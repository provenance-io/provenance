package exchange

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/testutil/assertions"
)

// Annoyingly, sdkmath.NewInt(0) and sdkmath.ZeroInt() are not internally equal to an Int that
// started with a value and was reduced to zero.
// In other words, assert.Equal(t, sdkmath.ZeroInt(), sdkmath.NewInt(1).SubRaw(1)) fails.
// With those, Int.abs = (big.nat) <nil>.
// With this, Int.abs = (big.nat){}.
// So when an object has an sdkmath.Int that should have been reduced to zero, you'll need to use this.
var ZeroAmtAfterSub = sdkmath.NewInt(1).SubRaw(1)

// copyOrderSplit creates a copy of this order split.
// Unlike the other copiers in here, the Order is not deep copied, it will be the same reference.
func copyOrderSplit(split *OrderSplit) *OrderSplit {
	if split == nil {
		return nil
	}

	return &OrderSplit{
		// just copying the reference here to prevent infinite recursion.
		Order:  split.Order,
		Assets: copyCoin(split.Assets),
		Price:  copyCoin(split.Price),
	}
}

// copyOrderSplits copies a slice of order splits.
func copyOrderSplits(splits []*OrderSplit) []*OrderSplit {
	if splits == nil {
		return nil
	}

	rv := make([]*OrderSplit, len(splits))
	for i, split := range splits {
		rv[i] = copyOrderSplit(split)
	}
	return rv
}

// copyOrderFulfillment returns a deep copy of an order fulfillement.
func copyOrderFulfillment(f *OrderFulfillment) *OrderFulfillment {
	if f == nil {
		return nil
	}

	return &OrderFulfillment{
		Order:             copyOrder(f.Order),
		Splits:            copyOrderSplits(f.Splits),
		AssetsFilledAmt:   copySDKInt(f.AssetsFilledAmt),
		AssetsUnfilledAmt: copySDKInt(f.AssetsUnfilledAmt),
		PriceAppliedAmt:   copySDKInt(f.PriceAppliedAmt),
		PriceLeftAmt:      copySDKInt(f.PriceLeftAmt),
		IsFinalized:       f.IsFinalized,
		FeesToPay:         copyCoins(f.FeesToPay),
		OrderFeesLeft:     copyCoins(f.OrderFeesLeft),
		PriceFilledAmt:    copySDKInt(f.PriceFilledAmt),
		PriceUnfilledAmt:  copySDKInt(f.PriceUnfilledAmt),
	}
}

// orderSplitString is similar to %v except with easier to understand Coin and Int entries.
func orderSplitString(s *OrderSplit) string {
	if s == nil {
		return "nil"
	}

	fields := []string{
		// Just using superficial info for the order to prevent infinite loops.
		fmt.Sprintf("Order:{OrderID:%d,OrderType:%s,...}", s.Order.GetOrderID(), s.Order.GetOrderType()),
		fmt.Sprintf("Assets:%q", s.Assets),
		fmt.Sprintf("Price:%q", s.Price),
	}
	return fmt.Sprintf("{%s}", strings.Join(fields, ", "))
}

// orderSplitsString is similar to %v except with easier to understand Coin and Int entries.
func orderSplitsString(splits []*OrderSplit) string {
	if splits == nil {
		return "nil"
	}
	vals := make([]string, len(splits))
	for i, s := range splits {
		vals[i] = orderSplitString(s)
	}
	return fmt.Sprintf("[%s]", strings.Join(vals, ", "))
}

// orderFulfillmentString is similar to %v except with easier to understand Coin and Int entries.
func orderFulfillmentString(f *OrderFulfillment) string {
	if f == nil {
		return "nil"
	}

	fields := []string{
		fmt.Sprintf("Order:%s", orderString(f.Order)),
		fmt.Sprintf("Splits:%s", orderSplitsString(f.Splits)),
		fmt.Sprintf("AssetsFilledAmt:%s", f.AssetsFilledAmt),
		fmt.Sprintf("AssetsUnfilledAmt:%s", f.AssetsUnfilledAmt),
		fmt.Sprintf("PriceAppliedAmt:%s", f.PriceAppliedAmt),
		fmt.Sprintf("PriceLeftAmt:%s", f.PriceLeftAmt),
		fmt.Sprintf("IsFinalized:%t", f.IsFinalized),
		fmt.Sprintf("FeesToPay:%s", coinsString(f.FeesToPay)),
		fmt.Sprintf("OrderFeesLeft:%s", coinsString(f.OrderFeesLeft)),
		fmt.Sprintf("PriceFilledAmt:%s", f.PriceFilledAmt),
		fmt.Sprintf("PriceUnfilledAmt:%s", f.PriceUnfilledAmt),
	}
	return fmt.Sprintf("{%s}", strings.Join(fields, ", "))
}

func TestNewOrderFulfillment(t *testing.T) {
	tests := []struct {
		name     string
		order    *Order
		expected *OrderFulfillment
		expPanic string
	}{
		{
			name:     "nil sub-order",
			order:    NewOrder(1),
			expPanic: nilSubTypeErr(1),
		},
		{
			name: "ask order",
			order: NewOrder(2).WithAsk(&AskOrder{
				MarketId: 10,
				Assets:   sdk.NewInt64Coin("adolla", 92),
				Price:    sdk.NewInt64Coin("pdolla", 15),
			}),
			expected: &OrderFulfillment{
				Order: &Order{
					OrderId: 2,
					Order: &Order_AskOrder{
						AskOrder: &AskOrder{
							MarketId: 10,
							Assets:   sdk.NewInt64Coin("adolla", 92),
							Price:    sdk.NewInt64Coin("pdolla", 15),
						},
					},
				},
				Splits:            nil,
				AssetsFilledAmt:   sdkmath.ZeroInt(),
				AssetsUnfilledAmt: sdkmath.NewInt(92),
				PriceAppliedAmt:   sdkmath.ZeroInt(),
				PriceLeftAmt:      sdkmath.NewInt(15),
				IsFinalized:       false,
				FeesToPay:         nil,
				OrderFeesLeft:     nil,
				PriceFilledAmt:    sdkmath.ZeroInt(),
				PriceUnfilledAmt:  sdkmath.ZeroInt(),
			},
		},
		{
			name: "bid order",
			order: NewOrder(3).WithBid(&BidOrder{
				MarketId: 11,
				Assets:   sdk.NewInt64Coin("adolla", 93),
				Price:    sdk.NewInt64Coin("pdolla", 16),
			}),
			expected: &OrderFulfillment{
				Order: &Order{
					OrderId: 3,
					Order: &Order_BidOrder{
						BidOrder: &BidOrder{
							MarketId: 11,
							Assets:   sdk.NewInt64Coin("adolla", 93),
							Price:    sdk.NewInt64Coin("pdolla", 16),
						},
					},
				},
				Splits:            nil,
				AssetsFilledAmt:   sdkmath.ZeroInt(),
				AssetsUnfilledAmt: sdkmath.NewInt(93),
				PriceAppliedAmt:   sdkmath.ZeroInt(),
				PriceLeftAmt:      sdkmath.NewInt(16),
				IsFinalized:       false,
				FeesToPay:         nil,
				OrderFeesLeft:     nil,
				PriceFilledAmt:    sdkmath.ZeroInt(),
				PriceUnfilledAmt:  sdkmath.ZeroInt(),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual *OrderFulfillment
			defer func() {
				if t.Failed() {
					t.Logf("  Actual: %s", orderFulfillmentString(actual))
					t.Logf("Expected: %s", orderFulfillmentString(tc.expected))
				}
			}()

			testFunc := func() {
				actual = NewOrderFulfillment(tc.order)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "NewOrderFulfillment")
			assert.Equal(t, tc.expected, actual, "NewOrderFulfillment result")
		})
	}
}

// TODO[1658]: TestNewOrderFulfillments(t *testing.T)

func TestOrderFulfillment_GetAssetsFilled(t *testing.T) {
	coin := func(amt int64) sdk.Coin {
		return sdk.Coin{Denom: "asset", Amount: sdkmath.NewInt(amt)}
	}
	askOrder := NewOrder(444).WithAsk(&AskOrder{Assets: coin(5555)})
	bidOrder := NewOrder(666).WithBid(&BidOrder{Assets: coin(7777)})

	newOF := func(order *Order, amt int64) OrderFulfillment {
		return OrderFulfillment{
			Order:           order,
			AssetsFilledAmt: sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name string
		f    OrderFulfillment
		exp  sdk.Coin
	}{
		{name: "positive ask", f: newOF(askOrder, 2), exp: coin(2)},
		{name: "zero ask", f: newOF(askOrder, 0), exp: coin(0)},
		{name: "negative ask", f: newOF(askOrder, -3), exp: coin(-3)},
		{name: "positive bid", f: newOF(bidOrder, 2), exp: coin(2)},
		{name: "zero bid", f: newOF(bidOrder, 0), exp: coin(0)},
		{name: "negative bid", f: newOF(bidOrder, -3), exp: coin(-3)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coin
			testFunc := func() {
				actual = tc.f.GetAssetsFilled()
			}
			require.NotPanics(t, testFunc, "GetAssetsFilled()")
			assert.Equal(t, tc.exp.String(), actual.String(), "GetAssetsFilled() result")
		})
	}
}

func TestOrderFulfillment_GetAssetsUnfilled(t *testing.T) {
	coin := func(amt int64) sdk.Coin {
		return sdk.Coin{Denom: "asset", Amount: sdkmath.NewInt(amt)}
	}
	askOrder := NewOrder(444).WithAsk(&AskOrder{Assets: coin(5555)})
	bidOrder := NewOrder(666).WithBid(&BidOrder{Assets: coin(7777)})

	newOF := func(order *Order, amt int64) OrderFulfillment {
		return OrderFulfillment{
			Order:             order,
			AssetsUnfilledAmt: sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name string
		f    OrderFulfillment
		exp  sdk.Coin
	}{
		{name: "positive ask", f: newOF(askOrder, 2), exp: coin(2)},
		{name: "zero ask", f: newOF(askOrder, 0), exp: coin(0)},
		{name: "negative ask", f: newOF(askOrder, -3), exp: coin(-3)},
		{name: "positive bid", f: newOF(bidOrder, 2), exp: coin(2)},
		{name: "zero bid", f: newOF(bidOrder, 0), exp: coin(0)},
		{name: "negative bid", f: newOF(bidOrder, -3), exp: coin(-3)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coin
			testFunc := func() {
				actual = tc.f.GetAssetsUnfilled()
			}
			require.NotPanics(t, testFunc, "GetAssetsUnfilled()")
			assert.Equal(t, tc.exp.String(), actual.String(), "GetAssetsUnfilled() result")
		})
	}
}

func TestOrderFulfillment_GetPriceApplied(t *testing.T) {
	coin := func(amt int64) sdk.Coin {
		return sdk.Coin{Denom: "price", Amount: sdkmath.NewInt(amt)}
	}
	askOrder := NewOrder(444).WithAsk(&AskOrder{Price: coin(5555)})
	bidOrder := NewOrder(666).WithBid(&BidOrder{Price: coin(7777)})

	newOF := func(order *Order, amt int64) OrderFulfillment {
		return OrderFulfillment{
			Order:           order,
			PriceAppliedAmt: sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name string
		f    OrderFulfillment
		exp  sdk.Coin
	}{
		{name: "positive ask", f: newOF(askOrder, 2), exp: coin(2)},
		{name: "zero ask", f: newOF(askOrder, 0), exp: coin(0)},
		{name: "negative ask", f: newOF(askOrder, -3), exp: coin(-3)},
		{name: "positive bid", f: newOF(bidOrder, 2), exp: coin(2)},
		{name: "zero bid", f: newOF(bidOrder, 0), exp: coin(0)},
		{name: "negative bid", f: newOF(bidOrder, -3), exp: coin(-3)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coin
			testFunc := func() {
				actual = tc.f.GetPriceApplied()
			}
			require.NotPanics(t, testFunc, "GetPriceApplied()")
			assert.Equal(t, tc.exp.String(), actual.String(), "GetPriceApplied() result")
		})
	}
}

func TestOrderFulfillment_GetPriceLeft(t *testing.T) {
	coin := func(amt int64) sdk.Coin {
		return sdk.Coin{Denom: "price", Amount: sdkmath.NewInt(amt)}
	}
	askOrder := NewOrder(444).WithAsk(&AskOrder{Price: coin(5555)})
	bidOrder := NewOrder(666).WithBid(&BidOrder{Price: coin(7777)})

	newOF := func(order *Order, amt int64) OrderFulfillment {
		return OrderFulfillment{
			Order:        order,
			PriceLeftAmt: sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name string
		f    OrderFulfillment
		exp  sdk.Coin
	}{
		{name: "positive ask", f: newOF(askOrder, 2), exp: coin(2)},
		{name: "zero ask", f: newOF(askOrder, 0), exp: coin(0)},
		{name: "negative ask", f: newOF(askOrder, -3), exp: coin(-3)},
		{name: "positive bid", f: newOF(bidOrder, 2), exp: coin(2)},
		{name: "zero bid", f: newOF(bidOrder, 0), exp: coin(0)},
		{name: "negative bid", f: newOF(bidOrder, -3), exp: coin(-3)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coin
			testFunc := func() {
				actual = tc.f.GetPriceLeft()
			}
			require.NotPanics(t, testFunc, "GetPriceLeft()")
			assert.Equal(t, tc.exp.String(), actual.String(), "GetPriceLeft() result")
		})
	}
}

func TestOrderFulfillment_GetPriceFilled(t *testing.T) {
	coin := func(amt int64) sdk.Coin {
		return sdk.Coin{Denom: "price", Amount: sdkmath.NewInt(amt)}
	}
	askOrder := NewOrder(444).WithAsk(&AskOrder{Price: coin(5555)})
	bidOrder := NewOrder(666).WithBid(&BidOrder{Price: coin(7777)})

	newOF := func(order *Order, amt int64) OrderFulfillment {
		return OrderFulfillment{
			Order:          order,
			PriceFilledAmt: sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name string
		f    OrderFulfillment
		exp  sdk.Coin
	}{
		{name: "positive ask", f: newOF(askOrder, 2), exp: coin(2)},
		{name: "zero ask", f: newOF(askOrder, 0), exp: coin(0)},
		{name: "negative ask", f: newOF(askOrder, -3), exp: coin(-3)},
		{name: "positive bid", f: newOF(bidOrder, 2), exp: coin(2)},
		{name: "zero bid", f: newOF(bidOrder, 0), exp: coin(0)},
		{name: "negative bid", f: newOF(bidOrder, -3), exp: coin(-3)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coin
			testFunc := func() {
				actual = tc.f.GetPriceFilled()
			}
			require.NotPanics(t, testFunc, "GetPriceFilled()")
			assert.Equal(t, tc.exp.String(), actual.String(), "GetPriceFilled() result")
		})
	}
}

func TestOrderFulfillment_GetPriceUnfilled(t *testing.T) {
	coin := func(amt int64) sdk.Coin {
		return sdk.Coin{Denom: "price", Amount: sdkmath.NewInt(amt)}
	}
	askOrder := NewOrder(444).WithAsk(&AskOrder{Price: coin(5555)})
	bidOrder := NewOrder(666).WithBid(&BidOrder{Price: coin(7777)})

	newOF := func(order *Order, amt int64) OrderFulfillment {
		return OrderFulfillment{
			Order:            order,
			PriceUnfilledAmt: sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name string
		f    OrderFulfillment
		exp  sdk.Coin
	}{
		{name: "positive ask", f: newOF(askOrder, 2), exp: coin(2)},
		{name: "zero ask", f: newOF(askOrder, 0), exp: coin(0)},
		{name: "negative ask", f: newOF(askOrder, -3), exp: coin(-3)},
		{name: "positive bid", f: newOF(bidOrder, 2), exp: coin(2)},
		{name: "zero bid", f: newOF(bidOrder, 0), exp: coin(0)},
		{name: "negative bid", f: newOF(bidOrder, -3), exp: coin(-3)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coin
			testFunc := func() {
				actual = tc.f.GetPriceUnfilled()
			}
			require.NotPanics(t, testFunc, "GetPriceUnfilled()")
			assert.Equal(t, tc.exp.String(), actual.String(), "GetPriceUnfilled() result")
		})
	}
}

func TestOrderFulfillment_IsFullyFilled(t *testing.T) {
	newOF := func(amt int64) OrderFulfillment {
		return OrderFulfillment{
			AssetsUnfilledAmt: sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name string
		f    OrderFulfillment
		exp  bool
	}{
		{name: "positive assets unfilled", f: newOF(2), exp: false},
		{name: "zero assets unfilled", f: newOF(0), exp: true},
		{name: "negative assets unfilled", f: newOF(-3), exp: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = tc.f.IsFullyFilled()
			}
			require.NotPanics(t, testFunc, "IsFullyFilled()")
			assert.Equal(t, tc.exp, actual, "IsFullyFilled() result")
		})
	}
}

func TestOrderFulfillment_IsCompletelyUnfulfilled(t *testing.T) {
	newOF := func(amt int64) OrderFulfillment {
		return OrderFulfillment{
			AssetsFilledAmt: sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name string
		f    OrderFulfillment
		exp  bool
	}{
		{name: "positive assets filled", f: newOF(2), exp: false},
		{name: "zero assets filled", f: newOF(0), exp: true},
		{name: "negative assets filled", f: newOF(-3), exp: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = tc.f.IsCompletelyUnfulfilled()
			}
			require.NotPanics(t, testFunc, "IsCompletelyUnfulfilled()")
			assert.Equal(t, tc.exp, actual, "IsCompletelyUnfulfilled() result")
		})
	}
}

func TestOrderFulfillment_GetOrderID(t *testing.T) {
	newOF := func(orderID uint64) OrderFulfillment {
		return OrderFulfillment{
			Order: NewOrder(orderID),
		}
	}

	tests := []struct {
		name string
		f    OrderFulfillment
		exp  uint64
	}{
		{name: "zero", f: newOF(0), exp: 0},
		{name: "one", f: newOF(1), exp: 1},
		{name: "five", f: newOF(5), exp: 5},
		{name: "max uint32+1", f: newOF(4_294_967_296), exp: 4_294_967_296},
		{name: "max uint64", f: newOF(18_446_744_073_709_551_615), exp: 18_446_744_073_709_551_615},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual uint64
			testFunc := func() {
				actual = tc.f.GetOrderID()
			}
			require.NotPanics(t, testFunc, "GetOrderID()")
			assert.Equal(t, tc.exp, actual, "GetOrderID() result")
		})
	}
}

func TestOrderFulfillment_IsAskOrder(t *testing.T) {
	tests := []struct {
		name string
		f    OrderFulfillment
		exp  bool
	}{
		{name: "ask", f: OrderFulfillment{Order: NewOrder(444).WithAsk(&AskOrder{})}, exp: true},
		{name: "bid", f: OrderFulfillment{Order: NewOrder(666).WithBid(&BidOrder{})}, exp: false},
		{name: "nil", f: OrderFulfillment{Order: NewOrder(888)}, exp: false},
		{name: "unknown", f: OrderFulfillment{Order: newUnknownOrder(7)}, exp: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = tc.f.IsAskOrder()
			}
			require.NotPanics(t, testFunc, "IsAskOrder()")
			assert.Equal(t, tc.exp, actual, "IsAskOrder() result")
		})
	}
}

func TestOrderFulfillment_IsBidOrder(t *testing.T) {
	tests := []struct {
		name string
		f    OrderFulfillment
		exp  bool
	}{
		{name: "ask", f: OrderFulfillment{Order: NewOrder(444).WithAsk(&AskOrder{})}, exp: false},
		{name: "bid", f: OrderFulfillment{Order: NewOrder(666).WithBid(&BidOrder{})}, exp: true},
		{name: "nil", f: OrderFulfillment{Order: NewOrder(888)}, exp: false},
		{name: "unknown", f: OrderFulfillment{Order: newUnknownOrder(9)}, exp: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = tc.f.IsBidOrder()
			}
			require.NotPanics(t, testFunc, "IsBidOrder()")
			assert.Equal(t, tc.exp, actual, "IsBidOrder() result")
		})
	}
}

func TestOrderFulfillment_GetMarketID(t *testing.T) {
	askOrder := func(marketID uint32) OrderFulfillment {
		return OrderFulfillment{
			Order: NewOrder(999).WithAsk(&AskOrder{MarketId: marketID}),
		}
	}
	bidOrder := func(marketID uint32) OrderFulfillment {
		return OrderFulfillment{
			Order: NewOrder(999).WithBid(&BidOrder{MarketId: marketID}),
		}
	}

	tests := []struct {
		name string
		f    OrderFulfillment
		exp  uint32
	}{
		{name: "ask zero", f: askOrder(0), exp: 0},
		{name: "ask one", f: askOrder(1), exp: 1},
		{name: "ask five", f: askOrder(5), exp: 5},
		{name: "ask max uint16+1", f: askOrder(65_536), exp: 65_536},
		{name: "ask max uint32", f: askOrder(4_294_967_295), exp: 4_294_967_295},
		{name: "bid zero", f: bidOrder(0), exp: 0},
		{name: "bid one", f: bidOrder(1), exp: 1},
		{name: "bid five", f: bidOrder(5), exp: 5},
		{name: "bid max uint16+1", f: bidOrder(65_536), exp: 65_536},
		{name: "bid max uint32", f: bidOrder(4_294_967_295), exp: 4_294_967_295},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual uint32
			testFunc := func() {
				actual = tc.f.GetMarketID()
			}
			require.NotPanics(t, testFunc, "GetMarketID()")
			assert.Equal(t, tc.exp, actual, "GetMarketID() result")
		})
	}
}

func TestOrderFulfillment_GetOwner(t *testing.T) {
	askOrder := func(seller string) OrderFulfillment {
		return OrderFulfillment{
			Order: NewOrder(999).WithAsk(&AskOrder{Seller: seller}),
		}
	}
	bidOrder := func(buyer string) OrderFulfillment {
		return OrderFulfillment{
			Order: NewOrder(999).WithBid(&BidOrder{Buyer: buyer}),
		}
	}
	owner := sdk.AccAddress("owner_______________").String()

	tests := []struct {
		name string
		f    OrderFulfillment
		exp  string
	}{
		{name: "ask empty", f: askOrder(""), exp: ""},
		{name: "ask not a bech32", f: askOrder("owner"), exp: "owner"},
		{name: "ask beche32", f: askOrder(owner), exp: owner},
		{name: "bid empty", f: bidOrder(""), exp: ""},
		{name: "bid not a bech32", f: bidOrder("owner"), exp: "owner"},
		{name: "bid beche32", f: bidOrder(owner), exp: owner},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = tc.f.GetOwner()
			}
			require.NotPanics(t, testFunc, "GetOwner()")
			assert.Equal(t, tc.exp, actual, "GetOwner() result")
		})
	}
}

func TestOrderFulfillment_GetAssets(t *testing.T) {
	coin := func(amt int64) sdk.Coin {
		return sdk.Coin{Denom: "asset", Amount: sdkmath.NewInt(amt)}
	}
	askOrder := func(amt int64) OrderFulfillment {
		return OrderFulfillment{
			Order: NewOrder(999).WithAsk(&AskOrder{Assets: coin(amt)}),
		}
	}
	bidOrder := func(amt int64) OrderFulfillment {
		return OrderFulfillment{
			Order: NewOrder(999).WithBid(&BidOrder{Assets: coin(amt)}),
		}
	}

	tests := []struct {
		name string
		f    OrderFulfillment
		exp  sdk.Coin
	}{
		{name: "ask positive", f: askOrder(123), exp: coin(123)},
		{name: "ask zero", f: askOrder(0), exp: coin(0)},
		{name: "ask negative", f: askOrder(-9), exp: coin(-9)},
		{name: "bid positive", f: bidOrder(345), exp: coin(345)},
		{name: "bid zero", f: bidOrder(0), exp: coin(0)},
		{name: "bid negative", f: bidOrder(-8), exp: coin(-8)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coin
			testFunc := func() {
				actual = tc.f.GetAssets()
			}
			require.NotPanics(t, testFunc, "GetAssets()")
			assert.Equal(t, tc.exp.String(), actual.String(), "GetAssets() result")
		})
	}
}

func TestOrderFulfillment_GetPrice(t *testing.T) {
	coin := func(amt int64) sdk.Coin {
		return sdk.Coin{Denom: "price", Amount: sdkmath.NewInt(amt)}
	}
	askOrder := func(amt int64) OrderFulfillment {
		return OrderFulfillment{
			Order: NewOrder(999).WithAsk(&AskOrder{Price: coin(amt)}),
		}
	}
	bidOrder := func(amt int64) OrderFulfillment {
		return OrderFulfillment{
			Order: NewOrder(999).WithBid(&BidOrder{Price: coin(amt)}),
		}
	}

	tests := []struct {
		name string
		f    OrderFulfillment
		exp  sdk.Coin
	}{
		{name: "ask positive", f: askOrder(123), exp: coin(123)},
		{name: "ask zero", f: askOrder(0), exp: coin(0)},
		{name: "ask negative", f: askOrder(-9), exp: coin(-9)},
		{name: "bid positive", f: bidOrder(345), exp: coin(345)},
		{name: "bid zero", f: bidOrder(0), exp: coin(0)},
		{name: "bid negative", f: bidOrder(-8), exp: coin(-8)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coin
			testFunc := func() {
				actual = tc.f.GetPrice()
			}
			require.NotPanics(t, testFunc, "GetPrice()")
			assert.Equal(t, tc.exp.String(), actual.String(), "GetPrice() result")
		})
	}
}

func TestOrderFulfillment_GetSettlementFees(t *testing.T) {
	coin := func(amt int64) *sdk.Coin {
		return &sdk.Coin{Denom: "fees", Amount: sdkmath.NewInt(amt)}
	}
	askOrder := func(coin *sdk.Coin) OrderFulfillment {
		return OrderFulfillment{
			Order: NewOrder(999).WithAsk(&AskOrder{SellerSettlementFlatFee: coin}),
		}
	}
	bidOrder := func(coins sdk.Coins) OrderFulfillment {
		return OrderFulfillment{
			Order: NewOrder(999).WithBid(&BidOrder{BuyerSettlementFees: coins}),
		}
	}

	tests := []struct {
		name string
		f    OrderFulfillment
		exp  sdk.Coins
	}{
		{name: "ask nil", f: askOrder(nil), exp: nil},
		{name: "ask zero", f: askOrder(coin(0)), exp: sdk.Coins{*coin(0)}},
		{name: "ask positive", f: askOrder(coin(3)), exp: sdk.Coins{*coin(3)}},
		{name: "bid nil", f: bidOrder(nil), exp: nil},
		{name: "bid empty", f: bidOrder(sdk.Coins{}), exp: sdk.Coins{}},
		{name: "bid positive", f: bidOrder(sdk.Coins{*coin(3)}), exp: sdk.Coins{*coin(3)}},
		{name: "bid zero", f: bidOrder(sdk.Coins{*coin(0)}), exp: sdk.Coins{*coin(0)}},
		{name: "bid negative", f: bidOrder(sdk.Coins{*coin(-2)}), exp: sdk.Coins{*coin(-2)}},
		{
			name: "bid multiple",
			f:    bidOrder(sdk.Coins{*coin(987), sdk.NewInt64Coin("six", 6), sdk.Coin{Denom: "zeg", Amount: sdkmath.NewInt(-1)}}),
			exp:  sdk.Coins{*coin(987), sdk.NewInt64Coin("six", 6), sdk.Coin{Denom: "zeg", Amount: sdkmath.NewInt(-1)}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coins
			testFunc := func() {
				actual = tc.f.GetSettlementFees()
			}
			require.NotPanics(t, testFunc, "GetSettlementFees()")
			assert.Equal(t, tc.exp.String(), actual.String(), "GetSettlementFees() result")
		})
	}
}

func TestOrderFulfillment_PartialFillAllowed(t *testing.T) {
	askOrder := func(allowPartial bool) OrderFulfillment {
		return OrderFulfillment{
			Order: NewOrder(999).WithAsk(&AskOrder{AllowPartial: allowPartial}),
		}
	}
	bidOrder := func(allowPartial bool) OrderFulfillment {
		return OrderFulfillment{
			Order: NewOrder(999).WithBid(&BidOrder{AllowPartial: allowPartial}),
		}
	}

	tests := []struct {
		name string
		f    OrderFulfillment
		exp  bool
	}{
		{name: "ask true", f: askOrder(true), exp: true},
		{name: "ask false", f: askOrder(false), exp: false},
		{name: "bid true", f: bidOrder(true), exp: true},
		{name: "bid false", f: bidOrder(false), exp: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = tc.f.PartialFillAllowed()
			}
			require.NotPanics(t, testFunc, "PartialFillAllowed()")
			assert.Equal(t, tc.exp, actual, "PartialFillAllowed() result")
		})
	}
}

func TestOrderFulfillment_GetOrderType(t *testing.T) {
	tests := []struct {
		name string
		f    OrderFulfillment
		exp  string
	}{
		{name: "ask", f: OrderFulfillment{Order: NewOrder(444).WithAsk(&AskOrder{})}, exp: OrderTypeAsk},
		{name: "bid", f: OrderFulfillment{Order: NewOrder(666).WithBid(&BidOrder{})}, exp: OrderTypeBid},
		{name: "nil", f: OrderFulfillment{Order: NewOrder(888)}, exp: "<nil>"},
		{name: "unknown", f: OrderFulfillment{Order: newUnknownOrder(8)}, exp: "*exchange.unknownOrderType"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = tc.f.GetOrderType()
			}
			require.NotPanics(t, testFunc, "GetOrderType()")
			assert.Equal(t, tc.exp, actual, "GetOrderType() result")
		})
	}
}

func TestOrderFulfillment_GetOrderTypeByte(t *testing.T) {
	tests := []struct {
		name string
		f    OrderFulfillment
		exp  byte
	}{
		{name: "ask", f: OrderFulfillment{Order: NewOrder(444).WithAsk(&AskOrder{})}, exp: OrderTypeByteAsk},
		{name: "bid", f: OrderFulfillment{Order: NewOrder(666).WithBid(&BidOrder{})}, exp: OrderTypeByteBid},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual byte
			testFunc := func() {
				actual = tc.f.GetOrderTypeByte()
			}
			require.NotPanics(t, testFunc, "GetOrderTypeByte()")
			assert.Equal(t, tc.exp, actual, "GetOrderTypeByte() result")
		})
	}
}

func TestOrderFulfillment_GetHoldAmount(t *testing.T) {
	tests := []struct {
		name string
		f    OrderFulfillment
	}{
		{
			name: "ask",
			f: OrderFulfillment{
				Order: NewOrder(111).WithAsk(&AskOrder{
					Assets:                  sdk.Coin{Denom: "asset", Amount: sdkmath.NewInt(55)},
					SellerSettlementFlatFee: &sdk.Coin{Denom: "fee", Amount: sdkmath.NewInt(3)},
				}),
			},
		},
		{
			name: "bid",
			f: OrderFulfillment{
				Order: NewOrder(111).WithBid(&BidOrder{
					Price: sdk.Coin{Denom: "price", Amount: sdkmath.NewInt(55)},
					BuyerSettlementFees: sdk.Coins{
						{Denom: "feea", Amount: sdkmath.NewInt(3)},
						{Denom: "feeb", Amount: sdkmath.NewInt(4)},
					},
				}),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			expected := tc.f.GetHoldAmount()
			var actual sdk.Coins
			testFunc := func() {
				actual = tc.f.GetHoldAmount()
			}
			require.NotPanics(t, testFunc, "GetHoldAmount()")
			assert.Equal(t, expected, actual, "GetHoldAmount() result")
		})
	}
}

// TODO[1658]: func TestOrderFulfillment_DistributeAssets(t *testing.T)

// TODO[1658]: func TestDistributeAssets(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_DistributePrice(t *testing.T)

// TODO[1658]: func TestDistributePrice(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_SplitOrder(t *testing.T)

// TODO[1658]: func TestSumAssetsAndPrice(t *testing.T)

// TODO[1658]: func TestSumPriceLeft(t *testing.T)

// TODO[1658]: func TestBuildSettlement(t *testing.T)

// TODO[1658]: func TestValidateCanSettle(t *testing.T)

// TODO[1658]: func TestAllocateAssets(t *testing.T)

// TODO[1658]: func TestSplitPartial(t *testing.T)

// TODO[1658]: func TestAllocatePrice(t *testing.T)

// TODO[1658]: func TestSetFeesToPay(t *testing.T)

// TODO[1658]: func TestValidateFulfillments(t *testing.T)

// assertEqualOrderFulfillments asserts that the two order fulfillments are equal.
// Returns true if equal.
// If not equal, and neither are nil, equality on each field is also asserted in order to help identify the problem.
func assertEqualOrderFulfillments(t *testing.T, expected, actual *OrderFulfillment, message string, args ...interface{}) bool {
	if assert.Equalf(t, expected, actual, message, args...) {
		return true
	}
	// If either is nil, that's easy to understand in the above failure, so there's nothing more to do.
	if expected == nil || actual == nil {
		return false
	}

	msg := func(val string) string {
		if len(message) == 0 {
			return val
		}
		return val + "\n" + message
	}

	// Assert equality on each individual field so that we can more easily find the problem.
	// If any of the Ints fail with a complaint about Int.abs = (big.nat) <nil> vs {}, use ZeroAmtAfterSub for the expected.
	assert.Equalf(t, expected.Order, actual.Order, msg("OrderFulfillment.Order"), args...)
	assert.Equalf(t, expected.Splits, actual.Splits, msg("OrderFulfillment.Splits"), args...)
	assert.Equalf(t, expected.AssetsFilledAmt, actual.AssetsFilledAmt, msg("OrderFulfillment.AssetsFilledAmt"), args...)
	assert.Equalf(t, expected.AssetsUnfilledAmt, actual.AssetsUnfilledAmt, msg("OrderFulfillment.AssetsUnfilledAmt"), args...)
	assert.Equalf(t, expected.PriceAppliedAmt, actual.PriceAppliedAmt, msg("OrderFulfillment.PriceAppliedAmt"), args...)
	assert.Equalf(t, expected.PriceLeftAmt, actual.PriceLeftAmt, msg("OrderFulfillment.PriceLeftAmt"), args...)
	assert.Equalf(t, expected.IsFinalized, actual.IsFinalized, msg("OrderFulfillment.IsFinalized"), args...)
	assert.Equalf(t, expected.FeesToPay, actual.FeesToPay, msg("OrderFulfillment.FeesToPay"), args...)
	assert.Equalf(t, expected.OrderFeesLeft, actual.OrderFeesLeft, msg("OrderFulfillment.OrderFeesLeft"), args...)
	assert.Equalf(t, expected.PriceFilledAmt, actual.PriceFilledAmt, msg("OrderFulfillment.PriceFilledAmt"), args...)
	assert.Equalf(t, expected.PriceUnfilledAmt, actual.PriceUnfilledAmt, msg("OrderFulfillment.PriceUnfilledAmt"), args...)
	t.Logf("  Actual: %s", orderFulfillmentString(actual))
	t.Logf("Expected: %s", orderFulfillmentString(expected))
	return false
}

func TestOrderFulfillment_Apply(t *testing.T) {
	assetCoin := func(amt int64) sdk.Coin {
		return sdk.Coin{Denom: "acoin", Amount: sdkmath.NewInt(amt)}
	}
	priceCoin := func(amt int64) sdk.Coin {
		return sdk.Coin{Denom: "pcoin", Amount: sdkmath.NewInt(amt)}
	}
	askOrder := func(orderID uint64, assetsAmt, priceAmt int64) *Order {
		return NewOrder(orderID).WithAsk(&AskOrder{
			MarketId: 86420,
			Seller:   "seller",
			Assets:   assetCoin(assetsAmt),
			Price:    priceCoin(priceAmt),
		})
	}
	bidOrder := func(orderID uint64, assetsAmt, priceAmt int64) *Order {
		return NewOrder(orderID).WithBid(&BidOrder{
			MarketId: 86420,
			Buyer:    "buyer",
			Assets:   assetCoin(assetsAmt),
			Price:    priceCoin(priceAmt),
		})
	}

	tests := []struct {
		name      string
		receiver  *OrderFulfillment
		order     *OrderFulfillment
		assetsAmt sdkmath.Int
		priceAmt  sdkmath.Int
		expErr    string
		expResult *OrderFulfillment
	}{
		{
			name: "fills order in full",
			receiver: &OrderFulfillment{
				Order:             askOrder(1, 20, 55),
				Splits:            nil,
				AssetsFilledAmt:   sdkmath.ZeroInt(),
				AssetsUnfilledAmt: sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.ZeroInt(),
				PriceLeftAmt:      sdkmath.NewInt(55),
			},
			order:     &OrderFulfillment{Order: bidOrder(2, 40, 110)},
			assetsAmt: sdkmath.NewInt(20),
			priceAmt:  sdkmath.NewInt(55),
			expResult: &OrderFulfillment{
				Order: askOrder(1, 20, 55),
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: bidOrder(2, 40, 110)},
						Assets: assetCoin(20),
						Price:  priceCoin(55),
					},
				},
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: ZeroAmtAfterSub,
				PriceAppliedAmt:   sdkmath.NewInt(55),
				PriceLeftAmt:      ZeroAmtAfterSub,
			},
		},
		{
			name: "partially fills order",
			receiver: &OrderFulfillment{
				Order:             askOrder(1, 20, 55),
				Splits:            nil,
				AssetsFilledAmt:   sdkmath.ZeroInt(),
				AssetsUnfilledAmt: sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.ZeroInt(),
				PriceLeftAmt:      sdkmath.NewInt(55),
			},
			order:     &OrderFulfillment{Order: bidOrder(2, 40, 110)},
			assetsAmt: sdkmath.NewInt(11),
			priceAmt:  sdkmath.NewInt(22),
			expResult: &OrderFulfillment{
				Order: askOrder(1, 20, 55),
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: bidOrder(2, 40, 110)},
						Assets: assetCoin(11),
						Price:  priceCoin(22),
					},
				},
				AssetsFilledAmt:   sdkmath.NewInt(11),
				AssetsUnfilledAmt: sdkmath.NewInt(9),
				PriceAppliedAmt:   sdkmath.NewInt(22),
				PriceLeftAmt:      sdkmath.NewInt(33),
			},
		},
		{
			name: "already partially filled, fills rest",
			receiver: &OrderFulfillment{
				Order: askOrder(1, 20, 55),
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: bidOrder(3, 60, 220)},
						Assets: assetCoin(9),
						Price:  priceCoin(33),
					},
				},
				AssetsFilledAmt:   sdkmath.NewInt(9),
				AssetsUnfilledAmt: sdkmath.NewInt(11),
				PriceAppliedAmt:   sdkmath.NewInt(33),
				PriceLeftAmt:      sdkmath.NewInt(22),
			},
			order:     &OrderFulfillment{Order: bidOrder(2, 40, 110)},
			assetsAmt: sdkmath.NewInt(11),
			priceAmt:  sdkmath.NewInt(22),
			expResult: &OrderFulfillment{
				Order: askOrder(1, 20, 55),
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: bidOrder(3, 60, 220)},
						Assets: assetCoin(9),
						Price:  priceCoin(33),
					},
					{
						Order:  &OrderFulfillment{Order: bidOrder(2, 40, 110)},
						Assets: assetCoin(11),
						Price:  priceCoin(22),
					},
				},
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: ZeroAmtAfterSub,
				PriceAppliedAmt:   sdkmath.NewInt(55),
				PriceLeftAmt:      ZeroAmtAfterSub,
			},
		},
		{
			name: "ask assets overfill",
			receiver: &OrderFulfillment{
				Order:             askOrder(1, 20, 55),
				Splits:            nil,
				AssetsFilledAmt:   sdkmath.ZeroInt(),
				AssetsUnfilledAmt: sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.ZeroInt(),
				PriceLeftAmt:      sdkmath.NewInt(55),
			},
			order:     &OrderFulfillment{Order: bidOrder(2, 40, 110)},
			assetsAmt: sdkmath.NewInt(21),
			priceAmt:  sdkmath.NewInt(55),
			expErr:    "cannot fill ask order 1 having assets left \"20acoin\" with \"21acoin\" from bid order 2: overfill",
		},
		{
			name: "bid assets overfill",
			receiver: &OrderFulfillment{
				Order:             bidOrder(1, 20, 55),
				Splits:            nil,
				AssetsFilledAmt:   sdkmath.ZeroInt(),
				AssetsUnfilledAmt: sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.ZeroInt(),
				PriceLeftAmt:      sdkmath.NewInt(55),
			},
			order:     &OrderFulfillment{Order: askOrder(2, 40, 110)},
			assetsAmt: sdkmath.NewInt(21),
			priceAmt:  sdkmath.NewInt(55),
			expErr:    "cannot fill bid order 1 having assets left \"20acoin\" with \"21acoin\" from ask order 2: overfill",
		},
		{
			name: "ask price overfill",
			receiver: &OrderFulfillment{
				Order:             askOrder(1, 20, 55),
				Splits:            nil,
				AssetsFilledAmt:   sdkmath.ZeroInt(),
				AssetsUnfilledAmt: sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.ZeroInt(),
				PriceLeftAmt:      sdkmath.NewInt(55),
			},
			order:     &OrderFulfillment{Order: bidOrder(2, 40, 110)},
			assetsAmt: sdkmath.NewInt(20),
			priceAmt:  sdkmath.NewInt(56),
			expResult: &OrderFulfillment{
				Order: askOrder(1, 20, 55),
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: bidOrder(2, 40, 110)},
						Assets: assetCoin(20),
						Price:  priceCoin(56),
					},
				},
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: ZeroAmtAfterSub,
				PriceAppliedAmt:   sdkmath.NewInt(56),
				PriceLeftAmt:      sdkmath.NewInt(-1),
			},
		},
		{
			name: "bid price overfill",
			receiver: &OrderFulfillment{
				Order:             bidOrder(1, 20, 55),
				Splits:            nil,
				AssetsFilledAmt:   sdkmath.ZeroInt(),
				AssetsUnfilledAmt: sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.ZeroInt(),
				PriceLeftAmt:      sdkmath.NewInt(55),
			},
			order:     &OrderFulfillment{Order: askOrder(2, 40, 110)},
			assetsAmt: sdkmath.NewInt(20),
			priceAmt:  sdkmath.NewInt(56),
			expErr:    "cannot fill bid order 1 having price left \"55pcoin\" to ask order 2 at a price of \"56pcoin\": overfill",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := copyOrderFulfillment(tc.receiver)
			if tc.expResult == nil {
				tc.expResult = copyOrderFulfillment(tc.receiver)
			}

			var err error
			testFunc := func() {
				err = tc.receiver.Apply(tc.order, tc.assetsAmt, tc.priceAmt)
			}
			require.NotPanics(t, testFunc, "Apply")
			assertions.AssertErrorValue(t, err, tc.expErr, "Apply error")
			if !assertEqualOrderFulfillments(t, tc.expResult, tc.receiver, "order fulfillment after .Apply") {
				t.Logf("Original: %s", orderFulfillmentString(orig))
			}
		})
	}
}

func TestOrderFulfillment_ApplyLeftoverPrice(t *testing.T) {
	type testCase struct {
		name           string
		receiver       *OrderFulfillment
		askSplit       *OrderSplit
		amt            sdkmath.Int
		expFulfillment *OrderFulfillment
		expAskSplit    *OrderSplit
		expPanic       string
	}

	newTestCase := func(name string, bidSplitIndexes ...int) testCase {
		// Picture a bid order with 150 assets at a cost of 5555 being split among 3 ask orders evenly (50 each).
		// Each ask order has 53 to sell: 50 are coming from this bid order, and 1 and 2 each from two other bids.
		// During initial splitting, the bid will pay each ask 5555 * 50 / 150 = 1851.
		// 1851 * 3 = 5553, so there's 2 leftover.
		// The other 3 are being bought for 30 each (90 total).

		bidOrderID := uint64(200)
		bidOrder := NewOrder(bidOrderID).WithBid(&BidOrder{
			Price: sdk.NewInt64Coin("pcoin", 5555),
		})

		tc := testCase{
			name: name,
			receiver: &OrderFulfillment{
				Order:           bidOrder,
				PriceAppliedAmt: sdkmath.NewInt(5553),
				PriceLeftAmt:    sdkmath.NewInt(2),
			},
			amt: sdkmath.NewInt(2),
			expFulfillment: &OrderFulfillment{
				Order:           bidOrder,
				PriceAppliedAmt: sdkmath.NewInt(5555),
				PriceLeftAmt:    ZeroAmtAfterSub,
			},
			askSplit: &OrderSplit{
				Order: &OrderFulfillment{
					Order: NewOrder(1).WithAsk(&AskOrder{
						Price: sdk.NewInt64Coin("pcoin", 1500),
					}),
					PriceAppliedAmt: sdkmath.NewInt(1941), // 5555 * 50 / 150 = 1851 from main bid + 90 from the others.
					PriceLeftAmt:    sdkmath.NewInt(-441), // = 1500 - 1941
				},
				Price: sdk.NewInt64Coin("pcoin", 1851),
			},
			expAskSplit: &OrderSplit{
				Order: &OrderFulfillment{
					Order: NewOrder(1).WithAsk(&AskOrder{
						Price: sdk.NewInt64Coin("pcoin", 1500),
					}),
					PriceAppliedAmt: sdkmath.NewInt(1943),
					PriceLeftAmt:    sdkmath.NewInt(-443),
				},
				Price: sdk.NewInt64Coin("pcoin", 1853),
			},
		}

		bidSplits := []*OrderSplit{
			{
				// This is the primary bid split that we'll be looking to update.
				Order: &OrderFulfillment{Order: NewOrder(bidOrderID).WithBid(&BidOrder{})},
				Price: sdk.NewInt64Coin("pcoin", 1851), // == 5555 / 3 (truncated)
			},
			{
				Order: &OrderFulfillment{Order: NewOrder(bidOrderID + 1).WithBid(&BidOrder{})},
				Price: sdk.NewInt64Coin("pcoin", 30),
			},
			{
				Order: &OrderFulfillment{Order: NewOrder(bidOrderID + 2).WithBid(&BidOrder{})},
				Price: sdk.NewInt64Coin("pcoin", 60),
			},
			{
				// This one is similar to [0], but with a different order id.
				// It'll be used to test the case where the bid split isn't found.
				Order: &OrderFulfillment{Order: NewOrder(bidOrderID + 3).WithBid(&BidOrder{})},
				Price: sdk.NewInt64Coin("pcoin", 1851),
			},
		}

		for _, i := range bidSplitIndexes {
			tc.askSplit.Order.Splits = append(tc.askSplit.Order.Splits, bidSplits[i])
			if i == 0 {
				tc.expAskSplit.Order.Splits = append(tc.expAskSplit.Order.Splits, &OrderSplit{
					Order: &OrderFulfillment{Order: NewOrder(bidOrderID).WithBid(&BidOrder{})},
					Price: sdk.NewInt64Coin("pcoin", 1853), // == 5555 * 50 / 150 + 2 leftover.
				})
			} else {
				tc.expAskSplit.Order.Splits = append(tc.expAskSplit.Order.Splits, copyOrderSplit(bidSplits[i]))
			}
			if i == 3 {
				tc.expPanic = "could not apply leftover amount 2 from bid order 200 to ask order 1: bid split not found"
			}
		}

		return tc
	}

	tests := []testCase{
		newTestCase("applies to first bid split", 0, 1, 2),
		newTestCase("applies to second bid split", 2, 0, 1),
		newTestCase("applies to third bid split", 1, 2, 0),
		newTestCase("bid split not found", 1, 2, 3),
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			origFulfillment := copyOrderFulfillment(tc.receiver)
			origSplit := copyOrderSplit(tc.askSplit)

			testFunc := func() {
				tc.receiver.ApplyLeftoverPrice(tc.askSplit, tc.amt)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "ApplyLeftoverPrice")
			if !assertEqualOrderFulfillments(t, tc.expFulfillment, tc.receiver, "OrderFulfillment after .ApplyLeftoverPrice") {
				t.Logf("Original: %s", orderFulfillmentString(origFulfillment))
			}
			if !assert.Equal(t, tc.expAskSplit, tc.askSplit, "askSplit after ApplyLeftoverPrice") {
				t.Logf("Original askSplit: %s", orderSplitString(origSplit))
				t.Logf("  Actual askSplit: %s", orderSplitString(tc.askSplit))
				t.Logf("Expected askSplit: %s", orderSplitString(tc.expAskSplit))
			}
		})
	}
}

func TestOrderFulfillment_Finalize(t *testing.T) {
	sdkNewInt64CoinP := func(denom string, amount int64) *sdk.Coin {
		rv := sdk.NewInt64Coin(denom, amount)
		return &rv
	}

	tests := []struct {
		name           string
		receiver       *OrderFulfillment
		sellerFeeRatio *FeeRatio
		expResult      *OrderFulfillment
		expErr         string
	}{
		{
			name: "ask assets filled zero",
			receiver: &OrderFulfillment{
				Order:           NewOrder(3).WithAsk(&AskOrder{}),
				AssetsFilledAmt: sdkmath.ZeroInt(),
			},
			expErr: "no assets filled in ask order 3",
		},
		{
			name: "ask assets filled negative",
			receiver: &OrderFulfillment{
				Order:           NewOrder(3).WithAsk(&AskOrder{}),
				AssetsFilledAmt: sdkmath.NewInt(-8),
			},
			expErr: "no assets filled in ask order 3",
		},
		{
			name: "bid assets filled zero",
			receiver: &OrderFulfillment{
				Order:           NewOrder(3).WithBid(&BidOrder{}),
				AssetsFilledAmt: sdkmath.ZeroInt(),
			},
			expErr: "no assets filled in bid order 3",
		},
		{
			name: "bid assets filled negative",
			receiver: &OrderFulfillment{
				Order:           NewOrder(3).WithBid(&BidOrder{}),
				AssetsFilledAmt: sdkmath.NewInt(-8),
			},
			expErr: "no assets filled in bid order 3",
		},

		{
			name: "ask partial price not evenly divisible",
			receiver: &OrderFulfillment{
				Order: NewOrder(3).WithAsk(&AskOrder{
					Assets: sdk.NewInt64Coin("apple", 50),
					Price:  sdk.NewInt64Coin("pear", 101),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				AssetsUnfilledAmt: sdkmath.NewInt(40),
			},
			expErr: `ask order 3 having assets "50apple" cannot be partially filled by "10apple": price "101pear" is not evenly divisible`,
		},
		{
			name: "bid partial price not evenly divisible",
			receiver: &OrderFulfillment{
				Order: NewOrder(3).WithBid(&BidOrder{
					Assets: sdk.NewInt64Coin("apple", 50),
					Price:  sdk.NewInt64Coin("pear", 101),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				AssetsUnfilledAmt: sdkmath.NewInt(40),
			},
			expErr: `bid order 3 having assets "50apple" cannot be partially filled by "10apple": price "101pear" is not evenly divisible`,
		},

		{
			name: "ask partial fees not evenly divisible",
			receiver: &OrderFulfillment{
				Order: NewOrder(3).WithAsk(&AskOrder{
					Assets:                  sdk.NewInt64Coin("apple", 50),
					Price:                   sdk.NewInt64Coin("pear", 100),
					SellerSettlementFlatFee: sdkNewInt64CoinP("fig", 201),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				AssetsUnfilledAmt: sdkmath.NewInt(40),
			},
			expErr: `ask order 3 having assets "50apple" cannot be partially filled by "10apple": fee "201fig" is not evenly divisible`,
		},
		{
			name: "bid partial fees not evenly divisible",
			receiver: &OrderFulfillment{
				Order: NewOrder(3).WithBid(&BidOrder{
					Assets:              sdk.NewInt64Coin("apple", 50),
					Price:               sdk.NewInt64Coin("pear", 100),
					BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("fig", 200), sdk.NewInt64Coin("grape", 151)),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				AssetsUnfilledAmt: sdkmath.NewInt(40),
			},
			expErr: `bid order 3 having assets "50apple" cannot be partially filled by "10apple": fee "151grape" is not evenly divisible`,
		},

		{
			name: "ask ratio calc error",
			receiver: &OrderFulfillment{
				Order: NewOrder(3).WithAsk(&AskOrder{
					Assets:                  sdk.NewInt64Coin("apple", 50),
					Price:                   sdk.NewInt64Coin("pear", 100),
					SellerSettlementFlatFee: sdkNewInt64CoinP("fig", 200),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				AssetsUnfilledAmt: sdkmath.NewInt(40),
				PriceAppliedAmt:   sdkmath.NewInt(29),
				PriceLeftAmt:      sdkmath.NewInt(71),
			},
			sellerFeeRatio: &FeeRatio{
				Price: sdk.NewInt64Coin("plum", 1),
				Fee:   sdk.NewInt64Coin("fig", 3),
			},
			expErr: "could not calculate ask order 3 ratio fee: cannot apply ratio 1plum:3fig to price 29pear: incorrect price denom",
		},

		{
			name: "ask full no ratio",
			receiver: &OrderFulfillment{
				Order: NewOrder(3).WithAsk(&AskOrder{
					Assets:                  sdk.NewInt64Coin("apple", 50),
					Price:                   sdk.NewInt64Coin("pear", 100),
					SellerSettlementFlatFee: sdkNewInt64CoinP("fig", 200),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: sdkmath.NewInt(0),
			},
			sellerFeeRatio: nil,
			expResult: &OrderFulfillment{
				Order: NewOrder(3).WithAsk(&AskOrder{
					Assets:                  sdk.NewInt64Coin("apple", 50),
					Price:                   sdk.NewInt64Coin("pear", 100),
					SellerSettlementFlatFee: sdkNewInt64CoinP("fig", 200),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: sdkmath.NewInt(0),
				IsFinalized:       true,
				PriceFilledAmt:    sdkmath.NewInt(100),
				PriceUnfilledAmt:  sdkmath.ZeroInt(),
				FeesToPay:         sdk.NewCoins(sdk.NewInt64Coin("fig", 200)),
				OrderFeesLeft:     nil,
			},
		},
		{
			name: "ask full, exact ratio",
			receiver: &OrderFulfillment{
				Order: NewOrder(3).WithAsk(&AskOrder{
					Assets:                  sdk.NewInt64Coin("apple", 50),
					Price:                   sdk.NewInt64Coin("pear", 100),
					SellerSettlementFlatFee: sdkNewInt64CoinP("fig", 200),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: sdkmath.NewInt(0),
				PriceAppliedAmt:   sdkmath.NewInt(110),
				PriceLeftAmt:      sdkmath.NewInt(-10),
			},
			sellerFeeRatio: &FeeRatio{
				Price: sdk.NewInt64Coin("pear", 10),
				Fee:   sdk.NewInt64Coin("grape", 1),
			},
			expResult: &OrderFulfillment{
				Order: NewOrder(3).WithAsk(&AskOrder{
					Assets:                  sdk.NewInt64Coin("apple", 50),
					Price:                   sdk.NewInt64Coin("pear", 100),
					SellerSettlementFlatFee: sdkNewInt64CoinP("fig", 200),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: sdkmath.NewInt(0),
				PriceAppliedAmt:   sdkmath.NewInt(110),
				PriceLeftAmt:      sdkmath.NewInt(-10),
				IsFinalized:       true,
				PriceFilledAmt:    sdkmath.NewInt(100),
				PriceUnfilledAmt:  sdkmath.ZeroInt(),
				FeesToPay:         sdk.NewCoins(sdk.NewInt64Coin("fig", 200), sdk.NewInt64Coin("grape", 11)),
				OrderFeesLeft:     nil,
			},
		},
		{
			name: "ask full, loose ratio",
			receiver: &OrderFulfillment{
				Order: NewOrder(3).WithAsk(&AskOrder{
					Assets:                  sdk.NewInt64Coin("apple", 50),
					Price:                   sdk.NewInt64Coin("pear", 100),
					SellerSettlementFlatFee: sdkNewInt64CoinP("fig", 200),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: sdkmath.NewInt(0),
				PriceAppliedAmt:   sdkmath.NewInt(110),
				PriceLeftAmt:      sdkmath.NewInt(-10),
			},
			sellerFeeRatio: &FeeRatio{
				Price: sdk.NewInt64Coin("pear", 13),
				Fee:   sdk.NewInt64Coin("grape", 1),
			},
			expResult: &OrderFulfillment{
				Order: NewOrder(3).WithAsk(&AskOrder{
					Assets:                  sdk.NewInt64Coin("apple", 50),
					Price:                   sdk.NewInt64Coin("pear", 100),
					SellerSettlementFlatFee: sdkNewInt64CoinP("fig", 200),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: sdkmath.NewInt(0),
				PriceAppliedAmt:   sdkmath.NewInt(110),
				PriceLeftAmt:      sdkmath.NewInt(-10),
				IsFinalized:       true,
				PriceFilledAmt:    sdkmath.NewInt(100),
				PriceUnfilledAmt:  sdkmath.ZeroInt(),
				FeesToPay:         sdk.NewCoins(sdk.NewInt64Coin("fig", 200), sdk.NewInt64Coin("grape", 9)),
				OrderFeesLeft:     nil,
			},
		},
		{
			name: "ask partial no ratio",
			receiver: &OrderFulfillment{
				Order: NewOrder(3).WithAsk(&AskOrder{
					Assets:                  sdk.NewInt64Coin("apple", 50),
					Price:                   sdk.NewInt64Coin("pear", 100),
					SellerSettlementFlatFee: sdkNewInt64CoinP("fig", 200),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: sdkmath.NewInt(30),
			},
			sellerFeeRatio: nil,
			expResult: &OrderFulfillment{
				Order: NewOrder(3).WithAsk(&AskOrder{
					Assets:                  sdk.NewInt64Coin("apple", 50),
					Price:                   sdk.NewInt64Coin("pear", 100),
					SellerSettlementFlatFee: sdkNewInt64CoinP("fig", 200),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: sdkmath.NewInt(30),
				IsFinalized:       true,
				PriceFilledAmt:    sdkmath.NewInt(40),
				PriceUnfilledAmt:  sdkmath.NewInt(60),
				FeesToPay:         sdk.NewCoins(sdk.NewInt64Coin("fig", 80)),
				OrderFeesLeft:     sdk.NewCoins(sdk.NewInt64Coin("fig", 120)),
			},
		},
		{
			name: "ask partial, exact ratio",
			receiver: &OrderFulfillment{
				Order: NewOrder(3).WithAsk(&AskOrder{
					Assets:                  sdk.NewInt64Coin("apple", 50),
					Price:                   sdk.NewInt64Coin("pear", 100),
					SellerSettlementFlatFee: sdkNewInt64CoinP("fig", 200),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: sdkmath.NewInt(30),
				PriceAppliedAmt:   sdkmath.NewInt(110),
				PriceLeftAmt:      sdkmath.NewInt(-10),
			},
			sellerFeeRatio: &FeeRatio{
				Price: sdk.NewInt64Coin("pear", 10),
				Fee:   sdk.NewInt64Coin("grape", 1),
			},
			expResult: &OrderFulfillment{
				Order: NewOrder(3).WithAsk(&AskOrder{
					Assets:                  sdk.NewInt64Coin("apple", 50),
					Price:                   sdk.NewInt64Coin("pear", 100),
					SellerSettlementFlatFee: sdkNewInt64CoinP("fig", 200),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: sdkmath.NewInt(30),
				PriceAppliedAmt:   sdkmath.NewInt(110),
				PriceLeftAmt:      sdkmath.NewInt(-10),
				IsFinalized:       true,
				PriceFilledAmt:    sdkmath.NewInt(40),
				PriceUnfilledAmt:  sdkmath.NewInt(60),
				FeesToPay:         sdk.NewCoins(sdk.NewInt64Coin("fig", 80), sdk.NewInt64Coin("grape", 11)),
				OrderFeesLeft:     sdk.NewCoins(sdk.NewInt64Coin("fig", 120)),
			},
		},
		{
			name: "ask partial, loose ratio",
			receiver: &OrderFulfillment{
				Order: NewOrder(3).WithAsk(&AskOrder{
					Assets:                  sdk.NewInt64Coin("apple", 50),
					Price:                   sdk.NewInt64Coin("pear", 100),
					SellerSettlementFlatFee: sdkNewInt64CoinP("fig", 200),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: sdkmath.NewInt(30),
				PriceAppliedAmt:   sdkmath.NewInt(110),
				PriceLeftAmt:      sdkmath.NewInt(-10),
			},
			sellerFeeRatio: &FeeRatio{
				Price: sdk.NewInt64Coin("pear", 13),
				Fee:   sdk.NewInt64Coin("fig", 1),
			},
			expResult: &OrderFulfillment{
				Order: NewOrder(3).WithAsk(&AskOrder{
					Assets:                  sdk.NewInt64Coin("apple", 50),
					Price:                   sdk.NewInt64Coin("pear", 100),
					SellerSettlementFlatFee: sdkNewInt64CoinP("fig", 200),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: sdkmath.NewInt(30),
				PriceAppliedAmt:   sdkmath.NewInt(110),
				PriceLeftAmt:      sdkmath.NewInt(-10),
				IsFinalized:       true,
				PriceFilledAmt:    sdkmath.NewInt(40),
				PriceUnfilledAmt:  sdkmath.NewInt(60),
				FeesToPay:         sdk.NewCoins(sdk.NewInt64Coin("fig", 89)),
				OrderFeesLeft:     sdk.NewCoins(sdk.NewInt64Coin("fig", 120)),
			},
		},

		{
			name: "bid full no leftovers",
			receiver: &OrderFulfillment{
				Order: NewOrder(3).WithBid(&BidOrder{
					Assets:              sdk.NewInt64Coin("apple", 50),
					Price:               sdk.NewInt64Coin("pear", 100),
					BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("fig", 200), sdk.NewInt64Coin("grape", 13)),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: sdkmath.NewInt(0),
				PriceAppliedAmt:   sdkmath.NewInt(100),
				PriceLeftAmt:      sdkmath.NewInt(0),
			},
			expResult: &OrderFulfillment{
				Order: NewOrder(3).WithBid(&BidOrder{
					Assets:              sdk.NewInt64Coin("apple", 50),
					Price:               sdk.NewInt64Coin("pear", 100),
					BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("fig", 200), sdk.NewInt64Coin("grape", 13)),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: sdkmath.NewInt(0),
				PriceAppliedAmt:   sdkmath.NewInt(100),
				PriceLeftAmt:      sdkmath.NewInt(0),
				IsFinalized:       true,
				PriceFilledAmt:    sdkmath.NewInt(100),
				PriceUnfilledAmt:  sdkmath.NewInt(0),
				FeesToPay:         sdk.NewCoins(sdk.NewInt64Coin("fig", 200), sdk.NewInt64Coin("grape", 13)),
				OrderFeesLeft:     nil,
			},
		},
		{
			name: "bid full with leftovers",
			receiver: &OrderFulfillment{
				Order: NewOrder(3).WithBid(&BidOrder{
					Assets:              sdk.NewInt64Coin("apple", 50),
					Price:               sdk.NewInt64Coin("pear", 1000), // 1000 / 50 = 20 per asset.
					BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("fig", 200), sdk.NewInt64Coin("grape", 13)),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: sdkmath.NewInt(0),
				PriceAppliedAmt:   sdkmath.NewInt(993),
				PriceLeftAmt:      sdkmath.NewInt(7),
				Splits: []*OrderSplit{
					{
						// This one will get 1 once the loop defaults to 1.
						// So, 7 * split assets / 50 filled must be 0.
						Order: &OrderFulfillment{
							Order: NewOrder(101).WithAsk(&AskOrder{
								Assets: sdk.NewInt64Coin("apple", 5),
								Price:  sdk.NewInt64Coin("pear", 100),
							}),
							Splits: []*OrderSplit{{
								Order:  &OrderFulfillment{Order: NewOrder(3).WithBid(&BidOrder{})},
								Assets: sdk.NewInt64Coin("apple", 5),
								Price:  sdk.NewInt64Coin("pear", 100),
							}},
							AssetsFilledAmt:   sdkmath.NewInt(5),
							AssetsUnfilledAmt: sdkmath.NewInt(0),
							PriceAppliedAmt:   sdkmath.NewInt(100),
							PriceLeftAmt:      sdkmath.NewInt(0),
						},
						Assets: sdk.NewInt64Coin("apple", 5),
						Price:  sdk.NewInt64Coin("pear", 100),
					},
					{
						// This one will not get anything more.
						// It's in the same situation as the one above, but the leftover will run out first.
						Order: &OrderFulfillment{
							Order: NewOrder(102).WithAsk(&AskOrder{
								Assets: sdk.NewInt64Coin("apple", 4),
								Price:  sdk.NewInt64Coin("pear", 80),
							}),
							Splits: []*OrderSplit{{
								Order:  &OrderFulfillment{Order: NewOrder(3).WithBid(&BidOrder{})},
								Assets: sdk.NewInt64Coin("apple", 4),
								Price:  sdk.NewInt64Coin("pear", 80),
							}},
							AssetsFilledAmt:   sdkmath.NewInt(4),
							AssetsUnfilledAmt: sdkmath.NewInt(0),
							PriceAppliedAmt:   sdkmath.NewInt(80),
							PriceLeftAmt:      sdkmath.NewInt(0),
						},
						Assets: sdk.NewInt64Coin("apple", 4),
						Price:  sdk.NewInt64Coin("pear", 80),
					},
					{
						// This one will get 4 in the first pass of the loop.
						// I.e. 7 * split assets / 50 = 4. Assets 29 to 39
						Order: &OrderFulfillment{
							Order: NewOrder(103).WithAsk(&AskOrder{
								Assets: sdk.NewInt64Coin("apple", 35),
								Price:  sdk.NewInt64Coin("pear", 693),
							}),
							Splits: []*OrderSplit{{
								Order:  &OrderFulfillment{Order: NewOrder(3).WithBid(&BidOrder{})},
								Assets: sdk.NewInt64Coin("apple", 35),
								Price:  sdk.NewInt64Coin("pear", 693),
							}},
							AssetsFilledAmt:   sdkmath.NewInt(35),
							AssetsUnfilledAmt: sdkmath.NewInt(0),
							PriceAppliedAmt:   sdkmath.NewInt(693),
							PriceLeftAmt:      sdkmath.NewInt(0),
						},
						Assets: sdk.NewInt64Coin("apple", 35),
						Price:  sdk.NewInt64Coin("pear", 693),
					},
					{
						// This one will get 2 due to price left.
						// I also need this one to have 7 * assets / 50 = 0, so it doesn't get 1 more on the first pass.
						Order: &OrderFulfillment{
							Order: NewOrder(104).WithAsk(&AskOrder{
								Assets: sdk.NewInt64Coin("apple", 6),
								Price:  sdk.NewInt64Coin("pear", 122),
							}),
							Splits: []*OrderSplit{{
								Order:  &OrderFulfillment{Order: NewOrder(3).WithBid(&BidOrder{})},
								Assets: sdk.NewInt64Coin("apple", 6),
								Price:  sdk.NewInt64Coin("pear", 120),
							}},
							AssetsFilledAmt:   sdkmath.NewInt(6),
							AssetsUnfilledAmt: sdkmath.NewInt(0),
							PriceAppliedAmt:   sdkmath.NewInt(120),
							PriceLeftAmt:      sdkmath.NewInt(2),
						},
						Assets: sdk.NewInt64Coin("apple", 6),
						Price:  sdk.NewInt64Coin("pear", 120),
					},
				},
			},
			expResult: &OrderFulfillment{
				Order: NewOrder(3).WithBid(&BidOrder{
					Assets:              sdk.NewInt64Coin("apple", 50),
					Price:               sdk.NewInt64Coin("pear", 1000),
					BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("fig", 200), sdk.NewInt64Coin("grape", 13)),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: sdkmath.NewInt(0),
				PriceAppliedAmt:   sdkmath.NewInt(1000),
				PriceLeftAmt:      ZeroAmtAfterSub,
				Splits: []*OrderSplit{
					{
						// This one should get 1 once the loop defaults to 1.
						Order: &OrderFulfillment{
							Order: NewOrder(101).WithAsk(&AskOrder{
								Assets: sdk.NewInt64Coin("apple", 5),
								Price:  sdk.NewInt64Coin("pear", 100),
							}),
							Splits: []*OrderSplit{{
								Order:  &OrderFulfillment{Order: NewOrder(3).WithBid(&BidOrder{})},
								Assets: sdk.NewInt64Coin("apple", 5),
								Price:  sdk.NewInt64Coin("pear", 101),
							}},
							AssetsFilledAmt:   sdkmath.NewInt(5),
							AssetsUnfilledAmt: sdkmath.NewInt(0),
							PriceAppliedAmt:   sdkmath.NewInt(101),
							PriceLeftAmt:      sdkmath.NewInt(-1),
						},
						Assets: sdk.NewInt64Coin("apple", 5),
						Price:  sdk.NewInt64Coin("pear", 101),
					},
					{
						// This one will not get anything more.
						Order: &OrderFulfillment{
							Order: NewOrder(102).WithAsk(&AskOrder{
								Assets: sdk.NewInt64Coin("apple", 4),
								Price:  sdk.NewInt64Coin("pear", 80),
							}),
							Splits: []*OrderSplit{{
								Order:  &OrderFulfillment{Order: NewOrder(3).WithBid(&BidOrder{})},
								Assets: sdk.NewInt64Coin("apple", 4),
								Price:  sdk.NewInt64Coin("pear", 80),
							}},
							AssetsFilledAmt:   sdkmath.NewInt(4),
							AssetsUnfilledAmt: sdkmath.NewInt(0),
							PriceAppliedAmt:   sdkmath.NewInt(80),
							PriceLeftAmt:      sdkmath.NewInt(0),
						},
						Assets: sdk.NewInt64Coin("apple", 4),
						Price:  sdk.NewInt64Coin("pear", 80),
					},
					{
						// This one should get 4 in the first pass of the loop.
						Order: &OrderFulfillment{
							Order: NewOrder(103).WithAsk(&AskOrder{
								Assets: sdk.NewInt64Coin("apple", 35),
								Price:  sdk.NewInt64Coin("pear", 693),
							}),
							Splits: []*OrderSplit{{
								Order:  &OrderFulfillment{Order: NewOrder(3).WithBid(&BidOrder{})},
								Assets: sdk.NewInt64Coin("apple", 35),
								Price:  sdk.NewInt64Coin("pear", 697),
							}},
							AssetsFilledAmt:   sdkmath.NewInt(35),
							AssetsUnfilledAmt: sdkmath.NewInt(0),
							PriceAppliedAmt:   sdkmath.NewInt(697),
							PriceLeftAmt:      sdkmath.NewInt(-4),
						},
						Assets: sdk.NewInt64Coin("apple", 35),
						Price:  sdk.NewInt64Coin("pear", 697),
					},
					{
						// this one will get 2 due to price left.
						// I also need this one to have 7 * assets / 50 = 0, so it doesn't get 1 more on the first pass.
						Order: &OrderFulfillment{
							Order: NewOrder(104).WithAsk(&AskOrder{
								Assets: sdk.NewInt64Coin("apple", 6),
								Price:  sdk.NewInt64Coin("pear", 122),
							}),
							Splits: []*OrderSplit{{
								Order:  &OrderFulfillment{Order: NewOrder(3).WithBid(&BidOrder{})},
								Assets: sdk.NewInt64Coin("apple", 6),
								Price:  sdk.NewInt64Coin("pear", 122),
							}},
							AssetsFilledAmt:   sdkmath.NewInt(6),
							AssetsUnfilledAmt: sdkmath.NewInt(0),
							PriceAppliedAmt:   sdkmath.NewInt(122),
							PriceLeftAmt:      ZeroAmtAfterSub,
						},
						Assets: sdk.NewInt64Coin("apple", 6),
						Price:  sdk.NewInt64Coin("pear", 122),
					},
				},
				IsFinalized:      true,
				PriceFilledAmt:   sdkmath.NewInt(1000),
				PriceUnfilledAmt: sdkmath.NewInt(0),
				FeesToPay:        sdk.NewCoins(sdk.NewInt64Coin("fig", 200), sdk.NewInt64Coin("grape", 13)),
				OrderFeesLeft:    nil,
			},
		},
		{
			name: "bid partial no leftovers",
			receiver: &OrderFulfillment{
				Order: NewOrder(3).WithBid(&BidOrder{
					Assets:              sdk.NewInt64Coin("apple", 75),
					Price:               sdk.NewInt64Coin("pear", 150),
					BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("fig", 300), sdk.NewInt64Coin("grape", 12)),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: sdkmath.NewInt(25),
				PriceAppliedAmt:   sdkmath.NewInt(100),
				PriceLeftAmt:      sdkmath.NewInt(50),
			},
			expResult: &OrderFulfillment{
				Order: NewOrder(3).WithBid(&BidOrder{
					Assets:              sdk.NewInt64Coin("apple", 75),
					Price:               sdk.NewInt64Coin("pear", 150),
					BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("fig", 300), sdk.NewInt64Coin("grape", 12)),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: sdkmath.NewInt(25),
				PriceAppliedAmt:   sdkmath.NewInt(100),
				PriceLeftAmt:      sdkmath.NewInt(50),
				IsFinalized:       true,
				PriceFilledAmt:    sdkmath.NewInt(100),
				PriceUnfilledAmt:  sdkmath.NewInt(50),
				FeesToPay:         sdk.NewCoins(sdk.NewInt64Coin("fig", 200), sdk.NewInt64Coin("grape", 8)),
				OrderFeesLeft:     sdk.NewCoins(sdk.NewInt64Coin("fig", 100), sdk.NewInt64Coin("grape", 4)),
			},
		},
		{
			name: "bid partial with leftovers",
			receiver: &OrderFulfillment{
				Order: NewOrder(3).WithBid(&BidOrder{
					Assets:              sdk.NewInt64Coin("apple", 75),
					Price:               sdk.NewInt64Coin("pear", 1500), // 1020 / 51 = 20 per asset.
					BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("fig", 300), sdk.NewInt64Coin("grape", 12)),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: sdkmath.NewInt(25),
				PriceAppliedAmt:   sdkmath.NewInt(993),
				PriceLeftAmt:      sdkmath.NewInt(507),
				Splits: []*OrderSplit{
					{
						// This one will get 1 once the loop defaults to 1.
						// So, 7 * split assets / 50 filled must be 0.
						Order: &OrderFulfillment{
							Order: NewOrder(101).WithAsk(&AskOrder{
								Assets: sdk.NewInt64Coin("apple", 5),
								Price:  sdk.NewInt64Coin("pear", 100),
							}),
							Splits: []*OrderSplit{{
								Order:  &OrderFulfillment{Order: NewOrder(3).WithBid(&BidOrder{})},
								Assets: sdk.NewInt64Coin("apple", 5),
								Price:  sdk.NewInt64Coin("pear", 100),
							}},
							AssetsFilledAmt:   sdkmath.NewInt(5),
							AssetsUnfilledAmt: sdkmath.NewInt(0),
							PriceAppliedAmt:   sdkmath.NewInt(100),
							PriceLeftAmt:      sdkmath.NewInt(0),
						},
						Assets: sdk.NewInt64Coin("apple", 5),
						Price:  sdk.NewInt64Coin("pear", 100),
					},
					{
						// This one will not get anything more.
						// It's in the same situation as the one above, but the leftover will run out first.
						Order: &OrderFulfillment{
							Order: NewOrder(102).WithAsk(&AskOrder{
								Assets: sdk.NewInt64Coin("apple", 4),
								Price:  sdk.NewInt64Coin("pear", 80),
							}),
							Splits: []*OrderSplit{{
								Order:  &OrderFulfillment{Order: NewOrder(3).WithBid(&BidOrder{})},
								Assets: sdk.NewInt64Coin("apple", 4),
								Price:  sdk.NewInt64Coin("pear", 80),
							}},
							AssetsFilledAmt:   sdkmath.NewInt(4),
							AssetsUnfilledAmt: sdkmath.NewInt(0),
							PriceAppliedAmt:   sdkmath.NewInt(80),
							PriceLeftAmt:      sdkmath.NewInt(0),
						},
						Assets: sdk.NewInt64Coin("apple", 4),
						Price:  sdk.NewInt64Coin("pear", 80),
					},
					{
						// This one will get 4 in the first pass of the loop.
						// I.e. 7 * split assets / 50 = 4. Assets 29 to 39
						Order: &OrderFulfillment{
							Order: NewOrder(103).WithAsk(&AskOrder{
								Assets: sdk.NewInt64Coin("apple", 35),
								Price:  sdk.NewInt64Coin("pear", 693),
							}),
							Splits: []*OrderSplit{{
								Order:  &OrderFulfillment{Order: NewOrder(3).WithBid(&BidOrder{})},
								Assets: sdk.NewInt64Coin("apple", 35),
								Price:  sdk.NewInt64Coin("pear", 693),
							}},
							AssetsFilledAmt:   sdkmath.NewInt(35),
							AssetsUnfilledAmt: sdkmath.NewInt(0),
							PriceAppliedAmt:   sdkmath.NewInt(693),
							PriceLeftAmt:      sdkmath.NewInt(0),
						},
						Assets: sdk.NewInt64Coin("apple", 35),
						Price:  sdk.NewInt64Coin("pear", 693),
					},
					{
						// This one will get 2 due to price left.
						// I also need this one to have 7 * assets / 50 = 0, so it doesn't get 1 more on the first pass.
						Order: &OrderFulfillment{
							Order: NewOrder(104).WithAsk(&AskOrder{
								Assets: sdk.NewInt64Coin("apple", 6),
								Price:  sdk.NewInt64Coin("pear", 122),
							}),
							Splits: []*OrderSplit{{
								Order:  &OrderFulfillment{Order: NewOrder(3).WithBid(&BidOrder{})},
								Assets: sdk.NewInt64Coin("apple", 6),
								Price:  sdk.NewInt64Coin("pear", 120),
							}},
							AssetsFilledAmt:   sdkmath.NewInt(6),
							AssetsUnfilledAmt: sdkmath.NewInt(0),
							PriceAppliedAmt:   sdkmath.NewInt(120),
							PriceLeftAmt:      sdkmath.NewInt(2),
						},
						Assets: sdk.NewInt64Coin("apple", 6),
						Price:  sdk.NewInt64Coin("pear", 120),
					},
				},
			},
			expResult: &OrderFulfillment{
				Order: NewOrder(3).WithBid(&BidOrder{
					Assets:              sdk.NewInt64Coin("apple", 75),
					Price:               sdk.NewInt64Coin("pear", 1500),
					BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("fig", 300), sdk.NewInt64Coin("grape", 12)),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: sdkmath.NewInt(25),
				PriceAppliedAmt:   sdkmath.NewInt(1000),
				PriceLeftAmt:      sdkmath.NewInt(500),
				Splits: []*OrderSplit{
					{
						// This one should get 1 once the loop defaults to 1.
						Order: &OrderFulfillment{
							Order: NewOrder(101).WithAsk(&AskOrder{
								Assets: sdk.NewInt64Coin("apple", 5),
								Price:  sdk.NewInt64Coin("pear", 100),
							}),
							Splits: []*OrderSplit{{
								Order:  &OrderFulfillment{Order: NewOrder(3).WithBid(&BidOrder{})},
								Assets: sdk.NewInt64Coin("apple", 5),
								Price:  sdk.NewInt64Coin("pear", 101),
							}},
							AssetsFilledAmt:   sdkmath.NewInt(5),
							AssetsUnfilledAmt: sdkmath.NewInt(0),
							PriceAppliedAmt:   sdkmath.NewInt(101),
							PriceLeftAmt:      sdkmath.NewInt(-1),
						},
						Assets: sdk.NewInt64Coin("apple", 5),
						Price:  sdk.NewInt64Coin("pear", 101),
					},
					{
						// This one will not get anything more.
						Order: &OrderFulfillment{
							Order: NewOrder(102).WithAsk(&AskOrder{
								Assets: sdk.NewInt64Coin("apple", 4),
								Price:  sdk.NewInt64Coin("pear", 80),
							}),
							Splits: []*OrderSplit{{
								Order:  &OrderFulfillment{Order: NewOrder(3).WithBid(&BidOrder{})},
								Assets: sdk.NewInt64Coin("apple", 4),
								Price:  sdk.NewInt64Coin("pear", 80),
							}},
							AssetsFilledAmt:   sdkmath.NewInt(4),
							AssetsUnfilledAmt: sdkmath.NewInt(0),
							PriceAppliedAmt:   sdkmath.NewInt(80),
							PriceLeftAmt:      sdkmath.NewInt(0),
						},
						Assets: sdk.NewInt64Coin("apple", 4),
						Price:  sdk.NewInt64Coin("pear", 80),
					},
					{
						// This one should get 4 in the first pass of the loop.
						Order: &OrderFulfillment{
							Order: NewOrder(103).WithAsk(&AskOrder{
								Assets: sdk.NewInt64Coin("apple", 35),
								Price:  sdk.NewInt64Coin("pear", 693),
							}),
							Splits: []*OrderSplit{{
								Order:  &OrderFulfillment{Order: NewOrder(3).WithBid(&BidOrder{})},
								Assets: sdk.NewInt64Coin("apple", 35),
								Price:  sdk.NewInt64Coin("pear", 697),
							}},
							AssetsFilledAmt:   sdkmath.NewInt(35),
							AssetsUnfilledAmt: sdkmath.NewInt(0),
							PriceAppliedAmt:   sdkmath.NewInt(697),
							PriceLeftAmt:      sdkmath.NewInt(-4),
						},
						Assets: sdk.NewInt64Coin("apple", 35),
						Price:  sdk.NewInt64Coin("pear", 697),
					},
					{
						// this one will get 2 due to price left.
						// I also need this one to have 7 * assets / 50 = 0, so it doesn't get 1 more on the first pass.
						Order: &OrderFulfillment{
							Order: NewOrder(104).WithAsk(&AskOrder{
								Assets: sdk.NewInt64Coin("apple", 6),
								Price:  sdk.NewInt64Coin("pear", 122),
							}),
							Splits: []*OrderSplit{{
								Order:  &OrderFulfillment{Order: NewOrder(3).WithBid(&BidOrder{})},
								Assets: sdk.NewInt64Coin("apple", 6),
								Price:  sdk.NewInt64Coin("pear", 122),
							}},
							AssetsFilledAmt:   sdkmath.NewInt(6),
							AssetsUnfilledAmt: sdkmath.NewInt(0),
							PriceAppliedAmt:   sdkmath.NewInt(122),
							PriceLeftAmt:      ZeroAmtAfterSub,
						},
						Assets: sdk.NewInt64Coin("apple", 6),
						Price:  sdk.NewInt64Coin("pear", 122),
					},
				},
				IsFinalized:      true,
				PriceFilledAmt:   sdkmath.NewInt(1000),
				PriceUnfilledAmt: sdkmath.NewInt(500),
				FeesToPay:        sdk.NewCoins(sdk.NewInt64Coin("fig", 200), sdk.NewInt64Coin("grape", 8)),
				OrderFeesLeft:    sdk.NewCoins(sdk.NewInt64Coin("fig", 100), sdk.NewInt64Coin("grape", 4)),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := copyOrderFulfillment(tc.receiver)
			if tc.expResult == nil {
				tc.expResult = copyOrderFulfillment(tc.receiver)
				tc.expResult.PriceFilledAmt = sdkmath.ZeroInt()
				tc.expResult.PriceUnfilledAmt = sdkmath.ZeroInt()
			}

			var err error
			testFunc := func() {
				err = tc.receiver.Finalize(tc.sellerFeeRatio)
			}
			require.NotPanics(t, testFunc, "Finalize")
			assertions.AssertErrorValue(t, err, tc.expErr, "Finalize error")
			if !assertEqualOrderFulfillments(t, tc.expResult, tc.receiver, "receiver after Finalize") {
				t.Logf("Original: %s", orderFulfillmentString(orig))
			}
		})
	}
}

// TODO[1658]: func TestOrderFulfillment_Validate(t *testing.T)

func TestOrderFulfillment_Validate2(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	coinP := func(amount int64, denom string) *sdk.Coin {
		return &sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	coins := func(amount int64, denom string) sdk.Coins {
		return sdk.Coins{coin(amount, denom)}
	}

	tests := []struct {
		name     string
		receiver OrderFulfillment
		expErr   string
	}{
		{
			name:     "nil order type",
			receiver: OrderFulfillment{Order: NewOrder(2)},
			expErr:   nilSubTypeErr(2),
		},
		{
			name:     "unknown order type",
			receiver: OrderFulfillment{Order: newUnknownOrder(3)},
			expErr:   unknownSubTypeErr(3),
		},
		{
			name: "not finalized, ask",
			receiver: OrderFulfillment{
				Order:       NewOrder(4).WithAsk(&AskOrder{}),
				IsFinalized: false,
			},
			expErr: "fulfillment for ask order 4 has not been finalized",
		},
		{
			name: "not finalized, bid",
			receiver: OrderFulfillment{
				Order:       NewOrder(4).WithBid(&BidOrder{}),
				IsFinalized: false,
			},
			expErr: "fulfillment for bid order 4 has not been finalized",
		},

		{
			name: "assets unfilled negative",
			receiver: OrderFulfillment{
				Order: NewOrder(5).WithAsk(&AskOrder{
					Assets: coin(55, "apple"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(-12),
				AssetsFilledAmt:   sdkmath.NewInt(97),
			},
			expErr: "ask order 5 having assets \"55apple\" has negative assets left \"-12apple\" after filling \"97apple\"",
		},
		{
			name: "assets filled zero",
			receiver: OrderFulfillment{
				Order: NewOrder(6).WithBid(&BidOrder{
					Assets: coin(55, "apple"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(55),
				AssetsFilledAmt:   sdkmath.NewInt(0),
			},
			expErr: "cannot fill non-positive assets \"0apple\" on bid order 6 having assets \"55apple\"",
		},
		{
			name: "assets filled negative",
			receiver: OrderFulfillment{
				Order: NewOrder(7).WithAsk(&AskOrder{
					Assets: coin(55, "apple"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(56),
				AssetsFilledAmt:   sdkmath.NewInt(-1),
			},
			expErr: "cannot fill non-positive assets \"-1apple\" on ask order 7 having assets \"55apple\"",
		},
		{
			name: "assets tracked too low",
			receiver: OrderFulfillment{
				Order: NewOrder(8).WithBid(&BidOrder{
					Assets: coin(55, "apple"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(44),
				AssetsFilledAmt:   sdkmath.NewInt(10),
			},
			expErr: "tracked assets \"54apple\" does not equal bid order 8 assets \"55apple\"",
		},
		{
			name: "assets tracked too high",
			receiver: OrderFulfillment{
				Order: NewOrder(8).WithAsk(&AskOrder{
					Assets: coin(55, "apple"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(44),
				AssetsFilledAmt:   sdkmath.NewInt(12),
			},
			expErr: "tracked assets \"56apple\" does not equal ask order 8 assets \"55apple\"",
		},

		{
			name: "price left equals order price",
			receiver: OrderFulfillment{
				Order: NewOrder(19).WithAsk(&AskOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98789),
			},
			expErr: "price left \"98789plum\" is not less than ask order 19 price \"98789plum\"",
		},
		{
			name: "price left more than order price",
			receiver: OrderFulfillment{
				Order: NewOrder(20).WithBid(&BidOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98790),
			},
			expErr: "price left \"98790plum\" is not less than bid order 20 price \"98789plum\"",
		},
		{
			name: "price applied zero",
			receiver: OrderFulfillment{
				Order: NewOrder(21).WithBid(&BidOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98788),
				PriceAppliedAmt:   sdkmath.NewInt(0),
			},
			expErr: "cannot apply non-positive price \"0plum\" to bid order 21 having price \"98789plum\"",
		},
		{
			name: "price applied negative",
			receiver: OrderFulfillment{
				Order: NewOrder(22).WithAsk(&AskOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98788),
				PriceAppliedAmt:   sdkmath.NewInt(-1),
			},
			expErr: "cannot apply non-positive price \"-1plum\" to ask order 22 having price \"98789plum\"",
		},
		{
			name: "price tracked too low",
			receiver: OrderFulfillment{
				Order: NewOrder(23).WithAsk(&AskOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98000),
				PriceAppliedAmt:   sdkmath.NewInt(788),
			},
			expErr: "tracked price \"98788plum\" does not equal ask order 23 price \"98789plum\"",
		},
		{
			name: "price tracked too high",
			receiver: OrderFulfillment{
				Order: NewOrder(24).WithBid(&BidOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98001),
				PriceAppliedAmt:   sdkmath.NewInt(789),
			},
			expErr: "tracked price \"98790plum\" does not equal bid order 24 price \"98789plum\"",
		},
		{
			name: "price filled zero",
			receiver: OrderFulfillment{
				Order: NewOrder(25).WithAsk(&AskOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98000),
				PriceAppliedAmt:   sdkmath.NewInt(789),
				PriceUnfilledAmt:  sdkmath.NewInt(98789),
				PriceFilledAmt:    sdkmath.NewInt(0),
			},
			expErr: "cannot fill ask order 25 having price \"98789plum\" with non-positive price \"0plum\"",
		},
		{
			name: "price filled negative",
			receiver: OrderFulfillment{
				Order: NewOrder(26).WithBid(&BidOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98000),
				PriceAppliedAmt:   sdkmath.NewInt(789),
				PriceUnfilledAmt:  sdkmath.NewInt(98790),
				PriceFilledAmt:    sdkmath.NewInt(-1),
			},
			expErr: "cannot fill bid order 26 having price \"98789plum\" with non-positive price \"-1plum\"",
		},
		{
			name: "total price too low",
			receiver: OrderFulfillment{
				Order: NewOrder(27).WithAsk(&AskOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98000),
				PriceAppliedAmt:   sdkmath.NewInt(789),
				PriceUnfilledAmt:  sdkmath.NewInt(90000),
				PriceFilledAmt:    sdkmath.NewInt(8788),
			},
			expErr: "filled price \"8788plum\" plus unfilled price \"90000plum\" does not equal order price \"98789plum\" for ask order 27",
		},
		{
			name: "total price too high",
			receiver: OrderFulfillment{
				Order: NewOrder(28).WithBid(&BidOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98000),
				PriceAppliedAmt:   sdkmath.NewInt(789),
				PriceUnfilledAmt:  sdkmath.NewInt(90001),
				PriceFilledAmt:    sdkmath.NewInt(8789),
			},
			expErr: "filled price \"8789plum\" plus unfilled price \"90001plum\" does not equal order price \"98789plum\" for bid order 28",
		},
		{
			name: "price unfilled negative",
			receiver: OrderFulfillment{
				Order: NewOrder(29).WithAsk(&AskOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98000),
				PriceAppliedAmt:   sdkmath.NewInt(789),
				PriceUnfilledAmt:  sdkmath.NewInt(-1),
				PriceFilledAmt:    sdkmath.NewInt(98790),
			},
			expErr: "ask order 29 having price \"98789plum\" has negative price \"-1plum\" after filling \"98790plum\"",
		},

		{
			name: "nil splits",
			receiver: OrderFulfillment{
				Order: NewOrder(100).WithBid(&BidOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98000),
				PriceAppliedAmt:   sdkmath.NewInt(789),
				PriceUnfilledAmt:  sdkmath.NewInt(90000),
				PriceFilledAmt:    sdkmath.NewInt(8789),
				Splits:            nil,
			},
			expErr: "no splits applied to bid order 100",
		},
		{
			name: "empty splits",
			receiver: OrderFulfillment{
				Order: NewOrder(101).WithAsk(&AskOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98000),
				PriceAppliedAmt:   sdkmath.NewInt(789),
				PriceUnfilledAmt:  sdkmath.NewInt(90000),
				PriceFilledAmt:    sdkmath.NewInt(8789),
				Splits:            []*OrderSplit{},
			},
			expErr: "no splits applied to ask order 101",
		},
		{
			name: "multiple asset denoms in splits",
			receiver: OrderFulfillment{
				Order: NewOrder(102).WithBid(&BidOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98000),
				PriceAppliedAmt:   sdkmath.NewInt(789),
				PriceUnfilledAmt:  sdkmath.NewInt(90000),
				PriceFilledAmt:    sdkmath.NewInt(8789),
				Splits: []*OrderSplit{
					{Assets: coin(3, "apple"), Price: coin(700, "plum")},
					{Assets: coin(5, "acai"), Price: coin(89, "plum")},
				},
			},
			expErr: "multiple asset denoms \"5acai,3apple\" in splits applied to bid order 102 having assets \"55apple\"",
		},
		{
			name: "wrong splits assets denom",
			receiver: OrderFulfillment{
				Order: NewOrder(103).WithAsk(&AskOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98000),
				PriceAppliedAmt:   sdkmath.NewInt(789),
				PriceUnfilledAmt:  sdkmath.NewInt(90000),
				PriceFilledAmt:    sdkmath.NewInt(8789),
				Splits: []*OrderSplit{
					{Assets: coin(3, "acai"), Price: coin(700, "plum")},
					{Assets: coin(5, "acai"), Price: coin(89, "plum")},
				},
			},
			expErr: "splits asset denom \"8acai\" does not equal order assets denom \"55apple\" on ask order 103",
		},
		{
			name: "splits assets total too low",
			receiver: OrderFulfillment{
				Order: NewOrder(104).WithBid(&BidOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98000),
				PriceAppliedAmt:   sdkmath.NewInt(789),
				PriceUnfilledAmt:  sdkmath.NewInt(90000),
				PriceFilledAmt:    sdkmath.NewInt(8789),
				Splits: []*OrderSplit{
					{Assets: coin(3, "apple"), Price: coin(800, "plum")},
					{Assets: coin(6, "apple"), Price: coin(89, "plum")},
				},
			},
			expErr: "splits asset total \"9apple\" does not equal filled assets \"10apple\" on bid order 104",
		},
		{
			name: "splits assets total too high",
			receiver: OrderFulfillment{
				Order: NewOrder(105).WithAsk(&AskOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98000),
				PriceAppliedAmt:   sdkmath.NewInt(789),
				PriceUnfilledAmt:  sdkmath.NewInt(90000),
				PriceFilledAmt:    sdkmath.NewInt(8789),
				Splits: []*OrderSplit{
					{Assets: coin(3, "apple"), Price: coin(700, "plum")},
					{Assets: coin(8, "apple"), Price: coin(89, "plum")},
				},
			},
			expErr: "splits asset total \"11apple\" does not equal filled assets \"10apple\" on ask order 105",
		},
		{
			name: "multiple price denoms in splits",
			receiver: OrderFulfillment{
				Order: NewOrder(106).WithBid(&BidOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98000),
				PriceAppliedAmt:   sdkmath.NewInt(789),
				PriceUnfilledAmt:  sdkmath.NewInt(90000),
				PriceFilledAmt:    sdkmath.NewInt(8789),
				Splits: []*OrderSplit{
					{Assets: coin(3, "apple"), Price: coin(700, "potato")},
					{Assets: coin(7, "apple"), Price: coin(89, "plum")},
				},
			},
			expErr: "multiple price denoms \"89plum,700potato\" in splits applied to bid order 106 having price \"98789plum\"",
		},
		{
			name: "wrong splits price denom",
			receiver: OrderFulfillment{
				Order: NewOrder(107).WithAsk(&AskOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98000),
				PriceAppliedAmt:   sdkmath.NewInt(789),
				PriceUnfilledAmt:  sdkmath.NewInt(90000),
				PriceFilledAmt:    sdkmath.NewInt(8789),
				Splits: []*OrderSplit{
					{Assets: coin(3, "apple"), Price: coin(700, "potato")},
					{Assets: coin(7, "apple"), Price: coin(89, "potato")},
				},
			},
			expErr: "splits price denom \"789potato\" does not equal order price denom \"98789plum\" on ask order 107",
		},
		{
			name: "splits price total too low",
			receiver: OrderFulfillment{
				Order: NewOrder(108).WithBid(&BidOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98000),
				PriceAppliedAmt:   sdkmath.NewInt(789),
				PriceUnfilledAmt:  sdkmath.NewInt(90000),
				PriceFilledAmt:    sdkmath.NewInt(8789),
				Splits: []*OrderSplit{
					{Assets: coin(3, "apple"), Price: coin(700, "plum")},
					{Assets: coin(7, "apple"), Price: coin(88, "plum")},
				},
			},
			expErr: "splits price total \"788plum\" does not equal filled price \"789plum\" on bid order 108",
		},
		{
			name: "splits price total too high",
			receiver: OrderFulfillment{
				Order: NewOrder(109).WithAsk(&AskOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98000),
				PriceAppliedAmt:   sdkmath.NewInt(789),
				PriceUnfilledAmt:  sdkmath.NewInt(90000),
				PriceFilledAmt:    sdkmath.NewInt(8789),
				Splits: []*OrderSplit{
					{Assets: coin(3, "apple"), Price: coin(701, "plum")},
					{Assets: coin(7, "apple"), Price: coin(89, "plum")},
				},
			},
			expErr: "splits price total \"790plum\" does not equal filled price \"789plum\" on ask order 109",
		},

		{
			name: "order fees left has negative",
			receiver: OrderFulfillment{
				Order: NewOrder(201).WithAsk(&AskOrder{
					Assets:                  coin(55, "apple"),
					Price:                   coin(98789, "plum"),
					SellerSettlementFlatFee: coinP(5, "fig"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98000),
				PriceAppliedAmt:   sdkmath.NewInt(789),
				PriceUnfilledAmt:  sdkmath.NewInt(90000),
				PriceFilledAmt:    sdkmath.NewInt(8789),
				Splits: []*OrderSplit{
					{Assets: coin(3, "apple"), Price: coin(700, "plum")},
					{Assets: coin(7, "apple"), Price: coin(89, "plum")},
				},
				OrderFeesLeft: sdk.Coins{coin(2, "fig"), coin(-3, "grape"), coin(4, "honeydew")},
			},
			expErr: "settlement fees left \"2fig,-3grape,4honeydew\" is negative for ask order 201 having fees \"5fig\"",
		},
		{
			name: "more fees left than in ask order",
			receiver: OrderFulfillment{
				Order: NewOrder(202).WithAsk(&AskOrder{
					Assets:                  coin(55, "apple"),
					Price:                   coin(98789, "plum"),
					SellerSettlementFlatFee: coinP(5, "fig"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98000),
				PriceAppliedAmt:   sdkmath.NewInt(789),
				PriceUnfilledAmt:  sdkmath.NewInt(90000),
				PriceFilledAmt:    sdkmath.NewInt(8789),
				Splits: []*OrderSplit{
					{Assets: coin(3, "apple"), Price: coin(700, "plum")},
					{Assets: coin(7, "apple"), Price: coin(89, "plum")},
				},
				OrderFeesLeft: coins(6, "fig"),
			},
			expErr: "settlement fees left \"6fig\" is greater than ask order 202 settlement fees \"5fig\"",
		},
		{
			name: "fees left in ask order without fees",
			receiver: OrderFulfillment{
				Order: NewOrder(203).WithAsk(&AskOrder{
					Assets:                  coin(55, "apple"),
					Price:                   coin(98789, "plum"),
					SellerSettlementFlatFee: nil,
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98000),
				PriceAppliedAmt:   sdkmath.NewInt(789),
				PriceUnfilledAmt:  sdkmath.NewInt(90000),
				PriceFilledAmt:    sdkmath.NewInt(8789),
				Splits: []*OrderSplit{
					{Assets: coin(3, "apple"), Price: coin(700, "plum")},
					{Assets: coin(7, "apple"), Price: coin(89, "plum")},
				},
				OrderFeesLeft: coins(1, "fig"),
			},
			expErr: "settlement fees left \"1fig\" is greater than ask order 203 settlement fees \"\"",
		},
		{
			name: "more fees left than in bid order",
			receiver: OrderFulfillment{
				Order: NewOrder(204).WithBid(&BidOrder{
					Assets:              coin(55, "apple"),
					Price:               coin(98789, "plum"),
					BuyerSettlementFees: sdk.Coins{coin(5, "fig"), coin(6, "grape")},
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98000),
				PriceAppliedAmt:   sdkmath.NewInt(789),
				PriceUnfilledAmt:  sdkmath.NewInt(90000),
				PriceFilledAmt:    sdkmath.NewInt(8789),
				Splits: []*OrderSplit{
					{Assets: coin(3, "apple"), Price: coin(700, "plum")},
					{Assets: coin(7, "apple"), Price: coin(89, "plum")},
				},
				OrderFeesLeft: coins(6, "fig"),
			},
			expErr: "settlement fees left \"6fig\" is greater than bid order 204 settlement fees \"5fig,6grape\"",
		},
		{
			name: "fees left in bid order without fees",
			receiver: OrderFulfillment{
				Order: NewOrder(205).WithBid(&BidOrder{
					Assets:              coin(55, "apple"),
					Price:               coin(98789, "plum"),
					BuyerSettlementFees: nil,
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(45),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				PriceLeftAmt:      sdkmath.NewInt(98000),
				PriceAppliedAmt:   sdkmath.NewInt(789),
				PriceUnfilledAmt:  sdkmath.NewInt(90000),
				PriceFilledAmt:    sdkmath.NewInt(8789),
				Splits: []*OrderSplit{
					{Assets: coin(3, "apple"), Price: coin(700, "plum")},
					{Assets: coin(7, "apple"), Price: coin(89, "plum")},
				},
				OrderFeesLeft: coins(1, "fig"),
			},
			expErr: "settlement fees left \"1fig\" is greater than bid order 205 settlement fees \"\"",
		},

		{
			name: "fully filled, price unfilled positive",
			receiver: OrderFulfillment{
				Order: NewOrder(250).WithAsk(&AskOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(0),
				AssetsFilledAmt:   sdkmath.NewInt(55),
				PriceLeftAmt:      sdkmath.NewInt(789),
				PriceAppliedAmt:   sdkmath.NewInt(98000),
				PriceUnfilledAmt:  sdkmath.NewInt(788),
				PriceFilledAmt:    sdkmath.NewInt(98001),
				Splits:            []*OrderSplit{{Assets: coin(55, "apple"), Price: coin(98000, "plum")}},
			},
			expErr: "fully filled ask order 250 has non-zero unfilled price \"788plum\"",
		},
		{
			name: "fully filled, order fees left positive",
			receiver: OrderFulfillment{
				Order: NewOrder(252).WithAsk(&AskOrder{
					Assets:                  coin(55, "apple"),
					Price:                   coin(98789, "plum"),
					SellerSettlementFlatFee: coinP(5, "fig"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(0),
				AssetsFilledAmt:   sdkmath.NewInt(55),
				PriceLeftAmt:      sdkmath.NewInt(0),
				PriceAppliedAmt:   sdkmath.NewInt(98789),
				PriceUnfilledAmt:  sdkmath.NewInt(0),
				PriceFilledAmt:    sdkmath.NewInt(98789),
				Splits:            []*OrderSplit{{Assets: coin(55, "apple"), Price: coin(98789, "plum")}},
				OrderFeesLeft:     coins(1, "fig"),
			},
			expErr: "fully filled ask order 252 has non-zero settlement fees left \"1fig\"",
		},

		{
			name: "ask order, price applied less than filled",
			receiver: OrderFulfillment{
				Order: NewOrder(301).WithAsk(&AskOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(10),
				AssetsFilledAmt:   sdkmath.NewInt(45),
				PriceLeftAmt:      sdkmath.NewInt(8789),
				PriceAppliedAmt:   sdkmath.NewInt(90000),
				PriceUnfilledAmt:  sdkmath.NewInt(789),
				PriceFilledAmt:    sdkmath.NewInt(98000),
				Splits:            []*OrderSplit{{Assets: coin(45, "apple"), Price: coin(90000, "plum")}},
			},
			expErr: "ask order 301 having assets \"55apple\" and price \"98789plum\" cannot be filled by \"45apple\" at price \"90000plum\": insufficient price",
		},
		{
			name: "ask order, partial, multiple fees left denoms",
			receiver: OrderFulfillment{
				Order: NewOrder(302).WithAsk(&AskOrder{
					Assets:                  coin(55, "apple"),
					Price:                   coin(98789, "plum"),
					SellerSettlementFlatFee: coinP(3, "fig"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(10),
				AssetsFilledAmt:   sdkmath.NewInt(45),
				PriceLeftAmt:      sdkmath.NewInt(789),
				PriceAppliedAmt:   sdkmath.NewInt(98000),
				PriceUnfilledAmt:  sdkmath.NewInt(8789),
				PriceFilledAmt:    sdkmath.NewInt(90000),
				Splits:            []*OrderSplit{{Assets: coin(45, "apple"), Price: coin(98000, "plum")}},
				OrderFeesLeft:     sdk.Coins{coin(1, "fig"), coin(0, "grape")},
			},
			expErr: "partial fulfillment for ask order 302 having seller settlement fees \"3fig\" has multiple denoms in fees left \"1fig,0grape\"",
		},
		{
			name: "ask order, tracked fees less than order fees",
			receiver: OrderFulfillment{
				Order: NewOrder(303).WithAsk(&AskOrder{
					Assets:                  coin(55, "apple"),
					Price:                   coin(98789, "plum"),
					SellerSettlementFlatFee: coinP(123, "fig"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(10),
				AssetsFilledAmt:   sdkmath.NewInt(45),
				PriceLeftAmt:      sdkmath.NewInt(789),
				PriceAppliedAmt:   sdkmath.NewInt(98000),
				PriceUnfilledAmt:  sdkmath.NewInt(8789),
				PriceFilledAmt:    sdkmath.NewInt(90000),
				Splits:            []*OrderSplit{{Assets: coin(45, "apple"), Price: coin(98000, "plum")}},
				OrderFeesLeft:     coins(22, "fig"),
				FeesToPay:         coins(100, "fig"),
			},
			expErr: "tracked settlement fees \"122fig\" is less than ask order 303 settlement fees \"123fig\"",
		},

		{
			name: "bid order, price applied less than price filled",
			receiver: OrderFulfillment{
				Order: NewOrder(275).WithBid(&BidOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(10),
				AssetsFilledAmt:   sdkmath.NewInt(45),
				PriceLeftAmt:      sdkmath.NewInt(790),
				PriceAppliedAmt:   sdkmath.NewInt(97999),
				PriceUnfilledAmt:  sdkmath.NewInt(789),
				PriceFilledAmt:    sdkmath.NewInt(98000),
				Splits:            []*OrderSplit{{Assets: coin(45, "apple"), Price: coin(97999, "plum")}},
			},
			expErr: "price applied \"97999plum\" does not equal price filled \"98000plum\" for bid order 275 having price \"98789plum\"",
		},
		{
			name: "bid order, price applied more than price filled",
			receiver: OrderFulfillment{
				Order: NewOrder(276).WithBid(&BidOrder{
					Assets: coin(55, "apple"),
					Price:  coin(98789, "plum"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(10),
				AssetsFilledAmt:   sdkmath.NewInt(45),
				PriceLeftAmt:      sdkmath.NewInt(788),
				PriceAppliedAmt:   sdkmath.NewInt(98001),
				PriceUnfilledAmt:  sdkmath.NewInt(789),
				PriceFilledAmt:    sdkmath.NewInt(98000),
				Splits:            []*OrderSplit{{Assets: coin(45, "apple"), Price: coin(98001, "plum")}},
			},
			expErr: "price applied \"98001plum\" does not equal price filled \"98000plum\" for bid order 276 having price \"98789plum\"",
		},
		{
			name: "bid order, tracked fees less than order fees",
			receiver: OrderFulfillment{
				Order: NewOrder(277).WithBid(&BidOrder{
					Assets:              coin(55, "apple"),
					Price:               coin(98789, "plum"),
					BuyerSettlementFees: coins(123, "fig"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(1),
				AssetsFilledAmt:   sdkmath.NewInt(54),
				PriceLeftAmt:      sdkmath.NewInt(89),
				PriceAppliedAmt:   sdkmath.NewInt(98700),
				PriceUnfilledAmt:  sdkmath.NewInt(89),
				PriceFilledAmt:    sdkmath.NewInt(98700),
				Splits:            []*OrderSplit{{Assets: coin(54, "apple"), Price: coin(98700, "plum")}},
				OrderFeesLeft:     coins(2, "fig"),
				FeesToPay:         coins(120, "fig"),
			},
			expErr: "tracked settlement fees \"122fig\" does not equal bid order 277 settlement fees \"123fig\"",
		},
		{
			name: "bid order, tracked fees more than order fees",
			receiver: OrderFulfillment{
				Order: NewOrder(277).WithBid(&BidOrder{
					Assets:              coin(55, "apple"),
					Price:               coin(98789, "plum"),
					BuyerSettlementFees: coins(123, "fig"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(1),
				AssetsFilledAmt:   sdkmath.NewInt(54),
				PriceLeftAmt:      sdkmath.NewInt(89),
				PriceAppliedAmt:   sdkmath.NewInt(98700),
				PriceUnfilledAmt:  sdkmath.NewInt(89),
				PriceFilledAmt:    sdkmath.NewInt(98700),
				Splits:            []*OrderSplit{{Assets: coin(54, "apple"), Price: coin(98700, "plum")}},
				OrderFeesLeft:     coins(4, "fig"),
				FeesToPay:         coins(120, "fig"),
			},
			expErr: "tracked settlement fees \"124fig\" does not equal bid order 277 settlement fees \"123fig\"",
		},

		{
			name: "partial ask, but not allowed",
			receiver: OrderFulfillment{
				Order: NewOrder(301).WithAsk(&AskOrder{
					Assets:       coin(55, "apple"),
					Price:        coin(98789, "plum"),
					AllowPartial: false,
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(1),
				AssetsFilledAmt:   sdkmath.NewInt(54),
				PriceLeftAmt:      sdkmath.NewInt(89),
				PriceAppliedAmt:   sdkmath.NewInt(98700),
				PriceUnfilledAmt:  sdkmath.NewInt(89),
				PriceFilledAmt:    sdkmath.NewInt(98700),
				Splits:            []*OrderSplit{{Assets: coin(54, "apple"), Price: coin(98700, "plum")}},
			},
			expErr: "cannot fill ask order 301 having assets \"55apple\" with \"54apple\": order does not allow partial fill",
		},
		{
			name: "partial bid, but not allowed",
			receiver: OrderFulfillment{
				Order: NewOrder(302).WithBid(&BidOrder{
					Assets:       coin(55, "apple"),
					Price:        coin(98789, "plum"),
					AllowPartial: false,
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(1),
				AssetsFilledAmt:   sdkmath.NewInt(54),
				PriceLeftAmt:      sdkmath.NewInt(89),
				PriceAppliedAmt:   sdkmath.NewInt(98700),
				PriceUnfilledAmt:  sdkmath.NewInt(89),
				PriceFilledAmt:    sdkmath.NewInt(98700),
				Splits:            []*OrderSplit{{Assets: coin(54, "apple"), Price: coin(98700, "plum")}},
			},
			expErr: "cannot fill bid order 302 having assets \"55apple\" with \"54apple\": order does not allow partial fill",
		},

		{
			name: "ask, fully filled, exact",
			receiver: OrderFulfillment{
				Order: NewOrder(501).WithAsk(&AskOrder{
					Assets:       coin(50, "apple"),
					Price:        coin(100, "plum"),
					AllowPartial: true,
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(0),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				PriceLeftAmt:      sdkmath.NewInt(0),
				PriceAppliedAmt:   sdkmath.NewInt(100),
				PriceUnfilledAmt:  sdkmath.NewInt(0),
				PriceFilledAmt:    sdkmath.NewInt(100),
				Splits:            []*OrderSplit{{Assets: coin(50, "apple"), Price: coin(100, "plum")}},
			},
			expErr: "",
		},
		{
			name: "ask, fully filled, extra price",
			receiver: OrderFulfillment{
				Order: NewOrder(502).WithAsk(&AskOrder{
					Assets:       coin(50, "apple"),
					Price:        coin(100, "plum"),
					AllowPartial: true,
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(0),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				PriceLeftAmt:      sdkmath.NewInt(-5),
				PriceAppliedAmt:   sdkmath.NewInt(105),
				PriceUnfilledAmt:  sdkmath.NewInt(0),
				PriceFilledAmt:    sdkmath.NewInt(100),
				Splits:            []*OrderSplit{{Assets: coin(50, "apple"), Price: coin(105, "plum")}},
			},
			expErr: "",
		},
		{
			name: "ask, partially filled, exact",
			receiver: OrderFulfillment{
				Order: NewOrder(503).WithAsk(&AskOrder{
					Assets:       coin(50, "apple"),
					Price:        coin(100, "plum"),
					AllowPartial: true,
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(10),
				AssetsFilledAmt:   sdkmath.NewInt(40),
				PriceLeftAmt:      sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.NewInt(80),
				PriceUnfilledAmt:  sdkmath.NewInt(20),
				PriceFilledAmt:    sdkmath.NewInt(80),
				Splits:            []*OrderSplit{{Assets: coin(40, "apple"), Price: coin(80, "plum")}},
			},
			expErr: "",
		},
		{
			name: "ask, partially filled, extra price",
			receiver: OrderFulfillment{
				Order: NewOrder(504).WithAsk(&AskOrder{
					Assets:       coin(50, "apple"),
					Price:        coin(100, "plum"),
					AllowPartial: true,
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(10),
				AssetsFilledAmt:   sdkmath.NewInt(40),
				PriceLeftAmt:      sdkmath.NewInt(10),
				PriceAppliedAmt:   sdkmath.NewInt(90),
				PriceUnfilledAmt:  sdkmath.NewInt(20),
				PriceFilledAmt:    sdkmath.NewInt(80),
				Splits:            []*OrderSplit{{Assets: coin(40, "apple"), Price: coin(90, "plum")}},
			},
			expErr: "",
		},
		{
			name: "bid, fully filled",
			receiver: OrderFulfillment{
				Order: NewOrder(505).WithBid(&BidOrder{
					Assets:       coin(50, "apple"),
					Price:        coin(100, "plum"),
					AllowPartial: true,
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(0),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				PriceLeftAmt:      sdkmath.NewInt(0),
				PriceAppliedAmt:   sdkmath.NewInt(100),
				PriceUnfilledAmt:  sdkmath.NewInt(0),
				PriceFilledAmt:    sdkmath.NewInt(100),
				Splits:            []*OrderSplit{{Assets: coin(50, "apple"), Price: coin(100, "plum")}},
			},
			expErr: "",
		},
		{
			name: "bid, partially filled",
			receiver: OrderFulfillment{
				Order: NewOrder(506).WithBid(&BidOrder{
					Assets:       coin(50, "apple"),
					Price:        coin(100, "plum"),
					AllowPartial: true,
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(10),
				AssetsFilledAmt:   sdkmath.NewInt(40),
				PriceLeftAmt:      sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.NewInt(80),
				PriceUnfilledAmt:  sdkmath.NewInt(20),
				PriceFilledAmt:    sdkmath.NewInt(80),
				Splits:            []*OrderSplit{{Assets: coin(40, "apple"), Price: coin(80, "plum")}},
			},
			expErr: "",
		},
		{
			name: "ask, full, no fees, some to pay",
			receiver: OrderFulfillment{
				Order: NewOrder(507).WithAsk(&AskOrder{
					Assets:                  coin(50, "apple"),
					Price:                   coin(100, "plum"),
					AllowPartial:            true,
					SellerSettlementFlatFee: nil,
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(0),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				PriceLeftAmt:      sdkmath.NewInt(0),
				PriceAppliedAmt:   sdkmath.NewInt(100),
				PriceUnfilledAmt:  sdkmath.NewInt(0),
				PriceFilledAmt:    sdkmath.NewInt(100),
				Splits:            []*OrderSplit{{Assets: coin(50, "apple"), Price: coin(100, "plum")}},
				OrderFeesLeft:     nil,
				FeesToPay:         coins(20, "fig"),
			},
			expErr: "",
		},
		{
			name: "ask, full, with fees, paying exact",
			receiver: OrderFulfillment{
				Order: NewOrder(508).WithAsk(&AskOrder{
					Assets:                  coin(50, "apple"),
					Price:                   coin(100, "plum"),
					AllowPartial:            true,
					SellerSettlementFlatFee: coinP(200, "fig"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(0),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				PriceLeftAmt:      sdkmath.NewInt(0),
				PriceAppliedAmt:   sdkmath.NewInt(100),
				PriceUnfilledAmt:  sdkmath.NewInt(0),
				PriceFilledAmt:    sdkmath.NewInt(100),
				Splits:            []*OrderSplit{{Assets: coin(50, "apple"), Price: coin(100, "plum")}},
				OrderFeesLeft:     nil,
				FeesToPay:         coins(200, "fig"),
			},
			expErr: "",
		},
		{
			name: "ask, full, with fees, paying more",
			receiver: OrderFulfillment{
				Order: NewOrder(509).WithAsk(&AskOrder{
					Assets:                  coin(50, "apple"),
					Price:                   coin(100, "plum"),
					AllowPartial:            true,
					SellerSettlementFlatFee: coinP(200, "fig"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(0),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				PriceLeftAmt:      sdkmath.NewInt(0),
				PriceAppliedAmt:   sdkmath.NewInt(100),
				PriceUnfilledAmt:  sdkmath.NewInt(0),
				PriceFilledAmt:    sdkmath.NewInt(100),
				Splits:            []*OrderSplit{{Assets: coin(50, "apple"), Price: coin(100, "plum")}},
				OrderFeesLeft:     nil,
				FeesToPay:         coins(205, "fig"),
			},
			expErr: "",
		},
		{
			name: "ask, partial, no fees, some to pay",
			receiver: OrderFulfillment{
				Order: NewOrder(510).WithAsk(&AskOrder{
					Assets:                  coin(50, "apple"),
					Price:                   coin(100, "plum"),
					AllowPartial:            true,
					SellerSettlementFlatFee: nil,
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(10),
				AssetsFilledAmt:   sdkmath.NewInt(40),
				PriceLeftAmt:      sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.NewInt(80),
				PriceUnfilledAmt:  sdkmath.NewInt(20),
				PriceFilledAmt:    sdkmath.NewInt(80),
				Splits:            []*OrderSplit{{Assets: coin(40, "apple"), Price: coin(80, "plum")}},
				OrderFeesLeft:     nil,
				FeesToPay:         coins(20, "fig"),
			},
			expErr: "",
		},
		{
			name: "ask, partial, with fees, paying exact",
			receiver: OrderFulfillment{
				Order: NewOrder(511).WithAsk(&AskOrder{
					Assets:                  coin(50, "apple"),
					Price:                   coin(100, "plum"),
					AllowPartial:            true,
					SellerSettlementFlatFee: coinP(20, "fig"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(10),
				AssetsFilledAmt:   sdkmath.NewInt(40),
				PriceLeftAmt:      sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.NewInt(80),
				PriceUnfilledAmt:  sdkmath.NewInt(20),
				PriceFilledAmt:    sdkmath.NewInt(80),
				Splits:            []*OrderSplit{{Assets: coin(40, "apple"), Price: coin(80, "plum")}},
				OrderFeesLeft:     nil,
				FeesToPay:         coins(20, "fig"),
			},
			expErr: "",
		},
		{
			name: "ask, partial, with fees, paying more",
			receiver: OrderFulfillment{
				Order: NewOrder(512).WithAsk(&AskOrder{
					Assets:                  coin(50, "apple"),
					Price:                   coin(100, "plum"),
					AllowPartial:            true,
					SellerSettlementFlatFee: coinP(20, "fig"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(10),
				AssetsFilledAmt:   sdkmath.NewInt(40),
				PriceLeftAmt:      sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.NewInt(80),
				PriceUnfilledAmt:  sdkmath.NewInt(20),
				PriceFilledAmt:    sdkmath.NewInt(80),
				Splits:            []*OrderSplit{{Assets: coin(40, "apple"), Price: coin(80, "plum")}},
				OrderFeesLeft:     coins(4, "fig"),
				FeesToPay:         coins(55, "fig"),
			},
			expErr: "",
		},
		{
			name: "ask, partial, with fees, none being paid",
			receiver: OrderFulfillment{
				Order: NewOrder(513).WithAsk(&AskOrder{
					Assets:                  coin(50, "apple"),
					Price:                   coin(100, "plum"),
					AllowPartial:            true,
					SellerSettlementFlatFee: coinP(200, "fig"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(10),
				AssetsFilledAmt:   sdkmath.NewInt(40),
				PriceLeftAmt:      sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.NewInt(80),
				PriceUnfilledAmt:  sdkmath.NewInt(20),
				PriceFilledAmt:    sdkmath.NewInt(80),
				Splits:            []*OrderSplit{{Assets: coin(40, "apple"), Price: coin(80, "plum")}},
				OrderFeesLeft:     coins(200, "fig"),
				FeesToPay:         nil,
			},
			expErr: "",
		},
		{
			name: "bid, partial, with fees, none being paid",
			receiver: OrderFulfillment{
				Order: NewOrder(514).WithBid(&BidOrder{
					Assets:              coin(50, "apple"),
					Price:               coin(100, "plum"),
					AllowPartial:        true,
					BuyerSettlementFees: coins(20, "fig"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(10),
				AssetsFilledAmt:   sdkmath.NewInt(40),
				PriceLeftAmt:      sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.NewInt(80),
				PriceUnfilledAmt:  sdkmath.NewInt(20),
				PriceFilledAmt:    sdkmath.NewInt(80),
				Splits:            []*OrderSplit{{Assets: coin(40, "apple"), Price: coin(80, "plum")}},
				OrderFeesLeft:     coins(20, "fig"),
				FeesToPay:         nil,
			},
			expErr: "",
		},
		{
			name: "bid, partial, with fees, all being paid",
			receiver: OrderFulfillment{
				Order: NewOrder(515).WithBid(&BidOrder{
					Assets:              coin(50, "apple"),
					Price:               coin(100, "plum"),
					AllowPartial:        true,
					BuyerSettlementFees: coins(20, "fig"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(10),
				AssetsFilledAmt:   sdkmath.NewInt(40),
				PriceLeftAmt:      sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.NewInt(80),
				PriceUnfilledAmt:  sdkmath.NewInt(20),
				PriceFilledAmt:    sdkmath.NewInt(80),
				Splits:            []*OrderSplit{{Assets: coin(40, "apple"), Price: coin(80, "plum")}},
				OrderFeesLeft:     nil,
				FeesToPay:         coins(20, "fig"),
			},
			expErr: "",
		},
		{
			name: "bid, partial, with fees, some being paid",
			receiver: OrderFulfillment{
				Order: NewOrder(515).WithBid(&BidOrder{
					Assets:              coin(50, "apple"),
					Price:               coin(100, "plum"),
					AllowPartial:        true,
					BuyerSettlementFees: coins(20, "fig"),
				}),
				IsFinalized:       true,
				AssetsUnfilledAmt: sdkmath.NewInt(10),
				AssetsFilledAmt:   sdkmath.NewInt(40),
				PriceLeftAmt:      sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.NewInt(80),
				PriceUnfilledAmt:  sdkmath.NewInt(20),
				PriceFilledAmt:    sdkmath.NewInt(80),
				Splits:            []*OrderSplit{{Assets: coin(40, "apple"), Price: coin(80, "plum")}},
				OrderFeesLeft:     coins(4, "fig"),
				FeesToPay:         coins(16, "fig"),
			},
			expErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.receiver.Validate2()
			}
			require.NotPanics(t, testFunc, "Validate2")
			assertions.AssertErrorValue(t, err, tc.expErr, "Validate2 error")
		})
	}
}

func TestFulfill(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}

	tests := []struct {
		name    string
		ofA     *OrderFulfillment
		ofB     *OrderFulfillment
		expA    *OrderFulfillment
		expB    *OrderFulfillment
		expErr  string
		swapErr string
	}{
		{
			name:    "ask ask",
			ofA:     &OrderFulfillment{Order: NewOrder(1).WithAsk(&AskOrder{})},
			ofB:     &OrderFulfillment{Order: NewOrder(2).WithAsk(&AskOrder{})},
			expErr:  "cannot fulfill ask order 1 with ask order 2: order type mismatch",
			swapErr: "cannot fulfill ask order 2 with ask order 1: order type mismatch",
		},
		{
			name:    "bid bid",
			ofA:     &OrderFulfillment{Order: NewOrder(4).WithBid(&BidOrder{})},
			ofB:     &OrderFulfillment{Order: NewOrder(3).WithBid(&BidOrder{})},
			expErr:  "cannot fulfill bid order 4 with bid order 3: order type mismatch",
			swapErr: "cannot fulfill bid order 3 with bid order 4: order type mismatch",
		},
		{
			name:   "diff asset denom",
			ofA:    &OrderFulfillment{Order: NewOrder(5).WithAsk(&AskOrder{Assets: coin(15, "apple")})},
			ofB:    &OrderFulfillment{Order: NewOrder(6).WithBid(&BidOrder{Assets: coin(16, "banana")})},
			expErr: "cannot fill bid order 6 having assets \"16banana\" with ask order 5 having assets \"15apple\": denom mismatch",
		},
		{
			name: "diff price denom",
			ofA: &OrderFulfillment{
				Order: NewOrder(7).WithAsk(&AskOrder{
					Assets: coin(15, "apple"),
					Price:  coin(17, "pear"),
				}),
			},
			ofB: &OrderFulfillment{
				Order: NewOrder(8).WithBid(&BidOrder{
					Assets: coin(16, "apple"),
					Price:  coin(18, "plum"),
				}),
			},
			expErr: "cannot fill ask order 7 having price \"17pear\" with bid order 8 having price \"18plum\": denom mismatch",
		},
		{
			name: "cannot get assets left",
			ofA: &OrderFulfillment{
				Order: NewOrder(9).WithAsk(&AskOrder{
					Assets: coin(15, "apple"),
					Price:  coin(16, "plum"),
				}),
				AssetsUnfilledAmt: sdkmath.NewInt(-1),
			},
			ofB: &OrderFulfillment{
				Order: NewOrder(10).WithBid(&BidOrder{
					Assets: coin(17, "apple"),
					Price:  coin(18, "plum"),
				}),
				AssetsUnfilledAmt: sdkmath.NewInt(3),
			},
			expErr: "cannot fill ask order 9 having assets left \"-1apple\" with bid order 10 " +
				"having assets left \"3apple\": zero or negative assets left",
		},
		{
			name: "error from apply",
			ofA: &OrderFulfillment{
				Order: NewOrder(11).WithAsk(&AskOrder{
					Assets: coin(15, "apple"),
					Price:  coin(90, "plum"),
				}),
				AssetsUnfilledAmt: sdkmath.NewInt(15),
				AssetsFilledAmt:   sdkmath.NewInt(0),
				PriceLeftAmt:      sdkmath.NewInt(90),
				PriceAppliedAmt:   sdkmath.NewInt(0),
			},
			ofB: &OrderFulfillment{
				Order: NewOrder(12).WithBid(&BidOrder{
					Assets: coin(30, "apple"),
					Price:  coin(180, "plum"),
				}),
				AssetsUnfilledAmt: sdkmath.NewInt(30),
				AssetsFilledAmt:   sdkmath.NewInt(0),
				PriceLeftAmt:      sdkmath.NewInt(89),
				PriceAppliedAmt:   sdkmath.NewInt(91),
			},
			expErr: "cannot fill bid order 12 having price left \"89plum\" to ask order 11 at a price of \"90plum\": overfill",
		},
		{
			name: "both filled in full",
			ofA: &OrderFulfillment{
				Order: NewOrder(101).WithAsk(&AskOrder{
					Assets: coin(33, "apple"),
					Price:  coin(57, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(0),
				AssetsUnfilledAmt: sdkmath.NewInt(33),
				PriceAppliedAmt:   sdkmath.NewInt(0),
				PriceLeftAmt:      sdkmath.NewInt(57),
			},
			ofB: &OrderFulfillment{
				Order: NewOrder(102).WithBid(&BidOrder{
					Assets: coin(33, "apple"),
					Price:  coin(57, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(0),
				AssetsUnfilledAmt: sdkmath.NewInt(33),
				PriceAppliedAmt:   sdkmath.NewInt(0),
				PriceLeftAmt:      sdkmath.NewInt(57),
			},
			expA: &OrderFulfillment{
				Order: NewOrder(101).WithAsk(&AskOrder{
					Assets: coin(33, "apple"),
					Price:  coin(57, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(33),
				AssetsUnfilledAmt: ZeroAmtAfterSub,
				PriceAppliedAmt:   sdkmath.NewInt(57),
				PriceLeftAmt:      ZeroAmtAfterSub,
				Splits:            []*OrderSplit{{Assets: coin(33, "apple"), Price: coin(57, "plum")}},
			},
			expB: &OrderFulfillment{
				Order: NewOrder(102).WithBid(&BidOrder{
					Assets: coin(33, "apple"),
					Price:  coin(57, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(33),
				AssetsUnfilledAmt: ZeroAmtAfterSub,
				PriceAppliedAmt:   sdkmath.NewInt(57),
				PriceLeftAmt:      ZeroAmtAfterSub,
				Splits:            []*OrderSplit{{Assets: coin(33, "apple"), Price: coin(57, "plum")}},
			},
		},
		{
			name: "ask, unfilled, gets partially filled",
			ofA: &OrderFulfillment{
				Order: NewOrder(103).WithAsk(&AskOrder{
					Assets: coin(50, "apple"),
					Price:  coin(100, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(0),
				AssetsUnfilledAmt: sdkmath.NewInt(50),
				PriceAppliedAmt:   sdkmath.NewInt(0),
				PriceLeftAmt:      sdkmath.NewInt(100),
			},
			ofB: &OrderFulfillment{
				Order: NewOrder(104).WithBid(&BidOrder{
					Assets: coin(20, "apple"),
					Price:  coin(80, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(0),
				AssetsUnfilledAmt: sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.NewInt(0),
				PriceLeftAmt:      sdkmath.NewInt(80),
			},
			expA: &OrderFulfillment{
				Order: NewOrder(103).WithAsk(&AskOrder{
					Assets: coin(50, "apple"),
					Price:  coin(100, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: sdkmath.NewInt(30),
				PriceAppliedAmt:   sdkmath.NewInt(80),
				PriceLeftAmt:      sdkmath.NewInt(20),
				Splits:            []*OrderSplit{{Assets: coin(20, "apple"), Price: coin(80, "plum")}},
			},
			expB: &OrderFulfillment{
				Order: NewOrder(104).WithBid(&BidOrder{
					Assets: coin(20, "apple"),
					Price:  coin(80, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: ZeroAmtAfterSub,
				PriceAppliedAmt:   sdkmath.NewInt(80),
				PriceLeftAmt:      ZeroAmtAfterSub,
				Splits:            []*OrderSplit{{Assets: coin(20, "apple"), Price: coin(80, "plum")}},
			},
		},
		{
			name: "ask, partially filled, gets partially filled more",
			ofA: &OrderFulfillment{
				Order: NewOrder(105).WithAsk(&AskOrder{
					Assets: coin(50, "apple"),
					Price:  coin(100, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(1),
				AssetsUnfilledAmt: sdkmath.NewInt(49),
				PriceAppliedAmt:   sdkmath.NewInt(2),
				PriceLeftAmt:      sdkmath.NewInt(98),
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: NewOrder(88).WithBid(&BidOrder{})},
						Assets: coin(1, "apple"),
						Price:  coin(2, "plum"),
					},
				},
			},
			ofB: &OrderFulfillment{
				Order: NewOrder(106).WithBid(&BidOrder{
					Assets: coin(20, "apple"),
					Price:  coin(80, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(0),
				AssetsUnfilledAmt: sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.NewInt(0),
				PriceLeftAmt:      sdkmath.NewInt(80),
			},
			expA: &OrderFulfillment{
				Order: NewOrder(105).WithAsk(&AskOrder{
					Assets: coin(50, "apple"),
					Price:  coin(100, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(21),
				AssetsUnfilledAmt: sdkmath.NewInt(29),
				PriceAppliedAmt:   sdkmath.NewInt(82),
				PriceLeftAmt:      sdkmath.NewInt(18),
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: NewOrder(88).WithBid(&BidOrder{})},
						Assets: coin(1, "apple"),
						Price:  coin(2, "plum"),
					},
					{Assets: coin(20, "apple"), Price: coin(80, "plum")},
				},
			},
			expB: &OrderFulfillment{
				Order: NewOrder(106).WithBid(&BidOrder{
					Assets: coin(20, "apple"),
					Price:  coin(80, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: ZeroAmtAfterSub,
				PriceAppliedAmt:   sdkmath.NewInt(80),
				PriceLeftAmt:      ZeroAmtAfterSub,
				Splits:            []*OrderSplit{{Assets: coin(20, "apple"), Price: coin(80, "plum")}},
			},
		},
		{
			name: "ask, partially filled, gets fully filled",
			ofA: &OrderFulfillment{
				Order: NewOrder(107).WithAsk(&AskOrder{
					Assets: coin(50, "apple"),
					Price:  coin(100, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(30),
				AssetsUnfilledAmt: sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.NewInt(20),
				PriceLeftAmt:      sdkmath.NewInt(80),
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: NewOrder(86).WithBid(&BidOrder{})},
						Assets: coin(30, "apple"),
						Price:  coin(20, "plum"),
					},
				},
			},
			ofB: &OrderFulfillment{
				Order: NewOrder(108).WithBid(&BidOrder{
					Assets: coin(20, "apple"),
					Price:  coin(80, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(0),
				AssetsUnfilledAmt: sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.NewInt(0),
				PriceLeftAmt:      sdkmath.NewInt(80),
			},
			expA: &OrderFulfillment{
				Order: NewOrder(107).WithAsk(&AskOrder{
					Assets: coin(50, "apple"),
					Price:  coin(100, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: ZeroAmtAfterSub,
				PriceAppliedAmt:   sdkmath.NewInt(100),
				PriceLeftAmt:      ZeroAmtAfterSub,
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: NewOrder(86).WithBid(&BidOrder{})},
						Assets: coin(30, "apple"),
						Price:  coin(20, "plum"),
					},
					{Assets: coin(20, "apple"), Price: coin(80, "plum")},
				},
			},
			expB: &OrderFulfillment{
				Order: NewOrder(108).WithBid(&BidOrder{
					Assets: coin(20, "apple"),
					Price:  coin(80, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: ZeroAmtAfterSub,
				PriceAppliedAmt:   sdkmath.NewInt(80),
				PriceLeftAmt:      ZeroAmtAfterSub,
				Splits:            []*OrderSplit{{Assets: coin(20, "apple"), Price: coin(80, "plum")}},
			},
		},
		{
			name: "bid, unfilled, gets partially filled",
			ofA: &OrderFulfillment{
				Order: NewOrder(151).WithAsk(&AskOrder{
					Assets: coin(20, "apple"),
					Price:  coin(30, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(0),
				AssetsUnfilledAmt: sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.NewInt(0),
				PriceLeftAmt:      sdkmath.NewInt(30),
			},
			ofB: &OrderFulfillment{
				Order: NewOrder(152).WithBid(&BidOrder{
					Assets: coin(50, "apple"),
					Price:  coin(100, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(0),
				AssetsUnfilledAmt: sdkmath.NewInt(50),
				PriceAppliedAmt:   sdkmath.NewInt(0),
				PriceLeftAmt:      sdkmath.NewInt(100),
			},
			expA: &OrderFulfillment{
				Order: NewOrder(151).WithAsk(&AskOrder{
					Assets: coin(20, "apple"),
					Price:  coin(30, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: ZeroAmtAfterSub,
				PriceAppliedAmt:   sdkmath.NewInt(40),
				PriceLeftAmt:      sdkmath.NewInt(-10),
				Splits:            []*OrderSplit{{Assets: coin(20, "apple"), Price: coin(40, "plum")}},
			},
			expB: &OrderFulfillment{
				Order: NewOrder(152).WithBid(&BidOrder{
					Assets: coin(50, "apple"),
					Price:  coin(100, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: sdkmath.NewInt(30),
				PriceAppliedAmt:   sdkmath.NewInt(40),
				PriceLeftAmt:      sdkmath.NewInt(60),
				Splits:            []*OrderSplit{{Assets: coin(20, "apple"), Price: coin(40, "plum")}},
			},
		},
		{
			name: "bid, unfilled, gets partially filled with truncation",
			ofA: &OrderFulfillment{
				Order: NewOrder(153).WithAsk(&AskOrder{
					Assets: coin(20, "apple"),
					Price:  coin(30, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(0),
				AssetsUnfilledAmt: sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.NewInt(0),
				PriceLeftAmt:      sdkmath.NewInt(30),
			},
			ofB: &OrderFulfillment{
				Order: NewOrder(154).WithBid(&BidOrder{
					Assets: coin(57, "apple"),
					Price:  coin(331, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(0),
				AssetsUnfilledAmt: sdkmath.NewInt(57),
				PriceAppliedAmt:   sdkmath.NewInt(0),
				PriceLeftAmt:      sdkmath.NewInt(331),
			},
			expA: &OrderFulfillment{
				Order: NewOrder(153).WithAsk(&AskOrder{
					Assets: coin(20, "apple"),
					Price:  coin(30, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: ZeroAmtAfterSub,
				PriceAppliedAmt:   sdkmath.NewInt(116), // 331 * 20 / 57 = 116.140350877193
				PriceLeftAmt:      sdkmath.NewInt(-86),
				Splits:            []*OrderSplit{{Assets: coin(20, "apple"), Price: coin(116, "plum")}},
			},
			expB: &OrderFulfillment{
				Order: NewOrder(154).WithBid(&BidOrder{
					Assets: coin(57, "apple"),
					Price:  coin(331, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: sdkmath.NewInt(37),
				PriceAppliedAmt:   sdkmath.NewInt(116),
				PriceLeftAmt:      sdkmath.NewInt(215),
				Splits:            []*OrderSplit{{Assets: coin(20, "apple"), Price: coin(116, "plum")}},
			},
		},
		{
			name: "bid, partially filled, gets partially filled more",
			ofA: &OrderFulfillment{
				Order: NewOrder(155).WithAsk(&AskOrder{
					Assets: coin(20, "apple"),
					Price:  coin(30, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(0),
				AssetsUnfilledAmt: sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.NewInt(0),
				PriceLeftAmt:      sdkmath.NewInt(30),
			},
			ofB: &OrderFulfillment{
				Order: NewOrder(156).WithBid(&BidOrder{
					Assets: coin(50, "apple"),
					Price:  coin(100, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(10),
				AssetsUnfilledAmt: sdkmath.NewInt(40),
				PriceAppliedAmt:   sdkmath.NewInt(20),
				PriceLeftAmt:      sdkmath.NewInt(80),
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: NewOrder(77).WithAsk(&AskOrder{})},
						Assets: coin(10, "apple"),
						Price:  coin(20, "plum"),
					},
				},
			},
			expA: &OrderFulfillment{
				Order: NewOrder(155).WithAsk(&AskOrder{
					Assets: coin(20, "apple"),
					Price:  coin(30, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: ZeroAmtAfterSub,
				PriceAppliedAmt:   sdkmath.NewInt(40),
				PriceLeftAmt:      sdkmath.NewInt(-10),
				Splits:            []*OrderSplit{{Assets: coin(20, "apple"), Price: coin(40, "plum")}},
			},
			expB: &OrderFulfillment{
				Order: NewOrder(156).WithBid(&BidOrder{
					Assets: coin(50, "apple"),
					Price:  coin(100, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(30),
				AssetsUnfilledAmt: sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.NewInt(60),
				PriceLeftAmt:      sdkmath.NewInt(40),
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: NewOrder(77).WithAsk(&AskOrder{})},
						Assets: coin(10, "apple"),
						Price:  coin(20, "plum"),
					},
					{Assets: coin(20, "apple"), Price: coin(40, "plum")},
				},
			},
		},
		{
			name: "bid, partially filled, gets fully filled",
			ofA: &OrderFulfillment{
				Order: NewOrder(157).WithAsk(&AskOrder{
					Assets: coin(20, "apple"),
					Price:  coin(30, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(0),
				AssetsUnfilledAmt: sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.NewInt(0),
				PriceLeftAmt:      sdkmath.NewInt(30),
			},
			ofB: &OrderFulfillment{
				Order: NewOrder(158).WithBid(&BidOrder{
					Assets: coin(50, "apple"),
					Price:  coin(100, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(30),
				AssetsUnfilledAmt: sdkmath.NewInt(20),
				PriceAppliedAmt:   sdkmath.NewInt(60),
				PriceLeftAmt:      sdkmath.NewInt(40),
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: NewOrder(75).WithAsk(&AskOrder{})},
						Assets: coin(30, "apple"),
						Price:  coin(60, "plum"),
					},
				},
			},
			expA: &OrderFulfillment{
				Order: NewOrder(157).WithAsk(&AskOrder{
					Assets: coin(20, "apple"),
					Price:  coin(30, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: ZeroAmtAfterSub,
				PriceAppliedAmt:   sdkmath.NewInt(40),
				PriceLeftAmt:      sdkmath.NewInt(-10),
				Splits:            []*OrderSplit{{Assets: coin(20, "apple"), Price: coin(40, "plum")}},
			},
			expB: &OrderFulfillment{
				Order: NewOrder(158).WithBid(&BidOrder{
					Assets: coin(50, "apple"),
					Price:  coin(100, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: ZeroAmtAfterSub,
				PriceAppliedAmt:   sdkmath.NewInt(100),
				PriceLeftAmt:      ZeroAmtAfterSub,
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: NewOrder(75).WithAsk(&AskOrder{})},
						Assets: coin(30, "apple"),
						Price:  coin(60, "plum"),
					},
					{Assets: coin(20, "apple"), Price: coin(40, "plum")},
				},
			},
		},
		{
			name: "both partially filled, both get fully filled",
			ofA: &OrderFulfillment{
				Order: NewOrder(201).WithAsk(&AskOrder{
					Assets: coin(50, "apple"),
					Price:  coin(100, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: sdkmath.NewInt(30),
				PriceAppliedAmt:   sdkmath.NewInt(40),
				PriceLeftAmt:      sdkmath.NewInt(60),
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: NewOrder(1002).WithBid(&BidOrder{})},
						Assets: coin(20, "apple"),
						Price:  coin(40, "plum"),
					},
				},
			},
			ofB: &OrderFulfillment{
				Order: NewOrder(202).WithBid(&BidOrder{
					Assets: coin(50, "apple"),
					Price:  coin(100, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: sdkmath.NewInt(30),
				PriceAppliedAmt:   sdkmath.NewInt(40),
				PriceLeftAmt:      sdkmath.NewInt(60),
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: NewOrder(1003).WithAsk(&AskOrder{})},
						Assets: coin(20, "apple"),
						Price:  coin(40, "plum"),
					},
				},
			},
			expA: &OrderFulfillment{
				Order: NewOrder(201).WithAsk(&AskOrder{
					Assets: coin(50, "apple"),
					Price:  coin(100, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: ZeroAmtAfterSub,
				PriceAppliedAmt:   sdkmath.NewInt(100),
				PriceLeftAmt:      ZeroAmtAfterSub,
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: NewOrder(1002).WithBid(&BidOrder{})},
						Assets: coin(20, "apple"),
						Price:  coin(40, "plum"),
					},
					{Assets: coin(30, "apple"), Price: coin(60, "plum")},
				},
			},
			expB: &OrderFulfillment{
				Order: NewOrder(202).WithBid(&BidOrder{
					Assets: coin(50, "apple"),
					Price:  coin(100, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: ZeroAmtAfterSub,
				PriceAppliedAmt:   sdkmath.NewInt(100),
				PriceLeftAmt:      ZeroAmtAfterSub,
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: NewOrder(1003).WithAsk(&AskOrder{})},
						Assets: coin(20, "apple"),
						Price:  coin(40, "plum"),
					},
					{Assets: coin(30, "apple"), Price: coin(60, "plum")},
				},
			},
		},
		{
			name: "both partially filled, ask gets fully filled",
			ofA: &OrderFulfillment{
				Order: NewOrder(203).WithAsk(&AskOrder{
					Assets: coin(50, "apple"),
					Price:  coin(100, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: sdkmath.NewInt(30),
				PriceAppliedAmt:   sdkmath.NewInt(40),
				PriceLeftAmt:      sdkmath.NewInt(60),
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: NewOrder(1004).WithBid(&BidOrder{})},
						Assets: coin(20, "apple"),
						Price:  coin(40, "plum"),
					},
				},
			},
			ofB: &OrderFulfillment{
				Order: NewOrder(204).WithBid(&BidOrder{
					Assets: coin(60, "apple"),
					Price:  coin(120, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: sdkmath.NewInt(40),
				PriceAppliedAmt:   sdkmath.NewInt(40),
				PriceLeftAmt:      sdkmath.NewInt(80),
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: NewOrder(1005).WithAsk(&AskOrder{})},
						Assets: coin(20, "apple"),
						Price:  coin(40, "plum"),
					},
				},
			},
			expA: &OrderFulfillment{
				Order: NewOrder(203).WithAsk(&AskOrder{
					Assets: coin(50, "apple"),
					Price:  coin(100, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: ZeroAmtAfterSub,
				PriceAppliedAmt:   sdkmath.NewInt(100),
				PriceLeftAmt:      ZeroAmtAfterSub,
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: NewOrder(1004).WithBid(&BidOrder{})},
						Assets: coin(20, "apple"),
						Price:  coin(40, "plum"),
					},
					{Assets: coin(30, "apple"), Price: coin(60, "plum")},
				},
			},
			expB: &OrderFulfillment{
				Order: NewOrder(204).WithBid(&BidOrder{
					Assets: coin(60, "apple"),
					Price:  coin(120, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: sdkmath.NewInt(10),
				PriceAppliedAmt:   sdkmath.NewInt(100),
				PriceLeftAmt:      sdkmath.NewInt(20),
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: NewOrder(1005).WithAsk(&AskOrder{})},
						Assets: coin(20, "apple"),
						Price:  coin(40, "plum"),
					},
					{Assets: coin(30, "apple"), Price: coin(60, "plum")},
				},
			},
		},
		{
			name: "both partially filled, bid gets fully filled",
			ofA: &OrderFulfillment{
				Order: NewOrder(205).WithAsk(&AskOrder{
					Assets: coin(60, "apple"),
					Price:  coin(120, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: sdkmath.NewInt(40),
				PriceAppliedAmt:   sdkmath.NewInt(40),
				PriceLeftAmt:      sdkmath.NewInt(80),
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: NewOrder(1006).WithBid(&BidOrder{})},
						Assets: coin(20, "apple"),
						Price:  coin(40, "plum"),
					},
				},
			},
			ofB: &OrderFulfillment{
				Order: NewOrder(206).WithBid(&BidOrder{
					Assets: coin(50, "apple"),
					Price:  coin(100, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(20),
				AssetsUnfilledAmt: sdkmath.NewInt(30),
				PriceAppliedAmt:   sdkmath.NewInt(40),
				PriceLeftAmt:      sdkmath.NewInt(60),
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: NewOrder(1007).WithAsk(&AskOrder{})},
						Assets: coin(20, "apple"),
						Price:  coin(40, "plum"),
					},
				},
			},
			expA: &OrderFulfillment{
				Order: NewOrder(205).WithAsk(&AskOrder{
					Assets: coin(60, "apple"),
					Price:  coin(120, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: sdkmath.NewInt(10),
				PriceAppliedAmt:   sdkmath.NewInt(100),
				PriceLeftAmt:      sdkmath.NewInt(20),
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: NewOrder(1006).WithBid(&BidOrder{})},
						Assets: coin(20, "apple"),
						Price:  coin(40, "plum"),
					},
					{Assets: coin(30, "apple"), Price: coin(60, "plum")},
				},
			},
			expB: &OrderFulfillment{
				Order: NewOrder(206).WithBid(&BidOrder{
					Assets: coin(50, "apple"),
					Price:  coin(100, "plum"),
				}),
				AssetsFilledAmt:   sdkmath.NewInt(50),
				AssetsUnfilledAmt: ZeroAmtAfterSub,
				PriceAppliedAmt:   sdkmath.NewInt(100),
				PriceLeftAmt:      ZeroAmtAfterSub,
				Splits: []*OrderSplit{
					{
						Order:  &OrderFulfillment{Order: NewOrder(1007).WithAsk(&AskOrder{})},
						Assets: coin(20, "apple"),
						Price:  coin(40, "plum"),
					},
					{Assets: coin(30, "apple"), Price: coin(60, "plum")},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.expErr) != 0 && len(tc.swapErr) == 0 {
				tc.swapErr = tc.expErr
			}
			if len(tc.expErr) == 0 {
				for _, split := range tc.expA.Splits {
					if split.Order == nil {
						split.Order = tc.expB
					}
				}
				for _, split := range tc.expB.Splits {
					if split.Order == nil {
						split.Order = tc.expA
					}
				}
			}

			of1, of2 := copyOrderFulfillment(tc.ofA), copyOrderFulfillment(tc.ofB)
			var err error
			testFunc := func() {
				err = Fulfill(of1, of2)
			}
			require.NotPanics(t, testFunc, "Fulfill(A, B)")
			assertions.AssertErrorValue(t, err, tc.expErr, "Fulfill(A, B) error")
			if len(tc.expErr) == 0 {
				if !assertEqualOrderFulfillments(t, tc.expA, of1, "Fulfill(A, B): A") {
					t.Logf("Original: %s", orderFulfillmentString(tc.ofA))
				}
				if !assertEqualOrderFulfillments(t, tc.expB, of2, "Fulfill(A, B): B") {
					t.Logf("Original: %s", orderFulfillmentString(tc.ofB))
				}
			}

			of1, of2 = copyOrderFulfillment(tc.ofB), copyOrderFulfillment(tc.ofA)
			require.NotPanics(t, testFunc, "Fulfill(B, A)")
			assertions.AssertErrorValue(t, err, tc.swapErr, "Fulfill(B, A) error")
			if len(tc.expErr) == 0 {
				if !assertEqualOrderFulfillments(t, tc.expB, of1, "Fulfill(B, A): B") {
					t.Logf("Original: %s", orderFulfillmentString(tc.ofA))
				}
				if !assertEqualOrderFulfillments(t, tc.expA, of2, "Fulfill(B, A): A") {
					t.Logf("Original: %s", orderFulfillmentString(tc.ofB))
				}
			}
		})
	}
}

func TestGetFulfillmentAssetsAmt(t *testing.T) {
	newAskOF := func(orderID uint64, assetsUnfilled int64, assetDenom string) *OrderFulfillment {
		return &OrderFulfillment{
			Order: NewOrder(orderID).WithAsk(&AskOrder{
				Assets: sdk.NewInt64Coin(assetDenom, 999),
			}),
			AssetsUnfilledAmt: sdkmath.NewInt(assetsUnfilled),
		}
	}
	newBidOF := func(orderID uint64, assetsUnfilled int64, assetDenom string) *OrderFulfillment {
		return &OrderFulfillment{
			Order: NewOrder(orderID).WithBid(&BidOrder{
				Assets: sdk.NewInt64Coin(assetDenom, 999),
			}),
			AssetsUnfilledAmt: sdkmath.NewInt(assetsUnfilled),
		}
	}

	cases := []struct {
		name        string
		of1Unfilled int64
		of2Unfilled int64
		expAmt      int64
	}{
		{name: "of1 zero", of1Unfilled: 0, of2Unfilled: 3, expAmt: 0},
		{name: "of1 negative", of1Unfilled: -4, of2Unfilled: 3, expAmt: 0},
		{name: "of2 zero", of1Unfilled: 5, of2Unfilled: 0, expAmt: 0},
		{name: "of2 negative", of1Unfilled: 5, of2Unfilled: -6, expAmt: 0},
		{name: "equal", of1Unfilled: 8, of2Unfilled: 8, expAmt: 8},
		{name: "of1 has fewer", of1Unfilled: 9, of2Unfilled: 10, expAmt: 9},
		{name: "of2 has fewer", of1Unfilled: 12, of2Unfilled: 11, expAmt: 11},
	}

	type testCase struct {
		name   string
		of1    *OrderFulfillment
		of2    *OrderFulfillment
		expAmt sdkmath.Int
		expErr string
	}

	tests := make([]testCase, 0, len(cases)*4)

	for _, c := range cases {
		newTests := []testCase{
			{
				name:   "ask bid " + c.name,
				of1:    newAskOF(1, c.of1Unfilled, "one"),
				of2:    newBidOF(2, c.of2Unfilled, "two"),
				expAmt: sdkmath.NewInt(c.expAmt),
			},
			{
				name:   "bid ask " + c.name,
				of1:    newBidOF(1, c.of1Unfilled, "one"),
				of2:    newAskOF(2, c.of2Unfilled, "two"),
				expAmt: sdkmath.NewInt(c.expAmt),
			},
			{
				name:   "ask ask " + c.name,
				of1:    newAskOF(1, c.of1Unfilled, "one"),
				of2:    newAskOF(2, c.of2Unfilled, "two"),
				expAmt: sdkmath.NewInt(c.expAmt),
			},
			{
				name:   "bid bid " + c.name,
				of1:    newBidOF(1, c.of1Unfilled, "one"),
				of2:    newBidOF(2, c.of2Unfilled, "two"),
				expAmt: sdkmath.NewInt(c.expAmt),
			},
		}
		if c.expAmt == 0 {
			newTests[0].expErr = fmt.Sprintf("cannot fill ask order 1 having assets left \"%done\" "+
				"with bid order 2 having assets left \"%dtwo\": zero or negative assets left",
				c.of1Unfilled, c.of2Unfilled)
			newTests[1].expErr = fmt.Sprintf("cannot fill bid order 1 having assets left \"%done\" "+
				"with ask order 2 having assets left \"%dtwo\": zero or negative assets left",
				c.of1Unfilled, c.of2Unfilled)
			newTests[2].expErr = fmt.Sprintf("cannot fill ask order 1 having assets left \"%done\" "+
				"with ask order 2 having assets left \"%dtwo\": zero or negative assets left",
				c.of1Unfilled, c.of2Unfilled)
			newTests[3].expErr = fmt.Sprintf("cannot fill bid order 1 having assets left \"%done\" "+
				"with bid order 2 having assets left \"%dtwo\": zero or negative assets left",
				c.of1Unfilled, c.of2Unfilled)
		}
		tests = append(tests, newTests...)
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.expErr) > 0 {
				tc.expAmt = sdkmath.ZeroInt()
			}
			origOF1 := copyOrderFulfillment(tc.of1)
			origOF2 := copyOrderFulfillment(tc.of2)

			var amt sdkmath.Int
			var err error
			testFunc := func() {
				amt, err = GetFulfillmentAssetsAmt(tc.of1, tc.of2)
			}
			require.NotPanics(t, testFunc, "GetFulfillmentAssetsAmt")
			assertions.AssertErrorValue(t, err, tc.expErr, "GetFulfillmentAssetsAmt error")
			assert.Equal(t, tc.expAmt, amt, "GetFulfillmentAssetsAmt amount")
			assertEqualOrderFulfillments(t, origOF1, tc.of1, "of1 after GetFulfillmentAssetsAmt")
			assertEqualOrderFulfillments(t, origOF2, tc.of2, "of2 after GetFulfillmentAssetsAmt")
		})
	}
}

// TODO[1658]: func TestGetFulfillmentPriceAmt(t *testing.T)

func TestNewPartialFulfillment(t *testing.T) {
	sdkNewInt64CoinP := func(denom string, amt int64) *sdk.Coin {
		rv := sdk.NewInt64Coin(denom, amt)
		return &rv
	}

	tests := []struct {
		name     string
		f        *OrderFulfillment
		exp      *PartialFulfillment
		expPanic string
	}{
		{
			name: "ask order fees left",
			f: &OrderFulfillment{
				Order: NewOrder(54).WithAsk(&AskOrder{
					MarketId:                12,
					Seller:                  "the seller",
					Assets:                  sdk.NewInt64Coin("apple", 1234),
					Price:                   sdk.NewInt64Coin("pear", 9876),
					SellerSettlementFlatFee: sdkNewInt64CoinP("fig", 55),
					AllowPartial:            true,
				}),
				AssetsFilledAmt:   sdkmath.NewInt(234),
				AssetsUnfilledAmt: sdkmath.NewInt(1000),
				PriceAppliedAmt:   sdkmath.NewInt(10000),
				PriceLeftAmt:      sdkmath.NewInt(-124),
				OrderFeesLeft:     sdk.NewCoins(sdk.NewInt64Coin("fig", 50)),
				PriceFilledAmt:    sdkmath.NewInt(876),
				PriceUnfilledAmt:  sdkmath.NewInt(9000),
			},
			exp: &PartialFulfillment{
				NewOrder: NewOrder(54).WithAsk(&AskOrder{
					MarketId:                12,
					Seller:                  "the seller",
					Assets:                  sdk.NewInt64Coin("apple", 1000),
					Price:                   sdk.NewInt64Coin("pear", 9000),
					SellerSettlementFlatFee: sdkNewInt64CoinP("fig", 50),
					AllowPartial:            true,
				}),
				AssetsFilled: sdk.NewInt64Coin("apple", 234),
				PriceFilled:  sdk.NewInt64Coin("pear", 876),
			},
		},
		{
			name: "ask order no fees left",
			f: &OrderFulfillment{
				Order: NewOrder(54).WithAsk(&AskOrder{
					MarketId:                12,
					Seller:                  "the seller",
					Assets:                  sdk.NewInt64Coin("apple", 1234),
					Price:                   sdk.NewInt64Coin("pear", 9876),
					SellerSettlementFlatFee: sdkNewInt64CoinP("fig", 55),
					AllowPartial:            false,
				}),
				AssetsFilledAmt:   sdkmath.NewInt(234),
				AssetsUnfilledAmt: sdkmath.NewInt(1000),
				PriceAppliedAmt:   sdkmath.NewInt(10000),
				PriceLeftAmt:      sdkmath.NewInt(-124),
				OrderFeesLeft:     nil,
				PriceFilledAmt:    sdkmath.NewInt(876),
				PriceUnfilledAmt:  sdkmath.NewInt(9000),
			},
			exp: &PartialFulfillment{
				NewOrder: NewOrder(54).WithAsk(&AskOrder{
					MarketId:                12,
					Seller:                  "the seller",
					Assets:                  sdk.NewInt64Coin("apple", 1000),
					Price:                   sdk.NewInt64Coin("pear", 9000),
					SellerSettlementFlatFee: nil,
					AllowPartial:            false,
				}),
				AssetsFilled: sdk.NewInt64Coin("apple", 234),
				PriceFilled:  sdk.NewInt64Coin("pear", 876),
			},
			expPanic: "",
		},
		{
			name: "ask order multiple fees left",
			f: &OrderFulfillment{
				Order: NewOrder(54).WithAsk(&AskOrder{
					MarketId:                12,
					Seller:                  "the seller",
					Assets:                  sdk.NewInt64Coin("apple", 1234),
					Price:                   sdk.NewInt64Coin("pear", 9876),
					SellerSettlementFlatFee: sdkNewInt64CoinP("fig", 55),
					AllowPartial:            true,
				}),
				AssetsFilledAmt:   sdkmath.NewInt(234),
				AssetsUnfilledAmt: sdkmath.NewInt(1000),
				PriceAppliedAmt:   sdkmath.NewInt(10000),
				PriceLeftAmt:      sdkmath.NewInt(-124),
				OrderFeesLeft:     sdk.NewCoins(sdk.NewInt64Coin("fig", 50), sdk.NewInt64Coin("grape", 1)),
				PriceFilledAmt:    sdkmath.NewInt(876),
				PriceUnfilledAmt:  sdkmath.NewInt(9000),
			},
			expPanic: "partially filled ask order 54 somehow has multiple denoms in fees left \"50fig,1grape\"",
		},
		{
			name: "bid order",
			f: &OrderFulfillment{
				Order: NewOrder(54).WithBid(&BidOrder{
					MarketId:            12,
					Buyer:               "the buyer",
					Assets:              sdk.NewInt64Coin("apple", 1234),
					Price:               sdk.NewInt64Coin("pear", 9876),
					BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("fig", 55), sdk.NewInt64Coin("grape", 12)),
					AllowPartial:        true,
				}),
				AssetsFilledAmt:   sdkmath.NewInt(234),
				AssetsUnfilledAmt: sdkmath.NewInt(1000),
				PriceAppliedAmt:   sdkmath.NewInt(9875),
				PriceLeftAmt:      sdkmath.NewInt(1),
				OrderFeesLeft:     sdk.NewCoins(sdk.NewInt64Coin("fig", 50)),
				PriceFilledAmt:    sdkmath.NewInt(876),
				PriceUnfilledAmt:  sdkmath.NewInt(9000),
			},
			exp: &PartialFulfillment{
				NewOrder: NewOrder(54).WithBid(&BidOrder{
					MarketId:            12,
					Buyer:               "the buyer",
					Assets:              sdk.NewInt64Coin("apple", 1000),
					Price:               sdk.NewInt64Coin("pear", 9000),
					BuyerSettlementFees: sdk.NewCoins(sdk.NewInt64Coin("fig", 50)),
					AllowPartial:        true,
				}),
				AssetsFilled: sdk.NewInt64Coin("apple", 234),
				PriceFilled:  sdk.NewInt64Coin("pear", 876),
			},
		},
		{
			name: "nil order type",
			f: &OrderFulfillment{
				Order:           NewOrder(57),
				AssetsFilledAmt: sdkmath.NewInt(5),
				PriceFilledAmt:  sdkmath.NewInt(6),
			},
			expPanic: nilSubTypeErr(57),
		},
		{
			name: "unknown order type",
			f: &OrderFulfillment{
				Order:           newUnknownOrder(58),
				AssetsFilledAmt: sdkmath.NewInt(5),
				PriceFilledAmt:  sdkmath.NewInt(6),
			},
			expPanic: unknownSubTypeErr(58),
		},
		// I don't feel like creating a 3rd order type that implements SubOrderI which would be needed in order to
		// have a test case reach the final "order %d has unknown type %q" panic at the end of the func.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			origF := copyOrderFulfillment(tc.f)

			var actual *PartialFulfillment
			testFunc := func() {
				actual = NewPartialFulfillment(tc.f)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "NewPartialFulfillment")
			if !assert.Equal(t, tc.exp, actual, "NewPartialFulfillment result") {
				t.Logf("  Actual: %s", partialFulfillmentString(actual))
				t.Logf("Expected: %s", partialFulfillmentString(tc.exp))
			}
			assertEqualOrderFulfillments(t, origF, tc.f, "OrderFulfillment after NewPartialFulfillment")
		})
	}
}

func TestBuildFulfillments(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	coins := func(amount int64, denom string) sdk.Coins {
		return sdk.Coins{coin(amount, denom)}
	}

	askOrder := func(orderID uint64, assets sdk.Coin, price sdk.Coin, allowPartial bool, fees ...sdk.Coin) *Order {
		ao := &AskOrder{
			Seller:       "seller",
			Assets:       assets,
			Price:        price,
			AllowPartial: allowPartial,
		}
		if len(fees) > 1 {
			t.Fatalf("cannot create ask order %d with more than 1 fees %q", orderID, fees)
		}
		if len(fees) > 0 {
			ao.SellerSettlementFlatFee = &fees[0]
		}
		return NewOrder(orderID).WithAsk(ao)
	}
	bidOrder := func(orderID uint64, assets sdk.Coin, price sdk.Coin, allowPartial bool, fees ...sdk.Coin) *Order {
		return NewOrder(orderID).WithBid(&BidOrder{
			Buyer:               "buyer",
			Assets:              assets,
			Price:               price,
			AllowPartial:        allowPartial,
			BuyerSettlementFees: fees,
		})
	}
	filledOF := func(order *Order, priceAmt int64, splits []*OrderSplit, feesToPay ...sdk.Coins) *OrderFulfillment {
		rv := &OrderFulfillment{
			Order:             order,
			AssetsFilledAmt:   order.GetAssets().Amount,
			AssetsUnfilledAmt: ZeroAmtAfterSub,
			Splits:            splits,
			IsFinalized:       true,
			PriceFilledAmt:    order.GetPrice().Amount,
			PriceUnfilledAmt:  sdkmath.NewInt(0),
		}
		if priceAmt != 0 {
			rv.PriceAppliedAmt = sdkmath.NewInt(priceAmt)
		} else {
			rv.PriceAppliedAmt = order.GetPrice().Amount
		}
		rv.PriceLeftAmt = rv.PriceFilledAmt.Sub(rv.PriceAppliedAmt)
		if len(feesToPay) > 0 {
			rv.FeesToPay = feesToPay[0]
		}
		return rv
	}

	tests := []struct {
		name           string
		askOrders      []*Order
		bidOrders      []*Order
		sellerFeeRatio *FeeRatio
		expectedMaker  func(t *testing.T, askOrders, bidOrders []*Order) *Fulfillments
		expErr         string
	}{
		{
			name:           "one ask one bid, both fully filled",
			askOrders:      []*Order{askOrder(5, coin(10, "apple"), coin(55, "prune"), false, coin(8, "fig"))},
			bidOrders:      []*Order{bidOrder(6, coin(10, "apple"), coin(60, "prune"), false, coin(33, "fig"))},
			sellerFeeRatio: &FeeRatio{Price: coin(5, "prune"), Fee: coin(1, "fig")},
			expectedMaker: func(t *testing.T, askOrders, bidOrders []*Order) *Fulfillments {
				rv := &Fulfillments{
					AskOFs: []*OrderFulfillment{
						filledOF(askOrders[0], 60,
							[]*OrderSplit{{Assets: coin(10, "apple"), Price: coin(60, "prune")}}, // will be BidOFs[0].
							coins(20, "fig"),
						),
					},
					BidOFs: []*OrderFulfillment{
						filledOF(bidOrders[0], 0,
							[]*OrderSplit{{Assets: coin(10, "apple"), Price: coin(60, "prune")}}, // will be AskOFs[0].
							coins(33, "fig"),
						),
					},
					PartialOrder: nil,
				}

				rv.AskOFs[0].Splits[0].Order = rv.BidOFs[0]
				rv.BidOFs[0].Splits[0].Order = rv.AskOFs[0]

				return rv
			},
		},
		{
			name:           "one ask one bid, ask partially filled",
			askOrders:      []*Order{askOrder(7, coin(15, "apple"), coin(75, "prune"), true)},
			bidOrders:      []*Order{bidOrder(8, coin(10, "apple"), coin(60, "prune"), false)},
			sellerFeeRatio: &FeeRatio{Price: coin(5, "prune"), Fee: coin(1, "fig")},
			expectedMaker: func(t *testing.T, askOrders, bidOrders []*Order) *Fulfillments {
				rv := &Fulfillments{
					AskOFs: []*OrderFulfillment{
						{
							Order:             askOrders[0],
							AssetsFilledAmt:   sdkmath.NewInt(10),
							AssetsUnfilledAmt: sdkmath.NewInt(5),
							PriceAppliedAmt:   sdkmath.NewInt(60),
							PriceLeftAmt:      sdkmath.NewInt(15),
							Splits:            []*OrderSplit{{Assets: coin(10, "apple"), Price: coin(60, "prune")}}, // will be BidOFs[0].
							IsFinalized:       true,
							FeesToPay:         coins(12, "fig"),
							OrderFeesLeft:     nil,
							PriceFilledAmt:    sdkmath.NewInt(50),
							PriceUnfilledAmt:  sdkmath.NewInt(25),
						},
					},
					BidOFs: []*OrderFulfillment{
						filledOF(bidOrders[0], 0,
							[]*OrderSplit{{Assets: coin(10, "apple"), Price: coin(60, "prune")}}, // will be AskOFs[0].
						),
					},
					PartialOrder: &PartialFulfillment{
						NewOrder:     askOrder(7, coin(5, "apple"), coin(25, "prune"), true),
						AssetsFilled: coin(10, "apple"),
						PriceFilled:  coin(50, "prune"),
					},
				}

				rv.AskOFs[0].Splits[0].Order = rv.BidOFs[0]
				rv.BidOFs[0].Splits[0].Order = rv.AskOFs[0]

				return rv
			},
		},
		{
			name:           "one ask one bid, ask partially filled not allowed",
			askOrders:      []*Order{askOrder(7, coin(15, "apple"), coin(75, "prune"), false)},
			bidOrders:      []*Order{bidOrder(8, coin(10, "apple"), coin(60, "prune"), false)},
			sellerFeeRatio: &FeeRatio{Price: coin(5, "prune"), Fee: coin(1, "fig")},
			expErr:         "cannot fill ask order 7 having assets \"15apple\" with \"10apple\": order does not allow partial fill",
		},
		{
			name:           "one ask one bid, bid partially filled",
			askOrders:      []*Order{askOrder(9, coin(10, "apple"), coin(50, "prune"), false)},
			bidOrders:      []*Order{bidOrder(10, coin(15, "apple"), coin(90, "prune"), true)},
			sellerFeeRatio: &FeeRatio{Price: coin(5, "prune"), Fee: coin(1, "fig")},
			expectedMaker: func(t *testing.T, askOrders, bidOrders []*Order) *Fulfillments {
				rv := &Fulfillments{
					AskOFs: []*OrderFulfillment{
						filledOF(askOrders[0], 60,
							[]*OrderSplit{{Assets: coin(10, "apple"), Price: coin(60, "prune")}}, // will be BidOFs[0].
							coins(12, "fig"),
						),
					},
					BidOFs: []*OrderFulfillment{
						{
							Order:             bidOrders[0],
							AssetsFilledAmt:   sdkmath.NewInt(10),
							AssetsUnfilledAmt: sdkmath.NewInt(5),
							PriceAppliedAmt:   sdkmath.NewInt(60),
							PriceLeftAmt:      sdkmath.NewInt(30),
							Splits:            []*OrderSplit{{Assets: coin(10, "apple"), Price: coin(60, "prune")}}, // will be AskOFs[0].
							IsFinalized:       true,
							FeesToPay:         nil,
							OrderFeesLeft:     nil,
							PriceFilledAmt:    sdkmath.NewInt(60),
							PriceUnfilledAmt:  sdkmath.NewInt(30),
						},
					},
					PartialOrder: &PartialFulfillment{
						NewOrder:     bidOrder(10, coin(5, "apple"), coin(30, "prune"), true),
						AssetsFilled: coin(10, "apple"),
						PriceFilled:  coin(60, "prune"),
					},
				}

				rv.AskOFs[0].Splits[0].Order = rv.BidOFs[0]
				rv.BidOFs[0].Splits[0].Order = rv.AskOFs[0]

				return rv
			},
		},
		{
			name:           "one ask one bid, bid partially filled not allowed",
			askOrders:      []*Order{askOrder(9, coin(10, "apple"), coin(50, "prune"), false)},
			bidOrders:      []*Order{bidOrder(10, coin(15, "apple"), coin(90, "prune"), false)},
			sellerFeeRatio: &FeeRatio{Price: coin(5, "prune"), Fee: coin(1, "fig")},
			expErr:         "cannot fill bid order 10 having assets \"15apple\" with \"10apple\": order does not allow partial fill",
		},
		{
			name:      "one ask filled by five bids",
			askOrders: []*Order{askOrder(21, coin(12, "apple"), coin(60, "prune"), false)},
			bidOrders: []*Order{
				bidOrder(22, coin(1, "apple"), coin(10, "prune"), false),
				bidOrder(22, coin(1, "apple"), coin(12, "prune"), false),
				bidOrder(22, coin(2, "apple"), coin(1, "prune"), false),
				bidOrder(22, coin(3, "apple"), coin(15, "prune"), false),
				bidOrder(22, coin(5, "apple"), coin(25, "prune"), false),
			},
			sellerFeeRatio: &FeeRatio{Price: coin(5, "prune"), Fee: coin(1, "fig")},
			expectedMaker: func(t *testing.T, askOrders, bidOrders []*Order) *Fulfillments {
				rv := &Fulfillments{
					AskOFs: []*OrderFulfillment{
						filledOF(askOrders[0], 63,
							[]*OrderSplit{
								{Assets: coin(1, "apple"), Price: coin(10, "prune")}, // Will be BidOFs[0].
								{Assets: coin(1, "apple"), Price: coin(12, "prune")}, // Will be BidOFs[1].
								{Assets: coin(2, "apple"), Price: coin(1, "prune")},  // Will be BidOFs[2].
								{Assets: coin(3, "apple"), Price: coin(15, "prune")}, // Will be BidOFs[3].
								{Assets: coin(5, "apple"), Price: coin(25, "prune")}, // Will be BidOFs[4].
							},
							coins(13, "fig"),
						),
					},
					BidOFs: []*OrderFulfillment{
						filledOF(bidOrders[0], 0,
							[]*OrderSplit{{Assets: coin(1, "apple"), Price: coin(10, "prune")}}, // Will be AskOFs[0].
						),
						filledOF(bidOrders[1], 0,
							[]*OrderSplit{{Assets: coin(1, "apple"), Price: coin(12, "prune")}}, // Will be AskOFs[0].
						),
						filledOF(bidOrders[2], 0,
							[]*OrderSplit{{Assets: coin(2, "apple"), Price: coin(1, "prune")}}, // Will be AskOFs[0].

						),
						filledOF(bidOrders[3], 0,
							[]*OrderSplit{{Assets: coin(3, "apple"), Price: coin(15, "prune")}}, // Will be AskOFs[0].
						),
						filledOF(bidOrders[4], 0,
							[]*OrderSplit{{Assets: coin(5, "apple"), Price: coin(25, "prune")}}, // Will be AskOFs[0].
						),
					},
				}

				rv.AskOFs[0].Splits[0].Order = rv.BidOFs[0]
				rv.AskOFs[0].Splits[1].Order = rv.BidOFs[1]
				rv.AskOFs[0].Splits[2].Order = rv.BidOFs[2]
				rv.AskOFs[0].Splits[3].Order = rv.BidOFs[3]
				rv.AskOFs[0].Splits[4].Order = rv.BidOFs[4]

				rv.BidOFs[0].Splits[0].Order = rv.AskOFs[0]
				rv.BidOFs[1].Splits[0].Order = rv.AskOFs[0]
				rv.BidOFs[2].Splits[0].Order = rv.AskOFs[0]
				rv.BidOFs[3].Splits[0].Order = rv.AskOFs[0]
				rv.BidOFs[4].Splits[0].Order = rv.AskOFs[0]

				return rv
			},
		},
		{
			name: "one indivisible bid filled by five asks",
			askOrders: []*Order{
				askOrder(31, coin(1, "apple"), coin(16, "prune"), false),
				askOrder(33, coin(1, "apple"), coin(16, "prune"), false),
				askOrder(35, coin(1, "apple"), coin(17, "prune"), false),
				askOrder(37, coin(13, "apple"), coin(209, "prune"), false),
				askOrder(39, coin(15, "apple"), coin(241, "prune"), false),
			},
			bidOrders:      []*Order{bidOrder(30, coin(31, "apple"), coin(500, "prune"), false)},
			sellerFeeRatio: &FeeRatio{Price: coin(5, "prune"), Fee: coin(1, "fig")},
			expectedMaker: func(t *testing.T, askOrders, bidOrders []*Order) *Fulfillments {
				rv := &Fulfillments{
					AskOFs: []*OrderFulfillment{
						filledOF(askOrders[0], 17,
							[]*OrderSplit{{Assets: coin(1, "apple"), Price: coin(17, "prune")}}, // Will be BidOfs[0]
							coins(4, "fig"),
						),
						filledOF(askOrders[1], 0,
							[]*OrderSplit{{Assets: coin(1, "apple"), Price: coin(16, "prune")}}, // Will be BidOfs[0]
							coins(4, "fig"),
						),
						filledOF(askOrders[2], 0,
							[]*OrderSplit{{Assets: coin(1, "apple"), Price: coin(17, "prune")}}, // Will be BidOfs[0]
							coins(4, "fig"),
						),
						filledOF(askOrders[3], 0,
							[]*OrderSplit{{Assets: coin(13, "apple"), Price: coin(209, "prune")}}, // Will be BidOfs[0]
							coins(42, "fig"),
						),
						filledOF(askOrders[4], 0,
							[]*OrderSplit{{Assets: coin(15, "apple"), Price: coin(241, "prune")}}, // Will be BidOfs[0]
							coins(49, "fig"),
						),
					},
					BidOFs: []*OrderFulfillment{
						{
							Order:             bidOrders[0],
							AssetsFilledAmt:   sdkmath.NewInt(31),
							AssetsUnfilledAmt: ZeroAmtAfterSub,
							PriceAppliedAmt:   sdkmath.NewInt(500),
							PriceLeftAmt:      ZeroAmtAfterSub,
							Splits: []*OrderSplit{
								{Assets: coin(1, "apple"), Price: coin(17, "prune")},   // Will be AskOFs[0]
								{Assets: coin(1, "apple"), Price: coin(16, "prune")},   // Will be AskOFs[1]
								{Assets: coin(1, "apple"), Price: coin(17, "prune")},   // Will be AskOFs[2]
								{Assets: coin(13, "apple"), Price: coin(209, "prune")}, // Will be AskOFs[3]
								{Assets: coin(15, "apple"), Price: coin(241, "prune")}, // Will be AskOFs[4]
							},
							IsFinalized:      true,
							PriceFilledAmt:   sdkmath.NewInt(500),
							PriceUnfilledAmt: sdkmath.NewInt(0),
						},
					},
				}

				rv.AskOFs[0].Splits[0].Order = rv.BidOFs[0]
				rv.AskOFs[1].Splits[0].Order = rv.BidOFs[0]
				rv.AskOFs[2].Splits[0].Order = rv.BidOFs[0]
				rv.AskOFs[3].Splits[0].Order = rv.BidOFs[0]
				rv.AskOFs[4].Splits[0].Order = rv.BidOFs[0]

				rv.BidOFs[0].Splits[0].Order = rv.AskOFs[0]
				rv.BidOFs[0].Splits[1].Order = rv.AskOFs[1]
				rv.BidOFs[0].Splits[2].Order = rv.AskOFs[2]
				rv.BidOFs[0].Splits[3].Order = rv.AskOFs[3]
				rv.BidOFs[0].Splits[4].Order = rv.AskOFs[4]

				return rv
			},
		},
		{
			name: "three asks three bids, each fully fills the other",
			askOrders: []*Order{
				askOrder(51, coin(8, "apple"), coin(55, "prune"), false, coin(18, "grape")),
				askOrder(53, coin(12, "apple"), coin(18, "prune"), false, coin(1, "grape")),
				askOrder(55, coin(344, "apple"), coin(12345, "prune"), false, coin(99, "grape")),
			},
			bidOrders: []*Order{
				bidOrder(52, coin(8, "apple"), coin(55, "prune"), false, coin(3, "fig")),
				bidOrder(54, coin(12, "apple"), coin(18, "prune"), false, coin(7, "fig")),
				bidOrder(56, coin(344, "apple"), coin(12345, "prune"), false, coin(2, "fig")),
			},
			expectedMaker: func(t *testing.T, askOrders, bidOrders []*Order) *Fulfillments {
				rv := &Fulfillments{
					AskOFs: []*OrderFulfillment{
						filledOF(askOrders[0], 0,
							[]*OrderSplit{{Assets: coin(8, "apple"), Price: coin(55, "prune")}}, // Will be BidOFs[0]
							coins(18, "grape"),
						),
						filledOF(askOrders[1], 0,
							[]*OrderSplit{{Assets: coin(12, "apple"), Price: coin(18, "prune")}}, // Will be BidOFs[1]
							coins(1, "grape"),
						),
						filledOF(askOrders[2], 0,
							[]*OrderSplit{{Assets: coin(344, "apple"), Price: coin(12345, "prune")}}, // Will be BidOFs[2]
							coins(99, "grape"),
						),
					},
					BidOFs: []*OrderFulfillment{
						filledOF(bidOrders[0], 0,
							[]*OrderSplit{{Assets: coin(8, "apple"), Price: coin(55, "prune")}}, // Will be AskOFs[0]
							coins(3, "fig"),
						),
						filledOF(bidOrders[1], 0,
							[]*OrderSplit{{Assets: coin(12, "apple"), Price: coin(18, "prune")}}, // Will be AskOFs[1]
							coins(7, "fig"),
						),
						filledOF(bidOrders[2], 0,
							[]*OrderSplit{{Assets: coin(344, "apple"), Price: coin(12345, "prune")}}, // Will be AskOFs[2]
							coins(2, "fig"),
						),
					},
					PartialOrder: nil,
				}

				rv.AskOFs[0].Splits[0].Order = rv.BidOFs[0]
				rv.AskOFs[1].Splits[0].Order = rv.BidOFs[1]
				rv.AskOFs[2].Splits[0].Order = rv.BidOFs[2]

				rv.BidOFs[0].Splits[0].Order = rv.AskOFs[0]
				rv.BidOFs[1].Splits[0].Order = rv.AskOFs[1]
				rv.BidOFs[2].Splits[0].Order = rv.AskOFs[2]

				return rv
			},
		},
		{
			name: "three asks two bids, all fully filled",
			askOrders: []*Order{
				askOrder(11, coin(10, "apple"), coin(50, "prune"), false),
				askOrder(13, coin(20, "apple"), coin(100, "prune"), false),
				askOrder(15, coin(50, "apple"), coin(250, "prune"), false),
			},
			bidOrders: []*Order{
				bidOrder(12, coin(23, "apple"), coin(115, "prune"), false),
				bidOrder(14, coin(57, "apple"), coin(285, "prune"), false),
			},
			sellerFeeRatio: &FeeRatio{Price: coin(5, "prune"), Fee: coin(1, "fig")},
			expectedMaker: func(t *testing.T, askOrders, bidOrders []*Order) *Fulfillments {
				rv := &Fulfillments{
					AskOFs: []*OrderFulfillment{
						filledOF(askOrders[0], 0,
							[]*OrderSplit{{Assets: coin(10, "apple"), Price: coin(50, "prune")}}, // Will be BidOFs[0]
							coins(10, "fig"),
						),
						filledOF(askOrders[1], 0,
							[]*OrderSplit{
								{Assets: coin(13, "apple"), Price: coin(65, "prune")}, // Will be BidOFs[0]
								{Assets: coin(7, "apple"), Price: coin(35, "prune")},  // Will be BidOFs[1]
							},
							coins(20, "fig"),
						),
						filledOF(askOrders[2], 0,
							[]*OrderSplit{{Assets: coin(50, "apple"), Price: coin(250, "prune")}}, // Will be BidOFs[1]
							coins(50, "fig"),
						),
					},
					BidOFs: []*OrderFulfillment{
						filledOF(bidOrders[0], 0,
							[]*OrderSplit{
								{Assets: coin(10, "apple"), Price: coin(50, "prune")}, // Will be AskOfs[0]
								{Assets: coin(13, "apple"), Price: coin(65, "prune")}, // Will be AskOfs[1]
							},
						),
						filledOF(bidOrders[1], 0,
							[]*OrderSplit{
								{Assets: coin(7, "apple"), Price: coin(35, "prune")},   // Will be AskOFs[1]
								{Assets: coin(50, "apple"), Price: coin(250, "prune")}, // Will be AskOFs[2]
							},
						),
					},
				}

				rv.AskOFs[0].Splits[0].Order = rv.BidOFs[0]
				rv.AskOFs[1].Splits[0].Order = rv.BidOFs[0]
				rv.AskOFs[1].Splits[1].Order = rv.BidOFs[1]
				rv.AskOFs[2].Splits[0].Order = rv.BidOFs[1]
				rv.BidOFs[0].Splits[0].Order = rv.AskOFs[0]
				rv.BidOFs[0].Splits[1].Order = rv.AskOFs[1]
				rv.BidOFs[1].Splits[0].Order = rv.AskOFs[1]
				rv.BidOFs[1].Splits[1].Order = rv.AskOFs[2]

				return rv
			},
		},
		{
			name: "three asks two bids, ask partially filled",
			askOrders: []*Order{
				askOrder(73, coin(10, "apple"), coin(50, "prune"), false),
				askOrder(75, coin(15, "apple"), coin(75, "prune"), false),
				askOrder(77, coin(25, "apple"), coin(125, "prune"), true),
			},
			bidOrders: []*Order{
				bidOrder(74, coin(5, "apple"), coin(25, "prune"), false),
				bidOrder(76, coin(40, "apple"), coin(200, "prune"), false),
			},
			expectedMaker: func(t *testing.T, askOrders, bidOrders []*Order) *Fulfillments {
				rv := &Fulfillments{
					AskOFs: []*OrderFulfillment{
						filledOF(askOrders[0], 0,
							[]*OrderSplit{
								{Assets: coin(5, "apple"), Price: coin(25, "prune")}, // Will be BidOFs[0]
								{Assets: coin(5, "apple"), Price: coin(25, "prune")}, // Will be BidOFs[1]
							},
						),
						filledOF(askOrders[1], 0,
							[]*OrderSplit{{Assets: coin(15, "apple"), Price: coin(75, "prune")}}, // Will be BidOFs[1]
						),
						{
							Order:             askOrders[2],
							AssetsFilledAmt:   sdkmath.NewInt(20),
							AssetsUnfilledAmt: sdkmath.NewInt(5),
							PriceAppliedAmt:   sdkmath.NewInt(100),
							PriceLeftAmt:      sdkmath.NewInt(25),
							Splits:            []*OrderSplit{{Assets: coin(20, "apple"), Price: coin(100, "prune")}}, // Will be BidOFs[1],
							IsFinalized:       true,
							PriceFilledAmt:    sdkmath.NewInt(100),
							PriceUnfilledAmt:  sdkmath.NewInt(25),
						},
					},
					BidOFs: []*OrderFulfillment{
						filledOF(bidOrders[0], 0,
							[]*OrderSplit{{Assets: coin(5, "apple"), Price: coin(25, "prune")}}, // Will be AskOFs[0],
						),
						filledOF(bidOrders[1], 0,
							[]*OrderSplit{
								{Assets: coin(5, "apple"), Price: coin(25, "prune")},   // Will be AskOFs[0],
								{Assets: coin(15, "apple"), Price: coin(75, "prune")},  // Will be AskOFs[1],
								{Assets: coin(20, "apple"), Price: coin(100, "prune")}, // Will be AskOFs[2],
							},
						),
					},
					PartialOrder: &PartialFulfillment{
						NewOrder:     askOrder(77, coin(5, "apple"), coin(25, "prune"), true),
						AssetsFilled: coin(20, "apple"),
						PriceFilled:  coin(100, "prune"),
					},
				}

				rv.AskOFs[0].Splits[0].Order = rv.BidOFs[0]
				rv.AskOFs[0].Splits[1].Order = rv.BidOFs[1]
				rv.AskOFs[1].Splits[0].Order = rv.BidOFs[1]
				rv.AskOFs[2].Splits[0].Order = rv.BidOFs[1]

				rv.BidOFs[0].Splits[0].Order = rv.AskOFs[0]
				rv.BidOFs[1].Splits[0].Order = rv.AskOFs[0]
				rv.BidOFs[1].Splits[1].Order = rv.AskOFs[1]
				rv.BidOFs[1].Splits[2].Order = rv.AskOFs[2]

				return rv
			},
		},
		{
			name: "three asks two bids, ask partially filled not allowed",
			askOrders: []*Order{
				askOrder(73, coin(10, "apple"), coin(50, "prune"), false),
				askOrder(75, coin(15, "apple"), coin(75, "prune"), false),
				askOrder(77, coin(25, "apple"), coin(125, "prune"), false),
			},
			bidOrders: []*Order{
				bidOrder(74, coin(5, "apple"), coin(25, "prune"), false),
				bidOrder(76, coin(40, "apple"), coin(200, "prune"), false),
			},
			expErr: "cannot fill ask order 77 having assets \"25apple\" with \"20apple\": order does not allow partial fill",
		},
		{
			name: "three asks two bids, bid partially filled",
			askOrders: []*Order{
				askOrder(121, coin(55, "apple"), coin(275, "prune"), false),
				askOrder(123, coin(12, "apple"), coin(60, "prune"), false),
				askOrder(125, coin(13, "apple"), coin(65, "prune"), false),
			},
			bidOrders: []*Order{
				bidOrder(124, coin(65, "apple"), coin(325, "prune"), false),
				bidOrder(126, coin(20, "apple"), coin(100, "prune"), true),
			},
			expectedMaker: func(t *testing.T, askOrders, bidOrders []*Order) *Fulfillments {
				rv := &Fulfillments{
					AskOFs: []*OrderFulfillment{
						filledOF(askOrders[0], 0,
							[]*OrderSplit{{Assets: coin(55, "apple"), Price: coin(275, "prune")}}, // Will be BidOFs[0]
						),
						filledOF(askOrders[1], 0,
							[]*OrderSplit{
								{Assets: coin(10, "apple"), Price: coin(50, "prune")}, // Will be BidOFs[0]
								{Assets: coin(2, "apple"), Price: coin(10, "prune")},  // Will be BidOFs[1]
							},
						),
						filledOF(askOrders[2], 0,
							[]*OrderSplit{{Assets: coin(13, "apple"), Price: coin(65, "prune")}}, // Will be BidOFs[1]
						),
					},
					BidOFs: []*OrderFulfillment{
						filledOF(bidOrders[0], 0,
							[]*OrderSplit{
								{Assets: coin(55, "apple"), Price: coin(275, "prune")}, // Will be AskOFs[0]
								{Assets: coin(10, "apple"), Price: coin(50, "prune")},  // Will be AskOFs[1]
							},
						),
						{
							Order:             bidOrders[1],
							AssetsFilledAmt:   sdkmath.NewInt(15),
							AssetsUnfilledAmt: sdkmath.NewInt(5),
							PriceAppliedAmt:   sdkmath.NewInt(75),
							PriceLeftAmt:      sdkmath.NewInt(25),
							Splits: []*OrderSplit{
								{Assets: coin(2, "apple"), Price: coin(10, "prune")},  // Will be AskOFs[1]
								{Assets: coin(13, "apple"), Price: coin(65, "prune")}, // Will be AskOFs[2]
							},
							IsFinalized:      true,
							PriceFilledAmt:   sdkmath.NewInt(75),
							PriceUnfilledAmt: sdkmath.NewInt(25),
						},
					},
					PartialOrder: &PartialFulfillment{
						NewOrder:     bidOrder(126, coin(5, "apple"), coin(25, "prune"), true),
						AssetsFilled: coin(15, "apple"),
						PriceFilled:  coin(75, "prune"),
					},
				}

				rv.AskOFs[0].Splits[0].Order = rv.BidOFs[0]
				rv.AskOFs[1].Splits[0].Order = rv.BidOFs[0]
				rv.AskOFs[1].Splits[1].Order = rv.BidOFs[1]
				rv.AskOFs[2].Splits[0].Order = rv.BidOFs[1]

				rv.BidOFs[0].Splits[0].Order = rv.AskOFs[0]
				rv.BidOFs[0].Splits[1].Order = rv.AskOFs[1]
				rv.BidOFs[1].Splits[0].Order = rv.AskOFs[1]
				rv.BidOFs[1].Splits[1].Order = rv.AskOFs[2]

				return rv
			},
		},
		{
			// TODO[1658]: Either update the process or delete this 2 asks 1 bid unit test.
			name: "two asks one bid",
			askOrders: []*Order{
				askOrder(91, coin(10, "apple"), coin(49, "prune"), false),
				askOrder(93, coin(10, "apple"), coin(51, "prune"), false),
			},
			bidOrders: []*Order{bidOrder(92, coin(20, "apple"), coin(100, "prune"), false)},
			expectedMaker: func(t *testing.T, askOrders, bidOrders []*Order) *Fulfillments {
				return &Fulfillments{
					AskOFs: []*OrderFulfillment{
						filledOF(askOrders[0], 0,
							[]*OrderSplit{{Assets: coin(10, "apple"), Price: coin(49, "prune")}}, // Will be BidOFs[0]
						),
						filledOF(askOrders[1], 0,
							[]*OrderSplit{{Assets: coin(10, "apple"), Price: coin(51, "prune")}}, // Will be BidOFs[0]
						),
					},
					BidOFs: []*OrderFulfillment{
						filledOF(askOrders[1], 0,
							[]*OrderSplit{{Assets: coin(10, "apple"), Price: coin(49, "prune")}}, // Will be AskOFs[0]
						),
						filledOF(askOrders[1], 0,
							[]*OrderSplit{{Assets: coin(10, "apple"), Price: coin(51, "prune")}}, // Will be AskOFs[1]
						),
					},
				}
			},
		},
		{
			name: "ask order in bid order list",
			askOrders: []*Order{
				askOrder(1, coin(1, "apple"), coin(1, "prune"), false),
				bidOrder(2, coin(1, "apple"), coin(1, "prune"), false),
				askOrder(3, coin(1, "apple"), coin(1, "prune"), false),
			},
			bidOrders: []*Order{bidOrder(4, coin(3, "apple"), coin(3, "prune"), false)},
			expErr:    "bid order 2 is not an ask order but is in the askOrders list",
		},
		{
			name:      "bid order in ask order list",
			askOrders: []*Order{askOrder(4, coin(3, "apple"), coin(3, "prune"), false)},
			bidOrders: []*Order{
				bidOrder(1, coin(1, "apple"), coin(1, "prune"), false),
				askOrder(2, coin(1, "apple"), coin(1, "prune"), false),
				bidOrder(3, coin(1, "apple"), coin(1, "prune"), false),
			},
			expErr: "ask order 2 is not a bid order but is in the bidOrders list",
		},
		// neither filled in full - I'm not sure how I can trigger this.
		{
			name:      "ask finalize error",
			askOrders: []*Order{askOrder(15, coin(13, "apple"), coin(17, "prune"), true)},
			bidOrders: []*Order{bidOrder(16, coin(5, "apple"), coin(20, "prune"), true)},
			expErr:    "ask order 15 having assets \"13apple\" cannot be partially filled by \"5apple\": price \"17prune\" is not evenly divisible",
		},
		{
			name:      "bid finalize error",
			askOrders: []*Order{askOrder(15, coin(5, "apple"), coin(5, "prune"), true)},
			bidOrders: []*Order{bidOrder(16, coin(13, "apple"), coin(17, "prune"), true)},
			expErr:    "bid order 16 having assets \"13apple\" cannot be partially filled by \"5apple\": price \"17prune\" is not evenly divisible",
		},
		{
			name:      "validate error",
			askOrders: []*Order{askOrder(123, coin(5, "apple"), coin(6, "prune"), true)},
			bidOrders: []*Order{bidOrder(124, coin(5, "apple"), coin(5, "prune"), true)},
			expErr:    "ask order 123 having assets \"5apple\" and price \"6prune\" cannot be filled by \"5apple\" at price \"5prune\": insufficient price",
		},
		{
			name:      "nil askOrders",
			askOrders: nil,
			bidOrders: []*Order{bidOrder(124, coin(5, "apple"), coin(5, "prune"), true)},
			expErr:    "no assets filled in bid order 124",
		},
		{
			name:      "empty askOrders",
			askOrders: []*Order{},
			bidOrders: []*Order{bidOrder(124, coin(5, "apple"), coin(5, "prune"), true)},
			expErr:    "no assets filled in bid order 124",
		},
		{
			name:      "nil bidOrders",
			askOrders: []*Order{askOrder(123, coin(5, "apple"), coin(6, "prune"), true)},
			bidOrders: nil,
			expErr:    "no assets filled in ask order 123",
		},
		{
			name:      "empty bidOrders",
			askOrders: []*Order{askOrder(123, coin(5, "apple"), coin(6, "prune"), true)},
			bidOrders: []*Order{},
			expErr:    "no assets filled in ask order 123",
		},
		{
			name: "ask not filled at all", // this gets caught by Finalize.
			askOrders: []*Order{
				askOrder(123, coin(10, "apple"), coin(10, "prune"), true),
				askOrder(125, coin(5, "apple"), coin(5, "prune"), true),
			},
			bidOrders: []*Order{bidOrder(124, coin(5, "apple"), coin(5, "prune"), true)},
			expErr:    "no assets filled in ask order 125",
		},
		{
			name:      "bid not filled at all", // this gets caught by Finalize.
			askOrders: []*Order{askOrder(123, coin(5, "apple"), coin(5, "prune"), true)},
			bidOrders: []*Order{
				bidOrder(122, coin(10, "apple"), coin(10, "prune"), true),
				bidOrder(124, coin(5, "apple"), coin(5, "prune"), true),
			},
			expErr: "no assets filled in bid order 124",
		},
		// both ask and bid partially filled - I'm not sure how to trigger this.
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var expected *Fulfillments
			if tc.expectedMaker != nil {
				expected = tc.expectedMaker(t, tc.askOrders, tc.bidOrders)
			}
			var actual *Fulfillments
			var err error
			testFunc := func() {
				actual, err = BuildFulfillments(tc.askOrders, tc.bidOrders, tc.sellerFeeRatio)
			}
			require.NotPanics(t, testFunc, "BuildFulfillments")
			assertions.AssertErrorValue(t, err, tc.expErr, "BuildFulfillments error")
			if !assert.Equal(t, expected, actual, "BuildFulfillments result") && expected != nil && actual != nil {
				// Try to help identify the error in the failure logs.
				expAskOFs := orderFulfillmentsString(expected.AskOFs)
				actAskOFs := orderFulfillmentsString(actual.AskOFs)
				if assert.Equal(t, expAskOFs, actAskOFs, "AskOFs") {
					// Some difference don't come through in the strings, so dig even deeper.
					for i := range expected.AskOFs {
						assertEqualOrderFulfillments(t, expected.AskOFs[i], actual.AskOFs[i], "AskOFs[%d]", i)
					}
				}
				expBidOFs := orderFulfillmentsString(expected.BidOFs)
				actBidOFs := orderFulfillmentsString(actual.BidOFs)
				if assert.Equal(t, expBidOFs, actBidOFs, "BidOFs") {
					for i := range expected.BidOFs {
						assertEqualOrderFulfillments(t, expected.BidOFs[i], actual.BidOFs[i], "BidOFs[%d]", i)
					}
				}
				expPartial := partialFulfillmentString(expected.PartialOrder)
				actPartial := partialFulfillmentString(actual.PartialOrder)
				assert.Equal(t, expPartial, actPartial, "PartialOrder")
			}
		})
	}
}

// copyIndexedAddrAmts creates a deep copy of an indexedAddrAmts.
func copyIndexedAddrAmts(orig *indexedAddrAmts) *indexedAddrAmts {
	if orig == nil {
		return nil
	}

	rv := &indexedAddrAmts{
		addrs:   nil,
		amts:    nil,
		indexes: nil,
	}

	if orig.addrs != nil {
		rv.addrs = make([]string, 0, len(orig.addrs))
		rv.addrs = append(rv.addrs, orig.addrs...)
	}

	if orig.amts != nil {
		rv.amts = make([]sdk.Coins, len(orig.amts))
		for i, amt := range orig.amts {
			rv.amts[i] = copyCoins(amt)
		}
	}

	if orig.indexes != nil {
		rv.indexes = make(map[string]int, len(orig.indexes))
		for k, v := range orig.indexes {
			rv.indexes[k] = v
		}
	}

	return rv
}

// String converts a indexedAddrAmtsString to a string.
// This is mostly because test failure output of sdk.Coin and sdk.Coins is impossible to understand.
func indexedAddrAmtsString(i *indexedAddrAmts) string {
	if i == nil {
		return "nil"
	}

	addrs := "nil"
	if i.addrs != nil {
		addrsVals := make([]string, len(i.addrs))
		for j, addr := range i.addrs {
			addrsVals[j] = fmt.Sprintf("%q", addr)
		}
		addrs = fmt.Sprintf("%T{%s}", i.addrs, strings.Join(addrsVals, ", "))
	}

	amts := "nil"
	if i.amts != nil {
		amtsVals := make([]string, len(i.amts))
		for j, amt := range i.amts {
			amtsVals[j] = fmt.Sprintf("%q", amt)
		}
		amts = fmt.Sprintf("[]%T{%s}", i.amts, strings.Join(amtsVals, ", "))
	}

	indexes := "nil"
	if i.indexes != nil {
		indexVals := make([]string, 0, len(i.indexes))
		for k, v := range i.indexes {
			indexVals = append(indexVals, fmt.Sprintf("%q: %d", k, v))
		}
		sort.Strings(indexVals)
		indexes = fmt.Sprintf("%T{%s}", i.indexes, strings.Join(indexVals, ", "))
	}

	return fmt.Sprintf("%T{addrs:%s, amts:%s, indexes:%s}", i, addrs, amts, indexes)
}

func TestNewIndexedAddrAmts(t *testing.T) {
	expected := &indexedAddrAmts{
		addrs:   nil,
		amts:    nil,
		indexes: make(map[string]int),
	}
	actual := newIndexedAddrAmts()
	assert.Equal(t, expected, actual, "newIndexedAddrAmts result")
	key := "test"
	require.NotPanics(t, func() {
		_ = actual.indexes[key]
	}, "getting value of actual.indexes[%q]", key)
}

func TestIndexedAddrAmts_Add(t *testing.T) {
	coins := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "sdk.ParseCoinsNormalized(%q)", coins)
		return rv
	}
	negCoins := sdk.Coins{sdk.Coin{Denom: "neg", Amount: sdkmath.NewInt(-1)}}

	tests := []struct {
		name     string
		receiver *indexedAddrAmts
		addr     string
		coins    []sdk.Coin
		expected *indexedAddrAmts
		expPanic string
	}{
		{
			name:     "empty, add one coin",
			receiver: newIndexedAddrAmts(),
			addr:     "addr1",
			coins:    coins("1one"),
			expected: &indexedAddrAmts{
				addrs:   []string{"addr1"},
				amts:    []sdk.Coins{coins("1one")},
				indexes: map[string]int{"addr1": 0},
			},
		},
		{
			name:     "empty, add two coins",
			receiver: newIndexedAddrAmts(),
			addr:     "addr1",
			coins:    coins("1one,2two"),
			expected: &indexedAddrAmts{
				addrs:   []string{"addr1"},
				amts:    []sdk.Coins{coins("1one,2two")},
				indexes: map[string]int{"addr1": 0},
			},
		},
		{
			name:     "empty, add neg coins",
			receiver: newIndexedAddrAmts(),
			addr:     "addr1",
			coins:    negCoins,
			expPanic: "cannot index and add invalid coin amount \"-1neg\"",
		},
		{
			name: "one addr, add to existing new denom",
			receiver: &indexedAddrAmts{
				addrs:   []string{"addr1"},
				amts:    []sdk.Coins{coins("1one")},
				indexes: map[string]int{"addr1": 0},
			},
			addr:  "addr1",
			coins: coins("2two"),
			expected: &indexedAddrAmts{
				addrs:   []string{"addr1"},
				amts:    []sdk.Coins{coins("1one,2two")},
				indexes: map[string]int{"addr1": 0},
			},
		},
		{
			name: "one addr, add to existing same denom",
			receiver: &indexedAddrAmts{
				addrs:   []string{"addr1"},
				amts:    []sdk.Coins{coins("1one")},
				indexes: map[string]int{"addr1": 0},
			},
			addr:  "addr1",
			coins: coins("3one"),
			expected: &indexedAddrAmts{
				addrs:   []string{"addr1"},
				amts:    []sdk.Coins{coins("4one")},
				indexes: map[string]int{"addr1": 0},
			},
		},
		{
			name: "one addr, add negative to existing",
			receiver: &indexedAddrAmts{
				addrs:   []string{"addr1"},
				amts:    []sdk.Coins{coins("1one")},
				indexes: map[string]int{"addr1": 0},
			},
			addr:     "addr1",
			coins:    negCoins,
			expPanic: "cannot index and add invalid coin amount \"-1neg\"",
		},
		{
			name: "one addr, add to new",
			receiver: &indexedAddrAmts{
				addrs:   []string{"addr1"},
				amts:    []sdk.Coins{coins("1one")},
				indexes: map[string]int{"addr1": 0},
			},
			addr:  "addr2",
			coins: coins("2two"),
			expected: &indexedAddrAmts{
				addrs:   []string{"addr1", "addr2"},
				amts:    []sdk.Coins{coins("1one"), coins("2two")},
				indexes: map[string]int{"addr1": 0, "addr2": 1},
			},
		},
		{
			name: "one addr, add to new opposite order",
			receiver: &indexedAddrAmts{
				addrs:   []string{"addr2"},
				amts:    []sdk.Coins{coins("2two")},
				indexes: map[string]int{"addr2": 0},
			},
			addr:  "addr1",
			coins: coins("1one"),
			expected: &indexedAddrAmts{
				addrs:   []string{"addr2", "addr1"},
				amts:    []sdk.Coins{coins("2two"), coins("1one")},
				indexes: map[string]int{"addr2": 0, "addr1": 1},
			},
		},
		{
			name: "one addr, add negative to new",
			receiver: &indexedAddrAmts{
				addrs:   []string{"addr1"},
				amts:    []sdk.Coins{coins("1one")},
				indexes: map[string]int{"addr1": 0},
			},
			addr:     "addr2",
			coins:    negCoins,
			expPanic: "cannot index and add invalid coin amount \"-1neg\"",
		},
		{
			name: "three addrs, add to first",
			receiver: &indexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("2two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
			addr:  "addr1",
			coins: coins("10one"),
			expected: &indexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("11one"), coins("2two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
		},
		{
			name: "three addrs, add to second",
			receiver: &indexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("2two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
			addr:  "addr2",
			coins: coins("10two"),
			expected: &indexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("12two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
		},
		{
			name: "three addrs, add to third",
			receiver: &indexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("2two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
			addr:  "addr3",
			coins: coins("10three"),
			expected: &indexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("2two"), coins("13three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
		},
		{
			name: "three addrs, add two coins to second",
			receiver: &indexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("2two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
			addr:  "addr2",
			coins: coins("10four,20two"),
			expected: &indexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("10four,22two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
		},
		{
			name: "three addrs, add to new",
			receiver: &indexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("2two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
			addr:  "good buddy",
			coins: coins("10four"),
			expected: &indexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3", "good buddy"},
				amts:    []sdk.Coins{coins("1one"), coins("2two"), coins("3three"), coins("10four")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2, "good buddy": 3},
			},
		},
		{
			name: "three addrs, add negative to second",
			receiver: &indexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("2two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
			addr:     "addr2",
			coins:    negCoins,
			expPanic: "cannot index and add invalid coin amount \"-1neg\"",
		},
		{
			name: "three addrs, add negative to new",
			receiver: &indexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("2two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
			addr:     "addr4",
			coins:    negCoins,
			expPanic: "cannot index and add invalid coin amount \"-1neg\"",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := copyIndexedAddrAmts(tc.receiver)
			defer func() {
				if t.Failed() {
					t.Logf("Original: %s", indexedAddrAmtsString(orig))
					t.Logf("  Actual: %s", indexedAddrAmtsString(tc.receiver))
					t.Logf("Expected: %s", indexedAddrAmtsString(tc.expected))
				}
			}()

			testFunc := func() {
				tc.receiver.add(tc.addr, tc.coins...)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "add(%q, %q)", tc.addr, tc.coins)
			if len(tc.expPanic) == 0 {
				assert.Equal(t, tc.expected, tc.receiver, "receiver after add(%q, %q)", tc.addr, tc.coins)
			}
		})
	}
}

func TestIndexedAddrAmts_GetAsInputs(t *testing.T) {
	coins := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "sdk.ParseCoinsNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name     string
		receiver *indexedAddrAmts
		expected []banktypes.Input
		expPanic string
	}{
		{name: "nil receiver", receiver: nil, expected: nil},
		{name: "no addrs", receiver: newIndexedAddrAmts(), expected: nil},
		{
			name: "one addr negative amount",
			receiver: &indexedAddrAmts{
				addrs: []string{"addr1"},
				amts:  []sdk.Coins{{{Denom: "neg", Amount: sdkmath.NewInt(-1)}}},
				indexes: map[string]int{
					"addr1": 0,
				},
			},
			expPanic: "invalid indexed amount \"addr1\" for address \"-1neg\": cannot be zero or negative",
		},
		{
			name: "one addr zero amount",
			receiver: &indexedAddrAmts{
				addrs: []string{"addr1"},
				amts:  []sdk.Coins{{{Denom: "zero", Amount: sdkmath.NewInt(0)}}},
				indexes: map[string]int{
					"addr1": 0,
				},
			},
			expPanic: "invalid indexed amount \"addr1\" for address \"0zero\": cannot be zero or negative",
		},
		{
			name: "one addr positive amount",
			receiver: &indexedAddrAmts{
				addrs: []string{"addr1"},
				amts:  []sdk.Coins{coins("1one")},
				indexes: map[string]int{
					"addr1": 0,
				},
			},
			expected: []banktypes.Input{
				{Address: "addr1", Coins: coins("1one")},
			},
		},
		{
			name: "two addrs",
			receiver: &indexedAddrAmts{
				addrs: []string{"addr1", "addr2"},
				amts:  []sdk.Coins{coins("1one"), coins("2two,3three")},
				indexes: map[string]int{
					"addr1": 0,
					"addr2": 1,
				},
			},
			expected: []banktypes.Input{
				{Address: "addr1", Coins: coins("1one")},
				{Address: "addr2", Coins: coins("2two,3three")},
			},
		},
		{
			name: "three addrs",
			receiver: &indexedAddrAmts{
				addrs: []string{"addr1", "addr2", "addr3"},
				amts:  []sdk.Coins{coins("1one"), coins("2two,3three"), coins("4four,5five,6six")},
				indexes: map[string]int{
					"addr1": 0,
					"addr2": 1,
					"addr3": 2,
				},
			},
			expected: []banktypes.Input{
				{Address: "addr1", Coins: coins("1one")},
				{Address: "addr2", Coins: coins("2two,3three")},
				{Address: "addr3", Coins: coins("4four,5five,6six")},
			},
		},
		{
			name: "three addrs, negative in third",
			receiver: &indexedAddrAmts{
				addrs: []string{"addr1", "addr2", "addr3"},
				amts: []sdk.Coins{
					coins("1one"),
					coins("2two,3three"),
					{
						{Denom: "acoin", Amount: sdkmath.NewInt(4)},
						{Denom: "bcoin", Amount: sdkmath.NewInt(5)},
						{Denom: "ncoin", Amount: sdkmath.NewInt(-6)},
						{Denom: "zcoin", Amount: sdkmath.NewInt(7)},
					},
				},
				indexes: map[string]int{
					"addr1": 0,
					"addr2": 1,
					"addr3": 2,
				},
			},
			expPanic: "invalid indexed amount \"addr3\" for address \"4acoin,5bcoin,-6ncoin,7zcoin\": cannot be zero or negative",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := copyIndexedAddrAmts(tc.receiver)
			var actual []banktypes.Input
			testFunc := func() {
				actual = tc.receiver.getAsInputs()
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "getAsInputs()")
			assert.Equal(t, tc.expected, actual, "getAsInputs() result")
			if !assert.Equal(t, orig, tc.receiver, "receiver before and after getAsInputs()") {
				t.Logf("Before: %s", indexedAddrAmtsString(orig))
				t.Logf(" After: %s", indexedAddrAmtsString(tc.receiver))
			}
		})
	}
}

func TestIndexedAddrAmts_GetAsOutputs(t *testing.T) {
	coins := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "sdk.ParseCoinsNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name     string
		receiver *indexedAddrAmts
		expected []banktypes.Output
		expPanic string
	}{
		{name: "nil receiver", receiver: nil, expected: nil},
		{name: "no addrs", receiver: newIndexedAddrAmts(), expected: nil},
		{
			name: "one addr negative amount",
			receiver: &indexedAddrAmts{
				addrs: []string{"addr1"},
				amts:  []sdk.Coins{{{Denom: "neg", Amount: sdkmath.NewInt(-1)}}},
				indexes: map[string]int{
					"addr1": 0,
				},
			},
			expPanic: "invalid indexed amount \"addr1\" for address \"-1neg\": cannot be zero or negative",
		},
		{
			name: "one addr zero amount",
			receiver: &indexedAddrAmts{
				addrs: []string{"addr1"},
				amts:  []sdk.Coins{{{Denom: "zero", Amount: sdkmath.NewInt(0)}}},
				indexes: map[string]int{
					"addr1": 0,
				},
			},
			expPanic: "invalid indexed amount \"addr1\" for address \"0zero\": cannot be zero or negative",
		},
		{
			name: "one addr positive amount",
			receiver: &indexedAddrAmts{
				addrs: []string{"addr1"},
				amts:  []sdk.Coins{coins("1one")},
				indexes: map[string]int{
					"addr1": 0,
				},
			},
			expected: []banktypes.Output{
				{Address: "addr1", Coins: coins("1one")},
			},
		},
		{
			name: "two addrs",
			receiver: &indexedAddrAmts{
				addrs: []string{"addr1", "addr2"},
				amts:  []sdk.Coins{coins("1one"), coins("2two,3three")},
				indexes: map[string]int{
					"addr1": 0,
					"addr2": 1,
				},
			},
			expected: []banktypes.Output{
				{Address: "addr1", Coins: coins("1one")},
				{Address: "addr2", Coins: coins("2two,3three")},
			},
		},
		{
			name: "three addrs",
			receiver: &indexedAddrAmts{
				addrs: []string{"addr1", "addr2", "addr3"},
				amts:  []sdk.Coins{coins("1one"), coins("2two,3three"), coins("4four,5five,6six")},
				indexes: map[string]int{
					"addr1": 0,
					"addr2": 1,
					"addr3": 2,
				},
			},
			expected: []banktypes.Output{
				{Address: "addr1", Coins: coins("1one")},
				{Address: "addr2", Coins: coins("2two,3three")},
				{Address: "addr3", Coins: coins("4four,5five,6six")},
			},
		},
		{
			name: "three addrs, negative in third",
			receiver: &indexedAddrAmts{
				addrs: []string{"addr1", "addr2", "addr3"},
				amts: []sdk.Coins{
					coins("1one"),
					coins("2two,3three"),
					{
						{Denom: "acoin", Amount: sdkmath.NewInt(4)},
						{Denom: "bcoin", Amount: sdkmath.NewInt(5)},
						{Denom: "ncoin", Amount: sdkmath.NewInt(-6)},
						{Denom: "zcoin", Amount: sdkmath.NewInt(7)},
					},
				},
				indexes: map[string]int{
					"addr1": 0,
					"addr2": 1,
					"addr3": 2,
				},
			},
			expPanic: "invalid indexed amount \"addr3\" for address \"4acoin,5bcoin,-6ncoin,7zcoin\": cannot be zero or negative",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := copyIndexedAddrAmts(tc.receiver)
			var actual []banktypes.Output
			testFunc := func() {
				actual = tc.receiver.getAsOutputs()
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "getAsOutputs()")
			assert.Equal(t, tc.expected, actual, "getAsOutputs() result")
			if !assert.Equal(t, orig, tc.receiver, "receiver before and after getAsInputs()") {
				t.Logf("Before: %s", indexedAddrAmtsString(orig))
				t.Logf(" After: %s", indexedAddrAmtsString(tc.receiver))
			}
		})
	}
}

// fulfillmentsString is similar to %v except with easier to understand Coin entries.
func fulfillmentsString(f *Fulfillments) string {
	if f == nil {
		return "nil"
	}

	fields := []string{
		fmt.Sprintf("AskOFs: %s", orderFulfillmentsString(f.AskOFs)),
		fmt.Sprintf("BidOFs: %s", orderFulfillmentsString(f.BidOFs)),
		fmt.Sprintf("PartialOrder: %s", partialFulfillmentString(f.PartialOrder)),
	}
	return fmt.Sprintf("{%s}", strings.Join(fields, ", "))
}

// orderFulfillmentsString is similar to %v except with easier to understand Coin entries.
func orderFulfillmentsString(ofs []*OrderFulfillment) string {
	if ofs == nil {
		return "nil"
	}

	vals := make([]string, len(ofs))
	for i, f := range ofs {
		vals[i] = orderFulfillmentString(f)
	}
	return fmt.Sprintf("[%s]", strings.Join(vals, ", "))
}

// partialFulfillmentString is similar to %v except with easier to understand Coin entries.
func partialFulfillmentString(p *PartialFulfillment) string {
	if p == nil {
		return "nil"
	}

	fields := []string{
		fmt.Sprintf("NewOrder:%s", orderString(p.NewOrder)),
		fmt.Sprintf("AssetsFilled:%s", coinPString(&p.AssetsFilled)),
		fmt.Sprintf("PriceFilled:%s", coinPString(&p.PriceFilled)),
	}
	return fmt.Sprintf("{%s}", strings.Join(fields, ", "))
}

// transferString is similar to %v except with easier to understand Coin entries.
func settlementTransfersString(s *SettlementTransfers) string {
	if s == nil {
		return "nil"
	}

	orderTransfers := "nil"
	if s.OrderTransfers != nil {
		transVals := make([]string, len(s.OrderTransfers))
		for i, trans := range s.OrderTransfers {
			transVals[i] = transferString(trans)
		}
		orderTransfers = fmt.Sprintf("[%s]", strings.Join(transVals, ", "))
	}

	feeInputs := "nil"
	if s.FeeInputs != nil {
		feeVals := make([]string, len(s.FeeInputs))
		for i, input := range s.FeeInputs {
			feeVals[i] = bankInputString(input)
		}
		feeInputs = fmt.Sprintf("[%s]", strings.Join(feeVals, ", "))
	}

	return fmt.Sprintf("{OrderTransfers:%s, FeeInputs:%s}", orderTransfers, feeInputs)
}

func TestBuildSettlementTransfers(t *testing.T) {
	coin := func(amt int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amt)}
	}
	igc := coin(2468, "ignorable") // igc => "ignorable coin"
	askOrder := func(orderID uint64, seller string, assets, price sdk.Coin) *Order {
		return NewOrder(orderID).WithAsk(&AskOrder{
			MarketId: 97531,
			Seller:   seller,
			Assets:   assets,
			Price:    price,
		})
	}
	bidOrder := func(orderID uint64, buyer string, assets, price sdk.Coin) *Order {
		return NewOrder(orderID).WithBid(&BidOrder{
			MarketId: 97531,
			Buyer:    buyer,
			Assets:   assets,
			Price:    price,
		})
	}
	askSplit := func(orderID uint64, seller string, assets, price sdk.Coin) *OrderSplit {
		return &OrderSplit{
			Order:  &OrderFulfillment{Order: askOrder(orderID, seller, igc, igc)},
			Assets: assets,
			Price:  price,
		}
	}
	bidSplit := func(orderID uint64, seller string, assets, price sdk.Coin) *OrderSplit {
		return &OrderSplit{
			Order:  &OrderFulfillment{Order: bidOrder(orderID, seller, igc, igc)},
			Assets: assets,
			Price:  price,
		}
	}
	input := func(addr string, coins ...sdk.Coin) banktypes.Input {
		return banktypes.Input{Address: addr, Coins: coins}
	}
	output := func(addr string, coins ...sdk.Coin) banktypes.Output {
		return banktypes.Output{Address: addr, Coins: coins}
	}

	tests := []struct {
		name     string
		f        *Fulfillments
		expected *SettlementTransfers
		expPanic string
	}{
		{
			name: "just an ask, no fees",
			f: &Fulfillments{
				AskOFs: []*OrderFulfillment{
					{
						Order:           askOrder(1, "seller", coin(2, "oasset"), coin(3, "oprice")),
						AssetsFilledAmt: sdkmath.NewInt(4),
						PriceAppliedAmt: sdkmath.NewInt(5),
						Splits: []*OrderSplit{
							bidSplit(6, "buyer", coin(7, "sasset"), coin(8, "sprice")),
							bidSplit(9, "buyer", coin(10, "sasset"), coin(11, "sprice")),
						},
						FeesToPay: nil,
					},
				},
			},
			expected: &SettlementTransfers{
				OrderTransfers: []*Transfer{
					{
						Inputs:  []banktypes.Input{input("seller", coin(4, "oasset"))},
						Outputs: []banktypes.Output{output("buyer", coin(17, "sasset"))},
					},
					{
						Inputs:  []banktypes.Input{input("buyer", coin(19, "sprice"))},
						Outputs: []banktypes.Output{output("seller", coin(5, "oprice"))},
					},
				},
				FeeInputs: nil,
			},
		},
		{
			name: "just an ask, with fees",
			f: &Fulfillments{
				AskOFs: []*OrderFulfillment{
					{
						Order:           askOrder(1, "seller", coin(2, "oasset"), coin(3, "oprice")),
						AssetsFilledAmt: sdkmath.NewInt(4),
						PriceAppliedAmt: sdkmath.NewInt(5),
						Splits: []*OrderSplit{
							bidSplit(6, "buyer", coin(7, "sasset"), coin(8, "sprice")),
							bidSplit(9, "buyer", coin(10, "sasset"), coin(11, "sprice")),
						},
						FeesToPay: sdk.NewCoins(coin(12, "feea"), coin(13, "feeb")),
					},
				},
			},
			expected: &SettlementTransfers{
				OrderTransfers: []*Transfer{
					{
						Inputs:  []banktypes.Input{input("seller", coin(4, "oasset"))},
						Outputs: []banktypes.Output{output("buyer", coin(17, "sasset"))},
					},
					{
						Inputs:  []banktypes.Input{input("buyer", coin(19, "sprice"))},
						Outputs: []banktypes.Output{output("seller", coin(5, "oprice"))},
					},
				},
				FeeInputs: []banktypes.Input{input("seller", coin(12, "feea"), coin(13, "feeb"))},
			},
		},
		{
			name: "just a bid, no fees",
			f: &Fulfillments{
				BidOFs: []*OrderFulfillment{
					{
						Order:           bidOrder(1, "buyer", coin(2, "oasset"), coin(3, "oprice")),
						AssetsFilledAmt: sdkmath.NewInt(4),
						PriceAppliedAmt: sdkmath.NewInt(5),
						Splits: []*OrderSplit{
							askSplit(6, "seller", coin(7, "sasset"), coin(8, "sprice")),
							askSplit(9, "seller", coin(10, "sasset"), coin(11, "sprice")),
						},
						FeesToPay: nil,
					},
				},
			},
			expected: &SettlementTransfers{
				OrderTransfers: []*Transfer{
					{
						Inputs:  []banktypes.Input{input("seller", coin(17, "sasset"))},
						Outputs: []banktypes.Output{output("buyer", coin(4, "oasset"))},
					},
					{
						Inputs:  []banktypes.Input{input("buyer", coin(5, "oprice"))},
						Outputs: []banktypes.Output{output("seller", coin(19, "sprice"))},
					},
				},
				FeeInputs: nil,
			},
		},
		{
			name: "just a bid, with fees",
			f: &Fulfillments{
				BidOFs: []*OrderFulfillment{
					{
						Order:           bidOrder(1, "buyer", coin(2, "oasset"), coin(3, "oprice")),
						AssetsFilledAmt: sdkmath.NewInt(4),
						PriceAppliedAmt: sdkmath.NewInt(5),
						Splits: []*OrderSplit{
							askSplit(6, "seller", coin(7, "sasset"), coin(8, "sprice")),
							askSplit(9, "seller", coin(10, "sasset"), coin(11, "sprice")),
						},
						FeesToPay: sdk.NewCoins(coin(12, "feea"), coin(13, "feeb")),
					},
				},
			},
			expected: &SettlementTransfers{
				OrderTransfers: []*Transfer{
					{
						Inputs:  []banktypes.Input{input("seller", coin(17, "sasset"))},
						Outputs: []banktypes.Output{output("buyer", coin(4, "oasset"))},
					},
					{
						Inputs:  []banktypes.Input{input("buyer", coin(5, "oprice"))},
						Outputs: []banktypes.Output{output("seller", coin(19, "sprice"))},
					},
				},
				FeeInputs: []banktypes.Input{input("buyer", coin(12, "feea"), coin(13, "feeb"))},
			},
		},
		{
			name: "two asks two bids",
			f: &Fulfillments{
				AskOFs: []*OrderFulfillment{
					{
						Order:           askOrder(1, "order seller", coin(2, "oasset"), coin(3, "oprice")),
						AssetsFilledAmt: sdkmath.NewInt(4),
						PriceAppliedAmt: sdkmath.NewInt(5),
						Splits: []*OrderSplit{
							bidSplit(6, "split buyer one", coin(7, "sasset"), coin(8, "sprice")),
							bidSplit(9, "split buyer two", coin(10, "sasset"), coin(11, "sprice")),
						},
						FeesToPay: sdk.NewCoins(coin(12, "sellfee")),
					},
					{
						Order:           askOrder(13, "order seller", coin(14, "oasset"), coin(15, "oprice")),
						AssetsFilledAmt: sdkmath.NewInt(16),
						PriceAppliedAmt: sdkmath.NewInt(17),
						Splits: []*OrderSplit{
							bidSplit(18, "split buyer one", coin(19, "sasset"), coin(20, "sprice")),
						},
						FeesToPay: sdk.NewCoins(coin(21, "sellfee")),
					},
				},
				BidOFs: []*OrderFulfillment{
					{
						Order:           bidOrder(22, "order buyer one", coin(23, "oasset"), coin(24, "oprice")),
						AssetsFilledAmt: sdkmath.NewInt(25),
						PriceAppliedAmt: sdkmath.NewInt(26),
						Splits: []*OrderSplit{
							askSplit(27, "split seller one", coin(28, "sasset"), coin(29, "sprice")),
							askSplit(30, "split seller one", coin(31, "sasset"), coin(32, "sprice")),
						},
						FeesToPay: sdk.NewCoins(coin(33, "buyfee")),
					},
					{
						Order:           bidOrder(34, "order buyer two", coin(35, "oasset"), coin(36, "oprice")),
						AssetsFilledAmt: sdkmath.NewInt(37),
						PriceAppliedAmt: sdkmath.NewInt(38),
						Splits: []*OrderSplit{
							askSplit(39, "split seller one", coin(40, "sasset"), coin(41, "sprice")),
						},
						FeesToPay: sdk.NewCoins(coin(42, "buyfee")),
					},
				},
			},
			expected: &SettlementTransfers{
				OrderTransfers: []*Transfer{
					{
						Inputs: []banktypes.Input{input("order seller", coin(4, "oasset"))},
						Outputs: []banktypes.Output{
							output("split buyer one", coin(7, "sasset")),
							output("split buyer two", coin(10, "sasset")),
						},
					},
					{
						Inputs: []banktypes.Input{
							input("split buyer one", coin(8, "sprice")),
							input("split buyer two", coin(11, "sprice")),
						},
						Outputs: []banktypes.Output{output("order seller", coin(5, "oprice"))},
					},
					{
						Inputs:  []banktypes.Input{input("order seller", coin(16, "oasset"))},
						Outputs: []banktypes.Output{output("split buyer one", coin(19, "sasset"))},
					},
					{
						Inputs:  []banktypes.Input{input("split buyer one", coin(20, "sprice"))},
						Outputs: []banktypes.Output{output("order seller", coin(17, "oprice"))},
					},
					{
						Inputs:  []banktypes.Input{input("split seller one", coin(59, "sasset"))},
						Outputs: []banktypes.Output{output("order buyer one", coin(25, "oasset"))},
					},
					{
						Inputs:  []banktypes.Input{input("order buyer one", coin(26, "oprice"))},
						Outputs: []banktypes.Output{output("split seller one", coin(61, "sprice"))},
					},
					{
						Inputs:  []banktypes.Input{input("split seller one", coin(40, "sasset"))},
						Outputs: []banktypes.Output{output("order buyer two", coin(37, "oasset"))},
					},
					{
						Inputs:  []banktypes.Input{input("order buyer two", coin(38, "oprice"))},
						Outputs: []banktypes.Output{output("split seller one", coin(41, "sprice"))},
					},
				},
				FeeInputs: []banktypes.Input{
					input("order seller", coin(33, "sellfee")),
					input("order buyer one", coin(33, "buyfee")),
					input("order buyer two", coin(42, "buyfee")),
				},
			},
		},
		{
			name: "negative ask asset",
			f: &Fulfillments{
				AskOFs: []*OrderFulfillment{
					{
						Order:           askOrder(1, "seller", coin(2, "oasset"), coin(3, "oprice")),
						AssetsFilledAmt: sdkmath.NewInt(-4),
						PriceAppliedAmt: sdkmath.NewInt(5),
						Splits: []*OrderSplit{
							bidSplit(6, "buyer", coin(7, "sasset"), coin(8, "sprice")),
							bidSplit(9, "buyer", coin(10, "sasset"), coin(11, "sprice")),
						},
					},
				},
			},
			expPanic: "invalid coin set -4oasset: coin -4oasset amount is not positive",
		},
		{
			name: "negative ask price",
			f: &Fulfillments{
				AskOFs: []*OrderFulfillment{
					{
						Order:           askOrder(1, "seller", coin(2, "oasset"), coin(3, "oprice")),
						AssetsFilledAmt: sdkmath.NewInt(4),
						PriceAppliedAmt: sdkmath.NewInt(-5),
						Splits: []*OrderSplit{
							bidSplit(6, "buyer", coin(7, "sasset"), coin(8, "sprice")),
							bidSplit(9, "buyer", coin(10, "sasset"), coin(11, "sprice")),
						},
					},
				},
			},
			expPanic: "invalid coin set -5oprice: coin -5oprice amount is not positive",
		},
		{
			name: "negative bid asset",
			f: &Fulfillments{
				BidOFs: []*OrderFulfillment{
					{
						Order:           bidOrder(1, "buyer", coin(2, "oasset"), coin(3, "oprice")),
						AssetsFilledAmt: sdkmath.NewInt(-4),
						PriceAppliedAmt: sdkmath.NewInt(5),
						Splits: []*OrderSplit{
							askSplit(6, "seller", coin(7, "sasset"), coin(8, "sprice")),
							askSplit(9, "seller", coin(10, "sasset"), coin(11, "sprice")),
						},
					},
				},
			},
			expPanic: "invalid coin set -4oasset: coin -4oasset amount is not positive",
		},
		{
			name: "negative bid price",
			f: &Fulfillments{
				BidOFs: []*OrderFulfillment{
					{
						Order:           bidOrder(1, "buyer", coin(2, "oasset"), coin(3, "oprice")),
						AssetsFilledAmt: sdkmath.NewInt(4),
						PriceAppliedAmt: sdkmath.NewInt(-5),
						Splits: []*OrderSplit{
							askSplit(6, "seller", coin(7, "sasset"), coin(8, "sprice")),
							askSplit(9, "seller", coin(10, "sasset"), coin(11, "sprice")),
						},
					},
				},
			},
			expPanic: "invalid coin set -5oprice: coin -5oprice amount is not positive",
		},
		{
			name: "ask with negative fees",
			f: &Fulfillments{
				BidOFs: []*OrderFulfillment{
					{
						Order:           bidOrder(1, "buyer", coin(2, "oasset"), coin(3, "oprice")),
						AssetsFilledAmt: sdkmath.NewInt(4),
						PriceAppliedAmt: sdkmath.NewInt(5),
						Splits: []*OrderSplit{
							askSplit(6, "seller", coin(7, "sasset"), coin(8, "sprice")),
							askSplit(9, "seller", coin(10, "sasset"), coin(11, "sprice")),
						},
						FeesToPay: sdk.Coins{coin(-12, "feecoin")},
					},
				},
			},
			expPanic: "invalid coin set -12feecoin: coin -12feecoin amount is not positive",
		},
		{
			name: "bid with negative fees",
			f: &Fulfillments{
				BidOFs: []*OrderFulfillment{
					{
						Order:           bidOrder(1, "buyer", coin(2, "oasset"), coin(3, "oprice")),
						AssetsFilledAmt: sdkmath.NewInt(4),
						PriceAppliedAmt: sdkmath.NewInt(5),
						Splits: []*OrderSplit{
							askSplit(6, "seller", coin(7, "sasset"), coin(8, "sprice")),
							askSplit(9, "seller", coin(10, "sasset"), coin(11, "sprice")),
						},
						FeesToPay: sdk.Coins{coin(-12, "feecoin")},
					},
				},
			},
			expPanic: "invalid coin set -12feecoin: coin -12feecoin amount is not positive",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual *SettlementTransfers
			defer func() {
				if t.Failed() {
					t.Logf("  Actual: %s", settlementTransfersString(actual))
					t.Logf("Expected: %s", settlementTransfersString(tc.expected))
					t.Logf("Fulfillments: %s", fulfillmentsString(tc.f))
				}
			}()
			testFunc := func() {
				actual = BuildSettlementTransfers(tc.f)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "BuildSettlementTransfers")
			assert.Equal(t, tc.expected, actual, "BuildSettlementTransfers result")
		})
	}
}

// transferString is similar to %v except with easier to understand Coin entries.
func transferString(t *Transfer) string {
	if t == nil {
		return "nil"
	}
	inputs := "nil"
	if t.Inputs != nil {
		inputVals := make([]string, len(t.Inputs))
		for i, input := range t.Inputs {
			inputVals[i] = bankInputString(input)
		}
		inputs = fmt.Sprintf("[%s]", strings.Join(inputVals, ", "))
	}
	outputs := "nil"
	if t.Outputs != nil {
		outputVals := make([]string, len(t.Outputs))
		for i, output := range t.Outputs {
			outputVals[i] = bankOutputString(output)
		}
		outputs = fmt.Sprintf("[%s]", strings.Join(outputVals, ", "))
	}
	return fmt.Sprintf("T{Inputs:%s, Outputs: %s}", inputs, outputs)
}

// bankInputString is similar to %v except with easier to understand Coin entries.
func bankInputString(i banktypes.Input) string {
	return fmt.Sprintf("I{Address:%q,Coins:%q}", i.Address, i.Coins)
}

// bankOutputString is similar to %v except with easier to understand Coin entries.
func bankOutputString(o banktypes.Output) string {
	return fmt.Sprintf("O{Address:%q,Coins:%q}", o.Address, o.Coins)
}

func TestGetAssetTransfer(t *testing.T) {
	coin := func(amt int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amt)}
	}
	igc := coin(2468, "ignorable") // igc => "ignorable coin"
	askOrder := func(orderID uint64, seller string, assets sdk.Coin) *Order {
		return NewOrder(orderID).WithAsk(&AskOrder{
			Seller: seller,
			Assets: assets,
		})
	}
	bidOrder := func(orderID uint64, buyer string, assets sdk.Coin) *Order {
		return NewOrder(orderID).WithBid(&BidOrder{
			Buyer:  buyer,
			Assets: assets,
		})
	}
	orderSplit := func(order *Order, assets sdk.Coin) *OrderSplit {
		return &OrderSplit{
			Order:  &OrderFulfillment{Order: order},
			Assets: assets,
			Price:  igc,
		}
	}
	input := func(addr string, coins ...sdk.Coin) banktypes.Input {
		return banktypes.Input{Address: addr, Coins: coins}
	}
	output := func(addr string, coins ...sdk.Coin) banktypes.Output {
		return banktypes.Output{Address: addr, Coins: coins}
	}

	tests := []struct {
		name     string
		f        *OrderFulfillment
		exp      *Transfer
		expPanic string
	}{
		{
			name: "ask, one split",
			f: &OrderFulfillment{
				Order:           askOrder(1, "seller", coin(25, "carrot")),
				AssetsFilledAmt: sdkmath.NewInt(33),
				Splits:          []*OrderSplit{orderSplit(bidOrder(2, "buyer", igc), coin(88, "banana"))},
			},
			exp: &Transfer{
				Inputs:  []banktypes.Input{input("seller", coin(33, "carrot"))},
				Outputs: []banktypes.Output{output("buyer", coin(88, "banana"))},
			},
		},
		{
			name: "ask, two splits diff addrs",
			f: &OrderFulfillment{
				Order:           askOrder(3, "SELLER", coin(26, "carrot")),
				AssetsFilledAmt: sdkmath.NewInt(4321),
				Splits: []*OrderSplit{
					orderSplit(bidOrder(4, "buyer 1", igc), coin(89, "banana")),
					orderSplit(bidOrder(5, "second buyer", igc), coin(45, "apple")),
				},
			},
			exp: &Transfer{
				Inputs: []banktypes.Input{input("SELLER", coin(4321, "carrot"))},
				Outputs: []banktypes.Output{
					output("buyer 1", coin(89, "banana")),
					output("second buyer", coin(45, "apple")),
				},
			},
		},
		{
			name: "ask, two splits same addr, two denoms",
			f: &OrderFulfillment{
				Order:           askOrder(6, "SeLleR", coin(27, "carrot")),
				AssetsFilledAmt: sdkmath.NewInt(5511),
				Splits: []*OrderSplit{
					orderSplit(bidOrder(7, "buyer", igc), coin(90, "banana")),
					orderSplit(bidOrder(8, "buyer", igc), coin(46, "apple")),
				},
			},
			exp: &Transfer{
				Inputs:  []banktypes.Input{input("SeLleR", coin(5511, "carrot"))},
				Outputs: []banktypes.Output{output("buyer", coin(46, "apple"), coin(90, "banana"))},
			},
		},
		{
			name: "ask, two splits same addr, one denom",
			f: &OrderFulfillment{
				Order:           askOrder(9, "sellsell", coin(28, "carrot")),
				AssetsFilledAmt: sdkmath.NewInt(42),
				Splits: []*OrderSplit{
					orderSplit(bidOrder(10, "buybuy", igc), coin(55, "apple")),
					orderSplit(bidOrder(11, "buybuy", igc), coin(34, "apple")),
				},
			},
			exp: &Transfer{
				Inputs:  []banktypes.Input{input("sellsell", coin(42, "carrot"))},
				Outputs: []banktypes.Output{output("buybuy", coin(89, "apple"))},
			},
		},
		{
			name: "ask, negative price in split",
			f: &OrderFulfillment{
				Order:           askOrder(12, "goodsell", coin(29, "carrot")),
				AssetsFilledAmt: sdkmath.NewInt(91),
				Splits:          []*OrderSplit{orderSplit(bidOrder(13, "buygood", igc), coin(-4, "banana"))},
			},
			expPanic: "cannot index and add invalid coin amount \"-4banana\"",
		},
		{
			name: "ask, negative price applied",
			f: &OrderFulfillment{
				Order:           askOrder(14, "solong", coin(30, "carrot")),
				AssetsFilledAmt: sdkmath.NewInt(-5),
				Splits:          []*OrderSplit{orderSplit(bidOrder(15, "hello", igc), coin(66, "banana"))},
			},
			expPanic: "invalid coin set -5carrot: coin -5carrot amount is not positive",
		},

		{
			name: "bid, one split",
			f: &OrderFulfillment{
				Order:           bidOrder(1, "buyer", coin(25, "carrot")),
				AssetsFilledAmt: sdkmath.NewInt(33),
				Splits:          []*OrderSplit{orderSplit(askOrder(2, "seller", igc), coin(88, "banana"))},
			},
			exp: &Transfer{
				Inputs:  []banktypes.Input{input("seller", coin(88, "banana"))},
				Outputs: []banktypes.Output{output("buyer", coin(33, "carrot"))},
			},
		},
		{
			name: "bid, two splits diff addrs",
			f: &OrderFulfillment{
				Order:           bidOrder(3, "BUYER", coin(26, "carrot")),
				AssetsFilledAmt: sdkmath.NewInt(4321),
				Splits: []*OrderSplit{
					orderSplit(askOrder(4, "seller 1", igc), coin(89, "banana")),
					orderSplit(askOrder(5, "second seller", igc), coin(45, "apple")),
				},
			},
			exp: &Transfer{
				Inputs: []banktypes.Input{
					input("seller 1", coin(89, "banana")),
					input("second seller", coin(45, "apple")),
				},
				Outputs: []banktypes.Output{output("BUYER", coin(4321, "carrot"))},
			},
		},
		{
			name: "bid, two splits same addr, two denoms",
			f: &OrderFulfillment{
				Order:           bidOrder(6, "BuYeR", coin(27, "carrot")),
				AssetsFilledAmt: sdkmath.NewInt(5511),
				Splits: []*OrderSplit{
					orderSplit(askOrder(7, "seller", igc), coin(90, "banana")),
					orderSplit(askOrder(8, "seller", igc), coin(46, "apple")),
				},
			},
			exp: &Transfer{
				Inputs:  []banktypes.Input{input("seller", coin(46, "apple"), coin(90, "banana"))},
				Outputs: []banktypes.Output{output("BuYeR", coin(5511, "carrot"))},
			},
		},
		{
			name: "bid, two splits same addr, one denom",
			f: &OrderFulfillment{
				Order:           bidOrder(9, "buybuy", coin(28, "carrot")),
				AssetsFilledAmt: sdkmath.NewInt(42),
				Splits: []*OrderSplit{
					orderSplit(bidOrder(10, "sellsell", igc), coin(55, "apple")),
					orderSplit(bidOrder(11, "sellsell", igc), coin(34, "apple")),
				},
			},
			exp: &Transfer{
				Inputs:  []banktypes.Input{input("sellsell", coin(89, "apple"))},
				Outputs: []banktypes.Output{output("buybuy", coin(42, "carrot"))},
			},
		},
		{
			name: "bid, negative price in split",
			f: &OrderFulfillment{
				Order:           bidOrder(12, "goodbuy", coin(29, "carrot")),
				AssetsFilledAmt: sdkmath.NewInt(91),
				Splits:          []*OrderSplit{orderSplit(askOrder(13, "sellgood", igc), coin(-4, "banana"))},
			},
			expPanic: "cannot index and add invalid coin amount \"-4banana\"",
		},
		{
			name: "bid, negative price applied",
			f: &OrderFulfillment{
				Order:           bidOrder(14, "heythere", coin(30, "carrot")),
				AssetsFilledAmt: sdkmath.NewInt(-5),
				Splits:          []*OrderSplit{orderSplit(askOrder(15, "afterwhile", igc), coin(66, "banana"))},
			},
			expPanic: "invalid coin set -5carrot: coin -5carrot amount is not positive",
		},

		{
			name:     "nil inside order",
			f:        &OrderFulfillment{Order: NewOrder(20)},
			expPanic: "unknown order type <nil>",
		},
		{
			name:     "unknown inside order",
			f:        &OrderFulfillment{Order: newUnknownOrder(21)},
			expPanic: "unknown order type *exchange.unknownOrderType",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual *Transfer
			defer func() {
				if t.Failed() {
					t.Logf("  Actual: %s", transferString(actual))
					t.Logf("Expected: %s", transferString(tc.exp))
					t.Logf("OrderFulfillment: %s", orderFulfillmentString(tc.f))
				}
			}()
			testFunc := func() {
				actual = GetAssetTransfer(tc.f)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "GetAssetTransfer")
			assert.Equal(t, tc.exp, actual, "GetAssetTransfer result")
		})
	}
}

func TestGetPriceTransfer(t *testing.T) {
	coin := func(amt int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amt)}
	}
	igc := coin(2468, "ignorable") // igc => "ignorable coin"
	askOrder := func(orderID uint64, seller string, price sdk.Coin) *Order {
		return NewOrder(orderID).WithAsk(&AskOrder{
			Seller: seller,
			Price:  price,
		})
	}
	bidOrder := func(orderID uint64, buyer string, price sdk.Coin) *Order {
		return NewOrder(orderID).WithBid(&BidOrder{
			Buyer: buyer,
			Price: price,
		})
	}
	orderSplit := func(order *Order, price sdk.Coin) *OrderSplit {
		return &OrderSplit{
			Order:  &OrderFulfillment{Order: order},
			Price:  price,
			Assets: igc,
		}
	}
	input := func(addr string, coins ...sdk.Coin) banktypes.Input {
		return banktypes.Input{Address: addr, Coins: coins}
	}
	output := func(addr string, coins ...sdk.Coin) banktypes.Output {
		return banktypes.Output{Address: addr, Coins: coins}
	}

	tests := []struct {
		name     string
		f        *OrderFulfillment
		exp      *Transfer
		expPanic string
	}{
		{
			name: "ask, one split",
			f: &OrderFulfillment{
				Order:           askOrder(1, "seller", coin(25, "carrot")),
				PriceAppliedAmt: sdkmath.NewInt(33),
				Splits:          []*OrderSplit{orderSplit(bidOrder(2, "buyer", igc), coin(88, "banana"))},
			},
			exp: &Transfer{
				Inputs:  []banktypes.Input{input("buyer", coin(88, "banana"))},
				Outputs: []banktypes.Output{output("seller", coin(33, "carrot"))},
			},
		},
		{
			name: "ask, two splits diff addrs",
			f: &OrderFulfillment{
				Order:           askOrder(3, "SELLER", coin(26, "carrot")),
				PriceAppliedAmt: sdkmath.NewInt(4321),
				Splits: []*OrderSplit{
					orderSplit(bidOrder(4, "buyer 1", igc), coin(89, "banana")),
					orderSplit(bidOrder(5, "second buyer", igc), coin(45, "apple")),
				},
			},
			exp: &Transfer{
				Inputs: []banktypes.Input{
					input("buyer 1", coin(89, "banana")),
					input("second buyer", coin(45, "apple")),
				},
				Outputs: []banktypes.Output{output("SELLER", coin(4321, "carrot"))},
			},
		},
		{
			name: "ask, two splits same addr, two denoms",
			f: &OrderFulfillment{
				Order:           askOrder(6, "SeLleR", coin(27, "carrot")),
				PriceAppliedAmt: sdkmath.NewInt(5511),
				Splits: []*OrderSplit{
					orderSplit(bidOrder(7, "buyer", igc), coin(90, "banana")),
					orderSplit(bidOrder(8, "buyer", igc), coin(46, "apple")),
				},
			},
			exp: &Transfer{
				Inputs:  []banktypes.Input{input("buyer", coin(46, "apple"), coin(90, "banana"))},
				Outputs: []banktypes.Output{output("SeLleR", coin(5511, "carrot"))},
			},
		},
		{
			name: "ask, two splits same addr, one denom",
			f: &OrderFulfillment{
				Order:           askOrder(9, "sellsell", coin(28, "carrot")),
				PriceAppliedAmt: sdkmath.NewInt(42),
				Splits: []*OrderSplit{
					orderSplit(bidOrder(10, "buybuy", igc), coin(55, "apple")),
					orderSplit(bidOrder(11, "buybuy", igc), coin(34, "apple")),
				},
			},
			exp: &Transfer{
				Inputs:  []banktypes.Input{input("buybuy", coin(89, "apple"))},
				Outputs: []banktypes.Output{output("sellsell", coin(42, "carrot"))},
			},
		},
		{
			name: "ask, negative price in split",
			f: &OrderFulfillment{
				Order:           askOrder(12, "goodsell", coin(29, "carrot")),
				PriceAppliedAmt: sdkmath.NewInt(91),
				Splits:          []*OrderSplit{orderSplit(bidOrder(13, "buygood", igc), coin(-4, "banana"))},
			},
			expPanic: "cannot index and add invalid coin amount \"-4banana\"",
		},
		{
			name: "ask, negative price applied",
			f: &OrderFulfillment{
				Order:           askOrder(14, "solong", coin(30, "carrot")),
				PriceAppliedAmt: sdkmath.NewInt(-5),
				Splits:          []*OrderSplit{orderSplit(bidOrder(15, "hello", igc), coin(66, "banana"))},
			},
			expPanic: "invalid coin set -5carrot: coin -5carrot amount is not positive",
		},

		{
			name: "bid, one split",
			f: &OrderFulfillment{
				Order:           bidOrder(1, "buyer", coin(25, "carrot")),
				PriceAppliedAmt: sdkmath.NewInt(33),
				Splits:          []*OrderSplit{orderSplit(askOrder(2, "seller", igc), coin(88, "banana"))},
			},
			exp: &Transfer{
				Inputs:  []banktypes.Input{input("buyer", coin(33, "carrot"))},
				Outputs: []banktypes.Output{output("seller", coin(88, "banana"))},
			},
		},
		{
			name: "bid, two splits diff addrs",
			f: &OrderFulfillment{
				Order:           bidOrder(3, "BUYER", coin(26, "carrot")),
				PriceAppliedAmt: sdkmath.NewInt(4321),
				Splits: []*OrderSplit{
					orderSplit(askOrder(4, "seller 1", igc), coin(89, "banana")),
					orderSplit(askOrder(5, "second seller", igc), coin(45, "apple")),
				},
			},
			exp: &Transfer{
				Inputs: []banktypes.Input{input("BUYER", coin(4321, "carrot"))},
				Outputs: []banktypes.Output{
					output("seller 1", coin(89, "banana")),
					output("second seller", coin(45, "apple")),
				},
			},
		},
		{
			name: "bid, two splits same addr, two denoms",
			f: &OrderFulfillment{
				Order:           bidOrder(6, "BuYeR", coin(27, "carrot")),
				PriceAppliedAmt: sdkmath.NewInt(5511),
				Splits: []*OrderSplit{
					orderSplit(askOrder(7, "seller", igc), coin(90, "banana")),
					orderSplit(askOrder(8, "seller", igc), coin(46, "apple")),
				},
			},
			exp: &Transfer{
				Inputs:  []banktypes.Input{input("BuYeR", coin(5511, "carrot"))},
				Outputs: []banktypes.Output{output("seller", coin(46, "apple"), coin(90, "banana"))},
			},
		},
		{
			name: "bid, two splits same addr, one denom",
			f: &OrderFulfillment{
				Order:           bidOrder(9, "buybuy", coin(28, "carrot")),
				PriceAppliedAmt: sdkmath.NewInt(42),
				Splits: []*OrderSplit{
					orderSplit(bidOrder(10, "sellsell", igc), coin(55, "apple")),
					orderSplit(bidOrder(11, "sellsell", igc), coin(34, "apple")),
				},
			},
			exp: &Transfer{
				Inputs:  []banktypes.Input{input("buybuy", coin(42, "carrot"))},
				Outputs: []banktypes.Output{output("sellsell", coin(89, "apple"))},
			},
		},
		{
			name: "bid, negative price in split",
			f: &OrderFulfillment{
				Order:           bidOrder(12, "goodbuy", coin(29, "carrot")),
				PriceAppliedAmt: sdkmath.NewInt(91),
				Splits:          []*OrderSplit{orderSplit(askOrder(13, "sellgood", igc), coin(-4, "banana"))},
			},
			expPanic: "cannot index and add invalid coin amount \"-4banana\"",
		},
		{
			name: "bid, negative price applied",
			f: &OrderFulfillment{
				Order:           bidOrder(14, "heythere", coin(30, "carrot")),
				PriceAppliedAmt: sdkmath.NewInt(-5),
				Splits:          []*OrderSplit{orderSplit(askOrder(15, "afterwhile", igc), coin(66, "banana"))},
			},
			expPanic: "invalid coin set -5carrot: coin -5carrot amount is not positive",
		},

		{
			name:     "nil inside order",
			f:        &OrderFulfillment{Order: NewOrder(20)},
			expPanic: "unknown order type <nil>",
		},
		{
			name:     "unknown inside order",
			f:        &OrderFulfillment{Order: newUnknownOrder(21)},
			expPanic: "unknown order type *exchange.unknownOrderType",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual *Transfer
			defer func() {
				if t.Failed() {
					t.Logf("  Actual: %s", transferString(actual))
					t.Logf("Expected: %s", transferString(tc.exp))
					t.Logf("OrderFulfillment: %s", orderFulfillmentString(tc.f))
				}
			}()
			testFunc := func() {
				actual = GetPriceTransfer(tc.f)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "GetPriceTransfer")
			assert.Equal(t, tc.exp, actual, "GetPriceTransfer result")
		})
	}
}
