package types

import (
	fmt "fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// ProposalTypeAddMsgFee to add a new msg based fee
	ProposalTypeAddMsgFee string = "AddMsgFee"
	// ProposalTypeUpdateMsgFee to update an existing msg based fee
	ProposalTypeUpdateMsgFee string = "UpdateMsgFee"
	// ProposalTypeRemoveMsgFee to remove an existing msg based fee
	ProposalTypeRemoveMsgFee string = "RemoveMsgFee"
)

var (
	_ govtypes.Content = &AddMsgFeeProposal{}
	_ govtypes.Content = &UpdateMsgFeeProposal{}
	_ govtypes.Content = &RemoveMsgFeeProposal{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeAddMsgFee)
	govtypes.RegisterProposalTypeCodec(AddMsgFeeProposal{}, "provenance/msgfees/AddMsgFeeProposal")

	govtypes.RegisterProposalType(ProposalTypeUpdateMsgFee)
	govtypes.RegisterProposalTypeCodec(UpdateMsgFeeProposal{}, "provenance/msgfees/UpdateMsgFeeProposal")

	govtypes.RegisterProposalType(ProposalTypeRemoveMsgFee)
	govtypes.RegisterProposalTypeCodec(RemoveMsgFeeProposal{}, "provenance/msgfees/RemoveMsgFeeProposal")
}

func NewAddMsgFeeProposal(
	title string,
	description string,
	msg string,
	additionalFee sdk.Coin) *AddMsgFeeProposal {
	return &AddMsgFeeProposal{
		Title:         title,
		Description:   description,
		MsgTypeUrl:    msg,
		AdditionalFee: additionalFee,
	}
}

func (ambfp AddMsgFeeProposal) ProposalRoute() string { return RouterKey }
func (ambfp AddMsgFeeProposal) ProposalType() string  { return ProposalTypeAddMsgFee }
func (ambfp AddMsgFeeProposal) ValidateBasic() error {
	if len(ambfp.MsgTypeUrl) == 0 {
		return ErrEmptyMsgType
	}

	if !ambfp.AdditionalFee.IsPositive() {
		return ErrInvalidFee
	}

	return govtypes.ValidateAbstract(&ambfp)
}
func (ambfp AddMsgFeeProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Add Msg Fee Proposal:
Title:         %s
Description:   %s
Msg:           %s
AdditionalFee: %s
`, ambfp.Title, ambfp.Description, ambfp.MsgTypeUrl, ambfp.AdditionalFee))
	return b.String()
}

func NewUpdateMsgFeeProposal(
	title string,
	description string,
	msg string,
	additionalFee sdk.Coin) *UpdateMsgFeeProposal {
	return &UpdateMsgFeeProposal{
		Title:         title,
		Description:   description,
		MsgTypeUrl:    msg,
		AdditionalFee: additionalFee,
	}
}

func (umbfp UpdateMsgFeeProposal) ProposalRoute() string { return RouterKey }

func (umbfp UpdateMsgFeeProposal) ProposalType() string { return ProposalTypeUpdateMsgFee }

func (umbfp UpdateMsgFeeProposal) ValidateBasic() error {
	if len(umbfp.MsgTypeUrl) == 0 {
		return ErrEmptyMsgType
	}

	if !umbfp.AdditionalFee.IsPositive() {
		return ErrInvalidFee
	}

	return govtypes.ValidateAbstract(&umbfp)
}

func (umbfp UpdateMsgFeeProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Update Msg Fee Proposal:
Title:         %s
Description:   %s
Msg:           %s
AdditionalFee: %s
`, umbfp.Title, umbfp.Description, umbfp.MsgTypeUrl, umbfp.AdditionalFee))
	return b.String()
}

func NewRemoveMsgFeeProposal(
	title string,
	description string,
	msgTypeURL string,
) *RemoveMsgFeeProposal {
	return &RemoveMsgFeeProposal{
		Title:       title,
		Description: description,
		MsgTypeUrl:  msgTypeURL,
	}
}

func (rmbfp RemoveMsgFeeProposal) ProposalRoute() string { return RouterKey }

func (rmbfp RemoveMsgFeeProposal) ProposalType() string { return ProposalTypeRemoveMsgFee }

func (rmbfp RemoveMsgFeeProposal) ValidateBasic() error {
	if len(rmbfp.MsgTypeUrl) == 0 {
		return ErrEmptyMsgType
	}
	return govtypes.ValidateAbstract(&rmbfp)
}

func (rmbfp RemoveMsgFeeProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Remove Msg Fee Proposal:
  Title:       %s
  Description: %s
  MsgTypeUrl:         %s
`, rmbfp.Title, rmbfp.Description, rmbfp.MsgTypeUrl))
	return b.String()
}
