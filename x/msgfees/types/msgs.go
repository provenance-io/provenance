package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	yaml "gopkg.in/yaml.v2"
)

const (
	TypeAssessCustomMsgFee = "assess_custom_msg_fee"
)

// Compile time interface checks.
var (
	_ sdk.Msg = &MsgAssessCustomMsgFeeRequest{}
)

func NewMsgFee(msgTypeURL string, additionalFee sdk.Coin) MsgFee {
	return MsgFee{
		MsgTypeUrl: msgTypeURL, AdditionalFee: additionalFee,
	}
}

func (msg *MsgFee) ValidateBasic() error {
	if msg == nil {
		return ErrEmptyMsgType
	}

	if msg.AdditionalFee.IsZero() {
		return ErrInvalidFee
	}
	if err := msg.AdditionalFee.Validate(); err == nil {
		return err
	}

	return nil
}

func (acmfr MsgAssessCustomMsgFeeRequest) ValidateBasic() error {
	return nil
}

func (acmfr MsgAssessCustomMsgFeeRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{}
}

func (acmfr MsgAssessCustomMsgFeeRequest) String() string {
	out, _ := yaml.Marshal(acmfr)
	return string(out)
}

// Route returns the module route
func (acmfr MsgAssessCustomMsgFeeRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (acmfr MsgAssessCustomMsgFeeRequest) Type() string {
	return TypeAssessCustomMsgFee
}
