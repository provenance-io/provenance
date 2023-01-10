package types

import (
	"errors"
	"fmt"
	"strconv"

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
	recipientBasisPoints string,
) MsgAssessCustomMsgFeeRequest {
	return MsgAssessCustomMsgFeeRequest{
		Name:                 name,
		Amount:               amount,
		Recipient:            recipient,
		From:                 from,
		RecipientBasisPoints: recipientBasisPoints,
	}
}

// ValidateBasic runs stateless validation checks on the message.
func (msg MsgAssessCustomMsgFeeRequest) ValidateBasic() error {
	if len(msg.Recipient) != 0 {
		if _, err := sdk.AccAddressFromBech32(msg.Recipient); err != nil {
			return err
		}
	}
	if _, err := sdk.AccAddressFromBech32(msg.From); err != nil {
		return err
	}
	if !msg.Amount.IsPositive() {
		return errors.New("amount must be greater than zero")
	}
	if _, err := msg.GetBips(); err != nil {
		return err
	}
	return nil
}

// GetBips converts the msg RecipientBasisPoints to a uint32 basis point value 0 - 10,000
func (msg MsgAssessCustomMsgFeeRequest) GetBips() (uint32, error) {
	if msg.RecipientBasisPoints == "" {
		return AssessCustomMsgFeeBips, nil
	}
	bips, err := strconv.ParseUint(msg.RecipientBasisPoints, 10, 32)
	if err == nil {
		if bips > 10_000 {
			return 0, fmt.Errorf("recipient basis points can only be between 0 and 10,000 : %v", msg.RecipientBasisPoints)
		}
	}

	return uint32(bips), err
}

// GetSignBytes encodes the message for signing
func (msg MsgAssessCustomMsgFeeRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}
