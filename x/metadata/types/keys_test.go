package types

import (
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/require"
)

func TestScopeKey(t *testing.T) {
	scopeUUID := uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")
	sessionUUID := uuid.MustParse("c25c7bd4-c639-4367-a842-f64fa5fccc19")

	scopeKey := ScopeMetadataAddress(scopeUUID)
	// A scope metadata address should have a matching key prefix
	require.EqualValues(t, ScopeKeyPrefix, scopeKey[0:1])

	sessionKey := SessionMetadataAddress(scopeUUID, sessionUUID)
	// A session metadata address should have a matching key prefix
	require.EqualValues(t, SessionKeyPrefix, sessionKey[0:1])
}
