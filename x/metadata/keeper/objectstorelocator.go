package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/metadata/types"
)

func (keeper Keeper) GetOsLocatorRecord(ctx sdk.Context, ownerAddress sdk.AccAddress)(osLocator types.ObjectStoreLocator,found  bool) {
	key, err := types.GetOsLocatorAddressKeyPrefix(ownerAddress)
	if err != nil {
		return types.ObjectStoreLocator{}, false
	}
	store := ctx.KVStore(keeper.storeKey)
	b := store.Get(key)
	if b == nil {
		return types.ObjectStoreLocator{}, false
	}
	keeper.cdc.MustUnmarshalBinaryBare(b, &osLocator)
	return osLocator, true
}

// Logger returns a module-specific logger.
func (keeper Keeper) OSLocatorExists(ctx sdk.Context, ownerAddr string) bool {
	 address, err := sdk.AccAddressFromBech32(ownerAddr)
	 if err != nil {
		fmt.Errorf("failed to add locator for a given owner address, invalid address: %s\n", ownerAddr)
	}
	key, err := types.GetOsLocatorAddressKeyPrefix(address)

	if err != nil {
		return false
	}
	store := ctx.KVStore(keeper.storeKey)

	return store.Has(key)
}

// SetNameRecord binds a name to an address. An error is returned if no account exists for the address.
func (keeper Keeper) SetOSLocatorRecord(ctx sdk.Context, ownerAddr sdk.AccAddress, uri string) error {
	var err error

	if account := keeper.authKeeper.GetAccount(ctx, ownerAddr); account == nil {
		return types.ErrInvalidAddress
	}
	key, err := types.GetOsLocatorAddressKeyPrefix(ownerAddr)
	if err != nil {
		return err
	}
	store := ctx.KVStore(keeper.storeKey)
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
