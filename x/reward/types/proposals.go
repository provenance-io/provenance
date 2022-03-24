package types

import (
	"gopkg.in/yaml.v2"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// ProposalTypeAddRewardProgram to add a new rewards program
	ProposalTypeAddRewardProgram string = "AddRewardProgram"
)

var (
	_ govtypes.Content = &AddRewardProgramProposal{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeAddRewardProgram)
	govtypes.RegisterProposalTypeCodec(AddRewardProgramProposal{}, "provenance/msgfees/AddRewardProgramProposal")
}

func NewAddRewardProgramProposal(
	title string,
	description string,
	rewardProgram RewardProgram) *AddRewardProgramProposal {
	return &AddRewardProgramProposal{
		Title:         title,
		Description:   description,
		RewardProgram: &rewardProgram,
	}
}

func (arpp AddRewardProgramProposal) ProposalRoute() string { return RouterKey }

func (arpp AddRewardProgramProposal) ProposalType() string { return ProposalTypeAddRewardProgram }

func (arpp AddRewardProgramProposal) ValidateBasic() error {
	return arpp.RewardProgram.ValidateBasic()
}

func (arpp AddRewardProgramProposal) String() string {
	out, _ := yaml.Marshal(arpp)
	return string(out)
}
