package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/name/types"
)

// GetParams returns the total set of name parameters with fall thorugh to default values.
func (keeper Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	return types.Params{
		MaxSegmentLength:       keeper.GetMaxSegmentLength(ctx),
		MinSegmentLength:       keeper.GetMinSegmentLength(ctx),
		MaxNameLevels:          keeper.GetMaxNameLevels(ctx),
		AllowUnrestrictedNames: keeper.GetAllowUnrestrictedNames(ctx),
	}
}

// SetParams sets the distribution parameters to the param space.
func (keeper Keeper) SetParams(ctx sdk.Context, params types.Params) {
	keeper.paramSpace.SetParamSet(ctx, &params)
}

// GetMaxNameLevels returns the current maximum number of name segments allowed (or default if unset)
func (keeper Keeper) GetMaxNameLevels(ctx sdk.Context) (max uint32) {
	max = types.DefaultMaxSegments
	if keeper.paramSpace.Has(ctx, types.ParamStoreKeyMaxNameLevels) {
		keeper.paramSpace.Get(ctx, types.ParamStoreKeyMaxNameLevels, &max)
	}
	return
}

// GetMaxSegmentLength returns the current maximum length allowed for a name segment (or default if unset)
func (keeper Keeper) GetMaxSegmentLength(ctx sdk.Context) (max uint32) {
	max = types.DefaultMaxSegmentLength
	if keeper.paramSpace.Has(ctx, types.ParamStoreKeyMaxSegmentLength) {
		keeper.paramSpace.Get(ctx, types.ParamStoreKeyMaxSegmentLength, &max)
	}
	return
}

// GetMinSegmentLength returns the current minimum allowed name segment length (or default if unset)
func (keeper Keeper) GetMinSegmentLength(ctx sdk.Context) (min uint32) {
	min = types.DefaultMinSegmentLength
	if keeper.paramSpace.Has(ctx, types.ParamStoreKeyMinSegmentLength) {
		keeper.paramSpace.Get(ctx, types.ParamStoreKeyMinSegmentLength, &min)
	}
	return
}

// GetAllowUnrestrictedNames returns the current unrestricted names allowed parameter (or default if unset)
func (keeper Keeper) GetAllowUnrestrictedNames(ctx sdk.Context) (enabled bool) {
	enabled = types.DefaultAllowUnrestrictedNames
	if keeper.paramSpace.Has(ctx, types.ParamStoreKeyAllowUnrestrictedNames) {
		keeper.paramSpace.Get(ctx, types.ParamStoreKeyAllowUnrestrictedNames, &enabled)
	}
	return
}
