package reward

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/reward/keeper"
)

// EndBlocker called every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) error {
	ctx.Logger().Info(fmt.Sprintf("In endblocker"))

	// check if epoch has ended
	ctx.Logger().Info(fmt.Sprintf("Size of events is %d", len(ctx.EventManager().GetABCIEventHistory())))
	rewardPrograms, err := k.GetAllActiveRewards(ctx)
	if err != nil {
		return err
	}

	// only rewards programs who are eligible will be iterated through here
	for _, rewardProgram := range rewardPrograms {
		epochRewardDistibutionForEpoch, err := k.GetEpochRewardDistribution(ctx, rewardProgram.EpochId, rewardProgram.Id)
		if err != nil {
			return err
		}
		currentEpoch := k.EpochKeeper.GetEpochInfo(ctx,rewardProgram.EpochId)
		// epoch reward distribution does it exist till the block has ended, highly unlikely but could happen
		if epochRewardDistibutionForEpoch.EpochId == "" {
			epochRewardDistibutionForEpoch.EpochId = rewardProgram.EpochId
			epochRewardDistibutionForEpoch.RewardProgramId = rewardProgram.Id
			epochRewardDistibutionForEpoch.TotalShares = 0
			epochRewardDistibutionForEpoch.EpochEnded = false
			k.EvaluateRules(ctx, currentEpoch.CurrentEpoch, rewardProgram, *epochRewardDistibutionForEpoch)
		} else if epochRewardDistibutionForEpoch.EpochEnded == false { // if hook epoch end has already been called, this should not get called.
			// end the epoch
			epochRewardDistibutionForEpoch.EpochEnded = false
			k.EvaluateRules(ctx, currentEpoch.CurrentEpoch, rewardProgram, *epochRewardDistibutionForEpoch)
		}
	}
	return nil
}
