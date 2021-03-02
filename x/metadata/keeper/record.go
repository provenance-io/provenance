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

	recordID := record.SessionId.GetRecordAddress(record.Name)
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
	// get scope, collect required signers, get session (if it exists, if it is a new one make sure the contract-spec is allowed if restricted via scope spec), collect signers from that contract spec… verify update is correctly signed… pull record specification, check against the record update (this is a name match lookup against record name)

	if err := proposed.ValidateBasic(); err != nil {
		return err
	}

	scopeUUID, err := existing.SessionId.ScopeUUID()
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
	requiredSignatures := []string{}
	for _, p := range scope.Owners {
		requiredSignatures = append(requiredSignatures, p.Address)
	}

	// Signatures required of all existing data owners.
	for _, owner := range requiredSignatures {
		found := false
		for _, signer := range signers {
			if owner == signer {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("missing signature from existing owner %s; required for update", owner)
		}
	}

	// TODO finish full validation of update once specs are complete

	return nil
}
