package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/provenance-io/provenance/x/marker/types"

	markertypes "github.com/provenance-io/provenance/x/marker/types"
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

	// TODO Implement ValidateBasic
	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	marker := k.extractMarker(ctx, msg)

	if k.markerExists(ctx, marker.GetDenom()) {
		err = k.setMarkerPermissions(ctx, marker)
	} else {
		err = k.addMarker(ctx, marker.GetManager(), marker)
	}
	if err != nil {
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

func (k icaServer) setMarkerPermissions(ctx sdk.Context, reflectedMarker markertypes.MarkerAccountI) error {
	marker, err := k.GetMarkerByDenom(ctx, reflectedMarker.GetDenom())
	if err != nil {
		// TODO Some log
		return err
	}
	if err := k.clearAccessList(ctx, marker); err != nil {
		// TODO Some log
		return err
	}

	for _, grant := range reflectedMarker.GetAccessList() {
		// If these are either of the permissions then throw error
		if grant.HasAccess(markertypes.Access_Mint) || grant.HasAccess(markertypes.Access_Burn) {
			err := markertypes.ErrReflectAccessTypeInvalid
			ctx.Logger().Error("unable to reflect grant from marker", "err", err)
			return err
		}

		if err := marker.GrantAccess(&grant); err != nil {
			ctx.Logger().Error("unable to add access grant to marker", "err", err)
			return sdkerrors.ErrUnauthorized.Wrap(err.Error())
		}
	}

	k.SetMarker(ctx, marker)

	// TODO We may want a specific replace permissions event here

	return nil
}

func (k icaServer) addMarker(ctx sdk.Context, signer sdk.AccAddress, marker markertypes.MarkerAccountI) error {
	// TODO Can be moved to ValidateBasic
	if marker.HasFixedSupply() {
		err := markertypes.ErrReflectSupplyFixed
		ctx.Logger().Error("unable to add a reflected marker with fixed supply", "err", err)
		return err
	}

	// TODO Can be moved to ValidateBasic
	if marker.GetStatus() != markertypes.StatusActive {
		err := markertypes.ErrReflectMarkerStatus
		ctx.Logger().Error("unable to add a reflected marker that is not active", "err", err)
		return err
	}

	// Must be in proposed state
	marker.SetStatus(markertypes.StatusProposed)

	// TODO Check error
	if err := k.Keeper.AddMarkerAccount(ctx, marker); err != nil {
		ctx.Logger().Error("unable to add marker", "err", err)
		return sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	// This may not be needed because we already have the permissions on the marker object when we call AddMarkerAccount
	/*if err := k.setMarkerPermissions(ctx, marker); err != nil {
		return err
	}*/

	// TODO Check error
	if err := k.Keeper.FinalizeMarker(ctx, signer, marker.GetDenom()); err != nil {
		ctx.Logger().Error("unable to finalize marker", "err", err)
		return sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	// TODO Check error
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

func (k icaServer) clearAccessList(ctx sdk.Context, marker markertypes.MarkerAccountI) error {
	for _, grant := range marker.GetAccessList() {
		if err := marker.RevokeAccess(grant.GetAddress()); err != nil {
			return err
		}
	}

	return nil
}

func (k icaServer) extractMarker(ctx sdk.Context, msg *types.MsgIcaReflectMarkerRequest) markertypes.MarkerAccountI {
	manager := msg.GetSigners()[0]

	// TODO Check if this works with the ibc denom
	addr := types.MustGetMarkerAddress(msg.GetIbcDenom())
	account := authtypes.NewBaseAccount(addr, nil, 0, 0)

	marker := markertypes.NewMarkerAccount(
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
