package v042

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/metadata/types"
)

func MigrateAddresses(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) error {
	err := MigrateOSLocatorKeys(ctx, storeKey, cdc)
	if err != nil {
		return err
	}
	err = MigrateAddressScopeCacheKey(ctx, storeKey, cdc)
	if err != nil {
		return err
	}
	err = MigrateValueOwnerScopeCacheKey(ctx, storeKey, cdc)
	if err != nil {
		return err
	}
	err = MigrateAddressScopeSpecCacheKey(ctx, storeKey, cdc)
	if err != nil {
		return err
	}
	return MigrateAddressContractSpecCacheKey(ctx, storeKey, cdc)
}

func MigrateOSLocatorKeys(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	oldStore := prefix.NewStore(store, OSLocatorAddressKeyPrefixLegacy)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		var osLocator types.ObjectStoreLocator
		err := cdc.UnmarshalInterface(oldStoreIter.Value(), &osLocator)
		if err != nil {
			return err
		}
		legacyAddress, err := sdk.AccAddressFromBech32(osLocator.Owner)
		if err != nil {
			return err
		}

		newStoreKey := types.GetOSLocatorKey(legacyAddress)

		bz, err := cdc.Marshal(&osLocator)
		if err != nil {
			return err
		}

		store.Set(newStoreKey, bz)
		oldStore.Delete(oldStoreIter.Key())
	}
	return nil
}

func MigrateAddressScopeCacheKey(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	oldStore := prefix.NewStore(store, AddressScopeCacheKeyPrefixLegacy)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		addr := oldStoreIter.Key()[1:21]
		legacyAddress := sdk.AccAddress(addr)

		newStoreKey := types.GetAddressScopeCacheIteratorPrefix(legacyAddress)
		metaaddress := oldStoreIter.Key()[21:]

		store.Set(append(newStoreKey, metaaddress...), oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}

	return nil
}

func MigrateValueOwnerScopeCacheKey(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	oldStore := prefix.NewStore(store, ValueOwnerScopeCacheKeyPrefixLegacy)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		addr := oldStoreIter.Key()[1:21]
		legacyAddress := sdk.AccAddress(addr)

		newStoreKey := types.GetValueOwnerScopeCacheIteratorPrefix(legacyAddress)
		metaaddress := oldStoreIter.Key()[21:]

		store.Set(append(newStoreKey, metaaddress...), oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}

	return nil
}

func MigrateAddressScopeSpecCacheKey(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	oldStore := prefix.NewStore(store, AddressScopeSpecCacheKeyPrefixLegacy)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		addr := oldStoreIter.Key()[1:21]
		legacyAddress := sdk.AccAddress(addr)

		newStoreKey := types.GetAddressScopeSpecCacheIteratorPrefix(legacyAddress)
		metaaddress := oldStoreIter.Key()[21:]

		store.Set(append(newStoreKey, metaaddress...), oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}

	return nil
}

func MigrateAddressContractSpecCacheKey(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	oldStore := prefix.NewStore(store, AddressContractSpecCacheKeyPrefixLegacy)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		addr := oldStoreIter.Key()[1:21]
		legacyAddress := sdk.AccAddress(addr)

		newStoreKey := types.GetAddressContractSpecCacheIteratorPrefix(legacyAddress)
		metaaddress := oldStoreIter.Key()[21:]

		store.Set(append(newStoreKey, metaaddress...), oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}

	return nil
}
