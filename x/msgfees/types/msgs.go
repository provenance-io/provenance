package types

import (
	"errors"
	"fmt"

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

func NewMsgAssessCustomMsgFeeRequest(
	name string,
	amount sdk.Coin,
	recipient string,
	from string,
) MsgAssessCustomMsgFeeRequest {
	return MsgAssessCustomMsgFeeRequest{
		Name:      name,
		Amount:    amount,
		Recipient: recipient,
		From:      from,
	}
}

func (msg MsgAssessCustomMsgFeeRequest) ValidateBasic() error {
	if len(msg.Recipient) != 0 {
		_, err := sdk.AccAddressFromBech32(msg.Recipient)
		if err != nil {
			return err
		}
	}
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return err
	}
	if msg.Amount.Denom != UsdDenom && msg.Amount.Denom != NhashDenom {
		return fmt.Errorf("denom must be in usd or nhash : %s", msg.Amount.Denom)
	}
	if !msg.Amount.IsPositive() {
		return errors.New("amount must be greater than zero")
	}
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
