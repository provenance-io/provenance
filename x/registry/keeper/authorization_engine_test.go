package keeper_test

import (
	"github.com/provenance-io/provenance/x/registry/types"
)

// These tests exercise the policy engine's fail-closed branches directly via
// ValidateRoleChangeAuthorization. A misconfigured policy (unspecified enums, singular assignments
// resolving to more than one address, selectors used in unsupported ways) must surface an error and
// must NEVER be silently treated as satisfied.

// engineEntry builds an in-memory registry entry (owned by s.user1) seeded with the given roles.
func (s *KeeperTestSuite) engineEntry(roles []types.RolesEntry) *types.RegistryEntry {
	return &types.RegistryEntry{
		Key:   &types.RegistryKey{AssetClassId: s.validNFTClass.Id, NftId: s.validNFT.Id},
		Roles: roles,
	}
}

// singleSigPolicy wraps one role selector in a REQUIRED_ALL single-path policy for the SERVICER role.
func singleSigPolicy(ra *types.RoleAssignment) types.RoleAuthorization {
	return types.RoleAuthorization{
		Role: types.RegistryRole_REGISTRY_ROLE_SERVICER,
		Authorizations: []*types.Authorization{{
			Description: "engine fail-closed path",
			Signatures: []*types.SignatureRequirement{{
				Type:  types.SignatureType_SIGNATURE_TYPE_REQUIRED_ALL,
				Roles: []*types.RoleAssignment{ra},
			}},
		}},
	}
}

func (s *KeeperTestSuite) TestEngine_AssignmentCurrentResolvesMultiple() {
	// CONTROLLER currently held by two addresses; singular ASSIGNMENT_CURRENT must reject.
	entry := s.engineEntry([]types.RolesEntry{
		{Role: types.RegistryRole_REGISTRY_ROLE_CONTROLLER, Addresses: []string{s.user1, s.user2}},
	})
	policy := singleSigPolicy(&types.RoleAssignment{
		RoleSelector: &types.RoleAssignment_RegistryRole{RegistryRole: types.RegistryRole_REGISTRY_ROLE_CONTROLLER},
		Assignment:   types.Assignment_ASSIGNMENT_CURRENT,
	})

	err := s.app.RegistryKeeper.ValidateRoleChangeAuthorization(s.ctx, policy, entry, nil, []string{s.user1, s.user2})
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "ASSIGNMENT_CURRENT for role CONTROLLER resolved to 2 addresses, expected exactly one")
}

func (s *KeeperTestSuite) TestEngine_AssignmentNewResolvesMultiple() {
	entry := s.engineEntry(nil)
	policy := singleSigPolicy(&types.RoleAssignment{
		RoleSelector: &types.RoleAssignment_RegistryRole{RegistryRole: types.RegistryRole_REGISTRY_ROLE_SERVICER},
		Assignment:   types.Assignment_ASSIGNMENT_NEW,
	})

	newAddrs := []string{s.user1, s.user2}
	err := s.app.RegistryKeeper.ValidateRoleChangeAuthorization(s.ctx, policy, entry, newAddrs, newAddrs)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "ASSIGNMENT_NEW for role SERVICER resolved to 2 addresses, expected exactly one")
}

func (s *KeeperTestSuite) TestEngine_UnspecifiedAssignment() {
	entry := s.engineEntry([]types.RolesEntry{
		{Role: types.RegistryRole_REGISTRY_ROLE_CONTROLLER, Addresses: []string{s.user1}},
	})
	policy := singleSigPolicy(&types.RoleAssignment{
		RoleSelector: &types.RoleAssignment_RegistryRole{RegistryRole: types.RegistryRole_REGISTRY_ROLE_CONTROLLER},
		Assignment:   types.Assignment_ASSIGNMENT_UNSPECIFIED,
	})

	err := s.app.RegistryKeeper.ValidateRoleChangeAuthorization(s.ctx, policy, entry, nil, []string{s.user1})
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "unsupported assignment")
}

func (s *KeeperTestSuite) TestEngine_NftRoleWithNonCurrentAssignment() {
	entry := s.engineEntry(nil)
	policy := singleSigPolicy(&types.RoleAssignment{
		RoleSelector: &types.RoleAssignment_NftRole{NftRole: types.NftRole_NFT_ROLE_NFT_OWNER},
		Assignment:   types.Assignment_ASSIGNMENT_NEW,
	})

	err := s.app.RegistryKeeper.ValidateRoleChangeAuthorization(s.ctx, policy, entry, nil, []string{s.user1})
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "NftRole may only be used with ASSIGNMENT_CURRENT* variants")
}

func (s *KeeperTestSuite) TestEngine_UnsupportedNftRole() {
	entry := s.engineEntry(nil)
	policy := singleSigPolicy(&types.RoleAssignment{
		RoleSelector: &types.RoleAssignment_NftRole{NftRole: types.NftRole_NFT_ROLE_UNSPECIFIED},
		Assignment:   types.Assignment_ASSIGNMENT_CURRENT,
	})

	err := s.app.RegistryKeeper.ValidateRoleChangeAuthorization(s.ctx, policy, entry, nil, []string{s.user1})
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "unsupported NftRole")
}

func (s *KeeperTestSuite) TestEngine_NoRoleSelectorSet() {
	entry := s.engineEntry(nil)
	policy := singleSigPolicy(&types.RoleAssignment{
		Assignment: types.Assignment_ASSIGNMENT_CURRENT,
	})

	err := s.app.RegistryKeeper.ValidateRoleChangeAuthorization(s.ctx, policy, entry, nil, []string{s.user1})
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "role assignment has no role selector set")
}

func (s *KeeperTestSuite) TestEngine_UnsupportedSignatureType() {
	entry := s.engineEntry([]types.RolesEntry{
		{Role: types.RegistryRole_REGISTRY_ROLE_CONTROLLER, Addresses: []string{s.user1}},
	})
	policy := types.RoleAuthorization{
		Role: types.RegistryRole_REGISTRY_ROLE_SERVICER,
		Authorizations: []*types.Authorization{{
			Description: "bad signature type",
			Signatures: []*types.SignatureRequirement{{
				Type: types.SignatureType_SIGNATURE_TYPE_UNSPECIFIED,
				Roles: []*types.RoleAssignment{{
					RoleSelector: &types.RoleAssignment_RegistryRole{RegistryRole: types.RegistryRole_REGISTRY_ROLE_CONTROLLER},
					Assignment:   types.Assignment_ASSIGNMENT_CURRENT,
				}},
			}},
		}},
	}

	err := s.app.RegistryKeeper.ValidateRoleChangeAuthorization(s.ctx, policy, entry, nil, []string{s.user1})
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "unsupported signature requirement type")
}

// TestEngine_RequiredAnyIfSet_NoRolesSet confirms REQUIRED_ANY_IF_SET passes (vacuously) when none
// of its roles resolve to addresses, while REQUIRED_ANY rejects in the same situation. This is the
// empty-role distinction between the two ANY signature types.
func (s *KeeperTestSuite) TestEngine_RequiredAnyIfSet_NoRolesSet() {
	// Entry has no SECURED_PARTY_FOR_ENOTE set, so the selector resolves empty.
	entry := s.engineEntry(nil)
	selector := &types.RoleAssignment{
		RoleSelector: &types.RoleAssignment_RegistryRole{RegistryRole: types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE},
		Assignment:   types.Assignment_ASSIGNMENT_CURRENT_ANY,
	}

	s.Run("required_any_if_set passes when no role resolves", func() {
		policy := types.RoleAuthorization{
			Role: types.RegistryRole_REGISTRY_ROLE_SERVICER,
			Authorizations: []*types.Authorization{{
				Signatures: []*types.SignatureRequirement{{
					Type:  types.SignatureType_SIGNATURE_TYPE_REQUIRED_ANY_IF_SET,
					Roles: []*types.RoleAssignment{selector},
				}},
			}},
		}
		err := s.app.RegistryKeeper.ValidateRoleChangeAuthorization(s.ctx, policy, entry, nil, nil)
		s.Require().NoError(err)
	})

	s.Run("required_any rejects when no role resolves", func() {
		policy := types.RoleAuthorization{
			Role: types.RegistryRole_REGISTRY_ROLE_SERVICER,
			Authorizations: []*types.Authorization{{
				Signatures: []*types.SignatureRequirement{{
					Type:  types.SignatureType_SIGNATURE_TYPE_REQUIRED_ANY,
					Roles: []*types.RoleAssignment{selector},
				}},
			}},
		}
		err := s.app.RegistryKeeper.ValidateRoleChangeAuthorization(s.ctx, policy, entry, nil, nil)
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "no required signature found in any role")
	})
}
