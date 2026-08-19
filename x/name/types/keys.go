package types

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"cosmossdk.io/collections"
	callcodec "cosmossdk.io/collections/codec"
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

// HashedStringKeyCodec keys a name record by the sha256 of its segments.
type HashedStringKeyCodec struct{}

var _ callcodec.KeyCodec[string] = HashedStringKeyCodec{}

func (c HashedStringKeyCodec) Encode(buffer []byte, key string) (int, error) {
	hash := c.ComputeHash(key)
	if len(buffer) < len(hash) {
		return 0, fmt.Errorf("%w: buffer of %d bytes is too small for a %d byte name key",
			ErrNameInvalid, len(buffer), len(hash))
	}
	copy(buffer, hash)
	return len(hash), nil
}

func (c HashedStringKeyCodec) EncodeNonTerminal(buffer []byte, key string) (int, error) {
	return c.Encode(buffer, key)
}

func (c HashedStringKeyCodec) EncodeJSON(key string) ([]byte, error) {
	return json.Marshal(key)
}

func (c HashedStringKeyCodec) Stringify(key string) string {
	return key
}

func (c HashedStringKeyCodec) KeyType() string {
	return "hashedstring"
}

func (c HashedStringKeyCodec) Size(_ string) int {
	return sha256.Size
}

func (c HashedStringKeyCodec) Decode(buffer []byte) (int, string, error) {
	return len(buffer), base64.StdEncoding.EncodeToString(buffer), nil
}

func (c HashedStringKeyCodec) DecodeNonTerminal(buffer []byte) (int, string, error) {
	if len(buffer) < sha256.Size {
		return 0, "", fmt.Errorf("%w: a hashed name key needs %d bytes, got %d",
			ErrNameInvalid, sha256.Size, len(buffer))
	}
	return sha256.Size, base64.StdEncoding.EncodeToString(buffer[:sha256.Size]), nil
}
func (c HashedStringKeyCodec) DecodeJSON(b []byte) (string, error) {
	var s string
	err := json.Unmarshal(b, &s)
	return s, err
}

func (c HashedStringKeyCodec) SizeNonTerminal(_ string) int {
	return sha256.Size
}

func (c HashedStringKeyCodec) ComputeHash(name string) []byte {
	sum := sha256.Sum256([]byte(trimNameSegments(name)))
	return sum[:]
}

func ComputeNameHash(name string) ([]byte, error) {
	for _, segment := range strings.Split(name, ".") {
		if len(strings.TrimSpace(segment)) == 0 {
			return nil, fmt.Errorf("name segment cannot be empty: %w", ErrNameInvalid)
		}
	}
	return HashedStringKeyCodec{}.ComputeHash(name), nil
}

// trimNameSegments removes whitespace around each segment so " a . b " and "a.b"
// produce the same key.
func trimNameSegments(name string) string {
	segments := strings.Split(name, ".")
	for i, segment := range segments {
		segments[i] = strings.TrimSpace(segment)
	}
	return strings.Join(segments, ".")
}

// ValidateAddress validates an account address
func ValidateAddress(addr sdk.AccAddress) error {
	return sdk.VerifyAddressFormat(addr)
}

// GetNameKeyBytes returns the name key in the same format as before (0x03 + sha256(name))
func GetNameKeyBytes(name string) ([]byte, error) {
	hash, err := ComputeNameHash(name)
	if err != nil {
		return nil, err
	}
	return append(append([]byte{}, NameKeyPrefix...), hash...), nil
}

// GetNameKey returns a name in storage form by trimming and lower-casing each segment.
func GetNameKey(name string) (string, error) {
	normalized := NormalizeName(name)
	for _, segment := range strings.Split(normalized, ".") {
		if len(segment) == 0 {
			return "", fmt.Errorf("name segment cannot be empty: %w", ErrNameInvalid)
		}
	}
	return normalized, nil
}

// LegacyComputeNameHash reproduces the old name-key hash by hashing segments in reverse order.
//
// Deprecated: used only by the 3->4 migration and its tests. Use ComputeNameHash instead.
func LegacyComputeNameHash(name string) ([]byte, error) {
	comps := strings.Split(name, ".")
	hsh := sha256.New()
	for i := len(comps) - 1; i >= 0; i-- {
		comp := strings.TrimSpace(comps[i])
		if len(comp) == 0 {
			return nil, fmt.Errorf("name segment cannot be empty: %w", ErrNameInvalid)
		}
		if _, err := hsh.Write([]byte(comp)); err != nil {
			return nil, err
		}
	}
	return hsh.Sum(nil), nil
}

// LegacyGetNameKeyBytes returns the full store key a name occupied before the 3->4 migration.
func LegacyGetNameKeyBytes(name string) ([]byte, error) {
	hash, err := LegacyComputeNameHash(name)
	if err != nil {
		return nil, err
	}
	return append(append([]byte{}, NameKeyPrefix...), hash...), nil
}
