package v042

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/attribute/types"
)

func MigrateAddressLength(ctx sdk.Context, storeKey sdk.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	oldStore := prefix.NewStore(store, AttributeKeyPrefixLegacy)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		var attribute types.Attribute
		err := cdc.UnmarshalInterface(oldStoreIter.Value(), &attribute)
		if err != nil {
			return err
		}
		legacyAddress, err := sdk.AccAddressFromBech32(attribute.Address)
		if err != nil {
			return err
		}

		updatedAddress := ConvertLegacyAddress(legacyAddress)
		attribute.Address = updatedAddress.String()
		newStoreKey := types.AccountAttributeKey(updatedAddress, attribute)

		bz, err := cdc.Marshal(&attribute)
		if err != nil {
			return err
		}

		store.Set(newStoreKey, bz)
		oldStore.Delete(oldStoreIter.Key())
	}
	return nil
}

func ConvertLegacyAddress(legacyAddr sdk.AccAddress) sdk.AccAddress {
	padding := make([]byte, 12)
	updatedAddr := append(legacyAddr.Bytes(), padding...)
	return sdk.AccAddress(updatedAddr)
}
