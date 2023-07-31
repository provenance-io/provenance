package types

// IBC events
const (
	EventTypeQueryResult = "query_result"
	EventTypeTimeout     = "timeout"
	// this line is used by starport scaffolding # ibc/packet/event

	AttributeKeyAckSuccess = "success"
	AttributeKeyAck        = "acknowledgement"
	AttributeKeyAckError   = "error"
	AttributeKeySequence   = "sequence"
)
