package keeper_test

import (
	"slices"

	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/nav"
)

// assertEqualPageResponses is a wrapper on assertions.AssertEqualPageResponses.
func (s *TestSuite) assertEqualPageResponses(exp, act *query.PageResponse, msgAndArgs ...interface{}) bool {
	s.T().Helper()
	return assertions.AssertEqualPageResponses(s.T(), exp, act, msgAndArgs...)
}

func (s *TestSuite) TestKeeper_GetNAV() {
	navs := nav.NAVRecords{
		s.newNAVRec("1white", "2purple", 1, "one"),
		s.newNAVRec("1white", "4blue", 1, "one"),
		s.newNAVRec("2white", "6green", 1, "one"),
		s.newNAVRec("2white", "8yellow", 2, "two"),
		s.newNAVRec("3white", "10orange", 3, "three"),
		s.newNAVRec("3white", "12red", 3, "three"),
		s.newNAVRec("8purple", "2gray", 4, "four"),
		s.newNAVRec("8blue", "4gray", 5, "five"),
		s.newNAVRec("8green", "6gray", 5, "five"),
		s.newNAVRec("8yellow", "8gray", 5, "five"),
		s.newNAVRec("8orange", "10gray", 5, "five"),
		s.newNAVRec("8red", "12gray", 5, "five"),
		s.newNAVRec("10brown", "13black", 6, "six"),
		s.newNAVRec("10black", "9brown", 6, "six"),
		s.newNAVRec("10black", "4green", 6, "six"),
		s.newNAVRec("12indigo", "5gray", 7, "seven"),
	}

	ctx, _ := s.ctx.CacheContext()
	s.storeNAVs(ctx, navs)

	newReq := func(asset, price string) *nav.QueryGetNAVRequest {
		return &nav.QueryGetNAVRequest{AssetDenom: asset, PriceDenom: price}
	}

	tests := []struct {
		name    string
		req     *nav.QueryGetNAVRequest
		expResp *nav.QueryGetNAVResponse
		expErr  string
	}{
		{
			name:   "nil req",
			req:    nil,
			expErr: "rpc error: code = InvalidArgument desc = empty request",
		},
		{
			name:   "no asset denom",
			req:    newReq("", "yellow"),
			expErr: "rpc error: code = InvalidArgument desc = empty asset denom",
		},
		{
			name:   "no price denom",
			req:    newReq("pink", ""),
			expErr: "rpc error: code = InvalidArgument desc = empty price denom",
		},
		{
			name:    "no such nav",
			req:     newReq("gold", "pink"),
			expResp: &nav.QueryGetNAVResponse{Nav: nil},
		},
		{
			name:    "known asset, unknown price",
			req:     newReq("white", "pink"),
			expResp: &nav.QueryGetNAVResponse{Nav: nil},
		},
		{
			name:    "unknown asset, known price",
			req:     newReq("pink", "gray"),
			expResp: &nav.QueryGetNAVResponse{Nav: nil},
		},
		{
			name:    "asset with only one price: right price denom",
			req:     newReq("indigo", "gray"),
			expResp: &nav.QueryGetNAVResponse{Nav: s.newNAVRec("12indigo", "5gray", 7, "seven")},
		},
		{
			name:    "asset with only one price: wrong price denom",
			req:     newReq("indigo", "yellow"),
			expResp: &nav.QueryGetNAVResponse{Nav: nil},
		},
		{
			name:    "known entry with alternate capitalization",
			req:     newReq("indigo", "grAy"),
			expResp: &nav.QueryGetNAVResponse{Nav: nil},
		},
		{
			name:    "asset with many prices: right price denom",
			req:     newReq("white", "red"),
			expResp: &nav.QueryGetNAVResponse{Nav: s.newNAVRec("3white", "12red", 3, "three")},
		},
		{
			name:    "asset with many prices: wrong price denom",
			req:     newReq("white", "silver"),
			expResp: &nav.QueryGetNAVResponse{Nav: nil},
		},
		{
			name:    "asset is prefix only",
			req:     newReq("whit", "red"),
			expResp: &nav.QueryGetNAVResponse{Nav: nil},
		},
		{
			name:    "asset has known prefix",
			req:     newReq("whites", "red"),
			expResp: &nav.QueryGetNAVResponse{Nav: nil},
		},
		{
			name:    "price is prefix only",
			req:     newReq("white", "re"),
			expResp: &nav.QueryGetNAVResponse{Nav: nil},
		},
		{
			name:    "price has known prefix",
			req:     newReq("white", "reds"),
			expResp: &nav.QueryGetNAVResponse{Nav: nil},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			qsrvr := s.getQueryServer()
			var resp *nav.QueryGetNAVResponse
			var err error
			testFunc := func() {
				resp, err = qsrvr.GetNAV(ctx, tc.req)
			}
			s.Require().NotPanics(testFunc, "GetNAV")
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "GetNAV error")
			s.Assert().Equal(tc.expResp, resp, "GetNAV response")
		})
	}
}

func (s *TestSuite) TestKeeper_GetAllNAVs() {
	navs := nav.NAVRecords{
		s.newNAVRec("1white", "2purple", 1, "one"),
		s.newNAVRec("1white", "4blue", 1, "one"),
		s.newNAVRec("2white", "6green", 1, "one"),
		s.newNAVRec("2white", "8yellow", 2, "two"),
		s.newNAVRec("3white", "10orange", 3, "three"),
		s.newNAVRec("3white", "12red", 3, "three"),
		s.newNAVRec("8purple", "2gray", 4, "four"),
		s.newNAVRec("8blue", "4gray", 5, "five"),
		s.newNAVRec("8green", "6gray", 5, "five"),
		s.newNAVRec("8yellow", "8gray", 5, "five"),
		s.newNAVRec("8orange", "10gray", 5, "five"),
		s.newNAVRec("8red", "12gray", 5, "five"),
		s.newNAVRec("10brown", "13black", 6, "six"),
		s.newNAVRec("10black", "9brown", 6, "six"),
		s.newNAVRec("10black", "4green", 6, "six"),
		s.newNAVRec("12indigo", "5gray", 7, "seven"),
	}
	sortedNAVs := make(nav.NAVRecords, len(navs))
	copy(sortedNAVs, navs)
	slices.SortFunc(sortedNAVs, compareNAVRecords)

	tests := []struct {
		name    string
		ini     nav.NAVRecords // if defined as nil, will be all the navs.
		req     *nav.QueryGetAllNAVsRequest
		expResp *nav.QueryGetAllNAVsResponse
		expErr  string
	}{
		{
			name:    "no navs in state",
			ini:     nav.NAVRecords{},
			req:     nil,
			expResp: &nav.QueryGetAllNAVsResponse{},
		},
		{
			name: "only one nav in state: get all",
			ini:  nav.NAVRecords{s.newNAVRec("12gold", "77silver", 83, "one")},
			req:  nil,
			expResp: &nav.QueryGetAllNAVsResponse{
				Navs:       nav.NAVRecords{s.newNAVRec("12gold", "77silver", 83, "one")},
				Pagination: &query.PageResponse{Total: 1},
			},
		},
		{
			name: "only one nav in state: get by asset denom",
			ini:  nav.NAVRecords{s.newNAVRec("12gold", "77silver", 83, "one")},
			req:  &nav.QueryGetAllNAVsRequest{AssetDenom: "gold"},
			expResp: &nav.QueryGetAllNAVsResponse{
				Navs:       nav.NAVRecords{s.newNAVRec("12gold", "77silver", 83, "one")},
				Pagination: &query.PageResponse{Total: 1},
			},
		},
		{
			name:    "only one nav in state: get other denom",
			ini:     nav.NAVRecords{s.newNAVRec("12gold", "77silver", 83, "one")},
			req:     &nav.QueryGetAllNAVsRequest{AssetDenom: "silver"},
			expResp: &nav.QueryGetAllNAVsResponse{},
		},
		{
			name:    "many navs: unknown req denom",
			req:     &nav.QueryGetAllNAVsRequest{AssetDenom: "pink"},
			expResp: &nav.QueryGetAllNAVsResponse{},
		},
		{
			name: "many navs: known req denom with one result",
			req:  &nav.QueryGetAllNAVsRequest{AssetDenom: "indigo"},
			expResp: &nav.QueryGetAllNAVsResponse{
				Navs:       nav.NAVRecords{s.newNAVRec("12indigo", "5gray", 7, "seven")},
				Pagination: &query.PageResponse{Total: 1},
			},
		},
		{
			name: "many navs: known req denom with two results",
			req:  &nav.QueryGetAllNAVsRequest{AssetDenom: "black"},
			expResp: &nav.QueryGetAllNAVsResponse{
				Navs: nav.NAVRecords{
					s.newNAVRec("10black", "9brown", 6, "six"),
					s.newNAVRec("10black", "4green", 6, "six"),
				},
				Pagination: &query.PageResponse{Total: 2},
			},
		},
		{
			name: "many navs: known req denom with several results: no pagination",
			req:  &nav.QueryGetAllNAVsRequest{AssetDenom: "white"},
			expResp: &nav.QueryGetAllNAVsResponse{
				Navs: nav.NAVRecords{
					s.newNAVRec("1white", "4blue", 1, "one"),
					s.newNAVRec("2white", "6green", 1, "one"),
					s.newNAVRec("3white", "10orange", 3, "three"),
					s.newNAVRec("1white", "2purple", 1, "one"),
					s.newNAVRec("3white", "12red", 3, "three"),
					s.newNAVRec("2white", "8yellow", 2, "two"),
				},
				Pagination: &query.PageResponse{Total: 6},
			},
		},
		{
			name: "many navs: known req denom with several results: paginated",
			req: &nav.QueryGetAllNAVsRequest{
				AssetDenom: "white",
				Pagination: &query.PageRequest{Offset: 1, Limit: 3},
			},
			expResp: &nav.QueryGetAllNAVsResponse{
				Navs: nav.NAVRecords{
					s.newNAVRec("2white", "6green", 1, "one"),
					s.newNAVRec("3white", "10orange", 3, "three"),
					s.newNAVRec("1white", "2purple", 1, "one"),
				},
				Pagination: &query.PageResponse{NextKey: []byte("red")},
			},
		},
		{
			name: "many navs: get all: no pagination",
			req:  nil,
			expResp: &nav.QueryGetAllNAVsResponse{
				Navs:       sortedNAVs,
				Pagination: &query.PageResponse{Total: uint64(len(navs))},
			},
		},
		{
			name: "many navs: get all: paginated",
			req: &nav.QueryGetAllNAVsRequest{
				Pagination: &query.PageRequest{
					Key:   []byte(sortedNAVs[5].Assets.Denom + string(byte(0x0)) + sortedNAVs[5].Price.Denom),
					Limit: 10,
				},
			},
			expResp: &nav.QueryGetAllNAVsResponse{
				Navs: sortedNAVs[5:15],
				Pagination: &query.PageResponse{
					NextKey: []byte(sortedNAVs[15].Assets.Denom + string(byte(0x0)) + sortedNAVs[15].Price.Denom),
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.expResp != nil && tc.expResp.Pagination == nil {
				tc.expResp.Pagination = &query.PageResponse{}
			}

			if tc.ini == nil {
				tc.ini = navs
			}
			ctx, _ := s.ctx.CacheContext()
			s.storeNAVs(ctx, tc.ini)

			qsrvr := s.getQueryServer()
			var actResp *nav.QueryGetAllNAVsResponse
			var err error
			testFunc := func() {
				actResp, err = qsrvr.GetAllNAVs(ctx, tc.req)
			}
			s.Require().NotPanics(testFunc, "GetAllNAVs")
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "GetAllNAVs error")

			if tc.expResp == nil {
				s.Assert().Nil(actResp, "GetAllNAVs response")
				return
			}
			s.Require().NotNil(actResp, "GetAllNAVs response")
			ok := s.assertEqualNAVRecords(tc.expResp.Navs, actResp.Navs, "GetAllNAVs response")
			ok = s.assertEqualPageResponses(tc.expResp.Pagination, actResp.Pagination, "GetAllNAVs response") && ok
			if ok {
				return
			}

			s.Assert().Equal(tc.expResp, actResp, "GetAllNAVs response")
		})
	}
}
