package keeper

import (
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/store/prefix"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// flatFeeKeyMakers are the key and prefix maker funcs for a specific flat fee entry.
type flatFeeKeyMakers struct {
	key    func(marketID uint32, denom string) []byte
	prefix func(marketID uint32) []byte
}

var (
	// createAskKeyMakers are the key and prefix makers for the create-ask flat fees.
	createAskKeyMakers = flatFeeKeyMakers{
		key:    MakeKeyMarketCreateAskFlatFee,
		prefix: GetKeyPrefixMarketCreateAskFlatFee,
	}
	// createBidKeyMakers are the key and prefix makers for the create-bid flat fees.
	createBidKeyMakers = flatFeeKeyMakers{
		key:    MakeKeyMarketCreateBidFlatFee,
		prefix: GetKeyPrefixMarketCreateBidFlatFee,
	}
	// sellerSettlementFlatKeyMakers are the key and prefix makers for the seller settlement flat fees.
	sellerSettlementFlatKeyMakers = flatFeeKeyMakers{
		key:    MakeKeyMarketSellerSettlementFlatFee,
		prefix: GetKeyPrefixMarketSellerSettlementFlatFee,
	}
	// sellerSettlementFlatKeyMakers are the key and prefix makers for the buyer settlement flat fees.
	buyerSettlementFlatKeyMakers = flatFeeKeyMakers{
		key:    MakeKeyMarketBuyerSettlementFlatFee,
		prefix: GetKeyPrefixMarketBuyerSettlementFlatFee,
	}
)

// getFlatFee is a generic getter for a flat fee coin entry.
func getFlatFee(store sdk.KVStore, marketID uint32, denom string, maker flatFeeKeyMakers) *sdk.Coin {
	key := maker.key(marketID, denom)
	if store.Has(key) {
		value := string(store.Get(key))
		amt, ok := sdkmath.NewIntFromString(value)
		if ok {
			return &sdk.Coin{Denom: denom, Amount: amt}
		}
	}
	return nil
}

// setFlatFee is a generic setter for a single flat fee coin entry.
func setFlatFee(store sdk.KVStore, marketID uint32, coin sdk.Coin, maker flatFeeKeyMakers) {
	key := maker.key(marketID, coin.Denom)
	value := coin.Amount.String()
	store.Set(key, []byte(value))
}

// getAllCoins gets all the coin entries from the store with the given prefix.
// The denom comes from the part of the key after the prefix, and the amount comes from the values.
func (k Keeper) getAllCoins(ctx sdk.Context, pre []byte) []sdk.Coin {
	// Using a prefixed store here so that iter.Key() doesn't contain the prefix (just the denom).
	pStore := prefix.NewStore(ctx.KVStore(k.storeKey), pre)
	iter := pStore.Iterator(nil, nil)
	defer iter.Close()

	var coins []sdk.Coin
	for ; iter.Valid(); iter.Next() {
		denom := string(iter.Key())
		value := string(iter.Value())
		amt, ok := sdkmath.NewIntFromString(value)
		if !ok {
			continue
		}
		coins = append(coins, sdk.Coin{Denom: denom, Amount: amt})
	}

	return coins
}

// setAllCoins is a generic setter for a set of flat fee options.
// This will delete all previous options then save the ones provided. I.e. if options doesn't have a
// denom that currently exists in the store, those denoms will no longer be in the store after this.
func (k Keeper) setAllCoins(ctx sdk.Context, marketID uint32, options []sdk.Coin, maker flatFeeKeyMakers) {
	store := ctx.KVStore(k.storeKey)
	deleteAll(store, maker.prefix(marketID))
	for _, coin := range options {
		setFlatFee(store, marketID, coin, maker)
	}
}

// SetCreateAskFlatFees sets the create-ask flat fees for a market.
func (k Keeper) SetCreateAskFlatFees(ctx sdk.Context, marketID uint32, options []sdk.Coin) {
	k.setAllCoins(ctx, marketID, options, createAskKeyMakers)
}

// GetCreateAskFlatFees gets the create-ask flat fee options for a market.
func (k Keeper) GetCreateAskFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return k.getAllCoins(ctx, createAskKeyMakers.prefix(marketID))
}

// IsAcceptableCreateAskFlatFee returns true if the provide fee has a denom accepted as a create-ask flat fee,
// and the fee amount is at least as much as the required amount of that denom.
func (k Keeper) IsAcceptableCreateAskFlatFee(ctx sdk.Context, marketID uint32, fee sdk.Coin) bool {
	reqFee := getFlatFee(ctx.KVStore(k.storeKey), marketID, fee.Denom, createAskKeyMakers)
	return reqFee != nil && reqFee.Amount.LTE(fee.Amount)
}

// SetCreateBidFlatFees sets the create-bid flat fees for a market.
func (k Keeper) SetCreateBidFlatFees(ctx sdk.Context, marketID uint32, options []sdk.Coin) {
	k.setAllCoins(ctx, marketID, options, createBidKeyMakers)
}

// GetCreateBidFlatFees gets the create-bid flat fee options for a market.
func (k Keeper) GetCreateBidFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return k.getAllCoins(ctx, createBidKeyMakers.prefix(marketID))
}

// IsAcceptableCreateBidFlatFee returns true if the provide fee has a denom accepted as a create-bid flat fee,
// and the fee amount is at least as much as the required amount of that denom.
func (k Keeper) IsAcceptableCreateBidFlatFee(ctx sdk.Context, marketID uint32, fee sdk.Coin) bool {
	reqFee := getFlatFee(ctx.KVStore(k.storeKey), marketID, fee.Denom, createBidKeyMakers)
	return reqFee != nil && reqFee.Amount.LTE(fee.Amount)
}

// SetSellerSettlementFlatFees sets the seller settlement flat fees for a market.
func (k Keeper) SetSellerSettlementFlatFees(ctx sdk.Context, marketID uint32, options []sdk.Coin) {
	k.setAllCoins(ctx, marketID, options, sellerSettlementFlatKeyMakers)
}

// GetSellerSettlementFlatFees gets the seller settlement flat fee options for a market.
func (k Keeper) GetSellerSettlementFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return k.getAllCoins(ctx, sellerSettlementFlatKeyMakers.prefix(marketID))
}

// IsAcceptableSellerSettlementFlatFee returns true if the provide fee has a denom accepted as a seller settlement
// flat fee, and the fee amount is at least as much as the required amount of that denom.
func (k Keeper) IsAcceptableSellerSettlementFlatFee(ctx sdk.Context, marketID uint32, fee sdk.Coin) bool {
	reqFee := getFlatFee(ctx.KVStore(k.storeKey), marketID, fee.Denom, sellerSettlementFlatKeyMakers)
	return reqFee != nil && reqFee.Amount.LTE(fee.Amount)
}

// SetBuyerSettlementFlatFees sets the buyer settlement flat fees for a market.
func (k Keeper) SetBuyerSettlementFlatFees(ctx sdk.Context, marketID uint32, options []sdk.Coin) {
	k.setAllCoins(ctx, marketID, options, buyerSettlementFlatKeyMakers)
}

// GetBuyerSettlementFlatFees gets the buyer settlement flat fee options for a market.
func (k Keeper) GetBuyerSettlementFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return k.getAllCoins(ctx, buyerSettlementFlatKeyMakers.prefix(marketID))
}
