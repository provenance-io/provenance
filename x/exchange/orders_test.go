package exchange

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil/assertions"
)

// copyOrder creates a copy of the provided order.
func copyOrder(order *Order) *Order {
	if order == nil {
		return nil
	}

	rv := &Order{
		OrderId: order.OrderId,
	}
	if order.Order != nil {
		switch v := order.Order.(type) {
		case *Order_AskOrder:
			rv.Order = &Order_AskOrder{AskOrder: copyAskOrder(v.AskOrder)}
		case *Order_BidOrder:
			rv.Order = &Order_BidOrder{BidOrder: copyBidOrder(v.BidOrder)}
		case *unknownOrderType:
			rv.Order = &unknownOrderType{}
		default:
			panic(fmt.Sprintf("unknown order type %q", order.GetOrderType()))
		}
	}

	return rv
}

// copyAskOrder creates a copy of the provided ask order.
func copyAskOrder(askOrder *AskOrder) *AskOrder {
	if askOrder == nil {
		return nil
	}
	return &AskOrder{
		MarketId:                askOrder.MarketId,
		Seller:                  askOrder.Seller,
		Assets:                  copyCoin(askOrder.Assets),
		Price:                   copyCoin(askOrder.Price),
		SellerSettlementFlatFee: copyCoinP(askOrder.SellerSettlementFlatFee),
		AllowPartial:            askOrder.AllowPartial,
		ExternalId:              askOrder.ExternalId,
	}
}

// copyBidOrder creates a copy of the provided bid order.
func copyBidOrder(bidOrder *BidOrder) *BidOrder {
	if bidOrder == nil {
		return nil
	}
	return &BidOrder{
		MarketId:            bidOrder.MarketId,
		Buyer:               bidOrder.Buyer,
		Assets:              copyCoin(bidOrder.Assets),
		Price:               copyCoin(bidOrder.Price),
		BuyerSettlementFees: copyCoins(bidOrder.BuyerSettlementFees),
		AllowPartial:        bidOrder.AllowPartial,
		ExternalId:          bidOrder.ExternalId,
	}
}

// orderString is similar to %v except with easier to understand Coin and Int entries.
func orderString(order *Order) string {
	if order == nil {
		return "nil"
	}
	fields := make([]string, 1, 8)
	fields[0] = fmt.Sprintf("OrderId:%d", order.OrderId)

	switch {
	case order.IsAskOrder():
		fields = append(fields, fmt.Sprintf("AskOrder:%s", askOrderString(order.GetAskOrder())))
	case order.IsBidOrder():
		fields = append(fields, fmt.Sprintf("BidOrder:%s", bidOrderString(order.GetBidOrder())))
	default:
		if order.GetOrder() != nil {
			fields = append(fields, fmt.Sprintf("orderType:%q", order.GetOrderType()))
		} else {
			fields = append(fields, "Order:nil")
		}
	}

	return fmt.Sprintf("{%s}", strings.Join(fields, ", "))
}

// askOrderString is similar to %v except with easier to understand Coin and Int entries.
func askOrderString(askOrder *AskOrder) string {
	if askOrder == nil {
		return "nil"
	}

	fields := []string{
		fmt.Sprintf("MarketId:%d", askOrder.MarketId),
		fmt.Sprintf("Seller:%q", askOrder.Seller),
		fmt.Sprintf("Assets:%q", askOrder.Assets),
		fmt.Sprintf("Price:%q", askOrder.Price),
		fmt.Sprintf("SellerSettlementFlatFee:%s", coinPString(askOrder.SellerSettlementFlatFee)),
		fmt.Sprintf("AllowPartial:%t", askOrder.AllowPartial),
		fmt.Sprintf("ExternalID:%s", askOrder.ExternalId),
	}
	return fmt.Sprintf("{%s}", strings.Join(fields, ", "))
}

// bidOrderString is similar to %v except with easier to understand Coin and Int entries.
func bidOrderString(bidOrder *BidOrder) string {
	if bidOrder == nil {
		return "nil"
	}

	fields := []string{
		fmt.Sprintf("MarketId:%d", bidOrder.MarketId),
		fmt.Sprintf("Buyer:%q", bidOrder.Buyer),
		fmt.Sprintf("Assets:%q", bidOrder.Assets),
		fmt.Sprintf("Price:%q", bidOrder.Price),
		fmt.Sprintf("BuyerSettlementFees:%s", coinsString(bidOrder.BuyerSettlementFees)),
		fmt.Sprintf("AllowPartial:%t", bidOrder.AllowPartial),
		fmt.Sprintf("ExternalID:%s", bidOrder.ExternalId),
	}
	return fmt.Sprintf("{%s}", strings.Join(fields, ", "))
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

// newUnknownOrder returns a new order with the given id and an unknownOrderType.
func newUnknownOrder(orderID uint64) *Order {
	return &Order{OrderId: orderID, Order: &unknownOrderType{}}
}

// badSubTypeErr creates the expected error when a sub-order type is bad.
func badSubTypeErr(orderID uint64, badType string) string {
	return fmt.Sprintf("order %d has unknown sub-order type %s: does not implement SubOrderI", orderID, badType)
}

// nilSubTypeErr creates the expected error when a sub-order type is nil.
func nilSubTypeErr(orderID uint64) string {
	return badSubTypeErr(orderID, "<nil>")
}

// unknownSubTypeErr creates the expected error when a sub-order type is the unknownOrderType.
func unknownSubTypeErr(orderID uint64) string {
	return badSubTypeErr(orderID, "*exchange.unknownOrderType")
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

func TestValidateExternalID(t *testing.T) {
	tests := []struct {
		name       string
		externalID string
		expErr     string
	}{
		{
			name:       "empty",
			externalID: "",
			expErr:     "",
		},
		{
			name:       "max length",
			externalID: strings.Repeat("m", MaxExternalIDLength),
			expErr:     "",
		},
		{
			name:       "max length + 1",
			externalID: strings.Repeat("n", MaxExternalIDLength+1),
			expErr: fmt.Sprintf("invalid external id %q: max length %d",
				strings.Repeat("n", MaxExternalIDLength+1), MaxExternalIDLength),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = ValidateExternalID(tc.externalID)
			}
			require.NotPanics(t, testFunc, "ValidateExternalID(%q)", tc.externalID)
			assertions.AssertErrorValue(t, err, tc.expErr, "ValidateExternalID(%q)", tc.externalID)
		})
	}
}

func TestOrderSizes(t *testing.T) {
	// This unit test is mostly just to see the sizes of different orders and compare
	// that to the initial array size used in getOrderStoreKeyValue.
	expectedLength := 199 // = the 200 set in getOrderStoreKeyValue minus one for the order type byte.

	denomChars := "abcd"
	bigCoins := make(sdk.Coins, len(denomChars))
	for i := range bigCoins {
		str := fmt.Sprintf("%[1]d,000,000,000,000,000,000,000,000,000,00%[1]d", i) // i quetta + i
		amount := newInt(t, str)
		denom := strings.Repeat(denomChars[i:i+1], 128)
		bigCoins[i] = sdk.NewCoin(denom, amount)
	}
	coinP := func(denom string, amount int64) *sdk.Coin {
		rv := sdk.NewInt64Coin(denom, amount)
		return &rv
	}

	tests := []struct {
		name       string
		order      *Order
		expTooLong bool
	}{
		{
			name: "ask, normal",
			order: NewOrder(1).WithAsk(&AskOrder{
				MarketId:                1,
				Seller:                  sdk.AccAddress("seller______________").String(),
				Assets:                  sdk.NewInt64Coin("pm.sale.pool.sxyvff21dz5rrpamteeiin", 1),
				Price:                   sdk.NewInt64Coin("usd", 1_000_000),
				SellerSettlementFlatFee: coinP("nhash", 1_000_000_000_000),
				AllowPartial:            true,
			}),
			expTooLong: false,
		},
		{
			name: "ask, max market id",
			order: NewOrder(1).WithAsk(&AskOrder{
				MarketId:                4_294_967_295,
				Seller:                  sdk.AccAddress("seller______________").String(),
				Assets:                  sdk.NewInt64Coin("pm.sale.pool.sxyvff21dz5rrpamteeiin", 1),
				Price:                   sdk.NewInt64Coin("usd", 1_000_000),
				SellerSettlementFlatFee: coinP("nhash", 1_000_000_000_000),
				AllowPartial:            true,
			}),
			expTooLong: false,
		},
		{
			name: "ask, 32 byte addr",
			order: NewOrder(1).WithAsk(&AskOrder{
				MarketId:                1,
				Seller:                  sdk.AccAddress("seller__________________________").String(),
				Assets:                  sdk.NewInt64Coin("pm.sale.pool.sxyvff21dz5rrpamteeiin", 1),
				Price:                   sdk.NewInt64Coin("usd", 1_000_000),
				SellerSettlementFlatFee: coinP("nhash", 1_000_000_000_000),
				AllowPartial:            true,
			}),
			expTooLong: false,
		},
		{
			name: "ask, big",
			order: NewOrder(1).WithAsk(&AskOrder{
				MarketId:                4_294_967_295,
				Seller:                  sdk.AccAddress("seller__________________________").String(),
				Assets:                  bigCoins[0],
				Price:                   bigCoins[1],
				SellerSettlementFlatFee: &bigCoins[2],
				AllowPartial:            true,
			}),
			expTooLong: true,
		},
		{
			name: "bid, normal",
			order: NewOrder(1).WithBid(&BidOrder{
				MarketId:            1,
				Buyer:               sdk.AccAddress("buyer_______________").String(),
				Assets:              sdk.NewInt64Coin("pm.sale.pool.sxyvff21dz5rrpamteeiin", 1),
				Price:               sdk.NewInt64Coin("usd", 1_000_000),
				BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("nhash", 1_000_000_000_000)),
				AllowPartial:        true,
			}),
			expTooLong: false,
		},
		{
			name: "bid, max market id",
			order: NewOrder(1).WithBid(&BidOrder{
				MarketId:            4_294_967_295,
				Buyer:               sdk.AccAddress("buyer_______________").String(),
				Assets:              sdk.NewInt64Coin("pm.sale.pool.sxyvff21dz5rrpamteeiin", 1),
				Price:               sdk.NewInt64Coin("usd", 1_000_000),
				BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("nhash", 1_000_000_000_000)),
				AllowPartial:        true,
			}),
			expTooLong: false,
		},
		{
			name: "bid, 32 byte addr",
			order: NewOrder(1).WithBid(&BidOrder{
				MarketId:            1,
				Buyer:               sdk.AccAddress("buyer___________________________").String(),
				Assets:              sdk.NewInt64Coin("pm.sale.pool.sxyvff21dz5rrpamteeiin", 1),
				Price:               sdk.NewInt64Coin("usd", 1_000_000),
				BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("nhash", 1_000_000_000_000)),
				AllowPartial:        true,
			}),
			expTooLong: false,
		},
		{
			name: "bid, big",
			order: NewOrder(1).WithBid(&BidOrder{
				MarketId:            4_294_967_295,
				Buyer:               sdk.AccAddress("buyer___________________________").String(),
				Assets:              bigCoins[0],
				Price:               bigCoins[1],
				BuyerSettlementFees: bigCoins[2:],
				AllowPartial:        true,
			}),
			expTooLong: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var data []byte
			var err error
			switch {
			case tc.order.IsAskOrder():
				askOrder := tc.order.GetAskOrder()
				fields := []string{
					fmt.Sprintf("MarketID:     %d", askOrder.MarketId),
					fmt.Sprintf("Seller:       %q", askOrder.Seller),
					fmt.Sprintf("Assets:       %q", askOrder.Assets),
					fmt.Sprintf("Price:        %q", askOrder.Price),
					fmt.Sprintf("AllowPartial: %t", askOrder.AllowPartial),
					fmt.Sprintf("SellerSettlementFlatFee: %s", coinPString(askOrder.SellerSettlementFlatFee)),
				}
				t.Logf("AskOrder:\n%s\n", strings.Join(fields, "\n"))
				data, err = askOrder.Marshal()
			case tc.order.IsBidOrder():
				bidOrder := tc.order.GetBidOrder()
				fields := []string{
					fmt.Sprintf("MarketID:     %d", bidOrder.MarketId),
					fmt.Sprintf("Buyer:        %q", bidOrder.Buyer),
					fmt.Sprintf("Assets:       %q", bidOrder.Assets),
					fmt.Sprintf("Price:        %q", bidOrder.Price),
					fmt.Sprintf("AllowPartial: %t", bidOrder.AllowPartial),
					fmt.Sprintf("BuyerSettlementFees: %s", coinsString(bidOrder.BuyerSettlementFees)),
				}
				t.Logf("BidOrder:\n%s\n", strings.Join(fields, "\n"))
				data, err = bidOrder.Marshal()
			default:
				t.Fatalf("unknown order type %T", tc.order.GetOrder())
			}
			require.NoError(t, err, "%s.Marshal()", tc.order.GetOrderType())

			length := len(data)
			t.Logf("Data Length: %d", length)
			if tc.expTooLong {
				assert.Greater(t, length, expectedLength, "%s.Marshal() length", tc.order.GetOrderType())
			} else {
				assert.LessOrEqual(t, len(data), expectedLength, "%s.Marshal() length", tc.order.GetOrderType())
			}
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
			order: newUnknownOrder(4),
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
			order: newUnknownOrder(4),
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
			expErr: nilSubTypeErr(3),
		},
		{
			name:   "unknown order type",
			order:  newUnknownOrder(4),
			expErr: unknownSubTypeErr(4),
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
			expPanic: nilSubTypeErr(3),
		},
		{
			name:     "unknown order type",
			order:    newUnknownOrder(4),
			expPanic: unknownSubTypeErr(4),
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
			expPanic: nilSubTypeErr(3),
		},
		{
			name:     "unknown order type",
			order:    newUnknownOrder(4),
			expPanic: unknownSubTypeErr(4),
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
			expPanic: nilSubTypeErr(3),
		},
		{
			name:     "unknown order type",
			order:    newUnknownOrder(4),
			expPanic: unknownSubTypeErr(4),
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
			expPanic: nilSubTypeErr(3),
		},
		{
			name:     "unknown order type",
			order:    newUnknownOrder(4),
			expPanic: unknownSubTypeErr(4),
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
			expPanic: nilSubTypeErr(3),
		},
		{
			name:     "unknown order type",
			order:    newUnknownOrder(4),
			expPanic: unknownSubTypeErr(4),
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
			expPanic: nilSubTypeErr(3),
		},
		{
			name:     "unknown order type",
			order:    newUnknownOrder(4),
			expPanic: unknownSubTypeErr(4),
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

func TestOrder_GetExternalID(t *testing.T) {
	tests := []struct {
		name     string
		order    *Order
		expected string
		expPanic string
	}{
		{
			name:     "AskOrder",
			order:    NewOrder(1).WithAsk(&AskOrder{ExternalId: "ask12345"}),
			expected: "ask12345",
		},
		{
			name:     "BidOrder",
			order:    NewOrder(2).WithBid(&BidOrder{ExternalId: "bid987654"}),
			expected: "bid987654",
		},
		{
			name:     "nil inside order",
			order:    NewOrder(3),
			expPanic: nilSubTypeErr(3),
		},
		{
			name:     "unknown order type",
			order:    newUnknownOrder(4),
			expPanic: unknownSubTypeErr(4),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = tc.order.GetExternalID()
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "GetExternalID()")
			assert.Equal(t, tc.expected, actual, "GetExternalID() result")
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
			order:    newUnknownOrder(4),
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
			expPanic: nilSubTypeErr(3),
		},
		{
			name:     "unknown order type",
			order:    newUnknownOrder(4),
			expPanic: unknownSubTypeErr(4),
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
			expPanic: nilSubTypeErr(3),
		},
		{
			name:     "unknown order type",
			order:    newUnknownOrder(4),
			expPanic: unknownSubTypeErr(4),
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
			exp:   []string{"invalid order id", "cannot be zero"},
		},
		{
			name:  "nil sub-order",
			Order: NewOrder(1),
			exp:   []string{nilSubTypeErr(1)},
		},
		{
			name:  "unknown sub-order type",
			Order: newUnknownOrder(3),
			exp:   []string{unknownSubTypeErr(3)},
		},
		{
			name:  "ask order error",
			Order: NewOrder(1).WithAsk(&AskOrder{MarketId: 0, Price: zeroCoin}),
			exp:   []string{"invalid market id", "cannot be zero"},
		},
		{
			name:  "bid order error",
			Order: NewOrder(1).WithBid(&BidOrder{MarketId: 0, Price: zeroCoin}),
			exp:   []string{"invalid market id", "cannot be zero"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.Order.Validate()
			assertions.AssertErrorContents(t, err, tc.exp, "Validate() error")
		})
	}
}

func TestOrder_Split(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	askOrder := func(orderID uint64, assetAmt, priceAmt int64, fees ...sdk.Coin) *Order {
		askOrder := &AskOrder{
			MarketId:     55,
			Seller:       "samuel",
			Assets:       coin(assetAmt, "apple"),
			Price:        coin(priceAmt, "peach"),
			AllowPartial: true,
		}
		if len(fees) > 1 {
			t.Fatalf("a max of 1 fee can be provided to askOrder, actual: %s", sdk.Coins(fees))
		}
		if len(fees) > 0 {
			askOrder.SellerSettlementFlatFee = &fees[0]
		}
		return NewOrder(orderID).WithAsk(askOrder)
	}
	bidOrder := func(orderID uint64, assetAmt, priceAmt int64, fees ...sdk.Coin) *Order {
		bidOrder := &BidOrder{
			MarketId:     55,
			Buyer:        "brian",
			Assets:       coin(assetAmt, "apple"),
			Price:        coin(priceAmt, "peach"),
			AllowPartial: true,
		}
		if len(fees) > 0 {
			bidOrder.BuyerSettlementFees = fees
		}
		return NewOrder(orderID).WithBid(bidOrder)
	}

	tests := []struct {
		name            string
		order           *Order
		assetsFilledAmt sdkmath.Int
		expFilled       *Order
		expUnfilled     *Order
		expErr          string
		expPanic        string
	}{
		{
			name:            "nil inside order",
			order:           NewOrder(88),
			assetsFilledAmt: sdkmath.NewInt(0),
			expPanic:        nilSubTypeErr(88),
		},
		{
			name:            "unknown inside order",
			order:           newUnknownOrder(89),
			assetsFilledAmt: sdkmath.NewInt(0),
			expPanic:        unknownSubTypeErr(89),
		},
		{
			name:            "assets filled is negative: ask",
			order:           askOrder(3, 10, 100),
			assetsFilledAmt: sdkmath.NewInt(-1),
			expErr:          "cannot split ask order 3 having asset \"10apple\" at \"-1apple\": amount filled not positive",
		},
		{
			name:            "assets filled is negative: bid",
			order:           bidOrder(4, 10, 100),
			assetsFilledAmt: sdkmath.NewInt(-1),
			expErr:          "cannot split bid order 4 having asset \"10apple\" at \"-1apple\": amount filled not positive",
		},
		{
			name:            "assets filled is zero: ask",
			order:           askOrder(9, 10, 100),
			assetsFilledAmt: sdkmath.NewInt(0),
			expErr:          "cannot split ask order 9 having asset \"10apple\" at \"0apple\": amount filled not positive",
		},
		{
			name:            "assets filled is zero: bid",
			order:           bidOrder(10, 10, 100),
			assetsFilledAmt: sdkmath.NewInt(0),
			expErr:          "cannot split bid order 10 having asset \"10apple\" at \"0apple\": amount filled not positive",
		},
		{
			name:            "assets filled equals order assets: ask",
			order:           askOrder(7, 10, 100),
			assetsFilledAmt: sdkmath.NewInt(10),
			expErr:          "cannot split ask order 7 having asset \"10apple\" at \"10apple\": amount filled equals order assets",
		},
		{
			name:            "assets filled equals order assets: bid",
			order:           bidOrder(8, 10, 100),
			assetsFilledAmt: sdkmath.NewInt(10),
			expErr:          "cannot split bid order 8 having asset \"10apple\" at \"10apple\": amount filled equals order assets",
		},
		{
			name:            "assets filled is more than order assets: ask",
			order:           askOrder(5, 10, 100),
			assetsFilledAmt: sdkmath.NewInt(11),
			expErr:          "cannot split ask order 5 having asset \"10apple\" at \"11apple\": overfilled",
		},
		{
			name:            "assets filled is more than order assets: bid",
			order:           bidOrder(6, 10, 100),
			assetsFilledAmt: sdkmath.NewInt(11),
			expErr:          "cannot split bid order 6 having asset \"10apple\" at \"11apple\": overfilled",
		},
		{
			name:            "partial not allowed: ask",
			order:           NewOrder(1).WithAsk(&AskOrder{AllowPartial: false, Assets: coin(2, "peach")}),
			assetsFilledAmt: sdkmath.NewInt(1),
			expErr:          "cannot split ask order 1 having assets \"2peach\" at \"1peach\": order does not allow partial fulfillment",
		},
		{
			name:            "partial not allowed: bid",
			order:           NewOrder(2).WithBid(&BidOrder{AllowPartial: false, Assets: coin(2, "peach")}),
			assetsFilledAmt: sdkmath.NewInt(1),
			expErr:          "cannot split bid order 2 having assets \"2peach\" at \"1peach\": order does not allow partial fulfillment",
		},
		{
			name:            "price not divisible: ask",
			order:           askOrder(11, 70, 501),
			assetsFilledAmt: sdkmath.NewInt(7),
			expErr: "ask order 11 having assets \"70apple\" cannot be partially filled " +
				"by \"7apple\": price \"501peach\" is not evenly divisible",
		},
		{
			name:            "price not divisible: bid",
			order:           bidOrder(12, 70, 501),
			assetsFilledAmt: sdkmath.NewInt(7),
			expErr: "bid order 12 having assets \"70apple\" cannot be partially filled " +
				"by \"7apple\": price \"501peach\" is not evenly divisible",
		},
		{
			name:            "fee not divisible: ask",
			order:           askOrder(13, 70, 500, coin(23, "fig")),
			assetsFilledAmt: sdkmath.NewInt(7),
			expErr: "ask order 13 having assets \"70apple\" cannot be partially filled " +
				"by \"7apple\": fee \"23fig\" is not evenly divisible",
		},
		{
			name:            "fees not divisible: bid",
			order:           bidOrder(14, 70, 500, coin(20, "fig"), coin(34, "grape")),
			assetsFilledAmt: sdkmath.NewInt(7),
			expErr: "bid order 14 having assets \"70apple\" cannot be partially filled " +
				"by \"7apple\": fee \"34grape\" is not evenly divisible",
		},
		{
			name:            "no fees: ask",
			order:           askOrder(21, 70, 500),
			assetsFilledAmt: sdkmath.NewInt(7),
			expFilled:       askOrder(21, 7, 50),
			expUnfilled:     askOrder(21, 63, 450),
		},
		{
			name:            "no fees: bid",
			order:           bidOrder(22, 70, 500),
			assetsFilledAmt: sdkmath.NewInt(7),
			expFilled:       bidOrder(22, 7, 50),
			expUnfilled:     bidOrder(22, 63, 450),
		},
		{
			name:            "with fees: ask",
			order:           askOrder(23, 10, 500, coin(5, "fig")),
			assetsFilledAmt: sdkmath.NewInt(8),
			expFilled:       askOrder(23, 8, 400, coin(4, "fig")),
			expUnfilled:     askOrder(23, 2, 100, coin(1, "fig")),
		},
		{
			name:            "with fees: bid",
			order:           bidOrder(24, 10, 500, coin(5, "fig"), coin(15, "grape")),
			assetsFilledAmt: sdkmath.NewInt(8),
			expFilled:       bidOrder(24, 8, 400, coin(4, "fig"), coin(12, "grape")),
			expUnfilled:     bidOrder(24, 2, 100, coin(1, "fig"), coin(3, "grape")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var filled, unfilled *Order
			var err error
			testFunc := func() {
				filled, unfilled, err = tc.order.Split(tc.assetsFilledAmt)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "Split(%s)", tc.assetsFilledAmt)
			assertions.AssertErrorValue(t, err, tc.expErr, "Split(%s) error", tc.assetsFilledAmt)
			assert.Equal(t, tc.expFilled, filled, "Split(%s) filled order", tc.assetsFilledAmt)
			assert.Equal(t, tc.expUnfilled, unfilled, "Split(%s) unfilled order", tc.assetsFilledAmt)
			// If the expected filled isn't null, but unfilled is, make sure that the original was returned.
			if tc.expFilled != nil && tc.expUnfilled == nil {
				assert.Same(t, tc.order, filled, "Split(%s) filled order address", tc.assetsFilledAmt)
			}
			// If the expected unfilled isn't null, but filled is, make sure that the original was returned.
			if tc.expUnfilled != nil && tc.expFilled == nil {
				assert.Same(t, tc.order, unfilled, "Split(%s) unfilled order address", tc.assetsFilledAmt)
			}
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
	largeAmt := newInt(t, "25,000,000,000,000,000,000,000")
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
	largeAmt := newInt(t, "25,000,000,000,000,000,000,000")
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
	largeAmt := newInt(t, "25,000,000,000,000,000,000,000")
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

func TestAskOrder_GetExternalID(t *testing.T) {
	tests := []struct {
		name  string
		order AskOrder
		exp   string
	}{
		{name: "empty", order: AskOrder{ExternalId: ""}, exp: ""},
		{name: "something", order: AskOrder{ExternalId: "something"}, exp: "something"},
		{name: "a uuid", order: AskOrder{ExternalId: "36585FC1-C11D-42A4-B1F7-92B0D8229BC7"}, exp: "36585FC1-C11D-42A4-B1F7-92B0D8229BC7"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = tc.order.GetExternalID()
			}
			require.NotPanics(t, testFunc, "GetExternalID()")
			assert.Equal(t, tc.exp, actual, "GetExternalID() result")
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
			exp: []string{"invalid market id", "cannot be zero"},
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

func TestAskOrder_CopyChange(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	coinP := func(amount int64, denom string) *sdk.Coin {
		rv := coin(amount, denom)
		return &rv
	}

	tests := []struct {
		name      string
		order     AskOrder
		newAssets sdk.Coin
		newPrice  sdk.Coin
		newFee    *sdk.Coin
		expected  *AskOrder
	}{
		{
			name: "new assets",
			order: AskOrder{
				MarketId:                3,
				Seller:                  "sseelleerr",
				Assets:                  coin(8, "apple"),
				Price:                   coin(55, "peach"),
				SellerSettlementFlatFee: coinP(12, "fig"),
				AllowPartial:            true,
			},
			newAssets: coin(14, "avocado"),
			newPrice:  coin(55, "peach"),
			newFee:    coinP(12, "fig"),
			expected: &AskOrder{
				MarketId:                3,
				Seller:                  "sseelleerr",
				Assets:                  coin(14, "avocado"),
				Price:                   coin(55, "peach"),
				SellerSettlementFlatFee: coinP(12, "fig"),
				AllowPartial:            true,
			},
		},
		{
			name: "new price",
			order: AskOrder{
				MarketId:                99,
				Seller:                  "sseeLLeerr",
				Assets:                  coin(8, "apple"),
				Price:                   coin(55, "peach"),
				SellerSettlementFlatFee: coinP(12, "fig"),
				AllowPartial:            false,
			},
			newAssets: coin(8, "apple"),
			newPrice:  coin(38, "plum"),
			newFee:    coinP(12, "fig"),
			expected: &AskOrder{
				MarketId:                99,
				Seller:                  "sseeLLeerr",
				Assets:                  coin(8, "apple"),
				Price:                   coin(38, "plum"),
				SellerSettlementFlatFee: coinP(12, "fig"),
				AllowPartial:            false,
			},
		},
		{
			name: "new fees",
			order: AskOrder{
				MarketId:                33,
				Seller:                  "SseelleerR",
				Assets:                  coin(8, "apple"),
				Price:                   coin(55, "peach"),
				SellerSettlementFlatFee: coinP(12, "fig"),
				AllowPartial:            true,
			},
			newAssets: coin(8, "apple"),
			newPrice:  coin(55, "peach"),
			newFee:    coinP(88, "grape"),
			expected: &AskOrder{
				MarketId:                33,
				Seller:                  "SseelleerR",
				Assets:                  coin(8, "apple"),
				Price:                   coin(55, "peach"),
				SellerSettlementFlatFee: coinP(88, "grape"),
				AllowPartial:            true,
			},
		},
		{
			name: "new everything",
			order: AskOrder{
				MarketId:                34,
				Seller:                  "SSEELLEERR",
				Assets:                  coin(8, "apple"),
				Price:                   coin(55, "peach"),
				SellerSettlementFlatFee: coinP(12, "fig"),
				AllowPartial:            false,
			},
			newAssets: coin(14, "avocado"),
			newPrice:  coin(38, "plum"),
			newFee:    coinP(88, "grape"),
			expected: &AskOrder{
				MarketId:                34,
				Seller:                  "SSEELLEERR",
				Assets:                  coin(14, "avocado"),
				Price:                   coin(38, "plum"),
				SellerSettlementFlatFee: coinP(88, "grape"),
				AllowPartial:            false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual *AskOrder
			testFunc := func() {
				actual = tc.order.CopyChange(tc.newAssets, tc.newPrice, tc.newFee)
			}
			require.NotPanics(t, testFunc, "CopyChange")
			if !assert.Equal(t, tc.expected, actual, "CopyChange result") {
				t.Logf("  Actual: %s", askOrderString(actual))
				t.Logf("Expected: %s", askOrderString(tc.expected))
			}
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
	largeAmt := newInt(t, "25,000,000,000,000,000,000,000")
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
	largeAmt := newInt(t, "25,000,000,000,000,000,000,000")
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
	largeAmt := newInt(t, "25,000,000,000,000,000,000,000")
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

func TestBidOrder_GetExternalID(t *testing.T) {
	tests := []struct {
		name  string
		order BidOrder
		exp   string
	}{
		{name: "empty", order: BidOrder{ExternalId: ""}, exp: ""},
		{name: "something", order: BidOrder{ExternalId: "something"}, exp: "something"},
		{name: "a uuid", order: BidOrder{ExternalId: "36585FC1-C11D-42A4-B1F7-92B0D8229BC7"}, exp: "36585FC1-C11D-42A4-B1F7-92B0D8229BC7"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = tc.order.GetExternalID()
			}
			require.NotPanics(t, testFunc, "GetExternalID()")
			assert.Equal(t, tc.exp, actual, "GetExternalID() result")
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
			exp: []string{"invalid market id: cannot be zero"},
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

func TestBidOrder_CopyChange(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}

	tests := []struct {
		name      string
		order     BidOrder
		newAssets sdk.Coin
		newPrice  sdk.Coin
		newFees   sdk.Coins
		expected  *BidOrder
	}{
		{
			name: "new assets",
			order: BidOrder{
				MarketId:            3,
				Buyer:               "bbuuyyeerr",
				Assets:              coin(8, "apple"),
				Price:               coin(55, "peach"),
				BuyerSettlementFees: sdk.Coins{coin(12, "fig")},
				AllowPartial:        true,
			},
			newAssets: coin(14, "avocado"),
			newPrice:  coin(55, "peach"),
			newFees:   sdk.Coins{coin(12, "fig")},
			expected: &BidOrder{
				MarketId:            3,
				Buyer:               "bbuuyyeerr",
				Assets:              coin(14, "avocado"),
				Price:               coin(55, "peach"),
				BuyerSettlementFees: sdk.Coins{coin(12, "fig")},
				AllowPartial:        true,
			},
		},
		{
			name: "new price",
			order: BidOrder{
				MarketId:            99,
				Buyer:               "bbuuyyeerr",
				Assets:              coin(8, "apple"),
				Price:               coin(55, "peach"),
				BuyerSettlementFees: sdk.Coins{coin(12, "fig")},
				AllowPartial:        false,
			},
			newAssets: coin(8, "apple"),
			newPrice:  coin(38, "plum"),
			newFees:   sdk.Coins{coin(12, "fig")},
			expected: &BidOrder{
				MarketId:            99,
				Buyer:               "bbuuyyeerr",
				Assets:              coin(8, "apple"),
				Price:               coin(38, "plum"),
				BuyerSettlementFees: sdk.Coins{coin(12, "fig")},
				AllowPartial:        false,
			},
		},
		{
			name: "new fees",
			order: BidOrder{
				MarketId:            33,
				Buyer:               "bbuuyyeerr",
				Assets:              coin(8, "apple"),
				Price:               coin(55, "peach"),
				BuyerSettlementFees: sdk.Coins{coin(12, "fig")},
				AllowPartial:        true,
			},
			newAssets: coin(8, "apple"),
			newPrice:  coin(55, "peach"),
			newFees:   sdk.Coins{coin(88, "grape")},
			expected: &BidOrder{
				MarketId:            33,
				Buyer:               "bbuuyyeerr",
				Assets:              coin(8, "apple"),
				Price:               coin(55, "peach"),
				BuyerSettlementFees: sdk.Coins{coin(88, "grape")},
				AllowPartial:        true,
			},
		},
		{
			name: "new everything",
			order: BidOrder{
				MarketId:            34,
				Buyer:               "BBUUYYEERR",
				Assets:              coin(8, "apple"),
				Price:               coin(55, "peach"),
				BuyerSettlementFees: sdk.Coins{coin(12, "fig")},
				AllowPartial:        false,
			},
			newAssets: coin(14, "avocado"),
			newPrice:  coin(38, "plum"),
			newFees:   sdk.Coins{coin(88, "grape"), coin(123, "honeydew")},
			expected: &BidOrder{
				MarketId:            34,
				Buyer:               "BBUUYYEERR",
				Assets:              coin(14, "avocado"),
				Price:               coin(38, "plum"),
				BuyerSettlementFees: sdk.Coins{coin(88, "grape"), coin(123, "honeydew")},
				AllowPartial:        false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual *BidOrder
			testFunc := func() {
				actual = tc.order.CopyChange(tc.newAssets, tc.newPrice, tc.newFees)
			}
			require.NotPanics(t, testFunc, "CopyChange")
			if !assert.Equal(t, tc.expected, actual, "CopyChange result") {
				t.Logf("  Actual: %s", bidOrderString(actual))
				t.Logf("Expected: %s", bidOrderString(tc.expected))
			}
		})
	}
}

func TestNewFilledOrder(t *testing.T) {
	expected := &FilledOrder{
		order:       NewOrder(8),
		actualPrice: sdk.NewInt64Coin("prune", 123),
		actualFees:  sdk.NewCoins(sdk.NewInt64Coin("fig", 999)),
	}
	var actual *FilledOrder
	testFunc := func() {
		actual = NewFilledOrder(expected.order, expected.actualPrice, expected.actualFees)
	}
	require.NotPanics(t, testFunc, "NewFilledOrder")
	assert.Equal(t, expected, actual, "NewFilledOrder result")
}

func TestFilledOrderGetters(t *testing.T) {
	askOrder := &AskOrder{
		MarketId:                333,
		Seller:                  "SEllER",
		Assets:                  sdk.NewInt64Coin("apple", 55),
		Price:                   sdk.NewInt64Coin("peach", 111),
		SellerSettlementFlatFee: &sdk.Coin{Denom: "fig", Amount: sdkmath.NewInt(8)},
		AllowPartial:            true,
		ExternalId:              "ask order abc",
	}
	ask := NewOrder(51).WithAsk(askOrder)
	askActualPrice := sdk.NewInt64Coin("peach", 123)
	askActualFees := sdk.NewCoins(sdk.NewInt64Coin("fig", 13))
	filledAsk := NewFilledOrder(ask, askActualPrice, askActualFees)

	bidOrder := &BidOrder{
		MarketId:            444,
		Buyer:               "BUyER",
		Assets:              sdk.NewInt64Coin("apple", 56),
		Price:               sdk.NewInt64Coin("peach", 112),
		BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("fig", 9)),
		AllowPartial:        true,
		ExternalId:          "bid order def",
	}
	bid := NewOrder(52).WithBid(bidOrder)
	bidActualPrice := sdk.NewInt64Coin("peach", 124)
	bidActualFees := sdk.NewCoins(sdk.NewInt64Coin("fig", 14))
	filledBid := NewFilledOrder(bid, bidActualPrice, bidActualFees)

	tests := []struct {
		name   string
		getter func(fo *FilledOrder) interface{}
		expAsk interface{}
		expBid interface{}
	}{
		{
			name:   "GetOriginalOrder",
			getter: func(of *FilledOrder) interface{} { return of.GetOriginalOrder() },
			expAsk: ask,
			expBid: bid,
		},
		{
			name:   "GetOrderID",
			getter: func(of *FilledOrder) interface{} { return of.GetOrderID() },
			expAsk: ask.OrderId,
			expBid: bid.OrderId,
		},
		{
			name:   "IsAskOrder",
			getter: func(of *FilledOrder) interface{} { return of.IsAskOrder() },
			expAsk: true,
			expBid: false,
		},
		{
			name:   "IsBidOrder",
			getter: func(of *FilledOrder) interface{} { return of.IsBidOrder() },
			expAsk: false,
			expBid: true,
		},
		{
			name:   "GetMarketID",
			getter: func(of *FilledOrder) interface{} { return of.GetMarketID() },
			expAsk: askOrder.MarketId,
			expBid: bidOrder.MarketId,
		},
		{
			name:   "GetOwner",
			getter: func(of *FilledOrder) interface{} { return of.GetOwner() },
			expAsk: askOrder.Seller,
			expBid: bidOrder.Buyer,
		},
		{
			name:   "GetAssets",
			getter: func(of *FilledOrder) interface{} { return of.GetAssets() },
			expAsk: askOrder.Assets,
			expBid: bidOrder.Assets,
		},
		{
			name:   "GetPrice",
			getter: func(of *FilledOrder) interface{} { return of.GetPrice() },
			expAsk: askActualPrice,
			expBid: bidActualPrice,
		},
		{
			name:   "GetOriginalPrice",
			getter: func(of *FilledOrder) interface{} { return of.GetOriginalPrice() },
			expAsk: askOrder.Price,
			expBid: bidOrder.Price,
		},
		{
			name:   "GetSettlementFees",
			getter: func(of *FilledOrder) interface{} { return of.GetSettlementFees() },
			expAsk: askActualFees,
			expBid: bidActualFees,
		},
		{
			name:   "GetOriginalSettlementFees",
			getter: func(of *FilledOrder) interface{} { return of.GetOriginalSettlementFees() },
			expAsk: sdk.Coins{*askOrder.SellerSettlementFlatFee},
			expBid: bidOrder.BuyerSettlementFees,
		},
		{
			name:   "PartialFillAllowed",
			getter: func(of *FilledOrder) interface{} { return of.PartialFillAllowed() },
			expAsk: askOrder.AllowPartial,
			expBid: bidOrder.AllowPartial,
		},
		{
			name:   "GetExternalID",
			getter: func(of *FilledOrder) interface{} { return of.GetExternalID() },
			expAsk: askOrder.ExternalId,
			expBid: bidOrder.ExternalId,
		},
		{
			name:   "GetOrderType",
			getter: func(of *FilledOrder) interface{} { return of.GetOrderType() },
			expAsk: OrderTypeAsk,
			expBid: OrderTypeBid,
		},
		{
			name:   "GetOrderTypeByte",
			getter: func(of *FilledOrder) interface{} { return of.GetOrderTypeByte() },
			expAsk: OrderTypeByteAsk,
			expBid: OrderTypeByteBid,
		},
		{
			name:   "GetHoldAmount",
			getter: func(of *FilledOrder) interface{} { return of.GetHoldAmount() },
			expAsk: askOrder.GetHoldAmount(),
			expBid: bidOrder.GetHoldAmount(),
		},
		{
			name:   "Validate",
			getter: func(of *FilledOrder) interface{} { return of.Validate() },
			expAsk: error(nil),
			expBid: error(nil),
		},
	}

	tester := func(name string, of *FilledOrder, getter func(*FilledOrder) interface{}, expected interface{}) func(t *testing.T) {
		return func(t *testing.T) {
			var actual interface{}
			testFunc := func() {
				actual = getter(of)
			}
			require.NotPanics(t, testFunc, "%s()")
			assert.Equal(t, expected, actual, "%s() result", name)
		}
	}

	for _, tc := range tests {
		t.Run(tc.name+": ask", tester(tc.name, filledAsk, tc.getter, tc.expAsk))
		t.Run(tc.name+": bid", tester(tc.name, filledBid, tc.getter, tc.expBid))
	}
}
