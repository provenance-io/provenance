package cli_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	cmtcli "github.com/cometbft/cometbft/libs/cli"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/testutil/queries"
	oraclecli "github.com/provenance-io/provenance/x/oracle/client/cli"
	"github.com/provenance-io/provenance/x/oracle/types"
	oracletypes "github.com/provenance-io/provenance/x/oracle/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg        network.Config
	network    *network.Network
	keyring    keyring.Keyring
	keyringDir string

	accountAddr      sdk.AccAddress
	accountKey       *secp256k1.PrivKey
	accountAddresses []sdk.AccAddress

	port   string
	oracle string
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	pioconfig.SetProvenanceConfig("", 0)
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

	s.port = oracletypes.PortID
	s.oracle = "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"

	oracleData := oracletypes.NewGenesisState(
		s.port,
		s.oracle,
	)

	oracleDataBz, err := s.cfg.Codec.MarshalJSON(oracleData)
	s.Require().NoError(err, "should be able to marshal trigger genesis state when setting up suite")
	genesisState[oracletypes.ModuleName] = oracleDataBz

	s.cfg.GenesisState = genesisState

	s.cfg.ChainID = antewrapper.SimAppChainID

	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err, "network.New")

	_, err = testutil.WaitForHeight(s.network, 6)
	s.Require().NoError(err, "WaitForHeight")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	testutil.Cleanup(s.network, s.T())
}

func (s *IntegrationTestSuite) GenerateAccountsWithKeyrings(number int) {
	path := hd.CreateHDPath(118, 0, 0).String()
	s.keyringDir = s.T().TempDir()
	kr, err := keyring.New(s.T().Name(), "test", s.keyringDir, nil, s.cfg.Codec)
	s.Require().NoError(err, "Keyring.New")
	s.keyring = kr
	for i := 0; i < number; i++ {
		keyId := fmt.Sprintf("test_key%v", i)
		info, _, err := kr.NewMnemonic(keyId, keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		s.Require().NoError(err, "Keyring.NewMneomonic")
		addr, err := info.GetAddress()
		if err != nil {
			panic(err)
		}
		s.accountAddresses = append(s.accountAddresses, addr)
	}
}

func (s *IntegrationTestSuite) TestQueryOracleAddress() {
	testCases := []struct {
		name            string
		expectErrMsg    string
		expectedCode    uint32
		expectedAddress string
	}{
		{
			name:            "success - query for oracle address",
			expectedAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			expectedCode:    0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := s.network.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, oraclecli.GetQueryOracleAddressCmd(), []string{fmt.Sprintf("--%s=json", cmtcli.OutputFlag)})
			if len(tc.expectErrMsg) > 0 {
				s.EqualError(err, tc.expectErrMsg, "should have correct error message for invalid QueryOracleAddress")
			} else {
				var response types.QueryOracleAddressResponse
				s.NoError(err, "should have no error message for valid QueryOracleAddress")
				err = s.cfg.Codec.UnmarshalJSON(out.Bytes(), &response)
				s.NoError(err, "should have no error message when unmarshalling response to QueryOracleAddress")
				s.Equal(tc.expectedAddress, response.Address, "should have the correct oracle address")
			}
		})
	}
}

func (s *IntegrationTestSuite) TestOracleUpdate() {
	testCases := []struct {
		name         string
		address      string
		expectErrMsg string
		expectedCode uint32
		signer       string
	}{
		{
			name:         "success - address updated",
			address:      "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			expectedCode: 0,
			signer:       s.accountAddresses[0].String(),
		},
		{
			name:         "failure - unable to pass validate basic with bad address",
			address:      "badaddress",
			expectErrMsg: "invalid address for oracle: decoding bech32 failed: invalid separator index -1: invalid proposal message",
			expectedCode: 12,
			signer:       s.accountAddresses[0].String(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			clientCtx := s.network.Validators[0].ClientCtx.WithKeyringDir(s.keyringDir).WithKeyring(s.keyring)

			cmd := oraclecli.GetCmdOracleUpdate()
			args := []string{
				tc.address,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, tc.signer),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
				"--title", "Update the oracle", "--summary", "Update it real good",
				fmt.Sprintf("--%s=json", cmtcli.OutputFlag),
			}

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			outBz := out.Bytes()
			s.T().Logf("ExecTestCLICmd %q %q\nOutput:\n%s", cmd.Name(), args, string(outBz))
			s.Assert().NoError(err, "should have no error for valid OracleUpdate request")

			txResp := queries.GetTxFromResponse(s.T(), s.network, outBz)
			s.Assert().Equal(int(tc.expectedCode), int(txResp.Code), "should have correct response code for valid OracleUpdate request")
			s.Assert().Contains(txResp.RawLog, tc.expectErrMsg, "should have correct error for invalid OracleUpdate request")
		})
	}
}

func (s *IntegrationTestSuite) TestSendQuery() {
	testCases := []struct {
		name         string
		query        string
		channel      string
		expectErrMsg string
		expectedCode uint32
		signer       string
	}{
		{
			name:         "success - a valid message was attempted to be sent on ibc",
			query:        "{}",
			channel:      "channel-1",
			expectedCode: 9,
			signer:       s.accountAddresses[0].String(),
		},
		{
			name:         "failure - invalid query data",
			query:        "abc",
			expectErrMsg: "query data must be json",
			channel:      "channel-1",
			signer:       s.accountAddresses[0].String(),
		},
		{
			name:         "failure - invalid channel format",
			query:        "{}",
			expectErrMsg: "invalid channel id",
			channel:      "a",
			signer:       s.accountAddresses[0].String(),
		},
		{
			name:         "failure - invalid signer",
			query:        "{}",
			expectErrMsg: "failed to convert address field to address: abc.info: key not found",
			channel:      "channel-1",
			signer:       "abc",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			clientCtx := s.network.Validators[0].ClientCtx.WithKeyringDir(s.keyringDir).WithKeyring(s.keyring)

			cmd := oraclecli.GetCmdSendQuery()
			args := []string{
				tc.channel,
				tc.query,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, tc.signer),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
				fmt.Sprintf("--%s=json", cmtcli.OutputFlag),
			}

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			outBz := out.Bytes()
			s.T().Logf("ExecTestCLICmd %q %q\nOutput:\n%s", cmd.Name(), args, string(outBz))

			if len(tc.expectErrMsg) > 0 {
				s.Assert().ErrorContains(err, tc.expectErrMsg, "should have correct error for invalid SendQuery request")
			} else {
				s.Assert().NoError(err, "should have no error for valid SendQuery request")
				txResp := queries.GetTxFromResponse(s.T(), s.network, outBz)
				s.Assert().Equal(int(tc.expectedCode), int(txResp.Code), "should have correct response code for valid SendQuery request")
			}
		})
	}
}
