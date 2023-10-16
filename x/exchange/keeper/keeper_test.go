package keeper_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/keeper"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type TestSuite struct {
	suite.Suite

	app       *app.App
	ctx       sdk.Context
	stdlibCtx context.Context

	k          keeper.Keeper
	acctKeeper exchange.AccountKeeper
	attrKeeper exchange.AttributeKeeper
	bankKeeper exchange.BankKeeper
	holdKeeper exchange.HoldKeeper

	bondDenom  string
	initBal    sdk.Coins
	initAmount int64

	addr1 sdk.AccAddress
	addr2 sdk.AccAddress
	addr3 sdk.AccAddress
	addr4 sdk.AccAddress
	addr5 sdk.AccAddress

	marketAddr1 sdk.AccAddress
	marketAddr2 sdk.AccAddress
	marketAddr3 sdk.AccAddress

	feeCollector string
}

func (s *TestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.stdlibCtx = sdk.WrapSDKContext(s.ctx)
	s.k = s.app.ExchangeKeeper
	s.acctKeeper = s.app.AccountKeeper
	s.attrKeeper = s.app.AttributeKeeper
	s.bankKeeper = s.app.BankKeeper
	s.holdKeeper = s.app.HoldKeeper

	s.bondDenom = s.app.StakingKeeper.BondDenom(s.ctx)
	s.initAmount = 1_000_000_000
	s.initBal = sdk.NewCoins(sdk.NewCoin(s.bondDenom, sdk.NewInt(s.initAmount)))

	addrs := app.AddTestAddrsIncremental(s.app, s.ctx, 5, sdk.NewInt(s.initAmount))
	s.addr1 = addrs[0]
	s.addr2 = addrs[1]
	s.addr3 = addrs[2]
	s.addr4 = addrs[3]
	s.addr5 = addrs[4]

	s.marketAddr1 = exchange.GetMarketAddress(1)
	s.marketAddr2 = exchange.GetMarketAddress(2)
	s.marketAddr3 = exchange.GetMarketAddress(3)

	s.feeCollector = s.k.GetFeeCollectorName()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

// sliceStrings converts each val into a string using the provided stringer.
func sliceStrings[T any](vals []T, stringer func(T) string) []string {
	if vals == nil {
		return nil
	}
	strs := make([]string, len(vals))
	for i, v := range vals {
		strs[i] = fmt.Sprintf("[%d]:%q", i, stringer(v))
	}
	return strs
}

// sliceString converts each val into a string using the provided stringer and joins them with ", ".
func sliceString[T any](vals []T, stringer func(T) string) string {
	if vals == nil {
		return "<nil>"
	}
	return strings.Join(sliceStrings(vals, stringer), ", ")
}

// coins creates an sdk.Coins from a string, requiring it to work.
func (s *TestSuite) coins(coins string) sdk.Coins {
	s.T().Helper()
	rv, err := sdk.ParseCoinsNormalized(coins)
	s.Require().NoError(err, "ParseCoinsNormalized(%q)", coins)
	return rv
}

// coin creates a new coin without doing any validation on it.
func (s *TestSuite) coin(coin string) sdk.Coin {
	rv, err := sdk.ParseCoinNormalized(coin)
	s.Require().NoError(err, "ParseCoinNormalized(%q)", coin)
	return rv
}

// coinP creates a reference to a new coin without doing any validation on it.
func (s *TestSuite) coinP(coin string) *sdk.Coin {
	rv := s.coin(coin)
	return &rv
}

// coinsString converts a slice of coin entries into a string.
// This is different from sdk.Coins.String because the entries aren't sorted.
func (s *TestSuite) coinsString(coins []sdk.Coin) string {
	return sliceString(coins, func(coin sdk.Coin) string {
		return fmt.Sprintf("%q", coin)
	})
}

// coinPString converts the provided coin to a quoted string, or "<nil>".
func (s *TestSuite) coinPString(coin *sdk.Coin) string {
	if coin == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%q", coin)
}

// ratio creates a FeeRatio from a "<price>:<fee>" string.
func (s *TestSuite) ratio(ratioStr string) exchange.FeeRatio {
	rv, err := exchange.ParseFeeRatio(ratioStr)
	s.Require().NoError(err, "ParseFeeRatio(%q)", ratioStr)
	return *rv
}

// int is a shorter way to call sdkmath.NewInt.
func (s *TestSuite) int(amount int64) sdkmath.Int {
	return sdkmath.NewInt(amount)
}

// intStr creates an sdkmath.Int from a string, requiring it to work.
func (s *TestSuite) intStr(amount string) sdkmath.Int {
	s.T().Helper()
	rv, ok := sdkmath.NewIntFromString(amount)
	s.Require().True(ok, "NewIntFromString(%q) ok bool", amount)
	return rv
}

// ratiosStrings converts the ratios into strings. It's because comparsions on sdk.Coin (or sdkmath.Int) are annoying.
func (s *TestSuite) ratiosStrings(ratios []exchange.FeeRatio) []string {
	return sliceStrings(ratios, exchange.FeeRatio.String)
}

// joinErrs joins the provided error strings into a single one to match what errors.Join does.
func (s *TestSuite) joinErrs(errs ...string) string {
	return strings.Join(errs, "\n")
}

// getAddrName returns the name of the variable in this TestSuite holding the provided address.
func (s *TestSuite) getAddrName(addr sdk.AccAddress) string {
	switch string(addr) {
	case string(s.addr1):
		return "addr1"
	case string(s.addr2):
		return "addr2"
	case string(s.addr3):
		return "addr3"
	case string(s.addr4):
		return "addr4"
	case string(s.addr5):
		return "addr5"
	case string(s.marketAddr1):
		return "marketAddr1"
	case string(s.marketAddr2):
		return "marketAddr2"
	case string(s.marketAddr3):
		return "marketAddr3"
	default:
		return addr.String()
	}
}

// getAddrStrName returns the name of the variable in this TestSuite holding the provided address.
func (s *TestSuite) getAddrStrName(addrStr string) string {
	addr, err := sdk.AccAddressFromBech32(addrStr)
	if err != nil {
		return addrStr
	}
	return s.getAddrName(addr)
}

// getStore gets the exchange store.
func (s *TestSuite) getStore() sdk.KVStore {
	return s.k.GetStore(s.ctx)
}

// clearExchangeState deletes everything from the exchange state store.
func (s *TestSuite) clearExchangeState() {
	keeper.DeleteAll(s.getStore(), nil)
}

// stateEntryString converts the provided key and value into a "<key>"="<value>" string.
func (s *TestSuite) stateEntryString(key, value []byte) string {
	return fmt.Sprintf("%q=%q", key, value)
}

// dumpHoldState creates a string for each entry in the hold state store.
// Each entry has the format `"<key>"="<value>"`.
func (s *TestSuite) dumpHoldState() []string {
	var rv []string
	keeper.Iterate(s.getStore(), nil, func(key, value []byte) bool {
		rv = append(rv, s.stateEntryString(key, value))
		return false
	})
	return rv
}

// requireFundAccount calls testutil.FundAccount, making sure it doesn't panic or return an error.
func (s *TestSuite) requireFundAccount(addr sdk.AccAddress, coins string) {
	assertions.RequireNotPanicsNoErrorf(s.T(), func() error {
		return testutil.FundAccount(s.app.BankKeeper, s.ctx, addr, s.coins(coins))
	}, "FundAccount(%s, %q)", s.getAddrName(addr), coins)
}

// assertErrorValue is a wrapper for assertions.AssertErrorValue for this TestSuite.
func (s *TestSuite) assertErrorValue(theError error, expected string, msgAndArgs ...interface{}) bool {
	return assertions.AssertErrorValue(s.T(), theError, expected, msgAndArgs...)
}

// assertErrorContents is a wrapper for assertions.AssertErrorContents for this TestSuite.
func (s *TestSuite) assertErrorContents(theError error, contains []string, msgAndArgs ...interface{}) bool {
	return assertions.AssertErrorContents(s.T(), theError, contains, msgAndArgs...)
}

func (s *TestSuite) TestKeeper_GetAuthority() {
	expected := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	var actual string
	testFunc := func() {
		actual = s.k.GetAuthority()
	}
	s.Require().NotPanics(testFunc, "GetAuthority()")
	s.Assert().Equal(expected, actual, "GetAuthority() result")
}

func (s *TestSuite) TestKeeper_IsAuthority() {
	tests := []struct {
		name string
		addr string
		exp  bool
	}{
		{name: "empty string", addr: "", exp: false},
		{name: "whitespace", addr: strings.Repeat(" ", len(s.k.GetAuthority())), exp: false},
		{name: "authority", addr: s.k.GetAuthority(), exp: true},
		{name: "authority upper-case", addr: strings.ToUpper(s.k.GetAuthority()), exp: true},
		{name: "authority space", addr: s.k.GetAuthority() + " ", exp: false},
		{name: "space authority", addr: " " + s.k.GetAuthority(), exp: false},
		{name: "other addr", addr: s.addr1.String(), exp: false},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var actual bool
			testFunc := func() {
				actual = s.k.IsAuthority(tc.addr)
			}
			s.Require().NotPanics(testFunc, "IsAuthority(%q)", tc.addr)
			s.Assert().Equal(tc.exp, actual, "IsAuthority(%q) result", tc.addr)
		})
	}
}

func (s *TestSuite) TestKeeper_ValidateAuthority() {
	tests := []struct {
		name   string
		addr   string
		expErr bool
	}{
		{name: "empty string", addr: "", expErr: true},
		{name: "whitespace", addr: strings.Repeat(" ", len(s.k.GetAuthority())), expErr: true},
		{name: "authority", addr: s.k.GetAuthority(), expErr: false},
		{name: "authority upper-case", addr: strings.ToUpper(s.k.GetAuthority()), expErr: false},
		{name: "authority space", addr: s.k.GetAuthority() + " ", expErr: true},
		{name: "space authority", addr: " " + s.k.GetAuthority(), expErr: true},
		{name: "other addr", addr: s.addr1.String(), expErr: true},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			expErr := ""
			if tc.expErr {
				expErr = fmt.Sprintf("expected %q got %q: expected gov account as only signer for proposal message", s.k.GetAuthority(), tc.addr)
			}
			var err error
			testFunc := func() {
				err = s.k.ValidateAuthority(tc.addr)
			}
			s.Require().NotPanics(testFunc, "ValidateAuthority(%q)", tc.addr)
			s.assertErrorValue(err, expErr, "ValidateAuthority(%q) error", tc.addr)
		})
	}
}

func (s *TestSuite) TestKeeper_GetFeeCollectorName() {
	expected := authtypes.FeeCollectorName
	var actual string
	testFunc := func() {
		actual = s.k.GetFeeCollectorName()
	}
	s.Require().NotPanics(testFunc, "GetFeeCollectorName()")
	s.Assert().Equal(expected, actual, "GetFeeCollectorName() result")
}

func (s *TestSuite) TestKeeper_DoTransfer() {
	tests := []struct {
		name     string
		bk       *MockBankKeeper
		inputs   []banktypes.Input
		outputs  []banktypes.Output
		expErr   string
		expSends []*SendCoinsArgs
		expIO    bool
	}{
		{
			name:    "1 in, 1 out: different denoms",
			inputs:  []banktypes.Input{{Address: s.addr1.String(), Coins: s.coins("10apple")}},
			outputs: []banktypes.Output{{Address: s.addr2.String(), Coins: s.coins("10banana")}},
			expErr:  "input coins \"10apple\" does not equal output coins \"10banana\"",
		},
		{
			name:    "1 in, 1 out: different amounts",
			inputs:  []banktypes.Input{{Address: s.addr1.String(), Coins: s.coins("10apple")}},
			outputs: []banktypes.Output{{Address: s.addr2.String(), Coins: s.coins("11apple")}},
			expErr:  "input coins \"10apple\" does not equal output coins \"11apple\"",
		},
		{
			name:    "1 in, 1 out: bad in addr",
			inputs:  []banktypes.Input{{Address: "badInAddr", Coins: s.coins("10apple")}},
			outputs: []banktypes.Output{{Address: s.addr2.String(), Coins: s.coins("10apple")}},
			expErr:  "invalid inputs[0] address \"badInAddr\": decoding bech32 failed: string not all lowercase or all uppercase",
		},
		{
			name:    "1 in, 1 out: bad out addr",
			inputs:  []banktypes.Input{{Address: s.addr1.String(), Coins: s.coins("10apple")}},
			outputs: []banktypes.Output{{Address: "badOutAddr", Coins: s.coins("10apple")}},
			expErr:  "invalid outputs[0] address \"badOutAddr\": decoding bech32 failed: string not all lowercase or all uppercase",
		},
		{
			name:    "1 in, 1 out: err from SendCoins",
			bk:      NewMockBankKeeper().WithSendCoinsResults("test error X from SendCoins"),
			inputs:  []banktypes.Input{{Address: s.addr1.String(), Coins: s.coins("10apple")}},
			outputs: []banktypes.Output{{Address: s.addr2.String(), Coins: s.coins("10apple")}},
			expErr:  "test error X from SendCoins",
			expSends: []*SendCoinsArgs{
				{ctxHasQuarantineBypass: true, fromAddr: s.addr1, toAddr: s.addr2, amt: s.coins("10apple")},
			},
		},
		{
			name:    "1 in, 1 out: okay",
			inputs:  []banktypes.Input{{Address: s.addr2.String(), Coins: s.coins("15banana")}},
			outputs: []banktypes.Output{{Address: s.addr3.String(), Coins: s.coins("15banana")}},
			expSends: []*SendCoinsArgs{
				{ctxHasQuarantineBypass: true, fromAddr: s.addr2, toAddr: s.addr3, amt: s.coins("15banana")},
			},
		},
		{
			name:   "1 in, 3 out: err from InputOutputCoins",
			bk:     NewMockBankKeeper().WithInputOutputCoinsResults("test error V from InputOutputCoins"),
			inputs: []banktypes.Input{{Address: s.addr5.String(), Coins: s.coins("60cactus")}},
			outputs: []banktypes.Output{
				{Address: s.addr4.String(), Coins: s.coins("18cactus")},
				{Address: s.addr3.String(), Coins: s.coins("5cactus")},
				{Address: s.addr2.String(), Coins: s.coins("37cactus")},
			},
			expErr: "test error V from InputOutputCoins",
			expIO:  true,
		},
		{
			name:   "1 in, 3 out: okay",
			inputs: []banktypes.Input{{Address: s.addr5.String(), Coins: s.coins("60cactus")}},
			outputs: []banktypes.Output{
				{Address: s.addr4.String(), Coins: s.coins("18cactus")},
				{Address: s.addr3.String(), Coins: s.coins("5cactus")},
				{Address: s.addr2.String(), Coins: s.coins("37cactus")},
			},
			expIO: true,
		},
		{
			name: "3 in, 1 out: err from InputOutputCoins",
			bk:   NewMockBankKeeper().WithInputOutputCoinsResults("test error P from InputOutputCoins"),
			inputs: []banktypes.Input{
				{Address: s.addr1.String(), Coins: s.coins("51date")},
				{Address: s.addr2.String(), Coins: s.coins("3date")},
				{Address: s.addr3.String(), Coins: s.coins("16date")},
			},
			outputs: []banktypes.Output{
				{Address: s.addr4.String(), Coins: s.coins("70apple")},
			},
			expErr: "test error P from InputOutputCoins",
			expIO:  true,
		},
		{
			name: "3 in, 1 out: okay",
			inputs: []banktypes.Input{
				{Address: s.addr1.String(), Coins: s.coins("51date")},
				{Address: s.addr2.String(), Coins: s.coins("3date")},
				{Address: s.addr3.String(), Coins: s.coins("16date")},
			},
			outputs: []banktypes.Output{{Address: s.addr4.String(), Coins: s.coins("70apple")}},
			expIO:   true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.bk == nil {
				tc.bk = NewMockBankKeeper()
			}
			expCalls := BankCalls{
				SendCoinsCalls:                    tc.expSends,
				SendCoinsFromAccountToModuleCalls: nil,
				InputOutputCoinsCalls:             nil,
			}
			if tc.expIO {
				expCalls.InputOutputCoinsCalls = append(expCalls.InputOutputCoinsCalls, &InputOutputCoinsArgs{
					ctxHasQuarantineBypass: true,
					inputs:                 tc.inputs,
					outputs:                tc.outputs,
				})
			}

			kpr := s.k.WithBankKeeper(tc.bk)
			var err error
			testFunc := func() {
				err = kpr.DoTransfer(s.ctx, tc.inputs, tc.outputs)
			}
			s.Require().NotPanics(testFunc, "DoTransfer")
			s.assertErrorValue(err, tc.expErr, "DoTransfer error")
			s.assertBankKeeperCalls(tc.bk, expCalls, "DoTransfer")
		})
	}
}

func (s *TestSuite) TestKeeper_CalculateExchangeSplit() {
	tests := []struct {
		name   string
		params *exchange.Params
		feeAmt sdk.Coins
		expAmt sdk.Coins
	}{
		{
			name:   "no params in state",
			params: nil,
			feeAmt: s.coins("100apple,20banana"),
			expAmt: s.coins("5apple,1banana"),
		},
		{
			name:   "default params in state",
			params: exchange.DefaultParams(),
			feeAmt: s.coins("100apple,20banana"),
			expAmt: s.coins("5apple,1banana"),
		},
		{
			name: "denom with a specific split: evenly divisible",
			params: &exchange.Params{
				DefaultSplit: 500,
				DenomSplits:  []exchange.DenomSplit{{Denom: "apple", Split: 100}},
			},
			feeAmt: s.coins("500apple"),
			expAmt: s.coins("5apple"),
		},
		{
			name: "denom with a specific split: not evenly divisible",
			params: &exchange.Params{
				DefaultSplit: 500,
				DenomSplits:  []exchange.DenomSplit{{Denom: "apple", Split: 100}},
			},
			feeAmt: s.coins("501apple"),
			expAmt: s.coins("6apple"),
		},
		{
			name: "denom without a specific split: evenly divisible",
			params: &exchange.Params{
				DefaultSplit: 1000,
				DenomSplits:  []exchange.DenomSplit{{Denom: "apple", Split: 100}},
			},
			feeAmt: s.coins("30banana"),
			expAmt: s.coins("3banana"),
		},
		{
			name: "denom without a specific split: not evenly divisible",
			params: &exchange.Params{
				DefaultSplit: 1000,
				DenomSplits:  []exchange.DenomSplit{{Denom: "apple", Split: 100}},
			},
			feeAmt: s.coins("39banana"),
			expAmt: s.coins("4banana"),
		},
		{
			name: "denom with a zero split",
			params: &exchange.Params{
				DefaultSplit: 750,
				DenomSplits:  []exchange.DenomSplit{{Denom: "apple", Split: 0}},
			},
			feeAmt: s.coins("500000apple"),
			expAmt: nil,
		},
		{
			name: "four denoms: two specific, one undefined, one zero",
			params: &exchange.Params{
				DefaultSplit: 650,
				DenomSplits: []exchange.DenomSplit{
					{Denom: "apple", Split: 123},
					{Denom: "banana", Split: 0},
					{Denom: "peach", Split: 55},
				},
			},
			feeAmt: s.coins("123456apple,5000000000banana,400fig,160070peach"),
			// 123456 * 1.23% = 1518.5088, 400 * 6.5% = 26.0, 160070 * 0.55% = 880.385,
			expAmt: s.coins("1519apple,26fig,881peach"),
		},
		{
			name: "zero fee",
			params: &exchange.Params{
				DefaultSplit: 300,
				DenomSplits:  []exchange.DenomSplit{{Denom: "apple", Split: 600}},
			},
			feeAmt: sdk.Coins{sdk.NewInt64Coin("apple", 0), sdk.NewInt64Coin("banana", 0)},
			expAmt: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.k.SetParams(s.ctx, tc.params)

			var actAmt sdk.Coins
			testFunc := func() {
				actAmt = s.k.CalculateExchangeSplit(s.ctx, tc.feeAmt)
			}
			s.Require().NotPanics(testFunc, "CalculateExchangeSplit(%q)", tc.feeAmt)
			s.Assert().Equal(tc.expAmt.String(), actAmt.String(), "CalculateExchangeSplit(%q) result", tc.feeAmt)
		})
	}
}

func (s *TestSuite) TestKeeper_CollectFee() {
	// define our own default params for these tests.
	defaultParams := &exchange.Params{
		DefaultSplit: 250,
		DenomSplits: []exchange.DenomSplit{
			{Denom: "fig", Split: 1000},
			{Denom: "zucchini", Split: 0},
		},
	}

	tests := []struct {
		name     string
		params   *exchange.Params
		bk       *MockBankKeeper
		marketID uint32
		payer    sdk.AccAddress
		feeAmt   sdk.Coins
		expErr   string
		expCalls BankCalls
	}{
		{
			name:     "zero fee",
			marketID: 1,
			payer:    s.addr1,
			feeAmt:   sdk.Coins{sdk.NewInt64Coin("apple", 0)},
			expErr:   "",
			expCalls: BankCalls{},
		},
		{
			name:     "err collecting fee",
			bk:       NewMockBankKeeper().WithSendCoinsResults("test error F from SendCoins"),
			marketID: 1,
			payer:    s.addr1,
			feeAmt:   s.coins("750apple"),
			expErr:   "error transferring 750apple from " + s.addr1.String() + " to market 1: test error F from SendCoins",
			expCalls: BankCalls{
				SendCoinsCalls: []*SendCoinsArgs{
					{ctxHasQuarantineBypass: false, fromAddr: s.addr1, toAddr: s.marketAddr1, amt: s.coins("750apple")},
				},
			},
		},
		{
			name:     "err collecting exchange split",
			bk:       NewMockBankKeeper().WithSendCoinsFromAccountToModuleResults("test error U from SendCoinsFromAccountToModule"),
			marketID: 2,
			payer:    s.addr4,
			feeAmt:   s.coins("750apple"),
			expErr:   "error collecting exchange fee 19apple (based off 750apple) from market 2: test error U from SendCoinsFromAccountToModule",
			expCalls: BankCalls{
				SendCoinsCalls: []*SendCoinsArgs{
					{fromAddr: s.addr4, toAddr: s.marketAddr2, amt: s.coins("750apple")},
				},
				SendCoinsFromAccountToModuleCalls: []*SendCoinsFromAccountToModuleArgs{
					{senderAddr: s.marketAddr2, recipientModule: s.feeCollector, amt: s.coins("19apple")},
				},
			},
		},
		{
			name:     "no exchange split",
			params:   &exchange.Params{DefaultSplit: 0},
			marketID: 3,
			payer:    s.addr2,
			feeAmt:   s.coins("1000000apple,5000000fig"),
			expErr:   "",
			expCalls: BankCalls{
				SendCoinsCalls: []*SendCoinsArgs{
					{fromAddr: s.addr2, toAddr: s.marketAddr3, amt: s.coins("1000000apple,5000000fig")},
				},
			},
		},
		{
			name:     "with exchange split",
			marketID: 1,
			payer:    s.addr3,
			feeAmt:   s.coins("1005apple,5000fig,999999zucchini"),
			expCalls: BankCalls{
				SendCoinsCalls: []*SendCoinsArgs{
					{fromAddr: s.addr3, toAddr: s.marketAddr1, amt: s.coins("1005apple,5000fig,999999zucchini")},
				},
				SendCoinsFromAccountToModuleCalls: []*SendCoinsFromAccountToModuleArgs{
					{senderAddr: s.marketAddr1, recipientModule: s.feeCollector, amt: s.coins("26apple,500fig")},
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.params == nil {
				tc.params = defaultParams
			}
			s.k.SetParams(s.ctx, tc.params)
			if tc.bk == nil {
				tc.bk = NewMockBankKeeper()
			}

			kpr := s.k.WithBankKeeper(tc.bk)
			var err error
			testFunc := func() {
				err = kpr.CollectFee(s.ctx, tc.marketID, tc.payer, tc.feeAmt)
			}
			s.Require().NotPanics(testFunc, "CollectFee(%d, %s, %q)", tc.marketID, s.getAddrName(tc.payer), tc.feeAmt)
			s.assertErrorValue(err, tc.expErr, "CollectFee(%d, %s, %q)", tc.marketID, s.getAddrName(tc.payer), tc.feeAmt)
			s.assertBankKeeperCalls(tc.bk, tc.expCalls, "CollectFee(%d, %s, %q)", tc.marketID, s.getAddrName(tc.payer), tc.feeAmt)
		})
	}
}

func (s *TestSuite) TestKeeper_CollectFees() {
	// define our own default params for these tests.
	defaultParams := &exchange.Params{
		DefaultSplit: 250,
		DenomSplits: []exchange.DenomSplit{
			{Denom: "fig", Split: 1000},
			{Denom: "zucchini", Split: 0},
		},
	}

	tests := []struct {
		name     string
		params   *exchange.Params
		bk       *MockBankKeeper
		marketID uint32
		inputs   []banktypes.Input
		expErr   string
		expCalls BankCalls
	}{
		{
			name:     "nil inputs",
			marketID: 1,
			inputs:   nil,
			expErr:   "",
			expCalls: BankCalls{},
		},
		{
			name:     "nil inputs",
			marketID: 1,
			inputs:   []banktypes.Input{},
			expErr:   "",
			expCalls: BankCalls{},
		},
		{
			name:     "one input: bad address",
			marketID: 2,
			inputs:   []banktypes.Input{{Address: "badAddr", Coins: s.coins("1000apple")}},
			expErr:   "invalid inputs[0] address address \"badAddr\": decoding bech32 failed: invalid bech32 string length 7",
			expCalls: BankCalls{},
		},
		{
			name:     "one input",
			marketID: 2,
			inputs:   []banktypes.Input{{Address: s.addr1.String(), Coins: s.coins("1000apple")}},
			expErr:   "",
			expCalls: BankCalls{
				SendCoinsCalls: []*SendCoinsArgs{
					{fromAddr: s.addr1, toAddr: s.marketAddr2, amt: s.coins("1000apple")},
				},
				SendCoinsFromAccountToModuleCalls: []*SendCoinsFromAccountToModuleArgs{
					{senderAddr: s.marketAddr2, recipientModule: s.feeCollector, amt: s.coins("25apple")},
				},
			},
		},
		{
			name:     "three inputs: zero coins",
			marketID: 3,
			inputs: []banktypes.Input{
				{Address: s.addr3.String(), Coins: sdk.Coins{sdk.NewInt64Coin("apple", 0)}},
				{Address: s.addr4.String(), Coins: sdk.Coins{sdk.NewInt64Coin("fig", 0)}},
				{Address: s.addr5.String(), Coins: sdk.Coins{sdk.NewInt64Coin("zucchini", 0)}},
			},
			expErr:   "",
			expCalls: BankCalls{},
		},
		{
			name:     "three inputs: error from InputOutputCoins",
			bk:       NewMockBankKeeper().WithInputOutputCoinsResults("test error Z from InputOutputCoins"),
			marketID: 1,
			inputs: []banktypes.Input{
				{Address: s.addr1.String(), Coins: s.coins("10apple,1fig,1zucchini")},
				{Address: s.addr3.String(), Coins: s.coins("30fig")},
				{Address: s.addr5.String(), Coins: s.coins("50zucchini")},
			},
			expErr: "error collecting fees for market 1: test error Z from InputOutputCoins",
			expCalls: BankCalls{
				InputOutputCoinsCalls: []*InputOutputCoinsArgs{
					{
						inputs: []banktypes.Input{
							{Address: s.addr1.String(), Coins: s.coins("10apple,1fig,1zucchini")},
							{Address: s.addr3.String(), Coins: s.coins("30fig")},
							{Address: s.addr5.String(), Coins: s.coins("50zucchini")},
						},
						outputs: []banktypes.Output{
							{Address: s.marketAddr1.String(), Coins: s.coins("10apple,31fig,51zucchini")},
						},
					},
				},
			},
		},
		{
			name:     "three inputs: error from SendCoinsFromAccountToModule",
			bk:       NewMockBankKeeper().WithSendCoinsFromAccountToModuleResults("test error L from SendCoinsFromAccountToModule"),
			marketID: 1,
			inputs: []banktypes.Input{
				{Address: s.addr1.String(), Coins: s.coins("1000apple,1fig,10zucchini")},
				{Address: s.addr3.String(), Coins: s.coins("3000fig")},
				{Address: s.addr5.String(), Coins: s.coins("5000zucchini")},
			},
			expErr: "error collecting exchange fee 25apple,301fig (based off 1000apple,3001fig,5010zucchini) from market 1: test error L from SendCoinsFromAccountToModule",
			expCalls: BankCalls{
				InputOutputCoinsCalls: []*InputOutputCoinsArgs{
					{
						inputs: []banktypes.Input{
							{Address: s.addr1.String(), Coins: s.coins("1000apple,1fig,10zucchini")},
							{Address: s.addr3.String(), Coins: s.coins("3000fig")},
							{Address: s.addr5.String(), Coins: s.coins("5000zucchini")},
						},
						outputs: []banktypes.Output{
							{Address: s.marketAddr1.String(), Coins: s.coins("1000apple,3001fig,5010zucchini")},
						},
					},
				},
				SendCoinsFromAccountToModuleCalls: []*SendCoinsFromAccountToModuleArgs{
					{senderAddr: s.marketAddr1, recipientModule: s.feeCollector, amt: s.coins("25apple,301fig")},
				},
			},
		},
		{
			name:     "three inputs: zero split",
			params:   &exchange.Params{DefaultSplit: 0},
			marketID: 2,
			inputs: []banktypes.Input{
				{Address: s.addr1.String(), Coins: s.coins("1000apple,1fig,10zucchini")},
				{Address: s.addr3.String(), Coins: s.coins("3000fig")},
				{Address: s.addr5.String(), Coins: s.coins("5000zucchini")},
			},
			expCalls: BankCalls{
				InputOutputCoinsCalls: []*InputOutputCoinsArgs{
					{
						inputs: []banktypes.Input{
							{Address: s.addr1.String(), Coins: s.coins("1000apple,1fig,10zucchini")},
							{Address: s.addr3.String(), Coins: s.coins("3000fig")},
							{Address: s.addr5.String(), Coins: s.coins("5000zucchini")},
						},
						outputs: []banktypes.Output{
							{Address: s.marketAddr2.String(), Coins: s.coins("1000apple,3001fig,5010zucchini")},
						},
					},
				},
			},
		},
		{
			name:     "three inputs: with split",
			marketID: 3,
			inputs: []banktypes.Input{
				{Address: s.addr1.String(), Coins: s.coins("1000apple,1fig,10zucchini")},
				{Address: s.addr3.String(), Coins: s.coins("3000fig")},
				{Address: s.addr5.String(), Coins: s.coins("5000zucchini")},
			},
			expCalls: BankCalls{
				InputOutputCoinsCalls: []*InputOutputCoinsArgs{
					{
						inputs: []banktypes.Input{
							{Address: s.addr1.String(), Coins: s.coins("1000apple,1fig,10zucchini")},
							{Address: s.addr3.String(), Coins: s.coins("3000fig")},
							{Address: s.addr5.String(), Coins: s.coins("5000zucchini")},
						},
						outputs: []banktypes.Output{
							{Address: s.marketAddr3.String(), Coins: s.coins("1000apple,3001fig,5010zucchini")},
						},
					},
				},
				SendCoinsFromAccountToModuleCalls: []*SendCoinsFromAccountToModuleArgs{
					{senderAddr: s.marketAddr3, recipientModule: s.feeCollector, amt: s.coins("25apple,301fig")},
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.params == nil {
				tc.params = defaultParams
			}
			s.k.SetParams(s.ctx, tc.params)
			if tc.bk == nil {
				tc.bk = NewMockBankKeeper()
			}

			kpr := s.k.WithBankKeeper(tc.bk)
			var err error
			testFunc := func() {
				err = kpr.CollectFees(s.ctx, tc.marketID, tc.inputs)
			}
			s.Require().NotPanics(testFunc, "CollectFees(%d, ...)", tc.marketID)
			s.assertErrorValue(err, tc.expErr, "CollectFees(%d, ...)", tc.marketID)
			s.assertBankKeeperCalls(tc.bk, tc.expCalls, "CollectFees(%d, ...)", tc.marketID)
		})
	}
}
