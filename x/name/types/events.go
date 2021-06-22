package types

const (
	// EventTypeNameBound is the type of event generated when a name is bound to an address.
	EventTypeNameBound string = "name_bound"
	// EventTypeNameUnbound is the type of event generated when a name is unbound from an address (deleted).
	EventTypeNameUnbound string = "name_unbound"

	// KeyAttributeName is the key for a name.
	KeyAttributeName string = "name"
	// KeyAttributeAddress is the key for an address.
	KeyAttributeAddress string = "address"
)

func NewEventNameBound(address string, name string) *EventNameBound {
	return &EventNameBound{
		Address: address,
		Name:    name,
	}
}

func NewEventNameUnbound(address string, name string) *EventNameUnbound {
	return &EventNameUnbound{
		Address: address,
		Name:    name,
	}
}
