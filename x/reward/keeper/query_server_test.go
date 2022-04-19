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
	rewardProgram := newRewardProgram("jackthecat", 1)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, *rewardProgram)

	// Test with 1 item
	response, err = queryClient.RewardPrograms(gocontext.Background(), &types.RewardProgramsRequest{})
	s.Assert().Nil(err, "Error should be nil")
	s.Assert().Equal(len(response.RewardPrograms), 1, "response should contain the added element")
	s.Assert().True(rewardProgram.Equal(response.RewardPrograms[0]), "reward programs should match")

	// Test with 2 items
	rewardProgram2 := newRewardProgram("catthejack", 2)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, *rewardProgram2)

	response, err = queryClient.RewardPrograms(gocontext.Background(), &types.RewardProgramsRequest{})
	s.Assert().Nil(err, "Error should be nil")
	s.Assert().Equal(len(response.RewardPrograms), 2, "response should contain the added element")
	s.Assert().True(rewardProgram2.Equal(response.RewardPrograms[1]), "reward programs should match")
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

func (s *KeeperTestSuite) TestRewardProgramByID() {
	s.SetupTest()
	queryClient := s.queryClient

	request := types.RewardProgramByIDRequest{
		Id: 1,
	}

	// Test on reward program that doesn't exist
	response, err := queryClient.RewardProgramByID(gocontext.Background(), &request)
	s.Assert().NotNil(response, "response should not be nil")
	s.Assert().Nil(response.RewardProgram, "A reward program that does not exist should be nil")
	s.Assert().Nil(err, "error should be nil")

	// Create the reward programs
	rewardProgram := newRewardProgram("jackthecat", 1)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, *rewardProgram)
	rewardProgram2 := newRewardProgram("catthejack", 2)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, *rewardProgram2)

	// Tests that it can get the correct reward
	response, err = queryClient.RewardProgramByID(gocontext.Background(), &request)
	s.Assert().NotNil(response.RewardProgram, "response should not be nil")
	s.Assert().Nil(err, "error should be nil")
	s.Assert().True(rewardProgram.Equal(*response.RewardProgram), "reward programs should match")
}

func newRewardProgram(coinName string, id uint64) *types.RewardProgram {
	action := types.NewActionDelegate()
	coin := sdk.NewInt64Coin(coinName, 10000)
	maxCoin := sdk.NewInt64Coin(coinName, 100)

	rewardProgram := types.NewRewardProgram(id, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", coin, maxCoin, "day", 1, 10, types.NewEligibilityCriteria("criteria", &action), false, 1, 10)

	return &rewardProgram
}
