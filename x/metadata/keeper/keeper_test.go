package keeper_test

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/x/authz"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/metadata/keeper"

	"github.com/provenance-io/provenance/x/metadata/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	suite.Suite

	app         *simapp.App
	ctx         sdk.Context
	queryClient types.QueryClient

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress

	pubkey3   cryptotypes.PubKey
	user3     string
	user3Addr sdk.AccAddress

	objectLocator metadatatypes.ObjectStoreLocator
	ownerAddr     sdk.AccAddress
	encryptionKey sdk.AccAddress
	uri           string

	objectLocator1 metadatatypes.ObjectStoreLocator
	ownerAddr1     sdk.AccAddress
	encryptionKey1 sdk.AccAddress
	uri1           string
}

func (s *KeeperTestSuite) SetupTest() {
	msgfeestypes.DefaultFloorGasPrice = sdk.NewInt64Coin("atom", 0)
	s.app = simapp.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, s.app.MetadataKeeper)
	s.queryClient = types.NewQueryClient(queryHelper)

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()
	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	s.pubkey3 = secp256k1.GenPrivKey().PubKey()
	s.user3Addr = sdk.AccAddress(s.pubkey3.Address())
	s.user3 = s.user3Addr.String()

	// add os locator
	s.ownerAddr = s.user1Addr
	s.uri = "http://foo.com"
	s.encryptionKey = sdk.AccAddress{}
	s.objectLocator = metadatatypes.NewOSLocatorRecord(s.ownerAddr, s.encryptionKey, s.uri)

	s.ownerAddr1 = s.user2Addr
	s.uri1 = "http://bar.com"
	s.encryptionKey1 = sdk.AccAddress(s.pubkey1.Address())
	s.objectLocator1 = metadatatypes.NewOSLocatorRecord(s.ownerAddr1, s.encryptionKey1, s.uri1)
	//set up genesis
	var metadataData metadatatypes.GenesisState
	metadataData.Params = metadatatypes.DefaultParams()
	metadataData.OSLocatorParams = metadatatypes.DefaultOSLocatorParams()
	metadataData.ObjectStoreLocators = append(metadataData.ObjectStoreLocators, s.objectLocator, s.objectLocator1)
	s.app.MetadataKeeper.InitGenesis(s.ctx, &metadataData)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// ownerPartyList returns a party with role OWNER for each address provided.
// This func is used in other keeper test files.
func ownerPartyList(addresses ...string) []types.Party {
	retval := make([]types.Party, len(addresses))
	for i, addr := range addresses {
		retval[i] = types.Party{Address: addr, Role: types.PartyType_PARTY_TYPE_OWNER}
	}
	return retval
}

func (s *KeeperTestSuite) TestValidatePartiesInvolved() {

	cases := map[string]struct {
		parties         []types.Party
		requiredParties []types.PartyType
		wantErr         bool
		errorMsg        string
	}{
		"valid, matching no parties involved": {
			parties:         []types.Party{},
			requiredParties: []types.PartyType{},
			wantErr:         false,
			errorMsg:        "",
		},
		"invalid, parties contain no required parties": {
			parties:         []types.Party{},
			requiredParties: []types.PartyType{types.PartyType_PARTY_TYPE_AFFILIATE},
			wantErr:         true,
			errorMsg:        "missing required party type [PARTY_TYPE_AFFILIATE] from parties",
		},
		"invalid, missing one required party": {
			parties:         []types.Party{{Address: "address", Role: types.PartyType_PARTY_TYPE_CUSTODIAN}},
			requiredParties: []types.PartyType{types.PartyType_PARTY_TYPE_AFFILIATE},
			wantErr:         true,
			errorMsg:        "missing required party type [PARTY_TYPE_AFFILIATE] from parties",
		},
		"invalid, missing twp required parties": {
			parties:         []types.Party{{Address: "address", Role: types.PartyType_PARTY_TYPE_CUSTODIAN}},
			requiredParties: []types.PartyType{types.PartyType_PARTY_TYPE_AFFILIATE, types.PartyType_PARTY_TYPE_INVESTOR},
			wantErr:         true,
			errorMsg:        "missing required party types [PARTY_TYPE_AFFILIATE PARTY_TYPE_INVESTOR] from parties",
		},
		"valid, required parties fulfilled": {
			parties:         []types.Party{{Address: "address", Role: types.PartyType_PARTY_TYPE_CUSTODIAN}},
			requiredParties: []types.PartyType{types.PartyType_PARTY_TYPE_CUSTODIAN},
			wantErr:         false,
			errorMsg:        "",
		},
	}

	for n, tc := range cases {
		tc := tc

		s.T().Run(n, func(t *testing.T) {
			err := s.app.MetadataKeeper.ValidatePartiesInvolved(tc.parties, tc.requiredParties)
			if tc.wantErr {
				assert.EqualError(t, err, tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestValidateAllOwnerPartiesAreSigners() {

	cases := map[string]struct {
		owners     []types.Party
		signers    []string
		msgTypeURL string
		errorMsg   string
	}{
		"no owners - no signers": {
			owners:     []types.Party{},
			signers:    []string{},
			msgTypeURL: "",
			errorMsg:   "",
		},
		"one owner - is signer": {
			owners:     []types.Party{{Address: "signer1", Role: types.PartyType_PARTY_TYPE_OWNER}},
			signers:    []string{"signer1"},
			msgTypeURL: "",
			errorMsg:   "",
		},
		"one owner - is one of two signers": {
			owners:     []types.Party{{Address: "signer1", Role: types.PartyType_PARTY_TYPE_OWNER}},
			signers:    []string{"signer1", "signer2"},
			msgTypeURL: "",
			errorMsg:   "",
		},
		"one owner - is not one of two signers": {
			owners:     []types.Party{{Address: "missingowner", Role: types.PartyType_PARTY_TYPE_OWNER}},
			signers:    []string{"signer1", "signer2"},
			msgTypeURL: "",
			errorMsg:   "missing signature from [missingowner (PARTY_TYPE_OWNER)]",
		},
		"two owners - both are signers": {
			owners: []types.Party{
				{Address: "owner1", Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: "owner2", Role: types.PartyType_PARTY_TYPE_OWNER}},
			msgTypeURL: "",
			signers:    []string{"owner2", "owner1"},
			errorMsg:   "",
		},
		"two owners - only one is signer": {
			owners: []types.Party{
				{Address: "owner1", Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: "missingowner", Role: types.PartyType_PARTY_TYPE_OWNER}},
			signers:    []string{"owner2", "owner1"},
			msgTypeURL: "",
			errorMsg:   "missing signature from [missingowner (PARTY_TYPE_OWNER)]",
		},
		"two parties - one owner one other - only owner is signer": {
			owners: []types.Party{
				{Address: "owner", Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: "affiliate", Role: types.PartyType_PARTY_TYPE_AFFILIATE}},
			signers:    []string{"owner"},
			msgTypeURL: "",
			errorMsg:   "missing signature from [affiliate (PARTY_TYPE_AFFILIATE)]",
		},
		"two parties - one owner one other - only other is signer": {
			owners: []types.Party{
				{Address: "owner", Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: "affiliate", Role: types.PartyType_PARTY_TYPE_AFFILIATE}},
			signers:    []string{"affiliate"},
			msgTypeURL: "",
			errorMsg:   "missing signature from [owner (PARTY_TYPE_OWNER)]",
		},
		// authz test cases
		"two parties - one missing signature with authz grant - two signers": {
			owners: []types.Party{
				{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}}, // grantee of singer 1
			signers:    []string{s.user2, s.user3},
			msgTypeURL: types.TypeURLMsgWriteScopeRequest,
			errorMsg:   "",
		},
		"two parties - one missing signature without authz grant - one signer": {
			owners: []types.Party{
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}},
			signers:    []string{s.user2},
			msgTypeURL: types.TypeURLMsgWriteScopeRequest,
			errorMsg:   fmt.Sprintf("missing signature from [%s (PARTY_TYPE_OWNER)]", s.user3),
		},
		"two parties - one missing signature with a special case message type - authz grant - two signers": {
			owners: []types.Party{
				{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}}, // grantee of singer 2
			signers:    []string{s.user1, s.user3},
			msgTypeURL: types.TypeURLMsgAddScopeDataAccessRequest,
			errorMsg:   "",
		},
		"two parties - one missing signature with a special case message type - authz grant on parent message type - two signers": {
			owners: []types.Party{
				{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}}, // grantee of singer 2
			signers:    []string{s.user1, s.user3},
			msgTypeURL: types.TypeURLMsgAddContractSpecToScopeSpecRequest,
			errorMsg:   "",
		},
		"two parties - one missing signature with a special case message type without authz grant - one signer": {
			owners: []types.Party{
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
				{Address: s.user3, Role: types.PartyType_PARTY_TYPE_OWNER}},
			signers:    []string{s.user3},
			msgTypeURL: types.TypeURLMsgDeleteRecordRequest,
			errorMsg:   fmt.Sprintf("missing signature from [%s (PARTY_TYPE_OWNER)]", s.user2),
		},
	}

	// Add a few authorizations

	// A missing signature with an authz grant on MsgAddScopeOwnerRequest
	now := s.ctx.BlockHeader().Time
	s.Require().NotNil(now)
	granter := s.user1Addr
	grantee := s.user3Addr
	a := authz.NewGenericAuthorization(types.TypeURLMsgWriteScopeRequest)
	err := s.app.AuthzKeeper.SaveGrant(s.ctx, grantee, granter, a, now.Add(time.Hour))
	s.Require().NoError(err)

	// A missing signature on a special case message type with an authz grant on that type.
	// Add (a child msg type) TypeURLMsgAddScopeDataAccessRequest  (of a parent) TypeURLMsgWriteScopeRequest
	granter = s.user2Addr
	grantee = s.user3Addr
	a = authz.NewGenericAuthorization(types.TypeURLMsgAddScopeDataAccessRequest)
	err = s.app.AuthzKeeper.SaveGrant(s.ctx, grantee, granter, a, now.Add(time.Hour))
	s.Require().NoError(err)

	// A missing signature on a special case message type with an authz grant on its parent type.
	// Add grant on the parent type of TypeURLMsgAddContractSpecToScopeSpecRequest.
	granter = s.user2Addr
	grantee = s.user3Addr
	a = authz.NewGenericAuthorization(types.TypeURLMsgWriteScopeSpecificationRequest)
	err = s.app.AuthzKeeper.SaveGrant(s.ctx, grantee, granter, a, now.Add(time.Hour))
	s.Require().NoError(err)

	// Test cases
	for n, tc := range cases {
		s.T().Run(n, func(t *testing.T) {
			err := s.app.MetadataKeeper.ValidateAllPartiesAreSignersWithAuthz(s.ctx, tc.owners, tc.signers, tc.msgTypeURL)
			if len(tc.errorMsg) == 0 {
				assert.NoError(t, err, "%s unexpected error", n)
			} else {
				assert.EqualError(t, err, tc.errorMsg, "%s error", n)
			}
		})
	}
}

func (s *KeeperTestSuite) TestValidateAllOwnersAreSigners() {

	tests := map[string]struct {
		owners     []string
		signers    []string
		msgTypeURL string
		errorMsg   string
	}{
		"Scope Spec with 1 owner: no signers - error": {
			owners:     []string{s.user1},
			signers:    []string{},
			msgTypeURL: "",
			errorMsg:   fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1),
		},
		"Scope Spec with 1 owner: not in signers list - error": {
			owners:     []string{s.user1},
			signers:    []string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			msgTypeURL: "",
			errorMsg:   fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1),
		},
		"Scope Spec with 1 owner: in signers list with non-owners - ok": {
			owners:     []string{s.user1},
			signers:    []string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user1, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			msgTypeURL: "",
			errorMsg:   "",
		},
		"Scope Spec with 1 owner: only signer in list - ok": {
			owners:     []string{s.user1},
			signers:    []string{s.user1},
			msgTypeURL: "",
			errorMsg:   "",
		},
		"Scope Spec with 2 owners: no signers - error": {
			owners:     []string{s.user1, s.user2},
			signers:    []string{},
			msgTypeURL: "",
			errorMsg: fmt.Sprintf("missing signatures from existing owners %v; required for update",
				[]string{s.user1, s.user2}),
		},
		"Scope Spec with 2 owners: neither in signers list - error": {
			owners:     []string{s.user1, s.user2},
			signers:    []string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			msgTypeURL: "",
			errorMsg: fmt.Sprintf("missing signatures from existing owners %v; required for update",
				[]string{s.user1, s.user2}),
		},
		"Scope Spec with 2 owners: one in signers list with non-owners - error": {
			owners:     []string{s.user1, s.user2},
			signers:    []string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user1, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			msgTypeURL: "",
			errorMsg:   fmt.Sprintf("missing signature from existing owner %s; required for update", s.user2),
		},
		"Scope Spec with 2 owners: the other in signers list with non-owners - error": {
			owners:     []string{s.user1, s.user2},
			signers:    []string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user2, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			msgTypeURL: "",
			errorMsg:   fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1),
		},
		"Scope Spec with 2 owners: both in signers list with non-owners - ok": {
			owners:     []string{s.user1, s.user2},
			signers:    []string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user2, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user1},
			msgTypeURL: "",
			errorMsg:   "",
		},
		"Scope Spec with 2 owners: only both in signers list - ok": {
			owners:     []string{s.user1, s.user2},
			signers:    []string{s.user1, s.user2},
			msgTypeURL: "",
			errorMsg:   "",
		},
		"Scope Spec with 2 owners: only both in signers list, opposite order - ok": {
			owners:     []string{s.user1, s.user2},
			signers:    []string{s.user2, s.user1},
			msgTypeURL: "",
			errorMsg:   "",
		},
		// authz test cases
		"Scope Spec with 2 owners - both in signers list - authz": {
			owners:     []string{s.user2, s.user3},
			signers:    []string{s.user2, s.user3},
			msgTypeURL: types.TypeURLMsgAddScopeDataAccessRequest,
			errorMsg:   "",
		},
		"Scope Spec with 2 owners - one signer - authz - error": {
			owners:     []string{s.user2, s.user3},
			signers:    []string{s.user2},
			msgTypeURL: types.TypeURLMsgWriteScopeRequest,
			errorMsg:   fmt.Sprintf("missing signature from existing owner %s; required for update", s.user3),
		},
		"Scope Spec with 3 owners - one signer with a special case message type - with grant - authz": {
			owners:     []string{s.user1, s.user2, s.user3},
			signers:    []string{s.user1, s.user3}, // signer 3 is grantee of singer 2
			msgTypeURL: types.TypeURLMsgAddScopeDataAccessRequest,
			errorMsg:   "",
		},
		"Scope Spec with 3 owners - two signers with a special case message type - grant on parent of special case message type - authz": {
			owners:     []string{s.user1, s.user2, s.user3},
			signers:    []string{s.user1, s.user3}, // signer 3 grantee of signer 2
			msgTypeURL: types.TypeURLMsgDeleteContractSpecFromScopeSpecRequest,
			errorMsg:   "",
		},
		"Scope Spec with 2 owners - one signer - no grant - authz - error": {
			owners:     []string{s.user2, s.user3},
			signers:    []string{s.user3},
			msgTypeURL: types.TypeURLMsgDeleteRecordRequest,
			errorMsg:   fmt.Sprintf("missing signature from existing owner %s; required for update", s.user2),
		},
	}

	// Add a few authorizations

	// A missing signature with an authz grant on MsgAddScopeOwnerRequest
	now := s.ctx.BlockHeader().Time
	s.Require().NotNil(now)
	granter := s.user1Addr
	grantee := s.user3Addr
	a := authz.NewGenericAuthorization(types.TypeURLMsgAddScopeOwnerRequest)
	err := s.app.AuthzKeeper.SaveGrant(s.ctx, grantee, granter, a, now.Add(time.Hour))
	s.Require().NoError(err)

	// A missing signature on a special case message type with an authz grant on that type.
	// Add (a child msg type) TypeURLMsgAddScopeDataAccessRequest  (of a parent) TypeURLMsgWriteScopeRequest
	granter = s.user2Addr
	grantee = s.user3Addr
	a = authz.NewGenericAuthorization(types.TypeURLMsgAddScopeDataAccessRequest)
	err = s.app.AuthzKeeper.SaveGrant(s.ctx, grantee, granter, a, now.Add(time.Hour))
	s.Require().NoError(err)

	// A missing signature on a special case message type with an authz grant on its parent type.
	// Add grant on the parent type of TypeURLMsgAddContractSpecToScopeSpecRequest.
	granter = s.user2Addr
	grantee = s.user3Addr
	a = authz.NewGenericAuthorization(types.TypeURLMsgWriteScopeSpecificationRequest)
	err = s.app.AuthzKeeper.SaveGrant(s.ctx, grantee, granter, a, now.Add(time.Hour))
	s.Require().NoError(err)

	for n, tc := range tests {
		s.T().Run(n, func(t *testing.T) {
			err := s.app.MetadataKeeper.ValidateAllOwnersAreSignersWithAuthz(s.ctx, tc.owners, tc.signers, tc.msgTypeURL)
			if len(tc.errorMsg) == 0 {
				assert.NoError(t, err, "ValidateAllOwnersAreSigners unexpected error")
			} else {
				assert.EqualError(t, err, tc.errorMsg, "ValidateAllOwnersAreSigners error")
			}
		})
	}
}

func (s *KeeperTestSuite) TestFindMissing() {
	tests := map[string]struct {
		required []string
		entries  []string
		expected []string
	}{
		"empty required - empty entries - empty out": {
			[]string{},
			[]string{},
			[]string{},
		},
		"empty required - 2 entries - empty out": {
			[]string{},
			[]string{"one", "two"},
			[]string{},
		},
		"one required - is only entry - empty out": {
			[]string{"one"},
			[]string{"one"},
			[]string{},
		},
		"one required - is first of two entries - empty out": {
			[]string{"one"},
			[]string{"one", "two"},
			[]string{},
		},
		"one required - is second of two entries - empty out": {
			[]string{"one"},
			[]string{"two", "one"},
			[]string{},
		},
		"one required - empty entries - required out": {
			[]string{"one"},
			[]string{},
			[]string{"one"},
		},
		"one required - one other in entries - required out": {
			[]string{"one"},
			[]string{"two"},
			[]string{"one"},
		},
		"one required - two other in entries - required out": {
			[]string{"one"},
			[]string{"two", "three"},
			[]string{"one"},
		},
		"two required - both in entries - empty out": {
			[]string{"one", "two"},
			[]string{"one", "two"},
			[]string{},
		},
		"two required - reversed in entries - empty out": {
			[]string{"one", "two"},
			[]string{"two", "one"},
			[]string{},
		},
		"two required - only first in entries - second out": {
			[]string{"one", "two"},
			[]string{"one"},
			[]string{"two"},
		},
		"two required - only second in entries - first out": {
			[]string{"one", "two"},
			[]string{"two"},
			[]string{"one"},
		},
		"two required - first and other in entries - second out": {
			[]string{"one", "two"},
			[]string{"one", "other"},
			[]string{"two"},
		},
		"two required - second and other in entries - first out": {
			[]string{"one", "two"},
			[]string{"two", "other"},
			[]string{"one"},
		},
		"two required - empty entries - required out": {
			[]string{"one", "two"},
			[]string{},
			[]string{"one", "two"},
		},
		"two required - neither in one entries - required out": {
			[]string{"one", "two"},
			[]string{"neither"},
			[]string{"one", "two"},
		},
		"two required - neither in three entries - required out": {
			[]string{"one", "two"},
			[]string{"neither", "nor", "nothing"},
			[]string{"one", "two"},
		},
		"two required - first not in three entries 0 - first out": {
			[]string{"one", "two"},
			[]string{"two", "nor", "nothing"},
			[]string{"one"},
		},
		"two required - first not in three entries 1 - first out": {
			[]string{"one", "two"},
			[]string{"neither", "two", "nothing"},
			[]string{"one"},
		},
		"two required - first not in three entries 2 - first out": {
			[]string{"one", "two"},
			[]string{"neither", "nor", "two"},
			[]string{"one"},
		},
		"two required - second not in three entries 0 - second out": {
			[]string{"one", "two"},
			[]string{"one", "nor", "nothing"},
			[]string{"two"},
		},
		"two required - second not in three entries 1 - second out": {
			[]string{"one", "two"},
			[]string{"neither", "one", "nothing"},
			[]string{"two"},
		},
		"two required - second not in three entries 2 - second out": {
			[]string{"one", "two"},
			[]string{"neither", "nor", "one"},
			[]string{"two"},
		},
	}

	for n, tc := range tests {
		s.T().Run(n, func(t *testing.T) {
			actual := keeper.FindMissing(tc.required, tc.entries)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func (s *KeeperTestSuite) TestFindMissingMdAddr() {
	tests := map[string]struct {
		required []string
		entries  []string
		expected []string
	}{
		"empty required - empty entries - empty out": {
			[]string{},
			[]string{},
			[]string{},
		},
		"empty required - 2 entries - empty out": {
			[]string{},
			[]string{"one", "two"},
			[]string{},
		},
		"one required - is only entry - empty out": {
			[]string{"one"},
			[]string{"one"},
			[]string{},
		},
		"one required - is first of two entries - empty out": {
			[]string{"one"},
			[]string{"one", "two"},
			[]string{},
		},
		"one required - is second of two entries - empty out": {
			[]string{"one"},
			[]string{"two", "one"},
			[]string{},
		},
		"one required - empty entries - required out": {
			[]string{"one"},
			[]string{},
			[]string{"one"},
		},
		"one required - one other in entries - required out": {
			[]string{"one"},
			[]string{"two"},
			[]string{"one"},
		},
		"one required - two other in entries - required out": {
			[]string{"one"},
			[]string{"two", "three"},
			[]string{"one"},
		},
		"two required - both in entries - empty out": {
			[]string{"one", "two"},
			[]string{"one", "two"},
			[]string{},
		},
		"two required - reversed in entries - empty out": {
			[]string{"one", "two"},
			[]string{"two", "one"},
			[]string{},
		},
		"two required - only first in entries - second out": {
			[]string{"one", "two"},
			[]string{"one"},
			[]string{"two"},
		},
		"two required - only second in entries - first out": {
			[]string{"one", "two"},
			[]string{"two"},
			[]string{"one"},
		},
		"two required - first and other in entries - second out": {
			[]string{"one", "two"},
			[]string{"one", "other"},
			[]string{"two"},
		},
		"two required - second and other in entries - first out": {
			[]string{"one", "two"},
			[]string{"two", "other"},
			[]string{"one"},
		},
		"two required - empty entries - required out": {
			[]string{"one", "two"},
			[]string{},
			[]string{"one", "two"},
		},
		"two required - neither in one entries - required out": {
			[]string{"one", "two"},
			[]string{"neither"},
			[]string{"one", "two"},
		},
		"two required - neither in three entries - required out": {
			[]string{"one", "two"},
			[]string{"neither", "nor", "nothing"},
			[]string{"one", "two"},
		},
		"two required - first not in three entries 0 - first out": {
			[]string{"one", "two"},
			[]string{"two", "nor", "nothing"},
			[]string{"one"},
		},
		"two required - first not in three entries 1 - first out": {
			[]string{"one", "two"},
			[]string{"neither", "two", "nothing"},
			[]string{"one"},
		},
		"two required - first not in three entries 2 - first out": {
			[]string{"one", "two"},
			[]string{"neither", "nor", "two"},
			[]string{"one"},
		},
		"two required - second not in three entries 0 - second out": {
			[]string{"one", "two"},
			[]string{"one", "nor", "nothing"},
			[]string{"two"},
		},
		"two required - second not in three entries 1 - second out": {
			[]string{"one", "two"},
			[]string{"neither", "one", "nothing"},
			[]string{"two"},
		},
		"two required - second not in three entries 2 - second out": {
			[]string{"one", "two"},
			[]string{"neither", "nor", "one"},
			[]string{"two"},
		},
	}

	// For these tests, we shouldn't need valid metadata addresses.
	// So just convert the strings into byte arrays and use those
	// as MetadataAddresses. That way, I can just use the same tests
	// as the ones from TestFindMissing()
	stringsToAddrs := func(vals []string) []types.MetadataAddress {
		rv := make([]types.MetadataAddress, len(vals))
		for i, val := range vals {
			rv[i] = types.MetadataAddress(val)
		}
		return rv
	}

	for n, tc := range tests {
		s.T().Run(n, func(t *testing.T) {
			required := stringsToAddrs(tc.required)
			entries := stringsToAddrs(tc.entries)
			expected := stringsToAddrs(tc.expected)
			actual := keeper.FindMissingMdAddr(required, entries)
			assert.Equal(t, expected, actual)
		})
	}
}

func (s *KeeperTestSuite) TestParams() {
	s.T().Run("param tests", func(t *testing.T) {
		p := s.app.MetadataKeeper.GetParams(s.ctx)
		assert.NotNil(t, p)

		osp := s.app.MetadataKeeper.GetOSLocatorParams(s.ctx)
		assert.NotNil(t, osp)
		assert.Equal(t, osp.MaxUriLength, s.app.MetadataKeeper.GetMaxURILength(s.ctx))
	})
}

func (s *KeeperTestSuite) TestGetOSLocator() {
	s.Run("get os locator by owner address", func() {
		r, found := s.app.MetadataKeeper.GetOsLocatorRecord(s.ctx, s.user1Addr)
		s.Require().NotEmpty(r)
		s.Require().True(found)
	})
	s.Run("not found by owner address", func() {
		r, found := s.app.MetadataKeeper.GetOsLocatorRecord(s.ctx, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()))
		s.Require().Empty(r)
		s.Require().False(found)
	})
}

func (s *KeeperTestSuite) TestAddOSLocator() {
	s.Run("add os locator", func() {
		// create account and check default values
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user3Addr)
		s.Require().NotNil(acc)
		s.Require().Equal(s.user3Addr, acc.GetAddress())
		s.Require().EqualValues(nil, acc.GetPubKey())
		s.Require().EqualValues(0, acc.GetSequence())
		// set and get the new account.
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		acc1 := s.app.AccountKeeper.GetAccount(s.ctx, s.user3Addr)
		s.Require().NotNil(acc1)
		// create os locator with ^^ account
		err := s.app.MetadataKeeper.SetOSLocator(s.ctx, s.user3Addr, sdk.AccAddress{}, "https://bob.com/alice")
		s.Require().Empty(err)
		r, found := s.app.MetadataKeeper.GetOsLocatorRecord(s.ctx, s.user1Addr)
		s.Require().NotEmpty(r)
		s.Require().True(found)
	})

	s.Run("add os locator account does not exist.", func() {
		// create account and check default values
		err := s.app.MetadataKeeper.SetOSLocator(s.ctx, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()), sdk.AccAddress{}, "https://bob.com/alice")
		s.Require().NotEmpty(err)
	})

	s.Run("add os bad uri.", func() {
		pubkey4 := secp256k1.GenPrivKey().PubKey()
		user4Addr := sdk.AccAddress(pubkey4.Address())
		// create account and check default values
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, user4Addr)
		s.Require().NotNil(acc)
		s.Require().Equal(user4Addr, acc.GetAddress())
		s.Require().EqualValues(nil, acc.GetPubKey())
		s.Require().EqualValues(0, acc.GetSequence())
		// set and get the new account.
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		acc1 := s.app.AccountKeeper.GetAccount(s.ctx, user4Addr)
		s.Require().NotNil(acc1)
		// create os locator with ^^ account
		err := s.app.MetadataKeeper.SetOSLocator(s.ctx, user4Addr, s.encryptionKey, "foo.com")
		s.Require().NotEmpty(err)
		r, found := s.app.MetadataKeeper.GetOsLocatorRecord(s.ctx, user4Addr)
		s.Require().Empty(r)
		s.Require().False(found)
	})
}

func (s *KeeperTestSuite) TestModifyOSLocator() {
	s.Run("modify os locator", func() {
		// modify os locator
		err := s.app.MetadataKeeper.ModifyOSLocator(s.ctx, s.user1Addr, s.encryptionKey, "https://bob.com/alice")
		s.Require().Empty(err)
		r, found := s.app.MetadataKeeper.GetOsLocatorRecord(s.ctx, s.user1Addr)
		s.Require().NotEmpty(r)
		s.Require().True(found)
		s.Require().Equal(s.encryptionKey.String(), r.EncryptionKey)
		s.Require().Equal("https://bob.com/alice", r.LocatorUri)
	})
	s.Run("modify os locator invalid uri", func() {
		// modify os locator
		err := s.app.MetadataKeeper.ModifyOSLocator(s.ctx, s.user1Addr, s.encryptionKey, "://bob.com/alice")
		s.Require().NotEmpty(err)
	})

	s.Run("modify os locator invalid uri length", func() {
		// modify os locator
		err := s.app.MetadataKeeper.ModifyOSLocator(s.ctx, s.user1Addr, s.encryptionKey1, "https://www.google.com/search?q=long+url+example&oq=long+uril+&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8&aqs=chrome.1.69i57j0i13l9.4447j0j15&sourceid=chrome&ie=UTF-8")
		s.Require().NotEmpty(err)
		s.Require().Equal("uri length greater than allowed", err.Error())
	})
}

func (s *KeeperTestSuite) TestDeleteOSLocator() {
	s.Run("delete os locator", func() {
		// modify os locator
		err := s.app.MetadataKeeper.RemoveOSLocator(s.ctx, s.user1Addr)
		s.Require().Empty(err)
		r, found := s.app.MetadataKeeper.GetOsLocatorRecord(s.ctx, s.user1Addr)
		s.Require().Empty(r)
		s.Require().False(found)

	})
}

func (s *KeeperTestSuite) TestUnionDistinct() {
	tests := []struct {
		name   string
		inputs [][]string
		output []string
	}{
		{
			"empty in empty out",
			[][]string{},
			[]string{},
		},
		{
			"one set in same set out",
			[][]string{{"a", "b", "c"}},
			[]string{"a", "b", "c"},
		},
		{
			"two dup sets in single entries out",
			[][]string{{"a", "b", "c"}, {"a", "b", "c"}},
			[]string{"a", "b", "c"},
		},
		{
			"unique sets in combined for out",
			[][]string{{"a", "b", "c"}, {"d", "e"}},
			[]string{"a", "b", "c", "d", "e"},
		},
		{
			"empty set filled set in combined for out",
			[][]string{{}, {"a", "b", "c"}},
			[]string{"a", "b", "c"},
		},
		{
			"filled set empty set in combined for out",
			[][]string{{"a", "b", "c"}, {}},
			[]string{"a", "b", "c"},
		},
		{
			"two sets with one common entry in combined correctly for out",
			[][]string{{"a", "b", "c"}, {"d", "a", "e"}},
			[]string{"a", "b", "c", "d", "e"},
		},
		{
			"set with one entry and set with two entries in combined correctly for out",
			[][]string{{"a"}, {"a", "b"}},
			[]string{"a", "b"},
		},
		{
			"set with two entries set with one entry in combined correctly for out",
			[][]string{{"a", "b"}, {"a"}},
			[]string{"a", "b"},
		},
		{
			"set with dups and set with two entries in combined correctly for out",
			[][]string{{"a", "a"}, {"a", "b"}},
			[]string{"a", "b"},
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			output := s.app.MetadataKeeper.UnionDistinct(tc.inputs...)
			sort.Strings(output)
			sort.Strings(tc.output)
			assert.Equal(t, tc.output, output)
		})
	}
}

func (s *KeeperTestSuite) TestGetMsgTypeURLs() {

	expected := []string{types.TypeURLMsgAddScopeDataAccessRequest, types.TypeURLMsgWriteScopeRequest}
	urls := s.app.MetadataKeeper.GetMessageTypeURLs(types.TypeURLMsgAddScopeDataAccessRequest)
	assert.Equal(s.T(), expected, urls)

	expected = []string{types.TypeURLMsgDeleteScopeDataAccessRequest, types.TypeURLMsgWriteScopeRequest}
	urls = s.app.MetadataKeeper.GetMessageTypeURLs(types.TypeURLMsgDeleteScopeDataAccessRequest)
	assert.Equal(s.T(), expected, urls)

	expected = []string{types.TypeURLMsgDeleteScopeOwnerRequest, types.TypeURLMsgWriteScopeRequest}
	urls = s.app.MetadataKeeper.GetMessageTypeURLs(types.TypeURLMsgDeleteScopeOwnerRequest)
	assert.Equal(s.T(), expected, urls)

	expected = []string{types.TypeURLMsgWriteRecordRequest, types.TypeURLMsgWriteSessionRequest}
	urls = s.app.MetadataKeeper.GetMessageTypeURLs(types.TypeURLMsgWriteRecordRequest)
	assert.Equal(s.T(), expected, urls)

	expected = []string{types.TypeURLMsgAddContractSpecToScopeSpecRequest, types.TypeURLMsgWriteScopeSpecificationRequest}
	urls = s.app.MetadataKeeper.GetMessageTypeURLs(types.TypeURLMsgAddContractSpecToScopeSpecRequest)
	assert.Equal(s.T(), expected, urls)

	expected = []string{types.TypeURLMsgDeleteContractSpecFromScopeSpecRequest, types.TypeURLMsgWriteScopeSpecificationRequest}
	urls = s.app.MetadataKeeper.GetMessageTypeURLs(types.TypeURLMsgDeleteContractSpecFromScopeSpecRequest)
	assert.Equal(s.T(), expected, urls)

	expected = []string{types.TypeURLMsgAddContractSpecToScopeSpecRequest, types.TypeURLMsgWriteScopeSpecificationRequest}
	urls = s.app.MetadataKeeper.GetMessageTypeURLs(types.TypeURLMsgAddContractSpecToScopeSpecRequest)
	assert.Equal(s.T(), expected, urls)

	expected = []string{types.TypeURLMsgWriteRecordSpecificationRequest, types.TypeURLMsgWriteContractSpecificationRequest}
	urls = s.app.MetadataKeeper.GetMessageTypeURLs(types.TypeURLMsgWriteRecordSpecificationRequest)
	assert.Equal(s.T(), expected, urls)

	expected = []string{types.TypeURLMsgDeleteRecordSpecificationRequest, types.TypeURLMsgDeleteContractSpecificationRequest}
	urls = s.app.MetadataKeeper.GetMessageTypeURLs(types.TypeURLMsgDeleteRecordSpecificationRequest)
	assert.Equal(s.T(), expected, urls)
}
