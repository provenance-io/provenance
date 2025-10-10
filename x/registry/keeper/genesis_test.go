package keeper_test

import (
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/registry/types"
)

func (s *KeeperTestSuite) TestInitGenesis() {
	tests := []struct {
		name      string
		state     *types.GenesisState
		expPanic  string
		expErr    bool
		expStored []types.RegistryEntry
	}{
		{
			name:      "nil genesis state",
			state:     nil,
			expPanic:  "runtime error: invalid memory address or nil pointer dereference",
			expErr:    false,
			expStored: nil,
		},
		{
			name:      "empty genesis state",
			state:     &types.GenesisState{},
			expPanic:  "",
			expErr:    false,
			expStored: nil,
		},
		{
			name: "single entry",
			state: &types.GenesisState{
				Entries: []types.RegistryEntry{
					{
						Key: &types.RegistryKey{
							AssetClassId: "class1",
							NftId:        "nft1",
						},
						Roles: []types.RolesEntry{
							{
								Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
								Addresses: []string{s.user1},
							},
						},
					},
				},
			},
			expPanic: "",
			expErr:   false,
			expStored: []types.RegistryEntry{
				{
					Key: &types.RegistryKey{
						AssetClassId: "class1",
						NftId:        "nft1",
					},
					Roles: []types.RolesEntry{
						{
							Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
							Addresses: []string{s.user1},
						},
					},
				},
			},
		},
		{
			name: "multiple entries",
			state: &types.GenesisState{
				Entries: []types.RegistryEntry{
					{
						Key: &types.RegistryKey{
							AssetClassId: "class1",
							NftId:        "nft1",
						},
						Roles: []types.RolesEntry{
							{
								Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
								Addresses: []string{s.user1},
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
								Addresses: []string{s.user2},
							},
						},
					},
				},
			},
			expPanic: "",
			expErr:   false,
			expStored: []types.RegistryEntry{
				{
					Key: &types.RegistryKey{
						AssetClassId: "class1",
						NftId:        "nft1",
					},
					Roles: []types.RolesEntry{
						{
							Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
							Addresses: []string{s.user1},
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
							Addresses: []string{s.user2},
						},
					},
				},
			},
		},
		{
			name: "multiple roles for same entry",
			state: &types.GenesisState{
				Entries: []types.RegistryEntry{
					{
						Key: &types.RegistryKey{
							AssetClassId: "class1",
							NftId:        "nft1",
						},
						Roles: []types.RolesEntry{
							{
								Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
								Addresses: []string{s.user1},
							},
							{
								Role:      types.RegistryRole_REGISTRY_ROLE_SERVICER,
								Addresses: []string{s.user1, s.user2},
							},
						},
					},
				},
			},
			expPanic: "",
			expErr:   false,
			expStored: []types.RegistryEntry{
				{
					Key: &types.RegistryKey{
						AssetClassId: "class1",
						NftId:        "nft1",
					},
					Roles: []types.RolesEntry{
						{
							Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
							Addresses: []string{s.user1},
						},
						{
							Role:      types.RegistryRole_REGISTRY_ROLE_SERVICER,
							Addresses: []string{s.user1, s.user2},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Clear any existing state
			s.app.RegistryKeeper.Registry.Clear(s.ctx, nil)

			testFunc := func() {
				s.app.RegistryKeeper.InitGenesis(s.ctx, tc.state)
			}

			if tc.expPanic != "" {
				assertions.RequirePanicEquals(s.T(), testFunc, tc.expPanic, "InitGenesis")
			} else {
				s.Require().NotPanics(testFunc, "InitGenesis")

				// Verify stored entries
				if tc.expStored != nil {
					for _, expected := range tc.expStored {
						actual, err := s.app.RegistryKeeper.GetRegistry(s.ctx, expected.Key)
						s.Require().NoError(err)
						s.Require().NotNil(actual)
						s.Require().Equal(expected.Key, actual.Key)
						s.Require().Equal(expected.Roles, actual.Roles)
					}
				} else {
					// Verify no entries were stored
					entries, _, err := s.app.RegistryKeeper.GetRegistries(s.ctx, nil, "")
					s.Require().NoError(err)
					s.Require().Empty(entries)
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestExportGenesis() {
	tests := []struct {
		name       string
		setup      func()
		expEntries []types.RegistryEntry
	}{
		{
			name:       "empty state",
			setup:      func() {},
			expEntries: nil,
		},
		{
			name: "single entry",
			setup: func() {
				key := &types.RegistryKey{
					AssetClassId: "class1",
					NftId:        "nft1",
				}
				roles := []types.RolesEntry{
					{
						Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
						Addresses: []string{s.user1},
					},
				}
				err := s.app.RegistryKeeper.CreateRegistry(s.ctx, key, roles)
				s.Require().NoError(err)
			},
			expEntries: []types.RegistryEntry{
				{
					Key: &types.RegistryKey{
						AssetClassId: "class1",
						NftId:        "nft1",
					},
					Roles: []types.RolesEntry{
						{
							Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
							Addresses: []string{s.user1},
						},
					},
				},
			},
		},
		{
			name: "multiple entries",
			setup: func() {
				entries := []struct {
					key   *types.RegistryKey
					roles []types.RolesEntry
				}{
					{
						key: &types.RegistryKey{
							AssetClassId: "class1",
							NftId:        "nft1",
						},
						roles: []types.RolesEntry{
							{
								Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
								Addresses: []string{s.user1},
							},
						},
					},
					{
						key: &types.RegistryKey{
							AssetClassId: "class2",
							NftId:        "nft2",
						},
						roles: []types.RolesEntry{
							{
								Role:      types.RegistryRole_REGISTRY_ROLE_SERVICER,
								Addresses: []string{s.user2},
							},
						},
					},
				}

				for _, e := range entries {
					err := s.app.RegistryKeeper.CreateRegistry(s.ctx, e.key, e.roles)
					s.Require().NoError(err)
				}
			},
			expEntries: []types.RegistryEntry{
				{
					Key: &types.RegistryKey{
						AssetClassId: "class1",
						NftId:        "nft1",
					},
					Roles: []types.RolesEntry{
						{
							Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
							Addresses: []string{s.user1},
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
							Addresses: []string{s.user2},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Clear state before each test
			s.app.RegistryKeeper.Registry.Clear(s.ctx, nil)

			// Setup test data
			tc.setup()

			// Export genesis
			var genState *types.GenesisState
			testFunc := func() {
				genState = s.app.RegistryKeeper.ExportGenesis(s.ctx)
			}
			s.Require().NotPanics(testFunc, "ExportGenesis")
			s.Require().NotNil(genState)

			// Verify entries
			if tc.expEntries == nil {
				s.Require().Empty(genState.Entries)
			} else {
				s.Require().Len(genState.Entries, len(tc.expEntries))

				// Compare entries (order may vary)
				for _, expected := range tc.expEntries {
					found := false
					for _, actual := range genState.Entries {
						if expected.Key.AssetClassId == actual.Key.AssetClassId && expected.Key.NftId == actual.Key.NftId {
							s.Require().Equal(expected.Roles, actual.Roles)
							found = true
							break
						}
					}
					s.Require().True(found, "Expected entry not found: %v", expected.Key)
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestGenesisRoundTrip() {
	// Create some test data
	key1 := &types.RegistryKey{
		AssetClassId: "class1",
		NftId:        "nft1",
	}
	roles1 := []types.RolesEntry{
		{
			Role:      types.RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			Addresses: []string{s.user1},
		},
	}

	key2 := &types.RegistryKey{
		AssetClassId: "class2",
		NftId:        "nft2",
	}
	roles2 := []types.RolesEntry{
		{
			Role:      types.RegistryRole_REGISTRY_ROLE_SERVICER,
			Addresses: []string{s.user2},
		},
		{
			Role:      types.RegistryRole_REGISTRY_ROLE_CONTROLLER,
			Addresses: []string{s.user1, s.user2},
		},
	}

	// Create registries
	err := s.app.RegistryKeeper.CreateRegistry(s.ctx, key1, roles1)
	s.Require().NoError(err)
	err = s.app.RegistryKeeper.CreateRegistry(s.ctx, key2, roles2)
	s.Require().NoError(err)

	// Export genesis
	genesis1 := s.app.RegistryKeeper.ExportGenesis(s.ctx)
	s.Require().NotNil(genesis1)
	s.Require().Len(genesis1.Entries, 2)

	// Clear the state
	s.app.RegistryKeeper.Registry.Clear(s.ctx, nil)

	// Verify state is empty
	entries, _, err := s.app.RegistryKeeper.GetRegistries(s.ctx, nil, "")
	s.Require().NoError(err)
	s.Require().Empty(entries)

	// Import genesis
	testFunc := func() {
		s.app.RegistryKeeper.InitGenesis(s.ctx, genesis1)
	}
	s.Require().NotPanics(testFunc, "InitGenesis")

	// Export again
	genesis2 := s.app.RegistryKeeper.ExportGenesis(s.ctx)
	s.Require().NotNil(genesis2)

	// Compare the two genesis states
	s.Require().Len(genesis2.Entries, len(genesis1.Entries))

	// Verify each entry
	for _, e1 := range genesis1.Entries {
		found := false
		for _, e2 := range genesis2.Entries {
			if e1.Key.AssetClassId == e2.Key.AssetClassId && e1.Key.NftId == e2.Key.NftId {
				s.Require().Equal(e1.Roles, e2.Roles)
				found = true
				break
			}
		}
		s.Require().True(found, "Entry not found in second export: %v", e1.Key)
	}
}
