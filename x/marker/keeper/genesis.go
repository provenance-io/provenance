package keeper

import (
	"cosmossdk.io/collections"

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
	acc := k.authKeeper.GetAllAccounts(ctx)
	for i := range acc {
		if m, ok := acc[i].(types.MarkerAccountI); ok {
			if err := m.Validate(); err == nil {
				if err := k.markers.Set(ctx, m.GetAddress(), m.GetAddress()); err != nil {
					panic(err)
				}
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
				if err := data.Markers[i].SetAccountNumber(k.authKeeper.NextAccountNumber(ctx)); err != nil {
					panic(err)
				}
			}

			k.SetMarker(ctx, &data.Markers[i])
		}
	}

	for _, denyAddress := range data.DenySendAddresses {
		markerAddr := sdk.MustAccAddressFromBech32(denyAddress.MarkerAddress)
		denyAddress := sdk.MustAccAddressFromBech32(denyAddress.DenyAddress)
		k.AddSendDeny(ctx, markerAddr, denyAddress)
	}
	for _, mNavs := range data.NetAssetValues {
		address := sdk.MustAccAddressFromBech32(mNavs.Address)
		for _, nav := range mNavs.NetAssetValues {
			navCopy := nav
			if err := k.navs.Set(ctx, collections.Join(address, navCopy.Price.Denom), navCopy); err != nil {
				panic(err)
			}
		}
	}
}

// ExportGenesis exports the current keeper state of the marker module.ExportGenesis
// We do not export anything because our marker accounts will be exported/imported by the Account Module.
func (k Keeper) ExportGenesis(ctx sdk.Context) (data *types.GenesisState) {
	params := k.GetParams(ctx)

	var markers []types.MarkerAccount
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
			AllowForcedTransfer:    marker.AllowsForcedTransfer(),
			RequiredAttributes:     marker.GetRequiredAttributes(),
		})
		return false
	}
	k.IterateMarkers(ctx, appendToMarkers)

	var denyAddresses []types.DenySendAddress
	_ = k.denySend.Walk(ctx, nil, func(key collections.Pair[sdk.AccAddress, sdk.AccAddress], _ bool) (bool, error) {
		denyAddresses = append(denyAddresses, types.DenySendAddress{
			MarkerAddress: key.K1().String(),
			DenyAddress:   key.K2().String(),
		})
		return false, nil
	})

	markerNetAssetValues := make([]types.MarkerNetAssetValues, len(markers))
	for i := range markers {
		var markerNavs types.MarkerNetAssetValues
		var navs []types.NetAssetValue
		err := k.IterateNetAssetValues(ctx, markers[i].GetAddress(), func(nav types.NetAssetValue) (stop bool) {
			navs = append(navs, nav)
			return false
		})
		if err != nil {
			panic(err)
		}
		markerNavs.Address = markers[i].GetAddress().String()
		markerNavs.NetAssetValues = navs
		markerNetAssetValues[i] = markerNavs
	}

	return types.NewGenesisState(params, markers, denyAddresses, markerNetAssetValues)
}
