package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/testutil/assertions"

	. "github.com/provenance-io/provenance/x/registry/types"
)

func TestAllMsgsGetSigners(t *testing.T) {
	msgMakers := []testutil.MsgMaker{
		func(signer string) sdk.Msg { return &MsgRegisterNFT{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgUnregisterNFT{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgRegistryBulkUpdate{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgProposeRoleChange{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgApproveRoleChange{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgGrantRole{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgRevokeRole{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgSetRoles{Signer: signer} },
	}
	msgMakersMulti := []testutil.MsgMakerMulti{}

	testutil.RunGetSignersTests(t, AllRequestMsgs, msgMakers, msgMakersMulti)
}

func TestMsgRegisterNFT_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress("reg_signer____________").String()
	validKey := &RegistryKey{AssetClassId: "aclass", NftId: "nft1"}
	validRoles := []RolesEntry{{Role: RegistryRole_REGISTRY_ROLE_ORIGINATOR, Addresses: []string{validAddr}}}

	tests := []struct {
		name string
		msg  MsgRegisterNFT
		exp  string
	}{
		{name: "valid", msg: MsgRegisterNFT{Signer: validAddr, Key: validKey, Roles: validRoles}},
		{name: "empty signer", msg: MsgRegisterNFT{Signer: "", Key: validKey, Roles: validRoles}, exp: "invalid signer: empty address"},
		{name: "bad signer", msg: MsgRegisterNFT{Signer: "bad", Key: validKey, Roles: validRoles}, exp: "invalid signer: decoding bech32"},
		{name: "nil key", msg: MsgRegisterNFT{Signer: validAddr, Key: nil, Roles: validRoles}, exp: "invalid key: registry key cannot be nil"},
		{name: "invalid key", msg: MsgRegisterNFT{Signer: validAddr, Key: &RegistryKey{AssetClassId: "", NftId: "nft1"}, Roles: validRoles}, exp: "invalid key: asset class id: must be between"},
		{name: "role unspecified", msg: MsgRegisterNFT{Signer: validAddr, Key: validKey, Roles: []RolesEntry{{Role: RegistryRole_REGISTRY_ROLE_UNSPECIFIED, Addresses: []string{validAddr}}}}, exp: "invalid role: role: cannot be unspecified"},
		{name: "role empty addresses", msg: MsgRegisterNFT{Signer: validAddr, Key: validKey, Roles: []RolesEntry{{Role: RegistryRole_REGISTRY_ROLE_ORIGINATOR, Addresses: []string{}}}}, exp: "invalid role: addresses cannot be empty"},
		{name: "role bad address", msg: MsgRegisterNFT{Signer: validAddr, Key: validKey, Roles: []RolesEntry{{Role: RegistryRole_REGISTRY_ROLE_ORIGINATOR, Addresses: []string{"bad"}}}}, exp: "invalid role: address[0]: decoding bech32"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.exp == "" {
				assertions.RequireErrorContents(t, err, nil)
			} else {
				assertions.RequireErrorContents(t, err, []string{tc.exp})
			}
		})
	}
}

func TestMsgGrantRole_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress("grant_signer___________").String()
	otherAddr := sdk.AccAddress("other_addr____________").String()
	validKey := &RegistryKey{AssetClassId: "aclass", NftId: "nft1"}

	tests := []struct {
		name string
		msg  MsgGrantRole
		exp  string
	}{
		{name: "valid", msg: MsgGrantRole{Signer: validAddr, Key: validKey, Role: RegistryRole_REGISTRY_ROLE_SERVICER, Addresses: []string{otherAddr}}},
		{name: "empty signer", msg: MsgGrantRole{Signer: "", Key: validKey, Role: RegistryRole_REGISTRY_ROLE_SERVICER, Addresses: []string{otherAddr}}, exp: "invalid signer: empty address"},
		{name: "bad signer", msg: MsgGrantRole{Signer: "bad", Key: validKey, Role: RegistryRole_REGISTRY_ROLE_SERVICER, Addresses: []string{otherAddr}}, exp: "invalid signer: decoding bech32"},
		{name: "nil key", msg: MsgGrantRole{Signer: validAddr, Key: nil, Role: RegistryRole_REGISTRY_ROLE_SERVICER, Addresses: []string{otherAddr}}, exp: "invalid key: registry key cannot be nil"},
		{name: "invalid role", msg: MsgGrantRole{Signer: validAddr, Key: validKey, Role: RegistryRole_REGISTRY_ROLE_UNSPECIFIED, Addresses: []string{otherAddr}}, exp: "invalid role: cannot be unspecified"},
		{name: "no addresses", msg: MsgGrantRole{Signer: validAddr, Key: validKey, Role: RegistryRole_REGISTRY_ROLE_SERVICER, Addresses: []string{}}, exp: "invalid addresses: addresses cannot be empty"},
		{name: "bad address", msg: MsgGrantRole{Signer: validAddr, Key: validKey, Role: RegistryRole_REGISTRY_ROLE_SERVICER, Addresses: []string{"bad"}}, exp: "invalid addresses: decoding bech32"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.exp == "" {
				assertions.RequireErrorContents(t, err, nil)
			} else {
				assertions.RequireErrorContents(t, err, []string{tc.exp})
			}
		})
	}
}

func TestMsgRevokeRole_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress("revoke_signer__________").String()
	otherAddr := sdk.AccAddress("other_addr2___________").String()
	validKey := &RegistryKey{AssetClassId: "aclass", NftId: "nft1"}

	tests := []struct {
		name string
		msg  MsgRevokeRole
		exp  string
	}{
		{name: "valid", msg: MsgRevokeRole{Signer: validAddr, Key: validKey, Role: RegistryRole_REGISTRY_ROLE_ORIGINATOR, Addresses: []string{otherAddr}}},
		{name: "empty signer", msg: MsgRevokeRole{Signer: "", Key: validKey, Role: RegistryRole_REGISTRY_ROLE_ORIGINATOR, Addresses: []string{otherAddr}}, exp: "invalid signer: empty address"},
		{name: "bad signer", msg: MsgRevokeRole{Signer: "bad", Key: validKey, Role: RegistryRole_REGISTRY_ROLE_ORIGINATOR, Addresses: []string{otherAddr}}, exp: "invalid signer: decoding bech32"},
		{name: "nil key", msg: MsgRevokeRole{Signer: validAddr, Key: nil, Role: RegistryRole_REGISTRY_ROLE_ORIGINATOR, Addresses: []string{otherAddr}}, exp: "invalid key: registry key cannot be nil"},
		{name: "invalid role", msg: MsgRevokeRole{Signer: validAddr, Key: validKey, Role: RegistryRole_REGISTRY_ROLE_UNSPECIFIED, Addresses: []string{otherAddr}}, exp: "invalid role: cannot be unspecified"},
		{name: "no addresses", msg: MsgRevokeRole{Signer: validAddr, Key: validKey, Role: RegistryRole_REGISTRY_ROLE_ORIGINATOR, Addresses: []string{}}, exp: "invalid addresses: addresses cannot be empty"},
		{name: "bad address", msg: MsgRevokeRole{Signer: validAddr, Key: validKey, Role: RegistryRole_REGISTRY_ROLE_ORIGINATOR, Addresses: []string{"bad"}}, exp: "invalid addresses: decoding bech32"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.exp == "" {
				assertions.RequireErrorContents(t, err, nil)
			} else {
				assertions.RequireErrorContents(t, err, []string{tc.exp})
			}
		})
	}
}

func TestMsgSetRoles_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress("setroles_signer________").String()
	otherAddr := sdk.AccAddress("other_addr3___________").String()
	validKey := &RegistryKey{AssetClassId: "aclass", NftId: "nft1"}
	validUpdate := RoleUpdate{Role: RegistryRole_REGISTRY_ROLE_CONTROLLER, Addresses: []string{otherAddr}}

	tests := []struct {
		name string
		msg  MsgSetRoles
		exp  string
	}{
		{name: "valid", msg: MsgSetRoles{Signer: validAddr, Key: validKey, RoleUpdates: []RoleUpdate{validUpdate}}},
		{name: "valid clear role", msg: MsgSetRoles{Signer: validAddr, Key: validKey, RoleUpdates: []RoleUpdate{{Role: RegistryRole_REGISTRY_ROLE_CONTROLLER}}}},
		{name: "empty signer", msg: MsgSetRoles{Signer: "", Key: validKey, RoleUpdates: []RoleUpdate{validUpdate}}, exp: "invalid signer: empty address"},
		{name: "bad signer", msg: MsgSetRoles{Signer: "bad", Key: validKey, RoleUpdates: []RoleUpdate{validUpdate}}, exp: "invalid signer: decoding bech32"},
		{name: "nil key", msg: MsgSetRoles{Signer: validAddr, Key: nil, RoleUpdates: []RoleUpdate{validUpdate}}, exp: "invalid key: registry key cannot be nil"},
		{name: "no role_updates", msg: MsgSetRoles{Signer: validAddr, Key: validKey, RoleUpdates: []RoleUpdate{}}, exp: "invalid role_updates: at least one role update is required"},
		{name: "unspecified role", msg: MsgSetRoles{Signer: validAddr, Key: validKey, RoleUpdates: []RoleUpdate{{Role: RegistryRole_REGISTRY_ROLE_UNSPECIFIED}}}, exp: "invalid role_updates: 0: role cannot be unspecified"},
		{name: "unknown role enum", msg: MsgSetRoles{Signer: validAddr, Key: validKey, RoleUpdates: []RoleUpdate{{Role: RegistryRole(999), Addresses: []string{otherAddr}}}}, exp: "invalid role_updates: 0: role unknown registry_role enum value: 999"},
		{name: "bad address in update", msg: MsgSetRoles{Signer: validAddr, Key: validKey, RoleUpdates: []RoleUpdate{{Role: RegistryRole_REGISTRY_ROLE_CONTROLLER, Addresses: []string{"bad"}}}}, exp: "invalid role_updates: 0: invalid address"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.exp == "" {
				assertions.RequireErrorContents(t, err, nil)
			} else {
				assertions.RequireErrorContents(t, err, []string{tc.exp})
			}
		})
	}
}

func TestMsgProposeRoleChange_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress("propose_signer_________").String()
	otherAddr := sdk.AccAddress("propose_target_________").String()
	validKey := &RegistryKey{AssetClassId: "aclass", NftId: "nft1"}
	validUpdate := RoleUpdate{Role: RegistryRole_REGISTRY_ROLE_CONTROLLER, Addresses: []string{otherAddr}}

	tests := []struct {
		name string
		msg  MsgProposeRoleChange
		exp  string
	}{
		{name: "valid", msg: MsgProposeRoleChange{Signer: validAddr, Key: validKey, RoleUpdates: []RoleUpdate{validUpdate}}},
		{name: "valid clear role", msg: MsgProposeRoleChange{Signer: validAddr, Key: validKey, RoleUpdates: []RoleUpdate{{Role: RegistryRole_REGISTRY_ROLE_CONTROLLER}}}},
		{name: "empty signer", msg: MsgProposeRoleChange{Signer: "", Key: validKey, RoleUpdates: []RoleUpdate{validUpdate}}, exp: "invalid signer: empty address"},
		{name: "bad signer", msg: MsgProposeRoleChange{Signer: "bad", Key: validKey, RoleUpdates: []RoleUpdate{validUpdate}}, exp: "invalid signer: decoding bech32"},
		{name: "nil key", msg: MsgProposeRoleChange{Signer: validAddr, Key: nil, RoleUpdates: []RoleUpdate{validUpdate}}, exp: "invalid key: registry key cannot be nil"},
		{name: "no role_updates", msg: MsgProposeRoleChange{Signer: validAddr, Key: validKey, RoleUpdates: []RoleUpdate{}}, exp: "invalid role_updates: at least one role update is required"},
		{name: "unspecified role", msg: MsgProposeRoleChange{Signer: validAddr, Key: validKey, RoleUpdates: []RoleUpdate{{Role: RegistryRole_REGISTRY_ROLE_UNSPECIFIED}}}, exp: "invalid role_updates: 0: role cannot be unspecified"},
		{name: "unknown role enum", msg: MsgProposeRoleChange{Signer: validAddr, Key: validKey, RoleUpdates: []RoleUpdate{{Role: RegistryRole(999), Addresses: []string{otherAddr}}}}, exp: "invalid role_updates: 0: role unknown registry_role enum value: 999"},
		{name: "bad address in update", msg: MsgProposeRoleChange{Signer: validAddr, Key: validKey, RoleUpdates: []RoleUpdate{{Role: RegistryRole_REGISTRY_ROLE_CONTROLLER, Addresses: []string{"bad"}}}}, exp: "invalid role_updates: 0: invalid address"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.exp == "" {
				assertions.RequireErrorContents(t, err, nil)
			} else {
				assertions.RequireErrorContents(t, err, []string{tc.exp})
			}
		})
	}
}

func TestMsgApproveRoleChange_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress("approve_signer_________").String()

	tests := []struct {
		name string
		msg  MsgApproveRoleChange
		exp  string
	}{
		{name: "valid", msg: MsgApproveRoleChange{Signer: validAddr, ChangeId: "abc123"}},
		{name: "empty signer", msg: MsgApproveRoleChange{Signer: "", ChangeId: "abc123"}, exp: "invalid signer: empty address"},
		{name: "bad signer", msg: MsgApproveRoleChange{Signer: "bad", ChangeId: "abc123"}, exp: "invalid signer: decoding bech32"},
		{name: "empty change_id", msg: MsgApproveRoleChange{Signer: validAddr, ChangeId: ""}, exp: "invalid change_id: change_id is required"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.exp == "" {
				assertions.RequireErrorContents(t, err, nil)
			} else {
				assertions.RequireErrorContents(t, err, []string{tc.exp})
			}
		})
	}
}

func TestMsgUnregisterNFT_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress("unreg_signer___________").String()
	validKey := &RegistryKey{AssetClassId: "aclass", NftId: "nft1"}

	tests := []struct {
		name string
		msg  MsgUnregisterNFT
		exp  string
	}{
		{name: "valid", msg: MsgUnregisterNFT{Signer: validAddr, Key: validKey}},
		{name: "empty signer", msg: MsgUnregisterNFT{Signer: "", Key: validKey}, exp: "invalid signer: empty address"},
		{name: "bad signer", msg: MsgUnregisterNFT{Signer: "bad", Key: validKey}, exp: "invalid signer: decoding bech32"},
		{name: "nil key", msg: MsgUnregisterNFT{Signer: validAddr, Key: nil}, exp: "invalid key: registry key cannot be nil"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.exp == "" {
				assertions.RequireErrorContents(t, err, nil)
			} else {
				assertions.RequireErrorContents(t, err, []string{tc.exp})
			}
		})
	}
}

func TestMsgRegistryBulkUpdate_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress("bulk_signer____________").String()
	entry := RegistryEntry{Key: &RegistryKey{AssetClassId: "aclass", NftId: "nft1"}, Roles: []RolesEntry{{Role: RegistryRole_REGISTRY_ROLE_SERVICER, Addresses: []string{validAddr}}}}

	// Build over-max entries
	over := make([]RegistryEntry, MaxRegistryBulkEntries+1)
	for i := range over {
		over[i] = entry
	}

	tests := []struct {
		name string
		msg  MsgRegistryBulkUpdate
		exp  string
	}{
		{name: "valid", msg: MsgRegistryBulkUpdate{Signer: validAddr, Entries: []RegistryEntry{entry}}},
		{name: "empty signer", msg: MsgRegistryBulkUpdate{Signer: "", Entries: []RegistryEntry{entry}}, exp: "invalid signer: empty address"},
		{name: "bad signer", msg: MsgRegistryBulkUpdate{Signer: "bad", Entries: []RegistryEntry{entry}}, exp: "invalid signer: decoding bech32"},
		{name: "no entries", msg: MsgRegistryBulkUpdate{Signer: validAddr, Entries: []RegistryEntry{}}, exp: "entries cannot be empty"},
		{name: "too many entries", msg: MsgRegistryBulkUpdate{Signer: validAddr, Entries: over}, exp: "entries cannot be empty or greater than"},
		{name: "invalid entry", msg: MsgRegistryBulkUpdate{Signer: validAddr, Entries: []RegistryEntry{{Key: nil}}}, exp: "invalid entry: 0: key: registry key cannot be nil"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.exp == "" {
				assertions.RequireErrorContents(t, err, nil)
			} else {
				assertions.RequireErrorContents(t, err, []string{tc.exp})
			}
		})
	}
}
