package cli_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	testcli "github.com/provenance-io/provenance/testutil/cli"
	msgfeescli "github.com/provenance-io/provenance/x/msgfees/client/cli"
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

	_, err = testutil.WaitForHeight(s.testnet, 1)
	s.Require().NoError(err, "waiting for height 1")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	testutil.Cleanup(s.testnet, s.T())
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
		expectErrMsg  string
		expectedCode  uint32
	}{
		{
			name:          "add msg based fee proposal - valid",
			typeArg:       "add",
			title:         "test add msg based fee",
			description:   "description",
			deposit:       sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String(),
			msgType:       "--msg-type=/provenance.metadata.v1.MsgWriteRecordRequest",
			additionalFee: fmt.Sprintf("--additional-fee=%s", sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 12)).String()),
			recipient:     "",
			bips:          "",
			expectErrMsg:  "",
			expectedCode:  0,
		},
		{
			name:          "add msg based fee proposal with recipient - valid",
			typeArg:       "add",
			title:         "test update msg based fee",
			description:   "description",
			deposit:       sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String(),
			msgType:       "--msg-type=/provenance.metadata.v1.MsgWriteRecordRequest",
			additionalFee: fmt.Sprintf("--additional-fee=%s", sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 12)).String()),
			recipient:     "",
			bips:          "",
			expectErrMsg:  "",
			expectedCode:  0,
		},
		{
			name:          "add msg based fee proposal with recipient and default basis points - valid",
			typeArg:       "add",
			title:         "test update msg based fee",
			description:   "description",
			deposit:       sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String(),
			msgType:       "--msg-type=/provenance.metadata.v1.MsgWriteRecordRequest",
			additionalFee: fmt.Sprintf("--additional-fee=%s", sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 12)).String()),
			recipient:     fmt.Sprintf("--recipient=%s", s.testnet.Validators[0].Address.String()),
			bips:          "",
			expectErrMsg:  "",
			expectedCode:  0,
		},
		{
			name:          "add msg based fee proposal - invalid msg type url",
			typeArg:       "add",
			title:         "test add msg based fee",
			description:   "description",
			deposit:       sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String(),
			msgType:       "--msg-type=invalid",
			additionalFee: fmt.Sprintf("--additional-fee=%s", sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 12)).String()),
			recipient:     "",
			bips:          "",
			expectErrMsg:  "unable to resolve type URL invalid",
			expectedCode:  0,
		},
		{
			name:          "add msg based fee proposal - invalid additional fee",
			typeArg:       "add",
			title:         "test add msg based fee",
			description:   "description",
			deposit:       sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String(),
			msgType:       "--msg-type=/provenance.metadata.v1.MsgWriteRecordRequest",
			additionalFee: fmt.Sprintf("--additional-fee=%s", "blah"),
			recipient:     "",
			bips:          "",
			expectErrMsg:  "invalid decimal coin expression: blah",
			expectedCode:  0,
		},
		{
			name:          "update msg based fee proposal with recipient - valid",
			typeArg:       "update",
			title:         "test update msg based fee",
			description:   "description",
			deposit:       sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String(),
			msgType:       "--msg-type=/provenance.metadata.v1.MsgAddContractSpecToScopeSpecRequest",
			additionalFee: fmt.Sprintf("--additional-fee=%s", sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 12)).String()),
			recipient:     "",
			bips:          "",
			expectErrMsg:  "",
			expectedCode:  0,
		},
		{
			name:          "update msg based fee proposal with recipient and default basis points - valid",
			typeArg:       "update",
			title:         "test update msg based fee",
			description:   "description",
			deposit:       sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String(),
			msgType:       "--msg-type=/provenance.metadata.v1.MsgAddContractSpecToScopeSpecRequest",
			additionalFee: fmt.Sprintf("--additional-fee=%s", sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 12)).String()),
			recipient:     fmt.Sprintf("--recipient=%s", s.testnet.Validators[0].Address.String()),
			bips:          "",
			expectErrMsg:  "",
			expectedCode:  0,
		},
		{
			name:          "update msg based fee proposal with recipient and modified basis points - valid",
			typeArg:       "update",
			title:         "test update msg based fee",
			description:   "description",
			deposit:       sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String(),
			msgType:       "--msg-type=/provenance.metadata.v1.MsgAddContractSpecToScopeSpecRequest",
			additionalFee: fmt.Sprintf("--additional-fee=%s", sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 12)).String()),
			recipient:     fmt.Sprintf("--recipient=%s", s.testnet.Validators[0].Address.String()),
			bips:          fmt.Sprintf("--bips=%s", "5001"),
			expectErrMsg:  "",
			expectedCode:  0,
		},
		{
			name:          "remove msg based fee proposal - valid",
			typeArg:       "remove",
			title:         "test update msg based fee",
			description:   "description",
			deposit:       sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String(),
			msgType:       "--msg-type=/provenance.metadata.v1.MsgAddContractSpecToScopeSpecRequest",
			additionalFee: "",
			recipient:     "",
			bips:          "",
			expectErrMsg:  "",
			expectedCode:  0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := msgfeescli.GetCmdMsgFeesProposal()
			args := []string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
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

			testcli.NewCLITxExecutor(cmd, args).
				WithExpErrMsg(tc.expectErrMsg).
				WithExpCode(tc.expectedCode).
				Execute(s.T(), s.testnet)
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
		expectErrMsg string
		expectedCode uint32
	}{
		{
			name:         "update nhash to usd mil proposal - valid",
			title:        "title",
			description:  "description",
			rate:         "10",
			deposit:      sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String(),
			expectErrMsg: "",
			expectedCode: 0,
		},
		{
			name:         "update nhash to usd mil proposal - invalid - rate param error",
			title:        "title",
			description:  "description",
			rate:         "invalid-rate",
			deposit:      sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String(),
			expectErrMsg: "unable to parse nhash value: invalid-rate",
			expectedCode: 0,
		},
		{
			name:         "update nhash to usd mil proposal - invalid - deposit param",
			title:        "title",
			description:  "description",
			rate:         "10",
			deposit:      "invalid-deposit",
			expectErrMsg: "invalid decimal coin expression: invalid-deposit",
			expectedCode: 0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := msgfeescli.GetUpdateNhashPerUsdMilProposal()
			args := []string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			}
			args = append(args, tc.name, tc.description, tc.rate, tc.deposit)

			testcli.NewCLITxExecutor(cmd, args).
				WithExpErrMsg(tc.expectErrMsg).
				WithExpCode(tc.expectedCode).
				Execute(s.T(), s.testnet)
		})
	}
}

func (s *IntegrationTestSuite) TestUpdateConversionFeeDenomProposal() {
	testCases := []struct {
		name               string
		title              string
		description        string
		conversionFeeDenom string
		deposit            string
		expectErrMsg       string
		expectedCode       uint32
	}{
		{
			name:               "update nhash to usd mil proposal - valid",
			title:              "title",
			description:        "description",
			conversionFeeDenom: "jackthecat",
			deposit:            sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String(),
			expectErrMsg:       "",
			expectedCode:       0,
		},
		{
			name:               "update nhash to usd mil proposal - invalid - deposit param",
			title:              "title",
			description:        "description",
			conversionFeeDenom: "jackthecat",
			deposit:            "invalid-deposit",
			expectErrMsg:       "invalid decimal coin expression: invalid-deposit",
			expectedCode:       0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := msgfeescli.GetUpdateConversionFeeDenomProposal()
			args := []string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			}
			args = append(args, tc.name, tc.description, tc.conversionFeeDenom, tc.deposit)

			testcli.NewCLITxExecutor(cmd, args).
				WithExpErrMsg(tc.expectErrMsg).
				WithExpCode(tc.expectedCode).
				Execute(s.T(), s.testnet)
		})
	}
}
