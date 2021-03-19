package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/metadata/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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
	recordID   types.MetadataAddress

	sessionUUID uuid.UUID
	sessionID   types.MetadataAddress
	sessionName string

	contractSpecUUID uuid.UUID
	contractSpecID   types.MetadataAddress

	recordSpecID types.MetadataAddress
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
	s.recordID = types.RecordMetadataAddress(s.scopeUUID, s.recordName)

	s.sessionUUID = uuid.New()
	s.sessionID = types.SessionMetadataAddress(s.scopeUUID, s.sessionUUID)
	s.sessionName = "TestSession"

	s.sessionUUID = uuid.New()
	s.sessionID = types.SessionMetadataAddress(s.scopeUUID, s.sessionUUID)

	s.contractSpecUUID = uuid.New()
	s.contractSpecID = types.ContractSpecMetadataAddress(s.contractSpecUUID)

	s.recordSpecID = types.RecordSpecMetadataAddress(s.contractSpecUUID, s.recordName)
}

func TestRecordKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(RecordKeeperTestSuite))
}

// func ownerPartyList defined in keeper_test.go

func (s *RecordKeeperTestSuite) TestMetadataRecordGetSetRemove() {

	r, found := s.app.MetadataKeeper.GetRecord(s.ctx, s.recordID)
	s.NotNil(r)
	s.False(found)

	process := types.NewProcess("processname", &types.Process_Hash{Hash: "HASH"}, "process_method")
	record := types.NewRecord(s.recordName, s.sessionID, *process, []types.RecordInput{}, []types.RecordOutput{})

	s.NotNil(record)
	s.app.MetadataKeeper.SetRecord(s.ctx, *record)

	r, found = s.app.MetadataKeeper.GetRecord(s.ctx, s.recordID)
	s.True(found)
	s.NotNil(r)

	s.app.MetadataKeeper.RemoveRecord(s.ctx, s.recordID)
	r, found = s.app.MetadataKeeper.GetRecord(s.ctx, s.recordID)
	s.False(found)
	s.NotNil(r)
}

func (s *RecordKeeperTestSuite) TestMetadataRecordIterator() {
	for i := 1; i <= 10; i++ {
		process := types.NewProcess("processname", &types.Process_Hash{Hash: "HASH"}, "process_method")
		recordName := fmt.Sprintf("%s%v", s.recordName, i)
		record := types.NewRecord(recordName, s.sessionID, *process, []types.RecordInput{}, []types.RecordOutput{})
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
	record := types.NewRecord(s.recordName, s.sessionID, *process, []types.RecordInput{}, []types.RecordOutput{})
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
	scopeID := types.ScopeMetadataAddress(uuid.New())
	scope := types.NewScope(scopeID, s.scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, s.user1)
	s.app.MetadataKeeper.SetScope(s.ctx, *scope)

	sessionID, err := scopeID.AsSessionAddress(uuid.New())
	s.Require().NoError(err, "error making sessionID")
	session := types.NewSession(
		s.sessionName,
		sessionID,
		s.contractSpecID,
		ownerPartyList(s.user1),
		&types.AuditFields{
			CreatedDate: time.Time{},
			CreatedBy:   s.user1,
			UpdatedDate: time.Time{},
			UpdatedBy:   "",
			Version:     0,
			Message:     "",
		},
	)
	s.app.MetadataKeeper.SetSession(s.ctx, *session)

	recordSpecID := types.RecordSpecMetadataAddress(s.contractSpecUUID, s.recordName)
	recordTypeName := "TestRecordTypeName"
	recordSpec := types.NewRecordSpecification(
		recordSpecID,
		s.recordName,
		[]*types.InputSpecification{},
		recordTypeName,
		types.DefinitionType_DEFINITION_TYPE_RECORD,
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
	)
	inputSpecSourceHash := "inputspecsourcehash"
	inputSpec := types.NewInputSpecification(
		"TestInput",
		"TestInputType",
		types.NewInputSpecificationSourceHash(inputSpecSourceHash),
	)
	recordSpec.Inputs = append(recordSpec.Inputs, inputSpec)
	s.app.MetadataKeeper.SetRecordSpecification(s.ctx, *recordSpec)

	recordName2 := s.recordName + "2"
	recordSpec2ID := types.RecordSpecMetadataAddress(s.contractSpecUUID, recordName2)
	recordSpec2 := types.NewRecordSpecification(
		recordSpec2ID,
		recordName2,
		[]*types.InputSpecification{},
		recordTypeName,
		types.DefinitionType_DEFINITION_TYPE_RECORD_LIST,
		[]types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
	)
	inputSpec2SourceHash := "inputspec2sourcehash"
	inputSpec2 := types.NewInputSpecification(
		"TestInput2",
		"TestInput2Type",
		types.NewInputSpecificationSourceHash(inputSpec2SourceHash),
	)
	recordSpec2.Inputs = append(recordSpec2.Inputs, inputSpec2)
	s.app.MetadataKeeper.SetRecordSpecification(s.ctx, *recordSpec2)

	process := types.NewProcess("processname", &types.Process_Hash{Hash: "HASH"}, "process_method")
	goodInput := types.NewRecordInput(
		inputSpec.Name,
		&types.RecordInput_Hash{Hash: inputSpecSourceHash},
		inputSpec.TypeName,
		types.RecordInputStatus_Proposed,
	)
	goodInput2 := types.NewRecordInput(
		inputSpec2.Name,
		&types.RecordInput_Hash{Hash: inputSpec2SourceHash},
		inputSpec2.TypeName,
		types.RecordInputStatus_Proposed,
	)
	otherInput := types.NewRecordInput(
		"otherInput",
		&types.RecordInput_Hash{Hash: inputSpecSourceHash},
		inputSpec.TypeName,
		types.RecordInputStatus_Proposed,
	)

	randomScopeID := types.ScopeMetadataAddress(uuid.New())
	randomSessionID, err := randomScopeID.AsSessionAddress(uuid.New())
	s.Require().NoError(err, "error making randomSessionID")
	randomInScopeSessionID, err := scope.ScopeId.AsSessionAddress(uuid.New())
	s.Require().NoError(err, "error making randomInScopeSessionID")
	missingRecordSpecName := "missing"
	missingRecordSpecID, err := s.contractSpecID.AsRecordSpecAddress(missingRecordSpecName)
	s.Require().NoError(err, "error making randomInScopeSessionID")

	cases := map[string]struct {
		existing *types.Record
		proposed types.Record
		signers  []string
		errorMsg string
	}{
		"validate basic called on proposed": {
			existing: nil,
			proposed: types.Record{},
			signers:  []string{s.user1},
			errorMsg: "address is empty",
		},
		"existing and proposed names do not match": {
			existing: types.NewRecord("notamatch", s.sessionID, *process, []types.RecordInput{}, []types.RecordOutput{}),
			proposed: *types.NewRecord("not-a-match", s.sessionID, *process, []types.RecordInput{}, []types.RecordOutput{}),
			signers:  []string{s.user1},
			errorMsg: "the Name field of records cannot be changed",
		},
		"existing and proposed ids do not match": {
			existing: types.NewRecord(s.recordName, s.sessionID, *process, []types.RecordInput{}, []types.RecordOutput{}),
			proposed: *types.NewRecord(s.recordName, randomSessionID, *process, []types.RecordInput{}, []types.RecordOutput{}),
			signers:  []string{s.user1},
			errorMsg: "the SessionId field of records cannot be changed",
		},
		"scope not found": {
			existing: nil,
			proposed: *types.NewRecord(s.recordName, randomSessionID, *process, []types.RecordInput{}, []types.RecordOutput{}),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("scope not found for scope id %s", randomScopeID),
		},
		"missing signature from existing owner": {
			existing: nil,
			proposed: *types.NewRecord(s.recordName, sessionID, *process, []types.RecordInput{}, []types.RecordOutput{}),
			signers:  []string{},
			errorMsg: fmt.Sprintf("missing signature from %s (PARTY_TYPE_OWNER)", s.user1),
		},
		"session not found": {
			existing: nil,
			proposed: *types.NewRecord(s.recordName, randomInScopeSessionID, *process, []types.RecordInput{}, []types.RecordOutput{}),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("session not found for session id %s", randomInScopeSessionID),
		},
		"record specification not found": {
			existing: nil,
			proposed: *types.NewRecord(missingRecordSpecName, sessionID, *process, []types.RecordInput{}, []types.RecordOutput{}),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("record specification not found for record specification id %s (contract spec uuid %s and record name %s)",
				missingRecordSpecID, s.contractSpecUUID, missingRecordSpecName),
		},
		"missing input": {
			existing: nil,
			proposed: *types.NewRecord(s.recordName, sessionID, *process, []types.RecordInput{*otherInput}, []types.RecordOutput{}),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("missing input %s", goodInput.Name),
		},
		"extra input": {
			existing: nil,
			proposed: *types.NewRecord(s.recordName, sessionID, *process, []types.RecordInput{*goodInput, *otherInput}, []types.RecordOutput{}),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("extra input %s", otherInput.Name),
		},
		"duplicate input": {
			existing: nil,
			proposed: *types.NewRecord(s.recordName, sessionID, *process, []types.RecordInput{*goodInput, *goodInput}, []types.RecordOutput{}),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("input name %s provided twice", goodInput.Name),
		},
		"input type name wrong": {
			existing: nil,
			proposed: *types.NewRecord(
				s.recordName, sessionID, *process,
				[]types.RecordInput{
					{
						Name:     goodInput.Name,
						Source:   goodInput.Source,
						TypeName: "bad type name",
						Status:   goodInput.Status,
					},
				},
				[]types.RecordOutput{}),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("input %s has TypeName %s but spec calls for %s",
				goodInput.Name, "bad type name", inputSpec.TypeName),
		},
		"input source type wrong": {
			existing: nil,
			proposed: *types.NewRecord(
				s.recordName, sessionID, *process,
				[]types.RecordInput{
					{
						Name:     goodInput.Name,
						Source:   &types.RecordInput_RecordId{RecordId: s.recordID},
						TypeName: goodInput.TypeName,
						Status:   types.RecordInputStatus_Record,
					},
				},
				[]types.RecordOutput{}),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("input %s has source type %s but spec calls for %s",
				goodInput.Name, "record", "hash"),
		},
		"input source value wrong": {
			existing: nil,
			proposed: *types.NewRecord(
				s.recordName, sessionID, *process,
				[]types.RecordInput{
					{
						Name:     goodInput.Name,
						Source:   &types.RecordInput_Hash{Hash: "incorrectsourcehash"},
						TypeName: goodInput.TypeName,
						Status:   goodInput.Status,
					},
				},
				[]types.RecordOutput{}),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("input %s has source value %s but spec calls for %s",
				goodInput.Name, "incorrectsourcehash", inputSpecSourceHash),
		},
		"output count wrong - record - zero": {
			existing: nil,
			proposed: *types.NewRecord(
				s.recordName, sessionID, *process, []types.RecordInput{*goodInput}, []types.RecordOutput{}),
			signers:  []string{s.user1},
			errorMsg: "invalid output count (expected: 1, got: 0)",
		},
		"output count wrong - record - two": {
			existing: nil,
			proposed: *types.NewRecord(
				s.recordName, sessionID, *process, []types.RecordInput{*goodInput},
				[]types.RecordOutput{
					{
						Hash:   "justsomeoutput",
						Status: types.ResultStatus_RESULT_STATUS_PASS,
					},
					{
						Hash:   "justsomeoutput2",
						Status: types.ResultStatus_RESULT_STATUS_PASS,
					},
				}),
			signers:  []string{s.user1},
			errorMsg: "invalid output count (expected: 1, got: 2)",
		},
		"output count wrong - record list - zero": {
			existing: nil,
			proposed: *types.NewRecord(recordName2, sessionID, *process, []types.RecordInput{*goodInput2}, []types.RecordOutput{}),
			signers:  []string{s.user1},
			errorMsg: "invalid output count (expected > 0, got: 0)",
		},
		"valid - single output": {
			existing: nil,
			proposed: *types.NewRecord(
				s.recordName, sessionID, *process, []types.RecordInput{*goodInput},
				[]types.RecordOutput{
					{
						Hash:   "justsomeoutput",
						Status: types.ResultStatus_RESULT_STATUS_PASS,
					},
				}),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		"valid - list output": {
			existing: nil,
			proposed: *types.NewRecord(
				recordName2, sessionID, *process, []types.RecordInput{*goodInput2},
				[]types.RecordOutput{
					{
						Hash:   "justsomeoutput",
						Status: types.ResultStatus_RESULT_STATUS_PASS,
					},
					{
						Hash:   "justsomemoreoutput",
						Status: types.ResultStatus_RESULT_STATUS_PASS,
					},
				}),
			signers:  []string{s.user1},
			errorMsg: "",
		},
	}

	for n, tc := range cases {
		s.T().Run(n, func(t *testing.T) {
			err := s.app.MetadataKeeper.ValidateRecordUpdate(s.ctx, tc.existing, tc.proposed, tc.signers)
			if len(tc.errorMsg) != 0 {
				assert.EqualError(t, err, tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
