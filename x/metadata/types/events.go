package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// TxEndpoint is an enum for metadata TX endpoints.
type TxEndpoint string

const (
	TxEndpoint_WriteScope  TxEndpoint = "WriteScope"
	TxEndpoint_DeleteScope TxEndpoint = "DeleteScope"

	TxEndpoint_AddScopeDataAccess    TxEndpoint = "AddScopeDataAccess"
	TxEndpoint_DeleteScopeDataAccess TxEndpoint = "DeleteScopeDataAccess"

	TxEndpoint_WriteSession TxEndpoint = "WriteSession"

	TxEndpoint_WriteRecord  TxEndpoint = "WriteRecord"
	TxEndpoint_DeleteRecord TxEndpoint = "DeleteRecord"

	TxEndpoint_WriteScopeSpecification  TxEndpoint = "WriteScopeSpecification"
	TxEndpoint_DeleteScopeSpecification TxEndpoint = "DeleteScopeSpecification"

	TxEndpoint_WriteContractSpecification  TxEndpoint = "WriteContractSpecification"
	TxEndpoint_DeleteContractSpecification TxEndpoint = "DeleteContractSpecification"

	TxEndpoint_WriteRecordSpecification  TxEndpoint = "WriteRecordSpecification"
	TxEndpoint_DeleteRecordSpecification TxEndpoint = "DeleteRecordSpecification"

	TxEndpoint_WriteP8eContractSpec   TxEndpoint = "WriteP8eContractSpec"
	TxEndpoint_P8eMemorializeContract TxEndpoint = "P8eMemorializeContract"

	TxEndpoint_BindOSLocator   TxEndpoint = "BindOSLocator"
	TxEndpoint_DeleteOSLocator TxEndpoint = "DeleteOSLocator"
	TxEndpoint_ModifyOSLocator TxEndpoint = "ModifyOSLocator"
)

func NewEventTxCompleted(endpoint TxEndpoint, signers []sdk.AccAddress) *EventTxCompleted {
	retval := &EventTxCompleted{
		Module:   ModuleName,
		Endpoint: string(endpoint),
		Signers:  make([]string, len(signers)),
	}
	for i, s := range signers {
		retval.Signers[i] = s.String()
	}
	return retval
}

func NewEventScopeCreated(scope Scope) *EventScopeCreated {
	return &EventScopeCreated{
		ScopeAddr: scope.ScopeId.String(),
		Scope:     &scope,
	}
}

func NewEventScopeUpdated(scope, scopeReplaced Scope) *EventScopeUpdated {
	return &EventScopeUpdated{
		ScopeAddr:     scope.ScopeId.String(),
		Scope:         &scope,
		ScopeReplaced: &scopeReplaced,
	}
}

func NewEventScopeRemoved(scope Scope) *EventScopeRemoved {
	return &EventScopeRemoved{
		ScopeAddr: scope.ScopeId.String(),
		Scope:     &scope,
	}
}

func NewEventSessionCreated(session Session) *EventSessionCreated {
	return &EventSessionCreated{
		SessionAddr: session.SessionId.String(),
		Session:     &session,
	}
}

func NewEventSessionUpdated(session, sessionReplaced Session) *EventSessionUpdated {
	return &EventSessionUpdated{
		SessionAddr:     session.SessionId.String(),
		Session:         &session,
		SessionReplaced: &sessionReplaced,
	}
}

func NewEventSessionRemoved(session Session) *EventSessionRemoved {
	return &EventSessionRemoved{
		SessionAddr: session.SessionId.String(),
		Session:     &session,
	}
}

func NewEventRecordCreated(record Record) *EventRecordCreated {
	addr, _ := record.SessionId.AsRecordAddress(record.Name)
	return &EventRecordCreated{
		RecordAddr: addr.String(),
		Record:     &record,
	}
}

func NewEventRecordUpdated(record, recordReplaced Record) *EventRecordUpdated {
	addr, _ := record.SessionId.AsRecordAddress(record.Name)
	return &EventRecordUpdated{
		RecordAddr:     addr.String(),
		Record:         &record,
		RecordReplaced: &recordReplaced,
	}
}

func NewEventRecordRemoved(record Record) *EventRecordRemoved {
	addr, _ := record.SessionId.AsRecordAddress(record.Name)
	return &EventRecordRemoved{
		RecordAddr: addr.String(),
		Record:     &record,
	}
}

func NewEventScopeSpecificationCreated(scopeSpecification ScopeSpecification) *EventScopeSpecificationCreated {
	return &EventScopeSpecificationCreated{
		ScopeSpecificationAddr: scopeSpecification.SpecificationId.String(),
		ScopeSpecification:     &scopeSpecification,
	}
}

func NewEventScopeSpecificationUpdated(scopeSpecification, scopeSpecificationReplaced ScopeSpecification) *EventScopeSpecificationUpdated {
	return &EventScopeSpecificationUpdated{
		ScopeSpecificationAddr:     scopeSpecification.SpecificationId.String(),
		ScopeSpecification:         &scopeSpecification,
		ScopeSpecificationReplaced: &scopeSpecificationReplaced,
	}
}

func NewEventScopeSpecificationRemoved(scopeSpecification ScopeSpecification) *EventScopeSpecificationRemoved {
	return &EventScopeSpecificationRemoved{
		ScopeSpecificationAddr: scopeSpecification.SpecificationId.String(),
		ScopeSpecification:     &scopeSpecification,
	}
}

func NewEventContractSpecificationCreated(contractSpecification ContractSpecification) *EventContractSpecificationCreated {
	return &EventContractSpecificationCreated{
		ContractSpecificationAddr: contractSpecification.SpecificationId.String(),
		ContractSpecification:     &contractSpecification,
	}
}

func NewEventContractSpecificationUpdated(contractSpecification, contractSpecificationReplaced ContractSpecification) *EventContractSpecificationUpdated {
	return &EventContractSpecificationUpdated{
		ContractSpecificationAddr:     contractSpecification.SpecificationId.String(),
		ContractSpecification:         &contractSpecification,
		ContractSpecificationReplaced: &contractSpecificationReplaced,
	}
}

func NewEventContractSpecificationRemoved(contractSpecification ContractSpecification) *EventContractSpecificationRemoved {
	return &EventContractSpecificationRemoved{
		ContractSpecificationAddr: contractSpecification.SpecificationId.String(),
		ContractSpecification:     &contractSpecification,
	}
}

func NewEventRecordSpecificationCreated(recordSpecification RecordSpecification) *EventRecordSpecificationCreated {
	return &EventRecordSpecificationCreated{
		RecordSpecificationAddr: recordSpecification.SpecificationId.String(),
		RecordSpecification:     &recordSpecification,
	}
}

func NewEventRecordSpecificationUpdated(recordSpecification, recordSpecificationReplaced RecordSpecification) *EventRecordSpecificationUpdated {
	return &EventRecordSpecificationUpdated{
		RecordSpecificationAddr:     recordSpecification.SpecificationId.String(),
		RecordSpecification:         &recordSpecification,
		RecordSpecificationReplaced: &recordSpecificationReplaced,
	}
}

func NewEventRecordSpecificationRemoved(recordSpecification RecordSpecification) *EventRecordSpecificationRemoved {
	return &EventRecordSpecificationRemoved{
		RecordSpecificationAddr: recordSpecification.SpecificationId.String(),
		RecordSpecification:     &recordSpecification,
	}
}

func NewEventOSLocatorCreated(osLocator ObjectStoreLocator) *EventOSLocatorCreated {
	return &EventOSLocatorCreated{
		Address: osLocator.Owner,
		Uri:     osLocator.LocatorUri,
	}
}

func NewEventOSLocatorUpdated(osLocator, osLocatorReplaced ObjectStoreLocator) *EventOSLocatorUpdated {
	return &EventOSLocatorUpdated{
		Address:     osLocator.Owner,
		Uri:         osLocator.LocatorUri,
		UriReplaced: osLocatorReplaced.LocatorUri,
	}
}

func NewEventOSLocatorRemoved(osLocator ObjectStoreLocator) *EventOSLocatorRemoved {
	return &EventOSLocatorRemoved{
		Address: osLocator.Owner,
		Uri:     osLocator.LocatorUri,
	}
}
