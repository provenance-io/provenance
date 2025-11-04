package types

import "strconv"

func NewEventLedgerCreated(key *LedgerKey) *EventLedgerCreated {
	return &EventLedgerCreated{
		AssetClassId: key.AssetClassId,
		NftId:        key.NftId,
	}
}

func NewEventLedgerUpdated(key *LedgerKey, updateType UpdateType) *EventLedgerUpdated {
	return &EventLedgerUpdated{
		AssetClassId: key.AssetClassId,
		NftId:        key.NftId,
		UpdateType:   updateType,
	}
}

func NewEventLedgerEntryAdded(key *LedgerKey, correlationID string) *EventLedgerEntryAdded {
	return &EventLedgerEntryAdded{
		AssetClassId:  key.AssetClassId,
		NftId:         key.NftId,
		CorrelationId: correlationID,
	}
}

func NewEventLedgerEntryUpdated(key *LedgerKey, correlationID string) *EventLedgerEntryUpdated {
	return &EventLedgerEntryUpdated{
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

func NewEventLedgerClassCreated(lc *LedgerClass) *EventLedgerClassCreated {
	return &EventLedgerClassCreated{
		LedgerClassId: lc.LedgerClassId,
		AssetClassId:  lc.AssetClassId,
		Denom:         lc.Denom,
	}
}

func NewEventLedgerClassStatusTypeCreated(lcID string, t *LedgerClassStatusType) *EventLedgerClassTypeCreated {
	return &EventLedgerClassTypeCreated{
		LedgerClassId: lcID,
		TypeCreated:   CLASS_TYPE_CREATED_STATUS,
		Id:            strconv.Itoa(int(t.Id)),
		Code:          t.Code,
	}
}

func NewEventLedgerClassEntryTypeCreated(lcID string, t *LedgerClassEntryType) *EventLedgerClassTypeCreated {
	return &EventLedgerClassTypeCreated{
		LedgerClassId: lcID,
		TypeCreated:   CLASS_TYPE_CREATED_ENTRY,
		Id:            strconv.Itoa(int(t.Id)),
		Code:          t.Code,
	}
}

func NewEventLedgerClassBucketTypeCreated(lcID string, t *LedgerClassBucketType) *EventLedgerClassTypeCreated {
	return &EventLedgerClassTypeCreated{
		LedgerClassId: lcID,
		TypeCreated:   CLASS_TYPE_CREATED_BUCKET,
		Id:            strconv.Itoa(int(t.Id)),
		Code:          t.Code,
	}
}