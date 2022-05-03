package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gopkg.in/yaml.v2"
)

const (
	TypeMsgCreateRewardProgramRequest = "create_reward_program"
)

// Compile time interface checks.
var _ sdk.Msg = &MsgCreateRewardProgramRequest{}

// MsgCreateRewardProgramRequest creates a new create reward program request
func NewMsgCreateRewardProgramRequest() *MsgCreateRewardProgramRequest {
	return &MsgCreateRewardProgramRequest{}
}

// Route implements Msg
func (msg MsgCreateRewardProgramRequest) Route() string { return ModuleName }

// Type implements Msg
func (msg MsgCreateRewardProgramRequest) Type() string { return TypeMsgCreateRewardProgramRequest }

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgCreateRewardProgramRequest) ValidateBasic() error {

	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgCreateRewardProgramRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners indicates that the message must have been signed by the parent.
func (msg MsgCreateRewardProgramRequest) GetSigners() []sdk.AccAddress {
	// addr, err := sdk.AccAddressFromBech32(ms)
	// if err != nil {
	// 	panic(err)
	// }
	return []sdk.AccAddress{}
}

// GetSignBytes encodes the message for signing
func (msg MsgCreateRewardProgramRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}
