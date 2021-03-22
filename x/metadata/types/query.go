package types

import (
	"fmt"
	"github.com/google/uuid"
)

// WrapSession wraps a session in a SessionWrapper.
func WrapSession(session *Session) *SessionWrapper {
	wrapper := SessionWrapper{}
	if session != nil {
		wrapper.Session = session
		wrapper.SetIDFields(session.SessionId)
	}
	return &wrapper
}

// WrapSessionNotFound creates a SessionWrapper with the data id fields set using the provided MetadataAddress
func WrapSessionNotFound(sessionID MetadataAddress) *SessionWrapper {
	wrapper := SessionWrapper{}
	wrapper.SetIDFields(sessionID)
	return &wrapper
}

func (w *SessionWrapper) SetIDFields(sessionID MetadataAddress) {
	if !sessionID.IsSessionAddress() {
		return
	}
	w.SessionId = sessionID.String()
	sessionUUID, err := sessionID.SessionUUID()
	if err == nil {
		w.SessionUuid = sessionUUID.String()
	}
	scopeID, err := sessionID.AsScopeAddress()
	if err == nil {
		w.ScopeId = scopeID.String()
	}
	scopeUUID, err := sessionID.ScopeUUID()
	if err == nil {
		w.ScopeUuid = scopeUUID.String()
	}
}

// ParseScopeID parses the provided input into a scope MetadataAddress.
// The input can either be a uuid string or scope address bech32 string.
func ParseScopeID(scopeID string) (MetadataAddress, error) {
	addr, addrErr := MetadataAddressFromBech32(scopeID)
	if addrErr == nil {
		if addr.IsScopeAddress() {
			return addr, nil
		}
		return MetadataAddress{}, fmt.Errorf("address [%s] is not a scope address", scopeID)
	}
	uid, uidErr := uuid.Parse(scopeID)
	if uidErr == nil {
		return ScopeMetadataAddress(uid), nil
	}
	return MetadataAddress{}, fmt.Errorf("could not parse [%s] into either a scope address (%s) or uuid (%s)",
		scopeID, addrErr, uidErr)
}

// ParseSessionID parses the provided input into a session MetadataAddress.
// The scopeID field can be either a uuid or scope address bech32 string.
// The sessionID field can be either a uuid or session address bech32 string.
// If the sessionID field is a bech32 address, the scopeID field is ignored.
// Otherwise, the scope id field is parsed using ParseScopeID and converted to a session MetadataAddress using the uuid in the sessionID field.
func ParseSessionID(scopeID string, sessionID string) (MetadataAddress, error) {
	sessionAddr, sessionAddrErr := MetadataAddressFromBech32(sessionID)
	if sessionAddrErr == nil {
		if sessionAddr.IsSessionAddress() {
			return sessionAddr, nil
		}
		return MetadataAddress{}, fmt.Errorf("address [%s] is not a session address", sessionID)
	}
	scopeAddr, scopeAddrErr := ParseScopeID(scopeID)
	if scopeAddrErr != nil {
		return MetadataAddress{}, scopeAddrErr
	}
	sessionUUID, sessionUUIDErr := uuid.Parse(sessionID)
	if sessionUUIDErr == nil {
		return scopeAddr.AsSessionAddress(sessionUUID)
	}
	return MetadataAddress{}, fmt.Errorf("could not parse [%s] into either a session address (%s) or uuid (%s)",
		sessionID, sessionAddrErr, sessionUUIDErr)
}