package types

import (
	"github.com/cosmos/cosmos-sdk/types/address"
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

// GetMsgBasedFeeKey Takes in MsgName
func GetMsgBasedFeeKey(p string) []byte {
	return append(MsgBasedFeeKeyPrefix, address.MustLengthPrefix([]byte(p))...)
}

var (
	MsgBasedFeeKeyPrefix = []byte{0x00}
)
