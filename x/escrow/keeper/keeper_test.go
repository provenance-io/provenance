package keeper_test

import (
	"context"
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

	addrs := app.AddTestAddrsIncremental(s.app, s.sdkCtx, 5, sdk.NewInt(1_000_000_000))
	s.addr1 = addrs[0]
	s.addr2 = addrs[1]
	s.addr3 = addrs[2]
	s.addr4 = addrs[3]
	s.addr5 = addrs[4]

}

func (s *TestSuite) cz(coins string) sdk.Coins {
	s.T().Helper()
	rv, err := sdk.ParseCoinsNormalized(coins)
	s.Require().NoError(err, "ParseCoinsNormalized(%q)", coins)
	return rv
}

func (s *TestSuite) coin(amount int64, denom string) sdk.Coin {
	s.T().Helper()
	return sdk.Coin{
		Amount: s.int(amount),
		Denom:  denom,
	}
}

func (s *TestSuite) int(amount int64) sdkmath.Int {
	return sdkmath.NewInt(amount)
}

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

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
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
			spendable: s.cz("123rake"),
		},
		{
			name:      "empty funds",
			addr:      s.addr1,
			funds:     sdk.Coins{},
			spendable: s.cz("123bake"),
		},
		{
			name:      "two zero coins",
			addr:      s.addr1,
			funds:     sdk.Coins{s.coin(0, "acorn"), s.coin(0, "boin")},
			spendable: s.cz("123fake"),
		},
		{
			name:      "with negative coin",
			addr:      s.addr1,
			funds:     sdk.Coins{s.coin(10, "acorn"), s.coin(-3, "boin"), s.coin(22, "corn")},
			spendable: s.cz("10acorn,5boin,100corn"),
			expErr:    []string{"10acorn,-3boin,22corn", "escrow amounts", "cannot be negative", s.addr1.String()},
		},
		{
			name:      "no spendable for one coin",
			addr:      s.addr2,
			funds:     s.cz("10acorn,5boin,100corn"),
			spendable: s.cz("10acorn,100corn"),
			expErr:    []string{"spendable balance 0boin is less than escrow amount 5boin", s.addr2.String()},
		},
		{
			name:      "not enough spendable for a coin",
			addr:      s.addr3,
			funds:     s.cz("10acorn,5boin,100corn"),
			spendable: s.cz("10acorn,4boin,100corn"),
			expErr:    []string{"spendable balance 4boin is less than escrow amount 5boin", s.addr3.String()},
		},
		{
			name:      "all spendable of one coin being put in escrow",
			addr:      s.addr5,
			funds:     s.cz("5boin"),
			spendable: s.cz("10acorn,5boin,100corn"),
		},
		{
			name:      "all spendable being put in escrow",
			addr:      s.addr4,
			funds:     s.cz("10acorn,5boin,100corn"),
			spendable: s.cz("10acorn,5boin,100corn"),
		},
		{
			name:      "a zero coin that is not in spendable",
			addr:      s.addr5,
			funds:     sdk.Coins{s.coin(18, "acorn"), s.coin(0, "boin"), s.coin(55, "corn")},
			spendable: s.cz("20acorn,100corn"),
		},
		{
			name:      "three coins: all enough",
			addr:      s.addr1,
			funds:     s.cz("10acorn,20boin,30corn"),
			spendable: s.cz("10acorn,20boin,30corn"),
		},
		{
			name:      "three coins: none enough",
			addr:      s.addr1,
			funds:     s.cz("11acorn,21boin,31corn"),
			spendable: s.cz("10acorn,20boin,30corn"),
			expErr:    []string{"spendable balance 10acorn is less than escrow amount 11acorn", s.addr1.String()},
		},
		{
			name:      "three coins: first insufficient",
			addr:      s.addr1,
			funds:     s.cz("11acorn,20boin,30corn"),
			spendable: s.cz("10acorn,20boin,30corn"),
			expErr:    []string{"spendable balance 10acorn is less than escrow amount 11acorn", s.addr1.String()},
		},
		{
			name:      "three coins: second insufficient",
			addr:      s.addr1,
			funds:     s.cz("10acorn,21boin,30corn"),
			spendable: s.cz("10acorn,20boin,30corn"),
			expErr:    []string{"spendable balance 20boin is less than escrow amount 21boin", s.addr1.String()},
		},
		{
			name:      "three coins: third insufficient",
			addr:      s.addr1,
			funds:     s.cz("10acorn,20boin,31corn"),
			spendable: s.cz("10acorn,20boin,30corn"),
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

// TODO[1607]: func (s *TestSuite) TestKeeper_AddEscrow()

// TODO[1607]: func (s *TestSuite) TestKeeper_RemoveEscrow()

func (s *TestSuite) TestKeeper_GetEscrowCoin() {
	store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr1, "banana", s.int(99)), "SetEscrowCoinAmount(addr1, 99banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr1, "cucumber", s.int(3)), "SetEscrowCoinAmount(addr1, 3cucumber)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr2, "banana", s.int(18)), "SetEscrowCoinAmount(addr2, 18banana)")
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
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr1, "banana", s.int(99)), "SetEscrowCoinAmount(addr1, 99banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr2, "banana", s.int(18)), "SetEscrowCoinAmount(addr2, 18banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr2, "cucumber", s.int(3)), "SetEscrowCoinAmount(addr2, 3cucumber)")
	store.Set(keeper.CreateEscrowCoinKey(s.addr3, "grimcoin"), []byte("grimvalue"))
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "acorn", s.int(52)), "SetEscrowCoinAmount(addr4, 52acorn)")
	store.Set(keeper.CreateEscrowCoinKey(s.addr4, "badcoin"), []byte("badvalue"))
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "cucumber", s.int(12)), "SetEscrowCoinAmount(addr4, 12cucumber)")
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
			expCoins: s.cz("99banana"),
		},
		{
			name:     "addr with two denoms",
			addr:     s.addr2,
			expCoins: s.cz("18banana,3cucumber"),
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
			expCoins: s.cz("52acorn,12cucumber"),
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
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr1, "banana", s.int(99)), "SetEscrowCoinAmount(addr1, 99banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr2, "banana", s.int(18)), "SetEscrowCoinAmount(addr2, 18banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr2, "cucumber", s.int(3)), "SetEscrowCoinAmount(addr2, 3cucumber)")
	store.Set(keeper.CreateEscrowCoinKey(s.addr3, "grimcoin"), []byte("grimvalue"))
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "acorn", s.int(52)), "SetEscrowCoinAmount(addr4, 52acorn)")
	store.Set(keeper.CreateEscrowCoinKey(s.addr4, "badcoin"), []byte("badvalue"))
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "cucumber", s.int(12)), "SetEscrowCoinAmount(addr4, 12cucumber)")
	store.Set(keeper.CreateEscrowCoinKey(s.addr4, "dreadcoin"), []byte("dreadvalue"))
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "eggplant", s.int(747)), "SetEscrowCoinAmount(addr4, 747eggplant)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "acorn", s.int(358)), "SetEscrowCoinAmount(addr5, 358acorn)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "banana", s.int(101)), "SetEscrowCoinAmount(addr5, 101banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "cucumber", s.int(8)), "SetEscrowCoinAmount(addr5, 8cucumber)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "durian", s.int(5)), "SetEscrowCoinAmount(addr5, 5durian)")
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
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr1, "banana", s.int(99)), "SetEscrowCoinAmount(addr1, 99banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr2, "banana", s.int(18)), "SetEscrowCoinAmount(addr2, 18banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr2, "cucumber", s.int(3)), "SetEscrowCoinAmount(addr2, 3cucumber)")
	store.Set(keeper.CreateEscrowCoinKey(s.addr3, "grimcoin"), []byte("grimvalue"))
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "acorn", s.int(52)), "SetEscrowCoinAmount(addr4, 52acorn)")
	store.Set(keeper.CreateEscrowCoinKey(s.addr4, "badcoin"), []byte("badvalue"))
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "cucumber", s.int(12)), "SetEscrowCoinAmount(addr4, 12cucumber)")
	store.Set(keeper.CreateEscrowCoinKey(s.addr4, "dreadcoin"), []byte("dreadvalue"))
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "eggplant", s.int(747)), "SetEscrowCoinAmount(addr4, 747eggplant)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "acorn", s.int(358)), "SetEscrowCoinAmount(addr5, 358acorn)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "banana", s.int(101)), "SetEscrowCoinAmount(addr5, 101banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "cucumber", s.int(8)), "SetEscrowCoinAmount(addr5, 8cucumber)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "durian", s.int(5)), "SetEscrowCoinAmount(addr5, 5durian)")
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
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr1, "banana", s.int(99)), "SetEscrowCoinAmount(addr1, 99banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr2, "banana", s.int(18)), "SetEscrowCoinAmount(addr2, 18banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr2, "cucumber", s.int(3)), "SetEscrowCoinAmount(addr2, 3cucumber)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "acorn", s.int(52)), "SetEscrowCoinAmount(addr4, 52acorn)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "cucumber", s.int(12)), "SetEscrowCoinAmount(addr4, 12cucumber)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr4, "eggplant", s.int(747)), "SetEscrowCoinAmount(addr4, 747eggplant)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "acorn", s.int(358)), "SetEscrowCoinAmount(addr5, 358acorn)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "banana", s.int(101)), "SetEscrowCoinAmount(addr5, 101banana)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "cucumber", s.int(8)), "SetEscrowCoinAmount(addr5, 8cucumber)")
	s.Require().NoError(s.keeper.SetEscrowCoinAmount(store, s.addr5, "durian", s.int(5)), "SetEscrowCoinAmount(addr5, 5durian)")
	store = nil

	expected := []*escrow.AccountEscrow{
		{Address: s.addr1.String(), Amount: s.cz("99banana")},
		{Address: s.addr2.String(), Amount: s.cz("18banana,3cucumber")},
		{Address: s.addr4.String(), Amount: s.cz("52acorn,12cucumber,747eggplant")},
		{Address: s.addr5.String(), Amount: s.cz("358acorn,101banana,8cucumber,5durian")},
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
