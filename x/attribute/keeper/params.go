package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/attribute/types"
)

// GetParams returns the total set of account parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	return types.Params{
		MaxValueLength: k.GetMaxValueLength(ctx),
	}
}

// SetParams sets the account parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetMaxValueLength returns the current distribution community tax (or default if unset)
func (k Keeper) GetMaxValueLength(ctx sdk.Context) (maxValueLength uint32) {
	maxValueLength = types.DefaultMaxValueLength
	if k.paramSpace.Has(ctx, types.ParamStoreKeyMaxValueLength) {
		k.paramSpace.Get(ctx, types.ParamStoreKeyMaxValueLength, &maxValueLength)
	}
	return
}
