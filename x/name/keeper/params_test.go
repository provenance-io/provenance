package keeper_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/name/types"
	"github.com/stretchr/testify/suite"
)

type NameParamTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context
}

func (s *NameParamTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: time.Now()})
}

func TestNameParamTestSuite(t *testing.T) {
	suite.Run(t, new(NameParamTestSuite))
}

func (s *NameParamTestSuite) TestGetSetParams() {
	defaultParams := s.app.NameKeeper.GetParams(s.ctx)
	s.Require().Equal(types.DefaultMaxNameLevels, defaultParams.MaxNameLevels, "Default MaxNameLevels should match")
	s.Require().Equal(types.DefaultMaxSegmentLength, defaultParams.MaxSegmentLength, "Default MaxSegmentLength should match")
	s.Require().Equal(types.DefaultMinSegmentLength, defaultParams.MinSegmentLength, "Default MinSegmentLength should match")
	s.Require().Equal(types.DefaultAllowUnrestrictedNames, defaultParams.AllowUnrestrictedNames, "Default AllowUnrestrictedNames should match")

	newMaxNameLevels := uint32(10)
	newMaxSegmentLength := uint32(15)
	newMinSegmentLength := uint32(2)
	newAllowUnrestrictedNames := false

	newParams := types.Params{
		MaxNameLevels:          newMaxNameLevels,
		MaxSegmentLength:       newMaxSegmentLength,
		MinSegmentLength:       newMinSegmentLength,
		AllowUnrestrictedNames: newAllowUnrestrictedNames,
	}

	s.app.NameKeeper.SetParams(s.ctx, newParams)

	updatedParams := s.app.NameKeeper.GetParams(s.ctx)
	s.Require().Equal(newMaxNameLevels, updatedParams.MaxNameLevels, "Updated MaxNameLevels should match")
	s.Require().Equal(newMaxSegmentLength, updatedParams.MaxSegmentLength, "Updated MaxSegmentLength should match")
	s.Require().Equal(newMinSegmentLength, updatedParams.MinSegmentLength, "Updated MinSegmentLength should match")
	s.Require().Equal(newAllowUnrestrictedNames, updatedParams.AllowUnrestrictedNames, "Updated AllowUnrestrictedNames should match")
}
