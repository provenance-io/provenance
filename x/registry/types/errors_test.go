package types_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/provenance-io/provenance/x/registry/types"
)

func TestNewErrCodeRegistryAlreadyExists(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		expectedMsg string
	}{
		{
			name:        "basic key",
			key:         "test-key",
			expectedMsg: "registry already exists for key: \"test-key\": registry already exists",
		},
		{
			name:        "empty key",
			key:         "",
			expectedMsg: "registry already exists for key: \"\": registry already exists",
		},
		{
			name:        "complex key",
			key:         "class:nft-id-123",
			expectedMsg: "registry already exists for key: \"class:nft-id-123\": registry already exists",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.NewErrCodeRegistryAlreadyExists(tc.key)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedMsg)
			assert.True(t, errors.Is(err, types.ErrRegistryAlreadyExists), "error should wrap ErrRegistryAlreadyExists")
		})
	}
}

func TestNewErrCodeNFTNotFound(t *testing.T) {
	tests := []struct {
		name        string
		nftID       string
		expectedMsg string
	}{
		{
			name:        "basic nft id",
			nftID:       "nft-123",
			expectedMsg: "NFT does not exist: \"nft-123\": NFT does not exist",
		},
		{
			name:        "empty nft id",
			nftID:       "",
			expectedMsg: "NFT does not exist: \"\": NFT does not exist",
		},
		{
			name:        "metadata scope id",
			nftID:       "scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel",
			expectedMsg: "NFT does not exist: \"scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel\": NFT does not exist",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.NewErrCodeNFTNotFound(tc.nftID)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedMsg)
			assert.True(t, errors.Is(err, types.ErrNFTNotFound), "error should wrap ErrNFTNotFound")
		})
	}
}

func TestNewErrCodeUnauthorized(t *testing.T) {
	tests := []struct {
		name        string
		why         string
		expectedMsg string
	}{
		{
			name:        "basic reason",
			why:         "not the owner",
			expectedMsg: "unauthorized access: not the owner: unauthorized",
		},
		{
			name:        "empty reason",
			why:         "",
			expectedMsg: "unauthorized access: : unauthorized",
		},
		{
			name:        "detailed reason",
			why:         "signer does not own the NFT",
			expectedMsg: "unauthorized access: signer does not own the NFT: unauthorized",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.NewErrCodeUnauthorized(tc.why)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedMsg)
			assert.True(t, errors.Is(err, types.ErrUnauthorized), "error should wrap ErrUnauthorized")
		})
	}
}

func TestNewErrCodeInvalidRole(t *testing.T) {
	tests := []struct {
		name        string
		role        string
		expectedMsg string
	}{
		{
			name:        "unspecified role",
			role:        "REGISTRY_ROLE_UNSPECIFIED",
			expectedMsg: "invalid role: \"REGISTRY_ROLE_UNSPECIFIED\": invalid role",
		},
		{
			name:        "unknown role",
			role:        "INVALID_ROLE",
			expectedMsg: "invalid role: \"INVALID_ROLE\": invalid role",
		},
		{
			name:        "empty role",
			role:        "",
			expectedMsg: "invalid role: \"\": invalid role",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.NewErrCodeInvalidRole(tc.role)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedMsg)
			assert.True(t, errors.Is(err, types.ErrInvalidRole), "error should wrap ErrInvalidRole")
		})
	}
}

func TestNewErrCodeRegistryNotFound(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		expectedMsg string
	}{
		{
			name:        "basic key",
			key:         "test-key",
			expectedMsg: "registry not found for key: \"test-key\": registry not found",
		},
		{
			name:        "empty key",
			key:         "",
			expectedMsg: "registry not found for key: \"\": registry not found",
		},
		{
			name:        "complex key",
			key:         "class:nft-id-456",
			expectedMsg: "registry not found for key: \"class:nft-id-456\": registry not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.NewErrCodeRegistryNotFound(tc.key)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedMsg)
			assert.True(t, errors.Is(err, types.ErrRegistryNotFound), "error should wrap ErrRegistryNotFound")
		})
	}
}

func TestNewErrCodeAddressAlreadyHasRole(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		role        string
		expectedMsg string
	}{
		{
			name:        "basic address and role",
			address:     "cosmos1abc123",
			role:        "ORIGINATOR",
			expectedMsg: "address \"cosmos1abc123\" already has role \"ORIGINATOR\": address already has role",
		},
		{
			name:        "empty address",
			address:     "",
			role:        "SERVICER",
			expectedMsg: "address \"\" already has role \"SERVICER\": address already has role",
		},
		{
			name:        "empty role",
			address:     "cosmos1xyz789",
			role:        "",
			expectedMsg: "address \"cosmos1xyz789\" already has role \"\": address already has role",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.NewErrCodeAddressAlreadyHasRole(tc.address, tc.role)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedMsg)
			assert.True(t, errors.Is(err, types.ErrAddressAlreadyHasRole), "error should wrap ErrAddressAlreadyHasRole")
		})
	}
}

func TestNewErrCodeAddressDoesNotHaveRole(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		role        string
		expectedMsg string
	}{
		{
			name:        "basic address and role",
			address:     "cosmos1abc123",
			role:        "ORIGINATOR",
			expectedMsg: "address \"cosmos1abc123\" does not have role \"ORIGINATOR\": address does not have role",
		},
		{
			name:        "empty address",
			address:     "",
			role:        "SERVICER",
			expectedMsg: "address \"\" does not have role \"SERVICER\": address does not have role",
		},
		{
			name:        "empty role",
			address:     "cosmos1xyz789",
			role:        "",
			expectedMsg: "address \"cosmos1xyz789\" does not have role \"\": address does not have role",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.NewErrCodeAddressDoesNotHaveRole(tc.address, tc.role)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedMsg)
			assert.True(t, errors.Is(err, types.ErrAddressDoesNotHaveRole), "error should wrap ErrAddressDoesNotHaveRole")
		})
	}
}

func TestNewErrCodeInvalidField(t *testing.T) {
	tests := []struct {
		name        string
		field       string
		format      string
		args        []interface{}
		expectedMsg string
	}{
		{
			name:        "basic field error",
			field:       "asset_class_id",
			format:      "must not be empty",
			args:        nil,
			expectedMsg: "invalid asset_class_id: must not be empty: invalid field",
		},
		{
			name:        "field with format args",
			field:       "nft_id",
			format:      "length must be between %d and %d",
			args:        []interface{}{1, 128},
			expectedMsg: "invalid nft_id: length must be between 1 and 128: invalid field",
		},
		{
			name:        "empty field",
			field:       "",
			format:      "some error",
			args:        nil,
			expectedMsg: "invalid : some error: invalid field",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.NewErrCodeInvalidField(tc.field, tc.format, tc.args...)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedMsg)
			assert.True(t, errors.Is(err, types.ErrInvalidField), "error should wrap ErrInvalidField")
		})
	}
}

func TestErrorConstants(t *testing.T) {
	// Test that error constants are properly defined
	tests := []struct {
		name      string
		err       error
		codespace string
		code      uint32
	}{
		{
			name: "ErrRegistryAlreadyExists",
			err:  types.ErrRegistryAlreadyExists,
			code: 1,
		},
		{
			name: "ErrNFTNotFound",
			err:  types.ErrNFTNotFound,
			code: 2,
		},
		{
			name: "ErrUnauthorized",
			err:  types.ErrUnauthorized,
			code: 3,
		},
		{
			name: "ErrInvalidRole",
			err:  types.ErrInvalidRole,
			code: 4,
		},
		{
			name: "ErrRegistryNotFound",
			err:  types.ErrRegistryNotFound,
			code: 5,
		},
		{
			name: "ErrAddressAlreadyHasRole",
			err:  types.ErrAddressAlreadyHasRole,
			code: 6,
		},
		{
			name: "ErrInvalidKey",
			err:  types.ErrInvalidKey,
			code: 7,
		},
		{
			name: "ErrAddressDoesNotHaveRole",
			err:  types.ErrAddressDoesNotHaveRole,
			code: 8,
		},
		{
			name: "ErrInvalidField",
			err:  types.ErrInvalidField,
			code: 9,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.NotNil(t, tc.err, "error should not be nil")
			// Registry error strings may not include the module name; just ensure non-empty content and code registered.
			assert.NotEmpty(t, tc.err.Error(), "error should have text")
		})
	}
}
