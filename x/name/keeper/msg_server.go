package keeper

import (
	"context"
	"fmt"

	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/provenance-io/provenance/x/name/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the account MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// BindName binds a name to an address
func (s msgServer) BindName(goCtx context.Context, msg *types.MsgBindNameRequest) (*types.MsgBindNameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Validate
	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Error("unable to validate message", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	// Fetch the parent name record from the keeper.
	record, err := s.Keeper.GetRecordByName(ctx, msg.Parent.Name)
	if err != nil {
		ctx.Logger().Error("unable to find parent name record", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	// Ensure that if the parent name is restricted, it resolves to the given parent address (message signer).
	if record.Restricted {
		parentAddress, addrErr := sdk.AccAddressFromBech32(msg.Parent.Address)
		if addrErr != nil {
			ctx.Logger().Error("unable to parse parent address", "err", addrErr)
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, addrErr.Error())
		}
		if !s.Keeper.ResolvesTo(ctx, msg.Parent.Name, parentAddress) {
			errm := "parent name is restricted and does not resolve to the provided parent address"
			ctx.Logger().Error(errm)
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, errm)
		}
	}
	// Combine names, normalize, and check for existing record
	n := fmt.Sprintf("%s.%s", msg.Record.Name, msg.Parent.Name)
	name, err := s.Keeper.Normalize(ctx, n)
	if err != nil {
		ctx.Logger().Error("invalid name", "name", name)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	if s.Keeper.NameExists(ctx, name) {
		ctx.Logger().Error("name already bound", "name", name)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, types.ErrNameAlreadyBound.Error())
	}
	// Bind name to address
	address, err := sdk.AccAddressFromBech32(msg.Record.Address)
	if err != nil {
		ctx.Logger().Error("invalid address", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	if err := s.Keeper.SetNameRecord(ctx, name, address, msg.Record.Restricted); err != nil {
		ctx.Logger().Error("unable to bind name", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	// key: modulename+name+bind
	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, "name", "bind"},
			1,
			[]metrics.Label{telemetry.NewLabel("name", name), telemetry.NewLabel("address", msg.Record.Address)},
		)
	}()

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeNameBound,
			sdk.NewAttribute(types.KeyAttributeAddress, msg.Record.Address),
			sdk.NewAttribute(types.KeyAttributeName, msg.Record.Name),
		),
	)

	return &types.MsgBindNameResponse{}, nil
}

// DeleteName unbinds a name from an address
func (s msgServer) DeleteName(goCtx context.Context, msg *types.MsgDeleteNameRequest) (*types.MsgDeleteNameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Validate
	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Error("unable to validate message", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	// Normalize
	name, err := s.Keeper.Normalize(ctx, msg.Record.Name)
	if err != nil {
		ctx.Logger().Error("invalid name", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	// Parse address
	address, err := sdk.AccAddressFromBech32(msg.Record.Address)
	if err != nil {
		ctx.Logger().Error("invalid address", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	// Ensure the name exists
	if !s.Keeper.NameExists(ctx, name) {
		ctx.Logger().Error("invalid name", "name", name)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "name does not exist")
	}
	// Ensure permission
	if !s.Keeper.ResolvesTo(ctx, name, address) {
		ctx.Logger().Error("msg sender cannot delete name", "name", name)
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "msg sender cannot delete name")
	}
	// Delete
	if err := s.Keeper.DeleteRecord(ctx, name); err != nil {
		ctx.Logger().Error("error deleting name", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	// key: modulename+name+unbind
	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, "name", "unbind"},
			1,
			[]metrics.Label{telemetry.NewLabel("name", name), telemetry.NewLabel("address", msg.Record.Address)},
		)
	}()

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeNameUnbound,
			sdk.NewAttribute(types.KeyAttributeAddress, msg.Record.Address),
			sdk.NewAttribute(types.KeyAttributeName, msg.Record.Name),
		),
	)

	return &types.MsgDeleteNameResponse{}, nil
}
