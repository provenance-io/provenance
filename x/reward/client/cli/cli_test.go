package cli_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	provenanceconfig "github.com/provenance-io/provenance/internal/pioconfig"

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
	expiredRewardProgram  types.RewardProgram
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

	s.finishedRewardProgram = types.NewRewardProgram(
		"finished title",
		"finished description",
		2,
		s.accountAddr.String(),
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 100),
		now.Add(-60*60*time.Second),
		60*60,
		3,
		60*60*24,
		s.qualifyingActions,
	)
	s.finishedRewardProgram.ActualProgramEndTime = now
	s.finishedRewardProgram.State = types.RewardProgram_FINISHED

	s.pendingRewardProgram = types.NewRewardProgram(
		"pending title",
		"pending description",
		3,
		s.accountAddr.String(),
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 100),
		now.Add(60*60*time.Second),
		60*60,
		3,
		0,
		s.qualifyingActions,
	)
	s.pendingRewardProgram.State = types.RewardProgram_PENDING

	s.expiredRewardProgram = types.NewRewardProgram(
		"expired title",
		"expired description",
		4,
		s.accountAddr.String(),
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 100),
		now.Add(-60*60*time.Second),
		60*60,
		3,
		0,
		s.qualifyingActions,
	)
	s.expiredRewardProgram.State = types.RewardProgram_EXPIRED

	rewardData := rewardtypes.NewGenesisState(
		[]rewardtypes.RewardProgram{
			s.activeRewardProgram, s.pendingRewardProgram, s.finishedRewardProgram, s.expiredRewardProgram,
		},
		[]rewardtypes.ClaimPeriodRewardDistribution{},
		[]rewardtypes.RewardAccountState{},
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
			[]uint64{1, 2, 3, 4},
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
			[]uint64{2, 4},
		},
		{"query outstanding reward programs",
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

func (s *IntegrationTestSuite) TestTxClaimReward() {
	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
	}{
		{"claim rewards tx - valid",
			[]string{
				"1",
			},
			false,
			"",
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
			out, err := clitestutil.ExecTestCLICmd(clientCtx, rewardcli.GetCmdClaimReward(), tc.args)
			if tc.expectErr {
				s.Assert().Error(err)
				s.Assert().Equal(tc.expectErrMsg, err.Error())
			} else {
				var response sdk.TxResponse
				s.Assert().NoError(err)
				err = s.cfg.Codec.UnmarshalJSON(out.Bytes(), &response)
				marshalErr := clientCtx.JSONCodec.UnmarshalJSON(out.Bytes(), &response)
				s.Assert().NoError(marshalErr)
				s.Assert().Equal(tc.expectedCode, response.Code)
			}
		})
	}
}
