package keeper_test

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/escrow"
	"github.com/provenance-io/provenance/x/escrow/keeper"
)

type TestSuite struct {
	suite.Suite

	app        *app.App
	sdkCtx     sdk.Context
	stdlibCtx  context.Context
	keeper     keeper.Keeper
	bankKeeper bankkeeper.Keeper

	bondDenom string
	initBal   sdk.Coins

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
	s.keeper = s.app.EscrowKeeper
	s.bankKeeper = s.app.BankKeeper

	s.bondDenom = s.app.StakingKeeper.BondDenom(s.sdkCtx)
	initAmount := sdk.NewInt(1_000_000_000)
	s.initBal = sdk.NewCoins(sdk.NewCoin(s.bondDenom, initAmount))

	addrs := app.AddTestAddrsIncremental(s.app, s.sdkCtx, 5, initAmount)
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

// AssertErrorContents asserts that the provided error is as expected.
// If contains is empty, it asserts there is no error.
// Otherwise, it asseerts that the error contains each of the entries in the contains slice.
// Returns true if it's all good, false if one or more assertion failed.
func (s *TestSuite) AssertErrorContents(theError error, contains []string, msgAndArgs ...interface{}) bool {
	s.T().Helper()
	if len(contains) == 0 {
		return s.Assert().NoError(theError, msgAndArgs...)
	}
	if !s.Assert().Error(theError, msgAndArgs...) {
		s.T().Logf("Error was expected to contain:\n\t\t\"%s\"", strings.Join(contains, "\"\n\t\t\""))
		return false
	}

	hasAll := true
	for _, expInErr := range contains {
		hasAll = s.Assert().ErrorContains(theError, expInErr, msgAndArgs...) && hasAll
	}
	return hasAll
}

// panicTestFunc is a type declaration for a function that will be tested for panic.
type panicTestFunc func()

// didPanic safely executes the provided function and returns info about any panic it might have encountered.
func didPanic(f panicTestFunc) (didPanic bool, message interface{}, stack string) {
	didPanic = true

	defer func() {
		message = recover()
		if didPanic {
			stack = string(debug.Stack())
		}
	}()

	f()
	didPanic = false

	return
}

// AssertPanicContents asserts that, if contains is empty, the provided func does not panic
// Otherwise, asserts that the func panics and that its panic message contains each of the provided strings.
func (s *TestSuite) AssertPanicContents(f panicTestFunc, contains []string, msgAndArgs ...interface{}) bool {
	s.T().Helper()

	funcDidPanic, panicValue, panickedStack := didPanic(f)
	panicMsg := fmt.Sprintf("%v", panicValue)

	if len(contains) == 0 {
		if !funcDidPanic {
			return true
		}
		msg := fmt.Sprintf("func %#v should not panic, but did.", f)
		msg += fmt.Sprintf("\n\tPanic message:\t%q", panicMsg)
		msg += fmt.Sprintf("\n\t  Panic value:\t%#v", panicValue)
		msg += fmt.Sprintf("\n\t  Panic stack:\t%s", panickedStack)
		return s.Assert().Fail(msg, msgAndArgs...)
	}

	if !funcDidPanic {
		msg := fmt.Sprintf("func %#v should panic, but did not.", f)
		for _, exp := range contains {
			msg += fmt.Sprintf("\n\tExpected to contain:\t%q", exp)
		}
		return s.Assert().Fail(msg, msgAndArgs...)
	}

	var missing []string
	for _, exp := range contains {
		if !strings.Contains(panicMsg, exp) {
			missing = append(missing, exp)
		}
	}

	if len(missing) == 0 {
		return true
	}

	msg := fmt.Sprintf("func %#v panic message incorrect.", f)
	msg += fmt.Sprintf("\n\t   Panic message:\t%q", panicMsg)
	for _, exp := range missing {
		msg += fmt.Sprintf("\n\tDoes not contain:\t%q", exp)
	}
	msg += fmt.Sprintf("\n\tPanic value:\t%#v", panicValue)
	msg += fmt.Sprintf("\n\tPanic stack:\t%s", panickedStack)
	return s.Assert().Fail(msg, msgAndArgs)
}

// RequirePanicContents asserts that, if contains is empty, the provided func does not panic
// Otherwise, asserts that the func panics and that its panic message contains each of the provided strings.
//
// If the assertion fails, the test is halted.
func (s *TestSuite) RequirePanicContents(f panicTestFunc, contains []string, msgAndArgs ...interface{}) {
	s.T().Helper()
	if s.AssertPanicContents(f, contains, msgAndArgs...) {
		return
	}
	s.T().FailNow()
}

func (s *TestSuite) TestKeeper_ValidateNewEscrow() {
	tests := []struct {
		name      string
		addr      sdk.AccAddress
		funds     sdk.Coins
		spendable sdk.Coins
		expErr    []string
	}{
		{
			name:      "nil funds",
			addr:      s.addr1,
			funds:     nil,
			spendable: s.coins("123rake"),
		},
		{
			name:      "empty funds",
			addr:      s.addr1,
			funds:     sdk.Coins{},
			spendable: s.coins("123bake"),
		},
		{
			name:      "two zero coins",
			addr:      s.addr1,
			funds:     sdk.Coins{s.coin(0, "acorn"), s.coin(0, "boin")},
			spendable: s.coins("123fake"),
		},
		{
			name:      "with negative coin",
			addr:      s.addr1,
			funds:     sdk.Coins{s.coin(10, "acorn"), s.coin(-3, "boin"), s.coin(22, "corn")},
			spendable: s.coins("10acorn,5boin,100corn"),
			expErr:    []string{"10acorn,-3boin,22corn", "escrow amounts", "cannot be negative", s.addr1.String()},
		},
		{
			name:      "no spendable for one coin",
			addr:      s.addr2,
			funds:     s.coins("10acorn,5boin,100corn"),
			spendable: s.coins("10acorn,100corn"),
			expErr:    []string{"spendable balance 0boin is less than escrow amount 5boin", s.addr2.String()},
		},
		{
			name:      "not enough spendable for a coin",
			addr:      s.addr3,
			funds:     s.coins("10acorn,5boin,100corn"),
			spendable: s.coins("10acorn,4boin,100corn"),
			expErr:    []string{"spendable balance 4boin is less than escrow amount 5boin", s.addr3.String()},
		},
		{
			name:      "all spendable of one coin being put in escrow",
			addr:      s.addr5,
			funds:     s.coins("5boin"),
			spendable: s.coins("10acorn,5boin,100corn"),
		},
		{
			name:      "all spendable being put in escrow",
			addr:      s.addr4,
			funds:     s.coins("10acorn,5boin,100corn"),
			spendable: s.coins("10acorn,5boin,100corn"),
		},
		{
			name:      "a zero coin that is not in spendable",
			addr:      s.addr5,
			funds:     sdk.Coins{s.coin(18, "acorn"), s.coin(0, "boin"), s.coin(55, "corn")},
			spendable: s.coins("20acorn,100corn"),
		},
		{
			name:      "three coins: all enough",
			addr:      s.addr1,
			funds:     s.coins("10acorn,20boin,30corn"),
			spendable: s.coins("10acorn,20boin,30corn"),
		},
		{
			name:      "three coins: none enough",
			addr:      s.addr1,
			funds:     s.coins("11acorn,21boin,31corn"),
			spendable: s.coins("10acorn,20boin,30corn"),
			expErr:    []string{"spendable balance 10acorn is less than escrow amount 11acorn", s.addr1.String()},
		},
		{
			name:      "three coins: first insufficient",
			addr:      s.addr1,
			funds:     s.coins("11acorn,20boin,30corn"),
			spendable: s.coins("10acorn,20boin,30corn"),
			expErr:    []string{"spendable balance 10acorn is less than escrow amount 11acorn", s.addr1.String()},
		},
		{
			name:      "three coins: second insufficient",
			addr:      s.addr1,
			funds:     s.coins("10acorn,21boin,30corn"),
			spendable: s.coins("10acorn,20boin,30corn"),
			expErr:    []string{"spendable balance 20boin is less than escrow amount 21boin", s.addr1.String()},
		},
		{
			name:      "three coins: third insufficient",
			addr:      s.addr1,
			funds:     s.coins("10acorn,20boin,31corn"),
			spendable: s.coins("10acorn,20boin,30corn"),
			expErr:    []string{"spendable balance 30corn is less than escrow amount 31corn", s.addr1.String()},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			bk := NewMockBankKeeper().WithSpendable(tc.addr, tc.spendable)
			k := s.app.EscrowKeeper.WithBankKeeper(bk)

			var err error
			testFunc := func() {
				err = k.ValidateNewEscrow(s.sdkCtx, tc.addr, tc.funds)
			}
			s.Require().NotPanics(testFunc, "ValidateNewEscrow")
			s.AssertErrorContents(err, tc.expErr, "ValidateNewEscrow")
		})
	}
}

func (s *TestSuite) TestKeeper_AddEscrow() {
	store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr1, "banana", s.int(99)), "setEscrowCoinAmount(addr1, 99banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr1, "cucumber", s.int(3)), "setEscrowCoinAmount(addr1, 3cucumber)")
	// max uint64 = 18,446,744,073,709,551,615
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr2, "hugecoin", s.intStr("1844674407370955161500")), "setEscrowCoinAmount(addr2, 1844674407370955161500hugecoin)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr2, "mediumcoin", s.intStr("10000000000000000000")), "setEscrowCoinAmount(addr2, 10000000000000000000mediumcoin)")
	store.Set(keeper.CreateEscrowCoinKey(s.addr3, "badcoin"), []byte("badvalue"))
	store.Set(keeper.CreateEscrowCoinKey(s.addr3, "crudcoin"), []byte("crudvalue"))
	store = nil

	// Tests are ordered by address since the spendable balance depends on the previous state.
	tests := []struct {
		name     string
		addr     sdk.AccAddress
		funds    sdk.Coins
		spendBal sdk.Coins
		expErr   []string
		finalEsc sdk.Coins
	}{
		{
			name:     "nil funds",
			addr:     s.addr1,
			funds:    nil,
			finalEsc: s.coins("99banana,3cucumber"),
		},
		{
			name:     "empty funds",
			addr:     s.addr1,
			funds:    sdk.Coins{},
			finalEsc: s.coins("99banana,3cucumber"),
		},
		{
			name:     "insufficent spendable: some already in escrow",
			addr:     s.addr1,
			funds:    s.coins("2cucumber"),
			spendBal: s.coins("1cucumber"),
			expErr:   []string{"spendable balance 1cucumber is less than escrow amount 2cucumber"},
			finalEsc: s.coins("99banana,3cucumber"),
		},
		{
			name:     "sufficient spendable: add to existing entry",
			addr:     s.addr1,
			funds:    s.coins("2banana"),
			spendBal: s.coins("2banana,9cucumber,11durian"),
			finalEsc: s.coins("101banana,3cucumber"),
		},
		{
			name:     "small amount added to existing amount over max uint64",
			addr:     s.addr2,
			funds:    s.coins("99hugecoin"),
			spendBal: s.coins("5000000000000000000000hugecoin"),
			finalEsc: s.coins("1844674407370955161599hugecoin,10000000000000000000mediumcoin"),
		},
		{
			name:     "amount over max uint64 added to existing amount over max uint64",
			addr:     s.addr2,
			funds:    s.coins("2000000000000000000000hugecoin"),
			spendBal: s.coins("5000000000000000000000hugecoin"),
			finalEsc: s.coins("3844674407370955161599hugecoin,10000000000000000000mediumcoin"),
		},
		{
			name:     "amount over max uint64 added to new entry",
			addr:     s.addr2,
			funds:    s.coins("18446744073709551616bigcoin"),
			spendBal: s.coins("20000000000000000000bigcoin"),
			finalEsc: s.coins("18446744073709551616bigcoin,3844674407370955161599hugecoin,10000000000000000000mediumcoin"),
		},
		{
			name:     "amount under max uint64 added to another such amount resulting in more than max uint64",
			addr:     s.addr2,
			funds:    s.coins("10000000000000000000mediumcoin"),
			spendBal: s.coins("10000000000000000000mediumcoin"),
			finalEsc: s.coins("18446744073709551616bigcoin,3844674407370955161599hugecoin,20000000000000000000mediumcoin"),
		},
		{
			name:     "existing entry is invalid",
			addr:     s.addr3,
			funds:    s.coins("1badcoin"),
			spendBal: s.coins("1badcoin"),
			expErr: []string{
				"failed to get current badcoin escrow amount",
				"math/big: cannot unmarshal \"badvalue\" into a *big.Int",
			},
		},
		{
			name:     "existing entry is invalid but new is zero",
			addr:     s.addr3,
			funds:    s.coins("0badcoin"),
			spendBal: s.coins("1badcoin"),
		},
		{
			name:     "addr has bad entry but adding different denom",
			addr:     s.addr3,
			funds:    s.coins("4goodcoin"),
			spendBal: s.coins("1badcoin,2banana,4goodcoin"),
			finalEsc: s.coins("4goodcoin"),
		},
		{
			name:     "zero of bad denom with some of another",
			addr:     s.addr3,
			funds:    s.coins("0badcoin,8goodcoin"),
			spendBal: s.coins("8goodcoin"),
			finalEsc: s.coins("12goodcoin"),
		},
		{
			name:     "three denoms: two existing and bad",
			addr:     s.addr3,
			funds:    s.coins("57acorn,5badcoin,4crudcoin"),
			spendBal: s.coins("100acorn,100badcoin,100crudcoin"),
			expErr: []string{
				"failed to get current badcoin escrow amount",
				"math/big: cannot unmarshal \"badvalue\" into a *big.Int",
				"failed to get current crudcoin escrow amount",
				"math/big: cannot unmarshal \"crudvalue\" into a *big.Int",
			},
			finalEsc: s.coins("57acorn,12goodcoin"),
		},
		{
			name:     "sufficient spendable: new denoms to escrow",
			addr:     s.addr4,
			funds:    s.coins("37acorn,12banana"),
			spendBal: s.coins("37acorn,12banana"),
			finalEsc: s.coins("37acorn,12banana"),
		},
		{
			name:     "amount over max uint64 added to amount under uint64",
			addr:     s.addr4,
			funds:    s.coins("5000000000000000000000banana"),
			spendBal: s.coins("5000000000000000000000banana"),
			finalEsc: s.coins("37acorn,5000000000000000000012banana"),
		},
		{
			name:  "zero funds",
			addr:  s.addr5,
			funds: sdk.Coins{s.coin(0, "banana"), s.coin(0, "cucumber")},
		},
		{
			name:     "insufficient spendable: none in escrow yet",
			addr:     s.addr5,
			funds:    s.coins("49apple"),
			spendBal: s.coins("48apple"),
			expErr:   []string{"spendable balance 48apple is less than escrow amount 49apple"},
		},
		{
			name:     "new amount is invalid",
			addr:     s.addr5,
			funds:    sdk.Coins{s.coin(-1, "banana")},
			spendBal: s.coins("2banana"),
			expErr:   []string{s.addr5.String(), "-1banana", "cannot be negative"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if len(tc.expErr) > 0 {
				tc.expErr = append(tc.expErr, tc.addr.String())
			}
			bk := NewMockBankKeeper().WithSpendable(tc.addr, tc.spendBal)
			k := s.keeper.WithBankKeeper(bk)

			var err error
			testFunc := func() {
				err = k.AddEscrow(s.sdkCtx, tc.addr, tc.funds)
			}
			s.Require().NotPanics(testFunc, "AddEscrow")

			s.AssertErrorContents(err, tc.expErr, "AddEscrow error")

			finalEsc, _ := k.GetEscrowCoins(s.sdkCtx, tc.addr)
			s.Assert().Equal(tc.finalEsc.String(), finalEsc.String(), "final escrow")
		})
	}
}

func (s *TestSuite) TestKeeper_RemoveEscrow() {
	store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr1, "banana", s.int(99)), "setEscrowCoinAmount(addr1, 99banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr1, "cucumber", s.int(3)), "setEscrowCoinAmount(addr1, 3cucumber)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr2, "banana", s.int(18)), "setEscrowCoinAmount(addr2, 18banana)")
	store.Set(keeper.CreateEscrowCoinKey(s.addr3, "badcoin"), []byte("badvalue"))
	store.Set(keeper.CreateEscrowCoinKey(s.addr3, "crudcoin"), []byte("crudvalue"))
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr3, "goodcoin", s.int(2)), "setEscrowCoinAmount(addr3, 2goodcoin)")
	// max uint64 = 18,446,744,073,709,551,615
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "hugecoin", s.intStr("1844674407370955161500")), "setEscrowCoinAmount(addr4, 1844674407370955161500hugecoin)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "largecoin", s.intStr("1000000000000000000000")), "setEscrowCoinAmount(addr4, 1000000000000000000000largecoin)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "mediumcoin", s.intStr("20000000000000000000")), "setEscrowCoinAmount(addr4, 20000000000000000000mediumcoin)")
	store = nil

	// Tests are ordered by address since the spendable balance depends on the previous state.
	tests := []struct {
		name     string
		addr     sdk.AccAddress
		funds    sdk.Coins
		expErr   []string
		finalEsc sdk.Coins
	}{
		{
			name:     "remove some of two denoms",
			addr:     s.addr1,
			funds:    s.coins("1banana,1cucumber"),
			finalEsc: s.coins("98banana,2cucumber"),
		},
		{
			name:  "remove all of two denoms",
			addr:  s.addr1,
			funds: s.coins("98banana,2cucumber"),
		},
		{
			name:  "not enough in escrow",
			addr:  s.addr2,
			funds: s.coins("20banana"),
			expErr: []string{
				"cannot remove 20banana from escrow",
				"account only has 18banana in escrow",
			},
			finalEsc: s.coins("18banana"),
		},
		{
			name:     "only remove some",
			addr:     s.addr2,
			funds:    s.coins("10banana"),
			finalEsc: s.coins("8banana"),
		},
		{
			name:  "remove all of one denom",
			addr:  s.addr2,
			funds: s.coins("8banana"),
		},
		{
			name:  "bad existing entry",
			addr:  s.addr3,
			funds: s.coins("1badcoin"),
			expErr: []string{
				"failed to get current badcoin escrow amount",
				"math/big: cannot unmarshal \"badvalue\" into a *big.Int",
			},
			finalEsc: s.coins("2goodcoin"),
		},
		{
			name:     "bad existing entry but amount of that denom is zero",
			addr:     s.addr3,
			funds:    s.coins("0badcoin,1goodcoin"),
			finalEsc: s.coins("1goodcoin"),
		},
		{
			name:  "two bad denoms one good",
			addr:  s.addr3,
			funds: s.coins("1badcoin,2crudcoin,1goodcoin"),
			expErr: []string{
				"failed to get current badcoin escrow amount",
				"math/big: cannot unmarshal \"badvalue\" into a *big.Int",
				"failed to get current crudcoin escrow amount",
				"math/big: cannot unmarshal \"crudvalue\" into a *big.Int",
			},
		},
		{
			name:     "amount left in escrow still greater than max uint64",
			addr:     s.addr4,
			funds:    s.coins("1hugecoin"),
			finalEsc: s.coins("1844674407370955161499hugecoin,1000000000000000000000largecoin,20000000000000000000mediumcoin"),
		},
		{
			name:     "amount removed is greater than max uint64",
			addr:     s.addr4,
			funds:    s.coins("1844674407370955161400hugecoin"),
			finalEsc: s.coins("99hugecoin,1000000000000000000000largecoin,20000000000000000000mediumcoin"),
		},
		{
			name:     "exising amount more than max uint64 and amount removed is less with result also less",
			addr:     s.addr4,
			funds:    s.coins("10000000000000000000mediumcoin"),
			finalEsc: s.coins("99hugecoin,1000000000000000000000largecoin,10000000000000000000mediumcoin"),
		},
		{
			name:     "amount removed is more than max uint64 with result also more",
			addr:     s.addr4,
			funds:    s.coins("100000000000000000000largecoin"),
			finalEsc: s.coins("99hugecoin,900000000000000000000largecoin,10000000000000000000mediumcoin"),
		},
		{
			name:  "existing amount and amount to remove over max uint64 but insufficient",
			addr:  s.addr4,
			funds: s.coins("900000000000000000001largecoin"),
			expErr: []string{
				"cannot remove 900000000000000000001largecoin from escrow",
				"account only has 900000000000000000000largecoin in escrow",
			},
			finalEsc: s.coins("99hugecoin,900000000000000000000largecoin,10000000000000000000mediumcoin"),
		},
		{
			name:  "nil funds",
			addr:  s.addr5,
			funds: nil,
		},
		{
			name:  "empty funds",
			addr:  s.addr5,
			funds: sdk.Coins{},
		},
		{
			name:  "zero funds",
			addr:  s.addr5,
			funds: sdk.Coins{s.coin(0, "banana"), s.coin(0, "cucumber")},
		},
		{
			name:   "negative amount",
			addr:   s.addr5,
			funds:  sdk.Coins{s.coin(-1, "banana")},
			expErr: []string{"cannot remove \"-1banana\" from escrow", "amounts cannot be negative"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if len(tc.expErr) > 0 {
				tc.expErr = append(tc.expErr, tc.addr.String())
			}

			var err error
			testFunc := func() {
				err = s.keeper.RemoveEscrow(s.sdkCtx, tc.addr, tc.funds)
			}
			s.Require().NotPanics(testFunc, "RemoveEscrow")

			s.AssertErrorContents(err, tc.expErr, "RemoveEscrow error")

			finalEsc, _ := s.keeper.GetEscrowCoins(s.sdkCtx, tc.addr)
			s.Assert().Equal(tc.finalEsc.String(), finalEsc.String(), "final escrow")
		})
	}
}

func (s *TestSuite) TestKeeper_GetEscrowCoin() {
	store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr1, "banana", s.int(99)), "setEscrowCoinAmount(addr1, 99banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr1, "cucumber", s.int(3)), "setEscrowCoinAmount(addr1, 3cucumber)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr2, "banana", s.int(18)), "setEscrowCoinAmount(addr2, 18banana)")
	store.Set(keeper.CreateEscrowCoinKey(s.addr1, "badcoin"), []byte("badvalue"))
	store = nil

	tests := []struct {
		name    string
		addr    sdk.AccAddress
		denom   string
		expCoin sdk.Coin
		expErr  []string
	}{
		{
			name:    "nothing in escrow for addr",
			addr:    s.addr5,
			denom:   "nonecoin",
			expCoin: s.coin(0, "nonecoin"),
		},
		{
			name:    "addr has escrow but not this denom",
			addr:    s.addr2,
			denom:   "cucumber",
			expCoin: s.coin(0, "cucumber"),
		},
		{
			name:    "addr has only this denom in escrow",
			addr:    s.addr2,
			denom:   "banana",
			expCoin: s.coin(18, "banana"),
		},
		{
			name:    "addr has multiple denoms in escrow but not this one",
			addr:    s.addr1,
			denom:   "nonecoin",
			expCoin: s.coin(0, "nonecoin"),
		},
		{
			name:    "addr has multiple denoms in escrow this denom also in escrow by other addr",
			addr:    s.addr1,
			denom:   "banana",
			expCoin: s.coin(99, "banana"),
		},
		{
			name:    "addr has multiple denoms in escrow this denom is only in escrow by this addr",
			addr:    s.addr1,
			denom:   "cucumber",
			expCoin: s.coin(3, "cucumber"),
		},
		{
			name:    "bad value",
			addr:    s.addr1,
			denom:   "badcoin",
			expCoin: s.coin(0, "badcoin"),
			expErr:  []string{"could not get escrow coin amount for", s.addr1.String(), "math/big: cannot unmarshal \"badvalue\" into a *big.Int"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var coin sdk.Coin
			var err error
			testFunc := func() {
				coin, err = s.keeper.GetEscrowCoin(s.sdkCtx, tc.addr, tc.denom)
			}
			s.Require().NotPanics(testFunc, "GetEscrowCoin")
			s.AssertErrorContents(err, tc.expErr, "GetEscrowCoin error")
			s.Assert().Equal(tc.expCoin.String(), coin.String(), "GetEscrowCoin coin")
		})
	}
}

func (s *TestSuite) TestKeeper_GetEscrowCoins() {
	store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr1, "banana", s.int(99)), "setEscrowCoinAmount(addr1, 99banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr2, "banana", s.int(18)), "setEscrowCoinAmount(addr2, 18banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr2, "cucumber", s.int(3)), "setEscrowCoinAmount(addr2, 3cucumber)")
	store.Set(keeper.CreateEscrowCoinKey(s.addr3, "grimcoin"), []byte("grimvalue"))
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "acorn", s.int(52)), "setEscrowCoinAmount(addr4, 52acorn)")
	store.Set(keeper.CreateEscrowCoinKey(s.addr4, "badcoin"), []byte("badvalue"))
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "cucumber", s.int(12)), "setEscrowCoinAmount(addr4, 12cucumber)")
	store.Set(keeper.CreateEscrowCoinKey(s.addr4, "dreadcoin"), []byte("dreadvalue"))
	store = nil

	tests := []struct {
		name     string
		addr     sdk.AccAddress
		expCoins sdk.Coins
		expErr   []string
	}{
		{
			name:     "addr with only one denom",
			addr:     s.addr1,
			expCoins: s.coins("99banana"),
		},
		{
			name:     "addr with two denoms",
			addr:     s.addr2,
			expCoins: s.coins("18banana,3cucumber"),
		},
		{
			name:     "addr with only one denom and it's bad",
			addr:     s.addr3,
			expCoins: nil,
			expErr: []string{
				s.addr3.String(),
				"failed to read amount of grimcoin",
				"math/big: cannot unmarshal \"grimvalue\" into a *big.Int",
			},
		},
		{
			name:     "addr with two good denoms and two bad ones",
			addr:     s.addr4,
			expCoins: s.coins("52acorn,12cucumber"),
			expErr: []string{
				s.addr4.String(),
				"failed to read amount of badcoin",
				"math/big: cannot unmarshal \"badvalue\" into a *big.Int",
				"failed to read amount of dreadcoin",
				"math/big: cannot unmarshal \"dreadvalue\" into a *big.Int",
			},
		},
		{
			name:     "addr without anything",
			addr:     s.addr5,
			expCoins: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var coins sdk.Coins
			var err error
			testFunc := func() {
				coins, err = s.keeper.GetEscrowCoins(s.sdkCtx, tc.addr)
			}
			s.Require().NotPanics(testFunc, "GetEscrowCoins")
			s.AssertErrorContents(err, tc.expErr, "GetEscrowCoins error")
			s.Assert().Equal(tc.expCoins.String(), coins.String(), "GetEscrowCoins coins")
		})
	}
}

func (s *TestSuite) TestKeeper_IterateEscrow() {
	store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr1, "banana", s.int(99)), "setEscrowCoinAmount(addr1, 99banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr2, "banana", s.int(18)), "setEscrowCoinAmount(addr2, 18banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr2, "cucumber", s.int(3)), "setEscrowCoinAmount(addr2, 3cucumber)")
	store.Set(keeper.CreateEscrowCoinKey(s.addr3, "grimcoin"), []byte("grimvalue"))
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "acorn", s.int(52)), "setEscrowCoinAmount(addr4, 52acorn)")
	store.Set(keeper.CreateEscrowCoinKey(s.addr4, "badcoin"), []byte("badvalue"))
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "cucumber", s.int(12)), "setEscrowCoinAmount(addr4, 12cucumber)")
	store.Set(keeper.CreateEscrowCoinKey(s.addr4, "dreadcoin"), []byte("dreadvalue"))
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "eggplant", s.int(747)), "setEscrowCoinAmount(addr4, 747eggplant)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "acorn", s.int(358)), "setEscrowCoinAmount(addr5, 358acorn)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "banana", s.int(101)), "setEscrowCoinAmount(addr5, 101banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "cucumber", s.int(8)), "setEscrowCoinAmount(addr5, 8cucumber)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "durian", s.int(5)), "setEscrowCoinAmount(addr5, 5durian)")
	store = nil

	addrDNE := sdk.AccAddress("addr_does_not_exist_")

	var processed []string
	stopAfter := func(count int) func(coin sdk.Coin) bool {
		return func(coin sdk.Coin) bool {
			processed = append(processed, coin.String())
			return len(processed) >= count
		}
	}
	getAll := func(coin sdk.Coin) bool {
		processed = append(processed, coin.String())
		return false
	}

	tests := []struct {
		name        string
		addr        sdk.AccAddress
		process     func(sdk.Coin) bool
		expProc     []string
		expErr      []string
		expNotInErr []string
	}{
		{
			name:    "address is unknown",
			addr:    addrDNE,
			process: getAll,
			expProc: nil,
		},
		{
			name:    "address has one entry",
			addr:    s.addr1,
			process: getAll,
			expProc: []string{"99banana"},
		},
		{
			name:    "address has two entries: get all",
			addr:    s.addr2,
			process: getAll,
			expProc: []string{"18banana", "3cucumber"},
		},
		{
			name:    "address has two entries: stop after first",
			addr:    s.addr2,
			process: stopAfter(1),
			expProc: []string{"18banana"},
		},
		{
			name:    "address has single bad entry",
			addr:    s.addr3,
			process: getAll,
			expProc: nil,
			expErr: []string{
				s.addr3.String(),
				"failed to read amount of grimcoin",
				"math/big: cannot unmarshal \"grimvalue\" into a *big.Int",
			},
		},
		{
			name:    "address has three good and two bad: get all",
			addr:    s.addr4,
			process: getAll,
			expProc: []string{"52acorn", "12cucumber", "747eggplant"},
			expErr: []string{
				s.addr4.String(),
				"failed to read amount of badcoin",
				"math/big: cannot unmarshal \"badvalue\" into a *big.Int",
				"failed to read amount of dreadcoin",
				"math/big: cannot unmarshal \"dreadvalue\" into a *big.Int",
			},
		},
		{
			name:    "address has three good and two bad: stop after first",
			addr:    s.addr4,
			process: stopAfter(1),
			expProc: []string{"52acorn"},
			expErr:  nil,
		},
		{
			name:    "address has three good and two bad: stop after second",
			addr:    s.addr4,
			process: stopAfter(2),
			expProc: []string{"52acorn", "12cucumber"},
			expErr: []string{
				s.addr4.String(),
				"failed to read amount of badcoin",
				"math/big: cannot unmarshal \"badvalue\" into a *big.Int",
			},
			expNotInErr: []string{
				"failed to read amount of dreadcoin",
				"math/big: cannot unmarshal \"dreadvalue\" into a *big.Int",
			},
		},
		{
			name:    "address has four entries: get all",
			addr:    s.addr5,
			process: getAll,
			expProc: []string{"358acorn", "101banana", "8cucumber", "5durian"},
		},
		{
			name:    "address has four entries: stop after 3",
			addr:    s.addr5,
			process: stopAfter(3),
			expProc: []string{"358acorn", "101banana", "8cucumber"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			processed = nil
			var err error
			testFunc := func() {
				err = s.keeper.IterateEscrow(s.sdkCtx, tc.addr, tc.process)
			}
			s.Require().NotPanics(testFunc, "IterateEscrow")
			s.AssertErrorContents(err, tc.expErr, "IterateEscrow error")
			if err != nil && len(tc.expNotInErr) > 0 {
				errStr := err.Error()
				for _, unexp := range tc.expNotInErr {
					s.Assert().NotContains(errStr, unexp, "IterateEscrow error")
				}
			}
			s.Assert().Equal(tc.expProc, processed, "IterateEscrow entries processed")
		})
	}
}

func (s *TestSuite) TestKeeper_IterateAllEscrow() {
	// Since the addresses should have been created sequentially, that's the order the store records should be in.
	// I also picked easy-to-sort coin names.
	// That means that the order they're being defined here should be the order they are in state.
	store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr1, "banana", s.int(99)), "setEscrowCoinAmount(addr1, 99banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr2, "banana", s.int(18)), "setEscrowCoinAmount(addr2, 18banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr2, "cucumber", s.int(3)), "setEscrowCoinAmount(addr2, 3cucumber)")
	store.Set(keeper.CreateEscrowCoinKey(s.addr3, "grimcoin"), []byte("grimvalue"))
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "acorn", s.int(52)), "setEscrowCoinAmount(addr4, 52acorn)")
	store.Set(keeper.CreateEscrowCoinKey(s.addr4, "badcoin"), []byte("badvalue"))
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "cucumber", s.int(12)), "setEscrowCoinAmount(addr4, 12cucumber)")
	store.Set(keeper.CreateEscrowCoinKey(s.addr4, "dreadcoin"), []byte("dreadvalue"))
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "eggplant", s.int(747)), "setEscrowCoinAmount(addr4, 747eggplant)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "acorn", s.int(358)), "setEscrowCoinAmount(addr5, 358acorn)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "banana", s.int(101)), "setEscrowCoinAmount(addr5, 101banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "cucumber", s.int(8)), "setEscrowCoinAmount(addr5, 8cucumber)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "durian", s.int(5)), "setEscrowCoinAmount(addr5, 5durian)")
	store = nil

	entry := func(addr sdk.AccAddress, coin string) string {
		return addr.String() + ":" + coin
	}
	var processed []string
	stopAfter := func(count int) func(addr sdk.AccAddress, coin sdk.Coin) bool {
		return func(addr sdk.AccAddress, coin sdk.Coin) bool {
			processed = append(processed, entry(addr, coin.String()))
			return len(processed) >= count
		}
	}
	getAll := func(addr sdk.AccAddress, coin sdk.Coin) bool {
		processed = append(processed, entry(addr, coin.String()))
		return false
	}

	expProcessed := []string{
		entry(s.addr1, "99banana"),
		entry(s.addr2, "18banana"),
		entry(s.addr2, "3cucumber"),
		entry(s.addr4, "52acorn"),
		entry(s.addr4, "12cucumber"),
		entry(s.addr4, "747eggplant"),
		entry(s.addr5, "358acorn"),
		entry(s.addr5, "101banana"),
		entry(s.addr5, "8cucumber"),
		entry(s.addr5, "5durian"),
	}

	tests := []struct {
		name        string
		process     func(sdk.AccAddress, sdk.Coin) bool
		expProc     []string
		expErr      []string
		expNotInErr []string
	}{
		{
			name:    "Get all",
			process: getAll,
			expProc: expProcessed,
			expErr: []string{
				s.addr3.String(),
				"failed to read amount of grimcoin",
				"math/big: cannot unmarshal \"grimvalue\" into a *big.Int",
				s.addr4.String(),
				"failed to read amount of badcoin",
				"math/big: cannot unmarshal \"badvalue\" into a *big.Int",
				"failed to read amount of dreadcoin",
				"math/big: cannot unmarshal \"dreadvalue\" into a *big.Int",
			},
		},
		{
			name:    "stop after 1",
			process: stopAfter(1),
			expProc: expProcessed[0:1],
		},
		{
			name:    "stop after 4 (after 1st error)",
			process: stopAfter(4),
			expProc: expProcessed[0:4],
			expErr: []string{
				s.addr3.String(),
				"failed to read amount of grimcoin",
				"math/big: cannot unmarshal \"grimvalue\" into a *big.Int",
			},
			expNotInErr: []string{
				s.addr4.String(),
				"failed to read amount of badcoin",
				"math/big: cannot unmarshal \"badvalue\" into a *big.Int",
				"failed to read amount of dreadcoin",
				"math/big: cannot unmarshal \"dreadvalue\" into a *big.Int",
			},
		},
		{
			name:    "stop after 5 (after 2nd error)",
			process: stopAfter(5),
			expProc: expProcessed[0:5],
			expErr: []string{
				s.addr3.String(),
				"failed to read amount of grimcoin",
				"math/big: cannot unmarshal \"grimvalue\" into a *big.Int",
				s.addr4.String(),
				"failed to read amount of badcoin",
				"math/big: cannot unmarshal \"badvalue\" into a *big.Int",
			},
			expNotInErr: []string{
				"failed to read amount of dreadcoin",
				"math/big: cannot unmarshal \"dreadvalue\" into a *big.Int",
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			processed = nil
			var err error
			testFunc := func() {
				err = s.keeper.IterateAllEscrow(s.sdkCtx, tc.process)
			}
			s.Require().NotPanics(testFunc, "IterateAllEscrow")
			s.AssertErrorContents(err, tc.expErr, "IterateAllEscrow error")
			if err != nil && len(tc.expNotInErr) > 0 {
				errStr := err.Error()
				for _, unexp := range tc.expNotInErr {
					s.Assert().NotContains(errStr, unexp, "IterateAllEscrow error")
				}
			}
			s.Assert().Equal(tc.expProc, processed, "IterateAllEscrow entries processed")
		})
	}
}

func (s *TestSuite) TestKeeper_GetAllAccountEscrows() {
	store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr1, "banana", s.int(99)), "setEscrowCoinAmount(addr1, 99banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr2, "banana", s.int(18)), "setEscrowCoinAmount(addr2, 18banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr2, "cucumber", s.int(3)), "setEscrowCoinAmount(addr2, 3cucumber)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "acorn", s.int(52)), "setEscrowCoinAmount(addr4, 52acorn)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "cucumber", s.int(12)), "setEscrowCoinAmount(addr4, 12cucumber)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "eggplant", s.int(747)), "setEscrowCoinAmount(addr4, 747eggplant)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "acorn", s.int(358)), "setEscrowCoinAmount(addr5, 358acorn)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "banana", s.int(101)), "setEscrowCoinAmount(addr5, 101banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "cucumber", s.int(8)), "setEscrowCoinAmount(addr5, 8cucumber)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "durian", s.int(5)), "setEscrowCoinAmount(addr5, 5durian)")
	store = nil

	expected := []*escrow.AccountEscrow{
		{Address: s.addr1.String(), Amount: s.coins("99banana")},
		{Address: s.addr2.String(), Amount: s.coins("18banana,3cucumber")},
		{Address: s.addr4.String(), Amount: s.coins("52acorn,12cucumber,747eggplant")},
		{Address: s.addr5.String(), Amount: s.coins("358acorn,101banana,8cucumber,5durian")},
	}

	s.Run("no bad entries", func() {
		escrows, err := s.keeper.GetAllAccountEscrows(s.sdkCtx)
		s.Assert().NoError(err, "GetAllAccountEscrows error")
		s.Assert().Equal(expected, escrows, "GetAllAccountEscrows escrows")
	})

	store = s.sdkCtx.KVStore(s.keeper.GetStoreKey())
	store.Set(keeper.CreateEscrowCoinKey(s.addr3, "grimcoin"), []byte("grimvalue"))
	store.Set(keeper.CreateEscrowCoinKey(s.addr4, "badcoin"), []byte("badvalue"))
	store.Set(keeper.CreateEscrowCoinKey(s.addr4, "dreadcoin"), []byte("dreadvalue"))
	store = nil

	s.Run("a few bad entries", func() {
		expInErr := []string{
			s.addr3.String(),
			"failed to read amount of grimcoin",
			"math/big: cannot unmarshal \"grimvalue\" into a *big.Int",
			s.addr4.String(),
			"failed to read amount of badcoin",
			"math/big: cannot unmarshal \"badvalue\" into a *big.Int",
			"failed to read amount of dreadcoin",
			"math/big: cannot unmarshal \"dreadvalue\" into a *big.Int",
		}

		escrows, err := s.keeper.GetAllAccountEscrows(s.sdkCtx)
		s.AssertErrorContents(err, expInErr, "GetAllAccountEscrows error")
		s.Assert().Equal(expected, escrows, "GetAllAccountEscrows escrows")
	})
}
