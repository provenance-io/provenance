package keeper

import (
	"errors"
	"fmt"
	"slices"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/registry/types"
)

// RegistryKeeper defines the registry keeper
type Keeper struct {
	cdc      codec.BinaryCodec
	schema   collections.Schema
	Registry collections.Map[collections.Pair[string, string], types.RegistryEntry]

	NFTKeeper      NFTKeeper
	MetadataKeeper MetadataKeeper
}

// NewKeeper returns a new registry Keeper
func NewKeeper(cdc codec.BinaryCodec, storeService store.KVStoreService, nftKeeper NFTKeeper, metaDataKeeper MetadataKeeper) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	rk := Keeper{
		cdc: cdc,

		Registry: collections.NewMap(
			sb,
			collections.NewPrefix(registryPrefix),
			"registry",
			collections.PairKeyCodec(collections.StringKey, collections.StringKey),
			codec.CollValue[types.RegistryEntry](cdc),
		),

		NFTKeeper:      nftKeeper,
		MetadataKeeper: metaDataKeeper,
	}

	// Build and set the schema
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	rk.schema = schema

	return rk
}

// CreateDefaultRegistry generates a default registry for a given nft key.
func (k Keeper) CreateDefaultRegistry(ctx sdk.Context, ownerAddrStr string, key *types.RegistryKey) error {
	// Set the default roles for originator and servicer.
	roles := make([]types.RolesEntry, 1)
	roles[0] = types.RolesEntry{
		Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
		Addresses: []string{ownerAddrStr},
	}

	return k.CreateRegistry(ctx, key, roles)
}

func (k Keeper) CreateRegistry(ctx sdk.Context, key *types.RegistryKey, roles []types.RolesEntry) error {
	return k.Registry.Set(ctx, collections.Join(key.AssetClassId, key.NftId), types.RegistryEntry{
		Key:   key,
		Roles: roles,
	})
}

func (k Keeper) GrantRole(ctx sdk.Context, key *types.RegistryKey, role types.RegistryRole, addrs []string) error {
	if role == types.RegistryRole_REGISTRY_ROLE_UNSPECIFIED {
		return types.NewErrCodeInvalidRole(role.String())
	}

	registryEntry, err := k.Registry.Get(ctx, collections.Join(key.AssetClassId, key.NftId))
	if err != nil {
		return types.NewErrCodeRegistryNotFound(key.String())
	}

	// Identify which entry has the role we want.
	roleI := -1
	for i, roleEntry := range registryEntry.Roles {
		if roleEntry.Role == role {
			roleI = i
			break
		}
	}
	// If the role doesn't exist yet in the list, add it now.
	if roleI == -1 {
		roleI = len(registryEntry.Roles)
		registryEntry.Roles = append(registryEntry.Roles, types.RolesEntry{Role: role})
	}

	// Determine if any of the new grants are already authorized, and if so error out.
	authorized := registryEntry.Roles[roleI].Addresses
	for _, a := range addrs {
		if slices.Contains(authorized, a) {
			return types.NewErrCodeAddressAlreadyHasRole(a, role.String())
		}
	}

	// Add the addresses to the role.
	registryEntry.Roles[roleI].Addresses = append(registryEntry.Roles[roleI].Addresses, addrs...)

	// Store the updated entry.
	if err = k.Registry.Set(ctx, collections.Join(key.AssetClassId, key.NftId), registryEntry); err != nil {
		return fmt.Errorf("failed to set registry entry: %w", err)
	}

	return nil
}

func (k Keeper) RevokeRole(ctx sdk.Context, key *types.RegistryKey, role types.RegistryRole, addrs []string) error {
	registryEntry, err := k.Registry.Get(ctx, collections.Join(key.AssetClassId, key.NftId))
	if err != nil {
		return types.NewErrCodeRegistryNotFound(key.String())
	}

	// Identify which entry has the role we want.
	roleI := -1
	for i, roleEntry := range registryEntry.Roles {
		if roleEntry.Role == role {
			roleI = i
			break
		}
	}
	if roleI == -1 {
		return types.NewErrCodeInvalidRole(role.String())
	}

	// Make sure all addrs to remove are already granted the role.
	authorized := registryEntry.Roles[roleI].Addresses
	for _, a := range addrs {
		if !slices.Contains(authorized, a) {
			return types.NewErrCodeAddressDoesNotHaveRole(a, role.String())
		}
	}

	// Create a new list of addresss that doesn't have the addrs in it.
	registryEntry.Roles[roleI].Addresses = nil
	for _, addr := range authorized {
		if !slices.Contains(addrs, addr) {
			registryEntry.Roles[roleI].Addresses = append(registryEntry.Roles[roleI].Addresses, addr)
		}
	}

	// If the list is now empty, remove that role entry.
	if len(registryEntry.Roles[roleI].Addresses) == 0 {
		switch roleI {
		case 0:
			registryEntry.Roles = registryEntry.Roles[1:]
		case len(registryEntry.Roles) - 1:
			registryEntry.Roles = registryEntry.Roles[:roleI]
		default:
			registryEntry.Roles = append(registryEntry.Roles[:roleI], registryEntry.Roles[roleI+1:]...)
		}
	}

	// Save the updated registry entry
	if err := k.Registry.Set(ctx, collections.Join(key.AssetClassId, key.NftId), registryEntry); err != nil {
		return err
	}

	return nil
}

func (k Keeper) HasRole(ctx sdk.Context, key *types.RegistryKey, role types.RegistryRole, address string) (bool, error) {
	registryEntry, err := k.Registry.Get(ctx, collections.Join(key.AssetClassId, key.NftId))
	if err != nil {
		return false, err
	}

	// Search to see if the address has the role
	for _, roleEntry := range registryEntry.Roles {
		if roleEntry.Role == role {
			for _, roleAddr := range roleEntry.Addresses {
				if roleAddr == address {
					return true, nil
				}
			}
			// If we found the role (but not the addr), break out of the loop; nothing left to check.
			break
		}
	}

	return false, nil
}

// GetRegistry returns a registry entry for a given key. If the registry entry is not found, it returns nil, nil.
func (k Keeper) GetRegistry(ctx sdk.Context, key *types.RegistryKey) (*types.RegistryEntry, error) {
	registryEntry, err := k.Registry.Get(ctx, collections.Join(key.AssetClassId, key.NftId))
	if err != nil {
		// Eat the not found error as it is expected, and return nil.
		if errors.Is(err, collections.ErrNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &registryEntry, nil
}

// GetRegistries returns the registries paginated
func (k Keeper) GetRegistries(ctx sdk.Context, pagination *query.PageRequest, assetClassID string) ([]types.RegistryEntry, *query.PageResponse, error) {
	var opts []func(opt *query.CollectionsPaginateOptions[collections.Pair[string, string]])
	if len(assetClassID) > 0 {
		opts = append(opts, query.WithCollectionPaginationPairPrefix[string, string](assetClassID))
	}
	ptrs, pageRes, err := query.CollectionPaginate(ctx, k.Registry, pagination, func(_ collections.Pair[string, string], entry types.RegistryEntry) (*types.RegistryEntry, error) {
		return &entry, nil
	}, opts...)
	if err != nil {
		return nil, nil, err
	}

	entries := make([]types.RegistryEntry, len(ptrs))
	for i, p := range ptrs {
		entries[i] = *p
	}
	return entries, pageRes, nil
}
