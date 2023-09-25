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
	// AssetsFilledAmt is the total amount of assets being fulfilled for the order.
	AssetsFilledAmt sdkmath.Int
	// AssetsLeftAmt is the amount of order assets that have not yet been fulfilled for the order.
	AssetsLeftAmt sdkmath.Int
	// PriceFilledAmt is the total price amount involved in this order fulfillment.
	// If this is a bid order, the PriceFilledAmt is related to the order price.
	// If this is an ask order, the PriceFilledAmt is related to the prices of the bid orders fulfilling this order.
	PriceFilledAmt sdkmath.Int
	// PriceLeftAmt is the price that has not yet been fulfilled for the order.
	PriceLeftAmt sdkmath.Int
	// FeesToPay is the amount of fees to pay for this order.
	// This is not tracked as fulfillments are applied, it is only set during Finalize().
	FeesToPay sdk.Coins
	// FeesLeft is the amount fees left to pay (if this order is only partially filled).
	// This is not tracked as fulfillments are applied, it is only set during Finalize().
	FeesLeft sdk.Coins
	// Splits contains information on the orders being used to fulfill this order.
	Splits []*OrderSplit
}

var _ OrderI = (*OrderFulfillment)(nil)

func NewOrderFulfillment(order *Order) *OrderFulfillment {
	return &OrderFulfillment{
		Order:           order,
		AssetsFilledAmt: sdkmath.ZeroInt(),
		AssetsLeftAmt:   order.GetAssets().Amount,
		PriceFilledAmt:  sdkmath.ZeroInt(),
		PriceLeftAmt:    order.GetPrice().Amount,
	}
}

// GetAssetsFilled gets the coin value of the assets that have been filled in this fulfillment.
func (f OrderFulfillment) GetAssetsFilled() sdk.Coin {
	return sdk.Coin{Denom: f.GetAssets().Denom, Amount: f.AssetsFilledAmt}
}

// GetAssetsLeft gets the coin value of the assets left to fill in this fulfillment.
func (f OrderFulfillment) GetAssetsLeft() sdk.Coin {
	return sdk.Coin{Denom: f.GetAssets().Denom, Amount: f.AssetsLeftAmt}
}

// GetPriceFilled gets the coin value of the price that has been filled in this fulfillment.
func (f OrderFulfillment) GetPriceFilled() sdk.Coin {
	return sdk.Coin{Denom: f.GetPrice().Denom, Amount: f.PriceFilledAmt}
}

// GetPriceLeft gets the coin value of the price left to fill in this fulfillment.
func (f OrderFulfillment) GetPriceLeft() sdk.Coin {
	return sdk.Coin{Denom: f.GetPrice().Denom, Amount: f.PriceLeftAmt}
}

// IsFullyFilled returns true if this fulfillment's order has been fully accounted for.
func (f OrderFulfillment) IsFullyFilled() bool {
	return f.AssetsLeftAmt.IsZero()
}

// IsCompletelyUnfulfilled returns true if nothing in this order has been filled.
func (f OrderFulfillment) IsCompletelyUnfulfilled() bool {
	return len(f.Splits) == 0 || f.AssetsFilledAmt.IsZero()
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

// Validate does some final validation and sanity checking on this order fulfillment.
// It's assumed that Finalize has been called before calling this.
func (f OrderFulfillment) Validate() error {
	_ = &OrderFulfillment{
		Order:           nil,
		AssetsFilledAmt: sdkmath.Int{},
		AssetsLeftAmt:   sdkmath.Int{},
		PriceFilledAmt:  sdkmath.Int{},
		PriceLeftAmt:    sdkmath.Int{},
		FeesToPay:       nil,
		FeesLeft:        nil,
		Splits:          nil,
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
	orderAssets := f.GetAssets()
	if splitsAssets[0].Denom != orderAssets.Denom {
		return fmt.Errorf("splits asset denom %q does not match order assets %q on %s order %d",
			splitsAssets, orderAssets, f.GetOrderType(), f.GetOrderID())
	}
	if !splitsAssets[0].Amount.Equal(f.AssetsFilledAmt) {
		return fmt.Errorf("splits asset total %q does not match filled assets %q on %s order %d",
			splitsAssets, orderAssets, f.GetOrderType(), f.GetOrderID())
	}

	if len(splitsPrice) != 1 {
		return fmt.Errorf("multiple price denoms %q in splits applied to %s order %d",
			splitsPrice, f.GetOrderType(), f.GetOrderID())
	}
	orderPrice := f.GetPrice()
	if splitsPrice[0].Denom != orderPrice.Denom {
		return fmt.Errorf("splits price denom %q does not match order price %q on %s order %d",
			splitsPrice, orderPrice, f.GetOrderType(), f.GetOrderID())
	}
	if !splitsPrice[0].Amount.Equal(f.PriceFilledAmt) {
		return fmt.Errorf("splits price total %q does not match filled price %q on %s order %d",
			splitsPrice, orderPrice, f.GetOrderType(), f.GetOrderID())
	}

	if f.AssetsLeftAmt.IsNegative() {
		return fmt.Errorf("%s order %d having assets %q has negative assets left %q after filling %q",
			f.GetOrderType(), f.GetOrderID(), orderAssets, f.GetAssetsLeft(), f.GetAssetsFilled())
	}
	if f.PriceLeftAmt.IsNegative() {
		return fmt.Errorf("%s order %d having price %q has negative price left %q after filling %q",
			f.GetOrderType(), f.GetOrderID(), orderPrice, f.GetPriceLeft(), f.GetPriceFilled())
	}

	isFullyFilled := f.IsFullyFilled()
	switch {
	case f.IsAskOrder():
		// For ask orders, if being fully filled, the price filled needs to be at least the order price.
		if isFullyFilled && f.PriceFilledAmt.LT(orderPrice.Amount) {
			return fmt.Errorf("%s order %d having price %q cannot be filled at price %q: unsufficient price",
				f.GetOrderType(), f.GetOrderID(), orderPrice, f.GetPriceFilled())
		}
	case f.IsBidOrder():
		// If filled in full, the PriceFilledAmt must be equal to the order price.
		if isFullyFilled && !f.PriceFilledAmt.Equal(orderPrice.Amount) {
			return fmt.Errorf("%s order %d having price %q cannot be fully filled at price %q: price mismatch",
				f.GetOrderType(), f.GetOrderID(), orderPrice, f.GetPriceFilled())
		}
		// otherwise, the price filled must be less than the order price.
		if !isFullyFilled && f.PriceFilledAmt.GTE(orderPrice.Amount) {
			return fmt.Errorf("%s order %d having price %q cannot be partially filled at price %q: price mismatch",
				f.GetOrderType(), f.GetOrderID(), orderPrice, f.GetPriceFilled())
		}
	default:
		return fmt.Errorf("order %d has unknown type %s", f.GetOrderID(), f.GetOrderType())
	}

	if !isFullyFilled && !f.PartialFillAllowed() {
		return fmt.Errorf("cannot fill %s order %d having assets %q with assets %q: order does not allow partial fill",
			f.GetOrderType(), f.GetOrderID(), orderAssets, f.GetAssetsFilled())
	}

	return nil
}

// Apply adjusts this order fulfillment using the provided info.
func (f *OrderFulfillment) Apply(order *OrderFulfillment, assetsAmt, priceAmt sdkmath.Int) error {
	assets := sdk.NewCoin(order.GetAssets().Denom, assetsAmt)
	price := sdk.NewCoin(order.GetPrice().Denom, priceAmt)

	newAssetsLeftAmt := f.AssetsLeftAmt.Sub(assetsAmt)
	if newAssetsLeftAmt.IsNegative() {
		return fmt.Errorf("cannot fill %s order %d having assets left %q with %q from %s order %d: overfill",
			f.GetOrderType(), f.GetOrderID(), f.GetAssetsLeft(), assets, order.GetOrderType(), order.GetOrderID())
	}

	newPriceLeftAmt := f.PriceLeftAmt.Sub(priceAmt)
	// ask orders are allow to go negative on price left, but bid orders are not.
	if newPriceLeftAmt.IsNegative() && f.IsBidOrder() {
		return fmt.Errorf("cannot apply %s order %d having price left %q to %s order %d at a price of %q: overfill",
			f.GetOrderType(), f.GetOrderID(), f.GetPriceLeft(), order.GetOrderType(), order.GetOrderID(), price)
	}

	f.AssetsLeftAmt = newAssetsLeftAmt
	f.AssetsFilledAmt = f.AssetsFilledAmt.Add(assetsAmt)
	f.PriceLeftAmt = newPriceLeftAmt
	f.PriceFilledAmt = f.PriceFilledAmt.Add(priceAmt)
	f.Splits = append(f.Splits, &OrderSplit{
		Order:  order,
		Assets: assets,
		Price:  price,
	})
	return nil
}

// Finalize does some final calculations and validation for this order fulfillment.
// This order fulfillment and the ones in it maybe updated during this.
func (f *OrderFulfillment) Finalize(sellerFeeRatio *FeeRatio) error {
	if len(f.Splits) == 0 {
		return fmt.Errorf("no splits applied to %s order %d", f.GetOrderType(), f.GetOrderID())
	}
	if f.AssetsFilledAmt.IsZero() {
		return fmt.Errorf("cannot fill %s order %d with zero assets", f.GetOrderType(), f.GetOrderID())
	}
	if f.PriceFilledAmt.IsZero() {
		return fmt.Errorf("cannot fill %s order %d with zero price", f.GetOrderType(), f.GetOrderID())
	}

	if f.AssetsLeftAmt.IsNegative() {
		return fmt.Errorf("%s order %d having assets %q cannot fill be filled with %q: overfill",
			f.GetOrderType(), f.GetOrderID(), f.GetAssets(), f.GetAssetsFilled())
	}

	isAskOrder, isBidOrder := f.IsAskOrder(), f.IsBidOrder()
	orderAssets := f.GetAssets()
	orderPrice := f.GetPrice()
	orderFees := f.GetSettlementFees()
	targetPriceAmt := orderPrice.Amount

	isFullyFilled := f.IsFullyFilled()
	if !isFullyFilled {
		// Make sure the price can be split on a whole number.
		priceAssets := orderPrice.Amount.Mul(f.AssetsFilledAmt)
		priceRem := priceAssets.Mod(orderAssets.Amount)
		if !priceRem.IsZero() {
			return fmt.Errorf("%s order %d having assets %q cannot be partially filled by %q: "+
				"price %q is not evenly divisible",
				f.GetOrderType(), f.GetOrderID(), orderAssets, f.GetAssetsFilled(), orderPrice)
		}
		targetPriceAmt = priceAssets.Quo(orderAssets.Amount)

		// Make sure the fees can be split on a whole number.
		for _, orderFee := range orderFees {
			feeAssets := orderFee.Amount.Mul(f.AssetsFilledAmt)
			feeRem := feeAssets.Mul(orderAssets.Amount)
			if !feeRem.IsZero() {
				return fmt.Errorf("%s order %d having settlement fees %q cannot be partially filled by %q: "+
					"fee %q is not evenly divisible",
					f.GetOrderType(), f.GetOrderID(), orderFees, f.GetAssetsFilled(), orderFee)
			}
			feeAmtToPay := feeAssets.Quo(orderAssets.Amount)
			f.FeesToPay = f.FeesToPay.Add(sdk.NewCoin(orderFee.Denom, feeAmtToPay))
		}
		feesLeft, hasNeg := orderFees.SafeSub(f.FeesToPay...)
		if hasNeg {
			return fmt.Errorf("%s order %d having fees %q has negative fees left %q after applying %q",
				f.GetOrderType(), f.GetOrderID(), orderFees, feesLeft, f.FeesToPay)
		}
		f.FeesLeft = feesLeft
	} else {
		f.FeesToPay = orderFees
		f.FeesLeft = nil
	}

	switch {
	case isAskOrder:
		if !isFullyFilled {
			// For partially filled ask orders, we need to maintain the same price/asset ratio. Since ask orders
			// can receive more payment than requested, the PriceLeftAmt might be too low, so correct it now.
			f.PriceLeftAmt = orderPrice.Amount.Sub(targetPriceAmt)
		}
		// For ask orders, we need to calculate and add the ratio fee to the fees to pay.
		if sellerFeeRatio != nil {
			feeToPay, err := sellerFeeRatio.ApplyToLoosely(f.GetPriceFilled())
			if err != nil {
				return fmt.Errorf("could not calculate %s order %d ratio fee: %w",
					f.GetOrderType(), f.GetOrderID(), err)
			}
			f.FeesToPay = f.FeesToPay.Add(feeToPay)
		}
	case isBidOrder:
		// When adding things to f.PriceFilledAmt, we used truncation on the divisions.
		// So, at this point, it might be a little less than the target price.
		// If that's the case, we distribute the difference weighted by assets in order of the splits.
		// When adding things to f.PriceFilledAmt, we used truncation on the divisions.
		// So at this point, it might be a little less than the target price.
		// If that's the case, we distribute the difference weighted by assets in order of the splits.
		toDistribute := targetPriceAmt.Sub(f.PriceFilledAmt)
		if toDistribute.IsNegative() {
			return fmt.Errorf("%s order %d having price %q cannot pay %q for %q: overfill",
				f.GetOrderType(), f.GetOrderID(), orderPrice, f.PriceFilledAmt, f.GetAssetsFilled())
		}
		if toDistribute.IsPositive() {
			distLeft := toDistribute
			// First pass, we won't default to 1 (if the calc comes up zero).
			// This helps weight larger orders that are at the end of the list.
			// But it's possible for all the calcs to come up zero, so after
			// the first pass, use a minimum of 1 for each distribution.
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
					if distAmt.GT(distLeft) {
						distAmt = distLeft
					}
					f.PriceFilledAmt = f.PriceFilledAmt.Add(distAmt)
					f.PriceLeftAmt = f.PriceLeftAmt.Sub(distAmt)
					askSplit.Price.Amount = askSplit.Price.Amount.Add(distAmt)
					askSplit.Order.PriceFilledAmt = askSplit.Order.PriceFilledAmt.Add(distAmt)
					// Not updating askSplit.Order.PriceLeftAmt here since that's done more directly above.
					for _, bidSplit := range askSplit.Order.Splits {
						if bidSplit.Order.GetOrderID() == f.GetOrderID() {
							bidSplit.Price.Amount = bidSplit.Price.Amount.Add(distAmt)
							break
						}
					}
					distLeft = distLeft.Sub(distAmt)
					if !distLeft.IsPositive() {
						break
					}
				}
				minOne = true
			}

			// If being partially filled, the PriceFilledAmt must now equal the target price.
			if !isFullyFilled && !f.PriceFilledAmt.Equal(targetPriceAmt) {
				return fmt.Errorf("%s order %d having assets %q and price %q cannot be partially filled "+
					"with %q assets at price %q: expected price %q",
					f.GetOrderType(), f.GetOrderID(), orderAssets, orderPrice,
					f.GetAssetsFilled(), f.GetPriceFilled(), sdk.Coin{Denom: orderPrice.Denom, Amount: targetPriceAmt})
			}
		}
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

	assetsAmt, err := getFulfillmentAssetsAmt(askOF, bidOF)
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

// getFulfillmentAssetsAmt figures out the assets that can be fulfilled with the two provided orders.
// It's assumed that the askOF is for an ask order, and the bidOF is for a bid order.
func getFulfillmentAssetsAmt(askOF, bidOF *OrderFulfillment) (sdkmath.Int, error) {
	askAmtLeft, bidAmtLeft := askOF.AssetsLeftAmt, bidOF.AssetsLeftAmt
	if !askAmtLeft.IsPositive() || !bidAmtLeft.IsPositive() {
		return sdkmath.ZeroInt(), fmt.Errorf("cannot fill ask order %d having assets left %q "+
			"with bid order %d having assets left %q: zero or negative assets left",
			askOF.GetOrderID(), askOF.GetAssetsLeft(), bidOF.GetOrderID(), bidOF.GetAssetsLeft())
	}

	// Return the lesser of the two.
	if askAmtLeft.LTE(bidAmtLeft) {
		return askAmtLeft, nil
	}
	return bidAmtLeft, nil
}

type PartialFulfillment struct {
	NewOrder     *Order
	AssetsFilled sdk.Coin
	PriceFilled  sdk.Coin
}

func NewPartialFulfillment(f *OrderFulfillment) *PartialFulfillment {
	order := NewOrder(f.GetOrderID())
	if f.IsAskOrder() {
		askOrder := &AskOrder{
			MarketId:     f.GetMarketID(),
			Seller:       f.GetOwner(),
			Assets:       f.GetAssetsLeft(),
			Price:        f.GetPriceLeft(),
			AllowPartial: f.PartialFillAllowed(),
		}
		if !f.FeesLeft.IsZero() {
			if len(f.FeesLeft) > 1 {
				panic(fmt.Errorf("partially filled ask order %d somehow has multiple denoms in fees left %q",
					order.OrderId, f.FeesLeft))
			}
			askOrder.SellerSettlementFlatFee = &f.FeesLeft[0]
		}
		return &PartialFulfillment{
			NewOrder:     order,
			AssetsFilled: f.GetAssetsFilled(),
			PriceFilled:  f.GetPriceFilled(),
		}
	}

	if f.IsBidOrder() {
		bidOrder := &BidOrder{
			MarketId:            f.GetMarketID(),
			Buyer:               f.GetOwner(),
			Assets:              f.GetAssetsLeft(),
			Price:               f.GetPriceLeft(),
			BuyerSettlementFees: f.FeesLeft,
			AllowPartial:        f.PartialFillAllowed(),
		}
		return &PartialFulfillment{
			NewOrder:     order.WithBid(bidOrder),
			AssetsFilled: f.GetAssetsFilled(),
			PriceFilled:  f.GetPriceFilled(),
		}
	}

	panic(fmt.Errorf("order %d has unknown type %q", order.OrderId, f.GetOrderType()))
}

type Fulfillments struct {
	AskOFs       []*OrderFulfillment
	BidOFs       []*OrderFulfillment
	PartialOrder *PartialFulfillment
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

	for _, bidOF := range bidOFs {
		if err := bidOF.Validate(); err != nil {
			return nil, err
		}
	}
	for _, askOF := range askOFs {
		if err := askOF.Validate(); err != nil {
			return nil, err
		}
	}

	rv := &Fulfillments{
		AskOFs: askOFs,
		BidOFs: bidOFs,
	}
	if !askOFs[len(askOFs)-1].IsFullyFilled() {
		rv.PartialOrder = NewPartialFulfillment(askOFs[len(askOFs)-1])
	} else if !bidOFs[len(bidOFs)-1].IsFullyFilled() {
		rv.PartialOrder = NewPartialFulfillment(bidOFs[len(bidOFs)-1])
	}

	return rv, nil
}

// indexedAddrAmts is a set of addresses and amounts.
type indexedAddrAmts struct {
	addrs   []string
	amts    []sdk.Coins
	indexes map[string]int
}

func newIndexedAddrAmts() *indexedAddrAmts {
	return &indexedAddrAmts{
		indexes: make(map[string]int),
	}
}

// Add adds the coins to the input with the given address (creating it if needed).
func (i *indexedAddrAmts) Add(addr string, coins ...sdk.Coin) {
	n, known := i.indexes[addr]
	if !known {
		n = len(i.addrs)
		i.indexes[addr] = n
		i.addrs = append(i.addrs, addr)
		i.amts = append(i.amts, sdk.NewCoins())
	}
	i.amts[n] = i.amts[n].Add(coins...)
}

// GetAsInputs returns all the entries as bank Inputs.
func (i *indexedAddrAmts) GetAsInputs() []banktypes.Input {
	rv := make([]banktypes.Input, len(i.addrs))
	for n, addr := range i.addrs {
		rv[n] = banktypes.Input{Address: addr, Coins: i.amts[n]}
	}
	return rv
}

// GetAsOutputs returns all the entries as bank Outputs.
func (i *indexedAddrAmts) GetAsOutputs() []banktypes.Output {
	rv := make([]banktypes.Output, len(i.addrs))
	for n, addr := range i.addrs {
		rv[n] = banktypes.Output{Address: addr, Coins: i.amts[n]}
	}
	return rv
}

// Transfer contains bank inputs and outputs indicating a transfer that needs to be made.
type Transfer struct {
	Inputs  []banktypes.Input
	Outputs []banktypes.Output
}

// SettlementTransfers has everything needed to do all the transfers for a settlement.
type SettlementTransfers struct {
	OrderTransfers []*Transfer
	FeeInputs      []banktypes.Input
}

// BuildSettlementTransfers creates all the order transfers needed for the provided fulfillments.
func BuildSettlementTransfers(fulfillments *Fulfillments) *SettlementTransfers {
	indexedFees := newIndexedAddrAmts()

	rv := &SettlementTransfers{}
	applyOF := func(of *OrderFulfillment) {
		rv.OrderTransfers = append(rv.OrderTransfers, getAssetTransfer(of), getPriceTransfer(of))
		if !of.FeesToPay.IsZero() {
			// Using NewCoins in here (instead of Coins{...}) as a last-ditch negative amount panic check.
			fees := sdk.NewCoins(of.FeesToPay...)
			indexedFees.Add(of.GetOwner(), fees...)
		}
	}

	for _, askOF := range fulfillments.AskOFs {
		applyOF(askOF)
	}
	for _, bidOf := range fulfillments.BidOFs {
		applyOF(bidOf)
	}

	rv.FeeInputs = indexedFees.GetAsInputs()

	return rv
}

// getAssetTransfer gets the inputs and outputs to facilitate the transfers of assets for this order fulfillment.
func getAssetTransfer(f *OrderFulfillment) *Transfer {
	indexedSplits := newIndexedAddrAmts()
	for _, split := range f.Splits {
		indexedSplits.Add(split.Order.GetOwner(), split.Assets)
	}

	// Using NewCoins in here (instead of Coins{...}) as a last-ditch negative amount panic check.
	if f.IsAskOrder() {
		return &Transfer{
			Inputs:  []banktypes.Input{{Address: f.GetOwner(), Coins: sdk.NewCoins(f.GetAssetsFilled())}},
			Outputs: indexedSplits.GetAsOutputs(),
		}
	}
	if f.IsBidOrder() {
		return &Transfer{
			Inputs:  indexedSplits.GetAsInputs(),
			Outputs: []banktypes.Output{{Address: f.GetOwner(), Coins: sdk.NewCoins(f.GetAssetsFilled())}},
		}
	}

	// panicking in here if there's an error since it really should have happened earlier anyway.
	panic(fmt.Errorf("unknown order type %T", f.Order.GetOrder()))
}

// getPriceTransfer gets the inputs and outputs to facilitate the transfers for the price of this order fulfillment.
func getPriceTransfer(f *OrderFulfillment) *Transfer {
	indexedSplits := newIndexedAddrAmts()
	for _, split := range f.Splits {
		indexedSplits.Add(split.Order.GetOwner(), split.Price)
	}

	// Using NewCoins in here (instead of Coins{...}) as a last-ditch negative amount panic check.
	if f.IsAskOrder() {
		return &Transfer{
			Inputs:  indexedSplits.GetAsInputs(),
			Outputs: []banktypes.Output{{Address: f.GetOwner(), Coins: sdk.NewCoins(f.GetPriceFilled())}},
		}
	}
	if f.IsBidOrder() {
		return &Transfer{
			Inputs:  []banktypes.Input{{Address: f.GetOwner(), Coins: sdk.NewCoins(f.GetPriceFilled())}},
			Outputs: indexedSplits.GetAsOutputs(),
		}
	}

	// panicking in here if there's an error since it really should have happened earlier anyway.
	panic(fmt.Errorf("unknown order type %T", f.Order.GetOrder()))
}
