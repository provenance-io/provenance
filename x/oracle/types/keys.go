package types

import (
	icqtypes "github.com/strangelove-ventures/async-icq/v6/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "oracle"

	// StoreKey is string representation of the store key for marker
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_interquery"

	// Version defines the current version the IBC module supports
	Version = icqtypes.Version

	// PortID is the default port id that module binds to
	PortID = "oracle"
)

// The Oracle module's KVStore categorizes each item in the store using a single byte prefix
// Any additional bytes appended after this prefix are to help in making multiple unique entries per category
// The keys are relatively simple and are used for module setup and configuration
//
//	OracleStoreKey
//	- 0x01: sdk.AccAddress
//	  | 1 |
//
//
//	PortStoreKey
//	- 0x02: string
//	  | 1 |
var (
	// OracleStoreKey is the key for the module's oracle address
	OracleStoreKey = []byte{0x01}
	// PortStoreKey defines the key to store the port ID in store
	PortStoreKey = []byte{0x02}
)

// GetOracleStoreKey is a function to get the key for the oracle's address in store
func GetOracleStoreKey() []byte {
	return OracleStoreKey
}

// GetPortStoreKey is a function to get the key for the port in store
func GetPortStoreKey() []byte {
	return PortStoreKey
}
