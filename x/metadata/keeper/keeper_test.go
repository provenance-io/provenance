package keeper_test

import (
	gocontext "context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/suite"

	markertypes "github.com/provenance-io/provenance/x/marker/types"
	"github.com/provenance-io/provenance/x/metadata/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app         *app.App
	ctx         sdk.Context
	queryClient types.QueryClient

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress

	scopeUUID uuid.UUID
	scopeID   types.MetadataAddress

	specUUID uuid.UUID
	specID   types.MetadataAddress
}

func (s *KeeperTestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	s.scopeUUID = uuid.New()
	s.scopeID = types.ScopeMetadataAddress(s.scopeUUID)

	s.specUUID = uuid.New()
	s.specID = types.ScopeSpecMetadataAddress(s.specUUID)

	s.app = app
	s.ctx = ctx

	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.MetadataKeeper)
	queryClient := types.NewQueryClient(queryHelper)
	s.queryClient = queryClient
}

func (s *KeeperTestSuite) TestMetadataScopeGetSet() {
	scope, found := s.app.MetadataKeeper.GetScope(s.ctx, s.scopeID)
	s.NotNil(scope)
	s.False(found)

	ns := *types.NewScope(s.scopeID, s.specID, []string{s.user1}, []string{s.user1}, s.user1)
	s.NotNil(ns)
	s.app.MetadataKeeper.SetScope(s.ctx, ns)

	scope, found = s.app.MetadataKeeper.GetScope(s.ctx, s.scopeID)
	s.True(found)
	s.NotNil(scope)

	s.app.MetadataKeeper.DeleteScope(s.ctx, ns.ScopeId)
	scope, found = s.app.MetadataKeeper.GetScope(s.ctx, s.scopeID)
	s.False(found)
	s.NotNil(scope)
}

func (s *KeeperTestSuite) TestMetadataScopeIterator() {
	for i := 1; i <= 10; i++ {
		valueOwner := ""
		if i == 5 {
			valueOwner = s.user2
		}
		ns := types.NewScope(types.ScopeMetadataAddress(uuid.New()), nil, []string{s.user1}, []string{s.user1}, valueOwner)
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

func (s *KeeperTestSuite) TestValidateScopeUpdate() {
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
			errorMsg: "incorrect address length (must be at least 17, actual: 0)",
		},
		"valid proposed with nil existing doesn't error": {
			existing: types.Scope{},
			proposed: *types.NewScope(s.scopeID, nil, []string{s.user1}, []string{}, ""),
			signers:  []string{s.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"can't change scope id in update": {
			existing: *types.NewScope(s.scopeID, nil, []string{s.user1}, []string{}, ""),
			proposed: *types.NewScope(changedID, nil, []string{s.user1}, []string{}, ""),
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: fmt.Sprintf("cannot update scope identifier. expected %s, got %s", s.scopeID.String(), changedID.String()),
		},
		"missing existing owner signer on update fails": {
			existing: *types.NewScope(s.scopeID, nil, []string{s.user1}, []string{}, ""),
			proposed: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), []string{s.user1}, []string{}, ""),
			signers:  []string{s.user2},
			wantErr:  true,
			errorMsg: fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1),
		},
		"no error when update includes existing owner signer": {
			existing: *types.NewScope(s.scopeID, nil, []string{s.user1}, []string{}, ""),
			proposed: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), []string{s.user1}, []string{}, ""),
			signers:  []string{s.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"setting value owner when unset does not error": {
			existing: *types.NewScope(s.scopeID, nil, []string{s.user1}, []string{}, ""),
			proposed: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), []string{s.user1}, []string{}, s.user1),
			signers:  []string{s.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"setting value owner to user does not require their signature": {
			existing: *types.NewScope(s.scopeID, nil, []string{s.user1}, []string{}, ""),
			proposed: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), []string{s.user1}, []string{}, s.user2),
			signers:  []string{s.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"setting value owner to new user does not require their signature": {
			existing: *types.NewScope(s.scopeID, nil, []string{s.user1}, []string{}, s.user1),
			proposed: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), []string{s.user1}, []string{}, s.user2),
			signers:  []string{s.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"no change to value owner should not error": {
			existing: *types.NewScope(s.scopeID, nil, []string{s.user1}, []string{}, s.user1),
			proposed: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), []string{s.user1}, []string{}, s.user1),
			signers:  []string{s.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"setting a new value owner should not error with withdraw permission": {
			existing: *types.NewScope(s.scopeID, nil, []string{s.user1}, []string{}, markerAddr),
			proposed: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), []string{s.user1}, []string{}, s.user1),
			signers:  []string{s.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"setting a new value owner fails if missing withdraw permission": {
			existing: *types.NewScope(s.scopeID, nil, []string{s.user2}, []string{}, markerAddr),
			proposed: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), []string{s.user2}, []string{}, s.user2),
			signers:  []string{s.user2},
			wantErr:  true,
			errorMsg: fmt.Sprintf("missing signature for %s with authority to withdraw/remove existing value owner", markerAddr),
		},
		"setting a new value owner fails if missing deposit permission": {
			existing: *types.NewScope(s.scopeID, nil, []string{s.user2}, []string{}, ""),
			proposed: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), []string{s.user2}, []string{}, markerAddr),
			signers:  []string{s.user2},
			wantErr:  true,
			errorMsg: fmt.Sprintf("no signatures present with authority to add scope to marker %s", markerAddr),
		},
		"setting a new value owner fails for scope owner when value owner signature is missing": {
			existing: *types.NewScope(s.scopeID, nil, []string{s.user1}, []string{}, s.user2),
			proposed: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), []string{s.user1}, []string{}, s.user1),
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: fmt.Sprintf("missing signature from existing owner %s; required for update", s.user2),
		},
		"unsetting all fields on a scope should be successful": {
			existing: *types.NewScope(s.scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), []string{s.user1}, []string{}, s.user1),
			proposed: types.Scope{ScopeId: s.scopeID, OwnerAddress: []string{s.user1}},
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

func (s *KeeperTestSuite) TestScopeQuery() {
	app, ctx, queryClient, user1, user2 := s.app, s.ctx, s.queryClient, s.user1, s.user2

	testIDs := make([]types.MetadataAddress, 10)
	for i := 0; i < 10; i++ {
		valueOwner := ""
		if i == 5 {
			valueOwner = user2
		}
		testIDs[i] = types.ScopeMetadataAddress(uuid.New())
		ns := types.NewScope(testIDs[i], nil, []string{user1}, []string{user1}, valueOwner)
		app.MetadataKeeper.SetScope(ctx, *ns)
	}
	uuid, err := testIDs[0].ScopeUUID()
	s.NoError(err)

	_, err = queryClient.Scope(gocontext.Background(), &types.ScopeRequest{})
	s.EqualError(err, "rpc error: code = InvalidArgument desc = scope id cannot be empty")

	_, err = queryClient.Scope(gocontext.Background(), &types.ScopeRequest{ScopeId: "6332c1a4-foo1-bare-895b-invalid65cb6"})
	s.EqualError(err, "rpc error: code = InvalidArgument desc = invalid scope id: invalid UUID format")

	scopeResponse, err := queryClient.Scope(gocontext.Background(), &types.ScopeRequest{ScopeId: uuid.String()})
	s.NoError(err)
	s.NotNil(scopeResponse.Scope)
	s.Equal(testIDs[0], scopeResponse.Scope.ScopeId)

	// TODO assert records when available
	// TODO assert record groups when available

	// only one scope has value owner set (user2)
	valueResponse, err := queryClient.ValueOwnership(gocontext.Background(), &types.ValueOwnershipRequest{Address: user2})
	s.NoError(err)
	s.Len(valueResponse.ScopeIds, 1)

	// 10 entries as all scopes have user1 as data_owner
	ownerResponse, err := queryClient.Ownership(gocontext.Background(), &types.OwnershipRequest{Address: user1})
	s.NoError(err)
	s.Len(ownerResponse.ScopeIds, 10)

	// one entry for user2 (as value owner)
	ownerResponse, err = queryClient.Ownership(gocontext.Background(), &types.OwnershipRequest{Address: user2})
	s.NoError(err)
	s.Len(ownerResponse.ScopeIds, 1)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
