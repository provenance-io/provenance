package types

const (
	// EventTypeScopeCreated is the event type generated when new scopes are created.
	EventTypeScopeCreated string = "scope_created"
	// EventTypeScopeUpdated is the event type generated when existing scopes are updated.
	EventTypeScopeUpdated string = "scope_updated"
	// EventTypeScopeOwnership is the event type generated when existing scopes have a change ownership request processed.
	EventTypeScopeOwnership string = "scope_ownership"
	// EventTypeScopeRemoved is the event type generated when a scope is removed
	EventTypeScopeRemoved string = "scope_removed"

	// EventTypeSessionCreated is the event type generated when new record sessions are created.
	EventTypeSessionCreated string = "session_created"
	// EventTypeSessionUpdated is the event type generated when existing record sessions are updated.
	EventTypeSessionUpdated string = "session_updated"
	// EventTypeSessionRemoved is the event type generated when a scope is removed
	EventTypeSessionRemoved string = "session_removed"

	// EventTypeRecordCreated is the event type generated when new record sessions are created.
	EventTypeRecordCreated string = "record_created"
	// EventTypeRecordUpdated is the event type generated when existing record sessions are updated.
	EventTypeRecordUpdated string = "record_updated"
	// EventTypeRecordRemoved is the event type generated when a scope is removed
	EventTypeRecordRemoved string = "record_removed"

	// EventTypeScopeSpecificationCreated is the event type generated when a new scope specification is created.
	EventTypeScopeSpecificationCreated string = "scope_specification_created"
	// EventTypeScopeSpecificationUpdated is the event type generated when an existing scope specifications is updated.
	EventTypeScopeSpecificationUpdated string = "scope_specification_updated"
	// EventTypeScopeSpecificationRemoved is the event type generated when a scope specification is removed.
	EventTypeScopeSpecificationRemoved string = "scope_specification_removed"

	// EventTypeContractSpecificationCreated is the event type generated when a new contract specification is created.
	EventTypeContractSpecificationCreated string = "contract_specification_created"
	// EventTypeContractSpecificationUpdated is the event type generated when an existing contract specifications is updated.
	EventTypeContractSpecificationUpdated string = "contract_specification_updated"
	// EventTypeContractSpecificationRemoved is the event type generated when a contract specification is removed.
	EventTypeContractSpecificationRemoved string = "contract_specification_removed"

	// AttributeKeyScopeID is the attribute key for a scope ID attribute JSON value.
	AttributeKeyScopeID string = "scope_id"
	// AttributeKeyScope is the attribute key for a scope attribute JSON value.
	AttributeKeyScope string = "scope"
	// AttributeKeySessionID is the attribute key for a scope ID attribute JSON value.
	AttributeKeySessionID string = "session_id"
	// AttributeKeyRecordID is the attribute key for a record ID attribute JSON value.
	AttributeKeyRecordID string = "record_id"
	// AttributeKeyExecutionID is the attribute key for a scope ID attribute JSON value.
	AttributeKeyExecutionID string = "execution_id"
	// AttributeKeyModuleName is the attribute key for this module.
	AttributeKeyModuleName string = "module"
	// AttributeKeyTxHash is the attribute for the transaction hash.
	AttributeKeyTxHash = "tx_hash"
	// AttributeKeyScopeID is the attribute key for a scope ID attribute JSON value.
	AttributeKeyScopeSpecID string = "scope_spec_id"
	// AttributeKeyScope is the attribute key for a scope attribute JSON value.
	AttributeKeyScopeSpec string = "scope_spec"

	// AttributeValueCategory indicates the category for this value
	AttributeValueCategory = ModuleName
)
