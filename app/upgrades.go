package app

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v6/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"

	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/hold"
	ibchookstypes "github.com/provenance-io/provenance/x/ibchooks/types"
	ibcratelimit "github.com/provenance-io/provenance/x/ibcratelimit"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	oracletypes "github.com/provenance-io/provenance/x/oracle/types"
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
	"saffron-rc1": { // upgrade for v1.17.0-rc1
		Handler: func(ctx sdk.Context, app *App, vm module.VersionMap) (module.VersionMap, error) {
			var err error
			vm, err = runModuleMigrations(ctx, app, vm)
			if err != nil {
				return nil, err
			}
			// set ibchoooks defaults (no allowed async contracts)
			app.IBCHooksKeeper.SetParams(ctx, ibchookstypes.DefaultParams())

			removeInactiveValidatorDelegations(ctx, app)
			setupICQ(ctx, app)
			updateMaxSupply(ctx, app)
			setExchangeParams(ctx, app)

			return vm, nil
		},
		Added: []string{icqtypes.ModuleName, oracletypes.ModuleName, ibchookstypes.StoreKey, hold.ModuleName, exchange.ModuleName},
	},
	"saffron-rc2": { // upgrade for v1.17.0-rc2
		Handler: func(ctx sdk.Context, app *App, vm module.VersionMap) (module.VersionMap, error) {
			var err error
			vm, err = runModuleMigrations(ctx, app, vm)
			if err != nil {
				return nil, err
			}

			updateIbcMarkerDenomMetadata(ctx, app)

			return vm, nil
		},
	},
	"saffron-rc3": { // upgrade for v1.17.0-rc3
		Handler: func(ctx sdk.Context, app *App, vm module.VersionMap) (module.VersionMap, error) {
			var err error
			vm, err = runModuleMigrations(ctx, app, vm)
			if err != nil {
				return nil, err
			}

			updateIbcMarkerDenomMetadata(ctx, app)

			return vm, nil
		},
	},
	"saffron": { // upgrade for v1.17.0,
		Handler: func(ctx sdk.Context, app *App, vm module.VersionMap) (module.VersionMap, error) {
			var err error
			vm, err = runModuleMigrations(ctx, app, vm)
			if err != nil {
				return nil, err
			}

			// set ibchoooks defaults (no allowed async contracts)
			app.IBCHooksKeeper.SetParams(ctx, ibchookstypes.DefaultParams())

			removeInactiveValidatorDelegations(ctx, app)
			setupICQ(ctx, app)
			updateMaxSupply(ctx, app)

			addMarkerNavs(ctx, app, GetPioMainnet1DenomToNav())

			setExchangeParams(ctx, app)
			updateIbcMarkerDenomMetadata(ctx, app)

			return vm, nil
		},
		Added: []string{icqtypes.ModuleName, oracletypes.ModuleName, ibchookstypes.StoreKey, hold.ModuleName, exchange.ModuleName},
	},
	"tourmaline-rc1": { // upgrade for v1.18.0-rc1
		Added: []string{ibcratelimit.ModuleName},
		Handler: func(ctx sdk.Context, app *App, vm module.VersionMap) (module.VersionMap, error) {
			var err error
			vm, err = runModuleMigrations(ctx, app, vm)
			if err != nil {
				return nil, err
			}

			removeInactiveValidatorDelegations(ctx, app)
			convertNavUnits(ctx, app)

			return vm, nil
		},
	},
	"tourmaline": { // upgrade for v1.18.0
		Added: []string{ibcratelimit.ModuleName},
		Handler: func(ctx sdk.Context, app *App, vm module.VersionMap) (module.VersionMap, error) {
			var err error
			vm, err = runModuleMigrations(ctx, app, vm)
			if err != nil {
				return nil, err
			}

			removeInactiveValidatorDelegations(ctx, app)
			convertNavUnits(ctx, app)

			// This isn't in an rc because it was handled via gov prop for testnet.
			updateMsgFeesNhashPerMil(ctx, app)

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
			handler = func(ctx sdk.Context, plan upgradetypes.Plan, versionMap module.VersionMap) (module.VersionMap, error) {
				ctx.Logger().Info(fmt.Sprintf("Applying no-op upgrade to %q", plan.Name))
				return versionMap, nil
			}
		} else {
			ref := upgrade
			handler = func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
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
	unbondingTimeParam := app.StakingKeeper.GetParams(ctx).UnbondingTime
	ctx.Logger().Info(fmt.Sprintf("removing all delegations from validators that have been inactive (unbonded) for %d days", int64(unbondingTimeParam.Hours()/24)))
	removalCount := 0
	validators := app.StakingKeeper.GetAllValidators(ctx)
	for _, validator := range validators {
		if validator.IsUnbonded() {
			inactiveDuration := ctx.BlockTime().Sub(validator.UnbondingTime)
			if inactiveDuration >= unbondingTimeParam {
				ctx.Logger().Info(fmt.Sprintf("validator %v has been inactive (unbonded) for %d days and will be removed", validator.OperatorAddress, int64(inactiveDuration.Hours()/24)))
				valAddress, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
				if err != nil {
					ctx.Logger().Error(fmt.Sprintf("invalid operator address: %s: %v", validator.OperatorAddress, err))
					continue
				}
				delegations := app.StakingKeeper.GetValidatorDelegations(ctx, valAddress)
				for _, delegation := range delegations {
					ctx.Logger().Info(fmt.Sprintf("undelegate delegator %v from validator %v of all shares (%v)", delegation.DelegatorAddress, validator.OperatorAddress, delegation.GetShares()))
					_, err = app.StakingKeeper.Undelegate(ctx, delegation.GetDelegatorAddr(), valAddress, delegation.GetShares())
					if err != nil {
						ctx.Logger().Error(fmt.Sprintf("failed to undelegate delegator %s from validator %s: %v", delegation.GetDelegatorAddr().String(), valAddress.String(), err))
						continue
					}
				}
				removalCount++
			}
		}
	}
	ctx.Logger().Info(fmt.Sprintf("a total of %d inactive (unbonded) validators have had all their delegators removed", removalCount))
}

// setupICQ sets the correct default values for ICQKeeper.
// TODO: Remove with the saffron handlers.
func setupICQ(ctx sdk.Context, app *App) {
	ctx.Logger().Info("Updating ICQ params")
	app.ICQKeeper.SetParams(ctx, icqtypes.NewParams(true, []string{"/provenance.oracle.v1.Query/Oracle"}))
	ctx.Logger().Info("Done updating ICQ params")
}

// updateMaxSupply sets the value of max supply to the current value of MaxTotalSupply.
// TODO: Remove with the saffron handlers.
func updateMaxSupply(ctx sdk.Context, app *App) {
	ctx.Logger().Info("Updating MaxSupply marker param")
	params := app.MarkerKeeper.GetParams(ctx)
	//nolint:staticcheck // Populate new param with deprecated param
	params.MaxSupply = math.NewIntFromUint64(params.MaxTotalSupply)
	app.MarkerKeeper.SetParams(ctx, params)
	ctx.Logger().Info("Done updating MaxSupply marker param")
}

// addMarkerNavs adds navs to existing markers
// TODO: Remove with the saffron handlers.
func addMarkerNavs(ctx sdk.Context, app *App, denomToNav map[string]markertypes.NetAssetValue) {
	ctx.Logger().Info("Adding marker net asset values")
	for denom, nav := range denomToNav {
		marker, err := app.MarkerKeeper.GetMarkerByDenom(ctx, denom)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("unable to get marker %v: %v", denom, err))
			continue
		}
		if err := app.MarkerKeeper.AddSetNetAssetValues(ctx, marker, []markertypes.NetAssetValue{nav}, "upgrade_handler"); err != nil {
			ctx.Logger().Error(fmt.Sprintf("unable to set net asset value %v: %v", nav, err))
		}
	}
	ctx.Logger().Info("Done adding marker net asset values")
}

// setExchangeParams sets exchange module's params to the defaults.
// TODO: Remove with the saffron handlers.
func setExchangeParams(ctx sdk.Context, app *App) {
	ctx.Logger().Info("Ensuring exchange module params are set.")
	params := app.ExchangeKeeper.GetParams(ctx)
	if params != nil {
		ctx.Logger().Info("Exchange module params are already defined.")
	} else {
		params = exchange.DefaultParams()
		ctx.Logger().Info("Setting exchange module params to defaults.")
		app.ExchangeKeeper.SetParams(ctx, params)
	}
	ctx.Logger().Info("Done ensuring exchange module params are set.")
}

// updateIbcMarkerDenomMetadata iterates markers and creates denom metadata for ibc markers
// TODO: Remove with the saffron handlers.
func updateIbcMarkerDenomMetadata(ctx sdk.Context, app *App) {
	ctx.Logger().Info("Updating ibc marker denom metadata")
	app.MarkerKeeper.IterateMarkers(ctx, func(record markertypes.MarkerAccountI) bool {
		if !strings.HasPrefix(record.GetDenom(), "ibc/") {
			return false
		}

		hash, err := transfertypes.ParseHexHash(strings.TrimPrefix(record.GetDenom(), "ibc/"))
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("invalid denom trace hash: %s, error: %s", hash.String(), err))
			return false
		}
		denomTrace, found := app.TransferKeeper.GetDenomTrace(ctx, hash)
		if !found {
			ctx.Logger().Error(fmt.Sprintf("trace not found: %s, error: %s", hash.String(), err))
			return false
		}

		parts := strings.Split(denomTrace.Path, "/")
		if len(parts) == 2 && parts[0] == "transfer" {
			ctx.Logger().Info(fmt.Sprintf("Adding metadata to %s", record.GetDenom()))
			chainID := app.Ics20MarkerHooks.GetChainID(ctx, parts[0], parts[1], app.IBCKeeper)
			markerMetadata := banktypes.Metadata{
				Base:        record.GetDenom(),
				Name:        chainID + "/" + denomTrace.BaseDenom,
				Display:     chainID + "/" + denomTrace.BaseDenom,
				Description: denomTrace.BaseDenom + " from " + chainID,
			}
			app.BankKeeper.SetDenomMetaData(ctx, markerMetadata)
		}

		return false
	})
	ctx.Logger().Info("Done updating ibc marker denom metadata")
}

// convertNavUnits iterates all the net asset values and updates their units if they are using usd.
func convertNavUnits(ctx sdk.Context, app *App) {
	ctx.Logger().Info("Converting NAV units")
	err := app.MarkerKeeper.IterateAllNetAssetValues(ctx, func(markerAddr sdk.AccAddress, nav markertypes.NetAssetValue) (stop bool) {
		if nav.Price.Denom == markertypes.UsdDenom {
			nav.Price.Amount = nav.Price.Amount.Mul(math.NewInt(10))
			marker, err := app.MarkerKeeper.GetMarker(ctx, markerAddr)
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("Unable to get marker for address: %s, error: %s", markerAddr, err))
				return false
			}
			err = app.MarkerKeeper.SetNetAssetValue(ctx, marker, nav, "upgrade")
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("Unable to set net asset value for marker: %s, error: %s", markerAddr, err))
				return false
			}
		}
		return false
	})
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Unable to iterate all net asset values error: %s", err))
	}
}

// updateMsgFeesNhashPerMil updates the MsgFees Params to set the NhashPerUsdMil to 40,000,000.
func updateMsgFeesNhashPerMil(ctx sdk.Context, app *App) {
	var newVal uint64 = 40_000_000
	ctx.Logger().Info(fmt.Sprintf("Setting MsgFees Params NhashPerUsdMil to %d.", newVal))
	params := app.MsgFeesKeeper.GetParams(ctx)
	params.NhashPerUsdMil = newVal
	app.MsgFeesKeeper.SetParams(ctx, params)
	ctx.Logger().Info("Done setting MsgFees Params NhashPerUsdMil.")
}
