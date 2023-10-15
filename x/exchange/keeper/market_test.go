package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/exchange/keeper"
)

func (s *TestSuite) TestKeeper_IterateKnownMarketIDs() {
	var marketIDs []uint32
	stopAfter := func(n int) func(marketID uint32) bool {
		return func(marketID uint32) bool {
			marketIDs = append(marketIDs, marketID)
			return len(marketIDs) >= n
		}
	}
	getAll := func() func(marketID uint32) bool {
		return func(marketID uint32) bool {
			marketIDs = append(marketIDs, marketID)
			return false
		}
	}

	tests := []struct {
		name         string
		setup        func(s *TestSuite)
		cb           func(marketID uint32) bool
		expMarketIDs []uint32
	}{
		{
			name:         "no known market ids",
			setup:        nil,
			cb:           getAll(),
			expMarketIDs: nil,
		},
		{
			name: "one known market id",
			setup: func(s *TestSuite) {
				keeper.SetMarketKnown(s.getStore(), 88)
			},
			cb:           getAll(),
			expMarketIDs: []uint32{88},
		},
		{
			name: "three market ids: get all",
			setup: func(s *TestSuite) {
				store := s.getStore()
				keeper.SetMarketKnown(store, 88)
				keeper.SetMarketKnown(store, 3)
				keeper.SetMarketKnown(store, 50)
			},
			cb:           getAll(),
			expMarketIDs: []uint32{3, 50, 88},
		},
		{
			name: "three market ids: get one",
			setup: func(s *TestSuite) {
				store := s.getStore()
				keeper.SetMarketKnown(store, 88)
				keeper.SetMarketKnown(store, 3)
				keeper.SetMarketKnown(store, 50)
			},
			cb:           stopAfter(1),
			expMarketIDs: []uint32{3},
		},
		{
			name: "three market ids: get two",
			setup: func(s *TestSuite) {
				store := s.getStore()
				keeper.SetMarketKnown(store, 88)
				keeper.SetMarketKnown(store, 3)
				keeper.SetMarketKnown(store, 50)
			},
			cb:           stopAfter(2),
			expMarketIDs: []uint32{3, 50},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup(s)
			}

			marketIDs = nil
			testFunc := func() {
				s.k.IterateKnownMarketIDs(s.ctx, tc.cb)
			}
			s.Require().NotPanics(testFunc, "IterateKnownMarketIDs")
			s.Assert().Equal(tc.expMarketIDs, marketIDs, "IterateKnownMarketIDs market ids")
		})
	}
}

func (s *TestSuite) TestKeeper_GetCreateAskFlatFees() {
	setter := keeper.SetCreateAskFlatFees
	tests := []struct {
		name     string
		setup    func(s *TestSuite)
		marketID uint32
		expected []sdk.Coin
	}{
		{
			name:     "no entries at all",
			setup:    nil,
			marketID: 1,
			expected: nil,
		},
		{
			name: "no entries for market",
			setup: func(s *TestSuite) {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin(8, "acorn")})
				setter(store, 3, []sdk.Coin{s.coin(3, "apple")})
			},
			marketID: 2,
			expected: nil,
		},
		{
			name: "market with one entry",
			setup: func(s *TestSuite) {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin(8, "acorn")})
				setter(store, 2, []sdk.Coin{s.coin(5, "avocado")})
				setter(store, 3, []sdk.Coin{s.coin(3, "apple")})
			},
			marketID: 2,
			expected: []sdk.Coin{s.coin(5, "avocado")},
		},
		{
			name: "market with two coins",
			setup: func(s *TestSuite) {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin(1, "acorn")})
				setter(store, 2, []sdk.Coin{s.coin(8, "plum"), s.coin(2, "apple")})
				setter(store, 3, []sdk.Coin{s.coin(3, "acorn")})
			},
			marketID: 2,
			expected: []sdk.Coin{s.coin(2, "apple"), s.coin(8, "plum")},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup(s)
			}

			var actual []sdk.Coin
			testFunc := func() {
				actual = s.k.GetCreateAskFlatFees(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "GetCreateAskFlatFees(%d)", tc.marketID)
			s.Assert().Equal(s.coinsString(tc.expected), s.coinsString(actual), "GetCreateAskFlatFees(%d)", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_GetCreateBidFlatFees() {
	setter := keeper.SetCreateBidFlatFees
	tests := []struct {
		name     string
		setup    func(s *TestSuite)
		marketID uint32
		expected []sdk.Coin
	}{
		{
			name:     "no entries at all",
			setup:    nil,
			marketID: 1,
			expected: nil,
		},
		{
			name: "no entries for market",
			setup: func(s *TestSuite) {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin(8, "acorn")})
				setter(store, 3, []sdk.Coin{s.coin(3, "apple")})
			},
			marketID: 2,
			expected: nil,
		},
		{
			name: "market with one entry",
			setup: func(s *TestSuite) {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin(8, "acorn")})
				setter(store, 2, []sdk.Coin{s.coin(5, "avocado")})
				setter(store, 3, []sdk.Coin{s.coin(3, "apple")})
			},
			marketID: 2,
			expected: []sdk.Coin{s.coin(5, "avocado")},
		},
		{
			name: "market with two coins",
			setup: func(s *TestSuite) {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin(1, "acorn")})
				setter(store, 2, []sdk.Coin{s.coin(8, "plum"), s.coin(2, "apple")})
				setter(store, 3, []sdk.Coin{s.coin(3, "acorn")})
			},
			marketID: 2,
			expected: []sdk.Coin{s.coin(2, "apple"), s.coin(8, "plum")},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup(s)
			}

			var actual []sdk.Coin
			testFunc := func() {
				actual = s.k.GetCreateBidFlatFees(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "GetCreateBidFlatFees(%d)", tc.marketID)
			s.Assert().Equal(s.coinsString(tc.expected), s.coinsString(actual), "GetCreateBidFlatFees(%d)", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_GetSellerSettlementFlatFees() {
	setter := keeper.SetSellerSettlementFlatFees
	tests := []struct {
		name     string
		setup    func(s *TestSuite)
		marketID uint32
		expected []sdk.Coin
	}{
		{
			name:     "no entries at all",
			setup:    nil,
			marketID: 1,
			expected: nil,
		},
		{
			name: "no entries for market",
			setup: func(s *TestSuite) {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin(8, "acorn")})
				setter(store, 3, []sdk.Coin{s.coin(3, "apple")})
			},
			marketID: 2,
			expected: nil,
		},
		{
			name: "market with one entry",
			setup: func(s *TestSuite) {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin(8, "acorn")})
				setter(store, 2, []sdk.Coin{s.coin(5, "avocado")})
				setter(store, 3, []sdk.Coin{s.coin(3, "apple")})
			},
			marketID: 2,
			expected: []sdk.Coin{s.coin(5, "avocado")},
		},
		{
			name: "market with two coins",
			setup: func(s *TestSuite) {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin(1, "acorn")})
				setter(store, 2, []sdk.Coin{s.coin(8, "plum"), s.coin(2, "apple")})
				setter(store, 3, []sdk.Coin{s.coin(3, "acorn")})
			},
			marketID: 2,
			expected: []sdk.Coin{s.coin(2, "apple"), s.coin(8, "plum")},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup(s)
			}

			var actual []sdk.Coin
			testFunc := func() {
				actual = s.k.GetSellerSettlementFlatFees(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "GetSellerSettlementFlatFees(%d)", tc.marketID)
			s.Assert().Equal(s.coinsString(tc.expected), s.coinsString(actual), "GetSellerSettlementFlatFees(%d)", tc.marketID)
		})
	}
}

// TODO[1658]: func (s *TestSuite) TestKeeper_GetSellerSettlementRatios()

func (s *TestSuite) TestKeeper_GetBuyerSettlementFlatFees() {
	setter := keeper.SetBuyerSettlementFlatFees
	tests := []struct {
		name     string
		setup    func(s *TestSuite)
		marketID uint32
		expected []sdk.Coin
	}{
		{
			name:     "no entries at all",
			setup:    nil,
			marketID: 1,
			expected: nil,
		},
		{
			name: "no entries for market",
			setup: func(s *TestSuite) {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin(8, "acorn")})
				setter(store, 3, []sdk.Coin{s.coin(3, "apple")})
			},
			marketID: 2,
			expected: nil,
		},
		{
			name: "market with one entry",
			setup: func(s *TestSuite) {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin(8, "acorn")})
				setter(store, 2, []sdk.Coin{s.coin(5, "avocado")})
				setter(store, 3, []sdk.Coin{s.coin(3, "apple")})
			},
			marketID: 2,
			expected: []sdk.Coin{s.coin(5, "avocado")},
		},
		{
			name: "market with two coins",
			setup: func(s *TestSuite) {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin(1, "acorn")})
				setter(store, 2, []sdk.Coin{s.coin(8, "plum"), s.coin(2, "apple")})
				setter(store, 3, []sdk.Coin{s.coin(3, "acorn")})
			},
			marketID: 2,
			expected: []sdk.Coin{s.coin(2, "apple"), s.coin(8, "plum")},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup(s)
			}

			var actual []sdk.Coin
			testFunc := func() {
				actual = s.k.GetBuyerSettlementFlatFees(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "GetBuyerSettlementFlatFees(%d)", tc.marketID)
			s.Assert().Equal(s.coinsString(tc.expected), s.coinsString(actual), "GetBuyerSettlementFlatFees(%d)", tc.marketID)
		})
	}
}

// TODO[1658]: func (s *TestSuite) TestKeeper_GetBuyerSettlementRatios()

// TODO[1658]: func (s *TestSuite) TestKeeper_CalculateSellerSettlementRatioFee()

// TODO[1658]: func (s *TestSuite) TestKeeper_CalculateBuyerSettlementRatioFeeOptions()

// TODO[1658]: func (s *TestSuite) TestKeeper_ValidateCreateAskFlatFee()

// TODO[1658]: func (s *TestSuite) TestKeeper_ValidateCreateBidFlatFee()

// TODO[1658]: func (s *TestSuite) TestKeeper_ValidateSellerSettlementFlatFee()

// TODO[1658]: func (s *TestSuite) TestKeeper_ValidateAskPrice()

// TODO[1658]: func (s *TestSuite) TestKeeper_ValidateBuyerSettlementFee()

// TODO[1658]: func (s *TestSuite) TestKeeper_UpdateFees()

// TODO[1658]: func (s *TestSuite) TestKeeper_IsMarketActive()

// TODO[1658]: func (s *TestSuite) TestKeeper_UpdateMarketActive()

// TODO[1658]: func (s *TestSuite) TestKeeper_IsUserSettlementAllowed()

// TODO[1658]: func (s *TestSuite) TestKeeper_UpdateUserSettlementAllowed()

// TODO[1658]: func (s *TestSuite) TestKeeper_HasPermission()

// TODO[1658]: func (s *TestSuite) TestKeeper_CanSettleOrders()

// TODO[1658]: func (s *TestSuite) TestKeeper_CanSetIDs()

// TODO[1658]: func (s *TestSuite) TestKeeper_CanCancelOrdersForMarket()

// TODO[1658]: func (s *TestSuite) TestKeeper_CanWithdrawMarketFunds()

// TODO[1658]: func (s *TestSuite) TestKeeper_CanUpdateMarket()

// TODO[1658]: func (s *TestSuite) TestKeeper_CanManagePermissions()

// TODO[1658]: func (s *TestSuite) TestKeeper_CanManageReqAttrs()

// TODO[1658]: func (s *TestSuite) TestKeeper_GetUserPermissions()

// TODO[1658]: func (s *TestSuite) TestKeeper_GetAccessGrants()

// TODO[1658]: func (s *TestSuite) TestKeeper_UpdatePermissions()

// TODO[1658]: func (s *TestSuite) TestKeeper_GetReqAttrsAsk()

// TODO[1658]: func (s *TestSuite) TestKeeper_GetReqAttrsBid()

// TODO[1658]: func (s *TestSuite) TestKeeper_CanCreateAsk()

// TODO[1658]: func (s *TestSuite) TestKeeper_CanCreateBid()

// TODO[1658]: func (s *TestSuite) TestKeeper_UpdateReqAttrs()

// TODO[1658]: func (s *TestSuite) TestKeeper_GetMarketAccount()

// TODO[1658]: func (s *TestSuite) TestKeeper_GetMarketDetails()

// TODO[1658]: func (s *TestSuite) TestKeeper_UpdateMarketDetails()

// TODO[1658]: func (s *TestSuite) TestKeeper_CreateMarket()

// TODO[1658]: func (s *TestSuite) TestKeeper_GetMarket()

// TODO[1658]: func (s *TestSuite) TestKeeper_IterateMarkets()

// TODO[1658]: func (s *TestSuite) TestKeeper_GetMarketBrief()

// TODO[1658]: func (s *TestSuite) TestKeeper_WithdrawMarketFunds()

// TODO[1658]: func (s *TestSuite) TestKeeper_ValidateMarket()
