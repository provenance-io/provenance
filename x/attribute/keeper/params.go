package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/attribute/types"
)

// GetParams returns the attribute Params.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	params, err := k.params.Get(ctx)
	if err != nil {
		return types.Params{
			MaxValueLength: types.DefaultMaxValueLength,
		}
	}
	return params
}

// SetParams sets the account parameters to the param store.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	if err := k.params.Set(ctx, params); err != nil {
		panic(err)
	}
}

// GetMaxValueLength returns the max value for attribute length.
func (k Keeper) GetMaxValueLength(ctx sdk.Context) (maxValueLength uint32) {
	return k.GetParams(ctx).MaxValueLength
}
