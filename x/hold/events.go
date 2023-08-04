package hold

import sdk "github.com/cosmos/cosmos-sdk/types"

func NewEventHoldAdded(addr sdk.AccAddress, amount sdk.Coins) *EventHoldAdded {
	return &EventHoldAdded{
		Address: addr.String(),
		Amount:  amount.String(),
	}
}

func NewEventHoldReleased(addr sdk.AccAddress, amount sdk.Coins) *EventHoldReleased {
	return &EventHoldReleased{
		Address: addr.String(),
		Amount:  amount.String(),
	}
}
