package keeper

import (
	"errors"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"

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

// SetScopeValueOwners updates the value owner of all provided scopes.
// A coin will be minted for any scope that does not yet have a value owner.
// If the newValueOwner is empty, all of the coins for the provided scopes will be burned.
//
// Contract: The provided AccMDLinks must accurately indicate the current value owners of each scope in the AccAddr fields.
// If the AccAddr field is nil or empty, that indicates that there is no value owner yet for the scope.
func (k Keeper) SetScopeValueOwners(ctx sdk.Context, links types.AccMDLinks, newValueOwner string) error {
	links = links.WithNilsRemoved()
	if len(links) == 0 {
		return nil
	}

	if err := links.ValidateAllAreScopes(); err != nil {
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
			return sdkerrors.ErrUnauthorized.Wrapf("new value owner %s is not allowed to receive funds", newValueOwner)
		}
	}

	// Identify the addresses and the amounts to send to each, and also the amounts to mint or burn.
	var fromAddrs []sdk.AccAddress
	fromAddrAmts := make(map[string]sdk.Coins)
	var toMint, toBurn sdk.Coins
	for _, link := range links {
		coin := link.MDAddr.Coin()

		if len(link.AccAddr) == 0 {
			if doBurn {
				// There's no AccAddr which means it hasn't been minted yet, but we're burning so we can just ignore it.
				continue
			}
			// Make sure the lack of an existing owner isn't a lie.
			supply := k.bankKeeper.GetSupply(ctx, coin.Denom)
			if !supply.IsZero() {
				return fmt.Errorf("cannot mint scope coin for %q: supply %s is not zero", link.MDAddr, supply.Amount)
			}
			// Add it to the amount to mint and send it from the module account.
			toMint = toMint.Add(coin)
			link.AccAddr = k.moduleAddr
		}

		cur, seen := fromAddrAmts[string(link.AccAddr)]
		if !seen {
			fromAddrs = append(fromAddrs, link.AccAddr)
		}
		fromAddrAmts[string(link.AccAddr)] = cur.Add(coin)
		if doBurn {
			toBurn = toBurn.Add(coin)
		}
	}

	// Mint anything that needs minting.
	if !doBurn && !toMint.IsZero() {
		if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, toMint); err != nil {
			return fmt.Errorf("could not mint scope coins %q: %w", toMint, err)
		}
	}

	// Do all the sending!
	for _, fromAddr := range fromAddrs {
		if !toAddr.Equals(fromAddr) {
			if err := k.bankKeeper.SendCoins(ctx, fromAddr, toAddr, fromAddrAmts[string(fromAddr)]); err != nil {
				return fmt.Errorf("could not send scope coins %q from %s to %s: %w", fromAddrAmts[string(fromAddr)], fromAddr, toAddr, err)
			}
		}
	}

	// If we're burning it, it should all be in the module account, so we can burn it all!
	if doBurn && !toBurn.IsZero() {
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, toBurn); err != nil {
			return fmt.Errorf("could not burn scope coins %q", toBurn)
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
	rv.Addresses = findMissingComp(required.Addresses, found.Addresses, func(a1 sdk.AccAddress, a2 sdk.AccAddress) bool {
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

// SetNetAssetValue adds/updates a net asset value to scope
func (k Keeper) SetNetAssetValue(ctx sdk.Context, scopeID types.MetadataAddress, netAssetValue types.NetAssetValue, source string) error {
	netAssetValue.UpdatedBlockHeight = uint64(ctx.BlockHeight())
	if err := netAssetValue.Validate(); err != nil {
		return err
	}

	setNetAssetValueEvent := types.NewEventSetNetAssetValue(scopeID, netAssetValue.Price, source)
	if err := ctx.EventManager().EmitTypedEvent(setNetAssetValueEvent); err != nil {
		return err
	}

	key := types.NetAssetValueKey(scopeID, netAssetValue.Price.Denom)
	store := ctx.KVStore(k.storeKey)

	bz, err := k.cdc.Marshal(&netAssetValue)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	return nil
}

// IterateNetAssetValues iterates net asset values for scope
func (k Keeper) IterateNetAssetValues(ctx sdk.Context, scopeID types.MetadataAddress, handler func(state types.NetAssetValue) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	it := storetypes.KVStorePrefixIterator(store, types.NetAssetValueKeyPrefix(scopeID))
	defer it.Close()
	for ; it.Valid(); it.Next() {
		var scopeNav types.NetAssetValue
		err := k.cdc.Unmarshal(it.Value(), &scopeNav)
		if err != nil {
			return err
		} else if handler(scopeNav) {
			break
		}
	}
	return nil
}

// RemoveNetAssetValues removes all net asset values for a scope
func (k Keeper) RemoveNetAssetValues(ctx sdk.Context, scopeID types.MetadataAddress) {
	store := ctx.KVStore(k.storeKey)
	it := storetypes.KVStorePrefixIterator(store, types.NetAssetValueKeyPrefix(scopeID))
	var keys [][]byte
	for ; it.Valid(); it.Next() {
		keys = append(keys, it.Key())
	}
	it.Close()

	for _, key := range keys {
		store.Delete(key)
	}
}

// SetNetAssetValueWithBlockHeight adds/updates a net asset value to scope with a specific block height
func (k Keeper) SetNetAssetValueWithBlockHeight(ctx sdk.Context, scopeID types.MetadataAddress, netAssetValue types.NetAssetValue, source string, blockHeight uint64) error {
	netAssetValue.UpdatedBlockHeight = blockHeight
	if err := netAssetValue.Validate(); err != nil {
		return err
	}

	setNetAssetValueEvent := types.NewEventSetNetAssetValue(scopeID, netAssetValue.Price, source)
	if err := ctx.EventManager().EmitTypedEvent(setNetAssetValueEvent); err != nil {
		return err
	}

	key := types.NetAssetValueKey(scopeID, netAssetValue.Price.Denom)
	bz, err := k.cdc.Marshal(&netAssetValue)
	if err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	store.Set(key, bz)

	return nil
}
