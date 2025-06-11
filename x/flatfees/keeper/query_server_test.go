package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/flatfees/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"

	. "github.com/provenance-io/provenance/x/flatfees/keeper"
)

func TestQueryServerTestSuite(t *testing.T) {
	suite.Run(t, new(QueryServerTestSuite))
}

type QueryServerTestSuite struct {
	suite.Suite

	app         *simapp.App
	ctx         sdk.Context
	cfg         network.Config
	queryClient types.QueryClient

	params types.Params

	privkey1  cryptotypes.PrivKey
	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress
	acct1     sdk.AccountI

	privkey2  cryptotypes.PrivKey
	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress
	acct2     sdk.AccountI

	minGasPrice       sdk.Coin
	usdConversionRate uint64
}

func (s *QueryServerTestSuite) SetupTest() {
	s.app = simapp.SetupQuerier(s.T())
	s.ctx = s.app.BaseApp.NewContext(true)
	err := s.app.AccountKeeper.Params.Set(s.ctx, authtypes.DefaultParams())
	s.Require().NoError(err, "AccountKeeper.Params.Set(DefaultParams())")
	err = s.app.BankKeeper.SetParams(s.ctx, banktypes.DefaultParams())
	s.Require().NoError(err, "BankKeeper.SetParams(DefaultParams())")
	s.cfg = testutil.DefaultTestNetworkConfig()
	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, NewQueryServer(s.app.FlatFeesKeeper))
	s.queryClient = types.NewQueryClient(queryHelper)

	s.minGasPrice = sdk.Coin{
		Denom:  s.cfg.BondDenom,
		Amount: sdkmath.NewInt(10),
	}
	s.usdConversionRate = 7
	s.params = types.DefaultParams()
	s.params.ConversionFactor.ConvertedAmount.Denom = s.cfg.BondDenom
	err = s.app.FlatFeesKeeper.SetParams(s.ctx, s.params)
	s.Require().NoError(err, "FlatFeesKeeper.SetParams(DefaultParams())")

	s.privkey1 = secp256k1.GenPrivKey()
	s.pubkey1 = s.privkey1.PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	s.privkey2 = secp256k1.GenPrivKey()
	s.pubkey2 = s.privkey2.PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	s.acct1 = s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr)
	s.app.AccountKeeper.SetAccount(s.ctx, s.acct1)
	s.acct2 = s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user2Addr)
	s.app.AccountKeeper.SetAccount(s.ctx, s.acct2)
	s.Require().NoError(s.app.MarkerKeeper.AddMarkerAccount(s.ctx, markertypes.NewEmptyMarkerAccount("navcoin", s.acct1.GetAddress().String(), []markertypes.AccessGrant{})))
	s.Require().NoError(banktestutil.FundAccount(s.ctx, s.app.BankKeeper, s.acct1.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 100_000))))
}

func (s *QueryServerTestSuite) costCoin(amount int64) sdk.Coin {
	return sdk.NewInt64Coin(s.params.DefaultCost.Denom, amount)
}

func (s *QueryServerTestSuite) convertMsgFees(msgFees []*types.MsgFee) []*types.MsgFee {
	if msgFees == nil {
		return nil
	}
	rv := make([]*types.MsgFee, len(msgFees))
	for i, msgFee := range msgFees {
		rv[i] = s.params.ConversionFactor.ConvertMsgFee(msgFee)
	}
	return rv
}

func (s *QueryServerTestSuite) convertMsgFee(msgFee *types.MsgFee) *types.MsgFee {
	return s.params.ConversionFactor.ConvertMsgFee(msgFee)
}

// assertEqualPagination asserts that two provided page responses are equal, returning true iff equal.
func (s *QueryServerTestSuite) assertEqualPagination(expected, actual *query.PageResponse) bool {
	s.T().Helper()
	if expected != nil && actual != nil {
		ok := s.Assert().Equal(expected.NextKey, actual.NextKey, "Pagination.NextKey")
		ok = s.Assert().Equal(fmt.Sprintf("%d", expected.Total), fmt.Sprintf("%d", actual.Total), "Pagination.Total") && ok
		if !ok {
			return false
		}
	}
	return s.Assert().Equal(expected, actual, "Pagination")
}

func (s *QueryServerTestSuite) TestParams() {
	defaultParams := types.DefaultParams()
	tests := []struct {
		name    string
		params  *types.Params
		req     *types.QueryParamsRequest
		expResp *types.QueryParamsResponse
		expErr  string
	}{
		{
			name:    "no req",
			req:     nil,
			expResp: &types.QueryParamsResponse{Params: s.params},
		},
		{
			name:    "with req",
			req:     &types.QueryParamsRequest{},
			expResp: &types.QueryParamsResponse{Params: s.params},
		},
		{
			name: "all different",
			params: &types.Params{
				DefaultCost: sdk.NewInt64Coin("banana", 10),
				ConversionFactor: types.ConversionFactor{
					DefinitionAmount: sdk.NewInt64Coin("apple", 44),
					ConvertedAmount:  sdk.NewInt64Coin("orange", 78),
				},
			},
		},
		{
			name: "no conversion",
			params: &types.Params{
				DefaultCost: sdk.NewInt64Coin("banana", 10),
				ConversionFactor: types.ConversionFactor{
					DefinitionAmount: sdk.NewInt64Coin("banana", 1),
					ConvertedAmount:  sdk.NewInt64Coin("banana", 1),
				},
			},
		},
		{
			name: "zero default cost",
			params: &types.Params{
				DefaultCost:      sdk.NewInt64Coin(defaultParams.DefaultCost.Denom, 0),
				ConversionFactor: defaultParams.ConversionFactor,
			},
		},
		{
			name:   "defaults",
			params: &defaultParams,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if len(tc.expErr) == 0 && tc.expResp == nil && tc.params != nil {
				tc.expResp = &types.QueryParamsResponse{Params: *tc.params}
			}

			if tc.params != nil {
				err := s.app.FlatFeesKeeper.SetParams(s.ctx, *tc.params)
				s.Require().NoError(err, "SetParams(%s)", tc.params)
			}

			var actResp *types.QueryParamsResponse
			var err error
			testFunc := func() {
				actResp, err = s.queryClient.Params(s.ctx, tc.req)
			}
			s.Require().NotPanics(testFunc, "Params(...)")
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "Params(...) error")

			ok := true
			if tc.expResp != nil && actResp != nil {
				ok = assertEqualParams(s.T(), tc.expResp.Params, actResp.Params)
			}
			if ok {
				s.Assert().Equal(tc.expResp, actResp, "Params(...) response")
			}
		})
	}
}

func (s *QueryServerTestSuite) TestAllMsgFees() {
	msgFees := []*types.MsgFee{
		types.NewMsgFee("/cosmos.gov.v1.MsgSubmitProposal"),
		types.NewMsgFee("/cosmos.group.v1.MsgCreateGroup", s.costCoin(3)),
		types.NewMsgFee("/cosmos.group.v1.MsgCreateGroupPolicy", s.costCoin(4)),
		types.NewMsgFee("/cosmos.group.v1.MsgCreateGroupWithPolicy", s.costCoin(5)),
		types.NewMsgFee("/cosmos.group.v1.MsgExec", s.costCoin(6)),
		types.NewMsgFee("/cosmos.group.v1.MsgLeaveGroup", s.costCoin(7)),
		types.NewMsgFee("/cosmos.group.v1.MsgSubmitProposal", s.costCoin(8)),
		types.NewMsgFee("/cosmos.group.v1.MsgUpdateGroupAdmin", s.costCoin(9)),
		types.NewMsgFee("/cosmos.group.v1.MsgUpdateGroupMembers", s.costCoin(10)),
		types.NewMsgFee("/cosmos.group.v1.MsgUpdateGroupMetadata", s.costCoin(11)),
		types.NewMsgFee("/cosmos.group.v1.MsgUpdateGroupPolicyAdmin", s.costCoin(12)),
		types.NewMsgFee("/cosmos.group.v1.MsgUpdateGroupPolicyDecisionPolicy", s.costCoin(13)),
		types.NewMsgFee("/cosmos.group.v1.MsgUpdateGroupPolicyMetadata", s.costCoin(14)),
		types.NewMsgFee("/cosmos.group.v1.MsgVote", s.costCoin(15)),
		types.NewMsgFee("/cosmos.group.v1.MsgWithdrawProposal", s.costCoin(16)),
	}
	for i, msgFee := range msgFees {
		err := s.app.FlatFeesKeeper.SetMsgFee(s.ctx, *msgFee)
		s.Require().NoError(err, "[%d]: SetMsgFee(%s)", i, msgFee)
	}

	nextKeyFor := func(i int) []byte {
		return []byte(msgFees[i].MsgTypeUrl)
	}

	tests := []struct {
		name    string
		req     *types.QueryAllMsgFeesRequest
		expResp *types.QueryAllMsgFeesResponse
		expErr  string
	}{
		{
			name: "nil req",
			req:  nil,
			expResp: &types.QueryAllMsgFeesResponse{
				MsgFees:    s.convertMsgFees(msgFees),
				Pagination: &query.PageResponse{Total: 15},
			},
		},
		{
			name: "nil pagination: converted",
			req:  &types.QueryAllMsgFeesRequest{DoNotConvert: false, Pagination: nil},
			expResp: &types.QueryAllMsgFeesResponse{
				MsgFees:    s.convertMsgFees(msgFees),
				Pagination: &query.PageResponse{Total: 15},
			},
		},
		{
			name: "nil pagination: not converted",
			req:  &types.QueryAllMsgFeesRequest{DoNotConvert: true, Pagination: nil},
			expResp: &types.QueryAllMsgFeesResponse{
				MsgFees:    msgFees,
				Pagination: &query.PageResponse{Total: 15},
			},
		},
		{
			name: "empty pagination: converted",
			req:  &types.QueryAllMsgFeesRequest{DoNotConvert: false, Pagination: &query.PageRequest{}},
			expResp: &types.QueryAllMsgFeesResponse{
				MsgFees:    s.convertMsgFees(msgFees),
				Pagination: &query.PageResponse{Total: 15},
			},
		},
		{
			name: "empty pagination: not converted",
			req:  &types.QueryAllMsgFeesRequest{DoNotConvert: true, Pagination: &query.PageRequest{}},
			expResp: &types.QueryAllMsgFeesResponse{
				MsgFees:    msgFees,
				Pagination: &query.PageResponse{Total: 15},
			},
		},
		{
			name: "limit 1 with count",
			req:  &types.QueryAllMsgFeesRequest{Pagination: &query.PageRequest{Limit: 1, CountTotal: true}},
			expResp: &types.QueryAllMsgFeesResponse{
				MsgFees:    s.convertMsgFees(msgFees[0:1]),
				Pagination: &query.PageResponse{NextKey: nextKeyFor(1), Total: uint64(len(msgFees))},
			},
		},
		{
			name: "limit 3 with next key",
			req:  &types.QueryAllMsgFeesRequest{Pagination: &query.PageRequest{Limit: 3, Key: nextKeyFor(5)}},
			expResp: &types.QueryAllMsgFeesResponse{
				MsgFees:    s.convertMsgFees(msgFees[5:8]),
				Pagination: &query.PageResponse{NextKey: nextKeyFor(8)},
			},
		},
		{
			name: "limit 3 with next key and no conversion",
			req: &types.QueryAllMsgFeesRequest{
				DoNotConvert: true,
				Pagination:   &query.PageRequest{Limit: 3, Key: nextKeyFor(5)},
			},
			expResp: &types.QueryAllMsgFeesResponse{
				MsgFees:    msgFees[5:8],
				Pagination: &query.PageResponse{NextKey: nextKeyFor(8)},
			},
		},
		{
			name: "limit 3 with offset",
			req:  &types.QueryAllMsgFeesRequest{Pagination: &query.PageRequest{Limit: 3, Offset: 1}},
			expResp: &types.QueryAllMsgFeesResponse{
				MsgFees:    s.convertMsgFees(msgFees[1:4]),
				Pagination: &query.PageResponse{NextKey: nextKeyFor(4)},
			},
		},
		{
			name: "limit 4 reversed",
			req:  &types.QueryAllMsgFeesRequest{Pagination: &query.PageRequest{Limit: 4, Reverse: true}},
			expResp: &types.QueryAllMsgFeesResponse{
				MsgFees:    s.convertMsgFees(reversed(msgFees)[0:4]),
				Pagination: &query.PageResponse{NextKey: nextKeyFor(len(msgFees) - 5)},
			},
		},
		{
			name: "limit 5 reversed with next key and no conversion",
			req: &types.QueryAllMsgFeesRequest{
				DoNotConvert: true,
				Pagination:   &query.PageRequest{Limit: 5, Reverse: true, Key: nextKeyFor(6)},
			},
			expResp: &types.QueryAllMsgFeesResponse{
				MsgFees:    reversed(msgFees[2:7]),
				Pagination: &query.PageResponse{NextKey: nextKeyFor(1)},
			},
		},
		{
			name:   "invalid pagination",
			req:    &types.QueryAllMsgFeesRequest{Pagination: &query.PageRequest{Limit: 3, Offset: 1, Key: nextKeyFor(3)}},
			expErr: "rpc error: code = Internal desc = invalid request, either offset or key is expected, got both",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var actResp *types.QueryAllMsgFeesResponse
			var err error
			testFunc := func() {
				actResp, err = s.queryClient.AllMsgFees(s.ctx, tc.req)
			}
			s.Require().NotPanics(testFunc, "AllMsgFees(...)")
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "AllMsgFees(...) error")

			ok := true
			if tc.expResp != nil && actResp != nil {
				ok = assertEqualMsgFees(s.T(), tc.expResp.MsgFees, actResp.MsgFees) && ok
				ok = s.assertEqualPagination(tc.expResp.Pagination, actResp.Pagination) && ok
			}
			if ok {
				s.Assert().Equal(tc.expResp, actResp)
			}
		})
	}
}

func (s *QueryServerTestSuite) TestMsgFee() {
	defaultMsgFee := types.NewMsgFee("/cosmos.group.v1.MsgExec", s.params.DefaultCost)
	freeMsgFee := types.NewMsgFee("/cosmos.gov.v1.MsgSubmitProposal")
	nonDefaultMsgFee := types.NewMsgFee("/cosmos.group.v1.MsgCreateGroup", s.costCoin(3))
	otherDenomMsgFee := types.NewMsgFee("/cosmos.group.v1.MsgLeaveGroup", sdk.NewInt64Coin("banana", 7))
	msgFees := []*types.MsgFee{freeMsgFee, nonDefaultMsgFee, otherDenomMsgFee}
	for i, msgFee := range msgFees {
		err := s.app.FlatFeesKeeper.SetMsgFee(s.ctx, *msgFee)
		s.Require().NoError(err, "[%d]: SetMsgFee(%s)", i, msgFee)
	}

	tests := []struct {
		name    string
		req     *types.QueryMsgFeeRequest
		expResp *types.QueryMsgFeeResponse
		expErr  string
	}{
		{
			name: "nil req",
			req:  nil,
			// Note: The QueryClient never provides a nil req to the endpoint, so this ends up without a URL instead.
			expErr: "rpc error: code = InvalidArgument desc = unknown msg type url \"\"",
		},
		{
			name:   "empty msg type url",
			req:    &types.QueryMsgFeeRequest{MsgTypeUrl: ""},
			expErr: "rpc error: code = InvalidArgument desc = unknown msg type url \"\"",
		},
		{
			name:   "unknown msg type url",
			req:    &types.QueryMsgFeeRequest{MsgTypeUrl: "unknown"},
			expErr: "rpc error: code = InvalidArgument desc = unknown msg type url \"unknown\"",
		},
		{
			name:    "default, with conversion",
			req:     &types.QueryMsgFeeRequest{MsgTypeUrl: defaultMsgFee.MsgTypeUrl},
			expResp: &types.QueryMsgFeeResponse{MsgFee: s.convertMsgFee(defaultMsgFee)},
		},
		{
			name:    "default, no conversion",
			req:     &types.QueryMsgFeeRequest{MsgTypeUrl: defaultMsgFee.MsgTypeUrl, DoNotConvert: true},
			expResp: &types.QueryMsgFeeResponse{MsgFee: defaultMsgFee},
		},
		{
			name:    "free, with conversion",
			req:     &types.QueryMsgFeeRequest{MsgTypeUrl: freeMsgFee.MsgTypeUrl},
			expResp: &types.QueryMsgFeeResponse{MsgFee: s.convertMsgFee(freeMsgFee)},
		},
		{
			name:    "free, no conversion",
			req:     &types.QueryMsgFeeRequest{MsgTypeUrl: freeMsgFee.MsgTypeUrl, DoNotConvert: true},
			expResp: &types.QueryMsgFeeResponse{MsgFee: freeMsgFee},
		},
		{
			name:    "non-default, with conversion",
			req:     &types.QueryMsgFeeRequest{MsgTypeUrl: nonDefaultMsgFee.MsgTypeUrl},
			expResp: &types.QueryMsgFeeResponse{MsgFee: s.convertMsgFee(nonDefaultMsgFee)},
		},
		{
			name:    "non-default, no conversion",
			req:     &types.QueryMsgFeeRequest{MsgTypeUrl: nonDefaultMsgFee.MsgTypeUrl, DoNotConvert: true},
			expResp: &types.QueryMsgFeeResponse{MsgFee: nonDefaultMsgFee},
		},
		{
			name:    "other denom, with conversion",
			req:     &types.QueryMsgFeeRequest{MsgTypeUrl: otherDenomMsgFee.MsgTypeUrl},
			expResp: &types.QueryMsgFeeResponse{MsgFee: otherDenomMsgFee},
		},
		{
			name:    "other denom, no conversion",
			req:     &types.QueryMsgFeeRequest{MsgTypeUrl: otherDenomMsgFee.MsgTypeUrl, DoNotConvert: true},
			expResp: &types.QueryMsgFeeResponse{MsgFee: otherDenomMsgFee},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var actResp *types.QueryMsgFeeResponse
			var err error
			testFunc := func() {
				actResp, err = s.queryClient.MsgFee(s.ctx, tc.req)
			}
			s.Require().NotPanics(testFunc, "MsgFee(...)")
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "MsgFee(...) error")

			ok := true
			if tc.expResp != nil && actResp != nil {
				ok = assertEqualMsgFee(s.T(), tc.expResp.MsgFee, actResp.MsgFee)
			}
			if ok {
				s.Assert().Equal(tc.expResp, actResp, "MsgFee(...) response")
			}
		})
	}
}
