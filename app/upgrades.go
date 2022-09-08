package app

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	ica "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts"
	icacontrollertypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"
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
	"mango": {
		Handler: func(app *App, ctx sdk.Context, plan upgradetypes.Plan) (module.VersionMap, error) {
			params := app.MsgFeesKeeper.GetParams(ctx)
			app.MsgFeesKeeper.SetParams(ctx, params)
			versionMap := app.UpgradeKeeper.GetModuleVersionMap(ctx)
			return app.mm.RunMigrations(ctx, app.configurator, versionMap)
		},
	}, // upgrade for 1.11.1
	"mango-rc4":      {}, // upgrade for 1.11.1-rc4
	"neoncarrot-rc1": {}, // upgrade for 1.12.0-rc1
	"ochre-rc1": {
		// TODO: Required for v1.13.x: Fill in Added with modules new to 1.13.x https://github.com/provenance-io/provenance/issues/1007
		Added: []string{icacontrollertypes.StoreKey, icahosttypes.StoreKey},
		Handler: func(app *App, ctx sdk.Context, plan upgradetypes.Plan) (module.VersionMap, error) {
			versionMap := app.UpgradeKeeper.GetModuleVersionMap(ctx)
			UpgradeICA(ctx, app, &versionMap)
			return app.mm.RunMigrations(ctx, app.configurator, versionMap)
		},
	}, // upgrade for 1.13.0-rc1
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
func CustomUpgradeStoreLoader(app *App, info upgradetypes.Plan) baseapp.StoreLoader {
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

func UpgradeICA(ctx sdk.Context, app *App, versionMap *module.VersionMap) {
	app.Logger().Info("Initializing ICA")

	// Set the consensus version so InitGenesis is not ran
	// We are configuring the module here
	(*versionMap)[icatypes.ModuleName] = app.mm.Modules[icatypes.ModuleName].ConsensusVersion()

	// create ICS27 Controller submodule params
	controllerParams := icacontrollertypes.Params{
		ControllerEnabled: true,
	}

	// create ICS27 Host submodule params
	// TODO Verify which messages we want to run on the host/Provenance chain
	hostParams := icahosttypes.Params{
		HostEnabled: true,
		AllowMessages: []string{
			"*",
		},
	}

	// initialize ICS27 module
	icamodule, correctTypecast := app.mm.Modules[icatypes.ModuleName].(ica.AppModule)
	if !correctTypecast {
		panic("mm.Modules[icatypes.ModuleName] is not of type ica.AppModule")
	}
	icamodule.InitModule(ctx, controllerParams, hostParams)
	app.Logger().Info("Finished initializing ICA")
}
