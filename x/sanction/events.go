package sanction

import sdk "github.com/cosmos/cosmos-sdk/types"

// NewEventAddressSanctioned creates a new event for a sanctioned address.
func NewEventAddressSanctioned(addr sdk.AccAddress) *EventAddressSanctioned {
	return &EventAddressSanctioned{
		Address: addr.String(),
	}
}

// NewEventAddressUnsanctioned creates a new event for an unsanctioned address.
func NewEventAddressUnsanctioned(addr sdk.AccAddress) *EventAddressUnsanctioned {
	return &EventAddressUnsanctioned{
		Address: addr.String(),
	}
}

// NewEventTempAddressSanctioned creates a new event for a temporarily sanctioned address.
func NewEventTempAddressSanctioned(addr sdk.AccAddress) *EventTempAddressSanctioned {
	return &EventTempAddressSanctioned{
		Address: addr.String(),
	}
}

// NewEventTempAddressUnsanctioned creates a new event for a temporarily unsanctioned address.
func NewEventTempAddressUnsanctioned(addr sdk.AccAddress) *EventTempAddressUnsanctioned {
	return &EventTempAddressUnsanctioned{
		Address: addr.String(),
	}
}
