package v042

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/attribute/keeper"
	"github.com/provenance-io/provenance/x/attribute/types"
)

func MigrateAddressLength(attrKeeper keeper.Keeper, ctx sdk.Context) {
	attrs := make([]types.Attribute, 0)
	appendToRecords := func(attr types.Attribute) error {
		attrs = append(attrs, attr)
		return nil
	}
	if err := attrKeeper.IterateRecords(ctx, AttributeKeyPrefixLegacy, appendToRecords); err != nil {
		panic(err)
	}
	for _, legacyAttribute := range attrs {
		legacyAddr, err := sdk.AccAddressFromBech32(legacyAttribute.Address)
		if err != nil {
			panic(err)
		}
		legacyKey := AccountAttributeKeyLegacy(legacyAddr, legacyAttribute)
		padding := make([]byte, 12)
		updatedAddr := append(legacyAddr.Bytes(), padding...)
		updateAccAddr := sdk.AccAddress(updatedAddr)
		attrKeeper.UpdateAddributeAddress(ctx, legacyAttribute, updateAccAddr, legacyKey)
	}
}
