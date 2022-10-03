package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/provenance-io/provenance/x/reward/types"
)

// ClaimRewards for a given address and a given reward program id
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
	sent, err := k.sendRewards(ctx, rewards, addr, rewardProgram.GetTotalRewardPool().Denom)
	if err != nil {
		return nil, sdk.Coin{}, err
	}
	rewardProgram.ClaimedAmount = rewardProgram.ClaimedAmount.Add(sent)
	k.SetRewardProgram(ctx, rewardProgram)

	return rewards, sent, nil
}

// claimRewardsForProgram internal method used by ClaimRewards, which iterates over all the reward account states that the
// address is eligible for, and then claim them.
func (k Keeper) claimRewardsForProgram(ctx sdk.Context, rewardProgram types.RewardProgram, addr string) ([]*types.ClaimedRewardPeriodDetail, error) {
	var states []types.RewardAccountState
	address, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrap(err.Error())
	}
	err = k.IterateRewardAccountStatesByAddressAndRewardsID(ctx, address, rewardProgram.GetId(), func(state types.RewardAccountState) bool {
		if state.GetSharesEarned() > 0 && state.Address == address.String() {
			states = append(states, state)
		}
		return false
	})
	if err != nil {
		return nil, err
	}

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

// claimRewardForPeriod internal method to actually claim rewards for a period.
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

// sendRewards internal method called with ClaimedRewardPeriodDetail of a single reward program
func (k Keeper) sendRewards(ctx sdk.Context, rewards []*types.ClaimedRewardPeriodDetail, addr string, rewardProgramDenom string) (sdk.Coin, error) {
	amount := sdk.NewInt64Coin(rewardProgramDenom, 0)

	if len(rewards) == 0 {
		return amount, nil
	}

	for _, reward := range rewards {
		amount.Denom = reward.GetClaimPeriodReward().Denom
		amount = amount.Add(reward.GetClaimPeriodReward())
	}

	return k.sendCoinsToAccount(ctx, amount, addr)
}

// sendCoinsToAccount internal wrapper method, to mainly do `SendCoinsFromModuleToAccount`
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

// RefundRewardClaims refund all unclaimed rewards to the reward program creator
func (k Keeper) RefundRewardClaims(ctx sdk.Context, rewardProgram types.RewardProgram) error {
	amount := rewardProgram.TotalRewardPool.Sub(rewardProgram.RemainingPoolBalance).Sub(rewardProgram.ClaimedAmount)
	_, err := k.sendCoinsToAccount(ctx, amount, rewardProgram.GetDistributeFromAddress())
	return err
}

// ClaimAllRewards calls ClaimRewards, however differs from ClaimRewards in that it claims all the rewards that the address
// is eligible for across all reward programs.
func (k Keeper) ClaimAllRewards(ctx sdk.Context, addr string) ([]*types.RewardProgramClaimDetail, sdk.Coins, error) {
	allProgramDetails := []*types.RewardProgramClaimDetail{}
	allRewards := sdk.Coins{}

	programs, err := k.GetAllUnexpiredRewardPrograms(ctx)
	if err != nil {
		return nil, sdk.Coins{}, err
	}

	for _, rewardProgram := range programs {
		details, reward, err := k.ClaimRewards(ctx, rewardProgram.GetId(), addr)
		// err needs to propagated up, else tx will commit
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("Unable to claim reward program %d. Error: %v ", rewardProgram.GetId(), err))
			return nil, sdk.Coins{}, err
		}
		if reward.IsZero() {
			ctx.Logger().Info(fmt.Sprintf("Skipping reward program %d. It has no rewards.", rewardProgram.GetId()))
			continue
		}

		programDetails := types.RewardProgramClaimDetail{
			RewardProgramId:            rewardProgram.GetId(),
			TotalRewardClaim:           reward,
			ClaimedRewardPeriodDetails: details,
		}
		allProgramDetails = append(allProgramDetails, &programDetails)
		allRewards = allRewards.Add(reward)
	}

	return allProgramDetails, allRewards, nil
}
