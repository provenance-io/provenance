package keeper

import (
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
		panic(fmt.Errorf("invalid address, address must be for a record session"))
	}
	store := ctx.KVStore(k.storeKey)
	store.Delete(id)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSessionRemoved,
			sdk.NewAttribute(types.AttributeKeySessionID, id.String()),
		),
	)
}

// IterateSessions processes all stored scopes with the given handler.
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
		k.cdc.MustUnmarshalBinaryBare(it.Value(), &session)
		if handler(session) {
			break
		}
	}
	return nil
}
