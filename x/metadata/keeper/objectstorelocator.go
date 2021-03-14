package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/provenance-io/provenance/x/metadata/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler is a name record handler function for use with IterateRecords.
type ObjectStoreHandler func(record types.ObjectStoreLocator) error

func (k Keeper) GetOsLocatorRecord(ctx sdk.Context, ownerAddress sdk.AccAddress) (osLocator types.ObjectStoreLocator, found bool) {
	key, err := types.GetOsLocatorAddressKeyPrefix(ownerAddress)
	if err != nil {
		return types.ObjectStoreLocator{}, false
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(key)
	if b == nil {
		return types.ObjectStoreLocator{}, false
	}
	k.cdc.MustUnmarshalBinaryBare(b, &osLocator)
	return osLocator, true
}

// Logger returns a module-specific logger.
func (k Keeper) OSLocatorExists(ctx sdk.Context, ownerAddr string) bool {
	address, err := sdk.AccAddressFromBech32(ownerAddr)
	if err != nil {
		ctx.Logger().Error("failed to get locator", "err", err)
		return false
	}
	key, err := types.GetOsLocatorAddressKeyPrefix(address)

	if err != nil {
		ctx.Logger().Error("failed to get locator", "err", err)
		return false
	}
	store := ctx.KVStore(k.storeKey)

	return store.Has(key)
}

// SetNameRecord binds a name to an address. An error is returned if no account exists for the address.
func (k Keeper) SetOSLocatorRecord(ctx sdk.Context, ownerAddr sdk.AccAddress, uri string) error {
	var err error

	if account := k.authKeeper.GetAccount(ctx, ownerAddr); account == nil {
		return types.ErrInvalidAddress
	}
	key, err := types.GetOsLocatorAddressKeyPrefix(ownerAddr)
	if err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	if store.Has(key) {
		return types.ErrOSLocatorAlreadyBound
	}
	record := types.NewOSLocatorRecord(ownerAddr, uri)
	bz, err := types.ModuleCdc.MarshalBinaryBare(&record)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	// Now index by address
	addrPrefix, err := types.GetOsLocatorAddressKeyPrefix(ownerAddr)
	if err != nil {
		return err
	}
	indexKey := append(addrPrefix, key...) // [0x02] :: [addr-bytes]
	store.Set(indexKey, bz)
	return nil
}

// IterateLocatorsForURI gets address's associated with a given URI.
func (k Keeper) IterateLocatorsForURI(ctx sdk.Context, handler ObjectStoreHandler) error {
	store := ctx.KVStore(k.storeKey)
	prefix := types.OSLocatorAddressKeyPrefix
	it := sdk.KVStorePrefixIterator(store, prefix)

	defer it.Close()

	for ; it.Valid(); it.Next() {
		record := types.ObjectStoreLocator{}
		if err := types.ModuleCdc.UnmarshalBinaryBare(it.Value(), &record); err != nil {
			return err
		}
		if err := handler(record); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) GetOSLocatorByScopeUUID(ctx sdk.Context, scopeID string) (*types.OSLocatorScopeResponse, error) {
	id, err := uuid.Parse(scopeID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid scope uuid: %s", err.Error())
	}
	scopeAddress := types.ScopeMetadataAddress(id)

	s, found := k.GetScope(ctx, scopeAddress)

	if found == false {
		return nil, status.Errorf(codes.InvalidArgument, "invalid scope uuid: %s", err.Error())
	}

	signers := make([]sdk.AccAddress, 0, len(s.Owners))

	for i, p := range s.Owners {
		addr, err := sdk.AccAddressFromBech32(p.Address)
		if err != nil {
			panic(err)
		}
		signers[i] = addr
	}

	locators := make([]types.ObjectStoreLocator, 0, len(signers))
	for i, addr := range signers {
		loc, found := k.GetOsLocatorRecord(ctx, addr)
		if found == false {
			continue
		}
		locators[i] = loc
	}
	return &types.OSLocatorScopeResponse{Locator: locators}, nil
}

// Delete a os locator record from the kvstore.
func (k Keeper) deleteRecord(ctx sdk.Context, ownerAddr sdk.AccAddress) error {
	// Need the record to clear the address index
	_, found := k.GetOsLocatorRecord(ctx, ownerAddr)
	if found == false {
		return types.ErrAddressNotBound
	}

	// Delete the main name record
	key, err := types.GetOsLocatorAddressKeyPrefix(ownerAddr)
	if err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	store.Delete(key)
	// Delete the address index record
	addrPrefix, err := types.GetOsLocatorAddressKeyPrefix(ownerAddr)
	if err != nil {
		return err
	}
	indexKey := append(addrPrefix, key...) // [0x02] :: [addr-bytes]
	if store.Has(indexKey) {
		store.Delete(indexKey)
	}
	return nil
}

func (k Keeper) modifyRecord(ctx sdk.Context, ownerAddr sdk.AccAddress, uri string) error {
	// Need the record to clear the address index
	_, found := k.GetOsLocatorRecord(ctx, ownerAddr)
	if found == false {
		return types.ErrAddressNotBound
	}

	// Delete the main name record
	key, err := types.GetOsLocatorAddressKeyPrefix(ownerAddr)
	if err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	record := types.NewOSLocatorRecord(ownerAddr, uri)
	bz, err := types.ModuleCdc.MarshalBinaryBare(&record)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}
