package hold

import sdk "github.com/cosmos/cosmos-sdk/types"

// NewEventHoldAdded creates a new event signaling that a hold was added.
func NewEventHoldAdded(addr sdk.AccAddress, amount sdk.Coins, reason string) *EventHoldAdded {
	return &EventHoldAdded{
		Address: addr.String(),
		Amount:  amount.String(),
		Reason:  reason,
	}
}

// NewEventHoldReleased creates a new event signaling that a hold was released.
func NewEventHoldReleased(addr sdk.AccAddress, amount sdk.Coins) *EventHoldReleased {
	return &EventHoldReleased{
		Address: addr.String(),
		Amount:  amount.String(),
	}
}

// NewEventVestingAccountUnlocked creates a new event signaling that a vesting account was unlocked.
func NewEventVestingAccountUnlocked(addr sdk.AccAddress) *EventVestingAccountUnlocked {
	return &EventVestingAccountUnlocked{Address: addr.String()}
}
