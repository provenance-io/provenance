package keeper_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/provenance-io/provenance/app"
	internalcollections "github.com/provenance-io/provenance/internal/collections"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/hold"
	"github.com/provenance-io/provenance/x/hold/keeper"
)

type TestSuite struct {
	suite.Suite

	app        *app.App
	ctx        sdk.Context
	keeper     keeper.Keeper
	bankKeeper bankkeeper.Keeper

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
	s.ctx = s.app.BaseApp.NewContext(false)
	s.keeper = s.app.HoldKeeper
	s.bankKeeper = s.app.BankKeeper

	var err error
	s.bondDenom, err = s.app.StakingKeeper.BondDenom(s.ctx)
	s.Require().NoError(err, "app.StakingKeeper.BondDenom(s.ctx)")

	s.initAmount = 1_000_000_000
	s.initBal = sdk.NewCoins(sdk.NewCoin(s.bondDenom, sdkmath.NewInt(s.initAmount)))

	addrs := app.AddTestAddrsIncremental(s.app, s.ctx, 5, sdkmath.NewInt(s.initAmount))
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
	return assertions.AssertErrorContents(s.T(), theError, contains, msgAndArgs...)
}

// assertErrorValue asserts that the provided error equals the expected.
// If expected is empty, asserts that theError is nil.
// Returns true if it's all good, false if one or more assertion failed.
func (s *TestSuite) assertErrorValue(theError error, expected string, msgAndArgs ...interface{}) bool {
	return assertions.AssertErrorValue(s.T(), theError, expected, msgAndArgs...)
}

// requirePanicContents asserts that, if contains is empty, the provided func does not panic
// Otherwise, asserts that the func panics and that its panic message contains each of the provided strings.
//
// If the assertion fails, the test is halted.
func (s *TestSuite) requirePanicContents(f assertions.PanicTestFunc, contains []string, msgAndArgs ...interface{}) {
	assertions.RequirePanicContents(s.T(), f, contains, msgAndArgs...)
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
func (s *TestSuite) getStore() storetypes.KVStore {
	return s.ctx.KVStore(s.keeper.GetStoreKey())
}

// requireSetHoldCoinAmount calls setHoldCoinAmount making sure it doesn't panic or return an error.
func (s *TestSuite) requireSetHoldCoinAmount(store storetypes.KVStore, addr sdk.AccAddress, denom string, amount sdkmath.Int) {
	assertions.RequireNotPanicsNoErrorf(s.T(), func() error {
		return s.keeper.SetHoldCoinAmount(store, addr, denom, amount)
	}, "setHoldCoinAmount(%s, %s%s)", s.getAddrName(addr), amount, denom)
}

// setHoldCoinAmountRaw sets a hold coin amount to the provided "amount" string.
func (s *TestSuite) setHoldCoinAmountRaw(store storetypes.KVStore, addr sdk.AccAddress, denom string, amount string) {
	store.Set(keeper.CreateHoldCoinKey(addr, denom), []byte(amount))
}

// requireFundAccount calls testutil.FundAccount, making sure it doesn't panic or return an error.
func (s *TestSuite) requireFundAccount(addr sdk.AccAddress, coins string) {
	assertions.RequireNotPanicsNoErrorf(s.T(), func() error {
		return testutil.FundAccount(s.ctx, s.app.BankKeeper, addr, s.coins(coins))
	}, "FundAccount(%s, %q)", s.getAddrName(addr), coins)
}

// assertEqualEvents asserts that the expected events equal the actual events.
// Returns success (true = they're equal, false = they're different).
func (s *TestSuite) assertEqualEvents(expected, actual sdk.Events, msgAndArgs ...interface{}) bool {
	return assertions.AssertEqualEvents(s.T(), expected, actual, msgAndArgs...)
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

func (s *TestSuite) TestSetHoldCoinAmount() {
	stateEntry := func(addr sdk.AccAddress, denom string, amt sdkmath.Int) string {
		return s.stateEntryString(keeper.CreateHoldCoinKey(addr, denom), []byte(amt.String()))
	}
	tests := []struct {
		name       string
		setupStore func(storetypes.KVStore)
		addr       sdk.AccAddress
		denom      string
		amount     sdkmath.Int
		expErr     string
		expState   []string
	}{
		{
			name:     "empty store fresh address",
			addr:     s.addr1,
			denom:    "kitty",
			amount:   s.int(3),
			expState: []string{stateEntry(s.addr1, "kitty", s.int(3))},
		},
		{
			name: "addr has none of this denom but one of another",
			setupStore: func(store storetypes.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr1, "bag", s.int(3))
			},
			addr:   s.addr1,
			denom:  "purse",
			amount: s.int(8),
			expState: []string{
				stateEntry(s.addr1, "bag", s.int(3)),
				stateEntry(s.addr1, "purse", s.int(8)),
			},
		},
		{
			name: "another addr has the same denom",
			setupStore: func(store storetypes.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr2, "banana", s.int(88))
			},
			addr:   s.addr1,
			denom:  "banana",
			amount: s.int(99),
			expState: []string{
				stateEntry(s.addr1, "banana", s.int(99)),
				stateEntry(s.addr2, "banana", s.int(88)),
			},
		},
		{
			name: "update existing entry",
			setupStore: func(store storetypes.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr1, "eclair", s.int(3))
				s.requireSetHoldCoinAmount(store, s.addr2, "eclair", s.int(500))
			},
			addr:   s.addr1,
			denom:  "eclair",
			amount: s.int(4),
			expState: []string{
				stateEntry(s.addr1, "eclair", s.int(4)),
				stateEntry(s.addr2, "eclair", s.int(500)),
			},
		},
		{
			name: "delete existing entry",
			setupStore: func(store storetypes.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr1, "blanket", s.int(12))
				s.requireSetHoldCoinAmount(store, s.addr2, "blanket", s.int(44))
			},
			addr:   s.addr1,
			denom:  "blanket",
			amount: s.int(0),
			expState: []string{
				stateEntry(s.addr2, "blanket", s.int(44)),
			},
		},
		{
			name: "zero amount without an existing entry",
			setupStore: func(store storetypes.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr2, "blanket", s.int(44))
			},
			addr:   s.addr1,
			denom:  "blanket",
			amount: s.int(0),
			expState: []string{
				stateEntry(s.addr2, "blanket", s.int(44)),
			},
		},
		{
			name:   "empty denom string",
			addr:   s.addr1,
			denom:  "",
			amount: s.int(3),
			expErr: "cannot store hold with an empty denom for " + s.addr1.String(),
		},
		{
			name:   "negative amount string",
			addr:   s.addr1,
			denom:  "negcoin",
			amount: s.int(-3),
			expErr: "cannot store negative hold amount -3negcoin for " + s.addr1.String(),
		},
		// There's currently no way sdk.Amount.Marshall() returns an error, so that test case is omitted.
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearHoldState()
			store := s.getStore()
			if tc.setupStore != nil {
				tc.setupStore(store)
			}

			err := s.keeper.SetHoldCoinAmount(store, tc.addr, tc.denom, tc.amount)
			s.assertErrorValue(err, tc.expErr, "setHoldCoinAmount")

			state := s.dumpHoldState()
			s.Assert().Equal(tc.expState, state, "state after setHoldCoinAmount")
		})
	}
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
				err = k.ValidateNewHold(s.ctx, tc.addr, tc.funds)
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

	makeEvents := func(addr sdk.AccAddress, coins sdk.Coins, reason string) sdk.Events {
		event, err := sdk.TypedEventToEvent(hold.NewEventHoldAdded(addr, coins, reason))
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
		finalHold sdk.Coins
		expEvents sdk.Events
	}{
		{
			name:      "nil funds",
			addr:      s.addr1,
			funds:     nil,
			finalHold: s.coins("99banana,3cucumber"),
		},
		{
			name:      "empty funds",
			addr:      s.addr1,
			funds:     sdk.Coins{},
			finalHold: s.coins("99banana,3cucumber"),
		},
		{
			name:      "insufficent spendable: some already on hold",
			addr:      s.addr1,
			funds:     s.coins("2cucumber"),
			spendBal:  s.coins("1cucumber"),
			expErr:    []string{"spendable balance 1cucumber is less than hold amount 2cucumber"},
			finalHold: s.coins("99banana,3cucumber"),
		},
		{
			name:      "sufficient spendable: add to existing entry",
			addr:      s.addr1,
			funds:     s.coins("2banana"),
			spendBal:  s.coins("2banana,9cucumber,11durian"),
			finalHold: s.coins("101banana,3cucumber"),
			expEvents: makeEvents(s.addr1, s.coins("2banana"), "sufficient spendable: add to existing entry"),
		},
		{
			name:      "small amount added to existing amount over max uint64",
			addr:      s.addr2,
			funds:     s.coins("99hugecoin"),
			spendBal:  s.coins("5000000000000000000000hugecoin"),
			finalHold: s.coins("1844674407370955161599hugecoin,10000000000000000000mediumcoin"),
			expEvents: makeEvents(s.addr2, s.coins("99hugecoin"), "small amount added to existing amount over max uint64"),
		},
		{
			name:      "amount over max uint64 added to existing amount over max uint64",
			addr:      s.addr2,
			funds:     s.coins("2000000000000000000000hugecoin"),
			spendBal:  s.coins("5000000000000000000000hugecoin"),
			finalHold: s.coins("3844674407370955161599hugecoin,10000000000000000000mediumcoin"),
			expEvents: makeEvents(s.addr2, s.coins("2000000000000000000000hugecoin"), "amount over max uint64 added to existing amount over max uint64"),
		},
		{
			name:      "amount over max uint64 added to new entry",
			addr:      s.addr2,
			funds:     s.coins("18446744073709551616bigcoin"),
			spendBal:  s.coins("20000000000000000000bigcoin"),
			finalHold: s.coins("18446744073709551616bigcoin,3844674407370955161599hugecoin,10000000000000000000mediumcoin"),
			expEvents: makeEvents(s.addr2, s.coins("18446744073709551616bigcoin"), "amount over max uint64 added to new entry"),
		},
		{
			name:      "amount under max uint64 added to another such amount resulting in more than max uint64",
			addr:      s.addr2,
			funds:     s.coins("10000000000000000000mediumcoin"),
			spendBal:  s.coins("10000000000000000000mediumcoin"),
			finalHold: s.coins("18446744073709551616bigcoin,3844674407370955161599hugecoin,20000000000000000000mediumcoin"),
			expEvents: makeEvents(s.addr2, s.coins("10000000000000000000mediumcoin"), "amount under max uint64 added to another such amount resulting in more than max uint64"),
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
			finalHold: s.coins("4goodcoin"),
			expEvents: makeEvents(s.addr3, s.coins("4goodcoin"), "addr has bad entry but adding different denom"),
		},
		{
			name:      "zero of bad denom with some of another",
			addr:      s.addr3,
			funds:     s.coins("0badcoin,8goodcoin"),
			spendBal:  s.coins("8goodcoin"),
			finalHold: s.coins("12goodcoin"),
			expEvents: makeEvents(s.addr3, s.coins("8goodcoin"), "zero of bad denom with some of another"),
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
			finalHold: s.coins("57acorn,12goodcoin"),
			expEvents: makeEvents(s.addr3, s.coins("57acorn"), "three denoms: two existing and bad"),
		},
		{
			name:      "sufficient spendable: new denoms on hold",
			addr:      s.addr4,
			funds:     s.coins("37acorn,12banana"),
			spendBal:  s.coins("37acorn,12banana"),
			finalHold: s.coins("37acorn,12banana"),
			expEvents: makeEvents(s.addr4, s.coins("37acorn,12banana"), "sufficient spendable: new denoms on hold"),
		},
		{
			name:      "amount over max uint64 added to amount under uint64",
			addr:      s.addr4,
			funds:     s.coins("5000000000000000000000banana"),
			spendBal:  s.coins("5000000000000000000000banana"),
			finalHold: s.coins("37acorn,5000000000000000000012banana"),
			expEvents: makeEvents(s.addr4, s.coins("5000000000000000000000banana"), "amount over max uint64 added to amount under uint64"),
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
		{
			name:      "two zero coins plus one not",
			addr:      s.addr5,
			funds:     sdk.Coins{s.coin(1, "apple"), s.coin(0, "banana"), s.coin(0, "cucumber")},
			spendBal:  s.coins("8apple"),
			finalHold: s.coins("1apple"),
			expEvents: makeEvents(s.addr5, s.coins("1apple"), "two zero coins plus one not"),
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
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = k.AddHold(ctx, tc.addr, tc.funds, tc.name)
			}
			s.Require().NotPanics(testFunc, "AddHold")

			s.assertErrorContents(err, tc.expErr, "AddHold error")

			finalHold, _ := k.GetHoldCoins(s.ctx, tc.addr)
			s.Assert().Equal(tc.finalHold.String(), finalHold.String(), "final hold")

			events := em.Events()
			s.assertEqualEvents(tc.expEvents, events, "AddHold events")
		})
	}
}

func (s *TestSuite) TestKeeper_ReleaseHold() {
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
		event, err := sdk.TypedEventToEvent(hold.NewEventHoldReleased(addr, coins))
		s.Require().NoError(err, "TypedEventToEvent EventHoldReleased((%s, %q)", s.getAddrName(addr), coins)
		return sdk.Events{event}
	}

	// Tests are ordered by address since the spendable balance depends on the previous state.
	tests := []struct {
		name      string
		addr      sdk.AccAddress
		funds     sdk.Coins
		expErr    []string
		finalHold sdk.Coins
		expEvents sdk.Events
	}{
		{
			name:      "release some of two denoms",
			addr:      s.addr1,
			funds:     s.coins("1banana,1cucumber"),
			finalHold: s.coins("98banana,2cucumber"),
			expEvents: makeEvents(s.addr1, s.coins("1banana,1cucumber")),
		},
		{
			name:      "two zero coins one not",
			addr:      s.addr1,
			funds:     sdk.Coins{s.coin(0, "apple"), s.coin(1, "banana"), s.coin(0, "cucumber")},
			finalHold: s.coins("97banana,2cucumber"),
			expEvents: makeEvents(s.addr1, s.coins("1banana")),
		},
		{
			name:      "release all of two denoms",
			addr:      s.addr1,
			funds:     s.coins("97banana,2cucumber"),
			expEvents: makeEvents(s.addr1, s.coins("97banana,2cucumber")),
		},
		{
			name:  "not enough on hold",
			addr:  s.addr2,
			funds: s.coins("20banana"),
			expErr: []string{
				"cannot release 20banana from hold",
				"account only has 18banana on hold",
			},
			finalHold: s.coins("18banana"),
		},
		{
			name:      "only release some of one denom",
			addr:      s.addr2,
			funds:     s.coins("10banana"),
			finalHold: s.coins("8banana"),
			expEvents: makeEvents(s.addr2, s.coins("10banana")),
		},
		{
			name:      "release all of one denom",
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
			finalHold: s.coins("2goodcoin"),
		},
		{
			name:      "bad existing entry but amount of that denom is zero",
			addr:      s.addr3,
			funds:     s.coins("0badcoin,1goodcoin"),
			finalHold: s.coins("1goodcoin"),
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
			finalHold: s.coins("1844674407370955161499hugecoin,1000000000000000000000largecoin,20000000000000000000mediumcoin"),
			expEvents: makeEvents(s.addr4, s.coins("1hugecoin")),
		},
		{
			name:      "amount released is greater than max uint64",
			addr:      s.addr4,
			funds:     s.coins("1844674407370955161400hugecoin"),
			finalHold: s.coins("99hugecoin,1000000000000000000000largecoin,20000000000000000000mediumcoin"),
			expEvents: makeEvents(s.addr4, s.coins("1844674407370955161400hugecoin")),
		},
		{
			name:      "exising amount more than max uint64 and amount released is less with result also less",
			addr:      s.addr4,
			funds:     s.coins("10000000000000000000mediumcoin"),
			finalHold: s.coins("99hugecoin,1000000000000000000000largecoin,10000000000000000000mediumcoin"),
			expEvents: makeEvents(s.addr4, s.coins("10000000000000000000mediumcoin")),
		},
		{
			name:      "amount released is more than max uint64 with result also more",
			addr:      s.addr4,
			funds:     s.coins("100000000000000000000largecoin"),
			finalHold: s.coins("99hugecoin,900000000000000000000largecoin,10000000000000000000mediumcoin"),
			expEvents: makeEvents(s.addr4, s.coins("100000000000000000000largecoin")),
		},
		{
			name:  "existing amount and amount to release over max uint64 but insufficient",
			addr:  s.addr4,
			funds: s.coins("900000000000000000001largecoin"),
			expErr: []string{
				"cannot release 900000000000000000001largecoin from hold",
				"account only has 900000000000000000000largecoin on hold",
			},
			finalHold: s.coins("99hugecoin,900000000000000000000largecoin,10000000000000000000mediumcoin"),
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
			expErr: []string{"cannot release \"-1banana\" from hold", "amounts cannot be negative"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if len(tc.expErr) > 0 {
				tc.expErr = append(tc.expErr, tc.addr.String())
			}

			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = s.keeper.ReleaseHold(ctx, tc.addr, tc.funds)
			}
			s.Require().NotPanics(testFunc, "ReleaseHold")

			s.assertErrorContents(err, tc.expErr, "ReleaseHold error")

			finalHold, _ := s.keeper.GetHoldCoins(s.ctx, tc.addr)
			s.Assert().Equal(tc.finalHold.String(), finalHold.String(), "final hold")

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
				coin, err = s.keeper.GetHoldCoin(s.ctx, tc.addr, tc.denom)
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
				coins, err = s.keeper.GetHoldCoins(s.ctx, tc.addr)
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
				err = s.keeper.IterateHolds(s.ctx, tc.addr, tc.process)
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
				err = s.keeper.IterateAllHolds(s.ctx, tc.process)
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
		holds, err := s.keeper.GetAllAccountHolds(s.ctx)
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

		holds, err := s.keeper.GetAllAccountHolds(s.ctx)
		s.assertErrorContents(err, expInErr, "GetAllAccountHolds error")
		s.Assert().Equal(expected, holds, "GetAllAccountHolds holds")
	})
}

func (s *TestSuite) TestVestingAndHoldOverTime() {
	// This is a bit of a complex test that tracks a vesting account over time
	// while adding, removing, delegating, undelegating, holding, and releasing funds.
	// The output of the "setup: process steps" subtest will have a table with
	// various values at each point in time.
	// Then, in subtests, some checks are run on each step.

	// denom is the only denom we care about.
	denom := "fish"
	// amtOf gets the Int amount of the only denom we care about.
	amtOf := func(coins sdk.Coins) sdkmath.Int {
		return coins.AmountOf(denom)
	}
	// coins creates an sdk.Coins with just one coin in the provided amount with the only denom we care about.
	// if 0 is provided, an empty Coins is returned (not one with a zero coin of that denom).
	coins := func(amt int64) sdk.Coins {
		if amt == 0 {
			return sdk.Coins{}
		}
		return sdk.NewCoins(sdk.NewInt64Coin(denom, amt))
	}
	// appendJoin appends the suffix to each provided string, and joins them all together into one string.
	appendJoin := func(suffix string, strs []string) string {
		return strings.Join(strs, suffix) + suffix
	}
	// logf logs the step time followed by the provided stuff.
	logf := func(step uint32, format string, args ...interface{}) {
		s.T().Logf("%4ds: "+format, append([]interface{}{step}, args...)...)
	}

	// a stepResult contains values for a certain step.
	type stepResult struct {
		step                uint32
		balance             sdk.Coins // "B"
		delegated           sdk.Coins // "D"
		spendable           sdk.Coins // "S"
		locked              sdk.Coins // "L"
		lockedHold          sdk.Coins // "LH"
		lockedVest          sdk.Coins // "LV"
		accVesting          sdk.Coins // "UV"
		accVested           sdk.Coins // "V"
		accDelegatedVesting sdk.Coins // "DV"
		accDelegatedFree    sdk.Coins // "DF"
	}
	// logLabels are the field labels used to log the step table.
	logLabels := []string{
		"B", "D",
		"S", "L",
		"LH", "LV",
		"UV", "V",
		"DV", "DF",
	}
	// logVals returns all the values that go with the logLabels for logging.
	logVals := func(result *stepResult) []interface{} {
		vals := []sdk.Coins{
			result.balance, result.delegated,
			result.spendable, result.locked,
			result.lockedHold, result.lockedVest,
			result.accVesting, result.accVested,
			result.accDelegatedVesting, result.accDelegatedFree,
		}
		rv := make([]interface{}, len(vals))
		for i, val := range vals {
			rv[i] = amtOf(val)
		}
		return rv
	}
	// logStepData logs all the values at a given step.
	logStepData := func(step uint32, result *stepResult) {
		logf(step, "  "+appendJoin("=%4s  ", logLabels), logVals(result)...)
	}
	// checkResult makes sure that either expected is nil, or expected equals actual.
	checkResult := func(expected, actual sdk.Coins, name string) {
		if expected != nil {
			s.Assert().Equal(expected.String(), actual.String(), name)
		}
	}

	// a stepAction has funds to move for any given step.
	type stepAction struct {
		fund     int64 // positive = funds added, negative = funds released.
		delegate int64 // positive = funds delegated, negative = funds undelegated.
		hold     int64 // positive = funds put into hold, negative = funds released from hold.
	}

	// Create the vesting account (it's about time).
	addr := sdk.AccAddress("ContinuousVestingAcc")
	totalSeconds := int64(1000)
	totalDur := time.Duration(totalSeconds) * time.Second
	originalVesting := coins(totalSeconds)
	startTime := time.Unix(0, 0)
	endTime := startTime.Add(totalDur)
	ctx := s.ctx.WithBlockTime(startTime)
	baseAcc := s.app.AccountKeeper.NewAccountWithAddress(ctx, addr).(*authtypes.BaseAccount)
	cva, err := vesting.NewContinuousVestingAccount(baseAcc, originalVesting, startTime.Unix(), endTime.Unix())
	s.Require().NoError(err, "NewContinuousVestingAccount")
	s.app.AccountKeeper.SetAccount(ctx, cva)
	s.requireFundAccount(addr, originalVesting.String())

	// Create a fake "module" address that we use to send "delegated" funds to.
	modAddr := sdk.AccAddress("modAddr_____________")
	modAcc := s.app.AccountKeeper.NewAccountWithAddress(ctx, modAddr)
	s.app.AccountKeeper.SetAccount(ctx, modAcc)

	// Create another account that we can send funds to, to get them out of the main address' account.
	otherAddr := sdk.AccAddress("otherAddr___________")
	otherAcc := s.app.AccountKeeper.NewAccountWithAddress(ctx, otherAddr)
	s.app.AccountKeeper.SetAccount(ctx, otherAcc)

	// process defines both the actions to take and any values to check at any given step.
	process := map[uint32]struct {
		action stepAction
		check  stepResult
	}{
		0: {action: stepAction{},
			check: stepResult{
				balance: coins(1000), delegated: coins(0),
				spendable: coins(0), locked: coins(1000),
				lockedHold: coins(0), lockedVest: coins(1000),
				accVested: coins(0), accVesting: coins(1000),
			}},
		50: {action: stepAction{delegate: 100},
			check: stepResult{
				balance: coins(900), delegated: coins(100),
				spendable: coins(50), locked: coins(850),
				lockedHold: coins(0), lockedVest: coins(850),
				accVested: coins(50), accVesting: coins(950),
			}},
		100: {
			check: stepResult{
				balance: coins(900), delegated: coins(100),
				spendable: coins(100), locked: coins(800),
				lockedHold: coins(0), lockedVest: coins(800),
				accVested: coins(100), accVesting: coins(900),
			}},
		150: {action: stepAction{delegate: -50},
			check: stepResult{
				balance: coins(950), delegated: coins(50),
				spendable: coins(150), locked: coins(800),
				lockedHold: coins(0), lockedVest: coins(800),
			}},
		200: {action: stepAction{fund: 50},
			check: stepResult{
				balance: coins(1000), delegated: coins(50),
				spendable: coins(250), locked: coins(750),
				lockedHold: coins(0), lockedVest: coins(750),
			}},
		300: {action: stepAction{hold: 150},
			check: stepResult{
				balance: coins(1000), delegated: coins(50),
				spendable: coins(200), locked: coins(800),
				lockedHold: coins(150), lockedVest: coins(650),
				accVested: coins(300), accVesting: coins(700),
			}},
		400: {action: stepAction{hold: -100},
			check: stepResult{
				balance: coins(1000), delegated: coins(50),
				spendable: coins(400), locked: coins(600),
				lockedHold: coins(50), lockedVest: coins(550),
			}},
		500: {action: stepAction{delegate: 500}},
		650: {action: stepAction{fund: 300, delegate: 300}},
		700: {action: stepAction{hold: 200},
			check: stepResult{
				balance: coins(500), delegated: coins(850),
				spendable: coins(250), locked: coins(250),
				lockedHold: coins(250), lockedVest: coins(0),
			}},
		750: {action: stepAction{hold: -100, fund: -100},
			check: stepResult{
				balance: coins(400), delegated: coins(850),
				spendable: coins(250), locked: coins(150),
				lockedHold: coins(150), lockedVest: coins(0),
			}},
		800: {action: stepAction{delegate: -800},
			check: stepResult{
				balance: coins(1200), delegated: coins(50),
				spendable: coins(900), locked: coins(300),
				lockedHold: coins(150), lockedVest: coins(150),
				accVested: coins(800), accVesting: coins(200),
			}},
		850: {action: stepAction{delegate: 950}},
		1000: {
			check: stepResult{
				balance: coins(250), delegated: coins(1000),
				spendable: coins(100), locked: coins(150),
				lockedHold: coins(150), lockedVest: coins(0),
				accVested: coins(1000), accVesting: coins(0),
			}},
		1100: {action: stepAction{delegate: 100, hold: -100}},
		1150: {action: stepAction{delegate: -1050}},
		1200: {
			check: stepResult{
				balance: coins(1200), delegated: coins(50),
				spendable: coins(1150), locked: coins(50),
				lockedHold: coins(50), lockedVest: coins(0),
			}},
	}

	// Identify all defined step times.
	lastStep := uint32(0)
	stepsMap := make(map[uint32]bool)
	for key := range process {
		stepsMap[key] = true
		if key > lastStep {
			lastStep = key
		}
	}
	// Make sure we've got a step at 0 and for every 50 seconds.
	for i := uint32(0); i <= lastStep; i += 50 {
		stepsMap[i] = true
	}

	// Put all the step values in order.
	steps := internalcollections.Keys(stepsMap)
	sort.Slice(steps, func(i, j int) bool {
		return steps[i] < steps[j]
	})

	// Run through all the steps, take the appropriate actions, and look up and log the results.
	stepResults := make([]*stepResult, len(steps))
	s.Run("setup: process steps", func() {
		for i, step := range steps {
			reqNoPanicNoErr := func(f func() error, msg string, args ...interface{}) {
				assertions.RequireNotPanicsNoErrorf(s.T(), f, "%4ds: "+msg, append([]interface{}{step}, args...)...)
			}
			blockTime := startTime.Add(time.Duration(step) * time.Second)
			ctx = s.ctx.WithBlockTime(blockTime)

			action := process[step].action

			// Do things that might add to spendable first.
			if action.fund > 0 {
				amt := coins(action.fund)
				logf(step, "Adding funds: %s", amtOf(amt))
				reqNoPanicNoErr(func() error {
					return testutil.FundAccount(s.ctx, s.app.BankKeeper, addr, amt)
				}, "FundAccount(addr, %q)", amt)
			}
			if action.delegate < 0 {
				amt := coins(-1 * action.delegate)
				logf(step, "Undelegating: %s", amtOf(amt))
				reqNoPanicNoErr(func() error {
					return s.app.BankKeeper.UndelegateCoins(ctx, modAddr, addr, amt)
				}, "UndelegateCoins(%q)", amt)
			}
			if action.hold < 0 {
				amt := coins(-1 * action.hold)
				logf(step, "Releasing hold on: %s", amtOf(amt))
				reqNoPanicNoErr(func() error {
					return s.keeper.ReleaseHold(ctx, addr, amt)
				}, "ReleaseHold(addr, %q)", amt)
			}

			// Now do the things that might reduce spendable.
			if action.fund < 0 {
				amt := coins(-1 * action.fund)
				logf(step, "Sending funds: %s", amtOf(amt))
				reqNoPanicNoErr(func() error {
					return s.app.BankKeeper.SendCoins(ctx, addr, otherAddr, amt)
				}, "SendCoins(addr, otherAddr, %q)", amt)
			}
			if action.delegate > 0 {
				amt := coins(action.delegate)
				logf(step, "Delegating: %s", amtOf(amt))
				reqNoPanicNoErr(func() error {
					return s.app.BankKeeper.DelegateCoins(ctx, addr, modAddr, amt)
				}, "DelegateCoins(%q)", amt)
			}
			if action.hold > 0 {
				amt := coins(action.hold)
				logf(step, "Putting hold on: %s", amtOf(amt))
				reqNoPanicNoErr(func() error {
					return s.keeper.AddHold(ctx, addr, amt, fmt.Sprintf("test at %d", step))
				}, "AddHold(addr, %q)", amt)
			}

			var acc *vesting.ContinuousVestingAccount
			reqNoPanicNoErr(func() error {
				acc = s.app.AccountKeeper.GetAccount(ctx, addr).(*vesting.ContinuousVestingAccount)
				return nil
			}, "casting addr account to %T", acc)

			stepResults[i] = &stepResult{
				step:                step,
				balance:             s.app.BankKeeper.GetAllBalances(ctx, addr),
				delegated:           s.app.BankKeeper.GetAllBalances(ctx, modAddr),
				spendable:           s.app.BankKeeper.SpendableCoins(ctx, addr),
				locked:              s.app.BankKeeper.LockedCoins(ctx, addr),
				lockedHold:          s.keeper.GetLockedCoins(ctx, addr),
				lockedVest:          s.app.BankKeeper.UnvestedCoins(ctx, addr),
				accVesting:          acc.GetVestingCoins(blockTime),
				accVested:           acc.GetVestedCoins(blockTime),
				accDelegatedVesting: acc.GetDelegatedVesting(),
				accDelegatedFree:    acc.GetDelegatedFree(),
			}
			logStepData(step, stepResults[i])
		}
	})
	s.Require().False(s.T().Failed(), "Stopping early due to setup failure.")

	// Lastly, loop through all the results and make sure they're all as expected.
	for _, result := range stepResults {
		s.Run(fmt.Sprintf("%d seconds", result.step), func() {
			check := process[result.step].check
			checkResult(check.balance, result.balance, "balance")
			checkResult(check.delegated, result.delegated, "delegated")
			checkResult(check.spendable, result.spendable, "spendable")
			checkResult(check.locked, result.locked, "locked")
			checkResult(check.lockedHold, result.lockedHold, "locked on hold")
			checkResult(check.lockedVest, result.lockedVest, "locked vesting")
			checkResult(check.accVesting, result.accVesting, "vesting (account result)")
			checkResult(check.accVested, result.accVested, "vested (account result)")
			checkResult(check.accDelegatedVesting, result.accDelegatedVesting, "delegated vesting (account result)")
			checkResult(check.accDelegatedFree, result.accDelegatedFree, "delegated free (account result)")

			// A few more checks on every step to make sure things add up as expected.
			accDelegatedTotal := result.accDelegatedVesting.Add(result.accDelegatedFree...)
			s.Require().Equal(result.delegated.String(), accDelegatedTotal.String(),
				"delegated VS delegated vesting %q + delegated free %q",
				result.accDelegatedVesting, result.accDelegatedFree)
			balanceCheck := result.locked.Add(result.spendable...)
			s.Require().Equal(result.balance.String(), balanceCheck.String(),
				"balance VS locked %q + spendable %q",
				result.locked, result.spendable)
			lockedCheck := result.lockedVest.Add(result.lockedHold...)
			s.Require().Equal(result.locked.String(), lockedCheck.String(),
				"locked VS locked vesting %q + locked on hold %q",
				result.lockedVest, result.lockedHold)
		})
	}
}
