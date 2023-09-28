package exchange

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// OrderSplit contains an order, and the asset and price amounts that should come out of it.
type OrderSplit struct {
	// Order fulfillment associated with this split.
	Order *OrderFulfillment
	// Assets is the amount of assets from the order involved in this split.
	Assets sdk.Coin
	// Price is the amount of the price from the order involved in this split.
	Price sdk.Coin
}

// OrderFulfillment is used to figure out how an order should be fulfilled.
type OrderFulfillment struct {
	// Order is the original order with all its information.
	Order *Order

	// Splits contains information on the orders being used to fulfill this order.
	Splits []*OrderSplit

	// AssetsFilledAmt is the total amount of assets being fulfilled for the order.
	AssetsFilledAmt sdkmath.Int
	// AssetsUnfilledAmt is the amount of order assets that have not yet been fulfilled for the order.
	AssetsUnfilledAmt sdkmath.Int
	// PriceAppliedAmt is the total price amount involved in this order fulfillment.
	// If this is a bid order, it's the actual amount the buyer will pay.
	// If this is an ask order it's the actual amount the seller will receive.
	PriceAppliedAmt sdkmath.Int
	// PriceLeftAmt is the price that has not yet been fulfilled for the order.
	// This can be negative for ask orders that are being filled at a higher price than requested.
	PriceLeftAmt sdkmath.Int

	// IsFinalized is set to true once Finalize() is called without error.
	IsFinalized bool
	// FeesToPay is the amount of settlement fees the order owner should pay to settle this order.
	// This is only set during Finalize().
	FeesToPay sdk.Coins
	// OrderFeesLeft is the amount fees settlement left to pay (if this order is only partially filled).
	// This is only set during Finalize().
	OrderFeesLeft sdk.Coins
	// PriceFilledAmt is the amount of the order price that is being filled.
	// This is only set during Finalize().
	PriceFilledAmt sdkmath.Int
	// PriceUnfilledAmt is the amount of the order price that is not being filled.
	// This is only set during Finalize().
	PriceUnfilledAmt sdkmath.Int
}

var _ OrderI = (*OrderFulfillment)(nil)

// NewOrderFulfillment creates a new OrderFulfillment wrapping the provided order.
func NewOrderFulfillment(order *Order) *OrderFulfillment {
	return &OrderFulfillment{
		Order:             order,
		AssetsFilledAmt:   sdkmath.ZeroInt(),
		AssetsUnfilledAmt: order.GetAssets().Amount,
		PriceAppliedAmt:   sdkmath.ZeroInt(),
		PriceLeftAmt:      order.GetPrice().Amount,
		PriceFilledAmt:    sdkmath.ZeroInt(),
		PriceUnfilledAmt:  sdkmath.ZeroInt(),
	}
}

// GetAssetsFilled gets the coin value of the assets that have been filled in this fulfillment.
func (f OrderFulfillment) GetAssetsFilled() sdk.Coin {
	return sdk.Coin{Denom: f.GetAssets().Denom, Amount: f.AssetsFilledAmt}
}

// GetAssetsUnfilled gets the coin value of the assets left to fill in this fulfillment.
func (f OrderFulfillment) GetAssetsUnfilled() sdk.Coin {
	return sdk.Coin{Denom: f.GetAssets().Denom, Amount: f.AssetsUnfilledAmt}
}

// GetPriceApplied gets the coin value of the price that has been filled in this fulfillment.
func (f OrderFulfillment) GetPriceApplied() sdk.Coin {
	return sdk.Coin{Denom: f.GetPrice().Denom, Amount: f.PriceAppliedAmt}
}

// GetPriceLeft gets the coin value of the price left to fill in this fulfillment.
func (f OrderFulfillment) GetPriceLeft() sdk.Coin {
	return sdk.Coin{Denom: f.GetPrice().Denom, Amount: f.PriceLeftAmt}
}

// GetPriceFilled gets the coin value of the price filled in this fulfillment.
func (f OrderFulfillment) GetPriceFilled() sdk.Coin {
	return sdk.Coin{Denom: f.GetPrice().Denom, Amount: f.PriceFilledAmt}
}

// GetPriceUnfilled gets the coin value of the price unfilled in this fulfillment.
func (f OrderFulfillment) GetPriceUnfilled() sdk.Coin {
	return sdk.Coin{Denom: f.GetPrice().Denom, Amount: f.PriceUnfilledAmt}
}

// IsFullyFilled returns true if this fulfillment's order has been fully accounted for.
func (f OrderFulfillment) IsFullyFilled() bool {
	return !f.AssetsUnfilledAmt.IsPositive()
}

// IsCompletelyUnfulfilled returns true if nothing in this order has been filled.
func (f OrderFulfillment) IsCompletelyUnfulfilled() bool {
	return f.AssetsFilledAmt.IsZero()
}

// GetOrderID gets this fulfillment's order's id.
func (f OrderFulfillment) GetOrderID() uint64 {
	return f.Order.GetOrderID()
}

// IsAskOrder returns true if this is an ask order.
func (f OrderFulfillment) IsAskOrder() bool {
	return f.Order.IsAskOrder()
}

// IsBidOrder returns true if this is an ask order.
func (f OrderFulfillment) IsBidOrder() bool {
	return f.Order.IsBidOrder()
}

// GetMarketID gets this fulfillment's order's market id.
func (f OrderFulfillment) GetMarketID() uint32 {
	return f.Order.GetMarketID()
}

// GetOwner gets this fulfillment's order's owner.
func (f OrderFulfillment) GetOwner() string {
	return f.Order.GetOwner()
}

// GetAssets gets this fulfillment's order's assets.
func (f OrderFulfillment) GetAssets() sdk.Coin {
	return f.Order.GetAssets()
}

// GetPrice gets this fulfillment's order's price.
func (f OrderFulfillment) GetPrice() sdk.Coin {
	return f.Order.GetPrice()
}

// GetSettlementFees gets this fulfillment's order's settlement fees.
func (f OrderFulfillment) GetSettlementFees() sdk.Coins {
	return f.Order.GetSettlementFees()
}

// PartialFillAllowed gets this fulfillment's order's AllowPartial flag.
func (f OrderFulfillment) PartialFillAllowed() bool {
	return f.Order.PartialFillAllowed()
}

// GetOrderType gets this fulfillment's order's type string.
func (f OrderFulfillment) GetOrderType() string {
	return f.Order.GetOrderType()
}

// GetOrderTypeByte gets this fulfillment's order's type byte.
func (f OrderFulfillment) GetOrderTypeByte() byte {
	return f.Order.GetOrderTypeByte()
}

// GetHoldAmount gets this fulfillment's order's hold amount.
func (f OrderFulfillment) GetHoldAmount() sdk.Coins {
	return f.Order.GetHoldAmount()
}

// Apply adjusts this order fulfillment using the provided info.
func (f *OrderFulfillment) Apply(order *OrderFulfillment, assetsAmt, priceAmt sdkmath.Int) error {
	assets := sdk.NewCoin(order.GetAssets().Denom, assetsAmt)
	price := sdk.NewCoin(order.GetPrice().Denom, priceAmt)

	newAssetsUnfilledAmt := f.AssetsUnfilledAmt.Sub(assetsAmt)
	if newAssetsUnfilledAmt.IsNegative() {
		return fmt.Errorf("cannot fill %s order %d having assets left %q with %q from %s order %d: overfill",
			f.GetOrderType(), f.GetOrderID(), f.GetAssetsUnfilled(), assets, order.GetOrderType(), order.GetOrderID())
	}

	newPriceLeftAmt := f.PriceLeftAmt.Sub(priceAmt)
	// ask orders are allow to go negative on price left, but bid orders are not.
	if newPriceLeftAmt.IsNegative() && f.IsBidOrder() {
		return fmt.Errorf("cannot fill %s order %d having price left %q to %s order %d at a price of %q: overfill",
			f.GetOrderType(), f.GetOrderID(), f.GetPriceLeft(), order.GetOrderType(), order.GetOrderID(), price)
	}

	f.AssetsUnfilledAmt = newAssetsUnfilledAmt
	f.AssetsFilledAmt = f.AssetsFilledAmt.Add(assetsAmt)
	f.PriceLeftAmt = newPriceLeftAmt
	f.PriceAppliedAmt = f.PriceAppliedAmt.Add(priceAmt)
	f.Splits = append(f.Splits, &OrderSplit{
		Order:  order,
		Assets: assets,
		Price:  price,
	})
	return nil
}

// ApplyLeftoverPrice increases this fulfillment and the provided split
// using info in the split and the provided amount.
func (f *OrderFulfillment) ApplyLeftoverPrice(askSplit *OrderSplit, amt sdkmath.Int) {
	// Update this fulfillment to indicate that the amount has been applied.
	f.PriceLeftAmt = f.PriceLeftAmt.Sub(amt)
	f.PriceAppliedAmt = f.PriceAppliedAmt.Add(amt)

	// Update the ask split to include the extra amount.
	askSplit.Price.Amount = askSplit.Price.Amount.Add(amt)
	// And update the ask split's fulfillment similarly.
	askSplit.Order.PriceLeftAmt = askSplit.Order.PriceLeftAmt.Sub(amt)
	askSplit.Order.PriceAppliedAmt = askSplit.Order.PriceAppliedAmt.Add(amt)

	// Update the bid split entry for this order in the splits that the ask split has
	// to indicate the extra amount from this bid.
	orderID := f.GetOrderID()
	for _, bidSplit := range askSplit.Order.Splits {
		if bidSplit.Order.GetOrderID() == orderID {
			bidSplit.Price.Amount = bidSplit.Price.Amount.Add(amt)
			return
		}
	}

	// If we didn't find a bid split to update, something is horribly wrong.
	panic(fmt.Errorf("could not apply leftover amount %s from %s order %d to %s order %d: bid split not found",
		amt, f.GetOrderType(), orderID, askSplit.Order.GetOrderType(), askSplit.Order.GetOrderID()))
}

// Finalize does some final calculations and validation for this order fulfillment.
// This order fulfillment and the ones in it maybe updated during this.
func (f *OrderFulfillment) Finalize(sellerFeeRatio *FeeRatio) (err error) {
	// If this is returning an error, unset all the fields that get set in here.
	defer func() {
		if err != nil {
			f.IsFinalized = false
			f.PriceFilledAmt = sdkmath.ZeroInt()
			f.PriceUnfilledAmt = sdkmath.ZeroInt()
			f.FeesToPay = nil
			f.OrderFeesLeft = nil
		}
	}()

	// AssetsFilledAmt cannot be zero here because we'll be dividing by it.
	// AssetsFilledAmt cannot be negative here because we can't have negative values from the calcs.
	// Checking for assets filled > zero here (instead of in Validate) because we need to divide by it in here.
	if !f.AssetsFilledAmt.IsPositive() {
		return fmt.Errorf("no assets filled in %s order %d", f.GetOrderType(), f.GetOrderID())
	}

	isAskOrder, isBidOrder := f.IsAskOrder(), f.IsBidOrder()
	isFullyFilled := f.IsFullyFilled()
	orderFees := f.GetSettlementFees()
	orderAssets := f.GetAssets()
	orderPrice := f.GetPrice()

	f.PriceFilledAmt = orderPrice.Amount
	f.PriceUnfilledAmt = sdkmath.ZeroInt()
	f.FeesToPay = orderFees
	f.OrderFeesLeft = nil

	if !isFullyFilled {
		// Make sure the price can be split on a whole number, and figure out the price being filled.
		priceAssets := orderPrice.Amount.Mul(f.AssetsFilledAmt)
		priceRem := priceAssets.Mod(orderAssets.Amount)
		if !priceRem.IsZero() {
			return fmt.Errorf("%s order %d having assets %q cannot be partially filled by %q: "+
				"price %q is not evenly divisible",
				f.GetOrderType(), f.GetOrderID(), orderAssets, f.GetAssetsFilled(), orderPrice)
		}
		f.PriceFilledAmt = priceAssets.Quo(orderAssets.Amount)
		f.PriceUnfilledAmt = orderPrice.Amount.Sub(f.PriceFilledAmt)

		// Make sure the fees can be split on a whole number, and figure out how much is actually being paid of them.
		f.FeesToPay = nil
		for _, orderFee := range orderFees {
			feeAssets := orderFee.Amount.Mul(f.AssetsFilledAmt)
			feeRem := feeAssets.Mod(orderAssets.Amount)
			if !feeRem.IsZero() {
				return fmt.Errorf("%s order %d having assets %q cannot be partially filled by %q: "+
					"fee %q is not evenly divisible",
					f.GetOrderType(), f.GetOrderID(), orderAssets, f.GetAssetsFilled(), orderFee)
			}
			feeAmtToPay := feeAssets.Quo(orderAssets.Amount)
			f.FeesToPay = f.FeesToPay.Add(sdk.NewCoin(orderFee.Denom, feeAmtToPay))
		}
		f.OrderFeesLeft = orderFees.Sub(f.FeesToPay...)
	}

	switch {
	case isAskOrder:
		// For ask orders, we need to calculate and add the ratio fee to the fees to pay.
		// This should NOT affect the order fees left.
		if sellerFeeRatio != nil {
			ratioFeeToPay, ferr := sellerFeeRatio.ApplyToLoosely(f.GetPriceApplied())
			if ferr != nil {
				return fmt.Errorf("could not calculate %s order %d ratio fee: %w",
					f.GetOrderType(), f.GetOrderID(), ferr)
			}
			f.FeesToPay = f.FeesToPay.Add(ratioFeeToPay)
		}
	case isBidOrder:
		// When adding things to PriceAppliedAmt (and Splits .Price), we used truncation on the divisions.
		// When calculated PriceFilledAmt, we made sure it was a whole number based on total assets being distributed.
		// So, at this point, PriceAppliedAmt might be a little less than the PriceFilledAmt.
		// If that's the case, we'll distribute the difference among the splits.
		toDistribute := f.PriceFilledAmt.Sub(f.PriceAppliedAmt)
		if toDistribute.IsPositive() {
			distLeft := toDistribute
			// First, go through each split, and apply the leftovers to any asks that still have price left.
			for _, askSplit := range f.Splits {
				if askSplit.Order.PriceLeftAmt.IsPositive() {
					toDist := MinSDKInt(askSplit.Order.PriceLeftAmt, distLeft)
					f.ApplyLeftoverPrice(askSplit, toDist)
					distLeft = distLeft.Sub(toDist)
				}
			}

			// Now try to distribute the leftovers evenly weighted by assets.
			// First pass, we won't default to 1 (if the calc comes up zero).
			// This helps weigh larger orders that are at the end of the list.
			// Once they've all had a chance, use a minimum of 1.
			minOne := false
			for distLeft.IsPositive() {
				for _, askSplit := range f.Splits {
					distAmt := toDistribute.Mul(askSplit.Assets.Amount).Quo(f.AssetsFilledAmt)
					if distAmt.IsZero() {
						if !minOne {
							continue
						}
						distAmt = sdkmath.OneInt()
					}
					distAmt = MinSDKInt(distAmt, distLeft)

					f.ApplyLeftoverPrice(askSplit, distAmt)

					distLeft = distLeft.Sub(distAmt)
					if !distLeft.IsPositive() {
						break
					}
				}
				minOne = true
			}
		}
	}

	f.IsFinalized = true
	return nil
}

// Validate does some final validation and sanity checking on this order fulfillment.
// It's assumed that Finalize has been called before calling this.
func (f OrderFulfillment) Validate() error {
	if _, err := f.Order.GetSubOrder(); err != nil {
		return err
	}
	if !f.IsFinalized {
		return fmt.Errorf("fulfillment for %s order %d has not been finalized", f.GetOrderType(), f.GetOrderID())
	}

	orderAssets := f.GetAssets()
	if f.AssetsUnfilledAmt.IsNegative() {
		return fmt.Errorf("%s order %d having assets %q has negative assets left %q after filling %q",
			f.GetOrderType(), f.GetOrderID(), orderAssets, f.GetAssetsUnfilled(), f.GetAssetsFilled())
	}
	if !f.AssetsFilledAmt.IsPositive() {
		return fmt.Errorf("cannot fill non-positive assets %q on %s order %d having assets %q",
			f.GetAssetsFilled(), f.GetOrderType(), f.GetOrderID(), orderAssets)
	}
	trackedAssetsAmt := f.AssetsFilledAmt.Add(f.AssetsUnfilledAmt)
	if !orderAssets.Amount.Equal(trackedAssetsAmt) {
		return fmt.Errorf("tracked assets %q does not equal %s order %d assets %q",
			sdk.Coin{Denom: orderAssets.Denom, Amount: trackedAssetsAmt}, f.GetOrderType(), f.GetOrderID(), orderAssets)
	}

	orderPrice := f.GetPrice()
	if f.PriceLeftAmt.Equal(orderPrice.Amount) {
		return fmt.Errorf("price left %q equals %s order %d price %q",
			f.GetPriceLeft(), f.GetOrderType(), f.GetOrderID(), orderPrice)
	}
	if !f.PriceAppliedAmt.IsPositive() {
		return fmt.Errorf("cannot apply non-positive price %q to %s order %d having price %q",
			f.GetPriceApplied(), f.GetOrderType(), f.GetOrderID(), orderPrice)
	}
	trackedPriceAmt := f.PriceAppliedAmt.Add(f.PriceLeftAmt)
	if !orderPrice.Amount.Equal(trackedPriceAmt) {
		return fmt.Errorf("tracked price %q does not equal %s order %d price %q",
			sdk.Coin{Denom: orderPrice.Denom, Amount: trackedPriceAmt}, f.GetOrderType(), f.GetOrderID(), orderPrice)
	}
	if f.PriceUnfilledAmt.IsNegative() {
		return fmt.Errorf("%s order %d having price %q has negative price %q after filling %q",
			f.GetOrderType(), f.GetOrderID(), orderPrice, f.GetPriceUnfilled(), f.GetPriceFilled())
	}
	if !f.PriceFilledAmt.IsPositive() {
		return fmt.Errorf("cannot fill %s order %d having price %q with non-positive price %q",
			f.GetOrderType(), f.GetOrderID(), orderPrice, f.GetPriceFilled())
	}
	totalPriceAmt := f.PriceFilledAmt.Add(f.PriceUnfilledAmt)
	if !orderPrice.Amount.Equal(totalPriceAmt) {
		return fmt.Errorf("filled price %q plus unfilled price %q does not equal order price %q for %s order %d",
			f.GetPriceFilled(), f.GetPriceUnfilled(), orderPrice, f.GetOrderType(), f.GetOrderID())
	}

	if len(f.Splits) == 0 {
		return fmt.Errorf("no splits applied to %s order %d", f.GetOrderType(), f.GetOrderID())
	}

	var splitsAssets, splitsPrice sdk.Coins
	for _, split := range f.Splits {
		splitsAssets = splitsAssets.Add(split.Assets)
		splitsPrice = splitsPrice.Add(split.Price)
	}

	if len(splitsAssets) != 1 {
		return fmt.Errorf("multiple asset denoms %q in splits applied to %s order %d",
			splitsAssets, f.GetOrderType(), f.GetOrderID())
	}
	if splitsAssets[0].Denom != orderAssets.Denom {
		return fmt.Errorf("splits asset denom %q does not equal order assets %q on %s order %d",
			splitsAssets, orderAssets, f.GetOrderType(), f.GetOrderID())
	}
	if !splitsAssets[0].Amount.Equal(f.AssetsFilledAmt) {
		return fmt.Errorf("splits asset total %q does not equal filled assets %q on %s order %d",
			splitsAssets, orderAssets, f.GetOrderType(), f.GetOrderID())
	}

	if len(splitsPrice) != 1 {
		return fmt.Errorf("multiple price denoms %q in splits applied to %s order %d",
			splitsPrice, f.GetOrderType(), f.GetOrderID())
	}
	if splitsPrice[0].Denom != orderPrice.Denom {
		return fmt.Errorf("splits price denom %q does not equal order price %q on %s order %d",
			splitsPrice, orderPrice, f.GetOrderType(), f.GetOrderID())
	}
	if !splitsPrice[0].Amount.Equal(f.PriceFilledAmt) {
		return fmt.Errorf("splits price total %q does not equal applied price %q on %s order %d",
			splitsPrice, f.GetPriceApplied(), f.GetOrderType(), f.GetOrderID())
	}

	orderFees := f.GetSettlementFees()
	if f.OrderFeesLeft.IsAnyNegative() {
		return fmt.Errorf("%s order %d settlement fees left %q is negative",
			f.GetOrderType(), f.GetOrderID(), f.OrderFeesLeft)
	}
	if _, hasNeg := orderFees.SafeSub(f.OrderFeesLeft...); hasNeg {
		return fmt.Errorf("settlement fees left %q is greater than %s order %q settlement fees %q",
			f.OrderFeesLeft, f.GetOrderType(), f.GetOrderID(), orderFees)
	}

	isFullyFilled := f.IsFullyFilled()
	if isFullyFilled {
		if !f.AssetsUnfilledAmt.IsZero() {
			return fmt.Errorf("fully filled %s order %q has non-zero unfilled assets %q",
				f.GetOrderType(), f.GetOrderID(), f.GetAssetsUnfilled())
		}
		if !f.PriceUnfilledAmt.IsZero() {
			return fmt.Errorf("fully filled %s order %q has non-zero unfilled price %q",
				f.GetOrderType(), f.GetOrderID(), f.GetPriceUnfilled())
		}
		if !f.OrderFeesLeft.IsZero() {
			return fmt.Errorf("fully filled %s order %q has non-zero settlement fees left %q",
				f.GetOrderType(), f.GetOrderID(), f.OrderFeesLeft)
		}
	}

	switch {
	case f.IsAskOrder():
		// For ask orders, the applied amount needs to be at least the filled amount.
		if f.PriceFilledAmt.LT(f.PriceAppliedAmt) {
			return fmt.Errorf("%s order %d having assets %q and price %q cannot be filled by %q at price %q: unsufficient price",
				f.GetOrderType(), f.GetOrderID(), orderAssets, orderPrice, f.GetAssetsFilled(), f.GetPriceApplied())
		}
		// If not being fully filled on an order that has some fees, make sure that there's at most 1 denom in the fees left.
		if !isFullyFilled && len(orderFees) > 0 && len(f.OrderFeesLeft) > 1 {
			return fmt.Errorf("partial fulfillment for %s order %d having seller settlement fees %q has multiple denoms in fees left %q",
				f.GetOrderType(), f.GetOrderID(), orderFees, f.OrderFeesLeft)
		}
	case f.IsBidOrder():
		// If filled in full, the PriceAppliedAmt must be equal to the order price.
		if isFullyFilled && !f.PriceAppliedAmt.Equal(orderPrice.Amount) {
			return fmt.Errorf("%s order %d having price %q cannot be fully filled at price %q: price mismatch",
				f.GetOrderType(), f.GetOrderID(), orderPrice, f.GetPriceApplied())
		}
		// otherwise, the price filled must be less than the order price.
		if !isFullyFilled && orderPrice.Amount.LT(f.PriceAppliedAmt) {
			return fmt.Errorf("%s order %d having price %q cannot be partially filled at price %q: price mismatch",
				f.GetOrderType(), f.GetOrderID(), orderPrice, f.GetPriceApplied())
		}

		// For bid orders, fees to pay + fees left should equal the order fees.
		trackedFees := f.FeesToPay.Add(f.OrderFeesLeft...)
		if !CoinsEquals(trackedFees, orderFees) {
			return fmt.Errorf("tracked settlement fees %q does not equal %s order %d settlement fees %q",
				trackedFees, f.GetOrderType(), f.GetOrderID(), orderFees)
		}
	default:
		// The only way to trigger this would be to add a new order type but not add a case for it in this switch.
		panic(fmt.Errorf("case missing for %T in Validate", f.GetOrderType()))
	}

	// Saving this simple check for last in the hopes that a previous error exposes why this
	// order might accidentally be only partially filled.
	if !isFullyFilled && !f.PartialFillAllowed() {
		return fmt.Errorf("cannot fill %s order %d having assets %q with assets %q: order does not allow partial fill",
			f.GetOrderType(), f.GetOrderID(), orderAssets, f.GetAssetsFilled())
	}

	return nil
}

// Fulfill attempts to use the two provided order fulfillments to fulfill each other.
// The provided order fulfillments will be updated if everything goes okay.
func Fulfill(of1, of2 *OrderFulfillment) error {
	order1Type := of1.Order.GetOrderType()
	order2Type := of2.Order.GetOrderType()
	if order1Type == order2Type {
		return fmt.Errorf("cannot fulfill %s order %d with %s order %d",
			order1Type, of1.Order.OrderId, order2Type, of2.Order.OrderId)
	}

	var askOF, bidOF *OrderFulfillment
	if order1Type == OrderTypeAsk {
		askOF = of1
		bidOF = of2
	} else {
		askOF = of2
		bidOF = of1
	}

	askOrder, bidOrder := askOF.Order.GetAskOrder(), bidOF.Order.GetBidOrder()
	if askOrder.Price.Denom != bidOrder.Price.Denom {
		return fmt.Errorf("cannot fill bid order %d having price %q with ask order %d having price %q: denom mismatch",
			bidOF.GetOrderID(), bidOrder.Price, askOF.GetOrderID(), askOrder.Price)
	}

	assetsAmt, err := GetFulfillmentAssetsAmt(askOF, bidOF)
	if err != nil {
		return err
	}

	// We calculate the price amount based off the original bid order assets (as opposed to assets left)
	// for consistent truncation and remainders. Once we've identified all the fulfillment relationships,
	// we'll enumerate and redistribute those remainders.
	priceAmt := bidOrder.Price.Amount
	if !assetsAmt.Equal(bidOrder.Assets.Amount) {
		priceAmt = bidOrder.Price.Amount.Mul(assetsAmt).Quo(bidOrder.Assets.Amount)
	}

	askErr := askOF.Apply(bidOF, assetsAmt, priceAmt)
	bidErr := bidOF.Apply(askOF, assetsAmt, priceAmt)

	return errors.Join(askErr, bidErr)
}

// GetFulfillmentAssetsAmt figures out the assets that can be fulfilled with the two provided orders.
func GetFulfillmentAssetsAmt(of1, of2 *OrderFulfillment) (sdkmath.Int, error) {
	if !of1.AssetsUnfilledAmt.IsPositive() || !of2.AssetsUnfilledAmt.IsPositive() {
		return sdkmath.ZeroInt(), fmt.Errorf("cannot fill %s order %d having assets left %q "+
			"with %s order %d having assets left %q: zero or negative assets left",
			of1.GetOrderType(), of1.GetOrderID(), of1.GetAssetsUnfilled(),
			of2.GetOrderType(), of2.GetOrderID(), of2.GetAssetsUnfilled())
	}

	// Return the lesser of the two.
	if of1.AssetsUnfilledAmt.LTE(of2.AssetsUnfilledAmt) {
		return of1.AssetsUnfilledAmt, nil
	}
	return of2.AssetsUnfilledAmt, nil
}

// Fulfillments contains information on how orders are to be fulfilled.
type Fulfillments struct {
	// AskOFs are all the ask orders and how they are to be filled.
	AskOFs []*OrderFulfillment
	// BidOFs are all the bid orders and how they are to be filled.
	BidOFs []*OrderFulfillment
	// PartialOrder contains info on an order that is only being partially filled.
	// The transfers For part of its funds are included in the order fulfillments.
	PartialOrder *PartialFulfillment
}

// PartialFulfillment contains the remains of a partially filled order, and info on what was filled.
type PartialFulfillment struct {
	// NewOrder is an updated version of the partially filled order with reduced amounts.
	NewOrder *Order
	// AssetsFilled is the amount of order assets that were filled.
	AssetsFilled sdk.Coin
	// PriceFilled is the amount of the order price that was filled.
	PriceFilled sdk.Coin
}

// NewPartialFulfillment creates a new PartialFulfillment using the provided OrderFulfillment information.
func NewPartialFulfillment(f *OrderFulfillment) *PartialFulfillment {
	rv := &PartialFulfillment{
		NewOrder:     NewOrder(f.GetOrderID()),
		AssetsFilled: f.GetAssetsFilled(),
		PriceFilled:  f.GetPriceFilled(),
	}

	if f.IsAskOrder() {
		askOrder := &AskOrder{
			MarketId:     f.GetMarketID(),
			Seller:       f.GetOwner(),
			Assets:       f.GetAssetsUnfilled(),
			Price:        f.GetPriceUnfilled(),
			AllowPartial: f.PartialFillAllowed(),
		}
		if !f.OrderFeesLeft.IsZero() {
			if len(f.OrderFeesLeft) > 1 {
				panic(fmt.Errorf("partially filled ask order %d somehow has multiple denoms in fees left %q",
					f.GetOrderID(), f.OrderFeesLeft))
			}
			askOrder.SellerSettlementFlatFee = &f.OrderFeesLeft[0]
		}
		rv.NewOrder.WithAsk(askOrder)
		return rv
	}

	if f.IsBidOrder() {
		bidOrder := &BidOrder{
			MarketId:            f.GetMarketID(),
			Buyer:               f.GetOwner(),
			Assets:              f.GetAssetsUnfilled(),
			Price:               f.GetPriceUnfilled(),
			BuyerSettlementFees: f.OrderFeesLeft,
			AllowPartial:        f.PartialFillAllowed(),
		}
		rv.NewOrder.WithBid(bidOrder)
		return rv
	}

	// This is here in case another order type is created, but a case for it isn't added to this func.
	panic(fmt.Errorf("order %d has unknown type %q", f.GetOrderID(), f.GetOrderType()))
}

// BuildFulfillments creates all of the ask and bid order fulfillments.
func BuildFulfillments(askOrders, bidOrders []*Order, sellerFeeRatio *FeeRatio) (*Fulfillments, error) {
	askOFs := make([]*OrderFulfillment, len(askOrders))
	for i, askOrder := range askOrders {
		askOFs[i] = NewOrderFulfillment(askOrder)
	}
	bidOFs := make([]*OrderFulfillment, len(bidOrders))
	for i, bidOrder := range bidOrders {
		bidOFs[i] = NewOrderFulfillment(bidOrder)
	}

	var a, b int
	for a < len(askOFs) && b < len(bidOFs) {
		err := Fulfill(askOFs[a], bidOFs[b])
		if err != nil {
			return nil, err
		}
		askFilled := askOFs[a].IsFullyFilled()
		bidFilled := bidOFs[b].IsFullyFilled()
		if !askFilled && !bidFilled {
			return nil, fmt.Errorf("neither ask order %d nor bid order %d could be filled in full",
				askOFs[a].GetOrderID(), bidOFs[b].GetOrderID())
		}
		if askFilled {
			a++
		}
		if bidFilled {
			b++
		}
	}

	// Finalize all the fulfillments.
	// Need to finalize bid orders first due to possible extra price distribution.
	for _, bidOF := range bidOFs {
		if err := bidOF.Finalize(sellerFeeRatio); err != nil {
			return nil, err
		}
	}
	for _, askOF := range askOFs {
		if err := askOF.Finalize(sellerFeeRatio); err != nil {
			return nil, err
		}
	}

	// And make sure they're all valid.
	for _, askOF := range askOFs {
		if err := askOF.Validate(); err != nil {
			return nil, err
		}
	}
	for _, bidOF := range bidOFs {
		if err := bidOF.Validate(); err != nil {
			return nil, err
		}
	}

	// Make sure none of them are partially filled except possibly the last in each list.
	var partialFulfillments []*OrderFulfillment
	lastAskI, lastBidI := len(askOFs)-1, len(bidOFs)-1
	for i, askOF := range askOFs {
		if !askOF.IsFullyFilled() {
			if i != lastAskI {
				return nil, fmt.Errorf("ask order %d (at index %d) is not filled in full and is not the last ask order provided",
					askOF.GetOrderID(), i)
			}
			partialFulfillments = append(partialFulfillments, askOF)
		}
	}
	for i, bidOF := range bidOFs {
		if !bidOF.IsFullyFilled() {
			if i != lastBidI {
				return nil, fmt.Errorf("bid order %d (at index %d) is not filled in full and is not the last bid order provided",
					bidOF.GetOrderID(), i)
			}
			partialFulfillments = append(partialFulfillments, bidOF)
		}
	}

	// And make sure that only one order is being partially filled.
	if len(partialFulfillments) > 1 {
		return nil, fmt.Errorf("%s order %d and %s order %d cannot both be partially filled",
			partialFulfillments[0].GetOrderType(), partialFulfillments[0].GetOrderID(),
			partialFulfillments[1].GetOrderType(), partialFulfillments[1].GetOrderID())
	}

	rv := &Fulfillments{
		AskOFs: askOFs,
		BidOFs: bidOFs,
	}

	if len(partialFulfillments) > 0 {
		rv.PartialOrder = NewPartialFulfillment(partialFulfillments[0])
	}

	return rv, nil
}

// indexedAddrAmts is a set of addresses and amounts.
type indexedAddrAmts struct {
	// addrs are a list of all addresses that have amounts.
	addrs []string
	// amts are a list of the coin amounts for each address (by slice index).
	amts []sdk.Coins
	// indexes are the index value for each address.
	indexes map[string]int
}

func newIndexedAddrAmts() *indexedAddrAmts {
	return &indexedAddrAmts{
		indexes: make(map[string]int),
	}
}

// add adds the coins to the given address.
// Panics if a provided coin is invalid.
func (i *indexedAddrAmts) add(addr string, coins ...sdk.Coin) {
	for _, coin := range coins {
		if err := coin.Validate(); err != nil {
			panic(fmt.Errorf("cannot index and add invalid coin amount %q", coin))
		}
	}
	n, known := i.indexes[addr]
	if !known {
		n = len(i.addrs)
		i.indexes[addr] = n
		i.addrs = append(i.addrs, addr)
		i.amts = append(i.amts, sdk.NewCoins())
	}
	i.amts[n] = i.amts[n].Add(coins...)
}

// getAsInputs returns all the entries as bank Inputs.
// Panics if this is nil, has no addrs, or has a negative coin amount.
func (i *indexedAddrAmts) getAsInputs() []banktypes.Input {
	if i == nil || len(i.addrs) == 0 {
		return nil
	}
	rv := make([]banktypes.Input, len(i.addrs))
	for n, addr := range i.addrs {
		if !i.amts[n].IsAllPositive() {
			panic(fmt.Errorf("invalid indexed amount %q for address %q: cannot be zero or negative", addr, i.amts[n]))
		}
		rv[n] = banktypes.Input{Address: addr, Coins: i.amts[n]}
	}
	return rv
}

// getAsOutputs returns all the entries as bank Outputs.
// Panics if this is nil, has no addrs, or has a negative coin amount.
func (i *indexedAddrAmts) getAsOutputs() []banktypes.Output {
	if i == nil || len(i.addrs) == 0 {
		return nil
	}
	rv := make([]banktypes.Output, len(i.addrs))
	for n, addr := range i.addrs {
		if !i.amts[n].IsAllPositive() {
			panic(fmt.Errorf("invalid indexed amount %q for address %q: cannot be zero or negative", addr, i.amts[n]))
		}
		rv[n] = banktypes.Output{Address: addr, Coins: i.amts[n]}
	}
	return rv
}

// Transfer contains bank inputs and outputs indicating a transfer that needs to be made.
type Transfer struct {
	// Inputs are the inputs that make up this transfer.
	Inputs []banktypes.Input
	// Outputs are the outputs that make up this transfer.
	Outputs []banktypes.Output
}

// SettlementTransfers has everything needed to do all the transfers for a settlement.
type SettlementTransfers struct {
	// OrderTransfers are all of the asset and price transfers needed to facilitate a settlement.
	OrderTransfers []*Transfer
	// FeeInputs are all of the inputs needed to facilitate payment of fees to a market.
	FeeInputs []banktypes.Input
}

// BuildSettlementTransfers creates all the order transfers needed for the provided fulfillments.
// Assumes that all fulfillments have passed Validate.
// Panics if any amounts are negative.
func BuildSettlementTransfers(f *Fulfillments) *SettlementTransfers {
	allOFs := make([]*OrderFulfillment, 0, len(f.AskOFs)+len(f.BidOFs))
	allOFs = append(allOFs, f.AskOFs...)
	allOFs = append(allOFs, f.BidOFs...)

	indexedFees := newIndexedAddrAmts()
	rv := &SettlementTransfers{
		OrderTransfers: make([]*Transfer, 0, len(allOFs)*2),
	}

	for _, of := range allOFs {
		rv.OrderTransfers = append(rv.OrderTransfers, GetAssetTransfer(of), GetPriceTransfer(of))
		if !of.FeesToPay.IsZero() {
			// Using NewCoins in here as a last-ditch negative amount panic check.
			fees := sdk.NewCoins(of.FeesToPay...)
			indexedFees.add(of.GetOwner(), fees...)
		}
	}

	rv.FeeInputs = indexedFees.getAsInputs()

	return rv
}

// GetAssetTransfer gets the inputs and outputs to facilitate the transfers of assets for this order fulfillment.
// Assumes that the fulfillment has passed Validate already.
// Panics if any amounts are negative or if it's neither a bid nor ask order.
func GetAssetTransfer(f *OrderFulfillment) *Transfer {
	indexedSplits := newIndexedAddrAmts()
	for _, split := range f.Splits {
		indexedSplits.add(split.Order.GetOwner(), split.Assets)
	}

	// Using NewCoins in here (instead of Coins{...}) as a last-ditch negative amount panic check.
	if f.IsAskOrder() {
		return &Transfer{
			Inputs:  []banktypes.Input{{Address: f.GetOwner(), Coins: sdk.NewCoins(f.GetAssetsFilled())}},
			Outputs: indexedSplits.getAsOutputs(),
		}
	}
	if f.IsBidOrder() {
		return &Transfer{
			Inputs:  indexedSplits.getAsInputs(),
			Outputs: []banktypes.Output{{Address: f.GetOwner(), Coins: sdk.NewCoins(f.GetAssetsFilled())}},
		}
	}

	// panicking in here if there's an error since it really should have happened earlier anyway.
	panic(fmt.Errorf("unknown order type %T", f.Order.GetOrder()))
}

// GetPriceTransfer gets the inputs and outputs to facilitate the transfers for the price of this order fulfillment.
// Assumes that the fulfillment has passed Validate already.
// Panics if any amounts are negative or if it's neither a bid nor ask order.
func GetPriceTransfer(f *OrderFulfillment) *Transfer {
	indexedSplits := newIndexedAddrAmts()
	for _, split := range f.Splits {
		indexedSplits.add(split.Order.GetOwner(), split.Price)
	}

	// Using NewCoins in here (instead of Coins{...}) as a last-ditch negative amount panic check.
	if f.IsAskOrder() {
		return &Transfer{
			Inputs:  indexedSplits.getAsInputs(),
			Outputs: []banktypes.Output{{Address: f.GetOwner(), Coins: sdk.NewCoins(f.GetPriceApplied())}},
		}
	}
	if f.IsBidOrder() {
		return &Transfer{
			Inputs:  []banktypes.Input{{Address: f.GetOwner(), Coins: sdk.NewCoins(f.GetPriceApplied())}},
			Outputs: indexedSplits.getAsOutputs(),
		}
	}

	// panicking in here if there's an error since it really should have happened earlier anyway.
	panic(fmt.Errorf("unknown order type %T", f.Order.GetOrder()))
}
