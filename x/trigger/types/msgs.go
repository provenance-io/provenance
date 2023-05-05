package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var _ sdk.Msg = &MsgCreateTriggerRequest{}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgCreateTriggerRequest) ValidateBasic() error {
	return nil
}

// GetSigners indicates that the message must have been signed by the parent.
func (msg MsgCreateTriggerRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.GetAuthority())
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}
