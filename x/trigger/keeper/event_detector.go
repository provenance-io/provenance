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

// detectTransactionEvents Detects triggers that have been activated by transaction events.
func (k Keeper) detectTransactionEvents(ctx sdk.Context) (triggers []types.Trigger) {
	detectedTriggers := map[uint64]bool{}
	terminator := func(trigger types.Trigger, triggerEvent types.TriggerEventI) bool {
		return false
	}

	for _, event := range ctx.EventManager().GetABCIEventHistory() {
		matched := k.getMatchingTriggersUntil(ctx, event.GetType(), func(trigger types.Trigger, triggerEvent types.TriggerEventI) bool {
			if _, isDetected := detectedTriggers[trigger.Id]; isDetected {
				return false
			}
			txEvent := triggerEvent.(*types.TransactionEvent)
			detected := txEvent.Matches(event)
			detectedTriggers[trigger.Id] = detected
			return detected
		}, terminator)
		triggers = append(triggers, matched...)
	}
	return
}

// detectBlockHeightEvents Detects triggers that have been activated by block height events.
func (k Keeper) detectBlockHeightEvents(ctx sdk.Context) (triggers []types.Trigger) {
	match := func(_ types.Trigger, triggerEvent types.TriggerEventI) bool {
		blockHeightEvent := triggerEvent.(*types.BlockHeightEvent)
		return ctx.BlockHeight() >= int64(blockHeightEvent.GetBlockHeight())
	}
	terminator := func(_ types.Trigger, triggerEvent types.TriggerEventI) bool {
		blockHeightEvent := triggerEvent.(*types.BlockHeightEvent)
		return ctx.BlockHeight() < int64(blockHeightEvent.GetBlockHeight())
	}

	triggers = k.getMatchingTriggersUntil(ctx, types.BlockHeightPrefix, match, terminator)
	return
}

// detectTimeEvents Detects triggers that have been activated by block time events.
func (k Keeper) detectTimeEvents(ctx sdk.Context) (triggers []types.Trigger) {
	match := func(_ types.Trigger, triggerEvent types.TriggerEventI) bool {
		blockTimeEvent := triggerEvent.(*types.BlockTimeEvent)
		return ctx.BlockTime().UTC().Equal(blockTimeEvent.GetTime().UTC()) || ctx.BlockTime().UTC().After(blockTimeEvent.GetTime().UTC())
	}
	terminator := func(_ types.Trigger, triggerEvent types.TriggerEventI) bool {
		blockTimeEvent := triggerEvent.(*types.BlockTimeEvent)
		return ctx.BlockTime().UTC().Before(blockTimeEvent.GetTime().UTC())
	}

	triggers = k.getMatchingTriggersUntil(ctx, types.BlockTimePrefix, match, terminator)
	return
}

// getMatchingTriggersUntil Gets the triggers with a specified prefix that are ready to be activated and fulfill the given condition until a specific ending condition is reached.
func (k Keeper) getMatchingTriggersUntil(ctx sdk.Context, prefix string, match func(types.Trigger, types.TriggerEventI) bool, terminator func(types.Trigger, types.TriggerEventI) bool) (triggers []types.Trigger) {
	err := k.IterateEventListeners(ctx, prefix, func(trigger types.Trigger) (stop bool, err error) {
		event, _ := trigger.GetTriggerEventI()
		if match(trigger, event) {
			triggers = append(triggers, trigger)
		}
		return terminator(trigger, event), nil
	})
	if err != nil {
		panic(fmt.Errorf("unable to iterate event listeners for matching triggers: %w", err))
	}
	return
}
