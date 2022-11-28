package keeper

import (
	"context"
	"fmt"
	"time"

	"github.com/armon/go-metrics"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	ibckeeper "github.com/cosmos/ibc-go/v5/modules/apps/transfer/keeper"
	ibctypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"

	intertxtypes "github.com/provenance-io/provenance/x/intertx/types"
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
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	admin, err := sdk.AccAddressFromBech32(msg.Administrator)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	grantee, err := sdk.AccAddressFromBech32(msg.Grantee)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	if !m.AddressHasAccess(admin, types.Access_Admin) {
		return nil, sdkerrors.ErrUnauthorized.Wrap("administrator must have admin grant on marker")
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
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	if msg.Status >= types.StatusActive {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("a marker can not be created in an ACTIVE status")
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
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
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
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	for i := range msg.Access {
		access := msg.Access[i]
		if err := k.Keeper.AddAccess(ctx, msg.GetSigners()[0], msg.Denom, &access); err != nil {
			ctx.Logger().Error("unable to add access grant to marker", "err", err)
			return nil, sdkerrors.ErrUnauthorized.Wrap(err.Error())
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
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	addr, err := sdk.AccAddressFromBech32(msg.RemovedAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrap(err.Error())
	}

	if err := k.Keeper.RemoveAccess(ctx, msg.GetSigners()[0], msg.Denom, addr); err != nil {
		ctx.Logger().Error("unable to remove access grant from marker", "err", err)
		return nil, sdkerrors.ErrUnauthorized.Wrap(err.Error())
	}

	return &types.MsgDeleteAccessResponse{}, nil
}

// Finalize handles a message to finalize a marker
func (k msgServer) Finalize(goCtx context.Context, msg *types.MsgFinalizeRequest) (*types.MsgFinalizeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate transaction message.
	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	if err := k.Keeper.FinalizeMarker(ctx, msg.GetSigners()[0], msg.Denom); err != nil {
		ctx.Logger().Error("unable to finalize marker", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
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
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	if err := k.Keeper.ActivateMarker(ctx, msg.GetSigners()[0], msg.Denom); err != nil {
		ctx.Logger().Error("unable to activate marker", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
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
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	if err := k.Keeper.CancelMarker(ctx, msg.GetSigners()[0], msg.Denom); err != nil {
		ctx.Logger().Error("unable to cancel marker", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
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
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	if err := k.Keeper.DeleteMarker(ctx, msg.GetSigners()[0], msg.Denom); err != nil {
		ctx.Logger().Error("unable to delete marker", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
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
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	if err := k.Keeper.MintCoin(ctx, msg.GetSigners()[0], msg.Amount); err != nil {
		ctx.Logger().Error("unable to mint coin for marker", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
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
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	if err := k.Keeper.BurnCoin(ctx, msg.GetSigners()[0], msg.Amount); err != nil {
		ctx.Logger().Error("unable to burn coin from marker", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
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
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	to, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, err
	}

	if err := k.Keeper.WithdrawCoins(ctx, msg.GetSigners()[0], to, msg.Denom, msg.Amount); err != nil {
		ctx.Logger().Error("unable to withdraw coins from marker", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
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
//
//	sent using the normal bank send process)
func (k msgServer) Transfer(goCtx context.Context, msg *types.MsgTransferRequest) (*types.MsgTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate transaction message.
	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
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

// Transfer handles a message to send coins from one account to another (used with restricted coins that are not
//
//	sent using the normal bank send process)
func (k msgServer) IbcTransfer(goCtx context.Context, msg *types.MsgIbcTransferRequest) (*types.MsgIbcTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate transaction message.
	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	from, err := sdk.AccAddressFromBech32(msg.Transfer.Sender)
	if err != nil {
		return nil, err
	}
	admin, err := sdk.AccAddressFromBech32(msg.Administrator)
	if err != nil {
		return nil, err
	}

	err = k.IbcTransferCoin(ctx, msg.Transfer.SourcePort, msg.Transfer.SourceChannel, msg.Transfer.Token, from, admin, msg.Transfer.Receiver, msg.Transfer.TimeoutHeight, msg.Transfer.TimeoutTimestamp, func(ctx sdk.Context, ibcKeeper ibckeeper.Keeper, sender sdk.AccAddress, token sdk.Coin) (canTransfer bool, err error) {
		if !ibcKeeper.GetSendEnabled(ctx) {
			return false, ibctypes.ErrSendDisabled
		}

		if ibcKeeper.BankKeeper.BlockedAddr(sender) {
			return false, sdkerrors.ErrUnauthorized.Wrapf("%s is not allowed to send funds", sender)
		}

		return true, nil
	})
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
			[]string{types.ModuleName, types.EventTelemetryKeyIbcTransfer},
			1,
			[]metrics.Label{
				telemetry.NewLabel(types.EventTelemetryLabelToAddress, msg.Transfer.Receiver),
				telemetry.NewLabel(types.EventTelemetryLabelFromAddress, msg.Transfer.Sender),
				telemetry.NewLabel(types.EventTelemetryLabelDenom, msg.Transfer.Token.Denom),
				telemetry.NewLabel(types.EventTelemetryLabelAdministrator, msg.Administrator),
			},
		)
		if msg.Transfer.Token.Amount.IsInt64() {
			telemetry.SetGaugeWithLabels(
				[]string{types.ModuleName, types.EventTelemetryKeyIbcTransfer, msg.Transfer.Token.Denom},
				float32(msg.Transfer.Token.Amount.Int64()),
				[]metrics.Label{telemetry.NewLabel(types.EventTelemetryLabelDenom, msg.Transfer.Token.Denom)},
			)
		}
	}()

	return &types.MsgIbcTransferResponse{}, nil
}

// SetDenomMetadata handles a message setting metadata for a marker with the specified denom.
func (k msgServer) SetDenomMetadata(
	goCtx context.Context,
	msg *types.MsgSetDenomMetadataRequest,
) (*types.MsgSetDenomMetadataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate transaction message.
	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	admin, addrErr := sdk.AccAddressFromBech32(msg.Administrator)
	if addrErr != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(addrErr.Error())
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

// ReflectMarker  handles the  reflect marker request to create the marker on another network via ICA
func (k msgServer) ReflectMarker(goCtx context.Context, msg *types.MsgReflectMarkerRequest) (*types.MsgReflectMarkerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	marker, err := k.Keeper.GetMarkerByDenom(ctx, msg.MarkerDenom)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	if marker.HasFixedSupply() {
		return nil, fmt.Errorf("marker cannot have fixed supply")
	}

	if marker.GetStatus() != types.StatusActive {
		return nil, fmt.Errorf("marker must be in Active state : %s", marker.GetStatus())
	}

	filteredAccessList, err := filterAccessList(marker.GetAccessList(), msg.Administrator)

	owner, found := k.intertxKeeper.GetInterChainAccountAddress(ctx, msg.ConnectionId, msg.Administrator)
	if !found {
		return nil, fmt.Errorf("interchain account for address %s not found", msg.Administrator)
	}
	icaReflect := types.NewMsgIcaReflectMarkerRequest(
		msg.MarkerDenom,
		msg.IbcDenom,
		msg.Administrator,
		owner,
		marker.GetStatus(),
		marker.GetMarkerType(),
		filteredAccessList,
		marker.HasGovernanceEnabled(),
	)
	err = icaReflect.ValidateBasic()
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	submitTx, err := intertxtypes.NewMsgSubmitTx(icaReflect, msg.ConnectionId, msg.Administrator)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	err = submitTx.ValidateBasic()
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	err = k.intertxKeeper.SubmitTx(ctx, submitTx, time.Minute)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)
	return nil, nil
}

// filterAccessList creates a new access list with burn, mint removed.  Will fail if administrator is not in list with transfer rights
func filterAccessList(accessList []types.AccessGrant, administrator string) ([]types.AccessGrant, error) {
	// remove mint and burn from grants with transfer permission
	var filteredAccessList []types.AccessGrant
	for _, grant := range accessList {
		var hasTransfer bool
		var acctAccess []types.Access
		for _, access := range grant.Permissions {
			if access != types.Access_Burn && access != types.Access_Mint {
				acctAccess = append(acctAccess, access)
			}
			if access == types.Access_Transfer {
				hasTransfer = true
			}
		}
		if len(acctAccess) > 0 && hasTransfer {
			accessGrant := types.AccessGrant{
				Address:     grant.Address,
				Permissions: acctAccess,
			}
			filteredAccessList = append(filteredAccessList, accessGrant)
		}
	}

	// check if administrator address is in the final list
	var containsAdmin bool
	for _, accessList := range filteredAccessList {
		if accessList.Address == administrator {
			containsAdmin = true
			break
		}
	}

	if !containsAdmin {
		return nil, fmt.Errorf("marker does not have valid access rights")
	}
	return filteredAccessList, nil
}

// TODO Cleanup
func (k msgServer) IcaReflectMarker(goCtx context.Context, msg *types.MsgIcaReflectMarkerRequest) (*types.MsgIcaReflectMarkerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var err error

	if err = msg.ValidateBasic(); err != nil {
		ctx.Logger().Error("unable to pass validate basic", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	if !k.isMatchingDenom(ctx, msg.GetMarkerDenom(), msg.GetIbcDenom()) {
		ctx.Logger().Error("marker denom and ibc denom mismatch")
		return nil, types.ErrReflectDenomMismatch
	}

	marker := k.extractMarker(ctx, msg)

	if k.markerExists(ctx, marker.GetDenom()) {
		err = k.setMarkerPermissions(ctx, marker)
	} else {
		err = k.addMarker(ctx, marker.GetManager(), marker)
	}
	if err != nil {
		ctx.Logger().Error("failed to set permissions or add marker", "err", err)
		return nil, err
	}

	// TODO Check event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

	return &types.MsgIcaReflectMarkerResponse{}, nil
}

func (k msgServer) setMarkerPermissions(ctx sdk.Context, reflectedMarker types.MarkerAccountI) error {
	marker, err := k.GetMarkerByDenom(ctx, reflectedMarker.GetDenom())
	if err != nil {
		ctx.Logger().Error("unable to get marker by denom", "err", err)
		return sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	if err := k.clearAccessList(marker); err != nil {
		ctx.Logger().Error("failed to clear access list of existing marker", "err", err)
		return sdkerrors.ErrUnauthorized.Wrap(err.Error())
	}

	for _, grant := range reflectedMarker.GetAccessList() {
		grant := grant

		if err := marker.GrantAccess(&grant); err != nil {
			ctx.Logger().Error("unable to add access grant to marker", "err", err)
			return sdkerrors.ErrUnauthorized.Wrap(err.Error())
		}
	}

	k.SetMarker(ctx, marker)

	// TODO We may want a specific ReplacePermissionsEvent here

	return nil
}

func (k msgServer) addMarker(ctx sdk.Context, signer sdk.AccAddress, marker types.MarkerAccountI) error {
	// Must be in proposed state
	if err := marker.SetStatus(types.StatusProposed); err != nil {
		ctx.Logger().Error("unable to add set status to proposed for marker", "err", err)
		return sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	if err := k.Keeper.AddMarkerAccount(ctx, marker); err != nil {
		ctx.Logger().Error("unable to add marker account", "err", err)
		return sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	if err := k.Keeper.FinalizeMarker(ctx, signer, marker.GetDenom()); err != nil {
		ctx.Logger().Error("unable to finalize marker", "err", err)
		return sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	if err := k.Keeper.ActivateMarker(ctx, signer, marker.GetDenom()); err != nil {
		ctx.Logger().Error("unable to activate marker", "err", err)
		return sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	return nil
}

func (k msgServer) markerExists(ctx sdk.Context, denom string) bool {
	_, err := k.GetMarkerByDenom(ctx, denom)
	return err == nil
}

func (k msgServer) clearAccessList(marker types.MarkerAccountI) error {
	for _, grant := range marker.GetAccessList() {
		if err := marker.RevokeAccess(grant.GetAddress()); err != nil {
			return err
		}
	}

	return nil
}

func (k msgServer) extractMarker(ctx sdk.Context, msg *types.MsgIcaReflectMarkerRequest) types.MarkerAccountI {
	// TODO Check if this works with the ibc denom
	addr := types.MustGetMarkerAddress(msg.GetIbcDenom())
	account := authtypes.NewBaseAccount(addr, nil, 0, 0)

	marker := types.NewMarkerAccount(
		account,
		k.bankKeeper.GetSupply(ctx, msg.GetIbcDenom()),
		sdk.MustAccAddressFromBech32(msg.GetInvoker()),
		msg.GetAccessControl(),
		msg.GetStatus(),
		msg.GetMarkerType(),
	)
	marker.Manager = msg.GetInvoker()
	marker.SupplyFixed = false
	marker.AllowGovernanceControl = msg.GetAllowGovernanceControl()
	return marker
}

func (k msgServer) isMatchingDenom(ctx sdk.Context, markerDenom, ibcDenom string) bool {
	denom, found := k.extractDenom(ctx, ibcDenom)
	return found && denom == markerDenom
}

func (k msgServer) extractDenom(ctx sdk.Context, denom string) (string, bool) {
	hexHash := denom[len(ibctypes.DenomPrefix+"/"):]

	hash, err := ibctypes.ParseHexHash(hexHash)
	if err != nil {
		ctx.Logger().Error("unable to parse hex hash", "err", err)
		return "", false
	}

	denomTrace, found := k.ibcKeeper.GetDenomTrace(ctx, hash)
	if !found {
		ctx.Logger().Error("unable to get denom trace for hash")
		return "", false
	}

	return denomTrace.BaseDenom, true
}
