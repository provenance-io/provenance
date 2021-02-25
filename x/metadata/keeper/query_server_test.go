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
	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/provenance/x/metadata/types"
)

type QueryServerTestSuite struct {
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

func (s *QueryServerTestSuite) SetupTest() {
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

	s.groupUUID = uuid.New()
	s.groupId = types.GroupMetadataAddress(s.scopeUUID, s.groupUUID)

	s.recordName = "TestRecord"
	s.recordId = types.RecordMetadataAddress(s.scopeUUID, s.recordName)

	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.MetadataKeeper)
	queryClient := types.NewQueryClient(queryHelper)
	s.queryClient = queryClient
}

func (s *QueryServerTestSuite) TestScopeQuery() {
	app, ctx, queryClient, user1, user2, recordName, groupID := s.app, s.ctx, s.queryClient, s.user1, s.user2, s.recordName, s.groupId

	testIDs := make([]types.MetadataAddress, 10)
	for i := 0; i < 10; i++ {
		valueOwner := ""
		if i == 5 {
			valueOwner = user2
		}
		testIDs[i] = types.ScopeMetadataAddress(uuid.New())
		ns := types.NewScope(testIDs[i], nil, ownerPartyList(user1), []string{user1}, valueOwner)
		app.MetadataKeeper.SetScope(ctx, *ns)

		name := fmt.Sprintf("%s%v", recordName, i)
		process := types.NewProcess("processname", &types.Process_Hash{Hash: "HASH"}, "process_method")
		record := types.NewRecord(name, groupID, *process, []types.RecordInput{}, []types.RecordOutput{})
		app.MetadataKeeper.SetRecord(ctx, *record)
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

	name := fmt.Sprintf("%s%v", recordName, 0)
	s.Equal(1, len(scopeResponse.Records))
	s.Equal(name, scopeResponse.Records[0].Name)
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

func (s *QueryServerTestSuite) TestRecordQuery() {
	app, ctx, queryClient, scopeUUID, groupID, recordName := s.app, s.ctx, s.queryClient, s.scopeUUID, s.groupId, s.recordName

	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("%s%v", recordName, i)
		process := types.NewProcess("processname", &types.Process_Hash{Hash: "HASH"}, "process_method")
		record := types.NewRecord(name, groupID, *process, []types.RecordInput{}, []types.RecordOutput{})
		app.MetadataKeeper.SetRecord(ctx, *record)
	}

	_, err := queryClient.Record(gocontext.Background(), nil)
	s.EqualError(err, "rpc error: code = InvalidArgument desc = empty request")

	_, err = queryClient.Record(gocontext.Background(), &types.RecordRequest{})
	s.EqualError(err, "rpc error: code = InvalidArgument desc = scope id cannot be empty")

	_, err = queryClient.Record(gocontext.Background(), &types.RecordRequest{ScopeId: "6332c1a4-foo1-bare-895b-invalid65cb6"})
	s.EqualError(err, "rpc error: code = InvalidArgument desc = invalid scope id: invalid UUID format")

	rs, err := queryClient.Record(gocontext.Background(), &types.RecordRequest{ScopeId: scopeUUID.String()})
	s.NoError(err)
	s.Equal(10, len(rs.Records), "should be 10 records in set for record query by scope uuid")
	for i := 0; i < 10; i++ {
		s.Equal(scopeUUID.String(), rs.ScopeId)
	}

	name := fmt.Sprintf("%s%v", recordName, 0)
	rs, err = queryClient.Record(gocontext.Background(), &types.RecordRequest{ScopeId: scopeUUID.String(), Name: name})
	s.NoError(err)
	s.Equal(1, len(rs.Records), "should be 1 record in set for record query by scope uuid")
	s.Equal(scopeUUID.String(), rs.ScopeId)
	s.Equal(name, rs.Records[0].Name)
}

func TestQuerierTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
