package keeper_test

func (s *KeeperTestSuite) TestSetGetLastQueryPacketSeq() {
	sequence := s.app.OracleKeeper.GetLastQueryPacketSeq(s.ctx)
	s.Assert().Equal(int(1), int(sequence), "should have the correct initial sequence")
	s.app.OracleKeeper.SetLastQueryPacketSeq(s.ctx, 5)
	sequence = s.app.OracleKeeper.GetLastQueryPacketSeq(s.ctx)
	s.Assert().Equal(int(5), int(sequence), "should have the correct updated sequence")
}
