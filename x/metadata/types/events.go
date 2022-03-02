package types

import (
	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TelemetryCategory is an enum for metadata telemetry categories.
type TelemetryCategory string

// TelemetryObjectType is an enum for metadata telemetry object types.
type TelemetryObjectType string

// TelemetryAction is an enum for metadata telemetry actions.
type TelemetryAction string

const (
	// TKObject is the telemetry key for an object stored on-chain in the metadata module.
	TKObject string = "stored-object"
	// TKObjectAction is the telemetry key for an action taken on the chain.
	TKObjectAction string = "object-action"

	// TLCategory is a string name for labels defining an object category.
	TLCategory               string            = "category"
	TLCategory_Entry         TelemetryCategory = "entry"
	TLCategory_Specification TelemetryCategory = "specification"
	TLCategory_OSLocator     TelemetryCategory = "object-store-locator"

	// TLType is a string name for labels defining an object type.
	TLType              string              = "object-type"
	TLType_Scope        TelemetryObjectType = "scope"
	TLType_Session      TelemetryObjectType = "session"
	TLType_Record       TelemetryObjectType = "record"
	TLType_ScopeSpec    TelemetryObjectType = "scope-specification"
	TLType_ContractSpec TelemetryObjectType = "contract-specification"
	TLType_RecordSpec   TelemetryObjectType = "record-specification"
	TLType_OSLocator    TelemetryObjectType = "object-store-locator"

	// TLAction is a string name for labels defining an action taken.
	TLAction         string          = "action"
	TLAction_Created TelemetryAction = "created"
	TLAction_Updated TelemetryAction = "updated"
	TLAction_Deleted TelemetryAction = "deleted"
)

// AsLabel Returns this TelemetryCategory as a label.
func (t TelemetryCategory) AsLabel() metrics.Label {
	return telemetry.NewLabel(TLCategory, string(t))
}

// AsLabel Returns this TelemetryObjectType as a label.
func (t TelemetryObjectType) AsLabel() metrics.Label {
	return telemetry.NewLabel(TLType, string(t))
}

// AsLabel Returns this TelemetryAction as a label.
func (t TelemetryAction) AsLabel() metrics.Label {
	return telemetry.NewLabel(TLAction, string(t))
}

// GetIncObjFunc creates a function that will call telemetry.IncrCounterWithLabels for counting metadata module chain objects.
func GetIncObjFunc(objType TelemetryObjectType, action TelemetryAction) func() {
	val := 0 // Default is for action == TLAction_Updated
	if action == TLAction_Created {
		val = 1
	} else if action == TLAction_Deleted {
		val = -1
	}
	cat := TLCategory_OSLocator // Default is for objType == TLType_OSLocator
	if objType == TLType_Record || objType == TLType_Session || objType == TLType_Scope {
		cat = TLCategory_Entry
	} else if objType == TLType_RecordSpec || objType == TLType_ContractSpec || objType == TLType_ScopeSpec {
		cat = TLCategory_Specification
	}
	return func() {
		if val != 0 {
			telemetry.IncrCounterWithLabels(
				[]string{ModuleName, TKObject},
				float32(val),
				[]metrics.Label{cat.AsLabel(), objType.AsLabel()},
			)
		}
		telemetry.IncrCounterWithLabels(
			[]string{ModuleName, TKObjectAction},
			1,
			[]metrics.Label{cat.AsLabel(), objType.AsLabel(), action.AsLabel()},
		)
	}
}

// TxEndpoint is an enum for metadata TX endpoints.
type TxEndpoint string

const (
	TxEndpoint_WriteScope            TxEndpoint = "WriteScope"
	TxEndpoint_DeleteScope           TxEndpoint = "DeleteScope"
	TxEndpoint_AddScopeDataAccess    TxEndpoint = "AddScopeDataAccess"
	TxEndpoint_DeleteScopeDataAccess TxEndpoint = "DeleteScopeDataAccess"
	TxEndpoint_AddScopeOwner         TxEndpoint = "AddScopeOwner"
	TxEndpoint_DeleteScopeOwner      TxEndpoint = "DeleteScopeOwner"

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
