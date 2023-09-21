package exchange

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Define the type strings and bytes to use for each order type.
// These must all be unique. If you add entries, be sure to update TestOrderTypesAndBytes too.
const (
	OrderTypeAsk     = "ask"
	OrderTypeByteAsk = byte(0x00)

	OrderTypeBid     = "bid"
	OrderTypeByteBid = byte(0x01)
)

// SubOrderI is an interface with getters for the fields in a sub-order (i.e. AskOrder or BidOrder).
type SubOrderI interface {
	GetMarketID() uint32
	GetOwner() string
	GetAssets() sdk.Coins
	GetPrice() sdk.Coin
	GetSettlementFees() sdk.Coins
	PartialFillAllowed() bool
	GetOrderType() string
	GetOrderTypeByte() byte
	GetHoldAmount() sdk.Coins
	Validate() error
}

var (
	_ SubOrderI = (*AskOrder)(nil)
	_ SubOrderI = (*BidOrder)(nil)
)

// OrderI is an interface with getters for all the fields associated with an order and it's sub-order.
type OrderI interface {
	SubOrderI
	GetOrderID() uint64
	IsAskOrder() bool
	IsBidOrder() bool
}

var (
	_ OrderI = (*Order)(nil)
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

// GetOrderID gets the numerical identifier for this order.
func (o Order) GetOrderID() uint64 {
	return o.OrderId
}

// GetSubOrder gets this order's sub-order as a SubOrderI.
func (o Order) GetSubOrder() (SubOrderI, error) {
	switch v := o.GetOrder().(type) {
	case *Order_AskOrder:
		return v.AskOrder, nil
	case *Order_BidOrder:
		return v.BidOrder, nil
	default:
		// If this is called without the sub-order being set yet, it's a programming error, so panic.
		// If it's a type that doesn't implement SubOrderI, that needs to be done, so panic.
		return nil, fmt.Errorf("unknown sub-order type %T: does not implement SubOrderI", v)
	}
}

// MustGetSubOrder gets this order's sub-order as a SubOrderI.
// Panics if the sub-order is not set or is something unexpected.
func (o Order) MustGetSubOrder() SubOrderI {
	rv, err := o.GetSubOrder()
	if err != nil {
		panic(err)
	}
	return rv
}

// GetMarketID returns the market id for this order.
// Panics if the sub-order is not set or is something unexpected.
func (o Order) GetMarketID() uint32 {
	return o.MustGetSubOrder().GetMarketID()
}

// GetOwner returns the owner of this order.
// E.g. the seller for ask orders, or buyer for bid orders.
// Panics if the sub-order is not set or is something unexpected.
func (o Order) GetOwner() string {
	return o.MustGetSubOrder().GetOwner()
}

// GetAssets returns the assets for this order.
// Panics if the sub-order is not set or is something unexpected.
func (o Order) GetAssets() sdk.Coins {
	return o.MustGetSubOrder().GetAssets()
}

// GetPrice returns the price for this order.
// Panics if the sub-order is not set or is something unexpected.
func (o Order) GetPrice() sdk.Coin {
	return o.MustGetSubOrder().GetPrice()
}

// GetSettlementFees returns the settlement fees in this order.
func (o Order) GetSettlementFees() sdk.Coins {
	return o.MustGetSubOrder().GetSettlementFees()
}

// PartialFillAllowed returns true if this order allows partial fulfillment.
func (o Order) PartialFillAllowed() bool {
	return o.MustGetSubOrder().PartialFillAllowed()
}

// GetOrderType returns a string indicating what type this order is.
// E.g: OrderTypeAsk or OrderTypeBid
func (o Order) GetOrderType() string {
	so, err := o.GetSubOrder()
	if err != nil {
		return fmt.Sprintf("%T", o.Order)
	}
	return so.GetOrderType()
}

// GetOrderTypeByte returns the type byte for this order.
// E.g: OrderTypeByteAsk or OrderTypeByteBid
// Panics if the sub-order is not set or is something unexpected.
func (o Order) GetOrderTypeByte() byte {
	return o.MustGetSubOrder().GetOrderTypeByte()
}

// GetHoldAmount returns the amount that should be on hold for this order.
// Panics if the sub-order is not set or is something unexpected.
func (o Order) GetHoldAmount() sdk.Coins {
	return o.MustGetSubOrder().GetHoldAmount()
}

// Validate returns an error if anything in this order is invalid.
func (o Order) Validate() error {
	if o.OrderId == 0 {
		return errors.New("invalid order id: must not be zero")
	}
	so, err := o.GetSubOrder()
	if err != nil {
		return err
	}
	return so.Validate()
}

// GetMarketID returns the market id for this ask order.
func (a AskOrder) GetMarketID() uint32 {
	return a.MarketId
}

// GetOwner returns the owner of this ask order: the seller.
func (a AskOrder) GetOwner() string {
	return a.Seller
}

// GetAssets returns the assets for sale with this ask order.
func (a AskOrder) GetAssets() sdk.Coins {
	return a.Assets
}

// GetPrice returns the minimum price to accept for this ask order.
func (a AskOrder) GetPrice() sdk.Coin {
	return a.Price
}

// GetSettlementFees returns the seller settlement flat fees in this ask order.
func (a AskOrder) GetSettlementFees() sdk.Coins {
	if a.SellerSettlementFlatFee == nil {
		return nil
	}
	return sdk.Coins{*a.SellerSettlementFlatFee}
}

// PartialFillAllowed returns true if this ask order allows partial fulfillment.
func (a AskOrder) PartialFillAllowed() bool {
	return a.AllowPartial
}

// GetOrderType returns the order type string for this ask order: "ask".
func (a AskOrder) GetOrderType() string {
	return OrderTypeAsk
}

// GetOrderTypeByte returns the order type byte for this bid order: 0x00.
func (a AskOrder) GetOrderTypeByte() byte {
	return OrderTypeByteAsk
}

// GetHoldAmount returns the amount that should be on hold for this ask order.
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

// GetMarketID returns the market id for this bid order.
func (b BidOrder) GetMarketID() uint32 {
	return b.MarketId
}

// GetOwner returns the owner of this bid order: the buyer.
func (b BidOrder) GetOwner() string {
	return b.Buyer
}

// GetAssets returns the assets to buy for this bid order.
func (b BidOrder) GetAssets() sdk.Coins {
	return b.Assets
}

// GetPrice returns the price to pay for this bid order.
func (b BidOrder) GetPrice() sdk.Coin {
	return b.Price
}

// GetSettlementFees returns the buyer settlement fees in this bid order.
func (b BidOrder) GetSettlementFees() sdk.Coins {
	return b.BuyerSettlementFees
}

// PartialFillAllowed returns true if this bid order allows partial fulfillment.
func (b BidOrder) PartialFillAllowed() bool {
	return b.AllowPartial
}

// GetOrderType returns the order type string for this bid order: "bid".
func (b BidOrder) GetOrderType() string {
	return OrderTypeBid
}

// GetOrderTypeByte returns the order type byte for this bid order: 0x01.
func (b BidOrder) GetOrderTypeByte() byte {
	return OrderTypeByteBid
}

// GetHoldAmount returns the amount that should be on hold for this bid order.
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
