package types

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TxEndpoint is an enum for metadata TX endpoints.
type TxEndpoint string

const (
	TxEndpoint_WriteScope            TxEndpoint = "WriteScope"            //nolint:revive
	TxEndpoint_DeleteScope           TxEndpoint = "DeleteScope"           //nolint:revive
	TxEndpoint_AddScopeDataAccess    TxEndpoint = "AddScopeDataAccess"    //nolint:revive
	TxEndpoint_DeleteScopeDataAccess TxEndpoint = "DeleteScopeDataAccess" //nolint:revive
	TxEndpoint_AddScopeOwner         TxEndpoint = "AddScopeOwner"         //nolint:revive
	TxEndpoint_DeleteScopeOwner      TxEndpoint = "DeleteScopeOwner"      //nolint:revive
	TxEndpoint_UpdateValueOwners     TxEndpoint = "UpdateValueOwners"     //nolint:revive
	TxEndpoint_MigrateValueOwner     TxEndpoint = "MigrateValueOwner"     //nolint:revive

	TxEndpoint_WriteSession TxEndpoint = "WriteSession" //nolint:revive

	TxEndpoint_WriteRecord  TxEndpoint = "WriteRecord"  //nolint:revive
	TxEndpoint_DeleteRecord TxEndpoint = "DeleteRecord" //nolint:revive

	TxEndpoint_WriteScopeSpecification  TxEndpoint = "WriteScopeSpecification"  //nolint:revive
	TxEndpoint_DeleteScopeSpecification TxEndpoint = "DeleteScopeSpecification" //nolint:revive

	TxEndpoint_WriteContractSpecification  TxEndpoint = "WriteContractSpecification"  //nolint:revive
	TxEndpoint_DeleteContractSpecification TxEndpoint = "DeleteContractSpecification" //nolint:revive

	TxEndpoint_AddContractSpecToScopeSpec      TxEndpoint = "AddContractSpecToScopeSpec"      //nolint:revive
	TxEndpoint_DeleteContractSpecFromScopeSpec TxEndpoint = "DeleteContractSpecFromScopeSpec" //nolint:revive

	TxEndpoint_WriteRecordSpecification  TxEndpoint = "WriteRecordSpecification"  //nolint:revive
	TxEndpoint_DeleteRecordSpecification TxEndpoint = "DeleteRecordSpecification" //nolint:revive

	TxEndpoint_BindOSLocator   TxEndpoint = "BindOSLocator"   //nolint:revive
	TxEndpoint_DeleteOSLocator TxEndpoint = "DeleteOSLocator" //nolint:revive
	TxEndpoint_ModifyOSLocator TxEndpoint = "ModifyOSLocator" //nolint:revive
)

// NewEventTxCompleted creates a new event indicating transaction completion.
func NewEventTxCompleted(endpoint TxEndpoint, signers []string) *EventTxCompleted {
	return &EventTxCompleted{
		Module:   ModuleName,
		Endpoint: string(endpoint),
		Signers:  signers,
	}
}

// NewEventScopeCreated creates a new event for a created scope.
func NewEventScopeCreated(scopeID MetadataAddress) *EventScopeCreated {
	return &EventScopeCreated{
		ScopeAddr: scopeID.String(),
	}
}

// NewEventScopeUpdated creates a new event for an updated scope.
func NewEventScopeUpdated(scopeID MetadataAddress) *EventScopeUpdated {
	return &EventScopeUpdated{
		ScopeAddr: scopeID.String(),
	}
}

// NewEventScopeDeleted creates a new event for a deleted scope.
func NewEventScopeDeleted(scopeID MetadataAddress) *EventScopeDeleted {
	return &EventScopeDeleted{
		ScopeAddr: scopeID.String(),
	}
}

// NewEventSessionCreated creates a new event for a created session.
func NewEventSessionCreated(sessionID MetadataAddress) *EventSessionCreated {
	return &EventSessionCreated{
		SessionAddr: sessionID.String(),
		ScopeAddr:   sessionID.MustGetAsScopeAddress().String(),
	}
}

// NewEventSessionUpdated creates a new event for an updated session.
func NewEventSessionUpdated(sessionID MetadataAddress) *EventSessionUpdated {
	return &EventSessionUpdated{
		SessionAddr: sessionID.String(),
		ScopeAddr:   sessionID.MustGetAsScopeAddress().String(),
	}
}

// NewEventSessionDeleted creates a new event for a deleted session.
func NewEventSessionDeleted(sessionID MetadataAddress) *EventSessionDeleted {
	return &EventSessionDeleted{
		SessionAddr: sessionID.String(),
		ScopeAddr:   sessionID.MustGetAsScopeAddress().String(),
	}
}

// NewEventRecordCreated creates a new event for a created record.
func NewEventRecordCreated(recordID, sessionID MetadataAddress) *EventRecordCreated {
	return &EventRecordCreated{
		RecordAddr:  recordID.String(),
		SessionAddr: sessionID.String(),
		ScopeAddr:   recordID.MustGetAsScopeAddress().String(),
	}
}

// NewEventRecordUpdated creates a new event for an updated record.
func NewEventRecordUpdated(recordID, sessionID MetadataAddress) *EventRecordUpdated {
	return &EventRecordUpdated{
		RecordAddr:  recordID.String(),
		SessionAddr: sessionID.String(),
		ScopeAddr:   recordID.MustGetAsScopeAddress().String(),
	}
}

// NewEventRecordDeleted creates a new event for a deleted record.
func NewEventRecordDeleted(recordID MetadataAddress) *EventRecordDeleted {
	return &EventRecordDeleted{
		RecordAddr: recordID.String(),
		ScopeAddr:  recordID.MustGetAsScopeAddress().String(),
	}
}

// NewEventScopeSpecificationCreated creates a new event for a created scope specification.
func NewEventScopeSpecificationCreated(scopeSpecificationID MetadataAddress) *EventScopeSpecificationCreated {
	return &EventScopeSpecificationCreated{
		ScopeSpecificationAddr: scopeSpecificationID.String(),
	}
}

// NewEventScopeSpecificationUpdated creates a new event for an updated scope specification.
func NewEventScopeSpecificationUpdated(scopeSpecificationID MetadataAddress) *EventScopeSpecificationUpdated {
	return &EventScopeSpecificationUpdated{
		ScopeSpecificationAddr: scopeSpecificationID.String(),
	}
}

// NewEventScopeSpecificationDeleted creates a new event for a deleted scope specification.
func NewEventScopeSpecificationDeleted(scopeSpecificationID MetadataAddress) *EventScopeSpecificationDeleted {
	return &EventScopeSpecificationDeleted{
		ScopeSpecificationAddr: scopeSpecificationID.String(),
	}
}

// NewEventContractSpecificationCreated creates a new event for a created contract specification.
func NewEventContractSpecificationCreated(contractSpecificationID MetadataAddress) *EventContractSpecificationCreated {
	return &EventContractSpecificationCreated{
		ContractSpecificationAddr: contractSpecificationID.String(),
	}
}

// NewEventContractSpecificationUpdated creates a new event for an updated contract specification.
func NewEventContractSpecificationUpdated(contractSpecificationID MetadataAddress) *EventContractSpecificationUpdated {
	return &EventContractSpecificationUpdated{
		ContractSpecificationAddr: contractSpecificationID.String(),
	}
}

// NewEventContractSpecificationDeleted creates a new event for a deleted contract specification.
func NewEventContractSpecificationDeleted(contractSpecificationID MetadataAddress) *EventContractSpecificationDeleted {
	return &EventContractSpecificationDeleted{
		ContractSpecificationAddr: contractSpecificationID.String(),
	}
}

// NewEventRecordSpecificationCreated creates a new event for a created record specification.
func NewEventRecordSpecificationCreated(recordSpecificationID MetadataAddress) *EventRecordSpecificationCreated {
	return &EventRecordSpecificationCreated{
		RecordSpecificationAddr:   recordSpecificationID.String(),
		ContractSpecificationAddr: recordSpecificationID.MustGetAsContractSpecAddress().String(),
	}
}

// NewEventRecordSpecificationUpdated creates a new event for an updated record specification.
func NewEventRecordSpecificationUpdated(recordSpecificationID MetadataAddress) *EventRecordSpecificationUpdated {
	return &EventRecordSpecificationUpdated{
		RecordSpecificationAddr:   recordSpecificationID.String(),
		ContractSpecificationAddr: recordSpecificationID.MustGetAsContractSpecAddress().String(),
	}
}

// NewEventRecordSpecificationDeleted creates a new event for a deleted record specification.
func NewEventRecordSpecificationDeleted(recordSpecificationID MetadataAddress) *EventRecordSpecificationDeleted {
	return &EventRecordSpecificationDeleted{
		RecordSpecificationAddr:   recordSpecificationID.String(),
		ContractSpecificationAddr: recordSpecificationID.MustGetAsContractSpecAddress().String(),
	}
}

// NewEventOSLocatorCreated creates a new event for a created object store locator.
func NewEventOSLocatorCreated(owner string) *EventOSLocatorCreated {
	return &EventOSLocatorCreated{
		Owner: owner,
	}
}

// NewEventOSLocatorUpdated creates a new event for an updated object store locator.
func NewEventOSLocatorUpdated(owner string) *EventOSLocatorUpdated {
	return &EventOSLocatorUpdated{
		Owner: owner,
	}
}

// NewEventOSLocatorDeleted creates a new event for a deleted object store locator.
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
