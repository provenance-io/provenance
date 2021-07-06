package v042

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/name/types"
)

func MigrateAddresses(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) error {
	err := MigrateAddressLength(ctx, storeKey, cdc)
	if err != nil {
		return err
	}
	return MigrateNameAddress(ctx, storeKey, cdc)
}

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
		nameKey, err := types.GetNameKeyPrefix(nameRecord.Name)
		if err != nil {
			return err
		}
		updatedAddress := ConvertLegacyNameAddress(legacyAddress)
		nameRecord.Address = updatedAddress.String()
		updateAddressKey, err := types.GetAddressKeyPrefix(updatedAddress)
		if err != nil {
			return err
		}
		updatedKey := append(updateAddressKey, nameKey...)
		if err != nil {
			return err
		}
		bz, err := cdc.Marshal(&nameRecord)
		if err != nil {
			return err
		}
		store.Set(updatedKey, bz)
		oldStore.Delete(oldStoreIter.Key())
	}
	return nil
}

func MigrateNameAddress(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	oldStore := prefix.NewStore(store, types.NameKeyPrefix)

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
		bz, err := cdc.Marshal(&nameRecord)
		if err != nil {
			return err
		}
		namePrefixKey, _ := types.GetNameKeyPrefix(nameRecord.Name)
		if !bytes.Equal(namePrefixKey, namePrefixKey) {
			return fmt.Errorf("")
		}
		store.Set(namePrefixKey, bz)
		store.Delete(oldStoreIter.Key())
	}
	return nil
}
