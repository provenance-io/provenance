package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	triggertypes "github.com/provenance-io/provenance/x/trigger/types"
)

// RegisterTrigger Adds the trigger to the trigger, event listener, and gas store
func (k Keeper) RegisterTrigger(ctx sdk.Context, trigger triggertypes.Trigger) {
	if err := k.SetTrigger(ctx, trigger); err != nil {
		ctx.Logger().Error(
			"Failed to set trigger",
			"triggerID", trigger.GetId(),
			"err", err,
		)
	}
	if err := k.SetEventListener(ctx, trigger); err != nil {
		ctx.Logger().Error(
			"Failed to set event listener for trigger",
			"triggerID", trigger.GetId(),
			"err", err,
		)
	}
}

// UnregisterTrigger Removes the trigger from the trigger, and event listener
func (k Keeper) UnregisterTrigger(ctx sdk.Context, trigger triggertypes.Trigger) {
	k.RemoveTrigger(ctx, trigger.GetId())
	if _, err := k.RemoveEventListener(ctx, trigger); err != nil {
		ctx.Logger().Error(
			"Failed to remove event listener for trigger",
			"triggerID", trigger.GetId(),
			"err", err,
		)
	}
}
