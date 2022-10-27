package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/x/marker/types"
)

type icaServer struct {
	Keeper
}

func NewIcaServerImpl(keeper Keeper) types.IcaServer {
	return icaServer{Keeper: keeper}
}

var _ types.IcaServer = icaServer{}

func (k icaServer) ReflectMarker(goCtx context.Context, msg *types.MsgIcaReflectMarkerRequest) (*types.MsgIcaReflectMarkerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var err error

	if err = msg.ValidateBasic(); err != nil {
		ctx.Logger().Error("unable to pass validate basic", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	marker := k.extractMarker(msg)

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

func (k icaServer) setMarkerPermissions(ctx sdk.Context, reflectedMarker types.MarkerAccountI) error {
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
		// If these are either of the permissions then throw error
		/*if grant.HasAccess(types.Access_Mint) || grant.HasAccess(types.Access_Burn) {
			err := types.ErrReflectAccessTypeInvalid
			ctx.Logger().Error("unable to reflect grant from marker", "err", err)
			return err
		}*/

		if err := marker.GrantAccess(&grant); err != nil {
			ctx.Logger().Error("unable to add access grant to marker", "err", err)
			return sdkerrors.ErrUnauthorized.Wrap(err.Error())
		}
	}

	k.SetMarker(ctx, marker)

	// TODO We may want a specific ReplacePermissionsEvent here

	return nil
}

func (k icaServer) addMarker(ctx sdk.Context, signer sdk.AccAddress, marker types.MarkerAccountI) error {
	// Must be in proposed state
	if err := marker.SetStatus(types.StatusProposed); err != nil {
		ctx.Logger().Error("unable to add set status to proposed for marker", "err", err)
		return sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	if err := k.Keeper.AddMarkerAccount(ctx, marker); err != nil {
		ctx.Logger().Error("unable to add marker account", "err", err)
		return sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	// This may not be needed because we already have the permissions on the marker object when we call AddMarkerAccount
	/*if err := k.setMarkerPermissions(ctx, marker); err != nil {
		return err
	}*/

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

func (k icaServer) markerExists(ctx sdk.Context, denom string) bool {
	_, err := k.GetMarkerByDenom(ctx, denom)
	return err == nil
}

func (k icaServer) clearAccessList(marker types.MarkerAccountI) error {
	for _, grant := range marker.GetAccessList() {
		if err := marker.RevokeAccess(grant.GetAddress()); err != nil {
			return err
		}
	}

	return nil
}

func (k icaServer) extractMarker(msg *types.MsgIcaReflectMarkerRequest) types.MarkerAccountI {
	manager := msg.GetSigners()[0]

	// TODO Check if this works with the ibc denom
	addr := types.MustGetMarkerAddress(msg.GetIbcDenom())
	account := authtypes.NewBaseAccount(addr, nil, 0, 0)

	marker := types.NewMarkerAccount(
		account,
		sdk.NewInt64Coin(msg.GetIbcDenom(), 0),
		manager,
		msg.GetAccessControl(),
		msg.GetStatus(),
		msg.GetMarkerType(),
	)
	marker.SupplyFixed = false
	marker.AllowGovernanceControl = msg.GetAllowGovernanceControl()
	return marker
}
