package types

import (
	"encoding/base64"
	time "time"
)

const (
	// EventTypeAttributeAdded emitted when account attributes are added.
	EventTypeAttributeAdded string = "account_attribute_added"
	// EventTypeAttributeUpdated emitted when account attributes are updated.
	EventTypeAttributeUpdated string = "account_attribute_updated"
	// EventTypeAttributeDeleted emitted when account attributes are removed.
	EventTypeAttributeDeleted string = "account_attribute_deleted"
	// EventTypeAttributeDistinctDeleted emitted when a distinct account attribute is deleted.
	EventTypeAttributeDistinctDeleted string = "account_attribute_distinct_deleted"
	// EventTypeDeletedExpired emitted when attributes have expired and been deleted in begin blocker
	EventTypeDeletedExpired string = "attribute_deleted_expired"

	AttributeKeyAttribute           string = "attribute"
	AttributeKeyNameAttribute       string = "attribute_name"
	AttributeKeyAccountAddress      string = "account_address"
	AttributeKeyTotalExpiredDeleted string = "total_expired_deleted"

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
	// EventTelemetryLabelName name telemetry metrics label
	EventTelemetryLabelValue string = "value"
	// EventTelemetryLabelType type telemetry metrics label
	EventTelemetryLabelType string = "type"
	// EventTelemetryLabelOwner owner telemetry metrics label
	EventTelemetryLabelOwner string = "owner"
	// EventTelemetryKeyAccount acount telemetry metrics label
	EventTelemetryLabelAccount string = "account"
	// EventTelemetryKeyAccount size telemetry metrics label
	EventTelemetryLabelSize string = "size"
)

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
		UpdateExpiration:   updated,
	}
}

func NewEventAttributeDelete(name string, account string, owner string) *EventAttributeDelete {
	return &EventAttributeDelete{
		Name:    name,
		Owner:   owner,
		Account: account,
	}
}

func NewEventDistinctAttributeDelete(name string, value string, account string, owner string) *EventAttributeDistinctDelete {
	return &EventAttributeDistinctDelete{
		Name:    name,
		Value:   value,
		Owner:   owner,
		Account: account,
	}
}

func NewEventAttributeExpiredDelete(attribute Attribute) *EventAttributeExpiredDelete {
	var expiredTime string
	if attribute.ExpirationDate != nil {
		expiredTime = attribute.ExpirationDate.String()
	}

	return &EventAttributeExpiredDelete{
		Name:       attribute.Name,
		Value:      base64.StdEncoding.EncodeToString(attribute.GetValue()),
		Account:    attribute.Address,
		Expiration: expiredTime,
	}
}
