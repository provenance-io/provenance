package hold

import sdk "github.com/cosmos/cosmos-sdk/types"

func NewEventEscrowAdded(addr sdk.AccAddress, amount sdk.Coins) *EventEscrowAdded {
	return &EventEscrowAdded{
		Address: addr.String(),
		Amount:  amount.String(),
	}
}

func NewEventEscrowRemoved(addr sdk.AccAddress, amount sdk.Coins) *EventEscrowRemoved {
	return &EventEscrowRemoved{
		Address: addr.String(),
		Amount:  amount.String(),
	}
}
