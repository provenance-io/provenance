package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// GetParams returns the total set of metadata parameters.
func (k Keeper) GetOSLocatorParams(ctx sdk.Context) (osLocatorParams types.OSLocatorParams) {
	return types.OSLocatorParams{
		MaxUriLength: k.GetMaxURILength(ctx),
	}
}

// SetParams sets the distribution parameters to the param space.
func (k Keeper) SetOSLocatorParams(ctx sdk.Context, params types.OSLocatorParams) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetMaxURILength gets the configured parameter for max uri length on a locator record (or the default if unset)
func (k Keeper) GetMaxURILength(ctx sdk.Context) (max uint32) {
	max = types.DefaultMaxURILength
	if k.paramSpace.Has(ctx, types.ParamStoreKeyMaxValueLength) {
		k.paramSpace.Get(ctx, types.ParamStoreKeyMaxValueLength, &max)
	}
	return
}
