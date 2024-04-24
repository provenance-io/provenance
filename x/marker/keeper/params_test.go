package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/marker/types"
)

type ParamTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context
}

func (s *ParamTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: time.Now()})
}

func TestParamTestSuite(t *testing.T) {
	suite.Run(t, new(ParamTestSuite))
}

func (s *ParamTestSuite) TestGetSetParams() {
	defaultParams := s.app.MarkerKeeper.GetParams(s.ctx)
	s.Require().Equal(types.DefaultMaxTotalSupply, defaultParams.MaxTotalSupply, "Default MaxTotalSupply should match")
	s.Require().Equal(types.DefaultEnableGovernance, defaultParams.EnableGovernance, "Default EnableGovernance should match")
	s.Require().Equal(types.DefaultUnrestrictedDenomRegex, defaultParams.UnrestrictedDenomRegex, "Default UnrestrictedDenomRegex should match")
	s.Require().Equal(types.StringToBigInt(types.DefaultMaxSupply), defaultParams.MaxSupply, "Default MaxSupply should match")

	newMaxTotalSupply := uint64(2000000)
	newEnableGovernance := false
	newUnrestrictedDenomRegex := "xyz.*"
	newMaxSupply := "3000000"

	newParams := types.Params{
		MaxTotalSupply:         newMaxTotalSupply,
		EnableGovernance:       newEnableGovernance,
		UnrestrictedDenomRegex: newUnrestrictedDenomRegex,
		MaxSupply:              types.StringToBigInt(newMaxSupply),
	}

	s.app.MarkerKeeper.SetParams(s.ctx, newParams)

	updatedParams := s.app.MarkerKeeper.GetParams(s.ctx)
	s.Require().Equal(newMaxTotalSupply, updatedParams.MaxTotalSupply, "Updated MaxTotalSupply should match")
	s.Require().Equal(newEnableGovernance, updatedParams.EnableGovernance, "Updated EnableGovernance should match")
	s.Require().Equal(newUnrestrictedDenomRegex, updatedParams.UnrestrictedDenomRegex, "Updated UnrestrictedDenomRegex should match")
	s.Require().Equal(types.StringToBigInt(newMaxSupply), updatedParams.MaxSupply, "Updated MaxSupply should match")
}
