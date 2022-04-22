package keeper_test

import (
	gocontext "context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	epoch "github.com/provenance-io/provenance/x/epoch"
	epochTypes "github.com/provenance-io/provenance/x/epoch/types"
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
	rewardProgram := newRewardProgram("jackthecat", 1, 1)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, *rewardProgram)

	// Test with 1 item
	response, err = queryClient.RewardPrograms(gocontext.Background(), &types.RewardProgramsRequest{})
	s.Assert().Nil(err, "Error should be nil")
	s.Assert().Equal(len(response.RewardPrograms), 1, "response should contain the added element")
	s.Assert().True(rewardProgram.Equal(response.RewardPrograms[0]), "reward programs should match")

	// Test with 2 items
	rewardProgram2 := newRewardProgram("catthejack", 2, 1)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, *rewardProgram2)

	response, err = queryClient.RewardPrograms(gocontext.Background(), &types.RewardProgramsRequest{})
	s.Assert().Nil(err, "Error should be nil")
	s.Assert().Equal(len(response.RewardPrograms), 2, "response should contain the added element")
	s.Assert().True(rewardProgram2.Equal(response.RewardPrograms[1]), "reward programs should match")
}

func (s *KeeperTestSuite) TestActiveRewardPrograms() {
	s.SetupTest()
	queryClient := s.queryClient

	// Setup
	initialBlockHeight := int64(1)
	s.ctx = s.ctx.WithBlockHeight(initialBlockHeight)
	setupEpoch(s, uint64(initialBlockHeight))

	// Check against no reward programs
	resp, err := queryClient.ActiveRewardPrograms(gocontext.Background(), &types.ActiveRewardProgramsRequest{})
	s.Assert().Nil(err)
	s.Assert().Equal(0, len(resp.GetRewardPrograms()), "no reward programs returned when none exist.")

	// Create the reward programs
	rewardProgram := newRewardProgram("jackthecat", 1, 1)
	rewardProgram2 := newRewardProgram("catthejack", 2, 3)
	rewardProgram3 := newRewardProgram("thejackcat", 3, 1)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, *rewardProgram)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, *rewardProgram2)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, *rewardProgram3)

	// Check against ACTIVE reward programs
	resp, err = queryClient.ActiveRewardPrograms(gocontext.Background(), &types.ActiveRewardProgramsRequest{})
	s.Assert().Nil(err)
	s.Assert().Equal(2, len(resp.GetRewardPrograms()), "must contain all active reward programs")
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
	rewardProgram := newRewardProgram("jackthecat", 1, 1)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, *rewardProgram)
	rewardProgram2 := newRewardProgram("catthejack", 2, 1)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, *rewardProgram2)

	// Tests that it can get the correct reward
	response, err = queryClient.RewardProgramByID(gocontext.Background(), &request)
	s.Assert().NotNil(response.RewardProgram, "response should not be nil")
	s.Assert().Nil(err, "error should be nil")
	s.Assert().True(rewardProgram.Equal(*response.RewardProgram), "reward programs should match")
}

func (s *KeeperTestSuite) TestRewardClaims() {
	s.Assert().Fail("not implemented")
}

func (s *KeeperTestSuite) TestRewardClaimById() {
	s.Assert().Fail("not implemented")
}

func newRewardProgram(coinName string, id uint64, start uint64) *types.RewardProgram {
	action := types.NewActionDelegate()
	coin := sdk.NewInt64Coin(coinName, 10000)
	maxCoin := sdk.NewInt64Coin(coinName, 100)

	rewardProgram := types.NewRewardProgram(id, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", coin, maxCoin, "day", start, 2, types.NewEligibilityCriteria("criteria", &action), false, 1, 10)

	return &rewardProgram
}

func setupEpoch(s *KeeperTestSuite, height uint64) {
	epochInfos := s.app.EpochKeeper.AllEpochInfos(s.ctx)
	for _, epochInfo := range epochInfos {
		s.app.EpochKeeper.DeleteEpochInfo(s.ctx, epochInfo.Identifier)
	}

	epoch.InitGenesis(s.ctx, s.app.EpochKeeper, epochTypes.GenesisState{
		Epochs: []epochTypes.EpochInfo{
			{
				Identifier:              "day",
				StartHeight:             uint64(height),
				Duration:                (60 * 60 * 24) / 5,
				CurrentEpoch:            2,
				CurrentEpochStartHeight: uint64(height),
				EpochCountingStarted:    false,
			},
		},
	})
}
