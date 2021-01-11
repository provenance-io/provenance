package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/name/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the account MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) BindName(goCtx context.Context, msg *types.MsgBindNameRequest) (*types.MsgBindNameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO keeper.Bind
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeNameBound,
			sdk.NewAttribute(types.KeyAttributeAddress, msg.Record.Address),
			sdk.NewAttribute(types.KeyAttributeName, msg.Record.Name),
		),
	)

	return &types.MsgBindNameResponse{}, nil
}

func (k msgServer) DeleteName(goCtx context.Context, msg *types.MsgDeleteNameRequest) (*types.MsgDeleteNameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO keeper.Delete
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeNameUnbound,
			sdk.NewAttribute(types.KeyAttributeAddress, msg.Record.Address),
			sdk.NewAttribute(types.KeyAttributeName, msg.Record.Name),
		),
	)

	return &types.MsgDeleteNameResponse{}, nil
}
