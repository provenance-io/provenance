package types

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TxEndpoint is an enum for metadata TX endpoints.
type TxEndpoint string

const (
	TxEndpoint_WriteScope            TxEndpoint = "WriteScope"
	TxEndpoint_DeleteScope           TxEndpoint = "DeleteScope"
	TxEndpoint_AddScopeDataAccess    TxEndpoint = "AddScopeDataAccess"
	TxEndpoint_DeleteScopeDataAccess TxEndpoint = "DeleteScopeDataAccess"
	TxEndpoint_AddScopeOwner         TxEndpoint = "AddScopeOwner"
	TxEndpoint_DeleteScopeOwner      TxEndpoint = "DeleteScopeOwner"
	TxEndpoint_UpdateValueOwners     TxEndpoint = "UpdateValueOwners"
	TxEndpoint_MigrateValueOwner     TxEndpoint = "MigrateValueOwner"

	TxEndpoint_WriteSession TxEndpoint = "WriteSession"

	TxEndpoint_WriteRecord  TxEndpoint = "WriteRecord"
	TxEndpoint_DeleteRecord TxEndpoint = "DeleteRecord"

	TxEndpoint_WriteScopeSpecification  TxEndpoint = "WriteScopeSpecification"
	TxEndpoint_DeleteScopeSpecification TxEndpoint = "DeleteScopeSpecification"

	TxEndpoint_WriteContractSpecification  TxEndpoint = "WriteContractSpecification"
	TxEndpoint_DeleteContractSpecification TxEndpoint = "DeleteContractSpecification"

	TxEndpoint_AddContractSpecToScopeSpec      TxEndpoint = "AddContractSpecToScopeSpec"
	TxEndpoint_DeleteContractSpecFromScopeSpec TxEndpoint = "DeleteContractSpecFromScopeSpec"

	TxEndpoint_WriteRecordSpecification  TxEndpoint = "WriteRecordSpecification"
	TxEndpoint_DeleteRecordSpecification TxEndpoint = "DeleteRecordSpecification"

	TxEndpoint_BindOSLocator   TxEndpoint = "BindOSLocator"
	TxEndpoint_DeleteOSLocator TxEndpoint = "DeleteOSLocator"
	TxEndpoint_ModifyOSLocator TxEndpoint = "ModifyOSLocator"
)

func NewEventTxCompleted(endpoint TxEndpoint, signers []string) *EventTxCompleted {
	return &EventTxCompleted{
		Module:   ModuleName,
		Endpoint: string(endpoint),
		Signers:  signers,
	}
}

func NewEventScopeCreated(scopeID MetadataAddress) *EventScopeCreated {
	return &EventScopeCreated{
		ScopeAddr: scopeID.String(),
	}
}

func NewEventScopeUpdated(scopeID MetadataAddress) *EventScopeUpdated {
	return &EventScopeUpdated{
		ScopeAddr: scopeID.String(),
	}
}

func NewEventScopeDeleted(scopeID MetadataAddress) *EventScopeDeleted {
	return &EventScopeDeleted{
		ScopeAddr: scopeID.String(),
	}
}

func NewEventSessionCreated(sessionID MetadataAddress) *EventSessionCreated {
	return &EventSessionCreated{
		SessionAddr: sessionID.String(),
		ScopeAddr:   sessionID.MustGetAsScopeAddress().String(),
	}
}

func NewEventSessionUpdated(sessionID MetadataAddress) *EventSessionUpdated {
	return &EventSessionUpdated{
		SessionAddr: sessionID.String(),
		ScopeAddr:   sessionID.MustGetAsScopeAddress().String(),
	}
}

func NewEventSessionDeleted(sessionID MetadataAddress) *EventSessionDeleted {
	return &EventSessionDeleted{
		SessionAddr: sessionID.String(),
		ScopeAddr:   sessionID.MustGetAsScopeAddress().String(),
	}
}

func NewEventRecordCreated(recordID, sessionID MetadataAddress) *EventRecordCreated {
	return &EventRecordCreated{
		RecordAddr:  recordID.String(),
		SessionAddr: sessionID.String(),
		ScopeAddr:   recordID.MustGetAsScopeAddress().String(),
	}
}

func NewEventRecordUpdated(recordID, sessionID MetadataAddress) *EventRecordUpdated {
	return &EventRecordUpdated{
		RecordAddr:  recordID.String(),
		SessionAddr: sessionID.String(),
		ScopeAddr:   recordID.MustGetAsScopeAddress().String(),
	}
}

func NewEventRecordDeleted(recordID MetadataAddress) *EventRecordDeleted {
	return &EventRecordDeleted{
		RecordAddr: recordID.String(),
		ScopeAddr:  recordID.MustGetAsScopeAddress().String(),
	}
}

func NewEventScopeSpecificationCreated(scopeSpecificationID MetadataAddress) *EventScopeSpecificationCreated {
	return &EventScopeSpecificationCreated{
		ScopeSpecificationAddr: scopeSpecificationID.String(),
	}
}

func NewEventScopeSpecificationUpdated(scopeSpecificationID MetadataAddress) *EventScopeSpecificationUpdated {
	return &EventScopeSpecificationUpdated{
		ScopeSpecificationAddr: scopeSpecificationID.String(),
	}
}

func NewEventScopeSpecificationDeleted(scopeSpecificationID MetadataAddress) *EventScopeSpecificationDeleted {
	return &EventScopeSpecificationDeleted{
		ScopeSpecificationAddr: scopeSpecificationID.String(),
	}
}

func NewEventContractSpecificationCreated(contractSpecificationID MetadataAddress) *EventContractSpecificationCreated {
	return &EventContractSpecificationCreated{
		ContractSpecificationAddr: contractSpecificationID.String(),
	}
}

func NewEventContractSpecificationUpdated(contractSpecificationID MetadataAddress) *EventContractSpecificationUpdated {
	return &EventContractSpecificationUpdated{
		ContractSpecificationAddr: contractSpecificationID.String(),
	}
}

func NewEventContractSpecificationDeleted(contractSpecificationID MetadataAddress) *EventContractSpecificationDeleted {
	return &EventContractSpecificationDeleted{
		ContractSpecificationAddr: contractSpecificationID.String(),
	}
}

func NewEventRecordSpecificationCreated(recordSpecificationID MetadataAddress) *EventRecordSpecificationCreated {
	return &EventRecordSpecificationCreated{
		RecordSpecificationAddr:   recordSpecificationID.String(),
		ContractSpecificationAddr: recordSpecificationID.MustGetAsContractSpecAddress().String(),
	}
}

func NewEventRecordSpecificationUpdated(recordSpecificationID MetadataAddress) *EventRecordSpecificationUpdated {
	return &EventRecordSpecificationUpdated{
		RecordSpecificationAddr:   recordSpecificationID.String(),
		ContractSpecificationAddr: recordSpecificationID.MustGetAsContractSpecAddress().String(),
	}
}

func NewEventRecordSpecificationDeleted(recordSpecificationID MetadataAddress) *EventRecordSpecificationDeleted {
	return &EventRecordSpecificationDeleted{
		RecordSpecificationAddr:   recordSpecificationID.String(),
		ContractSpecificationAddr: recordSpecificationID.MustGetAsContractSpecAddress().String(),
	}
}

func NewEventOSLocatorCreated(owner string) *EventOSLocatorCreated {
	return &EventOSLocatorCreated{
		Owner: owner,
	}
}

func NewEventOSLocatorUpdated(owner string) *EventOSLocatorUpdated {
	return &EventOSLocatorUpdated{
		Owner: owner,
	}
}

func NewEventOSLocatorDeleted(owner string) *EventOSLocatorDeleted {
	return &EventOSLocatorDeleted{
		Owner: owner,
	}
}

// NewEventSetNetAssetValue returns a new instance of EventSetNetAssetValue
func NewEventSetNetAssetValue(scopeID MetadataAddress, price sdk.Coin, volume uint64, source string) *EventSetNetAssetValue {
	return &EventSetNetAssetValue{
		ScopeId: scopeID.String(),
		Price:   price.String(),
		Source:  source,
		Volume:  strconv.FormatUint(volume, 10),
	}
}
