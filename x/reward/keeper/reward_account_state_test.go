package keeper_test

import (
	"github.com/provenance-io/provenance/x/reward/types"
)

func (suite *KeeperTestSuite) TestNewRewardAccountState() {
	suite.SetupTest()

	accountState := types.NewRewardAccountState(
		1,
		2,
		"test",
	)

	suite.Assert().Equal(uint64(1), accountState.GetRewardProgramId(), "reward program id must match")
	suite.Assert().Equal(uint64(2), accountState.GetClaimPeriodId(), "reward claim period id must match")
	suite.Assert().Equal("test", accountState.GetAddress(), "address must match")
	suite.Assert().Equal(uint64(0), accountState.GetActionCounter(), "action counter must match")
}

func (suite *KeeperTestSuite) TestGetSetRewardAccountState() {
	suite.SetupTest()

	expectedState := types.NewRewardAccountState(
		1,
		2,
		"test",
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
	suite.Assert().Equal(expectedState.GetActionCounter(), actualState.GetActionCounter(), "action counter must match")
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
	)

	removed := suite.app.RewardKeeper.RemoveRewardAccountState(suite.ctx,
		expectedState.GetRewardProgramId(),
		expectedState.GetClaimPeriodId(),
		expectedState.GetAddress())

	suite.Assert().False(removed, "account state should be unable to be removed")
}

func (suite *KeeperTestSuite) TestIterateAccountStates() {
	suite.SetupTest()

	state1 := types.NewRewardAccountState(1, 2, "test")
	state2 := types.NewRewardAccountState(1, 3, "test")
	state3 := types.NewRewardAccountState(2, 1, "test")
	state4 := types.NewRewardAccountState(2, 2, "test")
	state5 := types.NewRewardAccountState(2, 2, "test2")

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

	state1 := types.NewRewardAccountState(1, 2, "test")
	state2 := types.NewRewardAccountState(1, 3, "test")
	state3 := types.NewRewardAccountState(2, 1, "test")
	state4 := types.NewRewardAccountState(2, 2, "test")
	state5 := types.NewRewardAccountState(2, 2, "test2")

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

	state1 := types.NewRewardAccountState(1, 2, "test")
	state2 := types.NewRewardAccountState(1, 3, "test")
	state3 := types.NewRewardAccountState(2, 1, "test")
	state4 := types.NewRewardAccountState(2, 2, "test")
	state5 := types.NewRewardAccountState(2, 2, "test2")

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
