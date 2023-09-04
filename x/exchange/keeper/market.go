package keeper

import (
	"errors"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/exchange"
)

// flatFeeKeyMakers are the key and prefix maker funcs for a specific flat fee entry.
type flatFeeKeyMakers struct {
	key    func(marketID uint32, denom string) []byte
	prefix func(marketID uint32) []byte
}

// ratioKeyMakers are the key and prefix maker funcs for a specific ratio fee entry.
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

// getAllFlatFees gets all the coin entries from the store with the given prefix.
// The denom comes from the part of the key after the prefix, and the amount comes from the values.
func getAllFlatFees(store sdk.KVStore, marketID uint32, maker flatFeeKeyMakers) []sdk.Coin {
	var coins []sdk.Coin
	iterate(store, maker.prefix(marketID), func(key, value []byte) bool {
		amt, ok := sdkmath.NewIntFromString(string(value))
		if ok {
			coins = append(coins, sdk.Coin{Denom: string(key), Amount: amt})
		}
		return false
	})
	return coins
}

// setAllFlatFees is a generic setter for a set of flat fee options.
// This will delete all previous options then save the ones provided. I.e. if options doesn't have a
// denom that currently exists in the store, those denoms will no longer be in the store after this.
func setAllFlatFees(store sdk.KVStore, marketID uint32, options []sdk.Coin, maker flatFeeKeyMakers) {
	deleteAll(store, maker.prefix(marketID))
	for _, coin := range options {
		setFlatFee(store, marketID, coin, maker)
	}
}

// updateFlatFees deletes all the entries with a denom in toDelete, then writes all the toWrite entries.
func updateFlatFees(store sdk.KVStore, marketID uint32, toDelete, toWrite []sdk.Coin, maker flatFeeKeyMakers) {
	for _, coin := range toDelete {
		key := maker.key(marketID, coin.Denom)
		store.Delete(key)
	}
	for _, coin := range toWrite {
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
func getAllFeeRatios(store sdk.KVStore, marketID uint32, maker ratioKeyMakers) []exchange.FeeRatio {
	var feeRatios []exchange.FeeRatio
	iterate(store, maker.prefix(marketID), func(key, value []byte) bool {
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
func setAllFeeRatios(store sdk.KVStore, marketID uint32, ratios []exchange.FeeRatio, maker ratioKeyMakers) {
	deleteAll(store, maker.prefix(marketID))
	for _, ratio := range ratios {
		setFeeRatio(store, marketID, ratio, maker)
	}
}

// updateFeeRatios deletes all entries with the denom pairs in toDelete, then writes all the toWrite entries.
func updateFeeRatios(store sdk.KVStore, marketID uint32, toDelete, toWrite []exchange.FeeRatio, maker ratioKeyMakers) {
	for _, ratio := range toDelete {
		key := maker.key(marketID, ratio)
		store.Delete(key)
	}
	for _, ratio := range toWrite {
		setFeeRatio(store, marketID, ratio, maker)
	}
}

// GetCreateAskFlatFees gets the create-ask flat fee options for a market.
func (k Keeper) GetCreateAskFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return getAllFlatFees(k.getStore(ctx), marketID, createAskFlatKeyMakers)
}

// SetCreateAskFlatFees sets the create-ask flat fees for a market.
func (k Keeper) SetCreateAskFlatFees(ctx sdk.Context, marketID uint32, options []sdk.Coin) {
	setAllFlatFees(k.getStore(ctx), marketID, options, createAskFlatKeyMakers)
}

// UpdateCreateAskFlatFees deletes all create-ask flat fees to delete then adds the ones to add.
func (k Keeper) UpdateCreateAskFlatFees(ctx sdk.Context, marketID uint32, toDelete, toAdd []sdk.Coin) {
	updateFlatFees(k.getStore(ctx), marketID, toDelete, toAdd, createAskFlatKeyMakers)
}

// IsAcceptableCreateAskFlatFee returns true if the provide fee has a denom accepted as a create-ask flat fee,
// and the fee amount is at least as much as the required amount of that denom.
func (k Keeper) IsAcceptableCreateAskFlatFee(ctx sdk.Context, marketID uint32, fee sdk.Coin) bool {
	reqFee := getFlatFee(k.getStore(ctx), marketID, fee.Denom, createAskFlatKeyMakers)
	return reqFee != nil && fee.Amount.GTE(reqFee.Amount)
}

// GetCreateBidFlatFees gets the create-bid flat fee options for a market.
func (k Keeper) GetCreateBidFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return getAllFlatFees(k.getStore(ctx), marketID, createBidFlatKeyMakers)
}

// SetCreateBidFlatFees sets the create-bid flat fees for a market.
func (k Keeper) SetCreateBidFlatFees(ctx sdk.Context, marketID uint32, options []sdk.Coin) {
	setAllFlatFees(k.getStore(ctx), marketID, options, createBidFlatKeyMakers)
}

// UpdateCreateBidFlatFees deletes all create-bid flat fees to delete then adds the ones to add.
func (k Keeper) UpdateCreateBidFlatFees(ctx sdk.Context, marketID uint32, toDelete, toAdd []sdk.Coin) {
	updateFlatFees(k.getStore(ctx), marketID, toDelete, toAdd, createBidFlatKeyMakers)
}

// IsAcceptableCreateBidFlatFee returns true if the provide fee has a denom accepted as a create-bid flat fee,
// and the fee amount is at least as much as the required amount of that denom.
func (k Keeper) IsAcceptableCreateBidFlatFee(ctx sdk.Context, marketID uint32, fee sdk.Coin) bool {
	reqFee := getFlatFee(k.getStore(ctx), marketID, fee.Denom, createBidFlatKeyMakers)
	return reqFee != nil && fee.Amount.GTE(reqFee.Amount)
}

// GetSellerSettlementFlatFees gets the seller settlement flat fee options for a market.
func (k Keeper) GetSellerSettlementFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return getAllFlatFees(k.getStore(ctx), marketID, sellerSettlementFlatKeyMakers)
}

// SetSellerSettlementFlatFees sets the seller settlement flat fees for a market.
func (k Keeper) SetSellerSettlementFlatFees(ctx sdk.Context, marketID uint32, options []sdk.Coin) {
	setAllFlatFees(k.getStore(ctx), marketID, options, sellerSettlementFlatKeyMakers)
}

// UpdateSellerSettlementFlatFees deletes all seller settlement flat fees to delete then adds the ones to add.
func (k Keeper) UpdateSellerSettlementFlatFees(ctx sdk.Context, marketID uint32, toDelete, toAdd []sdk.Coin) {
	updateFlatFees(k.getStore(ctx), marketID, toDelete, toAdd, sellerSettlementFlatKeyMakers)
}

// IsAcceptableSellerSettlementFlatFee returns true if the provide fee has a denom accepted as a seller settlement
// flat fee, and the fee amount is at least as much as the required amount of that denom.
func (k Keeper) IsAcceptableSellerSettlementFlatFee(ctx sdk.Context, marketID uint32, fee sdk.Coin) bool {
	reqFee := getFlatFee(k.getStore(ctx), marketID, fee.Denom, sellerSettlementFlatKeyMakers)
	return reqFee != nil && fee.Amount.GTE(reqFee.Amount)
}

// GetSellerSettlementRatios gets the seller settlement fee ratios for a market.
func (k Keeper) GetSellerSettlementRatios(ctx sdk.Context, marketID uint32) []exchange.FeeRatio {
	return getAllFeeRatios(k.getStore(ctx), marketID, sellerSettlementRatioKeyMakers)
}

// SetSellerSettlementRatios sets the seller settlement fee ratios for a market.
func (k Keeper) SetSellerSettlementRatios(ctx sdk.Context, marketID uint32, ratios []exchange.FeeRatio) {
	setAllFeeRatios(k.getStore(ctx), marketID, ratios, sellerSettlementRatioKeyMakers)
}

// UpdateSellerSettlementRatios deletes all seller settlement ratio entries to delete then adds the ones to add.
func (k Keeper) UpdateSellerSettlementRatios(ctx sdk.Context, marketID uint32, toDelete, toAdd []exchange.FeeRatio) {
	updateFeeRatios(k.getStore(ctx), marketID, toDelete, toAdd, sellerSettlementRatioKeyMakers)
}

// CalculateSellerSettlementRatioFee calculates the seller settlement fee required for the given price.
func (k Keeper) CalculateSellerSettlementRatioFee(ctx sdk.Context, marketID uint32, price sdk.Coin) (sdk.Coin, error) {
	ratio := getFeeRatio(k.getStore(ctx), marketID, price.Denom, price.Denom, sellerSettlementRatioKeyMakers)
	if ratio == nil {
		return sdk.Coin{}, fmt.Errorf("no seller settlement fee ratio found for denom %q", price.Denom)
	}
	rv, err := ratio.ApplyTo(price)
	if err != nil {
		err = fmt.Errorf("invalid seller settlement fees: %w", err)
	}
	return rv, err
}

// GetBuyerSettlementFlatFees gets the buyer settlement flat fee options for a market.
func (k Keeper) GetBuyerSettlementFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return getAllFlatFees(k.getStore(ctx), marketID, buyerSettlementFlatKeyMakers)
}

// SetBuyerSettlementFlatFees sets the buyer settlement flat fees for a market.
func (k Keeper) SetBuyerSettlementFlatFees(ctx sdk.Context, marketID uint32, options []sdk.Coin) {
	setAllFlatFees(k.getStore(ctx), marketID, options, buyerSettlementFlatKeyMakers)
}

// UpdateBuyerSettlementFlatFees deletes all buyer settlement flat fees to delete then adds the ones to add.
func (k Keeper) UpdateBuyerSettlementFlatFees(ctx sdk.Context, marketID uint32, toDelete, toAdd []sdk.Coin) {
	updateFlatFees(k.getStore(ctx), marketID, toDelete, toAdd, buyerSettlementFlatKeyMakers)
}

// GetBuyerSettlementRatios gets the buyer settlement fee ratios for a market.
func (k Keeper) GetBuyerSettlementRatios(ctx sdk.Context, marketID uint32) []exchange.FeeRatio {
	return getAllFeeRatios(k.getStore(ctx), marketID, buyerSettlementRatioKeyMakers)
}

// SetBuyerSettlementRatios sets the buyer settlement fee ratios for a market.
func (k Keeper) SetBuyerSettlementRatios(ctx sdk.Context, marketID uint32, ratios []exchange.FeeRatio) {
	setAllFeeRatios(k.getStore(ctx), marketID, ratios, buyerSettlementRatioKeyMakers)
}

// UpdateBuyerSettlementRatios deletes all buyer settlement ratio entries to delete then adds the ones to add.
func (k Keeper) UpdateBuyerSettlementRatios(ctx sdk.Context, marketID uint32, toDelete, toAdd []exchange.FeeRatio) {
	updateFeeRatios(k.getStore(ctx), marketID, toDelete, toAdd, buyerSettlementRatioKeyMakers)
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

// CalculateBuyerSettlementRatioFeeOptions calculates the buyer settlement ratio fee options available for the given price.
func (k Keeper) CalculateBuyerSettlementRatioFeeOptions(ctx sdk.Context, marketID uint32, price sdk.Coin) ([]sdk.Coin, error) {
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

// ValidateBuyerSettlementFee returns an error if the provided fee is not enough to cover both the
// buyer settlement flat and percent fees for the given price.
func (k Keeper) ValidateBuyerSettlementFee(ctx sdk.Context, marketID uint32, price sdk.Coin, fee sdk.Coins) error {
	flatKeyMaker := buyerSettlementFlatKeyMakers
	ratioKeyMaker := buyerSettlementRatioKeyMakers
	store := k.getStore(ctx)
	flatFeeReq := hasFlatFee(store, marketID, flatKeyMaker)
	ratioFeeReq := hasFeeRatio(store, marketID, ratioKeyMaker)

	if !flatFeeReq && !ratioFeeReq {
		// no fee required. All good.
		return nil
	}

	if len(fee) == 0 {
		// a fee is required, but we have none.
		return errors.New("insufficient buyer settlement fee: no fee provided")
	}

	// Loop through each coin in the fee looking for entries that satisfy the fee requirements.
	var flatFeeOk, ratioFeeOk bool
	var errs []error
	for _, feeCoin := range fee {
		var flatFeeAmt, ratioFeeAmt *sdkmath.Int

		if flatFeeReq {
			flatFee := getFlatFee(store, marketID, feeCoin.Denom, flatKeyMaker)
			switch {
			case flatFee == nil:
				errs = append(errs, fmt.Errorf("no flat fee options available for denom %s", feeCoin.Denom))
			case feeCoin.Amount.LT(flatFee.Amount):
				errs = append(errs, fmt.Errorf("%s is less than required flat fee %s", feeCoin, flatFee))
			case !ratioFeeReq:
				// This fee coin covers the flat fee, and there is no ratio fee needed, so we're all good.
				return nil
			case ratioFeeOk:
				// A previous fee coin covered the ratio fee and this one covers the flat fee, so we're all good.
				return nil
			default:
				flatFeeAmt = &flatFee.Amount
			}
		}

		if ratioFeeReq {
			ratio := getFeeRatio(store, marketID, price.Denom, feeCoin.Denom, ratioKeyMaker)
			if ratio == nil {
				errs = append(errs, fmt.Errorf("no ratio from price denom %s to fee denom %s",
					price.Denom, feeCoin.Denom))
			} else {
				ratioFee, err := ratio.ApplyTo(price)
				switch {
				case err != nil:
					errs = append(errs, err)
				case feeCoin.Amount.LT(ratioFee.Amount):
					errs = append(errs, fmt.Errorf("%s is less than required ratio fee %s (based on price %s and ratio %s)",
						feeCoin, ratioFee, price, ratio))
				case !flatFeeReq:
					// This fee coin covers the ratio fee, and there's no flat fee needed, so we're all good.
					return nil
				case flatFeeOk:
					// A previous fee coin covered the flat fee and this one covers the ratio fee, so we're all good.
					return nil
				default:
					ratioFeeAmt = &ratioFee.Amount
				}
			}
		}

		// If we have both a satisfactory flat and ratio fee, check this fee coin against the sum.
		if flatFeeAmt != nil && ratioFeeAmt != nil {
			reqAmt := flatFeeAmt.Add(*ratioFeeAmt)
			if feeCoin.Amount.LT(reqAmt) {
				errs = append(errs, fmt.Errorf("%s is less than combined fee %s%s = flat %s%s + ratio %s%s (based on price %s)",
					feeCoin, reqAmt, fee[0].Denom, flatFeeAmt, fee[0].Denom, ratioFeeAmt, fee[0].Denom, price))
			} else {
				// This one coin fee is so satisfying. (How satisfying was it?)
				// It's so satisfying, it covers both! Thank you for coming to my TED talk.
				return nil
			}
		}

		// Keep track of satisfied requirements for the next fee coin.
		if flatFeeAmt != nil {
			flatFeeOk = true
		}
		if ratioFeeAmt != nil {
			ratioFeeOk = true
		}
	}

	// Programmer Sanity check.
	// Either we returned earlier or we added at least one entry to errs.
	if len(errs) == 0 {
		panic("no specific errors identified")
	}

	// If a fee type was required, but not satisfied, add that to the errors for easier identification by users.
	if flatFeeReq && !flatFeeOk {
		flatFeeOptions := getAllFlatFees(store, marketID, flatKeyMaker)
		errs = append(errs, fmt.Errorf("required flat fee not satisfied, valid options: %s", flatFeeOptions))
	}
	if ratioFeeReq && !ratioFeeOk {
		allRatioOptions := getAllFeeRatios(store, marketID, ratioKeyMaker)
		errs = append(errs, fmt.Errorf("required ratio fee not satisfied, valid ratios: %s",
			mapToJoinedStrings(allRatioOptions, exchange.FeeRatio.String)))
	}

	// And add an error with the overall reason for this failure.
	errs = append(errs, fmt.Errorf("insufficient buyer settlement fee %s", fee))

	return errors.Join(errs...)
}

// mapToJoinedStrings runs the provided mapper on each of the vals and returns the resulting strings joined using ", ".
func mapToJoinedStrings[T any](vals []T, mapper func(entry T) string) string {
	return strings.Join(mapToStrings(vals, mapper), ", ")
}

// mapToStrings runs the provided mapper on each of the vals and returns the resulting slice of strings.
func mapToStrings[T any](vals []T, mapper func(entry T) string) []string {
	rv := make([]string, len(vals))
	for i, val := range vals {
		rv[i] = mapper(val)
	}
	return rv
}
