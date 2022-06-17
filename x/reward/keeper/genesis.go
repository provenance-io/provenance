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

	epochRewardDistributions := make([]types.EpochRewardDistribution, 0)
	epochRewardDistributionRecords := func(epochRewardDistribution types.EpochRewardDistribution) bool {
		epochRewardDistributions = append(epochRewardDistributions, epochRewardDistribution)
		return false
	}
	if err := k.IterateEpochRewardDistributions(ctx, epochRewardDistributionRecords); err != nil {
		panic(err)
	}

	actionDelegate, _ := k.GetActionDelegate(ctx)

	actionTransferDelegations, _ := k.GetActionTransferDelegations(ctx)

	return types.NewGenesisState(
		rewardProgramID,
		rewardPrograms,
		epochRewardDistributions,
		actionDelegate,
		actionTransferDelegations,
	)
}

// InitGenesis new reward genesis
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	if err := data.Validate(); err != nil {
		panic(err)
	}

	k.SetRewardProgramID(ctx, data.StartingRewardProgramId)

	for _, rewardProgram := range data.RewardPrograms {
		k.SetRewardProgram(ctx, rewardProgram)
	}

	for _, epochRewardDistributions := range data.EpochRewardDistributions {
		k.SetEpochRewardDistribution(ctx, epochRewardDistributions)
	}

	k.SetActionDelegate(ctx, data.ActionDelegate)
	k.SetActionTransferDelegations(ctx, data.ActionTransferDelegations)
}
