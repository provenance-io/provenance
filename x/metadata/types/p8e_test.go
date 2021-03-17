package types

import (
	"testing"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/google/uuid"

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

func createDefinitionSpec(name string, classname string, reference p8e.ProvenanceReference, defType int) p8e.DefinitionSpec {
	return p8e.DefinitionSpec{
		Name: name,
		ResourceLocation: &p8e.Location{Classname: classname,
			Ref: &reference,
		},
		Type: 1,
	}
}

func (s *P8eTestSuite) TestConvertP8eContractSpec() {
	validDefSpec := createDefinitionSpec("perform_input_checks", "io.provenance.loan.LoanProtos$PartiesList", p8e.ProvenanceReference{Hash: "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw=="}, 1)
	validDefSpecUUID := createDefinitionSpec("perform_input_checks", "io.provenance.loan.LoanProtos$PartiesList", p8e.ProvenanceReference{ScopeUuid: &p8e.UUID{Value: uuid.New().String()}, Name: "recordname"}, 1)

	invalidDefSpecNoName := createDefinitionSpec("", "io.provenance.loan.LoanProtos$PartiesList", p8e.ProvenanceReference{Hash: "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw=="}, 1)
	invalidDefSpecNoClass := createDefinitionSpec("perform_action", "", p8e.ProvenanceReference{Hash: "Adv+huolGTKofYCR0dw5GHm/R7sUWOwF32XR8r8r9kDy4il5U/LApxOWYHb05jhK4+eY4YzRMRiWcxU3Lx0+Mw=="}, 1)
	invalidDefSpecUUID := createDefinitionSpec("perform_input_checks", "io.provenance.loan.LoanProtos$PartiesList", p8e.ProvenanceReference{ScopeUuid: &p8e.UUID{Value: "not-a-uuid"}, Name: "recordname"}, 1)
	invalidDefSpecUUIDNoName := createDefinitionSpec("perform_input_checks", "io.provenance.loan.LoanProtos$PartiesList", p8e.ProvenanceReference{ScopeUuid: &p8e.UUID{Value: uuid.New().String()}}, 1)
	invalidDefSpecHash := createDefinitionSpec("ExampleContract", "io.provenance.contracts.ExampleContract", p8e.ProvenanceReference{Hash: "should fail to decode this"}, 1)

	cases := map[string]struct {
		v39CSpec p8e.ContractSpec
		signers  []string
		wantErr  bool
		errorMsg string
	}{
		"should convert a contract spec successfully": {
			createContractSpec([]*p8e.DefinitionSpec{&validDefSpec}, p8e.OutputSpec{Spec: &validDefSpec}, validDefSpec),
			[]string{s.user1},
			false,
			"",
		},
		"should convert a contract spec successfully with uuid input spec": {
			createContractSpec([]*p8e.DefinitionSpec{&validDefSpecUUID}, p8e.OutputSpec{Spec: &validDefSpec}, validDefSpec),
			[]string{s.user1},
			false,
			"",
		},
		"should fail to validate basic on contract specification": {
			createContractSpec([]*p8e.DefinitionSpec{&validDefSpec}, p8e.OutputSpec{Spec: &validDefSpec}, validDefSpec),
			[]string{},
			true,
			"invalid owner addresses count (expected > 0 got: 0)",
		},
		"should fail to validatebasic on input specification": {
			createContractSpec([]*p8e.DefinitionSpec{&invalidDefSpecNoName}, p8e.OutputSpec{Spec: &validDefSpec}, validDefSpec),
			[]string{s.user1},
			true,
			"input specification name cannot be empty",
		},
		"should fail to validatebasic on record specification": {
			createContractSpec([]*p8e.DefinitionSpec{&validDefSpec}, p8e.OutputSpec{Spec: &invalidDefSpecNoClass}, validDefSpec),
			[]string{s.user1},
			true,
			"record specification type name cannot be empty",
		},
		"should fail to decode resource location": {
			createContractSpec([]*p8e.DefinitionSpec{&validDefSpec}, p8e.OutputSpec{Spec: &validDefSpec}, invalidDefSpecHash),
			[]string{s.user1},
			true,
			"illegal base64 data at input byte 6",
		},
		"should fail to decode input spec uuid": {
			createContractSpec([]*p8e.DefinitionSpec{&invalidDefSpecUUID}, p8e.OutputSpec{Spec: &validDefSpec}, validDefSpec),
			[]string{s.user1},
			true,
			"invalid UUID length: 10",
		},
		"should fail to find name on record with uuid": {
			createContractSpec([]*p8e.DefinitionSpec{&invalidDefSpecUUIDNoName}, p8e.OutputSpec{Spec: &validDefSpec}, validDefSpec),
			[]string{s.user1},
			true,
			"must have a value for record name",
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
