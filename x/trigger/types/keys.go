package types

import "encoding/binary"

const (
	// ModuleName defines the module name
	ModuleName = "trigger"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	TriggerIDKeyLength = 8
)

var (
	TriggerKeyPrefix = []byte{0x01}
	TriggerIDKey     = []byte{0x02}
)

// GetTriggerKey converts a trigger into key format.
func GetTriggerKey(id TriggerID) []byte {
	triggerIDBytes := make([]byte, TriggerIDKeyLength)
	binary.BigEndian.PutUint64(triggerIDBytes, id)
	return append(TriggerKeyPrefix, triggerIDBytes...)
}

func GetTriggerIDKey() []byte {
	return TriggerIDKey
}

// GetTriggerIDFromBytes returns triggerID in uint64 format from a byte array
func GetTriggerIDFromBytes(bz []byte) (triggerID TriggerID) {
	return binary.BigEndian.Uint64(bz)
}

// GetRewardProgramIDBytes returns the byte representation of the rewardprogramID
func GetTriggerIDBytes(triggerID TriggerID) (triggerIDBz []byte) {
	triggerIDBz = make([]byte, TriggerIDKeyLength)
	binary.BigEndian.PutUint64(triggerIDBz, triggerID)
	return
}
