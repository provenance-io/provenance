package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/keeper"
)

func (s *TestSuite) TestKeeper_GetCommitmentAmount() {
	tests := []struct {
		name     string
		setup    func()
		marketID uint32
		addr     sdk.AccAddress
		expCoins sdk.Coins
	}{
		{
			name:     "empty state",
			setup:    nil,
			marketID: 3,
			addr:     s.addr3,
			expCoins: nil,
		},
		{
			name: "market has commitments from other addrs, not this one",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 3, s.addr1, s.coins("31apple"))
				keeper.SetCommitmentAmount(store, 3, s.addr2, s.coins("32apple"))
				keeper.SetCommitmentAmount(store, 3, s.addr4, s.coins("34apple"))
				keeper.SetCommitmentAmount(store, 3, s.addr5, s.coins("35apple"))
			},
			marketID: 3,
			addr:     s.addr3,
			expCoins: nil,
		},
		{
			name: "addr has commitments in other markets, not this one",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 1, s.addr3, s.coins("13apple"))
				keeper.SetCommitmentAmount(store, 2, s.addr3, s.coins("23apple"))
				keeper.SetCommitmentAmount(store, 4, s.addr3, s.coins("43apple"))
				keeper.SetCommitmentAmount(store, 5, s.addr3, s.coins("53apple"))
			},
			marketID: 3,
			addr:     s.addr3,
			expCoins: nil,
		},
		{
			name: "one coin committed",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 2, s.addr1, s.coins("21apple"))
				keeper.SetCommitmentAmount(store, 2, s.addr2, s.coins("22apple"))
				keeper.SetCommitmentAmount(store, 2, s.addr3, s.coins("23apple"))
				keeper.SetCommitmentAmount(store, 2, s.addr4, s.coins("24apple"))
				keeper.SetCommitmentAmount(store, 2, s.addr5, s.coins("25apple"))

				keeper.SetCommitmentAmount(store, 3, s.addr1, s.coins("31apple"))
				keeper.SetCommitmentAmount(store, 3, s.addr2, s.coins("32apple"))
				keeper.SetCommitmentAmount(store, 3, s.addr3, s.coins("33apple"))
				keeper.SetCommitmentAmount(store, 3, s.addr4, s.coins("34apple"))
				keeper.SetCommitmentAmount(store, 3, s.addr5, s.coins("35apple"))

				keeper.SetCommitmentAmount(store, 4, s.addr1, s.coins("41apple"))
				keeper.SetCommitmentAmount(store, 4, s.addr2, s.coins("42apple"))
				keeper.SetCommitmentAmount(store, 4, s.addr3, s.coins("43apple"))
				keeper.SetCommitmentAmount(store, 4, s.addr4, s.coins("44apple"))
				keeper.SetCommitmentAmount(store, 4, s.addr5, s.coins("45apple"))
			},
			marketID: 3,
			addr:     s.addr3,
			expCoins: s.coins("33apple"),
		},
		{
			name: "three coins committed",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 2, s.addr2, s.coins("22apple"))
				keeper.SetCommitmentAmount(store, 2, s.addr3, s.coins("23apple"))
				keeper.SetCommitmentAmount(store, 2, s.addr4, s.coins("24apple"))

				keeper.SetCommitmentAmount(store, 3, s.addr2, s.coins("32apple"))
				keeper.SetCommitmentAmount(store, 3, s.addr3, s.coins("33apple,133banana,233cherry"))
				keeper.SetCommitmentAmount(store, 3, s.addr4, s.coins("34apple"))

				keeper.SetCommitmentAmount(store, 4, s.addr2, s.coins("42apple"))
				keeper.SetCommitmentAmount(store, 4, s.addr3, s.coins("43apple"))
				keeper.SetCommitmentAmount(store, 4, s.addr4, s.coins("44apple"))
			},
			marketID: 3,
			addr:     s.addr3,
			expCoins: s.coins("33apple,133banana,233cherry"),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var coins sdk.Coins
			testFunc := func() {
				coins = s.k.GetCommitmentAmount(s.ctx, tc.marketID, tc.addr)
			}
			s.Require().NotPanics(testFunc, "GetCommitmentAmount(%d, %s)", tc.marketID, s.getAddrName(tc.addr))
			s.Assert().Equal(tc.expCoins.String(), coins.String(), "GetCommitmentAmount(%d, %q) result", tc.marketID, tc.addr.String())
		})
	}
}

func (s *TestSuite) TestKeeper_AddCommitment() {
	existingSetup := func() {
		store := s.getStore()
		keeper.SetCommitmentAmount(store, 1, s.addr1, s.coins("11apple"))
		keeper.SetCommitmentAmount(store, 1, s.addr2, s.coins("12apple"))
		keeper.SetCommitmentAmount(store, 1, s.addr4, s.coins("14apple"))
		keeper.SetCommitmentAmount(store, 2, s.addr1, s.coins("21apple"))
		keeper.SetCommitmentAmount(store, 2, s.addr2, s.coins("22apple"))
		keeper.SetCommitmentAmount(store, 2, s.addr4, s.coins("24apple"))
		keeper.SetCommitmentAmount(store, 3, s.addr1, s.coins("31apple"))
		keeper.SetCommitmentAmount(store, 3, s.addr2, s.coins("32apple"))
		keeper.SetCommitmentAmount(store, 3, s.addr4, s.coins("34apple"))
	}
	tests := []struct {
		name        string
		setup       func()
		holdKeeper  *MockHoldKeeper
		marketID    uint32
		addr        sdk.AccAddress
		amount      sdk.Coins
		expErr      string
		expHoldCall bool
		expEvent    bool
		expAmount   sdk.Coins
	}{
		{
			name:     "empty state: zero amount",
			marketID: 2,
			addr:     s.addr3,
			amount:   nil,
		},
		{
			name:     "empty state: negative amount",
			marketID: 2,
			addr:     s.addr3,
			amount:   sdk.Coins{sdk.Coin{Denom: "apple", Amount: sdkmath.NewInt(-23)}},
			expErr:   "cannot add negative commitment amount \"-23apple\" for " + s.addr3.String() + " in market 2",
		},
		{
			name:        "empty state: error adding hold",
			holdKeeper:  NewMockHoldKeeper().WithAddHoldResults("injected testing error"),
			marketID:    2,
			addr:        s.addr3,
			amount:      s.coins("23apple"),
			expErr:      "injected testing error",
			expHoldCall: true,
		},
		{
			name:        "empty state: new commitment",
			marketID:    2,
			addr:        s.addr3,
			amount:      s.coins("23apple"),
			expHoldCall: true,
			expEvent:    true,
			expAmount:   s.coins("23apple"),
		},
		{
			name: "almost empty state: additional commitment",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 2, s.addr3, s.coins("23apple"))
			},
			marketID:    2,
			addr:        s.addr3,
			amount:      s.coins("100apple"),
			expHoldCall: true,
			expEvent:    true,
			expAmount:   s.coins("123apple"),
		},
		{
			name:      "other existing commitments: zero amount",
			setup:     existingSetup,
			marketID:  2,
			addr:      s.addr2,
			amount:    nil,
			expAmount: s.coins("22apple"),
		},
		{
			name:      "other existing commitments: negative amount",
			setup:     existingSetup,
			marketID:  2,
			addr:      s.addr2,
			amount:    sdk.Coins{sdk.Coin{Denom: "apple", Amount: sdkmath.NewInt(-2)}},
			expErr:    "cannot add negative commitment amount \"-2apple\" for " + s.addr2.String() + " in market 2",
			expAmount: s.coins("22apple"),
		},
		{
			name:        "other existing commitments: error adding hold",
			setup:       existingSetup,
			holdKeeper:  NewMockHoldKeeper().WithAddHoldResults("injected testing error"),
			marketID:    2,
			addr:        s.addr2,
			amount:      s.coins("100apple"),
			expErr:      "injected testing error",
			expHoldCall: true,
			expAmount:   s.coins("22apple"),
		},
		{
			name:        "other existing commitments: new commitment",
			setup:       existingSetup,
			marketID:    2,
			addr:        s.addr3,
			amount:      s.coins("23apple"),
			expHoldCall: true,
			expEvent:    true,
			expAmount:   s.coins("23apple"),
		},
		{
			name:        "other existing commitments: additional commitment",
			setup:       existingSetup,
			marketID:    2,
			addr:        s.addr2,
			amount:      s.coins("100apple"),
			expHoldCall: true,
			expEvent:    true,
			expAmount:   s.coins("122apple"),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var expHoldCalls HoldCalls
			if tc.expHoldCall {
				expHoldCalls.AddHold = append(expHoldCalls.AddHold, NewAddHoldArgs(tc.addr, tc.amount, fmt.Sprintf("x/exchange: commitment to %d", tc.marketID)))
			}

			var expEvents sdk.Events
			if tc.expEvent {
				expEvents = append(expEvents, s.untypeEvent(exchange.NewEventFundsCommitted(tc.addr.String(), tc.marketID, tc.amount, tc.name)))
			}

			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			if tc.holdKeeper == nil {
				tc.holdKeeper = NewMockHoldKeeper()
			}
			kpr := s.k.WithHoldKeeper(tc.holdKeeper)
			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = kpr.AddCommitment(ctx, tc.marketID, tc.addr, tc.amount, tc.name)
			}
			s.Require().NotPanics(testFunc, "AddCommitment(%d, %s, %q)",
				tc.marketID, s.getAddrName(tc.addr), tc.amount)
			s.assertErrorValue(err, tc.expErr, "AddCommitment(%d, %s, %q) error",
				tc.marketID, s.getAddrName(tc.addr), tc.amount)

			s.assertHoldKeeperCalls(tc.holdKeeper, expHoldCalls, "AddCommitment(%d, %s, %q)",
				tc.marketID, s.getAddrName(tc.addr), tc.amount)

			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "events emitted during AddCommitment(%d, %s, %q)",
				tc.marketID, s.getAddrName(tc.addr), tc.amount)

			actAmount := s.k.GetCommitmentAmount(s.ctx, tc.marketID, tc.addr)
			s.Assert().Equal(tc.expAmount.String(), actAmount.String(), "GetCommitmentAmount(%d, %s) after AddCommitment(%d, %s, %q)",
				tc.marketID, s.getAddrName(tc.addr), tc.marketID, s.getAddrName(tc.addr), tc.amount)
		})
	}
}

func (s *TestSuite) TestKeeper_AddCommitments() {
	existingSetup := func() {
		store := s.getStore()
		keeper.SetCommitmentAmount(store, 1, s.addr1, s.coins("11apple"))
		keeper.SetCommitmentAmount(store, 1, s.addr2, s.coins("12apple"))
		keeper.SetCommitmentAmount(store, 1, s.addr4, s.coins("14apple"))
		keeper.SetCommitmentAmount(store, 2, s.addr1, s.coins("21apple"))
		keeper.SetCommitmentAmount(store, 2, s.addr2, s.coins("22apple"))
		keeper.SetCommitmentAmount(store, 2, s.addr4, s.coins("24apple"))
		keeper.SetCommitmentAmount(store, 3, s.addr1, s.coins("31apple"))
		keeper.SetCommitmentAmount(store, 3, s.addr2, s.coins("32apple"))
		keeper.SetCommitmentAmount(store, 3, s.addr4, s.coins("34apple"))
	}
	eventTag := "justsometag"
	reason := func(marketID uint32) string {
		return fmt.Sprintf("x/exchange: commitment to %d", marketID)
	}

	tests := []struct {
		name            string
		setup           func()
		holdKeeper      *MockHoldKeeper
		marketID        uint32
		toAdd           []exchange.AccountAmount
		expErr          string
		expAmounts      []exchange.AccountAmount
		expEvents       sdk.Events
		expAddHoldCalls []*AddHoldArgs
	}{
		{
			name:     "nil to add",
			setup:    existingSetup,
			marketID: 2,
			toAdd:    nil,
			expErr:   "",
			expAmounts: []exchange.AccountAmount{
				{Account: s.addr1.String(), Amount: s.coins("21apple")},
				{Account: s.addr2.String(), Amount: s.coins("22apple")},
				{Account: s.addr3.String(), Amount: nil},
				{Account: s.addr4.String(), Amount: s.coins("24apple")},
				{Account: s.addr5.String(), Amount: nil},
			},
		},
		{
			name:     "empty to add",
			setup:    existingSetup,
			marketID: 2,
			toAdd:    []exchange.AccountAmount{},
			expErr:   "",
			expAmounts: []exchange.AccountAmount{
				{Account: s.addr1.String(), Amount: s.coins("21apple")},
				{Account: s.addr2.String(), Amount: s.coins("22apple")},
				{Account: s.addr3.String(), Amount: nil},
				{Account: s.addr4.String(), Amount: s.coins("24apple")},
				{Account: s.addr5.String(), Amount: nil},
			},
		},
		{
			name:     "one to add: bad addr",
			setup:    existingSetup,
			marketID: 2,
			toAdd:    []exchange.AccountAmount{{Account: "oopsnotgonnawork", Amount: s.coins("26apple")}},
			expErr:   "invalid account \"oopsnotgonnawork\": decoding bech32 failed: invalid separator index -1",
			expAmounts: []exchange.AccountAmount{
				{Account: s.addr1.String(), Amount: s.coins("21apple")},
				{Account: s.addr2.String(), Amount: s.coins("22apple")},
				{Account: s.addr3.String(), Amount: nil},
				{Account: s.addr4.String(), Amount: s.coins("24apple")},
				{Account: s.addr5.String(), Amount: nil},
			},
		},
		{
			name:     "one to add: error adding",
			setup:    existingSetup,
			marketID: 2,
			toAdd: []exchange.AccountAmount{
				{Account: s.addr2.String(), Amount: sdk.Coins{sdk.Coin{Denom: "apple", Amount: sdkmath.NewInt(-2)}}},
			},
			expErr: "cannot add negative commitment amount \"-2apple\" for " + s.addr2.String() + " in market 2",
			expAmounts: []exchange.AccountAmount{
				{Account: s.addr1.String(), Amount: s.coins("21apple")},
				{Account: s.addr2.String(), Amount: s.coins("22apple")},
				{Account: s.addr3.String(), Amount: nil},
				{Account: s.addr4.String(), Amount: s.coins("24apple")},
				{Account: s.addr5.String(), Amount: nil},
			},
		},
		{
			name:     "one to add: okay",
			setup:    existingSetup,
			marketID: 2,
			toAdd:    []exchange.AccountAmount{{Account: s.addr2.String(), Amount: s.coins("2200apple")}},
			expAmounts: []exchange.AccountAmount{
				{Account: s.addr1.String(), Amount: s.coins("21apple")},
				{Account: s.addr2.String(), Amount: s.coins("2222apple")},
				{Account: s.addr3.String(), Amount: nil},
				{Account: s.addr4.String(), Amount: s.coins("24apple")},
				{Account: s.addr5.String(), Amount: nil},
			},
			expEvents: sdk.Events{
				s.untypeEvent(exchange.NewEventFundsCommitted(s.addr2.String(), 2, s.coins("2200apple"), eventTag)),
			},
			expAddHoldCalls: []*AddHoldArgs{NewAddHoldArgs(s.addr2, s.coins("2200apple"), reason(2))},
		},
		{
			name:       "five to add: some errors",
			setup:      existingSetup,
			holdKeeper: NewMockHoldKeeper().WithAddHoldResults("injected addr1 error"),
			marketID:   2,
			toAdd: []exchange.AccountAmount{
				{Account: s.addr1.String(), Amount: s.coins("1100apple")},                                            // fail
				{Account: s.addr2.String(), Amount: s.coins("2200apple,20banana")},                                   // okay
				{Account: s.addr3.String(), Amount: s.coins("3300apple")},                                            // okay
				{Account: s.addr4.String(), Amount: sdk.Coins{sdk.Coin{Denom: "apple", Amount: sdkmath.NewInt(-4)}}}, // fail
				{Account: s.addr5.String(), Amount: sdk.Coins{sdk.Coin{Denom: "apple", Amount: sdkmath.NewInt(0)}}},  // skip
			},
			expErr: s.joinErrs(
				"injected addr1 error",
				"cannot add negative commitment amount \"-4apple\" for "+s.addr4.String()+" in market 2",
			),
			expAmounts: []exchange.AccountAmount{
				{Account: s.addr1.String(), Amount: s.coins("21apple")},
				{Account: s.addr2.String(), Amount: s.coins("2222apple,20banana")},
				{Account: s.addr3.String(), Amount: s.coins("3300apple")},
				{Account: s.addr4.String(), Amount: s.coins("24apple")},
				{Account: s.addr5.String(), Amount: nil},
			},
			expEvents: sdk.Events{
				s.untypeEvent(exchange.NewEventFundsCommitted(s.addr2.String(), 2, s.coins("2200apple,20banana"), eventTag)),
				s.untypeEvent(exchange.NewEventFundsCommitted(s.addr3.String(), 2, s.coins("3300apple"), eventTag)),
			},
			expAddHoldCalls: []*AddHoldArgs{
				NewAddHoldArgs(s.addr1, s.coins("1100apple"), reason(2)),
				NewAddHoldArgs(s.addr2, s.coins("2200apple,20banana"), reason(2)),
				NewAddHoldArgs(s.addr3, s.coins("3300apple"), reason(2)),
			},
		},
		{
			name:     "five to add: all okay",
			setup:    existingSetup,
			marketID: 2,
			toAdd: []exchange.AccountAmount{
				{Account: s.addr1.String(), Amount: s.coins("1100apple")},
				{Account: s.addr2.String(), Amount: s.coins("2200apple,20banana")},
				{Account: s.addr3.String(), Amount: s.coins("3300apple")},
				{Account: s.addr5.String(), Amount: s.coins("5500apple")},
				{Account: s.addr5.String(), Amount: s.coins("50cherry")},
			},
			expAmounts: []exchange.AccountAmount{
				{Account: s.addr1.String(), Amount: s.coins("1121apple")},
				{Account: s.addr2.String(), Amount: s.coins("2222apple,20banana")},
				{Account: s.addr3.String(), Amount: s.coins("3300apple")},
				{Account: s.addr4.String(), Amount: s.coins("24apple")},
				{Account: s.addr5.String(), Amount: s.coins("5500apple,50cherry")},
			},
			expEvents: sdk.Events{
				s.untypeEvent(exchange.NewEventFundsCommitted(s.addr1.String(), 2, s.coins("1100apple"), eventTag)),
				s.untypeEvent(exchange.NewEventFundsCommitted(s.addr2.String(), 2, s.coins("2200apple,20banana"), eventTag)),
				s.untypeEvent(exchange.NewEventFundsCommitted(s.addr3.String(), 2, s.coins("3300apple"), eventTag)),
				s.untypeEvent(exchange.NewEventFundsCommitted(s.addr5.String(), 2, s.coins("5500apple"), eventTag)),
				s.untypeEvent(exchange.NewEventFundsCommitted(s.addr5.String(), 2, s.coins("50cherry"), eventTag)),
			},
			expAddHoldCalls: []*AddHoldArgs{
				NewAddHoldArgs(s.addr1, s.coins("1100apple"), reason(2)),
				NewAddHoldArgs(s.addr2, s.coins("2200apple,20banana"), reason(2)),
				NewAddHoldArgs(s.addr3, s.coins("3300apple"), reason(2)),
				NewAddHoldArgs(s.addr5, s.coins("5500apple"), reason(2)),
				NewAddHoldArgs(s.addr5, s.coins("50cherry"), reason(2)),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			expHoldCalls := HoldCalls{AddHold: tc.expAddHoldCalls}

			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			if tc.holdKeeper == nil {
				tc.holdKeeper = NewMockHoldKeeper()
			}

			kpr := s.k.WithHoldKeeper(tc.holdKeeper)
			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = kpr.AddCommitments(ctx, tc.marketID, tc.toAdd, eventTag)
			}
			s.Require().NotPanics(testFunc, "AddCommitments(%d)", tc.marketID)
			s.assertErrorValue(err, tc.expErr, "AddCommitments(%d) error", tc.marketID)

			s.assertHoldKeeperCalls(tc.holdKeeper, expHoldCalls, "AddCommitments(%d)", tc.marketID)

			actEvents := em.Events()
			s.assertEqualEvents(tc.expEvents, actEvents, "events emitted during AddCommitments(%d)", tc.marketID)

			var addr sdk.AccAddress
			for i, exp := range tc.expAmounts {
				addr, err = sdk.AccAddressFromBech32(exp.Account)
				if s.Assert().NoError(err, "expAmounts[%d]: AccAddressFromBech32(%q)", i, exp.Account) {
					actAmount := s.k.GetCommitmentAmount(s.ctx, tc.marketID, addr)
					s.Assert().Equal(exp.Amount.String(), actAmount.String(), "expAmounts[%d]: GetCommitmentAmount(%d, %s)",
						i, tc.marketID, s.getAddrName(addr))
				}
			}
		})
	}
}

func (s *TestSuite) TestKeeper_ReleaseCommitment() {
	tests := []struct {
		name        string
		setup       func()
		holdKeeper  *MockHoldKeeper
		marketID    uint32
		addr        sdk.AccAddress
		amount      sdk.Coins
		expErr      string
		expHoldRel  sdk.Coins
		expEventAmt sdk.Coins
		expAmount   sdk.Coins
	}{
		{
			name:     "negative amount",
			marketID: 5,
			addr:     s.addr1,
			amount:   sdk.Coins{sdk.Coin{Denom: "apple", Amount: sdkmath.NewInt(-3)}},
			expErr:   "cannot release negative commitment amount \"-3apple\" for " + s.addr1.String() + " in market 5",
		},
		{
			name:     "nothing committed",
			marketID: 3,
			addr:     s.addr1,
			amount:   nil,
			expErr:   "account " + s.addr1.String() + " does not have any funds committed to market 3",
		},
		{
			name: "amount to release more than committed: same denom",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 2, s.addr2, s.coins("10apple"))
			},
			marketID: 2,
			addr:     s.addr2,
			amount:   s.coins("11apple"),
			expErr: "commitment amount to release \"11apple\" is more than currently committed amount " +
				"\"10apple\" for " + s.addr2.String() + " in market 2",
			expAmount: s.coins("10apple"),
		},
		{
			name: "amount to release more than committed: extra denom",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 2, s.addr2, s.coins("10apple"))
			},
			marketID: 2,
			addr:     s.addr2,
			amount:   s.coins("10apple,1orange"),
			expErr: "commitment amount to release \"10apple,1orange\" is more than currently committed amount " +
				"\"10apple\" for " + s.addr2.String() + " in market 2",
			expAmount: s.coins("10apple"),
		},
		{
			name: "error releasing hold",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 2, s.addr2, s.coins("10apple"))
			},
			holdKeeper: NewMockHoldKeeper().WithReleaseHoldResults("oops, we injected an error"),
			marketID:   2,
			addr:       s.addr2,
			expErr:     "oops, we injected an error",
			expHoldRel: s.coins("10apple"),
			expAmount:  s.coins("10apple"),
		},
		{
			name: "release some of commitment",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 1, s.addr2, s.coins("12banana"))
				keeper.SetCommitmentAmount(store, 2, s.addr1, s.coins("21banana"))
				keeper.SetCommitmentAmount(store, 2, s.addr2, s.coins("10apple,22banana,88cherry"))
				keeper.SetCommitmentAmount(store, 2, s.addr3, s.coins("23banana"))
				keeper.SetCommitmentAmount(store, 3, s.addr2, s.coins("32banana"))
			},
			marketID:    2,
			addr:        s.addr2,
			amount:      s.coins("10apple,50cherry"),
			expHoldRel:  s.coins("10apple,50cherry"),
			expEventAmt: s.coins("10apple,50cherry"),
			expAmount:   s.coins("22banana,38cherry"),
		},
		{
			name: "release all of commitment",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 1, s.addr2, s.coins("12banana"))
				keeper.SetCommitmentAmount(store, 2, s.addr1, s.coins("21banana"))
				keeper.SetCommitmentAmount(store, 2, s.addr2, s.coins("10apple,22banana,88cherry"))
				keeper.SetCommitmentAmount(store, 2, s.addr3, s.coins("23banana"))
				keeper.SetCommitmentAmount(store, 3, s.addr2, s.coins("32banana"))
			},
			marketID:    2,
			addr:        s.addr2,
			amount:      nil,
			expHoldRel:  s.coins("10apple,22banana,88cherry"),
			expEventAmt: s.coins("10apple,22banana,88cherry"),
			expAmount:   nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var expHoldCalls HoldCalls
			if tc.expHoldRel != nil {
				expHoldCalls.ReleaseHold = append(expHoldCalls.ReleaseHold, NewReleaseHoldArgs(tc.addr, tc.expHoldRel))
			}

			var expEvents sdk.Events
			if tc.expEventAmt != nil {
				expEvents = append(expEvents, s.untypeEvent(exchange.NewEventFundsReleased(tc.addr.String(), tc.marketID, tc.expEventAmt, tc.name)))
			}

			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			if tc.holdKeeper == nil {
				tc.holdKeeper = NewMockHoldKeeper()
			}
			kpr := s.k.WithHoldKeeper(tc.holdKeeper)
			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = kpr.ReleaseCommitment(ctx, tc.marketID, tc.addr, tc.amount, tc.name)
			}
			s.Require().NotPanics(testFunc, "ReleaseCommitment(%d, %s, %q)",
				tc.marketID, s.getAddrName(tc.addr), tc.amount)
			s.assertErrorValue(err, tc.expErr, "ReleaseCommitment(%d, %s, %q) error",
				tc.marketID, s.getAddrName(tc.addr), tc.amount)

			s.assertHoldKeeperCalls(tc.holdKeeper, expHoldCalls, "ReleaseCommitment(%d, %s, %q)",
				tc.marketID, s.getAddrName(tc.addr), tc.amount)

			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "events emitted during ReleaseCommitment(%d, %s, %q)",
				tc.marketID, s.getAddrName(tc.addr), tc.amount)

			actAmount := s.k.GetCommitmentAmount(s.ctx, tc.marketID, tc.addr)
			s.Assert().Equal(tc.expAmount.String(), actAmount.String(), "GetCommitmentAmount(%d, %s) after ReleaseCommitment(%d, %s, %q)",
				tc.marketID, s.getAddrName(tc.addr), tc.marketID, s.getAddrName(tc.addr), tc.amount)
		})
	}
}

func (s *TestSuite) TestKeeper_ReleaseCommitments() {
	existingSetup := func() {
		store := s.getStore()
		keeper.SetCommitmentAmount(store, 1, s.addr1, s.coins("11apple"))
		keeper.SetCommitmentAmount(store, 1, s.addr2, s.coins("12apple"))
		keeper.SetCommitmentAmount(store, 1, s.addr4, s.coins("14apple"))
		keeper.SetCommitmentAmount(store, 2, s.addr1, s.coins("21apple"))
		keeper.SetCommitmentAmount(store, 2, s.addr2, s.coins("22apple"))
		keeper.SetCommitmentAmount(store, 2, s.addr4, s.coins("24apple"))
		keeper.SetCommitmentAmount(store, 3, s.addr1, s.coins("31apple"))
		keeper.SetCommitmentAmount(store, 3, s.addr2, s.coins("32apple"))
		keeper.SetCommitmentAmount(store, 3, s.addr4, s.coins("34apple"))
	}
	eventTag := "justsomeothertag"

	tests := []struct {
		name            string
		setup           func()
		holdKeeper      *MockHoldKeeper
		marketID        uint32
		toRelease       []exchange.AccountAmount
		expErr          string
		expAmounts      []exchange.AccountAmount
		expEvents       sdk.Events
		expRelHoldCalls []*ReleaseHoldArgs
	}{
		{
			name:      "nil to release",
			setup:     existingSetup,
			marketID:  2,
			toRelease: nil,
			expErr:    "",
			expAmounts: []exchange.AccountAmount{
				{Account: s.addr1.String(), Amount: s.coins("21apple")},
				{Account: s.addr2.String(), Amount: s.coins("22apple")},
				{Account: s.addr3.String(), Amount: nil},
				{Account: s.addr4.String(), Amount: s.coins("24apple")},
				{Account: s.addr5.String(), Amount: nil},
			},
		},
		{
			name:      "empty to release",
			setup:     existingSetup,
			marketID:  2,
			toRelease: []exchange.AccountAmount{},
			expErr:    "",
			expAmounts: []exchange.AccountAmount{
				{Account: s.addr1.String(), Amount: s.coins("21apple")},
				{Account: s.addr2.String(), Amount: s.coins("22apple")},
				{Account: s.addr3.String(), Amount: nil},
				{Account: s.addr4.String(), Amount: s.coins("24apple")},
				{Account: s.addr5.String(), Amount: nil},
			},
		},
		{
			name:      "one to release: bad addr",
			setup:     existingSetup,
			marketID:  2,
			toRelease: []exchange.AccountAmount{{Account: "oopsnotgonnawork", Amount: s.coins("26apple")}},
			expErr:    "invalid account \"oopsnotgonnawork\": decoding bech32 failed: invalid separator index -1",
			expAmounts: []exchange.AccountAmount{
				{Account: s.addr1.String(), Amount: s.coins("21apple")},
				{Account: s.addr2.String(), Amount: s.coins("22apple")},
				{Account: s.addr3.String(), Amount: nil},
				{Account: s.addr4.String(), Amount: s.coins("24apple")},
				{Account: s.addr5.String(), Amount: nil},
			},
		},
		{
			name:     "one to release: error releasing",
			setup:    existingSetup,
			marketID: 2,
			toRelease: []exchange.AccountAmount{
				{Account: s.addr2.String(), Amount: sdk.Coins{sdk.Coin{Denom: "apple", Amount: sdkmath.NewInt(-2)}}},
			},
			expErr: "cannot release negative commitment amount \"-2apple\" for " + s.addr2.String() + " in market 2",
			expAmounts: []exchange.AccountAmount{
				{Account: s.addr1.String(), Amount: s.coins("21apple")},
				{Account: s.addr2.String(), Amount: s.coins("22apple")},
				{Account: s.addr3.String(), Amount: nil},
				{Account: s.addr4.String(), Amount: s.coins("24apple")},
				{Account: s.addr5.String(), Amount: nil},
			},
		},
		{
			name:      "one to release: partial amount",
			setup:     existingSetup,
			marketID:  2,
			toRelease: []exchange.AccountAmount{{Account: s.addr2.String(), Amount: s.coins("3apple")}},
			expAmounts: []exchange.AccountAmount{
				{Account: s.addr1.String(), Amount: s.coins("21apple")},
				{Account: s.addr2.String(), Amount: s.coins("19apple")},
				{Account: s.addr3.String(), Amount: nil},
				{Account: s.addr4.String(), Amount: s.coins("24apple")},
				{Account: s.addr5.String(), Amount: nil},
			},
			expEvents: sdk.Events{
				s.untypeEvent(exchange.NewEventFundsReleased(s.addr2.String(), 2, s.coins("3apple"), eventTag)),
			},
			expRelHoldCalls: []*ReleaseHoldArgs{NewReleaseHoldArgs(s.addr2, s.coins("3apple"))},
		},
		{
			name:      "one to release: full amount",
			setup:     existingSetup,
			marketID:  2,
			toRelease: []exchange.AccountAmount{{Account: s.addr2.String(), Amount: nil}},
			expAmounts: []exchange.AccountAmount{
				{Account: s.addr1.String(), Amount: s.coins("21apple")},
				{Account: s.addr2.String(), Amount: nil},
				{Account: s.addr3.String(), Amount: nil},
				{Account: s.addr4.String(), Amount: s.coins("24apple")},
				{Account: s.addr5.String(), Amount: nil},
			},
			expEvents: sdk.Events{
				s.untypeEvent(exchange.NewEventFundsReleased(s.addr2.String(), 2, s.coins("22apple"), eventTag)),
			},
			expRelHoldCalls: []*ReleaseHoldArgs{NewReleaseHoldArgs(s.addr2, s.coins("22apple"))},
		},
		{
			name:       "five to release: some errors",
			setup:      existingSetup,
			holdKeeper: NewMockHoldKeeper().WithReleaseHoldResults("injected addr1 error"),
			marketID:   2,
			toRelease: []exchange.AccountAmount{
				{Account: s.addr1.String(), Amount: s.coins("1apple")},          // fail
				{Account: s.addr2.String(), Amount: s.coins("2apple,20banana")}, // fail
				{Account: s.addr3.String(), Amount: s.coins("3apple")},          // fail
				{Account: s.addr4.String(), Amount: nil},                        // okay
				{Account: s.addr5.String(), Amount: sdk.Coins{sdk.Coin{Denom: "apple", Amount: sdkmath.NewInt(-4)}}},
			},
			expErr: s.joinErrs(
				"injected addr1 error",
				"commitment amount to release \"2apple,20banana\" is more than currently committed amount "+
					"\"22apple\" for "+s.addr2.String()+" in market 2",
				"account "+s.addr3.String()+" does not have any funds committed to market 2",
				"cannot release negative commitment amount \"-4apple\" for "+s.addr5.String()+" in market 2",
			),
			expAmounts: []exchange.AccountAmount{
				{Account: s.addr1.String(), Amount: s.coins("21apple")},
				{Account: s.addr2.String(), Amount: s.coins("22apple")},
				{Account: s.addr3.String(), Amount: nil},
				{Account: s.addr4.String(), Amount: nil},
				{Account: s.addr5.String(), Amount: nil},
			},
			expEvents: sdk.Events{
				s.untypeEvent(exchange.NewEventFundsReleased(s.addr4.String(), 2, s.coins("24apple"), eventTag)),
			},
			expRelHoldCalls: []*ReleaseHoldArgs{
				NewReleaseHoldArgs(s.addr1, s.coins("1apple")),
				NewReleaseHoldArgs(s.addr4, s.coins("24apple")),
			},
		},
		{
			name:     "four to add: all okay",
			setup:    existingSetup,
			marketID: 2,
			toRelease: []exchange.AccountAmount{
				{Account: s.addr1.String(), Amount: nil},
				{Account: s.addr2.String(), Amount: s.coins("22apple")},
				{Account: s.addr4.String(), Amount: s.coins("5apple")},
				{Account: s.addr4.String(), Amount: s.coins("6apple")},
			},
			expAmounts: []exchange.AccountAmount{
				{Account: s.addr1.String(), Amount: nil},
				{Account: s.addr2.String(), Amount: nil},
				{Account: s.addr3.String(), Amount: nil},
				{Account: s.addr4.String(), Amount: s.coins("13apple")},
				{Account: s.addr5.String(), Amount: nil},
			},
			expEvents: sdk.Events{
				s.untypeEvent(exchange.NewEventFundsReleased(s.addr1.String(), 2, s.coins("21apple"), eventTag)),
				s.untypeEvent(exchange.NewEventFundsReleased(s.addr2.String(), 2, s.coins("22apple"), eventTag)),
				s.untypeEvent(exchange.NewEventFundsReleased(s.addr4.String(), 2, s.coins("5apple"), eventTag)),
				s.untypeEvent(exchange.NewEventFundsReleased(s.addr4.String(), 2, s.coins("6apple"), eventTag)),
			},
			expRelHoldCalls: []*ReleaseHoldArgs{
				NewReleaseHoldArgs(s.addr1, s.coins("21apple")),
				NewReleaseHoldArgs(s.addr2, s.coins("22apple")),
				NewReleaseHoldArgs(s.addr4, s.coins("5apple")),
				NewReleaseHoldArgs(s.addr4, s.coins("6apple")),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			expHoldCalls := HoldCalls{ReleaseHold: tc.expRelHoldCalls}

			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			if tc.holdKeeper == nil {
				tc.holdKeeper = NewMockHoldKeeper()
			}

			kpr := s.k.WithHoldKeeper(tc.holdKeeper)
			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = kpr.ReleaseCommitments(ctx, tc.marketID, tc.toRelease, eventTag)
			}
			s.Require().NotPanics(testFunc, "ReleaseCommitments(%d)", tc.marketID)
			s.assertErrorValue(err, tc.expErr, "ReleaseCommitments(%d) error", tc.marketID)

			s.assertHoldKeeperCalls(tc.holdKeeper, expHoldCalls, "ReleaseCommitments(%d)", tc.marketID)

			actEvents := em.Events()
			s.assertEqualEvents(tc.expEvents, actEvents, "events emitted during ReleaseCommitments(%d)", tc.marketID)

			var addr sdk.AccAddress
			for i, exp := range tc.expAmounts {
				addr, err = sdk.AccAddressFromBech32(exp.Account)
				if s.Assert().NoError(err, "expAmounts[%d]: AccAddressFromBech32(%q)", i, exp.Account) {
					actAmount := s.k.GetCommitmentAmount(s.ctx, tc.marketID, addr)
					s.Assert().Equal(exp.Amount.String(), actAmount.String(), "expAmounts[%d]: ReleaseCommitments(%d, %s)",
						i, tc.marketID, s.getAddrName(addr))
				}
			}
		})
	}
}

func (s *TestSuite) TestKeeper_ReleaseAllCommitmentsForMarket() {
	type commitment struct {
		marketID uint32
		addr     sdk.AccAddress
		amount   sdk.Coins
	}
	keptCommitments := []commitment{
		{marketID: 1, addr: s.addr1, amount: s.coins("11cherry")},
		{marketID: 1, addr: s.addr2, amount: s.coins("12cherry")},
		{marketID: 1, addr: s.addr3, amount: s.coins("13cherry")},
		{marketID: 1, addr: s.addr4, amount: s.coins("14cherry")},
		{marketID: 1, addr: s.addr5, amount: s.coins("15cherry")},
		{marketID: 4, addr: s.addr1, amount: s.coins("41cherry")},
		{marketID: 4, addr: s.addr2, amount: s.coins("42cherry")},
		{marketID: 4, addr: s.addr3, amount: s.coins("43cherry")},
		{marketID: 4, addr: s.addr4, amount: s.coins("44cherry")},
		{marketID: 4, addr: s.addr5, amount: s.coins("45cherry")},
		{marketID: 41, addr: s.addr1, amount: s.coins("411cherry")},
		{marketID: 41, addr: s.addr2, amount: s.coins("412cherry")},
		{marketID: 41, addr: s.addr3, amount: s.coins("413cherry")},
		{marketID: 41, addr: s.addr4, amount: s.coins("414cherry")},
		{marketID: 41, addr: s.addr5, amount: s.coins("415cherry")},
		{marketID: 44, addr: s.addr1, amount: s.coins("441cherry")},
		{marketID: 44, addr: s.addr2, amount: s.coins("442cherry")},
		{marketID: 44, addr: s.addr3, amount: s.coins("443cherry")},
		{marketID: 44, addr: s.addr4, amount: s.coins("444cherry")},
		{marketID: 44, addr: s.addr5, amount: s.coins("445cherry")},
		{marketID: 440, addr: s.addr1, amount: s.coins("4401cherry")},
		{marketID: 440, addr: s.addr2, amount: s.coins("4402cherry")},
		{marketID: 440, addr: s.addr3, amount: s.coins("4403cherry")},
		{marketID: 440, addr: s.addr4, amount: s.coins("4404cherry")},
		{marketID: 440, addr: s.addr5, amount: s.coins("4405cherry")},
		{marketID: 442, addr: s.addr1, amount: s.coins("4421cherry")},
		{marketID: 442, addr: s.addr2, amount: s.coins("4422cherry")},
		{marketID: 442, addr: s.addr3, amount: s.coins("4423cherry")},
		{marketID: 442, addr: s.addr4, amount: s.coins("4424cherry")},
		{marketID: 442, addr: s.addr5, amount: s.coins("4425cherry")},
		{marketID: 4410, addr: s.addr1, amount: s.coins("44101cherry")},
		{marketID: 4410, addr: s.addr2, amount: s.coins("44102cherry")},
		{marketID: 4410, addr: s.addr3, amount: s.coins("44103cherry")},
		{marketID: 4410, addr: s.addr4, amount: s.coins("44104cherry")},
		{marketID: 4410, addr: s.addr5, amount: s.coins("44105cherry")},
	}
	releasedCommitments := []commitment{
		{marketID: 441, addr: s.addr1, amount: s.coins("4411cherry")},
		{marketID: 441, addr: s.addr2, amount: s.coins("4412cherry")},
		{marketID: 441, addr: s.addr3, amount: s.coins("4413cherry")},
		{marketID: 441, addr: s.addr4, amount: s.coins("4414cherry")},
		{marketID: 441, addr: s.addr5, amount: s.coins("4415cherry")},
	}
	initCommitments := make([]commitment, 0, len(keptCommitments)+len(releasedCommitments))
	initCommitments = append(initCommitments, keptCommitments...)
	initCommitments = append(initCommitments, releasedCommitments...)
	s.clearExchangeState()
	store := s.getStore()
	for _, com := range initCommitments {
		keeper.SetCommitmentAmount(store, com.marketID, com.addr, com.amount)
	}

	kpr := s.k.WithHoldKeeper(NewMockHoldKeeper())
	testFunc := func() {
		kpr.ReleaseAllCommitmentsForMarket(s.ctx, 441)
	}
	s.Require().NotPanics(testFunc, "ReleaseAllCommitmentsForMarket(441)")

	for _, com := range releasedCommitments {
		actAmt := s.k.GetCommitmentAmount(s.ctx, com.marketID, com.addr)
		s.Assert().Equal("", actAmt.String(), "GetCommitmentAmount(%d, %s) (released)", com.marketID, s.getAddrName(com.addr))
	}
	for _, com := range keptCommitments {
		actAmt := s.k.GetCommitmentAmount(s.ctx, com.marketID, com.addr)
		s.Assert().Equal(com.amount.String(), actAmt.String(), "GetCommitmentAmount(%d, %s) (kept)", com.marketID, s.getAddrName(com.addr))
	}
}

func (s *TestSuite) TestKeeper_IterateCommitments() {
	var commitments []exchange.Commitment
	stopAfter := func(count int) func(com exchange.Commitment) bool {
		return func(com exchange.Commitment) bool {
			commitments = append(commitments, com)
			return len(commitments) >= count
		}
	}
	getAll := func(com exchange.Commitment) bool {
		commitments = append(commitments, com)
		return false
	}
	comStr := func(com exchange.Commitment) string {
		return fmt.Sprintf("%d %s %s", com.MarketId, s.getAddrStrName(com.Account), com.Amount)
	}

	tests := []struct {
		name    string
		setup   func()
		cb      func(com exchange.Commitment) bool
		expComs []exchange.Commitment
	}{
		{
			name:    "no commitments",
			cb:      getAll,
			expComs: nil,
		},
		{
			name: "one commitment",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 7, s.addr4, s.coins("74cherry"))
			},
			cb:      getAll,
			expComs: []exchange.Commitment{{MarketId: 7, Account: s.addr4.String(), Amount: s.coins("74cherry")}},
		},
		{
			name: "three commitments: get all",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 3, s.addr2, s.coins("32cherry"))
				keeper.SetCommitmentAmount(store, 7, s.addr4, s.coins("74cherry"))
				keeper.SetCommitmentAmount(store, 22, s.addr1, s.coins("221cherry"))
			},
			cb: getAll,
			expComs: []exchange.Commitment{
				{MarketId: 3, Account: s.addr2.String(), Amount: s.coins("32cherry")},
				{MarketId: 7, Account: s.addr4.String(), Amount: s.coins("74cherry")},
				{MarketId: 22, Account: s.addr1.String(), Amount: s.coins("221cherry")},
			},
		},
		{
			name: "three commitments: get one",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 3, s.addr2, s.coins("32cherry"))
				keeper.SetCommitmentAmount(store, 7, s.addr4, s.coins("74cherry"))
				keeper.SetCommitmentAmount(store, 22, s.addr1, s.coins("221cherry"))
			},
			cb:      stopAfter(1),
			expComs: []exchange.Commitment{{MarketId: 3, Account: s.addr2.String(), Amount: s.coins("32cherry")}},
		},
		{
			name: "three commitments: get two",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 3, s.addr2, s.coins("32cherry"))
				keeper.SetCommitmentAmount(store, 7, s.addr4, s.coins("74cherry"))
				keeper.SetCommitmentAmount(store, 22, s.addr1, s.coins("221cherry"))
			},
			cb: stopAfter(2),
			expComs: []exchange.Commitment{
				{MarketId: 3, Account: s.addr2.String(), Amount: s.coins("32cherry")},
				{MarketId: 7, Account: s.addr4.String(), Amount: s.coins("74cherry")},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			commitments = nil
			testFunc := func() {
				s.k.IterateCommitments(s.ctx, tc.cb)
			}
			s.Require().NotPanics(testFunc, "IterateCommitments")
			assertEqualSlice(s, tc.expComs, commitments, comStr, "IterateCommitments commitments")
		})
	}
}

func (s *TestSuite) TestKeeper_ValidateAndCollectCommitmentCreationFee() {
	tests := []struct {
		name        string
		market      exchange.Market
		bankKeeper  *MockBankKeeper
		marketID    uint32
		addr        sdk.AccAddress
		fee         *sdk.Coin
		expErr      string
		expBankCall bool
	}{
		{
			name:        "market does not require fee: some",
			market:      exchange.Market{MarketId: 1, FeeCreateCommitmentFlat: nil},
			marketID:    1,
			addr:        s.addr1,
			fee:         s.coinP("10cherry"),
			expBankCall: true,
		},
		{
			name:     "market does not require fee: none",
			market:   exchange.Market{MarketId: 1, FeeCreateCommitmentFlat: nil},
			marketID: 1,
			addr:     s.addr1,
			fee:      nil,
		},
		{
			name:     "market has one option: none",
			market:   exchange.Market{MarketId: 1, FeeCreateCommitmentFlat: s.coins("10cherry")},
			marketID: 1,
			addr:     s.addr1,
			fee:      nil,
			expErr:   "no commitment creation fee provided, must be one of: 10cherry",
		},
		{
			name:     "market has one option: not enough",
			market:   exchange.Market{MarketId: 1, FeeCreateCommitmentFlat: s.coins("10cherry")},
			marketID: 1,
			addr:     s.addr1,
			fee:      s.coinP("9cherry"),
			expErr:   "insufficient commitment creation fee: \"9cherry\" is less than required amount \"10cherry\"",
		},
		{
			name:        "market has one option: same",
			market:      exchange.Market{MarketId: 1, FeeCreateCommitmentFlat: s.coins("10cherry")},
			marketID:    1,
			addr:        s.addr1,
			fee:         s.coinP("10cherry"),
			expBankCall: true,
		},
		{
			name:        "market has one option: more",
			market:      exchange.Market{MarketId: 1, FeeCreateCommitmentFlat: s.coins("10cherry")},
			marketID:    1,
			addr:        s.addr1,
			fee:         s.coinP("11cherry"),
			expBankCall: true,
		},
		{
			name:     "market has one option: different denom",
			market:   exchange.Market{MarketId: 1, FeeCreateCommitmentFlat: s.coins("10cherry")},
			marketID: 1,
			addr:     s.addr1,
			fee:      s.coinP("99banana"),
			expErr:   "invalid commitment creation fee \"99banana\", must be one of: 10cherry",
		},
		{
			name:     "market has two options: none",
			market:   exchange.Market{MarketId: 1, FeeCreateCommitmentFlat: s.coins("10cherry,3orange")},
			marketID: 1,
			addr:     s.addr1,
			fee:      nil,
			expErr:   "no commitment creation fee provided, must be one of: 10cherry,3orange",
		},
		{
			name:        "market has two options: first",
			market:      exchange.Market{MarketId: 1, FeeCreateCommitmentFlat: s.coins("10cherry,3orange")},
			marketID:    1,
			addr:        s.addr1,
			fee:         s.coinP("10cherry"),
			expBankCall: true,
		},
		{
			name:        "market has two options: second",
			market:      exchange.Market{MarketId: 1, FeeCreateCommitmentFlat: s.coins("10cherry,3orange")},
			marketID:    1,
			addr:        s.addr1,
			fee:         s.coinP("3orange"),
			expBankCall: true,
		},
		{
			name:     "market has two options: different denom",
			market:   exchange.Market{MarketId: 1, FeeCreateCommitmentFlat: s.coins("10cherry,3orange")},
			marketID: 1,
			addr:     s.addr1,
			fee:      s.coinP("55pear"),
			expErr:   "invalid commitment creation fee \"55pear\", must be one of: 10cherry,3orange",
		},
		{
			name:       "error collecting fee",
			market:     exchange.Market{MarketId: 1, FeeCreateCommitmentFlat: s.coins("10cherry")},
			bankKeeper: NewMockBankKeeper().WithSendCoinsResults("test error injected"),
			marketID:   1,
			addr:       s.addr1,
			fee:        s.coinP("10cherry"),
			expErr: "error collecting commitment creation fee: error transferring 10cherry from " +
				"cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqjwl8sq to market 1: test error injected",
			expBankCall: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			s.k.SetParams(s.ctx, &exchange.Params{DefaultSplit: 0})
			s.requireCreateMarket(tc.market)

			var expBankCalls BankCalls
			if tc.expBankCall {
				expBankCalls.SendCoins = append(expBankCalls.SendCoins, &SendCoinsArgs{
					fromAddr: tc.addr,
					toAddr:   exchange.GetMarketAddress(tc.marketID),
					amt:      sdk.Coins{*tc.fee},
				})
			}

			if tc.bankKeeper == nil {
				tc.bankKeeper = NewMockBankKeeper()
			}
			kpr := s.k.WithBankKeeper(tc.bankKeeper)
			var err error
			testFunc := func() {
				err = kpr.ValidateAndCollectCommitmentCreationFee(s.ctx, tc.marketID, tc.addr, tc.fee)
			}
			s.Require().NotPanics(testFunc, "ValidateAndCollectCommitmentCreationFee(%d, %s, %s)",
				tc.marketID, s.getAddrName(tc.addr), tc.fee)
			s.assertErrorValue(err, tc.expErr, "ValidateAndCollectCommitmentCreationFee(%d, %s, %s) error",
				tc.marketID, s.getAddrName(tc.addr), tc.fee)
			s.assertBankKeeperCalls(tc.bankKeeper, expBankCalls, "ValidateAndCollectCommitmentCreationFee(%d, %s, %s)",
				tc.marketID, s.getAddrName(tc.addr), tc.fee)
		})
	}
}

// TODO[1789]: func (s *TestSuite) TestKeeper_CalculateCommitmentSettlementFee()

// TODO[1789]: func (s *TestSuite) TestKeeper_SettleCommitments()
