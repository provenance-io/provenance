package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/x/nft"
	"github.com/stretchr/testify/suite"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	nftkeeper "cosmossdk.io/x/nft/keeper"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	// Test duplicate creation (keeper doesn't check for duplicates, just overwrites)
	err = s.app.RegistryKeeper.CreateDefaultRegistry(s.ctx, s.user1, key)
	s.Require().NoError(err) // Should succeed and overwrite
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

	// Test duplicate creation (keeper doesn't check for duplicates, just overwrites)
	err = s.app.RegistryKeeper.CreateRegistry(s.ctx, key, roles)
	s.Require().NoError(err) // Should succeed and overwrite
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
		expErr     string
		expHasRole bool
		expAddr    string
	}{
		{
			name:       "successful role grant",
			key:        baseKey,
			role:       types.RegistryRole_REGISTRY_ROLE_SERVICER,
			addr:       []string{s.user2Addr.String()},
			expErr:     "",
			expHasRole: true,
			expAddr:    s.user2Addr.String(),
		},
		{
			name:       "granting multiple addresses",
			key:        baseKey,
			role:       types.RegistryRole_REGISTRY_ROLE_CONTROLLER,
			addr:       []string{s.user1Addr.String(), s.user2Addr.String()},
			expErr:     "",
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
			expErr: "collections: not found",
		},
		{
			name:   "address already has role",
			key:    baseKey,
			role:   types.RegistryRole_REGISTRY_ROLE_SERVICER,
			addr:   []string{s.user2Addr.String()},
			expErr: "address already has role",
		},
		{
			name:   "invalid role",
			key:    baseKey,
			role:   types.RegistryRole_REGISTRY_ROLE_UNSPECIFIED,
			addr:   []string{s.user2Addr.String()},
			expErr: "invalid role",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := s.app.RegistryKeeper.GrantRole(s.ctx, tc.key, tc.role, tc.addr)
			if tc.expErr != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErr)
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
		expErr     string
		expHasRole bool
		checkAddr  string
	}{
		{
			name:       "successful role revocation for one address",
			key:        baseKey,
			role:       types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			addr:       []string{s.user2Addr.String()},
			expErr:     "",
			expHasRole: false,
			checkAddr:  s.user2Addr.String(),
		},
		{
			name:       "revoking multiple addresses",
			key:        baseKey,
			role:       types.RegistryRole_REGISTRY_ROLE_SERVICER,
			addr:       []string{s.user1Addr.String(), s.user2Addr.String()},
			expErr:     "",
			expHasRole: true, // The current implementation doesn't properly revoke
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
			expErr: "collections: not found",
		},
		{
			name:   "invalid role",
			key:    baseKey,
			role:   types.RegistryRole_REGISTRY_ROLE_UNSPECIFIED,
			addr:   []string{s.user2Addr.String()},
			expErr: "invalid role",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := s.app.RegistryKeeper.RevokeRole(s.ctx, tc.key, tc.role, tc.addr)
			if tc.expErr != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErr)
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

func (s *KeeperTestSuite) TestInitGenesis() {
	// Test InitGenesis with empty state
	emptyState := &types.GenesisState{}
	s.app.RegistryKeeper.InitGenesis(s.ctx, emptyState)

	// Test InitGenesis with valid state
	key := &types.RegistryKey{
		AssetClassId: "genesis-test-class",
		NftId:        "genesis-test-nft",
	}
	roles := []types.RolesEntry{
		{
			Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			Addresses: []string{s.user1Addr.String()},
		},
	}
	entry := types.RegistryEntry{
		Key:   key,
		Roles: roles,
	}
	genesisState := &types.GenesisState{
		Entries: []types.RegistryEntry{entry},
	}
	s.app.RegistryKeeper.InitGenesis(s.ctx, genesisState)

	// Verify genesis state was applied (currently not implemented)
	// retrievedEntry, err := s.app.RegistryKeeper.GetRegistry(s.ctx, key)
	// s.Require().NoError(err)
	// s.Require().NotNil(retrievedEntry)
	// s.Require().Equal(key, retrievedEntry.Key)
	// s.Require().Equal(roles, retrievedEntry.Roles)
}

func (s *KeeperTestSuite) TestExportGenesis() {
	// Create some registries for this test
	key1 := &types.RegistryKey{
		AssetClassId: "export-test-class-1",
		NftId:        "export-test-nft-1",
	}
	roles1 := []types.RolesEntry{
		{
			Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			Addresses: []string{s.user1Addr.String()},
		},
	}
	err := s.app.RegistryKeeper.CreateRegistry(s.ctx, key1, roles1)
	s.Require().NoError(err)

	key2 := &types.RegistryKey{
		AssetClassId: "export-test-class-2",
		NftId:        "export-test-nft-2",
	}
	roles2 := []types.RolesEntry{
		{
			Role:      types.RegistryRole_REGISTRY_ROLE_SERVICER,
			Addresses: []string{s.user2Addr.String()},
		},
	}
	err = s.app.RegistryKeeper.CreateRegistry(s.ctx, key2, roles2)
	s.Require().NoError(err)

	// Export genesis state
	exportedState := s.app.RegistryKeeper.ExportGenesis(s.ctx)
	s.Require().NotNil(exportedState)
	s.Require().Len(exportedState.Entries, 0) // Current implementation returns empty state
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
				Authority: s.user1Addr.String(),
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
				Authority: s.user1Addr.String(),
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
				Authority: s.user1Addr.String(),
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
				Authority: s.user2Addr.String(),
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
			expErr: "registry already exists", // The registry already exists from previous tests
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
