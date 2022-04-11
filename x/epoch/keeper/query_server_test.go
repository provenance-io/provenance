package keeper_test

import (
	gocontext "context"

	"github.com/provenance-io/provenance/x/epoch/types"
)

func (suite *KeeperTestSuite) TestQueryEpochInfos() {
	suite.SetupTest()
	queryClient := suite.queryClient

	chainBlockHeight := suite.ctx.BlockHeight()

	// Invalid param
	epochInfosResponse, err := queryClient.EpochInfos(gocontext.Background(), &types.QueryEpochInfosRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(epochInfosResponse.Epochs, 3)

	// check if EpochInfos are correct
	suite.Require().Equal(epochInfosResponse.Epochs[0].Identifier, "day")
	suite.Require().Equal(epochInfosResponse.Epochs[0].StartHeight, uint64(0))
	suite.Require().Equal(epochInfosResponse.Epochs[0].Duration, uint64((24*60*60)/5))
	suite.Require().Equal(epochInfosResponse.Epochs[0].CurrentEpoch, uint64(chainBlockHeight))
	suite.Require().Equal(epochInfosResponse.Epochs[0].CurrentEpochStartHeight, uint64(0))

	suite.Require().Equal(epochInfosResponse.Epochs[1].Identifier, "month")
	suite.Require().Equal(epochInfosResponse.Epochs[1].StartHeight, uint64(0))
	suite.Require().Equal(epochInfosResponse.Epochs[1].Duration, uint64((24*60*60*30)/5))
	suite.Require().Equal(epochInfosResponse.Epochs[1].CurrentEpoch, uint64(chainBlockHeight))
	suite.Require().Equal(epochInfosResponse.Epochs[1].CurrentEpochStartHeight, uint64(0))

	suite.Require().Equal(epochInfosResponse.Epochs[2].Identifier, "week")
	suite.Require().Equal(epochInfosResponse.Epochs[2].StartHeight, uint64(0))
	suite.Require().Equal(epochInfosResponse.Epochs[2].Duration, uint64((24*60*60*7)/5))
	suite.Require().Equal(epochInfosResponse.Epochs[2].CurrentEpoch, uint64(chainBlockHeight))
	suite.Require().Equal(epochInfosResponse.Epochs[2].CurrentEpochStartHeight, uint64(0))
}
