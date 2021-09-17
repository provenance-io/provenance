package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/x/marker/types"
)

// InitGenesis creates the initial genesis state for the marker module.  Typically these
// accounts would be listed with the rest of the accounts and not created here.
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	k.SetParams(ctx, data.Params)
	if err := data.Validate(); err != nil {
		panic(err)
	}

	// ensure our store contains references to any marker accounts in auth genesis
	store := ctx.KVStore(k.storeKey)
	acc := k.authKeeper.GetAllAccounts(ctx)
	for i := range acc {
		if m, ok := acc[i].(types.MarkerAccountI); ok {
			if err := m.Validate(); err == nil {
				store.Set(types.MarkerStoreKey(m.GetAddress()), m.GetAddress())
			}
		}
	}
	// if any markers were included directly, add these as well.
	if data.Markers != nil {
		for i := range data.Markers {
			// marker base account may already exist and have an account number assigned
			if exists := k.authKeeper.GetAccount(ctx, data.Markers[i].GetAddress()); exists != nil {
				if err := data.Markers[i].SetAccountNumber(exists.GetAccountNumber()); err != nil {
					panic(err)
				}
			} else {
				// no existing account reference so take next account number.
				if err := data.Markers[i].SetAccountNumber(k.authKeeper.GetNextAccountNumber(ctx)); err != nil {
					panic(err)
				}
			}

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
		markers = append(markers, types.MarkerAccount{
			BaseAccount: &authtypes.BaseAccount{
				Address:       marker.GetAddress().String(),
				AccountNumber: marker.GetAccountNumber(),
				Sequence:      0,
			},
			Manager:                marker.GetManager().String(),
			AccessControl:          marker.GetAccessList(),
			Status:                 marker.GetStatus(),
			Denom:                  marker.GetDenom(),
			Supply:                 marker.GetSupply().Amount,
			MarkerType:             marker.GetMarkerType(),
			SupplyFixed:            marker.HasFixedSupply(),
			AllowGovernanceControl: marker.HasGovernanceEnabled(),
		})
		return false
	}

	k.IterateMarkers(ctx, appendToMarkers)
	return types.NewGenesisState(params, markers)
}
