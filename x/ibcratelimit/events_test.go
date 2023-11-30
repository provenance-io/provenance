package ibcratelimit_test

import (
	"testing"

	"github.com/provenance-io/provenance/x/ibcratelimit"
	"github.com/stretchr/testify/assert"
)

func TestNewEventAckRevertFailure(t *testing.T) {
	expected := &ibcratelimit.EventAckRevertFailure{
		Module: "module",
		Packet: "packet",
		Ack:    "ack",
	}
	event := ibcratelimit.NewEventAckRevertFailure(expected.Module, expected.Packet, expected.Ack)
	assert.Equal(t, expected, event, "should create the correct event type")
}

func TestNewEventTimeoutRevertFailure(t *testing.T) {
	expected := &ibcratelimit.EventTimeoutRevertFailure{
		Module: "module",
		Packet: "packet",
	}
	event := ibcratelimit.NewEventTimeoutRevertFailure(expected.Module, expected.Packet)
	assert.Equal(t, expected, event, "should create the correct event type")
}

func TestNewEventParamsUpdated(t *testing.T) {
	expected := &ibcratelimit.EventParamsUpdated{}
	event := ibcratelimit.NewEventParamsUpdated()
	assert.Equal(t, expected, event, "should create the correct event type")
}
