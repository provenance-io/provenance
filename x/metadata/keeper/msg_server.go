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
) (*types.MsgMemorializeContractResponse, error) {
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
) (*types.MsgChangeOwnershipResponse, error) {
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

func (k msgServer) AddScope(
	goCtx context.Context,
	msg *types.MsgAddScopeRequest,
) (*types.MsgAddScopeResponse, error) {
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
			sdk.NewAttribute(sdk.AttributeKeySender, strings.Join(msg.Signers, ",")),
		),
	)

	return &types.MsgAddScopeResponse{}, nil
}

func (k msgServer) DeleteScope(
	goCtx context.Context,
	msg *types.MsgDeleteScopeRequest,
) (*types.MsgDeleteScopeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	existing, _ := k.GetScope(ctx, msg.ScopeId)
	// validate that all fields can be unset with the given list of signers
	if err := k.ValidateScopeRemove(ctx, existing, types.Scope{ScopeId: msg.ScopeId}, msg.Signers); err != nil {
		return nil, err
	}

	k.RemoveScope(ctx, msg.ScopeId)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(types.AttributeKeyScopeID, string(msg.ScopeId)),
		),
	)

	return &types.MsgDeleteScopeResponse{}, nil
}

func (k msgServer) AddSession(
	goCtx context.Context,
	msg *types.MsgAddSessionRequest,
) (*types.MsgAddSessionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	existing, _ := k.GetSession(ctx, msg.Session.SessionId)
	if err := k.ValidateSessionUpdate(ctx, existing, *msg.Session, msg.Signers); err != nil {
		return nil, err
	}

	k.SetSession(ctx, *msg.Session)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, strings.Join(msg.Signers, ",")),
		),
	)

	return &types.MsgAddSessionResponse{}, nil
}

func (k msgServer) AddRecord(
	goCtx context.Context,
	msg *types.MsgAddRecordRequest,
) (*types.MsgAddRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	scopeUUID, err := msg.Record.SessionId.ScopeUUID()
	if err != nil {
		return nil, err
	}

	recordID := types.RecordMetadataAddress(scopeUUID, msg.Record.Name)

	existing, _ := k.GetRecord(ctx, recordID)
	if err := k.ValidateRecordUpdate(ctx, existing, *msg.Record, msg.Signers); err != nil {
		return nil, err
	}

	k.SetRecord(ctx, *msg.Record)

	if !existing.SessionId.Equals(msg.Record.SessionId) {
		k.RemoveSession(ctx, existing.SessionId)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, strings.Join(msg.Signers, ",")),
		),
	)

	return &types.MsgAddRecordResponse{}, nil
}

func (k msgServer) DeleteRecord(
	goCtx context.Context,
	msg *types.MsgDeleteRecordRequest,
) (*types.MsgDeleteRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	existing, _ := k.GetRecord(ctx, msg.RecordId)
	if err := k.ValidateRecordRemove(ctx, existing, msg.RecordId, msg.Signers); err != nil {
		return nil, err
	}

	k.RemoveRecord(ctx, msg.RecordId)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, strings.Join(msg.Signers, ",")),
		),
	)

	return &types.MsgDeleteRecordResponse{}, nil
}

func (k msgServer) AddScopeSpecification(
	goCtx context.Context,
	msg *types.MsgAddScopeSpecificationRequest,
) (*types.MsgAddScopeSpecificationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var existing *types.ScopeSpecification = nil
	if e, found := k.GetScopeSpecification(ctx, msg.Specification.SpecificationId); found {
		existing = &e
	}
	if err := k.ValidateScopeSpecUpdate(ctx, existing, msg.Specification, msg.Signers); err != nil {
		return nil, err
	}

	k.SetScopeSpecification(ctx, msg.Specification)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, strings.Join(msg.Signers, ",")),
		),
	)

	return &types.MsgAddScopeSpecificationResponse{}, nil
}

func (k msgServer) DeleteScopeSpecification(
	goCtx context.Context,
	msg *types.MsgDeleteScopeSpecificationRequest,
) (*types.MsgDeleteScopeSpecificationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	existing, found := k.GetScopeSpecification(ctx, msg.SpecificationId)
	if !found {
		return nil, fmt.Errorf("scope specification not found with id %s", msg.SpecificationId)
	}
	if err := k.ValidateAllOwnersAreSigners(existing.OwnerAddresses, msg.Signers); err != nil {
		return nil, err
	}

	k.RemoveScopeSpecification(ctx, msg.SpecificationId)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, strings.Join(msg.Signers, ",")),
		),
	)

	return &types.MsgDeleteScopeSpecificationResponse{}, nil
}

func (k msgServer) AddContractSpecification(
	goCtx context.Context,
	msg *types.MsgAddContractSpecificationRequest,
) (*types.MsgAddContractSpecificationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var existing *types.ContractSpecification = nil
	if e, found := k.GetContractSpecification(ctx, msg.Specification.SpecificationId); found {
		existing = &e
	}
	if err := k.ValidateContractSpecUpdate(ctx, existing, msg.Specification, msg.Signers); err != nil {
		return nil, err
	}

	k.SetContractSpecification(ctx, msg.Specification)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, strings.Join(msg.Signers, ",")),
		),
	)

	return &types.MsgAddContractSpecificationResponse{}, nil
}

func (k msgServer) DeleteContractSpecification(
	goCtx context.Context,
	msg *types.MsgDeleteContractSpecificationRequest,
) (*types.MsgDeleteContractSpecificationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	existing, found := k.GetContractSpecification(ctx, msg.SpecificationId)
	if !found {
		return nil, fmt.Errorf("contract specification not found with id %s", msg.SpecificationId)
	}
	if err := k.ValidateAllOwnersAreSigners(existing.OwnerAddresses, msg.Signers); err != nil {
		return nil, err
	}

	k.RemoveContractSpecification(ctx, msg.SpecificationId)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, strings.Join(msg.Signers, ",")),
		),
	)

	return &types.MsgDeleteContractSpecificationResponse{}, nil
}
