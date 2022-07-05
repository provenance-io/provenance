package cli_test

import (
	"fmt"
	provenanceconfig "github.com/provenance-io/provenance/internal/config"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/testutil"
	rewardcli "github.com/provenance-io/provenance/x/reward/client/cli"
	"github.com/provenance-io/provenance/x/reward/types"
	rewardtypes "github.com/provenance-io/provenance/x/reward/types"

	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network

	accountAddr sdk.AccAddress
	accountKey  *secp256k1.PrivKey

	activeRewardProgram   types.RewardProgram
	pendingRewardProgram  types.RewardProgram
	finishedRewardProgram types.RewardProgram
	qualifyingActions     []types.QualifyingAction
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
	minimumDelegation := sdk.NewInt64Coin(provenanceconfig.DefaultBondDenom, 0)
	maximumDelegation := sdk.NewInt64Coin(provenanceconfig.DefaultBondDenom, 10)
	s.qualifyingActions = []types.QualifyingAction{
		{
			Type: &types.QualifyingAction_Delegate{
				Delegate: &types.ActionDelegate{
					MinimumActions:               0,
					MaximumActions:               10,
					MinimumDelegationAmount:      &minimumDelegation,
					MaximumDelegationAmount:      &maximumDelegation,
					MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
					MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
				},
			},
		},
	}

	s.activeRewardProgram = types.NewRewardProgram(
		"active title",
		"active description",
		1,
		s.accountAddr.String(),
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 100),
		now,
		60*60,
		3,
		0,
		s.qualifyingActions,
	)
	s.activeRewardProgram.State = types.RewardProgram_STARTED
	s.activeRewardProgram.Id = 1

	s.finishedRewardProgram = types.NewRewardProgram(
		"finished title",
		"finished description",
		1,
		s.accountAddr.String(),
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 100),
		now.Add(-60*60*time.Second),
		60*60,
		3,
		1,
		s.qualifyingActions,
	)
	s.finishedRewardProgram.Id = 2
	s.finishedRewardProgram.State = types.RewardProgram_FINISHED

	s.pendingRewardProgram = types.NewRewardProgram(
		"pending title",
		"pending description",
		1,
		s.accountAddr.String(),
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 100),
		now.Add(60*60*time.Second),
		60*60,
		3,
		0,
		s.qualifyingActions,
	)
	s.pendingRewardProgram.Id = 3
	s.pendingRewardProgram.State = types.RewardProgram_PENDING

	rewardData := rewardtypes.NewGenesisState(
		rewardtypes.DefaultStartingRewardProgramID,
		[]rewardtypes.RewardProgram{
			s.activeRewardProgram, s.pendingRewardProgram, s.finishedRewardProgram,
		},
		[]rewardtypes.ClaimPeriodRewardDistribution{},
		rewardtypes.ActionDelegate{},
		rewardtypes.ActionTransferDelegations{},
	)

	rewardDataBz, err := s.cfg.Codec.MarshalJSON(rewardData)
	s.Require().NoError(err)
	genesisState[rewardtypes.ModuleName] = rewardDataBz

	s.cfg.GenesisState = genesisState

	s.cfg.ChainID = antewrapper.SimAppChainID

	s.network = network.New(s.T(), s.cfg)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.network.WaitForNextBlock()
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestQueryRewardPrograms() {
	testCases := []struct {
		name         string
		args         []string
		byId         bool
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
		expectedIds  []uint64
	}{
		{"query all reward programs",
			[]string{
				"all",
			},
			false,
			false,
			"",
			0,
			[]uint64{1, 2, 3},
		},
		{"query active reward programs",
			[]string{
				"active",
			},
			false,
			false,
			"",
			0,
			[]uint64{1},
		},
		{"query pending reward programs",
			[]string{
				"pending",
			},
			false,
			false,
			"",
			0,
			[]uint64{3},
		},
		{"query completed reward programs",
			[]string{
				"completed",
			},
			false,
			false,
			"",
			0,
			[]uint64{2},
		},
		{"query outstnding reward programs",
			[]string{
				"outstanding",
			},
			false,
			false,
			"",
			0,
			[]uint64{1, 3},
		},
		{"query by id reward programs",
			[]string{
				"2",
			},
			true,
			false,
			"",
			0,
			[]uint64{2},
		},
		{"query invalid query type",
			[]string{
				"invalid",
			},
			true,
			true,
			"invalid argument arg : invalid",
			0,
			[]uint64{},
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
			} else if tc.byId {
				var response types.RewardProgramByIDResponse
				s.Assert().NoError(err)
				err = s.cfg.Codec.UnmarshalJSON(out.Bytes(), &response)
				s.Assert().NoError(err)
				s.Assert().Equal(tc.expectedIds[0], response.RewardProgram.Id)
			} else {
				var response types.RewardProgramsResponse
				s.Assert().NoError(err)
				err = s.cfg.Codec.UnmarshalJSON(out.Bytes(), &response)
				s.Assert().NoError(err)
				s.Assert().Equal(len(tc.expectedIds), len(response.RewardPrograms))
				//s.Assert().Equal(tc.expectedOutput, response)
				for _, expectedId := range tc.expectedIds {
					s.Assert().True(containsId(response.RewardPrograms, expectedId))
				}
			}
		})
	}
}

func containsId(rewardPrograms []types.RewardProgram, id uint64) bool {
	for _, rewardProgram := range rewardPrograms {
		if rewardProgram.Id == id {
			return true
		}
	}
	return false
}

// func (s *IntegrationTestSuite) TestQueryRewardClaims() {
// 	rewardClaimsJson := fmt.Sprintf(`{"reward_claims":[{"address":"%s","shares_per_epoch_per_reward":[{"reward_program_id":"1","total_shares":"1","ephemeral_action_count":"1","latest_recorded_epoch":"1","claimed":false,"expired":false,"total_reward_claimed":{"denom":"","amount":"0"}}]},{"address":"cosmos1p3sl9tll0ygj3flwt5r2w0n6fx9p5ngq2tu6mq","shares_per_epoch_per_reward":[{"reward_program_id":"2","total_shares":"0","ephemeral_action_count":"0","latest_recorded_epoch":"0","claimed":false,"expired":false,"total_reward_claimed":{"denom":"nhash","amount":"0"}}]},{"address":"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h","shares_per_epoch_per_reward":[{"reward_program_id":"3","total_shares":"0","ephemeral_action_count":"0","latest_recorded_epoch":"0","claimed":false,"expired":false,"total_reward_claimed":{"denom":"nhash","amount":"0"}}]}]}`, s.network.Validators[0].Address.String())
// 	testCases := []struct {
// 		name           string
// 		args           []string
// 		expectErr      bool
// 		expectErrMsg   string
// 		expectedCode   uint32
// 		expectedOutput string
// 		compareLength  bool
// 	}{
// 		{"query all reward claims",
// 			[]string{
// 				"all",
// 			},
// 			false,
// 			"",
// 			0,
// 			"",
// 			true,
// 		},
// 		{"query existing reward claim by address",
// 			[]string{
// 				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
// 			},
// 			false,
// 			"",
// 			0,
// 			"{\"reward_claim\":{\"address\":\"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h\",\"shares_per_epoch_per_reward\":[{\"reward_program_id\":\"3\",\"total_shares\":\"0\",\"ephemeral_action_count\":\"0\",\"latest_recorded_epoch\":\"0\",\"claimed\":false,\"expired\":false,\"total_reward_claimed\":{\"denom\":\"jackthecat\",\"amount\":\"0\"}}]}}",
// 			false,
// 		},
// 		{"query non-existing reward claim by address",
// 			[]string{
// 				"failure",
// 			},
// 			true,
// 			"reward claim failure does not exist",
// 			0,
// 			"",
// 			false,
// 		},
// 		{"query non-existing reward claim by empty address",
// 			[]string{
// 				"",
// 			},
// 			true,
// 			"reward claim  does not exist",
// 			0,
// 			"",
// 			false,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			clientCtx := s.network.Validators[0].ClientCtx
// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, rewardcli.GetRewardClaimCmd(), tc.args)
// 			if tc.expectErr {
// 				s.Assert().Error(err)
// 				s.Assert().Equal(tc.expectErrMsg, err.Error())
// 			} else {
// 				s.Assert().NoError(err)
// 				if tc.compareLength {
// 					actualResponse := rewardtypes.RewardClaimsResponse{}
// 					expectedResponse := rewardtypes.RewardClaimsResponse{}
// 					json.Unmarshal([]byte(strings.TrimSpace(out.String())), &actualResponse)
// 					json.Unmarshal([]byte(strings.TrimSpace(rewardClaimsJson)), &expectedResponse)
// 					s.Assert().Equal(len(expectedResponse.GetRewardClaims()), len(actualResponse.GetRewardClaims()), "should have all reward claims")
// 				} else {
// 					s.Assert().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
// 				}
// 			}
// 		})
// 	}
// }

// func (s *IntegrationTestSuite) TestQueryEpochDistributionReward() {
// 	testCases := []struct {
// 		name           string
// 		args           []string
// 		expectErr      bool
// 		expectErrMsg   string
// 		expectedCode   uint32
// 		expectedOutput string
// 	}{
// 		{"query all epoch distribution rewards",
// 			[]string{
// 				"all",
// 			},
// 			false,
// 			"",
// 			0,
// 			"{\"epoch_reward_distribution\":[{\"epoch_id\":\"day\",\"reward_program_id\":\"1\",\"total_rewards_pool\":{\"denom\":\"jackthecat\",\"amount\":\"100\"},\"total_shares\":\"5\",\"epoch_ended\":false},{\"epoch_id\":\"day\",\"reward_program_id\":\"2\",\"total_rewards_pool\":{\"denom\":\"jackthecat\",\"amount\":\"100\"},\"total_shares\":\"3\",\"epoch_ended\":false},{\"epoch_id\":\"minute\",\"reward_program_id\":\"1\",\"total_rewards_pool\":{\"denom\":\"jackthecat\",\"amount\":\"1\"},\"total_shares\":\"1\",\"epoch_ended\":false},{\"epoch_id\":\"month\",\"reward_program_id\":\"1\",\"total_rewards_pool\":{\"denom\":\"jackthecat\",\"amount\":\"100\"},\"total_shares\":\"10\",\"epoch_ended\":false}]}",
// 		},
// 		{"query existing epoch reward distribution by valid ids",
// 			[]string{
// 				"1",
// 				"day",
// 			},
// 			false,
// 			"",
// 			0,
// 			"{\"epoch_reward_distribution\":{\"epoch_id\":\"day\",\"reward_program_id\":\"1\",\"total_rewards_pool\":{\"denom\":\"jackthecat\",\"amount\":\"100\"},\"total_shares\":\"5\",\"epoch_ended\":false}}",
// 		},
// 		{"query epoch reward distribution by invalid reward program id",
// 			[]string{
// 				"10",
// 				"day",
// 			},
// 			true,
// 			"epoch reward does not exist for reward-id: 10 epoch-id day",
// 			0,
// 			"",
// 		},
// 		{"query epoch reward distribution by invalid epoch id",
// 			[]string{
// 				"1",
// 				"blah",
// 			},
// 			true,
// 			"epoch reward does not exist for reward-id: 1 epoch-id blah",
// 			0,
// 			"",
// 		},
// 		{"query epoch reward distribution by invalid reward program id and epoch id",
// 			[]string{
// 				"10",
// 				"blah",
// 			},
// 			true,
// 			"epoch reward does not exist for reward-id: 10 epoch-id blah",
// 			0,
// 			"",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		fmt.Printf("Address: %s\n", s.network.Validators[0].Address.String())
// 		s.Run(tc.name, func() {
// 			clientCtx := s.network.Validators[0].ClientCtx
// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, rewardcli.GetClaimPeriodRewardDistributionCmd(), tc.args)
// 			if tc.expectErr {
// 				s.Assert().Error(err)
// 				s.Assert().Equal(tc.expectErrMsg, err.Error())
// 			} else {
// 				s.Assert().NoError(err)
// 				s.Assert().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
// 			}
// 		})
// 	}
// }

// func (s *IntegrationTestSuite) TestQueryEligibilityCriteria() {
// 	testCases := []struct {
// 		name           string
// 		args           []string
// 		expectErr      bool
// 		expectErrMsg   string
// 		expectedCode   uint32
// 		expectedOutput string
// 	}{
// 		{"query all eligibility criteria",
// 			[]string{
// 				"all",
// 			},
// 			false,
// 			"",
// 			0,
// 			"{\"eligibility_criteria\":[{\"name\":\"test1\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}},{\"name\":\"test2\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}},{\"name\":\"test3\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}]}",
// 		},
// 		{"query existing eligibility criteria by valid name",
// 			[]string{
// 				"test1",
// 			},
// 			false,
// 			"",
// 			0,
// 			"{\"eligibility_criteria\":{\"name\":\"test1\",\"action\":{\"@type\":\"/provenance.reward.v1.ActionDelegate\"}}}",
// 		},
// 		{"query eligibility criteria by invalid name",
// 			[]string{
// 				"blah",
// 			},
// 			true,
// 			"eligibility criteria does not exist for name: blah",
// 			0,
// 			"",
// 		},
// 		{"query eligibility criteria with empty name",
// 			[]string{
// 				"",
// 			},
// 			true,
// 			"eligibility criteria does not exist for name: ",
// 			0,
// 			"",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		fmt.Printf("Address: %s\n", s.network.Validators[0].Address.String())
// 		s.Run(tc.name, func() {
// 			clientCtx := s.network.Validators[0].ClientCtx
// 			out, err := clitestutil.ExecTestCLICmd(clientCtx, rewardcli.GetEligibilityCriteriaCmd(), tc.args)
// 			if tc.expectErr {
// 				s.Assert().Error(err)
// 				s.Assert().Equal(tc.expectErrMsg, err.Error())
// 			} else {
// 				s.Assert().NoError(err)
// 				s.Assert().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
// 			}
// 		})
// 	}
//  }

func (s *IntegrationTestSuite) TestGetCmdRewardProgramAdd() {
	actions := "{\"qualifying_actions\":[{\"delegate\":{\"minimum_actions\":\"0\",\"maximum_actions\":\"0\",\"minimum_delegation_amount\":{\"denom\":\"nhash\",\"amount\":\"0\"},\"maximum_delegation_amount\":{\"denom\":\"nhash\",\"amount\":\"100\"},\"minimum_active_stake_percentile\":\"0.000000000000000000\",\"maximum_active_stake_percentile\":\"1.000000000000000000\"}}]}"
	soon := time.Now().Add(time.Hour * 24)
	date := strings.Split(soon.String(), " ")

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
	}{
		{"add reward program tx - valid",
			[]string{
				"add-reward-program",
				"test add reward program",
				"description",
				fmt.Sprintf("--total-reward-pool=580%s", s.cfg.BondDenom),
				fmt.Sprintf("--max-reward-by-address=100%s", s.cfg.BondDenom),
				"--reward-period-days=364",
				"--claim-period-days=7",
				fmt.Sprintf("--start-time=%s", date[0]),
				"--expire-days=14",
				fmt.Sprintf("--qualifying-actions=%s", actions),
			},
			false,
			"",
			0,
		},
		{"add reward program tx - invalid total-reward-pool",
			[]string{
				"add-reward-program",
				"test add reward program",
				"description",
				"--total-reward-pool=invalid",
				fmt.Sprintf("--max-reward-by-address=100%s", s.cfg.BondDenom),
				"--reward-period-days=365",
				"--claim-period-days=10",
				"--start-time=2022-05-10",
				"--expire-days=14",
				fmt.Sprintf("--qualifying-actions=%s", actions),
			},
			true,
			"invalid decimal coin expression: invalid",
			0,
		},
		{"add reward program tx - invalid max-reward-by-address",
			[]string{
				"add-reward-program",
				"test add reward program",
				"description",
				fmt.Sprintf("--total-reward-pool=580%s", s.cfg.BondDenom),
				"--max-reward-by-address=invalid",
				"--claim-period-days=10",
				"--start-time=2022-05-10",
				"--expire-days=14",
				fmt.Sprintf("--qualifying-actions=%s", actions),
			},
			true,
			"invalid decimal coin expression: invalid",
			0,
		},
		{"add reward program tx - invalid claim period days",
			[]string{
				"add-reward-program",
				"test add reward program",
				"description",
				fmt.Sprintf("--total-reward-pool=580%s", s.cfg.BondDenom),
				fmt.Sprintf("--max-reward-by-address=100%s", s.cfg.BondDenom),
				"--reward-period-days=365",
				"--claim-period-days=-1",
				"--start-time=2022-05-10",
				"--expire-days=14",
				fmt.Sprintf("--qualifying-actions=%s", actions),
			},
			true,
			"invalid argument \"-1\" for \"--claim-period-days\" flag: strconv.ParseUint: parsing \"-1\": invalid syntax",
			0,
		},
		{"add reward program tx - invalid expire days",
			[]string{
				"add-reward-program",
				"test add reward program",
				"description",
				fmt.Sprintf("--total-reward-pool=580%s", s.cfg.BondDenom),
				fmt.Sprintf("--max-reward-by-address=100%s", s.cfg.BondDenom),
				"--reward-period-days=365",
				"--claim-period-days=10",
				"--start-time=2022-05-10",
				"--expire-days=-1",
				fmt.Sprintf("--qualifying-actions=%s", actions),
			},
			true,
			"invalid argument \"-1\" for \"--expire-days\" flag: strconv.ParseUint: parsing \"-1\": invalid syntax",
			0,
		},
		{"add reward program tx - invalid reward period days",
			[]string{
				"add-reward-program",
				"test add reward program",
				"description",
				fmt.Sprintf("--total-reward-pool=580%s", s.cfg.BondDenom),
				fmt.Sprintf("--max-reward-by-address=100%s", s.cfg.BondDenom),
				"--reward-period-days=-365",
				"--claim-period-days=10",
				"--start-time=2022-05-10",
				"--expire-days=1",
				fmt.Sprintf("--qualifying-actions=%s", actions),
			},
			true,
			"invalid argument \"-365\" for \"--reward-period-days\" flag: strconv.ParseUint: parsing \"-365\": invalid syntax",
			0,
		},
		{"add reward program tx - invalid start time",
			[]string{
				"add-reward-program",
				"test add reward program",
				"description",
				fmt.Sprintf("--total-reward-pool=580%s", s.cfg.BondDenom),
				fmt.Sprintf("--max-reward-by-address=100%s", s.cfg.BondDenom),
				"--reward-period-days=365",
				"--claim-period-days=10",
				"--start-time=invalid",
				"--expire-days=14",
				fmt.Sprintf("--qualifying-actions=%s", actions),
			},
			true,
			"error parsing start date must be of format YYYY-MM-dd: invalid",
			0,
		},
		{"add reward program tx - invalid ec",
			[]string{
				"add-reward-program",
				"test add reward program",
				"description",
				fmt.Sprintf("--total-reward-pool=580%s", s.cfg.BondDenom),
				fmt.Sprintf("--max-reward-by-address=100%s", s.cfg.BondDenom),
				"--reward-period-days=365",
				"--claim-period-days=10",
				"--start-time=2022-05-10",
				"--expire-days=14",
				fmt.Sprintf("--qualifying-actions=%s", "actions"),
			},
			true,
			"invalid character 'a' looking for beginning of value",
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

			out, err := clitestutil.ExecTestCLICmd(clientCtx, rewardcli.GetCmdRewardProgramAdd(), tc.args)
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
