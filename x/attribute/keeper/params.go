package keeper

import (
	"errors"
	"fmt"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/attribute/types"
)

// GetParams returns the attribute Params.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	params, err := k.params.Get(ctx)
	switch {
	case err == nil:
		return params
	case errors.Is(err, collections.ErrNotFound):
		// Params have never been set, so use the defaults.
		return types.Params{
			MaxValueLength: types.DefaultMaxValueLength,
		}
	default:
		panic(fmt.Errorf("attribute: could not read params: %w", err))
	}
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
