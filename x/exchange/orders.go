package exchange

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"

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

// MaxExternalIDLength is the maximum length that an external id can have.
// A 32 byte address as a bech32 string is 59 characters + the hrp.
// E.g. a 32 byte address with hrp "pb" will be 61 characters long.
// Technically, a bech32 HRP can be 1 to 83 characters. This 100 was chosen as  a balance meant
// to allow most of those while still limiting the length of keys that use these external ids.
const MaxExternalIDLength = 100

// SubOrderI is an interface with getters for the fields in a sub-order (i.e. AskOrder or BidOrder).
type SubOrderI interface {
	GetMarketID() uint32
	GetOwner() string
	GetAssets() sdk.Coin
	GetPrice() sdk.Coin
	GetSettlementFees() sdk.Coins
	PartialFillAllowed() bool
	GetExternalID() string
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

var _ OrderI = (*Order)(nil)

// findDuplicateIds returns all order ids that appear two or more times in the provided slice.
func findDuplicateIds(orderIDs []uint64) []uint64 {
	var rv []uint64
	seen := make(map[uint64]bool, len(orderIDs))
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

// validateCoin makes sure the provided coin is valid an not zero.
func validateCoin(field string, coin sdk.Coin) error {
	// The Coin.Validate() method allows the coin to be zero (but not negative).
	if err := coin.Validate(); err != nil {
		return fmt.Errorf("invalid %s: %w", field, err)
	} else if coin.IsZero() {
		return fmt.Errorf("invalid %s: cannot be zero", field)
	}
	return nil
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

// ValidateExternalID makes sure an external id is okay.
func ValidateExternalID(externalID string) error {
	if len(externalID) > MaxExternalIDLength {
		// TODO[1703]: Update unit tests that broke because I just changed the ValidateExternalID error message.
		return fmt.Errorf("invalid external id %q (length %d): max length %d",
			externalID[:5]+"..."+externalID[len(externalID)-5:], len(externalID), MaxExternalIDLength)
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

// GetOrderID gets the numerical identifier for this order.
func (o Order) GetOrderID() uint64 {
	return o.OrderId
}

// IsAskOrder returns true if this order is an ask order.
func (o Order) IsAskOrder() bool {
	return o.GetAskOrder() != nil
}

// IsBidOrder returns true if this order is a bid order.
func (o Order) IsBidOrder() bool {
	return o.GetBidOrder() != nil
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
		return nil, fmt.Errorf("order %d has unknown sub-order type %T: does not implement SubOrderI", o.OrderId, v)
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
func (o Order) GetAssets() sdk.Coin {
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

// GetUUID returns this order's UUID.
func (o Order) GetExternalID() string {
	return o.MustGetSubOrder().GetExternalID()
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
		return errors.New("invalid order id: cannot be zero")
	}
	so, err := o.GetSubOrder()
	if err != nil {
		return err
	}
	return so.Validate()
}

// Split splits this order by the provided assets filled.
func (o Order) Split(assetsFilledAmt sdkmath.Int) (filled *Order, unfilled *Order, err error) {
	orderAssets := o.GetAssets()
	orderAssetsAmt := orderAssets.Amount
	assetsFilled := sdk.Coin{Denom: orderAssets.Denom, Amount: assetsFilledAmt}

	switch {
	case !assetsFilledAmt.IsPositive():
		return nil, nil, fmt.Errorf("cannot split %s order %d having asset %q at %q: amount filled not positive",
			o.GetOrderType(), o.OrderId, orderAssets, assetsFilled)
	case assetsFilledAmt.Equal(orderAssetsAmt):
		return nil, nil, fmt.Errorf("cannot split %s order %d having asset %q at %q: amount filled equals order assets",
			o.GetOrderType(), o.OrderId, orderAssets, assetsFilled)
	case assetsFilledAmt.GT(orderAssetsAmt):
		return nil, nil, fmt.Errorf("cannot split %s order %d having asset %q at %q: overfilled",
			o.GetOrderType(), o.OrderId, orderAssets, assetsFilled)
	case !o.PartialFillAllowed():
		return nil, nil, fmt.Errorf("cannot split %s order %d having assets %q at %q: order does not allow partial fulfillment",
			o.GetOrderType(), o.OrderId, orderAssets, assetsFilled)
	}

	orderPrice := o.GetPrice()
	priceFilledAmt, priceRem := QuoRemInt(orderPrice.Amount.Mul(assetsFilledAmt), orderAssetsAmt)
	if !priceRem.IsZero() {
		return nil, nil, fmt.Errorf("%s order %d having assets %q cannot be partially filled by %q: "+
			"price %q is not evenly divisible",
			o.GetOrderType(), o.OrderId, orderAssets, assetsFilled, orderPrice)
	}

	orderFees := o.GetSettlementFees()
	var feesFilled, feesUnfilled sdk.Coins
	if !orderFees.IsZero() {
		for _, orderFee := range orderFees {
			feeFilled, feeRem := QuoRemInt(orderFee.Amount.Mul(assetsFilledAmt), orderAssetsAmt)
			if !feeRem.IsZero() {
				return nil, nil, fmt.Errorf("%s order %d having assets %q cannot be partially filled by %q: "+
					"fee %q is not evenly divisible",
					o.GetOrderType(), o.OrderId, orderAssets, assetsFilled, orderFee)
			}
			feesFilled = feesFilled.Add(sdk.NewCoin(orderFee.Denom, feeFilled))
		}
		feesUnfilled = orderFees.Sub(feesFilled...)
		if feesFilled.IsZero() {
			feesFilled = nil
		}
		if feesUnfilled.IsZero() {
			feesUnfilled = nil
		}
	}

	assetsUnfilled := sdk.NewCoin(orderAssets.Denom, orderAssetsAmt.Sub(assetsFilledAmt))
	priceFilled := sdk.NewCoin(orderPrice.Denom, priceFilledAmt)
	priceUnfilled := sdk.NewCoin(orderPrice.Denom, orderPrice.Amount.Sub(priceFilledAmt))

	switch v := o.Order.(type) {
	case *Order_AskOrder:
		var feeFilled, feeUnfilled *sdk.Coin
		if !feesFilled.IsZero() {
			feeFilled = &feesFilled[0]
		}
		if !feesUnfilled.IsZero() {
			feeUnfilled = &feesUnfilled[0]
		}

		filled = NewOrder(o.OrderId).WithAsk(v.AskOrder.CopyChange(assetsFilled, priceFilled, feeFilled))
		unfilled = NewOrder(o.OrderId).WithAsk(v.AskOrder.CopyChange(assetsUnfilled, priceUnfilled, feeUnfilled))
		return filled, unfilled, nil
	case *Order_BidOrder:
		filled = NewOrder(o.OrderId).WithBid(v.BidOrder.CopyChange(assetsFilled, priceFilled, feesFilled))
		unfilled = NewOrder(o.OrderId).WithBid(v.BidOrder.CopyChange(assetsUnfilled, priceUnfilled, feesUnfilled))
		return filled, unfilled, nil
	default:
		// This is here in case a new order type is added (that implements OrderI), but a case isn't added here.
		panic(fmt.Errorf("cannot split %s order %d: unknown order type", o.GetOrderType(), o.OrderId))
	}
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
func (a AskOrder) GetAssets() sdk.Coin {
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

// GetExternalID returns this ask order's external id.
func (a AskOrder) GetExternalID() string {
	return a.ExternalId
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
	rv := sdk.Coins{a.Assets}
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
		errs = append(errs, errors.New("invalid market id: cannot be zero"))
	}

	// The seller address must be valid and not empty.
	if _, err := sdk.AccAddressFromBech32(a.Seller); err != nil {
		errs = append(errs, fmt.Errorf("invalid seller: %w", err))
	}

	var priceDenom string
	if err := validateCoin("price", a.Price); err != nil {
		errs = append(errs, err)
	} else {
		priceDenom = a.Price.Denom
	}

	if err := validateCoin("assets", a.Assets); err != nil {
		errs = append(errs, err)
	} else if len(priceDenom) > 0 && a.Assets.Denom == priceDenom {
		errs = append(errs, fmt.Errorf("invalid assets: price denom %s cannot also be the assets denom", priceDenom))
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

	if err := ValidateExternalID(a.ExternalId); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// CopyChange creates a copy of this ask order with the provided assets, price and fee.
func (a AskOrder) CopyChange(newAssets, newPrice sdk.Coin, newFee *sdk.Coin) *AskOrder {
	return &AskOrder{
		MarketId:                a.MarketId,
		Seller:                  a.Seller,
		Assets:                  newAssets,
		Price:                   newPrice,
		SellerSettlementFlatFee: newFee,
		AllowPartial:            a.AllowPartial,
		ExternalId:              a.ExternalId,
	}
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
func (b BidOrder) GetAssets() sdk.Coin {
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

// GetExternalID returns this bid order's external id.
func (b BidOrder) GetExternalID() string {
	return b.ExternalId
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
		errs = append(errs, errors.New("invalid market id: cannot be zero"))
	}

	// The buyer address must be valid and not empty.
	if _, err := sdk.AccAddressFromBech32(b.Buyer); err != nil {
		errs = append(errs, fmt.Errorf("invalid buyer: %w", err))
	}

	var priceDenom string
	if err := validateCoin("price", b.Price); err != nil {
		errs = append(errs, err)
	} else {
		priceDenom = b.Price.Denom
	}

	if err := validateCoin("assets", b.Assets); err != nil {
		errs = append(errs, err)
	} else if len(priceDenom) > 0 && b.Assets.Denom == priceDenom {
		errs = append(errs, fmt.Errorf("invalid assets: price denom %s cannot also be the assets denom", priceDenom))
	}

	if len(b.BuyerSettlementFees) > 0 {
		// If there are buyer settlement fees, they must all be valid and positive (non-zero).
		if err := b.BuyerSettlementFees.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("invalid buyer settlement fees: %w", err))
		}
	}

	// Nothing to check on the AllowPartial boolean.

	if err := ValidateExternalID(b.ExternalId); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// CopyChange creates a copy of this bid order with the provided assets, price and fees.
func (b BidOrder) CopyChange(newAssets, newPrice sdk.Coin, newFees sdk.Coins) *BidOrder {
	return &BidOrder{
		MarketId:            b.MarketId,
		Buyer:               b.Buyer,
		Assets:              newAssets,
		Price:               newPrice,
		BuyerSettlementFees: newFees,
		AllowPartial:        b.AllowPartial,
		ExternalId:          b.ExternalId,
	}
}

// FilledOrder holds an order that has been filled (either in full or partially).
// The GetPrice() and GetSettlementFees() methods indicate the actual amounts involved in the fulfillment.
type FilledOrder struct {
	order       *Order
	actualPrice sdk.Coin
	actualFees  sdk.Coins
}

var _ OrderI = (*FilledOrder)(nil)

func NewFilledOrder(order *Order, actualPrice sdk.Coin, actualFees sdk.Coins) *FilledOrder {
	return &FilledOrder{
		order:       order,
		actualPrice: actualPrice,
		actualFees:  actualFees,
	}
}

// GetOriginalOrder gets the original order that this filled order represents.
func (o FilledOrder) GetOriginalOrder() *Order {
	return o.order
}

// GetOrderID gets the numerical identifier for this order.
func (o FilledOrder) GetOrderID() uint64 {
	return o.order.GetOrderID()
}

// IsAskOrder returns true if this order is an ask order.
func (o FilledOrder) IsAskOrder() bool {
	return o.order.IsAskOrder()
}

// IsBidOrder returns true if this order is a bid order.
func (o FilledOrder) IsBidOrder() bool {
	return o.order.IsBidOrder()
}

// GetMarketID returns the market id for this order.
func (o FilledOrder) GetMarketID() uint32 {
	return o.order.GetMarketID()
}

// GetOwner returns the owner of this order.
// E.g. the seller for ask orders, or buyer for bid orders.
func (o FilledOrder) GetOwner() string {
	return o.order.GetOwner()
}

// GetAssets returns the assets for this order.
func (o FilledOrder) GetAssets() sdk.Coin {
	return o.order.GetAssets()
}

// GetPrice returns the actual price involved in this order fulfillment.
func (o FilledOrder) GetPrice() sdk.Coin {
	return o.actualPrice
}

// GetOriginalPrice gets the original price of this order.
func (o FilledOrder) GetOriginalPrice() sdk.Coin {
	return o.order.GetPrice()
}

// GetSettlementFees returns the actual settlement fees involved in this order fulfillment.
func (o FilledOrder) GetSettlementFees() sdk.Coins {
	return o.actualFees
}

// GetOriginalSettlementFees gets the original settlement fees of this order.
func (o FilledOrder) GetOriginalSettlementFees() sdk.Coins {
	return o.order.GetSettlementFees()
}

// PartialFillAllowed returns true if this order allows partial fulfillment.
func (o FilledOrder) PartialFillAllowed() bool {
	return o.order.PartialFillAllowed()
}

// GetExternalID returns this order's external id.
func (o FilledOrder) GetExternalID() string {
	return o.order.GetExternalID()
}

// GetOrderType returns a string indicating what type this order is.
// E.g: OrderTypeAsk or OrderTypeBid
func (o FilledOrder) GetOrderType() string {
	return o.order.GetOrderType()
}

// GetOrderTypeByte returns the type byte for this order.
// E.g: OrderTypeByteAsk or OrderTypeByteBid
func (o FilledOrder) GetOrderTypeByte() byte {
	return o.order.GetOrderTypeByte()
}

// GetHoldAmount returns the amount that should be on hold for this order.
func (o FilledOrder) GetHoldAmount() sdk.Coins {
	return o.order.GetHoldAmount()
}

// Validate returns nil (because it's assumed that the order was validated long ago).
// This is just here to fulfill the OrderI interface.
func (o FilledOrder) Validate() error {
	return nil
}
