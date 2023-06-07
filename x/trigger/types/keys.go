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

	TriggerIDLength  = 8
	QueueIndexLength = 8
	GasLimitLength   = 8
)

var (
	TriggerKeyPrefix       = []byte{0x01}
	EventListenerKeyPrefix = []byte{0x02}
	QueueKeyPrefix         = []byte{0x03}
	GasLimitKeyPrefix      = []byte{0x04}
	NextTriggerIDKey       = []byte{0x05}
	QueueStartIndexKey     = []byte{0x06}
	QueueLengthKey         = []byte{0x07}
)

// GetEventListenerKey converts an event name and trigger ID into an event registry key format.
func GetEventListenerKey(eventName string, id TriggerID) []byte {
	triggerIDBytes := make([]byte, TriggerIDLength)
	binary.BigEndian.PutUint64(triggerIDBytes, id)

	key := GetEventListenerPrefix(eventName)
	key = append(key, triggerIDBytes...)
	return key
}

// GetEventListenerPrefix converts an event name into a prefix for event registry.
func GetEventListenerPrefix(eventName string) []byte {
	eventNameBytes := GetEventNameBytes(eventName)

	key := EventListenerKeyPrefix
	key = append(key, eventNameBytes...)
	return key
}

// GetTriggerKey converts a trigger into key format.
func GetTriggerKey(id TriggerID) []byte {
	triggerIDBytes := make([]byte, TriggerIDLength)
	binary.BigEndian.PutUint64(triggerIDBytes, id)

	key := TriggerKeyPrefix
	key = append(key, triggerIDBytes...)
	return key
}

// GetNextTriggerIDKey gets the key for getting the next trigger ID.
func GetNextTriggerIDKey() []byte {
	return NextTriggerIDKey
}

// GetQueueKeyPrefix gets the queue's key prefix
func GetQueueKeyPrefix() []byte {
	return QueueKeyPrefix
}

// GetQueueKey gets the key for storing in the queue.
func GetQueueKey(index uint64) []byte {
	indexBytes := GetQueueIndexBytes(index)

	key := QueueKeyPrefix
	key = append(key, indexBytes...)
	return key
}

// GetQueueStartIndexKey gets the key for storing the start index
func GetQueueStartIndexKey() []byte {
	return QueueStartIndexKey
}

// GetQueueLengthKey gets the key for storing the queue length
func GetQueueLengthKey() []byte {
	return QueueLengthKey
}

// GetQueueIndexFromBytes returns the index in uint64 format from a byte array
func GetQueueIndexFromBytes(bz []byte) uint64 {
	return binary.BigEndian.Uint64(bz)
}

// GetQueueIndexBytes returns the byte representation of the queue index
func GetQueueIndexBytes(index uint64) (queueIndexBz []byte) {
	queueIndexBz = make([]byte, QueueIndexLength)
	binary.BigEndian.PutUint64(queueIndexBz, index)
	return
}

// GetTriggerIDFromBytes returns triggerID in uint64 format from a byte array
func GetTriggerIDFromBytes(bz []byte) TriggerID {
	return binary.BigEndian.Uint64(bz)
}

// GetTriggerIDBytes returns the byte representation of the trigger ID
func GetTriggerIDBytes(triggerID TriggerID) (triggerIDBz []byte) {
	triggerIDBz = make([]byte, TriggerIDLength)
	binary.BigEndian.PutUint64(triggerIDBz, triggerID)
	return
}

// GetGasLimitKey converts a gas limit into key format.
func GetGasLimitKey(id TriggerID) []byte {
	triggerIDBytes := make([]byte, TriggerIDLength)
	binary.BigEndian.PutUint64(triggerIDBytes, id)

	key := GasLimitKeyPrefix
	key = append(key, triggerIDBytes...)
	return key
}

// GetGasLimitBytes returns the byte representation of the gas limit
func GetGasLimitBytes(gasLimit uint64) (gasLimitBz []byte) {
	gasLimitBz = make([]byte, GasLimitLength)
	binary.BigEndian.PutUint64(gasLimitBz, gasLimit)
	return
}

// GetGasLimitFromBytes returns gas limit in uint64 format from a byte array
func GetGasLimitFromBytes(bz []byte) uint64 {
	return binary.BigEndian.Uint64(bz)
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
