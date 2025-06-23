package types

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	"github.com/stretchr/testify/require"
)

func TestUtil(t *testing.T) {
	fmt.Println("--- Util StringToAny AnyToString Test ---")
	ir := cdctypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(ir)
	fmt.Println("Registry and Codec created.")

	originalString := "direct check string"
	fmt.Printf("Attempting StringToAny for string: '%s' ...\n", originalString)
	anyMsg, err := StringToAny(originalString)
	if err != nil {
		t.Fatalf("StringToAny failed: %v", err)
	}

	fmt.Printf("Attempting util.AnyToString for TypeURL %s...\n", anyMsg.TypeUrl)
	strMsg, err := AnyToString(cdc, anyMsg)
	require.NoError(t, err, "AnyToString failed")
	require.Equal(t, originalString, strMsg, "Decoded string does not match original")

	fmt.Printf("Unpacking successful. Final string: %q\n", strMsg)
	fmt.Println("--- Registry Direct Check Test Passed ---")
}

func TestAnyToJSONSchema(t *testing.T) {
	ir := cdctypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(ir)

	tests := []struct {
		name        string
		schemaStr   string
		expectError bool
		expected    map[string]interface{}
	}{
		{
			name:        "valid JSON schema",
			schemaStr:   `{"type": "object", "properties": {"name": {"type": "string"}}}`,
			expectError: false,
			expected: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
		{
			name:        "nil any",
			schemaStr:   "",
			expectError: false,
			expected:    nil,
		},
		{
			name:        "invalid JSON schema",
			schemaStr:   `{"type": "object", "properties": {"name": {"type": "string"}}`, // Missing closing brace
			expectError: true,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var anyMsg *cdctypes.Any
			var err error

			if tt.schemaStr == "" {
				anyMsg = nil
			} else {
				anyMsg, err = StringToAny(tt.schemaStr)
				require.NoError(t, err)
			}

			result, err := AnyToJSONSchema(cdc, anyMsg)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.expected == nil {
					require.Nil(t, result)
				} else {
					require.Equal(t, tt.expected, result)
				}
			}
		})
	}
}

func TestValidateJSONSchema(t *testing.T) {
	tests := []struct {
		name        string
		schema      map[string]interface{}
		data        []byte
		expectError bool
	}{
		{
			name: "valid object schema with valid data",
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type": "string",
					},
					"age": map[string]interface{}{
						"type": "integer",
					},
				},
				"required": []string{"name"},
			},
			data:        []byte(`{"name": "John", "age": 30}`),
			expectError: false,
		},
		{
			name: "valid object schema with missing required field",
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type": "string",
					},
				},
				"required": []string{"name"},
			},
			data:        []byte(`{"age": 30}`),
			expectError: true,
		},
		{
			name: "valid object schema with wrong type",
			schema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"age": map[string]interface{}{
						"type": "integer",
					},
				},
			},
			data:        []byte(`{"age": "thirty"}`),
			expectError: true,
		},
		{
			name: "array schema with valid data",
			schema: map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			data:        []byte(`["item1", "item2", "item3"]`),
			expectError: false,
		},
		{
			name: "array schema with invalid data",
			schema: map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			data:        []byte(`["item1", 123, "item3"]`),
			expectError: true,
		},
		{
			name: "string schema with valid data",
			schema: map[string]interface{}{
				"type": "string",
			},
			data:        []byte(`"hello world"`),
			expectError: false,
		},
		{
			name: "string schema with invalid data",
			schema: map[string]interface{}{
				"type": "string",
			},
			data:        []byte(`123`),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJSONSchema(tt.schema, tt.data)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewDefaultMarker(t *testing.T) {
	tests := []struct {
		name        string
		denom       sdk.Coin
		fromAddr    string
		expectError bool
	}{
		{
			name:        "valid marker creation",
			denom:       sdk.NewCoin("testcoin", sdkmath.NewInt(1000000)),
			fromAddr:    "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			expectError: false,
		},
		{
			name:        "invalid from address",
			denom:       sdk.NewCoin("testcoin", sdkmath.NewInt(1000000)),
			fromAddr:    "invalid-address",
			expectError: true,
		},
		{
			name:        "zero coin amount",
			denom:       sdk.NewCoin("testcoin", sdkmath.ZeroInt()),
			fromAddr:    "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marker, err := NewDefaultMarker(tt.denom, tt.fromAddr)

			if tt.expectError {
				require.Error(t, err)
				// Do not require.Nil(marker) because NewDefaultMarker returns a non-nil empty marker on error
			} else {
				require.NoError(t, err)
				require.NotNil(t, marker)
				
				// Verify marker properties
				require.Equal(t, tt.denom.Denom, marker.GetDenom())
				require.Equal(t, tt.denom.Amount, marker.GetSupply().Amount)
				require.Equal(t, markertypes.StatusProposed, marker.GetStatus())
				require.Equal(t, markertypes.MarkerType_RestrictedCoin, marker.GetMarkerType())
				require.True(t, marker.HasFixedSupply())
				require.False(t, marker.HasGovernanceEnabled())
				require.False(t, marker.AllowsForcedTransfer())
				require.Empty(t, marker.GetRequiredAttributes())

				// Verify access grants
				require.Len(t, marker.GetAccessList(), 1)
				accessGrant := marker.GetAccessList()[0]
				require.Equal(t, tt.fromAddr, accessGrant.Address)
				require.Contains(t, accessGrant.Permissions, markertypes.Access_Admin)
				require.Contains(t, accessGrant.Permissions, markertypes.Access_Mint)
				require.Contains(t, accessGrant.Permissions, markertypes.Access_Burn)
				require.Contains(t, accessGrant.Permissions, markertypes.Access_Withdraw)
				require.Contains(t, accessGrant.Permissions, markertypes.Access_Transfer)
			}
		})
	}
}

func TestStringToAnyAndAnyToString_EdgeCases(t *testing.T) {
	ir := cdctypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(ir)

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "empty string",
			input:       "",
			expectError: false,
		},
		{
			name:        "unicode string",
			input:       "Hello 世界 🌍",
			expectError: false,
		},
		{
			name:        "special characters",
			input:       "!@#$%^&*()_+-=[]{}|;':\",./<>?",
			expectError: false,
		},
		{
			name:        "newlines and tabs",
			input:       "line1\nline2\tline3",
			expectError: false,
		},
		{
			name:        "very long string",
			input:       string(make([]byte, 10000)), // 10KB string
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anyMsg, err := StringToAny(tt.input)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, anyMsg)

			// Verify we can convert back
			result, err := AnyToString(cdc, anyMsg)
			require.NoError(t, err)
			require.Equal(t, tt.input, result)
		})
	}
}

func TestAnyToJSONSchema_Integration(t *testing.T) {
	ir := cdctypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(ir)

	// Test the full integration: schema -> any -> json schema -> validation
	schemaStr := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer", "minimum": 0}
		},
		"required": ["name"]
	}`

	// Convert schema string to Any
	anyMsg, err := StringToAny(schemaStr)
	require.NoError(t, err)

	// Convert Any back to JSON schema
	schema, err := AnyToJSONSchema(cdc, anyMsg)
	require.NoError(t, err)
	require.NotNil(t, schema)

	// Test validation with valid data
	validData := []byte(`{"name": "Alice", "age": 25}`)
	err = ValidateJSONSchema(schema, validData)
	require.NoError(t, err)

	// Test validation with invalid data
	invalidData := []byte(`{"age": "not a number"}`)
	err = ValidateJSONSchema(schema, invalidData)
	require.Error(t, err)
}
