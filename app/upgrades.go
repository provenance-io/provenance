package app

import (
	"context"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	ibctmmigrations "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint/migrations"

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
			if err = convertAcctsToVesting(ctx, app, getAcctsToConvertToVesting()); err != nil {
				return nil, err
			}
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
func convertAcctsToVesting(ctx sdk.Context, app *App, addrs []sdk.AccAddress) error {
	ctx.Logger().Info(fmt.Sprintf("Converting %d specific accounts to vesting accounts.", len(addrs)))

	converter := newAcctConverter(ctx, app)
	for _, addr := range addrs {
		// If there's an error, it's logged, but we want to keep moving, so we don't care about it here.
		_ = converter.convert(ctx, addr)
	}

	// TODO[yellow]: Add stats to logs.
	ctx.Logger().Info(fmt.Sprintf("Done converting %d specific accounts to vesting accounts.", len(addrs)))
	return nil
}

// getAcctsToConvertToVesting returns the AccAddress of each account that should be converted to a vesting account.
func getAcctsToConvertToVesting() []sdk.AccAddress {
	bech32s := []string{
		// TODO[yellow]: Define all the accounts here.
	}

	var err error
	rv := make([]sdk.AccAddress, len(bech32s))
	for i, str := range bech32s {
		rv[i], err = sdk.AccAddressFromBech32(str)
		if err != nil {
			panic(fmt.Errorf("could not convert %q to AccAddress: %w", str, err))
		}
	}

	return rv
}

// acctConverter is a helper struct used to facilitate the conversion of accounts to vesting accounts.
type acctConverter struct {
	app *App

	blockTime time.Time
	startTime int64
	endTime   int64
	marketIDs []uint32
	authority string

	accountsAttempted  int
	accountsConverted  int
	undelegatedVesting sdkmath.Int
	delegatedVesting   sdkmath.Int
	totalVesting       sdkmath.Int
}

func newAcctConverter(ctx sdk.Context, app *App) *acctConverter {
	rv := &acctConverter{
		app:              app,
		blockTime:        ctx.BlockTime().UTC(),
		authority:        app.ExchangeKeeper.GetAuthority(),
		delegatedVesting: sdkmath.ZeroInt(),
		totalVesting:     sdkmath.ZeroInt(),
	}
	rv.startTime = rv.blockTime.Unix()
	rv.endTime = addMonths(rv.blockTime, 48)

	app.ExchangeKeeper.IterateKnownMarketIDs(ctx, func(marketID uint32) bool {
		rv.marketIDs = append(rv.marketIDs, marketID)
		return false
	})

	return rv
}

// convert will convert the provided account to a vesting account.
func (c *acctConverter) convert(sdkCtx sdk.Context, addr sdk.AccAddress) (err error) {
	logger := sdkCtx.Logger().With("account", addr.String())
	defer func() {
		if err != nil {
			logger.Error("Skipping account.", "reason", err)
		} else {
			logger.Debug("Account converted.")
		}
	}()

	c.accountsAttempted++
	ctx, writeCache := sdkCtx.CacheContext()

	acctI := c.app.AccountKeeper.GetAccount(ctx, addr)
	if acctI == nil {
		return fmt.Errorf("account does not exist")
	}

	baseAcct, ok := acctI.(*authtypes.BaseAccount)
	if !ok {
		return fmt.Errorf("account has unexpected type %T", acctI)
	}

	if c.app.WasmKeeper.HasContractInfo(ctx, addr) {
		// Don't do anything to wasm accounts because it might break the contract.
		return fmt.Errorf("account is a wasm account")
	}

	if err = c.cancelExchangeHolds(ctx, addr); err != nil {
		return fmt.Errorf("could not cancel exchange holds: %w", err)
	}

	locked := c.app.BankKeeper.LockedCoins(ctx, addr)
	hasLockedHash, lockedHash := locked.Find(nhashDenom)
	if hasLockedHash && lockedHash.IsPositive() {
		return fmt.Errorf("account has %s on hold", lockedHash)
	}

	delegated, err := c.app.StakingKeeper.GetDelegatorBonded(ctx, addr)
	if err != nil {
		return fmt.Errorf("could not look up delegated amount: %w", err)
	}

	nhashBal := c.app.BankKeeper.GetBalance(ctx, addr, nhashDenom)
	toVest := nhashBal.AddAmount(delegated)
	if !toVest.IsPositive() || nhashBal.IsNegative() || delegated.IsNegative() {
		return fmt.Errorf("account has invalid nhash amount %s = %s (balance) + %snhash (delegated)", toVest, nhashBal, delegated)
	}

	newAcct, err := vesting.NewContinuousVestingAccount(baseAcct, sdk.Coins{toVest}, c.startTime, c.endTime)
	if err != nil {
		return fmt.Errorf("could not create new continuous vesting account: %w", err)
	}

	if delegated.IsPositive() {
		newAcct.BaseVestingAccount.DelegatedVesting = sdk.NewCoins(sdk.NewCoin(nhashDenom, delegated))
	}

	c.app.AccountKeeper.SetAccount(ctx, newAcct)
	writeCache()

	c.accountsConverted++
	c.undelegatedVesting = c.undelegatedVesting.Add(nhashBal.Amount)
	c.delegatedVesting = c.delegatedVesting.Add(delegated)
	c.totalVesting = c.totalVesting.Add(toVest.Amount)

	return nil
}

// addMonths will return an epoch that is the given months after the start time.
// Part of the tokenomics2 upgrade.
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

// cancelExchangeHolds cancels everything in the exchange module that has a hold on some nhash for the given address.
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
