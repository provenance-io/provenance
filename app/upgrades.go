package app

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

var (
	noopHandler = func(ctx sdk.Context, plan upgradetypes.Plan, versionMap module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Applying no-op upgrade plan for release " + plan.Name)
		return versionMap, nil
	}
)

type appUpgradeHandler = func(sdk.Context, *App, upgradetypes.Plan) (module.VersionMap, error)

// appUpgrade is an internal structure for defining all things for an upgrade.
type appUpgrade struct {
	Added   []string
	Deleted []string
	Renamed []storetypes.StoreRename
	Handler appUpgradeHandler
}

var handlers = map[string]appUpgrade{
	"quicksilver-rc1": { // upgrade for v1.15.0-rc2
		Handler: func(ctx sdk.Context, app *App, plan upgradetypes.Plan) (module.VersionMap, error) {
			app.MarkerKeeper.RemoveIsSendEnabledEntries(ctx)
			app.AttributeKeeper.PopulateAddressAttributeNameTable(ctx)
			versionMap := app.UpgradeKeeper.GetModuleVersionMap(ctx)
			ctx.Logger().Info("Starting migrations. This may take a significant amount of time to complete. Do not restart node.")
			return app.mm.RunMigrations(ctx, app.configurator, versionMap)
		},
	},
	"quicksilver-rc2": {}, // upgrade for v1.15.0-rc3
	"quicksilver": { // upgrade for v1.15.0
		Handler: func(ctx sdk.Context, app *App, plan upgradetypes.Plan) (module.VersionMap, error) {
			app.MarkerKeeper.RemoveIsSendEnabledEntries(ctx)
			app.AttributeKeeper.PopulateAddressAttributeNameTable(ctx)
			versionMap := app.UpgradeKeeper.GetModuleVersionMap(ctx)
			ctx.Logger().Info("Starting migrations. This may take a significant amount of time to complete. Do not restart node.")
			return app.mm.RunMigrations(ctx, app.configurator, versionMap)
		},
	},
	"rust-rc1": {}, // upgrade for v1.16.0-rc1,
	"rust":     {}, // upgrade for v1.16.0,
	// TODO - Add new upgrade definitions here.
}

// InstallCustomUpgradeHandlers sets upgrade handlers for all entries in the handlers map.
func InstallCustomUpgradeHandlers(app *App) {
	// Register all explicit appUpgrades
	for name, upgrade := range handlers {
		// If the handler has been defined, add it here, otherwise, use no-op.
		var handler upgradetypes.UpgradeHandler
		if upgrade.Handler == nil {
			handler = noopHandler
		} else {
			ref := upgrade
			handler = func(ctx sdk.Context, plan upgradetypes.Plan, versionMap module.VersionMap) (module.VersionMap, error) {
				vM, err := ref.Handler(ctx, app, plan)
				if err != nil {
					ctx.Logger().Info(fmt.Sprintf("Failed to upgrade to: %s with err: %v", plan.Name, err))
				} else {
					ctx.Logger().Info(fmt.Sprintf("Successfully upgraded to: %s with version map: %v", plan.Name, vM))
				}
				return vM, err
			}
		}
		app.UpgradeKeeper.SetUpgradeHandler(name, handler)
	}
}

// GetUpgradeStoreLoader creates an StoreLoader for use in an upgrade.
// Returns nil if no upgrade info is found or the upgrade doesn't need a store loader.
func GetUpgradeStoreLoader(app *App, info upgradetypes.Plan) baseapp.StoreLoader {
	upgrade, found := handlers[info.Name]
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
func runModuleMigrations(ctx sdk.Context, app *App) (module.VersionMap, error) {
	// Even if this function is no longer called, do not delete it. Keep it around for the next time it's needed.m
	ctx.Logger().Info("Starting module migrations. This may take a significant amount of time to complete. Do not restart node.")
	versionMap := app.UpgradeKeeper.GetModuleVersionMap(ctx)
	var err error
	versionMap, err = app.mm.RunMigrations(ctx, app.configurator, versionMap)
	if err != nil {
		ctx.Logger().Error("Module migrations encountered an error.", "error", err)
		return nil, err
	}
	ctx.Logger().Info("Module migrations completed.")
	return versionMap, nil
}

// Create a use of runModuleMigrations so that the linter neither complains about it not being used,
// nor complains about a nolint:unused directive that isn't needed because the function is used.
var _ = runModuleMigrations
