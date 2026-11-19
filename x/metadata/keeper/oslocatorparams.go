package keeper

import (
	"errors"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// GetOSLocatorParams returns the metadata OSLocatorParams.
func (k Keeper) GetOSLocatorParams(ctx sdk.Context) (osLocatorParams types.OSLocatorParams) {
	params, err := k.osLocatorParamsCollection.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.OSLocatorParams{MaxUriLength: types.DefaultMaxURILength}
		}
		panic(err)
	}
	return params
}

// SetOSLocatorParams sets the metadata OSLocator parameters to the store.
func (k Keeper) SetOSLocatorParams(ctx sdk.Context, params types.OSLocatorParams) {
	if err := k.osLocatorParamsCollection.Set(ctx, params); err != nil {
		panic(err)
	}
}

// GetMaxURILength returns the configured parameter for max URI length on a locator record
func (k Keeper) GetMaxURILength(ctx sdk.Context) uint32 {
	return k.GetOSLocatorParams(ctx).MaxUriLength
}
