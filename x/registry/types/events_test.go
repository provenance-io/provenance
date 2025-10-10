package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

// assertEverythingSet asserts that the provided proto.Message can be converted to
// an untyped event with the expected type string. Then, if assertAllSet = true,
// this asserts that none of the event attributes are empty.
// Returns true on success, false if one or more things aren't right.
func assertEventContent(t *testing.T, tev proto.Message, typeString string, assertAllSet bool) bool {
	t.Helper()
	event, err := sdk.TypedEventToEvent(tev)
	if !assert.NoError(t, err, "TypedEventToEvent(%T)", tev) {
		return false
	}

	expType := "provenance.registry.v1." + typeString
	rv := assert.Equal(t, expType, event.Type, "%T event.Type", tev)
	if !assertAllSet {
		return rv
	}

	for i, attr := range event.Attributes {
		rv = assert.NotEmpty(t, attr.Key, "%T event.attributes[%d].Key", tev, i) && rv
		rv = assert.NotEqual(t, `""`, attr.Key, "%T event.attributes[%d].Key", tev, i) && rv
		rv = assert.NotEqual(t, `0`, attr.Key, "%T event.attributes[%d].Key", tev, i) && rv
		rv = assert.NotEqual(t, `"0"`, attr.Key, "%T event.attributes[%d].Key", tev, i) && rv
		rv = assert.NotEmpty(t, attr.Value, "%T event.attributes[%d].Value", tev, i) && rv
		rv = assert.NotEqual(t, `""`, attr.Value, "%T event.attributes[%d].Value", tev, i) && rv
		rv = assert.NotEqual(t, `0`, attr.Value, "%T event.attributes[%d].Value", tev, i) && rv
		rv = assert.NotEqual(t, `"0"`, attr.Value, "%T event.attributes[%d].Value", tev, i) && rv
	}
	return rv
}

// assertEverythingSet asserts that the provided proto.Message can be converted to
// an untyped event with the expected type string, and a value for all fields.
// Returns true on success, false if one or more things aren't right.
func assertEverythingSet(t *testing.T, tev proto.Message, typeString string) bool {
	return assertEventContent(t, tev, typeString, true)
}

func TestNewEventNFTRegistered(t *testing.T) {
	key := &RegistryKey{AssetClassId: "asset-clAss-id", NftId: "nFt-id"}
	exp := &EventNFTRegistered{
		AssetClassId: "asset-clAss-id",
		NftId:        "nFt-id",
	}

	var act *EventNFTRegistered
	testFunc := func() {
		act = NewEventNFTRegistered(key)
	}
	require.NotPanics(t, testFunc, "NewEventNFTRegistered")
	assert.Equal(t, exp, act, "NewEventNFTRegistered result")
	assertEverythingSet(t, act, "EventNFTRegistered")
}

func TestNewEventRoleGranted(t *testing.T) {
	tests := []struct {
		name  string
		key   *RegistryKey
		role  RegistryRole
		addrs []string
		exp   *EventRoleGranted
	}{
		{
			name:  "role: servicer",
			key:   &RegistryKey{AssetClassId: "Asset-class-id", NftId: "nft-iD"},
			role:  RegistryRole_REGISTRY_ROLE_SERVICER,
			addrs: []string{"addr0"},
			exp: &EventRoleGranted{
				AssetClassId: "Asset-class-id",
				NftId:        "nft-iD",
				Role:         "SERVICER",
				Addresses:    []string{"addr0"},
			},
		},
		{
			name:  "role: sub-servicer",
			key:   &RegistryKey{AssetClassId: "aSset-class-id", NftId: "nft-Id"},
			role:  RegistryRole_REGISTRY_ROLE_SUBSERVICER,
			addrs: []string{"addr0"},
			exp: &EventRoleGranted{
				AssetClassId: "aSset-class-id",
				NftId:        "nft-Id",
				Role:         "SUBSERVICER",
				Addresses:    []string{"addr0"},
			},
		},
		{
			name:  "role: controller",
			key:   &RegistryKey{AssetClassId: "asSet-class-id", NftId: "nfT-id"},
			role:  RegistryRole_REGISTRY_ROLE_CONTROLLER,
			addrs: []string{"addr0"},
			exp: &EventRoleGranted{
				AssetClassId: "asSet-class-id",
				NftId:        "nfT-id",
				Role:         "CONTROLLER",
				Addresses:    []string{"addr0"},
			},
		},
		{
			name:  "role: custodian",
			key:   &RegistryKey{AssetClassId: "assEt-class-id", NftId: "nFt-id"},
			role:  RegistryRole_REGISTRY_ROLE_CUSTODIAN,
			addrs: []string{"addr0"},
			exp: &EventRoleGranted{
				AssetClassId: "assEt-class-id",
				NftId:        "nFt-id",
				Role:         "CUSTODIAN",
				Addresses:    []string{"addr0"},
			},
		},
		{
			name:  "role: borrower",
			key:   &RegistryKey{AssetClassId: "asseT-class-id", NftId: "Nft-id"},
			role:  RegistryRole_REGISTRY_ROLE_BORROWER,
			addrs: []string{"addr0"},
			exp: &EventRoleGranted{
				AssetClassId: "asseT-class-id",
				NftId:        "Nft-id",
				Role:         "BORROWER",
				Addresses:    []string{"addr0"},
			},
		},
		{
			name:  "role: originator",
			key:   &RegistryKey{AssetClassId: "asset-Class-id", NftId: "nft-iD"},
			role:  RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			addrs: []string{"addr0"},
			exp: &EventRoleGranted{
				AssetClassId: "asset-Class-id",
				NftId:        "nft-iD",
				Role:         "ORIGINATOR",
				Addresses:    []string{"addr0"},
			},
		},
		{
			name:  "role: unspecified",
			key:   &RegistryKey{AssetClassId: "asset-cLass-id", NftId: "nft-Id"},
			role:  RegistryRole_REGISTRY_ROLE_UNSPECIFIED,
			addrs: []string{"addr0"},
			exp: &EventRoleGranted{
				AssetClassId: "asset-cLass-id",
				NftId:        "nft-Id",
				Role:         "UNSPECIFIED",
				Addresses:    []string{"addr0"},
			},
		},
		{
			name:  "role: unknown",
			key:   &RegistryKey{AssetClassId: "asset-clAss-id", NftId: "nfT-id"},
			role:  420,
			addrs: []string{"addr0"},
			exp: &EventRoleGranted{
				AssetClassId: "asset-clAss-id",
				NftId:        "nfT-id",
				Role:         "420",
				Addresses:    []string{"addr0"},
			},
		},
		{
			name:  "addrs: empty",
			key:   &RegistryKey{AssetClassId: "asset-claSs-id", NftId: "nFt-id"},
			role:  RegistryRole_REGISTRY_ROLE_SERVICER,
			addrs: []string{},
			exp: &EventRoleGranted{
				AssetClassId: "asset-claSs-id",
				NftId:        "nFt-id",
				Role:         "SERVICER",
				Addresses:    []string{},
			},
		},
		{
			name:  "addrs: nil",
			key:   &RegistryKey{AssetClassId: "asset-clasS-id", NftId: "Nft-id"},
			role:  RegistryRole_REGISTRY_ROLE_SERVICER,
			addrs: nil,
			exp: &EventRoleGranted{
				AssetClassId: "asset-clasS-id",
				NftId:        "Nft-id",
				Role:         "SERVICER",
				Addresses:    nil,
			},
		},
		{
			name:  "addrs: 2",
			key:   &RegistryKey{AssetClassId: "asset-class-Id", NftId: "nft-iD"},
			role:  RegistryRole_REGISTRY_ROLE_SERVICER,
			addrs: []string{"addr0", "addr1"},
			exp: &EventRoleGranted{
				AssetClassId: "asset-class-Id",
				NftId:        "nft-iD",
				Role:         "SERVICER",
				Addresses:    []string{"addr0", "addr1"},
			},
		},
		{
			name:  "addrs: 3",
			key:   &RegistryKey{AssetClassId: "asset-class-iD", NftId: "nft-Id"},
			role:  RegistryRole_REGISTRY_ROLE_SERVICER,
			addrs: []string{"addr10", "addr11", "addr12"},
			exp: &EventRoleGranted{
				AssetClassId: "asset-class-iD",
				NftId:        "nft-Id",
				Role:         "SERVICER",
				Addresses:    []string{"addr10", "addr11", "addr12"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act *EventRoleGranted
			testFunc := func() {
				act = NewEventRoleGranted(tc.key, tc.role, tc.addrs)
			}
			require.NotPanics(t, testFunc, "NewEventRoleGranted")
			assert.Equal(t, tc.exp, act, "NewEventRoleGranted result")
			assertEverythingSet(t, act, "EventRoleGranted")
		})
	}
}

func TestNewEventRoleRevoked(t *testing.T) {
	tests := []struct {
		name  string
		key   *RegistryKey
		role  RegistryRole
		addrs []string
		exp   *EventRoleRevoked
	}{
		{
			name:  "role: servicer",
			key:   &RegistryKey{AssetClassId: "Asset-class-id", NftId: "nft-iD"},
			role:  RegistryRole_REGISTRY_ROLE_SERVICER,
			addrs: []string{"addr0"},
			exp: &EventRoleRevoked{
				AssetClassId: "Asset-class-id",
				NftId:        "nft-iD",
				Role:         "SERVICER",
				Addresses:    []string{"addr0"},
			},
		},
		{
			name:  "role: sub-servicer",
			key:   &RegistryKey{AssetClassId: "aSset-class-id", NftId: "nft-Id"},
			role:  RegistryRole_REGISTRY_ROLE_SUBSERVICER,
			addrs: []string{"addr0"},
			exp: &EventRoleRevoked{
				AssetClassId: "aSset-class-id",
				NftId:        "nft-Id",
				Role:         "SUBSERVICER",
				Addresses:    []string{"addr0"},
			},
		},
		{
			name:  "role: controller",
			key:   &RegistryKey{AssetClassId: "asSet-class-id", NftId: "nfT-id"},
			role:  RegistryRole_REGISTRY_ROLE_CONTROLLER,
			addrs: []string{"addr0"},
			exp: &EventRoleRevoked{
				AssetClassId: "asSet-class-id",
				NftId:        "nfT-id",
				Role:         "CONTROLLER",
				Addresses:    []string{"addr0"},
			},
		},
		{
			name:  "role: custodian",
			key:   &RegistryKey{AssetClassId: "assEt-class-id", NftId: "nFt-id"},
			role:  RegistryRole_REGISTRY_ROLE_CUSTODIAN,
			addrs: []string{"addr0"},
			exp: &EventRoleRevoked{
				AssetClassId: "assEt-class-id",
				NftId:        "nFt-id",
				Role:         "CUSTODIAN",
				Addresses:    []string{"addr0"},
			},
		},
		{
			name:  "role: borrower",
			key:   &RegistryKey{AssetClassId: "asseT-class-id", NftId: "Nft-id"},
			role:  RegistryRole_REGISTRY_ROLE_BORROWER,
			addrs: []string{"addr0"},
			exp: &EventRoleRevoked{
				AssetClassId: "asseT-class-id",
				NftId:        "Nft-id",
				Role:         "BORROWER",
				Addresses:    []string{"addr0"},
			},
		},
		{
			name:  "role: originator",
			key:   &RegistryKey{AssetClassId: "asset-Class-id", NftId: "nft-iD"},
			role:  RegistryRole_REGISTRY_ROLE_ORIGINATOR,
			addrs: []string{"addr0"},
			exp: &EventRoleRevoked{
				AssetClassId: "asset-Class-id",
				NftId:        "nft-iD",
				Role:         "ORIGINATOR",
				Addresses:    []string{"addr0"},
			},
		},
		{
			name:  "role: unspecified",
			key:   &RegistryKey{AssetClassId: "asset-cLass-id", NftId: "nft-Id"},
			role:  RegistryRole_REGISTRY_ROLE_UNSPECIFIED,
			addrs: []string{"addr0"},
			exp: &EventRoleRevoked{
				AssetClassId: "asset-cLass-id",
				NftId:        "nft-Id",
				Role:         "UNSPECIFIED",
				Addresses:    []string{"addr0"},
			},
		},
		{
			name:  "role: unknown",
			key:   &RegistryKey{AssetClassId: "asset-clAss-id", NftId: "nfT-id"},
			role:  420,
			addrs: []string{"addr0"},
			exp: &EventRoleRevoked{
				AssetClassId: "asset-clAss-id",
				NftId:        "nfT-id",
				Role:         "420",
				Addresses:    []string{"addr0"},
			},
		},
		{
			name:  "addrs: empty",
			key:   &RegistryKey{AssetClassId: "asset-claSs-id", NftId: "nFt-id"},
			role:  RegistryRole_REGISTRY_ROLE_SERVICER,
			addrs: []string{},
			exp: &EventRoleRevoked{
				AssetClassId: "asset-claSs-id",
				NftId:        "nFt-id",
				Role:         "SERVICER",
				Addresses:    []string{},
			},
		},
		{
			name:  "addrs: nil",
			key:   &RegistryKey{AssetClassId: "asset-clasS-id", NftId: "Nft-id"},
			role:  RegistryRole_REGISTRY_ROLE_SERVICER,
			addrs: nil,
			exp: &EventRoleRevoked{
				AssetClassId: "asset-clasS-id",
				NftId:        "Nft-id",
				Role:         "SERVICER",
				Addresses:    nil,
			},
		},
		{
			name:  "addrs: 2",
			key:   &RegistryKey{AssetClassId: "asset-class-Id", NftId: "nft-iD"},
			role:  RegistryRole_REGISTRY_ROLE_SERVICER,
			addrs: []string{"addr0", "addr1"},
			exp: &EventRoleRevoked{
				AssetClassId: "asset-class-Id",
				NftId:        "nft-iD",
				Role:         "SERVICER",
				Addresses:    []string{"addr0", "addr1"},
			},
		},
		{
			name:  "addrs: 3",
			key:   &RegistryKey{AssetClassId: "asset-class-iD", NftId: "nft-Id"},
			role:  RegistryRole_REGISTRY_ROLE_SERVICER,
			addrs: []string{"addr10", "addr11", "addr12"},
			exp: &EventRoleRevoked{
				AssetClassId: "asset-class-iD",
				NftId:        "nft-Id",
				Role:         "SERVICER",
				Addresses:    []string{"addr10", "addr11", "addr12"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act *EventRoleRevoked
			testFunc := func() {
				act = NewEventRoleRevoked(tc.key, tc.role, tc.addrs)
			}
			require.NotPanics(t, testFunc, "NewEventRoleRevoked")
			assert.Equal(t, tc.exp, act, "NewEventRoleRevoked result")
			assertEverythingSet(t, act, "EventRoleRevoked")
		})
	}
}

func TestNewEventNFTUnregistered(t *testing.T) {
	key := &RegistryKey{AssetClassId: "asset_class_id", NftId: "nft_id"}
	exp := &EventNFTUnregistered{
		AssetClassId: "asset_class_id",
		NftId:        "nft_id",
	}

	var act *EventNFTUnregistered
	testFunc := func() {
		act = NewEventNFTUnregistered(key)
	}
	require.NotPanics(t, testFunc, "NewEventNFTUnregistered")
	assert.Equal(t, exp, act, "NewEventNFTUnregistered result")
	assertEverythingSet(t, act, "EventNFTUnregistered")
}

func TestNewEventRegistryBulkUpdated(t *testing.T) {
	key := &RegistryKey{AssetClassId: "asset-CLASS-id", NftId: "NFT-id"}
	exp := &EventRegistryBulkUpdated{
		AssetClassId: "asset-CLASS-id",
		NftId:        "NFT-id",
	}

	var act *EventRegistryBulkUpdated
	testFunc := func() {
		act = NewEventRegistryBulkUpdated(key)
	}
	require.NotPanics(t, testFunc, "NewEventRegistryBulkUpdated")
	assert.Equal(t, exp, act, "NewEventRegistryBulkUpdated result")
	assertEverythingSet(t, act, "EventRegistryBulkUpdated")
}
