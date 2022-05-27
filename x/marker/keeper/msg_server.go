package keeper

import (
	"context"

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

// GrantAllowance grants an allowance from the marker's funds to be used by the grantee.
func (k msgServer) GrantAllowance(goCtx context.Context, msg *types.MsgGrantAllowanceRequest) (*types.MsgGrantAllowanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	m, err := k.GetMarkerByDenom(ctx, msg.Denom)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	admin, err := sdk.AccAddressFromBech32(msg.Administrator)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	grantee, err := sdk.AccAddressFromBech32(msg.Grantee)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	if !m.AddressHasAccess(admin, types.Access_Admin) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "administrator must have admin grant on marker")
	}
	allowance, err := msg.GetFeeAllowanceI()
	if err != nil {
		return nil, err
	}
	err = k.Keeper.feegrantKeeper.GrantAllowance(ctx, m.GetAddress(), grantee, allowance)
	return &types.MsgGrantAllowanceResponse{}, err
}

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

	// Add marker requests must pass extra validation for denom (in addition to regular coin validation expression)
	if err = k.ValidateUnrestictedDenom(ctx, msg.Amount.Denom); err != nil {
		return nil, err
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

	if k.GetEnableGovernance(ctx) {
		ma.AllowGovernanceControl = true
	} else {
		ma.AllowGovernanceControl = msg.AllowGovernanceControl
	}

	if err := k.Keeper.AddMarkerAccount(ctx, ma); err != nil {
		ctx.Logger().Error("unable to add marker", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

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
		access := msg.Access[i]
		if err := k.Keeper.AddAccess(ctx, msg.GetSigners()[0], msg.Denom, &access); err != nil {
			ctx.Logger().Error("unable to add access grant to marker", "err", err)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, err.Error())
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

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

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

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

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

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

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

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

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

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

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, types.EventTelemetryKeyMint},
			1,
			[]metrics.Label{
				telemetry.NewLabel(types.EventTelemetryLabelDenom, msg.Amount.GetDenom()),
				telemetry.NewLabel(types.EventTelemetryLabelAdministrator, msg.Administrator),
			},
		)
		if msg.Amount.Amount.IsInt64() {
			telemetry.SetGaugeWithLabels(
				[]string{types.ModuleName, types.EventTelemetryKeyMint, msg.Amount.Denom},
				float32(msg.Amount.Amount.Int64()),
				[]metrics.Label{telemetry.NewLabel(types.EventTelemetryLabelDenom, msg.Amount.Denom)},
			)
		}
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

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, types.EventTelemetryKeyBurn},
			1,
			[]metrics.Label{
				telemetry.NewLabel(types.EventTelemetryLabelDenom, msg.Amount.GetDenom()),
				telemetry.NewLabel(types.EventTelemetryLabelAdministrator, msg.Administrator),
			},
		)
		if msg.Amount.Amount.IsInt64() {
			telemetry.SetGaugeWithLabels(
				[]string{types.ModuleName, types.EventTelemetryKeyBurn, msg.Amount.Denom},
				float32(msg.Amount.Amount.Int64()),
				[]metrics.Label{telemetry.NewLabel(types.EventTelemetryLabelDenom, msg.Amount.Denom)},
			)
		}
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

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, types.EventTelemetryKeyWithdraw},
			1,
			[]metrics.Label{
				telemetry.NewLabel(types.EventTelemetryLabelToAddress, msg.ToAddress),
				telemetry.NewLabel(types.EventTelemetryLabelDenom, msg.GetDenom()),
				telemetry.NewLabel(types.EventTelemetryLabelAdministrator, msg.Administrator),
			},
		)
		for _, coin := range msg.Amount {
			if coin.Amount.IsInt64() {
				telemetry.SetGaugeWithLabels(
					[]string{types.ModuleName, types.EventTelemetryKeyWithdraw, msg.Denom},
					float32(coin.Amount.Int64()),
					[]metrics.Label{telemetry.NewLabel(types.EventTelemetryLabelDenom, coin.Denom)},
				)
			}
		}
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

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, types.EventTelemetryKeyTransfer},
			1,
			[]metrics.Label{
				telemetry.NewLabel(types.EventTelemetryLabelToAddress, msg.ToAddress),
				telemetry.NewLabel(types.EventTelemetryLabelFromAddress, msg.FromAddress),
				telemetry.NewLabel(types.EventTelemetryLabelDenom, msg.Amount.Denom),
				telemetry.NewLabel(types.EventTelemetryLabelAdministrator, msg.Administrator),
			},
		)
		if msg.Amount.Amount.IsInt64() {
			telemetry.SetGaugeWithLabels(
				[]string{types.ModuleName, types.EventTelemetryKeyTransfer, msg.Amount.Denom},
				float32(msg.Amount.Amount.Int64()),
				[]metrics.Label{telemetry.NewLabel(types.EventTelemetryLabelDenom, msg.Amount.Denom)},
			)
		}
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
