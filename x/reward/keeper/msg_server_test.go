package keeper_test

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (s *KeeperTestSuite) TestCreateRewardProgramTransaction() {
	s.SetupTest()
	minimumDelegation := sdk.NewInt64Coin("nhash", 100)
	maximumDelegation := sdk.NewInt64Coin("nhash", 200)
	simapp.FundAccount(s.app.BankKeeper, s.ctx, s.accountAddresses[0], sdk.NewCoins(sdk.NewInt64Coin("nhash", 100000)))

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
	s.SetupTest()
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
	s.Assert().NoError(err, "No error should be returned")
	s.Assert().NotNil(program.Validate(), "should have validation issue for invalid program")
}

func (s *KeeperTestSuite) TestRewardClaimTransaction() {
	s.SetupTest()

	time := s.ctx.BlockTime()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv",
		sdk.NewInt64Coin("nhash", 1000),
		sdk.NewInt64Coin("nhash", 100),
		time,
		10,
		3,
		0,
		uint64(time.Day()),
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
	rewardProgram.State = types.RewardProgram_FINISHED
	rewardProgram.CurrentClaimPeriod = rewardProgram.GetClaimPeriods()
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

	for i := 1; i <= int(rewardProgram.GetClaimPeriods()); i++ {
		state := types.NewRewardAccountState(rewardProgram.GetId(), uint64(i), s.accountAddresses[0].String(), 1)
		state.ClaimStatus = types.RewardAccountState_CLAIMABLE
		s.app.RewardKeeper.SetRewardAccountState(s.ctx, state)
		distribution := types.NewClaimPeriodRewardDistribution(uint64(i), rewardProgram.GetId(), sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, true)
		s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, distribution)
	}

	msg := types.NewMsgClaimRewardRequest(1, s.accountAddresses[0].String())
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
	s.SetupTest()

	msg := types.NewMsgClaimRewardRequest(1, "invalid address")
	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
	result, err := s.handler(s.ctx, msg)

	s.Assert().Error(err, "msg server should handle an invalid reward claim")
	s.Assert().Nil(result, "should have no emitted events")
}

func (s *KeeperTestSuite) TestRewardClaimTransactionInvalidClaimer() {
	s.SetupTest()

	time := s.ctx.BlockTime()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv",
		sdk.NewInt64Coin("nhash", 1000),
		sdk.NewInt64Coin("nhash", 100),
		time,
		10,
		3,
		0,
		uint64(time.Day()),
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
	rewardProgram.State = types.RewardProgram_FINISHED
	rewardProgram.CurrentClaimPeriod = rewardProgram.GetClaimPeriods()
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

	for i := 1; i <= int(rewardProgram.GetClaimPeriods()); i++ {
		state := types.NewRewardAccountState(rewardProgram.GetId(), uint64(i), s.accountAddresses[0].String(), 1)
		state.ClaimStatus = types.RewardAccountState_CLAIMABLE
		s.app.RewardKeeper.SetRewardAccountState(s.ctx, state)
		distribution := types.NewClaimPeriodRewardDistribution(uint64(i), rewardProgram.GetId(), sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, true)
		s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, distribution)
	}

	fmt.Printf("Owner: %s\n", s.accountAddresses[0].String())
	fmt.Printf("Claim Attempter: %s\n", s.accountAddresses[1].String())
	msg := types.NewMsgClaimRewardRequest(1, s.accountAddresses[1].String())
	s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
	result, err := s.handler(s.ctx, msg)
	s.Assert().NoError(err, "msg server should handle valid reward claim")
	s.Assert().NotNil(result, "msg server should emit events")

	var response types.MsgClaimRewardResponse
	response.Unmarshal(result.Data)
	s.Assert().Equal(uint64(1), response.GetRewardProgramId(), "should have correct reward program id")
	s.Assert().Equal(0, len(response.GetClaimedRewardPeriodDetails()), "should have no details")
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 0), response.GetTotalRewardClaim(), "should have no reward claim")
}
