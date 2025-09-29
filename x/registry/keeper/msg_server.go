package keeper

import (
	"context"

	"cosmossdk.io/collections"

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

// RegisterNFT registers a new NFT in the registry.
// This creates a new registry entry with the specified roles and addresses.
func (k msgServer) RegisterNFT(ctx context.Context, msg *types.MsgRegisterNFT) (*types.MsgRegisterNFTResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Already exists check
	has, err := k.keeper.Registry.Has(sdkCtx, collections.Join(msg.Key.AssetClassId, msg.Key.NftId))
	if err != nil {
		return nil, err
	}
	if has {
		return nil, types.NewErrCodeRegistryAlreadyExists(msg.Key.String())
	}

	// Validate that the NFT exists
	if hasNFT := k.keeper.HasNFT(sdkCtx, &msg.Key.AssetClassId, &msg.Key.NftId); !hasNFT {
		return nil, types.NewErrCodeNFTNotFound(msg.Key.NftId)
	}

	// Validate that the signer owns the NFT
	nftOwner := k.keeper.GetNFTOwner(sdkCtx, &msg.Key.AssetClassId, &msg.Key.NftId)
	if len(nftOwner) == 0 || nftOwner.String() != msg.Signer {
		return nil, types.NewErrCodeUnauthorized("signer does not own the NFT")
	}

	err = k.keeper.CreateRegistry(sdkCtx, msg.Key, msg.Roles)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterNFTResponse{}, nil
}

// GrantRole grants a role to one or more addresses.
// This adds the specified addresses to the role for the given registry key.
func (k msgServer) GrantRole(ctx context.Context, msg *types.MsgGrantRole) (*types.MsgGrantRoleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// ensure the registry exists
	has, err := k.keeper.Registry.Has(sdkCtx, collections.Join(msg.Key.AssetClassId, msg.Key.NftId))
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.NewErrCodeRegistryNotFound(msg.Key.String())
	}

	// Validate that the signer owns the NFT
	nftOwner := k.keeper.GetNFTOwner(sdkCtx, &msg.Key.AssetClassId, &msg.Key.NftId)
	if len(nftOwner) == 0 || nftOwner.String() != msg.Signer {
		return nil, types.NewErrCodeUnauthorized("signer does not own the NFT")
	}

	err = k.keeper.GrantRole(sdkCtx, msg.Key, msg.Role, msg.Addresses)
	if err != nil {
		return nil, err
	}

	return &types.MsgGrantRoleResponse{}, nil
}

// RevokeRole revokes a role from one or more addresses.
// This removes the specified addresses from the role for the given registry key.
func (k msgServer) RevokeRole(ctx context.Context, msg *types.MsgRevokeRole) (*types.MsgRevokeRoleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// ensure the registry exists
	has, err := k.keeper.Registry.Has(sdkCtx, collections.Join(msg.Key.AssetClassId, msg.Key.NftId))
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.NewErrCodeRegistryNotFound(msg.Key.String())
	}

	// Validate that the signer owns the NFT
	nftOwner := k.keeper.GetNFTOwner(sdkCtx, &msg.Key.AssetClassId, &msg.Key.NftId)
	if len(nftOwner) == 0 || nftOwner.String() != msg.Signer {
		return nil, types.NewErrCodeUnauthorized("signer does not own the NFT")
	}

	if err := k.keeper.RevokeRole(sdkCtx, msg.Key, msg.Role, msg.Addresses); err != nil {
		return nil, err
	}

	return &types.MsgRevokeRoleResponse{}, nil
}

// UnregisterNFT unregisters an NFT from the registry.
// This removes the entire registry entry and associated data for the specified key.
func (k msgServer) UnregisterNFT(ctx context.Context, msg *types.MsgUnregisterNFT) (*types.MsgUnregisterNFTResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Ensure the registry exists
	has, err := k.keeper.Registry.Has(sdkCtx, collections.Join(msg.Key.AssetClassId, msg.Key.NftId))
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.NewErrCodeRegistryNotFound(msg.Key.String())
	}

	// Validate that the signer owns the NFT
	nftOwner := k.keeper.GetNFTOwner(sdkCtx, &msg.Key.AssetClassId, &msg.Key.NftId)
	if len(nftOwner) == 0 || nftOwner.String() != msg.Signer {
		return nil, types.NewErrCodeUnauthorized("signer does not own the NFT")
	}

	// Remove the registry entry
	if err := k.keeper.Registry.Remove(sdkCtx, collections.Join(msg.Key.AssetClassId, msg.Key.NftId)); err != nil {
		return nil, err
	}

	return &types.MsgUnregisterNFTResponse{}, nil
}

func (k msgServer) RegistryBulkUpdate(ctx context.Context, msg *types.MsgRegistryBulkUpdate) (*types.MsgRegistryBulkUpdateResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Upsert each provided registry entry using the keeper's create function
	// which performs the underlying set operation on the registry store.
	for _, entry := range msg.Entries {
		// Validate that the signer owns the NFT
		nftOwner := k.keeper.GetNFTOwner(sdkCtx, &entry.Key.AssetClassId, &entry.Key.NftId)
		if nftOwner == nil || nftOwner.String() != msg.Signer {
			return nil, types.NewErrCodeUnauthorized("signer does not own the NFT")
		}

		if err := k.keeper.CreateRegistry(sdkCtx, entry.Key, entry.Roles); err != nil {
			return nil, err
		}
	}

	return &types.MsgRegistryBulkUpdateResponse{}, nil
}
