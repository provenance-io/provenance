package types

func NewGenesisState(
	startingRewardProgramID uint64,
	rewardProgram []RewardProgram,
	claimPeriodRewardDistributions []ClaimPeriodRewardDistribution,
) *GenesisState {
	return &GenesisState{
		StartingRewardProgramId:        startingRewardProgramID,
		RewardPrograms:                 rewardProgram,
		ClaimPeriodRewardDistributions: claimPeriodRewardDistributions,
	}
}

// DefaultGenesis returns the default reward genesis state
func DefaultGenesis() *GenesisState {
	return NewGenesisState(
		DefaultStartingRewardProgramID,
		[]RewardProgram{},
		[]ClaimPeriodRewardDistribution{},
	)
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for _, rewardProgram := range gs.RewardPrograms {
		if err := rewardProgram.ValidateBasic(); err != nil {
			return err
		}
	}

	for _, claimPeriodRewardDistributions := range gs.ClaimPeriodRewardDistributions {
		if err := claimPeriodRewardDistributions.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}
