package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/trigger/types"
)

// DetectBlockEvents Detects triggers that have been activated by their corresponding events.
func (k Keeper) DetectBlockEvents(ctx sdk.Context) {
	triggers := k.DetectTransactionEvents(ctx)
	triggers = append(triggers, k.DetectBlockHeightEvents(ctx)...)
	triggers = append(triggers, k.DetectTimeEvents(ctx)...)

	for _, trigger := range triggers {
		k.UnregisterTrigger(ctx, trigger)
		k.QueueTrigger(ctx, trigger)
	}
}

// DetectTransactionEvents Detects triggers that have been activated by transaction events.
func (k Keeper) DetectTransactionEvents(ctx sdk.Context) (triggers []types.Trigger) {
	for _, event := range ctx.EventManager().GetABCIEventHistory() {
		matched := k.GetMatchingTriggers(ctx, event.GetType(), func(triggerEvent types.TriggerEventI) bool {
			txEvent := triggerEvent.(*types.TransactionEvent)
			return txEvent.Equals(event)
		})
		triggers = append(triggers, matched...)
	}
	return
}

// DetectBlockHeightEvents Detects triggers that have been activated by block height events.
func (k Keeper) DetectBlockHeightEvents(ctx sdk.Context) (triggers []types.Trigger) {
	triggers = k.GetMatchingTriggers(ctx, types.BlockHeightPrefix, func(triggerEvent types.TriggerEventI) bool {
		blockHeightEvent := triggerEvent.(*types.BlockHeightEvent)
		return ctx.BlockHeight() >= int64(blockHeightEvent.GetBlockHeight())
	})
	return
}

// DetectTimeEvents Detects triggers that have been activated by block time events.
func (k Keeper) DetectTimeEvents(ctx sdk.Context) (triggers []types.Trigger) {
	triggers = k.GetMatchingTriggers(ctx, types.BlockTimePrefix, func(triggerEvent types.TriggerEventI) bool {
		blockTimeEvent := triggerEvent.(*types.BlockTimeEvent)
		return ctx.BlockTime().Equal(blockTimeEvent.GetTime()) || ctx.BlockTime().After(blockTimeEvent.GetTime())
	})
	return
}

// GetMatchingTriggers Obtains the prefixed triggers that are waiting to be activated and match the supplied condition.
func (k Keeper) GetMatchingTriggers(ctx sdk.Context, prefix string, condition func(types.TriggerEventI) bool) (triggers []types.Trigger) {
	err := k.IterateEventListeners(ctx, prefix, func(trigger types.Trigger) (stop bool, err error) {
		event, _ := trigger.GetTriggerEventI()
		if condition(event) {
			triggers = append(triggers, trigger)
		}
		return false, nil
	})
	if err != nil {
		panic("unable to iterate event listeners for matching triggers")
	}
	return
}
