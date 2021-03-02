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
)

type specificationTestSuite struct {
	suite.Suite
}

func TestSpecificationTestSuite(t *testing.T) {
	suite.Run(t, new(specificationTestSuite))
}

func (s *specificationTestSuite) SetupSuite() {
	s.T().Parallel()
}

func (s *specificationTestSuite) TestScopeSpecValidateBasic() {
	tests := []struct {
		name     string
		spec     *ScopeSpecification
		want     string
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
				NewDescription(strings.Repeat("x", maxDescriptionNameLength + 1), "", "", ""),
				[]string{}, []PartyType{}, []MetadataAddress{},
			),
			fmt.Sprintf("description (ScopeSpecification.Description) Name exceeds maximum length (expected <= %d got: %d)", maxDescriptionNameLength, maxDescriptionNameLength + 1),
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
			"invalid owner address at index 0 on ScopeSpecification: decoding bech32 failed: invalid index of 1",
		},
		{
			"owner addresses - invalid address at index 3",
			NewScopeSpecification(
				ScopeSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32, specTestBech32, specTestBech32, ":invalid"},
				[]PartyType{}, []MetadataAddress{},
			),
			"invalid owner address at index 3 on ScopeSpecification: decoding bech32 failed: invalid index of 1",
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
				require.Equal(t, tt.want, err.Error(), "ScopeSpec ValidateBasic error")
			} else if len(tt.want) > 0 {
				t.Errorf("ScopeSpec ValidateBasic error = nil, expected: %s", tt.want)
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

// WeirdSoure is a thing that satisfies all needed pieces of the isContractSpecification_Source interface
// but isn't a valid thing to use as a Source according to the ContractSpecification.ValidateBasic() method.
type WeirdSource struct {
	Value uint32
}
func NewWeirdSource(value uint32) *WeirdSource {
	return &WeirdSource{
		Value: value,
	}
}
func (*WeirdSource) isContractSpecification_Source() {}
func (*WeirdSource) isRecordSpecification_Source() {}
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

func (s *specificationTestSuite) TestContractSpecValidateBasic() {
	contractSpecUuid1 := uuid.New()
	contractSpecUuid2 := uuid.New()
	tests := []struct {
		name     string
		spec     *ContractSpecification
		want     string
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
				[]MetadataAddress{},
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
				[]MetadataAddress{},
			),
			"invalid contract specification id prefix (expected: contractspec, got scopespec)",
		},

		// description - just make sure one of the Description.ValidateBasic pieces fails.
		{
			"Description - name too long",
			NewContractSpecification(
				ContractSpecMetadataAddress(uuid.New()),
				NewDescription(strings.Repeat("x", maxDescriptionNameLength + 1), "", "", ""),
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]MetadataAddress{},
			),
			fmt.Sprintf("description (ContractSpecification.Description) Name exceeds maximum length (expected <= %d got: %d)", maxDescriptionNameLength, maxDescriptionNameLength + 1),
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
				[]MetadataAddress{},
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
				[]MetadataAddress{},
			),
			fmt.Sprintf("invalid owner address at index %d: %s",
				0, "decoding bech32 failed: invalid index of 1"),
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
				[]MetadataAddress{},
			),
			fmt.Sprintf("invalid owner address at index %d: %s",
				2, "decoding bech32 failed: invalid index of 1"),
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
				[]MetadataAddress{},
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
				[]MetadataAddress{},
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
				[]MetadataAddress{},
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
				[]MetadataAddress{},
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
				[]MetadataAddress{},
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
				[]MetadataAddress{},
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
				strings.Repeat("l", maxContractSpecificationClassNameLength + 1),
				[]MetadataAddress{},
			),
			fmt.Sprintf("class name exceeds maximum length (expected <= %d got: %d)",
				maxContractSpecificationClassNameLength, maxContractSpecificationClassNameLength + 1),
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
				[]MetadataAddress{},
			),
			"",
		},

		// RecordSpecIDs tests
		{
			"RecordSpecIDs - invalid address at index 0",
			NewContractSpecification(
				ContractSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]MetadataAddress{MetadataAddress(specTestAddr)},
			),
			fmt.Sprintf("invalid record specification id at index %d: %s",
				0, "invalid metadata address type: 133"),
		},
		{
			"RecordSpecIDs - invalid address prefix at index 0",
			NewContractSpecification(
				ContractSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]MetadataAddress{ContractSpecMetadataAddress(uuid.New())},
			),
			"invalid record specification id prefix at index 0 (expected: recspec, got contractspec)",
		},
		{
			"RecordSpecIDs - wrong contract spec uuid at index 0",
			NewContractSpecification(
				ContractSpecMetadataAddress(contractSpecUuid1),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]MetadataAddress{RecordSpecMetadataAddress(contractSpecUuid2, "cs2")},
			),
			fmt.Sprintf("invalid record specification id contract specification uuid value at index %d (expected :%s, got %s)",
				0, contractSpecUuid1.String(), contractSpecUuid2.String()),
		},
		{
			"RecordSpecIDs - invalid address at index 2",
			NewContractSpecification(
				ContractSpecMetadataAddress(contractSpecUuid1),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]MetadataAddress{
					RecordSpecMetadataAddress(contractSpecUuid1, "name1"),
					RecordSpecMetadataAddress(contractSpecUuid1, "name2"),
					MetadataAddress(specTestAddr),
				},
			),
			fmt.Sprintf("invalid record specification id at index %d: %s",
				2, "invalid metadata address type: 133"),
		},
		{
			"RecordSpecIDs - invalid address prefix at index 2",
			NewContractSpecification(
				ContractSpecMetadataAddress(contractSpecUuid1),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]MetadataAddress{
					RecordSpecMetadataAddress(contractSpecUuid1, "name1"),
					RecordSpecMetadataAddress(contractSpecUuid1, "name2"),
					ContractSpecMetadataAddress(contractSpecUuid1),
				},
			),
			"invalid record specification id prefix at index 2 (expected: recspec, got contractspec)",
		},
		{
			"RecordSpecIDs - wrong contract spec uuid at index 0",
			NewContractSpecification(
				ContractSpecMetadataAddress(contractSpecUuid1),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				NewContractSpecificationSourceHash("somehash"),
				"someclass",
				[]MetadataAddress{
					RecordSpecMetadataAddress(contractSpecUuid1, "name1"),
					RecordSpecMetadataAddress(contractSpecUuid1, "name2"),
					RecordSpecMetadataAddress(contractSpecUuid2, "name3"),
				},
			),
			fmt.Sprintf("invalid record specification id contract specification uuid value at index %d (expected :%s, got %s)",
				2, contractSpecUuid1.String(), contractSpecUuid2.String()),
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
				[]MetadataAddress{RecordSpecMetadataAddress(contractSpecUuid1,"recspecname")},
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

func (s *specificationTestSuite) TestDescriptionValidateBasic() {
	tests := []struct {
		name     string
		desc     *Description
		want     string
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
				strings.Repeat("x", maxDescriptionNameLength + 1),
				"",
				"",
				"",
			),
			fmt.Sprintf("description Name exceeds maximum length (expected <= %d got: %d)", maxDescriptionNameLength, maxDescriptionNameLength + 1),
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
				strings.Repeat("z", maxDescriptionDescriptionLength + 1),
				"",
				"",
			),
			fmt.Sprintf("description Description exceeds maximum length (expected <= %d got: %d)", maxDescriptionDescriptionLength, maxDescriptionDescriptionLength + 1),
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
				strings.Repeat("h", maxURLLength + 1),
				"",
			),
			fmt.Sprintf("url WebsiteUrl exceeds maximum length (expected <= %d got: %d)", maxURLLength, maxURLLength + 1),
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
				"http://" + strings.Repeat("f", maxURLLength - 7),
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
				"https://" + strings.Repeat("s", maxURLLength - 8),
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
				"data:image/png;base64," + strings.Repeat("d", maxURLLength - 22),
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
				strings.Repeat("h", maxURLLength + 1),
			),
			fmt.Sprintf("url IconUrl exceeds maximum length (expected <= %d got: %d)", maxURLLength, maxURLLength + 1),
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
				"http://" + strings.Repeat("f", maxURLLength - 7),
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
				"https://" + strings.Repeat("s", maxURLLength - 8),
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
				"data:image/png;base64," + strings.Repeat("d", maxURLLength - 22),
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

func (s *specificationTestSuite) TestScopeSpecString() {
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

func (s *specificationTestSuite) TestContractSpecString() {
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
		[]MetadataAddress{RecordSpecMetadataAddress(contractSpecUuid, "somename")},
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
record_spec_ids:
- recspec1q42qmt038k7yc0azq46htdlhgwzg5052mucgmerfku3gf5e7t3ej4ecag50yfxlsd8m2udlxzca6vhgch80wy
`
	actual := contractSpec.String()
	// fmt.Printf("Actual:\n%s\n-----\n", actual)
	require.Equal(s.T(), expected, actual)
}

func (s *specificationTestSuite) TestDescriptionString() {
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
