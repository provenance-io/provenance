package keeper

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/gogoproto/proto"

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

func (k Keeper) EmitEvent(ctx context.Context, tev proto.Message) {
	err := sdk.UnwrapSDKContext(ctx).EventManager().EmitTypedEvent(tev)
	if err != nil {
		// The only reason we'd get an error here is if the event isn't defined correctly in the protos.
		// But we already know all of them are, so we should never see this.
		panic(fmt.Errorf("could not emit typed event %#v: %w", tev, err))
	}
}

// CreateDefaultRegistry generates a default registry for a given nft key.
func (k Keeper) CreateDefaultRegistry(ctx context.Context, ownerAddrStr string, key *types.RegistryKey) error {
	// Set the default roles for originator and servicer.
	roles := make([]types.RolesEntry, 1)
	roles[0] = types.RolesEntry{
		Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
		Addresses: []string{ownerAddrStr},
	}

	return k.CreateRegistry(ctx, key, roles)
}

// CreateRegistry stores a new registry entry in state.
// Returns an error if the registry already exists, or if there's a problem.
func (k Keeper) CreateRegistry(ctx context.Context, key *types.RegistryKey, roles []types.RolesEntry) error {
	// Already exists check
	has, err := k.Registry.Has(ctx, key.CollKey())
	if err != nil {
		return fmt.Errorf("could not check if registry already exists: %w", err)
	}
	if has {
		return types.NewErrCodeRegistryAlreadyExists(key.String())
	}

	err = k.Registry.Set(ctx, key.CollKey(), types.RegistryEntry{
		Key:   key,
		Roles: roles,
	})
	if err != nil {
		return fmt.Errorf("could not set registry entry: %w", err)
	}

	k.EmitEvent(ctx, types.NewEventNFTRegistered(key))
	for _, entry := range roles {
		k.EmitEvent(ctx, types.NewEventRoleGranted(key, entry.Role, entry.Addresses))
	}
	return nil
}

func (k Keeper) DeleteRegistry(ctx context.Context, key *types.RegistryKey) error {
	reg, err := k.GetRegistry(ctx, key)
	if err != nil {
		return fmt.Errorf("could not get registry: %w", err)
	}
	if reg == nil {
		return types.NewErrCodeRegistryNotFound(key.String())
	}

	err = k.Registry.Remove(ctx, key.CollKey())
	if err != nil {
		return fmt.Errorf("error removing registry: %w", err)
	}

	k.EmitEvent(ctx, types.NewEventNFTUnregistered(key))
	for _, entry := range reg.Roles {
		k.EmitEvent(ctx, types.NewEventRoleRevoked(key, entry.Role, entry.Addresses))
	}
	return nil
}

func (k Keeper) GrantRole(ctx context.Context, key *types.RegistryKey, role types.RegistryRole, addrs []string) error {
	if role == types.RegistryRole_REGISTRY_ROLE_UNSPECIFIED {
		return types.NewErrCodeInvalidRole(role.String())
	}

	registryEntry, err := k.Registry.Get(ctx, key.CollKey())
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
	if err = k.Registry.Set(ctx, key.CollKey(), registryEntry); err != nil {
		return fmt.Errorf("failed to set registry entry: %w", err)
	}

	k.EmitEvent(ctx, types.NewEventRoleGranted(key, role, addrs))
	return nil
}

func (k Keeper) RevokeRole(ctx context.Context, key *types.RegistryKey, role types.RegistryRole, addrs []string) error {
	registryEntry, err := k.Registry.Get(ctx, key.CollKey())
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
	if err = k.Registry.Set(ctx, key.CollKey(), registryEntry); err != nil {
		return err
	}

	k.EmitEvent(ctx, types.NewEventRoleRevoked(key, role, addrs))
	return nil
}

func (k Keeper) HasRole(ctx context.Context, key *types.RegistryKey, role types.RegistryRole, address string) (bool, error) {
	registryEntry, err := k.Registry.Get(ctx, key.CollKey())
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

func (k Keeper) ValidateRegistryExists(ctx context.Context, key *types.RegistryKey) error {
	has, err := k.Registry.Has(ctx, key.CollKey())
	if err != nil {
		return fmt.Errorf("error checking if registry exists: %w", err)
	}
	if !has {
		return types.NewErrCodeRegistryNotFound(key.String())
	}
	return nil
}

// GetRegistry returns a registry entry for a given key. If the registry entry is not found, it returns nil, nil.
func (k Keeper) GetRegistry(ctx context.Context, key *types.RegistryKey) (*types.RegistryEntry, error) {
	registryEntry, err := k.Registry.Get(ctx, key.CollKey())
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
func (k Keeper) GetRegistries(ctx context.Context, pagination *query.PageRequest, assetClassID string) ([]types.RegistryEntry, *query.PageResponse, error) {
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
