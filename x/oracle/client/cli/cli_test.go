package cli_test

import (
	"fmt"
	"testing"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
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

	params   oracletypes.Params
	sequence uint64
	port     string
	oracle   string
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	pioconfig.SetProvenanceConfig("", 0)
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
			sdk.NewCoin("nhash", sdk.NewInt(100000000)), sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(100000000)),
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

	s.sequence = uint64(1)
	s.params = oracletypes.DefaultParams()
	s.port = oracletypes.PortID
	s.oracle = "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"

	oracleData := oracletypes.NewGenesisState(
		s.port,
		s.params,
		s.sequence,
		s.oracle,
	)

	oracleDataBz, err := s.cfg.Codec.MarshalJSON(oracleData)
	s.Require().NoError(err, "should be able to marshal trigger genesis state when setting up suite")
	genesisState[oracletypes.ModuleName] = oracleDataBz

	s.cfg.GenesisState = genesisState

	s.cfg.ChainID = antewrapper.SimAppChainID

	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err, "network.New")

	_, err = s.network.WaitForHeight(6)
	s.Require().NoError(err, "WaitForHeight")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.Require().NoError(s.network.WaitForNextBlock(), "WaitForNextBlock")
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
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
			out, err := clitestutil.ExecTestCLICmd(clientCtx, oraclecli.GetQueryOracleAddressCmd(), []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
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
			expectErrMsg: "invalid address for oracle: decoding bech32 failed: invalid separator index -1",
			signer:       s.accountAddresses[0].String(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {

			clientCtx := s.network.Validators[0].ClientCtx.WithKeyringDir(s.keyringDir).WithKeyring(s.keyring)

			args := []string{
				tc.address,
			}
			flags := []string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, tc.signer),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			}
			args = append(args, flags...)

			out, err := clitestutil.ExecTestCLICmd(clientCtx, oraclecli.GetCmdOracleUpdate(), append(args, []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)}...))
			var response sdk.TxResponse
			marshalErr := clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response)
			if len(tc.expectErrMsg) > 0 {
				s.Assert().EqualError(err, tc.expectErrMsg, "should have correct error for invalid OracleUpdate request")
				s.Assert().Equal(tc.expectedCode, response.Code, "should have correct response code for invalid OracleUpdate request")
			} else {
				s.Assert().NoError(err, "should have no error for valid OracleUpdate request")
				s.Assert().NoError(marshalErr, out.String(), "should have no marshal error for valid OracleUpdate request")
				s.Assert().Equal(tc.expectedCode, response.Code, "should have correct response code for valid OracleUpdate request")
			}
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
			expectErrMsg: "abc.info: key not found",
			channel:      "channel-1",
			signer:       "abc",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {

			clientCtx := s.network.Validators[0].ClientCtx.WithKeyringDir(s.keyringDir).WithKeyring(s.keyring)

			args := []string{
				tc.channel,
				tc.query,
			}
			flags := []string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, tc.signer),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			}
			args = append(args, flags...)

			out, err := clitestutil.ExecTestCLICmd(clientCtx, oraclecli.GetCmdSendQuery(), append(args, []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)}...))
			var response sdk.TxResponse
			marshalErr := clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response)
			if len(tc.expectErrMsg) > 0 {
				s.Assert().EqualError(err, tc.expectErrMsg, "should have correct error for invalid SendQuery request")
				s.Assert().Equal(tc.expectedCode, response.Code, "should have correct response code for invalid SendQuery request")
			} else {
				s.Assert().NoError(err, "should have no error for valid SendQuery request")
				s.Assert().NoError(marshalErr, out.String(), "should have no marshal error for valid SendQuery request")
				s.Assert().Equal(tc.expectedCode, response.Code, "should have correct response code for valid SendQuery request")
			}
		})
	}
}
