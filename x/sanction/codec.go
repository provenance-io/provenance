package sanction

import (
	"encoding/binary"
	"encoding/json"
	fmt "fmt"

	"cosmossdk.io/collections"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/gogoproto/proto"
)

// RegisterInterfaces registers concrete implementations for this module.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	messages := make([]proto.Message, len(AllRequestMsgs))
	copy(messages, AllRequestMsgs)
	registry.RegisterImplementations((*sdk.Msg)(nil), messages...)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// AccAddressKey implements KeyCodec[sdk.AccAddress] for length-prefixed addresses
type AccAddressKey struct{}

func (a AccAddressKey) Encode(buffer []byte, addr sdk.AccAddress) (int, error) {
	size := a.Size(addr)
	if len(buffer) < size {
		return 0, fmt.Errorf("buffer too small: need %d, got %d", size, len(buffer))
	}

	buffer[0] = byte(len(addr))
	copy(buffer[1:], addr)
	return size, nil
}

func (a AccAddressKey) Decode(buffer []byte) (int, sdk.AccAddress, error) {
	if len(buffer) == 0 {
		return 0, nil, fmt.Errorf("empty buffer")
	}

	addrLen := int(buffer[0])
	size := 1 + addrLen

	if len(buffer) < size {
		return 0, nil, fmt.Errorf("insufficient bytes: need %d, got %d", size, len(buffer))
	}

	addr := sdk.AccAddress(buffer[1 : 1+addrLen])
	return size, addr, nil
}

func (a AccAddressKey) Size(addr sdk.AccAddress) int {
	return 1 + len(addr)
}

func (a AccAddressKey) EncodeJSON(addr sdk.AccAddress) ([]byte, error) {
	return json.Marshal(addr.String())
}

func (a AccAddressKey) DecodeJSON(b []byte) (sdk.AccAddress, error) {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return sdk.AccAddressFromBech32(s)
}

func (a AccAddressKey) Stringify(addr sdk.AccAddress) string {
	return addr.String()
}

func (a AccAddressKey) KeyType() string {
	return "sdk.AccAddress"
}

func (a AccAddressKey) EncodeNonTerminal(buffer []byte, addr sdk.AccAddress) (int, error) {
	return a.Encode(buffer, addr)
}

func (a AccAddressKey) DecodeNonTerminal(buffer []byte) (int, sdk.AccAddress, error) {
	return a.Decode(buffer)
}

func (a AccAddressKey) SizeNonTerminal(addr sdk.AccAddress) int {
	return a.Size(addr)
}

// TemporaryKeyCodec implements KeyCodec[collections.Pair[sdk.AccAddress, uint64]]
// Format: <addr len (1 byte)><addr><gov prop id (8 bytes)>
type TemporaryKeyCodec struct{}

func (t TemporaryKeyCodec) Encode(buffer []byte, key collections.Pair[sdk.AccAddress, uint64]) (int, error) {
	addr := key.K1()
	propID := key.K2()

	size := t.Size(key)
	if len(buffer) < size {
		return 0, fmt.Errorf("buffer too small: need %d, got %d", size, len(buffer))
	}

	buffer[0] = byte(len(addr))
	copy(buffer[1:], addr)
	binary.BigEndian.PutUint64(buffer[1+len(addr):], propID)

	return size, nil
}

func (t TemporaryKeyCodec) Decode(buffer []byte) (int, collections.Pair[sdk.AccAddress, uint64], error) {
	if len(buffer) < 1+8 {
		return 0, collections.Pair[sdk.AccAddress, uint64]{}, fmt.Errorf("buffer too short: got %d bytes", len(buffer))
	}

	addrLen := int(buffer[0])
	size := 1 + addrLen + 8

	if len(buffer) < size {
		return 0, collections.Pair[sdk.AccAddress, uint64]{}, fmt.Errorf("invalid address length: need %d, got %d", size, len(buffer))
	}

	addr := sdk.AccAddress(buffer[1 : 1+addrLen])
	propID := binary.BigEndian.Uint64(buffer[1+addrLen:])

	return size, collections.Join(addr, propID), nil
}

func (t TemporaryKeyCodec) Size(key collections.Pair[sdk.AccAddress, uint64]) int {
	return 1 + len(key.K1()) + 8
}

func (t TemporaryKeyCodec) EncodeJSON(key collections.Pair[sdk.AccAddress, uint64]) ([]byte, error) {
	type jsonKey struct {
		Address string `json:"address"`
		PropID  uint64 `json:"prop_id"`
	}
	return json.Marshal(jsonKey{
		Address: key.K1().String(),
		PropID:  key.K2(),
	})
}

func (t TemporaryKeyCodec) DecodeJSON(b []byte) (collections.Pair[sdk.AccAddress, uint64], error) {
	var data struct {
		Address string `json:"address"`
		PropID  uint64 `json:"prop_id"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return collections.Pair[sdk.AccAddress, uint64]{}, err
	}

	addr, err := sdk.AccAddressFromBech32(data.Address)
	if err != nil {
		return collections.Pair[sdk.AccAddress, uint64]{}, err
	}

	return collections.Join(addr, data.PropID), nil
}

func (t TemporaryKeyCodec) Stringify(key collections.Pair[sdk.AccAddress, uint64]) string {
	return fmt.Sprintf("%s/%d", key.K1().String(), key.K2())
}

func (t TemporaryKeyCodec) KeyType() string {
	return "sanction.TemporaryKey"
}

func (t TemporaryKeyCodec) EncodeNonTerminal(buffer []byte, key collections.Pair[sdk.AccAddress, uint64]) (int, error) {
	return t.Encode(buffer, key)
}

func (t TemporaryKeyCodec) DecodeNonTerminal(buffer []byte) (int, collections.Pair[sdk.AccAddress, uint64], error) {
	return t.Decode(buffer)
}

func (t TemporaryKeyCodec) SizeNonTerminal(key collections.Pair[sdk.AccAddress, uint64]) int {
	return t.Size(key)
}

func (t TemporaryKeyCodec) EncodePrefix(addr sdk.AccAddress) []byte {
	prefix := make([]byte, 1+len(addr))
	prefix[0] = byte(len(addr))
	copy(prefix[1:], addr)
	return prefix
}

// ProposalIndexKeyCodec implements KeyCodec[collections.Pair[uint64, sdk.AccAddress]]
// Format: <proposal id (8 bytes)><addr len (1 byte)><addr>
type ProposalIndexKeyCodec struct{}

func (p ProposalIndexKeyCodec) Encode(buffer []byte, key collections.Pair[uint64, sdk.AccAddress]) (int, error) {
	propID := key.K1()
	addr := key.K2()

	size := p.Size(key)
	if len(buffer) < size {
		return 0, fmt.Errorf("buffer too small: need %d, got %d", size, len(buffer))
	}

	binary.BigEndian.PutUint64(buffer[0:8], propID)
	buffer[8] = byte(len(addr))
	copy(buffer[9:], addr)

	return size, nil
}

func (p ProposalIndexKeyCodec) Decode(buffer []byte) (int, collections.Pair[uint64, sdk.AccAddress], error) {
	if len(buffer) < 8+1 {
		return 0, collections.Pair[uint64, sdk.AccAddress]{}, fmt.Errorf("buffer too short: got %d bytes", len(buffer))
	}

	propID := binary.BigEndian.Uint64(buffer[0:8])
	addrLen := int(buffer[8])
	size := 8 + 1 + addrLen

	if len(buffer) < size {
		return 0, collections.Pair[uint64, sdk.AccAddress]{}, fmt.Errorf("invalid address length: need %d, got %d", size, len(buffer))
	}

	addr := sdk.AccAddress(buffer[9 : 9+addrLen])

	return size, collections.Join(propID, addr), nil
}

func (p ProposalIndexKeyCodec) Size(key collections.Pair[uint64, sdk.AccAddress]) int {
	return 8 + 1 + len(key.K2())
}

func (p ProposalIndexKeyCodec) EncodeJSON(key collections.Pair[uint64, sdk.AccAddress]) ([]byte, error) {
	type jsonKey struct {
		PropID  uint64 `json:"prop_id"`
		Address string `json:"address"`
	}
	return json.Marshal(jsonKey{
		PropID:  key.K1(),
		Address: key.K2().String(),
	})
}

func (p ProposalIndexKeyCodec) DecodeJSON(b []byte) (collections.Pair[uint64, sdk.AccAddress], error) {
	var data struct {
		PropID  uint64 `json:"prop_id"`
		Address string `json:"address"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return collections.Pair[uint64, sdk.AccAddress]{}, err
	}

	addr, err := sdk.AccAddressFromBech32(data.Address)
	if err != nil {
		return collections.Pair[uint64, sdk.AccAddress]{}, err
	}

	return collections.Join(data.PropID, addr), nil
}

func (p ProposalIndexKeyCodec) Stringify(key collections.Pair[uint64, sdk.AccAddress]) string {
	return fmt.Sprintf("%d/%s", key.K1(), key.K2().String())
}

func (p ProposalIndexKeyCodec) KeyType() string {
	return "sanction.ProposalIndexKey"
}

func (p ProposalIndexKeyCodec) EncodeNonTerminal(buffer []byte, key collections.Pair[uint64, sdk.AccAddress]) (int, error) {
	return p.Encode(buffer, key)
}

func (p ProposalIndexKeyCodec) DecodeNonTerminal(buffer []byte) (int, collections.Pair[uint64, sdk.AccAddress], error) {
	return p.Decode(buffer)
}

func (p ProposalIndexKeyCodec) SizeNonTerminal(key collections.Pair[uint64, sdk.AccAddress]) int {
	return p.Size(key)
}
