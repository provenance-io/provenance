package cmd_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"

	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/provenance/cmd/provenanced/cmd"
	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/testutil"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
)

type SimulateTestSuite struct {
	suite.Suite

	cfg     testnet.Config
	testnet *testnet.Network

	sendMsgTypeUrl       string
	sendMsgAdditionalFee sdk.Coin

	accountKey  *secp256k1.PrivKey
	accountAddr sdk.AccAddress

	account2Key  *secp256k1.PrivKey
	account2Addr sdk.AccAddress

	floorGasPrice sdk.Coin
}

func (s *SimulateTestSuite) SetupTest() {
	s.accountKey = secp256k1.GenPrivKeyFromSecret([]byte("acc2"))
	addr, err := sdk.AccAddressFromHex(s.accountKey.PubKey().Address().String())
	s.Require().NoError(err)
	s.accountAddr = addr

	s.account2Key = secp256k1.GenPrivKeyFromSecret([]byte("acc22"))
	addr2, err2 := sdk.AccAddressFromHex(s.account2Key.PubKey().Address().String())
	s.Require().NoError(err2)
	s.account2Addr = addr2

	s.floorGasPrice = sdk.Coin{
		Denom:  "stake",
		Amount: sdk.NewInt(1000),
	}
	msgfeestypes.DefaultFloorGasPrice = s.floorGasPrice

	s.sendMsgTypeUrl = "/cosmos.bank.v1beta1.MsgSend"
	s.sendMsgAdditionalFee = sdk.NewCoin("stake", sdk.NewInt(1))

	s.T().Log("setting up integration test suite")

	cfg := testutil.DefaultTestNetworkConfig()

	genesisState := cfg.GenesisState
	var msgfeesData msgfeestypes.GenesisState
	msgfeesData.Params.FloorGasPrice = s.floorGasPrice
	msgfeesData.MsgFees = append(msgfeesData.MsgFees, msgfeestypes.NewMsgFee(s.sendMsgTypeUrl, s.sendMsgAdditionalFee))
	msgFeesDataBz, err := cfg.Codec.MarshalJSON(&msgfeesData)
	s.Require().NoError(err)
	genesisState[msgfeestypes.ModuleName] = msgFeesDataBz

	cfg.NumValidators = 1

	cfg.GenesisState = genesisState

	s.cfg = cfg
	cfg.ChainID = antewrapper.SimAppChainID
	s.testnet = testnet.New(s.T(), cfg)

	_, err = s.testnet.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *SimulateTestSuite) TearDownTest() {
	testutil.CleanUp(s.testnet, s.T())
}

func TestSimulateTestSuite(t *testing.T) {
	suite.Run(t, new(SimulateTestSuite))
}

func (s *SimulateTestSuite) TestSimulateCmd() {
	signedTx := s.GenerateAndSignSend(s.testnet.Validators[0].Address.String(), s.accountAddr.String(), fmt.Sprintf("3%s", s.cfg.BondDenom))
	testCases := []struct {
		name                   string
		args                   []string
		expectedAdditionalFees sdk.Coins
	}{
		{
			"should succeed with additional fees on send in same denom as gas",
			[]string{signedTx, "-o", "json", "--default-denom", "stake"},
			sdk.NewCoins(s.sendMsgAdditionalFee),
		},
	}
	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cmd.GetCmdPioSimulateTx()
			clientCtx := s.testnet.Validators[0].ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			var result msgfeestypes.CalculateTxFeesResponse
			err = s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			s.Require().NoError(err)
			s.Assert().Equal(tc.expectedAdditionalFees, result.AdditionalFees)
			expectedTotalFees := sdk.NewCoins(s.sendMsgAdditionalFee.Add(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(int64(result.EstimatedGas)*s.floorGasPrice.Amount.Int64()))))
			s.Assert().Equal(expectedTotalFees, result.TotalFees)
		})
	}
}

func (s *SimulateTestSuite) GenerateAndSignSend(from string, to string, coins string) string {
	tmpDir := s.T().TempDir()
	clientCtx := s.testnet.Validators[0].ClientCtx
	args := []string{
		from,
		to,
		coins,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		"--generate-only",
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}
	out, err := clitestutil.ExecTestCLICmd(clientCtx, bankcli.NewSendTxCmd(), args)
	s.Require().NoError(err)
	genFilePath := filepath.Join(tmpDir, fmt.Sprintf("unsigned.%s.%s.%s.json", from, to, coins))
	f, err := os.Create(genFilePath)
	s.Require().NoError(err)
	f.WriteString(out.String())
	out, err = clitestutil.ExecTestCLICmd(clientCtx, authcli.GetSignCommand(), []string{
		genFilePath,
		fmt.Sprintf("--chain-id=%s", s.testnet.Validators[0].ClientCtx.ChainID),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
	})
	s.Require().NoError(err)
	signedFilePath := fmt.Sprintf("%s.signed", genFilePath)
	f, err = os.Create(signedFilePath)
	s.Require().NoError(err)
	f.WriteString(out.String())
	return signedFilePath
}
