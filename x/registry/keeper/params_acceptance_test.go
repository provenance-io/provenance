package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/x/nft"
	nftkeeper "cosmossdk.io/x/nft/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/registry/keeper"
	"github.com/provenance-io/provenance/x/registry/types"
)

// ParamsAcceptanceTestSuite covers Phase C2: the registry module's governance-managed params.
// The params hold the default role authorization policies, which form the middle tier of the
// two-tier resolution (registry class policy -> module params default -> NFT-owner fallback).
type ParamsAcceptanceTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	registryKeeper keeper.Keeper
	nftKeeper      nftkeeper.Keeper
	msgServer      types.MsgServer

	authority string

	nftClass nft.Class

	nftOwner   string
	controller string
	stranger   string
	grantee    string
}

func TestParamsAcceptanceTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsAcceptanceTestSuite))
}

func (s *ParamsAcceptanceTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: time.Now()}).WithBlockTime(time.Now())

	s.nftKeeper = s.app.NFTKeeper
	s.registryKeeper = s.app.RegistryKeeper
	s.msgServer = keeper.NewMsgServer(s.registryKeeper)

	s.authority = authtypes.NewModuleAddress(govtypes.ModuleName).String()

	s.nftOwner = genAddr()
	s.controller = genAddr()
	s.stranger = genAddr()
	s.grantee = genAddr()

	s.nftClass = nft.Class{Id: "params-test-nft-class"}
	s.nftKeeper.SaveClass(s.ctx, s.nftClass)
}

func (s *ParamsAcceptanceTestSuite) mintNFT(id string) *types.RegistryKey {
	n := nft.NFT{ClassId: s.nftClass.Id, Id: id}
	ownerAddr, err := sdk.AccAddressFromBech32(s.nftOwner)
	s.Require().NoError(err)
	s.Require().NoError(s.nftKeeper.Mint(s.ctx, n, ownerAddr))
	return &types.RegistryKey{AssetClassId: s.nftClass.Id, NftId: id}
}

// TestDefaultParams confirms the module starts with no role policies, preserving the original
// NFT-owner-only authorization behavior. The CONTROLLER policy is only an example, not a default.
func (s *ParamsAcceptanceTestSuite) TestDefaultParams() {
	params := s.registryKeeper.GetParams(s.ctx)
	s.Require().NoError(params.Validate())
	s.Require().Empty(params.RoleAuthorizations, "default params should define no role policies")
}

// TestUpdateParamsAuthority confirms only the governance authority may update params.
func (s *ParamsAcceptanceTestSuite) TestUpdateParamsAuthority() {
	newParams := types.Params{
		RoleAuthorizations: []types.RoleAuthorization{servicerRoleAuthorization()},
	}

	s.Run("reject: non-authority signer", func() {
		_, err := s.msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{
			Authority: s.stranger,
			Params:    newParams,
		})
		s.Require().Error(err)
		s.Require().ErrorContains(err, "expected")
	})

	s.Run("success: governance authority", func() {
		_, err := s.msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{
			Authority: s.authority,
			Params:    newParams,
		})
		s.Require().NoError(err)

		got := s.registryKeeper.GetParams(s.ctx)
		s.Require().Len(got.RoleAuthorizations, 1)
		s.Require().Equal(types.RegistryRole_REGISTRY_ROLE_SERVICER, got.RoleAuthorizations[0].Role)
	})
}

// TestParamsDrivenDefaultTier proves that the middle (params) tier of the two-tier resolution is
// governance-managed: changing params changes the default authorization for entries with no class.
func (s *ParamsAcceptanceTestSuite) TestParamsDrivenDefaultTier() {
	roles := []types.RolesEntry{
		{Role: types.RegistryRole_REGISTRY_ROLE_CONTROLLER, Addresses: []string{s.controller}},
	}

	// Before changing params: SERVICER has no default policy, so granting it falls back to
	// NFT-owner authorization. The controller (not the owner) cannot grant it.
	keyBefore := s.mintNFT("nft-before-params")
	_, err := s.msgServer.RegisterNFT(s.ctx, &types.MsgRegisterNFT{
		Signer: s.nftOwner,
		Key:    keyBefore,
		Roles:  roles,
	})
	s.Require().NoError(err)

	_, err = s.msgServer.GrantRole(s.ctx, &types.MsgGrantRole{
		Signer:    s.controller,
		Key:       keyBefore,
		Role:      types.RegistryRole_REGISTRY_ROLE_SERVICER,
		Addresses: []string{s.grantee},
	})
	s.Require().Error(err, "without a SERVICER default policy, controller cannot grant SERVICER")

	// Governance adds a SERVICER default policy (satisfied by the current controller).
	_, err = s.msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{
		Authority: s.authority,
		Params: types.Params{
			RoleAuthorizations: []types.RoleAuthorization{servicerRoleAuthorization()},
		},
	})
	s.Require().NoError(err)

	// After changing params: the same operation on a fresh (classless) entry now uses the new
	// default policy, so the controller can grant SERVICER.
	keyAfter := s.mintNFT("nft-after-params")
	_, err = s.msgServer.RegisterNFT(s.ctx, &types.MsgRegisterNFT{
		Signer: s.nftOwner,
		Key:    keyAfter,
		Roles:  roles,
	})
	s.Require().NoError(err)

	_, err = s.msgServer.GrantRole(s.ctx, &types.MsgGrantRole{
		Signer:    s.controller,
		Key:       keyAfter,
		Role:      types.RegistryRole_REGISTRY_ROLE_SERVICER,
		Addresses: []string{s.grantee},
	})
	s.Require().NoError(err, "with a SERVICER default policy, controller can grant SERVICER")

	hasRole, err := s.registryKeeper.HasRole(s.ctx, keyAfter, types.RegistryRole_REGISTRY_ROLE_SERVICER, s.grantee)
	s.Require().NoError(err)
	s.Require().True(hasRole)
}

// TestParamsGenesisRoundTrip verifies params survive an export/import cycle.
func (s *ParamsAcceptanceTestSuite) TestParamsGenesisRoundTrip() {
	_, err := s.msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{
		Authority: s.authority,
		Params: types.Params{
			RoleAuthorizations: []types.RoleAuthorization{servicerRoleAuthorization()},
		},
	})
	s.Require().NoError(err)

	exported := s.registryKeeper.ExportGenesis(s.ctx)
	s.Require().NoError(exported.Validate())
	s.Require().Len(exported.Params.RoleAuthorizations, 1)

	freshApp := app.Setup(s.T())
	freshCtx := freshApp.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: time.Now()}).WithBlockTime(time.Now())
	freshApp.RegistryKeeper.InitGenesis(freshCtx, exported)

	got := freshApp.RegistryKeeper.GetParams(freshCtx)
	s.Require().Len(got.RoleAuthorizations, 1)
	s.Require().Equal(types.RegistryRole_REGISTRY_ROLE_SERVICER, got.RoleAuthorizations[0].Role)
}
