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
	_ sdk.Msg            = &CalculateFeePerMsgRequest{}
	_ legacytx.LegacyMsg = &CreateMsgBasedFeeRequest{} // For amino support.
	_ legacytx.LegacyMsg = &CalculateFeePerMsgRequest{} // For amino support.
)

func NewMsgFees(msgTypeURL string, minFeeRate sdk.Coins, feeRate sdk.Dec) MsgFees {
	return MsgFees{
		MsgTypeUrl: msgTypeURL, MinAdditionalFee: minFeeRate, FeeRate: feeRate,
	}
}

func (msg *CreateMsgBasedFeeRequest) ValidateBasic() error {
	if msg.MsgFees == nil {
		return ErrEmptyMsgType
	}

	if msg.MsgFees.MinAdditionalFee.Empty() && msg.MsgFees.FeeRate.IsZero() {
		return ErrInvalidFee
	}

	if !msg.MsgFees.MinAdditionalFee.Empty() && len(msg.MsgFees.MinAdditionalFee) > 0 {
		msg.MsgFees.MinAdditionalFee.Validate()
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
//func (msg MsgFees) UnpackInterfaces(unpacker types.AnyUnpacker) error {
//	var msgfees MsgFees
//	return unpacker.UnpackAny(msg.Msg,&msgfees)
//}


func (msg *CalculateFeePerMsgRequest) ValidateBasic() error {
	panic("implement me")
}

func (m *CalculateFeePerMsgRequest) GetSigners() []sdk.AccAddress {
	panic("implement me")
}

func (msg *CalculateFeePerMsgRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&msg))
}

func (msg *CalculateFeePerMsgRequest) Route() string {
	panic("implement me")
}

func (msg *CalculateFeePerMsgRequest) Type() string {
	panic("implement me")
}
