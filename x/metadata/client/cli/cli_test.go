package cli_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
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

type IntegrationCLITestSuite struct {
	suite.Suite

	cfg     testnet.Config
	testnet *testnet.Network

	asJson         string
	asText         string
	includeRequest string

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

	scopeAsJson string
	scopeAsText string

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

func TestIntegrationCLITestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationCLITestSuite))
}

func (s *IntegrationCLITestSuite) SetupSuite() {
	s.accountKey = secp256k1.GenPrivKeyFromSecret([]byte("acc2"))
	addr, err := sdk.AccAddressFromHex(s.accountKey.PubKey().Address().String())
	s.Require().NoError(err)
	s.accountAddr = addr
	s.T().Log("setting up integration test suite")

	cfg := testutil.DefaultTestNetworkConfig()

	s.asJson = fmt.Sprintf("--%s=json", tmcli.OutputFlag)
	s.asText = fmt.Sprintf("--%s=text", tmcli.OutputFlag)
	s.includeRequest = "--include-request"

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
		s.recordSpecID,
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

	s.sessionAsJson = fmt.Sprintf("{\"session_id\":\"%s\",\"specification_id\":\"%s\",\"parties\":[{\"address\":\"%s\",\"role\":\"PARTY_TYPE_OWNER\"}],\"name\":\"unit test session\",\"context\":null,\"audit\":{\"created_date\":\"0001-01-01T00:00:00Z\",\"created_by\":\"%s\",\"updated_date\":\"0001-01-01T00:00:00Z\",\"updated_by\":\"\",\"version\":0,\"message\":\"unit testing\"}}",
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
context: null
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

	s.recordAsJson = fmt.Sprintf("{\"name\":\"recordname\",\"session_id\":\"%s\",\"process\":{\"hash\":\"notarealprocesshash\",\"name\":\"record process\",\"method\":\"myMethod\"},\"inputs\":[{\"name\":\"inputname\",\"hash\":\"notarealrecordinputhash\",\"type_name\":\"inputtypename\",\"status\":\"RECORD_INPUT_STATUS_RECORD\"}],\"outputs\":[{\"hash\":\"notarealrecordoutputhash\",\"status\":\"RESULT_STATUS_PASS\"}],\"specification_id\":\"%s\"}",
		s.sessionID,
		s.recordSpecID,
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
session_id: %s
specification_id: %s`,
		s.sessionID,
		s.recordSpecID,
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

func (s *IntegrationCLITestSuite) TearDownSuite() {
	s.testnet.WaitForNextBlock()
	s.T().Log("tearing down integration test suite")
	s.testnet.Cleanup()
}

func indent(str string, spaces int) string {
	var sb strings.Builder
	lines := strings.Split(str, "\n")
	maxI := len(lines) - 1
	s := strings.Repeat(" ", spaces)
	for i, l := range strings.Split(str, "\n") {
		sb.WriteString(s)
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
	name             string
	args             []string
	expectedError    string
	expectedInOutput []string
}

func runQueryCmdTestCases(s *IntegrationCLITestSuite, cmd *cobra.Command, testCases []queryCmdTestCase) {
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if len(tc.expectedError) > 0 {
				actualError := ""
				if err != nil {
					actualError = err.Error()
				}
				require.Contains(t, actualError, tc.expectedError, "expected error")
				// Something deep down is double wrapping the errors.
				// E.g. "rpc error: code = InvalidArgument desc = foo: invalid request" has become
				// "rpc error: code = InvalidArgument desc = rpc error: code = InvalidArgument desc = foo: invalid request"
				// So we changed from the "Equal" test below to the "Contains" test above.
				// If you're bored, maybe try swapping back to see if things have been fixed.
				//require.Equal(t, tc.expectedError, actualError, "expected error")
			} else {
				require.NoError(t, err, "unexpected error")
			}
			if err == nil {
				result := strings.TrimSpace(out.String())
				for _, exp := range tc.expectedInOutput {
					assert.Contains(t, result, exp)
				}
			}
		})
	}
}

func (s *IntegrationCLITestSuite) TestGetMetadataParamsCmd() {
	cmd := cli.GetMetadataParamsCmd()

	testCases := []queryCmdTestCase{
		{
			"get params as json output",
			[]string{s.asJson},
			"",
			[]string{"\"params\":{}"},
		},
		{
			"get params as text output",
			[]string{s.asText},
			"",
			[]string{"params: {}"},
		},
		{
			"get params - invalid args",
			[]string{"bad-arg"},
			"unknown argument: bad-arg",
			[]string{},
		},
		{
			"get params as json output including request",
			[]string{s.asJson, s.includeRequest},
			"",
			[]string{"\"params\":{}", "\"request\":{}"},
		},
		{
			"get locator params as json",
			[]string{"locator", s.asJson},
			"",
			[]string{"\"params\":{", "\"max_uri_length\":2048"},
		},
		{
			"get locator params as text including request",
			[]string{"locator", s.asText, s.includeRequest},
			"",
			[]string{"params:", "max_uri_length: 2048", "request: {}"},
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetMetadataByIDCmd() {
	cmd := cli.GetMetadataByIDCmd()

	testCases := []queryCmdTestCase{
		{
			"get metadata by id - scope id as json",
			[]string{s.scopeID.String(), s.asJson},
			"",
			[]string{s.scopeAsJson},
		},
		{
			"get metadata by id - scope id as text",
			[]string{s.scopeID.String(), s.asText},
			"",
			[]string{indent(s.scopeAsText, 4)},
		},
		{
			"get metadata by id - scope id does not exist",
			[]string{"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
			"",
			[]string{"scope: null", "scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
		},
		{
			"get metadata by id - session id as json",
			[]string{s.sessionID.String(), s.asJson},
			"",
			[]string{s.sessionAsJson},
		},
		{
			"get metadata by id - session id as text",
			[]string{s.sessionID.String(), s.asText},
			"",
			[]string{indent(s.sessionAsText, 4)},
		},
		{
			"get metadata by id - session id does not exist",
			[]string{"session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"},
			"",
			[]string{"session: null", "session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"},
		},
		{
			"get metadata by id - record id as json",
			[]string{s.recordID.String(), s.asJson},
			"",
			[]string{s.recordAsJson},
		},
		{
			"get metadata by id - record id as text",
			[]string{s.recordID.String(), s.asText},
			"",
			[]string{indent(s.recordAsText, 4)},
		},
		{
			"get metadata by id - record id does not exist",
			[]string{"record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3"},
			"",
			[]string{"record: null", "record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3"},
		},
		{
			"get metadata by id - scope spec id as json",
			[]string{s.scopeSpecID.String(), s.asJson},
			"",
			[]string{s.scopeSpecAsJson},
		},
		{
			"get metadata by id - scope spec id as text",
			[]string{s.scopeSpecID.String(), s.asText},
			"",
			[]string{indent(s.scopeSpecAsText, 4)},
		},
		{
			"get metadata by id - scope spec id does not exist",
			[]string{"scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"},
			"",
			[]string{"specification: null", "scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"},
		},
		{
			"get metadata by id - contract spec id as json",
			[]string{s.contractSpecID.String(), s.asJson},
			"",
			[]string{s.contractSpecAsJson},
		},
		{
			"get metadata by id - contract spec id as text",
			[]string{s.contractSpecID.String(), s.asText},
			"",
			[]string{indent(s.contractSpecAsText, 4)},
		},
		{
			"get metadata by id - contract spec id does not exist",
			[]string{"contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"},
			"",
			[]string{"specification: null", "contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"},
		},
		{
			"get metadata by id - record spec id as json",
			[]string{s.recordSpecID.String(), s.asJson},
			"",
			[]string{s.recordSpecAsJson},
		},
		{
			"get metadata by id - record spec id as text",
			[]string{s.recordSpecID.String(), s.asText},
			"",
			[]string{indent(s.recordSpecAsText, 4)},
		},
		{
			"get metadata by id - record spec id does not exist",
			[]string{"recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44"},
			"",
			[]string{"specification: null", "recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44"},
		},
		{
			"get metadata by id - bad prefix",
			[]string{"foo1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
			"decoding bech32 failed: checksum failed. Expected kzwk8c, got xlkwel.",
			[]string{},
		},
		{
			"get metadata by id - no args",
			[]string{},
			"accepts 1 arg(s), received 0",
			[]string{},
		},
		{
			"get metadata by id - two args",
			[]string{"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel", "scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"},
			"accepts 1 arg(s), received 2",
			[]string{},
		},
		{
			"get metadata by id - uuid",
			[]string{"91978ba2-5f35-459a-86a7-feca1b0512e0"},
			"decoding bech32 failed: invalid index of 1",
			[]string{},
		},
		{
			"get metadata by id - bad arg",
			[]string{"not-an-id"},
			"decoding bech32 failed: invalid index of 1",
			[]string{},
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

// TODO: GetMetadataGetAllCmd

func (s *IntegrationCLITestSuite) TestGetMetadataScopeCmd() {
	cmd := cli.GetMetadataScopeCmd()

	indentedScopeText := indent(s.scopeAsText, 4)

	testCases := []queryCmdTestCase{
		{
			"get scope by metadata scope id as json output",
			[]string{s.scopeID.String(), s.asJson},
			"",
			[]string{s.scopeAsJson},
		},
		{
			"get scope by metadata scope id as text output",
			[]string{s.scopeID.String(), s.asText},
			"",
			[]string{indentedScopeText},
		},
		{
			"get scope by uuid as json output",
			[]string{s.scopeUUID.String(), s.asJson},
			"",
			[]string{s.scopeAsJson},
		},
		{
			"get scope by uuid as text output",
			[]string{s.scopeUUID.String(), s.asText},
			"",
			[]string{indentedScopeText},
		},
		{
			"get scope by metadata session id as json output",
			[]string{s.sessionID.String(), s.asJson},
			"",
			[]string{s.scopeAsJson},
		},
		{
			"get scope by metadata session id as text output",
			[]string{s.sessionID.String(), s.asText},
			"",
			[]string{indentedScopeText},
		},
		{
			"get scope by metadata record id as json output",
			[]string{s.recordID.String(), s.asJson},
			"",
			[]string{s.scopeAsJson},
		},
		{
			"get scope by metadata record id as text output",
			[]string{s.recordID.String(), s.asText},
			"",
			[]string{indentedScopeText},
		},
		{
			"get scope by metadata id - does not exist",
			[]string{"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel", s.asText},
			"",
			[]string{"scope: null", "scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
		},
		{
			"get scope by uuid - does not exist",
			[]string{"91978ba2-5f35-459a-86a7-feca1b0512e0", s.asText},
			"",
			[]string{"scope: null", "91978ba2-5f35-459a-86a7-feca1b0512e0"},
		},
		{
			"get scope bad input",
			[]string{"not-a-valid-arg", s.asText},
			"rpc error: code = InvalidArgument desc = could not parse [not-a-valid-arg] into either a scope address (decoding bech32 failed: invalid index of 1) or uuid (invalid UUID length: 15): invalid request",
			[]string{},
		},
		{
			"get scope no args",
			[]string{},
			"accepts 1 arg(s), received 0",
			[]string{},
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetMetadataSessionCmd() {
	cmd := cli.GetMetadataSessionCmd()

	indentedSessionText := indent(s.sessionAsText, 4)

	testCases := []queryCmdTestCase{
		{
			"session from session id as json",
			[]string{s.sessionID.String(), s.asJson},
			"",
			[]string{s.sessionAsJson},
		},
		{
			"session from session id as text",
			[]string{s.sessionID.String(), s.asText},
			"",
			[]string{indentedSessionText},
		},
		{
			"session id does not exist",
			[]string{"session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"},
			"",
			[]string{"session: null", "session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"},
		},
		{
			"sessions from scope id as json",
			[]string{s.scopeID.String(), s.asJson},
			"",
			[]string{s.sessionAsJson},
		},
		{
			"sessions from scope id as text",
			[]string{s.scopeID.String(), s.asText},
			"",
			[]string{indentedSessionText},
		},
		{
			"scope id does not exist",
			[]string{"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
			"no sessions found",
			[]string{},
		},
		{
			"sessions from scope uuid as json",
			[]string{s.scopeUUID.String(), s.asJson},
			"",
			[]string{s.sessionAsJson},
		},
		{
			"sessions from scope uuid as text",
			[]string{s.scopeUUID.String(), s.asText},
			"",
			[]string{indentedSessionText},
		},
		{
			"scope uuid does not exist",
			[]string{"91978ba2-5f35-459a-86a7-feca1b0512e0"},
			"no sessions found",
			[]string{},
		},
		{
			"session from scope uuid and session uuid as json",
			[]string{s.scopeUUID.String(), s.sessionUUID.String(), s.asJson},
			"",
			[]string{s.sessionAsJson},
		},
		{
			"session from scope uuid ad session uuid as text",
			[]string{s.scopeUUID.String(), s.sessionUUID.String(), s.asText},
			"",
			[]string{indentedSessionText},
		},
		{
			"scope uuid exists but session uuid does not exist",
			[]string{s.scopeUUID.String(), "5803f8bc-6067-4eb5-951f-2121671c2ec0"},
			"",
			[]string{"session:", "session: null"},
		},
		{
			"bad prefix",
			[]string{"scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"},
			"rpc error: code = InvalidArgument desc = address [scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m] is not a scope address: invalid request",
			[]string{},
		},
		{
			"bad arg 1",
			[]string{"bad"},
			"rpc error: code = InvalidArgument desc = could not parse [bad] into either a scope address (decoding bech32 failed: invalid bech32 string length 3) or uuid (invalid UUID length: 3): invalid request",
			[]string{},
		},
		{
			"good arg 1, bad arg 2",
			[]string{s.scopeUUID.String(), "still-bad"},
			"rpc error: code = InvalidArgument desc = could not parse [still-bad] into either a session address (decoding bech32 failed: invalid index of 1) or uuid (invalid UUID length: 9): invalid request",
			[]string{},
		},
		{
			"3 args",
			[]string{s.scopeUUID.String(), s.sessionID.String(), s.recordName},
			"accepts between 1 and 2 arg(s), received 3",
			[]string{},
		},
		{
			"no args",
			[]string{},
			"accepts between 1 and 2 arg(s), received 0",
			[]string{},
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetMetadataRecordCmd() {
	cmd := cli.GetMetadataRecordCmd()

	testCases := []queryCmdTestCase{
		{
			"record from record id as json",
			[]string{s.recordID.String(), s.asJson},
			"",
			[]string{s.recordAsJson},
		},
		{
			"record from record id as text",
			[]string{s.recordID.String(), s.asText},
			"",
			[]string{indent(s.recordAsText, 4)},
		},
		{
			"record id does not exist",
			[]string{"record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3"},
			"",
			[]string{"records:", "record: null", "record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3"},
		},
		{
			"records from session id as json",
			[]string{s.sessionID.String(), s.asJson},
			"",
			[]string{s.recordAsJson},
		},
		{
			"records from session id as text",
			[]string{s.sessionID.String(), s.asText},
			"",
			[]string{indent(s.recordAsText, 4)},
		},
		{
			"session id does not exist",
			[]string{"session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"},
			"",
			[]string{"records: []"},
		},
		{
			"records from scope id as json",
			[]string{s.scopeID.String(), s.asJson},
			"",
			[]string{s.recordAsJson},
		},
		{
			"records from scope id as text",
			[]string{s.scopeID.String(), s.asText},
			"",
			[]string{indent(s.recordAsText, 4)},
		},
		{
			"scope id does not exist",
			[]string{"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
			"",
			[]string{"records: []"},
		},
		{
			"records from scope uuid as json",
			[]string{s.scopeUUID.String(), s.asJson},
			"",
			[]string{s.recordAsJson},
		},
		{
			"records from scope uuid as text",
			[]string{s.scopeUUID.String(), s.asText},
			"",
			[]string{indent(s.recordAsText, 4)},
		},
		{
			"scope uuid does not exist",
			[]string{"91978ba2-5f35-459a-86a7-feca1b0512e0"},
			"",
			[]string{"records: []"},
		},
		{
			"record from scope uuid and record name as json",
			[]string{s.scopeUUID.String(), s.recordName, s.asJson},
			"",
			[]string{s.recordAsJson},
		},
		{
			"record from scope uuid and record name as text",
			[]string{s.scopeUUID.String(), s.recordName, s.asText},
			"",
			[]string{indent(s.recordAsText, 4)},
		},
		{
			"scope uuid exists but record name does not",
			[]string{s.scopeUUID.String(), "nope"},
			"",
			[]string{"record: null"},
		},
		{
			"bad prefix",
			[]string{"contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"},
			"rpc error: code = InvalidArgument desc = address [contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn] is not a scope address: invalid request",
			[]string{},
		},
		{
			"bad arg 1",
			[]string{"badbad"},
			"rpc error: code = InvalidArgument desc = could not parse [badbad] into either a scope address (decoding bech32 failed: invalid bech32 string length 6) or uuid (invalid UUID length: 6): invalid request",
			[]string{},
		},
		{
			"uuid arg 1 and whitespace args 2 and 3 as json",
			[]string{s.scopeUUID.String(), "  ", " ", s.asJson},
			"",
			[]string{s.recordAsJson},
		},
		{
			"uuid arg 1 and whitespace args 2 and 3 as text",
			[]string{s.scopeUUID.String(), "  ", " ", s.asText},
			"",
			[]string{indent(s.recordAsText, 4)},
		},
		{
			"no args",
			[]string{},
			"requires at least 1 arg(s), only received 0",
			[]string{""},
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetMetadataScopeSpecCmd() {
	cmd := cli.GetMetadataScopeSpecCmd()

	testCases := []queryCmdTestCase{
		{
			"scope spec from scope spec id as json",
			[]string{s.scopeSpecID.String(), s.asJson},
			"",
			[]string{s.scopeSpecAsJson},
		},
		{
			"scope spec from scope spec id as text",
			[]string{s.scopeSpecID.String(), s.asText},
			"",
			[]string{indent(s.scopeSpecAsText, 4)},
		},
		{
			"scope spec id bad prefix",
			[]string{"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
			"rpc error: code = InvalidArgument desc = address [scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel] is not a scope spec address: invalid request",
			[]string{},
		},
		{
			"scope spec id does not exist",
			[]string{"scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"},
			"",
			[]string{"specification: null", "scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"},
		},
		{
			"scope spec from scope spec uuid as json",
			[]string{s.scopeSpecUUID.String(), s.asJson},
			"",
			[]string{s.scopeSpecAsJson},
		},
		{
			"scope spec from scope spec uuid as text",
			[]string{s.scopeSpecUUID.String(), s.asText},
			"",
			[]string{indent(s.scopeSpecAsText, 4)},
		},
		{
			"scope spec uuid does not exist",
			[]string{"dc83ea70-eacd-40fe-9adf-1cf6148bf8a2"},
			"",
			[]string{"specification: null", "dc83ea70-eacd-40fe-9adf-1cf6148bf8a2"},
		},
		{
			"bad arg",
			[]string{"reallybad"},
			"rpc error: code = InvalidArgument desc = could not parse [reallybad] into either a scope spec address (decoding bech32 failed: invalid index of 1) or uuid (invalid UUID length: 9): invalid request",
			[]string{},
		},
		{
			"two args",
			[]string{"arg1", "arg2"},
			"accepts 1 arg(s), received 2",
			[]string{},
		},
		{
			"no args",
			[]string{},
			"accepts 1 arg(s), received 0",
			[]string{},
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetMetadataContractSpecCmd() {
	cmd := cli.GetMetadataContractSpecCmd()

	testCases := []queryCmdTestCase{
		{
			"contract spec from contract spec id as json",
			[]string{s.contractSpecID.String(), s.asJson},
			"",
			[]string{s.contractSpecAsJson},
		},
		{
			"contract spec from contract spec id as text",
			[]string{s.contractSpecID.String(), s.asText},
			"",
			[]string{indent(s.contractSpecAsText, 4)},
		},
		{
			"contract spec id does not exist",
			[]string{"contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"},
			"",
			[]string{"specification: null", "contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"},
		},
		{
			"contract spec from contract spec uuid as json",
			[]string{s.contractSpecUUID.String(), s.asJson},
			"",
			[]string{s.contractSpecAsJson},
		},
		{
			"contract spec from contract spec uuid as text",
			[]string{s.contractSpecUUID.String(), s.asText},
			"",
			[]string{indent(s.contractSpecAsText, 4)},
		},
		{
			"contract spec uuid does not exist",
			[]string{"def6bc0a-c9dd-4874-948f-5206e6060a84"},
			"",
			[]string{"specification: null", "def6bc0a-c9dd-4874-948f-5206e6060a84"},
		},
		{
			"contract spec from record spec id as json",
			[]string{s.recordSpecID.String(), s.asJson},
			"",
			[]string{s.contractSpecAsJson},
		},
		{
			"contract spec from record spec id as text",
			[]string{s.recordSpecID.String(), s.asText},
			"",
			[]string{indent(s.contractSpecAsText, 4)},
		},
		{
			"record spec id does not exist",
			[]string{"recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44"},
			"",
			[]string{"specification: null"},
		},
		{
			"bad prefix",
			[]string{"record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3"},
			"rpc error: code = InvalidArgument desc = invalid specification id: address [record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3] is not a contract spec address: invalid request",
			[]string{},
		},
		{
			"bad arg",
			[]string{"badbadarg"},
			"rpc error: code = InvalidArgument desc = invalid specification id: could not parse [badbadarg] into either a contract spec address (decoding bech32 failed: invalid index of 1) or uuid (invalid UUID length: 9): invalid request",
			[]string{},
		},
		{
			"two args",
			[]string{"arg1", "arg2"},
			"accepts 1 arg(s), received 2",
			[]string{},
		},
		{
			"no args",
			[]string{},
			"accepts 1 arg(s), received 0",
			[]string{},
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetMetadataRecordSpecCmd() {
	cmd := cli.GetMetadataRecordSpecCmd()

	testCases := []queryCmdTestCase{
		{
			"record spec from rec spec id as json",
			[]string{s.recordSpecID.String(), s.asJson},
			"",
			[]string{s.recordSpecAsJson},
		},
		{
			"record spec from rec spec id as text",
			[]string{s.recordSpecID.String(), s.asText},
			"",
			[]string{indent(s.recordSpecAsText, 4)},
		},
		{
			"rec spec id does not exist",
			[]string{"recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44"},
			"",
			[]string{"record_specifications: []"},
		},
		{
			"record specs from contract spec id as json",
			[]string{s.contractSpecID.String(), s.asJson},
			"",
			[]string{s.recordSpecAsJson},
		},
		{
			"record specs from contract spec id as text",
			[]string{s.contractSpecID.String(), s.asText},
			"",
			[]string{indent(s.recordSpecAsText, 4)},
		},
		{
			"contract spec id does not exist",
			[]string{"contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"},
			"",
			[]string{"record_specifications: []", "contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"},
		},
		{
			"record specs from contract spec uuid as json",
			[]string{s.contractSpecUUID.String(), s.asJson},
			"",
			[]string{s.recordSpecAsJson},
		},
		{
			"record specs from contract spec uuid as text",
			[]string{s.contractSpecUUID.String(), s.asText},
			"",
			[]string{indent(s.recordSpecAsText, 4)},
		},
		{
			"contract spec uuid does not exist",
			[]string{"def6bc0a-c9dd-4874-948f-5206e6060a84"},
			"",
			[]string{"record_specifications: []", "def6bc0a-c9dd-4874-948f-5206e6060a84"},
		},
		{
			"record spec from contract spec uuid and record spec name as json",
			[]string{s.contractSpecUUID.String(), s.recordName, s.asJson},
			"",
			[]string{s.recordSpecAsJson},
		},
		{
			"record spec from contract spec uuid and record spec name as text",
			[]string{s.contractSpecUUID.String(), s.recordName, s.asText},
			"",
			[]string{indent(s.recordSpecAsText, 4)},
		},
		{
			"contract spec uuid exists record spec name does not",
			[]string{s.contractSpecUUID.String(), s.contractSpecUUID.String()},
			"",
			[]string{"specification: null", s.contractSpecUUID.String()},
		},
		{
			"record specs from contract spec uuid and only whitespace name args",
			[]string{s.contractSpecUUID.String(), "   ", " ", "      "},
			"",
			[]string{indent(s.recordSpecAsText, 4)},
		},
		{
			"bad prefix",
			[]string{"session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"},
			"rpc error: code = InvalidArgument desc = invalid specification id: address [session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr] is not a contract spec address: invalid request",
			[]string{},
		},
		{
			"bad arg 1",
			[]string{"not-gonna-parse"},
			"rpc error: code = InvalidArgument desc = invalid specification id: could not parse [not-gonna-parse] into either a contract spec address (decoding bech32 failed: invalid index of 1) or uuid (invalid UUID length: 15): invalid request",
			[]string{},
		},
		{
			"no args",
			[]string{s.asJson},
			"requires at least 1 arg(s), only received 0",
			[]string{},
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetOwnershipCmd() {
	cmd := cli.GetOwnershipCmd()

	newUser := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()

	paginationText := `pagination:
  next_key: null
  total: "1"
`
	scopeUUIDsText := fmt.Sprintf(`scope_uuids:
- %s`,
		s.scopeUUID,
	)

	testCases := []queryCmdTestCase{
		{
			"scopes as json",
			[]string{s.user1, s.asJson},
			"",
			[]string{
				fmt.Sprintf("\"scope_uuids\":[\"%s\"]", s.scopeUUID),
				"\"pagination\":{\"next_key\":null,\"total\":\"1\"}",
			},
		},
		{
			"scopes as text",
			[]string{s.user1, s.asText},
			"",
			[]string{scopeUUIDsText, paginationText},
		},
		{
			"scope through value owner",
			[]string{s.user2},
			"",
			[]string{scopeUUIDsText},
		},
		{
			"no result",
			[]string{newUser},
			"",
			[]string{"scope_uuids: []", "total: \"0\""},
		},
		{
			"two args",
			[]string{s.user1, s.user2},
			"accepts 1 arg(s), received 2",
			[]string{},
		},
		{
			"no args",
			[]string{},
			"accepts 1 arg(s), received 0",
			[]string{},
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetValueOwnershipCmd() {
	cmd := cli.GetValueOwnershipCmd()

	paginationText := `pagination:
  next_key: null
  total: "1"
`
	scopeUUIDsText := fmt.Sprintf(`scope_uuids:
- %s`,
		s.scopeUUID,
	)

	testCases := []queryCmdTestCase{
		{
			"as json",
			[]string{s.user2, s.asJson},
			"",
			[]string{
				fmt.Sprintf("\"scope_uuids\":[\"%s\"]", s.scopeUUID),
				"\"pagination\":{\"next_key\":null,\"total\":\"1\"}",
			},
		},
		{
			"as text",
			[]string{s.user2, s.asText},
			"",
			[]string{scopeUUIDsText, paginationText},
		},
		{
			"no result",
			[]string{s.user1},
			"",
			[]string{"scope_uuids: []", "total: \"0\""},
		},
		{
			"two args",
			[]string{s.user1, s.user2},
			"accepts 1 arg(s), received 2",
			[]string{},
		},
		{
			"no args",
			[]string{},
			"accepts 1 arg(s), received 0",
			[]string{},
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetOSLocatorCmd() {
	cmd := cli.GetOSLocatorCmd()

	testCases := []queryCmdTestCase{
		{
			"get os locator by owner",
			[]string{s.user1Addr.String(), s.asJson},
			"",
			[]string{
				fmt.Sprintf("\"owner\":\"%s\"", s.user1Addr.String()),
				fmt.Sprintf("\"locator_uri\":\"%s\"}", "http://foo.com"),
			},
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

// ---------- tx cmd tests ----------

type txCmdTestCase struct {
	name         string
	cmd          *cobra.Command
	args         []string
	expectErr    bool
	expectErrMsg string
	respType     proto.Message
	expectedCode uint32
}

func runTxCmdTestCases(s *IntegrationCLITestSuite, testCases []txCmdTestCase) {
	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(t *testing.T) {
			cmdName := tc.cmd.Name()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.cmd, tc.args)

			if len(tc.expectErrMsg) > 0 {
				require.EqualError(t, err, tc.expectErrMsg, "%s expected error message", cmdName)
			} else if tc.expectErr {
				require.Error(t, err, "%s expected error", cmdName)
			} else {
				require.NoError(t, err, "%s unexpected error", cmdName)

				umErr := clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType)
				require.NoError(t, umErr, "%s UnmarshalJSON error", cmdName)

				txResp := tc.respType.(*sdk.TxResponse)
				assert.Equal(t, tc.expectedCode, txResp.Code, "%s response code", cmdName)
				// Note: If the above is failing because a 0 is expected, but it's getting a 1,
				//       it might mean that the keeper method is returning an error.

				if t.Failed() {
					t.Logf("tx:\n%v\n", txResp)
				}
			}
		})
	}
}

func (s *IntegrationCLITestSuite) TestMetadataScopeTxCommands() {

	scopeID := metadatatypes.ScopeMetadataAddress(uuid.New()).String()
	specID := metadatatypes.ScopeSpecMetadataAddress(uuid.New()).String()
	testCases := []txCmdTestCase{
		{
			"should successfully add metadata scope",
			cli.WriteScopeCmd(),
			[]string{
				scopeID,
				specID,
				s.testnet.Validators[0].Address.String(),
				s.testnet.Validators[0].Address.String(),
				s.testnet.Validators[0].Address.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, "", &sdk.TxResponse{}, 0,
		},
		{
			"should successfully add metadata scope with signers flag",
			cli.WriteScopeCmd(),
			[]string{
				metadatatypes.ScopeMetadataAddress(uuid.New()).String(),
				metadatatypes.ScopeSpecMetadataAddress(uuid.New()).String(),
				s.user1,
				s.user1,
				s.user1,
				fmt.Sprintf("--%s=%s", cli.FlagSigners, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, "", &sdk.TxResponse{}, 0,
		},
		{
			"should fail to add metadata scope, incorrect scope id",
			cli.WriteScopeCmd(),
			[]string{
				"not-a-uuid",
				metadatatypes.ScopeSpecMetadataAddress(uuid.New()).String(),
				s.user1,
				s.user1,
				s.user1,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "decoding bech32 failed: invalid index of 1", &sdk.TxResponse{}, 0,
		},
		{
			"should fail to add metadata scope, incorrect scope spec id",
			cli.WriteScopeCmd(),
			[]string{
				metadatatypes.ScopeMetadataAddress(uuid.New()).String(),
				"not-a-uuid",
				s.user1,
				s.user1,
				s.user1,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "decoding bech32 failed: invalid index of 1", &sdk.TxResponse{}, 0,
		},
		{
			"should fail to add metadata scope, validate basic will err on owner format",
			cli.WriteScopeCmd(),
			[]string{
				metadatatypes.ScopeMetadataAddress(uuid.New()).String(),
				metadatatypes.ScopeSpecMetadataAddress(uuid.New()).String(),
				"incorrect,incorrect",
				s.user1,
				s.user1,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "invalid owner on scope: decoding bech32 failed: invalid index of 1", &sdk.TxResponse{}, 0,
		},
		{
			"should fail to remove metadata scope, invalid scopeid",
			cli.RemoveScopeCmd(),
			[]string{
				"not-valid",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "decoding bech32 failed: invalid index of 1", &sdk.TxResponse{}, 0,
		},
		{
			"should fail to add/remove metadata scope data access, invalid scopeid",
			cli.AddRemoveScopeDataAccessCmd(),
			[]string{
				"add",
				"not-valid",
				s.user2,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "decoding bech32 failed: invalid index of 1", &sdk.TxResponse{}, 0,
		},
		{
			"should fail to add/remove metadata scope data access, invalid command requires add or remove",
			cli.AddRemoveScopeDataAccessCmd(),
			[]string{
				"notaddorremove",
				scopeID,
				s.user2,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "incorrect command notaddorremove : required remove or update", &sdk.TxResponse{}, 0,
		},
		{
			"should fail to add/remove metadata scope data access, not a scope id",
			cli.AddRemoveScopeDataAccessCmd(),
			[]string{
				"add",
				specID,
				s.user2,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, fmt.Sprintf("meta address is not a scope: %s", specID), &sdk.TxResponse{}, 0,
		},
		{
			"should fail to add/remove metadata scope data access, validatebasic fails",
			cli.AddRemoveScopeDataAccessCmd(),
			[]string{
				"add",
				scopeID,
				"notauser",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "data access address is invalid: notauser", &sdk.TxResponse{}, 0,
		},
		{
			"should successfully add metadata scope data access",
			cli.AddRemoveScopeDataAccessCmd(),
			[]string{
				"add",
				scopeID,
				s.user1,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, "", &sdk.TxResponse{}, 0,
		},
		{
			"should successfully remove metadata scope data access",
			cli.AddRemoveScopeDataAccessCmd(),
			[]string{
				"remove",
				scopeID,
				s.user1,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, "", &sdk.TxResponse{}, 0,
		},
		{
			"should successfully remove metadata scope",
			cli.RemoveScopeCmd(),
			[]string{
				scopeID,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, "", &sdk.TxResponse{}, 0,
		},
		{
			"should fail to metadata scope that no longer exists",
			cli.RemoveScopeCmd(),
			[]string{
				scopeID,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, "", &sdk.TxResponse{}, 1,
		},
	}

	runTxCmdTestCases(s, testCases)
}

func (s *IntegrationCLITestSuite) TestScopeSpecificationTxCommands() {
	addCommand := cli.WriteScopeSpecificationCmd()
	removeCommand := cli.RemoveScopeSpecificationCmd()
	specID := metadatatypes.ScopeSpecMetadataAddress(uuid.New())
	testCases := []txCmdTestCase{
		{
			"should successfully add scope specification",
			addCommand,
			[]string{
				specID.String(),
				s.testnet.Validators[0].Address.String(),
				"owner",
				s.contractSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, "", &sdk.TxResponse{}, 0,
		},
		{
			"should successfully update scope specification with descriptions",
			addCommand,
			[]string{
				specID.String(),
				s.testnet.Validators[0].Address.String(),
				"owner",
				s.contractSpecID.String(),
				"description-name",
				"description",
				"http://www.blockchain.com/",
				"http://www.blockchain.com/icon.png",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, "", &sdk.TxResponse{}, 0,
		},
		{
			"should fail to add scope specification, invalid spec id format",
			addCommand,
			[]string{
				"invalid",
				s.testnet.Validators[0].Address.String(),
				"owner",
				s.contractSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "decoding bech32 failed: invalid bech32 string length 7", &sdk.TxResponse{}, 0,
		},
		{
			"should fail to add scope specification validate basic error",
			addCommand,
			[]string{
				specID.String(),
				s.testnet.Validators[0].Address.String(),
				"owner",
				specID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "invalid contract specification id prefix at index 0 (expected: contractspec, got scopespec)", &sdk.TxResponse{}, 0,
		},
		{
			"should fail to remove scope specification invalid id",
			removeCommand,
			[]string{
				"notvalid",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "decoding bech32 failed: invalid index of 1", &sdk.TxResponse{}, 0,
		},
		{
			"should successfully remove scope specification",
			removeCommand,
			[]string{
				specID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, "", &sdk.TxResponse{}, 0,
		},
		{
			"should fail to remove scope specification that has already been removed",
			removeCommand,
			[]string{
				specID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, "", &sdk.TxResponse{}, 1,
		},
	}

	runTxCmdTestCases(s, testCases)
}

func (s *IntegrationCLITestSuite) TestAddObjectLocatorCmd() {
	userURI := "http://foo.com"
	userURIMod := "https://www.google.com/search?q=red+butte+garden&oq=red+butte+garden&aqs=chrome..69i57j46i131i175i199i433j0j0i457j0l6.3834j0j7&sourceid=chrome&ie=UTF-8#lpqa=d,2"

	testCases := []txCmdTestCase{
		{
			"Should successfully add os locator",
			cli.BindOsLocatorCmd(),
			[]string{
				s.testnet.Validators[0].Address.String(),
				userURI,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, "", &sdk.TxResponse{}, 0,
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
			false, "", &sdk.TxResponse{}, 0,
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
			false, "", &sdk.TxResponse{}, 0,
		},
	}

	runTxCmdTestCases(s, testCases)
}

func (s *IntegrationCLITestSuite) TestContractSpecificationTxCommands() {
	addCommand := cli.WriteContractSpecificationCmd()
	removeCommand := cli.RemoveContractSpecificationCmd()
	contractSpecUUID := uuid.New()
	specificationID := metadatatypes.ContractSpecMetadataAddress(contractSpecUUID)
	testCases := []txCmdTestCase{
		{
			"should successfully add contract specification with resource hash",
			addCommand,
			[]string{
				specificationID.String(),
				s.testnet.Validators[0].Address.String(),
				"owner",
				"hashvalue",
				"`myclassname`",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"should successfully update contract specification with resource hash using signer flag",
			addCommand,
			[]string{
				specificationID.String(),
				s.testnet.Validators[0].Address.String(),
				"owner",
				"hashvalue",
				"`myclassname`",
				fmt.Sprintf("--%s=%s", cli.FlagSigners, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"should successfully update contract specification with resource id",
			addCommand,
			[]string{
				specificationID.String(),
				s.testnet.Validators[0].Address.String(),
				"owner",
				specificationID.String(),
				"myclassname",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"should successfully update contract specification with description",
			addCommand,
			[]string{
				specificationID.String(),
				s.testnet.Validators[0].Address.String(),
				"owner",
				"hashvalue",
				"myclassname",
				"description-name",
				"description",
				"http://www.blockchain.com/",
				"http://www.blockchain.com/icon.png",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"should successfully remove contract specification",
			removeCommand,
			[]string{
				specificationID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"should fail to add contract specification on validate basic error",
			addCommand,
			[]string{
				"invalid-spec-id",
				s.testnet.Validators[0].Address.String(),
				"owner",
				"hashvalue",
				"myclassname",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true,
			"decoding bech32 failed: invalid index of 1",
			&sdk.TxResponse{},
			0,
		},
		{
			"should fail to remove contract specification invalid address",
			removeCommand,
			[]string{
				"not-a-id",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true,
			"decoding bech32 failed: invalid index of 1",
			&sdk.TxResponse{},
			0,
		},
		{
			"should fail to remove contract that no longer exists",
			removeCommand,
			[]string{
				specificationID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			"",
			&sdk.TxResponse{},
			1,
		},
	}

	runTxCmdTestCases(s, testCases)
}

func (s *IntegrationCLITestSuite) TestRecordSpecificationTxCommands() {
	cmd := cli.WriteRecordSpecificationCmd()
	addConractSpecCmd := cli.WriteContractSpecificationCmd()
	deleteRecordSpecCmd := cli.RemoveRecordSpecificationCmd()
	recordName := "testrecordspecid"
	contractSpecUUID := uuid.New()
	contractSpecID := metadatatypes.ContractSpecMetadataAddress(contractSpecUUID)
	specificationID := metadatatypes.RecordSpecMetadataAddress(contractSpecUUID, recordName)
	testCases := []txCmdTestCase{
		{
			"setup test with a record specification owned by signer",
			addConractSpecCmd,
			[]string{
				contractSpecID.String(),
				s.testnet.Validators[0].Address.String(),
				"owner",
				"hashvalue",
				"`myclassname`",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"should successfully add record specification",
			cmd,
			[]string{
				specificationID.String(),
				recordName,
				"record1,typename1,hashvalue",
				"typename",
				"record",
				"responsibleparties",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			"",
			&sdk.TxResponse{}, 0,
		},
		{
			"should successfully add record specification",
			cmd,
			[]string{
				specificationID.String(),
				recordName,
				"record1,typename1,hashvalue",
				"typename",
				"record_list",
				"responsibleparties",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			"",
			&sdk.TxResponse{}, 0,
		},
		{
			"should fail to add record specification, validate basic fail",
			cmd,
			[]string{
				specificationID.String(),
				"",
				"record1,typename1,hashvalue;record2,typename2,recspec1q5p7xh9vtktyc9ynp25ydq4cycqp3tp7wdplq95fp3gsaylex5npzlhnhp6",
				"typename",
				"record",
				"responsibleparties",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true,
			"record specification name cannot be empty",
			&sdk.TxResponse{}, 0,
		},
		{
			"should fail to add record specification, fail parsing inputs too few values",
			cmd,
			[]string{
				specificationID.String(),
				recordName,
				"record1,typename1;record2,typename2,hashvalue",
				"typename",
				"record",
				"responsibleparties",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true,
			"invalid number of values for input specification: 2",
			&sdk.TxResponse{}, 0,
		},
		{
			"should fail to add record specification, incorrect result type",
			cmd,
			[]string{
				specificationID.String(),
				recordName,
				"record1,typename1,hashvalue",
				"typename",
				"incorrect",
				"responsibleparties",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true,
			"record specification result type cannot be unspecified",
			&sdk.TxResponse{}, 0,
		},
		{
			"should fail to add record specification, incorrect signer format",
			cmd,
			[]string{
				specificationID.String(),
				recordName,
				"record1,typename1,hashvalue",
				"typename",
				"record",
				"responsibleparties",
				fmt.Sprintf("--%s=%s", cli.FlagSigners, "incorrect-signer-format"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true,
			"decoding bech32 failed: invalid index of 1",
			&sdk.TxResponse{}, 0,
		},
		{
			"should fail to delete record specification, incorrect id",
			deleteRecordSpecCmd,
			[]string{
				"incorrect-id",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true,
			"decoding bech32 failed: invalid index of 1",
			&sdk.TxResponse{}, 0,
		},
		{
			"should fail to delete record specification, not a record specification",
			deleteRecordSpecCmd,
			[]string{
				contractSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true,
			fmt.Sprintf("invalid contract specification id: %v", contractSpecID.String()),
			&sdk.TxResponse{}, 0,
		},
		{
			"should successfully delete record specification",
			deleteRecordSpecCmd,
			[]string{
				specificationID.String(),
				fmt.Sprintf("--%s=%s", cli.FlagSigners, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			"",
			&sdk.TxResponse{}, 0,
		},
		{
			"should fail to delete record specification that does not exist",
			deleteRecordSpecCmd,
			[]string{
				specificationID.String(),
				fmt.Sprintf("--%s=%s", cli.FlagSigners, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			"",
			&sdk.TxResponse{}, 1,
		},
	}

	runTxCmdTestCases(s, testCases)
}

func (s *IntegrationCLITestSuite) TestRecordTxCommands() {
	userAddress := s.testnet.Validators[0].Address.String()
	addRecordCmd := cli.WriteRecordCmd()
	scopeSpecID := metadatatypes.ScopeSpecMetadataAddress(uuid.New())
	scopeUUID := uuid.New()
	scopeID := metadatatypes.ScopeMetadataAddress(scopeUUID)
	contractSpecUUID := uuid.New()
	contractSpecName := "`myclassname`"
	contractSpecID := metadatatypes.ContractSpecMetadataAddress(contractSpecUUID)

	recordName := "recordnamefortests"
	recSpecID := metadatatypes.RecordSpecMetadataAddress(contractSpecUUID, recordName)

	recordId := metadatatypes.RecordMetadataAddress(scopeUUID, recordName)

	testCases := []txCmdTestCase{
		{
			"should successfully add scope specification for test setup",
			cli.WriteScopeSpecificationCmd(),
			[]string{
				scopeSpecID.String(),
				userAddress,
				"owner,originator",
				s.contractSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, userAddress),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, "", &sdk.TxResponse{}, 0,
		},
		{
			"should successfully add metadata scope for test setup",
			cli.WriteScopeCmd(),
			[]string{
				scopeID.String(),
				scopeSpecID.String(),
				userAddress,
				userAddress,
				userAddress,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, userAddress),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, "", &sdk.TxResponse{}, 0,
		},
		{
			"should successfully add contract specification with resource hash for test setup",
			cli.WriteContractSpecificationCmd(),
			[]string{
				contractSpecID.String(),
				userAddress,
				"owner,originator",
				"hashvalue",
				contractSpecName,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, userAddress),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"should successfully add record specification for test setup",
			cli.WriteRecordSpecificationCmd(),
			[]string{
				recSpecID.String(),
				recordName,
				"input1name,typename1,hashvalue",
				"typename",
				"record",
				"responsibleparties",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			"",
			&sdk.TxResponse{}, 0,
		},
		{
			"should successfully add record with and create new session",
			addRecordCmd,
			[]string{
				scopeID.String(),
				recSpecID.String(),
				recordName,
				"processname,hashvalue,methodname",
				"input1name,hashvalue,typename1,proposed",
				"outputhashvalue,pass",
				fmt.Sprintf("%s,owner;%s,originator", userAddress, userAddress),
				contractSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"should fail to add record incorrect scope id format",
			addRecordCmd,
			[]string{
				"not-a-scope-id",
				recSpecID.String(),
				recordName,
				"processname,hashvalue,methodname",
				"input1name,hashvalue,typename1,proposed",
				"outputhashvalue,pass",
				fmt.Sprintf("%s,owner;%s,originator", userAddress, userAddress),
				contractSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "decoding bech32 failed: invalid index of 1",
			&sdk.TxResponse{},
			0,
		},
		{
			"should fail to add record incorrect record id format",
			addRecordCmd,
			[]string{
				scopeID.String(),
				"not-a-record-id",
				recordName,
				"processname,hashvalue,methodname",
				"input1name,hashvalue,typename1,proposed",
				"outputhashvalue,pass",
				fmt.Sprintf("%s,owner;%s,originator", userAddress, userAddress),
				contractSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "decoding bech32 failed: invalid index of 1",
			&sdk.TxResponse{},
			0,
		},
		{
			"should fail to add record incorrect process format",
			addRecordCmd,
			[]string{
				scopeID.String(),
				recSpecID.String(),
				recordName,
				"hashvalue,methodname",
				"input1name,hashvalue,typename1,proposed",
				"outputhashvalue,pass",
				fmt.Sprintf("%s,owner;%s,originator", userAddress, userAddress),
				contractSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "invalid number of values for process: 2",
			&sdk.TxResponse{},
			0,
		},
		{
			"should fail to add record incorrect record inputs format",
			addRecordCmd,
			[]string{
				scopeID.String(),
				recSpecID.String(),
				recordName,
				"processname,hashvalue,methodname",
				"input1name,typename1,proposed",
				"outputhashvalue,pass",
				fmt.Sprintf("%s,owner;%s,originator", userAddress, userAddress),
				contractSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "invalid number of values for record input: 3",
			&sdk.TxResponse{},
			0,
		},
		{
			"should fail to add record incorrect record output format",
			addRecordCmd,
			[]string{
				scopeID.String(),
				recSpecID.String(),
				recordName,
				"processname,hashvalue,methodname",
				"input1name,hashvalue,typename1,proposed",
				"outputhashvalue",
				fmt.Sprintf("%s,owner;%s,originator", userAddress, userAddress),
				contractSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "invalid number of values for record output: 1",
			&sdk.TxResponse{},
			0,
		},
		{
			"should fail to add record incorrect parties involved format",
			addRecordCmd,
			[]string{
				scopeID.String(),
				recSpecID.String(),
				recordName,
				"processname,hashvalue,methodname",
				"input1name,hashvalue,typename1,proposed",
				"outputhashvalue,pass",
				fmt.Sprintf("%s;%s", userAddress, userAddress),
				contractSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "invalid number of values for parties: 1",
			&sdk.TxResponse{},
			0,
		},
		{
			"should fail to add record incorrect contract or session id format",
			addRecordCmd,
			[]string{
				scopeID.String(),
				recSpecID.String(),
				recordName,
				"processname,hashvalue,methodname",
				"input1name,hashvalue,typename1,proposed",
				"outputhashvalue,pass",
				fmt.Sprintf("%s,owner;%s,originator", userAddress, userAddress),
				scopeID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, fmt.Sprintf("id must be a contract or session id: %s", scopeID.String()),
			&sdk.TxResponse{},
			0,
		},
		{
			"should successfully remove record",
			cli.RemoveRecordCmd(),
			[]string{
				recordId.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, "",
			&sdk.TxResponse{},
			0,
		},
	}
	runTxCmdTestCases(s, testCases)
}

func (s *IntegrationCLITestSuite) TestWriteSessionCmd() {
	cmd := cli.WriteSessionCmd()

	owner := s.testnet.Validators[0].Address.String()
	sender := s.testnet.Validators[0].Address.String()
	scopeUUID := uuid.New()
	scopeID := metadatatypes.ScopeMetadataAddress(scopeUUID)

	writeScopeCmd := cli.WriteScopeCmd()
	ctx := s.testnet.Validators[0].ClientCtx
	out, err := clitestutil.ExecTestCLICmd(
		ctx,
		writeScopeCmd,
		[]string{
			scopeID.String(),
			s.scopeSpecID.String(),
			owner,
			owner,
			owner,
			fmt.Sprintf("--%s=%s", flags.FlagFrom, sender),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	require.NoError(s.T(), err, "adding base scope")
	scopeResp := sdk.TxResponse{}
	umErr := ctx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &scopeResp)
	require.NoError(s.T(), umErr, "%s UnmarshalJSON error", writeScopeCmd.Name())
	if scopeResp.Code != 0 {
		s.T().Logf("write-scope response code is not 0.\ntx response:\n%v\n", scopeResp)
		s.T().FailNow()
	}

	testCases := []txCmdTestCase{
		{
			"session-id no context",
			cmd,
			[]string{
				metadatatypes.SessionMetadataAddress(scopeUUID, uuid.New()).String(),
				s.contractSpecID.String(), fmt.Sprintf("%s,owner", owner), "somename",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, sender),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"scope-id session-uuid no context",
			cmd,
			[]string{
				scopeID.String(),
				uuid.New().String(),
				s.contractSpecID.String(), fmt.Sprintf("%s,owner", owner), "somename",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, sender),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"scope-uuid session-uuid no context",
			cmd,
			[]string{
				scopeUUID.String(),
				uuid.New().String(),
				s.contractSpecID.String(), fmt.Sprintf("%s,owner", owner), "somename",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, sender),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"session-id with context",
			cmd,
			[]string{
				metadatatypes.SessionMetadataAddress(scopeUUID, uuid.New()).String(),
				s.contractSpecID.String(), fmt.Sprintf("%s,owner", owner), "somename",
				"ChFIRUxMTyBQUk9WRU5BTkNFIQ==",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, sender),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"scope-id session-uuid with context",
			cmd,
			[]string{
				scopeID.String(),
				uuid.New().String(),
				s.contractSpecID.String(), fmt.Sprintf("%s,owner", owner), "somename",
				"ChFIRUxMTyBQUk9WRU5BTkNFIQ==",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, sender),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"scope-uuid session-uuid with context",
			cmd,
			[]string{
				scopeUUID.String(),
				uuid.New().String(),
				s.contractSpecID.String(), fmt.Sprintf("%s,owner", owner), "somename",
				"ChFIRUxMTyBQUk9WRU5BTkNFIQ==",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, sender),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"wrong id type",
			cmd,
			[]string{
				s.scopeSpecID.String(),
				s.contractSpecID.String(), fmt.Sprintf("%s,owner", owner), "somename",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, sender),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true,
			fmt.Sprintf("invalid address type in argument [%s]", s.scopeSpecID),
			&sdk.TxResponse{},
			0,
		},
		{
			"invalid first argument",
			cmd,
			[]string{
				"invalid",
				s.contractSpecID.String(), fmt.Sprintf("%s,owner", owner), "somename",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, sender),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true,
			fmt.Sprintf("argument [%s] is neither a bech32 address (%s) nor UUID (%s)", "invalid", "decoding bech32 failed: invalid bech32 string length 7", "invalid UUID length: 7"),
			&sdk.TxResponse{},
			0,
		},
		{
			"session-id with different context",
			cmd,
			[]string{
				metadatatypes.SessionMetadataAddress(scopeUUID, uuid.New()).String(),
				s.contractSpecID.String(), fmt.Sprintf("%s,owner", owner), "somename",
				"SEVMTE8gUFJPVkVOQU5DRSEK",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, sender),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
	}

	runTxCmdTestCases(s, testCases)
}
