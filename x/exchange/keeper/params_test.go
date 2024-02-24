package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/keeper"
)

func (s *TestSuite) TestKeeper_SetParams() {
	expEntry := func(denom string, value uint16) string {
		keyBz := append([]byte{0}, []byte(keeper.ParamsKeyTypeSplit+denom)...)
		valueBz := keeper.Uint16Bz(value)
		return s.stateEntryString(keyBz, valueBz)
	}
	expAcceptEntry := func(value string) string {
		keyBz := keeper.MakeKeyParamsFeeAcceptPaymentFlat()
		return s.stateEntryString(keyBz, []byte(value))
	}
	expCreateEntry := func(value string) string {
		keyBz := keeper.MakeKeyParamsFeeCreatePaymentFlat()
		return s.stateEntryString(keyBz, []byte(value))
	}

	tests := []struct {
		name     string
		params   *exchange.Params
		expState []string
	}{
		{
			name:     "nil params",
			params:   nil,
			expState: nil,
		},
		{
			name:   "default params",
			params: exchange.DefaultParams(),
			expState: []string{
				expAcceptEntry("100000000nhash"),
				expCreateEntry("100000000nhash"),
				expEntry("", uint16(exchange.DefaultDefaultSplit)),
			},
		},
		{
			name: "zero default and two specifics",
			params: &exchange.Params{
				DefaultSplit: 0,
				DenomSplits: []exchange.DenomSplit{
					{Denom: "cows", Split: 2000},
					{Denom: "chickens", Split: 255},
				},
			},
			expState: []string{
				expEntry("", 0),
				expEntry("chickens", 255),
				expEntry("cows", 2000),
			},
		},
		{
			name: "a default and four specifics",
			params: &exchange.Params{
				DefaultSplit: 300,
				DenomSplits: []exchange.DenomSplit{
					{Denom: "horses", Split: 500},
					{Denom: "llamas", Split: 800},
					{Denom: "pigs", Split: 1200},
					{Denom: "emus", Split: 9999},
				},
			},
			expState: []string{
				expEntry("", 300),
				expEntry("emus", 9999),
				expEntry("horses", 500),
				expEntry("llamas", 800),
				expEntry("pigs", 1200),
			},
		},
		{
			name: "just create payment",
			params: &exchange.Params{
				FeeCreatePaymentFlat: []sdk.Coin{sdk.NewInt64Coin("cherry", 5), sdk.NewInt64Coin("cranberry", 8)},
			},
			expState: []string{
				expCreateEntry("5cherry,8cranberry"),
				expEntry("", 0),
			},
		},
		{
			name: "just accept payment",
			params: &exchange.Params{
				FeeAcceptPaymentFlat: []sdk.Coin{sdk.NewInt64Coin("apple", 2), sdk.NewInt64Coin("apricot", 7)},
			},
			expState: []string{
				expAcceptEntry("2apple,7apricot"),
				expEntry("", 0),
			},
		},
		{
			name: "just payment fees",
			params: &exchange.Params{
				FeeCreatePaymentFlat: []sdk.Coin{sdk.NewInt64Coin("cherry", 22)},
				FeeAcceptPaymentFlat: []sdk.Coin{sdk.NewInt64Coin("apple", 31)},
			},
			expState: []string{
				expAcceptEntry("31apple"),
				expCreateEntry("22cherry"),
				expEntry("", 0),
			},
		},
		{
			name: "one split",
			params: &exchange.Params{
				DefaultSplit: 406,
				DenomSplits:  []exchange.DenomSplit{{Denom: "cats", Split: 5}},
			},
			expState: []string{
				expEntry("", 406),
				expEntry("cats", 5),
			},
		},
		{
			// This one also tests that previously set entries are deleted.
			name:     "nil params again",
			params:   nil,
			expState: nil,
		},
	}

	s.clearExchangeState()
	for _, tc := range tests {
		s.Run(tc.name, func() {
			testFunc := func() {
				s.k.SetParams(s.ctx, tc.params)
			}
			s.Require().NotPanics(testFunc, "SetParams")
			state := s.dumpExchangeState()
			s.Assert().Equal(tc.expState, state, "state after SetParams")
		})
	}
}

func (s *TestSuite) TestKeeper_GetParams() {
	coins := func(coinStr string) []sdk.Coin {
		rv, err := sdk.ParseCoinsNormalized(coinStr)
		s.Require().NoError(err, "ParseCoinsNormalized(%q)", coinStr)
		return rv
	}

	tests := []struct {
		name              string
		splits            []exchange.DenomSplit
		createPaymentFlat []sdk.Coin
		acceptPaymentFlat []sdk.Coin
		exp               *exchange.Params
	}{
		{
			name: "empty state",
			exp:  nil,
		},
		{
			name:   "just a default split",
			splits: []exchange.DenomSplit{{Denom: "", Split: 444}},
			exp:    &exchange.Params{DefaultSplit: 444},
		},
		{
			name:   "one specific split",
			splits: []exchange.DenomSplit{{Denom: "strawberry", Split: 444}},
			exp:    &exchange.Params{DenomSplits: []exchange.DenomSplit{{Denom: "strawberry", Split: 444}}},
		},
		{
			name:              "just create payment flat",
			createPaymentFlat: coins("5000cherry"),
			exp:               &exchange.Params{FeeCreatePaymentFlat: coins("5000cherry")},
		},
		{
			name:              "just accept payment flat",
			acceptPaymentFlat: coins("57apple"),
			exp:               &exchange.Params{FeeAcceptPaymentFlat: coins("57apple")},
		},
		{
			name: "a little of everything",
			splits: []exchange.DenomSplit{
				{Denom: "", Split: 432},
				{Denom: "pigs", Split: 550},
				{Denom: "chickens", Split: 2000},
				{Denom: "cows", Split: 98},
			},
			createPaymentFlat: coins("72cactus"),
			acceptPaymentFlat: coins("21apricot"),
			exp: &exchange.Params{
				DefaultSplit: 432,
				DenomSplits: []exchange.DenomSplit{
					{Denom: "chickens", Split: 2000},
					{Denom: "cows", Split: 98},
					{Denom: "pigs", Split: 550},
				},
				FeeCreatePaymentFlat: coins("72cactus"),
				FeeAcceptPaymentFlat: coins("21apricot"),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			store := s.getStore()
			for _, split := range tc.splits {
				keeper.SetParamsSplit(store, split.Denom, uint16(split.Split))
			}
			keeper.SetParamsFeeCreatePaymentFlat(store, tc.createPaymentFlat)
			keeper.SetParamsFeeAcceptPaymentFlat(store, tc.acceptPaymentFlat)

			var actual *exchange.Params
			testFunc := func() {
				actual = s.k.GetParams(s.ctx)
			}
			s.Require().NotPanics(testFunc, "GetParams()")
			s.Assert().Equal(tc.exp, actual, "GetParams() result")
		})
	}
}

func (s *TestSuite) TestKeeper_GetParamsOrDefaults() {
	coins := func(coinStr string) []sdk.Coin {
		rv, err := sdk.ParseCoinsNormalized(coinStr)
		s.Require().NoError(err, "ParseCoinsNormalized(%q)", coinStr)
		return rv
	}

	tests := []struct {
		name   string
		params *exchange.Params
		exp    *exchange.Params // if nil, params is used.
	}{
		{
			name:   "no params",
			params: nil,
			exp:    exchange.DefaultParams(),
		},
		{
			name:   "zero default no splits",
			params: &exchange.Params{DefaultSplit: 0},
		},
		{
			name: "zero default two splits",
			params: &exchange.Params{
				DefaultSplit: 0,
				DenomSplits: []exchange.DenomSplit{
					{Denom: "llamas", Split: 222},
					{Denom: "cows", Split: 88},
				},
			},
			exp: &exchange.Params{
				DefaultSplit: 0,
				DenomSplits: []exchange.DenomSplit{
					{Denom: "cows", Split: 88},
					{Denom: "llamas", Split: 222},
				},
			},
		},
		{
			name:   "non-zero default and no splits",
			params: &exchange.Params{DefaultSplit: 510},
		},
		{
			name: "just create payment flat",
			params: &exchange.Params{
				FeeCreatePaymentFlat: coins("12cherry"),
			},
		},
		{
			name: "just accept payment flat",
			params: &exchange.Params{
				FeeAcceptPaymentFlat: coins("74apple"),
			},
		},
		{
			name: "both payment flats",
			params: &exchange.Params{
				FeeCreatePaymentFlat: coins("11cherry"),
				FeeAcceptPaymentFlat: coins("73apple"),
			},
		},
		{
			name: "a little bit of everything",
			params: &exchange.Params{
				DefaultSplit: 3333,
				DenomSplits: []exchange.DenomSplit{
					{Denom: "chickens", Split: 72},
					{Denom: "pigs", Split: 111},
				},
				FeeCreatePaymentFlat: coins("91cactus"),
				FeeAcceptPaymentFlat: coins("5acai"),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.exp == nil {
				tc.exp = tc.params
			}
			s.k.SetParams(s.ctx, tc.params)

			var actual *exchange.Params
			testFunc := func() {
				actual = s.k.GetParamsOrDefaults(s.ctx)
			}
			s.Require().NotPanics(testFunc, "GetParamsOrDefaults()")
			s.Assert().Equal(tc.exp, actual, "GetParamsOrDefaults() result")
		})
	}
}

func (s *TestSuite) TestKeeper_GetExchangeSplit() {
	defaultSplit := uint16(exchange.DefaultDefaultSplit)
	tests := []struct {
		name   string
		params *exchange.Params
		denom  string
		exp    uint16
	}{
		{
			name:   "no params, empty string",
			params: nil,
			denom:  "",
			exp:    defaultSplit,
		},
		{
			name:   "no params, chickens",
			params: nil,
			denom:  "chickens",
			exp:    defaultSplit,
		},
		{
			name:   "default params, empty string",
			params: exchange.DefaultParams(),
			denom:  "",
			exp:    defaultSplit,
		},
		{
			name:   "default params, cows",
			params: exchange.DefaultParams(),
			denom:  "cows",
			exp:    defaultSplit,
		},
		{
			name: "split for llamas, emus",
			params: &exchange.Params{
				DefaultSplit: 300,
				DenomSplits:  []exchange.DenomSplit{{Denom: "llamas", Split: 100}},
			},
			denom: "emus",
			exp:   300,
		},
		{
			name: "split for llamas, llama (not plural)",
			params: &exchange.Params{
				DefaultSplit: 300,
				DenomSplits:  []exchange.DenomSplit{{Denom: "llamas", Split: 100}},
			},
			denom: "llama",
			exp:   300,
		},
		{
			name: "split for llamas, llamas",
			params: &exchange.Params{
				DefaultSplit: 300,
				DenomSplits:  []exchange.DenomSplit{{Denom: "llamas", Split: 100}},
			},
			denom: "llamas",
			exp:   100,
		},
		{
			name: "splits for cows, chickens: pigs",
			params: &exchange.Params{
				DefaultSplit: 200,
				DenomSplits: []exchange.DenomSplit{
					{Denom: "chickens", Split: 300},
					{Denom: "cows", Split: 400},
				},
			},
			denom: "pigs",
			exp:   200,
		},
		{
			name: "splits for cows, chickens: cows",
			params: &exchange.Params{
				DefaultSplit: 200,
				DenomSplits: []exchange.DenomSplit{
					{Denom: "chickens", Split: 300},
					{Denom: "cows", Split: 400},
				},
			},
			denom: "cows",
			exp:   400,
		},
		{
			name: "splits for cows, chickens: chickens",
			params: &exchange.Params{
				DefaultSplit: 200,
				DenomSplits: []exchange.DenomSplit{
					{Denom: "chickens", Split: 300},
					{Denom: "cows", Split: 400},
				},
			},
			denom: "chickens",
			exp:   300,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.k.SetParams(s.ctx, tc.params)
			var actual uint16
			testFunc := func() {
				actual = s.k.GetExchangeSplit(s.ctx, tc.denom)
			}
			s.Require().NotPanics(testFunc, "GetExchangeSplit(%q)", tc.denom)
			s.Assert().Equal(tc.exp, actual, "GetExchangeSplit(%q) result", tc.denom)
		})
	}
}
