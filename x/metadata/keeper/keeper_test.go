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

	groupUUID uuid.UUID
	groupId   types.MetadataAddress

	record     types.Record
	recordName string
	recordId   types.MetadataAddress
}

func (suite *KeeperTestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	suite.pubkey1 = secp256k1.GenPrivKey().PubKey()
	suite.user1Addr = sdk.AccAddress(suite.pubkey1.Address())
	suite.user1 = suite.user1Addr.String()

	suite.pubkey2 = secp256k1.GenPrivKey().PubKey()
	suite.user2Addr = sdk.AccAddress(suite.pubkey2.Address())
	suite.user2 = suite.user2Addr.String()

	suite.scopeUUID = uuid.New()
	suite.scopeID = types.ScopeMetadataAddress(suite.scopeUUID)

	suite.specUUID = uuid.New()
	suite.specID = types.ScopeSpecMetadataAddress(suite.specUUID)

	suite.app = app
	suite.ctx = ctx

	suite.groupUUID = uuid.New()
	suite.groupId = types.GroupMetadataAddress(suite.scopeUUID, suite.groupUUID)

	suite.recordName = "TestRecord"
	suite.recordId = types.RecordMetadataAddress(suite.scopeUUID, suite.recordName)

	suite.app.AccountKeeper.SetAccount(suite.ctx, suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, suite.user1Addr))

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.MetadataKeeper)
	queryClient := types.NewQueryClient(queryHelper)
	suite.queryClient = queryClient
}

func (suite *KeeperTestSuite) TestMetadataScopeGetSet() {
	s, found := suite.app.MetadataKeeper.GetScope(suite.ctx, suite.scopeID)
	suite.NotNil(s)
	suite.False(found)

	ns := *types.NewScope(suite.scopeID, suite.specID, []string{suite.user1}, []string{suite.user1}, suite.user1)
	suite.NotNil(ns)
	suite.app.MetadataKeeper.SetScope(suite.ctx, ns)

	s, found = suite.app.MetadataKeeper.GetScope(suite.ctx, suite.scopeID)
	suite.True(found)
	suite.NotNil(s)

	suite.app.MetadataKeeper.RemoveScope(suite.ctx, ns.ScopeId)
	s, found = suite.app.MetadataKeeper.GetScope(suite.ctx, suite.scopeID)
	suite.False(found)
	suite.NotNil(s)
}

func (suite *KeeperTestSuite) TestMetadataScopeIterator() {
	for i := 1; i <= 10; i++ {
		valueOwner := ""
		if i == 5 {
			valueOwner = suite.user2
		}
		ns := types.NewScope(types.ScopeMetadataAddress(uuid.New()), nil, []string{suite.user1}, []string{suite.user1}, valueOwner)
		suite.app.MetadataKeeper.SetScope(suite.ctx, *ns)
	}
	count := 0
	suite.app.MetadataKeeper.IterateScopes(suite.ctx, func(s types.Scope) (stop bool) {
		count++
		return false
	})
	suite.Equal(10, count, "iterator should return a full list of scopes")

	count = 0
	suite.app.MetadataKeeper.IterateScopesForAddress(suite.ctx, suite.user1Addr, func(scopeID types.MetadataAddress) (stop bool) {
		count++
		suite.True(scopeID.IsScopeAddress())
		return false
	})
	suite.Equal(10, count, "iterator should return ten scope addresses")

	count = 0
	suite.app.MetadataKeeper.IterateScopesForAddress(suite.ctx, suite.user2Addr, func(scopeID types.MetadataAddress) (stop bool) {
		count++
		suite.True(scopeID.IsScopeAddress())
		return false
	})
	suite.Equal(1, count, "iterator should return a single address for the scope with value owned by user2")

	count = 0
	suite.app.MetadataKeeper.IterateScopes(suite.ctx, func(s types.Scope) (stop bool) {
		count++
		return count >= 5
	})
	suite.Equal(5, count, "using iterator stop function should stop iterator early")
}

func (suite *KeeperTestSuite) TestValidateScopeUpdate() {
	markerAddr := markertypes.MustGetMarkerAddress("testcoin").String()
	err := suite.app.MarkerKeeper.AddMarkerAccount(suite.ctx, &markertypes.MarkerAccount{
		BaseAccount: &authtypes.BaseAccount{
			Address:       markerAddr,
			AccountNumber: 23,
		},
		AccessControl: []markertypes.AccessGrant{
			{
				Address:     suite.user1,
				Permissions: markertypes.AccessListByNames("deposit,withdraw"),
			},
		},
		Denom:      "testcoin",
		Supply:     sdk.NewInt(1000),
		MarkerType: markertypes.MarkerType_Coin,
		Status:     markertypes.StatusActive,
	})
	suite.NoError(err)
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
			signers:  []string{suite.user1},
			wantErr:  true,
			errorMsg: "incorrect address length (must be at least 17, actual: 0)",
		},
		"valid proposed with nil existing doesn't error": {
			existing: types.Scope{},
			proposed: *types.NewScope(suite.scopeID, nil, []string{suite.user1}, []string{}, ""),
			signers:  []string{suite.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"can't change scope id in update": {
			existing: *types.NewScope(suite.scopeID, nil, []string{suite.user1}, []string{}, ""),
			proposed: *types.NewScope(changedID, nil, []string{suite.user1}, []string{}, ""),
			signers:  []string{suite.user1},
			wantErr:  true,
			errorMsg: fmt.Sprintf("cannot update scope identifier. expected %s, got %s", suite.scopeID.String(), changedID.String()),
		},
		"missing existing owner signer on update fails": {
			existing: *types.NewScope(suite.scopeID, nil, []string{suite.user1}, []string{}, ""),
			proposed: *types.NewScope(suite.scopeID, types.ScopeSpecMetadataAddress(suite.scopeUUID), []string{suite.user1}, []string{}, ""),
			signers:  []string{suite.user2},
			wantErr:  true,
			errorMsg: fmt.Sprintf("missing signature from existing owner %s; required for update", suite.user1),
		},
		"no error when update includes existing owner signer": {
			existing: *types.NewScope(suite.scopeID, nil, []string{suite.user1}, []string{}, ""),
			proposed: *types.NewScope(suite.scopeID, types.ScopeSpecMetadataAddress(suite.scopeUUID), []string{suite.user1}, []string{}, ""),
			signers:  []string{suite.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"setting value owner when unset does not error": {
			existing: *types.NewScope(suite.scopeID, nil, []string{suite.user1}, []string{}, ""),
			proposed: *types.NewScope(suite.scopeID, types.ScopeSpecMetadataAddress(suite.scopeUUID), []string{suite.user1}, []string{}, suite.user1),
			signers:  []string{suite.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"setting value owner to user does not require their signature": {
			existing: *types.NewScope(suite.scopeID, nil, []string{suite.user1}, []string{}, ""),
			proposed: *types.NewScope(suite.scopeID, types.ScopeSpecMetadataAddress(suite.scopeUUID), []string{suite.user1}, []string{}, suite.user2),
			signers:  []string{suite.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"setting value owner to new user does not require their signature": {
			existing: *types.NewScope(suite.scopeID, nil, []string{suite.user1}, []string{}, suite.user1),
			proposed: *types.NewScope(suite.scopeID, types.ScopeSpecMetadataAddress(suite.scopeUUID), []string{suite.user1}, []string{}, suite.user2),
			signers:  []string{suite.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"no change to value owner should not error": {
			existing: *types.NewScope(suite.scopeID, nil, []string{suite.user1}, []string{}, suite.user1),
			proposed: *types.NewScope(suite.scopeID, types.ScopeSpecMetadataAddress(suite.scopeUUID), []string{suite.user1}, []string{}, suite.user1),
			signers:  []string{suite.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"setting a new value owner should not error with withdraw permission": {
			existing: *types.NewScope(suite.scopeID, nil, []string{suite.user1}, []string{}, markerAddr),
			proposed: *types.NewScope(suite.scopeID, types.ScopeSpecMetadataAddress(suite.scopeUUID), []string{suite.user1}, []string{}, suite.user1),
			signers:  []string{suite.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"setting a new value owner fails if missing withdraw permission": {
			existing: *types.NewScope(suite.scopeID, nil, []string{suite.user2}, []string{}, markerAddr),
			proposed: *types.NewScope(suite.scopeID, types.ScopeSpecMetadataAddress(suite.scopeUUID), []string{suite.user2}, []string{}, suite.user2),
			signers:  []string{suite.user2},
			wantErr:  true,
			errorMsg: fmt.Sprintf("missing signature for %s with authority to withdraw/remove existing value owner", markerAddr),
		},
		"setting a new value owner fails if missing deposit permission": {
			existing: *types.NewScope(suite.scopeID, nil, []string{suite.user2}, []string{}, ""),
			proposed: *types.NewScope(suite.scopeID, types.ScopeSpecMetadataAddress(suite.scopeUUID), []string{suite.user2}, []string{}, markerAddr),
			signers:  []string{suite.user2},
			wantErr:  true,
			errorMsg: fmt.Sprintf("no signatures present with authority to add scope to marker %s", markerAddr),
		},
		"setting a new value owner fails for scope owner when value owner signature is missing": {
			existing: *types.NewScope(suite.scopeID, nil, []string{suite.user1}, []string{}, suite.user2),
			proposed: *types.NewScope(suite.scopeID, types.ScopeSpecMetadataAddress(suite.scopeUUID), []string{suite.user1}, []string{}, suite.user1),
			signers:  []string{suite.user1},
			wantErr:  true,
			errorMsg: fmt.Sprintf("missing signature from existing owner %s; required for update", suite.user2),
		},
		"unsetting all fields on a scope should be successful": {
			existing: *types.NewScope(suite.scopeID, types.ScopeSpecMetadataAddress(suite.scopeUUID), []string{suite.user1}, []string{}, suite.user1),
			proposed: types.Scope{ScopeId: suite.scopeID, OwnerAddress: []string{suite.user1}},
			signers:  []string{suite.user1},
			wantErr:  false,
			errorMsg: "",
		},
	}

	for n, tc := range cases {
		tc := tc

		suite.Run(n, func() {
			err = suite.app.MetadataKeeper.ValidateScopeUpdate(suite.ctx, tc.existing, tc.proposed, tc.signers)
			if tc.wantErr {
				suite.Error(err)
				suite.Equal(tc.errorMsg, err.Error())
			} else {
				suite.NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestScopeQuery() {
	app, ctx, queryClient, user1, user2 := suite.app, suite.ctx, suite.queryClient, suite.user1, suite.user2

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
	suite.NoError(err)

	_, err = queryClient.Scope(gocontext.Background(), &types.ScopeRequest{})
	suite.EqualError(err, "rpc error: code = InvalidArgument desc = scope id cannot be empty")

	_, err = queryClient.Scope(gocontext.Background(), &types.ScopeRequest{ScopeId: "6332c1a4-foo1-bare-895b-invalid65cb6"})
	suite.EqualError(err, "rpc error: code = InvalidArgument desc = invalid scope id: invalid UUID format")

	scopeResponse, err := queryClient.Scope(gocontext.Background(), &types.ScopeRequest{ScopeId: uuid.String()})
	suite.NoError(err)
	suite.NotNil(scopeResponse.Scope)
	suite.Equal(testIDs[0], scopeResponse.Scope.ScopeId)

	// TODO assert records when available
	// TODO assert record groups when available

	// only one scope has value owner set (user2)
	valueResponse, err := queryClient.ValueOwnership(gocontext.Background(), &types.ValueOwnershipRequest{Address: user2})
	suite.NoError(err)
	suite.Len(valueResponse.ScopeIds, 1)

	// 10 entries as all scopes have user1 as data_owner
	ownerResponse, err := queryClient.Ownership(gocontext.Background(), &types.OwnershipRequest{Address: user1})
	suite.NoError(err)
	suite.Len(ownerResponse.ScopeIds, 10)

	// one entry for user2 (as value owner)
	ownerResponse, err = queryClient.Ownership(gocontext.Background(), &types.OwnershipRequest{Address: user2})
	suite.NoError(err)
	suite.Len(ownerResponse.ScopeIds, 1)
}

func (suite *KeeperTestSuite) TestMetadataRecordGetSetRemove() {

	r, found := suite.app.MetadataKeeper.GetRecord(suite.ctx, suite.recordId)
	suite.NotNil(r)
	suite.False(found)

	process := types.NewProcess("processname", &types.Process_Hash{Hash: "HASH"}, "process_method")
	record := types.NewRecord(suite.recordName, suite.groupId, *process, []types.RecordInput{}, []types.RecordOutput{})

	suite.NotNil(record)
	suite.app.MetadataKeeper.SetRecord(suite.ctx, *record)

	r, found = suite.app.MetadataKeeper.GetRecord(suite.ctx, suite.recordId)
	suite.True(found)
	suite.NotNil(r)

	suite.app.MetadataKeeper.RemoveRecord(suite.ctx, suite.recordId)
	r, found = suite.app.MetadataKeeper.GetRecord(suite.ctx, suite.recordId)
	suite.False(found)
	suite.NotNil(r)
}

func (suite *KeeperTestSuite) TestMetadataRecordIterator() {
	for i := 1; i <= 10; i++ {
		process := types.NewProcess("processname", &types.Process_Hash{Hash: "HASH"}, "process_method")
		recordName := fmt.Sprintf("%s%v", suite.recordName, i)
		record := types.NewRecord(recordName, suite.groupId, *process, []types.RecordInput{}, []types.RecordOutput{})
		suite.app.MetadataKeeper.SetRecord(suite.ctx, *record)
	}
	count := 0
	suite.app.MetadataKeeper.IterateRecords(suite.ctx, suite.scopeID, func(s types.Record) (stop bool) {
		count++
		return false
	})
	suite.Equal(10, count, "iterator should return a full list of records")

}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
