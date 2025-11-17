package types

import (
	"errors"
	fmt "fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/codec"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

// RegisterInterfaces registers concrete implementations for this module.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	messages := make([]proto.Message, len(AllRequestMsgs))
	copy(messages, AllRequestMsgs)
	registry.RegisterImplementations((*sdk.Msg)(nil), messages...)

	registry.RegisterImplementations(
		(*TriggerEventI)(nil),
		&TransactionEvent{},
		&BlockHeightEvent{},
		&BlockTimeEvent{},
	)

	registry.RegisterInterface(
		"provenance.trigger.v1.TransactionEvent",
		(*TriggerEventI)(nil),
		&TransactionEvent{},
	)

	registry.RegisterInterface(
		"provenance.trigger.v1.BlockHeightEvent",
		(*TriggerEventI)(nil),
		&BlockHeightEvent{},
	)

	registry.RegisterInterface(
		"provenance.trigger.v1.BlockTimeEvent",
		(*TriggerEventI)(nil),
		&BlockTimeEvent{},
	)
}

// FixedBytes32KeyCodec handles 32-byte fixed size keys without length prefix
// This is necessary to maintain compatibility with the old KVStore key format for event listeners.
type FixedBytes32KeyCodec struct{}

// Encode writes the key bytes into the buffer.
func (f FixedBytes32KeyCodec) Encode(buffer []byte, key []byte) (int, error) {
	if len(key) != 32 {
		return 0, fmt.Errorf("key must be exactly 32 bytes, got %d", len(key))
	}
	if len(buffer) < 32 {
		return 0, errors.New("buffer too small for 32-byte key")
	}
	copy(buffer, key)
	return 32, nil
}

// Decode reads from the provided bytes buffer to decode the key T.
func (f FixedBytes32KeyCodec) Decode(buffer []byte) (int, []byte, error) {
	if len(buffer) < 32 {
		return 0, nil, errors.New("buffer too small for 32-byte key")
	}
	key := make([]byte, 32)
	copy(key, buffer[:32])
	return 32, key, nil
}

// Size returns the buffer size need to encode key T in binary format.
func (f FixedBytes32KeyCodec) Size(key []byte) int {
	return 32 // Fixed size
}

// Stringify returns a string representation of T.
func (f FixedBytes32KeyCodec) Stringify(key []byte) string {
	return fmt.Sprintf("%x", key)
}

// KeyType returns a string identifier for the type of the key.
func (f FixedBytes32KeyCodec) KeyType() string {
	return "bytes32"
}

// EncodeNonTerminal writes the key bytes into the buffer.
func (f FixedBytes32KeyCodec) EncodeNonTerminal(buffer []byte, key []byte) (int, error) {
	return f.Encode(buffer, key)
}

// DecodeNonTerminal reads the buffer provided and returns the key T.
func (f FixedBytes32KeyCodec) DecodeNonTerminal(buffer []byte) (int, []byte, error) {
	return f.Decode(buffer)
}

// SizeNonTerminal returns the maximum size of the key K when used in multipart keys like Pair.
func (f FixedBytes32KeyCodec) SizeNonTerminal(key []byte) int {
	return f.Size(key)
}

// EncodeJSON encodes the value as JSON.
func (f FixedBytes32KeyCodec) EncodeJSON(value []byte) ([]byte, error) {
	return nil, errors.New("EncodeJSON not implemented for FixedBytes32KeyCodec")
}

// DecodeJSON decodes the provided JSON bytes into an instance of T.
func (f FixedBytes32KeyCodec) DecodeJSON(b []byte) ([]byte, error) {
	return nil, errors.New("DecodeJSON not implemented for FixedBytes32KeyCodec")
}

// EventListenerKeyCodec creates a codec for event listener triple keys
// Matches EXACT existing key structure: [32-byte-hash][8-byte-order][8-byte-triggerID]
func EventListenerKeyCodec() codec.KeyCodec[collections.Triple[[]byte, uint64, uint64]] {
	return collections.TripleKeyCodec(
		FixedBytes32KeyCodec{}, // Fixed 32-byte encoding (no length prefix)
		collections.Uint64Key,  // order (8 bytes, big-endian)
		collections.Uint64Key,  // triggerID (8 bytes, big-endian)
	)
}
