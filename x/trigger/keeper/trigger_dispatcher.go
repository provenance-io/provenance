package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
	types "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
)

const (
	MaximumActions  uint64 = 5
	MaximumQueueGas uint64 = 2000000
)

// ProcessTriggers Reads triggers from queues and attempts to run them.
func (k Keeper) ProcessTriggers(ctx sdk.Context) {
	var actionsProcessed uint64
	var gasConsumed uint64

	for !k.QueueIsEmpty(ctx) && actionsProcessed < MaximumActions {
		item := k.QueuePeek(ctx)
		triggerID := item.GetTrigger().Id
		gasLimit := k.GetGasLimit(ctx, triggerID)

		if gasLimit+gasConsumed > MaximumQueueGas {
			return
		}
		actionsProcessed++
		gasConsumed += gasLimit

		k.Dequeue(ctx)
		k.RemoveGasLimit(ctx, triggerID)

		actions := item.GetTrigger().Actions
		k.runActions(ctx, gasLimit, actions)
	}
}

// RunActions Runs all the actions and constrains them by gasLimit.
func (k Keeper) runActions(ctx sdk.Context, gasLimit uint64, actions []*types.Any) {
	cacheCtx, flush := ctx.CacheContext()
	gasMeter := sdk.NewGasMeter(gasLimit)
	cacheCtx = cacheCtx.WithGasMeter(gasMeter)

	msgs, err := sdktx.GetMsgs(actions, "RunActions")
	if err != nil {
		ctx.GasMeter().ConsumeGas(gasLimit, "failed trigger dispatch")
		k.Logger(ctx).Error(
			"GetMsgs",
			"actions", actions,
			"error", err,
		)
		return
	}
	results, err := k.handleMsgs(cacheCtx, msgs)
	if err != nil {
		ctx.GasMeter().ConsumeGas(gasLimit, "failed trigger dispatch")
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

// handleMsgs Handles each message and verifies gas limit has not been exceeded.
func (k Keeper) handleMsgs(ctx sdk.Context, msgs []sdk.Msg) ([]sdk.Result, error) {
	results := make([]sdk.Result, len(msgs))

	for i, msg := range msgs {
		handler := k.router.Handler(msg)
		if handler == nil {
			return nil, fmt.Errorf("no message handler found for message %s at position %d", sdk.MsgTypeURL(msg), i)
		}
		r, err := k.safeHandle(ctx, msg, handler)
		if err != nil {
			return nil, fmt.Errorf("error processing message %s at position %d: %w", sdk.MsgTypeURL(msg), i, err)
		}
		// Handler should always return non-nil sdk.Result.
		if r == nil {
			return nil, fmt.Errorf("got nil sdk.Result for message %s at position %d", sdk.MsgTypeURL(msg), i)
		}

		results[i] = *r
	}
	return results, nil
}

// safeHandle Handles one message and safely returns an error if it panics
func (k Keeper) safeHandle(ctx sdk.Context, msg sdk.Msg, handler baseapp.MsgServiceHandler) (result *sdk.Result, err error) {
	defer func() {
		if e := recover(); e != nil {
			// If it's an out-of-gas panic, convert it to a nicer error.
			if _, ok := e.(storetypes.ErrorOutOfGas); ok {
				result = nil
				err = fmt.Errorf("gas %d exceeded limit %d for message %q", ctx.GasMeter().GasConsumed(), ctx.GasMeter().Limit(), sdk.MsgTypeURL(msg))
				return
			}

			// If it's some other error, wrap it up and return it (instead of panicking).
			if er, ok := e.(error); ok {
				err = fmt.Errorf("panic (recovered) processing msg: %w", er)
				return
			}

			// Otherwise, it's some other panic. Just create a new error for it.
			err = fmt.Errorf("panic (recovered) processing msg: %v", e)
		}
	}()
	return handler(ctx, msg)
}
