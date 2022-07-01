package keeper_test

import (
	"reflect"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (suite *KeeperTestSuite) TestNewRewardAccountState() {
	suite.SetupTest()

	accountState := types.NewRewardAccountState(
		1,
		2,
		"test",
		3,
	)

	suite.Assert().Equal(uint64(1), accountState.GetRewardProgramId(), "reward program id must match")
	suite.Assert().Equal(uint64(2), accountState.GetClaimPeriodId(), "reward claim period id must match")
	suite.Assert().Equal(uint64(3), accountState.GetSharesEarned(), "earned shares must match")
	suite.Assert().Equal("test", accountState.GetAddress(), "address must match")
	suite.Assert().Equal(types.RewardAccountState_UNCLAIMED, accountState.GetClaimStatus(), "should be set to unclaimed initially")
	suite.Assert().True(reflect.DeepEqual(map[string]uint64{}, accountState.GetActionCounter()), "action counter must match")
}

func (suite *KeeperTestSuite) TestGetSetRewardAccountState() {
	suite.SetupTest()

	expectedState := types.NewRewardAccountState(
		1,
		2,
		"test",
		3,
	)

	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &expectedState)
	actualState, err := suite.app.RewardKeeper.GetRewardAccountState(suite.ctx,
		expectedState.GetRewardProgramId(),
		expectedState.GetClaimPeriodId(),
		expectedState.GetAddress())

	suite.Assert().Nil(err, "must not have error")
	suite.Assert().Equal(expectedState.GetRewardProgramId(), actualState.GetRewardProgramId(), "reward program id must match")
	suite.Assert().Equal(expectedState.GetClaimPeriodId(), actualState.GetClaimPeriodId(), "reward claim period id must match")
	suite.Assert().Equal(expectedState.GetAddress(), actualState.GetAddress(), "address must match")
	suite.Assert().Equal(expectedState.GetSharesEarned(), actualState.GetSharesEarned(), "shares earned must match")
	suite.Assert().Equal(expectedState.GetClaimStatus(), actualState.GetClaimStatus(), "should be set to unclaimed initially")
	suite.Assert().True(reflect.DeepEqual(expectedState.GetActionCounter(), expectedState.GetActionCounter()), "action counter must match")
}

func (suite *KeeperTestSuite) TestGetInvalidAccountState() {
	suite.SetupTest()

	actualState, err := suite.app.RewardKeeper.GetRewardAccountState(suite.ctx,
		99,
		99,
		"jackthecat")

	suite.Assert().Nil(err, "must not have error")
	suite.Assert().Error(actualState.ValidateBasic(), "account state validate basic must return error")
}

func (suite *KeeperTestSuite) TestRemoveValidAccountState() {
	suite.SetupTest()

	expectedState := types.NewRewardAccountState(
		1,
		2,
		"test",
		0,
	)

	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &expectedState)
	removed := suite.app.RewardKeeper.RemoveRewardAccountState(suite.ctx,
		expectedState.GetRewardProgramId(),
		expectedState.GetClaimPeriodId(),
		expectedState.GetAddress())

	actualState, err := suite.app.RewardKeeper.GetRewardAccountState(suite.ctx,
		expectedState.GetRewardProgramId(),
		expectedState.GetClaimPeriodId(),
		expectedState.GetAddress())

	suite.Assert().True(removed, "account state should successfully be removed")
	suite.Assert().Nil(err, "must not have error")
	suite.Assert().Error(actualState.ValidateBasic(), "account state validate basic must return error")
}

func (suite *KeeperTestSuite) TestRemoveInvalidAccountState() {
	suite.SetupTest()

	expectedState := types.NewRewardAccountState(
		1,
		2,
		"test",
		0,
	)

	removed := suite.app.RewardKeeper.RemoveRewardAccountState(suite.ctx,
		expectedState.GetRewardProgramId(),
		expectedState.GetClaimPeriodId(),
		expectedState.GetAddress())

	suite.Assert().False(removed, "account state should be unable to be removed")
}

func (suite *KeeperTestSuite) TestIterateAccountStates() {
	suite.SetupTest()

	state1 := types.NewRewardAccountState(1, 2, "test", 0)
	state2 := types.NewRewardAccountState(1, 3, "test", 0)
	state3 := types.NewRewardAccountState(2, 1, "test", 0)
	state4 := types.NewRewardAccountState(2, 2, "test", 0)
	state5 := types.NewRewardAccountState(2, 2, "test2", 0)

	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state1)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state2)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state3)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state4)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state5)

	counter := 0
	suite.app.RewardKeeper.IterateRewardAccountStates(suite.ctx, 2, 2, func(state types.RewardAccountState) bool {
		counter += 1
		return false
	})

	suite.Assert().Equal(2, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestEmptyIterateAccountStates() {
	suite.SetupTest()

	state1 := types.NewRewardAccountState(1, 2, "test", 0)
	state2 := types.NewRewardAccountState(1, 3, "test", 0)
	state3 := types.NewRewardAccountState(2, 1, "test", 0)
	state4 := types.NewRewardAccountState(2, 2, "test", 0)
	state5 := types.NewRewardAccountState(2, 2, "test2", 0)

	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state1)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state2)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state3)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state4)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state5)

	counter := 0
	suite.app.RewardKeeper.IterateRewardAccountStates(suite.ctx, 1, 4, func(state types.RewardAccountState) bool {
		counter += 1
		return false
	})

	suite.Assert().Equal(0, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestIterateAccountStatesHalt() {
	suite.SetupTest()

	state1 := types.NewRewardAccountState(1, 2, "test", 0)
	state2 := types.NewRewardAccountState(1, 3, "test", 0)
	state3 := types.NewRewardAccountState(2, 1, "test", 0)
	state4 := types.NewRewardAccountState(2, 2, "test", 0)
	state5 := types.NewRewardAccountState(2, 2, "test2", 0)

	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state1)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state2)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state3)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state4)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state5)

	counter := 0
	suite.app.RewardKeeper.IterateRewardAccountStates(suite.ctx, 1, 2, func(state types.RewardAccountState) bool {
		counter += 1
		return counter == 1
	})

	suite.Assert().Equal(1, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestGetRewardAccountStatesForClaimPeriod() {
	suite.SetupTest()

	state1 := types.NewRewardAccountState(1, 1, "test", 0)
	state3 := types.NewRewardAccountState(1, 1, "test2", 0)
	state2 := types.NewRewardAccountState(1, 2, "test", 0)
	state4 := types.NewRewardAccountState(1, 2, "test2", 0)

	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state1)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state2)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state3)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state4)

	results, err := suite.app.RewardKeeper.GetRewardAccountStatesForClaimPeriod(suite.ctx, 1, 1)
	suite.Assert().NoError(err, "no error should be returned in successful case")
	suite.Assert().Equal(2, len(results), "should have every state for the reward program's claim period")
}

func (suite *KeeperTestSuite) TestGetRewardAccountStatesForClaimPeriodInvalid() {
	suite.SetupTest()
	results, err := suite.app.RewardKeeper.GetRewardAccountStatesForClaimPeriod(suite.ctx, 1, 1)
	suite.Assert().NoError(err, "no error should be returned when no states exist")
	suite.Assert().Equal(0, len(results), "should have every state for the reward program's claim period")
}

func (suite *KeeperTestSuite) TestIterateRewardAccountStatesForClaimPeriod() {
	suite.SetupTest()

	state1 := types.NewRewardAccountState(1, 1, "test", 0)
	state3 := types.NewRewardAccountState(1, 1, "test2", 0)
	state2 := types.NewRewardAccountState(1, 2, "test", 0)
	state4 := types.NewRewardAccountState(1, 2, "test2", 0)

	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state1)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state2)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state3)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state4)

	counter := 0
	err := suite.app.RewardKeeper.IterateRewardAccountStatesForClaimPeriod(suite.ctx, 1, 1, func(state types.RewardAccountState) (stop bool) {
		counter += 1
		return false
	})
	suite.Assert().NoError(err, "no error should be returned in successful case")
	suite.Assert().Equal(2, counter, "should iterate every state for the reward program's claim period")
}

func (suite *KeeperTestSuite) TestIterateRewardAccountStatesForClaimPeriodHalt() {
	suite.SetupTest()

	state1 := types.NewRewardAccountState(1, 1, "test", 0)
	state3 := types.NewRewardAccountState(1, 1, "test2", 0)
	state2 := types.NewRewardAccountState(1, 2, "test", 0)
	state4 := types.NewRewardAccountState(1, 2, "test2", 0)

	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state1)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state2)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state3)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, &state4)

	counter := 0
	err := suite.app.RewardKeeper.IterateRewardAccountStatesForClaimPeriod(suite.ctx, 1, 1, func(state types.RewardAccountState) (stop bool) {
		counter += 1
		return true
	})
	suite.Assert().NoError(err, "no error should be returned in successful case")
	suite.Assert().Equal(1, counter, "should only iterate once because of stop")
}

func (suite *KeeperTestSuite) TestIterateRewardAccountStatesForClaimPeriodEmpty() {
	suite.SetupTest()

	counter := 0
	err := suite.app.RewardKeeper.IterateRewardAccountStatesForClaimPeriod(suite.ctx, 1, 1, func(state types.RewardAccountState) (stop bool) {
		counter += 1
		return false
	})
	suite.Assert().NoError(err, "no error should be returned in successful case")
	suite.Assert().Equal(0, counter, "should not iterate when empty")
}
