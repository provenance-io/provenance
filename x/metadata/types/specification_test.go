package types

import (
	"encoding/hex"
	"fmt"
	math_bits "math/bits"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	specTestHexString = "85EA54E8598B27EC37EAEEEEA44F1E78A9B5E671"
	specTestPubHex, _ = hex.DecodeString(specTestHexString)
	specTestAddr      = sdk.AccAddress(specTestPubHex)
	specTestBech32    = specTestAddr.String() // cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck

	specTest2HexString = "120359A88C1ACD69505FD0C584E837021E848B89"
	specTest2PubHex, _ = hex.DecodeString(specTest2HexString)
	specTest2Addr      = sdk.AccAddress(specTest2PubHex)
	specTest2Bech32    = specTest2Addr.String() // cosmos1zgp4n2yvrtxkj5zl6rzcf6phqg0gfzuf3v08r4
)

type SpecificationTestSuite struct {
	suite.Suite
}

func TestSpecificationTestSuite(t *testing.T) {
	suite.Run(t, new(SpecificationTestSuite))
}

func (s *SpecificationTestSuite) SetupSuite() {
	s.T().Parallel()
}

func (s *SpecificationTestSuite) TestScopeSpecValidateBasic() {
	tests := []struct {
		name string
		spec *ScopeSpecification
		want string
	}{
		// SpecificationId tests.
		{
			"invalid scope specification id - wrong address type",
			NewScopeSpecification(
				MetadataAddress(specTestAddr),
				nil, []string{}, []PartyType{}, []MetadataAddress{},
			),
			"invalid scope specification id: invalid metadata address type: 133",
		},
		{
			"invalid scope specification id - identifier",
			NewScopeSpecification(
				ScopeMetadataAddress(uuid.New()),
				nil, []string{}, []PartyType{}, []MetadataAddress{},
			),
			"invalid scope specification id prefix (expected: scopespec, got scope)",
		},
		// Description test to make sure Description.ValidateBasic is being used.
		{
			"invalid description name - too long",
			NewScopeSpecification(
				ScopeSpecMetadataAddress(uuid.New()),
				NewDescription(strings.Repeat("x", maxDescriptionNameLength+1), "", "", ""),
				[]string{}, []PartyType{}, []MetadataAddress{},
			),
			fmt.Sprintf("description (ScopeSpecification.Description) Name exceeds maximum length (expected <= %d got: %d)", maxDescriptionNameLength, maxDescriptionNameLength+1),
		},
		// OwnerAddresses tests
		{
			"owner addresses - cannot be empty",
			NewScopeSpecification(
				ScopeSpecMetadataAddress(uuid.New()),
				nil,
				[]string{},
				[]PartyType{}, []MetadataAddress{},
			),
			"the ScopeSpecification must have at least one owner",
		},
		{
			"owner addresses - invalid address at index 0",
			NewScopeSpecification(
				ScopeSpecMetadataAddress(uuid.New()),
				nil,
				[]string{":invalid", specTestBech32},
				[]PartyType{}, []MetadataAddress{},
			),
			"invalid owner address at index 0 on ScopeSpecification: decoding bech32 failed: invalid separator index -1",
		},
		{
			"owner addresses - invalid address at index 3",
			NewScopeSpecification(
				ScopeSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32, specTestBech32, specTestBech32, ":invalid"},
				[]PartyType{}, []MetadataAddress{},
			),
			"invalid owner address at index 3 on ScopeSpecification: decoding bech32 failed: invalid separator index -1",
		},
		// parties involved - cannot be empty
		{
			"parties involved - cannot be empty",
			NewScopeSpecification(
				ScopeSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{},
				[]MetadataAddress{},
			),
			"the ScopeSpecification must have at least one party involved",
		},
		// contract spec ids - must all pass same tests as scope spec id (contractspec prefix)
		{
			"contract spec ids - wrong address type at index 0",
			NewScopeSpecification(
				ScopeSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				[]MetadataAddress{MetadataAddress(specTestAddr)},
			),
			"invalid contract specification id at index 0: invalid metadata address type: 133",
		},
		{
			"contract spec ids - wrong prefix at index 0",
			NewScopeSpecification(
				ScopeSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				[]MetadataAddress{ScopeMetadataAddress(uuid.New())},
			),
			"invalid contract specification id prefix at index 0 (expected: contractspec, got scope)",
		},
		{
			"contract spec ids - wrong address type at index 2",
			NewScopeSpecification(
				ScopeSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				[]MetadataAddress{ContractSpecMetadataAddress(uuid.New()), ContractSpecMetadataAddress(uuid.New()), MetadataAddress(specTestAddr)},
			),
			"invalid contract specification id at index 2: invalid metadata address type: 133",
		},
		{
			"contract spec ids - wrong prefix at index 2",
			NewScopeSpecification(
				ScopeSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				[]MetadataAddress{ContractSpecMetadataAddress(uuid.New()), ContractSpecMetadataAddress(uuid.New()), ScopeMetadataAddress(uuid.New())},
			),
			"invalid contract specification id prefix at index 2 (expected: contractspec, got scope)",
		},
		// Simple valid case
		{
			"simple valid case",
			NewScopeSpecification(
				ScopeSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				[]MetadataAddress{ContractSpecMetadataAddress(uuid.New())},
			),
			"",
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.spec.ValidateBasic()
			if err != nil {
				require.Equal(t, tt.want, err.Error(), "ScopeSpecification ValidateBasic error")
			} else if len(tt.want) > 0 {
				t.Errorf("ScopeSpecification ValidateBasic error = nil, expected: %s", tt.want)
			}
		})
	}
}

func sovSpecTests(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func encodeVarintSpecTests(dAtA []byte, offset int, v uint64) int {
	offset -= sovSpecTests(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}

// WeirdSource is a thing that satisfies all needed pieces of the
// isContractSpecification_Source and isRecordSpecification_Source interfaces.
// But it isn't a valid thing to use as a Source according to the
// ContractSpecification.ValidateBasic() and RecordSpecification.ValidateBasic() methods.
type WeirdSource struct {
	Value uint32
}

func NewWeirdSource(value uint32) *WeirdSource {
	return &WeirdSource{
		Value: value,
	}
}
func (*WeirdSource) isContractSpecification_Source() {}
func (*WeirdSource) isInputSpecification_Source()    {}
func (m *WeirdSource) Size() (n int) {
	return 1 + sovSpecTests(uint64(n))
}
func (m *WeirdSource) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}
func (m *WeirdSource) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.Value != 0 {
		i = encodeVarintSpecTests(dAtA, i, uint64(m.Value))
		i--
		dAtA[i] = 0x2a
	}
	return len(dAtA) - i, nil
}

func (s *SpecificationTestSuite) TestContractSpecValidateBasic() {
	contractSpecUuid1 := uuid.New()
	tests := []struct {
		name string
		spec *ContractSpecification
		want string
	}{
		// SpecificationID tests
		{
			"SpecificationID - invalid format",
			NewContractSpecification(
				MetadataAddress(specTestAddr),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				NewContractSpecificationSourceHash("somehash"),
				"someclass",
			),
			"invalid contract specification id: invalid metadata address type: 133",
		},
		{
			"SpecificationID - invalid prefix",
			NewContractSpecification(
				ScopeSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				NewContractSpecificationSourceHash("somehash"),
				"someclass",
			),
			"invalid contract specification id prefix (expected: contractspec, got scopespec)",
		},

		// description - just make sure one of the Description.ValidateBasic pieces fails.
		{
			"Description - name too long",
			NewContractSpecification(
				ContractSpecMetadataAddress(uuid.New()),
				NewDescription(strings.Repeat("x", maxDescriptionNameLength+1), "", "", ""),
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				NewContractSpecificationSourceHash("somehash"),
				"someclass",
			),
			fmt.Sprintf("description (ContractSpecification.Description) Name exceeds maximum length (expected <= %d got: %d)", maxDescriptionNameLength, maxDescriptionNameLength+1),
		},

		// OwnerAddresses tests
		{
			"OwnerAddresses - empty",
			NewContractSpecification(
				ContractSpecMetadataAddress(uuid.New()),
				nil,
				[]string{},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				NewContractSpecificationSourceHash("somehash"),
				"someclass",
			),
			"invalid owner addresses count (expected > 0 got: 0)",
		},
		{
			"OwnerAddresses - invalid address at index 0",
			NewContractSpecification(
				ContractSpecMetadataAddress(uuid.New()),
				nil,
				[]string{":invalid"},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				NewContractSpecificationSourceHash("somehash"),
				"someclass",
			),
			fmt.Sprintf("invalid owner address at index %d: %s",
				0, "decoding bech32 failed: invalid separator index -1"),
		},
		{
			"OwnerAddresses - invalid address at index 2",
			NewContractSpecification(
				ContractSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32, specTestBech32, ":invalid"},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				NewContractSpecificationSourceHash("somehash"),
				"someclass",
			),
			fmt.Sprintf("invalid owner address at index %d: %s",
				2, "decoding bech32 failed: invalid separator index -1"),
		},

		// PartiesInvolved tests
		{
			"PartiesInvolved - empty",
			NewContractSpecification(
				ContractSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{},
				NewContractSpecificationSourceHash("somehash"),
				"someclass",
			),
			"invalid parties involved count (expected > 0 got: 0)",
		},

		// Source tests
		{
			"Source - nil",
			NewContractSpecification(
				ContractSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				nil,
				"someclass",
			),
			"a source is required",
		},
		{
			"Source - ResourceID - invalid",
			NewContractSpecification(
				ContractSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				NewContractSpecificationSourceResourceID(MetadataAddress(specTestAddr)),
				"someclass",
			),
			fmt.Sprintf("invalid source resource id: %s", "invalid metadata address type: 133"),
		},
		{
			"Source - Hash - empty",
			NewContractSpecification(
				ContractSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				NewContractSpecificationSourceHash(""),
				"someclass",
			),
			"source hash cannot be empty",
		},
		{
			"Source - unknown type",
			NewContractSpecification(
				ContractSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				NewWeirdSource(3),
				"someclass",
			),
			"unknown source type",
		},

		// ClassName tests
		{
			"ClassName - empty",
			NewContractSpecification(
				ContractSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				NewContractSpecificationSourceHash("somehash"),
				"",
			),
			"class name cannot be empty",
		},
		{
			"ClassName - just over max length",
			NewContractSpecification(
				ContractSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				NewContractSpecificationSourceHash("somehash"),
				strings.Repeat("l", maxContractSpecificationClassNameLength+1),
			),
			fmt.Sprintf("class name exceeds maximum length (expected <= %d got: %d)",
				maxContractSpecificationClassNameLength, maxContractSpecificationClassNameLength+1),
		},
		{
			"ClassName - at max length",
			NewContractSpecification(
				ContractSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				NewContractSpecificationSourceHash("somehash"),
				strings.Repeat("m", maxContractSpecificationClassNameLength),
			),
			"",
		},

		// A simple valid ContractSpecification
		{
			"simple valid test case",
			NewContractSpecification(
				ContractSpecMetadataAddress(contractSpecUuid1),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				NewContractSpecificationSourceHash("somehash"),
				"someclass",
			),
			"",
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.spec.ValidateBasic()
			if err != nil {
				require.Equal(t, tt.want, err.Error(), "ContractSpecification ValidateBasic error")
			} else if len(tt.want) > 0 {
				t.Errorf("ContractSpecification ValidateBasic error = nil, expected: %s", tt.want)
			}
		})
	}
}

func (s *SpecificationTestSuite) TestRecordSpecValidateBasic() {
	contractSpecUUID := uuid.New()
	tests := []struct {
		name string
		spec *RecordSpecification
		want string
	}{
		// SpecificationId tests
		{
			"SpecificationId - invalid format",
			&RecordSpecification{
				SpecificationId:    MetadataAddress(specTestAddr),
				Name:               "recspecname",
				Inputs:             []*InputSpecification{},
				TypeName:           "recspectypename",
				ResultType:         DefinitionType_DEFINITION_TYPE_RECORD,
				ResponsibleParties: []PartyType{PartyType_PARTY_TYPE_OWNER},
			},
			"invalid record specification id: invalid metadata address type: 133",
		},
		{
			"SpecificationId - invalid prefix (record)",
			&RecordSpecification{
				SpecificationId:    RecordMetadataAddress(contractSpecUUID, "recspecname"),
				Name:               "recspecname",
				Inputs:             []*InputSpecification{},
				TypeName:           "recspectypename",
				ResultType:         DefinitionType_DEFINITION_TYPE_RECORD,
				ResponsibleParties: []PartyType{PartyType_PARTY_TYPE_OWNER},
			},
			fmt.Sprintf("invalid record specification id prefix (expected: %s, got %s)",
				PrefixRecordSpecification, PrefixRecord),
		},
		{
			"SpecificationId - invalid prefix (contract spec)",
			&RecordSpecification{
				SpecificationId:    ContractSpecMetadataAddress(contractSpecUUID),
				Name:               "recspecname",
				Inputs:             []*InputSpecification{},
				TypeName:           "recspectypename",
				ResultType:         DefinitionType_DEFINITION_TYPE_RECORD,
				ResponsibleParties: []PartyType{PartyType_PARTY_TYPE_OWNER},
			},
			fmt.Sprintf("invalid record specification id prefix (expected: %s, got %s)",
				PrefixRecordSpecification, PrefixContractSpecification),
		},
		{
			"SpecificationId - incorrect name hash",
			&RecordSpecification{
				SpecificationId:    RecordSpecMetadataAddress(contractSpecUUID, "recspecothername"),
				Name:               "recspecname",
				Inputs:             []*InputSpecification{},
				TypeName:           "recspectypename",
				ResultType:         DefinitionType_DEFINITION_TYPE_RECORD,
				ResponsibleParties: []PartyType{PartyType_PARTY_TYPE_OWNER},
			},
			fmt.Sprintf("invalid record specification id value (expected: %s, got %s)",
				RecordSpecMetadataAddress(contractSpecUUID, "recspecname"),
				RecordSpecMetadataAddress(contractSpecUUID, "recspecothername")),
		},

		// Name tests
		{
			"Name - empty",
			&RecordSpecification{
				SpecificationId:    RecordSpecMetadataAddress(contractSpecUUID, "recspecname"),
				Name:               "",
				Inputs:             []*InputSpecification{},
				TypeName:           "recspectypename",
				ResultType:         DefinitionType_DEFINITION_TYPE_RECORD,
				ResponsibleParties: []PartyType{PartyType_PARTY_TYPE_OWNER},
			},
			"record specification name cannot be empty",
		},
		{
			"Name - too long",
			&RecordSpecification{
				SpecificationId:    RecordSpecMetadataAddress(contractSpecUUID, "recspecname"),
				Name:               strings.Repeat("r", maxRecordSpecificationNameLength+1),
				Inputs:             []*InputSpecification{},
				TypeName:           "recspectypename",
				ResultType:         DefinitionType_DEFINITION_TYPE_RECORD,
				ResponsibleParties: []PartyType{PartyType_PARTY_TYPE_OWNER},
			},
			fmt.Sprintf("record specification name exceeds maximum length (expected <= %d got: %d)",
				maxRecordSpecificationNameLength, maxRecordSpecificationNameLength+1),
		},
		{
			"Name - max length - okay",
			&RecordSpecification{
				SpecificationId:    RecordSpecMetadataAddress(contractSpecUUID, strings.Repeat("r", maxRecordSpecificationNameLength)),
				Name:               strings.Repeat("r", maxRecordSpecificationNameLength),
				Inputs:             []*InputSpecification{},
				TypeName:           "recspectypename",
				ResultType:         DefinitionType_DEFINITION_TYPE_RECORD,
				ResponsibleParties: []PartyType{PartyType_PARTY_TYPE_OWNER},
			},
			"",
		},

		// Inputs tests
		{
			"Inputs - invalid name at index 0",
			&RecordSpecification{
				SpecificationId: RecordSpecMetadataAddress(contractSpecUUID, "recspecname"),
				Name:            "recspecname",
				Inputs: []*InputSpecification{
					{
						Name:     "",
						TypeName: "typename1",
						Source:   NewInputSpecificationSourceHash("inputspecsourcehash1"),
					},
					{
						Name:     "name2",
						TypeName: "typename2",
						Source:   NewInputSpecificationSourceHash("inputspecsourcehash2"),
					},
					{
						Name:     "name3",
						TypeName: "typename3",
						Source:   NewInputSpecificationSourceHash("inputspecsourcehash3"),
					},
				},
				TypeName:           "recspectypename",
				ResultType:         DefinitionType_DEFINITION_TYPE_RECORD,
				ResponsibleParties: []PartyType{PartyType_PARTY_TYPE_OWNER},
			},
			fmt.Sprintf("invalid input specification at index %d: %s",
				0, "input specification name cannot be empty"),
		},
		{
			"Inputs - invalid name at index 2",
			&RecordSpecification{
				SpecificationId: RecordSpecMetadataAddress(contractSpecUUID, "recspecname"),
				Name:            "recspecname",
				Inputs: []*InputSpecification{
					{
						Name:     "name1",
						TypeName: "typename1",
						Source:   NewInputSpecificationSourceHash("inputspecsourcehash1"),
					},
					{
						Name:     "name2",
						TypeName: "typename2",
						Source:   NewInputSpecificationSourceHash("inputspecsourcehash2"),
					},
					{
						Name:     "",
						TypeName: "typename3",
						Source:   NewInputSpecificationSourceHash("inputspecsourcehash3"),
					},
				},
				TypeName:           "recspectypename",
				ResultType:         DefinitionType_DEFINITION_TYPE_RECORD,
				ResponsibleParties: []PartyType{PartyType_PARTY_TYPE_OWNER},
			},
			fmt.Sprintf("invalid input specification at index %d: %s",
				2, "input specification name cannot be empty"),
		},

		// TypeName tests
		{
			"TypeName - empty",
			&RecordSpecification{
				SpecificationId:    RecordSpecMetadataAddress(contractSpecUUID, "recspecname"),
				Name:               "recspecname",
				Inputs:             []*InputSpecification{},
				TypeName:           "",
				ResultType:         DefinitionType_DEFINITION_TYPE_RECORD,
				ResponsibleParties: []PartyType{PartyType_PARTY_TYPE_OWNER},
			},
			"record specification type name cannot be empty",
		},
		{
			"TypeName - too long",
			&RecordSpecification{
				SpecificationId:    RecordSpecMetadataAddress(contractSpecUUID, "recspecname"),
				Name:               "recspecname",
				Inputs:             []*InputSpecification{},
				TypeName:           strings.Repeat("t", maxRecordSpecificationTypeNameLength+1),
				ResultType:         DefinitionType_DEFINITION_TYPE_RECORD,
				ResponsibleParties: []PartyType{PartyType_PARTY_TYPE_OWNER},
			},
			fmt.Sprintf("record specification type name exceeds maximum length (expected <= %d got: %d)",
				maxRecordSpecificationTypeNameLength, maxRecordSpecificationTypeNameLength+1),
		},
		{
			"TypeName - max length - okay",
			&RecordSpecification{
				SpecificationId:    RecordSpecMetadataAddress(contractSpecUUID, "recspecname"),
				Name:               "recspecname",
				Inputs:             []*InputSpecification{},
				TypeName:           strings.Repeat("t", maxRecordSpecificationTypeNameLength),
				ResultType:         DefinitionType_DEFINITION_TYPE_RECORD,
				ResponsibleParties: []PartyType{PartyType_PARTY_TYPE_OWNER},
			},
			"",
		},

		// ResponsibleParties tests
		{
			"ResponsibleParties - empty",
			&RecordSpecification{
				SpecificationId:    RecordSpecMetadataAddress(contractSpecUUID, "recspecname"),
				Name:               "recspecname",
				Inputs:             []*InputSpecification{},
				TypeName:           "recspectypename",
				ResultType:         DefinitionType_DEFINITION_TYPE_RECORD,
				ResponsibleParties: []PartyType{},
			},
			"invalid responsible parties count (expected > 0 got: 0)",
		},

		// A simple valid RecordSpecification
		{
			"simple valid RecordSpecification",
			&RecordSpecification{
				SpecificationId:    RecordSpecMetadataAddress(contractSpecUUID, "recspecname"),
				Name:               "recspecname",
				Inputs:             []*InputSpecification{},
				TypeName:           "recspectypename",
				ResultType:         DefinitionType_DEFINITION_TYPE_RECORD,
				ResponsibleParties: []PartyType{PartyType_PARTY_TYPE_OWNER},
			},
			"",
		},

		// ResultType test
		{
			"result type cannot be unspecified",
			&RecordSpecification{
				SpecificationId:    RecordSpecMetadataAddress(contractSpecUUID, "recspecname"),
				Name:               "recspecname",
				Inputs:             []*InputSpecification{},
				TypeName:           "recspectypename",
				ResultType:         DefinitionType_DEFINITION_TYPE_UNSPECIFIED,
				ResponsibleParties: []PartyType{PartyType_PARTY_TYPE_OWNER},
			},
			"record specification result type cannot be unspecified",
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.spec.ValidateBasic()
			if err != nil {
				require.Equal(t, tt.want, err.Error(), "RecordSpecification ValidateBasic error")
			} else if len(tt.want) > 0 {
				t.Errorf("RecordSpecification ValidateBasic error = nil, expected: %s", tt.want)
			}
		})
	}
}

func (s *SpecificationTestSuite) TestInputSpecValidateBasic() {
	tests := []struct {
		name string
		spec *InputSpecification
		want string
	}{
		// Name tests
		{
			"Name - empty",
			&InputSpecification{
				Name:     "",
				TypeName: "typename",
				Source:   NewInputSpecificationSourceHash("inputspecsourcehash"),
			},
			"input specification name cannot be empty",
		},
		{
			"Name - too long",
			&InputSpecification{
				Name:     strings.Repeat("i", maxInputSpecificationNameLength+1),
				TypeName: "typename",
				Source:   NewInputSpecificationSourceHash("inputspecsourcehash"),
			},
			fmt.Sprintf("input specification name exceeds maximum length (expected <= %d got: %d)",
				maxInputSpecificationNameLength, maxInputSpecificationNameLength+1),
		},
		{
			"Name - at max length - okay",
			&InputSpecification{
				Name:     strings.Repeat("i", maxInputSpecificationNameLength),
				TypeName: "typename",
				Source:   NewInputSpecificationSourceHash("inputspecsourcehash"),
			},
			"",
		},

		// TypeName tests
		{
			"TypeName - empty",
			&InputSpecification{
				Name:     "name",
				TypeName: "",
				Source:   NewInputSpecificationSourceHash("inputspecsourcehash"),
			},
			"input specification type name cannot be empty",
		},
		{
			"TypeName - too long",
			&InputSpecification{
				Name:     "name",
				TypeName: strings.Repeat("i", maxInputSpecificationTypeNameLength+1),
				Source:   NewInputSpecificationSourceHash("inputspecsourcehash"),
			},
			fmt.Sprintf("input specification type name exceeds maximum length (expected <= %d got: %d)",
				maxInputSpecificationTypeNameLength, maxInputSpecificationTypeNameLength+1),
		},
		{
			"TypeName - at max length - okay",
			&InputSpecification{
				Name:     "name",
				TypeName: strings.Repeat("i", maxInputSpecificationTypeNameLength),
				Source:   NewInputSpecificationSourceHash("inputspecsourcehash"),
			},
			"",
		},

		// Source tests
		{
			"Source - nil",
			&InputSpecification{
				Name:     "name",
				TypeName: "typename",
				Source:   nil,
			},
			"input specification source is required",
		},
		{
			"Source - RecordId - not a metadata address",
			&InputSpecification{
				Name:     "name",
				TypeName: "typename",
				Source:   NewInputSpecificationSourceRecordID(MetadataAddress(specTestAddr)),
			},
			"invalid input specification source record id: invalid metadata address type: 133",
		},
		{
			"Source - RecordId - wrong prefix",
			&InputSpecification{
				Name:     "name",
				TypeName: "typename",
				Source:   NewInputSpecificationSourceRecordID(ContractSpecMetadataAddress(uuid.New())),
			},
			fmt.Sprintf("invalid input specification source record id prefix (expected: %s, got: %s)",
				PrefixRecord, PrefixContractSpecification),
		},
		{
			"Source - Hash - empty",
			&InputSpecification{
				Name:     "name",
				TypeName: "typename",
				Source:   NewInputSpecificationSourceHash(""),
			},
			"input specification source hash cannot be empty",
		},
		{
			"Source - Weird",
			&InputSpecification{
				Name:     "name",
				TypeName: "typename",
				Source:   NewWeirdSource(8),
			},
			"unknown input specification source type",
		},

		// A simple valid InputSpecification
		{
			"simple valid InputSpecification",
			&InputSpecification{
				Name:     "name",
				TypeName: "typename",
				Source:   NewInputSpecificationSourceHash("inputspecsourcehash"),
			},
			"",
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.spec.ValidateBasic()
			if err != nil {
				require.Equal(t, tt.want, err.Error(), "InputSpecification ValidateBasic error")
			} else if len(tt.want) > 0 {
				t.Errorf("InputSpecification ValidateBasic error = nil, expected: %s", tt.want)
			}
		})
	}
}

func (s *SpecificationTestSuite) TestDescriptionValidateBasic() {
	tests := []struct {
		name string
		desc *Description
		want string
	}{
		// Name tests
		{
			"invalid name - empty",
			NewDescription(
				"",
				"",
				"",
				"",
			),
			fmt.Sprintf("description Name cannot be empty"),
		},
		{
			"invalid name - too long",
			NewDescription(
				strings.Repeat("x", maxDescriptionNameLength+1),
				"",
				"",
				"",
			),
			fmt.Sprintf("description Name exceeds maximum length (expected <= %d got: %d)", maxDescriptionNameLength, maxDescriptionNameLength+1),
		},
		{
			"valid name - 1 char",
			NewDescription(
				"x",
				"",
				"",
				"",
			),
			"",
		},
		{
			"valid name - exactly max length",
			NewDescription(
				strings.Repeat("y", maxDescriptionNameLength),
				"",
				"",
				"",
			),
			"",
		},

		// Description tests
		{
			"invalid description - too long",
			NewDescription(
				"Unit Tests",
				strings.Repeat("z", maxDescriptionDescriptionLength+1),
				"",
				"",
			),
			fmt.Sprintf("description Description exceeds maximum length (expected <= %d got: %d)", maxDescriptionDescriptionLength, maxDescriptionDescriptionLength+1),
		},
		{
			"valid description - empty",
			NewDescription(
				"Unit Tests",
				"",
				"",
				"",
			),
			"",
		},
		{
			"valid description - 1 char",
			NewDescription(
				"Unit Tests",
				"z",
				"",
				"",
			),
			"",
		},
		{
			"valid description - exactly max length",
			NewDescription(
				"Unit Tests",
				strings.Repeat("z", maxDescriptionDescriptionLength),
				"",
				"",
			),
			"",
		},

		// Website url tests
		{
			"invalid website url - too long",
			NewDescription(
				"Unit Tests",
				"",
				strings.Repeat("h", maxURLLength+1),
				"",
			),
			fmt.Sprintf("url WebsiteUrl exceeds maximum length (expected <= %d got: %d)", maxURLLength, maxURLLength+1),
		},
		{
			"invalid website url - no protocol",
			NewDescription(
				"Unit Tests",
				"",
				"www.test.com",
				"",
			),
			fmt.Sprintf("url WebsiteUrl must use the http, https, or data protocol"),
		},
		{
			"valid website url - http",
			NewDescription(
				"Unit Tests",
				"",
				"http://www.test.com",
				"",
			),
			"",
		},
		{
			"valid website url - http at max length",
			NewDescription(
				"Unit Tests",
				"",
				"http://"+strings.Repeat("f", maxURLLength-7),
				"",
			),
			"",
		},
		{
			"valid website url - https",
			NewDescription(
				"Unit Tests",
				"",
				"https://www.test.com",
				"",
			),
			"",
		},
		{
			"valid website url - https at max length",
			NewDescription(
				"Unit Tests",
				"",
				"https://"+strings.Repeat("s", maxURLLength-8),
				"",
			),
			"",
		},
		{
			"valid website url - data",
			NewDescription(
				"Unit Tests",
				"",
				"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAYAAACNbyblAAAAHElEQVQI12P4//8/w38GIAXDIBKE0DHxgljNBAAO9TXL0Y4OHwAAAABJRU5ErkJggg==",
				"",
			),
			"",
		},
		{
			"valid website url - data minimal",
			NewDescription(
				"Unit Tests",
				"",
				"data:,",
				"",
			),
			"",
		},
		{
			"valid website url - data at max length",
			NewDescription(
				"Unit Tests",
				"",
				"data:image/png;base64,"+strings.Repeat("d", maxURLLength-22),
				"",
			),
			"",
		},

		// Icon url tests
		{
			"invalid icon url - too long",
			NewDescription(
				"Unit Tests",
				"",
				"",
				strings.Repeat("h", maxURLLength+1),
			),
			fmt.Sprintf("url IconUrl exceeds maximum length (expected <= %d got: %d)", maxURLLength, maxURLLength+1),
		},
		{
			"invalid icon url - no protocol",
			NewDescription(
				"Unit Tests",
				"",
				"",
				"www.test.com",
			),
			fmt.Sprintf("url IconUrl must use the http, https, or data protocol"),
		},
		{
			"valid icon url - http",
			NewDescription(
				"Unit Tests",
				"",
				"",
				"http://www.test.com",
			),
			"",
		},
		{
			"valid icon url - http at max length",
			NewDescription(
				"Unit Tests",
				"",
				"",
				"http://"+strings.Repeat("f", maxURLLength-7),
			),
			"",
		},
		{
			"valid icon url - https",
			NewDescription(
				"Unit Tests",
				"",
				"",
				"https://www.test.com",
			),
			"",
		},
		{
			"valid icon url - https at max length",
			NewDescription(
				"Unit Tests",
				"",
				"",
				"https://"+strings.Repeat("s", maxURLLength-8),
			),
			"",
		},
		{
			"valid website url - data",
			NewDescription(
				"Unit Tests",
				"",
				"",
				"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAYAAACNbyblAAAAHElEQVQI12P4//8/w38GIAXDIBKE0DHxgljNBAAO9TXL0Y4OHwAAAABJRU5ErkJggg==",
			),
			"",
		},
		{
			"valid website url - data minimal",
			NewDescription(
				"Unit Tests",
				"",
				"",
				"data:,",
			),
			"",
		},
		{
			"valid website url - data at max length",
			NewDescription(
				"Unit Tests",
				"",
				"",
				"data:image/png;base64,"+strings.Repeat("d", maxURLLength-22),
			),
			"",
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.desc.ValidateBasic("")
			if err != nil {
				require.Equal(t, tt.want, err.Error(), "Description ValidateBasic error")
			} else if len(tt.want) > 0 {
				t.Errorf("Description ValidateBasic error = nil, expected: %s", tt.want)
			}
		})
	}
}

func (s *SpecificationTestSuite) TestScopeSpecString() {
	scopeSpecUuid := uuid.MustParse("c2074a03-6f6d-4029-bfe2-c3a5eb7e68b1")
	contractSpecUuid := uuid.MustParse("540dadf1-3dbc-4c3f-a205-7575b7f74384")
	scopeSpec := NewScopeSpecification(
		ScopeSpecMetadataAddress(scopeSpecUuid),
		NewDescription(
			"TestScopeSpecString Description",
			"This is a description of a description used in a unit test.",
			"https://figure.com/",
			"https://figure.com/favicon.png",
		),
		[]string{specTestBech32},
		[]PartyType{PartyType_PARTY_TYPE_OWNER},
		[]MetadataAddress{ContractSpecMetadataAddress(contractSpecUuid)},
	)
	expected := `specification_id: scopespec1qnpqwjsrdak5q2dlutp6t6m7dzcscd7ff6
description:
  name: TestScopeSpecString Description
  description: This is a description of a description used in a unit test.
  website_url: https://figure.com/
  icon_url: https://figure.com/favicon.png
owner_addresses:
- cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck
parties_involved:
- 5
contract_spec_ids:
- contractspec1qd2qmt038k7yc0azq46htdlhgwzqg6cr9l
`
	actual := scopeSpec.String()
	// fmt.Printf("Actual:\n%s\n-----\n", actual)
	require.Equal(s.T(), expected, actual)
}

func (s *SpecificationTestSuite) TestContractSpecString() {
	contractSpecUuid := uuid.MustParse("540dadf1-3dbc-4c3f-a205-7575b7f74384")
	contractSpec := NewContractSpecification(
		ContractSpecMetadataAddress(contractSpecUuid),
		NewDescription(
			"TestContractSpecString Description",
			"This is a description of a description used in a unit test.",
			"https://figure.com/",
			"https://figure.com/favicon.png",
		),
		[]string{specTestBech32},
		[]PartyType{PartyType_PARTY_TYPE_OWNER},
		nil,
		"CS 201: Intro to Blockchain",
	)
	expected := `specification_id: contractspec1qd2qmt038k7yc0azq46htdlhgwzqg6cr9l
description:
  name: TestContractSpecString Description
  description: This is a description of a description used in a unit test.
  website_url: https://figure.com/
  icon_url: https://figure.com/favicon.png
owner_addresses:
- cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck
parties_involved:
- 5
source: null
class_name: 'CS 201: Intro to Blockchain'
`
	actual := contractSpec.String()
	// fmt.Printf("Actual:\n%s\n-----\n", actual)
	require.Equal(s.T(), expected, actual)
}

func (s *SpecificationTestSuite) TestRecordSpecString() {
	contractSpecUuid := uuid.MustParse("540dadf1-3dbc-4c3f-a205-7575b7f74384")
	recordName := "somename"
	recordSpec := NewRecordSpecification(
		RecordSpecMetadataAddress(contractSpecUuid, recordName),
		recordName,
		[]*InputSpecification{
			{
				"inputSpecName1",
				"inputSpecTypeName1",
				NewInputSpecificationSourceHash("inputSpecSourceHash1"),
			},
			{
				"inputSpecName2",
				"inputSpecTypeName2",
				NewInputSpecificationSourceRecordID(RecordMetadataAddress(
					uuid.MustParse("1784AE79-77F1-421C-AAF9-ECA4DD79E571"),
					"inputSpecRecordIdSource",
				)),
			},
		},
		"sometype",
		DefinitionType_DEFINITION_TYPE_RECORD,
		[]PartyType{PartyType_PARTY_TYPE_CUSTODIAN, PartyType_PARTY_TYPE_INVESTOR},
	)
	expected := `specification_id: recspec1q42qmt038k7yc0azq46htdlhgwzg5052mucgmerfku3gf5e7t3ej5fjh7rr
name: somename
inputs:
- name: inputSpecName1
  type_name: inputSpecTypeName1
  source:
    hash: inputSpecSourceHash1
- name: inputSpecName2
  type_name: inputSpecTypeName2
  source:
    record_id: record1qgtcftnewlc5y892l8k2fhteu4ceth857yw3fprr4lvhfptn5gg4cv4ure3
type_name: sometype
result_type: 2
responsible_parties:
- 4
- 3
`
	actual := recordSpec.String()
	// fmt.Printf("Actual:\n%s\n-----\n", actual)
	require.Equal(s.T(), expected, actual)
}

func (s *SpecificationTestSuite) TestInputSpecString() {
	tests := []struct {
		name         string
		outputActual bool
		spec         *InputSpecification
		expected     string
	}{
		{
			"source is record id",
			false,
			NewInputSpecification(
				"inputSpecRecordIdSource",
				"inputSpecRecordIdSourceTypeName",
				NewInputSpecificationSourceRecordID(RecordMetadataAddress(
					uuid.MustParse("1784AE79-77F1-421C-AAF9-ECA4DD79E571"),
					"inputSpecRecordIdSource",
				)),
			),
			`name: inputSpecRecordIdSource
type_name: inputSpecRecordIdSourceTypeName
source:
  record_id: record1qgtcftnewlc5y892l8k2fhteu4ceth857yw3fprr4lvhfptn5gg4cv4ure3
`,
		},
		{
			"source is hash",
			false,
			NewInputSpecification(
				"inputSpecHashSource",
				"inputSpecHashSourceTypeName",
				NewInputSpecificationSourceHash("somehash"),
			),
			`name: inputSpecHashSource
type_name: inputSpecHashSourceTypeName
source:
  hash: somehash
`,
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			actual := tt.spec.String()
			if tt.outputActual {
				fmt.Printf("Actual [%s]:\n%s\n-----\n", tt.name, actual)
			}
			require.Equal(t, tt.expected, actual)
		})
	}
}

func (s *SpecificationTestSuite) TestDescriptionString() {
	description := NewDescription(
		"TestDescriptionString",
		"This is a description of a description used in a unit test.",
		"https://provenance.io",
		"https://provenance.io/ico.png",
	)
	expected := `name: TestDescriptionString
description: This is a description of a description used in a unit test.
website_url: https://provenance.io
icon_url: https://provenance.io/ico.png
`
	actual := description.String()
	// fmt.Printf("Actual:\n%s\n-----\n", actual)
	require.Equal(s.T(), expected, actual)
}
