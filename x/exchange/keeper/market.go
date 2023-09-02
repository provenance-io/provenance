package keeper

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/exchange"
)

// flatFeeKeyMakers are the key and prefix maker funcs for a specific flat fee entry.
type flatFeeKeyMakers struct {
	key    func(marketID uint32, denom string) []byte
	prefix func(marketID uint32) []byte
}

type ratioKeyMakers struct {
	key    func(marketID uint32, ratio exchange.FeeRatio) []byte
	prefix func(marketID uint32) []byte
}

var (
	// createAskFlatKeyMakers are the key and prefix makers for the create-ask flat fees.
	createAskFlatKeyMakers = flatFeeKeyMakers{
		key:    MakeKeyMarketCreateAskFlatFee,
		prefix: GetKeyPrefixMarketCreateAskFlatFee,
	}
	// createBidFlatKeyMakers are the key and prefix makers for the create-bid flat fees.
	createBidFlatKeyMakers = flatFeeKeyMakers{
		key:    MakeKeyMarketCreateBidFlatFee,
		prefix: GetKeyPrefixMarketCreateBidFlatFee,
	}
	// sellerSettlementFlatKeyMakers are the key and prefix makers for the seller settlement flat fees.
	sellerSettlementFlatKeyMakers = flatFeeKeyMakers{
		key:    MakeKeyMarketSellerSettlementFlatFee,
		prefix: GetKeyPrefixMarketSellerSettlementFlatFee,
	}
	// sellerSettlementRatioKeyMakers are the key and prefix makers for the seller settlement fee ratios.
	sellerSettlementRatioKeyMakers = ratioKeyMakers{
		key:    MakeKeyMarketSellerSettlementRatio,
		prefix: GetKeyPrefixMarketSellerSettlementRatio,
	}
	// sellerSettlementFlatKeyMakers are the key and prefix makers for the buyer settlement flat fees.
	buyerSettlementFlatKeyMakers = flatFeeKeyMakers{
		key:    MakeKeyMarketBuyerSettlementFlatFee,
		prefix: GetKeyPrefixMarketBuyerSettlementFlatFee,
	}
	// buyerSettlementRatioKeyMakers are the key and prefix makers for the buyer settlement fee ratios.
	buyerSettlementRatioKeyMakers = ratioKeyMakers{
		key:    MakeKeyMarketBuyerSettlementRatio,
		prefix: GetKeyPrefixMarketBuyerSettlementRatio,
	}
)

// hasFlatFee returns true if this market has any flat fee for a given type.
func hasFlatFee(store sdk.KVStore, marketID uint32, maker flatFeeKeyMakers) bool {
	rv := false
	iterate(store, maker.prefix(marketID), func(key, value []byte) bool {
		rv = true
		return true
	})
	return rv
}

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
	var coins []sdk.Coin
	k.iterate(ctx, pre, func(key, value []byte) bool {
		amt, ok := sdkmath.NewIntFromString(string(value))
		if ok {
			coins = append(coins, sdk.Coin{Denom: string(key), Amount: amt})
		}
		return false
	})
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

// hasFeeRatio returns true if this market has any fee ratios for a given type.
func hasFeeRatio(store sdk.KVStore, marketID uint32, maker ratioKeyMakers) bool {
	rv := false
	iterate(store, maker.prefix(marketID), func(key, value []byte) bool {
		rv = true
		return true
	})
	return rv
}

// getFeeRatio is a generic getter for a fee ratio entry in the store.
func getFeeRatio(store sdk.KVStore, marketID uint32, priceDenom, feeDenom string, maker ratioKeyMakers) *exchange.FeeRatio {
	key := maker.key(marketID, exchange.FeeRatio{Price: sdk.Coin{Denom: priceDenom}, Fee: sdk.Coin{Denom: feeDenom}})
	if store.Has(key) {
		value := store.Get(key)
		priceAmount, feeAmount, err := ParseFeeRatioStoreValue(value)
		if err == nil {
			return &exchange.FeeRatio{
				Price: sdk.Coin{Denom: priceDenom, Amount: priceAmount},
				Fee:   sdk.Coin{Denom: feeDenom, Amount: feeAmount},
			}
		}
	}
	return nil
}

// setFeeRatio is a generic setter for a fee ratio entry in the store.
func setFeeRatio(store sdk.KVStore, marketID uint32, ratio exchange.FeeRatio, maker ratioKeyMakers) {
	key := maker.key(marketID, ratio)
	value := GetFeeRatioStoreValue(ratio)
	store.Set(key, value)
}

// getAllFeeRatios gets all the fee ratio entries from the store with the given prefix.
// The denoms come from the keys and amounts come from the values.
func (k Keeper) getAllFeeRatios(ctx sdk.Context, pre []byte) []exchange.FeeRatio {
	var feeRatios []exchange.FeeRatio
	k.iterate(ctx, pre, func(key, value []byte) bool {
		priceDenom, feeDenom, kerr := ParseKeySuffixSettlementRatio(key)
		if kerr == nil {
			priceAmount, feeAmount, verr := ParseFeeRatioStoreValue(value)
			if verr == nil {
				feeRatios = append(feeRatios, exchange.FeeRatio{
					Price: sdk.Coin{Denom: priceDenom, Amount: priceAmount},
					Fee:   sdk.Coin{Denom: feeDenom, Amount: feeAmount},
				})
			}
		}
		return false
	})
	return feeRatios
}

// setAllFeeRatios is a generic setter for a set of fee ratios.
// This will delete all previous options then save the ones provided. I.e. if ratios doesn't have a
// price/fee denom pair that currently exists in the store, those pairs will no longer be in the store after this.
func (k Keeper) setAllFeeRatios(ctx sdk.Context, marketID uint32, ratios []exchange.FeeRatio, maker ratioKeyMakers) {
	store := ctx.KVStore(k.storeKey)
	deleteAll(store, maker.prefix(marketID))
	for _, ratio := range ratios {
		setFeeRatio(store, marketID, ratio, maker)
	}
}

// SetCreateAskFlatFees sets the create-ask flat fees for a market.
func (k Keeper) SetCreateAskFlatFees(ctx sdk.Context, marketID uint32, options []sdk.Coin) {
	k.setAllCoins(ctx, marketID, options, createAskFlatKeyMakers)
}

// GetCreateAskFlatFees gets the create-ask flat fee options for a market.
func (k Keeper) GetCreateAskFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return k.getAllCoins(ctx, createAskFlatKeyMakers.prefix(marketID))
}

// IsAcceptableCreateAskFlatFee returns true if the provide fee has a denom accepted as a create-ask flat fee,
// and the fee amount is at least as much as the required amount of that denom.
func (k Keeper) IsAcceptableCreateAskFlatFee(ctx sdk.Context, marketID uint32, fee sdk.Coin) bool {
	reqFee := getFlatFee(ctx.KVStore(k.storeKey), marketID, fee.Denom, createAskFlatKeyMakers)
	return reqFee != nil && reqFee.Amount.LTE(fee.Amount)
}

// SetCreateBidFlatFees sets the create-bid flat fees for a market.
func (k Keeper) SetCreateBidFlatFees(ctx sdk.Context, marketID uint32, options []sdk.Coin) {
	k.setAllCoins(ctx, marketID, options, createBidFlatKeyMakers)
}

// GetCreateBidFlatFees gets the create-bid flat fee options for a market.
func (k Keeper) GetCreateBidFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return k.getAllCoins(ctx, createBidFlatKeyMakers.prefix(marketID))
}

// IsAcceptableCreateBidFlatFee returns true if the provide fee has a denom accepted as a create-bid flat fee,
// and the fee amount is at least as much as the required amount of that denom.
func (k Keeper) IsAcceptableCreateBidFlatFee(ctx sdk.Context, marketID uint32, fee sdk.Coin) bool {
	reqFee := getFlatFee(ctx.KVStore(k.storeKey), marketID, fee.Denom, createBidFlatKeyMakers)
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

// SetSellerSettlementRatios sets the seller settlement fee ratios for a market.
func (k Keeper) SetSellerSettlementRatios(ctx sdk.Context, marketID uint32, ratios []exchange.FeeRatio) {
	k.setAllFeeRatios(ctx, marketID, ratios, sellerSettlementRatioKeyMakers)
}

// GetSellerSettlementRatios gets the seller settlement fee ratios for a market.
func (k Keeper) GetSellerSettlementRatios(ctx sdk.Context, marketID uint32) []exchange.FeeRatio {
	return k.getAllFeeRatios(ctx, sellerSettlementRatioKeyMakers.prefix(marketID))
}

// CalculateSellerSettlementFeeFromRatio calculates the seller settlement fee required for the given price.
func (k Keeper) CalculateSellerSettlementFeeFromRatio(ctx sdk.Context, marketID uint32, price sdk.Coin) (sdk.Coin, error) {
	ratio := getFeeRatio(ctx.KVStore(k.storeKey), marketID, price.Denom, price.Denom, sellerSettlementRatioKeyMakers)
	if ratio == nil {
		return sdk.Coin{Amount: sdkmath.ZeroInt()}, fmt.Errorf("no seller settlement fee ratio found for denom %q", price.Denom)
	}
	rv, err := ratio.ApplyTo(price)
	if err != nil {
		err = fmt.Errorf("seller settlement fees: %w", err)
	}
	return rv, err
}

// SetBuyerSettlementFlatFees sets the buyer settlement flat fees for a market.
func (k Keeper) SetBuyerSettlementFlatFees(ctx sdk.Context, marketID uint32, options []sdk.Coin) {
	k.setAllCoins(ctx, marketID, options, buyerSettlementFlatKeyMakers)
}

// GetBuyerSettlementFlatFees gets the buyer settlement flat fee options for a market.
func (k Keeper) GetBuyerSettlementFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return k.getAllCoins(ctx, buyerSettlementFlatKeyMakers.prefix(marketID))
}

// SetBuyerSettlementRatios sets the buyer settlement fee ratios for a market.
func (k Keeper) SetBuyerSettlementRatios(ctx sdk.Context, marketID uint32, ratios []exchange.FeeRatio) {
	k.setAllFeeRatios(ctx, marketID, ratios, buyerSettlementRatioKeyMakers)
}

// GetBuyerSettlementRatios gets the buyer settlement fee ratios for a market.
func (k Keeper) GetBuyerSettlementRatios(ctx sdk.Context, marketID uint32) []exchange.FeeRatio {
	return k.getAllFeeRatios(ctx, buyerSettlementRatioKeyMakers.prefix(marketID))
}

// GetBuyerSettlementFeeRatiosForPriceDenom gets all the buyer settlement fee ratios in a market that have the give price denom.
func (k Keeper) GetBuyerSettlementFeeRatiosForPriceDenom(ctx sdk.Context, marketID uint32, priceDenom string) []exchange.FeeRatio {
	var ratios []exchange.FeeRatio
	k.iterate(ctx, GetKeyPrefixMarketBuyerSettlementRatioForPriceDenom(marketID, priceDenom), func(key, value []byte) bool {
		feeDenom := string(key)
		priceAmount, feeAmount, err := ParseFeeRatioStoreValue(value)
		if err == nil {
			ratios = append(ratios, exchange.FeeRatio{
				Price: sdk.Coin{Denom: priceDenom, Amount: priceAmount},
				Fee:   sdk.Coin{Denom: feeDenom, Amount: feeAmount},
			})
		}
		return false
	})
	return ratios
}

// CalculateBuyerSettlementFeeOptionsFromRatios calculates the buyer settlement fee options available for the given price.
func (k Keeper) CalculateBuyerSettlementFeeOptionsFromRatios(ctx sdk.Context, marketID uint32, price sdk.Coin) ([]sdk.Coin, error) {
	ratios := k.GetBuyerSettlementFeeRatiosForPriceDenom(ctx, marketID, price.Denom)
	if len(ratios) == 0 {
		return nil, fmt.Errorf("no buyer settlement fee ratios found with price denom %q", price.Denom)
	}

	var errs []error
	rv := make([]sdk.Coin, 0, len(ratios))
	for _, ratio := range ratios {
		fee, err := ratio.ApplyTo(price)
		if err != nil {
			errs = append(errs, fmt.Errorf("buyer settlement fees: %w", err))
		} else {
			rv = append(rv, fee)
		}
	}

	// We only return errors if no options are available.
	if len(rv) == 0 {
		errs = append(errs, fmt.Errorf("no applicable buyer settlement fee ratios found for price %s", price))
		return nil, errors.Join(errs...)
	}
	return rv, nil
}

// IsAcceptableBuyerSettlementFee returns true if the provided fee is enough to cover both the buyer settlement flat and percent fees.
func (k Keeper) IsAcceptableBuyerSettlementFee(ctx sdk.Context, marketID uint32, price sdk.Coin, fee sdk.Coins) bool {
	flatKeyMaker := buyerSettlementFlatKeyMakers
	ratioKeyMaker := buyerSettlementRatioKeyMakers
	store := ctx.KVStore(k.storeKey)
	flatFeeReq := hasFlatFee(store, marketID, flatKeyMaker)
	ratioFeeReq := hasFeeRatio(store, marketID, ratioKeyMaker)

	if !flatFeeReq && !ratioFeeReq {
		return true
	}

	switch len(fee) {
	case 0:
		return !flatFeeReq && !ratioFeeReq
	case 1:
		reqAmt := sdkmath.ZeroInt()

		if flatFeeReq {
			flatFee := getFlatFee(store, marketID, fee[0].Denom, flatKeyMaker)
			if flatFee == nil {
				return false
			}
			reqAmt = reqAmt.Add(flatFee.Amount)
		}

		if ratioFeeReq {
			ratio := getFeeRatio(store, marketID, price.Denom, fee[0].Denom, ratioKeyMaker)
			if ratio == nil {
				return false
			}
			if ratio != nil {
				ratioFee, err := ratio.ApplyTo(price)
				if err != nil {
					return false
				}
				reqAmt = reqAmt.Add(ratioFee.Amount)
			}
		}

		return reqAmt.LTE(fee[0].Amount)
	default:
		// Find all the flat fee options available for each fee denom,
		// and all ratio fee options available for each price denom to fee denom ratio.
		var flatFees []sdk.Coin
		var ratioFees []sdk.Coin
		for _, coin := range fee {
			var flatFee, ratioFee *sdk.Coin

			if flatFeeReq {
				flatFee = getFlatFee(store, marketID, coin.Denom, flatKeyMaker)
				if flatFee != nil && flatFee.Amount.LTE(coin.Amount) {
					// If there aren't any ratio fees required, this fee entry covers all required fees.
					if !ratioFeeReq {
						return true
					}
					// Keep track of this satisfaction. Yeah.
					flatFees = append(flatFees, *flatFee)
				} else {
					flatFee = nil
				}
			}

			if ratioFeeReq {
				ratio := getFeeRatio(store, marketID, price.Denom, coin.Denom, ratioKeyMaker)
				if ratio != nil {
					ratioFeeCalc, err := ratio.ApplyTo(price)
					if err == nil {
						// If there aren't any flat fees required, this fee entry covers all required fees.
						if !flatFeeReq {
							return true
						}
						// Keep track of this satisfaction. Uh huh.
						ratioFees = append(ratioFees, ratioFeeCalc)
						ratioFee = &ratioFeeCalc
					}
				}
			}

			// If this coin's denom is an option for BOTH flat and ratio portions, and is enough to cover the sum,
			// this one coin satisfies both together, and we're good to go.
			if flatFee != nil && ratioFee != nil && flatFee.Amount.Add(ratioFee.Amount).LTE(coin.Amount) {
				return true
			}
		}

		// If there's no flat fees, but one is required, the flat fee isn't covered in the fee provided.
		if len(flatFees) == 0 && hasFlatFee(store, marketID, flatKeyMaker) {
			return false
		}
		// If there's no ratio fees, but one is required, the ratio fee isn't covered in the fee provided.
		if len(ratioFees) == 0 && hasFeeRatio(store, marketID, ratioKeyMaker) {
			return false
		}
		// If there's exactly one of each, and they have the same denom,
		// the fee of that denom was NOT enough to cover both together.
		if len(flatFees) == 1 && len(ratioFees) == 1 && flatFees[0].Denom == ratioFees[0].Denom {
			return false
		}

		// Flat fee requirements are met because either:
		//  a) no flat fee was required.
		//  b) one or more fee coins are enough to cover the flat fee.
		// Ratio fee requirements are met because either:
		//  a) no ratio fee was required.
		//  b) one or more fee coins are enough to cover the ratio fee.
		// We also know that, if there's exactly one of each, they have different denoms, so
		// it's NOT the case where (flat or ratio fee) <= (the other type of fee) <= fee provided <= sum of required fees.
		// I.e. if there are exactly one of each, they each satisfy a different fee type requirement.
		return true
	}
}
