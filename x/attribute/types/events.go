package types

const (
	// The type of event generated when account attributes are added.
	EventTypeAttributeAdded string = "account_attribute_added"
	// The type of event generated when account attributes are removed.
	EventTypeAttributeDeleted string = "account_attribute_deleted"

	AttributeKeyAttribute      string = "attribute"
	AttributeKeyNameAttribute  string = "attribute_name"
	AttributeKeyAccountAddress string = "account_address"
)

func NewEventAttributeAdd(attribute Attribute, owner string) *EventAttributeAdd {
	return &EventAttributeAdd{
		Name:    attribute.Name,
		Value:   string(attribute.GetValue()),
		Type:    attribute.AttributeType.String(),
		Account: attribute.Address,
		Owner:   owner,
	}
}

func NewEventAttributeDelete(name string, account string, owner string) *EventAttributeDelete {
	return &EventAttributeDelete{
		Name:    name,
		Owner:   owner,
		Account: account,
	}
}
