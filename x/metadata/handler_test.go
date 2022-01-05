package metadata_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/google/uuid"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/metadata"
	"github.com/provenance-io/provenance/x/metadata/types"
	"github.com/provenance-io/provenance/x/metadata/types/p8e"
)

type MetadataHandlerTestSuite struct {
	suite.Suite

	app     *app.App
	ctx     sdk.Context
	handler sdk.Handler

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress
}

func (s *MetadataHandlerTestSuite) SetupTest() {
	s.app = app.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.handler = metadata.NewHandler(s.app.MetadataKeeper)

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	privKey, _ := secp256r1.GenPrivKey()
	s.pubkey2 = privKey.PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))
	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user2Addr))
}

func TestMetadataHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(MetadataHandlerTestSuite))
}

func createContractSpec(inputSpecs []*p8e.DefinitionSpec, outputSpec p8e.OutputSpec, definitionSpec p8e.DefinitionSpec) p8e.ContractSpec {
	return p8e.ContractSpec{ConsiderationSpecs: []*p8e.ConsiderationSpec{
		{FuncName: "additionalParties",
			InputSpecs:       inputSpecs,
			OutputSpec:       &outputSpec,
			ResponsibleParty: 1,
		},
	},
		Definition:      &definitionSpec,
		InputSpecs:      inputSpecs,
		PartiesInvolved: []p8e.PartyType{p8e.PartyType_PARTY_TYPE_AFFILIATE},
	}
}

func createDefinitionSpec(name string, classname string, reference p8e.ProvenanceReference, defType int) p8e.DefinitionSpec {
	return p8e.DefinitionSpec{
		Name: name,
		ResourceLocation: &p8e.Location{Classname: classname,
			Ref: &reference,
		},
		Type: 1,
	}
}

// TODO: WriteScope tests
// TODO: DeleteScope tests

func (s MetadataHandlerTestSuite) TestWriteSession() {
	cSpec := types.ContractSpecification{
		SpecificationId: types.ContractSpecMetadataAddress(uuid.New()),
		Description:     nil,
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		Source:          types.NewContractSpecificationSourceHash("somesource"),
		ClassName:       "someclass",
	}
	s.app.MetadataKeeper.SetContractSpecification(s.ctx, cSpec)
	sSpec := types.ScopeSpecification{
		SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
		Description:     nil,
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		ContractSpecIds: []types.MetadataAddress{cSpec.SpecificationId},
	}
	s.app.MetadataKeeper.SetScopeSpecification(s.ctx, sSpec)

	scopeUUID := uuid.New()
	scope := types.Scope{
		ScopeId:         types.ScopeMetadataAddress(scopeUUID),
		SpecificationId: sSpec.SpecificationId,
		Owners: []types.Party{{
			Address: s.user1,
			Role:    types.PartyType_PARTY_TYPE_OWNER,
		}},
		DataAccess:        nil,
		ValueOwnerAddress: "",
	}
	s.app.MetadataKeeper.SetScope(s.ctx, scope)

	someBytes, err := base64.StdEncoding.DecodeString("ChFIRUxMTyBQUk9WRU5BTkNFIQ==")
	require.NoError(s.T(), err, "trying to create someBytes")

	cases := []struct {
		name     string
		session  types.Session
		signers  []string
		errorMsg string
	}{
		{
			"valid without context",
			types.Session{
				SessionId:       types.SessionMetadataAddress(scopeUUID, uuid.New()),
				SpecificationId: cSpec.SpecificationId,
				Parties:         scope.Owners,
				Name:            "someclass",
				Context:         nil,
				Audit:           nil,
			},
			[]string{s.user1},
			"",
		},
		{
			"valid with context",
			types.Session{
				SessionId:       types.SessionMetadataAddress(scopeUUID, uuid.New()),
				SpecificationId: cSpec.SpecificationId,
				Parties:         scope.Owners,
				Name:            "someclass",
				Context:         someBytes,
				Audit:           nil,
			},
			[]string{s.user1},
			"",
		},
	}

	for _, tc := range cases {
		s.T().Run(tc.name, func(t *testing.T) {
			msg := types.MsgWriteSessionRequest{
				Session:             tc.session,
				Signers:             tc.signers,
				SessionIdComponents: nil,
				SpecUuid:            "",
			}
			_, err := s.handler(s.ctx, &msg)
			if len(tc.errorMsg) > 0 {
				assert.EqualError(t, err, tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s MetadataHandlerTestSuite) TestWriteDeleteRecord() {
	cSpecUUID := uuid.New()
	cSpec := types.ContractSpecification{
		SpecificationId: types.ContractSpecMetadataAddress(cSpecUUID),
		Description:     nil,
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		Source:          types.NewContractSpecificationSourceHash("somesource1"),
		ClassName:       "someclass1",
	}
	s.app.MetadataKeeper.SetContractSpecification(s.ctx, cSpec)
	defer func() {
		s.Assert().NoError(s.app.MetadataKeeper.RemoveContractSpecification(s.ctx, cSpec.SpecificationId), "removing contract spec")
	}()

	sSpecUUID := uuid.New()
	sSpec := types.ScopeSpecification{
		SpecificationId: types.ScopeSpecMetadataAddress(sSpecUUID),
		Description:     nil,
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		ContractSpecIds: []types.MetadataAddress{cSpec.SpecificationId},
	}
	s.app.MetadataKeeper.SetScopeSpecification(s.ctx, sSpec)
	defer func() {
		s.Assert().NoError(s.app.MetadataKeeper.RemoveScopeSpecification(s.ctx, sSpec.SpecificationId), "removing scope spec")
	}()

	rSpec := types.RecordSpecification{
		SpecificationId: types.RecordSpecMetadataAddress(cSpecUUID, "record"),
		Name:            "record",
		Inputs: []*types.InputSpecification{
			{
				Name:     "ri1",
				TypeName: "string",
				Source:   types.NewInputSpecificationSourceHash("ri1hash"),
			},
		},
		TypeName:           "string",
		ResultType:         types.DefinitionType_DEFINITION_TYPE_RECORD_LIST,
		ResponsibleParties: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
	}
	s.app.MetadataKeeper.SetRecordSpecification(s.ctx, rSpec)
	defer func() {
		s.Assert().NoError(s.app.MetadataKeeper.RemoveRecordSpecification(s.ctx, rSpec.SpecificationId), "removing record spec 1")
	}()

	scopeUUID := uuid.New()
	scope := types.Scope{
		ScopeId:         types.ScopeMetadataAddress(scopeUUID),
		SpecificationId: sSpec.SpecificationId,
		Owners: []types.Party{{
			Address: s.user1,
			Role:    types.PartyType_PARTY_TYPE_OWNER,
		}},
		DataAccess:        nil,
		ValueOwnerAddress: "",
	}
	s.app.MetadataKeeper.SetScope(s.ctx, scope)
	defer s.app.MetadataKeeper.RemoveScope(s.ctx, scope.ScopeId)

	session1UUID := uuid.New()
	session1 := types.Session{
		SessionId:       types.SessionMetadataAddress(scopeUUID, session1UUID),
		SpecificationId: cSpec.SpecificationId,
		Parties:         ownerPartyList(s.user1),
		Name:            "someclass1",
	}
	s.app.MetadataKeeper.SetSession(s.ctx, session1)
	defer s.app.MetadataKeeper.RemoveSession(s.ctx, session1.SessionId)

	session2UUID := uuid.New()
	session2 := types.Session{
		SessionId:       types.SessionMetadataAddress(scopeUUID, session2UUID),
		SpecificationId: cSpec.SpecificationId,
		Parties:         ownerPartyList(s.user1),
		Name:            "someclass1",
	}
	s.app.MetadataKeeper.SetSession(s.ctx, session2)
	defer s.app.MetadataKeeper.RemoveSession(s.ctx, session2.SessionId)

	record := types.Record{
		Name:      rSpec.Name,
		SessionId: session1.SessionId,
		Process: types.Process{
			ProcessId: &types.Process_Hash{Hash: "rprochash"},
			Name:      "rproc",
			Method:    "rprocmethod",
		},
		Inputs: []types.RecordInput{
			{
				Name:     rSpec.Inputs[0].Name,
				Source:   &types.RecordInput_Hash{Hash: "rhash"},
				TypeName: rSpec.Inputs[0].TypeName,
				Status:   types.RecordInputStatus_Proposed,
			},
		},
		Outputs: []types.RecordOutput{
			{
				Hash:   "rout1",
				Status: types.ResultStatus_RESULT_STATUS_PASS,
			},
			{
				Hash:   "rout2",
				Status: types.ResultStatus_RESULT_STATUS_PASS,
			},
		},
		SpecificationId: rSpec.SpecificationId,
	}
	recordID := types.RecordMetadataAddress(scopeUUID, rSpec.Name)
	// Not adding the record here because we're testing that stuff.

	s.T().Run("write invalid record", func(t *testing.T) {
		// Make a record with an unknown spec id. Try to write it and expect an error.
		badRecord := types.Record{
			Name:      rSpec.Name,
			SessionId: session1.SessionId,
			Process: types.Process{
				ProcessId: &types.Process_Hash{Hash: "badrprochash"},
				Name:      "badrproc",
				Method:    "badrprocmethod",
			},
			Inputs: []types.RecordInput{
				{
					Name:     rSpec.Inputs[0].Name,
					Source:   &types.RecordInput_Hash{Hash: "badrhash"},
					TypeName: rSpec.Inputs[0].TypeName,
					Status:   types.RecordInputStatus_Proposed,
				},
			},
			Outputs: []types.RecordOutput{
				{
					Hash:   "badrout1",
					Status: types.ResultStatus_RESULT_STATUS_PASS,
				},
				{
					Hash:   "badrout2",
					Status: types.ResultStatus_RESULT_STATUS_PASS,
				},
			},
			SpecificationId: types.RecordSpecMetadataAddress(uuid.New(), rSpec.Name),
		}
		msg := types.MsgWriteRecordRequest{
			Record:              badRecord,
			Signers:             []string{s.user1},
			SessionIdComponents: nil,
			ContractSpecUuid:    "",
			Parties:             ownerPartyList(s.user1),
		}
		_, err := s.handler(s.ctx, &msg)
		require.Error(t, err, "sending bad MsgWriteRecordRequest")
		require.Contains(t, err.Error(), "proposed specification id")
		require.Contains(t, err.Error(), "does not match expected")
	})

	s.T().Run("write record to session 1", func(t *testing.T) {
		msg := types.MsgWriteRecordRequest{
			Record:              record,
			Signers:             []string{s.user1},
			SessionIdComponents: nil,
			ContractSpecUuid:    "",
			Parties:             ownerPartyList(s.user1),
		}
		_, err := s.handler(s.ctx, &msg)
		require.NoError(t, err, "sending MsgWriteRecordRequest")
		r, rok := s.app.MetadataKeeper.GetRecord(s.ctx, recordID)
		if assert.True(t, rok, "GetRecord bool") {
			assert.Equal(t, record, r, "GetRecord record")
		}
	})

	s.T().Run("Update record to other session", func(t *testing.T) {
		record.SessionId = session2.SessionId
		msg := types.MsgWriteRecordRequest{
			Record:              record,
			Signers:             []string{s.user1},
			SessionIdComponents: nil,
			ContractSpecUuid:    "",
			Parties:             ownerPartyList(s.user1),
		}
		_, err := s.handler(s.ctx, &msg)
		require.NoError(t, err, "sending MsgWriteRecordRequest")
		r, rok := s.app.MetadataKeeper.GetRecord(s.ctx, recordID)
		if assert.True(t, rok, "GetRecord bool") {
			assert.Equal(t, record, r, "GetRecord record")
		}
		// Make sure the session was deleted since it's now empty.
		_, sok := s.app.MetadataKeeper.GetSession(s.ctx, session1.SessionId)
		assert.False(t, sok, "GetSession session 1 bool")
	})

	s.T().Run("delete the record", func(t *testing.T) {
		msg := types.MsgDeleteRecordRequest{
			RecordId: recordID,
			Signers:  []string{s.user1},
		}
		_, err := s.handler(s.ctx, &msg)
		require.NoError(t, err, "sending MsgDeleteRecordRequest")
		_, rok := s.app.MetadataKeeper.GetRecord(s.ctx, recordID)
		assert.False(t, rok, "GetRecord bool")
		// Make sure the session was deleted since it's now empty.
		_, sok := s.app.MetadataKeeper.GetSession(s.ctx, session2.SessionId)
		assert.False(t, sok, "GetSession session 2 bool")
	})
}

// TODO: WriteScopeSpecification tests
// TODO: DeleteScopeSpecification tests
// TODO: WriteContractSpecification tests
// TODO: DeleteContractSpecification tests

func (s MetadataHandlerTestSuite) TestAddContractSpecToScopeSpec() {
	cSpec := types.ContractSpecification{
		SpecificationId: types.ContractSpecMetadataAddress(uuid.New()),
		Description:     nil,
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		Source:          types.NewContractSpecificationSourceHash("somesource"),
		ClassName:       "someclass",
	}
	s.app.MetadataKeeper.SetContractSpecification(s.ctx, cSpec)
	sSpec := types.ScopeSpecification{
		SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
		Description:     nil,
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		ContractSpecIds: []types.MetadataAddress{cSpec.SpecificationId},
	}
	s.app.MetadataKeeper.SetScopeSpecification(s.ctx, sSpec)

	cSpec2 := types.ContractSpecification{
		SpecificationId: types.ContractSpecMetadataAddress(uuid.New()),
		Description:     nil,
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		Source:          types.NewContractSpecificationSourceHash("somesource"),
		ClassName:       "someclass",
	}
	s.app.MetadataKeeper.SetContractSpecification(s.ctx, cSpec2)

	unknownContractSpecId := types.ContractSpecMetadataAddress(uuid.New())
	unknownScopeSpecId := types.ScopeSpecMetadataAddress(uuid.New())

	cases := []struct {
		name           string
		contractSpecId types.MetadataAddress
		scopeSpecId    types.MetadataAddress
		signers        []string
		errorMsg       string
	}{
		{
			"fail to add contract spec, cannot find contract spec",
			unknownContractSpecId,
			sSpec.SpecificationId,
			[]string{s.user1},
			fmt.Sprintf("contract specification not found with id %s", unknownContractSpecId),
		},
		{
			"fail to add contract spec, cannot find scope spec",
			cSpec2.SpecificationId,
			unknownScopeSpecId,
			[]string{s.user1},
			fmt.Sprintf("scope specification not found with id %s", unknownScopeSpecId),
		},
		{
			"fail to add contract spec, scope spec already has contract spec",
			cSpec.SpecificationId,
			sSpec.SpecificationId,
			[]string{s.user1},
			fmt.Sprintf("scope spec %s already contains contract spec %s", sSpec.SpecificationId, cSpec.SpecificationId),
		},
		{
			"should successfully add contract spec to scope spec",
			cSpec2.SpecificationId,
			sSpec.SpecificationId,
			[]string{s.user1},
			"",
		},
	}

	for _, tc := range cases {
		s.T().Run(tc.name, func(t *testing.T) {
			msg := types.MsgAddContractSpecToScopeSpecRequest{
				ContractSpecificationId: tc.contractSpecId,
				ScopeSpecificationId:    tc.scopeSpecId,
				Signers:                 tc.signers,
			}
			_, err := s.handler(s.ctx, &msg)
			if len(tc.errorMsg) > 0 {
				assert.EqualError(t, err, tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s MetadataHandlerTestSuite) TestDeleteContractSpecFromScopeSpec() {
	cSpec := types.ContractSpecification{
		SpecificationId: types.ContractSpecMetadataAddress(uuid.New()),
		Description:     nil,
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		Source:          types.NewContractSpecificationSourceHash("somesource"),
		ClassName:       "someclass",
	}
	s.app.MetadataKeeper.SetContractSpecification(s.ctx, cSpec)
	cSpec2 := types.ContractSpecification{
		SpecificationId: types.ContractSpecMetadataAddress(uuid.New()),
		Description:     nil,
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		Source:          types.NewContractSpecificationSourceHash("somesource"),
		ClassName:       "someclass",
	}
	s.app.MetadataKeeper.SetContractSpecification(s.ctx, cSpec2)
	sSpec := types.ScopeSpecification{
		SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
		Description:     nil,
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
		ContractSpecIds: []types.MetadataAddress{cSpec.SpecificationId, cSpec2.SpecificationId},
	}
	s.app.MetadataKeeper.SetScopeSpecification(s.ctx, sSpec)

	unknownContractSpecId := types.ContractSpecMetadataAddress(uuid.New())
	unknownScopeSpecId := types.ScopeSpecMetadataAddress(uuid.New())

	cases := []struct {
		name           string
		contractSpecId types.MetadataAddress
		scopeSpecId    types.MetadataAddress
		signers        []string
		errorMsg       string
	}{
		{
			"fail to delete contract spec from scope spec, cannot find contract spec",
			unknownContractSpecId,
			sSpec.SpecificationId,
			[]string{s.user1},
			fmt.Sprintf("contract specification not found with id %s", unknownContractSpecId),
		},
		{
			"fail to delete contract spec from scope spec, cannot find scope spec",
			cSpec2.SpecificationId,
			unknownScopeSpecId,
			[]string{s.user1},
			fmt.Sprintf("scope specification not found with id %s", unknownScopeSpecId),
		},
		{
			"should succeed to add contract spec to scope spec",
			cSpec2.SpecificationId,
			sSpec.SpecificationId,
			[]string{s.user1},
			"",
		},
		{
			"fail to delete contract spec from scope spec, scope spec does not contain contract spec",
			cSpec2.SpecificationId,
			sSpec.SpecificationId,
			[]string{s.user1},
			fmt.Sprintf("contract specification %s not found on scope specification id %s", cSpec2.SpecificationId, sSpec.SpecificationId),
		},
	}

	for _, tc := range cases {
		s.T().Run(tc.name, func(t *testing.T) {
			msg := types.MsgDeleteContractSpecFromScopeSpecRequest{
				ContractSpecificationId: tc.contractSpecId,
				ScopeSpecificationId:    tc.scopeSpecId,
				Signers:                 tc.signers,
			}
			_, err := s.handler(s.ctx, &msg)
			if len(tc.errorMsg) > 0 {
				assert.EqualError(t, err, tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TODO: WriteRecordSpecification tests
// TODO: DeleteRecordSpecification tests

func (s MetadataHandlerTestSuite) TestAddP8EContractSpec() {
	validDefSpec := createDefinitionSpec("perform_input_checks", "io.provenance.loan.LoanProtos$PartiesList", p8e.ProvenanceReference{Hash: "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw=="}, 1)
	invalidDefSpec := createDefinitionSpec("perform_action", "", p8e.ProvenanceReference{Hash: "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw=="}, 1)

	cases := []struct {
		name     string
		v39CSpec p8e.ContractSpec
		signers  []string
		errorMsg string
	}{
		{
			"should successfully ADD contract spec in from v38 to v40",
			createContractSpec([]*p8e.DefinitionSpec{&validDefSpec}, p8e.OutputSpec{Spec: &validDefSpec}, validDefSpec),
			[]string{s.user1},
			"",
		},
		{
			"should successfully UPDATE contract spec in from v38 to v40",
			createContractSpec([]*p8e.DefinitionSpec{&validDefSpec}, p8e.OutputSpec{Spec: &validDefSpec}, validDefSpec),
			[]string{s.user1},
			"",
		},
		{
			"should fail to add due to invalid signers",
			createContractSpec([]*p8e.DefinitionSpec{&validDefSpec}, p8e.OutputSpec{Spec: &validDefSpec}, validDefSpec),
			[]string{s.user2},
			fmt.Sprintf("missing signature from existing owner %s; required for update", s.user1),
		},
		{
			"should fail on converting contract validate basic",
			createContractSpec([]*p8e.DefinitionSpec{&invalidDefSpec}, p8e.OutputSpec{Spec: &validDefSpec}, validDefSpec),
			[]string{s.user1},
			"input specification type name cannot be empty",
		},
	}

	for _, tc := range cases {
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := s.handler(s.ctx, &types.MsgWriteP8EContractSpecRequest{Contractspec: tc.v39CSpec, Signers: tc.signers})
			if len(tc.errorMsg) > 0 {
				assert.EqualError(t, err, tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TODO: P8EMemorializeContract tests
// TODO: BindOSLocator tests
// TODO: DeleteOSLocator tests
// TODO: ModifyOSLocator tests

func ownerPartyList(addresses ...string) []types.Party {
	retval := make([]types.Party, len(addresses))
	for i, addr := range addresses {
		retval[i] = types.Party{Address: addr, Role: types.PartyType_PARTY_TYPE_OWNER}
	}
	return retval
}

func (s MetadataHandlerTestSuite) TestAddAndDeleteScopeOwners() {
	scopeSpecID := types.ScopeSpecMetadataAddress(uuid.New())
	scopeSpec := types.NewScopeSpecification(scopeSpecID, nil, []string{s.user1}, []types.PartyType{types.PartyType_PARTY_TYPE_OWNER}, []types.MetadataAddress{})
	scopeID := types.ScopeMetadataAddress(uuid.New())
	scope := types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, "")
	dneScopeID := types.ScopeMetadataAddress(uuid.New())
	user3 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()

	cases := []struct {
		name     string
		msg      sdk.Msg
		signers  []string
		errorMsg string
	}{
		{
			"setup test with new scope specification",
			types.NewMsgWriteScopeSpecificationRequest(*scopeSpec, []string{s.user1}),
			[]string{s.user1},
			"",
		},
		{
			"setup test with new scope",
			types.NewMsgWriteScopeRequest(*scope, []string{s.user1}),
			[]string{s.user1},
			"",
		},
		{
			"should fail to ADD owners, msg validate basic failure",
			types.NewMsgAddScopeOwnerRequest(scopeID, []types.Party{}, []string{s.user1}),
			[]string{s.user1},
			"invalid owners: at least one party is required",
		},
		{
			"should fail to ADD owners, can not find scope",
			types.NewMsgAddScopeOwnerRequest(dneScopeID, []types.Party{{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER}}, []string{s.user1}),
			[]string{s.user1},
			fmt.Sprintf("scope not found with id %s", dneScopeID),
		},
		{
			"should fail to ADD owners, validate add failure",
			types.NewMsgAddScopeOwnerRequest(scopeID, []types.Party{{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER}}, []string{s.user1}),
			[]string{s.user1},
			fmt.Sprintf("party already exists with address %s and role %s", s.user1, types.PartyType_PARTY_TYPE_OWNER),
		},
		{
			"should successfully ADD owners",
			types.NewMsgAddScopeOwnerRequest(scopeID, []types.Party{{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER}}, []string{s.user1}),
			[]string{s.user1},
			"",
		},
		{
			"should fail to DELETE owners, msg validate basic failure",
			types.NewMsgDeleteScopeOwnerRequest(scopeID, []string{}, []string{s.user1, s.user2}),
			[]string{s.user1},
			"at least one owner address is required",
		},
		{
			"should fail to DELETE owners, validate add failure",
			types.NewMsgDeleteScopeOwnerRequest(dneScopeID, []string{s.user1}, []string{s.user1, s.user2}),
			[]string{s.user1},
			fmt.Sprintf("scope not found with id %s", dneScopeID),
		},
		{
			"should fail to DELETE owners, validate add failure",
			types.NewMsgDeleteScopeOwnerRequest(scopeID, []string{user3}, []string{s.user1, s.user2}),
			[]string{s.user1},
			fmt.Sprintf("address does not exist in scope owners: %s", user3),
		},
		{
			"should successfully DELETE owners",
			types.NewMsgDeleteScopeOwnerRequest(scopeID, []string{s.user2}, []string{s.user1, s.user2}),
			[]string{s.user1},
			"",
		},
	}

	for _, tc := range cases {
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := s.handler(s.ctx, tc.msg)
			if len(tc.errorMsg) > 0 {
				assert.EqualError(t, err, tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	s.T().Run("owner actually deleted and added", func(t *testing.T) {
		addrOriginator := "cosmos1rr4d0eu62pgt4edw38d2ev27798pfhdhm39zct"
		addrServicer := "cosmos1a7mmtar5ke5fxk5gn00dlag0zfmdkmhapmugk7"
		scopeA := types.Scope{
			ScopeId:           types.ScopeMetadataAddress(uuid.New()),
			SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
			DataAccess:        []string{addrOriginator, addrServicer},
			ValueOwnerAddress: addrServicer,
			Owners: []types.Party{
				{
					Address: addrOriginator,
					Role:    types.PartyType_PARTY_TYPE_ORIGINATOR,
				},
				{
					Address: addrServicer,
					Role:    types.PartyType_PARTY_TYPE_SERVICER,
				},
			},
		}

		scopeSpecA := types.ScopeSpecification{
			SpecificationId: scopeA.SpecificationId,
			Description: &types.Description{
				Name:        "com.figure.origination.loan",
				Description: "Figure loan origination",
			},
			OwnerAddresses:  []string{"cosmos1q8n4v4m0hm8v0a7n697nwtpzhfsz3f4d40lnsu"},
			PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_ORIGINATOR},
			ContractSpecIds: nil,
		}

		s.app.MetadataKeeper.SetScope(s.ctx, scopeA)
		s.app.MetadataKeeper.SetScopeSpecification(s.ctx, scopeSpecA)

		msgDel := types.NewMsgDeleteScopeOwnerRequest(
			scopeA.ScopeId,
			[]string{addrServicer},
			[]string{addrOriginator, addrServicer},
		)

		_, errDel := s.handler(s.ctx, msgDel)
		require.NoError(t, errDel, "Failed to make DeleteScopeOwnerRequest call")

		scopeB, foundB := s.app.MetadataKeeper.GetScope(s.ctx, scopeA.ScopeId)
		require.Truef(t, foundB, "Scope %s not found after DeleteScopeOwnerRequest call.", scopeA.ScopeId)

		assert.Equal(t, scopeA.ScopeId, scopeB.ScopeId, "del ScopeId")
		assert.Equal(t, scopeA.SpecificationId, scopeB.SpecificationId, "del SpecificationId")
		assert.Equal(t, scopeA.DataAccess, scopeB.DataAccess, "del DataAccess")
		assert.Equal(t, scopeA.ValueOwnerAddress, scopeB.ValueOwnerAddress, "del ValueOwnerAddress")
		assert.Equal(t, scopeA.Owners[0:1], scopeB.Owners, "del Owners")

		// Stop test if it's already failed.
		if t.Failed() {
			t.FailNow()
		}

		msgAdd := types.NewMsgAddScopeOwnerRequest(
			scopeA.ScopeId,
			[]types.Party{{addrServicer, types.PartyType_PARTY_TYPE_SERVICER}},
			[]string{addrOriginator},
		)

		_, errAdd := s.handler(s.ctx, msgAdd)
		require.NoError(t, errAdd, "Failed to make DeleteScopeOwnerRequest call")

		scopeC, foundC := s.app.MetadataKeeper.GetScope(s.ctx, scopeA.ScopeId)
		require.Truef(t, foundC, "Scope %s not found after AddScopeOwnerRequest call.", scopeA.ScopeId)

		assert.Equal(t, scopeA.ScopeId, scopeC.ScopeId, "add ScopeId")
		assert.Equal(t, scopeA.SpecificationId, scopeC.SpecificationId, "add SpecificationId")
		assert.Equal(t, scopeA.DataAccess, scopeC.DataAccess, "add DataAccess")
		assert.Equal(t, scopeA.ValueOwnerAddress, scopeC.ValueOwnerAddress, "add ValueOwnerAddress")
		assert.Equal(t, scopeA.Owners, scopeC.Owners, "add Owners")
	})
}

func (s MetadataHandlerTestSuite) TestAddAndDeleteScopeDataAccess() {
	scopeSpecID := types.ScopeSpecMetadataAddress(uuid.New())
	scopeSpec := types.NewScopeSpecification(scopeSpecID, nil, []string{s.user1}, []types.PartyType{types.PartyType_PARTY_TYPE_OWNER}, []types.MetadataAddress{})
	scopeID := types.ScopeMetadataAddress(uuid.New())
	scope := types.NewScope(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, "")
	dneScopeID := types.ScopeMetadataAddress(uuid.New())
	user3 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()

	cases := []struct {
		name     string
		msg      sdk.Msg
		signers  []string
		errorMsg string
	}{
		{
			"setup test with new scope specification",
			types.NewMsgWriteScopeSpecificationRequest(*scopeSpec, []string{s.user1}),
			[]string{s.user1},
			"",
		},
		{
			"setup test with new scope",
			types.NewMsgWriteScopeRequest(*scope, []string{s.user1}),
			[]string{s.user1},
			"",
		},
		{
			"should fail to ADD address to data access, msg validate basic failure",
			types.NewMsgAddScopeDataAccessRequest(scopeID, []string{}, []string{s.user1}),
			[]string{s.user1},
			"data access list cannot be empty",
		},
		{
			"should fail to ADD address to data access, validate add failure",
			types.NewMsgAddScopeDataAccessRequest(dneScopeID, []string{s.user1}, []string{s.user1}),
			[]string{s.user1},
			fmt.Sprintf("scope not found with id %s", dneScopeID),
		},
		{
			"should fail to ADD address to data access, validate add failure",
			types.NewMsgAddScopeDataAccessRequest(scopeID, []string{s.user1}, []string{s.user1}),
			[]string{s.user1},
			fmt.Sprintf("address already exists for data access %s", s.user1),
		},
		{
			"should successfully ADD address to data access",
			types.NewMsgAddScopeDataAccessRequest(scopeID, []string{s.user2}, []string{s.user1}),
			[]string{s.user1},
			"",
		},
		{
			"should fail to DELETE address from data access, msg validate basic failure",
			types.NewMsgDeleteScopeDataAccessRequest(scopeID, []string{}, []string{s.user1}),
			[]string{s.user1},
			"data access list cannot be empty",
		},
		{
			"should fail to DELETE address from data access, validate add failure",
			types.NewMsgDeleteScopeDataAccessRequest(dneScopeID, []string{s.user1}, []string{s.user1}),
			[]string{s.user1},
			fmt.Sprintf("scope not found with id %s", dneScopeID),
		},
		{
			"should fail to DELETE address from data access, validate add failure",
			types.NewMsgDeleteScopeDataAccessRequest(scopeID, []string{user3}, []string{s.user1}),
			[]string{s.user1},
			fmt.Sprintf("address does not exist in scope data access: %s", user3),
		},
		{
			"should successfully DELETE address from data access",
			types.NewMsgDeleteScopeDataAccessRequest(scopeID, []string{s.user2}, []string{s.user1}),
			[]string{s.user1},
			"",
		},
	}

	for _, tc := range cases {
		s.T().Run(tc.name, func(t *testing.T) {
			_, err := s.handler(s.ctx, tc.msg)
			if len(tc.errorMsg) > 0 {
				assert.EqualError(t, err, tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	s.T().Run("data access actually deleted and added", func(t *testing.T) {
		addrOriginator := "cosmos1rr4d0eu62pgt4edw38d2ev27798pfhdhm39zct"
		addrServicer := "cosmos1a7mmtar5ke5fxk5gn00dlag0zfmdkmhapmugk7"
		scopeA := types.Scope{
			ScopeId:           types.ScopeMetadataAddress(uuid.New()),
			SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
			DataAccess:        []string{addrOriginator, addrServicer},
			ValueOwnerAddress: addrServicer,
			Owners: []types.Party{
				{
					Address: addrOriginator,
					Role:    types.PartyType_PARTY_TYPE_ORIGINATOR,
				},
			},
		}

		scopeSpecA := types.ScopeSpecification{
			SpecificationId: scopeA.SpecificationId,
			Description: &types.Description{
				Name:        "com.figure.origination.loan",
				Description: "Figure loan origination",
			},
			OwnerAddresses:  []string{"cosmos1q8n4v4m0hm8v0a7n697nwtpzhfsz3f4d40lnsu"},
			PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_ORIGINATOR},
			ContractSpecIds: nil,
		}

		s.app.MetadataKeeper.SetScope(s.ctx, scopeA)
		s.app.MetadataKeeper.SetScopeSpecification(s.ctx, scopeSpecA)

		msgDel := types.NewMsgDeleteScopeDataAccessRequest(
			scopeA.ScopeId,
			[]string{addrServicer},
			[]string{addrOriginator},
		)

		_, errDel := s.handler(s.ctx, msgDel)
		require.NoError(t, errDel, "Failed to make DeleteScopeDataAccessRequest call")

		scopeB, foundB := s.app.MetadataKeeper.GetScope(s.ctx, scopeA.ScopeId)
		require.Truef(t, foundB, "Scope %s not found after DeleteScopeOwnerRequest call.", scopeA.ScopeId)

		assert.Equal(t, scopeA.ScopeId, scopeB.ScopeId, "del ScopeId")
		assert.Equal(t, scopeA.SpecificationId, scopeB.SpecificationId, "del SpecificationId")
		assert.Equal(t, scopeA.DataAccess[0:1], scopeB.DataAccess, "del DataAccess")
		assert.Equal(t, scopeA.ValueOwnerAddress, scopeB.ValueOwnerAddress, "del ValueOwnerAddress")
		assert.Equal(t, scopeA.Owners, scopeB.Owners, "del Owners")

		// Stop test if it's already failed.
		if t.Failed() {
			t.FailNow()
		}

		msgAdd := types.NewMsgAddScopeDataAccessRequest(
			scopeA.ScopeId,
			[]string{addrServicer},
			[]string{addrOriginator},
		)

		_, errAdd := s.handler(s.ctx, msgAdd)
		require.NoError(t, errAdd, "Failed to make AddScopeDataAccessRequest call")

		scopeC, foundC := s.app.MetadataKeeper.GetScope(s.ctx, scopeA.ScopeId)
		require.Truef(t, foundC, "Scope %s not found after AddScopeOwnerRequest call.", scopeA.ScopeId)

		assert.Equal(t, scopeA.ScopeId, scopeC.ScopeId, "add ScopeId")
		assert.Equal(t, scopeA.SpecificationId, scopeC.SpecificationId, "add SpecificationId")
		assert.Equal(t, scopeA.DataAccess, scopeC.DataAccess, "add DataAccess")
		assert.Equal(t, scopeA.ValueOwnerAddress, scopeC.ValueOwnerAddress, "add ValueOwnerAddress")
		assert.Equal(t, scopeA.Owners, scopeC.Owners, "add Owners")
	})
}

func (s MetadataHandlerTestSuite) TestIssue412WriteScopeOptionalField() {
	ownerAddress := "cosmos1vz99nyd2er8myeugsr4xm5duwhulhp5ae4dvpa"
	specIDStr := "scopespec1qjkyp28sldx5r9ueaxqc5adrc5wszy6nsh"
	specUUIDStr := "ac40a8f0-fb4d-4197-99e9-818a75a3c51d"
	specID, specIDErr := types.MetadataAddressFromBech32(specIDStr)
	require.NoError(s.T(), specIDErr, "converting scopeIDStr to a metadata address")

	s.T().Run("Ensure ID and UUID strings match", func(t *testing.T) {
		specIDFromID, e1 := types.MetadataAddressFromBech32(specIDStr)
		require.NoError(t, e1, "specIDFromIDStr")
		specUUIDFromID, e2 := specIDFromID.ScopeSpecUUID()
		require.NoError(t, e2, "specUUIDActualStr")
		specUUIDStrActual := specUUIDFromID.String()
		assert.Equal(t, specUUIDStr, specUUIDStrActual, "UUID strings")

		specIDFFromUUID := types.ScopeSpecMetadataAddress(uuid.MustParse(specUUIDStr))
		specIDStrActual := specIDFFromUUID.String()
		assert.Equal(t, specIDStr, specIDStrActual, "ID strings")

		assert.Equal(t, specIDFromID, specIDFFromUUID, "scope spec ids")
	})

	s.T().Run("Setting scope spec with just a spec ID", func(t *testing.T) {
		msg := types.MsgWriteScopeSpecificationRequest{
			Specification: types.ScopeSpecification{
				SpecificationId: specID,
				Description: &types.Description{
					Name:        "io.p8e.contracts.examplekotlin.helloWorld",
					Description: "A generic scope that allows for a lot of example hello world contracts.",
				},
				OwnerAddresses:  []string{ownerAddress},
				PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				ContractSpecIds: nil,
			},
			Signers:  []string{ownerAddress},
			SpecUuid: "",
		}
		res, err := s.handler(s.ctx, &msg)
		assert.NoError(t, err)
		assert.NotNil(t, 0, res)
	})

	s.T().Run("Setting scope spec with just a UUID", func(t *testing.T) {
		msg := types.MsgWriteScopeSpecificationRequest{
			Specification: types.ScopeSpecification{
				SpecificationId: nil,
				Description: &types.Description{
					Name:        "io.p8e.contracts.examplekotlin.helloWorld",
					Description: "A generic scope that allows for a lot of example hello world contracts.",
				},
				OwnerAddresses:  []string{ownerAddress},
				PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				ContractSpecIds: nil,
			},
			Signers:  []string{ownerAddress},
			SpecUuid: specUUIDStr,
		}
		res, err := s.handler(s.ctx, &msg)
		assert.NoError(t, err)
		assert.NotNil(t, 0, res)
	})

	s.T().Run("Setting scope spec with matching ID and UUID", func(t *testing.T) {
		msg := types.MsgWriteScopeSpecificationRequest{
			Specification: types.ScopeSpecification{
				SpecificationId: specID,
				Description: &types.Description{
					Name:        "io.p8e.contracts.examplekotlin.helloWorld",
					Description: "A generic scope that allows for a lot of example hello world contracts.",
				},
				OwnerAddresses:  []string{ownerAddress},
				PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				ContractSpecIds: nil,
			},
			Signers:  []string{ownerAddress},
			SpecUuid: specUUIDStr,
		}
		res, err := s.handler(s.ctx, &msg)
		assert.NoError(t, err)
		assert.NotNil(t, 0, res)
	})
}
