package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/name/types"
)

// GetParams returns the total set of distribution parameters.
func (k Keeper) GetParams(clientCtx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(clientCtx, &params)
	return params
}

// SetParams sets the distribution parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetMaxNameLevels returns the current maximum number of name segments allowed.
func (k Keeper) GetMaxNameLevels(ctx sdk.Context) (max uint32) {
	k.paramSpace.Get(ctx, types.ParamStoreKeyMaxNameLevels, &max)
	return max
}

// GetMaxSegmentLength returns the current maximum length allowed for a name segment.
func (k Keeper) GetMaxSegmentLength(ctx sdk.Context) (max uint32) {
	k.paramSpace.Get(ctx, types.ParamStoreKeyMaxSegmentLength, &max)
	return max
}

// GetMinSegmentLength returns the current minimum allowed name segment length.
// rate.
func (k Keeper) GetMinSegmentLength(ctx sdk.Context) (min uint32) {
	k.paramSpace.Get(ctx, types.ParamStoreKeyMinSegmentLength, &min)
	return min
}

// GetAllowUnrestrictedNames returns the current unrestricted names allowed parameter.
func (k Keeper) GetAllowUnrestrictedNames(ctx sdk.Context) (enabled bool) {
	k.paramSpace.Get(ctx, types.ParamStoreKeyAllowUnrestrictedNames, &enabled)
	return enabled
}
