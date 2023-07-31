package keeper_test

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/hold"
	"github.com/provenance-io/provenance/x/hold/keeper"
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
	s.keeper = s.app.HoldKeeper
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

// assertErrorContents asserts that the provided error is as expected.
// If contains is empty, it asserts there is no error.
// Otherwise, it asseerts that the error contains each of the entries in the contains slice.
// Returns true if it's all good, false if one or more assertion failed.
func (s *TestSuite) assertErrorContents(theError error, contains []string, msgAndArgs ...interface{}) bool {
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

// assertPanicContents asserts that, if contains is empty, the provided func does not panic
// Otherwise, asserts that the func panics and that its panic message contains each of the provided strings.
func (s *TestSuite) assertPanicContents(f panicTestFunc, contains []string, msgAndArgs ...interface{}) bool {
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

// requirePanicContents asserts that, if contains is empty, the provided func does not panic
// Otherwise, asserts that the func panics and that its panic message contains each of the provided strings.
//
// If the assertion fails, the test is halted.
func (s *TestSuite) requirePanicContents(f panicTestFunc, contains []string, msgAndArgs ...interface{}) {
	s.T().Helper()
	if s.assertPanicContents(f, contains, msgAndArgs...) {
		return
	}
	s.T().FailNow()
}

// assertNotPanicsNoErrorf asserts that the code inside the provided function does not panic
// and that it does not return an error.
// Returns true if it neither panics nor errors.
func (s *TestSuite) assertNotPanicsNoErrorf(f func() error, msg string, args ...interface{}) bool {
	s.T().Helper()
	var err error
	if !s.Assert().NotPanicsf(func() { err = f() }, msg, args...) {
		return false
	}
	return s.Assert().NoErrorf(err, msg, args...)
}

// requireNotPanicsNoErrorf asserts that the code inside the provided function does not panic
// and that it does not return an error.
//
// If the assertion fails, the test is halted.
func (s *TestSuite) requireNotPanicsNoErrorf(f func() error, msg string, args ...interface{}) {
	s.T().Helper()
	if s.assertNotPanicsNoErrorf(f, msg, args...) {
		return
	}
	s.T().FailNow()
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

// getStore returns the hold state store.
func (s *TestSuite) getStore() sdk.KVStore {
	return s.sdkCtx.KVStore(s.keeper.GetStoreKey())
}

// requireSetHoldCoinAmount calls setHoldCoinAmount making sure it doesn't panic or return an error.
func (s *TestSuite) requireSetHoldCoinAmount(store sdk.KVStore, addr sdk.AccAddress, denom string, amount sdkmath.Int) {
	testFunc := func() error {
		return s.keeper.SetHoldCoinAmount(store, addr, denom, amount)
	}
	s.requireNotPanicsNoErrorf(testFunc, "setHoldCoinAmount(%s, %s%s)", s.getAddrName(addr), amount, denom)
}

// setHoldCoinAmountRaw sets a hold coin amount to the provided "amount" string.
func (s *TestSuite) setHoldCoinAmountRaw(store sdk.KVStore, addr sdk.AccAddress, denom string, amount string) {
	store.Set(keeper.CreateHoldCoinKey(addr, denom), []byte(amount))
}

// requireFundAccount calls testutil.FundAccount, making sure it doesn't panic or error.
func (s *TestSuite) requireFundAccount(addr sdk.AccAddress, coins string) {
	testFunc := func() error {
		return testutil.FundAccount(s.app.BankKeeper, s.sdkCtx, addr, s.coins(coins))
	}
	s.requireNotPanicsNoErrorf(testFunc, "FundAccount(%s, %q)", s.getAddrName(addr), coins)
}

// clearHoldState will delete all entries from the hold store.
func (s *TestSuite) clearHoldState() {
	store := s.getStore()
	var keys [][]byte

	iter := store.Iterator(nil, nil)
	defer func() {
		if iter != nil {
			iter.Close()
		}
	}()

	for ; iter.Valid(); iter.Next() {
		s.Require().NoError(iter.Error(), "iter.Error()")
		keys = append(keys, iter.Key())
	}
	err := iter.Close()
	iter = nil
	s.Require().NoError(err, "iter.Close()")

	for _, key := range keys {
		store.Delete(key)
	}
}

// stateEntryString converts the provided key and value into a "<key>"="<value>" string.
func (s *TestSuite) stateEntryString(key, value []byte) string {
	return fmt.Sprintf("%q=%q", key, value)
}

// dumpHoldState creates a string for each entry in the hold state store.
// Each entry has the format `"<key>"="<value>"`.
func (s *TestSuite) dumpHoldState() []string {
	store := s.getStore()
	var rv []string

	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		s.Require().NoError(iter.Error(), "iter.Error()")
		key := iter.Key()
		value := iter.Value()
		rv = append(rv, s.stateEntryString(key, value))
	}

	return rv
}

// prependToEach prepends the provided prefix to each of the provide lines.
func (s *TestSuite) prependToEach(prefix string, lines []string) []string {
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return lines
}

// eventsToStrings converts events to strings representing the events, one line per attribute.
func (s *TestSuite) eventsToStrings(events sdk.Events) []string {
	var rv []string
	for i, event := range events {
		rv = append(rv, s.prependToEach(fmt.Sprintf("[%d]", i), s.eventToStrings(event))...)
	}
	return rv
}

// eventToStrings converts a single event to strings, one string per attribute.
func (s *TestSuite) eventToStrings(event sdk.Event) []string {
	return s.prependToEach(event.Type, s.attrsToStrings(event.Attributes))
}

// attrsToStrings creates and returns a string for each attribute.
func (s *TestSuite) attrsToStrings(attrs []abci.EventAttribute) []string {
	rv := make([]string, len(attrs))
	for i, attr := range attrs {
		rv[i] = fmt.Sprintf("[%d]: %q = %q", i, string(attr.Key), string(attr.Value))
		if attr.Index {
			rv[i] = rv[i] + " (indexed)"
		}
	}
	return rv
}

func (s *TestSuite) TestEventsToStrings() {
	// This test is just making sure that the strings generated by eventsToStrings have
	// all the needed info in them for accurate comparisons. Tests could erroneously pass
	// if eventsToStrings isn't doing what's expected, e.g. if it were always returning an empty slice.

	addrAdd := sdk.AccAddress("address_add_event___")
	coinsAdd := s.coins("97acorn,12banana")
	eventAddT := hold.NewEventHoldAdded(addrAdd, coinsAdd)
	eventAdd, err := sdk.TypedEventToEvent(eventAddT)
	s.Require().NoError(err, "TypedEventToEvent EventHoldAdded")

	addrRem := sdk.AccAddress("address_rem_event___")
	coinsRem := s.coins("13cucumber,81dill")
	eventRemT := hold.NewEventHoldRemoved(addrRem, coinsRem)
	eventRem, err := sdk.TypedEventToEvent(eventRemT)
	s.Require().NoError(err, "TypedEventToEvent EventHoldRemoved")

	events := sdk.Events{
		eventAdd,
		eventRem,
	}

	// Set the index flag on the first attribute of the first event so we make sure that makes a difference.
	events[0].Attributes[0].Index = true

	expected := []string{
		fmt.Sprintf("[0]provenance.hold.v1.EventHoldAdded[0]: \"address\" = \"\\\"%s\\\"\" (indexed)", addrAdd.String()),
		fmt.Sprintf("[0]provenance.hold.v1.EventHoldAdded[1]: \"amount\" = \"\\\"%s\\\"\"", coinsAdd.String()),
		fmt.Sprintf("[1]provenance.hold.v1.EventHoldRemoved[0]: \"address\" = \"\\\"%s\\\"\"", addrRem.String()),
		fmt.Sprintf("[1]provenance.hold.v1.EventHoldRemoved[1]: \"amount\" = \"\\\"%s\\\"\"", coinsRem.String()),
	}

	actual := s.eventsToStrings(events)
	s.Assert().Equal(expected, actual, "events strings")
}

// assertEqualEvents asserts that the expected events equal the actual events.
// Returns success (true = they're equal, false = they're different).
func (s *TestSuite) assertEqualEvents(expected, actual sdk.Events, msgAndArgs ...interface{}) bool {
	expectedStrs := s.eventsToStrings(expected)
	actualStrs := s.eventsToStrings(actual)
	return s.Assert().Equal(expectedStrs, actualStrs, msgAndArgs...)
}

func (s *TestSuite) TestKeeper_ValidateNewHold() {
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
			expErr:    []string{"10acorn,-3boin,22corn", "hold amounts", "cannot be negative", s.addr1.String()},
		},
		{
			name:      "no spendable for one coin",
			addr:      s.addr2,
			funds:     s.coins("10acorn,5boin,100corn"),
			spendable: s.coins("10acorn,100corn"),
			expErr:    []string{"spendable balance 0boin is less than hold amount 5boin", s.addr2.String()},
		},
		{
			name:      "not enough spendable for a coin",
			addr:      s.addr3,
			funds:     s.coins("10acorn,5boin,100corn"),
			spendable: s.coins("10acorn,4boin,100corn"),
			expErr:    []string{"spendable balance 4boin is less than hold amount 5boin", s.addr3.String()},
		},
		{
			name:      "all spendable of one coin being put on hold",
			addr:      s.addr5,
			funds:     s.coins("5boin"),
			spendable: s.coins("10acorn,5boin,100corn"),
		},
		{
			name:      "all spendable being put on hold",
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
			expErr:    []string{"spendable balance 10acorn is less than hold amount 11acorn", s.addr1.String()},
		},
		{
			name:      "three coins: first insufficient",
			addr:      s.addr1,
			funds:     s.coins("11acorn,20boin,30corn"),
			spendable: s.coins("10acorn,20boin,30corn"),
			expErr:    []string{"spendable balance 10acorn is less than hold amount 11acorn", s.addr1.String()},
		},
		{
			name:      "three coins: second insufficient",
			addr:      s.addr1,
			funds:     s.coins("10acorn,21boin,30corn"),
			spendable: s.coins("10acorn,20boin,30corn"),
			expErr:    []string{"spendable balance 20boin is less than hold amount 21boin", s.addr1.String()},
		},
		{
			name:      "three coins: third insufficient",
			addr:      s.addr1,
			funds:     s.coins("10acorn,20boin,31corn"),
			spendable: s.coins("10acorn,20boin,30corn"),
			expErr:    []string{"spendable balance 30corn is less than hold amount 31corn", s.addr1.String()},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			bk := NewMockBankKeeper().WithSpendable(tc.addr, tc.spendable)
			k := s.app.HoldKeeper.WithBankKeeper(bk)

			var err error
			testFunc := func() {
				err = k.ValidateNewHold(s.sdkCtx, tc.addr, tc.funds)
			}
			s.Require().NotPanics(testFunc, "ValidateNewHold")
			s.assertErrorContents(err, tc.expErr, "ValidateNewHold")
		})
	}
}

func (s *TestSuite) TestKeeper_AddHold() {
	store := s.getStore()
	s.requireSetHoldCoinAmount(store, s.addr1, "banana", s.int(99))
	s.requireSetHoldCoinAmount(store, s.addr1, "cucumber", s.int(3))
	// max uint64 = 18,446,744,073,709,551,615
	s.requireSetHoldCoinAmount(store, s.addr2, "hugecoin", s.intStr("1844674407370955161500"))
	s.requireSetHoldCoinAmount(store, s.addr2, "mediumcoin", s.intStr("10000000000000000000"))
	s.setHoldCoinAmountRaw(store, s.addr3, "badcoin", "badvalue")
	s.setHoldCoinAmountRaw(store, s.addr3, "crudcoin", "crudvalue")
	store = nil

	makeEvents := func(addr sdk.AccAddress, coins sdk.Coins) sdk.Events {
		event, err := sdk.TypedEventToEvent(hold.NewEventHoldAdded(addr, coins))
		s.Require().NoError(err, "TypedEventToEvent EventHoldAdded(%s, %q)", s.getAddrName(addr), coins)
		return sdk.Events{event}
	}

	// Tests are ordered by address since the spendable balance depends on the previous state.
	tests := []struct {
		name      string
		addr      sdk.AccAddress
		funds     sdk.Coins
		spendBal  sdk.Coins
		expErr    []string
		finalEsc  sdk.Coins
		expEvents sdk.Events
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
			name:     "insufficent spendable: some already on hold",
			addr:     s.addr1,
			funds:    s.coins("2cucumber"),
			spendBal: s.coins("1cucumber"),
			expErr:   []string{"spendable balance 1cucumber is less than hold amount 2cucumber"},
			finalEsc: s.coins("99banana,3cucumber"),
		},
		{
			name:      "sufficient spendable: add to existing entry",
			addr:      s.addr1,
			funds:     s.coins("2banana"),
			spendBal:  s.coins("2banana,9cucumber,11durian"),
			finalEsc:  s.coins("101banana,3cucumber"),
			expEvents: makeEvents(s.addr1, s.coins("2banana")),
		},
		{
			name:      "small amount added to existing amount over max uint64",
			addr:      s.addr2,
			funds:     s.coins("99hugecoin"),
			spendBal:  s.coins("5000000000000000000000hugecoin"),
			finalEsc:  s.coins("1844674407370955161599hugecoin,10000000000000000000mediumcoin"),
			expEvents: makeEvents(s.addr2, s.coins("99hugecoin")),
		},
		{
			name:      "amount over max uint64 added to existing amount over max uint64",
			addr:      s.addr2,
			funds:     s.coins("2000000000000000000000hugecoin"),
			spendBal:  s.coins("5000000000000000000000hugecoin"),
			finalEsc:  s.coins("3844674407370955161599hugecoin,10000000000000000000mediumcoin"),
			expEvents: makeEvents(s.addr2, s.coins("2000000000000000000000hugecoin")),
		},
		{
			name:      "amount over max uint64 added to new entry",
			addr:      s.addr2,
			funds:     s.coins("18446744073709551616bigcoin"),
			spendBal:  s.coins("20000000000000000000bigcoin"),
			finalEsc:  s.coins("18446744073709551616bigcoin,3844674407370955161599hugecoin,10000000000000000000mediumcoin"),
			expEvents: makeEvents(s.addr2, s.coins("18446744073709551616bigcoin")),
		},
		{
			name:      "amount under max uint64 added to another such amount resulting in more than max uint64",
			addr:      s.addr2,
			funds:     s.coins("10000000000000000000mediumcoin"),
			spendBal:  s.coins("10000000000000000000mediumcoin"),
			finalEsc:  s.coins("18446744073709551616bigcoin,3844674407370955161599hugecoin,20000000000000000000mediumcoin"),
			expEvents: makeEvents(s.addr2, s.coins("10000000000000000000mediumcoin")),
		},
		{
			name:     "existing entry is invalid",
			addr:     s.addr3,
			funds:    s.coins("1badcoin"),
			spendBal: s.coins("1badcoin"),
			expErr: []string{
				"failed to get current badcoin hold amount",
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
			name:      "addr has bad entry but adding different denom",
			addr:      s.addr3,
			funds:     s.coins("4goodcoin"),
			spendBal:  s.coins("1badcoin,2banana,4goodcoin"),
			finalEsc:  s.coins("4goodcoin"),
			expEvents: makeEvents(s.addr3, s.coins("4goodcoin")),
		},
		{
			name:      "zero of bad denom with some of another",
			addr:      s.addr3,
			funds:     s.coins("0badcoin,8goodcoin"),
			spendBal:  s.coins("8goodcoin"),
			finalEsc:  s.coins("12goodcoin"),
			expEvents: makeEvents(s.addr3, s.coins("8goodcoin")),
		},
		{
			name:     "three denoms: two existing and bad",
			addr:     s.addr3,
			funds:    s.coins("57acorn,5badcoin,4crudcoin"),
			spendBal: s.coins("100acorn,100badcoin,100crudcoin"),
			expErr: []string{
				"failed to get current badcoin hold amount",
				"math/big: cannot unmarshal \"badvalue\" into a *big.Int",
				"failed to get current crudcoin hold amount",
				"math/big: cannot unmarshal \"crudvalue\" into a *big.Int",
			},
			finalEsc:  s.coins("57acorn,12goodcoin"),
			expEvents: makeEvents(s.addr3, s.coins("57acorn")),
		},
		{
			name:      "sufficient spendable: new denoms on hold",
			addr:      s.addr4,
			funds:     s.coins("37acorn,12banana"),
			spendBal:  s.coins("37acorn,12banana"),
			finalEsc:  s.coins("37acorn,12banana"),
			expEvents: makeEvents(s.addr4, s.coins("37acorn,12banana")),
		},
		{
			name:      "amount over max uint64 added to amount under uint64",
			addr:      s.addr4,
			funds:     s.coins("5000000000000000000000banana"),
			spendBal:  s.coins("5000000000000000000000banana"),
			finalEsc:  s.coins("37acorn,5000000000000000000012banana"),
			expEvents: makeEvents(s.addr4, s.coins("5000000000000000000000banana")),
		},
		{
			name:  "zero funds",
			addr:  s.addr5,
			funds: sdk.Coins{s.coin(0, "banana"), s.coin(0, "cucumber")},
		},
		{
			name:     "insufficient spendable: none on hold yet",
			addr:     s.addr5,
			funds:    s.coins("49apple"),
			spendBal: s.coins("48apple"),
			expErr:   []string{"spendable balance 48apple is less than hold amount 49apple"},
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

			em := sdk.NewEventManager()
			ctx := s.sdkCtx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = k.AddHold(ctx, tc.addr, tc.funds)
			}
			s.Require().NotPanics(testFunc, "AddHold")

			s.assertErrorContents(err, tc.expErr, "AddHold error")

			finalEsc, _ := k.GetHoldCoins(s.sdkCtx, tc.addr)
			s.Assert().Equal(tc.finalEsc.String(), finalEsc.String(), "final hold")

			events := em.Events()
			s.assertEqualEvents(tc.expEvents, events, "AddHold events")
		})
	}
}

func (s *TestSuite) TestKeeper_RemoveHold() {
	store := s.getStore()
	s.requireSetHoldCoinAmount(store, s.addr1, "banana", s.int(99))
	s.requireSetHoldCoinAmount(store, s.addr1, "cucumber", s.int(3))
	s.requireSetHoldCoinAmount(store, s.addr2, "banana", s.int(18))
	s.setHoldCoinAmountRaw(store, s.addr3, "badcoin", "badvalue")
	s.setHoldCoinAmountRaw(store, s.addr3, "crudcoin", "crudvalue")
	s.requireSetHoldCoinAmount(store, s.addr3, "goodcoin", s.int(2))
	// max uint64 = 18,446,744,073,709,551,615
	s.requireSetHoldCoinAmount(store, s.addr4, "hugecoin", s.intStr("1844674407370955161500"))
	s.requireSetHoldCoinAmount(store, s.addr4, "largecoin", s.intStr("1000000000000000000000"))
	s.requireSetHoldCoinAmount(store, s.addr4, "mediumcoin", s.intStr("20000000000000000000"))
	store = nil

	makeEvents := func(addr sdk.AccAddress, coins sdk.Coins) sdk.Events {
		event, err := sdk.TypedEventToEvent(hold.NewEventHoldRemoved(addr, coins))
		s.Require().NoError(err, "TypedEventToEvent EventHoldRemoved((%s, %q)", s.getAddrName(addr), coins)
		return sdk.Events{event}
	}

	// Tests are ordered by address since the spendable balance depends on the previous state.
	tests := []struct {
		name      string
		addr      sdk.AccAddress
		funds     sdk.Coins
		expErr    []string
		finalEsc  sdk.Coins
		expEvents sdk.Events
	}{
		{
			name:      "remove some of two denoms",
			addr:      s.addr1,
			funds:     s.coins("1banana,1cucumber"),
			finalEsc:  s.coins("98banana,2cucumber"),
			expEvents: makeEvents(s.addr1, s.coins("1banana,1cucumber")),
		},
		{
			name:      "remove all of two denoms",
			addr:      s.addr1,
			funds:     s.coins("98banana,2cucumber"),
			expEvents: makeEvents(s.addr1, s.coins("98banana,2cucumber")),
		},
		{
			name:  "not enough on hold",
			addr:  s.addr2,
			funds: s.coins("20banana"),
			expErr: []string{
				"cannot remove 20banana from hold",
				"account only has 18banana on hold",
			},
			finalEsc: s.coins("18banana"),
		},
		{
			name:      "only remove some of one denom",
			addr:      s.addr2,
			funds:     s.coins("10banana"),
			finalEsc:  s.coins("8banana"),
			expEvents: makeEvents(s.addr2, s.coins("10banana")),
		},
		{
			name:      "remove all of one denom",
			addr:      s.addr2,
			funds:     s.coins("8banana"),
			expEvents: makeEvents(s.addr2, s.coins("8banana")),
		},
		{
			name:  "bad existing entry",
			addr:  s.addr3,
			funds: s.coins("1badcoin"),
			expErr: []string{
				"failed to get current badcoin hold amount",
				"math/big: cannot unmarshal \"badvalue\" into a *big.Int",
			},
			finalEsc: s.coins("2goodcoin"),
		},
		{
			name:      "bad existing entry but amount of that denom is zero",
			addr:      s.addr3,
			funds:     s.coins("0badcoin,1goodcoin"),
			finalEsc:  s.coins("1goodcoin"),
			expEvents: makeEvents(s.addr3, s.coins("1goodcoin")),
		},
		{
			name:  "two bad denoms one good",
			addr:  s.addr3,
			funds: s.coins("1badcoin,2crudcoin,1goodcoin"),
			expErr: []string{
				"failed to get current badcoin hold amount",
				"math/big: cannot unmarshal \"badvalue\" into a *big.Int",
				"failed to get current crudcoin hold amount",
				"math/big: cannot unmarshal \"crudvalue\" into a *big.Int",
			},
			expEvents: makeEvents(s.addr3, s.coins("1goodcoin")),
		},
		{
			name:      "amount left on hold still greater than max uint64",
			addr:      s.addr4,
			funds:     s.coins("1hugecoin"),
			finalEsc:  s.coins("1844674407370955161499hugecoin,1000000000000000000000largecoin,20000000000000000000mediumcoin"),
			expEvents: makeEvents(s.addr4, s.coins("1hugecoin")),
		},
		{
			name:      "amount removed is greater than max uint64",
			addr:      s.addr4,
			funds:     s.coins("1844674407370955161400hugecoin"),
			finalEsc:  s.coins("99hugecoin,1000000000000000000000largecoin,20000000000000000000mediumcoin"),
			expEvents: makeEvents(s.addr4, s.coins("1844674407370955161400hugecoin")),
		},
		{
			name:      "exising amount more than max uint64 and amount removed is less with result also less",
			addr:      s.addr4,
			funds:     s.coins("10000000000000000000mediumcoin"),
			finalEsc:  s.coins("99hugecoin,1000000000000000000000largecoin,10000000000000000000mediumcoin"),
			expEvents: makeEvents(s.addr4, s.coins("10000000000000000000mediumcoin")),
		},
		{
			name:      "amount removed is more than max uint64 with result also more",
			addr:      s.addr4,
			funds:     s.coins("100000000000000000000largecoin"),
			finalEsc:  s.coins("99hugecoin,900000000000000000000largecoin,10000000000000000000mediumcoin"),
			expEvents: makeEvents(s.addr4, s.coins("100000000000000000000largecoin")),
		},
		{
			name:  "existing amount and amount to remove over max uint64 but insufficient",
			addr:  s.addr4,
			funds: s.coins("900000000000000000001largecoin"),
			expErr: []string{
				"cannot remove 900000000000000000001largecoin from hold",
				"account only has 900000000000000000000largecoin on hold",
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
			expErr: []string{"cannot remove \"-1banana\" from hold", "amounts cannot be negative"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if len(tc.expErr) > 0 {
				tc.expErr = append(tc.expErr, tc.addr.String())
			}

			em := sdk.NewEventManager()
			ctx := s.sdkCtx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = s.keeper.RemoveHold(ctx, tc.addr, tc.funds)
			}
			s.Require().NotPanics(testFunc, "RemoveHold")

			s.assertErrorContents(err, tc.expErr, "RemoveHold error")

			finalEsc, _ := s.keeper.GetHoldCoins(s.sdkCtx, tc.addr)
			s.Assert().Equal(tc.finalEsc.String(), finalEsc.String(), "final hold")

			events := em.Events()
			s.assertEqualEvents(tc.expEvents, events, "AddHold events")
		})
	}
}

func (s *TestSuite) TestKeeper_GetHoldCoin() {
	store := s.getStore()
	s.requireSetHoldCoinAmount(store, s.addr1, "banana", s.int(99))
	s.requireSetHoldCoinAmount(store, s.addr1, "cucumber", s.int(3))
	s.requireSetHoldCoinAmount(store, s.addr2, "banana", s.int(18))
	s.setHoldCoinAmountRaw(store, s.addr1, "badcoin", "badvalue")
	store = nil

	tests := []struct {
		name    string
		addr    sdk.AccAddress
		denom   string
		expCoin sdk.Coin
		expErr  []string
	}{
		{
			name:    "nothing on hold for addr",
			addr:    s.addr5,
			denom:   "nonecoin",
			expCoin: s.coin(0, "nonecoin"),
		},
		{
			name:    "addr has hold but not this denom",
			addr:    s.addr2,
			denom:   "cucumber",
			expCoin: s.coin(0, "cucumber"),
		},
		{
			name:    "addr has only this denom on hold",
			addr:    s.addr2,
			denom:   "banana",
			expCoin: s.coin(18, "banana"),
		},
		{
			name:    "addr has multiple denoms on hold but not this one",
			addr:    s.addr1,
			denom:   "nonecoin",
			expCoin: s.coin(0, "nonecoin"),
		},
		{
			name:    "addr has multiple denoms on hold this denom also on hold by other addr",
			addr:    s.addr1,
			denom:   "banana",
			expCoin: s.coin(99, "banana"),
		},
		{
			name:    "addr has multiple denoms on hold this denom is only on hold by this addr",
			addr:    s.addr1,
			denom:   "cucumber",
			expCoin: s.coin(3, "cucumber"),
		},
		{
			name:    "bad value",
			addr:    s.addr1,
			denom:   "badcoin",
			expCoin: s.coin(0, "badcoin"),
			expErr:  []string{"could not get hold coin amount for", s.addr1.String(), "math/big: cannot unmarshal \"badvalue\" into a *big.Int"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var coin sdk.Coin
			var err error
			testFunc := func() {
				coin, err = s.keeper.GetHoldCoin(s.sdkCtx, tc.addr, tc.denom)
			}
			s.Require().NotPanics(testFunc, "GetHoldCoin")
			s.assertErrorContents(err, tc.expErr, "GetHoldCoin error")
			s.Assert().Equal(tc.expCoin.String(), coin.String(), "GetHoldCoin coin")
		})
	}
}

func (s *TestSuite) TestKeeper_GetHoldCoins() {
	store := s.getStore()
	s.requireSetHoldCoinAmount(store, s.addr1, "banana", s.int(99))
	s.requireSetHoldCoinAmount(store, s.addr2, "banana", s.int(18))
	s.requireSetHoldCoinAmount(store, s.addr2, "cucumber", s.int(3))
	s.setHoldCoinAmountRaw(store, s.addr3, "grimcoin", "grimvalue")
	s.requireSetHoldCoinAmount(store, s.addr4, "acorn", s.int(52))
	s.setHoldCoinAmountRaw(store, s.addr4, "badcoin", "badvalue")
	s.requireSetHoldCoinAmount(store, s.addr4, "cucumber", s.int(12))
	s.setHoldCoinAmountRaw(store, s.addr4, "dreadcoin", "dreadvalue")
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
				coins, err = s.keeper.GetHoldCoins(s.sdkCtx, tc.addr)
			}
			s.Require().NotPanics(testFunc, "GetHoldCoins")
			s.assertErrorContents(err, tc.expErr, "GetHoldCoins error")
			s.Assert().Equal(tc.expCoins.String(), coins.String(), "GetHoldCoins coins")
		})
	}
}

func (s *TestSuite) TestKeeper_IterateHolds() {
	store := s.getStore()
	s.requireSetHoldCoinAmount(store, s.addr1, "banana", s.int(99))
	s.requireSetHoldCoinAmount(store, s.addr2, "banana", s.int(18))
	s.requireSetHoldCoinAmount(store, s.addr2, "cucumber", s.int(3))
	s.setHoldCoinAmountRaw(store, s.addr3, "grimcoin", "grimvalue")
	s.requireSetHoldCoinAmount(store, s.addr4, "acorn", s.int(52))
	s.setHoldCoinAmountRaw(store, s.addr4, "badcoin", "badvalue")
	s.requireSetHoldCoinAmount(store, s.addr4, "cucumber", s.int(12))
	s.setHoldCoinAmountRaw(store, s.addr4, "dreadcoin", "dreadvalue")
	s.requireSetHoldCoinAmount(store, s.addr4, "eggplant", s.int(747))
	s.requireSetHoldCoinAmount(store, s.addr5, "acorn", s.int(358))
	s.requireSetHoldCoinAmount(store, s.addr5, "banana", s.int(101))
	s.requireSetHoldCoinAmount(store, s.addr5, "cucumber", s.int(8))
	s.requireSetHoldCoinAmount(store, s.addr5, "durian", s.int(5))
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
				err = s.keeper.IterateHolds(s.sdkCtx, tc.addr, tc.process)
			}
			s.Require().NotPanics(testFunc, "IterateHolds")
			s.assertErrorContents(err, tc.expErr, "IterateHolds error")
			if err != nil && len(tc.expNotInErr) > 0 {
				errStr := err.Error()
				for _, unexp := range tc.expNotInErr {
					s.Assert().NotContains(errStr, unexp, "IterateHolds error")
				}
			}
			s.Assert().Equal(tc.expProc, processed, "IterateHolds entries processed")
		})
	}
}

func (s *TestSuite) TestKeeper_IterateAllHolds() {
	// Since the addresses should have been created sequentially, that's the order the store records should be in.
	// I also picked easy-to-sort coin names.
	// That means that the order they're being defined here should be the order they are in state.
	store := s.getStore()
	s.requireSetHoldCoinAmount(store, s.addr1, "banana", s.int(99))
	s.requireSetHoldCoinAmount(store, s.addr2, "banana", s.int(18))
	s.requireSetHoldCoinAmount(store, s.addr2, "cucumber", s.int(3))
	s.setHoldCoinAmountRaw(store, s.addr3, "grimcoin", "grimvalue")
	s.requireSetHoldCoinAmount(store, s.addr4, "acorn", s.int(52))
	s.setHoldCoinAmountRaw(store, s.addr4, "badcoin", "badvalue")
	s.requireSetHoldCoinAmount(store, s.addr4, "cucumber", s.int(12))
	s.setHoldCoinAmountRaw(store, s.addr4, "dreadcoin", "dreadvalue")
	s.requireSetHoldCoinAmount(store, s.addr4, "eggplant", s.int(747))
	s.requireSetHoldCoinAmount(store, s.addr5, "acorn", s.int(358))
	s.requireSetHoldCoinAmount(store, s.addr5, "banana", s.int(101))
	s.requireSetHoldCoinAmount(store, s.addr5, "cucumber", s.int(8))
	s.requireSetHoldCoinAmount(store, s.addr5, "durian", s.int(5))
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
				err = s.keeper.IterateAllHolds(s.sdkCtx, tc.process)
			}
			s.Require().NotPanics(testFunc, "IterateAllHolds")
			s.assertErrorContents(err, tc.expErr, "IterateAllHolds error")
			if err != nil && len(tc.expNotInErr) > 0 {
				errStr := err.Error()
				for _, unexp := range tc.expNotInErr {
					s.Assert().NotContains(errStr, unexp, "IterateAllHolds error")
				}
			}
			s.Assert().Equal(tc.expProc, processed, "IterateAllHolds entries processed")
		})
	}
}

func (s *TestSuite) TestKeeper_GetAllAccountHolds() {
	store := s.getStore()
	s.requireSetHoldCoinAmount(store, s.addr1, "banana", s.int(99))
	s.requireSetHoldCoinAmount(store, s.addr2, "banana", s.int(18))
	s.requireSetHoldCoinAmount(store, s.addr2, "cucumber", s.int(3))
	s.requireSetHoldCoinAmount(store, s.addr4, "acorn", s.int(52))
	s.requireSetHoldCoinAmount(store, s.addr4, "cucumber", s.int(12))
	s.requireSetHoldCoinAmount(store, s.addr4, "eggplant", s.int(747))
	s.requireSetHoldCoinAmount(store, s.addr5, "acorn", s.int(358))
	s.requireSetHoldCoinAmount(store, s.addr5, "banana", s.int(101))
	s.requireSetHoldCoinAmount(store, s.addr5, "cucumber", s.int(8))
	s.requireSetHoldCoinAmount(store, s.addr5, "durian", s.int(5))
	store = nil

	expected := []*hold.AccountHold{
		{Address: s.addr1.String(), Amount: s.coins("99banana")},
		{Address: s.addr2.String(), Amount: s.coins("18banana,3cucumber")},
		{Address: s.addr4.String(), Amount: s.coins("52acorn,12cucumber,747eggplant")},
		{Address: s.addr5.String(), Amount: s.coins("358acorn,101banana,8cucumber,5durian")},
	}

	s.Run("no bad entries", func() {
		holds, err := s.keeper.GetAllAccountHolds(s.sdkCtx)
		s.Assert().NoError(err, "GetAllAccountHolds error")
		s.Assert().Equal(expected, holds, "GetAllAccountHolds holds")
	})

	store = s.getStore()
	s.setHoldCoinAmountRaw(store, s.addr3, "grimcoin", "grimvalue")
	s.setHoldCoinAmountRaw(store, s.addr4, "badcoin", "badvalue")
	s.setHoldCoinAmountRaw(store, s.addr4, "dreadcoin", "dreadvalue")
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

		holds, err := s.keeper.GetAllAccountHolds(s.sdkCtx)
		s.assertErrorContents(err, expInErr, "GetAllAccountHolds error")
		s.Assert().Equal(expected, holds, "GetAllAccountHolds holds")
	})
}
