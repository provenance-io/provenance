package hold

import sdk "github.com/cosmos/cosmos-sdk/types"

func NewEventHoldAdded(addr sdk.AccAddress, amount sdk.Coins, reason string) *EventHoldAdded {
	return &EventHoldAdded{
		Address: addr.String(),
		Amount:  amount.String(),
		Reason:  reason,
	}
}

func NewEventHoldReleased(addr sdk.AccAddress, amount sdk.Coins) *EventHoldReleased {
	return &EventHoldReleased{
		Address: addr.String(),
		Amount:  amount.String(),
	}
}
