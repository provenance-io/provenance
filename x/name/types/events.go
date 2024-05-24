package types

import "strconv"

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

func NewEventNameBound(address string, name string, restricted bool) *EventNameBound {
	return &EventNameBound{
		Address:    address,
		Name:       name,
		Restricted: restricted,
	}
}

func NewEventNameUnbound(address string, name string, restricted bool) *EventNameUnbound {
	return &EventNameUnbound{
		Address:    address,
		Name:       name,
		Restricted: restricted,
	}
}

func NewEventNameUpdate(address string, name string, restricted bool) *EventNameUpdate {
	return &EventNameUpdate{
		Address:    address,
		Name:       name,
		Restricted: restricted,
	}
}

// NewEventNameParamsUpdated returns a new instance of EventNameParamsUpdated
func NewEventNameParamsUpdated(allowUnrestrictedNames bool, maxNameLevels, minSegmentLength, maxSegmentLength uint32) *EventNameParamsUpdated {
	return &EventNameParamsUpdated{
		AllowUnrestrictedNames: strconv.FormatBool(allowUnrestrictedNames),
		MaxNameLevels:          strconv.FormatUint(uint64(maxNameLevels), 10),
		MinSegmentLength:       strconv.FormatUint(uint64(minSegmentLength), 10),
		MaxSegmentLength:       strconv.FormatUint(uint64(maxSegmentLength), 10),
	}
}
