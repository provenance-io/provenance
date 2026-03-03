package quarantine

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	fmt "fmt"
	"sort"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// ModuleName is the name of the module
	ModuleName = "quarantine"

	// StoreKey is the store key string for gov
	StoreKey = ModuleName
)

var (
	// OptInPrefix is the prefix for the quarantine account opt-in flags.
	OptInPrefix = []byte{0x00}

	// AutoResponsePrefix is the prefix for quarantine auto-response settings.
	AutoResponsePrefix = []byte{0x01}

	// RecordPrefix is the prefix for keys with the records of quarantined funds.
	RecordPrefix = []byte{0x02}

	// RecordIndexPrefix is the prefix for the index of record suffixes.
	RecordIndexPrefix = []byte{0x03}
)

var (
	// OptInKeyPrefix is the collections.Prefix for the opt-in map.
	OptInKeyPrefix = collections.NewPrefix(OptInPrefix)
	// AutoResponseKeyPrefix is the collections.Prefix for the auto-response map.
	AutoResponseKeyPrefix = collections.NewPrefix(AutoResponsePrefix)
	// RecordKeyPrefix is the collections.Prefix for the quarantine record map.
	RecordKeyPrefix = collections.NewPrefix(RecordPrefix)
	// RecordIndexKeyPrefix is the collections.Prefix for the record-suffix index map.
	RecordIndexKeyPrefix = collections.NewPrefix(RecordIndexPrefix)
)

// MakeKey concatenates the two byte slices into a new byte slice.
func MakeKey(parts ...[]byte) []byte {
	var size int
	for _, part := range parts {
		size += len(part)
	}
	rv := make([]byte, 0, size)
	for _, part := range parts {
		rv = append(rv, part...)
	}
	return rv
}

// CreateOptInKey creates the key for a quarantine opt-in record.
func CreateOptInKey(toAddr sdk.AccAddress) []byte {
	return MakeKey(OptInPrefix, address.MustLengthPrefix(toAddr))
}

// ParseOptInKey extracts the account address from the provided quarantine opt-in key.
func ParseOptInKey(key []byte) (toAddr sdk.AccAddress) {
	// key is of format:
	// 0x00<to addr len><to addr bytes>
	toAddrLen, toAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, 1, 1)
	toAddr, _ = sdk.ParseLengthPrefixedBytes(key, toAddrLenEndIndex+1, int(toAddrLen[0]))

	return toAddr
}

// CreateAutoResponseToAddrPrefix creates a prefix for the quarantine auto-responses for a receiving address.
func CreateAutoResponseToAddrPrefix(toAddr sdk.AccAddress) []byte {
	return MakeKey(AutoResponsePrefix, address.MustLengthPrefix(toAddr))
}

// CreateAutoResponseKey creates the key for a quarantine auto-response.
func CreateAutoResponseKey(toAddr, fromAddr sdk.AccAddress) []byte {
	return MakeKey(AutoResponsePrefix,
		address.MustLengthPrefix(toAddr),
		address.MustLengthPrefix(fromAddr))
}

// ParseAutoResponseKey extracts the to address and from address from the provided quarantine auto-response key.
func ParseAutoResponseKey(key []byte) (toAddr, fromAddr sdk.AccAddress) {
	// key is of format:
	// 0x01<to addr len><to addr bytes><from addr len><from addr bytes>
	var toAddrEndIndex int
	toAddrLen, toAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, 1, 1)
	toAddr, toAddrEndIndex = sdk.ParseLengthPrefixedBytes(key, toAddrLenEndIndex+1, int(toAddrLen[0]))

	fromAddrLen, fromAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, toAddrEndIndex+1, 1)
	fromAddr, _ = sdk.ParseLengthPrefixedBytes(key, fromAddrLenEndIndex+1, int(fromAddrLen[0]))

	return toAddr, fromAddr
}

// CreateRecordToAddrPrefix creates a prefix for the quarantine funds for a receiving address.
func CreateRecordToAddrPrefix(toAddr sdk.AccAddress) []byte {
	return MakeKey(RecordPrefix, address.MustLengthPrefix(toAddr))
}

// CreateRecordKey creates the key for a quarantine record.
//
// If there is only one fromAddr, it is used as the record suffix.
// If there are more than one, a hash of them is used as the suffix.
//
// Panics if no fromAddrs are provided.
func CreateRecordKey(toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) []byte {
	// This is designed such that a known record suffix can be provided
	// as a single "from address" to create the key with that suffix.
	toAddrPreBz := CreateRecordToAddrPrefix(toAddr)
	recordID := address.MustLengthPrefix(CreateRecordSuffix(fromAddrs))
	return MakeKey(toAddrPreBz, recordID)
}

// createRecordSuffix creates a single "address" to use for the provided from addresses.
// This is not to be confused with CreateRecordKey which creates the full key for a quarantine record.
// This only creates a portion of the key.
//
// If one fromAddr is provided, it's what's returned.
// If more than one is provided, they are sorted, combined, and hashed.
//
// Panics if none are provided.
func CreateRecordSuffix(fromAddrs []sdk.AccAddress) []byte {
	// This is designed such that a known record suffix can be provided
	// as a single "from address" to create the key with that suffix.
	switch len(fromAddrs) {
	case 0:
		panic(sdkerrors.ErrLogic.Wrap("at least one fromAddr is required"))
	case 1:
		if len(fromAddrs[0]) > 32 {
			return fromAddrs[0][:32]
		}
		return fromAddrs[0]
	default:
		// The same n addresses needs to always create the same result.
		// And we don't want to change the input slice.
		addrs := make([]sdk.AccAddress, len(fromAddrs))
		copy(addrs, fromAddrs)
		sort.Slice(addrs, func(i, j int) bool {
			return bytes.Compare(addrs[i], addrs[j]) < 0
		})
		var toHash []byte
		for _, addr := range addrs {
			toHash = append(toHash, addr...)
		}
		hash := sha256.Sum256(toHash)
		return hash[0:]
	}
}

// ParseRecordKey extracts the to address and record suffix from the provided quarantine funds key.
func ParseRecordKey(key []byte) (toAddr, recordSuffix sdk.AccAddress) {
	// key is of format:
	// 0x02<to addr len><to addr bytes><record suffix len><record suffix bytes>
	var toAddrEndIndex int
	toAddrLen, toAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, 1, 1)
	toAddr, toAddrEndIndex = sdk.ParseLengthPrefixedBytes(key, toAddrLenEndIndex+1, int(toAddrLen[0]))

	recordSuffixLen, recordSuffixLenEndIndex := sdk.ParseLengthPrefixedBytes(key, toAddrEndIndex+1, 1)
	recordSuffix, _ = sdk.ParseLengthPrefixedBytes(key, recordSuffixLenEndIndex+1, int(recordSuffixLen[0]))

	return toAddr, recordSuffix
}

// CreateRecordIndexToAddrPrefix creates a prefix for the quarantine record index entries for a receiving address.
func CreateRecordIndexToAddrPrefix(toAddr sdk.AccAddress) []byte {
	return MakeKey(RecordIndexPrefix, address.MustLengthPrefix(toAddr))
}

// CreateRecordIndexKey creates the key for the quarantine record suffix index.
func CreateRecordIndexKey(toAddr, fromAddr sdk.AccAddress) []byte {
	return MakeKey(RecordIndexPrefix,
		address.MustLengthPrefix(toAddr),
		address.MustLengthPrefix(fromAddr))
}

// ParseRecordIndexKey extracts the to address and from address from the provided quarantine record index key.
func ParseRecordIndexKey(key []byte) (toAddr, fromAddr sdk.AccAddress) {
	// key is of format:
	// 0x03<to addr len><to addr bytes><from addr len><from addr bytes>
	var toAddrEndIndex int
	toAddrLen, toAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, 1, 1)
	toAddr, toAddrEndIndex = sdk.ParseLengthPrefixedBytes(key, toAddrLenEndIndex+1, int(toAddrLen[0]))

	fromAddrLen, fromAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, toAddrEndIndex+1, 1)
	fromAddr, _ = sdk.ParseLengthPrefixedBytes(key, fromAddrLenEndIndex+1, int(fromAddrLen[0]))

	return toAddr, fromAddr
}

type addressLengthPrefixedKeyCodec struct{}

// AddressLengthPrefixedKey is the singleton key codec for sdk.AccAddress that
// replicates the length-prefix encoding used throughout the quarantine module.
var AddressLengthPrefixedKey addressLengthPrefixedKeyCodec

// Compile-time check: addressLengthPrefixedKeyCodec must satisfy KeyCodec.
var _ collcodec.KeyCodec[sdk.AccAddress] = addressLengthPrefixedKeyCodec{}

func (addressLengthPrefixedKeyCodec) Encode(buffer []byte, key sdk.AccAddress) (int, error) {
	if len(key) > 255 {
		return 0, fmt.Errorf("quarantine: address length %d exceeds maximum of 255", len(key))
	}
	buffer[0] = byte(len(key))
	copy(buffer[1:], key)
	return 1 + len(key), nil
}

func (addressLengthPrefixedKeyCodec) Decode(buffer []byte) (int, sdk.AccAddress, error) {
	if len(buffer) < 1 {
		return 0, nil, fmt.Errorf("quarantine: buffer too short to read address length byte")
	}
	addrLen := int(buffer[0])
	if len(buffer) < 1+addrLen {
		return 0, nil, fmt.Errorf("quarantine: buffer too short: need %d address bytes, have %d", addrLen, len(buffer)-1)
	}
	addr := make(sdk.AccAddress, addrLen)
	copy(addr, buffer[1:1+addrLen])
	return 1 + addrLen, addr, nil
}

func (addressLengthPrefixedKeyCodec) Size(key sdk.AccAddress) int {
	return 1 + len(key)
}

func (c addressLengthPrefixedKeyCodec) EncodeNonTerminal(buffer []byte, key sdk.AccAddress) (int, error) {
	return c.Encode(buffer, key)
}

func (c addressLengthPrefixedKeyCodec) DecodeNonTerminal(buffer []byte) (int, sdk.AccAddress, error) {
	return c.Decode(buffer)
}

func (c addressLengthPrefixedKeyCodec) SizeNonTerminal(key sdk.AccAddress) int {
	return c.Size(key)
}

func (addressLengthPrefixedKeyCodec) EncodeJSON(key sdk.AccAddress) ([]byte, error) {
	return json.Marshal(key.String())
}

func (addressLengthPrefixedKeyCodec) DecodeJSON(bz []byte) (sdk.AccAddress, error) {
	var addrStr string
	if err := json.Unmarshal(bz, &addrStr); err != nil {
		return nil, err
	}
	return sdk.AccAddressFromBech32(addrStr)
}

func (addressLengthPrefixedKeyCodec) Stringify(key sdk.AccAddress) string {
	return key.String()
}

func (addressLengthPrefixedKeyCodec) KeyType() string {
	return "address(len-prefixed)"
}

// Encodes AutoResponse as one byte using the existing on-chain format.
// - UNSPECIFIED → 0x00
// - ACCEPT → 0x01
// - DECLINE → 0x02
type autoResponseValueCodec struct{}

// AutoResponseValue is the singleton value codec for AutoResponse entries in
// the quarantine auto-response collection.
var AutoResponseValue autoResponseValueCodec

// AutoResponseValueCodec must satisfy ValueCodec.
var _ collcodec.ValueCodec[AutoResponse] = autoResponseValueCodec{}

func (autoResponseValueCodec) Encode(value AutoResponse) ([]byte, error) {
	return []byte{ToAutoB(value)}, nil
}
func (autoResponseValueCodec) Decode(bz []byte) (AutoResponse, error) {
	return ToAutoResponse(bz), nil
}
func (autoResponseValueCodec) EncodeJSON(value AutoResponse) ([]byte, error) {
	return json.Marshal(int32(value))
}
func (autoResponseValueCodec) DecodeJSON(bz []byte) (AutoResponse, error) {
	var n int32
	if err := json.Unmarshal(bz, &n); err != nil {
		return 0, fmt.Errorf("quarantine: failed to decode AutoResponse from JSON: %w", err)
	}
	return AutoResponse(n), nil
}
func (autoResponseValueCodec) Stringify(value AutoResponse) string {
	return value.String()
}
func (autoResponseValueCodec) ValueType() string {
	return "AutoResponse"
}

// optInValueCodec encodes the opt-in presence flag as a single 0x00 byte,
// exactly matching the value written by the pre-collections quarantine keeper.
type optInValueCodec struct{}

// OptInValue is the singleton value codec for the opt-in collection.
var OptInValue optInValueCodec

// OptInValueCodec must satisfy ValueCodec.
var _ collcodec.ValueCodec[bool] = optInValueCodec{}

func (optInValueCodec) Encode(_ bool) ([]byte, error) {
	return []byte{0x00}, nil
}
func (optInValueCodec) Decode(_ []byte) (bool, error) {
	return true, nil
}
func (optInValueCodec) EncodeJSON(_ bool) ([]byte, error) {
	return json.Marshal(true)
}
func (optInValueCodec) DecodeJSON(bz []byte) (bool, error) {
	var v bool
	return v, json.Unmarshal(bz, &v)
}
func (optInValueCodec) Stringify(_ bool) string {
	return "opted-in"
}
func (optInValueCodec) ValueType() string {
	return "opt-in-flag"
}
