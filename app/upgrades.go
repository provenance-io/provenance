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

type appUpgradeHandler = func(*App, sdk.Context, upgradetypes.Plan) (module.VersionMap, error)

type appUpgrade struct {
	Added   []string
	Deleted []string
	Renamed []storetypes.StoreRename
	Handler appUpgradeHandler
}

var handlers = map[string]appUpgrade{
	"lava": {}, // upgrade for 1.10.0
	"mango": {
		Handler: func(app *App, ctx sdk.Context, plan upgradetypes.Plan) (module.VersionMap, error) {
			params := app.MsgFeesKeeper.GetParams(ctx)
			app.MsgFeesKeeper.SetParams(ctx, params)
			versionMap := app.UpgradeKeeper.GetModuleVersionMap(ctx)
			return versionMap, nil
		},
	}, // upgrade for 1.11.x
	// TODO - Add new upgrade definitions here.
}

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
				vM, err := ref.Handler(app, ctx, plan)
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

// CustomUpgradeStoreLoader provides upgrade handlers for store and application module upgrades at specified versions
func CustomUpgradeStoreLoader(app *App, info storetypes.UpgradeInfo) baseapp.StoreLoader {
	// Current upgrade info is empty or we are at the wrong height, skip this.
	if info.Name == "" || info.Height-1 != app.LastBlockHeight() {
		return nil
	}
	// Find the upgrade handler that matches this currently executing upgrade.
	for name, upgrade := range handlers {
		// If the plan is executing this block, set the store locator to create any
		// missing modules, delete unused modules, or rename any keys required in the plan.
		if info.Name == name && !app.UpgradeKeeper.IsSkipHeight(info.Height) {
			storeUpgrades := storetypes.StoreUpgrades{
				Added:   upgrade.Added,
				Renamed: upgrade.Renamed,
				Deleted: upgrade.Deleted,
			}

			if isEmptyUpgrade(storeUpgrades) {
				app.Logger().Info("No store upgrades required",
					"plan", name,
					"height", info.Height,
				)
				return nil
			}

			app.Logger().Info("Store upgrades",
				"plan", name,
				"height", info.Height,
				"upgrade.added", upgrade.Added,
				"upgrade.deleted", upgrade.Deleted,
				"upgrade.renamed", upgrade.Renamed,
			)
			return upgradetypes.UpgradeStoreLoader(info.Height, &storeUpgrades)
		}
	}
	return nil
}

func isEmptyUpgrade(upgrades storetypes.StoreUpgrades) bool {
	return len(upgrades.Renamed) == 0 && len(upgrades.Deleted) == 0 && len(upgrades.Added) == 0
}
