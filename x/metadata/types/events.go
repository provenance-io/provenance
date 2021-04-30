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
	}
}

func NewEventScopeUpdated(scope Scope) *EventScopeUpdated {
	return &EventScopeUpdated{
		ScopeAddr: scope.ScopeId.String(),
	}
}

func NewEventScopeDeleted(scope Scope) *EventScopeDeleted {
	return &EventScopeDeleted{
		ScopeAddr: scope.ScopeId.String(),
	}
}

func NewEventSessionCreated(session Session) *EventSessionCreated {
	return &EventSessionCreated{
		SessionAddr: session.SessionId.String(),
		ScopeAddr:   session.SessionId.MustGetAsScopeAddress().String(),
	}
}

func NewEventSessionUpdated(session Session) *EventSessionUpdated {
	return &EventSessionUpdated{
		SessionAddr: session.SessionId.String(),
		ScopeAddr:   session.SessionId.MustGetAsScopeAddress().String(),
	}
}

func NewEventSessionDeleted(session Session) *EventSessionDeleted {
	return &EventSessionDeleted{
		SessionAddr: session.SessionId.String(),
		ScopeAddr:   session.SessionId.MustGetAsScopeAddress().String(),
	}
}

func NewEventRecordCreated(record Record) *EventRecordCreated {
	return &EventRecordCreated{
		RecordAddr:  record.SessionId.MustGetAsRecordAddress(record.Name).String(),
		SessionAddr: record.SessionId.String(),
		ScopeAddr:   record.SessionId.MustGetAsScopeAddress().String(),
	}
}

func NewEventRecordUpdated(record Record) *EventRecordUpdated {
	return &EventRecordUpdated{
		RecordAddr:  record.SessionId.MustGetAsRecordAddress(record.Name).String(),
		SessionAddr: record.SessionId.String(),
		ScopeAddr:   record.SessionId.MustGetAsScopeAddress().String(),
	}
}

func NewEventRecordDeleted(record Record) *EventRecordDeleted {
	return &EventRecordDeleted{
		RecordAddr:  record.SessionId.MustGetAsRecordAddress(record.Name).String(),
		SessionAddr: record.SessionId.String(),
		ScopeAddr:   record.SessionId.MustGetAsScopeAddress().String(),
	}
}

func NewEventScopeSpecificationCreated(scopeSpecification ScopeSpecification) *EventScopeSpecificationCreated {
	return &EventScopeSpecificationCreated{
		ScopeSpecificationAddr: scopeSpecification.SpecificationId.String(),
	}
}

func NewEventScopeSpecificationUpdated(scopeSpecification ScopeSpecification) *EventScopeSpecificationUpdated {
	return &EventScopeSpecificationUpdated{
		ScopeSpecificationAddr: scopeSpecification.SpecificationId.String(),
	}
}

func NewEventScopeSpecificationDeleted(scopeSpecification ScopeSpecification) *EventScopeSpecificationDeleted {
	return &EventScopeSpecificationDeleted{
		ScopeSpecificationAddr: scopeSpecification.SpecificationId.String(),
	}
}

func NewEventContractSpecificationCreated(contractSpecification ContractSpecification) *EventContractSpecificationCreated {
	return &EventContractSpecificationCreated{
		ContractSpecificationAddr: contractSpecification.SpecificationId.String(),
	}
}

func NewEventContractSpecificationUpdated(contractSpecification ContractSpecification) *EventContractSpecificationUpdated {
	return &EventContractSpecificationUpdated{
		ContractSpecificationAddr: contractSpecification.SpecificationId.String(),
	}
}

func NewEventContractSpecificationDeleted(contractSpecification ContractSpecification) *EventContractSpecificationDeleted {
	return &EventContractSpecificationDeleted{
		ContractSpecificationAddr: contractSpecification.SpecificationId.String(),
	}
}

func NewEventRecordSpecificationCreated(recordSpecification RecordSpecification) *EventRecordSpecificationCreated {
	return &EventRecordSpecificationCreated{
		RecordSpecificationAddr:   recordSpecification.SpecificationId.String(),
		ContractSpecificationAddr: recordSpecification.SpecificationId.MustGetAsContractSpecAddress().String(),
	}
}

func NewEventRecordSpecificationUpdated(recordSpecification RecordSpecification) *EventRecordSpecificationUpdated {
	return &EventRecordSpecificationUpdated{
		RecordSpecificationAddr:   recordSpecification.SpecificationId.String(),
		ContractSpecificationAddr: recordSpecification.SpecificationId.MustGetAsContractSpecAddress().String(),
	}
}

func NewEventRecordSpecificationDeleted(recordSpecification RecordSpecification) *EventRecordSpecificationDeleted {
	return &EventRecordSpecificationDeleted{
		RecordSpecificationAddr:   recordSpecification.SpecificationId.String(),
		ContractSpecificationAddr: recordSpecification.SpecificationId.MustGetAsContractSpecAddress().String(),
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

func NewEventOSLocatorDeleted(osLocator ObjectStoreLocator) *EventOSLocatorDeleted {
	return &EventOSLocatorDeleted{
		Address: osLocator.Owner,
		Uri:     osLocator.LocatorUri,
	}
}
