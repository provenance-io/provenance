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

	if k.markerExists(ctx, ibcDenom) {
		k.updateMarkerPermissions(ctx, ibcDenom, reflectedMarker)
	} else {
		k.addMarker(ctx, ibcDenom, reflectedMarker)
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

}

func (k icaServer) addMarker(ctx sdk.Context, denom string, marker markertypes.MarkerAccountI) {
	addr := types.MustGetMarkerAddress(denom)
	manager := marker.GetManager()
	newAccount := authtypes.NewBaseAccount(addr, nil, 0, 0)
	newMarker := marker.Clone()
	newMarker.Denom = denom
	newMarker.BaseAccount = newAccount

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

func (k icaServer) markerExists(ctx sdk.Context, denom string) bool {
	_, err := k.GetMarkerByDenom(ctx, denom)
	return err == nil
}
