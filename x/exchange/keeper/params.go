package keeper

import (
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/exchange"
)

// deleteAllParamsSplits deletes all the params splits in the store.
func deleteAllParamsSplits(store storetypes.KVStore) {
	keys := getAllKeys(store, GetKeyPrefixParamsSplit())
	for _, key := range keys {
		store.Delete(key)
	}
}

// setParamsSplit sets the provided params split for the provided denom.
func setParamsSplit(store storetypes.KVStore, denom string, split uint16) {
	key := MakeKeyParamsSplit(denom)
	value := uint16Bz(split)
	store.Set(key, value)
}

// getParamsSplit gets the params split amount for the provided denom, and whether the entry existed.
func getParamsSplit(store storetypes.KVStore, denom string) (uint16, bool) {
	key := MakeKeyParamsSplit(denom)
	if store.Has(key) {
		value := store.Get(key)
		return uint16FromBz(value)
	}
	return 0, false
}

// SetParams updates the params to match those provided.
// If nil is provided, all params are deleted.
func (k Keeper) SetParams(ctx sdk.Context, params *exchange.Params) {
	store := k.getStore(ctx)
	deleteAllParamsSplits(store)
	if params != nil {
		setParamsSplit(store, "", uint16(params.DefaultSplit))
		for _, split := range params.DenomSplits {
			setParamsSplit(store, split.Denom, uint16(split.Split))
		}
	}
}

// GetParams gets the exchange module params.
// If there aren't any params in state, nil is returned.
func (k Keeper) GetParams(ctx sdk.Context) *exchange.Params {
	var rv *exchange.Params
	k.iterate(ctx, GetKeyPrefixParamsSplit(), func(key, value []byte) bool {
		split, ok := uint16FromBz(value)
		if !ok {
			return false
		}
		if rv == nil {
			rv = &exchange.Params{}
		}
		denom := string(key)
		if len(denom) == 0 {
			rv.DefaultSplit = uint32(split)
		} else {
			rv.DenomSplits = append(rv.DenomSplits, exchange.DenomSplit{Denom: denom, Split: uint32(split)})
		}
		return false
	})
	return rv
}

// GetParamsOrDefaults gets the exchange module params from state if there are any.
// If state doesn't have any param info, the defaults are returned.
func (k Keeper) GetParamsOrDefaults(ctx sdk.Context) *exchange.Params {
	if rv := k.GetParams(ctx); rv != nil {
		return rv
	}
	return exchange.DefaultParams()
}

// GetExchangeSplit gets the split amount for the provided denom.
// If the denom is "", the default is returned.
// If there isn't a specific entry for the provided denom, the default is returned.
func (k Keeper) GetExchangeSplit(ctx sdk.Context, denom string) uint16 {
	store := k.getStore(ctx)
	if split, found := getParamsSplit(store, denom); found {
		return split
	}

	// If it wasn't found, and we weren't already looking for the default, look up the default now.
	if len(denom) > 0 {
		if split, found := getParamsSplit(store, ""); found {
			return split
		}
	}

	// If still not found, look to the hard-coded defaults.
	defaults := exchange.DefaultParams()

	// If looking for a specific denom, check the denom splits first.
	if len(denom) > 0 && len(defaults.DenomSplits) > 0 {
		for _, ds := range defaults.DenomSplits {
			if ds.Denom == denom {
				return uint16(ds.Split)
			}
		}
	}

	// Lastly, use the default from the defaults.
	return uint16(defaults.DefaultSplit)
}
