package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _, _ sdk.Msg = &MsgUpdateOracleRequest{}, &MsgQueryOracleRequest{}

func NewMsgQueryOracle(creator, channelId string, query []byte) *MsgQueryOracleRequest {
	return &MsgQueryOracleRequest{
		Authority: creator,
		Channel:   channelId,
		Query:     query,
	}
}

func NewMsgUpdateOracle(creator, addr string) *MsgUpdateOracleRequest {
	return &MsgUpdateOracleRequest{
		Authority: creator,
		Address:   addr,
	}
}

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
