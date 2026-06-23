package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/provenance-io/provenance/internal/antewrapper"
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
	// Ensure the registry exists and fetch the current entry for auth checks.
	entry, err := k.GetRegistry(ctx, msg.Key)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, types.NewErrCodeRegistryNotFound(msg.Key.String())
	}

	roleAuths := types.RoleAuthorizationMap()
	if roleAuth, ok := roleAuths[msg.Role]; ok {
		// Policy-based path: the incoming addresses are the "new" assignment. A single signer must
		// satisfy the policy on its own; multi-party changes go through ProposeRoleChange.
		if err := k.Keeper.ValidateRoleChangeAuthorization(ctx, roleAuth, entry, msg.Addresses, []string{msg.Signer}); err != nil {
			return nil, err
		}
	} else {
		// Legacy fallback: the signer must own the NFT.
		if err := k.ValidateNFTOwner(ctx, &msg.Key.AssetClassId, &msg.Key.NftId, msg.Signer); err != nil {
			return nil, err
		}
	}

	// Grant the role to the addresses.
	err = k.Keeper.GrantRole(ctx, msg.Key, msg.Role, msg.Addresses)
	if err != nil {
		return nil, err
	}

	return &types.MsgGrantRoleResponse{}, nil
}

// RevokeRole revokes a role from one or more addresses.
// This removes the specified addresses from the role for the given registry key.
func (k msgServer) RevokeRole(ctx context.Context, msg *types.MsgRevokeRole) (*types.MsgRevokeRoleResponse, error) {
	// Ensure the registry exists and fetch the current entry for auth checks.
	entry, err := k.GetRegistry(ctx, msg.Key)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, types.NewErrCodeRegistryNotFound(msg.Key.String())
	}

	roleAuths := types.RoleAuthorizationMap()
	if roleAuth, ok := roleAuths[msg.Role]; ok {
		// Policy-based path. For revoke, no new addresses are being assigned. A single signer must
		// satisfy the policy on its own; multi-party changes go through ProposeRoleChange.
		if err := k.Keeper.ValidateRoleChangeAuthorization(ctx, roleAuth, entry, nil, []string{msg.Signer}); err != nil {
			return nil, err
		}
	} else {
		// Legacy fallback: the signer must own the NFT.
		if err := k.ValidateNFTOwner(ctx, &msg.Key.AssetClassId, &msg.Key.NftId, msg.Signer); err != nil {
			return nil, err
		}
	}

	// Revoke the role from the addresses.
	err = k.Keeper.RevokeRole(ctx, msg.Key, msg.Role, msg.Addresses)
	if err != nil {
		return nil, err
	}

	return &types.MsgRevokeRoleResponse{}, nil
}

// SetRoles atomically sets the desired state for one or more roles on a registry entry.
// Each role update specifies a role and the complete desired set of addresses for that role.
// An empty address list clears the role.
func (k msgServer) SetRoles(ctx context.Context, msg *types.MsgSetRoles) (*types.MsgSetRolesResponse, error) {
	// Ensure the registry exists and fetch the current entry for auth checks.
	entry, err := k.GetRegistry(ctx, msg.Key)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, types.NewErrCodeRegistryNotFound(msg.Key.String())
	}

	roleAuths := types.RoleAuthorizationMap()
	for _, update := range msg.RoleUpdates {
		if roleAuth, ok := roleAuths[update.Role]; ok {
			// The new addresses for this role are the desired state from the update. A single signer
			// must satisfy the policy on its own; multi-party changes go through ProposeRoleChange.
			if err := k.Keeper.ValidateRoleChangeAuthorization(ctx, roleAuth, entry, update.Addresses, []string{msg.Signer}); err != nil {
				return nil, err
			}
		} else {
			// Legacy fallback: the signer must own the NFT.
			if err := k.ValidateNFTOwner(ctx, &msg.Key.AssetClassId, &msg.Key.NftId, msg.Signer); err != nil {
				return nil, err
			}
		}
	}

	if err := k.Keeper.SetRoles(ctx, msg.Key, msg.RoleUpdates); err != nil {
		return nil, err
	}

	return &types.MsgSetRolesResponse{}, nil
}

// ProposeRoleChange opens a pending role change that accumulates single-signer approvals and
// auto-applies once the role's authorization policy is satisfied.
func (k msgServer) ProposeRoleChange(ctx context.Context, msg *types.MsgProposeRoleChange) (*types.MsgProposeRoleChangeResponse, error) {
	id, applied, err := k.Keeper.ProposeRoleChange(ctx, msg.Signer, msg.Key, msg.RoleUpdates)
	if err != nil {
		return nil, err
	}
	return &types.MsgProposeRoleChangeResponse{ChangeId: id, Applied: applied}, nil
}

// ApproveRoleChange records a single-signer approval for a pending role change and auto-applies
// it once the role's authorization policy is satisfied.
func (k msgServer) ApproveRoleChange(ctx context.Context, msg *types.MsgApproveRoleChange) (*types.MsgApproveRoleChangeResponse, error) {
	applied, err := k.Keeper.ApproveRoleChange(ctx, msg.Signer, msg.ChangeId)
	if err != nil {
		return nil, err
	}
	return &types.MsgApproveRoleChangeResponse{Applied: applied}, nil
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

// These are addresses with special access to the RegistryBulkUpdate endpoint.
var (
	authority1 = "pb1q3xhmqrjukjuhmccy4p6xza6q0uxwclled4wrf"
	authority2 = "tp1q3xhmqrjukjuhmccy4p6xza6q0uxwcll2xszpr"
)

// RegistryBulkUpdate registers, or updates, multiple NFTs in the registry.
// This creates multiple registry entries, or updates if one exists.
func (k msgServer) RegistryBulkUpdate(ctx context.Context, msg *types.MsgRegistryBulkUpdate) (*types.MsgRegistryBulkUpdateResponse, error) {
	for i, entry := range msg.Entries {
		// Validate that the NFT exists
		if err := k.ValidateNFTExists(ctx, &entry.Key.AssetClassId, &entry.Key.NftId); err != nil {
			return nil, fmt.Errorf("[%d]: %w", i, err)
		}

		// Validate that the signer owns the NFT
		if msg.Signer != authority1 && msg.Signer != authority2 {
			if err := k.ValidateNFTOwner(ctx, &entry.Key.AssetClassId, &entry.Key.NftId, msg.Signer); err != nil {
				return nil, fmt.Errorf("[%d]: %w", i, err)
			}
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

		// And finally, charge for one registry creation.
		antewrapper.ConsumeMsg(sdk.UnwrapSDKContext(ctx), &types.MsgRegisterNFT{})
	}

	return &types.MsgRegistryBulkUpdateResponse{}, nil
}
