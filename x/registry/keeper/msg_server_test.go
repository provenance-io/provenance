package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/x/nft"
	nftkeeper "cosmossdk.io/x/nft/keeper"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/registry/keeper"
	"github.com/provenance-io/provenance/x/registry/types"
)

type MsgServerTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	registryKeeper keeper.Keeper
	nftKeeper      nftkeeper.Keeper

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	validNFTClass nft.Class
	validNFT      nft.NFT
	validNFT2     nft.NFT

	integrationApp *integration.App
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

func (s *MsgServerTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: time.Now()})

	s.nftKeeper = s.app.NFTKeeper
	s.registryKeeper = s.app.RegistryKeeper

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	s.ctx = s.ctx.WithBlockTime(time.Now())

	s.validNFTClass = nft.Class{
		Id: "test-nft-class-id",
	}
	s.nftKeeper.SaveClass(s.ctx, s.validNFTClass)

	s.validNFT = nft.NFT{
		ClassId: s.validNFTClass.Id,
		Id:      "test-nft-id",
	}

	s.validNFT2 = nft.NFT{
		ClassId: s.validNFTClass.Id,
		Id:      "test-nft-id-2",
	}

	s.nftKeeper.Mint(s.ctx, s.validNFT, s.user1Addr)
	s.nftKeeper.Mint(s.ctx, s.validNFT2, s.user1Addr)
}

func (s *MsgServerTestSuite) TestRegisterNFTGRPC() {
	servicerRole := types.RegistryRole_REGISTRY_ROLE_SERVICER

	msg := &types.MsgRegisterNFT{
		Authority: s.user1Addr.String(),
		Key: &types.RegistryKey{
			AssetClassId: s.validNFTClass.Id,
			NftId:        s.validNFT.Id,
		},
		Roles: []types.RolesEntry{
			{
				Role:      servicerRole,
				Addresses: []string{s.user1Addr.String()},
			},
		},
	}

	handler := s.app.MsgServiceRouter().Handler(msg)
	_, err := handler(s.ctx, msg)
	s.Require().NoError(err)

	registry, err := s.registryKeeper.GetRegistry(s.ctx, &types.RegistryKey{
		AssetClassId: s.validNFTClass.Id,
		NftId:        s.validNFT.Id,
	})
	s.Require().NoError(err)
	s.Require().NotNil(registry)
	s.Require().Equal(s.validNFTClass.Id, registry.Key.AssetClassId)
	s.Require().Equal(s.validNFT.Id, registry.Key.NftId)

	servicerRoleAddress := registry.Roles[0].Addresses[0]
	s.Require().Equal(s.user1Addr.String(), servicerRoleAddress)
}


func (s *MsgServerTestSuite) TestRegistryBulkUpdateNFTGRPC() {
	servicerRole := types.RegistryRole_REGISTRY_ROLE_SERVICER

	msg := &types.MsgRegistryBulkUpdate{
		Authority: s.user1Addr.String(),
		Entries: []types.RegistryEntry{
			{
				Key: &types.RegistryKey{
					AssetClassId: s.validNFTClass.Id,
					NftId:        s.validNFT.Id,
				},
				Roles: []types.RolesEntry{
					{
						Role:      servicerRole,
						Addresses: []string{s.user1Addr.String()},
					},
				},
			},
			{
				Key: &types.RegistryKey{
					AssetClassId: s.validNFTClass.Id,
					NftId:        s.validNFT2.Id,
				},
				Roles: []types.RolesEntry{
					{
						Role:      servicerRole,
						Addresses: []string{s.user1Addr.String()},
					},
				},
			},
		},
	}

	handler := s.app.MsgServiceRouter().Handler(msg)
	_, err := handler(s.ctx, msg)
	s.Require().NoError(err)

	registry, err := s.registryKeeper.GetRegistry(s.ctx, &types.RegistryKey{
		AssetClassId: s.validNFTClass.Id,
		NftId:        s.validNFT.Id,
	})
	s.Require().NoError(err)
	s.Require().NotNil(registry)
	s.Require().Equal(s.validNFTClass.Id, registry.Key.AssetClassId)
	s.Require().Equal(s.validNFT.Id, registry.Key.NftId)

	servicerRoleAddress := registry.Roles[0].Addresses[0]
	s.Require().Equal(s.user1Addr.String(), servicerRoleAddress)

	registry, err = s.registryKeeper.GetRegistry(s.ctx, &types.RegistryKey{
		AssetClassId: s.validNFTClass.Id,
		NftId:        s.validNFT2.Id,
	})
	s.Require().NoError(err)
	s.Require().NotNil(registry)
	s.Require().Equal(s.validNFTClass.Id, registry.Key.AssetClassId)
	s.Require().Equal(s.validNFT2.Id, registry.Key.NftId)

	servicerRoleAddress = registry.Roles[0].Addresses[0]
	s.Require().Equal(s.user1Addr.String(), servicerRoleAddress)
}
