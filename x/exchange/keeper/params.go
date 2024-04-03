package keeper

import (
	"strings"

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

// getParamsSplits gets the default split and all denom splits from state.
// Returns the default split amount, the specific denom splits, and whether there were any entries in state.
// If there are no splits entries in state (default or specific), returns 0, nil, false.
func getParamsSplits(store storetypes.KVStore) (uint32, []exchange.DenomSplit, bool) {
	var defaultSplit uint32
	var denomSplits []exchange.DenomSplit
	var haveVals bool

	iterate(store, GetKeyPrefixParamsSplit(), func(key, value []byte) bool {
		split, ok := uint16FromBz(value)
		if !ok {
			return false
		}

		haveVals = true
		denom := string(key)
		if len(denom) == 0 {
			defaultSplit = uint32(split)
		} else {
			denomSplits = append(denomSplits, exchange.DenomSplit{Denom: denom, Split: uint32(split)})
		}

		return false
	})

	return defaultSplit, denomSplits, haveVals
}

// setParamsFeePaymentFlat sets a payment flat fee params entry.
func setParamsFeePaymentFlat(store storetypes.KVStore, key []byte, opts []sdk.Coin) {
	if len(opts) == 0 || sdk.Coins(opts).IsZero() {
		store.Delete(key)
		return
	}
	val := sdk.Coins(opts).String()
	store.Set(key, []byte(val))
}

// getParamsPaymentFlatFee gets a payment flat fee params entry.
func getParamsPaymentFlatFee(store storetypes.KVStore, key []byte) []sdk.Coin {
	val := store.Get(key)
	if len(val) == 0 {
		return nil
	}
	// If we used sdk.ParseCoinsNormalized() here, they'd be sorted by denom in the result.
	// But we want to maintain the order of the entries, so do it a little differently.
	entries := strings.Split(string(val), ",")
	rv := make([]sdk.Coin, 0, len(entries))
	for _, entry := range entries {
		coin, err := exchange.ParseCoin(entry)
		if err == nil && !coin.IsZero() {
			rv = append(rv, coin)
		}
	}
	return rv
}

// setParamsFeeCreatePaymentFlat sets the params entry for the create-payment flat fee.
func setParamsFeeCreatePaymentFlat(store storetypes.KVStore, opts []sdk.Coin) {
	setParamsFeePaymentFlat(store, MakeKeyParamsFeeCreatePaymentFlat(), opts)
}

// setParamsFeeCreatePaymentFlat sets the params entry for the accept-payment flat fee.
func setParamsFeeAcceptPaymentFlat(store storetypes.KVStore, opts []sdk.Coin) {
	setParamsFeePaymentFlat(store, MakeKeyParamsFeeAcceptPaymentFlat(), opts)
}

// getParamsFeeCreatePaymentFlat gets the params entry for the create-payment flat fee.
func getParamsFeeCreatePaymentFlat(store storetypes.KVStore) []sdk.Coin {
	return getParamsPaymentFlatFee(store, MakeKeyParamsFeeCreatePaymentFlat())
}

// getParamsFeeAcceptPaymentFlat gets the params entry for the accept-payment flat fee.
func getParamsFeeAcceptPaymentFlat(store storetypes.KVStore) []sdk.Coin {
	return getParamsPaymentFlatFee(store, MakeKeyParamsFeeAcceptPaymentFlat())
}

// SetParams updates the params to match those provided.
// If nil is provided, all params are deleted.
func (k Keeper) SetParams(ctx sdk.Context, params *exchange.Params) {
	store := k.getStore(ctx)

	deleteAllParamsSplits(store)
	var feeCreate, feeAccept []sdk.Coin
	if params != nil {
		setParamsSplit(store, "", uint16(params.DefaultSplit))
		for _, split := range params.DenomSplits {
			setParamsSplit(store, split.Denom, uint16(split.Split))
		}
		feeCreate = params.FeeCreatePaymentFlat
		feeAccept = params.FeeAcceptPaymentFlat
	}

	setParamsFeeCreatePaymentFlat(store, feeCreate)
	setParamsFeeAcceptPaymentFlat(store, feeAccept)
}

// GetParams gets the exchange module params.
// If there aren't any params in state, nil is returned.
func (k Keeper) GetParams(ctx sdk.Context) *exchange.Params {
	store := k.getStore(ctx)

	var rv *exchange.Params
	if defaultSplit, denomSplits, haveSplits := getParamsSplits(store); haveSplits {
		rv = &exchange.Params{
			DefaultSplit: defaultSplit,
			DenomSplits:  denomSplits,
		}
	}

	if opts := getParamsFeeCreatePaymentFlat(store); len(opts) > 0 {
		if rv == nil {
			rv = &exchange.Params{}
		}
		rv.FeeCreatePaymentFlat = opts
	}

	if opts := getParamsFeeAcceptPaymentFlat(store); len(opts) > 0 {
		if rv == nil {
			rv = &exchange.Params{}
		}
		rv.FeeAcceptPaymentFlat = opts
	}

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
