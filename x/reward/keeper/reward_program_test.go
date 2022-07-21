package keeper_test

import (
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (suite *KeeperTestSuite) TestNewRewardProgram() {
	suite.SetupTest()

	time := time.Now().UTC()
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		time,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)

	suite.Assert().Equal("title", program.GetTitle(), "title should match input")
	suite.Assert().Equal("description", program.GetDescription(), "description should match input")
	suite.Assert().Equal(uint64(1), program.GetId(), "id should match input")
	suite.Assert().Equal("insert address", program.GetDistributeFromAddress(), "address should match input")
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 100000), program.GetTotalRewardPool(), "coin should match input")
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 1000), program.GetMaxRewardByAddress(), "max reward by address should match")
	suite.Assert().Equal(time.UTC(), program.GetProgramStartTime(), "program start time should match input")
	suite.Assert().Equal(uint64(60*60), program.GetClaimPeriodSeconds(), "claim period seconds should match input")
	suite.Assert().Equal(uint64(3), program.GetClaimPeriods(), "claim periods should match input")
	suite.Assert().Equal(0, len(program.GetQualifyingActions()), "qualifying actions should match input")
}

func (suite *KeeperTestSuite) TestGetSetRewardProgram() {
	suite.SetupTest()

	time := time.Now().Local().UTC()
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		time,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)

	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program)
	program2, err := suite.app.RewardKeeper.GetRewardProgram(suite.ctx, 1)

	suite.Assert().NoError(err, "no error should be returned when getting reward program")

	suite.Assert().Equal(program.GetTitle(), program2.GetTitle(), "title should match")
	suite.Assert().Equal(program.GetDescription(), program2.GetDescription(), "description should match")
	suite.Assert().Equal(program.GetId(), program2.GetId(), "id should match")
	suite.Assert().Equal(program.GetDistributeFromAddress(), program2.GetDistributeFromAddress(), "address should match")
	suite.Assert().Equal(program.GetTotalRewardPool(), program2.GetTotalRewardPool(), "coin should match")
	suite.Assert().Equal(program.GetMaxRewardByAddress(), program2.GetMaxRewardByAddress(), "max reward by address should")
	suite.Assert().Equal(program.GetProgramStartTime(), program2.GetProgramStartTime(), "program start time should match")
	suite.Assert().Equal(program.GetClaimPeriodSeconds(), program2.GetClaimPeriodSeconds(), "claim period seconds should match")
	suite.Assert().Equal(program.GetClaimPeriods(), program2.GetClaimPeriods(), "number of claim periods should match")
	suite.Assert().Equal(len(program.GetQualifyingActions()), len(program2.GetQualifyingActions()), "qualifying actions should match")
}

func (suite *KeeperTestSuite) TestRemoveValidRewardProgram() {
	suite.SetupTest()

	time := time.Now()
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		time,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)

	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program)
	removed := suite.app.RewardKeeper.RemoveRewardProgram(suite.ctx, 1)
	suite.Assert().True(removed, "remove should succeed")

	invalidProgram, err := suite.app.RewardKeeper.GetRewardProgram(suite.ctx, 1)
	suite.Assert().NoError(err)
	suite.Assert().Equal(uint64(0), invalidProgram.Id)
}

func (suite *KeeperTestSuite) TestRemoveInvalidRewardProgram() {
	suite.SetupTest()
	invalidProgram, err := suite.app.RewardKeeper.GetRewardProgram(suite.ctx, 1)
	suite.Assert().NoError(err)
	suite.Assert().Equal(uint64(0), invalidProgram.Id)
}

func (suite *KeeperTestSuite) TestIterateRewardPrograms() {
	suite.SetupTest()
	time := time.Now()
	program1 := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		time,
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
		time,
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
		time,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)

	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program1)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program2)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program3)

	counter := 0
	err := suite.app.RewardKeeper.IterateRewardPrograms(suite.ctx, func(rewardProgram types.RewardProgram) (stop bool) {
		counter += 1
		return false
	})
	suite.Assert().NoError(err, "no error should be returned")
	suite.Assert().Equal(3, counter, "should iterate through each reward program")
}

func (suite *KeeperTestSuite) TestIterateRewardProgramsHalt() {
	suite.SetupTest()
	time := time.Now()
	program1 := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		time,
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
		time,
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
		time,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)

	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program1)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program2)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program3)

	counter := 0
	err := suite.app.RewardKeeper.IterateRewardPrograms(suite.ctx, func(rewardProgram types.RewardProgram) (stop bool) {
		counter += 1
		return true
	})
	suite.Assert().NoError(err, "no error should be returned")
	suite.Assert().Equal(1, counter, "should stop when iteration is instructed to stop")
}

func (suite *KeeperTestSuite) TestIterateRewardProgramsEmpty() {
	suite.SetupTest()

	counter := 0
	err := suite.app.RewardKeeper.IterateRewardPrograms(suite.ctx, func(rewardProgram types.RewardProgram) (stop bool) {
		counter += 1
		return true
	})

	suite.Assert().NoError(err, "no error should be returned")
	suite.Assert().Equal(0, counter, "should stop when iteration is instructed to stop")
}

func (suite *KeeperTestSuite) TestGetOutstandingRewardPrograms() {
	suite.SetupTest()
	time := time.Now()
	program1 := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		time,
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
		time,
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
		time,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program2.State = types.RewardProgram_STARTED
	program3.State = types.RewardProgram_FINISHED

	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program1)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program2)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program3)

	programs, err := suite.app.RewardKeeper.GetOutstandingRewardPrograms(suite.ctx)
	suite.Assert().NoError(err, "no error should be returned")
	suite.Assert().Equal(2, len(programs), "should have all outstanding programs")
	suite.Assert().Equal(uint64(1), programs[0].GetId(), "should have program 1")
	suite.Assert().Equal(uint64(2), programs[1].GetId(), "should have program 2")
}

func (suite *KeeperTestSuite) TestGetOutstandingRewardProgramsEmpty() {
	suite.SetupTest()
	programs, err := suite.app.RewardKeeper.GetOutstandingRewardPrograms(suite.ctx)
	suite.Assert().NoError(err, "no error should be returned")
	suite.Assert().Equal(0, len(programs), "should have no outstanding programs")
}

func (suite *KeeperTestSuite) TestGetAllExpiredRewardPrograms() {
	suite.SetupTest()
	time := time.Now()
	program1 := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		time,
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
		time,
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
		time,
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
		time,
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
		time,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program2.State = types.RewardProgram_STARTED
	program3.State = types.RewardProgram_FINISHED
	program4.State = types.RewardProgram_EXPIRED
	program5.State = types.RewardProgram_EXPIRED

	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program1)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program2)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program3)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program4)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program5)

	programs, err := suite.app.RewardKeeper.GetAllExpiredRewardPrograms(suite.ctx)
	suite.Assert().NoError(err, "no error should be returned")
	suite.Assert().Equal(2, len(programs), "should have all outstanding programs")
	suite.Assert().Equal(uint64(4), programs[0].GetId(), "should have program 4")
	suite.Assert().Equal(uint64(5), programs[1].GetId(), "should have program 5")
}

func (suite *KeeperTestSuite) TestGetAllExpiredRewardProgramsEmpty() {
	suite.SetupTest()
	programs, err := suite.app.RewardKeeper.GetAllExpiredRewardPrograms(suite.ctx)
	suite.Assert().NoError(err, "no error should be returned")
	suite.Assert().Equal(0, len(programs), "should have no expired programs")
}

func (suite *KeeperTestSuite) TestGetUnexpiredRewardPrograms() {
	suite.SetupTest()
	time := time.Now()
	program1 := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		time,
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
		time,
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
		time,
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
		time,
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
		time,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program2.State = types.RewardProgram_STARTED
	program3.State = types.RewardProgram_FINISHED
	program4.State = types.RewardProgram_EXPIRED
	program5.State = types.RewardProgram_EXPIRED

	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program1)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program2)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program3)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program4)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program5)

	programs, err := suite.app.RewardKeeper.GetUnexpiredRewardPrograms(suite.ctx)
	suite.Assert().NoError(err, "no error should be returned")
	suite.Assert().Equal(3, len(programs), "should have all unexpired programs")
	suite.Assert().Equal(uint64(1), programs[0].GetId(), "should have program 1")
	suite.Assert().Equal(uint64(2), programs[1].GetId(), "should have program 2")
	suite.Assert().Equal(uint64(3), programs[2].GetId(), "should have program 3")
}

func (suite *KeeperTestSuite) TestGetUnexpiredRewardProgramsEmpty() {
	suite.SetupTest()
	programs, err := suite.app.RewardKeeper.GetUnexpiredRewardPrograms(suite.ctx)
	suite.Assert().NoError(err, "no error should be returned")
	suite.Assert().Equal(0, len(programs), "should have no expired programs")
}

func (suite *KeeperTestSuite) TestGetAllActiveRewardPrograms() {
	suite.SetupTest()
	time := time.Now()
	program1 := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		time,
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
		time,
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
		time,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program2.State = types.RewardProgram_STARTED
	program3.State = types.RewardProgram_FINISHED

	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program1)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program2)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program3)

	programs, err := suite.app.RewardKeeper.GetAllActiveRewardPrograms(suite.ctx)
	suite.Assert().NoError(err, "no error should be returned")
	suite.Assert().Equal(1, len(programs), "should have all active programs")
	suite.Assert().Equal(uint64(2), programs[0].GetId(), "should have program 2")
}

func (suite *KeeperTestSuite) TestGetAllActiveRewardProgramsEmpty() {
	suite.SetupTest()
	programs, err := suite.app.RewardKeeper.GetAllActiveRewardPrograms(suite.ctx)
	suite.Assert().NoError(err, "no error should be returned")
	suite.Assert().Equal(0, len(programs), "should have no active programs")
}

func (suite *KeeperTestSuite) TestGetAllRewardPrograms() {
	suite.SetupTest()
	time := time.Now()
	program1 := types.NewRewardProgram(
		"title",
		"description",
		1,
		"insert address",
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		time,
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
		time,
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
		time,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{},
	)
	program2.State = types.RewardProgram_STARTED
	program3.State = types.RewardProgram_FINISHED

	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program1)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program2)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, program3)

	programs, err := suite.app.RewardKeeper.GetAllRewardPrograms(suite.ctx)
	suite.Assert().NoError(err, "no error should be returned")
	suite.Assert().Equal(3, len(programs), "should have all active programs")
	suite.Assert().Equal(uint64(1), programs[0].GetId(), "should have program 1")
	suite.Assert().Equal(uint64(2), programs[1].GetId(), "should have program 2")
	suite.Assert().Equal(uint64(3), programs[2].GetId(), "should have program 3")
}

func (suite *KeeperTestSuite) TestGetAllRewardProgramsEmpty() {
	suite.SetupTest()
	programs, err := suite.app.RewardKeeper.GetAllRewardPrograms(suite.ctx)
	suite.Assert().NoError(err, "no error should be returned")
	suite.Assert().Equal(0, len(programs), "should have no active programs")
}

func (suite *KeeperTestSuite) TestCreateRewardProgram() {
	suite.SetupTest()
	simapp.FundAccount(suite.app.BankKeeper, suite.ctx, suite.accountAddresses[0], sdk.NewCoins(sdk.NewInt64Coin("nhash", 1000000000000)))

	err := suite.app.RewardKeeper.CreateRewardProgram(suite.ctx, types.RewardProgram{})
	suite.Assert().Error(err)

	now := time.Now()
	validProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		suite.accountAddresses[0].String(),
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
	err = suite.app.RewardKeeper.CreateRewardProgram(suite.ctx, validProgram)
	suite.Assert().NoError(err)
	actualProgram, err := suite.app.RewardKeeper.GetRewardProgram(suite.ctx, uint64(1))
	suite.Assert().NoError(err)
	suite.Equal(uint64(1), actualProgram.Id)
	lastYear := now.Add(-60 * 60 * 365 * time.Second)
	inValidProgramStartTime := types.NewRewardProgram(
		"title",
		"description",
		2,
		suite.accountAddresses[0].String(),
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
	err = suite.app.RewardKeeper.CreateRewardProgram(suite.ctx, inValidProgramStartTime)
	suite.Assert().Error(err)
	suite.Assert().True(strings.Contains(err.Error(), "start time is before current block time"))

	minDelegation := sdk.NewInt64Coin("nhash", 4)
	maxDelegation := sdk.NewInt64Coin("nhash", 40)

	invalidAmount := types.NewRewardProgram(
		"title",
		"description",
		2,
		suite.accountAddresses[0].String(),
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
	err = suite.app.RewardKeeper.CreateRewardProgram(suite.ctx, invalidAmount)
	suite.Assert().Error(err)
	suite.Assert().Equal("unable to send coin to module reward pool : 999999900000nhash is smaller than 10000000000000nhash: insufficient funds", err.Error())
}

func (suite *KeeperTestSuite) TestRefundRemainingBalance() {
	suite.SetupTest()
	time := suite.ctx.BlockTime()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv",
		sdk.NewInt64Coin("nhash", 1000),
		sdk.NewInt64Coin("nhash", 100),
		time,
		10,
		5,
		0,
		uint64(time.Day()),
		[]types.QualifyingAction{},
	)
	remainingBalance := rewardProgram.GetTotalRewardPool()
	rewardProgram.RemainingPoolBalance = remainingBalance
	rewardProgram.ClaimedAmount = sdk.NewInt64Coin("nhash", 0)

	addr, _ := sdk.AccAddressFromBech32("cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv")
	beforeBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, "nhash")
	err := suite.app.RewardKeeper.RefundRemainingBalance(suite.ctx, &rewardProgram)
	afterBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, "nhash")

	suite.Assert().NoError(err, "no error should be thrown")
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 0), rewardProgram.GetRemainingPoolBalance(), "no remaining balance should be left")
	suite.Assert().Equal(beforeBalance.Add(remainingBalance), afterBalance, "balance should be given remaining pool balance")
}

func (suite *KeeperTestSuite) TestRefundRemainingBalanceEmpty() {
	suite.SetupTest()
	time := suite.ctx.BlockTime()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv",
		sdk.NewInt64Coin("nhash", 1000),
		sdk.NewInt64Coin("nhash", 100),
		time,
		10,
		5,
		0,
		uint64(time.Day()),
		[]types.QualifyingAction{},
	)
	rewardProgram.RemainingPoolBalance = sdk.NewInt64Coin("nhash", 0)
	rewardProgram.ClaimedAmount = sdk.NewInt64Coin("nhash", 0)

	addr, _ := sdk.AccAddressFromBech32("cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv")
	beforeBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, "nhash")
	err := suite.app.RewardKeeper.RefundRemainingBalance(suite.ctx, &rewardProgram)
	afterBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, "nhash")

	suite.Assert().NoError(err, "no error should be thrown")
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 0), rewardProgram.GetRemainingPoolBalance(), "no remaining balance should be left")
	suite.Assert().Equal(beforeBalance, afterBalance, "balance should remain same because there is no remaining pool balance")
}
