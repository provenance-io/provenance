package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/provenance-io/provenance/x/exchange/keeper"
)

// concatBz combines all provided byte slices into a single one.
func concatBz(bzs ...[]byte) []byte {
	var rv []byte
	for _, bz := range bzs {
		rv = append(rv, bz...)
	}
	return rv
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
		name  string
		denom string
		exp   []byte
	}{
		{
			name:  "empty denom",
			denom: "",
			exp:   concatBz([]byte{keeper.KeyTypeParams}, []byte("split")),
		},
		{
			name:  "nhash",
			denom: "nhash",
			exp:   concatBz([]byte{keeper.KeyTypeParams}, []byte("split"), []byte("nhash")),
		},
		{
			name:  "hex string",
			denom: "f019431c60d643b288d91a075bd4c323", // Nothing special. Got it from uuidgen.
			exp:   concatBz([]byte{keeper.KeyTypeParams}, []byte("split"), []byte("f019431c60d643b288d91a075bd4c323")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = keeper.MakeKeyParamsSplit(tc.denom)
			}
			require.NotPanics(t, testFunc, "MakeKeyParamsSplit(%q)", tc.denom)
			assert.Equal(t, tc.exp, actual, "MakeKeyParamsSplit(%q)", tc.denom)
			assert.Equal(t, len(actual), cap(actual), "length (expected) vs capacity (actual)")
		})
	}
}

func TestMakeKeyPrefixMarket(t *testing.T) {
	tests := []struct {
		name     string
		marketID uint32
		exp      []byte
	}{
		{
			name:     "market id 0",
			marketID: 0,
			exp:      []byte{keeper.KeyTypeMarket, 0, 0, 0, 0},
		},
		{
			name:     "market id 1",
			marketID: 1,
			exp:      []byte{keeper.KeyTypeMarket, 0, 0, 0, 1},
		},
		{
			name:     "market id 255",
			marketID: 255,
			exp:      []byte{keeper.KeyTypeMarket, 0, 0, 0, 255},
		},
		{
			name:     "market id 256",
			marketID: 256,
			exp:      []byte{keeper.KeyTypeMarket, 0, 0, 1, 0},
		},
		{
			name:     "market id 65_536",
			marketID: 65_536,
			exp:      []byte{keeper.KeyTypeMarket, 0, 1, 0, 0},
		},
		{
			name:     "market id 16,777,216",
			marketID: 16_777_216,
			exp:      []byte{keeper.KeyTypeMarket, 1, 0, 0, 0},
		},
		{
			name:     "market id 4,294,967,295",
			marketID: 4_294_967_295,
			exp:      []byte{keeper.KeyTypeMarket, 255, 255, 255, 255},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = keeper.MakeKeyPrefixMarket(tc.marketID)
			}
			require.NotPanics(t, testFunc, "MakeKeyPrefixMarket(%d)", tc.marketID)
			assert.Equal(t, tc.exp, actual, "MakeKeyPrefixMarket(%d)", tc.marketID)
			assert.Equal(t, len(actual), cap(actual), "length (expected) vs capacity (actual)")
		})
	}
}

// TODO[1658]: func TestMakeKeyPrefixMarketCreateAskFlatFee(t *testing.T)

// TODO[1658]: func TestMakeKeyMarketCreateAskFlatFee(t *testing.T)

// TODO[1658]: func TestMakeKeyPrefixMarketCreateBidFlatFee(t *testing.T)

// TODO[1658]: func TestMakeKeyMarketCreateBidFlatFee(t *testing.T)

// TODO[1658]: func TestMakeKeyPrefixMarketSettlementSellerFlatFee(t *testing.T)

// TODO[1658]: func TestMakeKeyMarketSettlementSellerFlatFee(t *testing.T)

// TODO[1658]: func TestMakeKeyPrefixMarketSettlementSellerRatio(t *testing.T)

// TODO[1658]: func TestMakeKeyMarketSettlementSellerRatio(t *testing.T)

// TODO[1658]: func TestMakeKeyPrefixMarketSettlementBuyerFlatFee(t *testing.T)

// TODO[1658]: func TestMakeKeyMarketSettlementBuyerFlatFee(t *testing.T)

// TODO[1658]: func TestMakeKeyPrefixMarketSettlementBuyerRatio(t *testing.T)

// TODO[1658]: func TestMakeKeyMarketSettlementBuyerRatio(t *testing.T)

// TODO[1658]: func TestMakeKeyMarketInactive(t *testing.T)

// TODO[1658]: func TestMakeKeyMarketSelfSettle(t *testing.T)

// TODO[1658]: func TestMakeKeyPrefixMarketPermissions(t *testing.T)

// TODO[1658]: func TestMakeKeyPrefixMarketPermissionsForAddress(t *testing.T)

// TODO[1658]: func TestMakeKeyMarketPermissions(t *testing.T)

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
