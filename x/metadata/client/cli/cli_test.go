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

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzcli "github.com/cosmos/cosmos-sdk/x/authz/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/x/metadata/client/cli"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
)

type IntegrationCLITestSuite struct {
	suite.Suite

	cfg             testnet.Config
	testnet         *testnet.Network
	keyring         keyring.Keyring
	keyringDir      string
	keyringAccounts []keyring.Record

	asJson         string
	asText         string
	includeRequest string

	accountAddr    sdk.AccAddress
	accountAddrStr string

	user1Addr    sdk.AccAddress
	user1AddrStr string

	user2Addr    sdk.AccAddress
	user2AddrStr string

	user3Addr    sdk.AccAddress
	user3AddrStr string

	userOtherAddr sdk.AccAddress
	userOtherStr  string

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

	objectLocator1 metadatatypes.ObjectStoreLocator
	ownerAddr1     sdk.AccAddress
	encryptionKey1 sdk.AccAddress
	uri1           string

	objectLocator1AsText string
	objectLocator1AsJson string

	objectLocator2 metadatatypes.ObjectStoreLocator
	ownerAddr2     sdk.AccAddress
	encryptionKey2 sdk.AccAddress
	uri2           string

	objectLocator2AsText string
	objectLocator2AsJson string
}

func TestIntegrationCLITestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationCLITestSuite))
}

func (s *IntegrationCLITestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	pioconfig.SetProvenanceConfig("atom", 0)
	cfg := testutil.DefaultTestNetworkConfig()
	cfg.NumValidators = 1
	genesisState := cfg.GenesisState
	s.cfg = cfg
	s.generateAccountsWithKeyrings(4)

	var err error
	// An account
	s.accountAddr, err = s.keyringAccounts[0].GetAddress()
	s.Require().NoError(err, "getting keyringAccounts[0] address")
	s.accountAddrStr = s.accountAddr.String()

	// A user account
	s.user1Addr, err = s.keyringAccounts[1].GetAddress()
	s.Require().NoError(err, "getting keyringAccounts[1] address")
	s.user1AddrStr = s.user1Addr.String()

	// A second user account
	s.user2Addr, err = s.keyringAccounts[2].GetAddress()
	s.Require().NoError(err, "getting keyringAccounts[2] address")
	s.user2AddrStr = s.user2Addr.String()

	// A third user account
	s.user3Addr, err = s.keyringAccounts[3].GetAddress()
	s.Require().NoError(err, "getting keyringAccounts[3] address")
	s.user3AddrStr = s.user3Addr.String()

	// An account that isn't known
	s.userOtherAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	s.userOtherStr = s.userOtherAddr.String()

	// Configure Genesis auth data for adding test accounts
	var genAccounts []authtypes.GenesisAccount
	var authData authtypes.GenesisState
	authData.Params = authtypes.DefaultParams()
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.accountAddr, nil, 3, 0))
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.user1Addr, nil, 4, 1))
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.user2Addr, nil, 5, 1))
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.user3Addr, nil, 6, 0))
	accounts, err := authtypes.PackAccounts(genAccounts)
	s.Require().NoError(err)
	authData.Accounts = accounts
	authDataBz, err := cfg.Codec.MarshalJSON(&authData)
	s.Require().NoError(err)
	genesisState[authtypes.ModuleName] = authDataBz

	// Configure Genesis bank data for test accounts
	var genBalances []banktypes.Balance
	genBalances = append(genBalances, banktypes.Balance{Address: s.accountAddrStr, Coins: sdk.NewCoins(
		sdk.NewCoin(cfg.BondDenom, cfg.StakingTokens),
		sdk.NewCoin("authzhotdog", sdk.NewInt(100)),
	).Sort()})
	genBalances = append(genBalances, banktypes.Balance{Address: s.user1AddrStr, Coins: sdk.NewCoins(
		sdk.NewCoin(cfg.BondDenom, cfg.StakingTokens),
		sdk.NewCoin("authzhotdog", sdk.NewInt(100)),
	).Sort()})
	genBalances = append(genBalances, banktypes.Balance{Address: s.user2AddrStr, Coins: sdk.NewCoins(
		sdk.NewCoin(cfg.BondDenom, cfg.StakingTokens),
	).Sort()})
	genBalances = append(genBalances, banktypes.Balance{Address: s.user3AddrStr, Coins: sdk.NewCoins(
		sdk.NewCoin(cfg.BondDenom, cfg.StakingTokens),
	).Sort()})
	var bankGenState banktypes.GenesisState
	bankGenState.Params = banktypes.DefaultParams()
	bankGenState.Balances = genBalances
	bankDataBz, err := cfg.Codec.MarshalJSON(&bankGenState)
	s.Require().NoError(err)
	genesisState[banktypes.ModuleName] = bankDataBz

	s.asJson = fmt.Sprintf("--%s=json", tmcli.OutputFlag)
	s.asText = fmt.Sprintf("--%s=text", tmcli.OutputFlag)
	s.includeRequest = "--include-request"

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

	s.scope = metadatatypes.Scope{
		ScopeId:           s.scopeID,
		SpecificationId:   s.scopeSpecID,
		Owners:            ownerPartyList(s.user1AddrStr),
		DataAccess:        []string{s.user1AddrStr},
		ValueOwnerAddress: s.user2AddrStr,
	}

	s.session = *metadatatypes.NewSession(
		"unit test session",
		s.sessionID,
		s.contractSpecID,
		ownerPartyList(s.user1AddrStr),
		&metadatatypes.AuditFields{
			CreatedDate: time.Time{},
			CreatedBy:   s.user1AddrStr,
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
		[]string{s.user1AddrStr},
		[]metadatatypes.PartyType{metadatatypes.PartyType_PARTY_TYPE_OWNER},
		[]metadatatypes.MetadataAddress{s.contractSpecID},
	)

	s.contractSpec = *metadatatypes.NewContractSpecification(
		s.contractSpecID,
		nil,
		[]string{s.user1AddrStr},
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

	s.scopeAsJson = fmt.Sprintf("{\"scope_id\":\"%s\",\"specification_id\":\"%s\",\"owners\":[{\"address\":\"%s\",\"role\":\"PARTY_TYPE_OWNER\",\"optional\":false}],\"data_access\":[\"%s\"],\"value_owner_address\":\"%s\",\"require_party_rollup\":false}",
		s.scopeID,
		s.scopeSpecID,
		s.user1AddrStr,
		s.user1AddrStr,
		s.user2AddrStr,
	)
	s.scopeAsText = fmt.Sprintf(`data_access:
- %s
owners:
- address: %s
  optional: false
  role: PARTY_TYPE_OWNER
require_party_rollup: false
scope_id: %s
specification_id: %s
value_owner_address: %s`,
		s.user1AddrStr,
		s.user1AddrStr,
		s.scopeID,
		s.scopeSpecID,
		s.user2AddrStr,
	)

	s.sessionAsJson = fmt.Sprintf("{\"session_id\":\"%s\",\"specification_id\":\"%s\",\"parties\":[{\"address\":\"%s\",\"role\":\"PARTY_TYPE_OWNER\",\"optional\":false}],\"name\":\"unit test session\",\"context\":null,\"audit\":{\"created_date\":\"0001-01-01T00:00:00Z\",\"created_by\":\"%s\",\"updated_date\":\"0001-01-01T00:00:00Z\",\"updated_by\":\"\",\"version\":0,\"message\":\"unit testing\"}}",
		s.sessionID,
		s.contractSpecID,
		s.user1AddrStr,
		s.user1AddrStr,
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
  optional: false
  role: PARTY_TYPE_OWNER
session_id: %s
specification_id: %s`,
		s.user1AddrStr,
		s.user1AddrStr,
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
		s.user1AddrStr,
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
		s.user1AddrStr,
		s.scopeSpecID,
	)

	s.contractSpecAsJson = fmt.Sprintf("{\"specification_id\":\"%s\",\"description\":null,\"owner_addresses\":[\"%s\"],\"parties_involved\":[\"PARTY_TYPE_OWNER\"],\"hash\":\"notreallyasourcehash\",\"class_name\":\"contractclassname\"}",
		s.contractSpecID,
		s.user1AddrStr,
	)
	s.contractSpecAsText = fmt.Sprintf(`class_name: contractclassname
description: null
hash: notreallyasourcehash
owner_addresses:
- %s
parties_involved:
- PARTY_TYPE_OWNER
specification_id: %s`,
		s.user1AddrStr,
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
	locAsText := func(loc metadatatypes.ObjectStoreLocator) string {
		eKey := loc.EncryptionKey
		if len(eKey) == 0 {
			eKey = "\"\""
		}
		return fmt.Sprintf(`encryption_key: %s
locator_uri: %s
owner: %s`,
			eKey,
			loc.LocatorUri,
			loc.Owner,
		)
	}
	locAsJson := func(loc metadatatypes.ObjectStoreLocator) string {
		return fmt.Sprintf("{\"owner\":\"%s\",\"locator_uri\":\"%s\",\"encryption_key\":\"%s\"}",
			loc.Owner,
			loc.LocatorUri,
			loc.EncryptionKey,
		)
	}
	s.ownerAddr1 = s.user1Addr
	s.encryptionKey1 = sdk.AccAddress{}
	s.uri1 = "http://foo.com"
	s.objectLocator1 = metadatatypes.NewOSLocatorRecord(s.ownerAddr1, s.encryptionKey1, s.uri1)
	s.objectLocator1AsText = locAsText(s.objectLocator1)
	s.objectLocator1AsJson = locAsJson(s.objectLocator1)

	s.ownerAddr2 = s.user2Addr
	s.encryptionKey2 = s.user1Addr
	s.uri2 = "http://bar.com"
	s.objectLocator2 = metadatatypes.NewOSLocatorRecord(s.ownerAddr2, s.encryptionKey2, s.uri2)
	s.objectLocator2AsText = locAsText(s.objectLocator2)
	s.objectLocator2AsJson = locAsJson(s.objectLocator2)

	var metadataData metadatatypes.GenesisState
	s.Require().NoError(cfg.Codec.UnmarshalJSON(genesisState[metadatatypes.ModuleName], &metadataData))
	metadataData.Scopes = append(metadataData.Scopes, s.scope)
	metadataData.Sessions = append(metadataData.Sessions, s.session)
	metadataData.Records = append(metadataData.Records, s.record)
	metadataData.ScopeSpecifications = append(metadataData.ScopeSpecifications, s.scopeSpec)
	metadataData.ContractSpecifications = append(metadataData.ContractSpecifications, s.contractSpec)
	metadataData.RecordSpecifications = append(metadataData.RecordSpecifications, s.recordSpec)
	metadataData.ObjectStoreLocators = append(metadataData.ObjectStoreLocators, s.objectLocator1, s.objectLocator2)
	metadataDataBz, err := cfg.Codec.MarshalJSON(&metadataData)
	s.Require().NoError(err)
	genesisState[metadatatypes.ModuleName] = metadataDataBz

	cfg.GenesisState = genesisState

	cfg.ChainID = antewrapper.SimAppChainID
	s.testnet, err = testnet.New(s.T(), s.T().TempDir(), cfg)
	s.Require().NoError(err, "creating testnet")

	_, err = s.testnet.WaitForHeight(1)
	s.Require().NoError(err, "waiting for height 1")
}

func (s *IntegrationCLITestSuite) TearDownSuite() {
	testutil.CleanUp(s.testnet, s.T())
}

func (s *IntegrationCLITestSuite) generateAccountsWithKeyrings(number int) {
	path := hd.CreateHDPath(118, 0, 0).String()
	s.keyringDir = s.T().TempDir()
	kr, err := keyring.New(s.T().Name(), "test", s.keyringDir, nil, s.cfg.Codec)
	s.Require().NoError(err, "keyring creation")
	s.keyring = kr
	for i := 0; i < number; i++ {
		keyId := fmt.Sprintf("test_key%v", i)
		info, _, err := kr.NewMnemonic(keyId, keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		s.Require().NoError(err, "key creation")
		s.keyringAccounts = append(s.keyringAccounts, *info)
	}
}

func ownerPartyList(addresses ...string) []metadatatypes.Party {
	retval := make([]metadatatypes.Party, len(addresses))
	for i, addr := range addresses {
		retval[i] = metadatatypes.Party{Address: addr, Role: metadatatypes.PartyType_PARTY_TYPE_OWNER}
	}
	return retval
}

func indent(str string, spaces int) string {
	var sb strings.Builder
	lines := strings.Split(str, "\n")
	maxI := len(lines) - 1
	s := strings.Repeat(" ", spaces)
	for i, l := range lines {
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

func alternateCase(str string, startUpper bool) string {
	// A-Z -> 65-90
	// a-z -> 97-122
	ms := 0 // aka modShift
	if startUpper {
		ms = 1
	}
	var r strings.Builder
	for i, c := range str {
		switch {
		case (i+ms)%2 == 0 && c >= 65 && c <= 90:
			r.WriteByte(byte(c + 32))
		case (i+ms)%2 == 1 && c >= 97 && c <= 122:
			r.WriteByte(byte(c - 32))
		default:
			r.WriteByte(byte(c))
		}
	}
	return r.String()
}

func (s *IntegrationCLITestSuite) getClientCtx() client.Context {
	return s.getClientCtxWithoutKeyring().WithKeyringDir(s.keyringDir).WithKeyring(s.keyring)
}

func (s *IntegrationCLITestSuite) getClientCtxWithoutKeyring() client.Context {
	return s.testnet.Validators[0].ClientCtx
}

// ---------- query cmd tests ----------

type queryCmdTestCase struct {
	name             string
	args             []string
	expectedError    string
	expectedInOutput []string
}

func runQueryCmdTestCases(s *IntegrationCLITestSuite, cmdGen func() *cobra.Command, testCases []queryCmdTestCase) {
	// Providing the command using a generator (cmdGen), we get a new instance of the cmd each time, and the flags won't
	// carry over between tests on the same command.
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			clientCtx := s.getClientCtxWithoutKeyring()
			cmd := cmdGen()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if len(tc.expectedError) > 0 {
				actualError := ""
				if err != nil {
					actualError = err.Error()
				}
				require.Contains(t, actualError, tc.expectedError, "expected error")
				// Something deep down is double wrapping the errors.
				// E.g. "foo: invalid request" has become
				// "foo: invalid request"
				// So we changed from the "Equal" test below to the "Contains" test above.
				// If you're bored, maybe try swapping back to see if things have been fixed.
				//require.Equal(t, tc.expectedError, actualError, "expected error")
			} else {
				require.NoErrorf(t, err, "unexpected error: %s", err)
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
	cmd := func() *cobra.Command { return cli.GetMetadataParamsCmd() }

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
	cmd := func() *cobra.Command { return cli.GetMetadataByIDCmd() }

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
			"decoding bech32 failed: invalid checksum (expected kzwk8c got xlkwel)",
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
			"decoding bech32 failed: invalid separator index 32",
			[]string{},
		},
		{
			"get metadata by id - bad arg",
			[]string{"not-an-id"},
			"decoding bech32 failed: invalid separator index -1",
			[]string{},
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetMetadataGetAllCmd() {
	cmd := func() *cobra.Command { return cli.GetMetadataGetAllCmd() }

	indentedScopeText := indent(s.scopeAsText, 4)
	indentedSessionText := indent(s.sessionAsText, 4)
	indentedRecordText := indent(s.recordAsText, 4)
	indentedScopeSpecText := indent(s.scopeSpecAsText, 4)
	indentedContractSpecText := indent(s.contractSpecAsText, 4)
	indentedRecordSpecText := indent(s.recordSpecAsText, 4)
	indentedOSLoc1Text := yamlListEntry(s.objectLocator1AsText)
	indentedOSLoc2Text := yamlListEntry(s.objectLocator2AsText)

	testCases := []queryCmdTestCase{}

	testName := func(base string, basei int, namei int, name string, suffix string) string {
		return fmt.Sprintf("all %s %03d %s %s", base, basei*4+namei+1, name, suffix)
	}
	addTestCases := func(bases []string, asText []string, asJson []string) {
		for bi, base := range bases {
			for ni, name := range []string{base, strings.ToUpper(base), alternateCase(base, true), alternateCase(base, false)} {
				testCases = append(testCases,
					queryCmdTestCase{
						testName(bases[0], bi, ni, name, "as text"),
						[]string{name, s.asText},
						"",
						asText,
					},
					queryCmdTestCase{
						testName(bases[0], bi, ni, name, "as json"),
						[]string{name, s.asJson},
						"",
						asJson,
					},
				)
			}
		}
	}
	makeSpecInputs := func(bases ...string) []string {
		r := make([]string, 0, len(bases)*8)
		for _, b := range bases {
			for _, e := range []string{"specs", "spec", "specifications", "specification"} {
				for _, d := range []string{"", "-", " "} {
					r = append(r, b+d+e)
				}
			}
		}
		return r
	}

	addTestCases([]string{"scopes", "scope"}, []string{indentedScopeText}, []string{s.scopeAsJson})
	addTestCases([]string{"sessions", "session", "sess"}, []string{indentedSessionText}, []string{s.sessionAsJson})
	addTestCases([]string{"records", "record", "recs", "rec"}, []string{indentedRecordText}, []string{s.recordAsJson})

	addTestCases(makeSpecInputs("scope"), []string{indentedScopeSpecText}, []string{s.scopeSpecAsJson})
	testCases = append(testCases,
		queryCmdTestCase{
			"all scopespecs spaced args 1 scope specs as text",
			[]string{"scope", "specs", s.asText},
			"",
			[]string{indentedScopeSpecText},
		},
		queryCmdTestCase{
			"all scopespecs spaced args 2 scope specification as json",
			[]string{"scope", "specification", s.asJson},
			"",
			[]string{s.scopeSpecAsJson},
		},
		queryCmdTestCase{
			"all scopespecs spaced args 3 scop espec as json",
			[]string{"scop", "espec", s.asJson},
			"",
			[]string{s.scopeSpecAsJson},
		},
	)

	cSpecInputs := makeSpecInputs("contract")
	cSpecInputs = append(cSpecInputs, "cspecs", "cspec", "c-specs", "c-spec", "c specs", "c spec")
	addTestCases(cSpecInputs, []string{indentedContractSpecText}, []string{s.contractSpecAsJson})
	testCases = append(testCases,
		queryCmdTestCase{
			"all contractspecs spaced args 1 contract specs as text",
			[]string{"contract", "specs", s.asText},
			"",
			[]string{indentedContractSpecText},
		},
		queryCmdTestCase{
			"all contractspecs spaced args 2 contract specification as json",
			[]string{"contract", "specification", s.asJson},
			"",
			[]string{s.contractSpecAsJson},
		},
		queryCmdTestCase{
			"all contractspecs spaced args 3 cs pec as json",
			[]string{"cs", "pec", s.asJson},
			"",
			[]string{s.contractSpecAsJson},
		},
	)

	addTestCases(makeSpecInputs("record", "rec"), []string{indentedRecordSpecText}, []string{s.recordSpecAsJson})
	testCases = append(testCases,
		queryCmdTestCase{
			"all recordspecs spaced args 1 record specs as text",
			[]string{"record", "specs", s.asText},
			"",
			[]string{indentedRecordSpecText},
		},
		queryCmdTestCase{
			"all recordspecs spaced args 2 record specification as json",
			[]string{"record", "specification", s.asJson},
			"",
			[]string{s.recordSpecAsJson},
		},
		queryCmdTestCase{
			"all recordspecs spaced args 3 recor dspec as json",
			[]string{"recor", "dspec", s.asJson},
			"",
			[]string{s.recordSpecAsJson},
		},
	)

	addTestCases(
		[]string{"locators", "locator", "locs", "loc"},
		[]string{indentedOSLoc1Text, indentedOSLoc2Text},
		[]string{s.objectLocator1AsJson, s.objectLocator2AsJson},
	)

	testCases = append(testCases,
		queryCmdTestCase{
			"unknown type",
			[]string{"scoops"},
			"unknown entry type: scoops",
			[]string{},
		},
		queryCmdTestCase{
			"unknown type many args",
			[]string{"r", "e", "d", "o", "r", "k", "   ", "sp", "o", "rk"},
			"unknown entry type: redorksporks",
			[]string{},
		},
		queryCmdTestCase{
			"no args",
			[]string{},
			"requires at least 1 arg(s), only received 0",
			[]string{},
		},
	)

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetMetadataScopeCmd() {
	cmd := func() *cobra.Command { return cli.GetMetadataScopeCmd() }

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
			"could not parse [not-a-valid-arg] into either a scope address (decoding bech32 failed: invalid separator index -1) or uuid (invalid UUID length: 15): invalid request",
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
	cmd := func() *cobra.Command { return cli.GetMetadataSessionCmd() }

	indentedSessionText := indent(s.sessionAsText, 4)
	notAUsedUUID := uuid.New()

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
			"session from scope uuid and session uuid as text",
			[]string{s.scopeUUID.String(), s.sessionUUID.String(), s.asText},
			"",
			[]string{indentedSessionText},
		},
		{
			"scope uuid and session uuid but scope does not exist",
			[]string{notAUsedUUID.String(), s.sessionUUID.String()},
			"",
			[]string{"session:", "session: null"},
		},
		{
			"scope uuid and session uuid and scope exists but session uuid does not exist",
			[]string{s.scopeUUID.String(), "5803f8bc-6067-4eb5-951f-2121671c2ec0"},
			"",
			[]string{"session:", "session: null"},
		},
		{
			"session from scope id and session uuid as text",
			[]string{s.scopeID.String(), s.sessionUUID.String(), s.asText},
			"",
			[]string{indentedSessionText},
		},
		{
			"session from scope id and session uuid as json",
			[]string{s.scopeID.String(), s.sessionUUID.String(), s.asJson},
			"",
			[]string{s.sessionAsJson},
		},
		{
			"scope id and session uuid but scope id does not exist",
			[]string{metadatatypes.ScopeMetadataAddress(notAUsedUUID).String(), s.sessionUUID.String()},
			"",
			[]string{"session:", "session: null"},
		},
		{
			"scope id and session uuid and scope id exists but session uuid does not",
			[]string{s.scopeID.String(), notAUsedUUID.String()},
			"",
			[]string{"session:", "session: null"},
		},
		{
			"session from scope id and record name as text",
			[]string{s.scopeID.String(), s.recordName, s.asText},
			"",
			[]string{indentedSessionText},
		},
		{
			"session from scope id and record name as json",
			[]string{s.scopeID.String(), s.recordName, s.asJson},
			"",
			[]string{s.sessionAsJson},
		},
		{
			"scope id and record name but scope id does not exist",
			[]string{metadatatypes.ScopeMetadataAddress(notAUsedUUID).String(), s.recordName},
			fmt.Sprintf("record %s does not exist: invalid request",
				metadatatypes.RecordMetadataAddress(notAUsedUUID, s.recordName)),
			[]string{},
		},
		{
			"scope id and record name and scope id exists but record does not",
			[]string{s.scopeID.String(), "not-a-record-name-that-exists"},
			fmt.Sprintf("record %s does not exist: invalid request",
				metadatatypes.RecordMetadataAddress(s.scopeUUID, "not-a-record-name-that-exists")),
			[]string{},
		},
		{
			"session from scope uuid and record name as text",
			[]string{s.scopeUUID.String(), s.recordName, s.asText},
			"",
			[]string{indentedSessionText},
		},
		{
			"session from scope uuid and record name as json",
			[]string{s.scopeUUID.String(), s.recordName, s.asJson},
			"",
			[]string{s.sessionAsJson},
		},
		{
			"scope uuid and record name but scope uuid does not exist",
			[]string{notAUsedUUID.String(), s.recordName},
			fmt.Sprintf("record %s does not exist: invalid request",
				metadatatypes.RecordMetadataAddress(notAUsedUUID, s.recordName)),
			[]string{},
		},
		{
			"scope uuid and record name and scope uuid exists but record does not",
			[]string{s.scopeUUID.String(), "not-a-record"},
			fmt.Sprintf("record %s does not exist: invalid request",
				metadatatypes.RecordMetadataAddress(s.scopeUUID, "not-a-record")),
			[]string{},
		},
		{
			"session from record id as text",
			[]string{s.recordID.String(), s.asText},
			"",
			[]string{indentedSessionText},
		},
		{
			"session from record id as json",
			[]string{s.recordID.String(), s.asJson},
			"",
			[]string{s.sessionAsJson},
		},
		{
			"record id but scope does not exist",
			[]string{metadatatypes.RecordMetadataAddress(notAUsedUUID, s.recordName).String()},
			fmt.Sprintf("record %s does not exist: invalid request",
				metadatatypes.RecordMetadataAddress(notAUsedUUID, s.recordName)),
			[]string{},
		},
		{
			"record id in existing scope but record does not exist",
			[]string{metadatatypes.RecordMetadataAddress(s.scopeUUID, "not-a-record-name").String()},
			fmt.Sprintf("record %s does not exist: invalid request",
				metadatatypes.RecordMetadataAddress(s.scopeUUID, "not-a-record-name")),
			[]string{},
		},
		{
			"sessions all as text",
			[]string{"all", s.asText},
			"",
			[]string{indentedSessionText},
		},
		{
			"sessions all as json",
			[]string{"all", s.asJson},
			"",
			[]string{s.sessionAsJson},
		},
		{
			"bad prefix",
			[]string{"scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"},
			"address [scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m] is not a scope address: invalid request",
			[]string{},
		},
		{
			"bad arg 1",
			[]string{"bad"},
			"could not parse [bad] into either a scope address (decoding bech32 failed: invalid bech32 string length 3) or uuid (invalid UUID length: 3): invalid request",
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
	cmd := func() *cobra.Command { return cli.GetMetadataRecordCmd() }

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
			"address [contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn] is not a scope address: invalid request",
			[]string{},
		},
		{
			"bad arg 1",
			[]string{"badbad"},
			"could not parse [badbad] into either a scope address (decoding bech32 failed: invalid bech32 string length 6) or uuid (invalid UUID length: 6): invalid request",
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
	cmd := func() *cobra.Command { return cli.GetMetadataScopeSpecCmd() }

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
			"address [scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel] is not a scope spec address: invalid request",
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
			"could not parse [reallybad] into either a scope spec address (decoding bech32 failed: invalid separator index -1) or uuid (invalid UUID length: 9): invalid request",
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
	cmd := func() *cobra.Command { return cli.GetMetadataContractSpecCmd() }

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
			"invalid specification id: address [record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3] is not a contract spec address: invalid request",
			[]string{},
		},
		{
			"bad arg",
			[]string{"badbadarg"},
			"invalid specification id: could not parse [badbadarg] into either a contract spec address (decoding bech32 failed: invalid separator index -1) or uuid (invalid UUID length: 9): invalid request",
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
	cmd := func() *cobra.Command { return cli.GetMetadataRecordSpecCmd() }

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
			"invalid specification id: address [session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr] is not a contract spec address: invalid request",
			[]string{},
		},
		{
			"bad arg 1",
			[]string{"not-gonna-parse"},
			"invalid specification id: could not parse [not-gonna-parse] into either a contract spec address (decoding bech32 failed: invalid separator index -1) or uuid (invalid UUID length: 15): invalid request",
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
	cmd := func() *cobra.Command { return cli.GetOwnershipCmd() }

	newUser := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()

	paginationText := `pagination:
  next_key: null
  total: "0"
`
	scopeUUIDsText := fmt.Sprintf(`scope_uuids:
- %s`,
		s.scopeUUID,
	)

	testCases := []queryCmdTestCase{
		{
			"scopes as json",
			[]string{s.user1AddrStr, s.asJson},
			"",
			[]string{
				fmt.Sprintf("\"scope_uuids\":[\"%s\"]", s.scopeUUID),
				"\"pagination\":{\"next_key\":null,\"total\":\"0\"}",
			},
		},
		{
			"scopes as text",
			[]string{s.user1AddrStr, s.asText},
			"",
			[]string{scopeUUIDsText, paginationText},
		},
		{
			"scope through value owner",
			[]string{s.user2AddrStr},
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
			[]string{s.user1AddrStr, s.user2AddrStr},
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
	cmd := func() *cobra.Command { return cli.GetValueOwnershipCmd() }

	paginationText := `pagination:
  next_key: null
  total: "0"
`
	scopeUUIDsText := fmt.Sprintf(`scope_uuids:
- %s`,
		s.scopeUUID,
	)

	testCases := []queryCmdTestCase{
		{
			"as json",
			[]string{s.user2AddrStr, s.asJson},
			"",
			[]string{
				fmt.Sprintf("\"scope_uuids\":[\"%s\"]", s.scopeUUID),
				"\"pagination\":{\"next_key\":null,\"total\":\"0\"}",
			},
		},
		{
			"as text",
			[]string{s.user2AddrStr, s.asText},
			"",
			[]string{scopeUUIDsText, paginationText},
		},
		{
			"no result",
			[]string{s.user1AddrStr},
			"",
			[]string{"scope_uuids: []", "total: \"0\""},
		},
		{
			"two args",
			[]string{s.user1AddrStr, s.user2AddrStr},
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
	cmd := func() *cobra.Command { return cli.GetOSLocatorCmd() }

	indentedLocator1Text := indent(s.objectLocator1AsText, 2)
	indentedLocator2Text := indent(s.objectLocator2AsText, 2)
	listEntryLocator1 := yamlListEntry(s.objectLocator1AsText)
	listEntryLocator2 := yamlListEntry(s.objectLocator2AsText)
	unknownUUID := uuid.New()

	testCases := []queryCmdTestCase{
		{
			"params as text",
			[]string{"params", s.asText},
			"",
			[]string{
				"params:",
				fmt.Sprintf("max_uri_length: %d", metadatatypes.DefaultMaxURILength),
			},
		},
		{
			"params as json",
			[]string{"params", s.asJson},
			"",
			[]string{
				"\"params\":{",
				fmt.Sprintf("\"max_uri_length\":%d", metadatatypes.DefaultMaxURILength),
			},
		},
		{
			"all as text",
			[]string{"all", s.asText},
			"",
			[]string{listEntryLocator1, listEntryLocator2},
		},
		{
			"all as json",
			[]string{"all", s.asJson},
			"",
			[]string{s.objectLocator1AsJson, s.objectLocator2AsJson},
		},
		{
			"by owner locator 1 as text",
			[]string{s.user1AddrStr, s.asText},
			"",
			[]string{indentedLocator1Text},
		},
		{
			"by owner locator 1 as json",
			[]string{s.user1AddrStr, s.asJson},
			"",
			[]string{s.objectLocator1AsJson},
		},
		{
			"by owner locator 2 as text",
			[]string{s.user2AddrStr, s.asText},
			"",
			[]string{indentedLocator2Text},
		},
		{
			"by owner locator 2 as json",
			[]string{s.user2AddrStr, s.asJson},
			"",
			[]string{s.objectLocator2AsJson},
		},
		{
			"by owner unknown owner",
			[]string{s.userOtherAddr.String()},
			"no locator bound to address: unknown request",
			[]string{""},
		},
		{
			"by scope id as text",
			[]string{s.scopeID.String(), s.asText},
			"",
			[]string{listEntryLocator1},
		},
		{
			"by scope id as json",
			[]string{s.scopeID.String(), s.asJson},
			"",
			[]string{s.objectLocator1AsJson},
		},
		{
			"by scope id unknown scope id",
			[]string{metadatatypes.ScopeMetadataAddress(unknownUUID).String()},
			fmt.Sprintf("scope [%s] not found: invalid request",
				metadatatypes.ScopeMetadataAddress(unknownUUID)),
			[]string{s.objectLocator1AsJson},
		},
		{
			"by scope uuid as text",
			[]string{s.scopeUUID.String(), s.asText},
			"",
			[]string{listEntryLocator1},
		},
		{
			"by scope uuid as json",
			[]string{s.scopeUUID.String(), s.asJson},
			"",
			[]string{s.objectLocator1AsJson},
		},
		{
			"by scope uuid unknown scope uuid",
			[]string{unknownUUID.String()},
			fmt.Sprintf("scope [%s] not found: invalid request",
				unknownUUID),
			[]string{s.objectLocator1AsJson},
		},
		{
			"by uri locator 1 as text",
			[]string{s.uri1, s.asText},
			"",
			[]string{listEntryLocator1},
		},
		{
			"by uri locator 1 as json",
			[]string{s.uri1, s.asJson},
			"",
			[]string{s.objectLocator1AsJson},
		},
		{
			"by uri locator 2 as text",
			[]string{s.uri2, s.asText},
			"",
			[]string{listEntryLocator2},
		},
		{
			"by uri locator 2 as json",
			[]string{s.uri2, s.asJson},
			"",
			[]string{s.objectLocator2AsJson},
		},
		{
			"by uri unknown uri",
			[]string{"http://not-an-entry.corn"},
			"No records found.: unknown request",
			[]string{},
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

// ---------- tx cmd tests ----------

type txCmdTestCase struct {
	name string
	cmd  *cobra.Command
	args []string
	// expectErr is only used if len(expectErrMsg) == 0.
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
			clientCtx := s.getClientCtx()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.cmd, tc.args)

			if len(tc.expectErrMsg) > 0 {
				require.EqualError(t, err, tc.expectErrMsg, "%s expected error message", cmdName)
			} else if tc.expectErr {
				require.Error(t, err, "%s expected error", cmdName)
			} else {
				require.NoError(t, err, "%s unexpected error", cmdName)

				umErr := clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType)
				require.NoError(t, umErr, "%s UnmarshalJSON error", cmdName)

				txResp := tc.respType.(*sdk.TxResponse)
				assert.Equal(t, int32(tc.expectedCode), int32(txResp.Code), "%s response code", cmdName)
				// Note: If the above is failing because a 0 is expected, it might mean that the keeper method is returning an error.

				if t.Failed() {
					t.Logf("tx:\n%v\n", txResp)
				}
			}
		})
	}
}

func (s *IntegrationCLITestSuite) TestScopeTxCommands() {

	scopeID := metadatatypes.ScopeMetadataAddress(uuid.New()).String()
	scopeSpecID := metadatatypes.ScopeSpecMetadataAddress(uuid.New()).String()
	testCases := []txCmdTestCase{
		{
			"should successfully add scope specification for test setup",
			cli.WriteScopeSpecificationCmd(),
			[]string{
				scopeSpecID,
				s.accountAddrStr,
				"owner",
				s.contractSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, "", &sdk.TxResponse{}, 0,
		},
		{
			"should successfully add metadata scope",
			cli.WriteScopeCmd(),
			[]string{
				scopeID,
				scopeSpecID,
				s.accountAddrStr,
				s.accountAddrStr,
				s.accountAddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
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
				scopeSpecID,
				s.user1AddrStr,
				s.user1AddrStr,
				s.user1AddrStr,
				fmt.Sprintf("--%s=%s", cli.FlagSigners, s.accountAddrStr),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, "", &sdk.TxResponse{}, 0,
		},
		{
			"should successfully add metadata scope with party rollup",
			cli.WriteScopeCmd(),
			[]string{
				scopeID,
				scopeSpecID,
				s.accountAddrStr,
				s.accountAddrStr,
				s.accountAddrStr,
				fmt.Sprintf("--%s", cli.FlagRequirePartyRollup),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
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
				s.user1AddrStr,
				s.user1AddrStr,
				s.user1AddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "invalid scope id: decoding bech32 failed: invalid separator index -1", &sdk.TxResponse{}, 0,
		},
		{
			"should fail to add metadata scope, incorrect scope spec id",
			cli.WriteScopeCmd(),
			[]string{
				metadatatypes.ScopeMetadataAddress(uuid.New()).String(),
				"not-a-uuid",
				s.user1AddrStr,
				s.user1AddrStr,
				s.user1AddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "invalid spec id: decoding bech32 failed: invalid separator index -1", &sdk.TxResponse{}, 0,
		},
		{
			"should fail to add metadata scope, validate basic will err on owner format",
			cli.WriteScopeCmd(),
			[]string{
				metadatatypes.ScopeMetadataAddress(uuid.New()).String(),
				metadatatypes.ScopeSpecMetadataAddress(uuid.New()).String(),
				"incorrect1,incorrect2",
				s.user1AddrStr,
				s.user1AddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, `invalid owners: invalid party "incorrect1,incorrect2": invalid address "incorrect1": decoding bech32 failed: invalid separator index 9`, &sdk.TxResponse{}, 0,
		},
		{
			"should fail to remove metadata scope, invalid scopeid",
			cli.RemoveScopeCmd(),
			[]string{
				"not-valid",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "decoding bech32 failed: invalid separator index -1", &sdk.TxResponse{}, 0,
		},
		{
			"should fail to add/remove metadata scope data access, invalid scopeid",
			cli.AddRemoveScopeDataAccessCmd(),
			[]string{
				"add",
				"not-valid",
				s.user2AddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "decoding bech32 failed: invalid separator index -1", &sdk.TxResponse{}, 0,
		},
		{
			"should fail to add/remove metadata scope data access, invalid command requires add or remove",
			cli.AddRemoveScopeDataAccessCmd(),
			[]string{
				"notaddorremove",
				scopeID,
				s.user2AddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
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
				scopeSpecID,
				s.user2AddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, fmt.Sprintf("meta address is not a scope: %s", scopeSpecID), &sdk.TxResponse{}, 0,
		},
		{
			"should fail to add/remove metadata scope data access, validatebasic fails",
			cli.AddRemoveScopeDataAccessCmd(),
			[]string{
				"add",
				scopeID,
				"notauser",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
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
				s.user1AddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
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
				s.user1AddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, "", &sdk.TxResponse{}, 0,
		},

		{
			"should fail to add/remove metadata scope owners, invalid scopeid",
			cli.AddRemoveScopeOwnersCmd(),
			[]string{
				"add",
				"not-valid",
				s.user2AddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "decoding bech32 failed: invalid separator index -1", &sdk.TxResponse{}, 0,
		},
		{
			"should fail to add/remove metadata scope owner, invalid command requires add or remove",
			cli.AddRemoveScopeOwnersCmd(),
			[]string{
				"notaddorremove",
				scopeID,
				s.user2AddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "incorrect command notaddorremove : required remove or update", &sdk.TxResponse{}, 0,
		},
		{
			"should fail to add/remove metadata scope owner, not a scope id",
			cli.AddRemoveScopeOwnersCmd(),
			[]string{
				"add",
				scopeSpecID,
				s.user2AddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, fmt.Sprintf("meta address is not a scope: %s", scopeSpecID), &sdk.TxResponse{}, 0,
		},
		{
			"should fail to add/remove metadata scope owner, validatebasic fails",
			cli.AddRemoveScopeOwnersCmd(),
			[]string{
				"add",
				scopeID,
				"notauser",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "invalid owners: invalid party address [notauser]: decoding bech32 failed: invalid separator index -1", &sdk.TxResponse{}, 0,
		},
		{
			"should successfully remove metadata scope",
			cli.RemoveScopeCmd(),
			[]string{
				scopeID,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, "", &sdk.TxResponse{}, 0,
		},
		{
			"should fail to delete metadata scope that no longer exists",
			cli.RemoveScopeCmd(),
			[]string{
				scopeID,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, "", &sdk.TxResponse{}, 18,
		},
		{
			name: "should fail to write scope with optional party but without rollup",
			cmd:  cli.WriteScopeCmd(),
			args: []string{
				metadatatypes.ScopeMetadataAddress(uuid.New()).String(),
				scopeSpecID,
				fmt.Sprintf("%s,servicer,opt", s.accountAddrStr),
				s.accountAddrStr,
				s.accountAddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErrMsg: "parties can only be optional when require_party_rollup = true",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should successfully write scope with optional party and rollup",
			cmd:  cli.WriteScopeCmd(),
			args: []string{
				metadatatypes.ScopeMetadataAddress(uuid.New()).String(),
				scopeSpecID,
				fmt.Sprintf("%s,servicer,opt;%s,owner", s.accountAddrStr, s.accountAddrStr),
				s.accountAddrStr,
				s.accountAddrStr,
				fmt.Sprintf("--%s", cli.FlagRequirePartyRollup),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
	}

	runTxCmdTestCases(s, testCases)
}

func (s *IntegrationCLITestSuite) TestUpdateMigrateValueOwnersCmds() {
	scopeSpecID := metadatatypes.ScopeSpecMetadataAddress(uuid.New()).String()
	scopeID1 := metadatatypes.ScopeMetadataAddress(uuid.New()).String()
	scopeID2 := metadatatypes.ScopeMetadataAddress(uuid.New()).String()
	scopeID3 := metadatatypes.ScopeMetadataAddress(uuid.New()).String()

	feeFlag := func(amt int64) string {
		return fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, amt)).String())
	}
	fromFlag := func(addr string) string {
		return fmt.Sprintf("--%s=%s", flags.FlagFrom, addr)
	}
	skipConfFlag := fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation)
	broadcastBlockFlag := fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock)

	queryCmd := func() *cobra.Command {
		return cli.GetMetadataScopeCmd()
	}
	queryTests := func(scope1ValueOwner, scope2ValueOwner, scope3ValueOwner string) []queryCmdTestCase {
		return []queryCmdTestCase{
			{
				name: "scope 1 value owner", args: []string{scopeID1},
				expectedInOutput: []string{"value_owner_address: " + scope1ValueOwner},
			},
			{
				name: "scope 2 value owner", args: []string{scopeID2},
				expectedInOutput: []string{"value_owner_address: " + scope2ValueOwner},
			},
			{
				name: "scope 3 value owner", args: []string{scopeID3},
				expectedInOutput: []string{"value_owner_address: " + scope3ValueOwner},
			},
		}
	}

	tests := []struct {
		txs     []txCmdTestCase
		queries []queryCmdTestCase
	}{
		{
			// Some failures that come from cmd.RunE.
			txs: []txCmdTestCase{
				{
					name: "update: only 1 arg",
					cmd:  cli.UpdateValueOwnersCmd(),
					args: []string{
						s.user2AddrStr,
						fromFlag(s.user1AddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
					},
					expectErrMsg: "requires at least 2 arg(s), only received 1",
				},
				{
					name: "update: invalid value owner",
					cmd:  cli.UpdateValueOwnersCmd(),
					// <new value owner> <scope id> [<scope id 2> ...]
					args: []string{
						"notabech32", scopeID1, scopeID2,
						fromFlag(s.user1AddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
					},
					expectErrMsg: "invalid new value owner \"notabech32\": decoding bech32 failed: invalid separator index -1",
				},
				{
					name: "update: invalid scope id",
					cmd:  cli.UpdateValueOwnersCmd(),
					// <new value owner> <scope id> [<scope id 2> ...]
					args: []string{
						s.user1AddrStr, scopeID1, scopeSpecID,
					},
					expectErrMsg: fmt.Sprintf("invalid scope id %d %q: %s", 2, scopeSpecID, "not a scope identifier"),
				},
				{
					name: "update: invalid signers",
					cmd:  cli.UpdateValueOwnersCmd(),
					// <new value owner> <scope id> [<scope id 2> ...]
					args: []string{
						s.user1AddrStr, scopeID1, scopeID2,
						"--" + cli.FlagSigners, "notabech32",
						fromFlag(s.user1AddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
					},
					expectErrMsg: "decoding bech32 failed: invalid separator index -1",
				},
				{
					name: "migrate: only 1 arg",
					cmd:  cli.MigrateValueOwnerCmd(),
					args: []string{
						s.user2AddrStr,
						fromFlag(s.user1AddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
					},
					expectErrMsg: "accepts 2 arg(s), received 1",
				},
				{
					name: "migrate: 3 args",
					cmd:  cli.MigrateValueOwnerCmd(),
					args: []string{
						s.user1AddrStr, s.user2AddrStr, s.user3AddrStr,
						fromFlag(s.user1AddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
					},
					expectErrMsg: "accepts 2 arg(s), received 3",
				},
				{
					name: "migrate: invalid existing value owner",
					cmd:  cli.MigrateValueOwnerCmd(),
					// <new value owner> <scope id> [<scope id 2> ...]
					args: []string{
						"notabech32", s.user2AddrStr,
						fromFlag(s.user1AddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
					},
					expectErrMsg: "invalid existing value owner \"notabech32\": decoding bech32 failed: invalid separator index -1",
				},
				{
					name: "migrate: invalid proposed value owner",
					cmd:  cli.MigrateValueOwnerCmd(),
					// <new value owner> <scope id> [<scope id 2> ...]
					args: []string{
						s.user2AddrStr, "notabech32",
						fromFlag(s.user1AddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
					},
					expectErrMsg: "invalid proposed value owner \"notabech32\": decoding bech32 failed: invalid separator index -1",
				},
				{
					name: "migrate: invalid signers",
					cmd:  cli.MigrateValueOwnerCmd(),
					// <new value owner> <scope id> [<scope id 2> ...]
					args: []string{
						s.user1AddrStr, s.user2AddrStr,
						"--" + cli.FlagSigners, "notabech32",
						fromFlag(s.user1AddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
					},
					expectErrMsg: "decoding bech32 failed: invalid separator index -1",
				},
			},
		},
		{
			// Setup 3 scopes. 1 and 2 have user 1 for value owner, scope 3 has value owner 2.
			txs: []txCmdTestCase{
				{
					name: "setup: write scope spec",
					cmd:  cli.WriteScopeSpecificationCmd(),
					// [specification-id] [owner-addresses] [responsible-parties] [contract-specification-ids] [description-name, optional] [description, optional] [website-url, optional] [icon-url, optional]
					args: []string{
						scopeSpecID, s.accountAddrStr, "owner", s.contractSpecID.String(),
						fromFlag(s.accountAddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
					},
					respType: &sdk.TxResponse{},
				},
				{
					name: "setup: write scope 1",
					cmd:  cli.WriteScopeCmd(),
					// [scope-id] [spec-id] [owners] [data-access] [value-owner-address]
					args: []string{
						scopeID1, scopeSpecID,
						s.accountAddrStr, s.accountAddrStr, s.user1AddrStr,
						fromFlag(s.accountAddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
					},
					respType: &sdk.TxResponse{},
				},
				{
					name: "setup: write scope 2",
					cmd:  cli.WriteScopeCmd(),
					// [scope-id] [spec-id] [owners] [data-access] [value-owner-address]
					args: []string{
						scopeID2, scopeSpecID,
						s.accountAddrStr, s.accountAddrStr, s.user1AddrStr,
						fromFlag(s.accountAddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
					},
					respType: &sdk.TxResponse{},
				},
				{
					name: "setup: write scope 3",
					cmd:  cli.WriteScopeCmd(),
					// [scope-id] [spec-id] [owners] [data-access] [value-owner-address]
					args: []string{
						scopeID3, scopeSpecID,
						s.accountAddrStr, s.accountAddrStr, s.user2AddrStr,
						fromFlag(s.accountAddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
					},
					respType: &sdk.TxResponse{},
				},
			},
			queries: queryTests(s.user1AddrStr, s.user1AddrStr, s.user2AddrStr),
		},
		{
			// Some failures from the keeper.
			txs: []txCmdTestCase{
				{
					name: "update: incorrect signer",
					cmd:  cli.UpdateValueOwnersCmd(),
					// <new value owner> <scope id> [<scope id 2> ...]
					args: []string{
						s.accountAddrStr, scopeID1, scopeID2,
						fromFlag(s.accountAddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
					},
					respType:     &sdk.TxResponse{},
					expectedCode: 18,
				},
				{
					name: "update: missing signature",
					cmd:  cli.UpdateValueOwnersCmd(),
					// <new value owner> <scope id> [<scope id 2> ...]
					args: []string{
						s.user2AddrStr, scopeID1, scopeID2, scopeID3,
						fromFlag(s.user1AddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
					},
					respType:     &sdk.TxResponse{},
					expectedCode: 18,
				},
				{
					name: "migrate: incorrect signer",
					cmd:  cli.MigrateValueOwnerCmd(),
					// <new value owner> <scope id> [<scope id 2> ...]
					args: []string{
						s.user1AddrStr, s.user2AddrStr,
						fromFlag(s.accountAddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
					},
					respType:     &sdk.TxResponse{},
					expectedCode: 18,
				},
			},
		},
		{
			// A single update of two scopes.
			txs: []txCmdTestCase{{
				name: "update: scopes 1 and 2 to user 2",
				cmd:  cli.UpdateValueOwnersCmd(),
				// <new value owner> <scope id> [<scope id 2> ...]
				args: []string{
					s.user2AddrStr, scopeID1, scopeID2,
					fromFlag(s.user1AddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
				},
				respType: &sdk.TxResponse{},
			}},
			queries: queryTests(s.user2AddrStr, s.user2AddrStr, s.user2AddrStr),
		},
		{
			// A single update of 3 scopes.
			txs: []txCmdTestCase{{
				name: "update: scopes 1 2 and 3 to user 3",
				cmd:  cli.UpdateValueOwnersCmd(),
				// <new value owner> <scope id> [<scope id 2> ...]
				args: []string{
					s.user3AddrStr, scopeID1, scopeID2, scopeID3,
					fromFlag(s.user2AddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
				},
				respType: &sdk.TxResponse{},
			}},
			queries: queryTests(s.user3AddrStr, s.user3AddrStr, s.user3AddrStr),
		},
		{
			// Two updates of two different scopes.
			txs: []txCmdTestCase{
				{
					name: "update: scope 1 to user 1",
					cmd:  cli.UpdateValueOwnersCmd(),
					// <new value owner> <scope id> [<scope id 2> ...]
					args: []string{
						s.user1AddrStr, scopeID1,
						fromFlag(s.user3AddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
					},
					respType: &sdk.TxResponse{},
				},
				{
					name: "update: scope 2 to user 2",
					cmd:  cli.UpdateValueOwnersCmd(),
					// <new value owner> <scope id> [<scope id 2> ...]
					args: []string{
						s.user2AddrStr, scopeID2,
						fromFlag(s.user3AddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
					},
					respType: &sdk.TxResponse{},
				},
			},
			queries: queryTests(s.user1AddrStr, s.user2AddrStr, s.user3AddrStr),
		},
		{
			// A single migrate of 1 scope.
			txs: []txCmdTestCase{{
				name: "migrate: user 1 scope to user 2",
				cmd:  cli.MigrateValueOwnerCmd(),
				// <new value owner> <scope id> [<scope id 2> ...]
				args: []string{
					s.user1AddrStr, s.user2AddrStr,
					fromFlag(s.user1AddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
				},
				respType: &sdk.TxResponse{},
			}},
			queries: queryTests(s.user2AddrStr, s.user2AddrStr, s.user3AddrStr),
		},
		{
			// A single migrate of 2 scopes.
			txs: []txCmdTestCase{{
				name: "migrate: user 2 scopes to user 3",
				cmd:  cli.MigrateValueOwnerCmd(),
				// <new value owner> <scope id> [<scope id 2> ...]
				args: []string{
					s.user2AddrStr, s.user3AddrStr,
					fromFlag(s.user2AddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
				},
				respType: &sdk.TxResponse{},
			}},
			queries: queryTests(s.user3AddrStr, s.user3AddrStr, s.user3AddrStr),
		},
	}

	for i, tc := range tests {
		lead := fmt.Sprintf("%d: ", i)
		for c := range tc.txs {
			tc.txs[c].name = lead + tc.txs[c].name
		}
		for c := range tc.queries {
			tc.queries[c].name = lead + tc.queries[c].name
		}

		runTxCmdTestCases(s, tc.txs)
		if len(tc.queries) > 0 {
			runQueryCmdTestCases(s, queryCmd, tc.queries)
		}
	}

	// If there was a failure, output the user bech32s so it's easier to figure out what went wrong.
	if s.T().Failed() {
		s.T().Logf("accountAddrStr: %s", s.accountAddrStr)
		s.T().Logf("user1AddrStr: %s", s.user1AddrStr)
		s.T().Logf("user2AddrStr: %s", s.user2AddrStr)
		s.T().Logf("user3AddrStr: %s", s.user3AddrStr)
	}
}

func (s *IntegrationCLITestSuite) TestScopeSpecificationTxCommands() {
	addCommand := cli.WriteScopeSpecificationCmd()
	removeCommand := cli.RemoveScopeSpecificationCmd()
	specID := metadatatypes.ScopeSpecMetadataAddress(uuid.New())
	testCases := []txCmdTestCase{
		{
			name: "should successfully add scope specification",
			cmd:  addCommand,
			args: []string{
				specID.String(),
				s.accountAddrStr,
				"owner",
				s.contractSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should successfully update scope specification with descriptions",
			cmd:  addCommand,
			args: []string{
				specID.String(),
				s.accountAddrStr,
				"owner",
				s.contractSpecID.String(),
				"description-name",
				"description",
				"http://www.blockchain.com/",
				"http://www.blockchain.com/icon.png",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should fail to add scope specification, invalid spec id format",
			cmd:  addCommand,
			args: []string{
				"invalid",
				s.accountAddrStr,
				"owner",
				s.contractSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			expectErrMsg: "decoding bech32 failed: invalid bech32 string length 7",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should fail to add scope specification validate basic error",
			cmd:  addCommand,
			args: []string{
				specID.String(),
				s.accountAddrStr,
				"owner",
				specID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			expectErrMsg: "invalid contract specification id prefix at index 0 (expected: contractspec, got scopespec)",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should fail to add scope specification unknown party type",
			cmd:  addCommand,
			args: []string{
				metadatatypes.ScopeMetadataAddress(uuid.New()).String(),
				s.accountAddrStr,
				"badpartytype",
				s.contractSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			expectErrMsg: `unknown party type: "badpartytype"`,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should fail to remove scope specification invalid id",
			cmd:  removeCommand,
			args: []string{
				"notvalid",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			expectErrMsg: "decoding bech32 failed: invalid separator index -1",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should successfully remove scope specification",
			cmd:  removeCommand,
			args: []string{
				specID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should fail to remove scope specification that has already been removed",
			cmd:  removeCommand,
			args: []string{
				specID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 38,
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
				s.accountAddrStr,
				userURI,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
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
				s.accountAddrStr,
				userURIMod,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
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
				s.accountAddrStr,
				userURIMod,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
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
			name: "should successfully add contract specification with resource hash",
			cmd:  addCommand,
			args: []string{
				specificationID.String(),
				s.accountAddrStr,
				"owner",
				"hashvalue",
				"`myclassname`",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should successfully update contract specification with resource hash using signer flag",
			cmd:  addCommand,
			args: []string{
				specificationID.String(),
				s.accountAddrStr,
				"owner",
				"hashvalue",
				"`myclassname`",
				fmt.Sprintf("--%s=%s", cli.FlagSigners, s.accountAddrStr),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should successfully update contract specification with resource id",
			cmd:  addCommand,
			args: []string{
				specificationID.String(),
				s.accountAddrStr,
				"owner",
				specificationID.String(),
				"myclassname",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should successfully update contract specification with description",
			cmd:  addCommand,
			args: []string{
				specificationID.String(),
				s.accountAddrStr,
				"owner",
				"hashvalue",
				"myclassname",
				"description-name",
				"description",
				"http://www.blockchain.com/",
				"http://www.blockchain.com/icon.png",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should successfully remove contract specification",
			cmd:  removeCommand,
			args: []string{
				specificationID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should fail to add contract specification on validate basic error",
			cmd:  addCommand,
			args: []string{
				"invalid-spec-id",
				s.accountAddrStr,
				"owner",
				"hashvalue",
				"myclassname",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			expectErrMsg: "decoding bech32 failed: invalid separator index -1",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should fail to add contract specification bad party type",
			cmd:  addCommand,
			args: []string{
				metadatatypes.ContractSpecMetadataAddress(uuid.New()).String(),
				s.accountAddrStr,
				"badpartytype",
				"hashvalue",
				"`myclassname`",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			expectErrMsg: `unknown party type: "badpartytype"`,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should fail to remove contract specification invalid address",
			cmd:  removeCommand,
			args: []string{
				"not-a-id",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			expectErrMsg: "decoding bech32 failed: invalid separator index -1",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should fail to remove contract that no longer exists",
			cmd:  removeCommand,
			args: []string{
				specificationID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 38,
		},
	}

	runTxCmdTestCases(s, testCases)
}

func (s *IntegrationCLITestSuite) TestContractSpecificationScopeSpecAddRemoveTxCommands() {
	addCommand := cli.AddContractSpecToScopeSpecCmd()
	removeCommand := cli.RemoveContractSpecFromScopeSpecCmd()
	contractSpecUUID := uuid.New()
	specificationID := metadatatypes.ContractSpecMetadataAddress(contractSpecUUID)
	scopeSpecID := metadatatypes.ScopeSpecMetadataAddress(uuid.New())

	testCases := []txCmdTestCase{
		{
			"should successfully add contract specification for test initialization",
			cli.WriteContractSpecificationCmd(),
			[]string{
				specificationID.String(),
				s.accountAddrStr,
				"owner",
				"hashvalue",
				"`myclassname`",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
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
			"should successfully add scope specification for test setup",
			cli.WriteScopeSpecificationCmd(),
			[]string{
				scopeSpecID.String(),
				s.accountAddrStr,
				"owner",
				s.contractSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, "", &sdk.TxResponse{}, 0,
		},
		{
			"should fail to add contract spec to scope spec, invalid contract spec id",
			addCommand,
			[]string{
				"invalid-contract-specid",
				scopeSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true,
			"invalid contract specification id : invalid-contract-specid",
			&sdk.TxResponse{},
			0,
		},
		{
			"should fail to add contract spec to scope spec, not a contract spec id",
			addCommand,
			[]string{
				scopeSpecID.String(),
				scopeSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true,
			fmt.Sprintf("invalid contract specification id : %s", scopeSpecID.String()),
			&sdk.TxResponse{},
			0,
		},
		{
			"should fail to add contract spec to scope spec, invalid scope spec id",
			addCommand,
			[]string{
				specificationID.String(),
				"invalid-scope-spec-id",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true,
			"invalid scope specification id : invalid-scope-spec-id",
			&sdk.TxResponse{},
			0,
		},
		{
			"should fail to add contract spec to scope spec, not a scope spec",
			addCommand,
			[]string{
				specificationID.String(),
				specificationID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true,
			fmt.Sprintf("invalid scope specification id : %s", specificationID.String()),
			&sdk.TxResponse{},
			0,
		},
		{
			"should successfully add contract spec to scope spec",
			addCommand,
			[]string{
				specificationID.String(),
				scopeSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
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
			"should fail to remove contract spec to scope spec, invalid contract spec id",
			removeCommand,
			[]string{
				"invalid-contract-specid",
				scopeSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true,
			"invalid contract specification id : invalid-contract-specid",
			&sdk.TxResponse{},
			0,
		},
		{
			"should fail to remove contract spec to scope spec, not a contract spec id",
			removeCommand,
			[]string{
				scopeSpecID.String(),
				scopeSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true,
			fmt.Sprintf("invalid contract specification id : %s", scopeSpecID.String()),
			&sdk.TxResponse{},
			0,
		},
		{
			"should fail to remove contract spec to scope spec, invalid scope spec id",
			removeCommand,
			[]string{
				specificationID.String(),
				"invalid-scope-spec-id",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true,
			"invalid scope specification id : invalid-scope-spec-id",
			&sdk.TxResponse{},
			0,
		},
		{
			"should fail to remove contract spec to scope spec, not a scope spec id",
			removeCommand,
			[]string{
				specificationID.String(),
				specificationID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true,
			fmt.Sprintf("invalid scope specification id : %s", specificationID.String()),
			&sdk.TxResponse{},
			0,
		},
		{
			"should successfully remove contract spec to scope spec",
			removeCommand,
			[]string{
				specificationID.String(),
				scopeSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
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
			name: "setup test with a record specification owned by signer",
			cmd:  addConractSpecCmd,
			args: []string{
				contractSpecID.String(),
				s.accountAddrStr,
				"owner",
				"hashvalue",
				"`myclassname`",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should successfully add record specification",
			cmd:  cmd,
			args: []string{
				specificationID.String(),
				recordName,
				"record1,typename1,hashvalue",
				"typename",
				"record",
				"validator",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should successfully add record specification",
			cmd:  cmd,
			args: []string{
				specificationID.String(),
				recordName,
				"record1,typename1,hashvalue",
				"typename",
				"record_list",
				"investor",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should fail to add record specification, bad party type",
			cmd:  cmd,
			args: []string{
				metadatatypes.RecordSpecMetadataAddress(uuid.New(), recordName).String(),
				recordName,
				"record1,typename1,hashvalue",
				"typename",
				"record_list",
				"badpartytype",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			expectErrMsg: `unknown party type: "badpartytype"`,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should fail to add record specification, validate basic fail",
			cmd:  cmd,
			args: []string{
				specificationID.String(),
				"",
				"record1,typename1,hashvalue;record2,typename2,recspec1q5p7xh9vtktyc9ynp25ydq4cycqp3tp7wdplq95fp3gsaylex5npzlhnhp6",
				"typename",
				"record",
				"custodian",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			expectErrMsg: "record specification name cannot be empty",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should fail to add record specification, fail parsing inputs too few values",
			cmd:  cmd,
			args: []string{
				specificationID.String(),
				recordName,
				"record1,typename1;record2,typename2,hashvalue",
				"typename",
				"record",
				"originator",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			expectErrMsg: `invalid input specification "record1,typename1": expected 3 parts, have 2`,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should fail to add record specification, incorrect result type",
			cmd:  cmd,
			args: []string{
				specificationID.String(),
				recordName,
				"record1,typename1,hashvalue",
				"typename",
				"incorrect",
				"servicer,affiliate",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			expectErrMsg: "record specification result type cannot be unspecified",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should fail to add record specification, incorrect signer format",
			cmd:  cmd,
			args: []string{
				specificationID.String(),
				recordName,
				"record1,typename1,hashvalue",
				"typename",
				"record",
				"provenance",
				fmt.Sprintf("--%s=%s", cli.FlagSigners, "incorrect-signer-format"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			expectErrMsg: "decoding bech32 failed: invalid separator index -1",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should fail to delete record specification, incorrect id",
			cmd:  deleteRecordSpecCmd,
			args: []string{
				"incorrect-id",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			expectErrMsg: "decoding bech32 failed: invalid separator index -1",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should fail to delete record specification, not a record specification",
			cmd:  deleteRecordSpecCmd,
			args: []string{
				contractSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			expectErrMsg: fmt.Sprintf("invalid contract specification id: %v", contractSpecID.String()),
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should successfully delete record specification",
			cmd:  deleteRecordSpecCmd,
			args: []string{
				specificationID.String(),
				fmt.Sprintf("--%s=%s", cli.FlagSigners, s.accountAddrStr),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should fail to delete record specification that does not exist",
			cmd:  deleteRecordSpecCmd,
			args: []string{
				specificationID.String(),
				fmt.Sprintf("--%s=%s", cli.FlagSigners, s.accountAddrStr),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 38,
		},
	}

	runTxCmdTestCases(s, testCases)
}

func (s *IntegrationCLITestSuite) TestRecordTxCommands() {
	userAddress := s.accountAddrStr
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
			"should successfully add contract specification with resource hash for test setup",
			cli.WriteContractSpecificationCmd(),
			[]string{
				contractSpecID.String(),
				userAddress,
				"owner",
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
			"should successfully add scope specification for test setup",
			cli.WriteScopeSpecificationCmd(),
			[]string{
				scopeSpecID.String(),
				userAddress,
				"owner",
				s.contractSpecID.String() + "," + contractSpecID.String(),
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
			"should successfully add record specification for test setup",
			cli.WriteRecordSpecificationCmd(),
			[]string{
				recSpecID.String(),
				recordName,
				"input1name,typename1,hashvalue",
				"typename",
				"record",
				"owner",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
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
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
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
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "decoding bech32 failed: invalid separator index -1",
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
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, "decoding bech32 failed: invalid separator index -1",
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
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, `invalid process "hashvalue,methodname": expected 3 parts, have: 2`,
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
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, `invalid record input "input1name,typename1,proposed": expected 4 parts, have 3`,
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
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, `invalid record output "outputhashvalue": expected 2 parts, have 1`,
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
				fmt.Sprintf("%s,%s", userAddress, userAddress),
				contractSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, fmt.Sprintf(`invalid party "%s,%s": unknown party type: "%s"`, userAddress, userAddress, userAddress),
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
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
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
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
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

	owner := s.accountAddrStr
	sender := s.accountAddrStr
	scopeUUID := uuid.New()
	scopeID := metadatatypes.ScopeMetadataAddress(scopeUUID)

	writeScopeCmd := cli.WriteScopeCmd()
	ctx := s.getClientCtx()
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
	umErr := ctx.Codec.UnmarshalJSON(out.Bytes(), &scopeResp)
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
		{
			name: "invalid party type",
			cmd:  cmd,
			args: []string{
				metadatatypes.SessionMetadataAddress(scopeUUID, uuid.New()).String(),
				s.contractSpecID.String(),
				fmt.Sprintf("%s,badpartytype", owner),
				"somename",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, sender),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			expectErrMsg: `invalid party "` + owner + `,badpartytype": unknown party type: "badpartytype"`,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
	}

	runTxCmdTestCases(s, testCases)
}

// ---------- tx cmd CountAuthorization tests ----------

func (s *IntegrationCLITestSuite) TestCountAuthorizationIntactTxCommands() {
	// The scenario being tested is as follows:
	// There are two owners (1 & 2) and two signers (3 & 4).
	// 1 & 2 are required signers and 3 & 4 are the actual signers.
	// It should find a grant (granter: 1 -> grantee: 3) and complain
	// about 4 not being one of the required signers because grant (granter: 2 -> grantee: 4) does not exist.
	//
	// NOTE: Signing in DIRECT mode is only supported for transactions with one signer only.
	//       So we'll test two owners with one signer the following way:
	//			1. create scope spec
	//			2. create scope with two owners (1 & 2) and value owner (1)
	//			3. add CountAuthorization: grant (granter: 1 -> grantee: 3, allowedAuthorizations: 1).
	//			4. validate that it "fails" to delete scope due to missing grant from 2 -> 3
	//			5. add CountAuthorization: grant (granter: 2 -> grantee: 3, allowedAuthorizations: 2 (helps distinguish grants in debugger)).
	//			6. validate that it "removes" scope because signer 3 has now been assigned two grants (1 -> 3 and 2 -> 3).

	scopeID := metadatatypes.ScopeMetadataAddress(uuid.New()).String()
	scopeSpecID := metadatatypes.ScopeSpecMetadataAddress(uuid.New()).String()
	testCases := []txCmdTestCase{
		{
			name: "should successfully add scope specification for test setup",
			cmd:  cli.WriteScopeSpecificationCmd(),
			args: []string{
				scopeSpecID,
				s.accountAddrStr,
				"owner",
				s.contractSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should successfully add metadata scope with two owners - owner 1 as value owner",
			cmd:  cli.WriteScopeCmd(),
			args: []string{
				scopeID,
				scopeSpecID,
				fmt.Sprintf("%s;%s", s.user1AddrStr, s.user2AddrStr),
				s.user1AddrStr,
				s.user1AddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should successfully add count authorization from owner 1 to signer 3",
			cmd:  authzcli.NewCmdGrantAuthorization(),
			args: []string{
				s.user3AddrStr,
				"count",
				fmt.Sprintf("--%s=%d", authzcli.FlagAllowedAuthorizations, 1),
				fmt.Sprintf("--%s=%s", authzcli.FlagMsgType, metadatatypes.TypeURLMsgDeleteScopeRequest),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.user1AddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should fail to remove metadata scope with signer 3 due to missing authz grant from owner 2",
			cmd:  cli.RemoveScopeCmd(),
			args: []string{
				scopeID,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.user3AddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 18,
		},
		{
			name: "should successfully add count authorization from owner 2 to signer 3",
			cmd:  authzcli.NewCmdGrantAuthorization(),
			args: []string{
				s.user3AddrStr,
				"count",
				fmt.Sprintf("--%s=%d", authzcli.FlagAllowedAuthorizations, 2),
				fmt.Sprintf("--%s=%s", authzcli.FlagMsgType, metadatatypes.TypeURLMsgDeleteScopeRequest),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.user2AddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should successfully remove metadata scope with signer 3, found grants for owner 1 & 2",
			cmd:  cli.RemoveScopeCmd(),
			args: []string{
				scopeID,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.user3AddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
	}

	runTxCmdTestCases(s, testCases)
}
