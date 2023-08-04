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

// The Oracle module's KVStore categorizes each item in the store using a single byte prefix
// Any additional bytes appended after this prefix are to help in making multiple unique entries per category
// The keys in this section are relatively simple and are used for module setup and configuration
//
//	OracleStoreKey
//	- 0x01: sdk.AccAddress
//	  | 1 |
//
//	LastQueryPacketSeqKey
//	- 0x02: uint64
//	  | 1 |
//
//	PortStoreKey
//	- 0x03: string
//	  | 1 |
//
// The following keys are used to store ICQ oracle requests and responses
// The <sequence_id_bytes> are 8 bytes to uniquely identify a ICQ request
//
//	QueryRequestStoreKey
//	- 0x04<sequence_id_bytes>: QueryOracleRequest
//	  | 1 |        8         |
//
//	QueryResponseStoreKey
//	- 0x05<sequence_id_bytes>: QueryOracleResponse
//	  | 1 |        8         |
var (
	// OracleStoreKey is the key for the module's oracle address
	OracleStoreKey = []byte{0x01}
	// LastQueryPacketSeqKey is the key for the last packet sequence
	LastQueryPacketSeqKey = []byte{0x02}
	// PortStoreKey defines the key to store the port ID in store
	PortStoreKey = []byte{0x03}
	// QueryResponseStoreKeyPrefix is a prefix for storing request
	QueryRequestStoreKeyPrefix = []byte{0x04}
	// QueryResponseStoreKeyPrefix is a prefix for storing result
	QueryResponseStoreKeyPrefix = []byte{0x05}
)

// GetOracleStoreKey is a function to get the key for the oracle's address in store
func GetOracleStoreKey() []byte {
	return OracleStoreKey
}

// GetPortStoreKey is a function to get the key for the port in store
func GetPortStoreKey() []byte {
	return PortStoreKey
}

// GetLastQueryPacketSeqKey is a function to get the key for the last query packet sequence in store
func GetLastQueryPacketSeqKey() []byte {
	return LastQueryPacketSeqKey
}

// GetQueryRequestStoreKey is a function to generate key for each result in store
func GetQueryRequestStoreKey(packetSequence uint64) []byte {
	return append(QueryRequestStoreKeyPrefix, uint64ToBytes(packetSequence)...)
}

// GetQueryResponseStoreKey is a function to generate key for each result in store
func GetQueryResponseStoreKey(packetSequence uint64) []byte {
	return append(QueryResponseStoreKeyPrefix, uint64ToBytes(packetSequence)...)
}

func uint64ToBytes(num uint64) []byte {
	result := make([]byte, 8)
	binary.BigEndian.PutUint64(result, uint64(num))
	return result
}
