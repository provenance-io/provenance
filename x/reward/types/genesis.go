package types

func NewGenesisState(
	startingRewardProgramID uint64,
	rewardProgram []RewardProgram,
	epochRewardDistributions []EpochRewardDistribution,
	actionDelegate ActionDelegate,
	actionTransferDelegations ActionTransferDelegations,
) *GenesisState {
	return &GenesisState{
		StartingRewardProgramId:   startingRewardProgramID,
		RewardPrograms:            rewardProgram,
		EpochRewardDistributions:  epochRewardDistributions,
		ActionDelegate:            actionDelegate,
		ActionTransferDelegations: actionTransferDelegations,
	}
}

// DefaultGenesis returns the default reward genesis state
func DefaultGenesis() *GenesisState {
	return NewGenesisState(
		DefaultStartingRewardProgramID,
		[]RewardProgram{},
		[]EpochRewardDistribution{},
		ActionDelegate{},
		ActionTransferDelegations{},
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

	for _, epochRewardDistributions := range gs.EpochRewardDistributions {
		if err := epochRewardDistributions.ValidateBasic(); err != nil {
			return err
		}
	}

	actionDelegate := gs.ActionDelegate
	if err := actionDelegate.ValidateBasic(); err != nil {
		return err
	}

	actionTransferDelegations := gs.ActionTransferDelegations
	if err := actionTransferDelegations.ValidateBasic(); err != nil {
		return err
	}

	return nil
}
