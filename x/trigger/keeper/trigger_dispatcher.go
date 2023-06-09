package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

	types "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/group/errors"
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
		k.Logger(ctx).Error(
			"GetMsgs",
			"actions", actions,
			"error", err,
		)
		return
	}
	results, err := k.handleMsgs(cacheCtx, msgs)
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

// handleMsgs Handles each message and verifies gas limit has not been exceeded.
func (k Keeper) handleMsgs(ctx sdk.Context, msgs []sdk.Msg) ([]sdk.Result, error) {
	results := make([]sdk.Result, len(msgs))

	for i, msg := range msgs {
		handler := k.router.Handler(msg)
		if handler == nil {
			return nil, errorsmod.Wrapf(errors.ErrInvalid, "no message handler found for %q", sdk.MsgTypeURL(msg))
		}
		r, err := k.safeHandle(ctx, msg, handler)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "message %s at position %d", sdk.MsgTypeURL(msg), i)
		}
		// Handler should always return non-nil sdk.Result.
		if r == nil {
			return nil, fmt.Errorf("got nil sdk.Result for message %q at position %d", sdk.MsgTypeURL(msg), i)
		}

		results[i] = *r
	}
	return results, nil
}

// safeHandle Handles one message and safely returns an error if it panics
func (k Keeper) safeHandle(ctx sdk.Context, msg sdk.Msg, handler MsgServiceHandler) (result *sdk.Result, err error) {
	defer func() {
		if e := recover(); e != nil {
			_, ok := e.(storetypes.ErrorOutOfGas)
			if ok {
				result = nil
				err = fmt.Errorf("gas %d exceeded limit %d for message %q", ctx.GasMeter().GasConsumed(), ctx.GasMeter().Limit(), sdk.MsgTypeURL(msg))
				return
			}

			// It is not an out of gas panic so pass it up
			panic(e)
		}
	}()
	return handler(ctx, msg)
}
