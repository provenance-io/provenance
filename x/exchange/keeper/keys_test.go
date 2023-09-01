package keeper_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

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

type keyTestCase struct {
	maker       func() []byte
	expected    []byte
	expPanic    string
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
				{name: "KeyTypeMarket", value: keeper.KeyTypeMarket},
				{name: "KeyTypeOrder", value: keeper.KeyTypeOrder},
				{name: "KeyTypeMarketToOrderIndex", value: keeper.KeyTypeMarketToOrderIndex},
				{name: "KeyTypeAddressToOrderIndex", value: keeper.KeyTypeAddressToOrderIndex},
				{name: "KeyTypeAssetToOrderIndex", value: keeper.KeyTypeAssetToOrderIndex},
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
				{name: "MarketKeyTypeInactive", value: keeper.MarketKeyTypeInactive},
				{name: "MarketKeyTypeUserSettle", value: keeper.MarketKeyTypeUserSettle},
				{name: "MarketKeyTypePermissions", value: keeper.MarketKeyTypePermissions},
				{name: "MarketKeyTypeReqAttr", value: keeper.MarketKeyTypeReqAttr},
			},
		},
		{
			name: "order types",
			types: []byteEntry{
				{name: "OrderKeyTypeAsk", value: keeper.OrderKeyTypeAsk},
				{name: "OrderKeyTypeBid", value: keeper.OrderKeyTypeBid},
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

func TestMakeKeyPrefixParamsSplit(t *testing.T) {
	ktc := keyTestCase{
		maker: func() []byte {
			return keeper.MakeKeyPrefixParamsSplit()
		},
		expected: []byte{keeper.KeyTypeParams, 's', 'p', 'l', 'i', 't'},
	}
	checkKey(t, ktc, "MakeKeyPrefixParamsSplit")
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
					{name: "MakeKeyPrefixParamsSplit", value: keeper.MakeKeyPrefixParamsSplit()},
				},
			}
			checkKey(t, ktc, "MakeKeyParamsSplit(%q)", tc.denom)
		})
	}
}

func TestMakeKeyPrefixMarket(t *testing.T) {
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
					return keeper.MakeKeyPrefixMarket(tc.marketID)
				},
				expected: tc.expected,
			}
			checkKey(t, ktc, "MakeKeyPrefixMarket(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyPrefixMarketCreateAskFlatFee(t *testing.T) {
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
					return keeper.MakeKeyPrefixMarketCreateAskFlatFee(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "MakeKeyPrefixMarket", value: keeper.MakeKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyPrefixMarketCreateAskFlatFee(%d)", tc.marketID)
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
						name:  "MakeKeyPrefixMarket",
						value: keeper.MakeKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "MakeKeyPrefixMarketCreateBidFlatFee",
						value: keeper.MakeKeyPrefixMarketCreateAskFlatFee(tc.marketID),
					},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketCreateAskFlatFee(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyPrefixMarketCreateBidFlatFee(t *testing.T) {
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
					return keeper.MakeKeyPrefixMarketCreateBidFlatFee(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "MakeKeyPrefixMarket", value: keeper.MakeKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyPrefixMarketCreateBidFlatFee(%d)", tc.marketID)
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
						name:  "MakeKeyPrefixMarket",
						value: keeper.MakeKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "MakeKeyPrefixMarketCreateBidFlatFee",
						value: keeper.MakeKeyPrefixMarketCreateBidFlatFee(tc.marketID),
					},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketCreateBidFlatFee(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyPrefixMarketSellerSettlementFlatFee(t *testing.T) {
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
					return keeper.MakeKeyPrefixMarketSellerSettlementFlatFee(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "MakeKeyPrefixMarket", value: keeper.MakeKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyPrefixMarketSellerSettlementFlatFee(%d)", tc.marketID)
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
						name:  "MakeKeyPrefixMarket",
						value: keeper.MakeKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "MakeKeyPrefixMarketSellerSettlementFlatFee",
						value: keeper.MakeKeyPrefixMarketSellerSettlementFlatFee(tc.marketID),
					},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketSellerSettlementFlatFee(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyPrefixMarketSellerSettlementRatio(t *testing.T) {
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
					return keeper.MakeKeyPrefixMarketSellerSettlementRatio(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "MakeKeyPrefixMarket", value: keeper.MakeKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyPrefixMarketSellerSettlementRatio(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyMarketSellerSettlementRatio(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeSellerSettlementRatio
	coin := func(denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.OneInt()}
	}

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
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte, 0},
		},
		{
			name:     "market id 1 nhash to empty",
			marketID: 1,
			ratio:    exchange.FeeRatio{Price: coin("nhash"), Fee: coin("")},
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte, 'n', 'h', 'a', 's', 'h', 0},
		},
		{
			name:     "market id 1 empty to nhash",
			marketID: 1,
			ratio:    exchange.FeeRatio{Price: coin(""), Fee: coin("nhash")},
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte, 0, 'n', 'h', 'a', 's', 'h'},
		},
		{
			name:     "market id 1 nhash to nhash",
			marketID: 1,
			ratio:    exchange.FeeRatio{Price: coin("nhash"), Fee: coin("nhash")},
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte, 'n', 'h', 'a', 's', 'h', 0, 'n', 'h', 'a', 's', 'h'},
		},
		{
			name:     "market id 16,843,009 nhash to hex string",
			marketID: 16_843_009,
			ratio:    exchange.FeeRatio{Price: coin("nhash"), Fee: coin(hexString)},
			expected: append([]byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte, 'n', 'h', 'a', 's', 'h', 0}, hexString...),
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
						name:  "MakeKeyPrefixMarket",
						value: keeper.MakeKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "MakeKeyPrefixMarketSellerSettlementRatio",
						value: keeper.MakeKeyPrefixMarketSellerSettlementRatio(tc.marketID),
					},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketSellerSettlementRatio(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyPrefixMarketBuyerSettlementFlatFee(t *testing.T) {
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
					return keeper.MakeKeyPrefixMarketBuyerSettlementFlatFee(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "MakeKeyPrefixMarket", value: keeper.MakeKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyPrefixMarketBuyerSettlementFlatFee(%d)", tc.marketID)
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
						name:  "MakeKeyPrefixMarket",
						value: keeper.MakeKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "MakeKeyPrefixMarketBuyerSettlementFlatFee",
						value: keeper.MakeKeyPrefixMarketBuyerSettlementFlatFee(tc.marketID),
					},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketBuyerSettlementFlatFee(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyPrefixMarketBuyerSettlementRatio(t *testing.T) {
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
					return keeper.MakeKeyPrefixMarketBuyerSettlementRatio(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "MakeKeyPrefixMarket", value: keeper.MakeKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyPrefixMarketBuyerSettlementRatio(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyMarketBuyerSettlementRatio(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeBuyerSettlementRatio
	coin := func(denom string) sdk.Coin {
		return sdk.Coin{Denom: denom, Amount: sdkmath.OneInt()}
	}

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
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 0, marketTypeByte, 0},
		},
		{
			name:     "market id 1 nhash to empty",
			marketID: 1,
			ratio:    exchange.FeeRatio{Price: coin("nhash"), Fee: coin("")},
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte, 'n', 'h', 'a', 's', 'h', 0},
		},
		{
			name:     "market id 1 empty to nhash",
			marketID: 1,
			ratio:    exchange.FeeRatio{Price: coin(""), Fee: coin("nhash")},
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte, 0, 'n', 'h', 'a', 's', 'h'},
		},
		{
			name:     "market id 1 nhash to nhash",
			marketID: 1,
			ratio:    exchange.FeeRatio{Price: coin("nhash"), Fee: coin("nhash")},
			expected: []byte{keeper.KeyTypeMarket, 0, 0, 0, 1, marketTypeByte, 'n', 'h', 'a', 's', 'h', 0, 'n', 'h', 'a', 's', 'h'},
		},
		{
			name:     "market id 16,843,009 nhash to hex string",
			marketID: 16_843_009,
			ratio:    exchange.FeeRatio{Price: coin("nhash"), Fee: coin(hexString)},
			expected: append([]byte{keeper.KeyTypeMarket, 1, 1, 1, 1, marketTypeByte, 'n', 'h', 'a', 's', 'h', 0}, hexString...),
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
						name:  "MakeKeyPrefixMarket",
						value: keeper.MakeKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "MakeKeyPrefixMarketBuyerSettlementRatio",
						value: keeper.MakeKeyPrefixMarketBuyerSettlementRatio(tc.marketID),
					},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketBuyerSettlementRatio(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyMarketInactive(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeInactive

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
					return keeper.MakeKeyMarketInactive(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "MakeKeyPrefixMarket", value: keeper.MakeKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketInactive(%d)", tc.marketID)
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
					{name: "MakeKeyPrefixMarket", value: keeper.MakeKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketUserSettle(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyPrefixMarketPermissions(t *testing.T) {
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
					return keeper.MakeKeyPrefixMarketPermissions(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "MakeKeyPrefixMarket", value: keeper.MakeKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyPrefixMarketPermissions(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyPrefixMarketPermissionsForAddress(t *testing.T) {
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
					return keeper.MakeKeyPrefixMarketPermissionsForAddress(tc.marketID, tc.addr)
				},
				expected: tc.expected,
				expPanic: tc.expPanic,
			}
			if len(tc.expPanic) == 0 {
				ktc.expPrefixes = []expectedPrefix{
					{
						name:  "MakeKeyPrefixMarket",
						value: keeper.MakeKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "MakeKeyPrefixMarketPermissions",
						value: keeper.MakeKeyPrefixMarketPermissions(tc.marketID),
					},
				}
			}
			checkKey(t, ktc, "MakeKeyPrefixMarketPermissionsForAddress(%d)", tc.marketID)
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
						name:  "MakeKeyPrefixMarket",
						value: keeper.MakeKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "MakeKeyPrefixMarketPermissions",
						value: keeper.MakeKeyPrefixMarketPermissions(tc.marketID),
					},
					{
						name:  "MakeKeyPrefixMarketPermissionsForAddress",
						value: keeper.MakeKeyPrefixMarketPermissionsForAddress(tc.marketID, tc.addr),
					},
				}
			}
			checkKey(t, ktc, "MakeKeyMarketPermissions(%d)", tc.marketID)
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
					{name: "MakeKeyPrefixMarket", value: keeper.MakeKeyPrefixMarket(tc.marketID)},
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
					{name: "MakeKeyPrefixMarket", value: keeper.MakeKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketReqAttrBid(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyPrefixOrder(t *testing.T) {
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
					return keeper.MakeKeyPrefixOrder(tc.orderID)
				},
				expected: tc.expected,
			}
			checkKey(t, ktc, "MakeKeyPrefixOrder(%d)", tc.orderID)
		})
	}
}

func TestMakeKeyOrder(t *testing.T) {
	askOrder := func(orderID uint64) exchange.Order {
		return *exchange.NewOrder(orderID).WithAsk(&exchange.AskOrder{})
	}
	bidOrder := func(orderID uint64) exchange.Order {
		return *exchange.NewOrder(orderID).WithBid(&exchange.BidOrder{})
	}

	tests := []struct {
		name     string
		order    exchange.Order
		expected []byte
		expPanic string
	}{
		{
			name:     "order id 0 ask",
			order:    askOrder(0),
			expected: []byte{keeper.KeyTypeOrder, 0, 0, 0, 0, 0, 0, 0, 0, exchange.OrderTypeByteAsk},
		},
		{
			name:     "order id 0 bid",
			order:    bidOrder(0),
			expected: []byte{keeper.KeyTypeOrder, 0, 0, 0, 0, 0, 0, 0, 0, exchange.OrderTypeByteBid},
		},
		{
			name:     "order id 1 ask",
			order:    askOrder(1),
			expected: []byte{keeper.KeyTypeOrder, 0, 0, 0, 0, 0, 0, 0, 1, exchange.OrderTypeByteAsk},
		},
		{
			name:     "order id 1 bid",
			order:    bidOrder(1),
			expected: []byte{keeper.KeyTypeOrder, 0, 0, 0, 0, 0, 0, 0, 1, exchange.OrderTypeByteBid},
		},
		{
			name:     "order id 256 ask",
			order:    askOrder(256),
			expected: []byte{keeper.KeyTypeOrder, 0, 0, 0, 0, 0, 0, 1, 0, exchange.OrderTypeByteAsk},
		},
		{
			name:     "order id 256 bid",
			order:    bidOrder(256),
			expected: []byte{keeper.KeyTypeOrder, 0, 0, 0, 0, 0, 0, 1, 0, exchange.OrderTypeByteBid},
		},
		{
			name:     "order id 65,536 ask",
			order:    askOrder(65_536),
			expected: []byte{keeper.KeyTypeOrder, 0, 0, 0, 0, 0, 1, 0, 0, exchange.OrderTypeByteAsk},
		},
		{
			name:     "order id 65,536 bid",
			order:    bidOrder(65_536),
			expected: []byte{keeper.KeyTypeOrder, 0, 0, 0, 0, 0, 1, 0, 0, exchange.OrderTypeByteBid},
		},
		{
			name:     "order id 16,777,216 ask",
			order:    askOrder(16_777_216),
			expected: []byte{keeper.KeyTypeOrder, 0, 0, 0, 0, 1, 0, 0, 0, exchange.OrderTypeByteAsk},
		},
		{
			name:     "order id 16,777,216 bid",
			order:    bidOrder(16_777_216),
			expected: []byte{keeper.KeyTypeOrder, 0, 0, 0, 0, 1, 0, 0, 0, exchange.OrderTypeByteBid},
		},
		{
			name:     "order id 4,294,967,296 ask",
			order:    askOrder(4_294_967_296),
			expected: []byte{keeper.KeyTypeOrder, 0, 0, 0, 1, 0, 0, 0, 0, exchange.OrderTypeByteAsk},
		},
		{
			name:     "order id 4,294,967,296 bid",
			order:    bidOrder(4_294_967_296),
			expected: []byte{keeper.KeyTypeOrder, 0, 0, 0, 1, 0, 0, 0, 0, exchange.OrderTypeByteBid},
		},
		{
			name:     "order id 1,099,511,627,776 ask",
			order:    askOrder(1_099_511_627_776),
			expected: []byte{keeper.KeyTypeOrder, 0, 0, 1, 0, 0, 0, 0, 0, exchange.OrderTypeByteAsk},
		},
		{
			name:     "order id 1,099,511,627,776 bid",
			order:    bidOrder(1_099_511_627_776),
			expected: []byte{keeper.KeyTypeOrder, 0, 0, 1, 0, 0, 0, 0, 0, exchange.OrderTypeByteBid},
		},
		{
			name:     "order id 281,474,976,710,656 ask",
			order:    askOrder(281_474_976_710_656),
			expected: []byte{keeper.KeyTypeOrder, 0, 1, 0, 0, 0, 0, 0, 0, exchange.OrderTypeByteAsk},
		},
		{
			name:     "order id 281,474,976,710,656 bid",
			order:    bidOrder(281_474_976_710_656),
			expected: []byte{keeper.KeyTypeOrder, 0, 1, 0, 0, 0, 0, 0, 0, exchange.OrderTypeByteBid},
		},
		{
			name:     "order id 72,057,594,037,927,936 ask",
			order:    askOrder(72_057_594_037_927_936),
			expected: []byte{keeper.KeyTypeOrder, 1, 0, 0, 0, 0, 0, 0, 0, exchange.OrderTypeByteAsk},
		},
		{
			name:     "order id 72,057,594,037,927,936 bid",
			order:    bidOrder(72_057_594_037_927_936),
			expected: []byte{keeper.KeyTypeOrder, 1, 0, 0, 0, 0, 0, 0, 0, exchange.OrderTypeByteBid},
		},
		{
			name:     "order id 72,340,172,838,076,673 ask",
			order:    askOrder(72_340_172_838_076_673),
			expected: []byte{keeper.KeyTypeOrder, 1, 1, 1, 1, 1, 1, 1, 1, exchange.OrderTypeByteAsk},
		},
		{
			name:     "order id 72,340,172,838,076,673 bid",
			order:    bidOrder(72_340_172_838_076_673),
			expected: []byte{keeper.KeyTypeOrder, 1, 1, 1, 1, 1, 1, 1, 1, exchange.OrderTypeByteBid},
		},
		{
			name:     "order id 1,229,782,938,247,303,441 ask",
			order:    askOrder(1_229_782_938_247_303_441),
			expected: []byte{keeper.KeyTypeOrder, 17, 17, 17, 17, 17, 17, 17, 17, exchange.OrderTypeByteAsk},
		},
		{
			name:     "order id 1,229,782,938,247,303,441 bid",
			order:    bidOrder(1_229_782_938_247_303_441),
			expected: []byte{keeper.KeyTypeOrder, 17, 17, 17, 17, 17, 17, 17, 17, exchange.OrderTypeByteBid},
		},
		{
			name:     "order id 18,446,744,073,709,551,615 ask",
			order:    askOrder(18_446_744_073_709_551_615),
			expected: []byte{keeper.KeyTypeOrder, 255, 255, 255, 255, 255, 255, 255, 255, exchange.OrderTypeByteAsk},
		},
		{
			name:     "order id 18,446,744,073,709,551,615 bid",
			order:    bidOrder(18_446_744_073_709_551_615),
			expected: []byte{keeper.KeyTypeOrder, 255, 255, 255, 255, 255, 255, 255, 255, exchange.OrderTypeByteBid},
		},
		{
			name:     "nil inside order",
			order:    exchange.Order{OrderId: 5, Order: nil},
			expPanic: "GetOrderTypeByte() missing case for <nil>",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeKeyOrder(tc.order)
				},
				expected: tc.expected,
				expPanic: tc.expPanic,
			}
			if len(tc.expPanic) == 0 {
				ktc.expPrefixes = []expectedPrefix{
					{name: "MakeKeyPrefixOrder", value: keeper.MakeKeyPrefixOrder(tc.order.OrderId)},
				}
			}
			checkKey(t, ktc, "MakeKeyOrder %d %T", tc.order.OrderId, tc.order.Order)
		})
	}
}

func TestMakeIndexPrefixMarketToOrder(t *testing.T) {
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
					return keeper.MakeIndexPrefixMarketToOrder(tc.marketID)
				},
				expected: tc.expected,
			}
			checkKey(t, ktc, "MakeIndexPrefixMarketToOrder(%d)", tc.marketID)
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
					{name: "", value: keeper.MakeIndexPrefixMarketToOrder(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeIndexKeyMarketToOrder(%d, %d)", tc.marketID, tc.orderID)
		})
	}
}

func TestMakeIndexPrefixAddressToOrder(t *testing.T) {
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
					return keeper.MakeIndexPrefixAddressToOrder(tc.addr)
				},
				expected: tc.expected,
				expPanic: tc.expPanic,
			}
			checkKey(t, ktc, "MakeIndexPrefixAddressToOrder(%s)", string(tc.addr))
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
					{name: "", value: keeper.MakeIndexPrefixAddressToOrder(tc.addr)},
				}
			}
			checkKey(t, ktc, "MakeIndexKeyAddressToOrder(%s, %d)", string(tc.addr), tc.orderID)
		})
	}
}

func TestMakeIndexPrefixAssetToOrder(t *testing.T) {
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
					return keeper.MakeIndexPrefixAssetToOrder(tc.assetDenom)
				},
				expected: tc.expected,
			}
			checkKey(t, ktc, "MakeIndexPrefixAssetToOrder(%q)", tc.assetDenom)
		})
	}
}

func TestMakeIndexPrefixAssetToOrderAsks(t *testing.T) {
	orderKeyType := keeper.OrderKeyTypeAsk
	tests := []struct {
		name       string
		assetDenom string
		expected   []byte
	}{
		{
			name:       "empty",
			assetDenom: "",
			expected:   []byte{keeper.KeyTypeAssetToOrderIndex, orderKeyType},
		},
		{
			name:       "1 char denom",
			assetDenom: "p",
			expected:   []byte{keeper.KeyTypeAssetToOrderIndex, 'p', orderKeyType},
		},
		{
			name:       "nhash",
			assetDenom: "nhash",
			expected:   []byte{keeper.KeyTypeAssetToOrderIndex, 'n', 'h', 'a', 's', 'h', orderKeyType},
		},
		{
			name:       "hex string",
			assetDenom: hexString,
			expected: concatBz(
				[]byte{keeper.KeyTypeAssetToOrderIndex},
				[]byte(hexString),
				[]byte{orderKeyType},
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeIndexPrefixAssetToOrderAsks(tc.assetDenom)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "MakeIndexPrefixAssetToOrder", value: keeper.MakeIndexPrefixAssetToOrder(tc.assetDenom)},
				},
			}

			checkKey(t, ktc, "MakeIndexPrefixAssetToOrderAsks(%q)", tc.assetDenom)
		})
	}
}

func TestMakeIndexPrefixAssetToOrderBids(t *testing.T) {
	orderKeyType := keeper.OrderKeyTypeBid
	tests := []struct {
		name       string
		assetDenom string
		expected   []byte
	}{
		{
			name:       "empty",
			assetDenom: "",
			expected:   []byte{keeper.KeyTypeAssetToOrderIndex, orderKeyType},
		},
		{
			name:       "1 char denom",
			assetDenom: "p",
			expected:   []byte{keeper.KeyTypeAssetToOrderIndex, 'p', orderKeyType},
		},
		{
			name:       "nhash",
			assetDenom: "nhash",
			expected:   []byte{keeper.KeyTypeAssetToOrderIndex, 'n', 'h', 'a', 's', 'h', orderKeyType},
		},
		{
			name:       "hex string",
			assetDenom: hexString,
			expected: concatBz(
				[]byte{keeper.KeyTypeAssetToOrderIndex},
				[]byte(hexString),
				[]byte{orderKeyType},
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeIndexPrefixAssetToOrderBids(tc.assetDenom)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "MakeIndexPrefixAssetToOrder", value: keeper.MakeIndexPrefixAssetToOrder(tc.assetDenom)},
				},
			}

			checkKey(t, ktc, "MakeIndexPrefixAssetToOrderBids(%q)", tc.assetDenom)
		})
	}
}

func TestMakeIndexKeyAssetToOrder(t *testing.T) {
	askOrder := func(orderID uint64) exchange.Order {
		return *exchange.NewOrder(orderID).WithAsk(&exchange.AskOrder{})
	}
	bidOrder := func(orderID uint64) exchange.Order {
		return *exchange.NewOrder(orderID).WithBid(&exchange.BidOrder{})
	}

	tests := []struct {
		name       string
		assetDenom string
		order      exchange.Order
		expected   []byte
		expPanic   string
	}{
		{
			name:       "no asset order 0 ask",
			assetDenom: "",
			order:      askOrder(0),
			expected: []byte{
				keeper.KeyTypeAssetToOrderIndex,
				keeper.OrderKeyTypeAsk, 0, 0, 0, 0, 0, 0, 0, 0,
			},
		},
		{
			name:       "no asset order 0 bid",
			assetDenom: "",
			order:      bidOrder(0),
			expected: []byte{
				keeper.KeyTypeAssetToOrderIndex,
				keeper.OrderKeyTypeBid, 0, 0, 0, 0, 0, 0, 0, 0,
			},
		},
		{
			name:       "nhash order 1 ask",
			assetDenom: "nhash",
			order:      askOrder(1),
			expected: []byte{
				keeper.KeyTypeAssetToOrderIndex, 'n', 'h', 'a', 's', 'h',
				keeper.OrderKeyTypeAsk, 0, 0, 0, 0, 0, 0, 0, 1,
			},
		},
		{
			name:       "nhash order 1 bid",
			assetDenom: "nhash",
			order:      bidOrder(1),
			expected: []byte{
				keeper.KeyTypeAssetToOrderIndex, 'n', 'h', 'a', 's', 'h',
				keeper.OrderKeyTypeBid, 0, 0, 0, 0, 0, 0, 0, 1,
			},
		},
		{
			name:       "hex string order 5 ask",
			assetDenom: hexString,
			order:      askOrder(5),
			expected: concatBz(
				[]byte{keeper.KeyTypeAssetToOrderIndex}, []byte(hexString),
				[]byte{keeper.OrderKeyTypeAsk, 0, 0, 0, 0, 0, 0, 0, 5},
			),
		},
		{
			name:       "hex string order 5 bid",
			assetDenom: hexString,
			order:      bidOrder(5),
			expected: concatBz(
				[]byte{keeper.KeyTypeAssetToOrderIndex}, []byte(hexString),
				[]byte{keeper.OrderKeyTypeBid, 0, 0, 0, 0, 0, 0, 0, 5},
			),
		},
		{
			name:       "nhash order 4,294,967,296 ask",
			assetDenom: "nhash",
			order:      askOrder(4_294_967_296),
			expected: []byte{
				keeper.KeyTypeAssetToOrderIndex, 'n', 'h', 'a', 's', 'h',
				keeper.OrderKeyTypeAsk, 0, 0, 0, 1, 0, 0, 0, 0,
			},
		},
		{
			name:       "nhash order 4,294,967,296 bid",
			assetDenom: "nhash",
			order:      bidOrder(4_294_967_296),
			expected: []byte{
				keeper.KeyTypeAssetToOrderIndex, 'n', 'h', 'a', 's', 'h',
				keeper.OrderKeyTypeBid, 0, 0, 0, 1, 0, 0, 0, 0,
			},
		},
		{
			name:       "nhash order max ask",
			assetDenom: "nhash",
			order:      askOrder(18_446_744_073_709_551_615),
			expected: []byte{
				keeper.KeyTypeAssetToOrderIndex, 'n', 'h', 'a', 's', 'h',
				keeper.OrderKeyTypeAsk, 255, 255, 255, 255, 255, 255, 255, 255,
			},
		},
		{
			name:       "nhash order max bid",
			assetDenom: "nhash",
			order:      bidOrder(18_446_744_073_709_551_615),
			expected: []byte{
				keeper.KeyTypeAssetToOrderIndex, 'n', 'h', 'a', 's', 'h',
				keeper.OrderKeyTypeBid, 255, 255, 255, 255, 255, 255, 255, 255,
			},
		},
		{
			name:     "nil inside order",
			order:    exchange.Order{OrderId: 3, Order: nil},
			expPanic: "GetOrderTypeByte() missing case for <nil>",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ktc := keyTestCase{
				maker: func() []byte {
					return keeper.MakeIndexKeyAssetToOrder(tc.assetDenom, tc.order)
				},
				expected: tc.expected,
				expPanic: tc.expPanic,
			}
			if len(tc.expPanic) == 0 {
				ktc.expPrefixes = []expectedPrefix{
					{name: "MakeIndexPrefixAssetToOrder", value: keeper.MakeIndexPrefixAssetToOrder(tc.assetDenom)},
				}
				switch v := tc.order.Order.(type) {
				case *exchange.Order_AskOrder:
					ktc.expPrefixes = append(ktc.expPrefixes, expectedPrefix{
						name:  "MakeIndexPrefixAssetToOrderAsks",
						value: keeper.MakeIndexPrefixAssetToOrderAsks(tc.assetDenom),
					})
				case *exchange.Order_BidOrder:
					ktc.expPrefixes = append(ktc.expPrefixes, expectedPrefix{
						name:  "MakeIndexPrefixAssetToOrderBids",
						value: keeper.MakeIndexPrefixAssetToOrderBids(tc.assetDenom),
					})
				default:
					assert.Fail(t, "no expected prefix case defined for %T", v)
				}
			}

			checkKey(t, ktc, "MakeIndexKeyAssetToOrder(%q, %d)", tc.assetDenom, tc.order)
		})
	}
}
