package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "metadata"

	// StoreKey is string representation of the store key for metadata
	StoreKey = ModuleName

	// RouterKey to be used for routing msgs
	RouterKey = ModuleName

	// QuerierRoute to be used for queries
	QuerierRoute = ModuleName

	// CoinPoolName to be used for coin pool associated with mint/burn activities.
	CoinPoolName = ModuleName

	// DefaultParamspace is the name used for the parameter subspace for this module.
	DefaultParamspace = ModuleName
)

// KVStore Key Prefixes used for iterator/scans against the store and identification of key types
// Items are stored with the following key: values  these keys are handled using the MetadataAddress class.
//
// - 0x00<scope_key_bytes>: Scope
//
// - 0x01<scope_key_bytes><session_id_bytes>: Session
//
// - 0x02<scope_key_bytes><record_name_bytes>: Record
//
// - 0x03<session_specification_hash>: ContractSpecification
//
// - 0x04<scope_specification_id_bytes>: ScopeSpecification
//
// - 0x10<party_address><scope_key_bytes>: 0x01
//
// - 0x11<scope_spec_id><scope_id>: 0x01
//
// - 0x12<value_owner_address><scope_id>: 0x01
//
// - 0x13<owner_address><scope_spec_id>: 0x01
//
// - 0x14<contract_spec_id><scope_spec_id>: 0x01
//
// - 0x15<owner_address><contract_spec_id>: 0x01
var (
	// ScopeKeyPrefix is the key for scope records in metadata store
	ScopeKeyPrefix = []byte{0x00}
	// SessionKeyPrefix is the key for session records in metadata store
	SessionKeyPrefix = []byte{0x01}
	// RecordKeyPrefix is the key for records within scopes in metadata store
	RecordKeyPrefix = []byte{0x02}
	// ContractSpecificationKeyPrefix is the key for session specification instances in metadata store
	ContractSpecificationKeyPrefix = []byte{0x03}
	// ScopeSpecificationKeyPrefix is the key for scope specifications in metadata store
	ScopeSpecificationKeyPrefix = []byte{0x04}
	// RecordSpecificationKeyPrefix is the key for record specifications in metadata store
	RecordSpecificationKeyPrefix = []byte{0x05}

	// AddressScopeCacheKeyPrefix for scope to address cache lookup
	AddressScopeCacheKeyPrefix = []byte{0x10}
	// ScopeSpecScopeCacheKeyPrefix for scope to scope specification cache lookup
	ScopeSpecScopeCacheKeyPrefix = []byte{0x11}
	// ValueOwnerScopeCacheKeyPrefix for scope to value owner address cache lookup
	ValueOwnerScopeCacheKeyPrefix = []byte{0x12}

	// AddressScopeSpecCacheKeyPrefix for scope spec lookup by address
	AddressScopeSpecCacheKeyPrefix = []byte{0x13}
	// ContractSpecScopeSpecCacheKeyPrefix for scope spec lookup by contract spec
	ContractSpecScopeSpecCacheKeyPrefix = []byte{0x14}
	// AddressContractSpecCacheKeyPrefix for contract spec lookup by address
	AddressContractSpecCacheKeyPrefix = []byte{0x15}
)

// GetAddressScopeCacheIteratorPrefix returns an iterator prefix for all scope cache entries assigned to a given address
func GetAddressScopeCacheIteratorPrefix(addr sdk.AccAddress) []byte {
	return append(AddressScopeCacheKeyPrefix, addr.Bytes()...)
}

// GetAddressScopeCacheKey returns the store key for an address cache entry
func GetAddressScopeCacheKey(addr sdk.AccAddress, scopeID MetadataAddress) []byte {
	return append(GetAddressScopeCacheIteratorPrefix(addr), scopeID.Bytes()...)
}

// GetScopeSpecScopeCacheIteratorPrefix returns an iterator prefix for all scope cache entries assigned to a given address
func GetScopeSpecScopeCacheIteratorPrefix(scopeSpecID MetadataAddress) []byte {
	return append(ScopeSpecScopeCacheKeyPrefix, scopeSpecID.Bytes()...)
}

// GetScopeSpecScopeCacheKey returns the store key for an address cache entry
func GetScopeSpecScopeCacheKey(scopeSpecID MetadataAddress, scopeID MetadataAddress) []byte {
	return append(GetScopeSpecScopeCacheIteratorPrefix(scopeSpecID), scopeID.Bytes()...)
}

// GetValueOwnerScopeCacheIteratorPrefix returns an iterator prefix for all scope cache entries assigned to a given address
func GetValueOwnerScopeCacheIteratorPrefix(addr sdk.AccAddress) []byte {
	return append(ValueOwnerScopeCacheKeyPrefix, addr.Bytes()...)
}

// GetValueOwnerScopeCacheKey returns the store key for an address cache entry
func GetValueOwnerScopeCacheKey(addr sdk.AccAddress, scopeID MetadataAddress) []byte {
	return append(GetValueOwnerScopeCacheIteratorPrefix(addr), scopeID.Bytes()...)
}

// GetAddressScopeSpecCacheIteratorPrefix returns an iterator prefix for all scope spec cache entries assigned to a given address
func GetAddressScopeSpecCacheIteratorPrefix(addr sdk.AccAddress) []byte {
	return append(AddressScopeSpecCacheKeyPrefix, addr.Bytes()...)
}

// GetAddressScopeSpecCacheKey returns the store key for an address + scope spec cache entry
func GetAddressScopeSpecCacheKey(addr sdk.AccAddress, scopeSpecID MetadataAddress) []byte {
	return append(GetAddressScopeSpecCacheIteratorPrefix(addr), scopeSpecID.Bytes()...)
}

// GetContractSpecScopeSpecCacheIteratorPrefix returns an iterator prefix for all scope spec cache entries assigned to a given contract spec
func GetContractSpecScopeSpecCacheIteratorPrefix(contractSpecID MetadataAddress) []byte {
	return append(ContractSpecScopeSpecCacheKeyPrefix, contractSpecID.Bytes()...)
}

// GetContractSpecScopeSpecCacheKey returns the store key for a contract spec + scope spec cache entry
func GetContractSpecScopeSpecCacheKey(contractSpecID MetadataAddress, scopeSpecID MetadataAddress) []byte {
	return append(GetContractSpecScopeSpecCacheIteratorPrefix(contractSpecID), scopeSpecID.Bytes()...)
}

// GetAddressContractSpecCacheIteratorPrefix returns an iterator prefix for all contract spec cache entries assigned to a given address
func GetAddressContractSpecCacheIteratorPrefix(addr sdk.AccAddress) []byte {
	return append(AddressContractSpecCacheKeyPrefix, addr.Bytes()...)
}

// GetAddressContractSpecCacheKey returns the store key for an address + contract spec cache entry
func GetAddressContractSpecCacheKey(addr sdk.AccAddress, contractSpecID MetadataAddress) []byte {
	return append(GetAddressContractSpecCacheIteratorPrefix(addr), contractSpecID.Bytes()...)
}
