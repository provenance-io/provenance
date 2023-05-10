package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/trigger/types"
)

func (k Keeper) DetectBlockEvents(ctx sdk.Context) {
	k.DetectTransactionEvents(ctx)
	k.DetectBlockHeightEvents(ctx)
	k.DetectTimeEvents(ctx)
}

func (k Keeper) DetectTransactionEvents(ctx sdk.Context) {
	triggers := []types.Trigger{}

	for _, event := range ctx.EventManager().GetABCIEventHistory() {
		matched, err := k.GetMatchingTriggers(ctx, event.GetType(), func(triggerEvent types.TriggerEventI) bool {
			txEvent := triggerEvent.(*types.TransactionEvent)
			return txEvent.Equals(event)
		})
		if err != nil {
			// TODO What to do here? There has to be something bad
			return
		}
		triggers = append(triggers, matched...)
	}

	for _, trigger := range triggers {
		k.QueueTrigger(ctx, trigger)
	}
}

func (k Keeper) DetectBlockHeightEvents(ctx sdk.Context) {
	triggers, err := k.GetMatchingTriggers(ctx, "block-height", func(triggerEvent types.TriggerEventI) bool {
		blockHeightEvent := triggerEvent.(*types.BlockHeightEvent)
		return ctx.BlockHeight() >= int64(blockHeightEvent.GetBlockHeight())
	})
	if err != nil {
		// TODO What to do here? There has to be something bad
		return
	}

	for _, trigger := range triggers {
		k.QueueTrigger(ctx, trigger)
	}
}

func (k Keeper) DetectTimeEvents(ctx sdk.Context) {
	triggers, err := k.GetMatchingTriggers(ctx, "block-time", func(triggerEvent types.TriggerEventI) bool {
		blockTimeEvent := triggerEvent.(*types.BlockTimeEvent)
		return ctx.BlockTime().Equal(blockTimeEvent.GetTime()) || ctx.BlockTime().After(blockTimeEvent.GetTime())
	})
	if err != nil {
		// TODO What to do here? There has to be something bad
		return
	}

	for _, trigger := range triggers {
		k.QueueTrigger(ctx, trigger)
	}
}

func (k Keeper) GetMatchingTriggers(ctx sdk.Context, prefix string, condition func(types.TriggerEventI) bool) (triggers []types.Trigger, err error) {
	err = k.IterateEventListeners(ctx, prefix, func(trigger types.Trigger) (stop bool, err error) {
		event := trigger.Event.GetCachedValue().(types.TriggerEventI)

		if condition(event) {
			triggers = append(triggers, trigger)
		}

		return false, nil
	})
	if err != nil {
		// TODO Return error
		return triggers, err
	}
	return triggers, err
}
