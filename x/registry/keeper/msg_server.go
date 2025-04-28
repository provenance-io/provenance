package keeper

import (
	"context"

	"github.com/provenance-io/provenance/x/registry"
)

type msgServer struct {
	keeper RegistryKeeper
}

// NewMsgServer returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServer(keeper RegistryKeeper) registry.MsgServer {
	return &msgServer{keeper: keeper}
}

func (k msgServer) RegisterNFT(ctx context.Context, msg *registry.MsgRegisterNFT) (*registry.MsgRegisterNFTResponse, error) {
	return nil, nil
}

func (k msgServer) GrantRole(ctx context.Context, msg *registry.MsgGrantRole) (*registry.MsgGrantRoleResponse, error) {
	return nil, nil
}

func (k msgServer) RevokeRole(ctx context.Context, msg *registry.MsgRevokeRole) (*registry.MsgRevokeRoleResponse, error) {
	return nil, nil
}

func (k msgServer) UnregisterNFT(ctx context.Context, msg *registry.MsgUnregisterNFT) (*registry.MsgUnregisterNFTResponse, error) {
	return nil, nil
}
