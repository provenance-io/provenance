package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/hold"
	"github.com/provenance-io/provenance/x/hold/keeper"
)

func (s *TestSuite) TestKeeper_GetEscrow() {
	store := s.getStore()
	s.requireSetEscrowCoinAmount(store, s.addr2, "banana", s.int(144))
	s.requireSetEscrowCoinAmount(store, s.addr3, "banana", s.int(89))
	s.requireSetEscrowCoinAmount(store, s.addr3, "cactus", s.int(55))
	s.requireSetEscrowCoinAmount(store, s.addr3, "date", s.int(34))
	s.setEscrowCoinAmountRaw(store, s.addr4, "dratcoin", "dratvalue")
	store = nil

	req := func(addr string) *hold.GetEscrowRequest {
		return &hold.GetEscrowRequest{Address: addr}
	}
	resp := func(amount string) *hold.GetEscrowResponse {
		return &hold.GetEscrowResponse{Amount: s.coins(amount)}
	}

	tests := []struct {
		name    string
		request *hold.GetEscrowRequest
		expResp *hold.GetEscrowResponse
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
			var response *hold.GetEscrowResponse
			var err error
			testFunc := func() {
				response, err = s.keeper.GetEscrow(s.stdlibCtx, tc.request)
			}
			s.Require().NotPanics(testFunc, "GetEscrow")
			s.assertErrorContents(err, tc.expErr, "GetEscrow error")
			s.Assert().Equal(tc.expResp, response, "GetEscrow response")
		})
	}
}

func (s *TestSuite) TestKeeper_GetAllEscrow() {
	accEscrow := func(addr sdk.AccAddress, amount string) *hold.AccountEscrow {
		return &hold.AccountEscrow{
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
		return keeper.CreateEscrowCoinKey(addr, denom)[len(keeper.KeyPrefixEscrowCoin):]
	}

	// standardSetup puts two denoms on hold for each addrs with incremental amounts.
	// This is used unless the test has a specific setup function to use instead.
	standardSetup := func(s *TestSuite, store sdk.KVStore) {
		s.requireSetEscrowCoinAmount(store, s.addr1, "banana", s.int(99))
		s.requireSetEscrowCoinAmount(store, s.addr1, "cherry", s.int(12))
		s.requireSetEscrowCoinAmount(store, s.addr2, "banana", s.int(100))
		s.requireSetEscrowCoinAmount(store, s.addr2, "cherry", s.int(13))
		s.requireSetEscrowCoinAmount(store, s.addr3, "banana", s.int(101))
		s.requireSetEscrowCoinAmount(store, s.addr3, "cherry", s.int(14))
		s.requireSetEscrowCoinAmount(store, s.addr4, "banana", s.int(102))
		s.requireSetEscrowCoinAmount(store, s.addr4, "cherry", s.int(15))
		s.requireSetEscrowCoinAmount(store, s.addr5, "banana", s.int(103))
		s.requireSetEscrowCoinAmount(store, s.addr5, "cherry", s.int(16))
	}
	standardExp := []*hold.AccountEscrow{
		accEscrow(s.addr1, "99banana,12cherry"),
		accEscrow(s.addr2, "100banana,13cherry"),
		accEscrow(s.addr3, "101banana,14cherry"),
		accEscrow(s.addr4, "102banana,15cherry"),
		accEscrow(s.addr5, "103banana,16cherry"),
	}
	standardExpRev := make([]*hold.AccountEscrow, len(standardExp))
	for i, val := range standardExp {
		standardExpRev[len(standardExp)-i-1] = val
	}

	tests := []struct {
		name        string
		setup       func(s *TestSuite, store sdk.KVStore)
		request     *hold.GetAllEscrowRequest
		expEscrows  []*hold.AccountEscrow
		expPageResp *query.PageResponse
		expErr      []string
	}{
		{
			name:        "nil request",
			request:     nil,
			expEscrows:  standardExp,
			expPageResp: pageResp(5, nil),
		},
		{
			name:        "nil pagination",
			request:     &hold.GetAllEscrowRequest{Pagination: nil},
			expEscrows:  standardExp,
			expPageResp: pageResp(5, nil),
		},
		{
			name: "both offset and key",
			request: &hold.GetAllEscrowRequest{Pagination: &query.PageRequest{
				Key: nextKey(s.addr1, "banana"), Offset: 1,
			}},
			expErr: []string{"InvalidArgument", "either offset or key is expected, got both"},
		},
		{
			name: "found bad entry",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetEscrowCoinAmount(store, s.addr1, "banana", s.int(99))
				s.setEscrowCoinAmountRaw(store, s.addr2, "badcoin", "badvalue")
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
				s.requireSetEscrowCoinAmount(store, s.addr1, "banana", s.int(99))
				s.setEscrowCoinAmountRaw(store, s.addr2, "badcoin", "badvalue")
			},
			request: &hold.GetAllEscrowRequest{Pagination: &query.PageRequest{
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
				s.setEscrowCoinAmountRaw(store, s.addr5, "zoinkscoin", "zoinksvalue")
			},
			request: &hold.GetAllEscrowRequest{Pagination: &query.PageRequest{
				Offset: 0, Limit: 4, CountTotal: true,
			}},
			expEscrows:  standardExp[:4],
			expPageResp: pageResp(5, nextKey(s.addr5, "banana")),
		},
		{
			name: "multiple denoms per entry, count total",
			request: &hold.GetAllEscrowRequest{Pagination: &query.PageRequest{
				Offset: 1, Limit: 2, CountTotal: true,
			}},
			expEscrows:  standardExp[1:3],
			expPageResp: pageResp(5, nextKey(s.addr4, "banana")),
		},
		{
			name: "multiple denoms per entry, reversed, count total",
			request: &hold.GetAllEscrowRequest{Pagination: &query.PageRequest{
				Offset: 1, Limit: 2, CountTotal: true, Reverse: true,
			}},
			expEscrows:  standardExpRev[1:3],
			expPageResp: pageResp(5, nextKey(s.addr2, "cherry")),
		},
		{
			name: "with offset, partial results",
			request: &hold.GetAllEscrowRequest{Pagination: &query.PageRequest{
				Offset: 1, Limit: 2,
			}},
			expEscrows:  standardExp[1:3],
			expPageResp: pageResp(0, nextKey(s.addr4, "banana")),
		},
		{
			name: "with offset, reversed, partial results",
			request: &hold.GetAllEscrowRequest{Pagination: &query.PageRequest{
				Offset: 1, Limit: 2, Reverse: true,
			}},
			expEscrows:  standardExpRev[1:3],
			expPageResp: pageResp(0, nextKey(s.addr2, "cherry")),
		},
		{
			name: "with offset, all results",
			request: &hold.GetAllEscrowRequest{Pagination: &query.PageRequest{
				Offset: 1, Limit: 4,
			}},
			expEscrows:  standardExp[1:],
			expPageResp: pageResp(0, nil),
		},
		{
			name: "with offset, reversed, all results",
			request: &hold.GetAllEscrowRequest{Pagination: &query.PageRequest{
				Offset: 1, Reverse: true, Limit: 100,
			}},
			expEscrows:  standardExpRev[1:],
			expPageResp: pageResp(0, nil),
		},
		{
			name: "with key, partial results",
			request: &hold.GetAllEscrowRequest{Pagination: &query.PageRequest{
				Key: nextKey(s.addr2, "banana"), Limit: 2,
			}},
			expEscrows:  standardExp[1:3],
			expPageResp: pageResp(0, nextKey(s.addr4, "banana")),
			expErr:      nil,
		},
		{
			name: "with key, reversed, partial results",
			request: &hold.GetAllEscrowRequest{Pagination: &query.PageRequest{
				Key: nextKey(s.addr4, "cherry"), Limit: 2, Reverse: true,
			}},
			expEscrows:  standardExpRev[1:3],
			expPageResp: pageResp(0, nextKey(s.addr2, "cherry")),
		},
		{
			name: "with key, all results",
			request: &hold.GetAllEscrowRequest{Pagination: &query.PageRequest{
				Key: nextKey(s.addr2, "banana"), Limit: 4,
			}},
			expEscrows:  standardExp[1:],
			expPageResp: pageResp(0, nil),
			expErr:      nil,
		},
		{
			name: "with key, reversed, all results",
			request: &hold.GetAllEscrowRequest{Pagination: &query.PageRequest{
				Key: nextKey(s.addr4, "cherry"), Limit: 4, Reverse: true,
			}},
			expEscrows:  standardExpRev[1:],
			expPageResp: pageResp(0, nil),
		},
		{
			name:        "all results",
			request:     &hold.GetAllEscrowRequest{Pagination: &query.PageRequest{}},
			expEscrows:  standardExp,
			expPageResp: pageResp(5, nil),
		},
		{
			name:        "all results, reversed",
			request:     &hold.GetAllEscrowRequest{Pagination: &query.PageRequest{Reverse: true}},
			expEscrows:  standardExpRev,
			expPageResp: pageResp(5, nil),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearEscrowState()
			if tc.setup == nil {
				tc.setup = standardSetup
			}
			tc.setup(s, s.getStore())

			var response *hold.GetAllEscrowResponse
			var err error
			testFunc := func() {
				response, err = s.keeper.GetAllEscrow(s.stdlibCtx, tc.request)
			}
			s.Require().NotPanics(testFunc, "GetAllEscrow")
			s.assertErrorContents(err, tc.expErr, "GetAllEscrow error")
			if response != nil {
				s.Assert().Equal(tc.expEscrows, response.Escrows, "response escrows")
				s.Assert().Equal(int(tc.expPageResp.Total), int(response.Pagination.Total), "response pagination total")
				s.Assert().Equal(tc.expPageResp.NextKey, response.Pagination.NextKey, "response pagination next key")
			}
		})
	}
}
