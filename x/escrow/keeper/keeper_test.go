package keeper_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/provenance-io/provenance/app"
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
			s.AssertErrorContents(err, tc.expErr, "GetEscrowCoin")
			s.Assert().Equal(tc.expCoin.String(), coin.String(), "GetEscrowCoin")
		})
	}
}

// TODO[1607]: func (s *TestSuite) TestKeeper_GetEscrowCoins()

// TODO[1607]: func (s *TestSuite) TestKeeper_IterateEscrow()

// TODO[1607]: func (s *TestSuite) TestKeeper_IterateAllEscrow()

// TODO[1607]: func (s *TestSuite) TestKeeper_GetAllAccountEscrows()
