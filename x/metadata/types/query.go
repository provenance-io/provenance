package types

// -------------- ScopeWrapper --------------

// WrapScope wraps a scope in a ScopeWrapper and populates the _addr and _uuid fields.
func WrapScope(scope *Scope, includeIDInfo bool) *ScopeWrapper {
	wrapper := ScopeWrapper{}
	if scope != nil {
		wrapper.Scope = scope
		if includeIDInfo {
			wrapper.ScopeIdInfo = GetScopeIDInfo(scope.ScopeId)
			wrapper.ScopeSpecIdInfo = GetScopeSpecIDInfo(scope.SpecificationId)
		}
	}
	return &wrapper
}

// WrapScopeNotFound creates a ScopeWrapper with the _addr and _uuid fields set using the provided MetadataAddress.
func WrapScopeNotFound(scopeAddr MetadataAddress) *ScopeWrapper {
	return &ScopeWrapper{
		ScopeIdInfo: GetScopeIDInfo(scopeAddr),
	}
}

// -------------- SessionWrapper --------------

// WrapSession wraps a session in a SessionWrapper and populates the _addr and _uuid fields.
func WrapSession(session *Session, includeIDInfo bool) *SessionWrapper {
	wrapper := SessionWrapper{}
	if session != nil {
		wrapper.Session = session
		if includeIDInfo {
			wrapper.SessionIdInfo = GetSessionIDInfo(session.SessionId)
			wrapper.ContractSpecIdInfo = GetContractSpecIDInfo(session.SpecificationId)
		}
	}
	return &wrapper
}

// WrapSessionNotFound creates a SessionWrapper with the _addr and _uuid fields set using the provided MetadataAddress.
func WrapSessionNotFound(sessionAddr MetadataAddress) *SessionWrapper {
	return &SessionWrapper{
		SessionIdInfo: GetSessionIDInfo(sessionAddr),
	}
}

// -------------- RecordWrapper --------------

// WrapRecord wraps a record in a RecordWrapper and populates the _addr and _uuid fields.
func WrapRecord(record *Record, includeIDInfo bool) *RecordWrapper {
	wrapper := RecordWrapper{}
	if record != nil {
		wrapper.Record = record
		if includeIDInfo {
			wrapper.RecordIdInfo = GetRecordIDInfo(record.GetRecordAddress())
			wrapper.RecordSpecIdInfo = GetRecordSpecIDInfo(record.SpecificationId)
		}
	}
	return &wrapper
}

// WrapSessionNotFound creates a RecordWrapper with the _addr and _uuid fields set as possible using the provided MetadataAddress.
func WrapRecordNotFound(recordAddr MetadataAddress) *RecordWrapper {
	return &RecordWrapper{
		RecordIdInfo: GetRecordIDInfo(recordAddr),
	}
}

// -------------- ScopeSpecificationWrapper --------------

// WrapScopeSpec wraps a scope specification in a ScopeSpecificationWrapper and populates the _addr and _uuid fields.
func WrapScopeSpec(spec *ScopeSpecification, includeIDInfo bool) *ScopeSpecificationWrapper {
	wrapper := ScopeSpecificationWrapper{}
	if spec != nil {
		wrapper.Specification = spec
		if includeIDInfo {
			wrapper.ScopeSpecIdInfo = GetScopeSpecIDInfo(spec.SpecificationId)
		}
	}
	return &wrapper
}

// WrapScopeSpecNotFound creates a ScopeSpecificationWrapper with the _addr and _uuid fields set using the provided MetadataAddress.
func WrapScopeSpecNotFound(ma MetadataAddress) *ScopeSpecificationWrapper {
	return &ScopeSpecificationWrapper{
		ScopeSpecIdInfo: GetScopeSpecIDInfo(ma),
	}
}

// -------------- ContractSpecificationWrapper --------------

// WrapContractSpec wraps a contract specification in a ContractSpecificationWrapper and populates the _addr and _uuid fields.
func WrapContractSpec(spec *ContractSpecification, includeIDInfo bool) *ContractSpecificationWrapper {
	wrapper := ContractSpecificationWrapper{}
	if spec != nil {
		wrapper.Specification = spec
		if includeIDInfo {
			wrapper.ContractSpecIdInfo = GetContractSpecIDInfo(spec.SpecificationId)
		}
	}
	return &wrapper
}

// WrapContractSpecNotFound creates a ContractSpecificationWrapper with the _addr and _uuid fields set using the provided MetadataAddress.
func WrapContractSpecNotFound(ma MetadataAddress) *ContractSpecificationWrapper {
	return &ContractSpecificationWrapper{
		ContractSpecIdInfo: GetContractSpecIDInfo(ma),
	}
}

// -------------- RecordSpecificationWrapper --------------

// WrapRecordSpec wraps a record specification in a RecordSpecificationWrapper and populates the _addr and _uuid fields.
func WrapRecordSpec(spec *RecordSpecification, includeIDInfo bool) *RecordSpecificationWrapper {
	wrapper := RecordSpecificationWrapper{}
	if spec != nil {
		wrapper.Specification = spec
		if includeIDInfo {
			wrapper.RecordSpecIdInfo = GetRecordSpecIDInfo(spec.SpecificationId)
		}
	}
	return &wrapper
}

func WrapRecordSpecs(specs []*RecordSpecification, includeIDInfo bool) []*RecordSpecificationWrapper {
	retval := make([]*RecordSpecificationWrapper, len(specs))
	for i, s := range specs {
		retval[i] = WrapRecordSpec(s, includeIDInfo)
	}
	return retval
}

// WrapRecordSpecNotFound creates a RecordSpecificationWrapper with the _addr and _uuid fields set using the provided MetadataAddress.
func WrapRecordSpecNotFound(ma MetadataAddress) *RecordSpecificationWrapper {
	return &RecordSpecificationWrapper{
		RecordSpecIdInfo: GetRecordSpecIDInfo(ma),
	}
}
