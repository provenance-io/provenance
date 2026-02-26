package app

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	circuittypes "cosmossdk.io/x/circuit/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	ibctmmigrations "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint/migrations"

	flatfeestypes "github.com/provenance-io/provenance/x/flatfees/types"
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
	"daisy-rc1": {
		Deleted: []string{"interchainquery", "oracle"},
		Handler: func(ctx sdk.Context, app *App, vm module.VersionMap) (module.VersionMap, error) {
			var err error
			if vm, err = runModuleMigrations(ctx, app, vm); err != nil {
				return nil, err
			}

			foundation, team := getTestnetCircuitBreakerAddrs()
			setupCircuitBreakerPermissions(ctx, app, foundation, team)

			if err = pruneIBCExpiredConsensusStates(ctx, app); err != nil {
				return nil, err
			}
			if err = setupMsgStoreCodeFee(ctx, app); err != nil {
				return nil, err
			}
			removeInactiveValidatorDelegations(ctx, app)
			if err = convertFinishedVestingAccountsToBase(ctx, app); err != nil {
				return nil, err
			}

			if err = app.VaultKeeper.MigrateVaultAccountPaymentDenomDefaults(ctx); err != nil {
				return nil, err
			}

			if err = updateMsgFees(ctx, app); err != nil {
				return nil, err
			}
			return vm, nil
		},
	},
	"daisy": {
		Deleted: []string{"interchainquery", "oracle"},
		Handler: func(ctx sdk.Context, app *App, vm module.VersionMap) (module.VersionMap, error) {
			var err error
			if vm, err = runModuleMigrations(ctx, app, vm); err != nil {
				return nil, err
			}

			foundation, team := getMainnetCircuitBreakerAddrs()
			setupCircuitBreakerPermissions(ctx, app, foundation, team)

			if err = pruneIBCExpiredConsensusStates(ctx, app); err != nil {
				return nil, err
			}
			if err = setupMsgStoreCodeFee(ctx, app); err != nil {
				return nil, err
			}
			removeInactiveValidatorDelegations(ctx, app)
			if err = convertFinishedVestingAccountsToBase(ctx, app); err != nil {
				return nil, err
			}

			if err = app.VaultKeeper.MigrateVaultAccountPaymentDenomDefaults(ctx); err != nil {
				return nil, err
			}

			if err = updateMsgFees(ctx, app); err != nil {
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
		ctx.Logger().Error(fmt.Sprintf("Could not get all validators: %v.", verr))
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
					ctx.Logger().Error(fmt.Sprintf("Could not delegations for validator %s: %v.", valAddress, err))
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

// unlockVestingAccounts will convert the provided addrs from ContinuousVestingAccount to BaseAccount.
// This might be needed later, for another round of unlocks, so we're keeping it around in case.
func unlockVestingAccounts(ctx sdk.Context, app *App, addrs []sdk.AccAddress) {
	ctx.Logger().Info("Unlocking select vesting accounts.")

	ctx.Logger().Info(fmt.Sprintf("Identified %d accounts to unlock.", len(addrs)))
	for _, addr := range addrs {
		// We don't care about the error here because it's already logged, and we want to move on.
		_ = app.HoldKeeper.UnlockVestingAccount(ctx, addr)
	}

	ctx.Logger().Info("Done unlocking select vesting accounts.")
}

// Create a use of the standard helpers so that the linter neither complains about it not being used,
// nor complains about a nolint:unused directive that isn't needed because the function is used.
var (
	_ = runModuleMigrations
	_ = removeInactiveValidatorDelegations
	_ = pruneIBCExpiredConsensusStates
	_ = convertFinishedVestingAccountsToBase
	_ = unlockVestingAccounts
)

// setupMsgStoreCodeFee sets the flat fee for MsgStoreCode to $100 (100,000 musd).
// This allows smart contracts to be stored directly without governance proposals.
func setupMsgStoreCodeFee(ctx sdk.Context, app *App) error {
	ctx.Logger().Info("Setting up MsgStoreCode flat fee.")

	msgFee := flatfeestypes.MsgFee{
		MsgTypeUrl: "/cosmwasm.wasm.v1.MsgStoreCode",
		Cost:       sdk.NewCoins(sdk.NewInt64Coin("musd", 100000)),
	}

	if err := msgFee.Validate(); err != nil {
		return fmt.Errorf("invalid MsgStoreCode fee: %w", err)
	}

	if err := app.FlatFeesKeeper.SetMsgFee(ctx, msgFee); err != nil {
		return fmt.Errorf("failed to set MsgStoreCode fee: %w", err)
	}

	ctx.Logger().Info("MsgStoreCode flat fee set successfully.", "msg_type", msgFee.MsgTypeUrl, "fee_musd", "100000", "fee_usd", "$100")

	return nil
}

// updateMsgFees updates the flat fees for multiple message types.
// Metadata operations (record/session writes): Lower fees to encourage usage
// IBC client updates: Minimal fee for relayer operations
func updateMsgFees(ctx sdk.Context, app *App) error {
	ctx.Logger().Info("Updating message fees.")

	// Define all fee updates
	feeUpdates := []flatfeestypes.MsgFee{
		{
			MsgTypeUrl: "/provenance.metadata.v1.MsgWriteRecordRequest",
			Cost:       sdk.NewCoins(sdk.NewInt64Coin("musd", 10)), // $0.01
		},
		{
			MsgTypeUrl: "/provenance.metadata.v1.MsgWriteSessionRequest",
			Cost:       sdk.NewCoins(sdk.NewInt64Coin("musd", 50)), // $0.05
		},
		{
			MsgTypeUrl: "/ibc.core.client.v1.MsgUpdateClient",
			Cost:       sdk.NewCoins(sdk.NewInt64Coin("musd", 10)), // $0.01
		},
	}

	for _, msgFee := range feeUpdates {
		if err := msgFee.Validate(); err != nil {
			return fmt.Errorf("invalid msg fee for %s: %w", msgFee.MsgTypeUrl, err)
		}

		if err := app.FlatFeesKeeper.SetMsgFee(ctx, msgFee); err != nil {
			return fmt.Errorf("failed to set msg fee for %s: %w", msgFee.MsgTypeUrl, err)
		}
	}

	ctx.Logger().Info("All message fees updated successfully.", "total_updated", len(feeUpdates))
	return nil
}

// setupCircuitBreakerPermissions grants circuit breaker permissions
// during an upgrade using the caller-provided address lists.
// The upgrade handlers are responsible for selecting the correct
// mainnet/testnet address sets.
func setupCircuitBreakerPermissions(ctx sdk.Context, app *App, foundationAccounts, teamAccounts []string) {
	ctx.Logger().Info("Setting up circuit breaker permissions.")

	grant := func(addresses []string, level circuittypes.Permissions_Level) {
		for i, addrStr := range addresses {
			addr, err := sdk.AccAddressFromBech32(addrStr)
			if err != nil {
				ctx.Logger().Error(
					fmt.Sprintf(
						"Invalid address at index %d (%s): decoding bech32 failed. Skipping.",
						i, addrStr,
					),
				)
				continue
			}

			perms := circuittypes.Permissions{
				Level:         level,
				LimitTypeUrls: []string{},
			}

			if err := app.CircuitKeeper.Permissions.Set(ctx, addr, perms); err != nil {
				ctx.Logger().Error(
					fmt.Sprintf("Failed to grant permissions to %s. Skipping.", addrStr),
				)
				continue
			}

			ctx.Logger().Info(fmt.Sprintf("Granted Level %s to %s.", level, addrStr))
		}
	}
	// Apply Foundation Permissions (SUPER_ADMIN)
	grant(foundationAccounts, circuittypes.Permissions_LEVEL_SUPER_ADMIN)

	// Apply Team Permissions (ALL_MSGS)
	grant(teamAccounts, circuittypes.Permissions_LEVEL_ALL_MSGS)

	ctx.Logger().Info("Circuit breaker setup configured.")
}

// getTestnetCircuitBreakerAddrs returns the list of accounts for Testnet.
func getTestnetCircuitBreakerAddrs() (foundation []string, team []string) {
	foundation = []string{
		"tp1j3uudcktfm98k4f7e6d8z20laxt7zc690f950s", // Danny Wedul.
	}
	team = []string{
		// TODO: Add REAL Testnet team addresses (tp1...)
	}
	return foundation, team
}

// getMainnetCircuitBreakerAddrs returns the list of accounts for Mainnet.
func getMainnetCircuitBreakerAddrs() (foundation []string, team []string) {
	foundation = []string{
		"pb1yfstpvte3a9yyq57qtw5k2vjydpn58992ydnfz9h7d23uhzej3esvnkqay", // A group account owned by the foundation.
	}
	team = []string{
		"pb1yquakyj3p68ugvwmxrw53uv58rr2lv4xzt6ue2", // Danny Wedul.
		"pb1ktgsj6ap4cfq45g5dyw5vj47zhcrd6cl9p6z43", // Jason Davidson.
		"pb1ue08tptk5rlye5aujd2k3astacnmufqtstn82t", // Matt Conroy.
	}
	return foundation, team
}
