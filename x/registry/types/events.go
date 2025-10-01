package types

import (
	"fmt"
	"slices"
)

// NewEventNFTRegistered returns a new EventNFTRegistered.
func NewEventNFTRegistered(key *RegistryKey) *EventNFTRegistered {
	return &EventNFTRegistered{
		NftId:        key.NftId,
		AssetClassId: key.AssetClassId,
	}
}

// NewEventRoleGranted returns a new EventRoleGranted.
func NewEventRoleGranted(key *RegistryKey, role RegistryRole, addrs []string) *EventRoleGranted {
	return &EventRoleGranted{
		NftId:        key.NftId,
		AssetClassId: key.AssetClassId,
		Role:         role.ShortString(),
		Addresses:    addrs,
	}
}

// NewEventRoleRevoked returns a new EventRoleRevoked.
func NewEventRoleRevoked(key *RegistryKey, role RegistryRole, addrs []string) *EventRoleRevoked {
	return &EventRoleRevoked{
		NftId:        key.NftId,
		AssetClassId: key.AssetClassId,
		Role:         role.ShortString(),
		Addresses:    addrs,
	}
}

// NewEventNFTUnregistered returns a new EventNFTUnregistered.
func NewEventNFTUnregistered(key *RegistryKey) *EventNFTUnregistered {
	return &EventNFTUnregistered{
		NftId:        key.NftId,
		AssetClassId: key.AssetClassId,
	}
}

// NewEventRegistryBulkUpdated returns a new EventRegistryBulkUpdated.
func NewEventRegistryBulkUpdated(key *RegistryKey) *EventRegistryBulkUpdated {
	return &EventRegistryBulkUpdated{
		NftId:        key.NftId,
		AssetClassId: key.AssetClassId,
	}
}

// GetChangeEvents gets all the events that represent the changes from oldReg to newReg.
// Panics if they have different keys (unless oldReg or newReg is nil).
func GetChangeEvents(oldReg, newReg *RegistryEntry) ([]*EventRoleGranted, []*EventRoleRevoked) {
	if oldReg == nil {
		if newReg == nil {
			// Nothing to compare, nothing to return.
			return nil, nil
		}
		// All we have is a newReg, so everything is being granted.
		granted := make([]*EventRoleGranted, len(newReg.Roles))
		for i, entry := range newReg.Roles {
			granted[i] = NewEventRoleGranted(newReg.Key, entry.Role, entry.Addresses)
		}
		return granted, nil
	}

	if newReg == nil {
		// There's only an oldReg, so all roles are being revoked.
		revoked := make([]*EventRoleRevoked, len(oldReg.Roles))
		for i, entry := range oldReg.Roles {
			revoked[i] = NewEventRoleRevoked(oldReg.Key, entry.Role, entry.Addresses)
		}
		return nil, revoked
	}

	// We have both and oldReg and newReg.
	// Panic if they have different keys, since we don't know what's actually going on.
	if !oldReg.Key.Equals(newReg.Key) {
		panic(fmt.Errorf("old registry key %q does not equal new registry key %q", oldReg.Key, newReg.Key))
	}

	// Identify all of the grants being made.
	var granted []*EventRoleGranted
	for _, newEntry := range newReg.Roles {
		oldAddrs := oldReg.GetRoleAddrs(newEntry.Role)
		var newAddrs []string
		if len(oldAddrs) == 0 {
			// If there aren't any addresses already listed for this role, they're all new.
			newAddrs = newEntry.Addresses
		} else {
			// Some addresses already have this role, find just the new ones.
			for _, newAddr := range newEntry.Addresses {
				if !slices.Contains(oldAddrs, newAddr) {
					newAddrs = append(newAddrs, newAddr)
				}
			}
		}

		// If we found new addresses, create grant events for them.
		if len(newAddrs) > 0 {
			granted = append(granted, NewEventRoleGranted(newReg.Key, newEntry.Role, newAddrs))
		}
	}

	// Identify all the revokes being done.
	var revoked []*EventRoleRevoked
	for _, oldEntry := range oldReg.Roles {
		newAddrs := newReg.GetRoleAddrs(oldEntry.Role)
		var oldAddrs []string
		if len(newAddrs) == 0 {
			// If there aren't any addresses for this role in the updated registry, they're all being revoked.
			oldAddrs = oldEntry.Addresses
		} else {
			// Some addresses keep this role, find just the ones that were removed.
			for _, oldAddr := range oldEntry.Addresses {
				if !slices.Contains(newAddrs, oldAddr) {
					oldAddrs = append(oldAddrs, oldAddr)
				}
			}
		}

		// If we found some addresses that were removed, create revoke events for them.
		if len(oldAddrs) > 0 {
			revoked = append(revoked, NewEventRoleRevoked(oldReg.Key, oldEntry.Role, oldAddrs))
		}
	}

	return granted, revoked
}
