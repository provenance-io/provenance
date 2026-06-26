package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/x/nft"
	nftkeeper "cosmossdk.io/x/nft/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/registry/keeper"
	"github.com/provenance-io/provenance/x/registry/types"
)

// RoleEventAndQueryAcceptanceTestSuite covers Phase C4 (the comprehensive EventRoleUpdated, with
// registry_class_id / previous_addresses / signers) and Phase C5 (the read-only ValidateRoleChange
// dry-run query).
type RoleEventAndQueryAcceptanceTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	registryKeeper keeper.Keeper
	nftKeeper      nftkeeper.Keeper
	msgServer      types.MsgServer
	queryServer    *keeper.QueryServer

	nftClass nft.Class

	nftOwner          string
	currentController string
	newController     string
	securedParty      string
	stranger          string
}

func TestRoleEventAndQueryAcceptanceTestSuite(t *testing.T) {
	suite.Run(t, new(RoleEventAndQueryAcceptanceTestSuite))
}

func (s *RoleEventAndQueryAcceptanceTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: time.Now()}).WithBlockTime(time.Now())

	s.nftKeeper = s.app.NFTKeeper
	s.registryKeeper = s.app.RegistryKeeper
	s.msgServer = keeper.NewMsgServer(s.registryKeeper)
	s.queryServer = keeper.NewQueryServer(s.registryKeeper)

	s.nftOwner = genAddr()
	s.currentController = genAddr()
	s.newController = genAddr()
	s.securedParty = genAddr()
	s.stranger = genAddr()

	s.nftClass = nft.Class{Id: "event-query-test-nft-class"}
	s.nftKeeper.SaveClass(s.ctx, s.nftClass)

	// Install the CONTROLLER policy as the module default so classless entries are policy-governed.
	s.Require().NoError(s.registryKeeper.SetParams(s.ctx, types.Params{
		RoleAuthorizations: types.ControllerRoleAuthorizations(),
	}))
}

func (s *RoleEventAndQueryAcceptanceTestSuite) mintNFT(id string) *types.RegistryKey {
	n := nft.NFT{ClassId: s.nftClass.Id, Id: id}
	ownerAddr, err := sdk.AccAddressFromBech32(s.nftOwner)
	s.Require().NoError(err)
	s.Require().NoError(s.nftKeeper.Mint(s.ctx, n, ownerAddr))
	return &types.RegistryKey{AssetClassId: s.nftClass.Id, NftId: id}
}

// resetEvents installs a fresh event manager so each scenario inspects only its own events.
func (s *RoleEventAndQueryAcceptanceTestSuite) resetEvents() {
	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
}

// roleUpdatedEvents returns all EventRoleUpdated emitted since the last resetEvents.
func (s *RoleEventAndQueryAcceptanceTestSuite) roleUpdatedEvents() []*types.EventRoleUpdated {
	var out []*types.EventRoleUpdated
	for _, e := range s.ctx.EventManager().Events() {
		if e.Type != "provenance.registry.v1.EventRoleUpdated" {
			continue
		}
		msg, err := sdk.ParseTypedEvent(abci.Event{Type: e.Type, Attributes: e.Attributes})
		s.Require().NoError(err)
		ev, ok := msg.(*types.EventRoleUpdated)
		s.Require().True(ok)
		out = append(out, ev)
	}
	return out
}

// --- Phase C4: EventRoleUpdated -----------------------------------------------------------------

// TestEventRoleUpdated_LegacyPath: a role change authorized by the NFT owner (no policy for the
// role) emits EventRoleUpdated with the NFT-owner signer and an empty registry_class_id.
func (s *RoleEventAndQueryAcceptanceTestSuite) TestEventRoleUpdated_LegacyPath() {
	key := s.mintNFT("evt-legacy")
	s.Require().NoError(s.registryKeeper.CreateRegistry(s.ctx, key, nil, ""))
	grantee := genAddr()

	s.resetEvents()
	// SERVICER has no policy (only CONTROLLER does), so this uses the NFT-owner fallback.
	_, err := s.msgServer.GrantRole(s.ctx, &types.MsgGrantRole{
		Signer:    s.nftOwner,
		Key:       key,
		Role:      types.RegistryRole_REGISTRY_ROLE_SERVICER,
		Addresses: []string{grantee},
	})
	s.Require().NoError(err)

	evts := s.roleUpdatedEvents()
	s.Require().Len(evts, 1)
	ev := evts[0]
	s.Require().Equal(key.NftId, ev.NftId)
	s.Require().Equal(key.AssetClassId, ev.AssetClassId)
	s.Require().Equal("", ev.RegistryClassId)
	s.Require().Equal("SERVICER", ev.Role)
	s.Require().Equal([]string{grantee}, ev.Addresses)
	s.Require().Empty(ev.PreviousAddresses)
	s.Require().Len(ev.Signers, 1)
	s.Require().Equal("NFT_ROLE_NFT_OWNER", ev.Signers[0].Role)
	s.Require().Equal("ASSIGNMENT_CURRENT", ev.Signers[0].Assignment)
	s.Require().Equal([]string{s.nftOwner}, ev.Signers[0].Addresses)
}

// TestEventRoleUpdated_PolicyAccumulation: a CONTROLLER change applied via the propose/approve flow
// emits EventRoleUpdated whose signers describe the satisfying policy path (current + secured party
// + new controller) and whose previous/current addresses capture the transition.
func (s *RoleEventAndQueryAcceptanceTestSuite) TestEventRoleUpdated_PolicyAccumulation() {
	key := s.mintNFT("evt-policy")
	s.Require().NoError(s.registryKeeper.CreateRegistry(s.ctx, key, []types.RolesEntry{
		{Role: types.RegistryRole_REGISTRY_ROLE_CONTROLLER, Addresses: []string{s.currentController}},
		{Role: types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE, Addresses: []string{s.securedParty}},
	}, ""))

	resp, err := s.msgServer.ProposeRoleChange(s.ctx, &types.MsgProposeRoleChange{
		Signer: s.currentController,
		Key:    key,
		RoleUpdates: []types.RoleUpdate{{
			Role:      types.RegistryRole_REGISTRY_ROLE_CONTROLLER,
			Addresses: []string{s.newController},
		}},
	})
	s.Require().NoError(err)
	s.Require().False(resp.Applied)

	_, err = s.msgServer.ApproveRoleChange(s.ctx, &types.MsgApproveRoleChange{Signer: s.securedParty, ChangeId: resp.ChangeId})
	s.Require().NoError(err)

	// The applying approval emits the EventRoleUpdated; isolate it.
	s.resetEvents()
	applyResp, err := s.msgServer.ApproveRoleChange(s.ctx, &types.MsgApproveRoleChange{Signer: s.newController, ChangeId: resp.ChangeId})
	s.Require().NoError(err)
	s.Require().True(applyResp.Applied)

	evts := s.roleUpdatedEvents()
	s.Require().Len(evts, 1)
	ev := evts[0]
	s.Require().Equal("CONTROLLER", ev.Role)
	s.Require().Equal([]string{s.newController}, ev.Addresses)
	s.Require().Equal([]string{s.currentController}, ev.PreviousAddresses)

	// Signers: current controller (CURRENT), secured party for eNote (CURRENT), new controller (NEW).
	got := map[string][]string{}
	for _, sgn := range ev.Signers {
		got[sgn.Role+"|"+sgn.Assignment] = sgn.Addresses
	}
	s.Require().Equal([]string{s.currentController}, got["REGISTRY_ROLE_CONTROLLER|ASSIGNMENT_CURRENT"])
	s.Require().Equal([]string{s.securedParty}, got["REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE|ASSIGNMENT_CURRENT"])
	s.Require().Equal([]string{s.newController}, got["REGISTRY_ROLE_CONTROLLER|ASSIGNMENT_NEW"])
}

// TestEventRoleUpdated_Revoke: a revoke (removing an address from a role) emits EventRoleUpdated
// with the resulting (empty) address set and the prior addresses, authorized via the NFT-owner
// fallback for a role with no policy.
func (s *RoleEventAndQueryAcceptanceTestSuite) TestEventRoleUpdated_Revoke() {
	key := s.mintNFT("evt-revoke")
	servicer := genAddr()
	s.Require().NoError(s.registryKeeper.CreateRegistry(s.ctx, key, []types.RolesEntry{
		{Role: types.RegistryRole_REGISTRY_ROLE_SERVICER, Addresses: []string{servicer}},
	}, ""))

	s.resetEvents()
	// SERVICER has no policy, so the NFT owner authorizes the revoke.
	_, err := s.msgServer.RevokeRole(s.ctx, &types.MsgRevokeRole{
		Signer:    s.nftOwner,
		Key:       key,
		Role:      types.RegistryRole_REGISTRY_ROLE_SERVICER,
		Addresses: []string{servicer},
	})
	s.Require().NoError(err)

	evts := s.roleUpdatedEvents()
	s.Require().Len(evts, 1)
	ev := evts[0]
	s.Require().Equal("SERVICER", ev.Role)
	s.Require().Empty(ev.Addresses)
	s.Require().Equal([]string{servicer}, ev.PreviousAddresses)
	s.Require().Len(ev.Signers, 1)
	s.Require().Equal("NFT_ROLE_NFT_OWNER", ev.Signers[0].Role)
	s.Require().Equal([]string{s.nftOwner}, ev.Signers[0].Addresses)
}

// TestEventRoleUpdated_RegistryClassId: an entry registered under a class carries that class id on
// the event.
func (s *RoleEventAndQueryAcceptanceTestSuite) TestEventRoleUpdated_RegistryClassId() {
	maintainer := genAddr()
	_, err := s.msgServer.CreateRegistryClass(s.ctx, &types.MsgCreateRegistryClass{
		Signer:             maintainer,
		RegistryClassId:    "evt-class",
		AssetClassId:       s.nftClass.Id,
		Maintainer:         maintainer,
		RoleAuthorizations: []types.RoleAuthorization{servicerRoleAuthorization()},
	})
	s.Require().NoError(err)

	key := s.mintNFT("evt-classed")
	s.Require().NoError(s.registryKeeper.CreateRegistry(s.ctx, key, []types.RolesEntry{
		{Role: types.RegistryRole_REGISTRY_ROLE_CONTROLLER, Addresses: []string{s.currentController}},
	}, "evt-class"))

	grantee := genAddr()
	s.resetEvents()
	// The class policy lets the current controller manage servicers single-handedly.
	_, err = s.msgServer.GrantRole(s.ctx, &types.MsgGrantRole{
		Signer:    s.currentController,
		Key:       key,
		Role:      types.RegistryRole_REGISTRY_ROLE_SERVICER,
		Addresses: []string{grantee},
	})
	s.Require().NoError(err)

	evts := s.roleUpdatedEvents()
	s.Require().Len(evts, 1)
	ev := evts[0]
	s.Require().Equal("evt-class", ev.RegistryClassId)
	s.Require().Equal("SERVICER", ev.Role)
	s.Require().Empty(ev.PreviousAddresses)
	s.Require().Equal([]string{grantee}, ev.Addresses)
	s.Require().Len(ev.Signers, 1)
	s.Require().Equal("REGISTRY_ROLE_CONTROLLER", ev.Signers[0].Role)
	s.Require().Equal("ASSIGNMENT_CURRENT", ev.Signers[0].Assignment)
	s.Require().Equal([]string{s.currentController}, ev.Signers[0].Addresses)
}

// TestSetRoles_AdditionsOnlyForNewAssignment: MsgSetRoles takes the full desired address set, but
// the authorization engine's ASSIGNMENT_NEW must resolve only the newly-added addresses. A policy
// that requires the incoming (NEW) member to sign must be satisfiable single-signer by that member
// even when the desired set also contains already-current members.
func (s *RoleEventAndQueryAcceptanceTestSuite) TestSetRoles_AdditionsOnlyForNewAssignment() {
	// Policy: a SERVICER may be added if the newly-assigned servicer signs (SERVICER@NEW).
	selfAddServicer := types.RoleAuthorization{
		Role: types.RegistryRole_REGISTRY_ROLE_SERVICER,
		Authorizations: []*types.Authorization{{
			Description: "a new servicer may add themselves",
			Signatures: []*types.SignatureRequirement{{
				Type: types.SignatureType_SIGNATURE_TYPE_REQUIRED_ALL,
				Roles: []*types.RoleAssignment{{
					RoleSelector: &types.RoleAssignment_RegistryRole{
						RegistryRole: types.RegistryRole_REGISTRY_ROLE_SERVICER,
					},
					Assignment: types.Assignment_ASSIGNMENT_NEW,
				}},
			}},
		}},
	}
	maintainer := genAddr()
	_, err := s.msgServer.CreateRegistryClass(s.ctx, &types.MsgCreateRegistryClass{
		Signer:             maintainer,
		RegistryClassId:    "servicer-class",
		AssetClassId:       s.nftClass.Id,
		Maintainer:         maintainer,
		RoleAuthorizations: []types.RoleAuthorization{selfAddServicer},
	})
	s.Require().NoError(err)

	existingServicer := genAddr()
	newServicer := genAddr()
	key := s.mintNFT("evt-additions")
	s.Require().NoError(s.registryKeeper.CreateRegistry(s.ctx, key, []types.RolesEntry{
		{Role: types.RegistryRole_REGISTRY_ROLE_SERVICER, Addresses: []string{existingServicer}},
	}, "servicer-class"))

	s.resetEvents()
	// Desired state is [existing, new]; only the addition (newServicer) needs to sign. Before the
	// additions fix this errored because ASSIGNMENT_NEW resolved to both addresses.
	_, err = s.msgServer.SetRoles(s.ctx, &types.MsgSetRoles{
		Signer: newServicer,
		Key:    key,
		RoleUpdates: []types.RoleUpdate{{
			Role:      types.RegistryRole_REGISTRY_ROLE_SERVICER,
			Addresses: []string{existingServicer, newServicer},
		}},
	})
	s.Require().NoError(err)

	evts := s.roleUpdatedEvents()
	s.Require().Len(evts, 1)
	ev := evts[0]
	s.Require().Equal("SERVICER", ev.Role)
	s.Require().Equal([]string{existingServicer, newServicer}, ev.Addresses)
	s.Require().Equal([]string{existingServicer}, ev.PreviousAddresses)
	s.Require().Len(ev.Signers, 1)
	s.Require().Equal("REGISTRY_ROLE_SERVICER", ev.Signers[0].Role)
	s.Require().Equal("ASSIGNMENT_NEW", ev.Signers[0].Assignment)
	s.Require().Equal([]string{newServicer}, ev.Signers[0].Addresses)
}

// --- Phase C5: ValidateRoleChange query ---------------------------------------------------------

func (s *RoleEventAndQueryAcceptanceTestSuite) controllerUpdate() []types.RoleUpdate {
	return []types.RoleUpdate{{
		Role:      types.RegistryRole_REGISTRY_ROLE_CONTROLLER,
		Addresses: []string{s.newController},
	}}
}

func (s *RoleEventAndQueryAcceptanceTestSuite) TestValidateRoleChange_Authorized() {
	key := s.mintNFT("vrc-ok")
	s.Require().NoError(s.registryKeeper.CreateRegistry(s.ctx, key, []types.RolesEntry{
		{Role: types.RegistryRole_REGISTRY_ROLE_CONTROLLER, Addresses: []string{s.currentController}},
	}, ""))

	res, err := s.queryServer.ValidateRoleChange(s.ctx, &types.QueryValidateRoleChangeRequest{
		Key:         key,
		RoleUpdates: s.controllerUpdate(),
		Approvers:   []string{s.currentController, s.newController},
	})
	s.Require().NoError(err)
	s.Require().True(res.Authorized)
	s.Require().Empty(res.Error)
}

func (s *RoleEventAndQueryAcceptanceTestSuite) TestValidateRoleChange_MissingApprover() {
	key := s.mintNFT("vrc-missing")
	s.Require().NoError(s.registryKeeper.CreateRegistry(s.ctx, key, []types.RolesEntry{
		{Role: types.RegistryRole_REGISTRY_ROLE_CONTROLLER, Addresses: []string{s.currentController}},
	}, ""))

	// Only the current controller approves; the new controller's approval is still outstanding.
	res, err := s.queryServer.ValidateRoleChange(s.ctx, &types.QueryValidateRoleChangeRequest{
		Key:         key,
		RoleUpdates: s.controllerUpdate(),
		Approvers:   []string{s.currentController},
	})
	s.Require().NoError(err)
	s.Require().False(res.Authorized)
	s.Require().NotEmpty(res.Error)
	s.Require().Contains(res.Error, "CONTROLLER")
}

func (s *RoleEventAndQueryAcceptanceTestSuite) TestValidateRoleChange_LegacyFallback() {
	key := s.mintNFT("vrc-legacy")
	s.Require().NoError(s.registryKeeper.CreateRegistry(s.ctx, key, nil, ""))

	update := []types.RoleUpdate{{
		Role:      types.RegistryRole_REGISTRY_ROLE_SERVICER,
		Addresses: []string{genAddr()},
	}}

	s.Run("NFT owner approves", func() {
		res, err := s.queryServer.ValidateRoleChange(s.ctx, &types.QueryValidateRoleChangeRequest{
			Key:         key,
			RoleUpdates: update,
			Approvers:   []string{s.nftOwner},
		})
		s.Require().NoError(err)
		s.Require().True(res.Authorized)
		s.Require().Empty(res.Error)
	})

	s.Run("non-owner is rejected", func() {
		res, err := s.queryServer.ValidateRoleChange(s.ctx, &types.QueryValidateRoleChangeRequest{
			Key:         key,
			RoleUpdates: update,
			Approvers:   []string{s.stranger},
		})
		s.Require().NoError(err)
		s.Require().False(res.Authorized)
		s.Require().NotEmpty(res.Error)
	})
}

func (s *RoleEventAndQueryAcceptanceTestSuite) TestValidateRoleChange_InvalidRequest() {
	s.Run("nil request", func() {
		_, err := s.queryServer.ValidateRoleChange(s.ctx, nil)
		s.Require().Error(err)
	})

	s.Run("no role updates", func() {
		key := s.mintNFT("vrc-empty")
		s.Require().NoError(s.registryKeeper.CreateRegistry(s.ctx, key, nil, ""))
		_, err := s.queryServer.ValidateRoleChange(s.ctx, &types.QueryValidateRoleChangeRequest{Key: key})
		s.Require().Error(err)
	})
}
