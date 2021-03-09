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
				require.Nil(t, err, "unexpected error")
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
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			"",
			"{}",
		},
		{
			"get params as text output",
			[]string{fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			"",
			"{}",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetMetadataParamsCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

// TODO: TestGetMetadataByIDCmd

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
			"get scope by uuid as json output",
			[]string{s.scopeUUID.String(), s.asJson},
			"",
			scopeAsJson,
		},
		{
			"get scope by metadata id as json output",
			[]string{s.scopeID.String(), s.asJson},
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
			"get scope by metadata id as text output",
			[]string{s.scopeID.String(), s.asText},
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
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

// TODO: GetMetadataSessionsByScopeCmd

// TODO: TestGetMetadataSessionCmd

// TODO: TestGetMetadataRecordCmd

// TODO: TestGetMetadataScopeSpecCmd

// TODO: TestGetMetadataContractSpecCmd

// TODO: TestGetMetadataRecordSpecCmd

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
