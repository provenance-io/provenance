package keeper

import (
	"context"

	cerrs "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/provenance-io/provenance/x/exchange"
)

// MsgServer is an alias for a Keeper that implements the exchange.MsgServer interface.
type MsgServer struct {
	Keeper
}

func NewMsgServer(k Keeper) exchange.MsgServer {
	return MsgServer{
		Keeper: k,
	}
}

var _ exchange.MsgServer = MsgServer{}

func (k MsgServer) CreateAsk(goCtx context.Context, msg *exchange.MsgCreateAskRequest) (*exchange.MsgCreateAskResponse, error) {
	// TODO[1658]: Implement CreateAsk
	panic("not implemented")
}

func (k MsgServer) CreateBid(goCtx context.Context, msg *exchange.MsgCreateBidRequest) (*exchange.MsgCreateBidResponse, error) {
	// TODO[1658]: Implement CreateBid
	panic("not implemented")
}

func (k MsgServer) CancelOrder(goCtx context.Context, msg *exchange.MsgCancelOrderRequest) (*exchange.MsgCancelOrderResponse, error) {
	// TODO[1658]: Implement CancelOrder
	panic("not implemented")
}

func (k MsgServer) FillBids(goCtx context.Context, msg *exchange.MsgFillBidsRequest) (*exchange.MsgFillBidsResponse, error) {
	// TODO[1658]: Implement FillBids
	panic("not implemented")
}

func (k MsgServer) FillAsks(goCtx context.Context, msg *exchange.MsgFillAsksRequest) (*exchange.MsgFillAsksResponse, error) {
	// TODO[1658]: Implement FillAsks
	panic("not implemented")
}

func (k MsgServer) MarketSettle(goCtx context.Context, msg *exchange.MsgMarketSettleRequest) (*exchange.MsgMarketSettleResponse, error) {
	// TODO[1658]: Implement MarketSettle
	panic("not implemented")
}

func (k MsgServer) MarketWithdraw(goCtx context.Context, msg *exchange.MsgMarketWithdrawRequest) (*exchange.MsgMarketWithdrawResponse, error) {
	// TODO[1658]: Implement MarketWithdraw
	panic("not implemented")
}

func (k MsgServer) MarketUpdateDetails(goCtx context.Context, msg *exchange.MsgMarketUpdateDetailsRequest) (*exchange.MsgMarketUpdateDetailsResponse, error) {
	// TODO[1658]: Implement MarketUpdateDetails
	panic("not implemented")
}

func (k MsgServer) MarketUpdateEnabled(goCtx context.Context, msg *exchange.MsgMarketUpdateEnabledRequest) (*exchange.MsgMarketUpdateEnabledResponse, error) {
	// TODO[1658]: Implement MarketUpdateEnabled
	panic("not implemented")
}

func (k MsgServer) MarketUpdateUserSettle(goCtx context.Context, msg *exchange.MsgMarketUpdateUserSettleRequest) (*exchange.MsgMarketUpdateUserSettleResponse, error) {
	// TODO[1658]: Implement MarketUpdateUserSettle
	panic("not implemented")
}

func (k MsgServer) MarketManagePermissions(goCtx context.Context, msg *exchange.MsgMarketManagePermissionsRequest) (*exchange.MsgMarketManagePermissionsResponse, error) {
	// TODO[1658]: Implement MarketManagePermissions
	panic("not implemented")
}

func (k MsgServer) MarketManageReqAttrs(goCtx context.Context, msg *exchange.MsgMarketManageReqAttrsRequest) (*exchange.MsgMarketManageReqAttrsResponse, error) {
	// TODO[1658]: Implement MarketManageReqAttrs
	panic("not implemented")
}

func (k MsgServer) GovCreateMarket(goCtx context.Context, msg *exchange.MsgGovCreateMarketRequest) (*exchange.MsgGovCreateMarketResponse, error) {
	if msg.Authority != k.GetAuthority() {
		return nil, cerrs.Wrapf(govtypes.ErrInvalidSigner, "expected %s got %s", k.GetAuthority(), msg.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO[1658]: Implement GovCreateMarket
	_ = ctx
	panic("not implemented")
}

func (k MsgServer) GovManageFees(goCtx context.Context, msg *exchange.MsgGovManageFeesRequest) (*exchange.MsgGovManageFeesResponse, error) {
	if msg.Authority != k.GetAuthority() {
		return nil, cerrs.Wrapf(govtypes.ErrInvalidSigner, "expected %s got %s", k.GetAuthority(), msg.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO[1658]: Implement GovManageFees
	_ = ctx
	panic("not implemented")
}

// GovUpdateParams is a governance proposal endpoint for updating the exchange module's params.
func (k MsgServer) GovUpdateParams(goCtx context.Context, msg *exchange.MsgGovUpdateParamsRequest) (*exchange.MsgGovUpdateParamsResponse, error) {
	if msg.Authority != k.GetAuthority() {
		return nil, cerrs.Wrapf(govtypes.ErrInvalidSigner, "expected %s got %s", k.GetAuthority(), msg.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.SetParams(ctx, &msg.Params)

	if err := ctx.EventManager().EmitTypedEvent(exchange.NewEventParamsUpdated()); err != nil {
		return nil, err
	}
	return &exchange.MsgGovUpdateParamsResponse{}, nil
}
