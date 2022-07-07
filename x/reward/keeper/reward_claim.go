package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (k Keeper) ClaimRewards(ctx sdk.Context, rewardProgram types.RewardProgram, addr string) []types.ClaimedRewardPeriodDetail {
	claimedRewards := []types.ClaimedRewardPeriodDetail{}

	// Go through every claim period
	for period := 1; period <= int(rewardProgram.CurrentClaimPeriod); period++ {
		state, err := k.GetRewardAccountState(ctx, rewardProgram.GetId(), uint64(period), addr)
		if err != nil {
			// TODO How to handle error
			continue
		}
		if state.GetClaimStatus() != types.RewardAccountState_CLAIMABLE {
			continue
		}

		distribution, err := k.GetClaimPeriodRewardDistribution(ctx, uint64(period), rewardProgram.GetId())
		if err != nil {
			// Log that there is no reward distribution for this claim period
			continue
		}

		reward := k.CalculateParticipantReward(ctx, int64(state.GetSharesEarned()), distribution.GetTotalShares(), distribution.GetRewardsPool())
		claimedReward := types.ClaimedRewardPeriodDetail{
			ClaimPeriodId:     uint64(period),
			TotalShares:       state.GetSharesEarned(),
			ClaimPeriodReward: reward,
		}
		claimedRewards = append(claimedRewards, claimedReward)

		state.ClaimStatus = types.RewardAccountState_CLAIMED
		k.SetRewardAccountState(ctx, state)
	}

	return claimedRewards
}
