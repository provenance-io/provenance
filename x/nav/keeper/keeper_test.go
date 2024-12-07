package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/nav"

	. "github.com/provenance-io/provenance/x/nav/keeper"
)

type TestSuite struct {
	suite.Suite

	app       *app.App
	ctx       sdk.Context
	navKeeper *Keeper
}

func (s *TestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false)
	s.navKeeper = s.app.NAVKeeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestKeeper_SetNAVs() {
	tests := []struct {
		name   string
		height int64
		source string
		navs   []*nav.NetAssetValue
		expErr string
	}{
		// negative height
		// no navs
		// one nav: good
		// one nav: bad
		// three navs: all good
		// three navs: first bad
		// three navs: second bad
		// three navs: third bad
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctx := s.ctx.WithBlockHeight(tc.height)
			var err error
			testFunc := func() {
				err = s.navKeeper.SetNAVs(ctx, tc.source, tc.navs...)
			}
			s.Require().NotPanics(testFunc, "SetNAVs")
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "SetNAVs error")
			if len(tc.expErr) == 0 && err == nil && len(tc.navs) > 0 {
				for i := range tc.navs {
					exp := tc.navs[i].AsRecord(uint64(tc.height), tc.source)
					act, err2 := s.navKeeper.NAVs().Get(s.ctx, collections.Join(exp.Assets.Denom, exp.Price.Denom))
					if s.Assert().NoError(err2, "[%d]: trying to get %s from the navs collection", i, exp) {
						s.Assert().Equal(*exp, act, "[%d]: nav read from the collection")
					}
				}
			}
		})
	}
}

// TODO: func (s *TestSuite) TestKeeper_SetNAVsAtHeight() {}

// TODO: func (s *TestSuite) TestKeeper_SetNAVRecords() {}

// TODO: func (s *TestSuite) TestKeeper_GetNAVRecord() {}

// TODO: func (s *TestSuite) TestKeeper_GetNAVRecords() {}

// TODO: func (s *TestSuite) TestKeeper_GetAllNAVRecords() {}
