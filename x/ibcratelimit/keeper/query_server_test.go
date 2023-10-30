package keeper_test

import (
	"github.com/provenance-io/provenance/x/ibcratelimit"
)

func (s *TestSuite) TestQueryParams() {
	tests := []struct {
		name     string
		contract string
		expected ibcratelimit.ParamsResponse
	}{
		{
			name: "success - params have not been set",
		},
		{
			name:     "success - params have been set",
			contract: "randomaddress",
			expected: ibcratelimit.ParamsResponse{
				Params: ibcratelimit.Params{
					ContractAddress: "randomaddress",
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if len(tc.contract) > 0 {
				s.app.RateLimitingKeeper.SetParams(s.ctx, ibcratelimit.NewParams(tc.contract))
			}

			request := ibcratelimit.ParamsRequest{}
			response, err := s.queryClient.Params(s.ctx, &request)

			s.Assert().NoError(err, "should not throw an error")
			s.Assert().Equal(tc.expected, *response, "should return correct response")

			if len(tc.contract) > 0 {
				s.app.RateLimitingKeeper.SetParams(s.ctx, ibcratelimit.DefaultParams())
			}
		})
	}
}
