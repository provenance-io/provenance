package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/x/nft"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	nftkeeper "cosmossdk.io/x/nft/keeper"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/registry/keeper"
	"github.com/provenance-io/provenance/x/registry/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress

	nftKeeper nftkeeper.Keeper

	validNFTClass nft.Class
	validNFT      nft.NFT
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.nftKeeper = s.app.NFTKeeper

	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: time.Now()})

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	s.ctx = s.ctx.WithBlockTime(time.Now())

	s.validNFTClass = nft.Class{
		Id: "test-nft-class-id",
	}
	s.nftKeeper.SaveClass(s.ctx, s.validNFTClass)

	s.validNFT = nft.NFT{
		ClassId: s.validNFTClass.Id,
		Id:      "test-nft-id",
	}
	s.nftKeeper.Mint(s.ctx, s.validNFT, s.user1Addr)
}

func (s *KeeperTestSuite) TestCreateDefaultRegistry() {
	// Test successful creation
	key := &types.RegistryKey{
		AssetClassId: "default-registry-class",
		NftId:        "default-registry-nft",
	}

	err := s.app.RegistryKeeper.CreateDefaultRegistry(s.ctx, s.user1, key)
	s.Require().NoError(err)

	// Verify registry was created with default roles
	entry, err := s.app.RegistryKeeper.GetRegistry(s.ctx, key)
	s.Require().NoError(err)
	s.Require().NotNil(entry)
	s.Require().Equal(key, entry.Key)
	s.Require().Len(entry.Roles, 1)
	s.Require().Equal(types.RegistryRole_REGISTRY_ROLE_ORIGINATOR, entry.Roles[0].Role)
	s.Require().Equal([]string{s.user1}, entry.Roles[0].Addresses)

	// Test duplicate creation not allowed.
	expDupErr := "registry already exists for key: \"" + key.String() + "\": registry already exists"
	err = s.app.RegistryKeeper.CreateDefaultRegistry(s.ctx, s.user1, key)
	s.Require().EqualError(err, expDupErr)
}

func (s *KeeperTestSuite) TestCreateRegistry() {
	// Test successful creation with different key
	key := &types.RegistryKey{
		AssetClassId: "create-registry-class",
		NftId:        "create-registry-nft",
	}
	roles := []types.RolesEntry{
		{
			Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			Addresses: []string{s.user1Addr.String()},
		},
		{
			Role:      types.RegistryRole_REGISTRY_ROLE_SERVICER,
			Addresses: []string{s.user1Addr.String()},
		},
	}

	err := s.app.RegistryKeeper.CreateRegistry(s.ctx, key, roles)
	s.Require().NoError(err)

	// Verify registry was created
	entry, err := s.app.RegistryKeeper.GetRegistry(s.ctx, key)
	s.Require().NoError(err)
	s.Require().NotNil(entry)
	s.Require().Equal(key, entry.Key)
	s.Require().Equal(roles, entry.Roles)

	// Test duplicate creation not allowed.
	expDupErr := "registry already exists for key: \"" + key.String() + "\": registry already exists"
	err = s.app.RegistryKeeper.CreateRegistry(s.ctx, key, roles)
	s.Require().EqualError(err, expDupErr)
}

func (s *KeeperTestSuite) TestGrantRole() {
	// Create a base registry for testing
	baseKey := &types.RegistryKey{
		AssetClassId: s.validNFTClass.Id,
		NftId:        s.validNFT.Id,
	}
	baseRoles := []types.RolesEntry{
		{
			Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			Addresses: []string{s.user1Addr.String()},
		},
	}
	err := s.app.RegistryKeeper.CreateRegistry(s.ctx, baseKey, baseRoles)
	s.Require().NoError(err)

	tests := []struct {
		name       string
		key        *types.RegistryKey
		role       types.RegistryRole
		addr       []string
		expErr     error
		expHasRole bool
		expAddr    string
	}{
		{
			name:       "successful role grant",
			key:        baseKey,
			role:       types.RegistryRole_REGISTRY_ROLE_SERVICER,
			addr:       []string{s.user2Addr.String()},
			expErr:     nil,
			expHasRole: true,
			expAddr:    s.user2Addr.String(),
		},
		{
			name:       "granting multiple addresses",
			key:        baseKey,
			role:       types.RegistryRole_REGISTRY_ROLE_CONTROLLER,
			addr:       []string{s.user1Addr.String(), s.user2Addr.String()},
			expErr:     nil,
			expHasRole: true,
			expAddr:    s.user1Addr.String(),
		},
		{
			name: "non-existent registry",
			key: &types.RegistryKey{
				AssetClassId: "nonexistent",
				NftId:        "nonexistent",
			},
			role:   types.RegistryRole_REGISTRY_ROLE_SERVICER,
			addr:   []string{s.user2Addr.String()},
			expErr: types.ErrRegistryNotFound,
		},
		{
			name:   "address already has role",
			key:    baseKey,
			role:   types.RegistryRole_REGISTRY_ROLE_SERVICER,
			addr:   []string{s.user2Addr.String()},
			expErr: types.ErrAddressAlreadyHasRole,
		},
		{
			name:   "invalid role",
			key:    baseKey,
			role:   types.RegistryRole_REGISTRY_ROLE_UNSPECIFIED,
			addr:   []string{s.user2Addr.String()},
			expErr: types.ErrInvalidRole,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := s.app.RegistryKeeper.GrantRole(s.ctx, tc.key, tc.role, tc.addr)
			if tc.expErr != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expErr)
			} else {
				s.Require().NoError(err)

				// Verify role was granted
				hasRole, err := s.app.RegistryKeeper.HasRole(s.ctx, tc.key, tc.role, tc.expAddr)
				s.Require().NoError(err)
				s.Require().Equal(tc.expHasRole, hasRole)
			}
		})
	}
}

func (s *KeeperTestSuite) TestRevokeRole() {
	// Create a base registry for testing
	baseKey := &types.RegistryKey{
		AssetClassId: s.validNFTClass.Id,
		NftId:        s.validNFT.Id,
	}
	baseRoles := []types.RolesEntry{
		{
			Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			Addresses: []string{s.user1Addr.String(), s.user2Addr.String()},
		},
		{
			Role:      types.RegistryRole_REGISTRY_ROLE_SERVICER,
			Addresses: []string{s.user1Addr.String(), s.user2Addr.String()},
		},
	}
	err := s.app.RegistryKeeper.CreateRegistry(s.ctx, baseKey, baseRoles)
	s.Require().NoError(err)

	tests := []struct {
		name       string
		key        *types.RegistryKey
		role       types.RegistryRole
		addr       []string
		expErr     error
		expHasRole bool
		checkAddr  string
	}{
		{
			name:       "successful role revocation for one address",
			key:        baseKey,
			role:       types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			addr:       []string{s.user2Addr.String()},
			expErr:     nil,
			expHasRole: false,
			checkAddr:  s.user2Addr.String(),
		},
		{
			name:       "revoking multiple addresses",
			key:        baseKey,
			role:       types.RegistryRole_REGISTRY_ROLE_SERVICER,
			addr:       []string{s.user1Addr.String(), s.user2Addr.String()},
			expErr:     nil,
			expHasRole: false,
			checkAddr:  s.user1Addr.String(),
		},
		{
			name: "non-existent registry",
			key: &types.RegistryKey{
				AssetClassId: "nonexistent",
				NftId:        "nonexistent",
			},
			role:   types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			addr:   []string{s.user2Addr.String()},
			expErr: types.ErrRegistryNotFound,
		},
		{
			name:   "invalid role",
			key:    baseKey,
			role:   types.RegistryRole_REGISTRY_ROLE_UNSPECIFIED,
			addr:   []string{s.user2Addr.String()},
			expErr: types.ErrInvalidRole,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := s.app.RegistryKeeper.RevokeRole(s.ctx, tc.key, tc.role, tc.addr)
			if tc.expErr != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expErr)
			} else {
				s.Require().NoError(err)

				// Verify role was revoked
				hasRole, err := s.app.RegistryKeeper.HasRole(s.ctx, tc.key, tc.role, tc.checkAddr)
				s.Require().NoError(err)
				s.Require().Equal(tc.expHasRole, hasRole)
			}
		})
	}
}

func (s *KeeperTestSuite) TestHasRole() {
	// Create a base registry for testing
	baseKey := &types.RegistryKey{
		AssetClassId: s.validNFTClass.Id,
		NftId:        s.validNFT.Id,
	}
	baseRoles := []types.RolesEntry{
		{
			Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			Addresses: []string{s.user1Addr.String()},
		},
	}
	err := s.app.RegistryKeeper.CreateRegistry(s.ctx, baseKey, baseRoles)
	s.Require().NoError(err)

	tests := []struct {
		name       string
		key        *types.RegistryKey
		role       types.RegistryRole
		address    string
		expHasRole bool
		expErr     string
	}{
		{
			name:       "has role",
			key:        baseKey,
			role:       types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			address:    s.user1Addr.String(),
			expHasRole: true,
			expErr:     "",
		},
		{
			name:       "doesn't have role",
			key:        baseKey,
			role:       types.RegistryRole_REGISTRY_ROLE_SERVICER,
			address:    s.user2Addr.String(),
			expHasRole: false,
			expErr:     "",
		},
		{
			name: "non-existent registry",
			key: &types.RegistryKey{
				AssetClassId: "nonexistent",
				NftId:        "nonexistent",
			},
			role:       types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			address:    s.user1Addr.String(),
			expHasRole: false,
			expErr:     "collections: not found",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			hasRole, err := s.app.RegistryKeeper.HasRole(s.ctx, tc.key, tc.role, tc.address)
			if tc.expErr != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErr)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expHasRole, hasRole)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetRegistry() {
	// Create a base registry for testing
	baseKey := &types.RegistryKey{
		AssetClassId: s.validNFTClass.Id,
		NftId:        s.validNFT.Id,
	}
	baseRoles := []types.RolesEntry{
		{
			Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			Addresses: []string{s.user1Addr.String()},
		},
	}
	err := s.app.RegistryKeeper.CreateRegistry(s.ctx, baseKey, baseRoles)
	s.Require().NoError(err)

	tests := []struct {
		name     string
		key      *types.RegistryKey
		expErr   string
		expKey   *types.RegistryKey
		expRoles []types.RolesEntry
	}{
		{
			name:     "get existing registry",
			key:      baseKey,
			expErr:   "",
			expKey:   baseKey,
			expRoles: baseRoles,
		},
		{
			name: "get non-existent registry",
			key: &types.RegistryKey{
				AssetClassId: "nonexistent",
				NftId:        "nonexistent",
			},
			expErr:   "",
			expKey:   nil,
			expRoles: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			entry, err := s.app.RegistryKeeper.GetRegistry(s.ctx, tc.key)
			if tc.expErr != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErr)
			} else {
				s.Require().NoError(err)
				if tc.expKey != nil {
					s.Require().NotNil(entry)
					s.Require().Equal(tc.expKey, entry.Key)
					s.Require().Equal(tc.expRoles, entry.Roles)
				} else {
					s.Require().Nil(entry)
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetRegistries() {
	// Initially no registries
	regs, _, err := s.app.RegistryKeeper.GetRegistries(s.ctx, nil, "")
	s.Require().NoError(err)
	s.Require().Len(regs, 0)

	// Create test data
	roles := []types.RolesEntry{
		{Role: types.RegistryRole_REGISTRY_ROLE_ORIGINATOR, Addresses: []string{s.user1Addr.String()}},
	}
	keys := []*types.RegistryKey{
		{AssetClassId: "aclass", NftId: "nft1"},
		{AssetClassId: "aclass", NftId: "nft2"},
		{AssetClassId: "bclass", NftId: "nft1"},
		{AssetClassId: "bclass", NftId: "nft2"},
		{AssetClassId: "cclass", NftId: "nft1"},
		{AssetClassId: "dclass", NftId: "nft1"},
		{AssetClassId: "eclass", NftId: "nft1"},
	}
	for _, k := range keys {
		s.Require().NoError(s.app.RegistryKeeper.CreateRegistry(s.ctx, k, roles))
	}

	// Helper to get key strings in order
	keyStrs := func(es []types.RegistryEntry) []string {
		out := make([]string, len(es))
		for i, e := range es {
			out[i] = e.Key.String()
		}
		return out
	}

	// Test comprehensive pagination scenarios
	tests := []struct {
		name          string
		pageRequest   *query.PageRequest
		assetClassId  string
		expectedCount int
		expectError   bool
	}{
		{
			name:          "get all registries without pagination",
			pageRequest:   nil,
			assetClassId:  "",
			expectedCount: len(keys),
			expectError:   false,
		},
		{
			name:          "get all registries with count total",
			pageRequest:   &query.PageRequest{CountTotal: true},
			assetClassId:  "",
			expectedCount: len(keys),
			expectError:   false,
		},
		{
			name:          "get registries with limit 3",
			pageRequest:   &query.PageRequest{Limit: 3, CountTotal: true},
			assetClassId:  "",
			expectedCount: 3,
			expectError:   false,
		},
		{
			name:          "get registries with offset 2 and limit 3",
			pageRequest:   &query.PageRequest{Offset: 2, Limit: 3, CountTotal: true},
			assetClassId:  "",
			expectedCount: 3,
			expectError:   false,
		},
		{
			name:          "get registries with offset 5 and limit 5",
			pageRequest:   &query.PageRequest{Offset: 5, Limit: 5, CountTotal: true},
			assetClassId:  "",
			expectedCount: 2, // Only 2 remaining after offset 5
			expectError:   false,
		},
		{
			name:          "get registries with offset beyond total count",
			pageRequest:   &query.PageRequest{Offset: 10, Limit: 5, CountTotal: true},
			assetClassId:  "",
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "get registries with large limit",
			pageRequest:   &query.PageRequest{Limit: 100, CountTotal: true},
			assetClassId:  "",
			expectedCount: len(keys),
			expectError:   false,
		},
		{
			name:          "get registries with reverse order",
			pageRequest:   &query.PageRequest{Limit: uint64(len(keys)), Reverse: true, CountTotal: true},
			assetClassId:  "",
			expectedCount: len(keys),
			expectError:   false,
		},
		{
			name:          "filter by asset class 'aclass'",
			pageRequest:   &query.PageRequest{CountTotal: true},
			assetClassId:  "aclass",
			expectedCount: 2, // aclass has 2 NFTs
			expectError:   false,
		},
		{
			name:          "filter by asset class 'bclass' with limit 1",
			pageRequest:   &query.PageRequest{Limit: 1, CountTotal: true},
			assetClassId:  "bclass",
			expectedCount: 1,
			expectError:   false,
		},
		{
			name:          "filter by non-existent asset class",
			pageRequest:   &query.PageRequest{CountTotal: true},
			assetClassId:  "nonexistent",
			expectedCount: 0,
			expectError:   false,
		},
	}

	// Get baseline data for comparison
	allRegs, _, err := s.app.RegistryKeeper.GetRegistries(s.ctx, nil, "")
	s.Require().NoError(err)
	s.Require().Len(allRegs, len(keys))

	for _, tc := range tests {
		s.Run(tc.name, func() {
			regs, pageRes, err := s.app.RegistryKeeper.GetRegistries(s.ctx, tc.pageRequest, tc.assetClassId)

			if tc.expectError {
				s.Require().Error(err)
				return
			}

			s.Require().NoError(err)
			s.Require().NotNil(regs)
			s.Require().Len(regs, tc.expectedCount)

			// Verify pagination response if CountTotal was requested
			if tc.pageRequest != nil && tc.pageRequest.CountTotal {
				s.Require().NotNil(pageRes)
				if tc.assetClassId == "" && tc.expectedCount > 0 {
					// For queries without asset class filter, verify total count
					s.Require().Equal(uint64(len(keys)), pageRes.Total)
				}
			}

			// Verify all returned registries are valid
			for _, reg := range regs {
				s.Require().NotNil(reg.Key)
				s.Require().NotEmpty(reg.Key.AssetClassId)
				s.Require().NotEmpty(reg.Key.NftId)
				s.Require().NotEmpty(reg.Roles)

				// If filtering by asset class, verify all results match
				if tc.assetClassId != "" {
					s.Require().Equal(tc.assetClassId, reg.Key.AssetClassId)
				}
			}

			// Verify ordering for reverse tests
			if tc.pageRequest != nil && tc.pageRequest.Reverse && tc.assetClassId == "" {
				// For reverse order test, verify ordering is actually reversed
				if len(regs) > 1 {
					// Compare first and last elements to ensure reverse ordering
					allKeys := keyStrs(allRegs)
					resultKeys := keyStrs(regs)
					s.Require().Equal(allKeys[len(allKeys)-1], resultKeys[0], "first element should be last in original order")
				}
			}

			// Verify deterministic ordering by running the same query twice
			regs2, _, err2 := s.app.RegistryKeeper.GetRegistries(s.ctx, tc.pageRequest, tc.assetClassId)
			s.Require().NoError(err2)
			s.Require().Equal(keyStrs(regs), keyStrs(regs2), "results should be deterministic")
		})
	}

	// Test pagination consistency - verify that paginated results match complete results
	s.Run("pagination consistency", func() {
		// Get all results without pagination
		allRegs, _, err := s.app.RegistryKeeper.GetRegistries(s.ctx, &query.PageRequest{CountTotal: true}, "")
		s.Require().NoError(err)

		// Get results in pages of 2
		var paginatedRegs []types.RegistryEntry
		pageSize := uint64(2)
		offset := uint64(0)

		for {
			pageRegs, _, err := s.app.RegistryKeeper.GetRegistries(s.ctx, &query.PageRequest{
				Offset:     offset,
				Limit:      pageSize,
				CountTotal: true,
			}, "")
			s.Require().NoError(err)

			if len(pageRegs) == 0 {
				break
			}

			paginatedRegs = append(paginatedRegs, pageRegs...)
			offset += pageSize
		}

		// Verify paginated results match complete results
		s.Require().Equal(len(allRegs), len(paginatedRegs), "paginated results should have same count as complete results")
		s.Require().Equal(keyStrs(allRegs), keyStrs(paginatedRegs), "paginated results should match complete results in order")
	})
}

func (s *KeeperTestSuite) GenesisTest() {
	k := s.app.RegistryKeeper

	genesis1 := k.ExportGenesis(s.ctx)

	// Clear the state
	k.Registry.Clear(s.ctx, nil)

	// Import the genesis state
	k.InitGenesis(s.ctx, genesis1)

	// Export the genesis state
	genesis2 := k.ExportGenesis(s.ctx)

	// Compare the two genesis states
	s.Require().Equal(genesis1, genesis2)
}

func (s *KeeperTestSuite) TestRegisterNFTMsgServer() {
	tests := []struct {
		name     string
		msg      *types.MsgRegisterNFT
		expErr   string
		expKey   *types.RegistryKey
		expRoles []types.RolesEntry
	}{
		{
			name: "successful registration",
			msg: &types.MsgRegisterNFT{
				Signer: s.user1Addr.String(),
				Key: &types.RegistryKey{
					AssetClassId: s.validNFTClass.Id,
					NftId:        s.validNFT.Id,
				},
				Roles: []types.RolesEntry{
					{
						Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
						Addresses: []string{s.user1Addr.String()},
					},
				},
			},
			expErr: "",
			expKey: &types.RegistryKey{
				AssetClassId: s.validNFTClass.Id,
				NftId:        s.validNFT.Id,
			},
			expRoles: []types.RolesEntry{
				{
					Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
					Addresses: []string{s.user1Addr.String()},
				},
			},
		},
		{
			name: "duplicate registration",
			msg: &types.MsgRegisterNFT{
				Signer: s.user1Addr.String(),
				Key: &types.RegistryKey{
					AssetClassId: s.validNFTClass.Id,
					NftId:        s.validNFT.Id,
				},
				Roles: []types.RolesEntry{
					{
						Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
						Addresses: []string{s.user1Addr.String()},
					},
				},
			},
			expErr: "registry already exists",
		},
		{
			name: "non-existent NFT",
			msg: &types.MsgRegisterNFT{
				Signer: s.user1Addr.String(),
				Key: &types.RegistryKey{
					AssetClassId: "nonexistent",
					NftId:        "nonexistent",
				},
				Roles: []types.RolesEntry{
					{
						Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
						Addresses: []string{s.user1Addr.String()},
					},
				},
			},
			expErr: "NFT does not exist",
		},
		{
			name: "wrong authority",
			msg: &types.MsgRegisterNFT{
				Signer: s.user2Addr.String(),
				Key: &types.RegistryKey{
					AssetClassId: s.validNFTClass.Id,
					NftId:        s.validNFT.Id,
				},
				Roles: []types.RolesEntry{
					{
						Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
						Addresses: []string{s.user1Addr.String()},
					},
				},
			},
			expErr: "unauthorized access: signer does not own the NFT: unauthorized",
		},
	}

	msgServer := keeper.NewMsgServer(s.app.RegistryKeeper)

	for _, tc := range tests {
		s.Run(tc.name, func() {
			_, err := msgServer.RegisterNFT(s.ctx, tc.msg)
			if tc.expErr != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErr)
			} else {
				s.Require().NoError(err)

				// Verify registry was created
				entry, err := s.app.RegistryKeeper.GetRegistry(s.ctx, tc.expKey)
				s.Require().NoError(err)
				s.Require().NotNil(entry)
				s.Require().Equal(tc.expKey, entry.Key)
				s.Require().Equal(tc.expRoles, entry.Roles)
			}
		})
	}
}
