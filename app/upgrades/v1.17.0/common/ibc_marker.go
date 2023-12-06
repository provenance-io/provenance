package common

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	provenance "github.com/provenance-io/provenance/app"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

// updateIbcMarkerDenomMetadata iterates markers and creates denom metadata for ibc markers
// TODO: Remove with the saffron handlers.
func UpdateIbcMarkerDenomMetadata(ctx sdk.Context, app *provenance.App) {
	ctx.Logger().Info("Updating ibc marker denom metadata")
	app.MarkerKeeper.IterateMarkers(ctx, func(record markertypes.MarkerAccountI) bool {
		if !strings.HasPrefix(record.GetDenom(), "ibc/") {
			return false
		}

		hash, err := transfertypes.ParseHexHash(strings.TrimPrefix(record.GetDenom(), "ibc/"))
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("invalid denom trace hash: %s, error: %s", hash.String(), err))
			return false
		}
		denomTrace, found := app.TransferKeeper.GetDenomTrace(ctx, hash)
		if !found {
			ctx.Logger().Error(fmt.Sprintf("trace not found: %s, error: %s", hash.String(), err))
			return false
		}

		parts := strings.Split(denomTrace.Path, "/")
		if len(parts) == 2 && parts[0] == "transfer" {
			ctx.Logger().Info(fmt.Sprintf("Adding metadata to %s", record.GetDenom()))
			chainID := app.Ics20MarkerHooks.GetChainID(ctx, parts[0], parts[1], app.IBCKeeper)
			markerMetadata := banktypes.Metadata{
				Base:        record.GetDenom(),
				Name:        chainID + "/" + denomTrace.BaseDenom,
				Display:     chainID + "/" + denomTrace.BaseDenom,
				Description: denomTrace.BaseDenom + " from " + chainID,
			}
			app.BankKeeper.SetDenomMetaData(ctx, markerMetadata)
		}

		return false
	})
	ctx.Logger().Info("Done updating ibc marker denom metadata")
}
