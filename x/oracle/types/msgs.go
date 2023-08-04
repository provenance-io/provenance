package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	host "github.com/cosmos/ibc-go/v6/modules/core/24-host"
)

var _, _ sdk.Msg = &MsgUpdateOracleRequest{}, &MsgSendQueryOracleRequest{}

// NewMsgSendQueryOracle creates a new MsgSendQueryOracleRequest
func NewMsgSendQueryOracle(creator, channelId string, query []byte) *MsgSendQueryOracleRequest {
	return &MsgSendQueryOracleRequest{
		Authority: creator,
		Channel:   channelId,
		Query:     query,
	}
}

// NewMsgUpdateOracle creates a new MsgUpdateOracleRequest
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
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return fmt.Errorf("invalid address for oracle: %w", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return fmt.Errorf("invalid authority address: %w", err)
	}
	return nil
}

// GetSigners indicates that the message must have been signed by the parent.
func (msg MsgSendQueryOracleRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Authority)}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgSendQueryOracleRequest) ValidateBasic() error {
	if err := host.ChannelIdentifierValidator(msg.Channel); err != nil {
		return fmt.Errorf("invalid channel id: %w", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return fmt.Errorf("invalid authority address: %w", err)
	}
	if err := msg.Query.ValidateBasic(); err != nil {
		return fmt.Errorf("invalid query data: %w", err)
	}
	return nil
}
