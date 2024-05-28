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
	ibchookscli "github.com/provenance-io/provenance/x/ibchooks/client/cli"
	ibchookstypes "github.com/provenance-io/provenance/x/ibchooks/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg              network.Config
	testnet          *network.Network
	keyring          keyring.Keyring
	keyringDir       string
	accountAddr      sdk.AccAddress
	accountKey       *secp256k1.PrivKey
	accountAddresses []sdk.AccAddress
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
	s.generateAccountsWithKeyrings(2)

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

	ibchooksParams := ibchookstypes.DefaultParams()
	ibchooksGenesis := &ibchookstypes.GenesisState{Params: ibchooksParams}

	ibchooksDataBz, err := s.cfg.Codec.MarshalJSON(ibchooksGenesis)
	s.Require().NoError(err, "should be able to marshal ibchooks genesis state when setting up suite")
	genesisState[ibchookstypes.ModuleName] = ibchooksDataBz

	s.cfg.GenesisState = genesisState

	s.cfg.ChainID = antewrapper.SimAppChainID

	s.testnet, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err, "network.New")

	// s.testnet.Validators[0].ClientCtx = s.testnet.Validators[0].ClientCtx.WithKeyringDir(s.keyringDir).WithKeyring(s.keyring)
	_, err = testutil.WaitForHeight(s.testnet, 6)
	s.Require().NoError(err, "WaitForHeight")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	testutil.Cleanup(s.testnet, s.T())
}

func (s *IntegrationTestSuite) generateAccountsWithKeyrings(number int) {
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

func (s *IntegrationTestSuite) TestQueryParams() {
	clientCtx := s.testnet.Validators[0].ClientCtx
	cmd := ibchookscli.GetCmdQueryParams()

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{fmt.Sprintf("--%s=json", cmtcli.OutputFlag)})
	s.Require().NoError(err)

	var response ibchookstypes.QueryParamsResponse
	s.NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response))
	expectedParams := ibchookstypes.DefaultParams()
	s.Equal(expectedParams, response.Params, "should have the default params")
}

func (s *IntegrationTestSuite) TestUpdateParamsCmd() {
	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
		expectedCode uint32
	}{
		{
			name:         "success - update allowed async ack contracts",
			args:         []string{fmt.Sprintf("%v,%v", s.accountAddr.String(), s.accountAddresses[0].String())},
			expectedCode: 0,
		},
		{
			name:         "failure - invalid args",
			args:         []string{"contract1"},
			expectErrMsg: "invalid contract address: contract1",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := ibchookscli.NewUpdateParamsCmd()
			tc.args = append(tc.args,
				"--title", fmt.Sprintf("title: %v", tc.name),
				"--summary", fmt.Sprintf("summary: %v", tc.name),
				"--deposit=1000000stake",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
				fmt.Sprintf("--%s=json", cmtcli.OutputFlag),
			)

			testcli.NewTxExecutor(cmd, tc.args).
				WithExpErrMsg(tc.expectErrMsg).
				WithExpCode(tc.expectedCode).
				Execute(s.T(), s.testnet)
		})
	}
}
