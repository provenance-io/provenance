package types

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
func (w *SessionWrapper) SetIDFields(sessionAddr MetadataAddress) {
	if sessionAddr.IsSessionAddress() {
		w.SessionAddr = sessionAddr.String()
	}
	scopeAddr, err := sessionAddr.AsScopeAddress()
	if err == nil {
		w.ScopeAddr = scopeAddr.String()
	}
	sessionUUID, err := sessionAddr.SessionUUID()
	if err == nil {
		w.SessionUuid = sessionUUID.String()
	}
	scopeUUID, err := sessionAddr.ScopeUUID()
	if err == nil {
		w.ScopeUuid = scopeUUID.String()
	}
}

// WrapRecord wraps a record in a RecordWrapper and populates the _addr and _uuid fields.
func WrapRecord(record *Record) *RecordWrapper {
	wrapper := RecordWrapper{}
	if record != nil {
		wrapper.Record = record
		wrapper.SetIDFields(MetadataAddress{}, record.SessionId, record.Name)
	}
	return &wrapper
}

// WrapSessionNotFound creates a RecordWrapper with the _addr and _uuid fields set as possible using the provided MetadataAddress.
func WrapRecordNotFound(recordAddr MetadataAddress) *RecordWrapper {
	wrapper := RecordWrapper{}
	wrapper.SetIDFields(recordAddr, MetadataAddress{}, "")
	return &wrapper
}

// SetIDFields sets the _addr and _uuid fields (as possible) in the current SessionWrapper using the provided input.
func (w *RecordWrapper) SetIDFields(recordAddr MetadataAddress, sessionAddr MetadataAddress, name string) {
	var err error
	if recordAddr.Empty() && !sessionAddr.Empty() && len(name) > 0 {
		recordAddr, err = sessionAddr.AsRecordAddress(name)
	}
	if err == nil && recordAddr.IsRecordAddress() {
		w.RecordAddr = recordAddr.String()
	}
	if sessionAddr.IsSessionAddress() {
		w.SessionAddr = sessionAddr.String()
	}
	scopeAddr, err := recordAddr.AsScopeAddress()
	if err != nil {
		scopeAddr, err = sessionAddr.AsScopeAddress()
	}
	if err == nil {
		w.ScopeAddr = scopeAddr.String()
	}
	sessionUUID, err := sessionAddr.SessionUUID()
	if err == nil {
		w.SessionUuid = sessionUUID.String()
	}
	scopeUUID, err := scopeAddr.ScopeUUID()
	if err == nil {
		w.ScopeUuid = scopeUUID.String()
	}
}
