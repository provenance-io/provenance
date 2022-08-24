package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (suite *KeeperTestSuite) TestGetSetClaimPeriodRewardDistribution() {
	rewardDistribution := types.NewClaimPeriodRewardDistribution(
		1,
		2,
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 200000),
		3,
		true,
	)
	suite.app.RewardKeeper.SetClaimPeriodRewardDistribution(suite.ctx, rewardDistribution)
	retrieved, err := suite.app.RewardKeeper.GetClaimPeriodRewardDistribution(suite.ctx, 1, 2)
	suite.Assert().NoError(err)
	suite.Assert().Equal(rewardDistribution.ClaimPeriodId, retrieved.ClaimPeriodId, "claim period id must match")
	suite.Assert().Equal(rewardDistribution.RewardProgramId, retrieved.RewardProgramId, "reward program id must match")
	suite.Assert().Equal(rewardDistribution.RewardsPool, retrieved.RewardsPool, "rewards pool must match")
	suite.Assert().Equal(rewardDistribution.TotalRewardsPoolForClaimPeriod, retrieved.TotalRewardsPoolForClaimPeriod, "total rewards pool must match")
	suite.Assert().Equal(rewardDistribution.TotalShares, retrieved.TotalShares, "total shares must match")
	suite.Assert().Equal(rewardDistribution.ClaimPeriodEnded, retrieved.ClaimPeriodEnded, "claim period ended must match")

	retrieved, err = suite.app.RewardKeeper.GetClaimPeriodRewardDistribution(suite.ctx, 1, 99)
	suite.Assert().NoError(err)
	suite.Assert().Equal(uint64(0), retrieved.RewardProgramId, "reward program id must be invalid")
}

func (suite *KeeperTestSuite) TestIterateClaimPeriodRewardDistributions() {
	tests := []struct {
		name          string
		distributions []types.ClaimPeriodRewardDistribution
		halt          bool
		counter       int
	}{
		{
			"valid - can handle empty distributions",
			[]types.ClaimPeriodRewardDistribution{},
			false,
			0,
		},
		{
			"valid - can handle one distribution",
			[]types.ClaimPeriodRewardDistribution{
				types.NewClaimPeriodRewardDistribution(1, 1, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, false),
			},
			false,
			1,
		},
		{
			"valid - can handle multiple distributions",
			[]types.ClaimPeriodRewardDistribution{
				types.NewClaimPeriodRewardDistribution(1, 1, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, false),
				types.NewClaimPeriodRewardDistribution(1, 2, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, false),
				types.NewClaimPeriodRewardDistribution(1, 3, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, false),
			},
			false,
			3,
		},
		{
			"valid - can handle halting",
			[]types.ClaimPeriodRewardDistribution{
				types.NewClaimPeriodRewardDistribution(1, 1, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, false),
				types.NewClaimPeriodRewardDistribution(1, 2, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, false),
				types.NewClaimPeriodRewardDistribution(1, 3, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, false),
			},
			true,
			1,
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			counter := 0
			for _, distribution := range tc.distributions {
				suite.app.RewardKeeper.SetClaimPeriodRewardDistribution(suite.ctx, distribution)
			}
			err := suite.app.RewardKeeper.IterateClaimPeriodRewardDistributions(suite.ctx, func(ClaimPeriodRewardDistribution types.ClaimPeriodRewardDistribution) (stop bool) {
				counter += 1
				return tc.halt
			})
			assert.NoError(t, err, "No error is thrown")
			assert.Equal(t, tc.counter, counter, "iterated the correct number of times")
		})
	}
}

func (suite *KeeperTestSuite) TestGetAllClaimPeriodRewardDistributions() {
	tests := []struct {
		name          string
		distributions []types.ClaimPeriodRewardDistribution
		expected      int
	}{
		{
			"valid - can handle empty distributions",
			[]types.ClaimPeriodRewardDistribution{},
			0,
		},
		{
			"valid - can handle one distribution",
			[]types.ClaimPeriodRewardDistribution{
				types.NewClaimPeriodRewardDistribution(1, 1, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, false),
			},
			1,
		},
		{
			"valid - can handle multiple distributions",
			[]types.ClaimPeriodRewardDistribution{
				types.NewClaimPeriodRewardDistribution(1, 1, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, false),
				types.NewClaimPeriodRewardDistribution(1, 2, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, false),
				types.NewClaimPeriodRewardDistribution(1, 3, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, false),
			},
			3,
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			for _, distribution := range tc.distributions {
				suite.app.RewardKeeper.SetClaimPeriodRewardDistribution(suite.ctx, distribution)
			}
			results, err := suite.app.RewardKeeper.GetAllClaimPeriodRewardDistributions(suite.ctx)
			assert.NoError(t, err, "No error is thrown")
			assert.Equal(t, tc.expected, len(results), "returned the correct number of claim period reward distributions")
		})
	}
}

func (suite *KeeperTestSuite) TestClaimPeriodRewardDistributionIsValid() {
	tests := []struct {
		name         string
		distribution types.ClaimPeriodRewardDistribution
		valid        bool
	}{
		{
			"valid - claim reward distribution has valid id",
			types.NewClaimPeriodRewardDistribution(1, 1, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 0, false),
			true,
		},
		{
			"invalid - claim reward distribution has invalid id",
			types.NewClaimPeriodRewardDistribution(0, 0, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 0, false),
			false,
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			result := suite.app.RewardKeeper.ClaimPeriodRewardDistributionIsValid(&tc.distribution)
			assert.Equal(t, tc.valid, result, "the output should match the expected valid value")
		})
	}
}

func (suite *KeeperTestSuite) TestRemoveClaimPeriodRewardDistribution() {
	tests := []struct {
		name          string
		distributions []types.ClaimPeriodRewardDistribution
		expected      int
		removed       bool
	}{
		{
			"valid - can handle removal of invalid claim period",
			[]types.ClaimPeriodRewardDistribution{},
			0,
			false,
		},
		{
			"valid - can handle valid removal",
			[]types.ClaimPeriodRewardDistribution{
				types.NewClaimPeriodRewardDistribution(1, 1, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, false),
				types.NewClaimPeriodRewardDistribution(2, 1, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, false),
			},
			1,
			true,
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.name, func(t *testing.T) {
			for _, distribution := range tc.distributions {
				suite.app.RewardKeeper.SetClaimPeriodRewardDistribution(suite.ctx, distribution)
			}
			removed := suite.app.RewardKeeper.RemoveClaimPeriodRewardDistribution(suite.ctx, 1, 1)
			results, err := suite.app.RewardKeeper.GetAllClaimPeriodRewardDistributions(suite.ctx)
			assert.Equal(t, tc.removed, removed, "removal status does not match expectation")
			assert.NoError(t, err, "No error is thrown")
			assert.Equal(t, tc.expected, len(results), "returned the correct number of claim period reward distributions")
		})
	}
}
