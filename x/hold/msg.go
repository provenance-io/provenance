package hold

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// AllRequestMsgs defines all the Msg*Request messages.
var AllRequestMsgs = []sdk.Msg{
	&MsgUnlockVestingAccountsRequest{},
}

// NewMsgUnlockVestingAccounts creates a new MsgUnlockVestingAccounts
func NewMsgUnlockVestingAccounts(authority string, addresses []string) *MsgUnlockVestingAccountsRequest {
	return &MsgUnlockVestingAccountsRequest{
		Authority: authority,
		Addresses: addresses,
	}
}

// ValidateBasic performs basic validation of the message
func (msg MsgUnlockVestingAccountsRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address %q: %v", msg.Authority, err)
	}

	if len(msg.Addresses) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("addresses list cannot be empty")
	}

	// Validate each address and check for duplicates
	addressSet := make(map[string]int)
	for i, addr := range msg.Addresses {
		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid addresses[%d] %q: %v", i, addr, err)
		}

		if j, exists := addressSet[addr]; exists {
			return sdkerrors.ErrInvalidRequest.Wrapf("duplicate address %q at addresses[%d] and [%d]", addr, j, i)
		}
		addressSet[addr] = i
	}

	return nil
}
