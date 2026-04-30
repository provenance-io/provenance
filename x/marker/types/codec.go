package types

import (
	"encoding/json"
	"fmt"
	"strings"

	"cosmossdk.io/collections/codec"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkaddress "github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/gogoproto/proto"
)

// RegisterInterfaces registers concrete implementations for this module.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	messages := make([]proto.Message, len(AllRequestMsgs))
	copy(messages, AllRequestMsgs)
	registry.RegisterImplementations((*sdk.Msg)(nil), messages...)

	registry.RegisterImplementations(
		(*govtypesv1beta1.Content)(nil),
		&AddMarkerProposal{},
		&SupplyIncreaseProposal{},
		&SupplyDecreaseProposal{},
		&SetAdministratorProposal{},
		&RemoveAdministratorProposal{},
		&ChangeStatusProposal{},
		&WithdrawEscrowProposal{},
		&SetDenomMetadataProposal{},
	)

	registry.RegisterImplementations(
		(*authz.Authorization)(nil),
		&MarkerTransferAuthorization{},
		&MultiAuthorization{},
	)

	registry.RegisterInterface(
		"provenance.marker.v1.MarkerAccount",
		(*MarkerAccountI)(nil),
		&MarkerAccount{},
	)

	registry.RegisterInterface(
		"provenance.marker.v1.MarkerAccount",
		(*sdk.AccountI)(nil),
		&MarkerAccount{},
	)

	registry.RegisterInterface(
		"provenance.marker.v1.MarkerAccount",
		(*authtypes.GenesisAccount)(nil),
		&MarkerAccount{},
	)

	registry.RegisterInterface(
		"provenance.marker.v1.AccessGrant",
		(*AccessGrantI)(nil),
		&AccessGrant{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// MarkerAddrKeyCodec encodes sdk.AccAddress as [1-byte length][addr bytes].
var MarkerAddrKeyCodec codec.KeyCodec[sdk.AccAddress] = markerAddrKeyCodec{}

type markerAddrKeyCodec struct{}

func (markerAddrKeyCodec) Encode(buffer []byte, key sdk.AccAddress) (int, error) {
	lp := sdkaddress.MustLengthPrefix(key)
	copy(buffer, lp)
	return len(lp), nil
}

func (markerAddrKeyCodec) Decode(buffer []byte) (int, sdk.AccAddress, error) {
	if len(buffer) == 0 {
		return 0, nil, fmt.Errorf("empty buffer")
	}
	addrLen := int(buffer[0])
	if len(buffer) < 1+addrLen {
		return 0, nil, fmt.Errorf("buffer too short: need %d, have %d", 1+addrLen, len(buffer))
	}
	return 1 + addrLen, sdk.AccAddress(buffer[1 : 1+addrLen]), nil
}

func (markerAddrKeyCodec) Size(key sdk.AccAddress) int {
	return 1 + len(key)
}

func (markerAddrKeyCodec) EncodeJSON(key sdk.AccAddress) ([]byte, error) {
	return []byte(`"` + key.String() + `"`), nil
}

func (markerAddrKeyCodec) DecodeJSON(b []byte) (sdk.AccAddress, error) {
	s := strings.Trim(string(b), `"`)
	return sdk.AccAddressFromBech32(s)
}

func (markerAddrKeyCodec) Stringify(key sdk.AccAddress) string {
	return key.String()
}

func (markerAddrKeyCodec) KeyType() string {
	return "sdk.AccAddress"
}

func (markerAddrKeyCodec) EncodeNonTerminal(buffer []byte, key sdk.AccAddress) (int, error) {
	return markerAddrKeyCodec{}.Encode(buffer, key)
}

func (markerAddrKeyCodec) DecodeNonTerminal(buffer []byte) (int, sdk.AccAddress, error) {
	return markerAddrKeyCodec{}.Decode(buffer)
}

func (markerAddrKeyCodec) SizeNonTerminal(key sdk.AccAddress) int {
	return markerAddrKeyCodec{}.Size(key)
}

// SentinelValue encodes an empty []byte{} for KeySet-like maps.
var SentinelValue codec.ValueCodec[bool] = sentinelValueCodec{}

type sentinelValueCodec struct{}

func (sentinelValueCodec) Encode(value bool) ([]byte, error) {
	return []byte{}, nil
}

func (sentinelValueCodec) Decode(b []byte) (bool, error) {
	return true, nil
}

func (sentinelValueCodec) EncodeJSON(value bool) ([]byte, error) {
	return []byte("true"), nil
}

func (sentinelValueCodec) DecodeJSON(b []byte) (bool, error) {
	return true, nil
}

func (sentinelValueCodec) Stringify(value bool) string {
	return "true"
}

func (sentinelValueCodec) ValueType() string {
	return "bool(sentinel)"
}

// RawBytesValue stores the value as raw []byte without any encoding.
var RawBytesValue codec.ValueCodec[[]byte] = rawBytesValueCodec{}

type rawBytesValueCodec struct{}

func (rawBytesValueCodec) Encode(value []byte) ([]byte, error) {
	return value, nil
}

func (rawBytesValueCodec) Decode(b []byte) ([]byte, error) {
	dst := make([]byte, len(b))
	copy(dst, b)
	return dst, nil
}

func (rawBytesValueCodec) EncodeJSON(value []byte) ([]byte, error) {
	return json.Marshal(value)
}

func (rawBytesValueCodec) DecodeJSON(b []byte) ([]byte, error) {
	var v []byte
	err := json.Unmarshal(b, &v)
	return v, err
}

func (rawBytesValueCodec) Stringify(value []byte) string {
	return fmt.Sprintf("%x", value)
}

func (rawBytesValueCodec) ValueType() string {
	return "[]byte"
}

// DenomStringKeyCodec encodes a denom string as raw bytes.
var DenomStringKeyCodec codec.KeyCodec[string] = denomStringKeyCodec{}

type denomStringKeyCodec struct{}

func (denomStringKeyCodec) Encode(buffer []byte, key string) (int, error) {
	copy(buffer, key)
	return len(key), nil
}

func (denomStringKeyCodec) Decode(buffer []byte) (int, string, error) {
	return len(buffer), string(buffer), nil
}

func (denomStringKeyCodec) Size(key string) int {
	return len(key)
}

func (denomStringKeyCodec) EncodeJSON(key string) ([]byte, error) {
	return []byte(`"` + key + `"`), nil
}

func (denomStringKeyCodec) DecodeJSON(b []byte) (string, error) {
	return strings.Trim(string(b), `"`), nil
}

func (denomStringKeyCodec) Stringify(key string) string {
	return key
}

func (denomStringKeyCodec) KeyType() string {
	return "string(denom)"
}

func (denomStringKeyCodec) EncodeNonTerminal(_ []byte, _ string) (int, error) {
	return 0, fmt.Errorf("denom string key codec cannot be used in non-terminal position")
}

func (denomStringKeyCodec) DecodeNonTerminal(_ []byte) (int, string, error) {
	return 0, "", fmt.Errorf("denom string key codec cannot be used in non-terminal position")
}

func (denomStringKeyCodec) SizeNonTerminal(_ string) int {
	panic("denom string key codec cannot be used in non-terminal position")
}
