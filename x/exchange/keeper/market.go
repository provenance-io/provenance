package keeper

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// deleteCreateAskFlatFees deletes all the create-ask flat fees for the given market.
func deleteCreateAskFlatFees(store sdk.KVStore, marketID uint32) {
	keys := getAllKeys(store, MakeKeyPrefixMarketCreateAskFlatFee(marketID))
	for _, key := range keys {
		store.Delete(key)
	}
}

// setCreateAskFlatFee sets a create-ask flat fee option in the store.
func setCreateAskFlatFee(store sdk.KVStore, marketID uint32, coin sdk.Coin) {
	key := MakeKeyMarketCreateAskFlatFee(marketID, coin.Denom)
	value := coin.Amount.String()
	store.Set(key, []byte(value))
}

// getCreateAskFlatFee gets a create-ask flat fee option from the store.
// Returns nil if either no entry for that denom exists, or the entry does exist, but can't be read.
func getCreateAskFlatFee(store sdk.KVStore, marketID uint32, denom string) *sdk.Coin {
	key := MakeKeyMarketCreateAskFlatFee(marketID, denom)
	if store.Has(key) {
		value := string(store.Get(key))
		amt, ok := sdkmath.NewIntFromString(value)
		if ok {
			return &sdk.Coin{Denom: denom, Amount: amt}
		}
	}
	return nil
}

// SetCreateAskFlatFees sets the create-ask flat fees for a market.
func (k Keeper) SetCreateAskFlatFees(ctx sdk.Context, marketID uint32, options []sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	deleteCreateAskFlatFees(store, marketID)

	for _, coin := range options {
		setCreateAskFlatFee(store, marketID, coin)
	}
}

// GetCreateAskFlatFees gets the create-ask flat fees for a market.
func (k Keeper) GetCreateAskFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return getAllCoins(ctx.KVStore(k.storeKey), MakeKeyPrefixMarketCreateAskFlatFee(marketID))
}

// IsAcceptableCreateAskFlatFee returns true if the provide fee has a denom accepted as a create-ask flat fee,
// and the fee amount is at least as much as the required amount of that denom.
func (k Keeper) IsAcceptableCreateAskFlatFee(ctx sdk.Context, marketID uint32, fee sdk.Coin) bool {
	reqFee := getCreateAskFlatFee(ctx.KVStore(k.storeKey), marketID, fee.Denom)
	if reqFee == nil {
		return false
	}
	return reqFee.Amount.LTE(fee.Amount)
}

// deleteCreateBidFlatFees deletes all the create-bid flat fees for the given market.
func deleteCreateBidFlatFees(store sdk.KVStore, marketID uint32) {
	keys := getAllKeys(store, MakeKeyPrefixMarketCreateBidFlatFee(marketID))
	for _, key := range keys {
		store.Delete(key)
	}
}

// setCreateBidFlatFee sets a create-Bid flat fee option in the store.
func setCreateBidFlatFee(store sdk.KVStore, marketID uint32, coin sdk.Coin) {
	key := MakeKeyMarketCreateBidFlatFee(marketID, coin.Denom)
	value := coin.Amount.String()
	store.Set(key, []byte(value))
}

// getCreateBidFlatFee gets a create-bid flat fee option from the store.
// Returns nil if either no entry for that denom exists, or the entry does exist, but can't be read.
func getCreateBidFlatFee(store sdk.KVStore, marketID uint32, denom string) *sdk.Coin {
	key := MakeKeyMarketCreateBidFlatFee(marketID, denom)
	if store.Has(key) {
		value := string(store.Get(key))
		amt, ok := sdkmath.NewIntFromString(value)
		if ok {
			return &sdk.Coin{Denom: denom, Amount: amt}
		}
	}
	return nil
}

// SetCreateBidFlatFees sets the create-bid flat fees for a market.
func (k Keeper) SetCreateBidFlatFees(ctx sdk.Context, marketID uint32, options []sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	deleteCreateBidFlatFees(store, marketID)

	for _, coin := range options {
		setCreateBidFlatFee(store, marketID, coin)
	}
}

// GetCreateBidFlatFees gets the create-bid flat fees for a market.
func (k Keeper) GetCreateBidFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return getAllCoins(ctx.KVStore(k.storeKey), MakeKeyPrefixMarketCreateBidFlatFee(marketID))
}

// IsAcceptableCreateBidFlatFee returns true if the provide fee has a denom accepted as a create-bid flat fee,
// and the fee amount is at least as much as the required amount of that denom.
func (k Keeper) IsAcceptableCreateBidFlatFee(ctx sdk.Context, marketID uint32, fee sdk.Coin) bool {
	reqFee := getCreateBidFlatFee(ctx.KVStore(k.storeKey), marketID, fee.Denom)
	if reqFee == nil {
		return false
	}
	return reqFee.Amount.LTE(fee.Amount)
}
