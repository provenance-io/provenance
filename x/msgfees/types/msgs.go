package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
)

// Compile time interface checks.
var (
	_ sdk.Msg            = &CreateMsgBasedFeeRequest{}
	_ legacytx.LegacyMsg = &CreateMsgBasedFeeRequest{} // For amino support.
)

func (m *CreateMsgBasedFeeRequest) ValidateBasic() error {
	panic("implement me")
}

func (m *CreateMsgBasedFeeRequest) GetSigners() []sdk.AccAddress {
	panic("implement me")
}

// GetSignBytes encodes the message for signing
func (msg *CreateMsgBasedFeeRequest) GetSignBytes() []byte {
	panic("implement me")
}

func (m *CreateMsgBasedFeeRequest) Type() string {
	panic("implement me")
}

// Route implements Msg
func (msg *CreateMsgBasedFeeRequest) Route() string { return ModuleName }
