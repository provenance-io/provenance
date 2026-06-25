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

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/provenance-io/provenance/x/registry/types"
)

// Keeper defines the registry keeper.
type Keeper struct {
	cdc                codec.BinaryCodec
	schema             collections.Schema
	Registry           collections.Map[collections.Pair[string, string], types.RegistryEntry]
	PendingRoleChanges collections.Map[string, types.PendingRoleChange]
	RegistryClasses    collections.Map[string, types.RegistryClass]
	Params             collections.Item[types.Params]

	authority string

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

		PendingRoleChanges: collections.NewMap(
			sb,
			collections.NewPrefix(pendingRoleChangePrefix),
			"pending_role_changes",
			collections.StringKey,
			codec.CollValue[types.PendingRoleChange](cdc),
		),

		RegistryClasses: collections.NewMap(
			sb,
			collections.NewPrefix(registryClassPrefix),
			"registry_classes",
			collections.StringKey,
			codec.CollValue[types.RegistryClass](cdc),
		),

		Params: collections.NewItem(
			sb,
			collections.NewPrefix(paramsPrefix),
			"params",
			codec.CollValue[types.Params](cdc),
		),

		authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),

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

// EmitEvent emits an event.
func (k Keeper) EmitEvent(ctx context.Context, tev proto.Message) {
	err := sdk.UnwrapSDKContext(ctx).EventManager().EmitTypedEvent(tev)
	if err != nil {
		// The only reason we'd get an error here is if the event isn't defined correctly in the protos.
		// But we already know all of them are, so we should never see this.
		panic(fmt.Errorf("could not emit typed event %#v: %w", tev, err))
	}
}

// EmitEvents emits multiple events at once.
func (k Keeper) EmitEvents(ctx context.Context, tevs ...proto.Message) {
	err := sdk.UnwrapSDKContext(ctx).EventManager().EmitTypedEvents(tevs...)
	if err != nil {
		// The only reason we'd get an error here is if the event isn't defined correctly in the protos.
		// But we already know all of them are, so we should never see this.
		panic(fmt.Errorf("could not emit typed events %#v: %w", tevs, err))
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

	return k.CreateRegistry(ctx, key, roles, "")
}

// CreateRegistry stores a new registry entry in state.
// Returns an error if the registry already exists, or if there's a problem.
func (k Keeper) CreateRegistry(ctx context.Context, key *types.RegistryKey, roles []types.RolesEntry, registryClassID string) error {
	if key == nil {
		return fmt.Errorf("registry key must not be nil")
	}

	// Defense in depth: enforce the registry class invariant at the keeper layer so invalid state
	// cannot be introduced via direct keeper calls (other modules/tests) that bypass the msg and
	// genesis validation paths. When set, the class must exist and be scoped to this entry's asset
	// class; otherwise authorization would later resolve against the wrong policy tier.
	if err := k.validateRegistryClassForEntry(ctx, registryClassID, key.AssetClassId); err != nil {
		return err
	}

	// Already exists check
	has, err := k.Registry.Has(ctx, key.CollKey())
	if err != nil {
		return fmt.Errorf("could not check if registry already exists: %w", err)
	}
	if has {
		return types.NewErrCodeRegistryAlreadyExists(key.String())
	}

	err = k.Registry.Set(ctx, key.CollKey(), types.RegistryEntry{
		Key:             key,
		Roles:           roles,
		RegistryClassId: registryClassID,
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

// roleChangeSigners resolves the RoleSigner entries describing who authorized a role change for the
// given role. For a policy-governed role it returns the satisfying authorization path's signers; for
// a non-policy role it returns the NFT owner among the approvers (the legacy fallback). before is the
// pre-mutation entry, newAddrs the addresses being assigned to the role, and approvers the signers.
func (k Keeper) roleChangeSigners(ctx context.Context, before *types.RegistryEntry, role types.RegistryRole, newAddrs, approvers []string) []types.RoleSigner {
	roleAuths := k.roleAuthorizationsForEntry(ctx, before)
	if roleAuth, ok := roleAuths[role]; ok {
		return k.CollectSatisfyingSigners(ctx, roleAuth, before, newAddrs, approvers)
	}
	for _, a := range approvers {
		if k.ValidateNFTOwner(ctx, &before.Key.AssetClassId, &before.Key.NftId, a) == nil {
			return []types.RoleSigner{types.NewNFTOwnerSigner(a)}
		}
	}
	return nil
}

// emitRoleUpdated emits the comprehensive EventRoleUpdated for a single role, reading the
// post-mutation addresses from state. before is the pre-mutation entry (source of previous
// addresses and registry class id), and signers is the authorizing signer set.
func (k Keeper) emitRoleUpdated(ctx context.Context, before *types.RegistryEntry, role types.RegistryRole, signers []types.RoleSigner) {
	previous := before.GetRoleAddrs(role)
	var current []string
	after, err := k.GetRegistry(ctx, before.Key)
	if err != nil {
		// The mutation has already been committed; a read-back error means the store is
		// inconsistent. Fail fast rather than emit a misleading audit event with empty addresses.
		panic(fmt.Errorf("could not read back registry %q for EventRoleUpdated: %w", before.Key.String(), err))
	}
	// after may legitimately be nil if the role change removed the entry's last role and the entry
	// was cleaned up; in that case current is correctly empty.
	if after != nil {
		current = after.GetRoleAddrs(role)
	}
	k.EmitEvent(ctx, types.NewEventRoleUpdated(before.Key, before.RegistryClassId, role, previous, current, signers))
}

func (k Keeper) HasRole(ctx context.Context, key *types.RegistryKey, role types.RegistryRole, address string) (bool, error) {
	registryEntry, err := k.Registry.Get(ctx, key.CollKey())
	if err != nil {
		return false, err
	}

	// Search to see if the address has the role
	for _, roleEntry := range registryEntry.Roles {
		if roleEntry.Role == role {
			if slices.Contains(roleEntry.Addresses, address) {
				return true, nil
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

// SetRegistry sets a registry entry for a given key.
func (k Keeper) SetRegistry(ctx context.Context, entry types.RegistryEntry) error {
	return k.Registry.Set(ctx, entry.Key.CollKey(), entry)
}

// SetRoles atomically sets the desired state for one or more roles on a registry entry.
// Each RoleUpdate specifies a role and the complete desired set of addresses; an empty
// address list clears the role entirely.
func (k Keeper) SetRoles(ctx context.Context, key *types.RegistryKey, roleUpdates []types.RoleUpdate) error {
	entry, err := k.Registry.Get(ctx, key.CollKey())
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.NewErrCodeRegistryNotFound(key.String())
		}
		return fmt.Errorf("failed to get registry entry: %w", err)
	}

	// Snapshot the original entry to compute diff events. The loop below mutates entry.Roles
	// (and individual Addresses slices) in place, so a shallow copy would corrupt the snapshot
	// and produce incorrect diffs. Deep-copy the Roles slice and each Addresses slice.
	origCopy := entry
	origCopy.Roles = make([]types.RolesEntry, len(entry.Roles))
	for i, re := range entry.Roles {
		origCopy.Roles[i] = types.RolesEntry{Role: re.Role, Addresses: slices.Clone(re.Addresses)}
	}

	for _, update := range roleUpdates {
		roleI := -1
		for i, re := range entry.Roles {
			if re.Role == update.Role {
				roleI = i
				break
			}
		}

		if len(update.Addresses) == 0 {
			// Clear the role if present.
			if roleI >= 0 {
				entry.Roles = append(entry.Roles[:roleI], entry.Roles[roleI+1:]...)
			}
		} else {
			// Set the complete desired address list for the role.
			if roleI >= 0 {
				entry.Roles[roleI].Addresses = update.Addresses
			} else {
				entry.Roles = append(entry.Roles, types.RolesEntry{Role: update.Role, Addresses: update.Addresses})
			}
		}
	}

	if err = k.Registry.Set(ctx, key.CollKey(), entry); err != nil {
		return fmt.Errorf("failed to set registry entry: %w", err)
	}

	grantEvents, revokeEvents := types.GetChangeEvents(&origCopy, &entry)
	for _, tev := range grantEvents {
		k.EmitEvent(ctx, tev)
	}
	for _, tev := range revokeEvents {
		k.EmitEvent(ctx, tev)
	}
	return nil
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
