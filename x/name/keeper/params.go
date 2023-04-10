package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/name/types"
)

// GetParams returns the total set of name parameters with fall thorugh to default values.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	return types.Params{
		MaxSegmentLength:       k.GetMaxSegmentLength(ctx),
		MinSegmentLength:       k.GetMinSegmentLength(ctx),
		MaxNameLevels:          k.GetMaxNameLevels(ctx),
		AllowUnrestrictedNames: k.GetAllowUnrestrictedNames(ctx),
	}
}

// SetParams sets the distribution parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetMaxNameLevels returns the current maximum number of name segments allowed (or default if unset)
func (k Keeper) GetMaxNameLevels(ctx sdk.Context) (max uint32) {
	max = types.DefaultMaxSegments
	if k.paramSpace.Has(ctx, types.ParamStoreKeyMaxNameLevels) {
		k.paramSpace.Get(ctx, types.ParamStoreKeyMaxNameLevels, &max)
	}
	return
}

// GetMaxSegmentLength returns the current maximum length allowed for a name segment (or default if unset)
func (k Keeper) GetMaxSegmentLength(ctx sdk.Context) (max uint32) {
	max = types.DefaultMaxSegmentLength
	if k.paramSpace.Has(ctx, types.ParamStoreKeyMaxSegmentLength) {
		k.paramSpace.Get(ctx, types.ParamStoreKeyMaxSegmentLength, &max)
	}
	return
}

// GetMinSegmentLength returns the current minimum allowed name segment length (or default if unset)
func (k Keeper) GetMinSegmentLength(ctx sdk.Context) (min uint32) {
	min = types.DefaultMinSegmentLength
	if k.paramSpace.Has(ctx, types.ParamStoreKeyMinSegmentLength) {
		k.paramSpace.Get(ctx, types.ParamStoreKeyMinSegmentLength, &min)
	}
	return
}

// GetAllowUnrestrictedNames returns the current unrestricted names allowed parameter (or default if unset)
func (k Keeper) GetAllowUnrestrictedNames(ctx sdk.Context) (enabled bool) {
	enabled = types.DefaultAllowUnrestrictedNames
	if k.paramSpace.Has(ctx, types.ParamStoreKeyAllowUnrestrictedNames) {
		k.paramSpace.Get(ctx, types.ParamStoreKeyAllowUnrestrictedNames, &enabled)
	}
	return
}
