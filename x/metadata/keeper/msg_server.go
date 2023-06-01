package keeper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

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

// WriteScope adds or updates a scope.
func (k msgServer) WriteScope(
	goCtx context.Context,
	msg *types.MsgWriteScopeRequest,
) (*types.MsgWriteScopeResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "WriteScope")
	ctx := UnwrapMetadataContext(goCtx)

	//nolint:errcheck // the error was checked when msg.ValidateBasic was called before getting here.
	msg.ConvertOptionalFields()

	var existing *types.Scope
	if e, found := k.GetScope(ctx, msg.Scope.ScopeId); found {
		existing = &e
	}
	if err := k.ValidateWriteScope(ctx, existing, msg); err != nil {
		return nil, err
	}

	k.SetScope(ctx, msg.Scope)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_WriteScope, msg.GetSignerStrs()))
	return types.NewMsgWriteScopeResponse(msg.Scope.ScopeId), nil
}

// DeleteScope deletes a scope and all associated Records, Sessions.
func (k msgServer) DeleteScope(
	goCtx context.Context,
	msg *types.MsgDeleteScopeRequest,
) (*types.MsgDeleteScopeResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "DeleteScope")
	ctx := UnwrapMetadataContext(goCtx)

	if err := k.ValidateDeleteScope(ctx, msg); err != nil {
		return nil, err
	}

	k.RemoveScope(ctx, msg.ScopeId)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_DeleteScope, msg.GetSignerStrs()))
	return types.NewMsgDeleteScopeResponse(), nil
}

// AddScopeDataAccess adds data access AccAddress to scope
func (k msgServer) AddScopeDataAccess(
	goCtx context.Context,
	msg *types.MsgAddScopeDataAccessRequest,
) (*types.MsgAddScopeDataAccessResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "AddScopeDataAccess")
	ctx := UnwrapMetadataContext(goCtx)

	existing, found := k.GetScope(ctx, msg.ScopeId)
	if !found {
		return nil, fmt.Errorf("scope not found with id %s", msg.ScopeId)
	}

	if err := k.ValidateAddScopeDataAccess(ctx, existing, msg); err != nil {
		return nil, err
	}

	existing.AddDataAccess(msg.DataAccess)

	k.SetScope(ctx, existing)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_AddScopeDataAccess, msg.GetSignerStrs()))
	return types.NewMsgAddScopeDataAccessResponse(), nil
}

// DeleteScopeDataAccess removes data access AccAddress from scope
func (k msgServer) DeleteScopeDataAccess(
	goCtx context.Context,
	msg *types.MsgDeleteScopeDataAccessRequest,
) (*types.MsgDeleteScopeDataAccessResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "DeleteScopeDataAccess")
	ctx := UnwrapMetadataContext(goCtx)

	existing, found := k.GetScope(ctx, msg.ScopeId)
	if !found {
		return nil, fmt.Errorf("scope not found with id %s", msg.ScopeId)
	}

	if err := k.ValidateDeleteScopeDataAccess(ctx, existing, msg); err != nil {
		return nil, err
	}

	existing.RemoveDataAccess(msg.DataAccess)

	k.SetScope(ctx, existing)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_DeleteScopeDataAccess, msg.GetSignerStrs()))
	return types.NewMsgDeleteScopeDataAccessResponse(), nil
}

// AddScopeOwner adds new owner parties to a scope
func (k msgServer) AddScopeOwner(
	goCtx context.Context,
	msg *types.MsgAddScopeOwnerRequest,
) (*types.MsgAddScopeOwnerResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "AddScopeOwner")
	ctx := UnwrapMetadataContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	existing, found := k.GetScope(ctx, msg.ScopeId)
	if !found {
		return nil, fmt.Errorf("scope not found with id %s", msg.ScopeId)
	}

	proposed := existing
	addErr := proposed.AddOwners(msg.Owners)
	if addErr != nil {
		return nil, addErr
	}

	if err := k.ValidateUpdateScopeOwners(ctx, existing, proposed, msg); err != nil {
		return nil, err
	}

	k.SetScope(ctx, proposed)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_AddScopeOwner, msg.GetSignerStrs()))
	return types.NewMsgAddScopeOwnerResponse(), nil
}

// DeleteScopeOwner removes owner parties (by addresses) from a scope
func (k msgServer) DeleteScopeOwner(
	goCtx context.Context,
	msg *types.MsgDeleteScopeOwnerRequest,
) (*types.MsgDeleteScopeOwnerResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "DeleteScopeOwner")
	ctx := UnwrapMetadataContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	existing, found := k.GetScope(ctx, msg.ScopeId)
	if !found {
		return nil, fmt.Errorf("scope not found with id %s", msg.ScopeId)
	}

	proposed := existing
	rmErr := proposed.RemoveOwners(msg.Owners)
	if rmErr != nil {
		return nil, rmErr
	}

	if err := k.ValidateUpdateScopeOwners(ctx, existing, proposed, msg); err != nil {
		return nil, err
	}

	k.SetScope(ctx, proposed)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_DeleteScopeOwner, msg.GetSignerStrs()))
	return types.NewMsgDeleteScopeOwnerResponse(), nil
}

// UpdateValueOwners sets the value owner of one or more scopes.
func (k msgServer) UpdateValueOwners(
	goCtx context.Context,
	msg *types.MsgUpdateValueOwnersRequest,
) (*types.MsgUpdateValueOwnersResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "UpdateValueOwners")
	ctx := UnwrapMetadataContext(goCtx)

	scopes := make([]*types.Scope, len(msg.ScopeIds))
	for i, id := range msg.ScopeIds {
		scope, found := k.GetScope(ctx, id)
		if !found {
			return nil, fmt.Errorf("scope not found with id %s", id)
		}
		scopes[i] = &scope
	}

	err := k.ValidateUpdateValueOwners(ctx, scopes, msg.ValueOwnerAddress, msg)
	if err != nil {
		return nil, err
	}

	k.SetScopeValueOwners(ctx, scopes, msg.ValueOwnerAddress)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_UpdateValueOwners, msg.GetSignerStrs()))
	return &types.MsgUpdateValueOwnersResponse{}, nil
}

// MigrateValueOwner updates all scopes that have one value owner to have a another value owner.
func (k msgServer) MigrateValueOwner(
	goCtx context.Context,
	msg *types.MsgMigrateValueOwnerRequest,
) (*types.MsgMigrateValueOwnerResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "MigrateValueOwner")
	ctx := UnwrapMetadataContext(goCtx)

	var scopes []*types.Scope
	err := k.IterateScopesForValueOwner(ctx, msg.Existing, func(scopeID types.MetadataAddress) (stop bool) {
		scope, found := k.GetScope(ctx, scopeID)
		if found {
			scopes = append(scopes, &scope)
		}
		return false
	})
	if err != nil {
		return nil, err
	}
	if len(scopes) == 0 {
		return nil, fmt.Errorf("no scopes found with value owner %q", msg.Existing)
	}

	err = k.ValidateUpdateValueOwners(ctx, scopes, msg.Proposed, msg)
	if err != nil {
		return nil, err
	}

	k.SetScopeValueOwners(ctx, scopes, msg.Proposed)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_MigrateValueOwner, msg.GetSignerStrs()))
	return &types.MsgMigrateValueOwnerResponse{}, nil
}

// WriteSession adds or updates a session context.
func (k msgServer) WriteSession(
	goCtx context.Context,
	msg *types.MsgWriteSessionRequest,
) (*types.MsgWriteSessionResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "WriteSession")
	ctx := UnwrapMetadataContext(goCtx)

	//nolint:errcheck // the error was checked when msg.ValidateBasic was called before getting here.
	msg.ConvertOptionalFields()

	var existing *types.Session
	var existingAudit *types.AuditFields
	if e, found := k.GetSession(ctx, msg.Session.SessionId); found {
		existing = &e
		existingAudit = existing.Audit
	}
	if err := k.ValidateWriteSession(ctx, existing, msg); err != nil {
		return nil, err
	}

	msg.Session.Audit = existingAudit.UpdateAudit(ctx.BlockTime(), strings.Join(msg.Signers, ", "), "")

	k.SetSession(ctx, msg.Session)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_WriteSession, msg.GetSignerStrs()))
	return types.NewMsgWriteSessionResponse(msg.Session.SessionId), nil
}

// WriteRecord adds or updates a record.
func (k msgServer) WriteRecord(
	goCtx context.Context,
	msg *types.MsgWriteRecordRequest,
) (*types.MsgWriteRecordResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "WriteRecord")
	ctx := UnwrapMetadataContext(goCtx)

	//nolint:errcheck // the error was checked when msg.ValidateBasic was called before getting here.
	msg.ConvertOptionalFields()

	scopeUUID, err := msg.Record.SessionId.ScopeUUID()
	if err != nil {
		return nil, err
	}

	recordID := types.RecordMetadataAddress(scopeUUID, msg.Record.Name)

	var existing *types.Record
	if e, found := k.GetRecord(ctx, recordID); found {
		existing = &e
	}
	if err = k.ValidateWriteRecord(ctx, existing, msg); err != nil {
		return nil, err
	}

	k.SetRecord(ctx, msg.Record)

	// Remove the old session if it doesn't have any records in it anymore.
	// Note that the RemoveSession does the record checking part.
	if existing != nil && !existing.SessionId.Equals(msg.Record.SessionId) {
		k.RemoveSession(ctx, existing.SessionId)
	}

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_WriteRecord, msg.GetSignerStrs()))
	return types.NewMsgWriteRecordResponse(recordID), nil
}

// DeleteRecord deletes a record.
func (k msgServer) DeleteRecord(
	goCtx context.Context,
	msg *types.MsgDeleteRecordRequest,
) (*types.MsgDeleteRecordResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "DeleteRecord")
	ctx := UnwrapMetadataContext(goCtx)

	if err := k.ValidateDeleteRecord(ctx, msg.RecordId, msg); err != nil {
		return nil, err
	}

	k.RemoveRecord(ctx, msg.RecordId)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_DeleteRecord, msg.GetSignerStrs()))
	return types.NewMsgDeleteRecordResponse(), nil
}

// WriteScopeSpecification adds or updates a scope specification.
func (k msgServer) WriteScopeSpecification(
	goCtx context.Context,
	msg *types.MsgWriteScopeSpecificationRequest,
) (*types.MsgWriteScopeSpecificationResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "WriteScopeSpecification")
	ctx := UnwrapMetadataContext(goCtx)

	//nolint:errcheck // the error was checked when msg.ValidateBasic was called before getting here.
	msg.ConvertOptionalFields()

	var existing *types.ScopeSpecification
	if e, found := k.GetScopeSpecification(ctx, msg.Specification.SpecificationId); found {
		existing = &e
		if err := k.ValidateSignersWithoutParties(ctx, existing.OwnerAddresses, msg); err != nil {
			return nil, err
		}
	}
	if err := k.ValidateWriteScopeSpecification(ctx, existing, msg.Specification); err != nil {
		return nil, err
	}

	k.SetScopeSpecification(ctx, msg.Specification)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_WriteScopeSpecification, msg.GetSignerStrs()))
	return types.NewMsgWriteScopeSpecificationResponse(msg.Specification.SpecificationId), nil
}

// DeleteScopeSpecification deletes a scope specification.
func (k msgServer) DeleteScopeSpecification(
	goCtx context.Context,
	msg *types.MsgDeleteScopeSpecificationRequest,
) (*types.MsgDeleteScopeSpecificationResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "DeleteScopeSpecification")
	ctx := UnwrapMetadataContext(goCtx)

	existing, found := k.GetScopeSpecification(ctx, msg.SpecificationId)
	if !found {
		return nil, fmt.Errorf("scope specification not found with id %s", msg.SpecificationId)
	}
	if err := k.ValidateSignersWithoutParties(ctx, existing.OwnerAddresses, msg); err != nil {
		return nil, err
	}

	if err := k.RemoveScopeSpecification(ctx, msg.SpecificationId); err != nil {
		return nil, fmt.Errorf("cannot delete scope specification with id %s: %w", msg.SpecificationId, err)
	}

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_DeleteScopeSpecification, msg.GetSignerStrs()))
	return types.NewMsgDeleteScopeSpecificationResponse(), nil
}

// WriteContractSpecification adds or updates a contract specification.
func (k msgServer) WriteContractSpecification(
	goCtx context.Context,
	msg *types.MsgWriteContractSpecificationRequest,
) (*types.MsgWriteContractSpecificationResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "WriteContractSpecification")
	ctx := UnwrapMetadataContext(goCtx)

	//nolint:errcheck // the error was checked when msg.ValidateBasic was called before getting here.
	msg.ConvertOptionalFields()

	var existing *types.ContractSpecification
	if e, found := k.GetContractSpecification(ctx, msg.Specification.SpecificationId); found {
		existing = &e
		if err := k.ValidateSignersWithoutParties(ctx, existing.OwnerAddresses, msg); err != nil {
			return nil, err
		}
	}
	if err := k.ValidateWriteContractSpecification(ctx, existing, msg.Specification); err != nil {
		return nil, err
	}

	k.SetContractSpecification(ctx, msg.Specification)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_WriteContractSpecification, msg.GetSignerStrs()))
	return types.NewMsgWriteContractSpecificationResponse(msg.Specification.SpecificationId), nil
}

// DeleteContractSpecification deletes a contract specification.
func (k msgServer) DeleteContractSpecification(
	goCtx context.Context,
	msg *types.MsgDeleteContractSpecificationRequest,
) (*types.MsgDeleteContractSpecificationResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "DeleteContractSpecification")
	ctx := UnwrapMetadataContext(goCtx)

	existing, found := k.GetContractSpecification(ctx, msg.SpecificationId)
	if !found {
		return nil, fmt.Errorf("contract specification not found with id %s", msg.SpecificationId)
	}
	if err := k.ValidateSignersWithoutParties(ctx, existing.OwnerAddresses, msg); err != nil {
		return nil, err
	}

	// Remove all record specifications associated with this contract specification.
	recSpecs, recSpecErr := k.GetRecordSpecificationsForContractSpecificationID(ctx, msg.SpecificationId)
	if recSpecErr != nil {
		return nil, fmt.Errorf("could not get record specifications to delete with contract specification with id %s: %w",
			msg.SpecificationId, recSpecErr)
	}
	var delRecSpecErr error
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

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_DeleteContractSpecification, msg.GetSignerStrs()))
	return types.NewMsgDeleteContractSpecificationResponse(), nil
}

// AddContractSpecToScopeSpec adds contract specification to a scope specification.
func (k msgServer) AddContractSpecToScopeSpec(
	goCtx context.Context,
	msg *types.MsgAddContractSpecToScopeSpecRequest,
) (*types.MsgAddContractSpecToScopeSpecResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "AddContractSpecToScopeSpec")
	ctx := UnwrapMetadataContext(goCtx)
	_, found := k.GetContractSpecification(ctx, msg.ContractSpecificationId)
	if !found {
		return nil, fmt.Errorf("contract specification not found with id %s", msg.ContractSpecificationId)
	}

	scopeSpec, found := k.GetScopeSpecification(ctx, msg.ScopeSpecificationId)
	if !found {
		return nil, fmt.Errorf("scope specification not found with id %s", msg.ScopeSpecificationId)
	}
	if err := k.ValidateSignersWithoutParties(ctx, scopeSpec.OwnerAddresses, msg); err != nil {
		return nil, err
	}

	for _, cSpecID := range scopeSpec.ContractSpecIds {
		if cSpecID.Equals(msg.ContractSpecificationId) {
			return nil, fmt.Errorf("scope spec %s already contains contract spec %s", msg.ScopeSpecificationId, msg.ContractSpecificationId)
		}
	}

	scopeSpec.ContractSpecIds = append(scopeSpec.ContractSpecIds, msg.ContractSpecificationId)
	k.SetScopeSpecification(ctx, scopeSpec)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_AddContractSpecToScopeSpec, msg.GetSignerStrs()))
	return types.NewMsgAddContractSpecToScopeSpecResponse(), nil
}

// DeleteContractSpecFromScopeSpec deletes a contract specification from a scope specification.
func (k msgServer) DeleteContractSpecFromScopeSpec(
	goCtx context.Context,
	msg *types.MsgDeleteContractSpecFromScopeSpecRequest,
) (*types.MsgDeleteContractSpecFromScopeSpecResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "DeleteContractSpecFromScopeSpec")
	ctx := UnwrapMetadataContext(goCtx)
	scopeSpec, found := k.GetScopeSpecification(ctx, msg.ScopeSpecificationId)
	if !found {
		return nil, fmt.Errorf("scope specification not found with id %s", msg.ScopeSpecificationId)
	}
	if err := k.ValidateSignersWithoutParties(ctx, scopeSpec.OwnerAddresses, msg); err != nil {
		return nil, err
	}

	updateContractSpecIds := []types.MetadataAddress{}
	found = false
	for _, cSpecID := range scopeSpec.ContractSpecIds {
		if !cSpecID.Equals(msg.ContractSpecificationId) {
			updateContractSpecIds = append(updateContractSpecIds, cSpecID)
		} else {
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("contract specification %s not found in scope specification %s", msg.ContractSpecificationId, msg.ScopeSpecificationId)
	}

	scopeSpec.ContractSpecIds = updateContractSpecIds
	k.SetScopeSpecification(ctx, scopeSpec)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_DeleteContractSpecFromScopeSpec, msg.GetSignerStrs()))

	return types.NewMsgDeleteContractSpecFromScopeSpecResponse(), nil
}

// WriteRecordSpecification adds or updates a record specification.
func (k msgServer) WriteRecordSpecification(
	goCtx context.Context,
	msg *types.MsgWriteRecordSpecificationRequest,
) (*types.MsgWriteRecordSpecificationResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "WriteRecordSpecification")
	ctx := UnwrapMetadataContext(goCtx)

	//nolint:errcheck // the error was checked when msg.ValidateBasic was called before getting here.
	msg.ConvertOptionalFields()

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
	if err = k.ValidateSignersWithoutParties(ctx, contractSpec.OwnerAddresses, msg); err != nil {
		return nil, err
	}

	var existing *types.RecordSpecification
	if e, found := k.GetRecordSpecification(ctx, msg.Specification.SpecificationId); found {
		existing = &e
	}
	if err = k.ValidateWriteRecordSpecification(ctx, existing, msg.Specification); err != nil {
		return nil, err
	}

	k.SetRecordSpecification(ctx, msg.Specification)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_WriteRecordSpecification, msg.GetSignerStrs()))
	return types.NewMsgWriteRecordSpecificationResponse(msg.Specification.SpecificationId), nil
}

// DeleteRecordSpecification deletes a record specification.
func (k msgServer) DeleteRecordSpecification(
	goCtx context.Context,
	msg *types.MsgDeleteRecordSpecificationRequest,
) (*types.MsgDeleteRecordSpecificationResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "DeleteRecordSpecification")
	ctx := UnwrapMetadataContext(goCtx)

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
	if err := k.ValidateSignersWithoutParties(ctx, contractSpec.OwnerAddresses, msg); err != nil {
		return nil, err
	}

	if err := k.RemoveRecordSpecification(ctx, msg.SpecificationId); err != nil {
		return nil, fmt.Errorf("cannot delete record specification with id %s: %w", msg.SpecificationId, err)
	}

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_DeleteRecordSpecification, msg.GetSignerStrs()))
	return types.NewMsgDeleteRecordSpecificationResponse(), nil
}

// BindOSLocator binds an owner address to a uri.
func (k msgServer) BindOSLocator(
	goCtx context.Context,
	msg *types.MsgBindOSLocatorRequest,
) (*types.MsgBindOSLocatorResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "BindOSLocator")
	ctx := UnwrapMetadataContext(goCtx)
	// Validate
	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Error("unable to validate message", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	// already valid address, checked in ValidateBasic
	ownerAddress, _ := sdk.AccAddressFromBech32(msg.Locator.Owner)
	encryptionKey := sdk.AccAddress{}
	if strings.TrimSpace(msg.Locator.EncryptionKey) != "" {
		encryptionKey, _ = sdk.AccAddressFromBech32(msg.Locator.EncryptionKey)
	}
	if k.Keeper.OSLocatorExists(ctx, ownerAddress) {
		ctx.Logger().Error("Address already bound to an URI", "owner", msg.Locator.Owner)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(types.ErrOSLocatorAlreadyBound.Error())
	}

	// Bind owner to URI
	if err := k.Keeper.SetOSLocator(ctx, ownerAddress, encryptionKey, msg.Locator.LocatorUri); err != nil {
		ctx.Logger().Error("unable to bind name", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_BindOSLocator, msg.GetSignerStrs()))
	return types.NewMsgBindOSLocatorResponse(msg.Locator), nil
}

// DeleteOSLocator deletes an existing ObjectStoreLocator record.
func (k msgServer) DeleteOSLocator(
	goCtx context.Context,
	msg *types.MsgDeleteOSLocatorRequest,
) (*types.MsgDeleteOSLocatorResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "DeleteOSLocator")
	ctx := UnwrapMetadataContext(goCtx)
	// Validate
	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Error("unable to validate message", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	// already valid address, checked in ValidateBasic
	ownerAddr, _ := sdk.AccAddressFromBech32(msg.Locator.Owner)

	if !k.Keeper.OSLocatorExists(ctx, ownerAddr) {
		ctx.Logger().Error("Address not already bound to an URI", "owner", msg.Locator.Owner)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(types.ErrOSLocatorAlreadyBound.Error())
	}

	if !k.Keeper.VerifyCorrectOwner(ctx, ownerAddr) {
		ctx.Logger().Error("msg sender cannot delete os locator", "owner", ownerAddr)
		return nil, sdkerrors.ErrUnauthorized.Wrap("msg sender cannot delete os locator.")
	}

	// Delete
	if err := k.Keeper.RemoveOSLocator(ctx, ownerAddr); err != nil {
		ctx.Logger().Error("error deleting name", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_DeleteOSLocator, msg.GetSignerStrs()))
	return types.NewMsgDeleteOSLocatorResponse(msg.Locator), nil
}

// ModifyOSLocator updates an ObjectStoreLocator record by the current owner.
func (k msgServer) ModifyOSLocator(
	goCtx context.Context,
	msg *types.MsgModifyOSLocatorRequest,
) (*types.MsgModifyOSLocatorResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "ModifyOSLocator")
	ctx := UnwrapMetadataContext(goCtx)
	// Validate
	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Error("unable to validate message", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	// already valid address(es), checked in ValidateBasic
	ownerAddr, _ := sdk.AccAddressFromBech32(msg.Locator.Owner)
	encryptionKey := sdk.AccAddress{}
	if strings.TrimSpace(msg.Locator.EncryptionKey) != "" {
		encryptionKey, _ = sdk.AccAddressFromBech32(msg.Locator.EncryptionKey)
	}

	if !k.Keeper.OSLocatorExists(ctx, ownerAddr) {
		ctx.Logger().Error("Address not already bound to an URI", "owner", msg.Locator.Owner)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(types.ErrOSLocatorAlreadyBound.Error())
	}

	if !k.Keeper.VerifyCorrectOwner(ctx, ownerAddr) {
		ctx.Logger().Error("msg sender cannot modify os locator", "owner", ownerAddr)
		return nil, sdkerrors.ErrUnauthorized.Wrap("msg sender cannot delete os locator.")
	}
	// Modify
	if err := k.Keeper.ModifyOSLocator(ctx, ownerAddr, encryptionKey, msg.Locator.LocatorUri); err != nil {
		ctx.Logger().Error("error deleting name", "err", err)
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_ModifyOSLocator, msg.GetSignerStrs()))
	return types.NewMsgModifyOSLocatorResponse(msg.Locator), nil
}

// SetAccountData associates some basic data with a metadata address.
// Currently, only scope ids are supported.
func (k msgServer) SetAccountData(
	goCtx context.Context,
	msg *types.MsgSetAccountDataRequest,
) (*types.MsgSetAccountDataResponse, error) {
	ctx := UnwrapMetadataContext(goCtx)

	if err := k.ValidateSetAccountData(ctx, msg); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	if err := k.attrKeeper.SetAccountData(ctx, msg.MetadataAddr.String(), msg.Value); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	return &types.MsgSetAccountDataResponse{}, nil
}
