package types

import (
	"encoding/hex"
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

func TestAddScopeRoute(t *testing.T) {
	var scope = NewScope(
		ScopeMetadataAddress(uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")),
		ScopeSpecMetadataAddress(uuid.MustParse("22fc17a6-40dd-4d68-a95b-ec94e7572a09")),
		ownerPartyList("data_owner"),
		[]string{"data_accessor"},
		"value_owner",
	)
	var msg = NewMsgAddScopeRequest(*scope, []string{})

	require.Equal(t, msg.Route(), RouterKey)
	require.Equal(t, msg.Type(), "add_scope_request")
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
`
	require.Equal(t, yaml, msg.String())
	require.Equal(t, "{\"type\":\"provenance/metadata/AddScopeRequest\",\"value\":{\"scope\":{\"data_access\":[\"data_accessor\"],\"owners\":[{\"address\":\"data_owner\",\"role\":5}],\"scope_id\":\"scope1qzxcpvj6czy5g354dews3nlruxjsahhnsp\",\"specification_id\":\"scopespec1qs30c9axgrw5669ft0kffe6h9gysfe58v3\",\"value_owner_address\":\"value_owner\"}}}", string(msg.GetSignBytes()))
}

func TestAddScopeValidation(t *testing.T) {
	var scope = NewScope(
		ScopeMetadataAddress(uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")),
		ScopeSpecMetadataAddress(uuid.MustParse("22fc17a6-40dd-4d68-a95b-ec94e7572a09")),
		ownerPartyList("data_owner"),
		[]string{"data_accessor"},
		"value_owner",
	)
	var msg = NewMsgAddScopeRequest(*scope, []string{"invalid"})
	err := msg.ValidateBasic()
	require.Panics(t, func() { msg.GetSigners() }, "panics due to invalid addresses")
	require.Error(t, err, "invalid addresses")
	require.Equal(t, "invalid owner on scope: decoding bech32 failed: invalid index of 1", err.Error())

	msg.Scope = *NewScope(
		ScopeMetadataAddress(uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")),
		ScopeSpecMetadataAddress(uuid.MustParse("22fc17a6-40dd-4d68-a95b-ec94e7572a09")),
		[]Party{},
		[]string{},
		"",
	)
	err = msg.ValidateBasic()
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
	err = msg.ValidateBasic()
	require.NoError(t, err, "valid add scope request")
	requiredSigners := msg.GetSigners()
	require.Equal(t, 1, len(requiredSigners))
	hex, err := hex.DecodeString("85EA54E8598B27EC37EAEEEEA44F1E78A9B5E671")
	require.NoError(t, err)
	require.Equal(t, sdk.AccAddress(hex), requiredSigners[0])
}

func TestAddP8eContractSpecValidation(t *testing.T) {

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

	msg := NewMsgAddP8EContractSpecRequest(validContractSpec, []string{})
	err := msg.ValidateBasic()
	require.Error(t, err, "should fail due to signatures < 1")

	msg = NewMsgAddP8EContractSpecRequest(validContractSpec, []string{"invalid"})
	err = msg.ValidateBasic()
	require.Error(t, err, "should fail in convert validation due to address not being valid")

	msg = NewMsgAddP8EContractSpecRequest(validContractSpec, []string{"cosmos1s0kcwmhstu6urpp4080qjzatta02y0rarrcgrp"})
	err = msg.ValidateBasic()
	require.NoError(t, err)
}
