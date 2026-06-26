package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MaxLenRegistryClassID is the maximum length of a registry class id.
const MaxLenRegistryClassID = 128

// Validate validates the RegistryClass.
func (m *RegistryClass) Validate() error {
	if m == nil {
		return fmt.Errorf("registry class cannot be nil")
	}

	var errs []error
	if err := ValidateRegistryClassID(m.RegistryClassId); err != nil {
		errs = append(errs, fmt.Errorf("registry_class_id: %w", err))
	}

	if err := ValidateClassID(m.AssetClassId); err != nil {
		errs = append(errs, fmt.Errorf("asset_class_id: %w", err))
	}

	if _, err := sdk.AccAddressFromBech32(m.Maintainer); err != nil {
		errs = append(errs, fmt.Errorf("maintainer: %w", err))
	}

	errs = append(errs, validateRoleAuthorizations(m.RoleAuthorizations)...)

	return errors.Join(errs...)
}

// ValidateRegistryClassID validates the registry class id format.
func ValidateRegistryClassID(registryClassID string) error {
	if err := ValidateStringLength(registryClassID, 1, MaxLenRegistryClassID); err != nil {
		return err
	}
	if !alNumDashRx.MatchString(registryClassID) {
		return fmt.Errorf("%q must only contain alphanumeric, '-', '.' characters", registryClassID)
	}
	return nil
}

// validateRoleAuthorizations validates a set of role authorization policies. Each policy must
// reference a known, specified role, a role may be configured at most once, and every authorization
// path's signature requirements and role selectors must be well-formed.
func validateRoleAuthorizations(auths []RoleAuthorization) []error {
	var errs []error
	seen := make(map[RegistryRole]bool, len(auths))
	for i, auth := range auths {
		if err := auth.Role.Validate(); err != nil {
			errs = append(errs, NewErrCodeInvalidField("role_authorizations", "%d: role %s", i, err))
		} else if seen[auth.Role] {
			errs = append(errs, NewErrCodeInvalidField("role_authorizations", "%d: duplicate role %s", i, auth.Role.ShortString()))
		} else {
			seen[auth.Role] = true
		}

		if len(auth.Authorizations) == 0 {
			errs = append(errs, NewErrCodeInvalidField("role_authorizations", "%d: at least one authorization path is required", i))
		}
		for j, path := range auth.Authorizations {
			errs = append(errs, validateAuthorizationPath(i, j, path)...)
		}
	}
	return errs
}

// validateAuthorizationPath validates a single authorization path (one of the "any path satisfies"
// alternatives) within a role authorization policy. i/j are the policy/path indices for error
// context.
func validateAuthorizationPath(i, j int, path *Authorization) []error {
	var errs []error
	if path == nil {
		return []error{NewErrCodeInvalidField("role_authorizations", "%d: authorization %d is nil", i, j)}
	}
	if len(path.Signatures) == 0 {
		errs = append(errs, NewErrCodeInvalidField("role_authorizations", "%d: authorization %d: at least one signature requirement is required", i, j))
	}
	for k, req := range path.Signatures {
		errs = append(errs, validateSignatureRequirement(i, j, k, req)...)
	}
	return errs
}

// validateSignatureRequirement validates a single signature requirement: a known signature type and
// at least one well-formed role selector. i/j/k are the policy/path/requirement indices.
func validateSignatureRequirement(i, j, k int, req *SignatureRequirement) []error {
	var errs []error
	if req == nil {
		return []error{NewErrCodeInvalidField("role_authorizations", "%d: authorization %d: signature %d is nil", i, j, k)}
	}
	if err := req.Type.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("role_authorizations", "%d: authorization %d: signature %d: type %s", i, j, k, err))
	}
	if len(req.Roles) == 0 {
		errs = append(errs, NewErrCodeInvalidField("role_authorizations", "%d: authorization %d: signature %d: at least one role selector is required", i, j, k))
	}
	for l, ra := range req.Roles {
		errs = append(errs, validateRoleAssignment(i, j, k, l, ra)...)
	}
	return errs
}

// validateRoleAssignment validates a single role selector within a signature requirement. It
// enforces that exactly one selector kind is set, that referenced enum values are defined, and that
// NftRole selectors only use the CURRENT* assignment family. i/j/k/l are the nested indices.
func validateRoleAssignment(i, j, k, l int, ra *RoleAssignment) []error {
	prefix := func(format string, args ...interface{}) error {
		return NewErrCodeInvalidField("role_authorizations",
			"%d: authorization %d: signature %d: role %d: "+format, append([]interface{}{i, j, k, l}, args...)...)
	}
	if ra == nil {
		return []error{prefix("role selector is nil")}
	}

	var errs []error
	if err := ra.Assignment.Validate(); err != nil {
		errs = append(errs, prefix("assignment %s", err))
	}

	switch r := ra.RoleSelector.(type) {
	case *RoleAssignment_RegistryRole:
		if err := r.RegistryRole.Validate(); err != nil {
			errs = append(errs, prefix("registry_role %s", err))
		}
	case *RoleAssignment_NftRole:
		if err := r.NftRole.Validate(); err != nil {
			errs = append(errs, prefix("nft_role %s", err))
		} else if !ra.Assignment.IsCurrent() {
			errs = append(errs, prefix("nft_role may only be used with a CURRENT* assignment, got %s", ra.Assignment.String()))
		}
	case *RoleAssignment_RolePriority:
		errs = append(errs, validateRolePriority(prefix, ra.Assignment, r.RolePriority)...)
	default:
		errs = append(errs, prefix("a role selector (registry_role, nft_role, or role_priority) must be set"))
	}
	return errs
}

// validateRolePriority validates a role_priority fallback chain: it must contain at least one entry,
// and every entry must set exactly one selector with a defined enum value. NftRole entries inherit
// the assignment's CURRENT* constraint.
func validateRolePriority(prefix func(string, ...interface{}) error, assignment Assignment, rp *RolePriority) []error {
	if rp == nil || len(rp.Entries) == 0 {
		return []error{prefix("role_priority must contain at least one entry")}
	}
	var errs []error
	for m, e := range rp.Entries {
		switch r := e.Role.(type) {
		case *RolePriorityEntry_RegistryRole:
			if err := r.RegistryRole.Validate(); err != nil {
				errs = append(errs, prefix("role_priority entry %d: registry_role %s", m, err))
			}
		case *RolePriorityEntry_NftRole:
			if err := r.NftRole.Validate(); err != nil {
				errs = append(errs, prefix("role_priority entry %d: nft_role %s", m, err))
			} else if !assignment.IsCurrent() {
				errs = append(errs, prefix("role_priority entry %d: nft_role may only be used with a CURRENT* assignment, got %s", m, assignment.String()))
			}
		default:
			errs = append(errs, prefix("role_priority entry %d: a role (registry_role or nft_role) must be set", m))
		}
	}
	return errs
}
