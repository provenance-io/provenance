package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	triggertypes "github.com/provenance-io/provenance/x/trigger/types"
)

// RegisterTrigger Adds the trigger to the trigger, event listener, and gas store
func (k Keeper) RegisterTrigger(ctx sdk.Context, trigger triggertypes.Trigger) {
	k.SetTrigger(ctx, trigger)
	k.SetEventListener(ctx, trigger)

	gasLimit := ctx.GasMeter().GasRemaining() - SetGasLimitCost
	k.SetGasLimit(ctx, trigger.GetId(), gasLimit)
	ctx.GasMeter().ConsumeGas(gasLimit, "trigger creation")
}

// UnregisterTrigger Removes the trigger from the trigger, and event listener
func (k Keeper) UnregisterTrigger(ctx sdk.Context, trigger triggertypes.Trigger) {
	k.RemoveTrigger(ctx, trigger.GetId())
	k.RemoveEventListener(ctx, trigger)
}
