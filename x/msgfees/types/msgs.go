package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (

	// AssessCustomMsgFeeBips is the bips the recipient will get
	// This should be a message level data (present in TypeAssessCustomMsgFee = "assess_custom_msg_fee") i think so that it can be defined by the smart contract writer
	// or at the very least it can be a module param.
	// for now i am hard coding it to avoid breaking any clients and because of this ticket https://github.com/provenance-io/provenance/issues/1263
	AssessCustomMsgFeeBips = 10_000

	TypeAssessCustomMsgFee = "assess_custom_msg_fee"
)

// Compile time interface checks.
var (
	_ sdk.Msg = &MsgAssessCustomMsgFeeRequest{}
)

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

// Route returns the module route
func (msg MsgAssessCustomMsgFeeRequest) Route() string {
	return ModuleName
}

// Type returns the type name for this msg
func (msg MsgAssessCustomMsgFeeRequest) Type() string {
	return TypeAssessCustomMsgFee
}
