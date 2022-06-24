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
	programBalance := types.NewRewardProgramBalance(program.GetId(), program.GetTotalRewardPool())
	suite.app.RewardKeeper.SetRewardProgramBalance(suite.ctx, programBalance)

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

func (suite *KeeperTestSuite) TestStartNilRewardProgram() {
	suite.SetupTest()
	err := suite.app.RewardKeeper.StartRewardProgram(suite.ctx, nil)
	suite.Assert().Error(err, "must throw error for nil case")
}

func (suite *KeeperTestSuite) TestStartRewardProgramClaimPeriodWithNil() {
	suite.SetupTest()

	err := suite.app.RewardKeeper.StartRewardProgramClaimPeriod(suite.ctx, nil)
	suite.Assert().Error(err, "should throw error")
}

func (suite *KeeperTestSuite) TestStartRewardProgramClaimPeriodWithNoPeriods() {
	suite.SetupTest()
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
		[]types.QualifyingAction{},
	)

	err := suite.app.RewardKeeper.StartRewardProgramClaimPeriod(suite.ctx, &program)
	suite.Assert().Error(err, "should throw error")
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
	programBalance := types.NewRewardProgramBalance(program.GetId(), program.GetTotalRewardPool())
	suite.app.RewardKeeper.SetRewardProgramBalance(suite.ctx, programBalance)

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

func (suite *KeeperTestSuite) TestStartRewardProgramClaimPeriodDoesNotExceedBalance() {
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
		4,
		[]types.QualifyingAction{},
	)
	programBalance := types.NewRewardProgramBalance(program.GetId(), sdk.NewInt64Coin("nhash", 20))
	suite.app.RewardKeeper.SetRewardProgramBalance(suite.ctx, programBalance)

	suite.app.RewardKeeper.StartRewardProgramClaimPeriod(suite.ctx, &program)
	suite.Assert().Equal(uint64(1), program.CurrentClaimPeriod, "current claim period should incremented")
	suite.Assert().Equal(blockTime.Add(time.Duration(program.ClaimPeriodSeconds)*time.Second), program.ClaimPeriodEndTime, "claim period end time should be set")

	reward, err := suite.app.RewardKeeper.GetClaimPeriodRewardDistribution(suite.ctx, 1, 1)
	suite.Assert().NoError(err)
	suite.Assert().Equal(uint64(1), reward.GetRewardProgramId())
	suite.Assert().Equal(uint64(1), reward.GetClaimPeriodId())
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 20), reward.GetRewardsPool())
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

func (suite *KeeperTestSuite) TestEndRewardProgramNil() {
	suite.SetupTest()
	err := suite.app.RewardKeeper.EndRewardProgram(suite.ctx, nil)
	suite.Assert().Error(err, "should throw an error for nil")
}

// We are good up to this point

func (suite *KeeperTestSuite) TestCalculateRewardClaimPeriodRewardsNonMatchingDenoms() {
	suite.SetupTest()
	notMatching := types.NewClaimPeriodRewardDistribution(
		1,
		1,
		sdk.NewInt64Coin("hotdog", 0),
		sdk.NewInt64Coin("hotdog", 0),
		1,
		false,
	)

	_, err := suite.app.RewardKeeper.CalculateRewardClaimPeriodRewards(suite.ctx, sdk.NewInt64Coin("nhash", 0), notMatching)
	suite.Assert().Error(err, "error should be thrown when claim period reward distribution doesn't match the others")
}

func (suite *KeeperTestSuite) TestCalculateRewardClaimPeriodRewardsNoSharesForPeriod() {
	suite.SetupTest()
	matching := types.NewClaimPeriodRewardDistribution(
		1,
		1,
		sdk.NewInt64Coin("nhash", 0),
		sdk.NewInt64Coin("nhash", 0),
		0,
		false,
	)

	reward, err := suite.app.RewardKeeper.CalculateRewardClaimPeriodRewards(suite.ctx, sdk.NewInt64Coin("nhash", 0), matching)
	suite.Assert().NoError(err, "No error should be thrown when there are no claimed shares")
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 0), reward, "should be 0 of the input denom")
}

func (suite *KeeperTestSuite) TestCalculateRewardClaimPeriodRewardsEvenDistributionNoRemainder() {
	suite.SetupTest()
	distribution := types.NewClaimPeriodRewardDistribution(
		1,
		1,
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 0),
		2,
		false,
	)

	share1 := types.NewShare(1, 1, "address1", false, time.Time{}, 1)
	share2 := types.NewShare(1, 1, "address2", false, time.Time{}, 1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share2)

	reward, err := suite.app.RewardKeeper.CalculateRewardClaimPeriodRewards(suite.ctx, sdk.NewInt64Coin("nhash", 100), distribution)
	suite.Assert().NoError(err, "should return no error")
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 100), reward, "should distribute all the funds")
}

func (suite *KeeperTestSuite) TestCalculateRewardClaimPeriodRewardsEvenDistributionWithRemainder() {
	suite.SetupTest()
	distribution := types.NewClaimPeriodRewardDistribution(
		1,
		1,
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 0),
		3,
		false,
	)

	share1 := types.NewShare(1, 1, "address1", false, time.Time{}, 1)
	share2 := types.NewShare(1, 1, "address2", false, time.Time{}, 1)
	share3 := types.NewShare(1, 1, "address3", false, time.Time{}, 1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share2)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share3)

	reward, err := suite.app.RewardKeeper.CalculateRewardClaimPeriodRewards(suite.ctx, sdk.NewInt64Coin("nhash", 100), distribution)
	suite.Assert().NoError(err, "should return no error")
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 99), reward, "should distribute all the funds except for the remainder")
}

func (suite *KeeperTestSuite) TestCalculateRewardClaimPeriodRewardsUnevenDistribution() {
	suite.SetupTest()
	distribution := types.NewClaimPeriodRewardDistribution(
		1,
		1,
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 0),
		4,
		false,
	)

	share1 := types.NewShare(1, 1, "address1", false, time.Time{}, 2)
	share2 := types.NewShare(1, 1, "address2", false, time.Time{}, 1)
	share3 := types.NewShare(1, 1, "address3", false, time.Time{}, 1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share2)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share3)

	reward, err := suite.app.RewardKeeper.CalculateRewardClaimPeriodRewards(suite.ctx, sdk.NewInt64Coin("nhash", 100), distribution)
	suite.Assert().NoError(err, "should return no error")
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 100), reward, "should distribute all the funds")
}

func (suite *KeeperTestSuite) TestCalculateRewardClaimPeriodRewardsUsesMaxReward() {
	suite.SetupTest()
	distribution := types.NewClaimPeriodRewardDistribution(
		1,
		1,
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 0),
		2,
		false,
	)

	share1 := types.NewShare(1, 1, "address1", false, time.Time{}, 1)
	share2 := types.NewShare(1, 1, "address2", false, time.Time{}, 1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share2)

	reward, err := suite.app.RewardKeeper.CalculateRewardClaimPeriodRewards(suite.ctx, sdk.NewInt64Coin("nhash", 20), distribution)
	suite.Assert().NoError(err, "should return no error")
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 40), reward, "should distribute only up to the maximum reward for each participant")
}

func (suite *KeeperTestSuite) TestCalculateRewardClaimPeriodRewardsUsesDoesNotExceedProgramBalance() {
	suite.SetupTest()
	suite.Assert().Fail("not yet implemented")
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
	programBalance := types.NewRewardProgramBalance(program.GetId(), program.GetTotalRewardPool())
	suite.app.RewardKeeper.SetRewardProgramBalance(suite.ctx, programBalance)

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
	programBalance := types.NewRewardProgramBalance(program.GetId(), program.GetTotalRewardPool())
	suite.app.RewardKeeper.SetRewardProgramBalance(suite.ctx, programBalance)

	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &program)
	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)
	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)
	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)

	suite.Assert().Equal(program.State, types.RewardProgram_FINISHED, "reward program should be in finished state")
	suite.Assert().Equal(uint64(3), program.CurrentClaimPeriod, "current claim period should not be updated")
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
	programBalance := types.NewRewardProgramBalance(program.GetId(), program.GetTotalRewardPool())
	suite.app.RewardKeeper.SetRewardProgramBalance(suite.ctx, programBalance)

	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &program)
	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)
	// Normally you would need an additional claim period. However, it should end because the expected time is set.
	program.ExpectedProgramEndTime = currentTime
	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)

	suite.Assert().Equal(types.RewardProgram_FINISHED, program.State, "reward program should be in finished state")
	suite.Assert().Equal(uint64(2), program.CurrentClaimPeriod, "current claim period should not be updated")
	suite.Assert().Equal(blockTime.Add(time.Duration(program.ClaimPeriodSeconds)*time.Second), program.ClaimPeriodEndTime, "claim period end time should be set")
	suite.Assert().Equal(blockTime, program.ActualProgramEndTime, "claim period end time should be set")
}

func (suite *KeeperTestSuite) TestRewardProgramClaimPeriodEndNoBalance() {
	suite.SetupTest()

	currentTime := time.Now()
	suite.ctx = suite.ctx.WithBlockTime(currentTime)
	blockTime := suite.ctx.BlockTime()
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 0),
		sdk.NewInt64Coin("nhash", 1000),
		currentTime,
		60*60,
		3,
		[]types.QualifyingAction{},
	)
	programBalance := types.NewRewardProgramBalance(program.GetId(), program.GetTotalRewardPool())
	suite.app.RewardKeeper.SetRewardProgramBalance(suite.ctx, programBalance)

	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &program)
	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)

	suite.Assert().Equal(types.RewardProgram_FINISHED, program.State, "reward program should be in finished state")
	suite.Assert().Equal(uint64(1), program.CurrentClaimPeriod, "current claim period should not be updated")
	suite.Assert().Equal(program.ClaimPeriodEndTime, program.ClaimPeriodEndTime, "claim period end time should not be updated")
	suite.Assert().Equal(blockTime, program.ActualProgramEndTime, "actual end time should be set")
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

/*func (suite *KeeperTestSuite) TestCleanup() {
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
}*/

/*func (suite *KeeperTestSuite) TestUpdate() {
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
*/
