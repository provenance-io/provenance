package cli_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/testutil"

	epochtypes "github.com/provenance-io/provenance/x/epoch/types"
	rewardcli "github.com/provenance-io/provenance/x/reward/client/cli"
	"github.com/provenance-io/provenance/x/reward/types"
	rewardtypes "github.com/provenance-io/provenance/x/reward/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network

	accountAddr sdk.AccAddress
	accountKey  *secp256k1.PrivKey
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	s.accountKey = secp256k1.GenPrivKeyFromSecret([]byte("acc2"))
	addr, err := sdk.AccAddressFromHex(s.accountKey.PubKey().Address().String())
	s.Require().NoError(err)
	s.accountAddr = addr

	s.cfg = testutil.DefaultTestNetworkConfig()

	genesisState := s.cfg.GenesisState
	s.cfg.NumValidators = 1

	epochData := epochtypes.NewGenesisState([]epochtypes.EpochInfo{
		{Identifier: "minute",
			StartHeight:             0,
			Duration:                uint64((60) / 5),
			CurrentEpoch:            0,
			CurrentEpochStartHeight: 0,
		},
	})
	epochDataBz, err := s.cfg.Codec.MarshalJSON(epochData)
	s.Require().NoError(err)
	genesisState[epochtypes.ModuleName] = epochDataBz

	rewardData := rewardtypes.NewGenesisState(
		[]rewardtypes.RewardProgram{
			types.NewRewardProgram(
				1,
				s.accountAddr.String(),
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				"minute",
				0,
				1,
				rewardtypes.NewEligibilityCriteria("action-name", &rewardtypes.ActionDelegate{}),
				false,
				1,
				2,
			),
			types.NewRewardProgram(
				2,
				s.accountAddr.String(),
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				"minute",
				100,
				1,
				rewardtypes.NewEligibilityCriteria("action-name", &rewardtypes.ActionDelegate{}),
				false,
				1,
				2,
			),
		},
		[]rewardtypes.RewardClaim{},
		[]rewardtypes.EpochRewardDistribution{},
		[]rewardtypes.EligibilityCriteria{},
		rewardtypes.ActionDelegate{},
		rewardtypes.ActionTransferDelegations{},
	)

	rewardDataBz, err := s.cfg.Codec.MarshalJSON(rewardData)
	s.Require().NoError(err)
	genesisState[rewardtypes.ModuleName] = rewardDataBz

	s.cfg.GenesisState = genesisState

	s.cfg.ChainID = antewrapper.SimAppChainID

	s.network = network.New(s.T(), s.cfg)

	_, err = s.network.WaitForHeight(3)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestQueryRewardPrograms() {
	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectErrMsg   string
		expectedCode   uint32
		expectedOutput string
	}{
		{"query all reward programs",
			[]string{
				"all",
			},
			false,
			"",
			0,
			"{\"reward_programs\":[{\"id\":\"1\",\"distribute_from_address\":\"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h\",\"coin\":{\"denom\":\"jackthecat\",\"amount\":\"1\"},\"max_reward_by_address\":{\"denom\":\"jackthecat\",\"amount\":\"2\"},\"epoch_id\":\"minute\",\"start_epoch\":\"0\",\"number_epochs\":\"1\",\"eligibility_criteria\":{\"name\":\"action-name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}},\"expired\":false,\"minimum\":\"1\",\"maximum\":\"2\"},{\"id\":\"2\",\"distribute_from_address\":\"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h\",\"coin\":{\"denom\":\"jackthecat\",\"amount\":\"1\"},\"max_reward_by_address\":{\"denom\":\"jackthecat\",\"amount\":\"2\"},\"epoch_id\":\"minute\",\"start_epoch\":\"100\",\"number_epochs\":\"1\",\"eligibility_criteria\":{\"name\":\"action-name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}},\"expired\":false,\"minimum\":\"1\",\"maximum\":\"2\"}]}"},
		{"query all active reward programs",
			[]string{
				"active",
			},
			false,
			"",
			0,
			"{\"reward_programs\":[{\"id\":\"1\",\"distribute_from_address\":\"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h\",\"coin\":{\"denom\":\"jackthecat\",\"amount\":\"1\"},\"max_reward_by_address\":{\"denom\":\"jackthecat\",\"amount\":\"2\"},\"epoch_id\":\"minute\",\"start_epoch\":\"0\",\"number_epochs\":\"1\",\"eligibility_criteria\":{\"name\":\"action-name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}},\"expired\":false,\"minimum\":\"1\",\"maximum\":\"2\"}]}",
		},
	}

	// Wait for block 2 just in case
	s.network.WaitForNextBlock()

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := s.network.Validators[0].ClientCtx

			// Set the height of the client ctx just in case?
			// Not sure if needed
			latestHeight, _ := s.network.LatestHeight()
			clientCtx.WithHeight(latestHeight)

			out, err := clitestutil.ExecTestCLICmd(clientCtx, rewardcli.GetRewardProgramCmd(), tc.args)
			if tc.expectErr {
				s.Assert().Error(err)
				s.Assert().Equal(tc.expectErrMsg, err.Error())
			} else {
				s.Assert().NoError(err)
				s.Assert().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryRewardClaims() {
	s.Assert().FailNow("not implemented")
	//cli.QueryRewardProgramsCmd()
	// TODO Need a way to create a reward claim before these can be implemented
}

func (s *IntegrationTestSuite) TestQueryRewardClaimsById() {
	s.Assert().FailNow("not implemented")
	//cli.QueryRewardProgramsCmd()
	// TODO Need a way to create a reward claim before these can be implemented
}

func (s *IntegrationTestSuite) TestQueryEpochDistributionReward() {
	s.Assert().FailNow("not implemented")
	//cli.QueryRewardProgramsCmd()
	// TODO Need a way to create a reward claim before these can be implemented
}

func (s *IntegrationTestSuite) TestQueryEpochDistributionRewardById() {
	s.Assert().FailNow("not implemented")
	//cli.QueryRewardProgramsCmd()
	// TODO Need a way to create a reward claim before these can be implemented
}

func (s *IntegrationTestSuite) TestCmdRewardProgramProposal() {
	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
	}{
		{"add reward program proposal - valid",
			[]string{
				"add",
				"test add reward program",
				"description",
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				fmt.Sprintf("--coin=580%s", s.cfg.BondDenom),
				"--reward-program-id=1",
				fmt.Sprintf("--dist-address=%s", s.network.Validators[0].Address.String()),
				"--epoch-id=minute",
				"--epoch-offset=100",
				"--num-epochs=10",
				"--minimum=3",
				"--maximum=10",
				"--eligibility-criteria={\"name\":\"name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}",
			},
			false,
			"",
			0,
		},
		{"add reward program proposal - invalid reward id",
			[]string{
				"add",
				"test add reward program",
				"description",
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				"--coin=580nhash",
				"--reward-program-id=invalid",
				fmt.Sprintf("--dist-address=%s", s.network.Validators[0].Address.String()),
				"--epoch-id=day",
				"--epoch-offset=100",
				"--num-epochs=10",
				"--minimum=3",
				"--maximum=10",
				"--eligibility-criteria={\"name\":\"name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}",
			},
			true,
			"invalid argument \"invalid\" for \"--reward-program-id\" flag: strconv.ParseUint: parsing \"invalid\": invalid syntax",
			0,
		},
		{"add reward program proposal - invalid coin",
			[]string{
				"add",
				"test add reward program",
				"description",
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				"--coin=invalid",
				"--reward-program-id=1",
				fmt.Sprintf("--dist-address=%s", s.network.Validators[0].Address.String()),
				"--epoch-id=day",
				"--epoch-offset=100",
				"--num-epochs=10",
				"--minimum=3",
				"--maximum=10",
				"--eligibility-criteria={\"name\":\"name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}",
			},
			true,
			"invalid decimal coin expression: invalid",
			0,
		},
		{"add reward program proposal - invalid dist address",
			[]string{
				"add",
				"test add reward program",
				"description",
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				"--coin=580nhash",
				"--reward-program-id=1",
				"--dist-address=invalid",
				"--epoch-id=day",
				"--epoch-offset=100",
				"--num-epochs=10",
				"--minimum=3",
				"--maximum=10",
				"--eligibility-criteria={\"name\":\"name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}",
			},
			true,
			"invalid address for rewards program distribution from address: decoding bech32 failed: invalid bech32 string length 7",
			0,
		},
		{"add reward program proposal - invalid action",
			[]string{
				"invalid",
				"test add reward program",
				"description",
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				"--coin=580nhash",
				"--reward-program-id=1",
				fmt.Sprintf("--dist-address=%s", s.network.Validators[0].Address.String()),
				"--epoch-id=invalid",
				"--epoch-offset=100",
				"--num-epochs=10",
				"--minimum=3",
				"--maximum=10",
				"--eligibility-criteria={\"name\":\"name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}",
			},
			true,
			"unknown proposal type : invalid",
			0,
		},
		{"add reward program proposal - invalid epoch offset",
			[]string{
				"add",
				"test add reward program",
				"description",
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				"--coin=580nhash",
				"--reward-program-id=1",
				fmt.Sprintf("--dist-address=%s", s.network.Validators[0].Address.String()),
				"--epoch-id=day",
				"--epoch-offset=invalid",
				"--num-epochs=10",
				"--minimum=3",
				"--maximum=10",
				"--eligibility-criteria={\"name\":\"name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}",
			},
			true,
			"invalid argument \"invalid\" for \"--epoch-offset\" flag: strconv.ParseUint: parsing \"invalid\": invalid syntax",
			0,
		},
		{"add reward program proposal - invalid num epochs",
			[]string{
				"add",
				"test add reward program",
				"description",
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				"--coin=580nhash",
				"--reward-program-id=1",
				fmt.Sprintf("--dist-address=%s", s.network.Validators[0].Address.String()),
				"--epoch-id=day",
				"--epoch-offset=100",
				"--num-epochs=invalid",
				"--minimum=3",
				"--maximum=10",
				"--eligibility-criteria={\"name\":\"name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}",
			},
			true,
			"invalid argument \"invalid\" for \"--num-epochs\" flag: strconv.ParseUint: parsing \"invalid\": invalid syntax",
			0,
		},
		{"add reward program proposal - invalid minimum",
			[]string{
				"add",
				"test add reward program",
				"description",
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				"--coin=580nhash",
				"--reward-program-id=1",
				fmt.Sprintf("--dist-address=%s", s.network.Validators[0].Address.String()),
				"--epoch-id=day",
				"--epoch-offset=100",
				"--num-epochs=10",
				"--minimum=invalid",
				"--maximum=10",
				"--eligibility-criteria={\"name\":\"name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}",
			},
			true,
			"invalid argument \"invalid\" for \"--minimum\" flag: strconv.ParseUint: parsing \"invalid\": invalid syntax",
			0,
		},
		{"add reward program proposal - invalid maximum",
			[]string{
				"add",
				"test add reward program",
				"description",
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				"--coin=580nhash",
				"--reward-program-id=1",
				fmt.Sprintf("--dist-address=%s", s.network.Validators[0].Address.String()),
				"--epoch-id=day",
				"--epoch-offset=100",
				"--num-epochs=10",
				"--minimum=3",
				"--maximum=invalid",
				"--eligibility-criteria={\"name\":\"name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}",
			},
			true,
			"invalid argument \"invalid\" for \"--maximum\" flag: strconv.ParseUint: parsing \"invalid\": invalid syntax",
			0,
		},
		{"add reward program proposal - invalid eligibility criteria",
			[]string{
				"add",
				"test add reward program",
				"description",
				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
				"--coin=580nhash",
				"--reward-program-id=1",
				fmt.Sprintf("--dist-address=%s", s.network.Validators[0].Address.String()),
				"--epoch-id=day",
				"--epoch-offset=100",
				"--num-epochs=10",
				"--minimum=3",
				"--maximum=10",
				"--eligibility-criteria=invalid",
			},
			true,
			"unable to parse eligibility criteria : invalid character 'i' looking for beginning of value",
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := s.network.Validators[0].ClientCtx

			args := []string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.network.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			}
			tc.args = append(tc.args, args...)

			out, err := clitestutil.ExecTestCLICmd(clientCtx, rewardcli.GetCmdRewardProgramProposal(), tc.args)
			var response sdk.TxResponse
			marshalErr := clientCtx.JSONCodec.UnmarshalJSON(out.Bytes(), &response)
			if tc.expectErr {
				s.Assert().Error(err)
				s.Assert().Equal(tc.expectErrMsg, err.Error())
			} else {
				s.Assert().NoError(err)
				s.Assert().NoError(marshalErr, out.String())
			}
		})
	}
}
