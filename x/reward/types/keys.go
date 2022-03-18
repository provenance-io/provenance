package types

import (
	"strconv"
)

const (
	// ModuleName defines the module name
	ModuleName = "reward"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	RewardProgramKeyPrefix = []byte{0x01}
)

// GetNameKeyPrefix converts a name into key format.
func GetRewardProgramKeyPrefix(id int64) []byte {
	idByte := []byte(strconv.FormatInt(id, 10))
	return append(RewardProgramKeyPrefix, idByte...)
}
