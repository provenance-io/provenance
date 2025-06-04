package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, err
	}

	err = k.keeper.CreateRegistry(sdkCtx, authority, msg.Key, msg.Roles)
	if err != nil {
		return nil, err
	}

	return &registry.MsgRegisterNFTResponse{}, nil
}

func (k msgServer) GrantRole(ctx context.Context, msg *registry.MsgGrantRole) (*registry.MsgGrantRoleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, err
	}

	addresses := make([]*sdk.AccAddress, len(msg.Addresses))
	for i, address := range msg.Addresses {
		addr, err := sdk.AccAddressFromBech32(address)
		if err != nil {
			return nil, err
		}

		addresses[i] = &addr
	}

	err = k.keeper.GrantRole(sdkCtx, authority, msg.Key, msg.Role, addresses)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (k msgServer) RevokeRole(ctx context.Context, msg *registry.MsgRevokeRole) (*registry.MsgRevokeRoleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, err
	}

	addresses := make([]*sdk.AccAddress, len(msg.Addresses))
	for i, address := range msg.Addresses {
		addr, err := sdk.AccAddressFromBech32(address)
		if err != nil {
			return nil, err
		}

		addresses[i] = &addr
	}

	err = k.keeper.RevokeRole(sdkCtx, authority, msg.Key, msg.Role, addresses)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (k msgServer) UnregisterNFT(ctx context.Context, msg *registry.MsgUnregisterNFT) (*registry.MsgUnregisterNFTResponse, error) {
	return nil, nil
}
