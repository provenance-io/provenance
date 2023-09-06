package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOracleStoreKey(t *testing.T) {
	key := GetOracleStoreKey()
	assert.EqualValues(t, OracleStoreKey, key[0:1], "must return correct oracle key")
}

func TestGetPortStoreKey(t *testing.T) {
	key := GetPortStoreKey()
	assert.EqualValues(t, PortStoreKey, key[0:1], "must return correct port key")
}
