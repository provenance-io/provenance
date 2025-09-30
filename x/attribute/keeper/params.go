package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/attribute/types"
)

// GetParams returns the attribute Params.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.AttributeParamPrefix)
	if bz == nil {
		return types.Params{
			MaxValueLength: types.DefaultMaxValueLength,
		}
	}
	err := k.cdc.Unmarshal(bz, &params)
	if err != nil {
		panic(err)
	}
	return params
}

// SetParams sets the account parameters to the param store.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		panic(err)
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(types.AttributeParamPrefix, bz)
}

// GetMaxValueLength returns the max value for attribute length.
func (k Keeper) GetMaxValueLength(ctx sdk.Context) (maxValueLength uint32) {
	return k.GetParams(ctx).MaxValueLength
}
