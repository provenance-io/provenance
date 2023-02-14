package reward

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/provenance-io/provenance/x/reward/keeper"
	"github.com/provenance-io/provenance/x/reward/types"
)

// NewHandler returns a handler for reward messages.
func NewHandler(k keeper.Keeper) sdk.Handler {
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgCreateRewardProgramRequest:
			res, err := msgServer.CreateRewardProgram(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgEndRewardProgramRequest:
			res, err := msgServer.EndRewardProgram(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgClaimRewardsRequest:
			res, err := msgServer.ClaimRewards(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgClaimAllRewardsRequest:
			res, err := msgServer.ClaimAllRewards(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		default:
			return nil, sdkerrors.ErrUnknownRequest.Wrapf("unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}
