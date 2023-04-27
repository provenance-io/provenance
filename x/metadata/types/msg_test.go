package types

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func ownerPartyList(addresses ...string) []Party {
	retval := make([]Party, len(addresses))
	for i, addr := range addresses {
		retval[i] = Party{Address: addr, Role: PartyType_PARTY_TYPE_OWNER}
	}
	return retval
}

func TestAllGetSigners(t *testing.T) {
	// TODO[1329]: Write TestAllGetSigners
	t.Fatalf("not yet written")
}

func TestWriteScopeRoute(t *testing.T) {
	var scope = &Scope{
		ScopeId:           ScopeMetadataAddress(uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")),
		SpecificationId:   ScopeSpecMetadataAddress(uuid.MustParse("22fc17a6-40dd-4d68-a95b-ec94e7572a09")),
		Owners:            ownerPartyList("data_owner"),
		DataAccess:        []string{"data_accessor"},
		ValueOwnerAddress: "value_owner",
	}
	var msg = NewMsgWriteScopeRequest(*scope, []string{})

	require.Equal(t, sdk.MsgTypeURL(msg), "/provenance.metadata.v1.MsgWriteScopeRequest")
	expectedYaml := `scope:
  scope_id: scope1qzxcpvj6czy5g354dews3nlruxjsahhnsp
  specification_id: scopespec1qs30c9axgrw5669ft0kffe6h9gysfe58v3
  owners:
  - address: data_owner
    role: 5
    optional: false
  data_access:
  - data_accessor
  value_owner_address: value_owner
  require_party_rollup: false
signers: []
scope_uuid: ""
spec_uuid: ""
`
	bz, err := yaml.Marshal(msg)
	require.NoError(t, err, "yaml.Marshal(msg)")
	assert.Equal(t, expectedYaml, string(bz), "scope as yaml")

	bz, err = ModuleCdc.MarshalJSON(msg)
	require.NoError(t, err, "ModuleCdc.MarshalJSON(msg)")
	assert.Equal(t, "{\"scope\":{\"scope_id\":\"scope1qzxcpvj6czy5g354dews3nlruxjsahhnsp\",\"specification_id\":\"scopespec1qs30c9axgrw5669ft0kffe6h9gysfe58v3\",\"owners\":[{\"address\":\"data_owner\",\"role\":\"PARTY_TYPE_OWNER\",\"optional\":false}],\"data_access\":[\"data_accessor\"],\"value_owner_address\":\"value_owner\",\"require_party_rollup\":false},\"signers\":[],\"scope_uuid\":\"\",\"spec_uuid\":\"\"}", string(bz))
}

func TestWriteScopeValidation(t *testing.T) {
	var scope = &Scope{
		ScopeId:           ScopeMetadataAddress(uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")),
		SpecificationId:   ScopeSpecMetadataAddress(uuid.MustParse("22fc17a6-40dd-4d68-a95b-ec94e7572a09")),
		Owners:            ownerPartyList("data_owner"),
		DataAccess:        []string{"data_accessor"},
		ValueOwnerAddress: "value_owner",
	}
	var msg = NewMsgWriteScopeRequest(*scope, []string{"invalid"})
	err := msg.ValidateBasic()
	require.EqualError(t, err, "invalid scope owners: invalid party address [data_owner]: decoding bech32 failed: invalid separator index -1")
	require.Panics(t, func() { msg.GetSigners() }, "panics due to invalid addresses")

	err = msg.Scope.ValidateBasic()
	require.Error(t, err, "invalid addresses")
	require.Equal(t, "invalid scope owners: invalid party address [data_owner]: decoding bech32 failed: invalid separator index -1", err.Error())

	msg.Scope = Scope{
		ScopeId:         ScopeMetadataAddress(uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")),
		SpecificationId: ScopeSpecMetadataAddress(uuid.MustParse("22fc17a6-40dd-4d68-a95b-ec94e7572a09")),
		Owners:          []Party{},
		DataAccess:      []string{},
	}
	err = msg.Scope.ValidateBasic()
	require.Error(t, err, "no owners")
	require.Equal(t, "invalid scope owners: at least one party is required", err.Error())

	msg.Scope = Scope{
		ScopeId:         ScopeMetadataAddress(uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")),
		SpecificationId: ScopeSpecMetadataAddress(uuid.MustParse("22fc17a6-40dd-4d68-a95b-ec94e7572a09")),
		Owners:          ownerPartyList("cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"),
		DataAccess:      []string{},
	}
	msg.Signers = []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}
	err = msg.Scope.ValidateBasic()
	require.NoError(t, err, "valid add scope request")
	requiredSigners := msg.GetSigners()
	require.Equal(t, 1, len(requiredSigners))
	x, err := hex.DecodeString("85EA54E8598B27EC37EAEEEEA44F1E78A9B5E671")
	require.NoError(t, err)
	require.Equal(t, sdk.AccAddress(x), requiredSigners[0])
}

func TestAddScopeDataAccessValidateBasic(t *testing.T) {
	notAScopeId := RecordMetadataAddress(uuid.New(), "recordname")
	actualScopeId := ScopeMetadataAddress(uuid.New())

	cases := map[string]struct {
		msg      *MsgAddScopeDataAccessRequest
		wantErr  bool
		errorMsg string
	}{
		"should fail to validate basic, incorrect scope id type": {
			NewMsgAddScopeDataAccessRequest(notAScopeId, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			fmt.Sprintf("address is not a scope id: %v", notAScopeId.String()),
		},
		"should fail to validate basic, requires at least one data access address": {
			NewMsgAddScopeDataAccessRequest(actualScopeId, []string{}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			"at least one data access address is required",
		},
		"should fail to validate basic, incorrect data access address format": {
			NewMsgAddScopeDataAccessRequest(actualScopeId, []string{"notabech32address"}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			"data access address is invalid: notabech32address",
		},
		"should fail to validate basic, requires at least one signer": {
			NewMsgAddScopeDataAccessRequest(actualScopeId, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}, []string{}),
			true,
			"at least one signer is required",
		},
		"should successfully validate basic": {
			NewMsgAddScopeDataAccessRequest(actualScopeId, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			false,
			"",
		},
	}

	for n, tc := range cases {
		tc := tc

		t.Run(n, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.wantErr {
				require.Error(t, err)
				require.Equal(t, tc.errorMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeleteScopeDataAccessValidateBasic(t *testing.T) {
	notAScopeId := RecordMetadataAddress(uuid.New(), "recordname")
	actualScopeId := ScopeMetadataAddress(uuid.New())

	cases := map[string]struct {
		msg      *MsgDeleteScopeDataAccessRequest
		wantErr  bool
		errorMsg string
	}{
		"should fail to validate basic, incorrect scope id type": {
			NewMsgDeleteScopeDataAccessRequest(notAScopeId, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			fmt.Sprintf("address is not a scope id: %v", notAScopeId.String()),
		},
		"should fail to validate basic, requires at least one data access address": {
			NewMsgDeleteScopeDataAccessRequest(actualScopeId, []string{}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			"at least one data access address is required",
		},
		"should fail to validate basic, incorrect data access address format": {
			NewMsgDeleteScopeDataAccessRequest(actualScopeId, []string{"notabech32address"}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			"data access address is invalid: notabech32address",
		},
		"should fail to validate basic, requires at least one signer": {
			NewMsgDeleteScopeDataAccessRequest(actualScopeId, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}, []string{}),
			true,
			"at least one signer is required",
		},
		"should successfully validate basic": {
			NewMsgDeleteScopeDataAccessRequest(actualScopeId, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			false,
			"",
		},
	}

	for n, tc := range cases {
		tc := tc

		t.Run(n, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.wantErr {
				require.Error(t, err)
				require.Equal(t, tc.errorMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAddScopeOwnersValidateBasic(t *testing.T) {
	notAScopeId := RecordMetadataAddress(uuid.New(), "recordname")
	actualScopeId := ScopeMetadataAddress(uuid.New())

	cases := []struct {
		name     string
		msg      *MsgAddScopeOwnerRequest
		errorMsg string
	}{
		{
			"should fail to validate basic, incorrect scope id type",
			NewMsgAddScopeOwnerRequest(
				notAScopeId,
				[]Party{{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_OWNER}},
				[]string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			fmt.Sprintf("address is not a scope id: %v", notAScopeId.String()),
		},
		{
			"should fail to validate basic, requires at least one owner address",
			NewMsgAddScopeOwnerRequest(
				actualScopeId,
				[]Party{},
				[]string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"},
			),
			"invalid owners: at least one party is required",
		},
		{
			"should fail to validate basic, incorrect owner address format",
			NewMsgAddScopeOwnerRequest(
				actualScopeId,
				[]Party{{Address: "notabech32address", Role: PartyType_PARTY_TYPE_OWNER}},
				[]string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"},
			),
			"invalid owners: invalid party address [notabech32address]: decoding bech32 failed: invalid separator index -1",
		},
		{
			"should fail to validate basic, incorrect party type",
			NewMsgAddScopeOwnerRequest(
				actualScopeId,
				[]Party{{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_UNSPECIFIED}},
				[]string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"},
			),
			"invalid owners: invalid party type for party cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck",
		},
		{
			"should fail to validate basic, requires at least one signer",
			NewMsgAddScopeOwnerRequest(
				actualScopeId,
				[]Party{{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_OWNER}},
				[]string{},
			),
			"at least one signer is required",
		},
		{
			"should successfully validate basic",
			NewMsgAddScopeOwnerRequest(
				actualScopeId,
				[]Party{{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_OWNER}},
				[]string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"},
			),
			"",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if len(tc.errorMsg) > 0 {
				require.EqualError(t, err, tc.errorMsg, "MsgAddScopeOwnerRequest.ValidateBasic expected error")
			} else {
				require.NoError(t, err, "MsgAddScopeOwnerRequest.ValidateBasic unexpected error")
			}
		})
	}
}

func TestDeleteScopeOwnerValidateBasic(t *testing.T) {
	notAScopeId := RecordMetadataAddress(uuid.New(), "recordname")
	actualScopeId := ScopeMetadataAddress(uuid.New())

	cases := map[string]struct {
		msg      *MsgDeleteScopeOwnerRequest
		wantErr  bool
		errorMsg string
	}{
		"should fail to validate basic, incorrect scope id type": {
			NewMsgDeleteScopeOwnerRequest(notAScopeId, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			fmt.Sprintf("address is not a scope id: %v", notAScopeId.String()),
		},
		"should fail to validate basic, requires at least one owner address": {
			NewMsgDeleteScopeOwnerRequest(actualScopeId, []string{}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			"at least one owner address is required",
		},
		"should fail to validate basic, incorrect data access address format": {
			NewMsgDeleteScopeOwnerRequest(actualScopeId, []string{"notabech32address"}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			"owner address is invalid: notabech32address",
		},
		"should fail to validate basic, requires at least one signer": {
			NewMsgDeleteScopeOwnerRequest(actualScopeId, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}, []string{}),
			true,
			"at least one signer is required",
		},
		"should successfully validate basic": {
			NewMsgDeleteScopeOwnerRequest(actualScopeId, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			false,
			"",
		},
	}

	for n, tc := range cases {
		tc := tc

		t.Run(n, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.wantErr {
				require.Error(t, err)
				require.Equal(t, tc.errorMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgUpdateValueOwnersRequest_ValidateBasic(t *testing.T) {
	// TODO[1329]: Write TestMsgUpdateValueOwnersRequest_ValidateBasic
	t.Fatalf("not yet written")
}

func TestMsgAddContractSpecToScopeSpecRequestValidateBasic(t *testing.T) {
	contractSpecID := ContractSpecMetadataAddress(uuid.New())
	scopeSpecID := ScopeSpecMetadataAddress(uuid.New())

	cases := map[string]struct {
		msg      *MsgAddContractSpecToScopeSpecRequest
		wantErr  bool
		errorMsg string
	}{
		"should fail to validate basic, incorrect contract spec id type": {
			NewMsgAddContractSpecToScopeSpecRequest(scopeSpecID, scopeSpecID, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			fmt.Sprintf("address is not a contract specification id: %v", scopeSpecID.String()),
		},
		"should fail to validate basic, incorrect scope spec id type": {
			NewMsgAddContractSpecToScopeSpecRequest(contractSpecID, contractSpecID, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			fmt.Sprintf("address is not a scope specification id: %v", contractSpecID.String()),
		},
		"should fail to validate basic, requires at least one signer": {
			NewMsgAddContractSpecToScopeSpecRequest(contractSpecID, scopeSpecID, []string{}),
			true,
			"at least one signer is required",
		},
		"should successfully validate basic": {
			NewMsgAddContractSpecToScopeSpecRequest(contractSpecID, scopeSpecID, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			false,
			"",
		},
	}

	for n, tc := range cases {
		tc := tc

		t.Run(n, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.wantErr {
				require.Error(t, err)
				require.Equal(t, tc.errorMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgDeleteContractSpecFromScopeSpecRequestValidateBasic(t *testing.T) {
	contractSpecID := ContractSpecMetadataAddress(uuid.New())
	scopeSpecID := ScopeSpecMetadataAddress(uuid.New())

	cases := map[string]struct {
		msg      *MsgDeleteContractSpecFromScopeSpecRequest
		wantErr  bool
		errorMsg string
	}{
		"should fail to validate basic, incorrect contract spec id type": {
			NewMsgDeleteContractSpecFromScopeSpecRequest(scopeSpecID, scopeSpecID, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			fmt.Sprintf("address is not a contract specification id: %v", scopeSpecID.String()),
		},
		"should fail to validate basic, incorrect scope spec id type": {
			NewMsgDeleteContractSpecFromScopeSpecRequest(contractSpecID, contractSpecID, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			fmt.Sprintf("address is not a scope specification id: %v", contractSpecID.String()),
		},
		"should fail to validate basic, requires at least one signer": {
			NewMsgDeleteContractSpecFromScopeSpecRequest(contractSpecID, scopeSpecID, []string{}),
			true,
			"at least one signer is required",
		},
		"should successfully validate basic": {
			NewMsgDeleteContractSpecFromScopeSpecRequest(contractSpecID, scopeSpecID, []string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			false,
			"",
		},
	}

	for n, tc := range cases {
		tc := tc

		t.Run(n, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.wantErr {
				require.Error(t, err)
				require.Equal(t, tc.errorMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBindOSLocator(t *testing.T) {
	var bindRequestMsg = NewMsgBindOSLocatorRequest(ObjectStoreLocator{Owner: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", LocatorUri: "http://foo.com"})

	err := bindRequestMsg.ValidateBasic()
	require.NoError(t, err)
	signers := bindRequestMsg.GetSigners()
	require.Equal(t, "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", signers[0].String())
	require.Equal(t, "/provenance.metadata.v1.MsgBindOSLocatorRequest", sdk.MsgTypeURL(bindRequestMsg))

	bz, _ := ModuleCdc.MarshalJSON(bindRequestMsg)
	require.Equal(t, "{\"locator\":{\"owner\":\"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck\",\"locator_uri\":\"http://foo.com\",\"encryption_key\":\"\"}}", string(bz))
}

func TestModifyOSLocator(t *testing.T) {
	var modifyRequest = NewMsgModifyOSLocatorRequest(ObjectStoreLocator{Owner: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", LocatorUri: "http://foo.com"})

	err := modifyRequest.ValidateBasic()
	require.NoError(t, err)
	signers := modifyRequest.GetSigners()
	require.Equal(t, "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", signers[0].String())
	require.Equal(t, "/provenance.metadata.v1.MsgModifyOSLocatorRequest", sdk.MsgTypeURL(modifyRequest))

	bz, _ := ModuleCdc.MarshalJSON(modifyRequest)
	require.Equal(t, "{\"locator\":{\"owner\":\"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck\",\"locator_uri\":\"http://foo.com\",\"encryption_key\":\"\"}}", string(bz))
}

func TestDeleteOSLocator(t *testing.T) {
	var deleteRequest = NewMsgDeleteOSLocatorRequest(ObjectStoreLocator{Owner: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", LocatorUri: "http://foo.com"})

	err := deleteRequest.ValidateBasic()
	require.NoError(t, err)

	signers := deleteRequest.GetSigners()
	require.Equal(t, "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", signers[0].String())
	require.Equal(t, "/provenance.metadata.v1.MsgDeleteOSLocatorRequest", sdk.MsgTypeURL(deleteRequest))

	bz, _ := ModuleCdc.MarshalJSON(deleteRequest)
	require.Equal(t, "{\"locator\":{\"owner\":\"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck\",\"locator_uri\":\"http://foo.com\",\"encryption_key\":\"\"}}", string(bz))
}

func TestBindOSLocatorInvalid(t *testing.T) {
	var bindRequestMsg = NewMsgBindOSLocatorRequest(ObjectStoreLocator{Owner: "vamonos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", LocatorUri: "http://foo.com"})

	err := bindRequestMsg.ValidateBasic()
	require.Error(t, err)
}

func TestBindOSLocatorInvalidAddr(t *testing.T) {
	var bindRequestMsg = NewMsgBindOSLocatorRequest(ObjectStoreLocator{Owner: "", LocatorUri: "http://foo.com"})

	err := bindRequestMsg.ValidateBasic()
	require.Error(t, err)
}

func TestBindOSLocatorInvalidURI(t *testing.T) {
	var bindRequestMsg = NewMsgBindOSLocatorRequest(ObjectStoreLocator{Owner: "", LocatorUri: "foo://foo.com"})

	err := bindRequestMsg.ValidateBasic()
	require.Error(t, err)
}

// TestPrintMessageTypeStrings just prints out all the MsgTypeURLs.
// The output can be copy/pasted into the const area in msg.go
func TestPrintMessageTypeStrings(t *testing.T) {
	messageTypes := []sdk.Msg{
		&MsgWriteScopeRequest{},
		&MsgDeleteScopeRequest{},
		&MsgAddScopeDataAccessRequest{},
		&MsgDeleteScopeDataAccessRequest{},
		&MsgAddScopeOwnerRequest{},
		&MsgDeleteScopeOwnerRequest{},
		&MsgUpdateValueOwnersRequest{},
		&MsgWriteSessionRequest{},
		&MsgWriteRecordRequest{},
		&MsgDeleteRecordRequest{},
		&MsgWriteScopeSpecificationRequest{},
		&MsgDeleteScopeSpecificationRequest{},
		&MsgWriteContractSpecificationRequest{},
		&MsgDeleteContractSpecificationRequest{},
		&MsgAddContractSpecToScopeSpecRequest{},
		&MsgDeleteContractSpecFromScopeSpecRequest{},
		&MsgWriteRecordSpecificationRequest{},
		&MsgDeleteRecordSpecificationRequest{},
		&MsgBindOSLocatorRequest{},
		&MsgDeleteOSLocatorRequest{},
		&MsgModifyOSLocatorRequest{},
		// add  any new messages here
	}

	fmt.Println("const (")
	for _, msg := range messageTypes {
		varname := fmt.Sprintf("%T", msg)
		lastDot := strings.LastIndex(varname, ".")
		if lastDot >= 0 && lastDot+1 < len(varname) {
			varname = varname[lastDot+1:]
		}
		expected := sdk.MsgTypeURL(msg)
		// Note: 41 comes from the length of the longest Msg name: MsgDeleteContractSpecFromScopeSpecRequest
		fmt.Printf("\tTypeURL%-41s = %q\n", varname, expected)
	}
	fmt.Println(")")
}
