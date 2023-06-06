package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/trigger/types"
)

// DetectBlockEvents Detects triggers that have been activated by their corresponding events.
func (k Keeper) DetectBlockEvents(ctx sdk.Context) {
	triggers := k.detectTransactionEvents(ctx)
	triggers = append(triggers, k.detectBlockHeightEvents(ctx)...)
	triggers = append(triggers, k.detectTimeEvents(ctx)...)

	for _, trigger := range triggers {
		k.UnregisterTrigger(ctx, trigger)
		k.QueueTrigger(ctx, trigger)
	}
}

// DetectTransactionEvents Detects triggers that have been activated by transaction events.
func (k Keeper) detectTransactionEvents(ctx sdk.Context) (triggers []types.Trigger) {
	detectedTriggers := map[uint64]bool{}
	for _, event := range ctx.EventManager().GetABCIEventHistory() {
		matched := k.getMatchingTriggers(ctx, event.GetType(), func(trigger types.Trigger, triggerEvent types.TriggerEventI) bool {
			if _, isDetected := detectedTriggers[trigger.Id]; isDetected {
				return false
			}
			txEvent := triggerEvent.(*types.TransactionEvent)
			detected := txEvent.Equals(event)
			detectedTriggers[trigger.Id] = detected
			return detected
		})
		triggers = append(triggers, matched...)
	}
	return
}

// DetectBlockHeightEvents Detects triggers that have been activated by block height events.
func (k Keeper) detectBlockHeightEvents(ctx sdk.Context) (triggers []types.Trigger) {
	triggers = k.getMatchingTriggers(ctx, types.BlockHeightPrefix, func(_ types.Trigger, triggerEvent types.TriggerEventI) bool {
		blockHeightEvent := triggerEvent.(*types.BlockHeightEvent)
		return ctx.BlockHeight() >= int64(blockHeightEvent.GetBlockHeight())
	})
	return
}

// DetectTimeEvents Detects triggers that have been activated by block time events.
func (k Keeper) detectTimeEvents(ctx sdk.Context) (triggers []types.Trigger) {
	triggers = k.getMatchingTriggers(ctx, types.BlockTimePrefix, func(_ types.Trigger, triggerEvent types.TriggerEventI) bool {
		blockTimeEvent := triggerEvent.(*types.BlockTimeEvent)
		return ctx.BlockTime().UTC().Equal(blockTimeEvent.GetTime().UTC()) || ctx.BlockTime().UTC().After(blockTimeEvent.GetTime().UTC())
	})
	return
}

// GetMatchingTriggers Obtains the prefixed triggers that are waiting to be activated and match the supplied condition.
func (k Keeper) getMatchingTriggers(ctx sdk.Context, prefix string, condition func(types.Trigger, types.TriggerEventI) bool) (triggers []types.Trigger) {
	err := k.IterateEventListeners(ctx, prefix, func(trigger types.Trigger) (stop bool, err error) {
		event, _ := trigger.GetTriggerEventI()
		if condition(trigger, event) {
			triggers = append(triggers, trigger)
		}
		return false, nil
	})
	if err != nil {
		panic(fmt.Errorf("unable to iterate event listeners for matching triggers"))
	}
	return
}
