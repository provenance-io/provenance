package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScopeKey(t *testing.T) {
	var scopeKey, groupKey MetadataAddress
	scopeKey = ScopeMetadataAddress(scopeUUID)

	// A scope metadata address should have a matching key prefix
	require.EqualValues(t, ScopeKeyPrefix, scopeKey[0:1])

	groupKey = GroupMetadataAddress(scopeUUID, groupUUID)

	// A group metadata address should have a matching key prefix
	require.EqualValues(t, GroupKeyPrefix, groupKey[0:1])
}
