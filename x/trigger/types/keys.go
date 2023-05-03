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
)

// GetTriggerKey converts a trigger into key format.
func GetTriggerKey(id TriggerID) []byte {
	triggerIDBytes := make([]byte, TriggerIDKeyLength)
	binary.BigEndian.PutUint64(triggerIDBytes, id)
	return append(TriggerKeyPrefix, triggerIDBytes...)
}
