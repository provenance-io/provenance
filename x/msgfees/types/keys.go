package types

import (
	"crypto/sha256"
	"fmt"
	"strings"
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

	// CompositeKeyDelimiter is the delimiter of msgTypeUrl and recipient
	CompositeKeyDelimiter = "\n"
)

// GetMsgFeeKey takes in msgType name and returns key
func GetMsgFeeKey(msgType string) []byte {
	msgNameBytes := sha256.Sum256([]byte(msgType))
	return append(MsgFeeKeyPrefix, msgNameBytes[0:16]...)
}

var (
	MsgFeeKeyPrefix = []byte{0x00}
)

func GetCompositeKey(msgType string, recipient string) string {
	if len(recipient) == 0 {
		return msgType
	}
	return fmt.Sprintf("%s%s%s", msgType, CompositeKeyDelimiter, recipient)
}

// SplitCompositKey splits the composite key into msgType and recipient, if recipient is empty then it is for the fee module
func SplitCompositeKey(key string) (msgType, recipient string) {
	msgAccountPair := strings.Split(key, CompositeKeyDelimiter)
	addressKey := ""
	if len(msgAccountPair) == 2 && len(msgAccountPair[1]) > 0 {
		addressKey = msgAccountPair[1]
	}
	return msgAccountPair[0], addressKey
}
