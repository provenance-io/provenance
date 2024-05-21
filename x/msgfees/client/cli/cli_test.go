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

	testutil.MutateGenesisState(s.T(), &s.cfg, banktypes.ModuleName, &banktypes.GenesisState{}, func(bankGenState *banktypes.GenesisState) *banktypes.GenesisState {
		var genBalances []banktypes.Balance
		for i := range s.accountAddresses {
			genBalances = append(genBalances, banktypes.Balance{Address: s.accountAddresses[i].String(), Coins: sdk.NewCoins(
				sdk.NewInt64Coin("nhash", 100_000_000), sdk.NewInt64Coin(s.cfg.BondDenom, 100_000_000),
			).Sort()})
		}
		bankGenState.Params = banktypes.DefaultParams()
		bankGenState.Balances = genBalances
		return bankGenState
	})

	testutil.MutateGenesisState(s.T(), &s.cfg, authtypes.ModuleName, &authtypes.GenesisState{}, func(authData *authtypes.GenesisState) *authtypes.GenesisState {
		var genAccounts []authtypes.GenesisAccount
		authData.Params = authtypes.DefaultParams()
		genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.accountAddresses[0], nil, 3, 0))
		accounts, err := authtypes.PackAccounts(genAccounts)
		s.Require().NoError(err, "should be able to pack accounts for genesis state when setting up suite")
		authData.Accounts = accounts
		return authData
	})

	testutil.MutateGenesisState(s.T(), &s.cfg, types.ModuleName, &types.GenesisState{}, func(msgfeeGen *types.GenesisState) *types.GenesisState {
		err = s.cfg.Codec.UnmarshalJSON(s.cfg.GenesisState[types.ModuleName], msgfeeGen)
		s.Require().NoError(err, "UnmarshalJSON msgfee gen state")
		msgfeeGen.MsgFees = append(msgfeeGen.MsgFees, types.MsgFee{
			MsgTypeUrl:    "/provenance.metadata.v1.MsgAddContractSpecToScopeSpecRequest",
			AdditionalFee: sdk.NewInt64Coin(s.cfg.BondDenom, 3),
		})
		return msgfeeGen
	})

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
			name: "failure - invalid proposal type",
			args: []string{
				"invalid-type",
				"--msg-type=/provenance.metadata.v1.MsgWriteRecordRequest",
			},
			expectErrMsg: `unknown proposal type "invalid-type"`,
			signer:       s.accountAddresses[0].String(),
		},
		{
			name: "success - add msg fee with recipient and bips",
			args: []string{
				"add",
				"--msg-type=/provenance.metadata.v1.MsgWriteRecordRequest",
				"--additional-fee=612nhash",
				fmt.Sprintf("--recipient=%v", s.account2Addr.String()),
				"--bips=100",
				"--deposit=1000000stake",
			},
			expectedCode: 0,
			signer:       s.accountAddresses[0].String(),
		},
		{
			name: "failure - invalid fee format",
			args: []string{
				"add",
				"--msg-type=/provenance.metadata.v1.MsgWriteRecordRequest",
				"--additional-fee=invalid-fee",
			},
			expectErrMsg: "invalid decimal coin expression: invalid-fee",
			signer:       s.accountAddresses[0].String(),
		},
		{
			name: "success - update msg fee with recipient and bips",
			args: []string{
				"update",
				"--msg-type=/provenance.metadata.v1.MsgWriteRecordRequest",
				"--additional-fee=612000nhash",
				fmt.Sprintf("--recipient=%v", s.account2Addr.String()),
				"--bips=100",
				"--deposit=1000000stake",
			},
			expectedCode: 0,
			signer:       s.accountAddresses[0].String(),
		},
		{
			name: "failure - invalid recipient format",
			args: []string{
				"add",
				"--msg-type=/provenance.metadata.v1.MsgWriteRecordRequest",
				"--additional-fee=612nhash",
				"--recipient=invalid-recipient",
				"--bips=100",
				"--deposit=1000000stake",
			},
			expectErrMsg: "error validating basis points args: decoding bech32 failed: invalid separator index -1",
			signer:       s.accountAddresses[0].String(),
		},
		{
			name: "failure - bips out of range",
			args: []string{
				"add",
				"--msg-type=/provenance.metadata.v1.MsgWriteRecordRequest",
				"--additional-fee=612nhash",
				fmt.Sprintf("--recipient=%v", s.account2Addr.String()),
				"--bips=10001",
				"--deposit=1000000stake",
			},
			expectErrMsg: "error validating basis points args: recipient basis points can only be between 0 and 10,000 : 10001",
			signer:       s.accountAddresses[0].String(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetCmdMsgFeesProposal()
			tc.args = append(tc.args,
				"--title", "Update nhash per usd mil proposal", "--summary", "Updates the nhash per usd mil rate.",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, tc.signer),
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

func (s *IntegrationTestSuite) TestUpdateNhashPerUsdMilProposal() {
	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
		expectedCode uint32
		signer       string
	}{
		{
			name:         "success - valid update",
			args:         []string{"1234"},
			expectedCode: 0,
			signer:       s.accountAddresses[0].String(),
		},
		{
			name:         "failure - invalid nhash value",
			args:         []string{"invalid-nhash-value"},
			expectErrMsg: "unable to parse nhash value: invalid-nhash-value",
			signer:       s.accountAddresses[0].String(),
		},
		{
			name:         "failure - no nhash value",
			args:         []string{},
			expectErrMsg: "accepts 1 arg(s), received 0",
			signer:       s.accountAddresses[0].String(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetUpdateNhashPerUsdMilProposal()
			tc.args = append(tc.args,
				"--title", "Update nhash per usd mil proposal", "--summary", "Updates the nhash per usd mil rate.",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, tc.signer),
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

func (s *IntegrationTestSuite) TestUpdateConversionFeeDenomProposal() {
	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
		expectedCode uint32
		signer       string
	}{
		{
			name:         "success - valid denom update",
			args:         []string{"customcoin"},
			expectedCode: 0,
			signer:       s.accountAddresses[0].String(),
		},
		{
			name:         "failure - too many arguments",
			args:         []string{"customcoin", "unexpected"},
			expectErrMsg: "accepts 1 arg(s), received 2",
			signer:       s.accountAddresses[0].String(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetUpdateConversionFeeDenomProposal()
			tc.args = append(tc.args,
				"--title", "Update conversion fee denom proposal", "--summary", "Updates the conversion fee denom.",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, tc.signer),
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

// TODO: Add query tests
