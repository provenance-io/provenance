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
	testcli "github.com/provenance-io/provenance/testutil/cli"
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
	var addrErr error
	s.accountAddr, addrErr = sdk.AccAddressFromHexUnsafe(s.accountKey.PubKey().Address().String())
	s.Require().NoError(addrErr)

	s.cfg = testutil.DefaultTestNetworkConfig()
	s.cfg.NumValidators = 1
	s.cfg.ChainID = antewrapper.SimAppChainID
	s.GenerateAccountsWithKeyrings(2)

	testutil.MutateGenesisState(s.T(), &s.cfg, banktypes.ModuleName, &banktypes.GenesisState{}, func(bankGenState *banktypes.GenesisState) *banktypes.GenesisState {
		for i := range s.accountAddresses {
			bankGenState.Balances = append(bankGenState.Balances, banktypes.Balance{Address: s.accountAddresses[i].String(), Coins: sdk.NewCoins(
				sdk.NewInt64Coin("nhash", 100_000_000), sdk.NewInt64Coin(s.cfg.BondDenom, 100_000_000),
			).Sort()})
		}
		return bankGenState
	})

	testutil.MutateGenesisState(s.T(), &s.cfg, authtypes.ModuleName, &authtypes.GenesisState{}, func(authData *authtypes.GenesisState) *authtypes.GenesisState {
		var genAccounts []authtypes.GenesisAccount
		genAccounts = append(genAccounts,
			authtypes.NewBaseAccount(s.accountAddresses[0], nil, 3, 0),
			authtypes.NewBaseAccount(s.accountAddresses[1], nil, 4, 0),
		)
		accounts, err := authtypes.PackAccounts(genAccounts)
		s.Require().NoError(err, "should be able to pack accounts for genesis state when setting up suite")
		authData.Accounts = accounts
		return authData
	})

	s.port = oracletypes.PortID
	s.oracle = "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"

	testutil.MutateGenesisState(s.T(), &s.cfg, oracletypes.ModuleName, &oracletypes.GenesisState{}, func(oracleData *oracletypes.GenesisState) *oracletypes.GenesisState {
		oracleData.PortId = s.port
		oracleData.Oracle = s.oracle
		return oracleData
	})

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err, "network.New")

	s.network.Validators[0].ClientCtx = s.network.Validators[0].ClientCtx.WithKeyringDir(s.keyringDir).WithKeyring(s.keyring)

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
			testcli.NewCLITxExecutor(cmd, args).
				WithExpCode(tc.expectedCode).
				WithExpInRawLog([]string{tc.expectErrMsg}).
				Execute(s.T(), s.network)
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

			testcli.NewCLITxExecutor(cmd, args).
				WithExpInErrMsg([]string{tc.expectErrMsg}).
				WithExpCode(tc.expectedCode).
				Execute(s.T(), s.network)
		})
	}
}
