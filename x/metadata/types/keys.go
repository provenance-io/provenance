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
// - 0x01<scope_key_bytes><group_id_bytes>: RecordGroup
//
// - 0x02<scope_key_bytes><record_name_bytes>: Record
//
// - 0x03<group_specification_hash>: ContractSpecification
//
// - 0x04<scope_specification_id_bytes>: ScopeSpecification
//
// - 0x10<party_address><scope_key_bytes>: Role
var (
	// ScopeKeyPrefix is the key for scope records in metadata store
	ScopeKeyPrefix = []byte{0x00}
	// GroupKeyPrefix is the key for group records in metadata store
	GroupKeyPrefix = []byte{0x01}
	// RecordKeyPrefix is the key for records within scopes in metadata store
	RecordKeyPrefix = []byte{0x02}
	// ContractSpecificationPrefix is the key for group specification instances in metadata store
	ContractSpecificationPrefix = []byte{0x03}
	// ScopeSpecificationPrefix is the key for scope specifications in metadata store
	ScopeSpecificationPrefix = []byte{0x04}

	// AddressCacheKeyPrefix for scope to address cache lookup
	AddressCacheKeyPrefix = []byte{0x10}
	// ScopeSpecCacheKeyPrefix for scope to scope specification cache lookup
	ScopeSpecCacheKeyPrefix = []byte{0x11}
	// ValueOwnerCacheKeyPrefix for scope to value owner address cache lookup
	ValueOwnerCacheKeyPrefix = []byte{0x12}
)

// GetAddressCacheIteratorPrefix returns an iterator prefix for all scope cache entries assigned to a given address
func GetAddressCacheIteratorPrefix(addr sdk.AccAddress) []byte {
	return append(AddressCacheKeyPrefix, addr.Bytes()...)
}

// GetAddressCacheKey returns the store key for an address cache entry
func GetAddressCacheKey(addr sdk.AccAddress, scopeID MetadataAddress) []byte {
	return append(GetAddressCacheIteratorPrefix(addr), scopeID.Bytes()...)
}

// GetScopeSpecCacheIteratorPrefix returns an iterator prefix for all scope cache entries assigned to a given address
func GetScopeSpecCacheIteratorPrefix(scopeSpecID MetadataAddress) []byte {
	return append(ScopeSpecCacheKeyPrefix, scopeSpecID.Bytes()...)
}

// GetScopeSpecCacheKey returns the store key for an address cache entry
func GetScopeSpecCacheKey(scopeSpecID MetadataAddress, scopeID MetadataAddress) []byte {
	return append(GetScopeSpecCacheIteratorPrefix(scopeSpecID), scopeID.Bytes()...)
}

// GetValueOwnerCacheIteratorPrefix returns an iterator prefix for all scope cache entries assigned to a given address
func GetValueOwnerCacheIteratorPrefix(addr sdk.AccAddress) []byte {
	return append(ValueOwnerCacheKeyPrefix, addr.Bytes()...)
}

// GetValueOwnerCacheKey returns the store key for an address cache entry
func GetValueOwnerCacheKey(addr sdk.AccAddress, scopeID MetadataAddress) []byte {
	return append(GetValueOwnerCacheIteratorPrefix(addr), scopeID.Bytes()...)
}
