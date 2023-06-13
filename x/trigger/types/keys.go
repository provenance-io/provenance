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

	EventOrderLength = 8
	TriggerIDLength  = 8
	QueueIndexLength = 8
	GasLimitLength   = 8
)

// KVStore Key Prefixes used for iterator/scans against the store and identification of key types
// The following keys are used to store Triggers
// The <trigger_id_bytes> are 8 bytes to uniquely identify a trigger
// The remaining key is used to track the next valid trigger id
//
//   - 0x01<trigger_id_bytes>: Trigger
//     | 1 |        8        |
//
//   - 0x05: Trigger ID
//     | 1 |
//
// The keys prefixed with 0x02 are used to quickly iterate trigger events during the detection phase
// The <event_type_bytes> is 32 bytes representing the event type's name.
// The next 8 bytes <order_bytes> are to help order the events and improve searching
// The last 8 bytes <trigger_id_bytes> is a the trigger id that the event listener belongs to.
//   - 0x02<event_type_bytes><order_bytes><trigger_id_bytes>: []byte{}
//     | 1 |       32       |      8      |        8        |
//
// The key in this section is used to track gas limits for triggers.
// The <trigger_id_bytes> are 8 bytes that match the trigger that the gas limit belongs to.
//
//   - 0x04<trigger_id_bytes>: uint64 (GasLimit)
//     | 1 |        8        |
//
// These keys are used to represent a queue in the store
// The <queue_index_bytes> are 8 bytes representing the item's spot in the queue
// The two additional keys are used as metadata to track and manage the queue.
//
//   - 0x03<queue_index_bytes>: QueuedTrigger
//     | 1 |        8         |
//
//   - 0x06: uint64 (Queue Start Index)
//     | 1 |
//
//   - 0x07: uint64 (Queue Length)
//     | 1 |
var (
	// TriggerKeyPrefix is an initial byte to help group all trigger keys
	TriggerKeyPrefix = []byte{0x01}
	// EventListenerKeyPrefix is an initial byte to help group all event listener keys
	EventListenerKeyPrefix = []byte{0x02}
	// QueueKeyPrefix is an initial byte to help group all queue entry keys
	QueueKeyPrefix = []byte{0x03}
	// GasLimitKeyPrefix is an initial byte to help group all gas limit keys
	GasLimitKeyPrefix = []byte{0x04}
	// NextTriggerIDKey is the key to obtain the next valid trigger id
	NextTriggerIDKey = []byte{0x05}
	// QueueStartIndexKey is the key to obtain the queue's starting index
	QueueStartIndexKey = []byte{0x06}
	// QueueStartIndexKey is the key to obtain the queue's length
	QueueLengthKey = []byte{0x07}
)

// GetEventListenerKey converts an event name, order, and trigger ID into an event registry key format.
func GetEventListenerKey(eventName string, order uint64, id TriggerID) []byte {
	triggerIDBytes := make([]byte, TriggerIDLength)
	binary.BigEndian.PutUint64(triggerIDBytes, id)
	orderBytes := make([]byte, EventOrderLength)
	binary.BigEndian.PutUint64(orderBytes, order)

	key := GetEventListenerPrefix(eventName)
	key = append(key, orderBytes...)
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
		panic(fmt.Sprintf("invalid event name: %s", name))
	}
	hash := sha256.Sum256([]byte(eventName))
	return hash[:]
}
