package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"

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

	cachedValue := msg.Event.GetCachedValue()
	if cachedValue == nil {
		return nil, fmt.Errorf("cached value does not exist")
	}
	event := cachedValue.(types.TriggerEventI)
	fmt.Printf("Prefix Name is: %s\n", event.GetEventPrefix())

	msgs, err := sdktx.GetMsgs(msg.Actions, "RunAction - sdk.MsgCreateTriggerRequest")
	if len(msgs) == 0 || err != nil {
		return nil, err
	}

	trigger, err := s.NewTriggerWithID(ctx, msg.GetAuthority(), msg.GetEvent(), msg.GetActions())
	if err != nil {
		// Throw an error
		return nil, err
	}

	// TODO Maybe we can group these
	s.SetTrigger(ctx, trigger)
	s.SetEventListener(ctx, trigger)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTriggerCreated,
			sdk.NewAttribute(types.AttributeKeyTriggerID, fmt.Sprintf("%d", trigger.GetId())),
		),
	)

	return &types.MsgCreateTriggerResponse{Id: trigger.GetId()}, nil
}
