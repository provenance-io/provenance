package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultMsgFeeSplit = uint32(50)
)

func NewMsgFee(msgTypeURL string, additionalFee sdk.Coin, recipient string, recipientBasisPoints uint32) MsgFee {
	return MsgFee{
		MsgTypeUrl:           msgTypeURL,
		AdditionalFee:        additionalFee,
		Recipient:            recipient,
		RecipientBasisPoints: recipientBasisPoints,
	}
}

func (msg *MsgFee) Validate() error {
	if msg == nil {
		return ErrEmptyMsgType
	}

	if len(msg.MsgTypeUrl) == 0 {
		return fmt.Errorf("invalid msg type url")
	}

	if msg.AdditionalFee.IsZero() {
		return ErrInvalidFee
	}
	if err := msg.AdditionalFee.Validate(); err != nil {
		return err
	}
	if len(msg.Recipient) != 0 {
		_, err := sdk.AccAddressFromBech32(msg.Recipient)
		if err != nil {
			return err
		}
	}
	if msg.RecipientBasisPoints > 10_000 || msg.RecipientBasisPoints < 1 {
		return fmt.Errorf("recipient basis points can only be between 1 and 10,000 : %v", msg.RecipientBasisPoints)
	}

	return nil
}
