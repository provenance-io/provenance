package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/provenance-io/provenance/x/hold/keeper"
)

func (s *TestSuite) TestHoldAccountBalancesInvariantHelper() {
	s.requireFundAccount(s.addr1, "99banana,3cucumber")
	s.requireFundAccount(s.addr2, "12banana")
	s.requireFundAccount(s.addr3, "1844674407370955161500hugecoin")

	var zeroTime time.Time
	startTime := time.Unix(1700000000, 0) // chosen for the even number. It's 2023-11-14 22:13:20 (a Tuesday).
	endTime := startTime.Add(time.Duration(s.initAmount) * time.Second)

	// Turn addr4 into a vesting account: 1 stake per second.
	addr4Acc := s.app.AccountKeeper.GetAccount(s.sdkCtx, s.addr4)
	addr4AccBase, ok := addr4Acc.(*authtypes.BaseAccount)
	s.Require().True(ok, "can cast addr4 account %T to %T", addr4Acc, addr4AccBase)
	addr4Vest := vesting.NewContinuousVestingAccount(addr4AccBase, s.initBal, startTime.Unix(), endTime.Unix())
	s.app.AccountKeeper.SetAccount(s.sdkCtx, addr4Vest)

	// Turn addr5 into a vesting account too:  1 stake per second, but give it some extra funds.
	addr5Acc := s.app.AccountKeeper.GetAccount(s.sdkCtx, s.addr5)
	addr5AccBase, ok := addr5Acc.(*authtypes.BaseAccount)
	s.Require().True(ok, "can cast addr5 account %T to %T", addr5Acc, addr5AccBase)
	addr5Vest := vesting.NewContinuousVestingAccount(addr5AccBase, s.initBal, startTime.Unix(), endTime.Unix())
	s.app.AccountKeeper.SetAccount(s.sdkCtx, addr5Vest)
	s.requireFundAccount(s.addr5, "5000"+s.bondDenom)

	errIns := func(addr sdk.AccAddress, balance, onHold string) string {
		return "account " + addr.String() + " spendable balance " + balance + " is less than hold amount " + onHold
	}

	tests := []struct {
		name      string
		setup     func(s *TestSuite, store sdk.KVStore)
		time      time.Time
		expMsg    string
		expBroken bool
	}{
		{
			name:      "nothing on hold",
			expMsg:    "No accounts have funds on hold. No problems detected.",
			expBroken: false,
		},
		{
			name: "error in an entry",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr1, "banana", s.int(99))
				s.setHoldCoinAmountRaw(store, s.addr1, "badcoin", "badvalue")
			},
			expMsg: "Failed to get a record of all funds that are on hold: " +
				"failed to read amount of badcoin for account " + s.addr1.String() +
				": math/big: cannot unmarshal \"badvalue\" into a *big.Int",
			expBroken: true,
		},
		{
			name: "one addr has none of a denom",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr4, "banana", s.int(5))
			},
			expMsg:    "1 account has 5banana on hold. 1 problem detected: " + errIns(s.addr4, "0banana", "5banana"),
			expBroken: true,
		},
		{
			name: "one addr has insufficient of a denom",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr2, "banana", s.int(13))
			},
			expMsg:    "1 account has 13banana on hold. 1 problem detected: " + errIns(s.addr2, "12banana", "13banana"),
			expBroken: true,
		},
		{
			name: "one addr has insufficient of two denoms",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr1, "banana", s.int(101))
				s.requireSetHoldCoinAmount(store, s.addr1, "cucumber", s.int(4))
			},
			expMsg: "1 account has 101banana,4cucumber on hold. 1 problem detected: " +
				errIns(s.addr1, "99banana", "101banana"),
			expBroken: true,
		},
		{
			name: "two addrs have insufficient",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr1, "banana", s.int(101))
				s.requireSetHoldCoinAmount(store, s.addr2, "banana", s.int(14))
			},
			expMsg: "2 accounts have 115banana on hold. 2 problems detected:\n" +
				"1: " + errIns(s.addr1, "99banana", "101banana") + "\n" +
				"2: " + errIns(s.addr2, "12banana", "14banana"),
			expBroken: true,
		},
		{
			name: "one addr has exact amount",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr1, "cucumber", s.int(3))
			},
			expMsg:    "1 account has 3cucumber on hold. No problems detected.",
			expBroken: false,
		},
		{
			name: "one addr has more than hold",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr1, "banana", s.int(95))
			},
			expMsg:    "1 account has 95banana on hold. No problems detected.",
			expBroken: false,
		},
		{
			name: "five addrs all have everything on hold",
			setup: func(s *TestSuite, store sdk.KVStore) {
				for _, addr := range []sdk.AccAddress{s.addr1, s.addr2, s.addr3, s.addr4, s.addr5} {
					for _, coin := range s.bankKeeper.GetAllBalances(s.sdkCtx, addr) {
						s.requireSetHoldCoinAmount(store, addr, coin.Denom, coin.Amount)
					}
				}
			},
			time: endTime,
			expMsg: "5 accounts have 111banana,3cucumber,1844674407370955161500hugecoin,5000005000stake on hold. " +
				"No problems detected.",
			expBroken: false,
		},
		{
			name: "five addrs all have everything on hold one has too much though",
			setup: func(s *TestSuite, store sdk.KVStore) {
				for _, addr := range []sdk.AccAddress{s.addr1, s.addr2, s.addr3, s.addr4, s.addr5} {
					for _, coin := range s.bankKeeper.GetAllBalances(s.sdkCtx, addr) {
						s.requireSetHoldCoinAmount(store, addr, coin.Denom, coin.Amount)
					}
				}
				s.requireSetHoldCoinAmount(store, s.addr3, "hugecoin", s.intStr("1844674407370955161501"))
			},
			time: endTime,
			expMsg: "5 accounts have 111banana,3cucumber,1844674407370955161501hugecoin,5000005000stake on hold. " +
				"1 problem detected: " +
				errIns(s.addr3, "1844674407370955161500hugecoin", "1844674407370955161501hugecoin"),
			expBroken: true,
		},
		{
			name: "five addrs all have holds but none have enough funds",
			setup: func(s *TestSuite, store sdk.KVStore) {
				for _, addr := range []sdk.AccAddress{s.addr1, s.addr2, s.addr3, s.addr4, s.addr5} {
					for _, coin := range s.bankKeeper.GetAllBalances(s.sdkCtx, addr) {
						s.requireSetHoldCoinAmount(store, addr, coin.Denom, coin.Amount.Add(sdkmath.OneInt()))
					}
				}
			},
			time: endTime,
			expMsg: "5 accounts have 113banana,4cucumber,1844674407370955161501hugecoin,5000005005stake on hold. " +
				"5 problems detected:\n" +
				"1: " + errIns(s.addr1, "99banana", "100banana") + "\n" +
				"2: " + errIns(s.addr2, "12banana", "13banana") + "\n" +
				"3: " + errIns(s.addr3, "1844674407370955161500hugecoin", "1844674407370955161501hugecoin") + "\n" +
				"4: " + errIns(s.addr4, "1000000000stake", "1000000001stake") + "\n" +
				"5: " + errIns(s.addr5, "1000005000stake", "1000005001stake"),
			expBroken: true,
		},
		{
			name: "vesting account has lock on some unvested funds",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr4, s.bondDenom, s.int(5))
			},
			time:      startTime,
			expMsg:    "1 account has 5stake on hold. 1 problem detected: " + errIns(s.addr4, "0stake", "5stake"),
			expBroken: true,
		},
		{
			name: "vesting account has lock on all vested funds",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr4, s.bondDenom, s.int(5))
			},
			time:      startTime.Add(5 * time.Second),
			expMsg:    "1 account has 5stake on hold. No problems detected.",
			expBroken: false,
		},
		{
			name: "vesting account has lock on some unvested funds at time zero",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr4, s.bondDenom, s.int(5))
			},
			time:      zeroTime,
			expMsg:    "1 account has 5stake on hold. No problems detected.",
			expBroken: false,
		},
		{
			name: "vesting account has lock on extra funds",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr5, s.bondDenom, s.int(5000))
			},
			time:      startTime,
			expMsg:    "1 account has 5000stake on hold. No problems detected.",
			expBroken: false,
		},
		{
			name: "vesting account has lock on extra funds and some unvested funds",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr5, s.bondDenom, s.int(5001))
			},
			time:      startTime,
			expMsg:    "1 account has 5001stake on hold. 1 problem detected: " + errIns(s.addr5, "5000stake", "5001stake"),
			expBroken: true,
		},
		{
			name: "vesting account has lock on extra funds and all vested funds",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr5, s.bondDenom, s.int(5001))
			},
			time:      startTime.Add(1 * time.Second),
			expMsg:    "1 account has 5001stake on hold. No problems detected.",
			expBroken: false,
		},
		{
			name: "vesting account has lock on extra funds plus some at time zero",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr5, s.bondDenom, s.int(5001))
			},
			time:      zeroTime,
			expMsg:    "1 account has 5001stake on hold. No problems detected.",
			expBroken: false,
		},
		{
			name: "vesting account has lock on too much at time zero",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr5, s.bondDenom, s.int(s.initAmount+5001))
			},
			time:      zeroTime,
			expMsg:    "1 account has 1000005001stake on hold. 1 problem detected: " + errIns(s.addr5, "1000005000stake", "1000005001stake"),
			expBroken: true,
		},
		{
			name: "vesting account has lock on too much after end time",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr5, s.bondDenom, s.int(s.initAmount+5001))
			},
			time:      endTime.Add(1 * time.Second),
			expMsg:    "1 account has 1000005001stake on hold. 1 problem detected: " + errIns(s.addr5, "1000005000stake", "1000005001stake"),
			expBroken: true,
		},
		{
			name: "vesting account has lock on everything after end time",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr5, s.bondDenom, s.int(s.initAmount+5000))
			},
			time:      endTime.Add(1 * time.Second),
			expMsg:    "1 account has 1000005000stake on hold. No problems detected.",
			expBroken: false,
		},
		{
			name: "vesting account has lock on everything after time zero",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr5, s.bondDenom, s.int(s.initAmount+5000))
			},
			time:      zeroTime,
			expMsg:    "1 account has 1000005000stake on hold. No problems detected.",
			expBroken: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearHoldState()
			if tc.setup != nil {
				tc.setup(s, s.getStore())
			}

			ctx := s.sdkCtx.WithBlockTime(tc.time)
			var msg string
			var broken bool
			testFunc := func() {
				msg, broken = keeper.HoldAccountBalancesInvariantHelper(ctx, s.keeper)
			}
			s.Require().NotPanics(testFunc, "holdAccountBalancesInvariantHelper")
			s.Assert().Equal(tc.expBroken, broken, "broken bool")
			s.Assert().Equal(tc.expMsg, msg, "result message")
		})
	}
}
