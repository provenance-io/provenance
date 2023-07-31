package types

import "encoding/binary"

var (
	// QueryResponseStoreKeyPrefix is a prefix for storing request
	QueryRequestStoreKeyPrefix = "coin_rates_request"

	// QueryResponseStoreKeyPrefix is a prefix for storing result
	QueryResponseStoreKeyPrefix = "coin_rates_response"

	// LastQueryPacketSeqKey is the key for the last packet sequence
	LastQueryPacketSeqKey = "coin_rates_last_id"
)

// QueryRequestStoreKey is a function to generate key for each result in store
func QueryRequestStoreKey(packetSequence uint64) []byte {
	return append(KeyPrefix(QueryRequestStoreKeyPrefix), uint64ToBytes(packetSequence)...)
}

// QueryResponseStoreKey is a function to generate key for each result in store
func QueryResponseStoreKey(packetSequence uint64) []byte {
	return append(KeyPrefix(QueryResponseStoreKeyPrefix), uint64ToBytes(packetSequence)...)
}

func uint64ToBytes(num uint64) []byte {
	result := make([]byte, 8)
	binary.BigEndian.PutUint64(result, uint64(num))
	return result
}
