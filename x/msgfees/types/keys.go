package types

import (
	"crypto/sha256"
)

const (
	// ModuleName defines the module name
	ModuleName = "msgfees"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_msgfees"
)

// GetMsgFeeKey takes in msgType name and returns key
func GetMsgFeeKey(msgType string) []byte {
	msgNameBytes := sha256.Sum256([]byte(msgType))
	return append(MsgFeeKeyPrefix, msgNameBytes[0:16]...)
}

var (
	MsgFeeKeyPrefix = []byte{0x00}
)
