package v042

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (

	// AddressScopeCacheKeyPrefixLegacy for scope to address cache lookup
	AddressScopeCacheKeyPrefixLegacy = []byte{0x10}
	// ValueOwnerScopeCacheKeyPrefixLegacy for scope to value owner address cache lookup
	ValueOwnerScopeCacheKeyPrefixLegacy = []byte{0x12}

	// AddressScopeSpecCacheKeyPrefixLegacy for scope spec lookup by address
	AddressScopeSpecCacheKeyPrefixLegacy = []byte{0x13}
	// AddressContractSpecCacheKeyPrefixLegacy for contract spec lookup by address
	AddressContractSpecCacheKeyPrefixLegacy = []byte{0x15}
	// OSLocatorAddressKeyPrefixLegacy is the key for OSLocator Record by address
	OSLocatorAddressKeyPrefixLegacy = []byte{0x16}
	// AddrLengthLegacy is the length of all keys in versions < 43
	AddrLengthLegacy = 20
)

// GetOSLocatorKeyLegacy returns a store key for an object store locator entry
func GetOSLocatorKeyLegacy(addr sdk.AccAddress) []byte {
	if len(addr.Bytes()) != AddrLengthLegacy {
		panic(fmt.Sprintf("unexpected key length (%d ≠ %d)", len(addr.Bytes()), AddrLengthLegacy))
	}
	return append(OSLocatorAddressKeyPrefixLegacy, addr.Bytes()...)
}

// GetOSLocatorKey returns a store key for an object store locator entry
func GetAddressScopeCacheKeyLegacy(addr sdk.AccAddress) []byte {
	if len(addr.Bytes()) != AddrLengthLegacy {
		panic(fmt.Sprintf("unexpected key length (%d ≠ %d)", len(addr.Bytes()), AddrLengthLegacy))
	}
	return append(OSLocatorAddressKeyPrefixLegacy, addr.Bytes()...)
}
