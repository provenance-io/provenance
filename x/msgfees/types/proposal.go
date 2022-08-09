package types

import (
	"errors"
	fmt "fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	// ProposalTypeAddMsgFee to add a new msg based fee
	ProposalTypeAddMsgFee string = "AddMsgFee"
	// ProposalTypeUpdateMsgFee to update an existing msg based fee
	ProposalTypeUpdateMsgFee string = "UpdateMsgFee"
	// ProposalTypeRemoveMsgFee to remove an existing msg based fee
	ProposalTypeRemoveMsgFee string = "RemoveMsgFee"
	// ProposalTypeUpdateUsdConversionRate to update the usd conversion rate param
	ProposalTypeUpdateUsdConversionRate string = "UpdateUsdConversionRate"
)

var (
	_ govtypesv1beta1.Content = &AddMsgFeeProposal{}
	_ govtypesv1beta1.Content = &UpdateMsgFeeProposal{}
	_ govtypesv1beta1.Content = &RemoveMsgFeeProposal{}
	_ govtypesv1beta1.Content = &UpdateNhashPerUsdMilProposal{}
)

func init() {
	govtypesv1beta1.RegisterProposalType(ProposalTypeAddMsgFee)
	govtypesv1beta1.RegisterProposalType(ProposalTypeUpdateMsgFee)
	govtypesv1beta1.RegisterProposalType(ProposalTypeRemoveMsgFee)
	govtypesv1beta1.RegisterProposalType(ProposalTypeUpdateUsdConversionRate)
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

func (p AddMsgFeeProposal) ProposalRoute() string { return RouterKey }
func (p AddMsgFeeProposal) ProposalType() string  { return ProposalTypeAddMsgFee }
func (p AddMsgFeeProposal) ValidateBasic() error {
	if len(p.MsgTypeUrl) == 0 {
		return ErrEmptyMsgType
	}

	if !p.AdditionalFee.IsPositive() {
		return ErrInvalidFee
	}

	return govtypesv1beta1.ValidateAbstract(&p)
}
func (p AddMsgFeeProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Add Msg Fee Proposal:
Title:         %s
Description:   %s
Msg:           %s
AdditionalFee: %s
`, p.Title, p.Description, p.MsgTypeUrl, p.AdditionalFee))
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

func (p UpdateMsgFeeProposal) ProposalRoute() string { return RouterKey }

func (p UpdateMsgFeeProposal) ProposalType() string { return ProposalTypeUpdateMsgFee }

func (p UpdateMsgFeeProposal) ValidateBasic() error {
	if len(p.MsgTypeUrl) == 0 {
		return ErrEmptyMsgType
	}

	if !p.AdditionalFee.IsPositive() {
		return ErrInvalidFee
	}

	return govtypesv1beta1.ValidateAbstract(&p)
}

func (p UpdateMsgFeeProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Update Msg Fee Proposal:
Title:         %s
Description:   %s
Msg:           %s
AdditionalFee: %s
`, p.Title, p.Description, p.MsgTypeUrl, p.AdditionalFee))
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

func (p RemoveMsgFeeProposal) ProposalRoute() string { return RouterKey }

func (p RemoveMsgFeeProposal) ProposalType() string { return ProposalTypeRemoveMsgFee }

func (p RemoveMsgFeeProposal) ValidateBasic() error {
	if len(p.MsgTypeUrl) == 0 {
		return ErrEmptyMsgType
	}
	return govtypesv1beta1.ValidateAbstract(&p)
}

func (p RemoveMsgFeeProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Remove Msg Fee Proposal:
  Title:       %s
  Description: %s
  MsgTypeUrl:  %s
`, p.Title, p.Description, p.MsgTypeUrl))
	return b.String()
}

func NewUpdateNhashPerUsdMilProposal(
	title string,
	description string,
	nhashPerUsdMil uint64,
) *UpdateNhashPerUsdMilProposal {
	return &UpdateNhashPerUsdMilProposal{
		Title:          title,
		Description:    description,
		NhashPerUsdMil: nhashPerUsdMil,
	}
}

func (p UpdateNhashPerUsdMilProposal) ProposalRoute() string { return RouterKey }

func (p UpdateNhashPerUsdMilProposal) ProposalType() string {
	return ProposalTypeUpdateUsdConversionRate
}

func (p UpdateNhashPerUsdMilProposal) ValidateBasic() error {
	if p.NhashPerUsdMil < 1 {
		return errors.New("nhash per usd mil must be greater than 0")
	}
	return govtypesv1beta1.ValidateAbstract(&p)
}

func (p UpdateNhashPerUsdMilProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Update Nhash to Usd Mil Proposal:
  Title:             %s
  Description:       %s
  NhashPerUsdMil:    %v
`, p.Title, p.Description, p.NhashPerUsdMil))
	return b.String()
}
