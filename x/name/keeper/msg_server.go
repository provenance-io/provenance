package keeper

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-metrics"

	"cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/provenance-io/provenance/x/name/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the name MsgServer interface
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
			return nil, sdkerrors.ErrInvalidRequest.Wrapf("parent name %q is restricted and does not resolve to the provided parent address", record.Name)
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
	// Delete
	err = s.Keeper.DeleteRecord(ctx, name)
	if err != nil {
		ctx.Logger().Error("error deleting name", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	// Remove all attributes from assigned accounts
	err = s.Keeper.attrKeeper.PurgeAttribute(ctx, name, address)
	if err != nil {
		return nil, err
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

	existing, _ := s.Keeper.GetRecordByName(ctx, msg.GetRecord().Name)
	if existing == nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(types.ErrNameNotBound.Error())
	}

	if msg.GetAuthority() != s.Keeper.GetAuthority() && msg.GetAuthority() != existing.Address {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("expected %s or %s got %s", s.Keeper.GetAuthority(), existing.Address, msg.GetAuthority())
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

// CreateRootName binds a name to an address
func (s msgServer) CreateRootName(goCtx context.Context, msg *types.MsgCreateRootNameRequest) (*types.MsgCreateRootNameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if s.Keeper.GetAuthority() != msg.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "expected %s got %s", s.Keeper.GetAuthority(), msg.Authority)
	}

	// Routes to legacy proposal handler to avoid code duplication
	// Setting title and description to empty strings. These two fields are deprecated in the v1.
	err := HandleCreateRootNameProposal(ctx, s.Keeper, types.NewCreateRootNameProposal("", "", msg.Record.Name, sdk.AccAddress(msg.Record.Address), msg.Record.Restricted))
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateRootNameResponse{}, nil
}
