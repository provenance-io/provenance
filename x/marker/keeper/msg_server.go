package keeper

import (
	"context"
	"fmt"
	"regexp"

	"github.com/armon/go-metrics"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/x/marker/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the marker MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// Handle a message to add a new marker account.
func (k msgServer) AddMarker(goCtx context.Context, msg *types.MsgAddMarkerRequest) (*types.MsgAddMarkerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate transaction message.
	err := msg.ValidateBasic()
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	if msg.Status >= types.StatusActive {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "a marker can not be created in an ACTIVE status")
	}

	params := k.GetParams(ctx)

	// Add marker requests must pass extra validation for denom (in addition to regular coin validation expression)
	valExp := params.UnrestrictedDenomRegex
	if v, expErr := regexp.Compile(valExp); expErr != nil {
		return nil, expErr
	} else if !v.MatchString(msg.Amount.Denom) {
		return nil, fmt.Errorf("invalid denom (fails unrestricted marker denom validation %s): %s", valExp, msg.Amount.Denom)
	}

	addr := types.MustGetMarkerAddress(msg.Amount.Denom)
	var manager sdk.AccAddress
	if msg.Manager != "" {
		manager, err = sdk.AccAddressFromBech32(msg.Manager)
	} else {
		manager, err = sdk.AccAddressFromBech32(msg.FromAddress)
	}
	if err != nil {
		return nil, err
	}
	account := authtypes.NewBaseAccount(addr, nil, 0, 0)
	ma := types.NewMarkerAccount(
		account,
		sdk.NewCoin(msg.Amount.Denom, msg.Amount.Amount),
		manager,
		msg.AccessList,
		msg.Status,
		msg.MarkerType)
	ma.SupplyFixed = msg.SupplyFixed

	if params.EnableGovernance {
		ma.AllowGovernanceControl = true
	} else {
		ma.AllowGovernanceControl = msg.AllowGovernanceControl
	}

	if err := k.Keeper.AddMarkerAccount(ctx, ma); err != nil {
		ctx.Logger().Error("unable to add marker", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	markerAddEvent := types.NewEventMarkerAdd(msg.Amount.Denom, msg.Amount.Amount.String(), msg.Status.String(), msg.Manager, msg.MarkerType.String())
	if err := ctx.EventManager().EmitTypedEvent(markerAddEvent); err != nil {
		return nil, err
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, "add", "marker"},
			1,
			[]metrics.Label{
				telemetry.NewLabel(types.EventTelemetryLabelAccess, markerAddEvent.Amount),
				telemetry.NewLabel(types.EventTelemetryLabelDenom, markerAddEvent.Denom),
				telemetry.NewLabel(types.EventTelemetryLabelStatus, markerAddEvent.Status),
				telemetry.NewLabel(types.EventTelemetryLabelManager, markerAddEvent.Manager),
				telemetry.NewLabel(types.EventTelemetryLabelMarkerType, markerAddEvent.MarkerType),
			},
		)
	}()

	return &types.MsgAddMarkerResponse{}, nil
}

// AddAccess handles a message to grant access to a marker account.
func (k msgServer) AddAccess(goCtx context.Context, msg *types.MsgAddAccessRequest) (*types.MsgAddAccessResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate transaction message.
	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	for i := range msg.Access {
		if err := k.Keeper.AddAccess(ctx, msg.GetSigners()[0], msg.Denom, &msg.Access[i]); err != nil {
			ctx.Logger().Error("unable to add access grant to marker", "err", err)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, err.Error())
		}

		markerAddAccessEvent := types.NewEventMarkerAddAccess(msg.Access[i], msg.Denom, msg.Administrator)
		if err := ctx.EventManager().EmitTypedEvent(markerAddAccessEvent); err != nil {
			return nil, err
		}

		defer func() {
			telemetry.IncrCounterWithLabels(
				[]string{types.ModuleName, "add", "access"},
				1,
				[]metrics.Label{
					telemetry.NewLabel(types.EventTelemetryLabelAccess, "TODO"),
					telemetry.NewLabel(types.EventTelemetryLabelDenom, markerAddAccessEvent.Denom),
					telemetry.NewLabel(types.EventTelemetryLabelAdministrator, markerAddAccessEvent.Administrator),
				},
			)
		}()

	}

	return &types.MsgAddAccessResponse{}, nil
}

// DeleteAccess handles a message to revoke access to marker account.
func (k msgServer) DeleteAccess(goCtx context.Context, msg *types.MsgDeleteAccessRequest) (*types.MsgDeleteAccessResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate transaction message.
	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	addr, err := sdk.AccAddressFromBech32(msg.RemovedAddress)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, err.Error())
	}

	if err := k.Keeper.RemoveAccess(ctx, msg.GetSigners()[0], msg.Denom, addr); err != nil {
		ctx.Logger().Error("unable to remove access grant from marker", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, err.Error())
	}

	markerDeleteAccessEvent := types.NewEventMarkerDeleteAccess(msg.RemovedAddress, msg.Denom, msg.Administrator)
	if err := ctx.EventManager().EmitTypedEvent(markerDeleteAccessEvent); err != nil {
		return nil, err
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, "delete", "access"},
			1,
			[]metrics.Label{
				telemetry.NewLabel(types.EventTelemetryAddress, markerDeleteAccessEvent.RemoveAddress),
				telemetry.NewLabel(types.EventTelemetryLabelDenom, markerDeleteAccessEvent.Denom),
				telemetry.NewLabel(types.EventTelemetryLabelAdministrator, markerDeleteAccessEvent.Administrator),
			},
		)
	}()

	return &types.MsgDeleteAccessResponse{}, nil
}

// Finalize handles a message to finalize a marker
func (k msgServer) Finalize(goCtx context.Context, msg *types.MsgFinalizeRequest) (*types.MsgFinalizeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate transaction message.
	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	if err := k.Keeper.FinalizeMarker(ctx, msg.GetSigners()[0], msg.Denom); err != nil {
		ctx.Logger().Error("unable to finalize marker", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	markerFinalizeEvent := types.NewEventMarkerFinalize(msg.Denom, msg.Administrator)
	if err := ctx.EventManager().EmitTypedEvent(markerFinalizeEvent); err != nil {
		return nil, err
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, "finalize", "marker"},
			1,
			[]metrics.Label{
				telemetry.NewLabel(types.EventTelemetryLabelDenom, markerFinalizeEvent.Denom),
				telemetry.NewLabel(types.EventTelemetryLabelAdministrator, markerFinalizeEvent.Administrator),
			},
		)
	}()

	return &types.MsgFinalizeResponse{}, nil
}

// Activate handles a message to activate a marker
func (k msgServer) Activate(goCtx context.Context, msg *types.MsgActivateRequest) (*types.MsgActivateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate transaction message.
	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	if err := k.Keeper.ActivateMarker(ctx, msg.GetSigners()[0], msg.Denom); err != nil {
		ctx.Logger().Error("unable to activate marker", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	markerActivateEvent := types.NewEventMarkerActivate(msg.Denom, msg.Administrator)
	if err := ctx.EventManager().EmitTypedEvent(markerActivateEvent); err != nil {
		return nil, err
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, "activate", "marker"},
			1,
			[]metrics.Label{
				telemetry.NewLabel(types.EventTelemetryLabelDenom, markerActivateEvent.Denom),
				telemetry.NewLabel(types.EventTelemetryLabelAdministrator, markerActivateEvent.Administrator),
			},
		)
	}()

	return &types.MsgActivateResponse{}, nil
}

// Cancel handles a message to cancel a marker
func (k msgServer) Cancel(goCtx context.Context, msg *types.MsgCancelRequest) (*types.MsgCancelResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate transaction message.
	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	if err := k.Keeper.CancelMarker(ctx, msg.GetSigners()[0], msg.Denom); err != nil {
		ctx.Logger().Error("unable to cancel marker", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	markerCancelEvent := types.NewEventMarkerCancel(msg.Denom, msg.Administrator)
	if err := ctx.EventManager().EmitTypedEvent(markerCancelEvent); err != nil {
		return nil, err
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, "cancel", "marker"},
			1,
			[]metrics.Label{
				telemetry.NewLabel(types.EventTelemetryLabelDenom, markerCancelEvent.Denom),
				telemetry.NewLabel(types.EventTelemetryLabelAdministrator, markerCancelEvent.Administrator),
			},
		)
	}()

	return &types.MsgCancelResponse{}, nil
}

// Delete handles a message to delete a marker
func (k msgServer) Delete(goCtx context.Context, msg *types.MsgDeleteRequest) (*types.MsgDeleteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate transaction message.
	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	if err := k.Keeper.DeleteMarker(ctx, msg.GetSigners()[0], msg.Denom); err != nil {
		ctx.Logger().Error("unable to delete marker", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	markerDeleteEvent := types.NewEventMarkerDelete(msg.Denom, msg.Administrator)
	if err := ctx.EventManager().EmitTypedEvent(markerDeleteEvent); err != nil {
		return nil, err
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, "delete", "marker"},
			1,
			[]metrics.Label{
				telemetry.NewLabel(types.EventTelemetryLabelDenom, markerDeleteEvent.Denom),
				telemetry.NewLabel(types.EventTelemetryLabelAdministrator, markerDeleteEvent.Administrator),
			},
		)
	}()

	return &types.MsgDeleteResponse{}, nil
}

// Mint handles a message to mint additional supply for a marker.
func (k msgServer) Mint(goCtx context.Context, msg *types.MsgMintRequest) (*types.MsgMintResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate transaction message.
	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	if err := k.Keeper.MintCoin(ctx, msg.GetSigners()[0], msg.Amount); err != nil {
		ctx.Logger().Error("unable to mint coin for marker", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	markerMintEvent := types.NewEventMarkerMint(msg.Amount.Amount.String(), msg.Amount.GetDenom(), msg.Administrator)
	if err := ctx.EventManager().EmitTypedEvent(markerMintEvent); err != nil {
		return nil, err
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, "mint", "marker"},
			1,
			[]metrics.Label{
				telemetry.NewLabel(types.EventTelemetryLabelAmount, markerMintEvent.Amount),
				telemetry.NewLabel(types.EventTelemetryLabelDenom, markerMintEvent.Denom),
				telemetry.NewLabel(types.EventTelemetryLabelAdministrator, markerMintEvent.Administrator),
			},
		)
	}()

	return &types.MsgMintResponse{}, nil
}

// Burn handles a message to burn coin from a marker account.
func (k msgServer) Burn(goCtx context.Context, msg *types.MsgBurnRequest) (*types.MsgBurnResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate transaction message.
	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	if err := k.Keeper.BurnCoin(ctx, msg.GetSigners()[0], msg.Amount); err != nil {
		ctx.Logger().Error("unable to burn coin from marker", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	markerBurnEvent := types.NewEventMarkerBurn(msg.Amount.Amount.String(), msg.Amount.GetDenom(), msg.Administrator)
	if err := ctx.EventManager().EmitTypedEvent(markerBurnEvent); err != nil {
		return nil, err
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, "burn", "marker"},
			1,
			[]metrics.Label{
				telemetry.NewLabel(types.EventTelemetryLabelAmount, markerBurnEvent.Amount),
				telemetry.NewLabel(types.EventTelemetryLabelDenom, markerBurnEvent.Denom),
				telemetry.NewLabel(types.EventTelemetryLabelAdministrator, markerBurnEvent.Administrator),
			},
		)
	}()

	return &types.MsgBurnResponse{}, nil
}

// Withdraw handles a message to withdraw coins from the marker account.
func (k msgServer) Withdraw(goCtx context.Context, msg *types.MsgWithdrawRequest) (*types.MsgWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate transaction message.
	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	to, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, err
	}

	if err := k.Keeper.WithdrawCoins(ctx, msg.GetSigners()[0], to, msg.Denom, msg.Amount); err != nil {
		ctx.Logger().Error("unable to withdraw coins from marker", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	markerWithdrawEvent := types.NewEventMarkerWithdraw(msg.Amount.String(), msg.GetDenom(), msg.Administrator, msg.ToAddress)
	if err := ctx.EventManager().EmitTypedEvent(markerWithdrawEvent); err != nil {
		return nil, err
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, "withdraw", "marker"},
			1,
			[]metrics.Label{
				telemetry.NewLabel(types.EventTelemetryToAddress, markerWithdrawEvent.ToAddress),
				telemetry.NewLabel(types.EventTelemetryLabelAmount, markerWithdrawEvent.Coins),
				telemetry.NewLabel(types.EventTelemetryLabelDenom, markerWithdrawEvent.Denom),
				telemetry.NewLabel(types.EventTelemetryLabelAdministrator, markerWithdrawEvent.Administrator),
			},
		)
	}()

	return &types.MsgWithdrawResponse{}, nil
}

// Transfer handles a message to send coins from one account to another (used with restricted coins that are not
//	sent using the normal bank send process)
func (k msgServer) Transfer(goCtx context.Context, msg *types.MsgTransferRequest) (*types.MsgTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate transaction message.
	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}
	to, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, err
	}
	admin, err := sdk.AccAddressFromBech32(msg.Administrator)
	if err != nil {
		return nil, err
	}

	err = k.TransferCoin(ctx, from, to, admin, msg.Amount)
	if err != nil {
		return nil, err
	}

	markerTransferEvent := types.NewEventMarkerTransfer(msg.Amount.Amount.String(), msg.Amount.Denom, msg.Administrator, msg.ToAddress, msg.FromAddress)
	if err := ctx.EventManager().EmitTypedEvent(markerTransferEvent); err != nil {
		return nil, err
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, "transfer", "marker"},
			1,
			[]metrics.Label{
				telemetry.NewLabel(types.EventTelemetryToAddress, markerTransferEvent.ToAddress),
				telemetry.NewLabel(types.EventTelemetryFromAddress, markerTransferEvent.FromAddress),
				telemetry.NewLabel(types.EventTelemetryLabelAmount, markerTransferEvent.Amount),
				telemetry.NewLabel(types.EventTelemetryLabelDenom, markerTransferEvent.Denom),
				telemetry.NewLabel(types.EventTelemetryLabelAdministrator, markerTransferEvent.Administrator),
			},
		)
	}()
	return &types.MsgTransferResponse{}, nil
}

// SetDenomMetadata handles a message setting metadata for a marker with the specified denom.
func (k msgServer) SetDenomMetadata(
	goCtx context.Context,
	msg *types.MsgSetDenomMetadataRequest,
) (*types.MsgSetDenomMetadataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate transaction message.
	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	admin, addrErr := sdk.AccAddressFromBech32(msg.Administrator)
	if addrErr != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, addrErr.Error())
	}

	err := k.SetMarkerDenomMetadata(ctx, msg.Metadata, admin)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)
	return &types.MsgSetDenomMetadataResponse{}, nil
}
