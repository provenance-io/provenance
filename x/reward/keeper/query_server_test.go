package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/provenance/x/reward/types"
)

func TestQueryServerTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) TestQueryRewardPrograms() {
	queryClient := s.queryClient

	response, err := queryClient.RewardPrograms(s.ctx.Context(), &types.QueryRewardProgramsRequest{})
	s.Assert().Nil(err, "query should not error")
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
	rewardProgram.Id = 2
	rewardProgram.State = types.RewardProgram_STATE_STARTED
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)
	rewardProgram.Id = 3
	rewardProgram.State = types.RewardProgram_STATE_FINISHED
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)
	rewardProgram.Id = 4
	rewardProgram.State = types.RewardProgram_STATE_EXPIRED
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

	response, err = queryClient.RewardPrograms(s.ctx.Context(), &types.QueryRewardProgramsRequest{QueryType: types.QueryRewardProgramsRequest_QUERY_TYPE_ALL})
	s.Assert().NoError(err, "query should not error")
	s.Assert().Equal(len(response.RewardPrograms), 4, "response should contain the added element")

	response, err = queryClient.RewardPrograms(s.ctx.Context(), &types.QueryRewardProgramsRequest{QueryType: types.QueryRewardProgramsRequest_QUERY_TYPE_PENDING})
	s.Assert().NoError(err, "query should not error")
	s.Assert().Equal(len(response.RewardPrograms), 1, "response should contain the added element")
	s.Assert().Equal(uint64(1), response.RewardPrograms[0].Id, "response should contain pending program")

	response, err = queryClient.RewardPrograms(s.ctx.Context(), &types.QueryRewardProgramsRequest{QueryType: types.QueryRewardProgramsRequest_QUERY_TYPE_ACTIVE})
	s.Assert().NoError(err, "query should not error")
	s.Assert().Equal(len(response.RewardPrograms), 1, "response should contain the added element")
	s.Assert().Equal(uint64(2), response.RewardPrograms[0].Id, "response should contain active program")

	response, err = queryClient.RewardPrograms(s.ctx.Context(), &types.QueryRewardProgramsRequest{QueryType: types.QueryRewardProgramsRequest_QUERY_TYPE_OUTSTANDING})
	s.Assert().NoError(err, "query should not error")
	s.Assert().Equal(len(response.RewardPrograms), 2, "response should contain the added element")
	s.Assert().Equal(uint64(1), response.RewardPrograms[0].Id, "response should contain the pending program")
	s.Assert().Equal(uint64(2), response.RewardPrograms[1].Id, "response should contain the active program")

	response, err = queryClient.RewardPrograms(s.ctx.Context(), &types.QueryRewardProgramsRequest{QueryType: types.QueryRewardProgramsRequest_QUERY_TYPE_FINISHED})
	s.Assert().NoError(err, "query should not error")
	s.Assert().Equal(len(response.RewardPrograms), 2, "response should contain the added element")
	s.Assert().Equal(uint64(3), response.RewardPrograms[0].Id, "response should contain the finished program")
	s.Assert().Equal(uint64(4), response.RewardPrograms[1].Id, "response should contain the expired program")

	responseId, err := queryClient.RewardProgramByID(s.ctx.Context(), &types.QueryRewardProgramByIDRequest{Id: uint64(4)})
	s.Assert().NoError(err, "query should not error")
	s.Assert().Equal(uint64(4), responseId.RewardProgram.Id, "response should contain the reward program with id 4")

	responseId, err = queryClient.RewardProgramByID(s.ctx.Context(), &types.QueryRewardProgramByIDRequest{Id: uint64(1000)})
	s.Assert().Error(err, "query should not error")
	s.Assert().Equal(true, responseId == nil, "response should be nil, since program doesn't exist")
}

func (s *KeeperTestSuite) TestClaimPeriodRewardDistributions() {

	queryClient := s.queryClient
	for i := 0; i < 101; i++ {
		s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, types.NewClaimPeriodRewardDistribution(uint64(i+1), 1, sdk.NewInt64Coin("jackthecat", 100), sdk.NewInt64Coin("jackthecat", 10), int64(i), false))
	}
	pageRequest := &query.PageRequest{}
	pageRequest.Limit = 100
	pageRequest.CountTotal = true
	response, err := queryClient.ClaimPeriodRewardDistributions(s.ctx.Context(), &types.QueryClaimPeriodRewardDistributionsRequest{pageRequest})
	s.Assert().NoError(err, "query should not error")
	s.Assert().Equal(response.Pagination.Total, uint64(101))
	s.Assert().Equal(100, len(response.ClaimPeriodRewardDistributions))

	response, err = queryClient.ClaimPeriodRewardDistributions(s.ctx.Context(), &types.QueryClaimPeriodRewardDistributionsRequest{Pagination: &query.PageRequest{Limit: 10000}})
	s.Assert().NoError(err, "query should not error")
	// 0 since pageRequest.CountTotal = false by default
	s.Assert().Equal(response.Pagination.Total, uint64(0), "should only return 100")
	s.Assert().Equal(101, len(response.ClaimPeriodRewardDistributions), "should only return 100 (max allowed per page)")

	response, err = queryClient.ClaimPeriodRewardDistributions(s.ctx.Context(), &types.QueryClaimPeriodRewardDistributionsRequest{Pagination: &query.PageRequest{Limit: 1, Offset: 9}})
	s.Assert().NoError(err, "query should not error")
	// 0 since pageRequest.CountTotal = false by default
	s.Assert().Equal(response.Pagination.Total, uint64(0), "should only return 100")
	s.Assert().Equal(1, len(response.ClaimPeriodRewardDistributions), "should only return 1")
	s.Assert().Equal(uint64(10), response.ClaimPeriodRewardDistributions[0].ClaimPeriodId)
}

func (s *KeeperTestSuite) TestClaimPeriodRewardDistributionByID() {

	queryClient := s.queryClient
	for i := 0; i < 101; i++ {
		s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, types.NewClaimPeriodRewardDistribution(uint64(i+1), 1, sdk.NewInt64Coin("jackthecat", 100), sdk.NewInt64Coin("jackthecat", 10), int64(i), false))
	}
	response, err := queryClient.ClaimPeriodRewardDistributionsByID(s.ctx.Context(), &types.QueryClaimPeriodRewardDistributionsByIDRequest{RewardId: uint64(1), ClaimPeriodId: uint64(612)})
	s.Assert().NoError(err, "query should not error")
	s.Assert().Nil(response.ClaimPeriodRewardDistribution, "ClaimPeriodRewardDistribution should not be found")

	response, err = queryClient.ClaimPeriodRewardDistributionsByID(s.ctx.Context(), &types.QueryClaimPeriodRewardDistributionsByIDRequest{RewardId: uint64(1), ClaimPeriodId: uint64(99)})
	s.Assert().NoError(err, "query should not error")
	s.Assert().Equal(response.ClaimPeriodRewardDistribution.RewardProgramId, uint64(1))
	s.Assert().Equal(response.ClaimPeriodRewardDistribution.ClaimPeriodId, uint64(99))
}

func (s *KeeperTestSuite) TestRewardDistributionsByAddress() {

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

	queryClient := s.queryClient
	for i := 0; i < 402; i++ {
		s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, types.NewClaimPeriodRewardDistribution(uint64(i+1), 1, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 10), int64(i), false))
	}

	for i := 0; i < 201; i++ {
		s.app.RewardKeeper.SetRewardAccountState(s.ctx, types.NewRewardAccountState(1, uint64(i+1), s.accountAddr.String(), 10, map[string]uint64{}))
		s.app.RewardKeeper.SetRewardAccountState(s.ctx, types.NewRewardAccountState(1, uint64(i+1), s.accountAddresses[2].String(), 10, map[string]uint64{}))
	}
	s.Assert().NotEqual(s.accountAddr.String(), s.accountAddresses[2].String())
	pageRequest := &query.PageRequest{}
	pageRequest.Limit = 100
	pageRequest.CountTotal = true
	response, err := queryClient.RewardDistributionsByAddress(s.ctx.Context(), &types.QueryRewardDistributionsByAddressRequest{
		Address:     s.accountAddr.String(),
		ClaimStatus: types.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
		Pagination:  pageRequest,
	})
	s.Assert().NoError(err, "query should not error")
	s.Assert().Equal(100, len(response.RewardAccountState))
	s.Assert().Equal(uint64(201), response.Pagination.Total)
	pageRequest1 := &query.PageRequest{}
	pageRequest1.Limit = 100
	pageRequest1.CountTotal = true
	s.Assert().NotNil(response.Pagination.NextKey)
	pageRequest1.Key = response.Pagination.NextKey
	response1, err := queryClient.RewardDistributionsByAddress(s.ctx.Context(), &types.QueryRewardDistributionsByAddressRequest{
		Address:     s.accountAddr.String(),
		ClaimStatus: types.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
		Pagination:  pageRequest1,
	})

	s.Assert().NoError(err, "query should not error")
	s.Assert().Equal(100, len(response1.RewardAccountState))
	pageRequest2 := &query.PageRequest{}
	pageRequest2.Limit = 100
	pageRequest2.CountTotal = true
	s.Assert().NotNil(response1.Pagination.NextKey)
	pageRequest2.Key = response1.Pagination.NextKey
	response2, err := queryClient.RewardDistributionsByAddress(s.ctx.Context(), &types.QueryRewardDistributionsByAddressRequest{
		Address:     s.accountAddr.String(),
		ClaimStatus: types.RewardAccountState_CLAIM_STATUS_UNSPECIFIED,
		Pagination:  pageRequest2,
	})

	s.Assert().NoError(err, "query should not error")
	s.Assert().Equal(1, len(response2.RewardAccountState))

}
