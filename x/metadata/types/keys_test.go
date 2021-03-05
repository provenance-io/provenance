package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScopeKey(t *testing.T) {
	var scopeKey, sessionKey MetadataAddress
	scopeKey = ScopeMetadataAddress(scopeUUID)

	// A scope metadata address should have a matching key prefix
	require.EqualValues(t, ScopeKeyPrefix, scopeKey[0:1])

	sessionKey = SessionMetadataAddress(scopeUUID, sessionUUID)

	// A session metadata address should have a matching key prefix
	require.EqualValues(t, SessionKeyPrefix, sessionKey[0:1])
}
