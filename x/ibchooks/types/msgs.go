package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// constants.
const (
	TypeMsgEmitIBCAck = "emit-ibc-ack"
)

var _ sdk.Msg = &MsgEmitIBCAck{}

func (m MsgEmitIBCAck) Route() string { return RouterKey }
func (m MsgEmitIBCAck) Type() string  { return TypeMsgEmitIBCAck }
func (m MsgEmitIBCAck) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	return err
}

func (m MsgEmitIBCAck) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgEmitIBCAck) GetSigners() []sdk.AccAddress {
	sender, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{sender}
}
