package types

import (
	"crypto/sha256"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEventListenerKey(t *testing.T) {
	eventHash := GetEventNameBytes("event") // correctly hashed
	key1 := GetEventListenerKey(eventHash, 5, 1)
	key2 := GetEventListenerKey(eventHash, 2, 0)

	assert.EqualValues(t, eventHash, key1.K1(), "should have correct event hash for key1")
	assert.EqualValues(t, eventHash, key2.K1(), "should have correct event hash for key2")
	assert.EqualValues(t, eventHash, key1.K1(), "should have correct name bytes in GetEventListenerKey for key1")
	assert.EqualValues(t, eventHash, key2.K1(), "should have correct name bytes in GetEventListenerKey for key2")
	assert.EqualValues(t, int(5), key1.K2(), "should have correct order bytes in GetEventListenerKey for key1")
	assert.EqualValues(t, int(2), key2.K2(), "should have correct order bytes in GetEventListenerKey for key2")
	assert.EqualValues(t, int(1), key1.K3(), "should have correct trigger id bytes in GetEventListenerKey for key1")
	assert.EqualValues(t, int(0), key2.K3(), "should have correct trigger id bytes in GetEventListenerKey for key2")
	assert.PanicsWithValue(t, "invalid event name: ", func() { GetEventListenerKey([]byte(""), 2, 0) }, "should panic with error message when given invalid event name")
}

func TestGetEventListenerPrefix(t *testing.T) {
	eventHash := GetEventNameBytes("event")
	key := GetEventListenerPrefix(eventHash)

	assert.EqualValues(t, EventListenerKeyPrefix, key[0:1], "should receive correct prefix for GetEventListenerPrefix")
	assert.EqualValues(t, GetEventNameBytes("event"), key[1:33], "should receive correct name bytes for GetEventListenerPrefix")
}

func TestGetTriggerKey(t *testing.T) {
	key := GetTriggerKey(1)
	assert.EqualValues(t, TriggerKeyPrefix, key[0:1], "should have correct prefix for GetTriggerKey")
	assert.EqualValues(t, int(1), int(binary.BigEndian.Uint64(key[1:9])), "should have correct ID for GetTriggerKey")
}

func TestGetNextTriggerIDKey(t *testing.T) {
	assert.EqualValues(t, NextTriggerIDKey, GetNextTriggerIDKey()[0:1], "should return the correct key for GetNextTriggerIDKey")
}

func TestGetQueueKeyPrefix(t *testing.T) {
	assert.EqualValues(t, QueueKeyPrefix, GetQueueKeyPrefix()[0:1], "should return the correct prefix for GetQueueKeyPrefix")
}

func TestGetQueueKey(t *testing.T) {
	key := GetQueueKey(1)
	assert.EqualValues(t, QueueKeyPrefix, key[0:1], "should return the correct prefix for GetQueueKey")
	assert.EqualValues(t, int(1), int(binary.BigEndian.Uint64(key[1:9])), "should have the correct index bytes for GetQueueKey")
}

func TestGetQueueStartIndexKey(t *testing.T) {
	assert.EqualValues(t, QueueStartIndexKey, GetQueueStartIndexKey()[0:1], "should return the correct key for GetQueueStartIndexKey")
}

func TestGetQueueLengthKey(t *testing.T) {
	assert.EqualValues(t, QueueLengthKey, GetQueueLengthKey()[0:1], "should return the correct key for GetQueueLengthKey")
}

func TestGetQueueIndexToAndFromBytes(t *testing.T) {
	bytes := GetQueueIndexBytes(1)
	index := GetQueueIndexFromBytes(bytes)
	assert.EqualValues(t, int(1), int(index), "should correctly get queue index from key bytes")
}

func TestGetTriggerIDToAndFromBytes(t *testing.T) {
	bytes := GetTriggerIDBytes(1)
	index := GetTriggerIDFromBytes(bytes)
	assert.EqualValues(t, int(1), int(index), "should correctly get trigger id from key bytes")
}

func TestGetEventNameBytes(t *testing.T) {
	bytes := GetEventNameBytes("event")
	bytes2 := GetEventNameBytes("Event")

	hash := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace("event"))))
	expectedBytes := hash[:]

	assert.EqualValues(t, expectedBytes, bytes, "should have correct bytes for GetEventNameBytes")
	assert.EqualValues(t, expectedBytes, bytes2, "should have same bytes for capitals in GetEventNameBytes")
	assert.PanicsWithValue(t, "invalid event name: ", func() { GetEventNameBytes("") })
}
