package keeper

import (
	"context"

	"github.com/provenance-io/provenance/x/reward/types"
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

func (s msgServer) CreateRewardProgram(context.Context, *types.MsgCreateRewardProgramRequest) (*types.MsgCreateRewardProgramResponse, error) {
	return nil, nil
}
