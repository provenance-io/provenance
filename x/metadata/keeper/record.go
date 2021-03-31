package keeper

import (
	"fmt"
	"strings"

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
	var iterator func(r types.Record) (stop bool)
	if len(name) == 0 {
		iterator = func(r types.Record) (stop bool) {
			records = append(records, &r)
			return false
		}
	} else {
		name = strings.ToLower(name)
		iterator = func(r types.Record) (stop bool) {
			if name == strings.ToLower(r.Name) {
				records = append(records, &r)
				return true
			}
			return false
		}
	}
	err := k.IterateRecords(ctx, scopeAddress, iterator)
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
		err = k.cdc.UnmarshalBinaryBare(it.Value(), &record)
		if err != nil {
			k.Logger(ctx).Error("could not unmarshal record", "address", it.Key(), "error", err)
		} else if handler(record) {
			break
		}
	}
	return nil
}

// ValidateRecordUpdate checks the current record and the proposed record to determine if the the proposed changes are valid
// based on the existing state
func (k Keeper) ValidateRecordUpdate(ctx sdk.Context, existing *types.Record, proposed types.Record, signers []string, partiesInvolved []types.Party) error {
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
		// The existing specification id might be empty for old stuff.
		// And for now, we'll allow the proposed specification id to be missing and set it appropriately below.
		// But if we've got both, make sure they didn't change.
		if !existing.SpecificationId.Empty() && !proposed.SpecificationId.Empty() && !existing.SpecificationId.Equals(proposed.SpecificationId) {
			return fmt.Errorf("the SpecificationId of records cannot be changed")
		}
	}

	scopeID, err := proposed.SessionId.AsScopeAddress()
	if err != nil {
		return err
	}

	// get scope for existing record
	scope, found := k.GetScope(ctx, scopeID)
	if !found {
		return fmt.Errorf("scope not found for scope id %s", scopeID)
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

	// Get contract specification
	contractSpec, found := k.GetContractSpecification(ctx, session.SpecificationId)
	if !found {
		return fmt.Errorf("contract specification not found: %s", session.SpecificationId.String())
	}
	err = k.ValidatePartiesInvolved(partiesInvolved, contractSpec.PartiesInvolved)
	if err != nil {
		return err
	}
	// Get the record specification
	contractSpecUUID, err := session.SpecificationId.ContractSpecUUID()
	if err != nil {
		return err
	}
	recSpecID := types.RecordSpecMetadataAddress(contractSpecUUID, proposed.Name)

	if proposed.SpecificationId.Empty() {
		proposed.SpecificationId = recSpecID
	} else if !proposed.SpecificationId.Equals(recSpecID) {
		return fmt.Errorf("proposed specification id %s does not match expected specification id %s",
			proposed.SpecificationId, recSpecID)
	}

	recSpec, found := k.GetRecordSpecification(ctx, recSpecID)
	if !found {
		return fmt.Errorf("record specification not found for record specification id %s (contract spec uuid %s and record name %s)",
			recSpecID, contractSpecUUID, proposed.Name)
	}

	// Make sure all input specs are present as inputs.
	inputNames := make([]string, len(proposed.Inputs))
	inputMap := make(map[string]types.RecordInput)
	for i, input := range proposed.Inputs {
		if _, found := inputMap[input.Name]; found {
			return fmt.Errorf("input name %s provided twice", input.Name)
		}
		inputNames[i] = input.Name
		inputMap[input.Name] = input
	}
	inputSpecNames := make([]string, len(recSpec.Inputs))
	inputSpecMap := make(map[string]types.InputSpecification)
	for i, inputSpec := range recSpec.Inputs {
		inputSpecNames[i] = inputSpec.Name
		inputSpecMap[inputSpec.Name] = *inputSpec
	}
	missingInputNames := FindMissing(inputSpecNames, inputNames)
	if len(missingInputNames) == 1 {
		return fmt.Errorf("missing input %s", missingInputNames[0])
	} else if len(missingInputNames) > 1 {
		return fmt.Errorf("missing inputs %v", missingInputNames)
	}
	extraInputNames := FindMissing(inputNames, inputSpecNames)
	if len(extraInputNames) == 1 {
		return fmt.Errorf("extra input %s", extraInputNames[0])
	} else if len(extraInputNames) > 1 {
		return fmt.Errorf("extra inputs %v", extraInputNames)
	}

	// Make sure all the inputs conform to their spec.
	for _, name := range inputNames {
		input := inputMap[name]
		inputSpec := inputSpecMap[name]

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
			return fmt.Errorf("input spec %s has an unknown source type", inputSpec.Name)
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
			return fmt.Errorf("input %s has source type %s but spec calls for %s",
				input.Name, inputSourceType, inputSpecSourceType)
		}
		if inputSourceValue != inputSpecSourceValue {
			return fmt.Errorf("input %s has source value %s but spec calls for %s",
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
