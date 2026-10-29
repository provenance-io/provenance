package exchange

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	collcodec "cosmossdk.io/collections/codec"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/gogoproto/proto"
)

// RegisterInterfaces registers concrete implementations for this module.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	messages := make([]proto.Message, len(AllRequestMsgs))
	copy(messages, AllRequestMsgs)
	registry.RegisterImplementations((*sdk.Msg)(nil), messages...)

	registry.RegisterInterface(
		"provenance.exchange.v1.MarketAccount",
		(*sdk.AccountI)(nil),
		&MarketAccount{},
	)

	registry.RegisterInterface(
		"provenance.exchange.v1.MarketAccount",
		(*authtypes.GenesisAccount)(nil),
		&MarketAccount{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// RecordSeparator is the RE ascii control char used to separate records in a byte slice.
const RecordSeparator = byte(0x1E)

// The exchange module's collections all use standard key codecs. These are the value codecs
// that the collections and SDK packages don't provide.
var (
	// ByteValue stores a single byte. It's used for the order indexes, whose values are the
	// order's type byte. These are the same bytes used before collections.
	ByteValue collcodec.ValueCodec[byte] = byteValueCodec{}
	// CoinsValue stores coins as their string form, e.g. "10nhash,3usd".
	// These are the same bytes used for commitment amounts before collections.
	CoinsValue collcodec.ValueCodec[sdk.Coins] = coinsValueCodec{}
	// StringSliceValue stores a list of strings joined by the RecordSeparator (0x1E).
	// It's used for a market's required attributes, which can never contain that character.
	// These are the same value bytes used before collections.
	StringSliceValue collcodec.ValueCodec[[]string] = stringSliceValueCodec{}
)

// NewOrderValueCodec creates the value codec used for orders. It stores an order as
// <order type byte> | <protobuf encoding of the AskOrder or BidOrder>. These are the same bytes
// used before collections, so existing orders don't need to be rewritten.
//
// The order id is the map key, so it isn't part of the value: Decode returns an Order with an
// OrderId of zero, and it's up to the caller to fill it in from the key.
func NewOrderValueCodec(cdc codec.BinaryCodec) collcodec.ValueCodec[Order] {
	return orderValueCodec{cdc: cdc}
}

// orderValueCodec is the value codec for orders. See NewOrderValueCodec.
type orderValueCodec struct {
	cdc codec.BinaryCodec
}

// Encode converts the order into its store value.
func (c orderValueCodec) Encode(order Order) ([]byte, error) {
	switch {
	case order.GetAskOrder() != nil:
		data, err := c.cdc.Marshal(order.GetAskOrder())
		if err != nil {
			return nil, fmt.Errorf("error marshaling ask order: %w", err)
		}
		return append([]byte{OrderTypeByteAsk}, data...), nil
	case order.GetBidOrder() != nil:
		data, err := c.cdc.Marshal(order.GetBidOrder())
		if err != nil {
			return nil, fmt.Errorf("error marshaling bid order: %w", err)
		}
		return append([]byte{OrderTypeByteBid}, data...), nil
	default:
		return nil, fmt.Errorf("cannot encode order %d: unknown order type %T", order.OrderId, order.GetOrder())
	}
}

// Decode converts a store value back into an order. The returned order has an OrderId of zero.
func (c orderValueCodec) Decode(bz []byte) (Order, error) {
	if len(bz) == 0 {
		return Order{}, errors.New("order value is empty")
	}
	typeByte, data := bz[0], bz[1:]
	switch typeByte {
	case OrderTypeByteAsk:
		var ask AskOrder
		if err := c.cdc.Unmarshal(data, &ask); err != nil {
			return Order{}, fmt.Errorf("failed to unmarshal ask order: %w", err)
		}
		return *NewOrder(0).WithAsk(&ask), nil
	case OrderTypeByteBid:
		var bid BidOrder
		if err := c.cdc.Unmarshal(data, &bid); err != nil {
			return Order{}, fmt.Errorf("failed to unmarshal bid order: %w", err)
		}
		return *NewOrder(0).WithBid(&bid), nil
	default:
		return Order{}, fmt.Errorf("unknown order type byte %#x", typeByte)
	}
}

// EncodeJSON converts the order to JSON.
func (c orderValueCodec) EncodeJSON(order Order) ([]byte, error) {
	return json.Marshal(order)
}

// DecodeJSON is not supported for orders: the order is a oneof that plain JSON can't restore.
func (c orderValueCodec) DecodeJSON(_ []byte) (Order, error) {
	return Order{}, errors.New("decoding an order from JSON is not supported")
}

// Stringify returns a string representation of the order.
func (c orderValueCodec) Stringify(order Order) string {
	return order.String()
}

// ValueType returns the name of the type this codec handles.
func (c orderValueCodec) ValueType() string {
	return "exchange.Order"
}

// byteValueCodec is the value codec for a single byte. See ByteValue.
type byteValueCodec struct{}

// Encode converts the byte into its store value.
func (byteValueCodec) Encode(value byte) ([]byte, error) {
	return []byte{value}, nil
}

// Decode converts a store value back into a byte.
func (byteValueCodec) Decode(bz []byte) (byte, error) {
	if len(bz) != 1 {
		return 0, fmt.Errorf("byte value must have length 1, got %d", len(bz))
	}
	return bz[0], nil
}

// EncodeJSON converts the byte to JSON.
func (byteValueCodec) EncodeJSON(value byte) ([]byte, error) {
	return json.Marshal(value)
}

// DecodeJSON converts JSON back into a byte.
func (byteValueCodec) DecodeJSON(bz []byte) (byte, error) {
	var value byte
	err := json.Unmarshal(bz, &value)
	return value, err
}

// Stringify returns a string representation of the byte.
func (byteValueCodec) Stringify(value byte) string {
	return fmt.Sprintf("%#x", value)
}

// ValueType returns the name of the type this codec handles.
func (byteValueCodec) ValueType() string {
	return "byte"
}

// coinsValueCodec is the value codec for coins. See CoinsValue.
type coinsValueCodec struct{}

// Encode converts the coins into their store value.
func (coinsValueCodec) Encode(value sdk.Coins) ([]byte, error) {
	return []byte(value.String()), nil
}

// Decode converts a store value back into coins.
func (coinsValueCodec) Decode(bz []byte) (sdk.Coins, error) {
	rv, err := sdk.ParseCoinsNormalized(string(bz))
	if err != nil {
		return nil, fmt.Errorf("invalid coins value: %w", err)
	}
	return rv, nil
}

// EncodeJSON converts the coins to JSON.
func (coinsValueCodec) EncodeJSON(value sdk.Coins) ([]byte, error) {
	return json.Marshal(value)
}

// DecodeJSON converts JSON back into coins.
func (coinsValueCodec) DecodeJSON(bz []byte) (sdk.Coins, error) {
	var value sdk.Coins
	err := json.Unmarshal(bz, &value)
	return value, err
}

// Stringify returns a string representation of the coins.
func (coinsValueCodec) Stringify(value sdk.Coins) string {
	return value.String()
}

// ValueType returns the name of the type this codec handles.
func (coinsValueCodec) ValueType() string {
	return "sdk.Coins"
}

// stringSliceValueCodec is the value codec for a list of strings. See StringSliceValue.
type stringSliceValueCodec struct{}

// Encode converts the strings into their store value.
func (stringSliceValueCodec) Encode(value []string) ([]byte, error) {
	for i, entry := range value {
		if strings.ContainsRune(entry, rune(RecordSeparator)) {
			return nil, fmt.Errorf("entry %d %q contains the record separator %#x", i, entry, RecordSeparator)
		}
	}
	return []byte(strings.Join(value, string(RecordSeparator))), nil
}

// Decode converts a store value back into the strings.
func (stringSliceValueCodec) Decode(bz []byte) ([]string, error) {
	if len(bz) == 0 {
		return nil, nil
	}
	return strings.Split(string(bz), string(RecordSeparator)), nil
}

// EncodeJSON converts the strings to JSON.
func (stringSliceValueCodec) EncodeJSON(value []string) ([]byte, error) {
	return json.Marshal(value)
}

// DecodeJSON converts JSON back into the strings.
func (stringSliceValueCodec) DecodeJSON(bz []byte) ([]string, error) {
	var value []string
	err := json.Unmarshal(bz, &value)
	return value, err
}

// Stringify returns a string representation of the strings.
func (stringSliceValueCodec) Stringify(value []string) string {
	return strings.Join(value, ",")
}

// ValueType returns the name of the type this codec handles.
func (stringSliceValueCodec) ValueType() string {
	return "[]string"
}
