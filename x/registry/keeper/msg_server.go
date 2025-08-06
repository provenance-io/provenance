package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/registry/types"
)

type msgServer struct {
	keeper Keeper
}

// NewMsgServer returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServer(keeper Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

func (k msgServer) RegisterNFT(ctx context.Context, msg *types.MsgRegisterNFT) (*types.MsgRegisterNFTResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, err
	}

	err = k.keeper.CreateRegistry(sdkCtx, authority, msg.Key, msg.Roles)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterNFTResponse{}, nil
}

func (k msgServer) GrantRole(ctx context.Context, msg *types.MsgGrantRole) (*types.MsgGrantRoleResponse, error) {
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

func (k msgServer) RevokeRole(ctx context.Context, msg *types.MsgRevokeRole) (*types.MsgRevokeRoleResponse, error) {
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

func (k msgServer) UnregisterNFT(ctx context.Context, msg *types.MsgUnregisterNFT) (*types.MsgUnregisterNFTResponse, error) {
	return nil, nil
}
