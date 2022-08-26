package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultMsgFeeSplit = uint32(50)
)

func NewMsgFee(msgTypeURL string, additionalFee sdk.Coin, recipient string, split uint32) MsgFee {
	return MsgFee{
		MsgTypeUrl:    msgTypeURL,
		AdditionalFee: additionalFee,
		Recipient:     recipient,
		Split:         split,
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
	if msg.Split > 100 {
		return fmt.Errorf("split can only be between 0 and 100 : %v", msg.Split)
	}

	return nil
}
