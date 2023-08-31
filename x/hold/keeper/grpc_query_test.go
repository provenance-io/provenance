package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/hold"
	"github.com/provenance-io/provenance/x/hold/keeper"
)

func (s *TestSuite) TestKeeper_GetHolds() {
	store := s.getStore()
	s.requireSetHoldCoinAmount(store, s.addr2, "banana", s.int(144))
	s.requireSetHoldCoinAmount(store, s.addr3, "banana", s.int(89))
	s.requireSetHoldCoinAmount(store, s.addr3, "cactus", s.int(55))
	s.requireSetHoldCoinAmount(store, s.addr3, "date", s.int(34))
	s.setHoldCoinAmountRaw(store, s.addr4, "dratcoin", "dratvalue")
	store = nil

	req := func(addr string) *hold.GetHoldsRequest {
		return &hold.GetHoldsRequest{Address: addr}
	}
	resp := func(amount string) *hold.GetHoldsResponse {
		return &hold.GetHoldsResponse{Amount: s.coins(amount)}
	}

	tests := []struct {
		name    string
		request *hold.GetHoldsRequest
		expResp *hold.GetHoldsResponse
		expErr  []string
	}{
		{
			name:    "nil request",
			request: nil,
			expErr:  []string{"InvalidArgument", "empty request"},
		},
		{
			name:    "empty addr",
			request: req(""),
			expErr:  []string{"InvalidArgument", "address cannot be empty"},
		},
		{
			name:    "invalid addr",
			request: req("not-valid"),
			expErr:  []string{"InvalidArgument", "invalid address", "decoding bech32 failed"},
		},
		{
			name:    "nothing on hold",
			request: req(s.addr1.String()),
			expResp: resp(""),
		},
		{
			name:    "one denom on hold",
			request: req(s.addr2.String()),
			expResp: resp("144banana"),
		},
		{
			name:    "three denoms on hold",
			request: req(s.addr3.String()),
			expResp: resp("89banana,55cactus,34date"),
		},
		{
			name:    "error getting amount",
			request: req(s.addr4.String()),
			expErr: []string{
				s.addr4.String(), "failed to read amount of dratcoin",
				"math/big: cannot unmarshal \"dratvalue\" into a *big.Int",
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var response *hold.GetHoldsResponse
			var err error
			testFunc := func() {
				response, err = s.keeper.GetHolds(s.stdlibCtx, tc.request)
			}
			s.Require().NotPanics(testFunc, "GetHolds")
			s.assertErrorContents(err, tc.expErr, "GetHolds error")
			s.Assert().Equal(tc.expResp, response, "GetHolds response")
		})
	}
}

func (s *TestSuite) TestKeeper_GetAllHolds() {
	accHold := func(addr sdk.AccAddress, amount string) *hold.AccountHold {
		return &hold.AccountHold{
			Address: addr.String(),
			Amount:  s.coins(amount),
		}
	}
	pageResp := func(total uint64, nextKey []byte) *query.PageResponse {
		return &query.PageResponse{
			Total:   total,
			NextKey: nextKey,
		}
	}
	nextKey := func(addr sdk.AccAddress, denom string) []byte {
		return keeper.CreateHoldCoinKey(addr, denom)[len(keeper.KeyPrefixHoldCoin):]
	}

	// standardSetup puts two denoms on hold for each addrs with incremental amounts.
	// This is used unless the test has a specific setup function to use instead.
	standardSetup := func(s *TestSuite, store sdk.KVStore) {
		s.requireSetHoldCoinAmount(store, s.addr1, "banana", s.int(99))
		s.requireSetHoldCoinAmount(store, s.addr1, "cherry", s.int(12))
		s.requireSetHoldCoinAmount(store, s.addr2, "banana", s.int(100))
		s.requireSetHoldCoinAmount(store, s.addr2, "cherry", s.int(13))
		s.requireSetHoldCoinAmount(store, s.addr3, "banana", s.int(101))
		s.requireSetHoldCoinAmount(store, s.addr3, "cherry", s.int(14))
		s.requireSetHoldCoinAmount(store, s.addr4, "banana", s.int(102))
		s.requireSetHoldCoinAmount(store, s.addr4, "cherry", s.int(15))
		s.requireSetHoldCoinAmount(store, s.addr5, "banana", s.int(103))
		s.requireSetHoldCoinAmount(store, s.addr5, "cherry", s.int(16))
	}
	standardExp := []*hold.AccountHold{
		accHold(s.addr1, "99banana,12cherry"),
		accHold(s.addr2, "100banana,13cherry"),
		accHold(s.addr3, "101banana,14cherry"),
		accHold(s.addr4, "102banana,15cherry"),
		accHold(s.addr5, "103banana,16cherry"),
	}
	standardExpRev := make([]*hold.AccountHold, len(standardExp))
	for i, val := range standardExp {
		standardExpRev[len(standardExp)-i-1] = val
	}

	tests := []struct {
		name        string
		setup       func(s *TestSuite, store sdk.KVStore)
		request     *hold.GetAllHoldsRequest
		expHolds    []*hold.AccountHold
		expPageResp *query.PageResponse
		expErr      []string
	}{
		{
			name:        "nil request",
			request:     nil,
			expHolds:    standardExp,
			expPageResp: pageResp(5, nil),
		},
		{
			name:        "nil pagination",
			request:     &hold.GetAllHoldsRequest{Pagination: nil},
			expHolds:    standardExp,
			expPageResp: pageResp(5, nil),
		},
		{
			name: "both offset and key",
			request: &hold.GetAllHoldsRequest{Pagination: &query.PageRequest{
				Key: nextKey(s.addr1, "banana"), Offset: 1,
			}},
			expErr: []string{"InvalidArgument", "either offset or key is expected, got both"},
		},
		{
			name: "found bad entry",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr1, "banana", s.int(99))
				s.setHoldCoinAmountRaw(store, s.addr2, "badcoin", "badvalue")
			},
			request: nil,
			expErr: []string{
				s.addr2.String(), "failed to read amount of badcoin",
				"math/big: cannot unmarshal \"badvalue\" into a *big.Int",
			},
		},
		{
			name: "found bad entry using nextkey",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr1, "banana", s.int(99))
				s.setHoldCoinAmountRaw(store, s.addr2, "badcoin", "badvalue")
			},
			request: &hold.GetAllHoldsRequest{Pagination: &query.PageRequest{
				Key:   nextKey(s.addr1, "banana"),
				Limit: 5,
			}},
			expErr: []string{
				s.addr2.String(), "failed to read amount of badcoin",
				"math/big: cannot unmarshal \"badvalue\" into a *big.Int",
			},
		},
		{
			name: "bad entry but its out of the result range",
			setup: func(s *TestSuite, store sdk.KVStore) {
				standardSetup(s, store)
				s.setHoldCoinAmountRaw(store, s.addr5, "zoinkscoin", "zoinksvalue")
			},
			request: &hold.GetAllHoldsRequest{Pagination: &query.PageRequest{
				Offset: 0, Limit: 4, CountTotal: true,
			}},
			expHolds:    standardExp[:4],
			expPageResp: pageResp(5, nextKey(s.addr5, "banana")),
		},
		{
			name: "multiple denoms per entry, count total",
			request: &hold.GetAllHoldsRequest{Pagination: &query.PageRequest{
				Offset: 1, Limit: 2, CountTotal: true,
			}},
			expHolds:    standardExp[1:3],
			expPageResp: pageResp(5, nextKey(s.addr4, "banana")),
		},
		{
			name: "multiple denoms per entry, reversed, count total",
			request: &hold.GetAllHoldsRequest{Pagination: &query.PageRequest{
				Offset: 1, Limit: 2, CountTotal: true, Reverse: true,
			}},
			expHolds:    standardExpRev[1:3],
			expPageResp: pageResp(5, nextKey(s.addr2, "cherry")),
		},
		{
			name: "with offset, partial results",
			request: &hold.GetAllHoldsRequest{Pagination: &query.PageRequest{
				Offset: 1, Limit: 2,
			}},
			expHolds:    standardExp[1:3],
			expPageResp: pageResp(0, nextKey(s.addr4, "banana")),
		},
		{
			name: "with offset, reversed, partial results",
			request: &hold.GetAllHoldsRequest{Pagination: &query.PageRequest{
				Offset: 1, Limit: 2, Reverse: true,
			}},
			expHolds:    standardExpRev[1:3],
			expPageResp: pageResp(0, nextKey(s.addr2, "cherry")),
		},
		{
			name: "with offset, all results",
			request: &hold.GetAllHoldsRequest{Pagination: &query.PageRequest{
				Offset: 1, Limit: 4,
			}},
			expHolds:    standardExp[1:],
			expPageResp: pageResp(0, nil),
		},
		{
			name: "with offset, reversed, all results",
			request: &hold.GetAllHoldsRequest{Pagination: &query.PageRequest{
				Offset: 1, Reverse: true, Limit: 100,
			}},
			expHolds:    standardExpRev[1:],
			expPageResp: pageResp(0, nil),
		},
		{
			name: "with key, partial results",
			request: &hold.GetAllHoldsRequest{Pagination: &query.PageRequest{
				Key: nextKey(s.addr2, "banana"), Limit: 2,
			}},
			expHolds:    standardExp[1:3],
			expPageResp: pageResp(0, nextKey(s.addr4, "banana")),
			expErr:      nil,
		},
		{
			name: "with key, reversed, partial results",
			request: &hold.GetAllHoldsRequest{Pagination: &query.PageRequest{
				Key: nextKey(s.addr4, "cherry"), Limit: 2, Reverse: true,
			}},
			expHolds:    standardExpRev[1:3],
			expPageResp: pageResp(0, nextKey(s.addr2, "cherry")),
		},
		{
			name: "with key, all results",
			request: &hold.GetAllHoldsRequest{Pagination: &query.PageRequest{
				Key: nextKey(s.addr2, "banana"), Limit: 4,
			}},
			expHolds:    standardExp[1:],
			expPageResp: pageResp(0, nil),
			expErr:      nil,
		},
		{
			name: "with key, reversed, all results",
			request: &hold.GetAllHoldsRequest{Pagination: &query.PageRequest{
				Key: nextKey(s.addr4, "cherry"), Limit: 4, Reverse: true,
			}},
			expHolds:    standardExpRev[1:],
			expPageResp: pageResp(0, nil),
		},
		{
			name:        "all results",
			request:     &hold.GetAllHoldsRequest{Pagination: &query.PageRequest{}},
			expHolds:    standardExp,
			expPageResp: pageResp(5, nil),
		},
		{
			name:        "all results, reversed",
			request:     &hold.GetAllHoldsRequest{Pagination: &query.PageRequest{Reverse: true}},
			expHolds:    standardExpRev,
			expPageResp: pageResp(5, nil),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearHoldState()
			if tc.setup == nil {
				tc.setup = standardSetup
			}
			tc.setup(s, s.getStore())

			var response *hold.GetAllHoldsResponse
			var err error
			testFunc := func() {
				response, err = s.keeper.GetAllHolds(s.stdlibCtx, tc.request)
			}
			s.Require().NotPanics(testFunc, "GetAllHolds")
			s.assertErrorContents(err, tc.expErr, "GetAllHolds error")
			if response != nil {
				s.Assert().Equal(tc.expHolds, response.Holds, "response holds")
				s.Assert().Equal(int(tc.expPageResp.Total), int(response.Pagination.Total), "response pagination total")
				s.Assert().Equal(tc.expPageResp.NextKey, response.Pagination.NextKey, "response pagination next key")
			}
		})
	}
}
