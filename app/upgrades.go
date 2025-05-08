package app

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibctmmigrations "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint/migrations"

	"github.com/provenance-io/provenance/internal/provutils"
	"github.com/provenance-io/provenance/x/exchange"
)

// appUpgrade is an internal structure for defining all things for an upgrade.
type appUpgrade struct {
	// Added contains names of modules being added during an upgrade.
	Added []string
	// Deleted contains names of modules being removed during an upgrade.
	Deleted []string
	// Renamed contains info on modules being renamed during an upgrade.
	Renamed []storetypes.StoreRename
	// Handler is a function to execute during an upgrade.
	Handler func(sdk.Context, *App, module.VersionMap) (module.VersionMap, error)
}

// upgrades is where we define things that need to happen during an upgrade.
// If no Handler is defined for an entry, a no-op upgrade handler is still registered.
// If there's nothing that needs to be done for an upgrade, there still needs to be an
// entry in this map, but it can just be {}.
//
// If something is happening in the rc upgrade(s) that isn't being applied in the non-rc,
// or vice versa, please add comments explaining why in both entries.
//
// On the same line as the key, there should be a comment indicating the software version.
// Entries should be in chronological/alphabetical order, earliest first.
// I.e. Brand-new colors should be added to the bottom with the rcs first, then the non-rc.
var upgrades = map[string]appUpgrade{
	"yellow": { // Upgrade for v1.23.0.
		Handler: func(ctx sdk.Context, app *App, vm module.VersionMap) (module.VersionMap, error) {
			var err error
			if err = pruneIBCExpiredConsensusStates(ctx, app); err != nil {
				return nil, err
			}
			removeInactiveValidatorDelegations(ctx, app)
			if err = convertFinishedVestingAccountsToBase(ctx, app); err != nil {
				return nil, err
			}
			convertAcctsToVesting(ctx, app)
			return vm, nil
		},
	},
}

// InstallCustomUpgradeHandlers sets upgrade handlers for all entries in the upgrades map.
func InstallCustomUpgradeHandlers(app *App) {
	// Register all explicit appUpgrades
	for name, upgrade := range upgrades {
		// If the handler has been defined, add it here, otherwise, use no-op.
		var handler upgradetypes.UpgradeHandler
		if upgrade.Handler == nil {
			handler = func(goCtx context.Context, plan upgradetypes.Plan, versionMap module.VersionMap) (module.VersionMap, error) {
				ctx := sdk.UnwrapSDKContext(goCtx)
				ctx.Logger().Info(fmt.Sprintf("Applying no-op upgrade to %q", plan.Name))
				return versionMap, nil
			}
		} else {
			ref := upgrade
			handler = func(goCtx context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
				ctx := sdk.UnwrapSDKContext(goCtx)
				ctx.Logger().Info(fmt.Sprintf("Starting upgrade to %q", plan.Name), "version-map", vm)
				newVM, err := ref.Handler(ctx, app, vm)
				if err != nil {
					ctx.Logger().Error(fmt.Sprintf("Failed to upgrade to %q", plan.Name), "error", err)
				} else {
					ctx.Logger().Info(fmt.Sprintf("Successfully upgraded to %q", plan.Name), "version-map", newVM)
				}
				return newVM, err
			}
		}
		app.UpgradeKeeper.SetUpgradeHandler(name, handler)
	}
}

// GetUpgradeStoreLoader creates an StoreLoader for use in an upgrade.
// Returns nil if no upgrade info is found or the upgrade doesn't need a store loader.
func GetUpgradeStoreLoader(app *App, info upgradetypes.Plan) baseapp.StoreLoader {
	upgrade, found := upgrades[info.Name]
	if !found {
		return nil
	}

	if len(upgrade.Renamed) == 0 && len(upgrade.Deleted) == 0 && len(upgrade.Added) == 0 {
		app.Logger().Info("No store upgrades required",
			"plan", info.Name,
			"height", info.Height,
		)
		return nil
	}

	storeUpgrades := storetypes.StoreUpgrades{
		Added:   upgrade.Added,
		Renamed: upgrade.Renamed,
		Deleted: upgrade.Deleted,
	}
	app.Logger().Info("Store upgrades",
		"plan", info.Name,
		"height", info.Height,
		"upgrade.added", storeUpgrades.Added,
		"upgrade.deleted", storeUpgrades.Deleted,
		"upgrade.renamed", storeUpgrades.Renamed,
	)
	return upgradetypes.UpgradeStoreLoader(info.Height, &storeUpgrades)
}

// runModuleMigrations wraps standard logging around the call to app.mm.RunMigrations.
// In most cases, it should be the first thing done during a migration.
//
// If state is updated prior to this migration, you run the risk of writing state using
// a new format when the migration is expecting all state to be in the old format.
func runModuleMigrations(ctx sdk.Context, app *App, vm module.VersionMap) (module.VersionMap, error) {
	// Even if this function is no longer called, do not delete it. Keep it around for the next time it's needed.
	ctx.Logger().Info("Starting module migrations. This may take a significant amount of time to complete. Do not restart node.")
	newVM, err := app.mm.RunMigrations(ctx, app.configurator, vm)
	if err != nil {
		ctx.Logger().Error("Module migrations encountered an error.", "error", err)
		return nil, err
	}
	ctx.Logger().Info("Module migrations completed.")
	return newVM, nil
}

// removeInactiveValidatorDelegations unbonds all delegations from inactive validators, triggering their removal from the validator set.
// This should be applied in most upgrades.
func removeInactiveValidatorDelegations(ctx sdk.Context, app *App) {
	ctx.Logger().Info("Removing inactive validator delegations.")

	sParams, perr := app.StakingKeeper.GetParams(ctx)
	if perr != nil {
		ctx.Logger().Error(fmt.Sprintf("Could not get staking params: %v.", perr))
		return
	}

	unbondingTimeParam := sParams.UnbondingTime
	ctx.Logger().Info(fmt.Sprintf("Threshold: %d days", int64(unbondingTimeParam.Hours()/24)))

	validators, verr := app.StakingKeeper.GetAllValidators(ctx)
	if verr != nil {
		ctx.Logger().Error(fmt.Sprintf("Could not get all validators: %v.", perr))
		return
	}

	removalCount := 0
	for _, validator := range validators {
		if validator.IsUnbonded() {
			inactiveDuration := ctx.BlockTime().Sub(validator.UnbondingTime)
			if inactiveDuration >= unbondingTimeParam {
				ctx.Logger().Info(fmt.Sprintf("Validator %v has been inactive (unbonded) for %d days and will be removed.", validator.OperatorAddress, int64(inactiveDuration.Hours()/24)))
				valAddress, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
				if err != nil {
					ctx.Logger().Error(fmt.Sprintf("Invalid operator address: %s: %v.", validator.OperatorAddress, err))
					continue
				}

				delegations, err := app.StakingKeeper.GetValidatorDelegations(ctx, valAddress)
				if err != nil {
					ctx.Logger().Error(fmt.Sprintf("Could not delegations for validator %s: %v.", valAddress, perr))
					continue
				}

				for _, delegation := range delegations {
					ctx.Logger().Info(fmt.Sprintf("Undelegate delegator %v from validator %v of all shares (%v).", delegation.DelegatorAddress, validator.OperatorAddress, delegation.GetShares()))
					var delAddr sdk.AccAddress
					delegator := delegation.GetDelegatorAddr()
					delAddr, err = sdk.AccAddressFromBech32(delegator)
					if err != nil {
						ctx.Logger().Error(fmt.Sprintf("Failed to undelegate delegator %s from validator %s: could not parse delegator address: %v.", delegator, valAddress.String(), err))
						continue
					}
					_, _, err = app.StakingKeeper.Undelegate(ctx, delAddr, valAddress, delegation.GetShares())
					if err != nil {
						ctx.Logger().Error(fmt.Sprintf("Failed to undelegate delegator %s from validator %s: %v.", delegator, valAddress.String(), err))
						continue
					}
				}
				removalCount++
			}
		}
	}

	ctx.Logger().Info(fmt.Sprintf("A total of %d inactive (unbonded) validators have had all their delegators removed.", removalCount))
}

// pruneIBCExpiredConsensusStates prunes expired consensus states for IBC.
// This should be applied in most upgrades.
func pruneIBCExpiredConsensusStates(ctx sdk.Context, app *App) error {
	ctx.Logger().Info("Pruning expired consensus states for IBC.")
	_, err := ibctmmigrations.PruneExpiredConsensusStates(ctx, app.appCodec, app.IBCKeeper.ClientKeeper)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Unable to prune expired consensus states, error: %s.", err))
		return err
	}
	ctx.Logger().Info("Done pruning expired consensus states for IBC.")
	return nil
}

// convertFinishedVestingAccountsToBase will turn completed vesting accounts into regular BaseAccounts.
// This should be applied in most upgrades.
func convertFinishedVestingAccountsToBase(ctx sdk.Context, app *App) error {
	ctx.Logger().Info("Converting completed vesting accounts into base accounts.")
	blockTime := ctx.BlockTime().UTC().Unix()
	var updatedAccts []sdk.AccountI
	err := app.AccountKeeper.Accounts.Walk(ctx, nil, func(_ sdk.AccAddress, acctI sdk.AccountI) (stop bool, err error) {
		var baseVestAcct *vesting.BaseVestingAccount
		switch acct := acctI.(type) {
		case *vesting.ContinuousVestingAccount:
			baseVestAcct = acct.BaseVestingAccount
		case *vesting.DelayedVestingAccount:
			baseVestAcct = acct.BaseVestingAccount
		case *vesting.PeriodicVestingAccount:
			baseVestAcct = acct.BaseVestingAccount
		default:
			// We don't care about permanent locked because they never end.
			// Nothing should ever be a *vesting.BaseVestingAccount, so we ignore those too.
			// All other accounts aren't a vesting account, so there's nothing to do with them here.
			return false, nil
		}

		if baseVestAcct.EndTime <= blockTime {
			updatedAccts = append(updatedAccts, baseVestAcct.BaseAccount)
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("error walking accounts: %w", err)
	}

	if len(updatedAccts) == 0 {
		ctx.Logger().Info("No completed vesting accounts found.")
	} else {
		ctx.Logger().Info(fmt.Sprintf("Found %d completed vesting accounts. Updating them now.", len(updatedAccts)))
		for _, acct := range updatedAccts {
			app.AccountKeeper.SetAccount(ctx, acct)
		}
	}

	ctx.Logger().Info("Done converting completed vesting accounts into base accounts.")
	return nil
}

// Create a use of the standard helpers so that the linter neither complains about it not being used,
// nor complains about a nolint:unused directive that isn't needed because the function is used.
var (
	_ = runModuleMigrations
	_ = removeInactiveValidatorDelegations
	_ = pruneIBCExpiredConsensusStates
	_ = convertFinishedVestingAccountsToBase
)

// nhashDenom is the denom for nhash.
// Part of the yellow upgrade.
const nhashDenom = "nhash"

// convertAcctsToVesting will convert the provided accounts to vesting accounts.
// Part of the yellow upgrade.
func convertAcctsToVesting(ctx sdk.Context, app *App) {
	ctx.Logger().Info("Converting accounts to vesting accounts.")

	toConvert, toIgnore := getAcctsToConvertToVesting(ctx, app)
	ctx.Logger().Info(fmt.Sprintf("Identified %d accounts to convert.", len(toConvert)))
	ctx.Logger().Debug(fmt.Sprintf("Identified %d accounts to ignore.", len(toIgnore)))

	converter := newAcctConverter(ctx, app, toIgnore)
	for _, acct := range toConvert {
		// If there's an error, it's logged, but we want to keep moving, so we don't care about it here.
		_ = converter.convert(ctx, acct)
	}

	converter.logStats(ctx)

	dumpAccountsToFile(ctx, "accounts_converted", converter.converted)
	dumpAccountsToFile(ctx, "accounts_ignored", converter.ignored)
	if len(converter.skipped) > 0 {
		dumpAccountsToFile(ctx, "accounts_skipped", converter.skipped)
	}

	ctx.Logger().Info("Done converting accounts to vesting accounts.")
}

// acctInfo contains several pieces of information needed while converting accounts to vesting accounts.
// Part of the yellow upgrade.
type acctInfo struct {
	addr      sdk.AccAddress
	acctI     sdk.AccountI
	baseAcct  *authtypes.BaseAccount
	balance   sdkmath.Int
	delegated sdkmath.Int
	total     sdkmath.Int
	toVest    sdkmath.Int
	toKeep    sdkmath.Int
	delVest   sdkmath.Int
	delFree   sdkmath.Int
	startTime int64
	endTime   int64

	// vestingAcct is only populated upon conversion.
	newAcct *vesting.ContinuousVestingAccount

	reason string

	// These two pieces are just informational (for analysis).
	onHold      sdk.Coin
	groupPolicy *group.GroupPolicyInfo
}

// newAcctInfo creates a new acctInfo starting with the provided address and account.
// Part of the yellow upgrade.
func newAcctInfo(addr sdk.AccAddress, acctI sdk.AccountI) *acctInfo {
	return &acctInfo{
		addr:      addr,
		acctI:     acctI,
		balance:   sdkmath.ZeroInt(),
		delegated: sdkmath.ZeroInt(),
		total:     sdkmath.ZeroInt(),
		toVest:    sdkmath.ZeroInt(),
		toKeep:    sdkmath.ZeroInt(),
		delVest:   sdkmath.ZeroInt(),
		delFree:   sdkmath.ZeroInt(),
		onHold:    sdk.NewInt64Coin(nhashDenom, 0),
	}
}

// getAcctsToConvertToVestingOld returns info on each of the accounts that should be converted to a vesting account.
// Part of the yellow upgrade.
func getAcctsToConvertToVesting(ctx sdk.Context, app *App) (toConvert, toIgnore []*acctInfo) {
	blockTime := ctx.BlockTime().UTC()
	monthsToStart := int64(3)
	monthsToEnd := int64(48)
	startTime := addMonths(blockTime, int(monthsToStart))
	endTime := addMonths(blockTime, int(monthsToEnd))

	ibcAccts := identifyIBCAccounts(ctx, app)

	err := app.AccountKeeper.Accounts.Walk(ctx, nil, func(addr sdk.AccAddress, acctI sdk.AccountI) (stop bool, err error) {
		addrStr := addr.String()
		logger := ctx.Logger().With("account", addrStr)

		acct := newAcctInfo(addr, acctI)
		acct.startTime = startTime
		acct.endTime = endTime

		acct.balance = app.BankKeeper.GetBalance(ctx, addr, nhashDenom).Amount
		acct.delegated, err = app.StakingKeeper.GetDelegatorBonded(ctx, addr)
		if err != nil {
			acct.reason = fmt.Sprintf("error: could not look up delegated amount: %v", err)
			logger.Error(acct.reason)
			toIgnore = append(toIgnore, acct)
			return false, nil
		}

		acct.total = acct.balance.Add(acct.delegated)
		acct.toVest = acct.total.MulRaw(monthsToEnd - monthsToStart).QuoRaw(monthsToEnd)
		acct.toKeep = acct.total.Sub(acct.toVest)
		acct.delVest = acct.delegated
		if acct.delegated.GT(acct.toVest) {
			acct.delVest = acct.toVest
			acct.delFree = acct.delegated.Sub(acct.toVest)
		}

		var ok bool
		acct.baseAcct, ok = acctI.(*authtypes.BaseAccount)
		if !ok {
			// Only base accounts can be converted to vesting accounts.
			acct.reason = fmt.Sprintf("account has type %T", acctI)
			logger.Debug("Skipping account.", "reason", acct.reason)
			toIgnore = append(toIgnore, acct)
			return false, nil
		}

		if ibcAccts[addrStr] {
			// Don't do anything to IBC accounts because it might break that stuff.
			acct.reason = "account is an IBC-related account"
			logger.Debug("Skipping account.", "reason", acct.reason)
			toIgnore = append(toIgnore, acct)
			return false, nil
		}

		if app.WasmKeeper.HasContractInfo(ctx, addr) {
			// Don't do anything to wasm accounts because it might break the contract.
			acct.reason = "account is a smart contract"
			logger.Debug("Skipping account.", "reason", acct.reason)
			toIgnore = append(toIgnore, acct)
			return false, nil
		}

		app.ICAHostKeeper.GetAllInterchainAccounts(ctx)
		if app.Ics20WasmHooks.ContractKeeper.HasContractInfo(ctx, addr) {
			// Don't do anything to interchain wasm accounts because it might break the contract.
			acct.reason = "account is an ICS smart contract"
			logger.Debug("Skipping account.", "reason", acct.reason)
			toIgnore = append(toIgnore, acct)
			return false, nil
		}

		// Get a few more pieces of info that are useful for analysis. TODO[yellow]: Delete this.
		acct.onHold, err = app.HoldKeeper.GetHoldCoin(ctx, addr, nhashDenom)
		if err != nil {
			logger.Error("Could not get amount of nhash on hold.", "error", err)
		}

		gp, err := app.GroupKeeper.GroupPolicyInfo(ctx, &group.QueryGroupPolicyInfoRequest{Address: addrStr})
		switch {
		case err == nil:
			acct.groupPolicy = gp.Info
		case errors.Is(err, sdkerrors.ErrNotFound):
			// Not a group account. Nothing to do.
		default:
			logger.Error("Could not get group policy info.", "error", err)
		}

		if !acct.toVest.IsPositive() {
			acct.reason = "account does not own enough hash"
			logger.Debug("Skipping account.", "reason", acct.reason, "nhash", acct.total.String())
			toIgnore = append(toIgnore, acct)
			return false, nil
		}

		toConvert = append(toConvert, acct)

		return false, nil
	})

	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("error walking accounts: %v", err))
		return nil, nil
	}

	return toConvert, toIgnore
}

// addMonths will return an epoch that is the given months after the start time.
// Part of the yellow upgrade.
func addMonths(start time.Time, months int) int64 {
	newDay := start.Day()
	newMonth := start.Month()
	newYear := start.Year()
	if months > 12 {
		newYear += months / 12
		months %= 12
	}
	newMonth += time.Month(months)
	return time.Date(newYear, newMonth, newDay+1, 0, 0, 0, 0, start.Location()).Unix()
}

// identifyIBCAccounts looks up and returns all accounts that might be related to IBC. The keys are bech32 address strings.
// Part of the yellow upgrade.
func identifyIBCAccounts(ctx sdk.Context, app *App) map[string]bool {
	rv := make(map[string]bool)
	count := 0
	for _, icaAcct := range app.ICAHostKeeper.GetAllInterchainAccounts(ctx) {
		rv[icaAcct.AccountAddress] = true
		count++
		ctx.Logger().Debug(fmt.Sprintf("ICA host account: %q.", icaAcct.AccountAddress))
	}
	ctx.Logger().Debug(fmt.Sprintf("Found %d active ICA host accounts.", count))

	count = 0
	for _, channel := range app.ICAHostKeeper.GetAllActiveChannels(ctx) {
		addr := ibctransfertypes.GetEscrowAddress(channel.PortId, channel.ChannelId)
		addrStr := addr.String()
		rv[addrStr] = true
		count++
		ctx.Logger().Debug(fmt.Sprintf("Active IBC Escrow Account: %q.", addrStr))
	}
	ctx.Logger().Debug(fmt.Sprintf("Found %d active IBC escrow accounts.", count))

	chanReq := &channeltypes.QueryChannelsRequest{
		Pagination: &query.PageRequest{Limit: 18_446_744_073_709_551_615}, // Max uint64.
	}
	chanResp, err := app.IBCKeeper.Channels(ctx, chanReq)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Error getting info on all IBC channels: %v", err))
	} else {
		count = 0
		for _, channel := range chanResp.Channels {
			addr := ibctransfertypes.GetEscrowAddress(channel.PortId, channel.ChannelId)
			addrStr := addr.String()
			rv[addrStr] = true
			count++
			ctx.Logger().Debug(fmt.Sprintf("IBC escrow Account: %q.", addrStr))
		}
		ctx.Logger().Debug(fmt.Sprintf("Found %d IBC escrow accounts.", count))

		count = 0
		for _, channel := range chanResp.Channels {
			addr := ibctransfertypes.GetEscrowAddress(channel.Counterparty.PortId, channel.Counterparty.ChannelId)
			addrStr := addr.String()
			rv[addrStr] = true
			count++
			ctx.Logger().Debug(fmt.Sprintf("IBC counterparty escrow Account: %q.", addrStr))
		}
		ctx.Logger().Debug(fmt.Sprintf("Found %d IBC counterparty escrow accounts.", count))

		count = 0
		for _, channel := range chanResp.Channels {
			for _, connectionID := range channel.ConnectionHops {
				addrStr, found := app.ICAHostKeeper.GetInterchainAccountAddress(ctx, connectionID, channel.PortId)
				if found {
					rv[addrStr] = true
					count++
					ctx.Logger().Debug(fmt.Sprintf("ICA account: %q.", addrStr))
				}
			}
		}
		ctx.Logger().Debug(fmt.Sprintf("Found %d ICA accounts.", count))

		count = 0
		for _, channel := range chanResp.Channels {
			for _, connectionID := range channel.ConnectionHops {
				addrStr, found := app.ICAHostKeeper.GetInterchainAccountAddress(ctx, connectionID, channel.Counterparty.PortId)
				if found {
					rv[addrStr] = true
					count++
					ctx.Logger().Debug(fmt.Sprintf("ICA counterparty account: %q.", addrStr))
				}
			}
		}
		ctx.Logger().Debug(fmt.Sprintf("Found %d ICA counterparty accounts.", count))
	}

	ctx.Logger().Debug(fmt.Sprintf("Identified %d IBC-related accounts.", len(rv)))
	return rv
}

// acctConverter is a helper struct used to facilitate the conversion of accounts to vesting accounts.
// Part of the yellow upgrade.
type acctConverter struct {
	app *App

	marketIDs []uint32
	authority string

	ignored   []*acctInfo
	converted []*acctInfo
	skipped   []*acctInfo

	accountsAttempted int
	accountsConverted int

	totalBalances  sdkmath.Int
	totalVesting   sdkmath.Int
	totalDelegated sdkmath.Int
	totalDelVest   sdkmath.Int
	totalDelFree   sdkmath.Int
}

// newAcctConverter creates a new acctConverter and looks up all needed info.
// Part of the yellow upgrade.
func newAcctConverter(ctx sdk.Context, app *App, ignored []*acctInfo) *acctConverter {
	rv := &acctConverter{
		app:       app,
		authority: app.ExchangeKeeper.GetAuthority(),
		ignored:   ignored,

		totalBalances:  sdkmath.ZeroInt(),
		totalVesting:   sdkmath.ZeroInt(),
		totalDelegated: sdkmath.ZeroInt(),
		totalDelVest:   sdkmath.ZeroInt(),
		totalDelFree:   sdkmath.ZeroInt(),
	}

	app.ExchangeKeeper.IterateKnownMarketIDs(ctx, func(marketID uint32) bool {
		rv.marketIDs = append(rv.marketIDs, marketID)
		return false
	})

	for _, acct := range ignored {
		rv.recordSkipped(acct)
	}
	// recordSkipped adds the entries to .skipped, but we only want these in the ignored list (set above).
	rv.skipped = nil

	return rv
}

// recordConverted updates stats with an account that has been converted.
// Part of the yellow upgrade.
func (c *acctConverter) recordConverted(acct *acctInfo) {
	c.converted = append(c.converted, acct)
	c.accountsConverted++
	c.totalBalances = c.totalBalances.Add(acct.balance)
	c.totalVesting = c.totalVesting.Add(acct.toVest)
	c.totalDelegated = c.totalDelegated.Add(acct.delegated)
	c.totalDelVest = c.totalDelVest.Add(acct.delVest)
	c.totalDelFree = c.totalDelFree.Add(acct.delFree)
}

// recordSkipped updates stats with an account that is not being converted.
// Part of the yellow upgrade.
func (c *acctConverter) recordSkipped(acct *acctInfo) {
	c.skipped = append(c.skipped, acct)
	c.totalBalances = c.totalBalances.Add(acct.balance)
	c.totalDelegated = c.totalDelegated.Add(acct.delegated)
	c.totalDelFree = c.totalDelFree.Add(acct.delegated)
}

// convert will convert the provided account to a vesting account.
// Part of the yellow upgrade.
func (c *acctConverter) convert(sdkCtx sdk.Context, acct *acctInfo) (err error) {
	logger := sdkCtx.Logger().With("account", acct.addr.String())
	defer func() {
		if err != nil {
			logger.Error("Skipping account.", "reason", err)
			acct.reason = err.Error()
			c.recordSkipped(acct)
		} else {
			logger.Debug("Account converted.", "count", c.accountsConverted)
			c.recordConverted(acct)
		}
	}()

	c.accountsAttempted++
	ctx, writeCache := sdkCtx.CacheContext()

	if err = c.cancelExchangeHolds(ctx, acct.addr); err != nil {
		return fmt.Errorf("could not cancel exchange holds: %w", err)
	}

	locked := c.app.BankKeeper.LockedCoins(ctx, acct.addr)
	hasLockedHash, lockedHash := locked.Find(nhashDenom)
	if hasLockedHash && lockedHash.IsPositive() {
		return fmt.Errorf("account has %s on hold", lockedHash)
	}

	if !acct.total.IsPositive() || acct.balance.IsNegative() || acct.delegated.IsNegative() {
		return fmt.Errorf("account has invalid nhash amount %s = %s (balance) + %snhash (delegated)",
			acct.total, acct.balance, acct.delegated)
	}

	origVest := sdk.Coins{sdk.NewCoin(nhashDenom, acct.total)}
	newAcct, err := vesting.NewContinuousVestingAccount(acct.baseAcct, origVest, acct.startTime, acct.endTime)
	if err != nil {
		return fmt.Errorf("could not create new continuous vesting account: %w", err)
	}

	if acct.delegated.IsPositive() {
		newAcct.BaseVestingAccount.DelegatedVesting = sdk.NewCoins(sdk.NewCoin(nhashDenom, acct.delegated))
	}

	c.app.AccountKeeper.SetAccount(ctx, newAcct)
	writeCache()
	acct.newAcct = newAcct

	return nil
}

// cancelExchangeHolds cancels everything in the exchange module that has a hold on some nhash for the given address.
// Part of the yellow upgrade.
func (c *acctConverter) cancelExchangeHolds(ctx sdk.Context, addr sdk.AccAddress) error {
	// Release any commitments.
	for _, marketID := range c.marketIDs {
		committed := c.app.ExchangeKeeper.GetCommitmentAmount(ctx, marketID, addr)
		hasHash, nhashCoin := committed.Find(nhashDenom)
		if !hasHash || !nhashCoin.IsPositive() {
			continue
		}
		err := c.app.ExchangeKeeper.ReleaseCommitment(ctx, marketID, addr, sdk.Coins{nhashCoin}, "yellow upgrade")
		if err != nil {
			return fmt.Errorf("error releasing commitment of %s to market %d", nhashCoin, marketID)
		}
	}

	// Cancel all orders.
	var orderIDs []uint64
	c.app.ExchangeKeeper.IterateAddressOrders(ctx, addr, func(orderID uint64, _ byte) bool {
		order, err := c.app.ExchangeKeeper.GetOrder(ctx, orderID)
		if err != nil && order == nil {
			return false
		}
		hasHash, nhashCoin := order.GetHoldAmount().Find(nhashDenom)
		if hasHash && nhashCoin.IsPositive() {
			orderIDs = append(orderIDs, orderID)
		}
		return false
	})
	for _, orderID := range orderIDs {
		err := c.app.ExchangeKeeper.CancelOrder(ctx, orderID, c.authority)
		if err != nil {
			return fmt.Errorf("error canceling order %d: %w", orderID, err)
		}
	}

	// Cancel their payments.
	var externalIDs []string
	c.app.ExchangeKeeper.IteratePaymentsForSource(ctx, addr, func(payment *exchange.Payment) bool {
		hasHash, nhashCoin := payment.SourceAmount.Find(nhashDenom)
		if hasHash && nhashCoin.IsPositive() {
			externalIDs = append(externalIDs, payment.ExternalId)
		}
		return false
	})
	if len(externalIDs) > 0 {
		if err := c.app.ExchangeKeeper.CancelPayments(ctx, addr, externalIDs); err != nil {
			return fmt.Errorf("error canceling payments: %w", err)
		}
	}
	return nil
}

// logStats will output some stats about the conversion to the logger at the info level.
// Part of the yellow upgrade.
func (c *acctConverter) logStats(ctx sdk.Context) {
	logger := ctx.Logger()
	logger.Info(fmt.Sprintf("Accounts converted: %s", addCommas(strconv.Itoa(c.accountsConverted))))
	if c.accountsAttempted != c.accountsConverted {
		logger.Info(fmt.Sprintf("  Accounts skipped: %s", strconv.Itoa(c.accountsAttempted-c.accountsConverted)))
	}
	notVesting := c.totalBalances.Sub(c.totalVesting)
	logger.Info(fmt.Sprintf("Total balances --------------------> %25s hash", toHashString(c.totalBalances)))
	logger.Info(fmt.Sprintf("Total vesting  --------------------> %25s hash", toHashString(c.totalVesting)))
	logger.Info(fmt.Sprintf("Total not vesting  ----------------> %25s hash", toHashString(notVesting)))
	logger.Info(fmt.Sprintf("Total delegated  ------------------> %25s hash", toHashString(c.totalDelegated)))
	logger.Info(fmt.Sprintf("Total not delegated  --------------> %25s hash", toHashString(c.totalBalances.Sub(c.totalDelegated))))
	logger.Info(fmt.Sprintf("Total delegated vesting  ----------> %25s hash", toHashString(c.totalDelVest)))
	logger.Info(fmt.Sprintf("Total delegated not vesting  ------> %25s hash", toHashString(c.totalDelFree)))
	logger.Info(fmt.Sprintf("Total not delegated vesting  ------> %25s hash", toHashString(c.totalVesting.Sub(c.totalDelVest))))
	logger.Info(fmt.Sprintf("Total not delegated not vesting  --> %25s hash", toHashString(notVesting.Sub(c.totalDelFree))))
}

// toHashString returns a string of the provided nhash amount as a hash amount. Essentially, it multiplies the amount
// by 10^9 ensuring there's always 9 digits after the decimal and has at least one digit before the decimal.
// Part of the yellow upgrade.
func toHashString(amt sdkmath.Int) string {
	amtStr := amt.String()
	if len(amtStr) < 9 {
		return "0." + strings.Repeat("0", 9-len(amtStr)) + amtStr
	}
	if len(amtStr) == 9 {
		return "0." + amtStr
	}
	return addCommas(amtStr[:len(amtStr)-9]) + "." + amtStr[len(amtStr)-9:]
}

// addCommas will add commas to the provided string (assuming it's a big number).
// Part of the yellow upgrade.
func addCommas(amt string) string {
	if len(amt) <= 3 || strings.ContainsAny(amt, ",.") {
		return amt
	}

	lenAmt := len(amt)
	rv := make([]rune, 0, lenAmt+(lenAmt-1)/3)
	for i, digit := range amt {
		if i > 0 && (lenAmt-i)%3 == 0 {
			rv = append(rv, ',')
		}
		rv = append(rv, digit)
	}
	return string(rv)
}

// TODO[yellow]: Delete the dumpAccountsToFile and acctJSONEntry stuff.

// dumpConvertedToFile will write a file containing info about all the converted accounts.
// Part of the yellow upgrade.
func dumpAccountsToFile(ctx sdk.Context, baseName string, converted []*acctInfo) {
	filename := fmt.Sprintf("./%s_%d_%s.json", baseName, ctx.BlockHeight(), ctx.BlockTime().Format("2006-01-02_03-04-05"))
	// This uses .Error to make it easier to find in the logs.
	ctx.Logger().Error(fmt.Sprintf("Writing accounts to %q (not an error)", filename))
	file, err := os.Create(filename)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Error creating file: %q: %v", filename, err))
		return
	}

	var buf *bufio.Writer
	defer func() {
		if buf != nil && buf.Buffered() > 0 {
			ctx.Logger().Debug(fmt.Sprintf("Flushing final buffer with %d bytes.", buf.Buffered()))
			if err = buf.Flush(); err != nil {
				ctx.Logger().Error(fmt.Sprintf("Error flushing final buffer %q: %v.", filename, err))
			}
		}
		buf = nil

		if file != nil {
			if err = file.Sync(); err != nil {
				ctx.Logger().Error(fmt.Sprintf("Error syncing file %q: %v.", filename, err))
			}
			if err = file.Close(); err != nil {
				ctx.Logger().Error(fmt.Sprintf("Error closing file %q: %v.", filename, err))
			}
			file = nil
		}
		// This uses .Error to make it easier to find in the logs.
		ctx.Logger().Error(fmt.Sprintf("Done writing accounts to %q (not an error).", filename))
	}()

	buf = bufio.NewWriterSize(file, 32768)
	ctx.Logger().Debug(fmt.Sprintf("Buffer size: %d", buf.Size()))

	for i, acct := range converted {
		entry := acct.AsAcctJSONEntry()
		logger := ctx.Logger().With("addr", entry.Addr, "i", i)
		data, err := json.Marshal(entry)
		if err != nil {
			logger.Error("Could not marshal JSON for entry.", "error", err)
			continue
		}

		lead := provutils.Ternary(i == 0, "[\n  ", ",\n  ")
		_, err = buf.WriteString(lead)
		if err != nil {
			logger.Error("Could not write line lead to buffer.", "line_lead", lead, "error", err)
			return
		}

		if buf.Available() < len(data) {
			logger.Debug(fmt.Sprintf("Flushing buffer with %d bytes.", buf.Buffered()))
			if err = buf.Flush(); err != nil {
				ctx.Logger().Error("Could not flush buffer.", "error", err)
				return
			}
		}

		_, err = buf.Write(data)
		if err != nil {
			logger.Error("Could not write entry to buffer.", "error", err)
			return
		}
	}

	_, err = buf.WriteString("\n]\n")
	if err != nil {
		ctx.Logger().Error("Could not write ending.", "error", err)
		return
	}
}

// acctJSONEntry contains the info we want in the json file.
// Part of the yellow upgrade.
type acctJSONEntry struct {
	Addr        string `json:"addr"`
	Total       string `json:"total"`
	Undelegated string `json:"undelegated"`
	Delegated   string `json:"delegated"`
	ToKeep      string `json:"to_keep"`
	ToVest      string `json:"to_vest"`
	DelVest     string `json:"del_vest"`
	DelFree     string `json:"del_free"`

	// These are just extra pieces of info that might be useful for analysis.
	Sequence string `json:"sequence"`
	OnHold   string `json:"on_hold"`
	PubKey   string `json:"key"`
	Group    string `json:"group"`
	Type     string `json:"type"`

	Reason string `json:"reason,omitempty"`
}

// AsAcctJSONEntry returns an *acctJSONEntry representing this acctInfo; satisfies the canBeJSONEntry interface.
func (a *acctInfo) AsAcctJSONEntry() *acctJSONEntry {
	rv := &acctJSONEntry{
		Addr:        a.addr.String(),
		Undelegated: a.balance.String(),
		Delegated:   provutils.Ternary(a.delegated.IsZero(), "", a.delegated.String()),
		Total:       provutils.Ternary(a.total.IsZero(), "", a.total.String()),
		ToKeep:      provutils.Ternary(a.toKeep.IsZero(), "", a.toKeep.String()),
		ToVest:      provutils.Ternary(a.toVest.IsZero(), "", a.toVest.String()),
		DelVest:     provutils.Ternary(a.delVest.IsZero(), "", a.delVest.String()),
		DelFree:     provutils.Ternary(a.delFree.IsZero(), "", a.delFree.String()),
		Reason:      a.reason,

		OnHold: provutils.Ternary(a.onHold.IsZero(), "", a.onHold.Amount.String()),
	}

	if a.baseAcct != nil {
		rv.Sequence = strconv.FormatUint(a.baseAcct.Sequence, 10)
		rv.PubKey = provutils.Ternary(a.baseAcct.GetPubKey() == nil, "", "yes")
	}
	if a.groupPolicy != nil {
		rv.Group = strconv.FormatUint(a.groupPolicy.GroupId, 10) + "=" + a.groupPolicy.Metadata
	}

	switch {
	case a.groupPolicy != nil:
		rv.Type = "group"
	case a.baseAcct == nil:
		rv.Type = "unknown"
	case a.baseAcct.Sequence == 0 && a.baseAcct.GetPubKey() == nil:
		rv.Type = "new"
	default:
		rv.Type = "standard"
	}

	return rv
}
