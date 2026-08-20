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
	comps, err := getNameSegments(name)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256([]byte(strings.Join(comps, ".")))
	key = NameKeyPrefix
	return append(key, sum[:]...), nil
}

// GetLegacyNameKeyPrefix converts a name into the key format used by the name module before version 3.
// Those keys are a hash of just the name segments, so names with different segments, e.g. "leaf.mid.root"
// and "midleaf.root", can end up sharing a key. It is only still around so that the migration can find them.
func GetLegacyNameKeyPrefix(name string) (key []byte, err error) {
	comps, err := getNameSegments(name)
	if err != nil {
		return nil, err
	}
	hsh := sha256.New()
	for i := len(comps) - 1; i >= 0; i-- {
		if _, err = hsh.Write([]byte(comps[i])); err != nil {
			return nil, err
		}
	}
	key = NameKeyPrefix
	return append(key, hsh.Sum(nil)...), nil
}

// getNameSegments splits a name into its space-trimmed segments, returning an error if any of them are empty.
func getNameSegments(name string) ([]string, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("name can not be empty: %w", ErrNameInvalid)
	}
	comps := strings.Split(name, ".")
	for i, comp := range comps {
		comps[i] = strings.TrimSpace(comp)
		if len(comps[i]) == 0 {
			return nil, fmt.Errorf("name segment cannot be empty: %w", ErrNameInvalid)
		}
	}
	return comps, nil
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
