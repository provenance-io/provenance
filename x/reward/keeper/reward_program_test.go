package keeper_test

import (
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (s *KeeperTestSuite) TestNewRewardProgram() {
	now := time.Now().UTC()
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)

	s.Assert().Equal("title", program.GetTitle(), "title should match input")
	s.Assert().Equal("description", program.GetDescription(), "description should match input")
	s.Assert().Equal(uint64(1), program.GetId(), "id should match input")
	s.Assert().Equal("insert address", program.GetDistributeFromAddress(), "address should match input")
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 100000), program.GetTotalRewardPool(), "coin should match input")
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 1000), program.GetMaxRewardByAddress(), "max reward by address should match")
	s.Assert().Equal(now.UTC(), program.GetProgramStartTime(), "program start time should match input")
	s.Assert().Equal(uint64(60*60), program.GetClaimPeriodSeconds(), "claim period seconds should match input")
	s.Assert().Equal(uint64(3), program.GetClaimPeriods(), "claim periods should match input")
	s.Assert().Equal(0, len(program.GetQualifyingActions()), "qualifying actions should match input")
}

func (s *KeeperTestSuite) TestGetSetRewardProgram() {
	now := time.Now().Local().UTC()
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)

	s.app.RewardKeeper.SetRewardProgram(s.ctx, program)
	program2, err := s.app.RewardKeeper.GetRewardProgram(s.ctx, 1)

	s.Assert().NoError(err, "no error should be returned when getting reward program")

	s.Assert().Equal(program.GetTitle(), program2.GetTitle(), "title should match")
	s.Assert().Equal(program.GetDescription(), program2.GetDescription(), "description should match")
	s.Assert().Equal(program.GetId(), program2.GetId(), "id should match")
	s.Assert().Equal(program.GetDistributeFromAddress(), program2.GetDistributeFromAddress(), "address should match")
	s.Assert().Equal(program.GetTotalRewardPool(), program2.GetTotalRewardPool(), "coin should match")
	s.Assert().Equal(program.GetMaxRewardByAddress(), program2.GetMaxRewardByAddress(), "max reward by address should")
	s.Assert().Equal(program.GetProgramStartTime(), program2.GetProgramStartTime(), "program start time should match")
	s.Assert().Equal(program.GetClaimPeriodSeconds(), program2.GetClaimPeriodSeconds(), "claim period seconds should match")
	s.Assert().Equal(program.GetClaimPeriods(), program2.GetClaimPeriods(), "number of claim periods should match")
	s.Assert().Equal(len(program.GetQualifyingActions()), len(program2.GetQualifyingActions()), "qualifying actions should match")
}

func (s *KeeperTestSuite) TestEndingRewardProgram() {
	now := time.Now()
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		3,
		0,
		[]types.QualifyingAction{},
	)
	program.State = types.RewardProgram_STATE_STARTED
	program.Id = 10

	program.CurrentClaimPeriod = 2
	program.ClaimPeriodEndTime = now
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program)
	s.app.RewardKeeper.EndingRewardProgram(s.ctx, program)

	endingRewardProgram, err := s.app.RewardKeeper.GetRewardProgram(s.ctx, 10)
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(2), endingRewardProgram.ClaimPeriods)
	s.Assert().Equal(uint64(0), endingRewardProgram.MaxRolloverClaimPeriods)
	s.Assert().Equal(now.UTC(), endingRewardProgram.ExpectedProgramEndTime)
	s.Assert().Equal(now.UTC(), endingRewardProgram.ProgramEndTimeMax)

	program.State = types.RewardProgram_STATE_PENDING
	program.Id = 20
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program)
	endingRewardProgram, err = s.app.RewardKeeper.GetRewardProgram(s.ctx, 20)
	s.Assert().NoError(err)
	s.Assert().Equal(uint64(20), endingRewardProgram.Id)

	s.app.RewardKeeper.EndingRewardProgram(s.ctx, program)
	endingRewardProgram, err = s.app.RewardKeeper.GetRewardProgram(s.ctx, 20)
	s.Assert().Error(err)
	s.Assert().Equal(uint64(0), endingRewardProgram.Id)
}

func (s *KeeperTestSuite) TestRemoveValidRewardProgram() {
	now := time.Now()
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)

	s.app.RewardKeeper.SetRewardProgram(s.ctx, program)
	removed := s.app.RewardKeeper.RemoveRewardProgram(s.ctx, 1)
	s.Assert().True(removed, "remove should succeed")

	invalidProgram, err := s.app.RewardKeeper.GetRewardProgram(s.ctx, 1)
	s.Assert().Error(err)
	s.Assert().Equal(uint64(0), invalidProgram.Id)
}

func (s *KeeperTestSuite) TestRemoveInvalidRewardProgram() {
	invalidProgram, err := s.app.RewardKeeper.GetRewardProgram(s.ctx, 1)
	s.Assert().Error(err)
	s.Assert().Equal(uint64(0), invalidProgram.Id)
}

func (s *KeeperTestSuite) TestIterateRewardPrograms() {
	now := time.Now()
	program1 := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program2 := types.NewRewardProgram(
		"title",
		"description",
		2,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program3 := types.NewRewardProgram(
		"title",
		"description",
		3,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)

	s.app.RewardKeeper.SetRewardProgram(s.ctx, program1)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program2)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program3)

	counter := 0
	err := s.app.RewardKeeper.IterateRewardPrograms(s.ctx, func(rewardProgram types.RewardProgram) (stop bool, err error) {
		counter += 1
		return false, nil
	})
	s.Assert().NoError(err, "no error should be returned")
	s.Assert().Equal(3, counter, "should iterate through each reward program")
}

func (s *KeeperTestSuite) TestIterateRewardProgramsHalt() {
	now := time.Now()
	program1 := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program2 := types.NewRewardProgram(
		"title",
		"description",
		2,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program3 := types.NewRewardProgram(
		"title",
		"description",
		3,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)

	s.app.RewardKeeper.SetRewardProgram(s.ctx, program1)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program2)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program3)

	counter := 0
	err := s.app.RewardKeeper.IterateRewardPrograms(s.ctx, func(rewardProgram types.RewardProgram) (stop bool, err error) {
		counter += 1
		return true, nil
	})
	s.Assert().NoError(err, "no error should be returned")
	s.Assert().Equal(1, counter, "should stop when iteration is instructed to stop")
}

func (s *KeeperTestSuite) TestIterateRewardProgramsEmpty() {
	counter := 0
	err := s.app.RewardKeeper.IterateRewardPrograms(s.ctx, func(rewardProgram types.RewardProgram) (stop bool, err error) {
		counter += 1
		return true, nil
	})

	s.Assert().NoError(err, "no error should be returned")
	s.Assert().Equal(0, counter, "should stop when iteration is instructed to stop")
}

func (s *KeeperTestSuite) TestGetAllOutstandingRewardPrograms() {
	now := time.Now()
	program1 := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program2 := types.NewRewardProgram(
		"title",
		"description",
		2,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program3 := types.NewRewardProgram(
		"title",
		"description",
		3,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program2.State = types.RewardProgram_STATE_STARTED
	program3.State = types.RewardProgram_STATE_FINISHED

	s.app.RewardKeeper.SetRewardProgram(s.ctx, program1)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program2)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program3)

	programs, err := s.app.RewardKeeper.GetAllOutstandingRewardPrograms(s.ctx)
	s.Assert().NoError(err, "no error should be returned")
	s.Assert().Equal(2, len(programs), "should have all outstanding programs")
	s.Assert().Equal(uint64(1), programs[0].GetId(), "should have program 1")
	s.Assert().Equal(uint64(2), programs[1].GetId(), "should have program 2")
}

func (s *KeeperTestSuite) TestGetAllOutstandingRewardProgramsEmpty() {
	programs, err := s.app.RewardKeeper.GetAllOutstandingRewardPrograms(s.ctx)
	s.Assert().NoError(err, "no error should be returned")
	s.Assert().Equal(0, len(programs), "should have no outstanding programs")
}

func (s *KeeperTestSuite) TestGetAllExpiredRewardPrograms() {
	now := time.Now()
	program1 := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program2 := types.NewRewardProgram(
		"title",
		"description",
		2,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program3 := types.NewRewardProgram(
		"title",
		"description",
		3,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program4 := types.NewRewardProgram(
		"title",
		"description",
		4,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program5 := types.NewRewardProgram(
		"title",
		"description",
		5,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program2.State = types.RewardProgram_STATE_STARTED
	program3.State = types.RewardProgram_STATE_FINISHED
	program4.State = types.RewardProgram_STATE_EXPIRED
	program5.State = types.RewardProgram_STATE_EXPIRED

	s.app.RewardKeeper.SetRewardProgram(s.ctx, program1)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program2)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program3)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program4)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program5)

	programs, err := s.app.RewardKeeper.GetAllExpiredRewardPrograms(s.ctx)
	s.Assert().NoError(err, "no error should be returned")
	s.Assert().Equal(2, len(programs), "should have all outstanding programs")
	s.Assert().Equal(uint64(4), programs[0].GetId(), "should have program 4")
	s.Assert().Equal(uint64(5), programs[1].GetId(), "should have program 5")
}

func (s *KeeperTestSuite) TestGetAllExpiredRewardProgramsEmpty() {
	programs, err := s.app.RewardKeeper.GetAllExpiredRewardPrograms(s.ctx)
	s.Assert().NoError(err, "no error should be returned")
	s.Assert().Equal(0, len(programs), "should have no expired programs")
}

func (s *KeeperTestSuite) TestGetAllUnexpiredRewardPrograms() {
	now := time.Now()
	program1 := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program2 := types.NewRewardProgram(
		"title",
		"description",
		2,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program3 := types.NewRewardProgram(
		"title",
		"description",
		3,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program4 := types.NewRewardProgram(
		"title",
		"description",
		4,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program5 := types.NewRewardProgram(
		"title",
		"description",
		5,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program2.State = types.RewardProgram_STATE_STARTED
	program3.State = types.RewardProgram_STATE_FINISHED
	program4.State = types.RewardProgram_STATE_EXPIRED
	program5.State = types.RewardProgram_STATE_EXPIRED

	s.app.RewardKeeper.SetRewardProgram(s.ctx, program1)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program2)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program3)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program4)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program5)

	programs, err := s.app.RewardKeeper.GetAllUnexpiredRewardPrograms(s.ctx)
	s.Assert().NoError(err, "no error should be returned")
	s.Assert().Equal(3, len(programs), "should have all unexpired programs")
	s.Assert().Equal(uint64(1), programs[0].GetId(), "should have program 1")
	s.Assert().Equal(uint64(2), programs[1].GetId(), "should have program 2")
	s.Assert().Equal(uint64(3), programs[2].GetId(), "should have program 3")
}

func (s *KeeperTestSuite) TestGetAllUnexpiredRewardProgramsEmpty() {
	programs, err := s.app.RewardKeeper.GetAllUnexpiredRewardPrograms(s.ctx)
	s.Assert().NoError(err, "no error should be returned")
	s.Assert().Equal(0, len(programs), "should have no expired programs")
}

func (s *KeeperTestSuite) TestGetAllActiveRewardPrograms() {
	now := time.Now()
	program1 := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program2 := types.NewRewardProgram(
		"title",
		"description",
		2,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program3 := types.NewRewardProgram(
		"title",
		"description",
		3,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program2.State = types.RewardProgram_STATE_STARTED
	program3.State = types.RewardProgram_STATE_FINISHED

	s.app.RewardKeeper.SetRewardProgram(s.ctx, program1)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program2)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program3)

	programs, err := s.app.RewardKeeper.GetAllActiveRewardPrograms(s.ctx)
	s.Assert().NoError(err, "no error should be returned")
	s.Assert().Equal(1, len(programs), "should have all active programs")
	s.Assert().Equal(uint64(2), programs[0].GetId(), "should have program 2")
}

func (s *KeeperTestSuite) TestGetAllActiveRewardProgramsEmpty() {
	programs, err := s.app.RewardKeeper.GetAllActiveRewardPrograms(s.ctx)
	s.Assert().NoError(err, "no error should be returned")
	s.Assert().Equal(0, len(programs), "should have no active programs")
}

func (s *KeeperTestSuite) TestGetAllRewardPrograms() {
	now := time.Now()
	program1 := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program2 := types.NewRewardProgram(
		"title",
		"description",
		2,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program3 := types.NewRewardProgram(
		"title",
		"description",
		3,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program2.State = types.RewardProgram_STATE_STARTED
	program3.State = types.RewardProgram_STATE_FINISHED

	s.app.RewardKeeper.SetRewardProgram(s.ctx, program1)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program2)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, program3)

	programs, err := s.app.RewardKeeper.GetAllRewardPrograms(s.ctx)
	s.Assert().NoError(err, "no error should be returned")
	s.Assert().Equal(3, len(programs), "should have all active programs")
	s.Assert().Equal(uint64(1), programs[0].GetId(), "should have program 1")
	s.Assert().Equal(uint64(2), programs[1].GetId(), "should have program 2")
	s.Assert().Equal(uint64(3), programs[2].GetId(), "should have program 3")
}

func (s *KeeperTestSuite) TestGetAllRewardProgramsEmpty() {
	programs, err := s.app.RewardKeeper.GetAllRewardPrograms(s.ctx)
	s.Assert().NoError(err, "no error should be returned")
	s.Assert().Equal(0, len(programs), "should have no active programs")
}

func (s *KeeperTestSuite) TestCreateRewardProgram() {
	testutil.FundAccount(s.app.BankKeeper, s.ctx, s.accountAddresses[0], sdk.NewCoins(sdk.NewInt64Coin("nhash", 1000000000000)))

	err := s.app.RewardKeeper.CreateRewardProgram(s.ctx, types.RewardProgram{})
	s.Assert().Error(err)

	now := time.Now()
	validProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		s.accountAddresses[0].String(),
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
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
	err = s.app.RewardKeeper.CreateRewardProgram(s.ctx, validProgram)
	s.Assert().NoError(err)
	actualProgram, err := s.app.RewardKeeper.GetRewardProgram(s.ctx, uint64(1))
	s.Assert().NoError(err)
	s.Equal(uint64(1), actualProgram.Id)
	lastYear := now.Add(-60 * 60 * 365 * time.Second)
	inValidProgramStartTime := types.NewRewardProgram(
		"title",
		"description",
		2,
		s.accountAddresses[0].String(),
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		lastYear,
		60*60,
		3,
		0,
		0,
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
	err = s.app.RewardKeeper.CreateRewardProgram(s.ctx, inValidProgramStartTime)
	s.Assert().Error(err)
	s.Assert().True(strings.Contains(err.Error(), "start time is before current block time"))

	minDelegation := sdk.NewInt64Coin("nhash", 4)
	maxDelegation := sdk.NewInt64Coin("nhash", 40)

	invalidAmount := types.NewRewardProgram(
		"title",
		"description",
		2,
		s.accountAddresses[0].String(),
		sdk.NewInt64Coin("nhash", 10000000000000),
		sdk.NewInt64Coin("nhash", 1000),
		now,
		60*60,
		3,
		0,
		0,
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
	err = s.app.RewardKeeper.CreateRewardProgram(s.ctx, invalidAmount)
	s.Assert().Error(err)
	s.Assert().True(strings.Contains(err.Error(), "unable to send coin to module reward pool : 999999900000nhash is smaller than 10000000000000nhash: insufficient funds"))
}

func (s *KeeperTestSuite) TestRefundRemainingBalance() {
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
		5,
		0,
		uint64(now.Day()),
		[]types.QualifyingAction{},
	)
	remainingBalance := rewardProgram.GetTotalRewardPool()
	rewardProgram.RemainingPoolBalance = remainingBalance
	rewardProgram.ClaimedAmount = sdk.NewInt64Coin("nhash", 0)

	addr, _ := sdk.AccAddressFromBech32("cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv")
	beforeBalance := s.app.BankKeeper.GetBalance(s.ctx, addr, "nhash")
	err := s.app.RewardKeeper.RefundRemainingBalance(s.ctx, &rewardProgram)
	afterBalance := s.app.BankKeeper.GetBalance(s.ctx, addr, "nhash")

	s.Assert().NoError(err, "no error should be thrown")
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 0), rewardProgram.GetRemainingPoolBalance(), "no remaining balance should be left")
	s.Assert().Equal(beforeBalance.Add(remainingBalance), afterBalance, "balance should be given remaining pool balance")
}

func (s *KeeperTestSuite) TestRefundRemainingBalanceEmpty() {
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
		5,
		0,
		uint64(now.Day()),
		[]types.QualifyingAction{},
	)
	rewardProgram.RemainingPoolBalance = sdk.NewInt64Coin("nhash", 0)
	rewardProgram.ClaimedAmount = sdk.NewInt64Coin("nhash", 0)

	addr, _ := sdk.AccAddressFromBech32("cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv")
	beforeBalance := s.app.BankKeeper.GetBalance(s.ctx, addr, "nhash")
	err := s.app.RewardKeeper.RefundRemainingBalance(s.ctx, &rewardProgram)
	afterBalance := s.app.BankKeeper.GetBalance(s.ctx, addr, "nhash")

	s.Assert().NoError(err, "no error should be thrown")
	s.Assert().Equal(sdk.NewInt64Coin("nhash", 0), rewardProgram.GetRemainingPoolBalance(), "no remaining balance should be left")
	s.Assert().Equal(beforeBalance, afterBalance, "balance should remain same because there is no remaining pool balance")
}

func (s *KeeperTestSuite) TestGetRewardProgramID() {
	id, err := s.app.RewardKeeper.GetRewardProgramID(s.ctx)
	s.Assert().NoError(err, "no error should be thrown")
	s.Assert().Equal(uint64(1), id, "id should match")

	next, err := s.app.RewardKeeper.GetNextRewardProgramID(s.ctx)
	s.Assert().NoError(err, "no error should be thrown")
	s.Assert().Equal(uint64(1), next, "id should match")

	id, err = s.app.RewardKeeper.GetRewardProgramID(s.ctx)
	s.Assert().NoError(err, "no error should be thrown")
	s.Assert().Equal(uint64(2), id, "id should match")
}
