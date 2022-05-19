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

func (msg MsgAssessCustomMsgFeeRequest) ValidateBasic() error {
	return nil
}

func (msg MsgAssessCustomMsgFeeRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg MsgAssessCustomMsgFeeRequest) GetSignerBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgAssessCustomMsgFeeRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// Route returns the module route
func (msg MsgAssessCustomMsgFeeRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgAssessCustomMsgFeeRequest) Type() string {
	return TypeAssessCustomMsgFee
}
