package hold

import sdk "github.com/cosmos/cosmos-sdk/types"

func NewEventHoldAdded(addr sdk.AccAddress, amount sdk.Coins) *EventHoldAdded {
	return &EventHoldAdded{
		Address: addr.String(),
		Amount:  amount.String(),
	}
}

func NewEventHoldRemoved(addr sdk.AccAddress, amount sdk.Coins) *EventHoldRemoved {
	return &EventHoldRemoved{
		Address: addr.String(),
		Amount:  amount.String(),
	}
}
