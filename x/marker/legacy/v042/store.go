package v042

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/marker/types"
)

func MigrateMarkerAddressKeys(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	oldStore := prefix.NewStore(store, MarkerStoreKeyPrefixLegacy)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		var marker = sdk.AccAddress(oldStoreIter.Value())
		newStoreKey := types.MarkerStoreKey(marker)
		store.Set(newStoreKey, oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}
	return nil
}

// MigrateMarkerPermissions inspects existing COIN markers for grants of Access_Transfer and removes the invalid
// grant to prevent validation errors with the strict access validation updates in 1.6.0
func MigrateMarkerPermissions(ctx sdk.Context, k MarkerKeeperI) error {
	var err error
	k.IterateMarkers(ctx, func(marker types.MarkerAccountI) (stop bool) {
		if marker.GetMarkerType() == types.MarkerType_Coin {
			invalid := marker.AddressListForPermission(types.Access_Transfer)
			// invalid permission grants exist, remediation required
			if len(invalid) > 0 {
				m := marker.Clone()
				var accessList []types.AccessGrant
				for _, ac := range m.AccessControl {
					if ac.HasAccess(types.Access_Transfer) {
						if err = ac.RemoveAccess(types.Access_Transfer); err != nil {
							return true // stop iterating
						}
					}
					accessList = append(accessList, ac)
				}
				m.AccessControl = accessList
				k.SetMarker(ctx, m)
			}
		}
		return false
	})
	return err
}
