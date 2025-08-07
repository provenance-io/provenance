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

	keyStr := msg.Key.String()

	// Already exists check
	has, err := k.keeper.Registry.Has(sdkCtx, keyStr)
	if err != nil {
		return nil, err
	}
	if has {
		return nil, types.NewErrCodeRegistryAlreadyExists(keyStr)
	}

	// Validate that the NFT exists
	if hasNFT := k.keeper.HasNFT(sdkCtx, &msg.Key.AssetClassId, &msg.Key.NftId); !hasNFT {
		return nil, types.NewErrCodeNFTNotFound(msg.Key.NftId)
	}

	// Validate that the authority owns the NFT
	if nftOwner := k.keeper.GetNFTOwner(sdkCtx, &msg.Key.AssetClassId, &msg.Key.NftId); nftOwner == nil || nftOwner.String() != msg.Authority {
		return nil, types.NewErrCodeUnauthorized("authority does not own the NFT")
	}

	err = k.keeper.CreateRegistry(sdkCtx, msg.Key, msg.Roles)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterNFTResponse{}, nil
}

func (k msgServer) GrantRole(ctx context.Context, msg *types.MsgGrantRole) (*types.MsgGrantRoleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	keyStr := msg.Key.String()

	// ensure the registry exists
	has, err := k.keeper.Registry.Has(sdkCtx, keyStr)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.NewErrCodeRegistryNotFound(keyStr)
	}

	// Validate that the authority owns the NFT
	nftOwner := k.keeper.GetNFTOwner(sdkCtx, &msg.Key.AssetClassId, &msg.Key.NftId)
	if nftOwner == nil || nftOwner.String() != msg.Authority {
		return nil, types.NewErrCodeUnauthorized("authority does not own the NFT")
	}

	err = k.keeper.GrantRole(sdkCtx, msg.Key, msg.Role, msg.Addresses)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (k msgServer) RevokeRole(ctx context.Context, msg *types.MsgRevokeRole) (*types.MsgRevokeRoleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	keyStr := msg.Key.String()

	// ensure the registry exists
	has, err := k.keeper.Registry.Has(sdkCtx, keyStr)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.NewErrCodeRegistryNotFound(keyStr)
	}

	// Validate that the authority owns the NFT
	nftOwner := k.keeper.GetNFTOwner(sdkCtx, &msg.Key.AssetClassId, &msg.Key.NftId)
	if nftOwner == nil || nftOwner.String() != msg.Authority {
		return nil, types.NewErrCodeUnauthorized("authority does not own the NFT")
	}

	if err := k.keeper.RevokeRole(sdkCtx, msg.Key, msg.Role, msg.Addresses); err != nil {
		return nil, err
	}

	return nil, nil
}

func (k msgServer) UnregisterNFT(ctx context.Context, msg *types.MsgUnregisterNFT) (*types.MsgUnregisterNFTResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Validate that the authority owns the NFT
	nftOwner := k.keeper.GetNFTOwner(sdkCtx, &msg.Key.AssetClassId, &msg.Key.NftId)
	if nftOwner == nil || nftOwner.String() != msg.Authority {
		return nil, types.NewErrCodeUnauthorized("authority does not own the NFT")
	}

	// TODO: Implement unregister functionality
	return &types.MsgUnregisterNFTResponse{}, nil
}
