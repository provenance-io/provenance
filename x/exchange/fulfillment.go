package exchange

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// OrderSplit contains an order, and the asset and price amounts that should come out of it.
// TODO[1658]: Remove this struct.
type OrderSplit struct {
	// Order fulfillment associated with this split.
	Order *OrderFulfillment
	// Assets is the amount of assets from the order involved in this split.
	Assets sdk.Coin
	// Price is the amount of the price from the order involved in this split.
	Price sdk.Coin
}

// Distribution indicates an address and an amount that will either go to, or come from that address.
type Distribution struct {
	// Address is an bech32 address string
	Address string
	// Amount is the amount that will either go to, or come from that address.
	Amount sdkmath.Int
}

// OrderFulfillment is used to figure out how an order should be fulfilled.
type OrderFulfillment struct {
	// Order is the original order with all its information.
	Order *Order

	// Splits contains information on the orders being used to fulfill this order.
	Splits []*OrderSplit // TODO[1658]: Remove this field.

	// AssetDists contains distribution info for this order's assets.
	AssetDists []*Distribution
	// AssetDists contains distribution info for this order's price.
	PriceDists []*Distribution

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

// NewOrderFulfillments creates a new OrderFulfillment for each of the provided orders.
func NewOrderFulfillments(orders []*Order) []*OrderFulfillment {
	rv := make([]*OrderFulfillment, len(orders))
	for i, o := range orders {
		rv[i] = NewOrderFulfillment(o)
	}
	return rv
}

// assetCoin returns a coin with the given amount and the same denom as this order's assets.
func (f OrderFulfillment) assetCoin(amt sdkmath.Int) sdk.Coin {
	return sdk.Coin{Denom: f.GetAssets().Denom, Amount: amt}
}

// priceCoin returns a coin with the given amount and the same denom as this order's price.
func (f OrderFulfillment) priceCoin(amt sdkmath.Int) sdk.Coin {
	return sdk.Coin{Denom: f.GetPrice().Denom, Amount: amt}
}

// GetAssetsFilled gets the coin value of the assets that have been filled in this fulfillment.
func (f OrderFulfillment) GetAssetsFilled() sdk.Coin {
	return f.assetCoin(f.AssetsFilledAmt)
}

// GetAssetsUnfilled gets the coin value of the assets left to fill in this fulfillment.
func (f OrderFulfillment) GetAssetsUnfilled() sdk.Coin {
	return f.assetCoin(f.AssetsUnfilledAmt)
}

// GetPriceApplied gets the coin value of the price that has been filled in this fulfillment.
func (f OrderFulfillment) GetPriceApplied() sdk.Coin {
	return f.priceCoin(f.PriceAppliedAmt)
}

// GetPriceLeft gets the coin value of the price left to fill in this fulfillment.
func (f OrderFulfillment) GetPriceLeft() sdk.Coin {
	return f.priceCoin(f.PriceLeftAmt)
}

// GetPriceFilled gets the coin value of the price filled in this fulfillment.
func (f OrderFulfillment) GetPriceFilled() sdk.Coin {
	return f.priceCoin(f.PriceFilledAmt)
}

// GetPriceUnfilled gets the coin value of the price unfilled in this fulfillment.
func (f OrderFulfillment) GetPriceUnfilled() sdk.Coin {
	return f.priceCoin(f.PriceUnfilledAmt)
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

// DistributeAssets records the distribution of assets in the provided amount to/from the given order.
func (f *OrderFulfillment) DistributeAssets(order OrderI, amount sdkmath.Int) error {
	if f.AssetsUnfilledAmt.LT(amount) {
		return fmt.Errorf("cannot fill %s order %d having assets left %q with %q from %s order %d: overfill",
			f.GetOrderType(), f.GetOrderID(), f.GetAssetsUnfilled(), f.assetCoin(amount), order.GetOrderType(), order.GetOrderID())
	}

	f.AssetsUnfilledAmt = f.AssetsUnfilledAmt.Sub(amount)
	f.AssetsFilledAmt = f.AssetsFilledAmt.Add(amount)
	f.AssetDists = append(f.AssetDists, &Distribution{
		Address: order.GetOwner(),
		Amount:  amount,
	})
	return nil
}

// DistributeAssets records the distribution of assets in the provided amount to/from the given fulfillments.
func DistributeAssets(of1, of2 *OrderFulfillment, amount sdkmath.Int) error {
	errs := []error{
		of1.DistributeAssets(of2, amount),
		of2.DistributeAssets(of1, amount),
	}
	return errors.Join(errs...)
}

// DistributePrice records the distribution of price in the provided amount to/from the given order.
func (f *OrderFulfillment) DistributePrice(order OrderI, amount sdkmath.Int) error {
	if f.PriceLeftAmt.LT(amount) && f.IsBidOrder() {
		return fmt.Errorf("cannot fill %s order %d having price left %q to %s order %d at a price of %q: overfill",
			f.GetOrderType(), f.GetOrderID(), f.GetPriceLeft(), order.GetOrderType(), order.GetOrderID(), f.priceCoin(amount))
	}

	f.PriceLeftAmt = f.PriceLeftAmt.Sub(amount)
	f.PriceAppliedAmt = f.PriceAppliedAmt.Add(amount)
	f.PriceDists = append(f.PriceDists, &Distribution{
		Address: order.GetOwner(),
		Amount:  amount,
	})
	return nil
}

// DistributePrice records the distribution of price in the provided amount to/from the given fulfillments.
func DistributePrice(of1, of2 *OrderFulfillment, amount sdkmath.Int) error {
	errs := []error{
		of1.DistributePrice(of2, amount),
		of2.DistributePrice(of1, amount),
	}
	return errors.Join(errs...)
}

// SplitOrder splits this order on the amount of assets filled.
// This order fulfillment is updated to have the filled order, and the unfilled portion is returned.
func (f *OrderFulfillment) SplitOrder() (*Order, error) {
	filled, unfilled, err := f.Order.Split(f.AssetsFilledAmt)
	if err != nil {
		return nil, err
	}

	f.Order = filled
	f.AssetsUnfilledAmt = sdkmath.ZeroInt()
	f.PriceLeftAmt = filled.GetPrice().Amount.Sub(f.PriceAppliedAmt)
	return unfilled, nil
}

// AsFilledOrder creates a FilledOrder from this order fulfillment.
func (f OrderFulfillment) AsFilledOrder() *FilledOrder {
	return NewFilledOrder(f.Order, f.GetPriceApplied(), f.FeesToPay)
}

// sumAssetsAndPrice gets the sum of assets, and the sum of prices of the provided orders.
func sumAssetsAndPrice(orders []*Order) (assets sdk.Coins, price sdk.Coins) {
	for _, o := range orders {
		assets = assets.Add(o.GetAssets())
		price = price.Add(o.GetPrice())
	}
	return
}

// sumPriceLeft gets the sum of price left of the provided fulfillments.
func sumPriceLeft(fulfillments []*OrderFulfillment) sdkmath.Int {
	rv := sdkmath.ZeroInt()
	for _, f := range fulfillments {
		rv = rv.Add(f.PriceLeftAmt)
	}
	return rv
}

// Settlement contains information on how a set of orders is to be settled.
type Settlement struct {
	// Transfers are all of the inputs and outputs needed to facilitate movement of assets and price.
	Transfers []*Transfer
	// FeeInputs are the inputs needed to facilitate payment of order fees.
	FeeInputs []banktypes.Input
	// FullyFilledOrders are all the orders that are fully filled in this settlement.
	// If there is an order that's being partially filled, it will not be included in this list.
	FullyFilledOrders []*FilledOrder
	// PartialOrderFilled is a partially filled order with amounts indicating how much was filled.
	// This is not included in FullyFilledOrders.
	PartialOrderFilled *FilledOrder
	// PartialOrderLeft is what's left of the partially filled order.
	PartialOrderLeft *Order
}

// BuildSettlement processes the provided orders, identifying how the provided orders can be settled.
func BuildSettlement(askOrders, bidOrders []*Order, sellerFeeRatioLookup func(denom string) (*FeeRatio, error)) (*Settlement, error) {
	if err := validateCanSettle(askOrders, bidOrders); err != nil {
		return nil, err
	}

	askOFs := NewOrderFulfillments(askOrders)
	bidOFs := NewOrderFulfillments(bidOrders)

	// Allocate the assets first.
	if err := allocateAssets(askOFs, bidOFs); err != nil {
		return nil, err
	}

	settlement := &Settlement{}
	// Identify any partial order and update its entry in the order fulfillments.
	if err := splitPartial(askOFs, bidOFs, settlement); err != nil {
		return nil, err
	}

	// Allocate the prices.
	if err := allocatePrice(askOFs, bidOFs); err != nil {
		return nil, err
	}

	// Set the fees in the fulfillments
	sellerFeeRatio, err := sellerFeeRatioLookup(askOFs[0].GetPrice().Denom)
	if err != nil {
		return nil, err
	}
	if err = setFeesToPay(askOFs, bidOFs, sellerFeeRatio); err != nil {
		return nil, err
	}

	// Make sure everything adds up.
	if err = validateFulfillments(askOFs, bidOFs); err != nil {
		return nil, err
	}

	// Create the transfers
	if err = buildTransfers(askOFs, bidOFs, settlement); err != nil {
		return nil, err
	}

	// Indicate what's been filled in full and partially.
	populateFilled(askOFs, bidOFs, settlement)

	return settlement, nil
}

// validateCanSettle does some superficial checking of the provided orders to make sure we can try to settle them.
func validateCanSettle(askOrders, bidOrders []*Order) error {
	var errs []error
	if len(askOrders) == 0 {
		errs = append(errs, errors.New("no ask orders provided"))
	}
	if len(bidOrders) == 0 {
		errs = append(errs, errors.New("no bid orders provided"))
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	for i, askOrder := range askOrders {
		if !askOrder.IsAskOrder() {
			errs = append(errs, fmt.Errorf("%s order %d is not an ask order but is in the askOrders list at index %d",
				askOrder.GetOrderType(), askOrder.GetOrderID(), i))
		}
	}
	for i, bidOrder := range bidOrders {
		if !bidOrder.IsBidOrder() {
			errs = append(errs, fmt.Errorf("%s order %d is not a bid order but is in the bidOrders list at index %d",
				bidOrder.GetOrderType(), bidOrder.GetOrderID(), i))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	totalAssetsForSale, totalAskPrice := sumAssetsAndPrice(askOrders)
	totalAssetsToBuy, totalBidPrice := sumAssetsAndPrice(bidOrders)

	if len(totalAssetsForSale) != 1 {
		errs = append(errs, fmt.Errorf("cannot settle with multiple ask order asset denoms %q", totalAssetsForSale))
	}
	if len(totalAskPrice) != 1 {
		errs = append(errs, fmt.Errorf("cannot settle with multiple ask order price denoms %q", totalAskPrice))
	}
	if len(totalAssetsToBuy) != 1 {
		errs = append(errs, fmt.Errorf("cannot settle with multiple bid order asset denoms %q", totalAssetsToBuy))
	}
	if len(totalBidPrice) != 1 {
		errs = append(errs, fmt.Errorf("cannot settle with multiple bid order price denoms %q", totalBidPrice))
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	if totalAssetsForSale[0].Denom != totalAssetsToBuy[0].Denom {
		errs = append(errs, fmt.Errorf("cannot settle different ask %q and bid %q asset denoms",
			totalAssetsForSale, totalAssetsToBuy))
	}
	if totalAskPrice[0].Denom != totalBidPrice[0].Denom {
		errs = append(errs, fmt.Errorf("cannot settle different ask %q and bid %q price denoms",
			totalAskPrice, totalBidPrice))
	}

	// Note: We don't compare the total asset and price amounts because we need to know what's
	// being partially filled before we can make an assertions about the amounts.

	return errors.Join(errs...)
}

// allocateAssets distributes the assets among the fulfillments.
func allocateAssets(askOFs, bidOFs []*OrderFulfillment) error {
	a, b := 0, 0
	for a < len(askOFs) && b < len(bidOFs) {
		assetsFilledAmt, err := GetFulfillmentAssetsAmt(askOFs[a], bidOFs[b])
		if err != nil {
			return err
		}
		if err = DistributeAssets(askOFs[a], bidOFs[b], assetsFilledAmt); err != nil {
			return err
		}

		askFilled := askOFs[a].AssetsUnfilledAmt.IsZero()
		bidFilled := bidOFs[b].AssetsUnfilledAmt.IsZero()
		if !askFilled && !bidFilled {
			return fmt.Errorf("neither %s order %d nor %s order %d could have assets filled in full",
				askOFs[a].GetOrderType(), askOFs[a].GetOrderID(), bidOFs[b].GetOrderType(), bidOFs[b].GetOrderID())
		}
		if askFilled {
			a++
		}
		if bidFilled {
			b++
		}
	}

	return nil
}

// splitPartial checks the provided fulfillments for a partial order and splits it out, updating the applicable fulfillment.
// This will possibly populate the PartialOrderLeft in the provided Settlement.
func splitPartial(askOFs, bidOFs []*OrderFulfillment, settlement *Settlement) error {
	if err := splitOrderFulfillments(askOFs, settlement); err != nil {
		return err
	}
	return splitOrderFulfillments(bidOFs, settlement)
}

// splitOrderFulfillments checks each of the OrderFulfillment for partial (or incomplete) fills.
// If an appropriate partial fill is found, its OrderFulfillment is update and Settlement.PartialOrderLeft is set.
func splitOrderFulfillments(fulfillments []*OrderFulfillment, settlement *Settlement) error {
	lastI := len(fulfillments) - 1
	for i, f := range fulfillments {
		if f.AssetsFilledAmt.IsZero() {
			return fmt.Errorf("%s order %d (at index %d) has no assets filled",
				f.GetOrderType(), f.GetOrderID(), i)
		}
		if !f.AssetsUnfilledAmt.IsZero() {
			if i != lastI {
				return fmt.Errorf("%s order %d (at index %d) is not filled in full and is not the last %s order provided",
					f.GetOrderType(), f.GetOrderID(), i, f.GetOrderType())
			}
			if settlement.PartialOrderLeft != nil {
				return fmt.Errorf("%s order %d and %s order %d cannot both be partially filled",
					settlement.PartialOrderLeft.GetOrderType(), settlement.PartialOrderLeft.GetOrderID(),
					f.GetOrderType(), f.GetOrderID())
			}
			var err error
			settlement.PartialOrderLeft, err = f.SplitOrder()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// allocatePrice distributes the prices among the fulfillments.
func allocatePrice(askOFs, bidOFs []*OrderFulfillment) error {
	// Check that the total ask price is not more than the total bid price.
	totalAskPriceAmt := sumPriceLeft(askOFs)
	totalBidPriceAmt := sumPriceLeft(bidOFs)
	if totalAskPriceAmt.GT(totalBidPriceAmt) {
		return fmt.Errorf("total ask price %q is greater than total bid price %q",
			askOFs[0].priceCoin(totalAskPriceAmt), bidOFs[0].priceCoin(totalBidPriceAmt))
	}

	// First pass at price distribution: Give all the asks their price.
	b := 0
	totalFilledFirstPass := sdkmath.ZeroInt()
	for _, askOF := range askOFs {
		for askOF.PriceLeftAmt.IsPositive() && bidOFs[b].PriceLeftAmt.IsPositive() {
			priceFilledAmt, err := GetFulfillmentPriceAmt(askOF, bidOFs[b])
			if err != nil {
				return err
			}
			if err = DistributePrice(askOF, bidOFs[b], priceFilledAmt); err != nil {
				return err
			}
			totalFilledFirstPass = totalFilledFirstPass.Add(priceFilledAmt)
			if !bidOFs[b].PriceLeftAmt.IsPositive() {
				b++
			}
		}
	}

	// Above, we made sure that ask price <= bid price.
	// So here, we can assume that total price filled = ask price.
	// If it's also the total bid price, we're done!
	if totalFilledFirstPass.Equal(totalBidPriceAmt) {
		return nil
	}

	// We also know that total price filled <= bid price, so bid price - total price left will be positive.
	totalLeftoverPriceAmt := totalBidPriceAmt.Sub(totalFilledFirstPass)
	// We need an assets total so we can allocate the leftovers by assets amounts.
	totalAssetsAmt := sdkmath.ZeroInt()
	for _, askOF := range askOFs {
		totalAssetsAmt = totalAssetsAmt.Add(askOF.AssetsFilledAmt)
	}

	// Now, distribute the leftovers per asset (truncated).
	// Start over on the asks, but start where we left off on the bids (since that's the first with any price left).
	a := -1
	firstPass := true
	leftoverPriceAmt := totalLeftoverPriceAmt
	for !leftoverPriceAmt.IsZero() {
		// Move on to the next ask, looping back to the first if we need to.
		a++
		if a == len(askOFs) {
			a = 0
			firstPass = false
		}

		// If there's no bids left to get this from, something's logically wrong with this process.
		if b >= len(bidOFs) {
			panic(fmt.Errorf("total ask price %q, total bid price %q, difference %q, left to allocate %q: "+
				"no bid orders left to allocate leftovers from",
				askOFs[0].priceCoin(totalAskPriceAmt), bidOFs[0].priceCoin(totalBidPriceAmt),
				askOFs[0].priceCoin(totalLeftoverPriceAmt), askOFs[0].priceCoin(leftoverPriceAmt)))
		}

		// Figure out how much additional price this order should get.
		addPriceAmt := totalLeftoverPriceAmt.Mul(askOFs[a].AssetsFilledAmt).Quo(totalAssetsAmt)
		if addPriceAmt.IsZero() {
			if firstPass {
				continue
			}
			// If this isn't the first time through the asks, distribute at least one price to each ask.
			addPriceAmt = sdkmath.OneInt()
		}
		// We can only add what's actually left over.
		addPriceAmt = MinSDKInt(addPriceAmt, leftoverPriceAmt)

		// If it can't all come out of the current bid, use the rest of what this bid has and move to the next bid.
		for !addPriceAmt.IsZero() && b < len(bidOFs) && bidOFs[b].PriceLeftAmt.LTE(addPriceAmt) {
			bidPriceLeft := bidOFs[b].PriceLeftAmt
			if err := DistributePrice(askOFs[a], bidOFs[b], bidPriceLeft); err != nil {
				return err
			}
			addPriceAmt = addPriceAmt.Sub(bidPriceLeft)
			leftoverPriceAmt = leftoverPriceAmt.Sub(bidPriceLeft)
			b++
		}

		// If there's still additional price left, it can all come out of the current bid.
		if !addPriceAmt.IsZero() && b < len(bidOFs) {
			if err := DistributePrice(askOFs[a], bidOFs[b], addPriceAmt); err != nil {
				return err
			}
			leftoverPriceAmt = leftoverPriceAmt.Sub(addPriceAmt)
			// If that was all of it, move on to the next.
			if bidOFs[b].PriceLeftAmt.IsZero() {
				b++
			}
		}
	}

	return nil
}

// setFeesToPay sets the FeesToPay on each fulfillment.
func setFeesToPay(askOFs, bidOFs []*OrderFulfillment, sellerFeeRatio *FeeRatio) error {
	var errs []error
	for _, askOF := range askOFs {
		feesToPay := askOF.GetSettlementFees()
		if sellerFeeRatio != nil {
			fee, err := sellerFeeRatio.ApplyToLoosely(askOF.GetPriceApplied())
			if err != nil {
				errs = append(errs, fmt.Errorf("failed calculate ratio fee for %s order %d: %w",
					askOF.GetOrderType(), askOF.GetOrderID(), err))
				continue
			}
			feesToPay = feesToPay.Add(fee)
		}
		askOF.FeesToPay = feesToPay
	}

	for _, bidOF := range bidOFs {
		bidOF.FeesToPay = bidOF.GetSettlementFees()
	}

	return errors.Join(errs...)
}

// validateFulfillments runs .Validate on each fulfillment, returning any problems.
func validateFulfillments(askOFs, bidOFs []*OrderFulfillment) error {
	var errs []error
	for _, askOF := range askOFs {
		if err := askOF.Validate(); err != nil {
			errs = append(errs, err)
		}
	}
	for _, bidOF := range bidOFs {
		if err := bidOF.Validate(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// buildTransfers creates the transfers, inputs for fee payments,
// and fee total and sets those fields in the provided Settlement.
// This will populate the Transfers and FeeInputs fields in the provided Settlement.
func buildTransfers(askOFs, bidOFs []*OrderFulfillment, settlement *Settlement) error {
	var errs []error
	indexedFees := newIndexedAddrAmts()
	transfers := make([]*Transfer, 0, len(askOFs)+len(bidOFs))

	record := func(of *OrderFulfillment, getter func(fulfillment *OrderFulfillment) (*Transfer, error)) {
		assetTrans, err := getter(of)
		if err != nil {
			errs = append(errs, err)
		} else {
			transfers = append(transfers, assetTrans)
		}

		if !of.FeesToPay.IsZero() {
			if of.FeesToPay.IsAnyNegative() {
				errs = append(errs, fmt.Errorf("%s order %d cannot pay %q in fees: negative amount",
					of.GetOrderType(), of.GetOrderID(), of.FeesToPay))
			} else {
				indexedFees.add(of.GetOwner(), of.FeesToPay...)
			}
		}
	}

	// If we got both the asset and price transfers from all OrderFulfillments, we'd be doubling
	// up on what's being traded. So we need to only get each from only one list.
	// Since we need to loop through both lists to make note of the fees, though, I thought it
	// would be good to get one from one list, and the other from the other. So I decided to get
	// the asset transfers from the asks, and price transfers from the bids so that each transfer
	// will have one input and multiple (or one) outputs. I'm not sure that it matters too much
	// where we get each transfer type from, though.

	for _, of := range askOFs {
		record(of, getAssetTransfer)
	}

	for _, of := range bidOFs {
		record(of, getPriceTransfer)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	settlement.Transfers = transfers
	settlement.FeeInputs = indexedFees.getAsInputs()

	return nil
}

// populateFilled creates all the FilledOrder entries and stores them in the provided Settlement.
// This will populate the FullyFilledOrders and PartialOrderFilled fields in the provided Settlement.
func populateFilled(askOFs, bidOFs []*OrderFulfillment, settlement *Settlement) {
	settlement.FullyFilledOrders = make([]*FilledOrder, 0, len(askOFs)+len(bidOFs))

	for _, f := range askOFs {
		if settlement.PartialOrderLeft != nil && settlement.PartialOrderLeft.GetOrderID() == f.GetOrderID() {
			settlement.PartialOrderFilled = f.AsFilledOrder()
		} else {
			settlement.FullyFilledOrders = append(settlement.FullyFilledOrders, f.AsFilledOrder())
		}
	}

	for _, f := range bidOFs {
		if settlement.PartialOrderLeft != nil && settlement.PartialOrderLeft.GetOrderID() == f.GetOrderID() {
			settlement.PartialOrderFilled = f.AsFilledOrder()
		} else {
			settlement.FullyFilledOrders = append(settlement.FullyFilledOrders, f.AsFilledOrder())
		}
	}
}

// getAssetTransfer gets the inputs and outputs to facilitate the transfers of assets for this order fulfillment.
func getAssetTransfer(f *OrderFulfillment) (*Transfer, error) {
	assetsFilled := f.GetAssetsFilled()
	if !assetsFilled.Amount.IsPositive() {
		return nil, fmt.Errorf("%s order %d cannot be filled with %q assets: amount not positive",
			f.GetOrderType(), f.GetOrderID(), assetsFilled)
	}

	indexedDists := newIndexedAddrAmts()
	sumDists := sdkmath.ZeroInt()
	for _, dist := range f.AssetDists {
		if !dist.Amount.IsPositive() {
			return nil, fmt.Errorf("%s order %d cannot have %q assets in a transfer: amount not positive",
				f.GetOrderType(), f.GetOrderID(), f.assetCoin(dist.Amount))
		}
		indexedDists.add(dist.Address, f.assetCoin(dist.Amount))
		sumDists = sumDists.Add(dist.Amount)
	}

	if !sumDists.Equal(assetsFilled.Amount) {
		return nil, fmt.Errorf("%s order %d assets filled %q does not equal assets distributed %q",
			f.GetOrderType(), f.GetOrderID(), assetsFilled, f.assetCoin(sumDists))
	}

	if f.IsAskOrder() {
		return &Transfer{
			Inputs:  []banktypes.Input{{Address: f.GetOwner(), Coins: sdk.NewCoins(assetsFilled)}},
			Outputs: indexedDists.getAsOutputs(),
		}, nil
	}
	if f.IsBidOrder() {
		return &Transfer{
			Inputs:  indexedDists.getAsInputs(),
			Outputs: []banktypes.Output{{Address: f.GetOwner(), Coins: sdk.NewCoins(assetsFilled)}},
		}, nil
	}

	// This is here in case a new SubTypeI is made that isn't accounted for in here.
	panic(fmt.Errorf("%s order %d: unknown order type", f.GetOrderType(), f.GetOrderID()))
}

// getPriceTransfer gets the inputs and outputs to facilitate the transfers for the price of this order fulfillment.
func getPriceTransfer(f *OrderFulfillment) (*Transfer, error) {
	priceApplied := f.GetPriceApplied()
	if !priceApplied.Amount.IsPositive() {
		return nil, fmt.Errorf("%s order %d cannot be filled at price %q: amount not positive",
			f.GetOrderType(), f.GetOrderID(), priceApplied)
	}

	indexedDists := newIndexedAddrAmts()
	sumDists := sdkmath.ZeroInt()
	for _, dist := range f.PriceDists {
		if !dist.Amount.IsPositive() {
			return nil, fmt.Errorf("%s order %d cannot have price %q in a transfer: amount not positive",
				f.GetOrderType(), f.GetOrderID(), f.priceCoin(dist.Amount))
		}
		indexedDists.add(dist.Address, f.priceCoin(dist.Amount))
		sumDists = sumDists.Add(dist.Amount)
	}

	if !sumDists.Equal(priceApplied.Amount) {
		return nil, fmt.Errorf("%s order %d price filled %q does not equal price distributed %q",
			f.GetOrderType(), f.GetOrderID(), priceApplied, f.priceCoin(sumDists))
	}

	if f.IsAskOrder() {
		return &Transfer{
			Inputs:  indexedDists.getAsInputs(),
			Outputs: []banktypes.Output{{Address: f.GetOwner(), Coins: sdk.NewCoins(priceApplied)}},
		}, nil
	}
	if f.IsBidOrder() {
		return &Transfer{
			Inputs:  []banktypes.Input{{Address: f.GetOwner(), Coins: sdk.NewCoins(priceApplied)}},
			Outputs: indexedDists.getAsOutputs(),
		}, nil
	}

	// This is here in case a new SubTypeI is made that isn't accounted for in here.
	panic(fmt.Errorf("%s order %d: unknown order type", f.GetOrderType(), f.GetOrderID()))
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
	// Checking for assets filled > zero here (instead of in Validate2) because we need to divide by it in here.
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

// Validate makes sure the assets filled and price applied are acceptable for this fulfillment.
func (f OrderFulfillment) Validate() (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	orderPrice := f.GetPrice()
	switch {
	case f.IsAskOrder():
		if orderPrice.Amount.GT(f.PriceAppliedAmt) {
			return fmt.Errorf("%s order %d price %q is more than price filled %q",
				f.GetOrderType(), f.GetOrderID(), orderPrice, f.GetPriceApplied())
		}
	case f.IsBidOrder():
		if !orderPrice.Amount.Equal(f.PriceAppliedAmt) {
			return fmt.Errorf("%s order %d price %q is not equal to price filled %q",
				f.GetOrderType(), f.GetOrderID(), orderPrice, f.GetPriceApplied())
		}
	default:
		// This is here in case something new implements SubOrderI but a case isn't added here.
		panic(fmt.Errorf("%s order %d: unknown order type", f.GetOrderType(), f.GetOrderID()))
	}

	orderAssets := f.GetAssets()
	if !orderAssets.Amount.Equal(f.AssetsFilledAmt) {
		return fmt.Errorf("%s order %d assets %q does not equal filled assets %q",
			f.GetOrderType(), f.GetOrderID(), orderAssets, f.GetAssetsFilled())
	}

	return nil
}

// Validate2 does some final validation and sanity checking on this order fulfillment.
// It's assumed that Finalize has been called before calling this.
func (f OrderFulfillment) Validate2() error {
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
	if f.PriceLeftAmt.GTE(orderPrice.Amount) {
		return fmt.Errorf("price left %q is not less than %s order %d price %q",
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
	if !f.PriceFilledAmt.IsPositive() {
		return fmt.Errorf("cannot fill %s order %d having price %q with non-positive price %q",
			f.GetOrderType(), f.GetOrderID(), orderPrice, f.GetPriceFilled())
	}
	totalPriceAmt := f.PriceFilledAmt.Add(f.PriceUnfilledAmt)
	if !orderPrice.Amount.Equal(totalPriceAmt) {
		return fmt.Errorf("filled price %q plus unfilled price %q does not equal order price %q for %s order %d",
			f.GetPriceFilled(), f.GetPriceUnfilled(), orderPrice, f.GetOrderType(), f.GetOrderID())
	}
	if f.PriceUnfilledAmt.IsNegative() {
		return fmt.Errorf("%s order %d having price %q has negative price %q after filling %q",
			f.GetOrderType(), f.GetOrderID(), orderPrice, f.GetPriceUnfilled(), f.GetPriceFilled())
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
		return fmt.Errorf("multiple asset denoms %q in splits applied to %s order %d having assets %q",
			splitsAssets, f.GetOrderType(), f.GetOrderID(), orderAssets)
	}
	if splitsAssets[0].Denom != orderAssets.Denom {
		return fmt.Errorf("splits asset denom %q does not equal order assets denom %q on %s order %d",
			splitsAssets, orderAssets, f.GetOrderType(), f.GetOrderID())
	}
	if !splitsAssets[0].Amount.Equal(f.AssetsFilledAmt) {
		return fmt.Errorf("splits asset total %q does not equal filled assets %q on %s order %d",
			splitsAssets, f.GetAssetsFilled(), f.GetOrderType(), f.GetOrderID())
	}

	if len(splitsPrice) != 1 {
		return fmt.Errorf("multiple price denoms %q in splits applied to %s order %d having price %q",
			splitsPrice, f.GetOrderType(), f.GetOrderID(), orderPrice)
	}
	if splitsPrice[0].Denom != orderPrice.Denom {
		return fmt.Errorf("splits price denom %q does not equal order price denom %q on %s order %d",
			splitsPrice, orderPrice, f.GetOrderType(), f.GetOrderID())
	}
	if !splitsPrice[0].Amount.Equal(f.PriceAppliedAmt) {
		return fmt.Errorf("splits price total %q does not equal filled price %q on %s order %d",
			splitsPrice, f.GetPriceApplied(), f.GetOrderType(), f.GetOrderID())
	}

	orderFees := f.GetSettlementFees()
	if f.OrderFeesLeft.IsAnyNegative() {
		return fmt.Errorf("settlement fees left %q is negative for %s order %d having fees %q",
			f.OrderFeesLeft, f.GetOrderType(), f.GetOrderID(), orderFees)
	}
	if _, hasNeg := orderFees.SafeSub(f.OrderFeesLeft...); hasNeg {
		return fmt.Errorf("settlement fees left %q is greater than %s order %d settlement fees %q",
			f.OrderFeesLeft, f.GetOrderType(), f.GetOrderID(), orderFees)
	}

	isFullyFilled := f.IsFullyFilled()
	if isFullyFilled {
		// IsFullyFilled returns true if unfilled assets is zero or negative.
		// We know from a previous check that unfilled is not negative.
		// So here, we know it's zero, and don't need to check that again.

		if !f.PriceUnfilledAmt.IsZero() {
			return fmt.Errorf("fully filled %s order %d has non-zero unfilled price %q",
				f.GetOrderType(), f.GetOrderID(), f.GetPriceUnfilled())
		}
		if !f.OrderFeesLeft.IsZero() {
			return fmt.Errorf("fully filled %s order %d has non-zero settlement fees left %q",
				f.GetOrderType(), f.GetOrderID(), f.OrderFeesLeft)
		}
	}

	switch {
	case f.IsAskOrder():
		// For ask orders, the applied amount needs to be at least the filled amount.
		if f.PriceAppliedAmt.LT(f.PriceFilledAmt) {
			return fmt.Errorf("%s order %d having assets %q and price %q cannot be filled by %q at price %q: insufficient price",
				f.GetOrderType(), f.GetOrderID(), orderAssets, orderPrice, f.GetAssetsFilled(), f.GetPriceApplied())
		}

		// If not being fully filled on an order that has some fees, make sure that there's at most 1 denom in the fees left.
		if !isFullyFilled && len(orderFees) > 0 && len(f.OrderFeesLeft) > 1 {
			return fmt.Errorf("partial fulfillment for %s order %d having seller settlement fees %q has multiple denoms in fees left %q",
				f.GetOrderType(), f.GetOrderID(), orderFees, f.OrderFeesLeft)
		}

		// For ask orders, the tracked fees must be at least the order fees.
		trackedFees := f.FeesToPay.Add(f.OrderFeesLeft...)
		if _, hasNeg := trackedFees.SafeSub(orderFees...); hasNeg {
			return fmt.Errorf("tracked settlement fees %q is less than %s order %d settlement fees %q",
				trackedFees, f.GetOrderType(), f.GetOrderID(), orderFees)
		}
	case f.IsBidOrder():
		if !f.PriceAppliedAmt.Equal(f.PriceFilledAmt) {
			return fmt.Errorf("price applied %q does not equal price filled %q for %s order %d having price %q",
				f.GetPriceApplied(), f.GetPriceFilled(), f.GetOrderType(), f.GetOrderID(), orderPrice)
		}

		// We now know that price applied = filled, and applied + left = order = filled + unfilled, so left = unfilled too.
		// We also know that applied > 0 and left < order. So 0 < applied < order.
		// If fully filled, we know that unfilled = 0, so applied = order = filled, so we don't need to check that again here.
		// If partially filled, we know that applied < order, so we don't need to check that either.

		// For bid orders, fees to pay + fees left should equal the order fees.
		trackedFees := f.FeesToPay.Add(f.OrderFeesLeft...)
		if !CoinsEquals(trackedFees, orderFees) {
			return fmt.Errorf("tracked settlement fees %q does not equal %s order %d settlement fees %q",
				trackedFees, f.GetOrderType(), f.GetOrderID(), orderFees)
		}
	default:
		// The only way to trigger this would be to add a new order type but not add a case for it in this switch.
		panic(fmt.Errorf("case missing for %T in Validate2", f.GetOrderType()))
	}

	// Saving this simple check for last in the hopes that a previous error exposes why this
	// order might accidentally be only partially filled.
	if !isFullyFilled && !f.PartialFillAllowed() {
		return fmt.Errorf("cannot fill %s order %d having assets %q with %q: order does not allow partial fill",
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
		return fmt.Errorf("cannot fulfill %s order %d with %s order %d: order type mismatch",
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
	if askOrder.Assets.Denom != bidOrder.Assets.Denom {
		return fmt.Errorf("cannot fill bid order %d having assets %q with ask order %d having assets %q: denom mismatch",
			bidOF.GetOrderID(), bidOrder.Assets, askOF.GetOrderID(), askOrder.Assets)
	}
	if askOrder.Price.Denom != bidOrder.Price.Denom {
		return fmt.Errorf("cannot fill ask order %d having price %q with bid order %d having price %q: denom mismatch",
			askOF.GetOrderID(), askOrder.Price, bidOF.GetOrderID(), bidOrder.Price)
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

	return MinSDKInt(of1.AssetsUnfilledAmt, of2.AssetsUnfilledAmt), nil
}

// GetFulfillmentPriceAmt figures out the price that can be fulfilled with the two provided orders.
func GetFulfillmentPriceAmt(of1, of2 *OrderFulfillment) (sdkmath.Int, error) {
	if !of1.PriceLeftAmt.IsPositive() || !of2.PriceLeftAmt.IsPositive() {
		return sdkmath.ZeroInt(), fmt.Errorf("cannot fill %s order %d having price left %q "+
			"with %s order %d having price left %q: zero or negative price left",
			of1.GetOrderType(), of1.GetOrderID(), of1.GetPriceLeft(),
			of2.GetOrderType(), of2.GetOrderID(), of2.GetPriceLeft())
	}

	return MinSDKInt(of1.PriceLeftAmt, of2.PriceLeftAmt), nil
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
		if !askOrder.IsAskOrder() {
			return nil, fmt.Errorf("%s order %d is not an ask order but is in the askOrders list",
				askOrder.GetOrderType(), askOrder.GetOrderID())
		}
		askOFs[i] = NewOrderFulfillment(askOrder)
	}
	bidOFs := make([]*OrderFulfillment, len(bidOrders))
	for i, bidOrder := range bidOrders {
		if !bidOrder.IsBidOrder() {
			return nil, fmt.Errorf("%s order %d is not a bid order but is in the bidOrders list",
				bidOrder.GetOrderType(), bidOrder.GetOrderID())
		}
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
			return nil, fmt.Errorf("neither %s order %d nor %s order %d could be filled in full",
				askOFs[a].GetOrderType(), askOFs[a].GetOrderID(), bidOFs[b].GetOrderType(), bidOFs[b].GetOrderID())
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
		if err := askOF.Validate2(); err != nil {
			return nil, err
		}
	}
	for _, bidOF := range bidOFs {
		if err := bidOF.Validate2(); err != nil {
			return nil, err
		}
	}

	// Make sure none of them are partially filled except possibly the last in each list.
	var partialFulfillments []*OrderFulfillment
	lastAskI, lastBidI := len(askOFs)-1, len(bidOFs)-1
	for i, askOF := range askOFs {
		if !askOF.IsFullyFilled() {
			if i != lastAskI {
				return nil, fmt.Errorf("%s order %d (at index %d) is not filled in full and is not the last ask order provided",
					askOF.GetOrderType(), askOF.GetOrderID(), i)
			}
			partialFulfillments = append(partialFulfillments, askOF)
		}
	}
	for i, bidOF := range bidOFs {
		if !bidOF.IsFullyFilled() {
			if i != lastBidI {
				return nil, fmt.Errorf("%s order %d (at index %d) is not filled in full and is not the last bid order provided",
					bidOF.GetOrderType(), bidOF.GetOrderID(), i)
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
// Assumes that all fulfillments have passed Validate2.
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
		rv.OrderTransfers = append(rv.OrderTransfers, GetAssetTransfer2(of), GetPriceTransfer2(of))
		if !of.FeesToPay.IsZero() {
			// Using NewCoins in here as a last-ditch negative amount panic check.
			fees := sdk.NewCoins(of.FeesToPay...)
			indexedFees.add(of.GetOwner(), fees...)
		}
	}

	rv.FeeInputs = indexedFees.getAsInputs()

	return rv
}

// GetAssetTransfer2 gets the inputs and outputs to facilitate the transfers of assets for this order fulfillment.
// Assumes that the fulfillment has passed Validate2 already.
// Panics if any amounts are negative or if it's neither a bid nor ask order.
func GetAssetTransfer2(f *OrderFulfillment) *Transfer {
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

// GetPriceTransfer2 gets the inputs and outputs to facilitate the transfers for the price of this order fulfillment.
// Assumes that the fulfillment has passed Validate2 already.
// Panics if any amounts are negative or if it's neither a bid nor ask order.
func GetPriceTransfer2(f *OrderFulfillment) *Transfer {
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
