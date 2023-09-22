package exchange

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil/assertions"
)

// copyCoins creates a copy of the provided coins slice with copies of each entry.
func copyCoins(coins sdk.Coins) sdk.Coins {
	if coins == nil {
		return nil
	}
	rv := make(sdk.Coins, len(coins))
	for i, coin := range coins {
		rv[i] = copyCoin(coin)
	}
	return rv
}

// copyCoin returns a copy of the provided coin.
func copyCoin(coin sdk.Coin) sdk.Coin {
	return sdk.NewInt64Coin(coin.Denom, coin.Amount.Int64())
}

// copyCoinP returns a copy of the provided *coin.
func copyCoinP(coin *sdk.Coin) *sdk.Coin {
	if coin == nil {
		return nil
	}
	rv := copyCoin(*coin)
	return &rv
}

func TestOrderTypesAndBytes(t *testing.T) {
	values := []struct {
		name string
		str  string
		b    byte
	}{
		{name: "Ask", str: OrderTypeAsk, b: OrderTypeByteAsk},
		{name: "Bid", str: OrderTypeBid, b: OrderTypeByteBid},
	}

	ot := func(name string) string {
		return "OrderType" + name
	}
	otb := func(name string) string {
		return "OrderTypeByte" + name
	}

	knownTypeStrings := make(map[string]string)
	knownTypeBytes := make(map[byte]string)
	var errs []error

	for _, value := range values {
		strKnownBy, strKnown := knownTypeStrings[value.str]
		if strKnown {
			errs = append(errs, fmt.Errorf("type string %q used by both %s and %s", value.str, ot(strKnownBy), ot(value.name)))
		}
		knownTypeStrings[value.str] = value.name

		bKnownBy, bKnown := knownTypeBytes[value.b]
		if bKnown {
			errs = append(errs, fmt.Errorf("type byte %#X used by both %s and %s", value.b, otb(bKnownBy), otb(value.name)))
		}
		knownTypeBytes[value.b] = value.name
	}

	err := errors.Join(errs...)
	assert.NoError(t, err, "checking for duplicate values")
}

func TestValidateOrderIDs(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		orderIDs []uint64
		expErr   string
	}{
		{
			name:     "control",
			field:    "testfieldname",
			orderIDs: []uint64{1, 18_446_744_073_709_551_615, 5, 65_536, 97},
			expErr:   "",
		},
		{
			name:     "nil slice",
			field:    "testfieldname",
			orderIDs: nil,
			expErr:   "no testfieldname order ids provided",
		},
		{
			name:     "empty slice",
			field:    "testfieldname",
			orderIDs: []uint64{},
			expErr:   "no testfieldname order ids provided",
		},
		{
			name:     "contains a zero",
			field:    "testfieldname",
			orderIDs: []uint64{1, 18_446_744_073_709_551_615, 0, 5, 65_536, 97},
			expErr:   "invalid testfieldname order ids: cannot contain order id zero",
		},
		{
			name:     "one duplicate entry",
			field:    "testfieldname",
			orderIDs: []uint64{1, 2, 3, 1, 4, 5},
			expErr:   "duplicate testfieldname order ids provided: [1]",
		},
		{
			name:     "three duplicate entries",
			field:    "testfieldname",
			orderIDs: []uint64{1, 2, 3, 1, 4, 5, 5, 6, 7, 3, 8, 9},
			expErr:   "duplicate testfieldname order ids provided: [1 5 3]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = ValidateOrderIDs(tc.field, tc.orderIDs)
			}
			require.NotPanics(t, testFunc, "ValidateOrderIDs(%q, %v)", tc.field, tc.orderIDs)
			assertions.AssertErrorValue(t, err, tc.expErr, "ValidateOrderIDs(%q, %v)", tc.field, tc.orderIDs)
		})
	}
}

func TestNewOrder(t *testing.T) {
	tests := []struct {
		name    string
		orderID uint64
	}{
		{name: "zero", orderID: 0},
		{name: "one", orderID: 1},
		{name: "two", orderID: 2},
		{name: "max uint8", orderID: 255},
		{name: "max uint16", orderID: 65_535},
		{name: "max uint32", orderID: 4_294_967_295},
		{name: "max uint64", orderID: 18_446_744_073_709_551_615},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			order := NewOrder(tc.orderID)
			assert.Equal(t, tc.orderID, order.OrderId, "order.OrderID")
			assert.Nil(t, order.Order, "order.Order")
		})
	}
}

func TestOrder_WithAsk(t *testing.T) {
	origOrderID := uint64(5)
	base := &Order{OrderId: origOrderID}
	origAsk := &AskOrder{
		MarketId:                12,
		Seller:                  "some seller",
		Assets:                  sdk.NewInt64Coin("water", 8),
		Price:                   sdk.NewInt64Coin("sand", 1),
		SellerSettlementFlatFee: &sdk.Coin{Denom: "banana", Amount: sdkmath.NewInt(3)},
		AllowPartial:            true,
	}
	ask := &AskOrder{
		MarketId:                origAsk.MarketId,
		Seller:                  origAsk.Seller,
		Assets:                  copyCoin(origAsk.Assets),
		Price:                   copyCoin(origAsk.Price),
		SellerSettlementFlatFee: copyCoinP(origAsk.SellerSettlementFlatFee),
		AllowPartial:            origAsk.AllowPartial,
	}

	var order *Order
	testFunc := func() {
		order = base.WithAsk(ask)
	}
	require.NotPanics(t, testFunc, "WithAsk")

	// Make sure the returned reference is the same as the base (receiver).
	assert.Same(t, base, order, "WithAsk result (actual) vs receiver (expected)")

	// Make sure the order id didn't change.
	assert.Equal(t, int(origOrderID), int(base.OrderId), "OrderId of result")

	// Make sure the AskOrder in the Order is the same reference as the one provided to WithAsk
	orderAsk := order.GetAskOrder()
	assert.Same(t, ask, orderAsk, "the ask in the resulting order (actual) vs what was provided (expected)")

	// Make sure nothing in the ask changed during WithAsk.
	assert.Equal(t, origAsk, orderAsk, "the ask in the resulting order (actual) vs what the ask was before being provided (expected)")
}

func TestOrder_WithBid(t *testing.T) {
	origOrderID := uint64(4)
	base := &Order{OrderId: origOrderID}
	origBid := &BidOrder{
		MarketId:            11,
		Buyer:               "some buyer",
		Assets:              sdk.NewInt64Coin("agua", 7),
		Price:               sdk.NewInt64Coin("dirt", 2),
		BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("grape", 3)),
		AllowPartial:        true,
	}
	bid := &BidOrder{
		MarketId:            origBid.MarketId,
		Buyer:               origBid.Buyer,
		Assets:              copyCoin(origBid.Assets),
		Price:               copyCoin(origBid.Price),
		BuyerSettlementFees: copyCoins(origBid.BuyerSettlementFees),
		AllowPartial:        origBid.AllowPartial,
	}

	var order *Order
	testFunc := func() {
		order = base.WithBid(bid)
	}
	require.NotPanics(t, testFunc, "WithBid")

	// Make sure the returned reference is the same as the base (receiver).
	assert.Same(t, base, order, "WithBid result (actual) vs receiver (expected)")

	// Make sure the order id didn't change.
	assert.Equal(t, int(origOrderID), int(base.OrderId), "OrderId of result")

	// Make sure the BidOrder in the Order is the same reference as the one provided to WithBid
	orderBid := order.GetBidOrder()
	assert.Same(t, bid, orderBid, "the bid in the resulting order (actual) vs what was provided (expected)")

	// Make sure nothing in the bid changed during WithBid.
	assert.Equal(t, origBid, orderBid, "the bid in the resulting order (actual) vs what the bid was before being provided (expected)")
}

// unknownOrderType is thing that implements isOrder_Order so it can be
// added to an Order, but isn't a type that there is anything defined for.
type unknownOrderType struct{}

var _ isOrder_Order = (*unknownOrderType)(nil)

func (o *unknownOrderType) isOrder_Order() {}
func (o *unknownOrderType) MarshalTo([]byte) (int, error) {
	return 0, nil
}
func (o *unknownOrderType) Size() int {
	return 0
}

func TestOrder_IsAskOrder(t *testing.T) {
	tests := []struct {
		name  string
		order *Order
		exp   bool
	}{
		{
			name:  "nil inside order",
			order: NewOrder(1),
			exp:   false,
		},
		{
			name:  "ask order",
			order: NewOrder(2).WithAsk(&AskOrder{}),
			exp:   true,
		},
		{
			name:  "bid order",
			order: NewOrder(3).WithBid(&BidOrder{}),
			exp:   false,
		},
		{
			name:  "unknown order type",
			order: &Order{OrderId: 4, Order: &unknownOrderType{}},
			exp:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = tc.order.IsAskOrder()
			}
			require.NotPanics(t, testFunc, "IsAskOrder()")
			assert.Equal(t, tc.exp, actual, "IsAskOrder() result")
		})
	}
}

func TestOrder_IsBidOrder(t *testing.T) {
	tests := []struct {
		name  string
		order *Order
		exp   bool
	}{
		{
			name:  "nil inside order",
			order: NewOrder(1),
			exp:   false,
		},
		{
			name:  "ask order",
			order: NewOrder(2).WithAsk(&AskOrder{}),
			exp:   false,
		},
		{
			name:  "bid order",
			order: NewOrder(3).WithBid(&BidOrder{}),
			exp:   true,
		},
		{
			name:  "unknown order type",
			order: &Order{OrderId: 4, Order: &unknownOrderType{}},
			exp:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = tc.order.IsBidOrder()
			}
			require.NotPanics(t, testFunc, "IsBidOrder()")
			assert.Equal(t, tc.exp, actual, "IsBidOrder() result")
		})
	}
}

func TestOrder_GetOrderID(t *testing.T) {
	tests := []struct {
		name  string
		order Order
		exp   uint64
	}{
		{name: "zero", order: Order{OrderId: 0}, exp: 0},
		{name: "one", order: Order{OrderId: 1}, exp: 1},
		{name: "twelve", order: Order{OrderId: 12}, exp: 12},
		{name: "max uint32 + 1", order: Order{OrderId: 4_294_967_296}, exp: 4_294_967_296},
		{name: "max uint64", order: Order{OrderId: 18_446_744_073_709_551_615}, exp: 18_446_744_073_709_551_615},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual uint64
			testFunc := func() {
				actual = tc.order.GetOrderID()
			}
			require.NotPanics(t, testFunc, "GetOrderID()")
			assert.Equal(t, tc.exp, actual, "GetOrderID() result")
		})
	}
}

const (
	nilSubTypeErr     = "unknown sub-order type <nil>: does not implement SubOrderI"
	unknownSubTypeErr = "unknown sub-order type *exchange.unknownOrderType: does not implement SubOrderI"
)

func TestOrder_GetSubOrder(t *testing.T) {
	askOrder := &AskOrder{
		MarketId:                1,
		Seller:                  sdk.AccAddress("Seller______________").String(),
		Assets:                  sdk.NewInt64Coin("assetcoin", 3),
		Price:                   sdk.NewInt64Coin("paycoin", 8),
		SellerSettlementFlatFee: &sdk.Coin{Denom: "feecoin", Amount: sdkmath.NewInt(1)},
		AllowPartial:            false,
	}
	bidOrder := &BidOrder{
		MarketId:            1,
		Buyer:               sdk.AccAddress("Buyer_______________").String(),
		Assets:              sdk.NewInt64Coin("assetcoin", 33),
		Price:               sdk.NewInt64Coin("paycoin", 88),
		BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("feecoin", 11)),
		AllowPartial:        true,
	}

	tests := []struct {
		name     string
		order    *Order
		expected SubOrderI
		expErr   string
	}{
		{
			name:     "AskOrder",
			order:    NewOrder(1).WithAsk(askOrder),
			expected: askOrder,
		},
		{
			name:     "BidOrder",
			order:    NewOrder(2).WithBid(bidOrder),
			expected: bidOrder,
		},
		{
			name:   "nil sub-order",
			order:  NewOrder(3),
			expErr: nilSubTypeErr,
		},
		{
			name:   "unknown order type",
			order:  &Order{OrderId: 4, Order: &unknownOrderType{}},
			expErr: unknownSubTypeErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var getActual SubOrderI
			var err error
			testGetSubOrder := func() {
				getActual, err = tc.order.GetSubOrder()
			}
			require.NotPanics(t, testGetSubOrder, "GetSubOrder()")
			assertions.AssertErrorValue(t, err, tc.expErr, "GetSubOrder() error")
			assert.Equal(t, tc.expected, getActual, "GetSubOrder() result")

			var mustActual SubOrderI
			testMustGetSubOrder := func() {
				mustActual = tc.order.MustGetSubOrder()
			}
			assertions.RequirePanicEquals(t, testMustGetSubOrder, tc.expErr, "MustGetSubOrder()")
			assert.Equal(t, tc.expected, mustActual, "MustGetSubOrder() result")
		})
	}
}

func TestOrder_GetMarketID(t *testing.T) {
	tests := []struct {
		name     string
		order    *Order
		expected uint32
		expPanic string
	}{
		{
			name:     "AskOrder",
			order:    NewOrder(1).WithAsk(&AskOrder{MarketId: 123}),
			expected: 123,
		},
		{
			name:     "BidOrder",
			order:    NewOrder(2).WithBid(&BidOrder{MarketId: 437}),
			expected: 437,
		},
		{
			name:     "nil inside order",
			order:    NewOrder(3),
			expPanic: nilSubTypeErr,
		},
		{
			name:     "unknown order type",
			order:    &Order{OrderId: 4, Order: &unknownOrderType{}},
			expPanic: unknownSubTypeErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual uint32
			testFunc := func() {
				actual = tc.order.GetMarketID()
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "GetMarketID()")
			assert.Equal(t, tc.expected, actual, "GetMarketID() result")
		})
	}
}

func TestOrder_GetOwner(t *testing.T) {
	tests := []struct {
		name     string
		order    *Order
		expected string
		expPanic string
	}{
		{
			name:     "AskOrder",
			order:    NewOrder(1).WithAsk(&AskOrder{Seller: "I'm a seller!"}),
			expected: "I'm a seller!",
		},
		{
			name:     "BidOrder",
			order:    NewOrder(2).WithBid(&BidOrder{Buyer: "gimmie gimmie"}),
			expected: "gimmie gimmie",
		},
		{
			name:     "nil inside order",
			order:    NewOrder(3),
			expPanic: nilSubTypeErr,
		},
		{
			name:     "unknown order type",
			order:    &Order{OrderId: 4, Order: &unknownOrderType{}},
			expPanic: unknownSubTypeErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var owner string
			testFunc := func() {
				owner = tc.order.GetOwner()
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "GetOwner()")
			assert.Equal(t, tc.expected, owner, "GetOwner() result")
		})
	}
}

func TestOrder_GetAssets(t *testing.T) {
	tests := []struct {
		name     string
		order    *Order
		expected sdk.Coin
		expPanic string
	}{
		{
			name:     "AskOrder",
			order:    NewOrder(1).WithAsk(&AskOrder{Assets: sdk.NewInt64Coin("acorns", 85)}),
			expected: sdk.NewInt64Coin("acorns", 85),
		},
		{
			name:     "BidOrder",
			order:    NewOrder(2).WithBid(&BidOrder{Assets: sdk.NewInt64Coin("boogers", 3)}),
			expected: sdk.NewInt64Coin("boogers", 3),
		},
		{
			name:     "nil inside order",
			order:    NewOrder(3),
			expPanic: nilSubTypeErr,
		},
		{
			name:     "unknown order type",
			order:    &Order{OrderId: 4, Order: &unknownOrderType{}},
			expPanic: unknownSubTypeErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var assets sdk.Coin
			testFunc := func() {
				assets = tc.order.GetAssets()
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "GetAssets()")
			assert.Equal(t, tc.expected.String(), assets.String(), "GetAssets() result")
		})
	}
}

func TestOrder_GetPrice(t *testing.T) {
	tests := []struct {
		name     string
		order    *Order
		expected sdk.Coin
		expPanic string
	}{
		{
			name:     "AskOrder",
			order:    NewOrder(1).WithAsk(&AskOrder{Price: sdk.NewInt64Coin("acorns", 85)}),
			expected: sdk.NewInt64Coin("acorns", 85),
		},
		{
			name:     "BidOrder",
			order:    NewOrder(2).WithBid(&BidOrder{Price: sdk.NewInt64Coin("boogers", 3)}),
			expected: sdk.NewInt64Coin("boogers", 3),
		},
		{
			name:     "nil inside order",
			order:    NewOrder(3),
			expPanic: nilSubTypeErr,
		},
		{
			name:     "unknown order type",
			order:    &Order{OrderId: 4, Order: &unknownOrderType{}},
			expPanic: unknownSubTypeErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var price sdk.Coin
			testFunc := func() {
				price = tc.order.GetPrice()
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "GetPrice()")
			assert.Equal(t, tc.expected.String(), price.String(), "GetPrice() result")
		})
	}
}

func TestOrder_GetSettlementFees(t *testing.T) {
	tests := []struct {
		name     string
		order    *Order
		expected sdk.Coins
		expPanic string
	}{
		{
			name:     "AskOrder",
			order:    NewOrder(1).WithAsk(&AskOrder{SellerSettlementFlatFee: &sdk.Coin{Denom: "askcoin", Amount: sdkmath.NewInt(3)}}),
			expected: sdk.NewCoins(sdk.NewInt64Coin("askcoin", 3)),
		},
		{
			name:     "BidOrder",
			order:    NewOrder(2).WithBid(&BidOrder{BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("bidcoin", 15))}),
			expected: sdk.NewCoins(sdk.NewInt64Coin("bidcoin", 15)),
		},
		{
			name:     "nil inside order",
			order:    NewOrder(3),
			expPanic: nilSubTypeErr,
		},
		{
			name:     "unknown order type",
			order:    &Order{OrderId: 4, Order: &unknownOrderType{}},
			expPanic: unknownSubTypeErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coins
			testFunc := func() {
				actual = tc.order.GetSettlementFees()
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "GetSettlementFees()")
			assert.Equal(t, tc.expected.String(), actual.String(), "GetSettlementFees() result")
		})
	}
}

func TestOrder_PartialFillAllowed(t *testing.T) {
	tests := []struct {
		name     string
		order    *Order
		expected bool
		expPanic string
	}{
		{
			name:     "AskOrder",
			order:    NewOrder(1).WithAsk(&AskOrder{AllowPartial: true}),
			expected: true,
		},
		{
			name:     "BidOrder",
			order:    NewOrder(2).WithBid(&BidOrder{AllowPartial: true}),
			expected: true,
		},
		{
			name:     "nil inside order",
			order:    NewOrder(3),
			expPanic: nilSubTypeErr,
		},
		{
			name:     "unknown order type",
			order:    &Order{OrderId: 4, Order: &unknownOrderType{}},
			expPanic: unknownSubTypeErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = tc.order.PartialFillAllowed()
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "PartialFillAllowed()")
			assert.Equal(t, tc.expected, actual, "PartialFillAllowed() result")
		})
	}
}

func TestOrder_GetOrderType(t *testing.T) {
	tests := []struct {
		name     string
		order    *Order
		expected string
	}{
		{
			name:     "AskOrder",
			order:    NewOrder(1).WithAsk(&AskOrder{}),
			expected: OrderTypeAsk,
		},
		{
			name:     "BidOrder",
			order:    NewOrder(2).WithBid(&BidOrder{}),
			expected: OrderTypeBid,
		},
		{
			name:     "nil inside order",
			order:    NewOrder(3),
			expected: "<nil>",
		},
		{
			name:     "unknown order type",
			order:    &Order{OrderId: 4, Order: &unknownOrderType{}},
			expected: "*exchange.unknownOrderType",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = tc.order.GetOrderType()
			}
			require.NotPanics(t, testFunc, "GetOrderType()")
			assert.Equal(t, tc.expected, actual, "GetOrderType() result")
		})
	}
}

func TestOrder_GetOrderTypeByte(t *testing.T) {
	tests := []struct {
		name     string
		order    *Order
		expected byte
		expPanic string
	}{
		{
			name:     "AskOrder",
			order:    NewOrder(1).WithAsk(&AskOrder{}),
			expected: OrderTypeByteAsk,
		},
		{
			name:     "BidOrder",
			order:    NewOrder(2).WithBid(&BidOrder{}),
			expected: OrderTypeByteBid,
		},
		{
			name:     "nil inside order",
			order:    NewOrder(3),
			expPanic: nilSubTypeErr,
		},
		{
			name:     "unknown order type",
			order:    &Order{OrderId: 4, Order: &unknownOrderType{}},
			expPanic: unknownSubTypeErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual byte
			testFunc := func() {
				actual = tc.order.GetOrderTypeByte()
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "GetOrderTypeByte()")
			assert.Equal(t, tc.expected, actual, "GetOrderTypeByte() result")
		})
	}
}

func TestOrder_GetHoldAmount(t *testing.T) {
	tests := []struct {
		name     string
		order    *Order
		expected sdk.Coins
		expPanic string
	}{
		{
			name: "AskOrder",
			order: NewOrder(1).WithAsk(&AskOrder{
				Assets:                  sdk.NewInt64Coin("acorns", 85),
				SellerSettlementFlatFee: &sdk.Coin{Denom: "bananas", Amount: sdkmath.NewInt(12)},
			}),
			expected: sdk.NewCoins(sdk.NewInt64Coin("acorns", 85), sdk.NewInt64Coin("bananas", 12)),
		},
		{
			name: "BidOrder",
			order: NewOrder(2).WithBid(&BidOrder{
				Price:               sdk.NewInt64Coin("acorns", 85),
				BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("bananas", 4), sdk.NewInt64Coin("cucumber", 7)),
			}),
			expected: sdk.NewCoins(
				sdk.NewInt64Coin("acorns", 85),
				sdk.NewInt64Coin("bananas", 4),
				sdk.NewInt64Coin("cucumber", 7),
			),
		},
		{
			name:     "nil inside order",
			order:    NewOrder(3),
			expPanic: nilSubTypeErr,
		},
		{
			name:     "unknown order type",
			order:    &Order{OrderId: 4, Order: &unknownOrderType{}},
			expPanic: unknownSubTypeErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coins
			testFunc := func() {
				actual = tc.order.GetHoldAmount()
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "GetHoldAmount()")
			assert.Equal(t, tc.expected.String(), actual.String(), "GetHoldAmount() result")
		})
	}
}

func TestOrder_Validate(t *testing.T) {
	// Annoyingly, sdkmath.Int{} (the zero value) causes panic whenever you
	// try to do anything with it. So we have to give it something.
	zeroCoin := sdk.Coin{Amount: sdkmath.ZeroInt()}
	tests := []struct {
		name  string
		Order *Order
		exp   []string
	}{
		{
			name:  "order id is zero",
			Order: NewOrder(0),
			exp:   []string{"invalid order id", "must not be zero"},
		},
		{
			name:  "nil sub-order",
			Order: NewOrder(1),
			exp:   []string{nilSubTypeErr},
		},
		{
			name:  "unknown sub-order type",
			Order: &Order{OrderId: 1, Order: &unknownOrderType{}},
			exp:   []string{unknownSubTypeErr},
		},
		{
			name:  "ask order error",
			Order: NewOrder(1).WithAsk(&AskOrder{MarketId: 0, Price: zeroCoin}),
			exp:   []string{"invalid market id", "must not be zero"},
		},
		{
			name:  "bid order error",
			Order: NewOrder(1).WithBid(&BidOrder{MarketId: 0, Price: zeroCoin}),
			exp:   []string{"invalid market id", "must not be zero"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.Order.Validate()
			assertions.AssertErrorContents(t, err, tc.exp, "Validate() error")
		})
	}
}

func TestAskOrder_GetMarketID(t *testing.T) {
	tests := []struct {
		name  string
		order AskOrder
		exp   uint32
	}{
		{name: "zero", order: AskOrder{MarketId: 0}, exp: 0},
		{name: "one", order: AskOrder{MarketId: 1}, exp: 1},
		{name: "five", order: AskOrder{MarketId: 5}, exp: 5},
		{name: "twenty-four", order: AskOrder{MarketId: 24}, exp: 24},
		{name: "max uint16+1", order: AskOrder{MarketId: 65_536}, exp: 65_536},
		{name: "max uint32", order: AskOrder{MarketId: 4_294_967_295}, exp: 4_294_967_295},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual uint32
			testFunc := func() {
				actual = tc.order.GetMarketID()
			}
			require.NotPanics(t, testFunc, "GetMarketID()")
			assert.Equal(t, tc.exp, actual, "GetMarketID() result")
		})
	}
}

func TestAskOrder_GetOwner(t *testing.T) {
	acc20 := sdk.AccAddress("seller______________").String()
	acc32 := sdk.AccAddress("__seller__seller__seller__seller").String()

	tests := []struct {
		name  string
		order AskOrder
		exp   string
	}{
		{name: "empty", order: AskOrder{Seller: ""}, exp: ""},
		{name: "not a bech32", order: AskOrder{Seller: "nopenopenope"}, exp: "nopenopenope"},
		{name: "20-byte bech32", order: AskOrder{Seller: acc20}, exp: acc20},
		{name: "32-byte bech32", order: AskOrder{Seller: acc32}, exp: acc32},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = tc.order.GetOwner()
			}
			require.NotPanics(t, testFunc, "GetOwner()")
			assert.Equal(t, tc.exp, actual, "GetOwner() result")
		})
	}
}

func TestAskOrder_GetAssets(t *testing.T) {
	largeAmt, ok := sdkmath.NewIntFromString("25000000000000000000000")
	require.Truef(t, ok, "NewIntFromString(\"25000000000000000000000\")")
	largeCoin := sdk.NewCoin("large", largeAmt)
	negCoin := sdk.Coin{Denom: "neg", Amount: sdkmath.NewInt(-88)}

	tests := []struct {
		name  string
		order AskOrder
		exp   sdk.Coin
	}{
		{name: "one", order: AskOrder{Assets: sdk.NewInt64Coin("one", 1)}, exp: sdk.NewInt64Coin("one", 1)},
		{name: "zero", order: AskOrder{Assets: sdk.NewInt64Coin("zero", 0)}, exp: sdk.NewInt64Coin("zero", 0)},
		{name: "negative", order: AskOrder{Assets: negCoin}, exp: negCoin},
		{name: "large amount", order: AskOrder{Assets: largeCoin}, exp: largeCoin},
		{name: "zero-value", order: AskOrder{Assets: sdk.Coin{}}, exp: sdk.Coin{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coin
			testFunc := func() {
				actual = tc.order.GetAssets()
			}
			require.NotPanics(t, testFunc, "GetAssets()")
			assert.Equal(t, tc.exp.String(), actual.String(), "GetAssets() result")
		})
	}
}

func TestAskOrder_GetPrice(t *testing.T) {
	largeAmt, ok := sdkmath.NewIntFromString("25000000000000000000000")
	require.Truef(t, ok, "NewIntFromString(\"25000000000000000000000\")")
	largeCoin := sdk.NewCoin("large", largeAmt)
	negCoin := sdk.Coin{Denom: "neg", Amount: sdkmath.NewInt(-88)}

	tests := []struct {
		name  string
		order AskOrder
		exp   sdk.Coin
	}{
		{name: "one", order: AskOrder{Price: sdk.NewInt64Coin("one", 1)}, exp: sdk.NewInt64Coin("one", 1)},
		{name: "zero", order: AskOrder{Price: sdk.NewInt64Coin("zero", 0)}, exp: sdk.NewInt64Coin("zero", 0)},
		{name: "negative", order: AskOrder{Price: negCoin}, exp: negCoin},
		{name: "large amount", order: AskOrder{Price: largeCoin}, exp: largeCoin},
		{name: "zero-value", order: AskOrder{Price: sdk.Coin{}}, exp: sdk.Coin{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coin
			testFunc := func() {
				actual = tc.order.GetPrice()
			}
			require.NotPanics(t, testFunc, "GetPrice()")
			assert.Equal(t, tc.exp.String(), actual.String(), "GetPrice() result")
		})
	}
}

func TestAskOrder_GetSettlementFees(t *testing.T) {
	coin := func(amount int64, denom string) *sdk.Coin {
		return &sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	largeAmt, ok := sdkmath.NewIntFromString("25000000000000000000000")
	require.Truef(t, ok, "NewIntFromString(\"25000000000000000000000\")")
	largeCoin := sdk.NewCoin("large", largeAmt)

	tests := []struct {
		name  string
		order AskOrder
		exp   sdk.Coins
	}{
		{name: "nil", order: AskOrder{SellerSettlementFlatFee: nil}, exp: nil},
		{name: "zero amount", order: AskOrder{SellerSettlementFlatFee: coin(0, "zero")}, exp: sdk.Coins{*coin(0, "zero")}},
		{name: "one amount", order: AskOrder{SellerSettlementFlatFee: coin(1, "one")}, exp: sdk.Coins{*coin(1, "one")}},
		{name: "negative amount", order: AskOrder{SellerSettlementFlatFee: coin(-51, "neg")}, exp: sdk.Coins{*coin(-51, "neg")}},
		{name: "large amount", order: AskOrder{SellerSettlementFlatFee: &largeCoin}, exp: sdk.Coins{largeCoin}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coins
			testFunc := func() {
				actual = tc.order.GetSettlementFees()
			}
			require.NotPanics(t, testFunc, "GetSettlementFees()")
			assert.Equal(t, tc.exp.String(), actual.String(), "GetSettlementFees() result")
		})
	}
}

func TestAskOrder_PartialFillAllowed(t *testing.T) {
	tests := []struct {
		name  string
		order AskOrder
		exp   bool
	}{
		{name: "false", order: AskOrder{AllowPartial: false}, exp: false},
		{name: "true", order: AskOrder{AllowPartial: true}, exp: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = tc.order.PartialFillAllowed()
			}
			require.NotPanics(t, testFunc, "PartialFillAllowed()")
			assert.Equal(t, tc.exp, actual, "PartialFillAllowed() result")
		})
	}
}

func TestAskOrder_GetOrderType(t *testing.T) {
	expected := OrderTypeAsk
	order := AskOrder{}
	var actual string
	testFunc := func() {
		actual = order.GetOrderType()
	}
	require.NotPanics(t, testFunc, "GetOrderType()")
	assert.Equal(t, expected, actual, "GetOrderType() result")
}

func TestAskOrder_GetOrderTypeByte(t *testing.T) {
	expected := OrderTypeByteAsk
	order := AskOrder{}
	var actual byte
	testFunc := func() {
		actual = order.GetOrderTypeByte()
	}
	require.NotPanics(t, testFunc, "GetOrderTypeByte()")
	assert.Equal(t, expected, actual, "GetOrderTypeByte() result")
}

func TestAskOrder_GetHoldAmount(t *testing.T) {
	tests := []struct {
		name  string
		order AskOrder
		exp   sdk.Coins
	}{
		{
			name: "just assets",
			order: AskOrder{
				Assets: sdk.NewInt64Coin("acorn", 12),
			},
			exp: sdk.NewCoins(sdk.NewInt64Coin("acorn", 12)),
		},
		{
			name: "settlement fee is different denom from price",
			order: AskOrder{
				Assets:                  sdk.NewInt64Coin("acorn", 12),
				Price:                   sdk.NewInt64Coin("cucumber", 8),
				SellerSettlementFlatFee: &sdk.Coin{Denom: "durian", Amount: sdkmath.NewInt(52)},
			},
			exp: sdk.NewCoins(
				sdk.NewInt64Coin("acorn", 12),
				sdk.NewInt64Coin("durian", 52),
			),
		},
		{
			name: "settlement fee is same denom as price",
			order: AskOrder{
				Assets:                  sdk.NewInt64Coin("acorn", 12),
				Price:                   sdk.NewInt64Coin("cucumber", 8),
				SellerSettlementFlatFee: &sdk.Coin{Denom: "cucumber", Amount: sdkmath.NewInt(52)},
			},
			exp: sdk.NewCoins(sdk.NewInt64Coin("acorn", 12)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coins
			testFunc := func() {
				actual = tc.order.GetHoldAmount()
			}
			require.NotPanics(t, testFunc, "GetHoldAmount()")
			assert.Equal(t, tc.exp, actual, "GetHoldAmount() result")
		})
	}
}

func TestAskOrder_Validate(t *testing.T) {
	coin := func(amount int64, denom string) *sdk.Coin {
		return &sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}

	tests := []struct {
		name  string
		order AskOrder
		exp   []string
	}{
		{
			name: "control",
			order: AskOrder{
				MarketId:                1,
				Seller:                  sdk.AccAddress("control_address_____").String(),
				Assets:                  *coin(99, "bender"),
				Price:                   *coin(42, "farnsworth"),
				SellerSettlementFlatFee: coin(1, "farnsworth"),
				AllowPartial:            false,
			},
			exp: nil,
		},
		{
			name: "allow partial",
			order: AskOrder{
				MarketId:                1,
				Seller:                  sdk.AccAddress("control_address_____").String(),
				Assets:                  *coin(99, "bender"),
				Price:                   *coin(42, "farnsworth"),
				SellerSettlementFlatFee: coin(1, "farnsworth"),
				AllowPartial:            true,
			},
			exp: nil,
		},
		{
			name: "nil seller settlement flat fee",
			order: AskOrder{
				MarketId:                1,
				Seller:                  sdk.AccAddress("control_address_____").String(),
				Assets:                  *coin(99, "bender"),
				Price:                   *coin(42, "farnsworth"),
				SellerSettlementFlatFee: nil,
				AllowPartial:            false,
			},
			exp: nil,
		},
		{
			name: "market id zero",
			order: AskOrder{
				MarketId: 0,
				Seller:   sdk.AccAddress("another_address_____").String(),
				Assets:   *coin(99, "bender"),
				Price:    *coin(42, "farnsworth"),
			},
			exp: []string{"invalid market id", "must not be zero"},
		},
		{
			name: "invalid seller",
			order: AskOrder{
				MarketId: 1,
				Seller:   "shady_address_______",
				Assets:   *coin(99, "bender"),
				Price:    *coin(42, "farnsworth"),
			},
			exp: []string{"invalid seller", "invalid separator index -1"},
		},
		{
			name: "no seller",
			order: AskOrder{
				MarketId: 1,
				Seller:   "",
				Assets:   *coin(99, "bender"),
				Price:    *coin(42, "farnsworth"),
			},
			exp: []string{"invalid seller", "empty address string is not allowed"},
		},
		{
			name: "zero price",
			order: AskOrder{
				MarketId: 1,
				Seller:   sdk.AccAddress("another_address_____").String(),
				Assets:   *coin(99, "bender"),
				Price:    *coin(0, "farnsworth"),
			},
			exp: []string{"invalid price", "cannot be zero"},
		},
		{
			name: "negative price",
			order: AskOrder{
				MarketId: 1,
				Seller:   sdk.AccAddress("another_address_____").String(),
				Assets:   *coin(99, "bender"),
				Price:    *coin(-24, "farnsworth"),
			},
			exp: []string{"invalid price", "negative coin amount: -24"},
		},
		{
			name: "invalid price denom",
			order: AskOrder{
				MarketId: 1,
				Seller:   sdk.AccAddress("another_address_____").String(),
				Assets:   *coin(99, "bender"),
				Price:    *coin(42, "7"),
			},
			exp: []string{"invalid price", "invalid denom: 7"},
		},
		{
			name: "zero-value price",
			order: AskOrder{
				MarketId: 1,
				Seller:   sdk.AccAddress("seller_address______").String(),
				Assets:   *coin(99, "bender"),
				Price:    sdk.Coin{},
			},
			exp: []string{"invalid price", "invalid denom"},
		},
		{
			name: "zero amount in assets",
			order: AskOrder{
				MarketId: 1,
				Seller:   sdk.AccAddress("another_address_____").String(),
				Assets:   *coin(0, "leela"),
				Price:    *coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets", "cannot be zero"},
		},
		{
			name: "negative amount in assets",
			order: AskOrder{
				MarketId: 1,
				Seller:   sdk.AccAddress("another_address_____").String(),
				Assets:   *coin(-1, "leela"),
				Price:    *coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets", "negative coin amount: -1"},
		},
		{
			name: "invalid denom in assets",
			order: AskOrder{
				MarketId: 1,
				Seller:   sdk.AccAddress("another_address_____").String(),
				Assets:   *coin(1, "x"),
				Price:    *coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets", "invalid denom: x"},
		},
		{
			name: "zero-value assets",
			order: AskOrder{
				MarketId: 1,
				Seller:   sdk.AccAddress("another_address_____").String(),
				Assets:   sdk.Coin{},
				Price:    *coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets", "invalid denom: "},
		},
		{
			name: "price denom in assets",
			order: AskOrder{
				MarketId: 1,
				Seller:   sdk.AccAddress("another_address_____").String(),
				Assets:   *coin(2, "farnsworth"),
				Price:    *coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets", "price denom farnsworth cannot also be the assets denom"},
		},
		{
			name: "invalid seller settlement flat fee denom",
			order: AskOrder{
				MarketId:                1,
				Seller:                  sdk.AccAddress("another_address_____").String(),
				Assets:                  *coin(99, "bender"),
				Price:                   *coin(42, "farnsworth"),
				SellerSettlementFlatFee: coin(13, "x"),
			},
			exp: []string{"invalid seller settlement flat fee", "invalid denom: x"},
		},
		{
			name: "zero seller settlement flat fee denom",
			order: AskOrder{
				MarketId:                1,
				Seller:                  sdk.AccAddress("another_address_____").String(),
				Assets:                  *coin(99, "bender"),
				Price:                   *coin(42, "farnsworth"),
				SellerSettlementFlatFee: coin(0, "nibbler"),
			},
			exp: []string{"invalid seller settlement flat fee", "nibbler amount cannot be zero"},
		},
		{
			name: "negative seller settlement flat fee",
			order: AskOrder{
				MarketId:                1,
				Seller:                  sdk.AccAddress("another_address_____").String(),
				Assets:                  *coin(99, "bender"),
				Price:                   *coin(42, "farnsworth"),
				SellerSettlementFlatFee: coin(-3, "nibbler"),
			},
			exp: []string{"invalid seller settlement flat fee", "negative coin amount: -3"},
		},
		{
			name: "multiple problems",
			order: AskOrder{
				Price:                   *coin(0, ""),
				SellerSettlementFlatFee: coin(0, ""),
			},
			exp: []string{
				"invalid market id",
				"invalid seller",
				"invalid price",
				"invalid assets",
				"invalid seller settlement flat fee",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.order.Validate()
			assertions.AssertErrorContents(t, err, tc.exp, "Validate() error")
		})
	}
}

func TestBidOrder_GetMarketID(t *testing.T) {
	tests := []struct {
		name  string
		order BidOrder
		exp   uint32
	}{
		{name: "zero", order: BidOrder{MarketId: 0}, exp: 0},
		{name: "one", order: BidOrder{MarketId: 1}, exp: 1},
		{name: "five", order: BidOrder{MarketId: 5}, exp: 5},
		{name: "twenty-four", order: BidOrder{MarketId: 24}, exp: 24},
		{name: "max uint16+1", order: BidOrder{MarketId: 65_536}, exp: 65_536},
		{name: "max uint32", order: BidOrder{MarketId: 4_294_967_295}, exp: 4_294_967_295},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual uint32
			testFunc := func() {
				actual = tc.order.GetMarketID()
			}
			require.NotPanics(t, testFunc, "GetMarketID()")
			assert.Equal(t, tc.exp, actual, "GetMarketID() result")
		})
	}
}

func TestBidOrder_GetOwner(t *testing.T) {
	acc20 := sdk.AccAddress("buyer_______________").String()
	acc32 := sdk.AccAddress("___buyer___buyer___buyer___buyer").String()

	tests := []struct {
		name  string
		order BidOrder
		exp   string
	}{
		{name: "empty", order: BidOrder{Buyer: ""}, exp: ""},
		{name: "not a bech32", order: BidOrder{Buyer: "nopenopenope"}, exp: "nopenopenope"},
		{name: "20-byte bech32", order: BidOrder{Buyer: acc20}, exp: acc20},
		{name: "32-byte bech32", order: BidOrder{Buyer: acc32}, exp: acc32},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = tc.order.GetOwner()
			}
			require.NotPanics(t, testFunc, "GetOwner()")
			assert.Equal(t, tc.exp, actual, "GetOwner() result")
		})
	}
}

func TestBidOrder_GetAssets(t *testing.T) {
	largeAmt, ok := sdkmath.NewIntFromString("25000000000000000000000")
	require.Truef(t, ok, "NewIntFromString(\"25000000000000000000000\")")
	largeCoin := sdk.NewCoin("large", largeAmt)
	negCoin := sdk.Coin{Denom: "neg", Amount: sdkmath.NewInt(-88)}

	tests := []struct {
		name  string
		order BidOrder
		exp   sdk.Coin
	}{
		{name: "one", order: BidOrder{Assets: sdk.NewInt64Coin("one", 1)}, exp: sdk.NewInt64Coin("one", 1)},
		{name: "zero", order: BidOrder{Assets: sdk.NewInt64Coin("zero", 0)}, exp: sdk.NewInt64Coin("zero", 0)},
		{name: "negative", order: BidOrder{Assets: negCoin}, exp: negCoin},
		{name: "large amount", order: BidOrder{Assets: largeCoin}, exp: largeCoin},
		{name: "zero-value", order: BidOrder{Assets: sdk.Coin{}}, exp: sdk.Coin{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coin
			testFunc := func() {
				actual = tc.order.GetAssets()
			}
			require.NotPanics(t, testFunc, "GetAssets()")
			assert.Equal(t, tc.exp.String(), actual.String(), "GetAssets() result")
		})
	}
}

func TestBidOrder_GetPrice(t *testing.T) {
	largeAmt, ok := sdkmath.NewIntFromString("25000000000000000000000")
	require.Truef(t, ok, "NewIntFromString(\"25000000000000000000000\")")
	largeCoin := sdk.NewCoin("large", largeAmt)
	negCoin := sdk.Coin{Denom: "neg", Amount: sdkmath.NewInt(-88)}

	tests := []struct {
		name  string
		order BidOrder
		exp   sdk.Coin
	}{
		{name: "one", order: BidOrder{Price: sdk.NewInt64Coin("one", 1)}, exp: sdk.NewInt64Coin("one", 1)},
		{name: "zero", order: BidOrder{Price: sdk.NewInt64Coin("zero", 0)}, exp: sdk.NewInt64Coin("zero", 0)},
		{name: "negative", order: BidOrder{Price: negCoin}, exp: negCoin},
		{name: "large amount", order: BidOrder{Price: largeCoin}, exp: largeCoin},
		{name: "zero-value", order: BidOrder{Price: sdk.Coin{}}, exp: sdk.Coin{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coin
			testFunc := func() {
				actual = tc.order.GetPrice()
			}
			require.NotPanics(t, testFunc, "GetPrice()")
			assert.Equal(t, tc.exp.String(), actual.String(), "GetPrice() result")
		})
	}
}

func TestBidOrder_GetSettlementFees(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	largeAmt, ok := sdkmath.NewIntFromString("25000000000000000000000")
	require.Truef(t, ok, "NewIntFromString(\"25000000000000000000000\")")
	largeCoin := sdk.NewCoin("large", largeAmt)

	tests := []struct {
		name  string
		order BidOrder
		exp   sdk.Coins
	}{
		{name: "nil", order: BidOrder{BuyerSettlementFees: nil}, exp: nil},
		{name: "empty", order: BidOrder{BuyerSettlementFees: sdk.Coins{}}, exp: sdk.Coins{}},
		{name: "zero amount", order: BidOrder{BuyerSettlementFees: sdk.Coins{coin(0, "zero")}}, exp: sdk.Coins{coin(0, "zero")}},
		{name: "one amount", order: BidOrder{BuyerSettlementFees: sdk.Coins{coin(1, "one")}}, exp: sdk.Coins{coin(1, "one")}},
		{name: "negative amount", order: BidOrder{BuyerSettlementFees: sdk.Coins{coin(-51, "neg")}}, exp: sdk.Coins{coin(-51, "neg")}},
		{name: "large amount", order: BidOrder{BuyerSettlementFees: sdk.Coins{largeCoin}}, exp: sdk.Coins{largeCoin}},
		{
			name: "multiple coins",
			order: BidOrder{
				BuyerSettlementFees: sdk.Coins{largeCoin, coin(1, "one"), coin(0, "zero")},
			},
			exp: sdk.Coins{largeCoin, coin(1, "one"), coin(0, "zero")},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coins
			testFunc := func() {
				actual = tc.order.GetSettlementFees()
			}
			require.NotPanics(t, testFunc, "GetSettlementFees()")
			assert.Equal(t, tc.exp.String(), actual.String(), "GetSettlementFees() result")
		})
	}
}

func TestBidOrder_PartialFillAllowed(t *testing.T) {
	tests := []struct {
		name  string
		order BidOrder
		exp   bool
	}{
		{name: "false", order: BidOrder{AllowPartial: false}, exp: false},
		{name: "true", order: BidOrder{AllowPartial: true}, exp: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = tc.order.PartialFillAllowed()
			}
			require.NotPanics(t, testFunc, "PartialFillAllowed()")
			assert.Equal(t, tc.exp, actual, "PartialFillAllowed() result")
		})
	}
}

func TestBidOrder_GetOrderType(t *testing.T) {
	expected := OrderTypeBid
	order := BidOrder{}
	var actual string
	testFunc := func() {
		actual = order.GetOrderType()
	}
	require.NotPanics(t, testFunc, "GetOrderType()")
	assert.Equal(t, expected, actual, "GetOrderType() result")

}

func TestBidOrder_GetOrderTypeByte(t *testing.T) {
	expected := OrderTypeByteBid
	order := BidOrder{}
	var actual byte
	testFunc := func() {
		actual = order.GetOrderTypeByte()
	}
	require.NotPanics(t, testFunc, "GetOrderTypeByte()")
	assert.Equal(t, expected, actual, "GetOrderTypeByte() result")
}

func TestBidOrder_GetHoldAmount(t *testing.T) {
	tests := []struct {
		name  string
		order BidOrder
		exp   sdk.Coins
	}{
		{
			name: "just price",
			order: BidOrder{
				Price: sdk.NewInt64Coin("cucumber", 8),
			},
			exp: sdk.NewCoins(sdk.NewInt64Coin("cucumber", 8)),
		},
		{
			name: "price and settlement fee with shared denom",
			order: BidOrder{
				Price:               sdk.NewInt64Coin("cucumber", 8),
				BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("acorn", 5), sdk.NewInt64Coin("cucumber", 1)),
			},
			exp: sdk.NewCoins(
				sdk.NewInt64Coin("acorn", 5),
				sdk.NewInt64Coin("cucumber", 9),
			),
		},
		{
			name: "price and settlement fee with different denom",
			order: BidOrder{
				Price:               sdk.NewInt64Coin("cucumber", 8),
				BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("acorn", 5), sdk.NewInt64Coin("banana", 1)),
			},
			exp: sdk.NewCoins(
				sdk.NewInt64Coin("acorn", 5),
				sdk.NewInt64Coin("banana", 1),
				sdk.NewInt64Coin("cucumber", 8),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coins
			testFunc := func() {
				actual = tc.order.GetHoldAmount()
			}
			require.NotPanics(t, testFunc, "GetHoldAmount()")
			assert.Equal(t, tc.exp, actual, "GetHoldAmount() result")
		})
	}
}

func TestBidOrder_Validate(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	coins := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "sdk.ParseCoinsNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name  string
		order BidOrder
		exp   []string
	}{
		{
			name: "control",
			order: BidOrder{
				MarketId:            1,
				Buyer:               sdk.AccAddress("control_address_____").String(),
				Assets:              coin(99, "bender"),
				Price:               coin(42, "farnsworth"),
				BuyerSettlementFees: coins("1farnsworth"),
				AllowPartial:        false,
			},
			exp: nil,
		},
		{
			name: "allow partial",
			order: BidOrder{
				MarketId:            1,
				Buyer:               sdk.AccAddress("control_address_____").String(),
				Assets:              coin(99, "bender"),
				Price:               coin(42, "farnsworth"),
				BuyerSettlementFees: coins("1farnsworth"),
				AllowPartial:        true,
			},
			exp: nil,
		},
		{
			name: "nil buyer settlement fees",
			order: BidOrder{
				MarketId:            1,
				Buyer:               sdk.AccAddress("control_address_____").String(),
				Assets:              coin(99, "bender"),
				Price:               coin(42, "farnsworth"),
				BuyerSettlementFees: nil,
				AllowPartial:        false,
			},
			exp: nil,
		},
		{
			name: "empty buyer settlement fees",
			order: BidOrder{
				MarketId:            1,
				Buyer:               sdk.AccAddress("control_address_____").String(),
				Assets:              coin(99, "bender"),
				Price:               coin(42, "farnsworth"),
				BuyerSettlementFees: sdk.Coins{},
				AllowPartial:        false,
			},
			exp: nil,
		},
		{
			name: "market id zero",
			order: BidOrder{
				MarketId: 0,
				Buyer:    sdk.AccAddress("another_address_____").String(),
				Assets:   coin(99, "bender"),
				Price:    coin(42, "farnsworth"),
			},
			exp: []string{"invalid market id: must not be zero"},
		},
		{
			name: "invalid buyer",
			order: BidOrder{
				MarketId: 1,
				Buyer:    "shady_address_______",
				Assets:   coin(99, "bender"),
				Price:    coin(42, "farnsworth"),
			},
			exp: []string{"invalid buyer", "invalid separator index -1"},
		},
		{
			name: "no buyer",
			order: BidOrder{
				MarketId: 1,
				Buyer:    "",
				Assets:   coin(99, "bender"),
				Price:    coin(42, "farnsworth"),
			},
			exp: []string{"invalid buyer", "empty address string is not allowed"},
		},
		{
			name: "zero price",
			order: BidOrder{
				MarketId: 1,
				Buyer:    sdk.AccAddress("another_address_____").String(),
				Assets:   coin(99, "bender"),
				Price:    coin(0, "farnsworth"),
			},
			exp: []string{"invalid price", "cannot be zero"},
		},
		{
			name: "negative price",
			order: BidOrder{
				MarketId: 1,
				Buyer:    sdk.AccAddress("another_address_____").String(),
				Assets:   coin(99, "bender"),
				Price:    coin(-24, "farnsworth"),
			},
			exp: []string{"invalid price", "negative coin amount: -24"},
		},
		{
			name: "invalid price denom",
			order: BidOrder{
				MarketId: 1,
				Buyer:    sdk.AccAddress("another_address_____").String(),
				Assets:   coin(99, "bender"),
				Price:    coin(42, "7"),
			},
			exp: []string{"invalid price", "invalid denom: 7"},
		},
		{
			name: "zero-value price",
			order: BidOrder{
				MarketId: 1,
				Buyer:    sdk.AccAddress("buyer_address_______").String(),
				Assets:   coin(99, "bender"),
				Price:    sdk.Coin{},
			},
			exp: []string{"invalid price", "invalid denom"},
		},
		{
			name: "zero amount in assets",
			order: BidOrder{
				MarketId: 1,
				Buyer:    sdk.AccAddress("another_address_____").String(),
				Assets:   coin(0, "leela"),
				Price:    coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets", "cannot be zero"},
		},
		{
			name: "negative amount in assets",
			order: BidOrder{
				MarketId: 1,
				Buyer:    sdk.AccAddress("another_address_____").String(),
				Assets:   coin(-1, "leela"),
				Price:    coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets", "negative coin amount: -1"},
		},
		{
			name: "invalid denom in assets",
			order: BidOrder{
				MarketId: 1,
				Buyer:    sdk.AccAddress("another_address_____").String(),
				Assets:   coin(1, "x"),
				Price:    coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets", "invalid denom: x"},
		},
		{
			name: "zero-value assets",
			order: BidOrder{
				MarketId: 1,
				Buyer:    sdk.AccAddress("another_address_____").String(),
				Assets:   sdk.Coin{},
				Price:    coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets", "invalid denom"},
		},
		{
			name: "price denom in assets",
			order: BidOrder{
				MarketId: 1,
				Buyer:    sdk.AccAddress("another_address_____").String(),
				Assets:   coin(2, "farnsworth"),
				Price:    coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets", "price denom farnsworth cannot also be the assets denom"},
		},
		{
			name: "invalid denom in buyer settlement fees",
			order: BidOrder{
				MarketId:            1,
				Buyer:               sdk.AccAddress("another_address_____").String(),
				Assets:              coin(99, "bender"),
				Price:               coin(42, "farnsworth"),
				BuyerSettlementFees: sdk.Coins{coin(1, "farnsworth"), coin(2, "x")},
			},
			exp: []string{"invalid buyer settlement fees", "invalid denom: x"},
		},
		{
			name: "negative buyer settlement fees",
			order: BidOrder{
				MarketId:            1,
				Buyer:               sdk.AccAddress("another_address_____").String(),
				Assets:              coin(99, "bender"),
				Price:               coin(42, "farnsworth"),
				BuyerSettlementFees: sdk.Coins{coin(3, "farnsworth"), coin(-3, "nibbler")},
			},
			exp: []string{"invalid buyer settlement fees", "coin nibbler amount is not positive"},
		},
		{
			name: "multiple problems",
			order: BidOrder{
				Price:               coin(0, ""),
				BuyerSettlementFees: sdk.Coins{coin(0, "")},
			},
			exp: []string{
				"invalid market id",
				"invalid buyer",
				"invalid price",
				"invalid assets",
				"invalid buyer settlement fees",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.order.Validate()
			assertions.AssertErrorContents(t, err, tc.exp, "Validate() error")
		})
	}
}
