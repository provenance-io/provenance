package keeper

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/x/exchange"
)

// getCommitmentAmount gets the amount that the given address has committed to the provided market.
func getCommitmentAmount(store sdk.KVStore, marketID uint32, addr sdk.AccAddress) sdk.Coins {
	key := MakeKeyCommitment(marketID, addr)
	value := store.Get(key)
	if len(value) == 0 {
		return nil
	}
	// Skipping the error check here because I'd just be returning nil on error anyway.
	rv, _ := parseCommitmentValue(value)
	return rv
}

// parseCommitmentValue parses the store value of a commitment.
func parseCommitmentValue(value []byte) (sdk.Coins, error) {
	rv, err := sdk.ParseCoinsNormalized(string(value))
	if err != nil {
		return nil, fmt.Errorf("invalid commitment value: %w", err)
	}
	return rv, nil
}

// parseCommitmentKeyValue parses a store key and value into a commitment object.
// The keyPrefix and keySuffix are concatenated to get the full key.
// If you already have the full key, just provide it in one of those and provide nil for the other.
func parseCommitmentKeyValue(keyPrefix, keySuffix, value []byte) (*exchange.Commitment, error) {
	marketID, addr, err := ParseKeyCommitment(append(keyPrefix, keySuffix...))
	if err != nil {
		return nil, err
	}
	amount, err := parseCommitmentValue(value)
	if err != nil {
		return nil, err
	}
	return &exchange.Commitment{Account: addr.String(), MarketId: marketID, Amount: amount}, nil
}

// setCommitmentAmount sets the amount that the given address has committed to the provided market.
// If the amount is zero, the entry is deleted.
func setCommitmentAmount(store sdk.KVStore, marketID uint32, addr sdk.AccAddress, amount sdk.Coins) {
	key := MakeKeyCommitment(marketID, addr)
	if !amount.IsZero() {
		value := amount.String()
		store.Set(key, []byte(value))
	} else {
		store.Delete(key)
	}
}

// addCommitmentAmount adds the provided amount to the funds committed by the addr to the given market.
func addCommitmentAmount(store sdk.KVStore, marketID uint32, addr sdk.AccAddress, amount sdk.Coins) {
	cur := getCommitmentAmount(store, marketID, addr)
	setCommitmentAmount(store, marketID, addr, cur.Add(amount...))
}

// GetCommitmentAmount gets the amount the given address has committed to the provided market.
func (k Keeper) GetCommitmentAmount(ctx sdk.Context, marketID uint32, addr sdk.AccAddress) sdk.Coins {
	return getCommitmentAmount(k.getStore(ctx), marketID, addr)
}

// AddCommitment commits the provided amount by the addr to the given market, and places a hold on them.
// If the addr already has funds committed to the market, the provided amount is added to that.
// Otherwise a new commitment record is created.
func (k Keeper) AddCommitment(ctx sdk.Context, marketID uint32, addr sdk.AccAddress, amount sdk.Coins) error {
	if amount.IsZero() {
		return nil
	}
	if amount.IsAnyNegative() {
		return fmt.Errorf("cannot add negative commitment amount %q for %s in market %d", amount, addr, marketID)
	}
	addCommitmentAmount(k.getStore(ctx), marketID, addr, amount)
	return k.holdKeeper.AddHold(ctx, addr, amount, fmt.Sprintf("x/exchange: commitment to %d", marketID))
}

// ReleaseCommitment reduces the funds committed by an address to a market and releases the hold on those funds.
// If an amount is provided, just that amount is released.
// If the provided amount is zero, all funds committed by the address to the market are released.
func (k Keeper) ReleaseCommitment(ctx sdk.Context, marketID uint32, addr sdk.AccAddress, amount sdk.Coins) error {
	if amount.IsAnyNegative() {
		return fmt.Errorf("cannot release negative commitment amount %q for %s in market %d", amount, addr, marketID)
	}
	store := k.getStore(ctx)
	cur := getCommitmentAmount(store, marketID, addr)
	var newAmt, toRelease sdk.Coins
	if !amount.IsZero() {
		var isNeg bool
		newAmt, isNeg = cur.SafeSub(amount...)
		if isNeg {
			return fmt.Errorf("commitment amount to release %q is more than currently committed amount %q for %s in market %d",
				amount, cur, addr, marketID)
		}
		toRelease = amount
	} else {
		toRelease = cur
	}
	setCommitmentAmount(store, marketID, addr, newAmt)
	return k.holdKeeper.ReleaseHold(ctx, addr, toRelease)
}

// IterateCommitments iterates over all commitment entries in the store.
func (k Keeper) IterateCommitments(ctx sdk.Context, cb func(commitment exchange.Commitment) bool) {
	keyPrefix := GetKeyPrefixCommitments()
	k.iterate(ctx, keyPrefix, func(keySuffix, value []byte) bool {
		commitment, err := parseCommitmentKeyValue(keyPrefix, keySuffix, value)
		if err != nil || commitment == nil {
			return false
		}
		return cb(*commitment)
	})
}

// lookupNav gets a nav from the provided known navs, or if not known, gets it from the marker module.
func (k Keeper) lookupNav(ctx sdk.Context, markerDenom, priceDenom string, known []exchange.NetAssetPrice) *exchange.NetAssetPrice {
	for _, nav := range known {
		if nav.Assets.Denom == markerDenom && nav.Price.Denom == priceDenom {
			return &nav
		}
	}
	return k.GetNav(ctx, markerDenom, priceDenom)
}

// CalculateCommitmentSettlementFee calculates the fee that the exchange must be paid (by the market) for the provided
// commitment settlement request. If the market does not have a bips defined, an empty result is returned (without error).
// If no inputs are given, the result will only have the ToFeeNav field (if it exists).
func (k Keeper) CalculateCommitmentSettlementFee(ctx sdk.Context, req *exchange.MsgMarketCommitmentSettleRequest) (*exchange.QueryCommitmentSettlementFeeCalcResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("settlement request cannot be nil")
	}
	if err := req.Validate(false); err != nil {
		return nil, err
	}

	rv := &exchange.QueryCommitmentSettlementFeeCalcResponse{}
	store := k.getStore(ctx)
	bips := getCommitmentSettlementBips(store, req.MarketId)
	if bips == 0 {
		return rv, nil
	}

	convDenom := getIntermediaryDenom(store, req.MarketId)
	if len(convDenom) == 0 {
		return nil, fmt.Errorf("market %d does not have an intermediary denom", req.MarketId)
	}

	feeDenom := pioconfig.GetProvenanceConfig().FeeDenom
	rv.ToFeeNav = k.lookupNav(ctx, convDenom, feeDenom, req.Navs)
	if rv.ToFeeNav == nil {
		return nil, fmt.Errorf("no nav found from intermediary denom %q to fee denom %q", convDenom, feeDenom)
	}

	// If there aren't any inputs, there's nothing left to do here.
	if len(req.Inputs) == 0 {
		return rv, nil
	}

	rv.InputTotal = exchange.SumAccountAmounts(req.Inputs)

	var errs []error
	var convDecAmt sdkmath.LegacyDec
	for _, coin := range rv.InputTotal {
		switch coin.Denom {
		case feeDenom:
			rv.ConvertedTotal = rv.ConvertedTotal.Add(coin)
		case convDenom:
			convDecAmt = convDecAmt.Add(sdkmath.LegacyNewDecFromInt(coin.Amount))
		default:
			nav := k.lookupNav(ctx, coin.Denom, convDenom, req.Navs)
			if nav == nil {
				errs = append(errs, fmt.Errorf("no nav found from assets denom %q to intermediary denom %q", coin.Denom, convDenom))
			} else {
				newAmt := sdkmath.LegacyNewDecFromInt(coin.Amount.Mul(nav.Price.Amount)).QuoInt(nav.Assets.Amount)
				convDecAmt = convDecAmt.Add(newAmt)
				rv.ConversionNavs = append(rv.ConversionNavs, *nav)
			}
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	convAmt := convDecAmt.TruncateInt()
	if !convDecAmt.IsInteger() {
		convAmt = convAmt.AddRaw(1)
	}
	rv.ConvertedTotal = rv.ConvertedTotal.Add(sdk.NewCoin(convDenom, convAmt))

	var feeDenomTotal sdkmath.Int
	for _, coin := range rv.ConvertedTotal {
		switch coin.Denom {
		case feeDenom:
			feeDenomTotal = feeDenomTotal.Add(coin.Amount)
		case convDenom:
			asFee := exchange.QuoIntRoundUp(coin.Amount.Mul(rv.ToFeeNav.Price.Amount), rv.ToFeeNav.Assets.Amount)
			feeDenomTotal = feeDenomTotal.Add(asFee)
		default:
			// It shouldn't be possible to get here, but just in case...
			return nil, fmt.Errorf("unknown denom %q in the converted total: %q", coin.Denom, rv.ConvertedTotal)
		}
	}

	// Both the assets and price funds are in the inputs. So the sum of them is twice what
	// we usually think of as the "value of a trade." As we apply the bips, we will divide
	// by 20,000 (instead of 10,000) in order to account for that doubling.
	feeAmt := exchange.QuoIntRoundUp(feeDenomTotal.MulRaw(int64(bips)), TwentyKInt)
	rv.ExchangeFees = sdk.NewCoins(sdk.NewCoin(feeDenom, feeAmt))

	return rv, nil
}
