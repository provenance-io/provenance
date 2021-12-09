package types

import (
	fmt "fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// ProposalTypeAddMsgBasedFee to add a new msg based fee
	ProposalTypeAddMsgBasedFee string = "AddMsgBasedFee"
	// ProposalTypeUpdateMsgBasedFee to update an existing msg based fee
	ProposalTypeUpdateMsgBasedFee string = "UpdateMsgBasedFee"
	// ProposalTypeRemoveMsgBasedFee to remove an existing msg based fee
	ProposalTypeRemoveMsgBasedFee string = "RemoveMsgBasedFee"
)

var (
	_ govtypes.Content = &AddMsgBasedFeeProposal{}
	_ govtypes.Content = &UpdateMsgBasedFeeProposal{}
	_ govtypes.Content = &RemoveMsgBasedFeeProposal{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeAddMsgBasedFee)
	govtypes.RegisterProposalTypeCodec(AddMsgBasedFeeProposal{}, "provenance/msgfees/AddMsgBasedFeeProposal")

	govtypes.RegisterProposalType(ProposalTypeUpdateMsgBasedFee)
	govtypes.RegisterProposalTypeCodec(UpdateMsgBasedFeeProposal{}, "provenance/msgfees/UpdateMsgBasedFeeProposal")

	govtypes.RegisterProposalType(ProposalTypeRemoveMsgBasedFee)
	govtypes.RegisterProposalTypeCodec(RemoveMsgBasedFeeProposal{}, "provenance/msgfees/RemoveMsgBasedFeeProposal")
}

func NewAddMsgBasedFeeProposal(
	title string,
	description string,
	msg string,
	additionalFee sdk.Coin) *AddMsgBasedFeeProposal {
	return &AddMsgBasedFeeProposal{
		Title:         title,
		Description:   description,
		MsgTypeURL:    msg,
		AdditionalFee: additionalFee,
	}
}

func (ambfp AddMsgBasedFeeProposal) ProposalRoute() string { return RouterKey }
func (ambfp AddMsgBasedFeeProposal) ProposalType() string  { return ProposalTypeAddMsgBasedFee }
func (ambfp AddMsgBasedFeeProposal) ValidateBasic() error {
	if len(ambfp.MsgTypeURL) == 0 {
		return ErrEmptyMsgType
	}

	if !ambfp.AdditionalFee.IsPositive() {
		return ErrInvalidFee
	}

	return govtypes.ValidateAbstract(&ambfp)
}
func (ambfp AddMsgBasedFeeProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Add Msg Based Fee Proposal:
Title:         %s
Description:   %s
Msg:           %s
AdditionalFee: %s
`, ambfp.Title, ambfp.Description, ambfp.MsgTypeURL, ambfp.AdditionalFee))
	return b.String()
}

func NewUpdateMsgBasedFeeProposal(
	title string,
	description string,
	msg string,
	additionalFee sdk.Coin) *UpdateMsgBasedFeeProposal {
	return &UpdateMsgBasedFeeProposal{
		Title:         title,
		Description:   description,
		MsgTypeURL:    msg,
		AdditionalFee: additionalFee,
	}
}

func (umbfp UpdateMsgBasedFeeProposal) ProposalRoute() string { return RouterKey }

func (umbfp UpdateMsgBasedFeeProposal) ProposalType() string { return ProposalTypeUpdateMsgBasedFee }

func (umbfp UpdateMsgBasedFeeProposal) ValidateBasic() error {
	if len(umbfp.MsgTypeURL) == 0 {
		return ErrEmptyMsgType
	}

	if !umbfp.AdditionalFee.IsPositive() {
		return ErrInvalidFee
	}

	return govtypes.ValidateAbstract(&umbfp)
}

func (umbfp UpdateMsgBasedFeeProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Update Msg Based Fee Proposal:
Title:         %s
Description:   %s
Msg:           %s
AdditionalFee: %s
`, umbfp.Title, umbfp.Description, umbfp.MsgTypeURL, umbfp.AdditionalFee))
	return b.String()
}

func NewRemoveMsgBasedFeeProposal(
	title string,
	description string,
	msg string,
) *RemoveMsgBasedFeeProposal {
	return &RemoveMsgBasedFeeProposal{
		Title:       title,
		Description: description,
		MsgTypeURL:  msg,
	}
}

func (rmbfp RemoveMsgBasedFeeProposal) ProposalRoute() string { return RouterKey }

func (rmbfp RemoveMsgBasedFeeProposal) ProposalType() string { return ProposalTypeRemoveMsgBasedFee }

func (rmbfp RemoveMsgBasedFeeProposal) ValidateBasic() error {
	if len(rmbfp.MsgTypeURL) > 0 {
		return ErrEmptyMsgType
	}
	return govtypes.ValidateAbstract(&rmbfp)
}

func (rmbfp RemoveMsgBasedFeeProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Remove Msg Based Fee Proposal:
  Title:       %s
  Description: %s
  MsgTypeUrl:         %s
`, rmbfp.Title, rmbfp.Description, rmbfp.MsgTypeURL))
	return b.String()
}
