package app

import (
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v2/modules/core/03-connection/types"
	"github.com/cosmos/ibc-go/v2/modules/core/exported"
	ibcctmtypes "github.com/cosmos/ibc-go/v2/modules/light-clients/07-tendermint/types"
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

type moduleUpgradeVersion struct {
	ModuleName  string
	FromVersion uint64
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

			orderedMigration := []moduleUpgradeVersion{

				// x/auth’s migrations depends on x/bank (delegations etc).
				// This causes cases where running auth migration before bank’s would produce
				// a different app state hash than running bank’s before auth’s
				{"bank", 1},
				{"auth", 1},

				// order doesn't matter
				{"capability", 1},
				{"crisis", 1},
				{"distribution", 1},
				{"evidence", 1},
				{"genutil", 1},
				{"gov", 1},
				{"ibc", 1},
				{"mint", 1},
				{"params", 1},
				{"slashing", 1},
				{"staking", 1},
				{"transfer", 1},
				{"upgrade", 1},
				{"vesting", 1},

				// new modules that need to run init genesis
				{"authz", 0},
				{"feegrant", 0},

				// cosmwasm module
				{"wasm", 1},

				// provenance modules
				{"attribute", 1},
				{"marker", 1},
				{"metadata", 1},
				{"name", 1},
			}
			ctx.Logger().Info("NOTICE: Starting large migration on all modules for cosmos-sdk v0.44.0.  This will take a significant amount of time to complete.  Do not restart node.")
			return RunOrderedMigrations(app, ctx, orderedMigration)
		},
		Added: []string{authz.ModuleName, feegrant.ModuleName},
	},
	"green": {
		Handler: func(app *App, ctx sdk.Context, plan upgradetypes.Plan) (module.VersionMap, error) {
			app.IBCKeeper.ClientKeeper.IterateClients(ctx, func(clientId string, state exported.ClientState) bool {
				tc, ok := (state).(*ibcctmtypes.ClientState)
				if ok {
					tc.AllowUpdateAfterExpiry = true
					app.IBCKeeper.ClientKeeper.SetClientState(ctx, clientId, state)
				}
				return false
			})
			orderedMigration := []moduleUpgradeVersion{
				{"metadata", 2},
			}
			ctx.Logger().Info("NOTICE: Starting migrations. This may take a significant amount of time to complete. Do not restart node.")
			return RunOrderedMigrations(app, ctx, orderedMigration)
		},
	},
	// TODO - Add new upgrade definitions here.
}

// RunOrderedMigrations runs migrations in a defined order.
// NOTE: We needed to modify the behavior of the cosmos-sdk's migrations.  Order DOES matter in some cases and their migration uses a map and not a list.
// This does not guarantee order in the migration process. i.e., The x/bank module needs to run before the x/auth module for version 1 to 2
func RunOrderedMigrations(app *App, ctx sdk.Context, migrationOrder []moduleUpgradeVersion) (module.VersionMap, error) {
	ctx.Logger().Info(fmt.Sprintf("Starting all module migrations in order: %v", migrationOrder))
	updatedVersionMap := make(module.VersionMap)
	for _, moduleAndVersion := range migrationOrder {
		partialVersionMap := make(module.VersionMap)
		partialVersionMap[moduleAndVersion.ModuleName] = moduleAndVersion.FromVersion
		ctx.Logger().Info(fmt.Sprintf("Run migration on module %v from starting version %v", moduleAndVersion.ModuleName, partialVersionMap[moduleAndVersion.ModuleName]))
		mm := make(map[string]module.AppModule)
		mm[moduleAndVersion.ModuleName] = app.mm.Modules[moduleAndVersion.ModuleName]
		// create a special filtered module manager so we can control the library internal initialization process
		// that uses a non-deterministic map.  The following still has a map but with only one element at a time.
		mgr := module.Manager{
			Modules:            mm,
			OrderBeginBlockers: app.mm.OrderBeginBlockers,
			OrderEndBlockers:   app.mm.OrderEndBlockers,
			OrderExportGenesis: app.mm.OrderExportGenesis,
			OrderInitGenesis:   app.mm.OrderInitGenesis,
		}
		migratedVersionMap, err := mgr.RunMigrations(ctx, app.configurator, partialVersionMap)
		if err != nil {
			return nil, err
		}
		updatedVersionMap[moduleAndVersion.ModuleName] = migratedVersionMap[moduleAndVersion.ModuleName]
	}
	ctx.Logger().Info(fmt.Sprintf("Finished running all module migrations. Final versions: %v", updatedVersionMap))
	return updatedVersionMap, nil
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
