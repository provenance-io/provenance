package exchange

import (
	"github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewEventOrderCreated(order OrderI) *EventOrderCreated {
	return &EventOrderCreated{
		OrderId:    order.GetOrderID(),
		OrderType:  order.GetOrderType(),
		MarketId:   order.GetMarketID(),
		ExternalId: order.GetExternalID(),
	}
}

func NewEventOrderCancelled(order OrderI, cancelledBy string) *EventOrderCancelled {
	return &EventOrderCancelled{
		OrderId:     order.GetOrderID(),
		CancelledBy: cancelledBy,
		MarketId:    order.GetMarketID(),
		ExternalId:  order.GetExternalID(),
	}
}

func NewEventOrderFilled(order OrderI) *EventOrderFilled {
	return &EventOrderFilled{
		OrderId:    order.GetOrderID(),
		Assets:     order.GetAssets().String(),
		Price:      order.GetPrice().String(),
		Fees:       order.GetSettlementFees().String(),
		MarketId:   order.GetMarketID(),
		ExternalId: order.GetExternalID(),
	}
}

func NewEventOrderPartiallyFilled(order OrderI) *EventOrderPartiallyFilled {
	return &EventOrderPartiallyFilled{
		OrderId:    order.GetOrderID(),
		Assets:     order.GetAssets().String(),
		Price:      order.GetPrice().String(),
		Fees:       order.GetSettlementFees().String(),
		MarketId:   order.GetMarketID(),
		ExternalId: order.GetExternalID(),
	}
}

func NewEventOrderExternalIDUpdated(order OrderI) *EventOrderExternalIDUpdated {
	return &EventOrderExternalIDUpdated{
		OrderId:    order.GetOrderID(),
		MarketId:   order.GetMarketID(),
		ExternalId: order.GetExternalID(),
	}
}

func NewEventFundsCommitted(account string, marketID uint32, amount sdk.Coins, tag string) *EventFundsCommitted {
	return &EventFundsCommitted{
		Account:  account,
		MarketId: marketID,
		Amount:   amount.String(),
		Tag:      tag,
	}
}

func NewEventCommitmentReleased(account string, marketID uint32, amount sdk.Coins, tag string) *EventCommitmentReleased {
	return &EventCommitmentReleased{
		Account:  account,
		MarketId: marketID,
		Amount:   amount.String(),
		Tag:      tag,
	}
}

func NewEventMarketWithdraw(marketID uint32, amount sdk.Coins, destination sdk.AccAddress, withdrawnBy string) *EventMarketWithdraw {
	return &EventMarketWithdraw{
		MarketId:    marketID,
		Amount:      amount.String(),
		Destination: destination.String(),
		WithdrawnBy: withdrawnBy,
	}
}

func NewEventMarketDetailsUpdated(marketID uint32, updatedBy string) *EventMarketDetailsUpdated {
	return &EventMarketDetailsUpdated{
		MarketId:  marketID,
		UpdatedBy: updatedBy,
	}
}

// NewEventMarketAcceptingOrdersUpdated returns a new EventMarketOrdersEnabled if isAccepting == true,
// or a new EventMarketOrdersDisabled if isAccepting == false.
func NewEventMarketAcceptingOrdersUpdated(marketID uint32, updatedBy string, isAccepting bool) proto.Message {
	if isAccepting {
		return NewEventMarketOrdersEnabled(marketID, updatedBy)
	}
	return NewEventMarketOrdersDisabled(marketID, updatedBy)
}

func NewEventMarketOrdersEnabled(marketID uint32, updatedBy string) *EventMarketOrdersEnabled {
	return &EventMarketOrdersEnabled{
		MarketId:  marketID,
		UpdatedBy: updatedBy,
	}
}

func NewEventMarketOrdersDisabled(marketID uint32, updatedBy string) *EventMarketOrdersDisabled {
	return &EventMarketOrdersDisabled{
		MarketId:  marketID,
		UpdatedBy: updatedBy,
	}
}

// NewEventMarketUserSettleUpdated returns a new EventMarketUserSettleEnabled if isAllowed == true,
// or a new EventMarketUserSettleDisabled if isAllowed == false.
func NewEventMarketUserSettleUpdated(marketID uint32, updatedBy string, isAllowed bool) proto.Message {
	if isAllowed {
		return NewEventMarketUserSettleEnabled(marketID, updatedBy)
	}
	return NewEventMarketUserSettleDisabled(marketID, updatedBy)
}

func NewEventMarketUserSettleEnabled(marketID uint32, updatedBy string) *EventMarketUserSettleEnabled {
	return &EventMarketUserSettleEnabled{
		MarketId:  marketID,
		UpdatedBy: updatedBy,
	}
}

func NewEventMarketUserSettleDisabled(marketID uint32, updatedBy string) *EventMarketUserSettleDisabled {
	return &EventMarketUserSettleDisabled{
		MarketId:  marketID,
		UpdatedBy: updatedBy,
	}
}

// NewEventMarketAcceptingCommitmentsUpdated returns a new NewEventMarketCommitmentsEnabled if isAccepting == true,
// or a new NewEventMarketCommitmentsDisabled if isAccepting == false.
func NewEventMarketAcceptingCommitmentsUpdated(marketID uint32, updatedBy string, isAccepting bool) proto.Message {
	if isAccepting {
		return NewEventMarketCommitmentsEnabled(marketID, updatedBy)
	}
	return NewEventMarketCommitmentsDisabled(marketID, updatedBy)
}

func NewEventMarketCommitmentsEnabled(marketID uint32, updatedBy string) *EventMarketCommitmentsEnabled {
	return &EventMarketCommitmentsEnabled{
		MarketId:  marketID,
		UpdatedBy: updatedBy,
	}
}

func NewEventMarketCommitmentsDisabled(marketID uint32, updatedBy string) *EventMarketCommitmentsDisabled {
	return &EventMarketCommitmentsDisabled{
		MarketId:  marketID,
		UpdatedBy: updatedBy,
	}
}

func NewEventMarketIntermediaryDenomUpdated(marketID uint32, updatedBy string) *EventMarketIntermediaryDenomUpdated {
	return &EventMarketIntermediaryDenomUpdated{
		MarketId:  marketID,
		UpdatedBy: updatedBy,
	}
}

func NewEventMarketPermissionsUpdated(marketID uint32, updatedBy string) *EventMarketPermissionsUpdated {
	return &EventMarketPermissionsUpdated{
		MarketId:  marketID,
		UpdatedBy: updatedBy,
	}
}

func NewEventMarketReqAttrUpdated(marketID uint32, updatedBy string) *EventMarketReqAttrUpdated {
	return &EventMarketReqAttrUpdated{
		MarketId:  marketID,
		UpdatedBy: updatedBy,
	}
}

func NewEventMarketCreated(marketID uint32) *EventMarketCreated {
	return &EventMarketCreated{
		MarketId: marketID,
	}
}

func NewEventMarketFeesUpdated(marketID uint32) *EventMarketFeesUpdated {
	return &EventMarketFeesUpdated{
		MarketId: marketID,
	}
}

func NewEventParamsUpdated() *EventParamsUpdated {
	return &EventParamsUpdated{}
}

func NewEventPaymentCreated(payment *Payment) *EventPaymentCreated {
	return &EventPaymentCreated{
		Source:       payment.Source,
		SourceAmount: payment.SourceAmount.String(),
		Target:       payment.Target,
		TargetAmount: payment.TargetAmount.String(),
		ExternalId:   payment.ExternalId,
	}
}

func NewEventPaymentUpdated(payment *Payment, oldTarget string) *EventPaymentUpdated {
	return &EventPaymentUpdated{
		Source:       payment.Source,
		SourceAmount: payment.SourceAmount.String(),
		OldTarget:    oldTarget,
		NewTarget:    payment.Target,
		TargetAmount: payment.TargetAmount.String(),
		ExternalId:   payment.ExternalId,
	}
}

func NewEventPaymentAccepted(payment *Payment) *EventPaymentAccepted {
	return &EventPaymentAccepted{
		Source:       payment.Source,
		SourceAmount: payment.SourceAmount.String(),
		Target:       payment.Target,
		TargetAmount: payment.TargetAmount.String(),
		ExternalId:   payment.ExternalId,
	}
}

func NewEventPaymentRejected(payment *Payment) *EventPaymentRejected {
	return &EventPaymentRejected{
		Source:     payment.Source,
		Target:     payment.Target,
		ExternalId: payment.ExternalId,
	}
}

// NewEventsPaymentsRejected creates a payment-rejected event for each payment provided.
func NewEventsPaymentsRejected(payments []*Payment) []*EventPaymentRejected {
	rv := make([]*EventPaymentRejected, len(payments))
	for i, payment := range payments {
		rv[i] = NewEventPaymentRejected(payment)
	}
	return rv
}

func NewEventPaymentCancelled(payment *Payment) *EventPaymentCancelled {
	return &EventPaymentCancelled{
		Source:     payment.Source,
		Target:     payment.Target,
		ExternalId: payment.ExternalId,
	}
}

// NewEventsPaymentsCancelled creates a payment-cancelled event for each payment provided.
func NewEventsPaymentsCancelled(payments []*Payment) []*EventPaymentCancelled {
	rv := make([]*EventPaymentCancelled, len(payments))
	for i, payment := range payments {
		rv[i] = NewEventPaymentCancelled(payment)
	}
	return rv
}
