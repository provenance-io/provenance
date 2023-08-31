package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	var actual []byte
	testFunc := func() {
		actual = tc.maker()
	}
	// TODO[1658]: Replace this with a testutils panic checker.
	if len(tc.expPanic) > 0 {
		require.PanicsWithErrorf(t, tc.expPanic, testFunc, msg, args...)
	} else {
		require.NotPanicsf(t, testFunc, msg, args...)
	}
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
				{name: "MarketKeyTypeSettlementSellerFlat", value: keeper.MarketKeyTypeSettlementSellerFlat},
				{name: "MarketKeyTypeSettlementSellerRatio", value: keeper.MarketKeyTypeSettlementSellerRatio},
				{name: "MarketKeyTypeSettlementBuyerFlat", value: keeper.MarketKeyTypeSettlementBuyerFlat},
				{name: "MarketKeyTypeSettlementBuyerRatio", value: keeper.MarketKeyTypeSettlementBuyerRatio},
				{name: "MarketKeyTypeInactive", value: keeper.MarketKeyTypeInactive},
				{name: "MarketKeyTypeSelfSettle", value: keeper.MarketKeyTypeSelfSettle},
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

func TestMakeKeyPrefixMarketSettlementSellerFlatFee(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeSettlementSellerFlat

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
					return keeper.MakeKeyPrefixMarketSettlementSellerFlatFee(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "MakeKeyPrefixMarket", value: keeper.MakeKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyPrefixMarketSettlementSellerFlatFee(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyMarketSettlementSellerFlatFee(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeSettlementSellerFlat

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
					return keeper.MakeKeyMarketSettlementSellerFlatFee(tc.marketID, tc.denom)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{
						name:  "MakeKeyPrefixMarket",
						value: keeper.MakeKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "MakeKeyPrefixMarketSettlementSellerFlatFee",
						value: keeper.MakeKeyPrefixMarketSettlementSellerFlatFee(tc.marketID),
					},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketSettlementSellerFlatFee(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyPrefixMarketSettlementSellerRatio(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeSettlementSellerRatio

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
					return keeper.MakeKeyPrefixMarketSettlementSellerRatio(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "MakeKeyPrefixMarket", value: keeper.MakeKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyPrefixMarketSettlementSellerRatio(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyMarketSettlementSellerRatio(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeSettlementSellerRatio
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
					return keeper.MakeKeyMarketSettlementSellerRatio(tc.marketID, tc.ratio)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{
						name:  "MakeKeyPrefixMarket",
						value: keeper.MakeKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "MakeKeyPrefixMarketSettlementSellerRatio",
						value: keeper.MakeKeyPrefixMarketSettlementSellerRatio(tc.marketID),
					},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketSettlementSellerRatio(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyPrefixMarketSettlementBuyerFlatFee(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeSettlementBuyerFlat

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
					return keeper.MakeKeyPrefixMarketSettlementBuyerFlatFee(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "MakeKeyPrefixMarket", value: keeper.MakeKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyPrefixMarketSettlementBuyerFlatFee(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyMarketSettlementBuyerFlatFee(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeSettlementBuyerFlat

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
					return keeper.MakeKeyMarketSettlementBuyerFlatFee(tc.marketID, tc.denom)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{
						name:  "MakeKeyPrefixMarket",
						value: keeper.MakeKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "MakeKeyPrefixMarketSettlementBuyerFlatFee",
						value: keeper.MakeKeyPrefixMarketSettlementBuyerFlatFee(tc.marketID),
					},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketSettlementBuyerFlatFee(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyPrefixMarketSettlementBuyerRatio(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeSettlementBuyerRatio

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
					return keeper.MakeKeyPrefixMarketSettlementBuyerRatio(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "MakeKeyPrefixMarket", value: keeper.MakeKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyPrefixMarketSettlementBuyerRatio(%d)", tc.marketID)
		})
	}
}

func TestMakeKeyMarketSettlementBuyerRatio(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeSettlementBuyerRatio
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
					return keeper.MakeKeyMarketSettlementBuyerRatio(tc.marketID, tc.ratio)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{
						name:  "MakeKeyPrefixMarket",
						value: keeper.MakeKeyPrefixMarket(tc.marketID),
					},
					{
						name:  "MakeKeyPrefixMarketSettlementBuyerRatio",
						value: keeper.MakeKeyPrefixMarketSettlementBuyerRatio(tc.marketID),
					},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketSettlementBuyerRatio(%d)", tc.marketID)
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

func TestMakeKeyMarketSelfSettle(t *testing.T) {
	marketTypeByte := keeper.MarketKeyTypeSelfSettle

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
					return keeper.MakeKeyMarketSelfSettle(tc.marketID)
				},
				expected: tc.expected,
				expPrefixes: []expectedPrefix{
					{name: "MakeKeyPrefixMarket", value: keeper.MakeKeyPrefixMarket(tc.marketID)},
				},
			}
			checkKey(t, ktc, "MakeKeyMarketSelfSettle(%d)", tc.marketID)
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
			name:     "market id 0 nil addr",
			marketID: 0,
			addr:     nil,
			expPanic: "empty address not allowed",
		},
		{
			name:     "market id 0 empty addr",
			marketID: 0,
			addr:     sdk.AccAddress{},
			expPanic: "empty address not allowed",
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
			name:     "market id 0 nil addr",
			marketID: 0,
			addr:     nil,
			expPanic: "empty address not allowed",
		},
		{
			name:     "market id 0 empty addr",
			marketID: 0,
			addr:     sdk.AccAddress{},
			expPanic: "empty address not allowed",
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
		}, {
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

// TODO[1658]: func TestMakeKeyMarketReqAttrAsk(t *testing.T)

// TODO[1658]: func TestMakeKeyMarketReqAttrBid(t *testing.T)

// TODO[1658]: func TestMakeKeyPrefixOrder(t *testing.T)

// TODO[1658]: func TestMakeKeyOrder(t *testing.T)

// TODO[1658]: func TestMakeIndexPrefixMarketToOrder(t *testing.T)

// TODO[1658]: func TestMakeIndexKeyMarketToOrder(t *testing.T)

// TODO[1658]: func TestMakeIndexPrefixAddressToOrder(t *testing.T)

// TODO[1658]: func TestMakeIndexKeyAddressToOrder(t *testing.T)

// TODO[1658]: func TestMakeIndexPrefixAssetToOrder(t *testing.T)

// TODO[1658]: func TestMakeIndexPrefixAssetToOrderAsks(t *testing.T)

// TODO[1658]: func TestMakeIndexPrefixAssetToOrderBids(t *testing.T)

// TODO[1658]: func TestMakeIndexKeyAssetToOrder(t *testing.T)
