package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// GetOsLocatorRecord Gets the object store locator entry from the kvstore for the given owner address.
func (k Keeper) GetOsLocatorRecord(ctx sdk.Context, ownerAddr sdk.AccAddress) (osLocator types.ObjectStoreLocator, found bool) {
	key := types.GetOSLocatorKey(ownerAddr)
	store := ctx.KVStore(k.storeKey)
	b := store.Get(key)
	if b == nil {
		return types.ObjectStoreLocator{}, false
	}
	err := k.cdc.Unmarshal(b, &osLocator)
	if err != nil {
		ctx.Logger().Error("failed to unmarshal locator", "err", err)
		return types.ObjectStoreLocator{}, false
	}
	return osLocator, true
}

// OSLocatorExists checks if the provided bech32 owner address has a OSL entry in the kvstore.
func (k Keeper) OSLocatorExists(ctx sdk.Context, ownerAddr sdk.AccAddress) bool {
	key := types.GetOSLocatorKey(ownerAddr)
	store := ctx.KVStore(k.storeKey)
	return store.Has(key)
}

// SetOSLocator binds an OS Locator to an address in the kvstore.
// An error is returned if no account exists for the address.
// An error is returned if an OS Locator already exists for the address.
func (k Keeper) SetOSLocator(ctx sdk.Context, ownerAddr, encryptionKey sdk.AccAddress, uri string) error {
	urlToPersist, err := k.checkValidURI(uri, ctx)
	if err != nil {
		return err
	}
	if account := k.authKeeper.GetAccount(ctx, ownerAddr); account == nil {
		return types.ErrInvalidAddress
	}
	key := types.GetOSLocatorKey(ownerAddr)
	store := ctx.KVStore(k.storeKey)
	if store.Has(key) {
		return types.ErrOSLocatorAlreadyBound
	}

	record := types.NewOSLocatorRecord(ownerAddr, encryptionKey, urlToPersist.String())
	bz, err := k.cdc.Marshal(&record)
	if err != nil {
		return err
	}
	store.Set(key, bz)

	k.EmitEvent(ctx, types.NewEventOSLocatorCreated(record.Owner))
	defer types.GetIncObjFunc(types.TLType_OSLocator, types.TLAction_Created)
	return nil
}

// IterateOSLocators runs a function for every ObjectStoreLocator entry in the kvstore.
func (k Keeper) IterateOSLocators(ctx sdk.Context, cb func(account types.ObjectStoreLocator) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	prefix := types.OSLocatorAddressKeyPrefix
	it := sdk.KVStorePrefixIterator(store, prefix)

	defer it.Close()

	for ; it.Valid(); it.Next() {
		record := types.ObjectStoreLocator{}
		if err := k.cdc.Unmarshal(it.Value(), &record); err != nil {
			return err
		}
		if cb(record) {
			break
		}
	}
	return nil
}

// GetOSLocatorByScope gets all Object Store Locators associated with a scope.
func (k Keeper) GetOSLocatorByScope(ctx sdk.Context, scopeID string) ([]types.ObjectStoreLocator, error) {
	scopeAddr, err := ParseScopeID(scopeID)
	if err != nil {
		return []types.ObjectStoreLocator{}, err
	}

	scope, found := k.GetScope(ctx, scopeAddr)
	if !found {
		return []types.ObjectStoreLocator{}, fmt.Errorf("scope [%s] not found", scopeID)
	}

	// should always have valid owners, hence creating it with capacity
	signers := make([]sdk.AccAddress, len(scope.Owners))

	for i, p := range scope.Owners {
		addr, err := sdk.AccAddressFromBech32(p.Address)
		if err != nil {
			panic(err)
		}
		signers[i] = addr
	}

	// may not have object locators defined for all owners
	locators := make([]types.ObjectStoreLocator, 0, len(signers))
	for _, addr := range signers {
		loc, found := k.GetOsLocatorRecord(ctx, addr)
		if !found {
			continue
		}
		locators = append(locators, loc)
	}
	return locators, nil
}

// RemoveOSLocator removes an os locator record from the kvstore.
func (k Keeper) RemoveOSLocator(ctx sdk.Context, ownerAddr sdk.AccAddress) error {
	key := types.GetOSLocatorKey(ownerAddr)
	store := ctx.KVStore(k.storeKey)
	if !store.Has(key) {
		return types.ErrAddressNotBound
	}
	store.Delete(key)
	k.EmitEvent(ctx, types.NewEventOSLocatorDeleted(ownerAddr.String()))
	defer types.GetIncObjFunc(types.TLType_OSLocator, types.TLAction_Deleted)
	return nil
}

// ModifyOSLocator updates an existing os locator entry in the kvstore, returns an error if it doesn't exist.
func (k Keeper) ModifyOSLocator(ctx sdk.Context, ownerAddr, encryptionKey sdk.AccAddress, uri string) error {
	urlToPersist, err := k.checkValidURI(uri, ctx)
	if err != nil {
		return err
	}
	key := types.GetOSLocatorKey(ownerAddr)
	store := ctx.KVStore(k.storeKey)
	if !store.Has(key) {
		return types.ErrAddressNotBound
	}

	record := types.NewOSLocatorRecord(ownerAddr, encryptionKey, urlToPersist.String())
	bz, err := k.cdc.Marshal(&record)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	k.EmitEvent(ctx, types.NewEventOSLocatorUpdated(record.Owner))
	defer types.GetIncObjFunc(types.TLType_OSLocator, types.TLAction_Updated)
	return nil
}

// ImportOSLocatorRecord binds a name to an address in the kvstore.
// Different from SetOSLocator in that there is less validation here.
// The uri format is not checked, and the owner address account is not looked up.
// This also does not emit any events.
func (k Keeper) ImportOSLocatorRecord(ctx sdk.Context, ownerAddr, encryptionKey sdk.AccAddress, uri string) error {
	key := types.GetOSLocatorKey(ownerAddr)
	store := ctx.KVStore(k.storeKey)
	if store.Has(key) {
		return types.ErrOSLocatorAlreadyBound
	}

	record := types.NewOSLocatorRecord(ownerAddr, encryptionKey, uri)
	bz, err := k.cdc.Marshal(&record)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	defer types.GetIncObjFunc(types.TLType_OSLocator, types.TLAction_Created)
	return nil
}
