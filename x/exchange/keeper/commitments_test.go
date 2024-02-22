package keeper_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/keeper"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
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
	existingSetup := func(reqAttr ...string) func() {
		return func() {
			s.requireCreateMarket(exchange.Market{
				MarketId:                 1,
				AcceptingCommitments:     true,
				CommitmentSettlementBips: 51,
				IntermediaryDenom:        "cherry",
				ReqAttrCreateCommitment:  reqAttr,
			})
			s.requireCreateMarket(exchange.Market{
				MarketId:                 2,
				AcceptingCommitments:     true,
				CommitmentSettlementBips: 52,
				IntermediaryDenom:        "cherry",
				ReqAttrCreateCommitment:  reqAttr,
			})
			s.requireCreateMarket(exchange.Market{
				MarketId:                 3,
				AcceptingCommitments:     true,
				CommitmentSettlementBips: 53,
				IntermediaryDenom:        "cherry",
				ReqAttrCreateCommitment:  reqAttr,
			})
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
	}

	tests := []struct {
		name        string
		setup       func()
		holdKeeper  *MockHoldKeeper
		attrKeeper  *MockAttributeKeeper
		marketID    uint32
		addr        sdk.AccAddress
		amount      sdk.Coins
		expErr      string
		expHoldCall bool
		expAttrCall bool
		expEvent    bool
		expAmount   sdk.Coins
	}{
		{
			name:     "zero amount",
			marketID: 2,
			addr:     s.addr3,
			amount:   nil,
		},
		{
			name:     "negative amount",
			marketID: 2,
			addr:     s.addr3,
			amount:   sdk.Coins{sdk.Coin{Denom: "apple", Amount: sdkmath.NewInt(-23)}},
			expErr:   "cannot add negative commitment amount \"-23apple\" for " + s.addr3.String() + " in market 2",
		},
		{
			name:     "market does not exist",
			marketID: 2,
			addr:     s.addr3,
			amount:   s.coins("23apple"),
			expErr:   "market 2 does not exist",
		},
		{
			name: "market is not accepting commitments",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 8,
					AcceptingCommitments:     false,
					CommitmentSettlementBips: 52,
					IntermediaryDenom:        "cherry",
				})
			},
			marketID: 8,
			addr:     s.addr4,
			amount:   s.coins("34apple"),
			expErr:   "market 8 is not accepting commitments",
		},
		{
			name: "user does not have req attr",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 7,
					AcceptingCommitments:     true,
					CommitmentSettlementBips: 52,
					IntermediaryDenom:        "cherry",
					ReqAttrCreateCommitment:  []string{"com.can.do"},
				})
			},
			marketID:    7,
			addr:        s.addr1,
			amount:      s.coins("71banana"),
			expErr:      "account " + s.addr1.String() + " is not allowed to create commitments in market 7",
			expAttrCall: true,
		},
		{
			name: "empty state: error adding hold",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 2,
					AcceptingCommitments:     true,
					CommitmentSettlementBips: 52,
					IntermediaryDenom:        "cherry",
				})
			},
			holdKeeper:  NewMockHoldKeeper().WithAddHoldResults("injected testing error"),
			marketID:    2,
			addr:        s.addr3,
			amount:      s.coins("23apple"),
			expErr:      "injected testing error",
			expHoldCall: true,
		},
		{
			name: "empty state: new commitment",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 2,
					AcceptingCommitments:     true,
					CommitmentSettlementBips: 52,
					IntermediaryDenom:        "cherry",
				})
			},
			marketID:    2,
			addr:        s.addr3,
			amount:      s.coins("23apple"),
			expHoldCall: true,
			expEvent:    true,
			expAmount:   s.coins("23apple"),
		},
		{
			name: "empty state: new commitment with req attr",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 4,
					AcceptingCommitments:     true,
					CommitmentSettlementBips: 52,
					IntermediaryDenom:        "cherry",
					ReqAttrCreateCommitment:  []string{"com.can.do"},
				})
			},
			attrKeeper:  NewMockAttributeKeeper().WithGetAllAttributesAddrResult(s.addr4, []string{"com.can.do"}, ""),
			marketID:    4,
			addr:        s.addr4,
			amount:      s.coins("15peach"),
			expHoldCall: true,
			expAttrCall: true,
			expEvent:    true,
			expAmount:   s.coins("15peach"),
		},
		{
			name: "almost empty state: additional commitment",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 2,
					AcceptingCommitments:     true,
					CommitmentSettlementBips: 52,
					IntermediaryDenom:        "cherry",
				})
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
			name: "existing commitments: market is not accepting commitments",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 2,
					AcceptingCommitments:     false,
					CommitmentSettlementBips: 52,
					IntermediaryDenom:        "cherry",
				})
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 2, s.addr1, s.coins("21apple"))
				keeper.SetCommitmentAmount(store, 2, s.addr2, s.coins("22apple"))
				keeper.SetCommitmentAmount(store, 2, s.addr4, s.coins("24apple"))
			},
			marketID:  2,
			addr:      s.addr1,
			amount:    s.coins("79apple"),
			expErr:    "market 2 is not accepting commitments",
			expAmount: s.coins("21apple"),
		},
		{
			name: "existing commitments: user does not have required attribute",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 2,
					AcceptingCommitments:     true,
					CommitmentSettlementBips: 52,
					IntermediaryDenom:        "cherry",
					ReqAttrCreateCommitment:  []string{"just.some.com.okay"},
				})
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 2, s.addr1, s.coins("21apple"))
				keeper.SetCommitmentAmount(store, 2, s.addr2, s.coins("22apple"))
				keeper.SetCommitmentAmount(store, 2, s.addr4, s.coins("24apple"))
			},
			marketID:    2,
			addr:        s.addr4,
			amount:      s.coins("77banana"),
			expErr:      "account " + s.addr4.String() + " is not allowed to create commitments in market 2",
			expAttrCall: true,
			expAmount:   s.coins("24apple"),
		},
		{
			name:        "existing commitments: error adding hold",
			setup:       existingSetup(),
			holdKeeper:  NewMockHoldKeeper().WithAddHoldResults("injected testing error"),
			marketID:    2,
			addr:        s.addr2,
			amount:      s.coins("100apple"),
			expErr:      "injected testing error",
			expHoldCall: true,
			expAmount:   s.coins("22apple"),
		},
		{
			name:        "existing commitments: new commitment",
			setup:       existingSetup(),
			marketID:    2,
			addr:        s.addr3,
			amount:      s.coins("23apple"),
			expHoldCall: true,
			expEvent:    true,
			expAmount:   s.coins("23apple"),
		},
		{
			name:        "existing commitments: additional commitment",
			setup:       existingSetup(),
			marketID:    2,
			addr:        s.addr2,
			amount:      s.coins("100apple"),
			expHoldCall: true,
			expEvent:    true,
			expAmount:   s.coins("122apple"),
		},
		{
			name:        "existing commitments: additional commitment with req attr",
			setup:       existingSetup("magic.com.creator"),
			attrKeeper:  NewMockAttributeKeeper().WithGetAllAttributesAddrResult(s.addr2, []string{"magic.com.creator"}, ""),
			marketID:    2,
			addr:        s.addr2,
			amount:      s.coins("100apple"),
			expHoldCall: true,
			expAttrCall: true,
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
			var expAttrCalls AttributeCalls
			if tc.expAttrCall {
				expAttrCalls.GetAllAttributesAddr = append(expAttrCalls.GetAllAttributesAddr, tc.addr)
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
			if tc.attrKeeper == nil {
				tc.attrKeeper = NewMockAttributeKeeper()
			}
			kpr := s.k.WithHoldKeeper(tc.holdKeeper).WithAttributeKeeper(tc.attrKeeper)
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
			s.assertAttributeKeeperCalls(tc.attrKeeper, expAttrCalls, "AddCommitment(%d, %s, %q)",
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

func (s *TestSuite) TestKeeper_AddCommitmentsUnsafe() {
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
				err = kpr.AddCommitmentsUnsafe(ctx, tc.marketID, tc.toAdd, eventTag)
			}
			s.Require().NotPanics(testFunc, "addCommitmentsUnsafe(%d)", tc.marketID)
			s.assertErrorValue(err, tc.expErr, "addCommitmentsUnsafe(%d) error", tc.marketID)

			s.assertHoldKeeperCalls(tc.holdKeeper, expHoldCalls, "addCommitmentsUnsafe(%d)", tc.marketID)

			actEvents := em.Events()
			s.assertEqualEvents(tc.expEvents, actEvents, "events emitted during addCommitmentsUnsafe(%d)", tc.marketID)

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
				expEvents = append(expEvents, s.untypeEvent(exchange.NewEventCommitmentReleased(tc.addr.String(), tc.marketID, tc.expEventAmt, tc.name)))
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
				s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr2.String(), 2, s.coins("3apple"), eventTag)),
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
				s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr2.String(), 2, s.coins("22apple"), eventTag)),
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
				s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr4.String(), 2, s.coins("24apple"), eventTag)),
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
				s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr1.String(), 2, s.coins("21apple"), eventTag)),
				s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr2.String(), 2, s.coins("22apple"), eventTag)),
				s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr4.String(), 2, s.coins("5apple"), eventTag)),
				s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr4.String(), 2, s.coins("6apple"), eventTag)),
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
			s.assertEqualCommitments(tc.expComs, commitments, "IterateCommitments commitments")
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

func (s *TestSuite) TestKeeper_CalculateCommitmentSettlementFee() {
	// These tests all assume that the fee denom is nhash.
	s.Require().Equal("nhash", pioconfig.GetProvenanceConfig().FeeDenom, "pioconfig.GetProvenanceConfig().FeeDenom")

	tests := []struct {
		name         string
		setup        func()
		markerKeeper *MockMarkerKeeper
		expGetNav    []*GetNetAssetValueArgs
		req          *exchange.MsgMarketCommitmentSettleRequest
		expResp      *exchange.QueryCommitmentSettlementFeeCalcResponse
		expErr       string
	}{
		{
			name:   "nil req",
			req:    nil,
			expErr: "settlement request cannot be nil",
		},
		{
			name:   "invalid req",
			req:    &exchange.MsgMarketCommitmentSettleRequest{MarketId: 0},
			expErr: "invalid market id: cannot be zero",
		},
		{
			name: "no bips",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 3,
					CommitmentSettlementBips: 0,
					IntermediaryDenom:        "",
				})
			},
			req: &exchange.MsgMarketCommitmentSettleRequest{
				MarketId: 3,
				Inputs:   []exchange.AccountAmount{{Account: s.addr2.String(), Amount: s.coins("10apple")}},
				Outputs:  []exchange.AccountAmount{{Account: s.addr3.String(), Amount: s.coins("10apple")}},
			},
			expResp: &exchange.QueryCommitmentSettlementFeeCalcResponse{},
		},
		{
			name: "bips but no denom",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 3,
					CommitmentSettlementBips: 10,
					IntermediaryDenom:        "",
				})
			},
			req: &exchange.MsgMarketCommitmentSettleRequest{
				MarketId: 3,
				Inputs:   []exchange.AccountAmount{{Account: s.addr2.String(), Amount: s.coins("10apple")}},
				Outputs:  []exchange.AccountAmount{{Account: s.addr3.String(), Amount: s.coins("10apple")}},
			},
			expResp: nil,
			expErr:  "market 3 does not have an intermediary denom",
		},
		{
			name: "no nav from intermediary denom to fee denom",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 3,
					CommitmentSettlementBips: 10,
					IntermediaryDenom:        "cherry",
				})
			},
			expGetNav: []*GetNetAssetValueArgs{{markerDenom: "cherry", priceDenom: "nhash"}},
			req: &exchange.MsgMarketCommitmentSettleRequest{
				MarketId: 3,
				Inputs:   []exchange.AccountAmount{{Account: s.addr2.String(), Amount: s.coins("10apple")}},
				Outputs:  []exchange.AccountAmount{{Account: s.addr3.String(), Amount: s.coins("10apple")}},
			},
			expErr: "no nav found from intermediary denom \"cherry\" to fee denom \"nhash\"",
		},
		{
			name: "no inputs",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 3,
					CommitmentSettlementBips: 10,
					IntermediaryDenom:        "cherry",
				})
			},
			req: &exchange.MsgMarketCommitmentSettleRequest{
				MarketId: 3,
				Navs:     []exchange.NetAssetPrice{{Assets: s.coin("10cherry"), Price: s.coin("30nhash")}},
			},
			expResp: &exchange.QueryCommitmentSettlementFeeCalcResponse{
				ToFeeNav: &exchange.NetAssetPrice{Assets: s.coin("10cherry"), Price: s.coin("30nhash")},
			},
		},
		{
			name: "intermediary is fee: no inputs",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 3,
					CommitmentSettlementBips: 10,
					IntermediaryDenom:        "nhash",
				})
			},
			req: &exchange.MsgMarketCommitmentSettleRequest{
				MarketId: 3,
			},
			expResp: &exchange.QueryCommitmentSettlementFeeCalcResponse{
				ToFeeNav: &exchange.NetAssetPrice{Assets: s.coin("1nhash"), Price: s.coin("1nhash")},
			},
		},
		{
			name: "two inputs: no navs",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 3,
					CommitmentSettlementBips: 10,
					IntermediaryDenom:        "cherry",
				})
			},
			markerKeeper: NewMockMarkerKeeper().WithGetNetAssetValueResult(s.coin("10cherry"), s.coin("30nhash")),
			expGetNav: []*GetNetAssetValueArgs{
				{markerDenom: "cherry", priceDenom: "nhash"},
				{markerDenom: "apple", priceDenom: "cherry"},
				{markerDenom: "banana", priceDenom: "cherry"},
			},
			req: &exchange.MsgMarketCommitmentSettleRequest{
				MarketId: 3,
				Inputs: []exchange.AccountAmount{
					{Account: s.addr2.String(), Amount: s.coins("10apple,3banana")},
					{Account: s.addr3.String(), Amount: s.coins("2banana")},
				},
				Outputs: []exchange.AccountAmount{{Account: s.addr4.String(), Amount: s.coins("10apple,5banana")}},
			},
			expErr: s.joinErrs(
				"no nav found from assets denom \"apple\" to intermediary denom \"cherry\"",
				"no nav found from assets denom \"banana\" to intermediary denom \"cherry\"",
			),
		},
		{
			name: "intermediary is fee: two inputs",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 3,
					CommitmentSettlementBips: 1000,
					IntermediaryDenom:        "nhash",
				})
			},
			markerKeeper: NewMockMarkerKeeper().WithGetNetAssetValueResult(s.coin("10apple"), s.coin("57nhash")),
			expGetNav:    []*GetNetAssetValueArgs{{markerDenom: "apple", priceDenom: "nhash"}},
			req: &exchange.MsgMarketCommitmentSettleRequest{
				MarketId: 3,
				Inputs: []exchange.AccountAmount{
					{Account: s.addr2.String(), Amount: s.coins("10apple,3banana")},
					{Account: s.addr3.String(), Amount: s.coins("2banana")},
				},
				Outputs: []exchange.AccountAmount{{Account: s.addr4.String(), Amount: s.coins("10apple,5banana")}},
				Navs:    []exchange.NetAssetPrice{{Assets: s.coin("5banana"), Price: s.coin("13nhash")}},
			},
			expResp: &exchange.QueryCommitmentSettlementFeeCalcResponse{
				InputTotal: s.coins("10apple,5banana"),
				// 10apple*57nhash/10apple = 57nhash
				// 5banana*13nhash/5banana = 13nhash
				// sum = 70nhash
				ConvertedTotal: s.coins("70nhash"),
				// 70nhash * 1000/20000 = 3.5 => 4nhash
				ExchangeFees: s.coins("4nhash"),
				ConversionNavs: []exchange.NetAssetPrice{
					{Assets: s.coin("10apple"), Price: s.coin("57nhash")},
					{Assets: s.coin("5banana"), Price: s.coin("13nhash")},
				},
				ToFeeNav: &exchange.NetAssetPrice{Assets: s.coin("1nhash"), Price: s.coin("1nhash")},
			},
		},
		{
			name: "one input denom: fee denom: evenly divisible",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 3,
					CommitmentSettlementBips: 100,
					IntermediaryDenom:        "cherry",
				})
			},
			markerKeeper: NewMockMarkerKeeper().WithGetNetAssetValueResult(s.coin("10cherry"), s.coin("30nhash")),
			expGetNav:    []*GetNetAssetValueArgs{{markerDenom: "cherry", priceDenom: "nhash"}},
			req: &exchange.MsgMarketCommitmentSettleRequest{
				MarketId: 3,
				Inputs: []exchange.AccountAmount{
					{Account: s.addr2.String(), Amount: s.coins("2200nhash")},
					{Account: s.addr3.String(), Amount: s.coins("5400nhash")},
				},
				Outputs: []exchange.AccountAmount{{Account: s.addr4.String(), Amount: s.coins("7600nhash")}},
			},
			expResp: &exchange.QueryCommitmentSettlementFeeCalcResponse{
				InputTotal:     s.coins("7600nhash"),
				ConvertedTotal: s.coins("7600nhash"),
				// 7500nhash * 100/20000 = 38nhash
				ExchangeFees:   s.coins("38nhash"),
				ConversionNavs: nil,
				ToFeeNav:       &exchange.NetAssetPrice{Assets: s.coin("10cherry"), Price: s.coin("30nhash")},
			},
		},
		{
			name: "one input denom: fee denom: not evenly divisible",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 3,
					CommitmentSettlementBips: 100,
					IntermediaryDenom:        "cherry",
				})
			},
			markerKeeper: NewMockMarkerKeeper().WithGetNetAssetValueResult(s.coin("10cherry"), s.coin("30nhash")),
			expGetNav:    []*GetNetAssetValueArgs{{markerDenom: "cherry", priceDenom: "nhash"}},
			req: &exchange.MsgMarketCommitmentSettleRequest{
				MarketId: 3,
				Inputs: []exchange.AccountAmount{
					{Account: s.addr2.String(), Amount: s.coins("2200nhash")},
					{Account: s.addr3.String(), Amount: s.coins("5300nhash")},
				},
				Outputs: []exchange.AccountAmount{{Account: s.addr4.String(), Amount: s.coins("7500nhash")}},
			},
			expResp: &exchange.QueryCommitmentSettlementFeeCalcResponse{
				InputTotal:     s.coins("7500nhash"),
				ConvertedTotal: s.coins("7500nhash"),
				// 7500nhash * 100/20000 = 37.5 => 38nhash
				ExchangeFees:   s.coins("38nhash"),
				ConversionNavs: nil,
				ToFeeNav:       &exchange.NetAssetPrice{Assets: s.coin("10cherry"), Price: s.coin("30nhash")},
			},
		},
		{
			name: "one input denom: intermediary denom: evenly divisible",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 3,
					CommitmentSettlementBips: 100,
					IntermediaryDenom:        "cherry",
				})
			},
			markerKeeper: NewMockMarkerKeeper().WithGetNetAssetValueResult(s.coin("10cherry"), s.coin("30nhash")),
			expGetNav:    []*GetNetAssetValueArgs{{markerDenom: "cherry", priceDenom: "nhash"}},
			req: &exchange.MsgMarketCommitmentSettleRequest{
				MarketId: 3,
				Inputs: []exchange.AccountAmount{
					{Account: s.addr2.String(), Amount: s.coins("358cherry")},
					{Account: s.addr3.String(), Amount: s.coins("642cherry")},
				},
				Outputs: []exchange.AccountAmount{{Account: s.addr4.String(), Amount: s.coins("1000cherry")}},
			},
			expResp: &exchange.QueryCommitmentSettlementFeeCalcResponse{
				InputTotal:     s.coins("1000cherry"),
				ConvertedTotal: s.coins("1000cherry"),
				// 3000nhash * 100/20000 = 15nhash
				// 1000cherry*30nhash/10cherry = 3000nhash
				ExchangeFees:   s.coins("15nhash"),
				ConversionNavs: nil,
				ToFeeNav:       &exchange.NetAssetPrice{Assets: s.coin("10cherry"), Price: s.coin("30nhash")},
			},
		},
		{
			name: "one input denom: intermediary denom: not evenly divisible",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 2,
					CommitmentSettlementBips: 100,
					IntermediaryDenom:        "cherry",
				})
			},
			markerKeeper: NewMockMarkerKeeper().WithGetNetAssetValueResult(s.coin("10cherry"), s.coin("31nhash")),
			expGetNav:    []*GetNetAssetValueArgs{{markerDenom: "cherry", priceDenom: "nhash"}},
			req: &exchange.MsgMarketCommitmentSettleRequest{
				MarketId: 2,
				Inputs: []exchange.AccountAmount{
					{Account: s.addr2.String(), Amount: s.coins("35cherry")},
					{Account: s.addr3.String(), Amount: s.coins("88cherry")},
				},
				Outputs: []exchange.AccountAmount{{Account: s.addr4.String(), Amount: s.coins("123cherry")}},
			},
			expResp: &exchange.QueryCommitmentSettlementFeeCalcResponse{
				InputTotal:     s.coins("123cherry"),
				ConvertedTotal: s.coins("123cherry"),
				// 123cherry*31nhash/10cherry = 381.3 => 382nhash
				// 382nhash * 100/20000 = 1.9065 => 2nhash
				ExchangeFees:   s.coins("2nhash"),
				ConversionNavs: nil,
				ToFeeNav:       &exchange.NetAssetPrice{Assets: s.coin("10cherry"), Price: s.coin("31nhash")},
			},
		},
		{
			name: "one input denom: nav from req: evenly divisible",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 2,
					CommitmentSettlementBips: 100,
					IntermediaryDenom:        "cherry",
				})
			},
			markerKeeper: NewMockMarkerKeeper().WithGetNetAssetValueResult(s.coin("10cherry"), s.coin("30nhash")),
			expGetNav:    []*GetNetAssetValueArgs{{markerDenom: "cherry", priceDenom: "nhash"}},
			req: &exchange.MsgMarketCommitmentSettleRequest{
				MarketId: 2,
				Inputs: []exchange.AccountAmount{
					{Account: s.addr2.String(), Amount: s.coins("25apple")},
					{Account: s.addr3.String(), Amount: s.coins("17apple")},
				},
				Outputs: []exchange.AccountAmount{{Account: s.addr4.String(), Amount: s.coins("42apple")}},
				Navs:    []exchange.NetAssetPrice{{Assets: s.coin("21apple"), Price: s.coin("500cherry")}},
			},
			expResp: &exchange.QueryCommitmentSettlementFeeCalcResponse{
				InputTotal: s.coins("42apple"),
				// 42apple*500cherry/21apple = 1000cherry
				ConvertedTotal: s.coins("1000cherry"),
				// 1000cherry*30nhash/10cherry = 3000nhash
				// 3000nhas * 100/20000 = 15nhash
				ExchangeFees:   s.coins("15nhash"),
				ConversionNavs: []exchange.NetAssetPrice{{Assets: s.coin("21apple"), Price: s.coin("500cherry")}},
				ToFeeNav:       &exchange.NetAssetPrice{Assets: s.coin("10cherry"), Price: s.coin("30nhash")},
			},
		},
		{
			name: "one input denom: nav from req: not evenly divisible",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 2,
					CommitmentSettlementBips: 100,
					IntermediaryDenom:        "cherry",
				})
			},
			markerKeeper: NewMockMarkerKeeper().WithGetNetAssetValueResult(s.coin("10cherry"), s.coin("30nhash")),
			expGetNav:    []*GetNetAssetValueArgs{{markerDenom: "cherry", priceDenom: "nhash"}},
			req: &exchange.MsgMarketCommitmentSettleRequest{
				MarketId: 2,
				Inputs: []exchange.AccountAmount{
					{Account: s.addr2.String(), Amount: s.coins("25apple")},
					{Account: s.addr3.String(), Amount: s.coins("18apple")},
				},
				Outputs: []exchange.AccountAmount{{Account: s.addr4.String(), Amount: s.coins("43apple")}},
				Navs:    []exchange.NetAssetPrice{{Assets: s.coin("21apple"), Price: s.coin("500cherry")}},
			},
			expResp: &exchange.QueryCommitmentSettlementFeeCalcResponse{
				InputTotal: s.coins("43apple"),
				// 43apple*500cherry/21apple = 1023.80 => 1024cherry
				ConvertedTotal: s.coins("1024cherry"),
				// 1024cherry*30nhash/10cherry = 3072nhash
				// 3072nhash * 100/20000 = 15.36 => 16nhash
				ExchangeFees:   s.coins("16nhash"),
				ConversionNavs: []exchange.NetAssetPrice{{Assets: s.coin("21apple"), Price: s.coin("500cherry")}},
				ToFeeNav:       &exchange.NetAssetPrice{Assets: s.coin("10cherry"), Price: s.coin("30nhash")},
			},
		},
		{
			name: "one input denom: nav from marker keeper",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 2,
					CommitmentSettlementBips: 100,
					IntermediaryDenom:        "cherry",
				})
			},
			markerKeeper: NewMockMarkerKeeper().
				WithGetNetAssetValueResult(s.coin("10cherry"), s.coin("30nhash")).
				WithGetNetAssetValueResult(s.coin("21apple"), s.coin("500cherry")),
			expGetNav: []*GetNetAssetValueArgs{
				{markerDenom: "cherry", priceDenom: "nhash"},
				{markerDenom: "apple", priceDenom: "cherry"},
			},
			req: &exchange.MsgMarketCommitmentSettleRequest{
				MarketId: 2,
				Inputs: []exchange.AccountAmount{
					{Account: s.addr2.String(), Amount: s.coins("25apple")},
					{Account: s.addr3.String(), Amount: s.coins("18apple")},
				},
				Outputs: []exchange.AccountAmount{{Account: s.addr4.String(), Amount: s.coins("43apple")}},
			},
			expResp: &exchange.QueryCommitmentSettlementFeeCalcResponse{
				InputTotal: s.coins("43apple"),
				// 43apple*500cherry/21apple = 1023.8 => 1024cherry
				ConvertedTotal: s.coins("1024cherry"),
				// 1024cherry*30nhash/10cherry = 3072nhash
				// 3072nhash * 100/20000 = 15.36nhash => 16nhash
				ExchangeFees:   s.coins("16nhash"),
				ConversionNavs: []exchange.NetAssetPrice{{Assets: s.coin("21apple"), Price: s.coin("500cherry")}},
				ToFeeNav:       &exchange.NetAssetPrice{Assets: s.coin("10cherry"), Price: s.coin("30nhash")},
			},
		},
		{
			name: "four input denoms",
			setup: func() {
				s.requireCreateMarket(exchange.Market{
					MarketId:                 2,
					CommitmentSettlementBips: 25,
					IntermediaryDenom:        "cherry",
				})
			},
			markerKeeper: NewMockMarkerKeeper().
				WithGetNetAssetValueResult(s.coin("10cherry"), s.coin("31nhash")).
				WithGetNetAssetValueResult(s.coin("21apple"), s.coin("500cherry")).
				WithGetNetAssetValueResult(s.coin("1banana"), s.coin("6666cherry")), // ignored.
			expGetNav: []*GetNetAssetValueArgs{
				{markerDenom: "cherry", priceDenom: "nhash"},
				{markerDenom: "apple", priceDenom: "cherry"},
			},
			req: &exchange.MsgMarketCommitmentSettleRequest{
				MarketId: 2,
				Inputs: []exchange.AccountAmount{
					{Account: s.addr2.String(), Amount: s.coins("12apple")},
					{Account: s.addr3.String(), Amount: s.coins("3apple,88banana")},
					{Account: s.addr4.String(), Amount: s.coins("15apple,75cherry,300nhash")},
					{Account: s.addr5.String(), Amount: s.coins("5banana,12cherry")},
				},
				Outputs: []exchange.AccountAmount{
					{Account: s.addr5.String(), Amount: s.coins("30apple")},
					{Account: s.addr4.String(), Amount: s.coins("93banana")},
					{Account: s.addr2.String(), Amount: s.coins("87cherry,300nhash")},
				},
				Navs: []exchange.NetAssetPrice{{Assets: s.coin("44banana"), Price: s.coin("77cherry")}},
			},
			expResp: &exchange.QueryCommitmentSettlementFeeCalcResponse{
				InputTotal: s.coins("30apple,93banana,87cherry,300nhash"),
				// 30apple*500cherry/21apple = 714.285714285714
				// 93banana*77cherry/44banana = 162.75
				// 87cherry => 87
				// sum = 964.035714285714=> 965cherry
				ConvertedTotal: s.coins("965cherry,300nhash"),
				// 965cherry *31nhash/10cherry = 2991.5 => 2992nhash
				// sum = 3292nhash
				// 3292nhash * 25/20000 = 4.115 => 5nhash
				ExchangeFees: s.coins("5nhash"),
				ConversionNavs: []exchange.NetAssetPrice{
					{Assets: s.coin("21apple"), Price: s.coin("500cherry")},
					{Assets: s.coin("44banana"), Price: s.coin("77cherry")},
				},
				ToFeeNav: &exchange.NetAssetPrice{Assets: s.coin("10cherry"), Price: s.coin("31nhash")},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			expMarkerCalls := MarkerCalls{GetNetAssetValue: tc.expGetNav}

			if tc.markerKeeper == nil {
				tc.markerKeeper = NewMockMarkerKeeper()
			}
			kpr := s.k.WithMarkerKeeper(tc.markerKeeper)

			var resp *exchange.QueryCommitmentSettlementFeeCalcResponse
			var err error
			testFunc := func() {
				resp, err = kpr.CalculateCommitmentSettlementFee(s.ctx, tc.req)
			}
			s.Require().NotPanics(testFunc, "CalculateCommitmentSettlementFee")
			s.assertErrorValue(err, tc.expErr, "CalculateCommitmentSettlementFee error")
			if !s.Assert().Equal(tc.expResp, resp, "CalculateCommitmentSettlementFee response") && tc.expResp != nil && resp != nil {
				s.Assert().Equal(tc.expResp.ExchangeFees.String(), resp.ExchangeFees.String(), "ExchangeFees")
				s.Assert().Equal(tc.expResp.InputTotal.String(), resp.InputTotal.String(), "InputTotal")
				s.Assert().Equal(tc.expResp.ConvertedTotal.String(), resp.ConvertedTotal.String(), "ConvertedTotal")
				assertEqualSlice(s, tc.expResp.ConversionNavs, resp.ConversionNavs, exchange.NetAssetPrice.String, "ConversionNavs")
				if !s.Assert().Equal(tc.expResp.ToFeeNav, resp.ToFeeNav, "ToFeeNav") && tc.expResp.ToFeeNav != nil && resp.ToFeeNav != nil {
					s.Assert().Equal(tc.expResp.ToFeeNav.String(), resp.ToFeeNav.String(), "ToFeeNav strings")
				}
			}
			s.assertMarkerKeeperCalls(tc.markerKeeper, expMarkerCalls, "CalculateCommitmentSettlementFee")
		})
	}
}

func (s *TestSuite) TestKeeper_SettleCommitments() {
	appleMarker := s.markerAccount("100000apple")
	bananaMarker := s.markerAccount("100000banana")
	holdReason := func(marketID uint32) string {
		return fmt.Sprintf("x/exchange: commitment to %d", marketID)
	}
	navSource := func(marketID uint32) string {
		return fmt.Sprintf("x/exchange market %d", marketID)
	}

	tests := []struct {
		name           string
		setup          func()
		markerKeeper   *MockMarkerKeeper
		holdKeeper     *MockHoldKeeper
		bankKeeper     *MockBankKeeper
		req            *exchange.MsgMarketCommitmentSettleRequest
		expEvents      sdk.Events
		expMarkerCalls MarkerCalls
		expHoldCalls   HoldCalls
		expBankCalls   BankCalls
		expErr         string
	}{
		{
			name: "cannot build transfers",
			req: &exchange.MsgMarketCommitmentSettleRequest{
				Admin:    s.addr1.String(),
				MarketId: 3,
				Inputs: []exchange.AccountAmount{
					{Account: s.addr2.String(), Amount: s.coins("10apple")},
					{Account: s.addr3.String(), Amount: s.coins("20apple")},
				},
				Outputs: []exchange.AccountAmount{
					{Account: s.addr2.String(), Amount: s.coins("10apple")},
					{Account: s.addr3.String(), Amount: s.coins("19apple")},
				},
			},
			expErr: "failed to build transfers: failed to allocate 20apple to outputs: 1 left over",
		},
		{
			name: "cannot release commitments",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 2, s.addr2, s.coins("9apple"))
			},
			req: &exchange.MsgMarketCommitmentSettleRequest{
				Admin:    s.addr1.String(),
				MarketId: 2,
				Inputs:   []exchange.AccountAmount{{Account: s.addr2.String(), Amount: s.coins("10apple")}},
				Outputs:  []exchange.AccountAmount{{Account: s.addr3.String(), Amount: s.coins("10apple")}},
			},
			expErr: "failed to release commitments on inputs and fees: " +
				"commitment amount to release \"10apple\" is more than currently committed amount \"9apple\" for " +
				s.addr2.String() + " in market 2",
		},
		{
			name: "transfer failure",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 2, s.addr2, s.coins("10apple"))
			},
			req: &exchange.MsgMarketCommitmentSettleRequest{
				Admin:    s.addr1.String(),
				MarketId: 2,
				Inputs:   []exchange.AccountAmount{{Account: s.addr2.String(), Amount: s.coins("10apple")}},
				Outputs:  []exchange.AccountAmount{{Account: s.addr3.String(), Amount: s.coins("11apple")}},
				EventTag: "testtag1",
			},
			expEvents: sdk.Events{
				s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr2.String(), 2, s.coins("10apple"), "testtag1")),
			},
			expHoldCalls: HoldCalls{
				ReleaseHold: []*ReleaseHoldArgs{NewReleaseHoldArgs(s.addr2, s.coins("10apple"))},
			},
			expErr: "input coins \"10apple\" does not equal output coins \"11apple\"",
		},
		{
			name: "cannot add new commitments",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 4, s.addr3, s.coins("10apple"))
			},
			holdKeeper: NewMockHoldKeeper().WithAddHoldResults("no hold 4u"),
			req: &exchange.MsgMarketCommitmentSettleRequest{
				Admin:    s.addr1.String(),
				MarketId: 4,
				Inputs:   []exchange.AccountAmount{{Account: s.addr3.String(), Amount: s.coins("10apple")}},
				Outputs:  []exchange.AccountAmount{{Account: s.addr2.String(), Amount: s.coins("10apple")}},
				EventTag: "testtag2",
			},
			expEvents: sdk.Events{
				s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr3.String(), 4, s.coins("10apple"), "testtag2")),
			},
			expHoldCalls: HoldCalls{
				ReleaseHold: []*ReleaseHoldArgs{NewReleaseHoldArgs(s.addr3, s.coins("10apple"))},
				AddHold:     []*AddHoldArgs{NewAddHoldArgs(s.addr2, s.coins("10apple"), holdReason(4))},
			},
			expBankCalls: BankCalls{
				BlockedAddr: []sdk.AccAddress{s.addr2},
				SendCoins: []*SendCoinsArgs{
					{fromAddr: s.addr3, toAddr: s.addr2, amt: s.coins("10apple")},
				},
			},
			expErr: "failed to re-commit funds after transfer: no hold 4u",
		},
		{
			name: "one in/out with navs",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 4, s.addr3, s.coins("10apple,10banana"))
			},
			markerKeeper: NewMockMarkerKeeper().
				WithGetMarkerAccount(appleMarker).
				WithGetMarkerAccount(bananaMarker),
			req: &exchange.MsgMarketCommitmentSettleRequest{
				Admin:    s.addr1.String(),
				MarketId: 4,
				Inputs:   []exchange.AccountAmount{{Account: s.addr3.String(), Amount: s.coins("10apple,10banana")}},
				Outputs:  []exchange.AccountAmount{{Account: s.addr5.String(), Amount: s.coins("10apple,10banana")}},
				Navs: []exchange.NetAssetPrice{
					{Assets: s.coin("10apple"), Price: s.coin("33cherry")},
					{Assets: s.coin("11apple"), Price: s.coin("700nhash")},
					{Assets: s.coin("12banana"), Price: s.coin("62cherry")},
					{Assets: s.coin("13banana"), Price: s.coin("1500nhash")},
				},
				EventTag: "testtag3",
			},
			expEvents: sdk.Events{
				s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr3.String(), 4, s.coins("10apple,10banana"), "testtag3")),
				s.untypeEvent(exchange.NewEventFundsCommitted(s.addr5.String(), 4, s.coins("10apple,10banana"), "testtag3")),
			},
			expMarkerCalls: MarkerCalls{
				GetMarker: []sdk.AccAddress{appleMarker.GetAddress(), bananaMarker.GetAddress()},
				AddSetNetAssetValues: []*AddSetNetAssetValuesArgs{
					{
						marker: appleMarker,
						netAssetValues: []markertypes.NetAssetValue{
							markertypes.NewNetAssetValue(s.coin("33cherry"), 10),
							markertypes.NewNetAssetValue(s.coin("700nhash"), 11),
						},
						source: navSource(4),
					},
					{
						marker: bananaMarker,
						netAssetValues: []markertypes.NetAssetValue{
							markertypes.NewNetAssetValue(s.coin("62cherry"), 12),
							markertypes.NewNetAssetValue(s.coin("1500nhash"), 13),
						},
						source: navSource(4),
					},
				},
			},
			expHoldCalls: HoldCalls{
				ReleaseHold: []*ReleaseHoldArgs{NewReleaseHoldArgs(s.addr3, s.coins("10apple,10banana"))},
				AddHold:     []*AddHoldArgs{NewAddHoldArgs(s.addr5, s.coins("10apple,10banana"), holdReason(4))},
			},
			expBankCalls: BankCalls{
				BlockedAddr: []sdk.AccAddress{s.addr5},
				SendCoins: []*SendCoinsArgs{
					{fromAddr: s.addr3, toAddr: s.addr5, amt: s.coins("10apple,10banana")},
				},
			},
		},
		{
			name: "one in/out with fees",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 2, s.addr1, s.coins("1cherry"))
				keeper.SetCommitmentAmount(store, 2, s.addr2, s.coins("2cherry"))
				keeper.SetCommitmentAmount(store, 2, s.addr3, s.coins("3cherry"))
				keeper.SetCommitmentAmount(store, 2, s.addr4, s.coins("11apple,4cherry"))
				keeper.SetCommitmentAmount(store, 2, s.addr5, s.coins("5cherry"))
			},
			req: &exchange.MsgMarketCommitmentSettleRequest{
				Admin:    s.addr1.String(),
				MarketId: 2,
				Inputs:   []exchange.AccountAmount{{Account: s.addr4.String(), Amount: s.coins("10apple")}},
				Outputs:  []exchange.AccountAmount{{Account: s.addr2.String(), Amount: s.coins("10apple")}},
				Fees: []exchange.AccountAmount{
					{Account: s.addr1.String(), Amount: s.coins("1cherry")},
					{Account: s.addr2.String(), Amount: s.coins("2cherry")},
					{Account: s.addr3.String(), Amount: s.coins("3cherry")},
					{Account: s.addr4.String(), Amount: s.coins("4cherry")},
					{Account: s.addr5.String(), Amount: s.coins("5cherry")},
				},
				EventTag: "testtag4",
			},
			expEvents: sdk.Events{
				s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr4.String(), 2, s.coins("10apple,4cherry"), "testtag4")),
				s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr1.String(), 2, s.coins("1cherry"), "testtag4")),
				s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr2.String(), 2, s.coins("2cherry"), "testtag4")),
				s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr3.String(), 2, s.coins("3cherry"), "testtag4")),
				s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr5.String(), 2, s.coins("5cherry"), "testtag4")),
				s.untypeEvent(exchange.NewEventFundsCommitted(s.addr2.String(), 2, s.coins("10apple"), "testtag4")),
			},
			expHoldCalls: HoldCalls{
				ReleaseHold: []*ReleaseHoldArgs{
					NewReleaseHoldArgs(s.addr4, s.coins("10apple,4cherry")),
					NewReleaseHoldArgs(s.addr1, s.coins("1cherry")),
					NewReleaseHoldArgs(s.addr2, s.coins("2cherry")),
					NewReleaseHoldArgs(s.addr3, s.coins("3cherry")),
					NewReleaseHoldArgs(s.addr5, s.coins("5cherry")),
				},
				AddHold: []*AddHoldArgs{NewAddHoldArgs(s.addr2, s.coins("10apple"), holdReason(2))},
			},
			expBankCalls: BankCalls{
				BlockedAddr: []sdk.AccAddress{s.addr2, s.marketAddr2},
				SendCoins: []*SendCoinsArgs{
					{fromAddr: s.addr4, toAddr: s.addr2, amt: s.coins("10apple")},
				},
				InputOutputCoins: []*InputOutputCoinsArgs{
					{
						inputs: []banktypes.Input{
							{Address: s.addr1.String(), Coins: s.coins("1cherry")},
							{Address: s.addr2.String(), Coins: s.coins("2cherry")},
							{Address: s.addr3.String(), Coins: s.coins("3cherry")},
							{Address: s.addr4.String(), Coins: s.coins("4cherry")},
							{Address: s.addr5.String(), Coins: s.coins("5cherry")},
						},
						outputs: []banktypes.Output{
							{Address: s.marketAddr2.String(), Coins: s.coins("15cherry")},
						},
					},
				},
			},
		},
		{
			name: "multiple ins/outs/fees",
			setup: func() {
				store := s.getStore()
				keeper.SetCommitmentAmount(store, 2, s.addr1, s.coins("11apple,51banana,1cherry"))
				keeper.SetCommitmentAmount(store, 2, s.addr2, s.coins("12apple,2cherry,72orange"))
				keeper.SetCommitmentAmount(store, 2, s.addr3, s.coins("13apple,3cherry,50orange,41pear"))
				keeper.SetCommitmentAmount(store, 2, s.addr4, s.coins("14apple,4cherry"))
				keeper.SetCommitmentAmount(store, 2, s.addr5, s.coins("5cherry,500raspberry"))
			},
			req: &exchange.MsgMarketCommitmentSettleRequest{
				Admin:    s.k.GetAuthority(),
				MarketId: 2,
				Inputs: []exchange.AccountAmount{
					{Account: s.addr1.String(), Amount: s.coins("10apple,51banana")},
					{Account: s.addr2.String(), Amount: s.coins("10apple,35orange")},
					{Account: s.addr3.String(), Amount: s.coins("13apple,50orange,41pear")},
					{Account: s.addr4.String(), Amount: s.coins("10apple")},
					{Account: s.addr5.String(), Amount: s.coins("500raspberry")},
				},
				Outputs: []exchange.AccountAmount{
					{Account: s.addr1.String(), Amount: s.coins("77orange,65raspberry")},
					{Account: s.addr2.String(), Amount: s.coins("40pear,315raspberry")},
					{Account: s.addr4.String(), Amount: s.coins("50banana,120raspberry,8orange")},
					{Account: s.addr5.String(), Amount: s.coins("43apple,1banana,1pear")},
				},
				Fees: []exchange.AccountAmount{
					{Account: s.addr1.String(), Amount: s.coins("1cherry")},
					{Account: s.addr2.String(), Amount: s.coins("2cherry")},
					{Account: s.addr3.String(), Amount: s.coins("3cherry")},
					{Account: s.addr4.String(), Amount: s.coins("4cherry")},
					{Account: s.addr5.String(), Amount: s.coins("5cherry")},
				},
				EventTag: "messytag",
			},
			expEvents: sdk.Events{
				s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr1.String(), 2, s.coins("10apple,51banana,1cherry"), "messytag")),
				s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr2.String(), 2, s.coins("10apple,2cherry,35orange"), "messytag")),
				s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr3.String(), 2, s.coins("13apple,3cherry,50orange,41pear"), "messytag")),
				s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr4.String(), 2, s.coins("10apple,4cherry"), "messytag")),
				s.untypeEvent(exchange.NewEventCommitmentReleased(s.addr5.String(), 2, s.coins("5cherry,500raspberry"), "messytag")),

				s.untypeEvent(exchange.NewEventFundsCommitted(s.addr1.String(), 2, s.coins("77orange,65raspberry"), "messytag")),
				s.untypeEvent(exchange.NewEventFundsCommitted(s.addr2.String(), 2, s.coins("40pear,315raspberry"), "messytag")),
				s.untypeEvent(exchange.NewEventFundsCommitted(s.addr4.String(), 2, s.coins("50banana,120raspberry,8orange"), "messytag")),
				s.untypeEvent(exchange.NewEventFundsCommitted(s.addr5.String(), 2, s.coins("43apple,1banana,1pear"), "messytag")),
			},
			expHoldCalls: HoldCalls{
				ReleaseHold: []*ReleaseHoldArgs{
					NewReleaseHoldArgs(s.addr1, s.coins("10apple,51banana,1cherry")),
					NewReleaseHoldArgs(s.addr2, s.coins("10apple,2cherry,35orange")),
					NewReleaseHoldArgs(s.addr3, s.coins("13apple,3cherry,50orange,41pear")),
					NewReleaseHoldArgs(s.addr4, s.coins("10apple,4cherry")),
					NewReleaseHoldArgs(s.addr5, s.coins("5cherry,500raspberry")),
				},
				AddHold: []*AddHoldArgs{
					NewAddHoldArgs(s.addr1, s.coins("77orange,65raspberry"), holdReason(2)),
					NewAddHoldArgs(s.addr2, s.coins("40pear,315raspberry"), holdReason(2)),
					NewAddHoldArgs(s.addr4, s.coins("50banana,120raspberry,8orange"), holdReason(2)),
					NewAddHoldArgs(s.addr5, s.coins("43apple,1banana,1pear"), holdReason(2)),
				},
			},
			expBankCalls: BankCalls{
				BlockedAddr: []sdk.AccAddress{s.addr1, s.addr2, s.addr4, s.addr5, s.marketAddr2},
				InputOutputCoins: []*InputOutputCoinsArgs{
					{
						inputs: []banktypes.Input{
							{Address: s.addr2.String(), Coins: s.coins("35orange")},
							{Address: s.addr3.String(), Coins: s.coins("42orange")},
							{Address: s.addr5.String(), Coins: s.coins("65raspberry")},
						},
						outputs: []banktypes.Output{{Address: s.addr1.String(), Coins: s.coins("77orange,65raspberry")}},
					},
					{
						inputs: []banktypes.Input{
							{Address: s.addr3.String(), Coins: s.coins("40pear")},
							{Address: s.addr5.String(), Coins: s.coins("315raspberry")},
						},
						outputs: []banktypes.Output{{Address: s.addr2.String(), Coins: s.coins("40pear,315raspberry")}},
					},
					{
						inputs: []banktypes.Input{
							{Address: s.addr1.String(), Coins: s.coins("50banana")},
							{Address: s.addr3.String(), Coins: s.coins("8orange")},
							{Address: s.addr5.String(), Coins: s.coins("120raspberry")},
						},
						outputs: []banktypes.Output{{Address: s.addr4.String(), Coins: s.coins("50banana,120raspberry,8orange")}},
					},
					{
						inputs: []banktypes.Input{
							{Address: s.addr1.String(), Coins: s.coins("10apple,1banana")},
							{Address: s.addr2.String(), Coins: s.coins("10apple")},
							{Address: s.addr3.String(), Coins: s.coins("13apple,1pear")},
							{Address: s.addr4.String(), Coins: s.coins("10apple")},
						},
						outputs: []banktypes.Output{{Address: s.addr5.String(), Coins: s.coins("43apple,1banana,1pear")}},
					},
					{
						inputs: []banktypes.Input{
							{Address: s.addr1.String(), Coins: s.coins("1cherry")},
							{Address: s.addr2.String(), Coins: s.coins("2cherry")},
							{Address: s.addr3.String(), Coins: s.coins("3cherry")},
							{Address: s.addr4.String(), Coins: s.coins("4cherry")},
							{Address: s.addr5.String(), Coins: s.coins("5cherry")},
						},
						outputs: []banktypes.Output{{Address: s.marketAddr2.String(), Coins: s.coins("15cherry")}},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			if tc.markerKeeper == nil {
				tc.markerKeeper = NewMockMarkerKeeper()
			}
			if tc.bankKeeper == nil {
				tc.bankKeeper = NewMockBankKeeper()
			}
			if tc.holdKeeper == nil {
				tc.holdKeeper = NewMockHoldKeeper()
			}

			admin, aErr := sdk.AccAddressFromBech32(tc.req.Admin)
			s.Require().NoError(aErr, "AccAddressFromBech32(tc.req.Admin)")
			for _, exp := range tc.expBankCalls.SendCoins {
				exp.ctxHasQuarantineBypass = true
				exp.ctxTransferAgent = admin
			}
			for _, exp := range tc.expBankCalls.InputOutputCoins {
				exp.ctxHasQuarantineBypass = true
				exp.ctxTransferAgent = admin
			}

			kpr := s.k.
				WithMarkerKeeper(tc.markerKeeper).
				WithBankKeeper(tc.bankKeeper).
				WithHoldKeeper(tc.holdKeeper)
			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = kpr.SettleCommitments(ctx, tc.req)
			}
			s.Require().NotPanics(testFunc, "SettleCommitments")
			s.assertErrorValue(err, tc.expErr, "SettleCommitments error")

			actEvents := em.Events()
			s.assertEqualEvents(tc.expEvents, actEvents, "events emitted during SettleCommitments")
			s.assertMarkerKeeperCalls(tc.markerKeeper, tc.expMarkerCalls, "SettleCommitments")
			s.assertBankKeeperCalls(tc.bankKeeper, tc.expBankCalls, "SettleCommitments")
			s.assertHoldKeeperCalls(tc.holdKeeper, tc.expHoldCalls, "SettleCommitments")
		})
	}
}
