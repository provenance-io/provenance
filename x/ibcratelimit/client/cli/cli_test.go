package cli_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	cmtcli "github.com/cometbft/cometbft/libs/cli"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/testutil"
	testcli "github.com/provenance-io/provenance/testutil/cli"
	"github.com/provenance-io/provenance/x/ibcratelimit"
	ibcratelimitcli "github.com/provenance-io/provenance/x/ibcratelimit/client/cli"
)

type TestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network

	keyring        keyring.Keyring
	keyringEntries []testutil.TestKeyringEntry

	accountAddr      sdk.AccAddress
	accountKey       *secp256k1.PrivKey
	accountAddresses []sdk.AccAddress

	ratelimiter string
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	govv1.DefaultMinDepositRatio = sdkmath.LegacyZeroDec()
	s.accountKey = secp256k1.GenPrivKeyFromSecret([]byte("acc2"))
	addr, err := sdk.AccAddressFromHexUnsafe(s.accountKey.PubKey().Address().String())
	s.Require().NoError(err)
	s.accountAddr = addr

	s.cfg = testutil.DefaultTestNetworkConfig()
	genesisState := s.cfg.GenesisState

	s.cfg.NumValidators = 1
	s.GenerateAccountsWithKeyrings(2)

	var genBalances []banktypes.Balance
	for i := range s.accountAddresses {
		genBalances = append(genBalances, banktypes.Balance{Address: s.accountAddresses[i].String(), Coins: sdk.NewCoins(
			sdk.NewInt64Coin("nhash", 100_000_000), sdk.NewInt64Coin(s.cfg.BondDenom, 100_000_000),
		).Sort()})
	}
	var bankGenState banktypes.GenesisState
	bankGenState.Params = banktypes.DefaultParams()
	bankGenState.Balances = genBalances
	bankDataBz, err := s.cfg.Codec.MarshalJSON(&bankGenState)
	s.Require().NoError(err, "should be able to marshal bank genesis state when setting up suite")
	genesisState[banktypes.ModuleName] = bankDataBz

	var authData authtypes.GenesisState
	var genAccounts []authtypes.GenesisAccount
	authData.Params = authtypes.DefaultParams()
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.accountAddresses[0], nil, 3, 0))
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.accountAddresses[1], nil, 4, 0))
	accounts, err := authtypes.PackAccounts(genAccounts)
	s.Require().NoError(err, "should be able to pack accounts for genesis state when setting up suite")
	authData.Accounts = accounts
	authDataBz, err := s.cfg.Codec.MarshalJSON(&authData)
	s.Require().NoError(err, "should be able to marshal auth genesis state when setting up suite")
	genesisState[authtypes.ModuleName] = authDataBz

	s.ratelimiter = "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"
	ratelimitData := ibcratelimit.NewGenesisState(ibcratelimit.NewParams(s.ratelimiter))

	ratelimitDataBz, err := s.cfg.Codec.MarshalJSON(ratelimitData)
	s.Require().NoError(err, "should be able to marshal ibcratelimit genesis state when setting up suite")
	genesisState[ibcratelimit.ModuleName] = ratelimitDataBz

	s.cfg.GenesisState = genesisState

	s.cfg.ChainID = antewrapper.SimAppChainID

	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err, "network.New")

	s.network.Validators[0].ClientCtx = s.network.Validators[0].ClientCtx.WithKeyring(s.keyring)
	_, err = testutil.WaitForHeight(s.network, 6)
	s.Require().NoError(err, "WaitForHeight")
}

func (s *TestSuite) TearDownSuite() {
	testutil.Cleanup(s.network, s.T())
}

func (s *TestSuite) GenerateAccountsWithKeyrings(number int) {
	s.keyringEntries, s.keyring = testutil.GenerateTestKeyring(s.T(), number, s.cfg.Codec)
	s.accountAddresses = testutil.GetKeyringEntryAddresses(s.keyringEntries)
}

func (s *TestSuite) TestGetParams() {
	testCases := []struct {
		name            string
		expectErrMsg    string
		expectedCode    uint32
		expectedAddress string
	}{
		{
			name:            "success - query for params",
			expectedAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			expectedCode:    0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			clientCtx := s.network.Validators[0].ClientCtx
			cmd := ibcratelimitcli.GetParamsCmd()
			args := []string{fmt.Sprintf("--%s=json", cmtcli.OutputFlag)}

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			outBz := out.Bytes()
			s.T().Logf("ExecTestCLICmd %q %q\nOutput:\n%s", cmd.Name(), args, string(outBz))

			if len(tc.expectErrMsg) > 0 {
				s.EqualError(err, tc.expectErrMsg, "should have correct error message for invalid Params request")
			} else {
				var response ibcratelimit.Params
				s.NoError(err, "should have no error message for valid Params request")
				err = s.cfg.Codec.UnmarshalJSON(outBz, &response)
				s.NoError(err, "should have no error message when unmarshalling response to Params request")
				s.Equal(tc.expectedAddress, response.ContractAddress, "should have the correct ratelimit address")
			}
		})
	}
}

func (s *TestSuite) TestParamsUpdate() {
	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
		expectedCode uint32
		signer       string
	}{
		{
			name:         "success - address updated",
			args:         []string{"cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"},
			expectedCode: 0,
			signer:       s.accountAddresses[0].String(),
		},
		{
			name:         "failure - invalid number of args",
			args:         []string{"cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma", "invalid"},
			expectErrMsg: "accepts 1 arg(s), received 2",
			signer:       s.accountAddresses[0].String(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := ibcratelimitcli.GetCmdParamsUpdate()
			tc.args = append(tc.args,
				"--title", "Update ibc-rate-limit params", "--summary", "See title.",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, tc.signer),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
				fmt.Sprintf("--%s=json", cmtcli.OutputFlag),
			)

			testcli.NewTxExecutor(cmd, tc.args).
				WithExpErrMsg(tc.expectErrMsg).
				WithExpCode(tc.expectedCode).
				Execute(s.T(), s.network)
		})
	}
}
