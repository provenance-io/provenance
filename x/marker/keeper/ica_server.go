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

	// What do we do with ibc denom?
	ibcDenom := msg.GetIbcDenom()
	signer := msg.GetSigners()[0]

	if k.markerExists(ctx, ibcDenom) {
		k.updateMarkerPermissions(ctx, ibcDenom, reflectedMarker)
	} else {
		k.addMarker(ctx, signer, ibcDenom, reflectedMarker)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)
	return nil, nil
}

func (k icaServer) updateMarkerPermissions(ctx context.Context, denom string, marker markertypes.MarkerAccountI) {
	// Find the permissions that are different
	// Add these permissions
	// TODO Implement

}

func (k icaServer) addMarker(ctx sdk.Context, signer sdk.AccAddress, denom string, reflectedMarker markertypes.MarkerAccountI) error {
	addr := types.MustGetMarkerAddress(denom)
	account := authtypes.NewBaseAccount(addr, nil, 0, 0)
	marker := reflectedMarker.Clone()
	marker.AccessControl = []markertypes.AccessGrant{}
	marker.Denom = denom
	marker.BaseAccount = account

	// TODO Check error
	if err := k.Keeper.AddMarkerAccount(ctx, marker); err != nil {
		ctx.Logger().Error("unable to add marker", "err", err)
		return sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	// Add permissions
	for _, grant := range reflectedMarker.GetAccessList() {
		// TODO Check error
		if err := k.Keeper.AddAccess(ctx, signer, denom, &grant); err != nil {
			ctx.Logger().Error("unable to add access grant to marker", "err", err)
			return sdkerrors.ErrUnauthorized.Wrap(err.Error())
		}
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

	// TODO Check events
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

	return nil
}

func (k icaServer) markerExists(ctx sdk.Context, denom string) bool {
	_, err := k.GetMarkerByDenom(ctx, denom)
	return err == nil
}
