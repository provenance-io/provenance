package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/x/nft"
	nftkeeper "cosmossdk.io/x/nft/keeper"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/registry/keeper"
	"github.com/provenance-io/provenance/x/registry/types"
)

// RoleChangeAccumulationAcceptanceTestSuite implements the ticket (sc-512248) Controller
// authorization matrix using Option B: single-signer messages whose approvals accumulate in
// registry state until the role's policy is satisfied, at which point the change auto-applies.
//
// Required approvals for a Controller update:
//   - Current Controller
//   - Current Secured Party for eNote (only if set)
//   - New Controller
//
// Unlike the native multi-signer model, every message here is single-signer, so each approval is
// delegable via authz (see TestControllerUpdate_Accumulation_ViaAuthz).
type RoleChangeAccumulationAcceptanceTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	registryKeeper keeper.Keeper
	nftKeeper      nftkeeper.Keeper
	msgServer      types.MsgServer

	nftClass nft.Class

	nftOwner          string
	currentController string
	newController     string
	securedParty      string
	stranger          string
}

func TestRoleChangeAccumulationAcceptanceTestSuite(t *testing.T) {
	suite.Run(t, new(RoleChangeAccumulationAcceptanceTestSuite))
}

// genAddr generates a fresh bech32 account address.
func genAddr() string {
	return sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()
}

func (s *RoleChangeAccumulationAcceptanceTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: time.Now()}).WithBlockTime(time.Now())

	s.nftKeeper = s.app.NFTKeeper
	s.registryKeeper = s.app.RegistryKeeper
	s.msgServer = keeper.NewMsgServer(s.registryKeeper)

	s.nftOwner = genAddr()
	s.currentController = genAddr()
	s.newController = genAddr()
	s.securedParty = genAddr()
	s.stranger = genAddr()

	s.nftClass = nft.Class{Id: "accumulation-test-nft-class-id"}
	s.nftKeeper.SaveClass(s.ctx, s.nftClass)
}

func (s *RoleChangeAccumulationAcceptanceTestSuite) mintNFT(id string) *types.RegistryKey {
	n := nft.NFT{ClassId: s.nftClass.Id, Id: id}
	ownerAddr, err := sdk.AccAddressFromBech32(s.nftOwner)
	s.Require().NoError(err)
	s.Require().NoError(s.nftKeeper.Mint(s.ctx, n, ownerAddr))
	return &types.RegistryKey{AssetClassId: s.nftClass.Id, NftId: id}
}

func (s *RoleChangeAccumulationAcceptanceTestSuite) setupRegistry(key *types.RegistryKey, roles []types.RolesEntry) {
	s.Require().NoError(s.registryKeeper.CreateRegistry(s.ctx, key, roles))
}

// proposeNewController opens a pending GRANT of the CONTROLLER role to newController, signed by
// proposer, and returns the change id.
func (s *RoleChangeAccumulationAcceptanceTestSuite) proposeNewController(key *types.RegistryKey, proposer string) (string, bool) {
	resp, err := s.msgServer.ProposeRoleChange(s.ctx, &types.MsgProposeRoleChange{
		Signer:    proposer,
		Key:       key,
		Role:      types.RegistryRole_REGISTRY_ROLE_CONTROLLER,
		Operation: types.RoleChangeOperation_ROLE_CHANGE_OPERATION_GRANT,
		Addresses: []string{s.newController},
	})
	s.Require().NoError(err)
	return resp.ChangeId, resp.Applied
}

// approve records an approval for the pending change, signed by signer.
func (s *RoleChangeAccumulationAcceptanceTestSuite) approve(changeID, signer string) bool {
	resp, err := s.msgServer.ApproveRoleChange(s.ctx, &types.MsgApproveRoleChange{
		Signer:   signer,
		ChangeId: changeID,
	})
	s.Require().NoError(err)
	return resp.Applied
}

func (s *RoleChangeAccumulationAcceptanceTestSuite) roleAddresses(key *types.RegistryKey, role types.RegistryRole) []string {
	entry, err := s.registryKeeper.GetRegistry(s.ctx, key)
	s.Require().NoError(err)
	s.Require().NotNil(entry)
	for _, re := range entry.Roles {
		if re.Role == role {
			return re.Addresses
		}
	}
	return nil
}

// --- Scenario A: Controller set, no Secured Party for eNote ---------------------------

func (s *RoleChangeAccumulationAcceptanceTestSuite) TestControllerSet_NoSecuredParty() {
	controllerRole := types.RegistryRole_REGISTRY_ROLE_CONTROLLER

	baseRoles := func() []types.RolesEntry {
		return []types.RolesEntry{
			{Role: controllerRole, Addresses: []string{s.currentController}},
		}
	}

	s.Run("happy path: current controller proposes, new controller approves", func() {
		key := s.mintNFT("acc-a-happy")
		s.setupRegistry(key, baseRoles())

		changeID, applied := s.proposeNewController(key, s.currentController)
		s.Require().False(applied, "should not apply with only the current controller's approval")
		s.Require().NotContains(s.roleAddresses(key, controllerRole), s.newController)

		applied = s.approve(changeID, s.newController)
		s.Require().True(applied, "current + new controller approvals should satisfy the policy")
		s.Require().Contains(s.roleAddresses(key, controllerRole), s.newController)
	})

	s.Run("incomplete: only current controller approves", func() {
		key := s.mintNFT("acc-a-cc-only")
		s.setupRegistry(key, baseRoles())

		_, applied := s.proposeNewController(key, s.currentController)
		s.Require().False(applied)
		s.Require().NotContains(s.roleAddresses(key, controllerRole), s.newController)
	})

	s.Run("incomplete: only new controller approves", func() {
		key := s.mintNFT("acc-a-nc-only")
		s.setupRegistry(key, baseRoles())

		_, applied := s.proposeNewController(key, s.newController)
		s.Require().False(applied, "missing current controller approval")
		s.Require().NotContains(s.roleAddresses(key, controllerRole), s.newController)
	})
}

// --- Scenario B: Controller set, Secured Party for eNote set --------------------------

func (s *RoleChangeAccumulationAcceptanceTestSuite) TestControllerSet_WithSecuredParty() {
	controllerRole := types.RegistryRole_REGISTRY_ROLE_CONTROLLER
	securedPartyRole := types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE

	baseRoles := func() []types.RolesEntry {
		return []types.RolesEntry{
			{Role: controllerRole, Addresses: []string{s.currentController}},
			{Role: securedPartyRole, Addresses: []string{s.securedParty}},
		}
	}

	s.Run("happy path: current controller, secured party, and new controller approve", func() {
		key := s.mintNFT("acc-b-happy")
		s.setupRegistry(key, baseRoles())

		changeID, applied := s.proposeNewController(key, s.currentController)
		s.Require().False(applied)

		applied = s.approve(changeID, s.securedParty)
		s.Require().False(applied, "still missing new controller")

		applied = s.approve(changeID, s.newController)
		s.Require().True(applied, "all three required approvals present")
		s.Require().Contains(s.roleAddresses(key, controllerRole), s.newController)
	})

	s.Run("incomplete: current controller and secured party (missing new controller)", func() {
		key := s.mintNFT("acc-b-cc-sp")
		s.setupRegistry(key, baseRoles())

		changeID, _ := s.proposeNewController(key, s.currentController)
		applied := s.approve(changeID, s.securedParty)
		s.Require().False(applied)
		s.Require().NotContains(s.roleAddresses(key, controllerRole), s.newController)
	})

	s.Run("incomplete: current controller and new controller (missing secured party)", func() {
		key := s.mintNFT("acc-b-cc-nc")
		s.setupRegistry(key, baseRoles())

		changeID, _ := s.proposeNewController(key, s.currentController)
		applied := s.approve(changeID, s.newController)
		s.Require().False(applied, "secured party approval is required when it is set")
		s.Require().NotContains(s.roleAddresses(key, controllerRole), s.newController)
	})

	s.Run("incomplete: secured party and new controller (missing current controller)", func() {
		key := s.mintNFT("acc-b-sp-nc")
		s.setupRegistry(key, baseRoles())

		changeID, _ := s.proposeNewController(key, s.securedParty)
		applied := s.approve(changeID, s.newController)
		s.Require().False(applied, "current controller approval is always required (no unilateral takeover)")
		s.Require().NotContains(s.roleAddresses(key, controllerRole), s.newController)
	})

	s.Run("incomplete: a stranger approves", func() {
		key := s.mintNFT("acc-b-stranger")
		s.setupRegistry(key, baseRoles())

		changeID, _ := s.proposeNewController(key, s.currentController)
		applied := s.approve(changeID, s.stranger)
		s.Require().False(applied, "an unrelated approval does not satisfy the policy")
		s.Require().NotContains(s.roleAddresses(key, controllerRole), s.newController)
	})
}

// --- AuthZ delegation -----------------------------------------------------------------
//
// The single-signer accumulation model is the whole reason Option B satisfies the ticket's authz
// requirement: because each MsgApproveRoleChange has exactly one signer, a party can grant another
// account authority to submit its approval via authz MsgExec. This is impossible with the native
// multi-signer MsgGrantRole (see authorization_acceptance_test.go,
// TestControllerUpdate_ViaAuthzGrants).

func (s *RoleChangeAccumulationAcceptanceTestSuite) mustAddr(addr string) sdk.AccAddress {
	a, err := sdk.AccAddressFromBech32(addr)
	s.Require().NoError(err)
	return a
}

// grantApproveAuthz grants executor authority to submit MsgApproveRoleChange on behalf of granter.
func (s *RoleChangeAccumulationAcceptanceTestSuite) grantApproveAuthz(executor, granter string) {
	auth := authz.NewGenericAuthorization(sdk.MsgTypeURL(&types.MsgApproveRoleChange{}))
	s.Require().NoError(s.app.AuthzKeeper.SaveGrant(s.ctx, s.mustAddr(executor), s.mustAddr(granter), auth, nil))
}

// execApproveViaAuthz submits MsgApproveRoleChange (signed by granter) wrapped in a MsgExec run by
// executor.
func (s *RoleChangeAccumulationAcceptanceTestSuite) execApproveViaAuthz(executor, granter, changeID string) {
	inner := &types.MsgApproveRoleChange{Signer: granter, ChangeId: changeID}
	msgExec := authz.NewMsgExec(s.mustAddr(executor), []sdk.Msg{inner})
	_, err := s.app.AuthzKeeper.Exec(s.ctx, &msgExec)
	s.Require().NoError(err, "authz must dispatch a single-signer approval")
}

func (s *RoleChangeAccumulationAcceptanceTestSuite) TestControllerUpdate_Accumulation_ViaAuthz() {
	controllerRole := types.RegistryRole_REGISTRY_ROLE_CONTROLLER
	securedPartyRole := types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE
	executor := genAddr()

	key := s.mintNFT("acc-authz")
	s.setupRegistry(key, []types.RolesEntry{
		{Role: controllerRole, Addresses: []string{s.currentController}},
		{Role: securedPartyRole, Addresses: []string{s.securedParty}},
	})

	// Current controller proposes directly.
	changeID, applied := s.proposeNewController(key, s.currentController)
	s.Require().False(applied)

	// Secured party and new controller each delegate their approval to the executor via authz.
	s.grantApproveAuthz(executor, s.securedParty)
	s.grantApproveAuthz(executor, s.newController)

	s.execApproveViaAuthz(executor, s.securedParty, changeID)
	s.Require().NotContains(s.roleAddresses(key, controllerRole), s.newController,
		"not yet applied after only the secured party's delegated approval")

	s.execApproveViaAuthz(executor, s.newController, changeID)
	s.Require().Contains(s.roleAddresses(key, controllerRole), s.newController,
		"change auto-applies once all required approvals (incl. authz-delegated ones) accumulate")

	// The pending change should have been cleaned up after applying.
	pending, err := s.registryKeeper.GetPendingRoleChange(s.ctx, changeID)
	s.Require().NoError(err)
	s.Require().Nil(pending, "pending change should be removed after it applies")
}
