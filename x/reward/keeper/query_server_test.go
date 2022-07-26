package keeper_test

/*
func (s *KeeperTestSuite) TestQueryRewardPrograms() {
	s.SetupTest()
	queryClient := s.queryClient

	// Test against empty set
	response, err := queryClient.RewardPrograms(gocontext.Background(), &types.QueryRewardProgramsRequest{})
	s.Assert().Nil(err, "Error should be nil")
	s.Assert().Equal(len(response.RewardPrograms), 0, "response should contain empty list")

	// Create the reward program
	rewardProgram := newRewardProgram("jackthecat", 1, 1)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, *rewardProgram)

	// Test with 1 item
	response, err = queryClient.RewardPrograms(gocontext.Background(), &types.QueryRewardProgramsRequest{})
	s.Assert().Nil(err, "Error should be nil")
	s.Assert().Equal(len(response.RewardPrograms), 1, "response should contain the added element")
	s.Assert().True(rewardProgram.Equal(response.RewardPrograms[0]), "reward programs should match")

	// Test with 2 items
	rewardProgram2 := newRewardProgram("catthejack", 2, 1)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, *rewardProgram2)

	response, err = queryClient.RewardPrograms(gocontext.Background(), &types.QueryRewardProgramsRequest{})
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
	resp, err := queryClient.ActiveRewardPrograms(gocontext.Background(), &types.ActiveQueryRewardProgramsRequest{})
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
	resp, err = queryClient.ActiveRewardPrograms(gocontext.Background(), &types.ActiveQueryRewardProgramsRequest{})
	s.Assert().Nil(err)
	s.Assert().Equal(2, len(resp.GetRewardPrograms()), "must contain all active reward programs")
}

func (suite *KeeperTestSuite) TestModuleAccountBalance() {
	suite.SetupTest()
	queryClient := suite.queryClient

	queryClient.ModuleAccountBalance(gocontext.Background(), &types.QueryModuleAccountBalanceRequest{})
	suite.Require().Fail("not implemented")
}

func (s *KeeperTestSuite) TestRewardProgramByID() {
	s.SetupTest()
	queryClient := s.queryClient

	request := types.QueryRewardProgramByIDRequest{
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

func (s *KeeperTestSuite) TestEpochDistributionRewardsByID() {
	s.SetupTest()
	queryClient := s.queryClient

	// Test against empty set
	response, err := queryClient.EpochRewardDistributionsByID(gocontext.Background(), &types.EpochRewardDistributionByIDRequest{
		RewardId: 1,
		EpochId:  "day",
	})
	s.Assert().Nil(err, "error should be nil")
	s.Assert().Nil(response.GetEpochRewardDistribution(), "epoch reward distribution should be nil")

	// Test single entry
	rewardDistribution := types.NewEpochRewardDistribution("day", 1, sdk.NewInt64Coin("jackthecat", 100), 5, false)
	s.app.RewardKeeper.SetEpochRewardDistribution(s.ctx, rewardDistribution)
	response, err = queryClient.EpochRewardDistributionsByID(gocontext.Background(), &types.EpochRewardDistributionByIDRequest{
		RewardId: 1,
		EpochId:  "day",
	})
	s.Assert().Nil(err)
	s.Assert().NotNil(response.GetEpochRewardDistribution(), "epoch reward distribution should not be nil")
	s.Assert().Equal(rewardDistribution.EpochId, response.GetEpochRewardDistribution().EpochId, "epoch ids should match")
	s.Assert().Equal(rewardDistribution.RewardProgramId, response.GetEpochRewardDistribution().RewardProgramId, "reward ids should match")
	s.Assert().Equal(rewardDistribution.TotalRewardsPool, response.GetEpochRewardDistribution().TotalRewardsPool, "total rewards pool should match")
	s.Assert().Equal(rewardDistribution.TotalShares, response.GetEpochRewardDistribution().TotalShares, "total shares should match")
	s.Assert().False(response.GetEpochRewardDistribution().EpochEnded, "epoch should not be ended")

	// Test multiple entry
	rewardDistribution2 := types.NewEpochRewardDistribution("day", 2, sdk.NewInt64Coin("jackthecat", 200), 10, false)
	rewardDistribution3 := types.NewEpochRewardDistribution("day", 3, sdk.NewInt64Coin("jackthecat", 300), 15, false)
	s.app.RewardKeeper.SetEpochRewardDistribution(s.ctx, rewardDistribution2)
	s.app.RewardKeeper.SetEpochRewardDistribution(s.ctx, rewardDistribution3)
	response, err = queryClient.EpochRewardDistributionsByID(gocontext.Background(), &types.EpochRewardDistributionByIDRequest{
		RewardId: 2,
		EpochId:  "day",
	})
	s.Assert().Nil(err)
	s.Assert().NotNil(response.GetEpochRewardDistribution(), "epoch reward distribution should not be nil")
	s.Assert().Equal(rewardDistribution2.EpochId, response.GetEpochRewardDistribution().EpochId, "epoch ids should match")
	s.Assert().Equal(rewardDistribution2.RewardProgramId, response.GetEpochRewardDistribution().RewardProgramId, "reward ids should match")
	s.Assert().Equal(rewardDistribution2.TotalRewardsPool, response.GetEpochRewardDistribution().TotalRewardsPool, "total rewards pool should match")
	s.Assert().Equal(rewardDistribution2.TotalShares, response.GetEpochRewardDistribution().TotalShares, "total shares should match")
	s.Assert().False(response.GetEpochRewardDistribution().EpochEnded, "epoch should not be ended")
}

func (s *KeeperTestSuite) TestEligibilityCriteria() {
	s.SetupTest()
	queryClient := s.queryClient

	// Test against empty set
	response, err := queryClient.EligibilityCriteria(gocontext.Background(), &types.EligibilityCriteriaRequest{})
	s.Assert().Nil(err, "error should be nil")
	s.Assert().Equal(len(response.GetEligibilityCriteria()), 0, "response should contain empty list")

	// Test single entry
	criteria := types.NewEligibilityCriteria("test", &types.ActionDelegate{})
	s.app.RewardKeeper.SetEligibilityCriteria(s.ctx, criteria)
	response, err = queryClient.EligibilityCriteria(gocontext.Background(), &types.EligibilityCriteriaRequest{})
	s.Assert().Nil(err)
	s.Assert().Equal(1, len(response.GetEligibilityCriteria()), "response should contain only 1 eligibility criteria")
	s.Assert().Equal(criteria.Name, response.GetEligibilityCriteria()[0].Name, "eligibility criteria names should match")

	// Test multiple entry
	criteria2 := types.NewEligibilityCriteria("test2", &types.ActionDelegate{})
	criteria3 := types.NewEligibilityCriteria("test3", &types.ActionDelegate{})
	s.app.RewardKeeper.SetEligibilityCriteria(s.ctx, criteria2)
	s.app.RewardKeeper.SetEligibilityCriteria(s.ctx, criteria3)
	response, err = queryClient.EligibilityCriteria(gocontext.Background(), &types.EligibilityCriteriaRequest{})
	s.Assert().Nil(err)
	s.Assert().Equal(3, len(response.GetEligibilityCriteria()), "response should contain exactly 3 eligibility criteria")
}

func (s *KeeperTestSuite) TestEligibilityCriteriaByName() {
	s.SetupTest()
	queryClient := s.queryClient

	criteria := types.NewEligibilityCriteria("test", &types.ActionDelegate{})
	criteria2 := types.NewEligibilityCriteria("test2", &types.ActionDelegate{})
	s.app.RewardKeeper.SetEligibilityCriteria(s.ctx, criteria)
	s.app.RewardKeeper.SetEligibilityCriteria(s.ctx, criteria2)

	response, err := queryClient.EligibilityCriteriaByName(gocontext.Background(), &types.EligibilityCriteriaRequestByNameRequest{Name: "test"})
	s.Assert().Nil(err, "error should be nil")
	s.Assert().NotNil(response.GetEligibilityCriteria(), "EligibilityCriteria should not be nil")
	s.Assert().Equal(criteria.Name, response.GetEligibilityCriteria().Name, 1, "EligibilityCriteria names should match")

	response, err = queryClient.EligibilityCriteriaByName(gocontext.Background(), &types.EligibilityCriteriaRequestByNameRequest{Name: "test3"})
	s.Assert().Nil(err, "error should be nil")
	s.Assert().Nil(response.GetEligibilityCriteria(), "EligibilityCriteria should be nil")
}

func newRewardProgram(coinName string, id uint64, start uint64) *types.RewardProgram {
	action := types.NewActionDelegate()
	coin := sdk.NewInt64Coin(coinName, 10000)
	maxCoin := sdk.NewInt64Coin(coinName, 100)

	rewardProgram := types.NewRewardProgram("title", "description", 1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", coin, maxCoin, time.Now().UTC(), 60, 2, types.NewEligibilityCriteria("criteria", &action))

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
*/
