package keeper_test

import (
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
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
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

	record     types.Record
	recordName string
	recordId   types.MetadataAddress

	sessionUUID uuid.UUID
	sessionId   types.MetadataAddress

	contractSpecUUID uuid.UUID
	contractSpecId   types.MetadataAddress
}

func ownerPartyList(addresses ...string) []types.Party {
	retval := make([]types.Party, len(addresses))
	for i, addr := range addresses {
		retval[i] = types.Party{Address: addr, Role: types.PartyType_PARTY_TYPE_OWNER}
	}
	return retval
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

	s.recordName = "TestRecord"
	s.recordId = types.RecordMetadataAddress(s.scopeUUID, s.recordName)

	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))

	s.sessionUUID = uuid.New()
	s.sessionId = types.SessionMetadataAddress(s.scopeUUID, s.sessionUUID)

	s.sessionUUID = uuid.New()
	s.sessionId = types.SessionMetadataAddress(s.scopeUUID, s.sessionUUID)

	s.contractSpecUUID = uuid.New()
	s.contractSpecId = types.ContractSpecMetadataAddress(s.contractSpecUUID)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.MetadataKeeper)
	queryClient := types.NewQueryClient(queryHelper)
	s.queryClient = queryClient
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) TestMetadataScopeGetSet() {
	scope, found := s.app.MetadataKeeper.GetScope(s.ctx, s.scopeID)
	s.NotNil(scope)
	s.False(found)

	ns := *types.NewScope(s.scopeID, s.specID, ownerPartyList(s.user1), []string{s.user1}, s.user1)
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

func (s *KeeperTestSuite) TestMetadataScopeIterator() {
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

func (s *KeeperTestSuite) TestMetadataRecordGetSetRemove() {

	r, found := s.app.MetadataKeeper.GetRecord(s.ctx, s.recordId)
	s.NotNil(r)
	s.False(found)

	process := types.NewProcess("processname", &types.Process_Hash{Hash: "HASH"}, "process_method")
	record := types.NewRecord(s.recordName, s.sessionId, *process, []types.RecordInput{}, []types.RecordOutput{})

	s.NotNil(record)
	s.app.MetadataKeeper.SetRecord(s.ctx, *record)

	r, found = s.app.MetadataKeeper.GetRecord(s.ctx, s.recordId)
	s.True(found)
	s.NotNil(r)

	s.app.MetadataKeeper.RemoveRecord(s.ctx, s.recordId)
	r, found = s.app.MetadataKeeper.GetRecord(s.ctx, s.recordId)
	s.False(found)
	s.NotNil(r)
}

func (s *KeeperTestSuite) TestMetadataRecordIterator() {
	for i := 1; i <= 10; i++ {
		process := types.NewProcess("processname", &types.Process_Hash{Hash: "HASH"}, "process_method")
		recordName := fmt.Sprintf("%s%v", s.recordName, i)
		record := types.NewRecord(recordName, s.sessionId, *process, []types.RecordInput{}, []types.RecordOutput{})
		s.app.MetadataKeeper.SetRecord(s.ctx, *record)
	}
	count := 0
	s.app.MetadataKeeper.IterateRecords(s.ctx, s.scopeID, func(s types.Record) (stop bool) {
		count++
		return false
	})
	s.Equal(10, count, "iterator should return a full list of records")

}

func (s *KeeperTestSuite) TestValidateRecordRemove() {
	scope := types.NewScope(s.scopeID, s.specID, ownerPartyList(s.user1), []string{s.user1}, s.user1)
	s.app.MetadataKeeper.SetScope(s.ctx, *scope)

	process := types.NewProcess("processname", &types.Process_Hash{Hash: "HASH"}, "process_method")
	record := types.NewRecord(s.recordName, s.sessionId, *process, []types.RecordInput{}, []types.RecordOutput{})
	recordID := types.RecordMetadataAddress(s.scopeUUID, s.recordName)
	s.app.MetadataKeeper.SetRecord(s.ctx, *record)

	dneRecordID := types.RecordMetadataAddress(s.scopeUUID, "does-not-exist")

	cases := map[string]struct {
		existing types.Record
		proposed types.MetadataAddress
		signers  []string
		wantErr  bool
		errorMsg string
	}{
		"invalid, existing record doesn't have scope": {
			existing: types.Record{},
			proposed: types.MetadataAddress{},
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: "cannot get scope uuid: this metadata address does not contain a scope uuid",
		},
		"invalid, unable to find scope": {
			existing: *record,
			proposed: types.MetadataAddress{},
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: fmt.Sprintf("cannot remove record. expected %s, got ", recordID),
		},
		"invalid, record ids do not match": {
			existing: *record,
			proposed: dneRecordID,
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: fmt.Sprintf("cannot remove record. expected %s, got %s", recordID, dneRecordID),
		},
		"Invalid, missing signature record ids do not match": {
			existing: *record,
			proposed: recordID,
			signers:  []string{"no-matchin"},
			wantErr:  true,
			errorMsg: fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1),
		},
		"valid, passed all validation": {
			existing: *record,
			proposed: recordID,
			signers:  []string{s.user1},
			wantErr:  false,
			errorMsg: "",
		},
	}

	for n, tc := range cases {
		tc := tc

		s.Run(n, func() {
			err := s.app.MetadataKeeper.ValidateRecordRemove(s.ctx, tc.existing, tc.proposed, tc.signers)
			if tc.wantErr {
				s.Error(err)
				s.Equal(tc.errorMsg, err.Error())
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestValidateRecordUpdate() {
	scope := types.NewScope(s.scopeID, s.specID, ownerPartyList(s.user1), []string{s.user1}, s.user1)
	s.app.MetadataKeeper.SetScope(s.ctx, *scope)

	process := types.NewProcess("processname", &types.Process_Hash{Hash: "HASH"}, "process_method")
	record := types.NewRecord(s.recordName, s.sessionId, *process, []types.RecordInput{}, []types.RecordOutput{})

	randomScopeUUID := uuid.New()
	randomSessionId := types.SessionMetadataAddress(randomScopeUUID, uuid.New())

	cases := map[string]struct {
		existing types.Record
		proposed types.Record
		signers  []string
		wantErr  bool
		errorMsg string
	}{
		"nil previous, proposed throws address error": {
			existing: types.Record{},
			proposed: types.Record{},
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: "incorrect address length (must be at least 17, actual: 0)",
		},
		"invalid names, existing and proposed names do not match": {
			existing: *types.NewRecord("notamatch", s.sessionId, *process, []types.RecordInput{}, []types.RecordOutput{}),
			proposed: *types.NewRecord("not-a-match", s.sessionId, *process, []types.RecordInput{}, []types.RecordOutput{}),
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: "existing and proposed records do not match",
		},
		"invalid ids, existing and proposed ids do not match": {
			existing: *types.NewRecord(s.recordName, s.sessionId, *process, []types.RecordInput{}, []types.RecordOutput{}),
			proposed: *types.NewRecord(s.recordName, randomSessionId, *process, []types.RecordInput{}, []types.RecordOutput{}),
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: "existing and proposed records do not match",
		},
		"both valid records, scope not found": {
			existing: *types.NewRecord(s.recordName, randomSessionId, *process, []types.RecordInput{}, []types.RecordOutput{}),
			proposed: *types.NewRecord(s.recordName, randomSessionId, *process, []types.RecordInput{}, []types.RecordOutput{}),
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: fmt.Sprintf("scope not found for scope uuid %s", randomScopeUUID),
		},
		"missing signature from existing owner ": {
			existing: *record,
			proposed: *record,
			signers:  []string{},
			wantErr:  true,
			errorMsg: fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1),
		},
	}

	for n, tc := range cases {
		tc := tc

		s.Run(n, func() {
			err := s.app.MetadataKeeper.ValidateRecordUpdate(s.ctx, tc.existing, tc.proposed, tc.signers)
			if tc.wantErr {
				s.Error(err)
				s.Equal(tc.errorMsg, err.Error())
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMetadataSessionGetSetRemove() {

	r, found := s.app.MetadataKeeper.GetSession(s.ctx, s.sessionId)
	s.Empty(r)
	s.False(found)

	session := types.NewSession("name", s.sessionId, s.contractSpecId, []types.Party{
		{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: types.PartyType_PARTY_TYPE_AFFILIATE}},
		&types.AuditFields{CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", CreatedDate: time.Now(),
			UpdatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", UpdatedDate: time.Now(),
			Message: "message",
		})

	s.NotNil(session)
	s.app.MetadataKeeper.SetSession(s.ctx, *session)

	sess, found := s.app.MetadataKeeper.GetSession(s.ctx, s.sessionId)
	s.True(found)
	s.NotEmpty(sess)

	s.app.MetadataKeeper.RemoveSession(s.ctx, s.sessionId)
	sess, found = s.app.MetadataKeeper.GetSession(s.ctx, s.sessionId)
	s.False(found)
	s.Empty(sess)

	process := types.NewProcess("processname", &types.Process_Hash{Hash: "HASH"}, "process_method")
	record := types.NewRecord(s.recordName, s.sessionId, *process, []types.RecordInput{}, []types.RecordOutput{})
	s.app.MetadataKeeper.SetRecord(s.ctx, *record)
	s.app.MetadataKeeper.SetSession(s.ctx, *session)

	sess, found = s.app.MetadataKeeper.GetSession(s.ctx, s.sessionId)
	s.True(found)
	s.NotEmpty(sess)

	s.app.MetadataKeeper.RemoveSession(s.ctx, s.sessionId)
	sess, found = s.app.MetadataKeeper.GetSession(s.ctx, s.sessionId)
	s.True(found)
	s.NotEmpty(sess)

}

func (s *KeeperTestSuite) TestMetadataSessionIterator() {
	for i := 1; i <= 10; i++ {
		sessionId := types.SessionMetadataAddress(s.scopeUUID, uuid.New())
		session := types.NewSession("name", sessionId, s.contractSpecId, []types.Party{
			{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: types.PartyType_PARTY_TYPE_AFFILIATE}},
			&types.AuditFields{CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", CreatedDate: time.Now(),
				UpdatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", UpdatedDate: time.Now(),
				Message: "message",
			})
		s.app.MetadataKeeper.SetSession(s.ctx, *session)
	}
	count := 0
	s.app.MetadataKeeper.IterateSessions(s.ctx, s.scopeID, func(s types.Session) (stop bool) {
		count++
		return false
	})
	s.Equal(10, count, "iterator should return a full list of sessions")

}

func (s *KeeperTestSuite) TestMetadataAuditUpdate() {
	auditTime := time.Now()
	var initial *types.AuditFields
	result := initial.UpdateAudit(s.ctx, "creator", "initial")
	s.Equal(uint32(1), result.Version)
	s.Equal(s.ctx.BlockTime(), result.CreatedDate)
	s.Equal("creator", result.CreatedBy)
	s.Equal(time.Time{}, result.UpdatedDate)
	s.Equal("", result.UpdatedBy)
	s.Equal("initial", result.Message)

	initial = &types.AuditFields{CreatedDate: auditTime, CreatedBy: "creator", Version: 1}
	result = initial.UpdateAudit(s.ctx, "updater", "")
	s.Equal(uint32(2), result.Version)
	s.Equal(auditTime, result.CreatedDate)
	s.Equal("creator", result.CreatedBy)
	s.Equal(s.ctx.BlockTime(), result.UpdatedDate)
	s.Equal("updater", result.UpdatedBy)
	s.Equal("", result.Message)
}

func (s *KeeperTestSuite) TestMetadataValidateSessionUpdate() {
	scope := types.NewScope(s.scopeID, s.specID, ownerPartyList(s.user1), []string{s.user1}, s.user1)
	s.app.MetadataKeeper.SetScope(s.ctx, *scope)

	auditTime := time.Now()

	invalidScopeUUID := uuid.New()
	parties := []types.Party{{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: types.PartyType_PARTY_TYPE_AFFILIATE}}
	validSession := types.NewSession("processname", s.sessionId, s.contractSpecId, parties, nil)
	validSessionWithAudit := types.NewSession("processname", s.sessionId, s.contractSpecId, parties, &types.AuditFields{CreatedDate: auditTime, CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Version: 1})
	invalidIdSession := types.NewSession("processname", types.SessionMetadataAddress(invalidScopeUUID, uuid.New()), s.contractSpecId, parties, nil)
	invalidContractId := types.NewSession("processname", s.sessionId, types.ContractSpecMetadataAddress(uuid.New()), parties, nil)
	invalidParties := []types.Party{{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: types.PartyType_PARTY_TYPE_CUSTODIAN}}
	invalidPartiesSession := types.NewSession("processname", s.sessionId, s.contractSpecId, invalidParties, nil)
	invalidNameSession := types.NewSession("invalid", s.sessionId, s.contractSpecId, parties, nil)

	partiesInvolved := []types.PartyType{types.PartyType_PARTY_TYPE_AFFILIATE}
	contractSpec := types.NewContractSpecification(s.contractSpecId, types.NewDescription("name", "desc", "url", "icon"), []string{s.user1}, partiesInvolved, &types.ContractSpecification_Hash{"hash"}, "processname")
	s.app.MetadataKeeper.SetContractSpecification(s.ctx, *contractSpec)

	cases := map[string]struct {
		existing types.Session
		proposed types.Session
		signers  []string
		wantErr  bool
		errorMsg string
	}{
		"invalid session update, existing record does not have scope": {
			existing: types.Session{},
			proposed: types.Session{},
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: "incorrect address length (must be at least 17, actual: 0)",
		},
		"valid session update, existing and proposed satisfy validation": {
			existing: *validSession,
			proposed: *validSession,
			signers:  []string{s.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"valid session update without audit, existing has audit": {
			existing: *validSessionWithAudit,
			proposed: *validSession,
			signers:  []string{s.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"invalid session update has audit, existing has no audit": {
			existing: *validSession,
			proposed: *validSessionWithAudit,
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: "attempt to modify audit fields, modification not allowed",
		},
		"valid session update existing and new have matching audits": {
			existing: *validSessionWithAudit,
			proposed: *validSessionWithAudit,
			signers:  []string{s.user1},
			wantErr:  false,
			errorMsg: "",
		},
		"invalid session update, existing id does not match proposed": {
			existing: *validSession,
			proposed: *invalidIdSession,
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: fmt.Sprintf("cannot update session identifier. expected %s, got %s", validSession.SessionId, invalidIdSession.SessionId),
		},
		"invalid session update, scope does not exist": {
			existing: *invalidIdSession,
			proposed: *invalidIdSession,
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: fmt.Sprintf("scope not found for scope id %s", types.ScopeMetadataAddress(invalidScopeUUID)),
		},
		"invalid session update, contract spec does not exist": {
			existing: *validSession,
			proposed: *invalidContractId,
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: fmt.Sprintf("cannot find contract specification %s", invalidContractId.SpecificationId),
		},
		"invalid session update, involved parties do not match": {
			existing: *validSession,
			proposed: *invalidPartiesSession,
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: "missing party type from required parties PARTY_TYPE_AFFILIATE",
		},
		"invalid session update, missing required signers": {
			existing: *validSession,
			proposed: *validSession,
			signers:  []string{"unknown signer"},
			wantErr:  true,
			errorMsg: fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1),
		},
		"invalid session update, proposed name does not match contract spec": {
			existing: *validSession,
			proposed: *invalidNameSession,
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: "proposed name does not match contract spec. expected invalid, got processname)",
		},
		"invalid session update, modified audit message": {
			existing: *validSessionWithAudit,
			proposed: *types.NewSession("processname", s.sessionId, s.contractSpecId, parties, &types.AuditFields{CreatedDate: auditTime, CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Version: 1, Message: "fault"}),
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: "attempt to modify message audit field, modification not allowed",
		},
		"invalid session update, modified audit version": {
			existing: *validSessionWithAudit,
			proposed: *types.NewSession("processname", s.sessionId, s.contractSpecId, parties, &types.AuditFields{CreatedDate: auditTime, CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Version: 2}),
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: "attempt to modify version audit field, modification not allowed",
		},
		"invalid session update, modified audit update date": {
			existing: *validSessionWithAudit,
			proposed: *types.NewSession("processname", s.sessionId, s.contractSpecId, parties, &types.AuditFields{CreatedDate: auditTime, CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Version: 1, UpdatedDate: time.Now()}),
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: "attempt to modify updated-date audit field, modification not allowed",
		},
		"invalid session update, modified audit update by": {
			existing: *validSessionWithAudit,
			proposed: *types.NewSession("processname", s.sessionId, s.contractSpecId, parties, &types.AuditFields{CreatedDate: auditTime, CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Version: 1, UpdatedBy: "fault"}),
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: "attempt to modify updated-by audit field, modification not allowed",
		},
		"invalid session update, modified audit created by": {
			existing: *validSessionWithAudit,
			proposed: *types.NewSession("processname", s.sessionId, s.contractSpecId, parties, &types.AuditFields{CreatedDate: auditTime, CreatedBy: "fault", Version: 1, UpdatedBy: "fault"}),
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: "attempt to modify created-by audit field, modification not allowed",
		},
		"invalid session update, modified audit created date": {
			existing: *validSessionWithAudit,
			proposed: *types.NewSession("processname", s.sessionId, s.contractSpecId, parties, &types.AuditFields{CreatedDate: time.Now(), CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Version: 1}),
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: "attempt to modify created-date audit field, modification not allowed",
		},
	}

	for n, tc := range cases {
		tc := tc

		s.Run(n, func() {
			err := s.app.MetadataKeeper.ValidateSessionUpdate(s.ctx, tc.existing, tc.proposed, tc.signers)
			if tc.wantErr {
				s.Error(err)
				s.Equal(tc.errorMsg, err.Error())
			} else {
				s.NoError(err)
			}
		})
	}
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
			errorMsg:        "missing party type from required parties PARTY_TYPE_AFFILIATE",
		},
		"invalid, missing required parties": {
			parties:         []types.Party{{Address: "address", Role: types.PartyType_PARTY_TYPE_CUSTODIAN}},
			requiredParties: []types.PartyType{types.PartyType_PARTY_TYPE_AFFILIATE},
			wantErr:         true,
			errorMsg:        "missing party type from required parties PARTY_TYPE_AFFILIATE",
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

		s.Run(n, func() {
			err := s.app.MetadataKeeper.ValidatePartiesInvolved(tc.parties, tc.requiredParties)
			if tc.wantErr {
				s.Error(err)
				s.Equal(tc.errorMsg, err.Error())
			} else {
				s.NoError(err)
			}
		})
	}

}

func (s *KeeperTestSuite) TestValidateRequiredSignatures() {

	cases := map[string]struct {
		owners   []types.Party
		signers  []string
		wantErr  bool
		errorMsg string
	}{
		"valid, no signatures": {
			owners:   []types.Party{},
			signers:  []string{},
			wantErr:  false,
			errorMsg: "",
		},
		"valid, signatures match": {
			owners:   []types.Party{{Address: "signer1"}},
			signers:  []string{"signer1"},
			wantErr:  false,
			errorMsg: "",
		},
		"valid, all owner signatures satisfied": {
			owners:   []types.Party{{Address: "signer1"}},
			signers:  []string{"signer1", "signer2"},
			wantErr:  false,
			errorMsg: "",
		},
		"invalid, missing owner signature": {
			owners:   []types.Party{{Address: "missingowner"}},
			signers:  []string{"signer1", "signer2"},
			wantErr:  true,
			errorMsg: "missing signature from existing owner missingowner; required for update",
		},
	}

	for n, tc := range cases {
		tc := tc

		s.Run(n, func() {
			err := s.app.MetadataKeeper.ValidateRequiredSignatures(tc.owners, tc.signers)
			if tc.wantErr {
				s.Error(err)
				s.Equal(tc.errorMsg, err.Error())
			} else {
				s.NoError(err)
			}
		})
	}

}

func (s *KeeperTestSuite) TestValidateAllOwnersAreSigners() {
	tests := []struct {
		name    string
		owners  []string
		signers []string
		want    string
	}{
		{
			"Scope Spec with 1 owner: no signers - error",
			[]string{s.user1Addr.String()},
			[]string{},
			fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1Addr.String()),
		},
		{
			"Scope Spec with 1 owner: not in signers list - error",
			[]string{s.user1Addr.String()},
			[]string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1Addr.String()),
		},
		{
			"Scope Spec with 1 owner: in signers list with non-owners - ok",
			[]string{s.user1Addr.String()},
			[]string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user1Addr.String(), sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			"",
		},
		{
			"Scope Spec with 1 owner: only signer in list - ok",
			[]string{s.user1Addr.String()},
			[]string{s.user1Addr.String()},
			"",
		},
		{
			"Scope Spec with 2 owners: no signers - error",
			[]string{s.user1Addr.String(), s.user2Addr.String()},
			[]string{},
			fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1Addr.String()),
		},
		{
			"Scope Spec with 2 owners: neither in signers list - error",
			[]string{s.user1Addr.String(), s.user2Addr.String()},
			[]string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1Addr.String()),
		},
		{
			"Scope Spec with 2 owners: one in signers list with non-owners - error",
			[]string{s.user1Addr.String(), s.user2Addr.String()},
			[]string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user1Addr.String(), sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			fmt.Sprintf("missing signature from existing owner %s; required for update", s.user2Addr.String()),
		},
		{
			"Scope Spec with 2 owners: the other in signers list with non-owners - error",
			[]string{s.user1Addr.String(), s.user2Addr.String()},
			[]string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user2Addr.String(), sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()},
			fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1Addr.String()),
		},
		{
			"Scope Spec with 2 owners: both in signers list with non-owners - ok",
			[]string{s.user1Addr.String(), s.user2Addr.String()},
			[]string{sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user2Addr.String(), sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(), s.user1Addr.String()},
			"",
		},
		{
			"Scope Spec with 2 owners: only both in signers list - ok",
			[]string{s.user1Addr.String(), s.user2Addr.String()},
			[]string{s.user1Addr.String(), s.user2Addr.String()},
			"",
		},
		{
			"Scope Spec with 2 owners: only both in signers list, opposite order - ok",
			[]string{s.user1Addr.String(), s.user2Addr.String()},
			[]string{s.user2Addr.String(), s.user1Addr.String()},
			"",
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := s.app.MetadataKeeper.ValidateAllOwnersAreSigners(tt.owners, tt.signers)
			if err != nil {
				require.Equal(t, tt.want, err.Error(), "ScopeSpec Keeper ValidateScopeSpecAllOwnersAreSigners error")
			} else if len(tt.want) > 0 {
				t.Errorf("ScopeSpec Keeper ValidateAllOwnersAreSigners error = nil, expected: %s", tt.want)
			}
		})
	}
}
