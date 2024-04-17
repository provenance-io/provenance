package app

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibctmmigrations "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint/migrations"
	attributetypes "github.com/provenance-io/provenance/x/attribute/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
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
// On the same line as the key, there should be a comment indicating the software version.
// Entries currently in use (e.g. on mainnet or testnet) cannot be deleted.
// Entries should be in chronological order, earliest first. E.g. quicksilver-rc1 went to
// testnet first, then quicksilver-rc2 went to testnet, then quicksilver went to mainnet.
//
// If something is happening in the rc upgrade(s) that isn't being applied in the non-rc,
// or vice versa, please add comments explaining why in both entries.
var upgrades = map[string]appUpgrade{
	"umber-rc1": { // upgrade for v1.19.0-rc1
		Added:   []string{crisistypes.ModuleName},
		Deleted: []string{"reward"},
		Handler: func(ctx sdk.Context, app *App, vm module.VersionMap) (module.VersionMap, error) {
			var err error

			if err = pruneIBCExpiredConsensusStates(ctx, app); err != nil {
				return nil, err
			}

			err = migrateBaseappParams(ctx, app)
			if err != nil {
				return nil, err
			}

			err = migrateBankParams(ctx, app)
			if err != nil {
				return nil, err
			}

			migrateAttributeParams(ctx, app)
			migrateMarkerParams(ctx, app)
			migrateMetadataOSLocatorParams(ctx, app)
			migrateMsgFeesParams(ctx, app)

			vm, err = runModuleMigrations(ctx, app, vm)
			if err != nil {
				return nil, err
			}

			err = updateIBCClients(ctx, app)
			if err != nil {
				return nil, err
			}

			removeInactiveValidatorDelegations(ctx, app)

			return vm, nil
		},
	},
	"umber": { // upgrade for v1.19.0
		Added:   []string{crisistypes.ModuleName},
		Deleted: []string{"reward"},
		Handler: func(ctx sdk.Context, app *App, vm module.VersionMap) (module.VersionMap, error) {
			var err error

			if err = pruneIBCExpiredConsensusStates(ctx, app); err != nil {
				return nil, err
			}

			err = migrateBaseappParams(ctx, app)
			if err != nil {
				return nil, err
			}

			err = migrateBankParams(ctx, app)
			if err != nil {
				return nil, err
			}

			migrateAttributeParams(ctx, app)
			migrateMarkerParams(ctx, app)
			migrateMetadataOSLocatorParams(ctx, app)
			migrateMsgFeesParams(ctx, app)

			vm, err = runModuleMigrations(ctx, app, vm)
			if err != nil {
				return nil, err
			}

			err = updateIBCClients(ctx, app)
			if err != nil {
				return nil, err
			}

			removeInactiveValidatorDelegations(ctx, app)

			return vm, nil
		},
	},
	// TODO - Add new upgrade definitions here.
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

// Create a use of runModuleMigrations so that the linter neither complains about it not being used,
// nor complains about a nolint:unused directive that isn't needed because the function is used.
var _ = runModuleMigrations

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

// updateIBCClients updates the allowed clients for IBC.
// TODO: Remove with the umber handlers.
func updateIBCClients(ctx sdk.Context, app *App) error {
	ctx.Logger().Info("Updating IBC AllowedClients.")
	params := app.IBCKeeper.ClientKeeper.GetParams(ctx)
	params.AllowedClients = append(params.AllowedClients, exported.Localhost)
	app.IBCKeeper.ClientKeeper.SetParams(ctx, params)
	ctx.Logger().Info("Done updating IBC AllowedClients.")
	return nil
}

// migrateBaseappParams migrates to new ConsensusParamsKeeper
// TODO: Remove with the umber handlers.
func migrateBaseappParams(ctx sdk.Context, app *App) error {
	ctx.Logger().Info("Migrating legacy params.")
	legacyBaseAppSubspace := app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
	err := baseapp.MigrateParams(ctx, legacyBaseAppSubspace, app.ConsensusParamsKeeper.ParamsStore)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Unable to migrate legacy params to ConsensusParamsKeeper, error: %s.", err))
		return err
	}
	ctx.Logger().Info("Done migrating legacy params.")
	return nil
}

// migrateBankParams migrates the bank params from the params module to the bank module's state.
// The SDK has this as part of their bank v4 migration, but we're already on v4, so that one
// won't run on its own. This is the only part of that migration that we still need to have
// done, and this brings us in-line with format the bank state on v4.
// TODO: delete with the umber handlers.
func migrateBankParams(ctx sdk.Context, app *App) (err error) {
	ctx.Logger().Info("Migrating bank params.")
	defer func() {
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("Unable to migrate bank params, error: %s.", err))
		}
		ctx.Logger().Info("Done migrating bank params.")
	}()

	bankParamsSpace, ok := app.ParamsKeeper.GetSubspace(banktypes.ModuleName)
	if !ok {
		return fmt.Errorf("params subspace not found: %q", banktypes.ModuleName)
	}
	return app.BankKeeper.MigrateParamsProv(ctx, bankParamsSpace)
}

// migrateAttributeParams migrates to new Attribute Params store
// TODO: Remove with the umber handlers.
func migrateAttributeParams(ctx sdk.Context, app *App) {
	ctx.Logger().Info("Migrating attribute params.")
	attributeParamSpace := app.ParamsKeeper.Subspace(attributetypes.ModuleName).WithKeyTable(attributetypes.ParamKeyTable())
	maxValueLength := uint32(attributetypes.DefaultMaxValueLength)
	// TODO: remove attributetypes.ParamStoreKeyMaxValueLength with the umber handlers.
	if attributeParamSpace.Has(ctx, attributetypes.ParamStoreKeyMaxValueLength) {
		attributeParamSpace.Get(ctx, attributetypes.ParamStoreKeyMaxValueLength, &maxValueLength)
	}
	app.AttributeKeeper.SetParams(ctx, attributetypes.Params{MaxValueLength: uint32(maxValueLength)})
	ctx.Logger().Info("Done migrating attribute params.")
}

// migrateMarkerParams migrates to new Marker Params store
// TODO: Remove with the umber handlers.
func migrateMarkerParams(ctx sdk.Context, app *App) {
	ctx.Logger().Info("Migrating marker params.")
	markerParamSpace := app.ParamsKeeper.Subspace(markertypes.ModuleName).WithKeyTable(markertypes.ParamKeyTable())

	params := markertypes.DefaultParams()

	// TODO: remove markertypes.ParamStoreKeyMaxTotalSupply with the umber handlers.
	if markerParamSpace.Has(ctx, markertypes.ParamStoreKeyMaxTotalSupply) {
		var maxTotalSupply uint64
		markerParamSpace.Get(ctx, markertypes.ParamStoreKeyMaxTotalSupply, &maxTotalSupply)
		params.MaxTotalSupply = maxTotalSupply
	}

	// TODO: remove markertypes.ParamStoreKeyEnableGovernance with the umber handlers.
	if markerParamSpace.Has(ctx, markertypes.ParamStoreKeyEnableGovernance) {
		var enableGovernance bool
		markerParamSpace.Get(ctx, markertypes.ParamStoreKeyEnableGovernance, &enableGovernance)
		params.EnableGovernance = enableGovernance
	}

	// TODO: remove markertypes.ParamStoreKeyUnrestrictedDenomRegex with the umber handlers.
	if markerParamSpace.Has(ctx, markertypes.ParamStoreKeyUnrestrictedDenomRegex) {
		var unrestrictedDenomRegex string
		markerParamSpace.Get(ctx, markertypes.ParamStoreKeyUnrestrictedDenomRegex, &unrestrictedDenomRegex)
		params.UnrestrictedDenomRegex = unrestrictedDenomRegex
	}

	// TODO: remove markertypes.ParamStoreKeyMaxSupply with the umber handlers.
	if markerParamSpace.Has(ctx, markertypes.ParamStoreKeyMaxSupply) {
		var maxSupply string
		markerParamSpace.Get(ctx, markertypes.ParamStoreKeyMaxSupply, &maxSupply)
		params.MaxSupply = markertypes.StringToBigInt(maxSupply)
	}

	app.MarkerKeeper.SetParams(ctx, params)

	ctx.Logger().Info("Done migrating marker params.")
}

// migrateAttributeParams migrates to new Metadata Os Locator Params store
// TODO: Remove with the umber handlers.
func migrateMetadataOSLocatorParams(ctx sdk.Context, app *App) {
	ctx.Logger().Info("Migrating metadata os locator params.")
	metadataParamSpace := app.ParamsKeeper.Subspace(metadatatypes.ModuleName).WithKeyTable(metadatatypes.ParamKeyTable())
	maxValueLength := uint32(metadatatypes.DefaultMaxURILength)
	// TODO: remove metadatatypes.ParamStoreKeyMaxValueLength with the umber handlers.
	if metadataParamSpace.Has(ctx, metadatatypes.ParamStoreKeyMaxValueLength) {
		metadataParamSpace.Get(ctx, metadatatypes.ParamStoreKeyMaxValueLength, &maxValueLength)
	}
	app.MetadataKeeper.SetOSLocatorParams(ctx, metadatatypes.OSLocatorParams{MaxUriLength: uint32(maxValueLength)})
	ctx.Logger().Info("Done migrating metadata os locator params.")

}

// migrateMsgFeesParams migrates to new MsgFees Params store
// TODO: Remove with the umber handlers.
func migrateMsgFeesParams(ctx sdk.Context, app *App) {
	ctx.Logger().Info("Migrating msgfees params.")
	msgFeesParamSpace := app.ParamsKeeper.Subspace(msgfeestypes.ModuleName).WithKeyTable(msgfeestypes.ParamKeyTable())

	var floorGasPrice sdk.Coin
	if msgFeesParamSpace.Has(ctx, msgfeestypes.ParamStoreKeyFloorGasPrice) {
		msgFeesParamSpace.Get(ctx, msgfeestypes.ParamStoreKeyFloorGasPrice, &floorGasPrice)
	}

	var nhashPerUsdMil uint64
	if msgFeesParamSpace.Has(ctx, msgfeestypes.ParamStoreKeyNhashPerUsdMil) {
		msgFeesParamSpace.Get(ctx, msgfeestypes.ParamStoreKeyNhashPerUsdMil, &nhashPerUsdMil)
	}

	var conversionFeeDenom string
	if msgFeesParamSpace.Has(ctx, msgfeestypes.ParamStoreKeyConversionFeeDenom) {
		msgFeesParamSpace.Get(ctx, msgfeestypes.ParamStoreKeyConversionFeeDenom, &conversionFeeDenom)
	}

	migratedParams := msgfeestypes.Params{
		FloorGasPrice:      floorGasPrice,
		NhashPerUsdMil:     nhashPerUsdMil,
		ConversionFeeDenom: conversionFeeDenom,
	}
	app.MsgFeesKeeper.SetParams(ctx, migratedParams)

	ctx.Logger().Info("Done migrating msgfees params.")
}
