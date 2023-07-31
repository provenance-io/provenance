package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/hold/keeper"
)

func (s *TestSuite) TestHoldAccountBalancesInvariantHelper() {
	s.requireFundAccount(s.addr1, "99banana,3cucumber")
	s.requireFundAccount(s.addr2, "12banana")
	s.requireFundAccount(s.addr3, "1844674407370955161500hugecoin")

	tests := []struct {
		name      string
		setup     func(s *TestSuite, store sdk.KVStore)
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
			expMsg: "1 account has 5banana on hold. 1 problem detected: " +
				"account " + s.addr4.String() + " spendable balance 0banana is less than hold amount 5banana",
			expBroken: true,
		},
		{
			name: "one addr has insufficient of a denom",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr2, "banana", s.int(13))
			},
			expMsg: "1 account has 13banana on hold. 1 problem detected: " +
				"account " + s.addr2.String() + " spendable balance 12banana is less than hold amount 13banana",
			expBroken: true,
		},
		{
			name: "one addr has insufficient of two denoms",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr1, "banana", s.int(101))
				s.requireSetHoldCoinAmount(store, s.addr1, "cucumber", s.int(4))
			},
			expMsg: "1 account has 101banana,4cucumber on hold. 1 problem detected: " +
				"account " + s.addr1.String() + " spendable balance 99banana is less than hold amount 101banana",
			expBroken: true,
		},
		{
			name: "two addrs have insufficient",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetHoldCoinAmount(store, s.addr1, "banana", s.int(101))
				s.requireSetHoldCoinAmount(store, s.addr2, "banana", s.int(14))
			},
			expMsg: "2 accounts have 115banana on hold. 2 problems detected:" +
				"\n1: account " + s.addr1.String() + " spendable balance 99banana is less than hold amount 101banana" +
				"\n2: account " + s.addr2.String() + " spendable balance 12banana is less than hold amount 14banana",
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
			expMsg: "5 accounts have 111banana,3cucumber,1844674407370955161500hugecoin,5000000000stake on hold. " +
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
			expMsg: "5 accounts have 111banana,3cucumber,1844674407370955161501hugecoin,5000000000stake on hold. " +
				"1 problem detected: " +
				"account " + s.addr3.String() + " spendable balance 1844674407370955161500hugecoin " +
				"is less than hold amount 1844674407370955161501hugecoin",
			expBroken: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearHoldState()
			if tc.setup != nil {
				tc.setup(s, s.getStore())
			}
			msg, broken := keeper.HoldAccountBalancesInvariantHelper(s.sdkCtx, s.keeper)
			s.Assert().Equal(tc.expBroken, broken, "broken bool")
			s.Assert().Equal(tc.expMsg, msg, "result message")
		})
	}
}
