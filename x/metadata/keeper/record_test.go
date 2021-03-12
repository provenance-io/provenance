package keeper_test

import (
	"fmt"
	"github.com/google/uuid"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/provenance/x/metadata/types"
)

type RecordKeeperTestSuite struct {
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

	scopeSpecUUID uuid.UUID
	scopeSpecID   types.MetadataAddress

	recordName string
	recordId   types.MetadataAddress

	sessionUUID uuid.UUID
	sessionId   types.MetadataAddress

	contractSpecUUID uuid.UUID
	contractSpecId   types.MetadataAddress
}

func (s *RecordKeeperTestSuite) SetupTest() {
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

	s.scopeUUID = uuid.New()
	s.scopeID = types.ScopeMetadataAddress(s.scopeUUID)

	s.scopeSpecUUID = uuid.New()
	s.scopeSpecID = types.ScopeSpecMetadataAddress(s.scopeSpecUUID)

	s.recordName = "TestRecord"
	s.recordId = types.RecordMetadataAddress(s.scopeUUID, s.recordName)

	s.sessionUUID = uuid.New()
	s.sessionId = types.SessionMetadataAddress(s.scopeUUID, s.sessionUUID)

	s.sessionUUID = uuid.New()
	s.sessionId = types.SessionMetadataAddress(s.scopeUUID, s.sessionUUID)

	s.contractSpecUUID = uuid.New()
	s.contractSpecId = types.ContractSpecMetadataAddress(s.contractSpecUUID)
}

func TestRecordKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(RecordKeeperTestSuite))
}

// func ownerPartyList defined in keeper_test.go

func (s *RecordKeeperTestSuite) TestMetadataRecordGetSetRemove() {

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

func (s *RecordKeeperTestSuite) TestMetadataRecordIterator() {
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

func (s *RecordKeeperTestSuite) TestValidateRecordRemove() {
	scope := types.NewScope(s.scopeID, s.scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, s.user1)
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
			errorMsg: "cannot get scope uuid: this metadata address () does not contain a scope uuid",
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
		"Invalid, missing signature": {
			existing: *record,
			proposed: recordID,
			signers:  []string{"no-matchin"},
			wantErr:  true,
			errorMsg: fmt.Sprintf("missing signature from %s (PARTY_TYPE_OWNER)", s.user1),
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

func (s *RecordKeeperTestSuite) TestValidateRecordUpdate() {
	scope := types.NewScope(s.scopeID, s.scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, s.user1)
	s.app.MetadataKeeper.SetScope(s.ctx, *scope)

	process := types.NewProcess("processname", &types.Process_Hash{Hash: "HASH"}, "process_method")
	record := types.NewRecord(s.recordName, s.sessionId, *process, []types.RecordInput{}, []types.RecordOutput{})

	randomScopeUUID := uuid.New()
	randomSessionId := types.SessionMetadataAddress(randomScopeUUID, uuid.New())

	cases := map[string]struct {
		existing *types.Record
		proposed types.Record
		signers  []string
		wantErr  bool
		errorMsg string
	}{
		"nil previous, proposed throws address error": {
			existing: nil,
			proposed: types.Record{},
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: "incorrect address length (must be at least 17, actual: 0)",
		},
		"invalid names, existing and proposed names do not match": {
			existing: types.NewRecord("notamatch", s.sessionId, *process, []types.RecordInput{}, []types.RecordOutput{}),
			proposed: *types.NewRecord("not-a-match", s.sessionId, *process, []types.RecordInput{}, []types.RecordOutput{}),
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: "the Name field of records cannot be changed",
		},
		"invalid ids, existing and proposed ids do not match": {
			existing: types.NewRecord(s.recordName, s.sessionId, *process, []types.RecordInput{}, []types.RecordOutput{}),
			proposed: *types.NewRecord(s.recordName, randomSessionId, *process, []types.RecordInput{}, []types.RecordOutput{}),
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: "the SessionId field of records cannot be changed",
		},
		"both valid records, scope not found": {
			existing: types.NewRecord(s.recordName, randomSessionId, *process, []types.RecordInput{}, []types.RecordOutput{}),
			proposed: *types.NewRecord(s.recordName, randomSessionId, *process, []types.RecordInput{}, []types.RecordOutput{}),
			signers:  []string{s.user1},
			wantErr:  true,
			errorMsg: fmt.Sprintf("scope not found for scope uuid %s", randomScopeUUID),
		},
		"missing signature from existing owner ": {
			existing: record,
			proposed: *record,
			signers:  []string{},
			wantErr:  true,
			errorMsg: fmt.Sprintf("missing signature from %s (PARTY_TYPE_OWNER)", s.user1),
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
