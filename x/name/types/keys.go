package types

import (
	"crypto/sha256"
	"fmt"
	"strings"

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

// GetNameKeySuffix returns the name key suffix (hash part only)
func GetNameKeySuffix(name string) ([]byte, error) {
	normalized := NormalizeName(name)
	if strings.TrimSpace(normalized) == "" {
		return nil, fmt.Errorf("name can not be empty: %w", ErrNameInvalid)
	}

	comps := strings.Split(normalized, ".")
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

// GetAddressKeySuffix returns the address key suffix
func GetAddressKeySuffix(addr sdk.AccAddress, nameKeySuffix []byte) ([]byte, error) {
	if err := sdk.VerifyAddressFormat(addr); err != nil {
		return nil, err
	}
	key := make([]byte, 0, len(addr)+1+len(nameKeySuffix))
	key = append(key, address.MustLengthPrefix(addr)...)
	key = append(key, nameKeySuffix...)
	return key, nil
}
