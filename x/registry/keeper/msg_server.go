package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/gogoproto/proto"
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

	k.keeper.emitEvent(sdkCtx, types.NewEventNFTRegistered(msg.Key, msg.Signer))

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

	k.keeper.emitEvent(sdkCtx, types.NewEventRoleGranted(msg.Key, msg.Role, msg.Addresses, msg.Signer))

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

	k.keeper.emitEvent(sdkCtx, types.NewEventRoleRevoked(msg.Key, msg.Role, msg.Addresses, msg.Signer))

	return &types.MsgRevokeRoleResponse{}, nil
}

// UnregisterNFT unregisters an NFT from the registry.
// This removes the entire registry entry and associated data for the specified key.
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

// RegistryBulkUpdate registers, or updates, multiple NFTs in the registry.
// This creates multiple registry entries, or updates if one exists.
func (k msgServer) RegistryBulkUpdate(ctx context.Context, msg *types.MsgRegistryBulkUpdate) (*types.MsgRegistryBulkUpdateResponse, error) {
	for i, entry := range msg.Entries {
		// Validate that the NFT exists
		if err := k.ValidateNFTExists(ctx, &entry.Key.AssetClassId, &entry.Key.NftId); err != nil {
			return nil, fmt.Errorf("[%d]: %w", i, err)
		}

		// Validate that the signer owns the NFT
		if err := k.ValidateNFTOwner(ctx, &entry.Key.AssetClassId, &entry.Key.NftId, msg.Signer); err != nil {
			return nil, fmt.Errorf("[%d]: %w", i, err)
		}

		// Get the original registry, so we know what we're updating.
		orig, err := k.GetRegistry(ctx, entry.Key)
		if err != nil {
			return nil, fmt.Errorf("could not get existing registry entry [%d]: %w", i, err)
		}

		// Store the registry.
		err = k.Registry.Set(ctx, entry.Key.CollKey(), entry)
		if err != nil {
			return nil, fmt.Errorf("error setting registry entry [%d]: %w", i, err)
		}

		// Create and emit all the needed events.
		grantEvents, revokeEvents := types.GetChangeEvents(orig, &entry)
		allEvents := make([]proto.Message, 0, 1+len(grantEvents)+len(revokeEvents))
		if orig == nil {
			// If it didn't exist before, it was just created, so use that event.
			allEvents = append(allEvents, types.NewEventNFTRegistered(entry.Key))
		} else {
			// Otherwise, it's an update, so we use that event.
			allEvents = append(allEvents, types.NewEventRegistryBulkUpdated(entry.Key))
		}
		// Add all the grant events.
		for _, tev := range grantEvents {
			allEvents = append(allEvents, tev)
		}
		// Add all the revoke events.
		for _, tev := range revokeEvents {
			allEvents = append(allEvents, tev)
		}
		k.EmitEvents(ctx, allEvents...)
	}

	k.keeper.emitEvent(sdkCtx, types.NewEventRegistryBulkUpdate(uint64(len(msg.Entries)), msg.Signer))

	return &types.MsgRegistryBulkUpdateResponse{}, nil
}
