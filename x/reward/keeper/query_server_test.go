package keeper_test

import (
	gocontext "context"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (suite *KeeperTestSuite) TestQueryRewardPrograms() {
	suite.SetupTest()
	queryClient := suite.queryClient

	queryClient.RewardPrograms(gocontext.Background(), &types.RewardProgramsRequest{})
	suite.Require().Fail("Not yet implemented")
}

func (suite *KeeperTestSuite) TestActiveRewardPrograms() {
	suite.SetupTest()
	queryClient := suite.queryClient

	queryClient.ActiveRewardPrograms(gocontext.Background(), &types.ActiveRewardProgramsRequest{})
	suite.Require().Fail("Not yet implemented")
}

func (suite *KeeperTestSuite) TestModuleAccountBalance() {
	suite.SetupTest()
	queryClient := suite.queryClient

	queryClient.ModuleAccountBalance(gocontext.Background(), &types.QueryModuleAccountBalanceRequest{})
	suite.Require().Fail("Not yet implemented")
}

func (suite *KeeperTestSuite) TestRewardProgramByID() {
	suite.SetupTest()
	queryClient := suite.queryClient

	queryClient.RewardProgramByID(gocontext.Background(), &types.RewardProgramByIDRequest{})
	suite.Require().Fail("Not yet implemented")
}
