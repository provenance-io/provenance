package v042

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/provenance-io/provenance/x/metadata/types"
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
	// ScopeSpecScopeCacheKeyPrefix for scope to scope specification cache lookup
	ScopeSpecScopeCacheKeyPrefix = []byte{0x11}
	// OSLocatorAddressKeyPrefix is the key for OSLocator Record by address
	OSLocatorAddressKeyPrefix = []byte{0x21}
)

// GetOSLocatorKeyLegacy returns a store key for an object store locator entry
func GetOSLocatorKeyLegacy(addr sdk.AccAddress) []byte {
	if len(addr.Bytes()) != AddrLengthLegacy {
		panic(fmt.Sprintf("unexpected key length (%d ≠ %d)", len(addr.Bytes()), AddrLengthLegacy))
	}
	return append(OSLocatorAddressKeyPrefixLegacy, addr.Bytes()...)
}

// GetAddressScopeCacheIteratorPrefix returns an iterator prefix for all scope cache entries assigned to a given address
func GetAddressScopeCacheIteratorPrefixLegacy(addr sdk.AccAddress) []byte {
	return append(AddressScopeCacheKeyPrefixLegacy, address.MustLengthPrefix(addr.Bytes())...)
}

// GetAddressScopeCacheKey returns the store key for an address cache entry
func GetAddressScopeCacheKeyLegacy(addr sdk.AccAddress, scopeID types.MetadataAddress) []byte {
	return append(GetAddressScopeCacheIteratorPrefixLegacy(addr), scopeID.Bytes()...)
}

// GetScopeSpecScopeCacheIteratorPrefix returns an iterator prefix for all scope cache entries assigned to a given address
func GetScopeSpecScopeCacheIteratorPrefixLegacy(scopeSpecID types.MetadataAddress) []byte {
	return append(ScopeSpecScopeCacheKeyPrefix, scopeSpecID.Bytes()...)
}

// GetScopeSpecScopeCacheKey returns the store key for an address cache entry
func GetScopeSpecScopeCacheKeyLegacy(scopeSpecID types.MetadataAddress, scopeID types.MetadataAddress) []byte {
	return append(GetScopeSpecScopeCacheIteratorPrefixLegacy(scopeSpecID), scopeID.Bytes()...)
}

// GetValueOwnerScopeCacheIteratorPrefix returns an iterator prefix for all scope cache entries assigned to a given address
func GetValueOwnerScopeCacheIteratorPrefixLegacy(addr sdk.AccAddress) []byte {
	return append(ValueOwnerScopeCacheKeyPrefixLegacy, address.MustLengthPrefix(addr.Bytes())...)
}

// GetValueOwnerScopeCacheKey returns the store key for an address cache entry
func GetValueOwnerScopeCacheKeyLegacy(addr sdk.AccAddress, scopeID types.MetadataAddress) []byte {
	return append(GetValueOwnerScopeCacheIteratorPrefixLegacy(addr), scopeID.Bytes()...)
}

// GetAddressScopeSpecCacheIteratorPrefix returns an iterator prefix for all scope spec cache entries assigned to a given address
func GetAddressScopeSpecCacheIteratorPrefixLegacy(addr sdk.AccAddress) []byte {
	return append(AddressScopeSpecCacheKeyPrefixLegacy, address.MustLengthPrefix(addr.Bytes())...)
}

// GetAddressScopeSpecCacheKey returns the store key for an address + scope spec cache entry
func GetAddressScopeSpecCacheKeyLegacy(addr sdk.AccAddress, scopeSpecID types.MetadataAddress) []byte {
	return append(GetAddressScopeSpecCacheIteratorPrefixLegacy(addr), scopeSpecID.Bytes()...)
}

// GetAddressContractSpecCacheIteratorPrefix returns an iterator prefix for all contract spec cache entries assigned to a given address
func GetAddressContractSpecCacheIteratorPrefixLegacy(addr sdk.AccAddress) []byte {
	return append(AddressContractSpecCacheKeyPrefixLegacy, address.MustLengthPrefix(addr.Bytes())...)
}

// GetAddressContractSpecCacheKeyLegacy returns the store key for an address + contract spec cache entry
func GetAddressContractSpecCacheKeyLegacy(addr sdk.AccAddress, contractSpecID types.MetadataAddress) []byte {
	return append(GetAddressContractSpecCacheIteratorPrefixLegacy(addr), contractSpecID.Bytes()...)
}
