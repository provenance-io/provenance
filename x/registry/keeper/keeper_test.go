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
	"github.com/provenance-io/provenance/x/registry"
	"github.com/provenance-io/provenance/x/registry/keeper"
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

func (s *KeeperTestSuite) TestCreateRegistry() {
	key := &registry.RegistryKey{
		AssetClassId: s.validNFTClass.Id,
		NftId:        s.validNFT.Id,
	}
	roles := []registry.RolesEntry{
		{
			Role:      registry.RegistryRole_REGISTRY_ROLE_ORIGINATOR.String(),
			Addresses: []string{s.user1Addr.String()},
		},
		{
			Role:      registry.RegistryRole_REGISTRY_ROLE_SERVICER.String(),
			Addresses: []string{s.user1Addr.String()},
		},
	}

	// Test successful creation
	err := s.app.RegistryKeeper.CreateRegistry(s.ctx, s.user1Addr, key, roles)
	s.Require().NoError(err)

	// Test duplicate creation
	err = s.app.RegistryKeeper.CreateRegistry(s.ctx, s.user1Addr, key, roles)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "registry already exists")
}

func (s *KeeperTestSuite) TestGrantRole() {
	key := &registry.RegistryKey{
		AssetClassId: s.validNFTClass.Id,
		NftId:        s.validNFT.Id,
	}
	roles := []registry.RolesEntry{
		{
			Role:      registry.RegistryRole_REGISTRY_ROLE_ORIGINATOR.String(),
			Addresses: []string{s.user1Addr.String()},
		},
	}

	// Create registry first
	err := s.app.RegistryKeeper.CreateRegistry(s.ctx, s.user1Addr, key, roles)
	s.Require().NoError(err)

	// Test successful role grant
	err = s.app.RegistryKeeper.GrantRole(s.ctx, s.user1Addr, key, "admin", []*sdk.AccAddress{&s.user2Addr})
	s.Require().NoError(err)

	// Verify role was granted
	hasRole, err := s.app.RegistryKeeper.HasRole(s.ctx, key, "admin", s.user2Addr.String())
	s.Require().NoError(err)
	s.Require().True(hasRole)

	// Test granting to non-existent registry
	nonExistentKey := &registry.RegistryKey{
		AssetClassId: "nonexistent",
		NftId:        "nonexistent",
	}
	err = s.app.RegistryKeeper.GrantRole(s.ctx, s.user1Addr, nonExistentKey, "admin", []*sdk.AccAddress{&s.user2Addr})
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "registry not found")

	// Test granting to address that already has role
	err = s.app.RegistryKeeper.GrantRole(s.ctx, s.user1Addr, key, "admin", []*sdk.AccAddress{&s.user2Addr})
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "address already has role")
}

func (s *KeeperTestSuite) TestRevokeRole() {
	key := &registry.RegistryKey{
		AssetClassId: s.validNFTClass.Id,
		NftId:        s.validNFT.Id,
	}
	roles := []registry.RolesEntry{
		{
			Role:      registry.RegistryRole_REGISTRY_ROLE_ORIGINATOR.String(),
			Addresses: []string{s.user1Addr.String(), s.user2Addr.String()},
		},
	}

	// Create registry first
	err := s.app.RegistryKeeper.CreateRegistry(s.ctx, s.user1Addr, key, roles)
	s.Require().NoError(err)

	// Test successful role revocation
	err = s.app.RegistryKeeper.RevokeRole(s.ctx, s.user1Addr, key, "admin", []*sdk.AccAddress{&s.user2Addr})
	s.Require().NoError(err)

	// Verify role was revoked
	hasRole, err := s.app.RegistryKeeper.HasRole(s.ctx, key, "admin", s.user2Addr.String())
	s.Require().NoError(err)
	s.Require().False(hasRole)

	// Test revoking from non-existent registry
	nonExistentKey := &registry.RegistryKey{
		AssetClassId: "nonexistent",
		NftId:        "nonexistent",
	}
	err = s.app.RegistryKeeper.RevokeRole(s.ctx, s.user1Addr, nonExistentKey, "admin", []*sdk.AccAddress{&s.user2Addr})
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "registry not found")
}

func (s *KeeperTestSuite) TestHasRole() {
	key := &registry.RegistryKey{
		AssetClassId: s.validNFTClass.Id,
		NftId:        s.validNFT.Id,
	}
	roles := []registry.RolesEntry{
		{
			Role:      registry.RegistryRole_REGISTRY_ROLE_ORIGINATOR.String(),
			Addresses: []string{s.user1Addr.String()},
		},
	}

	// Create registry first
	err := s.app.RegistryKeeper.CreateRegistry(s.ctx, s.user1Addr, key, roles)
	s.Require().NoError(err)

	// Test has role
	hasRole, err := s.app.RegistryKeeper.HasRole(s.ctx, key, "REGISTRY_ROLE_ORIGINATOR", s.user1Addr.String())
	s.Require().NoError(err)
	s.Require().True(hasRole)

	// Test doesn't have role
	hasRole, err = s.app.RegistryKeeper.HasRole(s.ctx, key, "REGISTRY_ROLE_ORIGINATOR", s.user2Addr.String())
	s.Require().NoError(err)
	s.Require().False(hasRole)

	// Test non-existent registry
	nonExistentKey := &registry.RegistryKey{
		AssetClassId: "nonexistent",
		NftId:        "nonexistent",
	}
	hasRole, err = s.app.RegistryKeeper.HasRole(s.ctx, nonExistentKey, "admin", s.user1Addr.String())
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "registry not found")
}

func (s *KeeperTestSuite) TestGetRegistry() {
	key := &registry.RegistryKey{
		AssetClassId: s.validNFTClass.Id,
		NftId:        s.validNFT.Id,
	}
	roles := []registry.RolesEntry{
		{
			Role:      registry.RegistryRole_REGISTRY_ROLE_ORIGINATOR.String(),
			Addresses: []string{s.user1Addr.String()},
		},
	}

	// Create registry first
	err := s.app.RegistryKeeper.CreateRegistry(s.ctx, s.user1Addr, key, roles)
	s.Require().NoError(err)

	// Test get existing registry
	entry, err := s.app.RegistryKeeper.GetRegistry(s.ctx, key)
	s.Require().NoError(err)
	s.Require().Equal(key, entry.Key)
	s.Require().Equal(roles, entry.Roles)

	// Test get non-existent registry
	nonExistentKey := &registry.RegistryKey{
		AssetClassId: "nonexistent",
		NftId:        "nonexistent",
	}
	entry, err = s.app.RegistryKeeper.GetRegistry(s.ctx, nonExistentKey)
	s.Require().NoError(err)
	s.Require().Nil(entry)
}

func (s *KeeperTestSuite) TestRegisterNFTMsgServer() {
	key := &registry.RegistryKey{
		AssetClassId: s.validNFTClass.Id,
		NftId:        s.validNFT.Id,
	}
	roles := []registry.RolesEntry{
		{
			Role:      registry.RegistryRole_REGISTRY_ROLE_ORIGINATOR.String(),
			Addresses: []string{s.user1Addr.String()},
		},
	}

	// Create msg
	msg := &registry.MsgRegisterNFT{
		Authority: s.user1Addr.String(),
		Key:       key,
		Roles:     roles,
	}

	// Create msg server
	msgServer := keeper.NewMsgServer(s.app.RegistryKeeper)

	// Test successful registration
	_, err := msgServer.RegisterNFT(s.ctx, msg)
	s.Require().NoError(err)

	// Verify registry was created
	entry, err := s.app.RegistryKeeper.GetRegistry(s.ctx, key)
	s.Require().NoError(err)
	s.Require().Equal(key, entry.Key)
	s.Require().Equal(roles, entry.Roles)

	// Test duplicate registration
	_, err = msgServer.RegisterNFT(s.ctx, msg)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "registry already exists")
}
