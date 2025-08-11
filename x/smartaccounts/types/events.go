package types

// NewEventSmartAccountInit creates a new smart account init event
func NewEventSmartAccountInit(address string, credentialCount uint64) *EventSmartAccountInit {
	return &EventSmartAccountInit{
		Address:         address,
		CredentialCount: credentialCount,
	}
}

// NewEventFido2CredentialAdd creates a new EventFido2CredentialAdd instance
func NewEventFido2CredentialAdd(address string, credentialNumber uint64, credentialID string) *EventFido2CredentialAdd {
	return &EventFido2CredentialAdd{
		Address:          address,
		CredentialNumber: credentialNumber,
		CredentialId:     credentialID,
	}
}

// NewEventCosmosCredentialAdd creates a new EventCosmosCredentialAdd instance
func NewEventCosmosCredentialAdd(address string, credentialNumber uint64) *EventCosmosCredentialAdd {
	return &EventCosmosCredentialAdd{
		Address:          address,
		CredentialNumber: credentialNumber,
	}
}

// NewEventCredentialDelete creates a new EventCredentialDelete instance
func NewEventCredentialDelete(address string, credentialNumber uint64) *EventCredentialDelete {
	return &EventCredentialDelete{
		Address:          address,
		CredentialNumber: credentialNumber,
	}
}
