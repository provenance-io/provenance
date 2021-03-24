package types

func GetScopeIDInfo(scopeID MetadataAddress) *ScopeIdInfo {
	info := scopeID.GetDetails()
	return &ScopeIdInfo{
		ScopeId:          info.Address,
		ScopeIdPrefix:    info.AddressPrefix,
		ScopeIdScopeUuid: info.AddressPrimaryUUID,
		ScopeAddr:        info.Address.String(),
		ScopeUuid:        info.PrimaryUUID,
	}
}

func GetSessionIDInfo(sessionID MetadataAddress) *SessionIdInfo {
	info := sessionID.GetDetails()
	return &SessionIdInfo{
		SessionId:            info.Address,
		SessionIdPrefix:      info.AddressPrefix,
		SessionIdScopeUuid:   info.AddressPrimaryUUID,
		SessionIdSessionUuid: info.AddressSecondaryUUID,
		SessionAddr:          info.Address.String(),
		SessionUuid:          info.SecondaryUUID,
		ScopeIdInfo:          GetScopeIDInfo(info.ParentAddress),
	}
}

func GetRecordIDInfo(recordID MetadataAddress) *RecordIdInfo {
	info := recordID.GetDetails()
	return &RecordIdInfo{
		RecordId:           info.Address,
		RecordIdPrefix:     info.AddressPrefix,
		RecordIdScopeUuid:  info.AddressPrimaryUUID,
		RecordIdHashedName: info.AddressNameHash,
		RecordAddr:         info.Address.String(),
		ScopeIdInfo:        GetScopeIDInfo(info.ParentAddress),
	}
}

func GetScopeSpecIDInfo(scopeSpecID MetadataAddress) *ScopeSpecIdInfo {
	info := scopeSpecID.GetDetails()
	return &ScopeSpecIdInfo{
		ScopeSpecId:              info.Address,
		ScopeSpecIdPrefix:        info.AddressPrefix,
		ScopeSpecIdScopeSpecUuid: info.AddressPrimaryUUID,
		ScopeSpecAddr:            info.Address.String(),
		ScopeSpecUuid:            info.PrimaryUUID,
	}
}

func GetContractSpecIDInfo(contractSpecID MetadataAddress) *ContractSpecIdInfo {
	info := contractSpecID.GetDetails()
	return &ContractSpecIdInfo{
		ContractSpecId:                 info.Address,
		ContractSpecIdPrefix:           info.AddressPrefix,
		ContractSpecIdContractSpecUuid: info.AddressPrimaryUUID,
		ContractSpecAddr:               info.Address.String(),
		ContractSpecUuid:               info.PrimaryUUID,
	}
}

func GetRecordSpecIDInfo(recordSpecID MetadataAddress) *RecordSpecIdInfo {
	info := recordSpecID.GetDetails()
	return &RecordSpecIdInfo{
		RecordSpecId:                 info.Address,
		RecordSpecIdPrefix:           info.AddressPrefix,
		RecordSpecIdContractSpecUuid: info.AddressPrimaryUUID,
		RecordSpecIdHashedName:       info.AddressNameHash,
		RecordSpecAddr:               info.Address.String(),
		ContractSpecIdInfo:           GetContractSpecIDInfo(info.ParentAddress),
	}
}
