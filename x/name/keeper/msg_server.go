package keeper

import (
	"context"
	"fmt"

	"github.com/armon/go-metrics"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	exptypes "github.com/provenance-io/provenance/x/expiration/types"
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
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	// Fetch the parent name record from the keeper.
	record, err := s.Keeper.GetRecordByName(ctx, msg.Parent.Name)
	if err != nil {
		ctx.Logger().Error("unable to find parent name record", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	// Ensure that if the parent name is restricted, it resolves to the given parent address (message signer).
	if record.Restricted {
		parentAddress, addrErr := sdk.AccAddressFromBech32(msg.Parent.Address)
		if addrErr != nil {
			ctx.Logger().Error("unable to parse parent address", "err", addrErr)
			return nil, sdkerrors.ErrInvalidRequest.Wrap(addrErr.Error())
		}
		if !s.Keeper.ResolvesTo(ctx, msg.Parent.Name, parentAddress) {
			errm := "parent name is restricted and does not resolve to the provided parent address"
			ctx.Logger().Error(errm)
			return nil, sdkerrors.ErrInvalidRequest.Wrap(errm)
		}
	}
	// Combine names, normalize, and check for existing record
	n := fmt.Sprintf("%s.%s", msg.Record.Name, msg.Parent.Name)
	name, err := s.Keeper.Normalize(ctx, n)
	if err != nil {
		ctx.Logger().Error("invalid name", "name", name)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	if s.Keeper.NameExists(ctx, name) {
		ctx.Logger().Error("name already bound", "name", name)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(types.ErrNameAlreadyBound.Error())
	}
	// Bind name to address
	address, err := sdk.AccAddressFromBech32(msg.Record.Address)
	if err != nil {
		ctx.Logger().Error("invalid address", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	if err := s.Keeper.SetNameRecord(ctx, name, address, msg.Record.Restricted); err != nil {
		ctx.Logger().Error("unable to bind name", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	// create an expiration if an expiration value is specified
	if len(msg.Expiration) > 0 {
		// attempt to create expiration metadata for name
		if err := s.setExpirationForNameRecord(ctx, msg); err != nil {
			return nil, err
		}
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
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	// Normalize
	name, err := s.Keeper.Normalize(ctx, msg.Record.Name)
	if err != nil {
		ctx.Logger().Error("invalid name", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	// Parse address
	address, err := sdk.AccAddressFromBech32(msg.Record.Address)
	if err != nil {
		ctx.Logger().Error("invalid address", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}
	// Ensure the name exists
	if !s.Keeper.NameExists(ctx, name) {
		ctx.Logger().Error("invalid name", "name", name)
		return nil, sdkerrors.ErrInvalidRequest.Wrap("name does not exist")
	}
	// Ensure permission
	if !s.Keeper.ResolvesTo(ctx, name, address) {
		ctx.Logger().Error("msg sender cannot delete name", "name", name)
		return nil, sdkerrors.ErrUnauthorized.Wrap("msg sender cannot delete name")
	}

	// remove expiration record
	if err := s.expKeeper.RemoveExpiration(ctx, address.String(), address); err != nil {
		return nil, err
	}

	// Delete
	if err := s.Keeper.DeleteRecord(ctx, name); err != nil {
		ctx.Logger().Error("error deleting name", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
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

// ModifyName updates an existing name
func (s msgServer) ModifyName(goCtx context.Context, msg *types.MsgModifyNameRequest) (*types.MsgModifyNameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.GetAuthority() != s.Keeper.GetAuthority() {
		return nil, govtypes.ErrInvalidSigner.Wrapf("expected %s got %s", s.Keeper.GetAuthority(), msg.GetAuthority())
	}

	existing, _ := s.Keeper.GetRecordByName(ctx, msg.GetRecord().Name)
	if existing == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(types.ErrNameNotBound.Error())
	}

	addr, err := sdk.AccAddressFromBech32(msg.GetRecord().Address)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	if err := s.Keeper.UpdateNameRecord(ctx, msg.GetRecord().Name, addr, msg.GetRecord().Restricted); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	return &types.MsgModifyNameResponse{}, nil
}

// setExpirationForNameRecord registers an expiration record
// with the expiration module for the supplied types.NameRecord
func (s msgServer) setExpirationForNameRecord(ctx sdk.Context, msg *types.MsgBindNameRequest) error {
	// create expire name request
	expireMsg := types.NewMsgDeleteNameRequest(msg.Record)
	wrapper, anyErr := codectypes.NewAnyWithValue(expireMsg)
	if anyErr != nil {
		return anyErr
	}

	// if an expiration period isn't provided, use default expiration period
	expShort := msg.Expiration
	if len(expShort) == 0 {
		expShort = s.Keeper.GetDefaultNameExpiration(ctx, msg.Record)
	}

	// parse expiration into a duration
	duration, durErr := exptypes.ParseDuration(expShort)
	if durErr != nil {
		return durErr
	}

	// use record address as owner for name as expiration depositor
	owner := msg.Record.Address

	// create expiration metadata that will expire scope when executed.
	expTime := ctx.BlockTime().Add(*duration)
	expDeposit := sdk.NewCoins(s.expKeeper.GetDeposit(ctx))
	expiration := exptypes.NewExpiration(owner, owner, expTime, expDeposit, *wrapper)
	expErr := s.expKeeper.SetExpiration(ctx, *expiration)
	if expErr != nil {
		return expErr
	}

	return nil
}
