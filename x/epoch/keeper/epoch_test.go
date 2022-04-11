package keeper_test

import (
	"github.com/provenance-io/provenance/x/epoch/types"
)

func (suite *KeeperTestSuite) TestEpochLifeCycle() {
	suite.SetupTest()

	epochInfo := types.EpochInfo{
		Identifier:              "monthly",
		StartHeight:             0,
		Duration:                uint64((24 * 60 * 60 * 30) / 5),
		CurrentEpoch:            0,
		CurrentEpochStartHeight: 0,
	}
	suite.app.EpochKeeper.SetEpochInfo(suite.ctx, epochInfo)
	epochInfoSaved := suite.app.EpochKeeper.GetEpochInfo(suite.ctx, "monthly")
	suite.Require().Equal(epochInfo, epochInfoSaved)

	allEpochs := suite.app.EpochKeeper.AllEpochInfos(suite.ctx)
	suite.Require().Len(allEpochs, 4)
	suite.Require().Equal(allEpochs[0].Identifier, "day")   // alphabetical order
	suite.Require().Equal(allEpochs[1].Identifier, "month") // alphabetical order
	suite.Require().Equal(allEpochs[2].Identifier, "monthly")
	suite.Require().Equal(allEpochs[3].Identifier, "week")
}
