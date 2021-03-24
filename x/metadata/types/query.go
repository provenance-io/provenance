package types

// -------------- ScopeWrapper --------------

// WrapScope wraps a scope in a ScopeWrapper and populates the _addr and _uuid fields.
func WrapScope(scope *Scope) *ScopeWrapper {
	wrapper := ScopeWrapper{}
	if scope != nil {
		wrapper.Scope = scope
		wrapper.SetIDFields(scope.ScopeId)
	}
	return &wrapper
}

// WrapScopeNotFound creates a ScopeWrapper with the _addr and _uuid fields set using the provided MetadataAddress.
func WrapScopeNotFound(scopeAddr MetadataAddress) *ScopeWrapper {
	wrapper := ScopeWrapper{}
	wrapper.SetIDFields(scopeAddr)
	return &wrapper
}

// SetIDFields sets the _addr and _uuid fields (as possible) in the current ScopeWrapper by getting them from the provided MetadataAddress
// If the provided MetadataAddress contains a scope uuid, that will be used to set the scope addr and scope uuid fields.
func (w *ScopeWrapper) SetIDFields(ma MetadataAddress) {
	// This is written to set as much as possible regardless of the MetadataAddress type.
	scopeUUID, scopeUUIDErr := ma.ScopeUUID()
	if scopeUUIDErr == nil {
		w.ScopeUuid = scopeUUID.String()
		w.ScopeAddr = ScopeMetadataAddress(scopeUUID).String()
	}
}

// -------------- SessionWrapper --------------

// WrapSession wraps a session in a SessionWrapper and populates the _addr and _uuid fields.
func WrapSession(session *Session) *SessionWrapper {
	wrapper := SessionWrapper{}
	if session != nil {
		wrapper.Session = session
		wrapper.SetIDFields(session.SessionId)
	}
	return &wrapper
}

// WrapSessionNotFound creates a SessionWrapper with the _addr and _uuid fields set using the provided MetadataAddress.
func WrapSessionNotFound(sessionAddr MetadataAddress) *SessionWrapper {
	wrapper := SessionWrapper{}
	wrapper.SetIDFields(sessionAddr)
	return &wrapper
}

// SetIDFields sets the _addr and _uuid fields (as possible) in the current SessionWrapper by getting them from the provided MetadataAddress
// If the provided MetadataAddress contains a scope uuid, that will be used to set the scope addr and scope uuid fields.
// If it also contains a session uuid, then that and the session uuid will be used to set the session addr and session uuid fields.
func (w *SessionWrapper) SetIDFields(ma MetadataAddress) {
	// This is written to set as much as possible regardless of the MetadataAddress type.
	scopeUUID, scopeUUIDErr := ma.ScopeUUID()
	if scopeUUIDErr == nil {
		w.ScopeUuid = scopeUUID.String()
		w.ScopeAddr = ScopeMetadataAddress(scopeUUID).String()
	}
	sessionUUID, sessionUUIDErr := ma.SessionUUID()
	if sessionUUIDErr == nil {
		w.SessionUuid = sessionUUID.String()
		if scopeUUIDErr == nil {
			w.SessionAddr = SessionMetadataAddress(scopeUUID, sessionUUID).String()
		}
	}
}

// -------------- RecordWrapper --------------

// WrapRecord wraps a record in a RecordWrapper and populates the _addr and _uuid fields.
func WrapRecord(record *Record) *RecordWrapper {
	wrapper := RecordWrapper{}
	if record != nil {
		wrapper.Record = record
		wrapper.SetIDFields(record.SessionId, record.Name)
	}
	return &wrapper
}

// WrapSessionNotFound creates a RecordWrapper with the _addr and _uuid fields set as possible using the provided MetadataAddress.
func WrapRecordNotFound(recordAddr MetadataAddress) *RecordWrapper {
	wrapper := RecordWrapper{}
	wrapper.SetIDFields(recordAddr, "")
	return &wrapper
}

// SetIDFields sets the _addr and _uuid fields (as possible) in the current SessionWrapper using the provided input.
// If the provided MetadataAddress contains a scope uuid, that will be used to set the scope addr and scope uuid fields.
// If it also contains a session uuid, then that and the session uuid will be used to set the session addr and session uuid fields.
// If it is a record MetadataAddress then it will be used directly for the record addr field;
// otherwise, if we've got a scope uuid and a name, they will be used to create and set the record addr.
func (w *RecordWrapper) SetIDFields(ma MetadataAddress, name string) {
	// This is written to try to set as much as possible regardless of the types of MetadataAddresses provided.
	scopeUUID, scopeUUIDErr := ma.ScopeUUID()
	if scopeUUIDErr == nil {
		w.ScopeUuid = scopeUUID.String()
		w.ScopeAddr = ScopeMetadataAddress(scopeUUID).String()
	}
	sessionUUID, sessionUUIDErr := ma.SessionUUID()
	if sessionUUIDErr == nil {
		w.SessionUuid = sessionUUID.String()
		if scopeUUIDErr == nil {
			w.SessionAddr = SessionMetadataAddress(scopeUUID, sessionUUID).String()
		}
	}
	if ma.IsRecordAddress() {
		w.RecordAddr = ma.String()
	} else if scopeUUIDErr == nil && len(name) > 0 {
		w.RecordAddr = RecordMetadataAddress(scopeUUID, name).String()
	}
}

// -------------- ScopeSpecificationWrapper --------------

// WrapScopeSpec wraps a scope specification in a ScopeSpecificationWrapper and populates the _addr and _uuid fields.
func WrapScopeSpec(spec *ScopeSpecification) *ScopeSpecificationWrapper {
	wrapper := ScopeSpecificationWrapper{}
	if spec != nil {
		wrapper.Specification = spec
		wrapper.SetIDFields(spec.SpecificationId)
	}
	return &wrapper
}

// WrapScopeSpecNotFound creates a ScopeSpecificationWrapper with the _addr and _uuid fields set using the provided MetadataAddress.
func WrapScopeSpecNotFound(ma MetadataAddress) *ScopeSpecificationWrapper {
	wrapper := ScopeSpecificationWrapper{}
	wrapper.SetIDFields(ma)
	return &wrapper
}

// SetIDFields sets the _addr and _uuid fields (as possible) in the current ScopeSpecificationWrapper by getting them from the provided MetadataAddress
// If the provided MetadataAddress contains a scope spec uuid, that will be used to set the spec addr and spec uuid fields.
func (w *ScopeSpecificationWrapper) SetIDFields(ma MetadataAddress) {
	// This is written to set as much as possible regardless of the MetadataAddress type.
	specUUID, specUUIDErr := ma.ScopeSpecUUID()
	if specUUIDErr == nil {
		w.SpecificationUuid = specUUID.String()
		w.SpecificationAddr = ScopeSpecMetadataAddress(specUUID).String()
	}
}

// -------------- ContractSpecificationWrapper --------------

// WrapContractSpec wraps a contract specification in a ContractSpecificationWrapper and populates the _addr and _uuid fields.
func WrapContractSpec(spec *ContractSpecification) *ContractSpecificationWrapper {
	wrapper := ContractSpecificationWrapper{}
	if spec != nil {
		wrapper.Specification = spec
		wrapper.SetIDFields(spec.SpecificationId)
	}
	return &wrapper
}

// WrapContractSpecNotFound creates a ContractSpecificationWrapper with the _addr and _uuid fields set using the provided MetadataAddress.
func WrapContractSpecNotFound(ma MetadataAddress) *ContractSpecificationWrapper {
	wrapper := ContractSpecificationWrapper{}
	wrapper.SetIDFields(ma)
	return &wrapper
}

// SetIDFields sets the _addr and _uuid fields (as possible) in the current ContractSpecificationWrapper by getting them from the provided MetadataAddress
// If the provided MetadataAddress contains a contract spec uuid, that will be used to set the spec addr and spec uuid fields.
func (w *ContractSpecificationWrapper) SetIDFields(ma MetadataAddress) {
	// This is written to set as much as possible regardless of the MetadataAddress type.
	specUUID, specUUIDErr := ma.ContractSpecUUID()
	if specUUIDErr == nil {
		w.SpecificationUuid = specUUID.String()
		w.SpecificationAddr = ScopeSpecMetadataAddress(specUUID).String()
	}
}

// -------------- RecordSpecificationWrapper --------------

// WrapRecordSpec wraps a record specification in a RecordSpecificationWrapper and populates the _addr and _uuid fields.
func WrapRecordSpec(spec *RecordSpecification) *RecordSpecificationWrapper {
	wrapper := RecordSpecificationWrapper{}
	if spec != nil {
		wrapper.Specification = spec
		wrapper.SetIDFields(spec.SpecificationId)
	}
	return &wrapper
}

func WrapRecordSpecs(specs []*RecordSpecification) []*RecordSpecificationWrapper {
	retval := make([]*RecordSpecificationWrapper, len(specs))
	for i, s := range specs {
		retval[i] = WrapRecordSpec(s)
	}
	return retval
}

// WrapRecordSpecNotFound creates a RecordSpecificationWrapper with the _addr and _uuid fields set using the provided MetadataAddress.
func WrapRecordSpecNotFound(ma MetadataAddress) *RecordSpecificationWrapper {
	wrapper := RecordSpecificationWrapper{}
	wrapper.SetIDFields(ma)
	return &wrapper
}

// SetIDFields sets the _addr and _uuid fields (as possible) in the current RecordSpecificationWrapper by getting them from the provided MetadataAddress
// If the provided MetadataAddress is a record specification address, that will be used to set the spec addr.
// If the provided MetadataAddress contains a contract spec uuid, that will be used to set the contract spec addr and contract spec uuid fields.
func (w *RecordSpecificationWrapper) SetIDFields(ma MetadataAddress) {
	// This is written to set as much as possible regardless of the MetadataAddress type.
	if ma.IsRecordSpecificationAddress() {
		w.SpecificationAddr = ma.String()
	}
	cSpecUUID, cSpecUUIDErr := ma.ScopeUUID()
	if cSpecUUIDErr == nil {
		w.ContractSpecificationUuid = cSpecUUID.String()
		w.ContractSpecificationAddr = ScopeSpecMetadataAddress(cSpecUUID).String()
	}
}
