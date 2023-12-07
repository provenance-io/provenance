package keeper

import (
	"fmt"

	"github.com/gogo/protobuf/proto"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// IterateScopes processes all stored scopes with the given handler.
func (k Keeper) IterateScopes(ctx sdk.Context, handler func(types.Scope) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	it := storetypes.KVStorePrefixIterator(store, types.ScopeKeyPrefix)
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
	it := storetypes.KVStorePrefixIterator(store, prefix)
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

// IterateScopesForValueOwner iterates over all scope ids that have the provided value owner.
func (k Keeper) IterateScopesForValueOwner(ctx sdk.Context, valueOwner string, handler func(scopeID types.MetadataAddress) (stop bool)) error {
	addr, err := sdk.AccAddressFromBech32(valueOwner)
	if err != nil {
		return fmt.Errorf("cannot iterate over invalid value owner %q: %w", valueOwner, err)
	}

	store := ctx.KVStore(k.storeKey)
	prefix := types.GetValueOwnerScopeCacheIteratorPrefix(addr)
	it := storetypes.KVStorePrefixIterator(store, prefix)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var scopeID types.MetadataAddress
		if err = scopeID.Unmarshal(it.Key()[len(prefix):]); err != nil {
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
	it := storetypes.KVStorePrefixIterator(store, prefix)
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
	k.indexScope(store, &scope, oldScope)
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
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		k.RemoveRecord(ctx, iter.Key())
	}

	// Sessions will be removed as the last record in each is deleted.

	k.indexScope(store, nil, &scope)
	store.Delete(id)
	k.EmitEvent(ctx, types.NewEventScopeDeleted(scope.ScopeId))
	defer types.GetIncObjFunc(types.TLType_Scope, types.TLAction_Deleted)
}

// SetScopeValueOwners updates the value owner of all the provided scopes and stores each in the kv store.
//
// Contract: Each provided scope must not have been modified from its value as read from state.
// Changing one before providing it to this function can mess up indexing.
func (k Keeper) SetScopeValueOwners(ctx sdk.Context, scopes []*types.Scope, newValueOwner string) {
	// Not using SetScope in here to skip the re-reading of each scope.
	// It's expected that sometimes there will be quite a few (100+) scopes to update, so this will save on gas.
	store := ctx.KVStore(k.storeKey)

	for _, oldScope := range scopes {
		// Copy the old scope and update/store the copy.
		newScope := *oldScope
		newScope.ValueOwnerAddress = newValueOwner
		b := k.cdc.MustMarshal(&newScope)
		store.Set(newScope.ScopeId, b)
		k.indexScope(store, &newScope, oldScope)
		k.EmitEvent(ctx, types.NewEventScopeUpdated(oldScope.ScopeId))
	}
	types.GetIncObjFuncN(types.TLType_Scope, types.TLAction_Updated, len(scopes))()
}

// scopeIndexValues is a struct containing the values used to index a scope.
type scopeIndexValues struct {
	ScopeID         types.MetadataAddress
	Addresses       []sdk.AccAddress
	ValueOwner      sdk.AccAddress
	SpecificationID types.MetadataAddress
}

// getScopeIndexValues extracts the values used to index a scope.
func getScopeIndexValues(scope *types.Scope) *scopeIndexValues {
	if scope == nil {
		return nil
	}
	rv := scopeIndexValues{
		ScopeID:         scope.ScopeId,
		SpecificationID: scope.SpecificationId,
	}
	knownAddrs := make(map[string]bool)
	if addr, err := sdk.AccAddressFromBech32(scope.ValueOwnerAddress); err == nil {
		rv.ValueOwner = addr
		rv.Addresses = append(rv.Addresses, addr)
		knownAddrs[scope.ValueOwnerAddress] = true
	}
	for _, dataAccess := range scope.DataAccess {
		if !knownAddrs[dataAccess] {
			if addr, err := sdk.AccAddressFromBech32(dataAccess); err == nil {
				rv.Addresses = append(rv.Addresses, addr)
			}
			knownAddrs[dataAccess] = true
		}
	}
	for _, owner := range scope.Owners {
		if !knownAddrs[owner.Address] {
			if addr, err := sdk.AccAddressFromBech32(owner.Address); err == nil {
				rv.Addresses = append(rv.Addresses, addr)
			}
			knownAddrs[owner.Address] = true
		}
	}
	return &rv
}

// getMissingScopeIndexValues extracts the index values in the required set that are not in the found set.
func getMissingScopeIndexValues(required, found *scopeIndexValues) *scopeIndexValues {
	rv := &scopeIndexValues{}
	if required == nil {
		return rv
	}
	if found == nil {
		return required
	}
	rv.ScopeID = required.ScopeID
	rv.Addresses = findMissingComp(required.Addresses, found.Addresses, func(a1 sdk.AccAddress, a2 sdk.AccAddress) bool {
		return a1.Equals(a2)
	})
	if !required.ValueOwner.Equals(found.ValueOwner) {
		rv.ValueOwner = required.ValueOwner
	}
	if !required.SpecificationID.Equals(found.SpecificationID) {
		rv.SpecificationID = required.SpecificationID
	}
	return rv
}

// IndexKeys creates all of the index key byte arrays that this scopeIndexValues represents.
func (v scopeIndexValues) IndexKeys() [][]byte {
	if v.ScopeID.Empty() {
		return nil
	}
	rv := make([][]byte, 0, len(v.Addresses)+2)
	for _, addr := range v.Addresses {
		rv = append(rv, types.GetAddressScopeCacheKey(addr, v.ScopeID))
	}
	if len(v.ValueOwner) > 0 {
		rv = append(rv, types.GetValueOwnerScopeCacheKey(v.ValueOwner, v.ScopeID))
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
func (k Keeper) indexScope(store storetypes.KVStore, newScope, oldScope *types.Scope) {
	if newScope == nil && oldScope == nil {
		return
	}

	newScopeIndexValues := getScopeIndexValues(newScope)
	oldScopeIndexValues := getScopeIndexValues(oldScope)

	toAdd := getMissingScopeIndexValues(newScopeIndexValues, oldScopeIndexValues)
	toRemove := getMissingScopeIndexValues(oldScopeIndexValues, newScopeIndexValues)

	for _, indexKey := range toAdd.IndexKeys() {
		store.Set(indexKey, []byte{0x01})
	}
	for _, indexKey := range toRemove.IndexKeys() {
		store.Delete(indexKey)
	}
}

// ValidateWriteScope checks the current scope and the proposed scope to determine if the proposed changes are valid
// based on the existing state
func (k Keeper) ValidateWriteScope(
	ctx sdk.Context,
	existing *types.Scope,
	msg *types.MsgWriteScopeRequest,
) error {
	proposed := msg.Scope
	if err := proposed.ValidateBasic(); err != nil {
		return err
	}

	// IDs must match
	if existing != nil {
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

	// Existing owners are not required to sign when the ONLY change is from one value owner to another.
	// If the value owner wasn't previously set, and is being set now, existing owners must sign.
	// If anything else is changing, the existing owners must sign.
	existingValueOwner := ""
	onlyChangeIsValueOwner := false
	if existing != nil && len(existing.ValueOwnerAddress) > 0 {
		existingValueOwner = existing.ValueOwnerAddress
		// Make a copy of proposed scope and set its value owner to the existing one. If it then
		// equals the existing scope, then the only change in proposed is to the value owner field.
		proposedCopy := proposed
		proposedCopy.ValueOwnerAddress = existing.ValueOwnerAddress
		onlyChangeIsValueOwner = existing.Equals(proposedCopy)
	}

	var err error
	var validatedParties []*PartyDetails

	if !onlyChangeIsValueOwner {
		scopeSpec, found := k.GetScopeSpecification(ctx, proposed.SpecificationId)
		if !found {
			return fmt.Errorf("scope specification %s not found", proposed.SpecificationId)
		}

		if err = validateRolesPresent(proposed.Owners, scopeSpec.PartiesInvolved); err != nil {
			return err
		}
		if err = k.validateProvenanceRole(ctx, BuildPartyDetails(nil, proposed.Owners)); err != nil {
			return err
		}

		// Make sure everyone has signed.
		if (existing != nil && !existing.RequirePartyRollup) || (existing == nil && !proposed.RequirePartyRollup) {
			// Old:
			//   - All roles required by the scope spec must have a party in the owners.
			//   - If not new, all existing owners must sign.
			//   - Value owner signer restrictions are applied.
			if existing != nil && !existing.Equals(proposed) {
				if validatedParties, err = k.validateAllRequiredSigned(ctx, existing.GetAllOwnerAddresses(), msg); err != nil {
					return err
				}
			}
		} else {
			// New:
			//   - All roles required by the scope spec must have a party in the owners.
			//   - If not new, all required=false existing owners must be signers.
			//   - If not new, all roles required by the scope spec must have a signer and
			//     associated party from the existing scope.
			//   - Value owner signer restrictions are applied.
			// Note: This means that a scope can be initially written without consideration for signers and roles.
			if existing != nil {
				if validatedParties, err = k.validateAllRequiredPartiesSigned(ctx, existing.Owners, existing.Owners, scopeSpec.PartiesInvolved, msg); err != nil {
					return err
				}
			}
		}
	}

	usedSigners, err := k.ValidateScopeValueOwnerUpdate(ctx, existingValueOwner, proposed.ValueOwnerAddress, msg)
	if err != nil {
		return err
	}

	usedSigners.AlsoUse(GetUsedSigners(validatedParties))
	return k.validateSmartContractSigners(ctx, usedSigners, msg)
}

// ValidateDeleteScope checks the current scope and the proposed removal scope to determine if the proposed remove is valid
// based on the existing state
func (k Keeper) ValidateDeleteScope(ctx sdk.Context, msg *types.MsgDeleteScopeRequest) error {
	scope, found := k.GetScope(ctx, msg.ScopeId)
	if !found {
		return fmt.Errorf("scope not found with id %s", msg.ScopeId)
	}

	var err error
	var validatedParties []*PartyDetails

	// Make sure everyone has signed.
	if !scope.RequirePartyRollup {
		// Old:
		//   - All roles required by the scope spec must have a party in the owners.
		//   - If not new, all existing owners must sign.
		//   - Value owner signer restrictions are applied.
		// We don't care about the first one here.
		if validatedParties, err = k.validateAllRequiredSigned(ctx, scope.GetAllOwnerAddresses(), msg); err != nil {
			return err
		}
	} else {
		// New:
		//   - All roles required by the scope spec must have a party in the owners.
		//   - If not new, all required=false existing owners must be signers.
		//   - If not new, all roles required by the scope spec must have a signer and
		//     associated party from the existing scope.
		//   - Value owner signer restrictions are applied.
		// We don't care about that first one, and only care about the roles one if the spec exists.
		scopeSpec, found := k.GetScopeSpecification(ctx, scope.SpecificationId)
		if !found {
			if validatedParties, err = k.validateAllRequiredSigned(ctx, types.GetRequiredPartyAddresses(scope.Owners), msg); err != nil {
				return err
			}
		} else {
			if validatedParties, err = k.validateAllRequiredPartiesSigned(ctx, scope.Owners, scope.Owners, scopeSpec.PartiesInvolved, msg); err != nil {
				return err
			}
		}
	}

	usedSigners, err := k.ValidateScopeValueOwnerUpdate(ctx, scope.ValueOwnerAddress, "", msg)
	if err != nil {
		return err
	}

	usedSigners.AlsoUse(GetUsedSigners(validatedParties))
	return k.validateSmartContractSigners(ctx, usedSigners, msg)
}

// ValidateSetScopeAccountData makes sure that the msg signers have proper authority to
// set the account data of the provided metadata address.
// Assumes that msg.MetadataAddr is a scope id.
func (k Keeper) ValidateSetScopeAccountData(ctx sdk.Context, msg *types.MsgSetAccountDataRequest) error {
	scope, found := k.GetScope(ctx, msg.MetadataAddr)
	if !found {
		// Allow deletion of account data if the scope no longer exists.
		if len(msg.Value) == 0 {
			return nil
		}
		return fmt.Errorf("scope not found with id %s", msg.MetadataAddr.String())
	}

	var err error
	var validatedParties []*PartyDetails

	// This is similar to ValidateDeleteScope, but the value owner isn't considered,
	// and we expect the scope spec to still exist.

	if !scope.RequirePartyRollup {
		// Old:
		//   - All roles required by the scope spec must have a party in the owners.
		//   - If not new, all existing owners must sign.
		//   - Value owner signer restrictions are applied.
		validatedParties, err = k.validateAllRequiredSigned(ctx, scope.GetAllOwnerAddresses(), msg)
		if err != nil {
			return err
		}
	} else {
		// New:
		//   - All roles required by the scope spec must have a party in the owners.
		//   - If not new, all required=false existing owners must be signers.
		//   - If not new, all roles required by the scope spec must have a signer and
		//     associated party from the existing scope.
		//   - Value owner signer restrictions are applied.
		scopeSpec, specFound := k.GetScopeSpecification(ctx, scope.SpecificationId)
		if !specFound {
			return fmt.Errorf("scope specification %s not found for scope id %s", scope.SpecificationId.String(), scope.ScopeId.String())
		}
		validatedParties, err = k.validateAllRequiredPartiesSigned(ctx, scope.Owners, scope.Owners, scopeSpec.PartiesInvolved, msg)
		if err != nil {
			return err
		}
	}

	return k.validateSmartContractSigners(ctx, GetUsedSigners(validatedParties), msg)
}

// ValidateAddScopeDataAccess checks the current scope and the proposed
func (k Keeper) ValidateAddScopeDataAccess(
	ctx sdk.Context,
	existing types.Scope,
	msg *types.MsgAddScopeDataAccessRequest,
) error {
	if len(msg.DataAccess) < 1 {
		return fmt.Errorf("data access list cannot be empty")
	}

	for _, da := range msg.DataAccess {
		_, err := sdk.AccAddressFromBech32(da)
		if err != nil {
			return fmt.Errorf("failed to decode data access address %s : %w", da, err)
		}
		for _, eda := range existing.DataAccess {
			if da == eda {
				return fmt.Errorf("address already exists for data access %s", eda)
			}
		}
	}

	// Make sure everyone has signed.
	if !existing.RequirePartyRollup {
		// Old:
		//   - All roles required by the scope spec must have a party in the owners.
		//   - If not new, all existing owners must sign.
		//   - Value owner signer restrictions are applied.
		// We don't care about the first one here since owners aren't changing.
		// We don't care about the value owner check either since it's not changing.
		if err := k.ValidateSignersWithoutParties(ctx, existing.GetAllOwnerAddresses(), msg); err != nil {
			return err
		}
	} else {
		// New:
		//   - All roles required by the scope spec must have a party in the owners.
		//   - If not new, all required=false existing owners must be signers.
		//   - If not new, all roles required by the scope spec must have a signer and
		//     associated party from the existing scope.
		//   - Value owner signer restrictions are applied.
		// We don't care about the first one here since owners aren't changing.
		// We don't care about the value owner check either since it's not changing.
		scopeSpec, found := k.GetScopeSpecification(ctx, existing.SpecificationId)
		if !found {
			return fmt.Errorf("scope specification %s not found", existing.SpecificationId)
		}
		if err := k.ValidateSignersWithParties(ctx, existing.Owners, existing.Owners, scopeSpec.PartiesInvolved, msg); err != nil {
			return err
		}
	}

	return nil
}

// ValidateDeleteScopeDataAccess checks the current scope data access and the proposed removed items
func (k Keeper) ValidateDeleteScopeDataAccess(
	ctx sdk.Context,
	existing types.Scope,
	msg *types.MsgDeleteScopeDataAccessRequest,
) error {
	if len(msg.DataAccess) < 1 {
		return fmt.Errorf("data access list cannot be empty")
	}

dataAccessLoop:
	for _, da := range msg.DataAccess {
		for _, eda := range existing.DataAccess {
			if da == eda {
				continue dataAccessLoop
			}
		}
		return fmt.Errorf("address does not exist in scope data access: %s", da)
	}

	// Make sure everyone has signed.
	if !existing.RequirePartyRollup {
		// Old:
		//   - All roles required by the scope spec must have a party in the owners.
		//   - If not new, all existing owners must sign.
		//   - Value owner signer restrictions are applied.
		// We don't care about the first one here since owners aren't changing.
		// We don't care about the value owner check either since it's not changing.
		if err := k.ValidateSignersWithoutParties(ctx, existing.GetAllOwnerAddresses(), msg); err != nil {
			return err
		}
	} else {
		// New:
		//   - All roles required by the scope spec must have a party in the owners.
		//   - If not new, all required=false existing owners must be signers.
		//   - If not new, all roles required by the scope spec must have a signer and
		//     associated party from the existing scope.
		//   - Value owner signer restrictions are applied.
		// We don't care about the first one here since owners aren't changing.
		// We don't care about the value owner check either since it's not changing.
		scopeSpec, found := k.GetScopeSpecification(ctx, existing.SpecificationId)
		if !found {
			return fmt.Errorf("scope specification %s not found", existing.SpecificationId)
		}
		if err := k.ValidateSignersWithParties(ctx, existing.Owners, existing.Owners, scopeSpec.PartiesInvolved, msg); err != nil {
			return err
		}
	}

	return nil
}

// ValidateUpdateScopeOwners checks the current scopes owners and the proposed update
func (k Keeper) ValidateUpdateScopeOwners(
	ctx sdk.Context,
	existing,
	proposed types.Scope,
	msg types.MetadataMsg,
) error {
	if err := proposed.ValidateOwnersBasic(); err != nil {
		return err
	}

	scopeSpec, found := k.GetScopeSpecification(ctx, proposed.SpecificationId)
	if !found {
		return fmt.Errorf("scope specification %s not found", proposed.SpecificationId)
	}

	if err := validateRolesPresent(proposed.Owners, scopeSpec.PartiesInvolved); err != nil {
		return err
	}
	if err := k.validateProvenanceRole(ctx, BuildPartyDetails(nil, proposed.Owners)); err != nil {
		return err
	}

	// Make sure everyone has signed.
	if !existing.RequirePartyRollup {
		// Old:
		//   - All roles required by the scope spec must have a party in the owners.
		//   - If not new, all existing owners must sign.
		//   - Value owner signer restrictions are applied.
		// The value owner isn't changing so we don't care about that one.
		if err := k.ValidateSignersWithoutParties(ctx, existing.GetAllOwnerAddresses(), msg); err != nil {
			return err
		}
	} else {
		// New:
		//   - All roles required by the scope spec must have a party in the owners.
		//   - If not new, all required=false existing owners must be signers.
		//   - If not new, all roles required by the scope spec must have a signer and
		//     associated party from the existing scope.
		//   - Value owner signer restrictions are applied.
		// The value owner isn't changing so we don't care about that one.
		validatedParties, err := k.validateAllRequiredPartiesSigned(ctx, existing.Owners, existing.Owners, scopeSpec.PartiesInvolved, msg)
		if err != nil {
			return err
		}
		if err = k.validateSmartContractSigners(ctx, GetUsedSigners(validatedParties), msg); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) ValidateUpdateValueOwners(
	ctx sdk.Context,
	scopes []*types.Scope,
	newValueOwner string,
	msg types.MetadataMsg,
) error {
	var existingValueOwners []string
	knownValueOwners := make(map[string]bool)

	for _, scope := range scopes {
		if len(scope.ValueOwnerAddress) == 0 {
			return fmt.Errorf("scope %s does not yet have a value owner", scope.ScopeId)
		}
		if !knownValueOwners[scope.ValueOwnerAddress] {
			existingValueOwners = append(existingValueOwners, scope.ValueOwnerAddress)
			knownValueOwners[scope.ValueOwnerAddress] = true
		}
	}

	signers := NewSignersWrapper(msg.GetSignerStrs())
	usedSigners, err := k.validateScopeValueOwnerChangeToProposed(ctx, newValueOwner, signers)
	if err != nil {
		return err
	}

	for _, existing := range existingValueOwners {
		alsoUsedSigners, err := k.validateScopeValueOwnerChangeFromExisting(ctx, existing, signers, msg)
		if err != nil {
			return err
		}
		usedSigners.AlsoUse(alsoUsedSigners)
	}

	return k.validateSmartContractSigners(ctx, usedSigners, msg)
}
