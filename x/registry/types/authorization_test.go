package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/provenance-io/provenance/x/registry/types"
)

func TestSignatureType_Validate(t *testing.T) {
	tests := []struct {
		name   string
		typ    types.SignatureType
		expErr string
	}{
		{name: "required_all", typ: types.SignatureType_SIGNATURE_TYPE_REQUIRED_ALL},
		{name: "required_any", typ: types.SignatureType_SIGNATURE_TYPE_REQUIRED_ANY},
		{name: "unspecified", typ: types.SignatureType_SIGNATURE_TYPE_UNSPECIFIED, expErr: "cannot be unspecified"},
		{name: "undefined value", typ: types.SignatureType(9999), expErr: "unknown"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.typ.Validate()
			if tc.expErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErr)
			}
		})
	}
}

func TestNftRole_Validate(t *testing.T) {
	tests := []struct {
		name   string
		role   types.NftRole
		expErr string
	}{
		{name: "nft_owner", role: types.NftRole_NFT_ROLE_NFT_OWNER},
		{name: "unspecified", role: types.NftRole_NFT_ROLE_UNSPECIFIED, expErr: "cannot be unspecified"},
		{name: "undefined value", role: types.NftRole(9999), expErr: "unknown"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.role.Validate()
			if tc.expErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErr)
			}
		})
	}
}

func TestAssignment_Validate(t *testing.T) {
	tests := []struct {
		name       string
		assignment types.Assignment
		expErr     string
	}{
		{name: "current", assignment: types.Assignment_ASSIGNMENT_CURRENT},
		{name: "new", assignment: types.Assignment_ASSIGNMENT_NEW},
		{name: "unspecified", assignment: types.Assignment_ASSIGNMENT_UNSPECIFIED, expErr: "cannot be unspecified"},
		{name: "undefined value", assignment: types.Assignment(9999), expErr: "unknown"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.assignment.Validate()
			if tc.expErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErr)
			}
		})
	}
}

func TestAssignment_IsCurrent(t *testing.T) {
	currentVariants := []types.Assignment{
		types.Assignment_ASSIGNMENT_CURRENT,
		types.Assignment_ASSIGNMENT_CURRENT_ALL,
		types.Assignment_ASSIGNMENT_CURRENT_ANY,
	}
	for _, a := range currentVariants {
		require.True(t, a.IsCurrent(), "%s should be current", a)
	}

	nonCurrentVariants := []types.Assignment{
		types.Assignment_ASSIGNMENT_UNSPECIFIED,
		types.Assignment_ASSIGNMENT_NEW,
		types.Assignment_ASSIGNMENT_NEW_ALL,
		types.Assignment_ASSIGNMENT_NEW_ANY,
	}
	for _, a := range nonCurrentVariants {
		require.False(t, a.IsCurrent(), "%s should not be current", a)
	}
}

// TestControllerRoleAuthorizations confirms the shipped example CONTROLLER policy is well-formed:
// it is the fixture governance can install via params or a maintainer can adopt for a class, so it
// must pass the same validation any on-chain class policy must pass.
func TestControllerRoleAuthorizations(t *testing.T) {
	auths := types.ControllerRoleAuthorizations()
	require.Len(t, auths, 1)
	require.Equal(t, types.RegistryRole_REGISTRY_ROLE_CONTROLLER, auths[0].Role)

	// Validate it via the same path a registry class uses.
	errs := validateAuthsViaClass(t, auths)
	require.NoError(t, errs)
}

// validateAuthsViaClass runs the given role authorizations through RegistryClass.Validate so the
// example policy is exercised by the production validation code.
func validateAuthsViaClass(t *testing.T, auths []types.RoleAuthorization) error {
	t.Helper()
	c := validRegistryClass()
	c.RoleAuthorizations = auths
	return c.Validate()
}

func TestRoleAuthorizationMapFrom(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		m := types.RoleAuthorizationMapFrom(nil)
		require.Empty(t, m)
	})

	t.Run("indexed by role", func(t *testing.T) {
		auths := []types.RoleAuthorization{
			{Role: types.RegistryRole_REGISTRY_ROLE_CONTROLLER},
			{Role: types.RegistryRole_REGISTRY_ROLE_SERVICER},
		}
		m := types.RoleAuthorizationMapFrom(auths)
		require.Len(t, m, 2)
		require.Contains(t, m, types.RegistryRole_REGISTRY_ROLE_CONTROLLER)
		require.Contains(t, m, types.RegistryRole_REGISTRY_ROLE_SERVICER)
	})

	t.Run("last entry wins on duplicate role", func(t *testing.T) {
		auths := []types.RoleAuthorization{
			{Role: types.RegistryRole_REGISTRY_ROLE_CONTROLLER, Authorizations: []*types.Authorization{{Description: "first"}}},
			{Role: types.RegistryRole_REGISTRY_ROLE_CONTROLLER, Authorizations: []*types.Authorization{{Description: "second"}}},
		}
		m := types.RoleAuthorizationMapFrom(auths)
		require.Len(t, m, 1)
		require.Equal(t, "second", m[types.RegistryRole_REGISTRY_ROLE_CONTROLLER].Authorizations[0].Description)
	})
}
