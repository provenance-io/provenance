package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (suite *KeeperTestSuite) TestStartRewardProgram() {
	suite.SetupTest()

	currentTime := time.Now()
	blockTime := suite.ctx.BlockTime()
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
		[]types.QualifyingAction{},
	)

	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &program)

	suite.Assert().Equal(program.State, types.RewardProgram_STARTED, "reward program should be in started state")
	suite.Assert().Equal(uint64(1), program.CurrentClaimPeriod, "current claim period should be set to 1")
	suite.Assert().Equal(blockTime.Add(time.Duration(program.ClaimPeriodSeconds)*time.Second), program.ClaimPeriodEndTime, "claim period end time should be set")

	claimPeriodAmount := program.GetTotalRewardPool().Amount.Quo(sdk.NewInt(int64(program.GetClaimPeriods())))
	claimPeriodPool := sdk.NewCoin(program.GetTotalRewardPool().Denom, claimPeriodAmount)
	reward, err := suite.app.RewardKeeper.GetClaimPeriodRewardDistribution(suite.ctx, 1, 1)
	suite.Assert().NoError(err)
	suite.Assert().Equal(uint64(1), reward.GetRewardProgramId())
	suite.Assert().Equal(uint64(1), reward.GetClaimPeriodId())
	suite.Assert().Equal(claimPeriodPool, reward.GetRewardsPool())
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 0), reward.GetTotalRewardsPoolForClaimPeriod())
}

func (suite *KeeperTestSuite) TestStartRewardProgramWithNotEnoughBalance() {
	suite.SetupTest()
	suite.Assert().Fail("not yet implemented")
}

func (suite *KeeperTestSuite) TestStartRewardProgramClaimPeriod() {
	suite.SetupTest()

	currentTime := time.Now()
	blockTime := suite.ctx.BlockTime()
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
		[]types.QualifyingAction{},
	)

	suite.app.RewardKeeper.StartRewardProgramClaimPeriod(suite.ctx, &program)
	suite.Assert().Equal(uint64(1), program.CurrentClaimPeriod, "current claim period should incremented")
	suite.Assert().Equal(blockTime.Add(time.Duration(program.ClaimPeriodSeconds)*time.Second), program.ClaimPeriodEndTime, "claim period end time should be set")

	claimPeriodAmount := program.GetTotalRewardPool().Amount.Quo(sdk.NewInt(int64(program.GetClaimPeriods())))
	claimPeriodPool := sdk.NewCoin(program.GetTotalRewardPool().Denom, claimPeriodAmount)
	reward, err := suite.app.RewardKeeper.GetClaimPeriodRewardDistribution(suite.ctx, 1, 1)
	suite.Assert().NoError(err)
	suite.Assert().Equal(uint64(1), reward.GetRewardProgramId())
	suite.Assert().Equal(uint64(1), reward.GetClaimPeriodId())
	suite.Assert().Equal(claimPeriodPool, reward.GetRewardsPool())
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 0), reward.GetTotalRewardsPoolForClaimPeriod())
}

func (suite *KeeperTestSuite) TestEndRewardProgram() {
	suite.SetupTest()

	currentTime := time.Now()
	blockTime := suite.ctx.BlockTime()
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
		[]types.QualifyingAction{},
	)

	suite.app.RewardKeeper.EndRewardProgram(suite.ctx, &program)

	suite.Assert().Equal(program.State, types.RewardProgram_FINISHED, "reward program should be in finished state")
	suite.Assert().Equal(blockTime, program.ActualProgramEndTime, "actual program end time should be set")
}

func (suite *KeeperTestSuite) TestRewardProgramClaimPeriodEnd() {
	suite.SetupTest()

	currentTime := time.Now()
	blockTime := suite.ctx.BlockTime()
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
		[]types.QualifyingAction{},
	)

	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &program)
	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)

	suite.Assert().Equal(program.State, types.RewardProgram_STARTED, "reward program should be in started state")
	suite.Assert().Equal(uint64(2), program.CurrentClaimPeriod, "current claim period should be updated")
	suite.Assert().Equal(blockTime.Add(time.Duration(program.ClaimPeriodSeconds)*time.Second), program.ClaimPeriodEndTime, "claim period end time should be set")
}

func (suite *KeeperTestSuite) TestRewardProgramClaimPeriodEndTransition() {
	suite.SetupTest()

	currentTime := time.Now()
	blockTime := suite.ctx.BlockTime()
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
		[]types.QualifyingAction{},
	)

	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &program)
	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)
	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)
	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)

	suite.Assert().Equal(program.State, types.RewardProgram_FINISHED, "reward program should be in finished state")
	suite.Assert().Equal(uint64(4), program.CurrentClaimPeriod, "current claim period should be updated")
	suite.Assert().Equal(blockTime.Add(time.Duration(program.ClaimPeriodSeconds)*time.Second), program.ClaimPeriodEndTime, "claim period end time should be set")
	suite.Assert().Equal(blockTime, program.ActualProgramEndTime, "claim period end time should be set")
}

func (suite *KeeperTestSuite) TestRewardProgramClaimPeriodEndTransitionExpired() {
	suite.SetupTest()

	currentTime := time.Now()
	suite.ctx = suite.ctx.WithBlockTime(currentTime)
	blockTime := suite.ctx.BlockTime()
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
		[]types.QualifyingAction{},
	)

	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &program)
	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)
	// Normally you would need an additional claim period. However, it should end because the expected time is set.
	program.ExpectedProgramEndTime = currentTime
	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)

	suite.Assert().Equal(types.RewardProgram_FINISHED, program.State, "reward program should be in finished state")
	suite.Assert().Equal(uint64(3), program.CurrentClaimPeriod, "current claim period should be updated")
	suite.Assert().Equal(blockTime.Add(time.Duration(program.ClaimPeriodSeconds)*time.Second), program.ClaimPeriodEndTime, "claim period end time should be set")
	suite.Assert().Equal(blockTime, program.ActualProgramEndTime, "claim period end time should be set")
}

func (suite *KeeperTestSuite) TestRewardProgramClaimPeriodEndNoBalance() {
	suite.SetupTest()
	suite.Assert().Fail("not yet implemented")
}

func (suite *KeeperTestSuite) TestRewardProgramClaimPeriodEndExtraBalance() {
	suite.SetupTest()
	suite.Assert().Fail("not yet implemented")
}

func (suite *KeeperTestSuite) TestEndRewardProgramClaimPeriodUpdatesBalances() {
	suite.SetupTest()
	suite.Assert().Fail("not yet implemented")
}

func (suite *KeeperTestSuite) TestEndRewardProgramClaimPeriodHandlesInvalidLookups() {
	suite.SetupTest()
	suite.Assert().Fail("not yet implemented")
}

func (suite *KeeperTestSuite) TestSumRewardClaimPeriodRewards() {
	suite.SetupTest()
	suite.Assert().Fail("not yet implemented")
}

func (suite *KeeperTestSuite) TestSumRewardClaimPeriodRewardsUsesMaxReward() {
	suite.SetupTest()
	suite.Assert().Fail("not yet implemented")
}

func (suite *KeeperTestSuite) TestSumRewardClaimPeriodRewardsUsesDoesNotExceedProgramBalance() {
	suite.SetupTest()
	suite.Assert().Fail("not yet implemented")
}

func (suite *KeeperTestSuite) TestSumRewardClaimPeriodRewardsNoTotalShares() {
	suite.SetupTest()
	suite.Assert().Fail("not yet implemented")
}

func (suite *KeeperTestSuite) TestSumRewardClaimPeriodRewardsHandlesInvalidLookups() {
	suite.SetupTest()
	suite.Assert().Fail("not yet implemented")
}

func (suite *KeeperTestSuite) TestCleanup() {
	suite.SetupTest()

	currentTime := time.Now()
	blockTime := suite.ctx.BlockTime()

	hasShares := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		currentTime,
		60*60,
		3,
		[]types.QualifyingAction{},
	)
	hasShares.State = types.RewardProgram_FINISHED

	hasExpiredShares := types.NewRewardProgram(
		"title",
		"description",
		2,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		currentTime,
		60*60,
		3,
		[]types.QualifyingAction{},
	)
	hasExpiredShares.State = types.RewardProgram_FINISHED

	hasNoShares := types.NewRewardProgram(
		"title",
		"description",
		3,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		currentTime,
		60*60,
		3,
		[]types.QualifyingAction{},
	)
	hasNoShares.State = types.RewardProgram_FINISHED

	hasNotFinished := types.NewRewardProgram(
		"title",
		"description",
		4,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		currentTime,
		60*60,
		3,
		[]types.QualifyingAction{},
	)

	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, hasShares)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, hasExpiredShares)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, hasNoShares)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, hasNotFinished)

	share1 := types.NewShare(1, 1, "test", false, blockTime.Add(time.Duration(time.Hour)), 1)
	share2 := types.NewShare(2, 1, "test", false, blockTime, 1)
	share3 := types.NewShare(2, 2, "test", true, blockTime.Add(time.Duration(time.Hour)), 1)

	suite.app.RewardKeeper.SetShare(suite.ctx, &share1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share2)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share3)

	suite.app.RewardKeeper.Cleanup(suite.ctx)

	programs, err := suite.app.RewardKeeper.GetAllRewardPrograms(suite.ctx)
	suite.Assert().NoError(err)
	suite.Assert().Equal(2, len(programs))
	suite.Assert().Equal(uint64(1), programs[0].Id, "reward program 1 should still exist")
	suite.Assert().Equal(uint64(4), programs[1].Id, "reward program 4 should still exist")

	count := 0
	suite.app.RewardKeeper.IterateShares(suite.ctx, func(share types.Share) (stop bool) {
		count += 1
		return false
	})
	suite.Assert().Equal(1, count, "only clean shares should exist")
}

func (suite *KeeperTestSuite) TestUpdate() {
	suite.SetupTest()
	// Reward Program that has not started
	blockTime := suite.ctx.BlockTime()

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
		[]types.QualifyingAction{},
	)

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
		[]types.QualifyingAction{},
	)

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
		[]types.QualifyingAction{},
	)
	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &nextClaimPeriod)
	nextClaimPeriod.ClaimPeriodEndTime = blockTime

	// Reward program that is ready to end
	ending := types.NewRewardProgram(
		"title",
		"description",
		4,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		blockTime,
		uint64(time.Hour),
		1,
		[]types.QualifyingAction{},
	)
	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &ending)
	ending.ClaimPeriodEndTime = blockTime

	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, notStarted)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, starting)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, nextClaimPeriod)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, ending)

	// We call update
	suite.app.RewardKeeper.Update(suite.ctx)

	notStarted, _ = suite.app.RewardKeeper.GetRewardProgram(suite.ctx, notStarted.Id)
	starting, _ = suite.app.RewardKeeper.GetRewardProgram(suite.ctx, starting.Id)
	nextClaimPeriod, _ = suite.app.RewardKeeper.GetRewardProgram(suite.ctx, nextClaimPeriod.Id)
	ending, _ = suite.app.RewardKeeper.GetRewardProgram(suite.ctx, ending.Id)

	suite.Assert().Equal(uint64(0), notStarted.CurrentClaimPeriod, "claim period should be 0 for a program that is not started")
	suite.Assert().Equal(notStarted.State, types.RewardProgram_PENDING, "should be in pending state")

	suite.Assert().Equal(uint64(1), starting.CurrentClaimPeriod, "claim period should be 1 for a program that just started")
	suite.Assert().Equal(starting.State, types.RewardProgram_STARTED, "should be in started state")

	suite.Assert().Equal(uint64(2), nextClaimPeriod.CurrentClaimPeriod, "claim period should be 2 for a program that went to next claim period")
	suite.Assert().Equal(nextClaimPeriod.State, types.RewardProgram_STARTED, "should be in started state")

	suite.Assert().Equal(uint64(2), ending.CurrentClaimPeriod, "claim period should be incremented by 1")
	suite.Assert().Equal(ending.State, types.RewardProgram_FINISHED, "should be in finished state")
}
