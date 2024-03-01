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
			s.assertErrorValue(err, tc.expErr, "GetPayment(%s, %q)", s.getAddrName(tc.source), tc.externalID)
			s.assertEqualPayment(tc.expPayment, payment, "GetPayment(%s, %q)", s.getAddrName(tc.source), tc.externalID)
		})
	}
}

func (s *TestSuite) TestCreatePayment() {
	tests := []struct {
		name        string
		setup       func()
		holdKeeper  *MockHoldKeeper
		payment     *exchange.Payment
		expPayment  *exchange.Payment // Set to payment when expStored is true.
		expErr      string
		expStored   bool
		expIndex    bool
		expHoldCall bool
		expEvent    bool
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
			name:        "error adding hold",
			holdKeeper:  NewMockHoldKeeper().WithAddHoldResults("you know you can't do that"),
			payment:     s.newTestPayment(s.longAddr3, "3starfruit", s.longAddr2, "1tangerine", "nopenope"),
			expErr:      "error placing hold on payment source: you know you can't do that",
			expStored:   true,
			expIndex:    true,
			expHoldCall: true,
		},
		{
			name:        "fully filled",
			payment:     s.newTestPayment(s.longAddr3, "88starfruit", s.longAddr1, "12tangerine,8tomato", "long-addrs"),
			expStored:   true,
			expIndex:    true,
			expHoldCall: true,
			expEvent:    true,
		},
		{
			name:        "no external id",
			payment:     s.newTestPayment(s.addr2, "12strawberry,4starfruit", s.addr3, "", ""),
			expStored:   true,
			expIndex:    true,
			expHoldCall: true,
			expEvent:    true,
		},
		{
			name:        "no target",
			payment:     s.newTestPayment(s.longAddr2, "", nil, "3tomato", "soon"),
			expStored:   true,
			expIndex:    false,
			expHoldCall: true,
			expEvent:    true,
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
			if tc.expHoldCall {
				s.Require().NotNil(tc.payment, "tc.payment cannot be nil when tc.expHoldCall = true")
				expHoldCalls.AddHold = []*AddHoldArgs{
					{
						addr:   s.requireAccAddressFromBech32(tc.payment.Source, "valid payment source required when tc.expHoldCall = true"),
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
			s.assertErrorValue(err, tc.expErr, "CreatePayment(%s)", tc.payment)
			s.assertHoldKeeperCalls(tc.holdKeeper, expHoldCalls, "hold keeper calls during CreatePayment(%s)", tc.payment)

			var actPayment *exchange.Payment
			if tc.payment != nil && len(tc.payment.Source) > 0 {
				source, sErr := sdk.AccAddressFromBech32(tc.payment.Source)
				if sErr == nil {
					actPayment, _ = s.k.GetPayment(s.ctx, source, tc.payment.ExternalId)
				}
			}
			s.assertEqualPayment(tc.expPayment, actPayment, "payment read from state after CreatePayment(%s)", tc.payment)

			actEvents := em.Events()
			s.assertEqualEvents(expEvents, actEvents, "events emitted during CreatePayment(%s)", tc.payment)

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

// TODO[1703]: func (s *TestSuite) TestAcceptPayment()

// TODO[1703]: func (s *TestSuite) TestRejectPayment()

// TODO[1703]: func (s *TestSuite) TestRejectPayments()

// TODO[1703]: func (s *TestSuite) TestCancelPayments()

// TODO[1703]: func (s *TestSuite) TestUpdatePaymentTarget()

// TODO[1703]: func (s *TestSuite) TestGetPaymentsForTargetAndSource()

// TODO[1703]: func (s *TestSuite) TestIteratePayments()

// TODO[1703]: func (s *TestSuite) TestCalculatePaymentFeess()
