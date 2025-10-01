package types

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
