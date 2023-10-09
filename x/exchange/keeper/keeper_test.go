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
	sdkCtx    sdk.Context
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
}

func (s *TestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.sdkCtx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.stdlibCtx = sdk.WrapSDKContext(s.sdkCtx)
	s.k = s.app.ExchangeKeeper
	s.acctKeeper = s.app.AccountKeeper
	s.attrKeeper = s.app.AttributeKeeper
	s.bankKeeper = s.app.BankKeeper
	s.holdKeeper = s.app.HoldKeeper

	s.bondDenom = s.app.StakingKeeper.BondDenom(s.sdkCtx)
	s.initAmount = 1_000_000_000
	s.initBal = sdk.NewCoins(sdk.NewCoin(s.bondDenom, sdk.NewInt(s.initAmount)))

	addrs := app.AddTestAddrsIncremental(s.app, s.sdkCtx, 5, sdk.NewInt(s.initAmount))
	s.addr1 = addrs[0]
	s.addr2 = addrs[1]
	s.addr3 = addrs[2]
	s.addr4 = addrs[3]
	s.addr5 = addrs[4]
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

// coins creates an sdk.Coins from a string, requiring it to work.
func (s *TestSuite) coins(coins string) sdk.Coins {
	s.T().Helper()
	rv, err := sdk.ParseCoinsNormalized(coins)
	s.Require().NoError(err, "ParseCoinsNormalized(%q)", coins)
	return rv
}

// coin creates a new coin without doing any validation on it.
func (s *TestSuite) coin(amount int64, denom string) sdk.Coin {
	return sdk.Coin{
		Amount: s.int(amount),
		Denom:  denom,
	}
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
	return s.k.GetStore(s.sdkCtx)
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
		return testutil.FundAccount(s.app.BankKeeper, s.sdkCtx, addr, s.coins(coins))
	}, "FundAccount(%s, %q)", s.getAddrName(addr), coins)
}

// assertErrorValue is a wrapper for assertions.AssertErrorValue.
func (s *TestSuite) assertErrorValue(theError error, expected string, msgAndArgs ...interface{}) bool {
	return assertions.AssertErrorValue(s.T(), theError, expected, msgAndArgs...)
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

// TODO[1658]: func (s *TestSuite) TestKeeper_DoTransfer()

// TODO[1658]: func (s *TestSuite) TestKeeper_CalculateExchangeSplit()

// TODO[1658]: func (s *TestSuite) TestKeeper_CollectFee()

// TODO[1658]: func (s *TestSuite) TestKeeper_CollectFees()
