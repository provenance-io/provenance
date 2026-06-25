package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/registry/keeper"
	"github.com/provenance-io/provenance/x/registry/types"
)

func (s *KeeperTestSuite) TestQueryGetRegistry() {
	// Setup test data
	key := &types.RegistryKey{
		AssetClassId: s.validNFTClass.Id,
		NftId:        s.validNFT.Id,
	}
	roles := []types.RolesEntry{
		{
			Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			Addresses: []string{s.user1},
		},
	}
	err := s.app.RegistryKeeper.CreateRegistry(s.ctx, key, roles, "")
	s.Require().NoError(err)

	queryServer := keeper.NewQueryServer(s.app.RegistryKeeper)

	tests := []struct {
		name    string
		req     *types.QueryGetRegistryRequest
		expErr  string
		expResp *types.QueryGetRegistryResponse
	}{
		{
			name:   "nil request",
			req:    nil,
			expErr: "empty request",
		},
		{
			name: "nil key",
			req: &types.QueryGetRegistryRequest{
				Key: nil,
			},
			expErr: "registry key cannot be nil",
		},
		{
			name: "invalid key - empty asset class id",
			req: &types.QueryGetRegistryRequest{
				Key: &types.RegistryKey{
					AssetClassId: "",
					NftId:        "nft1",
				},
			},
			expErr: "must be between",
		},
		{
			name: "invalid key - empty nft id",
			req: &types.QueryGetRegistryRequest{
				Key: &types.RegistryKey{
					AssetClassId: "class1",
					NftId:        "",
				},
			},
			expErr: "must be between",
		},
		{
			name: "registry exists",
			req: &types.QueryGetRegistryRequest{
				Key: key,
			},
			expErr: "",
			expResp: &types.QueryGetRegistryResponse{
				Registry: types.RegistryEntry{
					Key:   key,
					Roles: roles,
				},
			},
		},
		{
			name: "registry does not exist",
			req: &types.QueryGetRegistryRequest{
				Key: &types.RegistryKey{
					AssetClassId: "nonexistent",
					NftId:        "nonexistent",
				},
			},
			expErr:  "",
			expResp: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			resp, err := queryServer.GetRegistry(s.ctx, tc.req)

			if tc.expErr != "" {
				assertions.RequireErrorContents(s.T(), err, []string{tc.expErr})
				s.Require().Nil(resp)
			} else {
				assertions.RequireErrorContents(s.T(), err, nil)
				if tc.expResp == nil {
					s.Require().Nil(resp)
				} else {
					s.Require().NotNil(resp)
					s.Require().Equal(tc.expResp.Registry.Key, resp.Registry.Key)
					s.Require().Equal(tc.expResp.Registry.Roles, resp.Registry.Roles)
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryGetRegistries() {
	// Setup test data
	roles := []types.RolesEntry{
		{
			Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			Addresses: []string{s.user1},
		},
	}

	keys := []*types.RegistryKey{
		{AssetClassId: "aclass", NftId: "nft1"},
		{AssetClassId: "aclass", NftId: "nft2"},
		{AssetClassId: "bclass", NftId: "nft1"},
		{AssetClassId: "bclass", NftId: "nft2"},
	}

	for _, key := range keys {
		err := s.app.RegistryKeeper.CreateRegistry(s.ctx, key, roles, "")
		s.Require().NoError(err)
	}

	queryServer := keeper.NewQueryServer(s.app.RegistryKeeper)

	tests := []struct {
		name     string
		req      *types.QueryGetRegistriesRequest
		expErr   string
		expCount int
	}{
		{
			name:   "nil request",
			req:    nil,
			expErr: "empty request",
		},
		{
			name: "all registries without filter",
			req: &types.QueryGetRegistriesRequest{
				Pagination: nil,
			},
			expErr:   "",
			expCount: 4, // All 4 registries created
		},
		{
			name: "filter by asset class 'aclass'",
			req: &types.QueryGetRegistriesRequest{
				Pagination:   nil,
				AssetClassId: "aclass",
			},
			expErr:   "",
			expCount: 2,
		},
		{
			name: "filter by asset class 'bclass'",
			req: &types.QueryGetRegistriesRequest{
				Pagination:   nil,
				AssetClassId: "bclass",
			},
			expErr:   "",
			expCount: 2,
		},
		{
			name: "filter by nonexistent asset class",
			req: &types.QueryGetRegistriesRequest{
				Pagination:   nil,
				AssetClassId: "nonexistent",
			},
			expErr:   "",
			expCount: 0,
		},
		{
			name: "with pagination limit",
			req: &types.QueryGetRegistriesRequest{
				Pagination: &query.PageRequest{
					Limit:      2,
					CountTotal: true,
				},
			},
			expErr:   "",
			expCount: 2, // Limited to 2 registries
		},
		{
			name: "with pagination offset",
			req: &types.QueryGetRegistriesRequest{
				Pagination: &query.PageRequest{
					Offset:     2,
					Limit:      2,
					CountTotal: true,
				},
			},
			expErr:   "",
			expCount: 2, // 2 registries after skipping first 2
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			resp, err := queryServer.GetRegistries(s.ctx, tc.req)

			if tc.expErr != "" {
				assertions.RequireErrorContents(s.T(), err, []string{tc.expErr})
				s.Require().Nil(resp)
			} else {
				assertions.RequireErrorContents(s.T(), err, nil)
				s.Require().NotNil(resp)
				s.Require().Len(resp.Registries, tc.expCount)

				// Verify asset class filter if specified
				if tc.req.AssetClassId != "" {
					for _, reg := range resp.Registries {
						s.Require().Equal(tc.req.AssetClassId, reg.Key.AssetClassId)
					}
				}

				// Verify pagination info if requested
				if tc.req.Pagination != nil && tc.req.Pagination.CountTotal {
					s.Require().NotNil(resp.Pagination)
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryHasRole() {
	// Setup test data
	key := &types.RegistryKey{
		AssetClassId: s.validNFTClass.Id,
		NftId:        s.validNFT.Id,
	}
	roles := []types.RolesEntry{
		{
			Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			Addresses: []string{s.user1},
		},
		{
			Role:      types.RegistryRole_REGISTRY_ROLE_SERVICER,
			Addresses: []string{s.user1, s.user2},
		},
	}
	err := s.app.RegistryKeeper.CreateRegistry(s.ctx, key, roles, "")
	s.Require().NoError(err)

	queryServer := keeper.NewQueryServer(s.app.RegistryKeeper)

	tests := []struct {
		name       string
		req        *types.QueryHasRoleRequest
		expErr     string
		expHasRole bool
	}{
		{
			name:   "nil request",
			req:    nil,
			expErr: "empty request",
		},
		{
			name: "nil key",
			req: &types.QueryHasRoleRequest{
				Key:     nil,
				Role:    types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
				Address: s.user1,
			},
			expErr: "registry key cannot be nil",
		},
		{
			name: "invalid key - empty asset class id",
			req: &types.QueryHasRoleRequest{
				Key: &types.RegistryKey{
					AssetClassId: "",
					NftId:        "nft1",
				},
				Role:    types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
				Address: s.user1,
			},
			expErr: "must be between",
		},
		{
			name: "invalid role - unspecified",
			req: &types.QueryHasRoleRequest{
				Key:     key,
				Role:    types.RegistryRole_REGISTRY_ROLE_UNSPECIFIED,
				Address: s.user1,
			},
			expErr: "cannot be unspecified",
		},
		{
			name: "empty address",
			req: &types.QueryHasRoleRequest{
				Key:     key,
				Role:    types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
				Address: "",
			},
			expErr: "empty address string is not allowed",
		},
		{
			name: "invalid address",
			req: &types.QueryHasRoleRequest{
				Key:     key,
				Role:    types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
				Address: "invalid",
			},
			expErr: "decoding bech32",
		},
		{
			name: "user has role",
			req: &types.QueryHasRoleRequest{
				Key:     key,
				Role:    types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
				Address: s.user1,
			},
			expErr:     "",
			expHasRole: true,
		},
		{
			name: "user has different role",
			req: &types.QueryHasRoleRequest{
				Key:     key,
				Role:    types.RegistryRole_REGISTRY_ROLE_SERVICER,
				Address: s.user2,
			},
			expErr:     "",
			expHasRole: true,
		},
		{
			name: "user does not have role",
			req: &types.QueryHasRoleRequest{
				Key:     key,
				Role:    types.RegistryRole_REGISTRY_ROLE_CONTROLLER,
				Address: s.user1,
			},
			expErr:     "",
			expHasRole: false,
		},
		{
			name: "registry does not exist",
			req: &types.QueryHasRoleRequest{
				Key: &types.RegistryKey{
					AssetClassId: "nonexistent",
					NftId:        "nonexistent",
				},
				Role:    types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
				Address: s.user1,
			},
			expErr: "registry not found",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			resp, err := queryServer.HasRole(s.ctx, tc.req)

			if tc.expErr != "" {
				assertions.RequireErrorContents(s.T(), err, []string{tc.expErr})
				s.Require().Nil(resp)
			} else {
				assertions.RequireErrorContents(s.T(), err, nil)
				s.Require().NotNil(resp)
				s.Require().Equal(tc.expHasRole, resp.HasRole)
			}
		})
	}
}

// createQueryTestClass stores a minimal, valid registry class for query-server tests and returns it.
func (s *KeeperTestSuite) createQueryTestClass(classID string) types.RegistryClass {
	class := types.RegistryClass{
		RegistryClassId: classID,
		AssetClassId:    s.validNFTClass.Id,
		Maintainer:      s.user1,
		RoleAuthorizations: []types.RoleAuthorization{
			singleSigPolicy(&types.RoleAssignment{
				RoleSelector: &types.RoleAssignment_RegistryRole{RegistryRole: types.RegistryRole_REGISTRY_ROLE_CONTROLLER},
				Assignment:   types.Assignment_ASSIGNMENT_CURRENT,
			}),
		},
	}
	s.Require().NoError(s.app.RegistryKeeper.CreateRegistryClass(s.ctx, class))
	return class
}

func (s *KeeperTestSuite) TestQueryRegistryClass() {
	stored := s.createQueryTestClass("query-class-1")
	queryServer := keeper.NewQueryServer(s.app.RegistryKeeper)

	tests := []struct {
		name   string
		req    *types.QueryRegistryClassRequest
		expErr string
	}{
		{
			name:   "nil request",
			req:    nil,
			expErr: "empty request",
		},
		{
			name:   "empty registry_class_id",
			req:    &types.QueryRegistryClassRequest{RegistryClassId: ""},
			expErr: "registry_class_id cannot be empty",
		},
		{
			name:   "class does not exist",
			req:    &types.QueryRegistryClassRequest{RegistryClassId: "nonexistent"},
			expErr: "registry class",
		},
		{
			name: "class exists",
			req:  &types.QueryRegistryClassRequest{RegistryClassId: stored.RegistryClassId},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			resp, err := queryServer.RegistryClass(s.ctx, tc.req)
			if tc.expErr != "" {
				assertions.RequireErrorContents(s.T(), err, []string{tc.expErr})
				s.Require().Nil(resp)
			} else {
				assertions.RequireErrorContents(s.T(), err, nil)
				s.Require().NotNil(resp)
				s.Require().Equal(stored.RegistryClassId, resp.RegistryClass.RegistryClassId)
				s.Require().Equal(stored.Maintainer, resp.RegistryClass.Maintainer)
				s.Require().Len(resp.RegistryClass.RoleAuthorizations, 1)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryRegistryClasses() {
	s.createQueryTestClass("query-classes-a")
	s.createQueryTestClass("query-classes-b")
	s.createQueryTestClass("query-classes-c")
	queryServer := keeper.NewQueryServer(s.app.RegistryKeeper)

	tests := []struct {
		name     string
		req      *types.QueryRegistryClassesRequest
		expErr   string
		expCount int
	}{
		{
			name:   "nil request",
			req:    nil,
			expErr: "empty request",
		},
		{
			name:     "all classes",
			req:      &types.QueryRegistryClassesRequest{},
			expCount: 3,
		},
		{
			name: "pagination limit",
			req: &types.QueryRegistryClassesRequest{
				Pagination: &query.PageRequest{Limit: 2, CountTotal: true},
			},
			expCount: 2,
		},
		{
			name: "pagination offset",
			req: &types.QueryRegistryClassesRequest{
				Pagination: &query.PageRequest{Offset: 2, Limit: 2},
			},
			expCount: 1,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			resp, err := queryServer.RegistryClasses(s.ctx, tc.req)
			if tc.expErr != "" {
				assertions.RequireErrorContents(s.T(), err, []string{tc.expErr})
				s.Require().Nil(resp)
			} else {
				assertions.RequireErrorContents(s.T(), err, nil)
				s.Require().NotNil(resp)
				s.Require().Len(resp.RegistryClasses, tc.expCount)
				if tc.req.Pagination != nil && tc.req.Pagination.CountTotal {
					s.Require().NotNil(resp.Pagination)
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueryParams() {
	queryServer := keeper.NewQueryServer(s.app.RegistryKeeper)

	s.Run("nil request", func() {
		resp, err := queryServer.Params(s.ctx, nil)
		assertions.RequireErrorContents(s.T(), err, []string{"empty request"})
		s.Require().Nil(resp)
	})

	s.Run("default params", func() {
		resp, err := queryServer.Params(s.ctx, &types.QueryParamsRequest{})
		assertions.RequireErrorContents(s.T(), err, nil)
		s.Require().NotNil(resp)
		s.Require().Empty(resp.Params.RoleAuthorizations)
	})

	s.Run("reflects set params", func() {
		s.Require().NoError(s.app.RegistryKeeper.SetParams(s.ctx, types.Params{
			RoleAuthorizations: types.ControllerRoleAuthorizations(),
		}))
		resp, err := queryServer.Params(s.ctx, &types.QueryParamsRequest{})
		assertions.RequireErrorContents(s.T(), err, nil)
		s.Require().NotNil(resp)
		s.Require().Len(resp.Params.RoleAuthorizations, 1)
		s.Require().Equal(types.RegistryRole_REGISTRY_ROLE_CONTROLLER, resp.Params.RoleAuthorizations[0].Role)
	})
}

// TestKeeperRegistryClassLifecycle exercises the keeper-level class store accessors directly:
// HasRegistryClass before/after creation and GetRegistryClasses listing.
func (s *KeeperTestSuite) TestKeeperRegistryClassLifecycle() {
	has, err := s.app.RegistryKeeper.HasRegistryClass(s.ctx, "lifecycle-class")
	s.Require().NoError(err)
	s.Require().False(has)

	got, err := s.app.RegistryKeeper.GetRegistryClass(s.ctx, "lifecycle-class")
	s.Require().NoError(err)
	s.Require().Nil(got)

	s.createQueryTestClass("lifecycle-class")

	has, err = s.app.RegistryKeeper.HasRegistryClass(s.ctx, "lifecycle-class")
	s.Require().NoError(err)
	s.Require().True(has)

	classes, _, err := s.app.RegistryKeeper.GetRegistryClasses(s.ctx, nil)
	s.Require().NoError(err)
	s.Require().Len(classes, 1)
	s.Require().Equal("lifecycle-class", classes[0].RegistryClassId)
}
