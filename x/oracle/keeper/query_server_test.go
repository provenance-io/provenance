package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/oracle/keeper"
	"github.com/provenance-io/provenance/x/oracle/types"
)

func (s *KeeperTestSuite) TestOracleAddress() {
	tests := []struct {
		name     string
		req      *types.QueryOracleAddressRequest
		expected *types.QueryOracleAddressResponse
		oracle   string
		err      string
	}{
		{
			name:     "failure - should handle nil request",
			req:      nil,
			expected: &types.QueryOracleAddressResponse{Address: ""},
		},
		{
			name:     "success - should return correct oracle address",
			req:      &types.QueryOracleAddressRequest{},
			expected: &types.QueryOracleAddressResponse{Address: ""},
		},
		{
			name:     "success - should return correct oracle address",
			req:      &types.QueryOracleAddressRequest{},
			oracle:   "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			expected: &types.QueryOracleAddressResponse{Address: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if len(tc.oracle) > 0 {
				s.app.OracleKeeper.SetOracle(s.ctx, sdk.MustAccAddressFromBech32(tc.oracle))
			}
			resp, err := s.app.OracleKeeper.OracleAddress(s.ctx, tc.req)
			if len(tc.err) > 0 {
				s.Assert().EqualError(err, tc.err, "should return the correct error")
				s.Assert().Nil(resp, "response should be nil")
			} else {
				s.Assert().NoError(err, "should not return an error")
				s.Assert().Equal(tc.expected, resp, "should return the correct response")
			}
		})
	}
}

func (s *KeeperTestSuite) TestOracle() {
	tests := []struct {
		name        string
		req         *types.QueryOracleRequest
		expected    *types.QueryOracleResponse
		oracle      string
		mockEnabled bool
		err         string
	}{
		{
			name: "failure - should handle nil request",
			req:  nil,
			err:  "rpc error: code = InvalidArgument desc = invalid request",
		},
		{
			name: "failure - should handle invalid query data",
			req: &types.QueryOracleRequest{
				Query: []byte("abc"),
			},
			err: "rpc error: code = InvalidArgument desc = invalid query data",
		},
		{
			name: "failure - should handle unset oracle",
			req: &types.QueryOracleRequest{
				Query: []byte("{}"),
			},
			err: "missing oracle address",
		},
		{
			name: "success - should handle error from contract",
			req: &types.QueryOracleRequest{
				Query: []byte("{}"),
			},
			err: "missing oracle address",
		},
		{
			name: "failure - should handle error from contract",
			req: &types.QueryOracleRequest{
				Query: []byte("{}"),
			},
			oracle: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			err:    "contract: not found",
		},
		{
			name: "success - should handle response from contract",
			req: &types.QueryOracleRequest{
				Query: []byte("{}"),
			},
			oracle: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			expected: &types.QueryOracleResponse{
				Data: []byte("{}"),
			},
			mockEnabled: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.mockEnabled {
				s.app.OracleKeeper = s.app.OracleKeeper.WithWasmQueryServer(keeper.MockWasmServer{})
			}

			if len(tc.oracle) > 0 {
				s.app.OracleKeeper.SetOracle(s.ctx, sdk.MustAccAddressFromBech32(tc.oracle))
			}

			resp, err := s.app.OracleKeeper.Oracle(s.ctx, tc.req)
			if len(tc.err) > 0 {
				s.Assert().EqualError(err, tc.err, "should return the correct error")
				s.Assert().Nil(resp, "response should be nil")
			} else {
				s.Assert().NoError(err, "should not return an error")
				s.Assert().Equal(tc.expected, resp, "should return the correct response")
			}
		})
	}
}
