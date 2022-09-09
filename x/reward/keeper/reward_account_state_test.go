package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (s *KeeperTestSuite) TestNewRewardAccountState() {
	accountState := types.NewRewardAccountState(
		1,
		2,
		"cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv",
		3,
		map[string]uint64{})

	s.Assert().Equal(uint64(1), accountState.GetRewardProgramId(), "reward program id must match")
	s.Assert().Equal(uint64(2), accountState.GetClaimPeriodId(), "reward claim period id must match")
	s.Assert().Equal(uint64(3), accountState.GetSharesEarned(), "earned shares must match")
	s.Assert().Equal("cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", accountState.GetAddress(), "address must match")
	s.Assert().Equal(types.RewardAccountState_CLAIM_STATUS_UNCLAIMABLE, accountState.GetClaimStatus(), "should be set to unclaimable initially")
	s.Assert().Equal(map[string]uint64{}, accountState.GetActionCounter(), "action counter must match")
}

func (s *KeeperTestSuite) TestGetSetRewardAccountState() {
	expectedState := types.NewRewardAccountState(
		1,
		2,
		"cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27",
		3,
		map[string]uint64{},
	)

	s.app.RewardKeeper.SetRewardAccountState(s.ctx, expectedState)
	actualState, err := s.app.RewardKeeper.GetRewardAccountState(s.ctx,
		expectedState.GetRewardProgramId(),
		expectedState.GetClaimPeriodId(),
		expectedState.GetAddress())

	s.Assert().Nil(err, "must not have error")
	s.Assert().Equal(expectedState.GetRewardProgramId(), actualState.GetRewardProgramId(), "reward program id must match")
	s.Assert().Equal(expectedState.GetClaimPeriodId(), actualState.GetClaimPeriodId(), "reward claim period id must match")
	s.Assert().Equal(expectedState.GetAddress(), actualState.GetAddress(), "address must match")
	s.Assert().Equal(expectedState.GetSharesEarned(), actualState.GetSharesEarned(), "shares earned must match")
	s.Assert().Equal(expectedState.GetClaimStatus(), actualState.GetClaimStatus(), "should be set to unclaimed initially")
	s.Assert().Equal(expectedState.GetActionCounter(), actualState.GetActionCounter(), "action counter must match")
}

func (s *KeeperTestSuite) TestGetInvalidAccountState() {
	actualState, err := s.app.RewardKeeper.GetRewardAccountState(s.ctx,
		99,
		99,
		"cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27")

	s.Assert().Nil(err, "must not have error")
	s.Assert().Error(actualState.Validate(), "account state validate basic must return error")
}

func (s *KeeperTestSuite) TestIterateAccountStates() {
	state1 := types.NewRewardAccountState(1, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state2)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state3)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state4)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state5)

	counter := 0
	s.app.RewardKeeper.IterateRewardAccountStates(s.ctx, 2, 2, func(state types.RewardAccountState) bool {
		counter += 1
		return false
	})

	s.Assert().Equal(2, counter, "should have correct number of iterations")
}

func (s *KeeperTestSuite) TestIterateAccountStatesByAddress() {
	state1 := types.NewRewardAccountState(1, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state2)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state3)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state4)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state5)

	counter := 0
	addr, err := sdk.AccAddressFromBech32("cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv")
	s.Assert().NoError(err, "no error should be thrown")
	s.app.RewardKeeper.IterateRewardAccountStatesByAddress(s.ctx, addr, func(state types.RewardAccountState) bool {
		counter += 1
		return false
	})

	s.Assert().Equal(4, counter, "should have correct number of iterations")
}

func (s *KeeperTestSuite) TestEmptyIterateAccountStates() {
	state1 := types.NewRewardAccountState(1, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state2)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state3)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state4)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state5)

	counter := 0
	s.app.RewardKeeper.IterateRewardAccountStates(s.ctx, 1, 4, func(state types.RewardAccountState) bool {
		counter += 1
		return false
	})

	s.Assert().Equal(0, counter, "should have correct number of iterations")
}

func (s *KeeperTestSuite) TestIterateAccountStatesHalt() {
	state1 := types.NewRewardAccountState(1, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state2)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state3)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state4)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state5)

	counter := 0
	s.app.RewardKeeper.IterateRewardAccountStates(s.ctx, 1, 2, func(state types.RewardAccountState) bool {
		counter += 1
		return counter == 1
	})

	s.Assert().Equal(1, counter, "should have correct number of iterations")
}

func (s *KeeperTestSuite) TestIterateAllAccountStates() {
	state1 := types.NewRewardAccountState(1, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state2)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state3)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state4)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state5)

	counter := 0
	s.app.RewardKeeper.IterateAllRewardAccountStates(s.ctx, func(state types.RewardAccountState) bool {
		counter += 1
		return false
	})

	s.Assert().Equal(5, counter, "should have correct number of iterations")
}

func (s *KeeperTestSuite) TestEmptyIterateAllAccountStates() {
	counter := 0
	s.app.RewardKeeper.IterateAllRewardAccountStates(s.ctx, func(state types.RewardAccountState) bool {
		counter += 1
		return false
	})

	s.Assert().Equal(0, counter, "should have correct number of iterations")
}

func (s *KeeperTestSuite) TestIterateAllAccountStatesHalt() {
	state1 := types.NewRewardAccountState(1, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state2)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state3)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state4)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state5)

	counter := 0
	s.app.RewardKeeper.IterateAllRewardAccountStates(s.ctx, func(state types.RewardAccountState) bool {
		counter += 1
		return counter == 1
	})

	s.Assert().Equal(1, counter, "should have correct number of iterations")
}

func (s *KeeperTestSuite) TestIterateRewardAccountStatesForRewardProgram() {
	state1 := types.NewRewardAccountState(1, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state2)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state3)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state4)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state5)

	counter := 0
	s.app.RewardKeeper.IterateRewardAccountStatesForRewardProgram(s.ctx, 2, func(state types.RewardAccountState) bool {
		counter += 1
		return false
	})

	s.Assert().Equal(3, counter, "should have correct number of iterations")
}

func (s *KeeperTestSuite) TestEmptyIterateRewardAccountStatesForRewardProgram() {
	counter := 0
	s.app.RewardKeeper.IterateRewardAccountStatesForRewardProgram(s.ctx, 1, func(state types.RewardAccountState) bool {
		counter += 1
		return false
	})

	s.Assert().Equal(0, counter, "should have correct number of iterations")
}

func (s *KeeperTestSuite) TestIterateRewardAccountStatesForRewardProgramHalt() {
	state1 := types.NewRewardAccountState(1, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state2)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state3)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state4)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state5)

	counter := 0
	s.app.RewardKeeper.IterateRewardAccountStatesForRewardProgram(s.ctx, 2, func(state types.RewardAccountState) bool {
		counter += 1
		return counter == 1
	})

	s.Assert().Equal(1, counter, "should have correct number of iterations")
}

func (s *KeeperTestSuite) TestGetRewardAccountStatesForClaimPeriod() {
	state1 := types.NewRewardAccountState(1, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state2)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state3)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state4)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state5)

	states, err := s.app.RewardKeeper.GetRewardAccountStatesForClaimPeriod(s.ctx, 2, 2)
	s.Assert().NoError(err, "no error should be thrown when there are account states.")
	s.Assert().Equal(2, len(states), "should have correct number of account states")
}

func (s *KeeperTestSuite) TestGetRewardAccountStatesForClaimPeriodHandlesEmpty() {
	states, err := s.app.RewardKeeper.GetRewardAccountStatesForClaimPeriod(s.ctx, 1, 1)
	s.Assert().NoError(err, "no error should be thrown when there are no account states.")
	s.Assert().Equal(0, len(states), "should have no account states")
}

func (s *KeeperTestSuite) TestGetRewardAccountStatesForRewardProgram() {
	state1 := types.NewRewardAccountState(1, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state2)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state3)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state4)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state5)

	states, err := s.app.RewardKeeper.GetRewardAccountStatesForRewardProgram(s.ctx, 2)
	s.Assert().NoError(err, "no error should be thrown when there are account states.")
	s.Assert().Equal(3, len(states), "should have correct number of account states")
}

func (s *KeeperTestSuite) TestGetRewardAccountStatesForRewardProgramHandlesEmpty() {
	states, err := s.app.RewardKeeper.GetRewardAccountStatesForClaimPeriod(s.ctx, 1, 1)
	s.Assert().NoError(err, "no error should be thrown when there are no account states.")
	s.Assert().Equal(0, len(states), "should have no account states")
}

func (s *KeeperTestSuite) TestMakeRewardClaimsClaimableForPeriod() {
	state1 := types.NewRewardAccountState(1, 2, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 3, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(2, 1, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(2, 2, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", 0, map[string]uint64{})
	state5 := types.NewRewardAccountState(2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})

	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state2)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state3)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state4)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state5)

	err := s.app.RewardKeeper.MakeRewardClaimsClaimableForPeriod(s.ctx, 2, 2)
	s.Assert().NoError(err, "no error should be thrown when there are account states.")

	state1, _ = s.app.RewardKeeper.GetRewardAccountState(s.ctx, 1, 2, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27")
	state2, _ = s.app.RewardKeeper.GetRewardAccountState(s.ctx, 1, 3, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27")
	state3, _ = s.app.RewardKeeper.GetRewardAccountState(s.ctx, 2, 1, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27")
	state4, _ = s.app.RewardKeeper.GetRewardAccountState(s.ctx, 2, 2, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27")
	state5, _ = s.app.RewardKeeper.GetRewardAccountState(s.ctx, 2, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")

	s.Assert().NotEqual(types.RewardAccountState_CLAIM_STATUS_CLAIMABLE, state1.GetClaimStatus(), "account state should not be updated to be claimable")
	s.Assert().NotEqual(types.RewardAccountState_CLAIM_STATUS_CLAIMABLE, state2.GetClaimStatus(), "account state should not be updated to be claimable")
	s.Assert().NotEqual(types.RewardAccountState_CLAIM_STATUS_CLAIMABLE, state3.GetClaimStatus(), "account state should not be updated to be claimable")
	s.Assert().Equal(types.RewardAccountState_CLAIM_STATUS_CLAIMABLE, state4.GetClaimStatus(), "account state should not be updated to be claimable")
	s.Assert().Equal(types.RewardAccountState_CLAIM_STATUS_CLAIMABLE, state5.GetClaimStatus(), "account state should not be updated to be claimable")
}

func (s *KeeperTestSuite) TestMakeRewardClaimsClaimableForPeriodHandlesEmpty() {
	err := s.app.RewardKeeper.MakeRewardClaimsClaimableForPeriod(s.ctx, 1, 1)
	s.Assert().NoError(err, "no error should be thrown when there are no account states.")
}

func (s *KeeperTestSuite) TestExpireRewardClaimsForRewardProgram() {
	state1 := types.NewRewardAccountState(1, 1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})
	state2 := types.NewRewardAccountState(1, 1, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", 0, map[string]uint64{})
	state3 := types.NewRewardAccountState(1, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", 0, map[string]uint64{})
	state4 := types.NewRewardAccountState(1, 2, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27", 0, map[string]uint64{})
	state4.ClaimStatus = types.RewardAccountState_CLAIM_STATUS_CLAIMED

	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state1)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state2)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state3)
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state4)

	err := s.app.RewardKeeper.ExpireRewardClaimsForRewardProgram(s.ctx, 1)
	s.Assert().NoError(err, "no error should be thrown when there are account states.")

	state1, _ = s.app.RewardKeeper.GetRewardAccountState(s.ctx, 1, 1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	state2, _ = s.app.RewardKeeper.GetRewardAccountState(s.ctx, 1, 1, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27")
	state3, _ = s.app.RewardKeeper.GetRewardAccountState(s.ctx, 1, 2, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	state4, _ = s.app.RewardKeeper.GetRewardAccountState(s.ctx, 1, 2, "cosmos1depk54cuajgkzea6zpgkq36tnjwdzv4afc3d27")

	s.Assert().Equal(types.RewardAccountState_CLAIM_STATUS_EXPIRED, state1.GetClaimStatus(), "account state should be updated to expired")
	s.Assert().Equal(types.RewardAccountState_CLAIM_STATUS_EXPIRED, state2.GetClaimStatus(), "account state should be updated to expired")
	s.Assert().Equal(types.RewardAccountState_CLAIM_STATUS_EXPIRED, state3.GetClaimStatus(), "account state should be updated to expired")
	s.Assert().Equal(types.RewardAccountState_CLAIM_STATUS_CLAIMED, state4.GetClaimStatus(), "account state should not be updated to expired if claimed")
}

func (s *KeeperTestSuite) TestExpireRewardClaimsForRewardProgramHandlesEmpty() {
	err := s.app.RewardKeeper.ExpireRewardClaimsForRewardProgram(s.ctx, 1)
	s.Assert().NoError(err, "no error should be thrown when there are no account states.")
}

func (s *KeeperTestSuite) TestParseRewardAccountLookUpKey() {
	addressFromSec256k1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	rewardProgramId := uint64(123456)
	claimPeriodId := uint64(7891011)
	accountStateAddressLookupKey := types.GetRewardAccountStateAddressLookupKey(addressFromSec256k1, rewardProgramId, claimPeriodId)
	lookup, err := types.ParseRewardAccountLookUpKey(accountStateAddressLookupKey, addressFromSec256k1)
	s.Assert().NoError(err, "no error expected for parsing GetRewardAccountStateAddressLookupKey.")
	s.Assert().Equal(addressFromSec256k1, lookup.Addr)
	s.Assert().Equal(rewardProgramId, lookup.RewardID)
	s.Assert().Equal(claimPeriodId, lookup.ClaimID)
}
