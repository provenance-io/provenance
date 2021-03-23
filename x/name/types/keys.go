package types

import (
	"crypto/sha256"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "name"

	// StoreKey is the store key string for distribution
	StoreKey = ModuleName

	// RouterKey is the message route for distribution
	RouterKey = ModuleName

	// QuerierRoute is the querier route for distribution
	QuerierRoute = ModuleName
)

var (
	// NameKeyPrefix is a prefix added to keys for adding/querying names.
	NameKeyPrefix = []byte{0x01}
	// AddressKeyPrefix is a prefix added to keys for indexing name records by address.
	AddressKeyPrefix = []byte{0x02}
)

// GetNameKeyPrefix converts a name into key format.
func GetNameKeyPrefix(name string) (key []byte, err error) {
	if strings.TrimSpace(name) == "" {
		err = fmt.Errorf("name can not be empty: %w", ErrNameInvalid)
		return
	}
	comps := strings.Split(name, ".")
	hsh := sha256.New()
	for i := len(comps) - 1; i >= 0; i-- {
		comp := strings.TrimSpace(comps[i])
		if len(comp) == 0 {
			err = fmt.Errorf("name segment cannot be empty: %w", ErrNameInvalid)
			return
		}
		if _, err = hsh.Write([]byte(comp)); err != nil {
			return
		}
	}
	sum := hsh.Sum(nil)
	key = NameKeyPrefix
	key = append(key, sum...)
	return
}

// GetAddressKeyPrefix returns a store key for a name record address
func GetAddressKeyPrefix(address sdk.AccAddress) (key []byte, err error) {
	err = sdk.VerifyAddressFormat(address.Bytes())
	if err == nil {
		key = AddressKeyPrefix
		key = append(key, address.Bytes()...)
	}
	return
}
