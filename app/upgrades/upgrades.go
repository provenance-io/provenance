package upgrades

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

var (
	noopHandler = func(ctx sdk.Context, plan upgradetypes.Plan) {
		ctx.Logger().Info("Applying no-op upgrade plan for release " + plan.Name)
	}
)

type appUpgrade struct {
	Added   []string
	Deleted []string
	Renamed []storetypes.StoreRename
	Handler upgradetypes.UpgradeHandler
}

var handlers = map[string]appUpgrade{
	"v0.1.8": {Added: []string{"metadata"}},

	// TODO - Add new upgrade definitions here.
}

func CustomUpgradeStoreLoader(keeper upgradekeeper.Keeper, info storetypes.UpgradeInfo) baseapp.StoreLoader {
	// Register explicit appUpgrades
	for name, upgrade := range handlers {
		// If the handler has been defined, add it here, otherwise, use no-op.
		var handler upgradetypes.UpgradeHandler
		if upgrade.Handler == nil {
			handler = noopHandler
		} else {
			handler = upgrade.Handler
		}
		keeper.SetUpgradeHandler(name, handler)

		// If the plan is executing this block, set the store locator to create any
		// missing modules, delete unused modules, or rename any keys required in the plan.
		if info.Name == name && !keeper.IsSkipHeight(info.Height) {
			storeUpgrades := storetypes.StoreUpgrades{
				Added:   upgrade.Added,
				Renamed: upgrade.Renamed,
				Deleted: upgrade.Deleted,
			}

			return upgradetypes.UpgradeStoreLoader(info.Height, &storeUpgrades)
		}
	}
	return nil
}
