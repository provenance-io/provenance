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

	// TODO Implement ValidateBasic
	if err := msg.ValidateBasic(); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	// TODO We need the correct error
	reflectedMarker, ok := msg.GetMarker().GetCachedValue().(markertypes.MarkerAccountI)
	if !ok {
		return nil, sdkerrors.ErrInvalidType
	}

	// TODO Check for no mint and burn

	ibcDenom := msg.GetIbcDenom()
	signer := msg.GetSigners()[0]

	var err error
	if k.markerExists(ctx, ibcDenom) {
		err = k.setMarkerPermissions(ctx, ibcDenom, reflectedMarker)
	} else {
		err = k.addMarker(ctx, signer, ibcDenom, reflectedMarker)
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

func (k icaServer) setMarkerPermissions(ctx sdk.Context, denom string, reflectedMarker markertypes.MarkerAccountI) error {
	marker, err := k.GetMarkerByDenom(ctx, denom)
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

func (k icaServer) addMarker(ctx sdk.Context, signer sdk.AccAddress, denom string, reflectedMarker markertypes.MarkerAccountI) error {
	addr := types.MustGetMarkerAddress(denom)
	account := authtypes.NewBaseAccount(addr, nil, 0, 0)
	marker := reflectedMarker.Clone()
	marker.AccessControl = []markertypes.AccessGrant{}
	marker.Denom = denom
	marker.BaseAccount = account

	if marker.SupplyFixed {
		err := markertypes.ErrReflectSupplyFixed
		ctx.Logger().Error("unable to add a reflected marker with fixed supply", "err", err)
		return err
	}

	// TODO Check error
	if err := k.Keeper.AddMarkerAccount(ctx, marker); err != nil {
		ctx.Logger().Error("unable to add marker", "err", err)
		return sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	if err := k.setMarkerPermissions(ctx, denom, reflectedMarker); err != nil {
		return err
	}

	// TODO Check error
	if err := k.Keeper.FinalizeMarker(ctx, signer, marker.Denom); err != nil {
		ctx.Logger().Error("unable to finalize marker", "err", err)
		return sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	// TODO Check error
	if err := k.Keeper.ActivateMarker(ctx, signer, marker.Denom); err != nil {
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
