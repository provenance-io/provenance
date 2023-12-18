package cli_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	cmtcli "github.com/cometbft/cometbft/libs/cli"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	rewardcli "github.com/provenance-io/provenance/x/reward/client/cli"
	"github.com/provenance-io/provenance/x/reward/types"
	rewardtypes "github.com/provenance-io/provenance/x/reward/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg        network.Config
	network    *network.Network
	keyring    keyring.Keyring
	keyringDir string

	accountAddr      sdk.AccAddress
	accountKey       *secp256k1.PrivKey
	accountAddresses []sdk.AccAddress

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
	pioconfig.SetProvenanceConfig("", 0)
	s.accountKey = secp256k1.GenPrivKeyFromSecret([]byte("acc2"))
	addr, err := sdk.AccAddressFromHexUnsafe(s.accountKey.PubKey().Address().String())
	s.Require().NoError(err)
	s.accountAddr = addr

	s.cfg = testutil.DefaultTestNetworkConfig()
	genesisState := s.cfg.GenesisState

	s.cfg.NumValidators = 1
	s.GenerateAccountsWithKeyrings(2)

	var genBalances []banktypes.Balance
	for i := range s.accountAddresses {
		genBalances = append(genBalances, banktypes.Balance{Address: s.accountAddresses[i].String(), Coins: sdk.NewCoins(
			sdk.NewInt64Coin("nhash", 100_000_000), sdk.NewInt64Coin(s.cfg.BondDenom, 100_000_000),
		).Sort()})
	}
	genBalances = append(genBalances, banktypes.Balance{Address: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma", Coins: sdk.NewCoins(
		sdk.NewInt64Coin("nhash", 100_000_000), sdk.NewInt64Coin(s.cfg.BondDenom, 100_000_000)).Sort()})
	var bankGenState banktypes.GenesisState
	bankGenState.Params = banktypes.DefaultParams()
	bankGenState.Balances = genBalances
	bankDataBz, err := s.cfg.Codec.MarshalJSON(&bankGenState)
	s.Require().NoError(err)
	genesisState[banktypes.ModuleName] = bankDataBz

	var authData authtypes.GenesisState
	var genAccounts []authtypes.GenesisAccount
	authData.Params = authtypes.DefaultParams()
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.accountAddresses[0], nil, 3, 0))
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.accountAddresses[1], nil, 4, 0))
	accounts, err := authtypes.PackAccounts(genAccounts)
	s.Require().NoError(err)
	authData.Accounts = accounts
	authDataBz, err := s.cfg.Codec.MarshalJSON(&authData)
	s.Require().NoError(err)
	genesisState[authtypes.ModuleName] = authDataBz

	now := time.Now().UTC()
	minimumDelegation := sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().BondDenom, 0)
	maximumDelegation := sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().BondDenom, 10)
	s.qualifyingActions = []types.QualifyingAction{
		{
			Type: &types.QualifyingAction_Delegate{
				Delegate: &types.ActionDelegate{
					MinimumActions:               0,
					MaximumActions:               10,
					MinimumDelegationAmount:      &minimumDelegation,
					MaximumDelegationAmount:      &maximumDelegation,
					MinimumActiveStakePercentile: sdkmath.LegacyNewDecWithPrec(0, 0),
					MaximumActiveStakePercentile: sdkmath.LegacyNewDecWithPrec(1, 0),
				},
			},
		},
	}

	s.activeRewardProgram = types.NewRewardProgram(
		"active title",
		"active description",
		1,
		s.accountAddresses[0].String(),
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 100),
		now,
		60*60,
		3,
		0,
		0,
		s.qualifyingActions,
	)
	s.activeRewardProgram.State = types.RewardProgram_STATE_STARTED

	s.finishedRewardProgram = types.NewRewardProgram(
		"finished title",
		"finished description",
		2,
		s.accountAddresses[0].String(),
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 100),
		now.Add(-60*60*time.Second),
		60*60,
		3,
		0,
		60*60*24,
		s.qualifyingActions,
	)
	s.finishedRewardProgram.ActualProgramEndTime = now
	s.finishedRewardProgram.State = types.RewardProgram_STATE_FINISHED

	s.pendingRewardProgram = types.NewRewardProgram(
		"pending title",
		"pending description",
		3,
		s.accountAddresses[0].String(),
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 100),
		now.Add(60*60*time.Second),
		60*60,
		3,
		0,
		0,
		s.qualifyingActions,
	)
	s.pendingRewardProgram.State = types.RewardProgram_STATE_PENDING

	s.expiredRewardProgram = types.NewRewardProgram(
		"expired title",
		"expired description",
		4,
		s.accountAddresses[0].String(),
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 100),
		now.Add(-60*60*time.Second),
		60*60,
		3,
		0,
		0,
		s.qualifyingActions,
	)
	s.expiredRewardProgram.State = types.RewardProgram_STATE_EXPIRED

	claimPeriodRewardDistributions := make([]rewardtypes.ClaimPeriodRewardDistribution, 101)
	for i := 0; i < 101; i++ {
		claimPeriodRewardDistributions[i] = rewardtypes.NewClaimPeriodRewardDistribution(uint64(i+1), 1, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 10), int64(i), false)
	}

	rewardAccountState := make([]rewardtypes.RewardAccountState, 101)
	for i := 0; i < 101; i++ {
		rewardAccountState[i] = rewardtypes.NewRewardAccountState(1, uint64(i+1), s.accountAddr.String(), 10, []*types.ActionCounter{})
		switch i % 4 {
		case 0:
			rewardAccountState[i].ClaimStatus = rewardtypes.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE
		case 1:
			rewardAccountState[i].ClaimStatus = rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMABLE
		case 2:
			rewardAccountState[i].ClaimStatus = rewardtypes.RewardAccountState_CLAIM_STATUS_CLAIMED
		case 3:
			rewardAccountState[i].ClaimStatus = rewardtypes.RewardAccountState_CLAIM_STATUS_EXPIRED
		}
	}

	rewardData := rewardtypes.NewGenesisState(
		uint64(5),
		[]rewardtypes.RewardProgram{
			s.activeRewardProgram, s.pendingRewardProgram, s.finishedRewardProgram, s.expiredRewardProgram,
		},
		claimPeriodRewardDistributions,
		rewardAccountState,
	)

	rewardDataBz, err := s.cfg.Codec.MarshalJSON(rewardData)
	s.Require().NoError(err)
	genesisState[rewardtypes.ModuleName] = rewardDataBz

	s.cfg.GenesisState = genesisState

	s.cfg.ChainID = antewrapper.SimAppChainID
	s.cfg.TimeoutCommit = 500 * time.Millisecond

	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err, "network.New")

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err, "WaitForHeight")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.Require().NoError(s.network.WaitForNextBlock())
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) GenerateAccountsWithKeyrings(number int) {
	path := hd.CreateHDPath(118, 0, 0).String()
	s.keyringDir = s.T().TempDir()
	kr, err := keyring.New(s.T().Name(), "test", s.keyringDir, nil, s.cfg.Codec)
	s.Require().NoError(err)
	s.keyring = kr
	for i := 0; i < number; i++ {
		keyId := fmt.Sprintf("test_key%v", i)
		info, _, err := kr.NewMnemonic(keyId, keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		s.Require().NoError(err)
		addr, err := info.GetAddress()
		if err != nil {
			panic(err)
		}
		s.accountAddresses = append(s.accountAddresses, addr)
	}
}

func (s *IntegrationTestSuite) TestQueryRewardPrograms() {
	testCases := []struct {
		name         string
		queryTypeArg string
		byId         bool
		expectErrMsg string
		expectedCode uint32
		expectedIds  []uint64
	}{
		{"query all reward programs",
			"all",
			false,
			"",
			0,
			[]uint64{1, 2, 3, 4, 5},
		},
		{"query active reward programs",
			"active",
			false,
			"",
			0,
			[]uint64{1},
		},
		{"query pending reward programs",
			"pending",
			false,
			"",
			0,
			[]uint64{3, 5},
		},
		{"query completed reward programs",
			"completed",
			false,
			"",
			0,
			[]uint64{2, 4},
		},
		{"query outstanding reward programs",
			"outstanding",
			false,
			"",
			0,
			[]uint64{1, 3, 5},
		},
		{"query by id reward programs",
			"2",
			true,
			"",
			0,
			[]uint64{2},
		},
		{"query by id reward programs",
			"99",
			true,
			"failed to query reward program 99: rpc error: code = Unknown desc = rpc error: code = Internal desc = unable to query for reward program by ID: reward program not found: unknown request",
			0,
			[]uint64{2},
		},
		{"query invalid query type",
			"invalid",
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
			out, err := clitestutil.ExecTestCLICmd(clientCtx, rewardcli.GetRewardProgramCmd(), []string{tc.queryTypeArg, fmt.Sprintf("--%s=json", cmtcli.OutputFlag)})
			if len(tc.expectErrMsg) > 0 {
				s.Assert().EqualError(err, tc.expectErrMsg)
			} else if tc.byId {
				var response types.QueryRewardProgramByIDResponse
				s.Assert().NoError(err)
				err = s.cfg.Codec.UnmarshalJSON(out.Bytes(), &response)
				s.Assert().NoError(err)
				s.Assert().Equal(tc.expectedIds[0], response.RewardProgram.Id)
			} else {
				var response types.QueryRewardProgramsResponse
				s.Assert().NoError(err)
				err = s.cfg.Codec.UnmarshalJSON(out.Bytes(), &response)
				s.Assert().NoError(err)
				var rewardProgramIds []uint64
				for _, rp := range response.RewardPrograms {
					rewardProgramIds = append(rewardProgramIds, rp.Id)
				}
				s.Assert().ElementsMatch(tc.expectedIds, rewardProgramIds, "should have all expected reward program ids")
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryClaimPeriodRewardDistributionAll() {
	defaultMaxQueryIds := make([]uint64, 100)
	for i := 0; i < 100; i++ {
		defaultMaxQueryIds[i] = uint64(i + 1)
	}
	testCases := []struct {
		name         string
		args         []string
		byId         bool
		expectErrMsg string
		expectedCode uint32
		expectedIds  []uint64
	}{
		{"query all reward programs get default page size of 100",
			[]string{
				"all",
			},
			false,
			"",
			0,
			defaultMaxQueryIds,
		},
		{"query all reward programs max out to default page size",
			[]string{
				"all",
				"--limit",
				"100",
			},
			false,
			"",
			0,
			defaultMaxQueryIds,
		},
		{"query all reward programs first page with 5 results",
			[]string{
				"all",
				"--limit",
				"5",
			},
			false,
			"",
			0,
			[]uint64{1, 2, 3, 4, 5},
		},
		{"query all reward programs second page with 5 results",
			[]string{
				"all",
				"--limit",
				"5",
				"--page",
				"2",
			},
			false,
			"",
			0,
			[]uint64{6, 7, 8, 9, 10},
		},
		{"query by program id and claim period",
			[]string{
				"1",
				"2",
			},
			true,
			"",
			0,
			[]uint64{1, 2},
		},
		{
			"query without claim period",
			[]string{
				"1",
			},
			true,
			"a reward_program_id and an claim_period_id are required",
			0,
			[]uint64{},
		},
		{
			"query with invalid reward program id format",
			[]string{
				"1",
				"a",
			},
			true,
			"strconv.Atoi: parsing \"a\": invalid syntax",
			0,
			[]uint64{},
		},
		{
			"query with invalid reward program id format",
			[]string{
				"a",
				"1",
			},
			true,
			"strconv.Atoi: parsing \"a\": invalid syntax",
			0,
			[]uint64{},
		},
		{
			"query with invalid reward program id format",
			[]string{
				"100",
				"100",
			},
			true,
			"reward does not exist for reward-id: 100 claim-id 100",
			0,
			[]uint64{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := s.network.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, rewardcli.GetClaimPeriodRewardDistributionCmd(), append(tc.args, []string{fmt.Sprintf("--%s=json", cmtcli.OutputFlag)}...))
			if len(tc.expectErrMsg) > 0 {
				s.Assert().EqualError(err, tc.expectErrMsg)
			} else if tc.byId {
				var response types.QueryClaimPeriodRewardDistributionsByIDResponse
				s.Assert().NoError(err)
				err = s.cfg.Codec.UnmarshalJSON(out.Bytes(), &response)
				s.Assert().NoError(err)
				s.Assert().Equal(tc.expectedIds[0], response.ClaimPeriodRewardDistribution.RewardProgramId)
				s.Assert().Equal(tc.expectedIds[1], response.ClaimPeriodRewardDistribution.ClaimPeriodId)
			} else {
				var response types.QueryClaimPeriodRewardDistributionsResponse
				s.Assert().NoError(err)
				err = s.cfg.Codec.UnmarshalJSON(out.Bytes(), &response)
				s.Assert().NoError(err)
				var claimPeriodRewardDistIds []uint64
				for _, cprd := range response.ClaimPeriodRewardDistributions {
					claimPeriodRewardDistIds = append(claimPeriodRewardDistIds, cprd.ClaimPeriodId)
				}
				s.Assert().ElementsMatch(tc.expectedIds, claimPeriodRewardDistIds, "should have all expected claim period ids")
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdRewardProgramAdd() {
	actions := "{\"qualifying_actions\":[{\"delegate\":{\"minimum_actions\":\"0\",\"maximum_actions\":\"1\",\"minimum_delegation_amount\":{\"denom\":\"nhash\",\"amount\":\"0\"},\"maximum_delegation_amount\":{\"denom\":\"nhash\",\"amount\":\"100\"},\"minimum_active_stake_percentile\":\"0.000000000000000000\",\"maximum_active_stake_percentile\":\"1.000000000000000000\"}}]}"
	soon := time.Now().Add(time.Hour * 24)

	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
		expectedCode uint32
	}{
		{"add reward program tx - valid",
			[]string{
				"test add reward program",
				"description",
				fmt.Sprintf("--total-reward-pool=580%s", s.cfg.BondDenom),
				fmt.Sprintf("--max-reward-by-address=100%s", s.cfg.BondDenom),
				"--claim-periods=52364",
				"--claim-period-days=7",
				fmt.Sprintf("--start-time=%s", soon.Format(time.RFC3339)),
				"--expire-days=14",
				fmt.Sprintf("--qualifying-actions=%s", actions),
			},
			"",
			0,
		},
		{"add reward program tx - invalid total-reward-pool",
			[]string{
				"test add reward program",
				"description",
				"--total-reward-pool=invalid",
				fmt.Sprintf("--max-reward-by-address=100%s", s.cfg.BondDenom),
				"--claim-periods=52",
				"--claim-period-days=10",
				fmt.Sprintf("--start-time=%s", soon.Format(time.RFC3339)),
				"--expire-days=14",
				fmt.Sprintf("--qualifying-actions=%s", actions),
			},
			"invalid decimal coin expression: invalid",
			0,
		},
		{"add reward program tx - invalid max-reward-by-address",
			[]string{
				"test add reward program",
				"description",
				fmt.Sprintf("--total-reward-pool=580%s", s.cfg.BondDenom),
				"--max-reward-by-address=invalid",
				"--claim-period-days=10",
				fmt.Sprintf("--start-time=%s", soon.Format(time.RFC3339)),
				"--expire-days=14",
				fmt.Sprintf("--qualifying-actions=%s", actions),
			},
			"invalid decimal coin expression: invalid",
			0,
		},
		{"add reward program tx - invalid claim period days",
			[]string{
				"test add reward program",
				"description",
				fmt.Sprintf("--total-reward-pool=580%s", s.cfg.BondDenom),
				fmt.Sprintf("--max-reward-by-address=100%s", s.cfg.BondDenom),
				"--claim-periods=52",
				"--claim-period-days=-1",
				fmt.Sprintf("--start-time=%s", soon.Format(time.RFC3339)),
				"--expire-days=14",
				fmt.Sprintf("--qualifying-actions=%s", actions),
			},
			"invalid argument \"-1\" for \"--claim-period-days\" flag: strconv.ParseUint: parsing \"-1\": invalid syntax",
			0,
		},
		{"add reward program tx - invalid expire days",
			[]string{
				"test add reward program",
				"description",
				fmt.Sprintf("--total-reward-pool=580%s", s.cfg.BondDenom),
				fmt.Sprintf("--max-reward-by-address=100%s", s.cfg.BondDenom),
				"--claim-periods=52",
				"--claim-period-days=10",
				fmt.Sprintf("--start-time=%s", soon.Format(time.RFC3339)),
				"--expire-days=-1",
				fmt.Sprintf("--qualifying-actions=%s", actions),
			},
			"invalid argument \"-1\" for \"--expire-days\" flag: strconv.ParseUint: parsing \"-1\": invalid syntax",
			0,
		},
		{"add reward program tx - invalid reward period days",
			[]string{
				"test add reward program",
				"description",
				fmt.Sprintf("--total-reward-pool=580%s", s.cfg.BondDenom),
				fmt.Sprintf("--max-reward-by-address=100%s", s.cfg.BondDenom),
				"--claim-periods=-52",
				"--claim-period-days=10",
				fmt.Sprintf("--start-time=%s", soon.Format(time.RFC3339)),
				"--expire-days=1",
				fmt.Sprintf("--qualifying-actions=%s", actions),
			},
			"invalid argument \"-52\" for \"--claim-periods\" flag: strconv.ParseUint: parsing \"-52\": invalid syntax",
			0,
		},
		{"add reward program tx - invalid start time",
			[]string{
				"test add reward program",
				"description",
				fmt.Sprintf("--total-reward-pool=580%s", s.cfg.BondDenom),
				fmt.Sprintf("--max-reward-by-address=100%s", s.cfg.BondDenom),
				"--claim-periods=52",
				"--claim-period-days=10",
				"--start-time=invalid",
				"--expire-days=14",
				fmt.Sprintf("--qualifying-actions=%s", actions),
			},
			"unable to parse time (invalid) required format is RFC3339 (2006-01-02T15:04:05Z07:00) , parsing time \"invalid\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"invalid\" as \"2006\"",
			0,
		},
		{"add reward program tx - invalid ec",
			[]string{
				"test add reward program",
				"description",
				fmt.Sprintf("--total-reward-pool=580%s", s.cfg.BondDenom),
				fmt.Sprintf("--max-reward-by-address=100%s", s.cfg.BondDenom),
				"--claim-periods=52",
				"--claim-period-days=10",
				fmt.Sprintf("--start-time=%s", soon.Format(time.RFC3339)),
				"--expire-days=14",
				fmt.Sprintf("--qualifying-actions=%s", "actions"),
			},
			"invalid character 'a' looking for beginning of value",
			0,
		},
		{"add reward program tx - invalid type for claim periods",
			[]string{
				"test add reward program",
				"description",
				fmt.Sprintf("--total-reward-pool=580%s", s.cfg.BondDenom),
				fmt.Sprintf("--max-reward-by-address=100%s", s.cfg.BondDenom),
				"--claim-periods=abc",
				"--claim-period-days=10",
				fmt.Sprintf("--start-time=%s", soon.Format(time.RFC3339)),
				"--expire-days=14",
				fmt.Sprintf("--qualifying-actions=%s", "actions"),
			},
			"invalid argument \"abc\" for \"--claim-periods\" flag: strconv.ParseUint: parsing \"abc\": invalid syntax",
			0,
		},
		{"add reward program tx - invalid type for claim period days",
			[]string{
				"test add reward program",
				"description",
				fmt.Sprintf("--total-reward-pool=580%s", s.cfg.BondDenom),
				fmt.Sprintf("--max-reward-by-address=100%s", s.cfg.BondDenom),
				"--claim-periods=52",
				"--claim-period-days=abc",
				fmt.Sprintf("--start-time=%s", soon.Format(time.RFC3339)),
				"--expire-days=14",
				fmt.Sprintf("--qualifying-actions=%s", "actions"),
			},
			"invalid argument \"abc\" for \"--claim-period-days\" flag: strconv.ParseUint: parsing \"abc\": invalid syntax",
			0,
		},
		{"add reward program tx - invalid type for claim period days",
			[]string{
				"test add reward program",
				"description",
				fmt.Sprintf("--total-reward-pool=580%s", s.cfg.BondDenom),
				fmt.Sprintf("--max-reward-by-address=100%s", s.cfg.BondDenom),
				"--claim-periods=52",
				"--claim-period-days=10",
				fmt.Sprintf("--start-time=%s", soon.Format(time.RFC3339)),
				"--expire-days=abc",
				fmt.Sprintf("--qualifying-actions=%s", "actions"),
			},
			"invalid argument \"abc\" for \"--expire-days\" flag: strconv.ParseUint: parsing \"abc\": invalid syntax",
			0,
		},
		{"add reward program tx - invalid type for claim period days",
			[]string{
				"test add reward program",
				"description",
				fmt.Sprintf("--total-reward-pool=580%s", s.cfg.BondDenom),
				fmt.Sprintf("--max-reward-by-address=100%s", s.cfg.BondDenom),
				"--claim-periods=52",
				"--claim-period-days=10",
				fmt.Sprintf("--start-time=%s", soon.Format(time.RFC3339)),
				"--expire-days=14",
				"--max-rollover-periods=abc",
				fmt.Sprintf("--qualifying-actions=%s", "actions"),
			},
			"invalid argument \"abc\" for \"--max-rollover-periods\" flag: strconv.ParseUint: parsing \"abc\": invalid syntax",
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync), // TODO[1760]: broadcast
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			}
			tc.args = append(tc.args, args...)

			out, err := clitestutil.ExecTestCLICmd(clientCtx, rewardcli.GetCmdRewardProgramAdd(), append(tc.args, []string{fmt.Sprintf("--%s=json", cmtcli.OutputFlag)}...))
			var response sdk.TxResponse
			marshalErr := clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response)
			if len(tc.expectErrMsg) > 0 {
				s.Assert().EqualError(err, tc.expectErrMsg)
			} else {
				s.Assert().NoError(err)
				s.Assert().NoError(marshalErr, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxClaimReward() {
	testCases := []struct {
		name           string
		claimRewardArg string
		expectErrMsg   string
		expectedCode   uint32
	}{
		{"claim rewards tx - valid",
			"1",
			"",
			0,
		},
		{"claim rewards tx - all",
			"all",
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
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync), // TODO[1760]: broadcast
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			}
			args = append(args, tc.claimRewardArg)
			out, err := clitestutil.ExecTestCLICmd(clientCtx, rewardcli.GetCmdClaimReward(), args)
			if len(tc.expectErrMsg) > 0 {
				s.Assert().EqualError(err, tc.expectErrMsg)
			} else {
				var response sdk.TxResponse
				s.Assert().NoError(err)
				err = s.cfg.Codec.UnmarshalJSON(out.Bytes(), &response)
				marshalErr := clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response)
				s.Assert().NoError(marshalErr)
				s.Assert().Equal(tc.expectedCode, response.Code)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxEndRewardProgram() {
	testCases := []struct {
		name               string
		endRewardProgramId string
		expectErrMsg       string
		expectedCode       uint32
		signer             string
	}{
		{
			name:               "end reward program - valid",
			endRewardProgramId: "1",
			expectErrMsg:       "",
			expectedCode:       0,
			signer:             s.accountAddresses[0].String(),
		},
		{
			name:               "end reward program - invalid id",
			endRewardProgramId: "999",
			expectErrMsg:       "",
			expectedCode:       3,
			signer:             s.accountAddresses[0].String(),
		},
		{
			name:               "end reward program - invalid state",
			endRewardProgramId: "2",
			expectErrMsg:       "",
			expectedCode:       5,
			signer:             s.accountAddresses[0].String(),
		},
		{
			name:               "end reward program - not authorized",
			endRewardProgramId: "1",
			expectErrMsg:       "",
			expectedCode:       4,
			signer:             s.accountAddresses[1].String(),
		},
		{
			name:               "end reward program - invalid id format",
			endRewardProgramId: "abc",
			expectErrMsg:       "invalid argument : abc",
			expectedCode:       0,
			signer:             s.accountAddresses[0].String(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := s.network.Validators[0].ClientCtx.WithKeyringDir(s.keyringDir).WithKeyring(s.keyring)
			args := []string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, tc.signer),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync), // TODO[1760]: broadcast
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			}
			args = append(args, tc.endRewardProgramId)
			out, err := clitestutil.ExecTestCLICmd(clientCtx, rewardcli.GetCmdEndRewardProgram(), args)
			if len(tc.expectErrMsg) > 0 {
				s.Assert().EqualError(err, tc.expectErrMsg)
			} else {
				var response sdk.TxResponse
				s.Assert().NoError(err)
				err = s.cfg.Codec.UnmarshalJSON(out.Bytes(), &response)
				marshalErr := clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response)
				s.Assert().NoError(marshalErr)
				s.Assert().Equal(tc.expectedCode, response.Code)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryAllRewardsPerAddress() {
	testCases := []struct {
		name           string
		addressArg     string
		stateArg       string
		byId           bool
		expectErrMsg   string
		expectedCode   uint32
		expectedIds    []uint64
		expectedLength int64
	}{
		{
			name:           "query all reward by address",
			addressArg:     s.accountAddr.String(),
			stateArg:       "all",
			byId:           false,
			expectErrMsg:   "",
			expectedCode:   0,
			expectedIds:    []uint64{1, 2, 3, 4},
			expectedLength: 100,
		},
		{
			name:           "query unclaimable reward by address",
			addressArg:     s.accountAddr.String(),
			stateArg:       "unclaimable",
			byId:           false,
			expectErrMsg:   "",
			expectedCode:   0,
			expectedIds:    []uint64{1, 5},
			expectedLength: 26,
		},
		{
			name:           "query claimable reward by address",
			addressArg:     s.accountAddr.String(),
			stateArg:       "claimable",
			byId:           false,
			expectErrMsg:   "",
			expectedCode:   0,
			expectedIds:    []uint64{2, 6},
			expectedLength: 25,
		},
		{
			name:           "query claimed reward by address",
			addressArg:     s.accountAddr.String(),
			stateArg:       "claimed",
			byId:           false,
			expectErrMsg:   "",
			expectedCode:   0,
			expectedIds:    []uint64{3, 7},
			expectedLength: 25,
		},
		{
			name:           "query expired reward by address",
			addressArg:     s.accountAddr.String(),
			stateArg:       "expired",
			byId:           false,
			expectErrMsg:   "",
			expectedCode:   0,
			expectedIds:    []uint64{4, 8},
			expectedLength: 25,
		},
		{
			name:           "query reward by address",
			addressArg:     s.accountAddr.String(),
			stateArg:       "invalid",
			byId:           false,
			expectErrMsg:   "failed to query reward distributions. invalid is not a valid query param",
			expectedCode:   0,
			expectedIds:    []uint64{},
			expectedLength: 0,
		},
		{
			name:           "query reward by invalid address",
			addressArg:     "invalid address",
			stateArg:       "expired",
			byId:           false,
			expectErrMsg:   "failed to query reward distributions: rpc error: code = Unknown desc = decoding bech32 failed: invalid character in string: ' ': invalid address: unknown request",
			expectedCode:   0,
			expectedIds:    []uint64{},
			expectedLength: 0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := s.network.Validators[0].ClientCtx
			args := []string{tc.addressArg, tc.stateArg, fmt.Sprintf("--%s=json", cmtcli.OutputFlag)}
			out, err := clitestutil.ExecTestCLICmd(clientCtx, rewardcli.GetRewardsByAddressCmd(), args)
			if len(tc.expectErrMsg) > 0 {
				s.Assert().EqualError(err, tc.expectErrMsg)
			} else if tc.byId {
				var response types.QueryRewardDistributionsByAddressResponse
				s.Assert().NoError(err)
				err = s.cfg.Codec.UnmarshalJSON(out.Bytes(), &response)
				s.Assert().NoError(err)
			} else {
				var response types.QueryRewardDistributionsByAddressResponse
				s.Assert().NoError(err)
				err = s.cfg.Codec.UnmarshalJSON(out.Bytes(), &response)
				s.Assert().NoError(err)
				var actualClaimIds []uint64
				for _, ras := range response.RewardAccountState {
					if ras.RewardProgramId == 1 {
						actualClaimIds = append(actualClaimIds, ras.ClaimId)
					}
				}
				for _, eId := range tc.expectedIds {
					s.Assert().Contains(actualClaimIds, eId, fmt.Sprintf("missing claim id %d for reward id 1", eId))
				}
				s.Assert().Equal(int(tc.expectedLength), len(response.RewardAccountState), "length of results does not match")
			}
		})
	}
}
