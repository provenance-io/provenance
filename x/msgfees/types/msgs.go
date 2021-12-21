package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
