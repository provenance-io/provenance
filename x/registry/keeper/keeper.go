package keeper

import (
	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/registry"
)

const RegistryEntryPrefix = "registry_entry"

type RegistryKeeper struct {
	cdc      codec.Codec
	storeKey storetypes.StoreKey
	memKey   storetypes.StoreKey
}

func NewKeeper(
	cdc codec.Codec,
	storeKey,
	memKey storetypes.StoreKey,
) RegistryKeeper {
	return RegistryKeeper{
		cdc:      cdc,
		storeKey: storeKey,
		memKey:   memKey,
	}
}

func (k RegistryKeeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+registry.ModuleName)
}

// SetRegistryEntry sets a registry entry
func (k RegistryKeeper) SetRegistryEntry(ctx sdk.Context, entry registry.RegistryEntry) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&entry)
	store.Set(registry.GetRegistryEntryKey(entry.Address), bz)
}

// GetRegistryEntry returns a registry entry
func (k RegistryKeeper) GetRegistryEntry(ctx sdk.Context, address string) (registry.RegistryEntry, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRegistryEntryKey(address))
	if bz == nil {
		return registry.RegistryEntry{}, false
	}
	var entry types.RegistryEntry
	k.cdc.MustUnmarshal(bz, &entry)
	return entry, true
}

// GetAllRegistryEntries returns all registry entries
func (k RegistryKeeper) GetAllRegistryEntries(ctx sdk.Context) []registry.RegistryEntry {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, []byte(RegistryEntryPrefix))
	iterator := prefixStore.Iterator(nil, nil)
	defer iterator.Close()

	var entries []RegistryEntry
	for ; iterator.Valid(); iterator.Next() {
		var entry registry.RegistryEntry
		k.cdc.MustUnmarshal(iterator.Value(), &entry)
		entries = append(entries, entry)
	}
	return entries
}

// InitGenesis initializes the genesis state
func (k RegistryKeeper) InitGenesis(ctx sdk.Context, state *types.GenesisState) {
	// Initialize genesis state
}

// ExportGenesis exports the genesis state
func (k RegistryKeeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{}
}
