package keeper_test

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/provenance-io/provenance/x/escrow"
	"github.com/provenance-io/provenance/x/escrow/keeper"
)

// clearEscrowState will delete all entries from the escrow store.
func (s *TestSuite) clearEscrowState() {
	store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
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

// dumpEscrowState creates a string for each entry in the escrow state store.
// Each entry has the format `"<key>"="<value>"`.
func (s *TestSuite) dumpEscrowState() []string {
	store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
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

// stateEntryString converts the provided key and value into a "<key>"="<value>" string.
func (s *TestSuite) stateEntryString(key, value []byte) string {
	return fmt.Sprintf("%q=%q", key, value)
}

func (s *TestSuite) TestKeeper_InitGenesis() {
	// SetupTest creates the accounts each with 1_000_000_000 of the bond denom.
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.sdkCtx, s.addr1, s.coins("99banana,53cactus")), "FundAccount(addr1)")
	s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.sdkCtx, s.addr2, s.coins("42banana")), "FundAccount(addr1)")
	addrDNE := sdk.AccAddress("addr_does_not_exist_")

	genStateWithEscrows := func(escrows ...*escrow.AccountEscrow) *escrow.GenesisState {
		return &escrow.GenesisState{Escrows: escrows}
	}
	accEscrow := func(addr sdk.AccAddress, amount sdk.Coins) *escrow.AccountEscrow {
		return &escrow.AccountEscrow{
			Address: addr.String(),
			Amount:  amount,
		}
	}
	aeStateEntries := func(ae *escrow.AccountEscrow) []string {
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
	expStateEntries := func(genState *escrow.GenesisState) []string {
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
		genState *escrow.GenesisState
		expPanic []string
	}{
		{
			name:     "nil gen state",
			genState: nil,
		},
		{
			name:     "default gen state",
			genState: escrow.DefaultGenesisState(),
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
				"spendable balance 53cactus is less than escrow amount 54cactus",
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
				"spendable balance 0banana is less than escrow amount 1banana",
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
				"spendable balance 0" + s.bondDenom + " is less than escrow amount 5" + s.bondDenom,
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
				s.Assert().Equal(expectedState, actualState, "escrow state store entries")
			}

			events := em.Events()
			s.Assert().Empty(events, "events emitted during InitGenesis")
		})
	}
}

func (s *TestSuite) TestKeeper_ExportGenesis() {
	genStateWithEscrows := func(escrows ...*escrow.AccountEscrow) *escrow.GenesisState {
		return &escrow.GenesisState{Escrows: escrows}
	}
	accEscrow := func(addr sdk.AccAddress, amount string) *escrow.AccountEscrow {
		return &escrow.AccountEscrow{
			Address: addr.String(),
			Amount:  s.coins(amount),
		}
	}

	tests := []struct {
		name        string
		setup       func(*TestSuite, sdk.KVStore)
		expGenState *escrow.GenesisState
		expPanic    []string
	}{
		{
			name:        "empty state",
			expGenState: &escrow.GenesisState{},
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
				tc.setup(s, s.sdkCtx.KVStore(s.keeper.GetStoreKey()))
			}

			var genState *escrow.GenesisState
			testFunc := func() {
				genState = s.keeper.ExportGenesis(s.sdkCtx)
			}
			s.requirePanicContents(testFunc, tc.expPanic, "ExportGenesis")
			s.Assert().Equal(tc.expGenState, genState, "exported genesis state")
		})
	}
}
