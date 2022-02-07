package keeper

import (
	"context"
	"fmt"

	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/attribute/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the attribute MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) AddAttribute(goCtx context.Context, msg *types.MsgAddAttributeRequest) (*types.MsgAddAttributeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	attrib := types.Attribute{
		Address:       msg.Account,
		Name:          msg.Name,
		AttributeType: msg.AttributeType,
		Value:         msg.Value,
	}

	ownerAddr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	err = k.Keeper.SetAttribute(ctx, attrib, ownerAddr)
	if err != nil {
		return nil, err
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, types.EventTelemetryKeyAdd},
			1,
			[]metrics.Label{
				telemetry.NewLabel(types.EventTelemetryLabelName, msg.Name),
				telemetry.NewLabel(types.EventTelemetryLabelType, msg.AttributeType.String()),
				telemetry.NewLabel(types.EventTelemetryLabelAccount, msg.Account),
				telemetry.NewLabel(types.EventTelemetryLabelOwner, msg.Owner),
			},
		)
	}()

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAttributeAdded,
			sdk.NewAttribute(types.AttributeKeyNameAttribute, msg.Name),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, msg.Account),
		),
	)

	return &types.MsgAddAttributeResponse{}, nil
}

func (k msgServer) UpdateAttribute(goCtx context.Context, msg *types.MsgUpdateAttributeRequest) (*types.MsgUpdateAttributeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	originalAttribute := types.Attribute{
		Address:       msg.Account,
		Name:          msg.Name,
		AttributeType: msg.OriginalAttributeType,
		Value:         msg.OriginalValue,
	}

	updateAttribute := types.Attribute{
		Address:       msg.Account,
		Name:          msg.Name,
		AttributeType: msg.UpdateAttributeType,
		Value:         msg.UpdateValue,
	}

	ownerAddr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	err = k.Keeper.UpdateAttribute(ctx, originalAttribute, updateAttribute, ownerAddr)
	if err != nil {
		return nil, err
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, types.EventTelemetryKeyUpdate},
			1,
			[]metrics.Label{
				telemetry.NewLabel(types.EventTelemetryLabelName, msg.Name),
				telemetry.NewLabel(types.EventTelemetryLabelValue, string(msg.OriginalValue)),
				telemetry.NewLabel(types.EventTelemetryLabelType, msg.OriginalAttributeType.String()),
				telemetry.NewLabel(types.EventTelemetryLabelValue, string(msg.UpdateValue)),
				telemetry.NewLabel(types.EventTelemetryLabelType, msg.UpdateAttributeType.String()),
				telemetry.NewLabel(types.EventTelemetryLabelAccount, msg.Account),
				telemetry.NewLabel(types.EventTelemetryLabelOwner, msg.Owner),
			},
		)
	}()

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAttributeUpdated,
			sdk.NewAttribute(types.AttributeKeyNameAttribute, msg.Name),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, msg.Account),
		),
	)

	return &types.MsgUpdateAttributeResponse{}, nil
}

func (k msgServer) DeleteAttribute(goCtx context.Context, msg *types.MsgDeleteAttributeRequest) (*types.MsgDeleteAttributeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := types.ValidateAttributeAddress(msg.Account)
	if err != nil {
		return nil, fmt.Errorf("invalid account address: %w", err)
	}

	ownerAddr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	err = k.Keeper.DeleteAttribute(ctx, msg.Account, msg.Name, nil, ownerAddr)
	if err != nil {
		return nil, err
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, types.EventTelemetryKeyDelete},
			1,
			[]metrics.Label{
				telemetry.NewLabel(types.EventTelemetryLabelName, msg.Name),
				telemetry.NewLabel(types.EventTelemetryLabelAccount, msg.Account),
				telemetry.NewLabel(types.EventTelemetryLabelOwner, msg.Owner),
			},
		)
	}()

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAttributeDeleted,
			sdk.NewAttribute(types.AttributeKeyNameAttribute, msg.Name),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, msg.Account),
		),
	)

	return &types.MsgDeleteAttributeResponse{}, nil
}

func (k msgServer) DeleteDistinctAttribute(goCtx context.Context, msg *types.MsgDeleteDistinctAttributeRequest) (*types.MsgDeleteDistinctAttributeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := types.ValidateAttributeAddress(msg.Account)
	if err != nil {
		return nil, fmt.Errorf("invalid account address: %w", err)
	}

	ownerAddr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	err = k.Keeper.DeleteAttribute(ctx, msg.Account, msg.Name, &msg.Value, ownerAddr)
	if err != nil {
		return nil, err
	}

	defer func() {
		telemetry.IncrCounterWithLabels(
			[]string{types.ModuleName, types.EventTelemetryKeyDistinctDelete},
			1,
			[]metrics.Label{
				telemetry.NewLabel(types.EventTelemetryLabelName, msg.Name),
				telemetry.NewLabel(types.EventTelemetryLabelValue, string(msg.Value)),
				telemetry.NewLabel(types.EventTelemetryLabelAccount, msg.Account),
				telemetry.NewLabel(types.EventTelemetryLabelOwner, msg.Owner),
			},
		)
	}()

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAttributeDistinctDeleted,
			sdk.NewAttribute(types.AttributeKeyNameAttribute, msg.Name),
			sdk.NewAttribute(types.AttributeKeyAccountAddress, msg.Account),
		),
	)

	return &types.MsgDeleteDistinctAttributeResponse{}, nil
}
