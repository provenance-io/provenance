package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	epoch "github.com/provenance-io/provenance/x/epoch"
	"github.com/provenance-io/provenance/x/reward"
	abci "github.com/tendermint/tendermint/abci/types"
	"testing"

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

	// queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	// TODO
	// types.RegisterQueryServer(queryHelper, nil)
	// suite.queryClient = types.NewQueryClient(queryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

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
	for i := 0; i < 5; i++ {
		s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + ((24 * 60 * 60 * 30) / 5) + 1)
		epoch.BeginBlocker(s.ctx, s.app.EpochKeeper)
	}
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

	validators[0] = keeper.TestingUpdateValidator(s.app.StakingKeeper, s.ctx, validators[0], true)
	validators[1] = keeper.TestingUpdateValidator(s.app.StakingKeeper, s.ctx, validators[1], true)
	validators[2] = keeper.TestingUpdateValidator(s.app.StakingKeeper, s.ctx, validators[2], true)

	// first add a validators[0] to delegate too
	//bond1to1 := stakingtypes.NewDelegation(addrDels[0], valAddrs[0], sdk.NewDec(9))

	delegation := sdk.NewInt64Coin("hotdog", 10)

	event0 := sdk.Event{
		Type:       sdk.EventTypeMessage,
		Attributes: []abci.EventAttribute{},
	}
	event0.Attributes = append(
		event0.Attributes,
		abci.EventAttribute{Key: []byte(stakingtypes.AttributeKeyValidator), Value: []byte(validators[0].OperatorAddress)},
	)

	event1 := sdk.Event{
		Type:       sdk.EventTypeMessage,
		Attributes: []abci.EventAttribute{},
	}
	event1.Attributes = append(
		event1.Attributes,
		abci.EventAttribute{Key: []byte(stakingtypes.AttributeValueCategory), Value: []byte(addrDels[0].String())},
	)

	events := []abci.Event{
		{Type: stakingtypes.EventTypeDelegate, Attributes: []abci.EventAttribute{{Key: []byte(stakingtypes.AttributeKeyValidator), Value: []byte(validators[0].OperatorAddress), Index: true}}},
		{Type: stakingtypes.EventTypeDelegate, Attributes: []abci.EventAttribute{{Key: []byte(sdk.AttributeKeyAmount), Value: []byte(delegation.Amount.String()), Index: true}}},
		{Type: stakingtypes.EventTypeDelegate, Attributes: []abci.EventAttribute{{Key: []byte(stakingtypes.AttributeKeyNewShares), Value: []byte(sdk.NewDec(10).String()), Index: true}}},
		{Type: sdk.EventTypeMessage, Attributes: []abci.EventAttribute{{Key: []byte(sdk.AttributeKeyModule), Value: []byte(stakingtypes.AttributeValueCategory), Index: true}}},
		{Type: sdk.EventTypeMessage, Attributes: []abci.EventAttribute{{Key: []byte(sdk.AttributeKeySender), Value: []byte(addrDels[0].String()), Index: true}}},
	}

	s.ctx = s.ctx.WithEventManager(sdk.NewEventManagerWithHistory(events))
	reward.EndBlocker(s.ctx, s.app.RewardKeeper)
	rewardProgramGet, err = s.app.RewardKeeper.GetRewardProgram(s.ctx, 1)
	s.Assert().Equal(rewardProgram.Coin, rewardProgramGet.Coin)
	s.Assert().Equal(rewardProgram.Coin, rewardProgramGet.Coin)
	s.Assert().Equal(rewardProgram.StartEpoch, rewardProgramGet.StartEpoch)
	s.Assert().Equal(rewardProgram.Id, rewardProgramGet.Id)
	s.Assert().Equal(rewardProgram.EligibilityCriteria.Action.TypeUrl, rewardProgramGet.EligibilityCriteria.Action.TypeUrl)
	s.Assert().Equal(rewardProgramGet.Expired, false)

	// get reward epoch distribution
	epochRewardDistribution, err := s.app.RewardKeeper.GetEpochRewardDistribution(s.ctx, "day", 1)
	s.Assert().Nil(err)
	s.Assert().NotNil(epochRewardDistribution)
	s.Assert().Equal(epochRewardDistribution.RewardProgramId, uint64(1))
	s.Assert().Equal(epochRewardDistribution.EpochId, "day")
	s.Assert().Equal(epochRewardDistribution.TotalShares, int64(1))
	s.Assert().Equal(epochRewardDistribution.TotalRewardsPool, coin)
}
