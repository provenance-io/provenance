package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestGetSetOracle() {
	tests := []struct {
		name    string
		address string
		err     string
	}{
		{
			name: "failure - address not set",
			err:  "missing oracle address",
		},
		{
			name:    "success - address can be set",
			address: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if len(tc.address) > 0 {
				s.app.OracleKeeper.SetOracle(s.ctx, sdk.MustAccAddressFromBech32(tc.address))
			}
			oracle, err := s.app.OracleKeeper.GetOracle(s.ctx)

			if len(tc.err) > 0 {
				s.Assert().EqualError(err, tc.err, "should throw the correct error")
			} else {
				s.Assert().NoError(err, "should not throw an error")
				s.Assert().Equal(tc.address, oracle.String(), "should get back the set address")
			}
		})
	}
}
