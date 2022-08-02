package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	markertypes "github.com/provenance-io/provenance/x/marker/types"
	"github.com/provenance-io/provenance/x/metadata/types"
)

// IterateScopes processes all stored scopes with the given handler.
func (k Keeper) IterateScopes(ctx sdk.Context, handler func(types.Scope) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	it := sdk.KVStorePrefixIterator(store, types.ScopeKeyPrefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var scope types.Scope
		k.cdc.MustUnmarshal(it.Value(), &scope)
		if handler(scope) {
			break
		}
	}
	return nil
}

// IterateScopesForAddress processes scopes associated with the provided address with the given handler.
func (k Keeper) IterateScopesForAddress(ctx sdk.Context, address sdk.AccAddress, handler func(scopeID types.MetadataAddress) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	prefix := types.GetAddressScopeCacheIteratorPrefix(address)
	it := sdk.KVStorePrefixIterator(store, prefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var scopeID types.MetadataAddress
		if err := scopeID.Unmarshal(it.Key()[len(prefix):]); err != nil {
			return err
		}
		if handler(scopeID) {
			break
		}
	}
	return nil
}

// IterateScopesForScopeSpec processes scopes associated with the provided scope specification id with the given handler.
func (k Keeper) IterateScopesForScopeSpec(ctx sdk.Context, scopeSpecID types.MetadataAddress,
	handler func(scopeID types.MetadataAddress) (stop bool),
) error {
	store := ctx.KVStore(k.storeKey)
	prefix := types.GetScopeSpecScopeCacheIteratorPrefix(scopeSpecID)
	it := sdk.KVStorePrefixIterator(store, prefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var scopeID types.MetadataAddress
		if err := scopeID.Unmarshal(it.Key()[len(prefix):]); err != nil {
			return err
		}
		if handler(scopeID) {
			break
		}
	}
	return nil
}

// GetScope returns the scope with the given id.
func (k Keeper) GetScope(ctx sdk.Context, id types.MetadataAddress) (scope types.Scope, found bool) {
	if !id.IsScopeAddress() {
		return scope, false
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(id.Bytes())
	if b == nil {
		return types.Scope{}, false
	}
	k.cdc.MustUnmarshal(b, &scope)
	return scope, true
}

// SetScope stores a scope in the module kv store.
func (k Keeper) SetScope(ctx sdk.Context, scope types.Scope) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&scope)

	var oldScope *types.Scope
	var event proto.Message = types.NewEventScopeCreated(scope.ScopeId)
	action := types.TLAction_Created
	if store.Has(scope.ScopeId) {
		event = types.NewEventScopeUpdated(scope.ScopeId)
		action = types.TLAction_Updated
		if oldScopeBytes := store.Get(scope.ScopeId); oldScopeBytes != nil {
			oldScope = &types.Scope{}
			if err := k.cdc.Unmarshal(oldScopeBytes, oldScope); err != nil {
				k.Logger(ctx).Error("could not unmarshal old scope", "err", err, "scopeId", scope.ScopeId.String(), "oldScopeBytes", oldScopeBytes)
				oldScope = nil
			}
		}
	}

	store.Set(scope.ScopeId, b)
	k.indexScope(ctx, &scope, oldScope)
	k.EmitEvent(ctx, event)
	defer types.GetIncObjFunc(types.TLType_Scope, action)
}

// RemoveScope removes a scope from the module kv store along with all its records and sessions.
func (k Keeper) RemoveScope(ctx sdk.Context, id types.MetadataAddress) {
	if !id.IsScopeAddress() {
		panic(fmt.Errorf("invalid address, address must be for a scope"))
	}
	// iterate and remove all records, groups
	store := ctx.KVStore(k.storeKey)

	scope, found := k.GetScope(ctx, id)
	if !found {
		return
	}

	// Remove all records
	prefix, err := id.ScopeRecordIteratorPrefix()
	if err != nil {
		panic(err)
	}
	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		k.RemoveRecord(ctx, iter.Key())
	}

	// Sessions will be removed as the last record in each is deleted.

	k.indexScope(ctx, nil, &scope)
	store.Delete(id)
	k.EmitEvent(ctx, types.NewEventScopeDeleted(scope.ScopeId))
	defer types.GetIncObjFunc(types.TLType_Scope, types.TLAction_Deleted)
}

// scopeIndexValues is a struct containing the values used to index a scope.
type scopeIndexValues struct {
	ScopeID         types.MetadataAddress
	Addresses       []string
	ValueOwner      string
	SpecificationID types.MetadataAddress
}

// getScopeIndexValues extracts the values used to index a scope.
func getScopeIndexValues(scope *types.Scope) *scopeIndexValues {
	if scope == nil {
		return nil
	}
	rv := scopeIndexValues{
		ScopeID:         scope.ScopeId,
		ValueOwner:      scope.ValueOwnerAddress,
		SpecificationID: scope.SpecificationId,
	}
	rv.Addresses = append(rv.Addresses, scope.DataAccess...)
	for _, p := range scope.Owners {
		rv.Addresses = appendIfNew(rv.Addresses, p.Address)
	}
	if len(scope.ValueOwnerAddress) > 0 {
		rv.Addresses = appendIfNew(rv.Addresses, scope.ValueOwnerAddress)
	}
	return &rv
}

// appendIfNew appends a string to a string slice if it's not already in the string slice.
func appendIfNew(vals []string, val string) []string {
	for _, v := range vals {
		if v == val {
			return vals
		}
	}
	return append(vals, val)
}

// getMissingScopeIndexValues extracts the index values in the required set that are not in the found set.
func getMissingScopeIndexValues(required, found *scopeIndexValues) scopeIndexValues {
	rv := scopeIndexValues{}
	if required == nil {
		return rv
	}
	if found == nil {
		return *required
	}
	rv.ScopeID = required.ScopeID
	rv.Addresses = FindMissing(required.Addresses, found.Addresses)
	if required.ValueOwner != found.ValueOwner {
		rv.ValueOwner = required.ValueOwner
	}
	if !required.SpecificationID.Equals(found.SpecificationID) {
		rv.SpecificationID = required.SpecificationID
	}
	return rv
}

// IndexKeys creates all of the index key byte arrays that this scopeIndexValues represents.
func (v scopeIndexValues) IndexKeys() [][]byte {
	rv := make([][]byte, 0)
	if v.ScopeID.Empty() {
		return rv
	}
	for _, addrStr := range v.Addresses {
		if addr, err := sdk.AccAddressFromBech32(addrStr); err == nil {
			rv = append(rv, types.GetAddressScopeCacheKey(addr, v.ScopeID))
		}
	}
	if len(v.ValueOwner) > 0 {
		if addr, err := sdk.AccAddressFromBech32(v.ValueOwner); err == nil {
			rv = append(rv, types.GetValueOwnerScopeCacheKey(addr, v.ScopeID))
		}
	}
	if !v.SpecificationID.Empty() {
		rv = append(rv, types.GetScopeSpecScopeCacheKey(v.SpecificationID, v.ScopeID))
	}
	return rv
}

// indexScope updates the index entries for a scope.
//
// When adding a new scope:  indexScope(ctx, scope, nil)
//
// When deleting a scope:  indexScope(ctx, nil, scope)
//
// When updating a scope:  indexScope(ctx, newScope, oldScope)
//
// If both newScope and oldScope are not nil, it is assumed that they have the same ScopeId.
// Failure to meet this assumption will result in strange and bad behavior.
func (k Keeper) indexScope(ctx sdk.Context, newScope, oldScope *types.Scope) {
	if newScope == nil && oldScope == nil {
		return
	}

	newScopeIndexValues := getScopeIndexValues(newScope)
	oldScopeIndexValues := getScopeIndexValues(oldScope)

	toAdd := getMissingScopeIndexValues(newScopeIndexValues, oldScopeIndexValues)
	toRemove := getMissingScopeIndexValues(oldScopeIndexValues, newScopeIndexValues)

	store := ctx.KVStore(k.storeKey)
	for _, indexKey := range toAdd.IndexKeys() {
		store.Set(indexKey, []byte{0x01})
	}
	for _, indexKey := range toRemove.IndexKeys() {
		store.Delete(indexKey)
	}
}

// ValidateScopeUpdate checks the current scope and the proposed scope to determine if the the proposed changes are valid
// based on the existing state
func (k Keeper) ValidateScopeUpdate(
	ctx sdk.Context,
	existing,
	proposed types.Scope,
	signers []string,
	msgTypeURL string,
) error {
	if err := proposed.ValidateBasic(); err != nil {
		return err
	}

	// IDs must match
	if len(existing.ScopeId) > 0 {
		if !proposed.ScopeId.Equals(existing.ScopeId) {
			return fmt.Errorf("cannot update scope identifier. expected %s, got %s", existing.ScopeId, proposed.ScopeId)
		}
	}

	if err := proposed.SpecificationId.Validate(); err != nil {
		return fmt.Errorf("invalid specification id: %w", err)
	}
	if !proposed.SpecificationId.IsScopeSpecificationAddress() {
		return fmt.Errorf("invalid specification id: is not scope specification id: %s", proposed.SpecificationId)
	}

	scopeSpec, found := k.GetScopeSpecification(ctx, proposed.SpecificationId)
	if !found {
		return fmt.Errorf("scope specification %s not found", proposed.SpecificationId)
	}
	if err := k.ValidateScopeOwners(proposed.Owners, scopeSpec); err != nil {
		return err
	}

	if len(existing.Owners) > 0 {
		// Existing owners are not required to sign when the ONLY change is from one value owner to another.
		// If the value owner wasn't previously set, and is being set now, existing owners must sign.
		// If anything else is changing, the existing owners must sign.
		proposedCopy := proposed
		if len(existing.ValueOwnerAddress) > 0 {
			proposedCopy.ValueOwnerAddress = existing.ValueOwnerAddress
		}
		if !existing.Equals(proposedCopy) {
			if err := k.ValidateAllPartiesAreSignersWithAuthz(ctx, existing.Owners, signers, msgTypeURL); err != nil {
				return err
			}
		}
	}

	if err := k.validateScopeUpdateValueOwner(ctx, existing.ValueOwnerAddress, proposed.ValueOwnerAddress, signers, msgTypeURL); err != nil {
		return err
	}

	return nil
}

// ValidateScopeRemove checks the current scope and the proposed removal scope to determine if the the proposed remove is valid
// based on the existing state
func (k Keeper) ValidateScopeRemove(ctx sdk.Context, scope types.Scope, signers []string, msgTypeURL string) error {
	if err := k.ValidateAllPartiesAreSignersWithAuthz(ctx, scope.Owners, signers, msgTypeURL); err != nil {
		return err
	}

	if err := k.validateScopeUpdateValueOwner(ctx, scope.ValueOwnerAddress, "", signers, msgTypeURL); err != nil {
		return err
	}

	return nil
}

func (k Keeper) validateScopeUpdateValueOwner(ctx sdk.Context, existing, proposed string, signers []string, msgTypeURL string) error {
	// If they're the same, we don't need to do anything.
	if existing == proposed {
		return nil
	}
	if len(existing) > 0 {
		isMarker, hasAuth := k.IsMarkerAndHasAuthority(ctx, existing, signers, markertypes.Access_Withdraw)
		if isMarker {
			// If the existing is a marker, make sure a signer has withdraw authority on it.
			if !hasAuth {
				return fmt.Errorf("missing signature for %s with authority to withdraw/remove existing value owner", existing)
			}
		} else {
			// If the existing isn't a marker, make sure they're one of the signers.
			// Not using ValidateAllOwnersAreSignersWithAuthz here because we want a slightly different error message.
			// Not using ValidateAllPartiesAreSignersWithAuthz here because of the error message and also it wants parties.
			found := false
			for _, signer := range signers {
				if existing == signer {
					found = true
					break
				}
			}
			if !found {
				stillMissing := k.checkAuthzForMissing(ctx, []string{existing}, signers, msgTypeURL)
				if len(stillMissing) == 0 {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("missing signature from existing value owner %s", existing)
			}
		}
	}
	if len(proposed) > 0 {
		isMarker, hasAuth := k.IsMarkerAndHasAuthority(ctx, proposed, signers, markertypes.Access_Deposit)
		// If the proposed is a marker, make sure a signer has deposit authority on it.
		if isMarker && !hasAuth {
			return fmt.Errorf("no signatures present with authority to add scope to marker %s", proposed)
		}
		// If it's not a marker, we don't really care what it's being set to.
	}
	return nil
}

// ValidateScopeAddDataAccess checks the current scope and the proposed
func (k Keeper) ValidateScopeAddDataAccess(
	ctx sdk.Context,
	dataAccessAddrs []string,
	existing types.Scope,
	signers []string,
	msgTypeURL string,
) error {
	if len(dataAccessAddrs) < 1 {
		return fmt.Errorf("data access list cannot be empty")
	}

	for _, da := range dataAccessAddrs {
		_, err := sdk.AccAddressFromBech32(da)
		if err != nil {
			return fmt.Errorf("failed to decode data access address %s : %v", da, err.Error())
		}
		for _, pda := range existing.DataAccess {
			if da == pda {
				return fmt.Errorf("address already exists for data access %s", pda)
			}
		}
	}

	if err := k.ValidateAllPartiesAreSignersWithAuthz(ctx, existing.Owners, signers, msgTypeURL); err != nil {
		return err
	}

	return nil
}

// ValidateScopeDeleteDataAccess checks the current scope data access and the proposed removed items
func (k Keeper) ValidateScopeDeleteDataAccess(
	ctx sdk.Context,
	dataAccessAddrs []string,
	existing types.Scope,
	signers []string,
	msgTypeURL string,
) error {
	if len(dataAccessAddrs) < 1 {
		return fmt.Errorf("data access list cannot be empty")
	}
	for _, da := range dataAccessAddrs {
		_, err := sdk.AccAddressFromBech32(da)
		if err != nil {
			return fmt.Errorf("failed to decode data access address %s : %v", da, err.Error())
		}
		found := false
		for _, pda := range existing.DataAccess {
			if da == pda {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("address does not exist in scope data access: %s", da)
		}
	}

	if err := k.ValidateAllPartiesAreSignersWithAuthz(ctx, existing.Owners, signers, msgTypeURL); err != nil {
		return err
	}

	return nil
}

// ValidateScopeUpdateOwners checks the current scopes owners and the proposed update
func (k Keeper) ValidateScopeUpdateOwners(
	ctx sdk.Context,
	existing,
	proposed types.Scope,
	signers []string,
	msgTypeURL string,
) error {
	if err := proposed.ValidateOwnersBasic(); err != nil {
		return err
	}

	scopeSpec, found := k.GetScopeSpecification(ctx, proposed.SpecificationId)
	if !found {
		return fmt.Errorf("scope specification %s not found", proposed.SpecificationId)
	}
	if err := k.ValidateScopeOwners(proposed.Owners, scopeSpec); err != nil {
		return err
	}
	if err := k.ValidateAllPartiesAreSignersWithAuthz(ctx, existing.Owners, signers, msgTypeURL); err != nil {
		return err
	}
	return nil
}

// ValidateScopeOwners is stateful validation for scope owners against a scope specification.
// This does NOT involve the Scope.ValidateOwnersBasic() function.
func (k Keeper) ValidateScopeOwners(owners []types.Party, spec types.ScopeSpecification) error {
	var missingPartyTypes []string
	for _, pt := range spec.PartiesInvolved {
		found := false
		for _, o := range owners {
			if o.Role == pt {
				found = true
				break
			}
		}
		if !found {
			// Get the party type without the "PARTY_TYPE_" prefix.
			missingPartyTypes = append(missingPartyTypes, pt.String()[11:])
		}
	}
	if len(missingPartyTypes) > 0 {
		return fmt.Errorf("missing party type%s required by spec: %v", pluralEnding(len(missingPartyTypes)), missingPartyTypes)
	}
	return nil
}
