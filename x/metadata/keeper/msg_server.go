package keeper

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
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
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "WriteScope")
	ctx := sdk.UnwrapSDKContext(goCtx)

	//nolint:errcheck // the error was checked when msg.ValidateBasic was called before getting here.
	msg.ConvertOptionalFields()

	existing, _ := k.GetScope(ctx, msg.Scope.ScopeId)
	if err := k.ValidateScopeUpdate(ctx, existing, msg.Scope, msg.Signers, msg.MsgTypeURL()); err != nil {
		return nil, err
	}

	k.SetScope(ctx, msg.Scope)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_WriteScope, msg.GetSigners()))
	return types.NewMsgWriteScopeResponse(msg.Scope.ScopeId), nil
}

func (k msgServer) DeleteScope(
	goCtx context.Context,
	msg *types.MsgDeleteScopeRequest,
) (*types.MsgDeleteScopeResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "DeleteScope")
	ctx := sdk.UnwrapSDKContext(goCtx)

	if len(msg.ScopeId) == 0 {
		return nil, errors.New("scope id cannot be empty")
	}
	existing, found := k.GetScope(ctx, msg.ScopeId)
	if !found {
		return nil, fmt.Errorf("scope not found with id %s", msg.ScopeId)
	}

	if err := k.ValidateScopeRemove(ctx, existing, msg.Signers, msg.MsgTypeURL()); err != nil {
		return nil, err
	}

	k.RemoveScope(ctx, msg.ScopeId)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_DeleteScope, msg.GetSigners()))
	return types.NewMsgDeleteScopeResponse(), nil
}

func (k msgServer) AddScopeDataAccess(
	goCtx context.Context,
	msg *types.MsgAddScopeDataAccessRequest,
) (*types.MsgAddScopeDataAccessResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "AddScopeDataAccess")
	ctx := sdk.UnwrapSDKContext(goCtx)

	existing, found := k.GetScope(ctx, msg.ScopeId)
	if !found {
		return nil, fmt.Errorf("scope not found with id %s", msg.ScopeId)
	}

	if err := k.ValidateScopeAddDataAccess(ctx, msg.DataAccess, existing, msg.Signers, msg.MsgTypeURL()); err != nil {
		return nil, err
	}

	existing.AddDataAccess(msg.DataAccess)

	k.SetScope(ctx, existing)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_AddScopeDataAccess, msg.GetSigners()))
	return types.NewMsgAddScopeDataAccessResponse(), nil
}

func (k msgServer) DeleteScopeDataAccess(
	goCtx context.Context,
	msg *types.MsgDeleteScopeDataAccessRequest,
) (*types.MsgDeleteScopeDataAccessResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "DeleteScopeDataAccess")
	ctx := sdk.UnwrapSDKContext(goCtx)

	existing, found := k.GetScope(ctx, msg.ScopeId)
	if !found {
		return nil, fmt.Errorf("scope not found with id %s", msg.ScopeId)
	}

	if err := k.ValidateScopeDeleteDataAccess(ctx, msg.DataAccess, existing, msg.Signers, msg.MsgTypeURL()); err != nil {
		return nil, err
	}

	existing.RemoveDataAccess(msg.DataAccess)

	k.SetScope(ctx, existing)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_DeleteScopeDataAccess, msg.GetSigners()))
	return types.NewMsgDeleteScopeDataAccessResponse(), nil
}

func (k msgServer) AddScopeOwner(
	goCtx context.Context,
	msg *types.MsgAddScopeOwnerRequest,
) (*types.MsgAddScopeOwnerResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "AddScopeOwner")
	ctx := sdk.UnwrapSDKContext(goCtx)

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

	if err := k.ValidateScopeUpdateOwners(ctx, existing, proposed, msg.Signers, msg.MsgTypeURL()); err != nil {
		return nil, err
	}

	k.SetScope(ctx, proposed)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_AddScopeOwner, msg.GetSigners()))
	return types.NewMsgAddScopeOwnerResponse(), nil
}

func (k msgServer) DeleteScopeOwner(
	goCtx context.Context,
	msg *types.MsgDeleteScopeOwnerRequest,
) (*types.MsgDeleteScopeOwnerResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "DeleteScopeOwner")
	ctx := sdk.UnwrapSDKContext(goCtx)

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

	if err := k.ValidateScopeUpdateOwners(ctx, existing, proposed, msg.Signers, msg.MsgTypeURL()); err != nil {
		return nil, err
	}

	k.SetScope(ctx, proposed)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_DeleteScopeOwner, msg.GetSigners()))
	return types.NewMsgDeleteScopeOwnerResponse(), nil
}

func (k msgServer) WriteSession(
	goCtx context.Context,
	msg *types.MsgWriteSessionRequest,
) (*types.MsgWriteSessionResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "WriteSession")
	ctx := sdk.UnwrapSDKContext(goCtx)

	//nolint:errcheck // the error was checked when msg.ValidateBasic was called before getting here.
	msg.ConvertOptionalFields()

	var existing *types.Session
	var existingAudit *types.AuditFields
	if e, found := k.GetSession(ctx, msg.Session.SessionId); found {
		existing = &e
		existingAudit = existing.Audit
	}
	if err := k.ValidateSessionUpdate(ctx, existing, &msg.Session, msg.Signers, msg.MsgTypeURL()); err != nil {
		return nil, err
	}

	msg.Session.Audit = existingAudit.UpdateAudit(ctx.BlockTime(), strings.Join(msg.Signers, ", "), "")

	k.SetSession(ctx, msg.Session)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_WriteSession, msg.GetSigners()))
	return types.NewMsgWriteSessionResponse(msg.Session.SessionId), nil
}

func (k msgServer) WriteRecord(
	goCtx context.Context,
	msg *types.MsgWriteRecordRequest,
) (*types.MsgWriteRecordResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "WriteRecord")
	ctx := sdk.UnwrapSDKContext(goCtx)

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
	if err := k.ValidateRecordUpdate(ctx, existing, &msg.Record, msg.Signers, msg.Parties, msg.MsgTypeURL()); err != nil {
		return nil, err
	}

	k.SetRecord(ctx, msg.Record)

	// Remove the old session if it doesn't have any records in it anymore.
	// Note that the RemoveSession does the record checking part.
	if existing != nil && !existing.SessionId.Equals(msg.Record.SessionId) {
		k.RemoveSession(ctx, existing.SessionId)
	}

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_WriteRecord, msg.GetSigners()))
	return types.NewMsgWriteRecordResponse(recordID), nil
}

func (k msgServer) DeleteRecord(
	goCtx context.Context,
	msg *types.MsgDeleteRecordRequest,
) (*types.MsgDeleteRecordResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "DeleteRecord")
	ctx := sdk.UnwrapSDKContext(goCtx)

	existing, _ := k.GetRecord(ctx, msg.RecordId)
	if err := k.ValidateRecordRemove(ctx, existing, msg.RecordId, msg.Signers, msg.MsgTypeURL()); err != nil {
		return nil, err
	}

	k.RemoveRecord(ctx, msg.RecordId)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_DeleteRecord, msg.GetSigners()))
	return types.NewMsgDeleteRecordResponse(), nil
}

func (k msgServer) WriteScopeSpecification(
	goCtx context.Context,
	msg *types.MsgWriteScopeSpecificationRequest,
) (*types.MsgWriteScopeSpecificationResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "WriteScopeSpecification")
	ctx := sdk.UnwrapSDKContext(goCtx)

	//nolint:errcheck // the error was checked when msg.ValidateBasic was called before getting here.
	msg.ConvertOptionalFields()

	var existing *types.ScopeSpecification
	if e, found := k.GetScopeSpecification(ctx, msg.Specification.SpecificationId); found {
		existing = &e
		if err := k.ValidateAllOwnersAreSignersWithAuthz(ctx, existing.OwnerAddresses, msg.Signers, msg.MsgTypeURL()); err != nil {
			return nil, err
		}
	}
	if err := k.ValidateScopeSpecUpdate(ctx, existing, msg.Specification); err != nil {
		return nil, err
	}

	k.SetScopeSpecification(ctx, msg.Specification)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_WriteScopeSpecification, msg.GetSigners()))
	return types.NewMsgWriteScopeSpecificationResponse(msg.Specification.SpecificationId), nil
}

func (k msgServer) DeleteScopeSpecification(
	goCtx context.Context,
	msg *types.MsgDeleteScopeSpecificationRequest,
) (*types.MsgDeleteScopeSpecificationResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "DeleteScopeSpecification")
	ctx := sdk.UnwrapSDKContext(goCtx)

	existing, found := k.GetScopeSpecification(ctx, msg.SpecificationId)
	if !found {
		return nil, fmt.Errorf("scope specification not found with id %s", msg.SpecificationId)
	}
	if err := k.ValidateAllOwnersAreSignersWithAuthz(ctx, existing.OwnerAddresses, msg.Signers, msg.MsgTypeURL()); err != nil {
		return nil, err
	}

	if err := k.RemoveScopeSpecification(ctx, msg.SpecificationId); err != nil {
		return nil, fmt.Errorf("cannot delete scope specification with id %s: %w", msg.SpecificationId, err)
	}

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_DeleteScopeSpecification, msg.GetSigners()))
	return types.NewMsgDeleteScopeSpecificationResponse(), nil
}

func (k msgServer) WriteContractSpecification(
	goCtx context.Context,
	msg *types.MsgWriteContractSpecificationRequest,
) (*types.MsgWriteContractSpecificationResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "WriteContractSpecification")
	ctx := sdk.UnwrapSDKContext(goCtx)

	//nolint:errcheck // the error was checked when msg.ValidateBasic was called before getting here.
	msg.ConvertOptionalFields()

	var existing *types.ContractSpecification
	if e, found := k.GetContractSpecification(ctx, msg.Specification.SpecificationId); found {
		existing = &e
		if err := k.ValidateAllOwnersAreSignersWithAuthz(ctx, existing.OwnerAddresses, msg.Signers, msg.MsgTypeURL()); err != nil {
			return nil, err
		}
	}
	if err := k.ValidateContractSpecUpdate(ctx, existing, msg.Specification); err != nil {
		return nil, err
	}

	k.SetContractSpecification(ctx, msg.Specification)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_WriteContractSpecification, msg.GetSigners()))
	return types.NewMsgWriteContractSpecificationResponse(msg.Specification.SpecificationId), nil
}

func (k msgServer) DeleteContractSpecification(
	goCtx context.Context,
	msg *types.MsgDeleteContractSpecificationRequest,
) (*types.MsgDeleteContractSpecificationResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "DeleteContractSpecification")
	ctx := sdk.UnwrapSDKContext(goCtx)

	existing, found := k.GetContractSpecification(ctx, msg.SpecificationId)
	if !found {
		return nil, fmt.Errorf("contract specification not found with id %s", msg.SpecificationId)
	}
	if err := k.ValidateAllOwnersAreSignersWithAuthz(ctx, existing.OwnerAddresses, msg.Signers, msg.MsgTypeURL()); err != nil {
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

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_DeleteContractSpecification, msg.GetSigners()))
	return types.NewMsgDeleteContractSpecificationResponse(), nil
}

func (k msgServer) AddContractSpecToScopeSpec(
	goCtx context.Context,
	msg *types.MsgAddContractSpecToScopeSpecRequest,
) (*types.MsgAddContractSpecToScopeSpecResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "AddContractSpecToScopeSpec")
	ctx := sdk.UnwrapSDKContext(goCtx)
	_, found := k.GetContractSpecification(ctx, msg.ContractSpecificationId)
	if !found {
		return nil, fmt.Errorf("contract specification not found with id %s", msg.ContractSpecificationId)
	}

	scopeSpec, found := k.GetScopeSpecification(ctx, msg.ScopeSpecificationId)
	if !found {
		return nil, fmt.Errorf("scope specification not found with id %s", msg.ScopeSpecificationId)
	}
	if err := k.ValidateAllOwnersAreSignersWithAuthz(ctx, scopeSpec.OwnerAddresses, msg.Signers, msg.MsgTypeURL()); err != nil {
		return nil, err
	}

	for _, cSpecID := range scopeSpec.ContractSpecIds {
		if cSpecID.Equals(msg.ContractSpecificationId) {
			return nil, fmt.Errorf("scope spec %s already contains contract spec %s", msg.ScopeSpecificationId, msg.ContractSpecificationId)
		}
	}

	scopeSpec.ContractSpecIds = append(scopeSpec.ContractSpecIds, msg.ContractSpecificationId)
	k.SetScopeSpecification(ctx, scopeSpec)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_AddContractSpecToScopeSpec, msg.GetSigners()))
	return types.NewMsgAddContractSpecToScopeSpecResponse(), nil
}

func (k msgServer) DeleteContractSpecFromScopeSpec(
	goCtx context.Context,
	msg *types.MsgDeleteContractSpecFromScopeSpecRequest,
) (*types.MsgDeleteContractSpecFromScopeSpecResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "DeleteContractSpecFromScopeSpec")
	ctx := sdk.UnwrapSDKContext(goCtx)
	_, found := k.GetContractSpecification(ctx, msg.ContractSpecificationId)
	if !found {
		return nil, fmt.Errorf("contract specification not found with id %s", msg.ContractSpecificationId)
	}

	scopeSpec, found := k.GetScopeSpecification(ctx, msg.ScopeSpecificationId)
	if !found {
		return nil, fmt.Errorf("scope specification not found with id %s", msg.ScopeSpecificationId)
	}
	if err := k.ValidateAllOwnersAreSignersWithAuthz(ctx, scopeSpec.OwnerAddresses, msg.Signers, msg.MsgTypeURL()); err != nil {
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
		return nil, fmt.Errorf("contract specification %s not found on scope specification id %s", msg.ContractSpecificationId, msg.ScopeSpecificationId)
	}

	scopeSpec.ContractSpecIds = updateContractSpecIds
	k.SetScopeSpecification(ctx, scopeSpec)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_DeleteContractSpecFromScopeSpec, msg.GetSigners()))

	return types.NewMsgDeleteContractSpecFromScopeSpecResponse(), nil
}

func (k msgServer) WriteRecordSpecification(
	goCtx context.Context,
	msg *types.MsgWriteRecordSpecificationRequest,
) (*types.MsgWriteRecordSpecificationResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "WriteRecordSpecification")
	ctx := sdk.UnwrapSDKContext(goCtx)

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
	if err := k.ValidateAllOwnersAreSignersWithAuthz(ctx, contractSpec.OwnerAddresses, msg.Signers, msg.MsgTypeURL()); err != nil {
		return nil, err
	}

	var existing *types.RecordSpecification
	if e, found := k.GetRecordSpecification(ctx, msg.Specification.SpecificationId); found {
		existing = &e
	}
	if err := k.ValidateRecordSpecUpdate(ctx, existing, msg.Specification); err != nil {
		return nil, err
	}

	k.SetRecordSpecification(ctx, msg.Specification)

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_WriteRecordSpecification, msg.GetSigners()))
	return types.NewMsgWriteRecordSpecificationResponse(msg.Specification.SpecificationId), nil
}

func (k msgServer) DeleteRecordSpecification(
	goCtx context.Context,
	msg *types.MsgDeleteRecordSpecificationRequest,
) (*types.MsgDeleteRecordSpecificationResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "DeleteRecordSpecification")
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
	if err := k.ValidateAllOwnersAreSignersWithAuthz(ctx, contractSpec.OwnerAddresses, msg.Signers, msg.MsgTypeURL()); err != nil {
		return nil, err
	}

	if err := k.RemoveRecordSpecification(ctx, msg.SpecificationId); err != nil {
		return nil, fmt.Errorf("cannot delete record specification with id %s: %w", msg.SpecificationId, err)
	}

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_DeleteRecordSpecification, msg.GetSigners()))
	return types.NewMsgDeleteRecordSpecificationResponse(), nil
}

func (k msgServer) WriteP8EContractSpec(
	goCtx context.Context,
	msg *types.MsgWriteP8EContractSpecRequest,
) (*types.MsgWriteP8EContractSpecResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "WriteP8EContractSpec")
	ctx := sdk.UnwrapSDKContext(goCtx)

	proposed, newrecords, err := types.ConvertP8eContractSpec(&msg.Contractspec, msg.Signers)
	if err != nil {
		return nil, err
	}

	var existing *types.ContractSpecification
	if e, found := k.GetContractSpecification(ctx, proposed.SpecificationId); found {
		existing = &e
		if err := k.ValidateAllOwnersAreSignersWithAuthz(ctx, existing.OwnerAddresses, msg.Signers, msg.MsgTypeURL()); err != nil {
			return nil, err
		}
	}

	if err := k.ValidateContractSpecUpdate(ctx, existing, proposed); err != nil {
		return nil, err
	}

	k.SetContractSpecification(ctx, proposed)

	recSpecIDs := make([]types.MetadataAddress, len(newrecords))
	for i, proposedRecord := range newrecords {
		var existing *types.RecordSpecification
		if e, found := k.GetRecordSpecification(ctx, proposedRecord.SpecificationId); found {
			existing = &e
		}
		if err := k.ValidateRecordSpecUpdate(ctx, existing, proposedRecord); err != nil {
			return nil, err
		}

		k.SetRecordSpecification(ctx, proposedRecord)
		recSpecIDs[i] = proposedRecord.SpecificationId
	}

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_WriteP8eContractSpec, msg.GetSigners()))
	return types.NewMsgWriteP8EContractSpecResponse(proposed.SpecificationId, recSpecIDs...), nil
}

func (k msgServer) P8EMemorializeContract(
	goCtx context.Context,
	msg *types.MsgP8EMemorializeContractRequest,
) (*types.MsgP8EMemorializeContractResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "P8EMemorializeContract")
	ctx := sdk.UnwrapSDKContext(goCtx)

	p8EData, err := types.ConvertP8eMemorializeContractRequest(msg)
	if err != nil {
		return nil, err
	}

	existingScope, found := k.GetScope(ctx, p8EData.Scope.ScopeId)
	if found {
		// We don't want to update the scope's owners or value owner.
		p8EData.Scope.Owners = existingScope.Owners
		p8EData.Scope.ValueOwnerAddress = existingScope.ValueOwnerAddress
		// We only want to add to the data access list.
		p8EData.Scope.DataAccess = k.UnionDistinct(existingScope.DataAccess, p8EData.Scope.DataAccess)
	}

	scopeResp, err := k.WriteScope(goCtx, &types.MsgWriteScopeRequest{
		Scope:   *p8EData.Scope,
		Signers: p8EData.Signers,
	})
	if err != nil {
		return nil, err
	}

	sessionResp, err := k.WriteSession(goCtx, &types.MsgWriteSessionRequest{
		Session: *p8EData.Session,
		Signers: p8EData.Signers,
	})
	if err != nil {
		return nil, err
	}

	recordIDInfos := make([]*types.RecordIdInfo, len(p8EData.RecordReqs))
	for i, recordReq := range p8EData.RecordReqs {
		// TODO: If there is a OriginalOutputHashes value, get the existing record and make sure it's output is as expected.
		recordResp, err := k.WriteRecord(goCtx, &types.MsgWriteRecordRequest{
			Record:  *recordReq.Record,
			Signers: p8EData.Signers,
			Parties: p8EData.Session.Parties,
		})
		if err != nil {
			return nil, err
		}
		recordIDInfos[i] = recordResp.RecordIdInfo
	}

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_P8eMemorializeContract, msg.GetSigners()))
	return types.NewMsgP8EMemorializeContractResponse(scopeResp.ScopeIdInfo, sessionResp.SessionIdInfo, recordIDInfos), nil
}

func (k msgServer) BindOSLocator(
	goCtx context.Context,
	msg *types.MsgBindOSLocatorRequest,
) (*types.MsgBindOSLocatorResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "BindOSLocator")
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Validate
	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Error("unable to validate message", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	// already valid address, checked in ValidateBasic
	ownerAddress, _ := sdk.AccAddressFromBech32(msg.Locator.Owner)
	encryptionKey := sdk.AccAddress{}
	if strings.TrimSpace(msg.Locator.EncryptionKey) != "" {
		encryptionKey, _ = sdk.AccAddressFromBech32(msg.Locator.EncryptionKey)
	}
	if k.Keeper.OSLocatorExists(ctx, ownerAddress) {
		ctx.Logger().Error("Address already bound to an URI", "owner", msg.Locator.Owner)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, types.ErrOSLocatorAlreadyBound.Error())
	}

	// Bind owner to URI
	if err := k.Keeper.SetOSLocator(ctx, ownerAddress, encryptionKey, msg.Locator.LocatorUri); err != nil {
		ctx.Logger().Error("unable to bind name", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_BindOSLocator, msg.GetSigners()))
	return types.NewMsgBindOSLocatorResponse(msg.Locator), nil
}

func (k msgServer) DeleteOSLocator(
	goCtx context.Context,
	msg *types.MsgDeleteOSLocatorRequest,
) (*types.MsgDeleteOSLocatorResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "DeleteOSLocator")
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Validate
	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Error("unable to validate message", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	// already valid address, checked in ValidateBasic
	ownerAddr, _ := sdk.AccAddressFromBech32(msg.Locator.Owner)

	if !k.Keeper.OSLocatorExists(ctx, ownerAddr) {
		ctx.Logger().Error("Address not already bound to an URI", "owner", msg.Locator.Owner)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, types.ErrOSLocatorAlreadyBound.Error())
	}

	if !k.Keeper.VerifyCorrectOwner(ctx, ownerAddr) {
		ctx.Logger().Error("msg sender cannot delete os locator", "owner", ownerAddr)
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "msg sender cannot delete os locator.")
	}

	// Delete
	if err := k.Keeper.RemoveOSLocator(ctx, ownerAddr); err != nil {
		ctx.Logger().Error("error deleting name", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_DeleteOSLocator, msg.GetSigners()))
	return types.NewMsgDeleteOSLocatorResponse(msg.Locator), nil
}

func (k msgServer) ModifyOSLocator(
	goCtx context.Context,
	msg *types.MsgModifyOSLocatorRequest,
) (*types.MsgModifyOSLocatorResponse, error) {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "tx", "ModifyOSLocator")
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Validate
	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Error("unable to validate message", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	// already valid address(es), checked in ValidateBasic
	ownerAddr, _ := sdk.AccAddressFromBech32(msg.Locator.Owner)
	encryptionKey := sdk.AccAddress{}
	if strings.TrimSpace(msg.Locator.EncryptionKey) != "" {
		encryptionKey, _ = sdk.AccAddressFromBech32(msg.Locator.EncryptionKey)
	}

	if !k.Keeper.OSLocatorExists(ctx, ownerAddr) {
		ctx.Logger().Error("Address not already bound to an URI", "owner", msg.Locator.Owner)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, types.ErrOSLocatorAlreadyBound.Error())
	}

	if !k.Keeper.VerifyCorrectOwner(ctx, ownerAddr) {
		ctx.Logger().Error("msg sender cannot modify os locator", "owner", ownerAddr)
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "msg sender cannot delete os locator.")
	}
	// Modify
	if err := k.Keeper.ModifyOSLocator(ctx, ownerAddr, encryptionKey, msg.Locator.LocatorUri); err != nil {
		ctx.Logger().Error("error deleting name", "err", err)
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	k.EmitEvent(ctx, types.NewEventTxCompleted(types.TxEndpoint_ModifyOSLocator, msg.GetSigners()))
	return types.NewMsgModifyOSLocatorResponse(msg.Locator), nil
}
