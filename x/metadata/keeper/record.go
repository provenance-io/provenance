package keeper

import (
	"fmt"

	"github.com/provenance-io/provenance/x/metadata/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const sourceTypeHash = "hash"
const sourceTypeRecord = "record"

// GetRecord returns the record with the given id.
func (k Keeper) GetRecord(ctx sdk.Context, id types.MetadataAddress) (record types.Record, found bool) {
	if !id.IsRecordAddress() {
		return record, false
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(id)
	if b == nil {
		return types.Record{}, false
	}
	k.cdc.MustUnmarshalBinaryBare(b, &record)
	return record, true
}

// GetRecords returns records with scopeAddress and name
func (k Keeper) GetRecords(ctx sdk.Context, scopeAddress types.MetadataAddress, name string) ([]*types.Record, error) {
	records := []*types.Record{}
	err := k.IterateRecords(ctx, scopeAddress, func(r types.Record) (stop bool) {
		if name == "" {
			records = append(records, &r)
		} else if name == r.Name {
			records = append(records, &r)
		}
		return false
	})
	if err != nil {
		return nil, err
	}
	return records, nil
}

// SetRecord stores a record in the module kv store.
func (k Keeper) SetRecord(ctx sdk.Context, record types.Record) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryBare(&record)
	eventType := types.EventTypeRecordCreated

	recordID, err := record.SessionId.AsRecordAddress(record.Name)
	if err != nil {
		ctx.Logger().Error("could not create record id",
			"session id", record.SessionId, "name", record.Name, "error", err)
		return
	}
	if store.Has(recordID) {
		eventType = types.EventTypeRecordUpdated
	}

	store.Set(recordID, b)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			eventType,
			sdk.NewAttribute(types.AttributeKeyScopeID, recordID.String()),
		),
	)
}

// RemoveRecord removes a scope from the module kv store.
func (k Keeper) RemoveRecord(ctx sdk.Context, id types.MetadataAddress) {
	if !id.IsRecordAddress() {
		panic(fmt.Errorf("invalid address, address must be for a record"))
	}
	store := ctx.KVStore(k.storeKey)
	store.Delete(id)

	sessionUUID, _ := id.SessionUUID()
	scopeUUID, _ := id.ScopeUUID()
	sessionID := types.SessionMetadataAddress(scopeUUID, sessionUUID)
	k.RemoveSession(ctx, sessionID)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRecordRemoved,
			sdk.NewAttribute(types.AttributeKeyRecordID, id.String()),
		),
	)
}

// IterateRecords processes all records in a scope with the given handler.
func (k Keeper) IterateRecords(ctx sdk.Context, scopeID types.MetadataAddress, handler func(types.Record) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	prefix, err := scopeID.ScopeRecordIteratorPrefix()
	if err != nil {
		return err
	}
	it := sdk.KVStorePrefixIterator(store, prefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var record types.Record
		k.cdc.MustUnmarshalBinaryBare(it.Value(), &record)
		if handler(record) {
			break
		}
	}
	return nil
}

// ValidateRecordUpdate checks the current record and the proposed record to determine if the the proposed changes are valid
// based on the existing state
func (k Keeper) ValidateRecordUpdate(ctx sdk.Context, existing *types.Record, proposed types.Record, signers []string) error {
	if err := proposed.ValidateBasic(); err != nil {
		return err
	}

	if existing != nil {
		if existing.Name != proposed.Name {
			return fmt.Errorf("the Name field of records cannot be changed")
		}
		if !existing.SessionId.Equals(proposed.SessionId) || existing.Name != proposed.Name {
			return fmt.Errorf("the SessionId field of records cannot be changed")
		}
	}

	scopeUUID, err := proposed.SessionId.ScopeUUID()
	if err != nil {
		return err
	}

	scopeID := types.ScopeMetadataAddress(scopeUUID)

	// get scope for existing record
	scope, found := k.GetScope(ctx, scopeID)
	if !found {
		return fmt.Errorf("scope not found for scope uuid %s", scopeUUID)
	}

	// Make sure all the scope owners have signed.
	if signErr := k.ValidateAllPartiesAreSigners(scope.Owners, signers); signErr != nil {
		return signErr
	}

	// Get the session (and make sure it exists)
	session, found := k.GetSession(ctx, proposed.SessionId)
	if !found {
		return fmt.Errorf("session not found for session id %s", proposed.SessionId)
	}

	// Get the record specification
	contractSpecUUID, err := session.SpecificationId.ContractSpecUUID()
	if err != nil {
		return err
	}
	recSpecID := types.RecordSpecMetadataAddress(contractSpecUUID, proposed.Name)
	recSpec, found := k.GetRecordSpecification(ctx, recSpecID)
	if !found {
		return fmt.Errorf("record specification not found for record specification id %s (contract spec uuid %s and record name %s)",
			proposed.SessionId, contractSpecUUID, proposed.Name)
	}

	// Validate the inputs against the spec
	for _, input := range proposed.Inputs {
		// Make sure there's a spec for this input
		var inputSpec *types.InputSpecification = nil
		for _, is := range recSpec.Inputs {
			if input.Name == is.Name {
				inputSpec = is
				break
			}
		}
		if inputSpec == nil {
			return fmt.Errorf("no input specification found with name %s in record specification %s",
				input.Name, recSpecID)
		}

		// Make sure the input TypeName is correct.
		if inputSpec.TypeName != input.TypeName {
			return fmt.Errorf("input %s has TypeName %s but spec calls for %s",
				input.Name, input.TypeName, inputSpec.TypeName)
		}

		// Get the input specification source type and value
		inputSpecSourceType := ""
		inputSpecSourceValue := ""
		switch source := inputSpec.Source.(type) {
		case *types.InputSpecification_RecordId:
			inputSpecSourceType = sourceTypeRecord
			inputSpecSourceValue = source.RecordId.String()
		case *types.InputSpecification_Hash:
			inputSpecSourceType = sourceTypeHash
			inputSpecSourceValue = source.Hash
		default:
			return fmt.Errorf("input spec %s has an unknown source type", input.Name)
		}

		// Get the input source type and value
		inputSourceType := ""
		inputSourceValue := ""
		switch source := input.Source.(type) {
		case *types.RecordInput_RecordId:
			inputSourceType = sourceTypeRecord
			inputSourceValue = source.RecordId.String()
		case *types.RecordInput_Hash:
			inputSourceType = sourceTypeHash
			inputSourceValue = source.Hash
		default:
			return fmt.Errorf("input %s has an unknown source type", input.Name)
		}

		// Make sure the input spec source type and value match the input source type and value
		if inputSourceType != inputSpecSourceType {
			return fmt.Errorf("input %s has %s source type but spec calls for %s",
				input.Name, inputSourceType, inputSpecSourceType)
		}
		if inputSourceValue != inputSpecSourceValue {
			return fmt.Errorf("input %s has source %s but spec calls for %s",
				input.Name, inputSourceValue, inputSpecSourceValue)
		}
	}

	// Validate the output count
	switch recSpec.ResultType {
	case types.DefinitionType_DEFINITION_TYPE_RECORD:
		if len(proposed.Outputs) != 1 {
			return fmt.Errorf("invalid output count (expected: 1, got: %d)", len(proposed.Outputs))
		}
	case types.DefinitionType_DEFINITION_TYPE_RECORD_LIST:
		if len(proposed.Outputs) == 0 {
			return fmt.Errorf("invalid output count (expected > 0, got: 0)")
		}
	}
	// case types.DefinitionType_DEFINITION_TYPE_PROPOSED: ignored
	// case types.DefinitionType_DEFINITION_TYPE_UNSPECIFIED: ignored

	// TODO: recSpec.ResponsibleParties validation?

	return nil
}

// ValidateRecordRemove checks the current record and the proposed removal scope to determine if the the proposed remove is valid
// based on the existing state
func (k Keeper) ValidateRecordRemove(ctx sdk.Context, existing types.Record, proposedID types.MetadataAddress, signers []string) error {
	scopeUUID, err := existing.SessionId.ScopeUUID()
	if err != nil {
		return fmt.Errorf("cannot get scope uuid: %s", err)
	}
	scopeID := types.ScopeMetadataAddress(scopeUUID)
	scope, found := k.GetScope(ctx, scopeID)
	if !found {
		return fmt.Errorf("unable to find scope %s", scope.ScopeId)
	}
	recordID := types.RecordMetadataAddress(scopeUUID, existing.Name)
	if !recordID.Equals(proposedID) {
		return fmt.Errorf("cannot remove record. expected %s, got %s", recordID, proposedID)
	}

	if err := k.ValidateAllPartiesAreSigners(scope.Owners, signers); err != nil {
		return err
	}

	return nil
}
