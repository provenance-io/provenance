package keeper

// The following keys and values are stored in state:
//
// Params:
// All params entries start with the type byte 0x00.
// The splits are stored as uint16 in big-endian order.
// Default split: 0x00 | split => uint16
// Specific splits: 0x00 | split | <denom> => uint16
//
// Markets:
// Some aspects of a market are stored using the accounts module and the MarketAccount type.
// Others are stored in the exchange module.
// All market-related entries start with the type byte 0x01 followed by the market_id.
// The <market_id> is a uint32 in big-endian order (4 bytes).
// The <permission type byte> is a single byte as uint8 with the same value as the enum entries.
//
// Market Account Address: 0x01 | <market_id> | 0x00 => <market account address> (bytes)
// Market Create Ask Flat Fee: 0x01 | <market_id> | 0x01 | <denom> => <amount> (string)
// Market Create Bid Flat Fee: 0x01 | <market_id> | 0x02 | <denom> => <amount> (string)
// Market Settlement Seller Flat Fee: 0x01 | <market_id> | 0x03 | <denom> => <amount> (string)
// Market Settlement Seller Fee Ratio: 0x01 | <market_id> | 0x04 | <price denom> | 0x00 | <fee denom> => comma-separated price and fee amount strings.
// Market Settlement Buyer Flat Fee: 0x01 | <market_id> | 0x05 | <denom> => <amount> (string)
// Market Settlement Buyer Fee Ratio: 0x01 | <market_id> | 0x06 | <price denom> | 0x00 | <fee denom> => comma-separated price and fee amount strings.
// Market inactive indicator: 0x01 | <market_id> | 0x07 => nil
// Market self-settle indicator: 0x01 | <market_id> | 0x08 => nil
// Market permissions: 0x01 | <market_id> | 0x09 | <addr len byte> | <address> | <permission type byte> => nil
// Market Required Attributes: 0x01 | <market_id> | 0x10 | <order type byte> => comma-separated list of required attribute entries.
//
// Orders:
// Order entries all have the following general format:
//   0x02 | <order_id> | <order type byte> => protobuf encoding of order type.
// <order type byte>s:
//   Ask: 0x00
//   Bid: 0x01
// Specific entry formats:
//   Ask Orders: 0x02 | <order_id> | 0x00 => protobuf(AskOrder)
//   Bid Orders: 0x02 | <order_id> | 0x01 => protobuf(BidOrder)
//
// A market to order index is maintained with the following format:
//    0x03 | <market_id> (4 bytes) | <order_id> (8 bytes) => nil
//
// An address to order index is maintained with the following format:
//    0x04 | len(<address>) (1 byte) | <address> | <order_id> (8 bytes) => nil
//
// An asset type to order index is maintained with the following format:
//    0x05 | <denom> | <order type byte> (1 byte) | <order_id> (8 bytes) => nil

var (
	// KeyTypeParams is the type byte for params entries.
	KeyTypeParams = []byte{0x00}
	// KeyTypeMarket is the type byte for market entries.
	KeyTypeMarket = []byte{0x01}
	// KeyTypeOrder is the type byte for order entries.
	KeyTypeOrder = []byte{0x02}
	// KeyTypeMarketToOrderIndex is the type byte for entries in the market to order index.
	KeyTypeMarketToOrderIndex = []byte{0x03}
	// KeyTypeAddressToOrderIndex is the type byte for entries in the address to order index.
	KeyTypeAddressToOrderIndex = []byte{0x04}
	// KeyTypeAssetToOrderIndex is the type byte for entries in the asset to order index.
	KeyTypeAssetToOrderIndex = []byte{0x05}

	// MarketKeyTypeAddress is the market-specific type byte for the market addresses.
	MarketKeyTypeAddress = []byte{0x00}
	// MarketKeyTypeCreateAskFlat is the market-specific type byte for the create ask flat fees.
	MarketKeyTypeCreateAskFlat = []byte{0x01}
	// MarketKeyTypeCreateBidFlat is the market-specific type byte for the create bid flat fees.
	MarketKeyTypeCreateBidFlat = []byte{0x02}
	// MarketKeyTypeSettlementSellerFlat is the market-specific type byte for the seller settlement flat fees.
	MarketKeyTypeSettlementSellerFlat = []byte{0x03}
	// MarketKeyTypeSettlementSellerRatio is the market-specific type byte for the seller settlement ratios.
	MarketKeyTypeSettlementSellerRatio = []byte{0x04}
	// MarketKeyTypeSettlementBuyerFlat is the market-specific type byte for the buyer settlement flat fees.
	MarketKeyTypeSettlementBuyerFlat = []byte{0x05}
	// MarketKeyTypeSettlementBuyerRatio is the market-specific type byte for the buyer settlement ratios.
	MarketKeyTypeSettlementBuyerRatio = []byte{0x06}
	// MarketKeyTypeInactive is the market-specific type byte for the inactive indicators.
	MarketKeyTypeInactive = []byte{0x07}
	// MarketKeyTypeSelfSettle is the market-specific type byte for the self-settle indicators.
	MarketKeyTypeSelfSettle = []byte{0x08}
	// MarketKeyTypePermissions is the market-specific type byte for the market permissions.
	MarketKeyTypePermissions = []byte{0x09}
	// MarketKeyTypeReqAttr is the market-specific type byte for the market's required attributes lists.
	MarketKeyTypeReqAttr = []byte{0x10}

	OrderKeyTypeAsk = []byte{0x00}
	OrderKeyTypeBid = []byte{0x01}
)
