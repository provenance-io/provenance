package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/registry/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServer returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServer(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// RegisterNFT registers a new NFT in the registry.
// This creates a new registry entry with the specified roles and addresses.
func (k msgServer) RegisterNFT(ctx context.Context, msg *types.MsgRegisterNFT) (*types.MsgRegisterNFTResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Already exists check
	has, err := k.Registry.Has(sdkCtx, msg.Key.CollKey())
	if err != nil {
		return nil, err
	}
	if has {
		return nil, types.NewErrCodeRegistryAlreadyExists(msg.Key.String())
	}

	// Validate that the NFT exists
	if err = k.ValidateNFTExists(ctx, &msg.Key.AssetClassId, &msg.Key.NftId); err != nil {
		return nil, err
	}

	// Validate that the signer owns the NFT
	if err = k.ValidateNFTOwner(ctx, &msg.Key.AssetClassId, &msg.Key.NftId, msg.Signer); err != nil {
		return nil, err
	}

	err = k.CreateRegistry(sdkCtx, msg.Key, msg.Roles)
	if err != nil {
		return nil, err
	}

	k.EmitEvent(sdkCtx, types.NewEventNFTRegistered(msg.Key))
	return &types.MsgRegisterNFTResponse{}, nil
}

// GrantRole grants a role to one or more addresses.
// This adds the specified addresses to the role for the given registry key.
func (k msgServer) GrantRole(ctx context.Context, msg *types.MsgGrantRole) (*types.MsgGrantRoleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// ensure the registry exists
	has, err := k.Registry.Has(sdkCtx, msg.Key.CollKey())
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.NewErrCodeRegistryNotFound(msg.Key.String())
	}

	// Validate that the signer owns the NFT
	if err = k.ValidateNFTOwner(ctx, &msg.Key.AssetClassId, &msg.Key.NftId, msg.Signer); err != nil {
		return nil, err
	}

	err = k.Keeper.GrantRole(sdkCtx, msg.Key, msg.Role, msg.Addresses)
	if err != nil {
		return nil, err
	}

	k.EmitEvent(sdkCtx, types.NewEventRoleGranted(msg.Key, msg.Role, msg.Addresses))
	return &types.MsgGrantRoleResponse{}, nil
}

// RevokeRole revokes a role from one or more addresses.
// This removes the specified addresses from the role for the given registry key.
func (k msgServer) RevokeRole(ctx context.Context, msg *types.MsgRevokeRole) (*types.MsgRevokeRoleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// ensure the registry exists
	has, err := k.Registry.Has(sdkCtx, msg.Key.CollKey())
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.NewErrCodeRegistryNotFound(msg.Key.String())
	}

	// Validate that the signer owns the NFT
	if err = k.ValidateNFTOwner(ctx, &msg.Key.AssetClassId, &msg.Key.NftId, msg.Signer); err != nil {
		return nil, err
	}

	if err := k.Keeper.RevokeRole(sdkCtx, msg.Key, msg.Role, msg.Addresses); err != nil {
		return nil, err
	}

	k.EmitEvent(sdkCtx, types.NewEventRoleRevoked(msg.Key, msg.Role, msg.Addresses))
	return &types.MsgRevokeRoleResponse{}, nil
}

// UnregisterNFT unregisters an NFT from the registry.
// This removes the entire registry entry for the specified key.
func (k msgServer) UnregisterNFT(ctx context.Context, msg *types.MsgUnregisterNFT) (*types.MsgUnregisterNFTResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Validate that the signer owns the NFT
	if err := k.ValidateNFTOwner(ctx, &msg.Key.AssetClassId, &msg.Key.NftId, msg.Signer); err != nil {
		return nil, err
	}

	// TODO: Implement unregister functionality

	k.EmitEvent(sdkCtx, types.NewEventNFTUnregistered(msg.Key))
	return &types.MsgUnregisterNFTResponse{}, nil
}

func (k msgServer) RegistryBulkUpdate(ctx context.Context, msg *types.MsgRegistryBulkUpdate) (*types.MsgRegistryBulkUpdateResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Upsert each provided registry entry using the keeper's create function
	// which performs the underlying set operation on the registry store.
	for _, entry := range msg.Entries {
		// Validate that the NFT exists
		if err := k.ValidateNFTExists(ctx, &entry.Key.AssetClassId, &entry.Key.NftId); err != nil {
			return nil, err
		}

		// Validate that the signer owns the NFT
		if err := k.ValidateNFTOwner(ctx, &entry.Key.AssetClassId, &entry.Key.NftId, msg.Signer); err != nil {
			return nil, err
		}

		if err := k.CreateRegistry(sdkCtx, entry.Key, entry.Roles); err != nil {
			return nil, err
		}
		k.EmitEvent(sdkCtx, types.NewEventRegistryBulkUpdated(entry.Key))
	}

	return &types.MsgRegistryBulkUpdateResponse{}, nil
}
