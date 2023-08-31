package keeper

import (
	"encoding/binary"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/provenance-io/provenance/x/exchange"
)

// The following keys and values are stored in state:
//
// Params:
//   All params entries start with the type byte 0x00.
//   The splits are stored as uint16 in big-endian order.
//   Default split: 0x00 | "split" => uint16
//   Specific splits: 0x00 | "split" | <denom> => uint16
//
// Markets:
//   Some aspects of a market are stored using the accounts module and the MarketAccount type.
//   Others are stored in the exchange module.
//   All market-related entries start with the type byte 0x01 followed by the <market_id>, then a market key type byte.
//   The <market_id> is a uint32 in big-endian order (4 bytes).
//
//   Market Create Ask Flat Fee: 0x01 | <market_id> | 0x00 | <denom> => <amount> (string)
//   Market Create Bid Flat Fee: 0x01 | <market_id> | 0x01 | <denom> => <amount> (string)
//   Market Settlement Seller Flat Fee: 0x01 | <market_id> | 0x02 | <denom> => <amount> (string)
//   Market Settlement Seller Fee Ratio: 0x01 | <market_id> | 0x03 | <price_denom> | 0x00 | <fee_denom> => comma-separated price and fee amount (string).
//   Market Settlement Buyer Flat Fee: 0x01 | <market_id> | 0x04 | <denom> => <amount> (string)
//   Market Settlement Buyer Fee Ratio: 0x01 | <market_id> | 0x05 | <price_denom> | 0x00 | <fee_denom> => comma-separated price and fee amount (string).
//   Market inactive indicator: 0x01 | <market_id> | 0x06 => nil
//   Market self-settle indicator: 0x01 | <market_id> | 0x07 => nil
//   Market permissions: 0x01 | <market_id> | 0x08 | <addr len byte> | <address> | <permission type byte> => nil
//   Market Required Attributes: 0x01 | <market_id> | 0x09 | <order_type_byte> => comma-separated list of required attribute entries.
//
//   The <permission_type_byte> is a single byte as uint8 with the same values as the enum entries.
//
// Orders:
//   Order entries all have the following general format:
//     0x02 | <order_id> (8 bytes) | <order_type_byte> => protobuf encoding of specific order type.
//   The <order_id> is a uint64 in big-endian order (8 bytes).
//   <order_type_byte> values:
//     Ask: 0x00
//     Bid: 0x01
//   So, the specific entry formats look like this:
//     Ask Orders: 0x02 | <order_id> (8 bytes) | 0x00 => protobuf(AskOrder)
//     Bid Orders: 0x02 | <order_id> (8 bytes) | 0x01 => protobuf(BidOrder)
//
// A market to order index is maintained with the following format:
//    0x03 | <market_id> (4 bytes) | <order_id> (8 bytes) => nil
//
// An address to order index is maintained with the following format:
//    0x04 | len(<address>) (1 byte) | <address> | <order_id> (8 bytes) => nil
//
// An asset type to order index is maintained with the following format:
//    0x05 | <asset_denom> | <order_type_byte> (1 byte) | <order_id> (8 bytes) => nil

const (
	// KeyTypeParams is the type byte for params entries.
	KeyTypeParams = byte(0x00)
	// KeyTypeMarket is the type byte for market entries.
	KeyTypeMarket = byte(0x01)
	// KeyTypeOrder is the type byte for order entries.
	KeyTypeOrder = byte(0x02)
	// KeyTypeMarketToOrderIndex is the type byte for entries in the market to order index.
	KeyTypeMarketToOrderIndex = byte(0x03)
	// KeyTypeAddressToOrderIndex is the type byte for entries in the address to order index.
	KeyTypeAddressToOrderIndex = byte(0x04)
	// KeyTypeAssetToOrderIndex is the type byte for entries in the asset to order index.
	KeyTypeAssetToOrderIndex = byte(0x05)

	// MarketKeyTypeCreateAskFlat is the market-specific type byte for the create ask flat fees.
	MarketKeyTypeCreateAskFlat = byte(0x00)
	// MarketKeyTypeCreateBidFlat is the market-specific type byte for the create bid flat fees.
	MarketKeyTypeCreateBidFlat = byte(0x01)
	// MarketKeyTypeSettlementSellerFlat is the market-specific type byte for the seller settlement flat fees.
	MarketKeyTypeSettlementSellerFlat = byte(0x02)
	// MarketKeyTypeSettlementSellerRatio is the market-specific type byte for the seller settlement ratios.
	MarketKeyTypeSettlementSellerRatio = byte(0x03)
	// MarketKeyTypeSettlementBuyerFlat is the market-specific type byte for the buyer settlement flat fees.
	MarketKeyTypeSettlementBuyerFlat = byte(0x04)
	// MarketKeyTypeSettlementBuyerRatio is the market-specific type byte for the buyer settlement ratios.
	MarketKeyTypeSettlementBuyerRatio = byte(0x05)
	// MarketKeyTypeInactive is the market-specific type byte for the inactive indicators.
	MarketKeyTypeInactive = byte(0x06)
	// MarketKeyTypeSelfSettle is the market-specific type byte for the self-settle indicators.
	MarketKeyTypeSelfSettle = byte(0x07)
	// MarketKeyTypePermissions is the market-specific type byte for the market permissions.
	MarketKeyTypePermissions = byte(0x08)
	// MarketKeyTypeReqAttr is the market-specific type byte for the market's required attributes lists.
	MarketKeyTypeReqAttr = byte(0x09)

	// OrderKeyTypeAsk is the order-specific type byte for ask orders.
	OrderKeyTypeAsk = exchange.OrderTypeByteAsk
	// OrderKeyTypeBid is the order-specific type byte for bid orders.
	OrderKeyTypeBid = exchange.OrderTypeByteBid
)

// prepKey creates a single byte slice consisting of the type byte and provided byte slice with some extra capacity in the underlying array.
// The idea is that you can append(...) to the result of this without it needed a new underlying array.
func prepKey(typeByte byte, bz []byte, extraCap int) []byte {
	rv := make([]byte, 0, 1+len(bz)+extraCap)
	rv = append(rv, typeByte)
	rv = append(rv, bz...)
	return rv
}

// uint16Bz converts the provided uint16 value to a big-endian byte slice of length 2.
func uint16Bz(val uint16) []byte {
	rv := make([]byte, 2)
	binary.BigEndian.PutUint16(rv, val)
	return rv
}

// uint32Bz converts the provided uint32 value to a big-endian byte slice of length 4.
func uint32Bz(val uint32) []byte {
	rv := make([]byte, 4)
	binary.BigEndian.PutUint32(rv, val)
	return rv
}

// uint64Bz converts the provided uint64 value to a big-endian byte slice of length 8.
func uint64Bz(val uint64) []byte {
	rv := make([]byte, 8)
	binary.BigEndian.PutUint64(rv, val)
	return rv
}

// MakeKeyParamsSplit creates the key to use for the params defining the splits.
// A denom of "" is used for the default split value.
func MakeKeyParamsSplit(denom string) []byte {
	rv := prepKey(KeyTypeParams, []byte("split"), len(denom))
	rv = append(rv, denom...)
	return rv
}

// keyPrefixMarket creates the root of a market's key with extra capacity for the rest.
func keyPrefixMarket(marketID uint32, extraCap int) []byte {
	return prepKey(KeyTypeMarket, uint32Bz(marketID), extraCap)
}

// MakeKeyPrefixMarket creates the key prefix for all of a market's entries.
func MakeKeyPrefixMarket(marketID uint32) []byte {
	return keyPrefixMarket(marketID, 0)
}

// keyPrefixMarketType creates the beginnings of a market key with the given market id and entry type byte.
// Similar to prepKey, the idea is that you can append to the result without needing a new underlying array.
func keyPrefixMarketType(marketID uint32, marketTypeByte byte, extraCap int) []byte {
	rv := keyPrefixMarket(marketID, extraCap+1)
	rv = append(rv, marketTypeByte)
	return rv
}

// marketKeyPrefixCreateAskFlatFee creates the key prefix for a market's create ask flat fees with extra capacity for the rest.
func marketKeyPrefixCreateAskFlatFee(marketID uint32, extraCap int) []byte {
	return keyPrefixMarketType(marketID, MarketKeyTypeCreateAskFlat, extraCap)
}

// MakeKeyPrefixMarketCreateAskFlatFee creates the key prefix for the create ask flat fees for the provided market.
func MakeKeyPrefixMarketCreateAskFlatFee(marketID uint32) []byte {
	return marketKeyPrefixCreateAskFlatFee(marketID, 0)
}

// MakeKeyMarketCreateAskFlatFee creates the key to use for a create ask flat fee for the given market and denom.
func MakeKeyMarketCreateAskFlatFee(marketID uint32, denom string) []byte {
	rv := marketKeyPrefixCreateAskFlatFee(marketID, len(denom))
	rv = append(rv, denom...)
	return rv
}

// marketKeyPrefixCreateBidFlatFee creates the key prefix for a market's create bid flat fees with extra capacity for the rest.
func marketKeyPrefixCreateBidFlatFee(marketID uint32, extraCap int) []byte {
	return keyPrefixMarketType(marketID, MarketKeyTypeCreateBidFlat, extraCap)
}

// MakeKeyPrefixMarketCreateBidFlatFee creates the key prefix for the create bid flat fees for the provided market.
func MakeKeyPrefixMarketCreateBidFlatFee(marketID uint32) []byte {
	return marketKeyPrefixCreateBidFlatFee(marketID, 0)
}

// MakeKeyMarketCreateBidFlatFee creates the key to use for a create bid flat fee for the given denom.
func MakeKeyMarketCreateBidFlatFee(marketID uint32, denom string) []byte {
	rv := marketKeyPrefixCreateBidFlatFee(marketID, len(denom))
	rv = append(rv, denom...)
	return rv
}

// marketKeyPrefixSettlementSellerFlatFee creates the key prefix for a market's settlement seller flat fees with extra capacity for the rest.
func marketKeyPrefixSettlementSellerFlatFee(marketID uint32, extraCap int) []byte {
	return keyPrefixMarketType(marketID, MarketKeyTypeSettlementSellerFlat, extraCap)
}

// MakeKeyPrefixMarketSettlementSellerFlatFee creates the key prefix for a market's settlement seller flat fees.
func MakeKeyPrefixMarketSettlementSellerFlatFee(marketID uint32) []byte {
	return marketKeyPrefixSettlementSellerFlatFee(marketID, 0)
}

// MakeKeyMarketSettlementSellerFlatFee creates the key for a market's settlement seller flat fee with the given denom.
func MakeKeyMarketSettlementSellerFlatFee(marketID uint32, denom string) []byte {
	rv := marketKeyPrefixSettlementSellerFlatFee(marketID, len(denom))
	rv = append(rv, denom...)
	return rv
}

// marketKeyPrefixSettlementSellerRatio creates the key prefix for a market's settlement seller ratios with extra capacity for the rest.
func marketKeyPrefixSettlementSellerRatio(marketID uint32, extraCap int) []byte {
	return keyPrefixMarketType(marketID, MarketKeyTypeSettlementSellerRatio, extraCap)
}

// MakeKeyPrefixMarketSettlementSellerRatio creates the key prefix for a market's settlement seller fee ratios.
func MakeKeyPrefixMarketSettlementSellerRatio(marketID uint32) []byte {
	return marketKeyPrefixSettlementSellerRatio(marketID, 0)
}

// MakeKeyMarketSettlementSellerRatio creates the key to use for the given settlement seller fee ratio in the given market.
func MakeKeyMarketSettlementSellerRatio(marketID uint32, ratio exchange.FeeRatio) []byte {
	rv := marketKeyPrefixSettlementSellerRatio(marketID, len(ratio.Price.Denom)+1+len(ratio.Fee.Denom))
	rv = append(rv, ratio.Price.Denom...)
	rv = append(rv, 0x00)
	rv = append(rv, ratio.Fee.Denom...)
	return rv
}

// marketKeyPrefixSettlementBuyerFlatFee creates the key prefix for a market's settlement buyer flat fees with extra capacity for the rest.
func marketKeyPrefixSettlementBuyerFlatFee(marketID uint32, extraCap int) []byte {
	return keyPrefixMarketType(marketID, MarketKeyTypeSettlementBuyerFlat, extraCap)
}

// MakeKeyPrefixMarketSettlementBuyerFlatFee creates the key prefix for a market's settlement buyer flat fees.
func MakeKeyPrefixMarketSettlementBuyerFlatFee(marketID uint32) []byte {
	return marketKeyPrefixSettlementBuyerFlatFee(marketID, 0)
}

// MakeKeyMarketSettlementBuyerFlatFee creates th ekey for a market's settlement buyer flat fee with the given denom.
func MakeKeyMarketSettlementBuyerFlatFee(marketID uint32, denom string) []byte {
	rv := marketKeyPrefixSettlementBuyerFlatFee(marketID, len(denom))
	rv = append(rv, denom...)
	return rv
}

// marketKeyPrefixSettlementBuyerRatio creates the key prefix for a market's settlement buyer ratios with extra capacity for the rest.
func marketKeyPrefixSettlementBuyerRatio(marketID uint32, extraCap int) []byte {
	return keyPrefixMarketType(marketID, MarketKeyTypeSettlementBuyerRatio, extraCap)
}

// MakeKeyPrefixMarketSettlementBuyerRatio creates the key prefix for a market's settlement buyer fee ratios.
func MakeKeyPrefixMarketSettlementBuyerRatio(marketID uint32) []byte {
	return marketKeyPrefixSettlementBuyerRatio(marketID, 0)
}

// MakeKeyMarketSettlementBuyerRatio creates the key to use for the given settlement buyer fee ratio in the given market.
func MakeKeyMarketSettlementBuyerRatio(marketID uint32, ratio exchange.FeeRatio) []byte {
	rv := marketKeyPrefixSettlementBuyerRatio(marketID, len(ratio.Price.Denom)+1+len(ratio.Fee.Denom))
	rv = append(rv, ratio.Price.Denom...)
	rv = append(rv, 0x00)
	rv = append(rv, ratio.Fee.Denom...)
	return rv
}

// MakeKeyMarketInactive creates the key to use to indicate that a market is inactive.
func MakeKeyMarketInactive(marketID uint32) []byte {
	return keyPrefixMarketType(marketID, MarketKeyTypeInactive, 0)
}

// MakeKeyMarketSelfSettle creates the key to use to indicate that a market allows self-settlement.
func MakeKeyMarketSelfSettle(marketID uint32) []byte {
	return keyPrefixMarketType(marketID, MarketKeyTypeSelfSettle, 0)
}

// marketKeyPrefixPermissions creates the key prefix for a market's permissions with extra capacity for the rest.
func marketKeyPrefixPermissions(marketID uint32, extraCap int) []byte {
	return keyPrefixMarketType(marketID, MarketKeyTypePermissions, extraCap)
}

// marketKeyPrefixPermissionsForAddress creates the key prefix for an address' permissions in a given market with extra capacity for the rest.
func marketKeyPrefixPermissionsForAddress(marketID uint32, addr sdk.AccAddress, extraCap int) []byte {
	if len(addr) == 0 {
		panic(errors.New("empty address not allowed"))
	}
	rv := marketKeyPrefixPermissions(marketID, 1+len(addr)+extraCap)
	rv = append(rv, address.MustLengthPrefix(addr)...)
	return rv
}

// MakeKeyPrefixMarketPermissions creates the key prefix for a market's permissions.
func MakeKeyPrefixMarketPermissions(marketID uint32) []byte {
	return marketKeyPrefixPermissions(marketID, 0)
}

// MakeKeyPrefixMarketPermissionsForAddress creates the key prefix for an address' permissions in a given market.
func MakeKeyPrefixMarketPermissionsForAddress(marketID uint32, addr sdk.AccAddress) []byte {
	return marketKeyPrefixPermissionsForAddress(marketID, addr, 0)
}

// MakeKeyMarketPermissions creates the key to use for a permission granted to an address for a market.
func MakeKeyMarketPermissions(marketID uint32, addr sdk.AccAddress, permission exchange.Permission) []byte {
	if permission < 0 || permission > 255 {
		panic(fmt.Errorf("permission value %d out of range for uint8", permission))
	}
	rv := marketKeyPrefixPermissionsForAddress(marketID, addr, 1)
	rv = append(rv, byte(permission))
	return rv
}

// marketKeyPrefixReqAttr creates the key prefix for a market's required attributes entries with an extra byte of capacity for the order type.
func marketKeyPrefixReqAttr(marketID uint32) []byte {
	return keyPrefixMarketType(marketID, MarketKeyTypeReqAttr, 1)
}

// MakeKeyMarketReqAttrAsk creates the key to use for a market's attributes required to create an ask order.
func MakeKeyMarketReqAttrAsk(marketID uint32) []byte {
	rv := marketKeyPrefixReqAttr(marketID)
	rv = append(rv, OrderKeyTypeAsk)
	return rv
}

// MakeKeyMarketReqAttrBid creates the key to use for a market's attributes required to create an bid order.
func MakeKeyMarketReqAttrBid(marketID uint32) []byte {
	rv := marketKeyPrefixReqAttr(marketID)
	rv = append(rv, OrderKeyTypeBid)
	return rv
}

// keyPrefixOrder creates the key prefix for orders with the provide extra capacity for additional elements.
func keyPrefixOrder(orderID uint64, extraCap int) []byte {
	return prepKey(KeyTypeOrder, uint64Bz(orderID), extraCap)
}

// MakeKeyPrefixOrder creates the key prefix for the given order id.
func MakeKeyPrefixOrder(orderID uint64) []byte {
	return keyPrefixOrder(orderID, 0)
}

// MakeKeyOrder makes the key to use for the given order.
func MakeKeyOrder(order exchange.Order) []byte {
	rv := keyPrefixOrder(order.GetOrderId(), 1)
	rv = append(rv, order.GetOrderTypeByte())
	return rv
}

// indexPrefixMarketToOrder creates the prefix for the market to order prefix entries with some extra space for the rest.
func indexPrefixMarketToOrder(marketID uint32, extraCap int) []byte {
	return prepKey(KeyTypeMarketToOrderIndex, uint32Bz(marketID), extraCap)
}

// MakeIndexPrefixMarketToOrder creates the prefix for the market to order index limited ot the given market id.
func MakeIndexPrefixMarketToOrder(marketID uint32) []byte {
	return indexPrefixMarketToOrder(marketID, 0)
}

// MakeIndexKeyMarketToOrder creates the key to use for the market to order index with the given ids.
func MakeIndexKeyMarketToOrder(marketID uint32, orderID uint64) []byte {
	rv := indexPrefixMarketToOrder(marketID, 8)
	rv = append(rv, uint64Bz(orderID)...)
	return rv
}

// indexPrefixAddressToOrder creates the prefix for the address to order index entries with some extra apace for the rest.
func indexPrefixAddressToOrder(addr sdk.AccAddress, extraCap int) []byte {
	if len(addr) == 0 {
		panic(errors.New("empty address not allowed"))
	}
	return prepKey(KeyTypeAddressToOrderIndex, address.MustLengthPrefix(addr), extraCap)
}

// MakeIndexPrefixAddressToOrder creates a key prefix for the address to order index limited to the given address.
func MakeIndexPrefixAddressToOrder(addr sdk.AccAddress) []byte {
	return indexPrefixAddressToOrder(addr, 0)
}

// MakeIndexKeyAddressToOrder creates the key to use for the address to order index with the given values.
func MakeIndexKeyAddressToOrder(addr sdk.AccAddress, orderID uint64) []byte {
	rv := indexPrefixAddressToOrder(addr, 8)
	rv = append(rv, uint64Bz(orderID)...)
	return rv
}

// indexPrefixAssetToOrder creates the prefix for the asset to order index enties with some extra space for the rest.
func indexPrefixAssetToOrder(assetDenom string, extraCap int) []byte {
	return prepKey(KeyTypeAssetToOrderIndex, []byte(assetDenom), extraCap)
}

// MakeIndexPrefixAssetToOrder creates a key prefix for the asset to order index limited to the given asset.
func MakeIndexPrefixAssetToOrder(assetDenom string) []byte {
	return indexPrefixAssetToOrder(assetDenom, 0)
}

// MakeIndexPrefixAssetToOrderAsks creates a key prefix for the asset to orders limited to the given asset and only ask orders.
func MakeIndexPrefixAssetToOrderAsks(assetDenom string) []byte {
	rv := indexPrefixAssetToOrder(assetDenom, 1)
	rv = append(rv, OrderKeyTypeAsk)
	return rv
}

// MakeIndexPrefixAssetToOrderBids creates a key prefix for the asset to orders limited to the given asset and only bid orders.
func MakeIndexPrefixAssetToOrderBids(assetDenom string) []byte {
	rv := indexPrefixAssetToOrder(assetDenom, 1)
	rv = append(rv, OrderKeyTypeBid)
	return rv
}

// MakeIndexKeyAssetToOrder creates the key to use for the asset to order index for the provided values.
func MakeIndexKeyAssetToOrder(assetDenom string, order exchange.Order) []byte {
	rv := indexPrefixAssetToOrder(assetDenom, 9)
	rv = append(rv, order.GetOrderTypeByte())
	rv = append(rv, uint64Bz(order.GetOrderId())...)
	return rv
}
