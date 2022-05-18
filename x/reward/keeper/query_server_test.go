package keeper_test

import (
	gocontext "context"
	"time"

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
	suite.Require().Fail("not implemented")
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
	s.SetupTest()
	queryClient := s.queryClient

	// Test against empty set
	response, err := queryClient.RewardClaims(gocontext.Background(), &types.RewardClaimsRequest{})
	s.Assert().Nil(err, "error should be nil")
	s.Assert().Equal(len(response.RewardClaims), 0, "response should contain empty list")

	// Test with 1 item
	rewardClaim := types.NewRewardClaim("testing", []types.SharesPerEpochPerRewardsProgram{
		{
			RewardProgramId:      1,
			TotalShares:          2,
			EphemeralActionCount: 3,
			LatestRecordedEpoch:  4,
			Claimed:              false,
			Expired:              false,
			TotalRewardClaimed:   sdk.NewInt64Coin("jackthecat", 100),
		},
	}, false)
	s.app.RewardKeeper.SetRewardClaim(s.ctx, rewardClaim)
	response, err = queryClient.RewardClaims(gocontext.Background(), &types.RewardClaimsRequest{})
	s.Assert().Nil(err, "error should be nil")
	s.Assert().Equal(1, len(response.RewardClaims), "response should contain one reward claim")

	// Test with multiple item
	rewardClaim = types.NewRewardClaim("testing2", []types.SharesPerEpochPerRewardsProgram{
		{
			RewardProgramId:      2,
			TotalShares:          3,
			EphemeralActionCount: 4,
			LatestRecordedEpoch:  5,
			Claimed:              false,
			Expired:              false,
			TotalRewardClaimed:   sdk.NewInt64Coin("jackthecat", 200),
		},
	}, false)
	s.app.RewardKeeper.SetRewardClaim(s.ctx, rewardClaim)
	response, err = queryClient.RewardClaims(gocontext.Background(), &types.RewardClaimsRequest{})
	s.Assert().Nil(err, "error should be nil")
	s.Assert().Equal(2, len(response.RewardClaims), "response should contain multiple reward claims")
}

func (s *KeeperTestSuite) TestRewardClaimById() {
	s.SetupTest()
	queryClient := s.queryClient

	rewardClaim := types.NewRewardClaim("testing", []types.SharesPerEpochPerRewardsProgram{
		{
			RewardProgramId:      1,
			TotalShares:          2,
			EphemeralActionCount: 3,
			LatestRecordedEpoch:  4,
			Claimed:              false,
			Expired:              false,
			TotalRewardClaimed:   sdk.NewInt64Coin("jackthecat", 100),
		},
	}, false)
	s.app.RewardKeeper.SetRewardClaim(s.ctx, rewardClaim)
	rewardClaim = types.NewRewardClaim("testing2", []types.SharesPerEpochPerRewardsProgram{
		{
			RewardProgramId:      2,
			TotalShares:          3,
			EphemeralActionCount: 4,
			LatestRecordedEpoch:  5,
			Claimed:              false,
			Expired:              false,
			TotalRewardClaimed:   sdk.NewInt64Coin("jackthecat", 200),
		},
	}, false)
	s.app.RewardKeeper.SetRewardClaim(s.ctx, rewardClaim)

	response, err := queryClient.RewardClaimByID(gocontext.Background(), &types.RewardClaimByIDRequest{Id: "testing"})
	s.Assert().Nil(err, "error should be nil")
	s.Assert().NotNil(response.GetRewardClaim(), "RewardClaim should not be nil")
	s.Assert().Equal(len(response.GetRewardClaim().SharesPerEpochPerReward), 1, "Length of shares should match")
	s.Assert().Equal("testing", response.GetRewardClaim().Address, "RewardClaim address should match")
	s.Assert().Equal(uint64(1), response.GetRewardClaim().SharesPerEpochPerReward[0].RewardProgramId, "RewardProgramId should match")
	s.Assert().Equal(int64(2), response.GetRewardClaim().SharesPerEpochPerReward[0].TotalShares, "TotalShares should match")
	s.Assert().Equal(int64(3), response.GetRewardClaim().SharesPerEpochPerReward[0].EphemeralActionCount, "EphemeralActionCount should match")
	s.Assert().Equal(uint64(4), response.GetRewardClaim().SharesPerEpochPerReward[0].LatestRecordedEpoch, "LatestRecordedEpoch should match")
	s.Assert().False(response.GetRewardClaim().SharesPerEpochPerReward[0].Claimed, "Claimed should match")
	s.Assert().False(response.GetRewardClaim().SharesPerEpochPerReward[0].Expired, "Expired should match")
	s.Assert().Equal(sdk.NewInt64Coin("jackthecat", 100), response.GetRewardClaim().SharesPerEpochPerReward[0].TotalRewardClaimed, "")

	response, err = queryClient.RewardClaimByID(gocontext.Background(), &types.RewardClaimByIDRequest{Id: "testing3"})
	s.Assert().Nil(err, "error should be nil")
	s.Assert().Nil(response.GetRewardClaim(), "RewardClaim should be nil")
}

func (s *KeeperTestSuite) TestEpochDistributionRewards() {
	s.SetupTest()
	queryClient := s.queryClient

	// Test against empty set
	response, err := queryClient.EpochRewardDistributions(gocontext.Background(), &types.EpochRewardDistributionRequest{})
	s.Assert().Nil(err, "error should be nil")
	s.Assert().Equal(len(response.GetEpochRewardDistribution()), 0, "response should contain empty list")

	// Test single entry
	rewardDistribution := types.NewEpochRewardDistribution("day", 1, sdk.NewInt64Coin("jackthecat", 100), 5, false)
	s.app.RewardKeeper.SetEpochRewardDistribution(s.ctx, rewardDistribution)
	response, err = queryClient.EpochRewardDistributions(gocontext.Background(), &types.EpochRewardDistributionRequest{})
	s.Assert().Nil(err)
	s.Assert().Equal(1, len(response.GetEpochRewardDistribution()), "response should contain only 1 epoch reward distribution")
	s.Assert().Equal(rewardDistribution.EpochId, response.GetEpochRewardDistribution()[0].EpochId, "epoch ids should match")
	s.Assert().Equal(rewardDistribution.RewardProgramId, response.GetEpochRewardDistribution()[0].RewardProgramId, "reward ids should match")
	s.Assert().Equal(rewardDistribution.TotalRewardsPool, response.GetEpochRewardDistribution()[0].TotalRewardsPool, "total rewards pool should match")
	s.Assert().Equal(rewardDistribution.TotalShares, response.GetEpochRewardDistribution()[0].TotalShares, "total shares should match")
	s.Assert().False(response.GetEpochRewardDistribution()[0].EpochEnded, "epoch should not be ended")

	// Test multiple entry
	rewardDistribution2 := types.NewEpochRewardDistribution("day", 2, sdk.NewInt64Coin("jackthecat", 200), 10, false)
	rewardDistribution3 := types.NewEpochRewardDistribution("day", 3, sdk.NewInt64Coin("jackthecat", 300), 15, false)
	s.app.RewardKeeper.SetEpochRewardDistribution(s.ctx, rewardDistribution2)
	s.app.RewardKeeper.SetEpochRewardDistribution(s.ctx, rewardDistribution3)
	response, err = queryClient.EpochRewardDistributions(gocontext.Background(), &types.EpochRewardDistributionRequest{})
	s.Assert().Nil(err)
	s.Assert().Equal(3, len(response.GetEpochRewardDistribution()), "response should contain exactly 3 epoch reward distribution")
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
