package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/provenance-io/provenance/x/registry/types"
)

func TestDefaultGenesis(t *testing.T) {
	genesis := types.DefaultGenesis()
	require.NotNil(t, genesis)
	assert.Empty(t, genesis.Entries, "default genesis should have no entries")

	// Validate default genesis should not error
	err := genesis.Validate()
	assert.NoError(t, err, "default genesis should be valid")
}

func TestGenesisState_Validate(t *testing.T) {
	validAddr := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"

	tests := []struct {
		name    string
		genesis *types.GenesisState
		expErr  string
	}{
		{
			name:    "nil genesis",
			genesis: nil,
			expErr:  "invalid memory address",
		},
		{
			name:    "empty genesis",
			genesis: &types.GenesisState{},
			expErr:  "",
		},
		{
			name: "valid single entry",
			genesis: &types.GenesisState{
				Entries: []types.RegistryEntry{
					{
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
				},
			},
			expErr: "",
		},
		{
			name: "valid multiple entries",
			genesis: &types.GenesisState{
				Entries: []types.RegistryEntry{
					{
						Key: &types.RegistryKey{
							AssetClassId: "class1",
							NftId:        "nft1",
						},
						Roles: []types.RolesEntry{
							{
								Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
								Addresses: []string{validAddr},
							},
						},
					},
					{
						Key: &types.RegistryKey{
							AssetClassId: "class2",
							NftId:        "nft2",
						},
						Roles: []types.RolesEntry{
							{
								Role:      types.RegistryRole_REGISTRY_ROLE_SERVICER,
								Addresses: []string{validAddr},
							},
						},
					},
				},
			},
			expErr: "",
		},
		{
			name: "invalid entry - nil key",
			genesis: &types.GenesisState{
				Entries: []types.RegistryEntry{
					{
						Key: nil,
						Roles: []types.RolesEntry{
							{
								Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
								Addresses: []string{validAddr},
							},
						},
					},
				},
			},
			expErr: "registry key cannot be nil",
		},
		{
			name: "invalid entry - empty asset class id",
			genesis: &types.GenesisState{
				Entries: []types.RegistryEntry{
					{
						Key: &types.RegistryKey{
							AssetClassId: "",
							NftId:        "nft1",
						},
						Roles: []types.RolesEntry{
							{
								Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
								Addresses: []string{validAddr},
							},
						},
					},
				},
			},
			expErr: "must be between",
		},
		{
			name: "invalid entry - empty nft id",
			genesis: &types.GenesisState{
				Entries: []types.RegistryEntry{
					{
						Key: &types.RegistryKey{
							AssetClassId: "class1",
							NftId:        "",
						},
						Roles: []types.RolesEntry{
							{
								Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
								Addresses: []string{validAddr},
							},
						},
					},
				},
			},
			expErr: "must be between",
		},
		{
			name: "invalid entry - nil roles",
			genesis: &types.GenesisState{
				Entries: []types.RegistryEntry{
					{
						Key: &types.RegistryKey{
							AssetClassId: "class1",
							NftId:        "nft1",
						},
						Roles: nil,
					},
				},
			},
			expErr: "roles cannot be empty",
		},
		{
			name: "invalid entry - empty roles",
			genesis: &types.GenesisState{
				Entries: []types.RegistryEntry{
					{
						Key: &types.RegistryKey{
							AssetClassId: "class1",
							NftId:        "nft1",
						},
						Roles: []types.RolesEntry{},
					},
				},
			},
			expErr: "roles cannot be empty",
		},
		{
			name: "invalid entry - unspecified role",
			genesis: &types.GenesisState{
				Entries: []types.RegistryEntry{
					{
						Key: &types.RegistryKey{
							AssetClassId: "class1",
							NftId:        "nft1",
						},
						Roles: []types.RolesEntry{
							{
								Role:      types.RegistryRole_REGISTRY_ROLE_UNSPECIFIED,
								Addresses: []string{validAddr},
							},
						},
					},
				},
			},
			expErr: "cannot be unspecified",
		},
		{
			name: "invalid entry - empty addresses",
			genesis: &types.GenesisState{
				Entries: []types.RegistryEntry{
					{
						Key: &types.RegistryKey{
							AssetClassId: "class1",
							NftId:        "nft1",
						},
						Roles: []types.RolesEntry{
							{
								Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
								Addresses: []string{},
							},
						},
					},
				},
			},
			expErr: "addresses cannot be empty",
		},
		{
			name: "invalid entry - invalid address",
			genesis: &types.GenesisState{
				Entries: []types.RegistryEntry{
					{
						Key: &types.RegistryKey{
							AssetClassId: "class1",
							NftId:        "nft1",
						},
						Roles: []types.RolesEntry{
							{
								Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
								Addresses: []string{"invalid-address"},
							},
						},
					},
				},
			},
			expErr: "decoding bech32",
		},
		{
			name: "invalid entry - duplicate addresses",
			genesis: &types.GenesisState{
				Entries: []types.RegistryEntry{
					{
						Key: &types.RegistryKey{
							AssetClassId: "class1",
							NftId:        "nft1",
						},
						Roles: []types.RolesEntry{
							{
								Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
								Addresses: []string{validAddr, validAddr},
							},
						},
					},
				},
			},
			expErr: "duplicate address",
		},
		{
			name: "invalid entry - too long asset class id",
			genesis: &types.GenesisState{
				Entries: []types.RegistryEntry{
					{
						Key: &types.RegistryKey{
							AssetClassId: string(make([]byte, types.MaxLenAssetClassID+1)),
							NftId:        "nft1",
						},
						Roles: []types.RolesEntry{
							{
								Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
								Addresses: []string{validAddr},
							},
						},
					},
				},
			},
			expErr: "must be between",
		},
		{
			name: "invalid entry - too long nft id",
			genesis: &types.GenesisState{
				Entries: []types.RegistryEntry{
					{
						Key: &types.RegistryKey{
							AssetClassId: "class1",
							NftId:        string(make([]byte, types.MaxLenNFTID+1)),
						},
						Roles: []types.RolesEntry{
							{
								Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
								Addresses: []string{validAddr},
							},
						},
					},
				},
			},
			expErr: "must be between",
		},
		{
			name: "valid entry - multiple roles",
			genesis: &types.GenesisState{
				Entries: []types.RegistryEntry{
					{
						Key: &types.RegistryKey{
							AssetClassId: "class1",
							NftId:        "nft1",
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
				},
			},
			expErr: "",
		},
		{
			name: "valid entry - multiple addresses",
			genesis: &types.GenesisState{
				Entries: []types.RegistryEntry{
					{
						Key: &types.RegistryKey{
							AssetClassId: "class1",
							NftId:        "nft1",
						},
						Roles: []types.RolesEntry{
							{
								Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
								Addresses: []string{validAddr, "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"},
							},
						},
					},
				},
			},
			expErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			if tc.genesis == nil {
				// Special handling for nil genesis test
				require.Panics(t, func() {
					_ = tc.genesis.Validate()
				}, "should panic on nil genesis")
				return
			}

			err = tc.genesis.Validate()

			if tc.expErr != "" {
				require.Error(t, err, "expected error but got none")
				assert.Contains(t, err.Error(), tc.expErr, "error message should contain expected text")
			} else {
				require.NoError(t, err, "expected no error but got: %v", err)
			}
		})
	}
}
