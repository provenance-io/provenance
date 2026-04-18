package types

import (
	"encoding/binary"
	"encoding/json"
	fmt "fmt"

	collcodec "cosmossdk.io/collections/codec"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/gogoproto/proto"
)

// RegisterInterfaces registers concrete implementations for this module.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	messages := make([]proto.Message, len(AllRequestMsgs))
	copy(messages, AllRequestMsgs)
	registry.RegisterImplementations((*sdk.Msg)(nil), messages...)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// AttrTriple key (0x02),Format: [addr_len][addr][hash(name)][hash(value)].
type AttrTriple struct {
	AddrBytes []byte
	NameHash  [32]byte
	ValueHash [32]byte
}

var AttrTripleKey attrTripleKeyCodec
var _ collcodec.KeyCodec[AttrTriple] = attrTripleKeyCodec{}

type attrTripleKeyCodec struct{}

func (attrTripleKeyCodec) Encode(buffer []byte, key AttrTriple) (int, error) {
	if len(key.AddrBytes) > 255 {
		return 0, fmt.Errorf("attribute: address length %d exceeds 255", len(key.AddrBytes))
	}
	n := 0
	// Use address.MustLengthPrefix to match AddrAttributeKey byte-for-byte.
	lp := address.MustLengthPrefix(key.AddrBytes)
	copy(buffer[n:], lp)
	n += len(lp)
	copy(buffer[n:], key.NameHash[:])
	n += 32
	copy(buffer[n:], key.ValueHash[:])
	n += 32
	return n, nil
}

func (attrTripleKeyCodec) Decode(buffer []byte) (int, AttrTriple, error) {
	// Empty address: 64 bytes (two 32-byte hashes).
	if len(buffer) == 64 {
		var nameHash, valueHash [32]byte
		copy(nameHash[:], buffer[0:32])
		copy(valueHash[:], buffer[32:64])
		return 64, AttrTriple{NameHash: nameHash, ValueHash: valueHash}, nil
	}
	if len(buffer) < 1 {
		return 0, AttrTriple{}, fmt.Errorf("attribute: buffer too short for AttrTriple")
	}
	n := 0
	addrLen := int(buffer[n])
	n++
	if len(buffer) < n+addrLen+64 {
		return 0, AttrTriple{}, fmt.Errorf("attribute: buffer too short: need %d bytes, have %d", n+addrLen+64, len(buffer))
	}
	addrBytes := make([]byte, addrLen)
	copy(addrBytes, buffer[n:n+addrLen])
	n += addrLen
	var nameHash, valueHash [32]byte
	copy(nameHash[:], buffer[n:n+32])
	n += 32
	copy(valueHash[:], buffer[n:n+32])
	n += 32
	return n, AttrTriple{AddrBytes: addrBytes, NameHash: nameHash, ValueHash: valueHash}, nil
}

func (c attrTripleKeyCodec) EncodeNonTerminal(buffer []byte, key AttrTriple) (int, error) {
	return c.Encode(buffer, key)
}
func (c attrTripleKeyCodec) DecodeNonTerminal(buffer []byte) (int, AttrTriple, error) {
	return c.Decode(buffer)
}
func (c attrTripleKeyCodec) SizeNonTerminal(key AttrTriple) int { return c.Size(key) }
func (attrTripleKeyCodec) Size(key AttrTriple) int {
	if len(key.AddrBytes) == 0 {
		return 64 // no length prefix + two 32-byte hashes
	}
	return 1 + len(key.AddrBytes) + 32 + 32
}
func (attrTripleKeyCodec) EncodeJSON(key AttrTriple) ([]byte, error) {
	return json.Marshal(fmt.Sprintf("attr(%x,%x,%x)", key.AddrBytes, key.NameHash, key.ValueHash))
}
func (attrTripleKeyCodec) DecodeJSON(_ []byte) (AttrTriple, error) {
	return AttrTriple{}, fmt.Errorf("AttrTriple JSON decode not supported")
}
func (attrTripleKeyCodec) Stringify(key AttrTriple) string {
	return fmt.Sprintf("attr(%x)", key.AddrBytes)
}
func (attrTripleKeyCodec) KeyType() string { return "AttrTriple" }

// BuildAttrTriple constructs an AttrTriple from an Attribute.
func BuildAttrTriple(attr Attribute) AttrTriple {
	addrBz := attr.GetAddressBytes()
	var nameHash [32]byte
	copy(nameHash[:], GetNameKeyBytes(attr.Name))
	var valueHash [32]byte
	copy(valueHash[:], attr.Hash())
	return AttrTriple{AddrBytes: addrBz, NameHash: nameHash, ValueHash: valueHash}
}

// NameAddrPair key (0x03),Format: [hash(name)][addr_len][addr].
type NameAddrPair struct {
	NameHash  [32]byte
	AddrBytes []byte
}

var NameAddrPairKey nameAddrPairKeyCodec
var _ collcodec.KeyCodec[NameAddrPair] = nameAddrPairKeyCodec{}

type nameAddrPairKeyCodec struct{}

func (nameAddrPairKeyCodec) Encode(buffer []byte, key NameAddrPair) (int, error) {
	if len(key.AddrBytes) > 255 {
		return 0, fmt.Errorf("attribute: address length %d exceeds 255", len(key.AddrBytes))
	}
	n := 0
	copy(buffer[n:], key.NameHash[:])
	n += 32
	lp := address.MustLengthPrefix(key.AddrBytes)
	copy(buffer[n:], lp)
	n += len(lp)
	return n, nil
}

func (nameAddrPairKeyCodec) Decode(buffer []byte) (int, NameAddrPair, error) {
	if len(buffer) < 33 {
		return 0, NameAddrPair{}, fmt.Errorf("attribute: buffer too short for NameAddrPair")
	}
	n := 0
	var nameHash [32]byte
	copy(nameHash[:], buffer[n:n+32])
	n += 32
	addrLen := int(buffer[n])
	n++
	if len(buffer) < n+addrLen {
		return 0, NameAddrPair{}, fmt.Errorf("attribute: buffer too short for NameAddrPair address bytes")
	}
	addrBytes := make([]byte, addrLen)
	copy(addrBytes, buffer[n:n+addrLen])
	n += addrLen
	return n, NameAddrPair{NameHash: nameHash, AddrBytes: addrBytes}, nil
}

func (c nameAddrPairKeyCodec) EncodeNonTerminal(buffer []byte, key NameAddrPair) (int, error) {
	return c.Encode(buffer, key)
}
func (c nameAddrPairKeyCodec) DecodeNonTerminal(buffer []byte) (int, NameAddrPair, error) {
	return c.Decode(buffer)
}
func (c nameAddrPairKeyCodec) SizeNonTerminal(key NameAddrPair) int { return c.Size(key) }
func (nameAddrPairKeyCodec) Size(key NameAddrPair) int              { return 32 + 1 + len(key.AddrBytes) }
func (nameAddrPairKeyCodec) EncodeJSON(key NameAddrPair) ([]byte, error) {
	return json.Marshal(fmt.Sprintf("nap(%x,%x)", key.NameHash, key.AddrBytes))
}
func (nameAddrPairKeyCodec) DecodeJSON(_ []byte) (NameAddrPair, error) {
	return NameAddrPair{}, fmt.Errorf("NameAddrPair JSON decode not supported")
}
func (nameAddrPairKeyCodec) Stringify(key NameAddrPair) string {
	return fmt.Sprintf("nap(%x)", key.NameHash)
}
func (nameAddrPairKeyCodec) KeyType() string { return "NameAddrPair" }

// BuildNameAddrPair constructs a NameAddrPair from name and raw address bytes.
func BuildNameAddrPair(name string, addrBytes []byte) NameAddrPair {
	var nameHash [32]byte
	copy(nameHash[:], GetNameKeyBytes(name))
	return NameAddrPair{NameHash: nameHash, AddrBytes: addrBytes}
}

// ExpireTriple key (0x04),Format: [epoch(8)][addr_len][addr][hash(name)][hash(value)]
type ExpireTriple struct {
	EpochSecs int64
	AddrBytes []byte
	NameHash  [32]byte
	ValueHash [32]byte
}

var ExpireTripleKey expireTripleKeyCodec
var _ collcodec.KeyCodec[ExpireTriple] = expireTripleKeyCodec{}

type expireTripleKeyCodec struct{}

func (expireTripleKeyCodec) Encode(buffer []byte, key ExpireTriple) (int, error) {
	if len(key.AddrBytes) > 255 {
		return 0, fmt.Errorf("attribute: address length %d exceeds 255", len(key.AddrBytes))
	}
	n := 0
	binary.BigEndian.PutUint64(buffer[n:], uint64(key.EpochSecs)) //nolint:gosec // EpochSecs is non-negative and fits in uint64
	n += 8
	lp := address.MustLengthPrefix(key.AddrBytes)
	copy(buffer[n:], lp)
	n += len(lp)
	copy(buffer[n:], key.NameHash[:])
	n += 32
	copy(buffer[n:], key.ValueHash[:])
	n += 32
	return n, nil
}

func (expireTripleKeyCodec) Decode(buffer []byte) (int, ExpireTriple, error) {
	if len(buffer) < 9 {
		return 0, ExpireTriple{}, fmt.Errorf("attribute: buffer too short for ExpireTriple")
	}
	n := 0
	epochSecs := int64(binary.BigEndian.Uint64(buffer[n:])) //nolint:gosec // EpochSecs is non-negative and fits in uint64
	n += 8
	addrLen := int(buffer[n])
	n++
	if len(buffer) < n+addrLen+64 {
		return 0, ExpireTriple{}, fmt.Errorf("attribute: buffer too short for ExpireTriple body")
	}
	addrBytes := make([]byte, addrLen)
	copy(addrBytes, buffer[n:n+addrLen])
	n += addrLen
	var nameHash, valueHash [32]byte
	copy(nameHash[:], buffer[n:n+32])
	n += 32
	copy(valueHash[:], buffer[n:n+32])
	n += 32
	return n, ExpireTriple{EpochSecs: epochSecs, AddrBytes: addrBytes, NameHash: nameHash, ValueHash: valueHash}, nil
}

func (c expireTripleKeyCodec) EncodeNonTerminal(buffer []byte, key ExpireTriple) (int, error) {
	return c.Encode(buffer, key)
}
func (c expireTripleKeyCodec) DecodeNonTerminal(buffer []byte) (int, ExpireTriple, error) {
	return c.Decode(buffer)
}
func (c expireTripleKeyCodec) SizeNonTerminal(key ExpireTriple) int { return c.Size(key) }
func (expireTripleKeyCodec) Size(key ExpireTriple) int              { return 8 + 1 + len(key.AddrBytes) + 32 + 32 }
func (expireTripleKeyCodec) EncodeJSON(key ExpireTriple) ([]byte, error) {
	return json.Marshal(fmt.Sprintf("expire(%d,%x)", key.EpochSecs, key.AddrBytes))
}
func (expireTripleKeyCodec) DecodeJSON(_ []byte) (ExpireTriple, error) {
	return ExpireTriple{}, fmt.Errorf("ExpireTriple JSON decode not supported")
}
func (expireTripleKeyCodec) Stringify(key ExpireTriple) string {
	return fmt.Sprintf("expire(%d)", key.EpochSecs)
}
func (expireTripleKeyCodec) KeyType() string { return "ExpireTriple" }

// BuildExpireTriple constructs an ExpireTriple from an Attribute. Returns (key, false) when no expiration.
func BuildExpireTriple(attr Attribute) (ExpireTriple, bool) {
	if attr.ExpirationDate == nil {
		return ExpireTriple{}, false
	}
	var nameHash [32]byte
	copy(nameHash[:], GetNameKeyBytes(attr.Name))
	var valueHash [32]byte
	copy(valueHash[:], attr.Hash())
	return ExpireTriple{
		EpochSecs: attr.ExpirationDate.Unix(),
		AddrBytes: attr.GetAddressBytes(),
		NameHash:  nameHash,
		ValueHash: valueHash,
	}, true
}

// Uint64Value — encodes uint64 as big-endian 8 bytes. Identical to sdk.Uint64ToBigEndian.
var Uint64Value uint64ValueCodec
var _ collcodec.ValueCodec[uint64] = uint64ValueCodec{}

type uint64ValueCodec struct{}

func (uint64ValueCodec) Encode(value uint64) ([]byte, error) {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, value)
	return bz, nil
}
func (uint64ValueCodec) Decode(bz []byte) (uint64, error) {
	if len(bz) != 8 {
		return 0, fmt.Errorf("attribute: uint64 value must be 8 bytes, got %d", len(bz))
	}
	return binary.BigEndian.Uint64(bz), nil
}
func (uint64ValueCodec) EncodeJSON(value uint64) ([]byte, error) { return json.Marshal(value) }
func (uint64ValueCodec) DecodeJSON(bz []byte) (uint64, error) {
	var v uint64
	return v, json.Unmarshal(bz, &v)
}
func (uint64ValueCodec) Stringify(value uint64) string { return fmt.Sprintf("%d", value) }
func (uint64ValueCodec) ValueType() string             { return "uint64" }

// SentinelValue — encodes []byte{} on disk. Identical to store.Set(key, []byte{}).
var SentinelValue sentinelValueCodec
var _ collcodec.ValueCodec[bool] = sentinelValueCodec{}

type sentinelValueCodec struct{}

func (sentinelValueCodec) Encode(_ bool) ([]byte, error)     { return []byte{}, nil }
func (sentinelValueCodec) Decode(_ []byte) (bool, error)     { return true, nil }
func (sentinelValueCodec) EncodeJSON(_ bool) ([]byte, error) { return json.Marshal(true) }
func (sentinelValueCodec) DecodeJSON(bz []byte) (bool, error) {
	var v bool
	return v, json.Unmarshal(bz, &v)
}
func (sentinelValueCodec) Stringify(_ bool) string { return "sentinel" }
func (sentinelValueCodec) ValueType() string       { return "sentinel" }
