package v042

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/name/types"
)

func MigrateAddresses(ctx sdk.Context, storeKey sdk.StoreKey) error {
	store := ctx.KVStore(storeKey)
	oldStore := prefix.NewStore(store, AddressKeyPrefixLegacy)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		var nameRecord types.NameRecord
		err := types.ModuleCdc.UnmarshalInterface(oldStoreIter.Value(), &nameRecord)
		if err != nil {
			return err
		}
		legacyAddress, err := sdk.AccAddressFromBech32(nameRecord.Address)
		if err != nil {
			return err
		}
		nameKey, err := types.GetNameKeyPrefix(nameRecord.Name)
		if err != nil {
			return err
		}
		updateAddress, err := types.GetAddressKeyPrefix(legacyAddress)
		if err != nil {
			return err
		}
		updateAddress = append(updateAddress, nameKey...)
		store.Set(updateAddress, oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}
	return nil
}
