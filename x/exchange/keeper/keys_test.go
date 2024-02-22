package keeper_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/keeper"
)

const hexString = "8fde739c8158424b93dfc27b08e40285" // randomly generated using uuidgen

// concatBz combines all provided byte slices into a single one.
func concatBz(bzs ...[]byte) []byte {
	var rv []byte
	for _, bz := range bzs {
		rv = append(rv, bz...)
	}
	return rv
}

// expectedPrefix gives a name to a prefix that a key is expected to have.
type expectedPrefix struct {
	name  string
	value []byte
}

// keyTestCase is used with checkKey to run some standardized checks on a test case for a store key.
type keyTestCase struct {
	// maker is a function that creates the key value to check.
	// A maker is used (instead of just providing the value to check) so that
	// the function in question can be checked for panics.
	maker func() []byte
	// expected is the expected result of the maker.
	expected []byte
	// expPanic is the panic message expected when the maker is called.
	expPanic string
	// expPrefixes are all the prefixes that the result of the maker is expected to have.
	expPrefixes []expectedPrefix
}

// checkKey calls the maker and asserts that the result is as expected.
// Also asserts that the result has the provided prefixes.
func checkKey(t *testing.T, tc keyTestCase, msg string, args ...interface{}) {
	t.Helper()
	var expPanic []string
	if len(tc.expPanic) > 0 {
		expPanic = []string{tc.expPanic}
	}

	var actual []byte
	testFunc := func() {
		actual = tc.maker()
	}
	assertions.RequirePanicContentsf(t, testFunc, expPanic, msg, args...)
	assert.Equalf(t, tc.expected, actual, msg+" result", args...)

	if len(actual) > 0 {
		assert.Equalf(t, len(actual), cap(actual), msg+" result length (expected) vs capacity (actual)", args...)
		for _, expPre := range tc.expPrefixes {
			actPre := actual
			if len(actPre) > len(expPre.value) {
				actPre = actPre[:len(expPre.value)]
			}
			assert.Equalf(t, expPre.value, actPre, msg+" result %s prefix", append(args, expPre.name))
		}
	}
}

func TestKeyTypeUniqueness(t *testing.T) {
	type byteEntry struct {
		name  string
		value byte
	}

	tests := []struct {
		name  string
		types []byteEntry
	}{
		{
			name: "base type bytes",
			types: []byteEntry{
				{name: "KeyTypeParams", value: keeper.KeyTypeParams},
				{name: "KeyTypeLastMarketID", value: keeper.KeyTypeLastMarketID},
				{name: "KeyTypeKnownMarketID", value: keeper.KeyTypeKnownMarketID},
				{name: "KeyTypeLastOrderID", value: keeper.KeyTypeLastOrderID},
				{name: "KeyTypeMarket", value: keeper.KeyTypeMarket},
				{name: "KeyTypeOrder", value: keeper.KeyTypeOrder},
				{name: "KeyTypeMarketToOrderIndex", value: keeper.KeyTypeMarketToOrderIndex},
				{name: "KeyTypeAddressToOrderIndex", value: keeper.KeyTypeAddressToOrderIndex},
				{name: "KeyTypeAssetToOrderIndex", value: keeper.KeyTypeAssetToOrderIndex},
				{name: "KeyTypeMarketExternalIDToOrderIndex", value: keeper.KeyTypeMarketExternalIDToOrderIndex},
				{name: "KeyTypeCommitment", value: keeper.KeyTypeCommitment},
			},
		},
		{
			name: "market type bytes",
			types: []byteEntry{
				{name: "MarketKeyTypeCreateAskFlat", value: keeper.MarketKeyTypeCreateAskFlat},
				{name: "MarketKeyTypeCreateBidFlat", value: keeper.MarketKeyTypeCreateBidFlat},
				{name: "MarketKeyTypeSellerSettlementFlat", value: keeper.MarketKeyTypeSellerSettlementFlat},
				{name: "MarketKeyTypeSellerSettlementRatio", value: keeper.MarketKeyTypeSellerSettlementRatio},
				{name: "MarketKeyTypeBuyerSettlementFlat", value: keeper.MarketKeyTypeBuyerSettlementFlat},
				{name: "MarketKeyTypeBuyerSettlementRatio", value: keeper.MarketKeyTypeBuyerSettlementRatio},
				{name: "MarketKeyTypeNotAcceptingOrders", value: keeper.MarketKeyTypeNotAcceptingOrders},
				{name: "MarketKeyTypeUserSettle", value: keeper.MarketKeyTypeUserSettle},
				{name: "MarketKeyTypePermissions", value: keeper.MarketKeyTypePermissions},
				{name: "MarketKeyTypeReqAttr", value: keeper.MarketKeyTypeReqAttr},
				{name: "MarketKeyTypeAcceptingCommitments", value: keeper.MarketKeyTypeAcceptingCommitments},
				{name: "MarketKeyTypeCreateCommitmentFlat", value: keeper.MarketKeyTypeCreateCommitmentFlat},
				{name: "MarketKeyTypeCommitmentSettlementBips", value: keeper.MarketKeyTypeCommitmentSettlementBips},
				{name: "MarketKeyTypeIntermediaryDenom", value: keeper.MarketKeyTypeIntermediaryDenom},
			},
		},
		{
			name: "order types",
			types: []byteEntry{
				{name: "OrderKeyTypeAsk", value: keeper.OrderKeyTypeAsk},
				{name: "OrderKeyTypeBid", value: keeper.OrderKeyTypeBid},
			},
		},
		{
			name: "required attribute types",
			types: []byteEntry{
				{name: "OrderKeyTypeAsk", value: keeper.OrderKeyTypeAsk},
				{name: "OrderKeyTypeBid", value: keeper.OrderKeyTypeBid},
				{name: "KeyTypeCommitment", value: keeper.KeyTypeCommitment},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			seen := make(map[byte]string)
			for _, entry := range tc.types {
				prev, found := seen[entry.value]
				assert.False(t, found, "byte %#x used for both %s and %s", prev, entry.name)
				seen[entry.value] = entry.name
			}
		})
	}
}

func TestParseLengthPrefixedAddr(t *testing.T) {
	tests := []struct {
		name    string
		key     []byte
		expAddr sdk.AccAddress
		expRest []byte
		expErr  string
	}{
		{
			name:   "nil slice",
			key:    nil,
			expErr: "slice is empty",
		},
		{
			name:   "empty slice",
			key:    []byte{},
			expErr: "slice is empty",
		},
		{
			name:   "first byte is zero",
			key:    []byte{0, 1, 2, 3},
			expErr: "length byte is zero",
		},
		{
			name:   "too short for length byte 1",
			key:    []byte{1},
			expErr: "length byte is 1, but slice only has 0 left",
		},
		{
			name:   "really too short for length byte 10",
			key:    []byte{10, 1},
			expErr: "length byte is 10, but slice only has 1 left",
		},
		{
			name:   "barely too short for length byte 10",
			key:    []byte{10, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			expErr: "length byte is 10, but slice only has 9 left",
		},
		{
			name:    "length 5 with nothing left",
			key:     []byte{5, 1, 2, 3, 4, 5},
			expAddr: sdk.AccAddress{1, 2, 3, 4, 5},
			expRest: nil,
		},
		{
			name:    "length 5 with 1 byte left",
			key:     []byte{5, 1, 2, 3, 4, 5, 11},
			expAddr: sdk.AccAddress{1, 2, 3, 4, 5},
			expRest: []byte{11},
		},
		{
			name:    "length 5 with 5 bytes left",
			key:     []byte{5, 1, 2, 3, 4, 5, 11, 22, 33, 44, 55},
			expAddr: sdk.AccAddress{1, 2, 3, 4, 5},
			expRest: []byte{11, 22, 33, 44, 55},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var addr sdk.AccAddress
			var rest []byte
			var err error
			testFunc := func() {
				addr, rest, err = keeper.ParseLengthPrefixedAddr(tc.key)
			}
			require.NotPanics(t, testFunc, "parseLengthPrefixedAddr(%v)", tc.key)
			assertions.AssertErrorValue(t, err, tc.expErr, "parseLengthPrefixedAddr(%v) error", tc.key)
			assert.Equal(t, tc.expAddr, addr, "parseLengthPrefixedAddr(%v) address", tc.key)
			assert.Equal(t, tc.expRest, rest, "parseLengthPrefixedAddr(%v) the rest", tc.key)
		})
	}
}

func TestGetKeyPrefixParamsSplit(t *testing.T) {
	ktc := keyTestCase{
		maker: func() []byte {
			return keeper.GetKeyPrefixParamsSplit()
		},
		expected: []byte{keeper.KeyTypeParams, 's', 'p', 'l', 'i', 't'},
	}
	checkKey(t, ktc, "GetKeyPrefixParamsSplit")
}

func TestMakeKeyParamsSplit(t *testing.T) {
	tests := []struct {
		name     string
		denom    string
		expected []byte
	}{
		{
			name:     "empty denom",
			denom:    "",
			expected: concatBz([]byte{keeper.KeyTypeParams}, []byte("split")),
		},
		{
			name:     "nhash",
			denom:    "nhash",
			expected: concatBz([]byte{keeper.KeyTypeParams}, []byte("split"), []byte("nhash")),
		},
		{
			name:     "hex string",
			denom:    hexString,
			expected: concatBz([]byte{keeper.KeyTypeParams}, []byte("split"), []byte(hexString)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeKeyParamsSplit(tc.denom)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "GetKeyPrefixParamsSplit", value: keeper.GetKeyPrefixParamsSplit()},
				},
			}
			checkKey(t, ktc, "MakeKeyParamsSplit(%q)", tc.denom)
		})
	}
}

func TestMakeKeyParamsFeeCreatePaymentFlat(t *testing.T) {
	ktc := keyTestCase{
		maker: func() []byte {
			return keeper.MakeKeyParamsFeeCreatePaymentFlat()
		},
		expected: append([]byte{keeper.KeyTypeParams}, []byte("fee_create_payment_flat")...),
	}
	checkKey(t, ktc, "MakeKeyParamsFeeCreatePaymentFlat")
}

func TestMakeKeyParamsFeeAcceptPaymentFlat(t *testing.T) {
	ktc := keyTestCase{
		maker: func() []byte {
			return keeper.MakeKeyParamsFeeAcceptPaymentFlat()
		},
		expected: append([]byte{keeper.KeyTypeParams}, []byte("fee_accept_payment_flat")...),
	}
	checkKey(t, ktc, "MakeKeyParamsFeeAcceptPaymentFlat")
}

func TestMakeKeyLastMarketID(t *testing.T) {
	ktc := keyTestCase{
		maker: func() []byte {
			return keeper.MakeKeyLastMarketID()
		},
		expected: []byte{keeper.KeyTypeLastMarketID},
	}
	checkKey(t, ktc, "MakeKeyLastMarketID")
}

func TestGetKeyPrefixKnownMarketID(t *testing.T) {
	ktc := keyTestCase{
		maker: func() []byte {
			return keeper.GetKeyPrefixKnownMarketID()
		},
		expected: []byte{keeper.KeyTypeKnownMarketID},
	}
	checkKey(t, ktc, "GetKeyPrefixKnownMarketID")
}

func TestMakeKeyKnownMarketID(t *testing.T) {
	tests := []struct {
		name     string
		marketID uint32
		expected []byte
	}{
		{
			name:     "market id 0",
			marketID: 0,
			expected: []byte{keeper.KeyTypeKnownMarketID, 0, 0, 0, 0},
		},
		{
			name:     "market id 1",
			marketID: 1,
			expected: []byte{keeper.KeyTypeKnownMarketID, 0, 0, 0, 1},
		},
		{
			name:     "market id 255",
			marketID: 255,
			expected: []byte{keeper.KeyTypeKnownMarketID, 0, 0, 0, 255},
		},
		{
			name:     "market id 256",
			marketID: 256,
			expected: []byte{keeper.KeyTypeKnownMarketID, 0, 0, 1, 0},
		},
		{
			name:     "market id 65_536",
			marketID: 65_536,
			expected: []byte{keeper.KeyTypeKnownMarketID, 0, 1, 0, 0},
		},
		{
			name:     "market id 16,777,216",
			marketID: 16_777_216,
			expected: []byte{keeper.KeyTypeKnownMarketID, 1, 0, 0, 0},
		},
		{
			name:     "market id 16,843,009",
			marketID: 16_843_009,
			expected: []byte{keeper.KeyTypeKnownMarketID, 1, 1, 1, 1},
		},
		{
			name:     "market id 33,686,018",
			marketID: 33_686_018,
			expected: []byte{keeper.KeyTypeKnownMarketID, 2, 2, 2, 2},
		},
		{
			name:     "market id 4,294,967,295",
			marketID: 4_294_967_295,
			expected: []byte{keeper.KeyTypeKnownMarketID, 255, 255, 255, 255},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeKeyKnownMarketID(tc.marketID)
				},
				expected: tc.expected,
				expPanic: "",
				expPrefixes: []expectedPrefix{
					{
						name:  "GetKeyPrefixKnownMarketID",
						value: keeper.GetKeyPrefixKnownMarketID(),
					},
				},
			}
			checkKey(t, ktc, "MakeKeyKnownMarketID(%d)", tc.marketID)
		})
	}
}

func TestParseKeySuffixKnownMarketID(t *testing.T) {
	tests := []struct {
		name    string
		suffix  []byte
		exp     uint32
		expFail bool
	}{
		{
			name:    "nil suffix",
			suffix:  nil,
			expFail: true,
		},
		{
			name:    "empty suffix",
			suffix:  []byte{},
			expFail: true,
		},
		{
			name:    "1 byte suffix",
			suffix:  []byte{1},
			expFail: true,
		},
		{
			name:    "2 byte suffix",
			suffix:  []byte{1, 2},
			expFail: true,
		},
		{
			name:    "3 byte suffix",
			suffix:  []byte{1, 2, 3},
			expFail: true,
		},
		{
			name:   "market id 16,909,060 but with extra byte",
			suffix: []byte{1, 2, 3, 4, 5},
			exp:    16_909_060,
		},
		{
			name:   "market id 16,909,060",
			suffix: []byte{1, 2, 3, 4},
			exp:    16_909_060,
		},
		{
			name:   "market id zero",
			suffix: []byte{0, 0, 0, 0},
			exp:    0,
		},
		{
			name:   "market id 1",
			suffix: []byte{0, 0, 0, 1},
			exp:    1,
		},
		{
			name:   "market id 255",
			suffix: []byte{0, 0, 0, 255},
			exp:    255,
		},
		{
			name:   "market id 256",
			suffix: []byte{0, 0, 1, 0},
			exp:    256,
		},
		{
			name:   "market id 65_536",
			suffix: []byte{0, 1, 0, 0},
			exp:    65_536,
		},
		{
			name:   "market id 16,777,216",
			suffix: []byte{1, 0, 0, 0},
			exp:    16_777_216,
		},
		{
			name:   "market id 4,294,967,295",
			suffix: []byte{255, 255, 255, 255},
			exp:    4_294_967_295,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var marketID uint32
			var ok bool
			testFunc := func() {
				marketID, ok = keeper.ParseKeySuffixKnownMarketID(tc.suffix)
			}
			require.NotPanics(t, testFunc, "ParseKeySuffixKnownMarketID")
			assert.Equal(t, !tc.expFail, ok, "ParseKeySuffixKnownMarketID ok bool")
			assert.Equal(t, tc.exp, marketID, "ParseKeySuffixKnownMarketID result")
		})
	}
}

func TestMakeKeyLastOrderID(t *testing.T) {
	ktc := keyTestCase{
		maker: func() []byte {
			return keeper.MakeKeyLastOrderID()
		},
		expected: []byte{keeper.KeyTypeLastOrderID},
	}
	checkKey(t, ktc, "MakeKeyLastOrderID")
}

func TestGetKeyPrefixMarket(t *testing.T) {
	tests := []struct {
		name     string
		marketID uint32
		expected []byte
	}{
		{
			name:     "market id 0",
			marketID: 0,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0},
		},
		{
			name:     "market id 1",
			marketID: 1,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1},
		},
		{
			name:     "market id 255",
			marketID: 255,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 255},
		},
		{
			name:     "market id 256",
			marketID: 256,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 1, 0},
		},
		{
			name:     "market id 65_536",
			marketID: 65_536,
			expected: []byte{keeper.KeyTypeMarket, 0, 1, 0, 0},
		},
		{
			name:     "market id 16,777,216",
			marketID: 16_777_216,
			expected: []byte{keeper.KeyTypeMarket, 1, 0, 0, 0},
		},
		{
			name:     "market id 16,843,009",
			marketID: 16_843_009,
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1},
		},
		{
			name:     "market id 4,294,967,295",
			marketID: 4_294_967_295,
			expected: []byte{keeper.KeyTypeMarket, 255, 255, 255, 255},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.GetKeyPrefixMarket(tc.marketID)
				},
				expected: tc.expected,
			}
			checkKey(t, ktc, "GetKeyPrefixMarket(%d)", tc.marketID)
		})
	}
}

func TestGetKeyPrefixMarketCreateAskFlatFee(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeCreateAskFlat

	tests := []struct {
		name     string
		marketID uint32
		expected []byte
	}{
		{
			name:     "market id 0",
			marketID: 0,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 1",
			marketID: 1,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte},
		},
		{
			name:     "market id 255",
			marketID: 255,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 255, marketTypeByte},
		},
		{
			name:     "market id 256",
			marketID: 256,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 1, 0, marketTypeByte},
		},
		{
			name:     "market id 65_536",
			marketID: 65_536,
			expected: []byte{keeper.KeyTypeMarket, 0, 1, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,777,216",
			marketID: 16_777_216,
			expected: []byte{keeper.KeyTypeMarket, 1, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,843,009",
			marketID: 16_843_009,
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte},
		},
		{
			name:     "market id 4,294,967,295",
			marketID: 4_294_967_295,
			expected: []byte{keeper.KeyTypeMarket, 255, 255, 255, 255, marketTypeByte},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.GetKeyPrefixMarketCreateAskFlatFee(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "GetKeyPrefixMarket", value: keeper.GetKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "GetKeyPrefixMarketCreateAskFlatFee(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyMarketCreateAskFlatFee(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeCreateAskFlat

	tests := []struct {
		name     string
		marketID uint32
		denom    string
		expected []byte
	}{
		{
			name:     "market id 0 no denom",
			marketID: 0,
			denom:    "",
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 0 nhash",
			marketID: 0,
			denom:    "nhash",
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte}, "nhash"...),
		},
		{
			name:     "market id 0 hex string",
			marketID: 0,
			denom:    hexString,
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte}, hexString...),
		},
		{
			name:     "market id 1 no denom",
			marketID: 1,
			denom:    "",
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte},
		},
		{
			name:     "market id 1 nhash",
			marketID: 1,
			denom:    "nhash",
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte}, "nhash"...),
		},
		{
			name:     "market id 1 hex string",
			marketID: 1,
			denom:    hexString,
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte}, hexString...),
		},
		{
			name:     "market id 16,843,009 no denom",
			marketID: 16_843_009,
			denom:    "",
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte},
		},
		{
			name:     "market id 16,843,009 nhash",
			marketID: 16_843_009,
			denom:    "nhash",
			expected: append([]byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte}, "nhash"...),
		},
		{
			name:     "market id 16,843,009 hex string",
			marketID: 16_843_009,
			denom:    hexString,
			expected: append([]byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte}, hexString...),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeKeyMarketCreateAskFlatFee(tc.marketID, tc.denom)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{
						name:  "GetKeyPrefixMarket",
						value: keeper.GetKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "GetKeyPrefixMarketCreateBidFlatFee",
						value: keeper.GetKeyPrefixMarketCreateAskFlatFee(tc.marketID),
					},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketCreateAskFlatFee(%d)", tc.marketID)
		})
	}
}

func TestGetKeyPrefixMarketCreateBidFlatFee(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeCreateBidFlat

	tests := []struct {
		name     string
		marketID uint32
		expected []byte
	}{
		{
			name:     "market id 0",
			marketID: 0,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 1",
			marketID: 1,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte},
		},
		{
			name:     "market id 255",
			marketID: 255,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 255, marketTypeByte},
		},
		{
			name:     "market id 256",
			marketID: 256,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 1, 0, marketTypeByte},
		},
		{
			name:     "market id 65_536",
			marketID: 65_536,
			expected: []byte{keeper.KeyTypeMarket, 0, 1, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,777,216",
			marketID: 16_777_216,
			expected: []byte{keeper.KeyTypeMarket, 1, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,843,009",
			marketID: 16_843_009,
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte},
		},
		{
			name:     "market id 4,294,967,295",
			marketID: 4_294_967_295,
			expected: []byte{keeper.KeyTypeMarket, 255, 255, 255, 255, marketTypeByte},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.GetKeyPrefixMarketCreateBidFlatFee(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "GetKeyPrefixMarket", value: keeper.GetKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "GetKeyPrefixMarketCreateBidFlatFee(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyMarketCreateBidFlatFee(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeCreateBidFlat

	tests := []struct {
		name     string
		marketID uint32
		denom    string
		expected []byte
	}{
		{
			name:     "market id 0 no denom",
			marketID: 0,
			denom:    "",
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 0 nhash",
			marketID: 0,
			denom:    "nhash",
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte}, "nhash"...),
		},
		{
			name:     "market id 0 hex string",
			marketID: 0,
			denom:    hexString,
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte}, hexString...),
		},
		{
			name:     "market id 1 no denom",
			marketID: 1,
			denom:    "",
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte},
		},
		{
			name:     "market id 1 nhash",
			marketID: 1,
			denom:    "nhash",
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte}, "nhash"...),
		},
		{
			name:     "market id 1 hex string",
			marketID: 1,
			denom:    hexString,
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte}, hexString...),
		},
		{
			name:     "market id 16,843,009 no denom",
			marketID: 16_843_009,
			denom:    "",
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte},
		},
		{
			name:     "market id 16,843,009 nhash",
			marketID: 16_843_009,
			denom:    "nhash",
			expected: append([]byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte}, "nhash"...),
		},
		{
			name:     "market id 16,843,009 hex string",
			marketID: 16_843_009,
			denom:    hexString,
			expected: append([]byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte}, hexString...),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeKeyMarketCreateBidFlatFee(tc.marketID, tc.denom)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{
						name:  "GetKeyPrefixMarket",
						value: keeper.GetKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "GetKeyPrefixMarketCreateBidFlatFee",
						value: keeper.GetKeyPrefixMarketCreateBidFlatFee(tc.marketID),
					},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketCreateBidFlatFee(%d)", tc.marketID)
		})
	}
}

func TestGetKeyPrefixMarketCreateCommitmentFlatFee(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeCreateCommitmentFlat

	tests := []struct {
		name     string
		marketID uint32
		expected []byte
	}{
		{
			name:     "market id 0",
			marketID: 0,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 1",
			marketID: 1,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte},
		},
		{
			name:     "market id 255",
			marketID: 255,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 255, marketTypeByte},
		},
		{
			name:     "market id 256",
			marketID: 256,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 1, 0, marketTypeByte},
		},
		{
			name:     "market id 65_536",
			marketID: 65_536,
			expected: []byte{keeper.KeyTypeMarket, 0, 1, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,777,216",
			marketID: 16_777_216,
			expected: []byte{keeper.KeyTypeMarket, 1, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,843,009",
			marketID: 16_843_009,
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte},
		},
		{
			name:     "market id 4,294,967,295",
			marketID: 4_294_967_295,
			expected: []byte{keeper.KeyTypeMarket, 255, 255, 255, 255, marketTypeByte},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.GetKeyPrefixMarketCreateCommitmentFlatFee(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "GetKeyPrefixMarket", value: keeper.GetKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "GetKeyPrefixMarketCreateCommitmentFlatFee(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyMarketCreateCommitmentFlatFee(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeCreateCommitmentFlat

	tests := []struct {
		name     string
		marketID uint32
		denom    string
		expected []byte
	}{
		{
			name:     "market id 0 no denom",
			marketID: 0,
			denom:    "",
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 0 nhash",
			marketID: 0,
			denom:    "nhash",
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte}, "nhash"...),
		},
		{
			name:     "market id 0 hex string",
			marketID: 0,
			denom:    hexString,
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte}, hexString...),
		},
		{
			name:     "market id 1 no denom",
			marketID: 1,
			denom:    "",
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte},
		},
		{
			name:     "market id 1 nhash",
			marketID: 1,
			denom:    "nhash",
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte}, "nhash"...),
		},
		{
			name:     "market id 1 hex string",
			marketID: 1,
			denom:    hexString,
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte}, hexString...),
		},
		{
			name:     "market id 16,843,009 no denom",
			marketID: 16_843_009,
			denom:    "",
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte},
		},
		{
			name:     "market id 16,843,009 nhash",
			marketID: 16_843_009,
			denom:    "nhash",
			expected: append([]byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte}, "nhash"...),
		},
		{
			name:     "market id 16,843,009 hex string",
			marketID: 16_843_009,
			denom:    hexString,
			expected: append([]byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte}, hexString...),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeKeyMarketCreateCommitmentFlatFee(tc.marketID, tc.denom)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{
						name:  "GetKeyPrefixMarket",
						value: keeper.GetKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "GetKeyPrefixMarketCreateCommitmentFlatFee",
						value: keeper.GetKeyPrefixMarketCreateCommitmentFlatFee(tc.marketID),
					},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketCreateCommitmentFlatFee(%d, %q)", tc.marketID, tc.denom)
		})
	}
}

func TestGetKeyPrefixMarketSellerSettlementFlatFee(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeSellerSettlementFlat

	tests := []struct {
		name     string
		marketID uint32
		expected []byte
	}{
		{
			name:     "market id 0",
			marketID: 0,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 1",
			marketID: 1,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte},
		},
		{
			name:     "market id 255",
			marketID: 255,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 255, marketTypeByte},
		},
		{
			name:     "market id 256",
			marketID: 256,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 1, 0, marketTypeByte},
		},
		{
			name:     "market id 65_536",
			marketID: 65_536,
			expected: []byte{keeper.KeyTypeMarket, 0, 1, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,777,216",
			marketID: 16_777_216,
			expected: []byte{keeper.KeyTypeMarket, 1, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,843,009",
			marketID: 16_843_009,
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte},
		},
		{
			name:     "market id 4,294,967,295",
			marketID: 4_294_967_295,
			expected: []byte{keeper.KeyTypeMarket, 255, 255, 255, 255, marketTypeByte},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.GetKeyPrefixMarketSellerSettlementFlatFee(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "GetKeyPrefixMarket", value: keeper.GetKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "GetKeyPrefixMarketSellerSettlementFlatFee(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyMarketSellerSettlementFlatFee(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeSellerSettlementFlat

	tests := []struct {
		name     string
		marketID uint32
		denom    string
		expected []byte
	}{
		{
			name:     "market id 0 no denom",
			marketID: 0,
			denom:    "",
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 0 nhash",
			marketID: 0,
			denom:    "nhash",
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte}, "nhash"...),
		},
		{
			name:     "market id 0 hex string",
			marketID: 0,
			denom:    hexString,
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte}, hexString...),
		},
		{
			name:     "market id 1 no denom",
			marketID: 1,
			denom:    "",
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte},
		},
		{
			name:     "market id 1 nhash",
			marketID: 1,
			denom:    "nhash",
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte}, "nhash"...),
		},
		{
			name:     "market id 1 hex string",
			marketID: 1,
			denom:    hexString,
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte}, hexString...),
		},
		{
			name:     "market id 16,843,009 no denom",
			marketID: 16_843_009,
			denom:    "",
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte},
		},
		{
			name:     "market id 16,843,009 nhash",
			marketID: 16_843_009,
			denom:    "nhash",
			expected: append([]byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte}, "nhash"...),
		},
		{
			name:     "market id 16,843,009 hex string",
			marketID: 16_843_009,
			denom:    hexString,
			expected: append([]byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte}, hexString...),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeKeyMarketSellerSettlementFlatFee(tc.marketID, tc.denom)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{
						name:  "GetKeyPrefixMarket",
						value: keeper.GetKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "GetKeyPrefixMarketSellerSettlementFlatFee",
						value: keeper.GetKeyPrefixMarketSellerSettlementFlatFee(tc.marketID),
					},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketSellerSettlementFlatFee(%d)", tc.marketID)
		})
	}
}

func TestGetKeySuffixSettlementRatio(t *testing.T) {
	coin := func(denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.OneInt()}
	}
	tests := []struct {
		name  string
		ratio exchange.FeeRatio
		exp   []byte
	}{
		{
			name:  "both denoms empty",
			ratio: exchange.FeeRatio{Price: coin(""), Fee: coin("")},
			exp:   []byte{keeper.RecordSeparator},
		},
		{
			name:  "empty price nhash fee",
			ratio: exchange.FeeRatio{Price: coin(""), Fee: coin("nhash")},
			exp:   []byte{keeper.RecordSeparator, 'n', 'h', 'a', 's', 'h'},
		},
		{
			name:  "nhash price empty fee",
			ratio: exchange.FeeRatio{Price: coin("nhash"), Fee: coin("")},
			exp:   []byte{'n', 'h', 'a', 's', 'h', keeper.RecordSeparator},
		},
		{
			name:  "nhash price nhash fee",
			ratio: exchange.FeeRatio{Price: coin("nhash"), Fee: coin("nhash")},
			exp:   []byte{'n', 'h', 'a', 's', 'h', keeper.RecordSeparator, 'n', 'h', 'a', 's', 'h'},
		},
		{
			name:  "nhash price hex string fee",
			ratio: exchange.FeeRatio{Price: coin("nhash"), Fee: coin(hexString)},
			exp:   append([]byte{'n', 'h', 'a', 's', 'h', keeper.RecordSeparator}, hexString...),
		},
		{
			name:  "hex string price nhash fee",
			ratio: exchange.FeeRatio{Price: coin(hexString), Fee: coin("nhash")},
			exp:   append([]byte(hexString), keeper.RecordSeparator, 'n', 'h', 'a', 's', 'h'),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.GetKeySuffixSettlementRatio(tc.ratio)
				},
				expected: tc.exp,
			}
			checkKey(t, ktc, "GetKeySuffixSettlementRatio(%s)", tc.ratio)
		})
	}
}

func TestParseKeySuffixSettlementRatio(t *testing.T) {
	tests := []struct {
		name          string
		suffix        []byte
		expPriceDenom string
		expFeeDenom   string
		expErr        string
	}{
		{
			name:          "both denoms empty",
			suffix:        []byte{keeper.RecordSeparator},
			expPriceDenom: "",
			expFeeDenom:   "",
		},
		{
			name:          "empty price nhash fee",
			suffix:        []byte{keeper.RecordSeparator, 'n', 'h', 'a', 's', 'h'},
			expPriceDenom: "",
			expFeeDenom:   "nhash",
		},
		{
			name:          "nhash price empty fee",
			suffix:        []byte{'n', 'h', 'a', 's', 'h', keeper.RecordSeparator},
			expPriceDenom: "nhash",
			expFeeDenom:   "",
		},
		{
			name:          "nhash price nhash fee",
			suffix:        []byte{'n', 'h', 'a', 's', 'h', keeper.RecordSeparator, 'n', 'h', 'a', 's', 'h'},
			expPriceDenom: "nhash",
			expFeeDenom:   "nhash",
		},
		{
			name:          "nhash price hex string fee",
			suffix:        append([]byte{'n', 'h', 'a', 's', 'h', keeper.RecordSeparator}, hexString...),
			expPriceDenom: "nhash",
			expFeeDenom:   hexString,
		},
		{
			name:          "hex string price nhash fee",
			suffix:        append([]byte(hexString), keeper.RecordSeparator, 'n', 'h', 'a', 's', 'h'),
			expPriceDenom: hexString,
			expFeeDenom:   "nhash",
		},
		{
			name:   "no record separator",
			suffix: []byte("nhashnhash"),
			expErr: "ratio key suffix \"nhashnhash\" has 1 parts, expected 2",
		},
		{
			name:   "two record separators",
			suffix: []byte{keeper.RecordSeparator, 'b', keeper.RecordSeparator},
			expErr: "ratio key suffix \"\\x1eb\\x1e\" has 3 parts, expected 2",
		},
		{
			name:   "nil suffix",
			suffix: nil,
			expErr: "ratio key suffix is empty",
		},
		{
			name:   "empty suffix",
			suffix: []byte{},
			expErr: "ratio key suffix is empty",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var priceDenom, feeDenom string
			var err error
			testFunc := func() {
				priceDenom, feeDenom, err = keeper.ParseKeySuffixSettlementRatio(tc.suffix)
			}
			require.NotPanics(t, testFunc, "ParseKeySuffixSettlementRatio(%q)", tc.suffix)
			assertions.AssertErrorValue(t, err, tc.expErr, "ParseKeySuffixSettlementRatio(%q) error", tc.suffix)
			assert.Equalf(t, tc.expPriceDenom, priceDenom, "ParseKeySuffixSettlementRatio(%q) price denom", tc.suffix)
			assert.Equalf(t, tc.expFeeDenom, feeDenom, "ParseKeySuffixSettlementRatio(%q) fee denom", tc.suffix)
		})
	}
}

func TestGetFeeRatioStoreValue(t *testing.T) {
	pCoin := func(amount string) sdk.Coin {
		amt, ok := sdkmath.NewIntFromString(amount)
		require.True(t, ok, "sdkmath.NewIntFromString(%q) ok boolean return value", amount)
		return sdk.Coin{Denom: "price", Amount: amt}
	}
	fCoin := func(amount string) sdk.Coin {
		amt, ok := sdkmath.NewIntFromString(amount)
		require.True(t, ok, "sdkmath.NewIntFromString(%q) ok boolean return value", amount)
		return sdk.Coin{Denom: "fee", Amount: amt}
	}
	rs := keeper.RecordSeparator

	tests := []struct {
		name  string
		ratio exchange.FeeRatio
		exp   []byte
	}{
		{
			name:  "zero to zero",
			ratio: exchange.FeeRatio{Price: pCoin("0"), Fee: fCoin("0")},
			exp:   []byte{'0', rs, '0'},
		},
		{
			name:  "zero to one",
			ratio: exchange.FeeRatio{Price: pCoin("0"), Fee: fCoin("1")},
			exp:   []byte{'0', rs, '1'},
		},
		{
			name:  "one to zero",
			ratio: exchange.FeeRatio{Price: pCoin("1"), Fee: fCoin("0")},
			exp:   []byte{'1', rs, '0'},
		},
		{
			name:  "one to one",
			ratio: exchange.FeeRatio{Price: pCoin("1"), Fee: fCoin("1")},
			exp:   []byte{'1', rs, '1'},
		},
		{
			name:  "100 to 3",
			ratio: exchange.FeeRatio{Price: pCoin("100"), Fee: fCoin("3")},
			exp:   []byte{'1', '0', '0', rs, '3'},
		},
		{
			name:  "3 to 100",
			ratio: exchange.FeeRatio{Price: pCoin("3"), Fee: fCoin("100")},
			exp:   []byte{'3', rs, '1', '0', '0'},
		},
		{
			name: "huge number to 8",
			// max uint64 = 18,446,744,073,709,551,615. This is 100 times that.
			ratio: exchange.FeeRatio{Price: pCoin("1844674407370955161500"), Fee: fCoin("8")},
			exp:   append([]byte("1844674407370955161500"), rs, '8'),
		},
		{
			name: "15 to huge number",
			// max uint64 = 18,446,744,073,709,551,615. This is 100 times that.
			ratio: exchange.FeeRatio{Price: pCoin("15"), Fee: fCoin("1844674407370955161500")},
			exp:   append([]byte{'1', '5', rs}, "1844674407370955161500"...),
		},
		{
			name:  "two huge numbers",
			ratio: exchange.FeeRatio{Price: pCoin("3454125219812878222609890"), Fee: fCoin("8876890151543931493173153")},
			exp: concatBz(
				[]byte("3454125219812878222609890"),
				[]byte{rs},
				[]byte("8876890151543931493173153"),
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.GetFeeRatioStoreValue(tc.ratio)
				},
				expected: tc.exp,
			}
			checkKey(t, ktc, "GetFeeRatioStoreValue(%s)", tc.ratio)
		})
	}
}

func TestParseFeeRatioStoreValue(t *testing.T) {
	intAmt := func(amt string) sdkmath.Int {
		rv, ok := sdkmath.NewIntFromString(amt)
		require.True(t, ok, "sdkmath.NewIntFromString(%q)", amt)
		return rv
	}
	rs := keeper.RecordSeparator

	tests := []struct {
		name           string
		value          []byte
		expPriceAmount sdkmath.Int
		expFeeAmount   sdkmath.Int
		expErr         string
	}{
		{
			name:           "zero to zero",
			value:          []byte{'0', rs, '0'},
			expPriceAmount: intAmt("0"),
			expFeeAmount:   intAmt("0"),
		},
		{
			name:           "zero to one",
			value:          []byte{'0', rs, '1'},
			expPriceAmount: intAmt("0"),
			expFeeAmount:   intAmt("1"),
		},
		{
			name:           "one to zero",
			value:          []byte{'1', rs, '0'},
			expPriceAmount: intAmt("1"),
			expFeeAmount:   intAmt("0"),
		},
		{
			name:           "one to one",
			value:          []byte{'1', rs, '1'},
			expPriceAmount: intAmt("1"),
			expFeeAmount:   intAmt("1"),
		},
		{
			name:           "100 to 3",
			value:          []byte{'1', '0', '0', rs, '3'},
			expPriceAmount: intAmt("100"),
			expFeeAmount:   intAmt("3"),
		},
		{
			name:           "3 to 100",
			value:          []byte{'3', rs, '1', '0', '0'},
			expPriceAmount: intAmt("3"),
			expFeeAmount:   intAmt("100"),
		},
		{
			name: "huge number to 8",
			// max uint64 = 18,446,744,073,709,551,615. This is 100 times that.
			value:          append([]byte("1844674407370955161500"), rs, '8'),
			expPriceAmount: intAmt("1844674407370955161500"),
			expFeeAmount:   intAmt("8"),
		},
		{
			name: "15 to huge number",
			// max uint64 = 18,446,744,073,709,551,615. This is 100 times that.
			value:          append([]byte{'1', '5', rs}, "1844674407370955161500"...),
			expPriceAmount: intAmt("15"),
			expFeeAmount:   intAmt("1844674407370955161500"),
		},
		{
			name: "two huge numbers",
			value: concatBz(
				[]byte("3454125219812878222609890"),
				[]byte{rs},
				[]byte("8876890151543931493173153"),
			),
			expPriceAmount: intAmt("3454125219812878222609890"),
			expFeeAmount:   intAmt("8876890151543931493173153"),
		},
		{
			name:   "invalid char in price",
			value:  []byte{'1', 'f', '0', rs, '5', '6', '7'},
			expErr: "cannot convert price amount \"1f0\" to sdkmath.Int",
		},
		{
			name:   "invalid char in fee",
			value:  []byte{'1', '3', '0', rs, '5', 'f', '7'},
			expErr: "cannot convert fee amount \"5f7\" to sdkmath.Int",
		},
		{
			name:  "invalid char in both",
			value: []byte{'1', 'f', '0', rs, '5', 'f', '7'},
			expErr: "cannot convert price amount \"1f0\" to sdkmath.Int" + "\n" +
				"cannot convert fee amount \"5f7\" to sdkmath.Int",
		},
		{
			name:  "empty to empty",
			value: []byte{rs},
			expErr: "cannot convert price amount \"\" to sdkmath.Int" + "\n" +
				"cannot convert fee amount \"\" to sdkmath.Int",
		},
		{
			name:   "no record separator",
			value:  []byte("100"),
			expErr: "ratio value \"100\" has 1 parts, expected 2",
		},
		{
			name:   "two record separators",
			value:  []byte{rs, '1', '0', '0', rs},
			expErr: "ratio value \"\\x1e100\\x1e\" has 3 parts, expected 2",
		},
		{
			name:   "nil value",
			value:  nil,
			expErr: "ratio value is empty",
		},
		{
			name:   "empty value",
			value:  []byte{},
			expErr: "ratio value is empty",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.expErr) > 0 {
				tc.expPriceAmount = sdkmath.ZeroInt()
				tc.expFeeAmount = sdkmath.ZeroInt()
			}

			var priceAmount, feeAmont sdkmath.Int
			var err error
			testFunc := func() {
				priceAmount, feeAmont, err = keeper.ParseFeeRatioStoreValue(tc.value)
			}
			require.NotPanics(t, testFunc, "ParseFeeRatioStoreValue(%q)", tc.value)

			assertions.AssertErrorValue(t, err, tc.expErr, "ParseFeeRatioStoreValue(%q) error", tc.value)
			assert.Equal(t, tc.expPriceAmount.String(), priceAmount.String(), "ParseFeeRatioStoreValue(%q) price amount", tc.value)
			assert.Equal(t, tc.expFeeAmount.String(), feeAmont.String(), "ParseFeeRatioStoreValue(%q) fee amount", tc.value)
		})
	}
}

func TestGetKeyPrefixMarketSellerSettlementRatio(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeSellerSettlementRatio

	tests := []struct {
		name     string
		marketID uint32
		expected []byte
	}{
		{
			name:     "market id 0",
			marketID: 0,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 1",
			marketID: 1,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte},
		},
		{
			name:     "market id 255",
			marketID: 255,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 255, marketTypeByte},
		},
		{
			name:     "market id 256",
			marketID: 256,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 1, 0, marketTypeByte},
		},
		{
			name:     "market id 65_536",
			marketID: 65_536,
			expected: []byte{keeper.KeyTypeMarket, 0, 1, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,777,216",
			marketID: 16_777_216,
			expected: []byte{keeper.KeyTypeMarket, 1, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,843,009",
			marketID: 16_843_009,
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte},
		},
		{
			name:     "market id 4,294,967,295",
			marketID: 4_294_967_295,
			expected: []byte{keeper.KeyTypeMarket, 255, 255, 255, 255, marketTypeByte},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.GetKeyPrefixMarketSellerSettlementRatio(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "GetKeyPrefixMarket", value: keeper.GetKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "GetKeyPrefixMarketSellerSettlementRatio(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyMarketSellerSettlementRatio(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeSellerSettlementRatio
	coin := func(denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.OneInt()}
	}
	rs := keeper.RecordSeparator

	tests := []struct {
		name     string
		marketID uint32
		ratio    exchange.FeeRatio
		expected []byte
	}{
		{
			name:     "market id 0 both denoms empty",
			marketID: 0,
			ratio:    exchange.FeeRatio{Price: coin(""), Fee: coin("")},
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte, rs},
		},
		{
			name:     "market id 1 nhash to empty",
			marketID: 1,
			ratio:    exchange.FeeRatio{Price: coin("nhash"), Fee: coin("")},
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte, 'n', 'h', 'a', 's', 'h', rs},
		},
		{
			name:     "market id 1 empty to nhash",
			marketID: 1,
			ratio:    exchange.FeeRatio{Price: coin(""), Fee: coin("nhash")},
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte, rs, 'n', 'h', 'a', 's', 'h'},
		},
		{
			name:     "market id 1 nhash to nhash",
			marketID: 1,
			ratio:    exchange.FeeRatio{Price: coin("nhash"), Fee: coin("nhash")},
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte, 'n', 'h', 'a', 's', 'h', rs, 'n', 'h', 'a', 's', 'h'},
		},
		{
			name:     "market id 16,843,009 nhash to hex string",
			marketID: 16_843_009,
			ratio:    exchange.FeeRatio{Price: coin("nhash"), Fee: coin(hexString)},
			expected: append([]byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte, 'n', 'h', 'a', 's', 'h', rs}, hexString...),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeKeyMarketSellerSettlementRatio(tc.marketID, tc.ratio)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{
						name:  "GetKeyPrefixMarket",
						value: keeper.GetKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "GetKeyPrefixMarketSellerSettlementRatio",
						value: keeper.GetKeyPrefixMarketSellerSettlementRatio(tc.marketID),
					},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketSellerSettlementRatio(%d)", tc.marketID)
		})
	}
}

func TestGetKeyPrefixMarketBuyerSettlementFlatFee(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeBuyerSettlementFlat

	tests := []struct {
		name     string
		marketID uint32
		expected []byte
	}{
		{
			name:     "market id 0",
			marketID: 0,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 1",
			marketID: 1,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte},
		},
		{
			name:     "market id 255",
			marketID: 255,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 255, marketTypeByte},
		},
		{
			name:     "market id 256",
			marketID: 256,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 1, 0, marketTypeByte},
		},
		{
			name:     "market id 65_536",
			marketID: 65_536,
			expected: []byte{keeper.KeyTypeMarket, 0, 1, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,777,216",
			marketID: 16_777_216,
			expected: []byte{keeper.KeyTypeMarket, 1, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,843,009",
			marketID: 16_843_009,
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte},
		},
		{
			name:     "market id 4,294,967,295",
			marketID: 4_294_967_295,
			expected: []byte{keeper.KeyTypeMarket, 255, 255, 255, 255, marketTypeByte},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.GetKeyPrefixMarketBuyerSettlementFlatFee(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "GetKeyPrefixMarket", value: keeper.GetKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "GetKeyPrefixMarketBuyerSettlementFlatFee(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyMarketBuyerSettlementFlatFee(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeBuyerSettlementFlat

	tests := []struct {
		name     string
		marketID uint32
		denom    string
		expected []byte
	}{
		{
			name:     "market id 0 no denom",
			marketID: 0,
			denom:    "",
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 0 nhash",
			marketID: 0,
			denom:    "nhash",
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte}, "nhash"...),
		},
		{
			name:     "market id 0 hex string",
			marketID: 0,
			denom:    hexString,
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte}, hexString...),
		},
		{
			name:     "market id 1 no denom",
			marketID: 1,
			denom:    "",
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte},
		},
		{
			name:     "market id 1 nhash",
			marketID: 1,
			denom:    "nhash",
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte}, "nhash"...),
		},
		{
			name:     "market id 1 hex string",
			marketID: 1,
			denom:    hexString,
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte}, hexString...),
		},
		{
			name:     "market id 16,843,009 no denom",
			marketID: 16_843_009,
			denom:    "",
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte},
		},
		{
			name:     "market id 16,843,009 nhash",
			marketID: 16_843_009,
			denom:    "nhash",
			expected: append([]byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte}, "nhash"...),
		},
		{
			name:     "market id 16,843,009 hex string",
			marketID: 16_843_009,
			denom:    hexString,
			expected: append([]byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte}, hexString...),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeKeyMarketBuyerSettlementFlatFee(tc.marketID, tc.denom)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{
						name:  "GetKeyPrefixMarket",
						value: keeper.GetKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "GetKeyPrefixMarketBuyerSettlementFlatFee",
						value: keeper.GetKeyPrefixMarketBuyerSettlementFlatFee(tc.marketID),
					},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketBuyerSettlementFlatFee(%d)", tc.marketID)
		})
	}
}

func TestGetKeyPrefixMarketBuyerSettlementRatio(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeBuyerSettlementRatio

	tests := []struct {
		name     string
		marketID uint32
		expected []byte
	}{
		{
			name:     "market id 0",
			marketID: 0,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 1",
			marketID: 1,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte},
		},
		{
			name:     "market id 255",
			marketID: 255,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 255, marketTypeByte},
		},
		{
			name:     "market id 256",
			marketID: 256,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 1, 0, marketTypeByte},
		},
		{
			name:     "market id 65_536",
			marketID: 65_536,
			expected: []byte{keeper.KeyTypeMarket, 0, 1, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,777,216",
			marketID: 16_777_216,
			expected: []byte{keeper.KeyTypeMarket, 1, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,843,009",
			marketID: 16_843_009,
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte},
		},
		{
			name:     "market id 4,294,967,295",
			marketID: 4_294_967_295,
			expected: []byte{keeper.KeyTypeMarket, 255, 255, 255, 255, marketTypeByte},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.GetKeyPrefixMarketBuyerSettlementRatio(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "GetKeyPrefixMarket", value: keeper.GetKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "GetKeyPrefixMarketBuyerSettlementRatio(%d)", tc.marketID)
		})
	}
}

func TestGetKeyPrefixMarketBuyerSettlementRatioForPriceDenom(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeBuyerSettlementRatio
	rs := keeper.RecordSeparator

	tests := []struct {
		name     string
		marketID uint32
		denom    string
		expected []byte
	}{
		{
			name:     "market id 0 no denom",
			marketID: 0,
			denom:    "",
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte, rs},
		},
		{
			name:     "market id 0 nhash",
			marketID: 0,
			denom:    "nhash",
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte, 'n', 'h', 'a', 's', 'h', rs},
		},
		{
			name:     "market id 0 hex string",
			marketID: 0,
			denom:    hexString,
			expected: concatBz(
				[]byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte},
				[]byte(hexString),
				[]byte{rs},
			),
		},
		{
			name:     "market id 1 no denom",
			marketID: 1,
			denom:    "",
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte, rs},
		},
		{
			name:     "market id 1 nhash",
			marketID: 1,
			denom:    "nhash",
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte, 'n', 'h', 'a', 's', 'h', rs},
		},
		{
			name:     "market id 1 hex string",
			marketID: 1,
			denom:    hexString,
			expected: concatBz(
				[]byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte},
				[]byte(hexString),
				[]byte{rs},
			),
		},
		{
			name:     "market id 255 nhash",
			marketID: 255,
			denom:    "nhash",
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 255, marketTypeByte, 'n', 'h', 'a', 's', 'h', rs},
		},
		{
			name:     "market id 256 nhash",
			marketID: 256,
			denom:    "nhash",
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 1, 0, marketTypeByte, 'n', 'h', 'a', 's', 'h', rs},
		},
		{
			name:     "market id 65_536 bananas",
			marketID: 65_536,
			denom:    "bananas",
			expected: []byte{keeper.KeyTypeMarket, 0, 1, 0, 0, marketTypeByte, 'b', 'a', 'n', 'a', 'n', 'a', 's', rs},
		},
		{
			name:     "market id 16,777,216 nhash",
			marketID: 16_777_216,
			denom:    "nhash",
			expected: []byte{keeper.KeyTypeMarket, 1, 0, 0, 0, marketTypeByte, 'n', 'h', 'a', 's', 'h', rs},
		},
		{
			name:     "market id 16,843,009 nhash",
			marketID: 16_843_009,
			denom:    "nhash",
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte, 'n', 'h', 'a', 's', 'h', rs},
		},
		{
			name:     "market id 4,294,967,295 no denom",
			marketID: 4_294_967_295,
			denom:    "",
			expected: []byte{keeper.KeyTypeMarket, 255, 255, 255, 255, marketTypeByte, rs},
		},
		{
			name:     "market id 4,294,967,295 nhash",
			marketID: 4_294_967_295,
			denom:    "nhash",
			expected: []byte{keeper.KeyTypeMarket, 255, 255, 255, 255, marketTypeByte, 'n', 'h', 'a', 's', 'h', rs},
		},
		{
			name:     "market id 4,294,967,295 hex string",
			marketID: 4_294_967_295,
			denom:    hexString,
			expected: concatBz(
				[]byte{keeper.KeyTypeMarket, 255, 255, 255, 255, marketTypeByte},
				[]byte(hexString),
				[]byte{rs},
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.GetKeyPrefixMarketBuyerSettlementRatioForPriceDenom(tc.marketID, tc.denom)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{
						name:  "GetKeyPrefixMarket",
						value: keeper.GetKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "GetKeyPrefixMarketBuyerSettlementRatio",
						value: keeper.GetKeyPrefixMarketBuyerSettlementRatio(tc.marketID),
					},
				},
			}
			checkKey(t, ktc, "GetKeyPrefixMarketBuyerSettlementRatio(%d, %q)", tc.marketID, tc.denom)
		})
	}
}

func TestMakeKeyMarketBuyerSettlementRatio(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeBuyerSettlementRatio
	coin := func(denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.OneInt()}
	}
	rs := keeper.RecordSeparator

	tests := []struct {
		name     string
		marketID uint32
		ratio    exchange.FeeRatio
		expected []byte
	}{
		{
			name:     "market id 0 both denoms empty",
			marketID: 0,
			ratio:    exchange.FeeRatio{Price: coin(""), Fee: coin("")},
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte, rs},
		},
		{
			name:     "market id 1 nhash to empty",
			marketID: 1,
			ratio:    exchange.FeeRatio{Price: coin("nhash"), Fee: coin("")},
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte, 'n', 'h', 'a', 's', 'h', rs},
		},
		{
			name:     "market id 1 empty to nhash",
			marketID: 1,
			ratio:    exchange.FeeRatio{Price: coin(""), Fee: coin("nhash")},
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte, rs, 'n', 'h', 'a', 's', 'h'},
		},
		{
			name:     "market id 1 nhash to nhash",
			marketID: 1,
			ratio:    exchange.FeeRatio{Price: coin("nhash"), Fee: coin("nhash")},
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte, 'n', 'h', 'a', 's', 'h', rs, 'n', 'h', 'a', 's', 'h'},
		},
		{
			name:     "market id 16,843,009 nhash to hex string",
			marketID: 16_843_009,
			ratio:    exchange.FeeRatio{Price: coin("nhash"), Fee: coin(hexString)},
			expected: append([]byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte, 'n', 'h', 'a', 's', 'h', rs}, hexString...),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeKeyMarketBuyerSettlementRatio(tc.marketID, tc.ratio)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{
						name:  "GetKeyPrefixMarket",
						value: keeper.GetKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "GetKeyPrefixMarketBuyerSettlementRatio",
						value: keeper.GetKeyPrefixMarketBuyerSettlementRatio(tc.marketID),
					},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketBuyerSettlementRatio(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyMarketNotAcceptingOrders(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeNotAcceptingOrders

	tests := []struct {
		name     string
		marketID uint32
		expected []byte
	}{
		{
			name:     "market id 0",
			marketID: 0,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 1",
			marketID: 1,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte},
		},
		{
			name:     "market id 255",
			marketID: 255,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 255, marketTypeByte},
		},
		{
			name:     "market id 256",
			marketID: 256,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 1, 0, marketTypeByte},
		},
		{
			name:     "market id 65_536",
			marketID: 65_536,
			expected: []byte{keeper.KeyTypeMarket, 0, 1, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,777,216",
			marketID: 16_777_216,
			expected: []byte{keeper.KeyTypeMarket, 1, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,843,009",
			marketID: 16_843_009,
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte},
		},
		{
			name:     "market id 4,294,967,295",
			marketID: 4_294_967_295,
			expected: []byte{keeper.KeyTypeMarket, 255, 255, 255, 255, marketTypeByte},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeKeyMarketNotAcceptingOrders(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "GetKeyPrefixMarket", value: keeper.GetKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketNotAcceptingOrders(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyMarketUserSettle(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeUserSettle

	tests := []struct {
		name     string
		marketID uint32
		expected []byte
	}{
		{
			name:     "market id 0",
			marketID: 0,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 1",
			marketID: 1,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte},
		},
		{
			name:     "market id 255",
			marketID: 255,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 255, marketTypeByte},
		},
		{
			name:     "market id 256",
			marketID: 256,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 1, 0, marketTypeByte},
		},
		{
			name:     "market id 65_536",
			marketID: 65_536,
			expected: []byte{keeper.KeyTypeMarket, 0, 1, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,777,216",
			marketID: 16_777_216,
			expected: []byte{keeper.KeyTypeMarket, 1, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,843,009",
			marketID: 16_843_009,
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte},
		},
		{
			name:     "market id 4,294,967,295",
			marketID: 4_294_967_295,
			expected: []byte{keeper.KeyTypeMarket, 255, 255, 255, 255, marketTypeByte},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeKeyMarketUserSettle(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "GetKeyPrefixMarket", value: keeper.GetKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketUserSettle(%d)", tc.marketID)
		})
	}
}

func TestGetKeyPrefixMarketPermissions(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypePermissions

	tests := []struct {
		name     string
		marketID uint32
		expected []byte
	}{
		{
			name:     "market id 0",
			marketID: 0,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 1",
			marketID: 1,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte},
		},
		{
			name:     "market id 255",
			marketID: 255,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 255, marketTypeByte},
		},
		{
			name:     "market id 256",
			marketID: 256,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 1, 0, marketTypeByte},
		},
		{
			name:     "market id 65_536",
			marketID: 65_536,
			expected: []byte{keeper.KeyTypeMarket, 0, 1, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,777,216",
			marketID: 16_777_216,
			expected: []byte{keeper.KeyTypeMarket, 1, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,843,009",
			marketID: 16_843_009,
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte},
		},
		{
			name:     "market id 4,294,967,295",
			marketID: 4_294_967_295,
			expected: []byte{keeper.KeyTypeMarket, 255, 255, 255, 255, marketTypeByte},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.GetKeyPrefixMarketPermissions(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "GetKeyPrefixMarket", value: keeper.GetKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "GetKeyPrefixMarketPermissions(%d)", tc.marketID)
		})
	}
}

func TestGetKeyPrefixMarketPermissionsForAddress(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypePermissions

	tests := []struct {
		name     string
		marketID uint32
		addr     sdk.AccAddress
		expected []byte
		expPanic string
	}{
		{
			name:     "nil addr",
			addr:     nil,
			expPanic: "empty address not allowed",
		},
		{
			name:     "empty addr",
			addr:     sdk.AccAddress{},
			expPanic: "empty address not allowed",
		},
		{
			name:     "256 byte addr",
			addr:     bytes.Repeat([]byte{'p'}, 256),
			expPanic: "address length should be max 255 bytes, got 256: unknown address",
		},
		{
			name:     "market id 0 5 byte addr",
			marketID: 0,
			addr:     sdk.AccAddress("abcde"),
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte, 5}, "abcde"...),
		},
		{
			name:     "market id 0 20 byte addr",
			marketID: 0,
			addr:     sdk.AccAddress("abcdefghijklmnopqrst"),
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte, 20}, "abcdefghijklmnopqrst"...),
		},
		{
			name:     "market id 0 32 byte addr",
			marketID: 0,
			addr:     sdk.AccAddress("abcdefghijklmnopqrstuvwxyzABCDEF"),
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte, 32}, "abcdefghijklmnopqrstuvwxyzABCDEF"...),
		},
		{
			name:     "market id 1 20 byte addr",
			marketID: 1,
			addr:     sdk.AccAddress("abcdefghijklmnopqrst"),
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte, 20}, "abcdefghijklmnopqrst"...),
		},
		{
			name:     "market id 1 32 byte addr",
			marketID: 1,
			addr:     sdk.AccAddress("abcdefghijklmnopqrstuvwxyzABCDEF"),
			expected: append([]byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte, 32}, "abcdefghijklmnopqrstuvwxyzABCDEF"...),
		},
		{
			name:     "market id 16,843,009 20 byte addr",
			marketID: 16_843_009,
			addr:     sdk.AccAddress("abcdefghijklmnopqrst"),
			expected: append([]byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte, 20}, "abcdefghijklmnopqrst"...),
		},
		{
			name:     "market id 16,843,009 32 byte addr",
			marketID: 16_843_009,
			addr:     sdk.AccAddress("abcdefghijklmnopqrstuvwxyzABCDEF"),
			expected: append([]byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte, 32}, "abcdefghijklmnopqrstuvwxyzABCDEF"...),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.GetKeyPrefixMarketPermissionsForAddress(tc.marketID, tc.addr)
				},
				expected: tc.expected,
				expPanic: tc.expPanic,
			}
			if len(tc.expPanic) == 0 {
				ktc.expPrefixes = []expectedPrefix{
					{
						name:  "GetKeyPrefixMarket",
						value: keeper.GetKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "GetKeyPrefixMarketPermissions",
						value: keeper.GetKeyPrefixMarketPermissions(tc.marketID),
					},
				}
			}
			checkKey(t, ktc, "GetKeyPrefixMarketPermissionsForAddress(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyMarketPermissions(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypePermissions

	tests := []struct {
		name     string
		marketID uint32
		addr     sdk.AccAddress
		perm     exchange.Permission
		expected []byte
		expPanic string
	}{
		{
			name:     "nil addr",
			addr:     nil,
			expPanic: "empty address not allowed",
		},
		{
			name:     "empty addr",
			addr:     sdk.AccAddress{},
			expPanic: "empty address not allowed",
		},
		{
			name:     "256 byte addr",
			addr:     bytes.Repeat([]byte{'p'}, 256),
			expPanic: "address length should be max 255 bytes, got 256: unknown address",
		},
		{
			name:     "market id 0 5 byte addr settle",
			marketID: 0,
			addr:     sdk.AccAddress("abcde"),
			perm:     exchange.Permission_settle,
			expected: concatBz(
				[]byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte, 5},
				[]byte("abcde"),
				[]byte{byte(exchange.Permission_settle)},
			),
		},
		{
			name:     "market id 0 20 byte addr cancel",
			marketID: 0,
			addr:     sdk.AccAddress("abcdefghijklmnopqrst"),
			perm:     exchange.Permission_cancel,
			expected: concatBz(
				[]byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte, 20},
				[]byte("abcdefghijklmnopqrst"),
				[]byte{byte(exchange.Permission_cancel)},
			),
		},
		{
			name:     "market id 0 32 byte addr withdraw",
			marketID: 0,
			addr:     sdk.AccAddress("abcdefghijklmnopqrstuvwxyzABCDEF"),
			perm:     exchange.Permission_withdraw,
			expected: concatBz(
				[]byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte, 32},
				[]byte("abcdefghijklmnopqrstuvwxyzABCDEF"),
				[]byte{byte(exchange.Permission_withdraw)},
			),
		},
		{
			name:     "market id 1 20 byte addr settle",
			marketID: 1,
			addr:     sdk.AccAddress("abcdefghijklmnopqrst"),
			perm:     exchange.Permission_settle,
			expected: concatBz(
				[]byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte, 20},
				[]byte("abcdefghijklmnopqrst"),
				[]byte{byte(exchange.Permission_settle)},
			),
		},
		{
			name:     "market id 20 20 byte addr cancel",
			marketID: 20,
			addr:     sdk.AccAddress("abcdefghijklmnopqrst"),
			perm:     exchange.Permission_cancel,
			expected: concatBz(
				[]byte{keeper.KeyTypeMarket, 0, 0, 0, 20, marketTypeByte, 20},
				[]byte("abcdefghijklmnopqrst"),
				[]byte{byte(exchange.Permission_cancel)},
			),
		},
		{
			name:     "market id 33 20 byte addr withdraw",
			marketID: 33,
			addr:     sdk.AccAddress("abcdefghijklmnopqrst"),
			perm:     exchange.Permission_withdraw,
			expected: concatBz(
				[]byte{keeper.KeyTypeMarket, 0, 0, 0, 33, marketTypeByte, 20},
				[]byte("abcdefghijklmnopqrst"),
				[]byte{byte(exchange.Permission_withdraw)},
			),
		},
		{
			name:     "market id 48 20 byte addr update",
			marketID: 48,
			addr:     sdk.AccAddress("abcdefghijklmnopqrst"),
			perm:     exchange.Permission_update,
			expected: concatBz(
				[]byte{keeper.KeyTypeMarket, 0, 0, 0, 48, marketTypeByte, 20},
				[]byte("abcdefghijklmnopqrst"),
				[]byte{byte(exchange.Permission_update)},
			),
		},
		{
			name:     "market id 52 20 byte addr permissions",
			marketID: 52,
			addr:     sdk.AccAddress("abcdefghijklmnopqrst"),
			perm:     exchange.Permission_permissions,
			expected: concatBz(
				[]byte{keeper.KeyTypeMarket, 0, 0, 0, 52, marketTypeByte, 20},
				[]byte("abcdefghijklmnopqrst"),
				[]byte{byte(exchange.Permission_permissions)},
			),
		},
		{
			name:     "market id 67 20 byte addr attributes",
			marketID: 67,
			addr:     sdk.AccAddress("abcdefghijklmnopqrst"),
			perm:     exchange.Permission_attributes,
			expected: concatBz(
				[]byte{keeper.KeyTypeMarket, 0, 0, 0, 67, marketTypeByte, 20},
				[]byte("abcdefghijklmnopqrst"),
				[]byte{byte(exchange.Permission_attributes)},
			),
		},
		{
			name:     "market id 258 32 byte addr update",
			marketID: 258,
			addr:     sdk.AccAddress("abcdefghijklmnopqrstuvwxyzABCDEF"),
			perm:     exchange.Permission_update,
			expected: concatBz(
				[]byte{keeper.KeyTypeMarket, 0, 0, 1, 2, marketTypeByte, 32},
				[]byte("abcdefghijklmnopqrstuvwxyzABCDEF"),
				[]byte{byte(exchange.Permission_update)},
			),
		},
		{
			name:     "market id 16,843,009 20 byte addr permissions",
			marketID: 16_843_009,
			addr:     sdk.AccAddress("abcdefghijklmnopqrst"),
			perm:     exchange.Permission_permissions,
			expected: concatBz(
				[]byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte, 20},
				[]byte("abcdefghijklmnopqrst"),
				[]byte{byte(exchange.Permission_permissions)},
			),
		},
		{
			name:     "market id 16,843,009 32 byte addr attributes",
			marketID: 16_843_009,
			addr:     sdk.AccAddress("abcdefghijklmnopqrstuvwxyzABCDEF"),
			perm:     exchange.Permission_attributes,
			expected: concatBz(
				[]byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte, 32},
				[]byte("abcdefghijklmnopqrstuvwxyzABCDEF"),
				[]byte{byte(exchange.Permission_attributes)},
			),
		},
		{
			name:     "market id 67,305,985 20 byte addr unspecified",
			marketID: 67_305_985,
			addr:     sdk.AccAddress("abcdefghijklmnopqrst"),
			perm:     exchange.Permission_unspecified,
			expected: concatBz(
				[]byte{keeper.KeyTypeMarket, 4, 3, 2, 1, marketTypeByte, 20},
				[]byte("abcdefghijklmnopqrst"),
				[]byte{byte(exchange.Permission_unspecified)},
			),
		},
		{
			name:     "market id 117,967,114 negative permission",
			marketID: 117_967_114,
			addr:     sdk.AccAddress("abcdefghijklmnopqrst"),
			perm:     -1,
			expPanic: "permission value -1 out of range for uint8",
		},
		{
			name:     "market id 117,967,114 permission 0",
			marketID: 117_967_114,
			addr:     sdk.AccAddress("abcdefghijklmnopqrst"),
			perm:     0,
			expected: concatBz(
				[]byte{keeper.KeyTypeMarket, 7, 8, 9, 10, marketTypeByte, 20},
				[]byte("abcdefghijklmnopqrst"),
				[]byte{0},
			),
		},
		{
			name:     "market id 1,887,473,824 permission 256",
			marketID: 1_887_473_824,
			addr:     sdk.AccAddress("abcdefghijklmnopqrst"),
			perm:     256,
			expPanic: "permission value 256 out of range for uint8",
		},
		{
			name:     "market id 1,887,473,824 permission 255",
			marketID: 1_887_473_824,
			addr:     sdk.AccAddress("abcdefghijklmnopqrst"),
			perm:     255,
			expected: concatBz(
				[]byte{keeper.KeyTypeMarket, 112, 128, 144, 160, marketTypeByte, 20},
				[]byte("abcdefghijklmnopqrst"),
				[]byte{255},
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeKeyMarketPermissions(tc.marketID, tc.addr, tc.perm)
				},
				expected: tc.expected,
				expPanic: tc.expPanic,
			}
			if len(tc.expPanic) == 0 {
				ktc.expPrefixes = []expectedPrefix{
					{
						name:  "GetKeyPrefixMarket",
						value: keeper.GetKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "GetKeyPrefixMarketPermissions",
						value: keeper.GetKeyPrefixMarketPermissions(tc.marketID),
					},
					{
						name:  "GetKeyPrefixMarketPermissionsForAddress",
						value: keeper.GetKeyPrefixMarketPermissionsForAddress(tc.marketID, tc.addr),
					},
				}
			}
			checkKey(t, ktc, "MakeKeyMarketPermissions(%d)", tc.marketID)
		})
	}
}

func TestParseKeySuffixMarketPermissions(t *testing.T) {
	tests := []struct {
		name    string
		suffix  []byte
		expAddr sdk.AccAddress
		expPerm exchange.Permission
		expErr  string
	}{
		{
			name:   "nil suffix",
			suffix: nil,
			expErr: "cannot parse address from market permissions key: slice is empty",
		},
		{
			name:   "empty suffix",
			suffix: []byte{},
			expErr: "cannot parse address from market permissions key: slice is empty",
		},
		{
			name:   "byte length too short",
			suffix: []byte{5, 1, 2},
			expErr: "cannot parse address from market permissions key: length byte is 5, but slice only has 2 left",
		},
		{
			name:   "no permission byte",
			suffix: []byte{5, 1, 2, 3, 4, 5},
			expErr: "cannot parse market permissions key: found 0 bytes after address, expected 1",
		},
		{
			name:   "two bytes after addr",
			suffix: []byte{5, 1, 2, 3, 4, 5, 11, 22},
			expErr: "cannot parse market permissions key: found 2 bytes after address, expected 1",
		},
		{
			name:    "5 byte addr settle",
			suffix:  []byte{5, 1, 2, 3, 4, 5, byte(exchange.Permission_settle)},
			expAddr: sdk.AccAddress{1, 2, 3, 4, 5},
			expPerm: exchange.Permission_settle,
		},
		{
			name:    "5 byte addr cancel",
			suffix:  []byte{5, 1, 2, 3, 4, 5, byte(exchange.Permission_cancel)},
			expAddr: sdk.AccAddress{1, 2, 3, 4, 5},
			expPerm: exchange.Permission_cancel,
		},
		{
			name:    "5 byte addr withdraw",
			suffix:  []byte{5, 1, 2, 3, 4, 5, byte(exchange.Permission_withdraw)},
			expAddr: sdk.AccAddress{1, 2, 3, 4, 5},
			expPerm: exchange.Permission_withdraw,
		},
		{
			name:    "5 byte addr update",
			suffix:  []byte{5, 1, 2, 3, 4, 5, byte(exchange.Permission_update)},
			expAddr: sdk.AccAddress{1, 2, 3, 4, 5},
			expPerm: exchange.Permission_update,
		},
		{
			name:    "5 byte addr permissions",
			suffix:  []byte{5, 1, 2, 3, 4, 5, byte(exchange.Permission_permissions)},
			expAddr: sdk.AccAddress{1, 2, 3, 4, 5},
			expPerm: exchange.Permission_permissions,
		},
		{
			name:    "5 byte addr attributes",
			suffix:  []byte{5, 1, 2, 3, 4, 5, byte(exchange.Permission_attributes)},
			expAddr: sdk.AccAddress{1, 2, 3, 4, 5},
			expPerm: exchange.Permission_attributes,
		},
		{
			name:    "5 byte addr unknown permission",
			suffix:  []byte{5, 1, 2, 3, 4, 5, 88},
			expAddr: sdk.AccAddress{1, 2, 3, 4, 5},
			expPerm: exchange.Permission(88),
		},
		{
			name:    "20 byte addr settle",
			suffix:  []byte{20, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, byte(exchange.Permission_settle)},
			expAddr: sdk.AccAddress{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			expPerm: exchange.Permission_settle,
		},
		{
			name:    "20 byte addr cancel",
			suffix:  []byte{20, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, byte(exchange.Permission_cancel)},
			expAddr: sdk.AccAddress{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			expPerm: exchange.Permission_cancel,
		},
		{
			name:    "20 byte addr withdraw",
			suffix:  []byte{20, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, byte(exchange.Permission_withdraw)},
			expAddr: sdk.AccAddress{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			expPerm: exchange.Permission_withdraw,
		},
		{
			name:    "20 byte addr update",
			suffix:  []byte{20, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, byte(exchange.Permission_update)},
			expAddr: sdk.AccAddress{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			expPerm: exchange.Permission_update,
		},
		{
			name:    "20 byte addr permissions",
			suffix:  []byte{20, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, byte(exchange.Permission_permissions)},
			expAddr: sdk.AccAddress{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			expPerm: exchange.Permission_permissions,
		},
		{
			name:    "20 byte addr attributes",
			suffix:  []byte{20, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, byte(exchange.Permission_attributes)},
			expAddr: sdk.AccAddress{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			expPerm: exchange.Permission_attributes,
		},
		{
			name: "32 byte addr settle",
			suffix: []byte{32, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, byte(exchange.Permission_settle)},
			expAddr: sdk.AccAddress{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
				17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			expPerm: exchange.Permission_settle,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var addr sdk.AccAddress
			var perm exchange.Permission
			var err error
			testFunc := func() {
				addr, perm, err = keeper.ParseKeySuffixMarketPermissions(tc.suffix)
			}
			require.NotPanics(t, testFunc, "ParseKeySuffixMarketPermissions")
			assertions.AssertErrorValue(t, err, tc.expErr, "ParseKeySuffixMarketPermissions error")
			assert.Equal(t, tc.expAddr, addr, "ParseKeySuffixMarketPermissions address")
			assert.Equal(t, tc.expPerm, perm, "ParseKeySuffixMarketPermissions permission")
		})
	}
}

func TestMakeKeyMarketReqAttrAsk(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeReqAttr
	orderTypeByte := keeper.OrderKeyTypeAsk

	tests := []struct {
		name     string
		marketID uint32
		expected []byte
	}{
		{
			name:     "market id 0",
			marketID: 0,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte, orderTypeByte},
		},
		{
			name:     "market id 1",
			marketID: 1,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte, orderTypeByte},
		},
		{
			name:     "market id 255",
			marketID: 255,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 255, marketTypeByte, orderTypeByte},
		},
		{
			name:     "market id 256",
			marketID: 256,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 1, 0, marketTypeByte, orderTypeByte},
		},
		{
			name:     "market id 65_536",
			marketID: 65_536,
			expected: []byte{keeper.KeyTypeMarket, 0, 1, 0, 0, marketTypeByte, orderTypeByte},
		},
		{
			name:     "market id 16,777,216",
			marketID: 16_777_216,
			expected: []byte{keeper.KeyTypeMarket, 1, 0, 0, 0, marketTypeByte, orderTypeByte},
		},
		{
			name:     "market id 16,843,009",
			marketID: 16_843_009,
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte, orderTypeByte},
		},
		{
			name:     "market id 4,294,967,295",
			marketID: 4_294_967_295,
			expected: []byte{keeper.KeyTypeMarket, 255, 255, 255, 255, marketTypeByte, orderTypeByte},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeKeyMarketReqAttrAsk(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "GetKeyPrefixMarket", value: keeper.GetKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketReqAttrAsk(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyMarketReqAttrBid(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeReqAttr
	orderTypeByte := keeper.OrderKeyTypeBid

	tests := []struct {
		name     string
		marketID uint32
		expected []byte
	}{
		{
			name:     "market id 0",
			marketID: 0,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte, orderTypeByte},
		},
		{
			name:     "market id 1",
			marketID: 1,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte, orderTypeByte},
		},
		{
			name:     "market id 255",
			marketID: 255,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 255, marketTypeByte, orderTypeByte},
		},
		{
			name:     "market id 256",
			marketID: 256,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 1, 0, marketTypeByte, orderTypeByte},
		},
		{
			name:     "market id 65_536",
			marketID: 65_536,
			expected: []byte{keeper.KeyTypeMarket, 0, 1, 0, 0, marketTypeByte, orderTypeByte},
		},
		{
			name:     "market id 16,777,216",
			marketID: 16_777_216,
			expected: []byte{keeper.KeyTypeMarket, 1, 0, 0, 0, marketTypeByte, orderTypeByte},
		},
		{
			name:     "market id 16,843,009",
			marketID: 16_843_009,
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte, orderTypeByte},
		},
		{
			name:     "market id 4,294,967,295",
			marketID: 4_294_967_295,
			expected: []byte{keeper.KeyTypeMarket, 255, 255, 255, 255, marketTypeByte, orderTypeByte},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeKeyMarketReqAttrBid(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "GetKeyPrefixMarket", value: keeper.GetKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketReqAttrBid(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyMarketReqAttrCommitment(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeReqAttr
	entryTypeByte := keeper.KeyTypeCommitment

	tests := []struct {
		name     string
		marketID uint32
		expected []byte
	}{
		{
			name:     "market id 0",
			marketID: 0,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte, entryTypeByte},
		},
		{
			name:     "market id 1",
			marketID: 1,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte, entryTypeByte},
		},
		{
			name:     "market id 255",
			marketID: 255,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 255, marketTypeByte, entryTypeByte},
		},
		{
			name:     "market id 256",
			marketID: 256,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 1, 0, marketTypeByte, entryTypeByte},
		},
		{
			name:     "market id 65_536",
			marketID: 65_536,
			expected: []byte{keeper.KeyTypeMarket, 0, 1, 0, 0, marketTypeByte, entryTypeByte},
		},
		{
			name:     "market id 16,777,216",
			marketID: 16_777_216,
			expected: []byte{keeper.KeyTypeMarket, 1, 0, 0, 0, marketTypeByte, entryTypeByte},
		},
		{
			name:     "market id 16,843,009",
			marketID: 16_843_009,
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte, entryTypeByte},
		},
		{
			name:     "market id 4,294,967,295",
			marketID: 4_294_967_295,
			expected: []byte{keeper.KeyTypeMarket, 255, 255, 255, 255, marketTypeByte, entryTypeByte},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeKeyMarketReqAttrCommitment(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "GetKeyPrefixMarket", value: keeper.GetKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketReqAttrCommitment(%d)", tc.marketID)
		})
	}
}

func TestParseReqAttrStoreValue(t *testing.T) {
	rs := keeper.RecordSeparator

	tests := []struct {
		name  string
		value []byte
		exp   []string
	}{
		{
			name:  "nil value",
			value: nil,
			exp:   nil,
		},
		{
			name:  "empty value",
			value: nil,
			exp:   nil,
		},
		{
			name:  "one long attribute",
			value: []byte("one.long.really-long.super-long.attribute"),
			exp:   []string{"one.long.really-long.super-long.attribute"},
		},
		{
			name:  "two attributes",
			value: []byte{'a', 't', 't', 'r', '1', rs, 's', 'e', 'c', 'o', 'n', 'd'},
			exp:   []string{"attr1", "second"},
		},
		{
			name: "five attributes",
			value: bytes.Join([][]byte{
				[]byte("this.is.attr.one"),
				[]byte("a.second.appears"),
				[]byte("thrice.is.twice.as.nice"),
				[]byte("golfers.delight"),
				[]byte("i.have.nothing.for.fifth"),
			}, []byte{rs}),
			exp: []string{
				"this.is.attr.one",
				"a.second.appears",
				"thrice.is.twice.as.nice",
				"golfers.delight",
				"i.have.nothing.for.fifth",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []string
			testFunc := func() {
				actual = keeper.ParseReqAttrStoreValue(tc.value)
			}
			require.NotPanics(t, testFunc, "ParseReqAttrStoreValue")
			assert.Equal(t, tc.exp, actual, "ParseReqAttrStoreValue result")
		})
	}
}

func TestMakeKeyMarketAcceptingCommitments(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeAcceptingCommitments

	tests := []struct {
		name     string
		marketID uint32
		expected []byte
	}{
		{
			name:     "market id 0",
			marketID: 0,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 1",
			marketID: 1,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte},
		},
		{
			name:     "market id 255",
			marketID: 255,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 255, marketTypeByte},
		},
		{
			name:     "market id 256",
			marketID: 256,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 1, 0, marketTypeByte},
		},
		{
			name:     "market id 65_536",
			marketID: 65_536,
			expected: []byte{keeper.KeyTypeMarket, 0, 1, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,777,216",
			marketID: 16_777_216,
			expected: []byte{keeper.KeyTypeMarket, 1, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,843,009",
			marketID: 16_843_009,
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte},
		},
		{
			name:     "market id 4,294,967,295",
			marketID: 4_294_967_295,
			expected: []byte{keeper.KeyTypeMarket, 255, 255, 255, 255, marketTypeByte},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeKeyMarketAcceptingCommitments(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "GetKeyPrefixMarket", value: keeper.GetKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketAcceptingCommitments(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyMarketCommitmentSettlementBips(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeCommitmentSettlementBips

	tests := []struct {
		name     string
		marketID uint32
		expected []byte
	}{
		{
			name:     "market id 0",
			marketID: 0,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 1",
			marketID: 1,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte},
		},
		{
			name:     "market id 255",
			marketID: 255,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 255, marketTypeByte},
		},
		{
			name:     "market id 256",
			marketID: 256,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 1, 0, marketTypeByte},
		},
		{
			name:     "market id 65_536",
			marketID: 65_536,
			expected: []byte{keeper.KeyTypeMarket, 0, 1, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,777,216",
			marketID: 16_777_216,
			expected: []byte{keeper.KeyTypeMarket, 1, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,843,009",
			marketID: 16_843_009,
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte},
		},
		{
			name:     "market id 4,294,967,295",
			marketID: 4_294_967_295,
			expected: []byte{keeper.KeyTypeMarket, 255, 255, 255, 255, marketTypeByte},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeKeyMarketCommitmentSettlementBips(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "GetKeyPrefixMarket", value: keeper.GetKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketCommitmentSettlementBips(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyMarketIntermediaryDenom(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeIntermediaryDenom

	tests := []struct {
		name     string
		marketID uint32
		expected []byte
	}{
		{
			name:     "market id 0",
			marketID: 0,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 1",
			marketID: 1,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte},
		},
		{
			name:     "market id 255",
			marketID: 255,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 255, marketTypeByte},
		},
		{
			name:     "market id 256",
			marketID: 256,
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 1, 0, marketTypeByte},
		},
		{
			name:     "market id 65_536",
			marketID: 65_536,
			expected: []byte{keeper.KeyTypeMarket, 0, 1, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,777,216",
			marketID: 16_777_216,
			expected: []byte{keeper.KeyTypeMarket, 1, 0, 0, 0, marketTypeByte},
		},
		{
			name:     "market id 16,843,009",
			marketID: 16_843_009,
			expected: []byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte},
		},
		{
			name:     "market id 4,294,967,295",
			marketID: 4_294_967_295,
			expected: []byte{keeper.KeyTypeMarket, 255, 255, 255, 255, marketTypeByte},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeKeyMarketIntermediaryDenom(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "GetKeyPrefixMarket", value: keeper.GetKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketIntermediaryDenom(%d)", tc.marketID)
		})
	}
}

func TestGetKeyPrefixOrder(t *testing.T) {
	ktc := keyTestCase{
		maker: func() []byte {
			return keeper.GetKeyPrefixOrder()
		},
		expected: []byte{keeper.KeyTypeOrder},
	}
	checkKey(t, ktc, "GetKeyPrefixOrder")
}

func TestMakeKeyOrder(t *testing.T) {
	tests := []struct {
		name     string
		orderID  uint64
		expected []byte
	}{
		{
			name:     "order id 0",
			orderID:  0,
			expected: []byte{keeper.KeyTypeOrder, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:     "order id 1",
			orderID:  1,
			expected: []byte{keeper.KeyTypeOrder, 0, 0, 0, 0, 0, 0, 0, 1},
		},
		{
			name:     "order id 256",
			orderID:  256,
			expected: []byte{keeper.KeyTypeOrder, 0, 0, 0, 0, 0, 0, 1, 0},
		},
		{
			name:     "order id 65,536",
			orderID:  65_536,
			expected: []byte{keeper.KeyTypeOrder, 0, 0, 0, 0, 0, 1, 0, 0},
		},
		{
			name:     "order id 16,777,216",
			orderID:  16_777_216,
			expected: []byte{keeper.KeyTypeOrder, 0, 0, 0, 0, 1, 0, 0, 0},
		},
		{
			name:     "order id 4,294,967,296",
			orderID:  4_294_967_296,
			expected: []byte{keeper.KeyTypeOrder, 0, 0, 0, 1, 0, 0, 0, 0},
		},
		{
			name:     "order id 1,099,511,627,776",
			orderID:  1_099_511_627_776,
			expected: []byte{keeper.KeyTypeOrder, 0, 0, 1, 0, 0, 0, 0, 0},
		},
		{
			name:     "order id 281,474,976,710,656",
			orderID:  281_474_976_710_656,
			expected: []byte{keeper.KeyTypeOrder, 0, 1, 0, 0, 0, 0, 0, 0},
		},
		{
			name:     "order id 72,057,594,037,927,936",
			orderID:  72_057_594_037_927_936,
			expected: []byte{keeper.KeyTypeOrder, 1, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:     "order id 72,340,172,838,076,673",
			orderID:  72_340_172_838_076_673,
			expected: []byte{keeper.KeyTypeOrder, 1, 1, 1, 1, 1, 1, 1, 1},
		},
		{
			name:     "order id 1,229,782,938,247,303,441",
			orderID:  1_229_782_938_247_303_441,
			expected: []byte{keeper.KeyTypeOrder, 17, 17, 17, 17, 17, 17, 17, 17},
		},
		{
			name:     "order id 18,446,744,073,709,551,615",
			orderID:  18_446_744_073_709_551_615,
			expected: []byte{keeper.KeyTypeOrder, 255, 255, 255, 255, 255, 255, 255, 255},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeKeyOrder(tc.orderID)
				},
				expected:    tc.expected,
				expPrefixes: []expectedPrefix{{name: "", value: keeper.GetKeyPrefixOrder()}},
			}
			checkKey(t, ktc, "MakeKeyOrder(%d)", tc.orderID)
		})
	}
}

func TestParseKeyOrder(t *testing.T) {
	tests := []struct {
		name       string
		key        []byte
		expOrderID uint64
		expOK      bool
	}{
		{name: "nil key", key: nil, expOrderID: 0, expOK: false},
		{name: "empty key", key: []byte{}, expOrderID: 0, expOK: false},
		{name: "7 byte key", key: []byte{1, 2, 3, 4, 5, 6, 7}, expOrderID: 0, expOK: false},
		{name: "10 byte key", key: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, expOrderID: 0, expOK: false},
		{name: "9 byte key unknown type", key: []byte{99, 1, 2, 3, 4, 5, 6, 7, 8}, expOrderID: 0, expOK: false},
		{
			name:       "8 byte key",
			key:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
			expOrderID: 72_623_859_790_382_856,
			expOK:      true,
		},
		{
			name:       "9 byte key ask",
			key:        []byte{keeper.OrderKeyTypeAsk, 1, 2, 3, 4, 5, 6, 7, 8},
			expOrderID: 72_623_859_790_382_856,
			expOK:      true,
		},
		{
			name:       "9 byte key bid",
			key:        []byte{keeper.OrderKeyTypeBid, 1, 2, 3, 4, 5, 6, 7, 8},
			expOrderID: 72_623_859_790_382_856,
			expOK:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var orderID uint64
			var ok bool
			testFunc := func() {
				orderID, ok = keeper.ParseKeyOrder(tc.key)
			}
			require.NotPanics(t, testFunc, "ParseKeyOrder")
			assert.Equal(t, tc.expOK, ok, "ParseKeyOrder bool ok")
			assert.Equal(t, tc.expOrderID, orderID, "ParseKeyOrder order ID")
		})
	}
}

func TestParseIndexKeySuffixOrderID(t *testing.T) {
	tests := []struct {
		name       string
		key        []byte
		expOrderID uint64
		expOK      bool
	}{
		{name: "nil", key: nil, expOrderID: 0, expOK: false},
		{name: "empty", key: []byte{}, expOrderID: 0, expOK: false},
		{name: "1 byte", key: []byte{1}, expOrderID: 0, expOK: false},
		{name: "7 byte", key: []byte{1, 2, 3, 4, 5, 6, 7}, expOrderID: 0, expOK: false},
		{name: "8 bytes: 0", key: []byte{0, 0, 0, 0, 0, 0, 0, 0}, expOrderID: 0, expOK: true},
		{name: "8 bytes: 1", key: []byte{0, 0, 0, 0, 0, 0, 0, 1}, expOrderID: 1, expOK: true},
		{
			name:       "8 bytes: 4,294,967,296",
			key:        []byte{0, 0, 0, 1, 0, 0, 0, 0},
			expOrderID: 4_294_967_296,
			expOK:      true,
		},
		{
			name:       "8 bytes: 72,623,859,790,382,856",
			key:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
			expOrderID: 72_623_859_790_382_856,
			expOK:      true,
		},
		{
			name:       "8 bytes: max uint64",
			key:        []byte{255, 255, 255, 255, 255, 255, 255, 255},
			expOrderID: 18_446_744_073_709_551_615,
			expOK:      true,
		},
		{name: "9 bytes: 0", key: []byte{9, 0, 0, 0, 0, 0, 0, 0, 0}, expOrderID: 0, expOK: true},
		{name: "9 bytes: 1", key: []byte{9, 0, 0, 0, 0, 0, 0, 0, 1}, expOrderID: 1, expOK: true},
		{
			name:       "9 bytes: 4,294,967,296",
			key:        []byte{9, 0, 0, 0, 1, 0, 0, 0, 0},
			expOrderID: 4_294_967_296,
			expOK:      true,
		},
		{
			name:       "9 bytes: 72,623,859,790,382,856",
			key:        []byte{9, 1, 2, 3, 4, 5, 6, 7, 8},
			expOrderID: 72_623_859_790_382_856,
			expOK:      true,
		},
		{
			name:       "9 bytes: max uint64",
			key:        []byte{9, 255, 255, 255, 255, 255, 255, 255, 255},
			expOrderID: 18_446_744_073_709_551_615,
			expOK:      true,
		},
		{
			name:       "20 bytes: 0",
			key:        []byte{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 0, 0, 0, 0, 0, 0, 0, 0},
			expOrderID: 0,
			expOK:      true,
		},
		{
			name:       "20 bytes: 1",
			key:        []byte{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 0, 0, 0, 0, 0, 0, 0, 1},
			expOrderID: 1,
			expOK:      true,
		},
		{
			name:       "20 bytes: 4,294,967,296",
			key:        []byte{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 0, 0, 0, 1, 0, 0, 0, 0},
			expOrderID: 4_294_967_296,
			expOK:      true,
		},
		{
			name:       "20 bytes: 72,623,859,790,382,856",
			key:        []byte{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 1, 2, 3, 4, 5, 6, 7, 8},
			expOrderID: 72_623_859_790_382_856,
			expOK:      true,
		},
		{
			name:       "20 bytes: max uint64",
			key:        []byte{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 255, 255, 255, 255, 255, 255, 255, 255},
			expOrderID: 18_446_744_073_709_551_615,
			expOK:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var orderID uint64
			var ok bool
			testFunc := func() {
				orderID, ok = keeper.ParseIndexKeySuffixOrderID(tc.key)
			}
			require.NotPanics(t, testFunc, "ParseIndexKeySuffixOrderID")
			assert.Equal(t, tc.expOrderID, orderID, "ParseIndexKeySuffixOrderID orderID")
			assert.Equal(t, tc.expOK, ok, "ParseIndexKeySuffixOrderID ok bool")
		})
	}
}

func TestGetIndexKeyPrefixMarketToOrder(t *testing.T) {
	tests := []struct {
		name     string
		marketID uint32
		expected []byte
	}{
		{
			name:     "market id 0",
			marketID: 0,
			expected: []byte{keeper.KeyTypeMarketToOrderIndex, 0, 0, 0, 0},
		},
		{
			name:     "market id 1",
			marketID: 1,
			expected: []byte{keeper.KeyTypeMarketToOrderIndex, 0, 0, 0, 1},
		},
		{
			name:     "market id 255",
			marketID: 255,
			expected: []byte{keeper.KeyTypeMarketToOrderIndex, 0, 0, 0, 255},
		},
		{
			name:     "market id 256",
			marketID: 256,
			expected: []byte{keeper.KeyTypeMarketToOrderIndex, 0, 0, 1, 0},
		},
		{
			name:     "market id 65_536",
			marketID: 65_536,
			expected: []byte{keeper.KeyTypeMarketToOrderIndex, 0, 1, 0, 0},
		},
		{
			name:     "market id 16,777,216",
			marketID: 16_777_216,
			expected: []byte{keeper.KeyTypeMarketToOrderIndex, 1, 0, 0, 0},
		},
		{
			name:     "market id 16,843,009",
			marketID: 16_843_009,
			expected: []byte{keeper.KeyTypeMarketToOrderIndex, 1, 1, 1, 1},
		},
		{
			name:     "market id 4,294,967,295",
			marketID: 4_294_967_295,
			expected: []byte{keeper.KeyTypeMarketToOrderIndex, 255, 255, 255, 255},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.GetIndexKeyPrefixMarketToOrder(tc.marketID)
				},
				expected: tc.expected,
			}
			checkKey(t, ktc, "GetIndexKeyPrefixMarketToOrder(%d)", tc.marketID)
		})
	}
}

func TestMakeIndexKeyMarketToOrder(t *testing.T) {
	tests := []struct {
		name     string
		marketID uint32
		orderID  uint64
		expected []byte
	}{
		{
			name:     "market 0 order 0",
			marketID: 0,
			orderID:  0,
			expected: []byte{keeper.KeyTypeMarketToOrderIndex, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:     "market 0 order 1",
			marketID: 0,
			orderID:  1,
			expected: []byte{keeper.KeyTypeMarketToOrderIndex, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		},
		{
			name:     "market 0 order 72,340,172,838,076,673",
			marketID: 0,
			orderID:  72_340_172_838_076_673,
			expected: []byte{keeper.KeyTypeMarketToOrderIndex, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1},
		},
		{
			name:     "market 2 order 0",
			marketID: 2,
			orderID:  0,
			expected: []byte{keeper.KeyTypeMarketToOrderIndex, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:     "market 2 order 1",
			marketID: 2,
			orderID:  1,
			expected: []byte{keeper.KeyTypeMarketToOrderIndex, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 1},
		},
		{
			name:     "market 2 order 72,340,172,838,076,673",
			marketID: 2,
			orderID:  72_340_172_838_076_673,
			expected: []byte{keeper.KeyTypeMarketToOrderIndex, 0, 0, 0, 2, 1, 1, 1, 1, 1, 1, 1, 1},
		},
		{
			name:     "market 33,686,018 order 0",
			marketID: 33_686_018,
			orderID:  0,
			expected: []byte{keeper.KeyTypeMarketToOrderIndex, 2, 2, 2, 2, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:     "market 33,686,018 order 1",
			marketID: 33_686_018,
			orderID:  1,
			expected: []byte{keeper.KeyTypeMarketToOrderIndex, 2, 2, 2, 2, 0, 0, 0, 0, 0, 0, 0, 1},
		},
		{
			name:     "market 33,686,018 order 72,340,172,838,076,673",
			marketID: 33_686_018,
			orderID:  72_340_172_838_076_673,
			expected: []byte{keeper.KeyTypeMarketToOrderIndex, 2, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1},
		},

		{
			name:     "market max order max",
			marketID: 4_294_967_295,
			orderID:  18_446_744_073_709_551_615,
			expected: []byte{keeper.KeyTypeMarketToOrderIndex, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeIndexKeyMarketToOrder(tc.marketID, tc.orderID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "", value: keeper.GetIndexKeyPrefixMarketToOrder(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeIndexKeyMarketToOrder(%d, %d)", tc.marketID, tc.orderID)
		})
	}
}

func TestParseIndexKeyMarketToOrder(t *testing.T) {
	tests := []struct {
		name        string
		key         []byte
		expMarketID uint32
		expOrderID  uint64
		expErr      string
	}{
		{
			name:   "nil key",
			key:    nil,
			expErr: "cannot parse market to order key: length 0, expected 8, 12, or 13",
		},
		{
			name:   "empty key",
			key:    []byte{},
			expErr: "cannot parse market to order key: length 0, expected 8, 12, or 13",
		},
		{
			name:   "7 bytes",
			key:    []byte{1, 2, 3, 4, 5, 6, 7},
			expErr: "cannot parse market to order key: length 7, expected 8, 12, or 13",
		},
		{
			name:   "9 bytes",
			key:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9},
			expErr: "cannot parse market to order key: length 9, expected 8, 12, or 13",
		},
		{
			name:   "10 bytes",
			key:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			expErr: "cannot parse market to order key: length 10, expected 8, 12, or 13",
		},
		{
			name:   "11 bytes",
			key:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			expErr: "cannot parse market to order key: length 11, expected 8, 12, or 13",
		},
		{
			name:   "14 bytes",
			key:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14},
			expErr: "cannot parse market to order key: length 14, expected 8, 12, or 13",
		},
		{
			name:       "8 bytes order id 0",
			key:        []byte{0, 0, 0, 0, 0, 0, 0, 0},
			expOrderID: 0,
		},
		{
			name:        "8 bytes order id 1",
			key:         []byte{0, 0, 0, 0, 0, 0, 0, 1},
			expMarketID: 0,
			expOrderID:  1,
		},
		{
			name:       "8 bytes order id 72,623,859,790,382,856",
			key:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
			expOrderID: 72_623_859_790_382_856,
		},
		{
			name:        "12 bytes market id 1 order id 1",
			key:         []byte{0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 1},
			expMarketID: 1,
			expOrderID:  1,
		},
		{
			name:        "12 bytes market id 16,843,009 order id 144,680,345,676,153,346",
			key:         []byte{1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2},
			expMarketID: 16_843_009,
			expOrderID:  144_680_345_676_153_346,
		},
		{
			name:        "13 bytes market id 1 order id 1",
			key:         []byte{keeper.KeyTypeMarketToOrderIndex, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 1},
			expMarketID: 1,
			expOrderID:  1,
		},
		{
			name:        "13 bytes market id 16,843,009 order id 144,680,345,676,153,346",
			key:         []byte{keeper.KeyTypeMarketToOrderIndex, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2},
			expMarketID: 16_843_009,
			expOrderID:  144_680_345_676_153_346,
		},
		{
			name:   "13 bytes first byte too high",
			key:    []byte{0x4, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 1},
			expErr: "cannot parse market to order key: unknown type byte 0x4, expected 0x3",
		},
		{
			name:   "13 bytes first byte too low",
			key:    []byte{0x2, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 1},
			expErr: "cannot parse market to order key: unknown type byte 0x2, expected 0x3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var marketID uint32
			var orderID uint64
			var err error
			testFunc := func() {
				marketID, orderID, err = keeper.ParseIndexKeyMarketToOrder(tc.key)
			}
			require.NotPanics(t, testFunc, "ParseIndexKeyMarketToOrder(%v)", tc.key)
			assertions.AssertErrorValue(t, err, tc.expErr, "ParseIndexKeyMarketToOrder(%v) error", tc.key)
			assert.Equal(t, tc.expMarketID, marketID, "ParseIndexKeyMarketToOrder(%v) market id", tc.key)
			assert.Equal(t, tc.expOrderID, orderID, "ParseIndexKeyMarketToOrder(%v) order id", tc.key)
		})
	}
}

func TestGetIndexKeyPrefixAddressToOrder(t *testing.T) {
	tests := []struct {
		name     string
		addr     sdk.AccAddress
		expected []byte
		expPanic string
	}{
		{
			name:     "nil addr",
			addr:     nil,
			expPanic: "empty address not allowed",
		},
		{
			name:     "empty addr",
			addr:     sdk.AccAddress{},
			expPanic: "empty address not allowed",
		},
		{
			name:     "256 byte addr",
			addr:     sdk.AccAddress(bytes.Repeat([]byte{'P'}, 256)),
			expPanic: "address length should be max 255 bytes, got 256: unknown address",
		},
		{
			name:     "5 byte addr",
			addr:     sdk.AccAddress("abcde"),
			expected: append([]byte{keeper.KeyTypeAddressToOrderIndex, 5}, "abcde"...),
		},
		{
			name:     "20 byte addr",
			addr:     sdk.AccAddress("abcdefghijklmnopqrst"),
			expected: append([]byte{keeper.KeyTypeAddressToOrderIndex, 20}, "abcdefghijklmnopqrst"...),
		},
		{
			name:     "32 byte addr",
			addr:     sdk.AccAddress("abcdefghijklmnopqrstuvwxyzABCDEF"),
			expected: append([]byte{keeper.KeyTypeAddressToOrderIndex, 32}, "abcdefghijklmnopqrstuvwxyzABCDEF"...),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.GetIndexKeyPrefixAddressToOrder(tc.addr)
				},
				expected: tc.expected,
				expPanic: tc.expPanic,
			}
			checkKey(t, ktc, "GetIndexKeyPrefixAddressToOrder(%s)", string(tc.addr))
		})
	}
}

func TestMakeIndexKeyAddressToOrder(t *testing.T) {
	tests := []struct {
		name     string
		addr     sdk.AccAddress
		orderID  uint64
		expected []byte
		expPanic string
	}{
		{
			name:     "nil addr",
			addr:     nil,
			expPanic: "empty address not allowed",
		},
		{
			name:     "empty addr",
			addr:     sdk.AccAddress{},
			expPanic: "empty address not allowed",
		},
		{
			name:     "256 byte addr",
			addr:     sdk.AccAddress(bytes.Repeat([]byte{'P'}, 256)),
			expPanic: "address length should be max 255 bytes, got 256: unknown address",
		},
		{
			name:    "5 byte addr order 1",
			addr:    sdk.AccAddress("abcde"),
			orderID: 1,
			expected: concatBz(
				[]byte{keeper.KeyTypeAddressToOrderIndex, 5},
				[]byte("abcde"),
				[]byte{0, 0, 0, 0, 0, 0, 0, 1},
			),
		},
		{
			name:    "20 byte addr order 1",
			addr:    sdk.AccAddress("abcdefghijklmnopqrst"),
			orderID: 1,
			expected: concatBz(
				[]byte{keeper.KeyTypeAddressToOrderIndex, 20},
				[]byte("abcdefghijklmnopqrst"),
				[]byte{0, 0, 0, 0, 0, 0, 0, 1},
			),
		},
		{
			name:    "32 byte addr order 1",
			addr:    sdk.AccAddress("abcdefghijklmnopqrstuvwxyzABCDEF"),
			orderID: 1,
			expected: concatBz(
				[]byte{keeper.KeyTypeAddressToOrderIndex, 32},
				[]byte("abcdefghijklmnopqrstuvwxyzABCDEF"),
				[]byte{0, 0, 0, 0, 0, 0, 0, 1},
			),
		},
		{
			name:    "20 byte addr order 5",
			addr:    sdk.AccAddress("ABCDEFGHIJKLMNOPQRST"),
			orderID: 5,
			expected: concatBz(
				[]byte{keeper.KeyTypeAddressToOrderIndex, 20},
				[]byte("ABCDEFGHIJKLMNOPQRST"),
				[]byte{0, 0, 0, 0, 0, 0, 0, 5},
			),
		},
		{
			name:    "20 byte addr order 72,623,859,790,382,856",
			addr:    sdk.AccAddress("ABCDEFGHIJKLMNOPQRST"),
			orderID: 72_623_859_790_382_856,
			expected: concatBz(
				[]byte{keeper.KeyTypeAddressToOrderIndex, 20},
				[]byte("ABCDEFGHIJKLMNOPQRST"),
				[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeIndexKeyAddressToOrder(tc.addr, tc.orderID)
				},
				expected: tc.expected,
				expPanic: tc.expPanic,
			}
			if len(tc.expPanic) == 0 {
				ktc.expPrefixes = []expectedPrefix{
					{name: "", value: keeper.GetIndexKeyPrefixAddressToOrder(tc.addr)},
				}
			}
			checkKey(t, ktc, "MakeIndexKeyAddressToOrder(%s, %d)", string(tc.addr), tc.orderID)
		})
	}
}

func TestParseIndexKeyAddressToOrder(t *testing.T) {
	tests := []struct {
		name       string
		key        []byte
		expAddr    sdk.AccAddress
		expOrderID uint64
		expErr     string
	}{
		{
			name:   "nil key",
			key:    nil,
			expErr: "cannot parse address to order index key: only has 0 bytes, expected at least 8",
		},
		{
			name:   "empty key",
			key:    []byte{},
			expErr: "cannot parse address to order index key: only has 0 bytes, expected at least 8",
		},
		{
			name:   "7 bytes",
			key:    []byte{1, 2, 3, 4, 5, 6, 7},
			expErr: "cannot parse address to order index key: only has 7 bytes, expected at least 8",
		},
		{
			name:       "just order id 1",
			key:        []byte{0, 0, 0, 0, 0, 0, 0, 1},
			expOrderID: 1,
		},
		{
			name:       "just order id 72,623,859,790,382,856",
			key:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
			expOrderID: 72_623_859_790_382_856,
		},
		{
			name:   "9 bytes",
			key:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9},
			expErr: "cannot parse address to order index key: unable to determine address from single byte 0x1",
		},
		{
			name:       "1 byte address order id 72,623,859,790,382,856",
			key:        []byte{1, 55, 1, 2, 3, 4, 5, 6, 7, 8},
			expAddr:    sdk.AccAddress{55},
			expOrderID: 72_623_859_790_382_856,
		},
		{
			name:   "length byte 2 but only 1 byte after it",
			key:    []byte{2, 55, 1, 2, 3, 4, 5, 6, 7, 8},
			expErr: "cannot parse address to order index key: unable to determine address from [2, 55, ...(length 2)]",
		},
		{
			name:   "length byte 2 but 3 bytes after it",
			key:    []byte{2, 55, 56, 57, 1, 2, 3, 4, 5, 6, 7, 8},
			expErr: "cannot parse address to order index key: unable to determine address from [2, 55, ...(length 4)]",
		},
		{
			name:       "length byte 4 order id 72,623,859,790,382,856",
			key:        []byte{4, 55, 56, 57, 58, 1, 2, 3, 4, 5, 6, 7, 8},
			expAddr:    sdk.AccAddress{55, 56, 57, 58},
			expOrderID: 72_623_859_790_382_856,
		},
		{
			name: "length byte 20 but only 19 byte after it",
			key: []byte{20, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110,
				111, 112, 113, 114, 115, 116, 117, 118, 119,
				1, 2, 3, 4, 5, 6, 7, 8},
			expErr: "cannot parse address to order index key: unable to determine address from [20, 101, ...(length 20)]",
		},
		{
			name: "length byte 20 but 21 byte after it",
			key: []byte{20, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110,
				111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121,
				1, 2, 3, 4, 5, 6, 7, 8},
			expErr: "cannot parse address to order index key: unable to determine address from [20, 101, ...(length 22)]",
		},
		{
			name: "20 byte address order id 72,623,859,790,382,856",
			key: []byte{20, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110,
				111, 112, 113, 114, 115, 116, 117, 118, 119, 120,
				1, 2, 3, 4, 5, 6, 7, 8},
			expAddr: sdk.AccAddress{101, 102, 103, 104, 105, 106, 107, 108, 109, 110,
				111, 112, 113, 114, 115, 116, 117, 118, 119, 120},
			expOrderID: 72_623_859_790_382_856,
		},
		{
			name:       "with type byte: 1 byte address order id 72,623,859,790,382,856",
			key:        []byte{keeper.KeyTypeAddressToOrderIndex, 1, 55, 1, 2, 3, 4, 5, 6, 7, 8},
			expAddr:    sdk.AccAddress{55},
			expOrderID: 72_623_859_790_382_856,
		},
		{
			name:   "with type byte: length byte 2 but only 1 byte after it",
			key:    []byte{keeper.KeyTypeAddressToOrderIndex, 2, 55, 1, 2, 3, 4, 5, 6, 7, 8},
			expErr: "cannot parse address to order index key: unable to determine address from [4, 2, ...(length 3)]",
		},
		{
			name:   "with type byte: length byte 2 but 5 bytes after it",
			key:    []byte{keeper.KeyTypeAddressToOrderIndex, 2, 55, 56, 57, 58, 1, 2, 3, 4, 5, 6, 7, 8},
			expErr: "cannot parse address to order index key: unable to determine address from [4, 2, ...(length 6)]",
		},
		{
			name:       "with type byte: length byte 4 order id 72,623,859,790,382,856",
			key:        []byte{keeper.KeyTypeAddressToOrderIndex, 4, 55, 56, 57, 58, 1, 2, 3, 4, 5, 6, 7, 8},
			expAddr:    sdk.AccAddress{55, 56, 57, 58},
			expOrderID: 72_623_859_790_382_856,
		},
		{
			name: "with type byte: length byte 20 but only 19 byte after it",
			key: []byte{keeper.KeyTypeAddressToOrderIndex, 20,
				101, 102, 103, 104, 105, 106, 107, 108, 109, 110,
				111, 112, 113, 114, 115, 116, 117, 118, 119,
				1, 2, 3, 4, 5, 6, 7, 8},
			expErr: "cannot parse address to order index key: unable to determine address from [4, 20, ...(length 21)]",
		},
		{
			name: "with type byte: length byte 20 but 21 byte after it",
			key: []byte{keeper.KeyTypeAddressToOrderIndex, 20,
				101, 102, 103, 104, 105, 106, 107, 108, 109, 110,
				111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121,
				1, 2, 3, 4, 5, 6, 7, 8},
			expErr: "cannot parse address to order index key: unable to determine address from [4, 20, ...(length 23)]",
		},
		{
			name: "with type byte: 20 byte address order id 72,623,859,790,382,856",
			key: []byte{keeper.KeyTypeAddressToOrderIndex, 20,
				101, 102, 103, 104, 105, 106, 107, 108, 109, 110,
				111, 112, 113, 114, 115, 116, 117, 118, 119, 120,
				1, 2, 3, 4, 5, 6, 7, 8},
			expAddr: sdk.AccAddress{101, 102, 103, 104, 105, 106, 107, 108, 109, 110,
				111, 112, 113, 114, 115, 116, 117, 118, 119, 120},
			expOrderID: 72_623_859_790_382_856,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var addr sdk.AccAddress
			var orderID uint64
			var err error
			testFunc := func() {
				addr, orderID, err = keeper.ParseIndexKeyAddressToOrder(tc.key)
			}
			require.NotPanics(t, testFunc, "ParseIndexKeyAddressToOrder(%v)", tc.key)
			assertions.AssertErrorValue(t, err, tc.expErr, "ParseIndexKeyAddressToOrder(%v) error", tc.key)
			assert.Equal(t, tc.expAddr, addr, "ParseIndexKeyAddressToOrder(%v) address", tc.key)
			assert.Equal(t, tc.expOrderID, orderID, "ParseIndexKeyAddressToOrder(%v) order id", tc.key)
		})
	}
}

func TestGetIndexKeyPrefixAssetToOrder(t *testing.T) {
	tests := []struct {
		name       string
		assetDenom string
		expected   []byte
	}{
		{
			name:       "empty",
			assetDenom: "",
			expected:   []byte{keeper.KeyTypeAssetToOrderIndex},
		},
		{
			name:       "1 char denom",
			assetDenom: "p",
			expected:   []byte{keeper.KeyTypeAssetToOrderIndex, 'p'},
		},
		{
			name:       "nhash",
			assetDenom: "nhash",
			expected:   []byte{keeper.KeyTypeAssetToOrderIndex, 'n', 'h', 'a', 's', 'h'},
		},
		{
			name:       "hex string",
			assetDenom: hexString,
			expected:   append([]byte{keeper.KeyTypeAssetToOrderIndex}, hexString...),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.GetIndexKeyPrefixAssetToOrder(tc.assetDenom)
				},
				expected: tc.expected,
			}
			checkKey(t, ktc, "GetIndexKeyPrefixAssetToOrder(%q)", tc.assetDenom)
		})
	}
}

func TestMakeIndexKeyAssetToOrder(t *testing.T) {
	tests := []struct {
		name       string
		assetDenom string
		orderID    uint64
		expected   []byte
		expPanic   string
	}{
		{
			name:       "no asset order 0",
			assetDenom: "",
			orderID:    0,
			expected:   []byte{keeper.KeyTypeAssetToOrderIndex, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name:       "nhash order 1",
			assetDenom: "nhash",
			orderID:    1,
			expected:   []byte{keeper.KeyTypeAssetToOrderIndex, 'n', 'h', 'a', 's', 'h', 0, 0, 0, 0, 0, 0, 0, 1},
		},
		{
			name:       "hex string order 5",
			assetDenom: hexString,
			orderID:    5,
			expected: concatBz(
				[]byte{keeper.KeyTypeAssetToOrderIndex},
				[]byte(hexString),
				[]byte{0, 0, 0, 0, 0, 0, 0, 5},
			),
		},
		{
			name:       "nhash order 4,294,967,296",
			assetDenom: "nhash",
			orderID:    4_294_967_296,
			expected: []byte{
				keeper.KeyTypeAssetToOrderIndex,
				'n', 'h', 'a', 's', 'h',
				0, 0, 0, 1, 0, 0, 0, 0,
			},
		},
		{
			name:       "nhash order max",
			assetDenom: "nhash",
			orderID:    18_446_744_073_709_551_615,
			expected: []byte{
				keeper.KeyTypeAssetToOrderIndex,
				'n', 'h', 'a', 's', 'h',
				255, 255, 255, 255, 255, 255, 255, 255,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeIndexKeyAssetToOrder(tc.assetDenom, tc.orderID)
				},
				expected: tc.expected,
				expPanic: tc.expPanic,
			}
			if len(tc.expPanic) == 0 {
				ktc.expPrefixes = []expectedPrefix{
					{name: "GetIndexKeyPrefixAssetToOrder", value: keeper.GetIndexKeyPrefixAssetToOrder(tc.assetDenom)},
				}
			}

			checkKey(t, ktc, "MakeIndexKeyAssetToOrder(%q, %d)", tc.assetDenom, tc.orderID)
		})
	}
}

func TestParseIndexKeyAssetToOrder(t *testing.T) {
	tests := []struct {
		name       string
		key        []byte
		expDenom   string
		expOrderID uint64
		expErr     string
	}{
		{
			name:   "nil key",
			key:    nil,
			expErr: "cannot parse asset to order key: only has 0 bytes, expected at least 8",
		},
		{
			name:   "empty key",
			key:    []byte{},
			expErr: "cannot parse asset to order key: only has 0 bytes, expected at least 8",
		},
		{
			name:   "7 bytes",
			key:    []byte{1, 2, 3, 4, 5, 6, 7},
			expErr: "cannot parse asset to order key: only has 7 bytes, expected at least 8",
		},
		{
			name:       "order id 1",
			key:        []byte{0, 0, 0, 0, 0, 0, 0, 1},
			expOrderID: 1,
		},
		{
			name:       "order id 578,437,695,752,307,201",
			key:        []byte{8, 7, 6, 5, 4, 3, 2, 1},
			expOrderID: 578_437_695_752_307_201,
		},
		{
			name:       "nhash order id 578,437,695,752,307,201",
			key:        []byte{'n', 'h', 'a', 's', 'h', 8, 7, 6, 5, 4, 3, 2, 1},
			expDenom:   "nhash",
			expOrderID: 578_437_695_752_307_201,
		},
		{
			name:       "hex string order id 578,437,695,752,307,201",
			key:        append([]byte(hexString), 8, 7, 6, 5, 4, 3, 2, 1),
			expDenom:   hexString,
			expOrderID: 578_437_695_752_307_201,
		},
		{
			name:       "with type byte nhash order id 578,437,695,752,307,201",
			key:        []byte{keeper.KeyTypeAssetToOrderIndex, 'n', 'h', 'a', 's', 'h', 8, 7, 6, 5, 4, 3, 2, 1},
			expDenom:   "nhash",
			expOrderID: 578_437_695_752_307_201,
		},
		{
			name: "with type byte hex string order id 578,437,695,752,307,201",
			key: concatBz(
				[]byte{keeper.KeyTypeAssetToOrderIndex},
				[]byte(hexString),
				[]byte{8, 7, 6, 5, 4, 3, 2, 1},
			),
			expDenom:   hexString,
			expOrderID: 578_437_695_752_307_201,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var denom string
			var orderiD uint64
			var err error
			testFunc := func() {
				denom, orderiD, err = keeper.ParseIndexKeyAssetToOrder(tc.key)
			}
			require.NotPanics(t, testFunc, "ParseIndexKeyAssetToOrder(%v)", tc.key)
			assertions.AssertErrorValue(t, err, tc.expErr, "ParseIndexKeyAssetToOrder(%v) error", tc.key)
			assert.Equal(t, tc.expDenom, denom, "ParseIndexKeyAssetToOrder(%v) denom", tc.key)
			assert.Equal(t, tc.expOrderID, orderiD, "ParseIndexKeyAssetToOrder(%v) order id", tc.key)
		})
	}
}

func TestMakeIndexKeyMarketExternalIDToOrder(t *testing.T) {
	tests := []struct {
		name       string
		marketID   uint32
		externalID string
		expected   []byte
		expPanic   string
	}{
		{
			name:       "empty external id",
			marketID:   1,
			externalID: "",
			expPanic:   "cannot create market external id to order index with empty external id",
		},
		{
			name:       "external id too long",
			marketID:   2,
			externalID: strings.Repeat("H", exchange.MaxExternalIDLength+1),
			expPanic: fmt.Sprintf("cannot create market external id to order index: invalid external id %q (length %d): max length %d",
				"HHHHH...HHHHH", exchange.MaxExternalIDLength+1, exchange.MaxExternalIDLength),
		},
		{
			name:       "market 0, a zeroed uuid",
			marketID:   0,
			externalID: "00000000-0000-0000-0000-000000000000",
			expected: append([]byte{keeper.KeyTypeMarketExternalIDToOrderIndex, 0, 0, 0, 0},
				"00000000-0000-0000-0000-000000000000"...),
		},
		{
			name:       "market 1, random uuid",
			marketID:   1,
			externalID: "814348F5-DE62-4954-81D1-65874E37C0BE",
			expected: append([]byte{keeper.KeyTypeMarketExternalIDToOrderIndex, 0, 0, 0, 1},
				"814348F5-DE62-4954-81D1-65874E37C0BE"...),
		},
		{
			name:       "market 67,305,985, a wierd external id",
			marketID:   67_305_985,
			externalID: "ThisIsWierd",
			expected:   append([]byte{keeper.KeyTypeMarketExternalIDToOrderIndex, 4, 3, 2, 1}, "ThisIsWierd"...),
		},
		{
			name:       "max market id and lots of Zs",
			marketID:   4_294_967_295,
			externalID: strings.Repeat("Z", exchange.MaxExternalIDLength),
			expected: append([]byte{keeper.KeyTypeMarketExternalIDToOrderIndex, 255, 255, 255, 255},
				strings.Repeat("Z", exchange.MaxExternalIDLength)...),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeIndexKeyMarketExternalIDToOrder(tc.marketID, tc.externalID)
				},
				expected: tc.expected,
				expPanic: tc.expPanic,
			}

			checkKey(t, ktc, "MakeIndexKeyMarketExternalIDToOrder(%d, %v)", tc.marketID, tc.externalID)
		})
	}
}

func TestGetKeyPrefixCommitments(t *testing.T) {
	ktc := keyTestCase{
		maker: func() []byte {
			return keeper.GetKeyPrefixCommitments()
		},
		expected: []byte{keeper.KeyTypeCommitment},
	}
	checkKey(t, ktc, "GetKeyPrefixCommitments")

}

func TestGetKeyPrefixCommitmentsToMarket(t *testing.T) {
	tests := []struct {
		name     string
		marketID uint32
		expected []byte
	}{
		{
			name:     "market id 0",
			marketID: 0,
			expected: []byte{keeper.KeyTypeCommitment, 0, 0, 0, 0},
		},
		{
			name:     "market id 1",
			marketID: 1,
			expected: []byte{keeper.KeyTypeCommitment, 0, 0, 0, 1},
		},
		{
			name:     "market id 255",
			marketID: 255,
			expected: []byte{keeper.KeyTypeCommitment, 0, 0, 0, 255},
		},
		{
			name:     "market id 256",
			marketID: 256,
			expected: []byte{keeper.KeyTypeCommitment, 0, 0, 1, 0},
		},
		{
			name:     "market id 65_536",
			marketID: 65_536,
			expected: []byte{keeper.KeyTypeCommitment, 0, 1, 0, 0},
		},
		{
			name:     "market id 16,777,216",
			marketID: 16_777_216,
			expected: []byte{keeper.KeyTypeCommitment, 1, 0, 0, 0},
		},
		{
			name:     "market id 16,843,009",
			marketID: 16_843_009,
			expected: []byte{keeper.KeyTypeCommitment, 1, 1, 1, 1},
		},
		{
			name:     "market id 4,294,967,295",
			marketID: 4_294_967_295,
			expected: []byte{keeper.KeyTypeCommitment, 255, 255, 255, 255},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.GetKeyPrefixCommitmentsToMarket(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "GetKeyPrefixCommitments", value: keeper.GetKeyPrefixCommitments()},
				},
			}
			checkKey(t, ktc, "GetKeyPrefixCommitmentsToMarket(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyCommitment(t *testing.T) {
	tests := []struct {
		name     string
		marketID uint32
		addr     sdk.AccAddress
		expected []byte
		expPanic string
	}{
		{
			name:     "nil addr",
			addr:     nil,
			expPanic: "empty address not allowed",
		},
		{
			name:     "empty addr",
			addr:     sdk.AccAddress{},
			expPanic: "empty address not allowed",
		},
		{
			name:     "256 byte addr",
			addr:     bytes.Repeat([]byte{'p'}, 256),
			expPanic: "address length should be max 255 bytes, got 256: unknown address",
		},
		{
			name:     "market id 0 5 byte addr",
			marketID: 0,
			addr:     sdk.AccAddress("abcde"),
			expected: append([]byte{keeper.KeyTypeCommitment, 0, 0, 0, 0, 5}, "abcde"...),
		},
		{
			name:     "market id 0 20 byte addr",
			marketID: 0,
			addr:     sdk.AccAddress("abcdefghijklmnopqrst"),
			expected: append([]byte{keeper.KeyTypeCommitment, 0, 0, 0, 0, 20}, "abcdefghijklmnopqrst"...),
		},
		{
			name:     "market id 0 32 byte addr",
			marketID: 0,
			addr:     sdk.AccAddress("abcdefghijklmnopqrstuvwxyzABCDEF"),
			expected: append([]byte{keeper.KeyTypeCommitment, 0, 0, 0, 0, 32}, "abcdefghijklmnopqrstuvwxyzABCDEF"...),
		},
		{
			name:     "market id 1 20 byte addr",
			marketID: 1,
			addr:     sdk.AccAddress("abcdefghijklmnopqrst"),
			expected: append([]byte{keeper.KeyTypeCommitment, 0, 0, 0, 1, 20}, "abcdefghijklmnopqrst"...),
		},
		{
			name:     "market id 1 32 byte addr",
			marketID: 1,
			addr:     sdk.AccAddress("abcdefghijklmnopqrstuvwxyzABCDEF"),
			expected: append([]byte{keeper.KeyTypeCommitment, 0, 0, 0, 1, 32}, "abcdefghijklmnopqrstuvwxyzABCDEF"...),
		},
		{
			name:     "market id 16,843,009 20 byte addr",
			marketID: 16_843_009,
			addr:     sdk.AccAddress("abcdefghijklmnopqrst"),
			expected: append([]byte{keeper.KeyTypeCommitment, 1, 1, 1, 1, 20}, "abcdefghijklmnopqrst"...),
		},
		{
			name:     "market id 16,843,009 32 byte addr",
			marketID: 16_843_009,
			addr:     sdk.AccAddress("abcdefghijklmnopqrstuvwxyzABCDEF"),
			expected: append([]byte{keeper.KeyTypeCommitment, 1, 1, 1, 1, 32}, "abcdefghijklmnopqrstuvwxyzABCDEF"...),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeKeyCommitment(tc.marketID, tc.addr)
				},
				expected: tc.expected,
				expPanic: tc.expPanic,
			}
			if len(tc.expPanic) == 0 {
				ktc.expPrefixes = []expectedPrefix{
					{name: "GetKeyPrefixCommitments", value: keeper.GetKeyPrefixCommitments()},
					{name: "GetKeyPrefixCommitmentsToMarket", value: keeper.GetKeyPrefixCommitmentsToMarket(tc.marketID)},
				}
			}
			checkKey(t, ktc, "MakeKeyCommitment(%d, %s)", tc.marketID, tc.addr)
		})
	}
}

func TestParseKeyCommitment(t *testing.T) {
	tests := []struct {
		name        string
		key         []byte
		expMarketID uint32
		expAddr     sdk.AccAddress
		expErr      string
	}{
		{
			name:   "nil",
			key:    nil,
			expErr: "cannot parse commitment key: only has 0 bytes, expected at least 7",
		},
		{
			name:   "empty",
			key:    []byte{},
			expErr: "cannot parse commitment key: only has 0 bytes, expected at least 7",
		},
		{
			name:   "6 bytes",
			key:    []byte{keeper.KeyTypeCommitment, 2, 3, 4, 5, 6},
			expErr: "cannot parse commitment key: only has 6 bytes, expected at least 7",
		},
		{
			name:   "type byte one low",
			key:    []byte{keeper.KeyTypeCommitment - 1, 2, 3, 4, 5, 6, 7},
			expErr: "cannot parse commitment key: incorrect type byte 0x62",
		},
		{
			name:   "type byte one high",
			key:    []byte{keeper.KeyTypeCommitment + 1, 2, 3, 4, 5, 6, 7},
			expErr: "cannot parse commitment key: incorrect type byte 0x64",
		},
		{
			name:   "addr length byte zero",
			key:    []byte{keeper.KeyTypeCommitment, 1, 2, 3, 4, 0, 7},
			expErr: "cannot parse address from commitment key: length byte is zero",
		},
		{
			name:   "addr length byte too large",
			key:    []byte{keeper.KeyTypeCommitment, 1, 2, 3, 4, 6, 1, 2, 3, 4, 5},
			expErr: "cannot parse address from commitment key: length byte is 6, but slice only has 5 left",
		},
		{
			name:   "addr length byte too small",
			key:    []byte{keeper.KeyTypeCommitment, 1, 2, 3, 4, 4, 1, 2, 3, 4, 5},
			expErr: "cannot parse address from commitment key: found 1 bytes after address, expected 0",
		},
		{
			name:        "market 1; 1 byte addr",
			key:         []byte{keeper.KeyTypeCommitment, 0, 0, 0, 1, 1, 7},
			expMarketID: 1,
			expAddr:     sdk.AccAddress{7},
		},
		{
			name: "market 2; 20 byte addr",
			key: []byte{keeper.KeyTypeCommitment, 0, 0, 0, 2, 20,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			expMarketID: 2,
			expAddr:     sdk.AccAddress{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		},
		{
			name: "market 67,305,985; 20 byte addr",
			key: []byte{keeper.KeyTypeCommitment, 4, 3, 2, 1, 20,
				20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			expMarketID: 67_305_985,
			expAddr:     sdk.AccAddress{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		},
		{
			name: "market 4,294,967,295; 32 byte addr",
			key: []byte{keeper.KeyTypeCommitment, 255, 255, 255, 255, 32,
				101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116,
				117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131, 132,
			},
			expMarketID: 4_294_967_295,
			expAddr: sdk.AccAddress{
				101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116,
				117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131, 132,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var marketID uint32
			var addr sdk.AccAddress
			var err error
			testFunc := func() {
				marketID, addr, err = keeper.ParseKeyCommitment(tc.key)
			}
			require.NotPanics(t, testFunc, "ParseKeyCommitment(%q)", tc.key)
			assertions.AssertErrorValue(t, err, tc.expErr, "ParseKeyCommitment(%q) error", tc.key)
			assert.Equal(t, tc.expMarketID, marketID, "ParseKeyCommitment(%q) market id", tc.key)
			assert.Equal(t, tc.expAddr, addr, "ParseKeyCommitment(%q) addr", tc.key)
		})
	}
}

func TestParseKeySuffixCommitment(t *testing.T) {
	tests := []struct {
		name    string
		key     []byte
		expAddr sdk.AccAddress
		expErr  string
	}{
		{
			name:   "nil",
			key:    nil,
			expErr: "cannot parse address from commitment key: slice is empty",
		},
		{
			name:   "empty",
			key:    []byte{},
			expErr: "cannot parse address from commitment key: slice is empty",
		},
		{
			name:   "addr length byte zero",
			key:    []byte{0},
			expErr: "cannot parse address from commitment key: length byte is zero",
		},
		{
			name:   "addr length byte too large",
			key:    []byte{6, 1, 2, 3, 4, 5},
			expErr: "cannot parse address from commitment key: length byte is 6, but slice only has 5 left",
		},
		{
			name:   "addr length byte too small",
			key:    []byte{4, 1, 2, 3, 4, 5},
			expErr: "cannot parse address from commitment key: found 1 bytes after address, expected 0",
		},
		{
			name:    "1 byte addr",
			key:     []byte{1, 7},
			expAddr: sdk.AccAddress{7},
		},
		{
			name:    "20 byte addr",
			key:     []byte{20, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			expAddr: sdk.AccAddress{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		},
		{
			name: "32 byte addr",
			key: []byte{32,
				101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116,
				117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131, 132,
			},
			expAddr: sdk.AccAddress{
				101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116,
				117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131, 132,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var addr sdk.AccAddress
			var err error
			testFunc := func() {
				addr, err = keeper.ParseKeySuffixCommitment(tc.key)
			}
			require.NotPanics(t, testFunc, "ParseKeySuffixCommitment(%q)", tc.key)
			assertions.AssertErrorValue(t, err, tc.expErr, "ParseKeySuffixCommitment(%q) error", tc.key)
			assert.Equal(t, tc.expAddr, addr, "ParseKeySuffixCommitment(%q) addr", tc.key)
		})
	}
}
