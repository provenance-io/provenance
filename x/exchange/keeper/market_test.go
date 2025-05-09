package keeper_test

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/x/exchange"
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
		setup        func()
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
			setup: func() {
				keeper.SetMarketKnown(s.getStore(), 88)
			},
			cb:           getAll(),
			expMarketIDs: []uint32{88},
		},
		{
			name: "three market ids: get all",
			setup: func() {
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
			setup: func() {
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
			setup: func() {
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
				tc.setup()
			}

			marketIDs = nil
			testFunc := func() {
				s.k.IterateKnownMarketIDs(s.ctx, tc.cb)
			}
			s.Require().NotPanics(testFunc, "IterateKnownMarketIDs")
			assertEqualSlice(s, tc.expMarketIDs, marketIDs, func(marketID uint32) string {
				return fmt.Sprintf("%d", marketID)
			}, "IterateKnownMarketIDs market ids")
		})
	}
}

func (s *TestSuite) TestKeeper_GetCreateAskFlatFees() {
	setter := keeper.SetCreateAskFlatFees
	tests := []struct {
		name     string
		setup    func()
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
			setup: func() {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin("8acorn")})
				setter(store, 3, []sdk.Coin{s.coin("3apple")})
			},
			marketID: 2,
			expected: nil,
		},
		{
			name: "market with one entry",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin("8acorn")})
				setter(store, 2, []sdk.Coin{s.coin("5avocado")})
				setter(store, 3, []sdk.Coin{s.coin("3apple")})
			},
			marketID: 2,
			expected: []sdk.Coin{s.coin("5avocado")},
		},
		{
			name: "market with two coins",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin("1acorn")})
				setter(store, 2, []sdk.Coin{s.coin("8plum"), s.coin("2apple")})
				setter(store, 3, []sdk.Coin{s.coin("3acorn")})
			},
			marketID: 2,
			expected: []sdk.Coin{s.coin("2apple"), s.coin("8plum")},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var actual []sdk.Coin
			testFunc := func() {
				actual = s.k.GetCreateAskFlatFees(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "GetCreateAskFlatFees(%d)", tc.marketID)
			s.Assert().Equal(s.coinsString(tc.expected), s.coinsString(actual),
				"GetCreateAskFlatFees(%d)", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_GetCreateBidFlatFees() {
	setter := keeper.SetCreateBidFlatFees
	tests := []struct {
		name     string
		setup    func()
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
			setup: func() {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin("8acorn")})
				setter(store, 3, []sdk.Coin{s.coin("3apple")})
			},
			marketID: 2,
			expected: nil,
		},
		{
			name: "market with one entry",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin("8acorn")})
				setter(store, 2, []sdk.Coin{s.coin("5avocado")})
				setter(store, 3, []sdk.Coin{s.coin("3apple")})
			},
			marketID: 2,
			expected: []sdk.Coin{s.coin("5avocado")},
		},
		{
			name: "market with two coins",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin("1acorn")})
				setter(store, 2, []sdk.Coin{s.coin("8plum"), s.coin("2apple")})
				setter(store, 3, []sdk.Coin{s.coin("3acorn")})
			},
			marketID: 2,
			expected: []sdk.Coin{s.coin("2apple"), s.coin("8plum")},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var actual []sdk.Coin
			testFunc := func() {
				actual = s.k.GetCreateBidFlatFees(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "GetCreateBidFlatFees(%d)", tc.marketID)
			s.Assert().Equal(s.coinsString(tc.expected), s.coinsString(actual),
				"GetCreateBidFlatFees(%d)", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_GetCreateCommitmentFlatFees() {
	setter := keeper.SetCreateCommitmentFlatFees
	tests := []struct {
		name     string
		setup    func()
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
			setup: func() {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin("8acorn")})
				setter(store, 3, []sdk.Coin{s.coin("3apple")})
			},
			marketID: 2,
			expected: nil,
		},
		{
			name: "market with one entry",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin("8acorn")})
				setter(store, 2, []sdk.Coin{s.coin("5avocado")})
				setter(store, 3, []sdk.Coin{s.coin("3apple")})
			},
			marketID: 2,
			expected: []sdk.Coin{s.coin("5avocado")},
		},
		{
			name: "market with two coins",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin("1acorn")})
				setter(store, 2, []sdk.Coin{s.coin("8plum"), s.coin("2apple")})
				setter(store, 3, []sdk.Coin{s.coin("3acorn")})
			},
			marketID: 2,
			expected: []sdk.Coin{s.coin("2apple"), s.coin("8plum")},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var actual []sdk.Coin
			testFunc := func() {
				actual = s.k.GetCreateCommitmentFlatFees(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "GetCreateCommitmentFlatFees(%d)", tc.marketID)
			s.Assert().Equal(s.coinsString(tc.expected), s.coinsString(actual),
				"GetCreateCommitmentFlatFees(%d)", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_GetSellerSettlementFlatFees() {
	setter := keeper.SetSellerSettlementFlatFees
	tests := []struct {
		name     string
		setup    func()
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
			setup: func() {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin("8acorn")})
				setter(store, 3, []sdk.Coin{s.coin("3apple")})
			},
			marketID: 2,
			expected: nil,
		},
		{
			name: "market with one entry",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin("8acorn")})
				setter(store, 2, []sdk.Coin{s.coin("5avocado")})
				setter(store, 3, []sdk.Coin{s.coin("3apple")})
			},
			marketID: 2,
			expected: []sdk.Coin{s.coin("5avocado")},
		},
		{
			name: "market with two coins",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin("1acorn")})
				setter(store, 2, []sdk.Coin{s.coin("8plum"), s.coin("2apple")})
				setter(store, 3, []sdk.Coin{s.coin("3acorn")})
			},
			marketID: 2,
			expected: []sdk.Coin{s.coin("2apple"), s.coin("8plum")},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var actual []sdk.Coin
			testFunc := func() {
				actual = s.k.GetSellerSettlementFlatFees(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "GetSellerSettlementFlatFees(%d)", tc.marketID)
			s.Assert().Equal(s.coinsString(tc.expected), s.coinsString(actual),
				"GetSellerSettlementFlatFees(%d)", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_GetSellerSettlementRatios() {
	setter := keeper.SetSellerSettlementRatios
	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		expected []exchange.FeeRatio
	}{
		{
			name:     "no entries at all",
			setup:    nil,
			marketID: 1,
			expected: nil,
		},
		{
			name: "no entries for market",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []exchange.FeeRatio{s.ratio("8peach:1fig")})
				setter(store, 3, []exchange.FeeRatio{s.ratio("10plum:1fig")})
			},
			marketID: 2,
			expected: nil,
		},
		{
			name: "market with one entry",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []exchange.FeeRatio{s.ratio("8peach:1fig")})
				setter(store, 2, []exchange.FeeRatio{s.ratio("50pear:3fig")})
				setter(store, 3, []exchange.FeeRatio{s.ratio("10plum:1fig")})
			},
			marketID: 2,
			expected: []exchange.FeeRatio{s.ratio("50pear:3fig")},
		},
		{
			name: "market with two coins",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []exchange.FeeRatio{s.ratio("8peach:1fig")})
				setter(store, 2, []exchange.FeeRatio{
					s.ratio("50pear:3fig"),
					s.ratio("100apple:7grape"),
				})
				setter(store, 3, []exchange.FeeRatio{s.ratio("10plum:1fig")})
			},
			marketID: 2,
			expected: []exchange.FeeRatio{
				s.ratio("100apple:7grape"),
				s.ratio("50pear:3fig"),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var actual []exchange.FeeRatio
			testFunc := func() {
				actual = s.k.GetSellerSettlementRatios(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "GetSellerSettlementRatios(%d)", tc.marketID)
			s.Assert().Equal(s.ratiosStrings(tc.expected), s.ratiosStrings(actual),
				"GetSellerSettlementRatios(%d)", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_GetBuyerSettlementFlatFees() {
	setter := keeper.SetBuyerSettlementFlatFees
	tests := []struct {
		name     string
		setup    func()
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
			setup: func() {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin("8acorn")})
				setter(store, 3, []sdk.Coin{s.coin("3apple")})
			},
			marketID: 2,
			expected: nil,
		},
		{
			name: "market with one entry",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin("8acorn")})
				setter(store, 2, []sdk.Coin{s.coin("5avocado")})
				setter(store, 3, []sdk.Coin{s.coin("3apple")})
			},
			marketID: 2,
			expected: []sdk.Coin{s.coin("5avocado")},
		},
		{
			name: "market with two coins",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []sdk.Coin{s.coin("1acorn")})
				setter(store, 2, []sdk.Coin{s.coin("8plum"), s.coin("2apple")})
				setter(store, 3, []sdk.Coin{s.coin("3acorn")})
			},
			marketID: 2,
			expected: []sdk.Coin{s.coin("2apple"), s.coin("8plum")},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var actual []sdk.Coin
			testFunc := func() {
				actual = s.k.GetBuyerSettlementFlatFees(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "GetBuyerSettlementFlatFees(%d)", tc.marketID)
			s.Assert().Equal(s.coinsString(tc.expected), s.coinsString(actual),
				"GetBuyerSettlementFlatFees(%d)", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_GetBuyerSettlementRatios() {
	setter := keeper.SetBuyerSettlementRatios
	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		expected []exchange.FeeRatio
	}{
		{
			name:     "no entries at all",
			setup:    nil,
			marketID: 1,
			expected: nil,
		},
		{
			name: "no entries for market",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []exchange.FeeRatio{s.ratio("8peach:1fig")})
				setter(store, 3, []exchange.FeeRatio{s.ratio("10plum:1fig")})
			},
			marketID: 2,
			expected: nil,
		},
		{
			name: "market with one entry",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []exchange.FeeRatio{s.ratio("8peach:1fig")})
				setter(store, 2, []exchange.FeeRatio{s.ratio("50pear:3fig")})
				setter(store, 3, []exchange.FeeRatio{s.ratio("10plum:1fig")})
			},
			marketID: 2,
			expected: []exchange.FeeRatio{s.ratio("50pear:3fig")},
		},
		{
			name: "market with two coins",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []exchange.FeeRatio{s.ratio("8peach:1fig")})
				setter(store, 2, []exchange.FeeRatio{
					s.ratio("50pear:3fig"),
					s.ratio("100apple:7grape"),
				})
				setter(store, 3, []exchange.FeeRatio{s.ratio("10plum:1fig")})
			},
			marketID: 2,
			expected: []exchange.FeeRatio{
				s.ratio("100apple:7grape"),
				s.ratio("50pear:3fig"),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var actual []exchange.FeeRatio
			testFunc := func() {
				actual = s.k.GetBuyerSettlementRatios(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "GetBuyerSettlementRatios(%d)", tc.marketID)
			s.Assert().Equal(s.ratiosStrings(tc.expected), s.ratiosStrings(actual),
				"GetBuyerSettlementRatios(%d)", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_GetCommitmentSettlementBips() {
	setter := keeper.SetCommitmentSettlementBips
	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		expected uint32
	}{
		{
			name:     "no entries at all",
			setup:    nil,
			marketID: 1,
			expected: 0,
		},
		{
			name: "no entry for market",
			setup: func() {
				store := s.getStore()
				setter(store, 1, 10)
				setter(store, 3, 30)
			},
			marketID: 2,
			expected: 0,
		},
		{
			name: "market has entry",
			setup: func() {
				store := s.getStore()
				setter(store, 1, 10)
				setter(store, 2, 20)
				setter(store, 3, 30)
			},
			marketID: 2,
			expected: 20,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var actual uint32
			testFunc := func() {
				actual = s.k.GetCommitmentSettlementBips(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "GetCommitmentSettlementBips(%d)", tc.marketID)
			s.Assert().Equal(int(tc.expected), int(actual), "GetCommitmentSettlementBips(%d)", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_GetIntermediaryDenom() {
	setter := keeper.SetIntermediaryDenom
	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		expected string
	}{
		{
			name:     "no entries at all",
			setup:    nil,
			marketID: 1,
			expected: "",
		},
		{
			name: "no entry for market",
			setup: func() {
				store := s.getStore()
				setter(store, 1, "one")
				setter(store, 3, "three")
			},
			marketID: 2,
			expected: "",
		},
		{
			name: "market has entry",
			setup: func() {
				store := s.getStore()
				setter(store, 1, "one")
				setter(store, 2, "two")
				setter(store, 3, "three")
			},
			marketID: 2,
			expected: "two",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var actual string
			testFunc := func() {
				actual = s.k.GetIntermediaryDenom(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "GetIntermediaryDenom(%d)", tc.marketID)
			s.Assert().Equal(tc.expected, actual, "GetIntermediaryDenom(%d)", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_CalculateSellerSettlementRatioFee() {
	setter := keeper.SetSellerSettlementRatios
	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		price    sdk.Coin
		expFee   *sdk.Coin
		expErr   string
	}{
		{
			name:     "no ratios in store",
			setup:    nil,
			marketID: 1,
			price:    s.coin("100plum"),
			expFee:   nil,
			expErr:   "",
		},
		{
			name: "no ratios for market",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []exchange.FeeRatio{s.ratio("8peach:1peach")})
				setter(store, 3, []exchange.FeeRatio{s.ratio("10plum:1plum")})
			},
			marketID: 2,
			price:    s.coin("100plum"),
			expFee:   nil,
			expErr:   "",
		},
		{
			name: "no ratio for price denom",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []exchange.FeeRatio{s.ratio("8peach:1peach")})
				setter(store, 2, []exchange.FeeRatio{
					s.ratio("10prune:1prune"),
					s.ratio("50pear:3pear"),
				})
				setter(store, 3, []exchange.FeeRatio{s.ratio("10plum:1plum")})
			},
			marketID: 2,
			price:    s.coin("100pears"),
			expErr:   "no seller settlement fee ratio found for denom \"pears\"",
		},
		{
			name: "ratio evenly applicable",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []exchange.FeeRatio{s.ratio("8peach:1peach")})
				setter(store, 2, []exchange.FeeRatio{s.ratio("50pear:3pear")})
				setter(store, 3, []exchange.FeeRatio{s.ratio("10plum:1plum")})
			},
			marketID: 2,
			price:    s.coin("350pear"),
			expFee:   s.coinP("21pear"),
		},
		{
			name: "ratio not evenly applicable",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []exchange.FeeRatio{s.ratio("8peach:1peach")})
				setter(store, 2, []exchange.FeeRatio{s.ratio("50pear:3pear")})
				setter(store, 3, []exchange.FeeRatio{s.ratio("10plum:1plum")})
			},
			marketID: 2,
			price:    s.coin("442pear"),
			expFee:   s.coinP("27pear"),
		},
		{
			name: "error applying ratio",
			setup: func() {
				setter(s.getStore(), 1, []exchange.FeeRatio{s.ratio("0peach:1peach")})
			},
			marketID: 1,
			price:    s.coin("100peach"),
			expErr:   "invalid seller settlement fees: cannot apply ratio 0peach:1peach to price 100peach: division by zero",
		},
		{
			name: "three ratios: first",
			setup: func() {
				setter(s.getStore(), 8, []exchange.FeeRatio{
					s.ratio("10plum:1plum"),
					s.ratio("25prune:2prune"),
					s.ratio("50pear:3pear"),
				})
			},
			marketID: 8,
			price:    s.coin("500plum"),
			expFee:   s.coinP("50plum"), // 500 * 1 = 500, 500 / 10 = 50 => 50.
		},
		{
			name: "three ratios: second",
			setup: func() {
				setter(s.getStore(), 777, []exchange.FeeRatio{
					s.ratio("10plum:1plum"),
					s.ratio("25prune:2prune"),
					s.ratio("50pear:3pear"),
				})
			},
			marketID: 777,
			price:    s.coin("732prune"),
			expFee:   s.coinP("59prune"), // 732 * 2 = 1464, 1464 / 25 = 58.56 => 59.
		},
		{
			name: "three ratios: third",
			setup: func() {
				setter(s.getStore(), 41, []exchange.FeeRatio{
					s.ratio("10plum:1plum"),
					s.ratio("25prune:2prune"),
					s.ratio("50pear:3pear"),
				})
			},
			marketID: 41,
			price:    s.coin("123456pear"),
			expFee:   s.coinP("7408pear"), // 123456 * 3 = 370368, 370368 / 50 = 7407.36 => 7408.
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var fee *sdk.Coin
			var err error
			testFunc := func() {
				fee, err = s.k.CalculateSellerSettlementRatioFee(s.ctx, tc.marketID, tc.price)
			}
			s.Require().NotPanics(testFunc, "CalculateSellerSettlementRatioFee(%d, %q)", tc.marketID, tc.price)
			s.assertErrorValue(err, tc.expErr, "CalculateSellerSettlementRatioFee(%d, %q)", tc.marketID, tc.price)
			s.Assert().Equal(s.coinPString(tc.expFee), s.coinPString(fee),
				"CalculateSellerSettlementRatioFee(%d, %q)", tc.marketID, tc.price)
		})
	}
}

func (s *TestSuite) TestKeeper_CalculateBuyerSettlementRatioFeeOptions() {
	setter := keeper.SetBuyerSettlementRatios

	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		price    sdk.Coin
		expOpts  []sdk.Coin
		expErr   string
	}{
		{
			name:     "no ratios in state",
			setup:    nil,
			marketID: 6,
			price:    s.coin("100peach"),
			expOpts:  nil,
			expErr:   "",
		},
		{
			name: "no ratios for market",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []exchange.FeeRatio{s.ratio("11plum:1fig")})
				setter(store, 3, []exchange.FeeRatio{s.ratio("33prune:2grape")})
			},
			marketID: 2,
			price:    s.coin("100peach"),
			expOpts:  nil,
			expErr:   "",
		},
		{
			name: "no ratios for price denom: fee denom",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []exchange.FeeRatio{s.ratio("11plum:1fig")})
				setter(store, 2, []exchange.FeeRatio{
					s.ratio("21pineapple:1fig"),
					s.ratio("22pear:3fig"),
					s.ratio("23peach:4fig"),
				})
				setter(store, 3, []exchange.FeeRatio{s.ratio("33prune:2grape")})
			},
			marketID: 2,
			price:    s.coin("100fig"),
			expErr:   "no buyer settlement fee ratios found for denom \"fig\"",
		},
		{
			name: "no ratios for price denom: other market's denom",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []exchange.FeeRatio{s.ratio("11plum:1fig")})
				setter(store, 2, []exchange.FeeRatio{
					s.ratio("21pineapple:1fig"),
					s.ratio("22pear:3fig"),
					s.ratio("23peach:4fig"),
				})
				setter(store, 3, []exchange.FeeRatio{s.ratio("33prune:2grape")})
			},
			marketID: 2,
			price:    s.coin("100prune"),
			expErr:   "no buyer settlement fee ratios found for denom \"prune\"",
		},
		{
			name: "one ratio: evenly divisible",
			setup: func() {
				setter(s.getStore(), 15, []exchange.FeeRatio{s.ratio("500pineapple:1fig")})
			},
			marketID: 15,
			price:    s.coin("7500pineapple"),
			expOpts:  []sdk.Coin{s.coin("15fig")},
		},
		{
			name: "one ratio: not evenly divisible",
			setup: func() {
				setter(s.getStore(), 15, []exchange.FeeRatio{s.ratio("500pineapple:1fig")})
			},
			marketID: 15,
			price:    s.coin("7503pineapple"),
			// 7503pineapple * 1fig/500pineapple = 15.006 => 16fig
			expOpts: s.coins("16fig"),
		},
		{
			name: "three ratios for denom: none divisible",
			setup: func() {
				setter(s.getStore(), 21, []exchange.FeeRatio{
					s.ratio("123plum:1fig"),
					s.ratio("234plum:5grape"),
					s.ratio("345plum:7honeydew"),
					s.ratio("100peach:3fig"),
					s.ratio("200peach:11grape"),
					s.ratio("300peach:13honeydew"),
				})
			},
			marketID: 21,
			price:    s.coin("3000plum"),
			// 3000plum * 1fig/123plum = 24.39 => 25fig
			// 3000plum * 5grape/234plum = 64.10 => 65grape
			// 3000plum * 7honeydew/345plum = 60.86 => 61honeydew
			expOpts: s.coins("25fig,65grape,61honeydew"),
		},
		{
			name: "three ratios for denom: first not divisible",
			setup: func() {
				setter(s.getStore(), 1, []exchange.FeeRatio{
					s.ratio("123plum:1fig"),
					s.ratio("234plum:5grape"),
					s.ratio("345plum:7honeydew"),
					s.ratio("100peach:3fig"),
					s.ratio("200peach:11grape"),
					s.ratio("300peach:13honeydew"),
				})
			},
			marketID: 1,
			price:    s.coin("26910plum"),
			// 26910plum * 1fig/123plum = 218.78 => 219fig
			// 26910plum * 5grape/234plum = 575grape
			// 26910plum * 7honeydew/345plum = 546honeydew
			expOpts: []sdk.Coin{s.coin("219fig"), s.coin("575grape"), s.coin("546honeydew")},
		},
		{
			name: "three ratios for denom: second not divisible",
			setup: func() {
				setter(s.getStore(), 1, []exchange.FeeRatio{
					s.ratio("123plum:1fig"),
					s.ratio("234plum:5grape"),
					s.ratio("345plum:7honeydew"),
					s.ratio("100peach:3fig"),
					s.ratio("200peach:11grape"),
					s.ratio("300peach:13honeydew"),
				})
			},
			marketID: 1,
			price:    s.coin("50100peach"),
			// 50100peach * 3fig/100peach = 1503fig
			// 50100peach * 11grape/200peach = 2755.5 => 2756grape
			// 50100peach * 13honeydew/300peach = 2171honeydew
			expOpts: []sdk.Coin{s.coin("1503fig"), s.coin("2756grape"), s.coin("2171honeydew")},
		},
		{
			name: "three ratios for denom: third not divisible",
			setup: func() {
				setter(s.getStore(), 1, []exchange.FeeRatio{
					s.ratio("123plum:1fig"),
					s.ratio("234plum:5grape"),
					s.ratio("345plum:7honeydew"),
					s.ratio("100peach:3fig"),
					s.ratio("200peach:11grape"),
					s.ratio("300peach:13honeydew"),
				})
			},
			marketID: 1,
			price:    s.coin("50200peach"),
			// 50200peach * 3fig/100peach = 1506fig
			// 50200peach * 11grape/200peach = 2761grape
			// 50200peach * 13honeydew/300peach = 2175.33 => 2176honeydew
			expOpts: []sdk.Coin{s.coin("1506fig"), s.coin("2761grape"), s.coin("2176honeydew")},
		},
		{
			name: "three ratios for denom: all divisible",
			setup: func() {
				setter(s.getStore(), 5, []exchange.FeeRatio{
					s.ratio("123plum:1fig"),
					s.ratio("234plum:5grape"),
					s.ratio("345plum:7honeydew"),
					s.ratio("100peach:3fig"),
					s.ratio("200peach:11grape"),
					s.ratio("300peach:13honeydew"),
				})
			},
			marketID: 5,
			price:    s.coin("6000peach"),
			expOpts:  []sdk.Coin{s.coin("180fig"), s.coin("330grape"), s.coin("260honeydew")},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var opts []sdk.Coin
			var err error
			testFunc := func() {
				opts, err = s.k.CalculateBuyerSettlementRatioFeeOptions(s.ctx, tc.marketID, tc.price)
			}
			s.Require().NotPanics(testFunc, "CalculateBuyerSettlementRatioFeeOptions(%d, %q)", tc.marketID, tc.price)
			s.assertErrorValue(err, tc.expErr, "CalculateBuyerSettlementRatioFeeOptions(%d, %q)", tc.marketID, tc.price)
			s.Assert().Equal(s.coinsString(tc.expOpts), s.coinsString(opts),
				"CalculateBuyerSettlementRatioFeeOptions(%d, %q)", tc.marketID, tc.price)
		})
	}
}

func (s *TestSuite) TestKeeper_ValidateCreateAskFlatFee() {
	setter := keeper.SetCreateAskFlatFees
	name := "ask order creation"
	nilFeeErr := func(opts string) string {
		return fmt.Sprintf("no %s fee provided, must be one of: %s", name, opts)
	}
	noFeeErr := func(fee string, opts string) string {
		return fmt.Sprintf("invalid %s fee %q, must be one of: %s", name, fee, opts)
	}
	lowFeeErr := func(fee string, opts string) string {
		return fmt.Sprintf("insufficient %s fee: %q is less than required amount %q", name, fee, opts)
	}

	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		fee      *sdk.Coin
		expErr   string
	}{
		{
			name:     "no fees in store: nil",
			setup:    nil,
			marketID: 1,
			fee:      nil,
			expErr:   "",
		},
		{
			name:     "no fees in store: not nil",
			setup:    nil,
			marketID: 1,
			fee:      s.coinP("8fig"),
			expErr:   "",
		},
		{
			name: "no fees for market: nil",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("2grape")})
			},
			marketID: 6,
			fee:      nil,
			expErr:   "",
		},
		{
			name: "no fees for market: not nil",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("2grape")})
			},
			marketID: 6,
			fee:      s.coinP("30fig"),
			expErr:   "",
		},
		{
			name: "one fee option: nil",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 6, []sdk.Coin{s.coin("11fig")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("1grape")})
			},
			marketID: 6,
			fee:      nil,
			expErr:   nilFeeErr("11fig"),
		},
		{
			name: "one fee option: diff denom",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 6, []sdk.Coin{s.coin("11fig")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("1grape")})
			},
			marketID: 6,
			fee:      s.coinP("5grape"),
			expErr:   noFeeErr("5grape", "11fig"),
		},
		{
			name: "one fee option: insufficient",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 6, []sdk.Coin{s.coin("11fig")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("1grape")})
			},
			marketID: 6,
			fee:      s.coinP("10fig"),
			expErr:   lowFeeErr("10fig", "11fig"),
		},
		{
			name: "one fee option: same",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 6, []sdk.Coin{s.coin("11fig")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("1grape")})
			},
			marketID: 6,
			fee:      s.coinP("11fig"),
			expErr:   "",
		},
		{
			name: "one fee option: more",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 6, []sdk.Coin{s.coin("11fig")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("1grape")})
			},
			marketID: 6,
			fee:      s.coinP("12fig"),
			expErr:   "",
		},
		{
			name: "three fee options: nil",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      nil,
			expErr:   nilFeeErr("10fig,3grape,7honeydew"),
		},
		{
			name: "three fee options: wrong denom",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("80apple"),
			expErr:   noFeeErr("80apple", "10fig,3grape,7honeydew"),
		},
		{
			name: "three fee options: first, low",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("9fig"),
			expErr:   lowFeeErr("9fig", "10fig"),
		},
		{
			name: "three fee options: first, ok",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("10fig"),
			expErr:   "",
		},
		{
			name: "three fee options: second, low",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("2grape"),
			expErr:   lowFeeErr("2grape", "3grape"),
		},
		{
			name: "three fee options: second, ok",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("3grape"),
			expErr:   "",
		},
		{
			name: "three fee options: third, low",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("6honeydew"),
			expErr:   lowFeeErr("6honeydew", "7honeydew"),
		},
		{
			name: "three fee options: third, ok",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("7honeydew"),
			expErr:   "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var err error
			testFunc := func() {
				err = s.k.ValidateCreateAskFlatFee(s.ctx, tc.marketID, tc.fee)
			}
			s.Require().NotPanics(testFunc, "ValidateCreateAskFlatFee(%d, %s)", tc.marketID, s.coinPString(tc.fee))
			s.assertErrorValue(err, tc.expErr, "ValidateCreateAskFlatFee(%d, %s) error", tc.marketID, s.coinPString(tc.fee))
		})
	}
}

func (s *TestSuite) TestKeeper_ValidateCreateBidFlatFee() {
	setter := keeper.SetCreateBidFlatFees
	name := "bid order creation"
	nilFeeErr := func(opts string) string {
		return fmt.Sprintf("no %s fee provided, must be one of: %s", name, opts)
	}
	noFeeErr := func(fee string, opts string) string {
		return fmt.Sprintf("invalid %s fee %q, must be one of: %s", name, fee, opts)
	}
	lowFeeErr := func(fee string, opts string) string {
		return fmt.Sprintf("insufficient %s fee: %q is less than required amount %q", name, fee, opts)
	}

	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		fee      *sdk.Coin
		expErr   string
	}{
		{
			name:     "no fees in store: nil",
			setup:    nil,
			marketID: 1,
			fee:      nil,
			expErr:   "",
		},
		{
			name:     "no fees in store: not nil",
			setup:    nil,
			marketID: 1,
			fee:      s.coinP("8fig"),
			expErr:   "",
		},
		{
			name: "no fees for market: nil",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("2grape")})
			},
			marketID: 6,
			fee:      nil,
			expErr:   "",
		},
		{
			name: "no fees for market: not nil",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("2grape")})
			},
			marketID: 6,
			fee:      s.coinP("30fig"),
			expErr:   "",
		},
		{
			name: "one fee option: nil",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 6, []sdk.Coin{s.coin("11fig")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("1grape")})
			},
			marketID: 6,
			fee:      nil,
			expErr:   nilFeeErr("11fig"),
		},
		{
			name: "one fee option: diff denom",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 6, []sdk.Coin{s.coin("11fig")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("1grape")})
			},
			marketID: 6,
			fee:      s.coinP("5grape"),
			expErr:   noFeeErr("5grape", "11fig"),
		},
		{
			name: "one fee option: insufficient",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 6, []sdk.Coin{s.coin("11fig")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("1grape")})
			},
			marketID: 6,
			fee:      s.coinP("10fig"),
			expErr:   lowFeeErr("10fig", "11fig"),
		},
		{
			name: "one fee option: same",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 6, []sdk.Coin{s.coin("11fig")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("1grape")})
			},
			marketID: 6,
			fee:      s.coinP("11fig"),
			expErr:   "",
		},
		{
			name: "one fee option: more",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 6, []sdk.Coin{s.coin("11fig")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("1grape")})
			},
			marketID: 6,
			fee:      s.coinP("12fig"),
			expErr:   "",
		},
		{
			name: "three fee options: nil",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      nil,
			expErr:   nilFeeErr("10fig,3grape,7honeydew"),
		},
		{
			name: "three fee options: wrong denom",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("80apple"),
			expErr:   noFeeErr("80apple", "10fig,3grape,7honeydew"),
		},
		{
			name: "three fee options: first, low",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("9fig"),
			expErr:   lowFeeErr("9fig", "10fig"),
		},
		{
			name: "three fee options: first, ok",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("10fig"),
			expErr:   "",
		},
		{
			name: "three fee options: second, low",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("2grape"),
			expErr:   lowFeeErr("2grape", "3grape"),
		},
		{
			name: "three fee options: second, ok",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("3grape"),
			expErr:   "",
		},
		{
			name: "three fee options: third, low",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("6honeydew"),
			expErr:   lowFeeErr("6honeydew", "7honeydew"),
		},
		{
			name: "three fee options: third, ok",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("7honeydew"),
			expErr:   "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var err error
			testFunc := func() {
				err = s.k.ValidateCreateBidFlatFee(s.ctx, tc.marketID, tc.fee)
			}
			s.Require().NotPanics(testFunc, "ValidateCreateBidFlatFee(%d, %s)", tc.marketID, s.coinPString(tc.fee))
			s.assertErrorValue(err, tc.expErr, "ValidateCreateBidFlatFee(%d, %s) error", tc.marketID, s.coinPString(tc.fee))
		})
	}
}

func (s *TestSuite) TestKeeper_ValidateCreateCommitmentFlatFee() {
	setter := keeper.SetCreateCommitmentFlatFees
	name := "commitment creation"
	nilFeeErr := func(opts string) string {
		return fmt.Sprintf("no %s fee provided, must be one of: %s", name, opts)
	}
	noFeeErr := func(fee string, opts string) string {
		return fmt.Sprintf("invalid %s fee %q, must be one of: %s", name, fee, opts)
	}
	lowFeeErr := func(fee string, opts string) string {
		return fmt.Sprintf("insufficient %s fee: %q is less than required amount %q", name, fee, opts)
	}

	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		fee      *sdk.Coin
		expErr   string
	}{
		{
			name:     "no fees in store: nil",
			setup:    nil,
			marketID: 1,
			fee:      nil,
			expErr:   "",
		},
		{
			name:     "no fees in store: not nil",
			setup:    nil,
			marketID: 1,
			fee:      s.coinP("8fig"),
			expErr:   "",
		},
		{
			name: "no fees for market: nil",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("2grape")})
			},
			marketID: 6,
			fee:      nil,
			expErr:   "",
		},
		{
			name: "no fees for market: not nil",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("2grape")})
			},
			marketID: 6,
			fee:      s.coinP("30fig"),
			expErr:   "",
		},
		{
			name: "one fee option: nil",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 6, []sdk.Coin{s.coin("11fig")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("1grape")})
			},
			marketID: 6,
			fee:      nil,
			expErr:   nilFeeErr("11fig"),
		},
		{
			name: "one fee option: diff denom",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 6, []sdk.Coin{s.coin("11fig")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("1grape")})
			},
			marketID: 6,
			fee:      s.coinP("5grape"),
			expErr:   noFeeErr("5grape", "11fig"),
		},
		{
			name: "one fee option: insufficient",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 6, []sdk.Coin{s.coin("11fig")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("1grape")})
			},
			marketID: 6,
			fee:      s.coinP("10fig"),
			expErr:   lowFeeErr("10fig", "11fig"),
		},
		{
			name: "one fee option: same",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 6, []sdk.Coin{s.coin("11fig")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("1grape")})
			},
			marketID: 6,
			fee:      s.coinP("11fig"),
			expErr:   "",
		},
		{
			name: "one fee option: more",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 6, []sdk.Coin{s.coin("11fig")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("1grape")})
			},
			marketID: 6,
			fee:      s.coinP("12fig"),
			expErr:   "",
		},
		{
			name: "three fee options: nil",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      nil,
			expErr:   nilFeeErr("10fig,3grape,7honeydew"),
		},
		{
			name: "three fee options: wrong denom",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("80apple"),
			expErr:   noFeeErr("80apple", "10fig,3grape,7honeydew"),
		},
		{
			name: "three fee options: first, low",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("9fig"),
			expErr:   lowFeeErr("9fig", "10fig"),
		},
		{
			name: "three fee options: first, ok",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("10fig"),
			expErr:   "",
		},
		{
			name: "three fee options: second, low",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("2grape"),
			expErr:   lowFeeErr("2grape", "3grape"),
		},
		{
			name: "three fee options: second, ok",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("3grape"),
			expErr:   "",
		},
		{
			name: "three fee options: third, low",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("6honeydew"),
			expErr:   lowFeeErr("6honeydew", "7honeydew"),
		},
		{
			name: "three fee options: third, ok",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("7honeydew"),
			expErr:   "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var err error
			testFunc := func() {
				err = s.k.ValidateCreateCommitmentFlatFee(s.ctx, tc.marketID, tc.fee)
			}
			s.Require().NotPanics(testFunc, "ValidateCreateCommitmentFlatFee(%d, %s)", tc.marketID, s.coinPString(tc.fee))
			s.assertErrorValue(err, tc.expErr, "ValidateCreateCommitmentFlatFee(%d, %s) error", tc.marketID, s.coinPString(tc.fee))
		})
	}
}

func (s *TestSuite) TestKeeper_ValidateSellerSettlementFlatFee() {
	setter := keeper.SetSellerSettlementFlatFees
	name := "seller settlement flat"
	nilFeeErr := func(opts string) string {
		return fmt.Sprintf("no %s fee provided, must be one of: %s", name, opts)
	}
	noFeeErr := func(fee string, opts string) string {
		return fmt.Sprintf("invalid %s fee %q, must be one of: %s", name, fee, opts)
	}
	lowFeeErr := func(fee string, opts string) string {
		return fmt.Sprintf("insufficient %s fee: %q is less than required amount %q", name, fee, opts)
	}

	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		fee      *sdk.Coin
		expErr   string
	}{
		{
			name:     "no fees in store: nil",
			setup:    nil,
			marketID: 1,
			fee:      nil,
			expErr:   "",
		},
		{
			name:     "no fees in store: not nil",
			setup:    nil,
			marketID: 1,
			fee:      s.coinP("8fig"),
			expErr:   "",
		},
		{
			name: "no fees for market: nil",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("2grape")})
			},
			marketID: 6,
			fee:      nil,
			expErr:   "",
		},
		{
			name: "no fees for market: not nil",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("2grape")})
			},
			marketID: 6,
			fee:      s.coinP("30fig"),
			expErr:   "",
		},
		{
			name: "one fee option: nil",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 6, []sdk.Coin{s.coin("11fig")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("1grape")})
			},
			marketID: 6,
			fee:      nil,
			expErr:   nilFeeErr("11fig"),
		},
		{
			name: "one fee option: diff denom",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 6, []sdk.Coin{s.coin("11fig")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("1grape")})
			},
			marketID: 6,
			fee:      s.coinP("5grape"),
			expErr:   noFeeErr("5grape", "11fig"),
		},
		{
			name: "one fee option: insufficient",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 6, []sdk.Coin{s.coin("11fig")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("1grape")})
			},
			marketID: 6,
			fee:      s.coinP("10fig"),
			expErr:   lowFeeErr("10fig", "11fig"),
		},
		{
			name: "one fee option: same",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 6, []sdk.Coin{s.coin("11fig")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("1grape")})
			},
			marketID: 6,
			fee:      s.coinP("11fig"),
			expErr:   "",
		},
		{
			name: "one fee option: more",
			setup: func() {
				store := s.getStore()
				setter(store, 5, []sdk.Coin{s.coin("10fig"), s.coin("3grape")})
				setter(store, 6, []sdk.Coin{s.coin("11fig")})
				setter(store, 7, []sdk.Coin{s.coin("12fig"), s.coin("1grape")})
			},
			marketID: 6,
			fee:      s.coinP("12fig"),
			expErr:   "",
		},
		{
			name: "three fee options: nil",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      nil,
			expErr:   nilFeeErr("10fig,3grape,7honeydew"),
		},
		{
			name: "three fee options: wrong denom",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("80apple"),
			expErr:   noFeeErr("80apple", "10fig,3grape,7honeydew"),
		},
		{
			name: "three fee options: first, low",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("9fig"),
			expErr:   lowFeeErr("9fig", "10fig"),
		},
		{
			name: "three fee options: first, ok",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("10fig"),
			expErr:   "",
		},
		{
			name: "three fee options: second, low",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("2grape"),
			expErr:   lowFeeErr("2grape", "3grape"),
		},
		{
			name: "three fee options: second, ok",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("3grape"),
			expErr:   "",
		},
		{
			name: "three fee options: third, low",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("6honeydew"),
			expErr:   lowFeeErr("6honeydew", "7honeydew"),
		},
		{
			name: "three fee options: third, ok",
			setup: func() {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("7honeydew"),
			expErr:   "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var err error
			testFunc := func() {
				err = s.k.ValidateSellerSettlementFlatFee(s.ctx, tc.marketID, tc.fee)
			}
			s.Require().NotPanics(testFunc, "ValidateSellerSettlementFlatFee(%d, %s)", tc.marketID, s.coinPString(tc.fee))
			s.assertErrorValue(err, tc.expErr, "ValidateSellerSettlementFlatFee(%d, %s) error", tc.marketID, s.coinPString(tc.fee))
		})
	}
}

func (s *TestSuite) TestKeeper_ValidateAskPrice() {
	tests := []struct {
		name              string
		setup             func()
		marketID          uint32
		price             sdk.Coin
		settlementFlatFee *sdk.Coin
		expErr            string
	}{
		{
			name:              "no ratios in store",
			setup:             nil,
			marketID:          1,
			price:             s.coin("1plum"),
			settlementFlatFee: nil,
			expErr:            "",
		},
		{
			name: "no ratios in market: no flat",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 1, []exchange.FeeRatio{s.ratio("10plum:1plum")})
				keeper.SetSellerSettlementRatios(store, 3, []exchange.FeeRatio{s.ratio("11plum:1plum")})
			},
			marketID:          2,
			price:             s.coin("1plum"),
			settlementFlatFee: nil,
			expErr:            "",
		},
		{
			name: "no ratios in market: price less than flat",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 1, []exchange.FeeRatio{s.ratio("10plum:1plum")})
				keeper.SetSellerSettlementRatios(store, 3, []exchange.FeeRatio{s.ratio("11plum:1plum")})
			},
			marketID:          2,
			price:             s.coin("1plum"),
			settlementFlatFee: s.coinP("2plum"),
			expErr:            "price 1plum is not more than seller settlement flat fee 2plum",
		},
		{
			name: "no ratios in market: price equals flat",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 1, []exchange.FeeRatio{s.ratio("10plum:1plum")})
				keeper.SetSellerSettlementRatios(store, 3, []exchange.FeeRatio{s.ratio("11plum:1plum")})
			},
			marketID:          2,
			price:             s.coin("2plum"),
			settlementFlatFee: s.coinP("2plum"),
			expErr:            "price 2plum is not more than seller settlement flat fee 2plum",
		},
		{
			name: "no ratios in market: price more than flat",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 1, []exchange.FeeRatio{s.ratio("10plum:1plum")})
				keeper.SetSellerSettlementRatios(store, 3, []exchange.FeeRatio{s.ratio("11plum:1plum")})
			},
			marketID:          2,
			price:             s.coin("3plum"),
			settlementFlatFee: s.coinP("2plum"),
			expErr:            "",
		},
		{
			name: "no ratios in market: fee diff denom with larger amount",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 1, []exchange.FeeRatio{s.ratio("10plum:1plum")})
				keeper.SetSellerSettlementRatios(store, 3, []exchange.FeeRatio{s.ratio("11plum:1plum")})
			},
			marketID:          2,
			price:             s.coin("2plum"),
			settlementFlatFee: s.coinP("3fig"),
			expErr:            "",
		},
		{
			name: "one ratio: wrong denom",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 1, []exchange.FeeRatio{s.ratio("10plum:1plum")})
				keeper.SetSellerSettlementRatios(store, 2, []exchange.FeeRatio{s.ratio("12plum:1plum")})
				keeper.SetSellerSettlementRatios(store, 3, []exchange.FeeRatio{s.ratio("15plum:1plum")})
			},
			marketID:          2,
			price:             s.coin("500peach"),
			settlementFlatFee: nil,
			expErr:            "no seller settlement fee ratio found for denom \"peach\"",
		},
		{
			name: "one ratio: no flat: price less than ratio",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 1, []exchange.FeeRatio{s.ratio("10plum:1plum")})
				keeper.SetSellerSettlementRatios(store, 2, []exchange.FeeRatio{s.ratio("12plum:13plum")})
				keeper.SetSellerSettlementRatios(store, 3, []exchange.FeeRatio{s.ratio("15plum:1plum")})
			},
			marketID:          2,
			price:             s.coin("12plum"),
			settlementFlatFee: nil,
			expErr:            "price 12plum is not more than seller settlement ratio fee 13plum",
		},
		{
			name: "one ratio: no flat: price equals ratio",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 1, []exchange.FeeRatio{s.ratio("10plum:1plum")})
				keeper.SetSellerSettlementRatios(store, 2, []exchange.FeeRatio{s.ratio("12plum:11plum")})
				keeper.SetSellerSettlementRatios(store, 3, []exchange.FeeRatio{s.ratio("15plum:1plum")})
			},
			marketID:          2,
			price:             s.coin("11plum"),
			settlementFlatFee: nil,
			expErr:            "price 11plum is not more than seller settlement ratio fee 11plum",
		},
		{
			name: "one ratio: no flat: price more than ratio",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 1, []exchange.FeeRatio{s.ratio("10plum:1plum")})
				keeper.SetSellerSettlementRatios(store, 2, []exchange.FeeRatio{s.ratio("12plum:11plum")})
				keeper.SetSellerSettlementRatios(store, 3, []exchange.FeeRatio{s.ratio("15plum:1plum")})
			},
			marketID:          2,
			price:             s.coin("13plum"),
			settlementFlatFee: nil,
			expErr:            "",
		},
		{
			name: "one ratio: diff flat: price less than ratio",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 1, []exchange.FeeRatio{s.ratio("10plum:1plum")})
				keeper.SetSellerSettlementRatios(store, 2, []exchange.FeeRatio{s.ratio("12plum:13plum")})
				keeper.SetSellerSettlementRatios(store, 3, []exchange.FeeRatio{s.ratio("15plum:1plum")})
			},
			marketID:          2,
			price:             s.coin("12plum"),
			settlementFlatFee: s.coinP("20peach"),
			expErr:            "price 12plum is not more than seller settlement ratio fee 13plum",
		},
		{
			name: "one ratio: diff flat: price equals ratio",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 1, []exchange.FeeRatio{s.ratio("10plum:1plum")})
				keeper.SetSellerSettlementRatios(store, 2, []exchange.FeeRatio{s.ratio("12plum:11plum")})
				keeper.SetSellerSettlementRatios(store, 3, []exchange.FeeRatio{s.ratio("15plum:1plum")})
			},
			marketID:          2,
			price:             s.coin("11plum"),
			settlementFlatFee: s.coinP("20peach"),
			expErr:            "price 11plum is not more than seller settlement ratio fee 11plum",
		},
		{
			name: "one ratio: diff flat: price more than ratio",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 1, []exchange.FeeRatio{s.ratio("10plum:1plum")})
				keeper.SetSellerSettlementRatios(store, 2, []exchange.FeeRatio{s.ratio("12plum:11plum")})
				keeper.SetSellerSettlementRatios(store, 3, []exchange.FeeRatio{s.ratio("15plum:1plum")})
			},
			marketID:          2,
			price:             s.coin("12plum"),
			settlementFlatFee: s.coinP("20peach"),
			expErr:            "",
		},
		{
			name: "one ratio: price more than flat, more than ratio, less than total",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 1, []exchange.FeeRatio{s.ratio("10plum:1plum")})
				keeper.SetSellerSettlementRatios(store, 2, []exchange.FeeRatio{s.ratio("12plum:11plum")})
				keeper.SetSellerSettlementRatios(store, 3, []exchange.FeeRatio{s.ratio("15plum:1plum")})
			},
			marketID:          2,
			price:             s.coin("12plum"),
			settlementFlatFee: s.coinP("2plum"),
			expErr:            "price 12plum is not more than total required seller settlement fee 13plum = 2plum flat + 11plum ratio",
		},
		{
			name: "one ratio: price equals total",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 1, []exchange.FeeRatio{s.ratio("10plum:1plum")})
				keeper.SetSellerSettlementRatios(store, 2, []exchange.FeeRatio{s.ratio("12plum:7plum")})
				keeper.SetSellerSettlementRatios(store, 3, []exchange.FeeRatio{s.ratio("15plum:1plum")})
			},
			marketID:          2,
			price:             s.coin("12plum"),
			settlementFlatFee: s.coinP("5plum"),
			expErr:            "price 12plum is not more than total required seller settlement fee 12plum = 5plum flat + 7plum ratio",
		},
		{
			name: "one ratio: price more than total",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 1, []exchange.FeeRatio{s.ratio("10plum:1plum")})
				keeper.SetSellerSettlementRatios(store, 2, []exchange.FeeRatio{s.ratio("12plum:7plum")})
				keeper.SetSellerSettlementRatios(store, 3, []exchange.FeeRatio{s.ratio("15plum:1plum")})
			},
			marketID:          2,
			price:             s.coin("12plum"),
			settlementFlatFee: s.coinP("4plum"),
			expErr:            "",
		},
		{
			name: "ratio cannot be evenly applied to price, but is enough",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 1, []exchange.FeeRatio{s.ratio("10plum:1plum")})
				keeper.SetSellerSettlementRatios(store, 2, []exchange.FeeRatio{s.ratio("12plum:7plum")})
				keeper.SetSellerSettlementRatios(store, 3, []exchange.FeeRatio{s.ratio("15plum:1plum")})
			},
			marketID:          2,
			price:             s.coin("123plum"),
			settlementFlatFee: nil,
			expErr:            "",
		},
		{
			name: "error applying ratio",
			setup: func() {
				keeper.SetSellerSettlementRatios(s.getStore(), 1, []exchange.FeeRatio{s.ratio("0plum:1plum")})
			},
			marketID:          1,
			price:             s.coin("100plum"),
			settlementFlatFee: nil,
			expErr:            "cannot apply ratio 0plum:1plum to price 100plum: division by zero",
		},
		{
			name: "three ratios: wrong denom",
			setup: func() {
				keeper.SetSellerSettlementRatios(s.getStore(), 1, []exchange.FeeRatio{
					s.ratio("100peach:1peach"),
					s.ratio("200pear:3pear"),
					s.ratio("300plum:7plum"),
				})
			},
			marketID:          1,
			price:             s.coin("5000prune"),
			settlementFlatFee: nil,
			expErr:            "no seller settlement fee ratio found for denom \"prune\"",
		},
		{
			name: "three ratios: price less than total",
			setup: func() {
				keeper.SetSellerSettlementRatios(s.getStore(), 1, []exchange.FeeRatio{
					s.ratio("5000peach:1peach"),
					s.ratio("200pear:199pear"),
					s.ratio("5000plum:7plum"),
				})
			},
			marketID:          1,
			price:             s.coin("20pear"),
			settlementFlatFee: s.coinP("1pear"),
			expErr:            "price 20pear is not more than total required seller settlement fee 21pear = 1pear flat + 20pear ratio",
		},
		{
			name: "three ratios: price more",
			setup: func() {
				keeper.SetSellerSettlementRatios(s.getStore(), 1, []exchange.FeeRatio{
					s.ratio("100peach:1peach"),
					s.ratio("200pear:3pear"),
					s.ratio("300plum:7plum"),
				})
			},
			marketID:          1,
			price:             s.coin("5000pear"),
			settlementFlatFee: nil,
			expErr:            "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var err error
			testFunc := func() {
				err = s.k.ValidateAskPrice(s.ctx, tc.marketID, tc.price, tc.settlementFlatFee)
			}
			s.Require().NotPanics(testFunc, "ValidateAskPrice(%d, %q, %s)",
				tc.marketID, tc.price, s.coinPString(tc.settlementFlatFee))
			s.assertErrorValue(err, tc.expErr, "ValidateAskPrice(%d, %q, %s)",
				tc.marketID, tc.price, s.coinPString(tc.settlementFlatFee))
		})
	}
}

func (s *TestSuite) TestKeeper_ValidateBuyerSettlementFee() {
	noFeeErr := "insufficient buyer settlement fee: no fee provided"
	flatErr := func(opts string) string {
		return "required flat fee not satisfied, valid options: " + opts
	}
	ratioErr := func(opts string) string {
		return "required ratio fee not satisfied, valid ratios: " + opts
	}
	insufficientErr := func(fee string) string {
		return "insufficient buyer settlement fee " + fee
	}

	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		price    sdk.Coin
		fee      sdk.Coins
		expErr   string
	}{
		{
			name:     "empty state: no fee",
			setup:    nil,
			marketID: 8,
			price:    s.coin("50peach"),
			fee:      nil,
			expErr:   "",
		},
		{
			name:     "empty state: with fee",
			setup:    nil,
			marketID: 8,
			price:    s.coin("100peach"),
			fee:      s.coins("120peach"), // This is okay because it's added to the price.
			expErr:   "",
		},
		{
			name: "no flat no ratio: no fee",
			setup: func() {
				store := s.getStore()
				keeper.SetBuyerSettlementFlatFees(store, 1, s.coins("10peach,12plum"))
				keeper.SetBuyerSettlementRatios(store, 1, []exchange.FeeRatio{s.ratio("100peach:3fig")})
				keeper.SetBuyerSettlementFlatFees(store, 3, s.coins("14peach,8plum"))
				keeper.SetBuyerSettlementRatios(store, 3, []exchange.FeeRatio{s.ratio("100peach:1grape")})
			},
			marketID: 2,
			price:    s.coin("5000peach"),
			fee:      nil,
			expErr:   "",
		},
		{
			name: "no flat no ratio: with fee",
			setup: func() {
				store := s.getStore()
				keeper.SetBuyerSettlementFlatFees(store, 1, s.coins("10peach,12plum"))
				keeper.SetBuyerSettlementRatios(store, 1, []exchange.FeeRatio{s.ratio("100peach:3fig")})
				keeper.SetBuyerSettlementFlatFees(store, 3, s.coins("14peach,8plum"))
				keeper.SetBuyerSettlementRatios(store, 3, []exchange.FeeRatio{s.ratio("100peach:1grape")})
			},
			marketID: 2,
			price:    s.coin("5000peach"),
			fee:      s.coins("5001peach"), // This is okay because it's added to the price.
			expErr:   "",
		},
		{
			name: "only flat: no fee",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("11peach,9plum"))
			},
			marketID: 2,
			price:    s.coin("54pear"),
			fee:      nil,
			expErr: s.joinErrs(
				flatErr("11peach,9plum"),
				noFeeErr,
			),
		},
		{
			name: "only flat: wrong denom",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("11peach,9plum"))
			},
			marketID: 2,
			price:    s.coin("54pear"),
			fee:      s.coins("3pear"),
			expErr: s.joinErrs(
				"no flat fee options available for denom pear",
				flatErr("11peach,9plum"),
				insufficientErr("3pear"),
			),
		},
		{
			name: "only flat: less than req",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("11peach,9plum"))
			},
			marketID: 2,
			price:    s.coin("54pear"),
			fee:      s.coins("10peach"),
			expErr: s.joinErrs(
				"10peach is less than required flat fee 11peach",
				flatErr("11peach,9plum"),
				insufficientErr("10peach"),
			),
		},
		{
			name: "only flat: equals req",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("11peach,9plum"))
			},
			marketID: 2,
			price:    s.coin("54pear"),
			fee:      s.coins("11peach"),
			expErr:   "",
		},
		{
			name: "only flat: more than req",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("11peach,9plum"))
			},
			marketID: 2,
			price:    s.coin("54pear"),
			fee:      s.coins("10peach,10plum"),
			expErr:   "",
		},
		{
			name: "only ratio: nofee",
			setup: func() {
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("54pear"),
			fee:      nil,
			expErr: s.joinErrs(
				ratioErr("100peach:3fig,100peach:1grape"),
				noFeeErr,
			),
		},
		{
			name: "only ratio: wrong price denom",
			setup: func() {
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500pear"),
			fee:      s.coins("5grape"),
			expErr: s.joinErrs(
				"no ratio from price denom pear to fee denom grape",
				ratioErr("100peach:3fig,100peach:1grape"),
				insufficientErr("5grape"),
			),
		},
		{
			name: "only ratio: wrong fee denom",
			setup: func() {
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("20honeydew"),
			expErr: s.joinErrs(
				"no ratio from price denom peach to fee denom honeydew",
				ratioErr("100peach:3fig,100peach:1grape"),
				insufficientErr("20honeydew"),
			),
		},
		{
			name: "only ratio: less than req",
			setup: func() {
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("14fig,4grape"),
			expErr: s.joinErrs(
				"14fig is less than required ratio fee 15fig (based on price 500peach and ratio 100peach:3fig)",
				"4grape is less than required ratio fee 5grape (based on price 500peach and ratio 100peach:1grape)",
				ratioErr("100peach:3fig,100peach:1grape"),
				insufficientErr("14fig,4grape"),
			),
		},
		{
			name: "only ratio: equals req",
			setup: func() {
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("5grape"),
			expErr:   "",
		},
		{
			name: "only ratio: more than req",
			setup: func() {
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("16fig"),
			expErr:   "",
		},
		{
			name: "both: no fee",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("10fig,2honeydew"))
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      nil,
			expErr: s.joinErrs(
				flatErr("10fig,2honeydew"),
				ratioErr("100peach:3fig,100peach:1grape"),
				noFeeErr,
			),
		},
		{
			name: "both: no flat denom",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("10fig,2honeydew"))
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("5grape"),
			expErr: s.joinErrs(
				"no flat fee options available for denom grape",
				flatErr("10fig,2honeydew"),
				insufficientErr("5grape"),
			),
		},
		{
			name: "both: no ratio denom",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("10fig,2honeydew"))
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("2honeydew"),
			expErr: s.joinErrs(
				"no ratio from price denom peach to fee denom honeydew",
				ratioErr("100peach:3fig,100peach:1grape"),
				insufficientErr("2honeydew"),
			),
		},
		{
			name: "both: neither flat nor ratio denom",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("10fig,2honeydew"))
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("33apple,44banana"),
			expErr: s.joinErrs(
				"no flat fee options available for denom apple",
				"no flat fee options available for denom banana",
				flatErr("10fig,2honeydew"),
				"no ratio from price denom peach to fee denom apple",
				"no ratio from price denom peach to fee denom banana",
				ratioErr("100peach:3fig,100peach:1grape"),
				insufficientErr("33apple,44banana"),
			),
		},
		{
			name: "both: one denom: less than either",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("10fig,2honeydew"))
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("9fig"),
			expErr: s.joinErrs(
				"9fig is less than required flat fee 10fig",
				flatErr("10fig,2honeydew"),
				"9fig is less than required ratio fee 15fig (based on price 500peach and ratio 100peach:3fig)",
				ratioErr("100peach:3fig,100peach:1grape"),
				insufficientErr("9fig"),
			),
		},
		{
			name: "both: one denom: less than ratio, more than flat",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("10fig,2honeydew"))
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("14fig"),
			expErr: s.joinErrs(
				"14fig is less than required ratio fee 15fig (based on price 500peach and ratio 100peach:3fig)",
				ratioErr("100peach:3fig,100peach:1grape"),
				insufficientErr("14fig"),
			),
		},
		{
			name: "both: one denom: less than flat, more than ratio",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("10fig,2honeydew"))
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("300peach"),
			fee:      s.coins("9fig"),
			expErr: s.joinErrs(
				"9fig is less than required flat fee 10fig",
				flatErr("10fig,2honeydew"),
				insufficientErr("9fig"),
			),
		},
		{
			name: "both: one denom: more than either, less than total req",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("10fig,2honeydew"))
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("24fig"),
			expErr: s.joinErrs(
				"24fig is less than combined fee 25fig = 10fig (flat) + 15fig (ratio based on price 500peach)",
				insufficientErr("24fig"),
			),
		},
		{
			name: "both: one denom: fee equals total req",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("10fig,2honeydew"))
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("25fig"),
			expErr:   "",
		},
		{
			name: "both: one denom: fee more than total req",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("10fig,2honeydew"))
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("26fig"),
			expErr:   "",
		},
		{
			name: "both: diff denoms: all less than req",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("10fig,6grape"))
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("9fig,4grape,80honeydew"),
			expErr: s.joinErrs(
				"9fig is less than required flat fee 10fig",
				"4grape is less than required flat fee 6grape",
				"no flat fee options available for denom honeydew",
				flatErr("10fig,6grape"),
				"9fig is less than required ratio fee 15fig (based on price 500peach and ratio 100peach:3fig)",
				"4grape is less than required ratio fee 5grape (based on price 500peach and ratio 100peach:1grape)",
				"no ratio from price denom peach to fee denom honeydew",
				ratioErr("100peach:3fig,100peach:1grape"),
				insufficientErr("9fig,4grape,80honeydew"),
			),
		},
		{
			name: "both: diff denoms: flat okay, ratio not",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("10fig,4grape"))
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("10fig,4grape,80honeydew"),
			expErr: s.joinErrs(
				"10fig is less than required ratio fee 15fig (based on price 500peach and ratio 100peach:3fig)",
				"4grape is less than required ratio fee 5grape (based on price 500peach and ratio 100peach:1grape)",
				"no ratio from price denom peach to fee denom honeydew",
				ratioErr("100peach:3fig,100peach:1grape"),
				insufficientErr("10fig,4grape,80honeydew"),
			),
		},
		{
			name: "both: diff denoms: ratio okay, flat not",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("16fig,6grape"))
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("15fig,5grape,80honeydew"),
			expErr: s.joinErrs(
				"15fig is less than required flat fee 16fig",
				"5grape is less than required flat fee 6grape",
				"no flat fee options available for denom honeydew",
				flatErr("16fig,6grape"),
				insufficientErr("15fig,5grape,80honeydew"),
			),
		},
		{
			name: "both: diff denoms: either enough for one fee type, flat first",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("14fig,6grape"))
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("14fig,5grape"),
			expErr:   "",
		},
		{
			name: "both: diff denoms: either enough for one fee type, ratio first",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("16fig,4grape"))
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("15fig,4grape"),
			expErr:   "",
		},
		{
			name: "both: two denoms: first is more than either, less than total, second less than either",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("10fig,4grape"))
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("24fig,3grape"),
			expErr: s.joinErrs(
				"3grape is less than required flat fee 4grape",
				"3grape is less than required ratio fee 5grape (based on price 500peach and ratio 100peach:1grape)",
				"24fig is less than combined fee 25fig = 10fig (flat) + 15fig (ratio based on price 500peach)",
				insufficientErr("24fig,3grape"),
			),
		},
		{
			name: "both: two denoms: first is more than either, less than total, second covers flat",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("10fig,4grape"))
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("24fig,4grape"),
			expErr:   "",
		},
		{
			name: "both: two denoms: first is more than either, less than total, second covers ratio",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("10fig,6grape"))
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("24fig,5grape"),
			expErr:   "",
		},
		{
			name: "both: two denoms: first less than either, second more than either, less than total",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("10fig,6grape"))
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("9fig,10grape"),
			expErr: s.joinErrs(
				"9fig is less than required flat fee 10fig",
				"9fig is less than required ratio fee 15fig (based on price 500peach and ratio 100peach:3fig)",
				"10grape is less than combined fee 11grape = 6grape (flat) + 5grape (ratio based on price 500peach)",
				insufficientErr("9fig,10grape"),
			),
		},
		{
			name: "both: two denoms: first covers flat, second more than either, less than total",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("10fig,6grape"))
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("10fig,10grape"),
			expErr:   "",
		},
		{
			name: "both: two denoms: first covers ratio, second more than either, less than total",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("16fig,6grape"))
				keeper.SetBuyerSettlementRatios(s.getStore(), 2, []exchange.FeeRatio{
					s.ratio("100peach:3fig"), s.ratio("100peach:1grape"),
				})
			},
			marketID: 2,
			price:    s.coin("500peach"),
			fee:      s.coins("15fig,10grape"),
			expErr:   "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var err error
			testFunc := func() {
				err = s.k.ValidateBuyerSettlementFee(s.ctx, tc.marketID, tc.price, tc.fee)
			}
			s.Require().NotPanics(testFunc, "ValidateBuyerSettlementFee(%d, %q, %q)", tc.marketID, tc.price, tc.fee)
			s.assertErrorValue(err, tc.expErr, "ValidateBuyerSettlementFee(%d, %q, %q)", tc.marketID, tc.price, tc.fee)
		})
	}
}

func (s *TestSuite) TestKeeper_UpdateFees() {
	type marketFees struct {
		marketID    uint32
		createAsk   string
		createBid   string
		createCom   string
		sellerFlat  string
		sellerRatio string
		buyerFlat   string
		buyerRatio  string
		comBips     string
	}
	getMarketFees := func(marketID uint32) marketFees {
		rv := marketFees{
			marketID:    marketID,
			createAsk:   sdk.Coins(s.k.GetCreateAskFlatFees(s.ctx, marketID)).String(),
			createBid:   sdk.Coins(s.k.GetCreateBidFlatFees(s.ctx, marketID)).String(),
			createCom:   sdk.Coins(s.k.GetCreateCommitmentFlatFees(s.ctx, marketID)).String(),
			sellerFlat:  sdk.Coins(s.k.GetSellerSettlementFlatFees(s.ctx, marketID)).String(),
			sellerRatio: exchange.FeeRatiosString(s.k.GetSellerSettlementRatios(s.ctx, marketID)),
			buyerFlat:   sdk.Coins(s.k.GetBuyerSettlementFlatFees(s.ctx, marketID)).String(),
			buyerRatio:  exchange.FeeRatiosString(s.k.GetBuyerSettlementRatios(s.ctx, marketID)),
		}
		bips := s.k.GetCommitmentSettlementBips(s.ctx, marketID)
		if bips != 0 {
			rv.comBips = fmt.Sprintf("%d", bips)
		}
		return rv
	}

	tests := []struct {
		name        string
		setup       func()
		msg         *exchange.MsgGovManageFeesRequest
		expFees     marketFees
		expNoChange []uint32
		expPanic    string
	}{
		{
			name:     "nil msg",
			msg:      nil,
			expPanic: "runtime error: invalid memory address or nil pointer dereference",
		},

		// Only create-ask flat fee changes.
		{
			name: "create ask: add one",
			setup: func() {
				keeper.SetCreateAskFlatFees(s.getStore(), 3, s.coins("10fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, AddFeeCreateAskFlat: s.coins("10fig")},
			expFees:     marketFees{marketID: 5, createAsk: "10fig"},
			expNoChange: []uint32{3},
		},
		{
			name: "create ask: remove one, exists",
			setup: func() {
				store := s.getStore()
				keeper.SetCreateAskFlatFees(store, 3, s.coins("10fig"))
				keeper.SetCreateAskFlatFees(store, 5, s.coins("10fig,8grape"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, RemoveFeeCreateAskFlat: s.coins("10fig")},
			expFees:     marketFees{marketID: 5, createAsk: "8grape"},
			expNoChange: []uint32{3},
		},
		{
			name: "create ask: remove one, unknown denom",
			setup: func() {
				store := s.getStore()
				keeper.SetCreateAskFlatFees(store, 3, s.coins("10grape"))
				keeper.SetCreateAskFlatFees(store, 5, s.coins("10fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, RemoveFeeCreateAskFlat: s.coins("10grape")},
			expFees:     marketFees{marketID: 5, createAsk: "10fig"},
			expNoChange: []uint32{3},
		},
		{
			name: "create ask: remove one, known denom, wrong amount",
			setup: func() {
				store := s.getStore()
				keeper.SetCreateAskFlatFees(store, 4, s.coins("10grape"))
				keeper.SetCreateAskFlatFees(store, 2, s.coins("10fig,8grape"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 2, RemoveFeeCreateAskFlat: s.coins("8fig")},
			expFees:     marketFees{marketID: 2, createAsk: "8grape"},
			expNoChange: []uint32{4},
		},
		{
			name: "create ask: add+remove with different denoms",
			setup: func() {
				store := s.getStore()
				keeper.SetCreateAskFlatFees(store, 18, s.coins("10grape"))
				keeper.SetCreateAskFlatFees(store, 8, s.coins("10fig,8grape"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:               8,
				RemoveFeeCreateAskFlat: s.coins("8grape"),
				AddFeeCreateAskFlat:    s.coins("2honeydew"),
			},
			expFees:     marketFees{marketID: 8, createAsk: "10fig,2honeydew"},
			expNoChange: []uint32{18},
		},
		{
			name: "create ask: add+remove with same denoms",
			setup: func() {
				store := s.getStore()
				keeper.SetCreateAskFlatFees(store, 18, s.coins("10grape"))
				keeper.SetCreateAskFlatFees(store, 1, s.coins("10fig,8grape"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:               1,
				RemoveFeeCreateAskFlat: s.coins("10fig"),
				AddFeeCreateAskFlat:    s.coins("7fig"),
			},
			expFees:     marketFees{marketID: 1, createAsk: "7fig,8grape"},
			expNoChange: []uint32{18},
		},
		{
			name: "create ask: complex",
			// Remove one with wrong amount and don't replace it (10fig)
			// Remove one with correct amount and replace it with another amount (7honeydew -> 5honeydew).
			// Add one with a denom that already has a different amount (3cactus stomping on 7cactus)
			// Add a brand new one (99plum).
			// Leave one unchanged (2grape).
			setup: func() {
				store := s.getStore()
				keeper.SetCreateAskFlatFees(store, 1, s.coins("10fig,4grape,2honeydew,5apple"))
				keeper.SetCreateAskFlatFees(store, 2, s.coins("9fig,3grape,1honeydew,6banana"))
				keeper.SetCreateAskFlatFees(store, 3, s.coins("12fig,2grape,7honeydew,7cactus"))
				keeper.SetCreateAskFlatFees(store, 4, s.coins("25fig,1grape,3honeydew,8durian"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:               3,
				RemoveFeeCreateAskFlat: s.coins("10fig,7honeydew"),
				AddFeeCreateAskFlat:    s.coins("5honeydew,3cactus,99plum"),
			},
			expFees:     marketFees{marketID: 3, createAsk: "3cactus,2grape,5honeydew,99plum"},
			expNoChange: []uint32{1, 2, 4},
		},

		// Only create-bid flat fee changes.
		{
			name: "create bid: add one",
			setup: func() {
				keeper.SetCreateBidFlatFees(s.getStore(), 3, s.coins("10fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, AddFeeCreateBidFlat: s.coins("10fig")},
			expFees:     marketFees{marketID: 5, createBid: "10fig"},
			expNoChange: []uint32{3},
		},
		{
			name: "create bid: remove one, exists",
			setup: func() {
				store := s.getStore()
				keeper.SetCreateBidFlatFees(store, 3, s.coins("10fig"))
				keeper.SetCreateBidFlatFees(store, 5, s.coins("10fig,8grape"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, RemoveFeeCreateBidFlat: s.coins("10fig")},
			expFees:     marketFees{marketID: 5, createBid: "8grape"},
			expNoChange: []uint32{3},
		},
		{
			name: "create bid: remove one, unknown denom",
			setup: func() {
				store := s.getStore()
				keeper.SetCreateBidFlatFees(store, 3, s.coins("10grape"))
				keeper.SetCreateBidFlatFees(store, 5, s.coins("10fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, RemoveFeeCreateBidFlat: s.coins("10grape")},
			expFees:     marketFees{marketID: 5, createBid: "10fig"},
			expNoChange: []uint32{3},
		},
		{
			name: "create bid: remove one, known denom, wrong amount",
			setup: func() {
				store := s.getStore()
				keeper.SetCreateBidFlatFees(store, 4, s.coins("10grape"))
				keeper.SetCreateBidFlatFees(store, 2, s.coins("10fig,8grape"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 2, RemoveFeeCreateBidFlat: s.coins("8fig")},
			expFees:     marketFees{marketID: 2, createBid: "8grape"},
			expNoChange: []uint32{4},
		},
		{
			name: "create bid: add+remove with different denoms",
			setup: func() {
				store := s.getStore()
				keeper.SetCreateBidFlatFees(store, 18, s.coins("10grape"))
				keeper.SetCreateBidFlatFees(store, 8, s.coins("10fig,8grape"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:               8,
				RemoveFeeCreateBidFlat: s.coins("8grape"),
				AddFeeCreateBidFlat:    s.coins("2honeydew"),
			},
			expFees:     marketFees{marketID: 8, createBid: "10fig,2honeydew"},
			expNoChange: []uint32{18},
		},
		{
			name: "create bid: add+remove with same denoms",
			setup: func() {
				store := s.getStore()
				keeper.SetCreateBidFlatFees(store, 18, s.coins("10grape"))
				keeper.SetCreateBidFlatFees(store, 1, s.coins("10fig,8grape"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:               1,
				RemoveFeeCreateBidFlat: s.coins("10fig"),
				AddFeeCreateBidFlat:    s.coins("7fig"),
			},
			expFees:     marketFees{marketID: 1, createBid: "7fig,8grape"},
			expNoChange: []uint32{18},
		},
		{
			name: "create bid: complex",
			// Remove one with wrong amount and don't replace it (10fig)
			// Remove one with correct amount and replace it with another amount (7honeydew -> 5honeydew).
			// Add one with a denom that already has a different amount (3cactus stomping on 7cactus)
			// Add a brand new one (99plum).
			// Leave one unchanged (2grape).
			setup: func() {
				store := s.getStore()
				keeper.SetCreateBidFlatFees(store, 1, s.coins("10fig,4grape,2honeydew,5apple"))
				keeper.SetCreateBidFlatFees(store, 2, s.coins("9fig,3grape,1honeydew,6banana"))
				keeper.SetCreateBidFlatFees(store, 3, s.coins("12fig,2grape,7honeydew,7cactus"))
				keeper.SetCreateBidFlatFees(store, 4, s.coins("25fig,1grape,3honeydew,8durian"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:               3,
				RemoveFeeCreateBidFlat: s.coins("10fig,7honeydew"),
				AddFeeCreateBidFlat:    s.coins("5honeydew,3cactus,99plum"),
			},
			expFees:     marketFees{marketID: 3, createBid: "3cactus,2grape,5honeydew,99plum"},
			expNoChange: []uint32{1, 2, 4},
		},

		// Only create-commitment flat fee changes.
		{
			name: "create commitment: add one",
			setup: func() {
				keeper.SetCreateCommitmentFlatFees(s.getStore(), 3, s.coins("10fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, AddFeeCreateCommitmentFlat: s.coins("10fig")},
			expFees:     marketFees{marketID: 5, createCom: "10fig"},
			expNoChange: []uint32{3},
		},
		{
			name: "create commitment: remove one, exists",
			setup: func() {
				store := s.getStore()
				keeper.SetCreateCommitmentFlatFees(store, 3, s.coins("10fig"))
				keeper.SetCreateCommitmentFlatFees(store, 5, s.coins("10fig,8grape"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, RemoveFeeCreateCommitmentFlat: s.coins("10fig")},
			expFees:     marketFees{marketID: 5, createCom: "8grape"},
			expNoChange: []uint32{3},
		},
		{
			name: "create commitment: remove one, unknown denom",
			setup: func() {
				store := s.getStore()
				keeper.SetCreateCommitmentFlatFees(store, 3, s.coins("10grape"))
				keeper.SetCreateCommitmentFlatFees(store, 5, s.coins("10fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, RemoveFeeCreateCommitmentFlat: s.coins("10grape")},
			expFees:     marketFees{marketID: 5, createCom: "10fig"},
			expNoChange: []uint32{3},
		},
		{
			name: "create commitment: remove one, known denom, wrong amount",
			setup: func() {
				store := s.getStore()
				keeper.SetCreateCommitmentFlatFees(store, 4, s.coins("10grape"))
				keeper.SetCreateCommitmentFlatFees(store, 2, s.coins("10fig,8grape"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 2, RemoveFeeCreateCommitmentFlat: s.coins("8fig")},
			expFees:     marketFees{marketID: 2, createCom: "8grape"},
			expNoChange: []uint32{4},
		},
		{
			name: "create commitment: add+remove with different denoms",
			setup: func() {
				store := s.getStore()
				keeper.SetCreateCommitmentFlatFees(store, 18, s.coins("10grape"))
				keeper.SetCreateCommitmentFlatFees(store, 8, s.coins("10fig,8grape"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                      8,
				RemoveFeeCreateCommitmentFlat: s.coins("8grape"),
				AddFeeCreateCommitmentFlat:    s.coins("2honeydew"),
			},
			expFees:     marketFees{marketID: 8, createCom: "10fig,2honeydew"},
			expNoChange: []uint32{18},
		},
		{
			name: "create commitment: add+remove with same denoms",
			setup: func() {
				store := s.getStore()
				keeper.SetCreateCommitmentFlatFees(store, 18, s.coins("10grape"))
				keeper.SetCreateCommitmentFlatFees(store, 1, s.coins("10fig,8grape"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                      1,
				RemoveFeeCreateCommitmentFlat: s.coins("10fig"),
				AddFeeCreateCommitmentFlat:    s.coins("7fig"),
			},
			expFees:     marketFees{marketID: 1, createCom: "7fig,8grape"},
			expNoChange: []uint32{18},
		},
		{
			name: "create commitment: complex",
			// Remove one with wrong amount and don't replace it (10fig)
			// Remove one with correct amount and replace it with another amount (7honeydew -> 5honeydew).
			// Add one with a denom that already has a different amount (3cactus stomping on 7cactus)
			// Add a brand new one (99plum).
			// Leave one unchanged (2grape).
			setup: func() {
				store := s.getStore()
				keeper.SetCreateCommitmentFlatFees(store, 1, s.coins("10fig,4grape,2honeydew,5apple"))
				keeper.SetCreateCommitmentFlatFees(store, 2, s.coins("9fig,3grape,1honeydew,6banana"))
				keeper.SetCreateCommitmentFlatFees(store, 3, s.coins("12fig,2grape,7honeydew,7cactus"))
				keeper.SetCreateCommitmentFlatFees(store, 4, s.coins("25fig,1grape,3honeydew,8durian"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                      3,
				RemoveFeeCreateCommitmentFlat: s.coins("10fig,7honeydew"),
				AddFeeCreateCommitmentFlat:    s.coins("5honeydew,3cactus,99plum"),
			},
			expFees:     marketFees{marketID: 3, createCom: "3cactus,2grape,5honeydew,99plum"},
			expNoChange: []uint32{1, 2, 4},
		},

		// Only seller settlement flat fee changes.
		{
			name: "seller flat: add one",
			setup: func() {
				keeper.SetSellerSettlementFlatFees(s.getStore(), 3, s.coins("10fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, AddFeeSellerSettlementFlat: s.coins("10fig")},
			expFees:     marketFees{marketID: 5, sellerFlat: "10fig"},
			expNoChange: []uint32{3},
		},
		{
			name: "seller flat: remove one, exists",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementFlatFees(store, 3, s.coins("10fig"))
				keeper.SetSellerSettlementFlatFees(store, 5, s.coins("10fig,8grape"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, RemoveFeeSellerSettlementFlat: s.coins("10fig")},
			expFees:     marketFees{marketID: 5, sellerFlat: "8grape"},
			expNoChange: []uint32{3},
		},
		{
			name: "seller flat: remove one, unknown denom",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementFlatFees(store, 3, s.coins("10grape"))
				keeper.SetSellerSettlementFlatFees(store, 5, s.coins("10fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, RemoveFeeSellerSettlementFlat: s.coins("10grape")},
			expFees:     marketFees{marketID: 5, sellerFlat: "10fig"},
			expNoChange: []uint32{3},
		},
		{
			name: "seller flat: remove one, known denom, wrong amount",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementFlatFees(store, 4, s.coins("10grape"))
				keeper.SetSellerSettlementFlatFees(store, 2, s.coins("10fig,8grape"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 2, RemoveFeeSellerSettlementFlat: s.coins("8fig")},
			expFees:     marketFees{marketID: 2, sellerFlat: "8grape"},
			expNoChange: []uint32{4},
		},
		{
			name: "seller flat: add+remove with different denoms",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementFlatFees(store, 18, s.coins("10grape"))
				keeper.SetSellerSettlementFlatFees(store, 8, s.coins("10fig,8grape"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                      8,
				RemoveFeeSellerSettlementFlat: s.coins("8grape"),
				AddFeeSellerSettlementFlat:    s.coins("2honeydew"),
			},
			expFees:     marketFees{marketID: 8, sellerFlat: "10fig,2honeydew"},
			expNoChange: []uint32{18},
		},
		{
			name: "seller flat: add+remove with same denoms",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementFlatFees(store, 18, s.coins("10grape"))
				keeper.SetSellerSettlementFlatFees(store, 1, s.coins("10fig,8grape"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                      1,
				RemoveFeeSellerSettlementFlat: s.coins("10fig"),
				AddFeeSellerSettlementFlat:    s.coins("7fig"),
			},
			expFees:     marketFees{marketID: 1, sellerFlat: "7fig,8grape"},
			expNoChange: []uint32{18},
		},
		{
			name: "seller flat: complex",
			// Remove one with wrong amount and don't replace it (10fig)
			// Remove one with correct amount and replace it with another amount (7honeydew -> 5honeydew).
			// Add one with a denom that already has a different amount (3cactus stomping on 7cactus)
			// Add a brand new one (99plum).
			// Leave one unchanged (2grape).
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementFlatFees(store, 1, s.coins("10fig,4grape,2honeydew,5apple"))
				keeper.SetSellerSettlementFlatFees(store, 2, s.coins("9fig,3grape,1honeydew,6banana"))
				keeper.SetSellerSettlementFlatFees(store, 3, s.coins("12fig,2grape,7honeydew,7cactus"))
				keeper.SetSellerSettlementFlatFees(store, 4, s.coins("25fig,1grape,3honeydew,8durian"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                      3,
				RemoveFeeSellerSettlementFlat: s.coins("10fig,7honeydew"),
				AddFeeSellerSettlementFlat:    s.coins("5honeydew,3cactus,99plum"),
			},
			expFees:     marketFees{marketID: 3, sellerFlat: "3cactus,2grape,5honeydew,99plum"},
			expNoChange: []uint32{1, 2, 4},
		},

		// Only buyer settlement flat fee changes.
		{
			name: "buyer flat: add one",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 3, s.coins("10fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, AddFeeBuyerSettlementFlat: s.coins("10fig")},
			expFees:     marketFees{marketID: 5, buyerFlat: "10fig"},
			expNoChange: []uint32{3},
		},
		{
			name: "buyer flat: remove one, exists",
			setup: func() {
				store := s.getStore()
				keeper.SetBuyerSettlementFlatFees(store, 3, s.coins("10fig"))
				keeper.SetBuyerSettlementFlatFees(store, 5, s.coins("10fig,8grape"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, RemoveFeeBuyerSettlementFlat: s.coins("10fig")},
			expFees:     marketFees{marketID: 5, buyerFlat: "8grape"},
			expNoChange: []uint32{3},
		},
		{
			name: "buyer flat: remove one, unknown denom",
			setup: func() {
				store := s.getStore()
				keeper.SetBuyerSettlementFlatFees(store, 3, s.coins("10grape"))
				keeper.SetBuyerSettlementFlatFees(store, 5, s.coins("10fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, RemoveFeeBuyerSettlementFlat: s.coins("10grape")},
			expFees:     marketFees{marketID: 5, buyerFlat: "10fig"},
			expNoChange: []uint32{3},
		},
		{
			name: "buyer flat: remove one, known denom, wrong amount",
			setup: func() {
				store := s.getStore()
				keeper.SetBuyerSettlementFlatFees(store, 4, s.coins("10grape"))
				keeper.SetBuyerSettlementFlatFees(store, 2, s.coins("10fig,8grape"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 2, RemoveFeeBuyerSettlementFlat: s.coins("8fig")},
			expFees:     marketFees{marketID: 2, buyerFlat: "8grape"},
			expNoChange: []uint32{4},
		},
		{
			name: "buyer flat: add+remove with different denoms",
			setup: func() {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 18, s.coins("10grape"))
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 8, s.coins("10fig,8grape"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                     8,
				RemoveFeeBuyerSettlementFlat: s.coins("8grape"),
				AddFeeBuyerSettlementFlat:    s.coins("2honeydew"),
			},
			expFees:     marketFees{marketID: 8, buyerFlat: "10fig,2honeydew"},
			expNoChange: []uint32{18},
		},
		{
			name: "buyer flat: add+remove with same denoms",
			setup: func() {
				store := s.getStore()
				keeper.SetBuyerSettlementFlatFees(store, 18, s.coins("10grape"))
				keeper.SetBuyerSettlementFlatFees(store, 1, s.coins("10fig,8grape"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                     1,
				RemoveFeeBuyerSettlementFlat: s.coins("10fig"),
				AddFeeBuyerSettlementFlat:    s.coins("7fig"),
			},
			expFees:     marketFees{marketID: 1, buyerFlat: "7fig,8grape"},
			expNoChange: []uint32{18},
		},
		{
			name: "buyer flat: complex",
			// Remove one with wrong amount and don't replace it (10fig)
			// Remove one with correct amount and replace it with another amount (7honeydew -> 5honeydew).
			// Add one with a denom that already has a different amount (3cactus stomping on 7cactus)
			// Add a brand new one (99plum).
			// Leave one unchanged (2grape).
			setup: func() {
				store := s.getStore()
				keeper.SetBuyerSettlementFlatFees(store, 1, s.coins("10fig,4grape,2honeydew,5apple"))
				keeper.SetBuyerSettlementFlatFees(store, 2, s.coins("9fig,3grape,1honeydew,6banana"))
				keeper.SetBuyerSettlementFlatFees(store, 3, s.coins("12fig,2grape,7honeydew,7cactus"))
				keeper.SetBuyerSettlementFlatFees(store, 4, s.coins("25fig,1grape,3honeydew,8durian"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                     3,
				RemoveFeeBuyerSettlementFlat: s.coins("10fig,7honeydew"),
				AddFeeBuyerSettlementFlat:    s.coins("5honeydew,3cactus,99plum"),
			},
			expFees:     marketFees{marketID: 3, buyerFlat: "3cactus,2grape,5honeydew,99plum"},
			expNoChange: []uint32{1, 2, 4},
		},

		// Only seller settlement ratio fee changes.
		{
			name: "seller ratio: add one",
			setup: func() {
				keeper.SetSellerSettlementRatios(s.getStore(), 3, s.ratios("100peach:3fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, AddFeeSellerSettlementRatios: s.ratios("50plum:1grape")},
			expFees:     marketFees{marketID: 5, sellerRatio: "50plum:1grape"},
			expNoChange: []uint32{3},
		},
		{
			name: "seller ratio: remove one, exists",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 3, s.ratios("100peach:3fig"))
				keeper.SetSellerSettlementRatios(store, 5, s.ratios("90peach:2fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, RemoveFeeSellerSettlementRatios: s.ratios("90peach:2fig")},
			expFees:     marketFees{marketID: 5},
			expNoChange: []uint32{3},
		},
		{
			name: "seller ratio: remove one, unknown denoms",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 3, s.ratios("100peach:3fig"))
				keeper.SetSellerSettlementRatios(store, 5, s.ratios("90peach:2fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, RemoveFeeSellerSettlementRatios: s.ratios("90plum:2grape")},
			expFees:     marketFees{marketID: 5, sellerRatio: "90peach:2fig"},
			expNoChange: []uint32{3},
		},
		{
			name: "seller ratio: remove one, known price denom",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 3, s.ratios("100peach:3fig"))
				keeper.SetSellerSettlementRatios(store, 5, s.ratios("90peach:2fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, RemoveFeeSellerSettlementRatios: s.ratios("90peach:2grape")},
			expFees:     marketFees{marketID: 5, sellerRatio: "90peach:2fig"},
			expNoChange: []uint32{3},
		},
		{
			name: "seller ratio: remove one, known fee denom",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 4, s.ratios("100peach:3fig"))
				keeper.SetSellerSettlementRatios(store, 2, s.ratios("90peach:2fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 2, RemoveFeeSellerSettlementRatios: s.ratios("90plum:2fig")},
			expFees:     marketFees{marketID: 2, sellerRatio: "90peach:2fig"},
			expNoChange: []uint32{4},
		},
		{
			name: "seller ratio: remove one, known denoms, wrong amounts",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 4, s.ratios("100peach:3fig"))
				keeper.SetSellerSettlementRatios(store, 2, s.ratios("90peach:2fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 2, RemoveFeeSellerSettlementRatios: s.ratios("89peach:3fig")},
			expFees:     marketFees{marketID: 2},
			expNoChange: []uint32{4},
		},
		{
			name: "seller ratio: add+remove with different denoms",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 7, s.ratios("100peach:3fig,100peach:1grape"))
				keeper.SetSellerSettlementRatios(store, 77, s.ratios("100peach:3fig,100peach:1grape"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                        7,
				RemoveFeeSellerSettlementRatios: s.ratios("100peach:3fig"),
				AddFeeSellerSettlementRatios:    s.ratios("100plum:3honeydew"),
			},
			expFees:     marketFees{marketID: 7, sellerRatio: "100peach:1grape,100plum:3honeydew"},
			expNoChange: []uint32{77},
		},
		{
			name: "seller ratio: add+remove with different price denom",
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 7, s.ratios("100peach:3fig,100peach:1grape"))
				keeper.SetSellerSettlementRatios(store, 77, s.ratios("100peach:3fig,100peach:1grape"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                        77,
				RemoveFeeSellerSettlementRatios: s.ratios("100peach:3fig"),
				AddFeeSellerSettlementRatios:    s.ratios("100plum:3fig"),
			},
			expFees:     marketFees{marketID: 77, sellerRatio: "100peach:1grape,100plum:3fig"},
			expNoChange: []uint32{7},
		},
		{
			name: "seller ratio: add+remove with different fee denom",
			setup: func() {
				keeper.SetSellerSettlementRatios(s.getStore(), 1, s.ratios("100peach:3fig,100peach:1grape"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                        1,
				RemoveFeeSellerSettlementRatios: s.ratios("100peach:3fig"),
				AddFeeSellerSettlementRatios:    s.ratios("100peach:2honeydew"),
			},
			expFees: marketFees{marketID: 1, sellerRatio: "100peach:1grape,100peach:2honeydew"},
		},
		{
			name: "seller ratio: add+remove with same denoms",
			setup: func() {
				keeper.SetSellerSettlementRatios(s.getStore(), 1, s.ratios("100peach:3fig,100peach:1grape"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                        1,
				RemoveFeeSellerSettlementRatios: s.ratios("100peach:3fig"),
				AddFeeSellerSettlementRatios:    s.ratios("90peach:2fig"),
			},
			expFees: marketFees{marketID: 1, sellerRatio: "90peach:2fig,100peach:1grape"},
		},
		{
			name: "seller ratio: complex",
			// Remove one with wrong amounts and don't replace it (10plum:3fig)
			// Remove and replace one to change amounts (100peach:3fig -> 90peach:2fig)
			// Add one with existing denoms and different amounts (110peach:2grape stomping on 100peach:1grape)
			// Add one with same price denom (100peach:1honeydew)
			// Add one with same fee denom (100pear:3fig)
			// Add one all new (100papaya:5guava)
			// Leave on untouched (100prune:2fig)
			setup: func() {
				store := s.getStore()
				keeper.SetSellerSettlementRatios(store, 1, s.ratios("100peach:3fig,100peach:1grape"))
				keeper.SetSellerSettlementRatios(store, 11, s.ratios("20plum:2fig,100peach:3fig,100peach:1grape,100prune:2fig"))
				keeper.SetSellerSettlementRatios(store, 111, s.ratios("100peach:3fig,100peach:1grape"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                        11,
				RemoveFeeSellerSettlementRatios: s.ratios("10plum:3fig,100peach:3fig"),
				AddFeeSellerSettlementRatios:    s.ratios("90peach:2fig,110peach:2grape,100peach:1honeydew,100pear:3fig,100papaya:5guava"),
			},
			expFees: marketFees{
				marketID:    11,
				sellerRatio: "100papaya:5guava,90peach:2fig,110peach:2grape,100peach:1honeydew,100pear:3fig,100prune:2fig",
			},
			expNoChange: nil,
			expPanic:    "",
		},

		// Only buyer settlement ratio fee changes.
		{
			name: "buyer ratio: add one",
			setup: func() {
				keeper.SetBuyerSettlementRatios(s.getStore(), 3, s.ratios("100peach:3fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, AddFeeBuyerSettlementRatios: s.ratios("50plum:1grape")},
			expFees:     marketFees{marketID: 5, buyerRatio: "50plum:1grape"},
			expNoChange: []uint32{3},
		},
		{
			name: "buyer ratio: remove one, exists",
			setup: func() {
				store := s.getStore()
				keeper.SetBuyerSettlementRatios(store, 3, s.ratios("100peach:3fig"))
				keeper.SetBuyerSettlementRatios(store, 5, s.ratios("90peach:2fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, RemoveFeeBuyerSettlementRatios: s.ratios("90peach:2fig")},
			expFees:     marketFees{marketID: 5},
			expNoChange: []uint32{3},
		},
		{
			name: "buyer ratio: remove one, unknown denoms",
			setup: func() {
				store := s.getStore()
				keeper.SetBuyerSettlementRatios(store, 3, s.ratios("100peach:3fig"))
				keeper.SetBuyerSettlementRatios(store, 5, s.ratios("90peach:2fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, RemoveFeeBuyerSettlementRatios: s.ratios("90plum:2grape")},
			expFees:     marketFees{marketID: 5, buyerRatio: "90peach:2fig"},
			expNoChange: []uint32{3},
		},
		{
			name: "buyer ratio: remove one, known price denom",
			setup: func() {
				store := s.getStore()
				keeper.SetBuyerSettlementRatios(store, 3, s.ratios("100peach:3fig"))
				keeper.SetBuyerSettlementRatios(store, 5, s.ratios("90peach:2fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 5, RemoveFeeBuyerSettlementRatios: s.ratios("90peach:2grape")},
			expFees:     marketFees{marketID: 5, buyerRatio: "90peach:2fig"},
			expNoChange: []uint32{3},
		},
		{
			name: "buyer ratio: remove one, known fee denom",
			setup: func() {
				store := s.getStore()
				keeper.SetBuyerSettlementRatios(store, 4, s.ratios("100peach:3fig"))
				keeper.SetBuyerSettlementRatios(store, 2, s.ratios("90peach:2fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 2, RemoveFeeBuyerSettlementRatios: s.ratios("90plum:2fig")},
			expFees:     marketFees{marketID: 2, buyerRatio: "90peach:2fig"},
			expNoChange: []uint32{4},
		},
		{
			name: "buyer ratio: remove one, known denoms, wrong amounts",
			setup: func() {
				store := s.getStore()
				keeper.SetBuyerSettlementRatios(store, 4, s.ratios("100peach:3fig"))
				keeper.SetBuyerSettlementRatios(store, 2, s.ratios("90peach:2fig"))
			},
			msg:         &exchange.MsgGovManageFeesRequest{MarketId: 2, RemoveFeeBuyerSettlementRatios: s.ratios("89peach:3fig")},
			expFees:     marketFees{marketID: 2},
			expNoChange: []uint32{4},
		},
		{
			name: "buyer ratio: add+remove with different denoms",
			setup: func() {
				store := s.getStore()
				keeper.SetBuyerSettlementRatios(store, 7, s.ratios("100peach:3fig,100peach:1grape"))
				keeper.SetBuyerSettlementRatios(store, 77, s.ratios("100peach:3fig,100peach:1grape"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                       7,
				RemoveFeeBuyerSettlementRatios: s.ratios("100peach:3fig"),
				AddFeeBuyerSettlementRatios:    s.ratios("100plum:3honeydew"),
			},
			expFees:     marketFees{marketID: 7, buyerRatio: "100peach:1grape,100plum:3honeydew"},
			expNoChange: []uint32{77},
		},
		{
			name: "buyer ratio: add+remove with different price denom",
			setup: func() {
				store := s.getStore()
				keeper.SetBuyerSettlementRatios(store, 7, s.ratios("100peach:3fig,100peach:1grape"))
				keeper.SetBuyerSettlementRatios(store, 77, s.ratios("100peach:3fig,100peach:1grape"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                       77,
				RemoveFeeBuyerSettlementRatios: s.ratios("100peach:3fig"),
				AddFeeBuyerSettlementRatios:    s.ratios("100plum:3fig"),
			},
			expFees:     marketFees{marketID: 77, buyerRatio: "100peach:1grape,100plum:3fig"},
			expNoChange: []uint32{7},
		},
		{
			name: "buyer ratio: add+remove with different fee denom",
			setup: func() {
				keeper.SetBuyerSettlementRatios(s.getStore(), 1, s.ratios("100peach:3fig,100peach:1grape"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                       1,
				RemoveFeeBuyerSettlementRatios: s.ratios("100peach:3fig"),
				AddFeeBuyerSettlementRatios:    s.ratios("100peach:2honeydew"),
			},
			expFees: marketFees{marketID: 1, buyerRatio: "100peach:1grape,100peach:2honeydew"},
		},
		{
			name: "buyer ratio: add+remove with same denoms",
			setup: func() {
				keeper.SetBuyerSettlementRatios(s.getStore(), 1, s.ratios("100peach:3fig,100peach:1grape"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                       1,
				RemoveFeeBuyerSettlementRatios: s.ratios("100peach:3fig"),
				AddFeeBuyerSettlementRatios:    s.ratios("90peach:2fig"),
			},
			expFees: marketFees{marketID: 1, buyerRatio: "90peach:2fig,100peach:1grape"},
		},
		{
			name: "buyer ratio: complex",
			// Remove one with wrong amounts and don't replace it (10plum:3fig)
			// Remove and replace one to change amounts (100peach:3fig -> 90peach:2fig)
			// Add one with existing denoms and different amounts (110peach:2grape stomping on 100peach:1grape)
			// Add one with same price denom (100peach:1honeydew)
			// Add one with same fee denom (100pear:3fig)
			// Add one all new (100papaya:5guava)
			// Leave one untouched (100prune:2fig)
			setup: func() {
				store := s.getStore()
				keeper.SetBuyerSettlementRatios(store, 1, s.ratios("100peach:3fig,100peach:1grape"))
				keeper.SetBuyerSettlementRatios(store, 11, s.ratios("20plum:2fig,100peach:3fig,100peach:1grape,100prune:2fig"))
				keeper.SetBuyerSettlementRatios(store, 111, s.ratios("100peach:3fig,100peach:1grape"))
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                       11,
				RemoveFeeBuyerSettlementRatios: s.ratios("10plum:3fig,100peach:3fig"),
				AddFeeBuyerSettlementRatios:    s.ratios("90peach:2fig,110peach:2grape,100peach:1honeydew,100pear:3fig,100papaya:5guava"),
			},
			expFees: marketFees{
				marketID:   11,
				buyerRatio: "100papaya:5guava,90peach:2fig,110peach:2grape,100peach:1honeydew,100pear:3fig,100prune:2fig",
			},
			expNoChange: nil,
			expPanic:    "",
		},

		// only commitment settlement bips
		{
			name: "not set yet: setting",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentSettlementBips(store, 1, 100)
				keeper.SetCommitmentSettlementBips(store, 2, 200)
				keeper.SetCommitmentSettlementBips(store, 4, 400)
				keeper.SetCommitmentSettlementBips(store, 5, 500)
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                       3,
				SetFeeCommitmentSettlementBips: 25,
			},
			expFees: marketFees{
				marketID: 3,
				comBips:  "25",
			},
			expNoChange: []uint32{1, 2, 4, 5},
		},
		{
			name: "not set yet: unsetting",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentSettlementBips(store, 1, 100)
				keeper.SetCommitmentSettlementBips(store, 3, 300)
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                         2,
				UnsetFeeCommitmentSettlementBips: true,
			},
			expFees:     marketFees{marketID: 2},
			expNoChange: []uint32{1, 3},
		},
		{
			name: "already set: setting",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentSettlementBips(store, 1, 100)
				keeper.SetCommitmentSettlementBips(store, 2, 200)
				keeper.SetCommitmentSettlementBips(store, 3, 300)
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                       2,
				SetFeeCommitmentSettlementBips: 25,
			},
			expFees: marketFees{
				marketID: 2,
				comBips:  "25",
			},
			expNoChange: []uint32{1, 3},
		},
		{
			name: "already set: unsetting",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentSettlementBips(store, 1, 100)
				keeper.SetCommitmentSettlementBips(store, 2, 200)
				keeper.SetCommitmentSettlementBips(store, 3, 300)
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                         2,
				UnsetFeeCommitmentSettlementBips: true,
			},
			expFees:     marketFees{marketID: 2},
			expNoChange: []uint32{1, 3},
		},

		// combo
		{
			name: "a little bit of everything",
			// For each type, add one, replace one, remove one, leave one.
			setup: func() {
				store := s.getStore()
				keeper.SetCreateAskFlatFees(store, 1, s.coins("10grape"))
				keeper.SetCreateBidFlatFees(store, 1, s.coins("11guava"))
				keeper.SetCreateCommitmentFlatFees(store, 1, s.coins("14fig"))
				keeper.SetSellerSettlementFlatFees(store, 1, s.coins("12grapefruit"))
				keeper.SetBuyerSettlementFlatFees(store, 1, s.coins("13gooseberry"))
				keeper.SetSellerSettlementRatios(store, 1, s.ratios("100papaya:3goumi"))
				keeper.SetBuyerSettlementRatios(store, 1, s.ratios("120pineapple:1guarana"))
				keeper.SetCommitmentSettlementBips(store, 1, 100)

				keeper.SetCreateAskFlatFees(store, 2, s.coins("201acai,202apple,203apricot"))
				keeper.SetCreateBidFlatFees(store, 2, s.coins("211banana,212biriba,212blueberry"))
				keeper.SetCreateCommitmentFlatFees(store, 2, s.coins("261raisin,262raspberry,263redcurrent"))
				keeper.SetSellerSettlementFlatFees(store, 2, s.coins("221cactus,222cantaloupe,223cherry"))
				keeper.SetBuyerSettlementFlatFees(store, 2, s.coins("231date,232dewberry,233durian"))
				keeper.SetSellerSettlementRatios(store, 2, s.ratios("241tangerine:1lemon,242tangerine:2lime,243tayberry:3lime"))
				keeper.SetBuyerSettlementRatios(store, 2, s.ratios("251mandarin:4nectarine,252mango:5nectarine,253mango:6nutmeg"))
				keeper.SetCommitmentSettlementBips(store, 2, 200)

				keeper.SetCreateAskFlatFees(store, 3, s.coins("30grape"))
				keeper.SetCreateBidFlatFees(store, 3, s.coins("31guava"))
				keeper.SetCreateCommitmentFlatFees(store, 3, s.coins("31fig"))
				keeper.SetSellerSettlementFlatFees(store, 3, s.coins("32grapefruit"))
				keeper.SetBuyerSettlementFlatFees(store, 3, s.coins("33gooseberry"))
				keeper.SetSellerSettlementRatios(store, 3, s.ratios("300papaya:3goumi"))
				keeper.SetBuyerSettlementRatios(store, 3, s.ratios("320pineapple:1guarana"))
				keeper.SetCommitmentSettlementBips(store, 3, 300)
			},
			msg: &exchange.MsgGovManageFeesRequest{
				MarketId:                        2,
				AddFeeCreateAskFlat:             s.coins("2002apple,204avocado"),
				RemoveFeeCreateAskFlat:          s.coins("202apple,203apricot"),
				AddFeeCreateBidFlat:             s.coins("214barbadine,2102blueberry"),
				RemoveFeeCreateBidFlat:          s.coins("211banana,212blueberry"),
				AddFeeCreateCommitmentFlat:      s.coins("2620raspberry,264rata"),
				RemoveFeeCreateCommitmentFlat:   s.coins("262raspberry,263redcurrent"),
				AddFeeSellerSettlementFlat:      s.coins("224cassaba,2201cactus"),
				RemoveFeeSellerSettlementFlat:   s.coins("221cactus,222cantaloupe"),
				AddFeeBuyerSettlementFlat:       s.coins("2302dewberry,234dragonfruit"),
				RemoveFeeBuyerSettlementFlat:    s.coins("231date,232dewberry"),
				AddFeeSellerSettlementRatios:    s.ratios("2402tangerine:20lime,244tamarillo:4lemon"),
				RemoveFeeSellerSettlementRatios: s.ratios("241tangerine:1lemon,242tangerine:2lime"),
				AddFeeBuyerSettlementRatios:     s.ratios("2502mango:50nectarine,254marula:7neem"),
				RemoveFeeBuyerSettlementRatios:  s.ratios("252mango:5nectarine,253mango:6nutmeg"),
				SetFeeCommitmentSettlementBips:  22,
			},
			expFees: marketFees{
				marketID:    2,
				createAsk:   "201acai,2002apple,204avocado",
				createBid:   "214barbadine,212biriba,2102blueberry",
				createCom:   "261raisin,2620raspberry,264rata",
				sellerFlat:  "2201cactus,224cassaba,223cherry",
				buyerFlat:   "2302dewberry,234dragonfruit,233durian",
				sellerRatio: "244tamarillo:4lemon,2402tangerine:20lime,243tayberry:3lime",
				buyerRatio:  "251mandarin:4nectarine,2502mango:50nectarine,254marula:7neem",
				comBips:     "22",
			},
			expNoChange: []uint32{1, 3},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			origMarketFees := make([]marketFees, len(tc.expNoChange))
			for i, marketID := range tc.expNoChange {
				origMarketFees[i] = getMarketFees(marketID)
			}

			var expectedEvents sdk.Events
			if tc.msg != nil {
				event := exchange.NewEventMarketFeesUpdated(tc.msg.MarketId)
				expectedEvents = append(expectedEvents, s.untypeEvent(event))
			}

			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			testFunc := func() {
				s.k.UpdateFees(ctx, tc.msg)
			}
			s.requirePanicEquals(testFunc, tc.expPanic, "UpdateFees")
			if len(tc.expPanic) > 0 || tc.msg == nil {
				return
			}

			updatedMarketFees := getMarketFees(tc.msg.MarketId)
			s.Assert().Equal(tc.expFees, updatedMarketFees, "fees of updated market %d", tc.msg.MarketId)
			for _, expected := range origMarketFees {
				actual := getMarketFees(expected.marketID)
				s.Assert().Equal(expected, actual, "fees of market %d (that should not have changed)", expected.marketID)
			}

			actualEvents := em.Events()
			s.assertEqualEvents(expectedEvents, actualEvents, "events emitted during UpdateFees")
		})
	}
}

func (s *TestSuite) TestKeeper_UpdateIntermediaryDenom() {
	setter := keeper.SetIntermediaryDenom
	tests := []struct {
		name      string
		setup     func()
		marketID  uint32
		denom     string
		updatedBy string
	}{
		{
			name:      "no entries",
			setup:     nil,
			marketID:  1,
			denom:     "banana",
			updatedBy: "alex",
		},
		{
			name: "no entry for market: new value",
			setup: func() {
				store := s.getStore()
				setter(store, 1, "one")
				setter(store, 3, "three")
			},
			marketID:  2,
			denom:     "two",
			updatedBy: "bailey",
		},
		{
			name: "no entry for market: empty",
			setup: func() {
				store := s.getStore()
				setter(store, 1, "one")
				setter(store, 3, "three")
			},
			marketID:  2,
			denom:     "",
			updatedBy: "charlie",
		},
		{
			name: "market has existing: same",
			setup: func() {
				store := s.getStore()
				setter(store, 1, "one")
				setter(store, 3, "two")
				setter(store, 3, "three")
			},
			marketID:  2,
			denom:     "two",
			updatedBy: "devin",
		},
		{
			name: "market has existing: different",
			setup: func() {
				store := s.getStore()
				setter(store, 1, "one")
				setter(store, 3, "two")
				setter(store, 3, "three")
			},
			marketID:  2,
			denom:     "twenty",
			updatedBy: "emerson",
		},
		{
			name: "market has existing: empty",
			setup: func() {
				store := s.getStore()
				setter(store, 1, "one")
				setter(store, 3, "two")
				setter(store, 3, "three")
			},
			marketID:  2,
			denom:     "",
			updatedBy: "forest",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			expEvents := sdk.Events{
				s.untypeEvent(exchange.NewEventMarketIntermediaryDenomUpdated(tc.marketID, tc.updatedBy)),
			}

			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			testFunc := func() {
				s.k.UpdateIntermediaryDenom(ctx, tc.marketID, tc.denom, tc.updatedBy)
			}
			s.Require().NotPanics(testFunc, "UpdateIntermediaryDenom(%d, %q, %q)", tc.marketID, tc.denom, tc.updatedBy)

			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "events emitted during UpdateIntermediaryDenom(%d, %q, %q)",
				tc.marketID, tc.denom, tc.updatedBy)

			actDenom := s.k.GetIntermediaryDenom(s.ctx, tc.marketID)
			s.Assert().Equal(tc.denom, actDenom, "denom after UpdateIntermediaryDenom(%d, %q, %q)", tc.marketID, tc.denom, tc.updatedBy)
		})
	}
}

func (s *TestSuite) TestKeeper_IsMarketKnown() {
	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		expected bool
	}{
		{
			name:     "empty state",
			setup:    nil,
			marketID: 1,
			expected: false,
		},
		{
			name: "unknown market id",
			setup: func() {
				store := s.getStore()
				keeper.SetMarketKnown(store, 1)
				keeper.SetMarketKnown(store, 3)
			},
			marketID: 2,
			expected: false,
		},
		{
			name: "market known",
			setup: func() {
				store := s.getStore()
				keeper.SetMarketKnown(store, 1)
				keeper.SetMarketKnown(store, 2)
				keeper.SetMarketKnown(store, 3)
			},
			marketID: 2,
			expected: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var actual bool
			testFunc := func() {
				actual = s.k.IsMarketKnown(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "IsMarketKnown(%d)", tc.marketID)
			s.Assert().Equal(tc.expected, actual, "IsMarketKnown(%d) result", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_IsMarketAcceptingOrders() {
	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		expected bool
	}{
		{
			name:     "empty state",
			setup:    nil,
			marketID: 1,
			expected: false,
		},
		{
			name: "unknown market id",
			setup: func() {
				store := s.getStore()
				keeper.SetMarketAcceptingOrders(store, 1, true)
				keeper.SetMarketAcceptingOrders(store, 3, true)
				keeper.SetMarketKnown(store, 1)
				keeper.SetMarketKnown(store, 3)
			},
			marketID: 2,
			expected: false,
		},
		{
			name: "market not active",
			setup: func() {
				store := s.getStore()
				keeper.SetMarketAcceptingOrders(store, 1, true)
				keeper.SetMarketAcceptingOrders(store, 2, false)
				keeper.SetMarketAcceptingOrders(store, 3, true)
				keeper.SetMarketKnown(store, 1)
				keeper.SetMarketKnown(store, 2)
				keeper.SetMarketKnown(store, 3)
			},
			marketID: 2,
			expected: false,
		},
		{
			name: "market active and known",
			setup: func() {
				store := s.getStore()
				keeper.SetMarketAcceptingOrders(store, 1, true)
				keeper.SetMarketAcceptingOrders(store, 2, true)
				keeper.SetMarketAcceptingOrders(store, 3, true)
				keeper.SetMarketKnown(store, 1)
				keeper.SetMarketKnown(store, 2)
				keeper.SetMarketKnown(store, 3)
			},
			marketID: 2,
			expected: true,
		},
		{
			name: "market inactive but known",
			setup: func() {
				store := s.getStore()
				keeper.SetMarketAcceptingOrders(store, 1, true)
				keeper.SetMarketAcceptingOrders(store, 2, false)
				keeper.SetMarketAcceptingOrders(store, 3, true)
				keeper.SetMarketKnown(store, 1)
				keeper.SetMarketKnown(store, 2)
				keeper.SetMarketKnown(store, 3)
			},
			marketID: 2,
			expected: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var actual bool
			testFunc := func() {
				actual = s.k.IsMarketAcceptingOrders(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "IsMarketAcceptingOrders(%d)", tc.marketID)
			s.Assert().Equal(tc.expected, actual, "IsMarketAcceptingOrders(%d) result", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_UpdateMarketAcceptingOrders() {
	tests := []struct {
		name      string
		setup     func()
		marketID  uint32
		active    bool
		updatedBy string
		expErr    string
	}{
		{
			name:      "empty state to active",
			marketID:  1,
			active:    true,
			updatedBy: "updatedBy___________",
			expErr:    "market 1 already has accepting-orders true",
		},
		{
			name:      "empty state to inactive",
			marketID:  1,
			active:    false,
			updatedBy: "updatedBy___________",
			expErr:    "",
		},
		{
			name: "active to active",
			setup: func() {
				store := s.getStore()
				keeper.SetMarketAcceptingOrders(store, 1, true)
				keeper.SetMarketAcceptingOrders(store, 2, false)
				keeper.SetMarketAcceptingOrders(store, 3, true)
				keeper.SetMarketAcceptingOrders(store, 4, true)
				keeper.SetMarketAcceptingOrders(store, 5, false)
				keeper.SetMarketKnown(store, 1)
				keeper.SetMarketKnown(store, 2)
				keeper.SetMarketKnown(store, 3)
				keeper.SetMarketKnown(store, 4)
				keeper.SetMarketKnown(store, 5)
			},
			marketID:  3,
			active:    true,
			updatedBy: "updatedBy___________",
			expErr:    "market 3 already has accepting-orders true",
		},
		{
			name: "active to inactive",
			setup: func() {
				store := s.getStore()
				keeper.SetMarketAcceptingOrders(store, 1, true)
				keeper.SetMarketAcceptingOrders(store, 2, false)
				keeper.SetMarketAcceptingOrders(store, 3, true)
				keeper.SetMarketAcceptingOrders(store, 4, true)
				keeper.SetMarketAcceptingOrders(store, 5, false)
				keeper.SetMarketKnown(store, 1)
				keeper.SetMarketKnown(store, 2)
				keeper.SetMarketKnown(store, 3)
				keeper.SetMarketKnown(store, 4)
				keeper.SetMarketKnown(store, 5)
			},
			marketID:  3,
			active:    false,
			updatedBy: "updated_by__________",
			expErr:    "",
		},
		{
			name: "inactive to active",
			setup: func() {
				store := s.getStore()
				keeper.SetMarketAcceptingOrders(store, 11, true)
				keeper.SetMarketAcceptingOrders(store, 12, false)
				keeper.SetMarketAcceptingOrders(store, 13, false)
				keeper.SetMarketAcceptingOrders(store, 14, true)
				keeper.SetMarketAcceptingOrders(store, 15, false)
				keeper.SetMarketKnown(store, 11)
				keeper.SetMarketKnown(store, 12)
				keeper.SetMarketKnown(store, 13)
				keeper.SetMarketKnown(store, 14)
				keeper.SetMarketKnown(store, 15)
			},
			marketID:  13,
			active:    true,
			updatedBy: "updated___by________",
			expErr:    "",
		},
		{
			name: "inactive to inactive",
			setup: func() {
				store := s.getStore()
				keeper.SetMarketAcceptingOrders(store, 11, true)
				keeper.SetMarketAcceptingOrders(store, 12, false)
				keeper.SetMarketAcceptingOrders(store, 13, false)
				keeper.SetMarketAcceptingOrders(store, 14, true)
				keeper.SetMarketAcceptingOrders(store, 15, false)
				keeper.SetMarketKnown(store, 11)
				keeper.SetMarketKnown(store, 12)
				keeper.SetMarketKnown(store, 13)
				keeper.SetMarketKnown(store, 14)
				keeper.SetMarketKnown(store, 15)
			},
			marketID:  13,
			active:    false,
			updatedBy: "__updated_____by____",
			expErr:    "market 13 already has accepting-orders false",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var expEvents sdk.Events
			if len(tc.expErr) == 0 {
				event := exchange.NewEventMarketAcceptingOrdersUpdated(tc.marketID, tc.updatedBy, tc.active)
				expEvents = append(expEvents, s.untypeEvent(event))
			}

			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = s.k.UpdateMarketAcceptingOrders(ctx, tc.marketID, tc.active, tc.updatedBy)
			}
			s.Require().NotPanics(testFunc, "UpdateMarketAcceptingOrders(%d, %t, %s)", tc.marketID, tc.active, string(tc.updatedBy))
			s.assertErrorValue(err, tc.expErr, "UpdateMarketAcceptingOrders(%d, %t, %s)", tc.marketID, tc.active, string(tc.updatedBy))

			events := em.Events()
			s.assertEqualEvents(expEvents, events, "events after UpdateMarketAcceptingOrders")

			if len(tc.expErr) == 0 {
				isActive := s.k.IsMarketAcceptingOrders(s.ctx, tc.marketID)
				s.Assert().Equal(tc.active, isActive, "IsMarketAcceptingOrders(%d) after UpdateMarketAcceptingOrders(%d, %t, ...)",
					tc.marketID, tc.marketID, tc.active)
			}
		})
	}
}

func (s *TestSuite) TestKeeper_IsUserSettlementAllowed() {
	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		expected bool
	}{
		{
			name:     "empty state",
			marketID: 1,
			expected: false,
		},
		{
			name: "unknown market id",
			setup: func() {
				store := s.getStore()
				keeper.SetUserSettlementAllowed(store, 1, true)
				keeper.SetUserSettlementAllowed(store, 3, true)
			},
			marketID: 2,
			expected: false,
		},
		{
			name: "not allowed",
			setup: func() {
				store := s.getStore()
				keeper.SetUserSettlementAllowed(store, 1, true)
				keeper.SetUserSettlementAllowed(store, 2, false)
				keeper.SetUserSettlementAllowed(store, 3, true)
			},
			marketID: 2,
			expected: false,
		},
		{
			name: "allowed",
			setup: func() {
				store := s.getStore()
				keeper.SetUserSettlementAllowed(store, 1, true)
				keeper.SetUserSettlementAllowed(store, 2, true)
				keeper.SetUserSettlementAllowed(store, 3, true)
			},
			marketID: 2,
			expected: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var actual bool
			testFunc := func() {
				actual = s.k.IsUserSettlementAllowed(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "IsUserSettlementAllowed(%d)", tc.marketID)
			s.Assert().Equal(tc.expected, actual, "IsUserSettlementAllowed(%d) result", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_UpdateUserSettlementAllowed() {
	tests := []struct {
		name      string
		setup     func()
		marketID  uint32
		allow     bool
		updatedBy string
		expErr    string
	}{
		{
			name:      "empty state to allowed",
			marketID:  1,
			allow:     true,
			updatedBy: "updatedBy___________",
			expErr:    "",
		},
		{
			name:      "empty state to not allowed",
			marketID:  1,
			allow:     false,
			updatedBy: "updatedBy___________",
			expErr:    "market 1 already has allow-user-settlement false",
		},
		{
			name: "allowed to allowed",
			setup: func() {
				store := s.getStore()
				keeper.SetUserSettlementAllowed(store, 1, true)
				keeper.SetUserSettlementAllowed(store, 2, false)
				keeper.SetUserSettlementAllowed(store, 3, true)
				keeper.SetUserSettlementAllowed(store, 4, true)
				keeper.SetUserSettlementAllowed(store, 5, false)
			},
			marketID:  3,
			allow:     true,
			updatedBy: "updatedBy___________",
			expErr:    "market 3 already has allow-user-settlement true",
		},
		{
			name: "allowed to not allowed",
			setup: func() {
				store := s.getStore()
				keeper.SetUserSettlementAllowed(store, 1, true)
				keeper.SetUserSettlementAllowed(store, 2, false)
				keeper.SetUserSettlementAllowed(store, 3, true)
				keeper.SetUserSettlementAllowed(store, 4, true)
				keeper.SetUserSettlementAllowed(store, 5, false)
			},
			marketID:  3,
			allow:     false,
			updatedBy: "updated_by__________",
			expErr:    "",
		},
		{
			name: "not allowed to allowed",
			setup: func() {
				store := s.getStore()
				keeper.SetUserSettlementAllowed(store, 11, true)
				keeper.SetUserSettlementAllowed(store, 12, false)
				keeper.SetUserSettlementAllowed(store, 13, false)
				keeper.SetUserSettlementAllowed(store, 14, true)
				keeper.SetUserSettlementAllowed(store, 15, false)
			},
			marketID:  13,
			allow:     true,
			updatedBy: "updated___by________",
			expErr:    "",
		},
		{
			name: "not allowed to not allowed",
			setup: func() {
				store := s.getStore()
				keeper.SetUserSettlementAllowed(store, 11, true)
				keeper.SetUserSettlementAllowed(store, 12, false)
				keeper.SetUserSettlementAllowed(store, 13, false)
				keeper.SetUserSettlementAllowed(store, 14, true)
				keeper.SetUserSettlementAllowed(store, 15, false)
			},
			marketID:  13,
			allow:     false,
			updatedBy: "__updated_____by____",
			expErr:    "market 13 already has allow-user-settlement false",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var expEvents sdk.Events
			if len(tc.expErr) == 0 {
				event := exchange.NewEventMarketUserSettleUpdated(tc.marketID, tc.updatedBy, tc.allow)
				expEvents = append(expEvents, s.untypeEvent(event))
			}

			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = s.k.UpdateUserSettlementAllowed(ctx, tc.marketID, tc.allow, tc.updatedBy)
			}
			s.Require().NotPanics(testFunc, "UpdateUserSettlementAllowed(%d, %t, %s)", tc.marketID, tc.allow, tc.updatedBy)
			s.assertErrorValue(err, tc.expErr, "UpdateUserSettlementAllowed(%d, %t, %s)", tc.marketID, tc.allow, tc.updatedBy)

			events := em.Events()
			s.assertEqualEvents(expEvents, events, "events after UpdateUserSettlementAllowed")

			if len(tc.expErr) == 0 {
				isActive := s.k.IsUserSettlementAllowed(s.ctx, tc.marketID)
				s.Assert().Equal(tc.allow, isActive, "IsUserSettlementAllowed(%d) after UpdateUserSettlementAllowed(%d, %t, ...)",
					tc.marketID, tc.marketID, tc.allow)
			}
		})
	}
}

func (s *TestSuite) TestKeeper_IsMarketAcceptingCommitments() {
	setter := keeper.SetMarketAcceptingCommitments
	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		expected bool
	}{
		{
			name:     "empty state",
			marketID: 1,
			expected: false,
		},
		{
			name: "unknown market id",
			setup: func() {
				store := s.getStore()
				setter(store, 1, true)
				setter(store, 3, true)
			},
			marketID: 2,
			expected: false,
		},
		{
			name: "not allowed",
			setup: func() {
				store := s.getStore()
				setter(store, 1, true)
				setter(store, 2, false)
				setter(store, 3, true)
			},
			marketID: 2,
			expected: false,
		},
		{
			name: "allowed",
			setup: func() {
				store := s.getStore()
				setter(store, 1, true)
				setter(store, 2, true)
				setter(store, 3, true)
			},
			marketID: 2,
			expected: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var actual bool
			testFunc := func() {
				actual = s.k.IsMarketAcceptingCommitments(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "IsMarketAcceptingCommitments(%d)", tc.marketID)
			s.Assert().Equal(tc.expected, actual, "IsMarketAcceptingCommitments(%d) result", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_UpdateMarketAcceptingCommitments() {
	setter := keeper.SetMarketAcceptingCommitments
	tests := []struct {
		name      string
		setup     func()
		marketID  uint32
		allow     bool
		updatedBy string
		expErr    string
	}{
		{
			name:      "empty state to allowed",
			marketID:  1,
			allow:     true,
			updatedBy: "updatedBy___________",
			expErr:    "",
		},
		{
			name:      "empty state to not allowed",
			marketID:  1,
			allow:     false,
			updatedBy: "updatedBy___________",
			expErr:    "market 1 already has accepting-commitments false",
		},
		{
			name: "allowed to allowed",
			setup: func() {
				store := s.getStore()
				setter(store, 1, true)
				setter(store, 2, false)
				setter(store, 3, true)
				setter(store, 4, true)
				setter(store, 5, false)
			},
			marketID:  3,
			allow:     true,
			updatedBy: "updatedBy___________",
			expErr:    "market 3 already has accepting-commitments true",
		},
		{
			name: "allowed to not allowed",
			setup: func() {
				store := s.getStore()
				setter(store, 1, true)
				setter(store, 2, false)
				setter(store, 3, true)
				setter(store, 4, true)
				setter(store, 5, false)
			},
			marketID:  3,
			allow:     false,
			updatedBy: "updated_by__________",
			expErr:    "",
		},
		{
			name: "not allowed to allowed",
			setup: func() {
				store := s.getStore()
				setter(store, 11, true)
				setter(store, 12, false)
				setter(store, 13, false)
				setter(store, 14, true)
				setter(store, 15, false)
			},
			marketID:  13,
			allow:     true,
			updatedBy: "updated___by________",
			expErr:    "",
		},
		{
			name: "not allowed to not allowed",
			setup: func() {
				store := s.getStore()
				setter(store, 11, true)
				setter(store, 12, false)
				setter(store, 13, false)
				setter(store, 14, true)
				setter(store, 15, false)
			},
			marketID:  13,
			allow:     false,
			updatedBy: "__updated_____by____",
			expErr:    "market 13 already has accepting-commitments false",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var expEvents sdk.Events
			if len(tc.expErr) == 0 {
				event := exchange.NewEventMarketAcceptingCommitmentsUpdated(tc.marketID, tc.updatedBy, tc.allow)
				expEvents = append(expEvents, s.untypeEvent(event))
			}

			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = s.k.UpdateMarketAcceptingCommitments(ctx, tc.marketID, tc.allow, tc.updatedBy)
			}
			s.Require().NotPanics(testFunc, "UpdateMarketAcceptingCommitments(%d, %t, %s)", tc.marketID, tc.allow, string(tc.updatedBy))
			s.assertErrorValue(err, tc.expErr, "UpdateMarketAcceptingCommitments(%d, %t, %s)", tc.marketID, tc.allow, string(tc.updatedBy))

			events := em.Events()
			s.assertEqualEvents(expEvents, events, "events after UpdateMarketAcceptingCommitments")

			if len(tc.expErr) == 0 {
				isActive := s.k.IsMarketAcceptingCommitments(s.ctx, tc.marketID)
				s.Assert().Equal(tc.allow, isActive, "IsMarketAcceptingCommitments(%d) after UpdateMarketAcceptingCommitments(%d, %t, ...)",
					tc.marketID, tc.marketID, tc.allow)
			}
		})
	}
}

func (s *TestSuite) TestKeeper_HasPermission() {
	goodAcc := sdk.AccAddress("goodAddr____________")
	goodAddr := goodAcc.String()
	authority := s.k.GetAuthority()
	tests := []struct {
		name       string
		setup      func()
		marketID   uint32
		address    string
		permission exchange.Permission
		expected   bool
	}{
		{
			name:       "empty state, empty address",
			marketID:   1,
			address:    "",
			permission: 1,
			expected:   false,
		},
		{
			name:       "empty state, bad address",
			marketID:   1,
			address:    "bad address",
			permission: 1,
			expected:   false,
		},
		{
			name:       "empty state, not authority",
			marketID:   1,
			address:    goodAddr,
			permission: 1,
			expected:   false,
		},
		{
			name:       "empty state, is authority",
			marketID:   1,
			address:    authority,
			permission: 1,
			expected:   true,
		},
		{
			name: "no market perms, empty address",
			setup: func() {
				store := s.getStore()
				keeper.GrantPermissions(store, 1, goodAcc, exchange.AllPermissions())
				keeper.GrantPermissions(store, 3, goodAcc, exchange.AllPermissions())
			},
			marketID:   2,
			address:    "",
			permission: 1,
			expected:   false,
		},
		{
			name: "no market perms, bad address",
			setup: func() {
				store := s.getStore()
				keeper.GrantPermissions(store, 1, goodAcc, exchange.AllPermissions())
				keeper.GrantPermissions(store, 3, goodAcc, exchange.AllPermissions())
			},
			marketID:   2,
			address:    "bad address",
			permission: 1,
			expected:   false,
		},
		{
			name: "no market perms, not authority",
			setup: func() {
				store := s.getStore()
				keeper.GrantPermissions(store, 1, goodAcc, exchange.AllPermissions())
				keeper.GrantPermissions(store, 3, goodAcc, exchange.AllPermissions())
			},
			marketID:   2,
			address:    goodAddr,
			permission: 1,
			expected:   false,
		},
		{
			name: "no market perms, is authority",
			setup: func() {
				store := s.getStore()
				keeper.GrantPermissions(store, 1, goodAcc, exchange.AllPermissions())
				keeper.GrantPermissions(store, 3, goodAcc, exchange.AllPermissions())
			},
			marketID:   2,
			address:    authority,
			permission: 1,
			expected:   true,
		},
		{
			name: "market with perms, empty address",
			setup: func() {
				store := s.getStore()
				keeper.GrantPermissions(store, 1, goodAcc, exchange.AllPermissions())
				keeper.GrantPermissions(store, 2, goodAcc, exchange.AllPermissions())
				keeper.GrantPermissions(store, 3, goodAcc, exchange.AllPermissions())
			},
			marketID:   2,
			address:    "",
			permission: 1,
			expected:   false,
		},
		{
			name: "market with perms, bad address",
			setup: func() {
				store := s.getStore()
				keeper.GrantPermissions(store, 1, goodAcc, exchange.AllPermissions())
				keeper.GrantPermissions(store, 2, goodAcc, exchange.AllPermissions())
				keeper.GrantPermissions(store, 3, goodAcc, exchange.AllPermissions())
			},
			marketID:   2,
			address:    "bad addr",
			permission: 1,
			expected:   false,
		},
		{
			name: "market with perms, unknown address",
			setup: func() {
				store := s.getStore()
				keeper.GrantPermissions(store, 1, goodAcc, exchange.AllPermissions())
				keeper.GrantPermissions(store, 2, goodAcc, exchange.AllPermissions())
				keeper.GrantPermissions(store, 3, goodAcc, exchange.AllPermissions())
			},
			marketID:   2,
			address:    sdk.AccAddress("other_address_______").String(),
			permission: 1,
			expected:   false,
		},
		{
			name: "market with perms, authority",
			setup: func() {
				store := s.getStore()
				keeper.GrantPermissions(store, 1, goodAcc, exchange.AllPermissions())
				keeper.GrantPermissions(store, 2, goodAcc, exchange.AllPermissions())
				keeper.GrantPermissions(store, 3, goodAcc, exchange.AllPermissions())
			},
			marketID:   2,
			address:    authority,
			permission: 1,
			expected:   true,
		},
		{
			name: "address has other perms on this market",
			setup: func() {
				store := s.getStore()
				keeper.GrantPermissions(store, 1, goodAcc, exchange.AllPermissions())
				keeper.GrantPermissions(store, 2, goodAcc, []exchange.Permission{
					exchange.Permission_settle, exchange.Permission_cancel})
				keeper.GrantPermissions(store, 3, goodAcc, exchange.AllPermissions())
			},
			marketID:   2,
			address:    goodAddr,
			permission: exchange.Permission_set_ids,
			expected:   false,
		},
		{
			name: "address only has just perm on this market",
			setup: func() {
				store := s.getStore()
				keeper.GrantPermissions(store, 1, goodAcc, exchange.AllPermissions())
				keeper.GrantPermissions(store, 2, goodAcc, []exchange.Permission{exchange.Permission_withdraw})
				keeper.GrantPermissions(store, 3, goodAcc, exchange.AllPermissions())
			},
			marketID:   2,
			address:    goodAddr,
			permission: exchange.Permission_withdraw,
			expected:   true,
		},
		{
			name: "address has all perms on market",
			setup: func() {
				store := s.getStore()
				keeper.GrantPermissions(store, 1, goodAcc, exchange.AllPermissions())
				keeper.GrantPermissions(store, 2, goodAcc, exchange.AllPermissions())
				keeper.GrantPermissions(store, 3, goodAcc, exchange.AllPermissions())
			},
			marketID:   2,
			address:    goodAddr,
			permission: exchange.Permission_permissions,
			expected:   true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var actual bool
			testFunc := func() {
				actual = s.k.HasPermission(s.ctx, tc.marketID, tc.address, tc.permission)
			}
			s.Require().NotPanics(testFunc, "HasPermission(%d, %q, %s)", tc.marketID, tc.address, tc.permission)
			s.Assert().Equal(tc.expected, actual, "HasPermission(%d, %q, %s) result", tc.marketID, tc.address, tc.permission)
		})
	}
}

// permChecker is the function signature of a permission checking function, e.g. CanSettleOrders.
type permChecker func(ctx sdk.Context, marketID uint32, address string) bool

// runPermTest runs a set of tests on a permission checking function, e.g. CanSettleOrders.
func (s *TestSuite) runPermTest(perm exchange.Permission, checker permChecker, name string) {
	allPermsAcc := sdk.AccAddress("allPerms____________")
	justPermAcc := sdk.AccAddress("justPerm____________")
	otherPermsAcc := sdk.AccAddress("otherPerms__________")
	noPermsAcc := sdk.AccAddress("noPerms_____________")
	authority := s.k.GetAuthority()

	allPerms := exchange.AllPermissions()
	otherPerms := make([]exchange.Permission, 0, len(allPermsAcc)-1)
	for _, p := range exchange.AllPermissions() {
		if p != perm {
			otherPerms = append(otherPerms, p)
		}
	}

	defaultSetup := func() {
		store := s.getStore()
		keeper.GrantPermissions(store, 10, allPermsAcc, allPerms)
		keeper.GrantPermissions(store, 10, justPermAcc, allPerms)
		keeper.GrantPermissions(store, 10, otherPermsAcc, allPerms)
		keeper.GrantPermissions(store, 10, noPermsAcc, allPerms)

		keeper.GrantPermissions(store, 11, allPermsAcc, allPerms)
		keeper.GrantPermissions(store, 11, justPermAcc, []exchange.Permission{perm})
		keeper.GrantPermissions(store, 11, otherPermsAcc, otherPerms)

		keeper.GrantPermissions(store, 12, allPermsAcc, allPerms)
		keeper.GrantPermissions(store, 12, justPermAcc, allPerms)
		keeper.GrantPermissions(store, 12, otherPermsAcc, allPerms)
		keeper.GrantPermissions(store, 12, noPermsAcc, allPerms)
	}

	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		admin    string
		expected bool
	}{
		{
			name:     "empty state: empty addr",
			marketID: 1,
			admin:    "",
			expected: false,
		},
		{
			name:     "empty state: authority",
			marketID: 1,
			admin:    authority,
			expected: true,
		},
		{
			name:     "empty state: addr with all perms",
			marketID: 1,
			admin:    allPermsAcc.String(),
			expected: false,
		},
		{
			name:     "empty state: addr with just perm",
			marketID: 1,
			admin:    justPermAcc.String(),
			expected: false,
		},
		{
			name:     "empty state: addr with all other perms",
			marketID: 1,
			admin:    otherPermsAcc.String(),
			expected: false,
		},
		{
			name:     "empty state: addr without any perms",
			marketID: 1,
			admin:    noPermsAcc.String(),
			expected: false,
		},

		{
			name:     "existing market: empty addr",
			setup:    defaultSetup,
			marketID: 11,
			admin:    "",
			expected: false,
		},
		{
			name:     "existing market: authority",
			setup:    defaultSetup,
			marketID: 11,
			admin:    authority,
			expected: true,
		},
		{
			name:     "existing market: addr with all perms",
			setup:    defaultSetup,
			marketID: 11,
			admin:    allPermsAcc.String(),
			expected: true,
		},
		{
			name:     "existing market: addr with just perm",
			setup:    defaultSetup,
			marketID: 11,
			admin:    justPermAcc.String(),
			expected: true,
		},
		{
			name:     "existing market: addr with all other perms",
			setup:    defaultSetup,
			marketID: 11,
			admin:    otherPermsAcc.String(),
			expected: false,
		},
		{
			name:     "existing market: addr without any perms",
			setup:    defaultSetup,
			marketID: 11,
			admin:    noPermsAcc.String(),
			expected: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var actual bool
			testFunc := func() {
				actual = checker(s.ctx, tc.marketID, tc.admin)
			}
			s.Require().NotPanics(testFunc, "%s(%d, %q)", name, tc.marketID, tc.admin)
			s.Assert().Equal(tc.expected, actual, "%s(%d, %q) result", name, tc.marketID, tc.admin)
		})
	}
}

func (s *TestSuite) TestKeeper_CanSettleOrders() {
	s.runPermTest(exchange.Permission_settle, s.k.CanSettleOrders, "CanSettleOrders")
}

func (s *TestSuite) TestKeeper_CanSettleCommitments() {
	s.runPermTest(exchange.Permission_settle, s.k.CanSettleCommitments, "CanSettleCommitments")
}

func (s *TestSuite) TestKeeper_CanSetIDs() {
	s.runPermTest(exchange.Permission_set_ids, s.k.CanSetIDs, "CanSetIDs")
}

func (s *TestSuite) TestKeeper_CanCancelOrdersForMarket() {
	s.runPermTest(exchange.Permission_cancel, s.k.CanCancelOrdersForMarket, "CanCancelOrdersForMarket")
}

func (s *TestSuite) TestKeeper_CanReleaseCommitmentsForMarket() {
	s.runPermTest(exchange.Permission_cancel, s.k.CanReleaseCommitmentsForMarket, "CanReleaseCommitmentsForMarket")
}

func (s *TestSuite) TestKeeper_CanTransferCommitmentsForMarket() {
	s.runPermTest(exchange.Permission_cancel, s.k.CanTransferCommitmentsForMarket, "CanTransferCommitmentsForMarket")
}

func (s *TestSuite) TestKeeper_CanWithdrawMarketFunds() {
	s.runPermTest(exchange.Permission_withdraw, s.k.CanWithdrawMarketFunds, "CanWithdrawMarketFunds")
}

func (s *TestSuite) TestKeeper_CanUpdateMarket() {
	s.runPermTest(exchange.Permission_update, s.k.CanUpdateMarket, "CanUpdateMarket")
}

func (s *TestSuite) TestKeeper_CanManagePermissions() {
	s.runPermTest(exchange.Permission_permissions, s.k.CanManagePermissions, "CanManagePermissions")
}

func (s *TestSuite) TestKeeper_CanManageReqAttrs() {
	s.runPermTest(exchange.Permission_attributes, s.k.CanManageReqAttrs, "CanManageReqAttrs")
}

func (s *TestSuite) TestKeeper_GetUserPermissions() {
	addrNone := sdk.AccAddress("address_none________")
	addrOne := sdk.AccAddress("address_one_________")
	addrTwo := sdk.AccAddress("address_two_________")
	addrAll := sdk.AccAddress("address_all_________")
	addrEven := sdk.AccAddress("address_even________")
	addrOdd := sdk.AccAddress("address_odd_________")

	onePerm := []exchange.Permission{exchange.Permission_settle}
	twoPerms := []exchange.Permission{exchange.Permission_cancel, exchange.Permission_attributes}
	allPerms := exchange.AllPermissions()
	evenPerms := make([]exchange.Permission, 0, 1+len(allPerms)/2)
	oddPerms := make([]exchange.Permission, 0, 1+len(allPerms)/2)
	for _, p := range allPerms {
		if p%2 == 0 {
			evenPerms = append(evenPerms, p)
		} else {
			oddPerms = append(oddPerms, p)
		}
	}

	defaultSetup := func() {
		store := s.getStore()
		keeper.GrantPermissions(store, 1, addrNone, allPerms)
		keeper.GrantPermissions(store, 1, addrOne, allPerms)
		keeper.GrantPermissions(store, 1, addrTwo, allPerms)
		keeper.GrantPermissions(store, 1, addrAll, allPerms)
		keeper.GrantPermissions(store, 1, addrEven, allPerms)
		keeper.GrantPermissions(store, 1, addrOdd, allPerms)

		keeper.GrantPermissions(store, 2, addrNone, nil)
		keeper.GrantPermissions(store, 2, addrOne, onePerm)
		keeper.GrantPermissions(store, 2, addrTwo, twoPerms)
		keeper.GrantPermissions(store, 2, addrAll, allPerms)
		keeper.GrantPermissions(store, 2, addrEven, evenPerms)
		keeper.GrantPermissions(store, 2, addrOdd, oddPerms)

		keeper.GrantPermissions(store, 3, addrNone, allPerms)
		keeper.GrantPermissions(store, 3, addrOne, allPerms)
		keeper.GrantPermissions(store, 3, addrTwo, allPerms)
		keeper.GrantPermissions(store, 3, addrAll, allPerms)
		keeper.GrantPermissions(store, 3, addrEven, allPerms)
		keeper.GrantPermissions(store, 3, addrOdd, allPerms)
	}

	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		addr     sdk.AccAddress
		expected []exchange.Permission
		expPanic string
	}{
		{
			name:     "nil addr",
			marketID: 1,
			addr:     nil,
			expPanic: "empty address not allowed",
		},
		{
			name:     "empty addr",
			marketID: 1,
			addr:     sdk.AccAddress{},
			expPanic: "empty address not allowed",
		},
		{
			name:     "empty state",
			marketID: 1,
			addr:     sdk.AccAddress("some_address________"),
			expected: nil,
		},
		{
			name:     "no perms in market",
			setup:    defaultSetup,
			marketID: 2,
			addr:     addrNone,
			expected: nil,
		},
		{
			name:     "one perm in market",
			setup:    defaultSetup,
			marketID: 2,
			addr:     addrOne,
			expected: onePerm,
		},
		{
			name:     "two perms in market",
			setup:    defaultSetup,
			marketID: 2,
			addr:     addrTwo,
			expected: twoPerms,
		},
		{
			name:     "odd perms",
			setup:    defaultSetup,
			marketID: 2,
			addr:     addrOdd,
			expected: oddPerms,
		},
		{
			name:     "even perms",
			setup:    defaultSetup,
			marketID: 2,
			addr:     addrEven,
			expected: evenPerms,
		},
		{
			name:     "all perms",
			setup:    defaultSetup,
			marketID: 2,
			addr:     addrAll,
			expected: allPerms,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var actual []exchange.Permission
			testFunc := func() {
				actual = s.k.GetUserPermissions(s.ctx, tc.marketID, tc.addr)
			}
			s.requirePanicEquals(testFunc, tc.expPanic, "GetUserPermissions(%d, %q)", tc.marketID, string(tc.addr))
			s.Assert().Equal(tc.expected, actual, "GetUserPermissions(%d, %q) result", tc.marketID, string(tc.addr))
		})
	}
}

func (s *TestSuite) TestKeeper_GetAccessGrants() {
	addrNone := sdk.AccAddress("address_none________")
	addrOne := sdk.AccAddress("address_one_________")
	addrTwo := sdk.AccAddress("address_two_________")
	addrAll := sdk.AccAddress("address_all_________")
	addrEven := sdk.AccAddress("address_even________")
	addrOdd := sdk.AccAddress("address_odd_________")

	onePerm := []exchange.Permission{exchange.Permission_settle}
	oneOtherPerm := []exchange.Permission{exchange.Permission_set_ids}
	twoPerms := []exchange.Permission{exchange.Permission_cancel, exchange.Permission_attributes}
	allPerms := exchange.AllPermissions()
	evenPerms := make([]exchange.Permission, 0, 1+len(allPerms)/2)
	oddPerms := make([]exchange.Permission, 0, 1+len(allPerms)/2)
	for _, p := range allPerms {
		if p%2 == 0 {
			evenPerms = append(evenPerms, p)
		} else {
			oddPerms = append(oddPerms, p)
		}
	}

	defaultSetup := func() {
		store := s.getStore()
		keeper.GrantPermissions(store, 1, addrNone, allPerms)
		keeper.GrantPermissions(store, 1, addrOne, allPerms)
		keeper.GrantPermissions(store, 1, addrTwo, allPerms)
		keeper.GrantPermissions(store, 1, addrAll, allPerms)
		keeper.GrantPermissions(store, 1, addrEven, allPerms)
		keeper.GrantPermissions(store, 1, addrOdd, allPerms)

		keeper.GrantPermissions(store, 2, addrOne, oneOtherPerm)

		keeper.GrantPermissions(store, 3, addrNone, nil)
		keeper.GrantPermissions(store, 3, addrOne, onePerm)
		keeper.GrantPermissions(store, 3, addrTwo, twoPerms)
		keeper.GrantPermissions(store, 3, addrAll, allPerms)
		keeper.GrantPermissions(store, 3, addrEven, evenPerms)
		keeper.GrantPermissions(store, 3, addrOdd, oddPerms)

		keeper.GrantPermissions(store, 4, addrNone, allPerms)
		keeper.GrantPermissions(store, 4, addrOne, allPerms)
		keeper.GrantPermissions(store, 4, addrTwo, allPerms)
		keeper.GrantPermissions(store, 4, addrAll, allPerms)
		keeper.GrantPermissions(store, 4, addrEven, allPerms)
		keeper.GrantPermissions(store, 4, addrOdd, allPerms)
	}

	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		expected []exchange.AccessGrant
	}{
		{
			name:     "empty state",
			marketID: 1,
			expected: nil,
		},
		{
			name:     "market without any permissions",
			setup:    defaultSetup,
			marketID: 5,
			expected: nil,
		},
		{
			name:     "market with just one permission",
			setup:    defaultSetup,
			marketID: 2,
			expected: []exchange.AccessGrant{
				{Address: addrOne.String(), Permissions: oneOtherPerm},
			},
		},
		{
			name:     "market with several permissions",
			setup:    defaultSetup,
			marketID: 3,
			expected: []exchange.AccessGrant{
				{Address: addrAll.String(), Permissions: allPerms},
				{Address: addrEven.String(), Permissions: evenPerms},
				{Address: addrOdd.String(), Permissions: oddPerms},
				{Address: addrOne.String(), Permissions: onePerm},
				{Address: addrTwo.String(), Permissions: twoPerms},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var actual []exchange.AccessGrant
			testFunc := func() {
				actual = s.k.GetAccessGrants(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "GetAccessGrants(%d)", tc.marketID)
			s.Assert().Equal(tc.expected, actual, "GetAccessGrants(%d) result", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_UpdatePermissions() {
	adminAddr := sdk.AccAddress("admin_address_woooo_").String()
	oneAcc := sdk.AccAddress("addr_one____________")
	oneAddr := oneAcc.String()
	twoAcc := sdk.AccAddress("addr_two____________")
	twoAddr := twoAcc.String()

	tests := []struct {
		name      string
		setup     func()
		msg       *exchange.MsgMarketManagePermissionsRequest
		expErr    string
		expPanic  string
		expGrants []exchange.AccessGrant
	}{
		{
			name:     "nil msg",
			msg:      nil,
			expPanic: "runtime error: invalid memory address or nil pointer dereference",
		},
		{
			name: "invalid revoke-all addr",
			msg: &exchange.MsgMarketManagePermissionsRequest{
				Admin:     adminAddr,
				RevokeAll: []string{"invalid"},
			},
			expPanic: "decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name: "invalid to-revoke addr",
			msg: &exchange.MsgMarketManagePermissionsRequest{
				Admin:    adminAddr,
				ToRevoke: []exchange.AccessGrant{{Address: "invalid"}},
			},
			expPanic: "decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name: "invalid to-grant addr",
			msg: &exchange.MsgMarketManagePermissionsRequest{
				Admin:   adminAddr,
				ToGrant: []exchange.AccessGrant{{Address: "invalid"}},
			},
			expPanic: "decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name: "revoke-all addr without any perms",
			setup: func() {
				store := s.getStore()
				keeper.GrantPermissions(store, 1, twoAcc, exchange.AllPermissions())
			},
			msg: &exchange.MsgMarketManagePermissionsRequest{
				Admin:     adminAddr,
				MarketId:  1,
				RevokeAll: []string{oneAddr},
			},
			expErr: "account " + oneAddr + " does not have any permissions for market 1",
		},
		{
			name: "to-revoke perm not granted",
			setup: func() {
				store := s.getStore()
				keeper.GrantPermissions(store, 1, oneAcc, []exchange.Permission{exchange.Permission_update})
				keeper.GrantPermissions(store, 1, twoAcc, exchange.AllPermissions())
			},
			msg: &exchange.MsgMarketManagePermissionsRequest{
				Admin:    adminAddr,
				MarketId: 1,
				ToRevoke: []exchange.AccessGrant{
					{Address: oneAddr, Permissions: []exchange.Permission{exchange.Permission_settle}},
				},
			},
			expErr: "account " + oneAddr + " does not have PERMISSION_SETTLE for market 1",
		},
		{
			name: "to-add perm already granted",
			setup: func() {
				store := s.getStore()
				keeper.GrantPermissions(store, 2, oneAcc, exchange.AllPermissions())
				keeper.GrantPermissions(store, 2, twoAcc, []exchange.Permission{exchange.Permission_update})
			},
			msg: &exchange.MsgMarketManagePermissionsRequest{
				Admin:    adminAddr,
				MarketId: 2,
				ToGrant: []exchange.AccessGrant{
					{Address: twoAddr, Permissions: []exchange.Permission{exchange.Permission_update}},
				},
			},
			expErr: "account " + twoAddr + " already has PERMISSION_UPDATE for market 2",
		},
		{
			name: "multiple errors",
			setup: func() {
				store := s.getStore()
				keeper.GrantPermissions(store, 3, sdk.AccAddress("bbbbbbbbbbbbbbbbbbbbb"), []exchange.Permission{
					exchange.Permission_attributes})
				keeper.GrantPermissions(store, 3, sdk.AccAddress("dddddddddddddddddddd"), []exchange.Permission{
					exchange.Permission_cancel, exchange.Permission_attributes})
				keeper.GrantPermissions(store, 3, sdk.AccAddress("ffffffffffffffffffff"), []exchange.Permission{
					exchange.Permission_permissions, exchange.Permission_withdraw})
				keeper.GrantPermissions(store, 3, sdk.AccAddress("gggggggggggggggggggg"), []exchange.Permission{
					exchange.Permission_withdraw, exchange.Permission_attributes})
				keeper.GrantPermissions(store, 3, sdk.AccAddress("hhhhhhhhhhhhhhhhhhhh"), []exchange.Permission{
					exchange.Permission_withdraw, exchange.Permission_set_ids})
			},
			msg: &exchange.MsgMarketManagePermissionsRequest{
				Admin:    adminAddr,
				MarketId: 3,
				RevokeAll: []string{
					sdk.AccAddress("aaaaaaaaaaaaaaaaaaaaa").String(),
					sdk.AccAddress("bbbbbbbbbbbbbbbbbbbbb").String(),
					sdk.AccAddress("ccccccccccccccccccccc").String(),
				},
				ToRevoke: []exchange.AccessGrant{
					{
						Address:     sdk.AccAddress("dddddddddddddddddddd").String(),
						Permissions: []exchange.Permission{exchange.Permission_update, exchange.Permission_cancel},
					},
					{
						Address:     sdk.AccAddress("eeeeeeeeeeeeeeeeeeee").String(),
						Permissions: []exchange.Permission{exchange.Permission_set_ids, exchange.Permission_withdraw},
					},
					{
						Address:     sdk.AccAddress("ffffffffffffffffffff").String(),
						Permissions: []exchange.Permission{exchange.Permission_permissions, exchange.Permission_settle},
					},
				},
				ToGrant: []exchange.AccessGrant{
					{
						Address:     sdk.AccAddress("gggggggggggggggggggg").String(),
						Permissions: []exchange.Permission{exchange.Permission_withdraw, exchange.Permission_attributes},
					},
					{
						Address:     sdk.AccAddress("hhhhhhhhhhhhhhhhhhhh").String(),
						Permissions: []exchange.Permission{exchange.Permission_cancel, exchange.Permission_set_ids},
					},
					{
						Address:     sdk.AccAddress("iiiiiiiiiiiiiiiiiiii").String(),
						Permissions: []exchange.Permission{exchange.Permission_update, exchange.Permission_settle},
					},
				},
			},
			expErr: s.joinErrs(
				"account "+sdk.AccAddress("aaaaaaaaaaaaaaaaaaaaa").String()+" does not have any permissions for market 3",
				"account "+sdk.AccAddress("ccccccccccccccccccccc").String()+" does not have any permissions for market 3",
				"account "+sdk.AccAddress("dddddddddddddddddddd").String()+" does not have PERMISSION_UPDATE for market 3",
				"account "+sdk.AccAddress("eeeeeeeeeeeeeeeeeeee").String()+" does not have PERMISSION_SET_IDS for market 3",
				"account "+sdk.AccAddress("eeeeeeeeeeeeeeeeeeee").String()+" does not have PERMISSION_WITHDRAW for market 3",
				"account "+sdk.AccAddress("ffffffffffffffffffff").String()+" does not have PERMISSION_SETTLE for market 3",
				"account "+sdk.AccAddress("gggggggggggggggggggg").String()+" already has PERMISSION_WITHDRAW for market 3",
				"account "+sdk.AccAddress("gggggggggggggggggggg").String()+" already has PERMISSION_ATTRIBUTES for market 3",
				"account "+sdk.AccAddress("hhhhhhhhhhhhhhhhhhhh").String()+" already has PERMISSION_SET_IDS for market 3",
			),
		},
		{
			name: "just a revoke all",
			setup: func() {
				keeper.GrantPermissions(s.getStore(), 5, oneAcc, []exchange.Permission{3})
				keeper.GrantPermissions(s.getStore(), 5, twoAcc, []exchange.Permission{4, 2})
				keeper.GrantPermissions(s.getStore(), 6, oneAcc, []exchange.Permission{3})
				keeper.GrantPermissions(s.getStore(), 6, twoAcc, []exchange.Permission{4, 2})
				keeper.GrantPermissions(s.getStore(), 7, oneAcc, []exchange.Permission{3})
				keeper.GrantPermissions(s.getStore(), 7, twoAcc, []exchange.Permission{4, 2})
			},
			msg: &exchange.MsgMarketManagePermissionsRequest{
				Admin:     adminAddr,
				MarketId:  6,
				RevokeAll: []string{twoAddr},
			},
			expGrants: []exchange.AccessGrant{
				{Address: oneAddr, Permissions: []exchange.Permission{3}},
			},
		},
		{
			name: "just a to-revoke",
			setup: func() {
				keeper.GrantPermissions(s.getStore(), 5, oneAcc, []exchange.Permission{3})
				keeper.GrantPermissions(s.getStore(), 5, twoAcc, []exchange.Permission{4, 2})
				keeper.GrantPermissions(s.getStore(), 6, oneAcc, []exchange.Permission{3})
				keeper.GrantPermissions(s.getStore(), 6, twoAcc, []exchange.Permission{4, 2})
				keeper.GrantPermissions(s.getStore(), 7, oneAcc, []exchange.Permission{3})
				keeper.GrantPermissions(s.getStore(), 7, twoAcc, []exchange.Permission{4, 2})
			},
			msg: &exchange.MsgMarketManagePermissionsRequest{
				Admin:    adminAddr,
				MarketId: 6,
				ToRevoke: []exchange.AccessGrant{
					{Address: twoAddr, Permissions: []exchange.Permission{2}},
				},
			},
			expGrants: []exchange.AccessGrant{
				{Address: oneAddr, Permissions: []exchange.Permission{3}},
				{Address: twoAddr, Permissions: []exchange.Permission{4}},
			},
		},
		{
			name: "just a to-grant",
			setup: func() {
				keeper.GrantPermissions(s.getStore(), 5, oneAcc, []exchange.Permission{3})
				keeper.GrantPermissions(s.getStore(), 5, twoAcc, []exchange.Permission{4, 2})
				keeper.GrantPermissions(s.getStore(), 6, oneAcc, []exchange.Permission{3})
				keeper.GrantPermissions(s.getStore(), 6, twoAcc, []exchange.Permission{4, 2})
				keeper.GrantPermissions(s.getStore(), 7, oneAcc, []exchange.Permission{3})
				keeper.GrantPermissions(s.getStore(), 7, twoAcc, []exchange.Permission{4, 2})
			},
			msg: &exchange.MsgMarketManagePermissionsRequest{
				Admin:    adminAddr,
				MarketId: 6,
				ToGrant:  []exchange.AccessGrant{{Address: twoAddr, Permissions: []exchange.Permission{1}}},
			},
			expGrants: []exchange.AccessGrant{
				{Address: oneAddr, Permissions: []exchange.Permission{3}},
				{Address: twoAddr, Permissions: []exchange.Permission{1, 2, 4}},
			},
		},
		{
			name: "revoke all grant one",
			setup: func() {
				store := s.getStore()
				keeper.GrantPermissions(store, 1, oneAcc, exchange.AllPermissions())
				keeper.GrantPermissions(store, 2, oneAcc, exchange.AllPermissions())
				keeper.GrantPermissions(store, 3, oneAcc, exchange.AllPermissions())
			},
			msg: &exchange.MsgMarketManagePermissionsRequest{
				Admin:     adminAddr,
				MarketId:  2,
				RevokeAll: []string{oneAddr},
				ToGrant:   []exchange.AccessGrant{{Address: oneAddr, Permissions: []exchange.Permission{5}}},
			},
			expGrants: []exchange.AccessGrant{{Address: oneAddr, Permissions: []exchange.Permission{5}}},
		},
		{
			name: "revoke one grant different",
			setup: func() {
				store := s.getStore()
				perms := []exchange.Permission{1, 4, 6}
				keeper.GrantPermissions(store, 1, oneAcc, perms)
				keeper.GrantPermissions(store, 2, oneAcc, perms)
				keeper.GrantPermissions(store, 3, oneAcc, perms)
			},
			msg: &exchange.MsgMarketManagePermissionsRequest{
				Admin:    adminAddr,
				MarketId: 2,
				ToRevoke: []exchange.AccessGrant{{Address: oneAddr, Permissions: []exchange.Permission{4}}},
				ToGrant:  []exchange.AccessGrant{{Address: oneAddr, Permissions: []exchange.Permission{5}}},
			},
			expGrants: []exchange.AccessGrant{{Address: oneAddr, Permissions: []exchange.Permission{1, 5, 6}}},
		},
		{
			name: "complex",
			// revoke two from addr with two
			// revoke all from addr with one, regrant all
			// revoke one from addr with all
			// grant two to new addr
			// revoke one from addr with two, replace with another
			setup: func() {
				store := s.getStore()
				keeper.GrantPermissions(store, 33, sdk.AccAddress("aaaaaaaaaaaaaaaaaaaa"), []exchange.Permission{2, 6})
				keeper.GrantPermissions(store, 33, sdk.AccAddress("bbbbbbbbbbbbbbbbbbbb"), []exchange.Permission{1})
				keeper.GrantPermissions(store, 33, sdk.AccAddress("cccccccccccccccccccc"), exchange.AllPermissions())
				keeper.GrantPermissions(store, 33, sdk.AccAddress("eeeeeeeeeeeeeeeeeeee"), []exchange.Permission{7, 3})
			},
			msg: &exchange.MsgMarketManagePermissionsRequest{
				Admin:     adminAddr,
				MarketId:  33,
				RevokeAll: []string{sdk.AccAddress("bbbbbbbbbbbbbbbbbbbb").String()},
				ToRevoke: []exchange.AccessGrant{
					{Address: sdk.AccAddress("aaaaaaaaaaaaaaaaaaaa").String(), Permissions: []exchange.Permission{2, 6}},
					{Address: sdk.AccAddress("cccccccccccccccccccc").String(), Permissions: []exchange.Permission{3}},
					{Address: sdk.AccAddress("eeeeeeeeeeeeeeeeeeee").String(), Permissions: []exchange.Permission{3}},
				},
				ToGrant: []exchange.AccessGrant{
					{Address: sdk.AccAddress("bbbbbbbbbbbbbbbbbbbb").String(), Permissions: exchange.AllPermissions()},
					{Address: sdk.AccAddress("dddddddddddddddddddd").String(), Permissions: []exchange.Permission{5, 4}},
					{Address: sdk.AccAddress("eeeeeeeeeeeeeeeeeeee").String(), Permissions: []exchange.Permission{6}},
				},
			},
			expGrants: []exchange.AccessGrant{
				{Address: sdk.AccAddress("bbbbbbbbbbbbbbbbbbbb").String(), Permissions: exchange.AllPermissions()},
				{Address: sdk.AccAddress("cccccccccccccccccccc").String(), Permissions: []exchange.Permission{1, 2, 4, 5, 6, 7}},
				{Address: sdk.AccAddress("dddddddddddddddddddd").String(), Permissions: []exchange.Permission{4, 5}},
				{Address: sdk.AccAddress("eeeeeeeeeeeeeeeeeeee").String(), Permissions: []exchange.Permission{6, 7}},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var expEvents sdk.Events
			if len(tc.expPanic) == 0 && len(tc.expErr) == 0 {
				event := exchange.NewEventMarketPermissionsUpdated(tc.msg.MarketId, tc.msg.Admin)
				expEvents = append(expEvents, s.untypeEvent(event))
			}

			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = s.k.UpdatePermissions(ctx, tc.msg)
			}
			s.requirePanicEquals(testFunc, tc.expPanic, "UpdatePermissions")
			if len(tc.expPanic) > 0 {
				return
			}

			s.assertErrorValue(err, tc.expErr, "UpdatePermissions error")

			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "events emitted during UpdatePermissions")

			if len(tc.expErr) > 0 {
				return
			}

			actGrants := s.k.GetAccessGrants(ctx, tc.msg.MarketId)
			s.Assert().Equal(tc.expGrants, actGrants, "access grants for market %d after UpdatePermissions", tc.msg.MarketId)
		})
	}
}

func (s *TestSuite) TestKeeper_GetReqAttrsAsk() {
	setter := keeper.SetReqAttrsAsk
	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		expected []string
	}{
		{
			name:     "empty state",
			setup:    nil,
			marketID: 1,
			expected: nil,
		},
		{
			name: "market without any",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []string{"bb.aa", "*.cc.bb.aa", "banana"})
				setter(store, 3, []string{"yy.zz", "*.xx.yy.zz", "banana"})
			},
			marketID: 2,
			expected: nil,
		},
		{
			name: "market with one",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []string{"bb.aa", "*.cc.bb.aa", "banana"})
				setter(store, 2, []string{"raspberry"})
				setter(store, 3, []string{"yy.zz", "*.xx.yy.zz", "banana"})
			},
			marketID: 2,
			expected: []string{"raspberry"},
		},
		{
			name: "market with two",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []string{"bb.aa", "*.cc.bb.aa", "banana"})
				setter(store, 2, []string{"raspberry"})
				setter(store, 3, []string{"knee", "elbow"})
				setter(store, 4, []string{"yy.zz", "*.xx.yy.zz", "banana"})
			},
			marketID: 3,
			expected: []string{"knee", "elbow"},
		},
		{
			name: "market with three",
			setup: func() {
				store := s.getStore()
				setter(store, 2, []string{"raspberry"})
				setter(store, 33, []string{"knee", "elbow"})
				setter(store, 444, []string{"head", "shoulders", "toes"})
			},
			marketID: 444,
			expected: []string{"head", "shoulders", "toes"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var actual []string
			testFunc := func() {
				actual = s.k.GetReqAttrsAsk(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "GetReqAttrsAsk(%d)", tc.marketID)
			s.Assert().Equal(tc.expected, actual, "GetReqAttrsAsk(%d)", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_GetReqAttrsBid() {
	setter := keeper.SetReqAttrsBid
	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		expected []string
	}{
		{
			name:     "empty state",
			setup:    nil,
			marketID: 1,
			expected: nil,
		},
		{
			name: "market without any",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []string{"bb.aa", "*.cc.bb.aa", "banana"})
				setter(store, 3, []string{"yy.zz", "*.xx.yy.zz", "banana"})
			},
			marketID: 2,
			expected: nil,
		},
		{
			name: "market with one",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []string{"bb.aa", "*.cc.bb.aa", "banana"})
				setter(store, 2, []string{"raspberry"})
				setter(store, 3, []string{"yy.zz", "*.xx.yy.zz", "banana"})
			},
			marketID: 2,
			expected: []string{"raspberry"},
		},
		{
			name: "market with two",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []string{"bb.aa", "*.cc.bb.aa", "banana"})
				setter(store, 2, []string{"raspberry"})
				setter(store, 3, []string{"knee", "elbow"})
				setter(store, 4, []string{"yy.zz", "*.xx.yy.zz", "banana"})
			},
			marketID: 3,
			expected: []string{"knee", "elbow"},
		},
		{
			name: "market with three",
			setup: func() {
				store := s.getStore()
				setter(store, 2, []string{"raspberry"})
				setter(store, 33, []string{"knee", "elbow"})
				setter(store, 444, []string{"head", "shoulders", "toes"})
			},
			marketID: 444,
			expected: []string{"head", "shoulders", "toes"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var actual []string
			testFunc := func() {
				actual = s.k.GetReqAttrsBid(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "GetReqAttrsBid(%d)", tc.marketID)
			s.Assert().Equal(tc.expected, actual, "GetReqAttrsBid(%d)", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_GetReqAttrsCommitment() {
	setter := keeper.SetReqAttrsCommitment
	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		expected []string
	}{
		{
			name:     "empty state",
			setup:    nil,
			marketID: 1,
			expected: nil,
		},
		{
			name: "market without any",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []string{"bb.aa", "*.cc.bb.aa", "banana"})
				setter(store, 3, []string{"yy.zz", "*.xx.yy.zz", "banana"})
			},
			marketID: 2,
			expected: nil,
		},
		{
			name: "market with one",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []string{"bb.aa", "*.cc.bb.aa", "banana"})
				setter(store, 2, []string{"raspberry"})
				setter(store, 3, []string{"yy.zz", "*.xx.yy.zz", "banana"})
			},
			marketID: 2,
			expected: []string{"raspberry"},
		},
		{
			name: "market with two",
			setup: func() {
				store := s.getStore()
				setter(store, 1, []string{"bb.aa", "*.cc.bb.aa", "banana"})
				setter(store, 2, []string{"raspberry"})
				setter(store, 3, []string{"knee", "elbow"})
				setter(store, 4, []string{"yy.zz", "*.xx.yy.zz", "banana"})
			},
			marketID: 3,
			expected: []string{"knee", "elbow"},
		},
		{
			name: "market with three",
			setup: func() {
				store := s.getStore()
				setter(store, 2, []string{"raspberry"})
				setter(store, 33, []string{"knee", "elbow"})
				setter(store, 444, []string{"head", "shoulders", "toes"})
			},
			marketID: 444,
			expected: []string{"head", "shoulders", "toes"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var actual []string
			testFunc := func() {
				actual = s.k.GetReqAttrsCommitment(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "GetReqAttrsCommitment(%d)", tc.marketID)
			s.Assert().Equal(tc.expected, actual, "GetReqAttrsCommitment(%d)", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_CanCreateAsk() {
	setter := keeper.SetReqAttrsAsk
	addr1 := sdk.AccAddress("addr_one____________")
	addr2 := sdk.AccAddress("addr_two____________")
	addr3 := sdk.AccAddress("addr_three__________")

	tests := []struct {
		name           string
		setup          func()
		attrKeeper     *MockAttributeKeeper
		marketID       uint32
		addr           sdk.AccAddress
		expected       bool
		expGetAttrCall bool
	}{
		{
			name:     "empty state",
			marketID: 1,
			addr:     sdk.AccAddress("empty_state_addr____"),
			expected: true,
		},
		{
			name: "no req attrs, addr without any attributes",
			setup: func() {
				store := s.getStore()
				setter(store, 7, []string{"bb.aa"})
				setter(store, 9, []string{"yy.zz", "*.lm.no"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, []string{"bb.aa"}, "").
				WithGetAllAttributesAddrResult(addr2, nil, "").
				WithGetAllAttributesAddrResult(addr3, []string{"jk.lm.nl", "yy.zz"}, ""),
			marketID: 8,
			addr:     addr2,
			expected: true,
		},
		{
			name: "no req attrs, addr with some attributes",
			setup: func() {
				store := s.getStore()
				setter(store, 7, []string{"bb.aa"})
				setter(store, 9, []string{"yy.zz", "*.lm.no"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, []string{"bb.aa"}, "").
				WithGetAllAttributesAddrResult(addr2, []string{"left", "right"}, "").
				WithGetAllAttributesAddrResult(addr3, []string{"jk.lm.nl", "yy.zz"}, ""),
			marketID: 8,
			addr:     addr2,
			expected: true,
		},
		{
			name: "error getting attributes",
			setup: func() {
				setter(s.getStore(), 4, []string{"bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, nil, "injected test error"),
			marketID:       4,
			addr:           addr1,
			expected:       false,
			expGetAttrCall: true,
		},
		{
			name: "one req attr, acc has",
			setup: func() {
				setter(s.getStore(), 88, []string{"bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr2, []string{"yy.zz", "bb.aa", "lm.no"}, ""),
			marketID:       88,
			addr:           addr2,
			expected:       true,
			expGetAttrCall: true,
		},
		{
			name: "one req attr, acc does not have",
			setup: func() {
				setter(s.getStore(), 88, []string{"bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr2, []string{"yy.zz", "cc.bb.aa", "lm.no"}, ""),
			marketID:       88,
			addr:           addr2,
			expected:       false,
			expGetAttrCall: true,
		},
		{
			name: "one req attr with wildcard, acc has",
			setup: func() {
				setter(s.getStore(), 42, []string{"*.lm.no"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr2, []string{"yy.zz", "cc.bb.aa", "jk.lm.no"}, ""),
			marketID:       42,
			addr:           addr2,
			expected:       true,
			expGetAttrCall: true,
		},
		{
			name: "one req attr with wildcard, acc has two that match",
			setup: func() {
				setter(s.getStore(), 42, []string{"*.lm.no"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr2, []string{"yy.zz", "ab.cd.lm.no", "cc.bb.aa", "jk.lm.no"}, ""),
			marketID:       42,
			addr:           addr2,
			expected:       true,
			expGetAttrCall: true,
		},
		{
			name: "one req attr with wildcard, acc does not have",
			setup: func() {
				setter(s.getStore(), 42, []string{"*.lm.no"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr2, []string{"yy.zz", "cc.bb.aa", "lm.no"}, ""),
			marketID:       42,
			addr:           addr2,
			expected:       false,
			expGetAttrCall: true,
		},
		{
			name: "two req attr, acc has neither",
			setup: func() {
				setter(s.getStore(), 123, []string{"one.bb.aa", "two.bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, []string{"one.bb.aa", "two.bb.aa"}, "").
				WithGetAllAttributesAddrResult(addr2, []string{"one.yy.zz", "two.yy.zz"}, "").
				WithGetAllAttributesAddrResult(addr3, []string{"one.bb.aa"}, ""),
			marketID:       123,
			addr:           addr2,
			expected:       false,
			expGetAttrCall: true,
		},
		{
			name: "two req attr, acc has just first",
			setup: func() {
				setter(s.getStore(), 123, []string{"one.bb.aa", "two.bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, []string{"one.bb.aa", "two.bb.aa"}, "").
				WithGetAllAttributesAddrResult(addr2, []string{"one.yy.zz", "two.yy.zz"}, "").
				WithGetAllAttributesAddrResult(addr3, []string{"one.bb.aa"}, ""),
			marketID:       123,
			addr:           addr3,
			expected:       false,
			expGetAttrCall: true,
		},
		{
			name: "two req attr, acc has just second",
			setup: func() {
				setter(s.getStore(), 123, []string{"one.bb.aa", "two.bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, []string{"one.bb.aa", "two.bb.aa"}, "").
				WithGetAllAttributesAddrResult(addr2, []string{"one.yy.zz", "two.yy.zz"}, "").
				WithGetAllAttributesAddrResult(addr3, []string{"two.bb.aa"}, ""),
			marketID:       123,
			addr:           addr3,
			expected:       false,
			expGetAttrCall: true,
		},
		{
			name: "two req attr, acc has both, same order",
			setup: func() {
				setter(s.getStore(), 123, []string{"one.bb.aa", "two.bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, []string{"one.bb.aa", "two.bb.aa"}, "").
				WithGetAllAttributesAddrResult(addr2, []string{"one.yy.zz", "two.yy.zz"}, "").
				WithGetAllAttributesAddrResult(addr3, []string{"two.bb.aa"}, ""),
			marketID:       123,
			addr:           addr1,
			expected:       true,
			expGetAttrCall: true,
		},
		{
			name: "two req attr, acc has both, opposite order",
			setup: func() {
				setter(s.getStore(), 123, []string{"one.bb.aa", "two.bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, []string{"two.bb.aa", "one.bb.aa"}, "").
				WithGetAllAttributesAddrResult(addr2, []string{"one.yy.zz", "two.yy.zz"}, "").
				WithGetAllAttributesAddrResult(addr3, []string{"two.bb.aa"}, ""),
			marketID:       123,
			addr:           addr1,
			expected:       true,
			expGetAttrCall: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var expCalls AttributeCalls
			if tc.expGetAttrCall {
				expCalls.GetAllAttributesAddr = append(expCalls.GetAllAttributesAddr, tc.addr)
			}

			if tc.attrKeeper == nil {
				tc.attrKeeper = NewMockAttributeKeeper()
			}
			kpr := s.k.WithAttributeKeeper(tc.attrKeeper)

			var actual bool
			testFunc := func() {
				actual = kpr.CanCreateAsk(s.ctx, tc.marketID, tc.addr)
			}
			s.Require().NotPanics(testFunc, "CanCreateAsk(%d, %q)", tc.marketID, string(tc.addr))
			s.Assert().Equal(tc.expected, actual, "CanCreateAsk(%d, %q) result", tc.marketID, string(tc.addr))
			s.assertAttributeKeeperCalls(tc.attrKeeper, expCalls, "CanCreateAsk(%d, %q)", tc.marketID, string(tc.addr))
		})
	}
}

func (s *TestSuite) TestKeeper_CanCreateBid() {
	setter := keeper.SetReqAttrsBid
	addr1 := sdk.AccAddress("addr_one____________")
	addr2 := sdk.AccAddress("addr_two____________")
	addr3 := sdk.AccAddress("addr_three__________")

	tests := []struct {
		name           string
		setup          func()
		attrKeeper     *MockAttributeKeeper
		marketID       uint32
		addr           sdk.AccAddress
		expected       bool
		expGetAttrCall bool
	}{
		{
			name:     "empty state",
			marketID: 1,
			addr:     sdk.AccAddress("empty_state_addr____"),
			expected: true,
		},
		{
			name: "no req attrs, addr without any attributes",
			setup: func() {
				store := s.getStore()
				setter(store, 7, []string{"bb.aa"})
				setter(store, 9, []string{"yy.zz", "*.lm.no"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, []string{"bb.aa"}, "").
				WithGetAllAttributesAddrResult(addr2, nil, "").
				WithGetAllAttributesAddrResult(addr3, []string{"jk.lm.nl", "yy.zz"}, ""),
			marketID: 8,
			addr:     addr2,
			expected: true,
		},
		{
			name: "no req attrs, addr with some attributes",
			setup: func() {
				store := s.getStore()
				setter(store, 7, []string{"bb.aa"})
				setter(store, 9, []string{"yy.zz", "*.lm.no"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, []string{"bb.aa"}, "").
				WithGetAllAttributesAddrResult(addr2, []string{"left", "right"}, "").
				WithGetAllAttributesAddrResult(addr3, []string{"jk.lm.nl", "yy.zz"}, ""),
			marketID: 8,
			addr:     addr2,
			expected: true,
		},
		{
			name: "error getting attributes",
			setup: func() {
				setter(s.getStore(), 4, []string{"bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, nil, "injected test error"),
			marketID:       4,
			addr:           addr1,
			expected:       false,
			expGetAttrCall: true,
		},
		{
			name: "one req attr, acc has",
			setup: func() {
				setter(s.getStore(), 88, []string{"bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr2, []string{"yy.zz", "bb.aa", "lm.no"}, ""),
			marketID:       88,
			addr:           addr2,
			expected:       true,
			expGetAttrCall: true,
		},
		{
			name: "one req attr, acc does not have",
			setup: func() {
				setter(s.getStore(), 88, []string{"bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr2, []string{"yy.zz", "cc.bb.aa", "lm.no"}, ""),
			marketID:       88,
			addr:           addr2,
			expected:       false,
			expGetAttrCall: true,
		},
		{
			name: "one req attr with wildcard, acc has",
			setup: func() {
				setter(s.getStore(), 42, []string{"*.lm.no"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr2, []string{"yy.zz", "cc.bb.aa", "jk.lm.no"}, ""),
			marketID:       42,
			addr:           addr2,
			expected:       true,
			expGetAttrCall: true,
		},
		{
			name: "one req attr with wildcard, acc has two that match",
			setup: func() {
				setter(s.getStore(), 42, []string{"*.lm.no"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr2, []string{"yy.zz", "ab.cd.lm.no", "cc.bb.aa", "jk.lm.no"}, ""),
			marketID:       42,
			addr:           addr2,
			expected:       true,
			expGetAttrCall: true,
		},
		{
			name: "one req attr with wildcard, acc does not have",
			setup: func() {
				setter(s.getStore(), 42, []string{"*.lm.no"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr2, []string{"yy.zz", "cc.bb.aa", "lm.no"}, ""),
			marketID:       42,
			addr:           addr2,
			expected:       false,
			expGetAttrCall: true,
		},
		{
			name: "two req attr, acc has neither",
			setup: func() {
				setter(s.getStore(), 123, []string{"one.bb.aa", "two.bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, []string{"one.bb.aa", "two.bb.aa"}, "").
				WithGetAllAttributesAddrResult(addr2, []string{"one.yy.zz", "two.yy.zz"}, "").
				WithGetAllAttributesAddrResult(addr3, []string{"one.bb.aa"}, ""),
			marketID:       123,
			addr:           addr2,
			expected:       false,
			expGetAttrCall: true,
		},
		{
			name: "two req attr, acc has just first",
			setup: func() {
				setter(s.getStore(), 123, []string{"one.bb.aa", "two.bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, []string{"one.bb.aa", "two.bb.aa"}, "").
				WithGetAllAttributesAddrResult(addr2, []string{"one.yy.zz", "two.yy.zz"}, "").
				WithGetAllAttributesAddrResult(addr3, []string{"one.bb.aa"}, ""),
			marketID:       123,
			addr:           addr3,
			expected:       false,
			expGetAttrCall: true,
		},
		{
			name: "two req attr, acc has just second",
			setup: func() {
				setter(s.getStore(), 123, []string{"one.bb.aa", "two.bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, []string{"one.bb.aa", "two.bb.aa"}, "").
				WithGetAllAttributesAddrResult(addr2, []string{"one.yy.zz", "two.yy.zz"}, "").
				WithGetAllAttributesAddrResult(addr3, []string{"two.bb.aa"}, ""),
			marketID:       123,
			addr:           addr3,
			expected:       false,
			expGetAttrCall: true,
		},
		{
			name: "two req attr, acc has both, same order",
			setup: func() {
				setter(s.getStore(), 123, []string{"one.bb.aa", "two.bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, []string{"one.bb.aa", "two.bb.aa"}, "").
				WithGetAllAttributesAddrResult(addr2, []string{"one.yy.zz", "two.yy.zz"}, "").
				WithGetAllAttributesAddrResult(addr3, []string{"two.bb.aa"}, ""),
			marketID:       123,
			addr:           addr1,
			expected:       true,
			expGetAttrCall: true,
		},
		{
			name: "two req attr, acc has both, opposite order",
			setup: func() {
				setter(s.getStore(), 123, []string{"one.bb.aa", "two.bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, []string{"two.bb.aa", "one.bb.aa"}, "").
				WithGetAllAttributesAddrResult(addr2, []string{"one.yy.zz", "two.yy.zz"}, "").
				WithGetAllAttributesAddrResult(addr3, []string{"two.bb.aa"}, ""),
			marketID:       123,
			addr:           addr1,
			expected:       true,
			expGetAttrCall: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var expCalls AttributeCalls
			if tc.expGetAttrCall {
				expCalls.GetAllAttributesAddr = append(expCalls.GetAllAttributesAddr, tc.addr)
			}

			if tc.attrKeeper == nil {
				tc.attrKeeper = NewMockAttributeKeeper()
			}
			kpr := s.k.WithAttributeKeeper(tc.attrKeeper)

			var actual bool
			testFunc := func() {
				actual = kpr.CanCreateBid(s.ctx, tc.marketID, tc.addr)
			}
			s.Require().NotPanics(testFunc, "CanCreateBid(%d, %q)", tc.marketID, string(tc.addr))
			s.Assert().Equal(tc.expected, actual, "CanCreateBid(%d, %q) result", tc.marketID, string(tc.addr))
			s.assertAttributeKeeperCalls(tc.attrKeeper, expCalls, "CanCreateBid(%d, %q)", tc.marketID, string(tc.addr))
		})
	}
}

func (s *TestSuite) TestKeeper_CanCreateCommitment() {
	setter := keeper.SetReqAttrsCommitment
	addr1 := sdk.AccAddress("addr_one____________")
	addr2 := sdk.AccAddress("addr_two____________")
	addr3 := sdk.AccAddress("addr_three__________")

	tests := []struct {
		name           string
		setup          func()
		attrKeeper     *MockAttributeKeeper
		marketID       uint32
		addr           sdk.AccAddress
		expected       bool
		expGetAttrCall bool
	}{
		{
			name:     "empty state",
			marketID: 1,
			addr:     sdk.AccAddress("empty_state_addr____"),
			expected: true,
		},
		{
			name: "no req attrs, addr without any attributes",
			setup: func() {
				store := s.getStore()
				setter(store, 7, []string{"bb.aa"})
				setter(store, 9, []string{"yy.zz", "*.lm.no"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, []string{"bb.aa"}, "").
				WithGetAllAttributesAddrResult(addr2, nil, "").
				WithGetAllAttributesAddrResult(addr3, []string{"jk.lm.nl", "yy.zz"}, ""),
			marketID: 8,
			addr:     addr2,
			expected: true,
		},
		{
			name: "no req attrs, addr with some attributes",
			setup: func() {
				store := s.getStore()
				setter(store, 7, []string{"bb.aa"})
				setter(store, 9, []string{"yy.zz", "*.lm.no"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, []string{"bb.aa"}, "").
				WithGetAllAttributesAddrResult(addr2, []string{"left", "right"}, "").
				WithGetAllAttributesAddrResult(addr3, []string{"jk.lm.nl", "yy.zz"}, ""),
			marketID: 8,
			addr:     addr2,
			expected: true,
		},
		{
			name: "error getting attributes",
			setup: func() {
				setter(s.getStore(), 4, []string{"bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, nil, "injected test error"),
			marketID:       4,
			addr:           addr1,
			expected:       false,
			expGetAttrCall: true,
		},
		{
			name: "one req attr, acc has",
			setup: func() {
				setter(s.getStore(), 88, []string{"bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr2, []string{"yy.zz", "bb.aa", "lm.no"}, ""),
			marketID:       88,
			addr:           addr2,
			expected:       true,
			expGetAttrCall: true,
		},
		{
			name: "one req attr, acc does not have",
			setup: func() {
				setter(s.getStore(), 88, []string{"bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr2, []string{"yy.zz", "cc.bb.aa", "lm.no"}, ""),
			marketID:       88,
			addr:           addr2,
			expected:       false,
			expGetAttrCall: true,
		},
		{
			name: "one req attr with wildcard, acc has",
			setup: func() {
				setter(s.getStore(), 42, []string{"*.lm.no"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr2, []string{"yy.zz", "cc.bb.aa", "jk.lm.no"}, ""),
			marketID:       42,
			addr:           addr2,
			expected:       true,
			expGetAttrCall: true,
		},
		{
			name: "one req attr with wildcard, acc has two that match",
			setup: func() {
				setter(s.getStore(), 42, []string{"*.lm.no"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr2, []string{"yy.zz", "ab.cd.lm.no", "cc.bb.aa", "jk.lm.no"}, ""),
			marketID:       42,
			addr:           addr2,
			expected:       true,
			expGetAttrCall: true,
		},
		{
			name: "one req attr with wildcard, acc does not have",
			setup: func() {
				setter(s.getStore(), 42, []string{"*.lm.no"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr2, []string{"yy.zz", "cc.bb.aa", "lm.no"}, ""),
			marketID:       42,
			addr:           addr2,
			expected:       false,
			expGetAttrCall: true,
		},
		{
			name: "two req attr, acc has neither",
			setup: func() {
				setter(s.getStore(), 123, []string{"one.bb.aa", "two.bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, []string{"one.bb.aa", "two.bb.aa"}, "").
				WithGetAllAttributesAddrResult(addr2, []string{"one.yy.zz", "two.yy.zz"}, "").
				WithGetAllAttributesAddrResult(addr3, []string{"one.bb.aa"}, ""),
			marketID:       123,
			addr:           addr2,
			expected:       false,
			expGetAttrCall: true,
		},
		{
			name: "two req attr, acc has just first",
			setup: func() {
				setter(s.getStore(), 123, []string{"one.bb.aa", "two.bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, []string{"one.bb.aa", "two.bb.aa"}, "").
				WithGetAllAttributesAddrResult(addr2, []string{"one.yy.zz", "two.yy.zz"}, "").
				WithGetAllAttributesAddrResult(addr3, []string{"one.bb.aa"}, ""),
			marketID:       123,
			addr:           addr3,
			expected:       false,
			expGetAttrCall: true,
		},
		{
			name: "two req attr, acc has just second",
			setup: func() {
				setter(s.getStore(), 123, []string{"one.bb.aa", "two.bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, []string{"one.bb.aa", "two.bb.aa"}, "").
				WithGetAllAttributesAddrResult(addr2, []string{"one.yy.zz", "two.yy.zz"}, "").
				WithGetAllAttributesAddrResult(addr3, []string{"two.bb.aa"}, ""),
			marketID:       123,
			addr:           addr3,
			expected:       false,
			expGetAttrCall: true,
		},
		{
			name: "two req attr, acc has both, same order",
			setup: func() {
				setter(s.getStore(), 123, []string{"one.bb.aa", "two.bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, []string{"one.bb.aa", "two.bb.aa"}, "").
				WithGetAllAttributesAddrResult(addr2, []string{"one.yy.zz", "two.yy.zz"}, "").
				WithGetAllAttributesAddrResult(addr3, []string{"two.bb.aa"}, ""),
			marketID:       123,
			addr:           addr1,
			expected:       true,
			expGetAttrCall: true,
		},
		{
			name: "two req attr, acc has both, opposite order",
			setup: func() {
				setter(s.getStore(), 123, []string{"one.bb.aa", "two.bb.aa"})
			},
			attrKeeper: NewMockAttributeKeeper().
				WithGetAllAttributesAddrResult(addr1, []string{"two.bb.aa", "one.bb.aa"}, "").
				WithGetAllAttributesAddrResult(addr2, []string{"one.yy.zz", "two.yy.zz"}, "").
				WithGetAllAttributesAddrResult(addr3, []string{"two.bb.aa"}, ""),
			marketID:       123,
			addr:           addr1,
			expected:       true,
			expGetAttrCall: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var expCalls AttributeCalls
			if tc.expGetAttrCall {
				expCalls.GetAllAttributesAddr = append(expCalls.GetAllAttributesAddr, tc.addr)
			}

			if tc.attrKeeper == nil {
				tc.attrKeeper = NewMockAttributeKeeper()
			}
			kpr := s.k.WithAttributeKeeper(tc.attrKeeper)

			var actual bool
			testFunc := func() {
				actual = kpr.CanCreateCommitment(s.ctx, tc.marketID, tc.addr)
			}
			s.Require().NotPanics(testFunc, "CanCreateCommitment(%d, %q)", tc.marketID, string(tc.addr))
			s.Assert().Equal(tc.expected, actual, "CanCreateCommitment(%d, %q) result", tc.marketID, string(tc.addr))
			s.assertAttributeKeeperCalls(tc.attrKeeper, expCalls, "CanCreateCommitment(%d, %q)", tc.marketID, string(tc.addr))
		})
	}

}

func (s *TestSuite) TestKeeper_UpdateReqAttrs() {
	tests := []struct {
		name     string
		setup    func()
		msg      *exchange.MsgMarketManageReqAttrsRequest
		expAsk   []string
		expBid   []string
		expCom   []string
		expErr   string
		expPanic string
	}{
		// panics and errors.
		{
			name:     "nil msg",
			setup:    nil,
			msg:      nil,
			expPanic: "runtime error: invalid memory address or nil pointer dereference",
		},
		{
			name: "invalid attrs",
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:                    "admin_addr_str",
				MarketId:                 1,
				CreateAskToAdd:           []string{"three-dashes-not-allowed", "this.one.is.okay", "bad,punctuation"},
				CreateAskToRemove:        []string{"internal spaces are bad"}, // no error from this.
				CreateBidToAdd:           []string{"twodashes-notallowed-either", "this.one.is.also.okay", "really*bad,punctuation"},
				CreateBidToRemove:        []string{"what&are*you(doing)?"}, // no error from this.
				CreateCommitmentToAdd:    []string{"no&ampersands", "this.one.is.still.okay", "horrible|punctuation"},
				CreateCommitmentToRemove: []string{"[oops>"}, // no error from this.
			},
			expErr: s.joinErrs(
				"invalid attribute \"three-dashes-not-allowed\"",
				"invalid attribute \"bad,punctuation\"",
				"invalid attribute \"twodashes-notallowed-either\"",
				"invalid attribute \"really*bad,punctuation\"",
				"invalid attribute \"no&ampersands\"",
				"invalid attribute \"horrible|punctuation\"",
			),
		},
		{
			name: "remove create-ask that is not required",
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:             "admin_addr_str",
				MarketId:          1,
				CreateAskToRemove: []string{"not.req"},
			},
			expErr: "cannot remove create-ask required attribute \"not.req\": attribute not currently required",
		},
		{
			name: "remove create-bid that is not required",
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:             "admin_addr_str",
				MarketId:          1,
				CreateBidToRemove: []string{"not.req"},
			},
			expErr: "cannot remove create-bid required attribute \"not.req\": attribute not currently required",
		},
		{
			name: "remove create-commitment that is not required",
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:                    "admin_addr_str",
				MarketId:                 1,
				CreateCommitmentToRemove: []string{"not.req"},
			},
			expErr: "cannot remove create-commitment required attribute \"not.req\": attribute not currently required",
		},
		{
			name: "add create-ask that is already required",
			setup: func() {
				keeper.SetReqAttrsAsk(s.getStore(), 7, []string{"already.req"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:          "admin_addr_str",
				MarketId:       7,
				CreateAskToAdd: []string{"already.req"},
			},
			expErr: "cannot add create-ask required attribute \"already.req\": attribute already required",
		},
		{
			name: "add create-bid that is already required",
			setup: func() {
				keeper.SetReqAttrsBid(s.getStore(), 4, []string{"already.req"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:          "admin_addr_str",
				MarketId:       4,
				CreateBidToAdd: []string{"already.req"},
			},
			expErr: "cannot add create-bid required attribute \"already.req\": attribute already required",
		},
		{
			name: "add create-commitment that is already required",
			setup: func() {
				keeper.SetReqAttrsCommitment(s.getStore(), 12, []string{"already.req"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:                 "admin_addr_str",
				MarketId:              12,
				CreateCommitmentToAdd: []string{"already.req"},
			},
			expErr: "cannot add create-commitment required attribute \"already.req\": attribute already required",
		},
		{
			name: "multiple errors",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 3, []string{"one.ask", "two.ask", "three.ask", "four.ask"})
				keeper.SetReqAttrsBid(store, 3, []string{"one.bid", "two.bid", "three.bid", "four.bid"})
				keeper.SetReqAttrsCommitment(store, 3, []string{"one.com", "two.com", "three.com", "four.com"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:                    "addr_str_of_admin",
				MarketId:                 3,
				CreateAskToAdd:           []string{"two.ask", "three .ask", "five.ask"},
				CreateAskToRemove:        []string{" four .ask", "five.ask", "six . ask"},
				CreateBidToAdd:           []string{"two.bid ", " three.bid", "five.   bid"},
				CreateBidToRemove:        []string{"four. bid ", "five . bid", "six.bid"},
				CreateCommitmentToAdd:    []string{"two. com", "three.com", " five.com "},
				CreateCommitmentToRemove: []string{"four. com ", "five .com", "six. com"},
			},
			expErr: s.joinErrs(
				"cannot remove create-ask required attribute \"five.ask\": attribute not currently required",
				"cannot remove create-ask required attribute \"six.ask\": attribute not currently required",
				"cannot add create-ask required attribute \"two.ask\": attribute already required",
				"cannot add create-ask required attribute \"three.ask\": attribute already required",
				"cannot remove create-bid required attribute \"five.bid\": attribute not currently required",
				"cannot remove create-bid required attribute \"six.bid\": attribute not currently required",
				"cannot add create-bid required attribute \"two.bid\": attribute already required",
				"cannot add create-bid required attribute \"three.bid\": attribute already required",
				"cannot remove create-commitment required attribute \"five.com\": attribute not currently required",
				"cannot remove create-commitment required attribute \"six.com\": attribute not currently required",
				"cannot add create-commitment required attribute \"two.com\": attribute already required",
				"cannot add create-commitment required attribute \"three.com\": attribute already required",
			),
		},

		// just create-ask manipulation.
		{
			name: "remove one create-ask from one",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 9, []string{"ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 9, []string{"bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 9, []string{"com.can.create.bananas"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:             "admin_addr_str",
				MarketId:          9,
				CreateAskToRemove: []string{"ask.can.create.bananas"},
			},
			expAsk: nil,
			expBid: []string{"bid.can.create.bananas"},
			expCom: []string{"com.can.create.bananas"},
		},
		{
			name: "remove one create-ask from two",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 9, []string{"ask.can.create.bananas", "also.ask.okay"})
				keeper.SetReqAttrsBid(store, 9, []string{"bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 9, []string{"com.can.create.bananas"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:             "admin_addr_str",
				MarketId:          9,
				CreateAskToRemove: []string{"also.ask.okay"},
			},
			expAsk: []string{"ask.can.create.bananas"},
			expBid: []string{"bid.can.create.bananas"},
			expCom: []string{"com.can.create.bananas"},
		},
		{
			name: "remove one create-ask with wildcard",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 9, []string{
					"ask.can.create.bananas", "one.ask.can.create.bananas",
					"*.ask.can.create.bananas", "two.ask.can.create.bananas",
				})
				keeper.SetReqAttrsBid(store, 9, []string{"bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 9, []string{"com.can.create.bananas"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:             "admin_addr_str",
				MarketId:          9,
				CreateAskToRemove: []string{"*.ask.can.create.bananas"},
			},
			expAsk: []string{"ask.can.create.bananas", "one.ask.can.create.bananas", "two.ask.can.create.bananas"},
			expBid: []string{"bid.can.create.bananas"},
			expCom: []string{"com.can.create.bananas"},
		},
		{
			name: "remove last two create-ask",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 55, []string{"one.ask.can.create.bananas", "two.ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 55, []string{"one.bid.can.create.bananas", "two.bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 55, []string{"one.com.can.create.bananas", "two.com.can.create.bananas"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:             "admin_addr_str",
				MarketId:          55,
				CreateAskToRemove: []string{"two.ask.can.create.bananas", "one.ask.can.create.bananas"},
			},
			expAsk: nil,
			expBid: []string{"one.bid.can.create.bananas", "two.bid.can.create.bananas"},
			expCom: []string{"one.com.can.create.bananas", "two.com.can.create.bananas"},
		},
		{
			name: "add one create-ask to empty",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsBid(store, 1, []string{"bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 1, []string{"com.can.create.bananas"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:          "admin_addr_str",
				MarketId:       1,
				CreateAskToAdd: []string{"ask.can.create.bananas"},
			},
			expAsk: []string{"ask.can.create.bananas"},
			expBid: []string{"bid.can.create.bananas"},
			expCom: []string{"com.can.create.bananas"},
		},
		{
			name: "add one create-ask to existing",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 1, []string{"ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 1, []string{"bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 1, []string{"com.can.create.bananas"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:          "admin_addr_str",
				MarketId:       1,
				CreateAskToAdd: []string{"*.ask.can.create.bananas"},
			},
			expAsk: []string{"ask.can.create.bananas", "*.ask.can.create.bananas"},
			expBid: []string{"bid.can.create.bananas"},
			expCom: []string{"com.can.create.bananas"},
		},
		{
			name: "remove one, add diff create-ask",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 4, []string{"four.ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 4, []string{"four.bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 4, []string{"four.com.can.create.bananas"})
				keeper.SetReqAttrsAsk(store, 5, []string{"five.ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 5, []string{"five.bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 5, []string{"five.com.can.create.bananas"})
				keeper.SetReqAttrsAsk(store, 6, []string{"six.ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 6, []string{"six.bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 6, []string{"six.com.can.create.bananas"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:             "admin_addr_str",
				MarketId:          5,
				CreateAskToAdd:    []string{"*.ask.can.create.bananas"},
				CreateAskToRemove: []string{"five.ask.can.create.bananas"},
			},
			expAsk: []string{"*.ask.can.create.bananas"},
			expBid: []string{"five.bid.can.create.bananas"},
			expCom: []string{"five.com.can.create.bananas"},
		},

		// just create-bid manipulation.
		{
			name: "remove one create-bid from one",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 9, []string{"ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 9, []string{"bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 9, []string{"com.can.create.bananas"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:             "admin_addr_str",
				MarketId:          9,
				CreateBidToRemove: []string{"bid.can.create.bananas"},
			},
			expAsk: []string{"ask.can.create.bananas"},
			expBid: nil,
			expCom: []string{"com.can.create.bananas"},
		},
		{
			name: "remove one create-bid from two",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 9, []string{"ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 9, []string{"bid.can.create.bananas", "also.bid.okay"})
				keeper.SetReqAttrsCommitment(store, 9, []string{"com.can.create.bananas"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:             "admin_addr_str",
				MarketId:          9,
				CreateBidToRemove: []string{"also.bid.okay"},
			},
			expAsk: []string{"ask.can.create.bananas"},
			expBid: []string{"bid.can.create.bananas"},
			expCom: []string{"com.can.create.bananas"},
		},
		{
			name: "remove one create-bid with wildcard",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 9, []string{"ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 9, []string{
					"bid.can.create.bananas", "one.bid.can.create.bananas",
					"*.bid.can.create.bananas", "two.bid.can.create.bananas",
				})
				keeper.SetReqAttrsCommitment(store, 9, []string{"com.can.create.bananas"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:             "admin_addr_str",
				MarketId:          9,
				CreateBidToRemove: []string{"*.bid.can.create.bananas"},
			},
			expAsk: []string{"ask.can.create.bananas"},
			expBid: []string{"bid.can.create.bananas", "one.bid.can.create.bananas", "two.bid.can.create.bananas"},
			expCom: []string{"com.can.create.bananas"},
		},
		{
			name: "remove last two create-bid",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 55, []string{"one.ask.can.create.bananas", "two.ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 55, []string{"one.bid.can.create.bananas", "two.bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 55, []string{"one.com.can.create.bananas", "two.com.can.create.bananas"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:             "admin_addr_str",
				MarketId:          55,
				CreateBidToRemove: []string{"two.bid.can.create.bananas", "one.bid.can.create.bananas"},
			},
			expAsk: []string{"one.ask.can.create.bananas", "two.ask.can.create.bananas"},
			expBid: nil,
			expCom: []string{"one.com.can.create.bananas", "two.com.can.create.bananas"},
		},
		{
			name: "add one create-bid to empty",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 1, []string{"ask.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 1, []string{"com.can.create.bananas"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:          "admin_addr_str",
				MarketId:       1,
				CreateBidToAdd: []string{"bid.can.create.bananas"},
			},
			expAsk: []string{"ask.can.create.bananas"},
			expBid: []string{"bid.can.create.bananas"},
			expCom: []string{"com.can.create.bananas"},
		},
		{
			name: "add one create-bid to existing",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 1, []string{"ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 1, []string{"bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 1, []string{"com.can.create.bananas"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:          "admin_addr_str",
				MarketId:       1,
				CreateBidToAdd: []string{"*.bid.can.create.bananas"},
			},
			expAsk: []string{"ask.can.create.bananas"},
			expBid: []string{"bid.can.create.bananas", "*.bid.can.create.bananas"},
			expCom: []string{"com.can.create.bananas"},
		},
		{
			name: "remove one, add diff create-bid",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 4, []string{"four.ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 4, []string{"four.bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 4, []string{"four.com.can.create.bananas"})
				keeper.SetReqAttrsAsk(store, 5, []string{"five.ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 5, []string{"five.bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 5, []string{"five.com.can.create.bananas"})
				keeper.SetReqAttrsAsk(store, 6, []string{"six.ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 6, []string{"six.bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 6, []string{"six.com.can.create.bananas"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:             "admin_addr_str",
				MarketId:          5,
				CreateBidToAdd:    []string{"*.bid.can.create.bananas"},
				CreateBidToRemove: []string{"five.bid.can.create.bananas"},
			},
			expAsk: []string{"five.ask.can.create.bananas"},
			expBid: []string{"*.bid.can.create.bananas"},
			expCom: []string{"five.com.can.create.bananas"},
		},

		// just create-commitment manipulation.
		{
			name: "remove one create-commitment from one",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 9, []string{"ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 9, []string{"bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 9, []string{"com.can.create.bananas"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:                    "admin_addr_str",
				MarketId:                 9,
				CreateCommitmentToRemove: []string{"com.can.create.bananas"},
			},
			expAsk: []string{"ask.can.create.bananas"},
			expBid: []string{"bid.can.create.bananas"},
			expCom: nil,
		},
		{
			name: "remove one create-commitment from two",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 9, []string{"ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 9, []string{"bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 9, []string{"com.can.create.bananas", "also.com.okay"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:                    "admin_addr_str",
				MarketId:                 9,
				CreateCommitmentToRemove: []string{"also.com.okay"},
			},
			expAsk: []string{"ask.can.create.bananas"},
			expBid: []string{"bid.can.create.bananas"},
			expCom: []string{"com.can.create.bananas"},
		},
		{
			name: "remove one create-commitment with wildcard",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 9, []string{"ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 9, []string{"bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 9, []string{
					"com.can.create.bananas", "one.com.can.create.bananas",
					"*.com.can.create.bananas", "two.com.can.create.bananas",
				})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:                    "admin_addr_str",
				MarketId:                 9,
				CreateCommitmentToRemove: []string{"*.com.can.create.bananas"},
			},
			expAsk: []string{"ask.can.create.bananas"},
			expBid: []string{"bid.can.create.bananas"},
			expCom: []string{"com.can.create.bananas", "one.com.can.create.bananas", "two.com.can.create.bananas"},
		},
		{
			name: "remove last two create-commitment",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 55, []string{"one.ask.can.create.bananas", "two.ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 55, []string{"one.bid.can.create.bananas", "two.bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 55, []string{"one.com.can.create.bananas", "two.com.can.create.bananas"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:                    "admin_addr_str",
				MarketId:                 55,
				CreateCommitmentToRemove: []string{"two.com.can.create.bananas", "one.com.can.create.bananas"},
			},
			expAsk: []string{"one.ask.can.create.bananas", "two.ask.can.create.bananas"},
			expBid: []string{"one.bid.can.create.bananas", "two.bid.can.create.bananas"},
			expCom: nil,
		},
		{
			name: "add one create-commitment to empty",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 1, []string{"ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 1, []string{"bid.can.create.bananas"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:                 "admin_addr_str",
				MarketId:              1,
				CreateCommitmentToAdd: []string{"com.can.create.bananas"},
			},
			expAsk: []string{"ask.can.create.bananas"},
			expBid: []string{"bid.can.create.bananas"},
			expCom: []string{"com.can.create.bananas"},
		},
		{
			name: "add one create-commitment to existing",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 1, []string{"ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 1, []string{"bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 1, []string{"com.can.create.bananas"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:                 "admin_addr_str",
				MarketId:              1,
				CreateCommitmentToAdd: []string{"*.com.can.create.bananas"},
			},
			expAsk: []string{"ask.can.create.bananas"},
			expBid: []string{"bid.can.create.bananas"},
			expCom: []string{"com.can.create.bananas", "*.com.can.create.bananas"},
		},
		{
			name: "remove one, add diff create-commitment",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 4, []string{"four.ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 4, []string{"four.bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 4, []string{"four.com.can.create.bananas"})
				keeper.SetReqAttrsAsk(store, 5, []string{"five.ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 5, []string{"five.bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 5, []string{"five.com.can.create.bananas"})
				keeper.SetReqAttrsAsk(store, 6, []string{"six.ask.can.create.bananas"})
				keeper.SetReqAttrsBid(store, 6, []string{"six.bid.can.create.bananas"})
				keeper.SetReqAttrsCommitment(store, 6, []string{"six.com.can.create.bananas"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:                    "admin_addr_str",
				MarketId:                 5,
				CreateCommitmentToAdd:    []string{"*.com.can.create.bananas"},
				CreateCommitmentToRemove: []string{"five.com.can.create.bananas"},
			},
			expAsk: []string{"five.ask.can.create.bananas"},
			expBid: []string{"five.bid.can.create.bananas"},
			expCom: []string{"*.com.can.create.bananas"},
		},

		// manipulation of all.
		{
			name: "add and remove two of each",
			setup: func() {
				store := s.getStore()
				keeper.SetReqAttrsAsk(store, 2, []string{"one.ask", "two.ask", "three.ask"})
				keeper.SetReqAttrsBid(store, 2, []string{"one.bid", "two.bid", "three.bid"})
				keeper.SetReqAttrsCommitment(store, 2, []string{"one.com", "two.com", "three.com"})
			},
			msg: &exchange.MsgMarketManageReqAttrsRequest{
				Admin:                    "admin_addr_string",
				MarketId:                 2,
				CreateAskToAdd:           []string{"*.other", "four.ask"},
				CreateAskToRemove:        []string{"one.ask", "three.ask"},
				CreateBidToAdd:           []string{"*.other", "five.bid"},
				CreateBidToRemove:        []string{"three.bid", "two.bid"},
				CreateCommitmentToAdd:    []string{"*.other", "six.com"},
				CreateCommitmentToRemove: []string{"three.com", "two.com"},
			},
			expAsk: []string{"two.ask", "*.other", "four.ask"},
			expBid: []string{"one.bid", "*.other", "five.bid"},
			expCom: []string{"one.com", "*.other", "six.com"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var expEvents sdk.Events
			if len(tc.expPanic) == 0 && len(tc.expErr) == 0 {
				event := exchange.NewEventMarketReqAttrUpdated(tc.msg.MarketId, tc.msg.Admin)
				expEvents = append(expEvents, s.untypeEvent(event))
			}

			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = s.k.UpdateReqAttrs(ctx, tc.msg)
			}
			s.requirePanicEquals(testFunc, tc.expPanic, "UpdateReqAttrs")
			s.assertErrorValue(err, tc.expErr, "UpdateReqAttrs error")

			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "events emitted during UpdateReqAttrs")

			if len(tc.expPanic) > 0 || len(tc.expErr) > 0 {
				return
			}

			reqAttrAsk := s.k.GetReqAttrsAsk(s.ctx, tc.msg.MarketId)
			reqAttrBid := s.k.GetReqAttrsBid(s.ctx, tc.msg.MarketId)
			reqAttrCom := s.k.GetReqAttrsCommitment(s.ctx, tc.msg.MarketId)
			s.Assert().Equal(tc.expAsk, reqAttrAsk, "create-ask req attrs after UpdateReqAttrs")
			s.Assert().Equal(tc.expBid, reqAttrBid, "create-bid req attrs after UpdateReqAttrs")
			s.Assert().Equal(tc.expCom, reqAttrCom, "create-commitment req attrs after UpdateReqAttrs")
		})
	}
}

func (s *TestSuite) TestKeeper_GetMarketAccount() {
	baseAcc := func(marketID uint32) *authtypes.BaseAccount {
		return &authtypes.BaseAccount{
			Address:       exchange.GetMarketAddress(marketID).String(),
			PubKey:        nil,
			AccountNumber: uint64(marketID),
			Sequence:      uint64(marketID) * 2,
		}
	}
	marketAcc := func(marketID uint32) *exchange.MarketAccount {
		return &exchange.MarketAccount{
			BaseAccount: baseAcc(marketID),
			MarketId:    marketID,
			MarketDetails: exchange.MarketDetails{
				Name:        fmt.Sprintf("market %d name", marketID),
				Description: fmt.Sprintf("This is a description of market %d. It's not very helpful.", marketID),
				WebsiteUrl:  fmt.Sprintf("https://example.com/market/%d", marketID),
				IconUri:     fmt.Sprintf("https://icon.example.com/market/%d/small", marketID),
			},
		}
	}

	tests := []struct {
		name      string
		accKeeper *MockAccountKeeper
		marketID  uint32
		expected  *exchange.MarketAccount
	}{
		{
			name: "no account for market",
			accKeeper: NewMockAccountKeeper().
				WithGetAccountResult(exchange.GetMarketAddress(1), marketAcc(1)).
				WithGetAccountResult(exchange.GetMarketAddress(3), marketAcc(3)),
			marketID: 2,
			expected: nil,
		},
		{
			name: "not a market account",
			accKeeper: NewMockAccountKeeper().
				WithGetAccountResult(exchange.GetMarketAddress(1), marketAcc(1)).
				WithGetAccountResult(exchange.GetMarketAddress(2), baseAcc(2)).
				WithGetAccountResult(exchange.GetMarketAddress(3), marketAcc(3)),
			marketID: 2,
			expected: nil,
		},
		{
			name:      "market account 1",
			accKeeper: NewMockAccountKeeper().WithGetAccountResult(exchange.GetMarketAddress(1), marketAcc(1)),
			marketID:  1,
			expected:  marketAcc(1),
		},
		{
			name:      "market account 65,536",
			accKeeper: NewMockAccountKeeper().WithGetAccountResult(exchange.GetMarketAddress(65_536), marketAcc(65_536)),
			marketID:  65_536,
			expected:  marketAcc(65_536),
		},
		{
			name:      "market account max uint32",
			accKeeper: NewMockAccountKeeper().WithGetAccountResult(exchange.GetMarketAddress(4_294_967_295), marketAcc(4_294_967_295)),
			marketID:  4_294_967_295,
			expected:  marketAcc(4_294_967_295),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			expCalls := AccountCalls{GetAccount: []sdk.AccAddress{exchange.GetMarketAddress(tc.marketID)}}

			if tc.accKeeper == nil {
				tc.accKeeper = NewMockAccountKeeper()
			}
			kpr := s.k.WithAccountKeeper(tc.accKeeper)

			var actual *exchange.MarketAccount
			testFunc := func() {
				actual = kpr.GetMarketAccount(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "GetMarketAccount(%d)", tc.marketID)
			s.Assert().Equal(tc.expected, actual, "GetMarketAccount(%d) result", tc.marketID)
			s.assertAccountKeeperCalls(tc.accKeeper, expCalls, "GetMarketAccount(%d)", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_GetMarketDetails() {
	baseAcc := func(marketID uint32) *authtypes.BaseAccount {
		return &authtypes.BaseAccount{
			Address:       exchange.GetMarketAddress(marketID).String(),
			PubKey:        nil,
			AccountNumber: uint64(marketID),
			Sequence:      uint64(marketID) * 2,
		}
	}
	marketDeets := func(marketID uint32) *exchange.MarketDetails {
		return &exchange.MarketDetails{
			Name:        fmt.Sprintf("market %d name", marketID),
			Description: fmt.Sprintf("This is a description of market %d. It's not very helpful.", marketID),
			WebsiteUrl:  fmt.Sprintf("https://example.com/market/%d", marketID),
			IconUri:     fmt.Sprintf("https://icon.example.com/market/%d/small", marketID),
		}
	}
	marketAcc := func(marketID uint32) *exchange.MarketAccount {
		return &exchange.MarketAccount{
			BaseAccount:   baseAcc(marketID),
			MarketId:      marketID,
			MarketDetails: *marketDeets(marketID),
		}
	}

	tests := []struct {
		name      string
		accKeeper *MockAccountKeeper
		marketID  uint32
		expected  *exchange.MarketDetails
	}{
		{
			name: "no account for market",
			accKeeper: NewMockAccountKeeper().
				WithGetAccountResult(exchange.GetMarketAddress(1), marketAcc(1)).
				WithGetAccountResult(exchange.GetMarketAddress(3), marketAcc(3)),
			marketID: 2,
			expected: nil,
		},
		{
			name: "not a market account",
			accKeeper: NewMockAccountKeeper().
				WithGetAccountResult(exchange.GetMarketAddress(1), marketAcc(1)).
				WithGetAccountResult(exchange.GetMarketAddress(2), baseAcc(2)).
				WithGetAccountResult(exchange.GetMarketAddress(3), marketAcc(3)),
			marketID: 2,
			expected: nil,
		},
		{
			name:      "market account 1",
			accKeeper: NewMockAccountKeeper().WithGetAccountResult(exchange.GetMarketAddress(1), marketAcc(1)),
			marketID:  1,
			expected:  marketDeets(1),
		},
		{
			name:      "market account 65,536",
			accKeeper: NewMockAccountKeeper().WithGetAccountResult(exchange.GetMarketAddress(65_536), marketAcc(65_536)),
			marketID:  65_536,
			expected:  marketDeets(65_536),
		},
		{
			name:      "market account max uint32",
			accKeeper: NewMockAccountKeeper().WithGetAccountResult(exchange.GetMarketAddress(4_294_967_295), marketAcc(4_294_967_295)),
			marketID:  4_294_967_295,
			expected:  marketDeets(4_294_967_295),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			expCalls := AccountCalls{GetAccount: []sdk.AccAddress{exchange.GetMarketAddress(tc.marketID)}}

			if tc.accKeeper == nil {
				tc.accKeeper = NewMockAccountKeeper()
			}
			kpr := s.k.WithAccountKeeper(tc.accKeeper)

			var actual *exchange.MarketDetails
			testFunc := func() {
				actual = kpr.GetMarketDetails(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "GetMarketDetails(%d)", tc.marketID)
			s.Assert().Equal(tc.expected, actual, "GetMarketDetails(%d) result", tc.marketID)
			s.assertAccountKeeperCalls(tc.accKeeper, expCalls, "GetMarketDetails(%d)", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_UpdateMarketDetails() {
	baseAcc := func(marketID uint32) *authtypes.BaseAccount {
		return &authtypes.BaseAccount{
			Address:       exchange.GetMarketAddress(marketID).String(),
			PubKey:        nil,
			AccountNumber: uint64(marketID),
			Sequence:      uint64(marketID) * 2,
		}
	}
	standardDeets := func(marketID uint32) exchange.MarketDetails {
		return exchange.MarketDetails{
			Name:        fmt.Sprintf("market %d name", marketID),
			Description: fmt.Sprintf("This is a description of market %d. It's not very helpful.", marketID),
			WebsiteUrl:  fmt.Sprintf("https://example.com/market/%d", marketID),
			IconUri:     fmt.Sprintf("https://icon.example.com/market/%d/small", marketID),
		}
	}
	marketAcc := func(marketID uint32, marketDeets exchange.MarketDetails) *exchange.MarketAccount {
		return &exchange.MarketAccount{
			BaseAccount:   baseAcc(marketID),
			MarketId:      marketID,
			MarketDetails: marketDeets,
		}
	}

	tests := []struct {
		name          string
		accKeeper     *MockAccountKeeper
		marketID      uint32
		marketDetails exchange.MarketDetails
		updatedBy     string
		expErr        string
		expGetAccCall bool
		expSetAccCall sdk.AccountI
	}{
		{
			name:          "invalid market details",
			marketID:      1,
			marketDetails: exchange.MarketDetails{Name: strings.Repeat("v", exchange.MaxName+1)},
			updatedBy:     "whatever",
			expErr:        fmt.Sprintf("name length %d exceeds maximum length of %d", exchange.MaxName+1, exchange.MaxName),
		},
		{
			name:          "no market account",
			marketID:      1,
			marketDetails: exchange.MarketDetails{Name: "what"},
			updatedBy:     "whatever",
			expErr:        "market 1 account not found",
			expGetAccCall: true,
		},
		{
			name:          "not a market account",
			accKeeper:     NewMockAccountKeeper().WithGetAccountResult(exchange.GetMarketAddress(3), baseAcc(3)),
			marketID:      3,
			marketDetails: exchange.MarketDetails{Name: "ignored"},
			updatedBy:     "whatever",
			expErr:        "market 3 account not found",
			expGetAccCall: true,
		},
		{
			name:          "no changes",
			accKeeper:     NewMockAccountKeeper().WithGetAccountResult(exchange.GetMarketAddress(3), marketAcc(3, standardDeets(3))),
			marketID:      3,
			marketDetails: standardDeets(3),
			updatedBy:     "whatever",
			expErr:        "no changes",
			expGetAccCall: true,
		},
		{
			name:          "deleting all fields",
			accKeeper:     NewMockAccountKeeper().WithGetAccountResult(exchange.GetMarketAddress(3), marketAcc(3, standardDeets(3))),
			marketID:      3,
			marketDetails: exchange.MarketDetails{},
			updatedBy:     "i_did_this",
			expGetAccCall: true,
			expSetAccCall: marketAcc(3, exchange.MarketDetails{}),
		},
		{
			name:          "setting all fields",
			accKeeper:     NewMockAccountKeeper().WithGetAccountResult(exchange.GetMarketAddress(5), marketAcc(5, exchange.MarketDetails{})),
			marketID:      5,
			marketDetails: standardDeets(5),
			updatedBy:     "changeling",
			expGetAccCall: true,
			expSetAccCall: marketAcc(5, standardDeets(5)),
		},
		{
			name:          "changing all fields",
			accKeeper:     NewMockAccountKeeper().WithGetAccountResult(exchange.GetMarketAddress(1), marketAcc(1, standardDeets(1))),
			marketID:      1,
			marketDetails: standardDeets(12345),
			updatedBy:     "evil_laugh",
			expGetAccCall: true,
			expSetAccCall: marketAcc(1, standardDeets(12345)),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var expEvents sdk.Events
			var expCalls AccountCalls
			if tc.expGetAccCall {
				expCalls.GetAccount = append(expCalls.GetAccount, exchange.GetMarketAddress(tc.marketID))
			}
			if tc.expSetAccCall != nil {
				expCalls.SetAccount = append(expCalls.SetAccount, tc.expSetAccCall)
				event := exchange.NewEventMarketDetailsUpdated(tc.marketID, tc.updatedBy)
				expEvents = append(expEvents, s.untypeEvent(event))
			}

			if tc.accKeeper == nil {
				tc.accKeeper = NewMockAccountKeeper()
			}
			kpr := s.k.WithAccountKeeper(tc.accKeeper)

			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = kpr.UpdateMarketDetails(ctx, tc.marketID, tc.marketDetails, tc.updatedBy)
			}
			s.Require().NotPanics(testFunc, "UpdateMarketDetails(%d, ...)", tc.marketDetails)
			s.assertErrorValue(err, tc.expErr, "UpdateMarketDetails(%d, ...) error", tc.marketDetails)
			s.assertAccountKeeperCalls(tc.accKeeper, expCalls, "UpdateMarketDetails(%d, ...)", tc.marketDetails)
			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "events after UpdateMarketDetails(%d, ...)", tc.marketDetails)
		})
	}
}

func (s *TestSuite) TestKeeper_CreateMarket() {
	setAccNum := func(id uint64) AccountModifier {
		return func(acc sdk.AccountI) sdk.AccountI {
			err := acc.SetAccountNumber(id)
			s.Require().NoError(err, "SetAccountNumber(%d)", id)
			return acc
		}
	}

	tests := []struct {
		name           string
		setup          func()
		accKeeper      *MockAccountKeeper
		newAccModifier AccountModifier
		market         exchange.Market
		expMarketID    uint32
		expErr         string
		expHasAccCall  bool
		expLastAutoID  uint32
	}{
		{
			name: "market has errors",
			market: exchange.Market{
				ReqAttrCreateAsk: []string{"not$money"},
				ReqAttrCreateBid: []string{"no spaces"},
				MarketDetails: exchange.MarketDetails{
					Description: strings.Repeat("w", 1+exchange.MaxDescription),
				},
			},
			expErr: s.joinErrs(
				"invalid attribute \"not$money\"",
				"invalid attribute \"no spaces\"",
				"description length 2001 exceeds maximum length of 2000",
			),
		},
		{
			name:          "market address already exists",
			accKeeper:     NewMockAccountKeeper().WithHasAccountResult(exchange.GetMarketAddress(1), true),
			market:        exchange.Market{MarketId: 1},
			expErr:        "market id 1 account " + exchange.GetMarketAddress(1).String() + " already exists",
			expHasAccCall: true,
		},
		{
			name:           "no market id, empty state",
			setup:          nil,
			newAccModifier: setAccNum(88),
			market:         exchange.Market{MarketDetails: exchange.MarketDetails{Name: "Empty Market"}},
			expMarketID:    1,
			expHasAccCall:  true,
			expLastAutoID:  1,
		},
		{
			name: "no market id, last one was 55",
			setup: func() {
				keeper.SetLastAutoMarketID(s.getStore(), 55)
			},
			newAccModifier: setAccNum(123),
			market:         exchange.Market{MarketDetails: exchange.MarketDetails{Name: "NAME", Description: "DESCRIPTION"}},
			expMarketID:    56,
			expHasAccCall:  true,
			expLastAutoID:  56,
		},
		{
			name: "market id 78, last one was 22",
			setup: func() {
				keeper.SetLastAutoMarketID(s.getStore(), 22)
			},
			newAccModifier: setAccNum(5),
			market:         exchange.Market{MarketId: 78},
			expMarketID:    78,
			expHasAccCall:  true,
			expLastAutoID:  22,
		},
		{
			name: "market id 5, last one was 18",
			setup: func() {
				keeper.SetLastAutoMarketID(s.getStore(), 18)
			},
			newAccModifier: setAccNum(99),
			market:         exchange.Market{MarketId: 5},
			expMarketID:    5,
			expHasAccCall:  true,
			expLastAutoID:  18,
		},
		{
			name:           "fully filled market",
			newAccModifier: setAccNum(324),
			market: exchange.Market{
				MarketId: 3,
				MarketDetails: exchange.MarketDetails{
					Name:        "Market Three",
					Description: "The third market.",
					WebsiteUrl:  "https://example.com/market/3/info",
					IconUri:     "https://icon.example.com/market/3/small",
				},
				FeeCreateAskFlat:        []sdk.Coin{sdk.NewInt64Coin("incaberry", 88)},
				FeeCreateBidFlat:        []sdk.Coin{sdk.NewInt64Coin("fig", 77)},
				FeeSellerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("grape", 66)},
				FeeSellerSettlementRatios: []exchange.FeeRatio{
					{Price: sdk.NewInt64Coin("plum", 100), Fee: sdk.NewInt64Coin("jackfruit", 3)},
				},
				FeeBuyerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("honeydew", 55)},
				FeeBuyerSettlementRatios: []exchange.FeeRatio{
					{Price: sdk.NewInt64Coin("peach", 500), Fee: sdk.NewInt64Coin("kiwi", 33)},
				},
				AcceptingOrders:     true,
				AllowUserSettlement: true,
				AccessGrants: []exchange.AccessGrant{
					{
						Address:     sdk.AccAddress("just_some_address___").String(),
						Permissions: exchange.AllPermissions(),
					},
				},
				ReqAttrCreateAsk: []string{"*.ask.whatever"},
				ReqAttrCreateBid: []string{"*.bid.whatever"},

				AcceptingCommitments:     true,
				FeeCreateCommitmentFlat:  []sdk.Coin{sdk.NewInt64Coin("orange", 77)},
				CommitmentSettlementBips: 15,
				IntermediaryDenom:        "cherry",
				ReqAttrCreateCommitment:  []string{"*.com.whatever"},
			},
			expMarketID:   3,
			expHasAccCall: true,
			expLastAutoID: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			if tc.accKeeper == nil {
				tc.accKeeper = NewMockAccountKeeper()
			}
			kpr := s.k.WithAccountKeeper(tc.accKeeper)

			origMarket := s.copyMarket(tc.market)
			var expEvents sdk.Events
			var expCalls AccountCalls
			if tc.expHasAccCall {
				id := tc.expMarketID
				if id == 0 {
					id = tc.market.MarketId
				}
				expCalls.HasAccount = append(expCalls.HasAccount, exchange.GetMarketAddress(id))
			}
			if tc.newAccModifier != nil {
				marketAddr := exchange.GetMarketAddress(tc.expMarketID)
				tc.accKeeper.WithNewAccountModifier(marketAddr, tc.newAccModifier)

				expMarketAcc := tc.newAccModifier(&exchange.MarketAccount{
					BaseAccount:   &authtypes.BaseAccount{Address: marketAddr.String()},
					MarketId:      tc.expMarketID,
					MarketDetails: tc.market.MarketDetails,
				})
				// Even though the account number isn't set when the account is provided to NewAccount,
				// It's all passed by reference. So the arg recorded in the NewAccount call gets updated too.
				expCalls.NewAccount = append(expCalls.NewAccount, expMarketAcc)
				expCalls.SetAccount = append(expCalls.SetAccount, expMarketAcc)

				event := exchange.NewEventMarketCreated(tc.expMarketID)
				expEvents = append(expEvents, s.untypeEvent(event))
			}

			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var marketID uint32
			var err error
			testFunc := func() {
				marketID, err = kpr.CreateMarket(ctx, tc.market)
			}
			s.Require().NotPanics(testFunc, "CreateMarket")
			s.assertErrorValue(err, tc.expErr, "CreateMarket error")
			s.Assert().Equal(int(tc.expMarketID), int(marketID), "CreateMarket market id")
			s.assertAccountKeeperCalls(tc.accKeeper, expCalls, "CreateMarket")
			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "events emitted during CreateMarket")
			s.Assert().Equal(origMarket, tc.market, "market arg after CreateMarket")

			if len(tc.expErr) > 0 || s.T().Failed() {
				return
			}

			expMarket := tc.market
			expMarket.MarketId = marketID
			market := kpr.GetMarket(s.ctx, marketID)
			s.Assert().Equal(&expMarket, market, "market read from state after CreateMarket")

			lastMarketID := keeper.GetLastAutoMarketID(s.getStore())
			s.Assert().Equal(int(tc.expLastAutoID), int(lastMarketID), "last auto-market id after CreateMarket")
		})
	}
}

func (s *TestSuite) TestKeeper_GetMarket() {
	tests := []struct {
		name      string
		accKeeper *MockAccountKeeper
		setup     func() *exchange.Market // Should return the expected market.
		marketID  uint32
	}{
		{
			name:     "unknown market",
			marketID: 5,
		},
		{
			name: "empty market",
			accKeeper: NewMockAccountKeeper().WithGetAccountResult(exchange.GetMarketAddress(55), &exchange.MarketAccount{
				BaseAccount: &authtypes.BaseAccount{
					Address:       exchange.GetMarketAddress(55).String(),
					PubKey:        nil,
					AccountNumber: 71,
					Sequence:      0,
				},
				MarketId:      55,
				MarketDetails: exchange.MarketDetails{},
			}),
			setup: func() *exchange.Market {
				market := exchange.Market{
					MarketId:        55,
					AcceptingOrders: true,
				}
				keeper.StoreMarket(s.getStore(), market)
				return &market
			},
			marketID: 55,
		},
		{
			name: "market without an account",
			setup: func() *exchange.Market {
				market := exchange.Market{
					MarketId:        71,
					AcceptingOrders: true,
					AccessGrants: []exchange.AccessGrant{
						{
							Address:     s.addr4.String(),
							Permissions: exchange.AllPermissions(),
						},
					},
				}
				keeper.StoreMarket(s.getStore(), market)
				return &market
			},
			marketID: 71,
		},
		{
			name: "market with everything",
			accKeeper: NewMockAccountKeeper().WithGetAccountResult(exchange.GetMarketAddress(420), &exchange.MarketAccount{
				BaseAccount: &authtypes.BaseAccount{
					Address:       exchange.GetMarketAddress(420).String(),
					PubKey:        nil,
					AccountNumber: 71,
					Sequence:      0,
				},
				MarketId: 420,
				MarketDetails: exchange.MarketDetails{
					Name:        "Market 420 name",
					Description: "Market 420 description",
					WebsiteUrl:  "Market 420 url",
					IconUri:     "Market 420 icon uri",
				},
			}),
			setup: func() *exchange.Market {
				otherMarket1 := exchange.Market{
					MarketId:            419,
					AllowUserSettlement: true,
				}
				otherMarket2 := exchange.Market{
					MarketId:         421,
					FeeCreateAskFlat: []sdk.Coin{sdk.NewInt64Coin("whatever", 421)},
				}
				expMarket := exchange.Market{
					MarketId: 420,
					MarketDetails: exchange.MarketDetails{
						Name:        "Market 420 name",
						Description: "Market 420 description",
						WebsiteUrl:  "Market 420 url",
						IconUri:     "Market 420 icon uri",
					},
					FeeCreateAskFlat:        []sdk.Coin{sdk.NewInt64Coin("acorn", 6), sdk.NewInt64Coin("apple", 5)},
					FeeCreateBidFlat:        []sdk.Coin{sdk.NewInt64Coin("banana", 3), sdk.NewInt64Coin("blueberry", 3)},
					FeeSellerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("farkleberry", 30), sdk.NewInt64Coin("fig", 20)},
					FeeSellerSettlementRatios: []exchange.FeeRatio{
						{Price: sdk.NewInt64Coin("pear", 350), Fee: sdk.NewInt64Coin("grape", 7)},
						{Price: sdk.NewInt64Coin("pear", 500), Fee: sdk.NewInt64Coin("grapefruit", 1)},
					},
					FeeBuyerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("honeycrisp", 12), sdk.NewInt64Coin("honeydew", 2)},
					FeeBuyerSettlementRatios: []exchange.FeeRatio{
						{Price: sdk.NewInt64Coin("plum", 377), Fee: sdk.NewInt64Coin("guava", 3)},
						{Price: sdk.NewInt64Coin("prune", 888), Fee: sdk.NewInt64Coin("guava", 5)},
					},
					AcceptingOrders:     false,
					AllowUserSettlement: true,
					AccessGrants: []exchange.AccessGrant{
						{
							Address: s.addr1.String(),
							Permissions: []exchange.Permission{
								exchange.Permission_settle, exchange.Permission_set_ids, exchange.Permission_cancel,
							},
						},
						{
							Address: s.addr2.String(),
							Permissions: []exchange.Permission{
								exchange.Permission_update, exchange.Permission_permissions, exchange.Permission_attributes,
							},
						},
						{
							Address:     s.addr3.String(),
							Permissions: exchange.AllPermissions(),
						},
						{
							Address:     s.addr4.String(),
							Permissions: []exchange.Permission{exchange.Permission_withdraw},
						},
					},
					ReqAttrCreateAsk: []string{"create-ask.my.market", "*.kyc.someone"},
					ReqAttrCreateBid: []string{"create-bid.my.market", "*.kyc.someone"},

					AcceptingCommitments:     true,
					FeeCreateCommitmentFlat:  []sdk.Coin{sdk.NewInt64Coin("orange", 77)},
					CommitmentSettlementBips: 15,
					IntermediaryDenom:        "cherry",
					ReqAttrCreateCommitment:  []string{"create-com.my.market", "*.kyc.someone"},
				}

				store := s.getStore()
				keeper.StoreMarket(store, otherMarket1)
				keeper.StoreMarket(store, expMarket)
				keeper.StoreMarket(store, otherMarket2)

				return &expMarket
			},
			marketID: 420,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			var expMarket *exchange.Market
			if tc.setup != nil {
				expMarket = tc.setup()
			}

			var expCalls AccountCalls
			if expMarket != nil {
				expCalls.GetAccount = append(expCalls.GetAccount, exchange.GetMarketAddress(tc.marketID))
			}

			if tc.accKeeper == nil {
				tc.accKeeper = NewMockAccountKeeper()
			}
			kpr := s.k.WithAccountKeeper(tc.accKeeper)

			var actMarket *exchange.Market
			testFunc := func() {
				actMarket = kpr.GetMarket(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "GetMarket(%d)", tc.marketID)
			s.Assert().Equal(expMarket, actMarket, "GetMarket(%d) result", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_IterateMarkets() {
	var markets []*exchange.Market
	stopAfter := func(count int) func(market *exchange.Market) bool {
		return func(market *exchange.Market) bool {
			markets = append(markets, market)
			return len(markets) >= count
		}
	}
	getAll := func(market *exchange.Market) bool {
		markets = append(markets, market)
		return false
	}

	standardDetails := func(marketID uint32) exchange.MarketDetails {
		return exchange.MarketDetails{
			Name:        fmt.Sprintf("Market %d", marketID),
			Description: fmt.Sprintf("Description fo market %d. It's not very informational.", marketID),
			WebsiteUrl:  fmt.Sprintf("http://example.com/market/%d/info", marketID),
			IconUri:     fmt.Sprintf("http://example.com/market/%d/icon/huge", marketID),
		}
	}
	standardMarket := func(marketID uint32) *exchange.Market {
		return &exchange.Market{
			MarketId:                marketID,
			MarketDetails:           standardDetails(marketID),
			FeeCreateAskFlat:        []sdk.Coin{sdk.NewInt64Coin("askflat", int64(marketID))},
			FeeCreateBidFlat:        []sdk.Coin{sdk.NewInt64Coin("bidflat", int64(marketID))},
			FeeSellerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("sellerflat", int64(marketID))},
			FeeSellerSettlementRatios: []exchange.FeeRatio{
				{Price: sdk.NewInt64Coin("sellerprice", 500+int64(marketID)), Fee: sdk.NewInt64Coin("sellerfee", int64(marketID))},
			},
			FeeBuyerSettlementFlat: []sdk.Coin{sdk.NewInt64Coin("buyerflat", int64(marketID))},
			FeeBuyerSettlementRatios: []exchange.FeeRatio{
				{Price: sdk.NewInt64Coin("buyerprice", 1500+int64(marketID)), Fee: sdk.NewInt64Coin("buyerfee", 100+int64(marketID))},
			},
			AcceptingOrders:     true,
			AllowUserSettlement: false,
			AccessGrants:        []exchange.AccessGrant{{Address: s.addr5.String(), Permissions: exchange.AllPermissions()}},
			ReqAttrCreateAsk:    []string{fmt.Sprintf("%d.ask.create", marketID)},
			ReqAttrCreateBid:    []string{fmt.Sprintf("%d.bid.create", marketID)},
		}
	}
	mustCreateMarket := func(kpr keeper.Keeper, market exchange.Market) {
		_, err := kpr.CreateMarket(s.ctx, market)
		s.Require().NoError(err, "CreateMarket(%d)", market.MarketId)
	}

	tests := []struct {
		name       string
		setup      func() keeper.Keeper
		cb         func(market *exchange.Market) bool
		expMarkets []*exchange.Market
	}{
		{
			name:       "empty state",
			cb:         getAll,
			expMarkets: nil,
		},
		{
			name: "just market 1",
			setup: func() keeper.Keeper {
				kpr := s.k.WithAccountKeeper(NewMockAccountKeeper())
				mustCreateMarket(kpr, *standardMarket(1))
				return kpr
			},
			cb:         getAll,
			expMarkets: []*exchange.Market{standardMarket(1)},
		},
		{
			name: "just market 20",
			setup: func() keeper.Keeper {
				kpr := s.k.WithAccountKeeper(NewMockAccountKeeper())
				mustCreateMarket(kpr, *standardMarket(20))
				return kpr
			},
			cb:         getAll,
			expMarkets: []*exchange.Market{standardMarket(20)},
		},
		{
			name: "markets 1 through 5: get all",
			setup: func() keeper.Keeper {
				kpr := s.k.WithAccountKeeper(NewMockAccountKeeper())
				mustCreateMarket(kpr, *standardMarket(1))
				mustCreateMarket(kpr, *standardMarket(4))
				mustCreateMarket(kpr, *standardMarket(2))
				mustCreateMarket(kpr, *standardMarket(5))
				mustCreateMarket(kpr, *standardMarket(3))
				return kpr
			},
			cb: getAll,
			expMarkets: []*exchange.Market{
				standardMarket(1),
				standardMarket(2),
				standardMarket(3),
				standardMarket(4),
				standardMarket(5),
			},
		},
		{
			name: "markets 1 through 5: get first",
			setup: func() keeper.Keeper {
				kpr := s.k.WithAccountKeeper(NewMockAccountKeeper())
				mustCreateMarket(kpr, *standardMarket(1))
				mustCreateMarket(kpr, *standardMarket(4))
				mustCreateMarket(kpr, *standardMarket(2))
				mustCreateMarket(kpr, *standardMarket(5))
				mustCreateMarket(kpr, *standardMarket(3))
				return kpr
			},
			cb:         stopAfter(1),
			expMarkets: []*exchange.Market{standardMarket(1)},
		},
		{
			name: "markets 1 through 5: get three",
			setup: func() keeper.Keeper {
				kpr := s.k.WithAccountKeeper(NewMockAccountKeeper())
				mustCreateMarket(kpr, *standardMarket(1))
				mustCreateMarket(kpr, *standardMarket(4))
				mustCreateMarket(kpr, *standardMarket(2))
				mustCreateMarket(kpr, *standardMarket(5))
				mustCreateMarket(kpr, *standardMarket(3))
				return kpr
			},
			cb: stopAfter(3),
			expMarkets: []*exchange.Market{
				standardMarket(1),
				standardMarket(2),
				standardMarket(3),
			},
		},
		{
			name: "five randomly numbered markets: get all",
			setup: func() keeper.Keeper {
				kpr := s.k.WithAccountKeeper(NewMockAccountKeeper())
				mustCreateMarket(kpr, *standardMarket(63))
				mustCreateMarket(kpr, *standardMarket(23))
				mustCreateMarket(kpr, *standardMarket(36))
				mustCreateMarket(kpr, *standardMarket(6))
				mustCreateMarket(kpr, *standardMarket(14))
				return kpr
			},
			cb: getAll,
			expMarkets: []*exchange.Market{
				standardMarket(6),
				standardMarket(14),
				standardMarket(23),
				standardMarket(36),
				standardMarket(63),
			},
		},
		{
			name: "five randomly numbered markets: get first",
			setup: func() keeper.Keeper {
				kpr := s.k.WithAccountKeeper(NewMockAccountKeeper())
				mustCreateMarket(kpr, *standardMarket(63))
				mustCreateMarket(kpr, *standardMarket(23))
				mustCreateMarket(kpr, *standardMarket(36))
				mustCreateMarket(kpr, *standardMarket(6))
				mustCreateMarket(kpr, *standardMarket(14))
				return kpr
			},
			cb:         stopAfter(1),
			expMarkets: []*exchange.Market{standardMarket(6)},
		},
		{
			name: "five randomly numbered markets: get three",
			setup: func() keeper.Keeper {
				kpr := s.k.WithAccountKeeper(NewMockAccountKeeper())
				mustCreateMarket(kpr, *standardMarket(63))
				mustCreateMarket(kpr, *standardMarket(23))
				mustCreateMarket(kpr, *standardMarket(36))
				mustCreateMarket(kpr, *standardMarket(6))
				mustCreateMarket(kpr, *standardMarket(14))
				return kpr
			},
			cb: stopAfter(3),
			expMarkets: []*exchange.Market{
				standardMarket(6),
				standardMarket(14),
				standardMarket(23),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			kpr := s.k
			if tc.setup != nil {
				kpr = tc.setup()
			}

			markets = nil
			testFunc := func() {
				kpr.IterateMarkets(s.ctx, tc.cb)
			}
			s.Require().NotPanics(testFunc, "IterateMarkets")
			s.Assert().Equal(tc.expMarkets, markets, "markets iterated")
		})
	}
}

func (s *TestSuite) TestKeeper_GetMarketBrief() {
	tests := []struct {
		name      string
		accKeeper *MockAccountKeeper
		marketID  uint32
		expected  *exchange.MarketBrief
	}{
		{
			name:     "no account",
			marketID: 1,
			expected: nil,
		},
		{
			name: "empty details",
			accKeeper: NewMockAccountKeeper().WithGetAccountResult(s.marketAddr2, &exchange.MarketAccount{
				BaseAccount: &authtypes.BaseAccount{
					Address:       s.marketAddr2.String(),
					AccountNumber: 777,
				},
				MarketId:      2,
				MarketDetails: exchange.MarketDetails{},
			}),
			marketID: 2,
			expected: &exchange.MarketBrief{
				MarketId:      2,
				MarketAddress: s.marketAddr2.String(),
				MarketDetails: exchange.MarketDetails{},
			},
		},
		{
			name: "full details",
			accKeeper: NewMockAccountKeeper().WithGetAccountResult(s.marketAddr3, &exchange.MarketAccount{
				BaseAccount: &authtypes.BaseAccount{
					Address:       s.marketAddr3.String(),
					AccountNumber: 777,
				},
				MarketId: 3,
				MarketDetails: exchange.MarketDetails{
					Name:        "Market Three",
					Description: "Market Three's description is a bit lacking here.",
					WebsiteUrl:  "website three",
					IconUri:     "icon three",
				},
			}),
			marketID: 3,
			expected: &exchange.MarketBrief{
				MarketId:      3,
				MarketAddress: s.marketAddr3.String(),
				MarketDetails: exchange.MarketDetails{
					Name:        "Market Three",
					Description: "Market Three's description is a bit lacking here.",
					WebsiteUrl:  "website three",
					IconUri:     "icon three",
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.accKeeper == nil {
				tc.accKeeper = NewMockAccountKeeper()
			}
			kpr := s.k.WithAccountKeeper(tc.accKeeper)

			var actual *exchange.MarketBrief
			testFunc := func() {
				actual = kpr.GetMarketBrief(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "GetMarketBrief(%d)", tc.marketID)
			s.Assert().Equal(tc.expected, actual, "GetMarketBrief(%d) result", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_WithdrawMarketFunds() {
	tests := []struct {
		name         string
		bankKeeper   *MockBankKeeper
		marketID     uint32
		toAddr       sdk.AccAddress
		amount       sdk.Coins
		withdrawnBy  string
		expErr       string
		expBlockCall bool
		expSendCall  bool
		expQBypass   bool
	}{
		{
			name:        "invalid admin",
			marketID:    1,
			toAddr:      s.addr1,
			amount:      sdk.NewCoins(sdk.NewInt64Coin("banana", 12)),
			withdrawnBy: "ohno",
			expErr:      "invalid admin \"ohno\": decoding bech32 failed: invalid bech32 string length 4",
		},
		{
			name:         "to blocked address",
			bankKeeper:   NewMockBankKeeper().WithBlockedAddrResults(true),
			marketID:     1,
			toAddr:       s.addr3,
			amount:       sdk.NewCoins(sdk.NewInt64Coin("yay", 4444)),
			withdrawnBy:  s.adminAddr.String(),
			expErr:       s.addr3.String() + " is not allowed to receive funds",
			expBlockCall: true,
		},
		{
			name:         "market 1: error from SendCoins",
			bankKeeper:   NewMockBankKeeper().WithSendCoinsResults("woopsie-daisy: an error story"),
			marketID:     1,
			toAddr:       s.addr1,
			amount:       sdk.NewCoins(sdk.NewInt64Coin("oops", 55)),
			withdrawnBy:  s.adminAddr.String(),
			expErr:       "failed to withdraw 55oops from market 1: woopsie-daisy: an error story",
			expBlockCall: true,
			expSendCall:  true,
			expQBypass:   false,
		},
		{
			name:         "market 8: error from SendCoins",
			bankKeeper:   NewMockBankKeeper().WithSendCoinsResults("ouch-ouch-ouch: a sequel error story"),
			marketID:     8,
			toAddr:       s.addr1,
			amount:       sdk.NewCoins(sdk.NewInt64Coin("awwww", 77), sdk.NewInt64Coin("hurts", 3)),
			withdrawnBy:  s.adminAddr.String(),
			expErr:       "failed to withdraw 77awwww,3hurts from market 8: ouch-ouch-ouch: a sequel error story",
			expBlockCall: true,
			expSendCall:  true,
			expQBypass:   false,
		},
		{
			name:         "market 1: okay to other",
			marketID:     1,
			toAddr:       s.addr3,
			amount:       sdk.NewCoins(sdk.NewInt64Coin("yay", 4444)),
			withdrawnBy:  s.adminAddr.String(),
			expBlockCall: true,
			expSendCall:  true,
			expQBypass:   false,
		},
		{
			name:         "market 8: okay to self",
			marketID:     8,
			toAddr:       s.addr5,
			amount:       sdk.NewCoins(sdk.NewInt64Coin("kaching", 500_000_001)),
			withdrawnBy:  s.addr5.String(),
			expBlockCall: true,
			expSendCall:  true,
			expQBypass:   true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			expCalls := BankCalls{}
			if tc.expBlockCall {
				expCalls.BlockedAddr = append(expCalls.BlockedAddr, tc.toAddr)
			}
			if tc.expSendCall {
				admin, err := sdk.AccAddressFromBech32(tc.withdrawnBy)
				s.Require().NoError(err, "AccAddressFromBech32(%q)", tc.withdrawnBy)
				expCalls.SendCoins = []*SendCoinsArgs{{
					ctxHasQuarantineBypass: tc.expQBypass,
					ctxTransferAgent:       admin,
					fromAddr:               exchange.GetMarketAddress(tc.marketID),
					toAddr:                 tc.toAddr,
					amt:                    tc.amount,
				}}
			}

			var expEvents sdk.Events
			if len(tc.expErr) == 0 {
				event := exchange.NewEventMarketWithdraw(tc.marketID, tc.amount, tc.toAddr, tc.withdrawnBy)
				expEvents = append(expEvents, s.untypeEvent(event))
			}

			if tc.bankKeeper == nil {
				tc.bankKeeper = NewMockBankKeeper()
			}
			kpr := s.k.WithBankKeeper(tc.bankKeeper)

			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = kpr.WithdrawMarketFunds(ctx, tc.marketID, tc.toAddr, tc.amount, tc.withdrawnBy)
			}
			s.Require().NotPanics(testFunc, "WithdrawMarketFunds(%d, %s, %q, %q)",
				tc.marketID, s.getAddrName(tc.toAddr), tc.amount, tc.withdrawnBy)
			s.assertErrorValue(err, tc.expErr, "WithdrawMarketFunds(%d, %s, %q, %q) error",
				tc.marketID, s.getAddrName(tc.toAddr), tc.amount, tc.withdrawnBy)
			s.assertBankKeeperCalls(tc.bankKeeper, expCalls, "WithdrawMarketFunds(%d, %s, %q, %q)",
				tc.marketID, s.getAddrName(tc.toAddr), tc.amount, tc.withdrawnBy)
			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "WithdrawMarketFunds(%d, %s, %q, %q) events",
				tc.marketID, s.getAddrName(tc.toAddr), tc.amount, tc.withdrawnBy)
		})
	}
}

func (s *TestSuite) TestKeeper_ValidateMarket() {
	noBuyerErr := func(denom string) string {
		return "seller settlement fee ratios have price denom \"" + denom + "\" but there are no " +
			"buyer settlement fee ratios with that price denom"
	}
	noSellerErr := func(denom string) string {
		return "buyer settlement fee ratios have price denom \"" + denom + "\" but there is not a " +
			"seller settlement fee ratio with that price denom"
	}

	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		expErr   string
	}{
		{
			name:     "market doesn't exist",
			marketID: 1,
			expErr:   "market 1 does not exist",
		},
		{
			name: "seller price denom not in buyer",
			setup: func() {
				keeper.StoreMarket(s.getStore(), exchange.Market{
					MarketId: 1,
					FeeSellerSettlementRatios: []exchange.FeeRatio{
						{Price: s.coin("500pear"), Fee: s.coin("3pear")},
						{Price: s.coin("500prune"), Fee: s.coin("2prune")},
					},
					FeeBuyerSettlementRatios: []exchange.FeeRatio{{Price: s.coin("500prune"), Fee: s.coin("2fig")}},
				})
			},
			marketID: 1,
			expErr:   noBuyerErr("pear"),
		},
		{
			name: "buyer price denom not in seller",
			setup: func() {
				keeper.StoreMarket(s.getStore(), exchange.Market{
					MarketId:                  1,
					FeeSellerSettlementRatios: []exchange.FeeRatio{{Price: s.coin("500pear"), Fee: s.coin("3pear")}},
					FeeBuyerSettlementRatios: []exchange.FeeRatio{
						{Price: s.coin("500pear"), Fee: s.coin("1grape")},
						{Price: s.coin("500prune"), Fee: s.coin("2fig")},
					},
				})
			},
			marketID: 1,
			expErr:   noSellerErr("prune"),
		},
		{
			name: "multiple errors",
			setup: func() {
				keeper.StoreMarket(s.getStore(), exchange.Market{
					MarketId: 1,
					FeeSellerSettlementRatios: []exchange.FeeRatio{
						{Price: s.coin("600papaya"), Fee: s.coin("1papaya")},
						{Price: s.coin("800peach"), Fee: s.coin("7peach")},
						{Price: s.coin("500pear"), Fee: s.coin("3pear")},
					},
					FeeBuyerSettlementRatios: []exchange.FeeRatio{
						{Price: s.coin("800papaya"), Fee: s.coin("3honeydew")},
						{Price: s.coin("500plum"), Fee: s.coin("3fig")},
						{Price: s.coin("600prune"), Fee: s.coin("9grape")},
					},
				})
			},
			marketID: 1,
			expErr: s.joinErrs(
				noBuyerErr("peach"), noBuyerErr("pear"),
				noSellerErr("plum"), noSellerErr("prune"),
			),
		},
		{
			name: "no ratios",
			setup: func() {
				keeper.StoreMarket(s.getStore(), exchange.Market{MarketId: 2})
			},
			marketID: 2,
			expErr:   "",
		},
		{
			name: "no buyer ratios",
			setup: func() {
				keeper.StoreMarket(s.getStore(), exchange.Market{
					MarketId:                  2,
					FeeSellerSettlementRatios: []exchange.FeeRatio{{Price: s.coin("500pear"), Fee: s.coin("3pear")}},
				})
			},
			marketID: 2,
			expErr:   "",
		},
		{
			name: "no seller ratios",
			setup: func() {
				keeper.StoreMarket(s.getStore(), exchange.Market{
					MarketId:                 2,
					FeeBuyerSettlementRatios: []exchange.FeeRatio{{Price: s.coin("500pear"), Fee: s.coin("3pear")}},
				})
			},
			marketID: 2,
			expErr:   "",
		},
		{
			name: "one ratio each, same price denoms",
			setup: func() {
				keeper.StoreMarket(s.getStore(), exchange.Market{
					MarketId:                  2,
					FeeSellerSettlementRatios: []exchange.FeeRatio{{Price: s.coin("500pear"), Fee: s.coin("3pear")}},
					FeeBuyerSettlementRatios:  []exchange.FeeRatio{{Price: s.coin("500pear"), Fee: s.coin("2fig")}},
				})
			},
			marketID: 2,
			expErr:   "",
		},
		{
			name: "two seller denoms, four buyer ratios with those denoms",
			setup: func() {
				keeper.StoreMarket(s.getStore(), exchange.Market{
					MarketId: 55,
					FeeSellerSettlementRatios: []exchange.FeeRatio{
						{Price: s.coin("300plum"), Fee: s.coin("1plum")},
						{Price: s.coin("800peach"), Fee: s.coin("77peach")},
					},
					FeeBuyerSettlementRatios: []exchange.FeeRatio{
						{Price: s.coin("500plum"), Fee: s.coin("3plum")},
						{Price: s.coin("600plum"), Fee: s.coin("2fig")},
						{Price: s.coin("800peach"), Fee: s.coin("78peach")},
						{Price: s.coin("900peach"), Fee: s.coin("6fig")},
					},
				})
			},
			marketID: 55,
			expErr:   "",
		},
		{
			name: "commitment bips without denom",
			setup: func() {
				keeper.StoreMarket(s.getStore(), exchange.Market{
					MarketId:                 24,
					CommitmentSettlementBips: 80,
				})
			},
			marketID: 24,
			expErr:   "no intermediary denom defined, but commitment settlement bips 80 is not zero",
		},
		{
			name: "intermediary denom without nav",
			setup: func() {
				keeper.StoreMarket(s.getStore(), exchange.Market{
					MarketId:                 32,
					CommitmentSettlementBips: 80,
					IntermediaryDenom:        "cherry",
				})
			},
			marketID: 32,
			expErr:   "no nav exists from intermediary denom \"cherry\" to fee denom \"nhash\"",
		},
		{
			name: "commitments allowed but no fees defined",
			setup: func() {
				keeper.StoreMarket(s.getStore(), exchange.Market{
					MarketId:             28,
					AcceptingCommitments: true,
				})
			},
			marketID: 28,
			expErr:   "commitments are allowed but no commitment-related fees are defined",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var err error
			testFunc := func() {
				err = s.k.ValidateMarket(s.ctx, tc.marketID)
			}
			s.Require().NotPanics(testFunc, "ValidateMarket(%d)", tc.marketID)
			s.assertErrorValue(err, tc.expErr, "ValidateMarket(%d) error", tc.marketID)
		})
	}
}

func (s *TestSuite) TestKeeper_CloseMarket() {
	askOrder := func(orderID uint64, marketID uint32, addr sdk.AccAddress, assets, price string) *exchange.Order {
		return exchange.NewOrder(orderID).WithAsk(&exchange.AskOrder{
			MarketId: marketID,
			Seller:   addr.String(),
			Assets:   s.coin(assets),
			Price:    s.coin(price),
		})
	}
	bidOrder := func(orderID uint64, marketID uint32, addr sdk.AccAddress, assets, price string) *exchange.Order {
		return exchange.NewOrder(orderID).WithBid(&exchange.BidOrder{
			MarketId: marketID,
			Buyer:    addr.String(),
			Assets:   s.coin(assets),
			Price:    s.coin(price),
		})
	}
	marketID := uint32(14)
	signer := s.k.GetAuthority()

	expMarket := exchange.Market{
		MarketId:                 marketID,
		AcceptingOrders:          false,
		AllowUserSettlement:      true,
		AcceptingCommitments:     false,
		CommitmentSettlementBips: 10,
		IntermediaryDenom:        "cherry",
	}

	marketOrders := []*exchange.Order{
		askOrder(1, marketID, s.addr1, "10apple", "20peach"),
		askOrder(2, marketID, s.addr1, "15apple", "25peach"),
		askOrder(3, marketID, s.addr1, "20apple", "26peach"),
		bidOrder(10, marketID, s.addr2, "45apple", "70peach"),
		bidOrder(22, marketID, s.addr3, "71acorn", "5plum"),
		askOrder(24, marketID, s.addr3, "67acorn", "5plum"),
	}
	nonMarketOrders := []*exchange.Order{
		askOrder(4, marketID-1, s.addr1, "10apple", "20peach"),
		bidOrder(23, marketID+1, s.addr3, "67acorn", "5plum"),
	}
	allOrders := make([]*exchange.Order, 0, len(marketOrders)+len(nonMarketOrders))
	allOrders = append(allOrders, marketOrders...)
	allOrders = append(allOrders, nonMarketOrders...)

	marketCommitments := []exchange.Commitment{
		{MarketId: marketID, Account: s.addr1.String(), Amount: s.coins("57cherry,12orange")},
		{MarketId: marketID, Account: s.addr3.String(), Amount: s.coins("88apple,52banana")},
		{MarketId: marketID, Account: s.addr5.String(), Amount: s.coins("14acorn,8peach,15plum")},
	}
	nonMarketCommitments := []exchange.Commitment{
		{MarketId: marketID - 1, Account: s.addr3.String(), Amount: s.coins("12apple,48banana")},
		{MarketId: marketID + 1, Account: s.addr3.String(), Amount: s.coins("832cherry")},
	}
	allCommitments := make([]exchange.Commitment, 0, len(marketCommitments)+len(nonMarketCommitments))
	allCommitments = append(allCommitments, marketCommitments...)
	allCommitments = append(allCommitments, nonMarketCommitments...)

	expHoldCalls := HoldCalls{
		ReleaseHold: []*ReleaseHoldArgs{
			NewReleaseHoldArgs(s.addr1, s.coins("10apple")),
			NewReleaseHoldArgs(s.addr1, s.coins("15apple")),
			NewReleaseHoldArgs(s.addr1, s.coins("20apple")),
			NewReleaseHoldArgs(s.addr2, s.coins("70peach")),
			NewReleaseHoldArgs(s.addr3, s.coins("5plum")),
			NewReleaseHoldArgs(s.addr3, s.coins("67acorn")),
			NewReleaseHoldArgs(s.addr1, s.coins("57cherry,12orange")),
			NewReleaseHoldArgs(s.addr3, s.coins("88apple,52banana")),
			NewReleaseHoldArgs(s.addr5, s.coins("14acorn,8peach,15plum")),
		},
	}
	expEvents := make(sdk.Events, 2, 2+len(marketOrders)+len(marketCommitments))
	expEvents[0] = s.untypeEvent(exchange.NewEventMarketOrdersDisabled(marketID, signer))
	expEvents[1] = s.untypeEvent(exchange.NewEventMarketCommitmentsDisabled(marketID, signer))
	for _, order := range marketOrders {
		expEvents = append(expEvents, s.untypeEvent(exchange.NewEventOrderCancelled(order, signer)))
	}
	for _, com := range marketCommitments {
		expEvents = append(expEvents, s.untypeEvent(exchange.NewEventCommitmentReleased(com.Account, marketID, com.Amount, "GovCloseMarket")))
	}

	initMarket := s.copyMarket(expMarket)
	initMarket.AcceptingOrders = true
	initMarket.AcceptingCommitments = true

	s.clearExchangeState()
	s.requireCreateMarket(initMarket)
	store := s.getStore()
	for _, order := range allOrders {
		s.Require().NoError(s.k.SetOrderInStore(store, *order), "SetOrderInStore(%d)", order.OrderId)
	}
	for i, com := range allCommitments {
		addr, err := sdk.AccAddressFromBech32(com.Account)
		s.Require().NoError(err, "[%d]: AccAddressFromBech32(%q) error", i, com.Account)
		keeper.SetCommitmentAmount(store, com.MarketId, addr, com.Amount)
	}
	store = nil

	// Setup done, make the CloseMarket call.
	em := sdk.NewEventManager()
	ctx := s.ctx.WithEventManager(em)
	holdKeeper := NewMockHoldKeeper()
	kpr := s.k.WithHoldKeeper(holdKeeper)
	testFunc := func() {
		kpr.CloseMarket(ctx, marketID, signer)
	}
	s.Require().NotPanics(testFunc, "CloseMarket(%d, %q)", marketID, signer)

	// And check everything.
	s.assertHoldKeeperCalls(holdKeeper, expHoldCalls, "CloseMarket(%d, %q)", marketID, signer)
	actEvents := em.Events()
	s.assertEqualEvents(expEvents, actEvents, "events emitted during CloseMarket(%d, %q)", marketID, signer)

	actMarket := s.k.GetMarket(s.ctx, marketID)
	s.Assert().Equal(&expMarket, actMarket, "market after CloseMarket(%d, %q)", marketID, signer)

	var actOrders []*exchange.Order
	err := s.k.IterateOrders(s.ctx, func(order *exchange.Order) bool {
		actOrders = append(actOrders, order)
		return false
	})
	if s.Assert().NoError(err, "IterateOrders") {
		s.assertEqualOrders(nonMarketOrders, actOrders, "orders left after CloseMarket(%d, %q)", marketID, signer)
	}

	var actCommitments []exchange.Commitment
	s.k.IterateCommitments(s.ctx, func(commitment exchange.Commitment) bool {
		actCommitments = append(actCommitments, commitment)
		return false
	})
	s.assertEqualCommitments(nonMarketCommitments, actCommitments, "commitments left after CloseMarket(%d, %q)", marketID, signer)
}
