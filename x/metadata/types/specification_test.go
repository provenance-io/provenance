package types

import (
	"encoding/hex"
	"fmt"
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
		wantErr  bool
	}{
		// SpecificationId tests.
		{
			"invalid scope specification id - wrong address type",
			NewScopeSpecification(
				MetadataAddress(specTestAddr),
				nil, []string{}, []PartyType{}, []MetadataAddress{},
			),
			"invalid scope specification id: invalid metadata address type (must be 0-4, actual: 133)",
			true,
		},
		{
			"invalid scope specification id - identifier",
			NewScopeSpecification(
				ScopeMetadataAddress(uuid.New()),
				nil, []string{}, []PartyType{}, []MetadataAddress{},
			),
			"invalid scope specification id prefix (expected: scopespec, got scope)",
			true,
		},
		// Description test to make sure Description.ValidateBasic is being used.
		{
			"invalid info name - too long",
			NewScopeSpecification(
				ScopeSpecMetadataAddress(uuid.New()),
				NewDescription(strings.Repeat("x", maxDescriptionNameLength + 1), "", "", ""),
				[]string{}, []PartyType{}, []MetadataAddress{},
			),
			fmt.Sprintf("description (ScopeSpecification.Description) Name exceeds maximum length (expected <= %d got: %d)", maxDescriptionNameLength, maxDescriptionNameLength + 1),
			true,
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
			true,
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
			true,
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
			true,
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
			true,
		},
		// group spec ids - must all pass same tests as scope spec id (groupspec prefix)
		{
			"group spec ids - wrong address type at index 0",
			NewScopeSpecification(
				ScopeSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				[]MetadataAddress{MetadataAddress(specTestAddr)},
			),
			"invalid group specification id at index 0: invalid metadata address type (must be 0-4, actual: 133)",
			true,
		},
		{
			"group spec ids - wrong prefix at index 0",
			NewScopeSpecification(
				ScopeSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				[]MetadataAddress{ScopeMetadataAddress(uuid.New())},
			),
			"invalid group specification id prefix at index 0 (expected: groupspec, got scope)",
			true,
		},
		{
			"group spec ids - wrong address type at index 2",
			NewScopeSpecification(
				ScopeSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				[]MetadataAddress{GroupSpecMetadataAddress(uuid.New()), GroupSpecMetadataAddress(uuid.New()), MetadataAddress(specTestAddr)},
			),
			"invalid group specification id at index 2: invalid metadata address type (must be 0-4, actual: 133)",
			true,
		},
		{
			"group spec ids - wrong prefix at index 2",
			NewScopeSpecification(
				ScopeSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				[]MetadataAddress{GroupSpecMetadataAddress(uuid.New()), GroupSpecMetadataAddress(uuid.New()), ScopeMetadataAddress(uuid.New())},
			),
			"invalid group specification id prefix at index 2 (expected: groupspec, got scope)",
			true,
		},
		// Simple valid case
		{
			"simple valid case",
			NewScopeSpecification(
				ScopeSpecMetadataAddress(uuid.New()),
				nil,
				[]string{specTestBech32},
				[]PartyType{PartyType_PARTY_TYPE_OWNER},
				[]MetadataAddress{GroupSpecMetadataAddress(uuid.New())},
			),
			"",
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.spec.ValidateBasic()
			if (err != nil) != tt.wantErr {
				t.Errorf("Scope ValidateBasic error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Equal(t, tt.want, err.Error())
			}
		})
	}
}

func (s *specificationTestSuite) TestDescriptionValidateBasic() {
	tests := []struct {
		name     string
		desc     *Description
		want     string
		wantErr  bool
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
			fmt.Sprintf("info Name cannot be empty"),
			true,
		},
		{
			"invalid name - too long",
			NewDescription(
				strings.Repeat("x", maxDescriptionNameLength + 1),
				"",
				"",
				"",
			),
			fmt.Sprintf("info Name exceeds maximum length (expected <= %d got: %d)", maxDescriptionNameLength, maxDescriptionNameLength + 1),
			true,
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
			false,
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
			false,
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
			fmt.Sprintf("info Description exceeds maximum length (expected <= %d got: %d)", maxDescriptionDescriptionLength, maxDescriptionDescriptionLength + 1),
			true,
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
			false,
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
			false,
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
			false,
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
			true,
		},
		{
			"invalid website url - no protocol",
			NewDescription(
				"Unit Tests",
				"",
				"www.test.com",
				"",
			),
			fmt.Sprintf("url WebsiteUrl must begin with either http:// or https://"),
			true,
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
			false,
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
			false,
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
			false,
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
			false,
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
			true,
		},
		{
			"invalid icon url - no protocol",
			NewDescription(
				"Unit Tests",
				"",
				"",
				"www.test.com",
			),
			fmt.Sprintf("url IconUrl must begin with either http:// or https://"),
			true,
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
			false,
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
			false,
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
			false,
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
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.desc.ValidateBasic("")
			if (err != nil) != tt.wantErr {
				t.Errorf("Scope ValidateBasic error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Equal(t, tt.want, err.Error())
			}
		})
	}
}

func (s *specificationTestSuite) TestScopeSpecString() {
	s.T().Run("scope specification string", func(t *testing.T) {
		scopeSpecUuid := uuid.MustParse("c2074a03-6f6d-4029-bfe2-c3a5eb7e68b1")
		groupSpecUuid := uuid.MustParse("540dadf1-3dbc-4c3f-a205-7575b7f74384")
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
			[]MetadataAddress{GroupSpecMetadataAddress(groupSpecUuid)},
		)
		expected := "specification_id: scopespec1qnpqwjsrdak5q2dlutp6t6m7dzcscd7ff6\n" +
		            "description:\n" +
		            "  name: TestScopeSpecString Description\n" +
		            "  description: This is a description of a description used in a unit test.\n" +
		            "  website_url: https://figure.com/\n" +
		            "  icon_url: https://figure.com/favicon.png\n" +
		            "owner_addresses:\n" +
		            "- cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck\n" +
		            "parties_involved:\n" +
		            "- 5\n" +
		            "session_spec_ids:\n" +
		            "- groupspec1qd2qmt038k7yc0azq46htdlhgwzquwslkg\n"
		actual := scopeSpec.String()
		fmt.Printf("scopeSpecUuid: [%s]\n", scopeSpecUuid)
		fmt.Printf("groupSpecUuid: [%s]\n", groupSpecUuid)
		fmt.Println(actual)
		require.Equal(t, expected, actual)
	})
}

func (s *specificationTestSuite) TestDescriptionString() {
	s.T().Run("description string", func(t *testing.T) {
		description := NewDescription(
			"TestDescriptionString",
			"This is a description of a description used in a unit test.",
			"https://homestarrunner.com/",
			"https://homestarrunner.com/assets/sbemails/sbemail_hsrlogo_color.png",
		)
		expected := "name: TestDescriptionString\n" +
		            "description: This is a description of a description used in a unit test.\n" +
		            "website_url: https://homestarrunner.com/\n" +
		            "icon_url: https://homestarrunner.com/assets/sbemails/sbemail_hsrlogo_color.png\n"
		actual := description.String()
		require.Equal(t, expected, actual)
	})
}