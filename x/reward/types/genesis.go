package types

func NewGenesisState(rewardProgram []RewardProgram,
	rewardClaim []RewardClaim,
	epochRewardDistribution []EpochRewardDistribution,
	eligibilityCriteria []EligibilityCriteria,
	actionDelegate ActionDelegate,
	actionTransferDelegations ActionTransferDelegations,
) *GenesisState {
	return &GenesisState{
		RewardPrograms:            rewardProgram,
		RewardClaims:              rewardClaim,
		EpochRewardDistributions:  epochRewardDistribution,
		EligibilityCriterias:      eligibilityCriteria,
		ActionDelegate:            actionDelegate,
		ActionTransferDelegations: actionTransferDelegations,
	}
}

// DefaultGenesis returns the default reward genesis state
func DefaultGenesis() *GenesisState {
	return NewGenesisState(
		[]RewardProgram{},
		[]RewardClaim{},
		[]EpochRewardDistribution{},
		[]EligibilityCriteria{},
		ActionDelegate{},
		ActionTransferDelegations{},
	)
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	return nil
}
