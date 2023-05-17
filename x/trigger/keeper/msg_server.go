package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/trigger/types"
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

// CreateTrigger creates new trigger from msg
func (s msgServer) CreateTrigger(goCtx context.Context, msg *types.MsgCreateTriggerRequest) (*types.MsgCreateTriggerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	trigger := s.NewTriggerWithID(ctx, msg.GetAuthority(), msg.GetEvent(), msg.GetActions())
	s.RegisterTrigger(ctx, trigger)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTriggerCreated,
			sdk.NewAttribute(types.AttributeKeyTriggerID, fmt.Sprintf("%d", trigger.GetId())),
		),
	)

	return &types.MsgCreateTriggerResponse{Id: trigger.GetId()}, nil
}

// DestroyTrigger destroys trigger from msg
func (s msgServer) DestroyTrigger(goCtx context.Context, msg *types.MsgDestroyTriggerRequest) (*types.MsgDestroyTriggerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	trigger, err := s.GetTrigger(ctx, msg.GetId())
	if err != nil {
		return nil, err
	}
	if trigger.GetOwner() != msg.GetAuthority() {
		return nil, types.ErrInvalidTriggerAuthority
	}
	s.UnregisterTrigger(ctx, trigger)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTriggerDestroyed,
			sdk.NewAttribute(types.AttributeKeyTriggerID, fmt.Sprintf("%d", trigger.GetId())),
		),
	)

	return &types.MsgDestroyTriggerResponse{}, nil
}
