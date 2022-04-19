package keeper_test

import (
	gocontext "context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (s *KeeperTestSuite) TestQueryRewardPrograms() {
	s.SetupTest()
	queryClient := s.queryClient

	// Test against empty set
	response, err := queryClient.RewardPrograms(gocontext.Background(), &types.RewardProgramsRequest{})
	s.Assert().Nil(err, "Error should be nil")
	s.Assert().Equal(len(response.RewardPrograms), 0, "response should contain empty list")

	// Create the reward program
	action := types.NewActionDelegate()
	coin := sdk.NewInt64Coin("jackthecat", 10000)
	maxCoin := sdk.NewInt64Coin("jackthecat", 100)
	rewardProgram := types.NewRewardProgram(1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", coin, maxCoin, "day", 1, 10, types.NewEligibilityCriteria("criteria", &action), false, 1, 10)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

	// Test with 1 item
	response, err = queryClient.RewardPrograms(gocontext.Background(), &types.RewardProgramsRequest{})
	s.Assert().Nil(err, "Error should be nil")
	s.Assert().Equal(len(response.RewardPrograms), 1, "response should contain the added element")
	s.Assert().True(rewardProgram.Equal(response.RewardPrograms[0]), "reward programs should match")
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
