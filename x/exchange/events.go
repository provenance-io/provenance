package exchange

import sdk "github.com/cosmos/cosmos-sdk/types"

func NewEventOrderCreated(order *Order) *EventOrderCreated {
	return &EventOrderCreated{
		OrderId:   order.GetOrderId(),
		OrderType: order.OrderType(),
	}
}

func NewEventOrderCancelled(orderID uint64, cancelledBy sdk.AccAddress) *EventOrderCancelled {
	return &EventOrderCancelled{
		OrderId:     orderID,
		CancelledBy: cancelledBy.String(),
	}
}

func NewEventOrderFilled(orderID uint64) *EventOrderFilled {
	return &EventOrderFilled{
		OrderId: orderID,
	}
}

func NewEventOrderPartiallyFilled(orderID uint64, assetsFilled, feesFilled sdk.Coins) *EventOrderPartiallyFilled {
	return &EventOrderPartiallyFilled{
		OrderId:      orderID,
		AssetsFilled: assetsFilled.String(),
		FeesFilled:   feesFilled.String(),
	}
}

func NewEventMarketWithdraw(marketID uint32, amountWithdrawn sdk.Coins, destination, withdrawnBy sdk.AccAddress) *EventMarketWithdraw {
	return &EventMarketWithdraw{
		MarketId:        marketID,
		AmountWithdrawn: amountWithdrawn.String(),
		Destination:     destination.String(),
		WithdrawnBy:     withdrawnBy.String(),
	}
}

func NewEventMarketDetailsUpdated(marketID uint32, updatedBy sdk.AccAddress) *EventMarketDetailsUpdated {
	return &EventMarketDetailsUpdated{
		MarketId:  marketID,
		UpdatedBy: updatedBy.String(),
	}
}

func NewEventMarketEnabled(marketID uint32, updatedBy sdk.AccAddress) *EventMarketEnabled {
	return &EventMarketEnabled{
		MarketId:  marketID,
		UpdatedBy: updatedBy.String(),
	}
}

func NewEventMarketDisabled(marketID uint32, updatedBy sdk.AccAddress) *EventMarketDisabled {
	return &EventMarketDisabled{
		MarketId:  marketID,
		UpdatedBy: updatedBy.String(),
	}
}

func NewEventMarketUserSettleUpdated(marketID uint32, updatedBy sdk.AccAddress) *EventMarketUserSettleUpdated {
	return &EventMarketUserSettleUpdated{
		MarketId:  marketID,
		UpdatedBy: updatedBy.String(),
	}
}

func NewEventMarketPermissionsUpdated(marketID uint32, updatedBy sdk.AccAddress) *EventMarketPermissionsUpdated {
	return &EventMarketPermissionsUpdated{
		MarketId:  marketID,
		UpdatedBy: updatedBy.String(),
	}
}

func NewEventMarketReqAttrUpdated(marketID uint32, updatedBy sdk.AccAddress) *EventMarketReqAttrUpdated {
	return &EventMarketReqAttrUpdated{
		MarketId:  marketID,
		UpdatedBy: updatedBy.String(),
	}
}

func NewEventCreateMarketSubmitted(marketID uint32, proposalID uint64, submittedBy sdk.AccAddress) *EventCreateMarketSubmitted {
	return &EventCreateMarketSubmitted{
		MarketId:    marketID,
		ProposalId:  proposalID,
		SubmittedBy: submittedBy.String(),
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
