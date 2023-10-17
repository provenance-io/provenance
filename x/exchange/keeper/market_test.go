package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
				setter(store, 1, []sdk.Coin{s.coin("8acorn")})
				setter(store, 3, []sdk.Coin{s.coin("3apple")})
			},
			marketID: 2,
			expected: nil,
		},
		{
			name: "market with one entry",
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
				tc.setup(s)
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
				setter(store, 1, []sdk.Coin{s.coin("8acorn")})
				setter(store, 3, []sdk.Coin{s.coin("3apple")})
			},
			marketID: 2,
			expected: nil,
		},
		{
			name: "market with one entry",
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
				tc.setup(s)
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
				setter(store, 1, []sdk.Coin{s.coin("8acorn")})
				setter(store, 3, []sdk.Coin{s.coin("3apple")})
			},
			marketID: 2,
			expected: nil,
		},
		{
			name: "market with one entry",
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
				tc.setup(s)
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
		setup    func(s *TestSuite)
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
			setup: func(s *TestSuite) {
				store := s.getStore()
				setter(store, 1, []exchange.FeeRatio{s.ratio("8peach:1fig")})
				setter(store, 3, []exchange.FeeRatio{s.ratio("10plum:1fig")})
			},
			marketID: 2,
			expected: nil,
		},
		{
			name: "market with one entry",
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
				tc.setup(s)
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
				setter(store, 1, []sdk.Coin{s.coin("8acorn")})
				setter(store, 3, []sdk.Coin{s.coin("3apple")})
			},
			marketID: 2,
			expected: nil,
		},
		{
			name: "market with one entry",
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
				tc.setup(s)
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
		setup    func(s *TestSuite)
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
			setup: func(s *TestSuite) {
				store := s.getStore()
				setter(store, 1, []exchange.FeeRatio{s.ratio("8peach:1fig")})
				setter(store, 3, []exchange.FeeRatio{s.ratio("10plum:1fig")})
			},
			marketID: 2,
			expected: nil,
		},
		{
			name: "market with one entry",
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
				tc.setup(s)
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

func (s *TestSuite) TestKeeper_CalculateSellerSettlementRatioFee() {
	setter := keeper.SetSellerSettlementRatios
	tests := []struct {
		name     string
		setup    func(s *TestSuite)
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
				setter(s.getStore(), 1, []exchange.FeeRatio{s.ratio("0peach:1peach")})
			},
			marketID: 1,
			price:    s.coin("100peach"),
			expErr:   "invalid seller settlement fees: cannot apply ratio 0peach:1peach to price 100peach: division by zero",
		},
		{
			name: "three ratios: first",
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
				tc.setup(s)
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
	noDivErr := func(ratio, price string) string {
		return fmt.Sprintf("buyer settlement fees: cannot apply ratio %s to price %s: price amount cannot be evenly divided by ratio price",
			ratio, price)
	}
	noRatiosErr := func(price string) string {
		return "no applicable buyer settlement fee ratios found for price " + price
	}

	tests := []struct {
		name     string
		setup    func(s *TestSuite)
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
				setter(s.getStore(), 15, []exchange.FeeRatio{s.ratio("500pineapple:1fig")})
			},
			marketID: 15,
			price:    s.coin("7500pineapple"),
			expOpts:  []sdk.Coin{s.coin("15fig")},
		},
		{
			name: "one ratio: not evenly divisible",
			setup: func(s *TestSuite) {
				setter(s.getStore(), 15, []exchange.FeeRatio{s.ratio("500pineapple:1fig")})
			},
			marketID: 15,
			price:    s.coin("7503pineapple"),
			expErr: s.joinErrs(
				noDivErr("500pineapple:1fig", "7503pineapple"),
				noRatiosErr("7503pineapple"),
			),
		},
		{
			name: "three ratios for denom: none divisible",
			setup: func(s *TestSuite) {
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
			expErr: s.joinErrs(
				noDivErr("123plum:1fig", "3000plum"),
				noDivErr("234plum:5grape", "3000plum"),
				noDivErr("345plum:7honeydew", "3000plum"),
				noRatiosErr("3000plum"),
			),
		},
		{
			name: "three ratios for denom: only first divisible",
			setup: func(s *TestSuite) {
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
			price:    s.coin("615plum"),
			expOpts:  []sdk.Coin{s.coin("5fig")},
		},
		{
			name: "three ratios for denom: only second divisible",
			setup: func(s *TestSuite) {
				setter(s.getStore(), 99, []exchange.FeeRatio{
					s.ratio("123plum:1fig"),
					s.ratio("234plum:5grape"),
					s.ratio("345plum:7honeydew"),
					s.ratio("100peach:3fig"),
					s.ratio("200peach:11grape"),
					s.ratio("300peach:13honeydew"),
				})
			},
			marketID: 99,
			price:    s.coin("1170plum"),
			expOpts:  []sdk.Coin{s.coin("25grape")},
		},
		{
			name: "three ratios for denom: only third divisible",
			setup: func(s *TestSuite) {
				setter(s.getStore(), 3, []exchange.FeeRatio{
					s.ratio("123plum:1fig"),
					s.ratio("234plum:5grape"),
					s.ratio("345plum:7honeydew"),
					s.ratio("100peach:3fig"),
					s.ratio("200peach:11grape"),
					s.ratio("300peach:13honeydew"),
				})
			},
			marketID: 3,
			price:    s.coin("1725plum"),
			expOpts:  []sdk.Coin{s.coin("35honeydew")},
		},
		{
			name: "three ratios for denom: first not divisible",
			setup: func(s *TestSuite) {
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
			expOpts:  []sdk.Coin{s.coin("575grape"), s.coin("546honeydew")},
		},
		{
			name: "three ratios for denom: second not divisible",
			setup: func(s *TestSuite) {
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
			expOpts:  []sdk.Coin{s.coin("1503fig"), s.coin("2171honeydew")},
		},
		{
			name: "three ratios for denom: third not divisible",
			setup: func(s *TestSuite) {
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
			expOpts:  []sdk.Coin{s.coin("1506fig"), s.coin("2761grape")},
		},
		{
			name: "three ratios for denom: all divisible",
			setup: func(s *TestSuite) {
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
				tc.setup(s)
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
		setup    func(s *TestSuite)
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      nil,
			expErr:   nilFeeErr("10fig,3grape,7honeydew"),
		},
		{
			name: "three fee options: wrong denom",
			setup: func(s *TestSuite) {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("80apple"),
			expErr:   noFeeErr("80apple", "10fig,3grape,7honeydew"),
		},
		{
			name: "three fee options: first, low",
			setup: func(s *TestSuite) {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("9fig"),
			expErr:   lowFeeErr("9fig", "10fig"),
		},
		{
			name: "three fee options: first, ok",
			setup: func(s *TestSuite) {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("10fig"),
			expErr:   "",
		},
		{
			name: "three fee options: second, low",
			setup: func(s *TestSuite) {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("2grape"),
			expErr:   lowFeeErr("2grape", "3grape"),
		},
		{
			name: "three fee options: second, ok",
			setup: func(s *TestSuite) {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("3grape"),
			expErr:   "",
		},
		{
			name: "three fee options: third, low",
			setup: func(s *TestSuite) {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("6honeydew"),
			expErr:   lowFeeErr("6honeydew", "7honeydew"),
		},
		{
			name: "three fee options: third, ok",
			setup: func(s *TestSuite) {
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
				tc.setup(s)
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
		setup    func(s *TestSuite)
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      nil,
			expErr:   nilFeeErr("10fig,3grape,7honeydew"),
		},
		{
			name: "three fee options: wrong denom",
			setup: func(s *TestSuite) {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("80apple"),
			expErr:   noFeeErr("80apple", "10fig,3grape,7honeydew"),
		},
		{
			name: "three fee options: first, low",
			setup: func(s *TestSuite) {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("9fig"),
			expErr:   lowFeeErr("9fig", "10fig"),
		},
		{
			name: "three fee options: first, ok",
			setup: func(s *TestSuite) {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("10fig"),
			expErr:   "",
		},
		{
			name: "three fee options: second, low",
			setup: func(s *TestSuite) {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("2grape"),
			expErr:   lowFeeErr("2grape", "3grape"),
		},
		{
			name: "three fee options: second, ok",
			setup: func(s *TestSuite) {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("3grape"),
			expErr:   "",
		},
		{
			name: "three fee options: third, low",
			setup: func(s *TestSuite) {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("6honeydew"),
			expErr:   lowFeeErr("6honeydew", "7honeydew"),
		},
		{
			name: "three fee options: third, ok",
			setup: func(s *TestSuite) {
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
				tc.setup(s)
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
		setup    func(s *TestSuite)
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      nil,
			expErr:   nilFeeErr("10fig,3grape,7honeydew"),
		},
		{
			name: "three fee options: wrong denom",
			setup: func(s *TestSuite) {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("80apple"),
			expErr:   noFeeErr("80apple", "10fig,3grape,7honeydew"),
		},
		{
			name: "three fee options: first, low",
			setup: func(s *TestSuite) {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("9fig"),
			expErr:   lowFeeErr("9fig", "10fig"),
		},
		{
			name: "three fee options: first, ok",
			setup: func(s *TestSuite) {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("10fig"),
			expErr:   "",
		},
		{
			name: "three fee options: second, low",
			setup: func(s *TestSuite) {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("2grape"),
			expErr:   lowFeeErr("2grape", "3grape"),
		},
		{
			name: "three fee options: second, ok",
			setup: func(s *TestSuite) {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("3grape"),
			expErr:   "",
		},
		{
			name: "three fee options: third, low",
			setup: func(s *TestSuite) {
				setter(s.getStore(), 8, []sdk.Coin{s.coin("10fig"), s.coin("3grape"), s.coin("7honeydew")})
			},
			marketID: 8,
			fee:      s.coinP("6honeydew"),
			expErr:   lowFeeErr("6honeydew", "7honeydew"),
		},
		{
			name: "three fee options: third, ok",
			setup: func(s *TestSuite) {
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
				tc.setup(s)
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
		setup             func(s *TestSuite)
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
				keeper.SetSellerSettlementRatios(s.getStore(), 1, []exchange.FeeRatio{s.ratio("0plum:1plum")})
			},
			marketID:          1,
			price:             s.coin("100plum"),
			settlementFlatFee: nil,
			expErr:            "cannot apply ratio 0plum:1plum to price 100plum: division by zero",
		},
		{
			name: "three ratios: wrong denom",
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
				tc.setup(s)
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
		setup    func(s *TestSuite)
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("11peach,9plum"))
			},
			marketID: 2,
			price:    s.coin("54pear"),
			fee:      s.coins("11peach"),
			expErr:   "",
		},
		{
			name: "only flat: more than req",
			setup: func(s *TestSuite) {
				keeper.SetBuyerSettlementFlatFees(s.getStore(), 2, s.coins("11peach,9plum"))
			},
			marketID: 2,
			price:    s.coin("54pear"),
			fee:      s.coins("10peach,10plum"),
			expErr:   "",
		},
		{
			name: "only ratio: nofee",
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
			setup: func(s *TestSuite) {
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
				tc.setup(s)
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
