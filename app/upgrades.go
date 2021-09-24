package app

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/modules/core/03-connection/types"
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
	"eigengrau": {
		Handler: func(app *App, ctx sdk.Context, plan upgradetypes.Plan) (module.VersionMap, error) {
			panic("Upgrade height for eigengrau must be skipped.  Use `--unsafe-skip-upgrades <height>` flag to skip upgrade")
		},
	},
	"feldgrau": {
		Handler: func(app *App, ctx sdk.Context, plan upgradetypes.Plan) (module.VersionMap, error) {
			app.IBCKeeper.ConnectionKeeper.SetParams(ctx, ibcconnectiontypes.DefaultParams())

			nhashName := "Hash"
			nhashSymbol := "HASH"
			nhash, found := app.BankKeeper.GetDenomMetaData(ctx, "nhash")
			if found {
				nhash.Name = nhashName
				nhash.Symbol = nhashSymbol
			} else {
				nhash = banktypes.Metadata{
					Description: "Hash is the staking token of the Provenance Blockchain",
					Base:        "nhash",
					Display:     "hash",
					Name:        nhashName,
					Symbol:      nhashSymbol,
					DenomUnits: []*banktypes.DenomUnit{
						{
							Denom:    "nhash",
							Exponent: 0,
							Aliases:  []string{},
						},
						{
							Denom:    "hash",
							Exponent: 9,
							Aliases:  []string{},
						},
					},
				}
			}
			app.BankKeeper.SetDenomMetaData(ctx, nhash)

			if sdk.GetConfig().GetBech32AccountAddrPrefix() == AccountAddressPrefixMainNet {
				s, ok := app.ParamsKeeper.GetSubspace(wasmtypes.DefaultParamspace)
				if !ok {
					panic("could not get wasm module parameter configuration")
				}
				s.Set(ctx, wasmtypes.ParamStoreKeyUploadAccess, wasmtypes.AccessTypeNobody.With(sdk.AccAddress{}))
			}

			migrationOrder := []string{
				// x/auth’s migrations depends on x/bank (delegations etc).
				// This causes cases where running auth migration before bank’s would produce
				/// a different app state hash than running bank’s before auth’s
				"bank",
				"auth",

				"authz",
				"capability",
				"crisis",
				"distribution",
				"evidence",
				"feegrant",
				"genutil",
				"gov",
				"ibc",
				"mint",
				"params",
				"slashing",
				"staking",
				"transfer",
				"upgrade",
				"vesting",

				// cosmwasm module
				"wasm",

				// provenance modules
				"attribute",
				"marker",
				"metadata",
				"name",
			}

			ctx.Logger().Info("NOTICE: Starting large migration on all modules for cosmos-sdk v0.44.0.  This will take a significant amount of time to complete.  Do not restart node.")
			ctx.Logger().Info("Starting all module migrations in order: %v", migrationOrder)
			// NOTE: cosmos-sdk used a map to run migrations, but order does matter.  This will assure they are executed in order
			updatedVersionMap := make(module.VersionMap)
			for _, moduleName := range migrationOrder {
				partialVersionMap := make(module.VersionMap)
				if moduleName == "authz" || moduleName == "feegrant" {
					// we want new modules to execute init genesis
					partialVersionMap[moduleName] = 0
				} else {
					// modules that start at from version 1
					partialVersionMap[moduleName] = 1
				}
				ctx.Logger().Info("Run migration on module %s from starting version %v", moduleName, partialVersionMap[moduleName])
				versionMap, err := app.mm.RunMigrations(ctx, app.configurator, partialVersionMap)
				if err != nil {
					return nil, err
				}
				updatedVersionMap[moduleName] = versionMap[moduleName]
			}
			ctx.Logger().Info("Finished running all module migrations. Final versions: %v", updatedVersionMap)
			return updatedVersionMap, nil
		},
		Added: []string{authz.ModuleName, feegrant.ModuleName},
	},
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
				return ref.Handler(app, ctx, plan)
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
