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
	"github.com/provenance-io/provenance/x/metadata/types"
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

	scopeAsJson     string
	scopeAsText     string
	fullScopeAsJson string
	fullScopeAsText string

	session     metadatatypes.Session
	sessionUUID uuid.UUID
	sessionID   metadatatypes.MetadataAddress

	sessionAsJson string
	sessionAsText string

	record     metadatatypes.Record
	recordName string
	recordID   metadatatypes.MetadataAddress

	recordAsJson string
	recordAsText string

	scopeSpec     metadatatypes.ScopeSpecification
	scopeSpecUUID uuid.UUID
	scopeSpecID   metadatatypes.MetadataAddress

	scopeSpecAsJson string
	scopeSpecAsText string

	contractSpec     metadatatypes.ContractSpecification
	contractSpecUUID uuid.UUID
	contractSpecID   metadatatypes.MetadataAddress

	contractSpecAsJson string
	contractSpecAsText string

	recordSpec   metadatatypes.RecordSpecification
	recordSpecID metadatatypes.MetadataAddress

	recordSpecAsJson string
	recordSpecAsText string

	//os locator's
	objectLocator metadatatypes.ObjectStoreLocator
	ownerAddr     sdk.AccAddress
	uri           string

	objectLocator1 metadatatypes.ObjectStoreLocator
	ownerAddr1     sdk.AccAddress
	uri1           string
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
		s.user2,
	)

	s.session = *metadatatypes.NewSession(
		"unit test session",
		s.sessionID,
		s.contractSpecID,
		ownerPartyList(s.user1),
		&metadatatypes.AuditFields{
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

	s.scopeAsJson = fmt.Sprintf("{\"scope_id\":\"%s\",\"specification_id\":\"%s\",\"owners\":[{\"address\":\"%s\",\"role\":\"PARTY_TYPE_OWNER\"}],\"data_access\":[\"%s\"],\"value_owner_address\":\"%s\"}",
		s.scopeID,
		s.scopeSpecID,
		s.user1,
		s.user1,
		s.user2,
	)
	s.scopeAsText = fmt.Sprintf(`data_access:
- %s
owners:
- address: %s
  role: PARTY_TYPE_OWNER
scope_id: %s
specification_id: %s
value_owner_address: %s`,
		s.user1,
		s.user1,
		s.scopeID,
		s.scopeSpecID,
		s.user2,
	)

	s.sessionAsJson = fmt.Sprintf("{\"session_id\":\"%s\",\"specification_id\":\"%s\",\"parties\":[{\"address\":\"%s\",\"role\":\"PARTY_TYPE_OWNER\"}],\"name\":\"unit test session\",\"audit\":{\"created_date\":\"0001-01-01T00:00:00Z\",\"created_by\":\"%s\",\"updated_date\":\"0001-01-01T00:00:00Z\",\"updated_by\":\"\",\"version\":0,\"message\":\"unit testing\"}}",
		s.sessionID,
		s.contractSpecID,
		s.user1,
		s.user1,
	)
	s.sessionAsText = fmt.Sprintf(`audit:
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

	s.recordAsJson = fmt.Sprintf("{\"name\":\"recordname\",\"session_id\":\"%s\",\"process\":{\"hash\":\"notarealprocesshash\",\"name\":\"record process\",\"method\":\"myMethod\"},\"inputs\":[{\"name\":\"inputname\",\"hash\":\"notarealrecordinputhash\",\"type_name\":\"inputtypename\",\"status\":\"RECORD_INPUT_STATUS_RECORD\"}],\"outputs\":[{\"hash\":\"notarealrecordoutputhash\",\"status\":\"RESULT_STATUS_PASS\"}]}",
		s.sessionID,
	)
	s.recordAsText = fmt.Sprintf(`inputs:
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

	s.fullScopeAsJson = fmt.Sprintf("{\"scope\":%s,\"sessions\":[%s],\"records\":[%s],\"scope_uuid\":\"%s\"}",
		s.scopeAsJson,
		s.sessionAsJson,
		s.recordAsJson,
		s.scopeUUID,
	)
	s.fullScopeAsText = fmt.Sprintf(`records:
%s
scope:
%s
scope_uuid: %s
sessions:
%s`,
		yamlListEntry(s.recordAsText),
		indent(s.scopeAsText),
		s.scopeUUID,
		yamlListEntry(s.sessionAsText),
	)

	s.scopeSpecAsJson = fmt.Sprintf("{\"specification_id\":\"%s\",\"description\":null,\"owner_addresses\":[\"%s\"],\"parties_involved\":[\"PARTY_TYPE_OWNER\"],\"contract_spec_ids\":[\"%s\"]}",
		s.scopeSpecID,
		s.user1,
		s.contractSpecID,
	)
	s.scopeSpecAsText = fmt.Sprintf(`contract_spec_ids:
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

	s.contractSpecAsJson = fmt.Sprintf("{\"specification_id\":\"%s\",\"description\":null,\"owner_addresses\":[\"%s\"],\"parties_involved\":[\"PARTY_TYPE_OWNER\"],\"hash\":\"notreallyasourcehash\",\"class_name\":\"contractclassname\"}",
		s.contractSpecID,
		s.user1,
	)
	s.contractSpecAsText = fmt.Sprintf(`class_name: contractclassname
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

	s.recordSpecAsJson = fmt.Sprintf("{\"specification_id\":\"%s\",\"name\":\"recordname\",\"inputs\":[{\"name\":\"inputname\",\"type_name\":\"inputtypename\",\"hash\":\"alsonotreallyasourcehash\"}],\"type_name\":\"recordtypename\",\"result_type\":\"DEFINITION_TYPE_RECORD\",\"responsible_parties\":[\"PARTY_TYPE_OWNER\"]}",
		s.recordSpecID,
	)
	s.recordSpecAsText = fmt.Sprintf(`inputs:
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

	//os locators
	// add os locator
	s.ownerAddr = s.user1Addr
	s.uri = "http://foo.com"
	s.objectLocator = metadatatypes.NewOSLocatorRecord(s.ownerAddr, s.uri)

	s.ownerAddr1 = s.user2Addr
	s.uri1 = "http://bar.com"
	s.objectLocator1 = metadatatypes.NewOSLocatorRecord(s.ownerAddr1, s.uri1)

	var metadataData metadatatypes.GenesisState
	s.Require().NoError(cfg.Codec.UnmarshalJSON(genesisState[metadatatypes.ModuleName], &metadataData))
	metadataData.Scopes = append(metadataData.Scopes, s.scope)
	metadataData.Sessions = append(metadataData.Sessions, s.session)
	metadataData.Records = append(metadataData.Records, s.record)
	metadataData.ScopeSpecifications = append(metadataData.ScopeSpecifications, s.scopeSpec)
	metadataData.ContractSpecifications = append(metadataData.ContractSpecifications, s.contractSpec)
	metadataData.RecordSpecifications = append(metadataData.RecordSpecifications, s.recordSpec)
	metadataData.ObjectStoreLocators = append(metadataData.ObjectStoreLocators, s.objectLocator, s.objectLocator1)
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

func indent(str string) string {
	var sb strings.Builder
	lines := strings.Split(str, "\n")
	maxI := len(lines) - 1
	for i, l := range strings.Split(str, "\n") {
		sb.WriteString("  ")
		sb.WriteString(l)
		if i != maxI {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func yamlListEntry(str string) string {
	var sb strings.Builder
	lines := strings.Split(str, "\n")
	maxI := len(lines) - 1
	for i, l := range strings.Split(str, "\n") {
		if i == 0 {
			sb.WriteString("- ")
		} else {
			sb.WriteString("  ")
		}
		sb.WriteString(l)
		if i != maxI {
			sb.WriteString("\n")
		}
	}
	return sb.String()
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

func (s *IntegrationTestSuite) TestGetMetadataByIDCmd() {
	cmd := cli.GetMetadataByIDCmd()

	testCases := []queryCmdTestCase{
		{
			"get metadata by id - scope id as json",
			[]string{s.scopeID.String(), s.asJson},
			"",
			s.fullScopeAsJson,
		},
		{
			"get metadata by id - scope id as text",
			[]string{s.scopeID.String(), s.asText},
			"",
			s.fullScopeAsText,
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
			s.sessionAsJson,
		},
		{
			"get metadata by id - session id as text",
			[]string{s.sessionID.String(), s.asText},
			"",
			s.sessionAsText,
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
			s.recordAsJson,
		},
		{
			"get metadata by id - record id as text",
			[]string{s.recordID.String(), s.asText},
			"",
			s.recordAsText,
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
			s.scopeSpecAsJson,
		},
		{
			"get metadata by id - scope spec id as text",
			[]string{s.scopeSpecID.String(), s.asText},
			"",
			s.scopeSpecAsText,
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
			s.contractSpecAsJson,
		},
		{
			"get metadata by id - contract spec id as text",
			[]string{s.contractSpecID.String(), s.asText},
			"",
			s.contractSpecAsText,
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
			s.recordSpecAsJson,
		},
		{
			"get metadata by id - record spec id as text",
			[]string{s.recordSpecID.String(), s.asText},
			"",
			s.recordSpecAsText,
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

	testCases := []queryCmdTestCase{
		{
			"get scope by metadata scope id as json output",
			[]string{s.scopeID.String(), s.asJson},
			"",
			s.scopeAsJson,
		},
		{
			"get scope by metadata scope id as text output",
			[]string{s.scopeID.String(), s.asText},
			"",
			s.scopeAsText,
		},
		{
			"get scope by uuid as json output",
			[]string{s.scopeUUID.String(), s.asJson},
			"",
			s.scopeAsJson,
		},
		{
			"get scope by uuid as text output",
			[]string{s.scopeUUID.String(), s.asText},
			"",
			s.scopeAsText,
		},
		{
			"get scope by metadata session id as json output",
			[]string{s.sessionID.String(), s.asJson},
			"",
			s.scopeAsJson,
		},
		{
			"get scope by metadata session id as text output",
			[]string{s.sessionID.String(), s.asText},
			"",
			s.scopeAsText,
		},
		{
			"get scope by metadata record id as json output",
			[]string{s.recordID.String(), s.asJson},
			"",
			s.scopeAsJson,
		},
		{
			"get scope by metadata record id as text output",
			[]string{s.recordID.String(), s.asText},
			"",
			s.scopeAsText,
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

func (s *IntegrationTestSuite) TestGetMetadataFullScopeCmd() {
	cmd := cli.GetMetadataFullScopeCmd()

	testCases := []queryCmdTestCase{
		{
			"from scope id as json",
			[]string{s.scopeID.String(), s.asJson},
			"",
			s.fullScopeAsJson,
		},
		{
			"from scope id as text",
			[]string{s.scopeID.String(), s.asText},
			"",
			s.fullScopeAsText,
		},
		{
			"from scope uuid as json",
			[]string{s.scopeUUID.String(), s.asJson},
			"",
			s.fullScopeAsJson,
		},
		{
			"from scope uuid as text",
			[]string{s.scopeUUID.String(), s.asText},
			"",
			s.fullScopeAsText,
		},
		{
			"from session id as json",
			[]string{s.sessionID.String(), s.asJson},
			"",
			s.fullScopeAsJson,
		},
		{
			"from session id as text",
			[]string{s.sessionID.String(), s.asText},
			"",
			s.fullScopeAsText,
		},
		{
			"from record id as json",
			[]string{s.recordID.String(), s.asJson},
			"",
			s.fullScopeAsJson,
		},
		{
			"from record id as text",
			[]string{s.recordID.String(), s.asText},
			"",
			s.fullScopeAsText,
		},
		{
			"scope id does not exist",
			[]string{"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
			"rpc error: code = NotFound desc = scope uuid 91978ba2-5f35-459a-86a7-feca1b0512e0 not found: key not found",
			"",
		},
		{
			"scope uuid does not exist",
			[]string{"91978ba2-5f35-459a-86a7-feca1b0512e0"},
			"rpc error: code = NotFound desc = scope uuid 91978ba2-5f35-459a-86a7-feca1b0512e0 not found: key not found",
			"",
		},
		{
			"bad prefix",
			[]string{"foo1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
			"argument foo1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel is neither a metadata address (decoding bech32 failed: checksum failed. Expected kzwk8c, got xlkwel.) nor uuid (invalid UUID format)",
			"",
		},
		{
			"bad arg",
			[]string{"not-an-argument"},
			"argument not-an-argument is neither a metadata address (decoding bech32 failed: invalid index of 1) nor uuid (invalid UUID length: 15)",
			"",
		},
		{
			"two args",
			[]string{"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel", "record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3"},
			"accepts 1 arg(s), received 2",
			"",
		},
		{
			"no args",
			[]string{},
			"accepts 1 arg(s), received 0",
			"",
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationTestSuite) TestGetMetadataSessionCmd() {
	cmd := cli.GetMetadataSessionCmd()

	sessionListAsJson := fmt.Sprintf("[%s]", s.sessionAsJson)
	sessionListAsText := yamlListEntry(s.sessionAsText)

	testCases := []queryCmdTestCase{
		{
			"session from session id as json",
			[]string{s.sessionID.String(), s.asJson},
			"",
			s.sessionAsJson,
		},
		{
			"session from session id as text",
			[]string{s.sessionID.String(), s.asText},
			"",
			s.sessionAsText,
		},
		{
			"session id does not exist",
			[]string{"session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"},
			"rpc error: code = NotFound desc = session id session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr not found: key not found",
			"",
		},
		{
			"sessions from scope id as json",
			[]string{s.scopeID.String(), s.asJson},
			"",
			sessionListAsJson,
		},
		{
			"sessions from scope id as text",
			[]string{s.scopeID.String(), s.asText},
			"",
			sessionListAsText,
		},
		{
			"scope id does not exist",
			[]string{"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
			"rpc error: code = NotFound desc = scope uuid 91978ba2-5f35-459a-86a7-feca1b0512e0 not found: key not found",
			"",
		},
		{
			"sessions from scope uuid as json",
			[]string{s.scopeUUID.String(), s.asJson},
			"",
			sessionListAsJson,
		},
		{
			"sessions from scope uuid as text",
			[]string{s.scopeUUID.String(), s.asText},
			"",
			sessionListAsText,
		},
		{
			"scope uuid does not exist",
			[]string{"91978ba2-5f35-459a-86a7-feca1b0512e0"},
			"rpc error: code = NotFound desc = scope uuid 91978ba2-5f35-459a-86a7-feca1b0512e0 not found: key not found",
			"",
		},
		{
			"session from scope uuid and session uuid as json",
			[]string{s.scopeUUID.String(), s.sessionUUID.String(), s.asJson},
			"",
			s.sessionAsJson,
		},
		{
			"session from scope uuid ad session uuid as text",
			[]string{s.scopeUUID.String(), s.sessionUUID.String(), s.asText},
			"",
			s.sessionAsText,
		},
		{
			"scope uuid exists but session uuid does not exist",
			[]string{s.scopeUUID.String(), "5803f8bc-6067-4eb5-951f-2121671c2ec0"},
			fmt.Sprintf("rpc error: code = NotFound desc = session id %s not found: key not found",
				metadatatypes.SessionMetadataAddress(s.scopeUUID, uuid.MustParse("5803f8bc-6067-4eb5-951f-2121671c2ec0")),
			),
			"",
		},
		{
			"bad prefix",
			[]string{"scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"},
			"unexpected metadata address prefix on scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m",
			"",
		},
		{
			"bad arg 1",
			[]string{"bad"},
			"argument bad is neither a metadata address (decoding bech32 failed: invalid bech32 string length 3) nor uuid (invalid UUID length: 3)",
			"",
		},
		{
			"good arg 1, bad arg 2",
			[]string{s.scopeUUID.String(), "still-bad"},
			"invalid UUID length: 9",
			"",
		},
		{
			"3 args",
			[]string{s.scopeUUID.String(), s.sessionID.String(), s.recordName},
			"accepts between 1 and 2 arg(s), received 3",
			"",
		},
		{
			"no args",
			[]string{},
			"accepts between 1 and 2 arg(s), received 0",
			"",
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationTestSuite) TestGetMetadataRecordCmd() {
	cmd := cli.GetMetadataRecordCmd()

	recordListAsJson := fmt.Sprintf("[%s]", s.recordAsJson)
	recordListAsText := yamlListEntry(s.recordAsText)

	testCases := []queryCmdTestCase{
		{
			"record from record id as json",
			[]string{s.recordID.String(), s.asJson},
			"",
			s.recordAsJson,
		},
		{
			"record from record id as text",
			[]string{s.recordID.String(), s.asText},
			"",
			s.recordAsText,
		},
		{
			"record id does not exist",
			[]string{"record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3"},
			"rpc error: code = NotFound desc = scope uuid 91978ba2-5f35-459a-86a7-feca1b0512e0 not found: key not found",
			"",
		},
		{
			"records from session id as json",
			[]string{s.sessionID.String(), s.asJson},
			"",
			recordListAsJson,
		},
		{
			"records from session id as text",
			[]string{s.sessionID.String(), s.asText},
			"",
			recordListAsText,
		},
		{
			"session id does not exist",
			[]string{"session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"},
			"rpc error: code = NotFound desc = scope uuid 91978ba2-5f35-459a-86a7-feca1b0512e0 not found: key not found",
			"",
		},
		{
			"records from scope id as json",
			[]string{s.scopeID.String(), s.asJson},
			"",
			recordListAsJson,
		},
		{
			"records from scope id as text",
			[]string{s.scopeID.String(), s.asText},
			"",
			recordListAsText,
		},
		{
			"scope id does not exist",
			[]string{"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
			"rpc error: code = NotFound desc = scope uuid 91978ba2-5f35-459a-86a7-feca1b0512e0 not found: key not found",
			"",
		},
		{
			"records from scope uuid as json",
			[]string{s.scopeUUID.String(), s.asJson},
			"",
			recordListAsJson,
		},
		{
			"records from scope uuid as text",
			[]string{s.scopeUUID.String(), s.asText},
			"",
			recordListAsText,
		},
		{
			"scope uuid does not exist",
			[]string{"91978ba2-5f35-459a-86a7-feca1b0512e0"},
			"rpc error: code = NotFound desc = scope uuid 91978ba2-5f35-459a-86a7-feca1b0512e0 not found: key not found",
			"",
		},
		{
			"record from scope uuid and record name as json",
			[]string{s.scopeUUID.String(), s.recordName, s.asJson},
			"",
			s.recordAsJson,
		},
		{
			"record from scope uuid and record name as text",
			[]string{s.scopeUUID.String(), s.recordName, s.asText},
			"",
			s.recordAsText,
		},
		{
			"scope uuid exists but record name does not",
			[]string{s.scopeUUID.String(), "nope"},
			fmt.Sprintf("no records with id %s found in scope with uuid %s",
				metadatatypes.RecordMetadataAddress(s.scopeUUID, "nope"),
				s.scopeUUID,
			),
			"",
		},
		{
			"bad prefix",
			[]string{"contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"},
			"unexpected metadata address prefix on contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn",
			"",
		},
		{
			"bad arg 1",
			[]string{"badbad"},
			"argument badbad is neither a metadata address (decoding bech32 failed: invalid bech32 string length 6) nor uuid (invalid UUID length: 6)",
			"",
		},
		{
			"uuid arg 1 and whitespace args 2 and 3 as json",
			[]string{s.scopeUUID.String(), "  ", " ", s.asJson},
			"",
			recordListAsJson,
		},
		{
			"uuid arg 1 and whitespace args 2 and 3 as text",
			[]string{s.scopeUUID.String(), "  ", " ", s.asText},
			"",
			recordListAsText,
		},
		{
			"no args",
			[]string{},
			"requires at least 1 arg(s), only received 0",
			"",
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationTestSuite) TestGetMetadataScopeSpecCmd() {
	cmd := cli.GetMetadataScopeSpecCmd()

	testCases := []queryCmdTestCase{
		{
			"scope spec from scope spec id as json",
			[]string{s.scopeSpecID.String(), s.asJson},
			"",
			s.scopeSpecAsJson,
		},
		{
			"scope spec from scope spec id as text",
			[]string{s.scopeSpecID.String(), s.asText},
			"",
			s.scopeSpecAsText,
		},
		{
			"scope spec id bad prefix",
			[]string{"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
			"id scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel is not a scope specification metadata address",
			"",
		},
		{
			"scope spec id does not exist",
			[]string{"scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"},
			"rpc error: code = NotFound desc = scope specification uuid dc83ea70-eacd-40fe-9adf-1cf6148bf8a2 not found: key not found",
			"",
		},
		{
			"scope spec from scope spec uuid as json",
			[]string{s.scopeSpecUUID.String(), s.asJson},
			"",
			s.scopeSpecAsJson,
		},
		{
			"scope spec from scope spec uuid as text",
			[]string{s.scopeSpecUUID.String(), s.asText},
			"",
			s.scopeSpecAsText,
		},
		{
			"scope spec uuid does not exist",
			[]string{"dc83ea70-eacd-40fe-9adf-1cf6148bf8a2"},
			"rpc error: code = NotFound desc = scope specification uuid dc83ea70-eacd-40fe-9adf-1cf6148bf8a2 not found: key not found",
			"",
		},
		{
			"bad arg",
			[]string{"reallybad"},
			"argument reallybad is neither a metadata address (decoding bech32 failed: invalid index of 1) nor uuid (invalid UUID length: 9)",
			"",
		},
		{
			"two args",
			[]string{"arg1", "arg2"},
			"accepts 1 arg(s), received 2",
			"",
		},
		{
			"no args",
			[]string{},
			"accepts 1 arg(s), received 0",
			"",
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationTestSuite) TestGetMetadataContractSpecCmd() {
	cmd := cli.GetMetadataContractSpecCmd()

	testCases := []queryCmdTestCase{
		{
			"contract spec from contract spec id as json",
			[]string{s.contractSpecID.String(), s.asJson},
			"",
			s.contractSpecAsJson,
		},
		{
			"contract spec from contract spec id as text",
			[]string{s.contractSpecID.String(), s.asText},
			"",
			s.contractSpecAsText,
		},
		{
			"contract spec id does not exist",
			[]string{"contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"},
			"rpc error: code = NotFound desc = contract specification with id contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn (uuid def6bc0a-c9dd-4874-948f-5206e6060a84) not found: key not found",
			"",
		},
		{
			"contract spec from contract spec uuid as json",
			[]string{s.contractSpecUUID.String(), s.asJson},
			"",
			s.contractSpecAsJson,
		},
		{
			"contract spec from contract spec uuid as text",
			[]string{s.contractSpecUUID.String(), s.asText},
			"",
			s.contractSpecAsText,
		},
		{
			"contract spec uuid does not exist",
			[]string{"def6bc0a-c9dd-4874-948f-5206e6060a84"},
			"rpc error: code = NotFound desc = contract specification with id contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn (uuid def6bc0a-c9dd-4874-948f-5206e6060a84) not found: key not found",
			"",
		},
		{
			"contract spec from record spec id as json",
			[]string{s.recordSpecID.String(), s.asJson},
			"",
			s.contractSpecAsJson,
		},
		{
			"contract spec from record spec id as text",
			[]string{s.recordSpecID.String(), s.asText},
			"",
			s.contractSpecAsText,
		},
		{
			"record spec id does not exist",
			[]string{"recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44"},
			"rpc error: code = NotFound desc = contract specification with id contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn (uuid def6bc0a-c9dd-4874-948f-5206e6060a84) not found: key not found",
			"",
		},
		{
			"bad prefix",
			[]string{"record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3"},
			"unexpected metadata address prefix on record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3: this metadata address (record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3) does not contain a contract specification uuid",
			"",
		},
		{
			"bad arg",
			[]string{"badbadarg"},
			"argument badbadarg is neither a metadata address (decoding bech32 failed: invalid index of 1) nor uuid (invalid UUID length: 9)",
			"",
		},
		{
			"two args",
			[]string{"arg1", "arg2"},
			"accepts 1 arg(s), received 2",
			"",
		},
		{
			"no args",
			[]string{},
			"accepts 1 arg(s), received 0",
			"",
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationTestSuite) TestGetMetadataRecordSpecCmd() {
	cmd := cli.GetMetadataRecordSpecCmd()

	recordSpecListAsJson := fmt.Sprintf("[%s]", s.recordSpecAsJson)
	recordSpecListAsText := yamlListEntry(s.recordSpecAsText)

	testCases := []queryCmdTestCase{
		{
			"record spec from rec spec id as json",
			[]string{s.recordSpecID.String(), s.asJson},
			"",
			s.recordSpecAsJson,
		},
		{
			"record spec from rec spec id as text",
			[]string{s.recordSpecID.String(), s.asText},
			"",
			s.recordSpecAsText,
		},
		{
			"rec spec id does not exist",
			[]string{"recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44"},
			"rpc error: code = NotFound desc = record specification not found for id recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44: key not found",
			"",
		},
		{
			"record specs from contract spec id as json",
			[]string{s.contractSpecID.String(), s.asJson},
			"",
			recordSpecListAsJson,
		},
		{
			"record specs from contract spec id as text",
			[]string{s.contractSpecID.String(), s.asText},
			"",
			recordSpecListAsText,
		},
		{
			"contract spec id does not exist",
			[]string{"contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"},
			"rpc error: code = NotFound desc = no record specifications found for contract spec uuid def6bc0a-c9dd-4874-948f-5206e6060a84: key not found",
			"",
		},
		{
			"record specs from contract spec uuid as json",
			[]string{s.contractSpecUUID.String(), s.asJson},
			"",
			recordSpecListAsJson,
		},
		{
			"record specs from contract spec uuid as text",
			[]string{s.contractSpecUUID.String(), s.asText},
			"",
			recordSpecListAsText,
		},
		{
			"contract spec uuid does not exist",
			[]string{"def6bc0a-c9dd-4874-948f-5206e6060a84"},
			"rpc error: code = NotFound desc = no record specifications found for contract spec uuid def6bc0a-c9dd-4874-948f-5206e6060a84: key not found",
			"",
		},
		{
			"record spec from contract spec uuid and record spec name as json",
			[]string{s.contractSpecUUID.String(), s.recordName, s.asJson},
			"",
			s.recordSpecAsJson,
		},
		{
			"record spec from contract spec uuid and record spec name as text",
			[]string{s.contractSpecUUID.String(), s.recordName, s.asText},
			"",
			s.recordSpecAsText,
		},
		{
			"contract spec uuid exists record spec name does not",
			[]string{s.contractSpecUUID.String(), "nopenopenope"},
			fmt.Sprintf("rpc error: code = NotFound desc = record specification not found for id %s: key not found",
				metadatatypes.RecordSpecMetadataAddress(s.contractSpecUUID, "nopenopenope"),
			),
			"",
		},
		{
			"record specs from contract spec uuid and only whitespace name args",
			[]string{s.contractSpecUUID.String(), "   ", " ", "      "},
			"",
			recordSpecListAsText,
		},
		{
			"bad prefix",
			[]string{"session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"},
			"unexpected metadata address prefix on session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr",
			"",
		},
		{
			"bad arg 1",
			[]string{"not-gonna-parse"},
			"argument not-gonna-parse is neither a metadata address (decoding bech32 failed: invalid index of 1) nor uuid (invalid UUID length: 15)",
			"",
		},
		{
			"no args",
			[]string{s.asJson},
			"requires at least 1 arg(s), only received 0",
			"",
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationTestSuite) TestGetOwnershipCmd() {
	cmd := cli.GetOwnershipCmd()

	newUser := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()

	ownedScopesAsJson := fmt.Sprintf("{\"scope_uuids\":[\"%s\"],\"pagination\":{\"next_key\":null,\"total\":\"1\"}}",
		s.scopeUUID,
	)
	ownedScopesAsText := fmt.Sprintf(`pagination:
  next_key: null
  total: "1"
scope_uuids:
- %s`,
		s.scopeUUID,
	)

	testCases := []queryCmdTestCase{
		{
			"scopes as json",
			[]string{s.user1, s.asJson},
			"",
			ownedScopesAsJson,
		},
		{
			"scopes as text",
			[]string{s.user1, s.asText},
			"",
			ownedScopesAsText,
		},
		{
			"scope through value owner",
			[]string{s.user2},
			"",
			ownedScopesAsText,
		},
		{
			"no result",
			[]string{newUser},
			fmt.Sprintf("no scopes are owned by address %s", newUser),
			"",
		},
		{
			"two args",
			[]string{s.user1, s.user2},
			"accepts 1 arg(s), received 2",
			"",
		},
		{
			"no args",
			[]string{},
			"accepts 1 arg(s), received 0",
			"",
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationTestSuite) TestGetValueOwnershipCmd() {
	cmd := cli.GetValueOwnershipCmd()

	ownedScopesAsJson := fmt.Sprintf("{\"scope_uuids\":[\"%s\"],\"pagination\":{\"next_key\":null,\"total\":\"1\"}}",
		s.scopeUUID,
	)
	ownedScopesAsText := fmt.Sprintf(`pagination:
  next_key: null
  total: "1"
scope_uuids:
- %s`,
		s.scopeUUID,
	)

	testCases := []queryCmdTestCase{
		{
			"as json",
			[]string{s.user2, s.asJson},
			"",
			ownedScopesAsJson,
		},
		{
			"as text",
			[]string{s.user2, s.asText},
			"",
			ownedScopesAsText,
		},
		{
			"no result",
			[]string{s.user1},
			fmt.Sprintf("address %s is not the value owner on any scopes",
				s.user1),
			"",
		},
		{
			"two args",
			[]string{s.user1, s.user2},
			"accepts 1 arg(s), received 2",
			"",
		},
		{
			"no args",
			[]string{},
			"accepts 1 arg(s), received 0",
			"",
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

// ---------- tx cmd tests ----------

func (s *IntegrationTestSuite) TestAddMetadataScopeCmd() {

	scopeUUID := uuid.New().String()
	pubkey := secp256k1.GenPrivKey().PubKey()
	userAddr := sdk.AccAddress(pubkey.Address())
	user := userAddr.String()

	testCases := []commonTestStruct{
		{
			"Should successfully add metadata scope",
			cli.AddMetadataScopeCmd(),
			[]string{
				scopeUUID,
				uuid.New().String(),
				user,
				user,
				user,
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
				user,
				user,
				user,
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
				user,
				user,
				user,
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
				user,
				user,
				user,
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
				user,
				user,
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
				user,
				"incorrect,incorrect",
				user,
				s.testnet.Validators[0].Address.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, user),
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
				user,
				user,
				"incorrect",
				s.testnet.Validators[0].Address.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, user),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
	}

	s.runTestCase(testCases)
}

type commonTestStruct struct {
	name         string
	cmd          *cobra.Command
	args         []string
	expectErr    bool
	respType     proto.Message
	expectedCode uint32
}

func (s *IntegrationTestSuite) TestRemoveMetadataScopeCmd() {

	userId := s.testnet.Validators[0].Address.String()
	scopeUUID := uuid.New().String()

	testCases := []commonTestStruct{
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

	s.runTestCase(testCases)
}

// os locator tx cmds
func (s *IntegrationTestSuite) TestAddObjectLocatorCmd() {
	userURI := "http://foo.com"
	userURIMod := "https://www.google.com/search?q=red+butte+garden&oq=red+butte+garden&aqs=chrome..69i57j46i131i175i199i433j0j0i457j0l6.3834j0j7&sourceid=chrome&ie=UTF-8#lpqa=d,2"
	testCases := []commonTestStruct{
		{
			"Should successfully add os locator",
			cli.AddOsLocatorCmd(),
			[]string{
				s.testnet.Validators[0].Address.String(),
				userURI,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"Should successfully Modify os locator",
			cli.ModifyOsLocatorCmd(),
			[]string{
				s.testnet.Validators[0].Address.String(),
				userURIMod,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"Should successfully delete os locator",
			cli.RemoveOsLocatorCmd(),
			[]string{
				s.testnet.Validators[0].Address.String(),
				userURIMod,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
	}
	s.runTestCase(testCases)
}

func (s *IntegrationTestSuite) TestGetOSLocatorCmd() {
	cmd := cli.GetOSLocatorCmd()

	testCases := []queryCmdTestCase{
		{
			"get os locator",
			[]string{s.user1Addr.String(), s.asJson},
			"",
			fmt.Sprintf("{\"owner\":\"%s\",\"locator_uri\":\"%s\"}", s.user1Addr.String(), "http://foo.com"),
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationTestSuite) TestGetAllOSLocatorCmd() {
	cmd := cli.GetOSLocatorCmd()

	testCases := []queryCmdTestCase{
		{
			"get os locator",
			[]string{s.user1Addr.String(), s.asJson},
			"",
			fmt.Sprintf("{\"owner\":\"%s\",\"locator_uri\":\"%s\"}", s.user1Addr.String(), "http://foo.com"),
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationTestSuite) TestAddRecordSpecificationCmd() {
	cmd := cli.AddRecordSpecificationCmd()
	recordName := "testrecordspecid"
	specificationID := types.RecordSpecMetadataAddress(s.contractSpecUUID, "testrecordspecid")
	testCases := []commonTestStruct{
		{
			"Should successfully add os locator",
			cmd,
			[]string{
				specificationID.String(),
				recordName,
				"record1,typename1,hash,hashy;record2,typename2,hash,hashy",
				"typename",
				"resulttypes",
				"responsibleparties",
				fmt.Sprintf("%s", s.user1),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
	}
	s.runTestCase(testCases)
}

func (s *IntegrationTestSuite) runTestCase(testCases []commonTestStruct) {
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
