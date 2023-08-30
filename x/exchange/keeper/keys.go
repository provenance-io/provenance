package keeper

import (
	"encoding/binary"
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
//   Market Account Address: 0x01 | <market_id> | 0x00 => <market_account_address> (bytes)
//   Market Create Ask Flat Fee: 0x01 | <market_id> | 0x01 | <denom> => <amount> (string)
//   Market Create Bid Flat Fee: 0x01 | <market_id> | 0x02 | <denom> => <amount> (string)
//   Market Settlement Seller Flat Fee: 0x01 | <market_id> | 0x03 | <denom> => <amount> (string)
//   Market Settlement Seller Fee Ratio: 0x01 | <market_id> | 0x04 | <price_denom> | 0x00 | <fee_denom> => comma-separated price and fee amount (string).
//   Market Settlement Buyer Flat Fee: 0x01 | <market_id> | 0x05 | <denom> => <amount> (string)
//   Market Settlement Buyer Fee Ratio: 0x01 | <market_id> | 0x06 | <price_denom> | 0x00 | <fee_denom> => comma-separated price and fee amount (string).
//   Market inactive indicator: 0x01 | <market_id> | 0x07 => nil
//   Market self-settle indicator: 0x01 | <market_id> | 0x08 => nil
//   Market permissions: 0x01 | <market_id> | 0x09 | <addr len byte> | <address> | <permission type byte> => nil
//   Market Required Attributes: 0x01 | <market_id> | 0x10 | <order_type_byte> => comma-separated list of required attribute entries.
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

	// MarketKeyTypeAddress is the market-specific type byte for the market addresses.
	MarketKeyTypeAddress = byte(0x00)
	// MarketKeyTypeCreateAskFlat is the market-specific type byte for the create ask flat fees.
	MarketKeyTypeCreateAskFlat = byte(0x01)
	// MarketKeyTypeCreateBidFlat is the market-specific type byte for the create bid flat fees.
	MarketKeyTypeCreateBidFlat = byte(0x02)
	// MarketKeyTypeSettlementSellerFlat is the market-specific type byte for the seller settlement flat fees.
	MarketKeyTypeSettlementSellerFlat = byte(0x03)
	// MarketKeyTypeSettlementSellerRatio is the market-specific type byte for the seller settlement ratios.
	MarketKeyTypeSettlementSellerRatio = byte(0x04)
	// MarketKeyTypeSettlementBuyerFlat is the market-specific type byte for the buyer settlement flat fees.
	MarketKeyTypeSettlementBuyerFlat = byte(0x05)
	// MarketKeyTypeSettlementBuyerRatio is the market-specific type byte for the buyer settlement ratios.
	MarketKeyTypeSettlementBuyerRatio = byte(0x06)
	// MarketKeyTypeInactive is the market-specific type byte for the inactive indicators.
	MarketKeyTypeInactive = byte(0x07)
	// MarketKeyTypeSelfSettle is the market-specific type byte for the self-settle indicators.
	MarketKeyTypeSelfSettle = byte(0x08)
	// MarketKeyTypePermissions is the market-specific type byte for the market permissions.
	MarketKeyTypePermissions = byte(0x09)
	// MarketKeyTypeReqAttr is the market-specific type byte for the market's required attributes lists.
	MarketKeyTypeReqAttr = byte(0x10)

	// OrderKeyTypeAsk is the order-specific type byte for ask orders.
	OrderKeyTypeAsk = exchange.OrderTypeByteAsk
	// OrderKeyTypeBid is the order-specific type byte for bid orders.
	OrderKeyTypeBid = exchange.OrderTypeByteBid
)

// TODO[1658]: Add comments to the funcs in keeper/keys.go.

// TODO[1658]: Split out the market keys for prefixing.

// concatBzPlusCap creates a single byte slice consisting of the two provided byte slices with some extra capacity in the underlying array.
// The idea is that you can append(...) to the result of this without it needed a new underlying array.
func concatBzPlusCap(typeByte byte, bz1, bz2 []byte, extraCap int) []byte {
	rv := make([]byte, 0, 1+len(bz1)+len(bz2)+extraCap)
	rv = append(rv, typeByte)
	rv = append(rv, bz1...)
	rv = append(rv, bz2...)
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

func MakeKeyParamsSplit(denom string) []byte {
	return concatBzPlusCap(KeyTypeParams, []byte("split"), []byte(denom), 0)
}

func marketKeyPrefix(marketID uint32, marketTypeByte byte, extraCap int) []byte {
	return concatBzPlusCap(KeyTypeMarket, uint32Bz(marketID), []byte{marketTypeByte}, extraCap)
}

func MakeKeyMarketAccountAddress(marketID uint32) []byte {
	return marketKeyPrefix(marketID, MarketKeyTypeAddress, 0)
}

func MakeKeyMarketCreateAskFlatFee(marketID uint32, denom string) []byte {
	rv := marketKeyPrefix(marketID, MarketKeyTypeCreateAskFlat, len(denom))
	rv = append(rv, denom...)
	return rv
}

func MakeKeyMarketCreateBidFlatFee(marketID uint32, denom string) []byte {
	rv := marketKeyPrefix(marketID, MarketKeyTypeCreateBidFlat, len(denom))
	rv = append(rv, denom...)
	return rv
}

func MakeKeyMarketSettlementSellerFlatFee(marketID uint32, denom string) []byte {
	rv := marketKeyPrefix(marketID, MarketKeyTypeSettlementSellerFlat, len(denom))
	rv = append(rv, denom...)
	return rv
}

func MakeKeyMarketSettlementSellerRatio(marketID uint32, ratio exchange.FeeRatio) []byte {
	rv := marketKeyPrefix(marketID, MarketKeyTypeSettlementSellerRatio, len(ratio.Price.Denom)+1+len(ratio.Fee.Denom))
	rv = append(rv, ratio.Price.Denom...)
	rv = append(rv, 0x00)
	rv = append(rv, ratio.Fee.Denom...)
	return rv
}

func MakeKeyMarketSettlementBuyerFlatFee(marketID uint32, denom string) []byte {
	rv := marketKeyPrefix(marketID, MarketKeyTypeSettlementBuyerFlat, len(denom))
	rv = append(rv, denom...)
	return rv
}

func MakeKeyMarketSettlementBuyerRatio(marketID uint32, ratio exchange.FeeRatio) []byte {
	rv := marketKeyPrefix(marketID, MarketKeyTypeSettlementBuyerRatio, len(ratio.Price.Denom)+1+len(ratio.Fee.Denom))
	rv = append(rv, ratio.Price.Denom...)
	rv = append(rv, 0x00)
	rv = append(rv, ratio.Fee.Denom...)
	return rv
}

func MakeKeyMarketInactive(marketID uint32) []byte {
	return marketKeyPrefix(marketID, MarketKeyTypeInactive, 0)
}

func MakeKeyMarketSelfSettle(marketID uint32) []byte {
	return marketKeyPrefix(marketID, MarketKeyTypeSelfSettle, 0)
}

func MakeKeyMarketPermissions(marketID uint32, addr sdk.AccAddress, permission exchange.Permission) []byte {
	if permission < 0 || permission > 255 {
		panic(fmt.Errorf("permission value %d out of range for uint8", permission))
	}
	rv := marketKeyPrefix(marketID, MarketKeyTypePermissions, len(addr)+2)
	rv = append(rv, address.MustLengthPrefix(addr)...)
	rv = append(rv, byte(permission))
	return rv
}

func MakeKeyMarketReqAttrAsk(marketID uint32) []byte {
	rv := marketKeyPrefix(marketID, MarketKeyTypeReqAttr, 1)
	rv = append(rv, OrderKeyTypeAsk)
	return rv
}

func MakeKeyMarketReqAttrBid(marketID uint32) []byte {
	rv := marketKeyPrefix(marketID, MarketKeyTypeReqAttr, 1)
	rv = append(rv, OrderKeyTypeBid)
	return rv
}

func keyPrefixOrder(orderID uint64, extraCap int) []byte {
	return concatBzPlusCap(KeyTypeOrder, uint64Bz(orderID), nil, extraCap)
}

func MakeOrderPrefix(orderID uint64) []byte {
	return keyPrefixOrder(orderID, 0)
}

func MakeOrderKey(order exchange.Order) []byte {
	rv := keyPrefixOrder(order.GetOrderId(), 1)
	rv = append(rv, order.OrderTypeByte())
	return rv
}

func indexPrefixMarketToOrder(marketID uint32, extraCap int) []byte {
	return concatBzPlusCap(KeyTypeMarketToOrderIndex, uint32Bz(marketID), nil, extraCap)
}

func MakeIndexPrefixMarketToOrder(marketID uint32) []byte {
	return indexPrefixMarketToOrder(marketID, 0)
}

func MakeIndexKeyMarketToOrder(marketID uint32, orderID uint64) []byte {
	rv := indexPrefixMarketToOrder(marketID, 8)
	rv = append(rv, uint64Bz(orderID)...)
	return rv
}

func indexPrefixAddressToOrder(addr sdk.AccAddress, extraCap int) []byte {
	return concatBzPlusCap(KeyTypeAddressToOrderIndex, address.MustLengthPrefix(addr), nil, extraCap)
}

func MakeIndexPrefixAddressToOrder(addr sdk.AccAddress) []byte {
	return indexPrefixAddressToOrder(addr, 0)
}

func MakeIndexKeyAddressToOrder(addr sdk.AccAddress, orderID uint64) []byte {
	rv := indexPrefixAddressToOrder(addr, 8)
	rv = append(rv, uint64Bz(orderID)...)
	return rv
}

func indexPrefixAssetToOrder(assetDenom string, extraCap int) []byte {
	return concatBzPlusCap(KeyTypeAssetToOrderIndex, []byte(assetDenom), nil, extraCap)
}

func MakeIndexPrefixAssetToOrder(assetDenom string) []byte {
	return indexPrefixAssetToOrder(assetDenom, 0)
}

func MakeIndexPrefixAssetToOrderAsks(assetDenom string) []byte {
	rv := indexPrefixAssetToOrder(assetDenom, 1)
	rv = append(rv, OrderKeyTypeAsk)
	return rv
}

func MakeIndexPrefixAssetToOrderBids(assetDenom string) []byte {
	rv := indexPrefixAssetToOrder(assetDenom, 1)
	rv = append(rv, OrderKeyTypeBid)
	return rv
}

func MakeIndexKeyAssetToOrder(assetDenom string, order exchange.Order) []byte {
	rv := indexPrefixAssetToOrder(assetDenom, 9)
	rv = append(rv, order.OrderTypeByte())
	rv = append(rv, uint64Bz(order.GetOrderId())...)
	return rv
}
