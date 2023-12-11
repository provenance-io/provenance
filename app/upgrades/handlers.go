package upgrades

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

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
		handler := NilStrategy
		if ref.UpgradeStrategy != nil {
			handler = CreateUpgradeHandler(ref.UpgradeStrategy, app)
		}

		app.Keepers().UpgradeKeeper.SetUpgradeHandler(ref.UpgradeName, handler)
	}
}

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

func NilStrategy(ctx sdk.Context, plan upgradetypes.Plan, versionMap module.VersionMap) (module.VersionMap, error) {
	ctx.Logger().Info(fmt.Sprintf("Applying no-op upgrade to %q", plan.Name))
	return versionMap, nil
}
