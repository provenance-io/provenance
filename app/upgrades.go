package app

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	ibctmmigrations "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint/migrations"

	"github.com/provenance-io/provenance/x/exchange"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
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
	"wisteria-rc1": {}, // Upgrade for v1.21.0-rc1.
	"wisteria":     {}, // Upgrade for v1.21.0.
	"tokenomics2": { // Upgrade for v1.22.0.
		Handler: func(ctx sdk.Context, app *App, vm module.VersionMap) (module.VersionMap, error) {
			var err error
			if err = pruneIBCExpiredConsensusStates(ctx, app); err != nil {
				return nil, err
			}
			if vm, err = runModuleMigrations(ctx, app, vm); err != nil {
				return nil, err
			}
			removeInactiveValidatorDelegations(ctx, app)
			if err = convertFinishedVestingAccountsToBase(ctx, app); err != nil {
				return nil, err
			}
			if _, err = applyNewTokenomics(ctx, app, getMainnetTokenomicsOptions()); err != nil {
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
				emitUpgradeEvent(ctx, plan)
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
					emitUpgradeEvent(ctx, plan)
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

// emitUpgradeEvent emits an event indicating that an upgrade happened.
func emitUpgradeEvent(ctx sdk.Context, plan upgradetypes.Plan) {
	ctx.EventManager().EmitEvent(sdk.NewEvent("upgrade_applied",
		sdk.NewAttribute("name", plan.Name),
		sdk.NewAttribute("height", fmt.Sprintf("%d", plan.Height)),
	))
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

// vestingScheduleRow defines a max nhash amount and the number of months it should take to vest.
// Part of the tokenomics2 upgrade.
type vestingScheduleRow struct {
	Max    sdkmath.Int
	Months int
}

// vestingScheduleTable is an ordered list of vestingScheduleRow entries.
// Part of the tokenomics2 upgrade.
type vestingScheduleTable []vestingScheduleRow

// GetMonths returns the number of months that the vesting schedule should take for the given amount of nhash.
// Returns -1 if none of the rows apply.
// Part of the tokenomics2 upgrade.
func (t vestingScheduleTable) GetMonths(nhashAmt sdkmath.Int) int {
	for _, row := range t {
		if nhashAmt.LTE(row.Max) {
			return row.Months
		}
	}
	return -1
}

// getStandardVestingSchedule returns the standard vesting schedule to use for the hash that isn't moved.
// Part of the tokenomics2 upgrade.
func getStandardVestingSchedule() vestingScheduleTable {
	// Define the vesting schedule in hash so it's more readable.
	rv := []vestingScheduleRow{
		{Max: sdkmath.NewInt(1_000_000), Months: 3},
		{Max: sdkmath.NewInt(5_000_000), Months: 12},
		{Max: sdkmath.NewInt(100_000_000), Months: 36},
		{Max: sdkmath.NewInt(500_000_000), Months: 48},
		{Max: sdkmath.NewInt(1_000_000_000), Months: 60},
		{Max: sdkmath.NewInt(2_000_000_000), Months: 72},
		{Max: sdkmath.NewInt(100_000_000_000), Months: 120},
	}
	// Now turn them into nhash amounts.
	for i := 0; i < len(rv); i++ {
		rv[i].Max = rv[i].Max.MulRaw(1_000_000_000)
	}
	return rv
}

// tokenomicsOptions contains data and info about how the tokenomics moves should happen.
// Part of the tokenomics2 upgrade.
type tokenomicsOptions struct {
	vestingSchedule vestingScheduleTable
	fullMove        map[string]bool
	noMove          map[string]bool
}

// GetMonths returns the number of months that the vesting schedule should take for the given amount of nhash.
// Returns -1 if none of the rows apply.
// Part of the tokenomics2 upgrade.
func (o tokenomicsOptions) GetMonths(nhashAmt sdkmath.Int) int {
	return o.vestingSchedule.GetMonths(nhashAmt)
}

// IsFullMoved returns true if the provided addr should have ALL of its hash moved out.
// Part of the tokenomics2 upgrade.
func (o tokenomicsOptions) IsFullMoved(addr sdk.AccAddress) bool {
	return o.fullMove[string(addr)]
}

// IsNoMove returns true if the provided addr should have NONE of its hash moved.
// Part of the tokenomics2 upgrade.
func (o tokenomicsOptions) IsNoMove(addr sdk.AccAddress) bool {
	return o.noMove[string(addr)]
}

// GetAmountToKeepAndMove gets the amount the addr should keep and have moved (in that order) for the given address given the provided current amount.
// Part of the tokenomics2 upgrade.
func (o tokenomicsOptions) GetAmountToKeepAndMove(addr sdk.AccAddress, curAmt sdkmath.Int) (toKeep sdkmath.Int, toMove sdkmath.Int) {
	switch {
	case o.IsFullMoved(addr):
		return sdkmath.ZeroInt(), curAmt
	case o.IsNoMove(addr):
		return curAmt, sdkmath.ZeroInt()
	default:
		toKeep = getAmountToKeep(curAmt)
		toMove = curAmt.Sub(toKeep)
		return toKeep, toMove
	}
}

// GetEndTimes will return a map of number of months to epoch of when that will be (based on the current block time).
// Part of the tokenomics2 upgrade.
func (o tokenomicsOptions) GetEndTimes(now time.Time) map[int]int64 {
	rv := make(map[int]int64)
	for _, row := range o.vestingSchedule {
		rv[row.Months] = addMonths(now, row.Months)
	}
	return rv
}

// getMainnetTokenomicsOptions gets the tokenomics options to use on mainnet.
// Part of the tokenomics2 upgrade.
func getMainnetTokenomicsOptions() *tokenomicsOptions {
	rv := &tokenomicsOptions{
		vestingSchedule: getStandardVestingSchedule(),
	}

	fullMoves := []string{
		// TODO: Recheck this list.
		"pb1kxrcc2cfxsuqyctky463evkvhd8l5cwj5uj8hd",
		"pb1ttwznseyrt9vdgnf9na8hm2stvjhqsfdx9x7kwn83lc37qjv56dsyhegv2",
		"pb1w4uxkfwapxgv3xa45ylqwx8z3j9ptyt3acxs77",
		"pb10y6w3jdcr3xckxzhneuj0ndczttvqcd0py9l4v",
		"pb134yele9n9l8k0cdpnjl38rmdmwd80ztmzgmnyv",
		"pb1r346k6ewgjqsh0g52tnqkjkw4h8a4cguf9zj06",
		"pb13456yr5zqm7v6jhd49yq5gdkred7znfrgsetwy",
		"pb1r0uscn2m2hd6qv8pxgq9l7z0frenz4fczzd2fg",
		"pb1mjcng3qrwtytlv6pf4mrw79szy3228ed4gvxhl",
		"pb1mfejechmw7n70drzxkem3fu463ea9sc87kzz2m",
		"pb1qalmzw4dkky3h3wj9wzue6ey6tl878ceexff5d",
		"pb14xcqjuvkqea7tsm5vfk7fttgcgwctyqzs9r8g7",
		"pb1gggsa2594d2usu0pdpn0a8meva2vk4mqzz0n0c",
		"pb1mhul7n97lx3dnecz2s3jmz8dtqfx2awmc24ny9z43lwc5sm758qs38ay8t",
		"pb1zrlq6fmj4kd2ags638k7pp40js626243srtjqr",
		"pb1k73600d34k0677jn9xcy64u4tdde92xl9dl83x",
		"pb1h62gjes0878mum3j4y9n06pkj6s4ecepccycnr6nuq86g4jt037sgfc62v",
		"pb1rypc389x0nnw6zsrcmukemdz86c90x2vkgwydl",
		"pb1xvnfrc6lr6tqe7wq3awjpyjsnkzvutd7kmntj7",
		"pb19uwry6fmq7435upeq880u2x357rrux3f7zj8mawy6ert3ttvtcesld3g08",
		"pb1dsv3qsfwfem9843se8rlx0y4cjv0vv3nkzn8k9",
	}
	rv.fullMove = make(map[string]bool)
	for _, addr := range fullMoves {
		rv.fullMove[addr] = true
	}

	noMoves := []string{
		// System module accounts.
		// TODO: Recheck this list.
		"pb1r726yra3euctv92qqfxh45xztewgp2qja2sl6e",
		"pb12k2pyuylm9t7ugdvz67h9pg4gmmvhn5vkgwhug",
		"pb1a53udazy8ayufvy0s434pfwjcedzqv34lzfqxj",
		"pb1wsxce0ls59rtj70fwcrxmtmmv32vpgmgyz8rwq",
		"pb1fp9wuhq58pz53wxvv3tnrxkw8m8s6swpymt3m7",
	}
	rv.noMove = make(map[string]bool)
	for _, addr := range noMoves {
		rv.noMove[addr] = true
	}
	return rv
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

// getAmountToKeep gets the amount to keep out of the provided curAmt (assuming its for an address that gets split).
// Part of the tokenomics2 upgrade.
func getAmountToKeep(curAmt sdkmath.Int) sdkmath.Int {
	return curAmt.QuoRaw(2)
}

// nhashDenom is the denom for nhash.
// Part of the tokenomics2 upgrade.
const nhashDenom = "nhash"

// reduceNhashInCoins returns a new Coins with the amount of nhash reduced by half.
// Part of the tokenomics2 upgrade.
func reduceNhashInCoins(coins sdk.Coins) sdk.Coins {
	if coins == nil {
		return coins
	}
	rv := make(sdk.Coins, 0, len(coins))
	for _, coin := range coins {
		if coin.Denom != nhashDenom {
			rv = append(rv, coin)
			continue
		}
		newAmt := getAmountToKeep(coin.Amount)
		if newAmt.IsPositive() {
			rv = append(rv, sdk.NewCoin(nhashDenom, newAmt))
		}
	}
	return rv
}

// getNewDelVestingAndFree returns updated delegated vesting and delegated free amounts
// given the new original vesting amount and current delegated vesting and free amounts.
// Part of the tokenomics2 upgrade.
func getNewDelVestingAndFree(newOrigVest, oldDelVest, oldDelFree sdk.Coins) (newDelVest, newDelFree sdk.Coins) {
	if oldDelVest.IsZero() {
		// If nothing has been delegated, there's no need to change anything.
		return oldDelVest, oldDelFree
	}
	newDelVest = make(sdk.Coins, 0, len(oldDelVest))
	newDelFree = oldDelFree
	for _, oldDelVestCoin := range oldDelVest {
		origVestAmt := newOrigVest.AmountOf(oldDelVestCoin.Denom)
		if oldDelVestCoin.Amount.GT(origVestAmt) {
			// If the old amount of delegated vesting is more than the new original vesting amount,
			// we reduce the amount of delegated vesting to the original vesting amount
			// and add the difference to delegated free.
			newDelVest = append(newDelVest, sdk.NewCoin(oldDelVestCoin.Denom, origVestAmt))
			newDelFree = newDelFree.Add(sdk.NewCoin(oldDelVestCoin.Denom, oldDelVestCoin.Amount.Sub(origVestAmt)))
		} else {
			// Otherwise, we don't need to change anything; keep the whole thing.
			newDelVest = append(newDelVest, oldDelVestCoin)
		}
	}
	return newDelVest, newDelFree
}

// getNewVestingAmount returns the new nhash amount that should be vesting.
// Part of the tokenomics2 upgrade.
func getNewVestingAmount(amt sdkmath.Int, months int) sdkmath.Int {
	free := amt.QuoRaw(int64(months))
	return amt.Sub(free)
}

// addrAmt associates an address with an amount.
// Part of the tokenomics2 upgrade.
type addrAmt struct {
	Addr sdk.AccAddress
	Amt  sdkmath.Int
}

// newAddrAmt creates a new addrAmt.
// Part of the tokenomics2 upgrade.
func newAddrAmt(acc sdk.AccAddress, amt sdkmath.Int) *addrAmt {
	return &addrAmt{Addr: acc, Amt: amt}
}

// toHashAmountString converts the provided amount of nhash into an amount of hash.
// Part of the tokenomics2 upgrade.
func toHashAmountString(amt sdkmath.Int) string {
	str := amt.String()
	// Make sure it's got at least 10 digits (one for before the decimal, and nine after).
	if len(str) < 10 {
		str = strings.Repeat("0", 10-len(str)) + str
	}
	// Split it into the whole and fractional hash amounts.
	whole := str[:len(str)-9]
	fract := str[len(str)-9:]
	// Add commas to the whole part.
	lw := len(whole)
	if lw > 3 {
		wholeRz := make([]rune, 0, lw+(lw-1)/3)
		for i, digit := range whole {
			if i > 0 && (lw-i)%3 == 0 {
				wholeRz = append(wholeRz, ',')
			}
			wholeRz = append(wholeRz, digit)
		}
		whole = string(wholeRz)
	}
	return whole + "." + fract
}

// cancelExchangeNhashHolds will release all nhash commitments, cancel
// all orders with holds on nhash, and any payments with source nhash.
// Part of the tokenomics2 upgrade.
func cancelExchangeNhashHolds(ctx sdk.Context, app *App) error {
	logger := ctx.Logger()
	logger.Info("Cancelling exchange-related holds.")

	admin := app.ExchangeKeeper.GetAuthority()

	// Start with commitments.
	logger.Debug("Looking up commitments to release.")
	var commitments []exchange.Commitment
	app.ExchangeKeeper.IterateCommitments(ctx, func(commitment exchange.Commitment) bool {
		if !commitment.Amount.AmountOf(nhashDenom).IsZero() {
			commitment.Amount = sdk.Coins{sdk.NewCoin(nhashDenom, commitment.Amount.AmountOf(nhashDenom))}
			commitments = append(commitments, commitment)
		}
		return false
	})
	if len(commitments) > 0 {
		logger.Debug(fmt.Sprintf("Found %d commitments to release.", len(commitments)))
		for _, comm := range commitments {
			addr, err := sdk.AccAddressFromBech32(comm.Account)
			if err != nil {
				return fmt.Errorf("invalid commitment account address %q: %w", comm.Account, err)
			}
			err = app.ExchangeKeeper.ReleaseCommitment(ctx, comm.MarketId, addr, comm.Amount, "upgrade")
			if err != nil {
				return fmt.Errorf("error releasing commitment: %w", err)
			}
		}
		logger.Debug(fmt.Sprintf("Done releasing %d commitments.", len(commitments)))
	} else {
		logger.Debug("No commitments to release.")
	}

	// Now do orders.
	logger.Debug("Looking up orders to cancel.")
	var orders []*exchange.Order
	err := app.ExchangeKeeper.IterateOrders(ctx, func(order *exchange.Order) bool {
		if !order.GetHoldAmount().AmountOf(nhashDenom).IsZero() {
			orders = append(orders, order)
		}
		return false
	})
	if err != nil {
		return fmt.Errorf("error iterating orders: %w", err)
	}
	if len(orders) > 0 {
		logger.Debug(fmt.Sprintf("Found %d orders to cancel.", len(orders)))
		for _, order := range orders {
			err = app.ExchangeKeeper.CancelOrder(ctx, order.OrderId, admin)
			if err != nil {
				return fmt.Errorf("error cancelling order %d: %w", order.OrderId, err)
			}
		}
		logger.Debug(fmt.Sprintf("Done cancelling %d orders.", len(orders)))
	} else {
		logger.Debug("No orders to release.")
	}

	// And lastly, do payments.
	logger.Debug("Looking up payments to cancel.")
	var payments []*exchange.Payment
	app.ExchangeKeeper.IteratePayments(ctx, func(payment *exchange.Payment) bool {
		if !payment.SourceAmount.AmountOf(nhashDenom).IsZero() {
			payments = append(payments, payment)
		}
		return false
	})
	if len(payments) > 0 {
		logger.Debug(fmt.Sprintf("Found %d payments to cancel.", len(orders)))
		for _, payment := range payments {
			source, err := sdk.AccAddressFromBech32(payment.Source)
			if err != nil {
				return fmt.Errorf("invalid payment source %q: %w", payment.Source, err)
			}
			target, err := sdk.AccAddressFromBech32(payment.Target)
			if err != nil {
				return fmt.Errorf("invalid payment target %q: %w", payment.Target, err)
			}
			err = app.ExchangeKeeper.RejectPayment(ctx, target, source, payment.ExternalId)
			if err != nil {
				return fmt.Errorf("could not remove payment: %w", err)
			}
		}
		logger.Debug(fmt.Sprintf("Done cancelling %d payments.", len(orders)))
	} else {
		logger.Debug("No payments to cancel.")
	}

	return nil
}

// withdrawAllRewardsAndCommissions withdraws all delegator rewards and validator commissions.
// Part of the tokenomics2 upgrade.
func withdrawAllRewardsAndCommissions(ctx sdk.Context, app *App) error {
	logger := ctx.Logger()
	logger.Info("Withdrawing all delegator rewards and validator commissions.")

	delAddrCdc := app.AccountKeeper.AddressCodec()
	valAddrCdc := app.StakingKeeper.ValidatorAddressCodec()

	logger.Debug("Getting all Delegations.")
	delegations, err := app.StakingKeeper.GetAllDelegations(ctx)
	if err != nil {
		return fmt.Errorf("error getting all delegations: %w", err)
	}
	logger.Debug(fmt.Sprintf("There are %d delegations.", len(delegations)))
	var delAddr sdk.AccAddress
	var valAddr sdk.ValAddress
	for _, del := range delegations {
		delAddr, err = delAddrCdc.StringToBytes(del.DelegatorAddress)
		if err != nil {
			return fmt.Errorf("invalid delegator address %q: %w", del.DelegatorAddress, err)
		}
		valAddr, err = valAddrCdc.StringToBytes(del.ValidatorAddress)
		if err != nil {
			return fmt.Errorf("invalid validator address %q: %w", del.ValidatorAddress, err)
		}
		_, err = app.DistrKeeper.WithdrawDelegationRewards(ctx, delAddr, valAddr)
		if err != nil {
			return fmt.Errorf("error withdrawing %s delegation rewards for %s: %w", del.DelegatorAddress, del.ValidatorAddress, err)
		}
	}
	logger.Debug("Done withdrawing delegator rewards.")

	logger.Debug("Getting all validators.")
	validators, err := app.StakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return fmt.Errorf("error getting all validators: %w", err)
	}
	logger.Debug(fmt.Sprintf("There are %d validators.", len(validators)))
	for _, val := range validators {
		valAddr, err = valAddrCdc.StringToBytes(val.OperatorAddress)
		if err != nil {
			return fmt.Errorf("invalid operator/validator address %q: %w", val.OperatorAddress, err)
		}
		_, err = app.DistrKeeper.WithdrawValidatorCommission(ctx, valAddr)
		if err != nil {
			return fmt.Errorf("error withdrawing validator %s commission: %w", val.OperatorAddress, err)
		}
	}
	logger.Debug("Done withdrawing validator commissions.")

	return nil
}

// adjustStakes adjusts the amount of hash staked, returning the original and new amounts staked.
// Part of the tokenomics2 upgrade.
func adjustStakes(ctx sdk.Context, app *App) (stakedOrig, stakedNew sdkmath.Int, err error) {
	stakedOrig, stakedNew = sdkmath.ZeroInt(), sdkmath.ZeroInt()
	logger := ctx.Logger()
	logger.Info("Adjusting staked amounts.")

	delAddrCdc := app.AccountKeeper.AddressCodec()
	valAddrCdc := app.StakingKeeper.ValidatorAddressCodec()

	totalMoved := sdkmath.ZeroInt()

	logger.Debug("Getting all delegations.") // TODO: Remove this section since it's done elsewhere now.
	delegations, err := app.StakingKeeper.GetAllDelegations(ctx)
	if err != nil {
		return stakedOrig, stakedNew, fmt.Errorf("error getting all delegations: %w", err)
	}
	logger.Debug(fmt.Sprintf("Found %d delegations to adjust.", delegations))
	for _, del := range delegations {
		var delAddr sdk.AccAddress
		var valAddr sdk.ValAddress
		delAddr, err = delAddrCdc.StringToBytes(del.DelegatorAddress)
		if err != nil {
			return stakedOrig, stakedNew, fmt.Errorf("invalid delegator address %q: %w", del.DelegatorAddress, err)
		}
		valAddr, err = valAddrCdc.StringToBytes(del.ValidatorAddress)
		if err != nil {
			return stakedOrig, stakedNew, fmt.Errorf("invalid validator address %q: %w", del.ValidatorAddress, err)
		}

		var validator stakingtypes.Validator
		validator, err = app.StakingKeeper.GetValidator(ctx, valAddr)
		if err != nil {
			return stakedOrig, stakedNew, fmt.Errorf("error getting validator %s: %w", del.ValidatorAddress, err)
		}

		curAmt := validator.TokensFromShares(del.Shares).TruncateInt()
		stakedOrig = stakedOrig.Add(curAmt)

		sharesToRemove := del.Shares.QuoInt64(2)
		newShares := del.Shares.Sub(sharesToRemove)

		if bytes.Equal(delAddr, valAddr) && !validator.Jailed {
			// If this is a self-delegation, and they're not jailed, make sure the
			// delegated amount doesn't go below the min self delegation.
			if validator.TokensFromShares(newShares).TruncateInt().LT(validator.MinSelfDelegation) {
				newShares, err = validator.SharesFromTokens(validator.MinSelfDelegation)
				if err != nil {
					return stakedOrig, stakedNew, fmt.Errorf("error calculating shares from min self delegation for %s: %w",
						del.ValidatorAddress, err)
				}
				sharesToRemove = del.Shares.Sub(newShares)
			}
		}

		if !validator.TokensFromShares(sharesToRemove).TruncateInt().IsPositive() {
			continue
		}

		if err = app.StakingKeeper.Hooks().BeforeDelegationSharesModified(ctx, delAddr, valAddr); err != nil {
			return stakedOrig, stakedNew, fmt.Errorf("error from hook before-delegation-shares-modified: %w", err)
		}

		del.Shares = newShares

		if err = app.StakingKeeper.SetDelegation(ctx, del); err != nil {
			return stakedOrig, stakedNew, fmt.Errorf("error updating delegation by %s to %s: %w",
				del.DelegatorAddress, del.ValidatorAddress, err)
		}

		if err = app.StakingKeeper.Hooks().AfterDelegationModified(ctx, delAddr, valAddr); err != nil {
			return stakedOrig, stakedNew, fmt.Errorf("error from hook after-delegation-shares-modified: %w", err)
		}

		var amtToRemove sdkmath.Int
		_, amtToRemove, err = app.StakingKeeper.RemoveValidatorTokensAndShares(ctx, validator, sharesToRemove)

		fromPool := stakingtypes.BondedPoolName
		if !validator.IsBonded() {
			fromPool = stakingtypes.NotBondedPoolName
		}

		coinsToRemove := sdk.Coins{sdk.NewCoin(nhashDenom, amtToRemove)}
		err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, fromPool, delAddr, coinsToRemove)
		if err != nil {
			return stakedOrig, stakedNew, fmt.Errorf("error sending previously delegated %s to %s from %q: %w",
				coinsToRemove, del.DelegatorAddress, fromPool, err)
		}
		totalMoved = totalMoved.Add(amtToRemove)
	}
	logger.Debug("Done adjusting delegations.")

	validators, err := app.StakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return stakedOrig, stakedNew, fmt.Errorf("error getting all validators: %w", err)
	}
	for _, val := range validators {
		stakedOrig = stakedOrig.Add(val.Tokens)
		var moved sdkmath.Int
		val, moved, err = adjustValidatorStakes(ctx, app, val)
		if err != nil {
			return stakedOrig, stakedNew, fmt.Errorf("could not adjust stakes to validator %s: %w", val.OperatorAddress, err)
		}
		stakedNew = stakedNew.Add(val.Tokens)
		totalMoved = totalMoved.Add(moved)
	}

	return stakedOrig, stakedNew, nil
}

// adjustValidatorStakes adjusts the amount of hash staked to a specific validator.
// Part of the tokenomics2 upgrade.
func adjustValidatorStakes(ctx sdk.Context, app *App, val stakingtypes.Validator) (stakingtypes.Validator, sdkmath.Int, error) {
	totalMoved := sdkmath.ZeroInt()
	valAddr, err := app.StakingKeeper.ValidatorAddressCodec().StringToBytes(val.OperatorAddress)
	if err != nil {
		return val, totalMoved, fmt.Errorf("invalid validator address %q: %w", val.OperatorAddress, err)
	}

	// Unbond half of each of the delegations, sending the amount back to the delegator.
	dels, err := app.StakingKeeper.GetValidatorDelegations(ctx, valAddr)
	if err != nil {
		return val, totalMoved, fmt.Errorf("error getting delegations for %s: %w", val.OperatorAddress, err)
	}
	if len(dels) > 0 {
		ctx.Logger().Debug(fmt.Sprintf("Adjusting %d delegations to %s.", len(dels), val.OperatorAddress))
	}
	for _, del := range dels {
		var delAddr sdk.AccAddress
		delAddr, err = app.AccountKeeper.AddressCodec().StringToBytes(del.DelegatorAddress)
		if err != nil {
			return val, totalMoved, fmt.Errorf("invalid delegator address %q: %w", del.DelegatorAddress, err)
		}

		sharesToRemove := del.Shares.QuoInt64(2)
		newShares := del.Shares.Sub(sharesToRemove)

		if bytes.Equal(delAddr, valAddr) && !val.Jailed {
			// If this is a self-delegation, and they're not jailed, make sure
			// the delegated amount doesn't go below their min self delegation,
			if val.TokensFromShares(newShares).TruncateInt().LT(val.MinSelfDelegation) {
				newShares, err = val.SharesFromTokens(val.MinSelfDelegation)
				if err != nil {
					return val, totalMoved, fmt.Errorf("error calculating shares from min self delegation for %s: %w",
						del.ValidatorAddress, err)
				}
				sharesToRemove = del.Shares.Sub(newShares)
			}
		}

		if !val.TokensFromShares(sharesToRemove).TruncateInt().IsPositive() {
			continue
		}

		if err = app.StakingKeeper.Hooks().BeforeDelegationSharesModified(ctx, delAddr, valAddr); err != nil {
			return val, totalMoved, fmt.Errorf("error from hook before-delegation-shares-modified: %w", err)
		}

		del.Shares = newShares

		if err = app.StakingKeeper.SetDelegation(ctx, del); err != nil {
			return val, totalMoved, fmt.Errorf("error updating delegation by %s to %s: %w",
				del.DelegatorAddress, del.ValidatorAddress, err)
		}

		if err = app.StakingKeeper.Hooks().AfterDelegationModified(ctx, delAddr, valAddr); err != nil {
			return val, totalMoved, fmt.Errorf("error from hook after-delegation-shares-modified: %w", err)
		}

		var amtToRemove sdkmath.Int
		val, amtToRemove, err = app.StakingKeeper.RemoveValidatorTokensAndShares(ctx, val, sharesToRemove)

		fromPool := app.AccountKeeper.GetModuleAddress(stakingtypes.BondedPoolName)
		if !val.IsBonded() {
			fromPool = app.AccountKeeper.GetModuleAddress(stakingtypes.NotBondedPoolName)
		}

		coinsToRemove := sdk.Coins{sdk.NewCoin(nhashDenom, amtToRemove)}
		err = app.BankKeeper.UndelegateCoins(ctx, fromPool, delAddr, coinsToRemove)
		if err != nil {
			return val, totalMoved, fmt.Errorf("error sending previously delegated %s to %s from %q: %w",
				coinsToRemove, del.DelegatorAddress, fromPool, err)
		}
		totalMoved = totalMoved.Add(amtToRemove)
	}

	// Reduce any unbonding delegations for this validator.
	undels, err := app.StakingKeeper.GetUnbondingDelegationsFromValidator(ctx, valAddr)
	if err != nil {
		return val, totalMoved, fmt.Errorf("could not get unbonding delegations from validator %s: %w", val.OperatorAddress, err)
	}
	if len(undels) > 0 {
		ctx.Logger().Debug(fmt.Sprintf("Adjusting %d undelegations from %s.", len(dels), val.OperatorAddress))
	}
	notBondedToMove := sdkmath.ZeroInt()
	for _, undel := range undels {
		updated := false
		for i, entry := range undel.Entries {
			toKeep := getAmountToKeep(entry.Balance)
			if toKeep.IsZero() && !entry.Balance.IsZero() {
				// Balance must be 1, just leave it alone.
				continue
			}

			toMove := entry.Balance.Sub(toKeep)
			notBondedToMove = notBondedToMove.Add(toMove)
			entry.Balance = toKeep
			undel.Entries[i] = entry
			updated = true
		}
		if updated {
			err = app.StakingKeeper.SetUnbondingDelegation(ctx, undel)
			if err != nil {
				return val, totalMoved, fmt.Errorf("error setting unbonding delegation from validator %s by %s: %w",
					undel.ValidatorAddress, undel.DelegatorAddress, err)
			}
		}
		// TODO: Actually send the amount to move back to the delegator (early).
	}

	// Update the redelegation entries for this validator.
	// Delegation entries exist for these, which were updated above, and the funds moved.
	// So all we need to do is reduce the re-delegation entry.
	redels, err := app.StakingKeeper.GetRedelegationsFromSrcValidator(ctx, valAddr)
	if err != nil {
		return val, totalMoved, fmt.Errorf("error getting redelegations from validator %s: %w", val.OperatorAddress, err)
	}
	for _, redel := range redels {
		// Get the delegation entry that should exist with the updated amount.
		_ = redel // TODO: Finish the redelegation loop.
	}

	return val, totalMoved, nil
}

// applyNewTokenomics will go through all the accounts and reorganize hash as described in our new tokenomics paper.
// Part of the tokenomics2 upgrade.
func applyNewTokenomics(ctx sdk.Context, app *App, opts *tokenomicsOptions) (sdk.Events, error) {
	ctx.Logger().Info("Applying new tokenomics.")

	// This will emit too many events, so squelch them in a throw-away event manager.
	em := sdk.NewEventManager()
	ctx = ctx.WithEventManager(em)

	// Use a cache context for all of this for (hopefully) easier node/chain recovery if it fails.
	var writeCache func()
	ctx, writeCache = ctx.CacheContext()

	blockTime := ctx.BlockTime().UTC()
	endTimes := opts.GetEndTimes(blockTime)
	oneMonthFromNow := addMonths(blockTime, 1)
	nhashInitialTotal := app.BankKeeper.GetSupply(ctx, nhashDenom).Amount

	// We can't have any holds during this. The exchange module is the only thing that adds holds.
	// We can't just cancel the holds directly, though, because that'll break the exchange module.
	if err := cancelExchangeNhashHolds(ctx, app); err != nil {
		return nil, fmt.Errorf("error cancelling nhash holds: %w", err)
	}

	// Withdraw all rewards and commissions.
	if err := withdrawAllRewardsAndCommissions(ctx, app); err != nil {
		return nil, fmt.Errorf("error withdrawing rewards and comissions: %w", err)
	}

	stakedOrig, stakedNew, err := adjustStakes(ctx, app)
	if err != nil {
		return nil, fmt.Errorf("error adjusting stakes: %w", err)
	}

	ctx.Logger().Info("Identifying amounts to move and accounts to update.")
	var updatedAccts []sdk.AccountI
	nhashUnvestedNew := sdkmath.ZeroInt()
	nhashUnvestedOld := sdkmath.ZeroInt() // This will be the unvested amount from existing vesting accounts after reduction.
	var toMove []*addrAmt

	i := 0
	err = app.AccountKeeper.Accounts.Walk(ctx, nil, func(addr sdk.AccAddress, acctI sdk.AccountI) (stop bool, err error) {
		if i > 0 && i%10_000 == 0 {
			ctx.Logger().Info(fmt.Sprintf("Accounts examined: %d.", i))
		}
		i++

		logger := ctx.Logger().With("addr", addr.String())
		nhashOrig := app.BankKeeper.GetBalance(ctx, addr, nhashDenom)
		if nhashOrig.IsNil() || nhashOrig.IsZero() {
			// No nhash, we don't care about it.
			logger.Debug("Account does not have any nhash. Skipping.")
			return false, nil
		}

		nhashToKeep, nhashToMove := opts.GetAmountToKeepAndMove(addr, nhashOrig.Amount)

		switch acct := acctI.(type) {
		case *authtypes.BaseAccount:
			switch {
			case app.WasmKeeper.HasContractInfo(ctx, addr):
				// Don't do anything to wasm accounts because it might break the contract.
				logger.Debug("Smart contract account with nhash.")
				nhashToKeep = nhashOrig.Amount
				nhashToMove = sdkmath.ZeroInt()
			case nhashToKeep.IsPositive():
				// For all other base accounts keeping nhash, we convert them to a vesting account.
				months := opts.GetMonths(nhashToKeep)
				toVest := getNewVestingAmount(nhashToKeep, months)
				if toVest.IsPositive() {
					amt := sdk.Coins{sdk.NewCoin(nhashDenom, toVest)}
					endTime := endTimes[months]
					logger.Debug("Changing base account to continuous vesting account", "unvested", amt, "start", oneMonthFromNow, "end", endTime)
					var newAcct sdk.AccountI
					newAcct, err = vesting.NewContinuousVestingAccount(acct, amt, oneMonthFromNow, endTime)
					if err != nil {
						return true, fmt.Errorf("could not create continuous vesting account for %s with %s starting on %d ending on %d: %w",
							acct.Address, amt, oneMonthFromNow, endTime, err)
					}
					// TODO: Look up delegation amount and set DelegatedVesting and DelegatedFree.
					updatedAccts = append(updatedAccts, newAcct)
					nhashUnvestedNew = nhashUnvestedNew.Add(toVest)
				} else {
					logger.Debug("Base account does not have enough for a vesting account.")
				}
			}

		case *authtypes.ModuleAccount:
			// Don't do anything with the module accounts.
			logger.Debug("Identified module account with nhash.", "name", acct.Name)
			nhashToKeep = nhashOrig.Amount
			nhashToMove = sdkmath.ZeroInt()

		case *vesting.BaseVestingAccount:
			// There shouldn't be any of these.
			return true, fmt.Errorf("invalid account type %T in %s", acct, addr)

		case *vesting.ContinuousVestingAccount:
			if acct.OriginalVesting.AmountOf(nhashDenom).IsPositive() {
				// Cut the nhash in their original vesting amount in half.
				oldOrigVest := acct.OriginalVesting
				oldDelVest := acct.DelegatedVesting
				oldDelFree := acct.DelegatedFree
				acct.OriginalVesting = reduceNhashInCoins(oldOrigVest)
				// TODO: Look up delegation amount and set DelegatedVesting and DelegatedFree.
				acct.DelegatedVesting, acct.DelegatedFree = getNewDelVestingAndFree(acct.OriginalVesting, oldDelVest, oldDelFree)
				updatedAccts = append(updatedAccts, acct)
				logger.Debug("Updated ContinuousVestingAccount.",
					"OriginalVesting_old", oldOrigVest, "OriginalVesting_new", acct.OriginalVesting,
					"DelegatedVesting_old", oldDelVest, "DelegatedVesting_new", acct.DelegatedVesting,
					"DelegatedFree_old", oldDelFree, "DelegatedFree_new", acct.DelegatedFree,
				)

				unvested := acct.OriginalVesting.AmountOf(nhashDenom).Sub(acct.GetVestedCoins(blockTime).AmountOf(nhashDenom))
				nhashUnvestedOld = nhashUnvestedOld.Add(unvested)
			} else {
				logger.Debug("ContinuousVestingAccount does not have any vesting nhash.")
			}

		case *vesting.DelayedVestingAccount:
			if acct.OriginalVesting.AmountOf(nhashDenom).IsPositive() {
				// Cut the nhash in their original vesting amount in half.
				oldOrigVest := acct.OriginalVesting
				oldDelVest := acct.DelegatedVesting
				oldDelFree := acct.DelegatedFree
				acct.OriginalVesting = reduceNhashInCoins(oldOrigVest)
				// TODO: Look up delegation amount and set DelegatedVesting and DelegatedFree.
				acct.DelegatedVesting, acct.DelegatedFree = getNewDelVestingAndFree(acct.OriginalVesting, oldDelVest, oldDelFree)
				updatedAccts = append(updatedAccts, acct)
				logger.Debug("Updated DelayedVestingAccount.",
					"OriginalVesting_old", oldOrigVest, "OriginalVesting_new", acct.OriginalVesting,
					"DelegatedVesting_old", oldDelVest, "DelegatedVesting_new", acct.DelegatedVesting,
					"DelegatedFree_old", oldDelFree, "DelegatedFree_new", acct.DelegatedFree,
				)

				unvested := acct.OriginalVesting.AmountOf(nhashDenom).Sub(acct.GetVestedCoins(blockTime).AmountOf(nhashDenom))
				nhashUnvestedOld = nhashUnvestedOld.Add(unvested)
			} else {
				logger.Debug("DelayedVestingAccount does not have any vesting nhash.")
			}

		case *vesting.PeriodicVestingAccount:
			if acct.OriginalVesting.AmountOf(nhashDenom).IsPositive() {
				// Cut the nhash in each period in half and update the orig vesting amount accordingly.
				oldOrigVest := acct.OriginalVesting
				oldDelVest := acct.DelegatedVesting
				oldDelFree := acct.DelegatedFree
				acct.OriginalVesting = nil
				for p, period := range acct.VestingPeriods {
					acct.VestingPeriods[p].Amount = reduceNhashInCoins(period.Amount)
					acct.OriginalVesting = acct.OriginalVesting.Add(acct.VestingPeriods[i].Amount...)
				}
				// TODO: Look up delegation amount and set DelegatedVesting and DelegatedFree.
				acct.DelegatedVesting, acct.DelegatedFree = getNewDelVestingAndFree(acct.OriginalVesting, oldDelVest, oldDelFree)
				updatedAccts = append(updatedAccts, acct)
				logger.Debug("Updated PeriodicVestingAccount.",
					"OriginalVesting_old", oldOrigVest, "OriginalVesting_new", acct.OriginalVesting,
					"DelegatedVesting_old", oldDelVest, "DelegatedVesting_new", acct.DelegatedVesting,
					"DelegatedFree_old", oldDelFree, "DelegatedFree_new", acct.DelegatedFree,
				)

				unvested := acct.OriginalVesting.AmountOf(nhashDenom).Sub(acct.GetVestedCoins(blockTime).AmountOf(nhashDenom))
				nhashUnvestedOld = nhashUnvestedOld.Add(unvested)
			} else {
				logger.Debug("PeriodicVestingAccount does not have any vesting nhash.")
			}

		case *vesting.PermanentLockedAccount:
			if acct.OriginalVesting.AmountOf(nhashDenom).IsPositive() {
				// Cut the nhash in the orig vesting in half.
				oldOrigVest := acct.OriginalVesting
				oldDelVest := acct.DelegatedVesting
				oldDelFree := acct.DelegatedFree
				acct.OriginalVesting = reduceNhashInCoins(oldOrigVest)
				// TODO: Look up delegation amount and set DelegatedVesting and DelegatedFree.
				acct.DelegatedVesting, acct.DelegatedFree = getNewDelVestingAndFree(acct.OriginalVesting, oldDelVest, oldDelFree)
				updatedAccts = append(updatedAccts, acct)
				logger.Debug("Updated PermanentLockedAccount.",
					"OriginalVesting_old", oldOrigVest, "OriginalVesting_new", acct.OriginalVesting,
					"DelegatedVesting_old", oldDelVest, "DelegatedVesting_new", acct.DelegatedVesting,
					"DelegatedFree_old", oldDelFree, "DelegatedFree_new", acct.DelegatedFree,
				)

				updatedAccts = append(updatedAccts, acct)
				nhashUnvestedOld = nhashUnvestedOld.Add(acct.OriginalVesting.AmountOf(nhashDenom))
			} else {
				logger.Debug("PermanentLockedAccount does not have any vesting nhash.")
			}

		case *exchange.MarketAccount:
			// Nothing special needs to happen with any market accounts.
			// We'll move the nhash, but can't convert it to a vesting account, so we leave the account entry alone.
			logger.Debug("Identified market account with nhash.", "market id", acct.MarketId, "name", acct.MarketDetails.Name)

		case *markertypes.MarkerAccount:
			if acct.Denom == nhashDenom {
				// The nhash in here was burned at some point anyway, so take all of it.
				nhashToKeep = sdkmath.ZeroInt()
				nhashToMove = nhashOrig.Amount
				logger.Debug("Identified nhash marker account with nhash.")
			} else {
				// Otherwise, nothing special needs to happen.
				logger.Debug("Identified marker account with nhash.", "denom", acct.Denom)
			}
			// Either way, we'll do the move but can't convert it to a vesting account, so we leave it alone.

		case *ica.InterchainAccount:
			// Nothing special to do for these either.
			logger.Debug("Identified nhash marker account with nhash.")

		default:
			// This switch should have all possible account types handled.
			return true, fmt.Errorf("unknown account type %T", acctI)
		}

		logger.Debug("Account has nhash.", "addr", addr, "nhash in account", nhashOrig.Amount,
			"nhash to keep", nhashToKeep, "nhash to move", nhashToMove)
		if nhashToMove.IsPositive() {
			toMove = append(toMove, newAddrAmt(addr, nhashToMove))
		}

		return false, nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking all accounts: %w", err)
	}

	// Update all the accounts that we've modified.
	// This needs to happen before the transfers so that the vest accounts are updated first.
	ctx.Logger().Info(fmt.Sprintf("Setting %d updated accounts.", len(updatedAccts)))
	for _, acct := range updatedAccts {
		app.AccountKeeper.SetAccount(ctx, acct)
	}

	ctx.Logger().Info("Moving nhash.")
	// Bypass the marker send restrictions because otherwise we don't have permission.
	sendCtx := markertypes.WithBypass(ctx)
	nhashMoved := sdkmath.ZeroInt()
	for _, entry := range toMove {
		coinsToMove := sdk.Coins{sdk.NewCoin(nhashDenom, entry.Amt)}
		ctx.Logger().Debug("Moving nhash.", "addr", entry.Addr, "amount", coinsToMove)
		err = app.BankKeeper.SendCoinsFromAccountToModule(sendCtx, entry.Addr, markertypes.CoinPoolName, coinsToMove)
		if err != nil {
			return nil, fmt.Errorf("could not send %s from %s to marker module account: %w", coinsToMove, entry.Addr, err)
		}
		nhashMoved = nhashMoved.Add(entry.Amt)
	}
	nhashNotMoved := nhashInitialTotal.Sub(nhashMoved)

	// TODO: Redistribute the moved nhash as prescribed.

	nhashFinalTotal := app.BankKeeper.GetSupply(ctx, nhashDenom).Amount
	nhashUnvestedTotal := nhashUnvestedNew.Add(nhashUnvestedOld)
	// 100,000,000,000.123456789 = 25 characters = longest expected amount string.
	ctx.Logger().Info(fmt.Sprintf("Original Hash Supply: %25s", toHashAmountString(nhashInitialTotal)))
	ctx.Logger().Info(fmt.Sprintf("     New Hash Supply: %25s", toHashAmountString(nhashFinalTotal)))
	ctx.Logger().Info(fmt.Sprintf("          Hash Moved: %25s", toHashAmountString(nhashMoved)))
	ctx.Logger().Info(fmt.Sprintf("      Hash Not Moved: %25s", toHashAmountString(nhashNotMoved)))
	ctx.Logger().Info(fmt.Sprintf("   Hash Now Unvested: %25s", toHashAmountString(nhashUnvestedTotal)))
	ctx.Logger().Info(fmt.Sprintf("Original Hash Staked: %25s", toHashAmountString(stakedOrig)))
	ctx.Logger().Info(fmt.Sprintf("     New Hash Staked: %25s", toHashAmountString(stakedNew)))

	// Alright. It's all good. Write the cache ctx and get out of here.
	ctx.Logger().Debug("Writing cache context.")
	writeCache()

	ctx.Logger().Info("Done applying new tokenomics.")
	return em.Events(), nil
}
