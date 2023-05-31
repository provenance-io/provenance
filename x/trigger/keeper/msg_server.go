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

	event, err := msg.GetTriggerEventI()
	if err != nil {
		return nil, err
	}
	if err = event.ValidateContext(ctx); err != nil {
		return nil, err
	}

	trigger := s.NewTriggerWithID(ctx, msg.GetAuthority(), msg.GetEvent(), msg.GetActions())
	s.RegisterTrigger(ctx, trigger)

	err = ctx.EventManager().EmitTypedEvent(&types.EventTriggerCreated{
		TriggerId: fmt.Sprintf("%d", trigger.GetId()),
	})
	if err != nil {
		return nil, err
	}

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
	s.RemoveGasLimit(ctx, trigger.GetId())

	err = ctx.EventManager().EmitTypedEvent(&types.EventTriggerDestroyed{
		TriggerId: fmt.Sprintf("%d", trigger.GetId()),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgDestroyTriggerResponse{}, nil
}
