package keeper

import (
	"errors"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/x/exchange"
)

// TODO[1658]: Recheck all the public funcs in here to make sure they're still needed.

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

// validateFlatFee returns an error if the provided fee is not sufficient to cover the required flat fee.
func validateFlatFee(store sdk.KVStore, marketID uint32, fee *sdk.Coin, name string, maker flatFeeKeyMakers) error {
	if !hasFlatFee(store, marketID, maker) {
		return nil
	}
	if fee == nil {
		opts := getAllFlatFees(store, marketID, maker)
		return fmt.Errorf("no %s fee provided, must be one of: %s", name, sdk.NewCoins(opts...).String())
	}
	reqFee := getFlatFee(store, marketID, fee.Denom, maker)
	if reqFee == nil {
		opts := getAllFlatFees(store, marketID, maker)
		return fmt.Errorf("invalid %s fee, must be one of: %s", name, sdk.NewCoins(opts...).String())
	}
	if fee.Amount.LT(reqFee.Amount) {
		return fmt.Errorf("insufficient %s fee: %q is less than required amount %q", name, fee, reqFee)
	}
	return nil
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

// getCreateAskFlatFees gets the create-ask flat fee options for a market.
func getCreateAskFlatFees(store sdk.KVStore, marketID uint32) []sdk.Coin {
	return getAllFlatFees(store, marketID, createAskFlatKeyMakers)
}

// GetCreateAskFlatFees gets the create-ask flat fee options for a market.
func (k Keeper) GetCreateAskFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return getCreateAskFlatFees(k.getStore(ctx), marketID)
}

// setCreateAskFlatFees sets the create-ask flat fees for a market.
func setCreateAskFlatFees(store sdk.KVStore, marketID uint32, options []sdk.Coin) {
	setAllFlatFees(store, marketID, options, createAskFlatKeyMakers)
}

// updateCreateAskFlatFees deletes all create-ask flat fees to delete then adds the ones to add.
func updateCreateAskFlatFees(store sdk.KVStore, marketID uint32, toDelete, toAdd []sdk.Coin) {
	updateFlatFees(store, marketID, toDelete, toAdd, createAskFlatKeyMakers)
}

// validateCreateAskFlatFee returns an error if the provided fee is not a sufficient create ask flat fee.
func validateCreateAskFlatFee(store sdk.KVStore, marketID uint32, fee *sdk.Coin) error {
	return validateFlatFee(store, marketID, fee, "ask order creation", createAskFlatKeyMakers)
}

// getCreateBidFlatFees gets the create-bid flat fee options for a market.
func getCreateBidFlatFees(store sdk.KVStore, marketID uint32) []sdk.Coin {
	return getAllFlatFees(store, marketID, createBidFlatKeyMakers)
}

// GetCreateBidFlatFees gets the create-bid flat fee options for a market.
func (k Keeper) GetCreateBidFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return getCreateBidFlatFees(k.getStore(ctx), marketID)
}

// setCreateBidFlatFees sets the create-bid flat fees for a market.
func setCreateBidFlatFees(store sdk.KVStore, marketID uint32, options []sdk.Coin) {
	setAllFlatFees(store, marketID, options, createBidFlatKeyMakers)
}

// updateCreateBidFlatFees deletes all create-bid flat fees to delete then adds the ones to add.
func updateCreateBidFlatFees(store sdk.KVStore, marketID uint32, toDelete, toAdd []sdk.Coin) {
	updateFlatFees(store, marketID, toDelete, toAdd, createBidFlatKeyMakers)
}

// validateCreateBidFlatFee returns an error if the provided fee is not a sufficient create bid flat fee.
func validateCreateBidFlatFee(store sdk.KVStore, marketID uint32, fee *sdk.Coin) error {
	return validateFlatFee(store, marketID, fee, "bid order creation", createBidFlatKeyMakers)
}

// getSellerSettlementFlatFees gets the seller settlement flat fee options for a market.
func getSellerSettlementFlatFees(store sdk.KVStore, marketID uint32) []sdk.Coin {
	return getAllFlatFees(store, marketID, sellerSettlementFlatKeyMakers)
}

// GetSellerSettlementFlatFees gets the seller settlement flat fee options for a market.
func (k Keeper) GetSellerSettlementFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return getSellerSettlementFlatFees(k.getStore(ctx), marketID)
}

// setSellerSettlementFlatFees sets the seller settlement flat fees for a market.
func setSellerSettlementFlatFees(store sdk.KVStore, marketID uint32, options []sdk.Coin) {
	setAllFlatFees(store, marketID, options, sellerSettlementFlatKeyMakers)
}

// updateSellerSettlementFlatFees deletes all seller settlement flat fees to delete then adds the ones to add.
func updateSellerSettlementFlatFees(store sdk.KVStore, marketID uint32, toDelete, toAdd []sdk.Coin) {
	updateFlatFees(store, marketID, toDelete, toAdd, sellerSettlementFlatKeyMakers)
}

// validateSellerSettlementFlatFee returns an error if the provided fee is not a sufficient seller settlement flat fee.
func validateSellerSettlementFlatFee(store sdk.KVStore, marketID uint32, fee *sdk.Coin) error {
	return validateFlatFee(store, marketID, fee, "seller settlement flat", createBidFlatKeyMakers)
}

// getSellerSettlementRatios gets the seller settlement fee ratios for a market.
func getSellerSettlementRatios(store sdk.KVStore, marketID uint32) []exchange.FeeRatio {
	return getAllFeeRatios(store, marketID, sellerSettlementRatioKeyMakers)
}

// GetSellerSettlementRatios gets the seller settlement fee ratios for a market.
func (k Keeper) GetSellerSettlementRatios(ctx sdk.Context, marketID uint32) []exchange.FeeRatio {
	return getSellerSettlementRatios(k.getStore(ctx), marketID)
}

// setSellerSettlementRatios sets the seller settlement fee ratios for a market.
func setSellerSettlementRatios(store sdk.KVStore, marketID uint32, ratios []exchange.FeeRatio) {
	setAllFeeRatios(store, marketID, ratios, sellerSettlementRatioKeyMakers)
}

// updateSellerSettlementRatios deletes all seller settlement ratio entries to delete then adds the ones to add.
func updateSellerSettlementRatios(store sdk.KVStore, marketID uint32, toDelete, toAdd []exchange.FeeRatio) {
	updateFeeRatios(store, marketID, toDelete, toAdd, sellerSettlementRatioKeyMakers)
}

// getSellerSettlementRatio gets the seller settlement fee ratio for the given market with the provided denom.
func getSellerSettlementRatio(store sdk.KVStore, marketID uint32, priceDenom string) (*exchange.FeeRatio, error) {
	ratio := getFeeRatio(store, marketID, priceDenom, priceDenom, sellerSettlementRatioKeyMakers)
	if ratio == nil {
		if hasFeeRatio(store, marketID, sellerSettlementRatioKeyMakers) {
			return nil, fmt.Errorf("no seller settlement fee ratio found for denom %q", priceDenom)
		}
	}
	return ratio, nil
}

// validateAskPrice validates that the provided ask price denom is acceptable.
func validateAskPrice(store sdk.KVStore, marketID uint32, price sdk.Coin, settlementFlatFee *sdk.Coin) error {
	ratio, err := getSellerSettlementRatio(store, marketID, price.Denom)
	if err != nil {
		return err
	}

	// If there is a settlement flat fee with a different denom as the price, a hold is placed on it.
	// If there's a settlement flat fee with the same denom as the price, it's paid out of the price along
	// with the ratio amount. Assuming the ratio is less than one, the price will always cover the ratio fee amount.
	// But if the flat fee is coming out of the price too, it's possible that the price might be less than the total
	// fee that will need to come out of it. We want to return an error if that's the case.
	if settlementFlatFee != nil && price.Denom == settlementFlatFee.Denom {
		if price.Amount.LT(settlementFlatFee.Amount) {
			return fmt.Errorf("price %s is less than seller settlement flat fee %s", price, settlementFlatFee)
		}
		if ratio != nil {
			ratioFee, err := ratio.ApplyToLoosely(price)
			if err != nil {
				return err
			}
			reqPrice := settlementFlatFee.Add(ratioFee)
			if price.IsLT(reqPrice) {
				return fmt.Errorf("price %s is less than total required seller settlement fee of %s = %s flat + %s ratio",
					price, reqPrice, settlementFlatFee, ratioFee)
			}
		}
	}

	return err
}

// calculateSellerSettlementRatioFee calculates the seller settlement fee required for the given price.
func calculateSellerSettlementRatioFee(store sdk.KVStore, marketID uint32, price sdk.Coin) (*sdk.Coin, error) {
	ratio, err := getSellerSettlementRatio(store, marketID, price.Denom)
	if err != nil {
		return nil, err
	}
	if ratio == nil {
		return nil, nil
	}
	rv, err := ratio.ApplyTo(price)
	if err != nil {
		return nil, fmt.Errorf("invalid seller settlement fees: %w", err)
	}
	return &rv, nil
}

// getBuyerSettlementFlatFees gets the buyer settlement flat fee options for a market.
func getBuyerSettlementFlatFees(store sdk.KVStore, marketID uint32) []sdk.Coin {
	return getAllFlatFees(store, marketID, buyerSettlementFlatKeyMakers)
}

// GetBuyerSettlementFlatFees gets the buyer settlement flat fee options for a market.
func (k Keeper) GetBuyerSettlementFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return getBuyerSettlementFlatFees(k.getStore(ctx), marketID)
}

// setBuyerSettlementFlatFees sets the buyer settlement flat fees for a market.
func setBuyerSettlementFlatFees(store sdk.KVStore, marketID uint32, options []sdk.Coin) {
	setAllFlatFees(store, marketID, options, buyerSettlementFlatKeyMakers)
}

// updateBuyerSettlementFlatFees deletes all buyer settlement flat fees to delete then adds the ones to add.
func updateBuyerSettlementFlatFees(store sdk.KVStore, marketID uint32, toDelete, toAdd []sdk.Coin) {
	updateFlatFees(store, marketID, toDelete, toAdd, buyerSettlementFlatKeyMakers)
}

// getBuyerSettlementRatios gets the buyer settlement fee ratios for a market.
func getBuyerSettlementRatios(store sdk.KVStore, marketID uint32) []exchange.FeeRatio {
	return getAllFeeRatios(store, marketID, buyerSettlementRatioKeyMakers)
}

// GetBuyerSettlementRatios gets the buyer settlement fee ratios for a market.
func (k Keeper) GetBuyerSettlementRatios(ctx sdk.Context, marketID uint32) []exchange.FeeRatio {
	return getBuyerSettlementRatios(k.getStore(ctx), marketID)
}

// setBuyerSettlementRatios sets the buyer settlement fee ratios for a market.
func setBuyerSettlementRatios(store sdk.KVStore, marketID uint32, ratios []exchange.FeeRatio) {
	setAllFeeRatios(store, marketID, ratios, buyerSettlementRatioKeyMakers)
}

// updateBuyerSettlementRatios deletes all buyer settlement ratio entries to delete then adds the ones to add.
func updateBuyerSettlementRatios(store sdk.KVStore, marketID uint32, toDelete, toAdd []exchange.FeeRatio) {
	updateFeeRatios(store, marketID, toDelete, toAdd, buyerSettlementRatioKeyMakers)
}

// getBuyerSettlementFeeRatiosForPriceDenom gets all the buyer settlement fee ratios in a market that have the give price denom.
func getBuyerSettlementFeeRatiosForPriceDenom(store sdk.KVStore, marketID uint32, priceDenom string) []exchange.FeeRatio {
	var ratios []exchange.FeeRatio
	iterate(store, GetKeyPrefixMarketBuyerSettlementRatioForPriceDenom(marketID, priceDenom), func(key, value []byte) bool {
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
	ratios := getBuyerSettlementFeeRatiosForPriceDenom(k.getStore(ctx), marketID, price.Denom)
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

// validateBuyerSettlementFee returns an error if the provided fee is not enough to cover both the
// buyer settlement flat and percent fees for the given price.
func validateBuyerSettlementFee(store sdk.KVStore, marketID uint32, price sdk.Coin, fee sdk.Coins) error {
	flatKeyMaker := buyerSettlementFlatKeyMakers
	ratioKeyMaker := buyerSettlementRatioKeyMakers
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
			exchange.FeeRatiosString(allRatioOptions)))
	}

	// And add an error with the overall reason for this failure.
	errs = append(errs, fmt.Errorf("insufficient buyer settlement fee %s", fee))

	return errors.Join(errs...)
}

// isMarketActive returns true if the provided market is accepting orders.
func isMarketActive(store sdk.KVStore, marketID uint32) bool {
	key := MakeKeyMarketInactive(marketID)
	return !store.Has(key)
}

// IsMarketActive returns true if the provided market is accepting orders.
func (k Keeper) IsMarketActive(ctx sdk.Context, marketID uint32) bool {
	return isMarketActive(k.getStore(ctx), marketID)
}

// setMarketActive sets whether the provided market is accepting orders.
func setMarketActive(store sdk.KVStore, marketID uint32, active bool) {
	key := MakeKeyMarketInactive(marketID)
	if active {
		store.Delete(key)
	} else {
		store.Set(key, nil)
	}
}

// SetMarketActive sets whether the provided market is accepting orders.
func (k Keeper) SetMarketActive(ctx sdk.Context, marketID uint32, active bool) {
	setMarketActive(k.getStore(ctx), marketID, active)
}

// isUserSettlementAllowed gets whether user-settlement is allowed for a market.
func isUserSettlementAllowed(store sdk.KVStore, marketID uint32) bool {
	key := MakeKeyMarketUserSettle(marketID)
	return store.Has(key)
}

// IsUserSettlementAllowed gets whether user-settlement is allowed for a market.
func (k Keeper) IsUserSettlementAllowed(ctx sdk.Context, marketID uint32) bool {
	return isUserSettlementAllowed(k.getStore(ctx), marketID)
}

// SetUserSettlementAllowed sets whether user-settlement is allowed for a market.
func setUserSettlementAllowed(store sdk.KVStore, marketID uint32, allowed bool) {
	key := MakeKeyMarketUserSettle(marketID)
	if allowed {
		store.Set(key, nil)
	} else {
		store.Delete(key)
	}
}

// SetUserSettlementAllowed sets whether user-settlement is allowed for a market.
func (k Keeper) SetUserSettlementAllowed(ctx sdk.Context, marketID uint32, allowed bool) {
	setUserSettlementAllowed(k.getStore(ctx), marketID, allowed)
}

// hasPermission returns true if the provided address has the given permission in the market in question.
func hasPermission(store sdk.KVStore, marketID uint32, addr sdk.AccAddress, permission exchange.Permission) bool {
	key := MakeKeyMarketPermissions(marketID, addr, permission)
	return store.Has(key)
}

// grantPermissions updates the store so that the given address has the provided permissions in a market.
func grantPermissions(store sdk.KVStore, marketID uint32, addr sdk.AccAddress, permissions []exchange.Permission) {
	for _, perm := range permissions {
		key := MakeKeyMarketPermissions(marketID, addr, perm)
		store.Set(key, nil)
	}
}

// revokePermissions updates the store so that the given address does NOT have the provided permissions for the market.
func revokePermissions(store sdk.KVStore, marketID uint32, addr sdk.AccAddress, permissions []exchange.Permission) {
	for _, perm := range permissions {
		key := MakeKeyMarketPermissions(marketID, addr, perm)
		store.Delete(key)
	}
}

// revokeAllUserPermissions updates the store so that the given address does not have any permissions for the market.
func revokeAllUserPermissions(store sdk.KVStore, marketID uint32, addr sdk.AccAddress) {
	key := GetKeyPrefixMarketPermissionsForAddress(marketID, addr)
	deleteAll(store, key)
}

// revokeAllMarketPermissions clears out all permissions for a market.
func revokeAllMarketPermissions(store sdk.KVStore, marketID uint32) {
	key := GetKeyPrefixMarketPermissions(marketID)
	deleteAll(store, key)
}

// setAllMarketPermissions clears out all market permissions then stores just the ones provided.
func setAllMarketPermissions(store sdk.KVStore, marketID uint32, grants []exchange.AccessGrant) {
	revokeAllMarketPermissions(store, marketID)
	for _, ag := range grants {
		grantPermissions(store, marketID, sdk.MustAccAddressFromBech32(ag.Address), ag.Permissions)
	}
}

// updatePermissions revokes all permissions from the provided revokeAll bech32 addresses, then revokes all permissions
// in the toRevoke list, and lastly, grants all the permissions in toGrant.
func updatePermissions(store sdk.KVStore, marketID uint32, revokeAll []string, toRevoke, toGrant []exchange.AccessGrant) {
	for _, revAddr := range revokeAll {
		revokeAllUserPermissions(store, marketID, sdk.MustAccAddressFromBech32(revAddr))
	}
	for _, ag := range toRevoke {
		revokePermissions(store, marketID, sdk.MustAccAddressFromBech32(ag.Address), ag.Permissions)
	}
	for _, ag := range toGrant {
		grantPermissions(store, marketID, sdk.MustAccAddressFromBech32(ag.Address), ag.Permissions)
	}
}

// HasPermission returns true if the provided addr has the permission in question for a given market.
func (k Keeper) HasPermission(ctx sdk.Context, marketID uint32, addr sdk.AccAddress, permission exchange.Permission) bool {
	return hasPermission(k.getStore(ctx), marketID, addr, permission)
}

// getUserPermissions gets all permissions that have been granted to a user in a market.
func getUserPermissions(store sdk.KVStore, marketID uint32, addr sdk.AccAddress) []exchange.Permission {
	var rv []exchange.Permission
	iterate(store, GetKeyPrefixMarketPermissionsForAddress(marketID, addr), func(key, _ []byte) bool {
		rv = append(rv, exchange.Permission(key[0]))
		return false
	})
	return rv
}

// GetUserPermissions gets all permissions that have been granted to a user in a market.
func (k Keeper) GetUserPermissions(ctx sdk.Context, marketID uint32, addr sdk.AccAddress) []exchange.Permission {
	return getUserPermissions(k.getStore(ctx), marketID, addr)
}

// getAccessGrants gets all the access grants for a market.
func getAccessGrants(store sdk.KVStore, marketID uint32) []exchange.AccessGrant {
	var rv []exchange.AccessGrant
	var lastAG exchange.AccessGrant
	iterate(store, GetKeyPrefixMarketPermissions(marketID), func(key, _ []byte) bool {
		addr, perm, err := ParseKeySuffixMarketPermissions(key)
		if err == nil {
			if addr.String() != lastAG.Address {
				lastAG = exchange.AccessGrant{Address: addr.String()}
				rv = append(rv, lastAG)
			}
			lastAG.Permissions = append(lastAG.Permissions, perm)
		}
		return false
	})
	return rv
}

// setAccessGrants deletes all access grants on a market and sets just the ones provided.
func setAccessGrants(store sdk.KVStore, marketID uint32, grants []exchange.AccessGrant) {
	revokeAllMarketPermissions(store, marketID)
	setAllMarketPermissions(store, marketID, grants)
}

// UpdateAccessGrants revokes all permissions from the provided revokeAll bech32 addresses, then revokes all permissions
// in the toRevoke list, and lastly, grants all the permissions in toGrant.
func (k Keeper) UpdateAccessGrants(ctx sdk.Context, marketID uint32, revokeAll []string, toRevoke, toGrant []exchange.AccessGrant) {
	updatePermissions(k.getStore(ctx), marketID, revokeAll, toRevoke, toGrant)
}

// reqAttrKeyMaker is a function that returns a key for required attributes.
type reqAttrKeyMaker func(marketID uint32) []byte

// getReqAttr gets the required attributes for a market using the provided key maker.
func getReqAttr(store sdk.KVStore, marketID uint32, maker reqAttrKeyMaker) []string {
	key := maker(marketID)
	value := store.Get(key)
	return ParseReqAttrStoreValue(value)
}

// setReqAttr sets the required attributes for a market using the provided key maker.
func setReqAttr(store sdk.KVStore, marketID uint32, reqAttrs []string, maker reqAttrKeyMaker) {
	key := maker(marketID)
	if len(reqAttrs) == 0 {
		store.Delete(key)
	} else {
		value := []byte(strings.Join(reqAttrs, string(RecordSeparator)))
		store.Set(key, value)
	}
}

// getReqAttrAsk gets the attributes required to create an ask order.
func getReqAttrAsk(store sdk.KVStore, marketID uint32) []string {
	return getReqAttr(store, marketID, MakeKeyMarketReqAttrAsk)
}

// GetReqAttrAsk gets the attributes required to create an ask order.
func (k Keeper) GetReqAttrAsk(ctx sdk.Context, marketID uint32) []string {
	return getReqAttrAsk(k.getStore(ctx), marketID)
}

// setReqAttrAsk sets the attributes required to create an ask order.
func setReqAttrAsk(store sdk.KVStore, marketID uint32, reqAttrs []string) {
	setReqAttr(store, marketID, reqAttrs, MakeKeyMarketReqAttrAsk)
}

// getReqAttrBid gets the attributes required to create a bid order.
func getReqAttrBid(store sdk.KVStore, marketID uint32) []string {
	return getReqAttr(store, marketID, MakeKeyMarketReqAttrBid)
}

// GetReqAttrBid gets the attributes required to create a bid order.
func (k Keeper) GetReqAttrBid(ctx sdk.Context, marketID uint32) []string {
	return getReqAttrBid(k.getStore(ctx), marketID)
}

// setReqAttrBid sets the attributes required to create a bid order.
func setReqAttrBid(store sdk.KVStore, marketID uint32, reqAttrs []string) {
	setReqAttr(store, marketID, reqAttrs, MakeKeyMarketReqAttrBid)
}

// getLastAutoMarketID gets the last auto-selected market id.
func getLastAutoMarketID(store sdk.KVStore) uint32 {
	key := MakeKeyLastMarketID()
	value := store.Get(key)
	return uint32FromBz(value)
}

// setLastAutoMarketID sets the last auto-selected market id to the provided value.
func setLastAutoMarketID(store sdk.KVStore, marketID uint32) {
	key := MakeKeyLastMarketID()
	value := uint32Bz(marketID)
	store.Set(key, value)
}

// NextMarketID finds the next available market id, updates the last auto-selected
// market id store entry, and returns the unused id it found.
func (k Keeper) NextMarketID(ctx sdk.Context) uint32 {
	store := k.getStore(ctx)
	marketID := getLastAutoMarketID(store) + 1
	for {
		marketAddr := exchange.GetMarketAddress(marketID)
		if !k.accountKeeper.HasAccount(ctx, marketAddr) {
			break
		}
	}
	setLastAutoMarketID(store, marketID)
	return marketID
}

// setMarketKnown sets the known market id indicator in the store.
func setMarketKnown(store sdk.KVStore, marketID uint32) {
	key := MakeKeyKnownMarketID(marketID)
	store.Set(key, nil)
}

// validateMarketExists returns an error if the provided marketID does not exist.
func validateMarketExists(store sdk.KVStore, marketID uint32) error {
	key := MakeKeyKnownMarketID(marketID)
	if !store.Has(key) {
		return fmt.Errorf("market %d does not exist", marketID)
	}
	return nil
}

// GetAllMarketIDs gets all the known market ids from the store.
func (k Keeper) GetAllMarketIDs(ctx sdk.Context) []uint32 {
	var rv []uint32
	k.iterate(ctx, GetKeyPrefixKnownMarketID(), func(key, _ []byte) bool {
		rv = append(rv, ParseKeySuffixKnownMarketID(key))
		return false
	})
	return rv
}

// CreateMarket saves a new market to the store with all the info provided.
// If the marketId is zero, the next available one will be used.
func (k Keeper) CreateMarket(ctx sdk.Context, market exchange.Market) (marketID uint32, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = fmt.Errorf("could not set market: %w", e)
			} else {
				err = fmt.Errorf("could not set market: %v", r)
			}
		}
	}()

	if market.MarketId == 0 {
		market.MarketId = k.NextMarketID(ctx)
	}
	marketID = market.MarketId

	marketAddr := exchange.GetMarketAddress(marketID)
	if k.accountKeeper.HasAccount(ctx, marketAddr) {
		return 0, fmt.Errorf("market id %d %s already exists", marketID, marketAddr)
	}

	reqAttrCreateAsk, errAsk := k.NormalizeReqAttrs(ctx, market.ReqAttrCreateAsk)
	reqAttrCreateBid, errBid := k.NormalizeReqAttrs(ctx, market.ReqAttrCreateAsk)
	err = errors.Join(errAsk, errBid)
	if err != nil {
		return 0, err
	}

	marketAcc := &exchange.MarketAccount{
		BaseAccount:   &authtypes.BaseAccount{Address: marketAddr.String()},
		MarketId:      marketID,
		MarketDetails: market.MarketDetails,
	}
	k.accountKeeper.NewAccount(ctx, marketAcc)
	k.accountKeeper.SetAccount(ctx, marketAcc)

	store := k.getStore(ctx)
	setMarketKnown(store, marketID)
	setCreateAskFlatFees(store, marketID, market.FeeCreateAskFlat)
	setCreateBidFlatFees(store, marketID, market.FeeCreateBidFlat)
	setSellerSettlementFlatFees(store, marketID, market.FeeSellerSettlementFlat)
	setSellerSettlementRatios(store, marketID, market.FeeSellerSettlementRatios)
	setBuyerSettlementFlatFees(store, marketID, market.FeeBuyerSettlementFlat)
	setBuyerSettlementRatios(store, marketID, market.FeeBuyerSettlementRatios)
	setMarketActive(store, marketID, market.AcceptingOrders)
	setUserSettlementAllowed(store, marketID, market.AllowUserSettlement)
	setAccessGrants(store, marketID, market.AccessGrants)
	setReqAttrAsk(store, marketID, reqAttrCreateAsk)
	setReqAttrBid(store, marketID, reqAttrCreateBid)

	return marketID, err
}

// GetMarket reads all the market info from state and returns it.
// Returns nil if the market account doesn't exist or it's not a market account.
func (k Keeper) GetMarket(ctx sdk.Context, marketID uint32) *exchange.Market {
	marketAddr := exchange.GetMarketAddress(marketID)
	acc := k.accountKeeper.GetAccount(ctx, marketAddr)
	if acc == nil {
		return nil
	}
	marketAcc, ok := acc.(*exchange.MarketAccount)
	if !ok {
		return nil
	}

	store := k.getStore(ctx)
	market := &exchange.Market{MarketId: marketID}
	market.MarketDetails = marketAcc.MarketDetails
	market.FeeCreateAskFlat = getCreateAskFlatFees(store, marketID)
	market.FeeCreateBidFlat = getCreateBidFlatFees(store, marketID)
	market.FeeSellerSettlementFlat = getSellerSettlementFlatFees(store, marketID)
	market.FeeSellerSettlementRatios = getSellerSettlementRatios(store, marketID)
	market.FeeBuyerSettlementFlat = getBuyerSettlementFlatFees(store, marketID)
	market.FeeBuyerSettlementRatios = getBuyerSettlementRatios(store, marketID)
	market.AcceptingOrders = isMarketActive(store, marketID)
	market.AllowUserSettlement = isUserSettlementAllowed(store, marketID)
	market.AccessGrants = getAccessGrants(store, marketID)
	market.ReqAttrCreateAsk = getReqAttrAsk(store, marketID)
	market.ReqAttrCreateBid = getReqAttrBid(store, marketID)

	return market
}

// UpdateFees updates all the fees as provided in the MsgGovManageFeesRequest.
func (k Keeper) UpdateFees(ctx sdk.Context, msg *exchange.MsgGovManageFeesRequest) {
	store := k.getStore(ctx)
	updateCreateAskFlatFees(store, msg.MarketId, msg.RemoveFeeCreateAskFlat, msg.AddFeeCreateAskFlat)
	updateCreateBidFlatFees(store, msg.MarketId, msg.RemoveFeeCreateBidFlat, msg.AddFeeCreateBidFlat)
	updateSellerSettlementFlatFees(store, msg.MarketId, msg.RemoveFeeSellerSettlementFlat, msg.AddFeeSellerSettlementFlat)
	updateSellerSettlementRatios(store, msg.MarketId, msg.RemoveFeeSellerSettlementRatios, msg.AddFeeSellerSettlementRatios)
	updateBuyerSettlementFlatFees(store, msg.MarketId, msg.RemoveFeeBuyerSettlementFlat, msg.AddFeeBuyerSettlementFlat)
	updateBuyerSettlementRatios(store, msg.MarketId, msg.RemoveFeeBuyerSettlementRatios, msg.AddFeeBuyerSettlementRatios)
}

// hasReqAttrs returns true if either reqAttrs is empty or the provide address has all of them on their account.
func (k Keeper) hasReqAttrs(ctx sdk.Context, addr sdk.AccAddress, reqAttrs []string) bool {
	if len(reqAttrs) == 0 {
		return true
	}
	attrs, err := k.attrKeeper.GetAllAttributesAddr(ctx, addr)
	if err != nil {
		return false
	}
	accAttrs := make([]string, len(attrs))
	for i, attr := range attrs {
		accAttrs[i] = attr.Name
	}
	missing := exchange.FindUnmatchedReqAttrs(reqAttrs, accAttrs)
	return len(missing) == 0
}

// CanCreateAsk returns true if the provided address is allowed to create an ask order in the given market.
func (k Keeper) CanCreateAsk(ctx sdk.Context, marketID uint32, addr sdk.AccAddress) bool {
	reqAttrs := k.GetReqAttrAsk(ctx, marketID)
	return k.hasReqAttrs(ctx, addr, reqAttrs)
}

// CanCreateBid returns true if the provided address is allowed to create a bid order in the given market.
func (k Keeper) CanCreateBid(ctx sdk.Context, marketID uint32, addr sdk.AccAddress) bool {
	reqAttrs := k.GetReqAttrBid(ctx, marketID)
	return k.hasReqAttrs(ctx, addr, reqAttrs)
}

// CanWithdrawMarketFunds returns true if the provided admin bech32 address has permission to
// withdraw funds from the given market's account.
func (k Keeper) CanWithdrawMarketFunds(ctx sdk.Context, marketID uint32, admin string) bool {
	if admin == k.GetAuthority() {
		return true
	}
	adminAddr := sdk.MustAccAddressFromBech32(admin)
	return hasPermission(k.getStore(ctx), marketID, adminAddr, exchange.Permission_withdraw)
}

// WithdrawMarketFunds transfers funds from a market account to another account.
func (k Keeper) WithdrawMarketFunds(ctx sdk.Context, marketID uint32, toAddr sdk.AccAddress, amount sdk.Coins) error {
	marketAddr := exchange.GetMarketAddress(marketID)
	err := k.bankKeeper.SendCoins(ctx, marketAddr, toAddr, amount)
	if err != nil {
		return fmt.Errorf("failed to withdraw %s from market %d: %w", amount, marketID, err)
	}
	return nil
}
