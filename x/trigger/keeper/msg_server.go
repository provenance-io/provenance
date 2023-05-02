package keeper

import (
	"context"

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
	//ctx := sdk.UnwrapSDKContext(goCtx)
	triggerID := uint64(0)
	return &types.MsgCreateTriggerResponse{Id: triggerID}, nil
}
