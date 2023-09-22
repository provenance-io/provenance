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
func (i *IndexedAddrAmts) Add(addr string, coin sdk.Coin) {
	n, known := i.indexes[addr]
	if !known {
		n = len(i.addrs)
		i.indexes[addr] = n
		i.addrs = append(i.addrs, addr)
		i.amts = append(i.amts, sdk.NewCoins())
	}
	i.amts[n] = i.amts[n].Add(coin)
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

// Validate does nothing but satisfies the OrderI interface.
func (f OrderFulfillment) Validate() error {
	return nil
}

// Finalize does some final calculations and validation for this order fulfillment.
// This order fulfillment and the ones in it maybe updated during this.
func (f *OrderFulfillment) Finalize() error {
	if len(f.Splits) == 0 || f.AssetsFilledAmt.IsZero() {
		return fmt.Errorf("%s order %d not even partially filled", f.GetOrderType(), f.GetOrderID())
	}

	if f.AssetsLeftAmt.IsNegative() {
		return fmt.Errorf("%s order %d having assets %q cannot fill be filled with %q: overfill",
			f.GetOrderType(), f.GetOrderID(), f.GetAssets(), f.GetAssetsFilled())
	}

	isAskOrder, isBidOrder := f.IsAskOrder(), f.IsBidOrder()
	orderAssets := f.GetAssets()
	orderPrice := f.GetPrice()
	targetPriceAmt := orderPrice.Amount

	if !f.AssetsLeftAmt.IsZero() {
		if !f.PartialFillAllowed() {
			return fmt.Errorf("%s order %d having assets %q cannot be partially filled with %q: "+
				"order does not allow partial fulfillment",
				f.GetOrderType(), f.GetOrderID(), orderAssets, f.GetAssetsFilled())
		}

		priceAssets := orderPrice.Amount.Mul(f.AssetsFilledAmt)
		priceRem := priceAssets.Mod(orderAssets.Amount)
		if !priceRem.IsZero() {
			return fmt.Errorf("%s order %d having assets %q cannot be partially filled by %q: "+
				"price %q is not evenly divisible",
				f.GetOrderType(), f.GetOrderID(), orderAssets, f.GetAssetsFilled(), orderPrice)
		}
		targetPriceAmt = priceAssets.Quo(orderAssets.Amount)
	}

	if isAskOrder {
		// For ask orders, we need the updated price to have the same price/asset ratio
		// as when originally created. Since ask orders can receive more payment than they requested,
		// the PriceLeftAmt in here might be less than that, and we need to fix it.
		f.PriceLeftAmt = orderPrice.Amount.Sub(targetPriceAmt)
	}

	if isBidOrder {
		// When adding things to f.PriceFilledAmt, we used truncation on the divisions.
		// So at this point, it might be a little less than the target price.
		// If that's the case, we distribute the difference weighted by assets in order of the splits.
		toDistribute := targetPriceAmt.Sub(f.PriceFilledAmt)
		if toDistribute.IsNegative() {
			return fmt.Errorf("bid order %d having price %q cannot pay %q for %q: overfill",
				f.GetOrderID(), f.GetPrice(), f.PriceFilledAmt, f.GetAssetsFilled())
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
					// Not updating askSplit.Order.PriceLeftAmt here since that's done specially above.
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
	if !f.AssetsLeftAmt.IsZero() {
		var assetsFilledAmt, orderAssetsAmt sdkmath.Int // Temporary line so we can still compile.
		orderFees := f.GetSettlementFees()
		for _, orderFee := range orderFees {
			feeAssets := orderFee.Amount.Mul(assetsFilledAmt)
			feeRem := feeAssets.Mul(orderAssetsAmt)
			if !feeRem.IsZero() {
				return fmt.Errorf("%s order %d having settlement fees %q cannot be partially filled by %q: "+
					"fee %q is not evenly divisible",
					f.GetOrderType(), f.GetOrderID(), orderFees, f.GetAssetsFilled(), orderFee)
			}
			feeAmtLeft := feeAssets.Quo(orderAssetsAmt)
			f.FeesLeft = f.FeesLeft.Add(sdk.NewCoin(orderFee.Denom, feeAmtLeft))
		}
	}

	panic("not implemented")
}

// IsFilled returns true if this fulfillment's order has been fully accounted for.
func (f OrderFulfillment) IsFilled() bool {
	return f.AssetsLeftAmt.IsZero()
}

// getAssetInputsOutputs gets the inputs and outputs to facilitate the transfers of assets for this order fulfillment.
func (f OrderFulfillment) getAssetInputsOutputs() ([]banktypes.Input, []banktypes.Output, error) {
	indexedSplits := NewIndexedAddrAmts()
	for _, split := range f.Splits {
		indexedSplits.Add(split.Order.GetOwner(), split.Assets)
	}

	if f.IsAskOrder() {
		inputs := []banktypes.Input{{Address: f.GetOwner(), Coins: sdk.Coins{f.GetAssetsFilled()}}}
		outputs := indexedSplits.GetAsOutputs()
		return inputs, outputs, nil
	}
	if f.IsBidOrder() {
		inputs := indexedSplits.GetAsInputs()
		outputs := []banktypes.Output{{Address: f.GetOwner(), Coins: sdk.Coins{f.GetAssetsFilled()}}}
		return inputs, outputs, nil
	}
	return nil, nil, fmt.Errorf("unknown order type %T", f.Order.GetOrder())
}

// getPriceInputsOutputs gets the inputs and outputs to facilitate the transfers for the price of this order fulfillment.
func (f OrderFulfillment) getPriceInputsOutputs() ([]banktypes.Input, []banktypes.Output, error) {
	indexedSplits := NewIndexedAddrAmts()
	for _, split := range f.Splits {
		indexedSplits.Add(split.Order.GetOwner(), split.Price)
	}

	if f.IsAskOrder() {
		inputs := indexedSplits.GetAsInputs()
		outputs := []banktypes.Output{{Address: f.GetOwner(), Coins: sdk.Coins{f.GetPriceFilled()}}}
		return inputs, outputs, nil
	}
	if f.IsBidOrder() {
		inputs := []banktypes.Input{{Address: f.GetOwner(), Coins: sdk.Coins{f.GetPriceFilled()}}}
		outputs := indexedSplits.GetAsOutputs()
		return inputs, outputs, nil
	}
	return nil, nil, fmt.Errorf("unknown order type %T", f.Order.GetOrder())
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
	f.AssetsLeftAmt = newAssetsLeftAmt
	f.AssetsFilledAmt = f.AssetsFilledAmt.Add(assetsAmt)
	f.PriceLeftAmt = f.PriceLeftAmt.Sub(priceAmt)
	f.PriceFilledAmt = f.PriceFilledAmt.Add(priceAmt)
	f.Splits = append(f.Splits, &OrderSplit{
		Order:  order,
		Assets: assets,
		Price:  price,
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

	assetsAmt, err := getFulfillmentAssetsAmt(askOF, bidOF)
	if err != nil {
		return err
	}

	// We calculate the price amount based off the original order assets (as opposed to assets left)
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
	askAssetsLeft, bidAssetsLeft := askOF.GetAssetsLeft(), bidOF.GetAssetsLeft()
	askOrderID, bidOrderID := askOF.GetOrderID(), bidOF.GetOrderID()
	stdErr := func(msg string) error {
		return fmt.Errorf("cannot fill ask order %d having assets left %q with bid order %d having assets left %q: %s",
			askOrderID, askAssetsLeft, bidOrderID, bidAssetsLeft, msg)
	}

	askAmt, bidAmt := askOF.AssetsLeftAmt, bidOF.AssetsLeftAmt
	if askAmt.IsZero() || askAmt.IsNegative() || bidAmt.IsZero() || bidAmt.IsNegative() {
		return sdkmath.ZeroInt(), stdErr("zero or negative assets")
	}

	// If they're equal, we're all good, just return one of them.
	if askAmt.Equal(bidAmt) {
		return askAmt, nil
	}
	// Use the lesser of the two.
	if askAmt.LT(bidAmt) {
		return askAmt, nil
	}
	return bidAmt, nil
}

type OrderTransfers struct {
	AssetInputs  []banktypes.Input
	AssetOutputs []banktypes.Output
	PriceInputs  []banktypes.Input
	PriceOutputs []banktypes.Output
	FeeInputs    []banktypes.Input
	FeeTotal     sdk.Coins
}
