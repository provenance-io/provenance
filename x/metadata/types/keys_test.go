package types

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var (
	scopeUUID = uuid.MustParse("c8a159ef-635e-4ac1-810f-8fc6098cfd87")
	groupUUID = uuid.MustParse("a4ebe416-83da-4f23-9f83-26848e77100c")
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
