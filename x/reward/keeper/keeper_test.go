package keeper_test

import (
	"testing"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/reward/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
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
		types.NewRewardProgram(1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", coin,maxCoin, "day", 1, 10, types.NewEligibilityCriteria("criteria", &action),false,1,10),
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
		types.NewRewardProgram(1, "", sdk.NewInt64Coin("nhash", 100),sdk.NewInt64Coin("nhash", 10), "day", 1, 10, types.NewEligibilityCriteria("criteria", &action),false,1,10),
	}

	s.Assert().Panics(func() { s.app.RewardKeeper.InitGenesis(s.ctx, &rewardData) })
}

// create a test that creates a reward program for an epoch
// make a transfer, with delegation
func (s *KeeperTestSuite) TestCreateRewardClaim() {
	action := types.NewActionDelegate()
	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)
	rewardProgram := types.NewRewardProgram(1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", coin,maxCoin, "day", 1, 10, types.NewEligibilityCriteria("criteria", &action),false,1,10)
	s.app.RewardKeeper.SetRewardProgram(s.ctx,rewardProgram)
	rewardProgramGet,found := s.app.RewardKeeper.GetRewardProgram(s.ctx,1)
	s.Assert().Equal(true,found,"reward program should be found.")

	s.Assert().Equal(rewardProgram.Coin,rewardProgramGet.Coin)
	s.Assert().Equal(rewardProgram.Coin,rewardProgramGet.Coin)
	s.Assert().Equal(rewardProgram.StartEpoch,rewardProgramGet.StartEpoch)
	s.Assert().Equal(rewardProgram.Id,rewardProgramGet.Id)
	s.Assert().Equal(rewardProgram.EligibilityCriteria.Action.TypeUrl,rewardProgramGet.EligibilityCriteria.Action.TypeUrl)


}
