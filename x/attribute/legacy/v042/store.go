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
		attrAddress, err := sdk.AccAddressFromBech32(attribute.Address)
		if err != nil {
			return err
		}

		newStoreKey := types.AccountAttributeKey(attrAddress, attribute)

		bz, err := cdc.Marshal(&attribute)
		if err != nil {
			return err
		}

		store.Set(newStoreKey, bz)
		oldStore.Delete(oldStoreIter.Key())
	}
	return nil
}
