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
	action := types.NewActionDelegate(1, 100)
	coin := sdk.NewInt64Coin("jackthecat", 100)
	var rewardData types.GenesisState
	rewardData.RewardPrograms = []types.RewardProgram{
		types.NewRewardProgram(1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", coin, "day", 1, 10, types.NewEligibilityCriteria("criteria", &action)),
	}
	sharesPerEpoch := types.SharesPerEpochPerRewardsProgram{RewardProgramId: 1, Shares: 2, EpochId: "week", EpochEndHeight: 1000, Claimed: false, ExpirationHeight: 11000, Expired: false, TotalShares: 420, TotalRewards: coin}
	rewardData.RewardClaims = []types.RewardClaim{types.NewRewardClaim("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", []*types.SharesPerEpochPerRewardsProgram{&sharesPerEpoch})}
	rewardData.EligibilityCriterias = []types.EligibilityCriteria{types.NewEligibilityCriteria("delegate", &action)}
	rewardData.EpochRewardDistributions = []types.EpochRewardDistribution{types.NewEpochRewardDistribution("day", 1,
		coin,
		10,
	)}
	rewardData.ActionDelegate = types.NewActionDelegate(1, 100)
	rewardData.ActionTransferDelegations = types.NewActionTransferDelegations(1, 100)
	s.Assert().NotPanics(func() { s.app.RewardKeeper.InitGenesis(s.ctx, &rewardData) })
	s.Assert().NotPanics(func() { s.app.RewardKeeper.ExportGenesis(s.ctx) })

	rewardData.RewardPrograms = []types.RewardProgram{
		types.NewRewardProgram(1, "", sdk.NewInt64Coin("nhash", 100), "day", 1, 10, types.NewEligibilityCriteria("criteria", &action)),
	}

	s.Assert().Panics(func() { s.app.RewardKeeper.InitGenesis(s.ctx, &rewardData) })
}
