package v042

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/name/types"
)

func MigrateAddressLength(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	oldStore := prefix.NewStore(store, AddressKeyPrefixLegacy)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		var nameRecord types.NameRecord
		err := cdc.UnmarshalInterface(oldStoreIter.Value(), &nameRecord)
		if err != nil {
			return err
		}
		legacyAddress, err := sdk.AccAddressFromBech32(nameRecord.Address)
		if err != nil {
			return err
		}
		updatedAddress := ConvertLegacyNameAddress(legacyAddress)
		nameRecord.Address = updatedAddress.String()
		newStoreKey, err := types.GetAddressKeyPrefix(updatedAddress)
		if err != nil {
			return err
		}

		bz, err := cdc.Marshal(&nameRecord)
		if err != nil {
			return err
		}
		store.Set(newStoreKey, bz)
		oldStore.Delete(oldStoreIter.Key())
	}
	return nil
}
