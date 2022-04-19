package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"

	epoch "github.com/provenance-io/provenance/x/epoch"
	"github.com/provenance-io/provenance/x/reward"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/reward/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var (
	PKs = simapp.CreateTestPubKeys(500)
)

type KeeperTestSuite struct {
	suite.Suite

	app         *simapp.App
	ctx         sdk.Context
	queryClient types.QueryClient
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.RewardKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

//TODO because of this line k.AfterEpochEnd(ctx, epochInfo.Identifier, epochInfo.CurrentEpoch-1)  double check calcs on all these tests.
func (s *KeeperTestSuite) TestInitGenesisAddingAttributes() {
	action := types.NewActionDelegate()
	coin := sdk.NewInt64Coin("jackthecat", 10000)
	maxCoin := sdk.NewInt64Coin("jackthecat", 100)
	var rewardData types.GenesisState
	rewardData.RewardPrograms = []types.RewardProgram{
		types.NewRewardProgram(1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", coin, maxCoin, "day", 1, 10, types.NewEligibilityCriteria("criteria", &action), false, 1, 10),
	}
	sharesPerEpoch := types.SharesPerEpochPerRewardsProgram{RewardProgramId: 1, TotalShares: 2, LatestRecordedEpoch: 1000, Claimed: false, Expired: false, TotalRewardClaimed: coin}
	rewardData.RewardClaims = []types.RewardClaim{types.NewRewardClaim("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", []types.SharesPerEpochPerRewardsProgram{sharesPerEpoch})}
	rewardData.EligibilityCriterias = []types.EligibilityCriteria{types.NewEligibilityCriteria("delegate", &action)}
	rewardData.EpochRewardDistributions = []types.EpochRewardDistribution{types.NewEpochRewardDistribution("day", 1,
		coin,
		10,
		false,
	)}
	rewardData.ActionDelegate = types.NewActionDelegate()
	rewardData.ActionTransferDelegations = types.NewActionTransferDelegations()
	s.Assert().NotPanics(func() { s.app.RewardKeeper.InitGenesis(s.ctx, &rewardData) })
	s.Assert().NotPanics(func() { s.app.RewardKeeper.ExportGenesis(s.ctx) })

	rewardData.RewardPrograms = []types.RewardProgram{
		types.NewRewardProgram(1, "", sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 10), "day", 1, 10, types.NewEligibilityCriteria("criteria", &action), false, 1, 10),
	}

	s.Assert().Panics(func() { s.app.RewardKeeper.InitGenesis(s.ctx, &rewardData) })
}

// create a test that creates a reward program for an epoch
// go past 10 epochs, check that the reward program has expired
func (s *KeeperTestSuite) TestCheckRewardProgramExpired() {
	action := types.NewActionDelegate()
	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)
	rewardProgram := types.NewRewardProgram(1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", coin, maxCoin, "day", 1, 10, types.NewEligibilityCriteria("criteria", &action), false, 1, 10)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)
	rewardProgramGet, err := s.app.RewardKeeper.GetRewardProgram(s.ctx, 1)
	s.Assert().NoError(err)

	s.Assert().Equal(rewardProgram.Coin, rewardProgramGet.Coin)
	s.Assert().Equal(rewardProgram.Coin, rewardProgramGet.Coin)
	s.Assert().Equal(rewardProgram.StartEpoch, rewardProgramGet.StartEpoch)
	s.Assert().Equal(rewardProgram.Id, rewardProgramGet.Id)
	s.Assert().Equal(rewardProgram.EligibilityCriteria.Action.TypeUrl, rewardProgramGet.EligibilityCriteria.Action.TypeUrl)
	s.Assert().Equal(rewardProgram.Expired, false)

	// if we increase the block height
	for i := 0; i < 11; i++ {
		s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + ((24 * 60 * 60 * 30) / 5) + 1)
		epoch.BeginBlocker(s.ctx, s.app.EpochKeeper)
	}
	reward.EndBlocker(s.ctx, s.app.RewardKeeper)
	rewardProgramGet, err = s.app.RewardKeeper.GetRewardProgram(s.ctx, 1)
	s.Assert().Equal(rewardProgram.Coin, rewardProgramGet.Coin)
	s.Assert().Equal(rewardProgram.Coin, rewardProgramGet.Coin)
	s.Assert().Equal(rewardProgram.StartEpoch, rewardProgramGet.StartEpoch)
	s.Assert().Equal(rewardProgram.Id, rewardProgramGet.Id)
	s.Assert().Equal(rewardProgram.EligibilityCriteria.Action.TypeUrl, rewardProgramGet.EligibilityCriteria.Action.TypeUrl)
	s.Assert().Equal(rewardProgramGet.Expired, true)
}

// create a test that creates a reward program for an epoch
// make a transfer, with delegation
func (s *KeeperTestSuite) TestCreateRewardClaim() {
	action := types.NewActionDelegate()
	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)
	rewardProgram := types.NewRewardProgram(1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", coin, maxCoin, "day", 1, 10, types.NewEligibilityCriteria("criteria", &action), false, 0, 10)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)
	rewardProgramGet, err := s.app.RewardKeeper.GetRewardProgram(s.ctx, 1)
	s.Assert().NoError(err)

	s.Assert().Equal(rewardProgram.Coin, rewardProgramGet.Coin)
	s.Assert().Equal(rewardProgram.StartEpoch, rewardProgramGet.StartEpoch)
	s.Assert().Equal(rewardProgram.Id, rewardProgramGet.Id)
	s.Assert().Equal(rewardProgram.EligibilityCriteria.Action.TypeUrl, rewardProgramGet.EligibilityCriteria.Action.TypeUrl)
	s.Assert().Equal(false, rewardProgram.Expired)

	// go past 4 epochs, no events, nothing is incremented
	for i := 0; i < 4; i++ {
		s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + ((24 * 60 * 60 * 30) / 5) + 1)
		epoch.BeginBlocker(s.ctx, s.app.EpochKeeper)
		reward.EndBlocker(s.ctx, s.app.RewardKeeper)

	}

	// get reward epoch distribution, should be 0 total shares
	epochRewardDistribution, err := s.app.RewardKeeper.GetEpochRewardDistribution(s.ctx, "day", 1)
	s.Assert().Nil(err)
	s.Assert().NotNil(epochRewardDistribution)
	s.Assert().Equal(epochRewardDistribution.RewardProgramId, uint64(1))
	s.Assert().Equal("day", epochRewardDistribution.EpochId)
	s.Assert().Equal(int64(0), epochRewardDistribution.TotalShares)
	s.Assert().Equal(coin, epochRewardDistribution.TotalRewardsPool)
	s.Assert().Equal(false, epochRewardDistribution.EpochEnded)

	addrDels := simapp.AddTestAddrsIncremental(s.app, s.ctx, 3, sdk.NewInt(10000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrDels)
	// add a delegation
	//construct the validators
	amts := []sdk.Int{sdk.NewInt(9), sdk.NewInt(8), sdk.NewInt(7)}
	var validators [3]stakingtypes.Validator
	for i, amt := range amts {
		validators[i] = teststaking.NewValidator(s.T(), valAddrs[i], PKs[i])
		validators[i], _ = validators[i].AddTokensFromDel(amt)
	}

	delegation := sdk.NewInt64Coin("hotdog", 10)

	events := []abci.Event{
		{Type: stakingtypes.EventTypeDelegate, Attributes: []abci.EventAttribute{{Key: []byte(stakingtypes.AttributeKeyValidator), Value: []byte(validators[0].OperatorAddress), Index: true}}},
		{Type: stakingtypes.EventTypeDelegate, Attributes: []abci.EventAttribute{{Key: []byte(sdk.AttributeKeyAmount), Value: []byte(delegation.Amount.String()), Index: true}}},
		{Type: stakingtypes.EventTypeDelegate, Attributes: []abci.EventAttribute{{Key: []byte(stakingtypes.AttributeKeyNewShares), Value: []byte(sdk.NewDec(10).String()), Index: true}}},
		{Type: sdk.EventTypeMessage, Attributes: []abci.EventAttribute{{Key: []byte(sdk.AttributeKeyModule), Value: []byte(stakingtypes.AttributeValueCategory), Index: true},
			{Key: []byte(sdk.AttributeKeySender), Value: []byte(addrDels[0].String()), Index: true}}},
	}

	s.ctx = s.ctx.WithEventManager(sdk.NewEventManagerWithHistory(events))

	// increment a day, feed events in the rewards end blocker // 5 epochs
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + ((24 * 60 * 60 * 30) / 5) + 1)
	epoch.BeginBlocker(s.ctx, s.app.EpochKeeper)
	reward.EndBlocker(s.ctx, s.app.RewardKeeper)

	// get reward epoch distribution
	epochRewardDistribution, err = s.app.RewardKeeper.GetEpochRewardDistribution(s.ctx, "day", 1)
	s.Assert().Nil(err)
	s.Assert().NotNil(epochRewardDistribution)
	s.Assert().Equal(uint64(1), epochRewardDistribution.RewardProgramId, "reward program id is wrong.")
	s.Assert().Equal("day", epochRewardDistribution.EpochId, "epoch id is wrong.")
	s.Assert().Equal(int64(1), epochRewardDistribution.TotalShares, "total epoch distribution shares wrong.")
	s.Assert().Equal(coin, epochRewardDistribution.TotalRewardsPool, "total rewards pool incorrect.")
	s.Assert().Equal(false, epochRewardDistribution.EpochEnded, "epoch should not have ended.")

	// goto the end of epoch + 1 (11 days)
	// ctx history always has the delegate event.
	// call end blocker of rewards to increment rewards accumulated
	// call begin blocker on epoch to signal end of blocker
	// ^^ should not double count rewards
	for i := 0; i < 5; i++ {
		s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + ((24 * 60 * 60 * 30) / 5) + 1)
		epoch.BeginBlocker(s.ctx, s.app.EpochKeeper)
		reward.EndBlocker(s.ctx, s.app.RewardKeeper)
	}

	// get reward epoch distribution
	epochRewardDistribution, err = s.app.RewardKeeper.GetEpochRewardDistribution(s.ctx, "day", 1)
	s.Assert().Nil(err)
	s.Assert().NotNil(epochRewardDistribution)
	s.Assert().Equal(uint64(1), epochRewardDistribution.RewardProgramId)
	s.Assert().Equal("day", epochRewardDistribution.EpochId)
	s.Assert().Equal(int64(6), epochRewardDistribution.TotalShares)
	s.Assert().Equal(coin, epochRewardDistribution.TotalRewardsPool)
	s.Assert().Equal(false, epochRewardDistribution.EpochEnded)

	// last block of the epoch, should end but also increment
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + ((24 * 60 * 60 * 30) / 5) + 1)
	epoch.BeginBlocker(s.ctx, s.app.EpochKeeper)
	reward.EndBlocker(s.ctx, s.app.RewardKeeper)
	epochRewardDistribution, err = s.app.RewardKeeper.GetEpochRewardDistribution(s.ctx, "day", 1)
	s.Assert().Nil(err)
	s.Assert().NotNil(epochRewardDistribution)
	s.Assert().Equal(uint64(1), epochRewardDistribution.RewardProgramId)
	s.Assert().Equal("day", epochRewardDistribution.EpochId)
	s.Assert().Equal(int64(7), epochRewardDistribution.TotalShares)
	s.Assert().Equal(coin, epochRewardDistribution.TotalRewardsPool)
	s.Assert().Equal(true, epochRewardDistribution.EpochEnded)

	// epoch has ended, should not increment
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + ((24 * 60 * 60 * 30) / 5) + 1)
	epoch.BeginBlocker(s.ctx, s.app.EpochKeeper)
	reward.EndBlocker(s.ctx, s.app.RewardKeeper)
	epochRewardDistribution, err = s.app.RewardKeeper.GetEpochRewardDistribution(s.ctx, "day", 1)
	s.Assert().Nil(err)
	s.Assert().NotNil(epochRewardDistribution)
	s.Assert().Equal(uint64(1), epochRewardDistribution.RewardProgramId)
	s.Assert().Equal("day", epochRewardDistribution.EpochId)
	s.Assert().Equal(int64(7), epochRewardDistribution.TotalShares)
	s.Assert().Equal(coin, epochRewardDistribution.TotalRewardsPool)
	s.Assert().Equal(true, epochRewardDistribution.EpochEnded)

}

// create a test that creates a reward program for an epoch
// make a transfer, with delegation
// 10 epoch 10 reward shares
func (s *KeeperTestSuite) TestCreateRewardClaim_1() {
	action := types.NewActionDelegate()
	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)
	rewardProgram := types.NewRewardProgram(1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", coin, maxCoin, "day", 1, 10, types.NewEligibilityCriteria("criteria", &action), false, 0, 10)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)
	rewardProgramGet, err := s.app.RewardKeeper.GetRewardProgram(s.ctx, 1)
	s.Assert().NoError(err)

	s.Assert().Equal(rewardProgram.Coin, rewardProgramGet.Coin)
	s.Assert().Equal(rewardProgram.StartEpoch, rewardProgramGet.StartEpoch)
	s.Assert().Equal(rewardProgram.Id, rewardProgramGet.Id)
	s.Assert().Equal(rewardProgram.EligibilityCriteria.Action.TypeUrl, rewardProgramGet.EligibilityCriteria.Action.TypeUrl)
	s.Assert().Equal(rewardProgram.Expired, false)

	addrDels := simapp.AddTestAddrsIncremental(s.app, s.ctx, 3, sdk.NewInt(10000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrDels)
	// add a delegation
	//construct the validators
	amts := []sdk.Int{sdk.NewInt(9), sdk.NewInt(8), sdk.NewInt(7)}
	var validators [3]stakingtypes.Validator
	for i, amt := range amts {
		validators[i] = teststaking.NewValidator(s.T(), valAddrs[i], PKs[i])
		validators[i], _ = validators[i].AddTokensFromDel(amt)
	}

	delegation := sdk.NewInt64Coin("hotdog", 10)

	events := []abci.Event{
		{Type: stakingtypes.EventTypeDelegate, Attributes: []abci.EventAttribute{{Key: []byte(stakingtypes.AttributeKeyValidator), Value: []byte(validators[0].OperatorAddress), Index: true}}},
		{Type: stakingtypes.EventTypeDelegate, Attributes: []abci.EventAttribute{{Key: []byte(sdk.AttributeKeyAmount), Value: []byte(delegation.Amount.String()), Index: true}}},
		{Type: stakingtypes.EventTypeDelegate, Attributes: []abci.EventAttribute{{Key: []byte(stakingtypes.AttributeKeyNewShares), Value: []byte(sdk.NewDec(10).String()), Index: true}}},
		{Type: sdk.EventTypeMessage, Attributes: []abci.EventAttribute{{Key: []byte(sdk.AttributeKeyModule), Value: []byte(stakingtypes.AttributeValueCategory), Index: true},
			{Key: []byte(sdk.AttributeKeySender), Value: []byte(addrDels[0].String()), Index: true}}},
	}

	s.ctx = s.ctx.WithEventManager(sdk.NewEventManagerWithHistory(events))

	// go past 11 epochs
	for i := 0; i < 11; i++ {
		s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + ((24 * 60 * 60 * 30) / 5) + 1)
		epoch.BeginBlocker(s.ctx, s.app.EpochKeeper)
		reward.EndBlocker(s.ctx, s.app.RewardKeeper)
	}

	// get reward epoch distribution
	epochRewardDistribution, err := s.app.RewardKeeper.GetEpochRewardDistribution(s.ctx, "day", 1)
	s.Assert().Nil(err)
	s.Assert().NotNil(epochRewardDistribution)
	s.Assert().Equal(epochRewardDistribution.RewardProgramId, uint64(1))
	s.Assert().Equal(epochRewardDistribution.EpochId, "day")
	s.Assert().Equal(epochRewardDistribution.TotalShares, int64(10))
	s.Assert().Equal(epochRewardDistribution.TotalRewardsPool, coin)
	s.Assert().Equal(true, epochRewardDistribution.EpochEnded)

	// epoch has ended, should not increment
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + ((24 * 60 * 60 * 30) / 5) + 1)
	epoch.BeginBlocker(s.ctx, s.app.EpochKeeper)
	reward.EndBlocker(s.ctx, s.app.RewardKeeper)
	epochRewardDistribution, err = s.app.RewardKeeper.GetEpochRewardDistribution(s.ctx, "day", 1)

	s.Assert().Nil(err)
	s.Assert().NotNil(epochRewardDistribution)
	s.Assert().Equal(uint64(1), epochRewardDistribution.RewardProgramId, "reward program id incorrect.")
	s.Assert().Equal("day", epochRewardDistribution.EpochId, "epoch id incorrect")
	s.Assert().Equal(int64(10), epochRewardDistribution.TotalShares, "Total shares should stay the same")
	s.Assert().Equal(coin, epochRewardDistribution.TotalRewardsPool, "Reward pool totals are wrong")
	s.Assert().Equal(true, epochRewardDistribution.EpochEnded, "Epoch should remain ended")

	// now let's check rewards claims
	rewardClaim, err := s.app.RewardKeeper.GetRewardClaim(s.ctx, addrDels[0].String())
	s.Assert().Nil(err)
	s.Assert().NotNil(rewardClaim)

	s.Assert().Equal(addrDels[0].String(), rewardClaim.Address, "address should match event delegator address")
	s.Assert().Equal(int64(10), rewardClaim.SharesPerEpochPerReward[0].TotalShares, "total shares wrong")
	s.Assert().Equal(int64(10), rewardClaim.SharesPerEpochPerReward[0].EphemeralActionCount, "ephemeral shares wrong")

}

func (s *KeeperTestSuite) TestCreateRewardClaimTestMin() {
	action := types.NewActionDelegate()
	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)
	rewardProgram := types.NewRewardProgram(1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", coin, maxCoin, "day", 1, 10, types.NewEligibilityCriteria("criteria", &action), false, 1, 10)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)
	rewardProgramGet, err := s.app.RewardKeeper.GetRewardProgram(s.ctx, 1)
	s.Assert().NoError(err)

	s.Assert().Equal(rewardProgram.Coin, rewardProgramGet.Coin)
	s.Assert().Equal(rewardProgram.StartEpoch, rewardProgramGet.StartEpoch)
	s.Assert().Equal(rewardProgram.Id, rewardProgramGet.Id)
	s.Assert().Equal(rewardProgram.EligibilityCriteria.Action.TypeUrl, rewardProgramGet.EligibilityCriteria.Action.TypeUrl)
	s.Assert().Equal(rewardProgram.Expired, false)

	addrDels := simapp.AddTestAddrsIncremental(s.app, s.ctx, 3, sdk.NewInt(10000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrDels)
	// add a delegation
	//construct the validators
	amts := []sdk.Int{sdk.NewInt(9), sdk.NewInt(8), sdk.NewInt(7)}
	var validators [3]stakingtypes.Validator
	for i, amt := range amts {
		validators[i] = teststaking.NewValidator(s.T(), valAddrs[i], PKs[i])
		validators[i], _ = validators[i].AddTokensFromDel(amt)
	}

	delegation := sdk.NewInt64Coin("hotdog", 10)

	events := []abci.Event{
		{Type: stakingtypes.EventTypeDelegate, Attributes: []abci.EventAttribute{{Key: []byte(stakingtypes.AttributeKeyValidator), Value: []byte(validators[0].OperatorAddress), Index: true}}},
		{Type: stakingtypes.EventTypeDelegate, Attributes: []abci.EventAttribute{{Key: []byte(sdk.AttributeKeyAmount), Value: []byte(delegation.Amount.String()), Index: true}}},
		{Type: stakingtypes.EventTypeDelegate, Attributes: []abci.EventAttribute{{Key: []byte(stakingtypes.AttributeKeyNewShares), Value: []byte(sdk.NewDec(10).String()), Index: true}}},
		{Type: sdk.EventTypeMessage, Attributes: []abci.EventAttribute{{Key: []byte(sdk.AttributeKeyModule), Value: []byte(stakingtypes.AttributeValueCategory), Index: true},
			{Key: []byte(sdk.AttributeKeySender), Value: []byte(addrDels[0].String()), Index: true}}},
	}

	s.ctx = s.ctx.WithEventManager(sdk.NewEventManagerWithHistory(events))

	// go past 11 epochs
	for i := 0; i < 11; i++ {
		s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + ((24 * 60 * 60 * 30) / 5) + 1)
		epoch.BeginBlocker(s.ctx, s.app.EpochKeeper)
		reward.EndBlocker(s.ctx, s.app.RewardKeeper)
	}

	// get reward epoch distribution
	epochRewardDistribution, err := s.app.RewardKeeper.GetEpochRewardDistribution(s.ctx, "day", 1)
	s.Assert().Nil(err)
	s.Assert().NotNil(epochRewardDistribution)
	s.Assert().Equal(uint64(1), epochRewardDistribution.RewardProgramId, "reward program id incorrect.")
	s.Assert().Equal("day", epochRewardDistribution.EpochId, "epoch id incorrect")
	s.Assert().Equal(epochRewardDistribution.TotalShares, int64(9), "total shares wrong")
	s.Assert().Equal(coin, epochRewardDistribution.TotalRewardsPool, "total reward pool wrong.")
	s.Assert().Equal(true, epochRewardDistribution.EpochEnded)

	// now let's check rewards claims
	rewardClaim, err := s.app.RewardKeeper.GetRewardClaim(s.ctx, addrDels[0].String())
	s.Assert().Nil(err)
	s.Assert().NotNil(rewardClaim)

	s.Assert().Equal(addrDels[0].String(), rewardClaim.Address, "address should match event delegator address")
	s.Assert().Equal(int64(9), rewardClaim.SharesPerEpochPerReward[0].TotalShares, "total shares wrong")
	s.Assert().Equal(int64(10), rewardClaim.SharesPerEpochPerReward[0].EphemeralActionCount, "ephemeral shares wrong")

}

func (s *KeeperTestSuite) TestCreateRewardClaimTestMax() {
	action := types.NewActionDelegate()
	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)
	rewardProgram := types.NewRewardProgram(1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", coin, maxCoin, "day", 1, 10, types.NewEligibilityCriteria("criteria", &action), false, 0, 5)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)
	rewardProgramGet, err := s.app.RewardKeeper.GetRewardProgram(s.ctx, 1)
	s.Assert().NoError(err)

	s.Assert().Equal(rewardProgram.Coin, rewardProgramGet.Coin)
	s.Assert().Equal(rewardProgram.StartEpoch, rewardProgramGet.StartEpoch)
	s.Assert().Equal(rewardProgram.Id, rewardProgramGet.Id)
	s.Assert().Equal(rewardProgram.EligibilityCriteria.Action.TypeUrl, rewardProgramGet.EligibilityCriteria.Action.TypeUrl)
	s.Assert().Equal(rewardProgram.Expired, false)

	addrDels := simapp.AddTestAddrsIncremental(s.app, s.ctx, 3, sdk.NewInt(10000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrDels)
	// add a delegation
	//construct the validators
	amts := []sdk.Int{sdk.NewInt(9), sdk.NewInt(8), sdk.NewInt(7)}
	var validators [3]stakingtypes.Validator
	for i, amt := range amts {
		validators[i] = teststaking.NewValidator(s.T(), valAddrs[i], PKs[i])
		validators[i], _ = validators[i].AddTokensFromDel(amt)
	}

	delegation := sdk.NewInt64Coin("hotdog", 10)

	events := []abci.Event{
		{Type: stakingtypes.EventTypeDelegate, Attributes: []abci.EventAttribute{{Key: []byte(stakingtypes.AttributeKeyValidator), Value: []byte(validators[0].OperatorAddress), Index: true}}},
		{Type: stakingtypes.EventTypeDelegate, Attributes: []abci.EventAttribute{{Key: []byte(sdk.AttributeKeyAmount), Value: []byte(delegation.Amount.String()), Index: true}}},
		{Type: stakingtypes.EventTypeDelegate, Attributes: []abci.EventAttribute{{Key: []byte(stakingtypes.AttributeKeyNewShares), Value: []byte(sdk.NewDec(10).String()), Index: true}}},
		{Type: sdk.EventTypeMessage, Attributes: []abci.EventAttribute{{Key: []byte(sdk.AttributeKeyModule), Value: []byte(stakingtypes.AttributeValueCategory), Index: true},
			{Key: []byte(sdk.AttributeKeySender), Value: []byte(addrDels[0].String()), Index: true}}},
	}

	s.ctx = s.ctx.WithEventManager(sdk.NewEventManagerWithHistory(events))

	// go past 11 epochs
	for i := 0; i < 11; i++ {
		s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + ((24 * 60 * 60 * 30) / 5) + 1)
		epoch.BeginBlocker(s.ctx, s.app.EpochKeeper)
		reward.EndBlocker(s.ctx, s.app.RewardKeeper)
	}

	// get reward epoch distribution
	epochRewardDistribution, err := s.app.RewardKeeper.GetEpochRewardDistribution(s.ctx, "day", 1)
	s.Assert().Nil(err)
	s.Assert().NotNil(epochRewardDistribution)
	s.Assert().Equal(epochRewardDistribution.RewardProgramId, uint64(1))
	s.Assert().Equal(epochRewardDistribution.EpochId, "day")
	s.Assert().Equal(epochRewardDistribution.TotalShares, int64(5))
	s.Assert().Equal(epochRewardDistribution.TotalRewardsPool, coin)
	s.Assert().Equal(true, epochRewardDistribution.EpochEnded)

	// now let's check rewards claims
	rewardClaim, err := s.app.RewardKeeper.GetRewardClaim(s.ctx, addrDels[0].String())
	s.Assert().Nil(err)
	s.Assert().NotNil(rewardClaim)

	s.Assert().Equal(addrDels[0].String(), rewardClaim.Address, "address should match event delegator address")
	s.Assert().Equal(int64(5), rewardClaim.SharesPerEpochPerReward[0].TotalShares, "total shares wrong")
	s.Assert().Equal(int64(10), rewardClaim.SharesPerEpochPerReward[0].EphemeralActionCount, "ephemeral shares wrong")

}
