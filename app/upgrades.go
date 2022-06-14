package app

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	attributetypes "github.com/provenance-io/provenance/x/attribute/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
	nametypes "github.com/provenance-io/provenance/x/name/types"
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
	"lava":  {}, // upgrade for 1.10.0
	"mango": {}, // upgrade for 1.11.1
	// TODO - Add new upgrade definitions here.
}

// AddMsgBasedFees adds pio specific msg based fees for BindName, AddMarker, AddAttribute, WriteScope, P8EMemorializeContract requests for v1.8.0 upgrade
func AddMsgBasedFees(app *App, ctx sdk.Context) error {
	ctx.Logger().Info("Adding a 10 hash (10,000,000,000 nhash) msg fee to MsgBindNameRequest (1/6)")
	if err := app.MsgFeesKeeper.SetMsgFee(ctx, msgfeestypes.NewMsgFee(sdk.MsgTypeURL(&nametypes.MsgBindNameRequest{}), sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)))); err != nil {
		return err
	}
	ctx.Logger().Info("Adding a 100 hash (100,000,000,000 nhash) msg fee to MsgAddMarkerRequest (2/6)")
	if err := app.MsgFeesKeeper.SetMsgFee(ctx, msgfeestypes.NewMsgFee(sdk.MsgTypeURL(&markertypes.MsgAddMarkerRequest{}), sdk.NewCoin("nhash", sdk.NewInt(100_000_000_000)))); err != nil {
		return err
	}
	ctx.Logger().Info("Adding a 10 hash (10,000,000,000 nhash) msg fee to MsgAddAttributeRequest (3/6)")
	if err := app.MsgFeesKeeper.SetMsgFee(ctx, msgfeestypes.NewMsgFee(sdk.MsgTypeURL(&attributetypes.MsgAddAttributeRequest{}), sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)))); err != nil {
		return err
	}
	ctx.Logger().Info("Adding a 10 hash (10,000,000,000 nhash) msg fee to MsgWriteScopeRequest (4/6)")
	if err := app.MsgFeesKeeper.SetMsgFee(ctx, msgfeestypes.NewMsgFee(sdk.MsgTypeURL(&metadatatypes.MsgWriteScopeRequest{}), sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)))); err != nil {
		return err
	}
	ctx.Logger().Info("Adding a 10 hash (10,000,000,000 nhash) msg fee to MsgP8EMemorializeContractRequest (5/6)")
	if err := app.MsgFeesKeeper.SetMsgFee(ctx, msgfeestypes.NewMsgFee(sdk.MsgTypeURL(&metadatatypes.MsgP8EMemorializeContractRequest{}), sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)))); err != nil {
		return err
	}
	ctx.Logger().Info("Adding a 100 hash (100,000,000,000 nhash) msg fee to MsgSubmitProposal (6/6)")
	if err := app.MsgFeesKeeper.SetMsgFee(ctx, msgfeestypes.NewMsgFee(sdk.MsgTypeURL(&govtypes.MsgSubmitProposal{}), sdk.NewCoin("nhash", sdk.NewInt(100_000_000_000)))); err != nil {
		return err
	}
	return nil
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
