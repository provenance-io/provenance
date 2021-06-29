package keeper

import (
	"github.com/provenance-io/provenance/x/attribute/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis creates the initial genesis state for the attribute module.
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	k.SetParams(ctx, data.Params)
	if err := data.ValidateBasic(); err != nil {
		panic(err)
	}
	for _, attr := range data.Attributes {
		if err := k.importAttribute(ctx, attr); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis exports the current keeper state of the attribute module.
func (k Keeper) ExportGenesis(ctx sdk.Context) (data *types.GenesisState) {
	attrs := make([]types.Attribute, 0)
	params := k.GetParams(ctx)

	appendToRecords := func(attr types.Attribute) error {
		attrs = append(attrs, attr)
		return nil
	}

	if err := k.IterateRecords(ctx, types.AttributeKeyPrefixLegacy, appendToRecords); err != nil {
		panic(err)
	}

	return types.NewGenesisState(params, attrs)
}

// convert name records before 1.0.0 to the right proto encoding.
func (k Keeper) ConvertLegacyAmino(ctx sdk.Context) []types.Attribute {
	attrs := make([]types.Attribute, 0)
	appendToRecords := func(attr types.Attribute) error {
		attrs = append(attrs, attr)
		return nil
	}
	if err := k.IterateRecords(ctx, types.AttributeKeyPrefixAmino, appendToRecords); err != nil {
		panic(err)
	}
	for _, record := range attrs {
		if err := k.importAttribute(ctx, record); err != nil {
			panic(err)
		}
	}
	return attrs
}
