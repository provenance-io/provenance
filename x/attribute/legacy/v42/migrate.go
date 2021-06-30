package v42

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/attribute/keeper"
	"github.com/provenance-io/provenance/x/attribute/types"
)

func MigrateAddressLength(attrKeeper keeper.Keeper, ctx sdk.Context) []types.Attribute {
	attrs := make([]types.Attribute, 0)
	appendToRecords := func(attr types.Attribute) error {
		attrs = append(attrs, attr)
		return nil
	}
	if err := attrKeeper.IterateRecords(ctx, types.AttributeKeyPrefixLegacy, appendToRecords); err != nil {
		panic(err)
	}
	for _, record := range attrs {
		if err := attrKeeper.UpdateAttributeAddressLength(ctx, record); err != nil {
			panic(err)
		}
	}
	return attrs
}
