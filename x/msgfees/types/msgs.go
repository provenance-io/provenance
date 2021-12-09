package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)



func NewMsgBasedFee(msgTypeURL string, additionalFee sdk.Coin) MsgBasedFee {
	return MsgBasedFee{
		MsgTypeUrl: msgTypeURL, AdditionalFee: additionalFee,
	}
}

func (msg *MsgBasedFee) ValidateBasic() error {
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
