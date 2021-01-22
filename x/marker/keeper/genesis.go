package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/x/marker/types"
)

// InitGenesis creates the initial genesis state for the marker module.  Typically these
// accounts would be listed with the rest of the accounts and not created here.
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	if err := data.Validate(); err != nil {
		panic(err)
	}
	if data.Markers != nil {
		for i := range data.Markers {
			k.SetMarker(ctx, &data.Markers[i])
		}
	}
}

// ExportGenesis exports the current keeper state of the marker module.ExportGenesis
// We do not export anything because our marker accounts will be exported/imported by the Account Module.
func (k Keeper) ExportGenesis(ctx sdk.Context) (data *types.GenesisState) {
	params := k.GetParams(ctx)
	markers := make([]types.MarkerAccount, 0)

	appendToMarkers := func(marker types.MarkerAccountI) bool {
		markers = append(markers, *types.NewMarkerAccount(
			&authtypes.BaseAccount{
				Address:       marker.GetAddress().String(),
				AccountNumber: marker.GetAccountNumber(),
				Sequence:      0,
			},
			marker.GetSupply(),
			marker.GetManager(),
			marker.GetAccessList(),
			marker.GetStatus(),
			marker.GetMarkerType(),
		))
		return false
	}

	k.IterateMarkers(ctx, appendToMarkers)
	return types.NewGenesisState(params, markers)
}
