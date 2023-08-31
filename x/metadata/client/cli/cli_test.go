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
	attrcli "github.com/provenance-io/provenance/x/attribute/client/cli"
	attrtypes "github.com/provenance-io/provenance/x/attribute/types"
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

	scopeIDWithData metadatatypes.MetadataAddress

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

	// Set some account data on a scope. It should be fine even though the scope doesn't actually exist.
	s.scopeIDWithData = metadatatypes.ScopeMetadataAddress(uuid.MustParse("A11E57A6-7D51-4C43-91F9-AD1F4D16FA35"))
	attrData := attrtypes.DefaultGenesisState()
	attrData.Attributes = append(attrData.Attributes,
		attrtypes.Attribute{
			Name:          attrtypes.AccountDataName,
			Value:         []byte("This is some scope account data."),
			AttributeType: attrtypes.AttributeType_String,
			Address:       s.scopeIDWithData.String(),
		},
	)
	attrDataBz, err := cfg.Codec.MarshalJSON(attrData)
	s.Require().NoError(err, "MarshalJSON(attrData)")
	genesisState[attrtypes.ModuleName] = attrDataBz

	cfg.GenesisState = genesisState

	cfg.ChainID = antewrapper.SimAppChainID
	cfg.TimeoutCommit = 500 * time.Millisecond
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

// indent breaks the str into lines and adds the desired number of spaces at the start of each line.
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

// entryIndent is similar to indent except the very first line will include the - that a yaml entry would have.
// Panics if spaces < 2.
func entryIndent(str string, spaces int) string {
	var sb strings.Builder
	lines := strings.Split(str, "\n")
	maxI := len(lines) - 1
	s := strings.Repeat(" ", spaces)
	for i, l := range lines {
		if i == 0 {
			sb.WriteString(strings.Repeat(" ", spaces-2) + "- ")
		} else {
			sb.WriteString(s)
		}
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
	name   string
	args   []string
	expErr string
	expOut []string
}

func runQueryCmdTestCases(s *IntegrationCLITestSuite, cmdGen func() *cobra.Command, testCases []queryCmdTestCase) {
	// Providing the command using a generator (cmdGen), we get a new instance of the cmd each time, and the flags won't
	// carry over between tests on the same command.
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			cmd := cmdGen()
			cmdName := cmd.Name()
			var outStr string
			defer func() {
				if t.Failed() {
					t.Logf("Command: %s\nArgs: %q\nOutput:\n%s", cmdName, tc.args, outStr)
				}
			}()

			clientCtx := s.getClientCtxWithoutKeyring()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			outStr = out.String()

			if len(tc.expErr) > 0 {
				require.ErrorContains(t, err, tc.expErr, "%s error", cmdName)
				// Something deep down is double wrapping the errors.
				// E.g. "foo: invalid request" has become "foo: invalid request: invalid request"
				// So we changed from the "Equal" test below to the "Contains" test above.
				// If you're bored, maybe try swapping back to see if things have been fixed.
				//require.EqualError(t, err, tc.expErr, "%s error", cmdName)
			} else {
				require.NoErrorf(t, err, "%s error", cmdName)
			}
			for _, exp := range tc.expOut {
				if !assert.Contains(t, outStr, exp, "%s command output", cmdName) {
					// The expected entry is easily lost in the failure message, so log it now too.
					// Logging it instead of putting it in the assertion message so it lines up with the deferrable.
					t.Logf("Not Found:\n%s", exp)
				}
			}
		})
	}
}

func (s *IntegrationCLITestSuite) TestGetMetadataParamsCmd() {
	cmd := func() *cobra.Command { return cli.GetMetadataParamsCmd() }

	testCases := []queryCmdTestCase{
		{
			name:   "get params as json output",
			args:   []string{s.asJson},
			expOut: []string{"\"params\":{}"},
		},
		{
			name:   "get params as text output",
			args:   []string{s.asText},
			expOut: []string{"params: {}"},
		},
		{
			name:   "get params - invalid args",
			args:   []string{"bad-arg"},
			expErr: "unknown argument: bad-arg",
		},
		{
			name:   "get params as json output including request",
			args:   []string{s.asJson, s.includeRequest},
			expOut: []string{"\"params\":{}", "\"request\":{\"include_request\":true}"},
		},
		{
			name:   "get locator params as json",
			args:   []string{"locator", s.asJson},
			expOut: []string{"\"params\":{", "\"max_uri_length\":2048"},
		},
		{
			name:   "get locator params as text including request",
			args:   []string{"locator", s.asText, s.includeRequest},
			expOut: []string{"params:", "max_uri_length: 2048", "request:", "include_request: true"},
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetMetadataByIDCmd() {
	cmd := func() *cobra.Command { return cli.GetMetadataByIDCmd() }

	testCases := []queryCmdTestCase{
		{
			name:   "get metadata by id - scope id as json",
			args:   []string{s.scopeID.String(), s.asJson},
			expOut: []string{s.scopeAsJson},
		},
		{
			name:   "get metadata by id - scope id as text",
			args:   []string{s.scopeID.String(), s.asText},
			expOut: []string{entryIndent(s.scopeAsText, 2)},
		},
		{
			name:   "get metadata by id - scope id does not exist",
			args:   []string{"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
			expOut: []string{"not_found:\n- scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
		},
		{
			name:   "get metadata by id - session id as json",
			args:   []string{s.sessionID.String(), s.asJson},
			expOut: []string{s.sessionAsJson},
		},
		{
			name:   "get metadata by id - session id as text",
			args:   []string{s.sessionID.String(), s.asText},
			expOut: []string{entryIndent(s.sessionAsText, 2)},
		},
		{
			name:   "get metadata by id - session id does not exist",
			args:   []string{"session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"},
			expOut: []string{"not_found:\n- session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"},
		},
		{
			name:   "get metadata by id - record id as json",
			args:   []string{s.recordID.String(), s.asJson},
			expOut: []string{s.recordAsJson},
		},
		{
			name:   "get metadata by id - record id as text",
			args:   []string{s.recordID.String(), s.asText},
			expOut: []string{entryIndent(s.recordAsText, 2)},
		},
		{
			name:   "get metadata by id - record id does not exist",
			args:   []string{"record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3"},
			expOut: []string{"not_found:\n- record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3"},
		},
		{
			name:   "get metadata by id - scope spec id as json",
			args:   []string{s.scopeSpecID.String(), s.asJson},
			expOut: []string{s.scopeSpecAsJson},
		},
		{
			name:   "get metadata by id - scope spec id as text",
			args:   []string{s.scopeSpecID.String(), s.asText},
			expOut: []string{entryIndent(s.scopeSpecAsText, 2)},
		},
		{
			name:   "get metadata by id - scope spec id does not exist",
			args:   []string{"scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"},
			expOut: []string{"not_found:\n- scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"},
		},
		{
			name:   "get metadata by id - contract spec id as json",
			args:   []string{s.contractSpecID.String(), s.asJson},
			expOut: []string{s.contractSpecAsJson},
		},
		{
			name:   "get metadata by id - contract spec id as text",
			args:   []string{s.contractSpecID.String(), s.asText},
			expOut: []string{entryIndent(s.contractSpecAsText, 2)},
		},
		{
			name:   "get metadata by id - contract spec id does not exist",
			args:   []string{"contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"},
			expOut: []string{"not_found:\n- contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"},
		},
		{
			name:   "get metadata by id - record spec id as json",
			args:   []string{s.recordSpecID.String(), s.asJson},
			expOut: []string{s.recordSpecAsJson},
		},
		{
			name:   "get metadata by id - record spec id as text",
			args:   []string{s.recordSpecID.String(), s.asText},
			expOut: []string{entryIndent(s.recordSpecAsText, 2)},
		},
		{
			name:   "get metadata by id - record spec id does not exist",
			args:   []string{"recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44"},
			expOut: []string{"not_found:\n- recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44"},
		},
		{
			name:   "get metadata by id - bad prefix",
			args:   []string{"foo1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
			expErr: "decoding bech32 failed: invalid checksum (expected kzwk8c got xlkwel)",
		},
		{
			name:   "get metadata by id - no args",
			args:   []string{},
			expErr: "requires at least 1 arg(s), only received 0",
		},
		{
			name:   "get metadata by id - two args",
			args:   []string{"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel", "scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"},
			expOut: []string{"not_found:\n- scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel\n- scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"},
		},
		{
			name:   "get metadata by id - uuid",
			args:   []string{"91978ba2-5f35-459a-86a7-feca1b0512e0"},
			expErr: "decoding bech32 failed: invalid separator index 32",
		},
		{
			name:   "get metadata by id - bad arg",
			args:   []string{"not-an-id"},
			expErr: "decoding bech32 failed: invalid separator index -1",
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
				testCases = append(testCases, []queryCmdTestCase{
					{
						name:   testName(bases[0], bi, ni, name, "as text"),
						args:   []string{name, s.asText},
						expOut: asText,
					},
					{
						name:   testName(bases[0], bi, ni, name, "as json"),
						args:   []string{name, s.asJson},
						expOut: asJson,
					},
				}...)
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
	testCases = append(testCases, []queryCmdTestCase{
		{
			name:   "all scopespecs spaced args 1 scope specs as text",
			args:   []string{"scope", "specs", s.asText},
			expOut: []string{indentedScopeSpecText},
		},
		{
			name:   "all scopespecs spaced args 2 scope specification as json",
			args:   []string{"scope", "specification", s.asJson},
			expOut: []string{s.scopeSpecAsJson},
		},
		{
			name:   "all scopespecs spaced args 3 scop espec as json",
			args:   []string{"scop", "espec", s.asJson},
			expOut: []string{s.scopeSpecAsJson},
		},
	}...)

	cSpecInputs := makeSpecInputs("contract")
	cSpecInputs = append(cSpecInputs, "cspecs", "cspec", "c-specs", "c-spec", "c specs", "c spec")
	addTestCases(cSpecInputs, []string{indentedContractSpecText}, []string{s.contractSpecAsJson})
	testCases = append(testCases, []queryCmdTestCase{
		{
			name:   "all contractspecs spaced args 1 contract specs as text",
			args:   []string{"contract", "specs", s.asText},
			expOut: []string{indentedContractSpecText},
		},
		{
			name:   "all contractspecs spaced args 2 contract specification as json",
			args:   []string{"contract", "specification", s.asJson},
			expOut: []string{s.contractSpecAsJson},
		},
		{
			name:   "all contractspecs spaced args 3 cs pec as json",
			args:   []string{"cs", "pec", s.asJson},
			expOut: []string{s.contractSpecAsJson},
		},
	}...)

	addTestCases(makeSpecInputs("record", "rec"), []string{indentedRecordSpecText}, []string{s.recordSpecAsJson})
	testCases = append(testCases, []queryCmdTestCase{
		{
			name:   "all recordspecs spaced args 1 record specs as text",
			args:   []string{"record", "specs", s.asText},
			expOut: []string{indentedRecordSpecText},
		},
		{
			name:   "all recordspecs spaced args 2 record specification as json",
			args:   []string{"record", "specification", s.asJson},
			expOut: []string{s.recordSpecAsJson},
		},
		{
			name:   "all recordspecs spaced args 3 recor dspec as json",
			args:   []string{"recor", "dspec", s.asJson},
			expOut: []string{s.recordSpecAsJson},
		},
	}...)

	addTestCases(
		[]string{"locators", "locator", "locs", "loc"},
		[]string{indentedOSLoc1Text, indentedOSLoc2Text},
		[]string{s.objectLocator1AsJson, s.objectLocator2AsJson},
	)

	testCases = append(testCases, []queryCmdTestCase{
		{
			name:   "unknown type",
			args:   []string{"scoops"},
			expErr: "unknown entry type: scoops",
		},
		{
			name:   "unknown type many args",
			args:   []string{"r", "e", "d", "o", "r", "k", "   ", "sp", "o", "rk"},
			expErr: "unknown entry type: redorksporks",
		},
		{
			name:   "no args",
			args:   []string{},
			expErr: "requires at least 1 arg(s), only received 0",
		},
	}...)

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetMetadataScopeCmd() {
	cmd := func() *cobra.Command { return cli.GetMetadataScopeCmd() }

	indentedScopeText := indent(s.scopeAsText, 4)

	testCases := []queryCmdTestCase{
		{
			name:   "get scope by metadata scope id as json output",
			args:   []string{s.scopeID.String(), s.asJson},
			expOut: []string{s.scopeAsJson},
		},
		{
			name:   "get scope by metadata scope id as text output",
			args:   []string{s.scopeID.String(), s.asText},
			expOut: []string{indentedScopeText},
		},
		{
			name:   "get scope by uuid as json output",
			args:   []string{s.scopeUUID.String(), s.asJson},
			expOut: []string{s.scopeAsJson},
		},
		{
			name:   "get scope by uuid as text output",
			args:   []string{s.scopeUUID.String(), s.asText},
			expOut: []string{indentedScopeText},
		},
		{
			name:   "get scope by metadata session id as json output",
			args:   []string{s.sessionID.String(), s.asJson},
			expOut: []string{s.scopeAsJson},
		},
		{
			name:   "get scope by metadata session id as text output",
			args:   []string{s.sessionID.String(), s.asText},
			expOut: []string{indentedScopeText},
		},
		{
			name:   "get scope by metadata record id as json output",
			args:   []string{s.recordID.String(), s.asJson},
			expOut: []string{s.scopeAsJson},
		},
		{
			name:   "get scope by metadata record id as text output",
			args:   []string{s.recordID.String(), s.asText},
			expOut: []string{indentedScopeText},
		},
		{
			name:   "get scope by metadata id - does not exist",
			args:   []string{"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel", s.asText},
			expOut: []string{"scope: null", "scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
		},
		{
			name:   "get scope by uuid - does not exist",
			args:   []string{"91978ba2-5f35-459a-86a7-feca1b0512e0", s.asText},
			expOut: []string{"scope: null", "91978ba2-5f35-459a-86a7-feca1b0512e0"},
		},
		{
			name:   "get scope bad input",
			args:   []string{"not-a-valid-arg", s.asText},
			expErr: "could not parse [not-a-valid-arg] into either a scope address (decoding bech32 failed: invalid separator index -1) or uuid (invalid UUID length: 15): invalid request",
		},
		{
			name:   "get scope no args",
			args:   []string{},
			expErr: "accepts 1 arg(s), received 0",
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
			name:   "session from session id as json",
			args:   []string{s.sessionID.String(), s.asJson},
			expOut: []string{s.sessionAsJson},
		},
		{
			name:   "session from session id as text",
			args:   []string{s.sessionID.String(), s.asText},
			expOut: []string{indentedSessionText},
		},
		{
			name:   "session id does not exist",
			args:   []string{"session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"},
			expOut: []string{"session: null", "session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"},
		},
		{
			name:   "sessions from scope id as json",
			args:   []string{s.scopeID.String(), s.asJson},
			expOut: []string{s.sessionAsJson},
		},
		{
			name:   "sessions from scope id as text",
			args:   []string{s.scopeID.String(), s.asText},
			expOut: []string{indentedSessionText},
		},
		{
			name:   "scope id does not exist",
			args:   []string{"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
			expErr: "no sessions found",
		},
		{
			name:   "sessions from scope uuid as json",
			args:   []string{s.scopeUUID.String(), s.asJson},
			expOut: []string{s.sessionAsJson},
		},
		{
			name:   "sessions from scope uuid as text",
			args:   []string{s.scopeUUID.String(), s.asText},
			expOut: []string{indentedSessionText},
		},
		{
			name:   "scope uuid does not exist",
			args:   []string{"91978ba2-5f35-459a-86a7-feca1b0512e0"},
			expErr: "no sessions found",
		},
		{
			name:   "session from scope uuid and session uuid as json",
			args:   []string{s.scopeUUID.String(), s.sessionUUID.String(), s.asJson},
			expOut: []string{s.sessionAsJson},
		},
		{
			name:   "session from scope uuid and session uuid as text",
			args:   []string{s.scopeUUID.String(), s.sessionUUID.String(), s.asText},
			expOut: []string{indentedSessionText},
		},
		{
			name:   "scope uuid and session uuid but scope does not exist",
			args:   []string{notAUsedUUID.String(), s.sessionUUID.String()},
			expOut: []string{"session:", "session: null"},
		},
		{
			name:   "scope uuid and session uuid and scope exists but session uuid does not exist",
			args:   []string{s.scopeUUID.String(), "5803f8bc-6067-4eb5-951f-2121671c2ec0"},
			expOut: []string{"session:", "session: null"},
		},
		{
			name:   "session from scope id and session uuid as text",
			args:   []string{s.scopeID.String(), s.sessionUUID.String(), s.asText},
			expOut: []string{indentedSessionText},
		},
		{
			name:   "session from scope id and session uuid as json",
			args:   []string{s.scopeID.String(), s.sessionUUID.String(), s.asJson},
			expOut: []string{s.sessionAsJson},
		},
		{
			name:   "scope id and session uuid but scope id does not exist",
			args:   []string{metadatatypes.ScopeMetadataAddress(notAUsedUUID).String(), s.sessionUUID.String()},
			expOut: []string{"session:", "session: null"},
		},
		{
			name:   "scope id and session uuid and scope id exists but session uuid does not",
			args:   []string{s.scopeID.String(), notAUsedUUID.String()},
			expOut: []string{"session:", "session: null"},
		},
		{
			name:   "session from scope id and record name as text",
			args:   []string{s.scopeID.String(), s.recordName, s.asText},
			expOut: []string{indentedSessionText},
		},
		{
			name:   "session from scope id and record name as json",
			args:   []string{s.scopeID.String(), s.recordName, s.asJson},
			expOut: []string{s.sessionAsJson},
		},
		{
			name: "scope id and record name but scope id does not exist",
			args: []string{metadatatypes.ScopeMetadataAddress(notAUsedUUID).String(), s.recordName},
			expErr: fmt.Sprintf("record %s does not exist: invalid request",
				metadatatypes.RecordMetadataAddress(notAUsedUUID, s.recordName)),
		},
		{
			name: "scope id and record name and scope id exists but record does not",
			args: []string{s.scopeID.String(), "not-a-record-name-that-exists"},
			expErr: fmt.Sprintf("record %s does not exist: invalid request",
				metadatatypes.RecordMetadataAddress(s.scopeUUID, "not-a-record-name-that-exists")),
		},
		{
			name:   "session from scope uuid and record name as text",
			args:   []string{s.scopeUUID.String(), s.recordName, s.asText},
			expOut: []string{indentedSessionText},
		},
		{
			name:   "session from scope uuid and record name as json",
			args:   []string{s.scopeUUID.String(), s.recordName, s.asJson},
			expOut: []string{s.sessionAsJson},
		},
		{
			name: "scope uuid and record name but scope uuid does not exist",
			args: []string{notAUsedUUID.String(), s.recordName},
			expErr: fmt.Sprintf("record %s does not exist: invalid request",
				metadatatypes.RecordMetadataAddress(notAUsedUUID, s.recordName)),
		},
		{
			name: "scope uuid and record name and scope uuid exists but record does not",
			args: []string{s.scopeUUID.String(), "not-a-record"},
			expErr: fmt.Sprintf("record %s does not exist: invalid request",
				metadatatypes.RecordMetadataAddress(s.scopeUUID, "not-a-record")),
		},
		{
			name:   "session from record id as text",
			args:   []string{s.recordID.String(), s.asText},
			expOut: []string{indentedSessionText},
		},
		{
			name:   "session from record id as json",
			args:   []string{s.recordID.String(), s.asJson},
			expOut: []string{s.sessionAsJson},
		},
		{
			name: "record id but scope does not exist",
			args: []string{metadatatypes.RecordMetadataAddress(notAUsedUUID, s.recordName).String()},
			expErr: fmt.Sprintf("record %s does not exist: invalid request",
				metadatatypes.RecordMetadataAddress(notAUsedUUID, s.recordName)),
		},
		{
			name: "record id in existing scope but record does not exist",
			args: []string{metadatatypes.RecordMetadataAddress(s.scopeUUID, "not-a-record-name").String()},
			expErr: fmt.Sprintf("record %s does not exist: invalid request",
				metadatatypes.RecordMetadataAddress(s.scopeUUID, "not-a-record-name")),
		},
		{
			name:   "sessions all as text",
			args:   []string{"all", s.asText},
			expOut: []string{indentedSessionText},
		},
		{
			name:   "sessions all as json",
			args:   []string{"all", s.asJson},
			expOut: []string{s.sessionAsJson},
		},
		{
			name:   "bad prefix",
			args:   []string{"scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"},
			expErr: "address [scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m] is not a scope address: invalid request",
		},
		{
			name:   "bad arg 1",
			args:   []string{"bad"},
			expErr: "could not parse [bad] into either a scope address (decoding bech32 failed: invalid bech32 string length 3) or uuid (invalid UUID length: 3): invalid request",
		},
		{
			name:   "3 args",
			args:   []string{s.scopeUUID.String(), s.sessionID.String(), s.recordName},
			expErr: "accepts between 1 and 2 arg(s), received 3",
		},
		{
			name:   "no args",
			args:   []string{},
			expErr: "accepts between 1 and 2 arg(s), received 0",
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetMetadataRecordCmd() {
	cmd := func() *cobra.Command { return cli.GetMetadataRecordCmd() }

	testCases := []queryCmdTestCase{
		{
			name:   "record from record id as json",
			args:   []string{s.recordID.String(), s.asJson},
			expOut: []string{s.recordAsJson},
		},
		{
			name:   "record from record id as text",
			args:   []string{s.recordID.String(), s.asText},
			expOut: []string{indent(s.recordAsText, 4)},
		},
		{
			name:   "record id does not exist",
			args:   []string{"record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3"},
			expOut: []string{"records:", "record: null", "record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3"},
		},
		{
			name:   "records from session id as json",
			args:   []string{s.sessionID.String(), s.asJson},
			expOut: []string{s.recordAsJson},
		},
		{
			name:   "records from session id as text",
			args:   []string{s.sessionID.String(), s.asText},
			expOut: []string{indent(s.recordAsText, 4)},
		},
		{
			name:   "session id does not exist",
			args:   []string{"session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"},
			expOut: []string{"records: []"},
		},
		{
			name:   "records from scope id as json",
			args:   []string{s.scopeID.String(), s.asJson},
			expOut: []string{s.recordAsJson},
		},
		{
			name:   "records from scope id as text",
			args:   []string{s.scopeID.String(), s.asText},
			expOut: []string{indent(s.recordAsText, 4)},
		},
		{
			name:   "scope id does not exist",
			args:   []string{"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
			expOut: []string{"records: []"},
		},
		{
			name:   "records from scope uuid as json",
			args:   []string{s.scopeUUID.String(), s.asJson},
			expOut: []string{s.recordAsJson},
		},
		{
			name:   "records from scope uuid as text",
			args:   []string{s.scopeUUID.String(), s.asText},
			expOut: []string{indent(s.recordAsText, 4)},
		},
		{
			name:   "scope uuid does not exist",
			args:   []string{"91978ba2-5f35-459a-86a7-feca1b0512e0"},
			expOut: []string{"records: []"},
		},
		{
			name:   "record from scope uuid and record name as json",
			args:   []string{s.scopeUUID.String(), s.recordName, s.asJson},
			expOut: []string{s.recordAsJson},
		},
		{
			name:   "record from scope uuid and record name as text",
			args:   []string{s.scopeUUID.String(), s.recordName, s.asText},
			expOut: []string{indent(s.recordAsText, 4)},
		},
		{
			name:   "scope uuid exists but record name does not",
			args:   []string{s.scopeUUID.String(), "nope"},
			expOut: []string{"record: null"},
		},
		{
			name:   "bad prefix",
			args:   []string{"contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"},
			expErr: "address [contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn] is not a scope address: invalid request",
		},
		{
			name:   "bad arg 1",
			args:   []string{"badbad"},
			expErr: "could not parse [badbad] into either a scope address (decoding bech32 failed: invalid bech32 string length 6) or uuid (invalid UUID length: 6): invalid request",
		},
		{
			name:   "uuid arg 1 and whitespace args 2 and 3 as json",
			args:   []string{s.scopeUUID.String(), "  ", " ", s.asJson},
			expOut: []string{s.recordAsJson},
		},
		{
			name:   "uuid arg 1 and whitespace args 2 and 3 as text",
			args:   []string{s.scopeUUID.String(), "  ", " ", s.asText},
			expOut: []string{indent(s.recordAsText, 4)},
		},
		{
			name:   "no args",
			args:   []string{},
			expErr: "requires at least 1 arg(s), only received 0",
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetMetadataScopeSpecCmd() {
	cmd := func() *cobra.Command { return cli.GetMetadataScopeSpecCmd() }

	testCases := []queryCmdTestCase{
		{
			name:   "scope spec from scope spec id as json",
			args:   []string{s.scopeSpecID.String(), s.asJson},
			expOut: []string{s.scopeSpecAsJson},
		},
		{
			name:   "scope spec from scope spec id as text",
			args:   []string{s.scopeSpecID.String(), s.asText},
			expOut: []string{indent(s.scopeSpecAsText, 4)},
		},
		{
			name:   "scope spec id bad prefix",
			args:   []string{"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
			expErr: "address [scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel] is not a scope spec address: invalid request",
		},
		{
			name:   "scope spec id does not exist",
			args:   []string{"scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"},
			expOut: []string{"specification: null", "scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"},
		},
		{
			name:   "scope spec from scope spec uuid as json",
			args:   []string{s.scopeSpecUUID.String(), s.asJson},
			expOut: []string{s.scopeSpecAsJson},
		},
		{
			name:   "scope spec from scope spec uuid as text",
			args:   []string{s.scopeSpecUUID.String(), s.asText},
			expOut: []string{indent(s.scopeSpecAsText, 4)},
		},
		{
			name:   "scope spec uuid does not exist",
			args:   []string{"dc83ea70-eacd-40fe-9adf-1cf6148bf8a2"},
			expOut: []string{"specification: null", "dc83ea70-eacd-40fe-9adf-1cf6148bf8a2"},
		},
		{
			name:   "bad arg",
			args:   []string{"reallybad"},
			expErr: "could not parse [reallybad] into either a scope spec address (decoding bech32 failed: invalid separator index -1) or uuid (invalid UUID length: 9): invalid request",
		},
		{
			name:   "two args",
			args:   []string{"arg1", "arg2"},
			expErr: "accepts 1 arg(s), received 2",
		},
		{
			name:   "no args",
			args:   []string{},
			expErr: "accepts 1 arg(s), received 0",
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetMetadataContractSpecCmd() {
	cmd := func() *cobra.Command { return cli.GetMetadataContractSpecCmd() }

	testCases := []queryCmdTestCase{
		{
			name:   "contract spec from contract spec id as json",
			args:   []string{s.contractSpecID.String(), s.asJson},
			expOut: []string{s.contractSpecAsJson},
		},
		{
			name:   "contract spec from contract spec id as text",
			args:   []string{s.contractSpecID.String(), s.asText},
			expOut: []string{indent(s.contractSpecAsText, 4)},
		},
		{
			name:   "contract spec id does not exist",
			args:   []string{"contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"},
			expOut: []string{"specification: null", "contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"},
		},
		{
			name:   "contract spec from contract spec uuid as json",
			args:   []string{s.contractSpecUUID.String(), s.asJson},
			expOut: []string{s.contractSpecAsJson},
		},
		{
			name:   "contract spec from contract spec uuid as text",
			args:   []string{s.contractSpecUUID.String(), s.asText},
			expOut: []string{indent(s.contractSpecAsText, 4)},
		},
		{
			name:   "contract spec uuid does not exist",
			args:   []string{"def6bc0a-c9dd-4874-948f-5206e6060a84"},
			expOut: []string{"specification: null", "def6bc0a-c9dd-4874-948f-5206e6060a84"},
		},
		{
			name:   "contract spec from record spec id as json",
			args:   []string{s.recordSpecID.String(), s.asJson},
			expOut: []string{s.contractSpecAsJson},
		},
		{
			name:   "contract spec from record spec id as text",
			args:   []string{s.recordSpecID.String(), s.asText},
			expOut: []string{indent(s.contractSpecAsText, 4)},
		},
		{
			name:   "record spec id does not exist",
			args:   []string{"recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44"},
			expOut: []string{"specification: null"},
		},
		{
			name:   "bad prefix",
			args:   []string{"record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3"},
			expErr: "invalid specification id: address [record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3] is not a contract spec address: invalid request",
		},
		{
			name:   "bad arg",
			args:   []string{"badbadarg"},
			expErr: "invalid specification id: could not parse [badbadarg] into either a contract spec address (decoding bech32 failed: invalid separator index -1) or uuid (invalid UUID length: 9): invalid request",
		},
		{
			name:   "two args",
			args:   []string{"arg1", "arg2"},
			expErr: "accepts 1 arg(s), received 2",
		},
		{
			name:   "no args",
			args:   []string{},
			expErr: "accepts 1 arg(s), received 0",
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetMetadataRecordSpecCmd() {
	cmd := func() *cobra.Command { return cli.GetMetadataRecordSpecCmd() }

	testCases := []queryCmdTestCase{
		{
			name:   "record spec from rec spec id as json",
			args:   []string{s.recordSpecID.String(), s.asJson},
			expOut: []string{s.recordSpecAsJson},
		},
		{
			name:   "record spec from rec spec id as text",
			args:   []string{s.recordSpecID.String(), s.asText},
			expOut: []string{indent(s.recordSpecAsText, 4)},
		},
		{
			name:   "rec spec id does not exist",
			args:   []string{"recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44"},
			expOut: []string{"record_specifications: []"},
		},
		{
			name:   "record specs from contract spec id as json",
			args:   []string{s.contractSpecID.String(), s.asJson},
			expOut: []string{s.recordSpecAsJson},
		},
		{
			name:   "record specs from contract spec id as text",
			args:   []string{s.contractSpecID.String(), s.asText},
			expOut: []string{indent(s.recordSpecAsText, 4)},
		},
		{
			name:   "contract spec id does not exist",
			args:   []string{"contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"},
			expOut: []string{"record_specifications: []", "contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"},
		},
		{
			name:   "record specs from contract spec uuid as json",
			args:   []string{s.contractSpecUUID.String(), s.asJson},
			expOut: []string{s.recordSpecAsJson},
		},
		{
			name:   "record specs from contract spec uuid as text",
			args:   []string{s.contractSpecUUID.String(), s.asText},
			expOut: []string{indent(s.recordSpecAsText, 4)},
		},
		{
			name:   "contract spec uuid does not exist",
			args:   []string{"def6bc0a-c9dd-4874-948f-5206e6060a84"},
			expOut: []string{"record_specifications: []", "def6bc0a-c9dd-4874-948f-5206e6060a84"},
		},
		{
			name:   "record spec from contract spec uuid and record spec name as json",
			args:   []string{s.contractSpecUUID.String(), s.recordName, s.asJson},
			expOut: []string{s.recordSpecAsJson},
		},
		{
			name:   "record spec from contract spec uuid and record spec name as text",
			args:   []string{s.contractSpecUUID.String(), s.recordName, s.asText},
			expOut: []string{indent(s.recordSpecAsText, 4)},
		},
		{
			name:   "contract spec uuid exists record spec name does not",
			args:   []string{s.contractSpecUUID.String(), s.contractSpecUUID.String()},
			expOut: []string{"specification: null", s.contractSpecUUID.String()},
		},
		{
			name:   "record specs from contract spec uuid and only whitespace name args",
			args:   []string{s.contractSpecUUID.String(), "   ", " ", "      "},
			expOut: []string{indent(s.recordSpecAsText, 4)},
		},
		{
			name:   "bad prefix",
			args:   []string{"session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"},
			expErr: "invalid specification id: address [session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr] is not a contract spec address: invalid request",
		},
		{
			name:   "bad arg 1",
			args:   []string{"not-gonna-parse"},
			expErr: "invalid specification id: could not parse [not-gonna-parse] into either a contract spec address (decoding bech32 failed: invalid separator index -1) or uuid (invalid UUID length: 15): invalid request",
		},
		{
			name:   "no args",
			args:   []string{s.asJson},
			expErr: "requires at least 1 arg(s), only received 0",
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
			name: "scopes as json",
			args: []string{s.user1AddrStr, s.asJson},
			expOut: []string{
				fmt.Sprintf("\"scope_uuids\":[\"%s\"]", s.scopeUUID),
				"\"pagination\":{\"next_key\":null,\"total\":\"0\"}",
			},
		},
		{
			name:   "scopes as text",
			args:   []string{s.user1AddrStr, s.asText},
			expOut: []string{scopeUUIDsText, paginationText},
		},
		{
			name:   "scope through value owner",
			args:   []string{s.user2AddrStr},
			expOut: []string{scopeUUIDsText},
		},
		{
			name:   "no result",
			args:   []string{newUser},
			expOut: []string{"scope_uuids: []", "total: \"0\""},
		},
		{
			name:   "two args",
			args:   []string{s.user1AddrStr, s.user2AddrStr},
			expErr: "accepts 1 arg(s), received 2",
		},
		{
			name:   "no args",
			args:   []string{},
			expErr: "accepts 1 arg(s), received 0",
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
			name: "as json",
			args: []string{s.user2AddrStr, s.asJson},
			expOut: []string{
				fmt.Sprintf("\"scope_uuids\":[\"%s\"]", s.scopeUUID),
				"\"pagination\":{\"next_key\":null,\"total\":\"0\"}",
			},
		},
		{
			name:   "as text",
			args:   []string{s.user2AddrStr, s.asText},
			expOut: []string{scopeUUIDsText, paginationText},
		},
		{
			name:   "no result",
			args:   []string{s.user1AddrStr},
			expOut: []string{"scope_uuids: []", "total: \"0\""},
		},
		{
			name:   "two args",
			args:   []string{s.user1AddrStr, s.user2AddrStr},
			expErr: "accepts 1 arg(s), received 2",
		},
		{
			name:   "no args",
			args:   []string{},
			expErr: "accepts 1 arg(s), received 0",
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
			name: "params as text",
			args: []string{"params", s.asText},
			expOut: []string{
				"params:",
				fmt.Sprintf("max_uri_length: %d", metadatatypes.DefaultMaxURILength),
			},
		},
		{
			name: "params as json",
			args: []string{"params", s.asJson},
			expOut: []string{
				"\"params\":{",
				fmt.Sprintf("\"max_uri_length\":%d", metadatatypes.DefaultMaxURILength),
			},
		},
		{
			name:   "all as text",
			args:   []string{"all", s.asText},
			expOut: []string{listEntryLocator1, listEntryLocator2},
		},
		{
			name:   "all as json",
			args:   []string{"all", s.asJson},
			expOut: []string{s.objectLocator1AsJson, s.objectLocator2AsJson},
		},
		{
			name:   "by owner locator 1 as text",
			args:   []string{s.user1AddrStr, s.asText},
			expOut: []string{indentedLocator1Text},
		},
		{
			name:   "by owner locator 1 as json",
			args:   []string{s.user1AddrStr, s.asJson},
			expOut: []string{s.objectLocator1AsJson},
		},
		{
			name:   "by owner locator 2 as text",
			args:   []string{s.user2AddrStr, s.asText},
			expOut: []string{indentedLocator2Text},
		},
		{
			name:   "by owner locator 2 as json",
			args:   []string{s.user2AddrStr, s.asJson},
			expOut: []string{s.objectLocator2AsJson},
		},
		{
			name:   "by owner unknown owner",
			args:   []string{s.userOtherAddr.String()},
			expErr: "no locator bound to address: unknown request",
		},
		{
			name:   "by scope id as text",
			args:   []string{s.scopeID.String(), s.asText},
			expOut: []string{listEntryLocator1},
		},
		{
			name:   "by scope id as json",
			args:   []string{s.scopeID.String(), s.asJson},
			expOut: []string{s.objectLocator1AsJson},
		},
		{
			name: "by scope id unknown scope id",
			args: []string{metadatatypes.ScopeMetadataAddress(unknownUUID).String()},
			expErr: fmt.Sprintf("scope [%s] not found: invalid request",
				metadatatypes.ScopeMetadataAddress(unknownUUID)),
		},
		{
			name:   "by scope uuid as text",
			args:   []string{s.scopeUUID.String(), s.asText},
			expOut: []string{listEntryLocator1},
		},
		{
			name:   "by scope uuid as json",
			args:   []string{s.scopeUUID.String(), s.asJson},
			expOut: []string{s.objectLocator1AsJson},
		},
		{
			name: "by scope uuid unknown scope uuid",
			args: []string{unknownUUID.String()},
			expErr: fmt.Sprintf("scope [%s] not found: invalid request",
				unknownUUID),
		},
		{
			name:   "by uri locator 1 as text",
			args:   []string{s.uri1, s.asText},
			expOut: []string{listEntryLocator1},
		},
		{
			name:   "by uri locator 1 as json",
			args:   []string{s.uri1, s.asJson},
			expOut: []string{s.objectLocator1AsJson},
		},
		{
			name:   "by uri locator 2 as text",
			args:   []string{s.uri2, s.asText},
			expOut: []string{listEntryLocator2},
		},
		{
			name:   "by uri locator 2 as json",
			args:   []string{s.uri2, s.asJson},
			expOut: []string{s.objectLocator2AsJson},
		},
		{
			name:   "by uri unknown uri",
			args:   []string{"http://not-an-entry.corn"},
			expErr: "No records found.: unknown request",
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetAccountDataCmd() {
	cmd := func() *cobra.Command { return cli.GetAccountDataCmd() }

	tests := []queryCmdTestCase{
		{
			name:   "invalid address",
			args:   []string{"notanaddr"},
			expErr: `invalid metadata address "notanaddr": decoding bech32 failed: invalid separator index -1`,
			expOut: []string{`invalid metadata address "notanaddr"`},
		},
		{
			name:   "data returned",
			args:   []string{s.scopeIDWithData.String()},
			expOut: []string{"value: This is some scope account data."},
		},
	}

	runQueryCmdTestCases(s, cmd, tests)
}

// ---------- tx cmd tests ----------

type txCmdTestCase struct {
	name         string
	cmd          *cobra.Command
	args         []string
	expectErrMsg string
	respType     proto.Message // You only need to define this if you're expecting something other than a TxResponse.
	expectedCode uint32
}

func runTxCmdTestCases(s *IntegrationCLITestSuite, testCases []txCmdTestCase) {
	s.T().Helper()
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			cmdName := tc.cmd.Name()
			var outBz []byte
			defer func() {
				if t.Failed() {
					t.Logf("Command: %s\nArgs: %q\nOutput:\n%s", cmdName, tc.args, string(outBz))
				}
			}()
			clientCtx := s.getClientCtx()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.cmd, tc.args)
			outBz = out.Bytes()

			if len(tc.expectErrMsg) > 0 {
				require.EqualError(t, err, tc.expectErrMsg, "%s expected error message", cmdName)
			} else {
				require.NoError(t, err, "%s unexpected error", cmdName)

				if tc.respType == nil {
					tc.respType = &sdk.TxResponse{}
				}
				umErr := clientCtx.Codec.UnmarshalJSON(outBz, tc.respType)
				require.NoError(t, umErr, "%s UnmarshalJSON error", cmdName)

				txResp, isTxResp := tc.respType.(*sdk.TxResponse)
				if isTxResp && !assert.Equal(t, int(tc.expectedCode), int(txResp.Code), "%s response code", cmdName) {
					// Note: If the above is failing because a 0 is expected, it might mean that the keeper method is returning an error.
					t.Logf("txResp:\n%v", txResp)
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
			expectedCode: 0,
		},
		{
			name: "should successfully add metadata scope",
			cmd:  cli.WriteScopeCmd(),
			args: []string{
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
			expectedCode: 0,
		},
		{
			name: "should successfully add metadata scope with signers flag",
			cmd:  cli.WriteScopeCmd(),
			args: []string{
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
			expectedCode: 0,
		},
		{
			name: "should successfully add metadata scope with party rollup",
			cmd:  cli.WriteScopeCmd(),
			args: []string{
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
			expectedCode: 0,
		},
		{
			name: "should fail to add metadata scope, incorrect scope id",
			cmd:  cli.WriteScopeCmd(),
			args: []string{
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
			expectErrMsg: "invalid scope id: decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "should fail to add metadata scope, incorrect scope spec id",
			cmd:  cli.WriteScopeCmd(),
			args: []string{
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
			expectErrMsg: "invalid spec id: decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "should fail to add metadata scope, validate basic will err on owner format",
			cmd:  cli.WriteScopeCmd(),
			args: []string{
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
			expectErrMsg: `invalid owners: invalid party "incorrect1,incorrect2": invalid address "incorrect1": decoding bech32 failed: invalid separator index 9`,
		},
		{
			name: "should fail to remove metadata scope, invalid scopeid",
			cmd:  cli.RemoveScopeCmd(),
			args: []string{
				"not-valid",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErrMsg: "decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "should fail to add/remove metadata scope data access, invalid scopeid",
			cmd:  cli.AddRemoveScopeDataAccessCmd(),
			args: []string{
				"add",
				"not-valid",
				s.user2AddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErrMsg: "decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "should fail to add/remove metadata scope data access, invalid command requires add or remove",
			cmd:  cli.AddRemoveScopeDataAccessCmd(),
			args: []string{
				"notaddorremove",
				scopeID,
				s.user2AddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErrMsg: "incorrect command notaddorremove : required remove or update",
		},
		{
			name: "should fail to add/remove metadata scope data access, not a scope id",
			cmd:  cli.AddRemoveScopeDataAccessCmd(),
			args: []string{
				"add",
				scopeSpecID,
				s.user2AddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErrMsg: fmt.Sprintf("meta address is not a scope: %s", scopeSpecID),
		},
		{
			name: "should fail to add/remove metadata scope data access, validatebasic fails",
			cmd:  cli.AddRemoveScopeDataAccessCmd(),
			args: []string{
				"add",
				scopeID,
				"notauser",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErrMsg: "data access address is invalid: notauser",
		},
		{
			name: "should successfully add metadata scope data access",
			cmd:  cli.AddRemoveScopeDataAccessCmd(),
			args: []string{
				"add",
				scopeID,
				s.user1AddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
		},
		{
			name: "should successfully remove metadata scope data access",
			cmd:  cli.AddRemoveScopeDataAccessCmd(),
			args: []string{
				"remove",
				scopeID,
				s.user1AddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
		},

		{
			name: "should fail to add/remove metadata scope owners, invalid scopeid",
			cmd:  cli.AddRemoveScopeOwnersCmd(),
			args: []string{
				"add",
				"not-valid",
				s.user2AddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErrMsg: "decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "should fail to add/remove metadata scope owner, invalid command requires add or remove",
			cmd:  cli.AddRemoveScopeOwnersCmd(),
			args: []string{
				"notaddorremove",
				scopeID,
				s.user2AddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErrMsg: "incorrect command notaddorremove : required remove or update",
		},
		{
			name: "should fail to add/remove metadata scope owner, not a scope id",
			cmd:  cli.AddRemoveScopeOwnersCmd(),
			args: []string{
				"add",
				scopeSpecID,
				s.user2AddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErrMsg: fmt.Sprintf("meta address is not a scope: %s", scopeSpecID),
		},
		{
			name: "should fail to add/remove metadata scope owner, validatebasic fails",
			cmd:  cli.AddRemoveScopeOwnersCmd(),
			args: []string{
				"add",
				scopeID,
				"notauser",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErrMsg: "invalid owners: invalid party address [notauser]: decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "should successfully remove metadata scope",
			cmd:  cli.RemoveScopeCmd(),
			args: []string{
				scopeID,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
		},
		{
			name: "should fail to delete metadata scope that no longer exists",
			cmd:  cli.RemoveScopeCmd(),
			args: []string{
				scopeID,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 18,
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
				name:   "scope 1 value owner",
				args:   []string{scopeID1},
				expOut: []string{"value_owner_address: " + scope1ValueOwner},
			},
			{
				name:   "scope 2 value owner",
				args:   []string{scopeID2},
				expOut: []string{"value_owner_address: " + scope2ValueOwner},
			},
			{
				name:   "scope 3 value owner",
				args:   []string{scopeID3},
				expOut: []string{"value_owner_address: " + scope3ValueOwner},
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
					expectErrMsg: "invalid signer address \"notabech32\": decoding bech32 failed: invalid separator index -1",
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
					expectErrMsg: "invalid signer address \"notabech32\": decoding bech32 failed: invalid separator index -1",
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
					expectedCode: 0,
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
					expectedCode: 0,
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
					expectedCode: 0,
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
					expectedCode: 0,
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
				expectedCode: 0,
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
				expectedCode: 0,
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
					expectedCode: 0,
				},
				{
					name: "update: scope 2 to user 2",
					cmd:  cli.UpdateValueOwnersCmd(),
					// <new value owner> <scope id> [<scope id 2> ...]
					args: []string{
						s.user2AddrStr, scopeID2,
						fromFlag(s.user3AddrStr), skipConfFlag, broadcastBlockFlag, feeFlag(10),
					},
					expectedCode: 0,
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
				expectedCode: 0,
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
				expectedCode: 0,
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
			expectErrMsg: "decoding bech32 failed: invalid bech32 string length 7",
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
			expectErrMsg: "invalid contract specification id prefix at index 0 (expected: contractspec, got scopespec)",
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
			expectErrMsg: `unknown party type: "badpartytype"`,
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
			expectErrMsg: "decoding bech32 failed: invalid separator index -1",
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
			name: "Should successfully add os locator",
			cmd:  cli.BindOsLocatorCmd(),
			args: []string{
				s.accountAddrStr,
				userURI,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
		},
		{
			name: "Should successfully Modify os locator",
			cmd:  cli.ModifyOsLocatorCmd(),
			args: []string{
				s.accountAddrStr,
				userURIMod,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
		},
		{
			name: "Should successfully delete os locator",
			cmd:  cli.RemoveOsLocatorCmd(),
			args: []string{
				s.accountAddrStr,
				userURIMod,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
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
			expectErrMsg: "decoding bech32 failed: invalid separator index -1",
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
			expectErrMsg: `unknown party type: "badpartytype"`,
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
			expectErrMsg: "decoding bech32 failed: invalid separator index -1",
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
			name: "should successfully add contract specification for test initialization",
			cmd:  cli.WriteContractSpecificationCmd(),
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
			expectedCode: 0,
		},
		{
			name: "should successfully add scope specification for test setup",
			cmd:  cli.WriteScopeSpecificationCmd(),
			args: []string{
				scopeSpecID.String(),
				s.accountAddrStr,
				"owner",
				s.contractSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
		},
		{
			name: "should fail to add contract spec to scope spec, invalid contract spec id",
			cmd:  addCommand,
			args: []string{
				"invalid-contract-specid",
				scopeSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErrMsg: "invalid contract specification id : invalid-contract-specid",
		},
		{
			name: "should fail to add contract spec to scope spec, not a contract spec id",
			cmd:  addCommand,
			args: []string{
				scopeSpecID.String(),
				scopeSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErrMsg: fmt.Sprintf("invalid contract specification id : %s", scopeSpecID.String()),
		},
		{
			name: "should fail to add contract spec to scope spec, invalid scope spec id",
			cmd:  addCommand,
			args: []string{
				specificationID.String(),
				"invalid-scope-spec-id",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErrMsg: "invalid scope specification id : invalid-scope-spec-id",
		},
		{
			name: "should fail to add contract spec to scope spec, not a scope spec",
			cmd:  addCommand,
			args: []string{
				specificationID.String(),
				specificationID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErrMsg: fmt.Sprintf("invalid scope specification id : %s", specificationID.String()),
		},
		{
			name: "should successfully add contract spec to scope spec",
			cmd:  addCommand,
			args: []string{
				specificationID.String(),
				scopeSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
		},
		{
			name: "should fail to remove contract spec to scope spec, invalid contract spec id",
			cmd:  removeCommand,
			args: []string{
				"invalid-contract-specid",
				scopeSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErrMsg: "invalid contract specification id : invalid-contract-specid",
		},
		{
			name: "should fail to remove contract spec to scope spec, not a contract spec id",
			cmd:  removeCommand,
			args: []string{
				scopeSpecID.String(),
				scopeSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErrMsg: fmt.Sprintf("invalid contract specification id : %s", scopeSpecID.String()),
		},
		{
			name: "should fail to remove contract spec to scope spec, invalid scope spec id",
			cmd:  removeCommand,
			args: []string{
				specificationID.String(),
				"invalid-scope-spec-id",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErrMsg: "invalid scope specification id : invalid-scope-spec-id",
		},
		{
			name: "should fail to remove contract spec to scope spec, not a scope spec id",
			cmd:  removeCommand,
			args: []string{
				specificationID.String(),
				specificationID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErrMsg: fmt.Sprintf("invalid scope specification id : %s", specificationID.String()),
		},
		{
			name: "should successfully remove contract spec to scope spec",
			cmd:  removeCommand,
			args: []string{
				specificationID.String(),
				scopeSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
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
			expectErrMsg: `unknown party type: "badpartytype"`,
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
			expectErrMsg: "record specification name cannot be empty",
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
			expectErrMsg: `invalid input specification "record1,typename1": expected 3 parts, have 2`,
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
			expectErrMsg: "record specification result type cannot be unspecified",
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
			expectErrMsg: "invalid signer address \"incorrect-signer-format\": decoding bech32 failed: invalid separator index -1",
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
			expectErrMsg: "decoding bech32 failed: invalid separator index -1",
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
			expectErrMsg: fmt.Sprintf("invalid contract specification id: %v", contractSpecID.String()),
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
			name: "should successfully add contract specification with resource hash for test setup",
			cmd:  cli.WriteContractSpecificationCmd(),
			args: []string{
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
			expectedCode: 0,
		},
		{
			name: "should successfully add scope specification for test setup",
			cmd:  cli.WriteScopeSpecificationCmd(),
			args: []string{
				scopeSpecID.String(),
				userAddress,
				"owner",
				s.contractSpecID.String() + "," + contractSpecID.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, userAddress),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
		},
		{
			name: "should successfully add metadata scope for test setup",
			cmd:  cli.WriteScopeCmd(),
			args: []string{
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
			expectedCode: 0,
		},
		{
			name: "should successfully add record specification for test setup",
			cmd:  cli.WriteRecordSpecificationCmd(),
			args: []string{
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
			expectedCode: 0,
		},
		{
			name: "should successfully add record with and create new session",
			cmd:  addRecordCmd,
			args: []string{
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
			expectedCode: 0,
		},
		{
			name: "should fail to add record incorrect scope id format",
			cmd:  addRecordCmd,
			args: []string{
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
			expectErrMsg: "decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "should fail to add record incorrect record id format",
			cmd:  addRecordCmd,
			args: []string{
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
			expectErrMsg: "decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "should fail to add record incorrect process format",
			cmd:  addRecordCmd,
			args: []string{
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
			expectErrMsg: `invalid process "hashvalue,methodname": expected 3 parts, have: 2`,
		},
		{
			name: "should fail to add record incorrect record inputs format",
			cmd:  addRecordCmd,
			args: []string{
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
			expectErrMsg: `invalid record input "input1name,typename1,proposed": expected 4 parts, have 3`,
		},
		{
			name: "should fail to add record incorrect record output format",
			cmd:  addRecordCmd,
			args: []string{
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
			expectErrMsg: `invalid record output "outputhashvalue": expected 2 parts, have 1`,
		},
		{
			name: "should fail to add record incorrect parties involved format",
			cmd:  addRecordCmd,
			args: []string{
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
			expectErrMsg: fmt.Sprintf(`invalid party "%s,%s": unknown party type: "%s"`, userAddress, userAddress, userAddress),
		},
		{
			name: "should fail to add record incorrect contract or session id format",
			cmd:  addRecordCmd,
			args: []string{
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
			expectErrMsg: fmt.Sprintf("id must be a contract or session id: %s", scopeID.String()),
		},
		{
			name: "should successfully remove record",
			cmd:  cli.RemoveRecordCmd(),
			args: []string{
				recordId.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
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
			name: "session-id no context",
			cmd:  cmd,
			args: []string{
				metadatatypes.SessionMetadataAddress(scopeUUID, uuid.New()).String(),
				s.contractSpecID.String(), fmt.Sprintf("%s,owner", owner), "somename",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, sender),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
		},
		{
			name: "scope-id session-uuid no context",
			cmd:  cmd,
			args: []string{
				scopeID.String(),
				uuid.New().String(),
				s.contractSpecID.String(), fmt.Sprintf("%s,owner", owner), "somename",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, sender),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
		},
		{
			name: "scope-uuid session-uuid no context",
			cmd:  cmd,
			args: []string{
				scopeUUID.String(),
				uuid.New().String(),
				s.contractSpecID.String(), fmt.Sprintf("%s,owner", owner), "somename",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, sender),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
		},
		{
			name: "session-id with context",
			cmd:  cmd,
			args: []string{
				metadatatypes.SessionMetadataAddress(scopeUUID, uuid.New()).String(),
				s.contractSpecID.String(), fmt.Sprintf("%s,owner", owner), "somename",
				"ChFIRUxMTyBQUk9WRU5BTkNFIQ==",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, sender),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
		},
		{
			name: "scope-id session-uuid with context",
			cmd:  cmd,
			args: []string{
				scopeID.String(),
				uuid.New().String(),
				s.contractSpecID.String(), fmt.Sprintf("%s,owner", owner), "somename",
				"ChFIRUxMTyBQUk9WRU5BTkNFIQ==",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, sender),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
		},
		{
			name: "scope-uuid session-uuid with context",
			cmd:  cmd,
			args: []string{
				scopeUUID.String(),
				uuid.New().String(),
				s.contractSpecID.String(), fmt.Sprintf("%s,owner", owner), "somename",
				"ChFIRUxMTyBQUk9WRU5BTkNFIQ==",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, sender),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
		},
		{
			name: "wrong id type",
			cmd:  cmd,
			args: []string{
				s.scopeSpecID.String(),
				s.contractSpecID.String(), fmt.Sprintf("%s,owner", owner), "somename",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, sender),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErrMsg: fmt.Sprintf("invalid address type in argument [%s]", s.scopeSpecID),
		},
		{
			name: "invalid first argument",
			cmd:  cmd,
			args: []string{
				"invalid",
				s.contractSpecID.String(), fmt.Sprintf("%s,owner", owner), "somename",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, sender),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErrMsg: fmt.Sprintf("argument [%s] is neither a bech32 address (%s) nor UUID (%s)", "invalid", "decoding bech32 failed: invalid bech32 string length 7", "invalid UUID length: 7"),
		},
		{
			name: "session-id with different context",
			cmd:  cmd,
			args: []string{
				metadatatypes.SessionMetadataAddress(scopeUUID, uuid.New()).String(),
				s.contractSpecID.String(), fmt.Sprintf("%s,owner", owner), "somename",
				"SEVMTE8gUFJPVkVOQU5DRSEK",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, sender),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 0,
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
			expectErrMsg: `invalid party "` + owner + `,badpartytype": unknown party type: "badpartytype"`,
		},
	}

	runTxCmdTestCases(s, testCases)
}

func (s *IntegrationCLITestSuite) TestSetAccountDataCmd() {
	cmd := func() *cobra.Command {
		return cli.SetAccountDataCmd()
	}

	scopeUUID := uuid.New()
	scopeID := metadatatypes.ScopeMetadataAddress(scopeUUID)
	scopeIDStr := scopeID.String()

	stdFlagsPlus := func(args ...string) []string {
		return append(args,
			fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		)
	}

	writeScopeCmd := cli.WriteScopeCmd()
	ctx := s.getClientCtx()
	out, err := clitestutil.ExecTestCLICmd(
		ctx,
		writeScopeCmd,
		stdFlagsPlus(
			scopeID.String(),
			s.scopeSpecID.String(),
			s.accountAddrStr,
			s.accountAddrStr,
			s.accountAddrStr,
		),
	)
	require.NoError(s.T(), err, "adding base scope")
	scopeResp := sdk.TxResponse{}
	umErr := ctx.Codec.UnmarshalJSON(out.Bytes(), &scopeResp)
	require.NoError(s.T(), umErr, "%s UnmarshalJSON error", writeScopeCmd.Name())
	if !s.Assert().Equal(0, int(scopeResp.Code), "write scope response code") {
		s.T().Logf("tx response:\n%v", scopeResp)
		s.T().FailNow()
	}

	tests := []txCmdTestCase{
		{
			name:         "invalid address",
			cmd:          cmd(),
			args:         stdFlagsPlus("notanaddr"),
			expectErrMsg: `invalid metadata address "notanaddr": decoding bech32 failed: invalid separator index -1`,
		},
		{
			name:         "no value",
			cmd:          cmd(),
			args:         stdFlagsPlus(scopeIDStr),
			expectErrMsg: "exactly one of these must be provided: " + attrcli.AccountDataFlagsUse,
		},
		{
			name: "invalid signers",
			cmd:  cmd(),
			args: stdFlagsPlus(
				scopeIDStr,
				"--"+attrcli.FlagValue, "Some new value.",
				"--"+cli.FlagSigners, s.accountAddrStr+",notanaddr",
			),
			expectErrMsg: "invalid signer address \"notanaddr\": decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "incorrect signer",
			cmd:  cmd(),
			args: []string{
				scopeIDStr,
				"--" + attrcli.FlagValue, "Some new value.",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.user1AddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectedCode: 18,
		},
		{
			name: "all okay",
			cmd:  cmd(),
			args: stdFlagsPlus(
				scopeIDStr,
				"--"+attrcli.FlagValue, "This is the account data for a test scope.",
			),
			expectedCode: 0,
		},
	}

	runTxCmdTestCases(s, tests)

	// Now look up what we just set to make sure it got set.
	runQueryCmdTestCases(s, func() *cobra.Command { return cli.GetAccountDataCmd() },
		[]queryCmdTestCase{
			{
				name:   "using metadata query command",
				args:   []string{scopeIDStr},
				expOut: []string{"value: This is the account data for a test scope."},
			},
		})
	runQueryCmdTestCases(s, func() *cobra.Command { return attrcli.GetAccountDataCmd() },
		[]queryCmdTestCase{
			{
				name:   "using attribute query command",
				args:   []string{scopeIDStr},
				expOut: []string{"value: This is the account data for a test scope."},
			},
		})
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
			expectedCode: 0,
		},
	}

	runTxCmdTestCases(s, testCases)
}
