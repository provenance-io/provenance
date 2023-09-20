package exchange

import (
	"fmt"

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

// GetInputs returns all the entries as bank Inputs.
func (i *IndexedAddrAmts) GetInputs() []banktypes.Input {
	rv := make([]banktypes.Input, len(i.addrs))
	for n, addr := range i.addrs {
		rv[n] = banktypes.Input{Address: addr, Coins: i.amts[n]}
	}
	return rv
}

// GetOutputs returns all the entries as bank Outputs.
func (i *IndexedAddrAmts) GetOutputs() []banktypes.Output {
	rv := make([]banktypes.Output, len(i.addrs))
	for n, addr := range i.addrs {
		rv[n] = banktypes.Output{Address: addr, Coins: i.amts[n]}
	}
	return rv
}

// OrderSplit contains an order, and the asset and price amounts that should come out of it.
type OrderSplit struct {
	Order  *Order
	Assets sdk.Coins
	Price  sdk.Coin
}

// AskOrderFulfillment represents how an ask order is fulfilled.
type AskOrderFulfillment struct {
	Order      *Order
	AssetsLeft sdk.Coins
	TotalPrice sdk.Coin
	Splits     []*OrderSplit
}

type BidOrderFulfillment struct {
	Order       *Order
	TotalAssets sdk.Coins
	PriceLeft   sdk.Coin
	Splits      []*OrderSplit
}

func NewAskOrderFulfillment(order *Order) *AskOrderFulfillment {
	if !order.IsAskOrder() {
		panic(fmt.Errorf("cannot create AskOrderFulfillment for %s order %d",
			order.GetOrderType(), order.OrderId))
	}
	askOrder := order.GetAskOrder()
	return &AskOrderFulfillment{
		Order:      order,
		AssetsLeft: askOrder.Assets,
		TotalPrice: sdk.NewInt64Coin(askOrder.Price.Denom, 0),
	}
}

func NewBidOrderFulfillment(order *Order) *BidOrderFulfillment {
	if !order.IsBidOrder() {
		panic(fmt.Errorf("cannot create BidOrderFulfillment for %s order %d",
			order.GetOrderType(), order.OrderId))
	}
	bidOrder := order.GetBidOrder()
	return &BidOrderFulfillment{
		Order:     order,
		PriceLeft: bidOrder.Price,
	}
}

// AddSplit applies the given bid order in the amount of assets provided to this ask order.
func (f *AskOrderFulfillment) AddSplit(order *Order, assets sdk.Coins) error {
	askOrderID := f.Order.OrderId
	bidOrderID := order.OrderId
	if !order.IsBidOrder() {
		return fmt.Errorf("cannot fill ask order %d with %s order %d",
			askOrderID, order.GetOrderType(), bidOrderID)
	}

	askOrder := f.Order.GetAskOrder()
	bidOrder := order.GetBidOrder()

	if askOrder.Price.Denom != bidOrder.Price.Denom {
		return fmt.Errorf("cannot fill ask order %d having price %q with bid order %d having price %q: denom mismatch",
			askOrderID, askOrder.Price, bidOrderID, bidOrder.Price)
	}

	if assets.IsZero() {
		return fmt.Errorf("cannot fill ask order %d with zero assets from bid order %d", askOrderID, bidOrderID)
	}
	if assets.IsAnyNegative() {
		return fmt.Errorf("cannot fill ask order %d with negative assets %q from bid order %d",
			askOrderID, assets, bidOrderID)
	}
	// TODO[1658]: This should be checked against the assets left to fill in the bid order.
	_, hasNeg := bidOrder.Assets.SafeSub(assets...)
	if hasNeg {
		return fmt.Errorf("cannot fill ask order %d with assets %q from bid order %d: insufficient assets %q in bid order",
			askOrderID, assets, bidOrderID, bidOrder.Assets)
	}
	if len(bidOrder.Assets) > 1 && !CoinsEquals(bidOrder.Assets, assets) {
		return fmt.Errorf("cannot fill ask order %d with assets %q from bid order %d having assets %q: "+
			"unable to divide price for multiple asset types",
			askOrderID, assets, bidOrderID, bidOrder.Assets)
	}

	newAssetsLeft, hasNeg := f.AssetsLeft.SafeSub(assets...)
	if hasNeg {
		return fmt.Errorf("cannot fill ask order %d having %q left with assets %q from bid order %d",
			askOrderID, f.AssetsLeft, assets, bidOrderID)
	}

	split := &OrderSplit{
		Order:  order,
		Assets: assets,
	}

	if CoinsEquals(assets, bidOrder.Assets) {
		split.Price = bidOrder.Price
	} else {
		if len(assets) != 1 || len(bidOrder.Assets) != 1 {
			return fmt.Errorf("cannot prorate bid order %d assets %q for ask order %d with assets %q: multiple asset denoms",
				bidOrderID, assets, askOrderID, askOrder.Assets)
		}
		// Note, we truncate the division here.
		// Later, after all fulfillments have been identified, we will handle the remainders.
		priceAmt := bidOrder.Price.Amount.Mul(askOrder.Assets[0].Amount).Quo(assets[0].Amount)
		split.Price = sdk.NewCoin(f.TotalPrice.Denom, priceAmt)
	}

	f.Splits = append(f.Splits, split)
	f.AssetsLeft = newAssetsLeft
	f.TotalPrice = f.TotalPrice.Add(split.Price)
	return nil
}

func (f *BidOrderFulfillment) AddSplit(order *Order, assets sdk.Coins) error {
	askOrderID := order.OrderId
	bidOrderID := f.Order.OrderId
	if !order.IsAskOrder() {
		return fmt.Errorf("cannot fill bid order %d with %s order %d",
			askOrderID, order.GetOrderType(), bidOrderID)
	}

	askOrder := order.GetAskOrder()
	bidOrder := f.Order.GetBidOrder()

	if askOrder.Price.Denom != bidOrder.Price.Denom {
		return fmt.Errorf("cannot fill bid order %d having price %q with ask order %d having price %q: denom mismatch",
			bidOrderID, bidOrder.Price, askOrderID, askOrder.Price)
	}

	if assets.IsZero() {
		return fmt.Errorf("cannot fill bid order %d with zero assets from ask order %d", bidOrderID, askOrderID)
	}
	if assets.IsAnyNegative() {
		return fmt.Errorf("cannot fill bid order %d with negative assets %q from ask order %d",
			bidOrderID, assets, askOrderID)
	}
	// TODO[1658]: This should be checked against the assets left to fill in the ask order.
	_, hasNeg := askOrder.Assets.SafeSub(assets...)
	if hasNeg {
		return fmt.Errorf("cannot fill bid order %d with assets %q from ask order %d: insufficient assets %q in ask order",
			bidOrderID, assets, askOrderID, askOrder.Assets)
	}

	split := &OrderSplit{
		Order:  order,
		Assets: assets,
	}

	if CoinsEquals(assets, bidOrder.Assets) {
		split.Price = bidOrder.Price
	} else {
		if len(assets) != 1 || len(bidOrder.Assets) != 1 {
			return fmt.Errorf("cannot prorate bid order %d assets %q for ask order %d with assets %q: multiple asset denoms",
				bidOrderID, assets, askOrderID, askOrder.Assets)
		}
		// Note, we truncate the division here.
		// Later, after all fulfillments have been identified, we will handle the remainders.
		priceAmt := bidOrder.Price.Amount.Mul(bidOrder.Assets[0].Amount).Quo(assets[0].Amount)
		split.Price = sdk.NewCoin(f.PriceLeft.Denom, priceAmt)
	}
	newPriceLeftAmt := f.PriceLeft.Amount.Sub(split.Price.Amount)
	if newPriceLeftAmt.IsNegative() {
		return fmt.Errorf("cannot fill bid order %d having assets %q with assets %q for ask order %d: "+
			"price left %q is less than price needed %q",
			bidOrderID, bidOrder.Assets, assets, askOrderID, f.PriceLeft, split.Price)
	}

	f.Splits = append(f.Splits, split)
	f.TotalAssets = f.TotalAssets.Add(assets...)
	f.PriceLeft.Amount = newPriceLeftAmt
	return nil
}
