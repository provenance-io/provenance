package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

// ExportGenesis returns a GenesisState for a given context.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	rewardProgramId, err := k.GetRewardProgramID(ctx)
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

	rewardClaims := make([]types.RewardClaim, 0)
	rewardClaimRecords := func(rewardClaim types.RewardClaim) bool {
		rewardClaims = append(rewardClaims, rewardClaim)
		return false
	}
	if err := k.IterateRewardClaims(ctx, rewardClaimRecords); err != nil {
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

	eligibilityCriterias := make([]types.EligibilityCriteria, 0)
	eligibilityCriteriaRecords := func(eligibilityCriteria types.EligibilityCriteria) bool {
		eligibilityCriterias = append(eligibilityCriterias, eligibilityCriteria)
		return false
	}
	if err := k.IterateEligibilityCriterias(ctx, eligibilityCriteriaRecords); err != nil {
		panic(err)
	}

	actionDelegate, _ := k.GetActionDelegate(ctx)

	actionTransferDelegations, _ := k.GetActionTransferDelegations(ctx)

	return types.NewGenesisState(
		rewardProgramId,
		rewardPrograms,
		rewardClaims,
		epochRewardDistributions,
		eligibilityCriterias,
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

	for _, rewardClaims := range data.RewardClaims {
		k.SetRewardClaim(ctx, rewardClaims)
	}

	for _, epochRewardDistributions := range data.EpochRewardDistributions {
		k.SetEpochRewardDistribution(ctx, epochRewardDistributions)
	}

	for _, eligibilityCriteria := range data.EligibilityCriterias {
		k.SetEligibilityCriteria(ctx, eligibilityCriteria)
	}

	k.SetActionDelegate(ctx, data.ActionDelegate)
	k.SetActionTransferDelegations(ctx, data.ActionTransferDelegations)
}
