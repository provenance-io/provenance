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
// reference a known, specified role, and a role may be configured at most once.
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
	}
	return errs
}
