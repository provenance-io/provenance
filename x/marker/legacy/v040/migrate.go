package v040

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/x/auth/types"

	v039marker "github.com/provenance-io/provenance/x/marker/legacy/v039"
	v040marker "github.com/provenance-io/provenance/x/marker/types"
)

// Migrate accepts exported x/marker genesis state from v0.39 and migrates it
// to v0.40 x/marker genesis state. The migration includes:
//
// - Convert addresses from bytes to bech32 strings.
// - Re-encode in v0.40 GenesisState.
func Migrate(oldGenState v039marker.GenesisState) *v040marker.GenesisState {
	var markerAccounts = make([]v040marker.MarkerAccount, 0, len(oldGenState.Markers))
	for _, mark := range oldGenState.Markers {
		markerAccounts = append(markerAccounts, v040marker.MarkerAccount{
			BaseAccount: &types.BaseAccount{
				Address:       mark.Address.String(),
				AccountNumber: mark.AccountNumber,
				Sequence:      mark.Sequence,
			},
			Manager:       mark.Manager.String(),
			Status:        v040marker.MustGetMarkerStatus(mark.GetStatus()),
			Denom:         mark.Denom,
			Supply:        mark.GetSupply().Amount,
			AccessControl: migrateAccess(mark.AccessControls),
			// v039 only supported COIN type (ignore previous values as untyped field held trash)
			MarkerType: v040marker.MarkerType_Coin,
		})
	}
	return &v040marker.GenesisState{
		Params:  v040marker.DefaultParams(),
		Markers: markerAccounts,
	}
}

func migrateAccess(old []v039marker.AccessGrant) (new []v040marker.AccessGrant) {
	new = make([]v040marker.AccessGrant, len(old))
	for i, a := range old {
		perms := strings.Join(a.Permissions, ",")
		perms = strings.ToLower(perms)
		perms = strings.ReplaceAll(perms, "grant", "admin")
		new[i] = v040marker.AccessGrant{
			Address:     a.Address.String(),
			Permissions: v040marker.AccessListByNames(perms),
		}
	}
	return
}
