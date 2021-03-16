package types

import (
	"testing"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	p8e "github.com/provenance-io/provenance/x/metadata/types/p8e"
	"github.com/stretchr/testify/suite"
)

type P8eTestSuite struct {
	suite.Suite

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress
}

func (s *P8eTestSuite) SetupTest() {
	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()
}

func TestP8eTestSuite(t *testing.T) {
	suite.Run(t, new(P8eTestSuite))
}

func createContractSpec(inputSpecs []*p8e.DefinitionSpec, outputSpec p8e.OutputSpec, definitionSpec p8e.DefinitionSpec) p8e.ContractSpec {
	return p8e.ContractSpec{ConsiderationSpecs: []*p8e.ConsiderationSpec{
		{FuncName: "additionalParties",
			InputSpecs:       inputSpecs,
			OutputSpec:       &outputSpec,
			ResponsibleParty: 1,
		},
	},
		Definition:      &definitionSpec,
		InputSpecs:      inputSpecs,
		PartiesInvolved: []p8e.PartyType{p8e.PartyType_PARTY_TYPE_AFFILIATE},
	}
}
func (s *P8eTestSuite) TestConvertP8eContractSpec() {
	validInputSpec := p8e.DefinitionSpec{
		Name: "perform_input_checks",
		ResourceLocation: &p8e.Location{Classname: "io.provenance.loan.LoanProtos$PartiesList",
			Ref: &p8e.ProvenanceReference{Hash: "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw=="},
		},
		Type: 1,
	}
	invalidInputSpec := p8e.DefinitionSpec{
		Name: "",
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

	invalidOutputSpec := p8e.OutputSpec{Spec: &p8e.DefinitionSpec{
		Name: "additional_parties",
		ResourceLocation: &p8e.Location{
			Classname: "",
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

	cases := map[string]struct {
		v39CSpec p8e.ContractSpec
		signers  []string
		wantErr  bool
		errorMsg string
	}{
		"should convert a contract spec successfully": {
			createContractSpec([]*p8e.DefinitionSpec{&validInputSpec}, validOutputSpec, validDefinition),
			[]string{s.user1},
			false,
			"",
		},
		"should fail to validate basic on contract specification": {
			createContractSpec([]*p8e.DefinitionSpec{&validInputSpec}, validOutputSpec, validDefinition),
			[]string{},
			true,
			"invalid owner addresses count (expected > 0 got: 0)",
		},
		"should fail to validatebasic on input specification": {
			createContractSpec([]*p8e.DefinitionSpec{&invalidInputSpec}, validOutputSpec, validDefinition),
			[]string{s.user1},
			true,
			"input specification name cannot be empty",
		},
		"should fail to validatebasic on record specification": {
			createContractSpec([]*p8e.DefinitionSpec{&validInputSpec}, invalidOutputSpec, validDefinition),
			[]string{s.user1},
			true,
			"record specification type name cannot be empty",
		},
	}

	for n, tc := range cases {
		tc := tc

		s.Run(n, func() {
			_, _, err := ConvertP8eContractSpec(&tc.v39CSpec, tc.signers)
			if tc.wantErr {
				s.Error(err)
				s.Equal(tc.errorMsg, err.Error())
			} else {
				s.NoError(err)
			}
		})
	}

}
