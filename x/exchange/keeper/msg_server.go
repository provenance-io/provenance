package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

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

// CommitFunds marks funds in an account as manageable by a market.
func (k MsgServer) CommitFunds(goCtx context.Context, msg *exchange.MsgCommitFundsRequest) (*exchange.MsgCommitFundsResponse, error) {
	// TODO[1789]: MsgServer.CommitFunds
	_, _ = goCtx, msg
	panic("not implemented")
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

// MarketCommitmentSettle is a market endpoint to transfer committed funds.
func (k MsgServer) MarketCommitmentSettle(goCtx context.Context, msg *exchange.MsgMarketCommitmentSettleRequest) (*exchange.MsgMarketCommitmentSettleResponse, error) {
	// TODO[1789]: MsgServer.MarketCommitmentSettle
	_, _ = goCtx, msg
	panic("not implemented")
}

// MarketReleaseCommitments is a market endpoint return control of funds back to the account owner(s).
func (k MsgServer) MarketReleaseCommitments(goCtx context.Context, msg *exchange.MsgMarketReleaseCommitmentsRequest) (*exchange.MsgMarketReleaseCommitmentsResponse, error) {
	// TODO[1789]: MsgServer.MarketReleaseCommitments
	_, _ = goCtx, msg
	panic("not implemented")
}

// MarketSetOrderExternalID updates an order's external id field.
func (k MsgServer) MarketSetOrderExternalID(goCtx context.Context, msg *exchange.MsgMarketSetOrderExternalIDRequest) (*exchange.MsgMarketSetOrderExternalIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.CanSetIDs(ctx, msg.MarketId, msg.Admin) {
		return nil, permError("set external ids on orders for", msg.Admin, msg.MarketId)
	}
	err := k.SetOrderExternalID(ctx, msg.MarketId, msg.OrderId, msg.ExternalId)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgMarketSetOrderExternalIDResponse{}, nil
}

// MarketWithdraw is a market endpoint to withdraw fees that have been collected.
func (k MsgServer) MarketWithdraw(goCtx context.Context, msg *exchange.MsgMarketWithdrawRequest) (*exchange.MsgMarketWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.CanWithdrawMarketFunds(ctx, msg.MarketId, msg.Admin) {
		return nil, permError("withdraw from", msg.Admin, msg.MarketId)
	}
	toAddr := sdk.MustAccAddressFromBech32(msg.ToAddress)
	err := k.WithdrawMarketFunds(ctx, msg.MarketId, toAddr, msg.Amount, msg.Admin)
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
	err := k.UpdateMarketDetails(ctx, msg.MarketId, msg.MarketDetails, msg.Admin)
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
	err := k.UpdateMarketActive(ctx, msg.MarketId, msg.AcceptingOrders, msg.Admin)
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
	err := k.UpdateUserSettlementAllowed(ctx, msg.MarketId, msg.AllowUserSettlement, msg.Admin)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgMarketUpdateUserSettleResponse{}, nil
}

// MarketUpdateAllowCommitments is a market endpoint to update whether it accepts commitments.
func (k MsgServer) MarketUpdateAllowCommitments(goCtx context.Context, msg *exchange.MsgMarketUpdateAllowCommitmentsRequest) (*exchange.MsgMarketUpdateAllowCommitmentsResponse, error) {
	// TODO[1789]: MsgServer.MarketUpdateAllowCommitments
	_, _ = goCtx, msg
	panic("not implemented")
}

// MarketUpdateIntermediaryDenom sets a market's intermediary denom.
func (k MsgServer) MarketUpdateIntermediaryDenom(goCtx context.Context, msg *exchange.MsgMarketUpdateIntermediaryDenomRequest) (*exchange.MsgMarketUpdateIntermediaryDenomResponse, error) {
	// TODO[1789]: MsgServer.MarketUpdateIntermediaryDenom
	_, _ = goCtx, msg
	panic("not implemented")
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
	if !k.CanManageReqAttrs(ctx, msg.MarketId, msg.Admin) {
		return nil, permError("manage required attributes for", msg.Admin, msg.MarketId)
	}
	err := k.UpdateReqAttrs(ctx, msg)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgMarketManageReqAttrsResponse{}, nil
}

// GovCreateMarket is a governance proposal endpoint for creating a market.
func (k MsgServer) GovCreateMarket(goCtx context.Context, msg *exchange.MsgGovCreateMarketRequest) (*exchange.MsgGovCreateMarketResponse, error) {
	if err := k.ValidateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	_, err := k.CreateMarket(ctx, msg.Market)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	return &exchange.MsgGovCreateMarketResponse{}, nil
}

// GovManageFees is a governance proposal endpoint for updating a market's fees.
func (k MsgServer) GovManageFees(goCtx context.Context, msg *exchange.MsgGovManageFeesRequest) (*exchange.MsgGovManageFeesResponse, error) {
	if err := k.ValidateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.UpdateFees(ctx, msg)

	return &exchange.MsgGovManageFeesResponse{}, nil
}

// GovCloseMarket is a governance proposal endpoint that will disable order and commitment creation,
func (k MsgServer) GovCloseMarket(goCtx context.Context, msg *exchange.MsgGovCloseMarketRequest) (*exchange.MsgGovCloseMarketResponse, error) {
	// TODO[1789]: MsgServer.GovCloseMarket
	_, _ = goCtx, msg
	panic("not implemented")
}

// GovUpdateParams is a governance proposal endpoint for updating the exchange module's params.
func (k MsgServer) GovUpdateParams(goCtx context.Context, msg *exchange.MsgGovUpdateParamsRequest) (*exchange.MsgGovUpdateParamsResponse, error) {
	if err := k.ValidateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.SetParams(ctx, &msg.Params)
	k.emitEvent(ctx, exchange.NewEventParamsUpdated())

	return &exchange.MsgGovUpdateParamsResponse{}, nil
}
