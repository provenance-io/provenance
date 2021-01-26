package cli_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	tmcli "github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdksim "github.com/cosmos/cosmos-sdk/simapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"

	simapp "github.com/provenance-io/provenance/app"

	"github.com/provenance-io/provenance/x/attribute/client/cli"
	attributetypes "github.com/provenance-io/provenance/x/attribute/types"
	namecli "github.com/provenance-io/provenance/x/name/client/cli"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     testnet.Config
	testnet *testnet.Network

	accountAddr sdk.AccAddress
	accountKey  *secp256k1.PrivKey
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.accountKey = secp256k1.GenPrivKeyFromSecret([]byte("acc2"))
	addr, err := sdk.AccAddressFromHex(s.accountKey.PubKey().Address().String())
	s.Require().NoError(err)
	s.accountAddr = addr
	s.T().Log("setting up integration test suite")

	encCfg := simapp.MakeEncodingConfig()
	cfg := testnet.DefaultConfig()
	cfg.InterfaceRegistry = encCfg.InterfaceRegistry
	cfg.Codec = encCfg.Marshaler
	cfg.TxConfig = encCfg.TxConfig
	cfg.LegacyAmino = encCfg.Amino

	cfg.AppConstructor = func(val network.Validator) servertypes.Application {
		return simapp.New("test",
			val.Ctx.Logger, dbm.NewMemDB(), nil, true, make(map[int64]bool), val.Ctx.Config.RootDir, 0,
			encCfg,
			sdksim.EmptyAppOptions{},
			baseapp.SetPruning(storetypes.NewPruningOptionsFromString(val.AppConfig.Pruning)),
			baseapp.SetMinGasPrices(val.AppConfig.MinGasPrices),
		)
	}

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
		Address:       s.accountAddr.String(),
		AccountNumber: 1,
		Sequence:      0,
	})
	s.Require().NoError(err)
	authData.Accounts = append(authData.Accounts, genAccount)

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
	attributeData.Params.MaxValueLength = 32
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
			fmt.Sprintf(`{"account":"%s","attributes":[{"name":"example.attribute.count","value":"Mg==","attribute_type":"ATTRIBUTE_TYPE_INT","address":"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"},{"name":"example.attribute","value":"ZXhhbXBsZSBhdHRyaWJ1dGUgdmFsdWUgc3RyaW5n","attribute_type":"ATTRIBUTE_TYPE_STRING","address":"%s"}],"pagination":{"next_key":null,"total":"0"}}`, s.accountAddr.String(), s.accountAddr.String()),
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
			"{\"max_value_length\":32}",
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			"max_value_length: 32",
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

	testAttr, _ := codectypes.NewAnyWithValue(&attributetypes.Attribute{
		Address:       s.accountAddr.String(),
		AttributeType: 7,
		Name:          "test proto",
		Value:         []byte("test value"),
	})
	protoAttr, _ := testAttr.Marshal()

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
			"set attribute, valid int",
			cli.NewAddAccountAttributeCmd(),
			[]string{
				"txtest.attribute",
				s.testnet.Validators[0].Address.String(),
				"int",
				"1",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"set attribute, invalid int",
			cli.NewAddAccountAttributeCmd(),
			[]string{
				"txtest.attribute",
				s.testnet.Validators[0].Address.String(),
				"int",
				"3.14159",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 1,
		},
		{
			"set attribute, valid json",
			cli.NewAddAccountAttributeCmd(),
			[]string{
				"txtest.attribute",
				s.testnet.Validators[0].Address.String(),
				"json",
				"{\"key\":\"value\"}",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"set attribute, invalid json",
			cli.NewAddAccountAttributeCmd(),
			[]string{
				"txtest.attribute",
				s.testnet.Validators[0].Address.String(),
				"json",
				"{\"not proper json\"}",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 1,
		},
		{
			"set attribute, valid uuid",
			cli.NewAddAccountAttributeCmd(),
			[]string{
				"txtest.attribute",
				s.testnet.Validators[0].Address.String(),
				"uuid",
				"deadbeef-cafe-8675-3099-feedfacebabe",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"set attribute, invalid uuid",
			cli.NewAddAccountAttributeCmd(),
			[]string{
				"txtest.attribute",
				s.testnet.Validators[0].Address.String(),
				"uuid",
				"not-a-uuid-vegan-deadbeef-cafe",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 1,
		},
		{
			"set attribute, valid float",
			cli.NewAddAccountAttributeCmd(),
			[]string{
				"txtest.attribute",
				s.testnet.Validators[0].Address.String(),
				"float",
				"40.6",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"set attribute, invalid float",
			cli.NewAddAccountAttributeCmd(),
			[]string{
				"txtest.attribute",
				s.testnet.Validators[0].Address.String(),
				"float",
				"not a float",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 1,
		},
		{
			"set attribute, valid uri",
			cli.NewAddAccountAttributeCmd(),
			[]string{
				"txtest.attribute",
				s.testnet.Validators[0].Address.String(),
				"uri",
				"http://provenance.io",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"set attribute, invalid uri",
			cli.NewAddAccountAttributeCmd(),
			[]string{
				"txtest.attribute",
				s.testnet.Validators[0].Address.String(),
				"uri",
				"not a uri",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 1,
		},
		{
			"set attribute, valid bytes",
			cli.NewAddAccountAttributeCmd(),
			[]string{
				"txtest.attribute",
				s.testnet.Validators[0].Address.String(),
				"bytes",
				"ZXhhbXBsZSBhdHRyaWJ1dGUgdmFsdWUgc3RyaW5n",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"set attribute, valid proto",
			cli.NewAddAccountAttributeCmd(),
			[]string{
				"txtest.attribute",
				s.testnet.Validators[0].Address.String(),
				"proto",
				string(protoAttr),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"set attribute, unknown type",
			cli.NewAddAccountAttributeCmd(),
			[]string{
				"txtest.attribute",
				s.testnet.Validators[0].Address.String(),
				"notdog",
				"not a valid type",
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
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
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
		{"delete attribute, should delete txtest.attribute",
			cli.NewDeleteAccountAttributeCmd(),
			[]string{
				"deletetest.attribute",
				s.testnet.Validators[0].Address.String(),
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
				s.testnet.Validators[0].Address.String(),
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
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
