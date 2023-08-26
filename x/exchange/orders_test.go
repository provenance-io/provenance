package exchange

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		Assets:                  sdk.NewCoins(sdk.NewInt64Coin("water", 8)),
		Price:                   sdk.NewInt64Coin("sand", 1),
		SellerSettlementFlatFee: &sdk.Coin{Denom: "banana", Amount: sdkmath.NewInt(3)},
		AllowPartial:            true,
	}
	ask := &AskOrder{
		MarketId:                origAsk.MarketId,
		Seller:                  origAsk.Seller,
		Assets:                  copyCoins(origAsk.Assets),
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
		Assets:              sdk.NewCoins(sdk.NewInt64Coin("agua", 7)),
		Price:               sdk.NewInt64Coin("dirt", 2),
		BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("grape", 3)),
		AllowPartial:        true,
	}
	bid := &BidOrder{
		MarketId:            origBid.MarketId,
		Buyer:               origBid.Buyer,
		Assets:              copyCoins(origBid.Assets),
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

func TestOrder_OrderType(t *testing.T) {
	tests := []struct {
		name     string
		order    *Order
		expected string
		expPanic interface{}
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
			name:     "nil base order",
			order:    nil,
			expPanic: "OrderType() missing case for <nil>",
		},
		{
			name:     "nil inside order",
			order:    NewOrder(3),
			expPanic: "OrderType() missing case for <nil>",
		},
		{
			name:     "unknown order type",
			order:    &Order{OrderId: 4, Order: &unknownOrderType{}},
			expPanic: "OrderType() missing case for *exchange.unknownOrderType",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = tc.order.OrderType()
			}

			if tc.expPanic != nil {
				require.PanicsWithValue(t, tc.expPanic, testFunc, "OrderType")
			} else {
				require.NotPanics(t, testFunc, "OrderType")
				assert.Equal(t, tc.expected, actual, "OrderType result")
			}
		})
	}
}

func TestOrder_OrderTypeByte(t *testing.T) {
	tests := []struct {
		name     string
		order    *Order
		expected byte
		expPanic interface{}
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
			name:     "nil base order",
			order:    nil,
			expPanic: "OrderTypeByte() missing case for <nil>",
		},
		{
			name:     "nil inside order",
			order:    NewOrder(3),
			expPanic: "OrderTypeByte() missing case for <nil>",
		},
		{
			name:     "unknown order type",
			order:    &Order{OrderId: 4, Order: &unknownOrderType{}},
			expPanic: "OrderTypeByte() missing case for *exchange.unknownOrderType",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual byte
			testFunc := func() {
				actual = tc.order.OrderTypeByte()
			}

			if tc.expPanic != nil {
				require.PanicsWithValue(t, tc.expPanic, testFunc, "OrderTypeByte")
			} else {
				require.NotPanics(t, testFunc, "OrderTypeByte")
				assert.Equal(t, tc.expected, actual, "OrderTypeByte result")
			}
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
			exp:   []string{"invalid order id: must not be zero"},
		},
		{
			name:  "unknown order type",
			Order: &Order{OrderId: 1, Order: &unknownOrderType{}},
			exp:   []string{"unknown order type *exchange.unknownOrderType"},
		},
		{
			name:  "ask order error",
			Order: NewOrder(1).WithAsk(&AskOrder{MarketId: 0, Price: zeroCoin}),
			exp:   []string{"invalid market id: must not be zero"},
		},
		{
			name:  "bid order error",
			Order: NewOrder(1).WithBid(&BidOrder{MarketId: 0, Price: zeroCoin}),
			exp:   []string{"invalid market id: must not be zero"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.Order.Validate()
			if len(tc.exp) > 0 {
				if assert.Error(t, err, "Validate() error") {
					for _, exp := range tc.exp {
						assert.ErrorContains(t, err, exp, "Validate() error\nExpecting: %q", exp)
					}
				}
			}
		})
	}
}

// TODO[1658]: func TestAskOrder_Validate(t *testing.T)

// TODO[1658]: func TestBidOrder_Validate(t *testing.T)
