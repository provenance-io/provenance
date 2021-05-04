package types

import (
	"encoding/hex"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	p8e "github.com/provenance-io/provenance/x/metadata/types/p8e"
	"github.com/stretchr/testify/require"
)

func ownerPartyList(addresses ...string) []Party {
	retval := make([]Party, len(addresses))
	for i, addr := range addresses {
		retval[i] = Party{Address: addr, Role: PartyType_PARTY_TYPE_OWNER}
	}
	return retval
}

func TestWriteScopeRoute(t *testing.T) {
	var scope = NewScope(
		ScopeMetadataAddress(uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")),
		ScopeSpecMetadataAddress(uuid.MustParse("22fc17a6-40dd-4d68-a95b-ec94e7572a09")),
		ownerPartyList("data_owner"),
		[]string{"data_accessor"},
		"value_owner",
	)
	var msg = NewMsgWriteScopeRequest(*scope, []string{})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "write_scope_request")
	yaml := `scope:
  scope_id: scope1qzxcpvj6czy5g354dews3nlruxjsahhnsp
  specification_id: scopespec1qs30c9axgrw5669ft0kffe6h9gysfe58v3
  owners:
  - address: data_owner
    role: 5
  data_access:
  - data_accessor
  value_owner_address: value_owner
signers: []
scope_uuid: ""
spec_uuid: ""
`
	require.Equal(t, yaml, msg.String())
	require.Equal(t, "{\"type\":\"provenance/metadata/WriteScopeRequest\",\"value\":{\"scope\":{\"data_access\":[\"data_accessor\"],\"owners\":[{\"address\":\"data_owner\",\"role\":5}],\"scope_id\":\"scope1qzxcpvj6czy5g354dews3nlruxjsahhnsp\",\"specification_id\":\"scopespec1qs30c9axgrw5669ft0kffe6h9gysfe58v3\",\"value_owner_address\":\"value_owner\"}}}", string(msg.GetSignBytes()))
}

func TestWriteScopeValidation(t *testing.T) {
	var scope = NewScope(
		ScopeMetadataAddress(uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")),
		ScopeSpecMetadataAddress(uuid.MustParse("22fc17a6-40dd-4d68-a95b-ec94e7572a09")),
		ownerPartyList("data_owner"),
		[]string{"data_accessor"},
		"value_owner",
	)
	var msg = NewMsgWriteScopeRequest(*scope, []string{"invalid"})
	err := msg.ValidateBasic()
	require.EqualError(t, err, "invalid owner on scope: decoding bech32 failed: invalid index of 1")
	require.Panics(t, func() { msg.GetSigners() }, "panics due to invalid addresses")

	err = msg.Scope.ValidateBasic()
	require.Error(t, err, "invalid addresses")
	require.Equal(t, "invalid owner on scope: decoding bech32 failed: invalid index of 1", err.Error())

	msg.Scope = *NewScope(
		ScopeMetadataAddress(uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")),
		ScopeSpecMetadataAddress(uuid.MustParse("22fc17a6-40dd-4d68-a95b-ec94e7572a09")),
		[]Party{},
		[]string{},
		"",
	)
	err = msg.Scope.ValidateBasic()
	require.Error(t, err, "no owners")
	require.Equal(t, "scope must have at least one owner", err.Error())

	msg.Scope = *NewScope(
		ScopeMetadataAddress(uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")),
		ScopeSpecMetadataAddress(uuid.MustParse("22fc17a6-40dd-4d68-a95b-ec94e7572a09")),
		ownerPartyList("cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"),
		[]string{},
		"",
	)
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

	cases := map[string]struct {
		msg      *MsgAddScopeOwnerRequest
		wantErr  bool
		errorMsg string
	}{
		"should fail to validate basic, incorrect scope id type": {
			NewMsgAddScopeOwnerRequest(
				notAScopeId,
				[]Party{{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_OWNER}},
				[]string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"}),
			true,
			fmt.Sprintf("address is not a scope id: %v", notAScopeId.String()),
		},
		"should fail to validate basic, requires at least one owner address": {
			NewMsgAddScopeOwnerRequest(
				actualScopeId,
				[]Party{},
				[]string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"},
			),
			true,
			"at least one owner party is required",
		},
		"should fail to validate basic, incorrect owner address format": {
			NewMsgAddScopeOwnerRequest(
				actualScopeId,
				[]Party{{Address: "notabech32address", Role: PartyType_PARTY_TYPE_OWNER}},
				[]string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"},
			),
			true,
			"owner address is invalid: notabech32address",
		},
		"should fail to validate basic, incorrect party type": {
			NewMsgAddScopeOwnerRequest(
				actualScopeId,
				[]Party{{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_UNSPECIFIED}},
				[]string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"},
			),
			true,
			"invalid party type for owner: cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck",
		},
		"should fail to validate basic, requires at least one signer": {
			NewMsgAddScopeOwnerRequest(
				actualScopeId,
				[]Party{{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_OWNER}},
				[]string{},
			),
			true,
			"at least one signer is required",
		},
		"should successfully validate basic": {
			NewMsgAddScopeOwnerRequest(
				actualScopeId,
				[]Party{{Address: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", Role: PartyType_PARTY_TYPE_OWNER}},
				[]string{"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"},
			),
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

func TestWriteP8eContractSpecValidation(t *testing.T) {

	validInputSpec := p8e.DefinitionSpec{
		Name: "perform_input_checks",
		ResourceLocation: &p8e.Location{Classname: "io.provenance.loan.LoanProtos$PartiesList",
			Ref: &p8e.ProvenanceReference{Hash: "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw=="},
		},
		Type: 1,
	}

	validOutputSpec := p8e.OutputSpec{Spec: &p8e.DefinitionSpec{
		Name: "additional_parties",
		ResourceLocation: &p8e.Location{
			Classname: "io.provenance.loan.LoanProtos$PartiesList",
			Ref: &p8e.ProvenanceReference{
				Hash: "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw==",
			},
		},
		Type: 1,
	},
	}

	validDefinition := p8e.DefinitionSpec{
		Name: "ExampleContract",
		ResourceLocation: &p8e.Location{Classname: "io.provenance.contracts.ExampleContract",
			Ref: &p8e.ProvenanceReference{Hash: "E36eeTUk8GYXGXjIbZTm4s/Dw3G1e42SinH1195t4ekgcXXPhfIpfQaEJ21PTzKhdv6JjhzQJ2kAJXK+TRXmeQ=="},
		},
		Type: 1,
	}

	validContractSpec := p8e.ContractSpec{ConsiderationSpecs: []*p8e.ConsiderationSpec{
		{FuncName: "additionalParties",
			InputSpecs:       []*p8e.DefinitionSpec{&validInputSpec},
			OutputSpec:       &validOutputSpec,
			ResponsibleParty: 1,
		},
	},
		Definition:      &validDefinition,
		InputSpecs:      []*p8e.DefinitionSpec{&validInputSpec},
		PartiesInvolved: []p8e.PartyType{p8e.PartyType_PARTY_TYPE_AFFILIATE},
	}

	msg := NewMsgWriteP8EContractSpecRequest(validContractSpec, []string{})
	err := msg.ValidateBasic()
	require.Error(t, err, "should fail due to signatures < 1")

	msg = NewMsgWriteP8EContractSpecRequest(validContractSpec, []string{"invalid"})
	err = msg.ValidateBasic()
	require.Error(t, err, "should fail in convert validation due to address not being valid")

	msg = NewMsgWriteP8EContractSpecRequest(validContractSpec, []string{"cosmos1s0kcwmhstu6urpp4080qjzatta02y0rarrcgrp"})
	err = msg.ValidateBasic()
	require.NoError(t, err)
}

func TestBindOSLocator(t *testing.T) {
	var bindRequestMsg = NewMsgBindOSLocatorRequest(ObjectStoreLocator{Owner: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", LocatorUri: "http://foo.com"})

	err := bindRequestMsg.ValidateBasic()
	require.NoError(t, err)
	signers := bindRequestMsg.GetSigners()
	route := bindRequestMsg.Route()
	require.Equal(t, "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", signers[0].String())
	require.Equal(t, ModuleName, route)
	require.Equal(t, TypeMsgBindOSLocatorRequest, bindRequestMsg.Type())
	require.Equal(t, "{\"type\":\"provenance/metadata/BindOSLocatorRequest\",\"value\":{\"locator\":{\"locator_uri\":\"http://foo.com\",\"owner\":\"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck\"}}}", string(bindRequestMsg.GetSignBytes()))

}

func TestModifyOSLocator(t *testing.T) {
	var modifyRequest = NewMsgModifyOSLocatorRequest(ObjectStoreLocator{Owner: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", LocatorUri: "http://foo.com"})

	err := modifyRequest.ValidateBasic()
	require.NoError(t, err)
	signers := modifyRequest.GetSigners()
	require.Equal(t, "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", signers[0].String())
	require.Equal(t, ModuleName, modifyRequest.Route())
	require.Equal(t, TypeMsgModifyOSLocatorRequest, modifyRequest.Type())
	require.Equal(t, "{\"type\":\"provenance/metadata/ModifyOSLocatorRequest\",\"value\":{\"locator\":{\"locator_uri\":\"http://foo.com\",\"owner\":\"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck\"}}}", string(modifyRequest.GetSignBytes()))

}

func TestDeleteOSLocator(t *testing.T) {
	var deleteRequest = NewMsgDeleteOSLocatorRequest(ObjectStoreLocator{Owner: "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", LocatorUri: "http://foo.com"})

	err := deleteRequest.ValidateBasic()
	require.NoError(t, err)

	signers := deleteRequest.GetSigners()
	require.Equal(t, "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck", signers[0].String())
	require.Equal(t, ModuleName, deleteRequest.Route())
	require.Equal(t, TypeMsgDeleteOSLocatorRequest, deleteRequest.Type())
	require.Equal(t, "{\"type\":\"provenance/metadata/DeleteOSLocatorRequest\",\"value\":{\"locator\":{\"locator_uri\":\"http://foo.com\",\"owner\":\"cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck\"}}}", string(deleteRequest.GetSignBytes()))

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
