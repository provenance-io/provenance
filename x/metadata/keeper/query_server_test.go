package keeper_test

import (
	gocontext "context"
	"fmt"
	"testing"
	"time"

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

	record     types.Record
	recordName string
	recordID   types.MetadataAddress

	sessionUUID uuid.UUID
	sessionID   types.MetadataAddress
	sessionName string

	cSpecUUID uuid.UUID
	cSpecID   types.MetadataAddress
}

func (s *QueryServerTestSuite) SetupTest() {
	s.app = simapp.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, s.app.MetadataKeeper)
	s.queryClient = types.NewQueryClient(queryHelper)

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

	s.recordName = "TestRecord"
	s.recordID = types.RecordMetadataAddress(s.scopeUUID, s.recordName)

	s.sessionUUID = uuid.New()
	s.sessionID = types.SessionMetadataAddress(s.scopeUUID, s.sessionUUID)
	s.sessionName = "TestSession"

	s.cSpecUUID = uuid.New()
	s.cSpecID = types.ContractSpecMetadataAddress(s.cSpecUUID)

	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))
}

func TestQuerierTestSuite(t *testing.T) {
	suite.Run(t, new(QueryServerTestSuite))
}

func (s *QueryServerTestSuite) TestScopeQuery() {
	app, ctx, queryClient, user1, user2, recordName, sessionName := s.app, s.ctx, s.queryClient, s.user1, s.user2, s.recordName, s.sessionName

	testIDs := make([]types.MetadataAddress, 10)
	for i := 0; i < 10; i++ {
		valueOwner := ""
		if i == 5 {
			valueOwner = user2
		}

		scopeUUID := uuid.New()
		testIDs[i] = types.ScopeMetadataAddress(scopeUUID)
		ns := types.NewScope(testIDs[i], nil, ownerPartyList(user1), []string{user1}, valueOwner)
		app.MetadataKeeper.SetScope(ctx, *ns)

		sessionUUID := uuid.New()
		sessionID := types.SessionMetadataAddress(scopeUUID, sessionUUID)
		sName := fmt.Sprintf("%s%d", sessionName, i)
		session := types.NewSession(sName, sessionID, s.cSpecID, ownerPartyList(user1), nil)
		app.MetadataKeeper.SetSession(ctx, *session)

		rName := fmt.Sprintf("%s%d", recordName, i)
		process := types.NewProcess("processname", &types.Process_Hash{Hash: "HASH"}, "process_method")
		record := types.NewRecord(rName, sessionID, *process, []types.RecordInput{}, []types.RecordOutput{})
		app.MetadataKeeper.SetRecord(ctx, *record)
	}
	scope0UUID, err := testIDs[0].ScopeUUID()
	s.NoError(err, "ScopeUUID error")

	_, err = queryClient.Scope(gocontext.Background(), &types.ScopeRequest{})
	s.EqualError(err, "rpc error: code = InvalidArgument desc = scope uuid cannot be empty", "empty request error")

	_, err = queryClient.Scope(gocontext.Background(), &types.ScopeRequest{ScopeUuid: "6332c1a4-foo1-bare-895b-invalid65cb6"})
	s.EqualError(err, "rpc error: code = InvalidArgument desc = invalid scope uuid: invalid UUID format", "invalid uuid in request error")

	scopeResponse, err := queryClient.Scope(gocontext.Background(), &types.ScopeRequest{ScopeUuid: scope0UUID.String()})
	s.NoError(err, "valid request error")
	s.NotNil(scopeResponse.Scope, "scope in scope response")
	s.Equal(testIDs[0], scopeResponse.Scope.ScopeId, "scopeId")

	record0Name := fmt.Sprintf("%s%v", recordName, 0)
	s.Equal(1, len(scopeResponse.Records), "records count")
	s.Equal(record0Name, scopeResponse.Records[0].Name, "record name")

	session0Name := fmt.Sprintf("%s%v", sessionName, 0)
	s.Equal(1, len(scopeResponse.Sessions), "session count")
	s.Equal(session0Name, scopeResponse.Sessions[0].Name, "session name")

	// only one scope has value owner set (user2)
	valueResponse, err := queryClient.ValueOwnership(gocontext.Background(), &types.ValueOwnershipRequest{Address: user2})
	s.NoError(err)
	s.Len(valueResponse.ScopeUuids, 1)

	// 10 entries as all scopes have user1 as data_owner
	ownerResponse, err := queryClient.Ownership(gocontext.Background(), &types.OwnershipRequest{Address: user1})
	s.NoError(err)
	s.Len(ownerResponse.ScopeUuids, 10)

	// one entry for user2 (as value owner)
	ownerResponse, err = queryClient.Ownership(gocontext.Background(), &types.OwnershipRequest{Address: user2})
	s.NoError(err)
	s.Len(ownerResponse.ScopeUuids, 1)
}

func (s *QueryServerTestSuite) TestRecordQuery() {
	app, ctx, queryClient, scopeUUID, scopeID, sessionID, recordName := s.app, s.ctx, s.queryClient, s.scopeUUID, s.scopeID, s.sessionID, s.recordName

	recordNames := make([]string, 10)
	for i := 0; i < 10; i++ {
		recordNames[i] = fmt.Sprintf("%s%v", recordName, i)
		process := types.NewProcess("processname", &types.Process_Hash{Hash: "HASH"}, "process_method")
		record := types.NewRecord(recordNames[i], sessionID, *process, []types.RecordInput{}, []types.RecordOutput{})
		app.MetadataKeeper.SetRecord(ctx, *record)
	}

	_, err := queryClient.RecordsByScopeUUID(gocontext.Background(), &types.RecordsByScopeUUIDRequest{})
	s.EqualError(err, "rpc error: code = InvalidArgument desc = scope uuid cannot be empty")

	_, err = queryClient.RecordsByScopeUUID(gocontext.Background(), &types.RecordsByScopeUUIDRequest{ScopeUuid: "6332c1a4-foo1-bare-895b-invalid65cb6"})
	s.EqualError(err, "rpc error: code = InvalidArgument desc = invalid scope uuid: invalid UUID format")

	rsUUID, err := queryClient.RecordsByScopeUUID(gocontext.Background(), &types.RecordsByScopeUUIDRequest{ScopeUuid: scopeUUID.String()})
	s.NoError(err)
	s.Equal(10, len(rsUUID.Records), "should be 10 records in set for record query by scope uuid")
	s.Equal(scopeUUID.String(), rsUUID.ScopeUuid)
	s.Equal(scopeID.String(), rsUUID.ScopeId)

	rsUUID, err = queryClient.RecordsByScopeUUID(gocontext.Background(), &types.RecordsByScopeUUIDRequest{ScopeUuid: scopeUUID.String(), Name: recordNames[0]})
	s.NoError(err)
	s.Equal(1, len(rsUUID.Records), "should be 1 record in set for record query by scope uuid")
	s.Equal(scopeUUID.String(), rsUUID.ScopeUuid)
	s.Equal(scopeID.String(), rsUUID.ScopeId)
	s.Equal(recordNames[0], rsUUID.Records[0].Name)

	_, err = queryClient.RecordsByScopeID(gocontext.Background(), &types.RecordsByScopeIDRequest{})
	s.EqualError(err, "rpc error: code = InvalidArgument desc = scope id cannot be empty")

	_, err = queryClient.RecordsByScopeID(gocontext.Background(), &types.RecordsByScopeIDRequest{ScopeId: "foo"})
	s.EqualError(err, "rpc error: code = InvalidArgument desc = invalid scope id foo : decoding bech32 failed: invalid bech32 string length 3")

	rsID, err := queryClient.RecordsByScopeID(gocontext.Background(), &types.RecordsByScopeIDRequest{ScopeId: scopeID.String()})
	s.NoError(err)
	s.Equal(10, len(rsID.Records), "should be 10 records in set for record query by scope id")
	s.Equal(scopeUUID.String(), rsID.ScopeUuid)
	s.Equal(scopeID.String(), rsID.ScopeId)

	rsID, err = queryClient.RecordsByScopeID(gocontext.Background(), &types.RecordsByScopeIDRequest{ScopeId: scopeID.String(), Name: recordNames[0]})
	s.NoError(err)
	s.Equal(1, len(rsID.Records), "should be 1 record in set for record query by scope id")
	s.Equal(scopeUUID.String(), rsID.ScopeUuid)
	s.Equal(scopeID.String(), rsID.ScopeId)
	s.Equal(recordNames[0], rsID.Records[0].Name)
}

func (s *QueryServerTestSuite) TestSessionQuery() {
	app, ctx, queryClient, scopeID, scopeUUID, sessionID, sessionUUID, cSpecID := s.app, s.ctx, s.queryClient, s.scopeID, s.scopeUUID, s.sessionID, s.sessionUUID, s.cSpecID

	session := types.NewSession("name", sessionID, cSpecID, []types.Party{
		{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: types.PartyType_PARTY_TYPE_AFFILIATE}},
		&types.AuditFields{CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", CreatedDate: time.Now(),
			UpdatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", UpdatedDate: time.Now(),
			Message: "message",
		})
	app.MetadataKeeper.SetSession(ctx, *session)
	for i := 0; i < 9; i++ {
		sID := types.SessionMetadataAddress(scopeUUID, uuid.New())
		session := types.NewSession("name", sID, cSpecID, []types.Party{
			{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: types.PartyType_PARTY_TYPE_AFFILIATE}},
			&types.AuditFields{CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", CreatedDate: time.Now(),
				UpdatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", UpdatedDate: time.Now(),
				Message: "message",
			})
		app.MetadataKeeper.SetSession(ctx, *session)

	}

	_, err := queryClient.SessionContextByUUID(gocontext.Background(), &types.SessionContextByUUIDRequest{})
	s.EqualError(err, "rpc error: code = InvalidArgument desc = scope uuid cannot be empty")

	_, err = queryClient.SessionContextByUUID(gocontext.Background(), &types.SessionContextByUUIDRequest{ScopeUuid: "6332c1a4-foo1-bare-895b-invalid65cb6"})
	s.EqualError(err, "rpc error: code = InvalidArgument desc = invalid scope uuid: invalid UUID format")

	sc, err := queryClient.SessionContextByUUID(gocontext.Background(), &types.SessionContextByUUIDRequest{ScopeUuid: scopeUUID.String()})
	s.NoError(err)
	s.Equal(10, len(sc.Sessions), "should be 10 sessions in set for session context query by scope uuid")
	s.Equal(scopeID.String(), sc.ScopeId)
	s.Equal("", sc.SessionId)

	_, err = queryClient.SessionContextByUUID(gocontext.Background(), &types.SessionContextByUUIDRequest{ScopeUuid: scopeUUID.String(), SessionUuid: "6332c1a4-foo1-bare-895b-invalid65cb6"})
	s.EqualError(err, "rpc error: code = InvalidArgument desc = invalid session uuid: invalid UUID format")

	sc, err = queryClient.SessionContextByUUID(gocontext.Background(), &types.SessionContextByUUIDRequest{ScopeUuid: scopeUUID.String(), SessionUuid: sessionUUID.String()})
	s.NoError(err)
	s.Equal(1, len(sc.Sessions), "should be 1 session in set for session context query by scope uuid and session uuid")
	s.Equal(scopeID.String(), sc.ScopeId)
	s.Equal(sessionID.String(), sc.SessionId)

	_, err = queryClient.SessionContextByID(gocontext.Background(), &types.SessionContextByIDRequest{})
	s.EqualError(err, "rpc error: code = InvalidArgument desc = scope id cannot be empty")

	_, err = queryClient.SessionContextByID(gocontext.Background(), &types.SessionContextByIDRequest{ScopeId: "invalidbech32"})
	s.EqualError(err, "rpc error: code = InvalidArgument desc = incorrect scope id: decoding bech32 failed: invalid index of 1")

	scrs, err := queryClient.SessionContextByID(gocontext.Background(), &types.SessionContextByIDRequest{ScopeId: scopeID.String()})
	s.Equal(10, len(scrs.Sessions), "should be 10 sessions in set for session context query by scope uuid")
	s.Equal(scopeID.String(), scrs.ScopeId)
	s.Equal("", scrs.SessionId)

	_, err = queryClient.SessionContextByID(gocontext.Background(), &types.SessionContextByIDRequest{ScopeId: scopeID.String(), SessionId: "invalidbech32"})
	s.EqualError(err, "rpc error: code = InvalidArgument desc = incorrect scope id: decoding bech32 failed: invalid index of 1")

	scrs, err = queryClient.SessionContextByID(gocontext.Background(), &types.SessionContextByIDRequest{ScopeId: scopeID.String(), SessionId: sessionID.String()})
	s.Equal(1, len(scrs.Sessions), "should be 1 sessions in set for session context query by scope id and session id")
	s.Equal(scopeID.String(), scrs.ScopeId)
	s.Equal(sessionID.String(), scrs.SessionId)
}

// TODO: ScopeSpecification tests
// TODO: ContractSpecification tests
// TODO: ContractSpecificationExtended tests
// TODO: RecordSpecificationsForContractSpecification test
// TODO: RecordSpecification tests
// TODO: RecordSpecificationByID tests
