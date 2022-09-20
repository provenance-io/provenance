package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (s *KeeperTestSuite) TestCreateRewardProgramTransaction() {

	minimumDelegation := sdk.NewInt64Coin("nhash", 100)
	maximumDelegation := sdk.NewInt64Coin("nhash", 200)
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, s.accountAddresses[0], sdk.NewCoins(sdk.NewInt64Coin("nhash", 100000))), "funding account")

	msg := types.NewMsgCreateRewardProgramRequest(
		"title",
		"description",
		s.accountAddresses[0].String(),
		sdk.NewInt64Coin("nhash", 1000),
		sdk.NewInt64Coin("nhash", 100),
		time.Now(),
		4,
		2,
		1,
		4,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               10,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
					},
				},
			},
		},
	)

	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
	result, err := s.handler(s.ctx, msg)
	s.Assert().NoError(err, "msg server should handle a new valid reward program")
	s.Assert().Less(0, len(result.GetEvents()), "should have emitted events")
	s.Assert().Equal(result.Events[len(result.Events)-1].Type, "reward_program_created", "emitted event should have correct event type")
	s.Assert().Equal(1, len(result.Events[len(result.Events)-1].Attributes), "emitted event should have correct event number of attributes")
	s.Assert().Equal(result.Events[len(result.Events)-1].Attributes[0].Key, []byte("reward_program_id"), "should have correct key")
	s.Assert().Equal(result.Events[len(result.Events)-1].Attributes[0].Value, []byte("1"), "should have correct value")

	program, err := s.app.RewardKeeper.GetRewardProgram(s.ctx, 1)
	s.Assert().NoError(err, "No error should be returned")
	s.Assert().Nil(program.Validate(), "should not have a validation error")
}

func (s *KeeperTestSuite) TestCreateRewardProgramFailedTransaction() {

	minimumDelegation := sdk.NewInt64Coin("nhash", 100)
	maximumDelegation := sdk.NewInt64Coin("nhash", 200)

	msg := types.NewMsgCreateRewardProgramRequest(
		"title",
		"description",
		s.accountAddresses[0].String(),
		sdk.NewInt64Coin("nhash", 1000),
		sdk.NewInt64Coin("nhash", 100),
		time.Now(),
		4,
		2,
		1,
		4,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               10,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
					},
				},
			},
		},
	)

	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
	result, err := s.handler(s.ctx, msg)
	s.Assert().Error(err, "msg server should throw error for invalid creation")
	s.Assert().Nil(result, "result should be nil and have no events")

	program, err := s.app.RewardKeeper.GetRewardProgram(s.ctx, 1)
	s.Assert().Error(err, "error should be returned")
	s.Assert().NotNil(program.Validate(), "should have validation issue for invalid program")
}

func (s *KeeperTestSuite) TestRewardClaimTransaction() {

	now := s.ctx.BlockTime()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv",
		sdk.NewInt64Coin("nhash", 1000),
		sdk.NewInt64Coin("nhash", 100),
		now,
		10,
		3,
		0,
		uint64(now.Day()),
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Vote{
					Vote: &types.ActionVote{
						MinimumActions:          0,
						MaximumActions:          1,
						MinimumDelegationAmount: minDelegation,
					},
				},
			},
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               1,
						MinimumDelegationAmount:      &minDelegation,
						MaximumDelegationAmount:      &maxDelegation,
						MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
					},
				},
			},
		},
	)
	rewardProgram.State = types.RewardProgram_STATE_FINISHED
	rewardProgram.CurrentClaimPeriod = rewardProgram.GetClaimPeriods()
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

	for i := 1; i <= int(rewardProgram.GetClaimPeriods()); i++ {
		state := types.NewRewardAccountState(rewardProgram.GetId(), uint64(i), s.accountAddresses[0].String(), 1, []*types.ActionCounter{})
		state.ClaimStatus = types.RewardAccountState_CLAIM_STATUS_CLAIMABLE
		s.app.RewardKeeper.SetRewardAccountState(s.ctx, state)
		distribution := types.NewClaimPeriodRewardDistribution(uint64(i), rewardProgram.GetId(), sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, true)
		s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, distribution)
	}

	msg := types.NewMsgClaimRewardsRequest(1, s.accountAddresses[0].String())
	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
	result, err := s.handler(s.ctx, msg)

	s.Assert().NoError(err, "msg server should handle valid reward claim")
	s.Assert().NotNil(result, "msg server should emit events")
	s.Assert().Less(0, len(result.GetEvents()), "should have emitted events")
	s.Assert().Equal(result.Events[len(result.Events)-1].Type, "claim_rewards", "emitted event should have correct event type")
	s.Assert().Equal(2, len(result.Events[len(result.Events)-1].Attributes), "emitted event should have correct number of attributes")
	s.Assert().Equal(result.Events[len(result.Events)-1].Attributes[0].Key, []byte("reward_program_id"), "should have correct key")
	s.Assert().Equal(result.Events[len(result.Events)-1].Attributes[0].Value, []byte("1"), "should have correct program id")
	s.Assert().Equal(result.Events[len(result.Events)-1].Attributes[1].Key, []byte("rewards_claim_address"), "should have correct key")
	s.Assert().Equal(result.Events[len(result.Events)-1].Attributes[1].Value, []byte(s.accountAddresses[0].String()), "should have correct address value")
}

func (s *KeeperTestSuite) TestRewardClaimInvalidTransaction() {

	msg := types.NewMsgClaimRewardsRequest(1, "invalid address")
	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
	result, err := s.handler(s.ctx, msg)

	s.Assert().Error(err, "msg server should handle an invalid reward claim")
	s.Assert().Nil(result, "should have no emitted events")
}

func (s *KeeperTestSuite) TestRewardClaimTransactionInvalidClaimer() {

	now := s.ctx.BlockTime()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv",
		sdk.NewInt64Coin("nhash", 1000),
		sdk.NewInt64Coin("nhash", 100),
		now,
		10,
		3,
		0,
		uint64(now.Day()),
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Vote{
					Vote: &types.ActionVote{
						MinimumActions:          0,
						MaximumActions:          1,
						MinimumDelegationAmount: minDelegation,
					},
				},
			},
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               1,
						MinimumDelegationAmount:      &minDelegation,
						MaximumDelegationAmount:      &maxDelegation,
						MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
					},
				},
			},
		},
	)
	rewardProgram.State = types.RewardProgram_STATE_FINISHED
	rewardProgram.CurrentClaimPeriod = rewardProgram.GetClaimPeriods()
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

	for i := 1; i <= int(rewardProgram.GetClaimPeriods()); i++ {
		state := types.NewRewardAccountState(rewardProgram.GetId(), uint64(i), s.accountAddresses[0].String(), 1, []*types.ActionCounter{})
		state.ClaimStatus = types.RewardAccountState_CLAIM_STATUS_CLAIMABLE
		s.app.RewardKeeper.SetRewardAccountState(s.ctx, state)
		distribution := types.NewClaimPeriodRewardDistribution(uint64(i), rewardProgram.GetId(), sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, true)
		s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, distribution)
	}

	msg := types.NewMsgClaimRewardsRequest(1, s.accountAddresses[1].String())
	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
	result, err := s.handler(s.ctx, msg)
	s.Assert().NoError(err, "msg server should handle valid reward claim")
	s.Assert().NotNil(result, "msg server should emit events")

	var response types.MsgClaimRewardsResponse
	response.Unmarshal(result.Data)
	s.Assert().Equal(uint64(1), response.GetClaimDetails().RewardProgramId, "should have correct reward program id")
	s.Assert().Equal(0, len(response.GetClaimDetails().ClaimedRewardPeriodDetails), "should have no details")
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 0), response.GetClaimDetails().TotalRewardClaim, "should have no reward claim")
}

func (s *KeeperTestSuite) TestClaimAllRewardsTransaction() {
	now := s.ctx.BlockTime()

	for i := 0; i < 3; i++ {
		rewardProgram := types.NewRewardProgram(
			"title",
			"description",
			uint64(i+1),
			s.accountAddresses[0].String(),
			sdk.NewInt64Coin("nhash", 1000),
			sdk.NewInt64Coin("nhash", 100),
			now,
			10,
			3,
			0,
			uint64(now.Day()),
			[]types.QualifyingAction{
				{
					Type: &types.QualifyingAction_Vote{
						Vote: &types.ActionVote{
							MinimumActions:          0,
							MaximumActions:          1,
							MinimumDelegationAmount: minDelegation,
						},
					},
				},
				{
					Type: &types.QualifyingAction_Delegate{
						Delegate: &types.ActionDelegate{
							MinimumActions:               0,
							MaximumActions:               1,
							MinimumDelegationAmount:      &minDelegation,
							MaximumDelegationAmount:      &maxDelegation,
							MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
							MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
						},
					},
				},
			},
		)
		rewardProgram.State = types.RewardProgram_STATE_FINISHED
		rewardProgram.CurrentClaimPeriod = rewardProgram.GetClaimPeriods()
		s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

		for j := 1; j <= int(rewardProgram.GetClaimPeriods()); j++ {
			state := types.NewRewardAccountState(rewardProgram.GetId(), uint64(j), s.accountAddresses[0].String(), 1, []*types.ActionCounter{})
			state.ClaimStatus = types.RewardAccountState_CLAIM_STATUS_CLAIMABLE
			s.app.RewardKeeper.SetRewardAccountState(s.ctx, state)
			distribution := types.NewClaimPeriodRewardDistribution(uint64(j), rewardProgram.GetId(), sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, true)
			s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, distribution)
		}
	}

	msg := types.NewMsgClaimAllRewardsRequest(s.accountAddresses[0].String())
	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
	result, err := s.handler(s.ctx, msg)
	s.Assert().NoError(err, "msg server should handle valid reward claim")
	s.Assert().NotNil(result, "msg server should emit events")

	var response types.MsgClaimAllRewardsResponse
	response.Unmarshal(result.Data)
	details := response.ClaimDetails
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 900), response.TotalRewardClaim[0], "should total up the rewards from the periods")
	s.Assert().Equal(3, len(details), "should have every reward program")
	for i := 0; i < len(details); i++ {
		s.Assert().Equal(3, len(details[i].ClaimedRewardPeriodDetails), "should have claims from every period")
		s.Assert().Equal(sdk.NewInt64Coin("nhash", 300), details[i].TotalRewardClaim, "should total up the rewards from the periods")
		s.Assert().Equal(uint64(i+1), details[i].RewardProgramId, "should have the correct id")
	}
}

func (s *KeeperTestSuite) TestClaimAllRewardsNoProgramsTransaction() {
	msg := types.NewMsgClaimAllRewardsRequest(s.accountAddresses[0].String())
	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
	result, err := s.handler(s.ctx, msg)
	s.Assert().NoError(err, "no error should be returned in a valid call")

	var response types.MsgClaimAllRewardsResponse
	response.Unmarshal(result.Data)
	details := response.ClaimDetails

	s.Assert().Equal(0, len(response.TotalRewardClaim), "should have no nhash")
	s.Assert().Equal(0, len(details), "should have no reward program")
}

func (s *KeeperTestSuite) TestRewardClaimAllRewardsInvalidAddressTransaction() {
	now := s.ctx.BlockTime()

	for i := 0; i < 3; i++ {
		rewardProgram := types.NewRewardProgram(
			"title",
			"description",
			uint64(i+1),
			s.accountAddresses[0].String(),
			sdk.NewInt64Coin("nhash", 1000),
			sdk.NewInt64Coin("nhash", 100),
			now,
			10,
			3,
			0,
			uint64(now.Day()),
			[]types.QualifyingAction{
				{
					Type: &types.QualifyingAction_Vote{
						Vote: &types.ActionVote{
							MinimumActions:          0,
							MaximumActions:          1,
							MinimumDelegationAmount: minDelegation,
						},
					},
				},
				{
					Type: &types.QualifyingAction_Delegate{
						Delegate: &types.ActionDelegate{
							MinimumActions:               0,
							MaximumActions:               1,
							MinimumDelegationAmount:      &minDelegation,
							MaximumDelegationAmount:      &maxDelegation,
							MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
							MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
						},
					},
				},
			},
		)
		rewardProgram.State = types.RewardProgram_STATE_FINISHED
		rewardProgram.CurrentClaimPeriod = rewardProgram.GetClaimPeriods()
		s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

		for j := 1; j <= int(rewardProgram.GetClaimPeriods()); j++ {
			state := types.NewRewardAccountState(rewardProgram.GetId(), uint64(j), s.accountAddresses[0].String(), 1, []*types.ActionCounter{})
			state.ClaimStatus = types.RewardAccountState_CLAIM_STATUS_CLAIMABLE
			s.app.RewardKeeper.SetRewardAccountState(s.ctx, state)
			distribution := types.NewClaimPeriodRewardDistribution(uint64(j), rewardProgram.GetId(), sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, true)
			s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, distribution)
		}
	}

	msg := types.NewMsgClaimAllRewardsRequest("invalid address")
	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
	result, err := s.handler(s.ctx, msg)
	s.Assert().Error(err, "error should be returned else state store will commit")
	s.Assert().Nil(result)
}

func (s *KeeperTestSuite) TestClaimAllRewardsExpiredTransaction() {
	now := s.ctx.BlockTime()

	for i := 0; i < 3; i++ {
		rewardProgram := types.NewRewardProgram(
			"title",
			"description",
			uint64(i+1),
			s.accountAddresses[0].String(),
			sdk.NewInt64Coin("nhash", 1000),
			sdk.NewInt64Coin("nhash", 100),
			now,
			10,
			3,
			0,
			uint64(now.Day()),
			[]types.QualifyingAction{
				{
					Type: &types.QualifyingAction_Vote{
						Vote: &types.ActionVote{
							MinimumActions:          0,
							MaximumActions:          1,
							MinimumDelegationAmount: minDelegation,
						},
					},
				},
				{
					Type: &types.QualifyingAction_Delegate{
						Delegate: &types.ActionDelegate{
							MinimumActions:               0,
							MaximumActions:               1,
							MinimumDelegationAmount:      &minDelegation,
							MaximumDelegationAmount:      &maxDelegation,
							MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
							MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
						},
					},
				},
			},
		)
		rewardProgram.State = types.RewardProgram_STATE_FINISHED
		rewardProgram.CurrentClaimPeriod = rewardProgram.GetClaimPeriods()
		s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

		for j := 1; j <= int(rewardProgram.GetClaimPeriods()); j++ {
			state := types.NewRewardAccountState(rewardProgram.GetId(), uint64(j), s.accountAddresses[0].String(), 1, []*types.ActionCounter{})
			state.ClaimStatus = types.RewardAccountState_CLAIM_STATUS_EXPIRED
			s.app.RewardKeeper.SetRewardAccountState(s.ctx, state)
			distribution := types.NewClaimPeriodRewardDistribution(uint64(j), rewardProgram.GetId(), sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, true)
			s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, distribution)
		}
	}

	msg := types.NewMsgClaimAllRewardsRequest(s.accountAddresses[0].String())
	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
	result, err := s.handler(s.ctx, msg)
	s.Assert().NoError(err, "no error should be returned in a valid call")

	var response types.MsgClaimAllRewardsResponse
	response.Unmarshal(result.Data)
	details := response.ClaimDetails

	s.Assert().Equal(0, len(response.TotalRewardClaim), "should have no nhash")
	s.Assert().Equal(0, len(details), "should have no reward program")
}

func (s *KeeperTestSuite) TestEndRewardProgramRequest() {
	testCases := []struct {
		name         string
		id           uint64
		address      string
		expectErr    bool
		expectErrMsg string
	}{
		{"end reward program request - invalid reward program id",
			88,
			s.accountAddresses[0].String(),
			true,
			"reward program not found",
		},
		{"end reward program request - invalid executor",
			1,
			s.accountAddresses[1].String(),
			true,
			"not authorized to end the reward program",
		},
		{"end reward program request - invalid state for reward program",
			3,
			s.accountAddresses[0].String(),
			true,
			"unable to end a reward program that is finished or expired",
		},
		{"end reward program request - valid request in pending state",
			1,
			s.accountAddresses[0].String(),
			false,
			"",
		},
		{"end reward program request - valid requested in started state",
			2,
			s.accountAddresses[0].String(),
			false,
			"",
		},
	}

	now := s.ctx.BlockTime()
	for i := 0; i < 3; i++ {
		rewardProgram := types.NewRewardProgram(
			"title",
			"description",
			uint64(i+1),
			s.accountAddresses[0].String(),
			sdk.NewInt64Coin("nhash", 1000),
			sdk.NewInt64Coin("nhash", 100),
			now,
			10,
			3,
			0,
			uint64(now.Day()),
			[]types.QualifyingAction{
				{
					Type: &types.QualifyingAction_Vote{
						Vote: &types.ActionVote{
							MinimumActions:          0,
							MaximumActions:          1,
							MinimumDelegationAmount: minDelegation,
						},
					},
				},
			},
		)
		switch i + 1 {
		case 1:
			rewardProgram.State = types.RewardProgram_STATE_PENDING
		case 2:
			rewardProgram.State = types.RewardProgram_STATE_STARTED
			rewardProgram.CurrentClaimPeriod = 1
		case 3:
			rewardProgram.State = types.RewardProgram_STATE_FINISHED
			rewardProgram.CurrentClaimPeriod = rewardProgram.GetClaimPeriods()
		}

		s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			msg := types.NewMsgEndRewardProgramRequest(tc.id, tc.address)
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			result, err := s.handler(s.ctx, msg)
			if tc.expectErr {
				s.Assert().Error(err)
				s.Assert().Equal(tc.expectErrMsg, err.Error())
			} else {
				s.Assert().NoError(err)
				var response types.MsgEndRewardProgramResponse
				err = response.Unmarshal(result.Data)
				s.Assert().NoError(err)
			}
		})
	}

}
