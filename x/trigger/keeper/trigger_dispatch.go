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
		action := item.GetTrigger().Action

		err = k.RunAction(ctx, action)
		if err != nil {
			// TODO We got an issue
			fmt.Println(("error"))
		}
	}
}

func (k Keeper) RunAction(ctx sdk.Context, action *types.Any) error {
	cacheCtx, flush := ctx.CacheContext()

	msgs, err := sdktx.GetMsgs([]*types.Any{action}, "sdk.MsgProposal")
	if err != nil {
		// TODO Something was wrong with the message
		return err
	}
	results, err := k.HandleMsgs(cacheCtx, msgs)
	if err != nil {
		// TODO We had a problem handling the message
		return err
	}

	flush()
	for _, res := range results {
		// NOTE: The sdk msg handler creates a new EventManager, so events must be correctly propagated back to the current context
		ctx.EventManager().EmitEvents(res.GetEvents())
	}
	return nil
}

func (k Keeper) HandleMsgs(ctx sdk.Context, msgs []sdk.Msg) ([]sdk.Result, error) {
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

		results[i] = *r
	}
	return results, nil
}
