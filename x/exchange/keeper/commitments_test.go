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

// TODO[1789]: func (s *TestSuite) TestKeeper_ReleaseCommitment()

// TODO[1789]: func (s *TestSuite) TestKeeper_ReleaseCommitments()

// TODO[1789]: func (s *TestSuite) TestKeeper_ReleaseAllCommitmentsForMarket()

// TODO[1789]: func (s *TestSuite) TestKeeper_IterateCommitments()

// TODO[1789]: func (s *TestSuite) TestKeeper_ValidateAndCollectCommitmentCreationFee()

// TODO[1789]: func (s *TestSuite) TestKeeper_CalculateCommitmentSettlementFee()

// TODO[1789]: func (s *TestSuite) TestKeeper_SettleCommitments()
