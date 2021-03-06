package cli_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil"

	"github.com/provenance-io/provenance/x/attribute/client/cli"
	attributetypes "github.com/provenance-io/provenance/x/attribute/types"
	namecli "github.com/provenance-io/provenance/x/name/client/cli"
	nametypes "github.com/provenance-io/provenance/x/name/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     testnet.Config
	testnet *testnet.Network

	accountAddr sdk.AccAddress
	accountKey  *secp256k1.PrivKey
	accountStr  string

	account2Addr sdk.AccAddress
	account2Key  *secp256k1.PrivKey
	account2Str  string
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.accountKey = secp256k1.GenPrivKeyFromSecret([]byte("acc2"))
	addr, err := sdk.AccAddressFromHex(s.accountKey.PubKey().Address().String())
	s.Require().NoError(err)
	s.accountAddr = addr
	s.accountStr = addr.String()

	s.account2Key = secp256k1.GenPrivKeyFromSecret([]byte("acc3"))
	addr, err = sdk.AccAddressFromHex(s.account2Key.PubKey().Address().String())
	s.Require().NoError(err)
	s.account2Addr = addr
	s.account2Str = addr.String()
	s.T().Log("setting up integration test suite")

	cfg := testutil.DefaultTestNetworkConfig()

	genesisState := cfg.GenesisState
	cfg.NumValidators = 1

	// Configure Genesis data for name module
	var nameData nametypes.GenesisState
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("attribute", s.accountAddr, false))
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("example.attribute", s.accountAddr, false))
	nameData.Params.AllowUnrestrictedNames = false
	nameData.Params.MaxNameLevels = 3
	nameData.Params.MinSegmentLength = 3
	nameData.Params.MaxSegmentLength = 12
	nameDataBz, err := cfg.Codec.MarshalJSON(&nameData)
	s.Require().NoError(err)
	genesisState[nametypes.ModuleName] = nameDataBz

	var authData authtypes.GenesisState
	s.Require().NoError(cfg.Codec.UnmarshalJSON(genesisState[authtypes.ModuleName], &authData))
	genAccount, err := codectypes.NewAnyWithValue(&authtypes.BaseAccount{
		Address:       s.accountStr,
		AccountNumber: 1,
		Sequence:      0,
	})
	s.Require().NoError(err)
	authData.Accounts = append(authData.Accounts, genAccount)
	authDataBz, err := cfg.Codec.MarshalJSON(&authData)
	s.Require().NoError(err)
	genesisState[authtypes.ModuleName] = authDataBz

	balances := sdk.NewCoins(
		sdk.NewCoin(cfg.BondDenom, cfg.AccountTokens),
	)
	var bankData banktypes.GenesisState
	s.Require().NoError(cfg.Codec.UnmarshalJSON(genesisState[banktypes.ModuleName], &bankData))
	genBank := banktypes.Balance{Address: s.accountStr, Coins: balances.Sort()}
	s.Require().NoError(err)
	bankData.Balances = append(bankData.Balances, genBank)
	bankDataBz, err := cfg.Codec.MarshalJSON(&bankData)
	s.Require().NoError(err)
	genesisState[banktypes.ModuleName] = bankDataBz

	// Configure Genesis data for attribute module
	var attributeData attributetypes.GenesisState
	attributeData.Attributes = append(attributeData.Attributes,
		attributetypes.NewAttribute(
			"example.attribute",
			s.accountAddr,
			attributetypes.AttributeType_String,
			[]byte("example attribute value string")))
	attributeData.Attributes = append(attributeData.Attributes,
		attributetypes.NewAttribute(
			"example.attribute.count",
			s.accountAddr,
			attributetypes.AttributeType_Int,
			[]byte("2")))
	attributeData.Params.MaxValueLength = 128
	attributeDataBz, err := cfg.Codec.MarshalJSON(&attributeData)
	s.Require().NoError(err)
	genesisState[attributetypes.ModuleName] = attributeDataBz

	cfg.GenesisState = genesisState

	s.cfg = cfg

	s.testnet = testnet.New(s.T(), cfg)

	_, err = s.testnet.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.testnet.WaitForNextBlock()
	s.T().Log("tearing down integration test suite")
	s.testnet.Cleanup()
}

func (s *IntegrationTestSuite) TestGetAccountAttributeCmd() {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"should get attribute by name with json output",
			[]string{s.accountAddr.String(), "example.attribute", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			fmt.Sprintf(`{"account":"%s","attributes":[{"name":"example.attribute","value":"ZXhhbXBsZSBhdHRyaWJ1dGUgdmFsdWUgc3RyaW5n","attribute_type":"ATTRIBUTE_TYPE_STRING","address":"%s"}],"pagination":{"next_key":null,"total":"0"}}`, s.accountAddr.String(), s.accountAddr.String()),
		},
		{
			"should get attribute by name with text output",
			[]string{s.accountAddr.String(), "example.attribute", fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			fmt.Sprintf(`account: %s
attributes:
- address: %s
  attribute_type: ATTRIBUTE_TYPE_STRING
  name: example.attribute
  value: ZXhhbXBsZSBhdHRyaWJ1dGUgdmFsdWUgc3RyaW5n
pagination:
  next_key: null
  total: "0"`, s.accountAddr.String(), s.accountAddr.String()),
		},
		{
			"should fail to find unknown attribute output",
			[]string{s.accountAddr.String(), "example.none", fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			fmt.Sprintf(`account: %s
attributes: []
pagination:
  next_key: null
  total: "0"`, s.accountAddr.String()),
		},
		{
			"should fail to find unknown attribute by name with json output",
			[]string{s.accountAddr.String(), "example.none", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			fmt.Sprintf(`{"account":"%s","attributes":[],"pagination":{"next_key":null,"total":"0"}}`, s.accountAddr.String()),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetAccountAttributeCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *IntegrationTestSuite) TestScanAccountAttributesCmd() {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"should get attribute by suffix with json output",
			[]string{s.accountAddr.String(), "attribute", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			fmt.Sprintf(`{"account":"%s","attributes":[{"name":"example.attribute","value":"ZXhhbXBsZSBhdHRyaWJ1dGUgdmFsdWUgc3RyaW5n","attribute_type":"ATTRIBUTE_TYPE_STRING","address":"%s"}],"pagination":{"next_key":null,"total":"0"}}`, s.accountAddr.String(), s.accountAddr.String()),
		},
		{
			"should get attribute by suffix with text output",
			[]string{s.accountAddr.String(), "attribute", fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			fmt.Sprintf(`account: %s
attributes:
- address: %s
  attribute_type: ATTRIBUTE_TYPE_STRING
  name: example.attribute
  value: ZXhhbXBsZSBhdHRyaWJ1dGUgdmFsdWUgc3RyaW5n
pagination:
  next_key: null
  total: "0"`, s.accountAddr.String(), s.accountAddr.String()),
		},
		{
			"should fail to find unknown attribute suffix text output",
			[]string{s.accountAddr.String(), "none", fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			fmt.Sprintf(`account: %s
attributes: []
pagination:
  next_key: null
  total: "0"`, s.accountAddr.String()),
		},
		{
			"should get attribute by suffix with json output",
			[]string{s.accountAddr.String(), "none", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			fmt.Sprintf(`{"account":"%s","attributes":[],"pagination":{"next_key":null,"total":"0"}}`, s.accountAddr.String()),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.ScanAccountAttributesCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *IntegrationTestSuite) TestListAccountAttributesCmd() {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"should list all attributes for account with json output",
			[]string{s.accountAddr.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			fmt.Sprintf(`{"account":"%s","attributes":[{"name":"example.attribute.count","value":"Mg==","attribute_type":"ATTRIBUTE_TYPE_INT","address":"%s"},{"name":"example.attribute","value":"ZXhhbXBsZSBhdHRyaWJ1dGUgdmFsdWUgc3RyaW5n","attribute_type":"ATTRIBUTE_TYPE_STRING","address":"%s"}],"pagination":{"next_key":null,"total":"0"}}`, s.accountAddr.String(), s.accountAddr.String(), s.accountAddr.String()),
		},
		{
			"should list all attributes for account text output",
			[]string{s.accountAddr.String(), fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			fmt.Sprintf(`account: %s
attributes:
- address: %s
  attribute_type: ATTRIBUTE_TYPE_INT
  name: example.attribute.count
  value: Mg==
- address: %s
  attribute_type: ATTRIBUTE_TYPE_STRING
  name: example.attribute
  value: ZXhhbXBsZSBhdHRyaWJ1dGUgdmFsdWUgc3RyaW5n
pagination:
  next_key: null
  total: "0"`, s.accountAddr.String(), s.accountAddr.String(), s.accountAddr.String()),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.ListAccountAttributesCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *IntegrationTestSuite) TestGetAttributeParamsCmd() {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			"{\"max_value_length\":128}",
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			"max_value_length: 128",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetAttributeParamsCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *IntegrationTestSuite) TestAttributeTxCommands() {

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"bind a new attribute name for testing",
			namecli.GetBindNameCmd(),
			[]string{
				"txtest",
				s.testnet.Validators[0].Address.String(),
				"attribute",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"set attribute, valid string",
			cli.NewAddAccountAttributeCmd(),
			[]string{
				"txtest.attribute",
				s.testnet.Validators[0].Address.String(),
				"string",
				"test value",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"set attribute, invalid bech32 address",
			cli.NewAddAccountAttributeCmd(),
			[]string{
				"txtest.attribute",
				"invalidbech32",
				"string",
				"test value",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"set attribute, invalid type",
			cli.NewAddAccountAttributeCmd(),
			[]string{
				"txtest.attribute",
				s.account2Addr.String(),
				"blah",
				"3.14159",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 1,
		},
		{
			"set attribute, cannot encode",
			cli.NewAddAccountAttributeCmd(),
			[]string{
				"txtest.attribute",
				s.testnet.Validators[0].Address.String(),
				"bytes",
				"3.14159",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 1,
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
				s.Require().NoError(clientCtx.JSONCodec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestUpdateAccountAttributeTxCommands() {

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"bind a new attribute name for delete testing",
			namecli.GetBindNameCmd(),
			[]string{
				"updatetest",
				s.testnet.Validators[0].Address.String(),
				"attribute",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"add new attribute for updating",
			cli.NewAddAccountAttributeCmd(),
			[]string{
				"updatetest.attribute",
				s.account2Addr.String(),
				"string",
				"test value",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"fail to update attribute, account address failure",
			cli.NewUpdateAccountAttributeCmd(),
			[]string{
				"updatetest.attribute",
				"not-an-address",
				"string",
				"test value",
				"int",
				"10",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"fail to update attribute, incorrect original type",
			cli.NewUpdateAccountAttributeCmd(),
			[]string{
				"updatetest.attribute",
				s.account2Addr.String(),
				"invalid",
				"test value",
				"int",
				"10",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"fail to update attribute, incorrect update type",
			cli.NewUpdateAccountAttributeCmd(),
			[]string{
				"updatetest.attribute",
				s.account2Addr.String(),
				"string",
				"test value",
				"invalid",
				"10",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"fail to update attribute, validate basic fail",
			cli.NewUpdateAccountAttributeCmd(),
			[]string{
				"updatetest.attribute",
				s.account2Addr.String(),
				"string",
				"test value",
				"init",
				"nan",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"successful update of attribute",
			cli.NewUpdateAccountAttributeCmd(),
			[]string{
				"updatetest.attribute",
				s.account2Addr.String(),
				"string",
				"test value",
				"int",
				"10",
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
				s.Require().NoError(clientCtx.JSONCodec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestDeleteDistinctAccountAttributeTxCommands() {

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"bind a new attribute name for delete testing",
			namecli.GetBindNameCmd(),
			[]string{
				"distinct",
				s.testnet.Validators[0].Address.String(),
				"attribute",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"add new attribute for delete testing",
			cli.NewAddAccountAttributeCmd(),
			[]string{
				"distinct.attribute",
				s.account2Addr.String(),
				"string",
				"test value",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{"delete distinct attribute, should fail incorrect address",
			cli.NewDeleteDistinctAccountAttributeCmd(),
			[]string{
				"distinct.attribute",
				"not-a-address",
				"string",
				"test value",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{"delete distinct attribute, should fail incorrect type",
			cli.NewDeleteDistinctAccountAttributeCmd(),
			[]string{
				"distinct.attribute",
				s.account2Addr.String(),
				"invalid",
				"test value",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{"delete distinct attribute, should successfully delete",
			cli.NewDeleteDistinctAccountAttributeCmd(),
			[]string{
				"distinct.attribute",
				s.account2Addr.String(),
				"string",
				"test value",
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
				s.Require().NoError(clientCtx.JSONCodec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestDeleteAccountAttributeTxCommands() {

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"bind a new attribute name for delete testing",
			namecli.GetBindNameCmd(),
			[]string{
				"deletetest",
				s.testnet.Validators[0].Address.String(),
				"attribute",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"add new attribute for delete testing",
			cli.NewAddAccountAttributeCmd(),
			[]string{
				"deletetest.attribute",
				s.account2Addr.String(),
				"string",
				"test value",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{"delete attribute, should delete txtest.attribute",
			cli.NewDeleteAccountAttributeCmd(),
			[]string{
				"deletetest.attribute",
				s.account2Addr.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{"delete attribute, should fail to find txtest.attribute",
			cli.NewDeleteAccountAttributeCmd(),
			[]string{
				"deletetest.attribute",
				s.account2Addr.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 1,
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
				s.Require().NoError(clientCtx.JSONCodec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
