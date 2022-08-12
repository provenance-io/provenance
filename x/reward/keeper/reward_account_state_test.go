package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/reward/keeper"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"reflect"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (suite *KeeperTestSuite) TestNewRewardAccountState() {
	suite.SetupTest()

	accountState := types.NewRewardAccountState(
		1,
		2,
		"cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv",
		3,
		map[string]uint64{})

	suite.Assert().Equal(uint64(1), accountState.GetRewardProgramId(), "reward program id must match")
	suite.Assert().Equal(uint64(2), accountState.GetClaimPeriodId(), "reward claim period id must match")
	suite.Assert().Equal(uint64(3), accountState.GetSharesEarned(), "earned shares must match")
	suite.Assert().Equal("cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", accountState.GetAddress(), "address must match")
	suite.Assert().Equal(types.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE, accountState.GetClaimStatus(), "should be set to unclaimable initially")
	suite.Assert().True(reflect.DeepEqual(map[string]uint64{}, accountState.GetActionCounter()), "action counter must match")
}

func (suite *KeeperTestSuite) TestGetSetRewardAccountState() {
	suite.SetupTest()

	expectedState := types.NewRewardAccountState(
		1,
		2,
		"cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27",
		3,
		map[string]uint64{},
	)

	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, expectedState)
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
		"cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27")

	suite.Assert().Nil(err, "must not have error")
	suite.Assert().Error(actualState.Validate(), "account state validate basic must return error")
}

func (suite *KeeperTestSuite) TestIterateAccountStates() {
	suite.SetupTest()

	state1 := types.NewRewardAccountState(1, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state1)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state2)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state3)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state4)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state5)

	counter := 0
	suite.app.RewardKeeper.IterateRewardAccountStates(suite.ctx, 2, 2, func(state types.RewardAccountState) bool {
		counter += 1
		return false
	})

	suite.Assert().Equal(2, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestIterateAccountStatesByAddress() {
	suite.SetupTest()

	state1 := types.NewRewardAccountState(1, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state1)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state2)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state3)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state4)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state5)

	counter := 0
	addr, err := sdk.AccAddressFromBech32("cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv")
	suite.Assert().NoError(err, "no error should be thrown")
	suite.app.RewardKeeper.IterateRewardAccountStatesByAddress(suite.ctx, addr, func(state types.RewardAccountState) bool {
		counter += 1
		return false
	})

	suite.Assert().Equal(4, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestEmptyIterateAccountStates() {
	suite.SetupTest()

	state1 := types.NewRewardAccountState(1, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state1)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state2)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state3)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state4)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state5)

	counter := 0
	suite.app.RewardKeeper.IterateRewardAccountStates(suite.ctx, 1, 4, func(state types.RewardAccountState) bool {
		counter += 1
		return false
	})

	suite.Assert().Equal(0, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestIterateAccountStatesHalt() {
	suite.SetupTest()

	state1 := types.NewRewardAccountState(1, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state1)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state2)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state3)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state4)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state5)

	counter := 0
	suite.app.RewardKeeper.IterateRewardAccountStates(suite.ctx, 1, 2, func(state types.RewardAccountState) bool {
		counter += 1
		return counter == 1
	})

	suite.Assert().Equal(1, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestIterateAllAccountStates() {
	suite.SetupTest()

	state1 := types.NewRewardAccountState(1, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state1)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state2)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state3)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state4)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state5)

	counter := 0
	suite.app.RewardKeeper.IterateAllRewardAccountStates(suite.ctx, func(state types.RewardAccountState) bool {
		counter += 1
		return false
	})

	suite.Assert().Equal(5, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestEmptyIterateAllAccountStates() {
	suite.SetupTest()

	counter := 0
	suite.app.RewardKeeper.IterateAllRewardAccountStates(suite.ctx, func(state types.RewardAccountState) bool {
		counter += 1
		return false
	})

	suite.Assert().Equal(0, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestIterateAllAccountStatesHalt() {
	suite.SetupTest()

	state1 := types.NewRewardAccountState(1, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state1)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state2)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state3)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state4)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state5)

	counter := 0
	suite.app.RewardKeeper.IterateAllRewardAccountStates(suite.ctx, func(state types.RewardAccountState) bool {
		counter += 1
		return counter == 1
	})

	suite.Assert().Equal(1, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestIterateRewardAccountStatesForRewardProgram() {
	suite.SetupTest()

	state1 := types.NewRewardAccountState(1, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state1)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state2)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state3)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state4)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state5)

	counter := 0
	suite.app.RewardKeeper.IterateRewardAccountStatesForRewardProgram(suite.ctx, 2, func(state types.RewardAccountState) bool {
		counter += 1
		return false
	})

	suite.Assert().Equal(3, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestEmptyIterateRewardAccountStatesForRewardProgram() {
	suite.SetupTest()

	counter := 0
	suite.app.RewardKeeper.IterateRewardAccountStatesForRewardProgram(suite.ctx, 1, func(state types.RewardAccountState) bool {
		counter += 1
		return false
	})

	suite.Assert().Equal(0, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestIterateRewardAccountStatesForRewardProgramHalt() {
	suite.SetupTest()

	state1 := types.NewRewardAccountState(1, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state1)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state2)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state3)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state4)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state5)

	counter := 0
	suite.app.RewardKeeper.IterateRewardAccountStatesForRewardProgram(suite.ctx, 2, func(state types.RewardAccountState) bool {
		counter += 1
		return counter == 1
	})

	suite.Assert().Equal(1, counter, "should have correct number of iterations")
}

func (suite *KeeperTestSuite) TestGetRewardAccountStatesForClaimPeriod() {
	suite.SetupTest()

	state1 := types.NewRewardAccountState(1, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state1)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state2)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state3)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state4)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state5)

	states, err := suite.app.RewardKeeper.GetRewardAccountStatesForClaimPeriod(suite.ctx, 2, 2)
	suite.Assert().NoError(err, "no error should be thrown when there are account states.")
	suite.Assert().Equal(2, len(states), "should have correct number of account states")
}

func (suite *KeeperTestSuite) TestGetRewardAccountStatesForClaimPeriodHandlesEmpty() {
	suite.SetupTest()

	states, err := suite.app.RewardKeeper.GetRewardAccountStatesForClaimPeriod(suite.ctx, 1, 1)
	suite.Assert().NoError(err, "no error should be thrown when there are no account states.")
	suite.Assert().Equal(0, len(states), "should have no account states")
}

func (suite *KeeperTestSuite) TestGetRewardAccountStatesForRewardProgram() {
	suite.SetupTest()

	state1 := types.NewRewardAccountState(1, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state1)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state2)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state3)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state4)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state5)

	states, err := suite.app.RewardKeeper.GetRewardAccountStatesForRewardProgram(suite.ctx, 2)
	suite.Assert().NoError(err, "no error should be thrown when there are account states.")
	suite.Assert().Equal(3, len(states), "should have correct number of account states")
}

func (suite *KeeperTestSuite) TestGetRewardAccountStatesForRewardProgramHandlesEmpty() {
	suite.SetupTest()

	states, err := suite.app.RewardKeeper.GetRewardAccountStatesForClaimPeriod(suite.ctx, 1, 1)
	suite.Assert().NoError(err, "no error should be thrown when there are no account states.")
	suite.Assert().Equal(0, len(states), "should have no account states")
}

func (suite *KeeperTestSuite) TestMakeRewardClaimsClaimableForPeriod() {
	suite.SetupTest()

	state1 := types.NewRewardAccountState(1, 2, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state1)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state2)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state3)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state4)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state5)

	err := suite.app.RewardKeeper.MakeRewardClaimsClaimableForPeriod(suite.ctx, 2, 2)
	suite.Assert().NoError(err, "no error should be thrown when there are account states.")

	state1, _ = suite.app.RewardKeeper.GetRewardAccountState(suite.ctx, 1, 2, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27")
	state2, _ = suite.app.RewardKeeper.GetRewardAccountState(suite.ctx, 1, 3, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27")
	state3, _ = suite.app.RewardKeeper.GetRewardAccountState(suite.ctx, 2, 1, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27")
	state4, _ = suite.app.RewardKeeper.GetRewardAccountState(suite.ctx, 2, 2, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27")
	state5, _ = suite.app.RewardKeeper.GetRewardAccountState(suite.ctx, 2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")

	suite.Assert().NotEqual(types.RewardAccountState_CLAIM_STATUS_CLAIMABLE, state1.GetClaimStatus(), "account state should not be updated to be claimable")
	suite.Assert().NotEqual(types.RewardAccountState_CLAIM_STATUS_CLAIMABLE, state2.GetClaimStatus(), "account state should not be updated to be claimable")
	suite.Assert().NotEqual(types.RewardAccountState_CLAIM_STATUS_CLAIMABLE, state3.GetClaimStatus(), "account state should not be updated to be claimable")
	suite.Assert().Equal(types.RewardAccountState_CLAIM_STATUS_CLAIMABLE, state4.GetClaimStatus(), "account state should not be updated to be claimable")
	suite.Assert().Equal(types.RewardAccountState_CLAIM_STATUS_CLAIMABLE, state5.GetClaimStatus(), "account state should not be updated to be claimable")
}

func (suite *KeeperTestSuite) TestMakeRewardClaimsClaimableForPeriodHandlesEmpty() {
	suite.SetupTest()

	err := suite.app.RewardKeeper.MakeRewardClaimsClaimableForPeriod(suite.ctx, 1, 1)
	suite.Assert().NoError(err, "no error should be thrown when there are no account states.")
}

func (suite *KeeperTestSuite) TestExpireRewardClaimsForRewardProgram() {
	suite.SetupTest()

	state1 := types.NewRewardAccountState(1, 1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 1, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(1, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(1, 2, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", 0, map[string]uint64{})
	state4.ClaimStatus = types.RewardAccountState_CLAIM_STATUS_CLAIMED

	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state1)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state2)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state3)
	suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state4)

	err := suite.app.RewardKeeper.ExpireRewardClaimsForRewardProgram(suite.ctx, 1)
	suite.Assert().NoError(err, "no error should be thrown when there are account states.")

	state1, _ = suite.app.RewardKeeper.GetRewardAccountState(suite.ctx, 1, 1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	state2, _ = suite.app.RewardKeeper.GetRewardAccountState(suite.ctx, 1, 1, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27")
	state3, _ = suite.app.RewardKeeper.GetRewardAccountState(suite.ctx, 1, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	state4, _ = suite.app.RewardKeeper.GetRewardAccountState(suite.ctx, 1, 2, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27")

	suite.Assert().Equal(types.RewardAccountState_CLAIM_STATUS_EXPIRED, state1.GetClaimStatus(), "account state should be updated to expired")
	suite.Assert().Equal(types.RewardAccountState_CLAIM_STATUS_EXPIRED, state2.GetClaimStatus(), "account state should be updated to expired")
	suite.Assert().Equal(types.RewardAccountState_CLAIM_STATUS_EXPIRED, state3.GetClaimStatus(), "account state should be updated to expired")
	suite.Assert().Equal(types.RewardAccountState_CLAIM_STATUS_CLAIMED, state4.GetClaimStatus(), "account state should not be updated to expired if claimed")
}

func (suite *KeeperTestSuite) TestExpireRewardClaimsForRewardProgramHandlesEmpty() {
	suite.SetupTest()

	err := suite.app.RewardKeeper.ExpireRewardClaimsForRewardProgram(suite.ctx, 1)
	suite.Assert().NoError(err, "no error should be thrown when there are no account states.")
}

func (suite *KeeperTestSuite) TestParseRewardAccountLookUpKey() {
	suite.SetupTest()
	addressFromSec256k1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	rewardProgramId := uint64(123456)
	claimPeriodId := uint64(7891011)
	accountStateAddressLookupKey := types.GetRewardAccountStateAddressLookupKey(addressFromSec256k1, rewardProgramId, claimPeriodId)
	lookup, err := keeper.ParseRewardAccountLookUpKey(accountStateAddressLookupKey, addressFromSec256k1)
	suite.Assert().NoError(err, "no error expected for parsing GetRewardAccountStateAddressLookupKey.")
	suite.Assert().Equal(addressFromSec256k1, lookup.Addr)
	suite.Assert().Equal(rewardProgramId, lookup.RewardID)
	suite.Assert().Equal(claimPeriodId, lookup.ClaimID)
}
