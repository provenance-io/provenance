package keeper

import (
	"context"
	"fmt"

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
	// Validate that the NFT exists
	if err := k.ValidateNFTExists(ctx, &msg.Key.AssetClassId, &msg.Key.NftId); err != nil {
		return nil, err
	}

	// Validate that the signer owns the NFT
	if err := k.ValidateNFTOwner(ctx, &msg.Key.AssetClassId, &msg.Key.NftId, msg.Signer); err != nil {
		return nil, err
	}

	// Store the registry in state.
	err := k.CreateRegistry(ctx, msg.Key, msg.Roles)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterNFTResponse{}, nil
}

// GrantRole grants a role to one or more addresses.
// This adds the specified addresses to the role for the given registry key.
func (k msgServer) GrantRole(ctx context.Context, msg *types.MsgGrantRole) (*types.MsgGrantRoleResponse, error) {
	// ensure the registry exists
	if err := k.ValidateRegistryExists(ctx, msg.Key); err != nil {
		return nil, err
	}

	// Validate that the signer owns the NFT
	if err := k.ValidateNFTOwner(ctx, &msg.Key.AssetClassId, &msg.Key.NftId, msg.Signer); err != nil {
		return nil, err
	}

	// Grant the role to the addresses.
	err := k.Keeper.GrantRole(ctx, msg.Key, msg.Role, msg.Addresses)
	if err != nil {
		return nil, err
	}

	return &types.MsgGrantRoleResponse{}, nil
}

// RevokeRole revokes a role from one or more addresses.
// This removes the specified addresses from the role for the given registry key.
func (k msgServer) RevokeRole(ctx context.Context, msg *types.MsgRevokeRole) (*types.MsgRevokeRoleResponse, error) {
	// ensure the registry exists
	if err := k.ValidateRegistryExists(ctx, msg.Key); err != nil {
		return nil, err
	}

	// Validate that the signer owns the NFT
	if err := k.ValidateNFTOwner(ctx, &msg.Key.AssetClassId, &msg.Key.NftId, msg.Signer); err != nil {
		return nil, err
	}

	// Revoke the role from the addresses.
	err := k.Keeper.RevokeRole(ctx, msg.Key, msg.Role, msg.Addresses)
	if err != nil {
		return nil, err
	}

	return &types.MsgRevokeRoleResponse{}, nil
}

// UnregisterNFT unregisters an NFT from the registry.
// This removes the entire registry entry for the specified key.
func (k msgServer) UnregisterNFT(ctx context.Context, msg *types.MsgUnregisterNFT) (*types.MsgUnregisterNFTResponse, error) {
	// Validate that the signer owns the NFT
	if err := k.ValidateNFTOwner(ctx, &msg.Key.AssetClassId, &msg.Key.NftId, msg.Signer); err != nil {
		return nil, err
	}

	if err := k.DeleteRegistry(ctx, msg.Key); err != nil {
		return nil, err
	}

	return &types.MsgUnregisterNFTResponse{}, nil
}

func (k msgServer) RegistryBulkUpdate(ctx context.Context, msg *types.MsgRegistryBulkUpdate) (*types.MsgRegistryBulkUpdateResponse, error) {
	// Upsert each provided registry entry using the keeper's create function
	// which performs the underlying set operation on the registry store.
	for i, entry := range msg.Entries {
		// Validate that the NFT exists
		if err := k.ValidateNFTExists(ctx, &entry.Key.AssetClassId, &entry.Key.NftId); err != nil {
			return nil, err
		}

		// Validate that the signer owns the NFT
		if err := k.ValidateNFTOwner(ctx, &entry.Key.AssetClassId, &entry.Key.NftId, msg.Signer); err != nil {
			return nil, err
		}

		// Store the registry.
		err := k.Registry.Set(ctx, entry.Key.CollKey(), entry)
		if err != nil {
			return nil, fmt.Errorf("error setting registry entry [%d]: %w", i, err)
		}
		k.EmitEvent(ctx, types.NewEventRegistryBulkUpdated(entry.Key))
	}

	return &types.MsgRegistryBulkUpdateResponse{}, nil
}
