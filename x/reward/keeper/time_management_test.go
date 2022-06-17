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

	suite.Assert().True(program.Started, "reward program should be in started state")
	suite.Assert().Equal(uint64(1), program.CurrentSubPeriod, "current sub period should be set to 1")
	suite.Assert().Equal(blockTime.Add(time.Duration(program.SubPeriodSeconds)*time.Second), program.SubPeriodEndTime, "sub period end time should be set")
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

	suite.Assert().True(program.Finished, "reward program should be in finished state")
	suite.Assert().Equal(blockTime, program.FinishedTime, "finished time should be set")
}

func (suite *KeeperTestSuite) TestRewardProgramSubPeriodEnd() {
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
	suite.app.RewardKeeper.EndRewardProgramSubPeriod(suite.ctx, &program)

	suite.Assert().False(program.Finished, "reward program should not be in finished state")
	suite.Assert().Equal(uint64(2), program.CurrentSubPeriod, "current sub period should be updated")
	suite.Assert().Equal(blockTime.Add(time.Duration(program.SubPeriodSeconds)*time.Second), program.SubPeriodEndTime, "sub period end time should be set")
}

func (suite *KeeperTestSuite) TestRewardProgramSubPeriodEndTransition() {
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
	suite.app.RewardKeeper.EndRewardProgramSubPeriod(suite.ctx, &program)
	suite.app.RewardKeeper.EndRewardProgramSubPeriod(suite.ctx, &program)
	suite.app.RewardKeeper.EndRewardProgramSubPeriod(suite.ctx, &program)

	suite.Assert().True(program.Finished, "reward program should not be in finished state")
	suite.Assert().Equal(uint64(4), program.CurrentSubPeriod, "current sub period should be updated")
	suite.Assert().Equal(blockTime.Add(time.Duration(program.SubPeriodSeconds)*time.Second), program.SubPeriodEndTime, "sub period end time should be set")
	suite.Assert().Equal(blockTime, program.FinishedTime, "sub period end time should be set")
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
	hasShares.Finished = true

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
	hasExpiredShares.Finished = true

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
	hasNoShares.Finished = true

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

	// Reward Program that is ready to move onto next sub period
	nextSubPeriod := types.NewRewardProgram(
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
	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &nextSubPeriod)
	nextSubPeriod.SubPeriodEndTime = blockTime

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
	ending.SubPeriodEndTime = blockTime

	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, notStarted)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, starting)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, nextSubPeriod)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, ending)

	// We call update
	suite.app.RewardKeeper.Update(suite.ctx)

	notStarted, _ = suite.app.RewardKeeper.GetRewardProgram(suite.ctx, notStarted.Id)
	starting, _ = suite.app.RewardKeeper.GetRewardProgram(suite.ctx, starting.Id)
	nextSubPeriod, _ = suite.app.RewardKeeper.GetRewardProgram(suite.ctx, nextSubPeriod.Id)
	ending, _ = suite.app.RewardKeeper.GetRewardProgram(suite.ctx, ending.Id)

	suite.Assert().Equal(uint64(0), notStarted.CurrentSubPeriod, "sub period should be 0 for a program that is not started")
	suite.Assert().False(notStarted.Started, "should not be in started state")
	suite.Assert().False(notStarted.Finished, "should not be in finished state")

	suite.Assert().Equal(uint64(1), starting.CurrentSubPeriod, "sub period should be 1 for a program that just started")
	suite.Assert().True(starting.Started, "should be in started state")
	suite.Assert().False(starting.Finished, "should not be in finished state")

	suite.Assert().Equal(uint64(2), nextSubPeriod.CurrentSubPeriod, "sub period should be 2 for a program that went to next sub period")
	suite.Assert().True(nextSubPeriod.Started, "should be in started state")
	suite.Assert().False(nextSubPeriod.Finished, "should not be in finished state")

	suite.Assert().Equal(uint64(2), ending.CurrentSubPeriod, "sub period should be incremented by 1")
	suite.Assert().True(ending.Started, "should be in started state")
	suite.Assert().True(ending.Finished, "should be in finished state")
}
