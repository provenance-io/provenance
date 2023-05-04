package types

import (
	"crypto/sha256"
	"encoding/binary"
	fmt "fmt"
	"strings"
)

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
	EventNameKeyLength = 32
)

var (
	TriggerKeyPrefix     = []byte{0x01}
	EventLookupKeyPrefix = []byte{0x01}
	NextTriggerIDKey     = []byte{0x02}
)

// GetEventLookupKey converts a trigger into key format.
func GetEventLookupKey(eventName string, id TriggerID) []byte {
	eventNameBytes := GetEventNameBytes(eventName)
	triggerIDBytes := make([]byte, TriggerIDKeyLength)
	binary.BigEndian.PutUint64(triggerIDBytes, id)

	key := EventLookupKeyPrefix
	key = append(key, triggerIDBytes...)
	key = append(key, eventNameBytes...)
	key = append(key, triggerIDBytes...)
	return key
}

// GetTriggerKey converts a trigger into key format.
func GetTriggerKey(id TriggerID) []byte {
	triggerIDBytes := make([]byte, TriggerIDKeyLength)
	binary.BigEndian.PutUint64(triggerIDBytes, id)

	key := TriggerKeyPrefix
	key = append(key, triggerIDBytes...)
	return key
}

func GetNextTriggerIDKey() []byte {
	return NextTriggerIDKey
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

// GetEventNameBytes returns a set of bytes that uniquely identifies the given event name
func GetEventNameBytes(name string) []byte {
	eventName := strings.ToLower(strings.TrimSpace(name))
	if len(eventName) == 0 {
		panic(fmt.Sprintf("invalid event name %s", name))
	}
	hash := sha256.Sum256([]byte(eventName))
	return hash[:]
}
