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

	var existing *types.Session = nil
	var existingAudit *types.AuditFields = nil
	if e, found := k.GetSession(ctx, msg.Session.SessionId); found {
		existing = &e
		existingAudit = existing.Audit
	}
	if err := k.ValidateSessionUpdate(ctx, existing, *msg.Session, msg.Signers); err != nil {
		return nil, err
	}

	msg.Session.Audit = existingAudit.UpdateAudit(ctx.BlockTime(), strings.Join(msg.Signers, ", "), "")

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

	var existing *types.Record = nil
	if e, found := k.GetRecord(ctx, recordID); found {
		existing = &e
	}
	if err := k.ValidateRecordUpdate(ctx, existing, *msg.Record, msg.Signers); err != nil {
		return nil, err
	}

	k.SetRecord(ctx, *msg.Record)

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
		if err := k.ValidateAllOwnersAreSigners(existing.OwnerAddresses, msg.Signers); err != nil {
			return nil, err
		}
	}
	if err := k.ValidateScopeSpecUpdate(ctx, existing, msg.Specification); err != nil {
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

	if err := k.RemoveScopeSpecification(ctx, msg.SpecificationId); err != nil {
		return nil, fmt.Errorf("cannot delete scope specification with id %s: %w", msg.SpecificationId, err)
	}

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
		if err := k.ValidateAllOwnersAreSigners(existing.OwnerAddresses, msg.Signers); err != nil {
			return nil, err
		}
	}
	if err := k.ValidateContractSpecUpdate(ctx, existing, msg.Specification); err != nil {
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

	// Remove all record specifications associated with this contract specification.
	recSpecs, recSpecErr := k.GetRecordSpecificationsForContractSpecificationID(ctx, msg.SpecificationId)
	if recSpecErr != nil {
		return nil, fmt.Errorf("could not get record specifications to delete with contract specification with id %s: %w",
			msg.SpecificationId, recSpecErr)
	}
	var delRecSpecErr error = nil
	removedRecSpecs := []*types.RecordSpecification{}
	for _, recSpec := range recSpecs {
		if err := k.RemoveRecordSpecification(ctx, recSpec.SpecificationId); err != nil {
			delRecSpecErr = fmt.Errorf("failed to delete record specification %s (name: %s) while trying to delete contract specification %d: %w",
				recSpec.SpecificationId, recSpec.Name, msg.SpecificationId, err)
			break
		}
		removedRecSpecs = append(removedRecSpecs, recSpec)
	}
	if delRecSpecErr != nil {
		// Put the deleted record specifications back since not all of them could be deleted (and neither can this contract spec)
		for _, recSpec := range removedRecSpecs {
			k.SetRecordSpecification(ctx, *recSpec)
		}
		return nil, delRecSpecErr
	}

	// Remove the contract specification itself
	if err := k.RemoveContractSpecification(ctx, msg.SpecificationId); err != nil {
		return nil, fmt.Errorf("cannot delete contract specification with id %s: %w", msg.SpecificationId, err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, strings.Join(msg.Signers, ",")),
		),
	)

	return &types.MsgDeleteContractSpecificationResponse{}, nil
}

func (k msgServer) AddRecordSpecification(
	goCtx context.Context,
	msg *types.MsgAddRecordSpecificationRequest,
) (*types.MsgAddRecordSpecificationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	contractSpecID, err := msg.Specification.SpecificationId.AsContractSpecAddress()
	if err != nil {
		return nil, err
	}
	contractSpec, contractSpecFound := k.GetContractSpecification(ctx, contractSpecID)
	if !contractSpecFound {
		contractSpecUUID, _ := contractSpecID.ContractSpecUUID()
		return nil, fmt.Errorf("contract specification not found with id %s (uuid %s) required for adding or updating record specification with id %s",
			contractSpecID, contractSpecUUID, msg.Specification.SpecificationId)
	}
	if err := k.ValidateAllOwnersAreSigners(contractSpec.OwnerAddresses, msg.Signers); err != nil {
		return nil, err
	}

	var existing *types.RecordSpecification = nil
	if e, found := k.GetRecordSpecification(ctx, msg.Specification.SpecificationId); found {
		existing = &e
	}
	if err := k.ValidateRecordSpecUpdate(ctx, existing, msg.Specification); err != nil {
		return nil, err
	}

	k.SetRecordSpecification(ctx, msg.Specification)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, strings.Join(msg.Signers, ",")),
		),
	)

	return &types.MsgAddRecordSpecificationResponse{}, nil
}

func (k msgServer) DeleteRecordSpecification(
	goCtx context.Context,
	msg *types.MsgDeleteRecordSpecificationRequest,
) (*types.MsgDeleteRecordSpecificationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, found := k.GetRecordSpecification(ctx, msg.SpecificationId)
	if !found {
		return nil, fmt.Errorf("record specification not found with id %s", msg.SpecificationId)
	}
	contractSpecID, err := msg.SpecificationId.AsContractSpecAddress()
	if err != nil {
		return nil, err
	}
	contractSpec, contractSpecFound := k.GetContractSpecification(ctx, contractSpecID)
	if !contractSpecFound {
		return nil, fmt.Errorf("contract specification not found with id %s required for deleting record specification with id %s",
			contractSpecID, msg.SpecificationId)
	}
	if err := k.ValidateAllOwnersAreSigners(contractSpec.OwnerAddresses, msg.Signers); err != nil {
		return nil, err
	}

	if err := k.RemoveRecordSpecification(ctx, msg.SpecificationId); err != nil {
		return nil, fmt.Errorf("cannot delete record specification with id %s: %w", msg.SpecificationId, err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, strings.Join(msg.Signers, ",")),
		),
	)

	return &types.MsgDeleteRecordSpecificationResponse{}, nil
}

func (k msgServer) AddP8EContractSpec(
	goCtx context.Context,
	msg *types.MsgAddP8EContractSpecRequest,
) (*types.MsgAddP8EContractSpecResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	proposed, newrecords, err := types.ConvertP8eContractSpec(&msg.Contractspec, msg.Signers)
	if err != nil {
		return nil, err
	}

	var existing *types.ContractSpecification = nil
	if e, found := k.GetContractSpecification(ctx, proposed.SpecificationId); found {
		existing = &e
		if err := k.ValidateAllOwnersAreSigners(existing.OwnerAddresses, msg.Signers); err != nil {
			return nil, err
		}
	}

	if err := k.ValidateContractSpecUpdate(ctx, existing, proposed); err != nil {
		return nil, err
	}

	k.SetContractSpecification(ctx, proposed)

	for _, proposedRecord := range newrecords {
		var existing *types.RecordSpecification = nil
		if e, found := k.GetRecordSpecification(ctx, proposedRecord.SpecificationId); found {
			existing = &e
		}
		if err := k.ValidateRecordSpecUpdate(ctx, existing, proposedRecord); err != nil {
			return nil, err
		}

		k.SetRecordSpecification(ctx, proposedRecord)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, strings.Join(msg.Signers, ",")),
		),
	)

	return &types.MsgAddP8EContractSpecResponse{}, nil
}

func (k msgServer) P8EMemorializeContract(
	goCtx context.Context,
	msg *types.MsgP8EMemorializeContractRequest, //nolint:staticcheck // Ignore deprecation error here.
) (*types.MsgP8EMemorializeContractResponse, error) { //nolint:staticcheck // Ignore deprecation error here.
	ctx := sdk.UnwrapSDKContext(goCtx)

	p8EData, signers, err := types.ConvertP8eMemorializeContractRequest(msg)
	if err != nil {
		return nil, err
	}

	// Add the stuff that needs to come from the specs.
	err = k.IterateScopeSpecsForContractSpec(ctx, p8EData.Session.SpecificationId, func(scopeSpecID types.MetadataAddress) bool {
		p8EData.Scope.SpecificationId = scopeSpecID
		return true
	})
	if err != nil {
		return nil, err
	}
	if p8EData.Scope.SpecificationId.Empty() {
		return nil, fmt.Errorf("no scope specifications found associated with contract specification %s",
			p8EData.Session.SpecificationId)
	}
	var processID types.ProcessID
	contractSpec, found := k.GetContractSpecification(ctx, p8EData.Session.SpecificationId)
	if !found {
		return nil, fmt.Errorf("contract specification %s not found", p8EData.Session.SpecificationId)
	}
	switch source := contractSpec.Source.(type) {
	case *types.ContractSpecification_ResourceId:
		processID = &types.Process_Address{Address: source.ResourceId.String()}
	case *types.ContractSpecification_Hash:
		processID = &types.Process_Hash{Hash: source.Hash}
	default:
		return nil, fmt.Errorf("unexpected source type on contract specification %s", p8EData.Session.SpecificationId)
	}

	p8EData.Session.Name = contractSpec.ClassName

	for _, r := range p8EData.Records {
		r.Process.ProcessId = processID
		recSpecID, e := p8EData.Session.SpecificationId.AsRecordSpecAddress(r.Name)
		if e != nil {
			return nil, e
		}
		recSpec, found := k.GetRecordSpecification(ctx, recSpecID)
		if !found {
			return nil, fmt.Errorf("record specification %s not found", recSpecID)
		}
		for _, input := range r.Inputs {
			input.Status = types.RecordInputStatus(recSpec.ResultType)
		}
	}

	// Finally, store everything.
	_, err = k.AddScope(goCtx, &types.MsgAddScopeRequest{
		Scope:   *p8EData.Scope,
		Signers: signers,
	})
	if err != nil {
		return nil, err
	}

	_, err = k.AddSession(goCtx, &types.MsgAddSessionRequest{
		Session: p8EData.Session,
		Signers: signers,
	})
	if err != nil {
		return nil, err
	}

	for _, record := range p8EData.Records {
		_, err = k.AddRecord(goCtx, &types.MsgAddRecordRequest{
			SessionId: p8EData.Session.SessionId,
			Record:    record,
			Signers:   signers,
		})
		if err != nil {
			return nil, err
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, string(msg.Invoker)),
		),
	)

	return &types.MsgP8EMemorializeContractResponse{}, nil //nolint:staticcheck // Ignore deprecation error here.
}
