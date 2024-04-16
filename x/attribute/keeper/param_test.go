package keeper_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/attribute/types"
	"github.com/stretchr/testify/suite"
)

type ParamTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	startBlockTime time.Time
}

func (s *ParamTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	s.startBlockTime = time.Now()
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: s.startBlockTime})
}

func TestParamTestSuite(t *testing.T) {
	suite.Run(t, new(ParamTestSuite))
}

func (s *ParamTestSuite) TestGetSetParams() {
	defaultParams := s.app.AttributeKeeper.GetParams(s.ctx)
	s.Require().Equal(int(types.DefaultMaxValueLength), int(defaultParams.MaxValueLength), "GetParams() Default max value length should match")

	defaultValueLength := s.app.AttributeKeeper.GetMaxValueLength(s.ctx)
	s.Require().Equal(int(types.DefaultMaxValueLength), int(defaultValueLength), "GetMaxValueLength() Default max value length should match")

	newMaxValueLength := uint32(2048)
	newParams := types.Params{
		MaxValueLength: newMaxValueLength,
	}
	s.app.AttributeKeeper.SetParams(s.ctx, newParams)

	updatedParams := s.app.AttributeKeeper.GetParams(s.ctx)
	s.Require().Equal(int(newMaxValueLength), int(updatedParams.MaxValueLength), "GetParams() Updated max value length should match")

	updatedValueLength := s.app.AttributeKeeper.GetMaxValueLength(s.ctx)
	s.Require().Equal(int(newMaxValueLength), int(updatedValueLength), "GetMaxValueLength() Updated max value length should match")
}
