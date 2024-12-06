package nav_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/testutil/assertions"
	. "github.com/provenance-io/provenance/x/nav"
)

// newCoin is similar to NewInt64Coin with the args in the right order and without the validation (also shorter).
func newCoin(amt int64, denom string) sdk.Coin {
	return sdk.Coin{Denom: denom, Amount: sdkmath.NewInt(amt)}
}

// joinErrs joins multiple expected error strings into what errors.Join will make.
func joinErrs(errs ...string) string {
	return strings.Join(errs, "\n")
}

func TestNetAssetValue_String(t *testing.T) {
	tests := []struct {
		name string
		nav  *NetAssetValue
		exp  string
	}{
		{
			name: "nil",
			nav:  nil,
			exp:  "<nil>",
		},
		{
			name: "empty",
			nav:  &NetAssetValue{},
			exp:  "<nil>=<nil>",
		},
		{
			name: "zero assets",
			nav: &NetAssetValue{
				Assets: newCoin(0, "red"),
				Price:  newCoin(5, "orange"),
			},
			exp: "0red=5orange",
		},
		{
			name: "zero price",
			nav: &NetAssetValue{
				Assets: newCoin(12, "yellow"),
				Price:  newCoin(0, "green"),
			},
			exp: "12yellow=0green",
		},
		{
			name: "normal",
			nav: &NetAssetValue{
				Assets: newCoin(753, "blue"),
				Price:  newCoin(248, "purple"),
			},
			exp: "753blue=248purple",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.nav.String()
			}
			require.NotPanics(t, testFunc, "%#v.String()", tc.nav)
			assert.Equal(t, tc.exp, act, "%#v.String()", tc.nav)
		})
	}
}

func TestNetAssetValue_Validate(t *testing.T) {
	tests := []struct {
		name   string
		nav    *NetAssetValue
		expErr string
	}{
		{
			name:   "nil",
			nav:    nil,
			expErr: "nav cannot be nil",
		},
		{
			name: "invalid assets",
			nav: &NetAssetValue{
				Assets: newCoin(-3, "pink"),
				Price:  newCoin(7, "blue"),
			},
			expErr: "invalid assets \"-3pink\": negative coin amount: -3",
		},
		{
			name: "zero assets",
			nav: &NetAssetValue{
				Assets: newCoin(0, "pink"),
				Price:  newCoin(7, "blue"),
			},
			expErr: "invalid assets \"0pink\": cannot be zero",
		},
		{
			name: "invalid price",
			nav: &NetAssetValue{
				Assets: newCoin(3, "pink"),
				Price:  newCoin(-6, "blue"),
			},
			expErr: "invalid price \"-6blue\": negative coin amount: -6",
		},
		{
			name: "zero price",
			nav: &NetAssetValue{
				Assets: newCoin(3, "pink"),
				Price:  newCoin(0, "blue"),
			},
		},
		{
			name: "same denoms",
			nav: &NetAssetValue{
				Assets: newCoin(12, "fuscia"),
				Price:  newCoin(18, "fuscia"),
			},
			expErr: "nav assets \"12fuscia\" and price \"18fuscia\" must have different denoms",
		},
		{
			name: "normal",
			nav: &NetAssetValue{
				Assets: newCoin(8213, "yellow"),
				Price:  newCoin(55, "blue"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.nav.Validate()
			}
			require.NotPanics(t, testFunc, "%#v.Validate()", tc.nav)
			assertions.AssertErrorValue(t, err, tc.expErr, "%#v.Validate() error", tc.nav)
		})
	}
}

func TestNetAssetValue_AsRecord(t *testing.T) {
	tests := []struct {
		name   string
		nav    *NetAssetValue
		height uint64
		source string
		exp    *NetAssetValueRecord
	}{
		{
			name:   "nil",
			nav:    nil,
			height: 20,
			source: "brown",
			exp:    &NetAssetValueRecord{Height: 20, Source: "brown"},
		},
		{
			name:   "empty",
			nav:    &NetAssetValue{},
			height: 818,
			source: "varsity",
			exp:    &NetAssetValueRecord{Height: 818, Source: "varsity"},
		},
		{
			name:   "normal",
			nav:    &NetAssetValue{Assets: newCoin(12, "maroon"), Price: newCoin(73, "yellow")},
			height: 406,
			source: "huckleberry",
			exp: &NetAssetValueRecord{
				Assets: newCoin(12, "maroon"),
				Price:  newCoin(73, "yellow"),
				Height: 406,
				Source: "huckleberry",
			},
		},
		{
			name:   "zero height",
			nav:    &NetAssetValue{Assets: newCoin(12, "maroon"), Price: newCoin(73, "yellow")},
			height: 0,
			source: "huckleberry",
			exp: &NetAssetValueRecord{
				Assets: newCoin(12, "maroon"),
				Price:  newCoin(73, "yellow"),
				Height: 0,
				Source: "huckleberry",
			},
		},
		{
			name:   "no source",
			nav:    &NetAssetValue{Assets: newCoin(12, "maroon"), Price: newCoin(73, "yellow")},
			height: 406,
			source: "",
			exp: &NetAssetValueRecord{
				Assets: newCoin(12, "maroon"),
				Price:  newCoin(73, "yellow"),
				Height: 406,
				Source: "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act *NetAssetValueRecord
			testFunc := func() {
				act = tc.nav.AsRecord(tc.height, tc.source)
			}
			require.NotPanics(t, testFunc, "%#v.AsRecord(%d, %q)", tc.nav, tc.height, tc.source)
			assert.Equal(t, tc.exp, act, "%#v.AsRecord(%d, %q) result", tc.nav, tc.height, tc.source)
		})
	}
}

func TestNAVs_String(t *testing.T) {
	tests := []struct {
		name string
		navs NAVs
		exp  string
	}{
		{
			name: "nil",
			navs: nil,
			exp:  "<nil>",
		},
		{
			name: "empty",
			navs: NAVs{},
			exp:  "[]",
		},
		{
			name: "one nav",
			navs: NAVs{{Assets: newCoin(4, "yellow"), Price: newCoin(99, "blue")}},
			exp:  "[4yellow=99blue]",
		},
		{
			name: "one nil nav",
			navs: NAVs{nil},
			exp:  "[<nil>]",
		},
		{
			name: "three navs",
			navs: NAVs{
				{Assets: newCoin(1, "red"), Price: newCoin(2, "purple")},
				{Assets: newCoin(3, "blue"), Price: newCoin(4, "green")},
				{Assets: newCoin(5, "yellow"), Price: newCoin(6, "orange")},
			},
			exp: "[1red=2purple,3blue=4green,5yellow=6orange]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.navs.String()
			}
			require.NotPanics(t, testFunc, "%#v.String()", tc.navs)
			assert.Equal(t, tc.exp, act, "%#v.String() result", tc.navs)
		})
	}
}

func TestNAVs_Validate(t *testing.T) {
	tests := []struct {
		name   string
		navs   NAVs
		expErr string
	}{
		{
			name:   "nil",
			navs:   nil,
			expErr: "",
		},
		{
			name:   "empty",
			navs:   NAVs{},
			expErr: "",
		},
		{
			name:   "one entry: nil",
			navs:   NAVs{nil},
			expErr: "0: nav cannot be nil",
		},
		{
			name:   "one entry: invalid",
			navs:   NAVs{{Assets: newCoin(-3, "red"), Price: newCoin(4, "blue")}},
			expErr: "0: invalid assets \"-3red\": negative coin amount: -3",
		},
		{
			name:   "one entry: valid",
			navs:   NAVs{{Assets: newCoin(3, "red"), Price: newCoin(4, "blue")}},
			expErr: "",
		},
		{
			name: "two entries: both okay",
			navs: NAVs{
				{Assets: newCoin(3, "red"), Price: newCoin(4, "blue")},
				{Assets: newCoin(7, "green"), Price: newCoin(9, "purple")},
			},
			expErr: "",
		},
		{
			name: "two entries: first invalid",
			navs: NAVs{
				{Assets: newCoin(-3, "red"), Price: newCoin(4, "blue")},
				{Assets: newCoin(7, "green"), Price: newCoin(9, "purple")},
			},
			expErr: "0: invalid assets \"-3red\": negative coin amount: -3",
		},
		{
			name: "two entries: second invalid",
			navs: NAVs{
				{Assets: newCoin(3, "red"), Price: newCoin(4, "blue")},
				{Assets: newCoin(7, "green"), Price: newCoin(-9, "purple")},
			},
			expErr: "1: invalid price \"-9purple\": negative coin amount: -9",
		},
		{
			name: "two entries: both invalid",
			navs: NAVs{
				{Assets: newCoin(-3, "red"), Price: newCoin(4, "blue")},
				{Assets: newCoin(7, "green"), Price: newCoin(-9, "purple")},
			},
			expErr: joinErrs(
				"0: invalid assets \"-3red\": negative coin amount: -3",
				"1: invalid price \"-9purple\": negative coin amount: -9",
			),
		},
		{
			name: "two entries: same assets denoms",
			navs: NAVs{
				{Assets: newCoin(3, "red"), Price: newCoin(4, "blue")},
				{Assets: newCoin(7, "red"), Price: newCoin(9, "purple")},
			},
			expErr: "",
		},
		{
			name: "two entries: same price denoms",
			navs: NAVs{
				{Assets: newCoin(3, "red"), Price: newCoin(4, "purple")},
				{Assets: newCoin(7, "green"), Price: newCoin(9, "purple")},
			},
			expErr: "",
		},
		{
			name: "two entries: same assets and price denoms",
			navs: NAVs{
				{Assets: newCoin(3, "red"), Price: newCoin(4, "purple")},
				{Assets: newCoin(7, "red"), Price: newCoin(9, "purple")},
			},
			expErr: "cannot have multiple (2) navs with the same asset (\"red\") and price (\"purple\") denoms",
		},
		{
			name: "five entries: two sets of same assets and price denoms",
			navs: NAVs{
				{Assets: newCoin(1, "red"), Price: newCoin(2, "purple")},
				{Assets: newCoin(3, "red"), Price: newCoin(4, "purple")},
				{Assets: newCoin(5, "red"), Price: newCoin(5, "purple")},
				{Assets: newCoin(7, "purple"), Price: newCoin(8, "red")},
				{Assets: newCoin(9, "purple"), Price: newCoin(10, "red")},
			},
			expErr: joinErrs(
				"cannot have multiple (3) navs with the same asset (\"red\") and price (\"purple\") denoms",
				"cannot have multiple (2) navs with the same asset (\"purple\") and price (\"red\") denoms",
			),
		},
		{
			name: "two entries: opposite denoms",
			navs: NAVs{
				{Assets: newCoin(3, "red"), Price: newCoin(4, "purple")},
				{Assets: newCoin(7, "purple"), Price: newCoin(9, "red")},
			},
			expErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc1 := func() {
				err = tc.navs.Validate()
			}
			testFunc2 := func() {
				err = ValidateNAVs(tc.navs)
			}
			if assert.NotPanics(t, testFunc1, "%#v.Validate()", tc.navs) {
				assertions.AssertErrorValue(t, err, tc.expErr, "%#v.Validate() error", tc.navs)
			}
			if assert.NotPanics(t, testFunc2, "ValidateNAVs(%#v)", tc.navs) {
				assertions.AssertErrorValue(t, err, tc.expErr, "ValidateNAVs(%#v) error", tc.navs)
			}
		})
	}
}

func TestNAVs_AsRecords(t *testing.T) {
	tests := []struct {
		name   string
		navs   NAVs
		height uint64
		source string
		exp    NAVRecords
	}{
		{
			name:   "nil",
			navs:   nil,
			height: 12,
			source: "yellow",
			exp:    nil,
		},
		{
			name:   "empty",
			navs:   NAVs{},
			height: 14,
			source: "green",
			exp:    NAVRecords{},
		},
		{
			name:   "one entry: nil",
			navs:   NAVs{nil},
			height: 72,
			source: "black",
			exp:    NAVRecords{{Height: 72, Source: "black"}},
		},
		{
			name:   "one entry: empty",
			navs:   NAVs{{}},
			height: 15,
			source: "rainbow",
			exp:    NAVRecords{{Height: 15, Source: "rainbow"}},
		},
		{
			name:   "one entry: normal",
			navs:   NAVs{{Assets: newCoin(3, "red"), Price: newCoin(7, "orange")}},
			height: 4,
			source: "fuscia",
			exp:    NAVRecords{{Assets: newCoin(3, "red"), Price: newCoin(7, "orange"), Height: 4, Source: "fuscia"}},
		},
		{
			name: "three entries: all normal",
			navs: NAVs{
				{Assets: newCoin(22, "red"), Price: newCoin(555, "orange")},
				{Assets: newCoin(44, "red"), Price: newCoin(333, "yellow")},
				{Assets: newCoin(55, "red"), Price: newCoin(111, "brown")},
			},
			height: 42,
			source: "green",
			exp: NAVRecords{
				{Assets: newCoin(22, "red"), Price: newCoin(555, "orange"), Height: 42, Source: "green"},
				{Assets: newCoin(44, "red"), Price: newCoin(333, "yellow"), Height: 42, Source: "green"},
				{Assets: newCoin(55, "red"), Price: newCoin(111, "brown"), Height: 42, Source: "green"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act NAVRecords
			testFunc1 := func() {
				act = tc.navs.AsRecords(tc.height, tc.source)
			}
			testFunc2 := func() {
				act = NAVsAsRecords(tc.navs, tc.height, tc.source)
			}
			if assert.NotPanics(t, testFunc1, "%#v.AsRecords(%d, %q)", tc.navs, tc.height, tc.source) {
				assert.Equal(t, tc.exp, act, "%#v.AsRecords(%d, %q) result", tc.navs, tc.height, tc.source)
			}
			if assert.NotPanics(t, testFunc2, "NAVsAsRecords(%#v, %d, %q)", tc.navs, tc.height, tc.source) {
				assert.Equal(t, tc.exp, act, "NAVsAsRecords(%#v, %d, %q) result", tc.navs, tc.height, tc.source)
			}
		})
	}
}

func TestNetAssetValueRecord_String(t *testing.T) {
	tests := []struct {
		name string
		nav  *NetAssetValueRecord
		exp  string
	}{
		{
			name: "nil",
			nav:  nil,
			exp:  "<nil>",
		},
		{
			name: "empty",
			nav:  &NetAssetValueRecord{},
			exp:  "<nil>=<nil>@0 by \"\"",
		},
		{
			name: "zero assets",
			nav: &NetAssetValueRecord{
				Assets: newCoin(0, "red"),
				Price:  newCoin(5, "orange"),
				Height: 8,
				Source: "blue",
			},
			exp: "0red=5orange@8 by \"blue\"",
		},
		{
			name: "zero price",
			nav: &NetAssetValueRecord{
				Assets: newCoin(12, "yellow"),
				Price:  newCoin(0, "green"),
				Height: 14,
				Source: "pink",
			},
			exp: "12yellow=0green@14 by \"pink\"",
		},
		{
			name: "zero height",
			nav: &NetAssetValueRecord{
				Assets: newCoin(753, "blue"),
				Price:  newCoin(248, "purple"),
				Height: 0,
				Source: "orange",
			},
			exp: "753blue=248purple@0 by \"orange\"",
		},
		{
			name: "empty source",
			nav: &NetAssetValueRecord{
				Assets: newCoin(12, "yellow"),
				Price:  newCoin(0, "green"),
				Height: 14,
				Source: "",
			},
			exp: "12yellow=0green@14 by \"\"",
		},
		{
			name: "normal",
			nav: &NetAssetValueRecord{
				Assets: newCoin(753, "blue"),
				Price:  newCoin(248, "purple"),
				Height: 44,
				Source: "orange",
			},
			exp: "753blue=248purple@44 by \"orange\"",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.nav.String()
			}
			require.NotPanics(t, testFunc, "%#v.String()", tc.nav)
			assert.Equal(t, tc.exp, act, "%#v.String() result", tc.nav)
		})
	}
}

func TestNetAssetValueRecord_Validate(t *testing.T) {
	tests := []struct {
		name   string
		nav    *NetAssetValueRecord
		expErr string
	}{
		{
			name:   "nil",
			nav:    nil,
			expErr: "nav record cannot be nil",
		},
		{
			name: "normal",
			nav: &NetAssetValueRecord{
				Assets: newCoin(8213, "yellow"),
				Price:  newCoin(55, "blue"),
				Height: 3,
				Source: "yellow",
			},
		},
		{
			name: "invalid assets",
			nav: &NetAssetValueRecord{
				Assets: newCoin(-3, "pink"),
				Price:  newCoin(7, "blue"),
				Height: 3,
				Source: "yellow",
			},
			expErr: "invalid assets \"-3pink\": negative coin amount: -3",
		},
		{
			name: "zero assets",
			nav: &NetAssetValueRecord{
				Assets: newCoin(0, "pink"),
				Price:  newCoin(7, "blue"),
				Height: 3,
				Source: "yellow",
			},
			expErr: "invalid assets \"0pink\": cannot be zero",
		},
		{
			name: "invalid price",
			nav: &NetAssetValueRecord{
				Assets: newCoin(3, "pink"),
				Price:  newCoin(-6, "blue"),
				Height: 3,
				Source: "yellow",
			},
			expErr: "invalid price \"-6blue\": negative coin amount: -6",
		},
		{
			name: "zero price",
			nav: &NetAssetValueRecord{
				Assets: newCoin(3, "pink"),
				Price:  newCoin(0, "blue"),
				Height: 3,
				Source: "yellow",
			},
		},
		{
			name: "same denoms",
			nav: &NetAssetValueRecord{
				Assets: newCoin(12, "fuscia"),
				Price:  newCoin(18, "fuscia"),
				Height: 3,
				Source: "yellow",
			},
			expErr: "nav assets \"12fuscia\" and price \"18fuscia\" must have different denoms",
		},
		{
			name: "zero height",
			nav: &NetAssetValueRecord{
				Assets: newCoin(8213, "yellow"),
				Price:  newCoin(55, "blue"),
				Height: 0,
				Source: "yellow",
			},
		},
		{
			name: "no source",
			nav: &NetAssetValueRecord{
				Assets: newCoin(8213, "yellow"),
				Price:  newCoin(55, "blue"),
				Height: 0,
				Source: "",
			},
			expErr: "invalid source \"\": cannot be empty",
		},
		{
			name: "max length source",
			nav: &NetAssetValueRecord{
				Assets: newCoin(8213, "yellow"),
				Price:  newCoin(55, "blue"),
				Height: 0,
				Source: strings.Repeat("p", SourceMaxLen),
			},
		},
		{
			name: "max length +1 source",
			nav: &NetAssetValueRecord{
				Assets: newCoin(8213, "yellow"),
				Price:  newCoin(55, "blue"),
				Height: 0,
				Source: "pb1" + strings.Repeat("r", SourceMaxLen-6) + "abcd",
			},
			expErr: fmt.Sprintf("invalid source \"pb1rrrr...rrabcd\": length %d exceeds max %d", SourceMaxLen+1, SourceMaxLen),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.nav.Validate()
			}
			require.NotPanics(t, testFunc, "%#v.Validate()", tc.nav)
			assertions.AssertErrorValue(t, err, tc.expErr, "%#v.Validate() error", tc.nav)
		})
	}
}

func TestValidateSource(t *testing.T) {
	tests := []struct {
		name   string
		source string
		expErr string
	}{
		{
			name:   "empty",
			source: "",
			expErr: "invalid source \"\": cannot be empty",
		},
		{
			name:   "101 bytes",
			source: "abc" + strings.Repeat("d", 95) + "efg",
			expErr: "invalid source \"abcdddd...dddefg\": length 101 exceeds max 100",
		},
		{
			name:   "1001 bytes",
			source: "abcdefg" + strings.Repeat("h", 988) + "ijklmn",
			expErr: "invalid source \"abcdefg...ijklmn\": length 1001 exceeds max 100",
		},
		{name: "1 byte", source: "x"},
		{name: "just a space", source: " "},
		{name: "41 bytes", source: "pb1zg69v7yszg69v7yszg69v7yszg69v7ysu420dg"},
		{name: "61 bytes", source: "pb1zg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg6qj4dfv3"},
		{name: "99 bytes", source: strings.Repeat("P", 99)},
		{name: "100 bytes", source: strings.Repeat("P", 100)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = ValidateSource(tc.source)
			}
			require.NotPanics(t, testFunc, "ValidateSource(%q)", tc.source)
			assertions.AssertErrorValue(t, err, tc.expErr, "ValidateSource(%q) error", tc.source)
		})
	}
}

func TestNetAssetValueRecord_AsNAV(t *testing.T) {
	tests := []struct {
		name string
		nav  *NetAssetValueRecord
		exp  *NetAssetValue
	}{
		{
			name: "nil",
			nav:  nil,
			exp:  &NetAssetValue{},
		},
		{
			name: "empty",
			nav:  &NetAssetValueRecord{},
			exp:  &NetAssetValue{},
		},
		{
			name: "normal",
			nav: &NetAssetValueRecord{
				Assets: newCoin(18, "red"),
				Price:  newCoin(52, "blue"),
				Height: 7,
				Source: "white",
			},
			exp: &NetAssetValue{
				Assets: newCoin(18, "red"),
				Price:  newCoin(52, "blue"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act *NetAssetValue
			testFunc := func() {
				act = tc.nav.AsNAV()
			}
			require.NotPanics(t, testFunc, "%#v.AsNAV()", tc.nav)
			assert.Equal(t, tc.exp, act, "%#v.AsNAV() result", tc.nav)
		})
	}
}

func TestNetAssetValueRecord_Key(t *testing.T) {
	tests := []struct {
		name string
		nav  *NetAssetValueRecord
		exp  collections.Pair[string, string]
	}{
		{
			name: "nil",
			nav:  nil,
			exp:  collections.Join("", ""),
		},
		{
			name: "empty",
			nav:  &NetAssetValueRecord{},
			exp:  collections.Join("", ""),
		},
		{
			name: "normal",
			nav: &NetAssetValueRecord{
				Assets: newCoin(11, "green"),
				Price:  newCoin(57, "blue"),
				Height: 12,
				Source: "me",
			},
			exp: collections.Join("green", "blue"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act collections.Pair[string, string]
			testFunc := func() {
				act = tc.nav.Key()
			}
			require.NotPanics(t, testFunc, "%#v.Key()", tc.nav)
			assert.Equal(t, tc.exp, act, "%#v.Key() result", tc.nav)
		})
	}
}

func TestNAVRecords_String(t *testing.T) {
	tests := []struct {
		name string
		navs NAVRecords
		exp  string
	}{
		{
			name: "nil",
			navs: nil,
			exp:  "<nil>",
		},
		{
			name: "empty",
			navs: NAVRecords{},
			exp:  "[]",
		},
		{
			name: "one nav",
			navs: NAVRecords{{Assets: newCoin(4, "yellow"), Price: newCoin(99, "blue"), Height: 12, Source: "pink"}},
			exp:  "[4yellow=99blue@12 by \"pink\"]",
		},
		{
			name: "one nil nav",
			navs: NAVRecords{nil},
			exp:  "[<nil>]",
		},
		{
			name: "three navs",
			navs: NAVRecords{
				{Assets: newCoin(1, "red"), Price: newCoin(2, "blue"), Height: 71, Source: "purple"},
				{Assets: newCoin(3, "blue"), Price: newCoin(4, "yellow"), Height: 45, Source: "green"},
				{Assets: newCoin(5, "yellow"), Price: newCoin(6, "red"), Height: 9001, Source: "orange"},
			},
			exp: "[1red=2blue@71 by \"purple\",3blue=4yellow@45 by \"green\",5yellow=6red@9001 by \"orange\"]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.navs.String()
			}
			require.NotPanics(t, testFunc, "%#v.String()", tc.navs)
			assert.Equal(t, tc.exp, act, "%#v.String() result", tc.navs)
		})
	}
}

func TestNAVRecords_Validate(t *testing.T) {
	tests := []struct {
		name   string
		navs   NAVRecords
		expErr string
	}{
		{
			name:   "nil",
			navs:   nil,
			expErr: "",
		},
		{
			name:   "empty",
			navs:   NAVRecords{},
			expErr: "",
		},
		{
			name:   "one entry: nil",
			navs:   NAVRecords{nil},
			expErr: "0: nav record cannot be nil",
		},
		{
			name:   "one entry: invalid",
			navs:   NAVRecords{{Assets: newCoin(-3, "red"), Price: newCoin(4, "blue"), Height: 3, Source: "pink"}},
			expErr: "0: invalid assets \"-3red\": negative coin amount: -3",
		},
		{
			name:   "one entry: valid",
			navs:   NAVRecords{{Assets: newCoin(3, "red"), Price: newCoin(4, "blue"), Height: 3, Source: "pink"}},
			expErr: "",
		},
		{
			name: "two entries: both okay",
			navs: NAVRecords{
				{Assets: newCoin(3, "red"), Price: newCoin(4, "blue"), Height: 3, Source: "pink"},
				{Assets: newCoin(7, "green"), Price: newCoin(9, "purple"), Height: 4, Source: "orange"},
			},
			expErr: "",
		},
		{
			name: "two entries: first invalid",
			navs: NAVRecords{
				{Assets: newCoin(-3, "red"), Price: newCoin(4, "blue"), Height: 3, Source: "pink"},
				{Assets: newCoin(7, "green"), Price: newCoin(9, "purple"), Height: 4, Source: "orange"},
			},
			expErr: "0: invalid assets \"-3red\": negative coin amount: -3",
		},
		{
			name: "two entries: second invalid",
			navs: NAVRecords{
				{Assets: newCoin(3, "red"), Price: newCoin(4, "blue"), Height: 3, Source: "pink"},
				{Assets: newCoin(7, "green"), Price: newCoin(-9, "purple"), Height: 4, Source: "orange"},
			},
			expErr: "1: invalid price \"-9purple\": negative coin amount: -9",
		},
		{
			name: "two entries: both invalid",
			navs: NAVRecords{
				{Assets: newCoin(-3, "red"), Price: newCoin(4, "blue"), Height: 3, Source: "pink"},
				{Assets: newCoin(7, "green"), Price: newCoin(-9, "purple"), Height: 4, Source: "orange"},
			},
			expErr: joinErrs(
				"0: invalid assets \"-3red\": negative coin amount: -3",
				"1: invalid price \"-9purple\": negative coin amount: -9",
			),
		},
		{
			name: "two entries: same assets denoms",
			navs: NAVRecords{
				{Assets: newCoin(3, "red"), Price: newCoin(4, "blue"), Height: 3, Source: "pink"},
				{Assets: newCoin(7, "red"), Price: newCoin(9, "purple"), Height: 4, Source: "orange"},
			},
			expErr: "",
		},
		{
			name: "two entries: same price denoms",
			navs: NAVRecords{
				{Assets: newCoin(3, "red"), Price: newCoin(4, "purple"), Height: 3, Source: "pink"},
				{Assets: newCoin(7, "green"), Price: newCoin(9, "purple"), Height: 4, Source: "orange"},
			},
			expErr: "",
		},
		{
			name: "two entries: same assets and price denoms",
			navs: NAVRecords{
				{Assets: newCoin(3, "red"), Price: newCoin(4, "purple"), Height: 3, Source: "pink"},
				{Assets: newCoin(7, "red"), Price: newCoin(9, "purple"), Height: 4, Source: "orange"},
			},
			expErr: "cannot have multiple (2) navs with the same asset (\"red\") and price (\"purple\") denoms",
		},
		{
			name: "five entries: two sets of same assets and price denoms",
			navs: NAVRecords{
				{Assets: newCoin(1, "red"), Price: newCoin(2, "purple"), Height: 3, Source: "one"},
				{Assets: newCoin(3, "red"), Price: newCoin(4, "purple"), Height: 4, Source: "two"},
				{Assets: newCoin(5, "red"), Price: newCoin(5, "purple"), Height: 5, Source: "three"},
				{Assets: newCoin(7, "purple"), Price: newCoin(8, "red"), Height: 6, Source: "four"},
				{Assets: newCoin(9, "purple"), Price: newCoin(10, "red"), Height: 7, Source: "five"},
			},
			expErr: joinErrs(
				"cannot have multiple (3) navs with the same asset (\"red\") and price (\"purple\") denoms",
				"cannot have multiple (2) navs with the same asset (\"purple\") and price (\"red\") denoms",
			),
		},
		{
			name: "two entries: opposite denoms",
			navs: NAVRecords{
				{Assets: newCoin(3, "red"), Price: newCoin(4, "purple"), Height: 3, Source: "pink"},
				{Assets: newCoin(7, "purple"), Price: newCoin(9, "red"), Height: 4, Source: "orange"},
			},
			expErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc1 := func() {
				err = tc.navs.Validate()
			}
			testFunc2 := func() {
				err = ValidateNAVRecords(tc.navs)
			}
			if assert.NotPanics(t, testFunc1, "%#v.Validate()", tc.navs) {
				assertions.AssertErrorValue(t, err, tc.expErr, "%#v.Validate() error", tc.navs)
			}
			if assert.NotPanics(t, testFunc2, "ValidateNAVRecords(%#v)", tc.navs) {
				assertions.AssertErrorValue(t, err, tc.expErr, "ValidateNAVRecords(%#v) error", tc.navs)
			}
		})
	}
}

func TestDefaultGenesisState(t *testing.T) {
	exp := &GenesisState{}
	var act *GenesisState
	testFunc := func() {
		act = DefaultGenesisState()
	}
	require.NotPanics(t, testFunc, "DefaultGenesisState()")
	assert.Equal(t, exp, act, "DefaultGenesisState() result")
}

func TestGenesisState_Validate(t *testing.T) {
	tests := []struct {
		name   string
		gen    GenesisState
		expErr string
	}{
		{
			name: "empty",
			gen:  GenesisState{},
		},
		{
			name: "empty navs",
			gen:  GenesisState{Navs: []*NetAssetValueRecord{}},
		},
		{
			name: "two navs, okay",
			gen: GenesisState{Navs: []*NetAssetValueRecord{
				{Assets: newCoin(11, "brown"), Price: newCoin(47, "white"), Height: 71, Source: "gen"},
				{Assets: newCoin(43, "yellow"), Price: newCoin(12, "green"), Height: 71, Source: "gen"},
			}},
		},
		{
			name: "three navs: invalid second",
			gen: GenesisState{Navs: []*NetAssetValueRecord{
				{Assets: newCoin(11, "brown"), Price: newCoin(47, "white"), Height: 71, Source: "gen1"},
				{Assets: newCoin(43, "yellow"), Price: newCoin(12, "yellow"), Height: 55, Source: "gen2"},
				{Assets: newCoin(3, "white"), Price: newCoin(49, "brown"), Height: 101, Source: "gen3"},
			}},
			expErr: "invalid navs: 1: nav assets \"43yellow\" and price \"12yellow\" must have different denoms",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.gen.Validate()
			}
			require.NotPanics(t, testFunc, "GenesisState.Validate()")
			assertions.AssertErrorValue(t, err, tc.expErr, "GenesisState.Validate() error")
		})
	}
}

func TestNewEventSetNetAssetValue(t *testing.T) {
	tests := []struct {
		name string
		nav  *NetAssetValueRecord
		exp  *EventSetNetAssetValue
	}{
		{
			name: "nil",
			nav:  nil,
			exp:  &EventSetNetAssetValue{Assets: "<nil>", Price: "<nil>", Source: ""},
		},
		{
			name: "empty",
			nav:  &NetAssetValueRecord{},
			exp:  &EventSetNetAssetValue{Assets: "<nil>", Price: "<nil>", Source: ""},
		},
		{
			name: "normal",
			nav:  &NetAssetValueRecord{Assets: newCoin(55, "gold"), Price: newCoin(12, "silver"), Height: 69, Source: "spring"},
			exp:  &EventSetNetAssetValue{Assets: "55gold", Price: "12silver", Source: "spring"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act *EventSetNetAssetValue
			testFunc := func() {
				act = NewEventSetNetAssetValue(tc.nav)
			}
			require.NotPanics(t, testFunc, "NewEventSetNetAssetValue(%#v)", tc.nav)
			assert.Equal(t, tc.exp, act, "NewEventSetNetAssetValue(%#v) result", tc.nav)
		})
	}
}
