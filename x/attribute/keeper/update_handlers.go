package keeper

import (
	"github.com/provenance-io/provenance/x/attribute/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) ConvertAddressLength(ctx sdk.Context) []types.Attribute {
	attrs := make([]types.Attribute, 0)
	appendToRecords := func(attr types.Attribute) error {
		attrs = append(attrs, attr)
		return nil
	}
	if err := k.IterateRecords(ctx, types.AttributeKeyPrefixLegacy, appendToRecords); err != nil {
		panic(err)
	}
	for _, record := range attrs {
		if err := k.updateAttributeAddressLength(ctx, record); err != nil {
			panic(err)
		}
	}
	return attrs
}
