package app_test

import (
	"github.com/provenance-io/provenance/app"
	"github.com/stretchr/testify/suite"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type UpgradeTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context
}

func TestDumbTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
}

func (s *UpgradeTestSuite) TestIncreaseMaxCommissions() {
	// make sure at least one doesn't have the new value so that there's something that's actually being tested.
	expectedMaxRate := sdk.OneDec()
	canTest := false
	for i, validator := range s.app.StakingKeeper.GetAllValidators(s.ctx) {
		s.T().Logf("Before: validator[%d] has Commission.MaxRate = %s", i, validator.Commission.MaxRate.String())
		if !expectedMaxRate.Equal(validator.Commission.MaxRate) {
			canTest = true
		}
	}
	s.Require().True(canTest, "all validators already have Commission.MaxRate = %s", expectedMaxRate.String())

	app.IncreaseMaxCommissions(s.ctx, s.app)

	for i, validator := range s.app.StakingKeeper.GetAllValidators(s.ctx) {
		s.T().Logf("After: validator[%d] has Commission.MaxRate = %s", i, validator.Commission.MaxRate.String())
		s.Assert().Equal(expectedMaxRate, validator.Commission.MaxRate, "validator[%d].Commission.MaxRate", i)
	}
}
