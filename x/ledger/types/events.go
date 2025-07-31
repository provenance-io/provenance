package types

func NewEventLedgerCreated(key *LedgerKey) *EventLedgerCreated {
	return &EventLedgerCreated{
		AssetClassId: key.AssetClassId,
		NftId:        key.NftId,
	}
}

func NewEventLedgerUpdated(key *LedgerKey) *EventLedgerUpdated {
	return &EventLedgerUpdated{
		AssetClassId: key.AssetClassId,
		NftId:        key.NftId,
	}
}

func NewEventLedgerEntryAdded(key *LedgerKey, correlationID string) *EventLedgerEntryAdded {
	return &EventLedgerEntryAdded{
		AssetClassId:  key.AssetClassId,
		NftId:         key.NftId,
		CorrelationId: correlationID,
	}
}

func NewEventFundTransferWithSettlement(key *LedgerKey, correlationID string) *EventFundTransferWithSettlement {
	return &EventFundTransferWithSettlement{
		AssetClassId:  key.AssetClassId,
		NftId:         key.NftId,
		CorrelationId: correlationID,
	}
}

func NewEventLedgerDestroyed(key *LedgerKey) *EventLedgerDestroyed {
	return &EventLedgerDestroyed{
		AssetClassId: key.AssetClassId,
		NftId:        key.NftId,
	}
}
