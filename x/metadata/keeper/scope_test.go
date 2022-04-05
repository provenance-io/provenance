package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	simapp "github.com/provenance-io/provenance/app"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	"github.com/provenance-io/provenance/x/metadata/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type ScopeKeeperTestSuite struct {
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

	scopeUUID uuid.UUID
	scopeID   types.MetadataAddress

	scopeSpecUUID uuid.UUID
	scopeSpecID   types.MetadataAddress
}

func (s *ScopeKeeperTestSuite) SetupTest() {
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

	s.scopeUUID = uuid.New()
	s.scopeID = types.ScopeMetadataAddress(s.scopeUUID)

	s.scopeSpecUUID = uuid.New()
	s.scopeSpecID = types.ScopeSpecMetadataAddress(s.scopeSpecUUID)
}

func TestScopeKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(ScopeKeeperTestSuite))
}

// func ownerPartyList defined in keeper_test.go

type user struct {
	PrivKey cryptotypes.PrivKey
	PubKey  cryptotypes.PubKey
	Addr    sdk.AccAddress
	Bech32  string
}

func randomUser() user {
	rv := user{}
	rv.PrivKey = secp256k1.GenPrivKey()
	rv.PubKey = rv.PrivKey.PubKey()
	rv.Addr = sdk.AccAddress(rv.PubKey.Address())
	rv.Bech32 = rv.Addr.String()
	return rv
}

func (s *ScopeKeeperTestSuite) TestMetadataScopeGetSet() {
	scope, found := s.app.MetadataKeeper.GetScope(s.ctx, s.scopeID)
	s.NotNil(scope)
	s.False(found)

	ns := *types.NewScope(s.scopeID, s.scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, s.user1)
	s.NotNil(ns)
	s.app.MetadataKeeper.SetScope(s.ctx, ns)

	scope, found = s.app.MetadataKeeper.GetScope(s.ctx, s.scopeID)
	s.True(found)
	s.NotNil(scope)

	s.app.MetadataKeeper.RemoveScope(s.ctx, ns.ScopeId)
	scope, found = s.app.MetadataKeeper.GetScope(s.ctx, s.scopeID)
	s.False(found)
	s.NotNil(scope)
}

func (s *ScopeKeeperTestSuite) TestMetadataScopeIterator() {
	for i := 1; i <= 10; i++ {
		valueOwner := ""
		if i == 5 {
			valueOwner = s.user2
		}
		ns := types.NewScope(types.ScopeMetadataAddress(uuid.New()), nil, ownerPartyList(s.user1), []string{s.user1}, valueOwner)
		s.app.MetadataKeeper.SetScope(s.ctx, *ns)
	}
	count := 0
	s.app.MetadataKeeper.IterateScopes(s.ctx, func(s types.Scope) (stop bool) {
		count++
		return false
	})
	s.Equal(10, count, "iterator should return a full list of scopes")

	count = 0
	s.app.MetadataKeeper.IterateScopesForAddress(s.ctx, s.user1Addr, func(scopeID types.MetadataAddress) (stop bool) {
		count++
		s.True(scopeID.IsScopeAddress())
		return false
	})
	s.Equal(10, count, "iterator should return ten scope addresses")

	count = 0
	s.app.MetadataKeeper.IterateScopesForAddress(s.ctx, s.user2Addr, func(scopeID types.MetadataAddress) (stop bool) {
		count++
		s.True(scopeID.IsScopeAddress())
		return false
	})
	s.Equal(1, count, "iterator should return a single address for the scope with value owned by user2")

	count = 0
	s.app.MetadataKeeper.IterateScopes(s.ctx, func(s types.Scope) (stop bool) {
		count++
		return count >= 5
	})
	s.Equal(5, count, "using iterator stop function should stop iterator early")
}

func (s *ScopeKeeperTestSuite) TestValidateScopeUpdate() {
	markerAddr := markertypes.MustGetMarkerAddress("testcoin").String()
	err := s.app.MarkerKeeper.AddMarkerAccount(s.ctx, &markertypes.MarkerAccount{
		BaseAccount: &authtypes.BaseAccount{
			Address:       markerAddr,
			AccountNumber: 23,
		},
		AccessControl: []markertypes.AccessGrant{
			{
				Address:     s.user1,
				Permissions: markertypes.AccessListByNames("deposit,withdraw"),
			},
		},
		Denom:      "testcoin",
		Supply:     sdk.NewInt(1000),
		MarkerType: markertypes.MarkerType_Coin,
		Status:     markertypes.StatusActive,
	})
	s.NoError(err)

	scopeSpecID := types.ScopeSpecMetadataAddress(uuid.New())
	scopeSpec := types.NewScopeSpecification(scopeSpecID, nil, []string{s.user1}, []types.PartyType{types.PartyType_PARTY_TYPE_OWNER}, []types.MetadataAddress{})
	s.app.MetadataKeeper.SetScopeSpecification(s.ctx, *scopeSpec)

	scopeID := types.ScopeMetadataAddress(uuid.New())
	scopeID2 := types.ScopeMetadataAddress(uuid.New())

	// Give user 3 authority to sign for user 1 for scope updates.
	now := s.ctx.BlockHeader().Time
	s.Require().NotNil(now)
	granter := s.user1Addr
	grantee := s.user3Addr
	a := authz.NewGenericAuthorization(types.TypeURLMsgWriteScopeRequest)
	s.Require().NoError(s.app.AuthzKeeper.SaveGrant(s.ctx, grantee, granter, a, now.Add(time.Hour)), "authz SaveGrant user1 to user3")

	cases := []struct {
		name     string
		existing types.Scope
		proposed types.Scope
		signers  []string
		errorMsg string
	}{
		{
			name:     "nil previous, proposed throws address error",
			existing: types.Scope{},
			proposed: types.Scope{},
			signers:  []string{s.user1},
			errorMsg: "address is empty",
		},
		{
			name:     "valid proposed with nil existing doesn't error",
			existing: types.Scope{},
			proposed: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "can't change scope id in update",
			existing: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *types.NewScope(scopeID2, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("cannot update scope identifier. expected %s, got %s", scopeID.String(), scopeID2.String()),
		},
		{
			name:     "missing existing owner signer on update fails",
			existing: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, ""),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature from [%s (PARTY_TYPE_OWNER)]", s.user1),
		},
		{
			name:     "missing existing owner signer on update fails",
			existing: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user2), []string{}, ""),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature from [%s (PARTY_TYPE_OWNER)]", s.user1),
		},
		{
			name:     "no error when update includes existing owner signer",
			existing: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, ""),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "no error when there are no updates regardless of signatures",
			existing: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			signers:  []string{},
			errorMsg: "",
		},
		{
			name:     "setting value owner when unset does not error",
			existing: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "setting value owner when unset requires current owner signature",
			existing: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{},
			errorMsg: fmt.Sprintf("missing signature from [%s (PARTY_TYPE_OWNER)]", s.user1),
		},
		{
			name:     "setting value owner to user does not require their signature",
			existing: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "setting value owner to new user does not require their signature",
			existing: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "no change to value owner should not error",
			existing: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "setting a new value owner should not error with withdraw permission",
			existing: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, markerAddr),
			proposed: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "setting a new value owner fails if missing withdraw permission",
			existing: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user2), []string{}, markerAddr),
			proposed: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user2), []string{}, s.user2),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature for %s with authority to withdraw/remove existing value owner", markerAddr),
		},
		{
			name:     "setting a new value owner fails if missing deposit permission",
			existing: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user2), []string{}, ""),
			proposed: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user2), []string{}, markerAddr),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("no signatures present with authority to add scope to marker %s", markerAddr),
		},
		{
			name:     "setting a new value owner fails for scope owner when value owner signature is missing",
			existing: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2),
			proposed: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("missing signature from existing value owner %s", s.user2),
		},
		{
			name:     "unsetting all fields on a scope should be successful",
			existing: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: types.Scope{ScopeId: scopeID, SpecificationId: scopeSpecID, Owners: ownerPartyList(s.user1)},
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "setting specification id to nil should fail",
			existing: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *types.NewScope(scopeID, nil, ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			errorMsg: "invalid specification id: address is empty",
		},
		{
			name:     "setting unknown specification id should fail",
			existing: *types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *types.NewScope(scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("scope specification %s not found", types.ScopeSpecMetadataAddress(s.scopeUUID)),
		},
		{
			name:     "adding data access with authz grant should be successful",
			existing: types.Scope{scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1},
			proposed: types.Scope{scopeID, scopeSpecID, ownerPartyList(s.user1), []string{s.user2}, s.user1},
			signers:  []string{s.user3}, // user 1 has granted scope-write to user 3
			errorMsg: "",
		},
		{
			name:     "multi owner adding data access with authz grant should be successful",
			existing: types.Scope{scopeID, scopeSpecID, ownerPartyList(s.user1, s.user2), []string{}, s.user1},
			proposed: types.Scope{scopeID, scopeSpecID, ownerPartyList(s.user1, s.user2), []string{s.user2}, s.user1},
			signers:  []string{s.user2, s.user3}, // user 1 has granted scope-write to user 3
			errorMsg: "",
		},
		{
			name:     "changing value owner with authz grant should be successful",
			existing: types.Scope{scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1},
			proposed: types.Scope{scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2},
			signers:  []string{s.user3}, // user 1 has granted scope-write to user 3
			errorMsg: "",
		},
		{
			name:     "changing value owner by authz granter should be successful",
			existing: types.Scope{scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1},
			proposed: types.Scope{scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2},
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "changing value owner by non-authz grantee should fail",
			existing: types.Scope{scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1},
			proposed: types.Scope{scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2},
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature from existing value owner %s", s.user1),
		},
		{
			name:     "changing value owner from non-authz granter with different signer should fail",
			existing: types.Scope{scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2},
			proposed: types.Scope{scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1},
			signers:  []string{s.user3},
			errorMsg: fmt.Sprintf("missing signature from existing value owner %s", s.user2),
		},
		{
			name:     "setting value owner from nothing to non-owner only signed by non-owner should fail",
			existing: types.Scope{scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""},
			proposed: types.Scope{scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2},
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature from [%s (PARTY_TYPE_OWNER)]", s.user1),
		},
	}

	for _, tc := range cases {
		s.T().Run(tc.name, func(t *testing.T) {
			err = s.app.MetadataKeeper.ValidateScopeUpdate(s.ctx, tc.existing, tc.proposed, tc.signers, types.TypeURLMsgWriteScopeRequest)
			if len(tc.errorMsg) > 0 {
				assert.EqualError(t, err, tc.errorMsg, "ValidateScopeUpdate expected error")
			} else {
				assert.NoError(t, err, "ValidateScopeUpdate unexpected error")
			}
		})
	}
}

func (s ScopeKeeperTestSuite) TestValidateScopeRemove() {
	markerDenom := "testcoins2"
	markerAddr := markertypes.MustGetMarkerAddress(markerDenom).String()
	err := s.app.MarkerKeeper.AddMarkerAccount(s.ctx, &markertypes.MarkerAccount{
		BaseAccount: &authtypes.BaseAccount{
			Address:       markerAddr,
			AccountNumber: 24,
		},
		AccessControl: []markertypes.AccessGrant{
			{
				Address:     s.user1,
				Permissions: markertypes.AccessListByNames("deposit,withdraw"),
			},
		},
		Denom:      markerDenom,
		Supply:     sdk.NewInt(1000),
		MarkerType: markertypes.MarkerType_Coin,
		Status:     markertypes.StatusActive,
	})
	s.NoError(err)

	scopeNoValueOwner := types.Scope{
		ScopeId:           types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
		Owners:            ownerPartyList(s.user1, s.user2),
		DataAccess:        nil,
		ValueOwnerAddress: "",
	}

	scopeMarkerValueOwner := types.Scope{
		ScopeId:           types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
		Owners:            ownerPartyList(s.user2),
		DataAccess:        nil,
		ValueOwnerAddress: markerAddr,
	}

	scopeUserValueOwner := types.Scope{
		ScopeId:           types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
		Owners:            ownerPartyList(s.user2),
		DataAccess:        nil,
		ValueOwnerAddress: s.user1,
	}

	missing1Sig := func(addr string) string {
		return fmt.Sprintf("missing signature from [%s (PARTY_TYPE_OWNER)]", addr)
	}

	missing2Sigs := func(addr1, addr2 string) string {
		return fmt.Sprintf("missing signatures from [%s (PARTY_TYPE_OWNER) %s (PARTY_TYPE_OWNER)]", addr1, addr2)
	}

	tests := []struct {
		name     string
		scope    types.Scope
		signers  []string
		expected string
	}{
		{
			name:     "no value owner all signers",
			scope:    scopeNoValueOwner,
			signers:  []string{s.user1, s.user2},
			expected: "",
		},
		{
			name:     "no value owner all signers reversed",
			scope:    scopeNoValueOwner,
			signers:  []string{s.user1, s.user2},
			expected: "",
		},
		{
			name:     "no value owner extra signer",
			scope:    scopeNoValueOwner,
			signers:  []string{s.user1, s.user2, s.user3},
			expected: "",
		},
		{
			name:     "no value owner missing signer 1",
			scope:    scopeNoValueOwner,
			signers:  []string{s.user2},
			expected: missing1Sig(s.user1),
		},
		{
			name:     "no value owner missing signer 2",
			scope:    scopeNoValueOwner,
			signers:  []string{s.user1},
			expected: missing1Sig(s.user2),
		},
		{
			name:     "no value owner no signers",
			scope:    scopeNoValueOwner,
			signers:  []string{},
			expected: missing2Sigs(s.user1, s.user2),
		},
		{
			name:     "no value owner wrong signer",
			scope:    scopeNoValueOwner,
			signers:  []string{s.user3},
			expected: missing2Sigs(s.user1, s.user2),
		},
		{
			name:     "marker value owner signed by owner and user with auth",
			scope:    scopeMarkerValueOwner,
			signers:  []string{s.user1, s.user2},
			expected: "",
		},
		{
			name:     "marker value owner signed by owner and user with auth reversed",
			scope:    scopeMarkerValueOwner,
			signers:  []string{s.user2, s.user1},
			expected: "",
		},
		{
			name:     "marker value owner not signed by owner",
			scope:    scopeMarkerValueOwner,
			signers:  []string{s.user1},
			expected: missing1Sig(s.user2),
		},
		{
			name:     "marker value owner not signed by user with auth",
			scope:    scopeMarkerValueOwner,
			signers:  []string{s.user2},
			expected: fmt.Sprintf("missing signature for %s with authority to withdraw/remove existing value owner", markerAddr),
		},
		{
			name:     "user value owner signed by owner and value owner",
			scope:    scopeUserValueOwner,
			signers:  []string{s.user1, s.user2},
			expected: "",
		},
		{
			name:     "user value owner signed by owner and value owner reversed",
			scope:    scopeUserValueOwner,
			signers:  []string{s.user2, s.user1},
			expected: "",
		},
		{
			name:     "user value owner not signed by owner",
			scope:    scopeUserValueOwner,
			signers:  []string{s.user1},
			expected: missing1Sig(s.user2),
		},
		{
			name:     "user value owner not signed by value owner",
			scope:    scopeUserValueOwner,
			signers:  []string{s.user2},
			expected: fmt.Sprintf("missing signature from existing value owner %s", s.user1),
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			actual := s.app.MetadataKeeper.ValidateScopeRemove(s.ctx, tc.scope, tc.signers, types.TypeURLMsgDeleteScopeRequest)
			if len(tc.expected) > 0 {
				require.EqualError(t, actual, tc.expected)
			} else {
				require.NoError(t, actual)
			}
		})
	}
}

func (s *ScopeKeeperTestSuite) TestValidateScopeAddDataAccess() {
	scope := *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), ownerPartyList(s.user1), []string{s.user1}, s.user1)

	cases := map[string]struct {
		dataAccessAddrs []string
		existing        types.Scope
		signers         []string
		wantErr         bool
		errorMsg        string
	}{
		"should fail to validate add scope data access, does not have any users": {
			[]string{},
			scope,
			[]string{s.user1},
			true,
			"data access list cannot be empty",
		},
		"should fail to validate add scope data access, user is already on the data access list": {
			[]string{s.user1},
			scope,
			[]string{s.user1},
			true,
			fmt.Sprintf("address already exists for data access %s", s.user1),
		},
		"should fail to validate add scope data access, incorrect signer for scope": {
			[]string{s.user2},
			scope,
			[]string{s.user2},
			true,
			fmt.Sprintf("missing signature from [%s (PARTY_TYPE_OWNER)]", s.user1),
		},
		"should fail to validate add scope data access, incorrect address type": {
			[]string{"invalidaddr"},
			scope,
			[]string{s.user1},
			true,
			"failed to decode data access address invalidaddr : decoding bech32 failed: invalid separator index -1",
		},
		"should successfully validate add scope data access": {
			[]string{s.user2},
			scope,
			[]string{s.user1},
			false,
			"",
		},
	}

	for n, tc := range cases {
		tc := tc

		s.Run(n, func() {
			err := s.app.MetadataKeeper.ValidateScopeAddDataAccess(s.ctx, tc.dataAccessAddrs, tc.existing, tc.signers, types.TypeURLMsgAddScopeDataAccessRequest)
			if tc.wantErr {
				s.Error(err)
				s.Equal(tc.errorMsg, err.Error())
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *ScopeKeeperTestSuite) TestValidateScopeDeleteDataAccess() {
	scope := *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), ownerPartyList(s.user1), []string{s.user1, s.user2}, s.user1)

	cases := map[string]struct {
		dataAccessAddrs []string
		existing        types.Scope
		signers         []string
		wantErr         bool
		errorMsg        string
	}{
		"should fail to validate delete scope data access, does not have any users": {
			[]string{},
			scope,
			[]string{s.user1},
			true,
			"data access list cannot be empty",
		},
		"should fail to validate delete scope data access, address is not in data access list": {
			[]string{s.user2, s.user3},
			scope,
			[]string{s.user1},
			true,
			fmt.Sprintf("address does not exist in scope data access: %s", s.user3),
		},
		"should fail to validate delete scope data access, incorrect signer for scope": {
			[]string{s.user2},
			scope,
			[]string{s.user2},
			true,
			fmt.Sprintf("missing signature from [%s (PARTY_TYPE_OWNER)]", s.user1),
		},
		"should fail to validate delete scope data access, incorrect address type": {
			[]string{"invalidaddr"},
			scope,
			[]string{s.user1},
			true,
			"failed to decode data access address invalidaddr : decoding bech32 failed: invalid separator index -1",
		},
		"should successfully validate delete scope data access": {
			[]string{s.user1, s.user2},
			scope,
			[]string{s.user1},
			false,
			"",
		},
	}

	for n, tc := range cases {
		tc := tc

		s.Run(n, func() {
			err := s.app.MetadataKeeper.ValidateScopeDeleteDataAccess(s.ctx, tc.dataAccessAddrs, tc.existing, tc.signers, types.TypeURLMsgDeleteScopeDataAccessRequest)
			if tc.wantErr {
				s.Error(err)
				s.Equal(tc.errorMsg, err.Error())
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *ScopeKeeperTestSuite) TestValidateScopeUpdateOwners() {
	scopeSpecID := types.ScopeSpecMetadataAddress(uuid.New())
	scopeSpec := types.NewScopeSpecification(scopeSpecID, nil, []string{s.user1}, []types.PartyType{types.PartyType_PARTY_TYPE_OWNER}, []types.MetadataAddress{})
	s.app.MetadataKeeper.SetScopeSpecification(s.ctx, *scopeSpec)

	scopeWithOwners := func(owners []types.Party) types.Scope {
		return *types.NewScope(s.scopeID, scopeSpecID, owners, []string{s.user1}, s.user1)
	}
	originalOwners := ownerPartyList(s.user1)

	testCases := []struct {
		name     string
		existing types.Scope
		proposed types.Scope
		signers  []string
		errorMsg string
	}{
		{
			"should fail to validate update scope owners, fail to decode address",
			scopeWithOwners(originalOwners),
			scopeWithOwners([]types.Party{{Address: "shoulderror", Role: types.PartyType_PARTY_TYPE_AFFILIATE}}),
			[]string{s.user1},
			fmt.Sprintf("invalid scope owners: invalid party address [%s]: %s", "shoulderror", "decoding bech32 failed: invalid separator index -1"),
		},
		{
			"should fail to validate update scope owners, role cannot be unspecified",
			scopeWithOwners(originalOwners),
			scopeWithOwners([]types.Party{{Address: s.user1, Role: types.PartyType_PARTY_TYPE_UNSPECIFIED}}),
			[]string{s.user1},
			fmt.Sprintf("invalid scope owners: invalid party type for party %s", s.user1),
		},
		{
			"should fail to validate update scope owner, wrong signer new owner",
			scopeWithOwners(originalOwners),
			scopeWithOwners([]types.Party{{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER}}),
			[]string{s.user2},
			fmt.Sprintf("missing signature from [%s (PARTY_TYPE_OWNER)]", s.user1),
		},
		{
			"should successfully validate update scope owner, same owner different role",
			scopeWithOwners(ownerPartyList(s.user1, s.user2)),
			scopeWithOwners([]types.Party{
				{Address: s.user1, Role: types.PartyType_PARTY_TYPE_CUSTODIAN},
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
			}),
			[]string{s.user1, s.user2},
			"",
		},
		{
			"should successfully validate update scope owner, new owner",
			scopeWithOwners(originalOwners),
			scopeWithOwners([]types.Party{{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER}}),
			[]string{s.user1},
			"",
		},
		{
			"should fail to validate update scope owner, missing role",
			scopeWithOwners(originalOwners),
			scopeWithOwners([]types.Party{{Address: s.user1, Role: types.PartyType_PARTY_TYPE_CUSTODIAN}}),
			[]string{s.user1},
			"missing party type required by spec: [OWNER]",
		},
		{
			"should fail to validate update scope owner, empty list",
			scopeWithOwners(originalOwners),
			scopeWithOwners([]types.Party{}),
			[]string{s.user1},
			"invalid scope owners: at least one party is required",
		},
		{
			"should successfully validate update scope owner, 1st owner removed",
			scopeWithOwners(ownerPartyList(s.user1, s.user2)),
			scopeWithOwners(ownerPartyList(s.user2)),
			[]string{s.user1, s.user2},
			"",
		},
		{
			"should successfully validate update scope owner, 2nd owner removed",
			scopeWithOwners(ownerPartyList(s.user1, s.user2)),
			scopeWithOwners(ownerPartyList(s.user1)),
			[]string{s.user1, s.user2},
			"",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			err := s.app.MetadataKeeper.ValidateScopeUpdateOwners(s.ctx, tc.existing, tc.proposed, tc.signers, types.TypeURLMsgAddScopeOwnerRequest)
			if len(tc.errorMsg) > 0 {
				assert.EqualError(t, err, tc.errorMsg, "ValidateScopeUpdateOwners expected error")
			} else {
				assert.NoError(t, err, "ValidateScopeUpdateOwners unexpected error")
			}
		})
	}
}

func (s *ScopeKeeperTestSuite) TestScopeIndexing() {
	scopeID := types.ScopeMetadataAddress(uuid.New())

	specIDOrig := types.ScopeSpecMetadataAddress(uuid.New())
	specIDNew := types.ScopeSpecMetadataAddress(uuid.New())

	ownerConstant := randomUser()
	ownerToAdd := randomUser()
	ownerToRemove := randomUser()
	valueOwnerOrig := randomUser()
	valueOwnerNew := randomUser()

	scopeV1 := types.Scope{
		ScopeId:           scopeID,
		SpecificationId:   specIDOrig,
		Owners:            ownerPartyList(ownerConstant.Bech32, ownerToRemove.Bech32),
		DataAccess:        nil,
		ValueOwnerAddress: valueOwnerOrig.Bech32,
	}
	scopeV2 := types.Scope{
		ScopeId:           scopeID,
		SpecificationId:   specIDNew,
		Owners:            ownerPartyList(ownerConstant.Bech32, ownerToAdd.Bech32),
		DataAccess:        nil,
		ValueOwnerAddress: valueOwnerNew.Bech32,
	}

	store := s.ctx.KVStore(s.app.GetKey(types.ModuleName))

	s.T().Run("1 write new scope", func(t *testing.T) {
		expectedIndexes := []struct {
			key  []byte
			name string
		}{
			{types.GetAddressScopeCacheKey(ownerConstant.Addr, scopeID), "ownerConstant address index"},
			{types.GetAddressScopeCacheKey(ownerToRemove.Addr, scopeID), "ownerToRemove address index"},
			{types.GetAddressScopeCacheKey(valueOwnerOrig.Addr, scopeID), "valueOwnerOrig address index"},

			{types.GetValueOwnerScopeCacheKey(valueOwnerOrig.Addr, scopeID), "valueOwnerOrig value owner index"},

			{types.GetScopeSpecScopeCacheKey(specIDOrig, scopeID), "specIDOrig spec index"},
		}

		s.app.MetadataKeeper.SetScope(s.ctx, scopeV1)

		for _, expected := range expectedIndexes {
			assert.True(t, store.Has(expected.key), expected.name)
		}
	})

	s.T().Run("2 update scope", func(t *testing.T) {
		expectedIndexes := []struct {
			key  []byte
			name string
		}{
			{types.GetAddressScopeCacheKey(ownerConstant.Addr, scopeID), "ownerConstant address index"},
			{types.GetAddressScopeCacheKey(ownerToAdd.Addr, scopeID), "ownerToAdd address index"},
			{types.GetAddressScopeCacheKey(valueOwnerNew.Addr, scopeID), "valueOwnerNew address index"},

			{types.GetValueOwnerScopeCacheKey(valueOwnerNew.Addr, scopeID), "valueOwnerNew value owner index"},

			{types.GetScopeSpecScopeCacheKey(specIDNew, scopeID), "specIDNew spec index"},
		}
		unexpectedIndexes := []struct {
			key  []byte
			name string
		}{
			{types.GetAddressScopeCacheKey(ownerToRemove.Addr, scopeID), "ownerToRemove address index"},
			{types.GetAddressScopeCacheKey(valueOwnerOrig.Addr, scopeID), "valueOwnerOrig address index"},

			{types.GetValueOwnerScopeCacheKey(valueOwnerOrig.Addr, scopeID), "valueOwnerOrig value owner index"},

			{types.GetScopeSpecScopeCacheKey(specIDOrig, scopeID), "specIDOrig spec index"},
		}

		s.app.MetadataKeeper.SetScope(s.ctx, scopeV2)

		for _, expected := range expectedIndexes {
			assert.True(t, store.Has(expected.key), expected.name)
		}
		for _, unexpected := range unexpectedIndexes {
			assert.False(t, store.Has(unexpected.key), unexpected.name)
		}
	})

	s.T().Run("3 delete scope", func(t *testing.T) {
		unexpectedIndexes := []struct {
			key  []byte
			name string
		}{
			{types.GetAddressScopeCacheKey(ownerConstant.Addr, scopeID), "ownerConstant address index"},
			{types.GetAddressScopeCacheKey(ownerToRemove.Addr, scopeID), "ownerToRemove address index"},
			{types.GetAddressScopeCacheKey(ownerToAdd.Addr, scopeID), "ownerToAdd address index"},
			{types.GetAddressScopeCacheKey(valueOwnerOrig.Addr, scopeID), "valueOwnerOrig address index"},
			{types.GetAddressScopeCacheKey(valueOwnerNew.Addr, scopeID), "valueOwnerNew address index"},

			{types.GetValueOwnerScopeCacheKey(valueOwnerOrig.Addr, scopeID), "valueOwnerOrig value owner index"},
			{types.GetValueOwnerScopeCacheKey(valueOwnerNew.Addr, scopeID), "valueOwnerNew value owner index"},

			{types.GetScopeSpecScopeCacheKey(specIDOrig, scopeID), "specIDOrig spec index"},
			{types.GetScopeSpecScopeCacheKey(specIDNew, scopeID), "specIDNew spec index"},
		}

		s.app.MetadataKeeper.RemoveScope(s.ctx, scopeID)

		for _, unexpected := range unexpectedIndexes {
			assert.False(t, store.Has(unexpected.key), unexpected.name)
		}
	})
}
