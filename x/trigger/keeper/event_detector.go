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
		k.Logger(ctx).Debug(fmt.Sprintf("Trigger %d added to queue", trigger.Id))
		k.emitTriggerDetected(ctx, trigger)
		k.UnregisterTrigger(ctx, trigger)
		k.QueueTrigger(ctx, trigger)
	}
}

// detectTransactionEvents Detects triggers that have been activated by transaction events.
func (k Keeper) detectTransactionEvents(ctx sdk.Context) (triggers []types.Trigger) {
	detectedTriggers := map[uint64]bool{}
	terminator := func(_ types.Trigger, _ types.TriggerEventI) bool {
		return false
	}

	abciEventHistory, ok := ctx.EventManager().(sdk.EventManagerWithHistoryI)
	if !ok {
		panic("event manager does not implement EventManagerWithHistoryI")
	}

	for _, event := range abciEventHistory.GetABCIEventHistory() {
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
		curHeight := uint64(ctx.BlockHeight())
		return curHeight >= blockHeightEvent.GetBlockHeight()
	}
	terminator := func(_ types.Trigger, triggerEvent types.TriggerEventI) bool {
		blockHeightEvent := triggerEvent.(*types.BlockHeightEvent)
		curHeight := uint64(ctx.BlockHeight())
		return curHeight < blockHeightEvent.GetBlockHeight()
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
			k.Logger(ctx).Debug(fmt.Sprintf("Event detected for trigger %d", trigger.Id))
			triggers = append(triggers, trigger)
		}
		return terminator(trigger, event), nil
	})
	if err != nil {
		panic(fmt.Errorf("unable to iterate event listeners for matching triggers: %w", err))
	}
	return
}

// emitTriggerDetected Emits an EventTriggerDetection for the provided trigger.
func (k Keeper) emitTriggerDetected(ctx sdk.Context, trigger types.Trigger) {
	err := ctx.EventManager().EmitTypedEvent(&types.EventTriggerDetected{
		TriggerId: fmt.Sprintf("%d", trigger.GetId()),
	})
	if err != nil {
		ctx.Logger().Error("unable to emit EventTriggerDetected", "err", err)
	}
}
