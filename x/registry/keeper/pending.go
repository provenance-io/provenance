package keeper

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"cosmossdk.io/collections"

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

// ProposeRoleChange opens (or re-uses) a pending role change and records the proposer's approval.
// If the accumulated approvals already satisfy the role's policy, the change is applied
// immediately. Returns the change id and whether it was applied.
func (k Keeper) ProposeRoleChange(ctx context.Context, proposer string, key *types.RegistryKey, role types.RegistryRole, op types.RoleChangeOperation, addresses []string) (string, bool, error) {
	entry, err := k.GetRegistry(ctx, key)
	if err != nil {
		return "", false, err
	}
	if entry == nil {
		return "", false, types.NewErrCodeRegistryNotFound(key.String())
	}

	if _, ok := types.RoleAuthorizationMap()[role]; !ok {
		return "", false, types.NewErrCodeUnauthorized(
			fmt.Sprintf("role %s has no authorization policy; propose/approve flow requires a policy-governed role", role.ShortString()),
		)
	}

	id := types.NewPendingRoleChangeID(key, role, op, addresses)

	change, err := k.GetPendingRoleChange(ctx, id)
	if err != nil {
		return "", false, err
	}
	if change == nil {
		change = &types.PendingRoleChange{
			Id:        id,
			Key:       key,
			Role:      role,
			Operation: op,
			Addresses: addresses,
			Proposer:  proposer,
		}
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
func (k Keeper) recordApprovalAndMaybeApply(ctx context.Context, entry *types.RegistryEntry, change *types.PendingRoleChange, approver string) (bool, error) {
	if !slices.Contains(change.Approvals, approver) {
		change.Approvals = append(change.Approvals, approver)
	}
	k.EmitEvent(ctx, types.NewEventRoleChangeApproved(change.Id, approver))

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

// pendingChangeSatisfied reports whether the accumulated approvals satisfy the role's policy.
func (k Keeper) pendingChangeSatisfied(ctx context.Context, entry *types.RegistryEntry, change *types.PendingRoleChange) bool {
	roleAuth, ok := types.RoleAuthorizationMap()[change.Role]
	if !ok {
		return false
	}

	// New addresses are only meaningful for grant operations (ASSIGNMENT_NEW checks).
	var newAddrs []string
	if change.Operation == types.RoleChangeOperation_ROLE_CHANGE_OPERATION_GRANT {
		newAddrs = change.Addresses
	}

	return k.ValidateRoleChangeAuthorization(ctx, roleAuth, entry, newAddrs, change.Approvals) == nil
}

// applyPendingChange performs the role grant or revoke described by the change.
func (k Keeper) applyPendingChange(ctx context.Context, change *types.PendingRoleChange) error {
	switch change.Operation {
	case types.RoleChangeOperation_ROLE_CHANGE_OPERATION_GRANT:
		return k.GrantRole(ctx, change.Key, change.Role, change.Addresses)
	case types.RoleChangeOperation_ROLE_CHANGE_OPERATION_REVOKE:
		return k.RevokeRole(ctx, change.Key, change.Role, change.Addresses)
	default:
		return types.NewErrCodeInvalidField("operation", "unsupported operation %s", change.Operation.ShortString())
	}
}
