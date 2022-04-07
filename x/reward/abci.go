package reward

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/reward/keeper"
)

// EndBlocker called every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	ctx.Logger().Info(fmt.Sprintf("In endblocker"))

	// check if epoch has ended
	ctx.Logger().Info(fmt.Sprintf("Size of events is %d", len(ctx.EventManager().GetABCIEventHistory())))
	//logEvents(ctx)
	// Step 1
	rewardPrograms, err := k.GetAllActiveRewards(ctx)
	if err != nil {
		// TODO log it imo..we don't want blockchain to stop?
		ctx.Logger().Error(err.Error())
	}

	// only rewards programs who are eligible will be iterated through here
	// Step 2
	for _, rewardProgram := range rewardPrograms {
		epochRewardDistributionForEpoch, err := k.GetEpochRewardDistribution(ctx, rewardProgram.EpochId, rewardProgram.Id)
		if err != nil {
			continue
		}
		currentEpoch := k.EpochKeeper.GetEpochInfo(ctx, rewardProgram.EpochId)
		// epoch reward distribution does it exist till the block has ended, highly unlikely but could happen
		if epochRewardDistributionForEpoch.EpochId == "" {
			epochRewardDistributionForEpoch.EpochId = rewardProgram.EpochId
			epochRewardDistributionForEpoch.RewardProgramId = rewardProgram.Id
			epochRewardDistributionForEpoch.TotalShares = 0
			epochRewardDistributionForEpoch.EpochEnded = false
			k.EvaluateRules(ctx, currentEpoch.CurrentEpoch, rewardProgram, epochRewardDistributionForEpoch)
		} else if epochRewardDistributionForEpoch.EpochEnded == false { // if hook epoch end has already been called, this should not get called.
			// end the epoch
			k.EvaluateRules(ctx, currentEpoch.CurrentEpoch, rewardProgram, epochRewardDistributionForEpoch)
		}
	}
}

// this method is only for testing
func logEvents(ctx sdk.Context) {
	// check if epoch has ended
	for i, s := range ctx.EventManager().GetABCIEventHistory() {
		ctx.Logger().Info(fmt.Sprintf("events in end blocker %d", i))
		ctx.Logger().Info(fmt.Sprintf("events attributes are %s", s.Type))
		ctx.Logger().Info(fmt.Sprintf("Size of events is %d", len(ctx.EventManager().GetABCIEventHistory())))
		for _, s := range ctx.EventManager().GetABCIEventHistory() {
			ctx.Logger().Info(fmt.Sprintf("events type is %s", s.Type))
			for _, y := range s.Attributes {
				ctx.Logger().Info(fmt.Sprintf("event attribute key:events attribute value  %s:%s", y.Key, y.Value))
				ctx.Logger().Info(fmt.Sprintf("event attribute is %s attribute_key:attribute_value  %s:%s", s.Type, y.Key, y.Value))
				//4:24PM INF events type is coin_spent
				//4:24PM INF event attribute is coin_spent attribute_key:attribute_value  spender:tp1sha7e07l5knw4vdw2vgc3k06gd0fscz9r32yv6
				//4:24PM INF event attribute is coin_spent attribute_key:attribute_value  amount:76200000000000nhash
				//4:24PM INF events type is coin_received
				//4:24PM INF event attribute is coin_received attribute_key:attribute_value  receiver:tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt
				//4:24PM INF event attribute is coin_received attribute_key:attribute_value  amount:76200000000000nhash
				//4:24PM INF events type is transfer
				//4:24PM INF event attribute is transfer attribute_key:attribute_value  recipient:tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt
				//4:24PM INF event attribute is transfer attribute_key:attribute_value  sender:tp1sha7e07l5knw4vdw2vgc3k06gd0fscz9r32yv6
				//4:24PM INF event attribute is transfer attribute_key:attribute_value  amount:76200000000000nhash

			}
		}

	}
}
