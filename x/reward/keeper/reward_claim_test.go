package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

var (
	minDelegation = sdk.NewInt64Coin("nhash", 4)
	maxDelegation = sdk.NewInt64Coin("nhash", 40)
)

func (s *KeeperTestSuite) TestClaimRewards() {
	time := s.ctx.BlockTime()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv",
		sdk.NewInt64Coin("nhash", 1000),
		sdk.NewInt64Coin("nhash", 100),
		time,
		10,
		3,
		0,
		uint64(time.Day()),
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Vote{
					Vote: &types.ActionVote{
						MinimumActions:          0,
						MaximumActions:          1,
						MinimumDelegationAmount: minDelegation,
					},
				},
			},
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               1,
						MinimumDelegationAmount:      &minDelegation,
						MaximumDelegationAmount:      &maxDelegation,
						MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
					},
				},
			},
		},
	)
	rewardProgram.State = types.RewardProgram_STATE_FINISHED
	rewardProgram.CurrentClaimPeriod = rewardProgram.GetClaimPeriods()
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

	for i := 1; i <= int(rewardProgram.GetClaimPeriods()); i++ {
		state := types.NewRewardAccountState(rewardProgram.GetId(), uint64(i), "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 1, map[string]uint64{})
		state.ClaimStatus = types.RewardAccountState_CLAIM_STATUS_CLAIMABLE
		s.app.RewardKeeper.SetRewardAccountState(s.ctx, state)
		distribution := types.NewClaimPeriodRewardDistribution(uint64(i), rewardProgram.GetId(), sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, true)
		s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, distribution)
	}

	details, reward, err := s.app.RewardKeeper.ClaimRewards(s.ctx, rewardProgram.GetId(), "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv")
	s.Assert().NoError(err, "should throw no error")

	rewardProgram, err = s.app.RewardKeeper.GetRewardProgram(s.ctx, rewardProgram.GetId())
	s.Assert().NoError(err, "should throw no error")
	s.Assert().Equal(3, len(details), "should have rewards from every period")
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 300), reward, "should total up the rewards from the periods")
}

func (s *KeeperTestSuite) TestClaimRewardsHandlesInvalidProgram() {
	time := s.ctx.BlockTime()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv",
		sdk.NewInt64Coin("nhash", 1000),
		sdk.NewInt64Coin("nhash", 100),
		time,
		10,
		5,
		0,
		uint64(time.Day()),
		[]types.QualifyingAction{},
	)
	rewardProgram.State = types.RewardProgram_STATE_FINISHED
	rewardProgram.CurrentClaimPeriod = 5

	details, reward, err := s.app.RewardKeeper.ClaimRewards(s.ctx, rewardProgram.GetId(), "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv")
	s.Assert().Nil(details, "should have no reward details")
	s.Assert().Equal(sdk.Coin{}, reward, "should have no reward")
	s.Assert().Error(err, "should throw error")
}

func (s *KeeperTestSuite) TestClaimRewardsHandlesExpiredProgram() {
	time := s.ctx.BlockTime()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv",
		sdk.NewInt64Coin("nhash", 1000),
		sdk.NewInt64Coin("nhash", 100),
		time,
		10,
		5,
		0,
		uint64(time.Day()),
		[]types.QualifyingAction{},
	)
	rewardProgram.State = types.RewardProgram_STATE_EXPIRED
	rewardProgram.CurrentClaimPeriod = 5
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

	details, reward, err := s.app.RewardKeeper.ClaimRewards(s.ctx, rewardProgram.GetId(), "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv")
	s.Assert().Nil(details, "should have no reward details")
	s.Assert().Equal(sdk.Coin{}, reward, "should have no reward")
	s.Assert().Error(err, "should throw error")
}

func (s *KeeperTestSuite) TestRefundRewardClaims() {
	time := s.ctx.BlockTime()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv",
		sdk.NewInt64Coin("nhash", 1000),
		sdk.NewInt64Coin("nhash", 100),
		time,
		10,
		5,
		0,
		uint64(time.Day()),
		[]types.QualifyingAction{},
	)
	rewardProgram.RemainingPoolBalance = sdk.NewInt64Coin("nhash", 0)
	rewardProgram.ClaimedAmount = sdk.NewInt64Coin("nhash", 0)

	addr, _ := sdk.AccAddressFromBech32("cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv")
	beforeBalance := s.app.BankKeeper.GetBalance(s.ctx, addr, "nhash")
	err := s.app.RewardKeeper.RefundRewardClaims(s.ctx, rewardProgram)
	afterBalance := s.app.BankKeeper.GetBalance(s.ctx, addr, "nhash")

	s.Assert().NoError(err, "no error should be thrown")
	s.Assert().Equal(beforeBalance.Add(rewardProgram.TotalRewardPool), afterBalance, "unclaimed balance should be refunded")
}

func (s *KeeperTestSuite) TestRefundRewardClaimsEmpty() {
	time := s.ctx.BlockTime()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv",
		sdk.NewInt64Coin("nhash", 1000),
		sdk.NewInt64Coin("nhash", 100),
		time,
		10,
		5,
		0,
		uint64(time.Day()),
		[]types.QualifyingAction{},
	)
	rewardProgram.RemainingPoolBalance = rewardProgram.GetTotalRewardPool()
	rewardProgram.ClaimedAmount = sdk.NewInt64Coin("nhash", 0)

	addr, _ := sdk.AccAddressFromBech32("cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv")
	beforeBalance := s.app.BankKeeper.GetBalance(s.ctx, addr, "nhash")
	err := s.app.RewardKeeper.RefundRewardClaims(s.ctx, rewardProgram)
	afterBalance := s.app.BankKeeper.GetBalance(s.ctx, addr, "nhash")

	s.Assert().NoError(err, "no error should be thrown")
	s.Assert().Equal(beforeBalance, afterBalance, "balance should stay same since all claims are taken")
}

func (s *KeeperTestSuite) TestClaimAllRewards() {
	time := s.ctx.BlockTime()

	for i := 0; i < 3; i++ {
		rewardProgram := types.NewRewardProgram(
			"title",
			"description",
			uint64(i+1),
			"cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv",
			sdk.NewInt64Coin("nhash", 1000),
			sdk.NewInt64Coin("nhash", 100),
			time,
			10,
			3,
			0,
			uint64(time.Day()),
			[]types.QualifyingAction{
				{
					Type: &types.QualifyingAction_Vote{
						Vote: &types.ActionVote{
							MinimumActions:          0,
							MaximumActions:          1,
							MinimumDelegationAmount: minDelegation,
						},
					},
				},
				{
					Type: &types.QualifyingAction_Delegate{
						Delegate: &types.ActionDelegate{
							MinimumActions:               0,
							MaximumActions:               1,
							MinimumDelegationAmount:      &minDelegation,
							MaximumDelegationAmount:      &maxDelegation,
							MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
							MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
						},
					},
				},
			},
		)
		rewardProgram.State = types.RewardProgram_STATE_FINISHED
		rewardProgram.CurrentClaimPeriod = rewardProgram.GetClaimPeriods()
		s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

		for j := 1; j <= int(rewardProgram.GetClaimPeriods()); j++ {
			state := types.NewRewardAccountState(rewardProgram.GetId(), uint64(j), "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 1, map[string]uint64{})
			state.ClaimStatus = types.RewardAccountState_CLAIM_STATUS_CLAIMABLE
			s.app.RewardKeeper.SetRewardAccountState(s.ctx, state)
			distribution := types.NewClaimPeriodRewardDistribution(uint64(j), rewardProgram.GetId(), sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, true)
			s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, distribution)
		}
	}

	details, reward, err := s.app.RewardKeeper.ClaimAllRewards(s.ctx, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv")
	s.Assert().NoError(err, "should throw no error")
	s.Assert().Equal(3, len(details), "should have rewards from every program")
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 900), reward[0], "should total up the rewards from the periods")

	for i := 0; i < len(details); i++ {
		s.Assert().Equal(3, len(details[i].ClaimedRewardPeriodDetails), "should have claims from every period")
		s.Assert().Equal(sdk.NewInt64Coin("nhash", 300), details[i].TotalRewardClaim, "should total up the rewards from the periods")
		s.Assert().Equal(uint64(i+1), details[i].RewardProgramId, "should have the correct id")
	}
}

func (s *KeeperTestSuite) TestClaimAllRewardsExpired() {
	time := s.ctx.BlockTime()

	for i := 0; i < 3; i++ {
		rewardProgram := types.NewRewardProgram(
			"title",
			"description",
			uint64(i+1),
			"cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv",
			sdk.NewInt64Coin("nhash", 1000),
			sdk.NewInt64Coin("nhash", 100),
			time,
			10,
			3,
			0,
			uint64(time.Day()),
			[]types.QualifyingAction{
				{
					Type: &types.QualifyingAction_Vote{
						Vote: &types.ActionVote{
							MinimumActions:          0,
							MaximumActions:          1,
							MinimumDelegationAmount: minDelegation,
						},
					},
				},
				{
					Type: &types.QualifyingAction_Delegate{
						Delegate: &types.ActionDelegate{
							MinimumActions:               0,
							MaximumActions:               1,
							MinimumDelegationAmount:      &minDelegation,
							MaximumDelegationAmount:      &maxDelegation,
							MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
							MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
						},
					},
				},
			},
		)
		rewardProgram.State = types.RewardProgram_STATE_EXPIRED
		rewardProgram.CurrentClaimPeriod = rewardProgram.GetClaimPeriods()
		s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

		for j := 1; j <= int(rewardProgram.GetClaimPeriods()); j++ {
			state := types.NewRewardAccountState(rewardProgram.GetId(), uint64(j), "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 1, map[string]uint64{})
			state.ClaimStatus = types.RewardAccountState_CLAIM_STATUS_EXPIRED
			s.app.RewardKeeper.SetRewardAccountState(s.ctx, state)
			distribution := types.NewClaimPeriodRewardDistribution(uint64(j), rewardProgram.GetId(), sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, true)
			s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, distribution)
		}
	}

	details, reward, err := s.app.RewardKeeper.ClaimAllRewards(s.ctx, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv")
	s.Assert().NoError(err, "should throw no error")
	s.Assert().Equal(0, len(details), "should have rewards from every program")
	s.Assert().Equal(0, len(reward), "should total up the rewards from the periods")
}

func (s *KeeperTestSuite) TestClaimAllRewardsNoPrograms() {
	details, reward, err := s.app.RewardKeeper.ClaimAllRewards(s.ctx, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv")
	s.Assert().NoError(err, "should throw no error")
	s.Assert().Equal(0, len(details), "should have rewards from every program")
	s.Assert().Equal(0, len(reward), "should total up the rewards from the periods")
}
