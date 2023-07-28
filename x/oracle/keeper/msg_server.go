package keeper

import (
	"context"

	"github.com/provenance-io/provenance/x/oracle/types"
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

// UpdateOracle changes the oracle's address to the provided one
func (s msgServer) UpdateOracle(goCtx context.Context, msg *types.MsgUpdateOracleRequest) (*types.MsgUpdateOracleResponse, error) {
	//ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.MsgUpdateOracleResponse{}, nil
}

func (s msgServer) QueryOracle(goCtx context.Context, msg *types.MsgQueryOracleRequest) (*types.MsgQueryOracleResponse, error) {
	return &types.MsgQueryOracleResponse{}, nil
}
