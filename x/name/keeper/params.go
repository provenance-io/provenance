package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/name/types"
)

// GetParams returns the total set of name parameters with fallback to default values.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	params, err := k.paramsStore.Get(ctx)
	if err != nil {
		return types.DefaultParams()
	}
	return params
}

// SetParams sets the name parameters to the store.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	if err := k.paramsStore.Set(ctx, params); err != nil {
		k.Logger(ctx).Error("failed to set params", "error", err)
	}
}

// GetMaxNameLevels returns the current maximum number of name segments allowed.
func (k Keeper) GetMaxNameLevels(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).MaxNameLevels
}

// GetMaxSegmentLength returns the current maximum length allowed for a name segment.
func (k Keeper) GetMaxSegmentLength(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).MaxSegmentLength
}

// GetMinSegmentLength returns the current minimum allowed name segment length.
func (k Keeper) GetMinSegmentLength(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).MinSegmentLength
}

// GetAllowUnrestrictedNames returns whether unrestricted names are allowed.
func (k Keeper) GetAllowUnrestrictedNames(ctx sdk.Context) bool {
	return k.GetParams(ctx).AllowUnrestrictedNames
}
