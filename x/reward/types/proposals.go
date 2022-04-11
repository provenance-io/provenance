package types

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	rewardProgramId uint64,
	distributeFromAddress string,
	coin sdk.Coin,
	epochId string,
	epochStartOffset uint64,
	numberEpochs uint64,
	eligibilityCriteria EligibilityCriteria,
	minimum uint64,
	maximum uint64,
) *AddRewardProgramProposal {
	return &AddRewardProgramProposal{
		Title:                 title,
		Description:           description,
		RewardProgramId:       rewardProgramId,
		DistributeFromAddress: distributeFromAddress,
		Coin:                  coin,
		EpochId:               epochId,
		EpochStartOffset:      epochStartOffset,
		NumberEpochs:          numberEpochs,
		EligibilityCriteria:   eligibilityCriteria,
		Minimum:               minimum,
		Maximum:               maximum,
	}
}

func (arpp AddRewardProgramProposal) ProposalRoute() string { return RouterKey }

func (arpp AddRewardProgramProposal) ProposalType() string { return ProposalTypeAddRewardProgram }

func (arpp AddRewardProgramProposal) ValidateBasic() error {
	if arpp.RewardProgramId < 1 {
		return errors.New("reward program id is invalid")
	}
	if arpp.EpochStartOffset < 1 {
		return errors.New("reward program epoch start offset is invalid")
	}
	if _, err := sdk.AccAddressFromBech32(arpp.DistributeFromAddress); err != nil {
		return fmt.Errorf("invalid address for rewards program distribution from address: %w", err)
	}
	if len(arpp.EpochId) == 0 {
		return errors.New("epoch id cannot be empty")
	}
	if err := arpp.EligibilityCriteria.ValidateBasic(); err != nil {
		return fmt.Errorf("eligibility criteria is not valid: %w", err)
	}
	if !arpp.Coin.IsPositive() {
		return fmt.Errorf("reward program requires coins: %v", arpp.Coin)
	}
	if arpp.Maximum == 0 {
		return errors.New("maximum must be larger than 0")
	}
	if arpp.Minimum > arpp.Maximum {
		return fmt.Errorf("minimum (%v) cannot be larger than the maximum (%v)", arpp.Minimum, arpp.Maximum)
	}
	return nil
}

func (arpp AddRewardProgramProposal) String() string {
	out, _ := yaml.Marshal(arpp)
	return string(out)
}
