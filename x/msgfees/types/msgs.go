package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AllRequestMsgs defines all the Msg*Request messages.
var AllRequestMsgs = []sdk.Msg{
	(*MsgAssessCustomMsgFeeRequest)(nil),
	(*MsgAddMsgFeeProposalRequest)(nil),
	(*MsgUpdateMsgFeeProposalRequest)(nil),
	(*MsgRemoveMsgFeeProposalRequest)(nil),
	(*MsgUpdateConversionFeeDenomProposalRequest)(nil),
	(*MsgUpdateNhashPerUsdMilProposalRequest)(nil),
}

var errDep = errors.New("deprecated and unusable")

func (msg MsgAssessCustomMsgFeeRequest) ValidateBasic() error { return errDep }

func (msg MsgAddMsgFeeProposalRequest) ValidateBasic() error { return errDep }

func (msg MsgUpdateMsgFeeProposalRequest) ValidateBasic() error { return errDep }

func (msg MsgRemoveMsgFeeProposalRequest) ValidateBasic() error { return errDep }

func (msg MsgUpdateConversionFeeDenomProposalRequest) ValidateBasic() error { return errDep }

func (msg MsgUpdateNhashPerUsdMilProposalRequest) ValidateBasic() error { return errDep }
