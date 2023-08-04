package types

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetContractStoreKey(t *testing.T) {
	key := GetContractStoreKey()
	assert.EqualValues(t, ContractStoreKey, key[0:1], "must return correct contract key")
}

func TestGetPortStoreKey(t *testing.T) {
	key := GetPortStoreKey()
	assert.EqualValues(t, PortStoreKey, key[0:1], "must return correct port key")
}

func TestGetLastQueryPacketSeqKey(t *testing.T) {
	key := GetLastQueryPacketSeqKey()
	assert.EqualValues(t, LastQueryPacketSeqKey, key[0:1], "must return correct last query packet sequence key")
}

func TestQueryRequestStoreKey(t *testing.T) {
	key := GetQueryRequestStoreKey(5)
	assert.EqualValues(t, QueryRequestStoreKeyPrefix, key[0:1], "must have correct prefix")
	assert.EqualValues(t, int(5), int(binary.BigEndian.Uint64(key[1:9])), "must have correct sequence")
}

func TestQueryResponseStoreKey(t *testing.T) {
	key := GetQueryResponseStoreKey(5)
	assert.EqualValues(t, QueryResponseStoreKeyPrefix, key[0:1], "must have correct prefix")
	assert.EqualValues(t, int(5), int(binary.BigEndian.Uint64(key[1:9])), "must have correct sequence")
}
