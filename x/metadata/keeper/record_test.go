package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/stretchr/testify/suite"
)

type RecordKeeperTestSuite struct {
	suite.Suite

	app         *app.App
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
	s.app = simapp.Setup(s.T())
	ctx := s.FreshCtx()
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, s.app.MetadataKeeper)
	s.queryClient = types.NewQueryClient(queryHelper)

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()
	user1Acc := s.app.AccountKeeper.NewAccountWithAddress(ctx, s.user1Addr)
	s.Require().NoError(user1Acc.SetPubKey(s.pubkey1), "SetPubKey user1")
	s.app.AccountKeeper.SetAccount(ctx, user1Acc)

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

	s.contractSpecUUID = uuid.New()
	s.contractSpecID = types.ContractSpecMetadataAddress(s.contractSpecUUID)

	s.recordSpecID = types.RecordSpecMetadataAddress(s.contractSpecUUID, s.recordName)
}

func (s *RecordKeeperTestSuite) FreshCtx() sdk.Context {
	return keeper.AddAuthzCacheToContext(s.app.BaseApp.NewContext(false, tmproto.Header{}))
}

func TestRecordKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(RecordKeeperTestSuite))
}

// func ownerPartyList defined in keeper_test.go

func (s *RecordKeeperTestSuite) TestMetadataRecordGetSetRemove() {
	ctx := s.FreshCtx()
	r, found := s.app.MetadataKeeper.GetRecord(ctx, s.recordID)
	s.NotNil(r)
	s.False(found)

	process := types.NewProcess("processname", &types.Process_Hash{Hash: "HASH"}, "process_method")
	record := types.NewRecord(s.recordName, s.sessionID, *process, []types.RecordInput{}, []types.RecordOutput{}, s.recordSpecID)

	s.NotNil(record)
	s.app.MetadataKeeper.SetRecord(ctx, *record)

	r, found = s.app.MetadataKeeper.GetRecord(ctx, s.recordID)
	s.True(found)
	s.NotNil(r)

	s.app.MetadataKeeper.RemoveRecord(ctx, s.recordID)
	r, found = s.app.MetadataKeeper.GetRecord(ctx, s.recordID)
	s.False(found)
	s.NotNil(r)
}

func (s *RecordKeeperTestSuite) TestMetadataRecordIterator() {
	ctx := s.FreshCtx()
	for i := 1; i <= 10; i++ {
		process := types.NewProcess("processname", &types.Process_Hash{Hash: "HASH"}, "process_method")
		recordName := fmt.Sprintf("%s%v", s.recordName, i)
		record := types.NewRecord(recordName, s.sessionID, *process, []types.RecordInput{}, []types.RecordOutput{}, s.recordSpecID)
		s.app.MetadataKeeper.SetRecord(ctx, *record)
	}
	count := 0
	err := s.app.MetadataKeeper.IterateRecords(ctx, s.scopeID, func(s types.Record) (stop bool) {
		count++
		return false
	})
	s.Require().NoError(err, "IterateRecords")
	s.Assert().Equal(10, count, "iterator should return a full list of records")

}

func (s *RecordKeeperTestSuite) TestValidateDeleteRecord() {
	pt := func(addr string, role types.PartyType, opt bool) types.Party {
		return types.Party{
			Address:  addr,
			Role:     role,
			Optional: opt,
		}
	}
	ptz := func(parties ...types.Party) []types.Party {
		rv := make([]types.Party, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}

	ctx := s.FreshCtx()
	dneScopeUUID := uuid.New()
	dneSessionUUID := uuid.New()
	dneContractSpecUUID := uuid.New()
	dneRecordID := types.RecordMetadataAddress(s.scopeUUID, "does-not-exist")
	user3 := sdk.AccAddress("user_3______________").String()

	scope := types.NewScope(s.scopeID, s.scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, s.user1, false)
	s.app.MetadataKeeper.SetScope(ctx, *scope)

	auditFields := &types.AuditFields{
		CreatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", CreatedDate: time.Now(),
		UpdatedBy: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", UpdatedDate: time.Now(),
		Message: "message",
	}
	sessionParties := []types.Party{
		{Address: s.user1, Role: types.PartyType_PARTY_TYPE_INVESTOR, Optional: true},
		{Address: s.user2, Role: types.PartyType_PARTY_TYPE_SERVICER, Optional: false},
		{Address: user3, Role: types.PartyType_PARTY_TYPE_AFFILIATE, Optional: true},
		{Address: user3, Role: types.PartyType_PARTY_TYPE_INVESTOR, Optional: true},
	}
	session := types.NewSession(s.sessionName, s.sessionID, s.contractSpecID, sessionParties, auditFields)
	s.app.MetadataKeeper.SetSession(ctx, *session)

	sessionWOScopeID := types.SessionMetadataAddress(uuid.New(), s.sessionUUID)
	sessionWOScope := types.NewSession(s.sessionName, sessionWOScopeID, s.contractSpecID, sessionParties, auditFields)
	s.app.MetadataKeeper.SetSession(ctx, *sessionWOScope)

	newRecord := func(name string, sessionID types.MetadataAddress, specID types.MetadataAddress) *types.Record {
		return types.NewRecord(
			name,
			sessionID,
			*types.NewProcess("processname", &types.Process_Hash{Hash: "HASH"}, "process_method"),
			[]types.RecordInput{},
			[]types.RecordOutput{},
			specID,
		)
	}

	// record without a scope, session, or spec.
	recordLone := newRecord("so-alone",
		types.SessionMetadataAddress(dneScopeUUID, dneSessionUUID),
		types.RecordSpecMetadataAddress(dneContractSpecUUID, "so-alone"))
	s.app.MetadataKeeper.SetRecord(ctx, *recordLone)

	// record where scope exists, but session and spec do not.
	recordOnlyScope := newRecord("scope-only",
		types.SessionMetadataAddress(s.scopeUUID, dneSessionUUID),
		types.RecordSpecMetadataAddress(dneContractSpecUUID, "scope-only"))
	s.app.MetadataKeeper.SetRecord(ctx, *recordOnlyScope)

	// record where session exists, but scope and spec do not.
	recordOnlySession := newRecord("session-only",
		sessionWOScopeID,
		types.RecordSpecMetadataAddress(dneContractSpecUUID, "session-only"))
	s.app.MetadataKeeper.SetRecord(ctx, *recordOnlySession)

	// record where spec exists, but scope and session do not.
	reqRoles := []types.PartyType{types.PartyType_PARTY_TYPE_INVESTOR, types.PartyType_PARTY_TYPE_AFFILIATE}
	recSpecOnly := types.NewRecordSpecification(
		types.RecordSpecMetadataAddress(s.contractSpecUUID, "rec-spec-only"),
		"rec-spec-only",
		[]*types.InputSpecification{},
		"type-name",
		types.DefinitionType_DEFINITION_TYPE_RECORD,
		reqRoles,
	)
	s.app.MetadataKeeper.SetRecordSpecification(ctx, *recSpecOnly)
	recordOnlySpec := newRecord(recSpecOnly.Name,
		types.SessionMetadataAddress(dneScopeUUID, dneSessionUUID),
		recSpecOnly.SpecificationId)
	s.app.MetadataKeeper.SetRecord(ctx, *recordOnlySpec)

	// record where session and spec exist, but scope does not.
	recSpecNoScope := types.NewRecordSpecification(
		types.RecordSpecMetadataAddress(s.contractSpecUUID, "no-scope"),
		"no-scope",
		[]*types.InputSpecification{},
		"type-name",
		types.DefinitionType_DEFINITION_TYPE_RECORD,
		reqRoles,
	)
	s.app.MetadataKeeper.SetRecordSpecification(ctx, *recSpecNoScope)
	recordNoScope := newRecord(recSpecNoScope.Name,
		sessionWOScopeID,
		recSpecNoScope.SpecificationId)
	s.app.MetadataKeeper.SetRecord(ctx, *recordNoScope)

	// record where scope, session, and spec exist.
	recSpec := types.NewRecordSpecification(
		s.recordSpecID,
		s.recordName,
		[]*types.InputSpecification{},
		"type-name",
		types.DefinitionType_DEFINITION_TYPE_RECORD,
		reqRoles,
	)
	s.app.MetadataKeeper.SetRecordSpecification(ctx, *recSpec)
	record := newRecord(s.recordName, s.sessionID, s.recordSpecID)
	s.app.MetadataKeeper.SetRecord(ctx, *record)

	owner := types.PartyType_PARTY_TYPE_OWNER
	custodian := types.PartyType_PARTY_TYPE_CUSTODIAN

	rollupScopeSpec := types.ScopeSpecification{
		SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
		Description:     types.NewDescription("rollup_spec", "rollup scope specification", "", ""),
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{custodian},
		ContractSpecIds: []types.MetadataAddress{s.contractSpecID},
	}
	s.app.MetadataKeeper.SetScopeSpecification(ctx, rollupScopeSpec)
	rollupRecSpec := types.RecordSpecification{
		SpecificationId:    types.RecordSpecMetadataAddress(s.contractSpecUUID, "rollup_record"),
		Name:               "rollup_record",
		Inputs:             []*types.InputSpecification{},
		TypeName:           "rollup_type_name",
		ResultType:         types.DefinitionType_DEFINITION_TYPE_RECORD,
		ResponsibleParties: []types.PartyType{custodian},
	}
	s.app.MetadataKeeper.SetRecordSpecification(ctx, rollupRecSpec)

	rollupScope := types.Scope{
		ScopeId:            types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:    rollupScopeSpec.SpecificationId,
		Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, custodian, true), pt(user3, custodian, true)),
		RequirePartyRollup: true,
	}
	s.app.MetadataKeeper.SetScope(ctx, rollupScope)
	rollupRecord := types.Record{
		Name:            rollupRecSpec.Name,
		SessionId:       rollupScope.ScopeId.MustGetAsSessionAddress(uuid.New()),
		Process:         record.Process,
		Inputs:          record.Inputs,
		Outputs:         record.Outputs,
		SpecificationId: rollupRecSpec.SpecificationId,
	}
	s.app.MetadataKeeper.SetRecord(ctx, rollupRecord)
	dneRecordName := "dne_record"
	dneRollupRecSpecID := types.RecordSpecMetadataAddress(s.contractSpecUUID, dneRecordName)
	rollupRecordNoSpec := types.Record{
		Name:            dneRecordName,
		SessionId:       rollupRecord.SessionId,
		Process:         record.Process,
		Inputs:          record.Inputs,
		Outputs:         record.Outputs,
		SpecificationId: dneRollupRecSpecID,
	}
	s.app.MetadataKeeper.SetRecord(ctx, rollupRecordNoSpec)

	cases := []struct {
		name       string
		proposedID types.MetadataAddress
		signers    []string
		expected   []string
	}{
		{
			name:       "record does not exist",
			proposedID: dneRecordID,
			signers:    []string{s.user1},
			expected:   []string{"record does not exist to delete", dneRecordID.String()},
		},
		{
			name:       "no scope session or spec",
			proposedID: recordLone.GetRecordAddress(),
			signers:    []string{},
			expected:   nil,
		},
		{
			name:       "has scope but not session or spec, req signer is signer",
			proposedID: recordOnlyScope.GetRecordAddress(),
			signers:    []string{s.user1},
			expected:   nil,
		},
		{
			name:       "has scope but not session or spec",
			proposedID: recordOnlyScope.GetRecordAddress(),
			signers:    []string{s.user2},
			expected:   []string{"missing signature:", s.user1},
		},
		{
			name:       "has session but not scope or spec, signer was req signer",
			proposedID: recordOnlySession.GetRecordAddress(),
			signers:    []string{s.user2},
			expected:   nil,
		},
		{
			name:       "has session but not scope or spec, signer was not req signer",
			proposedID: recordOnlySession.GetRecordAddress(),
			signers:    []string{s.user1},
			expected:   nil,
		},
		{
			name:       "has spec, no scope or session",
			proposedID: recordOnlySpec.GetRecordAddress(),
			signers:    nil,
			expected:   nil,
		},
		{
			name:       "has session and spec no scope, reqRoles fulfilled",
			proposedID: recordNoScope.GetRecordAddress(),
			signers:    []string{s.user2, user3},
			expected:   nil,
		},
		{
			name:       "has session and spec no scope, missing both req roles",
			proposedID: recordNoScope.GetRecordAddress(),
			signers:    []string{s.user2},
			expected:   nil,
		},
		{
			name:       "has session and spec no scope, missing one req roles",
			proposedID: recordNoScope.GetRecordAddress(),
			signers:    []string{s.user2, s.user1},
			expected:   nil,
		},
		{
			name:       "control",
			proposedID: s.recordID,
			signers:    []string{s.user1, s.user2, user3},
			expected:   nil,
		},
		{
			name:       "missing session required signer",
			proposedID: s.recordID,
			signers:    []string{s.user1, user3},
			expected:   nil,
		},
		{
			name:       "with rollup no rec spec req owner signed",
			proposedID: rollupRecordNoSpec.GetRecordAddress(),
			signers:    []string{s.user1},
			expected:   nil,
		},
		{
			name:       "with rollup no rec spec req owner not signed",
			proposedID: rollupRecordNoSpec.GetRecordAddress(),
			signers:    []string{s.user2},
			expected:   []string{"missing signature:", s.user1},
		},
		{
			name:       "with rollup req owner not signed",
			proposedID: rollupRecord.GetRecordAddress(),
			signers:    []string{s.user2},
			expected:   []string{"missing required signature:", s.user1},
		},
		{
			name:       "with rollup missing signature required by spec",
			proposedID: rollupRecord.GetRecordAddress(),
			signers:    []string{s.user1},
			expected:   []string{"missing signers for roles required by spec:", "CUSTODIAN need 1 have 0"},
		},
		{
			name:       "with rollup has req owner and signature req by spec",
			proposedID: rollupRecord.GetRecordAddress(),
			signers:    []string{s.user1, s.user2},
			expected:   nil,
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {

			msg := &types.MsgDeleteRecordRequest{Signers: tc.signers}
			err := s.app.MetadataKeeper.ValidateDeleteRecord(s.FreshCtx(), tc.proposedID, msg)
			if len(tc.expected) > 0 {
				if s.Assert().Error(err, "ValidateDeleteRecord") {
					for _, exp := range tc.expected {
						s.Assert().ErrorContains(err, exp, "ValidateDeleteRecord")
					}
				}
			} else {
				s.Assert().NoError(err, "ValidateDeleteRecord")
			}
		})
	}
}

func (s *RecordKeeperTestSuite) TestValidateWriteRecord() {
	pt := func(addr string, role types.PartyType, opt bool) types.Party {
		return types.Party{
			Address:  addr,
			Role:     role,
			Optional: opt,
		}
	}
	ptz := func(parties ...types.Party) []types.Party {
		rv := make([]types.Party, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}

	ctx := s.FreshCtx()
	scopeUUID := uuid.New()
	scopeID := types.ScopeMetadataAddress(scopeUUID)
	scope := types.NewScope(scopeID, s.scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, s.user1, false)
	s.app.MetadataKeeper.SetScope(ctx, *scope)

	sessionUUID := uuid.New()
	sessionID := types.SessionMetadataAddress(scopeUUID, sessionUUID)
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
	s.app.MetadataKeeper.SetSession(ctx, *session)

	contractSpec := types.ContractSpecification{
		SpecificationId: s.contractSpecID,
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		ClassName:       "classname",
	}
	s.app.MetadataKeeper.SetContractSpecification(ctx, contractSpec)
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
	s.app.MetadataKeeper.SetRecordSpecification(ctx, *recordSpec)

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
	s.app.MetadataKeeper.SetRecordSpecification(ctx, *recordSpec2)
	recordSpecOtherRole := types.RecordSpecification{
		SpecificationId:    s.contractSpecID.MustGetAsRecordSpecAddress("other_role"),
		Name:               "other_role",
		Inputs:             recordSpec.Inputs,
		TypeName:           recordSpec.TypeName,
		ResultType:         recordSpec.ResultType,
		ResponsibleParties: []types.PartyType{types.PartyType_PARTY_TYPE_INVESTOR},
	}
	s.app.MetadataKeeper.SetRecordSpecification(ctx, recordSpecOtherRole)

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
	goodInputs := []types.RecordInput{*goodInput}
	goodOutputs := []types.RecordOutput{{Hash: "justsomeoutput", Status: types.ResultStatus_RESULT_STATUS_PASS}}

	randomScopeUUID := uuid.New()
	randomScopeID := types.ScopeMetadataAddress(randomScopeUUID)
	randomSessionUUID := uuid.New()
	randomSessionID := types.SessionMetadataAddress(randomScopeUUID, randomSessionUUID)
	randomInScopeSessionID := types.SessionMetadataAddress(scopeUUID, uuid.New())
	missingRecordSpecName := "missing"
	missingRecordSpecID := types.RecordSpecMetadataAddress(s.contractSpecUUID, missingRecordSpecName)

	anotherScopeUUID := uuid.New()
	anotherScopeID := types.ScopeMetadataAddress(anotherScopeUUID)
	anotherScope := types.NewScope(anotherScopeID, s.scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, s.user1, false)
	s.app.MetadataKeeper.SetScope(ctx, *anotherScope)

	anotherSessionUUID := uuid.New()
	anotherSessionID := types.SessionMetadataAddress(anotherScopeUUID, anotherSessionUUID)
	anotherSession := types.NewSession(
		s.sessionName,
		anotherSessionID,
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
	s.app.MetadataKeeper.SetSession(ctx, *anotherSession)
	sessionNewOwner := types.NewSession(
		s.sessionName,
		anotherSessionID,
		s.contractSpecID,
		ownerPartyList(s.user2),
		&types.AuditFields{
			CreatedDate: time.Time{},
			CreatedBy:   s.user2,
			UpdatedDate: time.Time{},
			UpdatedBy:   "",
			Version:     0,
			Message:     "",
		},
	)
	s.app.MetadataKeeper.SetSession(ctx, *sessionNewOwner)
	anotherRecord := types.NewRecord(
		// name string, sessionID MetadataAddress, process Process, inputs []RecordInput, outputs []RecordOutput, specificationID MetadataAddress
		s.recordName,
		anotherSessionID,
		*process,
		[]types.RecordInput{*goodInput},
		[]types.RecordOutput{
			{
				Hash:   "anotheroutput",
				Status: types.ResultStatus_RESULT_STATUS_PASS,
			},
		},
		s.recordSpecID,
	)
	anotherRecordID := types.RecordMetadataAddress(anotherScopeUUID, anotherRecord.Name)
	s.app.MetadataKeeper.SetRecord(ctx, *anotherRecord)

	missingRecordID := types.RecordMetadataAddress(uuid.New(), anotherRecord.Name)

	owner := types.PartyType_PARTY_TYPE_OWNER
	originator := types.PartyType_PARTY_TYPE_ORIGINATOR

	rollupCSpec := types.ContractSpecification{
		SpecificationId: types.ContractSpecMetadataAddress(uuid.New()),
		Description:     types.NewDescription("rollup_c_spec", "rollup contract specification", "", ""),
		OwnerAddresses:  []string{s.user2},
		PartiesInvolved: []types.PartyType{owner},
		Source:          contractSpec.Source,
		ClassName:       contractSpec.ClassName,
	}
	s.app.MetadataKeeper.SetContractSpecification(ctx, rollupCSpec)
	rollupRecSpec := types.RecordSpecification{
		SpecificationId:    rollupCSpec.SpecificationId.MustGetAsRecordSpecAddress("rollup1"),
		Name:               "rollup1",
		Inputs:             []*types.InputSpecification{},
		TypeName:           recordTypeName,
		ResultType:         types.DefinitionType_DEFINITION_TYPE_RECORD_LIST,
		ResponsibleParties: []types.PartyType{originator},
	}
	s.app.MetadataKeeper.SetRecordSpecification(ctx, rollupRecSpec)
	rollupRecSpecAff := types.RecordSpecification{
		SpecificationId:    rollupCSpec.SpecificationId.MustGetAsRecordSpecAddress("rollup_with_affiliate"),
		Name:               "rollup_with_affiliate",
		Inputs:             []*types.InputSpecification{},
		TypeName:           recordTypeName,
		ResultType:         types.DefinitionType_DEFINITION_TYPE_RECORD_LIST,
		ResponsibleParties: []types.PartyType{types.PartyType_PARTY_TYPE_AFFILIATE},
	}
	s.app.MetadataKeeper.SetRecordSpecification(ctx, rollupRecSpecAff)

	user3 := sdk.AccAddress("user_three__________").String()
	user4 := sdk.AccAddress("user_four___________").String()
	rollupScope := types.Scope{
		ScopeId:            types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:    s.scopeSpecID,
		Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, originator, true), pt(user3, originator, true)),
		RequirePartyRollup: true,
	}
	s.app.MetadataKeeper.SetScope(ctx, rollupScope)
	rollupSession1 := types.Session{
		SessionId:       rollupScope.ScopeId.MustGetAsSessionAddress(uuid.New()),
		SpecificationId: rollupCSpec.SpecificationId,
		Parties:         ptz(pt(user4, owner, false), pt(s.user2, originator, true)),
		Name:            "rollup_session_1",
		Context:         session.Context,
		Audit:           session.Audit,
	}
	s.app.MetadataKeeper.SetSession(ctx, rollupSession1)
	rollupSession2 := types.Session{
		SessionId:       rollupScope.ScopeId.MustGetAsSessionAddress(uuid.New()),
		SpecificationId: rollupCSpec.SpecificationId,
		Parties:         ptz(pt(user4, owner, false), pt(user3, originator, true)),
		Name:            "rollup_session_1",
		Context:         session.Context,
		Audit:           session.Audit,
	}
	s.app.MetadataKeeper.SetSession(ctx, rollupSession2)

	cases := map[string]struct {
		existing         *types.Record
		origOutputHashes []string
		proposed         *types.Record
		partiesInvolved  []types.Party
		signers          []string
		errorMsg         string
	}{
		"validate basic called on proposed": {
			existing:        nil,
			proposed:        &types.Record{},
			signers:         []string{s.user1},
			partiesInvolved: ownerPartyList(s.user1),
			errorMsg:        "address is empty",
		},
		"existing and proposed names do not match": {
			existing:         types.NewRecord("notamatch", sessionID, *process, []types.RecordInput{}, []types.RecordOutput{}, s.recordSpecID),
			origOutputHashes: []string{},
			proposed:         types.NewRecord("not-a-match", sessionID, *process, []types.RecordInput{}, []types.RecordOutput{}, s.recordSpecID),
			signers:          []string{s.user1},
			partiesInvolved:  ownerPartyList(s.user1),
			errorMsg:         "the Name field of records cannot be changed",
		},
		"original session id not found": {
			existing:         types.NewRecord(s.recordName, randomSessionID, *process, []types.RecordInput{}, []types.RecordOutput{}, s.recordSpecID),
			origOutputHashes: []string{},
			proposed: types.NewRecord(s.recordName, sessionID, *process, []types.RecordInput{*goodInput},
				[]types.RecordOutput{
					{
						Hash:   "justsomeoutput",
						Status: types.ResultStatus_RESULT_STATUS_PASS,
					},
				}, s.recordSpecID),
			signers:         []string{s.user1},
			partiesInvolved: ownerPartyList(s.user1),
			errorMsg:        "",
		},
		"scope not found": {
			existing:        nil,
			proposed:        types.NewRecord(s.recordName, randomSessionID, *process, []types.RecordInput{}, []types.RecordOutput{}, s.recordSpecID),
			signers:         []string{s.user1},
			partiesInvolved: ownerPartyList(s.user1),
			errorMsg:        fmt.Sprintf("scope not found with id %s", randomScopeID),
		},
		"missing signature from existing owner": {
			existing:        nil,
			proposed:        types.NewRecord(s.recordName, sessionID, *process, []types.RecordInput{}, []types.RecordOutput{}, s.recordSpecID),
			signers:         []string{},
			partiesInvolved: ownerPartyList(s.user1),
			errorMsg:        fmt.Sprintf("missing signature: %s", s.user1),
		},
		"session not found": {
			existing:        nil,
			proposed:        types.NewRecord(s.recordName, randomInScopeSessionID, *process, []types.RecordInput{}, []types.RecordOutput{}, s.recordSpecID),
			signers:         []string{s.user1},
			partiesInvolved: ownerPartyList(s.user1),
			errorMsg:        fmt.Sprintf("session not found for session id %s", randomInScopeSessionID),
		},
		"record specification not found": {
			existing:        nil,
			proposed:        types.NewRecord(missingRecordSpecName, sessionID, *process, []types.RecordInput{}, []types.RecordOutput{}, missingRecordSpecID),
			signers:         []string{s.user1},
			partiesInvolved: ownerPartyList(s.user1),
			errorMsg: fmt.Sprintf("record specification not found for record specification id %s (contract spec id %s and record name %q)",
				missingRecordSpecID, s.contractSpecID, missingRecordSpecName),
		},
		"missing input": {
			existing:        nil,
			proposed:        types.NewRecord(s.recordName, sessionID, *process, []types.RecordInput{*otherInput}, []types.RecordOutput{}, s.recordSpecID),
			signers:         []string{s.user1},
			partiesInvolved: ownerPartyList(s.user1),
			errorMsg:        fmt.Sprintf("missing input [%s]", goodInput.Name),
		},
		"extra input": {
			existing:        nil,
			proposed:        types.NewRecord(s.recordName, sessionID, *process, []types.RecordInput{*goodInput, *otherInput}, []types.RecordOutput{}, s.recordSpecID),
			signers:         []string{s.user1},
			partiesInvolved: ownerPartyList(s.user1),
			errorMsg:        fmt.Sprintf("extra input [%s]", otherInput.Name),
		},
		"duplicate input": {
			existing:        nil,
			proposed:        types.NewRecord(s.recordName, sessionID, *process, []types.RecordInput{*goodInput, *goodInput}, []types.RecordOutput{}, s.recordSpecID),
			signers:         []string{s.user1},
			partiesInvolved: ownerPartyList(s.user1),
			errorMsg:        fmt.Sprintf("input name %s provided twice", goodInput.Name),
		},
		"input type name wrong": {
			existing: nil,
			proposed: types.NewRecord(
				s.recordName, sessionID, *process,
				[]types.RecordInput{
					{
						Name:     goodInput.Name,
						Source:   goodInput.Source,
						TypeName: "bad type name",
						Status:   goodInput.Status,
					},
				},
				[]types.RecordOutput{},
				s.recordSpecID),
			signers:         []string{s.user1},
			partiesInvolved: ownerPartyList(s.user1),
			errorMsg: fmt.Sprintf("input %s has TypeName %s but spec calls for %s",
				goodInput.Name, "bad type name", inputSpec.TypeName),
		},
		"input source type wrong": {
			existing: nil,
			proposed: types.NewRecord(
				s.recordName, sessionID, *process,
				[]types.RecordInput{
					{
						Name:     goodInput.Name,
						Source:   &types.RecordInput_RecordId{RecordId: anotherRecordID},
						TypeName: goodInput.TypeName,
						Status:   types.RecordInputStatus_Record,
					},
				},
				[]types.RecordOutput{},
				s.recordSpecID),
			signers:         []string{s.user1},
			partiesInvolved: ownerPartyList(s.user1),
			errorMsg: fmt.Sprintf("input %s has source type %s but spec calls for %s",
				goodInput.Name, "record", "hash"),
		},
		"input source record id does not exist": {
			existing: nil,
			proposed: types.NewRecord(
				s.recordName, sessionID, *process,
				[]types.RecordInput{
					{
						Name:     goodInput.Name,
						Source:   &types.RecordInput_RecordId{RecordId: missingRecordID},
						TypeName: goodInput.TypeName,
						Status:   types.RecordInputStatus_Record,
					},
				},
				[]types.RecordOutput{},
				s.recordSpecID),
			signers:         []string{s.user1},
			partiesInvolved: ownerPartyList(s.user1),
			errorMsg: fmt.Sprintf("input %s source record id %s not found",
				goodInput.Name, missingRecordID),
		},
		"output count wrong - record - zero": {
			existing: nil,
			proposed: types.NewRecord(
				s.recordName, sessionID, *process, []types.RecordInput{*goodInput}, []types.RecordOutput{}, s.recordSpecID),
			signers:         []string{s.user1},
			partiesInvolved: ownerPartyList(s.user1),
			errorMsg:        "invalid output count (expected: 1, got: 0)",
		},
		"output count wrong - record - two": {
			existing: nil,
			proposed: types.NewRecord(
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
				},
				s.recordSpecID),
			signers:         []string{s.user1},
			partiesInvolved: ownerPartyList(s.user1),
			errorMsg:        "invalid output count (expected: 1, got: 2)",
		},
		"output count wrong - record list - zero": {
			existing:        nil,
			proposed:        types.NewRecord(recordName2, sessionID, *process, []types.RecordInput{*goodInput2}, []types.RecordOutput{}, recordSpec2ID),
			signers:         []string{s.user1},
			partiesInvolved: ownerPartyList(s.user1),
			errorMsg:        "invalid output count (expected > 0, got: 0)",
		},
		"valid - empty specification id": {
			existing: nil,
			proposed: types.NewRecord(
				s.recordName, sessionID, *process, []types.RecordInput{*goodInput},
				[]types.RecordOutput{
					{
						Hash:   "justsomeoutput",
						Status: types.ResultStatus_RESULT_STATUS_PASS,
					},
				},
				nil),
			signers:         []string{s.user1},
			partiesInvolved: ownerPartyList(s.user1),
			errorMsg:        "",
		},
		"valid - single output": {
			existing: nil,
			proposed: types.NewRecord(
				s.recordName, sessionID, *process, []types.RecordInput{*goodInput},
				[]types.RecordOutput{
					{
						Hash:   "justsomeoutput",
						Status: types.ResultStatus_RESULT_STATUS_PASS,
					},
				},
				s.recordSpecID),
			signers:         []string{s.user1},
			partiesInvolved: ownerPartyList(s.user1),
			errorMsg:        "",
		},
		"valid - list output": {
			existing: nil,
			proposed: types.NewRecord(
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
				},
				recordSpec2ID),
			signers:         []string{s.user1},
			partiesInvolved: ownerPartyList(s.user1),
			errorMsg:        "",
		},
		"missing role required by spec": {
			existing: nil,
			proposed: &types.Record{
				Name:            recordSpecOtherRole.Name,
				SessionId:       sessionID,
				Process:         *process,
				Inputs:          goodInputs,
				Outputs:         goodOutputs,
				SpecificationId: recordSpecOtherRole.SpecificationId,
			},
			signers:  []string{s.user1},
			errorMsg: "missing roles required by spec: INVESTOR need 1 have 0",
		},
		"missing session signature": {
			existing: nil,
			proposed: &types.Record{
				Name:            s.recordName,
				SessionId:       sessionID,
				Process:         *process,
				Inputs:          goodInputs,
				Outputs:         goodOutputs,
				SpecificationId: recordSpecID,
			},
			signers:  []string{s.user2},
			errorMsg: "missing signature: " + s.user1,
		},
		"missing previous session signature": {
			existing: &types.Record{
				Name:            s.recordName,
				SessionId:       sessionID,
				Process:         *process,
				Inputs:          goodInputs,
				Outputs:         goodOutputs,
				SpecificationId: recordSpecID,
			},
			proposed: &types.Record{
				Name:            s.recordName,
				SessionId:       sessionNewOwner.SessionId,
				Process:         *process,
				Inputs:          goodInputs,
				Outputs:         goodOutputs,
				SpecificationId: recordSpecID,
			},
			signers:  []string{s.user2},
			errorMsg: "missing signature: " + s.user1,
		},
		"with rollup missing signature from role required by spec": {
			proposed: &types.Record{
				Name:            rollupRecSpec.Name,
				SessionId:       rollupSession1.SessionId,
				Process:         *process,
				Inputs:          goodInputs,
				Outputs:         goodOutputs,
				SpecificationId: rollupRecSpec.SpecificationId,
			},
			signers:  []string{s.user1, user4, user3},
			errorMsg: "missing signers for roles required by spec: ORIGINATOR need 1 have 0",
		},
		"with rollup missing scope req signature": {
			proposed: &types.Record{
				Name:            rollupRecSpec.Name,
				SessionId:       rollupSession1.SessionId,
				Process:         *process,
				Inputs:          goodInputs,
				Outputs:         goodOutputs,
				SpecificationId: rollupRecSpec.SpecificationId,
			},
			signers:  []string{user4, s.user2},
			errorMsg: "missing required signature: " + s.user1 + " (OWNER)",
		},
		"with rollup missing session req signature": {
			proposed: &types.Record{
				Name:            rollupRecSpec.Name,
				SessionId:       rollupSession1.SessionId,
				Process:         *process,
				Inputs:          goodInputs,
				Outputs:         goodOutputs,
				SpecificationId: rollupRecSpec.SpecificationId,
			},
			signers:  []string{s.user1, s.user2},
			errorMsg: "missing required signature: " + user4 + " (OWNER)",
		},
		"with rollup missing previous session req signature": {
			existing: &types.Record{
				Name:            rollupRecSpec.Name,
				SessionId:       rollupSession1.SessionId,
				Process:         *process,
				Inputs:          goodInputs,
				Outputs:         goodOutputs,
				SpecificationId: rollupRecSpec.SpecificationId,
			},
			proposed: &types.Record{
				Name:            rollupRecSpec.Name,
				SessionId:       rollupSession2.SessionId,
				Process:         *process,
				Inputs:          goodInputs,
				Outputs:         goodOutputs,
				SpecificationId: rollupRecSpec.SpecificationId,
			},
			signers:  []string{s.user1, user3},
			errorMsg: "missing required signature: " + user4 + " (OWNER)",
		},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			msg := &types.MsgWriteRecordRequest{
				Record:  *tc.proposed,
				Signers: tc.signers,
			}
			err := s.app.MetadataKeeper.ValidateWriteRecord(s.FreshCtx(), tc.existing, msg)
			if len(tc.errorMsg) != 0 {
				s.Assert().EqualError(err, tc.errorMsg, "ValidateWriteRecord expected error")
			} else {
				s.Assert().NoError(err, "ValidateWriteRecord unexpected error")
				s.Assert().NotEmpty(msg.Record.SpecificationId, "proposed.SpecificationId after ValidateWriteRecord")
			}
		})
	}
}
