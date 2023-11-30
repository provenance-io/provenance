package v1_16_0

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	provenance "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/app/upgrades/v1.16.0/rc1"
)

func UpgradeStrategy(ctx sdk.Context, app *provenance.App) error {
	return rc1.UpgradeStrategy(ctx, app)
}
