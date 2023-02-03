package app

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	ica "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts"
	icacontrollertypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"
	rewardtypes "github.com/provenance-io/provenance/x/reward/types"
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
	"ochre-rc1": { // upgrade for 1.13.0-rc3
		Added: []string{group.ModuleName, rewardtypes.ModuleName, icacontrollertypes.StoreKey, icahosttypes.StoreKey},
		Handler: func(ctx sdk.Context, app *App, plan upgradetypes.Plan) (module.VersionMap, error) {
			versionMap := app.UpgradeKeeper.GetModuleVersionMap(ctx)
			UpgradeICA(ctx, app, &versionMap)
			ctx.Logger().Info("Starting migrations. This may take a significant amount of time to complete. Do not restart node.")
			return app.mm.RunMigrations(ctx, app.configurator, versionMap)
		},
	},
	"ochre-rc2": { // upgrade for 1.13.0-rc5
		Handler: func(ctx sdk.Context, app *App, plan upgradetypes.Plan) (module.VersionMap, error) {
			versionMap := app.UpgradeKeeper.GetModuleVersionMap(ctx)

			// We need to run Migrate3_V046_4_To_V046_5 here because testnet already upgraded to v0.46.x.
			// But we don't need to run it in the ochre upgrade plan because mainnet hasn't upgraded to v0.46.x yet, so it doesn't need fixing.
			bankBaseKeeper, ok := app.BankKeeper.(*bankkeeper.BaseKeeper)
			if !ok {
				return versionMap, fmt.Errorf("could not cast app.BankKeeper (type bankkeeper.Keeper) to bankkeeper.BaseKeeper")
			}
			bankMigrator := bankkeeper.NewMigrator(*bankBaseKeeper)
			err := bankMigrator.Migrate3_V046_4_To_V046_5(ctx)
			if err != nil {
				return versionMap, err
			}

			ctx.Logger().Info("Starting migrations. This may take a significant amount of time to complete. Do not restart node.")
			return app.mm.RunMigrations(ctx, app.configurator, versionMap)
		},
	},
	"ochre": { // upgrade for v1.13.0
		Added: []string{group.ModuleName, rewardtypes.ModuleName, icacontrollertypes.StoreKey, icahosttypes.StoreKey},
		Handler: func(ctx sdk.Context, app *App, plan upgradetypes.Plan) (module.VersionMap, error) {
			versionMap := app.UpgradeKeeper.GetModuleVersionMap(ctx)

			// This params fix was handled in testnet via gov prop. For mainnet, we want it done in here though.
			params := app.MsgFeesKeeper.GetParams(ctx)
			app.MsgFeesKeeper.SetParams(ctx, params)

			UpgradeICA(ctx, app, &versionMap)
			ctx.Logger().Info("Starting migrations. This may take a significant amount of time to complete. Do not restart node.")
			return app.mm.RunMigrations(ctx, app.configurator, versionMap)
		},
	},
	"paua-rc1": { // upgrade for v1.14.0-rc1
		Added: []string{quarantine.ModuleName, sanction.ModuleName},
		Handler: func(ctx sdk.Context, app *App, plan upgradetypes.Plan) (module.VersionMap, error) {
			ctx.Logger().Info("Starting migrations. This may take a significant amount of time to complete. Do not restart node.")
			IncreaseMaxCommissions(ctx, app)
			IncreaseMaxGas(ctx, app)
			if err := SetSanctionParams(ctx, app); err != nil {
				return nil, err
			}
			versionMap := app.UpgradeKeeper.GetModuleVersionMap(ctx)
			return app.mm.RunMigrations(ctx, app.configurator, versionMap)
		},
	},
	"paua": { // upgrade for v1.14.0
		Added: []string{quarantine.ModuleName, sanction.ModuleName},
		Handler: func(ctx sdk.Context, app *App, plan upgradetypes.Plan) (module.VersionMap, error) {
			ctx.Logger().Info("Starting migrations. This may take a significant amount of time to complete. Do not restart node.")
			IncreaseMaxCommissions(ctx, app)
			IncreaseMaxGas(ctx, app)
			if err := SetSanctionParams(ctx, app); err != nil {
				return nil, err
			}
			versionMap := app.UpgradeKeeper.GetModuleVersionMap(ctx)
			return app.mm.RunMigrations(ctx, app.configurator, versionMap)
		},
	},
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

func UpgradeICA(ctx sdk.Context, app *App, versionMap *module.VersionMap) {
	app.Logger().Info("Initializing ICA")

	// Set the consensus version so InitGenesis is not ran
	// We are configuring the module here
	(*versionMap)[icatypes.ModuleName] = app.mm.Modules[icatypes.ModuleName].ConsensusVersion()

	// create ICS27 Controller submodule params
	controllerParams := icacontrollertypes.Params{
		ControllerEnabled: false,
	}

	// create ICS27 Host submodule params
	hostParams := icahosttypes.Params{
		HostEnabled:   true,
		AllowMessages: []string{},
	}

	// initialize ICS27 module
	icamodule, correctTypecast := app.mm.Modules[icatypes.ModuleName].(ica.AppModule)
	if !correctTypecast {
		panic("mm.Modules[icatypes.ModuleName] is not of type ica.AppModule")
	}
	icamodule.InitModule(ctx, controllerParams, hostParams)
	app.Logger().Info("Finished initializing ICA")
}

func IncreaseMaxCommissions(ctx sdk.Context, app *App) {
	minMaxCom := sdk.OneDec()
	validators := app.StakingKeeper.GetAllValidators(ctx)
	ctx.Logger().Info("Increasing all validator's max commission to 100%", "count", len(validators))
	for _, validator := range validators {
		validator.Commission.MaxRate = minMaxCom
		app.StakingKeeper.SetValidator(ctx, validator)
	}
}

func IncreaseMaxGas(ctx sdk.Context, app *App) {
	ctx.Logger().Info("Increasing max gas per block to 120,000,000")
	params := app.GetConsensusParams(ctx)
	params.Block.MaxGas = 120_000_000
	app.StoreConsensusParams(ctx, params)
}

func SetSanctionParams(ctx sdk.Context, app *App) error {
	ctx.Logger().Info("Setting sanction params")
	params := &sanction.Params{
		ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("nhash", 1_000_000_000_000_000)),
		ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("nhash", 1_000_000_000_000_000)),
	}
	err := app.SanctionKeeper.SetParams(ctx, params)
	if err != nil {
		return fmt.Errorf("could not set sanction params: %w", err)
	}
	return nil
}
