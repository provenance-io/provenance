package types

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"cosmossdk.io/collections/codec"

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

// GetNameKeyPrefix converts a name into key format.
func GetNameKeyPrefix(name string) (key []byte, err error) {
	key = NameKeyPrefix
	return getNamePrefixByType(name, key)
}

// internal common code for legacy and current way.
func getNamePrefixByType(name string, key []byte) ([]byte, error) {
	var err error
	if strings.TrimSpace(name) == "" {
		err = fmt.Errorf("name can not be empty: %w", ErrNameInvalid)
		return nil, err
	}
	comps := strings.Split(name, ".")
	hsh := sha256.New()
	for i := len(comps) - 1; i >= 0; i-- {
		comp := strings.TrimSpace(comps[i])
		if len(comp) == 0 {
			err = fmt.Errorf("name segment cannot be empty: %w", ErrNameInvalid)
			return nil, err
		}
		if _, err = hsh.Write([]byte(comp)); err != nil {
			return nil, err
		}
	}
	sum := hsh.Sum(nil)
	key = append(key, sum...)
	return key, nil
}

// GetAddressKeyPrefix returns a store key for a name record address
func GetAddressKeyPrefix(addr sdk.AccAddress) (key []byte, err error) {
	err = sdk.VerifyAddressFormat(addr.Bytes())
	if err == nil {
		key = AddressKeyPrefix
		key = append(key, address.MustLengthPrefix(addr.Bytes())...)
	}
	return
}

func ValidateAddress(address sdk.AccAddress) error {
	return sdk.VerifyAddressFormat(address)
}

type rawBytesKeyCodec struct{}

// RawBytesKey is a codec for []byte keys without any length prefixing or transformation.
var RawBytesKey codec.KeyCodec[[]byte] = rawBytesKeyCodec{}

// Encode writes the raw bytes to the buffer.
// It expects the buffer to be at least len(key).
func (r rawBytesKeyCodec) Encode(buffer []byte, key []byte) (int, error) {
	if len(buffer) < len(key) {
		return 0, fmt.Errorf("buffer too small")
	}
	copy(buffer, key)
	return len(key), nil
}

// Decode reads the full slice as the key.
func (rawBytesKeyCodec) Decode(buffer []byte) (int, []byte, error) {
	return len(buffer), buffer, nil
}

// Size returns the exact size of the key.
func (rawBytesKeyCodec) Size(key []byte) int {
	return len(key)
}

// EncodeJSON encodes key to JSON (just as a base64 or raw array).
func (rawBytesKeyCodec) EncodeJSON(value []byte) ([]byte, error) {
	return json.Marshal(value)
}

// DecodeJSON decodes key from JSON.
func (rawBytesKeyCodec) DecodeJSON(b []byte) ([]byte, error) {
	var result []byte
	err := json.Unmarshal(b, &result)
	return result, err
}

// Stringify returns a readable representation of the key.
func (rawBytesKeyCodec) Stringify(key []byte) string {
	return string(key)
}

// KeyType returns the type name for this codec.
func (rawBytesKeyCodec) KeyType() string {
	return "rawbytes"
}

// EncodeNonTerminal is used in composite keys and behaves the same as Encode here.
func (rawBytesKeyCodec) EncodeNonTerminal(buffer []byte, key []byte) (int, error) {
	return rawBytesKeyCodec{}.Encode(buffer, key)
}

// DecodeNonTerminal behaves the same as Decode here.
func (rawBytesKeyCodec) DecodeNonTerminal(buffer []byte) (int, []byte, error) {
	return rawBytesKeyCodec{}.Decode(buffer)
}

// SizeNonTerminal behaves the same as Size.
func (rawBytesKeyCodec) SizeNonTerminal(key []byte) int {
	return len(key)
}
