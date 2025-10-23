package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/provenance-io/provenance/x/registry/types"
)

func TestRegistryKey_String(t *testing.T) {
	tests := []struct {
		name           string
		key            types.RegistryKey
		expectedPrefix string
	}{
		{
			name: "basic key",
			key: types.RegistryKey{
				AssetClassId: "test-class",
				NftId:        "test-nft",
			},
			expectedPrefix: "reg1",
		},
		{
			name: "empty values",
			key: types.RegistryKey{
				AssetClassId: "",
				NftId:        "",
			},
			expectedPrefix: "reg1",
		},
		{
			name: "special characters",
			key: types.RegistryKey{
				AssetClassId: "class-with-dashes",
				NftId:        "nft.with.dots",
			},
			expectedPrefix: "reg1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.key.String()
			assert.NotEmpty(t, result, "String() should not return empty")
			assert.True(t, strings.HasPrefix(result, tc.expectedPrefix), "should have reg1 prefix")
		})
	}
}

func TestRegistryKey_Validate(t *testing.T) {
	tests := []struct {
		name   string
		key    *types.RegistryKey
		expErr string
	}{
		{
			name:   "nil key",
			key:    nil,
			expErr: "registry key cannot be nil",
		},
		{
			name: "valid key",
			key: &types.RegistryKey{
				AssetClassId: "test-class",
				NftId:        "test-nft",
			},
			expErr: "",
		},
		{
			name: "empty asset class id",
			key: &types.RegistryKey{
				AssetClassId: "",
				NftId:        "test-nft",
			},
			expErr: "must be between",
		},
		{
			name: "empty nft id",
			key: &types.RegistryKey{
				AssetClassId: "test-class",
				NftId:        "",
			},
			expErr: "must be between",
		},
		{
			name: "too long asset class id",
			key: &types.RegistryKey{
				AssetClassId: strings.Repeat("a", types.MaxLenAssetClassID+1),
				NftId:        "test-nft",
			},
			expErr: "must be between",
		},
		{
			name: "too long nft id",
			key: &types.RegistryKey{
				AssetClassId: "test-class",
				NftId:        strings.Repeat("a", types.MaxLenNFTID+1),
			},
			expErr: "must be between",
		},
		{
			name: "max length asset class id",
			key: &types.RegistryKey{
				AssetClassId: strings.Repeat("a", types.MaxLenAssetClassID),
				NftId:        "test-nft",
			},
			expErr: "",
		},
		{
			name: "max length nft id",
			key: &types.RegistryKey{
				AssetClassId: "test-class",
				NftId:        strings.Repeat("a", types.MaxLenNFTID),
			},
			expErr: "",
		},
		{
			name: "asset class id with whitespace",
			key: &types.RegistryKey{
				AssetClassId: "test class",
				NftId:        "test-nft",
			},
			expErr: "\"test class\" must only contain alphanumeric, '-', '.' characters",
		},
		{
			name: "nft id with whitespace",
			key: &types.RegistryKey{
				AssetClassId: "test-class",
				NftId:        "test nft",
			},
			expErr: "\"test nft\" must only contain alphanumeric, '-', '.' characters",
		},
		{
			name: "valid metadata scope spec",
			key: &types.RegistryKey{
				AssetClassId: "scopespec1qsm47umrdacx2hmnwpjkxh6lta0sz95anh",
				NftId:        "test-nft",
			},
			expErr: "",
		},
		{
			name: "valid metadata scope",
			key: &types.RegistryKey{
				AssetClassId: "test-class",
				NftId:        "scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel",
			},
			expErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.key.Validate()
			if tc.expErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRegistryEntry_Validate(t *testing.T) {
	validAddr := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"

	tests := []struct {
		name   string
		entry  *types.RegistryEntry
		expErr string
	}{
		{
			name:   "nil entry",
			entry:  nil,
			expErr: "registry entry cannot be nil",
		},
		{
			name: "valid entry",
			entry: &types.RegistryEntry{
				Key: &types.RegistryKey{
					AssetClassId: "test-class",
					NftId:        "test-nft",
				},
				Roles: []types.RolesEntry{
					{
						Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
						Addresses: []string{validAddr},
					},
				},
			},
			expErr: "",
		},
		{
			name: "nil key",
			entry: &types.RegistryEntry{
				Key: nil,
				Roles: []types.RolesEntry{
					{
						Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
						Addresses: []string{validAddr},
					},
				},
			},
			expErr: "registry key cannot be nil",
		},
		{
			name: "empty roles",
			entry: &types.RegistryEntry{
				Key: &types.RegistryKey{
					AssetClassId: "test-class",
					NftId:        "test-nft",
				},
				Roles: []types.RolesEntry{},
			},
			expErr: "roles cannot be empty",
		},
		{
			name: "nil roles",
			entry: &types.RegistryEntry{
				Key: &types.RegistryKey{
					AssetClassId: "test-class",
					NftId:        "test-nft",
				},
				Roles: nil,
			},
			expErr: "roles cannot be empty",
		},
		{
			name: "invalid role",
			entry: &types.RegistryEntry{
				Key: &types.RegistryKey{
					AssetClassId: "test-class",
					NftId:        "test-nft",
				},
				Roles: []types.RolesEntry{
					{
						Role:      types.RegistryRole_REGISTRY_ROLE_UNSPECIFIED,
						Addresses: []string{validAddr},
					},
				},
			},
			expErr: "cannot be unspecified",
		},
		{
			name: "multiple valid roles",
			entry: &types.RegistryEntry{
				Key: &types.RegistryKey{
					AssetClassId: "test-class",
					NftId:        "test-nft",
				},
				Roles: []types.RolesEntry{
					{
						Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
						Addresses: []string{validAddr},
					},
					{
						Role:      types.RegistryRole_REGISTRY_ROLE_SERVICER,
						Addresses: []string{validAddr},
					},
				},
			},
			expErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.entry.Validate()
			if tc.expErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRolesEntry_Validate(t *testing.T) {
	validAddr := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"
	validAddr2 := "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"

	tests := []struct {
		name   string
		entry  *types.RolesEntry
		expErr string
	}{
		{
			name:   "nil entry",
			entry:  nil,
			expErr: "roles entry cannot be nil",
		},
		{
			name: "valid entry",
			entry: &types.RolesEntry{
				Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
				Addresses: []string{validAddr},
			},
			expErr: "",
		},
		{
			name: "unspecified role",
			entry: &types.RolesEntry{
				Role:      types.RegistryRole_REGISTRY_ROLE_UNSPECIFIED,
				Addresses: []string{validAddr},
			},
			expErr: "cannot be unspecified",
		},
		{
			name: "empty addresses",
			entry: &types.RolesEntry{
				Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
				Addresses: []string{},
			},
			expErr: "addresses cannot be empty",
		},
		{
			name: "nil addresses",
			entry: &types.RolesEntry{
				Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
				Addresses: nil,
			},
			expErr: "addresses cannot be empty",
		},
		{
			name: "invalid address",
			entry: &types.RolesEntry{
				Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
				Addresses: []string{"invalid-address"},
			},
			expErr: "decoding bech32",
		},
		{
			name: "duplicate addresses",
			entry: &types.RolesEntry{
				Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
				Addresses: []string{validAddr, validAddr},
			},
			expErr: "duplicate address",
		},
		{
			name: "multiple valid addresses",
			entry: &types.RolesEntry{
				Role:      types.RegistryRole_REGISTRY_ROLE_SERVICER,
				Addresses: []string{validAddr, validAddr2},
			},
			expErr: "",
		},
		{
			name: "address too long",
			entry: &types.RolesEntry{
				Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
				Addresses: []string{strings.Repeat("a", types.MaxLenAddress+1)},
			},
			expErr: "must be between",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.entry.Validate()
			if tc.expErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRegistryRole_Validate(t *testing.T) {
	tests := []struct {
		name   string
		role   types.RegistryRole
		expErr string
	}{
		{
			name:   "unspecified",
			role:   types.RegistryRole_REGISTRY_ROLE_UNSPECIFIED,
			expErr: "cannot be unspecified",
		},
		{
			name:   "originator",
			role:   types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			expErr: "",
		},
		{
			name:   "servicer",
			role:   types.RegistryRole_REGISTRY_ROLE_SERVICER,
			expErr: "",
		},
		{
			name:   "controller",
			role:   types.RegistryRole_REGISTRY_ROLE_CONTROLLER,
			expErr: "",
		},
		{
			name:   "invalid role value",
			role:   types.RegistryRole(999),
			expErr: "invalid role",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.role.Validate()
			if tc.expErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseRegistryRole(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		expected types.RegistryRole
		expErr   string
	}{
		{
			name:     "full name - originator",
			str:      "REGISTRY_ROLE_ORIGINATOR",
			expected: types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			expErr:   "",
		},
		{
			name:     "short name - originator",
			str:      "ORIGINATOR",
			expected: types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			expErr:   "",
		},
		{
			name:     "lowercase - originator",
			str:      "originator",
			expected: types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			expErr:   "",
		},
		{
			name:     "full name - servicer",
			str:      "REGISTRY_ROLE_SERVICER",
			expected: types.RegistryRole_REGISTRY_ROLE_SERVICER,
			expErr:   "",
		},
		{
			name:     "short name - servicer",
			str:      "SERVICER",
			expected: types.RegistryRole_REGISTRY_ROLE_SERVICER,
			expErr:   "",
		},
		{
			name:     "full name - controller",
			str:      "REGISTRY_ROLE_CONTROLLER",
			expected: types.RegistryRole_REGISTRY_ROLE_CONTROLLER,
			expErr:   "",
		},
		{
			name:     "short name - controller",
			str:      "CONTROLLER",
			expected: types.RegistryRole_REGISTRY_ROLE_CONTROLLER,
			expErr:   "",
		},
		{
			name:     "unspecified",
			str:      "REGISTRY_ROLE_UNSPECIFIED",
			expected: types.RegistryRole_REGISTRY_ROLE_UNSPECIFIED,
			expErr:   "cannot be unspecified",
		},
		{
			name:     "short unspecified",
			str:      "UNSPECIFIED",
			expected: types.RegistryRole_REGISTRY_ROLE_UNSPECIFIED,
			expErr:   "cannot be unspecified",
		},
		{
			name:     "invalid role",
			str:      "INVALID_ROLE",
			expected: types.RegistryRole_REGISTRY_ROLE_UNSPECIFIED,
			expErr:   "invalid role",
		},
		{
			name:     "empty string",
			str:      "",
			expected: types.RegistryRole_REGISTRY_ROLE_UNSPECIFIED,
			expErr:   "invalid role",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := types.ParseRegistryRole(tc.str)
			if tc.expErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestRegistryRole_ShortString(t *testing.T) {
	assert.Equal(t, types.RegistryRole_REGISTRY_ROLE_ORIGINATOR.ShortString(), "ORIGINATOR")
	assert.Equal(t, types.RegistryRole_REGISTRY_ROLE_SERVICER.ShortString(), "SERVICER")
	assert.Equal(t, types.RegistryRole_REGISTRY_ROLE_CONTROLLER.ShortString(), "CONTROLLER")
	assert.Equal(t, types.RegistryRole_REGISTRY_ROLE_UNSPECIFIED.ShortString(), "UNSPECIFIED")
}

func TestValidateClassID(t *testing.T) {
	tests := []struct {
		name    string
		classID string
		expErr  string
	}{
		{
			name:    "valid class id",
			classID: "test-class",
			expErr:  "",
		},
		{
			name:    "valid with dots",
			classID: "test.class.id",
			expErr:  "",
		},
		{
			name:    "valid with dashes",
			classID: "test-class-id",
			expErr:  "",
		},
		{
			name:    "valid alphanumeric",
			classID: "TestClass123",
			expErr:  "",
		},
		{
			name:    "empty class id",
			classID: "",
			expErr:  "must be between",
		},
		{
			name:    "too long",
			classID: strings.Repeat("a", types.MaxLenAssetClassID+1),
			expErr:  "must be between",
		},
		{
			name:    "max length",
			classID: strings.Repeat("a", types.MaxLenAssetClassID),
			expErr:  "",
		},
		{
			name:    "with whitespace",
			classID: "test class",
			expErr:  "\"test class\" must only contain alphanumeric, '-', '.' characters",
		},
		{
			name:    "valid metadata scope spec",
			classID: "scopespec1qsm47umrdacx2hmnwpjkxh6lta0sz95anh",
			expErr:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateClassID(tc.classID)
			if tc.expErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateNftID(t *testing.T) {
	tests := []struct {
		name   string
		nftID  string
		expErr string
	}{
		{
			name:   "valid nft id",
			nftID:  "test-nft",
			expErr: "",
		},
		{
			name:   "valid with dots",
			nftID:  "test.nft.id",
			expErr: "",
		},
		{
			name:   "valid with dashes",
			nftID:  "test-nft-id",
			expErr: "",
		},
		{
			name:   "valid alphanumeric",
			nftID:  "TestNFT123",
			expErr: "",
		},
		{
			name:   "empty nft id",
			nftID:  "",
			expErr: "must be between",
		},
		{
			name:   "too long",
			nftID:  strings.Repeat("a", types.MaxLenNFTID+1),
			expErr: "must be between",
		},
		{
			name:   "max length",
			nftID:  strings.Repeat("a", types.MaxLenNFTID),
			expErr: "",
		},
		{
			name:   "with whitespace",
			nftID:  "test nft",
			expErr: "must only contain alphanumeric",
		},
		{
			name:   "valid metadata scope",
			nftID:  "scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel",
			expErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateNftID(tc.nftID)
			if tc.expErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMetadataScopeID(t *testing.T) {
	tests := []struct {
		name            string
		bech32String    string
		expectedIsScope bool
	}{
		{
			name:            "valid scope",
			bech32String:    "scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel",
			expectedIsScope: true,
		},
		{
			name:            "valid scope spec (not a scope)",
			bech32String:    "scopespec1qsm47umrdacx2hmnwpjkxh6lta0sz95anh",
			expectedIsScope: false,
		},
		{
			name:            "invalid bech32",
			bech32String:    "invalid-bech32",
			expectedIsScope: false,
		},
		{
			name:            "empty string",
			bech32String:    "",
			expectedIsScope: false,
		},
		{
			name:            "regular string",
			bech32String:    "test-nft",
			expectedIsScope: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			addr, isScope := types.MetadataScopeID(tc.bech32String)
			assert.Equal(t, tc.expectedIsScope, isScope)
			if tc.expectedIsScope {
				assert.NotNil(t, addr)
			}
		})
	}
}

func TestMetadataScopeSpecID(t *testing.T) {
	tests := []struct {
		name                string
		bech32String        string
		expectedIsScopeSpec bool
	}{
		{
			name:                "valid scope spec",
			bech32String:        "scopespec1qsm47umrdacx2hmnwpjkxh6lta0sz95anh",
			expectedIsScopeSpec: true,
		},
		{
			name:                "valid scope (not a scope spec)",
			bech32String:        "scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel",
			expectedIsScopeSpec: false,
		},
		{
			name:                "invalid bech32",
			bech32String:        "invalid-bech32",
			expectedIsScopeSpec: false,
		},
		{
			name:                "empty string",
			bech32String:        "",
			expectedIsScopeSpec: false,
		},
		{
			name:                "regular string",
			bech32String:        "test-class",
			expectedIsScopeSpec: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			addr, isScopeSpec := types.MetadataScopeSpecID(tc.bech32String)
			assert.Equal(t, tc.expectedIsScopeSpec, isScopeSpec)
			if tc.expectedIsScopeSpec {
				assert.NotNil(t, addr)
			}
		})
	}
}
