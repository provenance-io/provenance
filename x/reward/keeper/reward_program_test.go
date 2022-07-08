package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/provenance/x/reward/types"
)

func TestRewardProgramKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestNewRewardProgram() {
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
		[]types.QualifyingAction{},
	)

	suite.Assert().Equal("title", program.GetTitle(), "title should match input")
	suite.Assert().Equal("description", program.GetDescription(), "description should match input")
	suite.Assert().Equal(uint64(1), program.GetId(), "id should match input")
	suite.Assert().Equal("insert address", program.GetDistributeFromAddress(), "address should match input")
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 100000), program.GetTotalRewardPool(), "coin should match input")
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 1000), program.GetMaxRewardByAddress(), "max reward by address should match")
	suite.Assert().Equal(time, program.GetProgramStartTime(), "program start time should match input")
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
	time := time.Now()
	program1 := types.NewRewardProgram(
		"title",
		"description",
		1,
		suite.accountAddresses[0].String(),
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		time,
		60*60,
		3,
		0,
		[]types.QualifyingAction{},
	)
	err = suite.app.RewardKeeper.CreateRewardProgram(suite.ctx, program1)
	suite.Assert().NoError(err)
	actualProgram, err := suite.app.RewardKeeper.GetRewardProgram(suite.ctx, uint64(1))
	suite.Assert().NoError(err)
	suite.Equal(uint64(1), actualProgram.Id)
}

func (suite *KeeperTestSuite) TestRefundRemainingBalance() {
	suite.SetupTest()
	suite.Assert().Fail("not yet implemented")
}

func (suite *KeeperTestSuite) TestRefundRemainingBalanceEmpty() {
	suite.SetupTest()
	suite.Assert().Fail("not yet implemented")
}
