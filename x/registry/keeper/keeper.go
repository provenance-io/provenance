package keeper

import (
	"errors"
	"fmt"
	"slices"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/registry/types"
)

// RegistryKeeper defines the registry keeper
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	schema   collections.Schema
	Registry collections.Map[string, types.RegistryEntry]

	NFTKeeper      NFTKeeper
	MetadataKeeper MetadataKeeper
}

// NewKeeper returns a new registry Keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, storeService store.KVStoreService, nftKeeper NFTKeeper, metaDataKeeper MetadataKeeper) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	rk := Keeper{
		cdc:      cdc,
		storeKey: storeKey,

		Registry: collections.NewMap(
			sb,
			collections.NewPrefix(registryPrefix),
			"registry",
			collections.StringKey,
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

// Generate a default registry for a given nft key.
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
	return k.Registry.Set(ctx, key.String(), types.RegistryEntry{
		Key:   key,
		Roles: roles,
	})
}

func (k Keeper) GrantRole(ctx sdk.Context, key *types.RegistryKey, role types.RegistryRole, addr []string) error {
	if role == types.RegistryRole_REGISTRY_ROLE_UNSPECIFIED {
		return types.NewErrCodeInvalidRole(role.String())
	}

	keyStr := key.String()

	registryEntry, err := k.Registry.Get(ctx, keyStr)
	if err != nil {
		return types.NewErrCodeRegistryNotFound(fmt.Sprintf("class id: %s, nft id: %s", key.AssetClassId, key.NftId))
	}

	// Return all the addresses that have the role.
	getRoleAddresses := func(role types.RegistryRole) []string {
		for _, roleEntry := range registryEntry.Roles {
			if roleEntry.Role == role {
				return roleEntry.Addresses
			}
		}

		return []string{}
	}
	authorized := getRoleAddresses(role)

	// Determine if any of the new grants are already authorized, and if so error out.
	for _, a := range addr {
		if slices.Contains(authorized, a) {
			return types.NewErrCodeAddressAlreadyHasRole(a, role.String())
		}
	}

	// Append new addresses to the authorized slice
	authorized = append(authorized, addr...)

	// Remove the old role entry from the registry
	updatedRoles := slices.DeleteFunc(registryEntry.Roles, func(s types.RolesEntry) bool {
		return s.Role == role
	})

	// Add the new authorized addresses to the role entry
	updatedRoles = append(updatedRoles, types.RolesEntry{
		Role:      role,
		Addresses: authorized,
	})

	// Update the registry with the new role entries
	registryEntry.Roles = updatedRoles
	if err := k.Registry.Set(ctx, keyStr, registryEntry); err != nil {
		return err
	}

	return nil
}

func (k Keeper) RevokeRole(ctx sdk.Context, key *types.RegistryKey, role types.RegistryRole, addr []string) error {
	keyStr := key.String()

	registryEntry, err := k.Registry.Get(ctx, keyStr)
	if err != nil {
		return types.NewErrCodeRegistryNotFound(fmt.Sprintf("class id: %s, nft id: %s", key.AssetClassId, key.NftId))
	}

	// Find the role entry that we'll be revoking the addresses from.
	var roleToDeleteFrom *types.RolesEntry
	for _, roleEntry := range registryEntry.Roles {
		if roleEntry.Role == role {
			roleToDeleteFrom = &roleEntry
			break
		}
	}

	if roleToDeleteFrom == nil {
		return types.NewErrCodeInvalidRole(role.String())
	}

	// Verify that each address has the role or error out.
	for _, a := range addr {
		if !slices.Contains(roleToDeleteFrom.Addresses, a) {
			return types.NewErrCodeAddressDoesNotHaveRole(a, role.String())
		}
	}

	// Remove the addresses to revoke from the role entry
	roleToDeleteFrom.Addresses = slices.DeleteFunc(roleToDeleteFrom.Addresses, func(s string) bool {
		for _, addrToRevoke := range addr {
			if s == addrToRevoke {
				return true
			}
		}

		return false
	})

	// Save the updated registry entry
	if err := k.Registry.Set(ctx, keyStr, registryEntry); err != nil {
		return err
	}

	return nil
}

func (k Keeper) HasRole(ctx sdk.Context, key *types.RegistryKey, role types.RegistryRole, address string) (bool, error) {
	keyStr := key.String()

	registryEntry, err := k.Registry.Get(ctx, keyStr)
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
			// If we found the role, break out of the loop.
			break
		}
	}

	return false, nil
}

// GetRegistry returns a registry entry for a given key. If the registry entry is not found, it returns nil, nil.
func (k Keeper) GetRegistry(ctx sdk.Context, key *types.RegistryKey) (*types.RegistryEntry, error) {
	keyStr := key.String()

	registryEntry, err := k.Registry.Get(ctx, keyStr)
	if err != nil {
		// Eat the not found error as it is expected, and return nil.
		if errors.Is(err, collections.ErrNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &registryEntry, nil
}

func (k Keeper) InitGenesis(ctx sdk.Context, state *types.GenesisState) {
	for _, entry := range state.Entries {
		if err := k.Registry.Set(ctx, entry.Key.String(), entry); err != nil {
			panic(err) // Genesis should not fail
		}
	}
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genesis := types.GenesisState{}

	registryEntriesIter, err := k.Registry.Iterate(ctx, nil)
	if err != nil {
		panic(err)
	}
	defer registryEntriesIter.Close()

	registryEntries := make([]types.RegistryEntry, 0)
	for ; registryEntriesIter.Valid(); registryEntriesIter.Next() {
		registryEntry, err := registryEntriesIter.Value()
		if err != nil {
			panic(err)
		}

		registryEntries = append(registryEntries, registryEntry)
	}

	genesis.Entries = registryEntries
	return &genesis
}
