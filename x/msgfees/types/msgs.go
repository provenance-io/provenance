package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
)

// Compile time interface checks.
var (
	_ sdk.Msg = &MsgAddFeeForMsgTypeRequest{}
	_ legacytx.LegacyMsg = &MsgAddFeeForMsgTypeRequest{} // For amino support.
)


func (m *MsgAddFeeForMsgTypeRequest) ValidateBasic() error {
	panic("implement me")
}

func (m *MsgAddFeeForMsgTypeRequest) GetSigners() []sdk.AccAddress {
	panic("implement me")
}

// GetSignBytes encodes the message for signing
func (msg *MsgAddFeeForMsgTypeRequest) GetSignBytes() []byte {
	panic("implement me")
}


func (m *MsgAddFeeForMsgTypeRequest) Type() string {
	panic("implement me")
}


// Route implements Msg
func (msg *MsgAddFeeForMsgTypeRequest) Route() string { return ModuleName }
