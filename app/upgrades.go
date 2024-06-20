package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	circuittypes "cosmossdk.io/x/circuit/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibctmmigrations "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint/migrations"

	attributetypes "github.com/provenance-io/provenance/x/attribute/types"
	ibchookstypes "github.com/provenance-io/provenance/x/ibchooks/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
	nametypes "github.com/provenance-io/provenance/x/name/types"
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
		Added:   []string{crisistypes.ModuleName, circuittypes.ModuleName, consensusparamtypes.ModuleName},
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
			migrateNameParams(ctx, app)
			migrateIbcHooksParams(ctx, app)

			vm, err = runModuleMigrations(ctx, app, vm)
			if err != nil {
				return nil, err
			}

			err = setNewGovParamsTestnet(ctx, app)
			if err != nil {
				return nil, err
			}

			updateIBCClients(ctx, app)

			scopeNavs, err := ReadNetAssetValues("upgrade_files/testnet_scope_navs.csv")
			if err != nil {
				return nil, err
			}
			addScopeNavsWithHeight(ctx, app, scopeNavs)

			removeInactiveValidatorDelegations(ctx, app)

			return vm, nil
		},
	},
	"umber": { // upgrade for v1.19.0
		Added:   []string{crisistypes.ModuleName, circuittypes.ModuleName, consensusparamtypes.ModuleName},
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
			migrateNameParams(ctx, app)
			migrateIbcHooksParams(ctx, app)

			vm, err = runModuleMigrations(ctx, app, vm)
			if err != nil {
				return nil, err
			}

			err = setNewGovParamsMainnet(ctx, app)
			if err != nil {
				return nil, err
			}

			updateIBCClients(ctx, app)

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
func updateIBCClients(ctx sdk.Context, app *App) {
	ctx.Logger().Info("Updating IBC AllowedClients.")
	params := app.IBCKeeper.ClientKeeper.GetParams(ctx)
	params.AllowedClients = append(params.AllowedClients, exported.Localhost)
	app.IBCKeeper.ClientKeeper.SetParams(ctx, params)
	ctx.Logger().Info("Done updating IBC AllowedClients.")
}

// migrateBaseappParams migrates to new ConsensusParamsKeeper
// TODO: Remove with the umber handlers.
func migrateBaseappParams(ctx sdk.Context, app *App) error {
	ctx.Logger().Info("Migrating consensus params.")
	legacyBaseAppSubspace := app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
	err := baseapp.MigrateParams(ctx, legacyBaseAppSubspace, app.ConsensusParamsKeeper.ParamsStore)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Unable to migrate legacy params to ConsensusParamsKeeper, error: %s.", err))
		return err
	}
	ctx.Logger().Info("Done migrating consensus params.")
	return nil
}

// migrateBankParams migrates the bank params from the params module to the bank module's state.
// The SDK has this as part of their bank v4 migration, but we're already on v4, so that one
// won't run on its own. This is the only part of that migration that we still need to have
// done, and this brings us in-line with format the bank state on v4.
// TODO: delete with the umber handlers.
func migrateBankParams(ctx sdk.Context, app *App) error {
	ctx.Logger().Info("Migrating bank params.")

	bankParams := banktypes.Params{DefaultSendEnabled: true}
	err := app.BankKeeper.SetParams(ctx, bankParams)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Unable to migrate bank params, error: %s.", err))
		return fmt.Errorf("could not store new bank params: %w", err)
	}

	ctx.Logger().Info("Done migrating bank params.")
	return nil
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
	app.AttributeKeeper.SetParams(ctx, attributetypes.Params{MaxValueLength: maxValueLength})
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
		params.MaxTotalSupply = maxTotalSupply //nolint:staticcheck
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
		var maxSupply sdkmath.Int
		markerParamSpace.Get(ctx, markertypes.ParamStoreKeyMaxSupply, &maxSupply)
		params.MaxSupply = maxSupply
	}

	app.MarkerKeeper.SetParams(ctx, params)

	ctx.Logger().Info("Done migrating marker params.")
}

// migrateAttributeParams migrates to new Metadata Os Locator Params store
// TODO: Remove with the umber handlers.
func migrateMetadataOSLocatorParams(ctx sdk.Context, app *App) {
	ctx.Logger().Info("Migrating metadata os locator params.")
	params := metadatatypes.DefaultOSLocatorParams()
	app.MetadataKeeper.SetOSLocatorParams(ctx, params)
	ctx.Logger().Info("Done migrating metadata os locator params.")
}

// migrateNameParams migrates to new Name Params store
// TODO: Remove with the umber handlers.
func migrateNameParams(ctx sdk.Context, app *App) {
	ctx.Logger().Info("Migrating name params.")
	nameParamSpace := app.ParamsKeeper.Subspace(nametypes.ModuleName).WithKeyTable(nametypes.ParamKeyTable())

	params := nametypes.DefaultParams()

	// TODO: all param keys from types/params with the umber handlers.
	if nameParamSpace.Has(ctx, nametypes.ParamStoreKeyMaxNameLevels) {
		nameParamSpace.Get(ctx, nametypes.ParamStoreKeyMaxNameLevels, &params.MaxNameLevels)
	}

	if nameParamSpace.Has(ctx, nametypes.ParamStoreKeyMaxSegmentLength) {
		nameParamSpace.Get(ctx, nametypes.ParamStoreKeyMaxSegmentLength, &params.MaxSegmentLength)
	}

	if nameParamSpace.Has(ctx, nametypes.ParamStoreKeyMinSegmentLength) {
		nameParamSpace.Get(ctx, nametypes.ParamStoreKeyMinSegmentLength, &params.MinSegmentLength)
	}

	if nameParamSpace.Has(ctx, nametypes.ParamStoreKeyAllowUnrestrictedNames) {
		nameParamSpace.Get(ctx, nametypes.ParamStoreKeyAllowUnrestrictedNames, &params.AllowUnrestrictedNames)
	}
	app.NameKeeper.SetParams(ctx, params)

	ctx.Logger().Info("Done migrating name params.")
}

// migrateMsgFeesParams migrates to new MsgFees Params store
// TODO: Remove with the umber handlers.
func migrateMsgFeesParams(ctx sdk.Context, app *App) {
	ctx.Logger().Info("Migrating msgfees params.")
	msgFeesParamSpace := app.ParamsKeeper.Subspace(msgfeestypes.ModuleName).WithKeyTable(msgfeestypes.ParamKeyTable())

	// TODO: all param keys from types/params with the umber handlers.
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

// migrateIbcHooksParams migrates existing ibchooks parameters from paramSpace to a direct KVStore.
// TODO: Remove with the umber handlers.
func migrateIbcHooksParams(ctx sdk.Context, app *App) {
	ctx.Logger().Info("Migrating ibchooks params.")
	ibcHooksParamSpace := app.ParamsKeeper.Subspace(ibchookstypes.ModuleName).WithKeyTable(ibchookstypes.ParamKeyTable())

	params := ibchookstypes.DefaultParams()

	// TODO: all param keys from types/params with the umber handlers.
	var allowlist []string
	if ibcHooksParamSpace.Has(ctx, ibchookstypes.KeyAsyncAckAllowList) {
		ibcHooksParamSpace.Get(ctx, ibchookstypes.KeyAsyncAckAllowList, &allowlist)
	}
	params.AllowedAsyncAckContracts = allowlist

	app.IBCHooksKeeper.SetParams(ctx, params)

	ctx.Logger().Info("Done migrating ibchooks params.")
}

// setNewGovParamsMainnet updates the newly added gov params fields to have the values we want for mainnet.
// TODO: Remove with the umber handlers.
func setNewGovParamsMainnet(ctx sdk.Context, app *App) error {
	expVP := time.Hour * 24
	params := govv1.Params{
		MinInitialDepositRatio:     "0.02",
		MinDepositRatio:            "0",
		ProposalCancelRatio:        "0.5",
		ProposalCancelDest:         "",
		ExpeditedVotingPeriod:      &expVP,
		ExpeditedThreshold:         "0.667",
		ExpeditedMinDeposit:        nil, // Will end up being the current MinDeposit value.
		BurnVoteQuorum:             false,
		BurnProposalDepositPrevote: true,
		BurnVoteVeto:               true,
	}
	return setNewGovParams(ctx, app, params, "mainnet")
}

// setNewGovParamsTestnet updates the newly added gov params fields to have the values we want for testnet.
// TODO: Remove with the umber handlers.
func setNewGovParamsTestnet(ctx sdk.Context, app *App) error {
	expVP := time.Minute * 5
	params := govv1.Params{
		MinInitialDepositRatio:     "0.00002",
		MinDepositRatio:            "0",
		ProposalCancelRatio:        "0",
		ProposalCancelDest:         "",
		ExpeditedVotingPeriod:      &expVP,
		ExpeditedThreshold:         "0.667",
		ExpeditedMinDeposit:        nil, // Will end up being the current MinDeposit value.
		BurnVoteQuorum:             false,
		BurnProposalDepositPrevote: false,
		BurnVoteVeto:               true,
	}
	return setNewGovParams(ctx, app, params, "testnet")
}

// setNewGovParams updates the gov params state to populate the new fields.
// Only the newly added fields are used from newParams.
// Fields that already existed will remain unchanged and are ignored in newParams.
// TODO: Remove with the umber handlers.
func setNewGovParams(ctx sdk.Context, app *App, newParams govv1.Params, chain string) error {
	ctx.Logger().Info(fmt.Sprintf("Setting new gov params for %s.", chain))

	params, err := app.GovKeeper.Params.Get(ctx)
	if err != nil {
		return fmt.Errorf("error getting gov params: %w", err)
	}

	params.MinInitialDepositRatio = sdkmath.LegacyMustNewDecFromStr(newParams.MinInitialDepositRatio).String()
	params.MinDepositRatio = sdkmath.LegacyMustNewDecFromStr(newParams.MinDepositRatio).String()
	params.ProposalCancelRatio = sdkmath.LegacyMustNewDecFromStr(newParams.ProposalCancelRatio).String()
	params.ProposalCancelDest = newParams.ProposalCancelDest
	params.ExpeditedVotingPeriod = newParams.ExpeditedVotingPeriod
	params.ExpeditedThreshold = sdkmath.LegacyMustNewDecFromStr(newParams.ExpeditedThreshold).String()
	if len(newParams.ExpeditedMinDeposit) != 0 {
		params.ExpeditedMinDeposit = newParams.ExpeditedMinDeposit
	} else {
		params.ExpeditedMinDeposit = params.MinDeposit
	}

	params.BurnVoteQuorum = newParams.BurnVoteQuorum
	params.BurnProposalDepositPrevote = newParams.BurnProposalDepositPrevote
	params.BurnVoteVeto = newParams.BurnVoteVeto

	err = app.GovKeeper.Params.Set(ctx, params)
	if err != nil {
		return fmt.Errorf("error setting updated gov params: %w", err)
	}

	ctx.Logger().Info(fmt.Sprintf("Done setting new gov params for %s.", chain))
	return nil
}

// addScopeNavsWithHeight sets net asset values with heights for markers
// TODO: Remove with the umber handlers.
func addScopeNavsWithHeight(ctx sdk.Context, app *App, navsWithHeight []NetAssetValueWithHeight) {
	ctx.Logger().Info("Adding scope net asset values with heights.")

	totalAdded := 0
	for _, navWithHeight := range navsWithHeight {
		uid, err := uuid.Parse(navWithHeight.ScopeUUID)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("invalid uuid %v : %v", navWithHeight.ScopeUUID, err))
			continue
		}
		scopeAddr := metadatatypes.ScopeMetadataAddress(uid)
		_, found := app.MetadataKeeper.GetScope(ctx, scopeAddr)
		if !found {
			ctx.Logger().Error(fmt.Sprintf("unable to find scope %v", navWithHeight.ScopeUUID))
			continue
		}

		if err := app.MetadataKeeper.SetNetAssetValueWithBlockHeight(ctx, scopeAddr, navWithHeight.NetAssetValue, "upgrade_handler", navWithHeight.Height); err != nil {
			ctx.Logger().Error(fmt.Sprintf("unable to set net asset value with height %v at height %d: %v", navWithHeight.NetAssetValue, navWithHeight.Height, err))
		}
		totalAdded++
	}

	ctx.Logger().Info(fmt.Sprintf("Done adding a total of %v scope net asset values with heights.", totalAdded))
}
