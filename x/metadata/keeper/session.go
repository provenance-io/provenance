package keeper

import (
	"errors"
	"fmt"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// GetSession returns the session with the given id.
func (k Keeper) GetSession(ctx sdk.Context, id types.MetadataAddress) (session types.Session, found bool) {
	if !id.IsSessionAddress() {
		return session, false
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(id.Bytes())
	if b == nil {
		return types.Session{}, false
	}
	k.cdc.MustUnmarshal(b, &session)
	return session, true
}

// SetSession stores a session in the module kv store.
func (k Keeper) SetSession(ctx sdk.Context, session types.Session) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&session)

	var event proto.Message = types.NewEventSessionCreated(session.SessionId)
	if store.Has(session.SessionId) {
		event = types.NewEventSessionUpdated(session.SessionId)
	}

	store.Set(session.SessionId, b)
	k.EmitEvent(ctx, event)
}

// RemoveSession removes a session from the module kv store if there are no records associated with it.
func (k Keeper) RemoveSession(ctx sdk.Context, id types.MetadataAddress) {
	if !id.IsSessionAddress() {
		panic(fmt.Errorf("invalid address, address must be for a session"))
	}
	store := ctx.KVStore(k.storeKey)

	if !store.Has(id) || k.sessionHasRecords(ctx, id) {
		return
	}

	store.Delete(id)
	k.EmitEvent(ctx, types.NewEventSessionDeleted(id))
}

func (k Keeper) sessionHasRecords(ctx sdk.Context, id types.MetadataAddress) bool {
	if !id.IsSessionAddress() {
		return false
	}

	scopeAddr, err := id.AsScopeAddress()
	if err != nil {
		return false
	}

	hasRecords := false
	_ = k.IterateRecords(ctx, scopeAddr, func(r types.Record) (stop bool) {
		if r.SessionId.Equals(id) {
			hasRecords = true
			return true
		}
		return false
	})

	return hasRecords
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
	it := storetypes.KVStorePrefixIterator(store, prefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var session types.Session
		err = k.cdc.Unmarshal(it.Value(), &session)
		if err != nil {
			k.Logger(ctx).Error("could not unmarshal session", "address", it.Key(), "error", err)
		} else if handler(session) {
			break
		}
	}
	return nil
}

// ValidateWriteSession checks the current session and the proposed session to determine if the proposed changes are valid
// based on the existing state
func (k Keeper) ValidateWriteSession(ctx sdk.Context, existing *types.Session, msg *types.MsgWriteSessionRequest) error {
	proposed := msg.Session
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
	if err = types.ValidateOptionalParties(scope.RequirePartyRollup, proposed.Parties); err != nil {
		return err
	}

	contractSpec, found := k.GetContractSpecification(ctx, proposed.SpecificationId)
	if !found {
		return fmt.Errorf("cannot find contract specification %s", proposed.SpecificationId)
	}

	scopeSpec, found := k.GetScopeSpecification(ctx, scope.SpecificationId)
	if !found {
		return fmt.Errorf("scope spec not found with id %s", scope.SpecificationId)
	}
	scopeSpecHasContractSpec := false
	for _, cSpecID := range scopeSpec.ContractSpecIds {
		if cSpecID.Equals(proposed.SpecificationId) {
			scopeSpecHasContractSpec = true
			break
		}
	}
	if !scopeSpecHasContractSpec {
		return fmt.Errorf("contract spec %s not listed in scope spec %s", proposed.SpecificationId, scopeSpec.SpecificationId)
	}

	// Make sure everyone has signed.
	if !scope.RequirePartyRollup {
		// Old:
		//   - All roles required by the contract spec must have a party in the session parties.
		//   - All scope owners must sign.
		if err = validateRolesPresent(proposed.Parties, contractSpec.PartiesInvolved); err != nil {
			return err
		}
		if err = k.validateProvenanceRole(ctx, types.BuildPartyDetails(nil, proposed.Parties)); err != nil {
			return err
		}
		if err = k.ValidateSignersWithoutParties(ctx, scope.GetAllOwnerAddresses(), msg); err != nil {
			return err
		}
	} else {
		// New:
		//   - All proposed session parties must be present in this scope's owners.
		//   - All optional=false scope owners must be signers.
		//   - If new, all roles required by the contract spec must have a signer and
		//     associated party in the proposed session.
		//   - If not new, all roles required by the contract spec must have a signer
		//     and associated party in the existing session.
		//   - If not new, all optional=false existing parties must also be signers.
		if err = validatePartiesArePresent(proposed.Parties, scope.Owners); err != nil {
			return fmt.Errorf("not all session parties in scope owners: %w", err)
		}
		var availableParties []types.Party
		var reqParties []types.Party
		if existing != nil {
			availableParties = existing.Parties
			// Providing the existing parties to ValidateSignersWithParties, so we need to check some stuff on proposed.
			if err = validateRolesPresent(proposed.Parties, contractSpec.PartiesInvolved); err != nil {
				return err
			}
			if err = k.validateProvenanceRole(ctx, types.BuildPartyDetails(nil, proposed.Parties)); err != nil {
				return err
			}
			reqParties = append(reqParties, existing.Parties...)
		} else {
			availableParties = proposed.Parties
			// We don't call validateRolesPresent or validateProvenanceRole here because proposed.Parties is being
			// provided to ValidateSignersWithParties, which does those.
		}
		reqParties = append(reqParties, scope.Owners...)
		if err = k.ValidateSignersWithParties(ctx, reqParties, availableParties, contractSpec.PartiesInvolved, msg); err != nil {
			return err
		}
	}

	if existing != nil {
		if err = k.ValidateAuditUpdate(ctx, existing.Audit, proposed.Audit); err != nil {
			return err
		}
	}

	if len(proposed.GetName()) == 0 && existing == nil {
		proposed.Name = contractSpec.ClassName
	}

	return nil
}

// ValidateAuditUpdate ensure that a given reference to audit fields represents no changes to
// existing audit field data.  NOTE: A nil proposed is considered "no update" and not an attempt to unset.
func (k Keeper) ValidateAuditUpdate(_ sdk.Context, existing, proposed *types.AuditFields) error {
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
