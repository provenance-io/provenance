package cli_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	pioconfig "github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	testcli "github.com/provenance-io/provenance/testutil/cli"
	assetcli "github.com/provenance-io/provenance/x/asset/client/cli"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     testnet.Config
	testnet *testnet.Network
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up asset module CLI integration test suite")

	// Set Provenance config with the correct bond denom before creating the testnet config
	pioconfig.SetProvenanceConfig("stake", 0)

	s.cfg = testutil.DefaultTestNetworkConfig()
	s.cfg.NumValidators = 1
	// Explicitly set bond denom to stake for correct fee handling
	s.cfg.BondDenom = "stake"

	var err error
	s.testnet, err = testnet.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err, "creating testnet")
	_, err = testutil.WaitForHeight(s.testnet, 1)
	s.Require().NoError(err, "waiting for height 1")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.testnet != nil {
		s.testnet.Cleanup()
	}
}

func (s *IntegrationTestSuite) TestAssetQueryCommands() {
	clientCtx := s.testnet.Validators[0].ClientCtx
	valAddr := s.testnet.Validators[0].Address.String()

	testCases := []struct {
		name           string
		cmd            *cobra.Command
		args           []string
		expectedOutput string
		expectErr      bool
	}{
		{
			name:           "list asset classes (should be empty)",
			cmd:            assetcli.GetCmdListAssetClasses(),
			args:           []string{fmt.Sprintf("--%s=json", flags.FlagOutput)},
			expectedOutput: "\"assetClasses\":[]",
		},
		{
			name:           "list assets for valid address (should be empty)",
			cmd:            assetcli.GetCmdListAssets(),
			args:           []string{valAddr, fmt.Sprintf("--%s=json", flags.FlagOutput)},
			expectedOutput: "\"assets\":[]",
		},
		{
			name:           "get class with invalid id",
			cmd:            assetcli.GetCmdGetClass(),
			args:           []string{"notaclassid", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			expectedOutput: "",
			expectErr:      true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				if tc.expectedOutput != "" {
					s.Require().Contains(strings.ReplaceAll(out.String(), "\n", ""), strings.ReplaceAll(tc.expectedOutput, "\n", ""))
				}
			}
		})
	}

	// TODO: Add more advanced query tests after creating assets/classes
}

func (s *IntegrationTestSuite) TestAssetTxCommands() {
	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			name: "create asset class - missing args",
			cmd:  assetcli.GetCmdCreateAssetClass(),
			args: []string{
				// Missing required arguments
			},
			expectErr:    true,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "create asset class - all args",
			cmd:  assetcli.GetCmdCreateAssetClass(),
			args: []string{
				"test-class-id",
				"Test Asset Class",
				"TAC",
				"Test asset class description",
				"https://example.com/uri",
				"uri-hash-123",
				"{\"type\":\"object\",\"properties\":{\"name\":{\"type\":\"string\"},\"value\":{\"type\":\"integer\"}}}",
				"ledger-class-id-123",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 381000000)).String()),
				"--yes",
				"--keyring-backend=test",
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 1,
		},
		{
			name: "create asset - missing args",
			cmd:  assetcli.GetCmdCreateAsset(),
			args: []string{
				// Missing required arguments
			},
			expectErr:    true,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "create asset - all args",
			cmd:  assetcli.GetCmdCreateAsset(),
			args: []string{
				"test-class-id",
				"test-asset-id",
				"https://example.com/asset-uri",
				"asset-uri-hash-123",
				"{\"type\":\"object\",\"properties\":{\"name\":{\"type\":\"string\"},\"value\":{\"type\":\"integer\"}}}",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 381000000)).String()),
				"--yes",
				"--keyring-backend=test",
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 1,
		},
		{
			name: "create pool - missing args",
			cmd:  assetcli.GetCmdCreatePool(),
			args: []string{
				// Missing required arguments
			},
			expectErr:    true,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "create pool - all args (expected fail, needs setup)",
			cmd:  assetcli.GetCmdCreatePool(),
			args: []string{
				"10pooltoken",
				"class-id-1,asset-id-1;class-id-2,asset-id-2",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 381000000)).String()),
				"--yes",
				"--keyring-backend=test",
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 1,
		},
		{
			name: "create participation - missing args",
			cmd:  assetcli.GetCmdCreateParticipation(),
			args: []string{
				// Missing required arguments
			},
			expectErr:    true,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "create participation - all args (expected fail, needs setup)",
			cmd:  assetcli.GetCmdCreateParticipation(),
			args: []string{
				"pool-id-123",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 381000000)).String()),
				"--yes",
				"--keyring-backend=test",
			},
			expectErr:    true,
			respType:     &sdk.TxResponse{},
			expectedCode: 1,
		},
		{
			name: "create securitization - missing args",
			cmd:  assetcli.GetCmdCreateSecuritization(),
			args: []string{
				// Missing required arguments
			},
			expectErr:    true,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "create securitization - all args (expected fail, needs setup)",
			cmd:  assetcli.GetCmdCreateSecuritization(),
			args: []string{
				"sec-id-123",
				"pool-1,pool-2",
				"100tranche1,200tranche2",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 381000000)).String()),
				"--yes",
				"--keyring-backend=test",
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 11,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			testcli.NewTxExecutor(tc.cmd, tc.args).
				WithExpErr(tc.expectErr).
				WithExpCode(tc.expectedCode).
				Execute(s.T(), s.testnet)
		})
	}
}
