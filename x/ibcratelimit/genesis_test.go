package ibcratelimit_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/ibcratelimit/types"
)

type GenesisTestSuite struct {
	suite.Suite

	app *simapp.App
	ctx sdk.Context
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) SetupTest() {
	suite.Setup()
}

func (s *GenesisTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{Time: time.Now().UTC()})
	s.ctx = s.ctx.WithBlockHeight(1)
}

func (suite *GenesisTestSuite) TestInitExportGenesis() {
	testAddress := sdk.AccAddress([]byte("addr1_______________")).String()
	suite.SetupTest()
	k := suite.app.RateLimitingICS4Wrapper

	initialGenesis := types.GenesisState{
		Params: types.Params{
			ContractAddress: testAddress,
		},
	}

	k.InitGenesis(suite.ctx, initialGenesis)

	suite.Require().Equal(testAddress, k.GetParams(suite.ctx).ContractAddress)

	exportedGenesis := k.ExportGenesis(suite.ctx)

	suite.Require().Equal(initialGenesis, *exportedGenesis)
}
