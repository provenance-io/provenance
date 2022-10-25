package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

	// What do we do with ibc denom?
	ibcDenom := msg.GetIbcDenom()

	// TODO We need the correct error
	reflectedMarker, ok := msg.GetMarker().GetCachedValue().(markertypes.MarkerAccountI)
	if !ok {
		return nil, sdkerrors.ErrInvalidType
	}

	// Check if the marker exists already
	marker, err := k.GetMarkerByDenom(ctx, marker.GetDenom())
	// If the marker exists then we want to just update some values on it and set it
	// What if the marker has a different owner

	// If the marker doesn't exist then we want to add it
	// Do we also want to Finalize and Activate it?

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)
	return nil, nil
}
