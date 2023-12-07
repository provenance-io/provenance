package ibcratelimit

// NewEventAckRevertFailure returns a new EventAckRevertFailure.
func NewEventAckRevertFailure(module, packet, ack string) *EventAckRevertFailure {
	return &EventAckRevertFailure{
		Module: module,
		Packet: packet,
		Ack:    ack,
	}
}

// NewEventTimeoutRevertFailure returns a new EventTimeoutRevertFailure.
func NewEventTimeoutRevertFailure(module, packet string) *EventTimeoutRevertFailure {
	return &EventTimeoutRevertFailure{
		Module: module,
		Packet: packet,
	}
}

// NewEventParamsUpdated returns a new EventParamsUpdated.
func NewEventParamsUpdated() *EventParamsUpdated {
	return &EventParamsUpdated{}
}
