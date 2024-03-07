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

// NewEventMarketActiveUpdated returns a new EventMarketEnabled if isActive == true,
// or a new EventMarketDisabled if isActive == false.
func NewEventMarketActiveUpdated(marketID uint32, updatedBy string, isActive bool) proto.Message {
	if isActive {
		return NewEventMarketEnabled(marketID, updatedBy)
	}
	return NewEventMarketDisabled(marketID, updatedBy)
}

func NewEventMarketEnabled(marketID uint32, updatedBy string) *EventMarketEnabled {
	return &EventMarketEnabled{
		MarketId:  marketID,
		UpdatedBy: updatedBy,
	}
}

func NewEventMarketDisabled(marketID uint32, updatedBy string) *EventMarketDisabled {
	return &EventMarketDisabled{
		MarketId:  marketID,
		UpdatedBy: updatedBy,
	}
}

// NewEventMarketUserSettleUpdated returns a new EventMarketUserSettleEnabled if isAllowed == true,
// or a new EventMarketUserSettleDisabled if isActive == false.
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
