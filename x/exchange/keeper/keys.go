package keeper

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"

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
// Last Market ID: 0x06 => uint32
//   This stores the last auto-selected market id.
//
// Known Market IDs: 0x07 | <market_id> => nil
//
// Markets:
//   Some aspects of a market are stored using the accounts module and the MarketAccount type.
//   Others are stored in the exchange module.
//   All market-related entries start with the type byte 0x01 followed by the <market_id>, then a market key type byte.
//   The <market_id> is a uint32 in big-endian order (4 bytes).
//
//   Market Create-Ask Flat Fee: 0x01 | <market_id> | 0x00 | <denom> => <amount> (string)
//   Market Create-Bid Flat Fee: 0x01 | <market_id> | 0x01 | <denom> => <amount> (string)
//   Market Seller Settlement Flat Fee: 0x01 | <market_id> | 0x02 | <denom> => <amount> (string)
//   Market Seller Settlement Fee Ratio: 0x01 | <market_id> | 0x03 | <price_denom> | 0x1E | <fee_denom> => price and fee amounts (strings) separated by 0x1E.
//   Market Buyer Settlement Flat Fee: 0x01 | <market_id> | 0x04 | <denom> => <amount> (string)
//   Market Buyer Settlement Fee Ratio: 0x01 | <market_id> | 0x05 | <price_denom> | 0x1E | <fee_denom> => price and fee amounts (strings) separated by 0x1E.
//   Market inactive indicator: 0x01 | <market_id> | 0x06 => nil
//   Market user-settle indicator: 0x01 | <market_id> | 0x07 => nil
//   Market permissions: 0x01 | <market_id> | 0x08 | <addr len byte> | <address> | <permission type byte> => nil
//   Market Required Attributes: 0x01 | <market_id> | 0x09 | <order_type_byte> => 0x1E-separated list of required attribute entries.
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
	// KeyTypeLastMarketID is the type byte for the last auto-selected market id.
	KeyTypeLastMarketID = byte(0x06)
	// KeyTypeKnownMarketID is the type byte for known market id entries.
	KeyTypeKnownMarketID = byte(0x07)
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

	// MarketKeyTypeCreateAskFlat is the market-specific type byte for the create-ask flat fees.
	MarketKeyTypeCreateAskFlat = byte(0x00)
	// MarketKeyTypeCreateBidFlat is the market-specific type byte for the create-bid flat fees.
	MarketKeyTypeCreateBidFlat = byte(0x01)
	// MarketKeyTypeSellerSettlementFlat is the market-specific type byte for the seller settlement flat fees.
	MarketKeyTypeSellerSettlementFlat = byte(0x02)
	// MarketKeyTypeSellerSettlementRatio is the market-specific type byte for the seller settlement ratios.
	MarketKeyTypeSellerSettlementRatio = byte(0x03)
	// MarketKeyTypeBuyerSettlementFlat is the market-specific type byte for the buyer settlement flat fees.
	MarketKeyTypeBuyerSettlementFlat = byte(0x04)
	// MarketKeyTypeBuyerSettlementRatio is the market-specific type byte for the buyer settlement ratios.
	MarketKeyTypeBuyerSettlementRatio = byte(0x05)
	// MarketKeyTypeInactive is the market-specific type byte for the inactive indicators.
	MarketKeyTypeInactive = byte(0x06)
	// MarketKeyTypeUserSettle is the market-specific type byte for the user-settle indicators.
	MarketKeyTypeUserSettle = byte(0x07)
	// MarketKeyTypePermissions is the market-specific type byte for the market permissions.
	MarketKeyTypePermissions = byte(0x08)
	// MarketKeyTypeReqAttr is the market-specific type byte for the market's required attributes lists.
	MarketKeyTypeReqAttr = byte(0x09)

	// OrderKeyTypeAsk is the order-specific type byte for ask orders.
	OrderKeyTypeAsk = exchange.OrderTypeByteAsk
	// OrderKeyTypeBid is the order-specific type byte for bid orders.
	OrderKeyTypeBid = exchange.OrderTypeByteBid

	// RecordSeparator is the RE ascii control char used to separate records in a byte slice.
	RecordSeparator = byte(0x1E)
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

// uint16FromBz converts the provided bytes into a uint16.
func uint16FromBz(bz []byte) uint16 {
	return binary.BigEndian.Uint16(bz)
}

// uint32Bz converts the provided uint32 value to a big-endian byte slice of length 4.
func uint32Bz(val uint32) []byte {
	rv := make([]byte, 4)
	binary.BigEndian.PutUint32(rv, val)
	return rv
}

// uint32FromBz converts the provided bytes into a uint32.
func uint32FromBz(bz []byte) uint32 {
	return binary.BigEndian.Uint32(bz)
}

// uint64Bz converts the provided uint64 value to a big-endian byte slice of length 8.
func uint64Bz(val uint64) []byte {
	rv := make([]byte, 8)
	binary.BigEndian.PutUint64(rv, val)
	return rv
}

// uint64FromBz converts the provided bytes into a uint64.
func uint64FromBz(bz []byte) uint64 {
	return binary.BigEndian.Uint64(bz)
}

// parseLengthPrefixedAddr extracts the length-prefixed sdk.AccAddress from the front of the provided slice.
func parseLengthPrefixedAddr(key []byte) (sdk.AccAddress, []byte, error) {
	if len(key) == 0 {
		return nil, nil, errors.New("slice is empty")
	}
	l := int(key[0])
	if l == 0 {
		return nil, nil, errors.New("length byte is zero")
	}
	if len(key) <= l {
		return nil, nil, fmt.Errorf("length byte is %d, but slice only has %d left", l, len(key)-1)
	}
	if len(key) == l+1 {
		return key[1:], nil, nil
	}
	return key[1 : l+1], key[l+1:], nil
}

// keyPrefixParamsSplit creates the key prefix for a params "split" entry.
func keyPrefixParamsSplit(extraCap int) []byte {
	return prepKey(KeyTypeParams, []byte("split"), extraCap)
}

// GetKeyPrefixParamsSplit creates the key prefix for all params splits entries.
func GetKeyPrefixParamsSplit() []byte {
	return keyPrefixParamsSplit(0)
}

// MakeKeyParamsSplit creates the key to use for the params defining the splits.
// A denom of "" is used for the default split value.
func MakeKeyParamsSplit(denom string) []byte {
	rv := keyPrefixParamsSplit(len(denom))
	rv = append(rv, denom...)
	return rv
}

// MakeKeyLastMarketID creates the key for the last auto-selected market id.
func MakeKeyLastMarketID() []byte {
	return []byte{KeyTypeLastMarketID}
}

// keyPrefixKnownMarketID creates the key prefix for a known market id entry.
func keyPrefixKnownMarketID(extraCap int) []byte {
	return prepKey(KeyTypeKnownMarketID, nil, extraCap)
}

// GetKeyPrefixKnownMarketID creates the key prefix for all known market id entries.
func GetKeyPrefixKnownMarketID() []byte {
	return keyPrefixKnownMarketID(0)
}

// MakeKeyKnownMarketID creates the key for a market's known market id entry.
func MakeKeyKnownMarketID(marketID uint32) []byte {
	suffix := uint32Bz(marketID)
	rv := keyPrefixKnownMarketID(len(suffix))
	rv = append(rv, suffix...)
	return rv
}

// ParseKeySuffixKnownMarketID parses the market id out of a known market id key that doesn't have the type byte.
// Input is expected to have the format <market id bytes>.
func ParseKeySuffixKnownMarketID(suffix []byte) uint32 {
	return uint32FromBz(suffix)
}

// keyPrefixMarket creates the root of a market's key with extra capacity for the rest.
func keyPrefixMarket(marketID uint32, extraCap int) []byte {
	return prepKey(KeyTypeMarket, uint32Bz(marketID), extraCap)
}

// GetKeyPrefixMarket creates the key prefix for all of a market's entries.
func GetKeyPrefixMarket(marketID uint32) []byte {
	return keyPrefixMarket(marketID, 0)
}

// keyPrefixMarketType creates the beginnings of a market key with the given market id and entry type byte.
// Similar to prepKey, the idea is that you can append to the result without needing a new underlying array.
func keyPrefixMarketType(marketID uint32, marketTypeByte byte, extraCap int) []byte {
	rv := keyPrefixMarket(marketID, extraCap+1)
	rv = append(rv, marketTypeByte)
	return rv
}

// marketKeyPrefixCreateAskFlatFee creates the key prefix for a market's create-ask flat fees with extra capacity for the rest.
func marketKeyPrefixCreateAskFlatFee(marketID uint32, extraCap int) []byte {
	return keyPrefixMarketType(marketID, MarketKeyTypeCreateAskFlat, extraCap)
}

// GetKeyPrefixMarketCreateAskFlatFee creates the key prefix for the create-ask flat fees for the provided market.
func GetKeyPrefixMarketCreateAskFlatFee(marketID uint32) []byte {
	return marketKeyPrefixCreateAskFlatFee(marketID, 0)
}

// MakeKeyMarketCreateAskFlatFee creates the key to use for a create-ask flat fee for the given market and denom.
func MakeKeyMarketCreateAskFlatFee(marketID uint32, denom string) []byte {
	rv := marketKeyPrefixCreateAskFlatFee(marketID, len(denom))
	rv = append(rv, denom...)
	return rv
}

// marketKeyPrefixCreateBidFlatFee creates the key prefix for a market's create-bid flat fees with extra capacity for the rest.
func marketKeyPrefixCreateBidFlatFee(marketID uint32, extraCap int) []byte {
	return keyPrefixMarketType(marketID, MarketKeyTypeCreateBidFlat, extraCap)
}

// GetKeyPrefixMarketCreateBidFlatFee creates the key prefix for the create-bid flat fees for the provided market.
func GetKeyPrefixMarketCreateBidFlatFee(marketID uint32) []byte {
	return marketKeyPrefixCreateBidFlatFee(marketID, 0)
}

// MakeKeyMarketCreateBidFlatFee creates the key to use for a create-bid flat fee for the given denom.
func MakeKeyMarketCreateBidFlatFee(marketID uint32, denom string) []byte {
	rv := marketKeyPrefixCreateBidFlatFee(marketID, len(denom))
	rv = append(rv, denom...)
	return rv
}

// marketKeyPrefixSellerSettlementFlatFee creates the key prefix for a market's seller settlement flat fees with extra capacity for the rest.
func marketKeyPrefixSellerSettlementFlatFee(marketID uint32, extraCap int) []byte {
	return keyPrefixMarketType(marketID, MarketKeyTypeSellerSettlementFlat, extraCap)
}

// GetKeyPrefixMarketSellerSettlementFlatFee creates the key prefix for a market's seller settlement flat fees.
func GetKeyPrefixMarketSellerSettlementFlatFee(marketID uint32) []byte {
	return marketKeyPrefixSellerSettlementFlatFee(marketID, 0)
}

// MakeKeyMarketSellerSettlementFlatFee creates the key for a market's seller settlement flat fee with the given denom.
func MakeKeyMarketSellerSettlementFlatFee(marketID uint32, denom string) []byte {
	rv := marketKeyPrefixSellerSettlementFlatFee(marketID, len(denom))
	rv = append(rv, denom...)
	return rv
}

// GetKeySuffixSettlementRatio gets the key suffix bytes to represent the provided fee ratio.
// Result has the format <price denom><RS><fee denom>
func GetKeySuffixSettlementRatio(ratio exchange.FeeRatio) []byte {
	rv := make([]byte, 0, len(ratio.Price.Denom)+1+len(ratio.Fee.Denom))
	rv = append(rv, ratio.Price.Denom...)
	rv = append(rv, RecordSeparator)
	rv = append(rv, ratio.Fee.Denom...)
	return rv
}

// ParseKeySuffixSettlementRatio parses the <price denom><RS><fee denom> portion
// of a settlement ratio key back into the denom strings.
func ParseKeySuffixSettlementRatio(suffix []byte) (priceDenom, feeDenom string, err error) {
	if len(suffix) == 0 {
		return "", "", errors.New("ratio key suffix is empty")
	}
	parts := strings.Split(string(suffix), string(RecordSeparator))
	if len(parts) == 2 {
		priceDenom = parts[0]
		feeDenom = parts[1]
	} else {
		err = fmt.Errorf("ratio key suffix %q has %d parts, expected 2", suffix, len(parts))
	}
	return
}

// GetFeeRatioStoreValue creates the byte slice to set in the store for a fee ratio's value.
// Result has the format <price amount><RS><fee amount> where both amounts are strings (of digits).
// E.g. "100\\1E3" (for a price amount of 100, and fee amount of 3).
func GetFeeRatioStoreValue(ratio exchange.FeeRatio) []byte {
	priceAmount := ratio.Price.Amount.String()
	feeAmount := ratio.Fee.Amount.String()
	rv := make([]byte, 0, len(priceAmount)+1+len(feeAmount))
	rv = append(rv, priceAmount...)
	rv = append(rv, RecordSeparator)
	rv = append(rv, feeAmount...)
	return rv
}

// ParseFeeRatioStoreValue parses a fee ratio's store value back into the amounts.
// Input is expected to have the format <price amount><RS><fee amount> where both amounts are strings (of digits).
// E.g. "100\\1E3" (for a price amount of 100, and fee amount of 3).
func ParseFeeRatioStoreValue(value []byte) (priceAmount, feeAmount sdkmath.Int, err error) {
	if len(value) == 0 {
		return sdkmath.ZeroInt(), sdkmath.ZeroInt(), errors.New("ratio value is empty")
	}

	parts := bytes.Split(value, []byte{RecordSeparator})
	if len(parts) == 2 {
		var ok bool
		priceAmount, ok = sdkmath.NewIntFromString(string(parts[0]))
		if !ok {
			err = fmt.Errorf("cannot convert price amount %q to sdkmath.Int", parts[0])
		}
		feeAmount, ok = sdkmath.NewIntFromString(string(parts[1]))
		if !ok {
			err = errors.Join(err, fmt.Errorf("cannot convert fee amount %q to sdkmath.Int", parts[1]))
		}
	} else {
		err = fmt.Errorf("ratio value %q has %d parts, expected 2", value, len(parts))
	}

	if err != nil {
		// The zero-value of sdkmath.Int{} will panic if anything tries to do anything with it (e.g. convert it to a string).
		// Additionally, if there was an error, we don't want either of them to have any left-over values.
		priceAmount = sdkmath.ZeroInt()
		feeAmount = sdkmath.ZeroInt()
	}

	return priceAmount, feeAmount, err
}

// marketKeyPrefixSellerSettlementRatio creates the key prefix for a market's seller settlement ratios with extra capacity for the rest.
func marketKeyPrefixSellerSettlementRatio(marketID uint32, extraCap int) []byte {
	return keyPrefixMarketType(marketID, MarketKeyTypeSellerSettlementRatio, extraCap)
}

// GetKeyPrefixMarketSellerSettlementRatio creates the key prefix for a market's seller settlement fee ratios.
func GetKeyPrefixMarketSellerSettlementRatio(marketID uint32) []byte {
	return marketKeyPrefixSellerSettlementRatio(marketID, 0)
}

// MakeKeyMarketSellerSettlementRatio creates the key to use for the given seller settlement fee ratio in the given market.
func MakeKeyMarketSellerSettlementRatio(marketID uint32, ratio exchange.FeeRatio) []byte {
	suffix := GetKeySuffixSettlementRatio(ratio)
	rv := marketKeyPrefixSellerSettlementRatio(marketID, len(suffix))
	rv = append(rv, suffix...)
	return rv
}

// marketKeyPrefixBuyerSettlementFlatFee creates the key prefix for a market's buyer settlement flat fees with extra capacity for the rest.
func marketKeyPrefixBuyerSettlementFlatFee(marketID uint32, extraCap int) []byte {
	return keyPrefixMarketType(marketID, MarketKeyTypeBuyerSettlementFlat, extraCap)
}

// GetKeyPrefixMarketBuyerSettlementFlatFee creates the key prefix for a market's buyer settlement flat fees.
func GetKeyPrefixMarketBuyerSettlementFlatFee(marketID uint32) []byte {
	return marketKeyPrefixBuyerSettlementFlatFee(marketID, 0)
}

// MakeKeyMarketBuyerSettlementFlatFee creates the key for a market's buyer settlement flat fee with the given denom.
func MakeKeyMarketBuyerSettlementFlatFee(marketID uint32, denom string) []byte {
	rv := marketKeyPrefixBuyerSettlementFlatFee(marketID, len(denom))
	rv = append(rv, denom...)
	return rv
}

// marketKeyPrefixBuyerSettlementRatio creates the key prefix for a market's buyer settlement ratios with extra capacity for the rest.
func marketKeyPrefixBuyerSettlementRatio(marketID uint32, extraCap int) []byte {
	return keyPrefixMarketType(marketID, MarketKeyTypeBuyerSettlementRatio, extraCap)
}

// GetKeyPrefixMarketBuyerSettlementRatio creates the key prefix for a market's buyer settlement fee ratios.
func GetKeyPrefixMarketBuyerSettlementRatio(marketID uint32) []byte {
	return marketKeyPrefixBuyerSettlementRatio(marketID, 0)
}

// GetKeyPrefixMarketBuyerSettlementRatioForPriceDenom creates the key prefix for a market's
// buyer settlement fee ratios that have the provided price denom.
func GetKeyPrefixMarketBuyerSettlementRatioForPriceDenom(marketID uint32, priceDenom string) []byte {
	suffix := GetKeySuffixSettlementRatio(exchange.FeeRatio{Price: sdk.Coin{Denom: priceDenom}, Fee: sdk.Coin{Denom: ""}})
	rv := marketKeyPrefixBuyerSettlementRatio(marketID, len(suffix))
	rv = append(rv, suffix...)
	return rv
}

// MakeKeyMarketBuyerSettlementRatio creates the key to use for the given buyer settlement fee ratio in the given market.
func MakeKeyMarketBuyerSettlementRatio(marketID uint32, ratio exchange.FeeRatio) []byte {
	suffix := GetKeySuffixSettlementRatio(ratio)
	rv := marketKeyPrefixBuyerSettlementRatio(marketID, len(suffix))
	rv = append(rv, suffix...)
	return rv
}

// MakeKeyMarketInactive creates the key to use to indicate that a market is inactive.
func MakeKeyMarketInactive(marketID uint32) []byte {
	return keyPrefixMarketType(marketID, MarketKeyTypeInactive, 0)
}

// MakeKeyMarketUserSettle creates the key to use to indicate that a market allows settlement by users.
func MakeKeyMarketUserSettle(marketID uint32) []byte {
	return keyPrefixMarketType(marketID, MarketKeyTypeUserSettle, 0)
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

// GetKeyPrefixMarketPermissions creates the key prefix for a market's permissions.
func GetKeyPrefixMarketPermissions(marketID uint32) []byte {
	return marketKeyPrefixPermissions(marketID, 0)
}

// GetKeyPrefixMarketPermissionsForAddress creates the key prefix for an address' permissions in a given market.
func GetKeyPrefixMarketPermissionsForAddress(marketID uint32, addr sdk.AccAddress) []byte {
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

// ParseKeySuffixMarketPermissions parses the <addr length byte><addr><permission byte> portion of a market permissions key.
func ParseKeySuffixMarketPermissions(suffix []byte) (sdk.AccAddress, exchange.Permission, error) {
	addr, remainder, err := parseLengthPrefixedAddr(suffix)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot parse address from market permissions key: %w", err)
	}
	if len(remainder) != 1 {
		return nil, 0, fmt.Errorf("cannot parse market permissions key: found %d bytes after address, expected 1", len(remainder))
	}
	return addr, exchange.Permission(remainder[0]), nil
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

// MakeKeyMarketReqAttrBid creates the key to use for a market's attributes required to create a bid order.
func MakeKeyMarketReqAttrBid(marketID uint32) []byte {
	rv := marketKeyPrefixReqAttr(marketID)
	rv = append(rv, OrderKeyTypeBid)
	return rv
}

// ParseReqAttrStoreValue parses a required attribute store value into it's string slice.
func ParseReqAttrStoreValue(value []byte) []string {
	if len(value) == 0 {
		return nil
	}
	return strings.Split(string(value), string(RecordSeparator))
}

// keyPrefixOrder creates the key prefix for orders with the provide extra capacity for additional elements.
func keyPrefixOrder(orderID uint64, extraCap int) []byte {
	return prepKey(KeyTypeOrder, uint64Bz(orderID), extraCap)
}

// GetKeyPrefixOrder creates the key prefix for the given order id.
func GetKeyPrefixOrder(orderID uint64) []byte {
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

// GetIndexKeyPrefixMarketToOrder creates the prefix for the market to order index limited ot the given market id.
func GetIndexKeyPrefixMarketToOrder(marketID uint32) []byte {
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

// GetIndexKeyPrefixAddressToOrder creates a key prefix for the address to order index limited to the given address.
func GetIndexKeyPrefixAddressToOrder(addr sdk.AccAddress) []byte {
	return indexPrefixAddressToOrder(addr, 0)
}

// MakeIndexKeyAddressToOrder creates the key to use for the address to order index with the given values.
func MakeIndexKeyAddressToOrder(addr sdk.AccAddress, orderID uint64) []byte {
	rv := indexPrefixAddressToOrder(addr, 8)
	rv = append(rv, uint64Bz(orderID)...)
	return rv
}

// indexPrefixAssetToOrder creates the prefix for the asset to order index entries with some extra space for the rest.
func indexPrefixAssetToOrder(assetDenom string, extraCap int) []byte {
	return prepKey(KeyTypeAssetToOrderIndex, []byte(assetDenom), extraCap)
}

// GetIndexKeyPrefixAssetToOrder creates a key prefix for the asset to order index limited to the given asset.
func GetIndexKeyPrefixAssetToOrder(assetDenom string) []byte {
	return indexPrefixAssetToOrder(assetDenom, 0)
}

// GetIndexKeyPrefixAssetToOrderAsks creates a key prefix for the asset to orders limited to the given asset and only ask orders.
func GetIndexKeyPrefixAssetToOrderAsks(assetDenom string) []byte {
	rv := indexPrefixAssetToOrder(assetDenom, 1)
	rv = append(rv, OrderKeyTypeAsk)
	return rv
}

// GetIndexKeyPrefixAssetToOrderBids creates a key prefix for the asset to orders limited to the given asset and only bid orders.
func GetIndexKeyPrefixAssetToOrderBids(assetDenom string) []byte {
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
