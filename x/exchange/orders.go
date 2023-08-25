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
	OrderTypeByteAsk = 0x00

	OrderTypeBid     = "bid"
	OrderTypeByteBid = 0x01
)

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

// OrderType returns a string indicating what type this order is.
// See: OrderTypeAsk, OrderTypeBid
// Panics if the order details are not set or are something unexpected.
func (o Order) OrderType() string {
	switch v := o.GetOrder().(type) {
	case *Order_AskOrder:
		return OrderTypeAsk
	case *Order_BidOrder:
		return OrderTypeBid
	default:
		// If OrderType() is called without the order being set yet, it's a programming error, so panic.
		// If it's a type without a case, the case needs to be added, so panic.
		panic(fmt.Sprintf("OrderType() missing case for %T", v))
	}
}

// OrderTypeByte returns the type byte for this order.
// See: OrderTypeByteAsk, OrderTypeByteBid
// Panics if the order details are not set or are something unexpected.
func (o Order) OrderTypeByte() byte {
	switch v := o.GetOrder().(type) {
	case *Order_AskOrder:
		return OrderTypeByteAsk
	case *Order_BidOrder:
		return OrderTypeByteBid
	default:
		// If OrderType() is called without the order being set yet, it's a programming error, so panic.
		// If it's a type without a case, the case needs to be added, so panic.
		panic(fmt.Sprintf("OrderTypeByte() missing case for %T", v))
	}
}

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
	if a.Price.IsZero() {
		errs = append(errs, errors.New("invalid price: cannot be zero"))
	} else {
		if err := a.Price.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid price: %w", err))
		} else {
			priceDenom = a.Price.Denom
		}
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

// Validate returns an error if anything in this ask order is invalid.
func (a BidOrder) Validate() error {
	var errs []error

	// The market id must be provided.
	if a.MarketId == 0 {
		errs = append(errs, errors.New("invalid market id: must not be zero"))
	}

	// The seller address must be valid and not empty.
	if _, err := sdk.AccAddressFromBech32(a.Buyer); err != nil {
		errs = append(errs, fmt.Errorf("invalid buyer: %w", err))
	}

	// The price must not be zero and must be a valid coin.
	// The Coin.Validate() method allows the coin to be zero (but not negative).
	var priceDenom string
	if a.Price.IsZero() {
		errs = append(errs, errors.New("invalid price: cannot be zero"))
	} else {
		if err := a.Price.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid price: %w", err))
		} else {
			priceDenom = a.Price.Denom
		}
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

	if len(a.BuyerSettlementFees) > 0 {
		// If there are buyer settlement fees, they must all be valid and positive (non-zero).
		if err := a.BuyerSettlementFees.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid buyer settlement fees: %w", err))
		}
	}

	// Nothing to check on the AllowPartial boolean.

	return errors.Join(errs...)
}
