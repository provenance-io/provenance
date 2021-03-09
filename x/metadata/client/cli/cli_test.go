package cli_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil"

	"github.com/provenance-io/provenance/x/metadata/client/cli"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     testnet.Config
	testnet *testnet.Network

	asJson string
	asText string

	accountAddr sdk.AccAddress
	accountKey  *secp256k1.PrivKey

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress

	scope     metadatatypes.Scope
	scopeUUID uuid.UUID
	scopeID   metadatatypes.MetadataAddress

	session     metadatatypes.Session
	sessionUUID uuid.UUID
	sessionID   metadatatypes.MetadataAddress

	record     metadatatypes.Record
	recordName string
	recordID   metadatatypes.MetadataAddress

	scopeSpec     metadatatypes.ScopeSpecification
	scopeSpecUUID uuid.UUID
	scopeSpecID   metadatatypes.MetadataAddress

	contractSpec     metadatatypes.ContractSpecification
	contractSpecUUID uuid.UUID
	contractSpecID   metadatatypes.MetadataAddress

	recordSpec   metadatatypes.RecordSpecification
	recordSpecID metadatatypes.MetadataAddress
}

func ownerPartyList(addresses ...string) []metadatatypes.Party {
	retval := make([]metadatatypes.Party, len(addresses))
	for i, addr := range addresses {
		retval[i] = metadatatypes.Party{Address: addr, Role: metadatatypes.PartyType_PARTY_TYPE_OWNER}
	}
	return retval
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.accountKey = secp256k1.GenPrivKeyFromSecret([]byte("acc2"))
	addr, err := sdk.AccAddressFromHex(s.accountKey.PubKey().Address().String())
	s.Require().NoError(err)
	s.accountAddr = addr
	s.T().Log("setting up integration test suite")

	cfg := testutil.DefaultTestNetworkConfig()

	s.asJson = fmt.Sprintf("--%s=json", tmcli.OutputFlag)
	s.asText = fmt.Sprintf("--%s=text", tmcli.OutputFlag)

	genesisState := cfg.GenesisState
	cfg.NumValidators = 1

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	s.scopeUUID = uuid.New()
	s.sessionUUID = uuid.New()
	s.recordName = "recordname"
	s.scopeSpecUUID = uuid.New()
	s.contractSpecUUID = uuid.New()

	s.scopeID = metadatatypes.ScopeMetadataAddress(s.scopeUUID)
	s.sessionID = metadatatypes.SessionMetadataAddress(s.scopeUUID, s.sessionUUID)
	s.recordID = metadatatypes.RecordMetadataAddress(s.scopeUUID, s.recordName)
	s.scopeSpecID = metadatatypes.ScopeSpecMetadataAddress(s.scopeSpecUUID)
	s.contractSpecID = metadatatypes.ContractSpecMetadataAddress(s.contractSpecUUID)
	s.recordSpecID = metadatatypes.RecordSpecMetadataAddress(s.contractSpecUUID, s.recordName)

	s.scope = *metadatatypes.NewScope(
		s.scopeID,
		s.scopeSpecID,
		ownerPartyList(s.user1),
		[]string{s.user1},
		s.user1,
	)

	s.session = *metadatatypes.NewSession(
		"unit test session",
		s.sessionID,
		s.contractSpecID,
		ownerPartyList(s.user1),
		metadatatypes.AuditFields{
			CreatedDate: time.Time{},
			CreatedBy:   s.user1,
			UpdatedDate: time.Time{},
			UpdatedBy:   "",
			Version:     0,
			Message:     "unit testing",
		},
	)

	s.record = *metadatatypes.NewRecord(
		s.recordName,
		s.sessionID,
		*metadatatypes.NewProcess(
			"record process",
			&metadatatypes.Process_Hash{Hash: "notarealprocesshash"},
			"myMethod",
		),
		[]metadatatypes.RecordInput{
			*metadatatypes.NewRecordInput(
				"inputname",
				&metadatatypes.RecordInput_Hash{Hash: "notarealrecordinputhash"},
				"inputtypename",
				metadatatypes.RecordInputStatus_Record,
			),
		},
		[]metadatatypes.RecordOutput{
			*metadatatypes.NewRecordOutput(
				"notarealrecordoutputhash",
				metadatatypes.ResultStatus_RESULT_STATUS_PASS,
			),
		},
	)

	s.scopeSpec = *metadatatypes.NewScopeSpecification(
		s.scopeSpecID,
		nil,
		[]string{s.user1},
		[]metadatatypes.PartyType{metadatatypes.PartyType_PARTY_TYPE_OWNER},
		[]metadatatypes.MetadataAddress{s.contractSpecID},
	)

	s.contractSpec = *metadatatypes.NewContractSpecification(
		s.contractSpecID,
		nil,
		[]string{s.user1},
		[]metadatatypes.PartyType{metadatatypes.PartyType_PARTY_TYPE_OWNER},
		metadatatypes.NewContractSpecificationSourceHash("notreallyasourcehash"),
		"contractclassname",
	)

	s.recordSpec = *metadatatypes.NewRecordSpecification(
		s.recordSpecID,
		s.recordName,
		[]*metadatatypes.InputSpecification{
			metadatatypes.NewInputSpecification(
				"inputname",
				"inputtypename",
				metadatatypes.NewInputSpecificationSourceHash("alsonotreallyasourcehash"),
			),
		},
		"recordtypename",
		metadatatypes.DefinitionType_DEFINITION_TYPE_RECORD,
		[]metadatatypes.PartyType{metadatatypes.PartyType_PARTY_TYPE_OWNER},
	)

	var metadataData metadatatypes.GenesisState
	s.Require().NoError(cfg.Codec.UnmarshalJSON(genesisState[metadatatypes.ModuleName], &metadataData))
	metadataData.Scopes = append(metadataData.Scopes, s.scope)
	metadataData.Sessions = append(metadataData.Sessions, s.session)
	metadataData.Records = append(metadataData.Records, s.record)
	metadataData.ScopeSpecifications = append(metadataData.ScopeSpecifications, s.scopeSpec)
	metadataData.ContractSpecifications = append(metadataData.ContractSpecifications, s.contractSpec)
	metadataData.RecordSpecifications = append(metadataData.RecordSpecifications, s.recordSpec)
	metadataDataBz, err := cfg.Codec.MarshalJSON(&metadataData)
	s.Require().NoError(err)
	genesisState[metadatatypes.ModuleName] = metadataDataBz

	// adding to auth genesis does not work due to cosmos sdk overwritting it
	var authData authtypes.GenesisState
	s.Require().NoError(cfg.Codec.UnmarshalJSON(genesisState[authtypes.ModuleName], &authData))
	genAccount, err := codectypes.NewAnyWithValue(authtypes.NewBaseAccount(s.accountAddr, s.accountKey.PubKey(), 1, 1))
	s.Require().NoError(err)
	authData.Accounts = append(authData.Accounts, genAccount)

	user1Account, err := codectypes.NewAnyWithValue(authtypes.NewBaseAccount(s.user1Addr, s.pubkey1, 2, 2))
	s.Require().NoError(err)
	authData.Accounts = append(authData.Accounts, user1Account)

	user2Account, err := codectypes.NewAnyWithValue(authtypes.NewBaseAccount(s.user2Addr, s.pubkey2, 3, 3))
	s.Require().NoError(err)
	authData.Accounts = append(authData.Accounts, user2Account)

	authDataBz, err := cfg.Codec.MarshalJSON(&authData)
	s.Require().NoError(err)
	genesisState[authtypes.ModuleName] = authDataBz

	cfg.GenesisState = genesisState

	s.cfg = cfg

	// TODO: This is overwritting some of our genesis states https://github.com/provenance-io/provenance/issues/81
	s.testnet = testnet.New(s.T(), cfg)

	_, err = s.testnet.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.testnet.WaitForNextBlock()
	s.T().Log("tearing down integration test suite")
	s.testnet.Cleanup()
}

// ---------- query cmd tests ----------

type queryCmdTestCase struct {
	name           string
	args           []string
	expectedError  string
	expectedOutput string
}

func runQueryCmdTestCases(s *IntegrationTestSuite, cmd *cobra.Command, testCases []queryCmdTestCase) {
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if len(tc.expectedError) > 0 {
				actualError := ""
				if err != nil {
					actualError = err.Error()
				}
				require.Equal(t, tc.expectedError, actualError, "expected error")
			} else {
				require.NoError(t, err, "unexpected error")
			}
			if err == nil {
				require.Equal(t, tc.expectedOutput, strings.TrimSpace(out.String()), "expected output")
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetMetadataParamsCmd() {
	cmd := cli.GetMetadataParamsCmd()

	testCases := []queryCmdTestCase{
		{
			"get params as json output",
			[]string{s.asJson},
			"",
			"{}",
		},
		{
			"get params as text output",
			[]string{s.asText},
			"",
			"{}",
		},
		{
			"get params - invalid args",
			[]string{"bad-arg"},
			"unknown command \"bad-arg\" for \"params\"",
			"{}",
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

// TODO: GetMetadataByIDCmd
func (s *IntegrationTestSuite) TestGetMetadataByIDCmd() {
	cmd := cli.GetMetadataByIDCmd()

	scopeAsJson := fmt.Sprintf("{\"scope_id\":\"%s\",\"specification_id\":\"%s\",\"owners\":[{\"address\":\"%s\",\"role\":\"%s\"}],\"data_access\":[\"%s\"],\"value_owner_address\":\"%s\"}",
		s.scope.ScopeId,
		s.scope.SpecificationId.String(),
		s.scope.Owners[0].Address,
		s.scope.Owners[0].Role.String(),
		s.scope.DataAccess[0],
		s.scope.ValueOwnerAddress,
	)

	sessionAsJson := fmt.Sprintf("{\"session_id\":\"%s\",\"specification_id\":\"%s\",\"parties\":[{\"address\":\"%s\",\"role\":\"PARTY_TYPE_OWNER\"}],\"name\":\"unit test session\",\"audit\":{\"created_date\":\"0001-01-01T00:00:00Z\",\"created_by\":\"%s\",\"updated_date\":\"0001-01-01T00:00:00Z\",\"updated_by\":\"\",\"version\":0,\"message\":\"unit testing\"}}",
		s.sessionID,
		s.contractSpecID,
		s.user1,
		s.user1,
	)
	sessionAsText := fmt.Sprintf(`audit:
  created_by: %s
  created_date: "0001-01-01T00:00:00Z"
  message: unit testing
  updated_by: ""
  updated_date: "0001-01-01T00:00:00Z"
  version: 0
name: unit test session
parties:
- address: %s
  role: PARTY_TYPE_OWNER
session_id: %s
specification_id: %s`,
		s.user1,
		s.user1,
		s.sessionID,
		s.contractSpecID,
	)

	recordAsJson := fmt.Sprintf("{\"name\":\"recordname\",\"session_id\":\"%s\",\"process\":{\"hash\":\"notarealprocesshash\",\"name\":\"record process\",\"method\":\"myMethod\"},\"inputs\":[{\"name\":\"inputname\",\"hash\":\"notarealrecordinputhash\",\"type_name\":\"inputtypename\",\"status\":\"RECORD_INPUT_STATUS_RECORD\"}],\"outputs\":[{\"hash\":\"notarealrecordoutputhash\",\"status\":\"RESULT_STATUS_PASS\"}]}",
		s.sessionID,
	)
	recordAsText := fmt.Sprintf(`inputs:
- hash: notarealrecordinputhash
  name: inputname
  status: RECORD_INPUT_STATUS_RECORD
  type_name: inputtypename
name: recordname
outputs:
- hash: notarealrecordoutputhash
  status: RESULT_STATUS_PASS
process:
  hash: notarealprocesshash
  method: myMethod
  name: record process
session_id: %s`,
		s.sessionID,
	)

	fullScopeAsJson := fmt.Sprintf("{\"scope\":%s,\"sessions\":[%s],\"records\":[%s],\"scope_uuid\":\"%s\"}",
		scopeAsJson, sessionAsJson, recordAsJson, s.scopeUUID)
	fullScopeAsText := fmt.Sprintf(`records:
- inputs:
  - hash: notarealrecordinputhash
    name: inputname
    status: RECORD_INPUT_STATUS_RECORD
    type_name: inputtypename
  name: recordname
  outputs:
  - hash: notarealrecordoutputhash
    status: RESULT_STATUS_PASS
  process:
    hash: notarealprocesshash
    method: myMethod
    name: record process
  session_id: %s
scope:
  data_access:
  - %s
  owners:
  - address: %s
    role: PARTY_TYPE_OWNER
  scope_id: %s
  specification_id: %s
  value_owner_address: %s
scope_uuid: %s
sessions:
- audit:
    created_by: %s
    created_date: "0001-01-01T00:00:00Z"
    message: unit testing
    updated_by: ""
    updated_date: "0001-01-01T00:00:00Z"
    version: 0
  name: unit test session
  parties:
  - address: %s
    role: PARTY_TYPE_OWNER
  session_id: %s
  specification_id: %s`,
		s.sessionID,
		s.user1,
		s.user1,
		s.scopeID,
		s.scopeSpecID,
		s.user1,
		s.scopeUUID,
		s.user1,
		s.user1,
		s.sessionID,
		s.contractSpecID,
	)

	scopeSpecAsJson := fmt.Sprintf("{\"specification_id\":\"%s\",\"description\":null,\"owner_addresses\":[\"%s\"],\"parties_involved\":[\"PARTY_TYPE_OWNER\"],\"contract_spec_ids\":[\"%s\"]}",
		s.scopeSpecID,
		s.user1,
		s.contractSpecID,
	)
	scopeSpecAsText := fmt.Sprintf(`contract_spec_ids:
- %s
description: null
owner_addresses:
- %s
parties_involved:
- PARTY_TYPE_OWNER
specification_id: %s`,
		s.contractSpecID,
		s.user1,
		s.scopeSpecID,
	)

	contractSpecAsJson := fmt.Sprintf("{\"specification_id\":\"%s\",\"description\":null,\"owner_addresses\":[\"%s\"],\"parties_involved\":[\"PARTY_TYPE_OWNER\"],\"hash\":\"notreallyasourcehash\",\"class_name\":\"contractclassname\"}",
		s.contractSpecID,
		s.user1,
	)
	contractSpecAsText := fmt.Sprintf(`class_name: contractclassname
description: null
hash: notreallyasourcehash
owner_addresses:
- %s
parties_involved:
- PARTY_TYPE_OWNER
specification_id: %s`,
		s.user1,
		s.contractSpecID,
	)

	recordSpecAsJson := fmt.Sprintf("{\"specification_id\":\"%s\",\"name\":\"recordname\",\"inputs\":[{\"name\":\"inputname\",\"type_name\":\"inputtypename\",\"hash\":\"alsonotreallyasourcehash\"}],\"type_name\":\"recordtypename\",\"result_type\":\"DEFINITION_TYPE_RECORD\",\"responsible_parties\":[\"PARTY_TYPE_OWNER\"]}",
		s.recordSpecID,
	)
	recordSpecAsText := fmt.Sprintf(`inputs:
- hash: alsonotreallyasourcehash
  name: inputname
  type_name: inputtypename
name: recordname
responsible_parties:
- PARTY_TYPE_OWNER
result_type: DEFINITION_TYPE_RECORD
specification_id: %s
type_name: recordtypename`,
		s.recordSpecID,
	)

	testCases := []queryCmdTestCase{
		{
			"get metadata by id - scope id as json",
			[]string{s.scopeID.String(), s.asJson},
			"",
			fullScopeAsJson,
		},
		{
			"get metadata by id - scope id as text",
			[]string{s.scopeID.String(), s.asText},
			"",
			fullScopeAsText,
		},
		{
			"get metadata by id - scope id does not exist",
			[]string{"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
			"rpc error: code = NotFound desc = scope uuid 91978ba2-5f35-459a-86a7-feca1b0512e0 not found: key not found",
			"",
		},
		{
			"get metadata by id - session id as json",
			[]string{s.sessionID.String(), s.asJson},
			"",
			sessionAsJson,
		},
		{
			"get metadata by id - session id as text",
			[]string{s.sessionID.String(), s.asText},
			"",
			sessionAsText,
		},
		{
			"get metadata by id - session id does not exist",
			[]string{"session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"},
			"rpc error: code = NotFound desc = session id session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr not found: key not found",
			"",
		},
		{
			"get metadata by id - record id as json",
			[]string{s.recordID.String(), s.asJson},
			"",
			recordAsJson,
		},
		{
			"get metadata by id - record id as text",
			[]string{s.recordID.String(), s.asText},
			"",
			recordAsText,
		},
		{
			"get metadata by id - record id does not exist",
			[]string{"record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3"},
			"rpc error: code = NotFound desc = scope uuid 91978ba2-5f35-459a-86a7-feca1b0512e0 not found: key not found",
			"",
		},
		{
			"get metadata by id - scope spec id as json",
			[]string{s.scopeSpecID.String(), s.asJson},
			"",
			scopeSpecAsJson,
		},
		{
			"get metadata by id - scope spec id as text",
			[]string{s.scopeSpecID.String(), s.asText},
			"",
			scopeSpecAsText,
		},
		{
			"get metadata by id - scope spec id does not exist",
			[]string{"scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"},
			"rpc error: code = NotFound desc = scope specification uuid dc83ea70-eacd-40fe-9adf-1cf6148bf8a2 not found: key not found",
			"",
		},
		{
			"get metadata by id - contract spec id as json",
			[]string{s.contractSpecID.String(), s.asJson},
			"",
			contractSpecAsJson,
		},
		{
			"get metadata by id - contract spec id as text",
			[]string{s.contractSpecID.String(), s.asText},
			"",
			contractSpecAsText,
		},
		{
			"get metadata by id - contract spec id does not exist",
			[]string{"contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"},
			"rpc error: code = NotFound desc = contract specification with id contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn (uuid def6bc0a-c9dd-4874-948f-5206e6060a84) not found: key not found",
			"",
		},
		{
			"get metadata by id - record spec id as json",
			[]string{s.recordSpecID.String(), s.asJson},
			"",
			recordSpecAsJson,
		},
		{
			"get metadata by id - record spec id as text",
			[]string{s.recordSpecID.String(), s.asText},
			"",
			recordSpecAsText,
		},
		{
			"get metadata by id - record spec id does not exist",
			[]string{"recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44"},
			"rpc error: code = NotFound desc = record specification not found for id recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44: key not found",
			"",
		},
		{
			"get metadata by id - bad prefix",
			[]string{"foo1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
			"decoding bech32 failed: checksum failed. Expected kzwk8c, got xlkwel.",
			"",
		},
		{
			"get metadata by id - no args",
			[]string{},
			"accepts 1 arg(s), received 0",
			"",
		},
		{
			"get metadata by id - two args",
			[]string{"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel", "scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"},
			"accepts 1 arg(s), received 2",
			"",
		},
		{
			"get metadata by id - uuid",
			[]string{"91978ba2-5f35-459a-86a7-feca1b0512e0"},
			"decoding bech32 failed: invalid index of 1",
			"",
		},
		{
			"get metadata by id - bad arg",
			[]string{"not-an-id"},
			"decoding bech32 failed: invalid index of 1",
			"",
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationTestSuite) TestGetMetadataScopeCmd() {
	cmd := cli.GetMetadataScopeCmd()

	scopeAsJson := fmt.Sprintf("{\"scope_id\":\"%s\",\"specification_id\":\"%s\",\"owners\":[{\"address\":\"%s\",\"role\":\"%s\"}],\"data_access\":[\"%s\"],\"value_owner_address\":\"%s\"}",
		s.scope.ScopeId,
		s.scope.SpecificationId.String(),
		s.scope.Owners[0].Address,
		s.scope.Owners[0].Role.String(),
		s.scope.DataAccess[0],
		s.scope.ValueOwnerAddress,
	)
	scopeAsText := fmt.Sprintf(`data_access:
- %s
owners:
- address: %s
  role: %s
scope_id: %s
specification_id: %s
value_owner_address: %s`,
		s.scope.DataAccess[0],
		s.scope.Owners[0].Address,
		s.scope.Owners[0].Role.String(),
		s.scope.ScopeId,
		s.scope.SpecificationId.String(),
		s.scope.ValueOwnerAddress,
	)

	testCases := []queryCmdTestCase{
		{
			"get scope by metadata scope id as json output",
			[]string{s.scopeID.String(), s.asJson},
			"",
			scopeAsJson,
		},
		{
			"get scope by metadata scope id as text output",
			[]string{s.scopeID.String(), s.asText},
			"",
			scopeAsText,
		},
		{
			"get scope by uuid as json output",
			[]string{s.scopeUUID.String(), s.asJson},
			"",
			scopeAsJson,
		},
		{
			"get scope by uuid as text output",
			[]string{s.scopeUUID.String(), s.asText},
			"",
			scopeAsText,
		},
		{
			"get scope by metadata session id as json output",
			[]string{s.sessionID.String(), s.asJson},
			"",
			scopeAsJson,
		},
		{
			"get scope by metadata session id as text output",
			[]string{s.sessionID.String(), s.asText},
			"",
			scopeAsText,
		},
		{
			"get scope by metadata record id as json output",
			[]string{s.recordID.String(), s.asJson},
			"",
			scopeAsJson,
		},
		{
			"get scope by metadata record id as text output",
			[]string{s.recordID.String(), s.asText},
			"",
			scopeAsText,
		},
		{
			"get scope by metadata id - does not exist",
			[]string{"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel", s.asText},
			"rpc error: code = NotFound desc = scope uuid 91978ba2-5f35-459a-86a7-feca1b0512e0 not found: key not found",
			"",
		},
		{
			"get scope by uuid - does not exist",
			[]string{"91978ba2-5f35-459a-86a7-feca1b0512e0", s.asText},
			"rpc error: code = NotFound desc = scope uuid 91978ba2-5f35-459a-86a7-feca1b0512e0 not found: key not found",
			"",
		},
		{
			"get scope bad arg",
			[]string{"not-a-valid-arg", s.asText},
			"argument not-a-valid-arg is neither a metadata address (decoding bech32 failed: invalid index of 1) nor uuid (invalid UUID length: 15)",
			"",
		},
		{
			"get scope no args",
			[]string{},
			"accepts 1 arg(s), received 0",
			"",
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

// TODO: GetMetadataFullScopeCmd
func (s *IntegrationTestSuite) TestGetMetadataFullScopeCmd() {
	cmd := cli.GetMetadataFullScopeCmd()

	testCases := []queryCmdTestCase{
		// TODO: scope id
		// TODO: scope uuid
		// TODO: session id
		// TODO: session uuid
		// TODO: record id
		// TODO: record uuid
		// TODO: entry does not exist
		// TODO: bad prefix
		// TODO: bad arg
		// TODO: two args
		// TODO: no args
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

// TODO: GetMetadataSessionCmd
func (s *IntegrationTestSuite) TestGetMetadataSessionCmd() {
	cmd := cli.GetMetadataSessionCmd()

	testCases := []queryCmdTestCase{
		// TODO: session id
		// TODO: scope id
		// TODO: scope uuid
		// TODO: scope uuid, session uuid
		// TODO: bad prefix
		// TODO: bad arg 1
		// TODO: uuid, bad arg 2
		// TODO: 3 args
		// TODO: no args
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

// TODO: GetMetadataRecordCmd
func (s *IntegrationTestSuite) TestGetMetadataRecordCmd() {
	cmd := cli.GetMetadataRecordCmd()

	testCases := []queryCmdTestCase{
		// TODO: record id
		// TODO: session id
		// TODO: scope id
		// TODO: scope uuid
		// TODO: scope uuid, record name
		// TODO: bad prefix
		// TODO: bad arg 1
		// TODO: uuid, whitespace arg 2 and 3
		// TODO: no args
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

// TODO: GetMetadataScopeSpecCmd
func (s *IntegrationTestSuite) TestGetMetadataScopeSpecCmd() {
	cmd := cli.GetMetadataScopeSpecCmd()

	testCases := []queryCmdTestCase{
		// TODO: scope spec id
		// TODO: scope spec uuid
		// TODO: bad prefix
		// TODO: bad arg
		// TODO: two args
		// TODO: no args
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

// TODO: GetMetadataContractSpecCmd
func (s *IntegrationTestSuite) TestGetMetadataContractSpecCmd() {
	cmd := cli.GetMetadataContractSpecCmd()

	testCases := []queryCmdTestCase{
		// TODO: contract spec id
		// TODO: contract spec uuid
		// TODO: record spec id
		// TODO: bad prefix
		// TODO: bad arg
		// TODO: two args
		// TODO: no args
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

// TODO: GetMetadataRecordSpecCmd
func (s *IntegrationTestSuite) TestGetMetadataRecordSpecCmd() {
	cmd := cli.GetMetadataRecordSpecCmd()

	testCases := []queryCmdTestCase{
		// TODO: rec spec id
		// TODO: contract spec id
		// TODO: contract spec uuid
		// TODO: contract spec uuid, name
		// TODO: bad prefix
		// TODO: bad arg 1
		// TODO: uuid, whitespace args 2 and 3
		// TODO: no args
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

// ---------- tx cmd tests ----------

func (s *IntegrationTestSuite) TestAddMetadataScopeCmd() {

	scopeUUID := uuid.New().String()

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"Should successfully add metadata scope",
			cli.AddMetadataScopeCmd(),
			[]string{
				scopeUUID,
				uuid.New().String(),
				s.user1,
				s.user1,
				s.user1,
				s.testnet.Validators[0].Address.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"Should fail to add metadata scope, incorrect scope uuid",
			cli.AddMetadataScopeCmd(),
			[]string{
				"not-a-uuid",
				uuid.New().String(),
				s.user1,
				s.user1,
				s.user1,
				s.testnet.Validators[0].Address.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"Should fail to add metadata scope, incorrect scope spec uuid",
			cli.AddMetadataScopeCmd(),
			[]string{
				uuid.New().String(),
				"not-a-uuid",
				s.user1,
				s.user1,
				s.user1,
				s.testnet.Validators[0].Address.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"Should fail to add metadata scope, incorrect scope spec uuid",
			cli.AddMetadataScopeCmd(),
			[]string{
				uuid.New().String(),
				"not-a-uuid",
				s.user1,
				s.user1,
				s.user1,
				s.testnet.Validators[0].Address.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"Should fail to add metadata scope, incorrect owner address format",
			cli.AddMetadataScopeCmd(),
			[]string{
				uuid.New().String(),
				uuid.New().String(),
				"incorrect,incorrect",
				s.user1,
				s.user1,
				s.testnet.Validators[0].Address.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"Should fail to add metadata scope, incorrect data access format",
			cli.AddMetadataScopeCmd(),
			[]string{
				uuid.New().String(),
				uuid.New().String(),
				s.user1,
				"incorrect,incorrect",
				s.user1,
				s.testnet.Validators[0].Address.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.user1),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"Should fail to add metadata scope, incorrect value owner address",
			cli.AddMetadataScopeCmd(),
			[]string{
				uuid.New().String(),
				uuid.New().String(),
				s.user1,
				s.user1,
				"incorrect",
				s.testnet.Validators[0].Address.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.user1),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestRemoveMetadataScopeCmd() {

	userId := s.testnet.Validators[0].Address.String()
	scopeUUID := uuid.New().String()

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"Should successfully add metadata scope for testing scope removal",
			cli.AddMetadataScopeCmd(),
			[]string{
				scopeUUID,
				uuid.New().String(),
				userId,
				userId,
				userId,
				userId,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"Should fail to remove metadata scope, invalid scopeid",
			cli.RemoveMetadataScopeCmd(),
			[]string{
				"not-valid",
				userId,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"Should fail to remove metadata scope, invalid userid",
			cli.RemoveMetadataScopeCmd(),
			[]string{
				scopeUUID,
				"not-valid",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"Should remove metadata scope",
			cli.RemoveMetadataScopeCmd(),
			[]string{
				scopeUUID,
				userId,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}
