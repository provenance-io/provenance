package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Message type constants
const (
	TypeMsgUnlockVestingAccounts = "unlock_vesting_accounts"
)

// AllRequestMsgs defines all the Msg*Request messages.
var AllRequestMsgs = []sdk.Msg{
	(*MsgUnlockVestingAccountsRequest)(nil),
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
	// Validate authority address
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	// Check that we have at least one address to unlock
	if len(msg.Addresses) == 0 {
		return fmt.Errorf("addresses list cannot be empty")
	}

	// Validate each address and check for duplicates (optimized)
	addressSet := make(map[string]struct{})
	for i, addr := range msg.Addresses {
		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid address at index %d: %s", i, err)
		}
		
		if _, exists := addressSet[addr]; exists {
			return fmt.Errorf("duplicate address: %s", addr)
		}
		addressSet[addr] = struct{}{}
	}

	return nil
}

// GetSigners returns the required signers for this message
func (msg MsgUnlockVestingAccountsRequest) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

// Type returns the message type
func (msg MsgUnlockVestingAccountsRequest) Type() string {
	return TypeMsgUnlockVestingAccounts
}