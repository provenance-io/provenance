package app

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	ibctmmigrations "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint/migrations"

	"github.com/provenance-io/provenance/internal/pioconfig"
	flatfeestypes "github.com/provenance-io/provenance/x/flatfees/types"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
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
	"zomp": { // Upgrade for v1.24.0
		Handler: func(ctx sdk.Context, app *App, vm module.VersionMap) (module.VersionMap, error) {
			unlockVestingAccounts(ctx, app, getMainnetUnlocks())
			return vm, nil
		},
	},
	"zomp-rc1": {}, // Upgrade for v1.24.0-rc1
	"zydeco-rc1": { // Upgrade for v1.25.0-rc1.
		Added:   []string{flatfeestypes.StoreKey},
		Deleted: []string{msgfeestypes.StoreKey},
		Handler: func(ctx sdk.Context, app *App, vm module.VersionMap) (module.VersionMap, error) {
			var err error
			if vm, err = runModuleMigrations(ctx, app, vm); err != nil {
				return nil, err
			}
			if err = pruneIBCExpiredConsensusStates(ctx, app); err != nil {
				return nil, err
			}
			removeInactiveValidatorDelegations(ctx, app)
			if err = convertFinishedVestingAccountsToBase(ctx, app); err != nil {
				return nil, err
			}
			if err = setupFlatFees(ctx, app); err != nil {
				return nil, err
			}
			return vm, nil
		},
	},
	"zydeco": { // Upgrade for v1.25.0.
		Handler: func(ctx sdk.Context, app *App, vm module.VersionMap) (module.VersionMap, error) {
			var err error
			if vm, err = runModuleMigrations(ctx, app, vm); err != nil {
				return nil, err
			}
			if err = pruneIBCExpiredConsensusStates(ctx, app); err != nil {
				return nil, err
			}
			removeInactiveValidatorDelegations(ctx, app)
			if err = convertFinishedVestingAccountsToBase(ctx, app); err != nil {
				return nil, err
			}
			if err = setupFlatFees(ctx, app); err != nil {
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
				ctx.EventManager().EmitEvent(sdk.NewEvent("upgrade", sdk.NewAttribute("name", plan.Name)))
				return versionMap, nil
			}
		} else {
			ref := upgrade
			handler = func(goCtx context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
				ctx := sdk.UnwrapSDKContext(goCtx)
				ctx.Logger().Info(fmt.Sprintf("Starting upgrade to %q", plan.Name), "version-map", vm)
				ctx.EventManager().EmitEvent(sdk.NewEvent("upgrade", sdk.NewAttribute("name", plan.Name)))
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

// unlockVestingAccounts will convert the provided addrs from ContinuousVestingAccount to BaseAccount.
func unlockVestingAccounts(ctx sdk.Context, app *App, addrs []sdk.AccAddress) {
	ctx.Logger().Info("Unlocking select vesting accounts.")

	ctx.Logger().Info(fmt.Sprintf("Identified %d accounts to unlock.", len(addrs)))
	for i, addr := range addrs {
		acctI := app.AccountKeeper.GetAccount(ctx, addr)
		if acctI == nil {
			ctx.Logger().Info(fmt.Sprintf("[%d/%d]: Cannot unlock account %s: account not found.", i+1, len(addrs), addr))
			continue
		}
		switch acct := acctI.(type) {
		case *vesting.ContinuousVestingAccount:
			app.AccountKeeper.SetAccount(ctx, acct.BaseAccount)
			ctx.Logger().Debug(fmt.Sprintf("[%d/%d]: Unlocked account: %s.", i+1, len(addrs), addr))
		case *vesting.DelayedVestingAccount:
			app.AccountKeeper.SetAccount(ctx, acct.BaseAccount)
			ctx.Logger().Debug(fmt.Sprintf("[%d/%d]: Unlocked account: %s.", i+1, len(addrs), addr))
		case *vesting.PeriodicVestingAccount:
			app.AccountKeeper.SetAccount(ctx, acct.BaseAccount)
			ctx.Logger().Debug(fmt.Sprintf("[%d/%d]: Unlocked account: %s.", i+1, len(addrs), addr))
		case *vesting.PermanentLockedAccount:
			app.AccountKeeper.SetAccount(ctx, acct.BaseAccount)
			ctx.Logger().Debug(fmt.Sprintf("[%d/%d]: Unlocked account: %s.", i+1, len(addrs), addr))
		default:
			ctx.Logger().Info(fmt.Sprintf("[%d/%d]: Cannot unlock account %s: not a vesting account: %T.", i+1, len(addrs), addr, acctI))
		}
	}

	ctx.Logger().Info("Done unlocking select vesting accounts.")
}

// getMainnetUnlocks gets a list of mainnet addresses to unlock.
func getMainnetUnlocks() []sdk.AccAddress {
	addrs := []string{
		"pb1yku08jr34cls842gv2c9r3ec9g3aqyhdhxn7gv",
		"pb1uf7txm4tce8tmg95yd5jwf9lsghdl0zk7yemmf",
		"pb12dy3ztax0yw3wlqkd5dan0q5fs8p9pzwspge63",
		"pb13n5l87am8fgkuaeyjx8jrk6flajuw6ywgupd3l",
		"pb1sms0k6d8ealyl3ttxsvptx7ztcthwpjq8memak",
		"pb1v2k2d2wf8kza6r24zwajfka0kc3qrj503luljr",
		"pb1qydxnkqygzn4qcmnvshln9v62334xgu46eq0ft",
		"pb1pywenrwq5mxg0k24nk5v4zty5n45c8pf2xcquz",
		"pb1nf0s5hlvzececp55rlw7243she76fwm6xktue2",
		"pb1rnpt0x6vsmgdekmpqa94qcn7377k3nzvalka4k",
		"pb16dhla87g7a2jexytcrh3jm3fhzrsjzwxz28e2t",
		"pb18gl8gyf44dvyjp4tr9jrzuhmdjdcwzxravfc6n",
		"pb108t72x4taapyvmtrvyslqfpwxkj0sh6q80vjvw",
		"pb1qwvmrd5dm3ah4ym5eu2s3k6p46hs65hvc0zqn3",
		"pb139x44e5l0e84h3awccfztaefurnf0z7j5w9rcaff4lfvxvv4hkvqxuz5d5",
		"pb1enagqpfq5x7hcae9khgatd9mm4r372mylnwrqp",
		"pb1evhycvsj44uk4j3u096j3g833qaw22qveyhul7",
		"pb1sjppc478pta6y3x5tu9l2938ejj0qzws46fhqh",
		"pb19zlct5wh6xfn3u4na7ujhnaudwt63p2qqhukth",
		"pb18xlkw9evkhwfk2hn7ss2f0xm6h7qra07fzpsgq",
		"pb1ts3zffs789nhdhsnfzfzuve9j72kwqpqsj0h7u",
		"pb1vw5fuucvfhnkhpfm2dezj4uwln2t8ckqjdhfft",
		"pb1f6vqww9f7g5arawrsvmwqh9zzyt0njyeyzlutu",
		"pb1x09ku54r4m5cllnwc0m8gqx32u76xvvq0ap026",
		"pb1pft6y8u8eay4v4ypj6hlzqvgtf20krfcnyfth5",
		"pb1vv530hgpgmt0mtsfwmtr4t2vm84x328z0xdhvw",
		"pb1q9y3g5wm9c8kljj4ktgcvpdyau8xzrxxa75zfz",
		"pb1ahp9lcj9m5vkhpdzw57d8xcjuygy6xvp94q9uw",
		"pb14af37xzm2hssdklq92r6jq86q95zqxg7nvemvh",
		"pb1nep5lpv8kg2q7yxvepdqwngdpuvgq5h62eapq5",
		"pb1043tjayw4fhfgm3x900w96u4ytxufgezw90f5p",
		"pb1uds0mh9q94fmjwxavkgxncls3n33j7kkqs5ns2",
		"pb1n9h425x77vrrl47qv25z7m7f0p396ncucxyg5r",
	}

	rv := make([]sdk.AccAddress, 0, len(addrs))
	for _, str := range addrs {
		// This will give an error if not on mainnet (where the HRP is "pb").
		// If that happens, we ignore the errors and essentially end up with an empty list.
		addr, err := sdk.AccAddressFromBech32(str)
		if err == nil {
			rv = append(rv, addr)
		}
	}

	return rv
}

// setupFlatFees defines the flatfees module params and msg costs.
// Part of the zydeco upgrade.
func setupFlatFees(ctx sdk.Context, app *App) error {
	ctx.Logger().Info("Setting up flat fees.")

	feeDefCoin := func(amount int64) sdk.Coin {
		return sdk.NewInt64Coin(flatfeestypes.DefaultFeeDefinitionDenom, amount)
	}

	params := flatfeestypes.Params{ // TODO[fees]: Set these params with more accurate values.
		DefaultCost: feeDefCoin(1),
		ConversionFactor: flatfeestypes.ConversionFactor{
			BaseAmount:      feeDefCoin(1),
			ConvertedAmount: sdk.NewInt64Coin(pioconfig.GetProvConfig().FeeDenom, 1),
		},
	}
	err := app.FlatFeesKeeper.SetParams(ctx, params)
	if err != nil {
		return fmt.Errorf("could not set x/flatfees params: %w", err)
	}

	// TODO[fees]: Define all the non-default msg fees.
	msgFees := []*flatfeestypes.MsgFee{
		// flatfeestypes.NewMsgFee("url", feeDefCoin(2)),
	}
	for _, msgFee := range msgFees {
		err = app.FlatFeesKeeper.SetMsgFee(ctx, *msgFee)
		if err != nil {
			return fmt.Errorf("could not set msg fee %s: %w", msgFee, err)
		}
	}

	ctx.Logger().Info("Done setting up flat fees.")
	return nil
}
