package keeper_test

import (
	"github.com/provenance-io/provenance/x/reward/types"
)

func (suite *KeeperTestSuite) TestNewShare() {
	suite.SetupTest()

	share := types.NewShare(
		1,
		2,
		"test",
		5,
	)

	suite.Assert().Equal(uint64(1), share.GetRewardProgramId(), "reward program id must match")
	suite.Assert().Equal(uint64(2), share.GetClaimPeriodId(), "reward claim period id must match")
	suite.Assert().Equal("test", share.GetAddress(), "address must match")
	suite.Assert().Equal(int64(5), share.GetAmount(), "share amount must match")
}

func (suite *KeeperTestSuite) TestGetSetShare() {
	suite.SetupTest()

	expectedShare := types.NewShare(
		1,
		2,
		"test",
		5,
	)

	suite.app.RewardKeeper.SetShare(suite.ctx, &expectedShare)
	actualShare, err := suite.app.RewardKeeper.GetShare(suite.ctx,
		expectedShare.GetRewardProgramId(),
		expectedShare.GetClaimPeriodId(),
		expectedShare.GetAddress())

	suite.Assert().Nil(err, "must not have error")
	suite.Assert().Equal(expectedShare.GetRewardProgramId(), actualShare.GetRewardProgramId(), "reward program id must match")
	suite.Assert().Equal(expectedShare.GetClaimPeriodId(), actualShare.GetClaimPeriodId(), "reward claim period id must match")
	suite.Assert().Equal(expectedShare.GetAddress(), actualShare.GetAddress(), "address must match")
	suite.Assert().Equal(expectedShare.GetAmount(), actualShare.GetAmount(), "share amount must match")
}

func (suite *KeeperTestSuite) TestGetInvalidShare() {
	suite.SetupTest()

	actualShare, err := suite.app.RewardKeeper.GetShare(suite.ctx,
		99,
		99,
		"jackthecat")

	suite.Assert().Nil(err, "must not have error")
	suite.Assert().Error(actualShare.ValidateBasic(), "share validate basic must return error")
}

func (suite *KeeperTestSuite) TestRemoveValidShare() {
	suite.SetupTest()

	expectedShare := types.NewShare(
		1,
		2,
		"test",
		5,
	)

	suite.app.RewardKeeper.SetShare(suite.ctx, &expectedShare)
	removed := suite.app.RewardKeeper.RemoveShare(suite.ctx,
		expectedShare.GetRewardProgramId(),
		expectedShare.GetClaimPeriodId(),
		expectedShare.GetAddress())

	actualShare, err := suite.app.RewardKeeper.GetShare(suite.ctx,
		expectedShare.GetRewardProgramId(),
		expectedShare.GetClaimPeriodId(),
		expectedShare.GetAddress())

	suite.Assert().True(removed, "share should successfully be removed")
	suite.Assert().Nil(err, "must not have error")
	suite.Assert().Error(actualShare.ValidateBasic(), "share validate basic must return error")
}

func (suite *KeeperTestSuite) TestRemoveInvalidShare() {
	suite.SetupTest()

	expectedShare := types.NewShare(
		1,
		2,
		"test",
		5,
	)

	removed := suite.app.RewardKeeper.RemoveShare(suite.ctx,
		expectedShare.GetRewardProgramId(),
		expectedShare.GetClaimPeriodId(),
		expectedShare.GetAddress())

	suite.Assert().False(removed, "share should be unable to be removed")
}

func (suite *KeeperTestSuite) TestIterateShares() {
	suite.SetupTest()

	share1 := types.NewShare(1, 2, "test", 5)
	share2 := types.NewShare(1, 3, "test", 5)
	share3 := types.NewShare(2, 1, "test", 5)
	share4 := types.NewShare(2, 2, "test", 5)
	share5 := types.NewShare(2, 2, "test2", 5)

	suite.app.RewardKeeper.SetShare(suite.ctx, &share1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share2)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share3)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share4)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share5)

	counter := 0
	suite.app.RewardKeeper.IterateShares(suite.ctx, func(share types.Share) bool {
		counter += 1
		return false
	})

	suite.Assert().Equal(5, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestIterateSharesHalt() {
	suite.SetupTest()

	share1 := types.NewShare(1, 2, "test", 5)
	share2 := types.NewShare(1, 3, "test", 5)
	share3 := types.NewShare(2, 1, "test", 5)
	share4 := types.NewShare(2, 2, "test", 5)
	share5 := types.NewShare(2, 2, "test2", 5)

	suite.app.RewardKeeper.SetShare(suite.ctx, &share1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share2)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share3)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share4)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share5)

	counter := 0
	suite.app.RewardKeeper.IterateShares(suite.ctx, func(share types.Share) bool {
		counter += 1

		return counter == 2
	})

	suite.Assert().Equal(2, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestEmptyIterateShares() {
	suite.SetupTest()

	counter := 0
	suite.app.RewardKeeper.IterateShares(suite.ctx, func(share types.Share) bool {
		counter += 1
		return false
	})

	suite.Assert().Equal(0, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestIterateRewardShares() {
	suite.SetupTest()

	share1 := types.NewShare(1, 2, "test", 5)
	share2 := types.NewShare(1, 3, "test", 5)
	share3 := types.NewShare(2, 1, "test", 5)
	share4 := types.NewShare(2, 2, "test", 5)
	share5 := types.NewShare(2, 2, "test2", 5)

	suite.app.RewardKeeper.SetShare(suite.ctx, &share1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share2)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share3)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share4)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share5)

	counter := 0
	suite.app.RewardKeeper.IterateRewardShares(suite.ctx, 1, func(share types.Share) bool {
		counter += 1
		return false
	})

	suite.Assert().Equal(2, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestEmptyIterateRewardShares() {
	suite.SetupTest()

	share3 := types.NewShare(2, 1, "test", 5)
	share4 := types.NewShare(2, 2, "test", 5)
	share5 := types.NewShare(2, 2, "test2", 5)

	suite.app.RewardKeeper.SetShare(suite.ctx, &share3)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share4)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share5)

	counter := 0
	suite.app.RewardKeeper.IterateRewardShares(suite.ctx, 1, func(share types.Share) bool {
		counter += 1
		return false
	})

	suite.Assert().Equal(0, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestIterateRewardSharesHalt() {
	suite.SetupTest()

	share1 := types.NewShare(1, 2, "test", 5)
	share2 := types.NewShare(1, 3, "test", 5)
	share3 := types.NewShare(2, 1, "test", 5)
	share4 := types.NewShare(2, 2, "test", 5)
	share5 := types.NewShare(2, 2, "test2", 5)

	suite.app.RewardKeeper.SetShare(suite.ctx, &share1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share2)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share3)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share4)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share5)

	counter := 0
	suite.app.RewardKeeper.IterateRewardShares(suite.ctx, 1, func(share types.Share) bool {
		counter += 1
		return counter == 1
	})

	suite.Assert().Equal(1, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestIterateRewardClaimPeriodShares() {
	suite.SetupTest()

	share1 := types.NewShare(1, 2, "test", 5)
	share2 := types.NewShare(1, 3, "test", 5)
	share3 := types.NewShare(2, 1, "test", 5)
	share4 := types.NewShare(2, 2, "test", 5)
	share5 := types.NewShare(2, 2, "test2", 5)

	suite.app.RewardKeeper.SetShare(suite.ctx, &share1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share2)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share3)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share4)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share5)

	counter := 0
	suite.app.RewardKeeper.IterateRewardClaimPeriodShares(suite.ctx, 2, 2, func(share types.Share) bool {
		counter += 1
		return false
	})

	suite.Assert().Equal(2, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestEmptyIterateRewardClaimPeriodShares() {
	suite.SetupTest()

	share1 := types.NewShare(1, 2, "test", 5)
	share2 := types.NewShare(1, 3, "test", 5)
	share3 := types.NewShare(2, 1, "test", 5)
	share4 := types.NewShare(2, 2, "test", 5)
	share5 := types.NewShare(2, 2, "test2", 5)

	suite.app.RewardKeeper.SetShare(suite.ctx, &share1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share2)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share3)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share4)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share5)

	counter := 0
	suite.app.RewardKeeper.IterateRewardClaimPeriodShares(suite.ctx, 1, 4, func(share types.Share) bool {
		counter += 1
		return false
	})

	suite.Assert().Equal(0, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestIterateRewardSharesClaimPeriodHalt() {
	suite.SetupTest()

	share1 := types.NewShare(1, 2, "test", 5)
	share2 := types.NewShare(1, 3, "test", 5)
	share3 := types.NewShare(2, 1, "test", 5)
	share4 := types.NewShare(2, 2, "test", 5)
	share5 := types.NewShare(2, 2, "test2", 5)

	suite.app.RewardKeeper.SetShare(suite.ctx, &share1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share2)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share3)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share4)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share5)

	counter := 0
	suite.app.RewardKeeper.IterateRewardClaimPeriodShares(suite.ctx, 1, 2, func(share types.Share) bool {
		counter += 1
		return counter == 1
	})

	suite.Assert().Equal(1, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestGetRewardClaimPeriodShares() {
	suite.SetupTest()

	share1 := types.NewShare(1, 2, "test", 5)
	share2 := types.NewShare(1, 3, "test", 5)
	share3 := types.NewShare(2, 1, "test", 5)
	share4 := types.NewShare(2, 2, "test", 5)
	share5 := types.NewShare(2, 2, "test2", 5)

	suite.app.RewardKeeper.SetShare(suite.ctx, &share1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share2)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share3)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share4)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share5)

	shares, err := suite.app.RewardKeeper.GetRewardClaimPeriodShares(suite.ctx, 2, 2)

	suite.Assert().NoError(err, "should have no error")
	suite.Assert().Equal(2, len(shares), "should have correct number of shares")
}

func (suite *KeeperTestSuite) TestGetRewardClaimPeriodSharesEmpty() {
	suite.SetupTest()

	share1 := types.NewShare(1, 2, "test", 5)
	share2 := types.NewShare(1, 3, "test", 5)
	share3 := types.NewShare(2, 1, "test", 5)
	share4 := types.NewShare(2, 2, "test", 5)
	share5 := types.NewShare(2, 2, "test2", 5)

	suite.app.RewardKeeper.SetShare(suite.ctx, &share1)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share2)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share3)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share4)
	suite.app.RewardKeeper.SetShare(suite.ctx, &share5)

	shares, err := suite.app.RewardKeeper.GetRewardClaimPeriodShares(suite.ctx, 5, 5)

	suite.Assert().NoError(err, "should have no error")
	suite.Assert().Equal(0, len(shares), "should have correct number of shares")
}
