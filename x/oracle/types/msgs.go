package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
)

// AllRequestMsgs defines all the Msg*Request messages.
var AllRequestMsgs = []sdk.Msg{
	(*MsgUpdateOracleRequest)(nil),
	(*MsgSendQueryOracleRequest)(nil),
}

// NewMsgSendQueryOracle creates a new MsgSendQueryOracleRequest
func NewMsgSendQueryOracle(creator, channelID string, query []byte) *MsgSendQueryOracleRequest {
	return &MsgSendQueryOracleRequest{
		Authority: creator,
		Channel:   channelID,
		Query:     query,
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgSendQueryOracleRequest) ValidateBasic() error {
	if err := host.ChannelIdentifierValidator(msg.Channel); err != nil {
		return fmt.Errorf("invalid channel id")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return fmt.Errorf("invalid authority address: %w", err)
	}
	if err := msg.Query.ValidateBasic(); err != nil {
		return fmt.Errorf("invalid query data: %w", err)
	}
	return nil
}

// NewMsgUpdateOracle creates a new MsgUpdateOracleRequest
func NewMsgUpdateOracle(creator, addr string) *MsgUpdateOracleRequest {
	return &MsgUpdateOracleRequest{
		Authority: creator,
		Address:   addr,
	}
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
