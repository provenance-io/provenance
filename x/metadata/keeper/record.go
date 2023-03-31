package keeper

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/metadata/types"
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
	k.cdc.MustUnmarshal(b, &record)
	return record, true
}

// GetRecords returns records for a scope optionally limited to a name.
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
	b := k.cdc.MustMarshal(&record)

	recordID := record.SessionId.MustGetAsRecordAddress(record.Name)

	var event proto.Message = types.NewEventRecordCreated(recordID, record.SessionId)
	action := types.TLAction_Created
	if store.Has(recordID) {
		event = types.NewEventRecordUpdated(recordID, record.SessionId)
		action = types.TLAction_Updated
	}

	store.Set(recordID, b)
	k.EmitEvent(ctx, event)
	defer types.GetIncObjFunc(types.TLType_Record, action)
}

// RemoveRecord removes a record from the module kv store.
func (k Keeper) RemoveRecord(ctx sdk.Context, id types.MetadataAddress) {
	if !id.IsRecordAddress() {
		panic(fmt.Errorf("invalid address, address must be for a record"))
	}
	record, found := k.GetRecord(ctx, id)
	if !found {
		return
	}
	store := ctx.KVStore(k.storeKey)
	store.Delete(id)
	k.EmitEvent(ctx, types.NewEventRecordDeleted(id))
	defer types.GetIncObjFunc(types.TLType_Record, types.TLAction_Deleted)

	// Remove the session too if there are no more records in it.
	k.RemoveSession(ctx, record.SessionId)
}

// IterateRecords processes stored records with the given handler.
// If the scopeID is an empty MetadataAddress, all records will be processed.
// Otherwise, just the records for the given scopeID will be processed.
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
		err = k.cdc.Unmarshal(it.Value(), &record)
		if err != nil {
			k.Logger(ctx).Error("could not unmarshal record", "address", it.Key(), "error", err)
		} else if handler(record) {
			break
		}
	}
	return nil
}

// ValidateWriteRecord checks the current record and the proposed record to determine if the proposed changes are valid
// based on the existing state
// Note: The proposed parameter is a reference here so that the SpecificationId can be set in cases when it's not provided.
func (k Keeper) ValidateWriteRecord(
	ctx sdk.Context,
	existing *types.Record,
	msg *types.MsgWriteRecordRequest,
) error {
	proposed := &msg.Record
	if err := proposed.ValidateBasic(); err != nil {
		return err
	}

	var oldSession *types.Session
	if existing != nil {
		if existing.Name != proposed.Name {
			return fmt.Errorf("the Name field of records cannot be changed")
		}
		if !existing.SessionId.Equals(proposed.SessionId) {
			// If the session is changing, add the original session's parties to the required parties list.
			// If it's somehow not found, just ignore that and let the record be updated as normal.
			os, found := k.GetSession(ctx, existing.SessionId)
			if found {
				oldSession = &os
			}
		}
		// The existing specification id might be empty for old stuff.
		// And for now, we'll allow the proposed specification id to be missing and set it appropriately below.
		// But if we've got both, make sure they didn't change.
		if !existing.SpecificationId.Empty() && !proposed.SpecificationId.Empty() && !existing.SpecificationId.Equals(proposed.SpecificationId) {
			return fmt.Errorf("the SpecificationId of records cannot be changed")
		}
	}

	scopeID, scopeIDErr := proposed.SessionId.AsScopeAddress()
	if scopeIDErr != nil {
		return scopeIDErr
	}

	// Get the scope, session, and record spec.
	scope, found := k.GetScope(ctx, scopeID)
	if !found {
		return fmt.Errorf("scope not found with id %s", scopeID)
	}
	session, found := k.GetSession(ctx, proposed.SessionId)
	if !found {
		return fmt.Errorf("session not found for session id %s", proposed.SessionId)
	}
	recSpecID, err := session.SpecificationId.AsRecordSpecAddress(proposed.Name)
	if err != nil {
		return fmt.Errorf("could not create record specification id from contract spec id %s and record name %q",
			session.SpecificationId, proposed.Name)
	}
	if proposed.SpecificationId.Empty() {
		proposed.SpecificationId = recSpecID
	} else if !proposed.SpecificationId.Equals(recSpecID) {
		return fmt.Errorf("proposed specification id %s does not match expected specification id %s",
			proposed.SpecificationId, recSpecID)
	}
	recSpec, found := k.GetRecordSpecification(ctx, recSpecID)
	if !found {
		return fmt.Errorf("record specification not found for record specification id %s (contract spec id %s and record name %q)",
			recSpecID, session.SpecificationId, proposed.Name)
	}

	// Make sure everyone has signed.
	if !scope.GetRequirePartyRollup() {
		// Old:
		//   - All roles required by the record spec must have a party in the session parties.
		//   - All session parties must sign.
		//   - If the record is being updated to a new session, all previous session parties must sign.
		if err = validateRolesPresent(session.Parties, recSpec.ResponsibleParties); err != nil {
			return err
		}
		reqSigs := session.GetAllPartyAddresses()
		if oldSession != nil {
			reqSigs = append(reqSigs, oldSession.GetAllPartyAddresses()...)
		}
		if _, err = k.ValidateSignersWithoutParties(ctx, reqSigs, msg); err != nil {
			return err
		}
	} else {
		// New:
		//   - All roles required by the record spec must have a signer and associated party in the session.
		//   - All optional=false parties in the scope owners and session parties must be signers.
		//   - If the record is changing sessions, all optional=false parties in the previous session must be signers.
		var reqParties []types.Party
		reqParties = append(reqParties, scope.Owners...)
		reqParties = append(reqParties, session.Parties...)
		if oldSession != nil {
			reqParties = append(reqParties, oldSession.Parties...)
		}
		if _, err = k.ValidateSignersWithParties(ctx, reqParties, session.Parties, recSpec.ResponsibleParties, msg); err != nil {
			return err
		}
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
	missingInputNames := findMissing(inputSpecNames, inputNames)
	if len(missingInputNames) > 0 {
		return fmt.Errorf("missing input%s %v", pluralEnding(len(missingInputNames)), missingInputNames)
	}
	extraInputNames := findMissing(inputNames, inputSpecNames)
	if len(extraInputNames) > 0 {
		return fmt.Errorf("extra input%s %v", pluralEnding(len(extraInputNames)), extraInputNames)
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
			if _, found := k.GetRecord(ctx, source.RecordId); !found {
				return fmt.Errorf("input %s source record id %s not found", input.Name, source.RecordId)
			}
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
		if inputSourceType == sourceTypeRecord && inputSourceValue != inputSpecSourceValue {
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

// ValidateDeleteRecord checks the current record and the proposed removal scope to determine if the proposed remove is valid
// based on the existing state
func (k Keeper) ValidateDeleteRecord(ctx sdk.Context, proposedID types.MetadataAddress, msg types.MetadataMsg) error {
	record, found := k.GetRecord(ctx, proposedID)
	if !found {
		return fmt.Errorf("record does not exist to delete: %s", proposedID)
	}

	// GetRecord found a record, so we know proposedID is good and will have a scope id.
	// That's why we can ignore the error from AsScopeAddress().
	scopeID, _ := proposedID.AsScopeAddress()
	var scope *types.Scope
	if s, found := k.GetScope(ctx, scopeID); found {
		scope = &s
	}
	var session *types.Session
	if s, found := k.GetSession(ctx, record.SessionId); found {
		session = &s
	}

	// Make sure everyone has signed.
	switch {
	case scope == nil && session == nil:
		// nothing to validate signers against. Just let the record get deleted too.
	case scope != nil && !scope.GetRequirePartyRollup():
		// Old:
		//   - All roles required by the record spec must have a party in the session parties.
		//   - All session parties must sign.
		//   - If the record is being updated to a new session, all previous session parties must sign.
		// Since we're deleting it, the only one that makes sense to do, is check the session party signers.
		// And if the session doesn't exist, just let the record get deleted.
		if session != nil {
			if _, err := k.ValidateSignersWithoutParties(ctx, session.GetAllPartyAddresses(), msg); err != nil {
				return err
			}
		}
	default:
		// New:
		//   - All roles required by the record spec must have a signer and associated party in the session.
		//   - All optional=false parties in the scope owners and session parties must be signers.
		//   - If the record is changing sessions, all optional=false parties in the previous session must be signers.
		// We don't need to worry about that last part.

		// Required parties come from both the session and scope since they might have different optional values.
		var reqParties []types.Party
		// available parties come from the session if it exists, otherwise the scope.
		var availableParties []types.Party
		if scope != nil {
			reqParties = append(reqParties, scope.Owners...)
		}
		if session != nil {
			reqParties = append(reqParties, session.Parties...)
			availableParties = append(availableParties, session.Parties...)
		} else {
			availableParties = append(availableParties, scope.Owners...)
		}

		// If the record spec doesn't exist, ignore the role/signer requirement.
		reqSpec, found := k.GetRecordSpecification(ctx, record.SpecificationId)
		if !found {
			if _, err := k.ValidateSignersWithoutParties(ctx, types.GetRequiredPartyAddresses(reqParties), msg); err != nil {
				return err
			}
		} else {
			if _, err := k.ValidateSignersWithParties(ctx, reqParties, availableParties, reqSpec.ResponsibleParties, msg); err != nil {
				return err
			}
		}
	}

	return nil
}
