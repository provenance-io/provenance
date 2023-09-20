package exchange

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Define the type strings and bytes to use for each order type.
// These must all be unique. If you add entries, be sure to update TestOrderTypesAndBytes too.
const (
	OrderTypeAsk     = "ask"
	OrderTypeByteAsk = byte(0x00)

	OrderTypeBid     = "bid"
	OrderTypeByteBid = byte(0x01)
)

// findDuplicateIds returns all order ids that appear two or more times in the provided slice.
func findDuplicateIds(orderIDs []uint64) []uint64 {
	var rv []uint64
	seen := make(map[uint64]bool)
	dups := make(map[uint64]bool)
	for _, orderID := range orderIDs {
		if seen[orderID] && !dups[orderID] {
			rv = append(rv, orderID)
			dups[orderID] = true
		}
		seen[orderID] = true
	}
	return rv
}

// ValidateOrderIDs makes sure that one or more order ids are provided,
// none of them are zero, and there aren't any duplicates.
func ValidateOrderIDs(field string, orderIDs []uint64) error {
	if len(orderIDs) == 0 {
		return fmt.Errorf("no %s order ids provided", field)
	}
	if ContainsUint64(orderIDs, 0) {
		return fmt.Errorf("invalid %s order ids: cannot contain order id zero", field)
	}
	dupOrderIDs := findDuplicateIds(orderIDs)
	if len(dupOrderIDs) > 0 {
		return fmt.Errorf("duplicate %s order ids provided: %v", field, dupOrderIDs)
	}
	return nil
}

// NewOrder creates a new empty Order with the provided order id.
// The order details are set using one of: WithAsk, WithBid.
func NewOrder(orderID uint64) *Order {
	return &Order{OrderId: orderID}
}

// WithAsk updates this to contain the provided AskOrder and returns itself.
func (o *Order) WithAsk(askOrder *AskOrder) *Order {
	o.Order = &Order_AskOrder{AskOrder: askOrder}
	return o
}

// WithBid updates this to contain the provided BidOrder and returns itself.
func (o *Order) WithBid(bidOrder *BidOrder) *Order {
	o.Order = &Order_BidOrder{BidOrder: bidOrder}
	return o
}

// IsAskOrder returns true if this order is an ask order.
func (o Order) IsAskOrder() bool {
	return o.GetAskOrder() != nil
}

// IsBidOrder returns true if this order is a bid order.
func (o Order) IsBidOrder() bool {
	return o.GetBidOrder() != nil
}

// GetOrderType returns a string indicating what type this order is.
// See: OrderTypeAsk, OrderTypeBid
// Panics if the order details are not set or are something unexpected.
func (o Order) GetOrderType() string {
	switch v := o.GetOrder().(type) {
	case *Order_AskOrder:
		return OrderTypeAsk
	case *Order_BidOrder:
		return OrderTypeBid
	default:
		// If GetOrderType() is called without the order being set yet, it's a programming error, so panic.
		// If it's a type without a case, the case needs to be added, so panic.
		panic(fmt.Sprintf("GetOrderType() missing case for %T", v))
	}
}

// GetOrderTypeByte returns the type byte for this order.
// See: OrderTypeByteAsk, OrderTypeByteBid
// Panics if the order details are not set or are something unexpected.
func (o Order) GetOrderTypeByte() byte {
	switch v := o.GetOrder().(type) {
	case *Order_AskOrder:
		return OrderTypeByteAsk
	case *Order_BidOrder:
		return OrderTypeByteBid
	default:
		// If GetOrderTypeByte() is called without the order being set yet, it's a programming error, so panic.
		// If it's a type without a case, the case needs to be added, so panic.
		panic(fmt.Sprintf("GetOrderTypeByte() missing case for %T", v))
	}
}

// GetMarketID returns the market id for this order.
// Panics if the order details are not set or are something unexpected.
func (o Order) GetMarketID() uint32 {
	switch v := o.GetOrder().(type) {
	case *Order_AskOrder:
		return v.AskOrder.MarketId
	case *Order_BidOrder:
		return v.BidOrder.MarketId
	default:
		// If GetMarketID() is called without the order being set yet, it's a programming error, so panic.
		// If it's a type without a case, the case needs to be added, so panic.
		panic(fmt.Sprintf("GetMarketID() missing case for %T", v))
	}
}

// GetOwner gets the address of the owner of this order.
// E.g. the seller for ask orders, or buyer for bid orders.
func (o Order) GetOwner() string {
	switch v := o.GetOrder().(type) {
	case *Order_AskOrder:
		return v.AskOrder.Seller
	case *Order_BidOrder:
		return v.BidOrder.Buyer
	default:
		// If GetOwner() is called without the order being set yet, it's a programming error, so panic.
		// If it's a type without a case, the case needs to be added, so panic.
		panic(fmt.Sprintf("GetOwner() missing case for %T", v))
	}
}

// GetAssets gets the assets in this order.
func (o Order) GetAssets() sdk.Coins {
	switch v := o.GetOrder().(type) {
	case *Order_AskOrder:
		return v.AskOrder.Assets
	case *Order_BidOrder:
		return v.BidOrder.Assets
	default:
		// If GetAssets() is called without the order being set yet, it's a programming error, so panic.
		// If it's a type without a case, the case needs to be added, so panic.
		panic(fmt.Sprintf("GetAssets() missing case for %T", v))
	}
}

// GetPrice gets the price in this order.
func (o Order) GetPrice() sdk.Coin {
	switch v := o.GetOrder().(type) {
	case *Order_AskOrder:
		return v.AskOrder.Price
	case *Order_BidOrder:
		return v.BidOrder.Price
	default:
		// If GetPrice() is called without the order being set yet, it's a programming error, so panic.
		// If it's a type without a case, the case needs to be added, so panic.
		panic(fmt.Sprintf("GetPrice() missing case for %T", v))
	}
}

// GetHoldAmount returns the total amount that should be on hold for this order.
func (o Order) GetHoldAmount() sdk.Coins {
	switch v := o.GetOrder().(type) {
	case *Order_AskOrder:
		return v.AskOrder.GetHoldAmount()
	case *Order_BidOrder:
		return v.BidOrder.GetHoldAmount()
	default:
		// If HoldSettlementFee() is called without the order being set yet, it's a programming error, so panic.
		// If it's a type without a case, the case needs to be added, so panic.
		panic(fmt.Sprintf("GetHoldAmount() missing case for %T", v))
	}
}

// Validate returns an error if anything in this order is invalid.
func (o Order) Validate() error {
	if o.OrderId == 0 {
		return errors.New("invalid order id: must not be zero")
	}
	switch v := o.GetOrder().(type) {
	case *Order_AskOrder:
		return v.AskOrder.Validate()
	case *Order_BidOrder:
		return v.BidOrder.Validate()
	default:
		return fmt.Errorf("unknown order type %T", v)
	}
}

// GetHoldAmount gets the amount to put on hold for this ask order.
func (a AskOrder) GetHoldAmount() sdk.Coins {
	rv := a.Assets
	if a.SellerSettlementFlatFee != nil && a.SellerSettlementFlatFee.Denom != a.Price.Denom {
		rv = rv.Add(*a.SellerSettlementFlatFee)
	}
	return rv
}

// Validate returns an error if anything in this ask order is invalid.
func (a AskOrder) Validate() error {
	var errs []error

	// The market id must be provided.
	if a.MarketId == 0 {
		errs = append(errs, errors.New("invalid market id: must not be zero"))
	}

	// The seller address must be valid and not empty.
	if _, err := sdk.AccAddressFromBech32(a.Seller); err != nil {
		errs = append(errs, fmt.Errorf("invalid seller: %w", err))
	}

	// The price must not be zero and must be a valid coin.
	// The Coin.Validate() method allows the coin to be zero (but not negative).
	var priceDenom string
	if err := a.Price.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("invalid price: %w", err))
	} else if a.Price.IsZero() {
		errs = append(errs, errors.New("invalid price: cannot be zero"))
	} else {
		priceDenom = a.Price.Denom
	}

	// The Coins.Validate method does NOT allow any coin entries to be zero (or negative).
	// It does allow there to not be any entries, though, which we don't want to allow here.
	// We also don't want to allow the price denom to also be in the assets.
	if err := a.Assets.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("invalid assets: %w", err))
	} else {
		switch {
		case len(a.Assets) == 0:
			errs = append(errs, errors.New("invalid assets: must not be empty"))
		case len(priceDenom) > 0:
			for _, asset := range a.Assets {
				if priceDenom == asset.Denom {
					errs = append(errs, fmt.Errorf("invalid assets: cannot contain price denom %s", priceDenom))
					break
				}
			}
		}
	}

	if a.SellerSettlementFlatFee != nil {
		// Again, a Coin can be zero according to Validate.
		// A seller settlement flat fee is optional, but if they provided something,
		// it must have a positive (non-zero) value.
		if err := a.SellerSettlementFlatFee.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid seller settlement flat fee: %w", err))
		} else if a.SellerSettlementFlatFee.IsZero() {
			errs = append(errs, fmt.Errorf("invalid seller settlement flat fee: %s amount cannot be zero", a.SellerSettlementFlatFee.Denom))
		}
	}

	// Nothing to check on the AllowPartial boolean.

	return errors.Join(errs...)
}

// GetHoldAmount gets the amount to put on hold for this ask order.
func (b BidOrder) GetHoldAmount() sdk.Coins {
	return b.BuyerSettlementFees.Add(b.Price)
}

// Validate returns an error if anything in this ask order is invalid.
func (b BidOrder) Validate() error {
	var errs []error

	// The market id must be provided.
	if b.MarketId == 0 {
		errs = append(errs, errors.New("invalid market id: must not be zero"))
	}

	// The seller address must be valid and not empty.
	if _, err := sdk.AccAddressFromBech32(b.Buyer); err != nil {
		errs = append(errs, fmt.Errorf("invalid buyer: %w", err))
	}

	// The price must not be zero and must be a valid coin.
	// The Coin.Validate() method allows the coin to be zero (but not negative).
	var priceDenom string
	if err := b.Price.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("invalid price: %w", err))
	} else if b.Price.IsZero() {
		errs = append(errs, errors.New("invalid price: cannot be zero"))
	} else {
		priceDenom = b.Price.Denom
	}

	// The Coins.Validate method does NOT allow any coin entries to be zero (or negative).
	// It does allow there to not be any entries, though, which we don't want to allow here.
	// We also don't want to allow the price denom to also be in the assets.
	if err := b.Assets.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("invalid assets: %w", err))
	} else {
		switch {
		case len(b.Assets) == 0:
			errs = append(errs, errors.New("invalid assets: must not be empty"))
		case len(priceDenom) > 0:
			for _, asset := range b.Assets {
				if priceDenom == asset.Denom {
					errs = append(errs, fmt.Errorf("invalid assets: cannot contain price denom %s", priceDenom))
					break
				}
			}
		}
	}

	if len(b.BuyerSettlementFees) > 0 {
		// If there are buyer settlement fees, they must all be valid and positive (non-zero).
		if err := b.BuyerSettlementFees.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid buyer settlement fees: %w", err))
		}
	}

	// Nothing to check on the AllowPartial boolean.

	return errors.Join(errs...)
}

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
				order.OrderId, assets, f.Order.OrderId, askOrder.Assets)
		}
		// Note, we truncate the division here.
		// Later, after all fulfillments have been identified, we will handle the remainders.
		priceAmt := bidOrder.Price.Amount.Mul(askOrder.Assets[0].Amount).Quo(assets[0].Amount)
		split.Price = sdk.NewCoin(bidOrder.Price.Denom, priceAmt)
	}

	f.Splits = append(f.Splits, split)
	f.AssetsLeft = newAssetsLeft
	f.TotalPrice = f.TotalPrice.Add(split.Price)
	return nil
}
