package keeper_test

import (
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/ibcratelimit"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	queryClient ibcratelimit.QueryClient
}

func (s *TestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.ctx = s.ctx.WithBlockHeight(0)

	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	ibcratelimit.RegisterQueryServer(queryHelper, s.app.RateLimitingKeeper)
	s.queryClient = ibcratelimit.NewQueryClient(queryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestGetSetParams() {
	tests := []struct {
		name     string
		contract string
	}{
		{
			name: "success - get empty params",
		},
		{
			name:     "success - set and get new params",
			contract: "contractaddress",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			params := ibcratelimit.NewParams(tc.contract)
			s.app.RateLimitingKeeper.SetParams(s.ctx, params)
			newParams, err := s.app.RateLimitingKeeper.GetParams(s.ctx)
			s.Assert().NoError(err)
			s.Assert().Equal(params, newParams, "should have expected params")
		})
	}
}

func (s *TestSuite) TestGetContractAddress() {
	tests := []struct {
		name     string
		contract string
	}{
		{
			name: "success - get empty contract",
		},
		{
			name:     "success - set and get new contract address",
			contract: "contractaddress",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			params := ibcratelimit.NewParams(tc.contract)
			s.app.RateLimitingKeeper.SetParams(s.ctx, params)
			contract := s.app.RateLimitingKeeper.GetContractAddress(s.ctx)
			s.Assert().Equal(tc.contract, contract, "should have expected contract")
		})
	}
}

func (s *TestSuite) TestContractConfigured() {
	tests := []struct {
		name     string
		contract string
		expected bool
	}{
		{
			name:     "success - get empty contract",
			expected: false,
		},
		{
			name:     "success - set and get new contract address",
			contract: "contractaddress",
			expected: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			params := ibcratelimit.NewParams(tc.contract)
			s.app.RateLimitingKeeper.SetParams(s.ctx, params)
			configured := s.app.RateLimitingKeeper.ContractConfigured(s.ctx)
			s.Assert().Equal(tc.expected, configured, "should have expected configured output")
		})
	}
}
