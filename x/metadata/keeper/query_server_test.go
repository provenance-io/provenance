package keeper_test

import (
	gocontext "context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/metadata/types"
)

type QueryServerTestSuite struct {
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

	scopeUUID uuid.UUID
	scopeID   types.MetadataAddress

	scopeSpecUUID uuid.UUID
	scopeSpecID   types.MetadataAddress

	record     types.Record
	recordName string
	recordID   types.MetadataAddress

	sessionUUID uuid.UUID
	sessionID   types.MetadataAddress
	sessionName string

	cSpecUUID uuid.UUID
	cSpecID   types.MetadataAddress
	recSpecID types.MetadataAddress
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

	s.scopeSpecUUID = uuid.New()
	s.scopeSpecID = types.ScopeSpecMetadataAddress(s.scopeSpecUUID)

	s.recordName = "TestRecord"
	s.recordID = types.RecordMetadataAddress(s.scopeUUID, s.recordName)

	s.sessionUUID = uuid.New()
	s.sessionID = types.SessionMetadataAddress(s.scopeUUID, s.sessionUUID)
	s.sessionName = "TestSession"

	s.cSpecUUID = uuid.New()
	s.cSpecID = types.ContractSpecMetadataAddress(s.cSpecUUID)
	s.recSpecID = types.RecordSpecMetadataAddress(s.cSpecUUID, s.recordName)

	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))
}

func TestQuerierTestSuite(t *testing.T) {
	suite.Run(t, new(QueryServerTestSuite))
}

// TODO: Params tests

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
		record := types.NewRecord(rName, sessionID, *process, []types.RecordInput{}, []types.RecordOutput{}, s.recSpecID)
		app.MetadataKeeper.SetRecord(ctx, *record)
	}
	scope0UUID, err := testIDs[0].ScopeUUID()
	s.NoError(err, "ScopeUUID error")

	_, err = queryClient.Scope(gocontext.Background(), &types.ScopeRequest{})
	s.EqualError(err, "rpc error: code = InvalidArgument desc = empty request parameters", "empty request error")

	_, err = queryClient.Scope(gocontext.Background(), &types.ScopeRequest{ScopeId: "6332c1a4-foo1-bare-895b-invalid65cb6"})
	s.EqualError(err, "rpc error: code = InvalidArgument desc = could not parse [6332c1a4-foo1-bare-895b-invalid65cb6] into either a scope address (decoding bech32 failed: invalid character not part of charset: 45) or uuid (invalid UUID format)", "invalid uuid in request error")

	// TODO: expand this to test new features/failures of the Scope query.

	fullReq0 := types.ScopeRequest{
		ScopeId:         scope0UUID.String(),
		IncludeSessions: true,
		IncludeRecords:  true,
	}
	scopeResponse, err := queryClient.Scope(gocontext.Background(), &fullReq0)
	s.NoError(err, "valid request error")
	s.NotNil(scopeResponse.Scope, "scope in scope response")
	s.Equal(testIDs[0], scopeResponse.Scope.Scope.ScopeId, "scopeId")

	record0Name := fmt.Sprintf("%s%v", recordName, 0)
	s.Equal(1, len(scopeResponse.Records), "records count")
	s.Equal(record0Name, scopeResponse.Records[0].Record.Name, "record name")

	session0Name := fmt.Sprintf("%s%v", sessionName, 0)
	s.Equal(1, len(scopeResponse.Sessions), "session count")
	s.Equal(session0Name, scopeResponse.Sessions[0].Session.Name, "session name")

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

// TODO: ScopesAll tests

func (s *QueryServerTestSuite) TestSessionsQuery() {
	app, ctx, queryClient := s.app, s.ctx, s.queryClient

	scope := types.NewScope(s.scopeID, s.scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, s.user1)
	app.MetadataKeeper.SetScope(ctx, *scope)

	session := types.NewSession("name", s.sessionID, s.cSpecID, []types.Party{
		{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: types.PartyType_PARTY_TYPE_AFFILIATE}},
		&types.AuditFields{CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", CreatedDate: time.Now(),
			UpdatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", UpdatedDate: time.Now(),
			Message: "message",
		})
	app.MetadataKeeper.SetSession(ctx, *session)
	var anotherSessionID types.MetadataAddress
	for i := 0; i < 9; i++ {
		anotherSessionID = types.SessionMetadataAddress(s.scopeUUID, uuid.New())
		sess := types.NewSession("name", anotherSessionID, s.cSpecID, []types.Party{
			{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: types.PartyType_PARTY_TYPE_AFFILIATE}},
			&types.AuditFields{CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", CreatedDate: time.Now(),
				UpdatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", UpdatedDate: time.Now(),
				Message: "message",
			})
		app.MetadataKeeper.SetSession(ctx, *sess)
	}
	record := types.NewRecord(s.recordName, s.sessionID,
		*types.NewProcess("procname", &types.Process_Hash{Hash: "PROC_HASH"}, "proc_method"),
		[]types.RecordInput{},
		[]types.RecordOutput{},
		types.RecordSpecMetadataAddress(s.cSpecUUID, s.recordName),
	)
	app.MetadataKeeper.SetRecord(ctx, *record)
	anotherRecordName := "another-record"
	anotherRecord := types.NewRecord(anotherRecordName, anotherSessionID,
		*types.NewProcess("anotherprocname", &types.Process_Hash{Hash: "ANOTHER_PROC_HASH"}, "another_proc_method"),
		[]types.RecordInput{},
		[]types.RecordOutput{},
		types.RecordSpecMetadataAddress(s.cSpecUUID, anotherRecordName),
	)
	app.MetadataKeeper.SetRecord(ctx, *anotherRecord)
	anotherRecordID := types.RecordMetadataAddress(s.scopeUUID, anotherRecordName)

	// Note: I tried just using uuid.New() here, but if something tries to decode it as a
	//       bech32 string (like when it's provided as a SessionID without any other parameters)
	//       it'll fail at different parts, making pre-defined expected errors... difficult.
	unknownUUID := uuid.MustParse("bb38cf02-1f78-4e72-a624-fa3d34383c81")

	zero := 0
	one := 1
	ten := 10

	testCases := []struct {
		name       string
		req        *types.SessionsRequest
		err        string
		count      *int
		scopeID    types.MetadataAddress
		sessionID  types.MetadataAddress
		nilSession bool
	}{
		// Note: Cannot do a "nil request" test because the cosmos-sdk queryClient stuff panics first.
		// {
		// 	name: "nil request",
		// 	req:  nil,
		// 	err:  "rpc error: code = InvalidArgument desc = empty request",
		// },
		{
			name: "empty request",
			req:  &types.SessionsRequest{},
			err:  "rpc error: code = InvalidArgument desc = empty request parameters",
		},

		// only scope id
		{
			name: "only scope id invalid - error",
			req:  &types.SessionsRequest{ScopeId: "6332c1a4-foo1-bare-895b-invalid65cb6"},
			err:  "rpc error: code = InvalidArgument desc = could not parse [6332c1a4-foo1-bare-895b-invalid65cb6] into either a scope address (decoding bech32 failed: invalid character not part of charset: 45) or uuid (invalid UUID format)",
		},
		{
			name:  "only scope id as uuid not found - empty",
			req:   &types.SessionsRequest{ScopeId: unknownUUID.String()},
			count: &zero,
		},
		{
			name:    "only scope id as uuid exists - results",
			req:     &types.SessionsRequest{ScopeId: s.scopeUUID.String()},
			count:   &ten,
			scopeID: s.scopeID,
		},
		{
			name:  "only scope id as addr not found - empty",
			req:   &types.SessionsRequest{ScopeId: types.ScopeMetadataAddress(unknownUUID).String()},
			count: &zero,
		},
		{
			name:    "only scope id as addr exists - results",
			req:     &types.SessionsRequest{ScopeId: s.scopeID.String()},
			count:   &ten,
			scopeID: s.scopeID,
		},

		// only session id
		{
			name: "only session id invalid - error",
			req:  &types.SessionsRequest{SessionId: "not-a-valid-session-id"},
			err:  "rpc error: code = InvalidArgument desc = could not parse [not-a-valid-session-id] into a session address: decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "only session id as uuid - error",
			req:  &types.SessionsRequest{SessionId: unknownUUID.String()},
			err:  fmt.Sprintf("rpc error: code = InvalidArgument desc = could not parse [%s] into a session address: decoding bech32 failed: invalid separator index 35", unknownUUID),
		},
		{
			name:       "only session id as addr not found - empty",
			req:        &types.SessionsRequest{SessionId: types.SessionMetadataAddress(unknownUUID, unknownUUID).String()},
			count:      &one,
			scopeID:    types.ScopeMetadataAddress(unknownUUID),
			sessionID:  types.SessionMetadataAddress(unknownUUID, unknownUUID),
			nilSession: true,
		},
		{
			name:      "only session id as addr exists - result",
			req:       &types.SessionsRequest{SessionId: s.sessionID.String()},
			count:     &one,
			scopeID:   s.scopeID,
			sessionID: s.sessionID,
		},

		// only record addr
		{
			name: "only record addr invalid - error",
			req:  &types.SessionsRequest{RecordAddr: "not-a-valid-record-id"},
			err:  "rpc error: code = InvalidArgument desc = could not parse [not-a-valid-record-id] into a record address: decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "only record addr not found - error",
			req:  &types.SessionsRequest{RecordAddr: types.RecordMetadataAddress(s.scopeUUID, "no-such-record").String()},
			err:  fmt.Sprintf("rpc error: code = InvalidArgument desc = record %s does not exist", types.RecordMetadataAddress(s.scopeUUID, "no-such-record")),
		},
		{
			name:      "only record addr exists - result",
			req:       &types.SessionsRequest{RecordAddr: types.RecordMetadataAddress(s.scopeUUID, s.recordName).String()},
			count:     &one,
			scopeID:   s.scopeID,
			sessionID: s.sessionID,
		},

		// only record name
		{
			name: "only record name - error",
			req:  &types.SessionsRequest{RecordName: s.recordName},
			err:  "rpc error: code = InvalidArgument desc = a scope is required to look up sessions by record name",
		},

		// scope id and session id
		{
			name: "scope id invalid session id ok - error",
			req:  &types.SessionsRequest{ScopeId: "not-a-scope-id", SessionId: s.sessionID.String()},
			err:  "rpc error: code = InvalidArgument desc = could not parse [not-a-scope-id] into either a scope address (decoding bech32 failed: invalid separator index -1) or uuid (invalid UUID length: 14)",
		},
		{
			name: "scope id as uuid exists session id invalid - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeUUID.String(), SessionId: "invalidSessionID"},
			err:  "rpc error: code = InvalidArgument desc = could not parse [invalidSessionID] into either a session address (decoding bech32 failed: string not all lowercase or all uppercase) or uuid (invalid UUID length: 16)",
		},
		{
			name: "scope id as addr exists session id invalid - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: "invalidSessionID"},
			err:  "rpc error: code = InvalidArgument desc = could not parse [invalidSessionID] into either a session address (decoding bech32 failed: string not all lowercase or all uppercase) or uuid (invalid UUID length: 16)",
		},
		{
			name: "scope id as uuid exists session id as addr wrong scope - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeUUID.String(), SessionId: types.SessionMetadataAddress(unknownUUID, s.sessionUUID).String()},
			err: fmt.Sprintf("rpc error: code = InvalidArgument desc = session %s is not in scope %s",
				types.SessionMetadataAddress(unknownUUID, s.sessionUUID), s.scopeID),
		},
		{
			name: "scope id as addr exists session id as addr wrong scope - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: types.SessionMetadataAddress(unknownUUID, s.sessionUUID).String()},
			err: fmt.Sprintf("rpc error: code = InvalidArgument desc = session %s is not in scope %s",
				types.SessionMetadataAddress(unknownUUID, s.sessionUUID), s.scopeID),
		},
		{
			name:       "scope id as uuid exists session id as uuid not found - empty",
			req:        &types.SessionsRequest{ScopeId: s.scopeUUID.String(), SessionId: unknownUUID.String()},
			count:      &one,
			scopeID:    s.scopeID,
			sessionID:  types.SessionMetadataAddress(s.scopeUUID, unknownUUID),
			nilSession: true,
		},
		{
			name:       "scope id as addr exists session id as uuid not found - empty",
			req:        &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: unknownUUID.String()},
			count:      &one,
			scopeID:    s.scopeID,
			sessionID:  types.SessionMetadataAddress(s.scopeUUID, unknownUUID),
			nilSession: true,
		},
		{
			name:      "scope id as uuid exists session id as uuid exists - result",
			req:       &types.SessionsRequest{ScopeId: s.scopeUUID.String(), SessionId: s.sessionUUID.String()},
			count:     &one,
			scopeID:   s.scopeID,
			sessionID: s.sessionID,
		},
		{
			name:      "scope id as addr exists session id as uuid exists - result",
			req:       &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: s.sessionUUID.String()},
			count:     &one,
			scopeID:   s.scopeID,
			sessionID: s.sessionID,
		},
		{
			name:      "scope id as uuid exists session id as addr exists - result",
			req:       &types.SessionsRequest{ScopeId: s.scopeUUID.String(), SessionId: s.sessionID.String()},
			count:     &one,
			scopeID:   s.scopeID,
			sessionID: s.sessionID,
		},
		{
			name:      "scope id as addr exists session id as addr exists - result",
			req:       &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: s.sessionID.String()},
			count:     &one,
			scopeID:   s.scopeID,
			sessionID: s.sessionID,
		},

		// scope id and record addr
		{
			name: "scope id invalid record addr ok - error",
			req:  &types.SessionsRequest{ScopeId: "notReallyAScopeID", RecordAddr: s.recordID.String()},
			err:  "rpc error: code = InvalidArgument desc = could not parse [notReallyAScopeID] into either a scope address (decoding bech32 failed: string not all lowercase or all uppercase) or uuid (invalid UUID length: 17)",
		},
		{
			name: "scope id exists record addr invalid - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), RecordAddr: "invalid-record-addr"},
			err:  "rpc error: code = InvalidArgument desc = could not parse [invalid-record-addr] into a record address: decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "scope id exists record addr wrong scope - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), RecordAddr: types.RecordMetadataAddress(unknownUUID, s.recordName).String()},
			err:  fmt.Sprintf("rpc error: code = InvalidArgument desc = record %s is not part of scope %s", types.RecordMetadataAddress(unknownUUID, s.recordName), s.scopeID),
		},
		{
			name: "scope id exists record addr not found - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), RecordAddr: types.RecordMetadataAddress(s.scopeUUID, "record-not-real").String()},
			err:  fmt.Sprintf("rpc error: code = InvalidArgument desc = record %s does not exist", types.RecordMetadataAddress(s.scopeUUID, "record-not-real")),
		},
		{
			name:      "scope id exists record addr exists - result",
			req:       &types.SessionsRequest{ScopeId: s.scopeID.String(), RecordAddr: s.recordID.String()},
			count:     &one,
			scopeID:   s.scopeID,
			sessionID: s.sessionID,
		},

		// scope id and record name
		{
			name: "scope id invalid record name ok - error",
			req:  &types.SessionsRequest{ScopeId: "illegitimate-scope", RecordName: s.recordName},
			err:  "rpc error: code = InvalidArgument desc = could not parse [illegitimate-scope] into either a scope address (decoding bech32 failed: invalid separator index -1) or uuid (invalid UUID length: 18)",
		},
		{
			name: "scope id as uuid not found record name ok - error",
			req:  &types.SessionsRequest{ScopeId: unknownUUID.String(), RecordName: s.recordName},
			err:  fmt.Sprintf("rpc error: code = InvalidArgument desc = record %s does not exist", types.RecordMetadataAddress(unknownUUID, s.recordName)),
		},
		{
			name: "scope id as addr not found record name ok - error",
			req:  &types.SessionsRequest{ScopeId: types.ScopeMetadataAddress(unknownUUID).String(), RecordName: s.recordName},
			err:  fmt.Sprintf("rpc error: code = InvalidArgument desc = record %s does not exist", types.RecordMetadataAddress(unknownUUID, s.recordName)),
		},
		{
			name: "scope id as uuid exists record name not found - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeUUID.String(), RecordName: "no-such-record"},
			err:  fmt.Sprintf("rpc error: code = InvalidArgument desc = record %s does not exist", types.RecordMetadataAddress(s.scopeUUID, "no-such-record")),
		},
		{
			name: "scope id as addr exists record name not found - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), RecordName: "still-no-such-record"},
			err:  fmt.Sprintf("rpc error: code = InvalidArgument desc = record %s does not exist", types.RecordMetadataAddress(s.scopeUUID, "still-no-such-record")),
		},
		{
			name:      "scope id as uuid exists record name exists - result",
			req:       &types.SessionsRequest{ScopeId: s.scopeUUID.String(), RecordName: s.recordName},
			count:     &one,
			scopeID:   s.scopeID,
			sessionID: s.sessionID,
		},
		{
			name:      "scope id as addr exists record name exists - result",
			req:       &types.SessionsRequest{ScopeId: s.scopeID.String(), RecordName: s.recordName},
			count:     &one,
			scopeID:   s.scopeID,
			sessionID: s.sessionID,
		},

		// session id and record addr
		{
			name: "session id invalid record addr ok - error",
			req:  &types.SessionsRequest{SessionId: "lol-nope-notASessionId", RecordAddr: s.recordID.String()},
			err:  "rpc error: code = InvalidArgument desc = could not parse [lol-nope-notASessionId] into either a session address (decoding bech32 failed: string not all lowercase or all uppercase) or uuid (invalid UUID length: 22)",
		},
		{
			name: "session id exists record addr invalid - error",
			req:  &types.SessionsRequest{SessionId: s.sessionID.String(), RecordAddr: "yeah-not-a-record-addr"},
			err:  "rpc error: code = InvalidArgument desc = could not parse [yeah-not-a-record-addr] into a record address: decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "session id exists record addr wrong scope - error",
			req:  &types.SessionsRequest{SessionId: s.sessionID.String(), RecordAddr: types.RecordMetadataAddress(unknownUUID, s.recordName).String()},
			err:  fmt.Sprintf("rpc error: code = InvalidArgument desc = session %s is not in scope %s", s.sessionID, types.ScopeMetadataAddress(unknownUUID)),
		},
		{
			name: "session id exists record addr wrong session - error",
			req:  &types.SessionsRequest{SessionId: s.sessionID.String(), RecordAddr: anotherRecordID.String()},
			err:  fmt.Sprintf("rpc error: code = InvalidArgument desc = record %s belongs to session %s (not %s)", anotherRecordID, anotherSessionID, s.sessionID),
		},
		{
			name:      "session id as addr exists record addr exists - result",
			req:       &types.SessionsRequest{SessionId: s.sessionID.String(), RecordAddr: s.recordID.String()},
			count:     &one,
			scopeID:   s.scopeID,
			sessionID: s.sessionID,
		},
		{
			name:      "session id as uuid exists record addr exists - result",
			req:       &types.SessionsRequest{SessionId: s.sessionUUID.String(), RecordAddr: s.recordID.String()},
			count:     &one,
			scopeID:   s.scopeID,
			sessionID: s.sessionID,
		},

		// session id and record name
		{
			name: "session id invalid record name ok - error",
			req:  &types.SessionsRequest{SessionId: "nopenope-nope-nope-nota-sessionuuidx", RecordName: s.recordName},
			err:  "rpc error: code = InvalidArgument desc = could not parse [nopenope-nope-nope-nota-sessionuuidx] into a session address: decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "session id as uuid record name ok - error",
			req:  &types.SessionsRequest{SessionId: unknownUUID.String(), RecordName: s.recordName},
			err:  fmt.Sprintf("rpc error: code = InvalidArgument desc = could not parse [%s] into a session address: decoding bech32 failed: invalid separator index 35", unknownUUID),
		},
		{
			name: "session id as addr not found record name ok - error",
			req:  &types.SessionsRequest{SessionId: types.SessionMetadataAddress(unknownUUID, unknownUUID).String(), RecordName: s.recordName},
			err:  fmt.Sprintf("rpc error: code = InvalidArgument desc = record %s does not exist", types.RecordMetadataAddress(unknownUUID, s.recordName)),
		},
		{
			name: "session id as addr exists record name not found - error",
			req:  &types.SessionsRequest{SessionId: s.sessionID.String(), RecordName: "not-gonna-find-this-record"},
			err:  fmt.Sprintf("rpc error: code = InvalidArgument desc = record %s does not exist", types.RecordMetadataAddress(s.scopeUUID, "not-gonna-find-this-record")),
		},
		{
			name: "session id as addr exists record name wrong session - error",
			req:  &types.SessionsRequest{SessionId: s.sessionID.String(), RecordName: anotherRecordName},
			err: fmt.Sprintf("rpc error: code = InvalidArgument desc = record %s belongs to session %s (not %s)",
				types.RecordMetadataAddress(s.scopeUUID, anotherRecordName), anotherSessionID, s.sessionID),
		},
		{
			name:      "session id as addr exists record name exists - result",
			req:       &types.SessionsRequest{SessionId: s.sessionID.String(), RecordName: s.recordName},
			count:     &one,
			scopeID:   s.scopeID,
			sessionID: s.sessionID,
		},

		// record addr and record name
		{
			name: "record addr invalid record name ok - error",
			req:  &types.SessionsRequest{RecordAddr: "abby-lane", RecordName: s.recordName},
			err:  "rpc error: code = InvalidArgument desc = could not parse [abby-lane] into a record address: decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "record addr exists record name wrong - error",
			req:  &types.SessionsRequest{RecordAddr: s.recordID.String(), RecordName: anotherRecordName},
			err:  fmt.Sprintf("rpc error: code = InvalidArgument desc = record %s does not have name %s", s.recordID, anotherRecordName),
		},
		{
			name:      "record addr exists record name matches - result",
			req:       &types.SessionsRequest{RecordAddr: s.recordID.String(), RecordName: s.recordName},
			count:     &one,
			scopeID:   s.scopeID,
			sessionID: s.sessionID,
		},

		// scope id and session id and record addr
		{
			name: "scope id invalid session id ok record addr ok - error",
			req:  &types.SessionsRequest{ScopeId: "bad scope", SessionId: s.scopeID.String(), RecordAddr: s.recordID.String()},
			err:  "rpc error: code = InvalidArgument desc = could not parse [bad scope] into either a scope address (decoding bech32 failed: invalid character in string: ' ') or uuid (invalid UUID length: 9)",
		},
		{
			name: "scope id exists session id invalid record addr ok - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: "bad session", RecordAddr: s.recordID.String()},
			err:  "rpc error: code = InvalidArgument desc = could not parse [bad session] into either a session address (decoding bech32 failed: invalid character in string: ' ') or uuid (invalid UUID length: 11)",
		},
		{
			name: "scope id exists session id exists record addr invalid - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: s.sessionID.String(), RecordAddr: "bad record"},
			err:  "rpc error: code = InvalidArgument desc = could not parse [bad record] into a record address: decoding bech32 failed: invalid character in string: ' '",
		},
		{
			name: "scope id wrong scope session id exists record addr exists - error",
			req:  &types.SessionsRequest{ScopeId: types.ScopeMetadataAddress(unknownUUID).String(), SessionId: s.sessionID.String(), RecordAddr: s.recordID.String()},
			err:  fmt.Sprintf("rpc error: code = InvalidArgument desc = record %s is not part of scope %s", s.recordID, types.ScopeMetadataAddress(unknownUUID)),
		},
		{
			name:      "scope id exists session id exists record addr exists - result",
			req:       &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: s.sessionID.String(), RecordAddr: s.recordID.String()},
			count:     &one,
			scopeID:   s.scopeID,
			sessionID: s.sessionID,
		},

		// scope id and session id and record name
		{
			name: "scope id invalid session id ok record name ok - error",
			req:  &types.SessionsRequest{ScopeId: "nopescope", SessionId: s.sessionID.String(), RecordName: s.recordName},
			err:  "rpc error: code = InvalidArgument desc = could not parse [nopescope] into either a scope address (decoding bech32 failed: invalid separator index -1) or uuid (invalid UUID length: 9)",
		},
		{
			name: "scope id exists session id invalid record name ok - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: "nosesh", RecordName: s.recordName},
			err:  "rpc error: code = InvalidArgument desc = could not parse [nosesh] into either a session address (decoding bech32 failed: invalid bech32 string length 6) or uuid (invalid UUID length: 6)",
		},
		{
			name: "scope id exists session id exists record name not found - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: s.sessionID.String(), RecordName: "nopenopenope"},
			err:  fmt.Sprintf("rpc error: code = InvalidArgument desc = record %s does not exist", types.RecordMetadataAddress(s.scopeUUID, "nopenopenope")),
		},
		{
			name:      "scope id exists session id exists record name exists - result",
			req:       &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: s.sessionID.String(), RecordName: s.recordName},
			count:     &one,
			scopeID:   s.scopeID,
			sessionID: s.sessionID,
		},

		// scope id and record addr and record name
		{
			name: "scope id invalid record addr ok record name ok - error",
			req:  &types.SessionsRequest{ScopeId: "veryBadScope", RecordAddr: s.recordID.String(), RecordName: s.recordName},
			err:  "rpc error: code = InvalidArgument desc = could not parse [veryBadScope] into either a scope address (decoding bech32 failed: string not all lowercase or all uppercase) or uuid (invalid UUID length: 12)",
		},
		{
			name: "scope id exists record addr invalid record name ok - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), RecordAddr: "kindofbadrecord", RecordName: s.recordName},
			err:  "rpc error: code = InvalidArgument desc = could not parse [kindofbadrecord] into a record address: decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "scope id exists record addr exists record name wrong - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), RecordAddr: s.recordID.String(), RecordName: "wrongagain"},
			err:  fmt.Sprintf("rpc error: code = InvalidArgument desc = record %s does not have name wrongagain", s.recordID),
		},
		{
			name: "scope id exists record addr wrong scope record name ok - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), RecordAddr: types.RecordMetadataAddress(unknownUUID, s.recordName).String(), RecordName: s.recordName},
			err:  fmt.Sprintf("rpc error: code = InvalidArgument desc = record %s is not part of scope %s", types.RecordMetadataAddress(unknownUUID, s.recordName), s.scopeID),
		},
		{
			name:      "scope id exists record addr exists record name matches - result",
			req:       &types.SessionsRequest{ScopeId: s.scopeID.String(), RecordAddr: s.recordID.String(), RecordName: s.recordName},
			count:     &one,
			scopeID:   s.scopeID,
			sessionID: s.sessionID,
		},

		// session id and record addr and record name
		{
			name: "session id invalid record addr ok record name ok - error",
			req:  &types.SessionsRequest{SessionId: "nopesess", RecordAddr: s.recordID.String(), RecordName: s.recordName},
			err:  "rpc error: code = InvalidArgument desc = could not parse [nopesess] into either a session address (decoding bech32 failed: invalid separator index -1) or uuid (invalid UUID length: 8)",
		},
		{
			name: "session id exists record addr invalid record name ok - error",
			req:  &types.SessionsRequest{SessionId: s.sessionID.String(), RecordAddr: "incorrect-record-id", RecordName: s.recordName},
			err:  "rpc error: code = InvalidArgument desc = could not parse [incorrect-record-id] into a record address: decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "session id exists record addr wrong scope record name ok - error",
			req:  &types.SessionsRequest{SessionId: s.sessionID.String(), RecordAddr: types.RecordMetadataAddress(unknownUUID, s.recordName).String(), RecordName: s.recordName},
			err:  fmt.Sprintf("rpc error: code = InvalidArgument desc = session %s is not in scope %s", s.sessionID, types.ScopeMetadataAddress(unknownUUID)),
		},
		{
			name: "session id exists record addr wrong session record name ok - error",
			req:  &types.SessionsRequest{SessionId: s.sessionID.String(), RecordAddr: s.recordID.String(), RecordName: anotherRecordName},
			err:  fmt.Sprintf("rpc error: code = InvalidArgument desc = record %s does not have name another-record", s.recordID),
		},
		{
			name:      "session id exists record addr exists record name ok - result",
			req:       &types.SessionsRequest{SessionId: s.sessionID.String(), RecordAddr: s.recordID.String(), RecordName: s.recordName},
			count:     &one,
			scopeID:   s.scopeID,
			sessionID: s.sessionID,
		},

		// scope id and session id and record addr and record name
		{
			name: "scope id invalid session id ok record addr ok record name ok - error",
			req:  &types.SessionsRequest{ScopeId: "negatoryscope", SessionId: s.sessionID.String(), RecordAddr: s.recordID.String(), RecordName: s.recordName},
			err:  "rpc error: code = InvalidArgument desc = could not parse [negatoryscope] into either a scope address (decoding bech32 failed: invalid separator index -1) or uuid (invalid UUID length: 13)",
		},
		{
			name: "scope id exists session id invalid record addr ok record name ok - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: "negatorysession", RecordAddr: s.recordID.String(), RecordName: s.recordName},
			err:  "rpc error: code = InvalidArgument desc = could not parse [negatorysession] into either a session address (decoding bech32 failed: invalid separator index -1) or uuid (invalid UUID length: 15)",
		},
		{
			name: "scope id exists session id exists record addr invalid record name ok - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: s.sessionID.String(), RecordAddr: "negatoryrecord", RecordName: s.recordName},
			err:  "rpc error: code = InvalidArgument desc = could not parse [negatoryrecord] into a record address: decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "scope id exists session id exists record addr exists record name wrong - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: s.sessionID.String(), RecordAddr: s.recordID.String(), RecordName: anotherRecordName},
			err:  fmt.Sprintf("rpc error: code = InvalidArgument desc = record %s does not have name %s", s.recordID, anotherRecordName),
		},
		{
			name:      "scope id exists session id exists record addr exists record name matches - result",
			req:       &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: s.sessionID.String(), RecordAddr: s.recordID.String(), RecordName: s.recordName},
			count:     &one,
			scopeID:   s.scopeID,
			sessionID: s.sessionID,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			expectError := len(tc.err) > 0
			checkExactCount := tc.count != nil
			checkScopeIds := len(tc.scopeID) > 0
			checkSessionIds := len(tc.sessionID) > 0
			sr, err := queryClient.Sessions(gocontext.Background(), tc.req)
			if expectError {
				assert.EqualError(t, err, tc.err, "expected error")
			} else {
				require.NoError(t, err, "unexpected error: %s", err)
			}
			if !expectError || checkExactCount || checkScopeIds || checkSessionIds || tc.nilSession {
				require.NotNil(t, sr, "result of Sessions query")
				if checkExactCount {
					assert.Equal(t, *tc.count, len(sr.Sessions), "number of sessions found")
				} else {
					require.Greater(t, 0, len(sr.Sessions), "at least one session result expected")
				}
				for i, x := range sr.Sessions {
					if checkScopeIds {
						assert.Equalf(t, tc.scopeID, x.SessionIdInfo.ScopeIdInfo.ScopeId, "scope id of result.Sessions[%d]", i)
					}
					if checkSessionIds {
						assert.Equalf(t, tc.sessionID, x.SessionIdInfo.SessionId, "session id of result.Sessions[%d]", i)
					}
					if tc.nilSession {
						assert.Nilf(t, x.Session, "expected result.Sessions[x].Session to be nil: result.Sessions[%d] = %s", i, x.String())
					} else {
						assert.NotNilf(t, x.Session, "expected result.Sessions[x].Session to have value: result.Sessions[%d] = %s", i, x.String())
					}
				}
			}
		})
	}

	// Run a couple tests on the Include* flags.
	s.T().Run("include scope", func(t *testing.T) {
		req := types.SessionsRequest{SessionId: s.sessionID.String(), IncludeScope: true}
		sr, err := queryClient.Sessions(gocontext.Background(), &req)
		require.NoErrorf(t, err, "unexpected error: %s", err)
		require.NotNil(t, sr, "result of Sessions query")
		assert.Equal(t, 0, len(sr.Records), "number of records")
		require.NotNil(t, sr.Scope, "scope wrapper")
		require.NotNil(t, sr.Scope.Scope, "the scope")
		assert.Equal(t, s.scopeID, sr.Scope.Scope.ScopeId, "scope id")
	})
	s.T().Run("include records", func(t *testing.T) {
		req := types.SessionsRequest{SessionId: s.sessionID.String(), IncludeRecords: true}
		sr, err := queryClient.Sessions(gocontext.Background(), &req)
		require.NoErrorf(t, err, "unexpected error: %s", err)
		require.NotNil(t, sr, "result of Sessions query")
		assert.Nilf(t, sr.Scope, "scope should be nil: %s", sr.String())
		require.Equal(t, 1, len(sr.Records), "number of records")
		require.NotNil(t, sr.Records[0].Record, "the record")
		assert.Equal(t, s.scopeID, sr.Records[0].RecordIdInfo.ScopeIdInfo.ScopeId, "scope id")
		assert.Equal(t, s.sessionID, sr.Records[0].Record.SessionId, "session id")
		assert.Equal(t, s.recordName, sr.Records[0].Record.Name, "record name")
	})
	s.T().Run("iinclude both scope and records", func(t *testing.T) {
		req := types.SessionsRequest{SessionId: s.sessionID.String(), IncludeScope: true, IncludeRecords: true}
		sr, err := queryClient.Sessions(gocontext.Background(), &req)
		require.NoErrorf(t, err, "unexpected error: %s", err)
		require.NotNil(t, sr, "result of Sessions query")
		require.NotNil(t, sr.Scope, "scope wrapper")
		require.NotNil(t, sr.Scope.Scope, "the scope")
		require.Equal(t, 1, len(sr.Records), "number of records")
		require.NotNil(t, sr.Records[0].Record, "the record")
		assert.Equal(t, s.scopeID, sr.Scope.Scope.ScopeId, "scope id")
		assert.Equal(t, s.scopeID, sr.Records[0].RecordIdInfo.ScopeIdInfo.ScopeId, "scope id")
		assert.Equal(t, s.sessionID, sr.Records[0].Record.SessionId, "session id")
		assert.Equal(t, s.recordName, sr.Records[0].Record.Name, "record name")
	})
}

// TODO: SessionsAll tests

func (s *QueryServerTestSuite) TestRecordsQuery() {
	app, ctx, queryClient, scopeUUID, scopeID, sessionID, recordName := s.app, s.ctx, s.queryClient, s.scopeUUID, s.scopeID, s.sessionID, s.recordName

	recordNames := make([]string, 10)
	for i := 0; i < 10; i++ {
		recordNames[i] = fmt.Sprintf("%s%v", recordName, i)
		process := types.NewProcess("processname", &types.Process_Hash{Hash: "HASH"}, "process_method")
		record := types.NewRecord(recordNames[i], sessionID, *process, []types.RecordInput{}, []types.RecordOutput{}, s.recSpecID)
		app.MetadataKeeper.SetRecord(ctx, *record)
	}

	_, err := queryClient.Records(gocontext.Background(), &types.RecordsRequest{})
	s.EqualError(err, "rpc error: code = InvalidArgument desc = empty request parameters")

	_, err = queryClient.Records(gocontext.Background(), &types.RecordsRequest{ScopeId: "foo"})
	s.EqualError(err, "rpc error: code = InvalidArgument desc = could not parse [foo] into either a scope address (decoding bech32 failed: invalid bech32 string length 3) or uuid (invalid UUID length: 3)")

	_, err = queryClient.Records(gocontext.Background(), &types.RecordsRequest{ScopeId: "6332c1a4-foo1-bare-895b-invalid65cb6"})
	s.EqualError(err, "rpc error: code = InvalidArgument desc = could not parse [6332c1a4-foo1-bare-895b-invalid65cb6] into either a scope address (decoding bech32 failed: invalid character not part of charset: 45) or uuid (invalid UUID format)")

	// TODO: expand this to test new features/failures of the Records query.

	rsUUID, err := queryClient.Records(gocontext.Background(), &types.RecordsRequest{ScopeId: scopeUUID.String()})
	s.NoError(err)
	s.Equal(10, len(rsUUID.Records), "should be 10 records in set for record query by scope uuid")
	s.Equal(scopeUUID.String(), rsUUID.Records[0].RecordIdInfo.ScopeIdInfo.ScopeUuid)
	s.Equal(scopeID.String(), rsUUID.Records[0].RecordIdInfo.ScopeIdInfo.ScopeAddr)

	rsUUID2, err2 := queryClient.Records(gocontext.Background(), &types.RecordsRequest{ScopeId: scopeUUID.String(), Name: recordNames[0]})
	s.NoError(err2)
	s.Equal(1, len(rsUUID2.Records), "should be 1 record in set for record query by scope uuid")
	s.Equal(scopeUUID.String(), rsUUID2.Records[0].RecordIdInfo.ScopeIdInfo.ScopeUuid)
	s.Equal(scopeID.String(), rsUUID2.Records[0].RecordIdInfo.ScopeIdInfo.ScopeAddr)
	s.Equal(recordNames[0], rsUUID2.Records[0].Record.Name)

	rsID, err := queryClient.Records(gocontext.Background(), &types.RecordsRequest{ScopeId: scopeID.String()})
	s.NoError(err)
	s.Equal(10, len(rsID.Records), "should be 10 records in set for record query by scope id")
	s.Equal(scopeUUID.String(), rsID.Records[0].RecordIdInfo.ScopeIdInfo.ScopeUuid)
	s.Equal(scopeID.String(), rsID.Records[0].RecordIdInfo.ScopeIdInfo.ScopeAddr)

	rsID, err = queryClient.Records(gocontext.Background(), &types.RecordsRequest{ScopeId: scopeID.String(), Name: recordNames[0]})
	s.NoError(err)
	s.Equal(1, len(rsID.Records), "should be 1 record in set for record query by scope id")
	s.Equal(scopeUUID.String(), rsID.Records[0].RecordIdInfo.ScopeIdInfo.ScopeUuid)
	s.Equal(scopeID.String(), rsID.Records[0].RecordIdInfo.ScopeIdInfo.ScopeAddr)
	s.Equal(recordNames[0], rsID.Records[0].Record.Name)
}

// TODO: RecordsAll tests
// TODO: Ownership tests
// TODO: ValueOwnership tests
// TODO: ScopeSpecification tests
// TODO: ScopeSpecificationsAll tests
// TODO: ContractSpecification tests
// TODO: ContractSpecificationsAll tests
// TODO: RecordSpecificationsForContractSpecification tests
// TODO: RecordSpecification tests
// TODO: RecordSpecificationsAll tests
// TODO: OSLocatorParams tests
// TODO: OSLocator tests
// TODO: OSLocatorsByURI tests
// TODO: OSLocatorsByScope tests
// TODO: OSAllLocators tests

// TODO: Helper IsBase64 tests
// TODO: Helper ParseScopeID tests
// TODO: Helper ParseSessionID tests
// TODO: Helper ParseSessionAddr tests
// TODO: Helper ParseRecordAddr tests
// TODO: Helper ParseScopeSpecID tests
// TODO: Helper ParseContractSpecID tests
// TODO: Helper ParseRecordSpecID tests
// TODO: Helper getPageRequest tests
