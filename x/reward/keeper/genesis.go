package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

// ExportGenesis returns a GenesisState for a given context.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	rewardProgramID, err := k.GetRewardProgramID(ctx)
	if err != nil {
		panic(err)
	}

	rewardPrograms := make([]types.RewardProgram, 0)
	rewardProgramRecords := func(rewardProgram types.RewardProgram) bool {
		rewardPrograms = append(rewardPrograms, rewardProgram)
		return false
	}
	if err := k.IterateRewardPrograms(ctx, rewardProgramRecords); err != nil {
		panic(err)
	}

	claimPeriodRewardDistributions := make([]types.ClaimPeriodRewardDistribution, 0)
	claimPeriodRewardDistributionRecords := func(ClaimPeriodRewardDistribution types.ClaimPeriodRewardDistribution) bool {
		claimPeriodRewardDistributions = append(claimPeriodRewardDistributions, ClaimPeriodRewardDistribution)
		return false
	}
	if err := k.IterateClaimPeriodRewardDistributions(ctx, claimPeriodRewardDistributionRecords); err != nil {
		panic(err)
	}

	rewardAccountStates := make([]types.RewardAccountState, 0)
	rewardAccountStateRecords := func(RewardAccountState types.RewardAccountState) bool {
		rewardAccountStates = append(rewardAccountStates, RewardAccountState)
		return false
	}

	for _, claim := range claimPeriodRewardDistributions {
		if err := k.IterateRewardAccountStates(ctx, claim.RewardProgramId, claim.ClaimPeriodId, rewardAccountStateRecords); err != nil {
			panic(err)
		}
	}

	return types.NewGenesisState(
		rewardProgramID,
		rewardPrograms,
		claimPeriodRewardDistributions,
		rewardAccountStates,
	)
}

// InitGenesis new reward genesis
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	if err := data.Validate(); err != nil {
		panic(err)
	}
	k.setRewardProgramID(ctx, data.RewardProgramId)

	for _, rewardProgram := range data.RewardPrograms {
		k.SetRewardProgram(ctx, rewardProgram)
	}

	for _, ClaimPeriodRewardDistributions := range data.ClaimPeriodRewardDistributions {
		k.SetClaimPeriodRewardDistribution(ctx, ClaimPeriodRewardDistributions)
	}

	for _, RewardAccountStates := range data.RewardAccountStates {
		k.SetRewardAccountState(ctx, RewardAccountStates)
	}
}
