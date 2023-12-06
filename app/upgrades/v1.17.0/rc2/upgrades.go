package rc2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	provenance "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/app/upgrades"
	"github.com/provenance-io/provenance/app/upgrades/v1.17.0/common"
)

func UpgradeStrategy(ctx sdk.Context, app *provenance.App, vm module.VersionMap) (module.VersionMap, error) {
	// Migrate all the modules
	newVM, err := upgrades.RunModuleMigrations(ctx, app, vm)
	if err != nil {
		return nil, err
	}

	return PerformUpgrade(ctx, app, newVM)
}

func PerformUpgrade(ctx sdk.Context, app *provenance.App, vm module.VersionMap) (module.VersionMap, error) {
	common.UpdateIbcMarkerDenomMetadata(ctx, app)
	return vm, nil
}
