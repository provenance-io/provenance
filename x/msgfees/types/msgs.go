package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
)

const (
	TypeCreateMsgBasedFeeRequest  = "createmsgbasedfee"
	TypeCalculateFeePerMsgRequest = "calculatefeepermsg"
)

// Compile time interface checks.
var (
	_ sdk.Msg            = &CreateMsgBasedFeeRequest{}
	_ sdk.Msg            = &CalculateFeePerMsgRequest{}
	_ legacytx.LegacyMsg = &CreateMsgBasedFeeRequest{}  // For amino support.
	_ legacytx.LegacyMsg = &CalculateFeePerMsgRequest{} // For amino support.
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

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
// func (msg MsgBasedFee) UnpackInterfaces(unpacker types.AnyUnpacker) error {
//	var msgfees MsgBasedFee
//	return unpacker.UnpackAny(msg.Msg,&msgfees)
//}

func NewCalculateFeePerMsgRequest(tx tx.Tx) CalculateFeePerMsgRequest {
	return CalculateFeePerMsgRequest{
		Tx: &tx,
	}
}

func (msg *CalculateFeePerMsgRequest) ValidateBasic() error {
	if msg.Tx == nil {
		return fmt.Errorf("msg transaction cannot be nil")
	}
	return nil
}

func (msg *CalculateFeePerMsgRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{}
}

func (msg *CalculateFeePerMsgRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&msg))
}

func (msg *CalculateFeePerMsgRequest) Route() string {
	return ModuleName
}

func (msg *CalculateFeePerMsgRequest) Type() string {
	return TypeCalculateFeePerMsgRequest
}
