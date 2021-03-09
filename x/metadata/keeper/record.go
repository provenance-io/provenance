package keeper

import (
	"fmt"

	"github.com/provenance-io/provenance/x/metadata/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

	recordID := record.SessionId.AsRecordAddress(record.Name)
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
func (k Keeper) ValidateRecordUpdate(ctx sdk.Context, existing, proposed types.Record, signers []string) error {
	if len(existing.SessionId) > 0 {
		if !existing.SessionId.Equals(proposed.SessionId) || existing.Name != proposed.Name {
			return fmt.Errorf("existing and proposed records do not match")
		}
	}

	if err := proposed.ValidateBasic(); err != nil {
		return err
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

	// Validate any changes to the ValueOwner property.
	if err := k.ValidateRequiredSignatures(scope.Owners, signers); err != nil {
		return err
	}

	// TODO finish full validation of update once specs are complete

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

	if err := k.ValidateRequiredSignatures(scope.Owners, signers); err != nil {
		return err
	}

	return nil
}
