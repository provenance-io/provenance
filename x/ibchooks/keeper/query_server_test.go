package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/x/ibchooks/types"
)

type QueryServerTestSuite struct {
	suite.Suite

	app         *simapp.App
	ctx         sdk.Context
	cfg         network.Config
	queryClient types.QueryClient
}

func (s *QueryServerTestSuite) SetupTest() {
	s.app = simapp.SetupQuerier(s.T())
	s.ctx = s.app.BaseApp.NewContext(true)
	s.cfg = testutil.DefaultTestNetworkConfig()
	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, s.app.IBCHooksKeeper)
	s.queryClient = types.NewQueryClient(queryHelper)
}

func TestQuerierTestSuite(t *testing.T) {
	suite.Run(t, new(QueryServerTestSuite))
}

func (s *QueryServerTestSuite) TestQueryParams() {
	params := types.DefaultParams()
	s.app.IBCHooksKeeper.SetParams(s.ctx, params)

	response, err := s.queryClient.Params(s.ctx.Context(), &types.QueryParamsRequest{})
	s.Require().NoError(err, "QueryParamsRequest expected no error, got %v", err)
	s.Require().NotNil(response, "QueryParamsRequest expected non-nil response, got nil")
	s.Require().Len(response.Params.AllowedAsyncAckContracts, 0, "QueryParamsRequest expected no allowed async ack contracts, got %v", response.Params.AllowedAsyncAckContracts)

	params = types.Params{
		AllowedAsyncAckContracts: []string{"cosmos1address1", "cosmos1address2"},
	}
	s.app.IBCHooksKeeper.SetParams(s.ctx, params)

	response, err = s.queryClient.Params(s.ctx.Context(), &types.QueryParamsRequest{})
	s.Require().NoError(err, "QueryParamsRequest expected no error, got %v", err)
	s.Require().NotNil(response, "QueryParamsRequest expected non-nil response, got nil")
	s.Require().Len(response.Params.AllowedAsyncAckContracts, 2, "QueryParamsRequest expected 2 allowed async ack contracts, got %v", response.Params.AllowedAsyncAckContracts)
	s.Require().Equal(params.AllowedAsyncAckContracts, response.Params.AllowedAsyncAckContracts, "QueryParamsRequest expected allowed async ack contracts %v, got %v", params.AllowedAsyncAckContracts, response.Params.AllowedAsyncAckContracts)
}
