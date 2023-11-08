package exchange

import (
	"errors"
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

// copyDistribution copies a distribution.
func copyDistribution(dist *distribution) *distribution {
	if dist == nil {
		return nil
	}

	return &distribution{
		Address: dist.Address,
		Amount:  copySDKInt(dist.Amount),
	}
}

// copyDistributions copies a slice of distributions.
func copyDistributions(dists []*distribution) []*distribution {
	return copySlice(dists, copyDistribution)
}

// copyOrderFulfillment returns a deep copy of an order fulfillment.
func copyOrderFulfillment(f *orderFulfillment) *orderFulfillment {
	if f == nil {
		return nil
	}

	return &orderFulfillment{
		Order:             copyOrder(f.Order),
		AssetDists:        copyDistributions(f.AssetDists),
		PriceDists:        copyDistributions(f.PriceDists),
		AssetsFilledAmt:   copySDKInt(f.AssetsFilledAmt),
		AssetsUnfilledAmt: copySDKInt(f.AssetsUnfilledAmt),
		PriceAppliedAmt:   copySDKInt(f.PriceAppliedAmt),
		PriceLeftAmt:      copySDKInt(f.PriceLeftAmt),
		FeesToPay:         copyCoins(f.FeesToPay),
	}
}

// copyOrderFulfillments returns a deep copy of a slice of order fulfillments.
func copyOrderFulfillments(fs []*orderFulfillment) []*orderFulfillment {
	return copySlice(fs, copyOrderFulfillment)
}

// copyIndexedAddrAmts creates a deep copy of an IndexedAddrAmts.
func copyIndexedAddrAmts(orig *IndexedAddrAmts) *IndexedAddrAmts {
	if orig == nil {
		return nil
	}

	rv := &IndexedAddrAmts{
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

// distributionString is similar to %v except with easier to understand Int entries.
func distributionString(dist *distribution) string {
	if dist == nil {
		return "nil"
	}
	return fmt.Sprintf("{Address:%q, Amount:%s}", dist.Address, dist.Amount)
}

// distributionsString is similar to %v except with easier to understand Int entries.
func distributionsString(dists []*distribution) string {
	return stringerJoin(dists, distributionString, ", ")
}

// orderFulfillmentString is similar to %v except with easier to understand Coin and Int entries.
func orderFulfillmentString(f *orderFulfillment) string {
	if f == nil {
		return "nil"
	}

	fields := []string{
		fmt.Sprintf("Order:%s", orderString(f.Order)),
		fmt.Sprintf("AssetDists:%s", distributionsString(f.AssetDists)),
		fmt.Sprintf("PriceDists:%s", distributionsString(f.PriceDists)),
		fmt.Sprintf("AssetsFilledAmt:%s", f.AssetsFilledAmt),
		fmt.Sprintf("AssetsUnfilledAmt:%s", f.AssetsUnfilledAmt),
		fmt.Sprintf("PriceAppliedAmt:%s", f.PriceAppliedAmt),
		fmt.Sprintf("PriceLeftAmt:%s", f.PriceLeftAmt),
		fmt.Sprintf("FeesToPay:%s", coinsString(f.FeesToPay)),
	}
	return fmt.Sprintf("{%s}", strings.Join(fields, ", "))
}

// orderFulfillmentsString is similar to %v except with easier to understand Coin entries.
func orderFulfillmentsString(ofs []*orderFulfillment) string {
	return stringerJoin(ofs, orderFulfillmentString, ", ")
}

// bankInputString is similar to %v except with easier to understand Coin entries.
func bankInputString(i banktypes.Input) string {
	return fmt.Sprintf("I{Address:%q,Coins:%q}", i.Address, i.Coins)
}

// bankInputsString returns a string with all the provided inputs.
func bankInputsString(ins []banktypes.Input) string {
	return stringerJoin(ins, bankInputString, ", ")
}

// bankOutputString is similar to %v except with easier to understand Coin entries.
func bankOutputString(o banktypes.Output) string {
	return fmt.Sprintf("O{Address:%q,Coins:%q}", o.Address, o.Coins)
}

// bankOutputsString returns a string with all the provided outputs.
func bankOutputsString(outs []banktypes.Output) string {
	return stringerJoin(outs, bankOutputString, ", ")
}

// transferString is similar to %v except with easier to understand Coin entries.
func transferString(t *Transfer) string {
	if t == nil {
		return "nil"
	}
	inputs := bankInputsString(t.Inputs)
	outputs := bankOutputsString(t.Outputs)
	return fmt.Sprintf("T{Inputs:%s, Outputs: %s}", inputs, outputs)
}

// String converts a indexedAddrAmtsString to a string.
// This is mostly because test failure output of sdk.Coin and sdk.Coins is impossible to understand.
func indexedAddrAmtsString(i *IndexedAddrAmts) string {
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
		amts = fmt.Sprintf("%T{%s}", i.amts, strings.Join(amtsVals, ", "))
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

// assertEqualOrderFulfillments asserts that the two order fulfillments are equal.
// Returns true if equal.
// If not equal, and neither are nil, equality on each field is also asserted in order to help identify the problem.
func assertEqualOrderFulfillments(t *testing.T, expected, actual *orderFulfillment, message string, args ...interface{}) bool {
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
	assert.Equalf(t, expected.Order, actual.Order, msg("orderFulfillment.Order"), args...)
	assert.Equalf(t, expected.AssetDists, actual.AssetDists, msg("orderFulfillment.AssetDists"), args...)
	assert.Equalf(t, expected.PriceDists, actual.PriceDists, msg("orderFulfillment.PriceDists"), args...)
	assert.Equalf(t, expected.AssetsFilledAmt, actual.AssetsFilledAmt, msg("orderFulfillment.AssetsFilledAmt"), args...)
	assert.Equalf(t, expected.AssetsUnfilledAmt, actual.AssetsUnfilledAmt, msg("orderFulfillment.AssetsUnfilledAmt"), args...)
	assert.Equalf(t, expected.PriceAppliedAmt, actual.PriceAppliedAmt, msg("orderFulfillment.PriceAppliedAmt"), args...)
	assert.Equalf(t, expected.PriceLeftAmt, actual.PriceLeftAmt, msg("orderFulfillment.PriceLeftAmt"), args...)
	assert.Equalf(t, expected.FeesToPay, actual.FeesToPay, msg("orderFulfillment.FeesToPay"), args...)
	t.Logf("  Actual: %s", orderFulfillmentString(actual))
	t.Logf("Expected: %s", orderFulfillmentString(expected))
	return false
}

// assertEqualOrderFulfillmentSlices asserts that the two order fulfillments are equal.
// Returns true if equal.
// If not equal, and neither are nil, equality on each field is also asserted in order to help identify the problem.
func assertEqualOrderFulfillmentSlices(t *testing.T, expected, actual []*orderFulfillment, message string, args ...interface{}) bool {
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
	idStringer := func(of *orderFulfillment) string {
		return fmt.Sprintf("%d", of.GetOrderID())
	}
	expIDs := stringerLines(expected, idStringer)
	actIDs := stringerLines(actual, idStringer)
	if !assert.Equalf(t, expIDs, actIDs, msg("OrderIDs"), args...) {
		// Wooo, should have actionable info in the failure, so we can be done.
		return false
	}

	// Try the comparisons as strings, one per line because that's easier with ints and coins.
	expStrs := stringerLines(expected, orderFulfillmentString)
	actStrs := stringerLines(actual, orderFulfillmentString)
	if !assert.Equalf(t, expStrs, actStrs, msg("orderFulfillment strings"), args...) {
		// Wooo, should have actionable info in the failure, so we can be done.
		return false
	}

	// Alright, do it the hard way one at a time.
	for i := range expected {
		assertEqualOrderFulfillments(t, expected[i], actual[i], fmt.Sprintf("[%d]%s", i, message), args...)
	}
	t.Logf("  Actual:\n%s", strings.Join(expStrs, "\n"))
	t.Logf("Expected:\n%s", strings.Join(actStrs, "\n"))
	return false
}

func TestBuildSettlement(t *testing.T) {
	assetDenom, priceDenom := "apple", "peach"
	feeDenoms := []string{"fig", "grape"}
	feeCoins := func(tracer string, amts []int64) sdk.Coins {
		if len(amts) == 0 {
			return nil
		}
		if len(amts) > len(feeDenoms) {
			t.Fatalf("cannot create %s with more than %d fees %v", tracer, len(feeDenoms), amts)
		}
		var rv sdk.Coins
		for i, amt := range amts {
			rv = rv.Add(sdk.NewInt64Coin(feeDenoms[i], amt))
		}
		return rv
	}
	askOrder := func(orderID uint64, assets, price int64, allowPartial bool, fees ...int64) *Order {
		if len(fees) > 1 {
			t.Fatalf("cannot create ask order %d with more than 1 fees %v", orderID, fees)
		}
		var fee *sdk.Coin
		if fc := feeCoins(fmt.Sprintf("ask order %d", orderID), fees); !fc.IsZero() {
			fee = &fc[0]
		}
		return NewOrder(orderID).WithAsk(&AskOrder{
			Seller:                  fmt.Sprintf("seller%d", orderID),
			Assets:                  sdk.NewInt64Coin(assetDenom, assets),
			Price:                   sdk.NewInt64Coin(priceDenom, price),
			SellerSettlementFlatFee: fee,
			AllowPartial:            allowPartial,
		})
	}
	bidOrder := func(orderID uint64, assets, price int64, allowPartial bool, fees ...int64) *Order {
		return NewOrder(orderID).WithBid(&BidOrder{
			Buyer:               fmt.Sprintf("buyer%d", orderID),
			Assets:              sdk.NewInt64Coin(assetDenom, assets),
			Price:               sdk.NewInt64Coin(priceDenom, price),
			BuyerSettlementFees: feeCoins(fmt.Sprintf("bid order %d", orderID), fees),
			AllowPartial:        allowPartial,
		})
	}
	ratio := func(price, fee int64) func(denom string) (*FeeRatio, error) {
		return func(denom string) (*FeeRatio, error) {
			return &FeeRatio{Price: sdk.NewInt64Coin(priceDenom, price), Fee: sdk.NewInt64Coin(feeDenoms[0], fee)}, nil
		}
	}
	filled := func(order *Order, price int64, fees ...int64) *FilledOrder {
		return NewFilledOrder(order,
			sdk.NewInt64Coin(priceDenom, price),
			feeCoins(fmt.Sprintf("filled order %d", order), fees))
	}
	assetsInput := func(orderID uint64, amount int64) banktypes.Input {
		return banktypes.Input{
			Address: fmt.Sprintf("seller%d", orderID),
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(assetDenom, amount)),
		}
	}
	assetsOutput := func(orderID uint64, amount int64) banktypes.Output {
		return banktypes.Output{
			Address: fmt.Sprintf("buyer%d", orderID),
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(assetDenom, amount)),
		}
	}
	priceInput := func(orderID uint64, amount int64) banktypes.Input {
		return banktypes.Input{
			Address: fmt.Sprintf("buyer%d", orderID),
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(priceDenom, amount)),
		}
	}
	priceOutput := func(orderID uint64, amount int64) banktypes.Output {
		return banktypes.Output{
			Address: fmt.Sprintf("seller%d", orderID),
			Coins:   sdk.NewCoins(sdk.NewInt64Coin(priceDenom, amount)),
		}
	}
	feeInput := func(address string, amts ...int64) banktypes.Input {
		return banktypes.Input{Address: address, Coins: feeCoins("bank input for "+address, amts)}
	}

	tests := []struct {
		name                 string
		askOrders            []*Order
		bidOrders            []*Order
		sellerFeeRatioLookup func(denom string) (*FeeRatio, error)
		expSettlement        *Settlement
		expErr               string
	}{
		{
			// error from validateCanSettle
			name:      "no ask orders",
			askOrders: []*Order{},
			bidOrders: []*Order{bidOrder(3, 1, 10, false)},
			expErr:    "no ask orders provided",
		},
		{
			name:      "error from ratio lookup",
			askOrders: []*Order{askOrder(3, 1, 10, false)},
			bidOrders: []*Order{bidOrder(4, 1, 10, false)},
			sellerFeeRatioLookup: func(denom string) (*FeeRatio, error) {
				return nil, errors.New("this is a test error")
			},
			expErr: "this is a test error",
		},
		{
			name:      "error from setFeesToPay",
			askOrders: []*Order{askOrder(3, 1, 10, false)},
			bidOrders: []*Order{bidOrder(4, 1, 10, false)},
			sellerFeeRatioLookup: func(denom string) (*FeeRatio, error) {
				return &FeeRatio{Price: sdk.NewInt64Coin("prune", 10), Fee: sdk.NewInt64Coin("fig", 1)}, nil
			},
			expErr: "failed calculate ratio fee for ask order 3: cannot apply ratio 10prune:1fig to price 10peach: incorrect price denom",
		},
		{
			name:      "one ask, three bids: last bid not used",
			askOrders: []*Order{askOrder(3, 10, 20, false)},
			bidOrders: []*Order{
				bidOrder(4, 7, 14, false),
				bidOrder(5, 3, 6, false),
				bidOrder(6, 1, 2, false),
			},
			expErr: "bid order 6 (at index 2) has no assets filled",
		},
		{
			name: "three asks, one bids: last ask not used",
			askOrders: []*Order{
				askOrder(11, 7, 14, false),
				askOrder(12, 3, 14, false),
				askOrder(13, 1, 14, false),
			},
			bidOrders: []*Order{bidOrder(14, 10, 20, false)},
			expErr:    "ask order 13 (at index 2) has no assets filled",
		},
		{
			name: "two asks, two bids: same assets total, total bid price not enough",
			askOrders: []*Order{
				askOrder(1, 10, 25, false),
				askOrder(2, 10, 15, false),
			},
			bidOrders: []*Order{
				bidOrder(8, 10, 20, false),
				bidOrder(9, 10, 19, false),
			},
			expErr: "total ask price \"40peach\" is greater than total bid price \"39peach\"",
		},
		{
			name: "two asks, two bids: ask partial, total bid price not enough",
			askOrders: []*Order{
				askOrder(1, 10, 20, false),
				askOrder(2, 10, 20, true),
			},
			bidOrders: []*Order{
				bidOrder(8, 10, 20, false),
				bidOrder(9, 9, 17, false),
			},
			expErr: "total ask price \"38peach\" is greater than total bid price \"37peach\"",
		},
		{
			name: "two asks, two bids: bid partial, total bid price not enough",
			askOrders: []*Order{
				askOrder(1, 10, 25, false),
				askOrder(2, 10, 15, false),
			},
			bidOrders: []*Order{
				bidOrder(8, 10, 19, false),
				bidOrder(9, 11, 22, true),
			},
			expErr: "total ask price \"40peach\" is greater than total bid price \"39peach\"",
		},
		{
			name:                 "one ask, one bid: both fully filled",
			askOrders:            []*Order{askOrder(52, 10, 100, false, 2)},
			bidOrders:            []*Order{bidOrder(11, 10, 105, false, 3, 4)},
			sellerFeeRatioLookup: ratio(4, 1),
			expSettlement: &Settlement{
				Transfers: []*Transfer{
					{Inputs: []banktypes.Input{assetsInput(52, 10)}, Outputs: []banktypes.Output{assetsOutput(11, 10)}},
					{Inputs: []banktypes.Input{priceInput(11, 105)}, Outputs: []banktypes.Output{priceOutput(52, 105)}},
				},
				FeeInputs: []banktypes.Input{
					feeInput("seller52", 29),
					feeInput("buyer11", 3, 4),
				},
				FullyFilledOrders: []*FilledOrder{
					filled(askOrder(52, 10, 100, false, 2), 105, 29),
					filled(bidOrder(11, 10, 105, false, 3, 4), 105, 3, 4),
				},
			},
		},
		{
			name:      "one ask, one bid: ask partially filled",
			askOrders: []*Order{askOrder(99, 10, 100, true)},
			bidOrders: []*Order{bidOrder(15, 9, 90, false)},
			expSettlement: &Settlement{
				Transfers: []*Transfer{
					{Inputs: []banktypes.Input{assetsInput(99, 9)}, Outputs: []banktypes.Output{assetsOutput(15, 9)}},
					{Inputs: []banktypes.Input{priceInput(15, 90)}, Outputs: []banktypes.Output{priceOutput(99, 90)}},
				},
				FullyFilledOrders:  []*FilledOrder{filled(bidOrder(15, 9, 90, false), 90)},
				PartialOrderFilled: filled(askOrder(99, 9, 90, true), 90),
				PartialOrderLeft:   askOrder(99, 1, 10, true),
			},
		},
		{
			name:      "one ask, one bid: ask partially filled, not allowed",
			askOrders: []*Order{askOrder(99, 10, 100, false)},
			bidOrders: []*Order{bidOrder(15, 9, 90, false)},
			expErr:    "cannot split ask order 99 having assets \"10apple\" at \"9apple\": order does not allow partial fulfillment",
		},
		{
			name:      "one ask, one bid: bid partially filled",
			askOrders: []*Order{askOrder(8, 9, 85, false, 2)},
			bidOrders: []*Order{bidOrder(12, 10, 100, true)},
			expSettlement: &Settlement{
				Transfers: []*Transfer{
					{Inputs: []banktypes.Input{assetsInput(8, 9)}, Outputs: []banktypes.Output{assetsOutput(12, 9)}},
					{Inputs: []banktypes.Input{priceInput(12, 90)}, Outputs: []banktypes.Output{priceOutput(8, 90)}},
				},
				FeeInputs:          []banktypes.Input{feeInput("seller8", 2)},
				FullyFilledOrders:  []*FilledOrder{filled(askOrder(8, 9, 85, false, 2), 90, 2)},
				PartialOrderFilled: filled(bidOrder(12, 9, 90, true), 90),
				PartialOrderLeft:   bidOrder(12, 1, 10, true),
			},
		},
		{
			name:      "one ask, one bid: bid partially filled, not allowed",
			askOrders: []*Order{askOrder(8, 9, 85, false, 2)},
			bidOrders: []*Order{bidOrder(12, 10, 100, false)},
			expErr:    "cannot split bid order 12 having assets \"10apple\" at \"9apple\": order does not allow partial fulfillment",
		},
		{
			name:      "one ask, five bids: all fully filled",
			askOrders: []*Order{askOrder(999, 130, 260, false)},
			bidOrders: []*Order{
				bidOrder(11, 71, 140, false),
				bidOrder(12, 10, 20, false, 5),
				bidOrder(13, 4, 12, false),
				bidOrder(14, 11, 22, false),
				bidOrder(15, 34, 68, false, 8),
			},
			sellerFeeRatioLookup: ratio(65, 3),
			expSettlement: &Settlement{
				Transfers: []*Transfer{
					{
						Inputs: []banktypes.Input{assetsInput(999, 130)},
						Outputs: []banktypes.Output{
							assetsOutput(11, 71),
							assetsOutput(12, 10),
							assetsOutput(13, 4),
							assetsOutput(14, 11),
							assetsOutput(15, 34),
						},
					},
					{Inputs: []banktypes.Input{priceInput(11, 140)}, Outputs: []banktypes.Output{priceOutput(999, 140)}},
					{Inputs: []banktypes.Input{priceInput(12, 20)}, Outputs: []banktypes.Output{priceOutput(999, 20)}},
					{Inputs: []banktypes.Input{priceInput(13, 12)}, Outputs: []banktypes.Output{priceOutput(999, 12)}},
					{Inputs: []banktypes.Input{priceInput(14, 22)}, Outputs: []banktypes.Output{priceOutput(999, 22)}},
					{Inputs: []banktypes.Input{priceInput(15, 68)}, Outputs: []banktypes.Output{priceOutput(999, 68)}},
				},
				FeeInputs: []banktypes.Input{
					feeInput("seller999", 13),
					feeInput("buyer12", 5),
					feeInput("buyer15", 8),
				},
				FullyFilledOrders: []*FilledOrder{
					filled(askOrder(999, 130, 260, false), 262, 13),
					filled(bidOrder(11, 71, 140, false), 140),
					filled(bidOrder(12, 10, 20, false, 5), 20, 5),
					filled(bidOrder(13, 4, 12, false), 12),
					filled(bidOrder(14, 11, 22, false), 22),
					filled(bidOrder(15, 34, 68, false, 8), 68, 8),
				},
			},
		},
		{
			name:      "one ask, five bids: ask partially filled",
			askOrders: []*Order{askOrder(999, 131, 262, true)},
			bidOrders: []*Order{
				bidOrder(11, 71, 140, false),
				bidOrder(12, 10, 20, false),
				bidOrder(13, 4, 12, false),
				bidOrder(14, 11, 22, false),
				bidOrder(15, 34, 68, false),
			},
			expSettlement: &Settlement{
				Transfers: []*Transfer{
					{
						Inputs: []banktypes.Input{assetsInput(999, 130)},
						Outputs: []banktypes.Output{
							assetsOutput(11, 71),
							assetsOutput(12, 10),
							assetsOutput(13, 4),
							assetsOutput(14, 11),
							assetsOutput(15, 34),
						},
					},
					{Inputs: []banktypes.Input{priceInput(11, 140)}, Outputs: []banktypes.Output{priceOutput(999, 140)}},
					{Inputs: []banktypes.Input{priceInput(12, 20)}, Outputs: []banktypes.Output{priceOutput(999, 20)}},
					{Inputs: []banktypes.Input{priceInput(13, 12)}, Outputs: []banktypes.Output{priceOutput(999, 12)}},
					{Inputs: []banktypes.Input{priceInput(14, 22)}, Outputs: []banktypes.Output{priceOutput(999, 22)}},
					{Inputs: []banktypes.Input{priceInput(15, 68)}, Outputs: []banktypes.Output{priceOutput(999, 68)}},
				},
				FeeInputs: nil,
				FullyFilledOrders: []*FilledOrder{
					filled(bidOrder(11, 71, 140, false), 140),
					filled(bidOrder(12, 10, 20, false), 20),
					filled(bidOrder(13, 4, 12, false), 12),
					filled(bidOrder(14, 11, 22, false), 22),
					filled(bidOrder(15, 34, 68, false), 68),
				},
				PartialOrderFilled: filled(askOrder(999, 130, 260, true), 262),
				PartialOrderLeft:   askOrder(999, 1, 2, true),
			},
		},
		{
			name:      "one ask, five bids: bid partially filled",
			askOrders: []*Order{askOrder(999, 130, 260, false)},
			bidOrders: []*Order{
				bidOrder(11, 71, 140, false),
				bidOrder(12, 10, 20, false),
				bidOrder(13, 4, 12, false),
				bidOrder(14, 11, 22, false),
				bidOrder(15, 35, 70, true),
			},
			expSettlement: &Settlement{
				Transfers: []*Transfer{
					{
						Inputs: []banktypes.Input{assetsInput(999, 130)},
						Outputs: []banktypes.Output{
							assetsOutput(11, 71),
							assetsOutput(12, 10),
							assetsOutput(13, 4),
							assetsOutput(14, 11),
							assetsOutput(15, 34),
						},
					},
					{Inputs: []banktypes.Input{priceInput(11, 140)}, Outputs: []banktypes.Output{priceOutput(999, 140)}},
					{Inputs: []banktypes.Input{priceInput(12, 20)}, Outputs: []banktypes.Output{priceOutput(999, 20)}},
					{Inputs: []banktypes.Input{priceInput(13, 12)}, Outputs: []banktypes.Output{priceOutput(999, 12)}},
					{Inputs: []banktypes.Input{priceInput(14, 22)}, Outputs: []banktypes.Output{priceOutput(999, 22)}},
					{Inputs: []banktypes.Input{priceInput(15, 68)}, Outputs: []banktypes.Output{priceOutput(999, 68)}},
				},
				FeeInputs: nil,
				FullyFilledOrders: []*FilledOrder{
					filled(askOrder(999, 130, 260, false), 262),
					filled(bidOrder(11, 71, 140, false), 140),
					filled(bidOrder(12, 10, 20, false), 20),
					filled(bidOrder(13, 4, 12, false), 12),
					filled(bidOrder(14, 11, 22, false), 22),
				},
				PartialOrderFilled: filled(bidOrder(15, 34, 68, true), 68),
				PartialOrderLeft:   bidOrder(15, 1, 2, true),
			},
		},
		{
			name: "five ask, one bids: all fully filled",
			askOrders: []*Order{
				askOrder(51, 37, 74, false, 1),
				askOrder(52, 21, 42, false, 2),
				askOrder(53, 15, 30, false, 3),
				askOrder(54, 9, 18, false, 4),
				askOrder(55, 55, 110, false, 5),
			},
			bidOrders: []*Order{bidOrder(777, 137, 280, false, 7)},
			expSettlement: &Settlement{
				Transfers: []*Transfer{
					{Inputs: []banktypes.Input{assetsInput(51, 37)}, Outputs: []banktypes.Output{assetsOutput(777, 37)}},
					{Inputs: []banktypes.Input{assetsInput(52, 21)}, Outputs: []banktypes.Output{assetsOutput(777, 21)}},
					{Inputs: []banktypes.Input{assetsInput(53, 15)}, Outputs: []banktypes.Output{assetsOutput(777, 15)}},
					{Inputs: []banktypes.Input{assetsInput(54, 9)}, Outputs: []banktypes.Output{assetsOutput(777, 9)}},
					{Inputs: []banktypes.Input{assetsInput(55, 55)}, Outputs: []banktypes.Output{assetsOutput(777, 55)}},
					{
						Inputs: []banktypes.Input{priceInput(777, 280)},
						Outputs: []banktypes.Output{
							priceOutput(51, 76),
							priceOutput(52, 43),
							priceOutput(53, 31),
							priceOutput(54, 18),
							priceOutput(55, 112),
						},
					},
				},
				FeeInputs: []banktypes.Input{
					feeInput("seller51", 1),
					feeInput("seller52", 2),
					feeInput("seller53", 3),
					feeInput("seller54", 4),
					feeInput("seller55", 5),
					feeInput("buyer777", 7),
				},
				FullyFilledOrders: []*FilledOrder{
					filled(askOrder(51, 37, 74, false, 1), 76, 1),
					filled(askOrder(52, 21, 42, false, 2), 43, 2),
					filled(askOrder(53, 15, 30, false, 3), 31, 3),
					filled(askOrder(54, 9, 18, false, 4), 18, 4),
					filled(askOrder(55, 55, 110, false, 5), 112, 5),
					filled(bidOrder(777, 137, 280, false, 7), 280, 7),
				},
			},
		},
		{
			name: "five ask, one bids: ask partially filled",
			askOrders: []*Order{
				askOrder(51, 37, 74, false),
				askOrder(52, 21, 42, false),
				askOrder(53, 15, 30, false),
				askOrder(54, 9, 18, false),
				askOrder(55, 57, 114, true),
			},
			bidOrders:            []*Order{bidOrder(777, 137, 280, false)},
			sellerFeeRatioLookup: ratio(50, 1),
			expSettlement: &Settlement{
				Transfers: []*Transfer{
					{Inputs: []banktypes.Input{assetsInput(51, 37)}, Outputs: []banktypes.Output{assetsOutput(777, 37)}},
					{Inputs: []banktypes.Input{assetsInput(52, 21)}, Outputs: []banktypes.Output{assetsOutput(777, 21)}},
					{Inputs: []banktypes.Input{assetsInput(53, 15)}, Outputs: []banktypes.Output{assetsOutput(777, 15)}},
					{Inputs: []banktypes.Input{assetsInput(54, 9)}, Outputs: []banktypes.Output{assetsOutput(777, 9)}},
					{Inputs: []banktypes.Input{assetsInput(55, 55)}, Outputs: []banktypes.Output{assetsOutput(777, 55)}},
					{
						Inputs: []banktypes.Input{priceInput(777, 280)},
						Outputs: []banktypes.Output{
							priceOutput(51, 76),
							priceOutput(52, 43),
							priceOutput(53, 31),
							priceOutput(54, 18),
							priceOutput(55, 112),
						},
					},
				},
				FeeInputs: []banktypes.Input{
					feeInput("seller51", 2),
					feeInput("seller52", 1),
					feeInput("seller53", 1),
					feeInput("seller54", 1),
					feeInput("seller55", 3),
				},
				FullyFilledOrders: []*FilledOrder{
					filled(askOrder(51, 37, 74, false), 76, 2),
					filled(askOrder(52, 21, 42, false), 43, 1),
					filled(askOrder(53, 15, 30, false), 31, 1),
					filled(askOrder(54, 9, 18, false), 18, 1),
					filled(bidOrder(777, 137, 280, false), 280),
				},
				PartialOrderFilled: filled(askOrder(55, 55, 110, true), 112, 3),
				PartialOrderLeft:   askOrder(55, 2, 4, true),
			},
		},
		{
			name: "five ask, one bids: bid partially filled",
			askOrders: []*Order{
				askOrder(51, 37, 74, false),
				askOrder(52, 21, 42, false),
				askOrder(53, 15, 30, false),
				askOrder(54, 9, 18, false),
				askOrder(55, 55, 110, false),
			},
			bidOrders: []*Order{bidOrder(777, 274, 560, true)},
			expSettlement: &Settlement{
				Transfers: []*Transfer{
					{Inputs: []banktypes.Input{assetsInput(51, 37)}, Outputs: []banktypes.Output{assetsOutput(777, 37)}},
					{Inputs: []banktypes.Input{assetsInput(52, 21)}, Outputs: []banktypes.Output{assetsOutput(777, 21)}},
					{Inputs: []banktypes.Input{assetsInput(53, 15)}, Outputs: []banktypes.Output{assetsOutput(777, 15)}},
					{Inputs: []banktypes.Input{assetsInput(54, 9)}, Outputs: []banktypes.Output{assetsOutput(777, 9)}},
					{Inputs: []banktypes.Input{assetsInput(55, 55)}, Outputs: []banktypes.Output{assetsOutput(777, 55)}},
					{
						Inputs: []banktypes.Input{priceInput(777, 280)},
						Outputs: []banktypes.Output{
							priceOutput(51, 76),
							priceOutput(52, 43),
							priceOutput(53, 31),
							priceOutput(54, 18),
							priceOutput(55, 112),
						},
					},
				},
				FullyFilledOrders: []*FilledOrder{
					filled(askOrder(51, 37, 74, false), 76),
					filled(askOrder(52, 21, 42, false), 43),
					filled(askOrder(53, 15, 30, false), 31),
					filled(askOrder(54, 9, 18, false), 18),
					filled(askOrder(55, 55, 110, false), 112),
				},
				PartialOrderFilled: filled(bidOrder(777, 137, 280, true), 280),
				PartialOrderLeft:   bidOrder(777, 137, 280, true),
			},
		},
		{
			name: "two asks, three bids: all fully filled",
			askOrders: []*Order{
				askOrder(11, 100, 1000, false),
				askOrder(22, 200, 2000, false),
			},
			bidOrders: []*Order{
				bidOrder(33, 75, 700, false),
				bidOrder(44, 130, 1302, false),
				bidOrder(55, 95, 1000, false),
			},
			expSettlement: &Settlement{
				Transfers: []*Transfer{
					{
						Inputs: []banktypes.Input{assetsInput(11, 100)},
						Outputs: []banktypes.Output{
							assetsOutput(33, 75),
							assetsOutput(44, 25),
						},
					},
					{
						Inputs: []banktypes.Input{assetsInput(22, 200)},
						Outputs: []banktypes.Output{
							assetsOutput(44, 105),
							assetsOutput(55, 95),
						},
					},
					{
						Inputs:  []banktypes.Input{priceInput(33, 700)},
						Outputs: []banktypes.Output{priceOutput(11, 700)},
					},
					{
						Inputs: []banktypes.Input{priceInput(44, 1302)},
						Outputs: []banktypes.Output{
							priceOutput(11, 300),
							priceOutput(22, 1002),
						},
					},
					{
						Inputs: []banktypes.Input{priceInput(55, 1000)},
						Outputs: []banktypes.Output{
							priceOutput(22, 999),
							priceOutput(11, 1),
						},
					},
				},
				FullyFilledOrders: []*FilledOrder{
					filled(askOrder(11, 100, 1000, false), 1001),
					filled(askOrder(22, 200, 2000, false), 2001),
					filled(bidOrder(33, 75, 700, false), 700),
					filled(bidOrder(44, 130, 1302, false), 1302),
					filled(bidOrder(55, 95, 1000, false), 1000),
				},
			},
		},
		{
			name: "two asks, three bids: ask partially filled",
			askOrders: []*Order{
				askOrder(11, 100, 1000, false),
				askOrder(22, 300, 3000, true),
			},
			bidOrders: []*Order{
				bidOrder(33, 75, 700, false),
				bidOrder(44, 130, 1302, false),
				bidOrder(55, 95, 1000, false),
			},
			sellerFeeRatioLookup: ratio(100, 1),
			expSettlement: &Settlement{
				Transfers: []*Transfer{
					{
						Inputs: []banktypes.Input{assetsInput(11, 100)},
						Outputs: []banktypes.Output{
							assetsOutput(33, 75),
							assetsOutput(44, 25),
						},
					},
					{
						Inputs: []banktypes.Input{assetsInput(22, 200)},
						Outputs: []banktypes.Output{
							assetsOutput(44, 105),
							assetsOutput(55, 95),
						},
					},
					{
						Inputs:  []banktypes.Input{priceInput(33, 700)},
						Outputs: []banktypes.Output{priceOutput(11, 700)},
					},
					{
						Inputs: []banktypes.Input{priceInput(44, 1302)},
						Outputs: []banktypes.Output{
							priceOutput(11, 300),
							priceOutput(22, 1002),
						},
					},
					{
						Inputs: []banktypes.Input{priceInput(55, 1000)},
						Outputs: []banktypes.Output{
							priceOutput(22, 999),
							priceOutput(11, 1),
						},
					},
				},
				FeeInputs: []banktypes.Input{
					feeInput("seller11", 11),
					feeInput("seller22", 21),
				},
				FullyFilledOrders: []*FilledOrder{
					filled(askOrder(11, 100, 1000, false), 1001, 11),
					filled(bidOrder(33, 75, 700, false), 700),
					filled(bidOrder(44, 130, 1302, false), 1302),
					filled(bidOrder(55, 95, 1000, false), 1000),
				},
				PartialOrderFilled: filled(askOrder(22, 200, 2000, true), 2001, 21),
				PartialOrderLeft:   askOrder(22, 100, 1000, true),
			},
		},
		{
			name: "two asks, three bids: ask partially filled, not allowed",
			askOrders: []*Order{
				askOrder(11, 100, 1000, true),
				askOrder(22, 300, 3000, false),
			},
			bidOrders: []*Order{
				bidOrder(33, 75, 700, true),
				bidOrder(44, 130, 1302, true),
				bidOrder(55, 95, 1000, true),
			},
			expErr: "cannot split ask order 22 having assets \"300apple\" at \"200apple\": order does not allow partial fulfillment",
		},
		{
			name: "two asks, three bids: bid partially filled",
			askOrders: []*Order{
				askOrder(11, 100, 1000, false),
				askOrder(22, 200, 2000, false),
			},
			bidOrders: []*Order{
				bidOrder(33, 75, 700, false),
				bidOrder(44, 130, 1352, false),
				bidOrder(55, 100, 1000, true, 40, 20),
			},
			expSettlement: &Settlement{
				Transfers: []*Transfer{
					{
						Inputs: []banktypes.Input{assetsInput(11, 100)},
						Outputs: []banktypes.Output{
							assetsOutput(33, 75),
							assetsOutput(44, 25),
						},
					},
					{
						Inputs: []banktypes.Input{assetsInput(22, 200)},
						Outputs: []banktypes.Output{
							assetsOutput(44, 105),
							assetsOutput(55, 95),
						},
					},
					{
						Inputs:  []banktypes.Input{priceInput(33, 700)},
						Outputs: []banktypes.Output{priceOutput(11, 700)},
					},
					{
						Inputs: []banktypes.Input{priceInput(44, 1352)},
						Outputs: []banktypes.Output{
							priceOutput(11, 300),
							priceOutput(22, 1052),
						},
					},
					{
						Inputs: []banktypes.Input{priceInput(55, 950)},
						Outputs: []banktypes.Output{
							priceOutput(22, 949),
							priceOutput(11, 1),
						},
					},
				},
				FeeInputs: []banktypes.Input{feeInput("buyer55", 38, 19)},
				FullyFilledOrders: []*FilledOrder{
					filled(askOrder(11, 100, 1000, false), 1001),
					filled(askOrder(22, 200, 2000, false), 2001),
					filled(bidOrder(33, 75, 700, false), 700),
					filled(bidOrder(44, 130, 1352, false), 1352),
				},
				PartialOrderFilled: filled(bidOrder(55, 95, 950, true, 38, 19), 950, 38, 19),
				PartialOrderLeft:   bidOrder(55, 5, 50, true, 2, 1),
			},
		},
		{
			name: "two asks, three bids: bid partially filled, not allowed",
			askOrders: []*Order{
				askOrder(11, 100, 1000, true),
				askOrder(22, 200, 2000, true),
			},
			bidOrders: []*Order{
				bidOrder(33, 75, 700, true),
				bidOrder(44, 130, 1352, true),
				bidOrder(55, 100, 1000, false, 40, 20),
			},
			expErr: "cannot split bid order 55 having assets \"100apple\" at \"95apple\": order does not allow partial fulfillment",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.sellerFeeRatioLookup == nil {
				tc.sellerFeeRatioLookup = func(denom string) (*FeeRatio, error) {
					return nil, nil
				}
			}
			var settlement *Settlement
			var err error
			testFunc := func() {
				settlement, err = BuildSettlement(tc.askOrders, tc.bidOrders, tc.sellerFeeRatioLookup)
			}
			require.NotPanics(t, testFunc, "BuildSettlement")
			assertions.RequireErrorValue(t, err, tc.expErr, "BuildSettlement error")
			if !assert.Equal(t, tc.expSettlement, settlement, "BuildSettlement result") {
				// Doing each field on its own now to try to help pinpoint the differences.
				expTrans := stringerLines(tc.expSettlement.Transfers, transferString)
				actTrans := stringerLines(settlement.Transfers, transferString)
				assert.Equal(t, expTrans, actTrans, "Transfers (as strings)")

				expFeeInputs := bankInputsString(tc.expSettlement.FeeInputs)
				actFeeInputs := bankInputsString(settlement.FeeInputs)
				assert.Equal(t, expFeeInputs, actFeeInputs, "FeeInputs (as strings)")

				idStringer := func(of *FilledOrder) string {
					return fmt.Sprintf("%d", of.GetOrderID())
				}
				expFilledIDs := stringerLines(tc.expSettlement.FullyFilledOrders, idStringer)
				actFilledIDs := stringerLines(settlement.FullyFilledOrders, idStringer)
				if assert.Equal(t, expFilledIDs, actFilledIDs, "FullyFilledOrders ids") {
					// If they're the same ids, compare each individually.
					for i := range tc.expSettlement.FullyFilledOrders {
						assert.Equal(t, tc.expSettlement.FullyFilledOrders[i], settlement.FullyFilledOrders[i], "FullyFilledOrders[%d]", i)
					}
				}

				assert.Equal(t, tc.expSettlement.PartialOrderFilled, settlement.PartialOrderFilled, "PartialOrderFilled")
				assert.Equal(t, tc.expSettlement.PartialOrderLeft, settlement.PartialOrderLeft, "PartialOrderLeft")
			}
		})
	}
}

func TestNewIndexedAddrAmts(t *testing.T) {
	expected := &IndexedAddrAmts{
		addrs:   nil,
		amts:    nil,
		indexes: make(map[string]int),
	}
	actual := NewIndexedAddrAmts()
	assert.Equal(t, expected, actual, "NewIndexedAddrAmts result")
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
		receiver *IndexedAddrAmts
		addr     string
		coins    []sdk.Coin
		expected *IndexedAddrAmts
		expPanic string
	}{
		{
			name:     "empty, add zero coins",
			receiver: NewIndexedAddrAmts(),
			addr:     "addr1",
			coins:    nil,
			expected: NewIndexedAddrAmts(),
		},
		{
			name:     "empty, add one coin with a zero value",
			receiver: NewIndexedAddrAmts(),
			addr:     "addr1",
			coins:    []sdk.Coin{{Denom: "zero", Amount: sdkmath.ZeroInt()}},
			expected: NewIndexedAddrAmts(),
		},
		{
			name:     "empty, add two coins with a zero value.",
			receiver: NewIndexedAddrAmts(),
			addr:     "addr1",
			coins: []sdk.Coin{
				{Denom: "zeroa", Amount: sdkmath.ZeroInt()},
				{Denom: "zerob", Amount: sdkmath.ZeroInt()},
			},
			expected: NewIndexedAddrAmts(),
		},
		{
			name:     "empty, add three coins, only one is not zero",
			receiver: NewIndexedAddrAmts(),
			addr:     "addr1",
			coins: []sdk.Coin{
				{Denom: "coina", Amount: sdkmath.ZeroInt()},
				{Denom: "coinb", Amount: sdkmath.OneInt()},
				{Denom: "coinc", Amount: sdkmath.ZeroInt()},
			},
			expected: &IndexedAddrAmts{
				addrs:   []string{"addr1"},
				amts:    []sdk.Coins{coins("1coinb")},
				indexes: map[string]int{"addr1": 0},
			},
		},
		{
			name:     "empty, add one coin",
			receiver: NewIndexedAddrAmts(),
			addr:     "addr1",
			coins:    coins("1one"),
			expected: &IndexedAddrAmts{
				addrs:   []string{"addr1"},
				amts:    []sdk.Coins{coins("1one")},
				indexes: map[string]int{"addr1": 0},
			},
		},
		{
			name:     "empty, add two coins",
			receiver: NewIndexedAddrAmts(),
			addr:     "addr1",
			coins:    coins("1one,2two"),
			expected: &IndexedAddrAmts{
				addrs:   []string{"addr1"},
				amts:    []sdk.Coins{coins("1one,2two")},
				indexes: map[string]int{"addr1": 0},
			},
		},
		{
			name:     "empty, add neg coins",
			receiver: NewIndexedAddrAmts(),
			addr:     "addr1",
			coins:    negCoins,
			expPanic: "cannot index and add invalid coin amount \"-1neg\"",
		},
		{
			name: "one addr, add to existing new denom",
			receiver: &IndexedAddrAmts{
				addrs:   []string{"addr1"},
				amts:    []sdk.Coins{coins("1one")},
				indexes: map[string]int{"addr1": 0},
			},
			addr:  "addr1",
			coins: coins("2two"),
			expected: &IndexedAddrAmts{
				addrs:   []string{"addr1"},
				amts:    []sdk.Coins{coins("1one,2two")},
				indexes: map[string]int{"addr1": 0},
			},
		},
		{
			name: "one addr, add to existing same denom",
			receiver: &IndexedAddrAmts{
				addrs:   []string{"addr1"},
				amts:    []sdk.Coins{coins("1one")},
				indexes: map[string]int{"addr1": 0},
			},
			addr:  "addr1",
			coins: coins("3one"),
			expected: &IndexedAddrAmts{
				addrs:   []string{"addr1"},
				amts:    []sdk.Coins{coins("4one")},
				indexes: map[string]int{"addr1": 0},
			},
		},
		{
			name: "one addr, add negative to existing",
			receiver: &IndexedAddrAmts{
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
			receiver: &IndexedAddrAmts{
				addrs:   []string{"addr1"},
				amts:    []sdk.Coins{coins("1one")},
				indexes: map[string]int{"addr1": 0},
			},
			addr:  "addr2",
			coins: coins("2two"),
			expected: &IndexedAddrAmts{
				addrs:   []string{"addr1", "addr2"},
				amts:    []sdk.Coins{coins("1one"), coins("2two")},
				indexes: map[string]int{"addr1": 0, "addr2": 1},
			},
		},
		{
			name: "one addr, add to new opposite order",
			receiver: &IndexedAddrAmts{
				addrs:   []string{"addr2"},
				amts:    []sdk.Coins{coins("2two")},
				indexes: map[string]int{"addr2": 0},
			},
			addr:  "addr1",
			coins: coins("1one"),
			expected: &IndexedAddrAmts{
				addrs:   []string{"addr2", "addr1"},
				amts:    []sdk.Coins{coins("2two"), coins("1one")},
				indexes: map[string]int{"addr2": 0, "addr1": 1},
			},
		},
		{
			name: "one addr, add negative to new",
			receiver: &IndexedAddrAmts{
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
			receiver: &IndexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("2two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
			addr:  "addr1",
			coins: coins("10one"),
			expected: &IndexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("11one"), coins("2two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
		},
		{
			name: "three addrs, add to second",
			receiver: &IndexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("2two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
			addr:  "addr2",
			coins: coins("10two"),
			expected: &IndexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("12two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
		},
		{
			name: "three addrs, add to third",
			receiver: &IndexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("2two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
			addr:  "addr3",
			coins: coins("10three"),
			expected: &IndexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("2two"), coins("13three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
		},
		{
			name: "three addrs, add two coins to second",
			receiver: &IndexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("2two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
			addr:  "addr2",
			coins: coins("10four,20two"),
			expected: &IndexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("10four,22two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
		},
		{
			name: "three addrs, add to new",
			receiver: &IndexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("2two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
			addr:  "good buddy",
			coins: coins("10four"),
			expected: &IndexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3", "good buddy"},
				amts:    []sdk.Coins{coins("1one"), coins("2two"), coins("3three"), coins("10four")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2, "good buddy": 3},
			},
		},
		{
			name: "three addrs, add negative to second",
			receiver: &IndexedAddrAmts{
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
			receiver: &IndexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("2two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
			addr:     "addr4",
			coins:    negCoins,
			expPanic: "cannot index and add invalid coin amount \"-1neg\"",
		},
		{
			name: "three addrs, add zero to existing",
			receiver: &IndexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("2two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
			addr:  "addr2",
			coins: []sdk.Coin{{Denom: "zero", Amount: sdkmath.ZeroInt()}},
			expected: &IndexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("2two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
		},
		{
			name: "three addrs, add zero to new",
			receiver: &IndexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("2two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
			addr:  "addr4",
			coins: []sdk.Coin{{Denom: "zero", Amount: sdkmath.ZeroInt()}},
			expected: &IndexedAddrAmts{
				addrs:   []string{"addr1", "addr2", "addr3"},
				amts:    []sdk.Coins{coins("1one"), coins("2two"), coins("3three")},
				indexes: map[string]int{"addr1": 0, "addr2": 1, "addr3": 2},
			},
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
				tc.receiver.Add(tc.addr, tc.coins...)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "Add(%q, %q)", tc.addr, tc.coins)
			if len(tc.expPanic) == 0 {
				assert.Equal(t, tc.expected, tc.receiver, "receiver after Add(%q, %q)", tc.addr, tc.coins)
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
		receiver *IndexedAddrAmts
		expected []banktypes.Input
		expPanic string
	}{
		{name: "nil receiver", receiver: nil, expected: nil},
		{name: "no addrs", receiver: NewIndexedAddrAmts(), expected: nil},
		{
			name: "one addr negative amount",
			receiver: &IndexedAddrAmts{
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
			receiver: &IndexedAddrAmts{
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
			receiver: &IndexedAddrAmts{
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
			receiver: &IndexedAddrAmts{
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
			receiver: &IndexedAddrAmts{
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
			receiver: &IndexedAddrAmts{
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
				actual = tc.receiver.GetAsInputs()
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "GetAsInputs()")
			assert.Equal(t, tc.expected, actual, "GetAsInputs() result")
			if !assert.Equal(t, orig, tc.receiver, "receiver before and after GetAsInputs()") {
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
		receiver *IndexedAddrAmts
		expected []banktypes.Output
		expPanic string
	}{
		{name: "nil receiver", receiver: nil, expected: nil},
		{name: "no addrs", receiver: NewIndexedAddrAmts(), expected: nil},
		{
			name: "one addr negative amount",
			receiver: &IndexedAddrAmts{
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
			receiver: &IndexedAddrAmts{
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
			receiver: &IndexedAddrAmts{
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
			receiver: &IndexedAddrAmts{
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
			receiver: &IndexedAddrAmts{
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
			receiver: &IndexedAddrAmts{
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
				actual = tc.receiver.GetAsOutputs()
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "GetAsOutputs()")
			assert.Equal(t, tc.expected, actual, "GetAsOutputs() result")
			if !assert.Equal(t, orig, tc.receiver, "receiver before and after GetAsInputs()") {
				t.Logf("Before: %s", indexedAddrAmtsString(orig))
				t.Logf(" After: %s", indexedAddrAmtsString(tc.receiver))
			}
		})
	}
}

func TestNewOrderFulfillment(t *testing.T) {
	tests := []struct {
		name     string
		order    *Order
		expected *orderFulfillment
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
			expected: &orderFulfillment{
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
				AssetsFilledAmt:   sdkmath.ZeroInt(),
				AssetsUnfilledAmt: sdkmath.NewInt(92),
				PriceAppliedAmt:   sdkmath.ZeroInt(),
				PriceLeftAmt:      sdkmath.NewInt(15),
				FeesToPay:         nil,
			},
		},
		{
			name: "bid order",
			order: NewOrder(3).WithBid(&BidOrder{
				MarketId: 11,
				Assets:   sdk.NewInt64Coin("adolla", 93),
				Price:    sdk.NewInt64Coin("pdolla", 16),
			}),
			expected: &orderFulfillment{
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
				AssetsFilledAmt:   sdkmath.ZeroInt(),
				AssetsUnfilledAmt: sdkmath.NewInt(93),
				PriceAppliedAmt:   sdkmath.ZeroInt(),
				PriceLeftAmt:      sdkmath.NewInt(16),
				FeesToPay:         nil,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual *orderFulfillment
			defer func() {
				if t.Failed() {
					t.Logf("  Actual: %s", orderFulfillmentString(actual))
					t.Logf("Expected: %s", orderFulfillmentString(tc.expected))
				}
			}()

			testFunc := func() {
				actual = newOrderFulfillment(tc.order)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "newOrderFulfillment")
			assert.Equal(t, tc.expected, actual, "newOrderFulfillment result")
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
		expected []*orderFulfillment
	}{
		{
			name:     "nil orders",
			orders:   nil,
			expected: []*orderFulfillment{},
		},
		{
			name:     "empty orders",
			orders:   []*Order{},
			expected: []*orderFulfillment{},
		},
		{
			name:     "1 ask order",
			orders:   []*Order{askOrders[0]},
			expected: []*orderFulfillment{newOrderFulfillment(askOrders[0])},
		},
		{
			name:     "1 bid order",
			orders:   []*Order{bidOrders[0]},
			expected: []*orderFulfillment{newOrderFulfillment(bidOrders[0])},
		},
		{
			name:   "4 ask orders",
			orders: []*Order{askOrders[0], askOrders[1], askOrders[2], askOrders[3]},
			expected: []*orderFulfillment{
				newOrderFulfillment(askOrders[0]),
				newOrderFulfillment(askOrders[1]),
				newOrderFulfillment(askOrders[2]),
				newOrderFulfillment(askOrders[3]),
			},
		},
		{
			name:   "4 bid orders",
			orders: []*Order{bidOrders[0], bidOrders[1], bidOrders[2], bidOrders[3]},
			expected: []*orderFulfillment{
				newOrderFulfillment(bidOrders[0]),
				newOrderFulfillment(bidOrders[1]),
				newOrderFulfillment(bidOrders[2]),
				newOrderFulfillment(bidOrders[3]),
			},
		},
		{
			name:   "1 bid 1 ask",
			orders: []*Order{askOrders[1], bidOrders[2]},
			expected: []*orderFulfillment{
				newOrderFulfillment(askOrders[1]),
				newOrderFulfillment(bidOrders[2]),
			},
		},
		{
			name:   "1 ask 1 bid",
			orders: []*Order{bidOrders[1], askOrders[2]},
			expected: []*orderFulfillment{
				newOrderFulfillment(bidOrders[1]),
				newOrderFulfillment(askOrders[2]),
			},
		},
		{
			name: "4 asks 4 bids",
			orders: []*Order{
				askOrders[0], askOrders[1], askOrders[2], askOrders[3],
				bidOrders[3], bidOrders[2], bidOrders[1], bidOrders[0],
			},
			expected: []*orderFulfillment{
				newOrderFulfillment(askOrders[0]),
				newOrderFulfillment(askOrders[1]),
				newOrderFulfillment(askOrders[2]),
				newOrderFulfillment(askOrders[3]),
				newOrderFulfillment(bidOrders[3]),
				newOrderFulfillment(bidOrders[2]),
				newOrderFulfillment(bidOrders[1]),
				newOrderFulfillment(bidOrders[0]),
			},
		},
		{
			name: "4 bids 4 asks",
			orders: []*Order{
				bidOrders[0], bidOrders[1], bidOrders[2], bidOrders[3],
				askOrders[3], askOrders[2], askOrders[1], askOrders[0],
			},
			expected: []*orderFulfillment{
				newOrderFulfillment(bidOrders[0]),
				newOrderFulfillment(bidOrders[1]),
				newOrderFulfillment(bidOrders[2]),
				newOrderFulfillment(bidOrders[3]),
				newOrderFulfillment(askOrders[3]),
				newOrderFulfillment(askOrders[2]),
				newOrderFulfillment(askOrders[1]),
				newOrderFulfillment(askOrders[0]),
			},
		},
		{
			name: "interweaved 4 asks 4 bids",
			orders: []*Order{
				bidOrders[3], askOrders[0], askOrders[3], bidOrders[1],
				bidOrders[0], askOrders[1], bidOrders[2], askOrders[2],
			},
			expected: []*orderFulfillment{
				newOrderFulfillment(bidOrders[3]),
				newOrderFulfillment(askOrders[0]),
				newOrderFulfillment(askOrders[3]),
				newOrderFulfillment(bidOrders[1]),
				newOrderFulfillment(bidOrders[0]),
				newOrderFulfillment(askOrders[1]),
				newOrderFulfillment(bidOrders[2]),
				newOrderFulfillment(askOrders[2]),
			},
		},
		{
			name: "duplicated entries",
			orders: []*Order{
				askOrders[3], bidOrders[2], askOrders[3], bidOrders[2],
			},
			expected: []*orderFulfillment{
				newOrderFulfillment(askOrders[3]),
				newOrderFulfillment(bidOrders[2]),
				newOrderFulfillment(askOrders[3]),
				newOrderFulfillment(bidOrders[2]),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []*orderFulfillment
			testFunc := func() {
				actual = newOrderFulfillments(tc.orders)
			}
			require.NotPanics(t, testFunc, "newOrderFulfillments")
			assertEqualOrderFulfillmentSlices(t, tc.expected, actual, "newOrderFulfillments result")
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
		receiver orderFulfillment
		amt      sdkmath.Int
		expected sdk.Coin
		expPanic string
	}{
		{
			name:     "nil order",
			receiver: orderFulfillment{Order: nil},
			amt:      sdkmath.NewInt(0),
			expPanic: "runtime error: invalid memory address or nil pointer dereference",
		},
		{
			name:     "nil inside order",
			receiver: orderFulfillment{Order: NewOrder(1)},
			amt:      sdkmath.NewInt(0),
			expPanic: nilSubTypeErr(1),
		},
		{
			name:     "unknown inside order",
			receiver: orderFulfillment{Order: newUnknownOrder(2)},
			amt:      sdkmath.NewInt(0),
			expPanic: unknownSubTypeErr(2),
		},
		{
			name:     "ask order",
			receiver: orderFulfillment{Order: askOrder(3, coin(4, "apple"), coin(5, "plum"))},
			amt:      sdkmath.NewInt(6),
			expected: coin(6, "apple"),
		},
		{
			name:     "ask order with negative assets",
			receiver: orderFulfillment{Order: askOrder(7, coin(-8, "apple"), coin(9, "plum"))},
			amt:      sdkmath.NewInt(10),
			expected: coin(10, "apple"),
		},
		{
			name:     "ask order, negative amt",
			receiver: orderFulfillment{Order: askOrder(11, coin(12, "apple"), coin(13, "plum"))},
			amt:      sdkmath.NewInt(-14),
			expected: coin(-14, "apple"),
		},
		{
			name:     "ask order with negative assets, negative amt",
			receiver: orderFulfillment{Order: askOrder(15, coin(-16, "apple"), coin(17, "plum"))},
			amt:      sdkmath.NewInt(-18),
			expected: coin(-18, "apple"),
		},
		{
			name:     "ask order, big amt",
			receiver: orderFulfillment{Order: askOrder(19, coin(20, "apple"), coin(21, "plum"))},
			amt:      newInt(t, "123,000,000,000,000,000,000,000,000,000,000,321"),
			expected: bigCoin("123,000,000,000,000,000,000,000,000,000,000,321", "apple"),
		},
		{
			name:     "bid order",
			receiver: orderFulfillment{Order: bidOrder(3, coin(4, "apple"), coin(5, "plum"))},
			amt:      sdkmath.NewInt(6),
			expected: coin(6, "apple"),
		},
		{
			name:     "bid order with negative assets",
			receiver: orderFulfillment{Order: bidOrder(7, coin(-8, "apple"), coin(9, "plum"))},
			amt:      sdkmath.NewInt(10),
			expected: coin(10, "apple"),
		},
		{
			name:     "bid order, negative amt",
			receiver: orderFulfillment{Order: bidOrder(11, coin(12, "apple"), coin(13, "plum"))},
			amt:      sdkmath.NewInt(-14),
			expected: coin(-14, "apple"),
		},
		{
			name:     "bid order with negative assets, negative amt",
			receiver: orderFulfillment{Order: bidOrder(15, coin(-16, "apple"), coin(17, "plum"))},
			amt:      sdkmath.NewInt(-18),
			expected: coin(-18, "apple"),
		},
		{
			name:     "bid order, big amt",
			receiver: orderFulfillment{Order: bidOrder(19, coin(20, "apple"), coin(21, "plum"))},
			amt:      newInt(t, "123,000,000,000,000,000,000,000,000,000,000,321"),
			expected: bigCoin("123,000,000,000,000,000,000,000,000,000,000,321", "apple"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coin
			testFunc := func() {
				actual = tc.receiver.AssetCoin(tc.amt)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "AssetCoin(%s)", tc.amt)
			assert.Equal(t, tc.expected.String(), actual.String(), "AssetCoin(%s) result", tc.amt)
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
		receiver orderFulfillment
		amt      sdkmath.Int
		expected sdk.Coin
		expPanic string
	}{
		{
			name:     "nil order",
			receiver: orderFulfillment{Order: nil},
			amt:      sdkmath.NewInt(0),
			expPanic: "runtime error: invalid memory address or nil pointer dereference",
		},
		{
			name:     "nil inside order",
			receiver: orderFulfillment{Order: NewOrder(1)},
			amt:      sdkmath.NewInt(0),
			expPanic: nilSubTypeErr(1),
		},
		{
			name:     "unknown inside order",
			receiver: orderFulfillment{Order: newUnknownOrder(2)},
			amt:      sdkmath.NewInt(0),
			expPanic: unknownSubTypeErr(2),
		},
		{
			name:     "ask order",
			receiver: orderFulfillment{Order: askOrder(3, coin(4, "apple"), coin(5, "plum"))},
			amt:      sdkmath.NewInt(6),
			expected: coin(6, "plum"),
		},
		{
			name:     "ask order with negative assets",
			receiver: orderFulfillment{Order: askOrder(7, coin(-8, "apple"), coin(9, "plum"))},
			amt:      sdkmath.NewInt(10),
			expected: coin(10, "plum"),
		},
		{
			name:     "ask order, negative amt",
			receiver: orderFulfillment{Order: askOrder(11, coin(12, "apple"), coin(13, "plum"))},
			amt:      sdkmath.NewInt(-14),
			expected: coin(-14, "plum"),
		},
		{
			name:     "ask order with negative assets, negative amt",
			receiver: orderFulfillment{Order: askOrder(15, coin(-16, "apple"), coin(17, "plum"))},
			amt:      sdkmath.NewInt(-18),
			expected: coin(-18, "plum"),
		},
		{
			name:     "ask order, big amt",
			receiver: orderFulfillment{Order: askOrder(19, coin(20, "apple"), coin(21, "plum"))},
			amt:      newInt(t, "123,000,000,000,000,000,000,000,000,000,000,321"),
			expected: bigCoin("123,000,000,000,000,000,000,000,000,000,000,321", "plum"),
		},
		{
			name:     "bid order",
			receiver: orderFulfillment{Order: bidOrder(3, coin(4, "apple"), coin(5, "plum"))},
			amt:      sdkmath.NewInt(6),
			expected: coin(6, "plum"),
		},
		{
			name:     "bid order with negative assets",
			receiver: orderFulfillment{Order: bidOrder(7, coin(-8, "apple"), coin(9, "plum"))},
			amt:      sdkmath.NewInt(10),
			expected: coin(10, "plum"),
		},
		{
			name:     "bid order, negative amt",
			receiver: orderFulfillment{Order: bidOrder(11, coin(12, "apple"), coin(13, "plum"))},
			amt:      sdkmath.NewInt(-14),
			expected: coin(-14, "plum"),
		},
		{
			name:     "bid order with negative assets, negative amt",
			receiver: orderFulfillment{Order: bidOrder(15, coin(-16, "apple"), coin(17, "plum"))},
			amt:      sdkmath.NewInt(-18),
			expected: coin(-18, "plum"),
		},
		{
			name:     "bid order, big amt",
			receiver: orderFulfillment{Order: bidOrder(19, coin(20, "apple"), coin(21, "plum"))},
			amt:      newInt(t, "123,000,000,000,000,000,000,000,000,000,000,321"),
			expected: bigCoin("123,000,000,000,000,000,000,000,000,000,000,321", "plum"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coin
			testFunc := func() {
				actual = tc.receiver.PriceCoin(tc.amt)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "PriceCoin(%s)", tc.amt)
			assert.Equal(t, tc.expected.String(), actual.String(), "PriceCoin(%s) result", tc.amt)
		})
	}
}

func TestOrderFulfillment_GetAssetsFilled(t *testing.T) {
	coin := func(amt int64) sdk.Coin {
		return sdk.Coin{Denom: "asset", Amount: sdkmath.NewInt(amt)}
	}
	askOrder := NewOrder(444).WithAsk(&AskOrder{Assets: coin(5555)})
	bidOrder := NewOrder(666).WithBid(&BidOrder{Assets: coin(7777)})

	newOF := func(order *Order, amt int64) orderFulfillment {
		return orderFulfillment{
			Order:           order,
			AssetsFilledAmt: sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name string
		f    orderFulfillment
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

	newOF := func(order *Order, amt int64) orderFulfillment {
		return orderFulfillment{
			Order:             order,
			AssetsUnfilledAmt: sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name string
		f    orderFulfillment
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

	newOF := func(order *Order, amt int64) orderFulfillment {
		return orderFulfillment{
			Order:           order,
			PriceAppliedAmt: sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name string
		f    orderFulfillment
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

	newOF := func(order *Order, amt int64) orderFulfillment {
		return orderFulfillment{
			Order:        order,
			PriceLeftAmt: sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name string
		f    orderFulfillment
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

func TestOrderFulfillment_GetOrderID(t *testing.T) {
	newOF := func(orderID uint64) orderFulfillment {
		return orderFulfillment{
			Order: NewOrder(orderID),
		}
	}

	tests := []struct {
		name string
		f    orderFulfillment
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
		f    orderFulfillment
		exp  bool
	}{
		{name: "ask", f: orderFulfillment{Order: NewOrder(444).WithAsk(&AskOrder{})}, exp: true},
		{name: "bid", f: orderFulfillment{Order: NewOrder(666).WithBid(&BidOrder{})}, exp: false},
		{name: "nil", f: orderFulfillment{Order: NewOrder(888)}, exp: false},
		{name: "unknown", f: orderFulfillment{Order: newUnknownOrder(7)}, exp: false},
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
		f    orderFulfillment
		exp  bool
	}{
		{name: "ask", f: orderFulfillment{Order: NewOrder(444).WithAsk(&AskOrder{})}, exp: false},
		{name: "bid", f: orderFulfillment{Order: NewOrder(666).WithBid(&BidOrder{})}, exp: true},
		{name: "nil", f: orderFulfillment{Order: NewOrder(888)}, exp: false},
		{name: "unknown", f: orderFulfillment{Order: newUnknownOrder(9)}, exp: false},
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
	askOrder := func(marketID uint32) orderFulfillment {
		return orderFulfillment{
			Order: NewOrder(999).WithAsk(&AskOrder{MarketId: marketID}),
		}
	}
	bidOrder := func(marketID uint32) orderFulfillment {
		return orderFulfillment{
			Order: NewOrder(999).WithBid(&BidOrder{MarketId: marketID}),
		}
	}

	tests := []struct {
		name string
		f    orderFulfillment
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
	askOrder := func(seller string) orderFulfillment {
		return orderFulfillment{
			Order: NewOrder(999).WithAsk(&AskOrder{Seller: seller}),
		}
	}
	bidOrder := func(buyer string) orderFulfillment {
		return orderFulfillment{
			Order: NewOrder(999).WithBid(&BidOrder{Buyer: buyer}),
		}
	}
	owner := sdk.AccAddress("owner_______________").String()

	tests := []struct {
		name string
		f    orderFulfillment
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
	askOrder := func(amt int64) orderFulfillment {
		return orderFulfillment{
			Order: NewOrder(999).WithAsk(&AskOrder{Assets: coin(amt)}),
		}
	}
	bidOrder := func(amt int64) orderFulfillment {
		return orderFulfillment{
			Order: NewOrder(999).WithBid(&BidOrder{Assets: coin(amt)}),
		}
	}

	tests := []struct {
		name string
		f    orderFulfillment
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
	askOrder := func(amt int64) orderFulfillment {
		return orderFulfillment{
			Order: NewOrder(999).WithAsk(&AskOrder{Price: coin(amt)}),
		}
	}
	bidOrder := func(amt int64) orderFulfillment {
		return orderFulfillment{
			Order: NewOrder(999).WithBid(&BidOrder{Price: coin(amt)}),
		}
	}

	tests := []struct {
		name string
		f    orderFulfillment
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
	askOrder := func(coin *sdk.Coin) orderFulfillment {
		return orderFulfillment{
			Order: NewOrder(999).WithAsk(&AskOrder{SellerSettlementFlatFee: coin}),
		}
	}
	bidOrder := func(coins sdk.Coins) orderFulfillment {
		return orderFulfillment{
			Order: NewOrder(999).WithBid(&BidOrder{BuyerSettlementFees: coins}),
		}
	}

	tests := []struct {
		name string
		f    orderFulfillment
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
	askOrder := func(allowPartial bool) orderFulfillment {
		return orderFulfillment{
			Order: NewOrder(999).WithAsk(&AskOrder{AllowPartial: allowPartial}),
		}
	}
	bidOrder := func(allowPartial bool) orderFulfillment {
		return orderFulfillment{
			Order: NewOrder(999).WithBid(&BidOrder{AllowPartial: allowPartial}),
		}
	}

	tests := []struct {
		name string
		f    orderFulfillment
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

func TestOrderFulfillment_GetExternalID(t *testing.T) {
	askOrder := func(externalID string) orderFulfillment {
		return orderFulfillment{
			Order: NewOrder(999).WithAsk(&AskOrder{ExternalId: externalID}),
		}
	}
	bidOrder := func(externalID string) orderFulfillment {
		return orderFulfillment{
			Order: NewOrder(999).WithBid(&BidOrder{ExternalId: externalID}),
		}
	}

	tests := []struct {
		name string
		f    orderFulfillment
		exp  string
	}{
		{name: "ask empty", f: askOrder(""), exp: ""},
		{name: "ask something", f: askOrder("something"), exp: "something"},
		{name: "ask uuid", f: askOrder("5E50E61A-43E7-4C35-86BF-04040B49C1CB"), exp: "5E50E61A-43E7-4C35-86BF-04040B49C1CB"},
		{name: "bid empty", f: bidOrder(""), exp: ""},
		{name: "bid something", f: bidOrder("SOMETHING"), exp: "SOMETHING"},
		{name: "bid uuid", f: bidOrder("1AD88BB9-86FA-42F1-A824-66AB27547904"), exp: "1AD88BB9-86FA-42F1-A824-66AB27547904"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = tc.f.GetExternalID()
			}
			require.NotPanics(t, testFunc, "GetExternalID()")
			assert.Equal(t, tc.exp, actual, "GetExternalID() result")
		})
	}
}

func TestOrderFulfillment_GetOrderType(t *testing.T) {
	tests := []struct {
		name string
		f    orderFulfillment
		exp  string
	}{
		{name: "ask", f: orderFulfillment{Order: NewOrder(444).WithAsk(&AskOrder{})}, exp: OrderTypeAsk},
		{name: "bid", f: orderFulfillment{Order: NewOrder(666).WithBid(&BidOrder{})}, exp: OrderTypeBid},
		{name: "nil", f: orderFulfillment{Order: NewOrder(888)}, exp: "<nil>"},
		{name: "unknown", f: orderFulfillment{Order: newUnknownOrder(8)}, exp: "*exchange.unknownOrderType"},
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
		f    orderFulfillment
		exp  byte
	}{
		{name: "ask", f: orderFulfillment{Order: NewOrder(444).WithAsk(&AskOrder{})}, exp: OrderTypeByteAsk},
		{name: "bid", f: orderFulfillment{Order: NewOrder(666).WithBid(&BidOrder{})}, exp: OrderTypeByteBid},
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
		f    orderFulfillment
	}{
		{
			name: "ask",
			f: orderFulfillment{
				Order: NewOrder(111).WithAsk(&AskOrder{
					Assets:                  sdk.Coin{Denom: "asset", Amount: sdkmath.NewInt(55)},
					SellerSettlementFlatFee: &sdk.Coin{Denom: "fee", Amount: sdkmath.NewInt(3)},
				}),
			},
		},
		{
			name: "bid",
			f: orderFulfillment{
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
	newOF := func(order *Order, assetsUnfilled, assetsFilled int64, dists ...*distribution) *orderFulfillment {
		rv := &orderFulfillment{
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
	askOF := func(orderID uint64, assetsUnfilled, assetsFilled int64, dists ...*distribution) *orderFulfillment {
		order := NewOrder(orderID).WithAsk(&AskOrder{Assets: sdk.NewInt64Coin("apple", 999)})
		return newOF(order, assetsUnfilled, assetsFilled, dists...)
	}
	bidOF := func(orderID uint64, assetsUnfilled, assetsFilled int64, dists ...*distribution) *orderFulfillment {
		order := NewOrder(orderID).WithBid(&BidOrder{Assets: sdk.NewInt64Coin("apple", 999)})
		return newOF(order, assetsUnfilled, assetsFilled, dists...)
	}
	dist := func(addr string, amt int64) *distribution {
		return &distribution{
			Address: addr,
			Amount:  sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name     string
		receiver *orderFulfillment
		order    OrderI
		amount   sdkmath.Int
		expRes   *orderFulfillment
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
			require.NotPanics(t, testFunc, "distributeAssets")
			assertions.AssertErrorValue(t, err, tc.expErr, "distributeAssets error")
			if !assertEqualOrderFulfillments(t, tc.expRes, tc.receiver, "orderFulfillment after distributeAssets") {
				t.Logf("Original: %s", orderFulfillmentString(orig))
				t.Logf("  Amount: %s", tc.amount)
			}
		})
	}
}

func TestDistributeAssets(t *testing.T) {
	seller, buyer := "SelleR", "BuyeR"
	newOF := func(order *Order, assetsUnfilled, assetsFilled int64, dists ...*distribution) *orderFulfillment {
		rv := &orderFulfillment{
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
	askOF := func(orderID uint64, assetsUnfilled, assetsFilled int64, dists ...*distribution) *orderFulfillment {
		order := NewOrder(orderID).WithAsk(&AskOrder{
			Seller: seller,
			Assets: sdk.NewInt64Coin("apple", 999),
		})
		return newOF(order, assetsUnfilled, assetsFilled, dists...)
	}
	bidOF := func(orderID uint64, assetsUnfilled, assetsFilled int64, dists ...*distribution) *orderFulfillment {
		order := NewOrder(orderID).WithBid(&BidOrder{
			Buyer:  buyer,
			Assets: sdk.NewInt64Coin("apple", 999),
		})
		return newOF(order, assetsUnfilled, assetsFilled, dists...)
	}
	dist := func(addr string, amt int64) *distribution {
		return &distribution{
			Address: addr,
			Amount:  sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name   string
		of1    *orderFulfillment
		of2    *orderFulfillment
		amount sdkmath.Int
		expOF1 *orderFulfillment
		expOF2 *orderFulfillment
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
				err = distributeAssets(tc.of1, tc.of2, tc.amount)
			}
			require.NotPanics(t, testFunc, "distributeAssets")
			assertions.AssertErrorValue(t, err, tc.expErr, "distributeAssets error")
			if !assertEqualOrderFulfillments(t, tc.expOF1, tc.of1, "of1 after distributeAssets") {
				t.Logf("Original: %s", orderFulfillmentString(origOF1))
				t.Logf("  Amount: %s", tc.amount)
			}
			if !assertEqualOrderFulfillments(t, tc.expOF2, tc.of2, "of2 after distributeAssets") {
				t.Logf("Original: %s", orderFulfillmentString(origOF2))
				t.Logf("  Amount: %s", tc.amount)
			}
		})
	}
}

func TestOrderFulfillment_DistributePrice(t *testing.T) {
	newOF := func(order *Order, priceLeft, priceApplied int64, dists ...*distribution) *orderFulfillment {
		rv := &orderFulfillment{
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
	askOF := func(orderID uint64, priceLeft, priceApplied int64, dists ...*distribution) *orderFulfillment {
		order := NewOrder(orderID).WithAsk(&AskOrder{Price: sdk.NewInt64Coin("peach", 999)})
		return newOF(order, priceLeft, priceApplied, dists...)
	}
	bidOF := func(orderID uint64, priceLeft, priceApplied int64, dists ...*distribution) *orderFulfillment {
		order := NewOrder(orderID).WithBid(&BidOrder{Price: sdk.NewInt64Coin("peach", 999)})
		return newOF(order, priceLeft, priceApplied, dists...)
	}
	dist := func(addr string, amt int64) *distribution {
		return &distribution{
			Address: addr,
			Amount:  sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name     string
		receiver *orderFulfillment
		order    OrderI
		amount   sdkmath.Int
		expRes   *orderFulfillment
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
			require.NotPanics(t, testFunc, "distributePrice")
			assertions.AssertErrorValue(t, err, tc.expErr, "distributePrice error")
			if !assertEqualOrderFulfillments(t, tc.expRes, tc.receiver, "orderFulfillment after distributePrice") {
				t.Logf("Original: %s", orderFulfillmentString(orig))
				t.Logf("  Amount: %s", tc.amount)
			}
		})
	}
}

func TestDistributePrice(t *testing.T) {
	seller, buyer := "SelleR", "BuyeR"
	newOF := func(order *Order, priceLeft, priceApplied int64, dists ...*distribution) *orderFulfillment {
		rv := &orderFulfillment{
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
	askOF := func(orderID uint64, priceLeft, priceApplied int64, dists ...*distribution) *orderFulfillment {
		order := NewOrder(orderID).WithAsk(&AskOrder{
			Seller: seller,
			Price:  sdk.NewInt64Coin("peach", 999),
		})
		return newOF(order, priceLeft, priceApplied, dists...)
	}
	bidOF := func(orderID uint64, priceLeft, priceApplied int64, dists ...*distribution) *orderFulfillment {
		order := NewOrder(orderID).WithBid(&BidOrder{
			Buyer: buyer,
			Price: sdk.NewInt64Coin("peach", 999),
		})
		return newOF(order, priceLeft, priceApplied, dists...)
	}
	dist := func(addr string, amt int64) *distribution {
		return &distribution{
			Address: addr,
			Amount:  sdkmath.NewInt(amt),
		}
	}

	tests := []struct {
		name   string
		of1    *orderFulfillment
		of2    *orderFulfillment
		amount sdkmath.Int
		expOF1 *orderFulfillment
		expOF2 *orderFulfillment
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
				err = distributePrice(tc.of1, tc.of2, tc.amount)
			}
			require.NotPanics(t, testFunc, "distributePrice")
			assertions.AssertErrorValue(t, err, tc.expErr, "distributePrice error")
			if !assertEqualOrderFulfillments(t, tc.expOF1, tc.of1, "of1 after distributePrice") {
				t.Logf("Original: %s", orderFulfillmentString(origOF1))
				t.Logf("  Amount: %s", tc.amount)
			}
			if !assertEqualOrderFulfillments(t, tc.expOF2, tc.of2, "of2 after distributePrice") {
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
		receiver    *orderFulfillment
		expUnfilled *Order
		expReceiver *orderFulfillment
		expErr      string
	}{
		{
			name: "order split error: ask",
			receiver: &orderFulfillment{
				Order:           askOrder(8, 10, 100),
				AssetsFilledAmt: sdkmath.NewInt(-1),
			},
			expErr: "cannot split ask order 8 having asset \"10apple\" at \"-1apple\": amount filled not positive",
		},
		{
			name: "order split error: bid",
			receiver: &orderFulfillment{
				Order:           bidOrder(9, 10, 100),
				AssetsFilledAmt: sdkmath.NewInt(-1),
			},
			expErr: "cannot split bid order 9 having asset \"10apple\" at \"-1apple\": amount filled not positive",
		},
		{
			name: "okay: ask",
			receiver: &orderFulfillment{
				Order:             askOrder(17, 10, 100, coin(20, "fig")),
				AssetsFilledAmt:   sdkmath.NewInt(9),
				AssetsUnfilledAmt: sdkmath.NewInt(1),
				PriceAppliedAmt:   sdkmath.NewInt(300),
				PriceLeftAmt:      sdkmath.NewInt(-200),
			},
			expUnfilled: askOrder(17, 1, 10, coin(2, "fig")),
			expReceiver: &orderFulfillment{
				Order:             askOrder(17, 9, 90, coin(18, "fig")),
				AssetsFilledAmt:   sdkmath.NewInt(9),
				AssetsUnfilledAmt: sdkmath.NewInt(0),
				PriceAppliedAmt:   sdkmath.NewInt(300),
				PriceLeftAmt:      sdkmath.NewInt(-210),
			},
		},
		{
			name: "okay: bid",
			receiver: &orderFulfillment{
				Order:             bidOrder(19, 10, 100, coin(20, "fig")),
				AssetsFilledAmt:   sdkmath.NewInt(9),
				AssetsUnfilledAmt: sdkmath.NewInt(1),
				PriceAppliedAmt:   sdkmath.NewInt(300),
				PriceLeftAmt:      sdkmath.NewInt(-200),
			},
			expUnfilled: bidOrder(19, 1, 10, coin(2, "fig")),
			expReceiver: &orderFulfillment{
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
			if !assertEqualOrderFulfillments(t, tc.expReceiver, tc.receiver, "orderFulfillment after SplitOrder") {
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
		receiver orderFulfillment
		expected *FilledOrder
	}{
		{
			name: "ask order",
			receiver: orderFulfillment{
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
			receiver: orderFulfillment{
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
		fulfillments []*orderFulfillment
		expected     sdkmath.Int
	}{
		{
			name:         "nil fulfillments",
			fulfillments: nil,
			expected:     sdkmath.NewInt(0),
		},
		{
			name:         "empty fulfillments",
			fulfillments: []*orderFulfillment{},
			expected:     sdkmath.NewInt(0),
		},
		{
			name:         "one fulfillment, positive",
			fulfillments: []*orderFulfillment{{PriceLeftAmt: sdkmath.NewInt(8)}},
			expected:     sdkmath.NewInt(8),
		},
		{
			name:         "one fulfillment, zero",
			fulfillments: []*orderFulfillment{{PriceLeftAmt: sdkmath.NewInt(0)}},
			expected:     sdkmath.NewInt(0),
		},
		{
			name:         "one fulfillment, negative",
			fulfillments: []*orderFulfillment{{PriceLeftAmt: sdkmath.NewInt(-3)}},
			expected:     sdkmath.NewInt(-3),
		},
		{
			name: "three fulfillments",
			fulfillments: []*orderFulfillment{
				{PriceLeftAmt: sdkmath.NewInt(10)},
				{PriceLeftAmt: sdkmath.NewInt(200)},
				{PriceLeftAmt: sdkmath.NewInt(3000)},
			},
			expected: sdkmath.NewInt(3210),
		},
		{
			name: "three fulfillments, one negative",
			fulfillments: []*orderFulfillment{
				{PriceLeftAmt: sdkmath.NewInt(10)},
				{PriceLeftAmt: sdkmath.NewInt(-200)},
				{PriceLeftAmt: sdkmath.NewInt(3000)},
			},
			expected: sdkmath.NewInt(2810),
		},
		{
			name: "three fulfillments, all negative",
			fulfillments: []*orderFulfillment{
				{PriceLeftAmt: sdkmath.NewInt(-10)},
				{PriceLeftAmt: sdkmath.NewInt(-200)},
				{PriceLeftAmt: sdkmath.NewInt(-3000)},
			},
			expected: sdkmath.NewInt(-3210),
		},
		{
			name: "three fulfillments, all large",
			fulfillments: []*orderFulfillment{
				{PriceLeftAmt: newInt(t, "3,000,000,000,000,000,000,000,000,000,000,300")},
				{PriceLeftAmt: newInt(t, "40,000,000,000,000,000,000,000,000,000,000,040")},
				{PriceLeftAmt: newInt(t, "500,000,000,000,000,000,000,000,000,000,000,005")},
			},
			expected: newInt(t, "543,000,000,000,000,000,000,000,000,000,000,345"),
		},
		{
			name: "four fullfillments, small negative zero large",
			fulfillments: []*orderFulfillment{
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
	newOF := func(order *Order, dists ...*distribution) *orderFulfillment {
		rv := &orderFulfillment{
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
	dist := func(addr string, amount int64) *distribution {
		return &distribution{Address: addr, Amount: sdkmath.NewInt(amount)}
	}

	tests := []struct {
		name      string
		askOFs    []*orderFulfillment
		bidOFs    []*orderFulfillment
		expAskOFs []*orderFulfillment
		expBidOfs []*orderFulfillment
		expErr    string
	}{
		{
			name:      "one ask, one bid: both full",
			askOFs:    []*orderFulfillment{newOF(askOrder(5, 10, "seller"))},
			bidOFs:    []*orderFulfillment{newOF(bidOrder(6, 10, "buyer"))},
			expAskOFs: []*orderFulfillment{newOF(askOrder(5, 10, "seller"), dist("buyer", 10))},
			expBidOfs: []*orderFulfillment{newOF(bidOrder(6, 10, "buyer"), dist("seller", 10))},
		},
		{
			name:      "one ask, one bid: ask partial",
			askOFs:    []*orderFulfillment{newOF(askOrder(5, 11, "seller"))},
			bidOFs:    []*orderFulfillment{newOF(bidOrder(16, 10, "buyer"))},
			expAskOFs: []*orderFulfillment{newOF(askOrder(5, 11, "seller"), dist("buyer", 10))},
			expBidOfs: []*orderFulfillment{newOF(bidOrder(16, 10, "buyer"), dist("seller", 10))},
		},
		{
			name:      "one ask, one bid: bid partial",
			askOFs:    []*orderFulfillment{newOF(askOrder(15, 10, "seller"))},
			bidOFs:    []*orderFulfillment{newOF(bidOrder(6, 11, "buyer"))},
			expAskOFs: []*orderFulfillment{newOF(askOrder(15, 10, "seller"), dist("buyer", 10))},
			expBidOfs: []*orderFulfillment{newOF(bidOrder(6, 11, "buyer"), dist("seller", 10))},
		},
		{
			name:   "one ask, two bids: last bid not touched",
			askOFs: []*orderFulfillment{newOF(askOrder(22, 10, "seller"))},
			bidOFs: []*orderFulfillment{
				newOF(bidOrder(64, 12, "buyer64")),
				newOF(bidOrder(78, 1, "buyer78")),
			},
			expAskOFs: []*orderFulfillment{newOF(askOrder(22, 10, "seller"), dist("buyer64", 10))},
			expBidOfs: []*orderFulfillment{
				newOF(bidOrder(64, 12, "buyer64"), dist("seller", 10)),
				newOF(bidOrder(78, 1, "buyer78")),
			},
		},
		{
			name: "two asks, one bids: last ask not touched",
			askOFs: []*orderFulfillment{
				newOF(askOrder(888, 10, "seller888")),
				newOF(askOrder(999, 10, "seller999")),
			},
			bidOFs: []*orderFulfillment{newOF(bidOrder(6, 10, "buyer"))},
			expAskOFs: []*orderFulfillment{
				newOF(askOrder(888, 10, "seller888"), dist("buyer", 10)),
				newOF(askOrder(999, 10, "seller999")),
			},
			expBidOfs: []*orderFulfillment{newOF(bidOrder(6, 10, "buyer"), dist("seller888", 10))},
		},
		{
			name: "two asks, three bids: both full",
			askOFs: []*orderFulfillment{
				newOF(askOrder(101, 15, "seller101")),
				newOF(askOrder(102, 25, "seller102")),
			},
			bidOFs: []*orderFulfillment{
				newOF(bidOrder(103, 10, "buyer103")),
				newOF(bidOrder(104, 8, "buyer104")),
				newOF(bidOrder(105, 22, "buyer105")),
			},
			expAskOFs: []*orderFulfillment{
				newOF(askOrder(101, 15, "seller101"), dist("buyer103", 10), dist("buyer104", 5)),
				newOF(askOrder(102, 25, "seller102"), dist("buyer104", 3), dist("buyer105", 22)),
			},
			expBidOfs: []*orderFulfillment{
				newOF(bidOrder(103, 10, "buyer103"), dist("seller101", 10)),
				newOF(bidOrder(104, 8, "buyer104"), dist("seller101", 5), dist("seller102", 3)),
				newOF(bidOrder(105, 22, "buyer105"), dist("seller102", 22)),
			},
		},
		{
			name: "two asks, three bids: ask partial",
			askOFs: []*orderFulfillment{
				newOF(askOrder(101, 15, "seller101")),
				newOF(askOrder(102, 26, "seller102")),
			},
			bidOFs: []*orderFulfillment{
				newOF(bidOrder(103, 10, "buyer103")),
				newOF(bidOrder(104, 8, "buyer104")),
				newOF(bidOrder(105, 22, "buyer105")),
			},
			expAskOFs: []*orderFulfillment{
				newOF(askOrder(101, 15, "seller101"), dist("buyer103", 10), dist("buyer104", 5)),
				newOF(askOrder(102, 26, "seller102"), dist("buyer104", 3), dist("buyer105", 22)),
			},
			expBidOfs: []*orderFulfillment{
				newOF(bidOrder(103, 10, "buyer103"), dist("seller101", 10)),
				newOF(bidOrder(104, 8, "buyer104"), dist("seller101", 5), dist("seller102", 3)),
				newOF(bidOrder(105, 22, "buyer105"), dist("seller102", 22)),
			},
		},
		{
			name: "two asks, three bids: bid partial",
			askOFs: []*orderFulfillment{
				newOF(askOrder(101, 15, "seller101")),
				newOF(askOrder(102, 25, "seller102")),
			},
			bidOFs: []*orderFulfillment{
				newOF(bidOrder(103, 10, "buyer103")),
				newOF(bidOrder(104, 8, "buyer104")),
				newOF(bidOrder(105, 23, "buyer105")),
			},
			expAskOFs: []*orderFulfillment{
				newOF(askOrder(101, 15, "seller101"), dist("buyer103", 10), dist("buyer104", 5)),
				newOF(askOrder(102, 25, "seller102"), dist("buyer104", 3), dist("buyer105", 22)),
			},
			expBidOfs: []*orderFulfillment{
				newOF(bidOrder(103, 10, "buyer103"), dist("seller101", 10)),
				newOF(bidOrder(104, 8, "buyer104"), dist("seller101", 5), dist("seller102", 3)),
				newOF(bidOrder(105, 23, "buyer105"), dist("seller102", 22)),
			},
		},
		{
			name:   "negative ask assets unfilled",
			askOFs: []*orderFulfillment{newOF(askOrder(101, 10, "seller"), dist("buyerx", 11))},
			bidOFs: []*orderFulfillment{newOF(bidOrder(102, 10, "buyer"))},
			expErr: "cannot fill ask order 101 having assets left \"-1apple\" with bid order 102 having " +
				"assets left \"10apple\": zero or negative assets left",
		},
		{
			name:   "negative bid assets unfilled",
			askOFs: []*orderFulfillment{newOF(askOrder(101, 10, "seller"))},
			bidOFs: []*orderFulfillment{newOF(bidOrder(102, 10, "buyer"), dist("sellerx", 11))},
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

func TestGetFulfillmentAssetsAmt(t *testing.T) {
	newAskOF := func(orderID uint64, assetsUnfilled int64, assetDenom string) *orderFulfillment {
		return &orderFulfillment{
			Order: NewOrder(orderID).WithAsk(&AskOrder{
				Assets: sdk.NewInt64Coin(assetDenom, 999),
			}),
			AssetsUnfilledAmt: sdkmath.NewInt(assetsUnfilled),
		}
	}
	newBidOF := func(orderID uint64, assetsUnfilled int64, assetDenom string) *orderFulfillment {
		return &orderFulfillment{
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
		of1    *orderFulfillment
		of2    *orderFulfillment
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
				amt, err = getFulfillmentAssetsAmt(tc.of1, tc.of2)
			}
			require.NotPanics(t, testFunc, "getFulfillmentAssetsAmt")
			assertions.AssertErrorValue(t, err, tc.expErr, "getFulfillmentAssetsAmt error")
			assert.Equal(t, tc.expAmt, amt, "getFulfillmentAssetsAmt amount")
			assertEqualOrderFulfillments(t, origOF1, tc.of1, "of1 after getFulfillmentAssetsAmt")
			assertEqualOrderFulfillments(t, origOF2, tc.of2, "of2 after getFulfillmentAssetsAmt")
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
	newOF := func(order *Order, dists ...*distribution) *orderFulfillment {
		rv := &orderFulfillment{
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
	dist := func(addr string, amount int64) *distribution {
		return &distribution{Address: addr, Amount: sdkmath.NewInt(amount)}
	}

	tests := []struct {
		name          string
		askOFs        []*orderFulfillment
		bidOFs        []*orderFulfillment
		settlement    *Settlement
		expAskOFs     []*orderFulfillment
		expBidOfs     []*orderFulfillment
		expSettlement *Settlement
		expErr        string
	}{
		{
			name:       "one ask: not touched",
			askOFs:     []*orderFulfillment{newOF(askOrder(8, 10, "seller8"))},
			settlement: &Settlement{},
			expErr:     "ask order 8 (at index 0) has no assets filled",
		},
		{
			name:          "one ask: partial",
			askOFs:        []*orderFulfillment{newOF(askOrder(8, 10, "seller8"), dist("buyer", 7))},
			settlement:    &Settlement{},
			expAskOFs:     []*orderFulfillment{newOF(askOrder(8, 7, "seller8"), dist("buyer", 7))},
			expSettlement: &Settlement{PartialOrderLeft: askOrder(8, 3, "seller8")},
		},
		{
			name:       "one ask: partial, settlement already has a partial",
			askOFs:     []*orderFulfillment{newOF(askOrder(8, 10, "seller8"), dist("buyer", 7))},
			settlement: &Settlement{PartialOrderLeft: bidOrder(55, 3, "buyer")},
			expErr:     "bid order 55 and ask order 8 cannot both be partially filled",
		},
		{
			name: "one ask: partial, not allowed",
			askOFs: []*orderFulfillment{
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
			askOFs: []*orderFulfillment{
				newOF(askOrder(8, 10, "seller8"), dist("buyer", 7)),
				newOF(askOrder(9, 12, "seller8")),
			},
			settlement: &Settlement{},
			expErr:     "ask order 8 (at index 0) is not filled in full and is not the last ask order provided",
		},
		{
			name: "two asks: last untouched",
			askOFs: []*orderFulfillment{
				newOF(askOrder(8, 10, "seller8"), dist("buyer", 10)),
				newOF(askOrder(9, 12, "seller8")),
			},
			settlement: &Settlement{},
			expErr:     "ask order 9 (at index 1) has no assets filled",
		},
		{
			name: "two asks: last partial",
			askOFs: []*orderFulfillment{
				newOF(askOrder(8, 10, "seller8"), dist("buyer", 10)),
				newOF(askOrder(9, 12, "seller9"), dist("buyer", 10)),
			},
			settlement: &Settlement{},
			expAskOFs: []*orderFulfillment{
				newOF(askOrder(8, 10, "seller8"), dist("buyer", 10)),
				newOF(askOrder(9, 10, "seller9"), dist("buyer", 10)),
			},
			expSettlement: &Settlement{PartialOrderLeft: askOrder(9, 2, "seller9")},
		},

		{
			name:       "one bid: not touched",
			bidOFs:     []*orderFulfillment{newOF(bidOrder(8, 10, "buyer8"))},
			settlement: &Settlement{},
			expErr:     "bid order 8 (at index 0) has no assets filled",
		},
		{
			name:          "one bid: partial",
			bidOFs:        []*orderFulfillment{newOF(bidOrder(8, 10, "buyer8"), dist("seller", 7))},
			settlement:    &Settlement{},
			expBidOfs:     []*orderFulfillment{newOF(bidOrder(8, 7, "buyer8"), dist("seller", 7))},
			expSettlement: &Settlement{PartialOrderLeft: bidOrder(8, 3, "buyer8")},
		},
		{
			name: "one bid: partial, not allowed",
			askOFs: []*orderFulfillment{
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
			bidOFs: []*orderFulfillment{
				newOF(bidOrder(8, 10, "buyer8"), dist("seller", 7)),
				newOF(bidOrder(9, 12, "buyer9")),
			},
			settlement: &Settlement{},
			expErr:     "bid order 8 (at index 0) is not filled in full and is not the last bid order provided",
		},
		{
			name: "two bids: last untouched",
			bidOFs: []*orderFulfillment{
				newOF(bidOrder(8, 10, "buyer8"), dist("seller", 10)),
				newOF(bidOrder(9, 12, "buyer9")),
			},
			settlement: &Settlement{},
			expErr:     "bid order 9 (at index 1) has no assets filled",
		},
		{
			name: "two bids: last partial",
			bidOFs: []*orderFulfillment{
				newOF(bidOrder(8, 10, "buyer8"), dist("seller", 10)),
				newOF(bidOrder(9, 12, "buyer9"), dist("seller", 10)),
			},
			settlement: &Settlement{},
			expBidOfs: []*orderFulfillment{
				newOF(bidOrder(8, 10, "buyer8"), dist("seller", 10)),
				newOF(bidOrder(9, 10, "buyer9"), dist("seller", 10)),
			},
			expSettlement: &Settlement{PartialOrderLeft: bidOrder(9, 2, "buyer9")},
		},
		{
			name:       "one ask, one bid: both partial",
			askOFs:     []*orderFulfillment{newOF(askOrder(8, 10, "seller8"), dist("buyer9", 7))},
			bidOFs:     []*orderFulfillment{newOF(bidOrder(9, 10, "buyer9"), dist("seller8", 7))},
			settlement: &Settlement{},
			expErr:     "ask order 8 and bid order 9 cannot both be partially filled",
		},
		{
			name: "three asks, three bids: no partial",
			askOFs: []*orderFulfillment{
				newOF(askOrder(51, 10, "seller51"), dist("buyer99", 10)),
				newOF(askOrder(77, 15, "seller77"), dist("buyer99", 10), dist("buyer8", 5)),
				newOF(askOrder(12, 20, "seller12"), dist("buyer8", 10), dist("buyer112", 10)),
			},
			bidOFs: []*orderFulfillment{
				newOF(bidOrder(99, 20, "buyer99"), dist("seller51", 10), dist("seller77", 10)),
				newOF(bidOrder(8, 15, "buyer8"), dist("seller77", 5), dist("seller12", 10)),
				newOF(bidOrder(112, 10, "buyer112"), dist("seller12", 10)),
			},
			settlement: &Settlement{},
			expAskOFs: []*orderFulfillment{
				newOF(askOrder(51, 10, "seller51"), dist("buyer99", 10)),
				newOF(askOrder(77, 15, "seller77"), dist("buyer99", 10), dist("buyer8", 5)),
				newOF(askOrder(12, 20, "seller12"), dist("buyer8", 10), dist("buyer112", 10)),
			},
			expBidOfs: []*orderFulfillment{
				newOF(bidOrder(99, 20, "buyer99"), dist("seller51", 10), dist("seller77", 10)),
				newOF(bidOrder(8, 15, "buyer8"), dist("seller77", 5), dist("seller12", 10)),
				newOF(bidOrder(112, 10, "buyer112"), dist("seller12", 10)),
			},
			expSettlement: &Settlement{PartialOrderLeft: nil},
		},
		{
			name: "three asks, three bids: partial ask",
			askOFs: []*orderFulfillment{
				newOF(askOrder(51, 10, "seller51"), dist("buyer99", 10)),
				newOF(askOrder(77, 15, "seller77"), dist("buyer99", 10), dist("buyer8", 5)),
				newOF(askOrder(12, 21, "seller12"), dist("buyer8", 10), dist("buyer112", 10)),
			},
			bidOFs: []*orderFulfillment{
				newOF(bidOrder(99, 20, "buyer99"), dist("seller51", 10), dist("seller77", 10)),
				newOF(bidOrder(8, 15, "buyer8"), dist("seller77", 5), dist("seller12", 10)),
				newOF(bidOrder(112, 10, "buyer112"), dist("seller12", 10)),
			},
			settlement: &Settlement{},
			expAskOFs: []*orderFulfillment{
				newOF(askOrder(51, 10, "seller51"), dist("buyer99", 10)),
				newOF(askOrder(77, 15, "seller77"), dist("buyer99", 10), dist("buyer8", 5)),
				newOF(askOrder(12, 20, "seller12"), dist("buyer8", 10), dist("buyer112", 10)),
			},
			expBidOfs: []*orderFulfillment{
				newOF(bidOrder(99, 20, "buyer99"), dist("seller51", 10), dist("seller77", 10)),
				newOF(bidOrder(8, 15, "buyer8"), dist("seller77", 5), dist("seller12", 10)),
				newOF(bidOrder(112, 10, "buyer112"), dist("seller12", 10)),
			},
			expSettlement: &Settlement{PartialOrderLeft: askOrder(12, 1, "seller12")},
		},
		{
			name: "three asks, three bids: no partial",
			askOFs: []*orderFulfillment{
				newOF(askOrder(51, 10, "seller51"), dist("buyer99", 10)),
				newOF(askOrder(77, 15, "seller77"), dist("buyer99", 10), dist("buyer8", 5)),
				newOF(askOrder(12, 20, "seller12"), dist("buyer8", 10), dist("buyer112", 10)),
			},
			bidOFs: []*orderFulfillment{
				newOF(bidOrder(99, 20, "buyer99"), dist("seller51", 10), dist("seller77", 10)),
				newOF(bidOrder(8, 15, "buyer8"), dist("seller77", 5), dist("seller12", 10)),
				newOF(bidOrder(112, 11, "buyer112"), dist("seller12", 10)),
			},
			settlement: &Settlement{},
			expAskOFs: []*orderFulfillment{
				newOF(askOrder(51, 10, "seller51"), dist("buyer99", 10)),
				newOF(askOrder(77, 15, "seller77"), dist("buyer99", 10), dist("buyer8", 5)),
				newOF(askOrder(12, 20, "seller12"), dist("buyer8", 10), dist("buyer112", 10)),
			},
			expBidOfs: []*orderFulfillment{
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
	newOF := func(order *Order, assetsFilledAmt int64) *orderFulfillment {
		rv := &orderFulfillment{
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
		fulfillments    []*orderFulfillment
		settlement      *Settlement
		expFulfillments []*orderFulfillment
		expSettlement   *Settlement
		expErr          string
	}{
		{
			name:         "one order, ask: nothing filled",
			fulfillments: []*orderFulfillment{newOF(askOrder(8, 53, false), 0)},
			settlement:   &Settlement{},
			expErr:       "ask order 8 (at index 0) has no assets filled",
		},
		{
			name:         "one order, bid: nothing filled",
			fulfillments: []*orderFulfillment{newOF(bidOrder(8, 53, false), 0)},
			settlement:   &Settlement{},
			expErr:       "bid order 8 (at index 0) has no assets filled",
		},
		{
			name:            "one order, ask: partially filled",
			fulfillments:    []*orderFulfillment{newOF(askOrder(8, 53, true), 13)},
			settlement:      &Settlement{},
			expFulfillments: []*orderFulfillment{newOF(askOrder(8, 13, true), 13)},
			expSettlement:   &Settlement{PartialOrderLeft: askOrder(8, 40, true)},
		},
		{
			name:            "one order, bid: partially filled",
			fulfillments:    []*orderFulfillment{newOF(bidOrder(8, 53, true), 13)},
			settlement:      &Settlement{},
			expFulfillments: []*orderFulfillment{newOF(bidOrder(8, 13, true), 13)},
			expSettlement:   &Settlement{PartialOrderLeft: bidOrder(8, 40, true)},
		},
		{
			name:         "one order, ask: partially filled, already have a partially filled",
			fulfillments: []*orderFulfillment{newOF(askOrder(8, 53, true), 13)},
			settlement:   &Settlement{PartialOrderLeft: bidOrder(951, 357, true)},
			expErr:       "bid order 951 and ask order 8 cannot both be partially filled",
		},
		{
			name:         "one order, bid: partially filled, already have a partially filled",
			fulfillments: []*orderFulfillment{newOF(bidOrder(8, 53, true), 13)},
			settlement:   &Settlement{PartialOrderLeft: askOrder(951, 357, true)},
			expErr:       "ask order 951 and bid order 8 cannot both be partially filled",
		},
		{
			name:         "one order, ask: partially filled, split not allowed",
			fulfillments: []*orderFulfillment{newOF(askOrder(8, 53, false), 13)},
			settlement:   &Settlement{},
			expErr:       "cannot split ask order 8 having assets \"53acorn\" at \"13acorn\": order does not allow partial fulfillment",
		},
		{
			name:         "one order, bid: partially filled, split not allowed",
			fulfillments: []*orderFulfillment{newOF(bidOrder(8, 53, false), 13)},
			settlement:   &Settlement{},
			expErr:       "cannot split bid order 8 having assets \"53acorn\" at \"13acorn\": order does not allow partial fulfillment",
		},
		{
			name:            "one order, ask: fully filled",
			fulfillments:    []*orderFulfillment{newOF(askOrder(8, 53, false), 53)},
			settlement:      &Settlement{},
			expFulfillments: []*orderFulfillment{newOF(askOrder(8, 53, false), 53)},
			expSettlement:   &Settlement{},
		},
		{
			name:            "one order, bid: fully filled",
			fulfillments:    []*orderFulfillment{newOF(bidOrder(8, 53, false), 53)},
			settlement:      &Settlement{},
			expFulfillments: []*orderFulfillment{newOF(bidOrder(8, 53, false), 53)},
			expSettlement:   &Settlement{},
		},
		{
			name:            "one order, ask: fully filled, already have a partially filled",
			fulfillments:    []*orderFulfillment{newOF(askOrder(8, 53, false), 53)},
			settlement:      &Settlement{PartialOrderLeft: bidOrder(951, 357, true)},
			expFulfillments: []*orderFulfillment{newOF(askOrder(8, 53, false), 53)},
			expSettlement:   &Settlement{PartialOrderLeft: bidOrder(951, 357, true)},
		},
		{
			name:            "one order, bid: fully filled, already have a partially filled",
			fulfillments:    []*orderFulfillment{newOF(bidOrder(8, 53, false), 53)},
			settlement:      &Settlement{PartialOrderLeft: askOrder(951, 357, true)},
			expFulfillments: []*orderFulfillment{newOF(bidOrder(8, 53, false), 53)},
			expSettlement:   &Settlement{PartialOrderLeft: askOrder(951, 357, true)},
		},
		{
			name: "three orders, ask: second partially filled",
			fulfillments: []*orderFulfillment{
				newOF(askOrder(8, 53, false), 53),
				newOF(askOrder(9, 17, true), 16),
				newOF(askOrder(10, 200, false), 0),
			},
			settlement: &Settlement{},
			expErr:     "ask order 9 (at index 1) is not filled in full and is not the last ask order provided",
		},
		{
			name: "three orders, bid: second partially filled",
			fulfillments: []*orderFulfillment{
				newOF(bidOrder(8, 53, false), 53),
				newOF(bidOrder(9, 17, true), 16),
				newOF(bidOrder(10, 200, false), 0),
			},
			settlement: &Settlement{},
			expErr:     "bid order 9 (at index 1) is not filled in full and is not the last bid order provided",
		},
		{
			name: "three orders, ask: last not touched",
			fulfillments: []*orderFulfillment{
				newOF(askOrder(8, 53, false), 53),
				newOF(askOrder(9, 17, false), 17),
				newOF(askOrder(10, 200, false), 0),
			},
			settlement: &Settlement{},
			expErr:     "ask order 10 (at index 2) has no assets filled",
		},
		{
			name: "three orders, bid: last not touched",
			fulfillments: []*orderFulfillment{
				newOF(bidOrder(8, 53, false), 53),
				newOF(bidOrder(9, 17, false), 17),
				newOF(bidOrder(10, 200, true), 0),
			},
			settlement: &Settlement{},
			expErr:     "bid order 10 (at index 2) has no assets filled",
		},
		{
			name: "three orders, ask: last partially filled",
			fulfillments: []*orderFulfillment{
				newOF(askOrder(8, 53, false), 53),
				newOF(askOrder(9, 17, false), 17),
				newOF(askOrder(10, 200, true), 183),
			},
			settlement: &Settlement{},
			expFulfillments: []*orderFulfillment{
				newOF(askOrder(8, 53, false), 53),
				newOF(askOrder(9, 17, false), 17),
				newOF(askOrder(10, 183, true), 183),
			},
			expSettlement: &Settlement{PartialOrderLeft: askOrder(10, 17, true)},
		},
		{
			name: "three orders, bid: last partially filled",
			fulfillments: []*orderFulfillment{
				newOF(bidOrder(8, 53, false), 53),
				newOF(bidOrder(9, 17, false), 17),
				newOF(bidOrder(10, 200, true), 183),
			},
			settlement: &Settlement{},
			expFulfillments: []*orderFulfillment{
				newOF(bidOrder(8, 53, false), 53),
				newOF(bidOrder(9, 17, false), 17),
				newOF(bidOrder(10, 183, true), 183),
			},
			expSettlement: &Settlement{PartialOrderLeft: bidOrder(10, 17, true)},
		},
		{
			name: "three orders, ask: last partially filled, split not allowed",
			fulfillments: []*orderFulfillment{
				newOF(askOrder(8, 53, true), 53),
				newOF(askOrder(9, 17, true), 17),
				newOF(askOrder(10, 200, false), 183),
			},
			settlement: &Settlement{},
			expErr:     "cannot split ask order 10 having assets \"200acorn\" at \"183acorn\": order does not allow partial fulfillment",
		},
		{
			name: "three orders, bid: last partially filled, split not allowed",
			fulfillments: []*orderFulfillment{
				newOF(bidOrder(8, 53, true), 53),
				newOF(bidOrder(9, 17, true), 17),
				newOF(bidOrder(10, 200, false), 183),
			},
			settlement: &Settlement{},
			expErr:     "cannot split bid order 10 having assets \"200acorn\" at \"183acorn\": order does not allow partial fulfillment",
		},
		{
			name: "three orders, ask: last partially filled, already have a partially filled",
			fulfillments: []*orderFulfillment{
				newOF(askOrder(8, 53, true), 53),
				newOF(askOrder(9, 17, true), 17),
				newOF(askOrder(10, 200, false), 183),
			},
			settlement: &Settlement{PartialOrderLeft: bidOrder(857, 43, true)},
			expErr:     "bid order 857 and ask order 10 cannot both be partially filled",
		},
		{
			name: "three orders, bid: last partially filled, already have a partially filled",
			fulfillments: []*orderFulfillment{
				newOF(bidOrder(8, 53, true), 53),
				newOF(bidOrder(9, 17, true), 17),
				newOF(bidOrder(10, 200, false), 183),
			},
			settlement: &Settlement{PartialOrderLeft: askOrder(857, 43, true)},
			expErr:     "ask order 857 and bid order 10 cannot both be partially filled",
		},
		{
			name: "three orders, ask: fully filled",
			fulfillments: []*orderFulfillment{
				newOF(askOrder(8, 53, false), 53),
				newOF(askOrder(9, 17, false), 17),
				newOF(askOrder(10, 200, false), 200),
			},
			settlement: &Settlement{},
			expFulfillments: []*orderFulfillment{
				newOF(askOrder(8, 53, false), 53),
				newOF(askOrder(9, 17, false), 17),
				newOF(askOrder(10, 200, false), 200),
			},
			expSettlement: &Settlement{},
		},
		{
			name: "three orders, bid: fully filled",
			fulfillments: []*orderFulfillment{
				newOF(bidOrder(8, 53, false), 53),
				newOF(bidOrder(9, 17, false), 17),
				newOF(bidOrder(10, 200, false), 200),
			},
			settlement: &Settlement{},
			expFulfillments: []*orderFulfillment{
				newOF(bidOrder(8, 53, false), 53),
				newOF(bidOrder(9, 17, false), 17),
				newOF(bidOrder(10, 200, false), 200),
			},
			expSettlement: &Settlement{},
		},
		{
			name: "three orders, ask: fully filled, already have a partially filled",
			fulfillments: []*orderFulfillment{
				newOF(askOrder(8, 53, false), 53),
				newOF(askOrder(9, 17, false), 17),
				newOF(askOrder(10, 200, false), 200),
			},
			settlement: &Settlement{},
			expFulfillments: []*orderFulfillment{
				newOF(askOrder(8, 53, false), 53),
				newOF(askOrder(9, 17, false), 17),
				newOF(askOrder(10, 200, false), 200),
			},
			expSettlement: &Settlement{},
		},
		{
			name: "three orders, bid: fully filled, already have a partially filled",
			fulfillments: []*orderFulfillment{
				newOF(bidOrder(8, 53, false), 53),
				newOF(bidOrder(9, 17, false), 17),
				newOF(bidOrder(10, 200, false), 200),
			},
			settlement: &Settlement{PartialOrderLeft: askOrder(857, 43, true)},
			expFulfillments: []*orderFulfillment{
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
	askOrder := func(orderID uint64, assetsAmt, priceAmt int64) *Order {
		return NewOrder(orderID).WithAsk(&AskOrder{
			MarketId: 123,
			Seller:   fmt.Sprintf("seller%d", orderID),
			Assets:   sdk.NewInt64Coin("apple", assetsAmt),
			Price:    sdk.NewInt64Coin("peach", priceAmt),
		})
	}
	bidOrder := func(orderID uint64, assetsAmt, priceAmt int64) *Order {
		return NewOrder(orderID).WithBid(&BidOrder{
			MarketId: 123,
			Buyer:    fmt.Sprintf("buyer%d", orderID),
			Assets:   sdk.NewInt64Coin("apple", assetsAmt),
			Price:    sdk.NewInt64Coin("peach", priceAmt),
		})
	}
	newOF := func(order *Order, dists ...*distribution) *orderFulfillment {
		rv := newOrderFulfillment(order)
		rv.AssetsFilledAmt, rv.AssetsUnfilledAmt = rv.AssetsUnfilledAmt, rv.AssetsFilledAmt
		if len(dists) > 0 {
			rv.PriceDists = dists
			for _, dist := range dists {
				rv.PriceAppliedAmt = rv.PriceAppliedAmt.Add(dist.Amount)
			}
			rv.PriceLeftAmt = rv.PriceLeftAmt.Sub(rv.PriceAppliedAmt)
		}
		return rv
	}
	dist := func(address string, amount int64) *distribution {
		return &distribution{Address: address, Amount: sdkmath.NewInt(amount)}
	}

	tests := []struct {
		name      string
		askOFs    []*orderFulfillment
		bidOFs    []*orderFulfillment
		expAskOFs []*orderFulfillment
		expBidOFs []*orderFulfillment
		expErr    string
	}{
		{
			name: "total ask price greater than total bid",
			askOFs: []*orderFulfillment{
				newOF(askOrder(3, 10, 20)),
				newOF(askOrder(4, 10, 20)),
				newOF(askOrder(5, 10, 20)),
			},
			bidOFs: []*orderFulfillment{
				newOF(bidOrder(6, 10, 20)),
				newOF(bidOrder(7, 10, 19)),
				newOF(bidOrder(8, 10, 20)),
			},
			expErr: "total ask price \"60peach\" is greater than total bid price \"59peach\"",
		},
		{
			name:      "one ask, one bid: same price",
			askOFs:    []*orderFulfillment{newOF(askOrder(3, 10, 60))},
			bidOFs:    []*orderFulfillment{newOF(bidOrder(6, 10, 60))},
			expAskOFs: []*orderFulfillment{newOF(askOrder(3, 10, 60), dist("buyer6", 60))},
			expBidOFs: []*orderFulfillment{newOF(bidOrder(6, 10, 60), dist("seller3", 60))},
		},
		{
			name:      "one ask, one bid: bid more",
			askOFs:    []*orderFulfillment{newOF(askOrder(3, 10, 60))},
			bidOFs:    []*orderFulfillment{newOF(bidOrder(6, 10, 65))},
			expAskOFs: []*orderFulfillment{newOF(askOrder(3, 10, 60), dist("buyer6", 60), dist("buyer6", 5))},
			expBidOFs: []*orderFulfillment{newOF(bidOrder(6, 10, 65), dist("seller3", 60), dist("seller3", 5))},
		},
		{
			name: "two asks, two bids: same total price, diff ask prices",
			askOFs: []*orderFulfillment{
				newOF(askOrder(3, 10, 21)),
				newOF(askOrder(4, 10, 19)),
			},
			bidOFs: []*orderFulfillment{
				newOF(bidOrder(6, 10, 20)),
				newOF(bidOrder(7, 10, 20)),
			},
			expAskOFs: []*orderFulfillment{
				newOF(askOrder(3, 10, 21), dist("buyer6", 20), dist("buyer7", 1)),
				newOF(askOrder(4, 10, 19), dist("buyer7", 19)),
			},
			expBidOFs: []*orderFulfillment{
				newOF(bidOrder(6, 10, 20), dist("seller3", 20)),
				newOF(bidOrder(7, 10, 20), dist("seller3", 1), dist("seller4", 19)),
			},
		},
		{
			name: "three asks, three bids: same total price",
			askOFs: []*orderFulfillment{
				newOF(askOrder(3, 10, 25)),
				newOF(askOrder(4, 10, 20)),
				newOF(askOrder(5, 10, 15)),
			},
			bidOFs: []*orderFulfillment{
				newOF(bidOrder(6, 10, 18)),
				newOF(bidOrder(7, 10, 30)),
				newOF(bidOrder(8, 10, 12)),
			},
			expAskOFs: []*orderFulfillment{
				newOF(askOrder(3, 10, 25), dist("buyer6", 18), dist("buyer7", 7)),
				newOF(askOrder(4, 10, 20), dist("buyer7", 20)),
				newOF(askOrder(5, 10, 15), dist("buyer7", 3), dist("buyer8", 12)),
			},
			expBidOFs: []*orderFulfillment{
				newOF(bidOrder(6, 10, 18), dist("seller3", 18)),
				newOF(bidOrder(7, 10, 30), dist("seller3", 7), dist("seller4", 20), dist("seller5", 3)),
				newOF(bidOrder(8, 10, 12), dist("seller5", 12)),
			},
		},
		{
			name: "three asks, three bids: bids more",
			askOFs: []*orderFulfillment{
				newOF(askOrder(3, 1, 10)),
				newOF(askOrder(4, 7, 25)),
				newOF(askOrder(5, 22, 30)),
			},
			bidOFs: []*orderFulfillment{
				newOF(bidOrder(6, 10, 20)),
				newOF(bidOrder(7, 10, 27)),
				newOF(bidOrder(8, 10, 30)),
			},
			// assets total = 30
			// ask price total = 65
			// bid price total = 77
			// leftover = 12
			expAskOFs: []*orderFulfillment{
				// 12 * 1 / 30 = 0.4 => 0, then 1
				newOF(askOrder(3, 1, 10), dist("buyer6", 10),
					dist("buyer8", 1)),
				// 12 * 7 / 30 = 2.8 => 2, then because there'll only be 1 left, 1
				newOF(askOrder(4, 7, 25), dist("buyer6", 10), dist("buyer7", 15),
					dist("buyer8", 2), dist("buyer8", 1)),
				// 12 * 22 / 30 = 8.8 => 8, then nothing because leftovers run out before getting back to it.
				newOF(askOrder(5, 22, 30), dist("buyer7", 12), dist("buyer8", 18),
					dist("buyer8", 8)),
			},
			expBidOFs: []*orderFulfillment{
				newOF(bidOrder(6, 10, 20), dist("seller3", 10), dist("seller4", 10)),
				newOF(bidOrder(7, 10, 27), dist("seller4", 15), dist("seller5", 12)),
				newOF(bidOrder(8, 10, 30), dist("seller5", 18), dist("seller4", 2),
					dist("seller5", 8), dist("seller3", 1), dist("seller4", 1)),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = allocatePrice(tc.askOFs, tc.bidOFs)
			}
			require.NotPanics(t, testFunc, "allocatePrice")
			assertions.AssertErrorValue(t, err, tc.expErr, "allocatePrice error")
			if len(tc.expErr) > 0 {
				return
			}
			assertEqualOrderFulfillmentSlices(t, tc.expAskOFs, tc.askOFs, "askOFs after allocatePrice")
			assertEqualOrderFulfillmentSlices(t, tc.expBidOFs, tc.bidOFs, "bidOFs after allocatePrice")
		})
	}
}

func TestGetFulfillmentPriceAmt(t *testing.T) {
	newAskOF := func(orderID uint64, assetsUnfilled int64, assetDenom string) *orderFulfillment {
		return &orderFulfillment{
			Order: NewOrder(orderID).WithAsk(&AskOrder{
				Price: sdk.NewInt64Coin(assetDenom, 999),
			}),
			PriceLeftAmt: sdkmath.NewInt(assetsUnfilled),
		}
	}
	newBidOF := func(orderID uint64, assetsUnfilled int64, assetDenom string) *orderFulfillment {
		return &orderFulfillment{
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
		of1    *orderFulfillment
		of2    *orderFulfillment
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
				amt, err = getFulfillmentPriceAmt(tc.of1, tc.of2)
			}
			require.NotPanics(t, testFunc, "getFulfillmentPriceAmt")
			assertions.AssertErrorValue(t, err, tc.expErr, "getFulfillmentPriceAmt error")
			assert.Equal(t, tc.expAmt, amt, "getFulfillmentPriceAmt amount")
			assertEqualOrderFulfillments(t, origOF1, tc.of1, "of1 after getFulfillmentPriceAmt")
			assertEqualOrderFulfillments(t, origOF2, tc.of2, "of2 after getFulfillmentPriceAmt")
		})
	}
}

func TestSetFeesToPay(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}
	askOF := func(orderID uint64, priceAppliedAmt int64, fees ...sdk.Coin) *orderFulfillment {
		askOrder := &AskOrder{Price: coin(50, "plum")}
		if len(fees) > 1 {
			t.Fatalf("cannot provide more than one fee to askOF(%d, %d, %q)",
				orderID, priceAppliedAmt, fees)
		}
		if len(fees) > 0 {
			askOrder.SellerSettlementFlatFee = &fees[0]
		}
		return &orderFulfillment{
			Order:           NewOrder(orderID).WithAsk(askOrder),
			PriceAppliedAmt: sdkmath.NewInt(priceAppliedAmt),
		}
	}
	bidOF := func(orderID uint64, priceAppliedAmt int64, fees ...sdk.Coin) *orderFulfillment {
		bidOrder := &BidOrder{Price: coin(50, "plum")}
		if len(fees) > 0 {
			bidOrder.BuyerSettlementFees = fees
		}
		return &orderFulfillment{
			Order:           NewOrder(orderID).WithBid(bidOrder),
			PriceAppliedAmt: sdkmath.NewInt(priceAppliedAmt),
		}
	}
	expOF := func(f *orderFulfillment, feesToPay ...sdk.Coin) *orderFulfillment {
		if len(feesToPay) > 0 {
			f.FeesToPay = sdk.NewCoins(feesToPay...)
		}
		return f
	}

	tests := []struct {
		name      string
		askOFs    []*orderFulfillment
		bidOFs    []*orderFulfillment
		ratio     *FeeRatio
		expAskOFs []*orderFulfillment
		expBidOFs []*orderFulfillment
		expErr    string
	}{
		{
			name: "cannot apply ratio",
			askOFs: []*orderFulfillment{
				askOF(7777, 55, coin(20, "grape")),
				askOF(5555, 71),
				askOF(6666, 100),
			},
			bidOFs: []*orderFulfillment{
				bidOF(1111, 100),
				bidOF(2222, 200, coin(20, "grape")),
				bidOF(3333, 300),
			},
			ratio: &FeeRatio{Price: coin(30, "peach"), Fee: coin(1, "fig")},
			expAskOFs: []*orderFulfillment{
				expOF(askOF(7777, 55, coin(20, "grape"))),
				expOF(askOF(5555, 71)),
				expOF(askOF(6666, 100)),
			},
			expBidOFs: []*orderFulfillment{
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
			askOFs: []*orderFulfillment{
				askOF(7777, 55, coin(20, "grape")),
				askOF(5555, 71),
				askOF(6666, 100),
			},
			bidOFs: []*orderFulfillment{
				bidOF(1111, 100),
				bidOF(2222, 200, coin(20, "grape")),
				bidOF(3333, 300),
			},
			ratio: nil,
			expAskOFs: []*orderFulfillment{
				expOF(askOF(7777, 55, coin(20, "grape")), coin(20, "grape")),
				expOF(askOF(5555, 71)),
				expOF(askOF(6666, 100)),
			},
			expBidOFs: []*orderFulfillment{
				expOF(bidOF(1111, 100)),
				expOF(bidOF(2222, 200, coin(20, "grape")), coin(20, "grape")),
				expOF(bidOF(3333, 300)),
			},
		},
		{
			name: "with ratio",
			askOFs: []*orderFulfillment{
				askOF(7777, 55, coin(20, "grape")),
				askOF(5555, 71),
				askOF(6666, 100),
			},
			bidOFs: []*orderFulfillment{
				bidOF(1111, 100),
				bidOF(2222, 200, coin(20, "grape")),
				bidOF(3333, 300),
			},
			ratio: &FeeRatio{Price: coin(30, "plum"), Fee: coin(1, "fig")},
			expAskOFs: []*orderFulfillment{
				expOF(askOF(7777, 55, coin(20, "grape")), coin(2, "fig"), coin(20, "grape")),
				expOF(askOF(5555, 71), coin(3, "fig")),
				expOF(askOF(6666, 100), coin(4, "fig")),
			},
			expBidOFs: []*orderFulfillment{
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
	goodAskOF := func(orderID uint64) *orderFulfillment {
		return &orderFulfillment{
			Order: NewOrder(orderID).WithAsk(&AskOrder{
				Assets: sdk.NewInt64Coin("apple", 50),
				Price:  sdk.NewInt64Coin("peach", 123),
			}),
			AssetsFilledAmt: sdkmath.NewInt(50),
			PriceAppliedAmt: sdkmath.NewInt(130),
		}
	}
	badAskOF := func(orderID uint64) *orderFulfillment {
		rv := goodAskOF(orderID)
		rv.AssetsFilledAmt = sdkmath.NewInt(49)
		return rv
	}
	badAskErr := func(orderID uint64) string {
		return badAskOF(orderID).Validate().Error()
	}
	goodBidOF := func(orderID uint64) *orderFulfillment {
		return &orderFulfillment{
			Order: NewOrder(orderID).WithBid(&BidOrder{
				Assets: sdk.NewInt64Coin("apple", 50),
				Price:  sdk.NewInt64Coin("peach", 123),
			}),
			AssetsFilledAmt: sdkmath.NewInt(50),
			PriceAppliedAmt: sdkmath.NewInt(123),
		}
	}
	badBidOF := func(orderID uint64) *orderFulfillment {
		rv := goodBidOF(orderID)
		rv.AssetsFilledAmt = sdkmath.NewInt(49)
		return rv
	}
	badBidErr := func(orderID uint64) string {
		return badBidOF(orderID).Validate().Error()
	}

	tests := []struct {
		name   string
		askOFs []*orderFulfillment
		bidOFs []*orderFulfillment
		expErr string
	}{
		{
			name:   "all good",
			askOFs: []*orderFulfillment{goodAskOF(10), goodAskOF(11), goodAskOF(12)},
			bidOFs: []*orderFulfillment{goodBidOF(20), goodBidOF(21), goodBidOF(22)},
			expErr: "",
		},
		{
			name:   "error in one ask",
			askOFs: []*orderFulfillment{goodAskOF(10), badAskOF(11), goodAskOF(12)},
			bidOFs: []*orderFulfillment{goodBidOF(20), goodBidOF(21), goodBidOF(22)},
			expErr: badAskErr(11),
		},
		{
			name:   "error in one bid",
			askOFs: []*orderFulfillment{goodAskOF(10), goodAskOF(11), goodAskOF(12)},
			bidOFs: []*orderFulfillment{goodBidOF(20), badBidOF(21), goodBidOF(22)},
			expErr: badBidErr(21),
		},
		{
			name:   "two errors in asks",
			askOFs: []*orderFulfillment{badAskOF(10), goodAskOF(11), badAskOF(12)},
			bidOFs: []*orderFulfillment{goodBidOF(20), goodBidOF(21), goodBidOF(22)},
			expErr: joinErrs(badAskErr(10), badAskErr(12)),
		},
		{
			name:   "two errors in bids",
			askOFs: []*orderFulfillment{goodAskOF(10), goodAskOF(11), goodAskOF(12)},
			bidOFs: []*orderFulfillment{badBidOF(20), goodBidOF(21), badBidOF(22)},
			expErr: joinErrs(badBidErr(20), badBidErr(22)),
		},
		{
			name:   "error in each",
			askOFs: []*orderFulfillment{goodAskOF(10), goodAskOF(11), badAskOF(12)},
			bidOFs: []*orderFulfillment{goodBidOF(20), badBidOF(21), goodBidOF(22)},
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
		f      orderFulfillment
		expErr string
	}{
		{
			name:   "nil inside order",
			f:      orderFulfillment{Order: NewOrder(8)},
			expErr: nilSubTypeErr(8),
		},
		{
			name:   "unknown inside order",
			f:      orderFulfillment{Order: newUnknownOrder(12)},
			expErr: unknownSubTypeErr(12),
		},
		{
			name: "order price greater than price applied: ask",
			f: orderFulfillment{
				Order:           askOrder(52, 10, 401),
				PriceAppliedAmt: sdkmath.NewInt(400),
				AssetsFilledAmt: sdkmath.NewInt(10),
			},
			expErr: "ask order 52 price \"401peach\" is more than price filled \"400peach\"",
		},
		{
			name: "order price equal to price applied: ask",
			f: orderFulfillment{
				Order:           askOrder(53, 10, 401),
				PriceAppliedAmt: sdkmath.NewInt(401),
				AssetsFilledAmt: sdkmath.NewInt(10),
			},
			expErr: "",
		},
		{
			name: "order price less than price applied: ask",
			f: orderFulfillment{
				Order:           askOrder(54, 10, 401),
				PriceAppliedAmt: sdkmath.NewInt(402),
				AssetsFilledAmt: sdkmath.NewInt(10),
			},
			expErr: "",
		},
		{
			name: "order price greater than price applied: bid",
			f: orderFulfillment{
				Order:           bidOrder(71, 17, 432),
				PriceAppliedAmt: sdkmath.NewInt(431),
				AssetsFilledAmt: sdkmath.NewInt(17),
			},
			expErr: "bid order 71 price \"432peach\" is not equal to price filled \"431peach\"",
		},
		{
			name: "order price equal to price applied: bid",
			f: orderFulfillment{
				Order:           bidOrder(72, 17, 432),
				PriceAppliedAmt: sdkmath.NewInt(432),
				AssetsFilledAmt: sdkmath.NewInt(17),
			},
			expErr: "",
		},
		{
			name: "order price less than price applied: bid",
			f: orderFulfillment{
				Order:           bidOrder(73, 17, 432),
				PriceAppliedAmt: sdkmath.NewInt(433),
				AssetsFilledAmt: sdkmath.NewInt(17),
			},
			expErr: "bid order 73 price \"432peach\" is not equal to price filled \"433peach\"",
		},
		{
			name: "order assets less than assets filled: ask",
			f: orderFulfillment{
				Order:           askOrder(101, 53, 12345),
				PriceAppliedAmt: sdkmath.NewInt(12345),
				AssetsFilledAmt: sdkmath.NewInt(54),
			},
			expErr: "ask order 101 assets \"53apple\" does not equal filled assets \"54apple\"",
		},
		{
			name: "order assets equal to assets filled: ask",
			f: orderFulfillment{
				Order:           askOrder(202, 53, 12345),
				PriceAppliedAmt: sdkmath.NewInt(12345),
				AssetsFilledAmt: sdkmath.NewInt(53),
			},
			expErr: "",
		},
		{
			name: "order assets more than assets filled: ask",
			f: orderFulfillment{
				Order:           askOrder(303, 53, 12345),
				PriceAppliedAmt: sdkmath.NewInt(12345),
				AssetsFilledAmt: sdkmath.NewInt(52),
			},
			expErr: "ask order 303 assets \"53apple\" does not equal filled assets \"52apple\"",
		},
		{
			name: "order assets less than assets filled: bid",
			f: orderFulfillment{
				Order:           bidOrder(404, 53, 12345),
				PriceAppliedAmt: sdkmath.NewInt(12345),
				AssetsFilledAmt: sdkmath.NewInt(54),
			},
			expErr: "bid order 404 assets \"53apple\" does not equal filled assets \"54apple\"",
		},
		{
			name: "order assets equal to assets filled: bid",
			f: orderFulfillment{
				Order:           bidOrder(505, 53, 12345),
				PriceAppliedAmt: sdkmath.NewInt(12345),
				AssetsFilledAmt: sdkmath.NewInt(53),
			},
			expErr: "",
		},
		{
			name: "order assets more than assets filled: bid",
			f: orderFulfillment{
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

func TestBuildTransfers(t *testing.T) {
	tests := []struct {
		name          string
		askOFs        []*orderFulfillment
		bidOFs        []*orderFulfillment
		expSettlement *Settlement
		expErr        string
	}{
		{
			name: "ask with negative assets filled",
			askOFs: []*orderFulfillment{
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
			bidOFs: []*orderFulfillment{
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
			askOFs: []*orderFulfillment{
				{
					Order: NewOrder(53).WithAsk(&AskOrder{
						Assets: sdk.NewInt64Coin("apple", 15),
						Price:  sdk.NewInt64Coin("plum", 42),
					}),
					AssetsFilledAmt: sdkmath.NewInt(15),
					AssetDists:      []*distribution{{Address: "buyer1", Amount: sdkmath.NewInt(15)}},
					PriceAppliedAmt: sdkmath.NewInt(42),
					PriceDists:      []*distribution{{Address: "seller1", Amount: sdkmath.NewInt(42)}},
					FeesToPay:       sdk.Coins{sdk.Coin{Denom: "fig", Amount: sdkmath.NewInt(-1)}},
				},
			},
			expErr: "ask order 53 cannot pay \"-1fig\" in fees: negative amount",
		},
		{
			name: "bid with negative fees to pay",
			bidOFs: []*orderFulfillment{
				{
					Order: NewOrder(35).WithBid(&BidOrder{
						Assets: sdk.NewInt64Coin("apple", 15),
						Price:  sdk.NewInt64Coin("plum", 42),
					}),
					AssetsFilledAmt: sdkmath.NewInt(15),
					AssetDists:      []*distribution{{Address: "seller1", Amount: sdkmath.NewInt(15)}},
					PriceAppliedAmt: sdkmath.NewInt(42),
					PriceDists:      []*distribution{{Address: "seller1", Amount: sdkmath.NewInt(42)}},
					FeesToPay:       sdk.Coins{sdk.Coin{Denom: "fig", Amount: sdkmath.NewInt(-1)}},
				},
			},
			expErr: "bid order 35 cannot pay \"-1fig\" in fees: negative amount",
		},
		{
			name: "two asks, three bids",
			askOFs: []*orderFulfillment{
				{
					Order: NewOrder(77).WithAsk(&AskOrder{
						Seller: "seller77",
						Assets: sdk.NewInt64Coin("apple", 15),
						Price:  sdk.NewInt64Coin("plum", 42),
					}),
					AssetsFilledAmt: sdkmath.NewInt(15),
					AssetDists: []*distribution{
						{Address: "buyer5511", Amount: sdkmath.NewInt(15)},
					},
					PriceAppliedAmt: sdkmath.NewInt(43),
					PriceDists: []*distribution{
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
					AssetDists: []*distribution{
						{Address: "buyer5511", Amount: sdkmath.NewInt(5)},
						{Address: "buyer78", Amount: sdkmath.NewInt(7)},
						{Address: "buyer9001", Amount: sdkmath.NewInt(31)},
					},
					PriceAppliedAmt: sdkmath.NewInt(90),
					PriceDists: []*distribution{
						{Address: "buyer78", Amount: sdkmath.NewInt(5)},
						{Address: "buyer9001", Amount: sdkmath.NewInt(83)},
						{Address: "buyer9001", Amount: sdkmath.NewInt(2)},
					},
					FeesToPay: nil,
				},
			},
			bidOFs: []*orderFulfillment{
				{
					Order: NewOrder(5511).WithBid(&BidOrder{
						Buyer:  "buyer5511",
						Assets: sdk.NewInt64Coin("apple", 20),
						Price:  sdk.NewInt64Coin("plum", 30),
					}),
					AssetsFilledAmt: sdkmath.NewInt(20),
					AssetDists: []*distribution{
						{Address: "seller77", Amount: sdkmath.NewInt(15)},
						{Address: "seller3", Amount: sdkmath.NewInt(5)},
					},
					PriceAppliedAmt: sdkmath.NewInt(30),
					PriceDists: []*distribution{
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
					AssetDists: []*distribution{
						{Address: "seller3", Amount: sdkmath.NewInt(7)},
					},
					PriceAppliedAmt: sdkmath.NewInt(15),
					PriceDists: []*distribution{
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
					AssetDists: []*distribution{
						{Address: "seller3", Amount: sdkmath.NewInt(31)},
					},
					PriceAppliedAmt: sdkmath.NewInt(86),
					PriceDists: []*distribution{
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
	newOF := func(order *Order, priceAppliedAmt int64, fees ...sdk.Coin) *orderFulfillment {
		rv := &orderFulfillment{
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
		askOFs        []*orderFulfillment
		bidOFs        []*orderFulfillment
		settlement    *Settlement
		expSettlement *Settlement
	}{
		{
			name: "no partial",
			askOFs: []*orderFulfillment{
				newOF(askOrder(2001, 53, 87), 92, coin(12, "fig")),
				newOF(askOrder(2002, 17, 33), 37),
				newOF(askOrder(2003, 22, 56), 60),
			},
			bidOFs: []*orderFulfillment{
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
			askOFs: []*orderFulfillment{
				newOF(askOrder(2001, 53, 87), 92, coin(12, "fig")),
				newOF(askOrder(2002, 17, 33), 37),
				newOF(askOrder(2003, 22, 56), 60),
			},
			bidOFs: []*orderFulfillment{
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
			askOFs: []*orderFulfillment{
				newOF(askOrder(2001, 53, 87), 92, coin(12, "fig")),
				newOF(askOrder(2002, 17, 33), 37),
				newOF(askOrder(2003, 22, 56), 60),
			},
			bidOFs: []*orderFulfillment{
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
	newOF := func(order *Order, assetsFilledAmt int64, assetDists ...*distribution) *orderFulfillment {
		rv := &orderFulfillment{
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
	dist := func(addr string, amount int64) *distribution {
		return &distribution{Address: addr, Amount: sdkmath.NewInt(amount)}
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
		f           *orderFulfillment
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
			assertEqualOrderFulfillments(t, orig, tc.f, "orderFulfillment before and after getAssetTransfer")
		})
	}
}

func TestGetPriceTransfer(t *testing.T) {
	seller, buyer := "sally", "brandon"
	assetDenom, priceDenom := "apple", "peach"
	newOF := func(order *Order, priceAppliedAmt int64, priceDists ...*distribution) *orderFulfillment {
		rv := &orderFulfillment{
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
	dist := func(addr string, amount int64) *distribution {
		return &distribution{Address: addr, Amount: sdkmath.NewInt(amount)}
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
		f           *orderFulfillment
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
			assertEqualOrderFulfillments(t, orig, tc.f, "orderFulfillment before and after getPriceTransfer")
		})
	}
}

func TestGetNAVs(t *testing.T) {
	coin := func(coinStr string) sdk.Coin {
		rv, err := parseCoin(coinStr)
		require.NoError(t, err, "parseCoin(%q)", coinStr)
		return rv
	}
	askOrder := func(orderID uint64, assets, price string) *FilledOrder {
		order := NewOrder(orderID).WithAsk(&AskOrder{Assets: coin(assets), Price: coin(price)})
		return NewFilledOrder(order, order.GetPrice(), nil)
	}
	bidOrder := func(orderID uint64, assets, price string) *FilledOrder {
		order := NewOrder(orderID).WithBid(&BidOrder{Assets: coin(assets), Price: coin(price)})
		return NewFilledOrder(order, order.GetPrice(), nil)
	}
	navStringer := func(nav *NetAssetValue) string {
		if nav == nil {
			return "<nil>"
		}
		return fmt.Sprintf("%s:%s", nav.Assets, nav.Price)
	}

	tests := []struct {
		name       string
		settlement *Settlement
		expNAVs    []*NetAssetValue
	}{
		{
			name:       "no orders",
			settlement: &Settlement{},
			expNAVs:    nil,
		},
		{
			name: "no filled bids, no partial",
			settlement: &Settlement{
				FullyFilledOrders: []*FilledOrder{
					askOrder(1, "10apple", "20plum"),
					askOrder(2, "15apple", "30plum"),
					askOrder(3, "20apple", "35plum"),
				},
			},
			expNAVs: nil,
		},
		{
			name: "no filled bids, partial ask",
			settlement: &Settlement{
				FullyFilledOrders: []*FilledOrder{
					askOrder(1, "10apple", "20plum"),
					askOrder(2, "15apple", "30plum"),
					askOrder(3, "20apple", "35plum"),
				},
				PartialOrderFilled: askOrder(4, "8apple", "12plum"),
			},
			expNAVs: nil,
		},
		{
			name: "no filled bids, partial bid",
			settlement: &Settlement{
				FullyFilledOrders: []*FilledOrder{
					askOrder(1, "10apple", "20plum"),
					askOrder(2, "15apple", "30plum"),
					askOrder(3, "20apple", "35plum"),
				},
				PartialOrderFilled: bidOrder(4, "8apple", "12plum"),
			},
			expNAVs: []*NetAssetValue{{Assets: coin("8apple"), Price: coin("12plum")}},
		},
		{
			name: "one filled bid, no partial",
			settlement: &Settlement{
				FullyFilledOrders: []*FilledOrder{
					askOrder(1, "10apple", "20plum"),
					askOrder(2, "15apple", "30plum"),
					bidOrder(3, "8apple", "12plum"),
					askOrder(4, "20apple", "35plum"),
				},
			},
			expNAVs: []*NetAssetValue{{Assets: coin("8apple"), Price: coin("12plum")}},
		},
		{
			name: "one filled bid, partial ask",
			settlement: &Settlement{
				FullyFilledOrders: []*FilledOrder{
					askOrder(1, "10apple", "20plum"),
					askOrder(2, "15apple", "30plum"),
					bidOrder(3, "8apple", "12plum"),
					askOrder(4, "20apple", "35plum"),
				},
				PartialOrderFilled: askOrder(5, "55apple", "114plum"),
			},
			expNAVs: []*NetAssetValue{{Assets: coin("8apple"), Price: coin("12plum")}},
		},
		{
			name: "one filled bid plus partial bid, same denoms",
			settlement: &Settlement{
				FullyFilledOrders: []*FilledOrder{
					askOrder(1, "10apple", "20plum"),
					askOrder(2, "15apple", "30plum"),
					bidOrder(3, "8apple", "12plum"),
					askOrder(4, "20apple", "35plum"),
				},
				PartialOrderFilled: bidOrder(5, "55apple", "114plum"),
			},
			expNAVs: []*NetAssetValue{{Assets: coin("63apple"), Price: coin("126plum")}},
		},
		{
			name: "one filled bid plus partial bid, diff asset denoms",
			settlement: &Settlement{
				FullyFilledOrders: []*FilledOrder{
					askOrder(1, "10apple", "20plum"),
					askOrder(2, "15apple", "30plum"),
					bidOrder(3, "8apple", "12plum"),
					askOrder(4, "20apple", "35plum"),
				},
				PartialOrderFilled: bidOrder(5, "55acorn", "114plum"),
			},
			expNAVs: []*NetAssetValue{
				{Assets: coin("8apple"), Price: coin("12plum")},
				{Assets: coin("55acorn"), Price: coin("114plum")},
			},
		},
		{
			name: "one filled bid plus partial bid, diff price denoms",
			settlement: &Settlement{
				FullyFilledOrders: []*FilledOrder{
					askOrder(1, "10apple", "20plum"),
					askOrder(2, "15apple", "30plum"),
					bidOrder(3, "8apple", "12plum"),
					askOrder(4, "20apple", "35plum"),
				},
				PartialOrderFilled: bidOrder(5, "55apple", "114pear"),
			},
			expNAVs: []*NetAssetValue{
				{Assets: coin("8apple"), Price: coin("12plum")},
				{Assets: coin("55apple"), Price: coin("114pear")},
			},
		},
		{
			name: "one filled bid plus partial bid, diff both denoms",
			settlement: &Settlement{
				FullyFilledOrders: []*FilledOrder{
					askOrder(1, "10apple", "20plum"),
					askOrder(2, "15apple", "30plum"),
					bidOrder(3, "8apple", "12plum"),
					askOrder(4, "20apple", "35plum"),
				},
				PartialOrderFilled: bidOrder(5, "55acorn", "114pear"),
			},
			expNAVs: []*NetAssetValue{
				{Assets: coin("8apple"), Price: coin("12plum")},
				{Assets: coin("55acorn"), Price: coin("114pear")},
			},
		},
		{
			name: "four bids, same denoms",
			settlement: &Settlement{
				FullyFilledOrders: []*FilledOrder{
					bidOrder(1, "10apple", "20plum"),
					bidOrder(2, "8apple", "12plum"),
					askOrder(3, "15apple", "30plum"),
					bidOrder(4, "20apple", "35plum"),
					askOrder(5, "27apple", "56plum"),
				},
				PartialOrderFilled: bidOrder(6, "55apple", "114plum"),
			},
			expNAVs: []*NetAssetValue{{Assets: coin("93apple"), Price: coin("181plum")}},
		},
		{
			name: "four bids, two asset denoms",
			settlement: &Settlement{
				FullyFilledOrders: []*FilledOrder{
					bidOrder(1, "10apple", "20plum"),
					bidOrder(2, "8acorn", "12plum"),
					askOrder(3, "15apple", "30plum"),
					bidOrder(4, "20apple", "35plum"),
					askOrder(5, "27apple", "56plum"),
					bidOrder(6, "55acorn", "114plum"),
				},
			},
			expNAVs: []*NetAssetValue{
				{Assets: coin("30apple"), Price: coin("55plum")},
				{Assets: coin("63acorn"), Price: coin("126plum")},
			},
		},
		{
			name: "four bids, two price denoms",
			settlement: &Settlement{
				FullyFilledOrders: []*FilledOrder{
					bidOrder(1, "10apple", "20pear"),
					bidOrder(2, "8apple", "12pear"),
					askOrder(3, "15apple", "30plum"),
					bidOrder(4, "20apple", "35plum"),
					askOrder(5, "27apple", "56plum"),
					bidOrder(6, "55apple", "114plum"),
				},
			},
			expNAVs: []*NetAssetValue{
				{Assets: coin("18apple"), Price: coin("32pear")},
				{Assets: coin("75apple"), Price: coin("149plum")},
			},
		},
		{
			name: "four bids, same denoms",
			settlement: &Settlement{
				FullyFilledOrders: []*FilledOrder{
					bidOrder(1, "10apple", "20pear"),
					bidOrder(2, "8apple", "12plum"),
					askOrder(3, "15apple", "30plum"),
					bidOrder(4, "20apple", "35plum"),
					askOrder(5, "27apple", "56plum"),
				},
				PartialOrderFilled: bidOrder(6, "55acorn", "114pear"),
			},
			expNAVs: []*NetAssetValue{
				{Assets: coin("10apple"), Price: coin("20pear")},
				{Assets: coin("28apple"), Price: coin("47plum")},
				{Assets: coin("55acorn"), Price: coin("114pear")},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var navs []*NetAssetValue
			testFunc := func() {
				navs = GetNAVs(tc.settlement)
			}
			require.NotPanics(t, testFunc, "GetNAVs")
			if !assert.Equal(t, tc.expNAVs, navs, "GetNavs result") {
				// Do the comparison as strings to hopefully make it easier to see what's different.
				expNavs := stringerLines(tc.expNAVs, navStringer)
				actNavs := stringerLines(navs, navStringer)
				assert.Equal(t, expNavs, actNavs, "GetNavs result as strings")
			}
		})
	}
}
