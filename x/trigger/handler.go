package reward

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler returns a handler for reward messages.
func NewHandler(k keeper.Keeper) sdk.Handler {
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		/*case *types.MsgCreateRewardProgramRequest:
		res, err := msgServer.CreateRewardProgram(sdk.WrapSDKContext(ctx), msg)
		return sdk.WrapServiceResult(ctx, res, err)*/
		default:
			return nil, sdkerrors.ErrUnknownRequest.Wrapf("unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}
