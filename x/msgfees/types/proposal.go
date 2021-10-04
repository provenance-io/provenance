package types

import (
	fmt "fmt"
	"strings"

	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// ProposalTypeAddMsgBasedFees to add a new msg based fees
	ProposalTypeAddMsgBasedFees string = "AddMsgBasedFees"
	// ProposalTypeUpdateMsgBasedFees to update an existing msg based fees
	ProposalTypeUpdateMsgBasedFees string = "UpdateMsgBasedFees"
	// ProposalTypeRemoveMsgBasedFees to remove an existing msg based fees
	ProposalTypeRemoveMsgBasedFees string = "RemoveMsgBasedFees"
)

var (
	_ govtypes.Content = &AddMsgBasedFeesProposal{}
	_ govtypes.Content = &UpdateMsgBasedFeesProposal{}
	_ govtypes.Content = &RemoveMsgBasedFeesProposal{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeAddMsgBasedFees)
	govtypes.RegisterProposalTypeCodec(AddMsgBasedFeesProposal{}, "provenance/msgfees/AddMsgBasedFeesProposal")

	govtypes.RegisterProposalType(ProposalTypeUpdateMsgBasedFees)
	govtypes.RegisterProposalTypeCodec(UpdateMsgBasedFeesProposal{}, "provenance/msgfees/UpdateMsgBasedFeesProposal")

	govtypes.RegisterProposalType(ProposalTypeRemoveMsgBasedFees)
	govtypes.RegisterProposalTypeCodec(RemoveMsgBasedFeesProposal{}, "provenance/msgfees/RemoveMsgBasedFeesProposal")
}

func NewAddMsgBasedFeesProposal(
	title string,
	description string,
	amount sdk.Coin,
	msg *types.Any,
	minFee sdk.Coins,
	feeRate sdk.Dec) *AddMsgBasedFeesProposal {

	return &AddMsgBasedFeesProposal{
		Title:       title,
		Description: description,
		Amount:      amount,
		Msg:         msg,
		MinFee:      minFee,
		FeeRate:     feeRate,
	}
}

func (ambfp AddMsgBasedFeesProposal) ProposalRoute() string { return RouterKey }
func (ambfp AddMsgBasedFeesProposal) ProposalType() string  { return ProposalTypeAddMsgBasedFees }
func (ambfp AddMsgBasedFeesProposal) ValidateBasic() error {
	if ambfp.Msg == nil {
		return ErrEmptyMsgType
	}

	if !ambfp.Amount.IsGTE(sdk.NewCoin(ambfp.Amount.Denom, sdk.OneInt())) {
		return ErrInvalidCoinAmount
	}

	if ambfp.MinFee.Empty() && ambfp.FeeRate.IsZero() {
		return ErrInvalidFee
	}

	return govtypes.ValidateAbstract(&ambfp)
}
func (ambfp AddMsgBasedFeesProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Add Msg Based Fees Proposal:
	Title:       %s
	Description: %s
	Amount:      %s
	Msg:         %s
	MinFee:      %s
	FeeRate:     %s
  `, ambfp.Title, ambfp.Description, ambfp.Amount, ambfp.Msg, ambfp.MinFee, ambfp.FeeRate))
	return b.String()
}

func NewUpdateMsgBasedFeesProposal(
	title string,
	description string,
	amount sdk.Coin,
	msg *types.Any,
	minFee sdk.Coins,
	feeRate sdk.Dec) *UpdateMsgBasedFeesProposal {

	return &UpdateMsgBasedFeesProposal{
		Title:       title,
		Description: description,
		Amount:      amount,
		Msg:         msg,
		MinFee:      minFee,
		FeeRate:     feeRate,
	}
}

func (umbfp UpdateMsgBasedFeesProposal) ProposalRoute() string { return RouterKey }

func (umbfp UpdateMsgBasedFeesProposal) ProposalType() string { return ProposalTypeUpdateMsgBasedFees }

func (umbfp UpdateMsgBasedFeesProposal) ValidateBasic() error {
	if umbfp.Msg == nil {
		return ErrEmptyMsgType
	}

	if !umbfp.Amount.IsGTE(sdk.NewCoin(umbfp.Amount.Denom, sdk.OneInt())) {
		return ErrInvalidCoinAmount
	}

	if umbfp.MinFee.Empty() && umbfp.FeeRate.IsZero() {
		return ErrInvalidFee
	}

	return govtypes.ValidateAbstract(&umbfp)
}

func (umbfp UpdateMsgBasedFeesProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Update Msg Based Fees Proposal:
	Title:       %s
	Description: %s
	Amount:      %s
	Msg:         %s
	MinFee:      %s
	FeeRate:     %s
  `, umbfp.Title, umbfp.Description, umbfp.Amount, umbfp.Msg, umbfp.MinFee, umbfp.FeeRate))
	return b.String()
}

func NewRemoveMsgBasedFeesProposal(
	title string,
	description string,
	msg *types.Any,
) *RemoveMsgBasedFeesProposal {

	return &RemoveMsgBasedFeesProposal{
		Title:       title,
		Description: description,
		Msg:         msg,
	}
}

func (rmbfp RemoveMsgBasedFeesProposal) ProposalRoute() string { return RouterKey }

func (rmbfp RemoveMsgBasedFeesProposal) ProposalType() string { return ProposalTypeRemoveMsgBasedFees }

func (rmbfp RemoveMsgBasedFeesProposal) ValidateBasic() error {
	if rmbfp.Msg == nil {
		return ErrEmptyMsgType
	}
	return govtypes.ValidateAbstract(&rmbfp)
}

func (rmbfp RemoveMsgBasedFeesProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Community Pool Spend Proposal:
  Title:       %s
  Description: %s
  Msg:         %s
`, rmbfp.Title, rmbfp.Description, rmbfp.Msg))
	return b.String()
}
