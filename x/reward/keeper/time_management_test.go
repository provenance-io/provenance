package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (s *KeeperTestSuite) TestStartRewardProgram() {
	currentTime := time.Now()
	blockTime := s.ctx.BlockTime()
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		currentTime,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program.RemainingPoolBalance = program.GetTotalRewardPool()

	s.app.RewardKeeper.StartRewardProgram(s.ctx, &program)

	s.Assert().Equal(program.State, types.RewardProgram_STATE_STARTED, "reward program should be in started state")
	s.Assert().Equal(uint64(1), program.CurrentClaimPeriod, "current claim period should be set to 1")
	s.Assert().Equal(blockTime.Add(time.Duration(program.ClaimPeriodSeconds)*time.Second), program.ClaimPeriodEndTime, "claim period end time should be set")

	claimPeriodAmount := program.GetTotalRewardPool().Amount.Quo(sdk.NewInt(int64(program.GetClaimPeriods())))
	claimPeriodPool := sdk.NewCoin(program.GetTotalRewardPool().Denom, claimPeriodAmount)
	reward, err := s.app.RewardKeeper.GetClaimPeriodRewardDistribution(s.ctx, 1, 1)
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(1), reward.GetRewardProgramId())
	s.Assert().Equal(uint64(1), reward.GetClaimPeriodId())
	s.Assert().Equal(claimPeriodPool, reward.GetRewardsPool())
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 0), reward.GetTotalRewardsPoolForClaimPeriod())

	events := s.ctx.EventManager().ABCIEvents()
	newEvent := events[len(events)-1]
	s.Assert().Equal("reward_program_started", newEvent.GetType(), "should emit the correct event type")
	s.Assert().Equal([]byte("reward_program_id"), newEvent.GetAttributes()[0].GetKey(), "should emit the correct attribute name")
	s.Assert().Equal([]byte("1"), newEvent.GetAttributes()[0].GetValue(), "should emit the correct attribute value")
}

func (s *KeeperTestSuite) TestStartRewardProgramNoBalance() {
	currentTime := time.Now()
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 0),
		sdk.NewInt64Coin("nhash", 0),
		currentTime,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program.RemainingPoolBalance = program.GetTotalRewardPool()

	err := s.app.RewardKeeper.StartRewardProgram(s.ctx, &program)
	s.Assert().Error(err, "an error should be thrown when there is no balance")
	s.Assert().Equal(program.State, types.RewardProgram_STATE_PENDING, "reward program should be in pending state")
	s.Assert().Equal(uint64(0), program.CurrentClaimPeriod, "current claim period should be set to 0")
}

func (s *KeeperTestSuite) TestStartNilRewardProgram() {
	err := s.app.RewardKeeper.StartRewardProgram(s.ctx, nil)
	s.Assert().Error(err, "must throw error for nil case")
}

func (s *KeeperTestSuite) TestStartRewardProgramClaimPeriodWithNil() {
	err := s.app.RewardKeeper.StartRewardProgramClaimPeriod(s.ctx, nil)
	s.Assert().Error(err, "should throw error")
}

func (s *KeeperTestSuite) TestStartRewardProgramClaimPeriodWithNoPeriods() {
	currentTime := time.Now()
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 100),
		currentTime,
		60*60,
		0,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)

	err := s.app.RewardKeeper.StartRewardProgramClaimPeriod(s.ctx, &program)
	s.Assert().Error(err, "should throw error")
}

func (s *KeeperTestSuite) TestStartRewardProgramClaimPeriod() {
	currentTime := time.Now()
	blockTime := s.ctx.BlockTime()
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 100),
		currentTime,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program.ExpectedProgramEndTime = s.ctx.BlockTime()
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program.RemainingPoolBalance = program.GetTotalRewardPool()

	s.app.RewardKeeper.StartRewardProgramClaimPeriod(s.ctx, &program)
	s.Assert().Equal(uint64(1), program.CurrentClaimPeriod, "current claim period should incremented")
	s.Assert().Equal(blockTime.Add(time.Duration(program.ClaimPeriodSeconds)*time.Second), program.ClaimPeriodEndTime, "claim period end time should be set")

	claimPeriodAmount := program.GetTotalRewardPool().Amount.Quo(sdk.NewInt(int64(program.GetClaimPeriods())))
	claimPeriodPool := sdk.NewCoin(program.GetTotalRewardPool().Denom, claimPeriodAmount)
	reward, err := s.app.RewardKeeper.GetClaimPeriodRewardDistribution(s.ctx, 1, 1)
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(1), reward.GetRewardProgramId())
	s.Assert().Equal(uint64(1), reward.GetClaimPeriodId())
	s.Assert().Equal(claimPeriodPool, reward.GetRewardsPool())
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 0), reward.GetTotalRewardsPoolForClaimPeriod())
	s.Assert().Equal(s.ctx.BlockTime(), program.ExpectedProgramEndTime, "expected program end time should not be updated.")
}

func (s *KeeperTestSuite) TestStartRewardProgramClaimPeriodUpdatesExpectedEndTime() {
	currentTime := time.Now()
	blockTime := s.ctx.BlockTime()
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 100),
		currentTime,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program.CurrentClaimPeriod = program.GetClaimPeriods()
	program.ExpectedProgramEndTime = s.ctx.BlockTime()
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program.RemainingPoolBalance = program.GetTotalRewardPool()

	s.app.RewardKeeper.StartRewardProgramClaimPeriod(s.ctx, &program)
	s.Assert().Equal(uint64(4), program.CurrentClaimPeriod, "current claim period should incremented")
	s.Assert().Equal(blockTime.Add(time.Duration(program.ClaimPeriodSeconds)*time.Second), program.ClaimPeriodEndTime, "claim period end time should be set")

	claimPeriodAmount := program.GetTotalRewardPool().Amount.Quo(sdk.NewInt(int64(program.GetClaimPeriods())))
	claimPeriodPool := sdk.NewCoin(program.GetTotalRewardPool().Denom, claimPeriodAmount)
	reward, err := s.app.RewardKeeper.GetClaimPeriodRewardDistribution(s.ctx, 4, 1)
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(1), reward.GetRewardProgramId())
	s.Assert().Equal(uint64(4), reward.GetClaimPeriodId())
	s.Assert().Equal(claimPeriodPool, reward.GetRewardsPool())
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 0), reward.GetTotalRewardsPoolForClaimPeriod())
	s.Assert().Equal(s.ctx.BlockTime().Add(time.Duration(program.ClaimPeriodSeconds)*time.Second), program.ExpectedProgramEndTime, "expected program end time should be updated for rollover.")
}

func (s *KeeperTestSuite) TestStartRewardProgramClaimPeriodDoesNotExceedBalance() {
	currentTime := time.Now()
	blockTime := s.ctx.BlockTime()
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 100),
		currentTime,
		60*60,
		4,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program.RemainingPoolBalance = sdk.NewInt64Coin("nhash", 20)

	s.app.RewardKeeper.StartRewardProgramClaimPeriod(s.ctx, &program)
	s.Assert().Equal(uint64(1), program.CurrentClaimPeriod, "current claim period should incremented")
	s.Assert().Equal(blockTime.Add(time.Duration(program.ClaimPeriodSeconds)*time.Second), program.ClaimPeriodEndTime, "claim period end time should be set")

	reward, err := s.app.RewardKeeper.GetClaimPeriodRewardDistribution(s.ctx, 1, 1)
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(1), reward.GetRewardProgramId())
	s.Assert().Equal(uint64(1), reward.GetClaimPeriodId())
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 20), reward.GetRewardsPool())
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 0), reward.GetTotalRewardsPoolForClaimPeriod())
}

func (s *KeeperTestSuite) TestEndRewardProgram() {
	currentTime := time.Now()
	blockTime := s.ctx.BlockTime()
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		currentTime,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)

	s.app.RewardKeeper.EndRewardProgram(s.ctx, &program)

	events := s.ctx.EventManager().ABCIEvents()
	newEvent := events[len(events)-1]
	s.Assert().Equal("reward_program_finished", newEvent.GetType(), "should emit the correct event type")
	s.Assert().Equal([]byte("reward_program_id"), newEvent.GetAttributes()[0].GetKey(), "should emit the correct attribute name")
	s.Assert().Equal([]byte("1"), newEvent.GetAttributes()[0].GetValue(), "should emit the correct attribute value")
	s.Assert().Equal(program.State, types.RewardProgram_STATE_FINISHED, "reward program should be in finished state")
	s.Assert().Equal(blockTime, program.ActualProgramEndTime, "actual program end time should be set")
}

func (s *KeeperTestSuite) TestEndRewardProgramNil() {
	err := s.app.RewardKeeper.EndRewardProgram(s.ctx, nil)
	s.Assert().Error(err, "should throw an error for nil")
}

func (s *KeeperTestSuite) TestExpireRewardProgram() {
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		time.Now(),
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)

	s.app.RewardKeeper.ExpireRewardProgram(s.ctx, &program)
	s.Assert().Equal(program.State, types.RewardProgram_STATE_EXPIRED, "reward program should be in expired state")
	events := s.ctx.EventManager().ABCIEvents()
	newEvent := events[len(events)-1]
	s.Assert().Equal("reward_program_expired", newEvent.GetType(), "should emit the correct event type")
	s.Assert().Equal([]byte("reward_program_id"), newEvent.GetAttributes()[0].GetKey(), "should emit the correct attribute name")
	s.Assert().Equal([]byte("1"), newEvent.GetAttributes()[0].GetValue(), "should emit the correct attribute value")
}

func (s *KeeperTestSuite) TestExpireRewardProgramNil() {
	err := s.app.RewardKeeper.ExpireRewardProgram(s.ctx, nil)
	s.Assert().Error(err, "should throw an error for nil")
}

func (s *KeeperTestSuite) TestExpireRewardProgramRefunds() {
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		time.Now(),
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program.RemainingPoolBalance = sdk.NewInt64Coin("nhash", 80000)
	program.ClaimedAmount = sdk.NewInt64Coin("nhash", 10000)

	addr, _ := sdk.AccAddressFromBech32("cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv")
	beforeBalance := s.app.BankKeeper.GetBalance(s.ctx, addr, "nhash")

	s.app.RewardKeeper.ExpireRewardProgram(s.ctx, &program)

	afterBalance := s.app.BankKeeper.GetBalance(s.ctx, addr, "nhash")

	s.Assert().Equal(beforeBalance.Add(sdk.NewInt64Coin("nhash", 90000)), afterBalance, "account should get remaining balance and claims")
	s.Assert().Equal(program.State, types.RewardProgram_STATE_EXPIRED, "reward program should be in expired state")
}

func (s *KeeperTestSuite) TestCalculateRewardClaimPeriodRewardsNonMatchingDenoms() {
	notMatching := types.NewClaimPeriodRewardDistribution(
		1,
		1,
		sdk.NewInt64Coin("hotdog", 0),
		sdk.NewInt64Coin("hotdog", 0),
		1,
		false,
	)

	_, err := s.app.RewardKeeper.CalculateRewardClaimPeriodRewards(s.ctx, sdk.NewInt64Coin("nhash", 0), notMatching)
	s.Assert().Error(err, "error should be thrown when claim period reward distribution doesn't match the others")
}

func (s *KeeperTestSuite) TestCalculateRewardClaimPeriodRewardsNoSharesForPeriod() {
	matching := types.NewClaimPeriodRewardDistribution(
		1,
		1,
		sdk.NewInt64Coin("nhash", 0),
		sdk.NewInt64Coin("nhash", 0),
		0,
		false,
	)

	reward, err := s.app.RewardKeeper.CalculateRewardClaimPeriodRewards(s.ctx, sdk.NewInt64Coin("nhash", 0), matching)
	s.Assert().NoError(err, "No error should be thrown when there are no claimed shares")
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 0), reward, "should be 0 of the input denom")
}

func (s *KeeperTestSuite) TestCalculateRewardClaimPeriodRewardsEvenDistributionNoRemainder() {
	distribution := types.NewClaimPeriodRewardDistribution(
		1,
		1,
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 0),
		2,
		false,
	)

	state1 := types.NewRewardAccountState(1, 1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 1, []*types.ActionCounter{})
	state2 := types.NewRewardAccountState(1, 1, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", 1, []*types.ActionCounter{})
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state2)

	reward, err := s.app.RewardKeeper.CalculateRewardClaimPeriodRewards(s.ctx, sdk.NewInt64Coin("nhash", 100), distribution)
	s.Assert().NoError(err, "should return no error")
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 100), reward, "should distribute all the funds")
}

func (s *KeeperTestSuite) TestCalculateRewardClaimPeriodRewardsEvenDistributionWithRemainder() {
	distribution := types.NewClaimPeriodRewardDistribution(
		1,
		1,
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 0),
		3,
		false,
	)

	state1 := types.NewRewardAccountState(1, 1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 1, []*types.ActionCounter{})
	state2 := types.NewRewardAccountState(1, 1, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", 1, []*types.ActionCounter{})
	state3 := types.NewRewardAccountState(1, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 1, []*types.ActionCounter{})
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state2)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state3)

	reward, err := s.app.RewardKeeper.CalculateRewardClaimPeriodRewards(s.ctx, sdk.NewInt64Coin("nhash", 100), distribution)
	s.Assert().NoError(err, "should return no error")
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 99), reward, "should distribute all the funds except for the remainder")
}

func (s *KeeperTestSuite) TestCalculateRewardClaimPeriodRewardsUnevenDistribution() {
	distribution := types.NewClaimPeriodRewardDistribution(
		1,
		1,
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 0),
		4,
		false,
	)

	state1 := types.NewRewardAccountState(1, 1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 2, []*types.ActionCounter{})
	state2 := types.NewRewardAccountState(1, 1, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", 1, []*types.ActionCounter{})
	state3 := types.NewRewardAccountState(1, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 1, []*types.ActionCounter{})
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state2)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state3)

	reward, err := s.app.RewardKeeper.CalculateRewardClaimPeriodRewards(s.ctx, sdk.NewInt64Coin("nhash", 100), distribution)
	s.Assert().NoError(err, "should return no error")
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 100), reward, "should distribute all the funds")
}

func (s *KeeperTestSuite) TestCalculateRewardClaimPeriodRewardsUsesMaxReward() {
	distribution := types.NewClaimPeriodRewardDistribution(
		1,
		1,
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 0),
		2,
		false,
	)

	state1 := types.NewRewardAccountState(1, 1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 1, []*types.ActionCounter{})
	state2 := types.NewRewardAccountState(1, 1, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", 1, []*types.ActionCounter{})
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state2)

	reward, err := s.app.RewardKeeper.CalculateRewardClaimPeriodRewards(s.ctx, sdk.NewInt64Coin("nhash", 20), distribution)
	s.Assert().NoError(err, "should return no error")
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 40), reward, "should distribute only up to the maximum reward for each participant")
}

func (s *KeeperTestSuite) TestCalculateParticipantReward() {
	reward := s.app.RewardKeeper.CalculateParticipantReward(s.ctx, 1, 2, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100))
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 50), reward, "should get correct cut of pool")
}

func (s *KeeperTestSuite) TestCalculateParticipantRewardLimitsToMaximum() {
	reward := s.app.RewardKeeper.CalculateParticipantReward(s.ctx, 1, 2, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 10))
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 10), reward, "should get correct cut of pool")
}

func (s *KeeperTestSuite) TestCalculateParticipantRewardCanHandleZeroTotalShares() {
	reward := s.app.RewardKeeper.CalculateParticipantReward(s.ctx, 1, 0, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100))
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 0), reward, "should have no reward")
}

func (s *KeeperTestSuite) TestCalculateParticipantRewardCanHandleZeroEarnedShares() {
	reward := s.app.RewardKeeper.CalculateParticipantReward(s.ctx, 0, 10, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100))
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 0), reward, "should have no reward")
}

func (s *KeeperTestSuite) TestCalculateParticipantRewardCanHandleZeroRewardPool() {
	reward := s.app.RewardKeeper.CalculateParticipantReward(s.ctx, 1, 1, sdk.NewInt64Coin("nhash", 0), sdk.NewInt64Coin("nhash", 100))
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 0), reward, "should have no reward")
}

func (s *KeeperTestSuite) TestCalculateParticipantRewardTruncates() {
	reward := s.app.RewardKeeper.CalculateParticipantReward(s.ctx, 1, 3, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100))
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 33), reward, "reward should truncate when < .5")

	reward = s.app.RewardKeeper.CalculateParticipantReward(s.ctx, 2, 3, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100))
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 66), reward, "reward should truncate when >= .5")
}

func (s *KeeperTestSuite) TestEndRewardProgramClaimPeriodHandlesInvalidLookups() {
	currentTime := time.Now()
	s.ctx = s.ctx.WithBlockTime(currentTime)

	program1 := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 0),
		sdk.NewInt64Coin("nhash", 1000),
		currentTime,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program2 := types.NewRewardProgram(
		"title",
		"description",
		2,
		"insert address",
		sdk.NewInt64Coin("nhash", 0),
		sdk.NewInt64Coin("nhash", 1000),
		currentTime,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program3 := types.NewRewardProgram(
		"title",
		"description",
		3,
		"insert address",
		sdk.NewInt64Coin("nhash", 0),
		sdk.NewInt64Coin("nhash", 1000),
		currentTime,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program1.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program2.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program3.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program2.RemainingPoolBalance = program2.GetTotalRewardPool()
	program3.RemainingPoolBalance = sdk.NewInt64Coin("jackthecat", 100)
	rewardDistribution := types.NewClaimPeriodRewardDistribution(0, 3, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, false)
	s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, rewardDistribution)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program1)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program2)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program3)

	err := s.app.RewardKeeper.EndRewardProgramClaimPeriod(s.ctx, &program1)
	s.Assert().Error(err, "an error should be thrown if there is no program balance for the program")
	err = s.app.RewardKeeper.EndRewardProgramClaimPeriod(s.ctx, &program2)
	s.Assert().Error(err, "an error should be thrown if there is no claim period reward distribution for the program")
	err = s.app.RewardKeeper.EndRewardProgramClaimPeriod(s.ctx, &program3)
	s.Assert().Error(err, "an error should be thrown if reward claim calculation fails")
}

func (s *KeeperTestSuite) TestEndRewardProgramClaimPeriodHandlesNilRewardProgram() {
	err := s.app.RewardKeeper.EndRewardProgramClaimPeriod(s.ctx, nil)
	s.Assert().Error(err, "error should be returned for nil reward program")
}

func (s *KeeperTestSuite) TestRewardProgramClaimPeriodEnd() {
	currentTime := time.Now()
	blockTime := s.ctx.BlockTime()
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 100000),
		currentTime,
		60*60,
		2,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	state1 := types.NewRewardAccountState(1, 1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 1, []*types.ActionCounter{})
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)
	program.RemainingPoolBalance = program.GetTotalRewardPool()

	s.app.RewardKeeper.StartRewardProgram(s.ctx, &program)

	// Update the distribution to replicate that a share was actually granted.
	rewardDistribution, _ := s.app.RewardKeeper.GetClaimPeriodRewardDistribution(s.ctx, 1, 1)
	rewardDistribution.TotalShares = 1
	s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, rewardDistribution)

	s.app.RewardKeeper.EndRewardProgramClaimPeriod(s.ctx, &program)

	reward, _ := s.app.RewardKeeper.GetClaimPeriodRewardDistribution(s.ctx, 1, 1)

	s.Assert().Equal(sdk.NewInt64Coin("nhash", 50000), program.RemainingPoolBalance, "balance should subtract the claim period reward")
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 50000), reward.TotalRewardsPoolForClaimPeriod, "total claim should be increased by the amount rewarded")
	s.Assert().Equal(program.State, types.RewardProgram_STATE_STARTED, "reward program should be in started state")
	s.Assert().Equal(uint64(2), program.CurrentClaimPeriod, "current claim period should be updated")
	s.Assert().Equal(blockTime.Add(time.Duration(program.ClaimPeriodSeconds)*time.Second), program.ClaimPeriodEndTime, "claim period end time should be set")
}

func (s *KeeperTestSuite) TestRewardProgramClaimPeriodEndTransition() {
	currentTime := time.Now()
	blockTime := s.ctx.BlockTime()
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 100000),
		currentTime,
		60*60,
		2,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program.RemainingPoolBalance = program.GetTotalRewardPool()

	state1 := types.NewRewardAccountState(1, 1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 1, []*types.ActionCounter{})
	state2 := types.NewRewardAccountState(1, 2, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", 1, []*types.ActionCounter{})
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state2)

	s.app.RewardKeeper.StartRewardProgram(s.ctx, &program)
	reward, _ := s.app.RewardKeeper.GetClaimPeriodRewardDistribution(s.ctx, 1, 1)
	reward.TotalShares = 1
	s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, reward)
	s.app.RewardKeeper.EndRewardProgramClaimPeriod(s.ctx, &program)
	reward, _ = s.app.RewardKeeper.GetClaimPeriodRewardDistribution(s.ctx, 2, 1)
	reward.TotalShares = 1
	s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, reward)
	s.app.RewardKeeper.EndRewardProgramClaimPeriod(s.ctx, &program)

	s.Assert().Equal(program.State, types.RewardProgram_STATE_FINISHED, "reward program should be in finished state")
	s.Assert().Equal(uint64(2), program.CurrentClaimPeriod, "current claim period should not be updated")
	s.Assert().Equal(blockTime.Add(time.Duration(program.ClaimPeriodSeconds)*time.Second), program.ClaimPeriodEndTime, "claim period end time should be set")
	s.Assert().Equal(blockTime, program.ActualProgramEndTime, "claim period end time should be set")
}

func (s *KeeperTestSuite) TestRewardProgramClaimPeriodEndTransitionExpired() {
	currentTime := time.Now()
	s.ctx = s.ctx.WithBlockTime(currentTime)
	blockTime := s.ctx.BlockTime()
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		currentTime,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program.RemainingPoolBalance = program.GetTotalRewardPool()

	s.app.RewardKeeper.StartRewardProgram(s.ctx, &program)
	s.app.RewardKeeper.EndRewardProgramClaimPeriod(s.ctx, &program)
	// Normally you would need an additional claim period. However, it should end because the expected time is set.
	program.ProgramEndTimeMax = currentTime
	s.app.RewardKeeper.EndRewardProgramClaimPeriod(s.ctx, &program)

	s.Assert().Equal(types.RewardProgram_STATE_FINISHED, program.State, "reward program should be in finished state")
	s.Assert().Equal(uint64(2), program.CurrentClaimPeriod, "current claim period should not be updated")
	s.Assert().Equal(blockTime.Add(time.Duration(program.ClaimPeriodSeconds)*time.Second), program.ClaimPeriodEndTime, "claim period end time should be set")
	s.Assert().Equal(blockTime, program.ActualProgramEndTime, "claim period end time should be set")
}

func (s *KeeperTestSuite) TestRewardProgramClaimPeriodEndNoBalance() {
	currentTime := time.Now()
	s.ctx = s.ctx.WithBlockTime(currentTime)
	blockTime := s.ctx.BlockTime()
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 1000),
		sdk.NewInt64Coin("nhash", 1000),
		currentTime,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program.RemainingPoolBalance = sdk.NewInt64Coin("nhash", 0)

	s.app.RewardKeeper.StartRewardProgram(s.ctx, &program)
	s.app.RewardKeeper.EndRewardProgramClaimPeriod(s.ctx, &program)

	s.Assert().Equal(types.RewardProgram_STATE_FINISHED, program.State, "reward program should be in finished state")
	s.Assert().Equal(uint64(1), program.CurrentClaimPeriod, "current claim period should not be updated")
	s.Assert().Equal(program.ClaimPeriodEndTime, program.ClaimPeriodEndTime, "claim period end time should not be updated")
	s.Assert().Equal(blockTime, program.ActualProgramEndTime, "actual end time should be set")
}

func (s *KeeperTestSuite) TestEndRewardProgramClaimPeriodUpdatesClaimStatus() {
	currentTime := time.Now()
	s.ctx = s.ctx.WithBlockTime(currentTime)
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 400),
		sdk.NewInt64Coin("nhash", 1000),
		currentTime,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program.RemainingPoolBalance = program.GetTotalRewardPool()

	state1 := types.NewRewardAccountState(1, 1, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", 1, []*types.ActionCounter{})
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)
	state2 := types.NewRewardAccountState(1, 1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 1, []*types.ActionCounter{})
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state2)

	s.app.RewardKeeper.StartRewardProgram(s.ctx, &program)
	reward, _ := s.app.RewardKeeper.GetClaimPeriodRewardDistribution(s.ctx, 1, 1)
	reward.TotalShares = 1
	s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, reward)
	s.app.RewardKeeper.EndRewardProgramClaimPeriod(s.ctx, &program)

	state1, _ = s.app.RewardKeeper.GetRewardAccountState(s.ctx, 1, 1, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27")
	state2, _ = s.app.RewardKeeper.GetRewardAccountState(s.ctx, 1, 1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")

	// Adjusted after ending period
	s.Assert().Equal(types.RewardAccountState_CLAIM_STATUS_CLAIMABLE, state1.GetClaimStatus(), "first claim status should be updated to claimable")
	s.Assert().Equal(types.RewardAccountState_CLAIM_STATUS_CLAIMABLE, state2.GetClaimStatus(), "second claim status should be updated to claimable")
}

func (s *KeeperTestSuite) TestEndRewardProgramClaimPeriodUpdatesBalances() {
	currentTime := time.Now()
	s.ctx = s.ctx.WithBlockTime(currentTime)
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 400),
		sdk.NewInt64Coin("nhash", 1000),
		currentTime,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program.RemainingPoolBalance = program.GetTotalRewardPool()

	state1 := types.NewRewardAccountState(1, 1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 1, []*types.ActionCounter{})
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)

	s.app.RewardKeeper.StartRewardProgram(s.ctx, &program)
	reward, _ := s.app.RewardKeeper.GetClaimPeriodRewardDistribution(s.ctx, 1, 1)
	reward.TotalShares = 1
	s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, reward)
	claimAmount, _ := s.app.RewardKeeper.CalculateRewardClaimPeriodRewards(s.ctx, program.GetMaxRewardByAddress(), reward)
	s.app.RewardKeeper.EndRewardProgramClaimPeriod(s.ctx, &program)

	// Adjusted after ending period
	reward, _ = s.app.RewardKeeper.GetClaimPeriodRewardDistribution(s.ctx, 1, 1)
	expectedProgramBalance := program.GetTotalRewardPool().Sub(claimAmount)
	s.Assert().Equal(claimAmount, reward.GetTotalRewardsPoolForClaimPeriod(), "the reward for the claim period should be added to total reward")
	s.Assert().Equal(expectedProgramBalance, program.GetRemainingPoolBalance(), "the reward for the claim period should be subtracted out of the program balance")
	s.Assert().Equal(types.RewardProgram_STATE_STARTED, program.State, "reward program should be in started state")
	s.Assert().Equal(uint64(2), program.CurrentClaimPeriod, "next iteration should start")
	s.Assert().Equal(true, reward.ClaimPeriodEnded, "claim period should be marked as ended")
}

func (s *KeeperTestSuite) TestEndRewardProgramClaimPeriodHandlesMinimumRolloverAmount() {
	currentTime := time.Now()
	s.ctx = s.ctx.WithBlockTime(currentTime)
	blockTime := s.ctx.BlockTime()
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 1000),
		sdk.NewInt64Coin("nhash", 500),
		currentTime,
		60*60,
		2,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 501)
	program.RemainingPoolBalance = program.GetTotalRewardPool()

	s.app.RewardKeeper.StartRewardProgram(s.ctx, &program)

	// Create the shares
	state1 := types.NewRewardAccountState(1, 1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 1, []*types.ActionCounter{})
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)
	reward, _ := s.app.RewardKeeper.GetClaimPeriodRewardDistribution(s.ctx, 1, 1)
	reward.TotalShares = 1
	s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, reward)

	// Should end because the balance should be below 501
	s.app.RewardKeeper.EndRewardProgramClaimPeriod(s.ctx, &program)

	s.Assert().Equal(types.RewardProgram_STATE_FINISHED, program.State, "reward program should be in finished state")
	s.Assert().Equal(uint64(1), program.CurrentClaimPeriod, "current claim period should not be updated")
	s.Assert().Equal(program.ClaimPeriodEndTime, program.ClaimPeriodEndTime, "claim period end time should not be updated")
	s.Assert().Equal(blockTime, program.ActualProgramEndTime, "actual end time should be set")
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 500), program.GetRemainingPoolBalance(), "balance should be updated")
}

func (s *KeeperTestSuite) TestUpdate() {
	// Reward Program that has not started
	currentTime := time.Now()
	s.ctx = s.ctx.WithBlockTime(currentTime)
	blockTime := s.ctx.BlockTime()

	notStarted := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		blockTime.Add(time.Duration(time.Hour)),
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	notStarted.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	notStarted.RemainingPoolBalance = notStarted.GetTotalRewardPool()

	// Reward Program that is starting
	starting := types.NewRewardProgram(
		"title",
		"description",
		2,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		blockTime,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	starting.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	starting.RemainingPoolBalance = starting.GetTotalRewardPool()

	// Reward Program that is ready to move onto next claim period
	nextClaimPeriod := types.NewRewardProgram(
		"title",
		"description",
		3,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		blockTime,
		uint64(time.Hour),
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	nextClaimPeriod.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	nextClaimPeriod.RemainingPoolBalance = nextClaimPeriod.GetTotalRewardPool()
	s.app.RewardKeeper.StartRewardProgram(s.ctx, &nextClaimPeriod)
	nextClaimPeriod.ClaimPeriodEndTime = blockTime

	// Reward program that runs out of funds
	ending := types.NewRewardProgram(
		"title",
		"description",
		4,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 100000),
		blockTime,
		uint64(time.Hour),
		1,
		0,
		0,
		[]types.QualifyingAction{},
	)
	ending.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	ending.RemainingPoolBalance = sdk.NewInt64Coin("nhash", 0)
	state1 := types.NewRewardAccountState(4, 1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 1, []*types.ActionCounter{})
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)
	s.app.RewardKeeper.StartRewardProgram(s.ctx, &ending)
	ending.ClaimPeriodEndTime = blockTime

	// Reward program that times out
	timeout := types.NewRewardProgram(
		"title",
		"description",
		5,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 100000),
		blockTime,
		0,
		1,
		0,
		0,
		[]types.QualifyingAction{},
	)
	timeout.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	timeout.ClaimPeriodEndTime = blockTime
	timeout.ProgramEndTimeMax = blockTime
	timeout.RemainingPoolBalance = timeout.GetTotalRewardPool()
	s.app.RewardKeeper.StartRewardProgram(s.ctx, &timeout)

	// Reward program that times out
	expiring := types.NewRewardProgram(
		"title",
		"description",
		6,
		"cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 100000),
		blockTime,
		0,
		1,
		0,
		0,
		[]types.QualifyingAction{},
	)
	remainingBalance := expiring.GetTotalRewardPool()
	expiring.ActualProgramEndTime = blockTime
	expiring.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	expiring.State = types.RewardProgram_STATE_FINISHED
	expiring.RemainingPoolBalance = remainingBalance

	s.app.RewardKeeper.SetRewardProgram(s.ctx, notStarted)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, starting)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, nextClaimPeriod)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, ending)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, timeout)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, expiring)

	addr, _ := sdk.AccAddressFromBech32("cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv")
	beforeBalance := s.app.BankKeeper.GetBalance(s.ctx, addr, "nhash")

	// We call update
	s.app.RewardKeeper.UpdateUnexpiredRewardsProgram(s.ctx)

	afterBalance := s.app.BankKeeper.GetBalance(s.ctx, addr, "nhash")
	notStarted, _ = s.app.RewardKeeper.GetRewardProgram(s.ctx, notStarted.Id)
	starting, _ = s.app.RewardKeeper.GetRewardProgram(s.ctx, starting.Id)
	nextClaimPeriod, _ = s.app.RewardKeeper.GetRewardProgram(s.ctx, nextClaimPeriod.Id)
	ending, _ = s.app.RewardKeeper.GetRewardProgram(s.ctx, ending.Id)
	timeout, _ = s.app.RewardKeeper.GetRewardProgram(s.ctx, timeout.Id)
	expiring, _ = s.app.RewardKeeper.GetRewardProgram(s.ctx, expiring.Id)

	s.Assert().Equal(uint64(0), notStarted.CurrentClaimPeriod, "claim period should be 0 for a program that is not started")
	s.Assert().Equal(notStarted.State, types.RewardProgram_STATE_PENDING, "should be in pending state")

	s.Assert().Equal(uint64(1), starting.CurrentClaimPeriod, "claim period should be 1 for a program that just started")
	s.Assert().Equal(starting.State, types.RewardProgram_STATE_STARTED, "should be in started state")

	s.Assert().Equal(uint64(2), nextClaimPeriod.CurrentClaimPeriod, "claim period should be 2 for a program that went to next claim period")
	s.Assert().Equal(nextClaimPeriod.State, types.RewardProgram_STATE_STARTED, "should be in started state")

	s.Assert().Equal(uint64(1), ending.CurrentClaimPeriod, "claim period should not increment")
	s.Assert().Equal(ending.State, types.RewardProgram_STATE_FINISHED, "should be in finished state")

	s.Assert().Equal(uint64(1), timeout.CurrentClaimPeriod, "claim period should not increment")
	s.Assert().Equal(timeout.State, types.RewardProgram_STATE_FINISHED, "should be in finished state")

	s.Assert().Equal(expiring.State, types.RewardProgram_STATE_EXPIRED, "should be in expired state")
	s.Assert().Equal(beforeBalance.Add(remainingBalance), afterBalance, "balance should be refunded")
}
