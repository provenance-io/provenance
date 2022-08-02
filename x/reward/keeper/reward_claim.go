package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (k Keeper) ClaimRewards(ctx sdk.Context, rewardProgramID uint64, addr string) ([]*types.ClaimedRewardPeriodDetail, sdk.Coin, error) {
	rewardProgram, err := k.GetRewardProgram(ctx, rewardProgramID)
	if err != nil || rewardProgram.Validate() != nil {
		return nil, sdk.Coin{}, fmt.Errorf("reward program %d does not exist", rewardProgramID)
	}

	if rewardProgram.State == types.RewardProgram_STATE_EXPIRED {
		return nil, sdk.Coin{}, fmt.Errorf("reward program %d has expired", rewardProgramID)
	}

	rewards, err := k.claimRewardsForProgram(ctx, rewardProgram, addr)
	if err != nil {
		return nil, sdk.Coin{}, err
	}
	sent, err := k.sendRewards(ctx, rewards, addr)
	if err != nil {
		// Rollback is handled by the chain automatically
		return nil, sdk.Coin{}, err
	}
	rewardProgram.ClaimedAmount = rewardProgram.ClaimedAmount.Add(sent)
	k.SetRewardProgram(ctx, rewardProgram)

	return rewards, sent, nil
}

func (k Keeper) claimRewardsForProgram(ctx sdk.Context, rewardProgram types.RewardProgram, addr string) ([]*types.ClaimedRewardPeriodDetail, error) {
	var states []types.RewardAccountState
	address, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}
	k.IterateRewardAccountStatesByAddressAndRewardsID(ctx, address, rewardProgram.GetId(), func(state types.RewardAccountState) bool {
		if state.GetSharesEarned() > 0 && state.Address == address.String() {
			states = append(states, state)
		}
		return false
	})

	rewards := make([]*types.ClaimedRewardPeriodDetail, 0, len(states))
	for _, account := range states {
		reward, found := k.claimRewardForPeriod(ctx, rewardProgram, account.ClaimPeriodId, addr)
		if !found {
			continue
		}
		rewards = append(rewards, &reward)
	}
	return rewards, nil
}

func (k Keeper) claimRewardForPeriod(ctx sdk.Context, rewardProgram types.RewardProgram, period uint64, addr string) (reward types.ClaimedRewardPeriodDetail, found bool) {
	state, err := k.GetRewardAccountState(ctx, rewardProgram.GetId(), period, addr)
	if err != nil {
		return reward, false
	}
	if state.GetClaimStatus() != types.RewardAccountState_CLAIM_STATUS_CLAIMABLE {
		return reward, false
	}

	distribution, err := k.GetClaimPeriodRewardDistribution(ctx, period, rewardProgram.GetId())
	if err != nil {
		return reward, false
	}

	participantReward := k.CalculateParticipantReward(ctx, int64(state.GetSharesEarned()), distribution.GetTotalShares(), distribution.GetRewardsPool(), rewardProgram.MaxRewardByAddress)
	reward = types.ClaimedRewardPeriodDetail{
		ClaimPeriodId:     period,
		TotalShares:       state.GetSharesEarned(),
		ClaimPeriodReward: participantReward,
	}

	state.ClaimStatus = types.RewardAccountState_CLAIM_STATUS_CLAIMED
	k.SetRewardAccountState(ctx, state)

	return reward, true
}

func (k Keeper) sendRewards(ctx sdk.Context, rewards []*types.ClaimedRewardPeriodDetail, addr string) (sdk.Coin, error) {
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
	if amount.IsZero() {
		return sdk.NewInt64Coin(amount.GetDenom(), 0), nil
	}

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

func (k Keeper) ClaimAllRewards(ctx sdk.Context, addr string) ([]*types.RewardProgramClaimDetail, sdk.Coin, error) {
	var allProgramDetails []*types.RewardProgramClaimDetail
	allRewards := sdk.NewInt64Coin("nhash", 0)
	err := k.IterateRewardPrograms(ctx, func(rewardProgram types.RewardProgram) (stop bool) {
		details, reward, err := k.ClaimRewards(ctx, rewardProgram.GetId(), addr)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("NOTICE: Unable to claim reward program %d. Error: %v ", rewardProgram.GetId(), err))
			return false
		}
		if reward.IsZero() {
			ctx.Logger().Info(fmt.Sprintf("NOTICE: Skipping reward program %d. It has no rewards.", rewardProgram.GetId()))
			return false
		}

		programDetails := types.RewardProgramClaimDetail{
			RewardProgramId:            rewardProgram.GetId(),
			TotalRewardClaim:           reward,
			ClaimedRewardPeriodDetails: details,
		}
		allProgramDetails = append(allProgramDetails, &programDetails)
		allRewards = allRewards.Add(reward)

		return false
	})
	if err != nil {
		return nil, sdk.Coin{}, err
	}

	return allProgramDetails, allRewards, nil
}
