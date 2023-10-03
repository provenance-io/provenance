package exchange

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestEqualsUint64(t *testing.T) {
	tests := []struct {
		name string
		a    uint64
		b    uint64
		exp  bool
	}{
		{name: "0 0", a: 0, b: 0, exp: true},
		{name: "0 1", a: 0, b: 1, exp: false},
		{name: "1 0", a: 1, b: 0, exp: false},
		{name: "1 1", a: 1, b: 1, exp: true},
		{name: "1 max uint32+1", a: 1, b: 4_294_967_296, exp: false},
		{name: "max uint32+1 1", a: 4_294_967_296, b: 1, exp: false},
		{name: "max uint32+1 max uint32+1", a: 4_294_967_296, b: 4_294_967_296, exp: true},
		{name: "1 max uint64", a: 1, b: 18_446_744_073_709_551_615, exp: false},
		{name: "max uint64 1", a: 18_446_744_073_709_551_615, b: 1, exp: false},
		{name: "max uint64 max uint64", a: 18_446_744_073_709_551_615, b: 18_446_744_073_709_551_615, exp: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = EqualsUint64(tc.a, tc.b)
			}
			require.NotPanics(t, testFunc, "EqualsUint64(%d, %d)", tc.a, tc.b)
			assert.Equal(t, tc.exp, actual, "EqualsUint64(%d, %d)", tc.a, tc.b)
		})
	}
}

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

func TestIntersectionUint64(t *testing.T) {
	tests := []struct {
		name string
		a    []uint64
		b    []uint64
		exp  []uint64
	}{
		{name: "nil nil", a: nil, b: nil, exp: nil},
		{name: "nil empty", a: nil, b: []uint64{}, exp: nil},
		{name: "empty nil", a: []uint64{}, b: nil, exp: nil},
		{name: "empty empty", a: []uint64{}, b: []uint64{}, exp: nil},
		{name: "nil one", a: nil, b: []uint64{1}, exp: nil},
		{name: "one nil", a: []uint64{1}, b: nil, exp: nil},
		{name: "one one same", a: []uint64{1}, b: []uint64{1}, exp: []uint64{1}},
		{name: "one one different", a: []uint64{1}, b: []uint64{2}, exp: nil},
		{name: "three one first", a: []uint64{1, 2, 3}, b: []uint64{1}, exp: []uint64{1}},
		{name: "three one second", a: []uint64{1, 2, 3}, b: []uint64{2}, exp: []uint64{2}},
		{name: "three one third", a: []uint64{1, 2, 3}, b: []uint64{3}, exp: []uint64{3}},
		{name: "three one none", a: []uint64{1, 2, 3}, b: []uint64{4}, exp: nil},

		{name: "three two none", a: []uint64{1, 2, 3}, b: []uint64{4, 5}, exp: nil},
		{name: "three two first rep", a: []uint64{1, 2, 3}, b: []uint64{1, 1}, exp: []uint64{1}},
		{name: "three two only first", a: []uint64{1, 2, 3}, b: []uint64{4, 1}, exp: []uint64{1}},
		{name: "three two second rep", a: []uint64{1, 2, 3}, b: []uint64{2, 2}, exp: []uint64{2}},
		{name: "three two only second", a: []uint64{1, 2, 3}, b: []uint64{4, 2}, exp: []uint64{2}},
		{name: "three two third rep", a: []uint64{1, 2, 3}, b: []uint64{3, 3}, exp: []uint64{3}},
		{name: "three two only third", a: []uint64{1, 2, 3}, b: []uint64{4, 3}, exp: []uint64{3}},
		{name: "three two not third", a: []uint64{1, 2, 3}, b: []uint64{2, 1}, exp: []uint64{1, 2}},
		{name: "three two not second", a: []uint64{1, 2, 3}, b: []uint64{3, 1}, exp: []uint64{1, 3}},
		{name: "three two not first", a: []uint64{1, 2, 3}, b: []uint64{3, 2}, exp: []uint64{2, 3}},

		{name: "three rep one same", a: []uint64{5, 5, 5}, b: []uint64{5}, exp: []uint64{5}},
		{name: "three rep one different", a: []uint64{5, 5, 5}, b: []uint64{6}, exp: nil},
		{name: "three rep two rep same", a: []uint64{5, 5, 5}, b: []uint64{5, 5}, exp: []uint64{5}},
		{name: "three rep two rep different", a: []uint64{5, 5, 5}, b: []uint64{6, 6}, exp: nil},
		{name: "three three one same", a: []uint64{1, 2, 3}, b: []uint64{4, 5, 1}, exp: []uint64{1}},
		{name: "three three two same", a: []uint64{1, 2, 3}, b: []uint64{3, 4, 2}, exp: []uint64{2, 3}},
		{name: "three three all same diff order", a: []uint64{1, 2, 3, 2, 1}, b: []uint64{2, 1, 1, 1, 2, 3, 1}, exp: []uint64{1, 2, 3}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []uint64
			testFunc := func() {
				actual = IntersectionUint64(tc.a, tc.b)
			}
			require.NotPanics(t, testFunc, "IntersectionUint64")
			assert.Equal(t, tc.exp, actual, "IntersectionUint64 result")
		})
	}
}

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

func TestCoinsEquals(t *testing.T) {
	coins := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "sdk.ParseCoinsNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name string
		a    sdk.Coins
		b    sdk.Coins
		exp  bool
	}{
		{name: "nil nil", a: nil, b: nil, exp: true},
		{name: "nil empty", a: nil, b: sdk.Coins{}, exp: true},
		{name: "empty nil", a: sdk.Coins{}, b: nil, exp: true},
		{name: "empty empty", a: nil, b: sdk.Coins{}, exp: true},
		{name: "nil one", a: nil, b: coins("1one"), exp: false},
		{name: "one nil", a: coins("1one"), b: nil, exp: false},
		{name: "one one same", a: coins("1one"), b: coins("1one"), exp: true},
		{name: "one one diff amount", a: coins("1one"), b: coins("2one"), exp: false},
		{name: "one one diff denom", a: coins("1one"), b: coins("1two"), exp: false},
		{name: "one one diff both", a: coins("1one"), b: coins("2two"), exp: false},
		{name: "two one first", a: coins("1one,2two"), b: coins("1one"), exp: false},
		{name: "two one second", a: coins("1one,2two"), b: coins("2two"), exp: false},
		{name: "two one neither", a: coins("1one,2two"), b: coins("3three"), exp: false},
		{name: "two two same", a: coins("1one,2two"), b: coins("1one,2two"), exp: true},
		{name: "two two diff first denom", a: coins("1one,2two"), b: coins("1three,2two"), exp: false},
		{name: "two two diff second denom", a: coins("1one,2two"), b: coins("1one,2three"), exp: false},
		{
			name: "one one same negative",
			a:    sdk.Coins{sdk.Coin{Denom: "one", Amount: sdkmath.NewInt(-1)}},
			b:    sdk.Coins{sdk.Coin{Denom: "one", Amount: sdkmath.NewInt(-1)}},
			exp:  true,
		},
		{
			name: "one one negative first",
			a:    sdk.Coins{sdk.Coin{Denom: "one", Amount: sdkmath.NewInt(-1)}},
			b:    sdk.Coins{sdk.Coin{Denom: "one", Amount: sdkmath.NewInt(1)}},
			exp:  false,
		},
		{
			name: "one one negative second",
			a:    sdk.Coins{sdk.Coin{Denom: "one", Amount: sdkmath.NewInt(1)}},
			b:    sdk.Coins{sdk.Coin{Denom: "one", Amount: sdkmath.NewInt(-1)}},
			exp:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = CoinsEquals(tc.a, tc.b)
			}
			require.NotPanics(t, testFunc, "CoinsEquals(%q, %q)", tc.a, tc.b)
			assert.Equal(t, tc.exp, actual, "CoinsEquals(%q, %q)", tc.a, tc.b)
		})
	}
}

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

func TestContainsCoin(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}

	tests := []struct {
		name   string
		vals   []sdk.Coin
		toFind sdk.Coin
		exp    bool
	}{
		{name: "nil vals", vals: nil, toFind: coin(1, "banana"), exp: false},
		{name: "empty vals", vals: []sdk.Coin{}, toFind: coin(1, "banana"), exp: false},
		{
			name:   "one val, diff denom and amount",
			vals:   []sdk.Coin{coin(3, "banana")},
			toFind: coin(8, "apple"),
			exp:    false,
		},
		{
			name:   "one val, same denom diff amount",
			vals:   []sdk.Coin{coin(3, "apple")},
			toFind: coin(8, "apple"),
			exp:    false,
		},
		{
			name:   "one val, diff denom same amount",
			vals:   []sdk.Coin{coin(8, "banana")},
			toFind: coin(8, "apple"),
			exp:    false,
		},
		{
			name:   "one val, same denom and amount",
			vals:   []sdk.Coin{coin(8, "apple")},
			toFind: coin(8, "apple"),
			exp:    true,
		},
		{
			name:   "one neg val, same",
			vals:   []sdk.Coin{coin(-3, "apple")},
			toFind: coin(-3, "apple"),
			exp:    true,
		},
		{
			name:   "one val without denom, same",
			vals:   []sdk.Coin{coin(22, "")},
			toFind: coin(22, ""),
			exp:    true,
		},
		{
			name:   "one val zero, diff denom",
			vals:   []sdk.Coin{coin(0, "banana")},
			toFind: coin(0, "apple"),
			exp:    false,
		},
		{
			name:   "one val zero, same denom",
			vals:   []sdk.Coin{coin(0, "banana")},
			toFind: coin(0, "banana"),
			exp:    true,
		},
		{
			name:   "three same vals, not to find",
			vals:   []sdk.Coin{coin(1, "apple"), coin(1, "apple"), coin(1, "apple")},
			toFind: coin(2, "apple"),
			exp:    false,
		},
		{
			name:   "three vals, not to find",
			vals:   []sdk.Coin{coin(1, "apple"), coin(2, "banana"), coin(3, "cactus")},
			toFind: coin(4, "durian"),
			exp:    false,
		},
		{
			name:   "three vals, first",
			vals:   []sdk.Coin{coin(1, "apple"), coin(2, "banana"), coin(3, "cactus")},
			toFind: coin(1, "apple"),
			exp:    true,
		},
		{
			name:   "three vals, second",
			vals:   []sdk.Coin{coin(1, "apple"), coin(2, "banana"), coin(3, "cactus")},
			toFind: coin(2, "banana"),
			exp:    true,
		},
		{
			name:   "three vals, third",
			vals:   []sdk.Coin{coin(1, "apple"), coin(2, "banana"), coin(3, "cactus")},
			toFind: coin(3, "cactus"),
			exp:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = ContainsCoin(tc.vals, tc.toFind)
			}
			require.NotPanics(t, testFunc, "ContainsCoin(%q, %q)", sdk.Coins(tc.vals), tc.toFind)
			assert.Equal(t, tc.exp, actual, "ContainsCoin(%q, %q)", sdk.Coins(tc.vals), tc.toFind)
		})
	}
}

func TestContainsCoinWithSameDenom(t *testing.T) {
	coin := func(amount int64, denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amount)}
	}

	tests := []struct {
		name   string
		vals   []sdk.Coin
		toFind sdk.Coin
		exp    bool
	}{
		{
			name:   "nil vals",
			vals:   nil,
			toFind: coin(1, "apple"),
			exp:    false,
		},
		{
			name:   "empty vals",
			vals:   []sdk.Coin{},
			toFind: coin(1, "apple"),
			exp:    false,
		},
		{
			name:   "one val, same amount, diff denom",
			vals:   []sdk.Coin{coin(1, "apple")},
			toFind: coin(1, "banana"),
			exp:    false,
		},
		{
			name:   "one val, same",
			vals:   []sdk.Coin{coin(1, "apple")},
			toFind: coin(1, "apple"),
			exp:    true,
		},
		{
			name:   "one val, same denom, diff amount",
			vals:   []sdk.Coin{coin(1, "apple")},
			toFind: coin(2, "apple"),
			exp:    true,
		},
		{
			name:   "three vals, not to find",
			vals:   []sdk.Coin{coin(1, "apple"), coin(2, "banana"), coin(3, "cactus")},
			toFind: coin(4, "durian"),
			exp:    false,
		},
		{
			name:   "three vals, first",
			vals:   []sdk.Coin{coin(1, "apple"), coin(2, "banana"), coin(3, "cactus")},
			toFind: coin(4, "apple"),
			exp:    true,
		},
		{
			name:   "three vals, second",
			vals:   []sdk.Coin{coin(1, "apple"), coin(2, "banana"), coin(3, "cactus")},
			toFind: coin(4, "banana"),
			exp:    true,
		},
		{
			name:   "three vals, third",
			vals:   []sdk.Coin{coin(1, "apple"), coin(2, "banana"), coin(3, "cactus")},
			toFind: coin(4, "cactus"),
			exp:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = ContainsCoinWithSameDenom(tc.vals, tc.toFind)
			}
			require.NotPanics(t, testFunc, "ContainsCoinWithSameDenom(%q, %q)", sdk.Coins(tc.vals), tc.toFind)
			assert.Equal(t, tc.exp, actual, "ContainsCoinWithSameDenom(%q, %q)", sdk.Coins(tc.vals), tc.toFind)
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
