package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (suite *KeeperTestSuite) TestClaimRewards() {
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
		3,
		uint64(time.Day()),
		[]types.QualifyingAction{},
	)
	rewardProgram.State = types.RewardProgram_FINISHED
	rewardProgram.CurrentClaimPeriod = rewardProgram.GetClaimPeriods()
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, rewardProgram)

	for i := 1; i <= int(rewardProgram.GetClaimPeriods()); i++ {
		state := types.NewRewardAccountState(rewardProgram.GetId(), uint64(i), "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv", 1)
		state.ClaimStatus = types.RewardAccountState_CLAIMABLE
		suite.app.RewardKeeper.SetRewardAccountState(suite.ctx, state)
		distribution := types.NewClaimPeriodRewardDistribution(uint64(i), rewardProgram.GetId(), sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 1, true)
		suite.app.RewardKeeper.SetClaimPeriodRewardDistribution(suite.ctx, distribution)
	}

	details, reward, err := suite.app.RewardKeeper.ClaimRewards(suite.ctx, rewardProgram.GetId(), "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv")
	suite.Assert().NoError(err, "should throw no error")

	rewardProgram, err = suite.app.RewardKeeper.GetRewardProgram(suite.ctx, rewardProgram.GetId())
	suite.Assert().NoError(err, "should throw no error")
	suite.Assert().Equal(3, len(details), "should have rewards from every period")
	suite.Assert().Equal(sdk.NewInt64Coin("nhash", 300), reward, "should total up the rewards from the periods")

}

func (suite *KeeperTestSuite) TestClaimRewardsHandlesInvalidProgram() {
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
		uint64(time.Day()),
		[]types.QualifyingAction{},
	)
	rewardProgram.State = types.RewardProgram_FINISHED
	rewardProgram.CurrentClaimPeriod = 5

	details, reward, err := suite.app.RewardKeeper.ClaimRewards(suite.ctx, rewardProgram.GetId(), "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv")
	suite.Assert().Nil(details, "should have no reward details")
	suite.Assert().Equal(sdk.Coin{}, reward, "should have no reward")
	suite.Assert().Error(err, "should throw error")
}

func (suite *KeeperTestSuite) TestClaimRewardsHandlesExpiredProgram() {
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
		uint64(time.Day()),
		[]types.QualifyingAction{},
	)
	rewardProgram.State = types.RewardProgram_EXPIRED
	rewardProgram.CurrentClaimPeriod = 5
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, rewardProgram)

	details, reward, err := suite.app.RewardKeeper.ClaimRewards(suite.ctx, rewardProgram.GetId(), "cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv")
	suite.Assert().Nil(details, "should have no reward details")
	suite.Assert().Equal(sdk.Coin{}, reward, "should have no reward")
	suite.Assert().Error(err, "should throw error")
}

func (suite *KeeperTestSuite) TestRefundRewardClaims() {
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
		uint64(time.Day()),
		[]types.QualifyingAction{},
	)
	rewardProgram.RemainingPoolBalance = sdk.NewInt64Coin("nhash", 0)
	rewardProgram.ClaimedAmount = sdk.NewInt64Coin("nhash", 0)

	addr, _ := sdk.AccAddressFromBech32("cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv")
	beforeBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, "nhash")
	err := suite.app.RewardKeeper.RefundRewardClaims(suite.ctx, rewardProgram)
	afterBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, "nhash")

	suite.Assert().NoError(err, "no error should be thrown")
	suite.Assert().Equal(beforeBalance.Add(rewardProgram.TotalRewardPool), afterBalance, "unclaimed balance should be refunded")
}

func (suite *KeeperTestSuite) TestRefundRewardClaimsEmpty() {
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
		uint64(time.Day()),
		[]types.QualifyingAction{},
	)
	rewardProgram.RemainingPoolBalance = rewardProgram.GetTotalRewardPool()
	rewardProgram.ClaimedAmount = sdk.NewInt64Coin("nhash", 0)

	addr, _ := sdk.AccAddressFromBech32("cosmos1ffnqn02ft2psvyv4dyr56nnv6plllf9pm2kpmv")
	beforeBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, "nhash")
	err := suite.app.RewardKeeper.RefundRewardClaims(suite.ctx, rewardProgram)
	afterBalance := suite.app.BankKeeper.GetBalance(suite.ctx, addr, "nhash")

	suite.Assert().NoError(err, "no error should be thrown")
	suite.Assert().Equal(beforeBalance, afterBalance, "balance should stay same since all claims are taken")
}
