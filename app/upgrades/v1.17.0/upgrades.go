package v1_17_0

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	provenance "github.com/provenance-io/provenance/app"
	rc1 "github.com/provenance-io/provenance/app/upgrades/v1.17.0/rc1"
	rc2 "github.com/provenance-io/provenance/app/upgrades/v1.17.0/rc2"
	rc3 "github.com/provenance-io/provenance/app/upgrades/v1.17.0/rc3"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

func UpgradeStrategy(ctx sdk.Context, app *provenance.App) error {
	if err := rc1.UpgradeStrategy(ctx, app); err != nil {
		return err
	}
	if err := rc2.UpgradeStrategy(ctx, app); err != nil {
		return err
	}
	if err := rc3.UpgradeStrategy(ctx, app); err != nil {
		return err
	}
	AddMarkerNavs(ctx, app, provenance.GetPioMainnet1DenomToNav())
	return nil
}

// addMarkerNavs adds navs to existing markers
func AddMarkerNavs(ctx sdk.Context, app *provenance.App, denomToNav map[string]markertypes.NetAssetValue) {
	ctx.Logger().Info("Adding marker net asset values")
	for denom, nav := range denomToNav {
		marker, err := app.MarkerKeeper.GetMarkerByDenom(ctx, denom)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("unable to get marker %v: %v", denom, err))
			continue
		}
		if err := app.MarkerKeeper.AddSetNetAssetValues(ctx, marker, []markertypes.NetAssetValue{nav}, "upgrade_handler"); err != nil {
			ctx.Logger().Error(fmt.Sprintf("unable to set net asset value %v: %v", nav, err))
		}
	}
	ctx.Logger().Info("Done adding marker net asset values")
}
