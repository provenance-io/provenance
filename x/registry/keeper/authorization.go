package keeper

import (
	"context"
	"fmt"
	"strings"

	"github.com/provenance-io/provenance/x/registry/types"
)

// resolveRegistryRoleAddresses returns the addresses for a registry role under the given
// assignment. CURRENT* variants resolve the addresses currently held for the role; NEW* variants
// resolve the incoming new addresses.
//
// The singular ASSIGNMENT_CURRENT and ASSIGNMENT_NEW variants are documented to resolve exactly
// one address and return an error if more than one would resolve.
func (k Keeper) resolveRegistryRoleAddresses(entry *types.RegistryEntry, role types.RegistryRole, assignment types.Assignment, newAddrs []string) ([]string, bool, error) {
	switch assignment {
	case types.Assignment_ASSIGNMENT_CURRENT,
		types.Assignment_ASSIGNMENT_CURRENT_ALL,
		types.Assignment_ASSIGNMENT_CURRENT_ANY:
		var current []string
		for _, re := range entry.Roles {
			if re.Role == role {
				current = re.Addresses
				break
			}
		}
		if assignment == types.Assignment_ASSIGNMENT_CURRENT && len(current) > 1 {
			return nil, false, types.NewErrCodeUnauthorized(
				fmt.Sprintf("ASSIGNMENT_CURRENT for role %s resolved to %d addresses, expected exactly one", role.ShortString(), len(current)),
			)
		}
		return current, len(current) > 0, nil
	case types.Assignment_ASSIGNMENT_NEW,
		types.Assignment_ASSIGNMENT_NEW_ALL,
		types.Assignment_ASSIGNMENT_NEW_ANY:
		if assignment == types.Assignment_ASSIGNMENT_NEW && len(newAddrs) > 1 {
			return nil, false, types.NewErrCodeUnauthorized(
				fmt.Sprintf("ASSIGNMENT_NEW for role %s resolved to %d addresses, expected exactly one", role.ShortString(), len(newAddrs)),
			)
		}
		return newAddrs, len(newAddrs) > 0, nil
	default:
		return nil, false, nil
	}
}

// resolveNFTRoleAddresses returns the address(es) for an NFT-module role (e.g. the NFT owner).
// NftRole is read-only from the registry's perspective and may only be used with the
// ASSIGNMENT_CURRENT* variants; any other assignment is a policy misconfiguration and returns an
// error.
func (k Keeper) resolveNFTRoleAddresses(ctx context.Context, entry *types.RegistryEntry, nftRole types.NftRole, assignment types.Assignment) ([]string, bool, error) {
	switch assignment {
	case types.Assignment_ASSIGNMENT_CURRENT,
		types.Assignment_ASSIGNMENT_CURRENT_ALL,
		types.Assignment_ASSIGNMENT_CURRENT_ANY:
		// NftRole is only valid with the CURRENT* assignment variants.
	default:
		return nil, false, types.NewErrCodeUnauthorized(
			fmt.Sprintf("NftRole may only be used with ASSIGNMENT_CURRENT* variants, got %s", assignment.String()),
		)
	}

	switch nftRole {
	case types.NftRole_NFT_ROLE_NFT_OWNER:
		owner := k.GetNFTOwner(ctx, &entry.Key.AssetClassId, &entry.Key.NftId)
		if len(owner) == 0 {
			return nil, false, nil
		}
		return []string{owner.String()}, true, nil
	default:
		return nil, false, nil
	}
}

// resolveRolePriorityAddresses iterates priority entries and returns addresses from the first
// non-empty role that exists, implementing the "first match wins" fallback chain.
func (k Keeper) resolveRolePriorityAddresses(ctx context.Context, entry *types.RegistryEntry, rp *types.RolePriority, assignment types.Assignment, newAddrs []string) ([]string, bool, error) {
	if rp == nil {
		return nil, false, nil
	}
	for _, e := range rp.Entries {
		switch r := e.Role.(type) {
		case *types.RolePriorityEntry_RegistryRole:
			addrs, exists, err := k.resolveRegistryRoleAddresses(entry, r.RegistryRole, assignment, newAddrs)
			if err != nil {
				return nil, false, err
			}
			if exists {
				return addrs, true, nil
			}
		case *types.RolePriorityEntry_NftRole:
			addrs, exists, err := k.resolveNFTRoleAddresses(ctx, entry, r.NftRole, assignment)
			if err != nil {
				return nil, false, err
			}
			if exists {
				return addrs, true, nil
			}
		}
	}
	return nil, false, nil
}

// resolveRoleAssignmentAddresses dispatches to the correct resolver based on the role selector type.
func (k Keeper) resolveRoleAssignmentAddresses(ctx context.Context, entry *types.RegistryEntry, ra *types.RoleAssignment, newAddrs []string) ([]string, bool, error) {
	if ra == nil {
		return nil, false, nil
	}
	switch r := ra.RoleSelector.(type) {
	case *types.RoleAssignment_RegistryRole:
		return k.resolveRegistryRoleAddresses(entry, r.RegistryRole, ra.Assignment, newAddrs)
	case *types.RoleAssignment_NftRole:
		return k.resolveNFTRoleAddresses(ctx, entry, r.NftRole, ra.Assignment)
	case *types.RoleAssignment_RolePriority:
		return k.resolveRolePriorityAddresses(ctx, entry, r.RolePriority, ra.Assignment, newAddrs)
	default:
		return nil, false, nil
	}
}

// evaluateSignatureRequirement checks whether a single SignatureRequirement is satisfied.
//
//   - REQUIRED_ALL: every resolved address from every role must be in signerSet.
//   - REQUIRED_ALL_IF_SET: for each role that resolves to non-empty addresses, all must sign;
//     roles that resolve to empty are skipped.
//   - REQUIRED_ANY: at least one address from any role must be in signerSet.
//   - REQUIRED_ANY_IF_SET: same as REQUIRED_ANY but empty roles are skipped.
func (k Keeper) evaluateSignatureRequirement(ctx context.Context, entry *types.RegistryEntry, req *types.SignatureRequirement, newAddrs []string, signerSet map[string]bool) error {
	if req == nil {
		return nil
	}
	switch req.Type {
	case types.SignatureType_SIGNATURE_TYPE_REQUIRED_ALL:
		for _, ra := range req.Roles {
			addrs, exists, err := k.resolveRoleAssignmentAddresses(ctx, entry, ra, newAddrs)
			if err != nil {
				return err
			}
			if !exists || len(addrs) == 0 {
				return types.NewErrCodeUnauthorized("required role resolved to no addresses")
			}
			for _, addr := range addrs {
				if !signerSet[addr] {
					return types.NewErrCodeUnauthorized(fmt.Sprintf("missing required signature for %q", addr))
				}
			}
		}

	case types.SignatureType_SIGNATURE_TYPE_REQUIRED_ALL_IF_SET:
		for _, ra := range req.Roles {
			addrs, exists, err := k.resolveRoleAssignmentAddresses(ctx, entry, ra, newAddrs)
			if err != nil {
				return err
			}
			if !exists || len(addrs) == 0 {
				continue
			}
			for _, addr := range addrs {
				if !signerSet[addr] {
					return types.NewErrCodeUnauthorized(fmt.Sprintf("missing required signature for %q", addr))
				}
			}
		}

	case types.SignatureType_SIGNATURE_TYPE_REQUIRED_ANY:
		found := false
		for _, ra := range req.Roles {
			addrs, _, err := k.resolveRoleAssignmentAddresses(ctx, entry, ra, newAddrs)
			if err != nil {
				return err
			}
			for _, addr := range addrs {
				if signerSet[addr] {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return types.NewErrCodeUnauthorized("no required signature found in any role")
		}

	case types.SignatureType_SIGNATURE_TYPE_REQUIRED_ANY_IF_SET:
		hasAnyRole := false
		found := false
		for _, ra := range req.Roles {
			addrs, exists, err := k.resolveRoleAssignmentAddresses(ctx, entry, ra, newAddrs)
			if err != nil {
				return err
			}
			if !exists || len(addrs) == 0 {
				continue
			}
			hasAnyRole = true
			for _, addr := range addrs {
				if signerSet[addr] {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if hasAnyRole && !found {
			return types.NewErrCodeUnauthorized("no required signature found in any role")
		}
	}
	return nil
}

// evaluateAuthorization checks whether every SignatureRequirement in an authorization path is
// satisfied. Returns nil if all requirements pass, or an error describing the first failure.
func (k Keeper) evaluateAuthorization(ctx context.Context, entry *types.RegistryEntry, auth *types.Authorization, newAddrs []string, signerSet map[string]bool) error {
	if auth == nil {
		return nil
	}
	for i, req := range auth.Signatures {
		if err := k.evaluateSignatureRequirement(ctx, entry, req, newAddrs, signerSet); err != nil {
			return fmt.Errorf("signature requirement %d: %w", i, err)
		}
	}
	return nil
}

// ValidateRoleChangeAuthorization validates that the provided signers are authorized to update a
// role on the given entry. It tries each authorization path in the RoleAuthorization; the
// operation is approved as soon as any single path is fully satisfied.
//
// newAddrs contains the addresses being assigned to the role (used for ASSIGNMENT_NEW checks).
// For revoke operations, newAddrs should be nil or empty.
func (k Keeper) ValidateRoleChangeAuthorization(ctx context.Context, roleAuth types.RoleAuthorization, entry *types.RegistryEntry, newAddrs []string, signers []string) error {
	signerSet := make(map[string]bool, len(signers))
	for _, s := range signers {
		signerSet[s] = true
	}

	var pathErrors []string
	for _, auth := range roleAuth.Authorizations {
		if err := k.evaluateAuthorization(ctx, entry, auth, newAddrs, signerSet); err == nil {
			return nil
		} else {
			desc := auth.Description
			if desc == "" {
				desc = fmt.Sprintf("path %d", len(pathErrors)+1)
			}
			pathErrors = append(pathErrors, fmt.Sprintf("%s: %v", desc, err))
		}
	}

	return types.NewErrCodeUnauthorized(
		fmt.Sprintf("no valid authorization path found for %s: %s", roleAuth.Role.ShortString(), strings.Join(pathErrors, "; ")),
	)
}

// CollectPolicyApprovers returns the set of addresses that the given role authorization policy
// could ever require to sign, resolved against the entry and the addresses being newly assigned.
// These are the only approvers whose signature can contribute to satisfying the policy; any other
// address is, by definition, a no-op approver. Resolution errors from a misconfigured path are
// ignored here (the path simply contributes no eligible addresses).
func (k Keeper) CollectPolicyApprovers(ctx context.Context, roleAuth types.RoleAuthorization, entry *types.RegistryEntry, newAddrs []string) map[string]bool {
	out := make(map[string]bool)
	for _, auth := range roleAuth.Authorizations {
		if auth == nil {
			continue
		}
		for _, req := range auth.Signatures {
			if req == nil {
				continue
			}
			for _, ra := range req.Roles {
				addrs, _, _ := k.resolveRoleAssignmentAddresses(ctx, entry, ra, newAddrs)
				for _, addr := range addrs {
					out[addr] = true
				}
			}
		}
	}
	return out
}
