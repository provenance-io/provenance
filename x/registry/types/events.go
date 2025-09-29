package types

// NewEventNFTRegistered creates a new EventNFTRegistered event.
func NewEventNFTRegistered(key *RegistryKey, signer string) *EventNFTRegistered {
	return &EventNFTRegistered{
		AssetClassId: key.AssetClassId,
		NftId:        key.NftId,
		Signer:       signer,
	}
}

// NewEventRoleGranted creates a new EventRoleGranted event.
func NewEventRoleGranted(key *RegistryKey, role RegistryRole, addresses []string, signer string) *EventRoleGranted {
	return &EventRoleGranted{
		AssetClassId: key.AssetClassId,
		NftId:        key.NftId,
		Role:         role,
		Addresses:    addresses,
		Signer:       signer,
	}
}

// NewEventRoleRevoked creates a new EventRoleRevoked event.
func NewEventRoleRevoked(key *RegistryKey, role RegistryRole, addresses []string, signer string) *EventRoleRevoked {
	return &EventRoleRevoked{
		AssetClassId: key.AssetClassId,
		NftId:        key.NftId,
		Role:         role,
		Addresses:    addresses,
		Signer:       signer,
	}
}

// NewEventNFTUnregistered creates a new EventNFTUnregistered event.
func NewEventNFTUnregistered(key *RegistryKey, signer string) *EventNFTUnregistered {
	return &EventNFTUnregistered{
		AssetClassId: key.AssetClassId,
		NftId:        key.NftId,
		Signer:       signer,
	}
}

// NewEventRegistryBulkUpdate creates a new EventRegistryBulkUpdate event.
func NewEventRegistryBulkUpdate(entriesCount uint64, signer string) *EventRegistryBulkUpdate {
	return &EventRegistryBulkUpdate{
		EntriesCount: entriesCount,
		Signer:       signer,
	}
}
