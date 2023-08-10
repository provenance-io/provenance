package keeper_test

import (
	"github.com/provenance-io/provenance/x/oracle/types"
)

func (s *KeeperTestSuite) TestExportGenesis() {
	genesis := s.app.OracleKeeper.ExportGenesis(s.ctx)
	s.Assert().Equal(types.DefaultParams(), genesis.Params, "should export the correct params")
	s.Assert().Equal("oracle", genesis.PortId, "should export the correct port")
	s.Assert().Equal(int(1), int(genesis.Sequence), "should export the correct sequence number")
	s.Assert().Equal("", genesis.Oracle, "should export the correct oracle address")
}

func (s *KeeperTestSuite) TestInitGenesis() {
	tests := []struct {
		name    string
		genesis *types.GenesisState
		err     string
	}{
		{
			name:    "success - valid genesis state",
			genesis: types.NewGenesisState("jackthecat", types.DefaultParams(), 1, "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
		},
		{
			name:    "success - valid genesis state with empty oracle",
			genesis: types.NewGenesisState("jackthecat", types.DefaultParams(), 1, ""),
		},
		{
			name:    "failure - invalid sequence number",
			genesis: types.NewGenesisState("jackthecat", types.DefaultParams(), 0, ""),
			err:     "sequence 0 is invalid, must be greater than 0",
		},
		{
			name:    "failure - invalid port",
			genesis: types.NewGenesisState("", types.DefaultParams(), 1, ""),
			err:     "identifier cannot be blank: invalid identifier",
		},
		{
			name:    "failure - invalid oracle",
			genesis: types.NewGenesisState("jackthecat", types.DefaultParams(), 1, "abc"),
			err:     "decoding bech32 failed: invalid bech32 string length 3",
		},
		{
			name:    "success - works with existing port",
			genesis: types.NewGenesisState("oracle", types.DefaultParams(), 1, ""),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {

			if len(tc.err) > 0 {
				s.Assert().PanicsWithError(tc.err, func() {
					s.app.OracleKeeper.InitGenesis(s.ctx, tc.genesis)
				}, "invalid init genesis should cause panic")
			} else {
				s.app.OracleKeeper.InitGenesis(s.ctx, tc.genesis)
				oracle, _ := s.app.OracleKeeper.GetOracle(s.ctx)
				s.Assert().Equal(tc.genesis.Params, s.app.OracleKeeper.GetParams(s.ctx), "should correctly set params")
				s.Assert().Equal(tc.genesis.PortId, s.app.OracleKeeper.GetPort(s.ctx), "should correctly set the port")
				s.Assert().True(s.app.OracleKeeper.IsBound(s.ctx, tc.genesis.PortId), "should bind the port")
				s.Assert().Equal(int(tc.genesis.Sequence), int(s.app.OracleKeeper.GetLastQueryPacketSeq(s.ctx)), "should set the last sequence number")
				s.Assert().Equal(tc.genesis.Oracle, oracle.String(), "should get the correct oracle address")
			}
		})
	}
}
