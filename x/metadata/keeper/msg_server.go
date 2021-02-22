package keeper

import (
	"context"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/metadata/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the distribution MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) MemorializeContract(
	goCtx context.Context,
	msg *types.MsgMemorializeContractRequest,
) (*types.MemorializeContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO (contract keeper class  methods to process contract execution, scope keeper methods to record it)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Notary),
		),
	)

	return nil, fmt.Errorf("not implemented")
}

func (k msgServer) ChangeOwnership(
	goCtx context.Context,
	msg *types.MsgChangeOwnershipRequest,
) (*types.ChangeOwnershipResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO (contract keeper class  methods to process contract execution, scope keeper methods to record it)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Notary),
		),
	)

	return nil, fmt.Errorf("not implemented")
}

func (k msgServer) AddScopeSpecification(
	goCtx context.Context,
	msg *types.MsgAddScopeSpecificationRequest,
) (*types.AddScopeSpecificationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	existing, _ := k.GetScopeSpecification(ctx, msg.Specification.SpecificationId)
	if err := k.ValidateScopeSpecUpdate(ctx, existing, *msg.Specification, msg.Signers); err != nil {
		return nil, err
	}

	k.SetScopeSpecification(ctx, *msg.Specification)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, strings.Join(msg.Signers, ",")),
		),
	)

	return &types.AddScopeSpecificationResponse{}, nil
}

func (k msgServer) RemoveScopeSpecification(
	goCtx context.Context,
	msg *types.MsgRemoveScopeSpecificationRequest,
) (*types.RemoveScopeSpecificationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	existing, found := k.GetScopeSpecification(ctx, msg.SpecificationId)
	if !found {
		return nil, fmt.Errorf("scope specification not found with id %s", msg.SpecificationId)
	}
	if err := k.ValidateScopeSpecAllOwnersAreSigners(existing, msg.Signers); err != nil {
		return nil, err
	}

	k.DeleteScopeSpecification(ctx, msg.SpecificationId)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, strings.Join(msg.Signers, ",")),
		),
	)

	return &types.RemoveScopeSpecificationResponse{}, nil
}

func (k msgServer) AddGroupSpecification(
	goCtx context.Context,
	msg *types.MsgAddGroupSpecificationRequest,
) (*types.AddGroupSpecificationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO (contract keeper class  methods to process request, keeper methods to record it)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, ""),
		),
	)

	return nil, fmt.Errorf("not implemented")
}

func (k msgServer) AddRecord(
	goCtx context.Context,
	msg *types.MsgAddRecordRequest,
) (*types.AddRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO (contract keeper class  methods to process request, keeper methods to record it)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, ""),
		),
	)

	return nil, fmt.Errorf("not implemented")
}

func (k msgServer) AddRecordGroup(
	goCtx context.Context,
	msg *types.MsgAddRecordGroupRequest,
) (*types.AddRecordGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO (contract keeper class  methods to process request, keeper methods to record it)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, ""),
		),
	)

	return nil, fmt.Errorf("not implemented")
}

func (k msgServer) AddScope(
	goCtx context.Context,
	msg *types.MsgAddScopeRequest,
) (*types.AddScopeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	existing, _ := k.GetScope(ctx, msg.Scope.ScopeId)
	if err := k.ValidateScopeUpdate(ctx, existing, msg.Scope, msg.Signers); err != nil {
		return nil, err
	}

	k.SetScope(ctx, msg.Scope)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, ""),
		),
	)

	return nil, fmt.Errorf("not implemented")
}

func (k msgServer) RemoveScope(
	goCtx context.Context,
	msg *types.MsgRemoveScopeRequest,
) (*types.RemoveScopeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	existing, _ := k.GetScope(ctx, msg.ScopeId)
	// validate that all fields can be unset with the given list of signers
	if err := k.ValidateScopeUpdate(ctx, existing, types.Scope{ScopeId: msg.ScopeId}, msg.Signers); err != nil {
		return nil, err
	}

	k.DeleteScope(ctx, msg.ScopeId)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, ""),
		),
	)

	return nil, fmt.Errorf("not implemented")
}
