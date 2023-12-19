package attribute

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/provenance-io/provenance/x/attribute/keeper"
	"github.com/provenance-io/provenance/x/attribute/types"
)

// NewHandler returns a handler for attribute messages.
// TODO[1760]: attribute: Delete the attribute NewHandler.
func NewHandler(k keeper.Keeper) func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgAddAttributeRequest:
			res, err := msgServer.AddAttribute(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgUpdateAttributeRequest:
			res, err := msgServer.UpdateAttribute(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDeleteAttributeRequest:
			res, err := msgServer.DeleteAttribute(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDeleteDistinctAttributeRequest:
			res, err := msgServer.DeleteDistinctAttribute(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		default:
			return nil, sdkerrors.ErrUnknownRequest.Wrapf("unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}
