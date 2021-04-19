package keeper

import (
	"context"
	"fmt"
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

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

func (k msgServer) WriteScope(
	goCtx context.Context,
	msg *types.MsgWriteScopeRequest,
) (*types.MsgWriteScopeResponse, error) {
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

	return &types.MsgWriteScopeResponse{
		ScopeIdInfo: types.GetScopeIDInfo(msg.Scope.ScopeId),
	}, nil
}

func (k msgServer) DeleteScope(
	goCtx context.Context,
	msg *types.MsgDeleteScopeRequest,
) (*types.MsgDeleteScopeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	existing, found := k.GetScope(ctx, msg.ScopeId)
	if !found {
		return nil, fmt.Errorf("scope not found with id %s", msg.ScopeId)
	}
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

func (k msgServer) WriteSession(
	goCtx context.Context,
	msg *types.MsgWriteSessionRequest,
) (*types.MsgWriteSessionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var existing *types.Session = nil
	var existingAudit *types.AuditFields = nil
	if e, found := k.GetSession(ctx, msg.Session.SessionId); found {
		existing = &e
		existingAudit = existing.Audit
	}
	if err := k.ValidateSessionUpdate(ctx, existing, msg.Session, msg.Signers); err != nil {
		return nil, err
	}

	msg.Session.Audit = existingAudit.UpdateAudit(ctx.BlockTime(), strings.Join(msg.Signers, ", "), "")

	k.SetSession(ctx, msg.Session)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, strings.Join(msg.Signers, ",")),
		),
	)

	return &types.MsgWriteSessionResponse{
		SessionIdInfo: types.GetSessionIDInfo(msg.Session.SessionId),
	}, nil
}

func (k msgServer) WriteRecord(
	goCtx context.Context,
	msg *types.MsgWriteRecordRequest,
) (*types.MsgWriteRecordResponse, error) {
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
	if err := k.ValidateRecordUpdate(ctx, existing, msg.Record, msg.Signers, msg.Parties); err != nil {
		return nil, err
	}

	k.SetRecord(ctx, msg.Record)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, strings.Join(msg.Signers, ",")),
		),
	)

	return &types.MsgWriteRecordResponse{
		RecordIdInfo: types.GetRecordIDInfo(recordID),
	}, nil
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

func (k msgServer) WriteScopeSpecification(
	goCtx context.Context,
	msg *types.MsgWriteScopeSpecificationRequest,
) (*types.MsgWriteScopeSpecificationResponse, error) {
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

	return &types.MsgWriteScopeSpecificationResponse{
		ScopeSpecIdInfo: types.GetScopeSpecIDInfo(msg.Specification.SpecificationId),
	}, nil
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

func (k msgServer) WriteContractSpecification(
	goCtx context.Context,
	msg *types.MsgWriteContractSpecificationRequest,
) (*types.MsgWriteContractSpecificationResponse, error) {
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

	return &types.MsgWriteContractSpecificationResponse{
		ContractSpecIdInfo: types.GetContractSpecIDInfo(msg.Specification.SpecificationId),
	}, nil
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

func (k msgServer) WriteRecordSpecification(
	goCtx context.Context,
	msg *types.MsgWriteRecordSpecificationRequest,
) (*types.MsgWriteRecordSpecificationResponse, error) {
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

	return &types.MsgWriteRecordSpecificationResponse{
		RecordSpecIdInfo: types.GetRecordSpecIDInfo(msg.Specification.SpecificationId),
	}, nil
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

func (k msgServer) WriteP8EContractSpec(
	goCtx context.Context,
	msg *types.MsgWriteP8EContractSpecRequest,
) (*types.MsgWriteP8EContractSpecResponse, error) {
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

	recSpecIDInfos := make([]*types.RecordSpecIdInfo, len(newrecords))
	for i, proposedRecord := range newrecords {
		var existing *types.RecordSpecification = nil
		if e, found := k.GetRecordSpecification(ctx, proposedRecord.SpecificationId); found {
			existing = &e
		}
		if err := k.ValidateRecordSpecUpdate(ctx, existing, proposedRecord); err != nil {
			return nil, err
		}

		k.SetRecordSpecification(ctx, proposedRecord)
		recSpecIDInfos[i] = types.GetRecordSpecIDInfo(proposedRecord.SpecificationId)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, strings.Join(msg.Signers, ",")),
		),
	)

	return &types.MsgWriteP8EContractSpecResponse{
		ContractSpecIdInfo: types.GetContractSpecIDInfo(proposed.SpecificationId),
		RecordSpecIdInfos:  recSpecIDInfos,
	}, nil
}

func (k msgServer) P8EMemorializeContract(
	goCtx context.Context,
	msg *types.MsgP8EMemorializeContractRequest,
) (*types.MsgP8EMemorializeContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	p8EData, signers, err := types.ConvertP8eMemorializeContractRequest(msg)
	if err != nil {
		return nil, err
	}

	// Add the stuff that needs to come from the specs.
	for _, r := range p8EData.Records {
		recSpecID, e := p8EData.Session.SpecificationId.AsRecordSpecAddress(r.Name)
		if e != nil {
			return nil, e
		}
		recSpec, found := k.GetRecordSpecification(ctx, recSpecID)
		if !found {
			return nil, fmt.Errorf("record specification %s not found", recSpecID)
		}
		inputStatus := types.RecordInputStatus_Unknown
		switch recSpec.ResultType {
		case types.DefinitionType_DEFINITION_TYPE_PROPOSED:
			inputStatus = types.RecordInputStatus_Proposed
		case types.DefinitionType_DEFINITION_TYPE_RECORD, types.DefinitionType_DEFINITION_TYPE_RECORD_LIST:
			inputStatus = types.RecordInputStatus_Record
		}
		// r.Inputs is a list of structs (not a list of references).
		// Need to use the list index to update the actual entry and make it stay.
		for i := range r.Inputs {
			r.Inputs[i].Status = inputStatus
		}
	}

	// Finally, store everything.
	scopeResp, err := k.WriteScope(goCtx, &types.MsgWriteScopeRequest{
		Scope:   *p8EData.Scope,
		Signers: signers,
	})
	if err != nil {
		return nil, err
	}

	sessionResp, err := k.WriteSession(goCtx, &types.MsgWriteSessionRequest{
		Session: *p8EData.Session,
		Signers: signers,
	})
	if err != nil {
		return nil, err
	}

	recordIDInfos := make([]*types.RecordIdInfo, len(p8EData.Records))
	for i, record := range p8EData.Records {
		recordResp, err := k.WriteRecord(goCtx, &types.MsgWriteRecordRequest{
			Record:  *record,
			Signers: signers,
			Parties: p8EData.Session.Parties,
		})
		if err != nil {
			return nil, err
		}
		recordIDInfos[i] = recordResp.RecordIdInfo
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Invoker),
		),
	)

	return &types.MsgP8EMemorializeContractResponse{
		ScopeIdInfo:   scopeResp.ScopeIdInfo,
		SessionIdInfo: sessionResp.SessionIdInfo,
		RecordIdInfos: recordIDInfos,
	}, nil
}

func (k msgServer) BindOSLocator(goCtx context.Context, msg *types.MsgBindOSLocatorRequest) (*types.MsgBindOSLocatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Validate
	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Error("unable to validate message", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	// already valid address, checked in ValidateBasic
	address, _ := sdk.AccAddressFromBech32(msg.Locator.Owner)

	if k.Keeper.OSLocatorExists(ctx, address) {
		ctx.Logger().Error("Address already bound to an URI", "owner", msg.Locator.Owner)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, types.ErrOSLocatorAlreadyBound.Error())
	}

	// Bind owner to URI
	if err := k.Keeper.SetOSLocatorRecord(ctx, address, msg.Locator.LocatorUri); err != nil {
		ctx.Logger().Error("unable to bind name", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	// Emit event and return
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeOsLocatorCreated,
			sdk.NewAttribute(types.AttributeKeyOSLocatorAddress, msg.Locator.Owner),
			sdk.NewAttribute(types.AttributeKeyOSLocatorURI, msg.Locator.LocatorUri),
		),
	)

	return &types.MsgBindOSLocatorResponse{}, nil
}

func (k msgServer) DeleteOSLocator(ctx context.Context, msg *types.MsgDeleteOSLocatorRequest) (*types.MsgDeleteOSLocatorResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Validate
	if err := msg.ValidateBasic(); err != nil {
		sdkCtx.Logger().Error("unable to validate message", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	// already valid address, checked in ValidateBasic
	ownerAddr, _ := sdk.AccAddressFromBech32(msg.Locator.Owner)

	if !k.Keeper.OSLocatorExists(sdkCtx, ownerAddr) {
		sdkCtx.Logger().Error("Address not already bound to an URI", "owner", msg.Locator.Owner)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, types.ErrOSLocatorAlreadyBound.Error())
	}

	if !k.Keeper.VerifyCorrectOwner(sdkCtx, ownerAddr) {
		sdkCtx.Logger().Error("msg sender cannot delete os locator", "owner", ownerAddr)
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "msg sender cannot delete os locator.")
	}

	// Delete
	if err := k.Keeper.DeleteRecord(sdkCtx, ownerAddr); err != nil {
		sdkCtx.Logger().Error("error deleting name", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	// Emit event and return
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeOsLocatorDeleted,
			sdk.NewAttribute(types.AttributeKeyOSLocatorAddress, msg.Locator.Owner),
			sdk.NewAttribute(types.AttributeKeyOSLocatorURI, msg.Locator.LocatorUri),
		),
	)
	return &types.MsgDeleteOSLocatorResponse{Locator: msg.Locator}, nil
}

func (k msgServer) ModifyOSLocator(ctx context.Context, msg *types.MsgModifyOSLocatorRequest) (*types.MsgModifyOSLocatorResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Validate
	if err := msg.ValidateBasic(); err != nil {
		sdkCtx.Logger().Error("unable to validate message", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	// already valid address, checked in ValidateBasic
	ownerAddr, _ := sdk.AccAddressFromBech32(msg.Locator.Owner)

	if !k.Keeper.OSLocatorExists(sdkCtx, ownerAddr) {
		sdkCtx.Logger().Error("Address not already bound to an URI", "owner", msg.Locator.Owner)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, types.ErrOSLocatorAlreadyBound.Error())
	}

	if !k.Keeper.VerifyCorrectOwner(sdkCtx, ownerAddr) {
		sdkCtx.Logger().Error("msg sender cannot modify os locator", "owner", ownerAddr)
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "msg sender cannot delete os locator.")
	}
	// Modify
	if err := k.Keeper.ModifyRecord(sdkCtx, ownerAddr, msg.Locator.LocatorUri); err != nil {
		sdkCtx.Logger().Error("error deleting name", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	// Emit event and return
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeOsLocatorModified,
			sdk.NewAttribute(types.AttributeKeyOSLocatorAddress, msg.Locator.Owner),
			sdk.NewAttribute(types.AttributeKeyOSLocatorURI, msg.Locator.LocatorUri),
		),
	)
	return &types.MsgModifyOSLocatorResponse{Locator: msg.Locator}, nil
}
