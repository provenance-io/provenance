package upgrades

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/provenance-io/provenance/app/keepers"
)

// runModuleMigrations wraps standard logging around the call to app.mm.RunMigrations.
// In most cases, it should be the first thing done during a migration.
//
// If state is updated prior to this migration, you run the risk of writing state using
// a new format when the migration is expecting all state to be in the old format.
func RunModuleMigrations(ctx sdk.Context, app AppUpgrader, vm module.VersionMap) (module.VersionMap, error) {
	// Even if this function is no longer called, do not delete it. Keep it around for the next time it's needed.
	ctx.Logger().Info("Starting module migrations. This may take a significant amount of time to complete. Do not restart node.")
	newVM, err := app.ModuleManager().RunMigrations(ctx, app.Configurator(), vm)
	if err != nil {
		ctx.Logger().Error("Module migrations encountered an error.", "error", err)
		return nil, err
	}
	ctx.Logger().Info("Module migrations completed.")
	return newVM, nil
}

// Create a use of runModuleMigrations so that the linter neither complains about it not being used,
// nor complains about a nolint:unused directive that isn't needed because the function is used.
var _ = RunModuleMigrations

func CreateUpgradeHandler(upgrade UpgradeStrategy, app AppUpgrader) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info(fmt.Sprintf("Starting upgrade to %q", plan.Name), "version-map", vm)

		// This is where we want to run the logic
		newVM, err := upgrade(ctx, app, vm)

		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("Failed to upgrade to %q", plan.Name), "error", err)
		} else {
			ctx.Logger().Info(fmt.Sprintf("Successfully upgraded to %q", plan.Name), "version-map", newVM)
		}
		return newVM, err
	}
}

// InstallCustomUpgradeHandlers sets upgrade handlers for all entries in the upgrades map.
func InstallCustomUpgradeHandlers(app AppUpgrader, upgrades []Upgrade) {
	// Register all explicit appUpgrades
	for _, upgrade := range upgrades {
		// If the handler has been defined, add it here, otherwise, use no-op.
		ref := upgrade
		handler := CreateUpgradeHandler(ref.UpgradeStrategy, app)
		app.Keepers().UpgradeKeeper.SetUpgradeHandler(ref.UpgradeName, handler)
	}
}

func AttemptUpgradeStoreLoaders(app StoreLoaderUpgrader, k *keepers.AppKeepers, upgrades []Upgrade) {
	// Use the dump of $home/data/upgrade-info.json:{"name":"$plan","height":321654} to determine
	// if we load a store upgrade from the handlers. No file == no error from read func.
	upgradeInfo, err := k.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	// Currently in an upgrade hold for this block.
	if upgradeInfo.Name != "" && upgradeInfo.Height == app.LastBlockHeight()+1 {
		if k.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
			app.Logger().Info("Skipping upgrade based on height",
				"plan", upgradeInfo.Name,
				"upgradeHeight", upgradeInfo.Height,
				"lastHeight", app.LastBlockHeight(),
			)
		} else {
			app.Logger().Info("Managing upgrade",
				"plan", upgradeInfo.Name,
				"upgradeHeight", upgradeInfo.Height,
				"lastHeight", app.LastBlockHeight(),
			)
			// See if we have a custom store loader to use for upgrades.
			storeLoader := GetUpgradeStoreLoader(app, upgradeInfo, upgrades)
			if storeLoader != nil {
				app.SetStoreLoader(storeLoader)
			}
		}
	}
}

// GetUpgradeStoreLoader creates an StoreLoader for use in an upgrade.
// Returns nil if no upgrade info is found or the upgrade doesn't need a store loader.
func GetUpgradeStoreLoader(app StoreLoaderUpgrader, info upgradetypes.Plan, upgrades []Upgrade) baseapp.StoreLoader {
	upgrade, found := FindUpgrade(info.Name, upgrades)
	if !found {
		return nil
	}

	if len(upgrade.StoreUpgrades.Renamed) == 0 && len(upgrade.StoreUpgrades.Deleted) == 0 && len(upgrade.StoreUpgrades.Added) == 0 {
		app.Logger().Info("No store upgrades required",
			"plan", info.Name,
			"height", info.Height,
		)
		return nil
	}

	app.Logger().Info("Store upgrades",
		"plan", info.Name,
		"height", info.Height,
		"upgrade.added", upgrade.StoreUpgrades.Added,
		"upgrade.deleted", upgrade.StoreUpgrades.Deleted,
		"upgrade.renamed", upgrade.StoreUpgrades.Renamed,
	)
	return upgradetypes.UpgradeStoreLoader(info.Height, &upgrade.StoreUpgrades)
}

func FindUpgrade(name string, upgrades []Upgrade) (*Upgrade, bool) {
	for _, upgrade := range upgrades {
		if upgrade.UpgradeName == name {
			return &upgrade, true
		}
	}
	return nil, false
}
