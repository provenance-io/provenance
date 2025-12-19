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

// ValidateBasic implements the sdk.Msg interface for MsgAssessCustomMsgFeeRequest
func (msg MsgAssessCustomMsgFeeRequest) ValidateBasic() error { return errDep }

// ValidateBasic implements the sdk.Msg interface for MsgAddMsgFeeProposalRequest
func (msg MsgAddMsgFeeProposalRequest) ValidateBasic() error { return errDep }

// ValidateBasic implements the sdk.Msg interface for MsgUpdateMsgFeeProposalRequest
func (msg MsgUpdateMsgFeeProposalRequest) ValidateBasic() error { return errDep }

// ValidateBasic implements the sdk.Msg interface for MsgRemoveMsgFeeProposalRequest
func (msg MsgRemoveMsgFeeProposalRequest) ValidateBasic() error { return errDep }

// ValidateBasic implements the sdk.Msg interface for MsgUpdateConversionFeeDenomProposalRequest
func (msg MsgUpdateConversionFeeDenomProposalRequest) ValidateBasic() error { return errDep }

// ValidateBasic implements the sdk.Msg interface for MsgUpdateNhashPerUsdMilProposalRequest
func (msg MsgUpdateNhashPerUsdMilProposalRequest) ValidateBasic() error { return errDep }
