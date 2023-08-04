package types

import (
	"encoding/binary"

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
	// ContractStoreKey is the key for the oracle's contract address
	ContractStoreKey = []byte{0x01}
	// LastQueryPacketSeqKey is the key for the last packet sequence
	LastQueryPacketSeqKey = []byte{0x02}
	// PortStoreKey defines the key to store the port ID in store
	PortStoreKey = []byte{0x03}
	// QueryResponseStoreKeyPrefix is a prefix for storing request
	QueryRequestStoreKeyPrefix = []byte{0x04}
	// QueryResponseStoreKeyPrefix is a prefix for storing result
	QueryResponseStoreKeyPrefix = []byte{0x05}
)

// ContractStoreKey is a function to get the key for the oracle's contract in store
func GetContractStoreKey() []byte {
	return ContractStoreKey
}

// QueryRequestStoreKey is a function to generate key for each result in store
func QueryRequestStoreKey(packetSequence uint64) []byte {
	return append(QueryRequestStoreKeyPrefix, uint64ToBytes(packetSequence)...)
}

// QueryResponseStoreKey is a function to generate key for each result in store
func QueryResponseStoreKey(packetSequence uint64) []byte {
	return append(QueryResponseStoreKeyPrefix, uint64ToBytes(packetSequence)...)
}

func uint64ToBytes(num uint64) []byte {
	result := make([]byte, 8)
	binary.BigEndian.PutUint64(result, uint64(num))
	return result
}
