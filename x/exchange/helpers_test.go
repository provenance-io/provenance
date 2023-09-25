package exchange

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO[1658]: func TestEqualsUint64(t *testing.T)

func TestContainsUint64(t *testing.T) {
	tests := []struct {
		name   string
		vals   []uint64
		toFind uint64
		exp    bool
	}{
		{
			name:   "nil vals",
			vals:   nil,
			toFind: 0,
			exp:    false,
		},
		{
			name:   "empty vals",
			vals:   []uint64{},
			toFind: 0,
			exp:    false,
		},
		{
			name:   "one val: same",
			vals:   []uint64{1},
			toFind: 1,
			exp:    true,
		},
		{
			name:   "one val: different",
			vals:   []uint64{1},
			toFind: 2,
			exp:    false,
		},
		{
			name:   "three vals: not found",
			vals:   []uint64{1, 2, 3},
			toFind: 0,
			exp:    false,
		},
		{
			name:   "three vals: first",
			vals:   []uint64{1, 2, 3},
			toFind: 1,
			exp:    true,
		},
		{
			name:   "three vals: second",
			vals:   []uint64{1, 2, 3},
			toFind: 2,
			exp:    true,
		},
		{
			name:   "three vals: third",
			vals:   []uint64{1, 2, 3},
			toFind: 3,
			exp:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = ContainsUint64(tc.vals, tc.toFind)
			}
			require.NotPanics(t, testFunc, "ContainsUint64(%q, %q)", tc.vals, tc.toFind)
			assert.Equal(t, tc.exp, actual, "ContainsUint64(%q, %q)", tc.vals, tc.toFind)
		})
	}
}

// TODO[1658]: func TestIntersectionUint64(t *testing.T)

func TestContainsString(t *testing.T) {
	tests := []struct {
		name   string
		vals   []string
		toFind string
		exp    bool
	}{
		{
			name:   "nil vals",
			vals:   nil,
			toFind: "",
			exp:    false,
		},
		{
			name:   "empty vals",
			vals:   []string{},
			toFind: "",
			exp:    false,
		},
		{
			name:   "one val: same",
			vals:   []string{"one"},
			toFind: "one",
			exp:    true,
		},
		{
			name:   "one val: different",
			vals:   []string{"one"},
			toFind: "two",
			exp:    false,
		},
		{
			name:   "one val: space at end of val",
			vals:   []string{"one "},
			toFind: "one",
			exp:    false,
		},
		{
			name:   "one val: space at end of toFind",
			vals:   []string{"one"},
			toFind: "one ",
			exp:    false,
		},
		{
			name:   "one val: space at start of val",
			vals:   []string{" one"},
			toFind: "one",
			exp:    false,
		},
		{
			name:   "one val: space at start of toFind",
			vals:   []string{"one"},
			toFind: " one",
			exp:    false,
		},
		{
			name:   "one val: different casing",
			vals:   []string{"one"},
			toFind: "oNe",
			exp:    false,
		},
		{
			name:   "three vals: not found",
			vals:   []string{"one", "two", "three"},
			toFind: "zero",
			exp:    false,
		},
		{
			name:   "three vals: first",
			vals:   []string{"one", "two", "three"},
			toFind: "one",
			exp:    true,
		},
		{
			name:   "three vals: second",
			vals:   []string{"one", "two", "three"},
			toFind: "two",
			exp:    true,
		},
		{
			name:   "three vals: third",
			vals:   []string{"one", "two", "three"},
			toFind: "three",
			exp:    true,
		},
		{
			name:   "three vals: empty string",
			vals:   []string{"one", "two", "three"},
			toFind: "",
			exp:    false,
		},
		{
			name:   "empty string in vals: finding empty string",
			vals:   []string{"one", "", "three"},
			toFind: "",
			exp:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = ContainsString(tc.vals, tc.toFind)
			}
			require.NotPanics(t, testFunc, "ContainsString(%q, %q)", tc.vals, tc.toFind)
			assert.Equal(t, tc.exp, actual, "ContainsString(%q, %q)", tc.vals, tc.toFind)
		})
	}
}

// TODO[1658]: func TestCoinsEquals(t *testing.T)

func TestCoinEquals(t *testing.T) {
	tests := []struct {
		name string
		a    sdk.Coin
		b    sdk.Coin
		exp  bool
	}{
		{
			name: "zero-value coins",
			a:    sdk.Coin{},
			b:    sdk.Coin{},
			exp:  true,
		},
		{
			name: "different amounts",
			a:    sdk.NewInt64Coin("pear", 2),
			b:    sdk.NewInt64Coin("pear", 3),
			exp:  false,
		},
		{
			name: "different denoms",
			a:    sdk.NewInt64Coin("pear", 2),
			b:    sdk.NewInt64Coin("onion", 2),
			exp:  false,
		},
		{
			name: "same denom and amount",
			a:    sdk.NewInt64Coin("pear", 2),
			b:    sdk.NewInt64Coin("pear", 2),
			exp:  true,
		},
		{
			name: "same denom zero amounts",
			a:    sdk.NewInt64Coin("pear", 0),
			b:    sdk.NewInt64Coin("pear", 0),
			exp:  true,
		},
		{
			name: "diff denom zero amounts",
			a:    sdk.NewInt64Coin("pear", 0),
			b:    sdk.NewInt64Coin("onion", 0),
			exp:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = CoinEquals(tc.a, tc.b)
			}
			require.NotPanics(t, testFunc, "CoinEquals(%q, %q)", tc.a, tc.b)
			assert.Equal(t, tc.exp, actual, "CoinEquals(%q, %q) result", tc.a, tc.b)
		})
	}
}

func TestIntersectionOfCoin(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}

	tests := []struct {
		name     string
		options1 []sdk.Coin
		options2 []sdk.Coin
		expected []sdk.Coin
	}{
		{name: "nil nil", options1: nil, options2: nil, expected: nil},
		{name: "nil empty", options1: nil, options2: []sdk.Coin{}, expected: nil},
		{name: "empty nil", options1: []sdk.Coin{}, options2: nil, expected: nil},
		{name: "empty empty", options1: []sdk.Coin{}, options2: []sdk.Coin{}, expected: nil},
		{
			name:     "one nil",
			options1: []sdk.Coin{coin(1, "finger")},
			options2: nil,
			expected: nil,
		},
		{
			name:     "nil one",
			options1: nil,
			options2: []sdk.Coin{coin(1, "finger")},
			expected: nil,
		},
		{
			name:     "one one same",
			options1: []sdk.Coin{coin(1, "finger")},
			options2: []sdk.Coin{coin(1, "finger")},
			expected: []sdk.Coin{coin(1, "finger")},
		},
		{
			name:     "one one different first amount",
			options1: []sdk.Coin{coin(2, "finger")},
			options2: []sdk.Coin{coin(1, "finger")},
			expected: nil,
		},
		{
			name:     "one one different first denom",
			options1: []sdk.Coin{coin(1, "toe")},
			options2: []sdk.Coin{coin(1, "finger")},
			expected: nil,
		},
		{
			name:     "one one different second amount",
			options1: []sdk.Coin{coin(1, "finger")},
			options2: []sdk.Coin{coin(2, "finger")},
			expected: nil,
		},
		{
			name:     "one one different second denom",
			options1: []sdk.Coin{coin(1, "finger")},
			options2: []sdk.Coin{coin(1, "toe")},
			expected: nil,
		},
		{
			name:     "three three two common",
			options1: []sdk.Coin{coin(1, "finger"), coin(2, "toe"), coin(3, "elbow")},
			options2: []sdk.Coin{coin(5, "toe"), coin(3, "elbow"), coin(1, "finger")},
			expected: []sdk.Coin{coin(1, "finger"), coin(3, "elbow")},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []sdk.Coin
			testFunc := func() {
				actual = IntersectionOfCoin(tc.options1, tc.options2)
			}
			require.NotPanics(t, testFunc, "IntersectionOfCoin")
			assert.Equal(t, tc.expected, actual, "IntersectionOfCoin result")
		})
	}
}

func TestMinSDKInt(t *testing.T) {
	newInt := func(val string) sdkmath.Int {
		rv, ok := sdkmath.NewIntFromString(val)
		require.True(t, ok, "sdkmath.NewIntFromString(%s) resulting bool", val)
		return rv
	}

	posBig := newInt("123456789012345678901234567890")
	negBig := posBig.Neg()
	posBigger := posBig.Add(sdkmath.OneInt())

	tests := []struct {
		name string
		a    sdkmath.Int
		b    sdkmath.Int
		exp  sdkmath.Int
	}{
		{name: "-big -big", a: negBig, b: negBig, exp: negBig},
		{name: "-big -2  ", a: negBig, b: sdkmath.NewInt(-2), exp: negBig},
		{name: "-big -1  ", a: negBig, b: sdkmath.NewInt(-1), exp: negBig},
		{name: "-big 0   ", a: negBig, b: sdkmath.NewInt(0), exp: negBig},
		{name: "-big 1   ", a: negBig, b: sdkmath.NewInt(1), exp: negBig},
		{name: "-big 5   ", a: negBig, b: sdkmath.NewInt(5), exp: negBig},
		{name: "-big big ", a: negBig, b: posBig, exp: negBig},

		{name: "-2 -big", a: sdkmath.NewInt(-2), b: negBig, exp: negBig},
		{name: "-2 -2  ", a: sdkmath.NewInt(-2), b: sdkmath.NewInt(-2), exp: sdkmath.NewInt(-2)},
		{name: "-2 -1  ", a: sdkmath.NewInt(-2), b: sdkmath.NewInt(-1), exp: sdkmath.NewInt(-2)},
		{name: "-2 0   ", a: sdkmath.NewInt(-2), b: sdkmath.NewInt(0), exp: sdkmath.NewInt(-2)},
		{name: "-2 1   ", a: sdkmath.NewInt(-2), b: sdkmath.NewInt(1), exp: sdkmath.NewInt(-2)},
		{name: "-2 5   ", a: sdkmath.NewInt(-2), b: sdkmath.NewInt(5), exp: sdkmath.NewInt(-2)},
		{name: "-2 big ", a: sdkmath.NewInt(-2), b: posBig, exp: sdkmath.NewInt(-2)},

		{name: "-1 -big", a: sdkmath.NewInt(-1), b: negBig, exp: negBig},
		{name: "-1 -2  ", a: sdkmath.NewInt(-1), b: sdkmath.NewInt(-2), exp: sdkmath.NewInt(-2)},
		{name: "-1 -1  ", a: sdkmath.NewInt(-1), b: sdkmath.NewInt(-1), exp: sdkmath.NewInt(-1)},
		{name: "-1 0   ", a: sdkmath.NewInt(-1), b: sdkmath.NewInt(0), exp: sdkmath.NewInt(-1)},
		{name: "-1 1   ", a: sdkmath.NewInt(-1), b: sdkmath.NewInt(1), exp: sdkmath.NewInt(-1)},
		{name: "-1 5   ", a: sdkmath.NewInt(-1), b: sdkmath.NewInt(5), exp: sdkmath.NewInt(-1)},
		{name: "-1 big ", a: sdkmath.NewInt(-1), b: posBig, exp: sdkmath.NewInt(-1)},

		{name: "0 -big", a: sdkmath.NewInt(0), b: negBig, exp: negBig},
		{name: "0 -2  ", a: sdkmath.NewInt(0), b: sdkmath.NewInt(-2), exp: sdkmath.NewInt(-2)},
		{name: "0 -1  ", a: sdkmath.NewInt(0), b: sdkmath.NewInt(-1), exp: sdkmath.NewInt(-1)},
		{name: "0 0   ", a: sdkmath.NewInt(0), b: sdkmath.NewInt(0), exp: sdkmath.NewInt(0)},
		{name: "0 1   ", a: sdkmath.NewInt(0), b: sdkmath.NewInt(1), exp: sdkmath.NewInt(0)},
		{name: "0 5   ", a: sdkmath.NewInt(0), b: sdkmath.NewInt(5), exp: sdkmath.NewInt(0)},
		{name: "0 big ", a: sdkmath.NewInt(0), b: posBig, exp: sdkmath.NewInt(0)},

		{name: "1 -big", a: sdkmath.NewInt(1), b: negBig, exp: negBig},
		{name: "1 -2  ", a: sdkmath.NewInt(1), b: sdkmath.NewInt(-2), exp: sdkmath.NewInt(-2)},
		{name: "1 -1  ", a: sdkmath.NewInt(1), b: sdkmath.NewInt(-1), exp: sdkmath.NewInt(-1)},
		{name: "1 0   ", a: sdkmath.NewInt(1), b: sdkmath.NewInt(0), exp: sdkmath.NewInt(0)},
		{name: "1 1   ", a: sdkmath.NewInt(1), b: sdkmath.NewInt(1), exp: sdkmath.NewInt(1)},
		{name: "1 5   ", a: sdkmath.NewInt(1), b: sdkmath.NewInt(5), exp: sdkmath.NewInt(1)},
		{name: "1 big ", a: sdkmath.NewInt(1), b: posBig, exp: sdkmath.NewInt(1)},

		{name: "5 -big", a: sdkmath.NewInt(5), b: negBig, exp: negBig},
		{name: "5 -2  ", a: sdkmath.NewInt(5), b: sdkmath.NewInt(-2), exp: sdkmath.NewInt(-2)},
		{name: "5 -1  ", a: sdkmath.NewInt(5), b: sdkmath.NewInt(-1), exp: sdkmath.NewInt(-1)},
		{name: "5 0   ", a: sdkmath.NewInt(5), b: sdkmath.NewInt(0), exp: sdkmath.NewInt(0)},
		{name: "5 1   ", a: sdkmath.NewInt(5), b: sdkmath.NewInt(1), exp: sdkmath.NewInt(1)},
		{name: "5 5   ", a: sdkmath.NewInt(5), b: sdkmath.NewInt(5), exp: sdkmath.NewInt(5)},
		{name: "5 big ", a: sdkmath.NewInt(5), b: posBig, exp: sdkmath.NewInt(5)},

		{name: "big -big", a: posBig, b: negBig, exp: negBig},
		{name: "big -2  ", a: posBig, b: sdkmath.NewInt(-2), exp: sdkmath.NewInt(-2)},
		{name: "big -1  ", a: posBig, b: sdkmath.NewInt(-1), exp: sdkmath.NewInt(-1)},
		{name: "big 0   ", a: posBig, b: sdkmath.NewInt(0), exp: sdkmath.NewInt(0)},
		{name: "big 1   ", a: posBig, b: sdkmath.NewInt(1), exp: sdkmath.NewInt(1)},
		{name: "big 5   ", a: posBig, b: sdkmath.NewInt(5), exp: sdkmath.NewInt(5)},
		{name: "big big ", a: posBig, b: posBig, exp: posBig},

		{name: "big bigger", a: posBig, b: posBigger, exp: posBig},
		{name: "bigger big", a: posBigger, b: posBig, exp: posBig},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdkmath.Int
			testFunc := func() {
				actual = MinSDKInt(tc.a, tc.b)
			}
			require.NotPanics(t, testFunc, "MinSDKInt(%s, %s)", tc.a, tc.b)
			assert.Equal(t, tc.exp, actual, "MinSDKInt(%s, %s)", tc.a, tc.b)
		})
	}
}
