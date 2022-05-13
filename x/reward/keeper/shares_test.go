package keeper_test

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (suite *KeeperTestSuite) TestNewShare() {
	suite.SetupTest()

	time := timestamppb.Now().AsTime()

	share := types.NewShare(
		1,
		"test",
		true,
		time,
		5,
	)

	suite.Assert().Equal(uint64(1), share.GetRewardProgramId(), "reward program id must match")
	suite.Assert().Equal("test", share.GetAddress(), "address must match")
	suite.Assert().Equal(true, share.GetClaimed(), "claim status must match")
	suite.Assert().Equal(time, share.GetExpireTime(), "expiration time must match")
	suite.Assert().Equal(int64(5), share.GetAmount(), "share amount must match")
}

func (suite *KeeperTestSuite) TestGetSetShare() {
	suite.Assert().Fail("not implemented")
}

func (suite *KeeperTestSuite) TestGetInvalidShare() {
	suite.Assert().Fail("not implemented")
}

func (suite *KeeperTestSuite) TestRemoveValidShare() {
	suite.Assert().Fail("not implemented")
}

func (suite *KeeperTestSuite) TestIterateShares() {
	suite.Assert().Fail("not implemented")
}

func (suite *KeeperTestSuite) TestIterateSharesHalt() {
	suite.Assert().Fail("not implemented")
}

func (suite *KeeperTestSuite) TestEmptyIterateShares() {
	suite.Assert().Fail("not implemented")
}

func (suite *KeeperTestSuite) TestIterateRewardShares() {
	suite.Assert().Fail("not implemented")
}

func (suite *KeeperTestSuite) TestEmptyIterateRewardShares() {
	suite.Assert().Fail("not implemented")
}

func (suite *KeeperTestSuite) TestIterateRewardSharesHalt() {
	suite.Assert().Fail("not implemented")
}

func (suite *KeeperTestSuite) TestIterateRewardEpochShares() {
	suite.Assert().Fail("not implemented")
}

func (suite *KeeperTestSuite) TestEmptyIterateRewardEpochShares() {
	suite.Assert().Fail("not implemented")
}

func (suite *KeeperTestSuite) TestIterateRewardSharesEpochHalt() {
	suite.Assert().Fail("not implemented")
}
