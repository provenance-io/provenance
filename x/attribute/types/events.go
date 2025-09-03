package types

import (
	"encoding/base64"
	"strconv"
	time "time"
)

const (
	// EventTypeAttributeAdded emitted when account attributes are added.
	EventTypeAttributeAdded string = "account_attribute_added"
	// EventTypeAttributeUpdated emitted when account attributes are updated.
	EventTypeAttributeUpdated string = "account_attribute_updated"
	// EventTypeAttributeExpirationUpdated emitted when a attribute's expiration date is updated.
	EventTypeAttributeExpirationUpdated string = "account_attribute_expiration_updated"
	// EventTypeAttributeDeleted emitted when account attributes are removed.
	EventTypeAttributeDeleted string = "account_attribute_deleted"
	// EventTypeAttributeDistinctDeleted emitted when a distinct account attribute is deleted.
	EventTypeAttributeDistinctDeleted string = "account_attribute_distinct_deleted"
	// EventTypeDeletedExpired emitted when attributes have expired and been deleted in begin blocker
	EventTypeDeletedExpired string = "attribute_deleted_expired"
	// AttributeKeyAttribute is the key used for attributes.
	AttributeKeyAttribute string = "attribute"
	// AttributeKeyNameAttribute is the telemetry label for attribute value.
	AttributeKeyNameAttribute string = "attribute_name"
	// AttributeKeyAccountAddress is the telemetry label for account.
	AttributeKeyAccountAddress string = "account_address"
	// AttributeKeyTotalExpired is the telemetry label for size.
	AttributeKeyTotalExpired string = "total_expired_deleted"

	// EventTelemetryKeyAdd add telemetry metrics key
	EventTelemetryKeyAdd string = "add"
	// EventTelemetryKeyUpdate add telemetry metrics key
	EventTelemetryKeyUpdate string = "update"
	// EventTelemetryKeyDelete delete telemetry metrics key
	EventTelemetryKeyDelete string = "delete"
	// EventTelemetryKeyDistinctDelete delete telemetry metrics key
	EventTelemetryKeyDistinctDelete string = "distinct_delete"
	// EventTelemetryLabelName name telemetry metrics label
	EventTelemetryLabelName string = "name"
	// EventTelemetryLabelValue name telemetry metrics label
	EventTelemetryLabelValue string = "value"
	// EventTelemetryLabelType type telemetry metrics label
	EventTelemetryLabelType string = "type"
	// EventTelemetryLabelOwner owner telemetry metrics label
	EventTelemetryLabelOwner string = "owner"
	// EventTelemetryLabelAccount acount telemetry metrics label
	EventTelemetryLabelAccount string = "account"
	// EventTelemetryLabelSize size telemetry metrics label
	EventTelemetryLabelSize string = "size"
)

// NewEventAttributeAdd creates a new event for adding an attribute.
func NewEventAttributeAdd(attribute Attribute, owner string) *EventAttributeAdd {
	var expirationDate string
	if attribute.ExpirationDate != nil {
		expirationDate = attribute.ExpirationDate.String()
	}
	return &EventAttributeAdd{
		Name:       attribute.Name,
		Value:      base64.StdEncoding.EncodeToString(attribute.GetValue()),
		Type:       attribute.AttributeType.String(),
		Account:    attribute.Address,
		Owner:      owner,
		Expiration: expirationDate,
	}
}

// NewEventAttributeUpdate creates a new event for updating an attribute.
func NewEventAttributeUpdate(originalAttribute Attribute, updateAttribute Attribute, owner string) *EventAttributeUpdate {
	return &EventAttributeUpdate{
		Name:          originalAttribute.Name,
		OriginalValue: base64.StdEncoding.EncodeToString(originalAttribute.GetValue()),
		OriginalType:  originalAttribute.AttributeType.String(),
		UpdateValue:   base64.StdEncoding.EncodeToString(updateAttribute.GetValue()),
		UpdateType:    updateAttribute.AttributeType.String(),
		Account:       originalAttribute.Address,
		Owner:         owner,
	}
}

// NewEventAttributeExpirationUpdate creates a new event for attribute expiration updates.
func NewEventAttributeExpirationUpdate(attribute Attribute, originalExpiration *time.Time, owner string) *EventAttributeExpirationUpdate {
	var original, updated string
	if attribute.ExpirationDate != nil {
		updated = attribute.ExpirationDate.String()
	}
	if originalExpiration != nil {
		original = originalExpiration.String()
	}
	return &EventAttributeExpirationUpdate{
		Name:               attribute.Name,
		Value:              base64.StdEncoding.EncodeToString(attribute.GetValue()),
		Account:            attribute.Address,
		Owner:              owner,
		OriginalExpiration: original,
		UpdatedExpiration:  updated,
	}
}

// NewEventAttributeDelete creates a new event for deleting an attribute.
func NewEventAttributeDelete(name string, account string, owner string) *EventAttributeDelete {
	return &EventAttributeDelete{
		Name:    name,
		Owner:   owner,
		Account: account,
	}
}

// NewEventDistinctAttributeDelete creates a new event for deleting a distinct attribute.
func NewEventDistinctAttributeDelete(name string, value string, account string, owner string) *EventAttributeDistinctDelete {
	return &EventAttributeDistinctDelete{
		Name:    name,
		Value:   value,
		Owner:   owner,
		Account: account,
	}
}

// NewEventAttributeExpired creates a new event signaling an attribute has expired.
func NewEventAttributeExpired(attribute Attribute) *EventAttributeExpired {
	var expiredTime string
	if attribute.ExpirationDate != nil {
		expiredTime = attribute.ExpirationDate.String()
	}

	return &EventAttributeExpired{
		Name:       attribute.Name,
		ValueHash:  string(attribute.Hash()),
		Account:    attribute.Address,
		Expiration: expiredTime,
	}
}

// NewEventAttributeParamsUpdated creates a new event for updated attribute params.
func NewEventAttributeParamsUpdated(params Params) *EventAttributeParamsUpdated {
	return &EventAttributeParamsUpdated{MaxValueLength: strconv.FormatUint(uint64(params.MaxValueLength), 10)}
}
