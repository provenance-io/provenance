package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/provenance-io/provenance/x/hold"
)

func (s *TestSuite) TestKeeper_GetLockedCoins() {
	addrNoHold := sdk.AccAddress("addrNoHold__________")
	addrWithHolds := sdk.AccAddress("addrWithHolds_______")
	addrWithBadHold := sdk.AccAddress("addrWithBadHold_____")

	s.requireFundAccount(addrNoHold, "100acorn,100banana,100badcoin")
	s.requireFundAccount(addrWithHolds, "100acorn,100banana,100badcoin")
	s.requireFundAccount(addrWithBadHold, "100acorn,100banana,100badcoin")

	store := s.getStore()
	s.requireSetHoldCoinAmount(store, addrWithHolds, "acorn", s.int(12))
	s.requireSetHoldCoinAmount(store, addrWithHolds, "banana", s.int(99))
	s.setHoldCoinAmountRaw(store, addrWithBadHold, "badcoin", "badvalue")
	store = nil

	tests := []struct {
		name     string
		ctx      sdk.Context
		addr     sdk.AccAddress
		expCoins sdk.Coins
		expPanic []string
	}{
		{
			name:     "no coins on hold",
			ctx:      s.sdkCtx,
			addr:     addrNoHold,
			expCoins: nil,
		},
		{
			name:     "some coins on hold",
			ctx:      s.sdkCtx,
			addr:     addrWithHolds,
			expCoins: s.coins("12acorn,99banana"),
		},
		{
			name:     "with bypass: some coins on hold",
			ctx:      hold.WithBypass(s.sdkCtx),
			addr:     addrWithHolds,
			expCoins: nil,
		},
		{
			name: "error getting hold coins",
			ctx:  s.sdkCtx,
			addr: addrWithBadHold,
			expPanic: []string{
				"failed to read amount of badcoin", addrWithBadHold.String(),
				"math/big: cannot unmarshal \"badvalue\" into a *big.Int",
			},
		},
		{
			name:     "with bypass: error getting hold coins",
			ctx:      hold.WithBypass(s.sdkCtx),
			addr:     addrWithBadHold,
			expCoins: nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var coins sdk.Coins
			testFunc := func() {
				coins = s.keeper.GetLockedCoins(tc.ctx, tc.addr)
			}
			s.assertPanicContents(testFunc, tc.expPanic, "GetLockedCoins")
			s.Assert().Equal(tc.expCoins.String(), coins.String(), "GetLockedCoins result")
		})
	}
}

func (s *TestSuite) TestBank_LockedCoins() {
	// This test makes sure that the hold module is properly wired
	// into the bank module's LockedCoins function.

	addrNoHold := sdk.AccAddress("addrNoHold__________")
	addrWithHolds := sdk.AccAddress("addrWithHolds_______")
	addrVestingAndHolds := sdk.AccAddress("addrVestingAndHolds_")
	addrWithBadHold := sdk.AccAddress("addrWithBadHold_____")

	vestingAccount := vestingtypes.NewPermanentLockedAccount(
		s.app.AccountKeeper.NewAccountWithAddress(s.sdkCtx, addrVestingAndHolds).(*authtypes.BaseAccount),
		s.coins("100vestcoin"),
	)
	s.app.AccountKeeper.SetAccount(s.sdkCtx, vestingAccount)
	s.requireFundAccount(addrNoHold, "100acorn,100banana,100badcoin")
	s.requireFundAccount(addrWithHolds, "100acorn,100banana,100badcoin")
	s.requireFundAccount(addrWithBadHold, "100acorn,100banana,100badcoin")
	s.requireFundAccount(addrVestingAndHolds, "100acorn,100banana,100badcoin")

	store := s.getStore()
	s.requireSetHoldCoinAmount(store, addrWithHolds, "acorn", s.int(12))
	s.requireSetHoldCoinAmount(store, addrWithHolds, "banana", s.int(99))
	s.setHoldCoinAmountRaw(store, addrWithBadHold, "badcoin", "badvalue")
	s.requireSetHoldCoinAmount(store, addrVestingAndHolds, "banana", s.int(8))
	store = nil

	tests := []struct {
		name     string
		ctx      sdk.Context
		addr     sdk.AccAddress
		expCoins sdk.Coins
		expPanic []string
	}{
		{
			name:     "no coins on hold",
			ctx:      s.sdkCtx,
			addr:     addrNoHold,
			expCoins: nil,
		},
		{
			name:     "some coins on hold",
			ctx:      s.sdkCtx,
			addr:     addrWithHolds,
			expCoins: s.coins("12acorn,99banana"),
		},
		{
			name:     "with bypass: some coins on hold",
			ctx:      hold.WithBypass(s.sdkCtx),
			addr:     addrWithHolds,
			expCoins: nil,
		},
		{
			name: "error getting hold coins",
			ctx:  s.sdkCtx,
			addr: addrWithBadHold,
			expPanic: []string{
				"failed to read amount of badcoin", addrWithBadHold.String(),
				"math/big: cannot unmarshal \"badvalue\" into a *big.Int",
			},
		},
		{
			name:     "with bypass: error getting hold coins",
			ctx:      hold.WithBypass(s.sdkCtx),
			addr:     addrWithBadHold,
			expCoins: nil,
		},
		{
			name:     "vesting and holds",
			ctx:      s.sdkCtx,
			addr:     addrVestingAndHolds,
			expCoins: s.coins("8banana,100vestcoin"),
		},
		{
			name:     "with bypass: vesting and holds",
			ctx:      hold.WithBypass(s.sdkCtx),
			addr:     addrVestingAndHolds,
			expCoins: s.coins("100vestcoin"),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var coins sdk.Coins
			testFunc := func() {
				coins = s.app.BankKeeper.LockedCoins(tc.ctx, tc.addr)
			}
			s.assertPanicContents(testFunc, tc.expPanic, "LockedCoins")
			s.Assert().Equal(tc.expCoins.String(), coins.String(), "LockedCoins result")
		})
	}
}

func (s *TestSuite) TestBank_SpendableCoins() {
	// This test makes sure that the hold module is properly wired
	// into the bank module's SpendableCoins function.

	addrNoHold := sdk.AccAddress("addrNoHold__________")
	addrWithHolds := sdk.AccAddress("addrWithHolds_______")
	addrVestingAndHolds := sdk.AccAddress("addrVestingAndHolds_")
	addrWithBadHold := sdk.AccAddress("addrWithBadHold_____")

	vestingAccount := vestingtypes.NewPermanentLockedAccount(
		s.app.AccountKeeper.NewAccountWithAddress(s.sdkCtx, addrVestingAndHolds).(*authtypes.BaseAccount),
		s.coins("100vestcoin"),
	)
	s.app.AccountKeeper.SetAccount(s.sdkCtx, vestingAccount)
	s.requireFundAccount(addrNoHold, "100acorn,100banana,100badcoin")
	s.requireFundAccount(addrWithHolds, "100acorn,100banana,100badcoin")
	s.requireFundAccount(addrWithBadHold, "100acorn,100banana,100badcoin")
	s.requireFundAccount(addrVestingAndHolds, "100acorn,100banana,100badcoin")

	store := s.getStore()
	s.requireSetHoldCoinAmount(store, addrWithHolds, "acorn", s.int(12))
	s.requireSetHoldCoinAmount(store, addrWithHolds, "banana", s.int(99))
	s.setHoldCoinAmountRaw(store, addrWithBadHold, "badcoin", "badvalue")
	s.requireSetHoldCoinAmount(store, addrVestingAndHolds, "banana", s.int(8))
	store = nil

	tests := []struct {
		name     string
		ctx      sdk.Context
		addr     sdk.AccAddress
		expCoins sdk.Coins
		expPanic []string
	}{
		{
			name:     "no coins on hold",
			ctx:      s.sdkCtx,
			addr:     addrNoHold,
			expCoins: s.coins("100acorn,100banana,100badcoin"),
		},
		{
			name:     "some coins on hold",
			ctx:      s.sdkCtx,
			addr:     addrWithHolds,
			expCoins: s.coins("88acorn,1banana,100badcoin"),
		},
		{
			name:     "with bypass: some coins on hold",
			ctx:      hold.WithBypass(s.sdkCtx),
			addr:     addrWithHolds,
			expCoins: s.coins("100acorn,100banana,100badcoin"),
		},
		{
			name: "error getting hold coins",
			ctx:  s.sdkCtx,
			addr: addrWithBadHold,
			expPanic: []string{
				"failed to read amount of badcoin", addrWithBadHold.String(),
				"math/big: cannot unmarshal \"badvalue\" into a *big.Int",
			},
		},
		{
			name:     "with bypass: error getting hold coins",
			ctx:      hold.WithBypass(s.sdkCtx),
			addr:     addrWithBadHold,
			expCoins: s.coins("100acorn,100banana,100badcoin"),
		},
		{
			name:     "vesting and holds",
			ctx:      s.sdkCtx,
			addr:     addrVestingAndHolds,
			expCoins: s.coins("100acorn,92banana,100badcoin"),
		},
		{
			name:     "with bypass: vesting and holds",
			ctx:      hold.WithBypass(s.sdkCtx),
			addr:     addrVestingAndHolds,
			expCoins: s.coins("100acorn,100banana,100badcoin"),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var coins sdk.Coins
			testFunc := func() {
				coins = s.app.BankKeeper.SpendableCoins(tc.ctx, tc.addr)
			}
			s.assertPanicContents(testFunc, tc.expPanic, "SpendableCoins")
			s.Assert().Equal(tc.expCoins.String(), coins.String(), "SpendableCoins result")
		})
	}
}

func (s *TestSuite) TestBank_Send() {
	// This test makes sure that the hold module is properly wired
	// into the bank module's SendCoins function.
	// It's assumed that it's then also wired into InputOutput Coins since both
	// use subUnlockedCoins which calls the LockedCoins function.

	fromAddr := sdk.AccAddress("fromAddr____________")
	toAddr := sdk.AccAddress("toAddr______________")
	s.requireFundAccount(fromAddr, "100acorn,100banana,100coolcoin")

	store := s.getStore()
	s.requireSetHoldCoinAmount(store, fromAddr, "acorn", s.int(12))
	s.requireSetHoldCoinAmount(store, fromAddr, "banana", s.int(99))
	store = nil

	s.Run("send less than balance but more than spendable", func() {
		err := s.bankKeeper.SendCoins(s.sdkCtx, fromAddr, toAddr, s.coins("5banana"))
		s.Assert().EqualError(err, "spendable balance 1banana is smaller than 5banana: insufficient funds", "SendCoins error")
		fromBal := s.bankKeeper.GetAllBalances(s.sdkCtx, fromAddr)
		s.Assert().Equal("100acorn,100banana,100coolcoin", fromBal.String(), "GetAllBalances(fromAddr)")
		toBal := s.bankKeeper.GetAllBalances(s.sdkCtx, toAddr)
		s.Assert().Equal("", toBal.String(), "GetAllBalances(toAddr)")
	})

	s.Run("with bypass: send less than balance but more than spendable", func() {
		// Note: This is a really bad idea. It will break the invariant unless the hold is also deleted.
		// It's better to remove the hold first, then do a send coins without any bypass.
		// But for this test, I want to see that the bypass is being passed on as expected.

		ctx := hold.WithBypass(s.sdkCtx)
		err := s.bankKeeper.SendCoins(ctx, fromAddr, toAddr, s.coins("5banana"))
		s.Assert().NoError(err, "SendCoins error")
		err = s.keeper.RemoveHold(s.sdkCtx, fromAddr, s.coins("4banana"))
		s.Assert().NoError(err, "RemoveHold error")
		fromBal := s.bankKeeper.GetAllBalances(s.sdkCtx, fromAddr)
		s.Assert().Equal("100acorn,95banana,100coolcoin", fromBal.String(), "GetAllBalances(fromAddr)")
		toBal := s.bankKeeper.GetAllBalances(s.sdkCtx, toAddr)
		s.Assert().Equal("5banana", toBal.String(), "GetAllBalances(toAddr)")
	})

	s.Run("send exactly spendable", func() {
		err := s.bankKeeper.SendCoins(s.sdkCtx, fromAddr, toAddr, s.coins("88acorn"))
		s.Assert().NoError(err, "SendCoins error")
		fromBal := s.bankKeeper.GetAllBalances(s.sdkCtx, fromAddr)
		s.Assert().Equal(s.coins("12acorn,95banana,100coolcoin").String(), fromBal.String(), "GetAllBalances(fromAddr)")
		toBal := s.bankKeeper.GetAllBalances(s.sdkCtx, toAddr)
		s.Assert().Equal("88acorn,5banana", toBal.String(), "GetAllBalances(toAddr)")
	})
}
