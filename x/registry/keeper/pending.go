package keeper

import (
	"context"
	"errors"
	"slices"

	"cosmossdk.io/collections"

	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/registry/types"
)

// GetPendingRoleChange returns the pending role change for the given id, or nil if none exists.
func (k Keeper) GetPendingRoleChange(ctx context.Context, id string) (*types.PendingRoleChange, error) {
	change, err := k.PendingRoleChanges.Get(ctx, id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &change, nil
}

// SetPendingRoleChange stores a pending role change.
func (k Keeper) SetPendingRoleChange(ctx context.Context, change types.PendingRoleChange) error {
	return k.PendingRoleChanges.Set(ctx, change.Id, change)
}

// RemovePendingRoleChange deletes a pending role change by id.
func (k Keeper) RemovePendingRoleChange(ctx context.Context, id string) error {
	return k.PendingRoleChanges.Remove(ctx, id)
}

// GetPendingRoleChanges returns the pending role changes, optionally filtered to a single registry
// key. Results are paginated over the deterministic change-id keyspace.
func (k Keeper) GetPendingRoleChanges(ctx context.Context, pagination *query.PageRequest, key *types.RegistryKey) ([]types.PendingRoleChange, *query.PageResponse, error) {
	ptrs, pageRes, err := query.CollectionFilteredPaginate(ctx, k.PendingRoleChanges, pagination,
		func(_ string, change types.PendingRoleChange) (bool, error) {
			if key == nil {
				return true, nil
			}
			return change.Key != nil &&
				change.Key.AssetClassId == key.AssetClassId &&
				change.Key.NftId == key.NftId, nil
		},
		func(_ string, change types.PendingRoleChange) (*types.PendingRoleChange, error) {
			return &change, nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	changes := make([]types.PendingRoleChange, len(ptrs))
	for i, p := range ptrs {
		changes[i] = *p
	}
	return changes, pageRes, nil
}

// ProposeRoleChange opens (or re-uses) a pending role change for a batch of desired-state role
// updates and records the proposer's approval. If the accumulated approvals already satisfy every
// affected role's policy, the change is applied immediately. Returns the change id and whether it
// was applied.
func (k Keeper) ProposeRoleChange(ctx context.Context, proposer string, key *types.RegistryKey, roleUpdates []types.RoleUpdate) (string, bool, error) {
	entry, err := k.GetRegistry(ctx, key)
	if err != nil {
		return "", false, err
	}
	if entry == nil {
		return "", false, types.NewErrCodeRegistryNotFound(key.String())
	}

	id := types.NewPendingRoleChangeID(key, roleUpdates)

	change, err := k.GetPendingRoleChange(ctx, id)
	if err != nil {
		return "", false, err
	}
	if change == nil {
		newChange := &types.PendingRoleChange{
			Id:          id,
			Key:         key,
			RoleUpdates: roleUpdates,
			Proposer:    proposer,
		}
		// Only open a new pending change when the proposer could actually contribute a valid
		// approval to at least one affected role. This keeps state from growing via proposals
		// opened by accounts that are not a required party for the change.
		if !k.approverEligible(ctx, entry, newChange, proposer) {
			return "", false, types.NewErrCodeUnauthorized(
				"proposer " + proposer + " is not eligible to approve any affected role",
			)
		}
		change = newChange
		k.EmitEvent(ctx, types.NewEventRoleChangeProposed(change))
	}

	applied, err := k.recordApprovalAndMaybeApply(ctx, entry, change, proposer)
	if err != nil {
		return "", false, err
	}
	return id, applied, nil
}

// ApproveRoleChange records an approval on an existing pending role change and applies it if the
// accumulated approvals now satisfy the role's policy. Returns whether the change was applied.
func (k Keeper) ApproveRoleChange(ctx context.Context, approver string, changeID string) (bool, error) {
	change, err := k.GetPendingRoleChange(ctx, changeID)
	if err != nil {
		return false, err
	}
	if change == nil {
		return false, types.NewErrCodePendingChangeNotFound(changeID)
	}

	entry, err := k.GetRegistry(ctx, change.Key)
	if err != nil {
		return false, err
	}
	if entry == nil {
		// The registry was removed out from under the pending change; clean it up.
		if rerr := k.RemovePendingRoleChange(ctx, changeID); rerr != nil {
			return false, rerr
		}
		return false, types.NewErrCodeRegistryNotFound(change.Key.String())
	}

	return k.recordApprovalAndMaybeApply(ctx, entry, change, approver)
}

// recordApprovalAndMaybeApply adds approver to the change's approval set, evaluates the role's
// authorization policy against the accumulated approvals, and either applies and removes the
// change (returns true) or persists the updated change (returns false).
//
// Approvals from addresses that are not eligible under any affected role's policy are silently
// ignored: they can never contribute to satisfying the change, so recording them would only let a
// third party grow the approval set without bound (state bloat / DoS).
func (k Keeper) recordApprovalAndMaybeApply(ctx context.Context, entry *types.RegistryEntry, change *types.PendingRoleChange, approver string) (bool, error) {
	if k.approverEligible(ctx, entry, change, approver) && !slices.Contains(change.Approvals, approver) {
		change.Approvals = append(change.Approvals, approver)
		k.EmitEvent(ctx, types.NewEventRoleChangeApproved(change.Id, approver))
	}

	if !k.pendingChangeSatisfied(ctx, entry, change) {
		if err := k.SetPendingRoleChange(ctx, *change); err != nil {
			return false, err
		}
		return false, nil
	}

	if err := k.applyPendingChange(ctx, change); err != nil {
		return false, err
	}
	if err := k.RemovePendingRoleChange(ctx, change.Id); err != nil {
		return false, err
	}
	k.EmitEvent(ctx, types.NewEventRoleChangeApplied(change))
	return true, nil
}

// approverEligible reports whether approver could contribute to satisfying any of the change's
// role-update gates: for a policy-governed role they must be one of the policy's referenced
// addresses, and for a non-policy role they must own the NFT (the legacy fallback).
func (k Keeper) approverEligible(ctx context.Context, entry *types.RegistryEntry, change *types.PendingRoleChange, approver string) bool {
	roleAuths := types.RoleAuthorizationMap()
	for _, update := range change.RoleUpdates {
		roleAuth, ok := roleAuths[update.Role]
		if !ok {
			// Non-policy role: only the NFT owner can satisfy the fallback.
			if k.ValidateNFTOwner(ctx, &change.Key.AssetClassId, &change.Key.NftId, approver) == nil {
				return true
			}
			continue
		}

		newAddrs := additions(entry.GetRoleAddrs(update.Role), update.Addresses)
		if k.CollectPolicyApprovers(ctx, roleAuth, entry, newAddrs)[approver] {
			return true
		}
	}
	return false
}

// pendingChangeSatisfied reports whether the accumulated approvals satisfy every affected role's
// authorization policy. A role with no policy falls back to NFT ownership: at least one approver
// must own the NFT. The change applies only when all role updates are satisfied (atomic gate).
func (k Keeper) pendingChangeSatisfied(ctx context.Context, entry *types.RegistryEntry, change *types.PendingRoleChange) bool {
	roleAuths := types.RoleAuthorizationMap()
	for _, update := range change.RoleUpdates {
		roleAuth, ok := roleAuths[update.Role]
		if !ok {
			// Legacy fallback: at least one approver must own the NFT.
			if !k.anyApproverOwnsNFT(ctx, change) {
				return false
			}
			continue
		}

		// Only addresses being newly added to the role are relevant for ASSIGNMENT_NEW checks.
		newAddrs := additions(entry.GetRoleAddrs(update.Role), update.Addresses)
		if k.ValidateRoleChangeAuthorization(ctx, roleAuth, entry, newAddrs, change.Approvals) != nil {
			return false
		}
	}
	return true
}

// anyApproverOwnsNFT reports whether any accumulated approver owns the change's NFT.
func (k Keeper) anyApproverOwnsNFT(ctx context.Context, change *types.PendingRoleChange) bool {
	for _, approver := range change.Approvals {
		if k.ValidateNFTOwner(ctx, &change.Key.AssetClassId, &change.Key.NftId, approver) == nil {
			return true
		}
	}
	return false
}

// additions returns the addresses in desired that are not already present in current.
func additions(current, desired []string) []string {
	if len(current) == 0 {
		return desired
	}
	have := make(map[string]bool, len(current))
	for _, a := range current {
		have[a] = true
	}
	var added []string
	for _, a := range desired {
		if !have[a] {
			added = append(added, a)
		}
	}
	return added
}

// applyPendingChange applies the batch of role updates as a single atomic desired-state set.
func (k Keeper) applyPendingChange(ctx context.Context, change *types.PendingRoleChange) error {
	return k.SetRoles(ctx, change.Key, change.RoleUpdates)
}
