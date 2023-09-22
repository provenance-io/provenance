package exchange

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// IndexedAddrAmts is a set of addresses and amounts.
type IndexedAddrAmts struct {
	addrs   []string
	amts    []sdk.Coins
	indexes map[string]int
}

func NewIndexedAddrAmts() *IndexedAddrAmts {
	return &IndexedAddrAmts{
		indexes: make(map[string]int),
	}
}

// Add adds the coins to the input with the given address (creating it if needed).
func (i *IndexedAddrAmts) Add(addr string, coins sdk.Coins) {
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
func (i *IndexedAddrAmts) GetAsInputs() []banktypes.Input {
	rv := make([]banktypes.Input, len(i.addrs))
	for n, addr := range i.addrs {
		rv[n] = banktypes.Input{Address: addr, Coins: i.amts[n]}
	}
	return rv
}

// GetAsOutputs returns all the entries as bank Outputs.
func (i *IndexedAddrAmts) GetAsOutputs() []banktypes.Output {
	rv := make([]banktypes.Output, len(i.addrs))
	for n, addr := range i.addrs {
		rv[n] = banktypes.Output{Address: addr, Coins: i.amts[n]}
	}
	return rv
}

// OrderSplit contains an order, and the asset and price amounts that should come out of it.
type OrderSplit struct {
	// Order fulfillment associated with this split.
	Order *OrderFulfillment
	// Assets is the amount of assets from the order involved in this split.
	Assets sdk.Coins
	// Price is the amount of the price from the order involved in this split.
	Price sdk.Coin
}

// OrderFulfillment is used to figure out how an order should be fulfilled.
type OrderFulfillment struct {
	// Order is the original order with all its information.
	Order *Order
	// AssetsFilled is the total assets being fulfilled for the order.
	AssetsFilled sdk.Coins
	// AssetsLeft is the amount of order assets that have not yet been fulfilled for the order.
	AssetsLeft sdk.Coins
	// PriceAmtFilled is the total price amount involved in this order fulfillment.
	// If this is a bid order, the PriceAmtFilled is related to the order price.
	// If this is an ask order, the PriceAmtFilled is related to the prices of the bid orders fulfilling this order.
	PriceAmtFilled sdkmath.Int
	// PriceAmtLeft is the price that has not yet been fulfilled for the order.
	PriceAmtLeft sdkmath.Int
	// FeesLeft is the amount fees left to pay (if this order is only partially filled).
	// This is not tracked as fulfillments are applied, it is only set during Finalize().
	FeesLeft sdk.Coins
	// Splits contains information on the orders being used to fulfill this order.
	Splits []*OrderSplit
}

var _ OrderI = (*OrderFulfillment)(nil)

func NewOrderFulfillment(order *Order) *OrderFulfillment {
	return &OrderFulfillment{
		Order:          order,
		AssetsLeft:     order.GetAssets(),
		PriceAmtFilled: sdkmath.ZeroInt(),
		PriceAmtLeft:   order.GetPrice().Amount,
	}
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
func (f OrderFulfillment) GetAssets() sdk.Coins {
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

// Finalize does some final calculations and validation for this order fulfillment.
// This order fulfillment and the ones in it maybe updated during this.
func (f *OrderFulfillment) Finalize() error {
	if len(f.Splits) == 0 || f.AssetsFilled.IsZero() {
		return fmt.Errorf("%s order %d not even partially filled", f.GetOrderType(), f.GetOrderID())
	}

	if f.AssetsLeft.IsAnyNegative() {
		return fmt.Errorf("%s order %d having assets %q cannot fill be filled with %q: overfill",
			f.GetOrderType(), f.GetOrderID(), f.GetAssets(), f.AssetsFilled)
	}

	isAskOrder, isBidOrder := f.IsAskOrder(), f.IsBidOrder()
	orderAssets := f.GetAssets()
	orderPrice := f.GetPrice()
	targetPriceAmt := orderPrice.Amount

	if !f.AssetsLeft.IsZero() {
		if !f.PartialFillAllowed() {
			return fmt.Errorf("%s order %d having assets %q cannot be partially filled with %q: "+
				"order does not allow partial fulfillment",
				f.GetOrderType(), f.GetOrderID(), orderAssets, f.AssetsFilled)
		}

		if len(orderAssets) != 1 {
			return fmt.Errorf("%s order %d having assets %q cannot be partially filled with %q: "+
				"orders with multiple asset types cannot be split",
				f.GetOrderType(), f.GetOrderID(), orderAssets, f.AssetsFilled)
		}

		assetsFilledAmt := f.AssetsFilled[0].Amount
		orderAssetsAmt := orderAssets[0].Amount
		priceAssets := orderPrice.Amount.Mul(assetsFilledAmt)
		priceRem := priceAssets.Mod(orderAssetsAmt)
		if !priceRem.IsZero() {
			return fmt.Errorf("%s order %d having assets %q cannot be partially filled by %q: "+
				"price %q is not evenly divisible",
				f.GetOrderType(), f.GetOrderID(), orderAssets, f.AssetsFilled, orderPrice)
		}
		targetPriceAmt = priceAssets.Quo(orderAssetsAmt)
	}

	if isAskOrder {
		// For ask orders, we need the updated price to have the same price/asset ratio
		// as when originally created. Since ask orders can receive more payment than they requested,
		// the PriceAmtLeft in here might be less than that, and we need to fix it.
		f.PriceAmtLeft = orderPrice.Amount.Sub(targetPriceAmt)
	}

	if isBidOrder {
		// When adding things to f.PriceAmtFilled, we used truncation on the divisions.
		// So at this point, it might be a little less than the target price.
		// If that's the case, we distribute the difference weighted by assets in order of the splits.
		toDistribute := targetPriceAmt.Sub(f.PriceAmtFilled)
		if toDistribute.IsNegative() {
			return fmt.Errorf("bid order %d having price %q cannot pay %q for %q: overfill",
				f.GetOrderID(), f.GetPrice(), f.PriceAmtFilled, f.AssetsFilled)
		}
		if toDistribute.IsPositive() {
			distLeft := toDistribute
			// First pass, we won't default to 1 (if the calc comes up zero).
			// This helps weight larger orders that are at the end of the list.
			// But it's possible for all the calcs to come up zero, so eventually
			// we might need to default to one.
			minOne := false
			for distLeft.IsPositive() {
				// OrderFulfillment won't let a bid order with multiple assets be split.
				// So we know that here, there's only one asset.
				for _, askSplit := range f.Splits {
					distAmt := toDistribute.Mul(askSplit.Assets[0].Amount).Quo(f.AssetsFilled[0].Amount)
					if distAmt.IsZero() {
						if !minOne {
							continue
						}
						distAmt = sdkmath.OneInt()
					}
					if distAmt.GT(distLeft) {
						distAmt = distLeft
					}
					f.PriceAmtFilled = f.PriceAmtFilled.Add(distAmt)
					f.PriceAmtLeft = f.PriceAmtLeft.Sub(distAmt)
					askSplit.Price.Amount = askSplit.Price.Amount.Add(distAmt)
					askSplit.Order.PriceAmtFilled = askSplit.Order.PriceAmtFilled.Add(distAmt)
					// Not updating askSplit.Order.PriceAmtLeft here since that's done specially above.
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
		}
	}

	// TODO[1658]: Finish up Finalize().
	if !f.AssetsLeft.IsZero() {
		var assetsFilledAmt, orderAssetsAmt sdkmath.Int // Temporary line so we can still compile.
		orderFees := f.GetSettlementFees()
		for _, orderFee := range orderFees {
			feeAssets := orderFee.Amount.Mul(assetsFilledAmt)
			feeRem := feeAssets.Mul(orderAssetsAmt)
			if !feeRem.IsZero() {
				return fmt.Errorf("%s order %d having settlement fees %q cannot be partially filled by %q: "+
					"fee %q is not evenly divisible",
					f.GetOrderType(), f.GetOrderID(), orderFees, f.AssetsFilled, orderFee)
			}
			feeAmtLeft := feeAssets.Quo(orderAssetsAmt)
			f.FeesLeft = f.FeesLeft.Add(sdk.NewCoin(orderFee.Denom, feeAmtLeft))
		}
	}

	panic("not implemented")
}

// Validate returns an error if there is a problem with this fulfillment.
func (f OrderFulfillment) Validate() error {
	var assetsFilled sdk.Coins
	pricePaid := sdk.NewInt64Coin(f.GetPrice().Denom, 0)
	for _, split := range f.Splits {
		assetsFilled = assetsFilled.Add(split.Assets...)
		pricePaid = pricePaid.Add(split.Price)
	}
	orderAssets := f.GetAssets()
	unfilledAssets, hasNeg := orderAssets.SafeSub(assetsFilled...)
	if hasNeg {
		return fmt.Errorf("assets fulfilled %q for order %d is more than the order assets %q", assetsFilled, f.GetOrderID(), orderAssets)
	}

	// If it's a bid order:
	// 	If filled in full, the pricePaid needs to equal the order price.
	//  If not filled in full:
	// 		The price * assets filled / order assets must be a whole number and equal the price paid.
	// 		The settlement fee * assets filled / order assets must be a whole number.
	// If it's an ask order:
	//  If filled in full, the pricePaid must be greater than or equal to the order price.
	//  If not filled in full:
	//		The price * assets filled / order assets must be a whole number and must be less than or equal to the price paid.
	//      The settlement fee * assets filled / order assets must be a whole number.
	if unfilledAssets.IsZero() {

	}

	if !unfilledAssets.IsZero() && !f.PartialFillAllowed() {
		return fmt.Errorf("cannot partially fill order %d having assets %q with assets %q: order does not allow partial fill",
			f.GetOrderID(), orderAssets, assetsFilled)
	}

	// TODO[1658]: Implement OrderFulfillment.Validate()
	panic("not implemented")
}

// IsFilled returns true if this fulfillment's order has been fully accounted for.
func (f OrderFulfillment) IsFilled() bool {
	return f.AssetsLeft.IsZero()
}

// getAssetInputsOutputs gets the inputs and outputs to facilitate the transfers of assets for this order fulfillment.
func (f OrderFulfillment) getAssetInputsOutputs() ([]banktypes.Input, []banktypes.Output, error) {
	indexedSplits := NewIndexedAddrAmts()
	for _, split := range f.Splits {
		indexedSplits.Add(split.Order.GetOwner(), split.Assets)
	}

	if f.IsAskOrder() {
		inputs := []banktypes.Input{{Address: f.GetOwner(), Coins: f.AssetsFilled}}
		outputs := indexedSplits.GetAsOutputs()
		return inputs, outputs, nil
	}
	if f.IsBidOrder() {
		inputs := indexedSplits.GetAsInputs()
		outputs := []banktypes.Output{{Address: f.GetOwner(), Coins: f.AssetsFilled}}
		return inputs, outputs, nil
	}
	return nil, nil, fmt.Errorf("unknown order type %T", f.Order.GetOrder())
}

// getPriceInputsOutputs gets the inputs and outputs to facilitate the transfers for the price of this order fulfillment.
// It is assumed that getAssetInputsOutputs has already been called and did not return an error.
func (f OrderFulfillment) getPriceInputsOutputs() ([]banktypes.Input, []banktypes.Output, error) {
	price := sdk.Coin{Denom: f.GetPrice().Denom, Amount: f.PriceAmtFilled}
	if f.PriceAmtLeft.IsNegative() && f.IsBidOrder() {
		return nil, nil, fmt.Errorf("bid order %d having price %q cannot pay %q: overfill",
			f.GetOrderID(), f.GetPrice(), price)
	}
	if !f.AssetsLeft.IsZero() {
		// Assuming getAssetInputsOutputs was called previously, which returns an error
		// if trying to partially fill an order with multiple asset types.
		// So, if we're here and there's assets left, there's only one denom.
		orderPrice := f.GetPrice()
		priceAssetsAmt := orderPrice.Amount.Mul(f.AssetsFilled[0].Amount)
		rem := priceAssetsAmt.Mod(f.GetAssets()[0].Amount)
		if !rem.IsZero() {
			return nil, nil, fmt.Errorf("%s order %d having assets %q cannot be partially filled by %q: "+
				"price %q is not evenly divisible",
				f.GetOrderType(), f.GetOrderID(), f.GetAssets(), f.AssetsFilled, orderPrice)
		}
	}

	// TODO[1658]: Finish getPriceInputsOutputs
	return nil, nil, nil
}

// Apply adjusts this order fulfillment using the provided info.
func (f *OrderFulfillment) Apply(order *OrderFulfillment, assets sdk.Coins, priceAmt sdkmath.Int) error {
	newAssetsLeft, hasNeg := f.AssetsLeft.SafeSub(assets...)
	if hasNeg {
		return fmt.Errorf("cannot fill %s order %d having assets left %q with %q from %s order %d: overfill",
			f.GetOrderType(), f.GetOrderID(), f.AssetsLeft, assets, order.GetOrderType(), order.GetOrderID())
	}
	f.AssetsLeft = newAssetsLeft
	f.AssetsFilled = f.AssetsFilled.Add(assets...)
	f.PriceAmtLeft = f.PriceAmtLeft.Sub(priceAmt)
	f.PriceAmtFilled = f.PriceAmtFilled.Add(priceAmt)
	f.Splits = append(f.Splits, &OrderSplit{
		Order:  order,
		Assets: assets,
		Price:  sdk.NewCoin(order.GetPrice().Denom, priceAmt),
	})
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

	askOrderID, bidOrderID := askOF.GetOrderID(), bidOF.GetOrderID()
	askOrder, bidOrder := askOF.Order.GetAskOrder(), bidOF.Order.GetBidOrder()

	if askOrder.Price.Denom != bidOrder.Price.Denom {
		return fmt.Errorf("cannot fill bid order %d having price %q with ask order %d having price %q: denom mismatch",
			bidOrderID, bidOrder.Price, askOrderID, askOrder.Price)
	}

	assets, err := getFulfillmentAssets(askOF, bidOF)
	if err != nil {
		return err
	}

	// We calculate the price amount based off the original order assets (as opposed to assets left)
	// for consistent truncation and remainders. Once we've identified all the fulfillment relationships,
	// we'll enumerate and redistribute those remainders.
	priceAmt := bidOrder.Price.Amount
	if !CoinsEquals(assets, bidOrder.GetAssets()) {
		if len(assets) != 1 {
			return fmt.Errorf("cannot split bid order %d having assets %q by %q: overfill",
				bidOrderID, bidOrder.Assets, assets)
		}
		priceAmt = bidOrder.Price.Amount.Mul(assets[0].Amount).Quo(bidOrder.Assets.Amount)
	}

	askErr := askOF.Apply(bidOF, assets, priceAmt)
	bidErr := bidOF.Apply(askOF, assets, priceAmt)

	return errors.Join(askErr, bidErr)
}

// getFulfillmentAssets figures out the assets that can be fulfilled with the two provided orders.
// It's assumed that the askOF is for an ask order, and the bidOF is for a bid order.
func getFulfillmentAssets(askOF, bidOF *OrderFulfillment) (sdk.Coins, error) {
	askAssets, bidAssets := askOF.AssetsLeft, bidOF.AssetsLeft
	askOrderID, bidOrderID := askOF.GetOrderID(), bidOF.GetOrderID()
	stdErr := func(msg string) error {
		return fmt.Errorf("cannot fill ask order %d having assets left %q with bid order %d having assets left %q: %s",
			askOrderID, askAssets, bidOrderID, bidAssets, msg)
	}

	if askAssets.IsZero() || askAssets.IsAnyNegative() || bidAssets.IsZero() || bidAssets.IsAnyNegative() {
		return nil, stdErr("zero or negative assets")
	}

	// If they're equal, we're all good, just return one of them.
	if CoinsEquals(askAssets, bidAssets) {
		return askAssets, nil
	}

	// Handling single denom case now since that's expected to be the most common thing, and it's easier.
	if len(askAssets) == 1 && len(bidAssets) == 1 {
		if askAssets[0].Denom != bidAssets[0].Denom {
			return nil, stdErr("asset denom mismatch")
		}
		// Take the lesser of the two.
		if askAssets[0].Amount.LT(bidAssets[0].Amount) {
			return askAssets, nil
		}
		return bidAssets, nil
	}

	// When splitting a bid, we distribute the price based on the assets.
	// If the bid has multiple asset denoms, that division is impossible.
	// So, if the bid has multiple asset denoms, ensure they're less than or equal to the ask assets and return them.
	if len(bidAssets) > 1 {
		_, hasNeg := askAssets.SafeSub(bidAssets...)
		if hasNeg {
			return nil, stdErr("bid orders with multiple assets cannot be split")
		}
		return bidAssets, nil
	}

	// At this point, we know there's multiple ask assets, and only one bid asset.
	// All we care about are the amounts of that denom, taking the lesser of the two (but not zero).
	askAmt := askAssets.AmountOf(bidAssets[0].Denom)
	if !askAmt.IsPositive() {
		return nil, stdErr("asset denom mismatch")
	}
	if bidAssets[0].Amount.LTE(askAmt) {
		return bidAssets, nil
	}
	return sdk.NewCoins(sdk.NewCoin(bidAssets[0].Denom, askAmt)), nil
}

type OrderTransfers struct {
	AssetInputs  []banktypes.Input
	AssetOutputs []banktypes.Output
	PriceInputs  []banktypes.Input
	PriceOutputs []banktypes.Output
	FeeInputs    []banktypes.Input
	FeeTotal     sdk.Coins
}
