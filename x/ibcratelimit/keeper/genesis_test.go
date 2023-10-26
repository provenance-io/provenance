package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/ibcratelimit"
)

type GenesisTestSuite struct {
	suite.Suite

	app *simapp.App
	ctx sdk.Context
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (s *GenesisTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.ctx = s.ctx.WithBlockHeight(0)
}

func (s *GenesisTestSuite) TestInitExportGenesis() {
	testAddress := sdk.AccAddress([]byte("addr1_______________")).String()
	k := s.app.RateLimitingKeeper

	initialGenesis := ibcratelimit.GenesisState{
		Params: ibcratelimit.Params{
			ContractAddress: testAddress,
		},
	}

	k.InitGenesis(s.ctx, &initialGenesis)

	s.Require().Equal(testAddress, k.GetContractAddress(s.ctx))

	exportedGenesis := k.ExportGenesis(s.ctx)

	s.Require().Equal(initialGenesis, *exportedGenesis)
}
