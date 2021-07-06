package v042

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// AddressKeyPrefix is a prefix added to keys for indexing name records by address.
	AddressNamePrefix = []byte{0x03}
	// AddressKeyPrefixLegacy is a prefix added to keys for indexing name records by address.
	AddressKeyPrefixLegacy = []byte{0x04}
	// NameAddressLengthLegacy is the legacy length of pre v043 address
	NameAddressLengthLegacy = 20
)

func ConvertLegacyNameAddress(legacyAddr sdk.AccAddress) sdk.AccAddress {
	padding := make([]byte, 12)
	updatedAddr := append(legacyAddr.Bytes(), padding...)
	return sdk.AccAddress(updatedAddr)
}

// GetAddressKeyPrefix returns a store key for a name record address
func GetAddressKeyPrefixLegacy(address sdk.AccAddress) (key []byte, err error) {
	err = ValidateAddress(address)
	if err == nil {
		key = AddressKeyPrefixLegacy
		key = append(key, address.Bytes()...)
	}
	return
}

func ValidateAddress(addr sdk.AccAddress) error {
	if len(addr.Bytes()) != NameAddressLengthLegacy {
		return fmt.Errorf("unexpected key length (%d ≠ %d)", len(addr.Bytes()), NameAddressLengthLegacy)
	}
	if err := sdk.VerifyAddressFormat(addr); err != nil {
		return err
	}
	return nil
}
