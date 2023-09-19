package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

// CreateAsk creates an ask order (to sell something you own).
func (k MsgServer) CreateAsk(goCtx context.Context, msg *exchange.MsgCreateAskRequest) (*exchange.MsgCreateAskResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	orderID, err := k.CreateAskOrder(ctx, msg.AskOrder, msg.OrderCreationFee)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgCreateAskResponse{OrderId: orderID}, nil
}

// CreateBid creates a bid order (to buy something you want).
func (k MsgServer) CreateBid(goCtx context.Context, msg *exchange.MsgCreateBidRequest) (*exchange.MsgCreateBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	orderID, err := k.CreateBidOrder(ctx, msg.BidOrder, msg.OrderCreationFee)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgCreateBidResponse{OrderId: orderID}, nil
}

// CancelOrder cancels an order.
func (k MsgServer) CancelOrder(goCtx context.Context, msg *exchange.MsgCancelOrderRequest) (*exchange.MsgCancelOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := k.Keeper.CancelOrder(ctx, msg.OrderId, msg.Signer)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgCancelOrderResponse{}, nil
}

// FillBids uses the assets in your account to fulfill one or more bids (similar to a fill-or-cancel ask).
func (k MsgServer) FillBids(goCtx context.Context, msg *exchange.MsgFillBidsRequest) (*exchange.MsgFillBidsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := k.Keeper.FillBids(ctx, msg)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgFillBidsResponse{}, nil
}

// FillAsks uses the funds in your account to fulfill one or more asks (similar to a fill-or-cancel bid).
func (k MsgServer) FillAsks(goCtx context.Context, msg *exchange.MsgFillAsksRequest) (*exchange.MsgFillAsksResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	err := k.Keeper.FillAsks(ctx, msg)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgFillAsksResponse{}, nil
}

// permError creates and returns an error indicating that an account does not have a needed permission.
func permError(desc string, account string, marketID uint32) error {
	return sdkerrors.ErrInvalidRequest.Wrapf("account %s does not have permission to %s market %d", account, desc, marketID)
}

// MarketSettle is a market endpoint to trigger the settlement of orders.
func (k MsgServer) MarketSettle(goCtx context.Context, msg *exchange.MsgMarketSettleRequest) (*exchange.MsgMarketSettleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.CanSettleOrders(ctx, msg.MarketId, msg.Admin) {
		return nil, permError("settle orders for", msg.Admin, msg.MarketId)
	}
	err := k.SettleOrders(ctx, msg.MarketId, msg.AskOrderIds, msg.BidOrderIds, msg.ExpectPartial)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgMarketSettleResponse{}, nil
}

// MarketWithdraw is a market endpoint to withdraw fees that have been collected.
func (k MsgServer) MarketWithdraw(goCtx context.Context, msg *exchange.MsgMarketWithdrawRequest) (*exchange.MsgMarketWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.CanWithdrawMarketFunds(ctx, msg.MarketId, msg.Admin) {
		return nil, permError("withdraw from", msg.Admin, msg.MarketId)
	}
	admin := sdk.MustAccAddressFromBech32(msg.Admin)
	toAddr := sdk.MustAccAddressFromBech32(msg.ToAddress)
	err := k.WithdrawMarketFunds(ctx, msg.MarketId, toAddr, msg.Amount, admin)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgMarketWithdrawResponse{}, nil
}

// MarketUpdateDetails is a market endpoint to update its details.
func (k MsgServer) MarketUpdateDetails(goCtx context.Context, msg *exchange.MsgMarketUpdateDetailsRequest) (*exchange.MsgMarketUpdateDetailsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.CanUpdateMarket(ctx, msg.MarketId, msg.Admin) {
		return nil, permError("update", msg.Admin, msg.MarketId)
	}
	admin := sdk.MustAccAddressFromBech32(msg.Admin)
	err := k.UpdateMarketDetails(ctx, msg.MarketId, &msg.MarketDetails, admin)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgMarketUpdateDetailsResponse{}, nil
}

// MarketUpdateEnabled is a market endpoint to update whether its accepting orders.
func (k MsgServer) MarketUpdateEnabled(goCtx context.Context, msg *exchange.MsgMarketUpdateEnabledRequest) (*exchange.MsgMarketUpdateEnabledResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.CanUpdateMarket(ctx, msg.MarketId, msg.Admin) {
		return nil, permError("update", msg.Admin, msg.MarketId)
	}
	admin := sdk.MustAccAddressFromBech32(msg.Admin)
	err := k.UpdateMarketActive(ctx, msg.MarketId, msg.AcceptingOrders, admin)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgMarketUpdateEnabledResponse{}, nil
}

// MarketUpdateUserSettle is a market endpoint to update whether it allows user-initiated settlement.
func (k MsgServer) MarketUpdateUserSettle(goCtx context.Context, msg *exchange.MsgMarketUpdateUserSettleRequest) (*exchange.MsgMarketUpdateUserSettleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.CanUpdateMarket(ctx, msg.MarketId, msg.Admin) {
		return nil, permError("update", msg.Admin, msg.MarketId)
	}
	admin := sdk.MustAccAddressFromBech32(msg.Admin)
	err := k.UpdateUserSettlementAllowed(ctx, msg.MarketId, msg.AllowUserSettlement, admin)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgMarketUpdateUserSettleResponse{}, nil
}

// MarketManagePermissions is a market endpoint to manage a market's user permissions.
func (k MsgServer) MarketManagePermissions(goCtx context.Context, msg *exchange.MsgMarketManagePermissionsRequest) (*exchange.MsgMarketManagePermissionsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.CanManagePermissions(ctx, msg.MarketId, msg.Admin) {
		return nil, permError("manage permissions for", msg.Admin, msg.MarketId)
	}
	err := k.UpdatePermissions(ctx, msg)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgMarketManagePermissionsResponse{}, nil
}

// MarketManageReqAttrs is a market endpoint to manage the attributes required to interact with it.
func (k MsgServer) MarketManageReqAttrs(goCtx context.Context, msg *exchange.MsgMarketManageReqAttrsRequest) (*exchange.MsgMarketManageReqAttrsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.CanWithdrawMarketFunds(ctx, msg.MarketId, msg.Admin) {
		return nil, permError("manage required attributes for", msg.Admin, msg.MarketId)
	}
	err := k.UpdateReqAttrs(ctx, msg)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgMarketManageReqAttrsResponse{}, nil
}

// wrongAuthErr returns the error to use when a message's authority isn't what's required.
func (k MsgServer) wrongAuthErr(badAuthority string) error {
	return govtypes.ErrInvalidSigner.Wrapf("expected %s got %s", k.GetAuthority(), badAuthority)
}

// GovCreateMarket is a governance proposal endpoint for creating a market.
func (k MsgServer) GovCreateMarket(goCtx context.Context, msg *exchange.MsgGovCreateMarketRequest) (*exchange.MsgGovCreateMarketResponse, error) {
	if !k.IsAuthority(msg.Authority) {
		return nil, k.wrongAuthErr(msg.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	marketID, err := k.CreateMarket(ctx, msg.Market)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	err = ctx.EventManager().EmitTypedEvent(exchange.NewEventMarketCreated(marketID))
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgGovCreateMarketResponse{}, nil
}

// GovManageFees is a governance proposal endpoint for updating a market's fees.
func (k MsgServer) GovManageFees(goCtx context.Context, msg *exchange.MsgGovManageFeesRequest) (*exchange.MsgGovManageFeesResponse, error) {
	if !k.IsAuthority(msg.Authority) {
		return nil, k.wrongAuthErr(msg.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	err := k.UpdateFees(ctx, msg)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	return &exchange.MsgGovManageFeesResponse{}, nil
}

// GovUpdateParams is a governance proposal endpoint for updating the exchange module's params.
func (k MsgServer) GovUpdateParams(goCtx context.Context, msg *exchange.MsgGovUpdateParamsRequest) (*exchange.MsgGovUpdateParamsResponse, error) {
	if !k.IsAuthority(msg.Authority) {
		return nil, k.wrongAuthErr(msg.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.SetParams(ctx, &msg.Params)

	if err := ctx.EventManager().EmitTypedEvent(exchange.NewEventParamsUpdated()); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgGovUpdateParamsResponse{}, nil
}
