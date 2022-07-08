package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (k Keeper) ClaimRewards(ctx sdk.Context, rewardProgramID uint64, addr string) ([]*types.ClaimedRewardPeriodDetail, sdk.Coin, error) {
	rewardProgram, err := k.GetRewardProgram(ctx, rewardProgramID)
	if err != nil || rewardProgram.Validate() != nil {
		return nil, sdk.Coin{}, fmt.Errorf("reward program %d does not exist", rewardProgramID)
	}

	if rewardProgram.State == types.RewardProgram_EXPIRED {
		return nil, sdk.Coin{}, fmt.Errorf("reward program %d has expired", rewardProgramID)
	}

	rewards := k.claimRewardsForProgram(ctx, rewardProgram, addr)
	sent, err := k.sendRewards(ctx, rewardProgram, rewards, addr)
	if err != nil {
		// Rollback is handled by the chain automatically
		// k.rollbackClaims(ctx, rewardProgram, rewards, addr)
		return nil, sdk.Coin{}, err
	}
	rewardProgram.ClaimedAmount = rewardProgram.ClaimedAmount.Add(sent)
	k.SetRewardProgram(ctx, rewardProgram)

	return rewards, sent, nil
}

func (k Keeper) claimRewardsForProgram(ctx sdk.Context, rewardProgram types.RewardProgram, addr string) []*types.ClaimedRewardPeriodDetail {
	rewards := []*types.ClaimedRewardPeriodDetail{}

	for period := 1; period <= int(rewardProgram.CurrentClaimPeriod); period++ {
		reward, found := k.claimRewardForPeriod(ctx, rewardProgram, uint64(period), addr)
		if !found {
			continue
		}
		rewards = append(rewards, &reward)
	}

	return rewards
}

func (k Keeper) claimRewardForPeriod(ctx sdk.Context, rewardProgram types.RewardProgram, period uint64, addr string) (reward types.ClaimedRewardPeriodDetail, found bool) {
	state, err := k.GetRewardAccountState(ctx, rewardProgram.GetId(), period, addr)
	if err != nil {
		return reward, false
	}
	if state.GetClaimStatus() != types.RewardAccountState_CLAIMABLE {
		return reward, false
	}

	distribution, err := k.GetClaimPeriodRewardDistribution(ctx, period, rewardProgram.GetId())
	if err != nil {
		return reward, false
	}

	participantReward := k.CalculateParticipantReward(ctx, int64(state.GetSharesEarned()), distribution.GetTotalShares(), distribution.GetRewardsPool())
	reward = types.ClaimedRewardPeriodDetail{
		ClaimPeriodId:     period,
		TotalShares:       state.GetSharesEarned(),
		ClaimPeriodReward: participantReward,
	}

	state.ClaimStatus = types.RewardAccountState_CLAIMED
	k.SetRewardAccountState(ctx, state)

	return reward, true
}

func (k Keeper) sendRewards(ctx sdk.Context, rewardProgram types.RewardProgram, rewards []*types.ClaimedRewardPeriodDetail, addr string) (sdk.Coin, error) {
	amount := sdk.NewInt64Coin("nhash", 0)

	if len(rewards) == 0 {
		return amount, nil
	}

	for i := 0; i < len(rewards); i++ {
		reward := rewards[i]
		amount.Denom = reward.GetClaimPeriodReward().Denom
		amount = amount.Add(reward.GetClaimPeriodReward())
	}

	return k.sendCoinsToAccount(ctx, amount, addr)
}

func (k Keeper) sendCoinsToAccount(ctx sdk.Context, amount sdk.Coin, addr string) (sdk.Coin, error) {
	acc, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return sdk.NewInt64Coin(amount.Denom, 0), err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, acc, sdk.NewCoins(amount))
	if err != nil {
		return sdk.NewInt64Coin(amount.Denom, 0), err
	}

	return amount, nil
}

func (k Keeper) RefundRewardClaims(ctx sdk.Context, rewardProgram types.RewardProgram) error {
	amount := rewardProgram.TotalRewardPool.Sub(rewardProgram.RemainingPoolBalance).Sub(rewardProgram.ClaimedAmount)
	_, err := k.sendCoinsToAccount(ctx, amount, rewardProgram.GetDistributeFromAddress())
	return err
}
