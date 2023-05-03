package keeper

import (
	"context"

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

	id, err := s.GetNextTriggerID(ctx)
	if err != nil {
		// Throw an error
		return nil, err
	}

	trigger := types.NewTrigger(id, msg.GetEvent(), msg.GetAction())
	err = s.SetTrigger(ctx, trigger)
	if err != nil {
		// Throw an error
		return nil, err
	}

	return &types.MsgCreateTriggerResponse{Id: trigger.GetId()}, nil
}
