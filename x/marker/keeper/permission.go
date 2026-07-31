package keeper

import (
	"errors"
	"fmt"

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
			return err
		}
		return nil
	}
	return k.markerPerms.Set(ctx, key, types.MarkerPermissions{Permissions: perms})
}

// RevokeAccessEntry removes all permissions addr has on a marker.
func (k Keeper) RevokeAccessEntry(ctx sdk.Context, markerAddr, addr sdk.AccAddress) error {
	return k.SetAccess(ctx, markerAddr, addr, nil)
}

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
	return grants, err
}

// RemoveAllAccessForMarker deletes every permission entry for a marker.
func (k Keeper) RemoveAllAccessForMarker(ctx sdk.Context, markerAddr sdk.AccAddress) error {
	rng := collections.NewPrefixedPairRange[sdk.AccAddress, sdk.AccAddress](markerAddr)
	return k.markerPerms.Clear(ctx, rng)
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
	return addrs, err
}

// GetMarkerWithPerms returns the marker with its AccessControl populated.
func (k Keeper) GetMarkerWithPerms(ctx sdk.Context, addr sdk.AccAddress) (types.MarkerAccountI, error) {
	m, err := k.GetMarker(ctx, addr)
	if err != nil || m == nil {
		return m, err
	}
	if ma, ok := m.(*types.MarkerAccount); ok {
		if err := k.PopulateMarkerPerms(ctx, ma); err != nil {
			return nil, err
		}
	}
	return m, nil
}

// GetMarkerByDenomWithPerms is the denom variant of GetMarkerWithPerms.
func (k Keeper) GetMarkerByDenomWithPerms(ctx sdk.Context, denom string) (types.MarkerAccountI, error) {
	addr, err := types.MarkerAddress(denom)
	if err != nil {
		return nil, err
	}
	m, err := k.GetMarkerWithPerms(ctx, addr)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, fmt.Errorf("marker %s not found for address: %s", denom, addr)
	}
	return m, nil
}

// PopulateMarkerPerms looks up and sets the marker's AccessControl from the perms store.
func (k Keeper) PopulateMarkerPerms(ctx sdk.Context, ma *types.MarkerAccount) error {
	grants, err := k.GetMarkerAccessList(ctx, ma.GetAddress())
	if err != nil {
		return err
	}
	ma.AccessControl = grants
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

func (k Keeper) ValidateAtLeastOneHasAccess(ctx sdk.Context, markerAddr sdk.AccAddress, addrs []sdk.AccAddress, role types.Access) error {
	if len(addrs) == 1 {
		return k.ValidateHasAccess(ctx, markerAddr, addrs[0], role)
	}
	for _, addr := range addrs {
		if k.HasAccess(ctx, markerAddr, addr, role) {
			return nil
		}
	}
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
