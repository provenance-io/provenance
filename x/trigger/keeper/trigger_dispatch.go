package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"

	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/group/errors"
)

func (k Keeper) RunTriggers(ctx sdk.Context) {
	for !k.QueueIsEmpty(ctx) {
		item, err := k.Peek(ctx)
		if err != nil {
			// TODO Something went wrong
			fmt.Println(("error"))
		}
		_ = k.DequeueTrigger(ctx)
		actions := item.GetTrigger().Actions
		gasLimit, err := k.GetGasLimit(ctx, item.GetTrigger().Id)
		if err != nil {
			// TODO Something went wrong
			fmt.Println(("error"))
		}
		k.RemoveGasLimit(ctx, item.GetTrigger().Id)

		err = k.RunActions(ctx, gasLimit, actions)
		if err != nil {
			// TODO We got an issue
			fmt.Println(("error"))
		}
	}
}

func (k Keeper) RunActions(ctx sdk.Context, gasLimit uint64, action []*types.Any) error {
	cacheCtx, flush := ctx.CacheContext()
	gasMeter := sdk.NewInfiniteGasMeter()
	cacheCtx = cacheCtx.WithGasMeter(gasMeter)

	msgs, err := sdktx.GetMsgs(action, "RunAction - sdk.MsgCreateTriggerRequest")
	if err != nil {
		// TODO Something was wrong with the message
		return err
	}
	results, err := k.HandleMsgs(cacheCtx, msgs, gasLimit)
	if err != nil {
		// TODO We had a problem handling one or more of the messages
		return err
	}

	flush()
	for _, res := range results {
		// NOTE: The sdk msg handler creates a new EventManager, so events must be correctly propagated back to the current context
		ctx.EventManager().EmitEvents(res.GetEvents())
	}
	return nil
}

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
