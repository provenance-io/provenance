package keeper_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	simapp "github.com/provenance-io/provenance/app"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	"github.com/provenance-io/provenance/x/metadata/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/stretchr/testify/suite"
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
	changedID := types.ScopeMetadataAddress(uuid.New())

	cases := map[string]struct {
		existing types.Scope
		proposed types.Scope
		signers  []string
		wantErr  bool
		errorMsg string
	}{
		"nil previous, proposed throws address error": {
			existing: types.Scope{},
			proposed: types.Scope{},
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: "address is empty",
		},
		"valid proposed with nil existing doesn't error": {
			existing: types.Scope{},
			proposed: *types.NewScope(s.scopeID, nil, ownerPartyList(s.user1), []string{}, ""),
			signers:  []string{s.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"can't change scope id in update": {
			existing: *types.NewScope(s.scopeID, nil, ownerPartyList(s.user1), []string{}, ""),
			proposed: *types.NewScope(changedID, nil, ownerPartyList(s.user1), []string{}, ""),
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: fmt.Sprintf("cannot update scope identifier. expected %s, got %s", s.scopeID.String(), changedID.String()),
		},
		"missing existing owner signer on update fails": {
			existing: *types.NewScope(s.scopeID, nil, ownerPartyList(s.user1), []string{}, ""),
			proposed: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), ownerPartyList(s.user1), []string{}, ""),
			signers:  []string{s.user2},
			wantErr:  true,
			errorMsg: fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1),
		},
		"no error when update includes existing owner signer": {
			existing: *types.NewScope(s.scopeID, nil, ownerPartyList(s.user1), []string{}, ""),
			proposed: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), ownerPartyList(s.user1), []string{}, ""),
			signers:  []string{s.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"setting value owner when unset does not error": {
			existing: *types.NewScope(s.scopeID, nil, ownerPartyList(s.user1), []string{}, ""),
			proposed: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"setting value owner to user does not require their signature": {
			existing: *types.NewScope(s.scopeID, nil, ownerPartyList(s.user1), []string{}, ""),
			proposed: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), ownerPartyList(s.user1), []string{}, s.user2),
			signers:  []string{s.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"setting value owner to new user does not require their signature": {
			existing: *types.NewScope(s.scopeID, nil, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), ownerPartyList(s.user1), []string{}, s.user2),
			signers:  []string{s.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"no change to value owner should not error": {
			existing: *types.NewScope(s.scopeID, nil, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"setting a new value owner should not error with withdraw permission": {
			existing: *types.NewScope(s.scopeID, nil, ownerPartyList(s.user1), []string{}, markerAddr),
			proposed: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"setting a new value owner fails if missing withdraw permission": {
			existing: *types.NewScope(s.scopeID, nil, ownerPartyList(s.user2), []string{}, markerAddr),
			proposed: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), ownerPartyList(s.user2), []string{}, s.user2),
			signers:  []string{s.user2},
			wantErr:  true,
			errorMsg: fmt.Sprintf("missing signature for %s with authority to withdraw/remove existing value owner", markerAddr),
		},
		"setting a new value owner fails if missing deposit permission": {
			existing: *types.NewScope(s.scopeID, nil, ownerPartyList(s.user2), []string{}, ""),
			proposed: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), ownerPartyList(s.user2), []string{}, markerAddr),
			signers:  []string{s.user2},
			wantErr:  true,
			errorMsg: fmt.Sprintf("no signatures present with authority to add scope to marker %s", markerAddr),
		},
		"setting a new value owner fails for scope owner when value owner signature is missing": {
			existing: *types.NewScope(s.scopeID, nil, ownerPartyList(s.user1), []string{}, s.user2),
			proposed: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: fmt.Sprintf("missing signature from existing owner %s; required for update", s.user2),
		},
		"unsetting all fields on a scope should be successful": {
			existing: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), ownerPartyList(s.user1), []string{}, s.user1),
			proposed: types.Scope{ScopeId: s.scopeID, Owners: ownerPartyList(s.user1)},
			signers:  []string{s.user1},
			wantErr:  false,
			errorMsg: "",
		},
	}

	for n, tc := range cases {
		tc := tc

		s.Run(n, func() {
			err = s.app.MetadataKeeper.ValidateScopeUpdate(s.ctx, tc.existing, tc.proposed, tc.signers)
			if tc.wantErr {
				s.Error(err)
				s.Equal(tc.errorMsg, err.Error())
			} else {
				s.NoError(err)
			}
		})
	}
}

// TODO: ValidateScopeRemove tests

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
			"failed to decode data access address invalidaddr : decoding bech32 failed: invalid index of 1",
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
			err := s.app.MetadataKeeper.ValidateScopeAddDataAccess(s.ctx, tc.dataAccessAddrs, tc.existing, tc.signers)
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
			"failed to decode data access address invalidaddr : decoding bech32 failed: invalid index of 1",
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
			err := s.app.MetadataKeeper.ValidateScopeDeleteDataAccess(s.ctx, tc.dataAccessAddrs, tc.existing, tc.signers)
			if tc.wantErr {
				s.Error(err)
				s.Equal(tc.errorMsg, err.Error())
			} else {
				s.NoError(err)
			}
		})
	}
}
