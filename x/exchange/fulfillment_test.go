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

// TODO[1658]: func TestNewOrderFulfillment(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_GetAssetsFilled(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_GetAssetsUnfilled(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_GetPriceApplied(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_GetPriceLeft(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_GetPriceFilled(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_GetPriceUnfilled(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_IsFullyFilled(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_IsCompletelyUnfulfilled(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_GetOrderID(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_IsAskOrder(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_IsBidOrder(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_GetMarketID(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_GetOwner(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_GetAssets(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_GetPrice(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_GetSettlementFees(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_PartialFillAllowed(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_GetOrderType(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_GetOrderTypeByte(t *testing.T)

// TODO[1658]: func TestOrderFulfillment_GetHoldAmount(t *testing.T)

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

// String converts everything in this to a string.
// This is mostly because test failure output of sdk.Coin and sdk.Coins is impossible to understand.
func (i *indexedAddrAmts) String() string {
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
					t.Logf("Original: %s", orig)
					t.Logf("  Actual: %s", tc.receiver)
					t.Logf("Expected: %s", tc.expected)
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
		{name: "nil receiver", receiver: nil, expPanic: "cannot get inputs from empty set"},
		{name: "no addrs", receiver: newIndexedAddrAmts(), expPanic: "cannot get inputs from empty set"},
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
				t.Logf("Before: %s", orig)
				t.Logf(" After: %s", tc.receiver)
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
		{name: "nil receiver", receiver: nil, expPanic: "cannot get inputs from empty set"},
		{name: "no addrs", receiver: newIndexedAddrAmts(), expPanic: "cannot get inputs from empty set"},
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
				t.Logf("Before: %s", orig)
				t.Logf(" After: %s", tc.receiver)
			}
		})
	}
}

// TODO[1658]: func TestBuildSettlementTransfers(t *testing.T)

// TODO[1658]: func TestGetAssetTransfer(t *testing.T)

// TODO[1658]: func TestGetPriceTransfer(t *testing.T)
