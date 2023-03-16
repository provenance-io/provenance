package app

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	"github.com/cosmos/cosmos-sdk/x/sanction"
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
	"quicksilver-rc1": { // upgrade for v1.15.0-rc1
		Handler: func(ctx sdk.Context, app *App, plan upgradetypes.Plan) (module.VersionMap, error) {
			RemoveIsSendEnabledEntries(ctx, app)
			versionMap := app.UpgradeKeeper.GetModuleVersionMap(ctx)
			ctx.Logger().Info("Starting migrations. This may take a significant amount of time to complete. Do not restart node.")
			return app.mm.RunMigrations(ctx, app.configurator, versionMap)
		},
	},
	"paua": { // upgrade for v1.14.0
		Added: []string{quarantine.ModuleName, sanction.ModuleName},
		Handler: func(ctx sdk.Context, app *App, plan upgradetypes.Plan) (module.VersionMap, error) {
			ctx.Logger().Info("Starting migrations. This may take a significant amount of time to complete. Do not restart node.")
			versionMap, err := app.mm.RunMigrations(ctx, app.configurator, app.UpgradeKeeper.GetModuleVersionMap(ctx))
			if err != nil {
				return nil, err
			}
			err = SetSanctionParams(ctx, app) // Needs to happen after RunMigrations adds the sanction module.
			if err != nil {
				return nil, err
			}
			IncreaseMaxCommissions(ctx, app)
			// Skipping IncreaseMaxGas(ctx, app) in mainnet for now.
			return versionMap, nil
		},
	},
	"paua-rc2": { // upgrade for v1.14.0-rc3
		Handler: func(ctx sdk.Context, app *App, plan upgradetypes.Plan) (module.VersionMap, error) {
			// Reapply the max commissions thing again so testnet gets the max change rate bump too.
			IncreaseMaxCommissions(ctx, app)
			UndoMaxGasIncrease(ctx, app)
			return app.UpgradeKeeper.GetModuleVersionMap(ctx), nil
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

func IncreaseMaxCommissions(ctx sdk.Context, app *App) {
	oneHundredPct := sdk.OneDec()
	fivePct := sdk.MustNewDecFromStr("0.05")
	validators := app.StakingKeeper.GetAllValidators(ctx)
	ctx.Logger().Info("Increasing all validator's max commission to 100% and max change rate to 5%", "count", len(validators))
	for _, validator := range validators {
		validator.Commission.MaxRate = oneHundredPct
		// Note: This MaxChangeRate bump was added after paua-rc1 was run on testnet.
		// So, even though it's called by the paua-rc1 upgrade handler now,
		// it wasn't part of the actual paua-rc1 upgrade that was performed on testnet.
		if validator.Commission.MaxChangeRate.LT(fivePct) {
			validator.Commission.MaxChangeRate = fivePct
		}
		app.StakingKeeper.SetValidator(ctx, validator)
	}
}

func UndoMaxGasIncrease(ctx sdk.Context, app *App) {
	ctx.Logger().Info("Setting max gas per block back to 60,000,000")
	params := app.GetConsensusParams(ctx)
	params.Block.MaxGas = 60_000_000
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

// RemoveIsSendEnabledEntries removes all entries in the bankkeepers send enabled table
func RemoveIsSendEnabledEntries(ctx sdk.Context, app *App) {
	sendEnabledItems := app.BankKeeper.GetAllSendEnabledEntries(ctx)
	for _, item := range sendEnabledItems {
		marker, err := app.MarkerKeeper.GetMarkerByDenom(ctx, item.Denom)
		if err == nil {
			app.BankKeeper.DeleteSendEnabled(ctx, marker.GetDenom())
		}
	}
}
