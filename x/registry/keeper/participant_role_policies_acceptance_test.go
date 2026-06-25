package keeper_test

import (
	"os"
	"path/filepath"
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

// ParticipantRolePoliciesAcceptanceTestSuite proves that the participant role policies described in
// the requirements (§"Participant Roles") are expressible as ordinary
// RegistryClass.role_authorizations data and are correctly evaluated by the policy engine — without
// any hard-coded, per-role chain logic. The same policies double as the example fixture shipped in
// x/registry/spec/examples/example_registry_class.json.
type ParticipantRolePoliciesAcceptanceTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	registryKeeper keeper.Keeper
	nftKeeper      nftkeeper.Keeper
	msgServer      types.MsgServer

	nftClass nft.Class

	maintainer string
	nftOwner   string
}

func TestParticipantRolePoliciesAcceptanceTestSuite(t *testing.T) {
	suite.Run(t, new(ParticipantRolePoliciesAcceptanceTestSuite))
}

const exampleClassID = "loan-registry-v1"

func (s *ParticipantRolePoliciesAcceptanceTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: time.Now()}).WithBlockTime(time.Now())

	s.nftKeeper = s.app.NFTKeeper
	s.registryKeeper = s.app.RegistryKeeper
	s.msgServer = keeper.NewMsgServer(s.registryKeeper)

	s.maintainer = genAddr()
	s.nftOwner = genAddr()

	s.nftClass = nft.Class{Id: "participant-roles-test-nft-class"}
	s.nftKeeper.SaveClass(s.ctx, s.nftClass)

	// Install the full participant registry class so every classed entry is governed by these policies.
	_, err := s.msgServer.CreateRegistryClass(s.ctx, &types.MsgCreateRegistryClass{
		Signer:             s.maintainer,
		RegistryClassId:    exampleClassID,
		AssetClassId:       s.nftClass.Id,
		Maintainer:         s.maintainer,
		RoleAuthorizations: participantRoleAuthorizations(),
	})
	s.Require().NoError(err)
}

// registerWithRoles mints an NFT owned by s.nftOwner, registers it under the participant class, and seeds
// the given initial roles. It returns the registry key.
func (s *ParticipantRolePoliciesAcceptanceTestSuite) registerWithRoles(id string, roles []types.RolesEntry) *types.RegistryKey {
	n := nft.NFT{ClassId: s.nftClass.Id, Id: id}
	ownerAddr, err := sdk.AccAddressFromBech32(s.nftOwner)
	s.Require().NoError(err)
	s.Require().NoError(s.nftKeeper.Mint(s.ctx, n, ownerAddr))

	key := &types.RegistryKey{AssetClassId: s.nftClass.Id, NftId: id}
	s.Require().NoError(s.registryKeeper.CreateRegistry(s.ctx, key, roles, exampleClassID))
	return key
}

// propose opens a pending role change and returns the change id and whether it applied immediately.
func (s *ParticipantRolePoliciesAcceptanceTestSuite) propose(key *types.RegistryKey, signer string, role types.RegistryRole, addrs []string) (string, bool) {
	resp, err := s.msgServer.ProposeRoleChange(s.ctx, &types.MsgProposeRoleChange{
		Signer:      signer,
		Key:         key,
		RoleUpdates: []types.RoleUpdate{{Role: role, Addresses: addrs}},
	})
	s.Require().NoError(err)
	return resp.ChangeId, resp.Applied
}

// proposeErr attempts a proposal expected to be rejected because the proposer is not an eligible
// approver of the change, and returns the error.
func (s *ParticipantRolePoliciesAcceptanceTestSuite) proposeErr(key *types.RegistryKey, signer string, role types.RegistryRole, addrs []string) error {
	_, err := s.msgServer.ProposeRoleChange(s.ctx, &types.MsgProposeRoleChange{
		Signer:      signer,
		Key:         key,
		RoleUpdates: []types.RoleUpdate{{Role: role, Addresses: addrs}},
	})
	return err
}

// approve records an approval for a pending change and returns whether it applied.
func (s *ParticipantRolePoliciesAcceptanceTestSuite) approve(changeID, signer string) bool {
	resp, err := s.msgServer.ApproveRoleChange(s.ctx, &types.MsgApproveRoleChange{
		Signer:   signer,
		ChangeId: changeID,
	})
	s.Require().NoError(err)
	return resp.Applied
}

// roleAddresses returns the addresses currently assigned to role on the entry.
func (s *ParticipantRolePoliciesAcceptanceTestSuite) roleAddresses(key *types.RegistryKey, role types.RegistryRole) []string {
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

// --- Consent-matrix tests -----------------------------------------------------------------------

// TestOriginator: an originator update requires the current originator (or NFT owner if unset) plus
// the incoming new originator.
func (s *ParticipantRolePoliciesAcceptanceTestSuite) TestOriginator() {
	role := types.RegistryRole_REGISTRY_ROLE_ORIGINATOR
	currentOriginator := genAddr()
	newOriginator := genAddr()

	s.Run("current + new originator approve", func() {
		key := s.registerWithRoles("orig-happy", []types.RolesEntry{{Role: role, Addresses: []string{currentOriginator}}})

		changeID, applied := s.propose(key, currentOriginator, role, []string{newOriginator})
		s.Require().False(applied, "current originator alone is not sufficient")

		applied = s.approve(changeID, newOriginator)
		s.Require().True(applied, "current + new originator satisfy the policy")
		s.Require().Equal([]string{newOriginator}, s.roleAddresses(key, role))
	})

	s.Run("NFT owner acts as current originator when none is set", func() {
		key := s.registerWithRoles("orig-nft-owner", nil)

		changeID, applied := s.propose(key, s.nftOwner, role, []string{newOriginator})
		s.Require().False(applied)

		applied = s.approve(changeID, newOriginator)
		s.Require().True(applied)
		s.Require().Equal([]string{newOriginator}, s.roleAddresses(key, role))
	})

	s.Run("stranger cannot originate", func() {
		key := s.registerWithRoles("orig-stranger", []types.RolesEntry{{Role: role, Addresses: []string{currentOriginator}}})

		err := s.proposeErr(key, genAddr(), role, []string{newOriginator})
		s.Require().Error(err, "an ineligible proposer is rejected outright")
		s.Require().Equal([]string{currentOriginator}, s.roleAddresses(key, role))
	})
}

// TestLienOwnerStandard: a standard lien owner transfer requires the current lien owner, the current
// Secured Party for Lien (if set), and the new lien owner.
func (s *ParticipantRolePoliciesAcceptanceTestSuite) TestLienOwnerStandard() {
	lienOwner := types.RegistryRole_REGISTRY_ROLE_LIEN_OWNER
	securedParty := types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_LIEN

	currentLienOwner := genAddr()
	sp := genAddr()
	newLienOwner := genAddr()

	baseRoles := func() []types.RolesEntry {
		return []types.RolesEntry{
			{Role: lienOwner, Addresses: []string{currentLienOwner}},
			{Role: securedParty, Addresses: []string{sp}},
		}
	}

	s.Run("current lien owner + secured party + new lien owner", func() {
		key := s.registerWithRoles("lien-happy", baseRoles())

		changeID, applied := s.propose(key, currentLienOwner, lienOwner, []string{newLienOwner})
		s.Require().False(applied)

		applied = s.approve(changeID, sp)
		s.Require().False(applied, "still missing the new lien owner")

		applied = s.approve(changeID, newLienOwner)
		s.Require().True(applied)
		s.Require().Equal([]string{newLienOwner}, s.roleAddresses(key, lienOwner))
	})

	s.Run("incomplete: missing secured party approval", func() {
		key := s.registerWithRoles("lien-no-sp", baseRoles())

		changeID, _ := s.propose(key, currentLienOwner, lienOwner, []string{newLienOwner})
		applied := s.approve(changeID, newLienOwner)
		s.Require().False(applied, "secured party for lien must also approve")
		s.Require().Equal([]string{currentLienOwner}, s.roleAddresses(key, lienOwner))
	})
}

// TestLienOwnerForeclosure: the Secured Party for Lien can unilaterally become the Lien Owner. A
// single proposal by the secured party (assigning the lien owner role to itself) satisfies the
// dedicated foreclosure authorization path.
func (s *ParticipantRolePoliciesAcceptanceTestSuite) TestLienOwnerForeclosure() {
	lienOwner := types.RegistryRole_REGISTRY_ROLE_LIEN_OWNER
	securedParty := types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_LIEN

	currentLienOwner := genAddr()
	sp := genAddr()

	s.Run("secured party unilaterally forecloses into lien owner", func() {
		key := s.registerWithRoles("lien-foreclose", []types.RolesEntry{
			{Role: lienOwner, Addresses: []string{currentLienOwner}},
			{Role: securedParty, Addresses: []string{sp}},
		})

		_, applied := s.propose(key, sp, lienOwner, []string{sp})
		s.Require().True(applied, "foreclosure path: current secured party becomes the new lien owner")
		s.Require().Equal([]string{sp}, s.roleAddresses(key, lienOwner))
	})

	s.Run("stranger cannot foreclose", func() {
		key := s.registerWithRoles("lien-foreclose-bad", []types.RolesEntry{
			{Role: lienOwner, Addresses: []string{currentLienOwner}},
			{Role: securedParty, Addresses: []string{sp}},
		})

		stranger := genAddr()
		_, applied := s.propose(key, stranger, lienOwner, []string{stranger})
		s.Require().False(applied)
		s.Require().Equal([]string{currentLienOwner}, s.roleAddresses(key, lienOwner))
	})
}

// TestControllerForeclosure: the Secured Party for eNote can unilaterally become the Controller.
func (s *ParticipantRolePoliciesAcceptanceTestSuite) TestControllerForeclosure() {
	controller := types.RegistryRole_REGISTRY_ROLE_CONTROLLER
	securedParty := types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE

	currentController := genAddr()
	sp := genAddr()

	key := s.registerWithRoles("ctrl-foreclose", []types.RolesEntry{
		{Role: controller, Addresses: []string{currentController}},
		{Role: securedParty, Addresses: []string{sp}},
	})

	_, applied := s.propose(key, sp, controller, []string{sp})
	s.Require().True(applied, "foreclosure path: current secured party for eNote becomes the controller")
	s.Require().Equal([]string{sp}, s.roleAddresses(key, controller))
}

// TestServicer exercises a policy whose conditional requirement uses a role_priority selector
// (Secured Party for eNote, falling back to Pledgee). Granting a servicer requires the current
// controller, the conditional approver, and the new servicer.
func (s *ParticipantRolePoliciesAcceptanceTestSuite) TestServicer() {
	servicer := types.RegistryRole_REGISTRY_ROLE_SERVICER
	controller := types.RegistryRole_REGISTRY_ROLE_CONTROLLER
	spEnote := types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE

	currentController := genAddr()
	sp := genAddr()
	newServicer := genAddr()

	s.Run("controller + secured party for eNote + new servicer", func() {
		key := s.registerWithRoles("svc-happy", []types.RolesEntry{
			{Role: controller, Addresses: []string{currentController}},
			{Role: spEnote, Addresses: []string{sp}},
		})

		changeID, applied := s.propose(key, currentController, servicer, []string{newServicer})
		s.Require().False(applied)

		applied = s.approve(changeID, sp)
		s.Require().False(applied, "still missing the new servicer")

		applied = s.approve(changeID, newServicer)
		s.Require().True(applied)
		s.Require().Equal([]string{newServicer}, s.roleAddresses(key, servicer))
	})

	s.Run("incomplete: missing conditional secured-party approval", func() {
		key := s.registerWithRoles("svc-no-sp", []types.RolesEntry{
			{Role: controller, Addresses: []string{currentController}},
			{Role: spEnote, Addresses: []string{sp}},
		})

		changeID, _ := s.propose(key, currentController, servicer, []string{newServicer})
		applied := s.approve(changeID, newServicer)
		s.Require().False(applied)
		s.Require().Empty(s.roleAddresses(key, servicer))
	})
}

// TestAllParticipantPoliciesCoexist confirms the full set validates and persists together.
func (s *ParticipantRolePoliciesAcceptanceTestSuite) TestAllParticipantPoliciesCoexist() {
	got, err := s.registryKeeper.GetRegistryClass(s.ctx, exampleClassID)
	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Require().Len(got.RoleAuthorizations, len(participantRoleAuthorizations()))
}

// TestMalformedPolicyRejected verifies the deepened create-time validation rejects malformed
// authorization paths instead of letting them fail closed only at evaluation time.
func (s *ParticipantRolePoliciesAcceptanceTestSuite) TestMalformedPolicyRejected() {
	cases := []struct {
		name    string
		auth    types.RoleAuthorization
		errPart string
	}{
		{
			name: "nft_role with NEW assignment",
			auth: types.RoleAuthorization{
				Role: types.RegistryRole_REGISTRY_ROLE_SERVICER,
				Authorizations: []*types.Authorization{{
					Description: "bad",
					Signatures: []*types.SignatureRequirement{
						sigReqAll(raNft(types.NftRole_NFT_ROLE_NFT_OWNER, types.Assignment_ASSIGNMENT_NEW)),
					},
				}},
			},
			errPart: "nft_role may only be used with a CURRENT*",
		},
		{
			name: "missing role selector",
			auth: types.RoleAuthorization{
				Role: types.RegistryRole_REGISTRY_ROLE_SERVICER,
				Authorizations: []*types.Authorization{{
					Description: "bad",
					Signatures: []*types.SignatureRequirement{
						{Type: types.SignatureType_SIGNATURE_TYPE_REQUIRED_ALL, Roles: []*types.RoleAssignment{
							{Assignment: types.Assignment_ASSIGNMENT_CURRENT},
						}},
					},
				}},
			},
			errPart: "a role selector",
		},
		{
			name: "unspecified signature type",
			auth: types.RoleAuthorization{
				Role: types.RegistryRole_REGISTRY_ROLE_SERVICER,
				Authorizations: []*types.Authorization{{
					Description: "bad",
					Signatures: []*types.SignatureRequirement{
						{Roles: []*types.RoleAssignment{raRegistry(types.RegistryRole_REGISTRY_ROLE_CONTROLLER, types.Assignment_ASSIGNMENT_CURRENT)}},
					},
				}},
			},
			errPart: "type",
		},
		{
			name: "authorization path with no signature requirements",
			auth: types.RoleAuthorization{
				Role:           types.RegistryRole_REGISTRY_ROLE_SERVICER,
				Authorizations: []*types.Authorization{{Description: "empty"}},
			},
			errPart: "at least one signature requirement",
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			_, err := s.msgServer.CreateRegistryClass(s.ctx, &types.MsgCreateRegistryClass{
				Signer:             s.maintainer,
				RegistryClassId:    "bad-" + tc.name,
				AssetClassId:       s.nftClass.Id,
				Maintainer:         s.maintainer,
				RoleAuthorizations: []types.RoleAuthorization{tc.auth},
			})
			s.Require().Error(err)
			s.Require().ErrorContains(err, tc.errPart)
		})
	}
}

// TestExampleFixtureInSync verifies the committed example fixture proto-JSON decodes to exactly the
// participant policies built in code. Set REGEN_EXAMPLE_FIXTURE=1 to (re)generate the fixture file.
func (s *ParticipantRolePoliciesAcceptanceTestSuite) TestExampleFixtureInSync() {
	path := filepath.Join("..", "spec", "examples", "example_registry_class.json")

	want := types.RegistryClass{
		RegistryClassId:    exampleClassID,
		AssetClassId:       "loan.asset",
		Maintainer:         "pb1maintainerplaceholder0000000000000000000",
		RoleAuthorizations: participantRoleAuthorizations(),
	}

	if os.Getenv("REGEN_EXAMPLE_FIXTURE") == "1" {
		bz, err := s.app.AppCodec().MarshalJSON(&want)
		s.Require().NoError(err)
		s.Require().NoError(os.WriteFile(path, append(bz, '\n'), 0o644))
	}

	bz, err := os.ReadFile(path)
	s.Require().NoError(err, "example fixture missing; run with REGEN_EXAMPLE_FIXTURE=1 to generate")

	var got types.RegistryClass
	s.Require().NoError(s.app.AppCodec().UnmarshalJSON(bz, &got))
	s.Require().Equal(want.RoleAuthorizations, got.RoleAuthorizations,
		"example fixture is out of sync with code; run with REGEN_EXAMPLE_FIXTURE=1 to regenerate")
}

// --- Participant policy builders (mirror requirements.md §"Participant Roles") ---------------------

// participantRoleAuthorizations returns the full set of participant role policies.
func participantRoleAuthorizations() []types.RoleAuthorization {
	return []types.RoleAuthorization{
		originatorPolicy(),
		lienOwnerPolicy(),
		controllerPolicy(),
		securedPartyForLienPolicy(),
		securedPartyForEnotePolicy(),
		pledgeePolicy(),
		servicerPolicy(),
		subservicerPolicy(),
	}
}

func originatorPolicy() types.RoleAuthorization {
	return types.RoleAuthorization{
		Role: types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
		Authorizations: []*types.Authorization{{
			Description: "Assign originator (requires current and new originator approval)",
			Signatures: []*types.SignatureRequirement{
				sigReqAll(raPriority(types.Assignment_ASSIGNMENT_CURRENT,
					rpRegistry(types.RegistryRole_REGISTRY_ROLE_ORIGINATOR),
					rpNft(types.NftRole_NFT_ROLE_NFT_OWNER),
				)),
				sigReqAllIfSet(raRegistry(types.RegistryRole_REGISTRY_ROLE_ORIGINATOR, types.Assignment_ASSIGNMENT_NEW)),
			},
		}},
	}
}

func lienOwnerPolicy() types.RoleAuthorization {
	return types.RoleAuthorization{
		Role: types.RegistryRole_REGISTRY_ROLE_LIEN_OWNER,
		Authorizations: []*types.Authorization{
			{
				Description: "Transfer requiring current lien owner approval",
				Signatures: []*types.SignatureRequirement{
					sigReqAll(raPriority(types.Assignment_ASSIGNMENT_CURRENT,
						rpRegistry(types.RegistryRole_REGISTRY_ROLE_LIEN_OWNER),
						rpNft(types.NftRole_NFT_ROLE_NFT_OWNER),
					)),
					sigReqAllIfSet(
						raRegistry(types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_LIEN, types.Assignment_ASSIGNMENT_CURRENT),
						raRegistry(types.RegistryRole_REGISTRY_ROLE_LIEN_OWNER, types.Assignment_ASSIGNMENT_NEW),
					),
				},
			},
			{
				Description: "Foreclosure: Secured Party for Lien can unilaterally become Lien Owner in case of default",
				Signatures: []*types.SignatureRequirement{
					sigReqAll(
						raRegistry(types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_LIEN, types.Assignment_ASSIGNMENT_CURRENT),
						raRegistry(types.RegistryRole_REGISTRY_ROLE_LIEN_OWNER, types.Assignment_ASSIGNMENT_NEW),
					),
				},
			},
		},
	}
}

func controllerPolicy() types.RoleAuthorization {
	return types.RoleAuthorization{
		Role: types.RegistryRole_REGISTRY_ROLE_CONTROLLER,
		Authorizations: []*types.Authorization{
			{
				Description: "Transfer requiring current controller approval",
				Signatures: []*types.SignatureRequirement{
					sigReqAll(raPriority(types.Assignment_ASSIGNMENT_CURRENT,
						rpRegistry(types.RegistryRole_REGISTRY_ROLE_CONTROLLER),
						rpNft(types.NftRole_NFT_ROLE_NFT_OWNER),
					)),
					sigReqAllIfSet(
						raRegistry(types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE, types.Assignment_ASSIGNMENT_CURRENT),
						raRegistry(types.RegistryRole_REGISTRY_ROLE_CONTROLLER, types.Assignment_ASSIGNMENT_NEW),
					),
				},
			},
			{
				Description: "Foreclosure: Secured Party for eNote can unilaterally become Controller in case of default",
				Signatures: []*types.SignatureRequirement{
					sigReqAll(
						raRegistry(types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE, types.Assignment_ASSIGNMENT_CURRENT),
						raRegistry(types.RegistryRole_REGISTRY_ROLE_CONTROLLER, types.Assignment_ASSIGNMENT_NEW),
					),
				},
			},
		},
	}
}

func securedPartyForLienPolicy() types.RoleAuthorization {
	return types.RoleAuthorization{
		Role: types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_LIEN,
		Authorizations: []*types.Authorization{
			{
				Description: "Update by current lien owner",
				Signatures: []*types.SignatureRequirement{
					sigReqAll(raPriority(types.Assignment_ASSIGNMENT_CURRENT,
						rpRegistry(types.RegistryRole_REGISTRY_ROLE_LIEN_OWNER),
						rpNft(types.NftRole_NFT_ROLE_NFT_OWNER),
					)),
					sigReqAllIfSet(
						raRegistry(types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_LIEN, types.Assignment_ASSIGNMENT_CURRENT),
						raRegistry(types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_LIEN, types.Assignment_ASSIGNMENT_NEW),
					),
				},
			},
			{
				Description: "Update by current secured party for lien",
				Signatures: []*types.SignatureRequirement{
					sigReqAll(raRegistry(types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_LIEN, types.Assignment_ASSIGNMENT_CURRENT)),
					sigReqAllIfSet(raRegistry(types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_LIEN, types.Assignment_ASSIGNMENT_NEW)),
				},
			},
		},
	}
}

func securedPartyForEnotePolicy() types.RoleAuthorization {
	return types.RoleAuthorization{
		Role: types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE,
		Authorizations: []*types.Authorization{
			{
				Description: "Update by current controller",
				Signatures: []*types.SignatureRequirement{
					sigReqAll(raPriority(types.Assignment_ASSIGNMENT_CURRENT,
						rpRegistry(types.RegistryRole_REGISTRY_ROLE_CONTROLLER),
						rpNft(types.NftRole_NFT_ROLE_NFT_OWNER),
					)),
					sigReqAllIfSet(
						raRegistry(types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE, types.Assignment_ASSIGNMENT_CURRENT),
						raRegistry(types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE, types.Assignment_ASSIGNMENT_NEW),
					),
				},
			},
			{
				Description: "Update by current secured party for eNote",
				Signatures: []*types.SignatureRequirement{
					sigReqAll(raRegistry(types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE, types.Assignment_ASSIGNMENT_CURRENT)),
					sigReqAllIfSet(raRegistry(types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE, types.Assignment_ASSIGNMENT_NEW)),
				},
			},
		},
	}
}

func pledgeePolicy() types.RoleAuthorization {
	return types.RoleAuthorization{
		Role: types.RegistryRole_REGISTRY_ROLE_PLEDGEE,
		Authorizations: []*types.Authorization{{
			Description: "Update pledgee with value owner approval",
			Signatures: []*types.SignatureRequirement{
				sigReqAll(raNft(types.NftRole_NFT_ROLE_NFT_OWNER, types.Assignment_ASSIGNMENT_CURRENT)),
				sigReqAllIfSet(
					raRegistry(types.RegistryRole_REGISTRY_ROLE_PLEDGEE, types.Assignment_ASSIGNMENT_CURRENT),
					raRegistry(types.RegistryRole_REGISTRY_ROLE_PLEDGEE, types.Assignment_ASSIGNMENT_NEW),
				),
			},
		}},
	}
}

func servicerPolicy() types.RoleAuthorization {
	return types.RoleAuthorization{
		Role: types.RegistryRole_REGISTRY_ROLE_SERVICER,
		Authorizations: []*types.Authorization{{
			Description: "Update servicer with owner/controller approval",
			Signatures: []*types.SignatureRequirement{
				sigReqAll(raPriority(types.Assignment_ASSIGNMENT_CURRENT,
					rpRegistry(types.RegistryRole_REGISTRY_ROLE_CONTROLLER),
					rpNft(types.NftRole_NFT_ROLE_NFT_OWNER),
				)),
				sigReqAllIfSet(
					raPriority(types.Assignment_ASSIGNMENT_CURRENT,
						rpRegistry(types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE),
						rpRegistry(types.RegistryRole_REGISTRY_ROLE_PLEDGEE),
					),
					raRegistry(types.RegistryRole_REGISTRY_ROLE_SERVICER, types.Assignment_ASSIGNMENT_CURRENT),
					raRegistry(types.RegistryRole_REGISTRY_ROLE_SERVICER, types.Assignment_ASSIGNMENT_NEW),
				),
			},
		}},
	}
}

func subservicerPolicy() types.RoleAuthorization {
	return types.RoleAuthorization{
		Role: types.RegistryRole_REGISTRY_ROLE_SUBSERVICER,
		Authorizations: []*types.Authorization{{
			Description: "Update sub-servicer with owner/controller and servicer approval",
			Signatures: []*types.SignatureRequirement{
				sigReqAll(raPriority(types.Assignment_ASSIGNMENT_CURRENT,
					rpRegistry(types.RegistryRole_REGISTRY_ROLE_CONTROLLER),
					rpNft(types.NftRole_NFT_ROLE_NFT_OWNER),
				)),
				sigReqAllIfSet(
					raPriority(types.Assignment_ASSIGNMENT_CURRENT,
						rpRegistry(types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE),
						rpRegistry(types.RegistryRole_REGISTRY_ROLE_PLEDGEE),
					),
					raRegistry(types.RegistryRole_REGISTRY_ROLE_SERVICER, types.Assignment_ASSIGNMENT_CURRENT),
					raRegistry(types.RegistryRole_REGISTRY_ROLE_SUBSERVICER, types.Assignment_ASSIGNMENT_CURRENT),
					raRegistry(types.RegistryRole_REGISTRY_ROLE_SUBSERVICER, types.Assignment_ASSIGNMENT_NEW),
				),
			},
		}},
	}
}

// --- terse builder helpers ----------------------------------------------------------------------

func sigReqAll(roles ...*types.RoleAssignment) *types.SignatureRequirement {
	return &types.SignatureRequirement{Type: types.SignatureType_SIGNATURE_TYPE_REQUIRED_ALL, Roles: roles}
}

func sigReqAllIfSet(roles ...*types.RoleAssignment) *types.SignatureRequirement {
	return &types.SignatureRequirement{Type: types.SignatureType_SIGNATURE_TYPE_REQUIRED_ALL_IF_SET, Roles: roles}
}

func raRegistry(role types.RegistryRole, assignment types.Assignment) *types.RoleAssignment {
	return &types.RoleAssignment{
		RoleSelector: &types.RoleAssignment_RegistryRole{RegistryRole: role},
		Assignment:   assignment,
	}
}

func raNft(nftRole types.NftRole, assignment types.Assignment) *types.RoleAssignment {
	return &types.RoleAssignment{
		RoleSelector: &types.RoleAssignment_NftRole{NftRole: nftRole},
		Assignment:   assignment,
	}
}

func raPriority(assignment types.Assignment, entries ...*types.RolePriorityEntry) *types.RoleAssignment {
	return &types.RoleAssignment{
		RoleSelector: &types.RoleAssignment_RolePriority{RolePriority: &types.RolePriority{Entries: entries}},
		Assignment:   assignment,
	}
}

func rpRegistry(role types.RegistryRole) *types.RolePriorityEntry {
	return &types.RolePriorityEntry{Role: &types.RolePriorityEntry_RegistryRole{RegistryRole: role}}
}

func rpNft(nftRole types.NftRole) *types.RolePriorityEntry {
	return &types.RolePriorityEntry{Role: &types.RolePriorityEntry_NftRole{NftRole: nftRole}}
}
