package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// GetOSLocatorParams returns the metadata OSLocatorParams.
func (k Keeper) GetOSLocatorParams(ctx sdk.Context) (osLocatorParams types.OSLocatorParams) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.OSLocatorParamPrefix)
	if bz == nil {
		return types.OSLocatorParams{
			MaxUriLength: types.DefaultMaxURILength,
		}
	}
	err := k.cdc.Unmarshal(bz, &osLocatorParams)
	if err != nil {
		panic(err)
	}
	return osLocatorParams
}

// SetOSLocatorParams sets the metadata OSLocator parameters to the store.
func (k Keeper) SetOSLocatorParams(ctx sdk.Context, params types.OSLocatorParams) {
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		panic(err)
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(types.OSLocatorParamPrefix, bz)
}

// GetMaxURILength returns the configured parameter for max URI length on a locator record
func (k Keeper) GetMaxURILength(ctx sdk.Context) (max uint32) {
	return k.GetOSLocatorParams(ctx).MaxUriLength
}
