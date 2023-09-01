package keeper_test

import (
	"github.com/provenance-io/provenance/x/oracle/types"
)

func (s *KeeperTestSuite) TestExportGenesis() {
	genesis := s.app.OracleKeeper.ExportGenesis(s.ctx)
	s.Assert().Equal("oracle", genesis.PortId, "should export the correct port")
	s.Assert().Equal("", genesis.Oracle, "should export the correct oracle address")
}

func (s *KeeperTestSuite) TestInitGenesis() {
	tests := []struct {
		name     string
		genesis  *types.GenesisState
		err      string
		mockPort bool
	}{
		{
			name:    "success - valid genesis state",
			genesis: types.NewGenesisState("jackthecat", "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
		},
		{
			name:    "success - valid genesis state with empty oracle",
			genesis: types.NewGenesisState("jackthecat", ""),
		},
		{
			name:    "failure - invalid port",
			genesis: types.NewGenesisState("", ""),
			err:     "identifier cannot be blank: invalid identifier",
		},
		{
			name:    "failure - invalid oracle",
			genesis: types.NewGenesisState("jackthecat", "abc"),
			err:     "decoding bech32 failed: invalid bech32 string length 3",
		},
		{
			name:    "success - works with existing port",
			genesis: types.NewGenesisState("oracle", ""),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.mockPort {
				s.app.OracleKeeper.BindPort(s.ctx, "test")
			}
			if len(tc.err) > 0 {
				s.Assert().PanicsWithError(tc.err, func() {
					s.app.OracleKeeper.InitGenesis(s.ctx, tc.genesis)
				}, "invalid init genesis should cause panic")
			} else {
				s.app.OracleKeeper.InitGenesis(s.ctx, tc.genesis)
				oracle, _ := s.app.OracleKeeper.GetOracle(s.ctx)
				s.Assert().Equal(tc.genesis.PortId, s.app.OracleKeeper.GetPort(s.ctx), "should correctly set the port")
				s.Assert().True(s.app.OracleKeeper.IsBound(s.ctx, tc.genesis.PortId), "should bind the port")
				s.Assert().Equal(tc.genesis.Oracle, oracle.String(), "should get the correct oracle address")
			}
		})
	}
}
