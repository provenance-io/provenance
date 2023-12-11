package v1_18_0

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/provenance-io/provenance/app/keepers"
	"github.com/provenance-io/provenance/app/upgrades"
	rc1 "github.com/provenance-io/provenance/app/upgrades/mainnet/v1.18.0/rc1"
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

	return newVM, nil
}

func PerformUpgrade(ctx sdk.Context, k *keepers.AppKeepers) (err error) {
	return rc1.PerformUpgrade(ctx, k)
}
