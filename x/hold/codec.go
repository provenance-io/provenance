package hold

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/collections"
	sdkmath "cosmossdk.io/math"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	messages := make([]proto.Message, len(AllRequestMsgs))
	copy(messages, AllRequestMsgs)
	registry.RegisterImplementations((*sdk.Msg)(nil), messages...)
}

// Custom value codec for sdkmath.Int
type IntValueCodec struct{}

func (c IntValueCodec) Encode(value sdkmath.Int) ([]byte, error) {
	return value.Marshal()
}

func (c IntValueCodec) Decode(b []byte) (sdkmath.Int, error) {
	if len(b) == 0 {
		return sdkmath.ZeroInt(), nil
	}
	var i sdkmath.Int
	err := i.Unmarshal(b)
	return i, err
}

func (c IntValueCodec) EncodeJSON(value sdkmath.Int) ([]byte, error) {
	return value.MarshalJSON()
}

func (c IntValueCodec) DecodeJSON(b []byte) (sdkmath.Int, error) {
	var i sdkmath.Int
	err := i.UnmarshalJSON(b)
	return i, err
}

func (c IntValueCodec) Stringify(value sdkmath.Int) string {
	return value.String()
}

func (c IntValueCodec) ValueType() string {
	return "sdkmath.Int"
}

// Custom key codec for address with 1-byte length prefix
type AddressKeyCodec struct{}

func (AddressKeyCodec) Encode(buffer []byte, addr sdk.AccAddress) (int, error) {
	if addr == nil {
		return 0, fmt.Errorf("cannot encode nil address")
	}
	if len(buffer) == 0 {
		return 0, fmt.Errorf("buffer is empty")
	}
	if len(addr) > 255 {
		return 0, collections.ErrEncoding
	}
	buffer[0] = byte(len(addr))
	copy(buffer[1:], addr)
	return 1 + len(addr), nil
}

func (AddressKeyCodec) Decode(buffer []byte) (int, sdk.AccAddress, error) {
	if len(buffer) == 0 {
		return 0, nil, collections.ErrEncoding
	}
	length := int(buffer[0])
	if len(buffer) < 1+length {
		return 0, nil, collections.ErrEncoding
	}
	addr := make(sdk.AccAddress, length)
	copy(addr, buffer[1:1+length])
	return 1 + length, addr, nil
}

func (AddressKeyCodec) Size(key sdk.AccAddress) int {
	return 1 + len(key)
}

func (AddressKeyCodec) EncodeJSON(value sdk.AccAddress) ([]byte, error) {
	return json.Marshal(value.String())
}

func (AddressKeyCodec) DecodeJSON(b []byte) (sdk.AccAddress, error) {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return sdk.AccAddressFromBech32(s)
}

func (AddressKeyCodec) Stringify(key sdk.AccAddress) string {
	return key.String()
}

func (AddressKeyCodec) KeyType() string {
	return "addressKeyCodec"
}

func (AddressKeyCodec) EncodeNonTerminal(buffer []byte, key sdk.AccAddress) (int, error) {
	return AddressKeyCodec{}.Encode(buffer, key)
}

func (AddressKeyCodec) DecodeNonTerminal(buffer []byte) (int, sdk.AccAddress, error) {
	return AddressKeyCodec{}.Decode(buffer)
}

func (AddressKeyCodec) SizeNonTerminal(key sdk.AccAddress) int {
	return AddressKeyCodec{}.Size(key)
}

// Custom key codec for denom (raw bytes, no length prefix)
type DenomKeyCodec struct{}

func (DenomKeyCodec) Encode(buffer []byte, denom string) (int, error) {
	denomBytes := []byte(denom)
	if len(buffer) < len(denomBytes) {
		return 0, collections.ErrEncoding
	}
	return copy(buffer, denomBytes), nil
}

func (DenomKeyCodec) Decode(buffer []byte) (int, string, error) {
	return len(buffer), string(buffer), nil
}

func (DenomKeyCodec) Size(key string) int {
	return len(key)
}

func (DenomKeyCodec) EncodeJSON(value string) ([]byte, error) {
	return json.Marshal(value)
}

func (DenomKeyCodec) DecodeJSON(b []byte) (string, error) {
	var s string
	err := json.Unmarshal(b, &s)
	return s, err
}

func (DenomKeyCodec) Stringify(key string) string {
	return key
}

func (DenomKeyCodec) KeyType() string {
	return "denomKeyCodec"
}

func (DenomKeyCodec) EncodeNonTerminal(buffer []byte, key string) (int, error) {
	return DenomKeyCodec{}.Encode(buffer, key)
}

func (DenomKeyCodec) DecodeNonTerminal(buffer []byte) (int, string, error) {
	return DenomKeyCodec{}.Decode(buffer)
}

func (DenomKeyCodec) SizeNonTerminal(key string) int {
	return DenomKeyCodec{}.Size(key)
}
