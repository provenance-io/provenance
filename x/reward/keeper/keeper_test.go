package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/reward"
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
	suite.createTestValidators(10)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.RewardKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

//TODO because of this line k.AfterEpochEnd(ctx, epochInfo.Identifier, epochInfo.CurrentEpoch-1)  double check calcs on all these tests.

/*func (s *KeeperTestSuite) TestInitGenesisAddingAttributes() {
	action := types.NewActionDelegate()
	coin := sdk.NewInt64Coin("jackthecat", 10000)
	maxCoin := sdk.NewInt64Coin("jackthecat", 100)
	var rewardData types.GenesisState
	now := time.Now().UTC()
	nextEpochTime := now.Add(time.Hour + 24)
	rewardData.RewardPrograms = []types.RewardProgram{
		types.NewRewardProgram(1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", coin, maxCoin, now, nextEpochTime, 60*24, 10, types.NewEligibilityCriteria("criteria", &action), false, false),
	}
	sharesPerEpoch := types.SharesPerEpochPerRewardsProgram{RewardProgramId: 1, TotalShares: 2, LatestRecordedEpoch: 1000, Claimed: false, Expired: false, TotalRewardClaimed: coin}
	rewardData.RewardClaims = []types.RewardClaim{types.NewRewardClaim("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", []types.SharesPerEpochPerRewardsProgram{sharesPerEpoch}, false)}
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
		types.NewRewardProgram(1, "", sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 10), time.Now().UTC(), time.Now().Add(time.Hour+24), 60*24, 10, types.NewEligibilityCriteria("criteria", &action), false, false),
	}

	s.Assert().Panics(func() { s.app.RewardKeeper.InitGenesis(s.ctx, &rewardData) })
}

// create a test that creates a reward program for an epoch
// go past 10 epochs, check that the reward program has expired
func (s *KeeperTestSuite) TestCheckRewardProgramExpired() {
	action := types.NewActionDelegate()
	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)
	now := time.Now().UTC()
	nextEpochTime := now.Add(time.Hour + 24)
	rewardProgram := types.NewRewardProgram(1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", coin, maxCoin, now, nextEpochTime, 60*24, 10, types.NewEligibilityCriteria("criteria", &action), false, false)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)
	rewardProgramGet, err := s.app.RewardKeeper.GetRewardProgram(s.ctx, 1)
	s.Assert().NoError(err)

	s.Assert().Equal(rewardProgram.Coin, rewardProgramGet.Coin)
	s.Assert().Equal(rewardProgram.Coin, rewardProgramGet.Coin)
	s.Assert().Equal(rewardProgram.ProgramStartTime, rewardProgramGet.ProgramStartTime)
	s.Assert().Equal(rewardProgram.Id, rewardProgramGet.Id)
	s.Assert().Equal(rewardProgram.EligibilityCriteria.Action.TypeUrl, rewardProgramGet.EligibilityCriteria.Action.TypeUrl)
	s.Assert().Equal(rewardProgram.Expired, false)

	// if we increase the block height
	for i := 0; i < 12; i++ {
		s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + ((24 * 60 * 60 * 30) / 5) + 1)
		epoch.BeginBlocker(s.ctx, s.app.EpochKeeper)
		reward.EndBlocker(s.ctx, s.app.RewardKeeper)
	}
	rewardProgramGet, err = s.app.RewardKeeper.GetRewardProgram(s.ctx, 1)
	s.Assert().Equal(rewardProgram.Coin, rewardProgramGet.Coin)
	s.Assert().Equal(rewardProgram.Coin, rewardProgramGet.Coin)
	s.Assert().Equal(rewardProgram.ProgramStartTime, rewardProgramGet.ProgramStartTime)
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
	now := time.Now().UTC()
	nextEpochTime := now.Add(time.Hour + 24)
	rewardProgram := types.NewRewardProgram(1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", coin, maxCoin, now, nextEpochTime, 60*24, 10, types.NewEligibilityCriteria("criteria", &action), false, false)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)
	rewardProgramGet, err := s.app.RewardKeeper.GetRewardProgram(s.ctx, 1)
	s.Assert().NoError(err)

	s.Assert().Equal(rewardProgram.Coin, rewardProgramGet.Coin)
	s.Assert().Equal(rewardProgram.MaxRewardByAddress, rewardProgramGet.MaxRewardByAddress)
	s.Assert().Equal(rewardProgram.ProgramStartTime, rewardProgramGet.ProgramStartTime)
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
	s.Assert().Equal(60*24, epochRewardDistribution.EpochId)
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
	epochRewardDistribution, err = s.app.RewardKeeper.GetEpochRewardDistribution(s.ctx, 60*24, 1)
	s.Assert().Nil(err)
	s.Assert().NotNil(epochRewardDistribution)
	s.Assert().Equal(uint64(1), epochRewardDistribution.RewardProgramId, "reward program id is wrong.")
	s.Assert().Equal(60*24, epochRewardDistribution.EpochId, "epoch id is wrong.")
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
	epochRewardDistribution, err = s.app.RewardKeeper.GetEpochRewardDistribution(s.ctx, 60*24, 1)
	s.Assert().Nil(err)
	s.Assert().NotNil(epochRewardDistribution)
	s.Assert().Equal(uint64(1), epochRewardDistribution.RewardProgramId)
	s.Assert().Equal(60*24, epochRewardDistribution.EpochId)
	s.Assert().Equal(int64(6), epochRewardDistribution.TotalShares)
	s.Assert().Equal(coin, epochRewardDistribution.TotalRewardsPool)
	s.Assert().Equal(false, epochRewardDistribution.EpochEnded)

	// last block of the epoch, should end but also increment
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + ((24 * 60 * 60 * 30) / 5) + 1)
	epoch.BeginBlocker(s.ctx, s.app.EpochKeeper)
	reward.EndBlocker(s.ctx, s.app.RewardKeeper)
	epochRewardDistribution, err = s.app.RewardKeeper.GetEpochRewardDistribution(s.ctx, 60*24, 1)
	s.Assert().Nil(err)
	s.Assert().NotNil(epochRewardDistribution)
	s.Assert().Equal(uint64(1), epochRewardDistribution.RewardProgramId)
	s.Assert().Equal(60*24, epochRewardDistribution.EpochId)
	s.Assert().Equal(int64(7), epochRewardDistribution.TotalShares)
	s.Assert().Equal(coin, epochRewardDistribution.TotalRewardsPool)
	s.Assert().Equal(true, epochRewardDistribution.EpochEnded)

	// epoch has ended, should not increment
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + ((24 * 60 * 60 * 30) / 5) + 1)
	epoch.BeginBlocker(s.ctx, s.app.EpochKeeper)
	reward.EndBlocker(s.ctx, s.app.RewardKeeper)
	epochRewardDistribution, err = s.app.RewardKeeper.GetEpochRewardDistribution(s.ctx, 60*24, 1)
	s.Assert().Nil(err)
	s.Assert().NotNil(epochRewardDistribution)
	s.Assert().Equal(uint64(1), epochRewardDistribution.RewardProgramId)
	s.Assert().Equal(60*24, epochRewardDistribution.EpochId)
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
	now := time.Now().UTC()
	nextEpochTime := now.Add(time.Hour + 24)
	rewardProgram := types.NewRewardProgram(1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", coin, maxCoin, now, nextEpochTime, 60*24, 10, types.NewEligibilityCriteria("criteria", &action), false)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)
	rewardProgramGet, err := s.app.RewardKeeper.GetRewardProgram(s.ctx, 1)
	s.Assert().NoError(err)

	s.Assert().Equal(rewardProgram.Coin, rewardProgramGet.Coin)
	s.Assert().Equal(rewardProgram.ProgramStartTime, rewardProgramGet.ProgramStartTime)
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
	epochRewardDistribution, err := s.app.RewardKeeper.GetEpochRewardDistribution(s.ctx, 60*24, 1)
	s.Assert().Nil(err)
	s.Assert().NotNil(epochRewardDistribution)
	s.Assert().Equal(epochRewardDistribution.RewardProgramId, uint64(1))
	s.Assert().Equal(epochRewardDistribution.EpochId, 60*24)
	s.Assert().Equal(epochRewardDistribution.TotalShares, int64(10))
	s.Assert().Equal(epochRewardDistribution.TotalRewardsPool, coin)
	s.Assert().Equal(true, epochRewardDistribution.EpochEnded)

	// epoch has ended, should not increment
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + ((24 * 60 * 60 * 30) / 5) + 1)
	epoch.BeginBlocker(s.ctx, s.app.EpochKeeper)
	reward.EndBlocker(s.ctx, s.app.RewardKeeper)
	epochRewardDistribution, err = s.app.RewardKeeper.GetEpochRewardDistribution(s.ctx, 60*24, 1)

	s.Assert().Nil(err)
	s.Assert().NotNil(epochRewardDistribution)
	s.Assert().Equal(uint64(1), epochRewardDistribution.RewardProgramId, "reward program id incorrect.")
	s.Assert().Equal(60*24, epochRewardDistribution.EpochId, "epoch id incorrect")
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
	now := time.Now().UTC()
	nextEpochTime := now.Add(time.Hour + 24)
	rewardProgram := types.NewRewardProgram(1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", coin, maxCoin, now, nextEpochTime, 60*24, 10, types.NewEligibilityCriteria("criteria", &action), false)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)
	rewardProgramGet, err := s.app.RewardKeeper.GetRewardProgram(s.ctx, 1)
	s.Assert().NoError(err)

	s.Assert().Equal(rewardProgram.Coin, rewardProgramGet.Coin)
	s.Assert().Equal(rewardProgram.ProgramStartTime, rewardProgramGet.ProgramStartTime)
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
	epochRewardDistribution, err := s.app.RewardKeeper.GetEpochRewardDistribution(s.ctx, 60*24, 1)
	s.Assert().Nil(err)
	s.Assert().NotNil(epochRewardDistribution)
	s.Assert().Equal(uint64(1), epochRewardDistribution.RewardProgramId, "reward program id incorrect.")
	s.Assert().Equal(60*24, epochRewardDistribution.EpochId, "epoch id incorrect")
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
	now := time.Now().UTC()
	nextEpochTime := now.Add(time.Hour + 24)
	rewardProgram := types.NewRewardProgram(1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", coin, maxCoin, now, nextEpochTime, 60*24, 10, types.NewEligibilityCriteria("criteria", &action), false)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)
	rewardProgramGet, err := s.app.RewardKeeper.GetRewardProgram(s.ctx, 1)
	s.Assert().NoError(err)

	s.Assert().Equal(rewardProgram.Coin, rewardProgramGet.Coin)
	s.Assert().Equal(rewardProgram.ProgramStartTime, rewardProgramGet.ProgramStartTime)
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
	epochRewardDistribution, err := s.app.RewardKeeper.GetEpochRewardDistribution(s.ctx, 60*24, 1)
	s.Assert().Nil(err)
	s.Assert().NotNil(epochRewardDistribution)
	s.Assert().Equal(epochRewardDistribution.RewardProgramId, uint64(1))
	s.Assert().Equal(epochRewardDistribution.EpochId, 60*24)
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

}*/

// Test no reward programs. Nothing should happen
func (s *KeeperTestSuite) TestDelegateAgainstNoRewardPrograms() {
	s.SetupTest()
	SetupEventHistoryWithDelegates(s)

	// Advance one day
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + (24 * 60 * 60))
	s.Assert().NotPanics(func() {
		reward.EndBlocker(s.ctx, s.app.RewardKeeper)
	})

	count := 0
	err := s.app.RewardKeeper.IterateShares(s.ctx, func(share types.Share) bool {
		count += 1
		return true
	})
	s.Assert().NoError(err, "iterate should not throw error")
	s.Assert().Equal(0, count, "no shares should be created")
}

// Test against inactive reward programs. They should not be updated
func (s *KeeperTestSuite) TestDelegateAgainstInactiveRewardPrograms() {
	s.SetupTest()
	SetupEventHistoryWithDelegates(s)

	// Create inactive reward program
	action := types.NewActionDelegate()
	action.MaximumActions = 10
	action.MinimumActions = 1
	action.MinimumActiveStakePercentile = 0.0
	action.MaximumActiveStakePercentile = 1.0
	minimumDelegation := sdk.NewInt64Coin("nhash", 0)
	maximumDelegation := sdk.NewInt64Coin("nhash", 10)
	action.MinimumDelegationAmount = &minimumDelegation
	action.MaximumDelegationAmount = &maximumDelegation

	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)

	now := time.Now().UTC()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		coin,
		maxCoin,
		now,
		60*60,
		3,
		types.NewEligibilityCriteria("criteria", &action),
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               10,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: 0,
						MaximumActiveStakePercentile: 1,
					},
				},
			},
		},
	)
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

	reward.EndBlocker(s.ctx, s.app.RewardKeeper)

	// Ensure no shares are granted
	count := 0
	err := s.app.RewardKeeper.IterateShares(s.ctx, func(share types.Share) bool {
		count += 1
		return true
	})
	s.Assert().NoError(err, "iterate should not throw error")
	s.Assert().Equal(0, count, "no shares should be created")

	programs, err := s.app.RewardKeeper.GetAllRewardPrograms(s.ctx)
	s.Assert().Equal(1, len(programs))
	s.Assert().NoError(err, "get all reward programs should not throw error")

	activePrograms, err := s.app.RewardKeeper.GetAllActiveRewardPrograms(s.ctx)
	s.Assert().NoError(err, "get all active reward programs should not throw error")
	s.Assert().Equal(0, len(activePrograms))
}

// Test against delegate reward program. No delegate happens.
func (s *KeeperTestSuite) TestNonDelegateAgainstRewardProgram() {
	s.SetupTest()
	setupEventHistory(s)

	// Create inactive reward program
	action := types.NewActionDelegate()
	action.MaximumActions = 10
	action.MinimumActions = 1
	action.MinimumActiveStakePercentile = 0.0
	action.MaximumActiveStakePercentile = 1.0
	minimumDelegation := sdk.NewInt64Coin("nhash", 0)
	maximumDelegation := sdk.NewInt64Coin("nhash", 10)
	action.MinimumDelegationAmount = &minimumDelegation
	action.MaximumDelegationAmount = &maximumDelegation

	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)

	now := time.Now().UTC()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		coin,
		maxCoin,
		now,
		60*60,
		3,
		types.NewEligibilityCriteria("criteria", &action),
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               10,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: 0,
						MaximumActiveStakePercentile: 1,
					},
				},
			},
		},
	)
	rewardProgram.Started = true
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

	reward.EndBlocker(s.ctx, s.app.RewardKeeper)

	// Ensure no shares are granted
	count := 0
	err := s.app.RewardKeeper.IterateShares(s.ctx, func(share types.Share) bool {
		count += 1
		return true
	})
	s.Assert().NoError(err, "iterate should not throw error")
	s.Assert().Equal(0, count, "no shares should be created")

	programs, err := s.app.RewardKeeper.GetAllRewardPrograms(s.ctx)
	s.Assert().Equal(1, len(programs))
	s.Assert().NoError(err, "get all reward programs should not throw error")

	activePrograms, err := s.app.RewardKeeper.GetAllActiveRewardPrograms(s.ctx)
	s.Assert().NoError(err, "get all active reward programs should not throw error")
	s.Assert().Equal(1, len(activePrograms))
}

func (suite *KeeperTestSuite) createDelegateEvents(validator, amount, sender, shares string) sdk.Events {
	attributes2 := []sdk.Attribute{
		sdk.NewAttribute("module", "staking"),
		sdk.NewAttribute("sender", sender),
	}
	attributes1 := []sdk.Attribute{
		sdk.NewAttribute("validator", validator),
		sdk.NewAttribute("amount", amount),
		sdk.NewAttribute("new_shares", shares),
	}
	event1 := sdk.NewEvent("delegate", attributes1...)
	event2 := sdk.NewEvent("message", attributes2...)
	events := sdk.Events{
		event1,
		event2,
	}
	return events
}

func (suite *KeeperTestSuite) createValidatorEvent(validator, amount, sender string) sdk.Events {
	attributes1 := []sdk.Attribute{
		sdk.NewAttribute("validator", validator),
		sdk.NewAttribute("amount", amount),
	}
	attributes2 := []sdk.Attribute{
		sdk.NewAttribute("module", "staking"),
		sdk.NewAttribute("sender", sender),
	}
	event1 := sdk.NewEvent("create_validator", attributes1...)
	event2 := sdk.NewEvent("message", attributes2...)
	events := sdk.Events{
		event1,
		event2,
	}
	return events
}

// Test against delegate reward program. Grant 1 share
func (s *KeeperTestSuite) TestSingleDelegate() {
	s.SetupTest()

	// Create inactive reward program
	action := types.NewActionDelegate()
	action.MaximumActions = 10
	action.MinimumActions = 1
	action.MinimumActiveStakePercentile = 0.0
	action.MaximumActiveStakePercentile = 1.0

	minimumDelegation := sdk.NewInt64Coin("nhash", 0)
	maximumDelegation := sdk.NewInt64Coin("nhash", 100)
	action.MinimumDelegationAmount = &minimumDelegation
	action.MaximumDelegationAmount = &maximumDelegation

	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)

	now := time.Now().UTC()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		coin,
		maxCoin,
		now,
		60*60,
		3,
		types.NewEligibilityCriteria("criteria", &action),
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               10,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: 0,
						MaximumActiveStakePercentile: 1,
					},
				},
			},
		},
	)
	rewardProgram.Started = true
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

	// We want to set the events here
	validators := getTestValidators(6, 6)
	delegates := s.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000")
	SetupEventHistory(s, delegates)
	reward.EndBlocker(s.ctx, s.app.RewardKeeper)

	// Ensure one share is granted
	count := 0
	err := s.app.RewardKeeper.IterateShares(s.ctx, func(share types.Share) bool {
		count += int(share.GetAmount())
		return true
	})
	s.Assert().NoError(err, "iterate should not throw error")
	s.Assert().Equal(1, count, "1 share should be created")
}

// Test against delegate reward program. Grant 2 share
func (s *KeeperTestSuite) TestMultipleDelegate() {
	s.SetupTest()

	// Create inactive reward program
	action := types.NewActionDelegate()
	action.MaximumActions = 10
	action.MinimumActions = 1
	action.MinimumActiveStakePercentile = 0.0
	action.MaximumActiveStakePercentile = 1.0

	minimumDelegation := sdk.NewInt64Coin("nhash", 0)
	maximumDelegation := sdk.NewInt64Coin("nhash", 100)
	action.MinimumDelegationAmount = &minimumDelegation
	action.MaximumDelegationAmount = &maximumDelegation

	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)

	now := time.Now().UTC()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		coin,
		maxCoin,
		now,
		60*60,
		3,
		types.NewEligibilityCriteria("criteria", &action),
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               10,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: 0,
						MaximumActiveStakePercentile: 1,
					},
				},
			},
		},
	)
	rewardProgram.Started = true
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

	// We want to set the events here
	validators := getTestValidators(6, 6)
	delegates := s.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000")
	delegates = delegates.AppendEvents(s.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000"))
	SetupEventHistory(s, delegates)
	reward.EndBlocker(s.ctx, s.app.RewardKeeper)

	// Ensure one share is granted
	count := 0
	err := s.app.RewardKeeper.IterateShares(s.ctx, func(share types.Share) bool {
		count += int(share.GetAmount())
		return true
	})
	s.Assert().NoError(err, "iterate should not throw error")
	s.Assert().Equal(2, count, "2 shares should be created")
}

// Test against delegate reward program. Not enough actions
func (s *KeeperTestSuite) TestDelegateBelowMinimumActions() {
	s.SetupTest()

	// Create inactive reward program
	action := types.NewActionDelegate()
	action.MaximumActions = 10
	action.MinimumActions = 1
	action.MinimumActiveStakePercentile = 0.0
	action.MaximumActiveStakePercentile = 1.0

	minimumDelegation := sdk.NewInt64Coin("nhash", 0)
	maximumDelegation := sdk.NewInt64Coin("nhash", 100)
	action.MinimumDelegationAmount = &minimumDelegation
	action.MaximumDelegationAmount = &maximumDelegation

	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)

	now := time.Now().UTC()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		coin,
		maxCoin,
		now,
		60*60,
		3,
		types.NewEligibilityCriteria("criteria", &action),
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               10,
						MaximumActions:               20,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: 0,
						MaximumActiveStakePercentile: 1,
					},
				},
			},
		},
	)
	rewardProgram.Started = true
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

	// We want to set the events here
	validators := getTestValidators(6, 6)
	delegates := s.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000")
	delegates = delegates.AppendEvents(s.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000"))
	SetupEventHistory(s, delegates)
	reward.EndBlocker(s.ctx, s.app.RewardKeeper)

	// Ensure no share is granted
	count := 0
	err := s.app.RewardKeeper.IterateShares(s.ctx, func(share types.Share) bool {
		count += int(share.GetAmount())
		return true
	})
	s.Assert().NoError(err, "iterate should not throw error")
	s.Assert().Equal(0, count, "no share should be created when below minimum actions")
}

// Test against delegate reward program. Too many actions
func (s *KeeperTestSuite) TestDelegateAboveMaximumActions() {
	s.SetupTest()

	// Create inactive reward program
	action := types.NewActionDelegate()
	action.MaximumActions = 10
	action.MinimumActions = 1
	action.MinimumActiveStakePercentile = 0.0
	action.MaximumActiveStakePercentile = 1.0

	minimumDelegation := sdk.NewInt64Coin("nhash", 0)
	maximumDelegation := sdk.NewInt64Coin("nhash", 100)
	action.MinimumDelegationAmount = &minimumDelegation
	action.MaximumDelegationAmount = &maximumDelegation

	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)

	now := time.Now().UTC()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		coin,
		maxCoin,
		now,
		60*60,
		3,
		types.NewEligibilityCriteria("criteria", &action),
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               0,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: 0,
						MaximumActiveStakePercentile: 1,
					},
				},
			},
		},
	)
	rewardProgram.Started = true
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

	// We want to set the events here
	validators := getTestValidators(6, 6)
	delegates := s.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000")
	delegates = delegates.AppendEvents(s.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000"))
	SetupEventHistory(s, delegates)
	reward.EndBlocker(s.ctx, s.app.RewardKeeper)

	// Ensure no share is granted
	count := 0
	err := s.app.RewardKeeper.IterateShares(s.ctx, func(share types.Share) bool {
		count += int(share.GetAmount())
		return true
	})
	s.Assert().NoError(err, "iterate should not throw error")
	s.Assert().Equal(0, count, "no share should be created when above maximum actions")
}

// Test against delegate reward program. Below delegation amount
func (s *KeeperTestSuite) TestDelegateBelowMinimumDelegation() {
	s.SetupTest()

	// Create inactive reward program
	action := types.NewActionDelegate()
	action.MaximumActions = 10
	action.MinimumActions = 1
	action.MinimumActiveStakePercentile = 0.0
	action.MaximumActiveStakePercentile = 1.0

	minimumDelegation := sdk.NewInt64Coin("nhash", 100)
	maximumDelegation := sdk.NewInt64Coin("nhash", 200)
	action.MinimumDelegationAmount = &minimumDelegation
	action.MaximumDelegationAmount = &maximumDelegation

	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)

	now := time.Now().UTC()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		coin,
		maxCoin,
		now,
		60*60,
		3,
		types.NewEligibilityCriteria("criteria", &action),
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               10,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: 0,
						MaximumActiveStakePercentile: 1,
					},
				},
			},
		},
	)
	rewardProgram.Started = true
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

	// We want to set the events here
	validators := getTestValidators(6, 6)
	delegates := s.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000")
	delegates = delegates.AppendEvents(s.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000"))
	SetupEventHistory(s, delegates)
	reward.EndBlocker(s.ctx, s.app.RewardKeeper)

	// Ensure no share is granted
	count := 0
	err := s.app.RewardKeeper.IterateShares(s.ctx, func(share types.Share) bool {
		count += int(share.GetAmount())
		return true
	})
	s.Assert().NoError(err, "iterate should not throw error")
	s.Assert().Equal(0, count, "no share should be created when below minimum delegation amount")
}

// Test against delegate reward program. Above delegation amount
func (s *KeeperTestSuite) TestDelegateAboveMaximumDelegation() {
	s.SetupTest()

	// Create inactive reward program
	action := types.NewActionDelegate()
	action.MaximumActions = 10
	action.MinimumActions = 1
	action.MinimumActiveStakePercentile = 0.0
	action.MaximumActiveStakePercentile = 1.0

	minimumDelegation := sdk.NewInt64Coin("nhash", 0)
	maximumDelegation := sdk.NewInt64Coin("nhash", 50)
	action.MinimumDelegationAmount = &minimumDelegation
	action.MaximumDelegationAmount = &maximumDelegation

	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)

	now := time.Now().UTC()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		coin,
		maxCoin,
		now,
		60*60,
		3,
		types.NewEligibilityCriteria("criteria", &action),
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               10,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: 0,
						MaximumActiveStakePercentile: 1,
					},
				},
			},
		},
	)
	rewardProgram.Started = true
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

	// We want to set the events here
	validators := getTestValidators(6, 6)
	delegates := s.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000")
	delegates = delegates.AppendEvents(s.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000"))
	SetupEventHistory(s, delegates)
	reward.EndBlocker(s.ctx, s.app.RewardKeeper)

	// Ensure no share is granted
	count := 0
	err := s.app.RewardKeeper.IterateShares(s.ctx, func(share types.Share) bool {
		count += int(share.GetAmount())
		return true
	})
	s.Assert().NoError(err, "iterate should not throw error")
	s.Assert().Equal(0, count, "no share should be created when above maximum delegation amount")
}

// Test against delegate reward program. Below percentile
func (s *KeeperTestSuite) TestDelegateBelowMinimumPercentile() {
	s.SetupTest()

	// Create inactive reward program
	action := types.NewActionDelegate()
	action.MaximumActions = 10
	action.MinimumActions = 1
	action.MinimumActiveStakePercentile = 0.0
	action.MaximumActiveStakePercentile = 1.0

	minimumDelegation := sdk.NewInt64Coin("nhash", 0)
	maximumDelegation := sdk.NewInt64Coin("nhash", 100)
	action.MinimumDelegationAmount = &minimumDelegation
	action.MaximumDelegationAmount = &maximumDelegation

	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)

	now := time.Now().UTC()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		coin,
		maxCoin,
		now,
		60*60,
		3,
		types.NewEligibilityCriteria("criteria", &action),
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               10,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: 0.70,
						MaximumActiveStakePercentile: 1.0,
					},
				},
			},
		},
	)
	rewardProgram.Started = true
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

	// We want to set the events here
	validators := getTestValidators(6, 6)
	delegates := s.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000")
	delegates = delegates.AppendEvents(s.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000"))
	SetupEventHistory(s, delegates)
	reward.EndBlocker(s.ctx, s.app.RewardKeeper)

	// Ensure no share is granted
	count := 0
	err := s.app.RewardKeeper.IterateShares(s.ctx, func(share types.Share) bool {
		count += int(share.GetAmount())
		return true
	})
	s.Assert().NoError(err, "iterate should not throw error")
	s.Assert().Equal(0, count, "no share should be created when below minimum delegation percentage")
}

// Test against delegate reward program. Above percentile
func (s *KeeperTestSuite) TestDelegateAboveMaximumPercentile() {
	s.SetupTest()

	// Create inactive reward program
	action := types.NewActionDelegate()
	action.MaximumActions = 10
	action.MinimumActions = 1
	action.MinimumActiveStakePercentile = 0.0
	action.MaximumActiveStakePercentile = 1.0

	minimumDelegation := sdk.NewInt64Coin("nhash", 0)
	maximumDelegation := sdk.NewInt64Coin("nhash", 100)
	action.MinimumDelegationAmount = &minimumDelegation
	action.MaximumDelegationAmount = &maximumDelegation

	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)

	now := time.Now().UTC()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		coin,
		maxCoin,
		now,
		60*60,
		3,
		types.NewEligibilityCriteria("criteria", &action),
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               10,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: 0.0,
						MaximumActiveStakePercentile: 0.20,
					},
				},
			},
		},
	)
	rewardProgram.Started = true
	s.app.RewardKeeper.SetRewardProgram(s.ctx, rewardProgram)

	// We want to set the events here
	validators := getTestValidators(6, 6)
	delegates := s.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000")
	delegates = delegates.AppendEvents(s.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000"))
	SetupEventHistory(s, delegates)
	reward.EndBlocker(s.ctx, s.app.RewardKeeper)

	// Ensure no share is granted
	count := 0
	err := s.app.RewardKeeper.IterateShares(s.ctx, func(share types.Share) bool {
		count += int(share.GetAmount())
		return true
	})
	s.Assert().NoError(err, "iterate should not throw error")
	s.Assert().Equal(0, count, "no share should be created when above maximum delegation percentage")
}
