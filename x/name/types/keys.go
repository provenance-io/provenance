package types

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
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

// Legacy prefixes - OLD values moved here for migration
var (
	LegacyNameKeyPrefix     = []byte{0x03}
	LegacyAddressKeyPrefix  = []byte{0x05}
	LegacyNameParamStoreKey = []byte{0x06}
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

var _ collcodec.KeyCodec[string] = HashedStringKeyCodec{}

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
	sum := sha256.Sum256([]byte(NormalizeName(name)))
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

// ValidateAddress validates an account address
func ValidateAddress(addr sdk.AccAddress) error {
	return sdk.VerifyAddressFormat(addr)
}
