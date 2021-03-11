package app

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

var (
	noopHandler = func(ctx sdk.Context, plan upgradetypes.Plan) {
		ctx.Logger().Info("Applying no-op upgrade plan for release " + plan.Name)
	}
)

type appUpgradeHandler = func(*App, sdk.Context, upgradetypes.Plan)

type appUpgrade struct {
	Added   []string
	Deleted []string
	Renamed []storetypes.StoreRename
	Handler appUpgradeHandler
}

var handlers = map[string]appUpgrade{
	"v0.2.0": {},
	"v0.2.1": {
		Handler: func(app *App, ctx sdk.Context, plan upgradetypes.Plan){
			existing := app.MarkerKeeper.GetParams(ctx)
			existing.UnrestrictedDenomRegex = markertypes.DefaultUnrestrictedDenomRegex
			app.MarkerKeeper.SetParams(ctx, existing)
		},
	},

	// TODO - Add new upgrade definitions here.
}

func CustomUpgradeStoreLoader(app *App, info storetypes.UpgradeInfo) baseapp.StoreLoader {
	// Register all explicit appUpgrades
	for name, upgrade := range handlers {
		// If the handler has been defined, add it here, otherwise, use no-op.
		var handler upgradetypes.UpgradeHandler
		if upgrade.Handler == nil {
			handler = noopHandler
		} else {
			handler = func(ctx sdk.Context, plan upgradetypes.Plan) {
				upgrade.Handler(app, ctx, plan)
			}
		}
		app.UpgradeKeeper.SetUpgradeHandler(name, handler)
	}

	for name, upgrade := range handlers {
		// If the plan is executing this block, set the store locator to create any
		// missing modules, delete unused modules, or rename any keys required in the plan.
		if info.Name == name && !app.UpgradeKeeper.IsSkipHeight(info.Height) {
			storeUpgrades := storetypes.StoreUpgrades{
				Added:   upgrade.Added,
				Renamed: upgrade.Renamed,
				Deleted: upgrade.Deleted,
			}
			app.Logger().Info("Store upgrades", "plan", name, "height", info.Height, "upgrade", upgrade)
			return upgradetypes.UpgradeStoreLoader(info.Height, &storeUpgrades)
		}
	}
	return nil
}
