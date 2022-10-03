package cli_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
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
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzcli "github.com/cosmos/cosmos-sdk/x/authz/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/x/expiration/client/cli"
	expirationtypes "github.com/provenance-io/provenance/x/expiration/types"
	metadatacli "github.com/provenance-io/provenance/x/metadata/client/cli"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"

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
	keyringAccounts []keyring.Record

	asJson string
	asText string

	acctErr error

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

	moduleAssetID1 string
	moduleAssetID2 string
	moduleAssetID3 string
	moduleAssetID4 string
	moduleAssetID5 string
	moduleAssetID6 string

	sameOwner string
	diffOwner string

	time    time.Time
	deposit sdk.Coin
	// signers []string

	scopeID metadatatypes.MetadataAddress

	noExpScopeId metadatatypes.MetadataAddress

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
	pioconfig.SetProvenanceConfig("", 0)
	cfg := testutil.DefaultTestNetworkConfig()

	cfg.NumValidators = 1
	genesisState := cfg.GenesisState
	s.generateAccountsWithKeyrings(7)

	// An account
	s.accountAddr, s.acctErr = s.keyringAccounts[0].GetAddress()
	s.Require().NoError(s.acctErr, "getting keyringAccounts[0] address")
	s.accountAddrStr = s.accountAddr.String()

	// A user account
	s.user1Addr, s.acctErr = s.keyringAccounts[1].GetAddress()
	s.Require().NoError(s.acctErr, "getting keyringAccounts[1] address")
	s.user1AddrStr = s.user1Addr.String()

	// A second user account
	s.user2Addr, s.acctErr = s.keyringAccounts[2].GetAddress()
	s.Require().NoError(s.acctErr, "getting keyringAccounts[2] address")
	s.user2AddrStr = s.user2Addr.String()

	// A third user account
	s.user3Addr, s.acctErr = s.keyringAccounts[3].GetAddress()
	s.Require().NoError(s.acctErr, "getting keyringAccounts[3] address")
	s.user3AddrStr = s.user3Addr.String()

	// A third user account
	s.user4Addr, s.acctErr = s.keyringAccounts[4].GetAddress()
	s.Require().NoError(s.acctErr, "getting keyringAccounts[4] address")
	s.user4AddrStr = s.user4Addr.String()

	// A third user account
	s.user5Addr, s.acctErr = s.keyringAccounts[5].GetAddress()
	s.Require().NoError(s.acctErr, "getting keyringAccounts[5] address")
	s.user5AddrStr = s.user5Addr.String()

	// A third user account
	s.user6Addr, s.acctErr = s.keyringAccounts[6].GetAddress()
	s.Require().NoError(s.acctErr, "getting keyringAccounts[6] address")
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
		expirationtypes.DefaultDeposit).Sort()})
	genBalances = append(genBalances, banktypes.Balance{Address: s.user1AddrStr, Coins: sdk.NewCoins(
		sdk.NewCoin(cfg.BondDenom, cfg.StakingTokens),
		expirationtypes.DefaultDeposit).Add(expirationtypes.DefaultDeposit).Sort()})
	genBalances = append(genBalances, banktypes.Balance{Address: s.user2AddrStr, Coins: sdk.NewCoins(
		sdk.NewCoin(cfg.BondDenom, cfg.StakingTokens),
		expirationtypes.DefaultDeposit).Sort()})
	genBalances = append(genBalances, banktypes.Balance{Address: s.user3AddrStr, Coins: sdk.NewCoins(
		sdk.NewCoin(cfg.BondDenom, cfg.StakingTokens),
		expirationtypes.DefaultDeposit).Sort()})
	genBalances = append(genBalances, banktypes.Balance{Address: s.user4AddrStr, Coins: sdk.NewCoins(
		sdk.NewCoin(cfg.BondDenom, cfg.StakingTokens),
		expirationtypes.DefaultDeposit).Sort()})
	genBalances = append(genBalances, banktypes.Balance{Address: s.user5AddrStr, Coins: sdk.NewCoins(
		sdk.NewCoin(cfg.BondDenom, cfg.StakingTokens),
		expirationtypes.DefaultDeposit).Sort()})
	genBalances = append(genBalances, banktypes.Balance{Address: s.user6AddrStr, Coins: sdk.NewCoins(
		sdk.NewCoin(cfg.BondDenom, cfg.StakingTokens),
		expirationtypes.DefaultDeposit).Sort()})

	var bankGenState banktypes.GenesisState
	bankGenState.Params = banktypes.DefaultParams()
	bankGenState.Balances = genBalances
	bankDataBz, err := cfg.Codec.MarshalJSON(&bankGenState)
	s.Require().NoError(err)
	genesisState[banktypes.ModuleName] = bankDataBz

	s.asJson = fmt.Sprintf("--%s=json", tmcli.OutputFlag)
	s.asText = fmt.Sprintf("--%s=text", tmcli.OutputFlag)

	s.moduleAssetID1 = s.user1AddrStr
	s.moduleAssetID2 = s.user2AddrStr
	s.moduleAssetID3 = s.user3AddrStr
	s.moduleAssetID4 = s.user4AddrStr
	s.moduleAssetID5 = s.user5AddrStr
	s.moduleAssetID6 = s.user6AddrStr

	s.sameOwner = s.user1AddrStr
	s.diffOwner = s.user3AddrStr

	s.time = time.Now().AddDate(0, 0, 5)
	s.deposit = expirationtypes.DefaultDeposit

	s.scopeID = metadatatypes.ScopeMetadataAddress(uuid.New())
	s.noExpScopeId = metadatatypes.ScopeMetadataAddress(uuid.New())

	s.expiration1 = *expirationtypes.NewExpiration(s.moduleAssetID1, s.sameOwner, s.time, s.deposit, s.anyMsg(s.sameOwner))
	s.expiration2 = *expirationtypes.NewExpiration(s.moduleAssetID2, s.sameOwner, s.time, s.deposit, s.anyMsg(s.sameOwner))
	s.expiration3 = *expirationtypes.NewExpiration(s.moduleAssetID3, s.diffOwner, s.time, s.deposit, s.anyMsg(s.diffOwner))
	s.expiration4 = *expirationtypes.NewExpiration(s.moduleAssetID4, s.user4AddrStr, s.time, s.deposit, s.anyMsg(s.user4AddrStr))
	s.expiration5 = *expirationtypes.NewExpiration(s.moduleAssetID5, s.user5AddrStr, s.time, s.deposit, s.anyMsg(s.user5AddrStr))
	s.expiration6 = *expirationtypes.NewExpiration(s.moduleAssetID6, s.user6AddrStr, s.time, s.deposit, s.anyMsg(s.user6AddrStr))

	utcFormat := "2006-01-02T15:04:05.000000000Z"
	// expected expirations as JSON
	s.expiration1AsJson = fmt.Sprintf("{\"expiration\":{\"module_asset_id\":\"%s\",\"owner\":\"%s\",\"time\":\"%v\",\"deposit\":{\"denom\":\"%s\",\"amount\":\"%v\"},\"message\":{\"@type\":\"/provenance.metadata.v1.MsgDeleteScopeRequest\",\"scope_id\":\"%s\",\"signers\":[\"%s\"]}}}",
		s.moduleAssetID1,
		s.sameOwner,
		s.time.UTC().Format(utcFormat),
		s.deposit.Denom,
		s.deposit.Amount,
		s.scopeID.String(),
		s.sameOwner,
	)
	s.expiration2AsJson = fmt.Sprintf("{\"expiration\":{\"module_asset_id\":\"%s\",\"owner\":\"%s\",\"time\":\"%v\",\"deposit\":{\"denom\":\"%s\",\"amount\":\"%v\"},\"message\":{\"@type\":\"/provenance.metadata.v1.MsgDeleteScopeRequest\",\"scope_id\":\"%s\",\"signers\":[\"%s\"]}}}",
		s.moduleAssetID2,
		s.sameOwner,
		s.time.UTC().Format(utcFormat),
		s.deposit.Denom,
		s.deposit.Amount,
		s.scopeID.String(),
		s.sameOwner,
	)
	s.expiration3AsJson = fmt.Sprintf("{\"expiration\":{\"module_asset_id\":\"%s\",\"owner\":\"%s\",\"time\":\"%v\",\"deposit\":{\"denom\":\"%s\",\"amount\":\"%v\"},\"message\":{\"@type\":\"/provenance.metadata.v1.MsgDeleteScopeRequest\",\"scope_id\":\"%s\",\"signers\":[\"%s\"]}}}",
		s.moduleAssetID3,
		s.diffOwner,
		s.time.UTC().Format(utcFormat),
		s.deposit.Denom,
		s.deposit.Amount,
		s.scopeID.String(),
		s.diffOwner,
	)

	// expected expirations as text
	s.expiration1AsText = fmt.Sprintf(`expiration:
  deposit:
    amount: "%v"
    denom: %s
  message:
    '@type': /provenance.metadata.v1.MsgDeleteScopeRequest
    scope_id: %s
    signers:
    - %s
  module_asset_id: %s
  owner: %s
  time: "%v"`,
		s.deposit.Amount,
		s.deposit.Denom,
		s.scopeID.String(),
		s.sameOwner,
		s.moduleAssetID1,
		s.sameOwner,
		s.time.UTC().Format(utcFormat),
	)
	s.expiration2AsText = fmt.Sprintf(`expiration:
  deposit:
    amount: "%v"
    denom: %s
  message:
    '@type': /provenance.metadata.v1.MsgDeleteScopeRequest
    scope_id: %s
    signers:
    - %s
  module_asset_id: %s
  owner: %s
  time: "%v"`,
		s.deposit.Amount,
		s.deposit.Denom,
		s.scopeID.String(),
		s.sameOwner,
		s.moduleAssetID2,
		s.sameOwner,
		s.time.UTC().Format(utcFormat),
	)
	s.expiration3AsText = fmt.Sprintf(`expiration:
  deposit:
    amount: "%v"
    denom: %s
  message:
    '@type': /provenance.metadata.v1.MsgDeleteScopeRequest
    scope_id: %s
    signers:
    - %s
  module_asset_id: %s
  owner: %s
  time: "%v"`,
		s.time.UTC().Format(utcFormat),
		s.deposit.Amount,
		s.deposit.Denom,
		s.scopeID.String(),
		s.diffOwner,
		s.moduleAssetID3,
		s.diffOwner,
	)

	var expirationData expirationtypes.GenesisState
	s.Require().NoError(cfg.Codec.UnmarshalJSON(genesisState[expirationtypes.ModuleName], &expirationData))
	expirationData.Expirations = append(expirationData.Expirations, s.expiration1)
	expirationData.Expirations = append(expirationData.Expirations, s.expiration2)
	expirationData.Expirations = append(expirationData.Expirations, s.expiration3)
	expirationData.Expirations = append(expirationData.Expirations, s.expiration4)
	expirationData.Expirations = append(expirationData.Expirations, s.expiration5)
	expirationData.Expirations = append(expirationData.Expirations, s.expiration6)
	expirationDataBz, err := cfg.Codec.MarshalJSON(&expirationData)
	s.Require().NoError(err)
	genesisState[expirationtypes.ModuleName] = expirationDataBz

	cfg.GenesisState = genesisState

	s.cfg = cfg
	cfg.ChainID = antewrapper.SimAppChainID
	s.testnet, err = testnet.New(s.T(), s.T().TempDir(), cfg)

	_, err = s.testnet.WaitForHeight(1)
	s.Require().NoError(err)
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

func (s *IntegrationCLITestSuite) getClientCtx() client.Context {
	return s.getClientCtxWithoutKeyring().WithKeyringDir(s.keyringDir).WithKeyring(s.keyring)
}

func (s *IntegrationCLITestSuite) getClientCtxWithoutKeyring() client.Context {
	return s.testnet.Validators[0].ClientCtx
}

func (s *IntegrationCLITestSuite) anyMsg(owner string) types.Any {
	msg := &metadatatypes.MsgDeleteScopeRequest{
		ScopeId: s.scopeID,
		Signers: []string{owner},
	}
	anyMsg, err := types.NewAnyWithValue(msg)
	if err != nil {
		panic(err)
	}
	return *anyMsg
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
			args:             []string{s.moduleAssetID1, s.asJson},
			expectedError:    "",
			expectedInOutput: []string{s.expiration1AsJson},
		},
		{
			name:             "get expiration by module asset id - id as text",
			args:             []string{s.moduleAssetID1, s.asText},
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
			expectedError:    "decoding bech32 failed: invalid checksum (expected kzwk8c got xlkwel)",
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
			args:             []string{s.moduleAssetID1, s.moduleAssetID2},
			expectedError:    "accepts 1 arg(s), received 2",
			expectedInOutput: []string{},
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetAllExpirationsCmd() {
	cmd := func() *cobra.Command { return cli.GetAllExpirationsCmd() }

	pageSizeArg := fmt.Sprintf("--%s=%d", flags.FlagLimit, 7)

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
			args:             []string{s.moduleAssetID1, pageSizeArg},
			expectedError:    fmt.Sprintf("unknown command \"%s\" for \"all\"", s.moduleAssetID1),
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
			args:             []string{s.sameOwner, s.moduleAssetID1, pageSizeArg},
			expectedError:    "accepts 1 arg(s), received 2",
			expectedInOutput: []string{},
		},
	}

	runQueryCmdTestCases(s, cmd, testCases)
}

func (s *IntegrationCLITestSuite) TestGetAllExpiredExpirationsCmd() {
	cmd := func() *cobra.Command { return cli.GetAllExpiredExpirationsCmd() }

	pageSizeArg := fmt.Sprintf("--%s=%d", flags.FlagLimit, 7)

	testCases := []queryCmdTestCase{
		{
			name:             "get expired expirations - as json",
			args:             []string{s.asJson, pageSizeArg},
			expectedError:    "",
			expectedInOutput: []string{},
		},
		{
			name:             "get expired expirations - as text",
			args:             []string{s.asText, pageSizeArg},
			expectedError:    "",
			expectedInOutput: []string{},
		},
		{
			name:             "get expired expirations - no args",
			args:             []string{"dd", pageSizeArg},
			expectedError:    fmt.Sprintf("unknown command \"%s\" for \"expired\"", "dd"),
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

				umErr := clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType)
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

func (s *IntegrationCLITestSuite) TestExpirationTxCommands() {
	scopeSpecID := metadatatypes.ScopeSpecMetadataAddress(uuid.New()).String()
	contractSpecID := metadatatypes.ContractSpecMetadataAddress(uuid.New()).String()

	testCases := []txCmdTestCase{
		{
			name: "should successfully extend expiration",
			cmd:  cli.ExtendExpirationCmd(),
			args: []string{
				s.expiration4.ModuleAssetId,
				"100d",
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
				"100d",
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
				s.expiration6.ModuleAssetId,
				"100d",
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
				"100d",
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
			"should successfully add metadata contract specification",
			metadatacli.WriteContractSpecificationCmd(),
			[]string{
				contractSpecID,
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
			"should successfully add metadata scope specification",
			metadatacli.WriteScopeSpecificationCmd(),
			[]string{
				scopeSpecID,
				s.accountAddrStr,
				"owner",
				contractSpecID,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, "", &sdk.TxResponse{}, 0,
		},
		{
			"should successfully add metadata scope with expiration",
			metadatacli.WriteScopeCmd(),
			[]string{
				s.scopeID.String(),
				scopeSpecID,
				s.accountAddrStr,
				s.accountAddrStr,
				s.accountAddrStr,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=%s", metadatacli.FlagExpires, "1y"),
			},
			false, "", &sdk.TxResponse{}, 0,
		},
		{
			"should successfully add metadata scope without expiration",
			metadatacli.WriteScopeCmd(),
			[]string{
				s.noExpScopeId.String(),
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
			name: "should successfully invoke metadata scope expiration",
			cmd:  cli.InvokeExpirationCmd(),
			args: []string{
				s.scopeID.String(),
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
			name: "should fail to invoke expiration on scope with no expiration set",
			cmd:  cli.InvokeExpirationCmd(),
			args: []string{
				s.noExpScopeId.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: 4,
		},
		{
			name: "should fail to extend expiration, not found",
			cmd:  cli.ExtendExpirationCmd(),
			args: []string{
				metadatatypes.ScopeMetadataAddress(uuid.New()).String(),
				"100d",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddrStr),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: expirationtypes.ErrNotFound.ABCICode(),
		},
		{
			name: "should fail to invoke expiration, not found",
			cmd:  cli.InvokeExpirationCmd(),
			args: []string{
				metadatatypes.ScopeMetadataAddress(uuid.New()).String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.expiration4.Owner),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr:    false,
			expectErrMsg: "",
			respType:     &sdk.TxResponse{},
			expectedCode: expirationtypes.ErrNotFound.ABCICode(),
		},
	}

	runTxCmdTestCases(s, testCases)
}
