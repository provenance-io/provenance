package cli_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	testcli "github.com/provenance-io/provenance/testutil/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	assetcli "github.com/provenance-io/provenance/x/asset/client/cli"
	"github.com/provenance-io/provenance/x/asset/types"
)

// CmdTestSuite is a test suite for the asset module CLI commands.
type CmdTestSuite struct {
	suite.Suite

	cfg     testnet.Config
	testnet *testnet.Network

	keyring        keyring.Keyring
	keyringEntries []testutil.TestKeyringEntry
	accountAddrs   []sdk.AccAddress

	addr0 sdk.AccAddress
	addr1 sdk.AccAddress
	addr2 sdk.AccAddress
}

// TestCmdTestSuite runs the CmdTestSuite.
func TestCmdTestSuite(t *testing.T) {
	suite.Run(t, new(CmdTestSuite))
}

// SetupSuite sets up the test suite by creating a test network and initializing test data.
func (s *CmdTestSuite) SetupSuite() {
	s.T().Log("setting up asset module CLI integration test suite")
	pioconfig.SetProvConfig("stake")
	govv1.DefaultMinDepositRatio = sdkmath.LegacyZeroDec()

	s.cfg = testutil.DefaultTestNetworkConfig()
	s.cfg.NumValidators = 1
	s.cfg.ChainID = antewrapper.SimAppChainID

	// Generate test accounts with keyring
	s.generateAccountsWithKeyring(3)
	s.addr0 = s.accountAddrs[0]
	s.addr1 = s.accountAddrs[1]
	s.addr2 = s.accountAddrs[2]

	// Add accounts to auth genesis state
	testutil.MutateGenesisState(s.T(), &s.cfg, authtypes.ModuleName, &authtypes.GenesisState{},
		func(authGen *authtypes.GenesisState) *authtypes.GenesisState {
			genAccs := make(authtypes.GenesisAccounts, 0, len(s.accountAddrs))
			for _, addr := range s.accountAddrs {
				genAccs = append(genAccs, authtypes.NewBaseAccount(addr, nil, 0, 0))
			}
			newAccounts, err := authtypes.PackAccounts(genAccs)
			s.Require().NoError(err, "PackAccounts")
			authGen.Accounts = append(authGen.Accounts, newAccounts...)
			return authGen
		})

	// Add balances to bank genesis state
	testutil.MutateGenesisState(s.T(), &s.cfg, banktypes.ModuleName, &banktypes.GenesisState{},
		func(bankGen *banktypes.GenesisState) *banktypes.GenesisState {
			balance := sdk.NewCoins(
				sdk.NewInt64Coin(s.cfg.BondDenom, 1_000_000_000),
			)
			for _, addr := range s.accountAddrs {
				bankGen.Balances = append(bankGen.Balances, banktypes.Balance{
					Address: addr.String(),
					Coins:   balance,
				})
			}
			return bankGen
		})

	// Note: The asset module doesn't have its own genesis state - it uses the NFT module.
	// Assets and classes will be created via transactions in the tests.

	// Create and start the test network
	var err error
	s.testnet, err = testnet.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err, "testnet.New")

	s.testnet.Validators[0].ClientCtx = s.testnet.Validators[0].ClientCtx.WithKeyring(s.keyring)

	_, err = testutil.WaitForHeight(s.testnet, 1)
	s.Require().NoError(err, "WaitForHeight")

	// Create initial test data by sending transactions
	s.createTestAssetClass("test-class-1", "Test Class 1", "TC1")
	s.createTestAssetClass("test-class-2", "Test Class 2", "TC2")
	s.createTestAsset("test-class-1", "test-asset-1", s.addr0.String())
	s.createTestAsset("test-class-1", "test-asset-2", s.addr1.String())
	err = testutil.WaitForNextBlock(s.testnet)
	s.Require().NoError(err, "WaitForNextBlock at end of suite setup")
}

// TearDownSuite tears down the test suite by cleaning up the test network.
func (s *CmdTestSuite) TearDownSuite() {
	testutil.Cleanup(s.testnet, s.T())
}

// generateAccountsWithKeyring creates a keyring and adds test accounts to it.
func (s *CmdTestSuite) generateAccountsWithKeyring(number int) {
	s.keyringEntries, s.keyring = testutil.GenerateTestKeyring(s.T(), number, s.cfg.Codec)
	s.accountAddrs = testutil.GetKeyringEntryAddresses(s.keyringEntries)
}

// getClientCtx returns a client context with the test keyring.
func (s *CmdTestSuite) getClientCtx() client.Context {
	return s.testnet.Validators[0].ClientCtx.WithKeyring(s.keyring)
}

// createTestAssetClass creates a test asset class via transaction.
func (s *CmdTestSuite) createTestAssetClass(id, name, symbol string) {
	args := []string{
		id,
		name,
		`{"type":"test"}`,
		"--symbol", symbol,
		"--description", fmt.Sprintf("Test asset class %s", id),
		"--uri", fmt.Sprintf("https://example.com/%s", id),
		"--uri-hash", fmt.Sprintf("%s-hash", id),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addr0.String()),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 1)).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	}

	testcli.NewTxExecutor(assetcli.GetCmdCreateAssetClass(), args).Execute(s.T(), s.testnet)
}

// createTestAsset creates a test asset via transaction.
func (s *CmdTestSuite) createTestAsset(classID, assetID, owner string) {
	args := []string{
		classID,
		assetID,
		`{"value":100}`,
		"--owner", owner,
		"--uri", fmt.Sprintf("https://example.com/%s/%s", classID, assetID),
		"--uri-hash", fmt.Sprintf("%s-hash", assetID),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addr0.String()),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 1)).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	}

	testcli.NewTxExecutor(assetcli.GetCmdCreateAsset(), args).Execute(s.T(), s.testnet)
}

// TestQueryAssetCmd tests the query asset command.
func (s *CmdTestSuite) TestQueryAssetCmd() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		expErrMsg string
		validate  func(*types.QueryAssetResponse)
	}{
		{
			name:      "valid asset query",
			args:      []string{"test-class-1", "test-asset-1", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			expectErr: false,
			validate: func(resp *types.QueryAssetResponse) {
				s.Require().NotNil(resp.Asset)
				s.Assert().Equal("test-class-1", resp.Asset.ClassId)
				s.Assert().Equal("test-asset-1", resp.Asset.Id)
			},
		},
		{
			name:      "non-existent asset",
			args:      []string{"test-class-1", "does-not-exist", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			expectErr: true,
			expErrMsg: "not found",
		},
		{
			name:      "non-existent class",
			args:      []string{"does-not-exist", "test-asset-1", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			expectErr: true,
			expErrMsg: "not found",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			clientCtx := s.getClientCtx()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, assetcli.GetCmdAsset(), tc.args)

			if tc.expectErr {
				s.Require().Error(err)
				if tc.expErrMsg != "" {
					s.Assert().Contains(err.Error(), tc.expErrMsg)
				}
			} else {
				s.Require().NoError(err)
				if tc.validate != nil {
					var resp types.QueryAssetResponse
					err = clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp)
					s.Require().NoError(err)
					tc.validate(&resp)
				}
			}
		})
	}
}

// TestQueryAssetsCmd tests the query assets command.
func (s *CmdTestSuite) TestQueryAssetsCmd() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		validate  func(*types.QueryAssetsResponse)
	}{
		{
			name:      "query all assets requires class or owner",
			args:      []string{fmt.Sprintf("--%s=json", flags.FlagOutput)},
			expectErr: true,
		},
		{
			name:      "query assets by class",
			args:      []string{"test-class-1", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			expectErr: false,
			validate: func(resp *types.QueryAssetsResponse) {
				s.Assert().GreaterOrEqual(len(resp.Assets), 2)
				for _, asset := range resp.Assets {
					s.Assert().Equal("test-class-1", asset.ClassId)
				}
			},
		},
		{
			name:      "query assets by owner",
			args:      []string{s.addr0.String(), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			expectErr: false,
			validate: func(resp *types.QueryAssetsResponse) {
				s.Assert().GreaterOrEqual(len(resp.Assets), 1)
			},
		},
		{
			name:      "query assets by class and owner",
			args:      []string{"test-class-1", s.addr0.String(), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			expectErr: false,
			validate: func(resp *types.QueryAssetsResponse) {
				for _, asset := range resp.Assets {
					s.Assert().Equal("test-class-1", asset.ClassId)
				}
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			clientCtx := s.getClientCtx()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, assetcli.GetCmdAssets(), tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				if tc.validate != nil {
					var resp types.QueryAssetsResponse
					err = clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp)
					s.Require().NoError(err)
					tc.validate(&resp)
				}
			}
		})
	}
}

// TestQueryAssetClassCmd tests the query asset class command.
func (s *CmdTestSuite) TestQueryAssetClassCmd() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		expErrMsg string
		validate  func(*types.QueryAssetClassResponse)
	}{
		{
			name:      "valid class query",
			args:      []string{"test-class-1", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			expectErr: false,
			validate: func(resp *types.QueryAssetClassResponse) {
				s.Require().NotNil(resp.AssetClass)
				s.Assert().Equal("test-class-1", resp.AssetClass.Id)
				s.Assert().Equal("Test Class 1", resp.AssetClass.Name)
				s.Assert().Equal("TC1", resp.AssetClass.Symbol)
			},
		},
		{
			name:      "non-existent class",
			args:      []string{"does-not-exist", fmt.Sprintf("--%s=json", flags.FlagOutput)},
			expectErr: true,
			expErrMsg: "not found",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			clientCtx := s.getClientCtx()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, assetcli.GetCmdAssetClass(), tc.args)

			if tc.expectErr {
				s.Require().Error(err)
				if tc.expErrMsg != "" {
					s.Assert().Contains(err.Error(), tc.expErrMsg)
				}
			} else {
				s.Require().NoError(err)
				if tc.validate != nil {
					var resp types.QueryAssetClassResponse
					err = clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp)
					s.Require().NoError(err)
					tc.validate(&resp)
				}
			}
		})
	}
}

// TestQueryAssetClassesCmd tests the query asset classes command.
func (s *CmdTestSuite) TestQueryAssetClassesCmd() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		validate  func(*types.QueryAssetClassesResponse)
	}{
		{
			name:      "query all classes",
			args:      []string{fmt.Sprintf("--%s=json", flags.FlagOutput)},
			expectErr: false,
			validate: func(resp *types.QueryAssetClassesResponse) {
				s.Assert().GreaterOrEqual(len(resp.AssetClasses), 2)
			},
		},
		{
			name:      "query with pagination",
			args:      []string{fmt.Sprintf("--%s=1", flags.FlagLimit), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			expectErr: false,
			validate: func(resp *types.QueryAssetClassesResponse) {
				s.Assert().LessOrEqual(len(resp.AssetClasses), 1)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			clientCtx := s.getClientCtx()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, assetcli.GetCmdAssetClasses(), tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				if tc.validate != nil {
					var resp types.QueryAssetClassesResponse
					err = clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp)
					s.Require().NoError(err)
					tc.validate(&resp)
				}
			}
		})
	}
}

// TestTxCreateAssetClassCmd tests the create asset class transaction command.
func (s *CmdTestSuite) TestTxCreateAssetClassCmd() {
	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
	}{
		{
			name: "valid create asset class",
			args: []string{
				"new-class-tx-1",
				"New Asset Class",
				`{"type":"new"}`,
				"--symbol", "NAC",
				"--description", "A new asset class",
				"--uri", "https://example.com/new-class",
				"--uri-hash", "new-class-hash",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addr0.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 1)).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			expectErr:    false,
			expectedCode: 0,
		},
		{
			name: "missing required args",
			args: []string{
				// Missing arguments
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addr0.String()),
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			clientCtx := s.getClientCtx()
			_, err := clitestutil.ExecTestCLICmd(clientCtx, assetcli.GetCmdCreateAssetClass(), tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				// For BroadcastSync mode, we only verify the command executes without error.
			}
		})
	}
}

// TestTxCreateAssetCmd tests the create asset transaction command.
func (s *CmdTestSuite) TestTxCreateAssetCmd() {
	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
	}{
		{
			name: "valid create asset",
			args: []string{
				"test-class-1",
				"new-asset-tx-1",
				`{"value":300}`,
				"--uri", "https://example.com/new-asset",
				"--uri-hash", "new-asset-hash",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addr0.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 1)).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			expectErr:    false,
			expectedCode: 0,
		},
		{
			name: "missing required args",
			args: []string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addr0.String()),
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			clientCtx := s.getClientCtx()
			_, err := clitestutil.ExecTestCLICmd(clientCtx, assetcli.GetCmdCreateAsset(), tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				// For BroadcastSync mode, we only verify the command executes without error.
			}
		})
	}
}

// TestTxBurnAssetCmd tests the burn asset transaction command.
func (s *CmdTestSuite) TestTxBurnAssetCmd() {
	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
	}{
		{
			name: "valid burn asset",
			args: []string{
				"test-class-1",
				"test-asset-1",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addr0.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 1)).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			expectErr:    false,
			expectedCode: 0,
		},
		{
			name: "missing required args",
			args: []string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addr0.String()),
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			clientCtx := s.getClientCtx()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, assetcli.GetCmdBurnAsset(), tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var txResp sdk.TxResponse
				err = clientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp)
				s.Require().NoError(err)
				s.Assert().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

// TestTxCreatePoolCmd tests the create pool transaction command.
func (s *CmdTestSuite) TestTxCreatePoolCmd() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name: "valid create pool",
			args: []string{
				"100pooltoken",
				"test-class-1,test-asset-1",
				"test-class-1,test-asset-2",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addr0.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 1)).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			expectErr: false,
		},
		{
			name: "invalid pool format",
			args: []string{
				"invalid",
				"test-class-1,test-asset-1",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addr0.String()),
			},
			expectErr: true,
		},
		{
			name: "invalid nft format",
			args: []string{
				"100pooltoken",
				"invalid-format",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addr0.String()),
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			clientCtx := s.getClientCtx()
			_, err := clitestutil.ExecTestCLICmd(clientCtx, assetcli.GetCmdCreatePool(), tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// TestTxCreateTokenizationCmd tests the create tokenization transaction command.
func (s *CmdTestSuite) TestTxCreateTokenizationCmd() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name: "valid create tokenization",
			args: []string{
				"1000token",
				"test-class-1",
				"test-asset-1",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addr0.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 1)).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			expectErr: false,
		},
		{
			name: "invalid token format",
			args: []string{
				"invalid",
				"test-class-1",
				"test-asset-1",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addr0.String()),
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			clientCtx := s.getClientCtx()
			_, err := clitestutil.ExecTestCLICmd(clientCtx, assetcli.GetCmdCreateTokenization(), tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// TestTxCreateSecuritizationCmd tests the create securitization transaction command.
func (s *CmdTestSuite) TestTxCreateSecuritizationCmd() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name: "valid create securitization",
			args: []string{
				"sec-1",
				"pool1,pool2",
				"100tranche1,200tranche2",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addr0.String()),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 1)).String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			},
			expectErr: false,
		},
		{
			name: "invalid tranche format",
			args: []string{
				"sec-1",
				"pool1,pool2",
				"invalid",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.addr0.String()),
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			clientCtx := s.getClientCtx()
			_, err := clitestutil.ExecTestCLICmd(clientCtx, assetcli.GetCmdCreateSecuritization(), tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// setupTestCase contains the stuff that runSetupTestCase should check.
type setupTestCase struct {
	// name is the name of the setup func being tested.
	name string
	// setup is the function being tested (usually wraps GetCmdXxx).
	setup func(cmd *cobra.Command)
	// expFlags is the list of flags expected to be added to the command after setup.
	expFlags []string
	// expAnnotations is the annotations expected for each of the expFlags.
	// The map is "flag name" -> "annotation type" -> values
	expAnnotations map[string]map[string][]string
	// expInUse is a set of strings that are expected to be in the command's Use string.
	expInUse []string
	// expExamples is a set of examples to ensure are on the command.
	expExamples []string
	// skipArgsCheck true causes the runner to skip the check ensuring that the command's Args func has been set.
	skipArgsCheck bool
	// skipAddingFromFlag true causes the runner to not add the from flag to the dummy command.
	skipAddingFromFlag bool
	// skipFlagInUseCheck true causes the runner to skip checking that each entry in expFlags also appears in the cmd.Use.
	skipFlagInUseCheck bool
}

// runSetupTestCase runs the provided setup func and checks that everything is set up as expected.
func runSetupTestCase(t *testing.T, tc setupTestCase) {
	if tc.expAnnotations == nil {
		tc.expAnnotations = make(map[string]map[string][]string)
	}
	cmd := &cobra.Command{
		Use: "dummy",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("the dummy command should not have been executed")
		},
	}
	if !tc.skipAddingFromFlag {
		cmd.Flags().String(flags.FlagFrom, "", "The from flag")
	}

	testFunc := func() {
		tc.setup(cmd)
	}
	require.NotPanics(t, testFunc, tc.name)

	pageFlags := []string{
		flags.FlagPage, flags.FlagPageKey, flags.FlagOffset,
		flags.FlagLimit, flags.FlagCountTotal, flags.FlagReverse,
	}
	for i, flagName := range tc.expFlags {
		t.Run(fmt.Sprintf("flag[%d]: --%s", i, flagName), func(t *testing.T) {
			flag := cmd.Flags().Lookup(flagName)
			if assert.NotNil(t, flag, "--%s", flagName) {
				expAnnotations, _ := tc.expAnnotations[flagName]
				actAnnotations := flag.Annotations
				assert.Equal(t, expAnnotations, actAnnotations, "--%s annotations", flagName)
				if !tc.skipFlagInUseCheck {
					expInUse := "--" + flagName
					for _, pageFlag := range pageFlags {
						if flagName == pageFlag {
							// Skip individual page flag checks
							return
						}
					}
					assert.Contains(t, cmd.Use, expInUse, "cmd.Use should have something about the %s flag", flagName)
				}
			}
		})
	}

	for i, exp := range tc.expInUse {
		t.Run(fmt.Sprintf("use[%d]: %s", i, truncate(exp, 20)), func(t *testing.T) {
			assert.Contains(t, cmd.Use, exp, "command use after %s", tc.name)
			if len(exp) > 0 && exp[0] != '[' {
				assert.NotContains(t, cmd.Use, "["+exp+"]", "command use after %s", tc.name)
			}
		})
	}

	examples := strings.Split(cmd.Example, "\n")
	for i, exp := range tc.expExamples {
		t.Run(fmt.Sprintf("examples[%d]", i), func(t *testing.T) {
			assert.Contains(t, examples, exp, "command examples after %s", tc.name)
		})
	}

	if !tc.skipArgsCheck {
		t.Run("args", func(t *testing.T) {
			assert.NotNil(t, cmd.Args, "command args after %s", tc.name)
		})
	}
}

// truncate truncates a string to the specified length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// TestQueryCmdAsset tests the setup of the asset query command.
func TestQueryCmdAsset(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name: "GetCmdAsset",
		setup: func(cmd *cobra.Command) {
			*cmd = *assetcli.GetCmdAsset()
		},
		expInUse: []string{
			"asset <class-id> <asset-id>",
		},
		skipAddingFromFlag: true,
	})
}

// TestQueryCmdAssets tests the setup of the assets query command.
func TestQueryCmdAssets(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name: "GetCmdAssets",
		setup: func(cmd *cobra.Command) {
			*cmd = *assetcli.GetCmdAssets()
		},
		expFlags: []string{
			flags.FlagLimit, flags.FlagOffset, flags.FlagPage,
			flags.FlagPageKey, flags.FlagCountTotal, flags.FlagReverse,
		},
		expInUse: []string{
			"assets",
		},
		skipAddingFromFlag: true,
		skipFlagInUseCheck: true,
	})
}

// TestQueryCmdAssetClass tests the setup of the asset class query command.
func TestQueryCmdAssetClass(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name: "GetCmdAssetClass",
		setup: func(cmd *cobra.Command) {
			*cmd = *assetcli.GetCmdAssetClass()
		},
		expInUse: []string{
			"class <id>",
		},
		skipAddingFromFlag: true,
	})
}

// TestQueryCmdAssetClasses tests the setup of the asset classes query command.
func TestQueryCmdAssetClasses(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name: "GetCmdAssetClasses",
		setup: func(cmd *cobra.Command) {
			*cmd = *assetcli.GetCmdAssetClasses()
		},
		expFlags: []string{
			flags.FlagLimit, flags.FlagOffset, flags.FlagPage,
			flags.FlagPageKey, flags.FlagCountTotal, flags.FlagReverse,
		},
		expInUse: []string{
			"classes",
		},
		skipAddingFromFlag: true,
		skipArgsCheck:      true,
		skipFlagInUseCheck: true,
	})
}

// TestTxCmdBurnAsset tests the setup of the burn asset transaction command.
func TestTxCmdBurnAsset(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name: "GetCmdBurnAsset",
		setup: func(cmd *cobra.Command) {
			*cmd = *assetcli.GetCmdBurnAsset()
		},
		expInUse: []string{
			"burn-asset <class-id> <id>",
		},
		skipAddingFromFlag: true,
	})
}

// TestTxCmdCreateAsset tests the setup of the create asset transaction command.
func TestTxCmdCreateAsset(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name: "GetCmdCreateAsset",
		setup: func(cmd *cobra.Command) {
			*cmd = *assetcli.GetCmdCreateAsset()
		},
		expFlags: []string{
			assetcli.FlagURI, assetcli.FlagURIHash, assetcli.FlagOwner,
		},
		expInUse: []string{
			"create-asset <class-id> <id> <data>",
			"[--owner <owner>]",
			"[--uri <uri>]",
			"[--uri-hash <uri-hash>]",
		},
		skipAddingFromFlag: true,
	})
}

// TestTxCmdCreateAssetClass tests the setup of the create asset class transaction command.
func TestTxCmdCreateAssetClass(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name: "GetCmdCreateAssetClass",
		setup: func(cmd *cobra.Command) {
			*cmd = *assetcli.GetCmdCreateAssetClass()
		},
		expFlags: []string{
			assetcli.FlagSymbol, assetcli.FlagDescription,
			assetcli.FlagURI, assetcli.FlagURIHash,
		},
		expInUse: []string{
			"create-class <id> <name> <data>",
			"[--symbol <symbol>]",
			"[--description <description>]",
			"[--uri <uri>]",
			"[--uri-hash <uri-hash>]",
		},
		skipAddingFromFlag: true,
	})
}

// TestTxCmdCreatePool tests the setup of the create pool transaction command.
func TestTxCmdCreatePool(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name: "GetCmdCreatePool",
		setup: func(cmd *cobra.Command) {
			*cmd = *assetcli.GetCmdCreatePool()
		},
		expInUse: []string{
			"create-pool <pool> <nft1>",
		},
		skipAddingFromFlag: true,
	})
}

// TestTxCmdCreateTokenization tests the setup of the create tokenization transaction command.
func TestTxCmdCreateTokenization(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name: "GetCmdCreateTokenization",
		setup: func(cmd *cobra.Command) {
			*cmd = *assetcli.GetCmdCreateTokenization()
		},
		expInUse: []string{
			"create-tokenization <token> <nft-class-id> <nft-id>",
		},
		skipAddingFromFlag: true,
	})
}

// TestTxCmdCreateSecuritization tests the setup of the create securitization transaction command.
func TestTxCmdCreateSecuritization(t *testing.T) {
	runSetupTestCase(t, setupTestCase{
		name: "GetCmdCreateSecuritization",
		setup: func(cmd *cobra.Command) {
			*cmd = *assetcli.GetCmdCreateSecuritization()
		},
		expInUse: []string{
			"create-securitization <id> <pools> <tranches>",
		},
		skipAddingFromFlag: true,
	})
}
