package rc2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/provenance-io/provenance/app/keepers"
	"github.com/provenance-io/provenance/app/upgrades"
	"github.com/provenance-io/provenance/app/upgrades/mainnet/v1.17.0/common"
)

func UpgradeStrategy(ctx sdk.Context, app upgrades.AppUpgrader, vm module.VersionMap) (module.VersionMap, error) {
	// Migrate all the modules
	newVM, err := upgrades.RunModuleMigrations(ctx, app, vm)
	if err != nil {
		return nil, err
	}

	if err = PerformUpgrade(ctx, app.Keepers()); err != nil {
		return nil, err
	}

	return newVM, err
}

func PerformUpgrade(ctx sdk.Context, k *keepers.AppKeepers) error {
	common.UpdateIbcMarkerDenomMetadata(ctx, k)
	return nil
}
