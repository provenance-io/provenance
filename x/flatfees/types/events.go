package types

// NewEventParamsUpdated creates the event emitted when UpdateParams succeeds.
func NewEventParamsUpdated() *EventParamsUpdated {
	return &EventParamsUpdated{}
}

// NewEventConversionFactorUpdated creates the event emitted when UpdateConversionFactor succeeds.
func NewEventConversionFactorUpdated(cf ConversionFactor) *EventConversionFactorUpdated {
	return &EventConversionFactorUpdated{
		DefinitionAmount: cf.DefinitionAmount.String(),
		ConvertedAmount:  cf.ConvertedAmount.String(),
	}
}

// NewEventMsgFeeSet creates the event emitted from the SetMsgFee keeper helper when a fee is set.
func NewEventMsgFeeSet(msgTypeURL string) *EventMsgFeeSet {
	return &EventMsgFeeSet{MsgTypeUrl: msgTypeURL}
}

// NewEventMsgFeeUnset creates the event emitted from the RemoveMsgFee keeper helper when a fee is removed.
func NewEventMsgFeeUnset(msgTypeURL string) *EventMsgFeeUnset {
	return &EventMsgFeeUnset{MsgTypeUrl: msgTypeURL}
}

// NewEventOracleAddressAdded creates the event emitted when AddOracleAddress succeeds.
func NewEventOracleAddressAdded(address string) *EventOracleAddressAdded {
	return &EventOracleAddressAdded{OracleAddress: address}
}

// NewEventOracleAddressRemoved creates the event emitted when RemoveOracleAddress succeeds.
func NewEventOracleAddressRemoved(address string) *EventOracleAddressRemoved {
	return &EventOracleAddressRemoved{OracleAddress: address}
}
