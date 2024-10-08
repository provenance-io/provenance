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
)

// AllRequestMsgs defines all the Msg*Request messages.
var AllRequestMsgs = []sdk.Msg{
	(*MsgAssessCustomMsgFeeRequest)(nil),
	(*MsgAddMsgFeeProposalRequest)(nil),
	(*MsgUpdateMsgFeeProposalRequest)(nil),
	(*MsgRemoveMsgFeeProposalRequest)(nil),
	(*MsgUpdateConversionFeeDenomProposalRequest)(nil),
	(*MsgUpdateNhashPerUsdMilProposalRequest)(nil),
}

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
	return uint32(bips), err //nolint:gosec // G115: ParseUint bitsize is 32, so we know this is okay.
}

func NewMsgAddMsgFeeProposalRequest(msgTypeURL string, additionalFee sdk.Coin, recipient string, recipientBasisPoints string, authority string) *MsgAddMsgFeeProposalRequest {
	return &MsgAddMsgFeeProposalRequest{
		MsgTypeUrl:           msgTypeURL,
		AdditionalFee:        additionalFee,
		Recipient:            recipient,
		RecipientBasisPoints: recipientBasisPoints,
		Authority:            authority,
	}
}

func (msg *MsgAddMsgFeeProposalRequest) ValidateBasic() error {
	if len(msg.MsgTypeUrl) == 0 {
		return ErrEmptyMsgType
	}

	if !msg.AdditionalFee.IsPositive() {
		return ErrInvalidFee
	}

	if err := msg.AdditionalFee.Validate(); err != nil {
		return err
	}

	if err := ValidateBips(msg.Recipient, msg.RecipientBasisPoints); err != nil {
		return err
	}

	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return err
	}

	return nil
}

func ValidateBips(recipient, recipientBasisPoints string) error {
	if len(recipient) != 0 {
		_, err := sdk.AccAddressFromBech32(recipient)
		if err != nil {
			return err
		}
	}

	if len(recipientBasisPoints) > 0 && len(recipient) > 0 {
		bips, err := strconv.ParseUint(recipientBasisPoints, 0, 64)
		if err != nil {
			return err
		}
		if bips > 10_000 {
			return fmt.Errorf("recipient basis points can only be between 0 and 10,000 : %v", recipientBasisPoints)
		}
	} else if len(recipientBasisPoints) > 0 && len(recipient) == 0 {
		return fmt.Errorf("recipient basis points provided without a recipient")
	}
	return nil
}

func NewMsgUpdateMsgFeeProposalRequest(msgTypeURL string, additionalFee sdk.Coin, recipient string, recipientBasisPoints string, authority string) *MsgUpdateMsgFeeProposalRequest {
	return &MsgUpdateMsgFeeProposalRequest{
		MsgTypeUrl:           msgTypeURL,
		AdditionalFee:        additionalFee,
		Recipient:            recipient,
		RecipientBasisPoints: recipientBasisPoints,
		Authority:            authority,
	}
}

func (msg *MsgUpdateMsgFeeProposalRequest) ValidateBasic() error {
	if len(msg.MsgTypeUrl) == 0 {
		return ErrEmptyMsgType
	}

	if !msg.AdditionalFee.IsPositive() {
		return ErrInvalidFee
	}

	if err := ValidateBips(msg.Recipient, msg.RecipientBasisPoints); err != nil {
		return err
	}

	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return err
	}

	return nil
}

func NewMsgRemoveMsgFeeProposalRequest(msgTypeURL string, authority string) *MsgRemoveMsgFeeProposalRequest {
	return &MsgRemoveMsgFeeProposalRequest{
		MsgTypeUrl: msgTypeURL,
		Authority:  authority,
	}
}

func (msg *MsgRemoveMsgFeeProposalRequest) ValidateBasic() error {
	if len(msg.MsgTypeUrl) == 0 {
		return ErrEmptyMsgType
	}

	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return err
	}

	return nil
}

func NewMsgUpdateNhashPerUsdMilProposalRequest(nhashPerUsdMil uint64, authority string) *MsgUpdateNhashPerUsdMilProposalRequest {
	return &MsgUpdateNhashPerUsdMilProposalRequest{
		NhashPerUsdMil: nhashPerUsdMil,
		Authority:      authority,
	}
}

func (msg *MsgUpdateConversionFeeDenomProposalRequest) ValidateBasic() error {
	if err := sdk.ValidateDenom(msg.ConversionFeeDenom); err != nil {
		return err
	}

	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return err
	}

	return nil
}

func NewMsgUpdateConversionFeeDenomProposalRequest(conversionFeeDenom string, authority string) *MsgUpdateConversionFeeDenomProposalRequest {
	return &MsgUpdateConversionFeeDenomProposalRequest{
		ConversionFeeDenom: conversionFeeDenom,
		Authority:          authority,
	}
}

func (msg *MsgUpdateNhashPerUsdMilProposalRequest) ValidateBasic() error {
	if msg.NhashPerUsdMil < 1 {
		return errors.New("nhash per usd mil must be greater than 0")
	}

	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return err
	}

	return nil
}
