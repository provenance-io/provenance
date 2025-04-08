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

// RegisterAddress handles MsgRegisterAddress
func (k msgServer) RegisterAddress(ctx context.Context, msg *registry.MsgRegisterAddress) (*registry.MsgRegisterAddressResponse, error) {
	// TODO: Implement
	return &registry.MsgRegisterAddressResponse{}, nil
}

// UpdateRoles handles MsgUpdateRoles
func (k msgServer) UpdateRoles(ctx context.Context, msg *registry.MsgUpdateRoles) (*registry.MsgUpdateRolesResponse, error) {
	// TODO: Implement
	return &registry.MsgUpdateRolesResponse{}, nil
}

// RemoveAddress handles MsgRemoveAddress
func (k msgServer) RemoveAddress(ctx context.Context, msg *registry.MsgRemoveAddress) (*registry.MsgRemoveAddressResponse, error) {
	// TODO: Implement
	return &registry.MsgRemoveAddressResponse{}, nil
}
