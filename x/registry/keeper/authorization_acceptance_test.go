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

// ControllerAuthAcceptanceTestSuite implements the acceptance tests from the ticket
// (sc-512248) that verify the static multi-signer authorization rules for updating
// the CONTROLLER role.
//
// Required signatures for a Controller update:
//   - Current Controller
//   - Current Secured Party for eNote (only if set)
//   - New Controller
type ControllerAuthAcceptanceTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	registryKeeper keeper.Keeper
	nftKeeper      nftkeeper.Keeper
	msgServer      types.MsgServer

	nftClass nft.Class

	// nftOwner owns the NFT and is the fallback authority when no controller is set.
	nftOwner string

	currentController string
	newController     string
	securedParty      string
	stranger          string
}

func TestControllerAuthAcceptanceTestSuite(t *testing.T) {
	suite.Run(t, new(ControllerAuthAcceptanceTestSuite))
}

func (s *ControllerAuthAcceptanceTestSuite) SetupTest() {
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

	s.nftClass = nft.Class{Id: "test-nft-class-id"}
	s.nftKeeper.SaveClass(s.ctx, s.nftClass)
}

// genAddr generates a fresh bech32 account address.
func genAddr() string {
	return sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()
}

// mintNFT mints an NFT with the given id owned by nftOwner and returns its registry key.
func (s *ControllerAuthAcceptanceTestSuite) mintNFT(id string) *types.RegistryKey {
	n := nft.NFT{ClassId: s.nftClass.Id, Id: id}
	ownerAddr, err := sdk.AccAddressFromBech32(s.nftOwner)
	s.Require().NoError(err)
	s.Require().NoError(s.nftKeeper.Mint(s.ctx, n, ownerAddr))
	return &types.RegistryKey{AssetClassId: s.nftClass.Id, NftId: id}
}

// setupRegistry creates a registry entry for the key with the provided roles.
func (s *ControllerAuthAcceptanceTestSuite) setupRegistry(key *types.RegistryKey, roles []types.RolesEntry) {
	s.Require().NoError(s.registryKeeper.CreateRegistry(s.ctx, key, roles))
}

// grantController attempts to grant the CONTROLLER role to newController, signed by the
// provided signers, and returns the resulting error (nil on success).
func (s *ControllerAuthAcceptanceTestSuite) grantController(key *types.RegistryKey, signers []string) error {
	_, err := s.msgServer.GrantRole(s.ctx, &types.MsgGrantRole{
		Signers:   signers,
		Key:       key,
		Role:      types.RegistryRole_REGISTRY_ROLE_CONTROLLER,
		Addresses: []string{s.newController},
	})
	return err
}

// roleAddresses returns the addresses currently assigned to the given role on the entry.
func (s *ControllerAuthAcceptanceTestSuite) roleAddresses(key *types.RegistryKey, role types.RegistryRole) []string {
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

func (s *ControllerAuthAcceptanceTestSuite) TestControllerSet_NoSecuredParty() {
	controllerRole := types.RegistryRole_REGISTRY_ROLE_CONTROLLER

	baseRoles := func() []types.RolesEntry {
		return []types.RolesEntry{
			{Role: controllerRole, Addresses: []string{s.currentController}},
		}
	}

	s.Run("happy path: current controller and new controller sign", func() {
		key := s.mintNFT("nft-a-happy")
		s.setupRegistry(key, baseRoles())

		err := s.grantController(key, []string{s.currentController, s.newController})
		s.Require().NoError(err, "current + new controller should authorize the update")

		s.Require().Contains(s.roleAddresses(key, controllerRole), s.newController,
			"new controller should be added to the CONTROLLER role")
	})

	s.Run("reject: only current controller signs", func() {
		key := s.mintNFT("nft-a-cc-only")
		s.setupRegistry(key, baseRoles())

		err := s.grantController(key, []string{s.currentController})
		s.Require().Error(err, "missing new controller signature should be rejected")
	})

	s.Run("reject: only new controller signs", func() {
		key := s.mintNFT("nft-a-nc-only")
		s.setupRegistry(key, baseRoles())

		err := s.grantController(key, []string{s.newController})
		s.Require().Error(err, "missing current controller signature should be rejected")
	})
}

// --- Scenario B: Controller set, Secured Party for eNote set --------------------------

func (s *ControllerAuthAcceptanceTestSuite) TestControllerSet_WithSecuredParty() {
	controllerRole := types.RegistryRole_REGISTRY_ROLE_CONTROLLER
	securedPartyRole := types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE

	baseRoles := func() []types.RolesEntry {
		return []types.RolesEntry{
			{Role: controllerRole, Addresses: []string{s.currentController}},
			{Role: securedPartyRole, Addresses: []string{s.securedParty}},
		}
	}

	s.Run("happy path: current controller, secured party, and new controller sign", func() {
		key := s.mintNFT("nft-b-happy")
		s.setupRegistry(key, baseRoles())

		err := s.grantController(key, []string{s.currentController, s.securedParty, s.newController})
		s.Require().NoError(err, "all three required signers should authorize the update")

		s.Require().Contains(s.roleAddresses(key, controllerRole), s.newController,
			"new controller should be added to the CONTROLLER role")
	})

	s.Run("reject: only current controller signs", func() {
		key := s.mintNFT("nft-b-cc-only")
		s.setupRegistry(key, baseRoles())

		err := s.grantController(key, []string{s.currentController})
		s.Require().Error(err)
	})

	s.Run("reject: only new controller signs", func() {
		key := s.mintNFT("nft-b-nc-only")
		s.setupRegistry(key, baseRoles())

		err := s.grantController(key, []string{s.newController})
		s.Require().Error(err)
	})

	s.Run("reject: only secured party signs", func() {
		key := s.mintNFT("nft-b-sp-only")
		s.setupRegistry(key, baseRoles())

		err := s.grantController(key, []string{s.securedParty})
		s.Require().Error(err)
	})

	s.Run("reject: current controller and secured party sign (missing new controller)", func() {
		key := s.mintNFT("nft-b-cc-sp")
		s.setupRegistry(key, baseRoles())

		err := s.grantController(key, []string{s.currentController, s.securedParty})
		s.Require().Error(err)
	})

	s.Run("reject: current controller and new controller sign (missing secured party)", func() {
		key := s.mintNFT("nft-b-cc-nc")
		s.setupRegistry(key, baseRoles())

		err := s.grantController(key, []string{s.currentController, s.newController})
		s.Require().Error(err, "secured party signature is required when it is set")
	})

	s.Run("reject: secured party and new controller sign (missing current controller)", func() {
		key := s.mintNFT("nft-b-sp-nc")
		s.setupRegistry(key, baseRoles())

		err := s.grantController(key, []string{s.securedParty, s.newController})
		s.Require().Error(err, "current controller signature is always required (no unilateral takeover)")
	})

	s.Run("reject: an unrelated stranger signs", func() {
		key := s.mintNFT("nft-b-stranger")
		s.setupRegistry(key, baseRoles())

		err := s.grantController(key, []string{s.stranger})
		s.Require().Error(err)
	})
}

// --- AuthZ grant scenarios ------------------------------------------------------------
//
// The ticket asks us to verify the same scenarios "but signed with authz grants". These tests
// document a hard architectural constraint discovered during the PoC: the Cosmos SDK authz
// module can only dispatch messages that have exactly one signer (see
// x/authz/keeper/keeper.go DispatchActions, which returns ErrAuthorizationNumOfSigners when
// len(signers) != 1).
//
// Because a Controller update always requires at least two signers (current controller and new
// controller, plus the secured party when set), a multi-signer MsgGrantRole cannot be executed
// through an authz MsgExec at all. Delegation via authz would therefore require a different
// multisig model (e.g. a single multisig account/key whose message has one signer), which is out
// of scope for this PoC.

// mustAddr converts a bech32 string into an sdk.AccAddress, failing the test on error.
func (s *ControllerAuthAcceptanceTestSuite) mustAddr(addr string) sdk.AccAddress {
	a, err := sdk.AccAddressFromBech32(addr)
	s.Require().NoError(err)
	return a
}

// grantGrantRoleAuthz grants the executor authority to execute MsgGrantRole on behalf of granter.
func (s *ControllerAuthAcceptanceTestSuite) grantGrantRoleAuthz(executor, granter string) {
	auth := authz.NewGenericAuthorization(sdk.MsgTypeURL(&types.MsgGrantRole{}))
	s.Require().NoError(s.app.AuthzKeeper.SaveGrant(s.ctx, s.mustAddr(executor), s.mustAddr(granter), auth, nil))
}

// execGrantControllerViaAuthz submits a MsgGrantRole (with the given signers) wrapped in a
// MsgExec run by executor, and returns the resulting error.
func (s *ControllerAuthAcceptanceTestSuite) execGrantControllerViaAuthz(executor string, key *types.RegistryKey, signers []string) error {
	inner := &types.MsgGrantRole{
		Signers:   signers,
		Key:       key,
		Role:      types.RegistryRole_REGISTRY_ROLE_CONTROLLER,
		Addresses: []string{s.newController},
	}
	msgExec := authz.NewMsgExec(s.mustAddr(executor), []sdk.Msg{inner})
	_, err := s.app.AuthzKeeper.Exec(s.ctx, &msgExec)
	return err
}

func (s *ControllerAuthAcceptanceTestSuite) TestControllerUpdate_ViaAuthzGrants() {
	controllerRole := types.RegistryRole_REGISTRY_ROLE_CONTROLLER
	securedPartyRole := types.RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE
	executor := genAddr()

	s.Run("authz cannot dispatch a multi-signer controller update", func() {
		key := s.mintNFT("nft-authz-multisig")
		s.setupRegistry(key, []types.RolesEntry{
			{Role: controllerRole, Addresses: []string{s.currentController}},
			{Role: securedPartyRole, Addresses: []string{s.securedParty}},
		})

		// Even with grants from every required party, authz refuses to dispatch because the
		// inner MsgGrantRole has more than one signer.
		s.grantGrantRoleAuthz(executor, s.currentController)
		s.grantGrantRoleAuthz(executor, s.securedParty)
		s.grantGrantRoleAuthz(executor, s.newController)

		err := s.execGrantControllerViaAuthz(executor, key, []string{s.currentController, s.securedParty, s.newController})
		s.Require().Error(err, "authz must reject a multi-signer message")
		s.Require().ErrorIs(err, authz.ErrAuthorizationNumOfSigners,
			"authz rejects multi-signer messages with ErrAuthorizationNumOfSigners")

		// The controller role must remain unchanged since the update never executed.
		s.Require().NotContains(s.roleAddresses(key, controllerRole), s.newController)
	})
}
