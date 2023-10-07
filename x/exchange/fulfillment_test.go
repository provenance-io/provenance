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

// copySlice copies a slice using the provided copier for each entry.
func copySlice[T any](vals []T, copier func(T) T) []T {
	if vals == nil {
		return nil
	}
	rv := make([]T, len(vals))
	for i, v := range vals {
		rv[i] = copier(v)
	}
	return rv
}

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

// copyDistribution copies a distribution.
func copyDistribution(dist *Distribution) *Distribution {
	if dist == nil {
		return nil
	}

	return &Distribution{
		Address: dist.Address,
		Amount:  copySDKInt(dist.Amount),
	}
}

// copyDistributions copies a slice of distributions.
func copyDistributions(dists []*Distribution) []*Distribution {
	return copySlice(dists, copyDistribution)
}

// copyOrderFulfillment returns a deep copy of an order fulfillment.
func copyOrderFulfillment(f *OrderFulfillment) *OrderFulfillment {
	if f == nil {
		return nil
	}

	return &OrderFulfillment{
		Order:             copyOrder(f.Order),
		Splits:            copyOrderSplits(f.Splits),
		AssetDists:        copyDistributions(f.AssetDists),
		PriceDists:        copyDistributions(f.PriceDists),
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

// copyOrderFulfillments returns a deep copy of a slice of order fulfillments.
func copyOrderFulfillments(fs []*OrderFulfillment) []*OrderFulfillment {
	return copySlice(fs, copyOrderFulfillment)
}

// copyInput returns a deep copy of a bank input.
func copyInput(input banktypes.Input) banktypes.Input {
	return banktypes.Input{
		Address: input.Address,
		Coins:   copyCoins(input.Coins),
	}
}

// copyInputs returns a deep copy of a slice of bank inputs.
func copyInputs(inputs []banktypes.Input) []banktypes.Input {
	return copySlice(inputs, copyInput)
}

// copyOutput returns a deep copy of a bank output.
func copyOutput(output banktypes.Output) banktypes.Output {
	return banktypes.Output{
		Address: output.Address,
		Coins:   copyCoins(output.Coins),
	}
}

// copyOutputs returns a deep copy of a slice of bank outputs.
func copyOutputs(outputs []banktypes.Output) []banktypes.Output {
	return copySlice(outputs, copyOutput)
}

// copyTransfer returns a deep copy of a transfer.
func copyTransfer(t *Transfer) *Transfer {
	if t == nil {
		return nil
	}
	return &Transfer{
		Inputs:  copyInputs(t.Inputs),
		Outputs: copyOutputs(t.Outputs),
	}
}

// copyTransfers returns a deep copy of a slice of transfers.
func copyTransfers(ts []*Transfer) []*Transfer {
	return copySlice(ts, copyTransfer)
}

// copyFilledOrder returns a deep copy of a filled order.
func copyFilledOrder(f *FilledOrder) *FilledOrder {
	if f == nil {
		return nil
	}
	return &FilledOrder{
		order:       copyOrder(f.order),
		actualPrice: copyCoin(f.actualPrice),
		actualFees:  copyCoins(f.actualFees),
	}
}

// copyFilledOrders returns a deep copy of a slice of filled order.
func copyFilledOrders(fs []*FilledOrder) []*FilledOrder {
	return copySlice(fs, copyFilledOrder)
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

// distributionString is similar to %v except with easier to understand Int entries.
func distributionString(dist *Distribution) string {
	if dist == nil {
		return "nil"
	}
	return fmt.Sprintf("{Address:%q, Amount:%s}", dist.Address, dist.Amount)
}

// distributionsString is similar to %v except with easier to understand Int entries.
func distributionsString(dists []*Distribution) string {
	if dists == nil {
		return "nil"
	}
	vals := make([]string, len(dists))
	for i, d := range dists {
		vals[i] = distributionString(d)
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
		fmt.Sprintf("AssetDists:%s", distributionsString(f.AssetDists)),
		fmt.Sprintf("PriceDists:%s", distributionsString(f.PriceDists)),
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

// assertEqualOrderFulfillments asserts that the two order fulfillments are equal.
// Returns true if equal.
// If not equal, and neither are nil, equality on each field is also asserted in order to help identify the problem.
func assertEqualOrderFulfillments(t *testing.T, expected, actual *OrderFulfillment, message string, args ...interface{}) bool {
	t.Helper()
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
	assert.Equalf(t, expected.AssetDists, actual.AssetDists, msg("OrderFulfillment.AssetDists"), args...)
	assert.Equalf(t, expected.PriceDists, actual.PriceDists, msg("OrderFulfillment.PriceDists"), args...)
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

// assertEqualOrderFulfillmentSlices asserts that the two order fulfillments are equal.
// Returns true if equal.
// If not equal, and neither are nil, equality on each field is also asserted in order to help identify the problem.
func assertEqualOrderFulfillmentSlices(t *testing.T, expected, actual []*OrderFulfillment, message string, args ...interface{}) bool {
	t.Helper()
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

	// Check the order ids (and lengths) since that's gonna be a huge clue to a problem
	expIDs := make([]string, len(expected))
	for i, exp := range expected {
		expIDs[i] = fmt.Sprintf("%d", exp.GetOrderID())
	}
	actIDs := make([]string, len(actual))
	for i, act := range actual {
		actIDs[i] = fmt.Sprintf("%d", act.GetOrderID())
	}
	if !assert.Equalf(t, expIDs, actIDs, msg("OrderIDs"), args...) {
		// Wooo, should have actionable info in the failure, so we can be done.
		return false
	}

	// Try the comparisons as strings, one per line because that's easier with ints and coins.
	expStrVals := make([]string, len(expected))
	for i, exp := range expected {
		expStrVals[i] = orderFulfillmentString(exp)
	}
	expStrs := strings.Join(expStrVals, "\n")
	actStrVals := make([]string, len(actual))
	for i, act := range actual {
		actStrVals[i] = orderFulfillmentString(act)
	}
	actStrs := strings.Join(actStrVals, "\n")
	if !assert.Equalf(t, expStrs, actStrs, msg("OrderFulfillment strings"), args...) {
		// Wooo, should have actionable info in the failure, so we can be done.
		return false
	}

	// Alright, do it the hard way one at a time.
	for i := range expected {
		assertEqualOrderFulfillments(t, expected[i], actual[i], fmt.Sprintf("[%d]%s", i, message), args...)
	}
	t.Logf("  Actual: %s", orderFulfillmentsString(actual))
	t.Logf("Expected: %s", orderFulfillmentsString(expected))
	return false
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

func TestNewOrderFulfillments(t *testing.T) {
	assetCoin := func(amount int64) sdk.Coin {
		return sdk.NewInt64Coin("anise", amount)
	}
	priceCoin := func(amount int64) sdk.Coin {
		return sdk.NewInt64Coin("paprika", amount)
	}
	feeCoin := func(amount int64) *sdk.Coin {
		rv := sdk.NewInt64Coin("fennel", amount)
		return &rv
	}

	askOrders := make([]*Order, 4) // ids 1, 2, 3, 4
	for j := range askOrders {
		i := int64(j) + 1
		order := &AskOrder{
			MarketId: uint32(90 + i),
			Seller:   fmt.Sprintf("seller-%d", i),
			Assets:   assetCoin(1000*i + 100*i + 10*i + i),
			Price:    priceCoin(100*i + 10*i + i),
		}
		if j%2 == 0 {
			order.SellerSettlementFlatFee = feeCoin(10*i + i)
		}
		if j >= 2 {
			order.AllowPartial = true
		}
		askOrders[j] = NewOrder(uint64(i)).WithAsk(order)
	}

	bidOrders := make([]*Order, 4) // ids 5, 6, 7, 8
	for j := range bidOrders {
		i := int64(j + 5)
		order := &BidOrder{
			MarketId: uint32(90 + i),
			Buyer:    fmt.Sprintf("buyer-%d", i),
			Assets:   assetCoin(1000*i + 100*i + 10*i + i),
			Price:    priceCoin(100*i + 10*i + i),
		}
		switch j {
		case 0:
			order.BuyerSettlementFees = sdk.Coins{*feeCoin(10*i + i)}
		case 2:
			order.BuyerSettlementFees = sdk.Coins{
				*feeCoin(10*i + i),
				sdk.NewInt64Coin("garlic", 10000*i+1000*i+100*i+10*i+i),
			}
		}
		if j >= 2 {
			order.AllowPartial = true
		}
		bidOrders[j] = NewOrder(uint64(i)).WithBid(order)
	}

	tests := []struct {
		name     string
		orders   []*Order
		expected []*OrderFulfillment
	}{
		{
			name:     "nil orders",
			orders:   nil,
			expected: []*OrderFulfillment{},
		},
		{
			name:     "empty orders",
			orders:   []*Order{},
			expected: []*OrderFulfillment{},
		},
		{
			name:     "1 ask order",
			orders:   []*Order{askOrders[0]},
			expected: []*OrderFulfillment{NewOrderFulfillment(askOrders[0])},
		},
		{
			name:     "1 bid order",
			orders:   []*Order{bidOrders[0]},
			expected: []*OrderFulfillment{NewOrderFulfillment(bidOrders[0])},
		},
		{
			name:   "4 ask orders",
			orders: []*Order{askOrders[0], askOrders[1], askOrders[2], askOrders[3]},
			expected: []*OrderFulfillment{
				NewOrderFulfillment(askOrders[0]),
				NewOrderFulfillment(askOrders[1]),
				NewOrderFulfillment(askOrders[2]),
				NewOrderFulfillment(askOrders[3]),
			},
		},
		{
			name:   "4 bid orders",
			orders: []*Order{bidOrders[0], bidOrders[1], bidOrders[2], bidOrders[3]},
			expected: []*OrderFulfillment{
				NewOrderFulfillment(bidOrders[0]),
				NewOrderFulfillment(bidOrders[1]),
				NewOrderFulfillment(bidOrders[2]),
				NewOrderFulfillment(bidOrders[3]),
			},
		},
		{
			name:   "1 bid 1 ask",
			orders: []*Order{askOrders[1], bidOrders[2]},
			expected: []*OrderFulfillment{
				NewOrderFulfillment(askOrders[1]),
				NewOrderFulfillment(bidOrders[2]),
			},
		},
		{
			name:   "1 ask 1 bid",
			orders: []*Order{bidOrders[1], askOrders[2]},
			expected: []*OrderFulfillment{
				NewOrderFulfillment(bidOrders[1]),
				NewOrderFulfillment(askOrders[2]),
			},
		},
		{
			name: "4 asks 4 bids",
			orders: []*Order{
				askOrders[0], askOrders[1], askOrders[2], askOrders[3],
				bidOrders[3], bidOrders[2], bidOrders[1], bidOrders[0],
			},
			expected: []*OrderFulfillment{
				NewOrderFulfillment(askOrders[0]),
				NewOrderFulfillment(askOrders[1]),
				NewOrderFulfillment(askOrders[2]),
				NewOrderFulfillment(askOrders[3]),
				NewOrderFulfillment(bidOrders[3]),
				NewOrderFulfillment(bidOrders[2]),
				NewOrderFulfillment(bidOrders[1]),
				NewOrderFulfillment(bidOrders[0]),
			},
		},
		{
			name: "4 bids 4 asks",
			orders: []*Order{
				bidOrders[0], bidOrders[1], bidOrders[2], bidOrders[3],
				askOrders[3], askOrders[2], askOrders[1], askOrders[0],
			},
			expected: []*OrderFulfillment{
				NewOrderFulfillment(bidOrders[0]),
				NewOrderFulfillment(bidOrders[1]),
				NewOrderFulfillment(bidOrders[2]),
				NewOrderFulfillment(bidOrders[3]),
				NewOrderFulfillment(askOrders[3]),
				NewOrderFulfillment(askOrders[2]),
				NewOrderFulfillment(askOrders[1]),
				NewOrderFulfillment(askOrders[0]),
			},
		},
		{
			name: "interweaved 4 asks 4 bids",
			orders: []*Order{
				bidOrders[3], askOrders[0], askOrders[3], bidOrders[1],
				bidOrders[0], askOrders[1], bidOrders[2], askOrders[2],
			},
			expected: []*OrderFulfillment{
				NewOrderFulfillment(bidOrders[3]),
				NewOrderFulfillment(askOrders[0]),
				NewOrderFulfillment(askOrders[3]),
				NewOrderFulfillment(bidOrders[1]),
				NewOrderFulfillment(bidOrders[0]),
				NewOrderFulfillment(askOrders[1]),
				NewOrderFulfillment(bidOrders[2]),
				NewOrderFulfillment(askOrders[2]),
			},
		},
		{
			name: "duplicated entries",
			orders: []*Order{
				askOrders[3], bidOrders[2], askOrders[3], bidOrders[2],
			},
			expected: []*OrderFulfillment{
				NewOrderFulfillment(askOrders[3]),
				NewOrderFulfillment(bidOrders[2]),
				NewOrderFulfillment(askOrders[3]),
				NewOrderFulfillment(bidOrders[2]),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []*OrderFulfillment
			testFunc := func() {
				actual = NewOrderFulfillments(tc.orders)
			}
			require.NotPanics(t, testFunc, "NewOrderFulfillments")
			assertEqualOrderFulfillmentSlices(t, tc.expected, actual, "NewOrderFulfillments result")
		})
	}
}

func TestOrderFulfillment_AssetCoin(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	bigCoin := func(amount, denom string) sdk.Coin {
		amt := newInt(t, amount)
		return sdk.Coin{Denom: denom, Amount: amt}
	}
	askOrder := func(orderID uint64, assets, price sdk.Coin) *Order {
		return NewOrder(orderID).WithAsk(&AskOrder{
			Assets: assets,
			Price:  price,
		})
	}
	bidOrder := func(orderID uint64, assets, price sdk.Coin) *Order {
		return NewOrder(orderID).WithBid(&BidOrder{
			Assets: assets,
			Price:  price,
		})
	}

	tests := []struct {
		name     string
		receiver OrderFulfillment
		amt      sdkmath.Int
		expected sdk.Coin
		expPanic string
	}{
		{
			name:     "nil order",
			receiver: OrderFulfillment{Order: nil},
			amt:      sdkmath.NewInt(0),
			expPanic: "runtime error: invalid memory address or nil pointer dereference",
		},
		{
			name:     "nil inside order",
			receiver: OrderFulfillment{Order: NewOrder(1)},
			amt:      sdkmath.NewInt(0),
			expPanic: nilSubTypeErr(1),
		},
		{
			name:     "unknown inside order",
			receiver: OrderFulfillment{Order: newUnknownOrder(2)},
			amt:      sdkmath.NewInt(0),
			expPanic: unknownSubTypeErr(2),
		},
		{
			name:     "ask order",
			receiver: OrderFulfillment{Order: askOrder(3, coin(4, "apple"), coin(5, "plum"))},
			amt:      sdkmath.NewInt(6),
			expected: coin(6, "apple"),
		},
		{
			name:     "ask order with negative assets",
			receiver: OrderFulfillment{Order: askOrder(7, coin(-8, "apple"), coin(9, "plum"))},
			amt:      sdkmath.NewInt(10),
			expected: coin(10, "apple"),
		},
		{
			name:     "ask order, negative amt",
			receiver: OrderFulfillment{Order: askOrder(11, coin(12, "apple"), coin(13, "plum"))},
			amt:      sdkmath.NewInt(-14),
			expected: coin(-14, "apple"),
		},
		{
			name:     "ask order with negative assets, negative amt",
			receiver: OrderFulfillment{Order: askOrder(15, coin(-16, "apple"), coin(17, "plum"))},
			amt:      sdkmath.NewInt(-18),
			expected: coin(-18, "apple"),
		},
		{
			name:     "ask order, big amt",
			receiver: OrderFulfillment{Order: askOrder(19, coin(20, "apple"), coin(21, "plum"))},
			amt:      newInt(t, "123,000,000,000,000,000,000,000,000,000,000,321"),
			expected: bigCoin("123,000,000,000,000,000,000,000,000,000,000,321", "apple"),
		},
		{
			name:     "bid order",
			receiver: OrderFulfillment{Order: bidOrder(3, coin(4, "apple"), coin(5, "plum"))},
			amt:      sdkmath.NewInt(6),
			expected: coin(6, "apple"),
		},
		{
			name:     "bid order with negative assets",
			receiver: OrderFulfillment{Order: bidOrder(7, coin(-8, "apple"), coin(9, "plum"))},
			amt:      sdkmath.NewInt(10),
			expected: coin(10, "apple"),
		},
		{
			name:     "bid order, negative amt",
			receiver: OrderFulfillment{Order: bidOrder(11, coin(12, "apple"), coin(13, "plum"))},
			amt:      sdkmath.NewInt(-14),
			expected: coin(-14, "apple"),
		},
		{
			name:     "bid order with negative assets, negative amt",
			receiver: OrderFulfillment{Order: bidOrder(15, coin(-16, "apple"), coin(17, "plum"))},
			amt:      sdkmath.NewInt(-18),
			expected: coin(-18, "apple"),
		},
		{
			name:     "bid order, big amt",
			receiver: OrderFulfillment{Order: bidOrder(19, coin(20, "apple"), coin(21, "plum"))},
			amt:      newInt(t, "123,000,000,000,000,000,000,000,000,000,000,321"),
			expected: bigCoin("123,000,000,000,000,000,000,000,000,000,000,321", "apple"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coin
			testFunc := func() {
				actual = tc.receiver.assetCoin(tc.amt)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "assetCoin(%s)", tc.amt)
			assert.Equal(t, tc.expected.String(), actual.String(), "assetCoin(%s) result", tc.amt)
		})
	}
}

func TestOrderFulfillment_PriceCoin(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	bigCoin := func(amount, denom string) sdk.Coin {
		amt := newInt(t, amount)
		return sdk.Coin{Denom: denom, Amount: amt}
	}
	askOrder := func(orderID uint64, assets, price sdk.Coin) *Order {
		return NewOrder(orderID).WithAsk(&AskOrder{
			Assets: assets,
			Price:  price,
		})
	}
	bidOrder := func(orderID uint64, assets, price sdk.Coin) *Order {
		return NewOrder(orderID).WithBid(&BidOrder{
			Assets: assets,
			Price:  price,
		})
	}

	tests := []struct {
		name     string
		receiver OrderFulfillment
		amt      sdkmath.Int
		expected sdk.Coin
		expPanic string
	}{
		{
			name:     "nil order",
			receiver: OrderFulfillment{Order: nil},
			amt:      sdkmath.NewInt(0),
			expPanic: "runtime error: invalid memory address or nil pointer dereference",
		},
		{
			name:     "nil inside order",
			receiver: OrderFulfillment{Order: NewOrder(1)},
			amt:      sdkmath.NewInt(0),
			expPanic: nilSubTypeErr(1),
		},
		{
			name:     "unknown inside order",
			receiver: OrderFulfillment{Order: newUnknownOrder(2)},
			amt:      sdkmath.NewInt(0),
			expPanic: unknownSubTypeErr(2),
		},
		{
			name:     "ask order",
			receiver: OrderFulfillment{Order: askOrder(3, coin(4, "apple"), coin(5, "plum"))},
			amt:      sdkmath.NewInt(6),
			expected: coin(6, "plum"),
		},
		{
			name:     "ask order with negative assets",
			receiver: OrderFulfillment{Order: askOrder(7, coin(-8, "apple"), coin(9, "plum"))},
			amt:      sdkmath.NewInt(10),
			expected: coin(10, "plum"),
		},
		{
			name:     "ask order, negative amt",
			receiver: OrderFulfillment{Order: askOrder(11, coin(12, "apple"), coin(13, "plum"))},
			amt:      sdkmath.NewInt(-14),
			expected: coin(-14, "plum"),
		},
		{
			name:     "ask order with negative assets, negative amt",
			receiver: OrderFulfillment{Order: askOrder(15, coin(-16, "apple"), coin(17, "plum"))},
			amt:      sdkmath.NewInt(-18),
			expected: coin(-18, "plum"),
		},
		{
			name:     "ask order, big amt",
			receiver: OrderFulfillment{Order: askOrder(19, coin(20, "apple"), coin(21, "plum"))},
			amt:      newInt(t, "123,000,000,000,000,000,000,000,000,000,000,321"),
			expected: bigCoin("123,000,000,000,000,000,000,000,000,000,000,321", "plum"),
		},
		{
			name:     "bid order",
			receiver: OrderFulfillment{Order: bidOrder(3, coin(4, "apple"), coin(5, "plum"))},
			amt:      sdkmath.NewInt(6),
			expected: coin(6, "plum"),
		},
		{
			name:     "bid order with negative assets",
			receiver: OrderFulfillment{Order: bidOrder(7, coin(-8, "apple"), coin(9, "plum"))},
			amt:      sdkmath.NewInt(10),
			expected: coin(10, "plum"),
		},
		{
			name:     "bid order, negative amt",
			receiver: OrderFulfillment{Order: bidOrder(11, coin(12, "apple"), coin(13, "plum"))},
			amt:      sdkmath.NewInt(-14),
			expected: coin(-14, "plum"),
		},
		{
			name:     "bid order with negative assets, negative amt",
			receiver: OrderFulfillment{Order: bidOrder(15, coin(-16, "apple"), coin(17, "plum"))},
			amt:      sdkmath.NewInt(-18),
			expected: coin(-18, "plum"),
		},
		{
			name:     "bid order, big amt",
			receiver: OrderFulfillment{Order: bidOrder(19, coin(20, "apple"), coin(21, "plum"))},
			amt:      newInt(t, "123,000,000,000,000,000,000,000,000,000,000,321"),
			expected: bigCoin("123,000,000,000,000,000,000,000,000,000,000,321", "plum"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coin
			testFunc := func() {
				actual = tc.receiver.priceCoin(tc.amt)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "priceCoin(%s)", tc.amt)
			assert.Equal(t, tc.expected.String(), actual.String(), "priceCoin(%s) result", tc.amt)
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

func TestOrderFulfillment_DistributeAssets(t *testing.T) {
	newOF := func(order *Order, assetsUnfilled, assetsFilled int64, dists ...*Distribution) *OrderFulfillment {
		rv := &OrderFulfillment{
			Order:             order,
			AssetsUnfilledAmt: sdkmath.NewInt(assetsUnfilled),
			AssetsFilledAmt:   sdkmath.NewInt(assetsFilled),
		}
		if assetsUnfilled == 0 {
			rv.AssetsUnfilledAmt = ZeroAmtAfterSub
		}
		if len(dists) > 0 {
			rv.AssetDists = dists
		}
		return rv

	}
	askOF := func(orderID uint64, assetsUnfilled, assetsFilled int64, dists ...*Distribution) *OrderFulfillment {
		order := NewOrder(orderID).WithAsk(&AskOrder{Assets: sdk.NewInt64Coin("apple", 999)})
		return newOF(order, assetsUnfilled, assetsFilled, dists...)
	}
	bidOF := func(orderID uint64, assetsUnfilled, assetsFilled int64, dists ...*Distribution) *OrderFulfillment {
		order := NewOrder(orderID).WithBid(&BidOrder{Assets: sdk.NewInt64Coin("apple", 999)})
		return newOF(order, assetsUnfilled, assetsFilled, dists...)
	}
	dist := func(addr string, amt int64) *Distribution {
		return &Distribution{
			Address: addr,
			Amount:  sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name     string
		receiver *OrderFulfillment
		order    OrderI
		amount   sdkmath.Int
		expRes   *OrderFulfillment
		expErr   string
	}{
		{
			name:     "assets unfilled less than amount: ask, ask",
			receiver: askOF(1, 5, 0),
			order:    NewOrder(2).WithAsk(&AskOrder{Seller: "seLLer"}),
			amount:   sdkmath.NewInt(6),
			expErr:   "cannot fill ask order 1 having assets left \"5apple\" with \"6apple\" from ask order 2: overfill",
		},
		{
			name:     "assets unfilled less than amount: ask, bid",
			receiver: askOF(3, 5, 0),
			order:    NewOrder(4).WithBid(&BidOrder{Buyer: "buYer"}),
			amount:   sdkmath.NewInt(6),
			expErr:   "cannot fill ask order 3 having assets left \"5apple\" with \"6apple\" from bid order 4: overfill",
		},
		{
			name:     "assets unfilled less than amount: bid, ask",
			receiver: bidOF(5, 5, 0),
			order:    NewOrder(6).WithAsk(&AskOrder{Seller: "seLLer"}),
			amount:   sdkmath.NewInt(6),
			expErr:   "cannot fill bid order 5 having assets left \"5apple\" with \"6apple\" from ask order 6: overfill",
		},
		{
			name:     "assets unfilled less than amount: bid, bid",
			receiver: bidOF(7, 5, 0),
			order:    NewOrder(8).WithBid(&BidOrder{Buyer: "buYer"}),
			amount:   sdkmath.NewInt(6),
			expErr:   "cannot fill bid order 7 having assets left \"5apple\" with \"6apple\" from bid order 8: overfill",
		},
		{
			name:     "assets unfilled equals amount: ask, bid",
			receiver: askOF(1, 12345, 0),
			order:    NewOrder(2).WithBid(&BidOrder{Buyer: "buYer"}),
			amount:   sdkmath.NewInt(12345),
			expRes:   askOF(1, 0, 12345, dist("buYer", 12345)),
		},
		{
			name:     "assets unfilled equals amount: bid, ask",
			receiver: bidOF(1, 12345, 0),
			order:    NewOrder(2).WithAsk(&AskOrder{Seller: "seLLer"}),
			amount:   sdkmath.NewInt(12345),
			expRes:   bidOF(1, 0, 12345, dist("seLLer", 12345)),
		},
		{
			name:     "assets unfilled more than amount: ask, bid",
			receiver: askOF(1, 12345, 0),
			order:    NewOrder(2).WithBid(&BidOrder{Buyer: "buYer"}),
			amount:   sdkmath.NewInt(300),
			expRes:   askOF(1, 12045, 300, dist("buYer", 300)),
		},
		{
			name:     "assets unfilled more than amount: bid, ask",
			receiver: bidOF(1, 12345, 0),
			order:    NewOrder(2).WithAsk(&AskOrder{Seller: "seLLer"}),
			amount:   sdkmath.NewInt(300),
			expRes:   bidOF(1, 12045, 300, dist("seLLer", 300)),
		},
		{
			name:     "already has 2 dists: ask, bid",
			receiver: askOF(1, 12300, 45, dist("bbbbb", 40), dist("YYYYY", 5)),
			order:    NewOrder(2).WithBid(&BidOrder{Buyer: "buYer"}),
			amount:   sdkmath.NewInt(2000),
			expRes:   askOF(1, 10300, 2045, dist("bbbbb", 40), dist("YYYYY", 5), dist("buYer", 2000)),
		},
		{
			name:     "already has 2 dists: bid, ask",
			receiver: bidOF(1, 12300, 45, dist("sssss", 40), dist("LLLLL", 5)),
			order:    NewOrder(2).WithAsk(&AskOrder{Seller: "seLLer"}),
			amount:   sdkmath.NewInt(2000),
			expRes:   bidOF(1, 10300, 2045, dist("sssss", 40), dist("LLLLL", 5), dist("seLLer", 2000)),
		},
		{
			name:     "amt more than filled, ask, bid",
			receiver: askOF(1, 45, 12300, dist("ssss", 12300)),
			order:    NewOrder(2).WithBid(&BidOrder{Buyer: "buYer"}),
			amount:   sdkmath.NewInt(45),
			expRes:   askOF(1, 0, 12345, dist("ssss", 12300), dist("buYer", 45)),
		},
		{
			name:     "amt more than filled, bid, ask",
			receiver: bidOF(1, 45, 12300, dist("ssss", 12300)),
			order:    NewOrder(2).WithAsk(&AskOrder{Seller: "seLLer"}),
			amount:   sdkmath.NewInt(45),
			expRes:   bidOF(1, 0, 12345, dist("ssss", 12300), dist("seLLer", 45)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := copyOrderFulfillment(tc.receiver)
			if tc.expRes == nil {
				tc.expRes = copyOrderFulfillment(tc.receiver)
			}
			var err error
			testFunc := func() {
				err = tc.receiver.DistributeAssets(tc.order, tc.amount)
			}
			require.NotPanics(t, testFunc, "DistributeAssets")
			assertions.AssertErrorValue(t, err, tc.expErr, "DistributeAssets error")
			if !assertEqualOrderFulfillments(t, tc.expRes, tc.receiver, "OrderFulfillment after DistributeAssets") {
				t.Logf("Original: %s", orderFulfillmentString(orig))
				t.Logf("  Amount: %s", tc.amount)
			}
		})
	}
}

func TestDistributeAssets(t *testing.T) {
	seller, buyer := "SelleR", "BuyeR"
	newOF := func(order *Order, assetsUnfilled, assetsFilled int64, dists ...*Distribution) *OrderFulfillment {
		rv := &OrderFulfillment{
			Order:             order,
			AssetsUnfilledAmt: sdkmath.NewInt(assetsUnfilled),
			AssetsFilledAmt:   sdkmath.NewInt(assetsFilled),
		}
		if assetsUnfilled == 0 {
			rv.AssetsUnfilledAmt = ZeroAmtAfterSub
		}
		if len(dists) > 0 {
			rv.AssetDists = dists
		}
		return rv

	}
	askOF := func(orderID uint64, assetsUnfilled, assetsFilled int64, dists ...*Distribution) *OrderFulfillment {
		order := NewOrder(orderID).WithAsk(&AskOrder{
			Seller: seller,
			Assets: sdk.NewInt64Coin("apple", 999),
		})
		return newOF(order, assetsUnfilled, assetsFilled, dists...)
	}
	bidOF := func(orderID uint64, assetsUnfilled, assetsFilled int64, dists ...*Distribution) *OrderFulfillment {
		order := NewOrder(orderID).WithBid(&BidOrder{
			Buyer:  buyer,
			Assets: sdk.NewInt64Coin("apple", 999),
		})
		return newOF(order, assetsUnfilled, assetsFilled, dists...)
	}
	dist := func(addr string, amt int64) *Distribution {
		return &Distribution{
			Address: addr,
			Amount:  sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name   string
		of1    *OrderFulfillment
		of2    *OrderFulfillment
		amount sdkmath.Int
		expOF1 *OrderFulfillment
		expOF2 *OrderFulfillment
		expErr string
	}{
		{
			name:   "amount more than of1 unfilled: ask bid",
			of1:    askOF(1, 5, 0),
			of2:    bidOF(2, 6, 0),
			amount: sdkmath.NewInt(6),
			expOF2: bidOF(2, 0, 6, dist(seller, 6)),
			expErr: "cannot fill ask order 1 having assets left \"5apple\" with \"6apple\" from bid order 2: overfill",
		},
		{
			name:   "amount more than of1 unfilled: bid ask",
			of1:    bidOF(1, 5, 0),
			of2:    askOF(2, 6, 0),
			amount: sdkmath.NewInt(6),
			expOF2: askOF(2, 0, 6, dist(buyer, 6)),
			expErr: "cannot fill bid order 1 having assets left \"5apple\" with \"6apple\" from ask order 2: overfill",
		},
		{
			name:   "amount more than of2 unfilled: ask, bid",
			of1:    askOF(1, 6, 0),
			of2:    bidOF(2, 5, 0),
			amount: sdkmath.NewInt(6),
			expOF1: askOF(1, 0, 6, dist(buyer, 6)),
			expErr: "cannot fill bid order 2 having assets left \"5apple\" with \"6apple\" from ask order 1: overfill",
		},
		{
			name:   "amount more than of2 unfilled: bid, ask",
			of1:    bidOF(1, 6, 0),
			of2:    askOF(2, 5, 0),
			amount: sdkmath.NewInt(6),
			expOF1: bidOF(1, 0, 6, dist(seller, 6)),
			expErr: "cannot fill ask order 2 having assets left \"5apple\" with \"6apple\" from bid order 1: overfill",
		},
		{
			name:   "amount more than both unfilled: ask, bid",
			of1:    askOF(1, 5, 0),
			of2:    bidOF(2, 4, 0),
			amount: sdkmath.NewInt(6),
			expErr: "cannot fill ask order 1 having assets left \"5apple\" with \"6apple\" from bid order 2: overfill" + "\n" +
				"cannot fill bid order 2 having assets left \"4apple\" with \"6apple\" from ask order 1: overfill",
		},
		{
			name:   "amount more than both unfilled: ask, bid",
			of1:    bidOF(1, 5, 0),
			of2:    askOF(2, 4, 0),
			amount: sdkmath.NewInt(6),
			expErr: "cannot fill bid order 1 having assets left \"5apple\" with \"6apple\" from ask order 2: overfill" + "\n" +
				"cannot fill ask order 2 having assets left \"4apple\" with \"6apple\" from bid order 1: overfill",
		},
		{
			name:   "ask bid",
			of1:    askOF(1, 10, 55, dist("bbb", 55)),
			of2:    bidOF(2, 10, 0),
			amount: sdkmath.NewInt(9),
			expOF1: askOF(1, 1, 64, dist("bbb", 55), dist(buyer, 9)),
			expOF2: bidOF(2, 1, 9, dist(seller, 9)),
		},
		{
			name:   "bid ask",
			of1:    bidOF(1, 10, 55, dist("sss", 55)),
			of2:    askOF(2, 10, 3, dist("bbb", 3)),
			amount: sdkmath.NewInt(10),
			expOF1: bidOF(1, 0, 65, dist("sss", 55), dist(seller, 10)),
			expOF2: askOF(2, 0, 13, dist("bbb", 3), dist(buyer, 10)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			origOF1 := copyOrderFulfillment(tc.of1)
			origOF2 := copyOrderFulfillment(tc.of2)
			if tc.expOF1 == nil {
				tc.expOF1 = copyOrderFulfillment(tc.of1)
			}
			if tc.expOF2 == nil {
				tc.expOF2 = copyOrderFulfillment(tc.of2)
			}

			var err error
			testFunc := func() {
				err = DistributeAssets(tc.of1, tc.of2, tc.amount)
			}
			require.NotPanics(t, testFunc, "DistributeAssets")
			assertions.AssertErrorValue(t, err, tc.expErr, "DistributeAssets error")
			if !assertEqualOrderFulfillments(t, tc.expOF1, tc.of1, "of1 after DistributeAssets") {
				t.Logf("Original: %s", orderFulfillmentString(origOF1))
				t.Logf("  Amount: %s", tc.amount)
			}
			if !assertEqualOrderFulfillments(t, tc.expOF2, tc.of2, "of2 after DistributeAssets") {
				t.Logf("Original: %s", orderFulfillmentString(origOF2))
				t.Logf("  Amount: %s", tc.amount)
			}
		})
	}
}

func TestOrderFulfillment_DistributePrice(t *testing.T) {
	newOF := func(order *Order, priceLeft, priceApplied int64, dists ...*Distribution) *OrderFulfillment {
		rv := &OrderFulfillment{
			Order:           order,
			PriceLeftAmt:    sdkmath.NewInt(priceLeft),
			PriceAppliedAmt: sdkmath.NewInt(priceApplied),
		}
		if priceLeft == 0 {
			rv.PriceLeftAmt = ZeroAmtAfterSub
		}
		if len(dists) > 0 {
			rv.PriceDists = dists
		}
		return rv

	}
	askOF := func(orderID uint64, priceLeft, priceApplied int64, dists ...*Distribution) *OrderFulfillment {
		order := NewOrder(orderID).WithAsk(&AskOrder{Price: sdk.NewInt64Coin("peach", 999)})
		return newOF(order, priceLeft, priceApplied, dists...)
	}
	bidOF := func(orderID uint64, priceLeft, priceApplied int64, dists ...*Distribution) *OrderFulfillment {
		order := NewOrder(orderID).WithBid(&BidOrder{Price: sdk.NewInt64Coin("peach", 999)})
		return newOF(order, priceLeft, priceApplied, dists...)
	}
	dist := func(addr string, amt int64) *Distribution {
		return &Distribution{
			Address: addr,
			Amount:  sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name     string
		receiver *OrderFulfillment
		order    OrderI
		amount   sdkmath.Int
		expRes   *OrderFulfillment
		expErr   string
	}{
		{
			name:     "assets unfilled less than amount: ask, ask",
			receiver: askOF(1, 5, 0),
			order:    NewOrder(2).WithAsk(&AskOrder{Seller: "seLLer"}),
			amount:   sdkmath.NewInt(6),
			expRes:   askOF(1, -1, 6, dist("seLLer", 6)),
		},
		{
			name:     "assets unfilled less than amount: ask, bid",
			receiver: askOF(3, 5, 0),
			order:    NewOrder(4).WithBid(&BidOrder{Buyer: "buYer"}),
			amount:   sdkmath.NewInt(6),
			expRes:   askOF(3, -1, 6, dist("buYer", 6)),
		},
		{
			name:     "assets unfilled less than amount: bid, ask",
			receiver: bidOF(5, 5, 0),
			order:    NewOrder(6).WithAsk(&AskOrder{Seller: "seLLer"}),
			amount:   sdkmath.NewInt(6),
			expErr:   "cannot fill bid order 5 having price left \"5peach\" to ask order 6 at a price of \"6peach\": overfill",
		},
		{
			name:     "assets unfilled less than amount: bid, bid",
			receiver: bidOF(7, 5, 0),
			order:    NewOrder(8).WithBid(&BidOrder{Buyer: "buYer"}),
			amount:   sdkmath.NewInt(6),
			expErr:   "cannot fill bid order 7 having price left \"5peach\" to bid order 8 at a price of \"6peach\": overfill",
		},
		{
			name:     "assets unfilled equals amount: ask, bid",
			receiver: askOF(1, 12345, 0),
			order:    NewOrder(2).WithBid(&BidOrder{Buyer: "buYer"}),
			amount:   sdkmath.NewInt(12345),
			expRes:   askOF(1, 0, 12345, dist("buYer", 12345)),
		},
		{
			name:     "assets unfilled equals amount: bid, ask",
			receiver: bidOF(1, 12345, 0),
			order:    NewOrder(2).WithAsk(&AskOrder{Seller: "seLLer"}),
			amount:   sdkmath.NewInt(12345),
			expRes:   bidOF(1, 0, 12345, dist("seLLer", 12345)),
		},
		{
			name:     "assets unfilled more than amount: ask, bid",
			receiver: askOF(1, 12345, 0),
			order:    NewOrder(2).WithBid(&BidOrder{Buyer: "buYer"}),
			amount:   sdkmath.NewInt(300),
			expRes:   askOF(1, 12045, 300, dist("buYer", 300)),
		},
		{
			name:     "assets unfilled more than amount: bid, ask",
			receiver: bidOF(1, 12345, 0),
			order:    NewOrder(2).WithAsk(&AskOrder{Seller: "seLLer"}),
			amount:   sdkmath.NewInt(300),
			expRes:   bidOF(1, 12045, 300, dist("seLLer", 300)),
		},
		{
			name:     "already has 2 dists: ask, bid",
			receiver: askOF(1, 12300, 45, dist("bbbbb", 40), dist("YYYYY", 5)),
			order:    NewOrder(2).WithBid(&BidOrder{Buyer: "buYer"}),
			amount:   sdkmath.NewInt(2000),
			expRes:   askOF(1, 10300, 2045, dist("bbbbb", 40), dist("YYYYY", 5), dist("buYer", 2000)),
		},
		{
			name:     "already has 2 dists: bid, ask",
			receiver: bidOF(1, 12300, 45, dist("sssss", 40), dist("LLLLL", 5)),
			order:    NewOrder(2).WithAsk(&AskOrder{Seller: "seLLer"}),
			amount:   sdkmath.NewInt(2000),
			expRes:   bidOF(1, 10300, 2045, dist("sssss", 40), dist("LLLLL", 5), dist("seLLer", 2000)),
		},
		{
			name:     "amt more than filled, ask, bid",
			receiver: askOF(1, 45, 12300, dist("ssss", 12300)),
			order:    NewOrder(2).WithBid(&BidOrder{Buyer: "buYer"}),
			amount:   sdkmath.NewInt(45),
			expRes:   askOF(1, 0, 12345, dist("ssss", 12300), dist("buYer", 45)),
		},
		{
			name:     "amt more than filled, bid, ask",
			receiver: bidOF(1, 45, 12300, dist("ssss", 12300)),
			order:    NewOrder(2).WithAsk(&AskOrder{Seller: "seLLer"}),
			amount:   sdkmath.NewInt(45),
			expRes:   bidOF(1, 0, 12345, dist("ssss", 12300), dist("seLLer", 45)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := copyOrderFulfillment(tc.receiver)
			if tc.expRes == nil {
				tc.expRes = copyOrderFulfillment(tc.receiver)
			}
			var err error
			testFunc := func() {
				err = tc.receiver.DistributePrice(tc.order, tc.amount)
			}
			require.NotPanics(t, testFunc, "DistributePrice")
			assertions.AssertErrorValue(t, err, tc.expErr, "DistributePrice error")
			if !assertEqualOrderFulfillments(t, tc.expRes, tc.receiver, "OrderFulfillment after DistributePrice") {
				t.Logf("Original: %s", orderFulfillmentString(orig))
				t.Logf("  Amount: %s", tc.amount)
			}
		})
	}
}

func TestDistributePrice(t *testing.T) {
	seller, buyer := "SelleR", "BuyeR"
	newOF := func(order *Order, priceLeft, priceApplied int64, dists ...*Distribution) *OrderFulfillment {
		rv := &OrderFulfillment{
			Order:           order,
			PriceLeftAmt:    sdkmath.NewInt(priceLeft),
			PriceAppliedAmt: sdkmath.NewInt(priceApplied),
		}
		if priceLeft == 0 {
			rv.PriceLeftAmt = ZeroAmtAfterSub
		}
		if len(dists) > 0 {
			rv.PriceDists = dists
		}
		return rv

	}
	askOF := func(orderID uint64, priceLeft, priceApplied int64, dists ...*Distribution) *OrderFulfillment {
		order := NewOrder(orderID).WithAsk(&AskOrder{
			Seller: seller,
			Price:  sdk.NewInt64Coin("peach", 999),
		})
		return newOF(order, priceLeft, priceApplied, dists...)
	}
	bidOF := func(orderID uint64, priceLeft, priceApplied int64, dists ...*Distribution) *OrderFulfillment {
		order := NewOrder(orderID).WithBid(&BidOrder{
			Buyer: buyer,
			Price: sdk.NewInt64Coin("peach", 999),
		})
		return newOF(order, priceLeft, priceApplied, dists...)
	}
	dist := func(addr string, amt int64) *Distribution {
		return &Distribution{
			Address: addr,
			Amount:  sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name   string
		of1    *OrderFulfillment
		of2    *OrderFulfillment
		amount sdkmath.Int
		expOF1 *OrderFulfillment
		expOF2 *OrderFulfillment
		expErr string
	}{
		{
			name:   "amount more than of1 unfilled: ask bid",
			of1:    askOF(1, 5, 0),
			of2:    bidOF(2, 6, 0),
			amount: sdkmath.NewInt(6),
			expOF1: askOF(1, -1, 6, dist(buyer, 6)),
			expOF2: bidOF(2, 0, 6, dist(seller, 6)),
		},
		{
			name:   "amount more than of1 unfilled: bid ask",
			of1:    bidOF(1, 5, 0),
			of2:    askOF(2, 6, 0),
			amount: sdkmath.NewInt(6),
			expOF2: askOF(2, 0, 6, dist(buyer, 6)),
			expErr: "cannot fill bid order 1 having price left \"5peach\" to ask order 2 at a price of \"6peach\": overfill",
		},
		{
			name:   "amount more than of2 unfilled: ask, bid",
			of1:    askOF(1, 6, 0),
			of2:    bidOF(2, 5, 0),
			amount: sdkmath.NewInt(6),
			expOF1: askOF(1, 0, 6, dist(buyer, 6)),
			expErr: "cannot fill bid order 2 having price left \"5peach\" to ask order 1 at a price of \"6peach\": overfill",
		},
		{
			name:   "amount more than of2 unfilled: bid, ask",
			of1:    bidOF(1, 6, 0),
			of2:    askOF(2, 5, 0),
			amount: sdkmath.NewInt(6),
			expOF1: bidOF(1, 0, 6, dist(seller, 6)),
			expOF2: askOF(2, -1, 6, dist(buyer, 6)),
		},
		{
			name:   "amount more than both unfilled: ask, bid",
			of1:    askOF(1, 5, 0),
			of2:    bidOF(2, 4, 0),
			amount: sdkmath.NewInt(6),
			expOF1: askOF(1, -1, 6, dist(buyer, 6)),
			expErr: "cannot fill bid order 2 having price left \"4peach\" to ask order 1 at a price of \"6peach\": overfill",
		},
		{
			name:   "amount more than both unfilled: ask, bid",
			of1:    bidOF(1, 5, 0),
			of2:    askOF(2, 4, 0),
			amount: sdkmath.NewInt(6),
			expOF2: askOF(2, -2, 6, dist(buyer, 6)),
			expErr: "cannot fill bid order 1 having price left \"5peach\" to ask order 2 at a price of \"6peach\": overfill",
		},
		{
			name:   "ask bid",
			of1:    askOF(1, 10, 55, dist("bbb", 55)),
			of2:    bidOF(2, 10, 0),
			amount: sdkmath.NewInt(9),
			expOF1: askOF(1, 1, 64, dist("bbb", 55), dist(buyer, 9)),
			expOF2: bidOF(2, 1, 9, dist(seller, 9)),
		},
		{
			name:   "bid ask",
			of1:    bidOF(1, 10, 55, dist("sss", 55)),
			of2:    askOF(2, 10, 3, dist("bbb", 3)),
			amount: sdkmath.NewInt(10),
			expOF1: bidOF(1, 0, 65, dist("sss", 55), dist(seller, 10)),
			expOF2: askOF(2, 0, 13, dist("bbb", 3), dist(buyer, 10)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			origOF1 := copyOrderFulfillment(tc.of1)
			origOF2 := copyOrderFulfillment(tc.of2)
			if tc.expOF1 == nil {
				tc.expOF1 = copyOrderFulfillment(tc.of1)
			}
			if tc.expOF2 == nil {
				tc.expOF2 = copyOrderFulfillment(tc.of2)
			}

			var err error
			testFunc := func() {
				err = DistributePrice(tc.of1, tc.of2, tc.amount)
			}
			require.NotPanics(t, testFunc, "DistributePrice")
			assertions.AssertErrorValue(t, err, tc.expErr, "DistributePrice error")
			if !assertEqualOrderFulfillments(t, tc.expOF1, tc.of1, "of1 after DistributePrice") {
				t.Logf("Original: %s", orderFulfillmentString(origOF1))
				t.Logf("  Amount: %s", tc.amount)
			}
			if !assertEqualOrderFulfillments(t, tc.expOF2, tc.of2, "of2 after DistributePrice") {
				t.Logf("Original: %s", orderFulfillmentString(origOF2))
				t.Logf("  Amount: %s", tc.amount)
			}
		})
	}
}

func TestOrderFulfillment_SplitOrder(t *testing.T) {
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
		name        string
		receiver    *OrderFulfillment
		expUnfilled *Order
		expReceiver *OrderFulfillment
		expErr      string
	}{
		{
			name: "order split error: ask",
			receiver: &OrderFulfillment{
				Order:           askOrder(8, 10, 100),
				AssetsFilledAmt: sdkmath.NewInt(-1),
			},
			expErr: "cannot split ask order 8 having asset \"10apple\" at \"-1apple\": amount filled not positive",
		},
		{
			name: "order split error: bid",
			receiver: &OrderFulfillment{
				Order:           bidOrder(9, 10, 100),
				AssetsFilledAmt: sdkmath.NewInt(-1),
			},
			expErr: "cannot split bid order 9 having asset \"10apple\" at \"-1apple\": amount filled not positive",
		},
		{
			name: "okay: ask",
			receiver: &OrderFulfillment{
				Order:             askOrder(17, 10, 100, coin(20, "fig")),
				AssetsFilledAmt:   sdkmath.NewInt(9),
				AssetsUnfilledAmt: sdkmath.NewInt(1),
				PriceAppliedAmt:   sdkmath.NewInt(300),
				PriceLeftAmt:      sdkmath.NewInt(-200),
			},
			expUnfilled: askOrder(17, 1, 10, coin(2, "fig")),
			expReceiver: &OrderFulfillment{
				Order:             askOrder(17, 9, 90, coin(18, "fig")),
				AssetsFilledAmt:   sdkmath.NewInt(9),
				AssetsUnfilledAmt: sdkmath.NewInt(0),
				PriceAppliedAmt:   sdkmath.NewInt(300),
				PriceLeftAmt:      sdkmath.NewInt(-210),
			},
		},
		{
			name: "okay: bid",
			receiver: &OrderFulfillment{
				Order:             bidOrder(19, 10, 100, coin(20, "fig")),
				AssetsFilledAmt:   sdkmath.NewInt(9),
				AssetsUnfilledAmt: sdkmath.NewInt(1),
				PriceAppliedAmt:   sdkmath.NewInt(300),
				PriceLeftAmt:      sdkmath.NewInt(-200),
			},
			expUnfilled: bidOrder(19, 1, 10, coin(2, "fig")),
			expReceiver: &OrderFulfillment{
				Order:             bidOrder(19, 9, 90, coin(18, "fig")),
				AssetsFilledAmt:   sdkmath.NewInt(9),
				AssetsUnfilledAmt: sdkmath.NewInt(0),
				PriceAppliedAmt:   sdkmath.NewInt(300),
				PriceLeftAmt:      sdkmath.NewInt(-210),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := copyOrderFulfillment(tc.receiver)
			if tc.expReceiver == nil {
				tc.expReceiver = copyOrderFulfillment(tc.receiver)
			}

			var unfilled *Order
			var err error
			testFunc := func() {
				unfilled, err = tc.receiver.SplitOrder()
			}
			require.NotPanics(t, testFunc, "SplitOrder")
			assertions.AssertErrorValue(t, err, tc.expErr, "SplitOrder error")
			assert.Equalf(t, tc.expUnfilled, unfilled, "SplitOrder unfilled order")
			if !assertEqualOrderFulfillments(t, tc.expReceiver, tc.receiver, "OrderFulfillment after SplitOrder") {
				t.Logf("Original: %s", orderFulfillmentString(orig))
			}
		})
	}
}

func TestOrderFulfillment_AsFilledOrder(t *testing.T) {
	askOrder := NewOrder(53).WithAsk(&AskOrder{
		MarketId:                765,
		Seller:                  "mefirst",
		Assets:                  sdk.NewInt64Coin("apple", 15),
		Price:                   sdk.NewInt64Coin("peach", 88),
		SellerSettlementFlatFee: &sdk.Coin{Denom: "fig", Amount: sdkmath.NewInt(6)},
		AllowPartial:            true,
	})
	bidOrder := NewOrder(9556).WithBid(&BidOrder{
		MarketId:            145,
		Buyer:               "gimmiegimmie",
		Assets:              sdk.NewInt64Coin("acorn", 1171),
		Price:               sdk.NewInt64Coin("prune", 5100),
		BuyerSettlementFees: sdk.NewCoins(sdk.Coin{Denom: "fig", Amount: sdkmath.NewInt(14)}),
		AllowPartial:        false,
	})

	tests := []struct {
		name     string
		receiver OrderFulfillment
		expected *FilledOrder
	}{
		{
			name: "ask order",
			receiver: OrderFulfillment{
				Order:           askOrder,
				PriceAppliedAmt: sdkmath.NewInt(132),
				FeesToPay:       sdk.NewCoins(sdk.Coin{Denom: "grape", Amount: sdkmath.NewInt(7)}),
			},
			expected: &FilledOrder{
				order:       askOrder,
				actualPrice: sdk.NewInt64Coin("peach", 132),
				actualFees:  sdk.NewCoins(sdk.Coin{Denom: "grape", Amount: sdkmath.NewInt(7)}),
			},
		},
		{
			name: "bid order",
			receiver: OrderFulfillment{
				Order:           bidOrder,
				PriceAppliedAmt: sdkmath.NewInt(5123),
				FeesToPay:       sdk.NewCoins(sdk.Coin{Denom: "grape", Amount: sdkmath.NewInt(23)}),
			},
			expected: &FilledOrder{
				order:       bidOrder,
				actualPrice: sdk.NewInt64Coin("prune", 5123),
				actualFees:  sdk.NewCoins(sdk.Coin{Denom: "grape", Amount: sdkmath.NewInt(23)}),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual *FilledOrder
			testFunc := func() {
				actual = tc.receiver.AsFilledOrder()
			}
			require.NotPanics(t, testFunc, "AsFilledOrder()")
			assert.Equal(t, tc.expected, actual, "AsFilledOrder() result")
		})
	}
}

func TestSumAssetsAndPrice(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	askOrder := func(orderID uint64, assets, price sdk.Coin) *Order {
		return NewOrder(orderID).WithAsk(&AskOrder{
			Assets: assets,
			Price:  price,
		})
	}
	bidOrder := func(orderID uint64, assets, price sdk.Coin) *Order {
		return NewOrder(orderID).WithBid(&BidOrder{
			Assets: assets,
			Price:  price,
		})
	}

	tests := []struct {
		name      string
		orders    []*Order
		expAssets sdk.Coins
		expPrice  sdk.Coins
		expPanic  string
	}{
		{
			name:      "nil orders",
			orders:    nil,
			expAssets: nil,
			expPrice:  nil,
		},
		{
			name:      "empty orders",
			orders:    []*Order{},
			expAssets: nil,
			expPrice:  nil,
		},
		{
			name: "nil inside order",
			orders: []*Order{
				askOrder(1, coin(2, "apple"), coin(3, "plum")),
				NewOrder(4),
				askOrder(5, coin(6, "apple"), coin(7, "plum")),
			},
			expPanic: nilSubTypeErr(4),
		},
		{
			name: "unknown inside order",
			orders: []*Order{
				askOrder(1, coin(2, "apple"), coin(3, "plum")),
				newUnknownOrder(4),
				askOrder(5, coin(6, "apple"), coin(7, "plum")),
			},
			expPanic: unknownSubTypeErr(4),
		},
		{
			name:      "one order, ask",
			orders:    []*Order{askOrder(1, coin(2, "apple"), coin(3, "plum"))},
			expAssets: sdk.NewCoins(coin(2, "apple")),
			expPrice:  sdk.NewCoins(coin(3, "plum")),
		},
		{
			name:      "one order, bid",
			orders:    []*Order{bidOrder(1, coin(2, "apple"), coin(3, "plum"))},
			expAssets: sdk.NewCoins(coin(2, "apple")),
			expPrice:  sdk.NewCoins(coin(3, "plum")),
		},
		{
			name: "2 orders, same denoms",
			orders: []*Order{
				askOrder(1, coin(2, "apple"), coin(3, "plum")),
				bidOrder(4, coin(5, "apple"), coin(6, "plum")),
			},
			expAssets: sdk.NewCoins(coin(7, "apple")),
			expPrice:  sdk.NewCoins(coin(9, "plum")),
		},
		{
			name: "2 orders, diff denoms",
			orders: []*Order{
				bidOrder(1, coin(2, "avocado"), coin(3, "peach")),
				askOrder(4, coin(5, "apple"), coin(6, "plum")),
			},
			expAssets: sdk.NewCoins(coin(2, "avocado"), coin(5, "apple")),
			expPrice:  sdk.NewCoins(coin(3, "peach"), coin(6, "plum")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var assets, price sdk.Coins
			testFunc := func() {
				assets, price = sumAssetsAndPrice(tc.orders)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "sumAssetsAndPrice")
			assert.Equal(t, tc.expAssets.String(), assets.String(), "sumAssetsAndPrice")
			assert.Equal(t, tc.expPrice.String(), price.String(), "sumAssetsAndPrice")
		})
	}
}

func TestSumPriceLeft(t *testing.T) {
	tests := []struct {
		name         string
		fulfillments []*OrderFulfillment
		expected     sdkmath.Int
	}{
		{
			name:         "nil fulfillments",
			fulfillments: nil,
			expected:     sdkmath.NewInt(0),
		},
		{
			name:         "empty fulfillments",
			fulfillments: []*OrderFulfillment{},
			expected:     sdkmath.NewInt(0),
		},
		{
			name:         "one fulfillment, positive",
			fulfillments: []*OrderFulfillment{{PriceLeftAmt: sdkmath.NewInt(8)}},
			expected:     sdkmath.NewInt(8),
		},
		{
			name:         "one fulfillment, zero",
			fulfillments: []*OrderFulfillment{{PriceLeftAmt: sdkmath.NewInt(0)}},
			expected:     sdkmath.NewInt(0),
		},
		{
			name:         "one fulfillment, negative",
			fulfillments: []*OrderFulfillment{{PriceLeftAmt: sdkmath.NewInt(-3)}},
			expected:     sdkmath.NewInt(-3),
		},
		{
			name: "three fulfillments",
			fulfillments: []*OrderFulfillment{
				{PriceLeftAmt: sdkmath.NewInt(10)},
				{PriceLeftAmt: sdkmath.NewInt(200)},
				{PriceLeftAmt: sdkmath.NewInt(3000)},
			},
			expected: sdkmath.NewInt(3210),
		},
		{
			name: "three fulfillments, one negative",
			fulfillments: []*OrderFulfillment{
				{PriceLeftAmt: sdkmath.NewInt(10)},
				{PriceLeftAmt: sdkmath.NewInt(-200)},
				{PriceLeftAmt: sdkmath.NewInt(3000)},
			},
			expected: sdkmath.NewInt(2810),
		},
		{
			name: "three fulfillments, all negative",
			fulfillments: []*OrderFulfillment{
				{PriceLeftAmt: sdkmath.NewInt(-10)},
				{PriceLeftAmt: sdkmath.NewInt(-200)},
				{PriceLeftAmt: sdkmath.NewInt(-3000)},
			},
			expected: sdkmath.NewInt(-3210),
		},
		{
			name: "three fulfillments, all large",
			fulfillments: []*OrderFulfillment{
				{PriceLeftAmt: newInt(t, "3,000,000,000,000,000,000,000,000,000,000,300")},
				{PriceLeftAmt: newInt(t, "40,000,000,000,000,000,000,000,000,000,000,040")},
				{PriceLeftAmt: newInt(t, "500,000,000,000,000,000,000,000,000,000,000,005")},
			},
			expected: newInt(t, "543,000,000,000,000,000,000,000,000,000,000,345"),
		},
		{
			name: "four fullfillments, small negative zero large",
			fulfillments: []*OrderFulfillment{
				{PriceLeftAmt: sdkmath.NewInt(654_789)},
				{PriceLeftAmt: sdkmath.NewInt(-789)},
				{PriceLeftAmt: sdkmath.NewInt(0)},
				{PriceLeftAmt: newInt(t, "543,000,000,000,000,000,000,000,000,000,000,345")},
			},
			expected: newInt(t, "543,000,000,000,000,000,000,000,000,000,654,345"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdkmath.Int
			testFunc := func() {
				actual = sumPriceLeft(tc.fulfillments)
			}
			require.NotPanics(t, testFunc, "sumPriceLeft")
			assert.Equal(t, tc.expected, actual, "sumPriceLeft")
		})
	}
}

func TestBuildSettlement(t *testing.T) {
	// TODO[1658]: func TestBuildSettlement(t *testing.T)
	t.Skipf("not written")
}

func TestValidateCanSettle(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	askOrder := func(orderID uint64, assets, price sdk.Coin) *Order {
		return NewOrder(orderID).WithAsk(&AskOrder{
			Assets: assets,
			Price:  price,
		})
	}
	bidOrder := func(orderID uint64, assets, price sdk.Coin) *Order {
		return NewOrder(orderID).WithBid(&BidOrder{
			Assets: assets,
			Price:  price,
		})
	}

	tests := []struct {
		name      string
		askOrders []*Order
		bidOrders []*Order
		expErr    string
	}{
		{
			name:      "nil ask orders",
			askOrders: nil,
			bidOrders: []*Order{bidOrder(8, coin(10, "apple"), coin(11, "peach"))},
			expErr:    "no ask orders provided",
		},
		{
			name:      "no bid orders",
			askOrders: []*Order{askOrder(7, coin(10, "apple"), coin(11, "peach"))},
			bidOrders: nil,
			expErr:    "no bid orders provided",
		},
		{
			name:      "no orders",
			askOrders: nil,
			bidOrders: nil,
			expErr:    joinErrs("no ask orders provided", "no bid orders provided"),
		},
		{
			name: "bid order in asks",
			askOrders: []*Order{
				askOrder(7, coin(10, "apple"), coin(11, "peach")),
				bidOrder(8, coin(10, "apple"), coin(11, "peach")),
				askOrder(9, coin(10, "apple"), coin(11, "peach")),
			},
			bidOrders: []*Order{bidOrder(22, coin(10, "apple"), coin(11, "peach"))},
			expErr:    "bid order 8 is not an ask order but is in the askOrders list at index 1",
		},
		{
			name: "nil inside order in asks",
			askOrders: []*Order{
				askOrder(7, coin(10, "apple"), coin(11, "peach")),
				NewOrder(8),
				askOrder(9, coin(10, "apple"), coin(11, "peach")),
			},
			bidOrders: []*Order{bidOrder(22, coin(10, "apple"), coin(11, "peach"))},
			expErr:    "<nil> order 8 is not an ask order but is in the askOrders list at index 1",
		},
		{
			name: "unknown inside order in asks",
			askOrders: []*Order{
				askOrder(7, coin(10, "apple"), coin(11, "peach")),
				newUnknownOrder(8),
				askOrder(9, coin(10, "apple"), coin(11, "peach")),
			},
			bidOrders: []*Order{bidOrder(22, coin(10, "apple"), coin(11, "peach"))},
			expErr:    "*exchange.unknownOrderType order 8 is not an ask order but is in the askOrders list at index 1",
		},
		{
			name:      "ask order in bids",
			askOrders: []*Order{askOrder(7, coin(10, "apple"), coin(11, "peach"))},
			bidOrders: []*Order{
				bidOrder(21, coin(10, "apple"), coin(11, "peach")),
				askOrder(22, coin(10, "apple"), coin(11, "peach")),
				bidOrder(23, coin(10, "apple"), coin(11, "peach")),
			},
			expErr: "ask order 22 is not a bid order but is in the bidOrders list at index 1",
		},
		{
			name:      "nil inside order in bids",
			askOrders: []*Order{askOrder(7, coin(10, "apple"), coin(11, "peach"))},
			bidOrders: []*Order{
				bidOrder(21, coin(10, "apple"), coin(11, "peach")),
				NewOrder(22),
				bidOrder(23, coin(10, "apple"), coin(11, "peach")),
			},
			expErr: "<nil> order 22 is not a bid order but is in the bidOrders list at index 1",
		},
		{
			name:      "unknown inside order in bids",
			askOrders: []*Order{askOrder(7, coin(10, "apple"), coin(11, "peach"))},
			bidOrders: []*Order{
				bidOrder(21, coin(10, "apple"), coin(11, "peach")),
				newUnknownOrder(22),
				bidOrder(23, coin(10, "apple"), coin(11, "peach")),
			},
			expErr: "*exchange.unknownOrderType order 22 is not a bid order but is in the bidOrders list at index 1",
		},
		{
			name: "orders in wrong args",
			askOrders: []*Order{
				askOrder(15, coin(10, "apple"), coin(11, "peach")),
				bidOrder(16, coin(10, "apple"), coin(11, "peach")),
				askOrder(17, coin(10, "apple"), coin(11, "peach")),
			},
			bidOrders: []*Order{
				bidOrder(91, coin(10, "apple"), coin(11, "peach")),
				askOrder(92, coin(10, "apple"), coin(11, "peach")),
				bidOrder(93, coin(10, "apple"), coin(11, "peach")),
			},
			expErr: joinErrs(
				"bid order 16 is not an ask order but is in the askOrders list at index 1",
				"ask order 92 is not a bid order but is in the bidOrders list at index 1",
			),
		},
		{
			name: "multiple ask asset denoms",
			askOrders: []*Order{
				askOrder(55, coin(10, "apple"), coin(11, "peach")),
				askOrder(56, coin(20, "avocado"), coin(22, "peach")),
			},
			bidOrders: []*Order{
				bidOrder(61, coin(10, "apple"), coin(11, "peach")),
			},
			expErr: "cannot settle with multiple ask order asset denoms \"10apple,20avocado\"",
		},
		{
			name: "multiple ask price denoms",
			askOrders: []*Order{
				askOrder(55, coin(10, "apple"), coin(11, "peach")),
				askOrder(56, coin(20, "apple"), coin(22, "plum")),
			},
			bidOrders: []*Order{
				bidOrder(61, coin(10, "apple"), coin(11, "peach")),
			},
			expErr: "cannot settle with multiple ask order price denoms \"11peach,22plum\"",
		},
		{
			name:      "multiple bid asset denoms",
			askOrders: []*Order{askOrder(88, coin(10, "apple"), coin(11, "peach"))},
			bidOrders: []*Order{
				bidOrder(12, coin(10, "apple"), coin(11, "peach")),
				bidOrder(13, coin(20, "avocado"), coin(22, "peach")),
			},
			expErr: "cannot settle with multiple bid order asset denoms \"10apple,20avocado\"",
		},
		{
			name:      "multiple bid price denoms",
			askOrders: []*Order{askOrder(88, coin(10, "apple"), coin(11, "peach"))},
			bidOrders: []*Order{
				bidOrder(12, coin(10, "apple"), coin(11, "peach")),
				bidOrder(13, coin(20, "apple"), coin(22, "plum")),
			},
			expErr: "cannot settle with multiple bid order price denoms \"11peach,22plum\"",
		},
		{
			name: "all different denoms",
			askOrders: []*Order{
				askOrder(55, coin(10, "apple"), coin(11, "peach")),
				askOrder(56, coin(20, "avocado"), coin(22, "plum")),
			},
			bidOrders: []*Order{
				bidOrder(12, coin(30, "acorn"), coin(33, "prune")),
				bidOrder(13, coin(40, "acai"), coin(44, "pear")),
			},
			expErr: joinErrs(
				"cannot settle with multiple ask order asset denoms \"10apple,20avocado\"",
				"cannot settle with multiple ask order price denoms \"11peach,22plum\"",
				"cannot settle with multiple bid order asset denoms \"40acai,30acorn\"",
				"cannot settle with multiple bid order price denoms \"44pear,33prune\"",
			),
		},
		{
			name: "different ask and bid asset denoms",
			askOrders: []*Order{
				askOrder(15, coin(10, "apple"), coin(11, "peach")),
				askOrder(16, coin(20, "apple"), coin(22, "peach")),
			},
			bidOrders: []*Order{
				bidOrder(2001, coin(30, "acorn"), coin(33, "peach")),
				bidOrder(2002, coin(40, "acorn"), coin(44, "peach")),
			},
			expErr: "cannot settle different ask \"30apple\" and bid \"70acorn\" asset denoms",
		},
		{
			name: "different ask and bid price denoms",
			askOrders: []*Order{
				askOrder(15, coin(10, "apple"), coin(11, "peach")),
				askOrder(16, coin(20, "apple"), coin(22, "peach")),
			},
			bidOrders: []*Order{
				bidOrder(2001, coin(30, "apple"), coin(33, "plum")),
				bidOrder(2002, coin(40, "apple"), coin(44, "plum")),
			},
			expErr: "cannot settle different ask \"33peach\" and bid \"77plum\" price denoms",
		},
		{
			name: "different ask and bid denoms",
			askOrders: []*Order{
				askOrder(15, coin(10, "apple"), coin(11, "peach")),
				askOrder(16, coin(20, "apple"), coin(22, "peach")),
			},
			bidOrders: []*Order{
				bidOrder(2001, coin(30, "acorn"), coin(33, "plum")),
				bidOrder(2002, coin(40, "acorn"), coin(44, "plum")),
			},
			expErr: joinErrs(
				"cannot settle different ask \"30apple\" and bid \"70acorn\" asset denoms",
				"cannot settle different ask \"33peach\" and bid \"77plum\" price denoms",
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = validateCanSettle(tc.askOrders, tc.bidOrders)
			}
			require.NotPanics(t, testFunc, "validateCanSettle")
			assertions.AssertErrorValue(t, err, tc.expErr, "validateCanSettle error")
		})
	}
}

func TestAllocateAssets(t *testing.T) {
	askOrder := func(orderID uint64, assetsAmt int64, seller string) *Order {
		return NewOrder(orderID).WithAsk(&AskOrder{
			Seller: seller,
			Assets: sdk.Coin{Denom: "apple", Amount: sdkmath.NewInt(assetsAmt)},
		})
	}
	bidOrder := func(orderID uint64, assetsAmt int64, buyer string) *Order {
		return NewOrder(orderID).WithBid(&BidOrder{
			Buyer:  buyer,
			Assets: sdk.Coin{Denom: "apple", Amount: sdkmath.NewInt(assetsAmt)},
		})
	}
	newOF := func(order *Order, dists ...*Distribution) *OrderFulfillment {
		rv := &OrderFulfillment{
			Order:             order,
			AssetsFilledAmt:   sdkmath.NewInt(0),
			AssetsUnfilledAmt: order.GetAssets().Amount,
		}
		if len(dists) > 0 {
			rv.AssetDists = dists
			for _, d := range dists {
				rv.AssetsFilledAmt = rv.AssetsFilledAmt.Add(d.Amount)
				rv.AssetsUnfilledAmt = rv.AssetsUnfilledAmt.Sub(d.Amount)
			}
		}
		return rv
	}
	dist := func(addr string, amount int64) *Distribution {
		return &Distribution{Address: addr, Amount: sdkmath.NewInt(amount)}
	}

	tests := []struct {
		name      string
		askOFs    []*OrderFulfillment
		bidOFs    []*OrderFulfillment
		expAskOFs []*OrderFulfillment
		expBidOfs []*OrderFulfillment
		expErr    string
	}{
		{
			name:      "one ask, one bid: both full",
			askOFs:    []*OrderFulfillment{newOF(askOrder(5, 10, "seller"))},
			bidOFs:    []*OrderFulfillment{newOF(bidOrder(6, 10, "buyer"))},
			expAskOFs: []*OrderFulfillment{newOF(askOrder(5, 10, "seller"), dist("buyer", 10))},
			expBidOfs: []*OrderFulfillment{newOF(bidOrder(6, 10, "buyer"), dist("seller", 10))},
		},
		{
			name:      "one ask, one bid: ask partial",
			askOFs:    []*OrderFulfillment{newOF(askOrder(5, 11, "seller"))},
			bidOFs:    []*OrderFulfillment{newOF(bidOrder(16, 10, "buyer"))},
			expAskOFs: []*OrderFulfillment{newOF(askOrder(5, 11, "seller"), dist("buyer", 10))},
			expBidOfs: []*OrderFulfillment{newOF(bidOrder(16, 10, "buyer"), dist("seller", 10))},
		},
		{
			name:      "one ask, one bid: bid partial",
			askOFs:    []*OrderFulfillment{newOF(askOrder(15, 10, "seller"))},
			bidOFs:    []*OrderFulfillment{newOF(bidOrder(6, 11, "buyer"))},
			expAskOFs: []*OrderFulfillment{newOF(askOrder(15, 10, "seller"), dist("buyer", 10))},
			expBidOfs: []*OrderFulfillment{newOF(bidOrder(6, 11, "buyer"), dist("seller", 10))},
		},
		{
			name:   "one ask, two bids: last bid not touched",
			askOFs: []*OrderFulfillment{newOF(askOrder(22, 10, "seller"))},
			bidOFs: []*OrderFulfillment{
				newOF(bidOrder(64, 12, "buyer64")),
				newOF(bidOrder(78, 1, "buyer78")),
			},
			expAskOFs: []*OrderFulfillment{newOF(askOrder(22, 10, "seller"), dist("buyer64", 10))},
			expBidOfs: []*OrderFulfillment{
				newOF(bidOrder(64, 12, "buyer64"), dist("seller", 10)),
				newOF(bidOrder(78, 1, "buyer78")),
			},
		},
		{
			name: "two asks, one bids: last ask not touched",
			askOFs: []*OrderFulfillment{
				newOF(askOrder(888, 10, "seller888")),
				newOF(askOrder(999, 10, "seller999")),
			},
			bidOFs: []*OrderFulfillment{newOF(bidOrder(6, 10, "buyer"))},
			expAskOFs: []*OrderFulfillment{
				newOF(askOrder(888, 10, "seller888"), dist("buyer", 10)),
				newOF(askOrder(999, 10, "seller999")),
			},
			expBidOfs: []*OrderFulfillment{newOF(bidOrder(6, 10, "buyer"), dist("seller888", 10))},
		},
		{
			name: "two asks, three bids: both full",
			askOFs: []*OrderFulfillment{
				newOF(askOrder(101, 15, "seller101")),
				newOF(askOrder(102, 25, "seller102")),
			},
			bidOFs: []*OrderFulfillment{
				newOF(bidOrder(103, 10, "buyer103")),
				newOF(bidOrder(104, 8, "buyer104")),
				newOF(bidOrder(105, 22, "buyer105")),
			},
			expAskOFs: []*OrderFulfillment{
				newOF(askOrder(101, 15, "seller101"), dist("buyer103", 10), dist("buyer104", 5)),
				newOF(askOrder(102, 25, "seller102"), dist("buyer104", 3), dist("buyer105", 22)),
			},
			expBidOfs: []*OrderFulfillment{
				newOF(bidOrder(103, 10, "buyer103"), dist("seller101", 10)),
				newOF(bidOrder(104, 8, "buyer104"), dist("seller101", 5), dist("seller102", 3)),
				newOF(bidOrder(105, 22, "buyer105"), dist("seller102", 22)),
			},
		},
		{
			name: "two asks, three bids: ask partial",
			askOFs: []*OrderFulfillment{
				newOF(askOrder(101, 15, "seller101")),
				newOF(askOrder(102, 26, "seller102")),
			},
			bidOFs: []*OrderFulfillment{
				newOF(bidOrder(103, 10, "buyer103")),
				newOF(bidOrder(104, 8, "buyer104")),
				newOF(bidOrder(105, 22, "buyer105")),
			},
			expAskOFs: []*OrderFulfillment{
				newOF(askOrder(101, 15, "seller101"), dist("buyer103", 10), dist("buyer104", 5)),
				newOF(askOrder(102, 26, "seller102"), dist("buyer104", 3), dist("buyer105", 22)),
			},
			expBidOfs: []*OrderFulfillment{
				newOF(bidOrder(103, 10, "buyer103"), dist("seller101", 10)),
				newOF(bidOrder(104, 8, "buyer104"), dist("seller101", 5), dist("seller102", 3)),
				newOF(bidOrder(105, 22, "buyer105"), dist("seller102", 22)),
			},
		},
		{
			name: "two asks, three bids: bid partial",
			askOFs: []*OrderFulfillment{
				newOF(askOrder(101, 15, "seller101")),
				newOF(askOrder(102, 25, "seller102")),
			},
			bidOFs: []*OrderFulfillment{
				newOF(bidOrder(103, 10, "buyer103")),
				newOF(bidOrder(104, 8, "buyer104")),
				newOF(bidOrder(105, 23, "buyer105")),
			},
			expAskOFs: []*OrderFulfillment{
				newOF(askOrder(101, 15, "seller101"), dist("buyer103", 10), dist("buyer104", 5)),
				newOF(askOrder(102, 25, "seller102"), dist("buyer104", 3), dist("buyer105", 22)),
			},
			expBidOfs: []*OrderFulfillment{
				newOF(bidOrder(103, 10, "buyer103"), dist("seller101", 10)),
				newOF(bidOrder(104, 8, "buyer104"), dist("seller101", 5), dist("seller102", 3)),
				newOF(bidOrder(105, 23, "buyer105"), dist("seller102", 22)),
			},
		},
		{
			name:   "negative ask assets unfilled",
			askOFs: []*OrderFulfillment{newOF(askOrder(101, 10, "seller"), dist("buyerx", 11))},
			bidOFs: []*OrderFulfillment{newOF(bidOrder(102, 10, "buyer"))},
			expErr: "cannot fill ask order 101 having assets left \"-1apple\" with bid order 102 having " +
				"assets left \"10apple\": zero or negative assets left",
		},
		{
			name:   "negative bid assets unfilled",
			askOFs: []*OrderFulfillment{newOF(askOrder(101, 10, "seller"))},
			bidOFs: []*OrderFulfillment{newOF(bidOrder(102, 10, "buyer"), dist("sellerx", 11))},
			expErr: "cannot fill ask order 101 having assets left \"10apple\" with bid order 102 having " +
				"assets left \"-1apple\": zero or negative assets left",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			origAskOFs := copyOrderFulfillments(tc.askOFs)
			origBidOFs := copyOrderFulfillments(tc.bidOFs)

			var err error
			testFunc := func() {
				err = allocateAssets(tc.askOFs, tc.bidOFs)
			}
			require.NotPanics(t, testFunc, "allocateAssets")
			assertions.AssertErrorValue(t, err, tc.expErr, "allocateAssets error")
			if len(tc.expErr) > 0 {
				return
			}
			if !assertEqualOrderFulfillmentSlices(t, tc.expAskOFs, tc.askOFs, "askOFs after allocateAssets") {
				t.Logf("Original: %s", orderFulfillmentsString(origAskOFs))
			}
			if !assertEqualOrderFulfillmentSlices(t, tc.expBidOfs, tc.bidOFs, "bidOFs after allocateAssets") {
				t.Logf("Original: %s", orderFulfillmentsString(origBidOFs))
			}
		})
	}
}

func TestSplitPartial(t *testing.T) {
	askOrder := func(orderID uint64, assetsAmt int64, seller string) *Order {
		return NewOrder(orderID).WithAsk(&AskOrder{
			Seller:       seller,
			Assets:       sdk.Coin{Denom: "apple", Amount: sdkmath.NewInt(assetsAmt)},
			Price:        sdk.Coin{Denom: "peach", Amount: sdkmath.NewInt(assetsAmt)},
			AllowPartial: true,
		})
	}
	bidOrder := func(orderID uint64, assetsAmt int64, buyer string) *Order {
		return NewOrder(orderID).WithBid(&BidOrder{
			Buyer:        buyer,
			Assets:       sdk.Coin{Denom: "apple", Amount: sdkmath.NewInt(assetsAmt)},
			Price:        sdk.Coin{Denom: "peach", Amount: sdkmath.NewInt(assetsAmt)},
			AllowPartial: true,
		})
	}
	newOF := func(order *Order, dists ...*Distribution) *OrderFulfillment {
		rv := &OrderFulfillment{
			Order:             order,
			AssetsFilledAmt:   sdkmath.NewInt(0),
			AssetsUnfilledAmt: order.GetAssets().Amount,
			PriceAppliedAmt:   sdkmath.NewInt(0),
			PriceLeftAmt:      order.GetPrice().Amount,
		}
		if len(dists) > 0 {
			rv.AssetDists = dists
			for _, d := range dists {
				rv.AssetsFilledAmt = rv.AssetsFilledAmt.Add(d.Amount)
				rv.AssetsUnfilledAmt = rv.AssetsUnfilledAmt.Sub(d.Amount)
			}
			if rv.AssetsUnfilledAmt.IsZero() {
				rv.AssetsUnfilledAmt = sdkmath.NewInt(0)
			}
		}
		return rv
	}
	dist := func(addr string, amount int64) *Distribution {
		return &Distribution{Address: addr, Amount: sdkmath.NewInt(amount)}
	}

	tests := []struct {
		name          string
		askOFs        []*OrderFulfillment
		bidOFs        []*OrderFulfillment
		settlement    *Settlement
		expAskOFs     []*OrderFulfillment
		expBidOfs     []*OrderFulfillment
		expSettlement *Settlement
		expErr        string
	}{
		{
			name:       "one ask: not touched",
			askOFs:     []*OrderFulfillment{newOF(askOrder(8, 10, "seller8"))},
			settlement: &Settlement{},
			expErr:     "ask order 8 (at index 0) has no assets filled",
		},
		{
			name:          "one ask: partial",
			askOFs:        []*OrderFulfillment{newOF(askOrder(8, 10, "seller8"), dist("buyer", 7))},
			settlement:    &Settlement{},
			expAskOFs:     []*OrderFulfillment{newOF(askOrder(8, 7, "seller8"), dist("buyer", 7))},
			expSettlement: &Settlement{PartialOrderLeft: askOrder(8, 3, "seller8")},
		},
		{
			name:       "one ask: partial, settlement already has a partial",
			askOFs:     []*OrderFulfillment{newOF(askOrder(8, 10, "seller8"), dist("buyer", 7))},
			settlement: &Settlement{PartialOrderLeft: bidOrder(55, 3, "buyer")},
			expErr:     "bid order 55 and ask order 8 cannot both be partially filled",
		},
		{
			name: "one ask: partial, not allowed",
			askOFs: []*OrderFulfillment{
				newOF(NewOrder(8).WithAsk(&AskOrder{
					Seller:       "seller8",
					Assets:       sdk.Coin{Denom: "apple", Amount: sdkmath.NewInt(10)},
					Price:        sdk.Coin{Denom: "peach", Amount: sdkmath.NewInt(10)},
					AllowPartial: false,
				}), dist("buyer", 7))},
			settlement: &Settlement{},
			expErr:     "cannot split ask order 8 having assets \"10apple\" at \"7apple\": order does not allow partial fulfillment",
		},
		{
			name: "two asks: first partial",
			askOFs: []*OrderFulfillment{
				newOF(askOrder(8, 10, "seller8"), dist("buyer", 7)),
				newOF(askOrder(9, 12, "seller8")),
			},
			settlement: &Settlement{},
			expErr:     "ask order 8 (at index 0) is not filled in full and is not the last ask order provided",
		},
		{
			name: "two asks: last untouched",
			askOFs: []*OrderFulfillment{
				newOF(askOrder(8, 10, "seller8"), dist("buyer", 10)),
				newOF(askOrder(9, 12, "seller8")),
			},
			settlement: &Settlement{},
			expErr:     "ask order 9 (at index 1) has no assets filled",
		},
		{
			name: "two asks: last partial",
			askOFs: []*OrderFulfillment{
				newOF(askOrder(8, 10, "seller8"), dist("buyer", 10)),
				newOF(askOrder(9, 12, "seller9"), dist("buyer", 10)),
			},
			settlement: &Settlement{},
			expAskOFs: []*OrderFulfillment{
				newOF(askOrder(8, 10, "seller8"), dist("buyer", 10)),
				newOF(askOrder(9, 10, "seller9"), dist("buyer", 10)),
			},
			expSettlement: &Settlement{PartialOrderLeft: askOrder(9, 2, "seller9")},
		},

		{
			name:       "one bid: not touched",
			bidOFs:     []*OrderFulfillment{newOF(bidOrder(8, 10, "buyer8"))},
			settlement: &Settlement{},
			expErr:     "bid order 8 (at index 0) has no assets filled",
		},
		{
			name:          "one bid: partial",
			bidOFs:        []*OrderFulfillment{newOF(bidOrder(8, 10, "buyer8"), dist("seller", 7))},
			settlement:    &Settlement{},
			expBidOfs:     []*OrderFulfillment{newOF(bidOrder(8, 7, "buyer8"), dist("seller", 7))},
			expSettlement: &Settlement{PartialOrderLeft: bidOrder(8, 3, "buyer8")},
		},
		{
			name: "one bid: partial, not allowed",
			askOFs: []*OrderFulfillment{
				newOF(NewOrder(8).WithBid(&BidOrder{
					Buyer:        "buyer8",
					Assets:       sdk.Coin{Denom: "apple", Amount: sdkmath.NewInt(10)},
					Price:        sdk.Coin{Denom: "peach", Amount: sdkmath.NewInt(10)},
					AllowPartial: false,
				}), dist("seller", 7))},
			settlement: &Settlement{},
			expErr:     "cannot split bid order 8 having assets \"10apple\" at \"7apple\": order does not allow partial fulfillment",
		},
		{
			name: "two bids: first partial",
			bidOFs: []*OrderFulfillment{
				newOF(bidOrder(8, 10, "buyer8"), dist("seller", 7)),
				newOF(bidOrder(9, 12, "buyer9")),
			},
			settlement: &Settlement{},
			expErr:     "bid order 8 (at index 0) is not filled in full and is not the last bid order provided",
		},
		{
			name: "two bids: last untouched",
			bidOFs: []*OrderFulfillment{
				newOF(bidOrder(8, 10, "buyer8"), dist("seller", 10)),
				newOF(bidOrder(9, 12, "buyer9")),
			},
			settlement: &Settlement{},
			expErr:     "bid order 9 (at index 1) has no assets filled",
		},
		{
			name: "two bids: last partial",
			bidOFs: []*OrderFulfillment{
				newOF(bidOrder(8, 10, "buyer8"), dist("seller", 10)),
				newOF(bidOrder(9, 12, "buyer9"), dist("seller", 10)),
			},
			settlement: &Settlement{},
			expBidOfs: []*OrderFulfillment{
				newOF(bidOrder(8, 10, "buyer8"), dist("seller", 10)),
				newOF(bidOrder(9, 10, "buyer9"), dist("seller", 10)),
			},
			expSettlement: &Settlement{PartialOrderLeft: bidOrder(9, 2, "buyer9")},
		},
		{
			name:       "one ask, one bid: both partial",
			askOFs:     []*OrderFulfillment{newOF(askOrder(8, 10, "seller8"), dist("buyer9", 7))},
			bidOFs:     []*OrderFulfillment{newOF(bidOrder(9, 10, "buyer9"), dist("seller8", 7))},
			settlement: &Settlement{},
			expErr:     "ask order 8 and bid order 9 cannot both be partially filled",
		},
		{
			name: "three asks, three bids: no partial",
			askOFs: []*OrderFulfillment{
				newOF(askOrder(51, 10, "seller51"), dist("buyer99", 10)),
				newOF(askOrder(77, 15, "seller77"), dist("buyer99", 10), dist("buyer8", 5)),
				newOF(askOrder(12, 20, "seller12"), dist("buyer8", 10), dist("buyer112", 10)),
			},
			bidOFs: []*OrderFulfillment{
				newOF(bidOrder(99, 20, "buyer99"), dist("seller51", 10), dist("seller77", 10)),
				newOF(bidOrder(8, 15, "buyer8"), dist("seller77", 5), dist("seller12", 10)),
				newOF(bidOrder(112, 10, "buyer112"), dist("seller12", 10)),
			},
			settlement: &Settlement{},
			expAskOFs: []*OrderFulfillment{
				newOF(askOrder(51, 10, "seller51"), dist("buyer99", 10)),
				newOF(askOrder(77, 15, "seller77"), dist("buyer99", 10), dist("buyer8", 5)),
				newOF(askOrder(12, 20, "seller12"), dist("buyer8", 10), dist("buyer112", 10)),
			},
			expBidOfs: []*OrderFulfillment{
				newOF(bidOrder(99, 20, "buyer99"), dist("seller51", 10), dist("seller77", 10)),
				newOF(bidOrder(8, 15, "buyer8"), dist("seller77", 5), dist("seller12", 10)),
				newOF(bidOrder(112, 10, "buyer112"), dist("seller12", 10)),
			},
			expSettlement: &Settlement{PartialOrderLeft: nil},
		},
		{
			name: "three asks, three bids: partial ask",
			askOFs: []*OrderFulfillment{
				newOF(askOrder(51, 10, "seller51"), dist("buyer99", 10)),
				newOF(askOrder(77, 15, "seller77"), dist("buyer99", 10), dist("buyer8", 5)),
				newOF(askOrder(12, 21, "seller12"), dist("buyer8", 10), dist("buyer112", 10)),
			},
			bidOFs: []*OrderFulfillment{
				newOF(bidOrder(99, 20, "buyer99"), dist("seller51", 10), dist("seller77", 10)),
				newOF(bidOrder(8, 15, "buyer8"), dist("seller77", 5), dist("seller12", 10)),
				newOF(bidOrder(112, 10, "buyer112"), dist("seller12", 10)),
			},
			settlement: &Settlement{},
			expAskOFs: []*OrderFulfillment{
				newOF(askOrder(51, 10, "seller51"), dist("buyer99", 10)),
				newOF(askOrder(77, 15, "seller77"), dist("buyer99", 10), dist("buyer8", 5)),
				newOF(askOrder(12, 20, "seller12"), dist("buyer8", 10), dist("buyer112", 10)),
			},
			expBidOfs: []*OrderFulfillment{
				newOF(bidOrder(99, 20, "buyer99"), dist("seller51", 10), dist("seller77", 10)),
				newOF(bidOrder(8, 15, "buyer8"), dist("seller77", 5), dist("seller12", 10)),
				newOF(bidOrder(112, 10, "buyer112"), dist("seller12", 10)),
			},
			expSettlement: &Settlement{PartialOrderLeft: askOrder(12, 1, "seller12")},
		},
		{
			name: "three asks, three bids: no partial",
			askOFs: []*OrderFulfillment{
				newOF(askOrder(51, 10, "seller51"), dist("buyer99", 10)),
				newOF(askOrder(77, 15, "seller77"), dist("buyer99", 10), dist("buyer8", 5)),
				newOF(askOrder(12, 20, "seller12"), dist("buyer8", 10), dist("buyer112", 10)),
			},
			bidOFs: []*OrderFulfillment{
				newOF(bidOrder(99, 20, "buyer99"), dist("seller51", 10), dist("seller77", 10)),
				newOF(bidOrder(8, 15, "buyer8"), dist("seller77", 5), dist("seller12", 10)),
				newOF(bidOrder(112, 11, "buyer112"), dist("seller12", 10)),
			},
			settlement: &Settlement{},
			expAskOFs: []*OrderFulfillment{
				newOF(askOrder(51, 10, "seller51"), dist("buyer99", 10)),
				newOF(askOrder(77, 15, "seller77"), dist("buyer99", 10), dist("buyer8", 5)),
				newOF(askOrder(12, 20, "seller12"), dist("buyer8", 10), dist("buyer112", 10)),
			},
			expBidOfs: []*OrderFulfillment{
				newOF(bidOrder(99, 20, "buyer99"), dist("seller51", 10), dist("seller77", 10)),
				newOF(bidOrder(8, 15, "buyer8"), dist("seller77", 5), dist("seller12", 10)),
				newOF(bidOrder(112, 10, "buyer112"), dist("seller12", 10)),
			},
			expSettlement: &Settlement{PartialOrderLeft: bidOrder(112, 1, "buyer112")},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			origAskOFs := copyOrderFulfillments(tc.askOFs)
			origBidOFs := copyOrderFulfillments(tc.bidOFs)

			var err error
			testFunc := func() {
				err = splitPartial(tc.askOFs, tc.bidOFs, tc.settlement)
			}
			require.NotPanics(t, testFunc, "splitPartial")
			assertions.AssertErrorValue(t, err, tc.expErr, "splitPartial error")
			if len(tc.expErr) > 0 {
				return
			}
			if !assertEqualOrderFulfillmentSlices(t, tc.expAskOFs, tc.askOFs, "askOFs after splitPartial") {
				t.Logf("Original: %s", orderFulfillmentsString(origAskOFs))
			}
			if !assertEqualOrderFulfillmentSlices(t, tc.expBidOfs, tc.bidOFs, "bidOFs after splitPartial") {
				t.Logf("Original: %s", orderFulfillmentsString(origBidOFs))
			}
			assert.Equalf(t, tc.expSettlement, tc.settlement, "settlement after splitPartial")
		})
	}
}

func TestSplitOrderFulfillments(t *testing.T) {
	acoin := func(amount int64) sdk.Coin {
		return sdk.NewInt64Coin("acorn", amount)
	}
	pcoin := func(amount int64) sdk.Coin {
		return sdk.NewInt64Coin("prune", amount)
	}
	askOrder := func(orderID uint64, assetsAmt int64, allowPartial bool) *Order {
		return NewOrder(orderID).WithAsk(&AskOrder{
			MarketId:     123,
			Seller:       "sEllEr",
			Assets:       acoin(assetsAmt),
			Price:        pcoin(assetsAmt),
			AllowPartial: allowPartial,
		})
	}
	bidOrder := func(orderID uint64, assetsAmt int64, allowPartial bool) *Order {
		return NewOrder(orderID).WithBid(&BidOrder{
			MarketId:     123,
			Buyer:        "bUyEr",
			Assets:       acoin(assetsAmt),
			Price:        pcoin(assetsAmt),
			AllowPartial: allowPartial,
		})
	}
	newOF := func(order *Order, assetsFilledAmt int64) *OrderFulfillment {
		rv := &OrderFulfillment{
			Order:             order,
			AssetsFilledAmt:   sdkmath.NewInt(assetsFilledAmt),
			AssetsUnfilledAmt: order.GetAssets().Amount.SubRaw(assetsFilledAmt),
			PriceAppliedAmt:   sdkmath.NewInt(assetsFilledAmt),
			PriceLeftAmt:      order.GetPrice().Amount.SubRaw(assetsFilledAmt),
		}
		// int(x).Sub(x) results in an object that is not .Equal to ZeroInt().
		// The Split function sets this to ZeroInt().
		if rv.AssetsUnfilledAmt.IsZero() {
			rv.AssetsUnfilledAmt = sdkmath.ZeroInt()
		}
		return rv
	}

	tests := []struct {
		name            string
		fulfillments    []*OrderFulfillment
		settlement      *Settlement
		expFulfillments []*OrderFulfillment
		expSettlement   *Settlement
		expErr          string
	}{
		{
			name:         "one order, ask: nothing filled",
			fulfillments: []*OrderFulfillment{newOF(askOrder(8, 53, false), 0)},
			settlement:   &Settlement{},
			expErr:       "ask order 8 (at index 0) has no assets filled",
		},
		{
			name:         "one order, bid: nothing filled",
			fulfillments: []*OrderFulfillment{newOF(bidOrder(8, 53, false), 0)},
			settlement:   &Settlement{},
			expErr:       "bid order 8 (at index 0) has no assets filled",
		},
		{
			name:            "one order, ask: partially filled",
			fulfillments:    []*OrderFulfillment{newOF(askOrder(8, 53, true), 13)},
			settlement:      &Settlement{},
			expFulfillments: []*OrderFulfillment{newOF(askOrder(8, 13, true), 13)},
			expSettlement:   &Settlement{PartialOrderLeft: askOrder(8, 40, true)},
		},
		{
			name:            "one order, bid: partially filled",
			fulfillments:    []*OrderFulfillment{newOF(bidOrder(8, 53, true), 13)},
			settlement:      &Settlement{},
			expFulfillments: []*OrderFulfillment{newOF(bidOrder(8, 13, true), 13)},
			expSettlement:   &Settlement{PartialOrderLeft: bidOrder(8, 40, true)},
		},
		{
			name:         "one order, ask: partially filled, already have a partially filled",
			fulfillments: []*OrderFulfillment{newOF(askOrder(8, 53, true), 13)},
			settlement:   &Settlement{PartialOrderLeft: bidOrder(951, 357, true)},
			expErr:       "bid order 951 and ask order 8 cannot both be partially filled",
		},
		{
			name:         "one order, bid: partially filled, already have a partially filled",
			fulfillments: []*OrderFulfillment{newOF(bidOrder(8, 53, true), 13)},
			settlement:   &Settlement{PartialOrderLeft: askOrder(951, 357, true)},
			expErr:       "ask order 951 and bid order 8 cannot both be partially filled",
		},
		{
			name:         "one order, ask: partially filled, split not allowed",
			fulfillments: []*OrderFulfillment{newOF(askOrder(8, 53, false), 13)},
			settlement:   &Settlement{},
			expErr:       "cannot split ask order 8 having assets \"53acorn\" at \"13acorn\": order does not allow partial fulfillment",
		},
		{
			name:         "one order, bid: partially filled, split not allowed",
			fulfillments: []*OrderFulfillment{newOF(bidOrder(8, 53, false), 13)},
			settlement:   &Settlement{},
			expErr:       "cannot split bid order 8 having assets \"53acorn\" at \"13acorn\": order does not allow partial fulfillment",
		},
		{
			name:            "one order, ask: fully filled",
			fulfillments:    []*OrderFulfillment{newOF(askOrder(8, 53, false), 53)},
			settlement:      &Settlement{},
			expFulfillments: []*OrderFulfillment{newOF(askOrder(8, 53, false), 53)},
			expSettlement:   &Settlement{},
		},
		{
			name:            "one order, bid: fully filled",
			fulfillments:    []*OrderFulfillment{newOF(bidOrder(8, 53, false), 53)},
			settlement:      &Settlement{},
			expFulfillments: []*OrderFulfillment{newOF(bidOrder(8, 53, false), 53)},
			expSettlement:   &Settlement{},
		},
		{
			name:            "one order, ask: fully filled, already have a partially filled",
			fulfillments:    []*OrderFulfillment{newOF(askOrder(8, 53, false), 53)},
			settlement:      &Settlement{PartialOrderLeft: bidOrder(951, 357, true)},
			expFulfillments: []*OrderFulfillment{newOF(askOrder(8, 53, false), 53)},
			expSettlement:   &Settlement{PartialOrderLeft: bidOrder(951, 357, true)},
		},
		{
			name:            "one order, bid: fully filled, already have a partially filled",
			fulfillments:    []*OrderFulfillment{newOF(bidOrder(8, 53, false), 53)},
			settlement:      &Settlement{PartialOrderLeft: askOrder(951, 357, true)},
			expFulfillments: []*OrderFulfillment{newOF(bidOrder(8, 53, false), 53)},
			expSettlement:   &Settlement{PartialOrderLeft: askOrder(951, 357, true)},
		},
		{
			name: "three orders, ask: second partially filled",
			fulfillments: []*OrderFulfillment{
				newOF(askOrder(8, 53, false), 53),
				newOF(askOrder(9, 17, true), 16),
				newOF(askOrder(10, 200, false), 0),
			},
			settlement: &Settlement{},
			expErr:     "ask order 9 (at index 1) is not filled in full and is not the last ask order provided",
		},
		{
			name: "three orders, bid: second partially filled",
			fulfillments: []*OrderFulfillment{
				newOF(bidOrder(8, 53, false), 53),
				newOF(bidOrder(9, 17, true), 16),
				newOF(bidOrder(10, 200, false), 0),
			},
			settlement: &Settlement{},
			expErr:     "bid order 9 (at index 1) is not filled in full and is not the last bid order provided",
		},
		{
			name: "three orders, ask: last not touched",
			fulfillments: []*OrderFulfillment{
				newOF(askOrder(8, 53, false), 53),
				newOF(askOrder(9, 17, false), 17),
				newOF(askOrder(10, 200, false), 0),
			},
			settlement: &Settlement{},
			expErr:     "ask order 10 (at index 2) has no assets filled",
		},
		{
			name: "three orders, bid: last not touched",
			fulfillments: []*OrderFulfillment{
				newOF(bidOrder(8, 53, false), 53),
				newOF(bidOrder(9, 17, false), 17),
				newOF(bidOrder(10, 200, true), 0),
			},
			settlement: &Settlement{},
			expErr:     "bid order 10 (at index 2) has no assets filled",
		},
		{
			name: "three orders, ask: last partially filled",
			fulfillments: []*OrderFulfillment{
				newOF(askOrder(8, 53, false), 53),
				newOF(askOrder(9, 17, false), 17),
				newOF(askOrder(10, 200, true), 183),
			},
			settlement: &Settlement{},
			expFulfillments: []*OrderFulfillment{
				newOF(askOrder(8, 53, false), 53),
				newOF(askOrder(9, 17, false), 17),
				newOF(askOrder(10, 183, true), 183),
			},
			expSettlement: &Settlement{PartialOrderLeft: askOrder(10, 17, true)},
		},
		{
			name: "three orders, bid: last partially filled",
			fulfillments: []*OrderFulfillment{
				newOF(bidOrder(8, 53, false), 53),
				newOF(bidOrder(9, 17, false), 17),
				newOF(bidOrder(10, 200, true), 183),
			},
			settlement: &Settlement{},
			expFulfillments: []*OrderFulfillment{
				newOF(bidOrder(8, 53, false), 53),
				newOF(bidOrder(9, 17, false), 17),
				newOF(bidOrder(10, 183, true), 183),
			},
			expSettlement: &Settlement{PartialOrderLeft: bidOrder(10, 17, true)},
		},
		{
			name: "three orders, ask: last partially filled, split not allowed",
			fulfillments: []*OrderFulfillment{
				newOF(askOrder(8, 53, true), 53),
				newOF(askOrder(9, 17, true), 17),
				newOF(askOrder(10, 200, false), 183),
			},
			settlement: &Settlement{},
			expErr:     "cannot split ask order 10 having assets \"200acorn\" at \"183acorn\": order does not allow partial fulfillment",
		},
		{
			name: "three orders, bid: last partially filled, split not allowed",
			fulfillments: []*OrderFulfillment{
				newOF(bidOrder(8, 53, true), 53),
				newOF(bidOrder(9, 17, true), 17),
				newOF(bidOrder(10, 200, false), 183),
			},
			settlement: &Settlement{},
			expErr:     "cannot split bid order 10 having assets \"200acorn\" at \"183acorn\": order does not allow partial fulfillment",
		},
		{
			name: "three orders, ask: last partially filled, already have a partially filled",
			fulfillments: []*OrderFulfillment{
				newOF(askOrder(8, 53, true), 53),
				newOF(askOrder(9, 17, true), 17),
				newOF(askOrder(10, 200, false), 183),
			},
			settlement: &Settlement{PartialOrderLeft: bidOrder(857, 43, true)},
			expErr:     "bid order 857 and ask order 10 cannot both be partially filled",
		},
		{
			name: "three orders, bid: last partially filled, already have a partially filled",
			fulfillments: []*OrderFulfillment{
				newOF(bidOrder(8, 53, true), 53),
				newOF(bidOrder(9, 17, true), 17),
				newOF(bidOrder(10, 200, false), 183),
			},
			settlement: &Settlement{PartialOrderLeft: askOrder(857, 43, true)},
			expErr:     "ask order 857 and bid order 10 cannot both be partially filled",
		},
		{
			name: "three orders, ask: fully filled",
			fulfillments: []*OrderFulfillment{
				newOF(askOrder(8, 53, false), 53),
				newOF(askOrder(9, 17, false), 17),
				newOF(askOrder(10, 200, false), 200),
			},
			settlement: &Settlement{},
			expFulfillments: []*OrderFulfillment{
				newOF(askOrder(8, 53, false), 53),
				newOF(askOrder(9, 17, false), 17),
				newOF(askOrder(10, 200, false), 200),
			},
			expSettlement: &Settlement{},
		},
		{
			name: "three orders, bid: fully filled",
			fulfillments: []*OrderFulfillment{
				newOF(bidOrder(8, 53, false), 53),
				newOF(bidOrder(9, 17, false), 17),
				newOF(bidOrder(10, 200, false), 200),
			},
			settlement: &Settlement{},
			expFulfillments: []*OrderFulfillment{
				newOF(bidOrder(8, 53, false), 53),
				newOF(bidOrder(9, 17, false), 17),
				newOF(bidOrder(10, 200, false), 200),
			},
			expSettlement: &Settlement{},
		},
		{
			name: "three orders, ask: fully filled, already have a partially filled",
			fulfillments: []*OrderFulfillment{
				newOF(askOrder(8, 53, false), 53),
				newOF(askOrder(9, 17, false), 17),
				newOF(askOrder(10, 200, false), 200),
			},
			settlement: &Settlement{},
			expFulfillments: []*OrderFulfillment{
				newOF(askOrder(8, 53, false), 53),
				newOF(askOrder(9, 17, false), 17),
				newOF(askOrder(10, 200, false), 200),
			},
			expSettlement: &Settlement{},
		},
		{
			name: "three orders, bid: fully filled, already have a partially filled",
			fulfillments: []*OrderFulfillment{
				newOF(bidOrder(8, 53, false), 53),
				newOF(bidOrder(9, 17, false), 17),
				newOF(bidOrder(10, 200, false), 200),
			},
			settlement: &Settlement{PartialOrderLeft: askOrder(857, 43, true)},
			expFulfillments: []*OrderFulfillment{
				newOF(bidOrder(8, 53, false), 53),
				newOF(bidOrder(9, 17, false), 17),
				newOF(bidOrder(10, 200, false), 200),
			},
			expSettlement: &Settlement{PartialOrderLeft: askOrder(857, 43, true)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = splitOrderFulfillments(tc.fulfillments, tc.settlement)
			}
			require.NotPanics(t, testFunc, "splitOrderFulfillments")
			assertions.AssertErrorValue(t, err, tc.expErr, "splitOrderFulfillments error")
			if len(tc.expErr) > 0 {
				return
			}
			assertEqualOrderFulfillmentSlices(t, tc.expFulfillments, tc.fulfillments, "fulfillments after splitOrderFulfillments")
			assert.Equal(t, tc.expSettlement, tc.settlement, "settlement after splitOrderFulfillments")
		})
	}
}

func TestAllocatePrice(t *testing.T) {
	// TODO[1658]: func TestAllocatePrice(t *testing.T)
	t.Skipf("not written")
}

func TestSetFeesToPay(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	askOF := func(orderID uint64, priceAppliedAmt int64, fees ...sdk.Coin) *OrderFulfillment {
		askOrder := &AskOrder{Price: coin(50, "plum")}
		if len(fees) > 1 {
			t.Fatalf("cannot provide more than one fee to askOF(%d, %d, %q)",
				orderID, priceAppliedAmt, fees)
		}
		if len(fees) > 0 {
			askOrder.SellerSettlementFlatFee = &fees[0]
		}
		return &OrderFulfillment{
			Order:           NewOrder(orderID).WithAsk(askOrder),
			PriceAppliedAmt: sdkmath.NewInt(priceAppliedAmt),
		}
	}
	bidOF := func(orderID uint64, priceAppliedAmt int64, fees ...sdk.Coin) *OrderFulfillment {
		bidOrder := &BidOrder{Price: coin(50, "plum")}
		if len(fees) > 0 {
			bidOrder.BuyerSettlementFees = fees
		}
		return &OrderFulfillment{
			Order:           NewOrder(orderID).WithBid(bidOrder),
			PriceAppliedAmt: sdkmath.NewInt(priceAppliedAmt),
		}
	}
	expOF := func(f *OrderFulfillment, feesToPay ...sdk.Coin) *OrderFulfillment {
		if len(feesToPay) > 0 {
			f.FeesToPay = sdk.NewCoins(feesToPay...)
		}
		return f
	}

	tests := []struct {
		name      string
		askOFs    []*OrderFulfillment
		bidOFs    []*OrderFulfillment
		ratio     *FeeRatio
		expAskOFs []*OrderFulfillment
		expBidOFs []*OrderFulfillment
		expErr    string
	}{
		{
			name: "cannot apply ratio",
			askOFs: []*OrderFulfillment{
				askOF(7777, 55, coin(20, "grape")),
				askOF(5555, 71),
				askOF(6666, 100),
			},
			bidOFs: []*OrderFulfillment{
				bidOF(1111, 100),
				bidOF(2222, 200, coin(20, "grape")),
				bidOF(3333, 300),
			},
			ratio: &FeeRatio{Price: coin(30, "peach"), Fee: coin(1, "fig")},
			expAskOFs: []*OrderFulfillment{
				expOF(askOF(7777, 55, coin(20, "grape"))),
				expOF(askOF(5555, 71)),
				expOF(askOF(6666, 100)),
			},
			expBidOFs: []*OrderFulfillment{
				expOF(bidOF(1111, 100)),
				expOF(bidOF(2222, 200, coin(20, "grape")), coin(20, "grape")),
				expOF(bidOF(3333, 300)),
			},
			expErr: joinErrs(
				"failed calculate ratio fee for ask order 7777: cannot apply ratio 30peach:1fig to price 55plum: incorrect price denom",
				"failed calculate ratio fee for ask order 5555: cannot apply ratio 30peach:1fig to price 71plum: incorrect price denom",
				"failed calculate ratio fee for ask order 6666: cannot apply ratio 30peach:1fig to price 100plum: incorrect price denom",
			),
		},
		{
			name: "no ratio",
			askOFs: []*OrderFulfillment{
				askOF(7777, 55, coin(20, "grape")),
				askOF(5555, 71),
				askOF(6666, 100),
			},
			bidOFs: []*OrderFulfillment{
				bidOF(1111, 100),
				bidOF(2222, 200, coin(20, "grape")),
				bidOF(3333, 300),
			},
			ratio: nil,
			expAskOFs: []*OrderFulfillment{
				expOF(askOF(7777, 55, coin(20, "grape")), coin(20, "grape")),
				expOF(askOF(5555, 71)),
				expOF(askOF(6666, 100)),
			},
			expBidOFs: []*OrderFulfillment{
				expOF(bidOF(1111, 100)),
				expOF(bidOF(2222, 200, coin(20, "grape")), coin(20, "grape")),
				expOF(bidOF(3333, 300)),
			},
		},
		{
			name: "with ratio",
			askOFs: []*OrderFulfillment{
				askOF(7777, 55, coin(20, "grape")),
				askOF(5555, 71),
				askOF(6666, 100),
			},
			bidOFs: []*OrderFulfillment{
				bidOF(1111, 100),
				bidOF(2222, 200, coin(20, "grape")),
				bidOF(3333, 300),
			},
			ratio: &FeeRatio{Price: coin(30, "plum"), Fee: coin(1, "fig")},
			expAskOFs: []*OrderFulfillment{
				expOF(askOF(7777, 55, coin(20, "grape")), coin(2, "fig"), coin(20, "grape")),
				expOF(askOF(5555, 71), coin(3, "fig")),
				expOF(askOF(6666, 100), coin(4, "fig")),
			},
			expBidOFs: []*OrderFulfillment{
				expOF(bidOF(1111, 100)),
				expOF(bidOF(2222, 200, coin(20, "grape")), coin(20, "grape")),
				expOF(bidOF(3333, 300)),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = setFeesToPay(tc.askOFs, tc.bidOFs, tc.ratio)
			}
			require.NotPanics(t, testFunc, "setFeesToPay")
			assertions.AssertErrorValue(t, err, tc.expErr, "setFeesToPay error")
			assertEqualOrderFulfillmentSlices(t, tc.expAskOFs, tc.askOFs, "askOFs after setFeesToPay")
			assertEqualOrderFulfillmentSlices(t, tc.expBidOFs, tc.bidOFs, "bidOFs after setFeesToPay")
		})
	}
}

func TestValidateFulfillments(t *testing.T) {
	goodAskOF := func(orderID uint64) *OrderFulfillment {
		return &OrderFulfillment{
			Order: NewOrder(orderID).WithAsk(&AskOrder{
				Assets: sdk.NewInt64Coin("apple", 50),
				Price:  sdk.NewInt64Coin("peach", 123),
			}),
			AssetsFilledAmt: sdkmath.NewInt(50),
			PriceAppliedAmt: sdkmath.NewInt(130),
		}
	}
	badAskOF := func(orderID uint64) *OrderFulfillment {
		rv := goodAskOF(orderID)
		rv.AssetsFilledAmt = sdkmath.NewInt(49)
		return rv
	}
	badAskErr := func(orderID uint64) string {
		return badAskOF(orderID).Validate().Error()
	}
	goodBidOF := func(orderID uint64) *OrderFulfillment {
		return &OrderFulfillment{
			Order: NewOrder(orderID).WithBid(&BidOrder{
				Assets: sdk.NewInt64Coin("apple", 50),
				Price:  sdk.NewInt64Coin("peach", 123),
			}),
			AssetsFilledAmt: sdkmath.NewInt(50),
			PriceAppliedAmt: sdkmath.NewInt(123),
		}
	}
	badBidOF := func(orderID uint64) *OrderFulfillment {
		rv := goodBidOF(orderID)
		rv.AssetsFilledAmt = sdkmath.NewInt(49)
		return rv
	}
	badBidErr := func(orderID uint64) string {
		return badBidOF(orderID).Validate().Error()
	}

	tests := []struct {
		name   string
		askOFs []*OrderFulfillment
		bidOFs []*OrderFulfillment
		expErr string
	}{
		{
			name:   "all good",
			askOFs: []*OrderFulfillment{goodAskOF(10), goodAskOF(11), goodAskOF(12)},
			bidOFs: []*OrderFulfillment{goodBidOF(20), goodBidOF(21), goodBidOF(22)},
			expErr: "",
		},
		{
			name:   "error in one ask",
			askOFs: []*OrderFulfillment{goodAskOF(10), badAskOF(11), goodAskOF(12)},
			bidOFs: []*OrderFulfillment{goodBidOF(20), goodBidOF(21), goodBidOF(22)},
			expErr: badAskErr(11),
		},
		{
			name:   "error in one bid",
			askOFs: []*OrderFulfillment{goodAskOF(10), goodAskOF(11), goodAskOF(12)},
			bidOFs: []*OrderFulfillment{goodBidOF(20), badBidOF(21), goodBidOF(22)},
			expErr: badBidErr(21),
		},
		{
			name:   "two errors in asks",
			askOFs: []*OrderFulfillment{badAskOF(10), goodAskOF(11), badAskOF(12)},
			bidOFs: []*OrderFulfillment{goodBidOF(20), goodBidOF(21), goodBidOF(22)},
			expErr: joinErrs(badAskErr(10), badAskErr(12)),
		},
		{
			name:   "two errors in bids",
			askOFs: []*OrderFulfillment{goodAskOF(10), goodAskOF(11), goodAskOF(12)},
			bidOFs: []*OrderFulfillment{badBidOF(20), goodBidOF(21), badBidOF(22)},
			expErr: joinErrs(badBidErr(20), badBidErr(22)),
		},
		{
			name:   "error in each",
			askOFs: []*OrderFulfillment{goodAskOF(10), goodAskOF(11), badAskOF(12)},
			bidOFs: []*OrderFulfillment{goodBidOF(20), badBidOF(21), goodBidOF(22)},
			expErr: joinErrs(badAskErr(12), badBidErr(21)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = validateFulfillments(tc.askOFs, tc.bidOFs)
			}
			require.NotPanics(t, testFunc, "validateFulfillments")
			assertions.AssertErrorValue(t, err, tc.expErr, "validateFulfillments error")
		})
	}
}

func TestBuildTransfers(t *testing.T) {
	tests := []struct {
		name          string
		askOFs        []*OrderFulfillment
		bidOFs        []*OrderFulfillment
		expSettlement *Settlement
		expErr        string
	}{
		{
			name: "ask with negative assets filled",
			askOFs: []*OrderFulfillment{
				{
					Order: NewOrder(18).WithAsk(&AskOrder{
						Assets: sdk.NewInt64Coin("apple", 15),
						Price:  sdk.NewInt64Coin("plum", 42),
					}),
					AssetsFilledAmt: sdkmath.NewInt(-1),
					PriceAppliedAmt: sdkmath.NewInt(-1),
				},
			},
			expErr: "ask order 18 cannot be filled with \"-1apple\" assets: amount not positive",
		},
		{
			name: "bid with negative assets filled",
			bidOFs: []*OrderFulfillment{
				{
					Order: NewOrder(12).WithBid(&BidOrder{
						Assets: sdk.NewInt64Coin("apple", 15),
						Price:  sdk.NewInt64Coin("plum", 42),
					}),
					AssetsFilledAmt: sdkmath.NewInt(-1),
					PriceAppliedAmt: sdkmath.NewInt(-1),
				},
			},
			expErr: "bid order 12 cannot be filled at price \"-1plum\": amount not positive",
		},
		{
			name: "ask with negative fees to pay",
			askOFs: []*OrderFulfillment{
				{
					Order: NewOrder(53).WithAsk(&AskOrder{
						Assets: sdk.NewInt64Coin("apple", 15),
						Price:  sdk.NewInt64Coin("plum", 42),
					}),
					AssetsFilledAmt: sdkmath.NewInt(15),
					AssetDists:      []*Distribution{{Address: "buyer1", Amount: sdkmath.NewInt(15)}},
					PriceAppliedAmt: sdkmath.NewInt(42),
					PriceDists:      []*Distribution{{Address: "seller1", Amount: sdkmath.NewInt(42)}},
					FeesToPay:       sdk.Coins{sdk.Coin{Denom: "fig", Amount: sdkmath.NewInt(-1)}},
				},
			},
			expErr: "ask order 53 cannot pay \"-1fig\" in fees: negative amount",
		},
		{
			name: "bid with negative fees to pay",
			bidOFs: []*OrderFulfillment{
				{
					Order: NewOrder(35).WithBid(&BidOrder{
						Assets: sdk.NewInt64Coin("apple", 15),
						Price:  sdk.NewInt64Coin("plum", 42),
					}),
					AssetsFilledAmt: sdkmath.NewInt(15),
					AssetDists:      []*Distribution{{Address: "seller1", Amount: sdkmath.NewInt(15)}},
					PriceAppliedAmt: sdkmath.NewInt(42),
					PriceDists:      []*Distribution{{Address: "seller1", Amount: sdkmath.NewInt(42)}},
					FeesToPay:       sdk.Coins{sdk.Coin{Denom: "fig", Amount: sdkmath.NewInt(-1)}},
				},
			},
			expErr: "bid order 35 cannot pay \"-1fig\" in fees: negative amount",
		},
		{
			name: "two asks, three bids",
			askOFs: []*OrderFulfillment{
				{
					Order: NewOrder(77).WithAsk(&AskOrder{
						Seller: "seller77",
						Assets: sdk.NewInt64Coin("apple", 15),
						Price:  sdk.NewInt64Coin("plum", 42),
					}),
					AssetsFilledAmt: sdkmath.NewInt(15),
					AssetDists: []*Distribution{
						{Address: "buyer5511", Amount: sdkmath.NewInt(15)},
					},
					PriceAppliedAmt: sdkmath.NewInt(43),
					PriceDists: []*Distribution{
						{Address: "buyer5511", Amount: sdkmath.NewInt(30)},
						{Address: "buyer78", Amount: sdkmath.NewInt(12)},
						{Address: "buyer9001", Amount: sdkmath.NewInt(1)},
					},
					FeesToPay: sdk.Coins{sdk.Coin{Denom: "fig", Amount: sdkmath.NewInt(11)}},
				},
				{
					Order: NewOrder(3).WithAsk(&AskOrder{
						Seller: "seller3",
						Assets: sdk.NewInt64Coin("apple", 43),
						Price:  sdk.NewInt64Coin("plum", 88),
					}),
					AssetsFilledAmt: sdkmath.NewInt(43),
					AssetDists: []*Distribution{
						{Address: "buyer5511", Amount: sdkmath.NewInt(5)},
						{Address: "buyer78", Amount: sdkmath.NewInt(7)},
						{Address: "buyer9001", Amount: sdkmath.NewInt(31)},
					},
					PriceAppliedAmt: sdkmath.NewInt(90),
					PriceDists: []*Distribution{
						{Address: "buyer78", Amount: sdkmath.NewInt(5)},
						{Address: "buyer9001", Amount: sdkmath.NewInt(83)},
						{Address: "buyer9001", Amount: sdkmath.NewInt(2)},
					},
					FeesToPay: nil,
				},
			},
			bidOFs: []*OrderFulfillment{
				{
					Order: NewOrder(5511).WithBid(&BidOrder{
						Buyer:  "buyer5511",
						Assets: sdk.NewInt64Coin("apple", 20),
						Price:  sdk.NewInt64Coin("plum", 30),
					}),
					AssetsFilledAmt: sdkmath.NewInt(20),
					AssetDists: []*Distribution{
						{Address: "seller77", Amount: sdkmath.NewInt(15)},
						{Address: "seller3", Amount: sdkmath.NewInt(5)},
					},
					PriceAppliedAmt: sdkmath.NewInt(30),
					PriceDists: []*Distribution{
						{Address: "seller77", Amount: sdkmath.NewInt(30)},
					},
					FeesToPay: sdk.Coins{sdk.Coin{Denom: "fig", Amount: sdkmath.NewInt(10)}},
				},
				{
					Order: NewOrder(78).WithBid(&BidOrder{
						Buyer:  "buyer78",
						Assets: sdk.NewInt64Coin("apple", 7),
						Price:  sdk.NewInt64Coin("plum", 15),
					}),
					AssetsFilledAmt: sdkmath.NewInt(7),
					AssetDists: []*Distribution{
						{Address: "seller3", Amount: sdkmath.NewInt(7)},
					},
					PriceAppliedAmt: sdkmath.NewInt(15),
					PriceDists: []*Distribution{
						{Address: "seller77", Amount: sdkmath.NewInt(12)},
						{Address: "seller3", Amount: sdkmath.NewInt(3)},
					},
					FeesToPay: sdk.Coins{sdk.Coin{Denom: "grape", Amount: sdkmath.NewInt(4)}},
				},
				{
					Order: NewOrder(9001).WithBid(&BidOrder{
						Buyer:  "buyer9001",
						Assets: sdk.NewInt64Coin("apple", 31),
						Price:  sdk.NewInt64Coin("plum", 86),
					}),
					AssetsFilledAmt: sdkmath.NewInt(31),
					AssetDists: []*Distribution{
						{Address: "seller3", Amount: sdkmath.NewInt(31)},
					},
					PriceAppliedAmt: sdkmath.NewInt(86),
					PriceDists: []*Distribution{
						{Address: "seller3", Amount: sdkmath.NewInt(83)},
						{Address: "seller77", Amount: sdkmath.NewInt(2)},
						{Address: "seller3", Amount: sdkmath.NewInt(1)},
					},
					FeesToPay: nil,
				},
			},
			expSettlement: &Settlement{
				Transfers: []*Transfer{
					{
						Inputs:  []banktypes.Input{{Address: "seller77", Coins: sdk.NewCoins(sdk.NewInt64Coin("apple", 15))}},
						Outputs: []banktypes.Output{{Address: "buyer5511", Coins: sdk.NewCoins(sdk.NewInt64Coin("apple", 15))}},
					},
					{
						Inputs: []banktypes.Input{{Address: "seller3", Coins: sdk.NewCoins(sdk.NewInt64Coin("apple", 43))}},
						Outputs: []banktypes.Output{
							{Address: "buyer5511", Coins: sdk.NewCoins(sdk.NewInt64Coin("apple", 5))},
							{Address: "buyer78", Coins: sdk.NewCoins(sdk.NewInt64Coin("apple", 7))},
							{Address: "buyer9001", Coins: sdk.NewCoins(sdk.NewInt64Coin("apple", 31))},
						},
					},
					{
						Inputs:  []banktypes.Input{{Address: "buyer5511", Coins: sdk.NewCoins(sdk.NewInt64Coin("plum", 30))}},
						Outputs: []banktypes.Output{{Address: "seller77", Coins: sdk.NewCoins(sdk.NewInt64Coin("plum", 30))}},
					},
					{
						Inputs: []banktypes.Input{{Address: "buyer78", Coins: sdk.NewCoins(sdk.NewInt64Coin("plum", 15))}},
						Outputs: []banktypes.Output{
							{Address: "seller77", Coins: sdk.NewCoins(sdk.NewInt64Coin("plum", 12))},
							{Address: "seller3", Coins: sdk.NewCoins(sdk.NewInt64Coin("plum", 3))},
						},
					},
					{
						Inputs: []banktypes.Input{{Address: "buyer9001", Coins: sdk.NewCoins(sdk.NewInt64Coin("plum", 86))}},
						Outputs: []banktypes.Output{
							{Address: "seller3", Coins: sdk.NewCoins(sdk.NewInt64Coin("plum", 84))},
							{Address: "seller77", Coins: sdk.NewCoins(sdk.NewInt64Coin("plum", 2))},
						},
					},
				},
				FeeInputs: []banktypes.Input{
					{Address: "seller77", Coins: sdk.NewCoins(sdk.NewInt64Coin("fig", 11))},
					{Address: "buyer5511", Coins: sdk.NewCoins(sdk.NewInt64Coin("fig", 10))},
					{Address: "buyer78", Coins: sdk.NewCoins(sdk.NewInt64Coin("grape", 4))},
				},
			},
			expErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			settlement := &Settlement{}
			if tc.expSettlement == nil {
				tc.expSettlement = &Settlement{}
			}
			var err error
			testFunc := func() {
				err = buildTransfers(tc.askOFs, tc.bidOFs, settlement)
			}
			require.NotPanics(t, testFunc, "buildTransfers")
			assertions.AssertErrorValue(t, err, tc.expErr, "buildTransfers error")
			if !assert.Equal(t, tc.expSettlement, settlement, "settlement after buildTransfers") {
				expTransStrs := make([]string, len(tc.expSettlement.Transfers))
				for i, t := range tc.expSettlement.Transfers {
					expTransStrs[i] = fmt.Sprintf("[%d]%s", i, transferString(t))
				}
				expTrans := strings.Join(expTransStrs, "\n")
				actTransStrs := make([]string, len(tc.expSettlement.Transfers))
				for i, t := range settlement.Transfers {
					actTransStrs[i] = fmt.Sprintf("[%d]%s", i, transferString(t))
				}
				actTrans := strings.Join(actTransStrs, "\n")
				assert.Equal(t, expTrans, actTrans, "transfers (as strings)")
			}
		})
	}
}

func TestPopulateFilled(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	askOrder := func(orderID uint64, assetsAmt, priceAmt int64) *Order {
		return NewOrder(orderID).WithAsk(&AskOrder{
			Assets: coin(assetsAmt, "acorn"),
			Price:  coin(priceAmt, "prune"),
		})
	}
	bidOrder := func(orderID uint64, assetsAmt, priceAmt int64) *Order {
		return NewOrder(orderID).WithBid(&BidOrder{
			Assets: coin(assetsAmt, "acorn"),
			Price:  coin(priceAmt, "prune"),
		})
	}
	newOF := func(order *Order, priceAppliedAmt int64, fees ...sdk.Coin) *OrderFulfillment {
		rv := &OrderFulfillment{
			Order:             order,
			AssetsFilledAmt:   order.GetAssets().Amount,
			AssetsUnfilledAmt: sdkmath.ZeroInt(),
			PriceAppliedAmt:   sdkmath.NewInt(priceAppliedAmt),
			PriceLeftAmt:      order.GetPrice().Amount.SubRaw(priceAppliedAmt),
		}
		if len(fees) > 0 {
			rv.FeesToPay = fees
		}
		return rv
	}
	filledOrder := func(order *Order, actualPrice int64, actualFees ...sdk.Coin) *FilledOrder {
		rv := &FilledOrder{
			order:       order,
			actualPrice: coin(actualPrice, order.GetPrice().Denom),
		}
		if len(actualFees) > 0 {
			rv.actualFees = actualFees
		}
		return rv
	}

	tests := []struct {
		name          string
		askOFs        []*OrderFulfillment
		bidOFs        []*OrderFulfillment
		settlement    *Settlement
		expSettlement *Settlement
	}{
		{
			name: "no partial",
			askOFs: []*OrderFulfillment{
				newOF(askOrder(2001, 53, 87), 92, coin(12, "fig")),
				newOF(askOrder(2002, 17, 33), 37),
				newOF(askOrder(2003, 22, 56), 60),
			},
			bidOFs: []*OrderFulfillment{
				newOF(bidOrder(3001, 30, 40), 40),
				newOF(bidOrder(3002, 27, 49), 49, coin(39, "fig")),
				newOF(bidOrder(3003, 35, 100), 100),
			},
			settlement: &Settlement{},
			expSettlement: &Settlement{
				FullyFilledOrders: []*FilledOrder{
					filledOrder(askOrder(2001, 53, 87), 92, coin(12, "fig")),
					filledOrder(askOrder(2002, 17, 33), 37),
					filledOrder(askOrder(2003, 22, 56), 60),
					filledOrder(bidOrder(3001, 30, 40), 40),
					filledOrder(bidOrder(3002, 27, 49), 49, coin(39, "fig")),
					filledOrder(bidOrder(3003, 35, 100), 100),
				},
			},
		},
		{
			name: "partial ask",
			askOFs: []*OrderFulfillment{
				newOF(askOrder(2001, 53, 87), 92, coin(12, "fig")),
				newOF(askOrder(2002, 17, 33), 37),
				newOF(askOrder(2003, 22, 56), 60),
			},
			bidOFs: []*OrderFulfillment{
				newOF(bidOrder(3001, 30, 40), 40),
				newOF(bidOrder(3002, 27, 49), 49, coin(39, "fig")),
				newOF(bidOrder(3003, 35, 100), 100),
			},
			settlement: &Settlement{PartialOrderLeft: askOrder(2002, 15, 63)},
			expSettlement: &Settlement{
				FullyFilledOrders: []*FilledOrder{
					filledOrder(askOrder(2001, 53, 87), 92, coin(12, "fig")),
					filledOrder(askOrder(2003, 22, 56), 60),
					filledOrder(bidOrder(3001, 30, 40), 40),
					filledOrder(bidOrder(3002, 27, 49), 49, coin(39, "fig")),
					filledOrder(bidOrder(3003, 35, 100), 100),
				},
				PartialOrderFilled: filledOrder(askOrder(2002, 17, 33), 37),
				PartialOrderLeft:   askOrder(2002, 15, 63),
			},
		},
		{
			name: "partial bid",
			askOFs: []*OrderFulfillment{
				newOF(askOrder(2001, 53, 87), 92, coin(12, "fig")),
				newOF(askOrder(2002, 17, 33), 37),
				newOF(askOrder(2003, 22, 56), 60),
			},
			bidOFs: []*OrderFulfillment{
				newOF(bidOrder(3001, 30, 40), 40),
				newOF(bidOrder(3002, 27, 49), 49, coin(39, "fig")),
				newOF(bidOrder(3003, 35, 100), 100),
			},
			settlement: &Settlement{PartialOrderLeft: bidOrder(3003, 15, 63)},
			expSettlement: &Settlement{
				FullyFilledOrders: []*FilledOrder{
					filledOrder(askOrder(2001, 53, 87), 92, coin(12, "fig")),
					filledOrder(askOrder(2002, 17, 33), 37),
					filledOrder(askOrder(2003, 22, 56), 60),
					filledOrder(bidOrder(3001, 30, 40), 40),
					filledOrder(bidOrder(3002, 27, 49), 49, coin(39, "fig")),
				},
				PartialOrderFilled: filledOrder(bidOrder(3003, 35, 100), 100),
				PartialOrderLeft:   bidOrder(3003, 15, 63),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testFunc := func() {
				populateFilled(tc.askOFs, tc.bidOFs, tc.settlement)
			}
			require.NotPanics(t, testFunc, "populateFilled")
			assert.Equal(t, tc.expSettlement, tc.settlement, "settlement after populateFilled")
		})
	}
}

func TestGetAssetTransfer(t *testing.T) {
	seller, buyer := "sally", "brandon"
	assetDenom, priceDenom := "apple", "peach"
	newOF := func(order *Order, assetsFilledAmt int64, assetDists ...*Distribution) *OrderFulfillment {
		rv := &OrderFulfillment{
			Order:           order,
			AssetsFilledAmt: sdkmath.NewInt(assetsFilledAmt),
		}
		if len(assetDists) > 0 {
			rv.AssetDists = assetDists
		}
		return rv
	}
	askOrder := func(orderID uint64) *Order {
		return NewOrder(orderID).WithAsk(&AskOrder{
			MarketId: 5555,
			Seller:   seller,
			Assets:   sdk.Coin{Denom: assetDenom, Amount: sdkmath.NewInt(111)},
			Price:    sdk.Coin{Denom: priceDenom, Amount: sdkmath.NewInt(999)},
		})
	}
	bidOrder := func(orderID uint64) *Order {
		return NewOrder(orderID).WithBid(&BidOrder{
			MarketId: 5555,
			Buyer:    buyer,
			Assets:   sdk.Coin{Denom: assetDenom, Amount: sdkmath.NewInt(111)},
			Price:    sdk.Coin{Denom: priceDenom, Amount: sdkmath.NewInt(999)},
		})
	}
	dist := func(addr string, amount int64) *Distribution {
		return &Distribution{Address: addr, Amount: sdkmath.NewInt(amount)}
	}
	input := func(addr string, amount int64) banktypes.Input {
		return banktypes.Input{
			Address: addr,
			Coins:   sdk.Coins{{Denom: assetDenom, Amount: sdkmath.NewInt(amount)}},
		}
	}
	output := func(addr string, amount int64) banktypes.Output {
		return banktypes.Output{
			Address: addr,
			Coins:   sdk.Coins{{Denom: assetDenom, Amount: sdkmath.NewInt(amount)}},
		}
	}

	tests := []struct {
		name        string
		f           *OrderFulfillment
		expTransfer *Transfer
		expErr      string
		expPanic    string
	}{
		{
			name:     "nil inside order",
			f:        newOF(NewOrder(975), 5, dist("five", 5)),
			expPanic: nilSubTypeErr(975),
		},
		{
			name:     "unknown inside order",
			f:        newOF(newUnknownOrder(974), 5, dist("five", 5)),
			expPanic: unknownSubTypeErr(974),
		},
		{
			name:   "assets filled negative: ask",
			f:      newOF(askOrder(159), -5),
			expErr: "ask order 159 cannot be filled with \"-5apple\" assets: amount not positive",
		},
		{
			name:   "assets filled negative: bid",
			f:      newOF(bidOrder(953), -5),
			expErr: "bid order 953 cannot be filled with \"-5apple\" assets: amount not positive",
		},
		{
			name:   "assets filled zero: ask",
			f:      newOF(askOrder(991), 0),
			expErr: "ask order 991 cannot be filled with \"0apple\" assets: amount not positive",
		},
		{
			name:   "assets filled zero: bid",
			f:      newOF(bidOrder(992), 0),
			expErr: "bid order 992 cannot be filled with \"0apple\" assets: amount not positive",
		},
		{
			name:   "asset dists has negative amount: ask",
			f:      newOF(askOrder(549), 10, dist("one", 1), dist("two", 2), dist("neg", -2), dist("nine", 9)),
			expErr: "ask order 549 cannot have \"-2apple\" assets in a transfer: amount not positive",
		},
		{
			name:   "asset dists has negative amount: bid",
			f:      newOF(bidOrder(545), 10, dist("one", 1), dist("two", 2), dist("neg", -2), dist("nine", 9)),
			expErr: "bid order 545 cannot have \"-2apple\" assets in a transfer: amount not positive",
		},
		{
			name:   "asset dists has zero: ask",
			f:      newOF(askOrder(683), 10, dist("one", 1), dist("two", 2), dist("zero", 0), dist("seven", 7)),
			expErr: "ask order 683 cannot have \"0apple\" assets in a transfer: amount not positive",
		},
		{
			name:   "asset dists has zero: bid",
			f:      newOF(bidOrder(777), 10, dist("one", 1), dist("two", 2), dist("zero", 0), dist("seven", 7)),
			expErr: "bid order 777 cannot have \"0apple\" assets in a transfer: amount not positive",
		},
		{
			name:   "asset dists sum less than assets filled: ask",
			f:      newOF(askOrder(8), 10, dist("one", 1), dist("two", 2), dist("three", 3), dist("three2", 3)),
			expErr: "ask order 8 assets filled \"10apple\" does not equal assets distributed \"9apple\"",
		},
		{
			name:   "asset dists sum less than assets filled: bid",
			f:      newOF(bidOrder(3), 10, dist("one", 1), dist("two", 2), dist("three", 3), dist("three2", 3)),
			expErr: "bid order 3 assets filled \"10apple\" does not equal assets distributed \"9apple\"",
		},
		{
			name:   "asset dists sum more than assets filled: ask",
			f:      newOF(askOrder(8), 10, dist("one", 1), dist("two", 2), dist("three", 3), dist("five", 5)),
			expErr: "ask order 8 assets filled \"10apple\" does not equal assets distributed \"11apple\"",
		},
		{
			name:   "asset dists sum more than assets filled: bid",
			f:      newOF(bidOrder(3), 10, dist("one", 1), dist("two", 2), dist("three", 3), dist("five", 5)),
			expErr: "bid order 3 assets filled \"10apple\" does not equal assets distributed \"11apple\"",
		},
		{
			name: "one dist: ask",
			f:    newOF(askOrder(12), 10, dist("ten", 10)),
			expTransfer: &Transfer{
				Inputs:  []banktypes.Input{input(seller, 10)},
				Outputs: []banktypes.Output{output("ten", 10)},
			},
		},
		{
			name: "one dist: bid",
			f:    newOF(bidOrder(13), 10, dist("ten", 10)),
			expTransfer: &Transfer{
				Inputs:  []banktypes.Input{input("ten", 10)},
				Outputs: []banktypes.Output{output(buyer, 10)},
			},
		},
		{
			name: "two dists, different addresses: ask",
			f:    newOF(askOrder(2111), 20, dist("eleven", 11), dist("nine", 9)),
			expTransfer: &Transfer{
				Inputs:  []banktypes.Input{input(seller, 20)},
				Outputs: []banktypes.Output{output("eleven", 11), output("nine", 9)},
			},
		},
		{
			name: "two dists, different addresses: bid",
			f:    newOF(bidOrder(1222), 20, dist("eleven", 11), dist("nine", 9)),
			expTransfer: &Transfer{
				Inputs:  []banktypes.Input{input("eleven", 11), input("nine", 9)},
				Outputs: []banktypes.Output{output(buyer, 20)},
			},
		},
		{
			name: "two dists, same addresses: ask",
			f:    newOF(askOrder(5353), 52, dist("billy", 48), dist("billy", 4)),
			expTransfer: &Transfer{
				Inputs:  []banktypes.Input{input(seller, 52)},
				Outputs: []banktypes.Output{output("billy", 52)},
			},
		},
		{
			name: "two dists, same addresses: bid",
			f:    newOF(bidOrder(3535), 52, dist("sol", 48), dist("sol", 4)),
			expTransfer: &Transfer{
				Inputs:  []banktypes.Input{input("sol", 52)},
				Outputs: []banktypes.Output{output(buyer, 52)},
			},
		},
		{
			name: "four dists: ask",
			f: newOF(askOrder(99221), 33,
				dist("buddy", 10), dist("brian", 13), dist("buddy", 8), dist("bella", 2)),
			expTransfer: &Transfer{
				Inputs:  []banktypes.Input{input(seller, 33)},
				Outputs: []banktypes.Output{output("buddy", 18), output("brian", 13), output("bella", 2)},
			},
		},
		{
			name: "four dists: bid",
			f: newOF(bidOrder(99221), 33,
				dist("sydney", 10), dist("sarah", 2), dist("sydney", 8), dist("spencer", 13)),
			expTransfer: &Transfer{
				Inputs:  []banktypes.Input{input("sydney", 18), input("sarah", 2), input("spencer", 13)},
				Outputs: []banktypes.Output{output(buyer, 33)},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := copyOrderFulfillment(tc.f)
			var transfer *Transfer
			var err error
			testFunc := func() {
				transfer, err = getAssetTransfer(tc.f)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "getAssetTransfer")
			assertions.AssertErrorValue(t, err, tc.expErr, "getAssetTransfer error")
			if !assert.Equal(t, tc.expTransfer, transfer, "getAssetTransfer transfers") {
				t.Logf("  Actual: %s", transferString(transfer))
				t.Logf("Expected: %s", transferString(tc.expTransfer))
			}
			assertEqualOrderFulfillments(t, orig, tc.f, "OrderFulfillment before and after getAssetTransfer")
		})
	}
}

func TestGetPriceTransfer(t *testing.T) {
	seller, buyer := "sally", "brandon"
	assetDenom, priceDenom := "apple", "peach"
	newOF := func(order *Order, priceAppliedAmt int64, priceDists ...*Distribution) *OrderFulfillment {
		rv := &OrderFulfillment{
			Order:           order,
			PriceAppliedAmt: sdkmath.NewInt(priceAppliedAmt),
		}
		if len(priceDists) > 0 {
			rv.PriceDists = priceDists
		}
		return rv
	}
	askOrder := func(orderID uint64) *Order {
		return NewOrder(orderID).WithAsk(&AskOrder{
			MarketId: 5555,
			Seller:   seller,
			Assets:   sdk.Coin{Denom: assetDenom, Amount: sdkmath.NewInt(111)},
			Price:    sdk.Coin{Denom: priceDenom, Amount: sdkmath.NewInt(999)},
		})
	}
	bidOrder := func(orderID uint64) *Order {
		return NewOrder(orderID).WithBid(&BidOrder{
			MarketId: 5555,
			Buyer:    buyer,
			Assets:   sdk.Coin{Denom: assetDenom, Amount: sdkmath.NewInt(111)},
			Price:    sdk.Coin{Denom: priceDenom, Amount: sdkmath.NewInt(999)},
		})
	}
	dist := func(addr string, amount int64) *Distribution {
		return &Distribution{Address: addr, Amount: sdkmath.NewInt(amount)}
	}
	input := func(addr string, amount int64) banktypes.Input {
		return banktypes.Input{
			Address: addr,
			Coins:   sdk.Coins{{Denom: priceDenom, Amount: sdkmath.NewInt(amount)}},
		}
	}
	output := func(addr string, amount int64) banktypes.Output {
		return banktypes.Output{
			Address: addr,
			Coins:   sdk.Coins{{Denom: priceDenom, Amount: sdkmath.NewInt(amount)}},
		}
	}

	tests := []struct {
		name        string
		f           *OrderFulfillment
		expTransfer *Transfer
		expErr      string
		expPanic    string
	}{
		{
			name:     "nil inside order",
			f:        newOF(NewOrder(975), 5, dist("five", 5)),
			expPanic: nilSubTypeErr(975),
		},
		{
			name:     "unknown inside order",
			f:        newOF(newUnknownOrder(974), 5, dist("five", 5)),
			expPanic: unknownSubTypeErr(974),
		},
		{
			name:   "price applied negative: ask",
			f:      newOF(askOrder(159), -5),
			expErr: "ask order 159 cannot be filled at price \"-5peach\": amount not positive",
		},
		{
			name:   "price applied negative: bid",
			f:      newOF(bidOrder(953), -5),
			expErr: "bid order 953 cannot be filled at price \"-5peach\": amount not positive",
		},
		{
			name:   "price applied zero: ask",
			f:      newOF(askOrder(991), 0),
			expErr: "ask order 991 cannot be filled at price \"0peach\": amount not positive",
		},
		{
			name:   "price applied zero: bid",
			f:      newOF(bidOrder(992), 0),
			expErr: "bid order 992 cannot be filled at price \"0peach\": amount not positive",
		},
		{
			name:   "price dists has negative amount: ask",
			f:      newOF(askOrder(549), 10, dist("one", 1), dist("two", 2), dist("neg", -2), dist("nine", 9)),
			expErr: "ask order 549 cannot have price \"-2peach\" in a transfer: amount not positive",
		},
		{
			name:   "price dists has negative amount: bid",
			f:      newOF(bidOrder(545), 10, dist("one", 1), dist("two", 2), dist("neg", -2), dist("nine", 9)),
			expErr: "bid order 545 cannot have price \"-2peach\" in a transfer: amount not positive",
		},
		{
			name:   "price dists has zero: ask",
			f:      newOF(askOrder(683), 10, dist("one", 1), dist("two", 2), dist("zero", 0), dist("seven", 7)),
			expErr: "ask order 683 cannot have price \"0peach\" in a transfer: amount not positive",
		},
		{
			name:   "price dists has zero: bid",
			f:      newOF(bidOrder(777), 10, dist("one", 1), dist("two", 2), dist("zero", 0), dist("seven", 7)),
			expErr: "bid order 777 cannot have price \"0peach\" in a transfer: amount not positive",
		},
		{
			name:   "price dists sum less than price applied: ask",
			f:      newOF(askOrder(8), 10, dist("one", 1), dist("two", 2), dist("three", 3), dist("three2", 3)),
			expErr: "ask order 8 price filled \"10peach\" does not equal price distributed \"9peach\"",
		},
		{
			name:   "price dists sum less than price applied: bid",
			f:      newOF(bidOrder(3), 10, dist("one", 1), dist("two", 2), dist("three", 3), dist("three2", 3)),
			expErr: "bid order 3 price filled \"10peach\" does not equal price distributed \"9peach\"",
		},
		{
			name:   "price dists sum more than price applied: ask",
			f:      newOF(askOrder(8), 10, dist("one", 1), dist("two", 2), dist("three", 3), dist("five", 5)),
			expErr: "ask order 8 price filled \"10peach\" does not equal price distributed \"11peach\"",
		},
		{
			name:   "price dists sum more than price applied: bid",
			f:      newOF(bidOrder(3), 10, dist("one", 1), dist("two", 2), dist("three", 3), dist("five", 5)),
			expErr: "bid order 3 price filled \"10peach\" does not equal price distributed \"11peach\"",
		},
		{
			name: "one dist: ask",
			f:    newOF(askOrder(12), 10, dist("ten", 10)),
			expTransfer: &Transfer{
				Inputs:  []banktypes.Input{input("ten", 10)},
				Outputs: []banktypes.Output{output(seller, 10)},
			},
		},
		{
			name: "one dist: bid",
			f:    newOF(bidOrder(13), 10, dist("ten", 10)),
			expTransfer: &Transfer{
				Inputs:  []banktypes.Input{input(buyer, 10)},
				Outputs: []banktypes.Output{output("ten", 10)},
			},
		},
		{
			name: "two dists, different addresses: ask",
			f:    newOF(askOrder(2111), 20, dist("eleven", 11), dist("nine", 9)),
			expTransfer: &Transfer{
				Inputs:  []banktypes.Input{input("eleven", 11), input("nine", 9)},
				Outputs: []banktypes.Output{output(seller, 20)},
			},
		},
		{
			name: "two dists, different addresses: bid",
			f:    newOF(bidOrder(1222), 20, dist("eleven", 11), dist("nine", 9)),
			expTransfer: &Transfer{
				Inputs:  []banktypes.Input{input(buyer, 20)},
				Outputs: []banktypes.Output{output("eleven", 11), output("nine", 9)},
			},
		},
		{
			name: "two dists, same addresses: ask",
			f:    newOF(askOrder(5353), 52, dist("billy", 48), dist("billy", 4)),
			expTransfer: &Transfer{
				Inputs:  []banktypes.Input{input("billy", 52)},
				Outputs: []banktypes.Output{output(seller, 52)},
			},
		},
		{
			name: "two dists, same addresses: bid",
			f:    newOF(bidOrder(3535), 52, dist("sol", 48), dist("sol", 4)),
			expTransfer: &Transfer{
				Inputs:  []banktypes.Input{input(buyer, 52)},
				Outputs: []banktypes.Output{output("sol", 52)},
			},
		},
		{
			name: "four dists: ask",
			f: newOF(askOrder(99221), 33,
				dist("buddy", 10), dist("brian", 13), dist("buddy", 8), dist("bella", 2)),
			expTransfer: &Transfer{
				Inputs:  []banktypes.Input{input("buddy", 18), input("brian", 13), input("bella", 2)},
				Outputs: []banktypes.Output{output(seller, 33)},
			},
		},
		{
			name: "four dists: bid",
			f: newOF(bidOrder(99221), 33,
				dist("sydney", 10), dist("sarah", 2), dist("sydney", 8), dist("spencer", 13)),
			expTransfer: &Transfer{
				Inputs:  []banktypes.Input{input(buyer, 33)},
				Outputs: []banktypes.Output{output("sydney", 18), output("sarah", 2), output("spencer", 13)},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := copyOrderFulfillment(tc.f)
			var transfer *Transfer
			var err error
			testFunc := func() {
				transfer, err = getPriceTransfer(tc.f)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "getPriceTransfer")
			assertions.AssertErrorValue(t, err, tc.expErr, "getPriceTransfer error")
			if !assert.Equal(t, tc.expTransfer, transfer, "getPriceTransfer transfers") {
				t.Logf("  Actual: %s", transferString(transfer))
				t.Logf("Expected: %s", transferString(tc.expTransfer))
			}
			assertEqualOrderFulfillments(t, orig, tc.f, "OrderFulfillment before and after getPriceTransfer")
		})
	}
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

func TestOrderFulfillment_Validate(t *testing.T) {
	askOrder := func(orderID uint64, assetsAmt, priceAmt int64) *Order {
		return NewOrder(orderID).WithAsk(&AskOrder{
			MarketId: 987,
			Seller:   "steve",
			Assets:   sdk.Coin{Denom: "apple", Amount: sdkmath.NewInt(assetsAmt)},
			Price:    sdk.Coin{Denom: "peach", Amount: sdkmath.NewInt(priceAmt)},
		})
	}
	bidOrder := func(orderID uint64, assetsAmt, priceAmt int64) *Order {
		return NewOrder(orderID).WithBid(&BidOrder{
			MarketId: 987,
			Buyer:    "bruce",
			Assets:   sdk.Coin{Denom: "apple", Amount: sdkmath.NewInt(assetsAmt)},
			Price:    sdk.Coin{Denom: "peach", Amount: sdkmath.NewInt(priceAmt)},
		})
	}

	tests := []struct {
		name   string
		f      OrderFulfillment
		expErr string
	}{
		{
			name:   "nil inside order",
			f:      OrderFulfillment{Order: NewOrder(8)},
			expErr: nilSubTypeErr(8),
		},
		{
			name:   "unknown inside order",
			f:      OrderFulfillment{Order: newUnknownOrder(12)},
			expErr: unknownSubTypeErr(12),
		},
		{
			name: "order price greater than price applied: ask",
			f: OrderFulfillment{
				Order:           askOrder(52, 10, 401),
				PriceAppliedAmt: sdkmath.NewInt(400),
				AssetsFilledAmt: sdkmath.NewInt(10),
			},
			expErr: "ask order 52 price \"401peach\" is more than price filled \"400peach\"",
		},
		{
			name: "order price equal to price applied: ask",
			f: OrderFulfillment{
				Order:           askOrder(53, 10, 401),
				PriceAppliedAmt: sdkmath.NewInt(401),
				AssetsFilledAmt: sdkmath.NewInt(10),
			},
			expErr: "",
		},
		{
			name: "order price less than price applied: ask",
			f: OrderFulfillment{
				Order:           askOrder(54, 10, 401),
				PriceAppliedAmt: sdkmath.NewInt(402),
				AssetsFilledAmt: sdkmath.NewInt(10),
			},
			expErr: "",
		},
		{
			name: "order price greater than price applied: bid",
			f: OrderFulfillment{
				Order:           bidOrder(71, 17, 432),
				PriceAppliedAmt: sdkmath.NewInt(431),
				AssetsFilledAmt: sdkmath.NewInt(17),
			},
			expErr: "bid order 71 price \"432peach\" is not equal to price filled \"431peach\"",
		},
		{
			name: "order price equal to price applied: bid",
			f: OrderFulfillment{
				Order:           bidOrder(72, 17, 432),
				PriceAppliedAmt: sdkmath.NewInt(432),
				AssetsFilledAmt: sdkmath.NewInt(17),
			},
			expErr: "",
		},
		{
			name: "order price less than price applied: bid",
			f: OrderFulfillment{
				Order:           bidOrder(73, 17, 432),
				PriceAppliedAmt: sdkmath.NewInt(433),
				AssetsFilledAmt: sdkmath.NewInt(17),
			},
			expErr: "bid order 73 price \"432peach\" is not equal to price filled \"433peach\"",
		},
		{
			name: "order assets less than assets filled: ask",
			f: OrderFulfillment{
				Order:           askOrder(101, 53, 12345),
				PriceAppliedAmt: sdkmath.NewInt(12345),
				AssetsFilledAmt: sdkmath.NewInt(54),
			},
			expErr: "ask order 101 assets \"53apple\" does not equal filled assets \"54apple\"",
		},
		{
			name: "order assets equal to assets filled: ask",
			f: OrderFulfillment{
				Order:           askOrder(202, 53, 12345),
				PriceAppliedAmt: sdkmath.NewInt(12345),
				AssetsFilledAmt: sdkmath.NewInt(53),
			},
			expErr: "",
		},
		{
			name: "order assets more than assets filled: ask",
			f: OrderFulfillment{
				Order:           askOrder(303, 53, 12345),
				PriceAppliedAmt: sdkmath.NewInt(12345),
				AssetsFilledAmt: sdkmath.NewInt(52),
			},
			expErr: "ask order 303 assets \"53apple\" does not equal filled assets \"52apple\"",
		},
		{
			name: "order assets less than assets filled: bid",
			f: OrderFulfillment{
				Order:           bidOrder(404, 53, 12345),
				PriceAppliedAmt: sdkmath.NewInt(12345),
				AssetsFilledAmt: sdkmath.NewInt(54),
			},
			expErr: "bid order 404 assets \"53apple\" does not equal filled assets \"54apple\"",
		},
		{
			name: "order assets equal to assets filled: bid",
			f: OrderFulfillment{
				Order:           bidOrder(505, 53, 12345),
				PriceAppliedAmt: sdkmath.NewInt(12345),
				AssetsFilledAmt: sdkmath.NewInt(53),
			},
			expErr: "",
		},
		{
			name: "order assets more than assets filled: bid",
			f: OrderFulfillment{
				Order:           bidOrder(606, 53, 12345),
				PriceAppliedAmt: sdkmath.NewInt(12345),
				AssetsFilledAmt: sdkmath.NewInt(52),
			},
			expErr: "bid order 606 assets \"53apple\" does not equal filled assets \"52apple\"",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.f.Validate()
			}
			require.NotPanics(t, testFunc, "Validate")
			assertions.AssertErrorValue(t, err, tc.expErr, "Validate error")
		})
	}
}

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
			newTests[0].expErr = fmt.Sprintf("cannot fill ask order 1 having assets left \"%done\" with bid "+
				"order 2 having assets left \"%dtwo\": zero or negative assets left",
				c.of1Unfilled, c.of2Unfilled)
			newTests[1].expErr = fmt.Sprintf("cannot fill bid order 1 having assets left \"%done\" with ask "+
				"order 2 having assets left \"%dtwo\": zero or negative assets left",
				c.of1Unfilled, c.of2Unfilled)
			newTests[2].expErr = fmt.Sprintf("cannot fill ask order 1 having assets left \"%done\" with ask "+
				"order 2 having assets left \"%dtwo\": zero or negative assets left",
				c.of1Unfilled, c.of2Unfilled)
			newTests[3].expErr = fmt.Sprintf("cannot fill bid order 1 having assets left \"%done\" with bid "+
				"order 2 having assets left \"%dtwo\": zero or negative assets left",
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

func TestGetFulfillmentPriceAmt(t *testing.T) {
	newAskOF := func(orderID uint64, assetsUnfilled int64, assetDenom string) *OrderFulfillment {
		return &OrderFulfillment{
			Order: NewOrder(orderID).WithAsk(&AskOrder{
				Price: sdk.NewInt64Coin(assetDenom, 999),
			}),
			PriceLeftAmt: sdkmath.NewInt(assetsUnfilled),
		}
	}
	newBidOF := func(orderID uint64, assetsUnfilled int64, assetDenom string) *OrderFulfillment {
		return &OrderFulfillment{
			Order: NewOrder(orderID).WithBid(&BidOrder{
				Price: sdk.NewInt64Coin(assetDenom, 999),
			}),
			PriceLeftAmt: sdkmath.NewInt(assetsUnfilled),
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
			newTests[0].expErr = fmt.Sprintf("cannot fill ask order 1 having price left \"%done\" with bid "+
				"order 2 having price left \"%dtwo\": zero or negative price left",
				c.of1Unfilled, c.of2Unfilled)
			newTests[1].expErr = fmt.Sprintf("cannot fill bid order 1 having price left \"%done\" with ask "+
				"order 2 having price left \"%dtwo\": zero or negative price left",
				c.of1Unfilled, c.of2Unfilled)
			newTests[2].expErr = fmt.Sprintf("cannot fill ask order 1 having price left \"%done\" with ask "+
				"order 2 having price left \"%dtwo\": zero or negative price left",
				c.of1Unfilled, c.of2Unfilled)
			newTests[3].expErr = fmt.Sprintf("cannot fill bid order 1 having price left \"%done\" with bid "+
				"order 2 having price left \"%dtwo\": zero or negative price left",
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
				amt, err = GetFulfillmentPriceAmt(tc.of1, tc.of2)
			}
			require.NotPanics(t, testFunc, "GetFulfillmentPriceAmt")
			assertions.AssertErrorValue(t, err, tc.expErr, "GetFulfillmentPriceAmt error")
			assert.Equal(t, tc.expAmt, amt, "GetFulfillmentPriceAmt amount")
			assertEqualOrderFulfillments(t, origOF1, tc.of1, "of1 after GetFulfillmentPriceAmt")
			assertEqualOrderFulfillments(t, origOF2, tc.of2, "of2 after GetFulfillmentPriceAmt")
		})
	}
}

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

func TestGetAssetTransfer2(t *testing.T) {
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
				actual = GetAssetTransfer2(tc.f)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "GetAssetTransfer2")
			assert.Equal(t, tc.exp, actual, "GetAssetTransfer2 result")
		})
	}
}

func TestGetPriceTransfer2(t *testing.T) {
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
				actual = GetPriceTransfer2(tc.f)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "GetPriceTransfer2")
			assert.Equal(t, tc.exp, actual, "GetPriceTransfer2 result")
		})
	}
}
