package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/name/types"
)

// GetParams returns the total set of name parameters with fallback to default values.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	params = types.DefaultParams() // Assuming DefaultParams initializes all defaults

	bz := store.Get(types.NameParamStoreKey) // General key for all parameters
	if bz != nil {
		k.cdc.MustUnmarshal(bz, &params) // Deserialize parameters from bytes
	}
	return params
}

// SetParams sets the name parameters to the store.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params) // Serialize parameters to bytes
	store.Set(types.NameParamStoreKey, bz)
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
