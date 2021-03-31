package keeper

import (
	"errors"
	"fmt"

	"github.com/provenance-io/provenance/x/metadata/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetSession returns the scope with the given id.
func (k Keeper) GetSession(ctx sdk.Context, id types.MetadataAddress) (session types.Session, found bool) {
	if !id.IsSessionAddress() {
		return session, false
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(id.Bytes())
	if b == nil {
		return types.Session{}, false
	}
	k.cdc.MustUnmarshalBinaryBare(b, &session)
	return session, true
}

// SetSession stores a session in the module kv store.
func (k Keeper) SetSession(ctx sdk.Context, session types.Session) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryBare(&session)
	eventType := types.EventTypeSessionCreated

	if store.Has(session.SessionId) {
		eventType = types.EventTypeSessionUpdated
	}

	store.Set(session.SessionId, b)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			eventType,
			sdk.NewAttribute(types.AttributeKeySessionID, session.SessionId.String()),
		),
	)
}

// RemoveSession removes a scope from the module kv store.
func (k Keeper) RemoveSession(ctx sdk.Context, id types.MetadataAddress) {
	if !id.IsSessionAddress() {
		panic(fmt.Errorf("invalid address, address must be for a session"))
	}
	store := ctx.KVStore(k.storeKey)

	hasRecords, err := k.hasSessionRecords(ctx, id)

	if err == nil && !hasRecords {
		store.Delete(id)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeSessionRemoved,
				sdk.NewAttribute(types.AttributeKeySessionID, id.String()),
			),
		)
	}
}

func (k Keeper) hasSessionRecords(ctx sdk.Context, id types.MetadataAddress) (bool, error) {
	if !id.IsSessionAddress() {
		return false, fmt.Errorf("invalid address, address must be for a session")
	}

	scopeUUID, err := id.ScopeUUID()
	if err != nil {
		return false, fmt.Errorf("unable to get scope uuid from session id: %s", err)
	}

	scopeAddr := types.ScopeMetadataAddress(scopeUUID)
	hasRecords := false
	err = k.IterateRecords(ctx, scopeAddr, func(r types.Record) (stop bool) {
		if r.SessionId.Equals(id) {
			hasRecords = true
			return true
		}
		return false
	})
	if err != nil {
		return false, err
	}

	return hasRecords, nil
}

// IterateSessions processes stored sessions with the given handler.
// If the scopeID is an empty MetadataAddress, all sessions will be processed.
// Otherwise, just the sessions for the given scopeID will be processed.
func (k Keeper) IterateSessions(ctx sdk.Context, scopeID types.MetadataAddress, handler func(types.Session) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	prefix, err := scopeID.ScopeSessionIteratorPrefix()
	if err != nil {
		return err
	}
	it := sdk.KVStorePrefixIterator(store, prefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var session types.Session
		err = k.cdc.UnmarshalBinaryBare(it.Value(), &session)
		if err != nil {
			k.Logger(ctx).Error("could not unmarshal session", "address", it.Key(), "error", err)
		} else if handler(session) {
			break
		}
	}
	return nil
}

// ValidateSessionUpdate checks the current session and the proposed session to determine if the the proposed changes are valid
// based on the existing state
func (k Keeper) ValidateSessionUpdate(ctx sdk.Context, existing *types.Session, proposed types.Session, signers []string) error {
	if err := proposed.ValidateBasic(); err != nil {
		return err
	}

	if existing != nil {
		if !proposed.SessionId.Equals(existing.SessionId) {
			return fmt.Errorf("cannot update session identifier. expected %s, got %s", existing.SessionId, proposed.SessionId)
		}
		if !proposed.SpecificationId.Equals(existing.SpecificationId) {
			return fmt.Errorf("cannot update specification identifier. expected %s, got %s", existing.SpecificationId, proposed.SpecificationId)
		}
		if len(proposed.GetName()) == 0 {
			return errors.New("proposed name to existing session must not be empty")
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
		return fmt.Errorf("scope not found for scope id %s", scopeID)
	}

	contractSpec, found := k.GetContractSpecification(ctx, proposed.SpecificationId)
	if !found {
		return fmt.Errorf("cannot find contract specification %s", proposed.SpecificationId)
	}

	if len(proposed.GetName()) == 0 && existing == nil {
		proposed.Name = contractSpec.ClassName
	}

	if err = k.ValidatePartiesInvolved(proposed.Parties, contractSpec.PartiesInvolved); err != nil {
		return err
	}

	if err = k.ValidateAllPartiesAreSigners(scope.Owners, signers); err != nil {
		return err
	}

	if existing != nil {
		if err = k.ValidateAuditUpdate(ctx, existing.Audit, proposed.Audit); err != nil {
			return err
		}
	}

	return nil
}

// ValidateAuditUpdate ensure that a given reference to audit fields represents no changes to
// existing audit field data.  NOTE: A nil proposed is considered "no update" and not an attempt to unset.
func (k Keeper) ValidateAuditUpdate(ctx sdk.Context, existing, proposed *types.AuditFields) error {
	if proposed == nil {
		return nil
	}
	if existing == nil {
		return errors.New("attempt to modify audit fields, modification not allowed")
	}
	if existing.CreatedBy != proposed.CreatedBy {
		return errors.New("attempt to modify created-by audit field, modification not allowed")
	}
	if existing.UpdatedBy != proposed.UpdatedBy {
		return errors.New("attempt to modify updated-by audit field, modification not allowed")
	}
	if existing.CreatedDate != proposed.CreatedDate {
		return errors.New("attempt to modify created-date audit field, modification not allowed")
	}
	if existing.UpdatedDate != proposed.UpdatedDate {
		return errors.New("attempt to modify updated-date audit field, modification not allowed")
	}
	if existing.Version != proposed.Version {
		return errors.New("attempt to modify version audit field, modification not allowed")
	}
	if existing.Message != proposed.Message {
		return errors.New("attempt to modify message audit field, modification not allowed")
	}

	return nil
}
