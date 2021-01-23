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

	markercli "github.com/provenance-io/provenance/x/marker/client/cli"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
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

	// Configure Genesis data for marker module
	var markerData markertypes.GenesisState
	markerData.Params.EnableGovernance = true
	markerData.Params.MaxTotalSupply = 1000000
	markerData.Markers = []markertypes.MarkerAccount{
		{
			BaseAccount: &authtypes.BaseAccount{
				Address:       markertypes.MustGetMarkerAddress("testcoin").String(),
				AccountNumber: 100,
				Sequence:      0,
			},
			Status:                 markertypes.StatusActive,
			SupplyFixed:            true,
			MarkerType:             markertypes.MarkerType_Coin,
			AllowGovernanceControl: false,
			Supply:                 sdk.NewInt(1000),
			Denom:                  "testcoin",
		},
		{
			BaseAccount: &authtypes.BaseAccount{
				Address:       markertypes.MustGetMarkerAddress("lockedcoin").String(),
				AccountNumber: 110,
				Sequence:      0,
			},
			Status:                 markertypes.StatusActive,
			SupplyFixed:            true,
			MarkerType:             markertypes.MarkerType_RestrictedCoin,
			AllowGovernanceControl: false,
			Supply:                 sdk.NewInt(1000),
			Denom:                  "lockedcoin",
		},
		{
			BaseAccount: &authtypes.BaseAccount{
				Address:       markertypes.MustGetMarkerAddress(cfg.BondDenom).String(),
				AccountNumber: 120,
				Sequence:      0,
			},
			Status:                 markertypes.StatusActive,
			SupplyFixed:            true,
			MarkerType:             markertypes.MarkerType_Coin,
			AllowGovernanceControl: true,
			Supply:                 cfg.BondedTokens.Mul(sdk.NewInt(int64(cfg.NumValidators))),
			Denom:                  cfg.BondDenom,
		},
	}
	markerDataBz, err := cfg.Codec.MarshalJSON(&markerData)
	s.Require().NoError(err)
	genesisState[markertypes.ModuleName] = markerDataBz

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

func (s *IntegrationTestSuite) TestMarkerQueryCommands() {
	testCases := []struct {
		name           string
		cmd            *cobra.Command
		args           []string
		expectedOutput string
	}{
		{
			"get testcoin marker",
			markercli.MarkerCmd(),
			[]string{
				"testcoin",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			`{"marker":{"@type":"/provenance.marker.v1.MarkerAccount","base_account":{"address":"cosmos1p3sl9tll0ygj3flwt5r2w0n6fx9p5ngq2tu6mq","pub_key":null,"account_number":"100","sequence":"0"},"manager":"","access_control":[],"status":"MARKER_STATUS_ACTIVE","denom":"testcoin","supply":"1000","marker_type":"MARKER_TYPE_COIN","supply_fixed":true,"allow_governance_control":false}}`,
		},
		{
			"get testcoin marker",
			markercli.MarkerCmd(),
			[]string{
				"testcoin",
				fmt.Sprintf("--%s=text", tmcli.OutputFlag),
			},
			`marker:
  '@type': /provenance.marker.v1.MarkerAccount
  access_control: []
  allow_governance_control: false
  base_account:
    account_number: "100"
    address: cosmos1p3sl9tll0ygj3flwt5r2w0n6fx9p5ngq2tu6mq
    pub_key: null
    sequence: "0"
  denom: testcoin
  manager: ""
  marker_type: MARKER_TYPE_COIN
  status: MARKER_STATUS_ACTIVE
  supply: "1000"
  supply_fixed: true`,
		},
		{
			"query non existant marker",
			markercli.MarkerCmd(),
			[]string{
				"doesntexist",
			},
			"",
		},
		{
			"get restricted  coin marker",
			markercli.MarkerCmd(),
			[]string{
				"lockedcoin",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			`{"marker":{"@type":"/provenance.marker.v1.MarkerAccount","base_account":{"address":"cosmos16437wt0xtqtuw0pn4vt8rlf8gr2plz2det0mt2","pub_key":null,"account_number":"110","sequence":"0"},"manager":"","access_control":[],"status":"MARKER_STATUS_ACTIVE","denom":"lockedcoin","supply":"1000","marker_type":"MARKER_TYPE_RESTRICTED","supply_fixed":true,"allow_governance_control":false}}`,
		},
		{
			"query access",
			markercli.MarkerAccessCmd(),
			[]string{
				s.cfg.BondDenom,
			},
			"accounts: []",
		},
		{
			"query escrow",
			markercli.MarkerEscrowCmd(),
			[]string{
				s.cfg.BondDenom,
			},
			"escrow: []",
		},
		{
			"query supply",
			markercli.MarkerSupplyCmd(),
			[]string{
				s.cfg.BondDenom,
			},
			fmt.Sprintf("amount:\n  amount: \"%s\"\n  denom: %s", s.cfg.BondedTokens.Mul(sdk.NewInt(int64(s.cfg.NumValidators))), s.cfg.BondDenom),
		},
	}
	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := s.testnet.Validators[0].ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *IntegrationTestSuite) TestMarkerTxCommands() {
	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"create a new marker",
			markercli.GetCmdAddMarker(),
			[]string{
				"1000hotdog",
				"--type=RESTRICTED",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"add single access",
			markercli.GetCmdAddAccess(),
			[]string{
				s.testnet.Validators[0].Address.String(),
				"hotdog",
				"admin",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"add multiple access",
			markercli.GetCmdAddAccess(),
			[]string{
				s.testnet.Validators[0].Address.String(),
				"hotdog",
				"mint,burn,transfer",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"mint supply",
			markercli.GetCmdMint(),
			[]string{
				"100hotdog",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"burn supply",
			markercli.GetCmdBurn(),
			[]string{
				"100hotdog",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"finalize",
			markercli.GetCmdFinalize(),
			[]string{
				"hotdog",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"activate",
			markercli.GetCmdActivate(),
			[]string{
				"hotdog",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"remove access",
			markercli.GetCmdDeleteAccess(),
			[]string{
				s.testnet.Validators[0].Address.String(),
				"hotdog",
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
				println(out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
