package v042

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/metadata/types"
)

func MigrateAddresses(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) error {
	return MigrateOSLocatorKeys(ctx, storeKey, cdc)
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

// func MigrateAddressScopeCacheKeys(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) error {
// 	store := ctx.KVStore(storeKey)
// 	oldStore := prefix.NewStore(store, AddressScopeCacheKeyPrefixLegacy)

// 	oldStoreIter := oldStore.Iterator(nil, nil)
// 	defer oldStoreIter.Close()

// 	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
// 		legacyAddr := oldStoreIter.Key()[0:20]
// 		scopeId := oldStoreIter.Key()[21:]
// 		err := sdk.VerifyAddressFormat(legacyAddr)
// 		if err == nil {
// 			return err
// 		}
// 		accAddress := sdk.AccAddress(legacyAddr)

// 		newStoreKey := types.GetAddressScopeCacheKey(accAddress)

// 		bz, err := cdc.Marshal(&osLocator)
// 		if err != nil {
// 			return err
// 		}

// 		store.Set(newStoreKey, bz)
// 		oldStore.Delete(oldStoreIter.Key())
// 	}
// 	return nil
// }
