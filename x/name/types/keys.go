package types

import (
	"slices"
	"strings"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	"cosmossdk.io/collections/indexes"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "name"

	// StoreKey is the store key string for distribution
	StoreKey = ModuleName

	// RouterKey is the message route for distribution
	RouterKey = ModuleName
)

var (
	// NameKeyPrefix is a prefix added to keys for adding/querying names.
	NameKeyPrefix = collections.NewPrefix(0x07)
	// AddressKeyPrefix is a prefix added to keys for indexing name records by address.
	AddressKeyPrefix = collections.NewPrefix(0x08)
	// NameParamStoreKey key for marker module's params
	NameParamStoreKey = collections.NewPrefix(0x09)
)

// NameRecordIndexes defines indexes for name records
type NameRecordIndexes struct {
	AddrIndex *indexes.Multi[sdk.AccAddress, string, NameRecord]
}

func (i NameRecordIndexes) IndexesList() []collections.Index[string, NameRecord] {
	return []collections.Index[string, NameRecord]{i.AddrIndex}
}

// ReversedNameKeyCodec keys a name record by its normalized name with the segments in reverse order.
// E.g. "ab.cd.ef" is stored using the key "ef.cd.ab".
type ReversedNameKeyCodec struct{}

var _ collcodec.KeyCodec[string] = ReversedNameKeyCodec{}

// Encode writes the reversed form of the provided name (see ReverseName) to the buffer.
func (ReversedNameKeyCodec) Encode(buffer []byte, key string) (int, error) {
	return collections.StringKey.Encode(buffer, ReverseName(key))
}

// Decode reads a reversed name from the buffer and returns the name with its segments back in their normal order.
func (ReversedNameKeyCodec) Decode(buffer []byte) (int, string, error) {
	n, reversed, err := collections.StringKey.Decode(buffer)
	if err != nil {
		return 0, "", err
	}
	return n, reverseSegments(reversed), nil
}

// Size returns the number of bytes needed to encode the provided name.
func (ReversedNameKeyCodec) Size(key string) int {
	return collections.StringKey.Size(ReverseName(key))
}

// EncodeNonTerminal writes the reversed form of the provided name to the buffer for use in a multipart key.
func (ReversedNameKeyCodec) EncodeNonTerminal(buffer []byte, key string) (int, error) {
	return collections.StringKey.EncodeNonTerminal(buffer, ReverseName(key))
}

// DecodeNonTerminal reads a reversed name from a multipart key and returns the name with its segments
// back in their normal order.
func (ReversedNameKeyCodec) DecodeNonTerminal(buffer []byte) (int, string, error) {
	n, reversed, err := collections.StringKey.DecodeNonTerminal(buffer)
	if err != nil {
		return 0, "", err
	}
	return n, reverseSegments(reversed), nil
}

// SizeNonTerminal returns the number of bytes needed to encode the provided name in a multipart key.
func (ReversedNameKeyCodec) SizeNonTerminal(key string) int {
	return collections.StringKey.SizeNonTerminal(ReverseName(key))
}

// EncodeJSON encodes the provided name as a JSON string (in its normal, non-reversed order).
func (ReversedNameKeyCodec) EncodeJSON(key string) ([]byte, error) {
	return collections.StringKey.EncodeJSON(key)
}

// DecodeJSON decodes a name from a JSON string.
func (ReversedNameKeyCodec) DecodeJSON(b []byte) (string, error) {
	return collections.StringKey.DecodeJSON(b)
}

// Stringify returns the provided name (in its normal, non-reversed order).
func (ReversedNameKeyCodec) Stringify(key string) string {
	return key
}

// KeyType returns the type of this key codec.
func (ReversedNameKeyCodec) KeyType() string {
	return "reversedname"
}

// ReverseName normalizes the provided name and reverses the order of its segments.
// E.g. "ab.cd.ef" becomes "ef.cd.ab". This is the form of a name used in its store key.
func ReverseName(name string) string {
	return reverseSegments(NormalizeName(name))
}

// reverseSegments reverses the order of the dot-separated segments in the provided string.
// It does not do any normalization.
func reverseSegments(name string) string {
	segments := strings.Split(name, ".")
	slices.Reverse(segments)
	return strings.Join(segments, ".")
}

// ValidateAddress validates an account address
func ValidateAddress(addr sdk.AccAddress) error {
	return sdk.VerifyAddressFormat(addr)
}
