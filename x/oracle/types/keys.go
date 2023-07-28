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

var (
	// PortKey defines the key to store the port ID in store
	PortKey = KeyPrefix("oracle-port-")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
