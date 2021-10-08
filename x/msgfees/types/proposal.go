package types

import (
	fmt "fmt"
	"strings"

	types "github.com/cosmos/cosmos-sdk/codec/types"
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
	amount sdk.Coin,
	msg *types.Any,
	minFee sdk.Coin,
	feeRate sdk.Dec) *AddMsgBasedFeeProposal {
	return &AddMsgBasedFeeProposal{
		Title:       title,
		Description: description,
		Amount:      amount,
		Msg:         msg,
		MinFee:      minFee,
		FeeRate:     feeRate,
	}
}

func (ambfp AddMsgBasedFeeProposal) ProposalRoute() string { return RouterKey }
func (ambfp AddMsgBasedFeeProposal) ProposalType() string  { return ProposalTypeAddMsgBasedFee }
func (ambfp AddMsgBasedFeeProposal) ValidateBasic() error {
	if ambfp.Msg == nil {
		return ErrEmptyMsgType
	}

	if !ambfp.Amount.IsGTE(sdk.NewCoin(ambfp.Amount.Denom, sdk.OneInt())) {
		return ErrInvalidCoinAmount
	}

	if !ambfp.MinFee.IsPositive() && ambfp.FeeRate.IsZero() {
		return ErrInvalidFee
	}

	return govtypes.ValidateAbstract(&ambfp)
}
func (ambfp AddMsgBasedFeeProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Add Msg Based Fee Proposal:
  Title:       %s
  Description: %s
  Amount:      %s
  Msg:         %s
  MinFee:      %s
  FeeRate:     %s
`, ambfp.Title, ambfp.Description, ambfp.Amount, ambfp.Msg.GetTypeUrl(), ambfp.MinFee, ambfp.FeeRate))
	return b.String()
}

func NewUpdateMsgBasedFeeProposal(
	title string,
	description string,
	amount sdk.Coin,
	msg *types.Any,
	minFee sdk.Coin,
	feeRate sdk.Dec) *UpdateMsgBasedFeeProposal {
	return &UpdateMsgBasedFeeProposal{
		Title:       title,
		Description: description,
		Amount:      amount,
		Msg:         msg,
		MinFee:      minFee,
		FeeRate:     feeRate,
	}
}

func (umbfp UpdateMsgBasedFeeProposal) ProposalRoute() string { return RouterKey }

func (umbfp UpdateMsgBasedFeeProposal) ProposalType() string { return ProposalTypeUpdateMsgBasedFee }

func (umbfp UpdateMsgBasedFeeProposal) ValidateBasic() error {
	if umbfp.Msg == nil {
		return ErrEmptyMsgType
	}

	if !umbfp.Amount.IsGTE(sdk.NewCoin(umbfp.Amount.Denom, sdk.OneInt())) {
		return ErrInvalidCoinAmount
	}

	if !umbfp.MinFee.IsPositive() && umbfp.FeeRate.IsZero() {
		return ErrInvalidFee
	}

	return govtypes.ValidateAbstract(&umbfp)
}

func (umbfp UpdateMsgBasedFeeProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Update Msg Based Fee Proposal:
  Title:       %s
  Description: %s
  Amount:      %s
  Msg:         %s
  MinFee:      %s
  FeeRate:     %s
`, umbfp.Title, umbfp.Description, umbfp.Amount, umbfp.Msg.GetTypeUrl(), umbfp.MinFee, umbfp.FeeRate))
	return b.String()
}

func NewRemoveMsgBasedFeeProposal(
	title string,
	description string,
	msg *types.Any,
) *RemoveMsgBasedFeeProposal {
	return &RemoveMsgBasedFeeProposal{
		Title:       title,
		Description: description,
		Msg:         msg,
	}
}

func (rmbfp RemoveMsgBasedFeeProposal) ProposalRoute() string { return RouterKey }

func (rmbfp RemoveMsgBasedFeeProposal) ProposalType() string { return ProposalTypeRemoveMsgBasedFee }

func (rmbfp RemoveMsgBasedFeeProposal) ValidateBasic() error {
	if rmbfp.Msg == nil {
		return ErrEmptyMsgType
	}
	return govtypes.ValidateAbstract(&rmbfp)
}

func (rmbfp RemoveMsgBasedFeeProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Remove Msg Based Fee Proposal:
  Title:       %s
  Description: %s
  Msg:         %s
`, rmbfp.Title, rmbfp.Description, rmbfp.Msg.GetTypeUrl()))
	return b.String()
}
