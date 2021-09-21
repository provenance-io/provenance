package v042

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/marker/types"
)

func MigrateMarkerAddressKeys(ctx sdk.Context, storeKey sdk.StoreKey) error {
	ctx.Logger().Info("Migrating Marker Module Markers (1/2)")
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
	ctx.Logger().Info("Migrating Marker Module Marker Permissions (2/2)")
	var err error
	k.IterateMarkers(ctx, func(marker types.MarkerAccountI) (stop bool) {
		if marker.GetMarkerType() == types.MarkerType_Coin {
			invalid := marker.AddressListForPermission(types.Access_Transfer)
			// invalid permission grants exist, remediation required
			if len(invalid) > 0 {
				m, ok := marker.(*types.MarkerAccount)
				if !ok {
					err = fmt.Errorf("unable to cast MarkerAccountI to MarkerAccount: %v", marker)
					return true
				}
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
