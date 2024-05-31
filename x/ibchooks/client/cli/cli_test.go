package cli_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	cmtcli "github.com/cometbft/cometbft/libs/cli"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	testcli "github.com/provenance-io/provenance/testutil/cli"
	ibchookscli "github.com/provenance-io/provenance/x/ibchooks/client/cli"
	"github.com/provenance-io/provenance/x/ibchooks/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg         network.Config
	testnet     *network.Network
	accountAddr sdk.AccAddress
	accountKey  *secp256k1.PrivKey
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

	testutil.MutateGenesisState(s.T(), &s.cfg, types.ModuleName, &types.GenesisState{}, func(ibchooks *types.GenesisState) *types.GenesisState {
		ibchooksParams := types.DefaultParams()
		ibchooksGenesis := &types.GenesisState{Params: ibchooksParams}
		return ibchooksGenesis
	})

	s.cfg.GenesisState = genesisState

	s.cfg.ChainID = antewrapper.SimAppChainID

	s.testnet, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err, "network.New")

	_, err = testutil.WaitForHeight(s.testnet, 6)
	s.Require().NoError(err, "WaitForHeight")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	testutil.Cleanup(s.testnet, s.T())
}

func (s *IntegrationTestSuite) TestQueryParams() {
	clientCtx := s.testnet.Validators[0].ClientCtx
	cmd := ibchookscli.GetCmdQueryParams()

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{fmt.Sprintf("--%s=json", cmtcli.OutputFlag)})
	s.Require().NoError(err)

	var response types.QueryParamsResponse
	s.NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response))
	expectedParams := types.DefaultParams()
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
			args:         []string{fmt.Sprintf("%v,%v", s.accountAddr.String(), sdk.AccAddress("input111111111111111").String())},
			expectedCode: 0,
		},
		{
			name:         "failure - invalid args",
			args:         []string{"contract1"},
			expectErrMsg: `invalid contract address: "contract1": decoding bech32 failed: invalid separator index 8`,
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
