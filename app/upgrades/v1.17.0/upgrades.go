package v1_17_0

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/provenance-io/provenance/app/keepers"
	navs "github.com/provenance-io/provenance/app/navs"
	"github.com/provenance-io/provenance/app/upgrades"
	rc1 "github.com/provenance-io/provenance/app/upgrades/v1.17.0/rc1"
	rc2 "github.com/provenance-io/provenance/app/upgrades/v1.17.0/rc2"
	rc3 "github.com/provenance-io/provenance/app/upgrades/v1.17.0/rc3"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
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
	if err = rc1.PerformUpgrade(ctx, k); err != nil {
		return err
	}
	if err = rc2.PerformUpgrade(ctx, k); err != nil {
		return err
	}
	if err = rc3.PerformUpgrade(ctx, k); err != nil {
		return err
	}

	AddMarkerNavs(ctx, k, navs.GetPioMainnet1DenomToNav())
	return nil
}

// addMarkerNavs adds navs to existing markers
func AddMarkerNavs(ctx sdk.Context, k *keepers.AppKeepers, denomToNav map[string]markertypes.NetAssetValue) {
	ctx.Logger().Info("Adding marker net asset values")
	for denom, nav := range denomToNav {
		marker, err := k.MarkerKeeper.GetMarkerByDenom(ctx, denom)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("unable to get marker %v: %v", denom, err))
			continue
		}
		if err := k.MarkerKeeper.AddSetNetAssetValues(ctx, marker, []markertypes.NetAssetValue{nav}, "upgrade_handler"); err != nil {
			ctx.Logger().Error(fmt.Sprintf("unable to set net asset value %v: %v", nav, err))
		}
	}
	ctx.Logger().Info("Done adding marker net asset values")
}
