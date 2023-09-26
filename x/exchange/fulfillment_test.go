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
		fmt.Sprintf("AssetsFilledAmt:%q", f.AssetsFilledAmt),
		fmt.Sprintf("AssetsUnfilledAmt:%q", f.AssetsUnfilledAmt),
		fmt.Sprintf("PriceAppliedAmt:%q", f.PriceAppliedAmt),
		fmt.Sprintf("PriceLeftAmt:%q", f.PriceLeftAmt),
		fmt.Sprintf("IsFinalized:%t", f.IsFinalized),
		fmt.Sprintf("FeesToPay:%s", coinsString(f.FeesToPay)),
		fmt.Sprintf("OrderFeesLeft:%s", coinsString(f.OrderFeesLeft)),
		fmt.Sprintf("PriceFilledAmt:%q", f.PriceFilledAmt),
		fmt.Sprintf("PriceUnfilledAmt:%q", f.PriceUnfilledAmt),
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
			expPanic: "unknown sub-order type <nil>: does not implement SubOrderI",
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
		{name: "unknown", f: OrderFulfillment{Order: &Order{Order: &unknownOrderType{}}}, exp: false},
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
		{name: "unknown", f: OrderFulfillment{Order: &Order{Order: &unknownOrderType{}}}, exp: false},
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
		{name: "unknown", f: OrderFulfillment{Order: &Order{Order: &unknownOrderType{}}}, exp: "*exchange.unknownOrderType"},
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

// TODO[1658]: func TestOrderFulfillment_Apply(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_ApplyLeftoverPrice(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_Finalize(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_Validate(t *testing.T)

// TODO[1658]: func TestFulfill(t *testing.T)

// TODO[1658]: func TestGetFulfillmentAssetsAmt(t *testing.T)

// TODO[1658]: func TestNewPartialFulfillment(t *testing.T)

// TODO[1658]: func TestBuildFulfillments(t *testing.T)

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
			f:        &OrderFulfillment{Order: &Order{OrderId: 21, Order: &unknownOrderType{}}},
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
			f:        &OrderFulfillment{Order: &Order{OrderId: 21, Order: &unknownOrderType{}}},
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
