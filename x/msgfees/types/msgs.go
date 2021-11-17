package types

import (
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
)

const (
	TypeCreateMsgBasedFeeRequest = "createmsgbasedfee"
)

// Compile time interface checks.
var (
	_ sdk.Msg            = &CreateMsgBasedFeeRequest{}
	_ legacytx.LegacyMsg = &CreateMsgBasedFeeRequest{} // For amino support.
)

func NewMsgBasedFee(msgTypeURL string, additionalFee sdk.Coin) MsgBasedFee {
	return MsgBasedFee{
		MsgTypeUrl: msgTypeURL, AdditionalFee: additionalFee,
	}
}

func (msg *CreateMsgBasedFeeRequest) ValidateBasic() error {
	if msg.MsgBasedFee == nil {
		return ErrEmptyMsgType
	}

	if msg.MsgBasedFee.AdditionalFee.IsZero() {
		return ErrInvalidFee
	}
	if err := msg.MsgBasedFee.AdditionalFee.Validate(); err == nil {
		return err
	}

	return nil
}

func (msg *CreateMsgBasedFeeRequest) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// GetSignBytes encodes the message for signing
func (msg *CreateMsgBasedFeeRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&msg))
}

func (msg *CreateMsgBasedFeeRequest) Type() string {
	return TypeCreateMsgBasedFeeRequest
}

// Route implements Msg
func (msg *CreateMsgBasedFeeRequest) Route() string { return ModuleName }

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
