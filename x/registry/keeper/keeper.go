package keeper

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/registry"
)

// RegistryKeeper defines the registry keeper
type RegistryKeeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	schema   collections.Schema

	RegistryEntries collections.Map[string, registry.RegistryEntry]
	Roles           collections.Map[collections.Pair[string, string], registry.RoleAddresses]
}

const (
	registryPrefix = "registry_entries"
	rolesPrefix    = "roles"
)

// NewKeeper returns a new registry Keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, storeService store.KVStoreService) RegistryKeeper {
	sb := collections.NewSchemaBuilder(storeService)

	rk := RegistryKeeper{
		cdc:      cdc,
		storeKey: storeKey,

		RegistryEntries: collections.NewMap(
			sb,
			collections.NewPrefix(registryPrefix),
			registryPrefix,
			collections.StringKey,
			codec.CollValue[registry.RegistryEntry](cdc),
		),
		Roles: collections.NewMap(
			sb,
			collections.NewPrefix(rolesPrefix),
			rolesPrefix,
			collections.PairKeyCodec(collections.StringKey, collections.StringKey),
			codec.CollValue[registry.RoleAddresses](cdc),
		),
	}

	// Build and set the schema
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	rk.schema = schema

	return rk
}

func (k RegistryKeeper) InitGenesis(ctx sdk.Context, state *registry.GenesisState) {
	// Initialize genesis state
}

func (k RegistryKeeper) ExportGenesis(ctx sdk.Context) *registry.GenesisState {
	return &registry.GenesisState{}
}
