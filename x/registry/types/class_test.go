package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/registry/types"
)

// validRoleAuthorization builds a minimal, well-formed SERVICER policy: a single authorization path
// requiring the current controller to sign. It is used as the valid baseline that the negative
// cases mutate.
func validRoleAuthorization() types.RoleAuthorization {
	return types.RoleAuthorization{
		Role: types.RegistryRole_REGISTRY_ROLE_SERVICER,
		Authorizations: []*types.Authorization{
			{
				Description: "controller may manage servicers",
				Signatures: []*types.SignatureRequirement{
					{
						Type: types.SignatureType_SIGNATURE_TYPE_REQUIRED_ALL,
						Roles: []*types.RoleAssignment{
							{
								RoleSelector: &types.RoleAssignment_RegistryRole{
									RegistryRole: types.RegistryRole_REGISTRY_ROLE_CONTROLLER,
								},
								Assignment: types.Assignment_ASSIGNMENT_CURRENT,
							},
						},
					},
				},
			},
		},
	}
}

// validRegistryClass returns a well-formed RegistryClass that each negative case mutates.
func validRegistryClass() *types.RegistryClass {
	return &types.RegistryClass{
		RegistryClassId:    "loan-registry-v1",
		AssetClassId:       "asset-class-1",
		Maintainer:         sdk.AccAddress("class_maintainer_addr_").String(),
		RoleAuthorizations: []types.RoleAuthorization{validRoleAuthorization()},
	}
}

func TestRegistryClass_Validate(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(c *types.RegistryClass)
		expErr string
	}{
		{
			name:   "valid",
			mutate: func(_ *types.RegistryClass) {},
		},
		{
			name:   "empty registry_class_id",
			mutate: func(c *types.RegistryClass) { c.RegistryClassId = "" },
			expErr: "registry_class_id",
		},
		{
			name:   "registry_class_id with illegal character",
			mutate: func(c *types.RegistryClass) { c.RegistryClassId = "bad id!" },
			expErr: "alphanumeric",
		},
		{
			name:   "empty asset_class_id",
			mutate: func(c *types.RegistryClass) { c.AssetClassId = "" },
			expErr: "asset_class_id",
		},
		{
			name:   "invalid maintainer",
			mutate: func(c *types.RegistryClass) { c.Maintainer = "not-bech32" },
			expErr: "maintainer",
		},
		{
			name: "duplicate role",
			mutate: func(c *types.RegistryClass) {
				c.RoleAuthorizations = []types.RoleAuthorization{validRoleAuthorization(), validRoleAuthorization()}
			},
			expErr: "duplicate role",
		},
		{
			name: "unspecified role",
			mutate: func(c *types.RegistryClass) {
				c.RoleAuthorizations[0].Role = types.RegistryRole_REGISTRY_ROLE_UNSPECIFIED
			},
			expErr: "role",
		},
		{
			name: "no authorization paths",
			mutate: func(c *types.RegistryClass) {
				c.RoleAuthorizations[0].Authorizations = nil
			},
			expErr: "at least one authorization path is required",
		},
		{
			name: "nil authorization path",
			mutate: func(c *types.RegistryClass) {
				c.RoleAuthorizations[0].Authorizations = []*types.Authorization{nil}
			},
			expErr: "authorization 0 is nil",
		},
		{
			name: "no signature requirements",
			mutate: func(c *types.RegistryClass) {
				c.RoleAuthorizations[0].Authorizations[0].Signatures = nil
			},
			expErr: "at least one signature requirement is required",
		},
		{
			name: "nil signature requirement",
			mutate: func(c *types.RegistryClass) {
				c.RoleAuthorizations[0].Authorizations[0].Signatures = []*types.SignatureRequirement{nil}
			},
			expErr: "signature 0 is nil",
		},
		{
			name: "unspecified signature type",
			mutate: func(c *types.RegistryClass) {
				c.RoleAuthorizations[0].Authorizations[0].Signatures[0].Type = types.SignatureType_SIGNATURE_TYPE_UNSPECIFIED
			},
			expErr: "type",
		},
		{
			name: "no role selectors",
			mutate: func(c *types.RegistryClass) {
				c.RoleAuthorizations[0].Authorizations[0].Signatures[0].Roles = nil
			},
			expErr: "at least one role selector is required",
		},
		{
			name: "nil role selector",
			mutate: func(c *types.RegistryClass) {
				c.RoleAuthorizations[0].Authorizations[0].Signatures[0].Roles = []*types.RoleAssignment{nil}
			},
			expErr: "role selector is nil",
		},
		{
			name: "unspecified assignment",
			mutate: func(c *types.RegistryClass) {
				c.RoleAuthorizations[0].Authorizations[0].Signatures[0].Roles[0].Assignment = types.Assignment_ASSIGNMENT_UNSPECIFIED
			},
			expErr: "assignment",
		},
		{
			name: "unspecified registry_role selector",
			mutate: func(c *types.RegistryClass) {
				c.RoleAuthorizations[0].Authorizations[0].Signatures[0].Roles[0].RoleSelector =
					&types.RoleAssignment_RegistryRole{RegistryRole: types.RegistryRole_REGISTRY_ROLE_UNSPECIFIED}
			},
			expErr: "registry_role",
		},
		{
			name: "no role selector kind set",
			mutate: func(c *types.RegistryClass) {
				c.RoleAuthorizations[0].Authorizations[0].Signatures[0].Roles[0].RoleSelector = nil
			},
			expErr: "a role selector (registry_role, nft_role, or role_priority) must be set",
		},
		{
			name: "valid nft_role with current assignment",
			mutate: func(c *types.RegistryClass) {
				c.RoleAuthorizations[0].Authorizations[0].Signatures[0].Roles[0] = &types.RoleAssignment{
					RoleSelector: &types.RoleAssignment_NftRole{NftRole: types.NftRole_NFT_ROLE_NFT_OWNER},
					Assignment:   types.Assignment_ASSIGNMENT_CURRENT,
				}
			},
		},
		{
			name: "nft_role with non-current assignment",
			mutate: func(c *types.RegistryClass) {
				c.RoleAuthorizations[0].Authorizations[0].Signatures[0].Roles[0] = &types.RoleAssignment{
					RoleSelector: &types.RoleAssignment_NftRole{NftRole: types.NftRole_NFT_ROLE_NFT_OWNER},
					Assignment:   types.Assignment_ASSIGNMENT_NEW,
				}
			},
			expErr: "nft_role may only be used with a CURRENT* assignment",
		},
		{
			name: "unspecified nft_role",
			mutate: func(c *types.RegistryClass) {
				c.RoleAuthorizations[0].Authorizations[0].Signatures[0].Roles[0] = &types.RoleAssignment{
					RoleSelector: &types.RoleAssignment_NftRole{NftRole: types.NftRole_NFT_ROLE_UNSPECIFIED},
					Assignment:   types.Assignment_ASSIGNMENT_CURRENT,
				}
			},
			expErr: "nft_role",
		},
		{
			name: "valid role_priority",
			mutate: func(c *types.RegistryClass) {
				c.RoleAuthorizations[0].Authorizations[0].Signatures[0].Roles[0] = &types.RoleAssignment{
					RoleSelector: &types.RoleAssignment_RolePriority{RolePriority: &types.RolePriority{
						Entries: []*types.RolePriorityEntry{
							{Role: &types.RolePriorityEntry_RegistryRole{RegistryRole: types.RegistryRole_REGISTRY_ROLE_CONTROLLER}},
							{Role: &types.RolePriorityEntry_NftRole{NftRole: types.NftRole_NFT_ROLE_NFT_OWNER}},
						},
					}},
					Assignment: types.Assignment_ASSIGNMENT_CURRENT,
				}
			},
		},
		{
			name: "empty role_priority",
			mutate: func(c *types.RegistryClass) {
				c.RoleAuthorizations[0].Authorizations[0].Signatures[0].Roles[0] = &types.RoleAssignment{
					RoleSelector: &types.RoleAssignment_RolePriority{RolePriority: &types.RolePriority{}},
					Assignment:   types.Assignment_ASSIGNMENT_CURRENT,
				}
			},
			expErr: "role_priority must contain at least one entry",
		},
		{
			name: "nil role_priority",
			mutate: func(c *types.RegistryClass) {
				c.RoleAuthorizations[0].Authorizations[0].Signatures[0].Roles[0] = &types.RoleAssignment{
					RoleSelector: &types.RoleAssignment_RolePriority{RolePriority: nil},
					Assignment:   types.Assignment_ASSIGNMENT_CURRENT,
				}
			},
			expErr: "role_priority must contain at least one entry",
		},
		{
			name: "role_priority entry with unspecified registry_role",
			mutate: func(c *types.RegistryClass) {
				c.RoleAuthorizations[0].Authorizations[0].Signatures[0].Roles[0] = &types.RoleAssignment{
					RoleSelector: &types.RoleAssignment_RolePriority{RolePriority: &types.RolePriority{
						Entries: []*types.RolePriorityEntry{
							{Role: &types.RolePriorityEntry_RegistryRole{RegistryRole: types.RegistryRole_REGISTRY_ROLE_UNSPECIFIED}},
						},
					}},
					Assignment: types.Assignment_ASSIGNMENT_CURRENT,
				}
			},
			expErr: "role_priority entry 0: registry_role",
		},
		{
			name: "role_priority entry nft_role with non-current assignment",
			mutate: func(c *types.RegistryClass) {
				c.RoleAuthorizations[0].Authorizations[0].Signatures[0].Roles[0] = &types.RoleAssignment{
					RoleSelector: &types.RoleAssignment_RolePriority{RolePriority: &types.RolePriority{
						Entries: []*types.RolePriorityEntry{
							{Role: &types.RolePriorityEntry_NftRole{NftRole: types.NftRole_NFT_ROLE_NFT_OWNER}},
						},
					}},
					Assignment: types.Assignment_ASSIGNMENT_NEW,
				}
			},
			expErr: "role_priority entry 0: nft_role may only be used with a CURRENT* assignment",
		},
		{
			name: "role_priority entry with no role set",
			mutate: func(c *types.RegistryClass) {
				c.RoleAuthorizations[0].Authorizations[0].Signatures[0].Roles[0] = &types.RoleAssignment{
					RoleSelector: &types.RoleAssignment_RolePriority{RolePriority: &types.RolePriority{
						Entries: []*types.RolePriorityEntry{{}},
					}},
					Assignment: types.Assignment_ASSIGNMENT_CURRENT,
				}
			},
			expErr: "role_priority entry 0: a role (registry_role or nft_role) must be set",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := validRegistryClass()
			tc.mutate(c)
			err := c.Validate()
			if tc.expErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErr)
			}
		})
	}
}

func TestRegistryClass_Validate_Nil(t *testing.T) {
	var c *types.RegistryClass
	err := c.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "registry class cannot be nil")
}

func TestValidateRegistryClassID(t *testing.T) {
	tests := []struct {
		name    string
		classID string
		expErr  string
	}{
		{name: "valid alphanumeric", classID: "loan-registry-v1"},
		{name: "valid with dots", classID: "a.b.c"},
		{name: "empty", classID: "", expErr: "must be between"},
		{name: "illegal character", classID: "has space", expErr: "alphanumeric"},
		{name: "underscore not allowed", classID: "a_b", expErr: "alphanumeric"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateRegistryClassID(tc.classID)
			if tc.expErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErr)
			}
		})
	}
}

func TestValidateRegistryClassID_TooLong(t *testing.T) {
	id := make([]byte, types.MaxLenRegistryClassID+1)
	for i := range id {
		id[i] = 'a'
	}
	err := types.ValidateRegistryClassID(string(id))
	require.Error(t, err)
	require.Contains(t, err.Error(), "must be between")
}
