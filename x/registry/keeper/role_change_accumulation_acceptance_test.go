package keeper_test

import (
	"strings"
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

	// These tests exercise the CONTROLLER authorization policy, which is not a chain default. Install
	// it as the module's default policy so classless entries are governed by it.
	s.Require().NoError(s.registryKeeper.SetParams(s.ctx, types.Params{
		RoleAuthorizations: types.ControllerRoleAuthorizations(),
	}))
}

func (s *RoleChangeAccumulationAcceptanceTestSuite) mintNFT(id string) *types.RegistryKey {
	n := nft.NFT{ClassId: s.nftClass.Id, Id: id}
	ownerAddr, err := sdk.AccAddressFromBech32(s.nftOwner)
	s.Require().NoError(err)
	s.Require().NoError(s.nftKeeper.Mint(s.ctx, n, ownerAddr))
	return &types.RegistryKey{AssetClassId: s.nftClass.Id, NftId: id}
}

func (s *RoleChangeAccumulationAcceptanceTestSuite) setupRegistry(key *types.RegistryKey, roles []types.RolesEntry) {
	s.Require().NoError(s.registryKeeper.CreateRegistry(s.ctx, key, roles, ""))
}

// proposeNewController opens a pending GRANT of the CONTROLLER role to newController, signed by
// proposer, and returns the change id.
func (s *RoleChangeAccumulationAcceptanceTestSuite) proposeNewController(key *types.RegistryKey, proposer string) (string, bool) {
	resp, err := s.msgServer.ProposeRoleChange(s.ctx, &types.MsgProposeRoleChange{
		Signer: proposer,
		Key:    key,
		RoleUpdates: []types.RoleUpdate{{
			Role:      types.RegistryRole_REGISTRY_ROLE_CONTROLLER,
			Addresses: []string{s.newController},
		}},
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

// proposeRevokeController opens a pending REVOKE of the CONTROLLER role from addr by proposing an
// empty desired-state for the role, signed by proposer, and returns the change id and whether it
// applied immediately.
func (s *RoleChangeAccumulationAcceptanceTestSuite) proposeRevokeController(key *types.RegistryKey, addr, proposer string) (string, bool) {
	current := s.roleAddresses(key, types.RegistryRole_REGISTRY_ROLE_CONTROLLER)
	desired := make([]string, 0, len(current))
	for _, a := range current {
		if a != addr {
			desired = append(desired, a)
		}
	}
	resp, err := s.msgServer.ProposeRoleChange(s.ctx, &types.MsgProposeRoleChange{
		Signer: proposer,
		Key:    key,
		RoleUpdates: []types.RoleUpdate{{
			Role:      types.RegistryRole_REGISTRY_ROLE_CONTROLLER,
			Addresses: desired,
		}},
	})
	s.Require().NoError(err)
	return resp.ChangeId, resp.Applied
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

		// The stranger is not referenced by the controller policy, so its approval must be
		// ignored entirely rather than accumulating in the pending change's approval set.
		pending, err := s.registryKeeper.GetPendingRoleChange(s.ctx, changeID)
		s.Require().NoError(err)
		s.Require().NotNil(pending)
		s.Require().NotContains(pending.Approvals, s.stranger, "ineligible approvals are not recorded")
		s.Require().Equal([]string{s.currentController}, pending.Approvals)
	})

	s.Run("incomplete: only current controller approves", func() {
		key := s.mintNFT("acc-b-cc-only")
		s.setupRegistry(key, baseRoles())

		_, applied := s.proposeNewController(key, s.currentController)
		s.Require().False(applied, "missing secured party and new controller approvals")
		s.Require().NotContains(s.roleAddresses(key, controllerRole), s.newController)
	})

	s.Run("incomplete: only new controller approves", func() {
		key := s.mintNFT("acc-b-nc-only")
		s.setupRegistry(key, baseRoles())

		_, applied := s.proposeNewController(key, s.newController)
		s.Require().False(applied, "missing current controller and secured party approvals")
		s.Require().NotContains(s.roleAddresses(key, controllerRole), s.newController)
	})

	s.Run("incomplete: only secured party approves", func() {
		key := s.mintNFT("acc-b-sp-only")
		s.setupRegistry(key, baseRoles())

		changeID, applied := s.proposeNewController(key, s.securedParty)
		s.Require().False(applied, "proposing as the secured party records only its own approval")

		// Re-stating the secured party's approval must remain insufficient.
		applied = s.approve(changeID, s.securedParty)
		s.Require().False(applied, "missing current controller and new controller approvals")
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

// grantProposeAuthz grants executor authority to submit MsgProposeRoleChange on behalf of granter.
func (s *RoleChangeAccumulationAcceptanceTestSuite) grantProposeAuthz(executor, granter string) {
	auth := authz.NewGenericAuthorization(sdk.MsgTypeURL(&types.MsgProposeRoleChange{}))
	s.Require().NoError(s.app.AuthzKeeper.SaveGrant(s.ctx, s.mustAddr(executor), s.mustAddr(granter), auth, nil))
}

// execProposeNewControllerViaAuthz submits MsgProposeRoleChange (signed by granter) to grant the
// CONTROLLER role to newController, wrapped in a MsgExec run by executor. Returns the change id.
func (s *RoleChangeAccumulationAcceptanceTestSuite) execProposeNewControllerViaAuthz(executor, granter string, key *types.RegistryKey) string {
	inner := &types.MsgProposeRoleChange{
		Signer: granter,
		Key:    key,
		RoleUpdates: []types.RoleUpdate{{
			Role:      types.RegistryRole_REGISTRY_ROLE_CONTROLLER,
			Addresses: []string{s.newController},
		}},
	}
	msgExec := authz.NewMsgExec(s.mustAddr(executor), []sdk.Msg{inner})
	_, err := s.app.AuthzKeeper.Exec(s.ctx, &msgExec)
	s.Require().NoError(err, "authz must dispatch a single-signer proposal")
	return types.NewPendingRoleChangeID(key, inner.RoleUpdates)
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

// TestControllerUpdate_NoSecuredParty_HappyPath_ViaAuthz proves the Scenario 1 happy path (no
// Secured Party set) also completes when both required approvals are delegated through authz.
func (s *RoleChangeAccumulationAcceptanceTestSuite) TestControllerUpdate_NoSecuredParty_HappyPath_ViaAuthz() {
	controllerRole := types.RegistryRole_REGISTRY_ROLE_CONTROLLER
	executor := genAddr()

	key := s.mintNFT("acc-authz-no-sp")
	s.setupRegistry(key, []types.RolesEntry{
		{Role: controllerRole, Addresses: []string{s.currentController}},
	})

	// Both the current controller (proposer) and the new controller delegate via authz.
	s.grantProposeAuthz(executor, s.currentController)
	s.grantApproveAuthz(executor, s.newController)

	changeID := s.execProposeNewControllerViaAuthz(executor, s.currentController, key)
	s.Require().NotContains(s.roleAddresses(key, controllerRole), s.newController,
		"not yet applied after only the current controller's delegated proposal")

	s.execApproveViaAuthz(executor, s.newController, changeID)
	s.Require().Contains(s.roleAddresses(key, controllerRole), s.newController,
		"current + new controller approvals (both authz-delegated) satisfy the policy")

	pending, err := s.registryKeeper.GetPendingRoleChange(s.ctx, changeID)
	s.Require().NoError(err)
	s.Require().Nil(pending, "pending change should be removed after it applies")
}

// TestControllerUpdate_RejectMatrix_ViaAuthz re-runs the ticket's reject scenarios with every
// approval (and the proposal) delegated through authz MsgExec, proving that an incomplete set of
// authz-delegated approvals never satisfies the policy. A neutral executor carries each party's
// delegated authority, so no required party ever signs a transaction directly.
func (s *RoleChangeAccumulationAcceptanceTestSuite) TestControllerUpdate_RejectMatrix_ViaAuthz() {
	controllerRole := types.RegistryRole_REGISTRY_ROLE_CONTROLLER
	securedPartyRole := types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE

	// authzApprovers opens the pending change as the first party (via authz) and submits the rest of
	// the partial approver set (via authz), then asserts the change never applied.
	authzApprovers := func(name string, withSecuredParty bool, approvers []string) {
		s.Run(name, func() {
			executor := genAddr()
			key := s.mintNFT("acc-authz-reject-" + strings.ReplaceAll(name, "_", "-"))

			roles := []types.RolesEntry{{Role: controllerRole, Addresses: []string{s.currentController}}}
			if withSecuredParty {
				roles = append(roles, types.RolesEntry{Role: securedPartyRole, Addresses: []string{s.securedParty}})
			}
			s.setupRegistry(key, roles)

			// Delegate every party's authority to the executor so nothing is signed directly.
			for _, p := range approvers {
				s.grantProposeAuthz(executor, p)
				s.grantApproveAuthz(executor, p)
			}

			// The first party opens the change via authz; the rest approve via authz.
			changeID := s.execProposeNewControllerViaAuthz(executor, approvers[0], key)
			for _, p := range approvers[1:] {
				s.execApproveViaAuthz(executor, p, changeID)
			}

			s.Require().NotContains(s.roleAddresses(key, controllerRole), s.newController,
				"incomplete authz-delegated approval set must not satisfy the policy")
			pending, err := s.registryKeeper.GetPendingRoleChange(s.ctx, changeID)
			s.Require().NoError(err)
			s.Require().NotNil(pending, "the pending change should remain open while approvals are incomplete")
		})
	}

	// Scenario 1: Controller set, no Secured Party for eNote.
	authzApprovers("no_sp_only_current_controller", false, []string{s.currentController})
	authzApprovers("no_sp_only_new_controller", false, []string{s.newController})

	// Scenario 2: Controller and Secured Party for eNote set.
	authzApprovers("sp_only_current_controller", true, []string{s.currentController})
	authzApprovers("sp_only_new_controller", true, []string{s.newController})
	authzApprovers("sp_only_secured_party", true, []string{s.securedParty})
	authzApprovers("sp_current_controller_and_secured_party", true, []string{s.currentController, s.securedParty})
	authzApprovers("sp_current_controller_and_new_controller", true, []string{s.currentController, s.newController})
	authzApprovers("sp_secured_party_and_new_controller", true, []string{s.securedParty, s.newController})
}

// --- Scenario C: Controller revoke accumulation ---------------------------------------

func (s *RoleChangeAccumulationAcceptanceTestSuite) TestControllerRevoke_Accumulation() {
	controllerRole := types.RegistryRole_REGISTRY_ROLE_CONTROLLER
	securedPartyRole := types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE

	s.Run("single controller, no secured party: applies on proposer approval", func() {
		key := s.mintNFT("acc-c-single")
		s.setupRegistry(key, []types.RolesEntry{
			{Role: controllerRole, Addresses: []string{s.currentController}},
		})

		_, applied := s.proposeRevokeController(key, s.currentController, s.currentController)
		s.Require().True(applied, "the sole current controller's approval satisfies a revoke")
		s.Require().NotContains(s.roleAddresses(key, controllerRole), s.currentController)
	})

	s.Run("secured party set: requires current controller and secured party", func() {
		key := s.mintNFT("acc-c-with-sp")
		s.setupRegistry(key, []types.RolesEntry{
			{Role: controllerRole, Addresses: []string{s.currentController}},
			{Role: securedPartyRole, Addresses: []string{s.securedParty}},
		})

		changeID, applied := s.proposeRevokeController(key, s.currentController, s.currentController)
		s.Require().False(applied, "secured party approval still required for a revoke")
		s.Require().Contains(s.roleAddresses(key, controllerRole), s.currentController)

		applied = s.approve(changeID, s.securedParty)
		s.Require().True(applied, "current controller + secured party satisfy the revoke")
		s.Require().NotContains(s.roleAddresses(key, controllerRole), s.currentController)
	})
}

// --- Invalidation: stale approvals are voided when role holders change underneath ------
//
// The accumulation engine re-resolves roles against live registry state on every approval, so an
// approval recorded by a party that is no longer the current role holder cannot satisfy the policy.
// This is the ticket's invalidation guarantee without a separate expiry mechanism.

func (s *RoleChangeAccumulationAcceptanceTestSuite) TestControllerUpdate_InvalidationOnRoleChange() {
	controllerRole := types.RegistryRole_REGISTRY_ROLE_CONTROLLER
	rotatedController := genAddr()

	key := s.mintNFT("acc-invalidation")
	s.setupRegistry(key, []types.RolesEntry{
		{Role: controllerRole, Addresses: []string{s.currentController}},
	})

	// Current controller proposes; only their approval is recorded so far.
	changeID, applied := s.proposeNewController(key, s.currentController)
	s.Require().False(applied)

	// The controller is rotated out from under the pending change (e.g. via a separate completed
	// flow). The pending change still carries the now-stale currentController approval.
	s.Require().NoError(s.registryKeeper.RevokeRole(s.ctx, key, controllerRole, []string{s.currentController}))
	s.Require().NoError(s.registryKeeper.GrantRole(s.ctx, key, controllerRole, []string{rotatedController}))

	// The new controller approving is no longer enough: the stale approval no longer counts as the
	// current controller, so the policy is not satisfied.
	applied = s.approve(changeID, s.newController)
	s.Require().False(applied, "stale current-controller approval is voided after the role holder changed")
	s.Require().NotContains(s.roleAddresses(key, controllerRole), s.newController)

	// Once the live current controller approves, the policy is satisfied and the change applies.
	applied = s.approve(changeID, rotatedController)
	s.Require().True(applied, "the live current controller's approval re-satisfies the policy")
	s.Require().Contains(s.roleAddresses(key, controllerRole), s.newController)
}

// --- Edge cases -----------------------------------------------------------------------

func (s *RoleChangeAccumulationAcceptanceTestSuite) TestPendingChange_EdgeCases() {
	controllerRole := types.RegistryRole_REGISTRY_ROLE_CONTROLLER
	securedPartyRole := types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE

	s.Run("duplicate approval is idempotent", func() {
		key := s.mintNFT("acc-edge-dup")
		s.setupRegistry(key, []types.RolesEntry{
			{Role: controllerRole, Addresses: []string{s.currentController}},
			{Role: securedPartyRole, Addresses: []string{s.securedParty}},
		})

		changeID, applied := s.proposeNewController(key, s.currentController)
		s.Require().False(applied)

		// The current controller approving again must not double-count or apply the change.
		applied = s.approve(changeID, s.currentController)
		s.Require().False(applied)

		pending, err := s.registryKeeper.GetPendingRoleChange(s.ctx, changeID)
		s.Require().NoError(err)
		s.Require().NotNil(pending)
		s.Require().Equal([]string{s.currentController}, pending.Approvals, "approver recorded exactly once")
	})

	s.Run("approving a non-existent change errors", func() {
		_, err := s.msgServer.ApproveRoleChange(s.ctx, &types.MsgApproveRoleChange{
			Signer:   s.currentController,
			ChangeId: "does-not-exist",
		})
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "pending role change")
	})

	s.Run("ineligible proposer cannot open a pending change", func() {
		key := s.mintNFT("acc-edge-ineligible-proposer")
		s.setupRegistry(key, []types.RolesEntry{
			{Role: controllerRole, Addresses: []string{s.currentController}},
			{Role: securedPartyRole, Addresses: []string{s.securedParty}},
		})

		// The stranger is not a required party for the controller policy, so it must not be able
		// to open (and persist) a new pending change.
		resp, err := s.msgServer.ProposeRoleChange(s.ctx, &types.MsgProposeRoleChange{
			Signer: s.stranger,
			Key:    key,
			RoleUpdates: []types.RoleUpdate{{
				Role:      controllerRole,
				Addresses: []string{s.newController},
			}},
		})
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "not eligible")
		s.Require().Nil(resp)

		// No pending change should have been persisted for this key.
		changes, _, err := s.registryKeeper.GetPendingRoleChanges(s.ctx, nil, key)
		s.Require().NoError(err)
		s.Require().Empty(changes, "no pending change is created for an ineligible proposer")
	})

	s.Run("registry removed underneath cleans up the pending change", func() {
		key := s.mintNFT("acc-edge-removed")
		s.setupRegistry(key, []types.RolesEntry{
			{Role: controllerRole, Addresses: []string{s.currentController}},
			{Role: securedPartyRole, Addresses: []string{s.securedParty}},
		})

		changeID, applied := s.proposeNewController(key, s.currentController)
		s.Require().False(applied)

		s.Require().NoError(s.registryKeeper.DeleteRegistry(s.ctx, key))

		_, err := s.msgServer.ApproveRoleChange(s.ctx, &types.MsgApproveRoleChange{
			Signer:   s.securedParty,
			ChangeId: changeID,
		})
		s.Require().Error(err)

		pending, err := s.registryKeeper.GetPendingRoleChange(s.ctx, changeID)
		s.Require().NoError(err)
		s.Require().Nil(pending, "the orphaned pending change is removed")
	})
}

// proposeBatch opens a pending change for an arbitrary batch of desired-state role updates, signed
// by proposer, and returns the change id and whether it applied immediately.
func (s *RoleChangeAccumulationAcceptanceTestSuite) proposeBatch(key *types.RegistryKey, proposer string, updates []types.RoleUpdate) (string, bool) {
	resp, err := s.msgServer.ProposeRoleChange(s.ctx, &types.MsgProposeRoleChange{
		Signer:      proposer,
		Key:         key,
		RoleUpdates: updates,
	})
	s.Require().NoError(err)
	return resp.ChangeId, resp.Applied
}

// --- Atomic grant+revoke: controller rotation in a single pending change ---------------
//
// A controller rotation (remove the old controller, add the new one) is expressed as a single
// desired-state role update. The grant and revoke are inseparable: the whole change applies
// atomically once the role's policy is satisfied, never leaving the role half-updated.

func (s *RoleChangeAccumulationAcceptanceTestSuite) TestControllerRotation_AtomicGrantRevoke() {
	controllerRole := types.RegistryRole_REGISTRY_ROLE_CONTROLLER
	securedPartyRole := types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE

	key := s.mintNFT("acc-rotation")
	s.setupRegistry(key, []types.RolesEntry{
		{Role: controllerRole, Addresses: []string{s.currentController}},
		{Role: securedPartyRole, Addresses: []string{s.securedParty}},
	})

	// Desired state replaces the controller entirely: drops currentController, adds newController.
	rotation := []types.RoleUpdate{{Role: controllerRole, Addresses: []string{s.newController}}}

	changeID, applied := s.proposeBatch(key, s.currentController, rotation)
	s.Require().False(applied)

	applied = s.approve(changeID, s.securedParty)
	s.Require().False(applied, "still missing the new controller's approval")

	// Nothing is half-applied while pending: the old controller is still present.
	s.Require().Contains(s.roleAddresses(key, controllerRole), s.currentController)
	s.Require().NotContains(s.roleAddresses(key, controllerRole), s.newController)

	applied = s.approve(changeID, s.newController)
	s.Require().True(applied, "current controller + secured party + new controller satisfy the rotation")

	// The grant and revoke applied together: role is now exactly the new controller.
	s.Require().Equal([]string{s.newController}, s.roleAddresses(key, controllerRole))
}

// --- Atomic batch across roles: policy role + NFT-owner-fallback role ------------------
//
// A single pending change can carry updates to several roles. It applies only when every affected
// role's gate is satisfied (policy for governed roles, NFT ownership for the rest), and all updates
// apply together.

func (s *RoleChangeAccumulationAcceptanceTestSuite) TestBatchRoleUpdate_AcrossRoles_AtomicApply() {
	controllerRole := types.RegistryRole_REGISTRY_ROLE_CONTROLLER
	servicerRole := types.RegistryRole_REGISTRY_ROLE_SERVICER
	oldServicer := genAddr()
	newServicer := genAddr()

	key := s.mintNFT("acc-batch-roles")
	s.setupRegistry(key, []types.RolesEntry{
		{Role: controllerRole, Addresses: []string{s.currentController}},
		{Role: servicerRole, Addresses: []string{oldServicer}},
	})

	// One batch: rotate the controller (policy-governed) and replace the servicer (no policy, so it
	// falls back to NFT ownership).
	batch := []types.RoleUpdate{
		{Role: controllerRole, Addresses: []string{s.newController}},
		{Role: servicerRole, Addresses: []string{newServicer}},
	}

	changeID, applied := s.proposeBatch(key, s.currentController, batch)
	s.Require().False(applied)

	// Both controller approvals are present (no secured party set), but the servicer update still
	// needs the NFT owner's approval, so the atomic batch must NOT apply.
	applied = s.approve(changeID, s.newController)
	s.Require().False(applied, "servicer update still requires the NFT owner's approval")
	s.Require().Equal([]string{s.currentController}, s.roleAddresses(key, controllerRole))
	s.Require().Equal([]string{oldServicer}, s.roleAddresses(key, servicerRole))

	// The NFT owner approves, satisfying the servicer (owner-fallback) gate. Now every gate is met
	// and the whole batch applies atomically.
	applied = s.approve(changeID, s.nftOwner)
	s.Require().True(applied, "all gates satisfied: controller policy and servicer owner-fallback")
	s.Require().Equal([]string{s.newController}, s.roleAddresses(key, controllerRole))
	s.Require().Equal([]string{newServicer}, s.roleAddresses(key, servicerRole))
}
