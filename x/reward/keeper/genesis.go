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

	ClaimPeriodRewardDistributions := make([]types.ClaimPeriodRewardDistribution, 0)
	ClaimPeriodRewardDistributionRecords := func(ClaimPeriodRewardDistribution types.ClaimPeriodRewardDistribution) bool {
		ClaimPeriodRewardDistributions = append(ClaimPeriodRewardDistributions, ClaimPeriodRewardDistribution)
		return false
	}
	if err := k.IterateClaimPeriodRewardDistributions(ctx, ClaimPeriodRewardDistributionRecords); err != nil {
		panic(err)
	}

	return types.NewGenesisState(
		rewardProgramID,
		rewardPrograms,
		ClaimPeriodRewardDistributions,
	)
}

// InitGenesis new reward genesis
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	if err := data.Validate(); err != nil {
		panic(err)
	}

	k.setRewardProgramID(ctx, data.StartingRewardProgramId)

	for _, rewardProgram := range data.RewardPrograms {
		k.SetRewardProgram(ctx, rewardProgram)
	}

	for _, ClaimPeriodRewardDistributions := range data.ClaimPeriodRewardDistributions {
		k.SetClaimPeriodRewardDistribution(ctx, ClaimPeriodRewardDistributions)
	}
}
