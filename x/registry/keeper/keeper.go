package keeper

import (
	"errors"
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

	NFTKeeper
	MetadataKeeper
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
		return err
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
	k.Registry.Set(ctx, keyStr, registryEntry)

	return nil
}

func (k Keeper) RevokeRole(ctx sdk.Context, key *types.RegistryKey, role types.RegistryRole, addr []string) error {
	if role == types.RegistryRole_REGISTRY_ROLE_UNSPECIFIED {
		return types.NewErrCodeInvalidRole(role.String())
	}

	keyStr := key.String()

	registryEntry, err := k.Registry.Get(ctx, keyStr)
	if err != nil {
		return err
	}

	// Remove any address from the current slice that is in the addresses to revoke slice
	var updatedAddresses []string
	for _, roleEntry := range registryEntry.Roles {
		// Find the role entry that matches the role to revoke
		if roleEntry.Role == role {
			for _, roleAddr := range roleEntry.Addresses {
				for _, addrToRevoke := range addr {
					// If the address to revoke is the same as the role address, skip it
					if roleAddr == addrToRevoke {
						continue
					}

					updatedAddresses = append(updatedAddresses, roleAddr)
				}
			}

			break
		}
	}

	// Delete the old permissioned addresses from the role entry
	registryEntry.Roles = slices.DeleteFunc(registryEntry.Roles, func(s types.RolesEntry) bool {
		if s.Role == role {
			return true
		}

		return false
	})

	// Add the new permissioned addresses to the role entry
	registryEntry.Roles = append(registryEntry.Roles, types.RolesEntry{
		Role:      role,
		Addresses: updatedAddresses,
	})

	// Save the updated registry entry
	k.Registry.Set(ctx, keyStr, registryEntry)

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
	// Initialize genesis state
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{}
}
