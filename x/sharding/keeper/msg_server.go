package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/sharding/types"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the account MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// Reads from the store
func (s msgServer) Read(goCtx context.Context, msg *types.MsgReadRequest) (*types.MsgReadResponse, error) {
	_ = sdk.UnwrapSDKContext(goCtx)

	return &types.MsgReadResponse{}, nil
}

// Writes to the store
func (s msgServer) Write(goCtx context.Context, msg *types.MsgWriteRequest) (*types.MsgWriteResponse, error) {
	_ = sdk.UnwrapSDKContext(goCtx)

	return &types.MsgWriteResponse{}, nil
}

// Read and writes to the store
func (s msgServer) Update(goCtx context.Context, msg *types.MsgUpdateRequest) (*types.MsgUpdateResponse, error) {
	_ = sdk.UnwrapSDKContext(goCtx)

	return &types.MsgUpdateResponse{}, nil
}
