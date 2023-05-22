package types

import (
	"crypto/sha256"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEventListenerKey(t *testing.T) {
	key1 := GetEventListenerKey("event", 1)
	key2 := GetEventListenerKey("event", 0)

	assert.EqualValues(t, EventListenerKeyPrefix, key1[0:1])
	assert.EqualValues(t, EventListenerKeyPrefix, key2[0:1])
	assert.EqualValues(t, GetEventNameBytes("event"), key1[1:33])
	assert.EqualValues(t, GetEventNameBytes("event"), key2[1:33])
	assert.EqualValues(t, uint64(1), uint64(binary.BigEndian.Uint64(key1[33:41])))
	assert.EqualValues(t, uint64(0), uint64(binary.BigEndian.Uint64(key2[33:41])))
	assert.Panics(t, func() { GetEventListenerKey("", 0) })
}

func TestGetEventListenerPrefix(t *testing.T) {
	key := GetEventListenerPrefix("event")

	assert.EqualValues(t, EventListenerKeyPrefix, key[0:1])
	assert.EqualValues(t, GetEventNameBytes("event"), key[1:33])
}

func TestGetTriggerKey(t *testing.T) {
	key := GetTriggerKey(1)
	assert.EqualValues(t, TriggerKeyPrefix, key[0:1])
	assert.EqualValues(t, uint64(1), uint64(binary.BigEndian.Uint64(key[1:9])))
}

func TestGetNextTriggerIDKey(t *testing.T) {
	assert.EqualValues(t, NextTriggerIDKey, GetNextTriggerIDKey()[0:1])
}

func TestGetQueueKeyPrefix(t *testing.T) {
	assert.EqualValues(t, QueueKeyPrefix, GetQueueKeyPrefix()[0:1])
}

func TestGetQueueKey(t *testing.T) {
	key := GetQueueKey(1)
	assert.EqualValues(t, QueueKeyPrefix, key[0:1])
	assert.EqualValues(t, uint64(1), uint64(binary.BigEndian.Uint64(key[1:9])))
}

func TestGetQueueStartIndexKey(t *testing.T) {
	assert.EqualValues(t, QueueStartIndexKey, GetQueueStartIndexKey()[0:1])
}

func TestGetQueueLengthKey(t *testing.T) {
	assert.EqualValues(t, QueueLengthKey, GetQueueLengthKey()[0:1])
}

func TestGetQueueIndexToAndFromBytes(t *testing.T) {
	bytes := GetQueueIndexBytes(1)
	index := GetQueueIndexFromBytes(bytes)
	assert.EqualValues(t, uint64(1), index)
}

func TestGetTriggerIDToAndFromBytes(t *testing.T) {
	bytes := GetTriggerIDBytes(1)
	index := GetTriggerIDFromBytes(bytes)
	assert.EqualValues(t, uint64(1), index)
}

func TestGetGasLimitKey(t *testing.T) {
	key := GetGasLimitKey(1)
	assert.EqualValues(t, GasLimitKeyPrefix, key[0:1])
	assert.EqualValues(t, uint64(1), uint64(binary.BigEndian.Uint64(key[1:9])))
}

func TestGetGasLimitToAndFromBytes(t *testing.T) {
	bytes := GetGasLimitBytes(1)
	index := GetGasLimitFromBytes(bytes)
	assert.EqualValues(t, uint64(1), index)
}

func TestGetEventNameBytes(t *testing.T) {
	bytes := GetEventNameBytes("event")
	bytes2 := GetEventNameBytes("Event")

	hash := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace("event"))))
	expectedBytes := hash[:]

	assert.EqualValues(t, expectedBytes, bytes)
	assert.EqualValues(t, expectedBytes, bytes2)
	assert.Panics(t, func() { GetEventNameBytes("") })
}
