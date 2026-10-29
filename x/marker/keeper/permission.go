package keeper

import (
	"errors"
	"fmt"
	"slices"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/marker/types"
)

// GetAccess returns the permissions one address has on a marker (empty if none).
// This is the cheapest permission lookup: one small store read, no marker fetch.
func (k Keeper) GetAccess(ctx sdk.Context, markerAddr, addr sdk.AccAddress) types.AccessList {
	v, err := k.markerPerms.Get(ctx, collections.Join(markerAddr, addr))
	if err != nil {
		return nil
	}
	return v.Permissions
}

// HasAccess returns true if addr has the given permission on the marker.
func (k Keeper) HasAccess(ctx sdk.Context, markerAddr, addr sdk.AccAddress, role types.Access) bool {
	for _, p := range k.GetAccess(ctx, markerAddr, addr) {
		if p == role {
			return true
		}
	}
	return false
}

// SetAccess sets (or deletes, if empty) the permission list for addr on a marker.
func (k Keeper) SetAccess(ctx sdk.Context, markerAddr, addr sdk.AccAddress, perms types.AccessList) error {
	key := collections.Join(markerAddr, addr)
	if len(perms) == 0 {
		err := k.markerPerms.Remove(ctx, key)
		if err != nil && !errors.Is(err, collections.ErrNotFound) {
			return fmt.Errorf("could not remove the permissions that %s has on marker %s: %w", addr, markerAddr, err)
		}
		return nil
	}
	// Store the permissions in a canonical (sorted, deduplicated) order so that a grant always has the
	// same bytes in state regardless of how it was written (e.g. a tx vs a genesis import).
	stored := slices.Clone(perms)
	slices.Sort(stored)
	stored = slices.Compact(stored)
	if err := k.markerPerms.Set(ctx, key, types.MarkerPermissions{Permissions: stored}); err != nil {
		return fmt.Errorf("could not set the permissions that %s has on marker %s: %w", addr, markerAddr, err)
	}
	return nil
}

// RevokeAccessEntry removes all permissions addr has on a marker.
func (k Keeper) RevokeAccessEntry(ctx sdk.Context, markerAddr, addr sdk.AccAddress) error {
	return k.SetAccess(ctx, markerAddr, addr, nil)
}

// GetMarkerAccessList rebuilds the full []AccessGrant for a marker from the perms store.
// GetMarkerAccessList rebuilds the full []AccessGrant for a marker from the perms store.
func (k Keeper) GetMarkerAccessList(ctx sdk.Context, markerAddr sdk.AccAddress) ([]types.AccessGrant, error) {
	var grants []types.AccessGrant
	rng := collections.NewPrefixedPairRange[sdk.AccAddress, sdk.AccAddress](markerAddr)
	err := k.markerPerms.Walk(ctx, rng,
		func(key collections.Pair[sdk.AccAddress, sdk.AccAddress], v types.MarkerPermissions) (bool, error) {
			grants = append(grants, types.AccessGrant{
				Address:     key.K2().String(),
				Permissions: v.Permissions,
			})
			return false, nil
		})
	if err != nil {
		return nil, fmt.Errorf("could not read the permissions on marker %s: %w", markerAddr, err)
	}
	return grants, nil
}

// RemoveAllAccessForMarker deletes every permission entry for a marker.
func (k Keeper) RemoveAllAccessForMarker(ctx sdk.Context, markerAddr sdk.AccAddress) error {
	rng := collections.NewPrefixedPairRange[sdk.AccAddress, sdk.AccAddress](markerAddr)
	if err := k.markerPerms.Clear(ctx, rng); err != nil {
		return fmt.Errorf("could not remove all the permissions on marker %s: %w", markerAddr, err)
	}
	return nil
}

// AddressesWithAccess returns every address holding the given permission on a marker.
func (k Keeper) AddressesWithAccess(ctx sdk.Context, markerAddr sdk.AccAddress, role types.Access) ([]sdk.AccAddress, error) {
	var addrs []sdk.AccAddress
	rng := collections.NewPrefixedPairRange[sdk.AccAddress, sdk.AccAddress](markerAddr)
	err := k.markerPerms.Walk(ctx, rng,
		func(key collections.Pair[sdk.AccAddress, sdk.AccAddress], v types.MarkerPermissions) (bool, error) {
			for _, p := range v.Permissions {
				if p == role {
					addrs = append(addrs, key.K2())
					break
				}
			}
			return false, nil
		})
	if err != nil {
		return nil, fmt.Errorf("could not read the addresses with %s on marker %s: %w", role, markerAddr, err)
	}
	return addrs, nil
}

// GetMarkerWithPerms looks up a marker by address and populates its AccessControl from the
// permission store.
func (k Keeper) GetMarkerWithPerms(ctx sdk.Context, addr sdk.AccAddress) (types.MarkerAccountI, error) {
	m, err := k.GetMarker(ctx, addr)
	if err != nil || m == nil {
		return m, err
	}
	if err := k.PopulateMarkerPerms(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// GetMarkerByDenomWithPerms is the denom variant of GetMarkerWithPerms.
func (k Keeper) GetMarkerByDenomWithPerms(ctx sdk.Context, denom string) (types.MarkerAccountI, error) {
	m, err := k.GetMarkerByDenom(ctx, denom)
	if err != nil || m == nil {
		return m, err
	}
	if err := k.PopulateMarkerPerms(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// PopulateMarkerPerms looks up and sets the marker's AccessControl from the perms store.
func (k Keeper) PopulateMarkerPerms(ctx sdk.Context, ma types.MarkerAccountI) error {
	grants, err := k.GetMarkerAccessList(ctx, ma.GetAddress())
	if err != nil {
		return fmt.Errorf("could not populate the permissions on marker %s: %w", ma.GetAddress(), err)
	}
	ma.SetAccessList(grants)
	return nil
}

// ValidateHasAccess returns an error if addr lacks the given permission on the marker.
func (k Keeper) ValidateHasAccess(ctx sdk.Context, markerAddr sdk.AccAddress, addr sdk.AccAddress, role types.Access) error {
	if k.HasAccess(ctx, markerAddr, addr, role) {
		return nil
	}
	// Match the historical type-method error format for test/tool compatibility.
	denom := ""
	if m, err := k.GetMarker(ctx, markerAddr); err == nil && m != nil {
		denom = m.GetDenom()
	}
	return fmt.Errorf("%s does not have %s on %s marker (%s)", addr, role, denom, markerAddr)
}

// AtLeastOneHasAccess returns true if any of the provided addresses has the given permission on the marker.
func (k Keeper) AtLeastOneHasAccess(ctx sdk.Context, markerAddr sdk.AccAddress, addrs []sdk.AccAddress, role types.Access) bool {
	for _, addr := range addrs {
		if k.HasAccess(ctx, markerAddr, addr, role) {
			return true
		}
	}
	return false
}

// ValidateAtLeastOneHasAccess returns an error if none of the provided addresses has the given permission on the marker.
func (k Keeper) ValidateAtLeastOneHasAccess(ctx sdk.Context, markerAddr sdk.AccAddress, addrs []sdk.AccAddress, role types.Access) error {
	if len(addrs) == 1 {
		return k.ValidateHasAccess(ctx, markerAddr, addrs[0], role)
	}
	if k.AtLeastOneHasAccess(ctx, markerAddr, addrs, role) {
		return nil
	}
	// Only look the marker up (for its denom) once we know we need the error.
	denom := ""
	if m, err := k.GetMarker(ctx, markerAddr); err == nil && m != nil {
		denom = m.GetDenom()
	}
	strs := make([]string, len(addrs))
	for i, addr := range addrs {
		strs[i] = addr.String()
	}
	return fmt.Errorf("none of %q have permission %s on %s marker (%s)", strs, role, denom, markerAddr)
}

// mergeAccess adds the permissions that the grant's address already has on the marker to the grant.
// The grant is updated in place, so afterward, it holds the full set of permissions to store.
func (k Keeper) mergeAccess(ctx sdk.Context, markerAddr sdk.AccAddress, grant types.AccessGrantI) error {
	if err := grant.Validate(); err != nil {
		return err
	}
	existing := k.GetAccess(ctx, markerAddr, grant.GetAddress())
	if len(existing) == 0 {
		return nil
	}
	return grant.MergeAdd(*types.NewAccessGrant(grant.GetAddress(), existing))
}

// validateGrantForMarker returns an error if the grant's permissions are not allowed on the marker.
// These are the same checks that MarkerAccount.Validate applies to each entry in its AccessControl.
func validateGrantForMarker(m types.MarkerAccountI, grant types.AccessGrantI) error {
	perms := grant.GetAccessList()
	if err := types.ValidateGrantsForMarkerType(m.GetMarkerType(), *types.NewAccessGrant(grant.GetAddress(), perms)); err != nil {
		return fmt.Errorf("invalid access privileges granted: %w", err)
	}
	if len(perms) > 0 && grant.GetAddress().Equals(m.GetAddress()) {
		return fmt.Errorf("permissions cannot be granted to '%s' marker account: %v", m.GetDenom(), perms)
	}
	return nil
}

// validateAccessControl checks the marker rules that depend on its permissions, reading the permissions
// from the store (see MarkerAccount.ValidateWithAccessControl). Any addresses in without are treated as
// having no permissions, so this can be used before revoking their access.
func (k Keeper) validateAccessControl(ctx sdk.Context, m types.MarkerAccountI, without ...sdk.AccAddress) error {
	hasOther := func(role types.Access) (bool, error) {
		addrs, err := k.AddressesWithAccess(ctx, m.GetAddress(), role)
		if err != nil {
			return false, err
		}
		for _, addr := range addrs {
			if !slices.ContainsFunc(without, func(w sdk.AccAddress) bool { return w.Equals(addr) }) {
				return true, nil
			}
		}
		return false, nil
	}

	if m.GetStatus() < types.StatusActive && len(m.GetManager()) == 0 {
		hasAdmin, err := hasOther(types.Access_Admin)
		if err != nil {
			return err
		}
		if !hasAdmin {
			return fmt.Errorf("a manager is required if there are no accounts with ACCESS_ADMIN and marker is not ACTIVE")
		}
	}
	if m.GetStatus() == types.StatusFinalized && m.GetSupply().IsZero() {
		hasMint, err := hasOther(types.Access_Mint)
		if err != nil {
			return err
		}
		if !hasMint {
			return fmt.Errorf("cannot create a marker with zero total supply and no authorization for minting more")
		}
	}
	return nil
}
