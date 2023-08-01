package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	query "github.com/cosmos/cosmos-sdk/types/query"
)

var _, _, _ sdk.Msg = &MsgUpdateOracleRequest{}, &MsgQueryOracleRequest{}, &MsgSendQueryAllBalances{}

func NewMsgSendQueryAllBalances(creator, channelId string, addr string, page *query.PageRequest) *MsgSendQueryAllBalances {
	return &MsgSendQueryAllBalances{
		Creator:    creator,
		ChannelId:  channelId,
		Address:    addr,
		Pagination: page,
	}
}

func (msg *MsgSendQueryAllBalances) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgSendQueryAllBalances) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
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
