package keeper

import (
	"context"
	"errors"
	"fmt"

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
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, _ := sdk.AccAddressFromBech32(msg.Account)

	err := k.ValidateAndCollectCommitmentCreationFee(ctx, msg.MarketId, addr, msg.CreationFee)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	err = k.AddCommitment(ctx, msg.MarketId, addr, msg.Amount, msg.EventTag)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgCommitFundsResponse{}, nil
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
	err := k.SettleOrders(ctx, msg)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgMarketSettleResponse{}, nil
}

// MarketCommitmentSettle is a market endpoint to transfer committed funds.
func (k MsgServer) MarketCommitmentSettle(goCtx context.Context, msg *exchange.MsgMarketCommitmentSettleRequest) (*exchange.MsgMarketCommitmentSettleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.CanSettleCommitments(ctx, msg.MarketId, msg.Admin) {
		return nil, permError("settle commitments for", msg.Admin, msg.MarketId)
	}
	err := k.SettleCommitments(ctx, msg)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	err = k.consumeCommitmentSettlementFee(ctx, msg)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgMarketCommitmentSettleResponse{}, nil
}

// MarketReleaseCommitments is a market endpoint return control of funds back to the account owner(s).
func (k MsgServer) MarketReleaseCommitments(goCtx context.Context, msg *exchange.MsgMarketReleaseCommitmentsRequest) (*exchange.MsgMarketReleaseCommitmentsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.CanReleaseCommitmentsForMarket(ctx, msg.MarketId, msg.Admin) {
		return nil, permError("release commitments for", msg.Admin, msg.MarketId)
	}
	err := k.ReleaseCommitments(ctx, msg.MarketId, msg.ToRelease, msg.EventTag)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgMarketReleaseCommitmentsResponse{}, nil
}

// MarketTransferCommitments is a market endpoint transfer funds from one market to another market back to the account owner(s).
func (k MsgServer) MarketTransferCommitments(goCtx context.Context, msg *exchange.MsgMarketTransferCommitmentsRequest) (*exchange.MsgMarketTransferCommitmentsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.CanTransferCommitmentsForMarket(ctx, msg.CurrentMarketId, msg.Admin) {
		return nil, permError("transfer commitments for", msg.Admin, msg.CurrentMarketId)
	}
	err := k.TransferCommitments(ctx, msg)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgMarketTransferCommitmentsResponse{}, nil
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
//
//nolint:staticcheck // This endpoint needs to keep existing, so use of the deprecated messages is needed.
func (k MsgServer) MarketUpdateEnabled(_ context.Context, _ *exchange.MsgMarketUpdateEnabledRequest) (*exchange.MsgMarketUpdateEnabledResponse, error) {
	return nil, fmt.Errorf("the MarketUpdateEnabled endpoint has been replaced by the MarketUpdateAcceptingOrders endpoint")
}

// MarketUpdateAcceptingOrders is a market endpoint to update whether its accepting orders.
func (k MsgServer) MarketUpdateAcceptingOrders(goCtx context.Context, msg *exchange.MsgMarketUpdateAcceptingOrdersRequest) (*exchange.MsgMarketUpdateAcceptingOrdersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.CanUpdateMarket(ctx, msg.MarketId, msg.Admin) {
		return nil, permError("update", msg.Admin, msg.MarketId)
	}
	err := k.UpdateMarketAcceptingOrders(ctx, msg.MarketId, msg.AcceptingOrders, msg.Admin)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgMarketUpdateAcceptingOrdersResponse{}, nil
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

// MarketUpdateAcceptingCommitments is a market endpoint to update whether it accepts commitments.
func (k MsgServer) MarketUpdateAcceptingCommitments(goCtx context.Context, msg *exchange.MsgMarketUpdateAcceptingCommitmentsRequest) (*exchange.MsgMarketUpdateAcceptingCommitmentsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.CanUpdateMarket(ctx, msg.MarketId, msg.Admin) {
		return nil, permError("update", msg.Admin, msg.MarketId)
	}
	if !k.IsAuthority(msg.Admin) {
		if err := validateMarketUpdateAcceptingCommitments(k.getStore(ctx), msg.MarketId, msg.AcceptingCommitments); err != nil {
			return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
		}
	}

	err := k.UpdateMarketAcceptingCommitments(ctx, msg.MarketId, msg.AcceptingCommitments, msg.Admin)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	return &exchange.MsgMarketUpdateAcceptingCommitmentsResponse{}, nil
}

// MarketUpdateIntermediaryDenom sets a market's intermediary denom.
func (k MsgServer) MarketUpdateIntermediaryDenom(goCtx context.Context, msg *exchange.MsgMarketUpdateIntermediaryDenomRequest) (*exchange.MsgMarketUpdateIntermediaryDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !k.CanUpdateMarket(ctx, msg.MarketId, msg.Admin) {
		return nil, permError("update", msg.Admin, msg.MarketId)
	}
	k.UpdateIntermediaryDenom(ctx, msg.MarketId, msg.IntermediaryDenom, msg.Admin)
	return &exchange.MsgMarketUpdateIntermediaryDenomResponse{}, nil
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

// CreatePayment creates a payment to facilitate a trade between two accounts.
func (k MsgServer) CreatePayment(goCtx context.Context, msg *exchange.MsgCreatePaymentRequest) (*exchange.MsgCreatePaymentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.Keeper.CreatePayment(ctx, &msg.Payment); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	if !msg.Payment.SourceAmount.IsZero() {
		k.consumeCreatePaymentFee(ctx, msg)
	}
	return &exchange.MsgCreatePaymentResponse{}, nil
}

// AcceptPayment is used by a target to accept a payment.
func (k MsgServer) AcceptPayment(goCtx context.Context, msg *exchange.MsgAcceptPaymentRequest) (*exchange.MsgAcceptPaymentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.Keeper.AcceptPayment(ctx, &msg.Payment); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	if !msg.Payment.TargetAmount.IsZero() {
		k.consumeAcceptPaymentFee(ctx, msg)
	}
	return &exchange.MsgAcceptPaymentResponse{}, nil
}

// RejectPayment can be used by a target to reject a payment.
func (k MsgServer) RejectPayment(goCtx context.Context, msg *exchange.MsgRejectPaymentRequest) (*exchange.MsgRejectPaymentResponse, error) {
	target, err := sdk.AccAddressFromBech32(msg.Target)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid target %q: %v", msg.Target, err)
	}
	source, err := sdk.AccAddressFromBech32(msg.Source)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid source %q: %v", msg.Source, err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	err = k.Keeper.RejectPayment(ctx, target, source, msg.ExternalId)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	return &exchange.MsgRejectPaymentResponse{}, nil
}

// RejectPayments can be used by a target to reject all payments from one or more sources.
func (k MsgServer) RejectPayments(goCtx context.Context, msg *exchange.MsgRejectPaymentsRequest) (*exchange.MsgRejectPaymentsResponse, error) {
	target, err := sdk.AccAddressFromBech32(msg.Target)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid target %q: %v", msg.Target, err)
	}
	sources := make([]sdk.AccAddress, 0, len(msg.Sources))
	for i, sourceStr := range msg.Sources {
		var source sdk.AccAddress
		source, err = sdk.AccAddressFromBech32(sourceStr)
		if err != nil {
			return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid sources[%d] %q: %v", i, sourceStr, err)
		}
		sources = append(sources, source)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	err = k.Keeper.RejectPayments(ctx, target, sources)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	return &exchange.MsgRejectPaymentsResponse{}, nil
}

// CancelPayments can be used by a source to cancel one or more payments.
func (k MsgServer) CancelPayments(goCtx context.Context, msg *exchange.MsgCancelPaymentsRequest) (*exchange.MsgCancelPaymentsResponse, error) {
	source, err := sdk.AccAddressFromBech32(msg.Source)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid source %q: %v", msg.Source, err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	err = k.Keeper.CancelPayments(ctx, source, msg.ExternalIds)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	return &exchange.MsgCancelPaymentsResponse{}, nil
}

// ChangePaymentTarget can be used by a source to change the target in one of their payments.
func (k MsgServer) ChangePaymentTarget(goCtx context.Context, msg *exchange.MsgChangePaymentTargetRequest) (*exchange.MsgChangePaymentTargetResponse, error) {
	source, err := sdk.AccAddressFromBech32(msg.Source)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid source %q: %v", msg.Source, err)
	}
	var newTarget sdk.AccAddress
	if len(msg.NewTarget) > 0 {
		newTarget, err = sdk.AccAddressFromBech32(msg.NewTarget)
		if err != nil {
			return nil, sdkerrors.ErrInvalidRequest.Wrapf("invalid new target %q: %v", msg.NewTarget, err)
		}
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	err = k.UpdatePaymentTarget(ctx, source, msg.ExternalId, newTarget)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	return &exchange.MsgChangePaymentTargetResponse{}, nil
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
// cancel all orders, and release all commitments.
func (k MsgServer) GovCloseMarket(goCtx context.Context, msg *exchange.MsgGovCloseMarketRequest) (*exchange.MsgGovCloseMarketResponse, error) {
	if err := k.ValidateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.CloseMarket(ctx, msg.MarketId, msg.Authority)

	return &exchange.MsgGovCloseMarketResponse{}, nil
}

// GovUpdateParams is a governance proposal endpoint for updating the exchange module's params.
//
//nolint:staticcheck // SA1019 Suppress warning for deprecated MsgGovUpdateParamsRequest usage
func (k MsgServer) GovUpdateParams(_ context.Context, _ *exchange.MsgGovUpdateParamsRequest) (*exchange.MsgGovUpdateParamsResponse, error) {
	return nil, errors.New("deprecated and unusable")
}

// UpdateParams is a governance proposal endpoint for updating the exchange module's params.
func (k MsgServer) UpdateParams(goCtx context.Context, msg *exchange.MsgUpdateParamsRequest) (*exchange.MsgUpdateParamsResponse, error) {
	if err := k.ValidateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.SetParams(ctx, &msg.Params)
	k.emitEvent(ctx, exchange.NewEventParamsUpdated())

	return &exchange.MsgUpdateParamsResponse{}, nil
}
