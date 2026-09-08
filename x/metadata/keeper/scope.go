package keeper

import (
	"errors"
	"fmt"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"

	"github.com/provenance-io/provenance/internal/provutils"
	"github.com/provenance-io/provenance/x/metadata/types"
)

// IterateScopes processes all stored scopes with the given handler.
func (k Keeper) IterateScopes(ctx sdk.Context, handler func(types.Scope) (stop bool)) error {
	return k.scopeCollections.Walk(ctx, nil, func(_ []byte, scope types.Scope) (stop bool, err error) {
		// Clear ValueOwnerAddress — historical state may have it set; it is not trusted.
		// See original readScopeBz comment.
		scope.ValueOwnerAddress = ""
		k.PopulateScopeValueOwner(ctx, &scope)
		return handler(scope), nil
	})
}

// IterateScopesForAddress processes scopes associated with the provided address with the given handler.
func (k Keeper) IterateScopesForAddress(ctx sdk.Context, addr sdk.AccAddress, handler func(scopeID types.MetadataAddress) (stop bool)) error {
	// Index key format: length_prefix(addr) + scopeID_17bytes
	// Use length_prefix(addr) as the prefix to iterate all scopes for this address.
	addrPrefix := address.MustLengthPrefix(addr)
	return k.addressScopeIndex.Walk(ctx,
		new(collections.Range[[]byte]).Prefix(addrPrefix),
		func(key []byte, _ []byte) (stop bool, err error) {
			// key = length_prefix(addr) + scopeID_17bytes
			// Strip the address prefix to get the scopeID bytes (full 17-byte MetadataAddress).
			scopeIDBytes := key[len(addrPrefix):]
			var scopeID types.MetadataAddress
			if err := scopeID.Unmarshal(scopeIDBytes); err != nil {
				return false, err
			}
			return handler(scopeID), nil
		})
}

// IterateScopesForScopeSpec processes scopes associated with the provided scope specification id with the given handler.
func (k Keeper) IterateScopesForScopeSpec(ctx sdk.Context, scopeSpecID types.MetadataAddress,
	handler func(scopeID types.MetadataAddress) (stop bool),
) error {
	// Index key format: scopeSpecID_17bytes + scopeID_17bytes
	specPrefix := scopeSpecID.Bytes() // full 17 bytes including type prefix byte
	return k.scopeSpecScopeIndex.Walk(ctx,
		new(collections.Range[[]byte]).Prefix(specPrefix),
		func(key []byte, _ []byte) (stop bool, err error) {
			scopeIDBytes := key[len(specPrefix):]
			var scopeID types.MetadataAddress
			if err := scopeID.Unmarshal(scopeIDBytes); err != nil {
				return false, err
			}
			return handler(scopeID), nil
		})
}

// GetScope returns the scope with the given id. The ValueOwnerAddress field will always be empty from this method.
// See also: GetScopeWithValueOwner, PopulateScopeValueOwner and GetScopeValueOwner.
func (k Keeper) GetScope(ctx sdk.Context, id types.MetadataAddress) (types.Scope, bool) {
	if !id.IsScopeAddress() {
		return types.Scope{}, false
	}
	val, err := k.scopeCollections.Get(ctx, mdKey(id))
	if err != nil {
		return types.Scope{}, false
	}
	// Clear ValueOwnerAddress — not trusted from state (managed by bank module).
	val.ValueOwnerAddress = ""
	return val, true
}

// GetScopeWithValueOwner will get a scope from state and also look up and set its value owner field.
func (k Keeper) GetScopeWithValueOwner(ctx sdk.Context, id types.MetadataAddress) (scope types.Scope, found bool) {
	scope, found = k.GetScope(ctx, id)
	if found {
		k.PopulateScopeValueOwner(ctx, &scope)
	}
	return scope, found
}

// PopulateScopeValueOwner will look up and set the ValueOwnerAddress in the provided scope.
func (k Keeper) PopulateScopeValueOwner(ctx sdk.Context, scope *types.Scope) {
	vo, err := k.GetScopeValueOwner(ctx, scope.ScopeId)
	if err == nil && len(vo) > 0 {
		scope.ValueOwnerAddress = vo.String()
	} else {
		scope.ValueOwnerAddress = ""
	}
}

// SetScope stores a scope in the module kv store.
func (k Keeper) SetScope(ctx sdk.Context, scope types.Scope) error {
	// If there's a value owner in the provided scope, update that record then remove it from the
	// scope before writing the scope. If it doesn't have a value owner, we don't do anything about it.
	// It shouldn't be possible to delete the value owner record once there is one for a scope.
	if len(scope.ValueOwnerAddress) > 0 {
		err := k.SetScopeValueOwner(ctx, scope.ScopeId, scope.ValueOwnerAddress)
		if err != nil {
			return fmt.Errorf("could not set value owner: %w", err)
		}
		scope.ValueOwnerAddress = ""
	}

	k.writeScopeToState(ctx, scope)
	return nil
}

// writeScopeToState writes the given scope to state, updates the related indexes, and emits the appropriate events.
// It's split out from SetScope only so that unit tests can write scopes that have something in the value owner field.
func (k Keeper) writeScopeToState(ctx sdk.Context, scope types.Scope) {
	var oldScope *types.Scope
	var event proto.Message = types.NewEventScopeCreated(scope.ScopeId)

	if has, _ := k.scopeCollections.Has(ctx, mdKey(scope.ScopeId)); has {
		event = types.NewEventScopeUpdated(scope.ScopeId)
		if existing, err := k.scopeCollections.Get(ctx, mdKey(scope.ScopeId)); err == nil {
			existing.ValueOwnerAddress = "" // clear before using for index comparison
			oldScope = &existing
		}
	}

	if err := k.scopeCollections.Set(ctx, mdKey(scope.ScopeId), scope); err != nil {
		panic(err)
	}
	k.indexScope(ctx, &scope, oldScope)
	k.EmitEvent(ctx, event)
}

// RemoveScope removes a scope from the module kv store along with all its records and sessions.
func (k Keeper) RemoveScope(ctx sdk.Context, id types.MetadataAddress) error {
	if err := id.ValidateIsScopeAddress(); err != nil {
		return err
	}

	// If the scope already does not exist, don't do anything.
	scope, found := k.GetScope(ctx, id)
	if !found {
		return nil
	}

	// Burn the scope's value owner coin.
	if err := k.SetScopeValueOwner(ctx, id, ""); err != nil {
		return fmt.Errorf("could not remove scope %s value owner: %w", id, err)
	}

	// Remove all records. The sessions are deleted by RemoveRecord as the last record in each is deleted.
	// Collect all record IDs first, then remove them.
	// (Each RemoveRecord may also remove its session once the last record is gone.)
	var recordIDs []types.MetadataAddress
	if err := k.IterateRecords(ctx, id, func(r types.Record) (stop bool) {
		recordID := r.SessionId.MustGetAsRecordAddress(r.Name)
		recordIDs = append(recordIDs, recordID)
		return false
	}); err != nil {
		return err
	}
	for _, recordID := range recordIDs {
		k.RemoveRecord(ctx, recordID)
	}

	k.indexScope(ctx, nil, &scope)
	if err := k.scopeCollections.Remove(ctx, mdKey(id)); err != nil {
		return err
	}
	k.EmitEvent(ctx, types.NewEventScopeDeleted(scope.ScopeId))
	return nil
}

// GetScopeValueOwner gets the value owner of a given scope.
func (k Keeper) GetScopeValueOwner(ctx sdk.Context, id types.MetadataAddress) (sdk.AccAddress, error) {
	if !id.IsScopeAddress() {
		return nil, fmt.Errorf("cannot get value owner for non-scope metadata address %q", id)
	}
	return k.bankKeeper.DenomOwner(ctx, id.Denom())
}

// GetScopeValueOwners gets the value owners for each given scope.
// The AccAddr will be nil for any scope that does not have a value owner.
func (k Keeper) GetScopeValueOwners(ctx sdk.Context, ids []types.MetadataAddress) (types.AccMDLinks, error) {
	var errs []error
	rv := make(types.AccMDLinks, 0, len(ids))
	for _, id := range ids {
		addr, err := k.GetScopeValueOwner(ctx, id)
		if err != nil {
			errs = append(errs, err)
		} else {
			rv = append(rv, types.NewAccMDLink(addr, id))
		}
	}
	return rv, errors.Join(errs...)
}

// SetScopeValueOwner updates the value owner of a scope.
// If there's no current value owner, the coin will be minted for the scope.
// If there's no new value owner, the coin will be burned for the scope.
func (k Keeper) SetScopeValueOwner(ctx sdk.Context, scopeID types.MetadataAddress, newValueOwner string) error {
	if err := scopeID.ValidateIsScopeAddress(); err != nil {
		return err
	}

	var toAddr sdk.AccAddress
	doBurn := false
	if len(newValueOwner) == 0 {
		// If there's no new value owner, we'll send everything to the module account so that we can then burn it.
		toAddr = k.moduleAddr
		doBurn = true
	} else {
		// Sending to another account, so make sure it's valid and not blocked.
		var err error
		toAddr, err = sdk.AccAddressFromBech32(newValueOwner)
		if err != nil {
			return fmt.Errorf("invalid new value owner address %q: %w", newValueOwner, err)
		}
		if k.bankKeeper.BlockedAddr(toAddr) {
			return sdkerrors.ErrUnauthorized.Wrapf("new value owner %q is not allowed to receive funds", newValueOwner)
		}
	}

	coin := scopeID.Coin()
	fromAddr, err := k.bankKeeper.DenomOwner(ctx, coin.Denom)
	if err != nil {
		return fmt.Errorf("could not get current value owner of %q: %w", scopeID, err)
	}
	if fromAddr.String() == newValueOwner {
		// no change, nothing more to do.
		return nil
	}

	coins := sdk.Coins{coin}
	if len(fromAddr) == 0 {
		// If there's no current value owner, we'll mint it and send it from the module account.
		fromAddr = k.moduleAddr
		if err = k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
			return fmt.Errorf("could not mint scope coin %q: %w", coins, err)
		}
	}

	if err = k.bankKeeper.SendCoins(ctx, fromAddr, toAddr, coins); err != nil {
		return fmt.Errorf("could not send scope coin %q from %s to %s: %w", coins, fromAddr, toAddr, err)
	}

	if doBurn {
		if err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins); err != nil {
			return fmt.Errorf("could not burn scope coin %q: %w", coins, err)
		}
	}

	return nil
}

// SetScopeValueOwners updates the value owner of one or more scopes.
func (k Keeper) SetScopeValueOwners(ctx sdk.Context, links types.AccMDLinks, newValueOwner string) error {
	if len(links) == 0 {
		return nil
	}

	if err := links.ValidateForScopes(); err != nil {
		return err
	}

	toAddr, err := sdk.AccAddressFromBech32(newValueOwner)
	if err != nil {
		return fmt.Errorf("invalid new value owner address %q: %w", newValueOwner, err)
	}
	if k.bankKeeper.BlockedAddr(toAddr) {
		return sdkerrors.ErrUnauthorized.Wrapf("new value owner %s is not allowed to receive funds", newValueOwner)
	}

	// Identify the addresses and the amounts to send to each.
	var fromAddrs []sdk.AccAddress
	fromAddrAmts := make(map[string]sdk.Coins)
	for _, link := range links {
		coin := link.MDAddr.Coin()
		cur, seen := fromAddrAmts[string(link.AccAddr)]
		if !seen {
			fromAddrs = append(fromAddrs, link.AccAddr)
		}
		fromAddrAmts[string(link.AccAddr)] = cur.Add(coin)
	}

	for _, fromAddr := range fromAddrs {
		if !toAddr.Equals(fromAddr) {
			if err = k.bankKeeper.SendCoins(ctx, fromAddr, toAddr, fromAddrAmts[string(fromAddr)]); err != nil {
				return fmt.Errorf("could not send scope coins %q from %s to %s: %w", fromAddrAmts[string(fromAddr)], fromAddr, toAddr, err)
			}
		}
	}

	return nil
}

// scopeIndexValues is a struct containing the values used to index a scope.
type scopeIndexValues struct {
	ScopeID         types.MetadataAddress
	Addresses       []sdk.AccAddress
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
	rv.Addresses = provutils.FindMissingFunc(required.Addresses, found.Addresses, func(a1, a2 sdk.AccAddress) bool {
		return a1.Equals(a2)
	})
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

	toAdd := getMissingScopeIndexValues(getScopeIndexValues(newScope), getScopeIndexValues(oldScope))
	toRemove := getMissingScopeIndexValues(getScopeIndexValues(oldScope), getScopeIndexValues(newScope))

	applyKey := func(indexKey []byte, set bool) {
		if len(indexKey) == 0 {
			return
		}
		// indexKey[0] is the collection prefix byte; indexKey[1:] is the collection key.
		subKey := indexKey[1:]
		switch indexKey[0] {
		case types.AddressScopeCacheKeyPrefix[0]: // 0x17 → AddressScopeIndex
			if set {
				_ = k.addressScopeIndex.Set(ctx, subKey, indexPresent)
			} else {
				_ = k.addressScopeIndex.Remove(ctx, subKey)
			}
		case types.ScopeSpecScopeCacheKeyPrefix[0]: // 0x11 → ScopeSpecScopeIndex
			if set {
				_ = k.scopeSpecScopeIndex.Set(ctx, subKey, indexPresent)
			} else {
				_ = k.scopeSpecScopeIndex.Remove(ctx, subKey)
			}
		}
	}

	for _, indexKey := range toAdd.IndexKeys() {
		applyKey(indexKey, true)
	}
	for _, indexKey := range toRemove.IndexKeys() {
		applyKey(indexKey, false)
	}
}

// ValidateWriteScope checks the current scope and the proposed scope to determine if the proposed changes are valid
// based on the existing state. Returns the addresses allowed to act as transfer agents.
func (k Keeper) ValidateWriteScope(
	ctx sdk.Context,
	msg *types.MsgWriteScopeRequest,
) ([]sdk.AccAddress, error) {
	proposed := msg.Scope
	if err := proposed.ValidateBasic(); err != nil {
		return nil, err
	}

	var existing *types.Scope
	if e, found := k.GetScope(ctx, msg.Scope.ScopeId); found {
		existing = &e
	}

	// If the scope already exists:
	//   - Lack of a proposed value owner means there is no desired change to it and we don't need to look it up.
	//   - Presence of a proposed value owner means we need to look up the existing one
	//     and require them to be a signer iff it's different from the proposed value owner.
	var existingVOAddrs []sdk.AccAddress
	if existing != nil && len(proposed.ValueOwnerAddress) > 0 {
		vo, err := k.GetScopeValueOwner(ctx, proposed.ScopeId)
		if err != nil {
			return nil, fmt.Errorf("error identifying current value owner of %q: %w", proposed.ScopeId, err)
		}
		// It is possible for scopes to not have a value owner.
		if len(vo) > 0 {
			existing.ValueOwnerAddress = vo.String()
			existingVOAddrs = append(existingVOAddrs, vo)
		}
	}

	// Existing owners are not required to sign when the ONLY change is from one value owner to another.
	// Signatures from existing owners are required if:
	//   - Anything other than the value owner is changing.
	//   - There's a proposed value owner and the scope exists, but does not yet have a value owner.
	onlyChangeIsValueOwner := false
	if existing != nil && len(existing.ValueOwnerAddress) > 0 && existing.ValueOwnerAddress != proposed.ValueOwnerAddress {
		// Make a copy of proposed scope and set its value owner to the existing one. If it then
		// equals the existing scope, then the only change in proposed is to the value owner field.
		proposedCopy := proposed
		proposedCopy.ValueOwnerAddress = existing.ValueOwnerAddress
		onlyChangeIsValueOwner = existing.Equals(proposedCopy)
	}

	var err error
	var validatedParties []*types.PartyDetails

	if !onlyChangeIsValueOwner {
		scopeSpec, found := k.GetScopeSpecification(ctx, proposed.SpecificationId)
		if !found {
			return nil, fmt.Errorf("scope specification %s not found", proposed.SpecificationId)
		}

		if err = validateRolesPresent(proposed.Owners, scopeSpec.PartiesInvolved); err != nil {
			return nil, err
		}
		if err = k.validateProvenanceRole(ctx, types.BuildPartyDetails(nil, proposed.Owners)); err != nil {
			return nil, err
		}

		// Make sure everyone has signed.
		if (existing != nil && !existing.RequirePartyRollup) || (existing == nil && !proposed.RequirePartyRollup) {
			// Old:
			//   - All roles required by the scope spec must have a party in the owners.
			//   - If not new, all existing owners must sign.
			//   - Value owner signer restrictions are applied.
			if existing != nil && !existing.Equals(proposed) {
				if validatedParties, err = k.validateAllRequiredSigned(ctx, existing.GetAllOwnerAddresses(), msg); err != nil {
					return nil, err
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
					return nil, err
				}
			}
		}
	}

	transferAgents, usedSigners, err := k.ValidateScopeValueOwnersSigners(ctx, existingVOAddrs, proposed.ValueOwnerAddress, msg)
	if err != nil {
		return nil, err
	}

	usedSigners.AlsoUse(types.GetUsedSigners(validatedParties))
	if err = k.validateSmartContractSigners(ctx, usedSigners, msg); err != nil {
		return nil, err
	}
	return transferAgents, nil
}

// ValidateDeleteScope checks the current scope and the proposed removal scope to determine if the proposed remove is valid
// based on the existing state
func (k Keeper) ValidateDeleteScope(ctx sdk.Context, msg *types.MsgDeleteScopeRequest) ([]sdk.AccAddress, error) {
	scope, found := k.GetScope(ctx, msg.ScopeId)
	if !found {
		return nil, fmt.Errorf("scope not found with id %s", msg.ScopeId)
	}

	var err error
	var validatedParties []*types.PartyDetails

	// Make sure everyone has signed.
	if !scope.RequirePartyRollup {
		// Old:
		//   - All roles required by the scope spec must have a party in the owners.
		//   - If not new, all existing owners must sign.
		//   - Value owner signer restrictions are applied.
		// We don't care about the first one here.
		if validatedParties, err = k.validateAllRequiredSigned(ctx, scope.GetAllOwnerAddresses(), msg); err != nil {
			return nil, err
		}
	} else {
		// New:
		//   - All roles required by the scope spec must have a party in the owners.
		//   - If not new, all required=false existing owners must be signers.
		//   - If not new, all roles required by the scope spec must have a signer and
		//     associated party from the existing scope.
		//   - Value owner signer restrictions are applied.
		// We don't care about that first one, and only care about the roles one if the spec exists.
		scopeSpec, specFound := k.GetScopeSpecification(ctx, scope.SpecificationId)
		if !specFound {
			if validatedParties, err = k.validateAllRequiredSigned(ctx, types.GetRequiredPartyAddresses(scope.Owners), msg); err != nil {
				return nil, err
			}
		} else {
			if validatedParties, err = k.validateAllRequiredPartiesSigned(ctx, scope.Owners, scope.Owners, scopeSpec.PartiesInvolved, msg); err != nil {
				return nil, err
			}
		}
	}

	var existingVOAddrs []sdk.AccAddress
	vo, err := k.GetScopeValueOwner(ctx, scope.ScopeId)
	if err != nil {
		return nil, fmt.Errorf("error identifying current value owner of %q: %w", scope.ScopeId, err)
	}
	// It is possible for older scopes to not have a value owner.
	if len(vo) > 0 {
		scope.ValueOwnerAddress = vo.String()
		existingVOAddrs = append(existingVOAddrs, vo)
	}

	transferAgents, usedSigners, err := k.ValidateScopeValueOwnersSigners(ctx, existingVOAddrs, "", msg)
	if err != nil {
		return nil, err
	}

	usedSigners.AlsoUse(types.GetUsedSigners(validatedParties))
	err = k.validateSmartContractSigners(ctx, usedSigners, msg)
	if err != nil {
		return nil, err
	}
	return transferAgents, nil
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
	var validatedParties []*types.PartyDetails

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

	return k.validateSmartContractSigners(ctx, types.GetUsedSigners(validatedParties), msg)
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
	if err := k.validateProvenanceRole(ctx, types.BuildPartyDetails(nil, proposed.Owners)); err != nil {
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
		if err = k.validateSmartContractSigners(ctx, types.GetUsedSigners(validatedParties), msg); err != nil {
			return err
		}
	}

	return nil
}

// ValidateUpdateValueOwners checks that the signer(s) of the provided msg are authorized to change the value owner
// of the scopes in the links provided. Also checks that the provided links are valid.
// Returns the transfer agents available for the SendCoins.
func (k Keeper) ValidateUpdateValueOwners(
	ctx sdk.Context,
	links types.AccMDLinks,
	proposed string,
	msg types.MetadataMsg,
) ([]sdk.AccAddress, error) {
	if len(links) == 0 {
		return nil, errors.New("no scopes found")
	}
	if err := links.ValidateForScopes(); err != nil {
		return nil, err
	}
	if ids := links.GetMDAddrsForAccAddr(proposed); len(ids) > 0 {
		if len(ids) == 1 {
			return nil, fmt.Errorf("scope %q already has the proposed value owner %q", ids[0], proposed)
		}
		return nil, fmt.Errorf("scopes %q already have the proposed value owner %q", ids, proposed)
	}

	transferAgents, _, err := k.ValidateScopeValueOwnersSigners(ctx, links.GetAccAddrs(), proposed, msg)
	return transferAgents, err
}

// AddSetNetAssetValues adds a set of net asset values to a scope
func (k Keeper) AddSetNetAssetValues(ctx sdk.Context, scopeID types.MetadataAddress, netAssetValues []types.NetAssetValue, source string) error {
	for _, nav := range netAssetValues {
		if nav.Price.Denom != types.UsdDenom {
			_, err := k.markerKeeper.GetMarkerByDenom(ctx, nav.Price.Denom)
			if err != nil {
				return fmt.Errorf("net asset value denom does not exist: %v", err.Error())
			}
		}

		if err := k.SetNetAssetValue(ctx, scopeID, nav, source); err != nil {
			return fmt.Errorf("cannot set net asset value : %v", err.Error())
		}
	}
	return nil
}

// GetNetAssetValue gets the NAV record for the given scopeID with the given price denom.
// If it doesn't exist then nil, nil is returned.
func (k Keeper) GetNetAssetValue(ctx sdk.Context, metadataDenom, priceDenom string) (*types.NetAssetValue, error) {
	scopeID, err := types.MetadataAddressFromDenom(metadataDenom)
	if err != nil {
		return nil, fmt.Errorf("could not get metadata address: %w", err)
	}

	navKey := append(address.MustLengthPrefix(scopeID.Bytes()), []byte(priceDenom)...)
	val, err := k.netAssetValues.Get(ctx, navKey)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, nil // genuinely not found
		}
		return nil, fmt.Errorf("could not read nav for %q with price denom %q: %w", scopeID, priceDenom, err)
	}
	if val.Volume < 1 {
		val.Volume = 1
	}
	return &val, nil
}

// SetNetAssetValue adds/updates a net asset value to scope
func (k Keeper) SetNetAssetValue(ctx sdk.Context, scopeID types.MetadataAddress, netAssetValue types.NetAssetValue, source string) error {
	netAssetValue.UpdatedBlockHeight = uint64(ctx.BlockHeight()) //nolint:gosec // G115
	if err := netAssetValue.Validate(); err != nil {
		return err
	}

	// Since this field was added we need to ensure the default value matches the previous behavior of always presuming one is used.
	if netAssetValue.Volume < 1 {
		netAssetValue.Volume = 1
	}

	setNetAssetValueEvent := types.NewEventSetNetAssetValue(scopeID, netAssetValue.Price, netAssetValue.Volume, source)
	if err := ctx.EventManager().EmitTypedEvent(setNetAssetValueEvent); err != nil {
		return err
	}

	navKey := append(address.MustLengthPrefix(scopeID.Bytes()), []byte(netAssetValue.Price.Denom)...)
	return k.netAssetValues.Set(ctx, navKey, netAssetValue)
}

// IterateNetAssetValues iterates net asset values for scope
func (k Keeper) IterateNetAssetValues(ctx sdk.Context, scopeID types.MetadataAddress, handler func(state types.NetAssetValue) (stop bool)) error {
	// NAV key = length_prefix(scopeAddr) + denom
	navPrefix := address.MustLengthPrefix(scopeID.Bytes())
	return k.netAssetValues.Walk(ctx,
		new(collections.Range[[]byte]).Prefix(navPrefix),
		func(_ []byte, nav types.NetAssetValue) (stop bool, err error) {
			return handler(nav), nil
		})
}

// RemoveNetAssetValues removes all net asset values for a scope
func (k Keeper) RemoveNetAssetValues(ctx sdk.Context, scopeID types.MetadataAddress) {
	navPrefix := address.MustLengthPrefix(scopeID.Bytes())

	// Collect all keys first, then delete (cannot mutate while iterating).
	var keys [][]byte
	_ = k.netAssetValues.Walk(ctx,
		new(collections.Range[[]byte]).Prefix(navPrefix),
		func(key []byte, _ types.NetAssetValue) (stop bool, err error) {
			keys = append(keys, key)
			return false, nil
		})
	for _, key := range keys {
		_ = k.netAssetValues.Remove(ctx, key)
	}
}

// SetNetAssetValueWithBlockHeight adds/updates a net asset value to scope with a specific block height
func (k Keeper) SetNetAssetValueWithBlockHeight(ctx sdk.Context, scopeID types.MetadataAddress, netAssetValue types.NetAssetValue, source string, blockHeight uint64) error {
	netAssetValue.UpdatedBlockHeight = blockHeight
	if err := netAssetValue.Validate(); err != nil {
		return err
	}

	setNetAssetValueEvent := types.NewEventSetNetAssetValue(scopeID, netAssetValue.Price, netAssetValue.Volume, source)
	if err := ctx.EventManager().EmitTypedEvent(setNetAssetValueEvent); err != nil {
		return err
	}

	navKey := append(address.MustLengthPrefix(scopeID.Bytes()), []byte(netAssetValue.Price.Denom)...)
	return k.netAssetValues.Set(ctx, navKey, netAssetValue)
}
