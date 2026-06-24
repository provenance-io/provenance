package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/x/nft"
	nftkeeper "cosmossdk.io/x/nft/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/registry/keeper"
	"github.com/provenance-io/provenance/x/registry/types"
)

// RegistryClassAcceptanceTestSuite covers Phase C1: registry classes provide maintainer-managed,
// asset-class-level authorization configuration that overrides the module's static default policy
// via two-tier resolution (class policy -> static default / NFT-owner fallback).
type RegistryClassAcceptanceTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	registryKeeper keeper.Keeper
	nftKeeper      nftkeeper.Keeper
	msgServer      types.MsgServer

	nftClass nft.Class

	nftOwner   string
	maintainer string
	controller string
	stranger   string
	grantee    string
}

func TestRegistryClassAcceptanceTestSuite(t *testing.T) {
	suite.Run(t, new(RegistryClassAcceptanceTestSuite))
}

func (s *RegistryClassAcceptanceTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: time.Now()}).WithBlockTime(time.Now())

	s.nftKeeper = s.app.NFTKeeper
	s.registryKeeper = s.app.RegistryKeeper
	s.msgServer = keeper.NewMsgServer(s.registryKeeper)

	s.nftOwner = genAddr()
	s.maintainer = genAddr()
	s.controller = genAddr()
	s.stranger = genAddr()
	s.grantee = genAddr()

	s.nftClass = nft.Class{Id: "registry-class-test-nft-class"}
	s.nftKeeper.SaveClass(s.ctx, s.nftClass)
}

// mintNFT mints an NFT owned by s.nftOwner and returns its registry key.
func (s *RegistryClassAcceptanceTestSuite) mintNFT(id string) *types.RegistryKey {
	n := nft.NFT{ClassId: s.nftClass.Id, Id: id}
	ownerAddr, err := sdk.AccAddressFromBech32(s.nftOwner)
	s.Require().NoError(err)
	s.Require().NoError(s.nftKeeper.Mint(s.ctx, n, ownerAddr))
	return &types.RegistryKey{AssetClassId: s.nftClass.Id, NftId: id}
}

// servicerRoleAuthorization builds a class policy for the SERVICER role that is satisfied by the
// current CONTROLLER's signature alone. This deviates from the static default (which only governs
// CONTROLLER and would otherwise fall back to NFT-owner authorization for SERVICER), letting us
// observe two-tier resolution.
func servicerRoleAuthorization() types.RoleAuthorization {
	return types.RoleAuthorization{
		Role: types.RegistryRole_REGISTRY_ROLE_SERVICER,
		Authorizations: []*types.Authorization{
			{
				Description: "current controller may manage servicers",
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

func (s *RegistryClassAcceptanceTestSuite) createClass(classID string, auths []types.RoleAuthorization) {
	_, err := s.msgServer.CreateRegistryClass(s.ctx, &types.MsgCreateRegistryClass{
		Signer:             s.maintainer,
		RegistryClassId:    classID,
		AssetClassId:       s.nftClass.Id,
		Maintainer:         s.maintainer,
		RoleAuthorizations: auths,
	})
	s.Require().NoError(err)
}

// TestCreateRegistryClass exercises class creation success and failure paths.
func (s *RegistryClassAcceptanceTestSuite) TestCreateRegistryClass() {
	s.Run("success: signer is maintainer", func() {
		_, err := s.msgServer.CreateRegistryClass(s.ctx, &types.MsgCreateRegistryClass{
			Signer:          s.maintainer,
			RegistryClassId: "class-create-ok",
			AssetClassId:    s.nftClass.Id,
			Maintainer:      s.maintainer,
		})
		s.Require().NoError(err)

		got, err := s.registryKeeper.GetRegistryClass(s.ctx, "class-create-ok")
		s.Require().NoError(err)
		s.Require().NotNil(got)
		s.Require().Equal(s.maintainer, got.Maintainer)
		s.Require().Equal(s.nftClass.Id, got.AssetClassId)
	})

	s.Run("reject: signer is not maintainer (ValidateBasic)", func() {
		msg := &types.MsgCreateRegistryClass{
			Signer:          s.stranger,
			RegistryClassId: "class-bad-signer",
			AssetClassId:    s.nftClass.Id,
			Maintainer:      s.maintainer,
		}
		err := msg.ValidateBasic()
		s.Require().Error(err)
		s.Require().ErrorContains(err, "signer")
	})

	s.Run("reject: duplicate registry class id", func() {
		s.createClass("class-dup", nil)
		_, err := s.msgServer.CreateRegistryClass(s.ctx, &types.MsgCreateRegistryClass{
			Signer:          s.maintainer,
			RegistryClassId: "class-dup",
			AssetClassId:    s.nftClass.Id,
			Maintainer:      s.maintainer,
		})
		s.Require().Error(err)
		s.Require().ErrorContains(err, "class-dup")
	})
}

// TestUpdateRegistryClassRoleAuthorization exercises the maintainer-only update path.
func (s *RegistryClassAcceptanceTestSuite) TestUpdateRegistryClassRoleAuthorization() {
	s.createClass("class-update", nil)

	s.Run("success: maintainer updates authorizations", func() {
		_, err := s.msgServer.UpdateRegistryClassRoleAuthorization(s.ctx, &types.MsgUpdateRegistryClassRoleAuthorization{
			Signer:             s.maintainer,
			RegistryClassId:    "class-update",
			RoleAuthorizations: []types.RoleAuthorization{servicerRoleAuthorization()},
		})
		s.Require().NoError(err)

		got, err := s.registryKeeper.GetRegistryClass(s.ctx, "class-update")
		s.Require().NoError(err)
		s.Require().NotNil(got)
		s.Require().Len(got.RoleAuthorizations, 1)
		s.Require().Equal(types.RegistryRole_REGISTRY_ROLE_SERVICER, got.RoleAuthorizations[0].Role)
	})

	s.Run("reject: non-maintainer cannot update", func() {
		_, err := s.msgServer.UpdateRegistryClassRoleAuthorization(s.ctx, &types.MsgUpdateRegistryClassRoleAuthorization{
			Signer:          s.stranger,
			RegistryClassId: "class-update",
		})
		s.Require().Error(err)
		s.Require().ErrorContains(err, "maintainer")
	})

	s.Run("reject: unknown registry class", func() {
		_, err := s.msgServer.UpdateRegistryClassRoleAuthorization(s.ctx, &types.MsgUpdateRegistryClassRoleAuthorization{
			Signer:          s.maintainer,
			RegistryClassId: "does-not-exist",
		})
		s.Require().Error(err)
		s.Require().ErrorContains(err, "does-not-exist")
	})
}

// TestRegisterNFTWithRegistryClass verifies that registering against a class requires the class to
// exist and persists the reference on the entry.
func (s *RegistryClassAcceptanceTestSuite) TestRegisterNFTWithRegistryClass() {
	s.createClass("class-register", []types.RoleAuthorization{servicerRoleAuthorization()})

	s.Run("reject: unknown registry class", func() {
		key := s.mintNFT("nft-missing-class")
		_, err := s.msgServer.RegisterNFT(s.ctx, &types.MsgRegisterNFT{
			Signer:          s.nftOwner,
			Key:             key,
			RegistryClassId: "no-such-class",
		})
		s.Require().Error(err)
		s.Require().ErrorContains(err, "no-such-class")
	})

	s.Run("success: entry stores registry class id", func() {
		key := s.mintNFT("nft-with-class")
		_, err := s.msgServer.RegisterNFT(s.ctx, &types.MsgRegisterNFT{
			Signer:          s.nftOwner,
			Key:             key,
			RegistryClassId: "class-register",
		})
		s.Require().NoError(err)

		entry, err := s.registryKeeper.GetRegistry(s.ctx, key)
		s.Require().NoError(err)
		s.Require().NotNil(entry)
		s.Require().Equal("class-register", entry.RegistryClassId)
	})
}

// TestTwoTierResolution proves that an entry referencing a class is governed by the class policy,
// while an otherwise-identical entry with no class falls back to the static default (NFT-owner) for
// the same role.
func (s *RegistryClassAcceptanceTestSuite) TestTwoTierResolution() {
	s.createClass("class-servicer", []types.RoleAuthorization{servicerRoleAuthorization()})

	roles := []types.RolesEntry{
		{Role: types.RegistryRole_REGISTRY_ROLE_CONTROLLER, Addresses: []string{s.controller}},
	}

	s.Run("class policy: current controller may grant SERVICER", func() {
		key := s.mintNFT("nft-classed")
		_, err := s.msgServer.RegisterNFT(s.ctx, &types.MsgRegisterNFT{
			Signer:          s.nftOwner,
			Key:             key,
			Roles:           roles,
			RegistryClassId: "class-servicer",
		})
		s.Require().NoError(err)

		// The controller (not the NFT owner) satisfies the class policy for SERVICER.
		_, err = s.msgServer.GrantRole(s.ctx, &types.MsgGrantRole{
			Signer:    s.controller,
			Key:       key,
			Role:      types.RegistryRole_REGISTRY_ROLE_SERVICER,
			Addresses: []string{s.grantee},
		})
		s.Require().NoError(err)

		hasRole, err := s.registryKeeper.HasRole(s.ctx, key, types.RegistryRole_REGISTRY_ROLE_SERVICER, s.grantee)
		s.Require().NoError(err)
		s.Require().True(hasRole)
	})

	s.Run("legacy fallback: controller cannot grant SERVICER without a class", func() {
		key := s.mintNFT("nft-unclassed")
		_, err := s.msgServer.RegisterNFT(s.ctx, &types.MsgRegisterNFT{
			Signer: s.nftOwner,
			Key:    key,
			Roles:  roles,
		})
		s.Require().NoError(err)

		// No class -> SERVICER is not in the static map -> NFT-owner authorization is required.
		_, err = s.msgServer.GrantRole(s.ctx, &types.MsgGrantRole{
			Signer:    s.controller,
			Key:       key,
			Role:      types.RegistryRole_REGISTRY_ROLE_SERVICER,
			Addresses: []string{s.grantee},
		})
		s.Require().Error(err)

		// The NFT owner can grant under the legacy fallback.
		_, err = s.msgServer.GrantRole(s.ctx, &types.MsgGrantRole{
			Signer:    s.nftOwner,
			Key:       key,
			Role:      types.RegistryRole_REGISTRY_ROLE_SERVICER,
			Addresses: []string{s.grantee},
		})
		s.Require().NoError(err)
	})
}

// TestGenesisRoundTrip verifies registry classes survive an export/import cycle.
func (s *RegistryClassAcceptanceTestSuite) TestGenesisRoundTrip() {
	s.createClass("class-genesis", []types.RoleAuthorization{servicerRoleAuthorization()})

	exported := s.registryKeeper.ExportGenesis(s.ctx)
	s.Require().NoError(exported.Validate())
	s.Require().Len(exported.RegistryClasses, 1)
	s.Require().Equal("class-genesis", exported.RegistryClasses[0].RegistryClassId)

	// Re-initialize into a fresh app and confirm the class is present.
	freshApp := app.Setup(s.T())
	freshCtx := freshApp.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: time.Now()}).WithBlockTime(time.Now())
	freshApp.RegistryKeeper.InitGenesis(freshCtx, exported)

	got, err := freshApp.RegistryKeeper.GetRegistryClass(freshCtx, "class-genesis")
	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Require().Equal(s.maintainer, got.Maintainer)
	s.Require().Len(got.RoleAuthorizations, 1)
}
