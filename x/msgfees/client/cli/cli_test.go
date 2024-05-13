package cli_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	cmtcli "github.com/cometbft/cometbft/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	testcli "github.com/provenance-io/provenance/testutil/cli"
	"github.com/provenance-io/provenance/x/msgfees/client/cli"
	"github.com/provenance-io/provenance/x/msgfees/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     testnet.Config
	testnet *testnet.Network

	accountAddr sdk.AccAddress
	accountKey  *secp256k1.PrivKey

	account2Addr  sdk.AccAddress
	account2Key   *secp256k1.PrivKey
	acc2NameCount int

	accountAddresses []sdk.AccAddress

	keyring    keyring.Keyring
	keyringDir string
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.accountKey = secp256k1.GenPrivKeyFromSecret([]byte("acc2"))
	addr, err := sdk.AccAddressFromHexUnsafe(s.accountKey.PubKey().Address().String())
	s.Require().NoError(err)
	s.accountAddr = addr

	s.account2Key = secp256k1.GenPrivKeyFromSecret([]byte("acc22"))
	addr2, err2 := sdk.AccAddressFromHexUnsafe(s.account2Key.PubKey().Address().String())
	s.Require().NoError(err2)
	s.account2Addr = addr2
	s.acc2NameCount = 50

	s.T().Log("setting up integration test suite")
	pioconfig.SetProvenanceConfig("atom", 0)
	govv1.DefaultMinDepositRatio = sdkmath.LegacyZeroDec()

	s.cfg = testutil.DefaultTestNetworkConfig()
	s.cfg.TimeoutCommit = 500 * time.Millisecond
	s.cfg.NumValidators = 1
	s.GenerateAccountsWithKeyrings(1)

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
	s.cfg.GenesisState[banktypes.ModuleName] = bankDataBz

	var authData authtypes.GenesisState
	var genAccounts []authtypes.GenesisAccount
	authData.Params = authtypes.DefaultParams()
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.accountAddresses[0], nil, 3, 0))
	accounts, err := authtypes.PackAccounts(genAccounts)
	s.Require().NoError(err, "should be able to pack accounts for genesis state when setting up suite")
	authData.Accounts = accounts
	authDataBz, err := s.cfg.Codec.MarshalJSON(&authData)
	s.Require().NoError(err, "should be able to marshal auth genesis state when setting up suite")
	s.cfg.GenesisState[authtypes.ModuleName] = authDataBz

	var msgfeeGen types.GenesisState
	err = s.cfg.Codec.UnmarshalJSON(s.cfg.GenesisState[types.ModuleName], &msgfeeGen)
	s.Require().NoError(err, "UnmarshalJSON msgfee gen state")
	msgfeeGen.MsgFees = append(msgfeeGen.MsgFees, types.MsgFee{
		MsgTypeUrl:    "/provenance.metadata.v1.MsgAddContractSpecToScopeSpecRequest",
		AdditionalFee: sdk.NewInt64Coin(s.cfg.BondDenom, 3),
	})
	s.cfg.GenesisState[types.ModuleName], err = s.cfg.Codec.MarshalJSON(&msgfeeGen)
	s.Require().NoError(err, "MarshalJSON msgfee gen state")

	s.testnet, err = testnet.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err, "creating testnet")

	s.testnet.Validators[0].ClientCtx = s.testnet.Validators[0].ClientCtx.WithKeyringDir(s.keyringDir).WithKeyring(s.keyring)
	_, err = testutil.WaitForHeight(s.testnet, 1)
	s.Require().NoError(err, "waiting for height 1")
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
		s.Require().NoError(err, "Keyring.NewMnemonic")
		addr, err := info.GetAddress()
		s.Require().NoError(err, "GetAddress")
		s.accountAddresses = append(s.accountAddresses, addr)
	}
}

func (s *IntegrationTestSuite) TearDownSuite() {
	testutil.Cleanup(s.testnet, s.T())
}

func (s *IntegrationTestSuite) TestMsgFeesProposal() {
	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
		expectedCode uint32
		signer       string
	}{
		{
			name: "success - add msg fee",
			args: []string{
				"add",
				"--msg-type=/provenance.metadata.v1.MsgWriteRecordRequest",
				"--additional-fee=612nhash",
			},
			expectedCode: 0,
			signer:       s.accountAddresses[0].String(),
		},
		{
			name: "success - update msg fee",
			args: []string{
				"update",
				"--msg-type=/provenance.metadata.v1.MsgWriteRecordRequest",
				"--additional-fee=612000nhash",
			},
			expectedCode: 0,
			signer:       s.accountAddresses[0].String(),
		},
		{
			name: "success - remove msg fee",
			args: []string{
				"remove",
				"--msg-type=/provenance.metadata.v1.MsgWriteRecordRequest",
			},
			expectedCode: 0,
			signer:       s.accountAddresses[0].String(),
		},
		{
			name:         "failure - invalid number of args",
			args:         []string{"add", "extra-arg"},
			expectErrMsg: "accepts 1 arg(s), received 2",
			signer:       s.accountAddresses[0].String(),
		},
		{
			name:         "failure - invalid proposal type",
			args:         []string{"invalid-type"},
			expectErrMsg: "unable to resolve type URL ",
			signer:       s.accountAddresses[0].String(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdMsgFeesProposal()
			tc.args = append(tc.args,
				"--title", "msg fees proposal", "--summary", "See title.",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, tc.signer),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
				fmt.Sprintf("--%s=json", cmtcli.OutputFlag),
			)

			testcli.NewCLITxExecutor(cmd, tc.args).
				WithExpErrMsg(tc.expectErrMsg).
				WithExpCode(tc.expectedCode).
				Execute(s.T(), s.testnet)
		})
	}
}

// TODO: Add query tests
