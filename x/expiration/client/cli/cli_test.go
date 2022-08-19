package cli_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzcli "github.com/cosmos/cosmos-sdk/x/authz/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/x/expiration/client/cli"
	expirationtypes "github.com/provenance-io/provenance/x/expiration/types"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type IntegrationCLITestSuite struct {
	suite.Suite

	cfg             testnet.Config
	testnet         *testnet.Network
	keyring         keyring.Keyring
	keyringDir      string
	keyringAccounts []keyring.Info

	asJson string
	asText string

	accountAddr    sdk.AccAddress
	accountAddrStr string

	user1Addr    sdk.AccAddress
	user1AddrStr string

	user2Addr    sdk.AccAddress
	user2AddrStr string

	user3Addr    sdk.AccAddress
	user3AddrStr string

	user4Addr    sdk.AccAddress
	user4AddrStr string

	user5Addr    sdk.AccAddress
	user5AddrStr string

	user6Addr    sdk.AccAddress
	user6AddrStr string

	userOtherAddr sdk.AccAddress
	userOtherStr  string

	moduleAssetId1 string
	moduleAssetId2 string
	moduleAssetId3 string
	moduleAssetId4 string
	moduleAssetId5 string
	moduleAssetId6 string

	sameOwner string
	diffOwner string

	blockHeight int64
	deposit     sdk.Coin
	signers     []string

	expiration1 expirationtypes.Expiration
	expiration2 expirationtypes.Expiration
	expiration3 expirationtypes.Expiration
	expiration4 expirationtypes.Expiration
	expiration5 expirationtypes.Expiration
	expiration6 expirationtypes.Expiration

	expiration1AsJson string
	expiration2AsJson string
	expiration3AsJson string

	expiration1AsText string
	expiration2AsText string
	expiration3AsText string
}

func TestIntegrationCLITestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationCLITestSuite))
}

func (s *IntegrationCLITestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := testutil.DefaultTestNetworkConfig()
	cfg.NumValidators = 1
	genesisState := cfg.GenesisState
	s.generateAccountsWithKeyrings(7)

	// An account
	s.accountAddr = s.keyringAccounts[0].GetAddress()
	s.accountAddrStr = s.accountAddr.String()

	// A user account
	s.user1Addr = s.keyringAccounts[1].GetAddress()
	s.user1AddrStr = s.user1Addr.String()

	// A second user account
	s.user2Addr = s.keyringAccounts[2].GetAddress()
	s.user2AddrStr = s.user2Addr.String()

	// A third user account
	s.user3Addr = s.keyringAccounts[3].GetAddress()
	s.user3AddrStr = s.user3Addr.String()

	// A third user account
	s.user4Addr = s.keyringAccounts[4].GetAddress()
	s.user4AddrStr = s.user4Addr.String()

	// A third user account
	s.user5Addr = s.keyringAccounts[5].GetAddress()
	s.user5AddrStr = s.user5Addr.String()

	// A third user account
	s.user6Addr = s.keyringAccounts[6].GetAddress()
	s.user6AddrStr = s.user6Addr.String()

	// An account that isn't known
	s.userOtherAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	s.userOtherStr = s.userOtherAddr.String()

	// Configure Genesis auth data for adding test accounts
	var genAccounts []authtypes.GenesisAccount
	var authData authtypes.GenesisState
	authData.Params = authtypes.DefaultParams()
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.accountAddr, nil, 3, 0))
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.user1Addr, nil, 4, 0))
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.user2Addr, nil, 5, 0))
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.user3Addr, nil, 6, 0))
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.user4Addr, nil, 7, 0))
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.user5Addr, nil, 8, 0))
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.user6Addr, nil, 9, 0))
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

	s.moduleAssetId1 = s.user1AddrStr
	s.moduleAssetId2 = s.user2AddrStr
	s.moduleAssetId3 = s.user3AddrStr
	s.moduleAssetId4 = s.user4AddrStr
	s.moduleAssetId5 = s.user5AddrStr
	s.moduleAssetId6 = s.user6AddrStr

	s.sameOwner = s.user1AddrStr
	s.diffOwner = s.user3AddrStr

	s.blockHeight = 1
	s.deposit = expirationtypes.DefaultDeposit

	s.expiration1 = *expirationtypes.NewExpiration(s.moduleAssetId1, s.sameOwner, s.blockHeight, s.deposit, nil)
	s.expiration2 = *expirationtypes.NewExpiration(s.moduleAssetId2, s.sameOwner, s.blockHeight, s.deposit, nil)
	s.expiration3 = *expirationtypes.NewExpiration(s.moduleAssetId3, s.diffOwner, s.blockHeight, s.deposit, nil)
	s.expiration4 = *expirationtypes.NewExpiration(s.moduleAssetId4, s.user4AddrStr, s.blockHeight, s.deposit, nil)
	s.expiration5 = *expirationtypes.NewExpiration(s.moduleAssetId5, s.user5AddrStr, s.blockHeight, s.deposit, nil)
	s.expiration6 = *expirationtypes.NewExpiration(s.moduleAssetId6, s.user6AddrStr, s.blockHeight, s.deposit, nil)

	// expected expirations as JSON
	s.expiration1AsJson = fmt.Sprintf("{\"expiration\":{\"module_asset_id\":\"%s\",\"owner\":\"%s\",\"block_height\":\"%d\",\"deposit\":{\"denom\":\"%s\",\"amount\":\"%v\"},\"messages\":%v}}",
		s.moduleAssetId1,
		s.sameOwner,
		s.blockHeight,
		s.deposit.Denom,
		s.deposit.Amount,
		[]types.Any{},
	)
	s.expiration2AsJson = fmt.Sprintf("{\"expiration\":{\"module_asset_id\":\"%s\",\"owner\":\"%s\",\"block_height\":\"%d\",\"deposit\":{\"denom\":\"%s\",\"amount\":\"%v\"},\"messages\":%v}}",
		s.moduleAssetId2,
		s.sameOwner,
		s.blockHeight,
		s.deposit.Denom,
		s.deposit.Amount,
		[]types.Any{},
	)
	s.expiration3AsJson = fmt.Sprintf("{\"expiration\":{\"module_asset_id\":\"%s\",\"owner\":\"%s\",\"block_height\":\"%d\",\"deposit\":{\"denom\":\"%s\",\"amount\":\"%v\"},\"messages\":%v}}",
		s.moduleAssetId3,
		s.diffOwner,
		s.blockHeight,
		s.deposit.Denom,
		s.deposit.Amount,
		[]types.Any{},
	)

	// expected expirations as text
	s.expiration1AsText = fmt.Sprintf(`expiration:
  block_height: "%d"
  deposit:
    amount: "%v"
    denom: %s
  messages: %v
  module_asset_id: %s
  owner: %s`,
		s.blockHeight,
		s.deposit.Amount,
		s.deposit.Denom,
		[]types.Any{},
		s.moduleAssetId1,
		s.sameOwner,
	)
	s.expiration2AsText = fmt.Sprintf(`expiration:
  block_height: "%d"
  deposit:
    amount: "%v"
    denom: %s
  messages: %v
  module_asset_id: %s
  owner: %s`,
		s.blockHeight,
		s.deposit.Amount,
		s.deposit.Denom,
		[]types.Any{},
		s.moduleAssetId2,
		s.sameOwner,
	)
	s.expiration3AsText = fmt.Sprintf(`expiration:
  block_height: "%d"
  deposit:
    amount: "%v"
    denom: %s
  messages: %v
  module_asset_id: %s
  owner: %s`,
		s.blockHeight,
		s.deposit.Amount,
		s.deposit.Denom,
		[]types.Any{},
		s.moduleAssetId3,
		s.diffOwner,
	)

	var expirationData expirationtypes.GenesisState
	s.Require().NoError(cfg.Codec.UnmarshalJSON(genesisState[expirationtypes.ModuleName], &expirationData))
	expirationData.Expirations = append(expirationData.Expirations, s.expiration1)
	expirationData.Expirations = append(expirationData.Expirations, s.expiration2)
	expirationData.Expirations = append(expirationData.Expirations, s.expiration3)
	expirationDataBz, err := cfg.Codec.MarshalJSON(&expirationData)
	s.Require().NoError(err)
	genesisState[expirationtypes.ModuleName] = expirationDataBz

	cfg.GenesisState = genesisState
	msgfeestypes.DefaultFloorGasPrice = sdk.NewCoin("atom", sdk.NewInt(0))

	s.cfg = cfg
	cfg.ChainID = antewrapper.SimAppChainID
	s.testnet = testnet.New(s.T(), cfg)

	_, err = s.testnet.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationCLITestSuite) TearDownSuite() {
	testutil.CleanUp(s.testnet, s.T())
}

func (s *IntegrationCLITestSuite) generateAccountsWithKeyrings(number int) {
	path := hd.CreateHDPath(118, 0, 0).String()
	s.keyringDir = s.T().TempDir()
	kr, err := keyring.New(s.T().Name(), "test", s.keyringDir, nil)
	s.Require().NoError(err, "keyring creation")
	s.keyring = kr
	for i := 0; i < number; i++ {
		keyId := fmt.Sprintf("test_key%v", i)
		info, _, err := kr.NewMnemonic(keyId, keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		s.Require().NoError(err, "key creation")
		s.keyringAccounts = append(s.keyringAccounts, info)
	}
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
				// E.g. "rpc error: code = InvalidArgument desc = foo: invalid request" has become
				// "rpc error: code = InvalidArgument desc = rpc error: code = InvalidArgument desc = foo: invalid request"
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

func (s *IntegrationCLITestSuite) TestGetExpirationByModuleAssetIdCmd() {
	cmd := func() *cobra.Command { return cli.GetExpirationCmd() }

	testCases := []queryCmdTestCase{
		{
			name:             "get expiration by module asset id - id as json",
			args:             []string{s.moduleAssetId1, s.asJson},
			expectedError:    "",
			expectedInOutput: []string{s.expiration1AsJson},
		},
		{
			name:             "get expiration by module asset id - id as text",
			args:             []string{s.moduleAssetId1, s.asText},
			expectedError:    "",
			expectedInOutput: []string{s.expiration1AsText},
		},
		{
			name:             "get expiration by module asset id - does not exist",
			args:             []string{s.userOtherStr},
			expectedError:    fmt.Sprintf("expiration for module asset id [%s] does not exist: expiration not found: invalid request", s.userOtherStr),
			expectedInOutput: []string{},
		},
		{
			name:             "get expiration by module asset id - bad prefix",
			args:             []string{"foo1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"},
			expectedError:    fmt.Sprintf("decoding bech32 failed: invalid checksum (expected kzwk8c got xlkwel)"),
			expectedInOutput: []string{},
		},
		{
			name:             "get expiration by module asset id - no args",
			args:             []string{},
			expectedError:    "accepts 1 arg(s), received 0",
			expectedInOutput: []string{},
		},
		{
			name:             "get expiration by module asset id - two args",
			args:             []string{s.moduleAssetId1, s.moduleAssetId2},
			expectedError:    "accepts 1 arg(s), received 2",
			expectedInOutput: []string{},
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetAllExpirationsCmd() {
	cmd := func() *cobra.Command { return cli.GetAllExpirationsCmd() }

	pageSizeArg := fmt.Sprintf("--%s=%d", flags.FlagLimit, 3)

	testCases := []queryCmdTestCase{
		{
			name:             "get all expirations - as json",
			args:             []string{s.asJson, pageSizeArg},
			expectedError:    "",
			expectedInOutput: []string{s.expiration1.ModuleAssetId, s.expiration2.ModuleAssetId, s.expiration3.ModuleAssetId},
		},
		{
			name:             "get all expirations - as text",
			args:             []string{s.asText, pageSizeArg},
			expectedError:    "",
			expectedInOutput: []string{s.expiration1.ModuleAssetId, s.expiration2.ModuleAssetId, s.expiration3.ModuleAssetId},
		},
		{
			name:             "get all expirations - one args",
			args:             []string{s.moduleAssetId1, pageSizeArg},
			expectedError:    fmt.Sprintf("unknown command \"%s\" for \"all\"", s.moduleAssetId1),
			expectedInOutput: []string{},
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetAllExpirationsByOwnerCmd() {
	cmd := func() *cobra.Command { return cli.GetAllExpirationsByOwnerCmd() }

	pageSizeArg := fmt.Sprintf("--%s=%d", flags.FlagLimit, 2)

	testCases := []queryCmdTestCase{
		{
			name:             "get all expirations by owner - as json",
			args:             []string{s.sameOwner, s.asJson, pageSizeArg},
			expectedError:    "",
			expectedInOutput: []string{s.expiration1.ModuleAssetId, s.expiration2.ModuleAssetId},
		},
		{
			name:             "get all expirations by owner - as text",
			args:             []string{s.sameOwner, s.asText, pageSizeArg},
			expectedError:    "",
			expectedInOutput: []string{s.expiration1.ModuleAssetId, s.expiration2.ModuleAssetId},
		},
		{
			name:             "get all expirations - two args",
			args:             []string{s.sameOwner, s.moduleAssetId1, pageSizeArg},
			expectedError:    "accepts 1 arg(s), received 2",
			expectedInOutput: []string{},
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
			clientCtx := s.getClientCtx()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.cmd, tc.args)

			if len(tc.expectErrMsg) > 0 {
				require.EqualError(t, err, tc.expectErrMsg, "%s expected error message", cmdName)
			} else if tc.expectErr {
				require.Error(t, err, "%s expected error", cmdName)
			} else {
				require.NoError(t, err, "%s unexpected error", cmdName)

				umErr := clientCtx.JSONCodec.UnmarshalJSON(out.Bytes(), tc.respType)
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

func getLatestHeight(s *IntegrationCLITestSuite) int64 {
	blockHeight, err := s.testnet.LatestHeight()
	if err != nil {
		s.Fail("failed to retrieve current block height for test")
	}
	return blockHeight
}

func (s *IntegrationCLITestSuite) TestExpirationTxCommands() {
	testCases := []txCmdTestCase{
		{
			name: "should successfully add expiration",
			cmd:  cli.AddExpirationCmd(),
			args: []string{
				s.expiration4.ModuleAssetId,
				s.expiration4.Owner,
				strconv.FormatInt(getLatestHeight(s)+1000, 10),
				s.expiration4.Deposit.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.expiration4.Owner),
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
			name: "should successfully add expiration with signers flag",
			cmd:  cli.AddExpirationCmd(),
			args: []string{
				s.expiration5.ModuleAssetId,
				s.expiration5.Owner,
				strconv.FormatInt(getLatestHeight(s)+1000, 10),
				s.expiration5.Deposit.String(),
				fmt.Sprintf("--%s=%s", cli.FlagSigners, s.expiration5.Owner),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.expiration5.Owner),
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
			name: "should fail to add expiration without authorization grant",
			cmd:  cli.AddExpirationCmd(),
			args: []string{
				s.expiration6.ModuleAssetId,
				s.expiration6.Owner,
				strconv.FormatInt(getLatestHeight(s)+1000, 10),
				s.expiration6.Deposit.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.user4AddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: expirationtypes.ErrInvalidSigners.ABCICode(),
		},
		{
			name: "should successfully grant add authorization from owner 6 to signer 4",
			cmd:  authzcli.NewCmdGrantAuthorization(),
			args: []string{
				s.user4AddrStr,
				"generic",
				fmt.Sprintf("--%s=%s", authzcli.FlagMsgType, sdk.MsgTypeURL(&expirationtypes.MsgAddExpirationRequest{})),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.expiration6.Owner),
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
			name: "should successfully add expiration with authorization grant",
			cmd:  cli.AddExpirationCmd(),
			args: []string{
				s.expiration6.ModuleAssetId,
				s.expiration6.Owner,
				strconv.FormatInt(getLatestHeight(s)+1000, 10),
				s.expiration6.Deposit.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.user4AddrStr),
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
			name: "should successfully extend expiration",
			cmd:  cli.ExtendExpirationCmd(),
			args: []string{
				s.expiration4.ModuleAssetId,
				s.expiration4.Owner,
				strconv.FormatInt(getLatestHeight(s)+2000, 10),
				s.expiration4.Deposit.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.expiration4.Owner),
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
			name: "should successfully extend expiration with signers flag",
			cmd:  cli.ExtendExpirationCmd(),
			args: []string{
				s.expiration5.ModuleAssetId,
				s.expiration5.Owner,
				strconv.FormatInt(getLatestHeight(s)+2000, 10),
				s.expiration5.Deposit.String(),
				fmt.Sprintf("--%s=%s", cli.FlagSigners, s.expiration5.Owner),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.expiration5.Owner),
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
			name: "should fail to extend expiration without authorization grant",
			cmd:  cli.ExtendExpirationCmd(),
			args: []string{
				s.expiration5.ModuleAssetId,
				s.expiration5.Owner,
				strconv.FormatInt(getLatestHeight(s)+2000, 10),
				s.expiration5.Deposit.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.expiration6.Owner),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: expirationtypes.ErrInvalidSigners.ABCICode(),
		},
		{
			name: "should successfully grant extend authorization from owner 6 to signer 4",
			cmd:  authzcli.NewCmdGrantAuthorization(),
			args: []string{
				s.user4AddrStr,
				"generic",
				fmt.Sprintf("--%s=%s", authzcli.FlagMsgType, sdk.MsgTypeURL(&expirationtypes.MsgExtendExpirationRequest{})),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.expiration6.Owner),
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
			name: "should successfully extend expiration with authorization grant",
			cmd:  cli.ExtendExpirationCmd(),
			args: []string{
				s.expiration6.ModuleAssetId,
				s.expiration6.Owner,
				strconv.FormatInt(getLatestHeight(s)+2000, 10),
				s.expiration6.Deposit.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.user4AddrStr),
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
			name: "should successfully delete expiration",
			cmd:  cli.DeleteExpirationCmd(),
			args: []string{
				s.expiration4.ModuleAssetId,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.expiration4.Owner),
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
			name: "should successfully delete expiration with signers flag",
			cmd:  cli.DeleteExpirationCmd(),
			args: []string{
				s.expiration5.ModuleAssetId,
				fmt.Sprintf("--%s=%s", cli.FlagSigners, s.expiration5.Owner),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.expiration5.Owner),
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
			name: "should fail to delete expiration without authorization grant",
			cmd:  cli.DeleteExpirationCmd(),
			args: []string{
				s.expiration6.ModuleAssetId,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.expiration4.Owner),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: expirationtypes.ErrInvalidSigners.ABCICode(),
		},
		{
			name: "should successfully grant delete authorization from owner 6 to signer 4",
			cmd:  authzcli.NewCmdGrantAuthorization(),
			args: []string{
				s.user4AddrStr,
				"generic",
				fmt.Sprintf("--%s=%s", authzcli.FlagMsgType, sdk.MsgTypeURL(&expirationtypes.MsgDeleteExpirationRequest{})),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.expiration6.Owner),
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
			name: "should successfully delete expiration with authorization grant",
			cmd:  cli.DeleteExpirationCmd(),
			args: []string{
				s.expiration6.ModuleAssetId,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.user4AddrStr),
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
			name: "should fail to add expiration, empty module asset id",
			cmd:  cli.AddExpirationCmd(),
			args: []string{
				"",
				s.expiration4.Owner,
				strconv.FormatInt(getLatestHeight(s)+1000, 10),
				s.expiration4.Deposit.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.expiration4.Owner),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    true,
			expectErrMsg: "empty module asset id: failed basic validation",
			respType:     &sdk.TxResponse{},
			expectedCode: 0,
		},
		{
			name: "should fail to add expiration, incorrect module asset id",
			cmd:  cli.AddExpirationCmd(),
			args: []string{
				"not-an-address",
				s.expiration4.Owner,
				strconv.FormatInt(getLatestHeight(s)+1000, 10),
				s.expiration4.Deposit.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.expiration4.Owner),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: sdkerrors.ErrInvalidAddress.ABCICode(),
		},
		{
			name: "should fail to add expiration, invalid signer(s)",
			cmd:  cli.AddExpirationCmd(),
			args: []string{
				s.expiration4.ModuleAssetId,
				s.expiration4.Owner,
				strconv.FormatInt(getLatestHeight(s)+1000, 10),
				s.expiration4.Deposit.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.expiration6.Owner),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: expirationtypes.ErrInvalidSigners.ABCICode(),
		},
		{
			name: "should fail to extend expiration, not found",
			cmd:  cli.ExtendExpirationCmd(),
			args: []string{
				s.expiration4.ModuleAssetId,
				s.expiration4.Owner,
				strconv.FormatInt(getLatestHeight(s)+2000, 10),
				s.expiration4.Deposit.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.expiration4.Owner),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: expirationtypes.ErrExpirationNotFound.ABCICode(),
		},
		{
			name: "should fail to delete expiration, not found",
			cmd:  cli.DeleteExpirationCmd(),
			args: []string{
				s.expiration4.ModuleAssetId,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.expiration4.Owner),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: expirationtypes.ErrExpirationNotFound.ABCICode(),
		},
	}

	runTxCmdTestCases(s, testCases)
}
