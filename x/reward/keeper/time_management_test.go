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
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program.Balance = program.GetTotalRewardPool()

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
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)

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
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program.Balance = program.GetTotalRewardPool()

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
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program.Balance = sdk.NewInt64Coin("nhash", 20)

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
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)

	suite.app.RewardKeeper.EndRewardProgram(suite.ctx, &program)

	suite.Assert().Equal(program.State, types.RewardProgram_FINISHED, "reward program should be in finished state")
	suite.Assert().Equal(blockTime, program.ActualProgramEndTime, "actual program end time should be set")
}

func (suite *KeeperTestSuite) TestEndRewardProgramNil() {
	suite.SetupTest()
	err := suite.app.RewardKeeper.EndRewardProgram(suite.ctx, nil)
	suite.Assert().Error(err, "should throw an error for nil")
}

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

	share1 := types.NewShare(1, 1, "address1", 1)
	share2 := types.NewShare(1, 1, "address2", 1)
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

	share1 := types.NewShare(1, 1, "address1", 1)
	share2 := types.NewShare(1, 1, "address2", 1)
	share3 := types.NewShare(1, 1, "address3", 1)
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

	share1 := types.NewShare(1, 1, "address1", 2)
	share2 := types.NewShare(1, 1, "address2", 1)
	share3 := types.NewShare(1, 1, "address3", 1)
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

	share1 := types.NewShare(1, 1, "address1", 1)
	share2 := types.NewShare(1, 1, "address2", 1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share2)

	reward, err := suite.app.RewardKeeper.CalculateRewardClaimPeriodRewards(suite.ctx, sdk.NewInt64Coin("nhash", 20), distribution)
	suite.Assert().NoError(err, "should return no error")
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 40), reward, "should distribute only up to the maximum reward for each participant")
}

func (suite *KeeperTestSuite) TestCalculateParticipantReward() {
	suite.SetupTest()
	reward := suite.app.RewardKeeper.CalculateParticipantReward(suite.ctx, 1, 2, sdk.NewInt64Coin("nhash", 100))
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 50), reward, "should get correct cut of pool")
}

func (suite *KeeperTestSuite) TestCalculateParticipantRewardCanHandleZeroTotalShares() {
	suite.SetupTest()
	reward := suite.app.RewardKeeper.CalculateParticipantReward(suite.ctx, 1, 0, sdk.NewInt64Coin("nhash", 100))
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 0), reward, "should have no reward")
}

func (suite *KeeperTestSuite) TestCalculateParticipantRewardCanHandleZeroEarnedShares() {
	suite.SetupTest()
	reward := suite.app.RewardKeeper.CalculateParticipantReward(suite.ctx, 0, 10, sdk.NewInt64Coin("nhash", 100))
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 0), reward, "should have no reward")
}

func (suite *KeeperTestSuite) TestCalculateParticipantRewardCanHandleZeroRewardPool() {
	suite.SetupTest()
	reward := suite.app.RewardKeeper.CalculateParticipantReward(suite.ctx, 1, 1, sdk.NewInt64Coin("nhash", 0))
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 0), reward, "should have no reward")
}

func (suite *KeeperTestSuite) TestCalculateParticipantRewardTruncates() {
	suite.SetupTest()

	reward := suite.app.RewardKeeper.CalculateParticipantReward(suite.ctx, 1, 3, sdk.NewInt64Coin("nhash", 100))
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 33), reward, "reward should truncate when < .5")

	reward = suite.app.RewardKeeper.CalculateParticipantReward(suite.ctx, 2, 3, sdk.NewInt64Coin("nhash", 100))
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 66), reward, "reward should truncate when >= .5")
}

func (suite *KeeperTestSuite) TestEndRewardProgramClaimPeriodHandlesInvalidLookups() {
	suite.SetupTest()

	currentTime := time.Now()
	suite.ctx = suite.ctx.WithBlockTime(currentTime)

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
		[]types.QualifyingAction{},
	)
	program1.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program2.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program3.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program2.Balance = program2.GetTotalRewardPool()
	program3.Balance = sdk.NewInt64Coin("jackthecat", 100)
	rewardDistribution := types.NewClaimPeriodRewardDistribution(0, 3, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, false)
	suite.app.RewardKeeper.SetClaimPeriodRewardDistribution(suite.ctx, rewardDistribution)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program1)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program2)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program3)

	err := suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program1)
	suite.Assert().Error(err, "an error should be thrown if there is no program balance for the program")
	err = suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program2)
	suite.Assert().Error(err, "an error should be thrown if there is no claim period reward distribution for the program")
	err = suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program3)
	suite.Assert().Error(err, "an error should be thrown if reward claim calculation fails")
}

func (suite *KeeperTestSuite) TestEndRewardProgramClaimPeriodHandlesNilRewardProgram() {
	suite.SetupTest()
	err := suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, nil)
	suite.Assert().Error(err, "error should be returned for nil reward program")
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
		sdk.NewInt64Coin("nhash", 100000),
		currentTime,
		60*60,
		2,
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	share := types.NewShare(1, 1, "address1", 1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share)
	program.Balance = program.GetTotalRewardPool()

	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &program)

	// Update the distribution to replicate that a share was actually granted.
	rewardDistribution, _ := suite.app.RewardKeeper.GetClaimPeriodRewardDistribution(suite.ctx, 1, 1)
	rewardDistribution.TotalShares = 1
	suite.app.RewardKeeper.SetClaimPeriodRewardDistribution(suite.ctx, rewardDistribution)

	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)

	reward, _ := suite.app.RewardKeeper.GetClaimPeriodRewardDistribution(suite.ctx, 1, 1)

	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 50000), program.Balance, "balance should subtract the claim period reward")
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 50000), reward.TotalRewardsPoolForClaimPeriod, "total claim should be increased by the amount rewarded")
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
		sdk.NewInt64Coin("nhash", 100000),
		currentTime,
		60*60,
		2,
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program.Balance = program.GetTotalRewardPool()

	share1 := types.NewShare(1, 1, "address1", 1)
	share2 := types.NewShare(1, 2, "address2", 1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share2)

	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &program)
	reward, _ := suite.app.RewardKeeper.GetClaimPeriodRewardDistribution(suite.ctx, 1, 1)
	reward.TotalShares = 1
	suite.app.RewardKeeper.SetClaimPeriodRewardDistribution(suite.ctx, reward)
	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)
	reward, _ = suite.app.RewardKeeper.GetClaimPeriodRewardDistribution(suite.ctx, 2, 1)
	reward.TotalShares = 1
	suite.app.RewardKeeper.SetClaimPeriodRewardDistribution(suite.ctx, reward)
	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)

	suite.Assert().Equal(program.State, types.RewardProgram_FINISHED, "reward program should be in finished state")
	suite.Assert().Equal(uint64(2), program.CurrentClaimPeriod, "current claim period should not be updated")
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
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program.Balance = program.GetTotalRewardPool()

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
		0,
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program.Balance = program.GetTotalRewardPool()

	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &program)
	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)

	suite.Assert().Equal(types.RewardProgram_FINISHED, program.State, "reward program should be in finished state")
	suite.Assert().Equal(uint64(1), program.CurrentClaimPeriod, "current claim period should not be updated")
	suite.Assert().Equal(program.ClaimPeriodEndTime, program.ClaimPeriodEndTime, "claim period end time should not be updated")
	suite.Assert().Equal(blockTime, program.ActualProgramEndTime, "actual end time should be set")
}

func (suite *KeeperTestSuite) TestRewardProgramClaimPeriodEndExtraBalance() {
	suite.SetupTest()

	currentTime := time.Now()
	suite.ctx = suite.ctx.WithBlockTime(currentTime)
	blockTime := suite.ctx.BlockTime()
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
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program.Balance = program.GetTotalRewardPool()

	share1 := types.NewShare(1, 1, "address1", 1)
	share2 := types.NewShare(1, 2, "address1", 1)
	share3 := types.NewShare(1, 3, "address1", 1)
	share4 := types.NewShare(1, 4, "address1", 1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share2)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share3)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share4)

	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &program)
	reward, _ := suite.app.RewardKeeper.GetClaimPeriodRewardDistribution(suite.ctx, 1, 1)
	reward.TotalShares = 1
	suite.app.RewardKeeper.SetClaimPeriodRewardDistribution(suite.ctx, reward)

	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)
	reward, _ = suite.app.RewardKeeper.GetClaimPeriodRewardDistribution(suite.ctx, 2, 1)
	reward.TotalShares = 1
	suite.app.RewardKeeper.SetClaimPeriodRewardDistribution(suite.ctx, reward)

	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)
	reward, _ = suite.app.RewardKeeper.GetClaimPeriodRewardDistribution(suite.ctx, 3, 1)
	reward.TotalShares = 1
	suite.app.RewardKeeper.SetClaimPeriodRewardDistribution(suite.ctx, reward)

	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)
	reward, _ = suite.app.RewardKeeper.GetClaimPeriodRewardDistribution(suite.ctx, 4, 1)
	reward.TotalShares = 1
	suite.app.RewardKeeper.SetClaimPeriodRewardDistribution(suite.ctx, reward)

	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)

	suite.Assert().Equal(types.RewardProgram_FINISHED, program.State, "reward program should be in finished state")
	suite.Assert().Equal(uint64(4), program.CurrentClaimPeriod, "an extra iteration should run")
	suite.Assert().Equal(blockTime, program.ActualProgramEndTime, "actual end time should be set")
}

func (suite *KeeperTestSuite) TestEndRewardProgramClaimPeriodUpdatesBalances() {
	suite.SetupTest()

	currentTime := time.Now()
	suite.ctx = suite.ctx.WithBlockTime(currentTime)
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
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	program.Balance = program.GetTotalRewardPool()

	share1 := types.NewShare(1, 1, "address1", 1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share1)

	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &program)
	reward, _ := suite.app.RewardKeeper.GetClaimPeriodRewardDistribution(suite.ctx, 1, 1)
	reward.TotalShares = 1
	suite.app.RewardKeeper.SetClaimPeriodRewardDistribution(suite.ctx, reward)
	claimAmount, _ := suite.app.RewardKeeper.CalculateRewardClaimPeriodRewards(suite.ctx, program.GetMaxRewardByAddress(), reward)
	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)

	// Adjusted after ending period
	reward, _ = suite.app.RewardKeeper.GetClaimPeriodRewardDistribution(suite.ctx, 1, 1)
	expectedProgramBalance := program.GetTotalRewardPool().Sub(claimAmount)
	suite.Assert().Equal(claimAmount, reward.GetTotalRewardsPoolForClaimPeriod(), "the reward for the claim period should be added to total reward")
	suite.Assert().Equal(expectedProgramBalance, program.GetBalance(), "the reward for the claim period should be subtracted out of the program balance")
	suite.Assert().Equal(types.RewardProgram_STARTED, program.State, "reward program should be in started state")
	suite.Assert().Equal(uint64(2), program.CurrentClaimPeriod, "next iteration should start")
}

func (suite *KeeperTestSuite) TestEndRewardProgramClaimPeriodHandlesMinimumRolloverAmount() {
	suite.SetupTest()

	currentTime := time.Now()
	suite.ctx = suite.ctx.WithBlockTime(currentTime)
	blockTime := suite.ctx.BlockTime()
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
		[]types.QualifyingAction{},
	)
	program.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 501)
	program.Balance = program.GetTotalRewardPool()

	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &program)

	// Create the shares
	share1 := types.NewShare(1, 1, "address1", 1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share1)
	reward, _ := suite.app.RewardKeeper.GetClaimPeriodRewardDistribution(suite.ctx, 1, 1)
	reward.TotalShares = 1
	suite.app.RewardKeeper.SetClaimPeriodRewardDistribution(suite.ctx, reward)

	// Should end because the balance should be below 501
	suite.app.RewardKeeper.EndRewardProgramClaimPeriod(suite.ctx, &program)

	suite.Assert().Equal(types.RewardProgram_FINISHED, program.State, "reward program should be in finished state")
	suite.Assert().Equal(uint64(1), program.CurrentClaimPeriod, "current claim period should not be updated")
	suite.Assert().Equal(program.ClaimPeriodEndTime, program.ClaimPeriodEndTime, "claim period end time should not be updated")
	suite.Assert().Equal(blockTime, program.ActualProgramEndTime, "actual end time should be set")
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 500), program.GetBalance(), "balance should be updated")
}

func (suite *KeeperTestSuite) TestUpdate() {
	suite.SetupTest()
	// Reward Program that has not started
	currentTime := time.Now()
	suite.ctx = suite.ctx.WithBlockTime(currentTime)
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
		0,
		[]types.QualifyingAction{},
	)
	notStarted.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	notStarted.Balance = notStarted.GetTotalRewardPool()

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
		[]types.QualifyingAction{},
	)
	starting.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	starting.Balance = starting.GetTotalRewardPool()

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
		[]types.QualifyingAction{},
	)
	nextClaimPeriod.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	nextClaimPeriod.Balance = nextClaimPeriod.GetTotalRewardPool()
	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &nextClaimPeriod)
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
		[]types.QualifyingAction{},
	)
	ending.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	ending.Balance = sdk.NewInt64Coin("nhash", 0)
	share1 := types.NewShare(4, 1, "address1", 1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share1)
	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &ending)
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
		[]types.QualifyingAction{},
	)
	timeout.MinimumRolloverAmount = sdk.NewInt64Coin("nhash", 1)
	timeout.ClaimPeriodEndTime = blockTime
	timeout.ExpectedProgramEndTime = blockTime
	timeout.Balance = timeout.GetTotalRewardPool()
	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &timeout)

	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, notStarted)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, starting)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, nextClaimPeriod)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, ending)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, timeout)

	// We call update
	suite.app.RewardKeeper.Update(suite.ctx)

	notStarted, _ = suite.app.RewardKeeper.GetRewardProgram(suite.ctx, notStarted.Id)
	starting, _ = suite.app.RewardKeeper.GetRewardProgram(suite.ctx, starting.Id)
	nextClaimPeriod, _ = suite.app.RewardKeeper.GetRewardProgram(suite.ctx, nextClaimPeriod.Id)
	ending, _ = suite.app.RewardKeeper.GetRewardProgram(suite.ctx, ending.Id)
	timeout, _ = suite.app.RewardKeeper.GetRewardProgram(suite.ctx, timeout.Id)

	suite.Assert().Equal(uint64(0), notStarted.CurrentClaimPeriod, "claim period should be 0 for a program that is not started")
	suite.Assert().Equal(notStarted.State, types.RewardProgram_PENDING, "should be in pending state")

	suite.Assert().Equal(uint64(1), starting.CurrentClaimPeriod, "claim period should be 1 for a program that just started")
	suite.Assert().Equal(starting.State, types.RewardProgram_STARTED, "should be in started state")

	suite.Assert().Equal(uint64(2), nextClaimPeriod.CurrentClaimPeriod, "claim period should be 2 for a program that went to next claim period")
	suite.Assert().Equal(nextClaimPeriod.State, types.RewardProgram_STARTED, "should be in started state")

	suite.Assert().Equal(uint64(1), ending.CurrentClaimPeriod, "claim period should not increment")
	suite.Assert().Equal(ending.State, types.RewardProgram_FINISHED, "should be in finished state")

	suite.Assert().Equal(uint64(1), timeout.CurrentClaimPeriod, "claim period should not increment")
	suite.Assert().Equal(timeout.State, types.RewardProgram_FINISHED, "should be in finished state")
}
