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
	pioconfig.SetProvConfig("stake")

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

	// Arrange: create classes and one asset to cover valid/empty states.
	{
		// Class with no assets.
		args := []string{
			"cli-class-empty",
			"Empty Class",
			"",
			"--symbol=EMPTY",
			"--description=Class with no assets",
			"--uri=https://example.com/empty",
			"--uri-hash=empty-hash",
			fmt.Sprintf("--%s=%s", flags.FlagFrom, valAddr),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%s", flags.FlagGasPrices, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 1)).String()),
			"--yes",
			"--keyring-backend=test",
		}
		testcli.NewTxExecutor(assetcli.GetCmdCreateAssetClass(), args).
			WithExpErr(false).
			WithExpCode(0).
			Execute(s.T(), s.testnet)
	}
	{
		// Class with one asset owned by validator.
		args := []string{
			"cli-class-1",
			"Test Class",
			"",
			"--symbol=TST",
			"--description=Class with one asset",
			"--uri=https://example.com/class1",
			"--uri-hash=class1-hash",
			fmt.Sprintf("--%s=%s", flags.FlagFrom, valAddr),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%s", flags.FlagGasPrices, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 1)).String()),
			"--yes",
			"--keyring-backend=test",
		}
		testcli.NewTxExecutor(assetcli.GetCmdCreateAssetClass(), args).
			WithExpErr(false).
			WithExpCode(0).
			Execute(s.T(), s.testnet)

		args = []string{
			"cli-class-1",
			"cli-asset-1",
			"",
			"--uri=https://example.com/asset1",
			"--uri-hash=asset1-hash",
			valAddr,
			fmt.Sprintf("--%s=%s", flags.FlagFrom, valAddr),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%s", flags.FlagGasPrices, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 1)).String()),
			"--yes",
			"--keyring-backend=test",
		}
		testcli.NewTxExecutor(assetcli.GetCmdCreateAsset(), args).
			WithExpErr(false).
			WithExpCode(0).
			Execute(s.T(), s.testnet)
	}

	// Asset (single) — valid and invalid data.
	{
		// Valid: existing asset.
		out, err := clitestutil.ExecTestCLICmd(clientCtx, assetcli.GetCmdAsset(), []string{"cli-class-1", "cli-asset-1", fmt.Sprintf("--%s=json", flags.FlagOutput)})
		s.Require().NoError(err)
		s.Require().Contains(strings.ReplaceAll(out.String(), "\n", ""), "\"asset\":{\"class_id\":\"cli-class-1\",\"id\":\"cli-asset-1\"")

		// Invalid: non-existent asset id.
		_, err = clitestutil.ExecTestCLICmd(clientCtx, assetcli.GetCmdAsset(), []string{"cli-class-1", "does-not-exist", fmt.Sprintf("--%s=json", flags.FlagOutput)})
		s.Require().Error(err)
	}

	// Assets (list) — valid and invalid data.
	{
		// Valid: address with assets.
		out, err := clitestutil.ExecTestCLICmd(clientCtx, assetcli.GetCmdAssets(), []string{valAddr, fmt.Sprintf("--%s=json", flags.FlagOutput)})
		s.Require().NoError(err)
		s.Require().Contains(strings.ReplaceAll(out.String(), "\n", ""), "\"id\":\"cli-asset-1\"")

		// Valid: class-id and address with no assets.
		out, err = clitestutil.ExecTestCLICmd(clientCtx, assetcli.GetCmdAssets(), []string{"cli-class-empty", valAddr, fmt.Sprintf("--%s=json", flags.FlagOutput)})
		s.Require().NoError(err)
		s.Require().Contains(strings.ReplaceAll(out.String(), "\n", ""), "\"assets\":[]")

		// Invalid: address format.
		_, err = clitestutil.ExecTestCLICmd(clientCtx, assetcli.GetCmdAssets(), []string{"cli-class-1", "badaddress", fmt.Sprintf("--%s=json", flags.FlagOutput)})
		s.Require().Error(err)
	}

	// Class — valid and invalid data.
	{
		// Valid: class with asset.
		out, err := clitestutil.ExecTestCLICmd(clientCtx, assetcli.GetCmdAssetClass(), []string{"cli-class-1", fmt.Sprintf("--%s=json", flags.FlagOutput)})
		s.Require().NoError(err)
		s.Require().Contains(strings.ReplaceAll(out.String(), "\n", ""), "\"class\":{\"id\":\"cli-class-1\"")

		// Valid: class without assets.
		out, err = clitestutil.ExecTestCLICmd(clientCtx, assetcli.GetCmdAssetClass(), []string{"cli-class-empty", fmt.Sprintf("--%s=json", flags.FlagOutput)})
		s.Require().NoError(err)
		s.Require().Contains(strings.ReplaceAll(out.String(), "\n", ""), "\"class\":{\"id\":\"cli-class-empty\"")

		// Invalid: non-existent class id.
		_, err = clitestutil.ExecTestCLICmd(clientCtx, assetcli.GetCmdAssetClass(), []string{"does-not-exist", fmt.Sprintf("--%s=json", flags.FlagOutput)})
		s.Require().Error(err)
	}
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
				"{\"type\":\"object\",\"properties\":{\"name\":{\"type\":\"string\"},\"value\":{\"type\":\"integer\"}}}",
				"--symbol=TAC",
				"--description=Test asset class description",
				"--uri=https://example.com/uri",
				"--uri-hash=uri-hash-123",
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
				"{\"type\":\"object\",\"properties\":{\"name\":{\"type\":\"string\"},\"value\":{\"type\":\"integer\"}}}",
				"--uri=https://example.com/asset-uri",
				"--uri-hash=asset-uri-hash-123",
				s.testnet.Validators[0].Address.String(),
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
			expectErr:    true,
			respType:     &sdk.TxResponse{},
			expectedCode: 1,
		},
		{
			name: "create tokenization - missing args",
			cmd:  assetcli.GetCmdCreateTokenization(),
			args: []string{
				// Missing required arguments
			},
			expectErr:    true,
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "create tokenization - all args (expected fail, needs setup)",
			cmd:  assetcli.GetCmdCreateTokenization(),
			args: []string{
				"1000tokenization",
				"class-id-123",
				"asset-id-123",
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
				// fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 381000000)).String()),
				fmt.Sprintf("--%s=%s", flags.FlagGasPrices, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 1)).String()),
				"--yes",
				"--keyring-backend=test",
			},
			expectErr:    false,
			respType:     &sdk.TxResponse{},
			expectedCode: 1,
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
