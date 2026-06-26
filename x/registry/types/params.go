package types

import (
	"errors"
)

// DefaultParams returns the default registry module parameters. By default the module defines no
// role authorization policies, which preserves the original chain behavior: every role change
// falls back to legacy NFT-owner authorization. Governance can install policies via
// MsgUpdateParams, and registry-class maintainers can define per-asset-class policies.
func DefaultParams() Params {
	return Params{}
}

// Validate validates the Params.
func (p Params) Validate() error {
	if errs := validateRoleAuthorizations(p.RoleAuthorizations); len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// RoleAuthorizationMap returns a map of RegistryRole -> RoleAuthorization for the params' default
// policies, for fast lookup during authorization resolution.
func (p Params) RoleAuthorizationMap() map[RegistryRole]RoleAuthorization {
	return RoleAuthorizationMapFrom(p.RoleAuthorizations)
}
