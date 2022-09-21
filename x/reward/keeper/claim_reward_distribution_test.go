package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (s *KeeperTestSuite) TestGetSetClaimPeriodRewardDistribution() {
	rewardDistribution := types.NewClaimPeriodRewardDistribution(
		1,
		2,
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 200000),
		3,
		true,
	)
	s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, rewardDistribution)
	retrieved, err := s.app.RewardKeeper.GetClaimPeriodRewardDistribution(s.ctx, 1, 2)
	s.Assert().NoError(err)
	s.Assert().Equal(rewardDistribution.ClaimPeriodId, retrieved.ClaimPeriodId, "claim period id must match")
	s.Assert().Equal(rewardDistribution.RewardProgramId, retrieved.RewardProgramId, "reward program id must match")
	s.Assert().Equal(rewardDistribution.RewardsPool, retrieved.RewardsPool, "rewards pool must match")
	s.Assert().Equal(rewardDistribution.TotalRewardsPoolForClaimPeriod, retrieved.TotalRewardsPoolForClaimPeriod, "total rewards pool must match")
	s.Assert().Equal(rewardDistribution.TotalShares, retrieved.TotalShares, "total shares must match")
	s.Assert().Equal(rewardDistribution.ClaimPeriodEnded, retrieved.ClaimPeriodEnded, "claim period ended must match")

	retrieved, err = s.app.RewardKeeper.GetClaimPeriodRewardDistribution(s.ctx, 1, 99)
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(0), retrieved.RewardProgramId, "reward program id must be invalid")
}

func (s *KeeperTestSuite) TestIterateClaimPeriodRewardDistributions() {
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
		s.T().Run(tc.name, func(t *testing.T) {
			counter := 0
			for _, distribution := range tc.distributions {
				s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, distribution)
			}
			err := s.app.RewardKeeper.IterateClaimPeriodRewardDistributions(s.ctx, func(ClaimPeriodRewardDistribution types.ClaimPeriodRewardDistribution) (stop bool) {
				counter += 1
				return tc.halt
			})
			assert.NoError(t, err, "No error is thrown")
			assert.Equal(t, tc.counter, counter, "iterated the correct number of times")
		})
	}
}

func (s *KeeperTestSuite) TestGetAllClaimPeriodRewardDistributions() {
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
		s.T().Run(tc.name, func(t *testing.T) {
			for _, distribution := range tc.distributions {
				s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, distribution)
			}
			results, err := s.app.RewardKeeper.GetAllClaimPeriodRewardDistributions(s.ctx)
			assert.NoError(t, err, "No error is thrown")
			assert.Equal(t, tc.expected, len(results), "returned the correct number of claim period reward distributions")
		})
	}
}

func (s *KeeperTestSuite) TestRemoveClaimPeriodRewardDistribution() {
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
		s.T().Run(tc.name, func(t *testing.T) {
			for _, distribution := range tc.distributions {
				s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, distribution)
			}
			removed := s.app.RewardKeeper.RemoveClaimPeriodRewardDistribution(s.ctx, 1, 1)
			results, err := s.app.RewardKeeper.GetAllClaimPeriodRewardDistributions(s.ctx)
			assert.Equal(t, tc.removed, removed, "removal status does not match expectation")
			assert.NoError(t, err, "No error is thrown")
			assert.Equal(t, tc.expected, len(results), "returned the correct number of claim period reward distributions")
		})
	}
}
