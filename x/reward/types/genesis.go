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
	for _, rewardProgram := range gs.RewardPrograms {
		if err := rewardProgram.ValidateBasic(); err != nil {
			return err
		}
	}

	for _, rewardClaims := range gs.RewardClaims {
		if err := rewardClaims.ValidateBasic(); err != nil {
			return err
		}
	}

	for _, epochRewardDistributions := range gs.EpochRewardDistributions {
		if err := epochRewardDistributions.ValidateBasic(); err != nil {
			return err
		}
	}

	for _, eligibilityCriteria := range gs.EligibilityCriterias {
		if err := eligibilityCriteria.ValidateBasic(); err != nil {
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
