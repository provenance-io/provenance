package exchange

import (
	"github.com/gogo/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewEventOrderCreated(order *Order) *EventOrderCreated {
	return &EventOrderCreated{
		OrderId:   order.GetOrderID(),
		OrderType: order.GetOrderType(),
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

func NewEventMarketWithdraw(marketID uint32, amount sdk.Coins, destination, withdrawnBy sdk.AccAddress) *EventMarketWithdraw {
	return &EventMarketWithdraw{
		MarketId:    marketID,
		Amount:      amount.String(),
		Destination: destination.String(),
		WithdrawnBy: withdrawnBy.String(),
	}
}

func NewEventMarketDetailsUpdated(marketID uint32, updatedBy sdk.AccAddress) *EventMarketDetailsUpdated {
	return &EventMarketDetailsUpdated{
		MarketId:  marketID,
		UpdatedBy: updatedBy.String(),
	}
}

// NewEventMarketActiveUpdated returns a new EventMarketEnabled if isActive == true,
// or a new EventMarketDisabled if isActive == false.
func NewEventMarketActiveUpdated(marketID uint32, updatedBy sdk.AccAddress, isActive bool) proto.Message {
	if isActive {
		return NewEventMarketEnabled(marketID, updatedBy)
	}
	return NewEventMarketDisabled(marketID, updatedBy)
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

// NewEventMarketUserSettleUpdated returns a new EventMarketUserSettleEnabled if isAllowed == true,
// or a new EventMarketUserSettleDisabled if isActive == false.
func NewEventMarketUserSettleUpdated(marketID uint32, updatedBy sdk.AccAddress, isAllowed bool) proto.Message {
	if isAllowed {
		return NewEventMarketUserSettleEnabled(marketID, updatedBy)
	}
	return NewEventMarketUserSettleDisabled(marketID, updatedBy)
}

func NewEventMarketUserSettleEnabled(marketID uint32, updatedBy sdk.AccAddress) *EventMarketUserSettleEnabled {
	return &EventMarketUserSettleEnabled{
		MarketId:  marketID,
		UpdatedBy: updatedBy.String(),
	}
}

func NewEventMarketUserSettleDisabled(marketID uint32, updatedBy sdk.AccAddress) *EventMarketUserSettleDisabled {
	return &EventMarketUserSettleDisabled{
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
