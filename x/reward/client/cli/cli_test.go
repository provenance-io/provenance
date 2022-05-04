package cli_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/testutil"

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
	now := time.Now().UTC()
	nextEpochTime := now.Add(time.Hour + 24)

	rewardData := rewardtypes.NewGenesisState(
		[]rewardtypes.RewardProgram{
			types.NewRewardProgram(
				1,
				s.accountAddr.String(),
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				now,
				nextEpochTime,
				"minute",
				1,
				rewardtypes.NewEligibilityCriteria("action-name", &rewardtypes.ActionDelegate{}),
				false,
			),
			types.NewRewardProgram(
				2,
				s.accountAddr.String(),
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				now,
				nextEpochTime,
				"minute",
				1,
				rewardtypes.NewEligibilityCriteria("action-name", &rewardtypes.ActionDelegate{}),
				false,
			),
		},
		[]rewardtypes.RewardClaim{
			types.NewRewardClaim(
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				[]rewardtypes.SharesPerEpochPerRewardsProgram{
					types.NewSharesPerEpochPerRewardsProgram(
						3,
						0,
						0,
						0,
						false,
						false,
						sdk.NewInt64Coin("jackthecat", 0),
					),
				},
			),
			types.NewRewardClaim(
				"cosmos1p3sl9tll0ygj3flwt5r2w0n6fx9p5ngq2tu6mq",
				[]rewardtypes.SharesPerEpochPerRewardsProgram{
					types.NewSharesPerEpochPerRewardsProgram(
						2,
						0,
						0,
						0,
						false,
						false,
						sdk.NewInt64Coin("jackthecat", 0),
					),
				},
			),
		},
		[]rewardtypes.EpochRewardDistribution{
			types.NewEpochRewardDistribution(
				"day",
				1,
				sdk.NewInt64Coin("jackthecat", 100),
				5,
				false,
			),
			types.NewEpochRewardDistribution(
				"day",
				2,
				sdk.NewInt64Coin("jackthecat", 100),
				3,
				false,
			),
			types.NewEpochRewardDistribution(
				"month",
				1,
				sdk.NewInt64Coin("jackthecat", 100),
				10,
				false,
			),
		},
		[]rewardtypes.EligibilityCriteria{
			types.NewEligibilityCriteria(
				"test1",
				&rewardtypes.ActionDelegate{},
			),
			types.NewEligibilityCriteria(
				"test2",
				&rewardtypes.ActionDelegate{},
			),
			types.NewEligibilityCriteria(
				"test3",
				&rewardtypes.ActionDelegate{},
			),
		},
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
			"{\"reward_programs\":[{\"id\":\"1\",\"distribute_from_address\":\"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h\",\"coin\":{\"denom\":\"jackthecat\",\"amount\":\"1\"},\"max_reward_by_address\":{\"denom\":\"jackthecat\",\"amount\":\"2\"},\"epoch_id\":\"minute\",\"start_epoch\":\"0\",\"number_epochs\":\"1\",\"eligibility_criteria\":{\"name\":\"action-name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}},\"expired\":false,\"minimum\":\"1\",\"maximum\":\"2\"},{\"id\":\"2\",\"distribute_from_address\":\"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h\",\"coin\":{\"denom\":\"jackthecat\",\"amount\":\"1\"},\"max_reward_by_address\":{\"denom\":\"jackthecat\",\"amount\":\"2\"},\"epoch_id\":\"minute\",\"start_epoch\":\"100\",\"number_epochs\":\"1\",\"eligibility_criteria\":{\"name\":\"action-name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}},\"expired\":false,\"minimum\":\"1\",\"maximum\":\"2\"}]}",
		},
		{"query all active reward programs",
			[]string{
				"active",
			},
			false,
			"",
			0,
			"{\"reward_programs\":[{\"id\":\"1\",\"distribute_from_address\":\"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h\",\"coin\":{\"denom\":\"jackthecat\",\"amount\":\"1\"},\"max_reward_by_address\":{\"denom\":\"jackthecat\",\"amount\":\"2\"},\"epoch_id\":\"minute\",\"start_epoch\":\"0\",\"number_epochs\":\"1\",\"eligibility_criteria\":{\"name\":\"action-name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}},\"expired\":false,\"minimum\":\"1\",\"maximum\":\"2\"}]}",
		},
		{"query existing reward program by id",
			[]string{
				"1",
			},
			false,
			"",
			0,
			"{\"reward_program\":{\"id\":\"1\",\"distribute_from_address\":\"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h\",\"coin\":{\"denom\":\"jackthecat\",\"amount\":\"1\"},\"max_reward_by_address\":{\"denom\":\"jackthecat\",\"amount\":\"2\"},\"epoch_id\":\"minute\",\"start_epoch\":\"0\",\"number_epochs\":\"1\",\"eligibility_criteria\":{\"name\":\"action-name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}},\"expired\":false,\"minimum\":\"1\",\"maximum\":\"2\"}}",
		},
		{"query non-existing reward program by id",
			[]string{
				"3",
			},
			true,
			"reward program 3 does not exist",
			0,
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := s.network.Validators[0].ClientCtx
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
	rewardClaimsJson := fmt.Sprintf(`{"reward_claims":[{"address":"%s","shares_per_epoch_per_reward":[{"reward_program_id":"1","total_shares":"1","ephemeral_action_count":"1","latest_recorded_epoch":"1","claimed":false,"expired":false,"total_reward_claimed":{"denom":"","amount":"0"}}]},{"address":"cosmos1p3sl9tll0ygj3flwt5r2w0n6fx9p5ngq2tu6mq","shares_per_epoch_per_reward":[{"reward_program_id":"2","total_shares":"0","ephemeral_action_count":"0","latest_recorded_epoch":"0","claimed":false,"expired":false,"total_reward_claimed":{"denom":"jackthecat","amount":"0"}}]},{"address":"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h","shares_per_epoch_per_reward":[{"reward_program_id":"3","total_shares":"0","ephemeral_action_count":"0","latest_recorded_epoch":"0","claimed":false,"expired":false,"total_reward_claimed":{"denom":"jackthecat","amount":"0"}}]}]}`, s.network.Validators[0].Address.String())
	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectErrMsg   string
		expectedCode   uint32
		expectedOutput string
		compareLength  bool
	}{
		{"query all reward claims",
			[]string{
				"all",
			},
			false,
			"",
			0,
			"",
			true,
		},
		{"query existing reward claim by address",
			[]string{
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			},
			false,
			"",
			0,
			"{\"reward_claim\":{\"address\":\"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h\",\"shares_per_epoch_per_reward\":[{\"reward_program_id\":\"3\",\"total_shares\":\"0\",\"ephemeral_action_count\":\"0\",\"latest_recorded_epoch\":\"0\",\"claimed\":false,\"expired\":false,\"total_reward_claimed\":{\"denom\":\"jackthecat\",\"amount\":\"0\"}}]}}",
			false,
		},
		{"query non-existing reward claim by address",
			[]string{
				"failure",
			},
			true,
			"reward claim failure does not exist",
			0,
			"",
			false,
		},
		{"query non-existing reward claim by empty address",
			[]string{
				"",
			},
			true,
			"reward claim  does not exist",
			0,
			"",
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := s.network.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, rewardcli.GetRewardClaimCmd(), tc.args)
			if tc.expectErr {
				s.Assert().Error(err)
				s.Assert().Equal(tc.expectErrMsg, err.Error())
			} else {
				s.Assert().NoError(err)
				if tc.compareLength {
					actualResponse := rewardtypes.RewardClaimsResponse{}
					expectedResponse := rewardtypes.RewardClaimsResponse{}
					json.Unmarshal([]byte(strings.TrimSpace(out.String())), &actualResponse)
					json.Unmarshal([]byte(strings.TrimSpace(rewardClaimsJson)), &expectedResponse)
					s.Assert().Equal(len(expectedResponse.GetRewardClaims()), len(actualResponse.GetRewardClaims()), "should have all reward claims")
				} else {
					s.Assert().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryEpochDistributionReward() {
	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectErrMsg   string
		expectedCode   uint32
		expectedOutput string
	}{
		{"query all epoch distribution rewards",
			[]string{
				"all",
			},
			false,
			"",
			0,
			"{\"epoch_reward_distribution\":[{\"epoch_id\":\"day\",\"reward_program_id\":\"1\",\"total_rewards_pool\":{\"denom\":\"jackthecat\",\"amount\":\"100\"},\"total_shares\":\"5\",\"epoch_ended\":false},{\"epoch_id\":\"day\",\"reward_program_id\":\"2\",\"total_rewards_pool\":{\"denom\":\"jackthecat\",\"amount\":\"100\"},\"total_shares\":\"3\",\"epoch_ended\":false},{\"epoch_id\":\"minute\",\"reward_program_id\":\"1\",\"total_rewards_pool\":{\"denom\":\"jackthecat\",\"amount\":\"1\"},\"total_shares\":\"1\",\"epoch_ended\":false},{\"epoch_id\":\"month\",\"reward_program_id\":\"1\",\"total_rewards_pool\":{\"denom\":\"jackthecat\",\"amount\":\"100\"},\"total_shares\":\"10\",\"epoch_ended\":false}]}",
		},
		{"query existing epoch reward distribution by valid ids",
			[]string{
				"1",
				"day",
			},
			false,
			"",
			0,
			"{\"epoch_reward_distribution\":{\"epoch_id\":\"day\",\"reward_program_id\":\"1\",\"total_rewards_pool\":{\"denom\":\"jackthecat\",\"amount\":\"100\"},\"total_shares\":\"5\",\"epoch_ended\":false}}",
		},
		{"query epoch reward distribution by invalid reward program id",
			[]string{
				"10",
				"day",
			},
			true,
			"epoch reward does not exist for reward-id: 10 epoch-id day",
			0,
			"",
		},
		{"query epoch reward distribution by invalid epoch id",
			[]string{
				"1",
				"blah",
			},
			true,
			"epoch reward does not exist for reward-id: 1 epoch-id blah",
			0,
			"",
		},
		{"query epoch reward distribution by invalid reward program id and epoch id",
			[]string{
				"10",
				"blah",
			},
			true,
			"epoch reward does not exist for reward-id: 10 epoch-id blah",
			0,
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc

		fmt.Printf("Address: %s\n", s.network.Validators[0].Address.String())
		s.Run(tc.name, func() {
			clientCtx := s.network.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, rewardcli.GetEpochRewardDistributionCmd(), tc.args)
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

func (s *IntegrationTestSuite) TestQueryEligibilityCriteria() {
	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectErrMsg   string
		expectedCode   uint32
		expectedOutput string
	}{
		{"query all eligibility criteria",
			[]string{
				"all",
			},
			false,
			"",
			0,
			"{\"eligibility_criteria\":[{\"name\":\"test1\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}},{\"name\":\"test2\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}},{\"name\":\"test3\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}]}",
		},
		{"query existing eligibility criteria by valid name",
			[]string{
				"test1",
			},
			false,
			"",
			0,
			"{\"eligibility_criteria\":{\"name\":\"test1\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}}",
		},
		{"query eligibility criteria by invalid name",
			[]string{
				"blah",
			},
			true,
			"eligibility criteria does not exist for name: blah",
			0,
			"",
		},
		{"query eligibility criteria with empty name",
			[]string{
				"",
			},
			true,
			"eligibility criteria does not exist for name: ",
			0,
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc

		fmt.Printf("Address: %s\n", s.network.Validators[0].Address.String())
		s.Run(tc.name, func() {
			clientCtx := s.network.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, rewardcli.GetEligibilityCriteriaCmd(), tc.args)
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

// func (s *IntegrationTestSuite) TestCmdRewardProgramProposal() {
// 	testCases := []struct {
// 		name         string
// 		args         []string
// 		expectErr    bool
// 		expectErrMsg string
// 		expectedCode uint32
// 	}{
// 		{"add reward program proposal - valid",
// 			[]string{
// 				"add",
// 				"test add reward program",
// 				"description",
// 				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
// 				fmt.Sprintf("--coin=580%s", s.cfg.BondDenom),
// 				"--reward-program-id=1",
// 				fmt.Sprintf("--dist-address=%s", s.network.Validators[0].Address.String()),
// 				"--epoch-id=minute",
// 				"--epoch-offset=100",
// 				"--num-epochs=10",
// 				"--minimum=3",
// 				"--maximum=10",
// 				"--eligibility-criteria={\"name\":\"name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}",
// 			},
// 			false,
// 			"",
// 			0,
// 		},
// 		{"add reward program proposal - invalid reward id",
// 			[]string{
// 				"add",
// 				"test add reward program",
// 				"description",
// 				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
// 				"--coin=580nhash",
// 				"--reward-program-id=invalid",
// 				fmt.Sprintf("--dist-address=%s", s.network.Validators[0].Address.String()),
// 				"--epoch-id=day",
// 				"--epoch-offset=100",
// 				"--num-epochs=10",
// 				"--minimum=3",
// 				"--maximum=10",
// 				"--eligibility-criteria={\"name\":\"name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}",
// 			},
// 			true,
// 			"invalid argument \"invalid\" for \"--reward-program-id\" flag: strconv.ParseUint: parsing \"invalid\": invalid syntax",
// 			0,
// 		},
// 		{"add reward program proposal - invalid coin",
// 			[]string{
// 				"add",
// 				"test add reward program",
// 				"description",
// 				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
// 				"--coin=invalid",
// 				"--reward-program-id=1",
// 				fmt.Sprintf("--dist-address=%s", s.network.Validators[0].Address.String()),
// 				"--epoch-id=day",
// 				"--epoch-offset=100",
// 				"--num-epochs=10",
// 				"--minimum=3",
// 				"--maximum=10",
// 				"--eligibility-criteria={\"name\":\"name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}",
// 			},
// 			true,
// 			"invalid decimal coin expression: invalid",
// 			0,
// 		},
// 		{"add reward program proposal - invalid dist address",
// 			[]string{
// 				"add",
// 				"test add reward program",
// 				"description",
// 				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
// 				"--coin=580nhash",
// 				"--reward-program-id=1",
// 				"--dist-address=invalid",
// 				"--epoch-id=day",
// 				"--epoch-offset=100",
// 				"--num-epochs=10",
// 				"--minimum=3",
// 				"--maximum=10",
// 				"--eligibility-criteria={\"name\":\"name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}",
// 			},
// 			true,
// 			"invalid address for rewards program distribution from address: decoding bech32 failed: invalid bech32 string length 7",
// 			0,
// 		},
// 		{"add reward program proposal - invalid action",
// 			[]string{
// 				"invalid",
// 				"test add reward program",
// 				"description",
// 				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
// 				"--coin=580nhash",
// 				"--reward-program-id=1",
// 				fmt.Sprintf("--dist-address=%s", s.network.Validators[0].Address.String()),
// 				"--epoch-id=invalid",
// 				"--epoch-offset=100",
// 				"--num-epochs=10",
// 				"--minimum=3",
// 				"--maximum=10",
// 				"--eligibility-criteria={\"name\":\"name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}",
// 			},
// 			true,
// 			"unknown proposal type : invalid",
// 			0,
// 		},
// 		{"add reward program proposal - invalid epoch offset",
// 			[]string{
// 				"add",
// 				"test add reward program",
// 				"description",
// 				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
// 				"--coin=580nhash",
// 				"--reward-program-id=1",
// 				fmt.Sprintf("--dist-address=%s", s.network.Validators[0].Address.String()),
// 				"--epoch-id=day",
// 				"--epoch-offset=invalid",
// 				"--num-epochs=10",
// 				"--minimum=3",
// 				"--maximum=10",
// 				"--eligibility-criteria={\"name\":\"name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}",
// 			},
// 			true,
// 			"invalid argument \"invalid\" for \"--epoch-offset\" flag: strconv.ParseUint: parsing \"invalid\": invalid syntax",
// 			0,
// 		},
// 		{"add reward program proposal - invalid num epochs",
// 			[]string{
// 				"add",
// 				"test add reward program",
// 				"description",
// 				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
// 				"--coin=580nhash",
// 				"--reward-program-id=1",
// 				fmt.Sprintf("--dist-address=%s", s.network.Validators[0].Address.String()),
// 				"--epoch-id=day",
// 				"--epoch-offset=100",
// 				"--num-epochs=invalid",
// 				"--minimum=3",
// 				"--maximum=10",
// 				"--eligibility-criteria={\"name\":\"name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}",
// 			},
// 			true,
// 			"invalid argument \"invalid\" for \"--num-epochs\" flag: strconv.ParseUint: parsing \"invalid\": invalid syntax",
// 			0,
// 		},
// 		{"add reward program proposal - invalid minimum",
// 			[]string{
// 				"add",
// 				"test add reward program",
// 				"description",
// 				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
// 				"--coin=580nhash",
// 				"--reward-program-id=1",
// 				fmt.Sprintf("--dist-address=%s", s.network.Validators[0].Address.String()),
// 				"--epoch-id=day",
// 				"--epoch-offset=100",
// 				"--num-epochs=10",
// 				"--minimum=invalid",
// 				"--maximum=10",
// 				"--eligibility-criteria={\"name\":\"name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}",
// 			},
// 			true,
// 			"invalid argument \"invalid\" for \"--minimum\" flag: strconv.ParseUint: parsing \"invalid\": invalid syntax",
// 			0,
// 		},
// 		{"add reward program proposal - invalid maximum",
// 			[]string{
// 				"add",
// 				"test add reward program",
// 				"description",
// 				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
// 				"--coin=580nhash",
// 				"--reward-program-id=1",
// 				fmt.Sprintf("--dist-address=%s", s.network.Validators[0].Address.String()),
// 				"--epoch-id=day",
// 				"--epoch-offset=100",
// 				"--num-epochs=10",
// 				"--minimum=3",
// 				"--maximum=invalid",
// 				"--eligibility-criteria={\"name\":\"name\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}",
// 			},
// 			true,
// 			"invalid argument \"invalid\" for \"--maximum\" flag: strconv.ParseUint: parsing \"invalid\": invalid syntax",
// 			0,
// 		},
// 		{"add reward program proposal - invalid eligibility criteria",
// 			[]string{
// 				"add",
// 				"test add reward program",
// 				"description",
// 				sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String(),
// 				"--coin=580nhash",
// 				"--reward-program-id=1",
// 				fmt.Sprintf("--dist-address=%s", s.network.Validators[0].Address.String()),
// 				"--epoch-id=day",
// 				"--epoch-offset=100",
// 				"--num-epochs=10",
// 				"--minimum=3",
// 				"--maximum=10",
// 				"--eligibility-criteria=invalid",
// 			},
// 			true,
// 			"unable to parse eligibility criteria : invalid character 'i' looking for beginning of value",
// 			0,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			clientCtx := s.network.Validators[0].ClientCtx

// 			args := []string{
// 				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.network.Validators[0].Address.String()),
// 				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
// 				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
// 				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
// 			}
// 			tc.args = append(tc.args, args...)

// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, rewardcli.GetCmdRewardProgramProposal(), tc.args)
// 			var response sdk.TxResponse
// 			marshalErr := clientCtx.JSONCodec.UnmarshalJSON(out.Bytes(), &response)
// 			if tc.expectErr {
// 				s.Assert().Error(err)
// 				s.Assert().Equal(tc.expectErrMsg, err.Error())
// 			} else {
// 				s.Assert().NoError(err)
// 				s.Assert().NoError(marshalErr, out.String())
// 			}
// 		})
// 	}
// }
