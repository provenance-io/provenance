package keeper

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/x/exchange"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

// getCommitmentAmount gets the amount that the given address has committed to the provided market.
func getCommitmentAmount(store storetypes.KVStore, marketID uint32, addr sdk.AccAddress) sdk.Coins {
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
func setCommitmentAmount(store storetypes.KVStore, marketID uint32, addr sdk.AccAddress, amount sdk.Coins) {
	key := MakeKeyCommitment(marketID, addr)
	if !amount.IsZero() {
		value := amount.String()
		store.Set(key, []byte(value))
	} else {
		store.Delete(key)
	}
}

// addCommitmentAmount adds the provided amount to the funds committed by the addr to the given market.
func addCommitmentAmount(store storetypes.KVStore, marketID uint32, addr sdk.AccAddress, amount sdk.Coins) {
	cur := getCommitmentAmount(store, marketID, addr)
	setCommitmentAmount(store, marketID, addr, cur.Add(amount...))
}

// validateMarketIsAcceptingCommitments makes sure the market exists and is accepting commitments.
func validateMarketIsAcceptingCommitments(store storetypes.KVStore, marketID uint32) error {
	if err := validateMarketExists(store, marketID); err != nil {
		return err
	}
	if !isMarketAcceptingCommitments(store, marketID) {
		return fmt.Errorf("market %d is not accepting commitments", marketID)
	}
	return nil
}

// validateUserCanCreateCommitment makes sure the user can create a commitment in the given market.
func (k Keeper) validateUserCanCreateCommitment(ctx sdk.Context, marketID uint32, addr sdk.AccAddress) error {
	if !k.CanCreateCommitment(ctx, marketID, addr) {
		return fmt.Errorf("account %s is not allowed to create commitments in market %d", addr, marketID)
	}
	return nil
}

// GetCommitmentAmount gets the amount the given address has committed to the provided market.
func (k Keeper) GetCommitmentAmount(ctx sdk.Context, marketID uint32, addr sdk.AccAddress) sdk.Coins {
	return getCommitmentAmount(k.getStore(ctx), marketID, addr)
}

// addCommitment commits the provided amount by the addr to the given market, and places a hold on them.
// If the addr already has funds committed to the market, the provided amount is added to that.
// Otherwise a new commitment record is created.
//
// If doMarketChecks = true, the market must be accepting commitments and the addr must have the required attributes.
func (k Keeper) addCommitment(ctx sdk.Context, marketID uint32, addr sdk.AccAddress, amount sdk.Coins, eventTag string, doMarketChecks bool) error {
	if amount.IsZero() {
		return nil
	}
	if amount.IsAnyNegative() {
		return fmt.Errorf("cannot add negative commitment amount %q for %s in market %d", amount, addr, marketID)
	}

	store := k.getStore(ctx)
	if doMarketChecks {
		if err := validateMarketIsAcceptingCommitments(store, marketID); err != nil {
			return err
		}
		if err := k.validateUserCanCreateCommitment(ctx, marketID, addr); err != nil {
			return err
		}
	}

	err := k.holdKeeper.AddHold(ctx, addr, amount, fmt.Sprintf("x/exchange: commitment to %d", marketID))
	if err != nil {
		return err
	}

	addCommitmentAmount(k.getStore(ctx), marketID, addr, amount)
	k.emitEvent(ctx, exchange.NewEventFundsCommitted(addr.String(), marketID, amount, eventTag))
	return nil
}

// AddCommitment commits the provided amount by the addr to the given market, and places a hold on them.
// If the addr already has funds committed to the market, the provided amount is added to that.
// Otherwise a new commitment record is created.
func (k Keeper) AddCommitment(ctx sdk.Context, marketID uint32, addr sdk.AccAddress, amount sdk.Coins, eventTag string) error {
	return k.addCommitment(ctx, marketID, addr, amount, eventTag, true)
}

// addCommitmentsUnsafe adds several commitments without checking that the market is accepting
// commitments and without checking that the address has the required attributes.
func (k Keeper) addCommitmentsUnsafe(ctx sdk.Context, marketID uint32, toAdd []exchange.AccountAmount, eventTag string) error {
	var errs []error
	for _, entry := range toAdd {
		addr, err := sdk.AccAddressFromBech32(entry.Account)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid account %q: %w", entry.Account, err))
			continue
		}
		err = k.addCommitment(ctx, marketID, addr, entry.Amount, eventTag, false)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// ReleaseCommitment reduces the funds committed by an address to a market and releases the hold on those funds.
// If an amount is provided, just that amount is released.
// If the provided amount is zero, all funds committed by the address to the market are released.
func (k Keeper) ReleaseCommitment(ctx sdk.Context, marketID uint32, addr sdk.AccAddress, amount sdk.Coins, eventTag string) error {
	if amount.IsAnyNegative() {
		return fmt.Errorf("cannot release negative commitment amount %q for %s in market %d", amount, addr, marketID)
	}

	store := k.getStore(ctx)
	cur := getCommitmentAmount(store, marketID, addr)
	if cur.IsZero() {
		return fmt.Errorf("account %s does not have any funds committed to market %d", addr, marketID)
	}

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

	err := k.holdKeeper.ReleaseHold(ctx, addr, toRelease)
	if err != nil {
		return err
	}

	setCommitmentAmount(store, marketID, addr, newAmt)
	k.emitEvent(ctx, exchange.NewEventCommitmentReleased(addr.String(), marketID, toRelease, eventTag))
	return nil
}

// ReleaseCommitments calls ReleaseCommitment for several entries.
func (k Keeper) ReleaseCommitments(ctx sdk.Context, marketID uint32, toRelease []exchange.AccountAmount, eventTag string) error {
	var errs []error
	for _, entry := range toRelease {
		addr, err := sdk.AccAddressFromBech32(entry.Account)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid account %q: %w", entry.Account, err))
			continue
		}
		err = k.ReleaseCommitment(ctx, marketID, addr, entry.Amount, eventTag)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// ReleaseAllCommitmentsForMarket releases all the commitments (and related holds)
// that have been made to the market.
func (k Keeper) ReleaseAllCommitmentsForMarket(ctx sdk.Context, marketID uint32) {
	var keySuffixes [][]byte
	keyPrefix := GetKeyPrefixCommitmentsToMarket(marketID)
	k.iterate(ctx, keyPrefix, func(keySuffix, _ []byte) bool {
		keySuffixes = append(keySuffixes, keySuffix)
		return false
	})

	var errs []error
	for _, keySuffix := range keySuffixes {
		addr, err := ParseKeySuffixCommitment(keySuffix)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to parse addr from key suffix %x: %w", keySuffix, err))
			continue
		}
		err = k.ReleaseCommitment(ctx, marketID, addr, nil, "GovCloseMarket")
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		k.logErrorf(ctx, "%d error(s) encountered releasing all commitments for market %d:\n%v",
			len(errs), marketID, errors.Join(errs...))
	}
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

// ValidateAndCollectCommitmentCreationFee verifies that the provided commitment
// creation fee is sufficient and collects it.
func (k Keeper) ValidateAndCollectCommitmentCreationFee(ctx sdk.Context, marketID uint32, addr sdk.AccAddress, fee *sdk.Coin) error {
	if err := validateCreateCommitmentFlatFee(k.getStore(ctx), marketID, fee); err != nil {
		return err
	}
	if fee == nil {
		return nil
	}

	err := k.CollectFee(ctx, marketID, addr, sdk.Coins{*fee})
	if err != nil {
		return fmt.Errorf("error collecting commitment creation fee: %w", err)
	}

	return nil
}

// lookupNav gets a nav from the provided known navs, or if not known, gets it from the marker or metadata module.
func (k Keeper) lookupNav(ctx sdk.Context, assetsDenom, priceDenom string, known []exchange.NetAssetPrice) *exchange.NetAssetPrice {
	for _, nav := range known {
		if nav.Assets.Denom == assetsDenom && nav.Price.Denom == priceDenom {
			return &nav
		}
	}
	return k.GetNav(ctx, assetsDenom, priceDenom)
}

// CalculateCommitmentSettlementFee calculates the fee that the exchange must be paid (by the market) for the provided
// commitment settlement request. If the market does not have a bips defined, an empty result is returned (without error).
// If no inputs are given, the result will only have the ToFeeNav field (if it exists).
func (k Keeper) CalculateCommitmentSettlementFee(ctx sdk.Context, req *exchange.MsgMarketCommitmentSettleRequest) (*exchange.QueryCommitmentSettlementFeeCalcResponse, error) {
	if req == nil {
		return nil, errors.New("settlement request cannot be nil")
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
	if convDenom != feeDenom {
		rv.ToFeeNav = k.lookupNav(ctx, convDenom, feeDenom, req.Navs)
		if rv.ToFeeNav == nil {
			return nil, fmt.Errorf("no nav found from intermediary denom %q to fee denom %q", convDenom, feeDenom)
		}
	} else {
		rv.ToFeeNav = &exchange.NetAssetPrice{Assets: sdk.NewInt64Coin(feeDenom, 1), Price: sdk.NewInt64Coin(feeDenom, 1)}
	}

	// If there aren't any inputs, there's nothing left to do here.
	if len(req.Inputs) == 0 {
		return rv, nil
	}

	rv.InputTotal = exchange.SumAccountAmounts(req.Inputs)

	var errs []error
	convDecAmt := sdkmath.LegacyZeroDec()
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

	feeDenomTotal := sdkmath.ZeroInt()
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

// SettleCommitments orchestrates the transfer of committed funds and collection of fees by the market.
func (k Keeper) SettleCommitments(ctx sdk.Context, req *exchange.MsgMarketCommitmentSettleRequest) error {
	admin, adminErr := sdk.AccAddressFromBech32(req.Admin)
	if adminErr != nil {
		return fmt.Errorf("invalid admin %q: %w", req.Admin, adminErr)
	}

	// Record all the navs.
	k.recordNAVs(ctx, req.MarketId, req.Navs)

	// Build the transfers
	inputs := exchange.SimplifyAccountAmounts(req.Inputs)
	outputs := exchange.SimplifyAccountAmounts(req.Outputs)
	fees := exchange.SimplifyAccountAmounts(req.Fees)
	transfers, err := exchange.BuildCommitmentTransfers(req.MarketId, inputs, outputs, fees)
	if err != nil {
		return fmt.Errorf("failed to build transfers: %w", err)
	}

	// Release the commitments on the inputs and fees
	inputsAndFees := make([]exchange.AccountAmount, 0, len(inputs)+len(fees))
	inputsAndFees = append(inputsAndFees, inputs...)
	inputsAndFees = append(inputsAndFees, fees...)
	err = k.ReleaseCommitments(ctx, req.MarketId, exchange.SimplifyAccountAmounts(inputsAndFees), req.EventTag)
	if err != nil {
		return fmt.Errorf("failed to release commitments on inputs and fees: %w", err)
	}

	// Do the transfers
	xFerCtx := markertypes.WithTransferAgents(ctx, admin)
	var xferErrs []error
	for _, transfer := range transfers {
		err = k.DoTransfer(xFerCtx, transfer.Inputs, transfer.Outputs)
		if err != nil {
			xferErrs = append(xferErrs, err)
		}
	}
	if len(xferErrs) > 0 {
		return errors.Join(xferErrs...)
	}

	// Commit the funds in the outputs.
	err = k.addCommitmentsUnsafe(ctx, req.MarketId, outputs, req.EventTag)
	if err != nil {
		return fmt.Errorf("failed to re-commit funds after transfer: %w", err)
	}

	return nil
}

// consumeCommitmentSettlementFee calculates and consumes the commitment settlement fee for the given request.
func (k Keeper) consumeCommitmentSettlementFee(ctx sdk.Context, req *exchange.MsgMarketCommitmentSettleRequest) error {
	calcResp, err := k.CalculateCommitmentSettlementFee(ctx, req)
	if err != nil {
		return fmt.Errorf("could not calculate commitment settlement fees: %w", err)
	}
	antewrapper.ConsumeAdditionalFee(ctx, calcResp.ExchangeFees)
	return nil
}
