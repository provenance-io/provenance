package types

import "fmt"

func NewGenesisState(
	rewardProgramID uint64,
	rewardProgram []RewardProgram,
	claimPeriodRewardDistributions []ClaimPeriodRewardDistribution,
	rewardAccountStates []RewardAccountState,
) *GenesisState {
	return &GenesisState{
		RewardProgramId:                rewardProgramID,
		RewardPrograms:                 rewardProgram,
		ClaimPeriodRewardDistributions: claimPeriodRewardDistributions,
		RewardAccountStates:            rewardAccountStates,
	}
}

// DefaultGenesis returns the default reward genesis state
func DefaultGenesis() *GenesisState {
	return NewGenesisState(
		uint64(1),
		[]RewardProgram{},
		[]ClaimPeriodRewardDistribution{},
		[]RewardAccountState{},
	)
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for _, rewardProgram := range gs.RewardPrograms {
		if rewardProgram.Id >= gs.RewardProgramId {
			return fmt.Errorf("reward program id (%v) must not equal or be less than a current reward program id (%v)", gs.RewardProgramId, rewardProgram.Id)
		}
		if err := rewardProgram.Validate(); err != nil {
			return err
		}
	}

	for _, claimPeriodRewardDistributions := range gs.ClaimPeriodRewardDistributions {
		if err := claimPeriodRewardDistributions.Validate(); err != nil {
			return err
		}
	}

	for _, rewardsAccountStates := range gs.RewardAccountStates {
		if err := rewardsAccountStates.Validate(); err != nil {
			return err
		}
	}

	return nil
}
