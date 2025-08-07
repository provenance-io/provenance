package types

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
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
	NameKeyPrefix = []byte{0x03}
	// AddressKeyPrefix is a prefix added to keys for indexing name records by address.
	AddressKeyPrefix = []byte{0x05}
	// NameParamStoreKey key for marker module's params
	NameParamStoreKey = []byte{0x06}
)

// NameRecordIndexes defines indexes for name records
type NameRecordIndexes struct {
	// AddrIndex maps (address, name) -> empty, like bank module's pattern
	AddrIndex *indexes.Multi[collections.Pair[[]byte, string], string, NameRecord]
}

// IndexesList implements collections.Indexes
func (i NameRecordIndexes) IndexesList() []collections.Index[string, NameRecord] {
	return []collections.Index[string, NameRecord]{i.AddrIndex}
}

type HashedStringKeyCodec struct{}

func (c HashedStringKeyCodec) Encode(buffer []byte, key string) (int, error) {
	hash := c.ComputeHash(key)
	if len(buffer) < len(hash) {
		return 0, fmt.Errorf("buffer too small")
	}
	copy(buffer, hash)
	return len(hash), nil
}

func (c HashedStringKeyCodec) Decode(buffer []byte) (int, string, error) {
	return len(buffer), base64.StdEncoding.EncodeToString(buffer), nil
}

func (c HashedStringKeyCodec) Size(key string) int {
	return sha256.Size
}

func (c HashedStringKeyCodec) EncodeJSON(key string) ([]byte, error) {
	return json.Marshal(key)
}

func (c HashedStringKeyCodec) DecodeJSON(b []byte) (string, error) {
	var s string
	err := json.Unmarshal(b, &s)
	return s, err
}

func (c HashedStringKeyCodec) Stringify(key string) string {
	return key
}

func (c HashedStringKeyCodec) KeyType() string {
	return "hashedstring"
}

func (c HashedStringKeyCodec) EncodeNonTerminal(buffer []byte, key string) (int, error) {
	hash := c.ComputeHash(key)
	if len(buffer) < len(hash) {
		return 0, fmt.Errorf("buffer too small")
	}
	copy(buffer, hash)
	return len(hash), nil
}

func (c HashedStringKeyCodec) DecodeNonTerminal(buffer []byte) (int, string, error) {
	return len(buffer), base64.StdEncoding.EncodeToString(buffer), nil
}

func (c HashedStringKeyCodec) SizeNonTerminal(key string) int {
	return sha256.Size
}

func (c HashedStringKeyCodec) ComputeHash(name string) []byte {
	comps := strings.Split(name, ".")
	hsh := sha256.New()
	for i := len(comps) - 1; i >= 0; i-- {
		comp := strings.TrimSpace(comps[i])
		hsh.Write([]byte(comp))
	}
	return hsh.Sum(nil)
}

// GetNameKeyBytes returns the name key in the same format as before (0x03 + sha256(name))
func GetNameKeyBytes(name string) ([]byte, error) {
	hash, err := computeNameHash(name)
	if err != nil {
		return nil, err
	}
	return append(append([]byte{}, NameKeyPrefix...), hash...), nil
}

func GetAddressKeyPrefix(addr sdk.AccAddress) ([]byte, error) {
	if err := sdk.VerifyAddressFormat(addr); err != nil {
		return nil, err
	}
	key := append([]byte{}, AddressKeyPrefix...) // safe copy
	key = append(key, address.MustLengthPrefix(addr)...)
	return key, nil
}

// GetAddressKeyBytes returns the full address key
func GetAddressKeyBytes(addr sdk.AccAddress, nameKey []byte) ([]byte, error) {
	prefix, err := GetAddressKeyPrefix(addr)
	if err != nil {
		return nil, err
	}
	return append(prefix, nameKey...), nil
}

// computeNameHash replicates the exact hashing logic from the old system
func computeNameHash(name string) ([]byte, error) {
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

// ValidateAddress validates an account address
func ValidateAddress(addr sdk.AccAddress) error {
	return sdk.VerifyAddressFormat(addr)
}
