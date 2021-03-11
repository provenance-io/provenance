package cmd_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/provenance-io/provenance/cmd/provenanced/cmd"
	"github.com/provenance-io/provenance/x/metadata/types"
)

func TestAddMetaAddressParser(t *testing.T) {
	scopeUUID := uuid.New()
	scopeID := types.ScopeMetadataAddress(scopeUUID)
	sessionUUID := uuid.New()
	sessionID := types.SessionMetadataAddress(scopeUUID, sessionUUID)
	recordID := types.RecordMetadataAddress(scopeUUID, "this is a name")
	contractSpecUUID := uuid.New()
	contractSpecID := types.ContractSpecMetadataAddress(contractSpecUUID)
	scopeSpecUUID := uuid.New()
	scopeSpecID := types.ScopeSpecMetadataAddress(scopeSpecUUID)

	tests := []struct {
		name      string
		addr      string
		expected  string
		expectErr bool
	}{
		{
			name:      "test not an address",
			addr:      "not an id",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "test scope address",
			addr:      scopeID.String(),
			expected:  fmt.Sprintf("Type: Scope\n\nScope UUID: %s\n", scopeUUID),
			expectErr: false,
		},
		{
			name:      "test session address",
			addr:      sessionID.String(),
			expected:  fmt.Sprintf("Type: Session\n\nScope Id: %s\nScope UUID: %s\nSession UUID: %s\n", scopeID, scopeUUID, sessionUUID),
			expectErr: false,
		},
		{
			name:      "test record address",
			addr:      recordID.String(),
			expected:  fmt.Sprintf("Type: Record\n\nScope Id: %s\nScope UUID: %s\n", scopeID, scopeUUID),
			expectErr: false,
		},
		{
			name:      "test contract spec id",
			addr:      contractSpecID.String(),
			expected:  fmt.Sprintf("Type: Contract Specification\n\nContract Specification UUID: %s\n", contractSpecUUID),
			expectErr: false,
		},
		{
			name:      "test scope specification address",
			addr:      scopeSpecID.String(),
			expected:  fmt.Sprintf("Type: Scope Specification\n\nScope Specification UUID: %s\n", scopeSpecUUID),
			expectErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			command := cmd.AddMetaAddressParser()
			command.SetArgs([]string{
				"parse", tc.addr})
			b := bytes.NewBufferString("")
			command.SetOut(b)
			if tc.expectErr {
				require.Error(t, command.Execute())
			} else {
				require.NoError(t, command.Execute())
				out, err := ioutil.ReadAll(b)
				require.NoError(t, err)
				require.Equal(t, tc.expected, string(out))
			}
		})
	}
}

func TestAddMetaAddressEncoder(t *testing.T) {
	scopeUUID := uuid.New()
	scopeID := types.ScopeMetadataAddress(scopeUUID)
	sessionUUID := uuid.New()
	sessionID := types.SessionMetadataAddress(scopeUUID, sessionUUID)
	recordName := "this is a name"
	recordID := types.RecordMetadataAddress(scopeUUID, recordName)
	contractSpecUUID := uuid.New()
	contractSpecID := types.ContractSpecMetadataAddress(contractSpecUUID)
	scopeSpecUUID := uuid.New()
	scopeSpecID := types.ScopeSpecMetadataAddress(scopeSpecUUID)

	tests := []struct {
		name      string
		args      []string
		expected  string
		expectErr bool
	}{
		{
			name:      "test scope address",
			args:      []string{"encode", "scope", scopeUUID.String()},
			expected:  scopeID.String(),
			expectErr: false,
		},
		{
			name:      "test scope address too many args",
			args:      []string{"encode", "scope", scopeUUID.String(), scopeUUID.String()},
			expected:  "",
			expectErr: true,
		},
		{
			name:      "test scope address invalid uuid",
			args:      []string{"encode", "scope", "not an id"},
			expected:  "",
			expectErr: true,
		},
		{
			name:      "test session address",
			args:      []string{"encode", "session", scopeUUID.String(), sessionUUID.String()},
			expected:  sessionID.String(),
			expectErr: false,
		},
		{
			name:      "test session address too few args",
			args:      []string{"encode", "session", scopeUUID.String()},
			expected:  "",
			expectErr: true,
		},
		{
			name:      "test session address too many args",
			args:      []string{"encode", "session", scopeUUID.String(), sessionUUID.String(), sessionUUID.String()},
			expected:  "",
			expectErr: true,
		},
		{
			name:      "test session address invalid first uuid",
			args:      []string{"encode", "session", "not a uuid", sessionUUID.String()},
			expected:  "",
			expectErr: true,
		},
		{
			name:      "test session address invalid second uuid",
			args:      []string{"encode", "session", scopeUUID.String(), "not a uuid"},
			expected:  "",
			expectErr: true,
		},
		{
			name:      "test record address",
			args:      []string{"encode", "record", scopeUUID.String(), recordName},
			expected:  recordID.String(),
			expectErr: false,
		},
		{
			name:      "test record address too few args",
			args:      []string{"encode", "record", scopeUUID.String()},
			expected:  "",
			expectErr: true,
		},
		{
			name:      "test record address too many args",
			args:      []string{"encode", "record", scopeUUID.String(), recordName, recordName},
			expected:  "",
			expectErr: true,
		},
		{
			name:      "test record address invalid uuid",
			args:      []string{"encode", "record", "not a uuid", recordName},
			expected:  "",
			expectErr: true,
		},
		{
			name:      "test contract spec id",
			args:      []string{"encode", "contract-specification", contractSpecUUID.String()},
			expected:  contractSpecID.String(),
			expectErr: false,
		},
		{
			name:      "test contract spec id too many args",
			args:      []string{"encode", "contract-specification", contractSpecUUID.String(), contractSpecUUID.String()},
			expected:  "",
			expectErr: true,
		},
		{
			name:      "test contract spec id invalid uuid",
			args:      []string{"encode", "contract-specification", "not a uuid"},
			expected:  "",
			expectErr: true,
		},
		{
			name:      "test scope specification address",
			args:      []string{"encode", "scope-specification", scopeSpecUUID.String()},
			expected:  scopeSpecID.String(),
			expectErr: false,
		},
		{
			name:      "test scope specification address too many args",
			args:      []string{"encode", "scope-specification", scopeSpecUUID.String(), scopeSpecUUID.String()},
			expected:  "",
			expectErr: true,
		},
		{
			name:      "test scope specification address invalid uuid",
			args:      []string{"encode", "scope-specification", "not a uuid"},
			expected:  "",
			expectErr: true,
		},
		{
			name:      "test scope invalid type",
			args:      []string{"encode", "invalid type", scopeUUID.String()},
			expected:  "",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			command := cmd.AddMetaAddressEncoder()
			command.SetArgs(tc.args)
			b := bytes.NewBufferString("")
			command.SetOut(b)
			if tc.expectErr {
				require.Error(t, command.Execute())
			} else {
				require.NoError(t, command.Execute())
				out, err := ioutil.ReadAll(b)
				require.NoError(t, err)
				require.Equal(t, tc.expected, string(out))
			}
		})
	}
}
