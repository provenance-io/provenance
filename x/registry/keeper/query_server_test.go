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
	err := s.app.RegistryKeeper.CreateRegistry(s.ctx, key, roles)
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
		err := s.app.RegistryKeeper.CreateRegistry(s.ctx, key, roles)
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
			name: "all registries - empty filter invalid",
			req: &types.QueryGetRegistriesRequest{
				Pagination: nil,
			},
			expErr:   "cannot be empty if provided",
			expCount: 0,
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
			expErr:   "cannot be empty if provided",
			expCount: 0,
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
			expErr:   "cannot be empty if provided",
			expCount: 0,
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
	err := s.app.RegistryKeeper.CreateRegistry(s.ctx, key, roles)
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
