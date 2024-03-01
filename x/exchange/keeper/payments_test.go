package keeper_test

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/keeper"
)

// newTestPayment creates a new Payment using the provided info.
func (s *TestSuite) newTestPayment(source sdk.AccAddress, sourceAmount string, target sdk.AccAddress, targetAmount string, externalID string) *exchange.Payment {
	s.T().Helper()
	return &exchange.Payment{
		Source:       source.String(),
		SourceAmount: s.coins(sourceAmount),
		Target:       target.String(),
		TargetAmount: s.coins(targetAmount),
		ExternalId:   externalID,
	}
}

func (s *TestSuite) TestGetPayment() {
	sourceHasTwoPayments1 := s.newTestPayment(s.longAddr2, "22strawberry", s.addr3, "12tomato", "l2-3-2")
	sourceHasTwoPayments2 := s.newTestPayment(s.longAddr2, "44strawberry", s.addr3, "14tomato", "l2-3-4")
	sourceHasTwoPaymentsSetup := func() {
		var payments []*exchange.Payment
		for _, i := range []int{1, 2, 3, 4, 5} {
			sa1 := fmt.Sprintf("%d0strawberry", i)
			ta1 := fmt.Sprintf("%dtomato", i)
			sa3 := fmt.Sprintf("%d%d%dstrawberry", i, i, i)
			ta3 := fmt.Sprintf("%d7tomato", i)
			payments = append(payments,
				s.newTestPayment(s.longAddr1, sa1, s.addr3, ta1, fmt.Sprintf("l1-3-%d", i)),
				s.newTestPayment(s.longAddr3, sa3, s.addr3, ta3, fmt.Sprintf("l1-3-%d", i)),
			)
		}
		payments = append(payments, sourceHasTwoPayments1, sourceHasTwoPayments2)
		s.requireSetPaymentsInStore(payments...)
	}

	tests := []struct {
		name       string
		setup      func()
		source     sdk.AccAddress
		externalID string
		expPayment *exchange.Payment
		expErr     string
	}{
		{
			name:       "no entry",
			source:     s.addr1,
			externalID: "oops",
			expPayment: nil,
			expErr:     "",
		},
		{
			name: "empty entry",
			setup: func() {
				key := keeper.MakeKeyPayment(s.addr2, "nope")
				s.k.GetStore(s.ctx).Set(key, []byte{})
			},
			source:     s.addr2,
			externalID: "nope",
			expPayment: nil,
			expErr:     "",
		},
		{
			name: "invalid entry",
			setup: func() {
				key := keeper.MakeKeyPayment(s.addr3, "bang")
				s.getStore().Set(key, []byte{'x'})
			},
			source:     s.addr3,
			externalID: "bang",
			expErr:     "failed to unmarshal payment: unexpected EOF",
		},
		{
			name: "fully filled payment",
			setup: func() {
				s.requireSetPaymentsInStore(s.newTestPayment(s.addr4, "12strawberry", s.addr5, "55tangerine", "buddy"))
			},
			source:     s.addr4,
			externalID: "buddy",
			expPayment: s.newTestPayment(s.addr4, "12strawberry", s.addr5, "55tangerine", "buddy"),
		},
		{
			name: "no external id",
			setup: func() {
				s.requireSetPaymentsInStore(s.newTestPayment(s.addr2, "8starfruit", s.addr4, "41tangerine", ""))
			},
			source:     s.addr2,
			externalID: "",
			expPayment: s.newTestPayment(s.addr2, "8starfruit", s.addr4, "41tangerine", ""),
		},
		{
			name: "max length external id",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.longAddr1, "71strawberry", s.addr5, "5001tomato",
						strings.Repeat("p", exchange.MaxExternalIDLength)),
				)
			},
			source:     s.longAddr1,
			externalID: strings.Repeat("p", exchange.MaxExternalIDLength),
			expPayment: s.newTestPayment(s.longAddr1, "71strawberry", s.addr5, "5001tomato",
				strings.Repeat("p", exchange.MaxExternalIDLength)),
		},
		{
			name:       "source has two payments: external id less than both",
			setup:      sourceHasTwoPaymentsSetup,
			source:     s.longAddr2,
			externalID: "l2-3-1",
			expPayment: nil,
			expErr:     "",
		},
		{
			name:       "source has two payments: external id is first",
			setup:      sourceHasTwoPaymentsSetup,
			source:     s.longAddr2,
			externalID: "l2-3-2",
			expPayment: sourceHasTwoPayments1,
		},
		{
			name:       "source has two payments: external id between",
			setup:      sourceHasTwoPaymentsSetup,
			source:     s.longAddr2,
			externalID: "l2-3-3",
			expPayment: nil,
			expErr:     "",
		},
		{
			name:       "source has two payments: external id is second",
			setup:      sourceHasTwoPaymentsSetup,
			source:     s.longAddr2,
			externalID: "l2-3-4",
			expPayment: sourceHasTwoPayments2,
		},
		{
			name:       "source has two payments: external id greater than both",
			setup:      sourceHasTwoPaymentsSetup,
			source:     s.longAddr2,
			externalID: "l2-3-5",
			expPayment: nil,
			expErr:     "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			var payment *exchange.Payment
			var err error
			testFunc := func() {
				payment, err = s.k.GetPayment(s.ctx, tc.source, tc.externalID)
			}
			s.Require().NotPanics(testFunc, "GetPayment(%s, %q)", s.getAddrName(tc.source), tc.externalID)
			s.assertErrorValue(err, tc.expErr, "GetPayment(%s, %q) error", s.getAddrName(tc.source), tc.externalID)
			s.assertEqualPayment(tc.expPayment, payment, "GetPayment(%s, %q) result", s.getAddrName(tc.source), tc.externalID)
		})
	}
}

func (s *TestSuite) TestCreatePayment() {
	tests := []struct {
		name       string
		setup      func()
		holdKeeper *MockHoldKeeper
		payment    *exchange.Payment
		expPayment *exchange.Payment // Set to payment when expStored is true.
		expErr     string
		expStored  bool
		expIndex   bool
		expAddHold bool
		expEvent   bool
	}{
		{
			name:    "nil payment",
			payment: nil,
			expErr:  "cannot create nil payment",
		},
		{
			name:    "no source",
			payment: s.newTestPayment(nil, "13strawberry", s.addr3, "", "badbad"),
			expErr:  "cannot create invalid payment: invalid source \"\": empty address string is not allowed",
		},
		{
			name: "payment already exists",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.addr2, "10strawberry", s.longAddr3, "2tomato", "do-not-reuse-this"),
				)
			},
			payment:    s.newTestPayment(s.addr2, "10000strawberry", s.addr5, "", "do-not-reuse-this"),
			expPayment: s.newTestPayment(s.addr2, "10strawberry", s.longAddr3, "2tomato", "do-not-reuse-this"),
			expErr: "failed to create payment: a payment already exists with source " +
				s.addr2.String() + " and external id \"do-not-reuse-this\"",
		},
		{
			name:       "error adding hold",
			holdKeeper: NewMockHoldKeeper().WithAddHoldResults("you know you can't do that"),
			payment:    s.newTestPayment(s.longAddr3, "3starfruit", s.longAddr2, "1tangerine", "nopenope"),
			expErr:     "error placing hold on payment source: you know you can't do that",
			expStored:  true,
			expIndex:   true,
			expAddHold: true,
		},
		{
			name:       "fully filled",
			payment:    s.newTestPayment(s.longAddr3, "88starfruit", s.longAddr1, "12tangerine,8tomato", "long-addrs"),
			expStored:  true,
			expIndex:   true,
			expAddHold: true,
			expEvent:   true,
		},
		{
			name:       "no external id",
			payment:    s.newTestPayment(s.addr2, "12strawberry,4starfruit", s.addr3, "", ""),
			expStored:  true,
			expIndex:   true,
			expAddHold: true,
			expEvent:   true,
		},
		{
			name:       "no target",
			payment:    s.newTestPayment(s.longAddr2, "", nil, "3tomato", "soon"),
			expStored:  true,
			expIndex:   false,
			expAddHold: true,
			expEvent:   true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()

			if tc.holdKeeper == nil {
				tc.holdKeeper = NewMockHoldKeeper()
			}

			if tc.expStored {
				tc.expPayment = tc.payment
			}

			var expIndexKey []byte
			if tc.payment != nil {
				source, sErr := sdk.AccAddressFromBech32(tc.payment.Source)
				target, tErr := sdk.AccAddressFromBech32(tc.payment.Target)
				if sErr == nil && tErr == nil {
					expIndexKey = keeper.MakeIndexKeyTargetToPayment(target, source, tc.payment.ExternalId)
				}
			}

			var expEvents sdk.Events
			if tc.expEvent {
				s.Require().NotNil(tc.payment, "tc.payment cannot be nil when tc.expEvent = true")
				expEvents = sdk.Events{s.untypeEvent(exchange.NewEventPaymentCreated(tc.payment))}
			}

			var expHoldCalls HoldCalls
			if tc.expAddHold {
				s.Require().NotNil(tc.payment, "tc.payment cannot be nil when tc.expAddHold = true")
				expHoldCalls.AddHold = []*AddHoldArgs{
					{
						addr:   s.requireAccAddressFromBech32(tc.payment.Source, "valid payment source required when tc.expAddHold = true"),
						funds:  tc.payment.SourceAmount,
						reason: fmt.Sprintf("x/exchange: payment %q", tc.payment.ExternalId),
					},
				}
			}

			if tc.setup != nil {
				tc.setup()
			}

			kpr := s.k.WithHoldKeeper(tc.holdKeeper)
			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = kpr.CreatePayment(ctx, tc.payment)
			}
			s.Require().NotPanics(testFunc, "CreatePayment(%s)", tc.payment)
			s.assertErrorValue(err, tc.expErr, "CreatePayment(%s) error", tc.payment)
			s.assertHoldKeeperCalls(tc.holdKeeper, expHoldCalls, "CreatePayment(%s) hold keeper calls", tc.payment)

			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "CreatePayment(%s) events", tc.payment)

			var actPayment *exchange.Payment
			if tc.payment != nil && len(tc.payment.Source) > 0 {
				source, sErr := sdk.AccAddressFromBech32(tc.payment.Source)
				if sErr == nil {
					actPayment, _ = s.k.GetPayment(s.ctx, source, tc.payment.ExternalId)
				}
			}
			s.assertEqualPayment(tc.expPayment, actPayment, "payment read from state after CreatePayment(%s)", tc.payment)

			if len(expIndexKey) > 0 {
				store := s.getStore()
				hasIndex := store.Has(expIndexKey)
				if s.Assert().Equal(tc.expIndex, hasIndex, "store.Has(the target to payment index key)") && tc.expIndex {
					indexValue := store.Get(expIndexKey)
					s.Assert().Empty(indexValue, "the target to payment index value")
				}
			}
		})
	}
}

func (s *TestSuite) TestAcceptPayment() {
	getPaymentStoreValue := func(payment *exchange.Payment) []byte {
		s.Require().NotNil(payment, "cannot create payment store value from nil payment")
		rv, err := s.k.GetCodec().Marshal(payment)
		s.Require().NoError(err, "Marshal(%s)", payment)
		return rv
	}
	setCustomPaymentInStore := func(key []byte, payment *exchange.Payment) {
		value := getPaymentStoreValue(payment)
		store := s.getStore()
		s.Require().NotPanics(func() {
			store.Set(key, value)
		}, "store.Set(%v, %s)", key, payment)
	}
	fullPaymentSource := s.longAddr1
	fullPaymentTarget := s.addr2
	fullPayment := s.newTestPayment(fullPaymentSource, "2starfruit,33strawberry", fullPaymentTarget, "8tangerine,3tomato", "just-some-id")
	fullPaymentKey := keeper.MakeKeyPayment(fullPaymentSource, fullPayment.ExternalId)

	tests := []struct {
		name           string
		setup          func()
		holdKeeper     *MockHoldKeeper
		bankKeeper     *MockBankKeeper
		payment        *exchange.Payment
		expErr         string
		expDeleted     bool
		expReleaseHold bool
		expBankCalls   BankCalls
		expEvent       bool
	}{
		{
			name:    "nil payment",
			payment: nil,
			expErr:  "cannot accept nil payment",
		},
		{
			name:    "invalid payment",
			payment: s.newTestPayment(nil, "8strawberry", s.addr1, "", ""),
			expErr:  "cannot accept invalid payment: invalid source \"\": empty address string is not allowed",
		},
		{
			name:    "no target",
			payment: s.newTestPayment(s.addr1, "", nil, "8tomato", ""),
			expErr:  "cannot accept a payment without a target",
		},
		{
			name: "payment does not exist",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.addr1, "1strawberry", s.addr5, "", "b"),
					s.newTestPayment(s.addr2, "2strawberry", s.addr5, "", "a"),
					s.newTestPayment(s.addr2, "4strawberry", s.addr5, "", "c"),
					s.newTestPayment(s.addr3, "5strawberry", s.addr5, "", "b"),
				)
			},
			payment: s.newTestPayment(s.addr2, "3strawberry", s.addr2, "", "b"),
			expErr:  "no payment found with source " + s.addr2.String() + " and external id \"b\"",
		},
		{
			name: "wrong source",
			setup: func() {
				setCustomPaymentInStore(fullPaymentKey,
					&exchange.Payment{
						Source:       s.addr5.String(),
						SourceAmount: fullPayment.SourceAmount,
						Target:       fullPayment.Target,
						TargetAmount: fullPayment.TargetAmount,
						ExternalId:   fullPayment.ExternalId,
					},
				)
			},
			payment: fullPayment,
			expErr:  "provided source " + fullPayment.Source + " does not equal existing source " + s.addr5.String(),
		},
		{
			name: "wrong source amount",
			setup: func() {
				s.requireSetPaymentsInStore(fullPayment)
			},
			payment: &exchange.Payment{
				Source:       fullPayment.Source,
				SourceAmount: s.coins("88starfruit"),
				Target:       fullPayment.Target,
				TargetAmount: fullPayment.TargetAmount,
				ExternalId:   fullPayment.ExternalId,
			},
			expErr: "provided source amount \"88starfruit\" does not equal existing source amount \"" +
				fullPayment.SourceAmount.String() + "\"",
		},
		{
			name: "existing does not have a target",
			setup: func() {
				s.requireSetPaymentsInStore(&exchange.Payment{
					Source:       fullPayment.Source,
					SourceAmount: fullPayment.SourceAmount,
					Target:       "",
					TargetAmount: fullPayment.TargetAmount,
					ExternalId:   fullPayment.ExternalId,
				})
			},
			payment: fullPayment,
			expErr:  "provided target " + fullPayment.Target + " does not equal existing target ",
		},
		{
			name: "wrong target",
			setup: func() {
				s.requireSetPaymentsInStore(fullPayment)
			},
			payment: &exchange.Payment{
				Source:       fullPayment.Source,
				SourceAmount: fullPayment.SourceAmount,
				Target:       s.longAddr3.String(),
				TargetAmount: fullPayment.TargetAmount,
				ExternalId:   fullPayment.ExternalId,
			},
			expErr: "provided target " + s.longAddr3.String() + " does not equal existing target " + fullPayment.Target,
		},
		{
			name: "wrong target amount",
			setup: func() {
				s.requireSetPaymentsInStore(fullPayment)
			},
			payment: &exchange.Payment{
				Source:       fullPayment.Source,
				SourceAmount: fullPayment.SourceAmount,
				Target:       fullPayment.Target,
				TargetAmount: nil,
				ExternalId:   fullPayment.ExternalId,
			},
			expErr: "provided target amount \"\" does not equal existing target amount \"" +
				fullPayment.TargetAmount.String() + "\"",
		},
		{
			name: "wrong external id",
			setup: func() {
				setCustomPaymentInStore(fullPaymentKey,
					&exchange.Payment{
						Source:       fullPayment.Source,
						SourceAmount: fullPayment.SourceAmount,
						Target:       fullPayment.Target,
						TargetAmount: fullPayment.TargetAmount,
						ExternalId:   "noway",
					},
				)
			},
			payment: fullPayment,
			expErr:  "provided external id \"" + fullPayment.ExternalId + "\" does not equal existing external id \"noway\"",
		},
		{
			name: "error releasing hold",
			setup: func() {
				s.requireSetPaymentsInStore(fullPayment)
			},
			holdKeeper:     NewMockHoldKeeper().WithReleaseHoldResults("just keep hodling on"),
			payment:        fullPayment,
			expErr:         "error releasing hold on payment source: just keep hodling on",
			expDeleted:     true,
			expReleaseHold: true,
		},
		{
			name: "error sending source funds",
			setup: func() {
				s.requireSetPaymentsInStore(fullPayment)
			},
			bankKeeper: NewMockBankKeeper().WithSendCoinsResults("first injected send error"),
			payment:    fullPayment,
			expErr: "error sending \"" + fullPayment.SourceAmount.String() + "\" from source " +
				fullPayment.Source + " to target " + fullPayment.Target + ": first injected send error",
			expDeleted:     true,
			expReleaseHold: true,
			expBankCalls: BankCalls{SendCoins: []*SendCoinsArgs{
				{fromAddr: fullPaymentSource, toAddr: fullPaymentTarget, amt: fullPayment.SourceAmount},
			}},
		},
		{
			name: "error sending target funds",
			setup: func() {
				s.requireSetPaymentsInStore(fullPayment)
			},
			bankKeeper: NewMockBankKeeper().WithSendCoinsResults("", "injected error for second send"),
			payment:    fullPayment,
			expErr: "error sending \"" + fullPayment.TargetAmount.String() + "\" from target " +
				fullPayment.Target + " to source " + fullPayment.Source + ": injected error for second send",
			expDeleted:     true,
			expReleaseHold: true,
			expBankCalls: BankCalls{SendCoins: []*SendCoinsArgs{
				{fromAddr: fullPaymentSource, toAddr: fullPaymentTarget, amt: fullPayment.SourceAmount},
				{fromAddr: fullPaymentTarget, toAddr: fullPaymentSource, amt: fullPayment.TargetAmount},
			}},
		},
		{
			name: "no source funds",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.addr3, "", s.addr1, "8tomato", "a-payment-request"),
				)
			},
			payment:        s.newTestPayment(s.addr3, "", s.addr1, "8tomato", "a-payment-request"),
			expDeleted:     true,
			expReleaseHold: true,
			expBankCalls: BankCalls{SendCoins: []*SendCoinsArgs{
				{fromAddr: s.addr1, toAddr: s.addr3, amt: s.coins("8tomato")},
			}},
			expEvent: true,
		},
		{
			name: "no target funds",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.addr5, "63strawberry", s.longAddr1, "", "a-peer-to-peer-payment"),
				)
			},
			payment:        s.newTestPayment(s.addr5, "63strawberry", s.longAddr1, "", "a-peer-to-peer-payment"),
			expDeleted:     true,
			expReleaseHold: true,
			expBankCalls: BankCalls{SendCoins: []*SendCoinsArgs{
				{fromAddr: s.addr5, toAddr: s.longAddr1, amt: s.coins("63strawberry")},
			}},
			expEvent: true,
		},
		{
			name: "funds going both ways",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.addr4, "3strawberry", s.addr3, "5000tangerine", "a-trade"),
				)
			},
			payment:        s.newTestPayment(s.addr4, "3strawberry", s.addr3, "5000tangerine", "a-trade"),
			expDeleted:     true,
			expReleaseHold: true,
			expBankCalls: BankCalls{SendCoins: []*SendCoinsArgs{
				{fromAddr: s.addr4, toAddr: s.addr3, amt: s.coins("3strawberry")},
				{fromAddr: s.addr3, toAddr: s.addr4, amt: s.coins("5000tangerine")},
			}},
			expEvent: true,
		},
		{
			name: "no external id",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.longAddr3, "66starfruit", s.addr1, "6tomato", ""),
				)
			},
			payment:        s.newTestPayment(s.longAddr3, "66starfruit", s.addr1, "6tomato", ""),
			expDeleted:     true,
			expReleaseHold: true,
			expBankCalls: BankCalls{SendCoins: []*SendCoinsArgs{
				{fromAddr: s.longAddr3, toAddr: s.addr1, amt: s.coins("66starfruit")},
				{fromAddr: s.addr1, toAddr: s.longAddr3, amt: s.coins("6tomato")},
			}},
			expEvent: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()

			if tc.holdKeeper == nil {
				tc.holdKeeper = NewMockHoldKeeper()
			}
			if tc.bankKeeper == nil {
				tc.bankKeeper = NewMockBankKeeper()
			}

			var expHoldCalls HoldCalls
			if tc.expReleaseHold {
				s.Require().NotNil(tc.payment, "tc.payment cannot be nil when tc.expReleaseHold = true")
				expHoldCalls.ReleaseHold = []*ReleaseHoldArgs{{
					addr:  s.requireAccAddressFromBech32(tc.payment.Source, "valid payment source required when tc.expReleaseHold = true"),
					funds: tc.payment.SourceAmount,
				}}
			}

			for i := range tc.expBankCalls.SendCoins {
				tc.expBankCalls.SendCoins[i].ctxHasQuarantineBypass = true
			}

			var expEvents sdk.Events
			if tc.expEvent {
				s.Require().NotNil(tc.payment, "tc.payment cannot be nil when tc.expEvent = true")
				expEvents = sdk.Events{s.untypeEvent(exchange.NewEventPaymentAccepted(tc.payment))}
			}

			if tc.setup != nil {
				tc.setup()
			}

			kpr := s.k.WithHoldKeeper(tc.holdKeeper).WithBankKeeper(tc.bankKeeper)
			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = kpr.AcceptPayment(ctx, tc.payment)
			}
			s.Require().NotPanics(testFunc, "AcceptPayment(%s)", tc.payment)
			s.assertErrorValue(err, tc.expErr, "AcceptPayment(%s) error", tc.payment)
			s.assertHoldKeeperCalls(tc.holdKeeper, expHoldCalls, "AcceptPayment(%s) hold keeper calls", tc.payment)
			s.assertBankKeeperCalls(tc.bankKeeper, tc.expBankCalls, "AcceptPayment(%s) bank keeper calls", tc.payment)

			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "AcceptPayment(%s) events", tc.payment)

			if tc.expDeleted {
				var actPayment *exchange.Payment
				var indexKey []byte
				if tc.payment != nil {
					source, sErr := sdk.AccAddressFromBech32(tc.payment.Source)
					if sErr == nil {
						actPayment, _ = s.k.GetPayment(s.ctx, source, tc.payment.ExternalId)
					}
					target, tErr := sdk.AccAddressFromBech32(tc.payment.Target)
					if sErr == nil && tErr == nil {
						indexKey = keeper.MakeIndexKeyTargetToPayment(target, source, tc.payment.ExternalId)
					}
				}
				s.Assert().Nil(actPayment, "GetPayment after AcceptPayment(%s)", tc.payment)
				hasIndex := s.getStore().Has(indexKey)
				s.Assert().False(hasIndex, "store.Has(target to payment index key) after AcceptPayment(%s)", tc.payment)
			}
		})
	}
}

func (s *TestSuite) TestRejectPayment() {
	tests := []struct {
		name         string
		setup        func()
		holdKeeper   *MockHoldKeeper
		target       sdk.AccAddress
		source       sdk.AccAddress
		externalID   string
		expErr       string
		expDeleted   bool
		expHoldCalls HoldCalls
		expEvent     *exchange.EventPaymentRejected
	}{
		{
			name:       "no target",
			target:     nil,
			source:     s.addr5,
			externalID: "banananas",
			expErr:     "a target is required in order to reject payment",
		},
		{
			name:       "no source",
			target:     s.addr4,
			source:     nil,
			externalID: "pinetree",
			expErr:     "a source is required in order to reject payment",
		},
		{
			name: "no payment",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.addr1, "1strawberry", s.addr5, "", "b"),
					s.newTestPayment(s.addr2, "2strawberry", s.addr5, "", "a"),
					s.newTestPayment(s.addr2, "4strawberry", s.addr5, "", "c"),
					s.newTestPayment(s.addr3, "5strawberry", s.addr5, "", "b"),
				)
			},
			target:     s.addr5,
			source:     s.addr2,
			externalID: "b",
			expErr:     "no payment found with source " + s.addr2.String() + " and external id \"b\"",
		},
		{
			name: "payment does not have a target yet",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.longAddr1, "888strawberry", nil, "12tangerine", "pants"),
				)
			},
			target:     s.longAddr3,
			source:     s.longAddr1,
			externalID: "pants",
			expErr:     "cannot reject a payment that does not have a target",
		},
		{
			name: "wrong target",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.longAddr1, "8strawberry", s.longAddr2, "55tangerine", "pants"),
				)
			},
			target:     s.longAddr3,
			source:     s.longAddr1,
			externalID: "pants",
			expErr:     "target " + s.longAddr3.String() + " cannot reject payment with target " + s.longAddr2.String(),
		},
		{
			name: "error releasing hold",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.longAddr1, "1strawberry", s.longAddr3, "352tangerine", "nonono"),
				)
			},
			holdKeeper:   NewMockHoldKeeper().WithReleaseHoldResults("oops, can't do that"),
			target:       s.longAddr3,
			source:       s.longAddr1,
			externalID:   "nonono",
			expErr:       "error releasing hold on payment source: oops, can't do that",
			expDeleted:   true,
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.longAddr1, funds: s.coins("1strawberry")}}},
		},
		{
			name: "no source funds",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.addr3, "", s.addr1, "41tomato", "gimmiegimmie"),
				)
			},
			target:       s.addr1,
			source:       s.addr3,
			externalID:   "gimmiegimmie",
			expDeleted:   true,
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr3, funds: nil}}},
			expEvent: exchange.NewEventPaymentRejected(
				s.newTestPayment(s.addr3, "", s.addr1, "41tomato", "gimmiegimmie")),
		},
		{
			name: "no target funds",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.addr2, "81starfruit", s.addr4, "", "all4u"),
				)
			},
			target:       s.addr4,
			source:       s.addr2,
			externalID:   "all4u",
			expDeleted:   true,
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr2, funds: s.coins("81starfruit")}}},
			expEvent: exchange.NewEventPaymentRejected(
				s.newTestPayment(s.addr2, "81starfruit", s.addr4, "", "all4u")),
		},
		{
			name: "with external id",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.longAddr1, "497strawberry", s.addr5, "13tangerine,12tomato", "I am an external id."),
				)
			},
			target:       s.addr5,
			source:       s.longAddr1,
			externalID:   "I am an external id.",
			expDeleted:   true,
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.longAddr1, funds: s.coins("497strawberry")}}},
			expEvent: exchange.NewEventPaymentRejected(
				s.newTestPayment(s.longAddr1, "497strawberry", s.addr5, "13tangerine,12tomato", "I am an external id.")),
		},
		{
			name: "without external id",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.addr2, "18starfruit,371strawberry", s.longAddr1, "945tomato", ""),
				)
			},
			target:       s.longAddr1,
			source:       s.addr2,
			externalID:   "",
			expDeleted:   true,
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr2, funds: s.coins("18starfruit,371strawberry")}}},
			expEvent: exchange.NewEventPaymentRejected(
				s.newTestPayment(s.addr2, "18starfruit,371strawberry", s.longAddr1, "945tomato", "")),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()

			if tc.holdKeeper == nil {
				tc.holdKeeper = NewMockHoldKeeper()
			}

			var expEvents sdk.Events
			if tc.expEvent != nil {
				expEvents = append(expEvents, s.untypeEvent(tc.expEvent))
			}

			if tc.setup != nil {
				tc.setup()
			}

			kpr := s.k.WithHoldKeeper(tc.holdKeeper)
			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = kpr.RejectPayment(ctx, tc.target, tc.source, tc.externalID)
			}
			sourceName := s.getAddrName(tc.source)
			targetName := s.getAddrName(tc.target)
			s.Require().NotPanics(testFunc, "RejectPayment(%s, %s, %q)", targetName, sourceName, tc.externalID)
			s.assertErrorValue(err, tc.expErr, "RejectPayment(%s, %s, %q) error", targetName, sourceName, tc.externalID)
			s.assertHoldKeeperCalls(tc.holdKeeper, tc.expHoldCalls, "RejectPayment(%s, %s, %q) hold calls",
				targetName, sourceName, tc.externalID)

			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "RejectPayment(%s, %s, %q) events",
				targetName, sourceName, tc.externalID)

			var actPayment *exchange.Payment
			var hasIndex bool
			if len(tc.source) > 0 {
				actPayment, _ = s.k.GetPayment(s.ctx, tc.source, tc.externalID)
				if len(tc.target) > 0 {
					indexKey := keeper.MakeIndexKeyTargetToPayment(tc.target, tc.source, tc.externalID)
					hasIndex = s.getStore().Has(indexKey)
				}
			}
			if tc.expDeleted {
				s.Assert().Nil(actPayment, "GetPayment after RejectPayment(%s, %s, %q)",
					targetName, sourceName, tc.externalID)
				s.Assert().False(hasIndex, "store.Has(target to payment index key) after RejectPayment(%s, %s, %q)",
					targetName, sourceName, tc.externalID)
			}
		})
	}
}

// TODO[1703]: func (s *TestSuite) TestRejectPayments()

// TODO[1703]: func (s *TestSuite) TestCancelPayments()

// TODO[1703]: func (s *TestSuite) TestUpdatePaymentTarget()

// TODO[1703]: func (s *TestSuite) TestGetPaymentsForTargetAndSource()

// TODO[1703]: func (s *TestSuite) TestIteratePayments()

// TODO[1703]: func (s *TestSuite) TestCalculatePaymentFeess()
