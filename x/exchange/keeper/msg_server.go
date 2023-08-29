package keeper

import (
	"context"

	"github.com/provenance-io/provenance/x/exchange"
)

// MsgServer is an alias for a Keeper that implements the exchange.MsgServer interface.
type MsgServer Keeper

func NewMsgServer(k Keeper) exchange.MsgServer {
	return MsgServer(k)
}

var _ exchange.MsgServer = MsgServer{}

func (k MsgServer) CreateAsk(goCtx context.Context, req *exchange.MsgCreateAskRequest) (*exchange.MsgCreateAskResponse, error) {
	// TODO[1658]: Implement CreateAsk
	panic("not implemented")
}

func (k MsgServer) CreateBid(goCtx context.Context, req *exchange.MsgCreateBidRequest) (*exchange.MsgCreateBidResponse, error) {
	// TODO[1658]: Implement CreateBid
	panic("not implemented")
}

func (k MsgServer) CancelOrder(goCtx context.Context, req *exchange.MsgCancelOrderRequest) (*exchange.MsgCancelOrderResponse, error) {
	// TODO[1658]: Implement CancelOrder
	panic("not implemented")
}

func (k MsgServer) FillBids(goCtx context.Context, req *exchange.MsgFillBidsRequest) (*exchange.MsgFillBidsResponse, error) {
	// TODO[1658]: Implement FillBids
	panic("not implemented")
}

func (k MsgServer) FillAsks(goCtx context.Context, req *exchange.MsgFillAsksRequest) (*exchange.MsgFillAsksResponse, error) {
	// TODO[1658]: Implement FillAsks
	panic("not implemented")
}

func (k MsgServer) MarketSettle(goCtx context.Context, req *exchange.MsgMarketSettleRequest) (*exchange.MsgMarketSettleResponse, error) {
	// TODO[1658]: Implement MarketSettle
	panic("not implemented")
}

func (k MsgServer) MarketWithdraw(goCtx context.Context, req *exchange.MsgMarketWithdrawRequest) (*exchange.MsgMarketWithdrawResponse, error) {
	// TODO[1658]: Implement MarketWithdraw
	panic("not implemented")
}

func (k MsgServer) MarketUpdateDetails(goCtx context.Context, req *exchange.MsgMarketUpdateDetailsRequest) (*exchange.MsgMarketUpdateDetailsResponse, error) {
	// TODO[1658]: Implement MarketUpdateDetails
	panic("not implemented")
}

func (k MsgServer) MarketUpdateEnabled(goCtx context.Context, req *exchange.MsgMarketUpdateEnabledRequest) (*exchange.MsgMarketUpdateEnabledResponse, error) {
	// TODO[1658]: Implement MarketUpdateEnabled
	panic("not implemented")
}

func (k MsgServer) MarketUpdateUserSettle(goCtx context.Context, req *exchange.MsgMarketUpdateUserSettleRequest) (*exchange.MsgMarketUpdateUserSettleResponse, error) {
	// TODO[1658]: Implement MarketUpdateUserSettle
	panic("not implemented")
}

func (k MsgServer) MarketManagePermissions(goCtx context.Context, req *exchange.MsgMarketManagePermissionsRequest) (*exchange.MsgMarketManagePermissionsResponse, error) {
	// TODO[1658]: Implement MarketManagePermissions
	panic("not implemented")
}

func (k MsgServer) MarketManageReqAttrs(goCtx context.Context, req *exchange.MsgMarketManageReqAttrsRequest) (*exchange.MsgMarketManageReqAttrsResponse, error) {
	// TODO[1658]: Implement MarketManageReqAttrs
	panic("not implemented")
}

func (k MsgServer) GovCreateMarket(goCtx context.Context, req *exchange.MsgGovCreateMarketRequest) (*exchange.MsgGovCreateMarketResponse, error) {
	// TODO[1658]: Implement GovCreateMarket
	panic("not implemented")
}

func (k MsgServer) GovManageFees(goCtx context.Context, req *exchange.MsgGovManageFeesRequest) (*exchange.MsgGovManageFeesResponse, error) {
	// TODO[1658]: Implement GovManageFees
	panic("not implemented")
}

func (k MsgServer) GovUpdateParams(goCtx context.Context, req *exchange.MsgGovUpdateParamsRequest) (*exchange.MsgGovUpdateParamsResponse, error) {
	// TODO[1658]: Implement GovUpdateParams
	panic("not implemented")
}
