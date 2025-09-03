package exchange

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

// NewEventOrderCreated creates a new event signaling that an order was created.
func NewEventOrderCreated(order OrderI) *EventOrderCreated {
	return &EventOrderCreated{
		OrderId:    order.GetOrderID(),
		OrderType:  order.GetOrderType(),
		MarketId:   order.GetMarketID(),
		ExternalId: order.GetExternalID(),
	}
}

// NewEventOrderCancelled creates a new event signaling that an order was cancelled.
func NewEventOrderCancelled(order OrderI, cancelledBy string) *EventOrderCancelled {
	return &EventOrderCancelled{
		OrderId:     order.GetOrderID(),
		CancelledBy: cancelledBy,
		MarketId:    order.GetMarketID(),
		ExternalId:  order.GetExternalID(),
	}
}

// NewEventOrderFilled creates a new event signaling that an order was filled.
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

// NewEventOrderPartiallyFilled creates a new event signaling that an order was partially filled.
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

// NewEventOrderExternalIDUpdated creates a new event for updated external order ID.
func NewEventOrderExternalIDUpdated(order OrderI) *EventOrderExternalIDUpdated {
	return &EventOrderExternalIDUpdated{
		OrderId:    order.GetOrderID(),
		MarketId:   order.GetMarketID(),
		ExternalId: order.GetExternalID(),
	}
}

// NewEventFundsCommitted creates a new event for committed funds.
func NewEventFundsCommitted(account string, marketID uint32, amount sdk.Coins, tag string) *EventFundsCommitted {
	return &EventFundsCommitted{
		Account:  account,
		MarketId: marketID,
		Amount:   amount.String(),
		Tag:      tag,
	}
}

// NewEventCommitmentReleased creates a new event for released commitments.
func NewEventCommitmentReleased(account string, marketID uint32, amount sdk.Coins, tag string) *EventCommitmentReleased {
	return &EventCommitmentReleased{
		Account:  account,
		MarketId: marketID,
		Amount:   amount.String(),
		Tag:      tag,
	}
}

// NewEventMarketWithdraw creates a new event for market withdrawal.
func NewEventMarketWithdraw(marketID uint32, amount sdk.Coins, destination sdk.AccAddress, withdrawnBy string) *EventMarketWithdraw {
	return &EventMarketWithdraw{
		MarketId:    marketID,
		Amount:      amount.String(),
		Destination: destination.String(),
		WithdrawnBy: withdrawnBy,
	}
}

// NewEventMarketDetailsUpdated creates a new event for market details update.
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

// NewEventMarketOrdersEnabled creates a new event when market orders are enabled.
func NewEventMarketOrdersEnabled(marketID uint32, updatedBy string) *EventMarketOrdersEnabled {
	return &EventMarketOrdersEnabled{
		MarketId:  marketID,
		UpdatedBy: updatedBy,
	}
}

// NewEventMarketOrdersDisabled creates a new event when market orders are disabled.
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

// NewEventMarketUserSettleEnabled creates a new event when user settle is enabled.
func NewEventMarketUserSettleEnabled(marketID uint32, updatedBy string) *EventMarketUserSettleEnabled {
	return &EventMarketUserSettleEnabled{
		MarketId:  marketID,
		UpdatedBy: updatedBy,
	}
}

// NewEventMarketUserSettleDisabled creates a new event when user settle is disabled.
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

// NewEventMarketCommitmentsEnabled creates a new event when market commitments are enabled.
func NewEventMarketCommitmentsEnabled(marketID uint32, updatedBy string) *EventMarketCommitmentsEnabled {
	return &EventMarketCommitmentsEnabled{
		MarketId:  marketID,
		UpdatedBy: updatedBy,
	}
}

// NewEventMarketCommitmentsDisabled creates a new event when market commitments are disabled.
func NewEventMarketCommitmentsDisabled(marketID uint32, updatedBy string) *EventMarketCommitmentsDisabled {
	return &EventMarketCommitmentsDisabled{
		MarketId:  marketID,
		UpdatedBy: updatedBy,
	}
}

// NewEventMarketIntermediaryDenomUpdated creates a new event when the intermediary denom is updated.
func NewEventMarketIntermediaryDenomUpdated(marketID uint32, updatedBy string) *EventMarketIntermediaryDenomUpdated {
	return &EventMarketIntermediaryDenomUpdated{
		MarketId:  marketID,
		UpdatedBy: updatedBy,
	}
}

// NewEventMarketPermissionsUpdated creates a new event when market permissions are updated.
func NewEventMarketPermissionsUpdated(marketID uint32, updatedBy string) *EventMarketPermissionsUpdated {
	return &EventMarketPermissionsUpdated{
		MarketId:  marketID,
		UpdatedBy: updatedBy,
	}
}

// NewEventMarketReqAttrUpdated creates a new event when required market attributes are updated.
func NewEventMarketReqAttrUpdated(marketID uint32, updatedBy string) *EventMarketReqAttrUpdated {
	return &EventMarketReqAttrUpdated{
		MarketId:  marketID,
		UpdatedBy: updatedBy,
	}
}

// NewEventMarketCreated creates a new event signaling that a market was created.
func NewEventMarketCreated(marketID uint32) *EventMarketCreated {
	return &EventMarketCreated{
		MarketId: marketID,
	}
}

// NewEventMarketFeesUpdated creates a new event when market fees are updated.
func NewEventMarketFeesUpdated(marketID uint32) *EventMarketFeesUpdated {
	return &EventMarketFeesUpdated{
		MarketId: marketID,
	}
}

// NewEventParamsUpdated creates a new event when parameters are updated.
func NewEventParamsUpdated() *EventParamsUpdated {
	return &EventParamsUpdated{}
}

// NewEventPaymentCreated creates a new event when a payment is created.
func NewEventPaymentCreated(payment *Payment) *EventPaymentCreated {
	return &EventPaymentCreated{
		Source:       payment.Source,
		SourceAmount: payment.SourceAmount.String(),
		Target:       payment.Target,
		TargetAmount: payment.TargetAmount.String(),
		ExternalId:   payment.ExternalId,
	}
}

// NewEventPaymentUpdated creates a new event when a payment is updated.
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

// NewEventPaymentAccepted creates a new event when a payment is accepted.
func NewEventPaymentAccepted(payment *Payment) *EventPaymentAccepted {
	return &EventPaymentAccepted{
		Source:       payment.Source,
		SourceAmount: payment.SourceAmount.String(),
		Target:       payment.Target,
		TargetAmount: payment.TargetAmount.String(),
		ExternalId:   payment.ExternalId,
	}
}

// NewEventPaymentRejected creates a new event when a payment is rejected.
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

// NewEventPaymentCancelled creates a new event when a payment is cancelled.
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
