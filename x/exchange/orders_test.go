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
		return "GetOrderType" + name
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
		expPanic string
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
			expPanic: "GetOrderType() missing case for <nil>",
		},
		{
			name:     "unknown order type",
			order:    &Order{OrderId: 4, Order: &unknownOrderType{}},
			expPanic: "GetOrderType() missing case for *exchange.unknownOrderType",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = tc.order.GetOrderType()
			}

			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "GetOrderType")
			assert.Equal(t, tc.expected, actual, "GetOrderType result")
		})
	}
}

func TestOrder_OrderTypeByte(t *testing.T) {
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
			expPanic: "GetOrderTypeByte() missing case for <nil>",
		},
		{
			name:     "unknown order type",
			order:    &Order{OrderId: 4, Order: &unknownOrderType{}},
			expPanic: "GetOrderTypeByte() missing case for *exchange.unknownOrderType",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual byte
			testFunc := func() {
				actual = tc.order.GetOrderTypeByte()
			}

			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "GetOrderTypeByte")
			assert.Equal(t, tc.expected, actual, "GetOrderTypeByte result")
		})
	}
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
			require.NotPanics(t, testFunc, "IsAskOrder")
			assert.Equal(t, tc.exp, actual, "IsAskOrder result")
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
			require.NotPanics(t, testFunc, "IsBidOrder")
			assert.Equal(t, tc.exp, actual, "IsBidOrder result")
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
			expPanic: "GetMarketID() missing case for <nil>",
		},
		{
			name:     "unknown order type",
			order:    &Order{OrderId: 4, Order: &unknownOrderType{}},
			expPanic: "GetMarketID() missing case for *exchange.unknownOrderType",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual uint32
			testFunc := func() {
				actual = tc.order.GetMarketID()
			}

			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "GetMarketID")
			assert.Equal(t, tc.expected, actual, "GetMarketID result")
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

			assertions.AssertErrorContents(t, err, tc.exp, "Validate() error")
		})
	}
}

func TestAskOrder_Validate(t *testing.T) {
	coin := func(amount int64, denom string) *sdk.Coin {
		return &sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	coins := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "sdk.ParseCoinsNormalized(%q)", coins)
		return rv
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
				Assets:                  coins("99bender"),
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
				Assets:                  coins("99bender"),
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
				Assets:                  coins("99bender"),
				Price:                   *coin(42, "farnsworth"),
				SellerSettlementFlatFee: nil,
				AllowPartial:            false,
			},
			exp: nil,
		},
		{
			name: "multiple assets",
			order: AskOrder{
				MarketId: 1,
				Seller:   sdk.AccAddress("another_address_____").String(),
				Assets:   coins("12amy,99bender,8fry,112leela,1zoidberg"),
				Price:    *coin(42, "farnsworth"),
			},
			exp: nil,
		},
		{
			name: "market id zero",
			order: AskOrder{
				MarketId: 0,
				Seller:   sdk.AccAddress("another_address_____").String(),
				Assets:   coins("99bender"),
				Price:    *coin(42, "farnsworth"),
			},
			exp: []string{"invalid market id: must not be zero"},
		},
		{
			name: "invalid seller",
			order: AskOrder{
				MarketId: 1,
				Seller:   "shady_address_______",
				Assets:   coins("99bender"),
				Price:    *coin(42, "farnsworth"),
			},
			exp: []string{"invalid seller: ", "invalid separator index -1"},
		},
		{
			name: "no seller",
			order: AskOrder{
				MarketId: 1,
				Seller:   "",
				Assets:   coins("99bender"),
				Price:    *coin(42, "farnsworth"),
			},
			exp: []string{"invalid seller: ", "empty address string is not allowed"},
		},
		{
			name: "zero price",
			order: AskOrder{
				MarketId: 1,
				Seller:   sdk.AccAddress("another_address_____").String(),
				Assets:   coins("99bender"),
				Price:    *coin(0, "farnsworth"),
			},
			exp: []string{"invalid price: cannot be zero"},
		},
		{
			name: "negative price",
			order: AskOrder{
				MarketId: 1,
				Seller:   sdk.AccAddress("another_address_____").String(),
				Assets:   coins("99bender"),
				Price:    *coin(-24, "farnsworth"),
			},
			exp: []string{"invalid price: ", "negative coin amount: -24"},
		},
		{
			name: "invalid price denom",
			order: AskOrder{
				MarketId: 1,
				Seller:   sdk.AccAddress("another_address_____").String(),
				Assets:   coins("99bender"),
				Price:    *coin(42, "7"),
			},
			exp: []string{"invalid price: ", "invalid denom: 7"},
		},
		{
			name: "zero amount in assets",
			order: AskOrder{
				MarketId: 1,
				Seller:   sdk.AccAddress("another_address_____").String(),
				Assets:   sdk.Coins{*coin(99, "bender"), *coin(0, "leela"), *coin(1, "zoidberg")},
				Price:    *coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets: ", "coin leela amount is not positive"},
		},
		{
			name: "negative amount in assets",
			order: AskOrder{
				MarketId: 1,
				Seller:   sdk.AccAddress("another_address_____").String(),
				Assets:   sdk.Coins{*coin(99, "bender"), *coin(-1, "leela"), *coin(1, "zoidberg")},
				Price:    *coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets: ", "coin leela amount is not positive"},
		},
		{
			name: "invalid denom in assets",
			order: AskOrder{
				MarketId: 1,
				Seller:   sdk.AccAddress("another_address_____").String(),
				Assets:   sdk.Coins{*coin(99, "bender"), *coin(1, "x"), *coin(1, "zoidberg")},
				Price:    *coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets: ", "invalid denom: x"},
		},
		{
			name: "nil assets",
			order: AskOrder{
				MarketId: 1,
				Seller:   sdk.AccAddress("another_address_____").String(),
				Assets:   nil,
				Price:    *coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets: must not be empty"},
		},
		{
			name: "empty assets",
			order: AskOrder{
				MarketId: 1,
				Seller:   sdk.AccAddress("another_address_____").String(),
				Assets:   sdk.Coins{},
				Price:    *coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets: must not be empty"},
		},
		{
			name: "price denom in assets",
			order: AskOrder{
				MarketId: 1,
				Seller:   sdk.AccAddress("another_address_____").String(),
				Assets:   coins("99bender,2farnsworth,44amy"),
				Price:    *coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets: cannot contain price denom farnsworth"},
		},
		{
			name: "invalid seller settlement flat fee denom",
			order: AskOrder{
				MarketId:                1,
				Seller:                  sdk.AccAddress("another_address_____").String(),
				Assets:                  coins("99bender"),
				Price:                   *coin(42, "farnsworth"),
				SellerSettlementFlatFee: coin(13, "x"),
			},
			exp: []string{"invalid seller settlement flat fee: ", "invalid denom: x"},
		},
		{
			name: "zero seller settlement flat fee denom",
			order: AskOrder{
				MarketId:                1,
				Seller:                  sdk.AccAddress("another_address_____").String(),
				Assets:                  coins("99bender"),
				Price:                   *coin(42, "farnsworth"),
				SellerSettlementFlatFee: coin(0, "nibbler"),
			},
			exp: []string{"invalid seller settlement flat fee: ", "nibbler amount cannot be zero"},
		},
		{
			name: "negative seller settlement flat fee",
			order: AskOrder{
				MarketId:                1,
				Seller:                  sdk.AccAddress("another_address_____").String(),
				Assets:                  coins("99bender"),
				Price:                   *coin(42, "farnsworth"),
				SellerSettlementFlatFee: coin(-3, "nibbler"),
			},
			exp: []string{"invalid seller settlement flat fee: ", "negative coin amount: -3"},
		},
		{
			name: "multiple problems",
			order: AskOrder{
				Price:                   *coin(0, ""),
				SellerSettlementFlatFee: coin(0, ""),
			},
			exp: []string{
				"invalid market id: ",
				"invalid seller: ",
				"invalid price: ",
				"invalid assets: ",
				"invalid seller settlement flat fee: ",
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
				Assets:              coins("99bender"),
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
				Assets:              coins("99bender"),
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
				Assets:              coins("99bender"),
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
				Assets:              coins("99bender"),
				Price:               coin(42, "farnsworth"),
				BuyerSettlementFees: sdk.Coins{},
				AllowPartial:        false,
			},
			exp: nil,
		},
		{
			name: "multiple assets",
			order: BidOrder{
				MarketId: 1,
				Buyer:    sdk.AccAddress("another_address_____").String(),
				Assets:   coins("12amy,99bender,8fry,112leela,1zoidberg"),
				Price:    coin(42, "farnsworth"),
			},
			exp: nil,
		},
		{
			name: "market id zero",
			order: BidOrder{
				MarketId: 0,
				Buyer:    sdk.AccAddress("another_address_____").String(),
				Assets:   coins("99bender"),
				Price:    coin(42, "farnsworth"),
			},
			exp: []string{"invalid market id: must not be zero"},
		},
		{
			name: "invalid buyer",
			order: BidOrder{
				MarketId: 1,
				Buyer:    "shady_address_______",
				Assets:   coins("99bender"),
				Price:    coin(42, "farnsworth"),
			},
			exp: []string{"invalid buyer: ", "invalid separator index -1"},
		},
		{
			name: "no buyer",
			order: BidOrder{
				MarketId: 1,
				Buyer:    "",
				Assets:   coins("99bender"),
				Price:    coin(42, "farnsworth"),
			},
			exp: []string{"invalid buyer: ", "empty address string is not allowed"},
		},
		{
			name: "zero price",
			order: BidOrder{
				MarketId: 1,
				Buyer:    sdk.AccAddress("another_address_____").String(),
				Assets:   coins("99bender"),
				Price:    coin(0, "farnsworth"),
			},
			exp: []string{"invalid price: cannot be zero"},
		},
		{
			name: "negative price",
			order: BidOrder{
				MarketId: 1,
				Buyer:    sdk.AccAddress("another_address_____").String(),
				Assets:   coins("99bender"),
				Price:    coin(-24, "farnsworth"),
			},
			exp: []string{"invalid price: ", "negative coin amount: -24"},
		},
		{
			name: "invalid price denom",
			order: BidOrder{
				MarketId: 1,
				Buyer:    sdk.AccAddress("another_address_____").String(),
				Assets:   coins("99bender"),
				Price:    coin(42, "7"),
			},
			exp: []string{"invalid price: ", "invalid denom: 7"},
		},
		{
			name: "zero amount in assets",
			order: BidOrder{
				MarketId: 1,
				Buyer:    sdk.AccAddress("another_address_____").String(),
				Assets:   sdk.Coins{coin(99, "bender"), coin(0, "leela"), coin(1, "zoidberg")},
				Price:    coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets: ", "coin leela amount is not positive"},
		},
		{
			name: "negative amount in assets",
			order: BidOrder{
				MarketId: 1,
				Buyer:    sdk.AccAddress("another_address_____").String(),
				Assets:   sdk.Coins{coin(99, "bender"), coin(-1, "leela"), coin(1, "zoidberg")},
				Price:    coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets: ", "coin leela amount is not positive"},
		},
		{
			name: "invalid denom in assets",
			order: BidOrder{
				MarketId: 1,
				Buyer:    sdk.AccAddress("another_address_____").String(),
				Assets:   sdk.Coins{coin(99, "bender"), coin(1, "x"), coin(1, "zoidberg")},
				Price:    coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets: ", "invalid denom: x"},
		},
		{
			name: "nil assets",
			order: BidOrder{
				MarketId: 1,
				Buyer:    sdk.AccAddress("another_address_____").String(),
				Assets:   nil,
				Price:    coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets: must not be empty"},
		},
		{
			name: "empty assets",
			order: BidOrder{
				MarketId: 1,
				Buyer:    sdk.AccAddress("another_address_____").String(),
				Assets:   sdk.Coins{},
				Price:    coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets: must not be empty"},
		},
		{
			name: "price denom in assets",
			order: BidOrder{
				MarketId: 1,
				Buyer:    sdk.AccAddress("another_address_____").String(),
				Assets:   coins("99bender,2farnsworth,44amy"),
				Price:    coin(42, "farnsworth"),
			},
			exp: []string{"invalid assets: cannot contain price denom farnsworth"},
		},
		{
			name: "invalid denom in buyer settlement fees",
			order: BidOrder{
				MarketId:            1,
				Buyer:               sdk.AccAddress("another_address_____").String(),
				Assets:              coins("99bender"),
				Price:               coin(42, "farnsworth"),
				BuyerSettlementFees: sdk.Coins{coin(1, "farnsworth"), coin(2, "x")},
			},
			exp: []string{"invalid buyer settlement fees: ", "invalid denom: x"},
		},
		{
			name: "negative buyer settlement fees",
			order: BidOrder{
				MarketId:            1,
				Buyer:               sdk.AccAddress("another_address_____").String(),
				Assets:              coins("99bender"),
				Price:               coin(42, "farnsworth"),
				BuyerSettlementFees: sdk.Coins{coin(3, "farnsworth"), coin(-3, "nibbler")},
			},
			exp: []string{"invalid buyer settlement fees: ", "coin nibbler amount is not positive"},
		},
		{
			name: "multiple problems",
			order: BidOrder{
				Price:               coin(0, ""),
				BuyerSettlementFees: sdk.Coins{coin(0, "")},
			},
			exp: []string{
				"invalid market id: ",
				"invalid buyer: ",
				"invalid price: ",
				"invalid assets: ",
				"invalid buyer settlement fees: ",
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
