package keeper_test

import "github.com/provenance-io/provenance/x/oracle/types"

func (s *KeeperTestSuite) TestGetParams() {
	params := s.app.OracleKeeper.GetParams(s.ctx)
	s.Assert().Equal(types.NewParams(), params, "should return correct params")
}
