package cli_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil"
	msgfeescli "github.com/provenance-io/provenance/x/msgfees/client/cli"
	msgfeetypes "github.com/provenance-io/provenance/x/msgfees/types"
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
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.accountKey = secp256k1.GenPrivKeyFromSecret([]byte("acc2"))
	addr, err := sdk.AccAddressFromHex(s.accountKey.PubKey().Address().String())
	s.Require().NoError(err)
	s.accountAddr = addr

	s.account2Key = secp256k1.GenPrivKeyFromSecret([]byte("acc22"))
	addr2, err2 := sdk.AccAddressFromHex(s.account2Key.PubKey().Address().String())
	s.Require().NoError(err2)
	s.account2Addr = addr2
	s.acc2NameCount = 50

	s.T().Log("setting up integration test suite")

	cfg := testutil.DefaultTestNetworkConfig()

	genesisState := cfg.GenesisState
	cfg.NumValidators = 1

	cfg.GenesisState = genesisState

	s.cfg = cfg
	msgfeetypes.DefaultFloorGasPrice = sdk.NewCoin("atom", sdk.NewInt(0))
	s.testnet = testnet.New(s.T(), cfg)

	_, err = s.testnet.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	testutil.CleanUp(s.testnet, s.T())
}

func (s *IntegrationTestSuite) TestMsgFeesTxGovProposals() {
	testCases := []struct {
		name          string
		typeArg       string
		title         string
		description   string
		deposit       string
		msgType       string
		additionalFee string
		recipient     string
		bips          string
		expectErr     bool
		expectErrMsg  string
		expectedCode  uint32
	}{
		{"add msg based fee proposal - valid",
			"add",
			"test add msg based fee",
			"description",
			sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
			"--msg-type=/provenance.metadata.v1.MsgWriteRecordRequest",
			fmt.Sprintf("--additional-fee=%s", sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(12))).String()),
			"",
			"",
			false,
			"",
			0,
		},
		{"add msg based fee proposal with recipient - valid",

			"add",
			"test update msg based fee",
			"description",
			sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
			"--msg-type=/provenance.metadata.v1.MsgWriteRecordRequest",
			fmt.Sprintf("--additional-fee=%s", sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(12))).String()),
			"",
			"",
			false,
			"",
			0,
		},
		{"add msg based fee proposal with recipient and default basis points - valid",

			"add",
			"test update msg based fee",
			"description",
			sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
			"--msg-type=/provenance.metadata.v1.MsgWriteRecordRequest",
			fmt.Sprintf("--additional-fee=%s", sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(12))).String()),
			fmt.Sprintf("--recipient=%s", s.testnet.Validators[0].Address.String()),
			"",
			false,
			"",
			0,
		},
		{"add msg based fee proposal - invalid msg type url",

			"add",
			"test add msg based fee",
			"description",
			sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
			"--msg-type=invalid",
			fmt.Sprintf("--additional-fee=%s", sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(12))).String()),
			"",
			"",
			true,
			"unable to resolve type URL invalid",
			0,
		},
		{"add msg based fee proposal - invalid additional fee",

			"add",
			"test add msg based fee",
			"description",
			sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
			"--msg-type=/provenance.metadata.v1.MsgWriteRecordRequest",
			fmt.Sprintf("--additional-fee=%s", "blah"),
			"",
			"",
			true,
			"invalid decimal coin expression: blah",
			0,
		},
		{"update msg based fee proposal with recipient - valid",
			"update",
			"test update msg based fee",
			"description",
			sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
			"--msg-type=/provenance.metadata.v1.MsgWriteRecordRequest",
			fmt.Sprintf("--additional-fee=%s", sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(12))).String()),
			"",
			"",
			false,
			"",
			0,
		},
		{"update msg based fee proposal with recipient and default basis points - valid",
			"update",
			"test update msg based fee",
			"description",
			sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
			"--msg-type=/provenance.metadata.v1.MsgWriteRecordRequest",
			fmt.Sprintf("--additional-fee=%s", sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(12))).String()),
			fmt.Sprintf("--recipient=%s", s.testnet.Validators[0].Address.String()),
			"",
			false,
			"",
			0,
		},
		{"update msg based fee proposal with recipient and modified basis points - valid",

			"update",
			"test update msg based fee",
			"description",
			sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
			"--msg-type=/provenance.metadata.v1.MsgWriteRecordRequest",
			fmt.Sprintf("--additional-fee=%s", sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(12))).String()),
			fmt.Sprintf("--recipient=%s", s.testnet.Validators[0].Address.String()),
			fmt.Sprintf("--bips=%s", "5001"),
			false,
			"",
			0,
		},
		{"remove msg based fee proposal - valid",
			"remove",
			"test update msg based fee",
			"description",
			sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
			"--msg-type=/provenance.metadata.v1.MsgWriteRecordRequest",
			"",
			"",
			"",
			false,
			"",
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := s.testnet.Validators[0].ClientCtx

			args := []string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			}

			args = append(args, tc.typeArg, tc.title, tc.description, tc.deposit, tc.msgType)
			if len(tc.additionalFee) != 0 {
				args = append(args, tc.additionalFee)
			}
			if len(tc.recipient) != 0 {
				args = append(args, tc.recipient)
			}
			if len(tc.bips) != 0 {
				args = append(args, tc.bips)
			}

			out, err := clitestutil.ExecTestCLICmd(clientCtx, msgfeescli.GetCmdMsgFeesProposal(), args)
			if tc.expectErr {
				s.Require().Error(err)
				s.Assert().Equal(tc.expectErrMsg, err.Error())
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONCodec.UnmarshalJSON(out.Bytes(), &sdk.TxResponse{}), out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestUpdateUsdConversionRateProposal() {
	testCases := []struct {
		name         string
		title        string
		description  string
		rate         string
		deposit      string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
	}{
		{"update nhash to usd mil proposal - valid",
			"title",
			"description",
			"10",
			sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
			false,
			"",
			0,
		},
		{"update nhash to usd mil proposal - invalid - rate param error",
			"title",
			"description",
			"invalid-rate",
			sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
			true,
			"unable to parse nhash value: invalid-rate",
			0,
		},
		{"update nhash to usd mil proposal - invalid - deposit param",
			"title",
			"description",
			"10",
			"invalid-deposit",
			true,
			"invalid decimal coin expression: invalid-deposit",
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := s.testnet.Validators[0].ClientCtx
			args := []string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			}
			args = append(args, tc.name, tc.description, tc.rate, tc.deposit)

			out, err := clitestutil.ExecTestCLICmd(clientCtx, msgfeescli.GetUpdateNhashPerUsdMilProposal(), args)
			if tc.expectErr {
				s.Require().Error(err)
				s.Assert().Equal(tc.expectErrMsg, err.Error())
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONCodec.UnmarshalJSON(out.Bytes(), &sdk.TxResponse{}), out.String())
			}
		})
	}
}
