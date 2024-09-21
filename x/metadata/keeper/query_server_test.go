package keeper_test

import (
	"bytes"
	gocontext "context"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/metadata/types"
)

func TestQuerierTestSuite(t *testing.T) {
	suite.Run(t, new(QueryServerTestSuite))
}

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
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false)
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

// AssertErrorValue is a wrapper on assertions.AssertErrorValue using this suite's current T().
func (s *QueryServerTestSuite) AssertErrorValue(theError error, expected string, msgAndArgs ...interface{}) bool {
	s.T().Helper()
	return assertions.AssertErrorValue(s.T(), theError, expected, msgAndArgs...)
}

// AssertEqualPageResponses will assert that the provided PageResponse values are equal
// with friendlier failure messages than if you just used s.Assert()..Equal(expected, actual).
func (s *QueryServerTestSuite) AssertEqualPageResponses(expected, actual *query.PageResponse, fieldName string) bool {
	s.T().Helper()
	// Check the individual fields first in a friendlier way than when all combined.
	pOK := true
	if expected != nil && actual != nil {
		// Compare the Total as ints because failures on uint64 show the values in hex, which isn't very handy.
		pOK = s.Assert().Equal(int(expected.Total), int(actual.Total), fieldName+".Total (as an int)") && pOK
		// Compare the NextKey directly. The failure message will list the bytes in multiple formats (int and char).
		pOK = s.Assert().Equalf(expected.NextKey, actual.NextKey, fieldName+".NextKey") && pOK
	}
	// If one of those failed, we can stop now.
	if !pOK {
		return false
	}
	// Run the comparison on them as a whole, just to be sure (and handle when one of them is nil).
	return s.Assert().Equal(expected, actual, fieldName)
}

// dataSetup is a collection metadata entries (scopes, sessions, records) generated for testing.
type dataSetup struct {
	// Scopes are each of the scopes in the order they're created.
	Scopes []*types.Scope
	// Sessions are each of the sessions for each scope in the order they're created.
	// E.g. Sessions[0] are all the sessions in Scopes[0].
	Sessions [][]*types.Session
	// Records are each of the records for each session for each scope in they order they're created.
	// E.g Records[1][0] are all of the records in Sessions[1][0] (which are all in Scopes[1]).
	Records [][][]*types.Record

	// ScopeIDs are the scope metadata addresses for each of the scopes.
	// E.g. ScopeIDs[0] is equal to Scopes[0].ScopeId.
	ScopeIDs []types.MetadataAddress
	// ScopeUUIDs are the uuids in the scope metadata addresses for each of the scopes.
	// E.g. ScopeUUIDs[0] is the uuid from Scopes[0].ScopeId (which is also ScopeIDs[0]).
	ScopeUUIDs []uuid.UUID
	// SessionIDs are the session metadata addresses for each of the sessions for each scope.
	// E.g. SessionIDs[1][0] is the session metadata address for Sessions[1][0], (which is in Scopes[1]) etc.
	SessionIDs [][]types.MetadataAddress
	// SessionUUIDs are the secondary uuids for each of the sessions for each scope.
	// E.g. SessionIDs[1][0] = SessionMetadataAddress(ScopeUUIDs[1], SessionUUIDs[0]).
	SessionUUIDs [][]uuid.UUID
	// RecordIDs are the record metadata addresses of each of the records in each session of each scope.
	// E.g. RecordIDs[2][1][0] is equal to Records[2][1][0].GetRecordAddress().
	RecordIDs [][][]types.MetadataAddress
	// RecordNames are the names of each of the records in each sesson of each scope.
	// E.g. RecordNames[2][1][0] is equal to Records[2][1][0].Name.
	RecordNames [][][]string

	// AllScopeSessions is all of the sessions for each scope. Each slice is sorted by address.
	// e.g. AllScopeSessions[1] is all of the sessions for Scopes[1], sorted the same as in state.
	// The secondary index of this field doesn't correlate to any other slices, only the first index.
	AllScopeSessions [][]*types.Session
	// AllScopeRecords is all of the records for each scope. Each slice is sorted by address.
	// e.g. AllScopeRecords[1] is all of the records for Scopes[1], sorted the same as in state.
	// The secondary index of this field doesn't correlate to any other slices, only the first index.
	AllScopeRecords [][]*types.Record

	// AllScopes is all of the Scopes sorted by address (same order as in state).
	// The index in this field doesn't correlate to any other slices.
	AllScopes []*types.Scope
	// AllSessions is all of the Sessions sorted by address (same order as in state).
	// The index in this field doesn't correlate to any other slices.
	AllSessions []*types.Session
	// AllRecords is all of the records sorted by address (same order as in state).
	// The index in this field doesn't correlate to any other slices.
	AllRecords []*types.Record
}

// createDataSetup will create all the scopes, sessions, and records needed to fill the provided counts.
// len(counts) = number of scopes.
// len(counts[i]) = number of sessions for scope[i].
// counts[i][j] = number of records for session j (which is part of scope i).
func createDataSetup(counts [][]int) *dataSetup {
	newAccAddr := func(i int) string {
		return sdk.AccAddress(fmt.Sprintf("%02d_acc_address______", i)).String()
	}
	newOwners := func(i int) []types.Party {
		rv := ownerPartyList(newAccAddr(i))
		if i%2 != 0 {
			rv = append(rv, types.Party{Address: newAccAddr(i + 10), Role: 1})
		}
		return rv
	}
	newScope := func(i int) *types.Scope {
		rv := &types.Scope{
			ScopeId:            types.ScopeMetadataAddress(newTestUUID(i)),
			SpecificationId:    types.ScopeSpecMetadataAddress(newTestUUID(i + 16)),
			Owners:             newOwners(i),
			DataAccess:         []string{newAccAddr(i + 20)},
			ValueOwnerAddress:  newAccAddr(i + 40),
			RequirePartyRollup: i%2 == 0,
		}
		if i > 2 {
			rv.DataAccess = append(rv.DataAccess, newAccAddr(i+30))
		}
		return rv
	}
	sessionAddr := func(iScope, iSession int) types.MetadataAddress {
		return types.SessionMetadataAddress(newTestUUID(iScope), newTestUUID(80+iScope*16+iSession))
	}
	cSpecAddr := func(iScope, iSession int) types.MetadataAddress {
		return types.ContractSpecMetadataAddress(newTestUUID(160 + iScope*16 + iSession))
	}
	newSession := func(iScope, iSession int) *types.Session {
		return &types.Session{
			SessionId:       sessionAddr(iScope, iSession),
			SpecificationId: cSpecAddr(iScope, iSession),
			Parties:         newOwners(iScope),
			Name:            fmt.Sprintf("Scope_%02d_TestSession_%02d", iScope, iSession),
		}
	}
	newRecord := func(iScope, iSession, iRecord int) *types.Record {
		rName := fmt.Sprintf("Scope_%02d_TestRecord_%02d_%02d", iScope, iSession, iRecord)
		id := fmt.Sprintf("%02d_%02d_%02d", iScope, iSession, iRecord)
		return &types.Record{
			Name:      rName,
			SessionId: sessionAddr(iScope, iSession),
			Process: types.Process{
				Name:      "process_name_" + id,
				ProcessId: &types.Process_Hash{Hash: "process_id_hash_" + id},
				Method:    "process_method_" + id,
			},
			Inputs: []types.RecordInput{
				{
					Name:     "inputs[0]_name_" + id,
					Source:   &types.RecordInput_Hash{Hash: "inputs[0]_source_hash_" + id},
					TypeName: "inputs[0]_type_" + id,
					Status:   types.RecordInputStatus_Record,
				},
			},
			Outputs: []types.RecordOutput{
				{
					Hash:   "outputs[0]_hash_" + id,
					Status: types.ResultStatus_RESULT_STATUS_PASS,
				},
			},
			SpecificationId: cSpecAddr(iScope, iSession).MustGetAsRecordSpecAddress(rName),
		}
	}

	rv := &dataSetup{
		Scopes:           make([]*types.Scope, len(counts)),
		Sessions:         make([][]*types.Session, len(counts)),
		Records:          make([][][]*types.Record, len(counts)),
		ScopeIDs:         make([]types.MetadataAddress, len(counts)),
		ScopeUUIDs:       make([]uuid.UUID, len(counts)),
		SessionIDs:       make([][]types.MetadataAddress, len(counts)),
		SessionUUIDs:     make([][]uuid.UUID, len(counts)),
		RecordIDs:        make([][][]types.MetadataAddress, len(counts)),
		RecordNames:      make([][][]string, len(counts)),
		AllScopeSessions: make([][]*types.Session, len(counts)),
		AllScopeRecords:  make([][]*types.Record, len(counts)),
	}

	for iScope := range rv.Records {
		scope := newScope(iScope)
		rv.Scopes[iScope] = scope
		rv.ScopeIDs[iScope] = scope.ScopeId
		rv.ScopeUUIDs[iScope], _ = scope.ScopeId.ScopeUUID()
		rv.AllScopes = append(rv.AllScopes, scope)

		sessionCount := len(counts[iScope])
		rv.Sessions[iScope] = make([]*types.Session, sessionCount)
		rv.SessionIDs[iScope] = make([]types.MetadataAddress, sessionCount)
		rv.SessionUUIDs[iScope] = make([]uuid.UUID, sessionCount)
		rv.Records[iScope] = make([][]*types.Record, sessionCount)
		rv.RecordIDs[iScope] = make([][]types.MetadataAddress, sessionCount)
		rv.RecordNames[iScope] = make([][]string, sessionCount)

		for iSession := range rv.Records[iScope] {
			session := newSession(iScope, iSession)
			rv.Sessions[iScope][iSession] = session
			rv.SessionIDs[iScope][iSession] = session.SessionId
			rv.SessionUUIDs[iScope][iSession], _ = session.SessionId.SessionUUID()
			rv.AllSessions = append(rv.AllSessions, session)
			rv.AllScopeSessions[iScope] = append(rv.AllScopeSessions[iScope], session)

			recordCount := counts[iScope][iSession]
			rv.Records[iScope][iSession] = make([]*types.Record, recordCount)
			rv.RecordIDs[iScope][iSession] = make([]types.MetadataAddress, recordCount)
			rv.RecordNames[iScope][iSession] = make([]string, recordCount)

			for iRecord := range rv.Records[iScope][iSession] {
				record := newRecord(iScope, iSession, iRecord)
				rv.Records[iScope][iSession][iRecord] = record
				rv.RecordIDs[iScope][iSession][iRecord] = record.GetRecordAddress()
				rv.RecordNames[iScope][iSession][iRecord] = record.Name
				rv.AllRecords = append(rv.AllRecords, record)
				rv.AllScopeRecords[iScope] = append(rv.AllScopeRecords[iScope], record)
			}
		}

		slices.SortFunc(rv.AllScopeSessions[iScope], compareSessions)
		slices.SortFunc(rv.AllScopeRecords[iScope], compareRecords)
	}

	slices.SortFunc(rv.AllScopes, compareScopes)
	slices.SortFunc(rv.AllSessions, compareSessions)
	slices.SortFunc(rv.AllRecords, compareRecords)

	return rv
}

// identifyScope will return a human-readable string indicating which scope from this dataSetup is in the provided wrapper.
func (d dataSetup) IdentifyScope(w *types.ScopeWrapper) string {
	if w == nil {
		return fmt.Sprintf("%#v", w)
	}
	for i, scope := range d.Scopes {
		if w.Scope != nil && w.Scope.ScopeId.Equals(scope.ScopeId) {
			return fmt.Sprintf("Scopes[%d]", i)
		}
		if w.Scope == nil && w.ScopeIdInfo != nil && w.ScopeIdInfo.ScopeId.Equals(scope.ScopeId) {
			return fmt.Sprintf("Scopes[%d]<not found>", i)
		}
	}
	return fmt.Sprintf("%#v", w)
}

// IdentifyScopes will call IdentifyScope on each of the provided scope wrappers.
func (d dataSetup) IdentifyScopes(scopes []*types.ScopeWrapper) []string {
	if scopes == nil {
		return nil
	}
	rv := make([]string, len(scopes))
	for i, scope := range scopes {
		rv[i] = d.IdentifyScope(scope)
	}
	return rv
}

// IdentifySession will return a human-readable string indicating which session from this dataSetup is in the provided wrapper.
func (d dataSetup) IdentifySession(w *types.SessionWrapper) string {
	if w == nil {
		return fmt.Sprintf("%#v", w)
	}
	for i := range d.Sessions {
		for j, session := range d.Sessions[i] {
			if w.Session != nil && w.Session.SessionId.Equals(session.SessionId) {
				return fmt.Sprintf("Sessions[%d][%d]", i, j)
			}
			if w.Session == nil && w.SessionIdInfo != nil && w.SessionIdInfo.SessionId.Equals(session.SessionId) {
				return fmt.Sprintf("Sessions[%d][%d]<not found>", i, j)
			}
		}
	}
	if w.Session != nil && len(w.Session.Name) > 0 {
		return w.Session.Name
	}
	return fmt.Sprintf("%#v", w)
}

// IdentifySessions will call IdentifySession on each of the provided session wrappers.
func (d dataSetup) IdentifySessions(sessions []*types.SessionWrapper) []string {
	if sessions == nil {
		return nil
	}
	rv := make([]string, len(sessions))
	for i, session := range sessions {
		rv[i] = d.IdentifySession(session)
	}
	return rv
}

// IdentifyRecord will return a human-readable string indicating which record from this dataSetup is in the provided wrapper.
func (d dataSetup) IdentifyRecord(w *types.RecordWrapper) string {
	if w == nil {
		return fmt.Sprintf("%#v", w)
	}
	for i := range d.Records {
		for j := range d.Records[i] {
			for k, record := range d.Records[i][j] {
				if w.Record != nil && w.Record.Name == record.Name && w.Record.SessionId.Equals(record.SessionId) {
					return fmt.Sprintf("Records[%d][%d][%d]", i, j, k)
				}
				if w.Record == nil && w.RecordIdInfo != nil && w.RecordIdInfo.RecordId.Equals(record.GetRecordAddress()) {
					return fmt.Sprintf("Records[%d][%d][%d]<not found>", i, j, k)
				}
			}
		}
	}
	if w.Record != nil && len(w.Record.Name) > 0 {
		return w.Record.Name
	}
	return fmt.Sprintf("%#v", w)
}

// IdentifyRecords will call IdentifyRecord on each of the provided record wrappers.
func (d dataSetup) IdentifyRecords(records []*types.RecordWrapper) []string {
	if records == nil {
		return nil
	}
	rv := make([]string, len(records))
	for i, record := range records {
		rv[i] = d.IdentifyRecord(record)
	}
	return rv
}

// newTestUUID creates a new UUID by converting the provided int into a two char
// hex code then repeating those two chars 8 times (for 16 total bytes).
func newTestUUID(i int) uuid.UUID {
	return uuid.UUID([]byte(strings.Repeat(fmt.Sprintf("%02X", i), 8)))
}

// compareScopes is used to sort scopes by their id (which is how they are sorted in state).
// It returns -1 if a < b, or 0 if a == b, or 1 if a > b.
// E.g To test if a > b, you could do if compareRecords(a, b) > 0.
func compareScopes(a, b *types.Scope) int {
	var addrA, addrB types.MetadataAddress
	if a != nil {
		addrA = a.ScopeId
	}
	if b != nil {
		addrB = b.ScopeId
	}
	return bytes.Compare(addrA, addrB)
}

// compareSessions is used to sort sessions by their id (which is how they are sorted in state).
// It returns -1 if a < b, or 0 if a == b, or 1 if a > b.
// E.g To test if a > b, you could do if compareRecords(a, b) > 0.
func compareSessions(a, b *types.Session) int {
	var addrA, addrB types.MetadataAddress
	if a != nil {
		addrA = a.SessionId
	}
	if b != nil {
		addrB = b.SessionId
	}
	return bytes.Compare(addrA, addrB)
}

// compareRecords is used to sort records by their id (which is how they are sorted in state).
// It's not all that efficient, but should be fine or small quantities of records.
// It returns -1 if a < b, or 0 if a == b, or 1 if a > b.
// E.g To test if a > b, you could do if compareRecords(a, b) > 0.
func compareRecords(a, b *types.Record) int {
	var addrA, addrB types.MetadataAddress
	if a != nil {
		addrA = a.GetRecordAddress()
	}
	if b != nil {
		addrB = b.GetRecordAddress()
	}
	return bytes.Compare(addrA, addrB)
}

// setData will write all the data in the provided dataSetup to state.
func (s *QueryServerTestSuite) setData(data *dataSetup) {
	for i, scope := range data.Scopes {
		assertions.RequireNotPanicsNoError(s.T(), func() error {
			return s.app.MetadataKeeper.SetScope(s.ctx, *scope)
		}, "[%d]: SetScope(%#v)", i, scope)
	}
	for i := range data.Sessions {
		for j, session := range data.Sessions[i] {
			s.Require().NotPanics(func() {
				s.app.MetadataKeeper.SetSession(s.ctx, *session)
			}, "[%d][%d]: SetSession(%#v)", i, j, session)
		}
	}
	for i := range data.Records {
		for j := range data.Records[i] {
			for k, record := range data.Records[i][j] {
				s.Require().NotPanics(func() {
					s.app.MetadataKeeper.SetRecord(s.ctx, *record)
				}, "[%d][%d][%d]: SetRecord(%#v)", i, j, k, record)
			}
		}
	}
}

// createDataSetup will create a dataSetup from the provided emptyRecords and write it to state.
func (s *QueryServerTestSuite) createData(emptyRecords [][]int) *dataSetup {
	data := createDataSetup(emptyRecords)
	s.setData(data)
	return data
}

// wrapScopes creates a ScopeWrapper with each of the provided scopes.
func wrapScopes(scopes []*types.Scope, includeIDInfo bool) []*types.ScopeWrapper {
	if scopes == nil {
		return nil
	}
	rv := make([]*types.ScopeWrapper, len(scopes))
	for i, scope := range scopes {
		rv[i] = types.WrapScope(scope, includeIDInfo)
	}
	return rv
}

// wrapSessions creates a SessionWrapper with each of the provided sessions.
func wrapSessions(sessions []*types.Session, includeIDInfo bool) []*types.SessionWrapper {
	if sessions == nil {
		return nil
	}
	rv := make([]*types.SessionWrapper, len(sessions))
	for i, session := range sessions {
		rv[i] = types.WrapSession(session, includeIDInfo)
	}
	return rv
}

// wrapRecords creates a RecordWrapper with each of the provided records.
func wrapRecords(records []*types.Record, includeIDInfo bool) []*types.RecordWrapper {
	if records == nil {
		return nil
	}
	rv := make([]*types.RecordWrapper, len(records))
	for i, record := range records {
		rv[i] = types.WrapRecord(record, includeIDInfo)
	}
	return rv
}

// TODO: Params tests

func (s *QueryServerTestSuite) TestScopeQuery() {
	// Make 3 scopes:
	//   scopes[0]: one session with one record.
	//   scopes[1]: one session with two records.
	//   scopes[2]: two sessions, one record and two records.
	data := s.createData([][]int{{1}, {2}, {1, 2}})
	// Valid Scopes indexes: [0] [1] [2]
	// Valid Session indexes: [0][0] [1][0] [2][0] [2][1]
	// Valid Record indexes: [0][0][0] [1][0][0] [1][0][1] [2][0][0] [2][1][0] [2][1][1]

	tests := []struct {
		name    string
		req     types.ScopeRequest
		expResp *types.ScopeResponse
		expErr  string
	}{
		{
			name:   "empty request",
			req:    types.ScopeRequest{},
			expErr: "empty request parameters: invalid request",
		},
		{
			name:   "invalid scope id",
			req:    types.ScopeRequest{ScopeId: "6332c1a4-foo1-bare-895b-invalid65cb6"},
			expErr: "could not parse [6332c1a4-foo1-bare-895b-invalid65cb6] into either a scope address (decoding bech32 failed: invalid character not part of charset: 45) or uuid (invalid UUID format): invalid request",
		},
		{
			name:   "invalid session",
			req:    types.ScopeRequest{SessionAddr: "nope"},
			expErr: "could not parse [nope] into a session address: decoding bech32 failed: invalid bech32 string length 4: invalid request",
		},
		{
			name:   "invalid record",
			req:    types.ScopeRequest{RecordAddr: "alsonope"},
			expErr: "could not parse [alsonope] into a record address: decoding bech32 failed: invalid separator index -1: invalid request",
		},
		{
			name: "scope uuid and other session",
			req: types.ScopeRequest{
				ScopeId:     data.ScopeUUIDs[0].String(),
				SessionAddr: data.SessionIDs[1][0].String(),
			},
			expErr: "session " + data.SessionIDs[1][0].String() + " is not in scope " + data.ScopeIDs[0].String() + ": invalid request",
		},
		{
			name: "scope addr and other session",
			req: types.ScopeRequest{
				ScopeId:     data.ScopeIDs[1].String(),
				SessionAddr: data.SessionIDs[2][0].String(),
			},
			expErr: "session " + data.SessionIDs[2][0].String() + " is not in scope " +
				data.ScopeIDs[1].String() + ": invalid request",
		},
		{
			name: "scope uuid and other record",
			req: types.ScopeRequest{
				ScopeId:    data.ScopeUUIDs[2].String(),
				RecordAddr: data.RecordIDs[0][0][0].String(),
			},
			expErr: "record " + data.RecordIDs[0][0][0].String() + " is not part of scope " +
				data.ScopeIDs[2].String() + ": invalid request",
		},
		{
			name: "scope addr and other record",
			req: types.ScopeRequest{
				ScopeId:    data.ScopeIDs[1].String(),
				RecordAddr: data.RecordIDs[2][0][0].String(),
			},
			expErr: "record " + data.RecordIDs[2][0][0].String() + " is not part of scope " +
				data.ScopeIDs[1].String() + ": invalid request",
		},
		{
			name: "session and other record",
			req: types.ScopeRequest{
				SessionAddr: data.SessionIDs[0][0].String(),
				RecordAddr:  data.RecordIDs[2][1][1].String(),
			},
			expErr: "session " + data.SessionIDs[0][0].String() + " and record " +
				data.RecordIDs[2][1][1].String() + " are not associated with the same scope: invalid request",
		},
		{
			name: "unknown scope uuid",
			req:  types.ScopeRequest{ScopeId: newTestUUID(21).String()},
			expResp: &types.ScopeResponse{
				Scope: types.WrapScopeNotFound(types.ScopeMetadataAddress(newTestUUID(21))),
			},
		},
		{
			name: "unknown scope addr",
			req:  types.ScopeRequest{ScopeId: types.ScopeMetadataAddress(newTestUUID(22)).String()},
			expResp: &types.ScopeResponse{
				Scope: types.WrapScopeNotFound(types.ScopeMetadataAddress(newTestUUID(22))),
			},
		},
		{
			name: "unknown session",
			req:  types.ScopeRequest{SessionAddr: types.SessionMetadataAddress(newTestUUID(23), newTestUUID(24)).String()},
			expResp: &types.ScopeResponse{
				Scope: types.WrapScopeNotFound(types.ScopeMetadataAddress(newTestUUID(23))),
			},
		},
		{
			name: "unknown record",
			req:  types.ScopeRequest{RecordAddr: types.RecordMetadataAddress(newTestUUID(25), newTestUUID(26).String()).String()},
			expResp: &types.ScopeResponse{
				Scope: types.WrapScopeNotFound(types.ScopeMetadataAddress(newTestUUID(25))),
			},
		},
		{
			name: "unknown scope: include request",
			req: types.ScopeRequest{
				ScopeId:        types.ScopeMetadataAddress(newTestUUID(27)).String(),
				IncludeRequest: true,
			},
			expResp: &types.ScopeResponse{
				Scope: types.WrapScopeNotFound(types.ScopeMetadataAddress(newTestUUID(27))),
				Request: &types.ScopeRequest{
					ScopeId:        types.ScopeMetadataAddress(newTestUUID(27)).String(),
					IncludeRequest: true,
				},
			},
		},
		{
			name:    "just scope: scope uuid",
			req:     types.ScopeRequest{ScopeId: data.ScopeUUIDs[0].String()},
			expResp: &types.ScopeResponse{Scope: types.WrapScope(data.Scopes[0], true)},
		},
		{
			name:    "just scope: scope addr",
			req:     types.ScopeRequest{ScopeId: data.ScopeIDs[1].String()},
			expResp: &types.ScopeResponse{Scope: types.WrapScope(data.Scopes[1], true)},
		},
		{
			name:    "just scope: session addr",
			req:     types.ScopeRequest{SessionAddr: data.SessionIDs[2][0].String()},
			expResp: &types.ScopeResponse{Scope: types.WrapScope(data.Scopes[2], true)},
		},
		{
			name:    "just scope: record addr",
			req:     types.ScopeRequest{RecordAddr: data.RecordIDs[1][0][1].String()},
			expResp: &types.ScopeResponse{Scope: types.WrapScope(data.Scopes[1], true)},
		},
		{
			name:    "just scope: scope uuid and session",
			req:     types.ScopeRequest{ScopeId: data.ScopeUUIDs[2].String(), SessionAddr: data.SessionIDs[2][1].String()},
			expResp: &types.ScopeResponse{Scope: types.WrapScope(data.Scopes[2], true)},
		},
		{
			name:    "just scope: scope addr and session",
			req:     types.ScopeRequest{ScopeId: data.ScopeIDs[0].String(), SessionAddr: data.SessionIDs[0][0].String()},
			expResp: &types.ScopeResponse{Scope: types.WrapScope(data.Scopes[0], true)},
		},
		{
			name:    "just scope: scope uuid and record",
			req:     types.ScopeRequest{ScopeId: data.ScopeUUIDs[1].String(), RecordAddr: data.RecordIDs[1][0][0].String()},
			expResp: &types.ScopeResponse{Scope: types.WrapScope(data.Scopes[1], true)},
		},
		{
			name:    "just scope: scope addr and record",
			req:     types.ScopeRequest{ScopeId: data.ScopeIDs[1].String(), RecordAddr: data.RecordIDs[1][0][1].String()},
			expResp: &types.ScopeResponse{Scope: types.WrapScope(data.Scopes[1], true)},
		},
		{
			name:    "just scope: session and record from same session",
			req:     types.ScopeRequest{SessionAddr: data.SessionIDs[0][0].String(), RecordAddr: data.RecordIDs[0][0][0].String()},
			expResp: &types.ScopeResponse{Scope: types.WrapScope(data.Scopes[0], true)},
		},
		{
			name:    "just scope: session and record from other session but same scope",
			req:     types.ScopeRequest{SessionAddr: data.SessionIDs[2][0].String(), RecordAddr: data.RecordIDs[2][1][0].String()},
			expResp: &types.ScopeResponse{Scope: types.WrapScope(data.Scopes[2], true)},
		},
		{
			name: "just scope: scope uuid and session and record",
			req: types.ScopeRequest{
				ScopeId:     data.ScopeUUIDs[1].String(),
				SessionAddr: data.SessionIDs[1][0].String(),
				RecordAddr:  data.RecordIDs[1][0][1].String(),
			},
			expResp: &types.ScopeResponse{Scope: types.WrapScope(data.Scopes[1], true)},
		},
		{
			name: "just scope: scope addr and session and record",
			req: types.ScopeRequest{
				ScopeId:     data.ScopeIDs[2].String(),
				SessionAddr: data.SessionIDs[2][1].String(),
				RecordAddr:  data.RecordIDs[2][1][1].String(),
			},
			expResp: &types.ScopeResponse{Scope: types.WrapScope(data.Scopes[2], true)},
		},
		{
			name: "with sessions: scope uuid",
			req:  types.ScopeRequest{ScopeId: data.ScopeUUIDs[0].String(), IncludeSessions: true},
			expResp: &types.ScopeResponse{
				Scope:    types.WrapScope(data.Scopes[0], true),
				Sessions: wrapSessions(data.AllScopeSessions[0], true),
			},
		},
		{
			name: "with sessions: scope addr",
			req:  types.ScopeRequest{ScopeId: data.ScopeUUIDs[2].String(), IncludeSessions: true},
			expResp: &types.ScopeResponse{
				Scope:    types.WrapScope(data.Scopes[2], true),
				Sessions: wrapSessions(data.AllScopeSessions[2], true),
			},
		},
		{
			name: "with sessions: session",
			req:  types.ScopeRequest{SessionAddr: data.SessionIDs[1][0].String(), IncludeSessions: true},
			expResp: &types.ScopeResponse{
				Scope:    types.WrapScope(data.Scopes[1], true),
				Sessions: wrapSessions(data.AllScopeSessions[1], true),
			},
		},
		{
			name: "with sessions: record",
			req:  types.ScopeRequest{RecordAddr: data.RecordIDs[2][0][0].String(), IncludeSessions: true},
			expResp: &types.ScopeResponse{
				Scope:    types.WrapScope(data.Scopes[2], true),
				Sessions: wrapSessions(data.AllScopeSessions[2], true),
			},
		},
		{
			name: "with records: scope uuid",
			req:  types.ScopeRequest{ScopeId: data.ScopeUUIDs[2].String(), IncludeRecords: true},
			expResp: &types.ScopeResponse{
				Scope:   types.WrapScope(data.Scopes[2], true),
				Records: wrapRecords(data.AllScopeRecords[2], true),
			},
		},
		{
			name: "with records: scope addr",
			req:  types.ScopeRequest{ScopeId: data.ScopeIDs[1].String(), IncludeRecords: true},
			expResp: &types.ScopeResponse{
				Scope:   types.WrapScope(data.Scopes[1], true),
				Records: wrapRecords(data.AllScopeRecords[1], true),
			},
		},
		{
			name: "with records: session",
			req:  types.ScopeRequest{SessionAddr: data.SessionIDs[2][0].String(), IncludeRecords: true},
			expResp: &types.ScopeResponse{
				Scope: types.WrapScope(data.Scopes[2], true),
				// Should be all the records in the scope, not just the ones in the provided session.
				Records: wrapRecords(data.AllScopeRecords[2], true),
			},
		},
		{
			name: "with records: record",
			req:  types.ScopeRequest{RecordAddr: data.RecordIDs[0][0][0].String(), IncludeRecords: true},
			expResp: &types.ScopeResponse{
				Scope:   types.WrapScope(data.Scopes[0], true),
				Records: wrapRecords(data.AllScopeRecords[0], true),
			},
		},
		{
			name: "with request: scope uuid",
			req:  types.ScopeRequest{ScopeId: data.ScopeUUIDs[0].String(), IncludeRequest: true},
			expResp: &types.ScopeResponse{
				Scope:   types.WrapScope(data.Scopes[0], true),
				Request: &types.ScopeRequest{ScopeId: data.ScopeUUIDs[0].String(), IncludeRequest: true},
			},
		},
		{
			name: "with request: scope addr",
			req:  types.ScopeRequest{ScopeId: data.ScopeIDs[1].String(), IncludeRequest: true},
			expResp: &types.ScopeResponse{
				Scope:   types.WrapScope(data.Scopes[1], true),
				Request: &types.ScopeRequest{ScopeId: data.ScopeIDs[1].String(), IncludeRequest: true},
			},
		},
		{
			name: "with request: session",
			req:  types.ScopeRequest{SessionAddr: data.SessionIDs[2][0].String(), IncludeRequest: true},
			expResp: &types.ScopeResponse{
				Scope:   types.WrapScope(data.Scopes[2], true),
				Request: &types.ScopeRequest{SessionAddr: data.SessionIDs[2][0].String(), IncludeRequest: true},
			},
		},
		{
			name: "with request: record",
			req:  types.ScopeRequest{RecordAddr: data.RecordIDs[2][1][0].String(), IncludeRequest: true},
			expResp: &types.ScopeResponse{
				Scope:   types.WrapScope(data.Scopes[2], true),
				Request: &types.ScopeRequest{RecordAddr: data.RecordIDs[2][1][0].String(), IncludeRequest: true},
			},
		},
		{
			name:    "no id info: scope uuid",
			req:     types.ScopeRequest{ScopeId: data.ScopeUUIDs[1].String(), ExcludeIdInfo: true},
			expResp: &types.ScopeResponse{Scope: types.WrapScope(data.Scopes[1], false)},
		},
		{
			name:    "no id info: scope addr",
			req:     types.ScopeRequest{ScopeId: data.ScopeIDs[0].String(), ExcludeIdInfo: true},
			expResp: &types.ScopeResponse{Scope: types.WrapScope(data.Scopes[0], false)},
		},
		{
			name:    "no id info: session",
			req:     types.ScopeRequest{SessionAddr: data.SessionIDs[1][0].String(), ExcludeIdInfo: true},
			expResp: &types.ScopeResponse{Scope: types.WrapScope(data.Scopes[1], false)},
		},
		{
			name:    "no id info: record",
			req:     types.ScopeRequest{RecordAddr: data.RecordIDs[2][0][0].String(), ExcludeIdInfo: true},
			expResp: &types.ScopeResponse{Scope: types.WrapScope(data.Scopes[2], false)},
		},
		{
			name: "with sessions, records, request and no id info: scope uuid",
			req: types.ScopeRequest{
				ScopeId:         data.ScopeUUIDs[0].String(),
				IncludeSessions: true,
				IncludeRecords:  true,
				ExcludeIdInfo:   true,
				IncludeRequest:  true,
			},
			expResp: &types.ScopeResponse{
				Scope:    types.WrapScope(data.Scopes[0], false),
				Sessions: wrapSessions(data.AllScopeSessions[0], false),
				Records:  wrapRecords(data.AllScopeRecords[0], false),
				Request: &types.ScopeRequest{
					ScopeId:         data.ScopeUUIDs[0].String(),
					IncludeSessions: true,
					IncludeRecords:  true,
					ExcludeIdInfo:   true,
					IncludeRequest:  true,
				},
			},
		},
		{
			name: "with sessions, records, request and no id info: scope addr",
			req: types.ScopeRequest{
				ScopeId:         data.ScopeIDs[1].String(),
				IncludeSessions: true,
				IncludeRecords:  true,
				ExcludeIdInfo:   true,
				IncludeRequest:  true,
			},
			expResp: &types.ScopeResponse{
				Scope:    types.WrapScope(data.Scopes[1], false),
				Sessions: wrapSessions(data.AllScopeSessions[1], false),
				Records:  wrapRecords(data.AllScopeRecords[1], false),
				Request: &types.ScopeRequest{
					ScopeId:         data.ScopeIDs[1].String(),
					IncludeSessions: true,
					IncludeRecords:  true,
					ExcludeIdInfo:   true,
					IncludeRequest:  true,
				},
			},
		},
		{
			name: "with sessions, records, request and no id info: session",
			req: types.ScopeRequest{
				SessionAddr:     data.SessionIDs[2][0].String(),
				IncludeSessions: true,
				IncludeRecords:  true,
				ExcludeIdInfo:   true,
				IncludeRequest:  true,
			},
			expResp: &types.ScopeResponse{
				Scope:    types.WrapScope(data.Scopes[2], false),
				Sessions: wrapSessions(data.AllScopeSessions[2], false),
				Records:  wrapRecords(data.AllScopeRecords[2], false),
				Request: &types.ScopeRequest{
					SessionAddr:     data.SessionIDs[2][0].String(),
					IncludeSessions: true,
					IncludeRecords:  true,
					ExcludeIdInfo:   true,
					IncludeRequest:  true,
				},
			},
		},
		{
			name: "with sessions, records, request and no id info: record",
			req: types.ScopeRequest{
				RecordAddr:      data.RecordIDs[2][1][0].String(),
				IncludeSessions: true,
				IncludeRecords:  true,
				ExcludeIdInfo:   true,
				IncludeRequest:  true,
			},
			expResp: &types.ScopeResponse{
				Scope:    types.WrapScope(data.Scopes[2], false),
				Sessions: wrapSessions(data.AllScopeSessions[2], false),
				Records:  wrapRecords(data.AllScopeRecords[2], false),
				Request: &types.ScopeRequest{
					RecordAddr:      data.RecordIDs[2][1][0].String(),
					IncludeSessions: true,
					IncludeRecords:  true,
					ExcludeIdInfo:   true,
					IncludeRequest:  true,
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var actResp *types.ScopeResponse
			var err error
			testFunc := func() {
				actResp, err = s.queryClient.Scope(gocontext.Background(), &tc.req)
			}
			s.Require().NotPanics(testFunc, "queryClient.Scope(...)")
			s.AssertErrorValue(err, tc.expErr, "error from queryClient.Scope(...)")
			if s.Assert().Equal(tc.expResp, actResp, "response from queryClient.Scope(...)") || tc.expResp == nil || actResp == nil {
				// If they're not equal and both not nil, I want to run some extra tests to maybe help identify what's wrong.
				// But if they're equal, we're all good. And if either is nil, that'll be obvious in the failure message, so we're done.
				return
			}

			expScopeName := data.IdentifyScope(tc.expResp.Scope)
			actScopeName := data.IdentifyScope(actResp.Scope)
			if s.Assert().Equal(expScopeName, actScopeName, "names for response.Scope") {
				// If those are equal, just make sure the actual scopes are equal too.
				s.Assert().Equal(tc.expResp.Scope, actResp.Scope, "response.Scope")
			}

			expSessionNames := data.IdentifySessions(tc.expResp.Sessions)
			actSessionNames := data.IdentifySessions(actResp.Sessions)
			if s.Assert().Equal(expSessionNames, actSessionNames, "names for response.Sessions") {
				// If those are equal, make sure each individual entry is equal too (just to be safe).
				for i := range tc.expResp.Sessions {
					s.Assert().Equal(tc.expResp.Sessions[i], actResp.Sessions[i], "response.Sessions[%d]", i)
				}
			}

			expRecordNames := data.IdentifyRecords(tc.expResp.Records)
			actRecordNames := data.IdentifyRecords(actResp.Records)
			if s.Assert().Equal(expRecordNames, actRecordNames, "names for response.Records") {
				for i := range tc.expResp.Records {
					s.Assert().Equal(tc.expResp.Records[i], actResp.Records[i], "response.Records[%d]", i)
				}
			}

			// The request is all strings and bools so it'll look just fine in the failure output.
			s.Assert().Equal(tc.expResp.Request, actResp.Request, "response.Request")
		})
	}
}

func (s *QueryServerTestSuite) TestScopesAll() {
	// six scopes with various numbers of sessions and records.
	data := s.createData([][]int{{1}, {2}, {3}, {2, 1}, {1, 1, 1}, {1, 2}})
	// Valid Scopes indexes (6): [0] [1] [2] [3] [4] [5]

	tests := []struct {
		name    string
		req     types.ScopesAllRequest
		expResp *types.ScopesAllResponse
		expErr  string
	}{
		{
			name:    "empty request",
			req:     types.ScopesAllRequest{},
			expResp: &types.ScopesAllResponse{Scopes: wrapScopes(data.AllScopes, true)},
		},
		{
			name:    "exclude id info",
			req:     types.ScopesAllRequest{ExcludeIdInfo: true},
			expResp: &types.ScopesAllResponse{Scopes: wrapScopes(data.AllScopes, false)},
		},
		{
			name: "include request",
			req:  types.ScopesAllRequest{IncludeRequest: true},
			expResp: &types.ScopesAllResponse{
				Scopes:  wrapScopes(data.AllScopes, true),
				Request: nil, // The test runner will populate this for us.
			},
		},
		{
			name: "exclude id info and include request",
			req:  types.ScopesAllRequest{ExcludeIdInfo: true, IncludeRequest: true},
			expResp: &types.ScopesAllResponse{
				Scopes:  wrapScopes(data.AllScopes, false),
				Request: nil, // The test runner will populate this for us.
			},
		},
		{
			name: "limit 2, offset 1",
			req:  types.ScopesAllRequest{Pagination: &query.PageRequest{Offset: 1, Limit: 2}},
			expResp: &types.ScopesAllResponse{
				Scopes:     wrapScopes(data.AllScopes[1:3], true),
				Pagination: &query.PageResponse{NextKey: data.ScopeUUIDs[3][:]},
			},
		},
		{
			name: "next key of 3rd, limit 2",
			req:  types.ScopesAllRequest{Pagination: &query.PageRequest{Limit: 2, Key: data.ScopeUUIDs[2][:]}},
			expResp: &types.ScopesAllResponse{
				Scopes:     wrapScopes(data.AllScopes[2:4], true),
				Pagination: &query.PageResponse{NextKey: data.ScopeUUIDs[4][:]},
			},
		},
		{
			name: "limit 1, reversed",
			req:  types.ScopesAllRequest{Pagination: &query.PageRequest{Limit: 1, Reverse: true}},
			expResp: &types.ScopesAllResponse{
				Scopes:     wrapScopes(data.AllScopes[5:6], true),
				Pagination: &query.PageResponse{NextKey: data.ScopeUUIDs[4][:]},
			},
		},
		{
			name: "limit 3, count total, exclude id info, include request",
			req: types.ScopesAllRequest{
				ExcludeIdInfo:  true,
				IncludeRequest: true,
				Pagination:     &query.PageRequest{Limit: 3, CountTotal: true},
			},
			expResp: &types.ScopesAllResponse{
				Scopes:     wrapScopes(data.AllScopes[0:3], false),
				Request:    nil, // The test runner will populate this for us.
				Pagination: &query.PageResponse{Total: 6, NextKey: data.ScopeUUIDs[3][:]},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.req.IncludeRequest && tc.expResp != nil {
				tc.expResp.Request = &tc.req
			}
			if tc.expResp != nil && tc.expResp.Pagination == nil {
				tc.expResp.Pagination = &query.PageResponse{}
			}

			var actResp *types.ScopesAllResponse
			var err error
			testFunc := func() {
				actResp, err = s.queryClient.ScopesAll(gocontext.Background(), &tc.req)
			}
			s.Require().NotPanics(testFunc, "queryClient.ScopesAll(...)")
			s.AssertErrorValue(err, tc.expErr, "error from queryClient.ScopesAll(...)")
			if s.Assert().Equal(tc.expResp, actResp, "response from queryClient.ScopesAll(...)") || tc.expResp == nil || actResp == nil {
				// If they're not equal and both not nil, I want to run some extra tests to maybe help identify what's wrong.
				// But if they're equal, we're all good. And if either is nil, that'll be obvious in the failure message, so we're done.
				return
			}

			expScopeNames := data.IdentifyScopes(tc.expResp.Scopes)
			actScopeNames := data.IdentifyScopes(actResp.Scopes)
			if s.Assert().Equal(expScopeNames, actScopeNames, "names for response.Scopes") {
				// If those are equal, make sure each individual entry is equal too (just to be safe).
				for i := range tc.expResp.Scopes {
					s.Assert().Equal(tc.expResp.Scopes[i], actResp.Scopes[i], "response.Scopes[%d]", i)
				}
			}

			s.Assert().Equal(tc.expResp.Request, actResp.Request, "response.Request")
			s.AssertEqualPageResponses(tc.expResp.Pagination, actResp.Pagination, "response.Pagination")
		})
	}
}

func (s *QueryServerTestSuite) TestSessionsQuery() {
	app, ctx, queryClient := s.app, s.ctx, s.queryClient

	scope := types.NewScope(s.scopeID, s.scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, s.user1, false)
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
		// 	err:  "empty request",
		// },
		{
			name: "empty request",
			req:  &types.SessionsRequest{},
			err:  "empty request parameters: invalid request",
		},

		// only scope id
		{
			name: "only scope id invalid - error",
			req:  &types.SessionsRequest{ScopeId: "6332c1a4-foo1-bare-895b-invalid65cb6"},
			err:  "could not parse [6332c1a4-foo1-bare-895b-invalid65cb6] into either a scope address (decoding bech32 failed: invalid character not part of charset: 45) or uuid (invalid UUID format): invalid request",
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
			err:  "could not parse [not-a-valid-session-id] into a session address: decoding bech32 failed: invalid separator index -1: invalid request",
		},
		{
			name: "only session id as uuid - error",
			req:  &types.SessionsRequest{SessionId: unknownUUID.String()},
			err:  fmt.Sprintf("could not parse [%s] into a session address: decoding bech32 failed: invalid separator index 35: invalid request", unknownUUID),
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
			err:  "could not parse [not-a-valid-record-id] into a record address: decoding bech32 failed: invalid separator index -1: invalid request",
		},
		{
			name: "only record addr not found - error",
			req:  &types.SessionsRequest{RecordAddr: types.RecordMetadataAddress(s.scopeUUID, "no-such-record").String()},
			err:  fmt.Sprintf("record %s does not exist: invalid request", types.RecordMetadataAddress(s.scopeUUID, "no-such-record")),
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
			err:  "a scope is required to look up sessions by record name: invalid request",
		},

		// scope id and session id
		{
			name: "scope id invalid session id ok - error",
			req:  &types.SessionsRequest{ScopeId: "not-a-scope-id", SessionId: s.sessionID.String()},
			err:  "could not parse [not-a-scope-id] into either a scope address (decoding bech32 failed: invalid separator index -1) or uuid (invalid UUID length: 14): invalid request",
		},
		{
			name: "scope id as uuid exists session id invalid - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeUUID.String(), SessionId: "invalidSessionID"},
			err:  "could not parse [invalidSessionID] into either a session address (decoding bech32 failed: string not all lowercase or all uppercase) or uuid (invalid UUID length: 16): invalid request",
		},
		{
			name: "scope id as addr exists session id invalid - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: "invalidSessionID"},
			err:  "could not parse [invalidSessionID] into either a session address (decoding bech32 failed: string not all lowercase or all uppercase) or uuid (invalid UUID length: 16): invalid request",
		},
		{
			name: "scope id as uuid exists session id as addr wrong scope - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeUUID.String(), SessionId: types.SessionMetadataAddress(unknownUUID, s.sessionUUID).String()},
			err: fmt.Sprintf("session %s is not in scope %s: invalid request",
				types.SessionMetadataAddress(unknownUUID, s.sessionUUID), s.scopeID),
		},
		{
			name: "scope id as addr exists session id as addr wrong scope - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: types.SessionMetadataAddress(unknownUUID, s.sessionUUID).String()},
			err: fmt.Sprintf("session %s is not in scope %s: invalid request",
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
			err:  "could not parse [notReallyAScopeID] into either a scope address (decoding bech32 failed: string not all lowercase or all uppercase) or uuid (invalid UUID length: 17): invalid request",
		},
		{
			name: "scope id exists record addr invalid - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), RecordAddr: "invalid-record-addr"},
			err:  "could not parse [invalid-record-addr] into a record address: decoding bech32 failed: invalid separator index -1: invalid request",
		},
		{
			name: "scope id exists record addr wrong scope - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), RecordAddr: types.RecordMetadataAddress(unknownUUID, s.recordName).String()},
			err:  fmt.Sprintf("record %s is not part of scope %s: invalid request", types.RecordMetadataAddress(unknownUUID, s.recordName), s.scopeID),
		},
		{
			name: "scope id exists record addr not found - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), RecordAddr: types.RecordMetadataAddress(s.scopeUUID, "record-not-real").String()},
			err:  fmt.Sprintf("record %s does not exist: invalid request", types.RecordMetadataAddress(s.scopeUUID, "record-not-real")),
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
			err:  "could not parse [illegitimate-scope] into either a scope address (decoding bech32 failed: invalid separator index -1) or uuid (invalid UUID length: 18): invalid request",
		},
		{
			name: "scope id as uuid not found record name ok - error",
			req:  &types.SessionsRequest{ScopeId: unknownUUID.String(), RecordName: s.recordName},
			err:  fmt.Sprintf("record %s does not exist: invalid request", types.RecordMetadataAddress(unknownUUID, s.recordName)),
		},
		{
			name: "scope id as addr not found record name ok - error",
			req:  &types.SessionsRequest{ScopeId: types.ScopeMetadataAddress(unknownUUID).String(), RecordName: s.recordName},
			err:  fmt.Sprintf("record %s does not exist: invalid request", types.RecordMetadataAddress(unknownUUID, s.recordName)),
		},
		{
			name: "scope id as uuid exists record name not found - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeUUID.String(), RecordName: "no-such-record"},
			err:  fmt.Sprintf("record %s does not exist: invalid request", types.RecordMetadataAddress(s.scopeUUID, "no-such-record")),
		},
		{
			name: "scope id as addr exists record name not found - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), RecordName: "still-no-such-record"},
			err:  fmt.Sprintf("record %s does not exist: invalid request", types.RecordMetadataAddress(s.scopeUUID, "still-no-such-record")),
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
			err:  "could not parse [lol-nope-notASessionId] into either a session address (decoding bech32 failed: string not all lowercase or all uppercase) or uuid (invalid UUID length: 22): invalid request",
		},
		{
			name: "session id exists record addr invalid - error",
			req:  &types.SessionsRequest{SessionId: s.sessionID.String(), RecordAddr: "yeah-not-a-record-addr"},
			err:  "could not parse [yeah-not-a-record-addr] into a record address: decoding bech32 failed: invalid separator index -1: invalid request",
		},
		{
			name: "session id exists record addr wrong scope - error",
			req:  &types.SessionsRequest{SessionId: s.sessionID.String(), RecordAddr: types.RecordMetadataAddress(unknownUUID, s.recordName).String()},
			err:  fmt.Sprintf("session %s is not in scope %s: invalid request", s.sessionID, types.ScopeMetadataAddress(unknownUUID)),
		},
		{
			name: "session id exists record addr wrong session - error",
			req:  &types.SessionsRequest{SessionId: s.sessionID.String(), RecordAddr: anotherRecordID.String()},
			err:  fmt.Sprintf("record %s belongs to session %s (not %s): invalid request", anotherRecordID, anotherSessionID, s.sessionID),
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
			err:  "could not parse [nopenope-nope-nope-nota-sessionuuidx] into a session address: decoding bech32 failed: invalid separator index -1: invalid request",
		},
		{
			name: "session id as uuid record name ok - error",
			req:  &types.SessionsRequest{SessionId: unknownUUID.String(), RecordName: s.recordName},
			err:  fmt.Sprintf("could not parse [%s] into a session address: decoding bech32 failed: invalid separator index 35: invalid request", unknownUUID),
		},
		{
			name: "session id as addr not found record name ok - error",
			req:  &types.SessionsRequest{SessionId: types.SessionMetadataAddress(unknownUUID, unknownUUID).String(), RecordName: s.recordName},
			err:  fmt.Sprintf("record %s does not exist: invalid request", types.RecordMetadataAddress(unknownUUID, s.recordName)),
		},
		{
			name: "session id as addr exists record name not found - error",
			req:  &types.SessionsRequest{SessionId: s.sessionID.String(), RecordName: "not-gonna-find-this-record"},
			err:  fmt.Sprintf("record %s does not exist: invalid request", types.RecordMetadataAddress(s.scopeUUID, "not-gonna-find-this-record")),
		},
		{
			name: "session id as addr exists record name wrong session - error",
			req:  &types.SessionsRequest{SessionId: s.sessionID.String(), RecordName: anotherRecordName},
			err: fmt.Sprintf("record %s belongs to session %s (not %s): invalid request",
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
			err:  "could not parse [abby-lane] into a record address: decoding bech32 failed: invalid separator index -1: invalid request",
		},
		{
			name: "record addr exists record name wrong - error",
			req:  &types.SessionsRequest{RecordAddr: s.recordID.String(), RecordName: anotherRecordName},
			err:  fmt.Sprintf("record %s does not have name %s: invalid request", s.recordID, anotherRecordName),
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
			err:  "could not parse [bad scope] into either a scope address (decoding bech32 failed: invalid character in string: ' ') or uuid (invalid UUID length: 9): invalid request",
		},
		{
			name: "scope id exists session id invalid record addr ok - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: "bad session", RecordAddr: s.recordID.String()},
			err:  "could not parse [bad session] into either a session address (decoding bech32 failed: invalid character in string: ' ') or uuid (invalid UUID length: 11): invalid request",
		},
		{
			name: "scope id exists session id exists record addr invalid - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: s.sessionID.String(), RecordAddr: "bad record"},
			err:  "could not parse [bad record] into a record address: decoding bech32 failed: invalid character in string: ' ': invalid request",
		},
		{
			name: "scope id wrong scope session id exists record addr exists - error",
			req:  &types.SessionsRequest{ScopeId: types.ScopeMetadataAddress(unknownUUID).String(), SessionId: s.sessionID.String(), RecordAddr: s.recordID.String()},
			err:  fmt.Sprintf("record %s is not part of scope %s: invalid request", s.recordID, types.ScopeMetadataAddress(unknownUUID)),
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
			err:  "could not parse [nopescope] into either a scope address (decoding bech32 failed: invalid separator index -1) or uuid (invalid UUID length: 9): invalid request",
		},
		{
			name: "scope id exists session id invalid record name ok - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: "nosesh", RecordName: s.recordName},
			err:  "could not parse [nosesh] into either a session address (decoding bech32 failed: invalid bech32 string length 6) or uuid (invalid UUID length: 6): invalid request",
		},
		{
			name: "scope id exists session id exists record name not found - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: s.sessionID.String(), RecordName: "nopenopenope"},
			err:  fmt.Sprintf("record %s does not exist: invalid request", types.RecordMetadataAddress(s.scopeUUID, "nopenopenope")),
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
			err:  "could not parse [veryBadScope] into either a scope address (decoding bech32 failed: string not all lowercase or all uppercase) or uuid (invalid UUID length: 12): invalid request",
		},
		{
			name: "scope id exists record addr invalid record name ok - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), RecordAddr: "kindofbadrecord", RecordName: s.recordName},
			err:  "could not parse [kindofbadrecord] into a record address: decoding bech32 failed: invalid separator index -1: invalid request",
		},
		{
			name: "scope id exists record addr exists record name wrong - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), RecordAddr: s.recordID.String(), RecordName: "wrongagain"},
			err:  fmt.Sprintf("record %s does not have name wrongagain: invalid request", s.recordID),
		},
		{
			name: "scope id exists record addr wrong scope record name ok - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), RecordAddr: types.RecordMetadataAddress(unknownUUID, s.recordName).String(), RecordName: s.recordName},
			err:  fmt.Sprintf("record %s is not part of scope %s: invalid request", types.RecordMetadataAddress(unknownUUID, s.recordName), s.scopeID),
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
			err:  "could not parse [nopesess] into either a session address (decoding bech32 failed: invalid separator index -1) or uuid (invalid UUID length: 8): invalid request",
		},
		{
			name: "session id exists record addr invalid record name ok - error",
			req:  &types.SessionsRequest{SessionId: s.sessionID.String(), RecordAddr: "incorrect-record-id", RecordName: s.recordName},
			err:  "could not parse [incorrect-record-id] into a record address: decoding bech32 failed: invalid separator index -1: invalid request",
		},
		{
			name: "session id exists record addr wrong scope record name ok - error",
			req:  &types.SessionsRequest{SessionId: s.sessionID.String(), RecordAddr: types.RecordMetadataAddress(unknownUUID, s.recordName).String(), RecordName: s.recordName},
			err:  fmt.Sprintf("session %s is not in scope %s: invalid request", s.sessionID, types.ScopeMetadataAddress(unknownUUID)),
		},
		{
			name: "session id exists record addr wrong session record name ok - error",
			req:  &types.SessionsRequest{SessionId: s.sessionID.String(), RecordAddr: s.recordID.String(), RecordName: anotherRecordName},
			err:  fmt.Sprintf("record %s does not have name another-record: invalid request", s.recordID),
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
			err:  "could not parse [negatoryscope] into either a scope address (decoding bech32 failed: invalid separator index -1) or uuid (invalid UUID length: 13): invalid request",
		},
		{
			name: "scope id exists session id invalid record addr ok record name ok - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: "negatorysession", RecordAddr: s.recordID.String(), RecordName: s.recordName},
			err:  "could not parse [negatorysession] into either a session address (decoding bech32 failed: invalid separator index -1) or uuid (invalid UUID length: 15): invalid request",
		},
		{
			name: "scope id exists session id exists record addr invalid record name ok - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: s.sessionID.String(), RecordAddr: "negatoryrecord", RecordName: s.recordName},
			err:  "could not parse [negatoryrecord] into a record address: decoding bech32 failed: invalid separator index -1: invalid request",
		},
		{
			name: "scope id exists session id exists record addr exists record name wrong - error",
			req:  &types.SessionsRequest{ScopeId: s.scopeID.String(), SessionId: s.sessionID.String(), RecordAddr: s.recordID.String(), RecordName: anotherRecordName},
			err:  fmt.Sprintf("record %s does not have name %s: invalid request", s.recordID, anotherRecordName),
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
	s.T().Run("include both scope and records", func(t *testing.T) {
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
	s.T().Run("include both scope and records without id info", func(t *testing.T) {
		req := types.SessionsRequest{SessionId: s.sessionID.String(), IncludeScope: true, IncludeRecords: true, ExcludeIdInfo: true}
		sr, err := queryClient.Sessions(gocontext.Background(), &req)
		require.NoErrorf(t, err, "unexpected error: %s", err)
		require.NotNil(t, sr, "result of Sessions query")

		if assert.NotNil(t, sr.Scope, "scope wrapper") {
			assert.NotNil(t, sr.Scope.Scope, "the scope")
			assert.Nil(t, sr.Scope.ScopeIdInfo, "scope id info")
			assert.Nil(t, sr.Scope.ScopeSpecIdInfo, "scope spec id info")
		}

		if assert.Len(t, sr.Sessions, 1, "sessions") {
			sw := sr.Sessions[0]
			if assert.NotNil(t, sw, "session wrapper") {
				assert.NotNil(t, sw.Session, "the session")
				assert.Nil(t, sw.SessionIdInfo, "session id info")
				assert.Nil(t, sw.ContractSpecIdInfo, "contract spec id info")
			}
		}

		if assert.Len(t, sr.Records, 1, "records") {
			rw := sr.Records[0]
			if assert.NotNil(t, rw, "record wrapper") {
				assert.NotNil(t, rw.Record, "the record")
				assert.Nil(t, rw.RecordIdInfo, "record id info")
				assert.Nil(t, rw.RecordSpecIdInfo, "record spec id info")
			}
		}
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
	s.EqualError(err, "empty request parameters: invalid request")

	_, err = queryClient.Records(gocontext.Background(), &types.RecordsRequest{ScopeId: "foo"})
	s.EqualError(err, "could not parse [foo] into either a scope address (decoding bech32 failed: invalid bech32 string length 3) or uuid (invalid UUID length: 3): invalid request")

	_, err = queryClient.Records(gocontext.Background(), &types.RecordsRequest{ScopeId: "6332c1a4-foo1-bare-895b-invalid65cb6"})
	s.EqualError(err, "could not parse [6332c1a4-foo1-bare-895b-invalid65cb6] into either a scope address (decoding bech32 failed: invalid character not part of charset: 45) or uuid (invalid UUID format): invalid request")

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

func (s *QueryServerTestSuite) TestScopeSpecificationQuery() {
	app, ctx, queryClient := s.app, s.ctx, s.queryClient

	scopeSpec := types.NewScopeSpecification(
		s.scopeSpecID,
		types.NewDescription("test-scope-spec", "testing", "https://provenance.io", ""),
		[]string{s.user1},
		[]types.PartyType{types.PartyType_PARTY_TYPE_AFFILIATE},
		[]types.MetadataAddress{s.cSpecID},
	)
	app.MetadataKeeper.SetScopeSpecification(ctx, *scopeSpec)

	contractSpec := types.NewContractSpecification(
		s.cSpecID,
		types.NewDescription("test-contract-spec", "testing", "https://provenance.io", ""),
		[]string{s.user1},
		[]types.PartyType{types.PartyType_PARTY_TYPE_AFFILIATE},
		types.NewContractSpecificationSourceHash("hash"),
		"",
	)
	app.MetadataKeeper.SetContractSpecification(ctx, *contractSpec)

	recordSpec := types.NewRecordSpecification(
		s.recSpecID,
		"test-record-spec",
		[]*types.InputSpecification{
			types.NewInputSpecification(
				"record-name",
				"type-name",
				types.NewInputSpecificationSourceHash("hash"),
			),
		},
		"type-name",
		1,
		[]types.PartyType{types.PartyType_PARTY_TYPE_AFFILIATE},
	)
	app.MetadataKeeper.SetRecordSpecification(ctx, *recordSpec)

	// Run a couple tests on the Include* flags.
	s.T().Run("include contract specs", func(t *testing.T) {
		req := types.ScopeSpecificationRequest{SpecificationId: s.scopeSpecID.String(), IncludeContractSpecs: true}
		res, err := queryClient.ScopeSpecification(ctx, &req)
		require.NoErrorf(t, err, "unexpected error: %s", err)
		require.NotNil(t, res, "result of Scope Specification query")
		assert.Equal(t, 1, len(res.ContractSpecs), "number of contract specs")
		assert.Equal(t, s.cSpecID, res.ContractSpecs[0].Specification.SpecificationId, "contract spec id")
		assert.Equal(t, "test-contract-spec", res.ContractSpecs[0].Specification.Description.Name, "contract spec name")
		assert.Equal(t, 0, len(res.RecordSpecs), "number of record specs")
	})
	s.T().Run("include record specs", func(t *testing.T) {
		req := types.ScopeSpecificationRequest{SpecificationId: s.scopeSpecID.String(), IncludeRecordSpecs: true}
		res, err := queryClient.ScopeSpecification(ctx, &req)
		require.NoErrorf(t, err, "unexpected error: %s", err)
		require.NotNil(t, res, "result of Scope Specification query")
		assert.Equal(t, 1, len(res.RecordSpecs), "number of record specs")
		assert.Equal(t, s.recSpecID, res.RecordSpecs[0].Specification.SpecificationId, "record spec id")
		assert.Equal(t, "test-record-spec", res.RecordSpecs[0].Specification.Name, "record spec name")
		assert.Equal(t, 0, len(res.ContractSpecs), "number of contract specs")
	})
	s.T().Run("include both contract and record specs", func(t *testing.T) {
		req := types.ScopeSpecificationRequest{SpecificationId: s.scopeSpecID.String(), IncludeContractSpecs: true, IncludeRecordSpecs: true}
		res, err := queryClient.ScopeSpecification(ctx, &req)
		require.NoErrorf(t, err, "unexpected error: %s", err)
		require.NotNil(t, res, "result of Scope Specification query")
		// contract spec assertions
		assert.Equal(t, 1, len(res.ContractSpecs), "number of contract specs")
		assert.Equal(t, s.cSpecID, res.ContractSpecs[0].Specification.SpecificationId, "contract spec id")
		assert.Equal(t, "test-contract-spec", res.ContractSpecs[0].Specification.Description.Name, "contract spec name")
		// record spec assertions
		assert.Equal(t, 1, len(res.RecordSpecs), "number of record specs")
		assert.Equal(t, s.recSpecID, res.RecordSpecs[0].Specification.SpecificationId, "record spec id")
		assert.Equal(t, "test-record-spec", res.RecordSpecs[0].Specification.Name, "record spec name")
	})
}

// TODO: ScopeSpecificationsAll tests
// TODO: ContractSpecification tests
// TODO: ContractSpecificationsAll tests
// TODO: RecordSpecificationsForContractSpecification tests
// TODO: RecordSpecification tests
// TODO: RecordSpecificationsAll tests

func (s *QueryServerTestSuite) TestGetByAddr() {
	recordName := "myrecord"
	uuids := []uuid.UUID{
		uuid.MustParse("7D3D0BFA-44CF-4536-888F-DEBF28ACC887"),
		uuid.MustParse("30FE5269-27F6-44AC-B900-BD13879932C3"),
		uuid.MustParse("8FA0F897-28C5-43D0-AC55-CC950A79AC9E"),
		uuid.MustParse("036E7E4C-014C-4103-A9F4-5759981D9CE8"),
		uuid.MustParse("1A0C65EC-C512-43B0-8F90-9CCF59A54FDE"),
		uuid.MustParse("F757BA39-ED71-4CE5-9FE9-A77659CEAC40"),
		uuid.MustParse("3B5F6C64-8A08-4EEB-BAE5-A92D18332DA9"),
		uuid.MustParse("A8BC40DE-A280-4599-8C8A-7A00B0A47CBA"),
	}

	scopeID1 := types.ScopeMetadataAddress(uuids[0])               // scope1qp7n6zl6gn852d5g3l0t729vezrsq304wz
	sessionID1 := types.SessionMetadataAddress(uuids[0], uuids[1]) // session1q97n6zl6gn852d5g3l0t729vezrnpljjdynlv39vhyqt6yu8nyevxx7k828
	recordID1 := types.RecordMetadataAddress(uuids[0], recordName) // record1qf7n6zl6gn852d5g3l0t729vezr7hw7d4pu05x7tj9hjcr4sla9v6fxd4pw

	scopeID2 := types.ScopeMetadataAddress(uuids[2])               // scope1qz86p7yh9rz5859v2hxf2zne4j0qef5nle
	sessionID2 := types.SessionMetadataAddress(uuids[2], uuids[3]) // session1qx86p7yh9rz5859v2hxf2zne4j0qxmn7fsq5csgr4869wkvcrkwws2wefy6
	recordID2 := types.RecordMetadataAddress(uuids[2], recordName) // record1q286p7yh9rz5859v2hxf2zne4j0whw7d4pu05x7tj9hjcr4sla9v6dur3el

	scopeSpecID := types.ScopeSpecMetadataAddress(uuids[4])            // scopespec1qsdqce0vc5fy8vy0jzwv7kd9fl0qknw7xs
	cSpecID := types.ContractSpecMetadataAddress(uuids[5])             // contractspec1q0m40w3ea4c5eevlaxnhvkww43qq7adrfr
	recSpecID := types.RecordSpecMetadataAddress(uuids[5], recordName) // recspec1qhm40w3ea4c5eevlaxnhvkww43qwhw7d4pu05x7tj9hjcr4sla9v64sc2mh

	scopeIDDNE := types.ScopeMetadataAddress(uuids[6])                    // scope1qqa47mry3gyya6a6uk5j6xpn9k5skuvtxy
	sessionIDDNE := types.SessionMetadataAddress(uuids[6], uuids[7])      // session1qya47mry3gyya6a6uk5j6xpn9k5630zqm63gq3ve3j985q9s537t5q3pgyg
	recordIDDNE := types.RecordMetadataAddress(uuids[6], recordName)      // record1qga47mry3gyya6a6uk5j6xpn9k57hw7d4pu05x7tj9hjcr4sla9v6xkxzus
	scopeSpecIDDNE := types.ScopeSpecMetadataAddress(uuids[6])            // scopespec1qsa47mry3gyya6a6uk5j6xpn9k5sc649p3
	cSpecIDDNE := types.ContractSpecMetadataAddress(uuids[6])             // contractspec1qva47mry3gyya6a6uk5j6xpn9k5s6h5s78
	recSpecIDDNE := types.RecordSpecMetadataAddress(uuids[6], recordName) // recspec1q5a47mry3gyya6a6uk5j6xpn9k57hw7d4pu05x7tj9hjcr4sla9v6negyuq

	ownerAddr := sdk.AccAddress("ownerAddr___________").String() // cosmos1damkuetjg9jxgujlta047h6lta047h6lmp9nkx

	manyAddrs := make([]string, 10000)
	curUUID := uuid.MustParse("4D09C7B0-5B86-43BF-BA9F-33CA3D78C1DB")
	incCurUUID := func() {
		for i := 15; i >= 0; i-- {
			if curUUID[i] != 255 {
				curUUID[i] = curUUID[i] + 1
				break
			}
			curUUID[i] = 0
		}
	}
	for i := range manyAddrs {
		switch i % 7 {
		case 0:
			incCurUUID()
			manyAddrs[i] = types.ScopeMetadataAddress(curUUID).String()
		case 1:
			manyAddrs[i] = types.SessionMetadataAddress(curUUID, uuid.New()).String()
		case 2:
			manyAddrs[i] = types.RecordMetadataAddress(curUUID, recordName).String()
		case 3:
			manyAddrs[i] = types.ScopeSpecMetadataAddress(curUUID).String()
		case 4:
			manyAddrs[i] = types.ContractSpecMetadataAddress(curUUID).String()
		case 5:
			manyAddrs[i] = types.RecordSpecMetadataAddress(curUUID, recordName).String()
		case 6:
			manyAddrs[i] = curUUID.String()
		}
	}

	scopeSpec := types.ScopeSpecification{
		SpecificationId: scopeSpecID,
		Description: &types.Description{
			Name:        "myTestScopeSpec",
			Description: "my test scope spec",
			WebsiteUrl:  "http://example.com",
		},
		OwnerAddresses:  []string{ownerAddr},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		ContractSpecIds: []types.MetadataAddress{cSpecID},
	}
	cSpec := types.ContractSpecification{
		SpecificationId: cSpecID,
		Description: &types.Description{
			Name:        "myTestContractSpec",
			Description: "my test contract spec",
			WebsiteUrl:  "http://example.com",
		},
		OwnerAddresses:  scopeSpec.OwnerAddresses,
		PartiesInvolved: scopeSpec.PartiesInvolved,
		Source:          &types.ContractSpecification_Hash{Hash: "1845ca13f061fc68f2343bba7f6be3be"}, // printf 'mycspec' | md5sum
		ClassName:       "myclass",
	}
	recSpec := types.RecordSpecification{
		SpecificationId: recSpecID,
		Name:            recordName,
		Inputs: []*types.InputSpecification{
			{
				Name:     "myinput",
				TypeName: "string",
				Source:   &types.InputSpecification_Hash{Hash: "8bf572e7226b661b87d47458b5aec2a7"},
			},
		},
		TypeName:           "myRecordType",
		ResultType:         types.DefinitionType_DEFINITION_TYPE_PROPOSED,
		ResponsibleParties: cSpec.PartiesInvolved,
	}

	scope1 := types.Scope{
		ScopeId:           scopeID1,
		SpecificationId:   scopeSpecID,
		Owners:            []types.Party{{Address: ownerAddr, Role: types.PartyType_PARTY_TYPE_OWNER}},
		ValueOwnerAddress: ownerAddr,
	}
	scope2 := types.Scope{
		ScopeId:           scopeID2,
		SpecificationId:   scopeSpecID,
		Owners:            []types.Party{{Address: ownerAddr, Role: types.PartyType_PARTY_TYPE_OWNER}},
		ValueOwnerAddress: ownerAddr,
	}
	session1 := types.Session{
		SessionId:       sessionID1,
		SpecificationId: cSpecID,
		Parties:         scope1.Owners,
		Name:            "mysession",
	}
	session1.Audit = session1.Audit.UpdateAudit(time.Unix(1650450000, 0).UTC(), ownerAddr, "test update")
	session2 := types.Session{
		SessionId:       sessionID2,
		SpecificationId: cSpecID,
		Parties:         scope2.Owners,
		Name:            "mysession",
	}
	session2.Audit = session2.Audit.UpdateAudit(time.Unix(1650450001, 0).UTC(), ownerAddr, "test update")
	record1 := types.Record{
		Name:      recordName,
		SessionId: sessionID1,
		Process: types.Process{
			ProcessId: &types.Process_Hash{Hash: "528706d1dfad2c63d619297c005cc841"}, // printf 'myrecord' | md5sum
			Name:      "myproc",
			Method:    "mymethod",
		},
		Inputs: []types.RecordInput{
			{
				Name:     recSpec.Inputs[0].Name,
				Source:   &types.RecordInput_Hash{Hash: "8bf572e7226b661b87d47458b5aec2a7"}, // printf 'myinput' | md5sum
				TypeName: recSpec.Inputs[0].TypeName,
				Status:   types.RecordInputStatus_Proposed,
			},
		},
		Outputs: []types.RecordOutput{
			{
				Hash:   "b7aebd675ceb38fcb6c11289c6ae5216", // printf 'myoutput' | md5sum
				Status: types.ResultStatus_RESULT_STATUS_PASS,
			},
		},
		SpecificationId: recSpecID,
	}
	record2 := types.Record{
		Name:      recordName,
		SessionId: sessionID2,
		Process: types.Process{
			ProcessId: &types.Process_Hash{Hash: "ac39c3a10703a3a7fae08207216a56d1"}, // printf 'myrecord2' | md5sum
			Name:      "myproc",
			Method:    "mymethod",
		},
		Inputs: []types.RecordInput{
			{
				Name:     recSpec.Inputs[0].Name,
				Source:   &types.RecordInput_Hash{Hash: "8bf572e7226b661b87d47458b5aec2a7"}, // printf 'myinput' | md5sum
				TypeName: recSpec.Inputs[0].TypeName,
				Status:   types.RecordInputStatus_Proposed,
			},
		},
		Outputs: []types.RecordOutput{
			{
				Hash:   "240bd2d0c49dc033e5de33679b434108", // printf 'myoutput2' | md5sum
				Status: types.ResultStatus_RESULT_STATUS_PASS,
			},
		},
		SpecificationId: recSpecID,
	}

	s.app.MetadataKeeper.SetContractSpecification(s.ctx, cSpec)
	s.app.MetadataKeeper.SetRecordSpecification(s.ctx, recSpec)
	s.app.MetadataKeeper.SetScopeSpecification(s.ctx, scopeSpec)

	s.app.MetadataKeeper.SetScope(s.ctx, scope1)
	s.app.MetadataKeeper.SetSession(s.ctx, session1)
	s.app.MetadataKeeper.SetRecord(s.ctx, record1)
	s.app.MetadataKeeper.SetScope(s.ctx, scope2)
	s.app.MetadataKeeper.SetSession(s.ctx, session2)
	s.app.MetadataKeeper.SetRecord(s.ctx, record2)

	req := func(addrs ...string) *types.GetByAddrRequest {
		return &types.GetByAddrRequest{Addrs: addrs}
	}

	tests := []struct {
		name    string
		req     *types.GetByAddrRequest
		expResp *types.GetByAddrResponse
		expErr  string
	}{
		{
			name:   "nil req",
			req:    nil,
			expErr: "empty request: invalid request",
		},
		{
			name:   "no addrs",
			req:    req(),
			expErr: "empty request: invalid request",
		},
		{
			name:    "many addresses",
			req:     req(manyAddrs...),
			expResp: &types.GetByAddrResponse{NotFound: manyAddrs},
		},
		{
			name:    "one addr bad",
			req:     req("notanaddr"),
			expResp: &types.GetByAddrResponse{NotFound: []string{"notanaddr"}},
		},
		{
			name:    "one addr scope found",
			req:     req(scopeID1.String()),
			expResp: &types.GetByAddrResponse{Scopes: []*types.Scope{&scope1}},
		},
		{
			name:    "one addr scope not found",
			req:     req(scopeIDDNE.String()),
			expResp: &types.GetByAddrResponse{NotFound: []string{scopeIDDNE.String()}},
		},
		{
			name:    "one addr session found",
			req:     req(sessionID1.String()),
			expResp: &types.GetByAddrResponse{Sessions: []*types.Session{&session1}},
		},
		{
			name:    "one addr session not found",
			req:     req(sessionIDDNE.String()),
			expResp: &types.GetByAddrResponse{NotFound: []string{sessionIDDNE.String()}},
		},
		{
			name:    "one addr record found",
			req:     req(recordID1.String()),
			expResp: &types.GetByAddrResponse{Records: []*types.Record{&record1}},
		},
		{
			name:    "one addr record not found",
			req:     req(recordIDDNE.String()),
			expResp: &types.GetByAddrResponse{NotFound: []string{recordIDDNE.String()}},
		},
		{
			name:    "one addr scope spec found",
			req:     req(scopeSpecID.String()),
			expResp: &types.GetByAddrResponse{ScopeSpecs: []*types.ScopeSpecification{&scopeSpec}},
		},
		{
			name:    "one addr scope spec not found",
			req:     req(scopeSpecIDDNE.String()),
			expResp: &types.GetByAddrResponse{NotFound: []string{scopeSpecIDDNE.String()}},
		},
		{
			name:    "one addr contract spec found",
			req:     req(cSpecID.String()),
			expResp: &types.GetByAddrResponse{ContractSpecs: []*types.ContractSpecification{&cSpec}},
		},
		{
			name:    "one addr contract spec not found",
			req:     req(cSpecIDDNE.String()),
			expResp: &types.GetByAddrResponse{NotFound: []string{cSpecIDDNE.String()}},
		},
		{
			name:    "one addr record spec found",
			req:     req(recSpecID.String()),
			expResp: &types.GetByAddrResponse{RecordSpecs: []*types.RecordSpecification{&recSpec}},
		},
		{
			name:    "one addr record spec not found",
			req:     req(recSpecIDDNE.String()),
			expResp: &types.GetByAddrResponse{NotFound: []string{recSpecIDDNE.String()}},
		},
		{
			name: "two scopes found one not",
			req:  req(scopeIDDNE.String(), scopeID2.String(), scopeID1.String()),
			expResp: &types.GetByAddrResponse{
				Scopes:   []*types.Scope{&scope2, &scope1},
				NotFound: []string{scopeIDDNE.String()},
			},
		},
		{
			name: "two sessions found one not",
			req:  req(sessionIDDNE.String(), sessionID2.String(), sessionID1.String()),
			expResp: &types.GetByAddrResponse{
				Sessions: []*types.Session{&session2, &session1},
				NotFound: []string{sessionIDDNE.String()},
			},
		},
		{
			name: "two records found one not",
			req:  req(recordIDDNE.String(), recordID2.String(), recordID1.String()),
			expResp: &types.GetByAddrResponse{
				Records:  []*types.Record{&record2, &record1},
				NotFound: []string{recordIDDNE.String()},
			},
		},
		{
			name: "scope session and record",
			req:  req(scopeID1.String(), sessionID1.String(), recordID1.String()),
			expResp: &types.GetByAddrResponse{
				Scopes:   []*types.Scope{&scope1},
				Sessions: []*types.Session{&session1},
				Records:  []*types.Record{&record1},
			},
		},
		{
			name: "scope spec contract spec and record spec",
			req:  req(scopeSpecID.String(), cSpecID.String(), recSpecID.String()),
			expResp: &types.GetByAddrResponse{
				ScopeSpecs:    []*types.ScopeSpecification{&scopeSpec},
				ContractSpecs: []*types.ContractSpecification{&cSpec},
				RecordSpecs:   []*types.RecordSpecification{&recSpec},
			},
		},
		{
			name: "everything and then some",
			req: req(
				cSpecID.String(), scopeID1.String(), recordID1.String(),
				cSpecIDDNE.String(), recordIDDNE.String(), "badaddr",
				recSpecID.String(), scopeSpecID.String(), sessionID2.String(),
				recordID2.String(), sessionID1.String(), "addrnotfound",
				sessionIDDNE.String(), "anothernotfound", recSpecIDDNE.String(),
				scopeID2.String(), scopeSpecIDDNE.String(), scopeIDDNE.String(),
			),
			expResp: &types.GetByAddrResponse{
				Scopes:        []*types.Scope{&scope1, &scope2},
				Sessions:      []*types.Session{&session2, &session1},
				Records:       []*types.Record{&record1, &record2},
				ScopeSpecs:    []*types.ScopeSpecification{&scopeSpec},
				ContractSpecs: []*types.ContractSpecification{&cSpec},
				RecordSpecs:   []*types.RecordSpecification{&recSpec},
				NotFound: []string{
					cSpecIDDNE.String(), recordIDDNE.String(), "badaddr",
					"addrnotfound", sessionIDDNE.String(), "anothernotfound",
					recSpecIDDNE.String(), scopeSpecIDDNE.String(), scopeIDDNE.String(),
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			resp, err := s.queryClient.GetByAddr(s.ctx, tc.req)
			if len(tc.expErr) > 0 {
				s.Assert().EqualError(err, tc.expErr, "GetByAddr error")
			} else {
				s.Assert().NoError(err, "GetByAddr error")
			}
			if tc.expResp == nil {
				s.Assert().Nil(resp, "GetByAddr response")
			} else {
				s.Assert().Equal(tc.expResp.Scopes, resp.Scopes, "GetByAddr response Scopes")
				s.Assert().Equal(tc.expResp.Sessions, resp.Sessions, "GetByAddr response Sessions")
				s.Assert().Equal(tc.expResp.Records, resp.Records, "GetByAddr response Records")
				s.Assert().Equal(tc.expResp.ScopeSpecs, resp.ScopeSpecs, "GetByAddr response ScopeSpecs")
				s.Assert().Equal(tc.expResp.ContractSpecs, resp.ContractSpecs, "GetByAddr response ContractSpecs")
				s.Assert().Equal(tc.expResp.RecordSpecs, resp.RecordSpecs, "GetByAddr response RecordSpecs")
				s.Assert().Equal(tc.expResp.NotFound, resp.NotFound, "GetByAddr response NotFound")
			}
		})
	}
}

func (s *QueryServerTestSuite) TestScopeNetAssetValuesQuery() {
	app, ctx, queryClient := s.app, s.ctx, s.queryClient
	scopeID := types.ScopeMetadataAddress(uuid.New())
	scopeIDNF := types.ScopeMetadataAddress(uuid.New())

	netAssetValues := make([]types.NetAssetValue, 5)
	for i := range netAssetValues {
		netAssetValues[i] = types.NetAssetValue{
			Price: sdk.Coin{
				Denom:  fmt.Sprintf("usd%v", i),
				Amount: sdkmath.NewInt(100 * int64(i+1)),
			},
		}
		err := app.MetadataKeeper.SetNetAssetValue(ctx, scopeID, netAssetValues[i], "source")
		s.Require().NoError(err)
	}

	tests := []struct {
		name       string
		req        *types.QueryScopeNetAssetValuesRequest
		expErr     string
		expNavsLen int
	}{
		{
			name:       "Valid Request with results",
			req:        &types.QueryScopeNetAssetValuesRequest{Id: scopeID.String()},
			expErr:     "",
			expNavsLen: len(netAssetValues),
		},
		{
			name:       "Valid Request without results",
			req:        &types.QueryScopeNetAssetValuesRequest{Id: scopeIDNF.String()},
			expErr:     "",
			expNavsLen: 0,
		},
		{
			name:   "Invalid Request - Bad Scope ID",
			req:    &types.QueryScopeNetAssetValuesRequest{Id: "note-scope-id"},
			expErr: "error extracting scope address",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			resp, err := queryClient.ScopeNetAssetValues(gocontext.Background(), tc.req)
			if tc.expErr != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErr)
			} else {
				s.Require().NoError(err)
				s.Require().Len(resp.NetAssetValues, tc.expNavsLen)
			}
		})
	}
}

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
