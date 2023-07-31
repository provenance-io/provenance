package keeper_test

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/hold"
	"github.com/provenance-io/provenance/x/hold/keeper"
)

func (s *TestSuite) TestKeeper_InitGenesis() {
	// SetupTest creates the accounts each with 1_000_000_000 of the bond denom.
	s.requireFundAccount(s.addr1, "99banana,53cactus")
	s.requireFundAccount(s.addr2, "42banana")
	addrDNE := sdk.AccAddress("addr_does_not_exist_")

	genStateWithEscrows := func(escrows ...*hold.AccountEscrow) *hold.GenesisState {
		return &hold.GenesisState{Escrows: escrows}
	}
	accEscrow := func(addr sdk.AccAddress, amount sdk.Coins) *hold.AccountEscrow {
		return &hold.AccountEscrow{
			Address: addr.String(),
			Amount:  amount,
		}
	}
	aeStateEntries := func(ae *hold.AccountEscrow) []string {
		addr, err := sdk.AccAddressFromBech32(ae.Address)
		s.Require().NoError(err, "sdk.AccAddressFromBech32(%q)", ae.Address)
		var rv []string
		var val []byte
		for _, coin := range ae.Amount {
			key := keeper.CreateEscrowCoinKey(addr, coin.Denom)
			val, err = coin.Amount.Marshal()
			s.Require().NoError(err, "%q.Amount.Marshal()", coin)
			rv = append(rv, s.stateEntryString(key, val))
		}
		return rv
	}
	expStateEntries := func(genState *hold.GenesisState) []string {
		var rv []string
		if genState != nil {
			for _, ae := range genState.Escrows {
				rv = append(rv, aeStateEntries(ae)...)
			}
			sort.Strings(rv)
		}
		return rv
	}

	tests := []struct {
		name     string
		genState *hold.GenesisState
		expPanic []string
	}{
		{
			name:     "nil gen state",
			genState: nil,
		},
		{
			name:     "default gen state",
			genState: hold.DefaultGenesisState(),
		},
		{
			name: "several escrows: all okay",
			genState: genStateWithEscrows(
				accEscrow(s.addr1, s.initBal.Add(s.coins("99banana,53cactus")...)),
				accEscrow(s.addr2, s.initBal.Add(s.coins("42banana")...)),
				accEscrow(s.addr3, s.initBal),
				accEscrow(s.addr4, s.initBal),
				accEscrow(s.addr5, s.initBal),
			),
		},
		{
			name: "several escrows: first insufficient",
			genState: genStateWithEscrows(
				accEscrow(s.addr1, s.initBal.Add(s.coins("99banana,54cactus")...)),
				accEscrow(s.addr2, s.initBal.Add(s.coins("42banana")...)),
				accEscrow(s.addr3, s.initBal),
				accEscrow(s.addr4, s.initBal),
				accEscrow(s.addr5, s.initBal),
			),
			expPanic: []string{
				"escrows[0]", s.addr1.String(),
				"spendable balance 53cactus is less than hold amount 54cactus",
			},
		},
		{
			name: "several escrows: last insufficient",
			genState: genStateWithEscrows(
				accEscrow(s.addr1, s.initBal.Add(s.coins("99banana,53cactus")...)),
				accEscrow(s.addr2, s.initBal.Add(s.coins("42banana")...)),
				accEscrow(s.addr3, s.initBal),
				accEscrow(s.addr4, s.initBal),
				accEscrow(s.addr5, s.initBal.Add(s.coins("1banana")...)),
			),
			expPanic: []string{
				"escrows[4]:", s.addr5.String(),
				"spendable balance 0banana is less than hold amount 1banana",
			},
		},
		{
			name: "unknown address",
			genState: genStateWithEscrows(
				accEscrow(s.addr1, s.coins("9banana,3cactus")),
				accEscrow(addrDNE, sdk.NewCoins(s.coin(5, s.bondDenom))),
				accEscrow(s.addr4, s.initBal),
			),
			expPanic: []string{
				"escrows[1]:", addrDNE.String(),
				"spendable balance 0" + s.bondDenom + " is less than hold amount 5" + s.bondDenom,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearEscrowState()
			expectedState := expStateEntries(tc.genState)

			em := sdk.NewEventManager()
			ctx := s.sdkCtx.WithEventManager(em)
			testFunc := func() {
				s.keeper.InitGenesis(ctx, tc.genState)
			}
			s.requirePanicContents(testFunc, tc.expPanic, "InitGenesis")

			if len(tc.expPanic) == 0 {
				actualState := s.dumpEscrowState()
				s.Assert().Equal(expectedState, actualState, "hold state store entries")
			}

			events := em.Events()
			s.Assert().Empty(events, "events emitted during InitGenesis")
		})
	}
}

func (s *TestSuite) TestKeeper_ExportGenesis() {
	genStateWithEscrows := func(escrows ...*hold.AccountEscrow) *hold.GenesisState {
		return &hold.GenesisState{Escrows: escrows}
	}
	accEscrow := func(addr sdk.AccAddress, amount string) *hold.AccountEscrow {
		return &hold.AccountEscrow{
			Address: addr.String(),
			Amount:  s.coins(amount),
		}
	}

	tests := []struct {
		name        string
		setup       func(*TestSuite, sdk.KVStore)
		expGenState *hold.GenesisState
		expPanic    []string
	}{
		{
			name:        "empty state",
			expGenState: &hold.GenesisState{},
		},
		{
			name: "one entry: good",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.requireSetEscrowCoinAmount(store, s.addr1, "banana", s.int(99))
			},
			expGenState: genStateWithEscrows(accEscrow(s.addr1, "99banana")),
		},
		{
			name: "one entry: bad",
			setup: func(s *TestSuite, store sdk.KVStore) {
				s.setEscrowCoinAmountRaw(store, s.addr1, "badcoin", "badvalue")
				s.requireSetEscrowCoinAmount(store, s.addr1, "banana", s.int(99))
			},
			expPanic: []string{
				s.addr1.String(), "failed to read amount of badcoin",
				"math/big: cannot unmarshal \"badvalue\" into a *big.Int",
			},
		},
		{
			name: "five addrs: all good",
			setup: func(suite *TestSuite, store sdk.KVStore) {
				s.requireSetEscrowCoinAmount(store, s.addr1, "banana", s.int(99))
				s.requireSetEscrowCoinAmount(store, s.addr1, "cucumber", s.int(3))
				s.requireSetEscrowCoinAmount(store, s.addr1, "durian", s.int(8))
				s.requireSetEscrowCoinAmount(store, s.addr2, "banana", s.int(12))
				s.requireSetEscrowCoinAmount(store, s.addr2, "eggplant", s.int(4))
				s.requireSetEscrowCoinAmount(store, s.addr3, "acorn", s.int(92))
				s.requireSetEscrowCoinAmount(store, s.addr3, "banana", s.int(71))
				s.requireSetEscrowCoinAmount(store, s.addr4, "banana", s.int(15))
				s.requireSetEscrowCoinAmount(store, s.addr5, "acorn", s.int(5))
				s.requireSetEscrowCoinAmount(store, s.addr5, "cabbage", s.int(157))
				s.requireSetEscrowCoinAmount(store, s.addr5, "dill", s.int(22))
				s.requireSetEscrowCoinAmount(store, s.addr5, "favabean", s.int(30))
			},
			expGenState: genStateWithEscrows(
				accEscrow(s.addr1, "99banana,3cucumber,8durian"),
				accEscrow(s.addr2, "12banana,4eggplant"),
				accEscrow(s.addr3, "92acorn,71banana"),
				accEscrow(s.addr4, "15banana"),
				accEscrow(s.addr5, "5acorn,157cabbage,22dill,30favabean"),
			),
		},
		{
			name: "five addrs: several bad",
			setup: func(suite *TestSuite, store sdk.KVStore) {
				s.requireSetEscrowCoinAmount(store, s.addr1, "banana", s.int(99))
				s.requireSetEscrowCoinAmount(store, s.addr1, "cucumber", s.int(3))
				s.requireSetEscrowCoinAmount(store, s.addr1, "durian", s.int(8))
				s.requireSetEscrowCoinAmount(store, s.addr2, "banana", s.int(12))
				s.setEscrowCoinAmountRaw(store, s.addr2, "badcoin", "badvalue")
				s.requireSetEscrowCoinAmount(store, s.addr2, "eggplant", s.int(4))
				s.requireSetEscrowCoinAmount(store, s.addr3, "acorn", s.int(92))
				s.requireSetEscrowCoinAmount(store, s.addr3, "banana", s.int(71))
				s.requireSetEscrowCoinAmount(store, s.addr4, "banana", s.int(15))
				s.setEscrowCoinAmountRaw(store, s.addr4, "crunkcoin", "crunkvalue")
				s.setEscrowCoinAmountRaw(store, s.addr4, "dumbcoin", "dumbvalue")
				s.requireSetEscrowCoinAmount(store, s.addr5, "acorn", s.int(5))
				s.requireSetEscrowCoinAmount(store, s.addr5, "cabbage", s.int(157))
				s.requireSetEscrowCoinAmount(store, s.addr5, "dill", s.int(22))
				s.requireSetEscrowCoinAmount(store, s.addr5, "favabean", s.int(30))
			},
			expPanic: []string{
				s.addr2.String(),
				"failed to read amount of badcoin",
				"math/big: cannot unmarshal \"badvalue\" into a *big.Int",
				s.addr4.String(),
				"failed to read amount of crunkcoin",
				"math/big: cannot unmarshal \"crunkvalue\" into a *big.Int",
				"failed to read amount of dumbcoin",
				"math/big: cannot unmarshal \"dumbvalue\" into a *big.Int",
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearEscrowState()
			if tc.setup != nil {
				tc.setup(s, s.getStore())
			}

			var genState *hold.GenesisState
			testFunc := func() {
				genState = s.keeper.ExportGenesis(s.sdkCtx)
			}
			s.requirePanicContents(testFunc, tc.expPanic, "ExportGenesis")
			s.Assert().Equal(tc.expGenState, genState, "exported genesis state")
		})
	}
}
