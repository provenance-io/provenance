package keeper_test

import (
	"bytes"
	"fmt"
	"sort"
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

// getAllPayments gets all the payments currently in state.
func (s *TestSuite) getAllPayments() []*exchange.Payment {
	var rv []*exchange.Payment
	s.k.IteratePayments(s.ctx, func(payment *exchange.Payment) bool {
		rv = append(rv, payment)
		return false
	})
	return rv
}

// getAllTargetToPaymentIndexEntries gets all the target to payment index keys currently in state.
func (s *TestSuite) getAllTargetToPaymentIndexEntries() [][]byte {
	var rv [][]byte
	keyPrefix := []byte{keeper.KeyTypeTargetToPaymentIndex}
	store := s.getStore()
	keeper.Iterate(store, keyPrefix, func(keySuffix, _ []byte) bool {
		key := concatBz(keyPrefix, keySuffix)
		rv = append(rv, key)
		return false
	})
	return rv
}

// assertTargetToPaymentIndexEntriesMatchPayments gets all the payments and target to payment index entries from state
// and makes sure that they're all as they should be.
func (s *TestSuite) assertTargetToPaymentIndexEntriesMatchPayments() bool {
	s.T().Helper()
	payments := s.getAllPayments()
	var expKeys [][]byte
	if len(payments) > 0 {
		for _, payment := range payments {
			target, _ := sdk.AccAddressFromBech32(payment.Target)
			source, _ := sdk.AccAddressFromBech32(payment.Source)
			if len(target) > 0 && len(source) > 0 {
				key := keeper.MakeIndexKeyTargetToPayment(target, source, payment.ExternalId)
				expKeys = append(expKeys, key)
			}
		}
	}
	sort.Slice(expKeys, func(i, j int) bool {
		return bytes.Compare(expKeys[i], expKeys[j]) < 0
	})

	actKeys := s.getAllTargetToPaymentIndexEntries()

	keyStringer := func(key []byte) string {
		target, source, externalID, err := keeper.ParseIndexKeyTargetToPayment(key)
		if err == nil {
			return fmt.Sprintf("%v", key)
		}
		return fmt.Sprintf("%s %s %q", s.getAddrName(target), s.getAddrName(source), externalID)
	}

	return assertEqualSlice(s, expKeys, actKeys, keyStringer, "target to payment index entries")
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

			s.assertTargetToPaymentIndexEntriesMatchPayments()
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
		skipIndCheck   bool
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
			payment:      fullPayment,
			expErr:       "provided source " + fullPayment.Source + " does not equal existing source " + s.addr5.String(),
			skipIndCheck: true,
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
			payment:      fullPayment,
			expErr:       "provided external id \"" + fullPayment.ExternalId + "\" does not equal existing external id \"noway\"",
			skipIndCheck: true,
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

			if !tc.skipIndCheck {
				s.assertTargetToPaymentIndexEntriesMatchPayments()
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

			s.assertTargetToPaymentIndexEntriesMatchPayments()
		})
	}
}

func (s *TestSuite) TestRejectPayments() {
	type paymentKey struct {
		source     sdk.AccAddress
		externalID string
	}
	newPKey := func(source sdk.AccAddress, externalID string) paymentKey {
		return paymentKey{source: source, externalID: externalID}
	}

	tests := []struct {
		name         string
		setup        func()
		holdKeeper   *MockHoldKeeper
		target       sdk.AccAddress
		sources      []sdk.AccAddress
		expErr       string
		expHoldCalls HoldCalls
		expEvents    []*exchange.EventPaymentRejected
		expDeleted   []paymentKey
		expRemain    []paymentKey
	}{
		{
			name:    "nil target",
			target:  nil,
			sources: []sdk.AccAddress{s.addr1},
			expErr:  "a target is required in order to reject payments",
		},
		{
			name:    "empty target",
			target:  sdk.AccAddress{},
			sources: []sdk.AccAddress{s.addr1},
			expErr:  "a target is required in order to reject payments",
		},
		{
			name:    "nil sources",
			target:  s.addr4,
			sources: nil,
			expErr:  "at least one source is required",
		},
		{
			name:    "empty sources",
			target:  s.addr3,
			sources: []sdk.AccAddress{},
			expErr:  "at least one source is required",
		},
		{
			name: "one source: zero payments",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.addr2, "", s.addr1, "2tomato", "a"),
					s.newTestPayment(s.addr4, "", s.addr1, "4tomato", "c"),
				)
			},
			holdKeeper: nil,
			target:     s.addr1,
			sources:    []sdk.AccAddress{s.addr3},
			expErr:     "source " + s.addr3.String() + " does not have any payments for target " + s.addr1.String(),
			expRemain:  []paymentKey{newPKey(s.addr2, "a"), newPKey(s.addr4, "c")},
		},
		{
			name: "error releasing a hold",
			setup: func() {
				s.requireSetPaymentsInStore(s.newTestPayment(s.addr4, "13strawberry", s.longAddr1, "8tangerine", "anid"))
			},
			holdKeeper:   NewMockHoldKeeper().WithReleaseHoldResults("stop right there"),
			target:       s.longAddr1,
			sources:      []sdk.AccAddress{s.addr4},
			expErr:       "error releasing hold on payment source: stop right there",
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr4, funds: s.coins("13strawberry")}}},
			expDeleted:   []paymentKey{newPKey(s.addr4, "anid")},
		},
		{
			name: "one source: one payment",
			setup: func() {
				s.requireSetPaymentsInStore(s.newTestPayment(s.longAddr1, "1starfruit", s.longAddr3, "3tangerine", ""))
			},
			target:       s.longAddr3,
			sources:      []sdk.AccAddress{s.longAddr1},
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.longAddr1, funds: s.coins("1starfruit")}}},
			expEvents: []*exchange.EventPaymentRejected{
				exchange.NewEventPaymentRejected(s.newTestPayment(s.longAddr1, "1starfruit", s.longAddr3, "3tangerine", "")),
			},
			expDeleted: []paymentKey{newPKey(s.longAddr1, "")},
		},
		{
			name: "one source: three payments",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.addr3, "18strawberry", s.addr2, "81tomato", "one"),
					s.newTestPayment(s.addr3, "28strawberry", s.addr2, "82tangerine", "two"),
					s.newTestPayment(s.addr3, "38starfruit", s.addr2, "83tangerine", "three"),
					s.newTestPayment(s.addr4, "500strawberry", s.addr2, "", "one"),
					s.newTestPayment(s.addr3, "6starfruit", s.addr4, "10tomato", "four"),
				)
			},
			target:  s.addr2,
			sources: []sdk.AccAddress{s.addr3},
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{
				{addr: s.addr3, funds: s.coins("18strawberry")},
				{addr: s.addr3, funds: s.coins("38starfruit")},
				{addr: s.addr3, funds: s.coins("28strawberry")},
			}},
			expEvents: []*exchange.EventPaymentRejected{
				exchange.NewEventPaymentRejected(s.newTestPayment(s.addr3, "18strawberry", s.addr2, "81tomato", "one")),
				exchange.NewEventPaymentRejected(s.newTestPayment(s.addr3, "38starfruit", s.addr2, "83tangerine", "three")),
				exchange.NewEventPaymentRejected(s.newTestPayment(s.addr3, "28strawberry", s.addr2, "82tangerine", "two")),
			},
			expDeleted: []paymentKey{newPKey(s.addr3, "one"), newPKey(s.addr3, "two"), newPKey(s.addr3, "three")},
			expRemain:  []paymentKey{newPKey(s.addr3, "four"), newPKey(s.addr4, "one")},
		},
		{
			name: "duplicated source",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.addr3, "", s.longAddr1, "8tangerine", "abc"),
					s.newTestPayment(s.addr2, "7strawberry", s.longAddr1, "", "abc"),
				)
			},
			target:  s.longAddr1,
			sources: []sdk.AccAddress{s.addr3, s.addr3, s.addr2},
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{
				{addr: s.addr3, funds: nil}, {addr: s.addr2, funds: s.coins("7strawberry")},
			}},
			expEvents: []*exchange.EventPaymentRejected{
				exchange.NewEventPaymentRejected(s.newTestPayment(s.addr3, "", s.longAddr1, "8tangerine", "abc")),
				exchange.NewEventPaymentRejected(s.newTestPayment(s.addr2, "7strawberry", s.longAddr1, "", "abc")),
			},
			expDeleted: []paymentKey{newPKey(s.addr3, "abc"), newPKey(s.addr2, "abc")},
		},
		{
			name: "three sources: third does not have any payments",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.addr1, "", s.longAddr2, "1tomato", "gimmie1"),
					s.newTestPayment(s.addr2, "", s.longAddr2, "2tomato", "gimmie2"),
					s.newTestPayment(s.addr2, "", s.longAddr2, "3tomato", "gimmie3"),
					s.newTestPayment(s.addr2, "", s.longAddr2, "4tomato", "gimmie4"),
				)
			},
			target:  s.longAddr2,
			sources: []sdk.AccAddress{s.addr1, s.addr2, s.addr3},
			expErr:  "source " + s.addr3.String() + " does not have any payments for target " + s.longAddr2.String(),
			expRemain: []paymentKey{
				newPKey(s.addr1, "gimmie1"),
				newPKey(s.addr2, "gimmie2"), newPKey(s.addr2, "gimmie3"), newPKey(s.addr2, "gimmie4"),
			},
		},
		{
			name: "three sources: multiple payments",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.addr1, "151strawberry,3starfruit", s.addr5, "", ""),
					s.newTestPayment(s.addr1, "", s.addr4, "14tomato", "222"),
					s.newTestPayment(s.addr2, "251strawberry", s.addr5, "12tomato", "111"),
					s.newTestPayment(s.addr2, "252strawberry", s.addr5, "8tangerine", "222"),
					s.newTestPayment(s.addr2, "24strawberry", s.addr4, "33tomato", "333"),
					s.newTestPayment(s.longAddr3, "3351strawberry", s.addr5, "53tomato", "111"),
					s.newTestPayment(s.longAddr3, "3352strawberry", s.addr5, "35tangerine", "222"),
					s.newTestPayment(s.longAddr3, "", s.addr4, "43tangerine", "333"),
					s.newTestPayment(s.longAddr3, "3353strawberry", s.addr5, "10tomato,8tangerine", "444"),
					s.newTestPayment(s.addr5, "51strawberry", s.addr1, "15tomato", "111"),
					s.newTestPayment(s.addr5, "52strawberry", s.addr2, "25tomato", "222"),
					s.newTestPayment(s.addr5, "533strawberry", s.longAddr3, "335tangerine", "333"),
				)
			},
			target:  s.addr5,
			sources: []sdk.AccAddress{s.addr2, s.longAddr3, s.addr1},
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{
				{addr: s.addr2, funds: s.coins("251strawberry")},
				{addr: s.addr2, funds: s.coins("252strawberry")},
				{addr: s.longAddr3, funds: s.coins("3351strawberry")},
				{addr: s.longAddr3, funds: s.coins("3352strawberry")},
				{addr: s.longAddr3, funds: s.coins("3353strawberry")},
				{addr: s.addr1, funds: s.coins("151strawberry,3starfruit")},
			}},
			expEvents: []*exchange.EventPaymentRejected{
				exchange.NewEventPaymentRejected(s.newTestPayment(s.addr2, "251strawberry", s.addr5, "12tomato", "111")),
				exchange.NewEventPaymentRejected(s.newTestPayment(s.addr2, "252strawberry", s.addr5, "8tangerine", "222")),
				exchange.NewEventPaymentRejected(s.newTestPayment(s.longAddr3, "3351strawberry", s.addr5, "53tomato", "111")),
				exchange.NewEventPaymentRejected(s.newTestPayment(s.longAddr3, "3352strawberry", s.addr5, "35tangerine", "222")),
				exchange.NewEventPaymentRejected(s.newTestPayment(s.longAddr3, "3353strawberry", s.addr5, "10tomato,8tangerine", "444")),
				exchange.NewEventPaymentRejected(s.newTestPayment(s.addr1, "151strawberry,3starfruit", s.addr5, "", "")),
			},
			expDeleted: []paymentKey{
				newPKey(s.addr2, "111"), newPKey(s.addr2, "222"),
				newPKey(s.longAddr3, "111"), newPKey(s.longAddr3, "222"), newPKey(s.longAddr3, "444"),
				newPKey(s.addr1, ""),
			},
			expRemain: []paymentKey{
				newPKey(s.addr1, "222"), newPKey(s.addr2, "333"), newPKey(s.longAddr3, "333"),
				newPKey(s.addr5, "111"), newPKey(s.addr5, "222"), newPKey(s.addr5, "333"),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()

			if tc.holdKeeper == nil {
				tc.holdKeeper = NewMockHoldKeeper()
			}

			var expEvents sdk.Events
			if len(tc.expEvents) > 0 {
				expEvents = make(sdk.Events, len(tc.expEvents))
				for i, event := range tc.expEvents {
					expEvents[i] = s.untypeEvent(event)
				}
			}

			if tc.setup != nil {
				tc.setup()
			}

			targetName := s.getAddrName(tc.target)
			sourceNames := make([]string, len(tc.sources))
			for i, source := range tc.sources {
				sourceNames[i] = s.getAddrName(source)
			}

			kpr := s.k.WithHoldKeeper(tc.holdKeeper)
			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = kpr.RejectPayments(ctx, tc.target, tc.sources)
			}
			s.Require().NotPanics(testFunc, "RejectPayments(%s, %v)", targetName, sourceNames)
			s.assertErrorValue(err, tc.expErr, "RejectPayments(%s, %v) error", targetName, sourceNames)
			s.assertHoldKeeperCalls(tc.holdKeeper, tc.expHoldCalls, "RejectPayments(%s, %v) hold calls", targetName, sourceNames)

			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "RejectPayments(%s, %v) events", targetName, sourceNames)

			for i, pKey := range tc.expDeleted {
				payment, _ := s.k.GetPayment(s.ctx, pKey.source, pKey.externalID)
				s.Assert().Nil(payment, "tc.expDeleted[%d]: GetPayment(%s, %q)", i, s.getAddrName(pKey.source), pKey.externalID)
			}

			// Check that all the tc.expRemain entries still remain.
			for i, pKey := range tc.expRemain {
				payment, _ := s.k.GetPayment(s.ctx, pKey.source, pKey.externalID)
				s.Assert().NotNil(payment, "tc.expRemain[%d]: GetPayment(%s, %q)", i, s.getAddrName(pKey.source), pKey.externalID)
			}

			// Check that no other payments remain that aren't in tc.expRemain
			payments := s.getAllPayments()
			for _, payment := range payments {
				wasExp := false
				for _, pKey := range tc.expRemain {
					if pKey.source.String() == payment.Source && pKey.externalID == payment.ExternalId {
						wasExp = true
						break
					}
				}
				s.Assert().True(wasExp, "payment with source %s and external id %q still exists but was not expected to",
					payment.Source, payment.ExternalId)
			}

			s.assertTargetToPaymentIndexEntriesMatchPayments()
		})
	}
}

func (s *TestSuite) TestCancelPayments() {
	type paymentKey struct {
		source     sdk.AccAddress
		externalID string
	}
	newPKey := func(source sdk.AccAddress, externalID string) paymentKey {
		return paymentKey{source: source, externalID: externalID}
	}

	tests := []struct {
		name         string
		setup        func()
		holdKeeper   *MockHoldKeeper
		source       sdk.AccAddress
		externalIDs  []string
		expErr       string
		expHoldCalls HoldCalls
		expEvents    []*exchange.EventPaymentCancelled
		expDeleted   []paymentKey
		expRemain    []paymentKey
	}{
		{
			name:        "nil source",
			source:      nil,
			externalIDs: []string{"x"},
			expErr:      "a source is required in order to cancel payments",
		},
		{
			name:        "empty source",
			source:      sdk.AccAddress{},
			externalIDs: []string{"x"},
			expErr:      "a source is required in order to cancel payments",
		},
		{
			name:        "nil external ids",
			source:      s.addr3,
			externalIDs: nil,
			expErr:      "at least one external id is required",
		},
		{
			name:        "empty external ids",
			source:      s.addr3,
			externalIDs: []string{},
			expErr:      "at least one external id is required",
		},
		{
			name:        "one external id: no such payment",
			source:      s.addr2,
			externalIDs: []string{"noexisty"},
			expErr:      "no payment found with source " + s.addr2.String() + " and external id \"noexisty\"",
		},
		{
			name: "one external id: error releasing hold",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.longAddr1, "", s.addr1, "6tangerine", "abc"),
				)
			},
			holdKeeper:   NewMockHoldKeeper().WithReleaseHoldResults("let it go"),
			source:       s.longAddr1,
			externalIDs:  []string{"abc"},
			expErr:       "error releasing hold on payment source: let it go",
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.longAddr1, funds: nil}}},
			expDeleted:   []paymentKey{newPKey(s.longAddr1, "abc")},
		},
		{
			name: "one external id: all good",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.longAddr2, "8starfruit", s.addr4, "", "whatever"),
				)
			},
			source:       s.longAddr2,
			externalIDs:  []string{"whatever"},
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.longAddr2, funds: s.coins("8starfruit")}}},
			expEvents: []*exchange.EventPaymentCancelled{
				exchange.NewEventPaymentCancelled(s.newTestPayment(s.longAddr2, "8starfruit", s.addr4, "", "whatever")),
			},
			expDeleted: []paymentKey{newPKey(s.longAddr2, "whatever")},
		},
		{
			name: "one empty external id: no such payment",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.addr4, "8starfruit", s.addr1, "", "something"),
				)
			},
			source:      s.addr4,
			externalIDs: []string{""},
			expErr:      "no payment found with source " + s.addr4.String() + " and external id \"\"",
			expRemain:   []paymentKey{newPKey(s.addr4, "something")},
		},
		{
			name: "one empty external id: all good",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.addr3, "3strawberry", s.longAddr1, "", ""),
					s.newTestPayment(s.addr3, "4strawberry", s.longAddr2, "", "123"),
					s.newTestPayment(s.addr2, "2starfruit", s.longAddr1, "", ""),
				)
			},
			source:       s.addr3,
			externalIDs:  []string{""},
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{{addr: s.addr3, funds: s.coins("3strawberry")}}},
			expEvents: []*exchange.EventPaymentCancelled{
				exchange.NewEventPaymentCancelled(s.newTestPayment(s.addr3, "3strawberry", s.longAddr1, "", "")),
			},
			expDeleted: []paymentKey{newPKey(s.addr3, "")},
			expRemain:  []paymentKey{newPKey(s.addr3, "123"), newPKey(s.addr2, "")},
		},
		{
			name: "duplicate external id",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.longAddr3, "3strawberry", s.addr1, "", ""),
					s.newTestPayment(s.longAddr3, "4strawberry", s.longAddr2, "", "123"),
					s.newTestPayment(s.longAddr3, "5starfruit", s.longAddr1, "", "456"),
				)
			},
			source:      s.longAddr3,
			externalIDs: []string{"456", "", "456"},
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{
				{addr: s.longAddr3, funds: s.coins("5starfruit")},
				{addr: s.longAddr3, funds: s.coins("3strawberry")},
			}},
			expEvents: []*exchange.EventPaymentCancelled{
				exchange.NewEventPaymentCancelled(s.newTestPayment(s.longAddr3, "5starfruit", s.longAddr1, "", "456")),
				exchange.NewEventPaymentCancelled(s.newTestPayment(s.longAddr3, "3strawberry", s.addr1, "", "")),
			},
			expDeleted: []paymentKey{newPKey(s.longAddr3, "456"), newPKey(s.longAddr3, "")},
			expRemain:  []paymentKey{newPKey(s.longAddr3, "123")},
		},
		{
			name: "three external ids: third does not exist",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.addr2, "3strawberry", s.addr5, "", ""),
					s.newTestPayment(s.addr2, "4strawberry", s.longAddr3, "", "123"),
					s.newTestPayment(s.addr2, "", s.longAddr3, "8tomato", "456"),
				)
			},
			source:      s.addr2,
			externalIDs: []string{"123", "456", " "},
			expErr:      "no payment found with source " + s.addr2.String() + " and external id \" \"",
			expRemain:   []paymentKey{newPKey(s.addr2, ""), newPKey(s.addr2, "123"), newPKey(s.addr2, "456")},
		},
		{
			name: "three external ids: error releasing hold on third",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.addr2, "3strawberry", s.addr5, "", ""),
					s.newTestPayment(s.addr2, "4strawberry", s.longAddr3, "", "123"),
					s.newTestPayment(s.addr2, "", s.longAddr3, "8tomato", "456"),
				)
			},
			holdKeeper:  NewMockHoldKeeper().WithReleaseHoldResults("", "", "third time fails"),
			source:      s.addr2,
			externalIDs: []string{"123", "456", ""},
			expErr:      "error releasing hold on payment source: third time fails",
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{
				{addr: s.addr2, funds: s.coins("4strawberry")},
				{addr: s.addr2, funds: nil},
				{addr: s.addr2, funds: s.coins("3strawberry")},
			}},
			expDeleted: []paymentKey{newPKey(s.addr2, ""), newPKey(s.addr2, "123"), newPKey(s.addr2, "456")},
		},
		{
			name: "three external ids: all good",
			setup: func() {
				s.requireSetPaymentsInStore(
					s.newTestPayment(s.addr2, "", s.addr3, "10tomato", "BB"),
					s.newTestPayment(s.addr2, "11strawberry", s.longAddr3, "", "CC"),
					s.newTestPayment(s.addr2, "12strawberry", s.addr4, "", "DD"),
					s.newTestPayment(s.addr3, "13strawberry", s.addr1, "", "AA"),
					s.newTestPayment(s.addr3, "14strawberry", s.addr2, "", "BB"),
					s.newTestPayment(s.addr3, "15strawberry", s.longAddr3, "", "CC"),
					s.newTestPayment(s.addr3, "16strawberry", s.addr4, "", "DD"),
					s.newTestPayment(s.addr3, "17strawberry", s.addr5, "", "EE"),
					s.newTestPayment(s.addr4, "18strawberry", s.addr2, "", "BB"),
					s.newTestPayment(s.addr4, "19strawberry", s.longAddr3, "", "CC"),
					s.newTestPayment(s.addr4, "", s.addr3, "20tomato", "DD"),
					s.newTestPayment(s.longAddr3, "", s.addr3, "21tomato", "CC"),
				)
			},
			source:      s.addr3,
			externalIDs: []string{"DD", "BB", "CC"},
			expHoldCalls: HoldCalls{ReleaseHold: []*ReleaseHoldArgs{
				{addr: s.addr3, funds: s.coins("16strawberry")},
				{addr: s.addr3, funds: s.coins("14strawberry")},
				{addr: s.addr3, funds: s.coins("15strawberry")},
			}},
			expEvents: []*exchange.EventPaymentCancelled{
				exchange.NewEventPaymentCancelled(s.newTestPayment(s.addr3, "16strawberry", s.addr4, "", "DD")),
				exchange.NewEventPaymentCancelled(s.newTestPayment(s.addr3, "14strawberry", s.addr2, "", "BB")),
				exchange.NewEventPaymentCancelled(s.newTestPayment(s.addr3, "15strawberry", s.longAddr3, "", "CC")),
			},
			expDeleted: []paymentKey{newPKey(s.addr3, "BB"), newPKey(s.addr3, "CC"), newPKey(s.addr3, "DD")},
			expRemain: []paymentKey{
				newPKey(s.addr2, "BB"), newPKey(s.addr2, "CC"), newPKey(s.addr2, "DD"),
				newPKey(s.addr3, "AA"), newPKey(s.addr3, "EE"),
				newPKey(s.addr4, "BB"), newPKey(s.addr4, "CC"), newPKey(s.addr4, "DD"),
				newPKey(s.longAddr3, "CC"),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()

			if tc.holdKeeper == nil {
				tc.holdKeeper = NewMockHoldKeeper()
			}

			var expEvents sdk.Events
			if len(tc.expEvents) > 0 {
				expEvents = make(sdk.Events, len(tc.expEvents))
				for i, event := range tc.expEvents {
					expEvents[i] = s.untypeEvent(event)
				}
			}

			if tc.setup != nil {
				tc.setup()
			}

			sourceName := s.getAddrName(tc.source)
			kpr := s.k.WithHoldKeeper(tc.holdKeeper)
			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = kpr.CancelPayments(ctx, tc.source, tc.externalIDs)
			}
			s.Require().NotPanics(testFunc, "CancelPayments(%s, %q)", sourceName, tc.externalIDs)
			s.assertErrorValue(err, tc.expErr, "CancelPayments(%s, %q) error", sourceName, tc.externalIDs)
			s.assertHoldKeeperCalls(tc.holdKeeper, tc.expHoldCalls, "CancelPayments(%s, %q) hold calls", sourceName, tc.externalIDs)

			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "CancelPayments(%s, %q) events", sourceName, tc.externalIDs)

			// check that none of the tc.expDeleted entries still exist.
			for i, pKey := range tc.expDeleted {
				payment, _ := s.k.GetPayment(s.ctx, pKey.source, pKey.externalID)
				s.Assert().Nil(payment, "tc.expDeleted[%d]: GetPayment(%s, %q)", i, s.getAddrName(pKey.source), pKey.externalID)
			}

			// Check that all the tc.expRemain entries still remain.
			for i, pKey := range tc.expRemain {
				payment, _ := s.k.GetPayment(s.ctx, pKey.source, pKey.externalID)
				s.Assert().NotNil(payment, "tc.expRemain[%d]: GetPayment(%s, %q)", i, s.getAddrName(pKey.source), pKey.externalID)
			}

			// Check that no other payments remain that aren't in tc.expRemain
			payments := s.getAllPayments()
			for _, payment := range payments {
				wasExp := false
				for _, pKey := range tc.expRemain {
					if pKey.source.String() == payment.Source && pKey.externalID == payment.ExternalId {
						wasExp = true
						break
					}
				}
				s.Assert().True(wasExp, "payment with source %s and external id %q still exists but was not expected to",
					payment.Source, payment.ExternalId)
			}

			s.assertTargetToPaymentIndexEntriesMatchPayments()
		})
	}
}

func (s *TestSuite) TestUpdatePaymentTarget() {
	tests := []struct {
		name         string
		setup        func()
		source       sdk.AccAddress
		externalID   string
		newTarget    sdk.AccAddress
		expErr       string
		expOldTarget sdk.AccAddress
		expPayment   *exchange.Payment
	}{
		{
			name: "error getting payment",
			setup: func() {
				key := keeper.MakeKeyPayment(s.addr1, "boom")
				s.getStore().Set(key, []byte{'x'})
			},
			source:     s.addr1,
			externalID: "boom",
			newTarget:  s.addr2,
			expErr: "error getting existing payment with source " + s.addr1.String() +
				" and external id \"boom\": failed to unmarshal payment: unexpected EOF",
		},
		{
			name:       "no such payment",
			source:     s.addr3,
			externalID: "what",
			newTarget:  s.addr1,
			expErr:     "no payment found with source " + s.addr3.String() + " and external id \"what\"",
		},
		{
			name: "no change in target",
			setup: func() {
				s.requireSetPaymentsInStore(s.newTestPayment(s.longAddr2, "51strawberry", s.addr3, "", "aBc"))
			},
			source:     s.longAddr2,
			externalID: "aBc",
			newTarget:  s.addr3,
			expErr: "payment with source " + s.longAddr2.String() +
				" and external id \"aBc\" already has target " + s.addr3.String(),
			expPayment: s.newTestPayment(s.longAddr2, "51strawberry", s.addr3, "", "aBc"),
		},
		{
			name: "empty to nil",
			setup: func() {
				s.requireSetPaymentsInStore(s.newTestPayment(s.longAddr1, "73strawberry", nil, "1tomato", "badtrade"))
			},
			source:     s.longAddr1,
			externalID: "badtrade",
			newTarget:  nil,
			expErr: "payment with source " + s.longAddr1.String() +
				" and external id \"badtrade\" already has target <empty>",
			expPayment: s.newTestPayment(s.longAddr1, "73strawberry", nil, "1tomato", "badtrade"),
		},
		{
			name: "empty to empty",
			setup: func() {
				s.requireSetPaymentsInStore(s.newTestPayment(s.longAddr3, "1strawberry", nil, "17tangerine", "goodtrade"))
			},
			source:     s.longAddr3,
			externalID: "goodtrade",
			newTarget:  sdk.AccAddress{},
			expErr: "payment with source " + s.longAddr3.String() +
				" and external id \"goodtrade\" already has target <empty>",
			expPayment: s.newTestPayment(s.longAddr3, "1strawberry", nil, "17tangerine", "goodtrade"),
		},
		{
			name: "empty to something",
			setup: func() {
				s.requireSetPaymentsInStore(s.newTestPayment(s.addr4, "18starfruit", nil, "", "partyparty"))
			},
			source:       s.addr4,
			externalID:   "partyparty",
			newTarget:    s.longAddr1,
			expOldTarget: nil,
			expPayment:   s.newTestPayment(s.addr4, "18starfruit", s.longAddr1, "", "partyparty"),
		},
		{
			name: "something to nil",
			setup: func() {
				s.requireSetPaymentsInStore(s.newTestPayment(s.longAddr3, "14starfruit", s.longAddr2, "77tomato", "another"))
			},
			source:       s.longAddr3,
			externalID:   "another",
			newTarget:    nil,
			expOldTarget: s.longAddr2,
			expPayment:   s.newTestPayment(s.longAddr3, "14starfruit", nil, "77tomato", "another"),
		},
		{
			name: "something to empty",
			setup: func() {
				s.requireSetPaymentsInStore(s.newTestPayment(s.longAddr2, "333strawberry", s.addr5, "100tangerine", "wooo"))
			},
			source:       s.longAddr2,
			externalID:   "wooo",
			newTarget:    sdk.AccAddress{},
			expOldTarget: s.addr5,
			expPayment:   s.newTestPayment(s.longAddr2, "333strawberry", nil, "100tangerine", "wooo"),
		},
		{
			name: "something to something else",
			setup: func() {
				s.requireSetPaymentsInStore(s.newTestPayment(s.addr5, "18starfruit", s.addr1, "", "QrsT"))
			},
			source:       s.addr5,
			externalID:   "QrsT",
			newTarget:    s.addr2,
			expOldTarget: s.addr1,
			expPayment:   s.newTestPayment(s.addr5, "18starfruit", s.addr2, "", "QrsT"),
		},
		{
			name: "changed to equal the source",
			setup: func() {
				s.requireSetPaymentsInStore(s.newTestPayment(s.addr2, "6strawberry", s.longAddr1, "2tangerine", ""))
			},
			source:       s.addr2,
			externalID:   "",
			newTarget:    s.addr2,
			expOldTarget: s.longAddr1,
			expPayment:   s.newTestPayment(s.addr2, "6strawberry", s.addr2, "2tangerine", ""),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()

			var expEvents sdk.Events
			var expOldIndex, expNewIndex []byte
			if len(tc.expErr) == 0 {
				s.Require().NotNil(tc.expPayment, "tc.expPayment cannot be nil when tc.expErr is empty")
				event := exchange.NewEventPaymentUpdated(tc.expPayment, tc.expOldTarget.String())
				expEvents = sdk.Events{s.untypeEvent(event)}

				source := s.requireAccAddressFromBech32(tc.expPayment.Source, "valid tc.expPayment.Source required when tc.expErr is empty")
				if len(tc.expOldTarget) > 0 {
					expOldIndex = keeper.MakeIndexKeyTargetToPayment(tc.expOldTarget, source, tc.expPayment.ExternalId)
				}
				if len(tc.expPayment.Target) > 0 {
					target := s.requireAccAddressFromBech32(tc.expPayment.Target, "valid (or empty) tc.expPayment.Target required when tc.expErr is empty")
					expNewIndex = keeper.MakeIndexKeyTargetToPayment(target, source, tc.expPayment.ExternalId)
				}
			}

			if tc.setup != nil {
				tc.setup()
			}

			sourceName := s.getAddrName(tc.source)
			newTargetName := s.getAddrName(tc.newTarget)
			em := sdk.NewEventManager()
			ctx := s.ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = s.k.UpdatePaymentTarget(ctx, tc.source, tc.externalID, tc.newTarget)
			}
			s.Require().NotPanics(testFunc, "UpdatePaymentTarget(%s, %q, %s)", sourceName, tc.externalID, newTargetName)
			s.assertErrorValue(err, tc.expErr, "UpdatePaymentTarget(%s, %q, %s) error", sourceName, tc.externalID, newTargetName)

			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "UpdatePaymentTarget(%s, %q, %s) events", sourceName, tc.externalID, newTargetName)

			actPayment, _ := s.k.GetPayment(s.ctx, tc.source, tc.externalID)
			s.assertEqualPayment(tc.expPayment, actPayment, "payment after UpdatePaymentTarget(%s, %q, %s)", sourceName, tc.externalID, newTargetName)

			store := s.getStore()
			if len(expOldIndex) > 0 {
				has := store.Has(expOldIndex)
				s.Assert().False(has, "The old target to payment index entry should not exist, but does.")
			}
			if len(expNewIndex) > 0 {
				has := store.Has(expNewIndex)
				s.Assert().True(has, "The new target to payment index entry does not exist, but should.")
			}

			s.assertTargetToPaymentIndexEntriesMatchPayments()
		})
	}
}

func (s *TestSuite) TestGetPaymentsForTargetAndSource() {
	s.clearExchangeState()
	paymentsAddr2FromAddr1 := []*exchange.Payment{
		s.newTestPayment(s.addr1, "18strawberry", s.addr2, "", "a"),
	}
	paymentsLongAddr1FromAddr4 := []*exchange.Payment{
		s.newTestPayment(s.addr4, "5starfruit", s.longAddr1, "3tangerine", "a"),
		s.newTestPayment(s.addr4, "6strawberry", s.longAddr1, "", "b"),
		s.newTestPayment(s.addr4, "", s.longAddr1, "5tangerine", "c"),
	}
	standardSetup := func() {
		s.requireSetPaymentsInStore(
			// s.addr1 should not be the target of any payments.
			// s.addr2 should be the target of exactly one payment.
			paymentsAddr2FromAddr1[0],
			// s.longAddr1 should be the target of many payments, several of which will be from s.addr4.
			s.newTestPayment(s.addr1, "1strawberry", s.longAddr1, "", "b"),
			s.newTestPayment(s.addr2, "2strawberry", s.longAddr1, "1tomato", ""),
			s.newTestPayment(s.addr3, "", s.longAddr1, "2tomato", ""),
			paymentsLongAddr1FromAddr4[0],
			paymentsLongAddr1FromAddr4[1],
			paymentsLongAddr1FromAddr4[2],
			s.newTestPayment(s.addr5, "5starfruit,2strawberry", s.longAddr1, "4tangerine", ""),
			s.newTestPayment(s.longAddr2, "", s.longAddr1, "4tangerine,8tomato", ""),
			s.newTestPayment(s.longAddr3, "8starfruit,18strawberry", s.longAddr1, "", ""),
			// An entry with target s.addr4 and source s.longAddr1 to make sure things aren't crossed.
			s.newTestPayment(s.longAddr1, "99strawberry", s.addr4, "", ""),
			// An entry with source s.addr4 and a different target.
			s.newTestPayment(s.addr4, "", s.longAddr2, "1tomato", "e"),
			// And an entry with source s.addr4 that doesn't have a target.
			s.newTestPayment(s.addr4, "55strawberry", nil, "66tomato", "f"),
			// I'm going to overwite this one with a bad entry, but want the index entry there still.
			s.newTestPayment(s.addr4, "999strawberry", s.longAddr1, "", "g"),
		)

		// Bad payment entry with target s.longAddr1 and source s.addr4
		store := s.getStore()
		badKey := keeper.MakeKeyPayment(s.addr4, "g")
		store.Set(badKey, []byte{'x'})

		// Add an index entry to a payment that does not exist between those two.
		badInd := keeper.MakeIndexKeyTargetToPayment(s.longAddr1, s.addr4, "nope")
		store.Set(badInd, []byte{})
	}
	standardSetup()

	tests := []struct {
		name     string
		target   sdk.AccAddress
		source   sdk.AccAddress
		expected []*exchange.Payment
	}{
		{name: "nil target", target: nil, source: s.addr4, expected: nil},
		{name: "empty target", target: sdk.AccAddress{}, source: s.addr4, expected: nil},
		{name: "nil source", target: s.longAddr1, source: nil, expected: nil},
		{name: "empty source", target: s.longAddr1, source: sdk.AccAddress{}, expected: nil},
		{name: "no payments at all for target", target: s.addr1, source: s.addr2, expected: nil},
		{name: "no payments for target and source", target: s.addr2, source: s.addr3, expected: nil},
		{
			name:     "one payment",
			target:   s.addr2,
			source:   s.addr1,
			expected: paymentsAddr2FromAddr1,
		},
		{
			name:     "three payments",
			target:   s.longAddr1,
			source:   s.addr4,
			expected: paymentsLongAddr1FromAddr4,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			targetName := s.getAddrName(tc.target)
			sourceName := s.getAddrName(tc.source)
			var actPayments []*exchange.Payment
			testFunc := func() {
				actPayments = s.k.GetPaymentsForTargetAndSource(s.ctx, tc.target, tc.source)
			}
			s.Require().NotPanics(testFunc, "GetPaymentsForTargetAndSource(%s, %s)", targetName, sourceName)
			s.assertEqualPayments(tc.expected, actPayments, "GetPaymentsForTargetAndSource(%s, %s) result", targetName, sourceName)
		})
	}
}

func (s *TestSuite) TestIteratePayments() {
	var payments []*exchange.Payment
	stopAfter := func(count int) func(*exchange.Payment) bool {
		return func(payment *exchange.Payment) bool {
			payments = append(payments, payment)
			return len(payments) >= count
		}
	}
	getAll := func(payment *exchange.Payment) bool {
		payments = append(payments, payment)
		return false
	}

	threePayments := []*exchange.Payment{
		s.newTestPayment(s.addr1, "", s.addr2, "3tomato", "p"),
		s.newTestPayment(s.addr1, "55tangerine", s.addr3, "", "r"),
		s.newTestPayment(s.addr2, "1starfruit", s.addr1, "4tomato", "p"),
	}
	threePaymentSetup := func() {
		s.requireSetPaymentsInStore(threePayments...)
		// Create an empty entry and invalid entry that, when sorted, aren't last.
		// The callback should not be called for these entries.
		// So stopAfter(3) should still return the three good entries.
		keyEmpty := keeper.MakeKeyPayment(s.addr1, "o")
		keyBad := keeper.MakeKeyPayment(s.addr1, "v")
		store := s.getStore()
		store.Set(keyEmpty, []byte{})
		store.Set(keyBad, []byte{'x'})
	}

	tests := []struct {
		name  string
		setup func()
		cb    func(payment *exchange.Payment) bool
		exp   []*exchange.Payment
	}{
		{
			name: "no payments",
			cb:   getAll,
			exp:  nil,
		},
		{
			name: "one payment",
			setup: func() {
				s.requireSetPaymentsInStore(threePayments[1])
			},
			cb:  getAll,
			exp: threePayments[1:2],
		},
		{
			name:  "three payments: get all",
			setup: threePaymentSetup,
			cb:    getAll,
			exp:   threePayments,
		},
		{
			name:  "three payments: get one",
			setup: threePaymentSetup,
			cb:    stopAfter(1),
			exp:   threePayments[0:1],
		},
		{
			name:  "three payments: get two",
			setup: threePaymentSetup,
			cb:    stopAfter(2),
			exp:   threePayments[0:2],
		},
		{
			name:  "empty and bad entries are ignored",
			setup: threePaymentSetup,
			cb:    stopAfter(3),
			// If they're not ignored, the actual result will have a nil 3rd entry.
			exp: threePayments,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.setup != nil {
				tc.setup()
			}

			payments = nil
			testFunc := func() {
				s.k.IteratePayments(s.ctx, tc.cb)
			}
			s.Require().NotPanics(testFunc, "IteratePayments")
			s.assertEqualPayments(tc.exp, payments, "IteratePayments payments")
		})
	}
}

func (s *TestSuite) TestCalculatePaymentFees() {
	tests := []struct {
		name     string
		params   *exchange.Params
		payment  *exchange.Payment
		expected *exchange.QueryPaymentFeeCalcResponse
	}{
		{
			name:     "nil payment",
			params:   exchange.DefaultParams(),
			payment:  nil,
			expected: &exchange.QueryPaymentFeeCalcResponse{},
		},
		{
			name:     "payment without amounts",
			params:   exchange.DefaultParams(),
			payment:  s.newTestPayment(nil, "", nil, "", "what"),
			expected: &exchange.QueryPaymentFeeCalcResponse{},
		},
		{
			name: "payment only has source amount: no create fee options in params",
			params: &exchange.Params{
				FeeCreatePaymentFlat: nil,
				FeeAcceptPaymentFlat: s.coins("3apricot"),
			},
			payment:  s.newTestPayment(nil, "1strawberry", nil, "", ""),
			expected: &exchange.QueryPaymentFeeCalcResponse{},
		},
		{
			name: "payment only has source amount: one create fee option in params",
			params: &exchange.Params{
				FeeCreatePaymentFlat: s.coins("5cherry"),
				FeeAcceptPaymentFlat: s.coins("3apricot"),
			},
			payment:  s.newTestPayment(nil, "1strawberry", nil, "", ""),
			expected: &exchange.QueryPaymentFeeCalcResponse{FeeCreate: s.coins("5cherry")},
		},
		{
			name: "payment only has source amount: two create fee options in params",
			params: &exchange.Params{
				FeeCreatePaymentFlat: []sdk.Coin{s.coin("2cucumber"), s.coin("1cactus")},
				FeeAcceptPaymentFlat: s.coins("3apricot"),
			},
			payment:  s.newTestPayment(nil, "1strawberry", nil, "", ""),
			expected: &exchange.QueryPaymentFeeCalcResponse{FeeCreate: s.coins("2cucumber")},
		},
		{
			name: "payment only has target amount: no accept fee options in params",
			params: &exchange.Params{
				FeeCreatePaymentFlat: s.coins("1cherry"),
				FeeAcceptPaymentFlat: nil,
			},
			payment:  s.newTestPayment(nil, "", nil, "1tomato", ""),
			expected: &exchange.QueryPaymentFeeCalcResponse{},
		},
		{
			name: "payment only has target amount: one accept fee option in params",
			params: &exchange.Params{
				FeeCreatePaymentFlat: s.coins("1cherry"),
				FeeAcceptPaymentFlat: s.coins("1apple"),
			},
			payment:  s.newTestPayment(nil, "", nil, "1tomato", ""),
			expected: &exchange.QueryPaymentFeeCalcResponse{FeeAccept: s.coins("1apple")},
		},
		{
			name: "payment only has target amount: two accept fee options in params",
			params: &exchange.Params{
				FeeCreatePaymentFlat: s.coins("1cherry"),
				FeeAcceptPaymentFlat: []sdk.Coin{s.coin("12avocado"), s.coin("5acorn")},
			},
			payment:  s.newTestPayment(nil, "", nil, "1tomato", ""),
			expected: &exchange.QueryPaymentFeeCalcResponse{FeeAccept: s.coins("12avocado")},
		},
		{
			name:     "both amounts: empty params",
			params:   nil,
			payment:  s.newTestPayment(nil, "1strawberry", nil, "1tomato", ""),
			expected: &exchange.QueryPaymentFeeCalcResponse{},
		},
		{
			name:     "both amounts: only create options in params",
			params:   &exchange.Params{FeeCreatePaymentFlat: []sdk.Coin{s.coin("3cucumber"), s.coin("2cactus")}},
			payment:  s.newTestPayment(nil, "1strawberry", nil, "1tomato", ""),
			expected: &exchange.QueryPaymentFeeCalcResponse{FeeCreate: s.coins("3cucumber")},
		},
		{
			name:     "both amounts: only accept options in params",
			params:   &exchange.Params{FeeAcceptPaymentFlat: []sdk.Coin{s.coin("3avocado"), s.coin("8acorn")}},
			payment:  s.newTestPayment(nil, "1strawberry", nil, "1tomato", ""),
			expected: &exchange.QueryPaymentFeeCalcResponse{FeeAccept: s.coins("3avocado")},
		},
		{
			name:     "both amounts: both options in params",
			params:   &exchange.Params{FeeCreatePaymentFlat: s.coins("9cherry"), FeeAcceptPaymentFlat: s.coins("2apple")},
			payment:  s.newTestPayment(nil, "1strawberry", nil, "1tomato", ""),
			expected: &exchange.QueryPaymentFeeCalcResponse{FeeCreate: s.coins("9cherry"), FeeAccept: s.coins("2apple")},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.clearExchangeState()
			if tc.params != nil {
				s.k.SetParams(s.ctx, tc.params)
			}

			var actual *exchange.QueryPaymentFeeCalcResponse
			testFunc := func() {
				actual = s.k.CalculatePaymentFees(s.ctx, tc.payment)
			}
			s.Require().NotPanics(testFunc, "CalculatePaymentFees(%s)", tc.payment)
			if !s.Assert().Equal(tc.expected, actual, "CalculatePaymentFees(%s) response", tc.payment) && tc.expected != nil && actual != nil {
				s.Assert().Equal(tc.expected.FeeCreate.String(), tc.expected.FeeAccept.String(), "CalculatePaymentFees(%s) response FeeAccept", tc.payment)
				s.Assert().Equal(tc.expected.FeeCreate.String(), tc.expected.FeeCreate.String(), "CalculatePaymentFees(%s) response FeeCreate", tc.payment)
			}
		})
	}
}
