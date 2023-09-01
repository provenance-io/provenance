package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/exchange"
)

// deleteAllParamsSplits deletes all the params splits in the store.
func deleteAllParamsSplits(store sdk.KVStore) {
	keys := getAllKeys(store, MakeKeyPrefixParamsSplit())
	for _, key := range keys {
		store.Delete(key)
	}
}

// setParamsSplit sets the provided params split for the provided denom.
func setParamsSplit(store sdk.KVStore, denom string, split uint16) {
	key := MakeKeyParamsSplit(denom)
	value := uint16Bz(split)
	store.Set(key, value)
}

// getParamsSplit gets the params split amount for the provided denom, and whether the entry existed.
func getParamsSplit(store sdk.KVStore, denom string) (uint16, bool) {
	key := MakeKeyParamsSplit(denom)
	if store.Has(key) {
		value := store.Get(key)
		return uint16FromBz(value), true
	}
	return 0, false
}

// SetParams updates the params to match those provided.
// If nil is provided, all params are deleted.
func (k Keeper) SetParams(ctx sdk.Context, params *exchange.Params) {
	store := ctx.KVStore(k.storeKey)
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
	// Using an open iterator on a prefixed store here so that iter.Key() doesn't contain the prefix.
	store := prefix.NewStore(ctx.KVStore(k.storeKey), MakeKeyPrefixParamsSplit())
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	var rv *exchange.Params
	for ; iter.Valid(); iter.Next() {
		denom := string(iter.Key())
		split := uint16FromBz(iter.Value())

		if rv == nil {
			rv = &exchange.Params{}
		}
		if len(denom) == 0 {
			rv.DefaultSplit = uint32(split)
		} else {
			rv.DenomSplits = append(rv.DenomSplits, exchange.DenomSplit{Denom: denom, Split: uint32(split)})
		}
	}

	return rv
}

// GetExchangeSplit gets the split amount for the provided denom.
// If the denom is "", the default is returned.
// If there isn't a specific entry for the provided denom, the default is returned.
func (k Keeper) GetExchangeSplit(ctx sdk.Context, denom string) uint16 {
	store := ctx.KVStore(k.storeKey)
	split, found := getParamsSplit(store, denom)
	// If it wasn't found, and we weren't already looking for the default, look up the default now.
	if !found && len(denom) > 0 {
		split, _ = getParamsSplit(store, "")
	}
	return split
}
