package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgUpdateOracleRequest{}

// GetSigners indicates that the message must have been signed by the parent.
func (msg MsgUpdateOracleRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Authority)}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgUpdateOracleRequest) ValidateBasic() error {
	return nil
}

// GetSigners indicates that the message must have been signed by the parent.
func (msg MsgQueryOracleRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Authority)}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgQueryOracleRequest) ValidateBasic() error {
	return nil
}
