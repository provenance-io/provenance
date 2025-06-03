package keeper

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/provenance-io/provenance/x/registry"
)

var _ RegistryKeeper = (*BaseRegistryKeeper)(nil)

type RegistryKeeper interface {
	CreateDefaultRegistry(ctx sdk.Context, authorityAddr sdk.AccAddress, key *registry.RegistryKey) error
	CreateRegistry(ctx sdk.Context, authorityAddr sdk.AccAddress, key *registry.RegistryKey, roles []registry.RolesEntry) error
	GrantRole(ctx sdk.Context, authorityAddr sdk.AccAddress, key *registry.RegistryKey, role string, addr []*sdk.AccAddress) error
	RevokeRole(ctx sdk.Context, authorityAddr sdk.AccAddress, key *registry.RegistryKey, role string, addr []*sdk.AccAddress) error
	HasRole(ctx sdk.Context, key *registry.RegistryKey, role string, address string) (bool, error)
	GetRegistry(ctx sdk.Context, key *registry.RegistryKey) (*registry.RegistryEntry, error)

	// As the registry it is important to have some basic nft information to share as it is a central point for
	// other modules to reference.
	AssetClassExists(ctx sdk.Context, assetClassId *string) bool
	HasNFT(ctx sdk.Context, assetClassId, nftId *string) bool
	GetNFTOwner(ctx sdk.Context, assetClassId, nftId *string) sdk.AccAddress

	InitGenesis(ctx sdk.Context, state *registry.GenesisState)
	ExportGenesis(ctx sdk.Context) *registry.GenesisState
}

// RegistryKeeper defines the registry keeper
type BaseRegistryKeeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	schema   collections.Schema

	Registry collections.Map[string, registry.RegistryEntry]

	NFTKeeper
	MetaDataKeeper
}

const (
	registryPrefix = "registry_entries"

	registryKeyHrp = "reg"
)

// NewKeeper returns a new registry Keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, storeService store.KVStoreService, nftKeeper NFTKeeper, metaDataKeeper MetaDataKeeper) RegistryKeeper {
	sb := collections.NewSchemaBuilder(storeService)

	rk := BaseRegistryKeeper{
		cdc:      cdc,
		storeKey: storeKey,

		Registry: collections.NewMap(
			sb,
			collections.NewPrefix(registryPrefix),
			registryPrefix,
			collections.StringKey,
			codec.CollValue[registry.RegistryEntry](cdc),
		),

		NFTKeeper:      nftKeeper,
		MetaDataKeeper: metaDataKeeper,
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
func (k BaseRegistryKeeper) CreateDefaultRegistry(ctx sdk.Context, authorityAddr sdk.AccAddress, key *registry.RegistryKey) error {
	ownerAddrStr := authorityAddr.String()

	// Set the default roles for originator and servicer.
	roles := make([]registry.RolesEntry, 1)
	roles[0] = registry.RolesEntry{
		Role:      registry.RegistryRole_REGISTRY_ROLE_ORIGINATOR.String(),
		Addresses: []string{ownerAddrStr},
	}

	return k.CreateRegistry(ctx, authorityAddr, key, roles)
}

func (k BaseRegistryKeeper) CreateRegistry(ctx sdk.Context, authorityAddr sdk.AccAddress, key *registry.RegistryKey, roles []registry.RolesEntry) error {
	keyStr, err := RegistryKeyToString(key)
	if err != nil {
		return err
	}

	has, err := k.Registry.Has(ctx, *keyStr)
	if err != nil {
		return fmt.Errorf("registry already exists")
	}
	if has {
		return fmt.Errorf("registry already exists")
	}

	// Verify that an NFT exists for the given key and that the authority owns the NFT
	hasNFT := k.HasNFT(ctx, &key.AssetClassId, &key.NftId)
	if !hasNFT {
		return fmt.Errorf("NFT does not exist")
	}

	nftOwner := k.GetNFTOwner(ctx, &key.AssetClassId, &key.NftId)
	if nftOwner == nil || nftOwner.String() != authorityAddr.String() {
		return fmt.Errorf("authority does not own the NFT")
	}

	k.Registry.Set(ctx, *keyStr, registry.RegistryEntry{
		Key:   key,
		Roles: roles,
	})
	return nil
}

func (k BaseRegistryKeeper) GrantRole(ctx sdk.Context, authorityAddr sdk.AccAddress, key *registry.RegistryKey, role string, addr []*sdk.AccAddress) error {
	keyStr, err := RegistryKeyToString(key)
	if err != nil {
		return err
	}

	has, err := k.Registry.Has(ctx, *keyStr)
	if err != nil {
		return err
	}
	if !has {
		return fmt.Errorf("registry not found")
	}

	// Determine if the authority owns the NFT
	nftOwner := k.GetNFTOwner(ctx, &key.AssetClassId, &key.NftId)
	if nftOwner == nil || nftOwner.String() != authorityAddr.String() {
		return fmt.Errorf("authority does not own the NFT")
	}

	registryEntry, err := k.Registry.Get(ctx, *keyStr)
	if err != nil {
		return err
	}

	// Convert the incoming addresses to strings
	addrStr := make([]string, len(addr))
	for i, a := range addr {
		addrStr[i] = a.String()

		// Look through the registry to see if the address already has the role.
		for _, roleEntry := range registryEntry.Roles {
			if roleEntry.Role == role {
				for _, roleAddr := range roleEntry.Addresses {
					if roleAddr == a.String() {
						return fmt.Errorf("address already has role")
					}
				}
			}
		}
	}

	// Update the registry with the new role
	registryEntry.Roles = append(registryEntry.Roles, registry.RolesEntry{
		Role:      role,
		Addresses: addrStr,
	})

	k.Registry.Set(ctx, *keyStr, registryEntry)

	return nil
}

func (k BaseRegistryKeeper) RevokeRole(ctx sdk.Context, authorityAddr sdk.AccAddress, key *registry.RegistryKey, role string, addr []*sdk.AccAddress) error {
	keyStr, err := RegistryKeyToString(key)
	if err != nil {
		return err
	}

	has, err := k.Registry.Has(ctx, *keyStr)
	if err != nil {
		return err
	}
	if !has {
		return fmt.Errorf("registry not found")
	}

	// Determine if the authority owns the NFT
	nftOwner := k.GetNFTOwner(ctx, &key.AssetClassId, &key.NftId)
	if nftOwner == nil || nftOwner.String() != authorityAddr.String() {
		return fmt.Errorf("authority does not own the NFT")
	}

	registryEntry, err := k.Registry.Get(ctx, *keyStr)
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
					if roleAddr == addrToRevoke.String() {
						continue
					}

					updatedAddresses = append(updatedAddresses, roleAddr)
				}
			}

			break
		}
	}

	// Delete the old permissioned addresses from the role entry
	slices.DeleteFunc(registryEntry.Roles, func(s registry.RolesEntry) bool {
		if s.Role == role {
			return true
		}

		return false
	})

	// Add the new permissioned addresses to the role entry
	registryEntry.Roles = append(registryEntry.Roles, registry.RolesEntry{
		Role:      role,
		Addresses: updatedAddresses,
	})

	// Save the updated registry entry
	k.Registry.Set(ctx, *keyStr, registryEntry)

	return nil
}

func (k BaseRegistryKeeper) HasRole(ctx sdk.Context, key *registry.RegistryKey, role string, address string) (bool, error) {
	keyStr, err := RegistryKeyToString(key)
	if err != nil {
		return false, err
	}

	has, err := k.Registry.Has(ctx, *keyStr)
	if err != nil {
		return false, err
	}
	if !has {
		return false, fmt.Errorf("registry not found")
	}

	registryEntry, err := k.Registry.Get(ctx, *keyStr)
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
func (k BaseRegistryKeeper) GetRegistry(ctx sdk.Context, key *registry.RegistryKey) (*registry.RegistryEntry, error) {
	keyStr, err := RegistryKeyToString(key)
	if err != nil {

		return nil, err
	}

	registryEntry, err := k.Registry.Get(ctx, *keyStr)
	if err != nil {
		// Eat the not found error as it is expected, and return nil.
		if errors.Is(err, collections.ErrNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &registryEntry, nil
}

func (k BaseRegistryKeeper) InitGenesis(ctx sdk.Context, state *registry.GenesisState) {
	// Initialize genesis state
}

func (k BaseRegistryKeeper) ExportGenesis(ctx sdk.Context) *registry.GenesisState {
	return &registry.GenesisState{}
}

// Combine the asset class id and nft id into a bech32 string.
// Using bech32 here just allows us a readable identifier for the registry.
func RegistryKeyToString(key *registry.RegistryKey) (*string, error) {
	joined := strings.Join([]string{key.AssetClassId, key.NftId}, ":")

	b32, err := bech32.ConvertAndEncode(registryKeyHrp, []byte(joined))
	if err != nil {
		return nil, err
	}

	return &b32, nil
}

func StringToRegistryKey(s string) (*registry.RegistryKey, error) {
	hrp, b, err := bech32.DecodeAndConvert(s)
	if err != nil {
		return nil, err
	}

	if hrp != registryKeyHrp {
		return nil, fmt.Errorf("invalid hrp: %s", hrp)
	}

	parts := strings.Split(string(b), ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid key: %s", s)
	}

	return &registry.RegistryKey{
		AssetClassId: parts[0],
		NftId:        parts[1],
	}, nil
}
