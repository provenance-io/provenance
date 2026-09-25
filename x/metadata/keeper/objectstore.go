package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// osLocatorKey returns the collection key for an OS Locator entry.
// Matches address.MustLengthPrefix(ownerAddr) — without the 0x21 collection prefix.
func osLocatorKey(ownerAddr sdk.AccAddress) []byte {
	return address.MustLengthPrefix(ownerAddr)
}

// GetOsLocatorRecord Gets the object store locator entry from the kvstore for the given owner address.
func (k Keeper) GetOsLocatorRecord(ctx sdk.Context, ownerAddr sdk.AccAddress) (osLocator types.ObjectStoreLocator, found bool) {
	val, err := k.osLocators.Get(ctx, osLocatorKey(ownerAddr))
	if err != nil {
		return types.ObjectStoreLocator{}, false
	}
	return val, true
}

// OSLocatorExists checks if the provided bech32 owner address has a OSL entry in the kvstore.
func (k Keeper) OSLocatorExists(ctx sdk.Context, ownerAddr sdk.AccAddress) bool {
	has, err := k.osLocators.Has(ctx, osLocatorKey(ownerAddr))
	return err == nil && has
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
	if k.OSLocatorExists(ctx, ownerAddr) {
		return types.ErrOSLocatorAlreadyBound
	}
	record := types.NewOSLocatorRecord(ownerAddr, encryptionKey, urlToPersist.String())
	if err := k.osLocators.Set(ctx, osLocatorKey(ownerAddr), record); err != nil {
		return err
	}
	k.EmitEvent(ctx, types.NewEventOSLocatorCreated(record.Owner))
	return nil
}

// IterateOSLocators runs a function for every ObjectStoreLocator entry in the kvstore.
func (k Keeper) IterateOSLocators(ctx sdk.Context, cb func(account types.ObjectStoreLocator) (stop bool)) error {
	return k.osLocators.Walk(ctx, nil, func(_ []byte, record types.ObjectStoreLocator) (stop bool, err error) {
		return cb(record), nil
	})
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
	if !k.OSLocatorExists(ctx, ownerAddr) {
		return types.ErrAddressNotBound
	}
	if err := k.osLocators.Remove(ctx, osLocatorKey(ownerAddr)); err != nil {
		return err
	}
	k.EmitEvent(ctx, types.NewEventOSLocatorDeleted(ownerAddr.String()))
	return nil
}

// ModifyOSLocator updates an existing os locator entry in the kvstore, returns an error if it doesn't exist.
func (k Keeper) ModifyOSLocator(ctx sdk.Context, ownerAddr, encryptionKey sdk.AccAddress, uri string) error {
	urlToPersist, err := k.checkValidURI(uri, ctx)
	if err != nil {
		return err
	}

	if !k.OSLocatorExists(ctx, ownerAddr) {
		return types.ErrAddressNotBound
	}
	record := types.NewOSLocatorRecord(ownerAddr, encryptionKey, urlToPersist.String())
	if err := k.osLocators.Set(ctx, osLocatorKey(ownerAddr), record); err != nil {
		return err
	}
	k.EmitEvent(ctx, types.NewEventOSLocatorUpdated(record.Owner))
	return nil
}

// ImportOSLocatorRecord binds a name to an address in the kvstore.
// Different from SetOSLocator in that there is less validation here.
// The uri format is not checked, and the owner address account is not looked up.
// This also does not emit any events.
func (k Keeper) ImportOSLocatorRecord(ctx sdk.Context, ownerAddr, encryptionKey sdk.AccAddress, uri string) error {
	if k.OSLocatorExists(ctx, ownerAddr) {
		return types.ErrOSLocatorAlreadyBound
	}
	record := types.NewOSLocatorRecord(ownerAddr, encryptionKey, uri)
	return k.osLocators.Set(ctx, osLocatorKey(ownerAddr), record)
}
