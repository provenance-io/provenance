package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/group/errors"
)

// ProcessTriggers Reads triggers from queues and attempts to run them.
func (k Keeper) ProcessTriggers(ctx sdk.Context) {
	for !k.QueueIsEmpty(ctx) {
		item := k.QueuePeek(ctx)
		k.Dequeue(ctx)

		triggerID := item.GetTrigger().Id
		gasLimit := k.GetGasLimit(ctx, triggerID)
		k.RemoveGasLimit(ctx, triggerID)

		actions := item.GetTrigger().Actions
		k.RunActions(ctx, gasLimit, actions)
	}
}

// RunActions Runs all the actions and constrains them by gasLimit.
func (k Keeper) RunActions(ctx sdk.Context, gasLimit uint64, actions []*types.Any) {
	cacheCtx, flush := ctx.CacheContext()
	gasMeter := sdk.NewInfiniteGasMeter()
	cacheCtx = cacheCtx.WithGasMeter(gasMeter)

	msgs, err := sdktx.GetMsgs(actions, "RunActions")
	if err != nil {
		k.Logger(ctx).Error(
			"GetMsgs",
			"actions", actions,
			"error", err,
		)
		return
	}
	results, err := k.HandleMsgs(cacheCtx, msgs, gasLimit)
	if err != nil {
		k.Logger(ctx).Error(
			"HandleMsgs",
			"error", err,
		)
		return
	}

	flush()
	for _, res := range results {
		ctx.EventManager().EmitEvents(res.GetEvents())
	}
}

// HandleMsgs Handles each message and verifies gas limit has not been exceeded.
func (k Keeper) HandleMsgs(ctx sdk.Context, msgs []sdk.Msg, gasLimit uint64) ([]sdk.Result, error) {
	results := make([]sdk.Result, len(msgs))
	for i, msg := range msgs {
		handler := k.router.Handler(msg)
		if handler == nil {
			return nil, errorsmod.Wrapf(errors.ErrInvalid, "no message handler found for %q", sdk.MsgTypeURL(msg))
		}
		r, err := handler(ctx, msg)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "message %s at position %d", sdk.MsgTypeURL(msg), i)
		}
		// Handler should always return non-nil sdk.Result.
		if r == nil {
			return nil, fmt.Errorf("got nil sdk.Result for message %q at position %d", msg, i)
		}

		if ctx.GasMeter().GasConsumed() > gasLimit {
			return nil, fmt.Errorf("gas %d exceeded limit %d for message %q at position %d", ctx.GasMeter().GasConsumed(), gasLimit, msg, i)
		}

		results[i] = *r
	}
	return results, nil
}
