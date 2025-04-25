package keeper

import (
	"errors"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/x/exchange"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	"github.com/provenance-io/provenance/x/quarantine"
)

// getLastAutoMarketID gets the last auto-selected market id.
func getLastAutoMarketID(store storetypes.KVStore) uint32 {
	key := MakeKeyLastMarketID()
	value := store.Get(key)
	rv, _ := uint32FromBz(value)
	return rv
}

// setLastAutoMarketID sets the last auto-selected market id to the provided value.
func setLastAutoMarketID(store storetypes.KVStore, marketID uint32) {
	key := MakeKeyLastMarketID()
	value := uint32Bz(marketID)
	store.Set(key, value)
}

// nextMarketID finds the next available market id, updates the last auto-selected
// market id store entry, and returns the unused id it found.
func nextMarketID(store storetypes.KVStore) uint32 {
	marketID := getLastAutoMarketID(store) + 1
	for {
		key := MakeKeyKnownMarketID(marketID)
		if !store.Has(key) {
			break
		}
		marketID++
	}
	setLastAutoMarketID(store, marketID)
	return marketID
}

// isMarketKnown returns true if the provided market id is a market that exists.
func isMarketKnown(store storetypes.KVStore, marketID uint32) bool {
	key := MakeKeyKnownMarketID(marketID)
	return store.Has(key)
}

// setMarketKnown sets the known market id indicator in the store.
func setMarketKnown(store storetypes.KVStore, marketID uint32) {
	key := MakeKeyKnownMarketID(marketID)
	store.Set(key, []byte{})
}

// validateMarketExists returns an error if the provided marketID does not exist.
func validateMarketExists(store storetypes.KVStore, marketID uint32) error {
	if !isMarketKnown(store, marketID) {
		return fmt.Errorf("market %d does not exist", marketID)
	}
	return nil
}

// IterateKnownMarketIDs iterates over all known market ids.
func (k Keeper) IterateKnownMarketIDs(ctx sdk.Context, cb func(marketID uint32) bool) {
	k.iterate(ctx, GetKeyPrefixKnownMarketID(), func(key, _ []byte) bool {
		marketID, ok := ParseKeySuffixKnownMarketID(key)
		return ok && cb(marketID)
	})
}

// flatFeeKeyMakers are the key and prefix maker functions for a specific flat fee entry.
type flatFeeKeyMakers struct {
	key    func(marketID uint32, denom string) []byte
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
	// createCommitmentFlatKeyMakers are the key and prefix makers for the create-commitment flat fees.
	createCommitmentFlatKeyMakers = flatFeeKeyMakers{
		key:    MakeKeyMarketCreateCommitmentFlatFee,
		prefix: GetKeyPrefixMarketCreateCommitmentFlatFee,
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

// hasFlatFee returns true if this market has any flat fee for a given type.
func hasFlatFee(store storetypes.KVStore, marketID uint32, maker flatFeeKeyMakers) bool {
	rv := false
	iterate(store, maker.prefix(marketID), func(_, _ []byte) bool {
		rv = true
		return true
	})
	return rv
}

// getFlatFee is a generic getter for a flat fee coin entry.
func getFlatFee(store storetypes.KVStore, marketID uint32, denom string, maker flatFeeKeyMakers) *sdk.Coin {
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
func setFlatFee(store storetypes.KVStore, marketID uint32, coin sdk.Coin, maker flatFeeKeyMakers) {
	key := maker.key(marketID, coin.Denom)
	value := coin.Amount.String()
	store.Set(key, []byte(value))
}

// validateFlatFee returns an error if the provided fee is not sufficient to cover the required flat fee.
func validateFlatFee(store storetypes.KVStore, marketID uint32, fee *sdk.Coin, name string, maker flatFeeKeyMakers) error {
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
		return fmt.Errorf("invalid %s fee %q, must be one of: %s", name, fee, sdk.NewCoins(opts...).String())
	}
	if fee.Amount.LT(reqFee.Amount) {
		return fmt.Errorf("insufficient %s fee: %q is less than required amount %q", name, fee, reqFee)
	}
	return nil
}

// getAllFlatFees gets all the coin entries from the store with the given prefix.
// The denom comes from the part of the key after the prefix, and the amount comes from the values.
func getAllFlatFees(store storetypes.KVStore, marketID uint32, maker flatFeeKeyMakers) []sdk.Coin {
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
func setAllFlatFees(store storetypes.KVStore, marketID uint32, options []sdk.Coin, maker flatFeeKeyMakers) {
	deleteAll(store, maker.prefix(marketID))
	for _, coin := range options {
		setFlatFee(store, marketID, coin, maker)
	}
}

// updateFlatFees deletes all the entries with a denom in toDelete, then writes all the toWrite entries.
func updateFlatFees(store storetypes.KVStore, marketID uint32, toDelete, toWrite []sdk.Coin, maker flatFeeKeyMakers) {
	for _, coin := range toDelete {
		key := maker.key(marketID, coin.Denom)
		store.Delete(key)
	}
	for _, coin := range toWrite {
		setFlatFee(store, marketID, coin, maker)
	}
}

// validateCreateAskFlatFee returns an error if the provided fee is not a sufficient create-ask flat fee.
func validateCreateAskFlatFee(store storetypes.KVStore, marketID uint32, fee *sdk.Coin) error {
	return validateFlatFee(store, marketID, fee, "ask order creation", createAskFlatKeyMakers)
}

// getCreateAskFlatFees gets the create-ask flat fee options for a market.
func getCreateAskFlatFees(store storetypes.KVStore, marketID uint32) []sdk.Coin {
	return getAllFlatFees(store, marketID, createAskFlatKeyMakers)
}

// setCreateAskFlatFees sets the create-ask flat fees for a market.
func setCreateAskFlatFees(store storetypes.KVStore, marketID uint32, options []sdk.Coin) {
	setAllFlatFees(store, marketID, options, createAskFlatKeyMakers)
}

// updateCreateAskFlatFees deletes all create-ask flat fees to delete then adds the ones to add.
func updateCreateAskFlatFees(store storetypes.KVStore, marketID uint32, toDelete, toAdd []sdk.Coin) {
	updateFlatFees(store, marketID, toDelete, toAdd, createAskFlatKeyMakers)
}

// validateCreateBidFlatFee returns an error if the provided fee is not a sufficient create-bid flat fee.
func validateCreateBidFlatFee(store storetypes.KVStore, marketID uint32, fee *sdk.Coin) error {
	return validateFlatFee(store, marketID, fee, "bid order creation", createBidFlatKeyMakers)
}

// getCreateBidFlatFees gets the create-bid flat fee options for a market.
func getCreateBidFlatFees(store storetypes.KVStore, marketID uint32) []sdk.Coin {
	return getAllFlatFees(store, marketID, createBidFlatKeyMakers)
}

// setCreateBidFlatFees sets the create-bid flat fees for a market.
func setCreateBidFlatFees(store storetypes.KVStore, marketID uint32, options []sdk.Coin) {
	setAllFlatFees(store, marketID, options, createBidFlatKeyMakers)
}

// updateCreateBidFlatFees deletes all create-bid flat fees to delete then adds the ones to add.
func updateCreateBidFlatFees(store storetypes.KVStore, marketID uint32, toDelete, toAdd []sdk.Coin) {
	updateFlatFees(store, marketID, toDelete, toAdd, createBidFlatKeyMakers)
}

// validateCreateCommitmentFlatFee returns an error if the provided fee is not a sufficient create-commitment flat fee.
func validateCreateCommitmentFlatFee(store storetypes.KVStore, marketID uint32, fee *sdk.Coin) error {
	return validateFlatFee(store, marketID, fee, "commitment creation", createCommitmentFlatKeyMakers)
}

// getCreateCommitmentFlatFees gets the create-commitment flat fee options for a market.
func getCreateCommitmentFlatFees(store storetypes.KVStore, marketID uint32) []sdk.Coin {
	return getAllFlatFees(store, marketID, createCommitmentFlatKeyMakers)
}

// setCreateCommitmentFlatFees sets the create-commitment flat fees for a market.
func setCreateCommitmentFlatFees(store storetypes.KVStore, marketID uint32, options []sdk.Coin) {
	setAllFlatFees(store, marketID, options, createCommitmentFlatKeyMakers)
}

// updateCreateCommitmentFlatFees deletes all create-commitment flat fees to delete then adds the ones to add.
func updateCreateCommitmentFlatFees(store storetypes.KVStore, marketID uint32, toDelete, toAdd []sdk.Coin) {
	updateFlatFees(store, marketID, toDelete, toAdd, createCommitmentFlatKeyMakers)
}

// validateSellerSettlementFlatFee returns an error if the provided fee is not a sufficient seller settlement flat fee.
func validateSellerSettlementFlatFee(store storetypes.KVStore, marketID uint32, fee *sdk.Coin) error {
	return validateFlatFee(store, marketID, fee, "seller settlement flat", sellerSettlementFlatKeyMakers)
}

// getSellerSettlementFlatFees gets the seller settlement flat fee options for a market.
func getSellerSettlementFlatFees(store storetypes.KVStore, marketID uint32) []sdk.Coin {
	return getAllFlatFees(store, marketID, sellerSettlementFlatKeyMakers)
}

// setSellerSettlementFlatFees sets the seller settlement flat fees for a market.
func setSellerSettlementFlatFees(store storetypes.KVStore, marketID uint32, options []sdk.Coin) {
	setAllFlatFees(store, marketID, options, sellerSettlementFlatKeyMakers)
}

// updateSellerSettlementFlatFees deletes all seller settlement flat fees to delete then adds the ones to add.
func updateSellerSettlementFlatFees(store storetypes.KVStore, marketID uint32, toDelete, toAdd []sdk.Coin) {
	updateFlatFees(store, marketID, toDelete, toAdd, sellerSettlementFlatKeyMakers)
}

// getBuyerSettlementFlatFees gets the buyer settlement flat fee options for a market.
func getBuyerSettlementFlatFees(store storetypes.KVStore, marketID uint32) []sdk.Coin {
	return getAllFlatFees(store, marketID, buyerSettlementFlatKeyMakers)
}

// setBuyerSettlementFlatFees sets the buyer settlement flat fees for a market.
func setBuyerSettlementFlatFees(store storetypes.KVStore, marketID uint32, options []sdk.Coin) {
	setAllFlatFees(store, marketID, options, buyerSettlementFlatKeyMakers)
}

// updateBuyerSettlementFlatFees deletes all buyer settlement flat fees to delete then adds the ones to add.
func updateBuyerSettlementFlatFees(store storetypes.KVStore, marketID uint32, toDelete, toAdd []sdk.Coin) {
	updateFlatFees(store, marketID, toDelete, toAdd, buyerSettlementFlatKeyMakers)
}

// ratioKeyMakers are the key and prefix maker functions for a specific ratio fee entry.
type ratioKeyMakers struct {
	key    func(marketID uint32, ratio exchange.FeeRatio) []byte
	prefix func(marketID uint32) []byte
}

var (
	// sellerSettlementRatioKeyMakers are the key and prefix makers for the seller settlement fee ratios.
	sellerSettlementRatioKeyMakers = ratioKeyMakers{
		key:    MakeKeyMarketSellerSettlementRatio,
		prefix: GetKeyPrefixMarketSellerSettlementRatio,
	}
	// buyerSettlementRatioKeyMakers are the key and prefix makers for the buyer settlement fee ratios.
	buyerSettlementRatioKeyMakers = ratioKeyMakers{
		key:    MakeKeyMarketBuyerSettlementRatio,
		prefix: GetKeyPrefixMarketBuyerSettlementRatio,
	}
)

// hasFeeRatio returns true if this market has any fee ratios for a given type.
func hasFeeRatio(store storetypes.KVStore, marketID uint32, maker ratioKeyMakers) bool {
	rv := false
	iterate(store, maker.prefix(marketID), func(_, _ []byte) bool {
		rv = true
		return true
	})
	return rv
}

// getFeeRatio is a generic getter for a fee ratio entry in the store.
func getFeeRatio(store storetypes.KVStore, marketID uint32, priceDenom, feeDenom string, maker ratioKeyMakers) *exchange.FeeRatio {
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
func setFeeRatio(store storetypes.KVStore, marketID uint32, ratio exchange.FeeRatio, maker ratioKeyMakers) {
	key := maker.key(marketID, ratio)
	value := GetFeeRatioStoreValue(ratio)
	store.Set(key, value)
}

// getAllFeeRatios gets all the fee ratio entries from the store with the given prefix.
// The denoms come from the keys and amounts come from the values.
func getAllFeeRatios(store storetypes.KVStore, marketID uint32, maker ratioKeyMakers) []exchange.FeeRatio {
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
func setAllFeeRatios(store storetypes.KVStore, marketID uint32, ratios []exchange.FeeRatio, maker ratioKeyMakers) {
	deleteAll(store, maker.prefix(marketID))
	for _, ratio := range ratios {
		setFeeRatio(store, marketID, ratio, maker)
	}
}

// updateFeeRatios deletes all entries with the denom pairs in toDelete, then writes all the toWrite entries.
func updateFeeRatios(store storetypes.KVStore, marketID uint32, toDelete, toWrite []exchange.FeeRatio, maker ratioKeyMakers) {
	for _, ratio := range toDelete {
		key := maker.key(marketID, ratio)
		store.Delete(key)
	}
	for _, ratio := range toWrite {
		setFeeRatio(store, marketID, ratio, maker)
	}
}

// getSellerSettlementRatio gets the seller settlement fee ratio for the given market with the provided denom.
func getSellerSettlementRatio(store storetypes.KVStore, marketID uint32, priceDenom string) (*exchange.FeeRatio, error) {
	ratio := getFeeRatio(store, marketID, priceDenom, priceDenom, sellerSettlementRatioKeyMakers)
	if ratio == nil {
		if hasFeeRatio(store, marketID, sellerSettlementRatioKeyMakers) {
			return nil, fmt.Errorf("no seller settlement fee ratio found for denom %q", priceDenom)
		}
	}
	return ratio, nil
}

// getSellerSettlementRatios gets the seller settlement fee ratios for a market.
func getSellerSettlementRatios(store storetypes.KVStore, marketID uint32) []exchange.FeeRatio {
	return getAllFeeRatios(store, marketID, sellerSettlementRatioKeyMakers)
}

// setSellerSettlementRatios sets the seller settlement fee ratios for a market.
func setSellerSettlementRatios(store storetypes.KVStore, marketID uint32, ratios []exchange.FeeRatio) {
	setAllFeeRatios(store, marketID, ratios, sellerSettlementRatioKeyMakers)
}

// updateSellerSettlementRatios deletes all seller settlement ratio entries to delete then adds the ones to add.
func updateSellerSettlementRatios(store storetypes.KVStore, marketID uint32, toDelete, toAdd []exchange.FeeRatio) {
	updateFeeRatios(store, marketID, toDelete, toAdd, sellerSettlementRatioKeyMakers)
}

// validateAskPrice validates that the provided ask price is acceptable.
func validateAskPrice(store storetypes.KVStore, marketID uint32, price sdk.Coin, settlementFlatFee *sdk.Coin) error {
	ratio, err := getSellerSettlementRatio(store, marketID, price.Denom)
	if err != nil {
		return err
	}

	// If there is a settlement flat fee with a different denom as the price, a hold is placed on it.
	// If there's a settlement flat fee with the same denom as the price, it's paid out of the price along
	// with the ratio amount. Assuming the ratio is less than one, the price will always be at least the ratio fee amount.
	// Here, we make sure that the price is more than any fee that is to come out of it.
	// Since the ask price is a minimum, and the ratio is less than 1 (fee amount goes up slower than price amount),
	// if it's okay with the provided price, it'll also be okay for a larger price too.

	checkFlat := settlementFlatFee != nil && !settlementFlatFee.Amount.IsZero() && price.Denom == settlementFlatFee.Denom
	if ratio == nil {
		// There's no ratio aspect to check, just maybe look at the flat.
		if checkFlat && price.Amount.LTE(settlementFlatFee.Amount) {
			return fmt.Errorf("price %s is not more than seller settlement flat fee %s", price, settlementFlatFee)
		}
		return nil
	}

	ratioFee, rerr := ratio.ApplyToLoosely(price)
	if rerr != nil {
		return rerr
	}

	if !checkFlat {
		// There's no flat aspect to check, just check the ratio.
		if price.Amount.LTE(ratioFee.Amount) {
			return fmt.Errorf("price %s is not more than seller settlement ratio fee %s", price, ratioFee)
		}
		return nil
	}

	// Check both together.
	reqPriceAmt := settlementFlatFee.Amount.Add(ratioFee.Amount)
	if price.Amount.LTE(reqPriceAmt) {
		return fmt.Errorf("price %s is not more than total required seller settlement fee %s = %s flat + %s ratio",
			price, sdk.NewCoin(price.Denom, reqPriceAmt), settlementFlatFee, ratioFee)
	}

	return nil
}

// calculateSellerSettlementRatioFee calculates the seller settlement fee required for the given price.
func calculateSellerSettlementRatioFee(store storetypes.KVStore, marketID uint32, price sdk.Coin) (*sdk.Coin, error) {
	ratio, err := getSellerSettlementRatio(store, marketID, price.Denom)
	if err != nil {
		return nil, err
	}
	if ratio == nil {
		return nil, nil
	}
	rv, err := ratio.ApplyToLoosely(price)
	if err != nil {
		return nil, fmt.Errorf("invalid seller settlement fees: %w", err)
	}
	return &rv, nil
}

// getBuyerSettlementRatios gets the buyer settlement fee ratios for a market.
func getBuyerSettlementRatios(store storetypes.KVStore, marketID uint32) []exchange.FeeRatio {
	return getAllFeeRatios(store, marketID, buyerSettlementRatioKeyMakers)
}

// setBuyerSettlementRatios sets the buyer settlement fee ratios for a market.
func setBuyerSettlementRatios(store storetypes.KVStore, marketID uint32, ratios []exchange.FeeRatio) {
	setAllFeeRatios(store, marketID, ratios, buyerSettlementRatioKeyMakers)
}

// updateBuyerSettlementRatios deletes all buyer settlement ratio entries to delete then adds the ones to add.
func updateBuyerSettlementRatios(store storetypes.KVStore, marketID uint32, toDelete, toAdd []exchange.FeeRatio) {
	updateFeeRatios(store, marketID, toDelete, toAdd, buyerSettlementRatioKeyMakers)
}

// getBuyerSettlementFeeRatiosForPriceDenom gets all the buyer settlement fee ratios in a market that have the give price denom.
func getBuyerSettlementFeeRatiosForPriceDenom(store storetypes.KVStore, marketID uint32, priceDenom string) ([]exchange.FeeRatio, error) {
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
	if len(ratios) == 0 && hasFeeRatio(store, marketID, buyerSettlementRatioKeyMakers) {
		return nil, fmt.Errorf("no buyer settlement fee ratios found for denom %q", priceDenom)
	}
	return ratios, nil
}

// calcBuyerSettlementRatioFeeOptions calculates the buyer settlement ratio fee options available for the given price.
func calcBuyerSettlementRatioFeeOptions(store storetypes.KVStore, marketID uint32, price sdk.Coin) ([]sdk.Coin, error) {
	ratios, err := getBuyerSettlementFeeRatiosForPriceDenom(store, marketID, price.Denom)
	if err != nil {
		return nil, err
	}
	if len(ratios) == 0 {
		return nil, nil
	}

	var errs []error
	rv := make([]sdk.Coin, 0, len(ratios))
	for _, ratio := range ratios {
		fee, ferr := ratio.ApplyToLoosely(price)
		if ferr != nil {
			errs = append(errs, fmt.Errorf("buyer settlement fees: %w", ferr))
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
func validateBuyerSettlementFee(store storetypes.KVStore, marketID uint32, price sdk.Coin, fee sdk.Coins) error {
	flatKeyMaker := buyerSettlementFlatKeyMakers
	ratioKeyMaker := buyerSettlementRatioKeyMakers
	flatFeeReq := hasFlatFee(store, marketID, flatKeyMaker)
	ratioFeeReq := hasFeeRatio(store, marketID, ratioKeyMaker)

	if !flatFeeReq && !ratioFeeReq {
		// no fee required. All good.
		return nil
	}

	// Loop through each coin in the fee looking for entries that satisfy the fee requirements.
	var flatFeeOk, ratioFeeOk bool
	var flatErrs []error
	var ratioErrs []error
	var combErrs []error
	for _, feeCoin := range fee {
		var flatFeeAmt, ratioFeeAmt *sdkmath.Int

		if flatFeeReq {
			flatFee := getFlatFee(store, marketID, feeCoin.Denom, flatKeyMaker)
			switch {
			case flatFee == nil:
				flatErrs = append(flatErrs, fmt.Errorf("no flat fee options available for denom %s", feeCoin.Denom))
			case feeCoin.Amount.LT(flatFee.Amount):
				flatErrs = append(flatErrs, fmt.Errorf("%s is less than required flat fee %s", feeCoin, flatFee))
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
				ratioErrs = append(ratioErrs, fmt.Errorf("no ratio from price denom %s to fee denom %s",
					price.Denom, feeCoin.Denom))
			} else {
				ratioFee, err := ratio.ApplyToLoosely(price)
				switch {
				case err != nil:
					ratioErrs = append(ratioErrs, err)
				case feeCoin.Amount.LT(ratioFee.Amount):
					ratioErrs = append(ratioErrs, fmt.Errorf("%s is less than required ratio fee %s (based on price %s and ratio %s)",
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
				combErrs = append(combErrs, fmt.Errorf("%s is less than combined fee %s%s = %s%s (flat) + %s%s (ratio based on price %s)",
					feeCoin, reqAmt, feeCoin.Denom, flatFeeAmt, feeCoin.Denom, ratioFeeAmt, feeCoin.Denom, price))
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

	// If we get here, the fee is insufficient.
	// Combine all the known errors and add some to help users fix problems.
	var errs []error
	if flatFeeReq && !flatFeeOk {
		errs = append(errs, flatErrs...)
		flatFeeOptions := getAllFlatFees(store, marketID, flatKeyMaker)
		errs = append(errs, fmt.Errorf("required flat fee not satisfied, valid options: %s", sdk.Coins(flatFeeOptions)))
	}
	if ratioFeeReq && !ratioFeeOk {
		errs = append(errs, ratioErrs...)
		allRatioOptions := getAllFeeRatios(store, marketID, ratioKeyMaker)
		errs = append(errs, fmt.Errorf("required ratio fee not satisfied, valid ratios: %s",
			exchange.FeeRatiosString(allRatioOptions)))
	}
	if len(combErrs) > 0 {
		// Both flatFeeOk and ratioFeeOk will be true here, but we'll include
		// all those errors since only one of those is actually okay.
		errs = append(errs, flatErrs...)
		errs = append(errs, ratioErrs...)
		errs = append(errs, combErrs...)
	}

	// And add an error with the overall reason for this failure.
	if len(fee) > 0 {
		errs = append(errs, fmt.Errorf("insufficient buyer settlement fee %s", fee))
	} else {
		errs = append(errs, errors.New("insufficient buyer settlement fee: no fee provided"))
	}

	return errors.Join(errs...)
}

// getCommitmentSettlementBips gets the commitment settlement bips for the given market.
func getCommitmentSettlementBips(store storetypes.KVStore, marketID uint32) uint32 {
	key := MakeKeyMarketCommitmentSettlementBips(marketID)
	value := store.Get(key)
	if len(value) == 0 {
		return 0
	}
	rv, _ := uint32FromBz(value)
	return rv
}

// setCommitmentSettlementBips sets the commitment settlement bips for a market.
func setCommitmentSettlementBips(store storetypes.KVStore, marketID uint32, bips uint32) {
	key := MakeKeyMarketCommitmentSettlementBips(marketID)
	if bips != 0 {
		value := uint32Bz(bips)
		store.Set(key, value)
	} else {
		store.Delete(key)
	}
}

// updateCommitmentSettlementBips updates the commitment settlement bips for a market.
// If unsetBips is true, the bips entry for the market will be deleted.
// If bips is not zero, the entry will be set to that value.
// If bips is zero and unsetBips is false, this does nothing.
func updateCommitmentSettlementBips(store storetypes.KVStore, marketID uint32, bips uint32, unsetBips bool) {
	if unsetBips {
		setCommitmentSettlementBips(store, marketID, 0)
	}
	if bips > 0 {
		setCommitmentSettlementBips(store, marketID, bips)
	}
}

// getIntermediaryDenom gets a market's intermediary denom.
func getIntermediaryDenom(store storetypes.KVStore, marketID uint32) string {
	key := MakeKeyMarketIntermediaryDenom(marketID)
	rv := store.Get(key)
	return string(rv)
}

// setIntermediaryDenom sets the market's intermediary denom to the provided value.
func setIntermediaryDenom(store storetypes.KVStore, marketID uint32, denom string) {
	key := MakeKeyMarketIntermediaryDenom(marketID)
	if len(denom) > 0 {
		store.Set(key, []byte(denom))
	} else {
		store.Delete(key)
	}
}

// GetCreateAskFlatFees gets the create-ask flat fee options for a market.
func (k Keeper) GetCreateAskFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return getCreateAskFlatFees(k.getStore(ctx), marketID)
}

// GetCreateBidFlatFees gets the create-bid flat fee options for a market.
func (k Keeper) GetCreateBidFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return getCreateBidFlatFees(k.getStore(ctx), marketID)
}

// GetCreateCommitmentFlatFees gets the create-commitment flat fee options for a market.
func (k Keeper) GetCreateCommitmentFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return getCreateCommitmentFlatFees(k.getStore(ctx), marketID)
}

// GetSellerSettlementFlatFees gets the seller settlement flat fee options for a market.
func (k Keeper) GetSellerSettlementFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return getSellerSettlementFlatFees(k.getStore(ctx), marketID)
}

// GetSellerSettlementRatios gets the seller settlement fee ratios for a market.
func (k Keeper) GetSellerSettlementRatios(ctx sdk.Context, marketID uint32) []exchange.FeeRatio {
	return getSellerSettlementRatios(k.getStore(ctx), marketID)
}

// GetBuyerSettlementFlatFees gets the buyer settlement flat fee options for a market.
func (k Keeper) GetBuyerSettlementFlatFees(ctx sdk.Context, marketID uint32) []sdk.Coin {
	return getBuyerSettlementFlatFees(k.getStore(ctx), marketID)
}

// GetBuyerSettlementRatios gets the buyer settlement fee ratios for a market.
func (k Keeper) GetBuyerSettlementRatios(ctx sdk.Context, marketID uint32) []exchange.FeeRatio {
	return getBuyerSettlementRatios(k.getStore(ctx), marketID)
}

// GetCommitmentSettlementBips gets the commitment settlement bips for the given market.
func (k Keeper) GetCommitmentSettlementBips(ctx sdk.Context, marketID uint32) uint32 {
	return getCommitmentSettlementBips(k.getStore(ctx), marketID)
}

// GetIntermediaryDenom gets a market's intermediary denom.
func (k Keeper) GetIntermediaryDenom(ctx sdk.Context, marketID uint32) string {
	return getIntermediaryDenom(k.getStore(ctx), marketID)
}

// CalculateSellerSettlementRatioFee calculates the seller settlement fee required for the given price.
func (k Keeper) CalculateSellerSettlementRatioFee(ctx sdk.Context, marketID uint32, price sdk.Coin) (*sdk.Coin, error) {
	return calculateSellerSettlementRatioFee(k.getStore(ctx), marketID, price)
}

// CalculateBuyerSettlementRatioFeeOptions calculates the buyer settlement ratio fee options available for the given price.
func (k Keeper) CalculateBuyerSettlementRatioFeeOptions(ctx sdk.Context, marketID uint32, price sdk.Coin) ([]sdk.Coin, error) {
	return calcBuyerSettlementRatioFeeOptions(k.getStore(ctx), marketID, price)
}

// ValidateCreateAskFlatFee returns an error if the provided fee is not a sufficient create-ask flat fee.
func (k Keeper) ValidateCreateAskFlatFee(ctx sdk.Context, marketID uint32, fee *sdk.Coin) error {
	return validateCreateAskFlatFee(k.getStore(ctx), marketID, fee)
}

// ValidateCreateBidFlatFee returns an error if the provided fee is not a sufficient create-bid flat fee.
func (k Keeper) ValidateCreateBidFlatFee(ctx sdk.Context, marketID uint32, fee *sdk.Coin) error {
	return validateCreateBidFlatFee(k.getStore(ctx), marketID, fee)
}

// ValidateCreateCommitmentFlatFee returns an error if the provided fee is not a sufficient create-commitment flat fee.
func (k Keeper) ValidateCreateCommitmentFlatFee(ctx sdk.Context, marketID uint32, fee *sdk.Coin) error {
	return validateCreateCommitmentFlatFee(k.getStore(ctx), marketID, fee)
}

// ValidateSellerSettlementFlatFee returns an error if the provided fee is not a sufficient seller settlement flat fee.
func (k Keeper) ValidateSellerSettlementFlatFee(ctx sdk.Context, marketID uint32, fee *sdk.Coin) error {
	return validateSellerSettlementFlatFee(k.getStore(ctx), marketID, fee)
}

// ValidateAskPrice validates that the provided ask price is acceptable.
func (k Keeper) ValidateAskPrice(ctx sdk.Context, marketID uint32, price sdk.Coin, settlementFlatFee *sdk.Coin) error {
	return validateAskPrice(k.getStore(ctx), marketID, price, settlementFlatFee)
}

// ValidateBuyerSettlementFee returns an error if the provided fee is not enough to cover both the
// buyer settlement flat and percent fees for the given price.
func (k Keeper) ValidateBuyerSettlementFee(ctx sdk.Context, marketID uint32, price sdk.Coin, fee sdk.Coins) error {
	return validateBuyerSettlementFee(k.getStore(ctx), marketID, price, fee)
}

// UpdateFees updates all the fees as provided in the MsgGovManageFeesRequest.
func (k Keeper) UpdateFees(ctx sdk.Context, msg *exchange.MsgGovManageFeesRequest) {
	store := k.getStore(ctx)
	updateCreateAskFlatFees(store, msg.MarketId, msg.RemoveFeeCreateAskFlat, msg.AddFeeCreateAskFlat)
	updateCreateBidFlatFees(store, msg.MarketId, msg.RemoveFeeCreateBidFlat, msg.AddFeeCreateBidFlat)
	updateCreateCommitmentFlatFees(store, msg.MarketId, msg.RemoveFeeCreateCommitmentFlat, msg.AddFeeCreateCommitmentFlat)
	updateSellerSettlementFlatFees(store, msg.MarketId, msg.RemoveFeeSellerSettlementFlat, msg.AddFeeSellerSettlementFlat)
	updateSellerSettlementRatios(store, msg.MarketId, msg.RemoveFeeSellerSettlementRatios, msg.AddFeeSellerSettlementRatios)
	updateBuyerSettlementFlatFees(store, msg.MarketId, msg.RemoveFeeBuyerSettlementFlat, msg.AddFeeBuyerSettlementFlat)
	updateBuyerSettlementRatios(store, msg.MarketId, msg.RemoveFeeBuyerSettlementRatios, msg.AddFeeBuyerSettlementRatios)
	updateCommitmentSettlementBips(store, msg.MarketId, msg.SetFeeCommitmentSettlementBips, msg.UnsetFeeCommitmentSettlementBips)

	k.emitEvent(ctx, exchange.NewEventMarketFeesUpdated(msg.MarketId))
}

// UpdateIntermediaryDenom sets the market's intermediary denom to the one provided.
func (k Keeper) UpdateIntermediaryDenom(ctx sdk.Context, marketID uint32, denom string, updatedBy string) {
	setIntermediaryDenom(k.getStore(ctx), marketID, denom)
	k.emitEvent(ctx, exchange.NewEventMarketIntermediaryDenomUpdated(marketID, updatedBy))
}

// validateMarketUpdateAcceptingCommitments checks that the market has things set up
// to change the accepting-commitments flag to the provided value.
func validateMarketUpdateAcceptingCommitments(store storetypes.KVStore, marketID uint32, newAllow bool) error {
	curAllow := isMarketAcceptingCommitments(store, marketID)
	if curAllow == newAllow {
		return fmt.Errorf("market %d already has accepting-commitments %t", marketID, curAllow)
	}

	if newAllow {
		bips := getCommitmentSettlementBips(store, marketID)
		cFee := getCreateCommitmentFlatFees(store, marketID)
		if bips == 0 && len(cFee) == 0 {
			return fmt.Errorf("market %d does not have any commitment fees defined", marketID)
		}
	}

	return nil
}

// isMarketAcceptingOrders returns true if the provided market's not-accepting-orders flag does not exist.
// See also isMarketKnown.
func isMarketAcceptingOrders(store storetypes.KVStore, marketID uint32) bool {
	key := MakeKeyMarketNotAcceptingOrders(marketID)
	return !store.Has(key)
}

// setMarketAcceptingOrders sets whether the provided market is accepting orders.
func setMarketAcceptingOrders(store storetypes.KVStore, marketID uint32, accepting bool) {
	key := MakeKeyMarketNotAcceptingOrders(marketID)
	if accepting {
		store.Delete(key)
	} else {
		store.Set(key, []byte{})
	}
}

// isUserSettlementAllowed gets whether user-settlement is allowed for a market.
func isUserSettlementAllowed(store storetypes.KVStore, marketID uint32) bool {
	key := MakeKeyMarketUserSettle(marketID)
	return store.Has(key)
}

// setUserSettlementAllowed sets whether user-settlement is allowed for a market.
func setUserSettlementAllowed(store storetypes.KVStore, marketID uint32, allowed bool) {
	key := MakeKeyMarketUserSettle(marketID)
	if allowed {
		store.Set(key, []byte{})
	} else {
		store.Delete(key)
	}
}

// isMarketAcceptingCommitments gets whether commitments are allowed for a market.
func isMarketAcceptingCommitments(store storetypes.KVStore, marketID uint32) bool {
	key := MakeKeyMarketAcceptingCommitments(marketID)
	return store.Has(key)
}

// setMarketAcceptingCommitments sets whether commitments are allowed for a market.
func setMarketAcceptingCommitments(store storetypes.KVStore, marketID uint32, allowed bool) {
	key := MakeKeyMarketAcceptingCommitments(marketID)
	if allowed {
		store.Set(key, []byte{})
	} else {
		store.Delete(key)
	}
}

// IsMarketKnown returns true if the provided market id is a known market's id.
func (k Keeper) IsMarketKnown(ctx sdk.Context, marketID uint32) bool {
	return isMarketKnown(k.getStore(ctx), marketID)
}

// IsMarketAcceptingOrders returns true if the provided market is accepting orders.
func (k Keeper) IsMarketAcceptingOrders(ctx sdk.Context, marketID uint32) bool {
	store := k.getStore(ctx)
	if !isMarketAcceptingOrders(store, marketID) {
		return false
	}
	return isMarketKnown(store, marketID)
}

// UpdateMarketAcceptingOrders updates the accepting orders flag for a market.
// An error is returned if the setting is already what is provided.
func (k Keeper) UpdateMarketAcceptingOrders(ctx sdk.Context, marketID uint32, accepting bool, updatedBy string) error {
	store := k.getStore(ctx)
	current := isMarketAcceptingOrders(store, marketID)
	if current == accepting {
		return fmt.Errorf("market %d already has accepting-orders %t", marketID, accepting)
	}
	setMarketAcceptingOrders(store, marketID, accepting)
	k.emitEvent(ctx, exchange.NewEventMarketAcceptingOrdersUpdated(marketID, updatedBy, accepting))
	return nil
}

// IsUserSettlementAllowed gets whether user-settlement is allowed for a market.
func (k Keeper) IsUserSettlementAllowed(ctx sdk.Context, marketID uint32) bool {
	return isUserSettlementAllowed(k.getStore(ctx), marketID)
}

// UpdateUserSettlementAllowed updates the allow-user-settlement flag for a market.
// An error is returned if the setting is already what is provided.
func (k Keeper) UpdateUserSettlementAllowed(ctx sdk.Context, marketID uint32, allow bool, updatedBy string) error {
	store := k.getStore(ctx)
	current := isUserSettlementAllowed(store, marketID)
	if current == allow {
		return fmt.Errorf("market %d already has allow-user-settlement %t", marketID, allow)
	}
	setUserSettlementAllowed(store, marketID, allow)
	k.emitEvent(ctx, exchange.NewEventMarketUserSettleUpdated(marketID, updatedBy, allow))
	return nil
}

// IsMarketAcceptingCommitments gets whether commitments are allowed for a market.
func (k Keeper) IsMarketAcceptingCommitments(ctx sdk.Context, marketID uint32) bool {
	return isMarketAcceptingCommitments(k.getStore(ctx), marketID)
}

// UpdateMarketAcceptingCommitments updates the accepting-commitments flag for a market.
// An error is returned if the setting is already what is provided.
func (k Keeper) UpdateMarketAcceptingCommitments(ctx sdk.Context, marketID uint32, accepting bool, updatedBy string) error {
	store := k.getStore(ctx)
	current := isMarketAcceptingCommitments(store, marketID)
	if current == accepting {
		return fmt.Errorf("market %d already has accepting-commitments %t", marketID, accepting)
	}
	setMarketAcceptingCommitments(store, marketID, accepting)
	k.emitEvent(ctx, exchange.NewEventMarketAcceptingCommitmentsUpdated(marketID, updatedBy, accepting))
	return nil
}

// storeHasPermission returns true if there is an entry in the store for the given market, address, and permissions.
func storeHasPermission(store storetypes.KVStore, marketID uint32, addr sdk.AccAddress, permission exchange.Permission) bool {
	key := MakeKeyMarketPermissions(marketID, addr, permission)
	return store.Has(key)
}

// grantPermissions updates the store so that the given address has the provided permissions in a market.
func grantPermissions(store storetypes.KVStore, marketID uint32, addr sdk.AccAddress, permissions []exchange.Permission) {
	for _, perm := range permissions {
		key := MakeKeyMarketPermissions(marketID, addr, perm)
		store.Set(key, []byte{})
	}
}

// revokePermissions updates the store so that the given address does NOT have the provided permissions for the market.
func revokePermissions(store storetypes.KVStore, marketID uint32, addr sdk.AccAddress, permissions []exchange.Permission) {
	for _, perm := range permissions {
		key := MakeKeyMarketPermissions(marketID, addr, perm)
		store.Delete(key)
	}
}

// revokeUserPermissions updates the store so that the given address does not have any permissions for the market.
func revokeUserPermissions(store storetypes.KVStore, marketID uint32, addr sdk.AccAddress) {
	key := GetKeyPrefixMarketPermissionsForAddress(marketID, addr)
	deleteAll(store, key)
}

// getUserPermissions gets all permissions that have been granted to a user in a market.
func getUserPermissions(store storetypes.KVStore, marketID uint32, addr sdk.AccAddress) []exchange.Permission {
	var rv []exchange.Permission
	iterate(store, GetKeyPrefixMarketPermissionsForAddress(marketID, addr), func(key, _ []byte) bool {
		rv = append(rv, exchange.Permission(key[0]))
		return false
	})
	return rv
}

// revokeAllMarketPermissions clears out all permissions for a market.
func revokeAllMarketPermissions(store storetypes.KVStore, marketID uint32) {
	key := GetKeyPrefixMarketPermissions(marketID)
	deleteAll(store, key)
}

// getAccessGrants gets all the access grants for a market.
func getAccessGrants(store storetypes.KVStore, marketID uint32) []exchange.AccessGrant {
	var rv []exchange.AccessGrant
	iterate(store, GetKeyPrefixMarketPermissions(marketID), func(key, _ []byte) bool {
		addr, perm, err := ParseKeySuffixMarketPermissions(key)
		if err == nil {
			last := len(rv) - 1
			if last < 0 || addr.String() != rv[last].Address {
				rv = append(rv, exchange.AccessGrant{Address: addr.String()})
				last++
			}
			rv[last].Permissions = append(rv[last].Permissions, perm)
		}
		return false
	})
	return rv
}

// setAccessGrants deletes all access grants on a market and sets just the ones provided.
func setAccessGrants(store storetypes.KVStore, marketID uint32, grants []exchange.AccessGrant) {
	revokeAllMarketPermissions(store, marketID)
	for _, ag := range grants {
		grantPermissions(store, marketID, sdk.MustAccAddressFromBech32(ag.Address), ag.Permissions)
	}
}

// HasPermission returns true if the provided address has the permission in question for a given market.
// Also returns true if the provided address is the authority address.
func (k Keeper) HasPermission(ctx sdk.Context, marketID uint32, address string, permission exchange.Permission) bool {
	if k.IsAuthority(address) {
		return true
	}
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return false
	}
	return storeHasPermission(k.getStore(ctx), marketID, addr, permission)
}

// CanSettleOrders returns true if the provided admin bech32 address has permission to
// settle orders for a market. Also returns true if the provided address is the authority address.
func (k Keeper) CanSettleOrders(ctx sdk.Context, marketID uint32, admin string) bool {
	return k.HasPermission(ctx, marketID, admin, exchange.Permission_settle)
}

// CanSettleCommitments returns true if the provided admin bech32 address has permission to
// settle commitments for a market. Also returns true if the provided address is the authority address.
func (k Keeper) CanSettleCommitments(ctx sdk.Context, marketID uint32, admin string) bool {
	return k.HasPermission(ctx, marketID, admin, exchange.Permission_settle)
}

// CanSetIDs returns true if the provided admin bech32 address has permission to
// set UUIDs on orders for a market. Also returns true if the provided address is the authority address.
func (k Keeper) CanSetIDs(ctx sdk.Context, marketID uint32, admin string) bool {
	return k.HasPermission(ctx, marketID, admin, exchange.Permission_set_ids)
}

// CanCancelOrdersForMarket returns true if the provided admin bech32 address has permission to
// cancel orders for a market. Also returns true if the provided address is the authority address.
func (k Keeper) CanCancelOrdersForMarket(ctx sdk.Context, marketID uint32, admin string) bool {
	return k.HasPermission(ctx, marketID, admin, exchange.Permission_cancel)
}

// CanReleaseCommitmentsForMarket returns true if the provided admin bech32 address has permission to
// release commitments for a market. Also returns true if the provided address is the authority address.
func (k Keeper) CanReleaseCommitmentsForMarket(ctx sdk.Context, marketID uint32, admin string) bool {
	return k.HasPermission(ctx, marketID, admin, exchange.Permission_cancel)
}

// CanTransferCommitmentsForMarket returns true if the provided admin bech32 address has permission to
// transfer commitments from one market to another. Also returns true if the provided address is the authority address.
func (k Keeper) CanTransferCommitmentsForMarket(ctx sdk.Context, marketID uint32, admin string) bool {
	return k.HasPermission(ctx, marketID, admin, exchange.Permission_cancel)
}

// CanWithdrawMarketFunds returns true if the provided admin bech32 address has permission to
// withdraw funds from the given market's account. Also returns true if the provided address is the authority address.
func (k Keeper) CanWithdrawMarketFunds(ctx sdk.Context, marketID uint32, admin string) bool {
	return k.HasPermission(ctx, marketID, admin, exchange.Permission_withdraw)
}

// CanUpdateMarket returns true if the provided admin bech32 address has permission to
// update market details and settings. Also returns true if the provided address is the authority address.
func (k Keeper) CanUpdateMarket(ctx sdk.Context, marketID uint32, admin string) bool {
	return k.HasPermission(ctx, marketID, admin, exchange.Permission_update)
}

// CanManagePermissions returns true if the provided admin bech32 address has permission to
// manage user permissions for a given market. Also returns true if the provided address is the authority address.
func (k Keeper) CanManagePermissions(ctx sdk.Context, marketID uint32, admin string) bool {
	return k.HasPermission(ctx, marketID, admin, exchange.Permission_permissions)
}

// CanManageReqAttrs returns true if the provided admin bech32 address has permission to
// manage required attributes for a given market. Also returns true if the provided address is the authority address.
func (k Keeper) CanManageReqAttrs(ctx sdk.Context, marketID uint32, admin string) bool {
	return k.HasPermission(ctx, marketID, admin, exchange.Permission_attributes)
}

// GetUserPermissions gets all permissions that have been granted to a user in a market.
func (k Keeper) GetUserPermissions(ctx sdk.Context, marketID uint32, addr sdk.AccAddress) []exchange.Permission {
	return getUserPermissions(k.getStore(ctx), marketID, addr)
}

// GetAccessGrants gets all the access grants for a market.
func (k Keeper) GetAccessGrants(ctx sdk.Context, marketID uint32) []exchange.AccessGrant {
	return getAccessGrants(k.getStore(ctx), marketID)
}

// UpdatePermissions updates users permissions in the store using the provided changes.
// The caller is responsible for making sure this update should be allowed (e.g. by calling CanManagePermissions first).
func (k Keeper) UpdatePermissions(ctx sdk.Context, msg *exchange.MsgMarketManagePermissionsRequest) error {
	marketID := msg.MarketId
	store := k.getStore(ctx)
	var errs []error

	for _, addrStr := range msg.RevokeAll {
		addr := sdk.MustAccAddressFromBech32(addrStr)
		perms := getUserPermissions(store, marketID, addr)
		if len(perms) == 0 {
			errs = append(errs, fmt.Errorf("account %s does not have any permissions for market %d", addrStr, marketID))
		}
		if len(errs) == 0 {
			revokeUserPermissions(store, marketID, addr)
		}
	}

	for _, ag := range msg.ToRevoke {
		addr := sdk.MustAccAddressFromBech32(ag.Address)
		for _, perm := range ag.Permissions {
			if !storeHasPermission(store, marketID, addr, perm) {
				errs = append(errs, fmt.Errorf("account %s does not have %s for market %d", ag.Address, perm.String(), marketID))
			}
		}
		if len(errs) == 0 {
			revokePermissions(store, marketID, addr, ag.Permissions)
		}
	}

	for _, ag := range msg.ToGrant {
		addr := sdk.MustAccAddressFromBech32(ag.Address)
		for _, perm := range ag.Permissions {
			if storeHasPermission(store, marketID, addr, perm) {
				errs = append(errs, fmt.Errorf("account %s already has %s for market %d", ag.Address, perm.String(), marketID))
			}
		}
		if len(errs) == 0 {
			grantPermissions(store, marketID, addr, ag.Permissions)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	k.emitEvent(ctx, exchange.NewEventMarketPermissionsUpdated(marketID, msg.Admin))
	return nil
}

// reqAttrKeyMaker is a function that returns a key for required attributes.
type reqAttrKeyMaker func(marketID uint32) []byte

// getReqAttrs gets the required attributes for a market using the provided key maker.
func getReqAttrs(store storetypes.KVStore, marketID uint32, maker reqAttrKeyMaker) []string {
	key := maker(marketID)
	value := store.Get(key)
	return ParseReqAttrStoreValue(value)
}

// setReqAttrs sets the required attributes for a market using the provided key maker.
func setReqAttrs(store storetypes.KVStore, marketID uint32, reqAttrs []string, maker reqAttrKeyMaker) {
	key := maker(marketID)
	if len(reqAttrs) == 0 {
		store.Delete(key)
	} else {
		value := []byte(strings.Join(reqAttrs, string(RecordSeparator)))
		store.Set(key, value)
	}
}

// updateReqAttrs updates the required attributes in the store that use the provided key maker by removing then adding
// the provided attributes to the existing entries.
func updateReqAttrs(store storetypes.KVStore, marketID uint32, toRemove, toAdd []string, field string, maker reqAttrKeyMaker) error {
	var errs []error
	curAttrs := getReqAttrs(store, marketID, maker)

	for _, attr := range toRemove {
		if !exchange.ContainsString(curAttrs, attr) {
			errs = append(errs, fmt.Errorf("cannot remove %s required attribute %q: attribute not currently required", field, attr))
		}
	}

	var updatedAttrs []string
	for _, attr := range curAttrs {
		if !exchange.ContainsString(toRemove, attr) {
			updatedAttrs = append(updatedAttrs, attr)
		}
	}

	for _, attr := range toAdd {
		if !exchange.ContainsString(curAttrs, attr) {
			updatedAttrs = append(updatedAttrs, attr)
		} else {
			errs = append(errs, fmt.Errorf("cannot add %s required attribute %q: attribute already required", field, attr))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	setReqAttrs(store, marketID, updatedAttrs, maker)
	return nil
}

// getReqAttrsAsk gets the attributes required to create an ask order.
func getReqAttrsAsk(store storetypes.KVStore, marketID uint32) []string {
	return getReqAttrs(store, marketID, MakeKeyMarketReqAttrAsk)
}

// setReqAttrsAsk sets the attributes required to create an ask order.
func setReqAttrsAsk(store storetypes.KVStore, marketID uint32, reqAttrs []string) {
	setReqAttrs(store, marketID, reqAttrs, MakeKeyMarketReqAttrAsk)
}

// updateReqAttrsAsk updates the attributes required to create an ask order in the store by removing and adding
// the provided entries to the existing entries.
// It is assumed that the attributes have been normalized prior to calling this.
func updateReqAttrsAsk(store storetypes.KVStore, marketID uint32, toRemove, toAdd []string) error {
	return updateReqAttrs(store, marketID, toRemove, toAdd, "create-ask", MakeKeyMarketReqAttrAsk)
}

// getReqAttrsBid gets the attributes required to create a bid order.
func getReqAttrsBid(store storetypes.KVStore, marketID uint32) []string {
	return getReqAttrs(store, marketID, MakeKeyMarketReqAttrBid)
}

// setReqAttrsBid sets the attributes required to create a bid order.
func setReqAttrsBid(store storetypes.KVStore, marketID uint32, reqAttrs []string) {
	setReqAttrs(store, marketID, reqAttrs, MakeKeyMarketReqAttrBid)
}

// updateReqAttrsBid updates the attributes required to create a bid order in the store by removing and adding
// the provided entries to the existing entries.
// It is assumed that the attributes have been normalized prior to calling this.
func updateReqAttrsBid(store storetypes.KVStore, marketID uint32, toRemove, toAdd []string) error {
	return updateReqAttrs(store, marketID, toRemove, toAdd, "create-bid", MakeKeyMarketReqAttrBid)
}

// getReqAttrsCommitment gets the attributes required to create a commitment.
func getReqAttrsCommitment(store storetypes.KVStore, marketID uint32) []string {
	return getReqAttrs(store, marketID, MakeKeyMarketReqAttrCommitment)
}

// setReqAttrsCommitment sets the attributes required to create a commitment.
func setReqAttrsCommitment(store storetypes.KVStore, marketID uint32, reqAttrs []string) {
	setReqAttrs(store, marketID, reqAttrs, MakeKeyMarketReqAttrCommitment)
}

// updateReqAttrsCommitment updates the attributes required to create a commitment in the store by removing and adding
// the provided entries to the existing entries.
// It is assumed that the attributes have been normalized prior to calling this.
func updateReqAttrsCommitment(store storetypes.KVStore, marketID uint32, toRemove, toAdd []string) error {
	return updateReqAttrs(store, marketID, toRemove, toAdd, "create-commitment", MakeKeyMarketReqAttrCommitment)
}

// acctHasReqAttrs returns true if either reqAttrs is empty or the provide address has all of them on their account.
func (k Keeper) acctHasReqAttrs(ctx sdk.Context, addr sdk.AccAddress, reqAttrs []string) bool {
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

// GetReqAttrsAsk gets the attributes required to create an ask order.
func (k Keeper) GetReqAttrsAsk(ctx sdk.Context, marketID uint32) []string {
	return getReqAttrsAsk(k.getStore(ctx), marketID)
}

// GetReqAttrsBid gets the attributes required to create a bid order.
func (k Keeper) GetReqAttrsBid(ctx sdk.Context, marketID uint32) []string {
	return getReqAttrsBid(k.getStore(ctx), marketID)
}

// GetReqAttrsCommitment gets the attributes required to create a commitment.
func (k Keeper) GetReqAttrsCommitment(ctx sdk.Context, marketID uint32) []string {
	return getReqAttrsCommitment(k.getStore(ctx), marketID)
}

// CanCreateAsk returns true if the provided address is allowed to create an ask order in the given market.
func (k Keeper) CanCreateAsk(ctx sdk.Context, marketID uint32, addr sdk.AccAddress) bool {
	reqAttrs := k.GetReqAttrsAsk(ctx, marketID)
	return k.acctHasReqAttrs(ctx, addr, reqAttrs)
}

// CanCreateBid returns true if the provided address is allowed to create a bid order in the given market.
func (k Keeper) CanCreateBid(ctx sdk.Context, marketID uint32, addr sdk.AccAddress) bool {
	reqAttrs := k.GetReqAttrsBid(ctx, marketID)
	return k.acctHasReqAttrs(ctx, addr, reqAttrs)
}

// CanCreateCommitment returns true if the provided address is allowed to create a commitment in the given market.
func (k Keeper) CanCreateCommitment(ctx sdk.Context, marketID uint32, addr sdk.AccAddress) bool {
	reqAttrs := k.GetReqAttrsCommitment(ctx, marketID)
	return k.acctHasReqAttrs(ctx, addr, reqAttrs)
}

// UpdateReqAttrs updates the required attributes in the store using the provided changes.
// The caller is responsible for making sure this update should be allowed (e.g. by calling CanManageReqAttrs first).
func (k Keeper) UpdateReqAttrs(ctx sdk.Context, msg *exchange.MsgMarketManageReqAttrsRequest) error {
	var errs []error
	// We don't care if the attributes to remove are valid so that we
	// can remove entries that are somehow now invalid.
	askToRemove, _ := exchange.NormalizeReqAttrs(msg.CreateAskToRemove)
	askToAdd, err := exchange.NormalizeReqAttrs(msg.CreateAskToAdd)
	if err != nil {
		errs = append(errs, err)
	}
	bidToRemove, _ := exchange.NormalizeReqAttrs(msg.CreateBidToRemove)
	bidToAdd, err := exchange.NormalizeReqAttrs(msg.CreateBidToAdd)
	if err != nil {
		errs = append(errs, err)
	}
	commitToRemove, _ := exchange.NormalizeReqAttrs(msg.CreateCommitmentToRemove)
	commitToAdd, err := exchange.NormalizeReqAttrs(msg.CreateCommitmentToAdd)
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	marketID := msg.MarketId
	store := k.getStore(ctx)

	if err = updateReqAttrsAsk(store, marketID, askToRemove, askToAdd); err != nil {
		errs = append(errs, err)
	}
	if err = updateReqAttrsBid(store, marketID, bidToRemove, bidToAdd); err != nil {
		errs = append(errs, err)
	}
	if err = updateReqAttrsCommitment(store, marketID, commitToRemove, commitToAdd); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	k.emitEvent(ctx, exchange.NewEventMarketReqAttrUpdated(marketID, msg.Admin))
	return nil
}

// getMarketAccountByAddr gets a market's account given its address.
// This is for when you've already called exchange.GetMarketAddress(marketID) and need it for other things too.
func (k Keeper) getMarketAccountByAddr(ctx sdk.Context, marketAddr sdk.AccAddress) *exchange.MarketAccount {
	acc := k.accountKeeper.GetAccount(ctx, marketAddr)
	if acc == nil {
		return nil
	}
	marketAcc, ok := acc.(*exchange.MarketAccount)
	if !ok {
		return nil
	}
	return marketAcc
}

// GetMarketAccount gets a market's account from the account module.
func (k Keeper) GetMarketAccount(ctx sdk.Context, marketID uint32) *exchange.MarketAccount {
	marketAddr := exchange.GetMarketAddress(marketID)
	return k.getMarketAccountByAddr(ctx, marketAddr)
}

// GetMarketDetails gets a market's details.
func (k Keeper) GetMarketDetails(ctx sdk.Context, marketID uint32) *exchange.MarketDetails {
	marketAcc := k.GetMarketAccount(ctx, marketID)
	if marketAcc == nil {
		return nil
	}
	return &marketAcc.MarketDetails
}

// UpdateMarketDetails updates a market's details. It returns an error if the market account
// isn't found or if there aren't any changes provided.
func (k Keeper) UpdateMarketDetails(ctx sdk.Context, marketID uint32, marketDetails exchange.MarketDetails, updatedBy string) error {
	if err := marketDetails.Validate(); err != nil {
		return err
	}

	marketAcc := k.GetMarketAccount(ctx, marketID)
	if marketAcc == nil {
		return fmt.Errorf("market %d account not found", marketID)
	}

	if marketAcc.MarketDetails.Equal(marketDetails) {
		return errors.New("no changes")
	}

	marketAcc.MarketDetails = marketDetails
	k.accountKeeper.SetAccount(ctx, marketAcc)
	k.emitEvent(ctx, exchange.NewEventMarketDetailsUpdated(marketID, updatedBy))

	return nil
}

// storeMarket writes all the market fields to the state store (except MarketDetails which are in the account).
func storeMarket(store storetypes.KVStore, market exchange.Market) {
	marketID := market.MarketId
	setMarketKnown(store, marketID)
	setCreateAskFlatFees(store, marketID, market.FeeCreateAskFlat)
	setCreateBidFlatFees(store, marketID, market.FeeCreateBidFlat)
	setCreateCommitmentFlatFees(store, marketID, market.FeeCreateCommitmentFlat)
	setSellerSettlementFlatFees(store, marketID, market.FeeSellerSettlementFlat)
	setSellerSettlementRatios(store, marketID, market.FeeSellerSettlementRatios)
	setBuyerSettlementFlatFees(store, marketID, market.FeeBuyerSettlementFlat)
	setBuyerSettlementRatios(store, marketID, market.FeeBuyerSettlementRatios)
	setMarketAcceptingOrders(store, marketID, market.AcceptingOrders)
	setUserSettlementAllowed(store, marketID, market.AllowUserSettlement)
	setAccessGrants(store, marketID, market.AccessGrants)
	setReqAttrsAsk(store, marketID, market.ReqAttrCreateAsk)
	setReqAttrsBid(store, marketID, market.ReqAttrCreateBid)
	setReqAttrsCommitment(store, marketID, market.ReqAttrCreateCommitment)
	setMarketAcceptingCommitments(store, marketID, market.AcceptingCommitments)
	setCommitmentSettlementBips(store, marketID, market.CommitmentSettlementBips)
	setIntermediaryDenom(store, marketID, market.IntermediaryDenom)
}

// initMarket is similar to CreateMarket but assumes the market has already been
// validated and also allows for the market account to already exist.
func (k Keeper) initMarket(ctx sdk.Context, store storetypes.KVStore, market exchange.Market) {
	if market.MarketId == 0 {
		market.MarketId = nextMarketID(store)
	}
	marketID := market.MarketId

	marketAddr := exchange.GetMarketAddress(marketID)
	marketAcc := k.getMarketAccountByAddr(ctx, marketAddr)
	if marketAcc != nil {
		if !market.MarketDetails.Equal(marketAcc.MarketDetails) {
			marketAcc.MarketDetails = market.MarketDetails
			k.accountKeeper.SetAccount(ctx, marketAcc)
		}
	} else {
		marketAcc = &exchange.MarketAccount{
			BaseAccount:   &authtypes.BaseAccount{Address: marketAddr.String()},
			MarketId:      marketID,
			MarketDetails: market.MarketDetails,
		}
		k.accountKeeper.NewAccount(ctx, marketAcc)
		k.accountKeeper.SetAccount(ctx, marketAcc)
	}

	storeMarket(store, market)
}

// CreateMarket saves a new market to the store with all the info provided.
// If the marketId is zero, the next available one will be used.
func (k Keeper) CreateMarket(ctx sdk.Context, market exchange.Market) (uint32, error) {
	// Note: The Market is passed in by value, so any alterations directly to it here will be lost upon return.
	var errAsk, errBid error
	market.ReqAttrCreateAsk, errAsk = exchange.NormalizeReqAttrs(market.ReqAttrCreateAsk)
	market.ReqAttrCreateBid, errBid = exchange.NormalizeReqAttrs(market.ReqAttrCreateBid)
	errDets := market.MarketDetails.Validate()
	if errAsk != nil || errBid != nil || errDets != nil {
		return 0, errors.Join(errAsk, errBid, errDets)
	}

	store := k.getStore(ctx)

	if market.MarketId == 0 {
		market.MarketId = nextMarketID(store)
	}

	marketAddr := exchange.GetMarketAddress(market.MarketId)
	if k.accountKeeper.HasAccount(ctx, marketAddr) {
		return 0, fmt.Errorf("market id %d account %s already exists", market.MarketId, marketAddr)
	}

	marketAcc := &exchange.MarketAccount{
		BaseAccount:   &authtypes.BaseAccount{Address: marketAddr.String()},
		MarketId:      market.MarketId,
		MarketDetails: market.MarketDetails,
	}
	k.accountKeeper.NewAccount(ctx, marketAcc)
	k.accountKeeper.SetAccount(ctx, marketAcc)

	storeMarket(store, market)
	k.emitEvent(ctx, exchange.NewEventMarketCreated(market.MarketId))

	return market.MarketId, nil
}

// GetMarket reads all the market info from state and returns it.
// Returns nil if the market account doesn't exist or it's not a market account.
func (k Keeper) GetMarket(ctx sdk.Context, marketID uint32) *exchange.Market {
	store := k.getStore(ctx)
	if err := validateMarketExists(store, marketID); err != nil {
		return nil
	}

	market := &exchange.Market{MarketId: marketID}
	market.FeeCreateAskFlat = getCreateAskFlatFees(store, marketID)
	market.FeeCreateBidFlat = getCreateBidFlatFees(store, marketID)
	market.FeeCreateCommitmentFlat = getCreateCommitmentFlatFees(store, marketID)
	market.FeeSellerSettlementFlat = getSellerSettlementFlatFees(store, marketID)
	market.FeeSellerSettlementRatios = getSellerSettlementRatios(store, marketID)
	market.FeeBuyerSettlementFlat = getBuyerSettlementFlatFees(store, marketID)
	market.FeeBuyerSettlementRatios = getBuyerSettlementRatios(store, marketID)
	market.AcceptingOrders = isMarketAcceptingOrders(store, marketID)
	market.AllowUserSettlement = isUserSettlementAllowed(store, marketID)
	market.AccessGrants = getAccessGrants(store, marketID)
	market.ReqAttrCreateAsk = getReqAttrsAsk(store, marketID)
	market.ReqAttrCreateBid = getReqAttrsBid(store, marketID)
	market.ReqAttrCreateCommitment = getReqAttrsCommitment(store, marketID)
	market.AcceptingCommitments = isMarketAcceptingCommitments(store, marketID)
	market.CommitmentSettlementBips = getCommitmentSettlementBips(store, marketID)
	market.IntermediaryDenom = getIntermediaryDenom(store, marketID)

	if marketAcc := k.GetMarketAccount(ctx, marketID); marketAcc != nil {
		market.MarketDetails = marketAcc.MarketDetails
	}

	return market
}

// IterateMarkets iterates over all markets.
// The callback should return whether to stop, i.e. true = stop iterating, false = keep going.
func (k Keeper) IterateMarkets(ctx sdk.Context, cb func(market *exchange.Market) bool) {
	k.IterateKnownMarketIDs(ctx, func(marketID uint32) bool {
		market := k.GetMarket(ctx, marketID)
		return market != nil && cb(market)
	})
}

// GetMarketBrief gets the MarketBrief for the given market id.
func (k Keeper) GetMarketBrief(ctx sdk.Context, marketID uint32) *exchange.MarketBrief {
	acc := k.GetMarketAccount(ctx, marketID)
	if acc == nil {
		return nil
	}

	return &exchange.MarketBrief{
		MarketId:      marketID,
		MarketAddress: acc.Address,
		MarketDetails: acc.MarketDetails,
	}
}

// WithdrawMarketFunds transfers funds from a market account to another account.
// The caller is responsible for making sure this withdrawal should be allowed (e.g. by calling CanWithdrawMarketFunds first).
func (k Keeper) WithdrawMarketFunds(ctx sdk.Context, marketID uint32, toAddr sdk.AccAddress, amount sdk.Coins, withdrawnBy string) error {
	admin, err := sdk.AccAddressFromBech32(withdrawnBy)
	if err != nil {
		return fmt.Errorf("invalid admin %q: %w", withdrawnBy, err)
	}
	if k.bankKeeper.BlockedAddr(toAddr) {
		return fmt.Errorf("%s is not allowed to receive funds", toAddr)
	}
	marketAddr := exchange.GetMarketAddress(marketID)
	xferCtx := markertypes.WithTransferAgents(ctx, admin)
	if toAddr.Equals(admin) {
		xferCtx = quarantine.WithBypass(xferCtx)
	}
	err = k.bankKeeper.SendCoins(xferCtx, marketAddr, toAddr, amount)
	if err != nil {
		return fmt.Errorf("failed to withdraw %s from market %d: %w", amount, marketID, err)
	}
	k.emitEvent(ctx, exchange.NewEventMarketWithdraw(marketID, amount, toAddr, withdrawnBy))
	return nil
}

// ValidateMarket checks the setup of the provided market, making sure there aren't any possibly problematic settings.
func (k Keeper) ValidateMarket(ctx sdk.Context, marketID uint32) error {
	store := k.getStore(ctx)
	if err := validateMarketExists(store, marketID); err != nil {
		return err
	}

	sellerRatios := getSellerSettlementRatios(store, marketID)
	buyerRatios := getBuyerSettlementRatios(store, marketID)
	errs := exchange.ValidateRatioDenoms(sellerRatios, buyerRatios)

	bips := getCommitmentSettlementBips(store, marketID)
	convDenom := getIntermediaryDenom(store, marketID)
	if bips > 0 && len(convDenom) == 0 {
		errs = append(errs, fmt.Errorf("no intermediary denom defined, but commitment settlement bips %d is not zero", bips))
	}

	if len(convDenom) > 0 {
		feeDenom := pioconfig.GetProvenanceConfig().FeeDenom
		feeNav := k.GetNav(ctx, convDenom, feeDenom)
		if feeNav == nil {
			errs = append(errs, fmt.Errorf("no nav exists from intermediary denom %q to fee denom %q", convDenom, feeDenom))
		}
	}

	allowComs := isMarketAcceptingCommitments(store, marketID)
	if allowComs {
		createComFee := getCreateCommitmentFlatFees(store, marketID)
		if bips == 0 && len(createComFee) == 0 {
			errs = append(errs, errors.New("commitments are allowed but no commitment-related fees are defined"))
		}
	}

	return errors.Join(errs...)
}

// CloseMarket disables order and commitment creation in a market,
// cancels all its existing orders, and releases all its commitments.
func (k Keeper) CloseMarket(ctx sdk.Context, marketID uint32, signer string) {
	_ = k.UpdateMarketAcceptingOrders(ctx, marketID, false, signer)
	_ = k.UpdateMarketAcceptingCommitments(ctx, marketID, false, signer)
	k.CancelAllOrdersForMarket(ctx, marketID, signer)
	k.ReleaseAllCommitmentsForMarket(ctx, marketID)
}
