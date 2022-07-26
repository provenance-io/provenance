package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (s *KeeperTestSuite) TestQueryRewardPrograms() {
	s.SetupTest()
	queryClient := s.queryClient

	// Test against empty set
	response, err := queryClient.RewardPrograms(s.ctx.Context(), &types.QueryRewardProgramsRequest{})
	s.Assert().Nil(err, "Error should be nil")
	s.Assert().Equal(len(response.RewardPrograms), 0, "response should contain empty list")

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		s.accountAddr.String(),
		sdk.NewCoin("nhash", sdk.NewInt(1000_000_000_000)),
		sdk.NewCoin("nhash", sdk.NewInt(10_000_000_000)),
		time.Now().Add(100*time.Millisecond),
		uint64(30),
		10,
		10,
		3,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Transfer{
					Transfer: &types.ActionTransfer{
						MinimumActions:          0,
						MaximumActions:          10,
						MinimumDelegationAmount: sdk.NewCoin("nhash", sdk.ZeroInt()),
					},
				},
			},
		},
	)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

	response, err = queryClient.RewardPrograms(s.ctx.Context(), &types.QueryRewardProgramsRequest{})
	s.Assert().Nil(err, "Error should be nil")
	s.Assert().Equal(len(response.RewardPrograms), 1, "response should contain the added element")
	s.Assert().True(rewardProgram.Equal(response.RewardPrograms[0]), "reward programs should match")
}

// TODO: RewardProgramByID

// TODO: ClaimPeriodRewardDistributions

// TODO: ClaimPeriodRewardDistributionsByID

// TODO: QueryRewardDistributionsByAddress
