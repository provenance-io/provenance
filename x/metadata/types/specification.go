package types

import (
	"errors"
	"fmt"
	"regexp"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gopkg.in/yaml.v2"
)

const (
	// TODO: Move these to params.
	// Default max length for Description.Name
	maxDescriptionNameLength = 200
	// Default max length for Description.Description
	maxDescriptionDescriptionLength = 5000
	// Default max length for a ContractSpecification.ClassName
	maxContractSpecificationClassNameLength = 1000
	// Default max url length
	maxURLLength = 2048
)

var (
	urlProtocolsAllowedRegexps = []*regexp.Regexp{
		regexp.MustCompile("https?://"),
		regexp.MustCompile("data:.*,"),
		// If you alter this, make sure to alter the associated error message.
	}
)

// NewScopeSpecification creates a new ScopeSpecification instance.
func NewScopeSpecification(
	specificationID MetadataAddress,
	description *Description,
	ownerAddresses []string,
	partiesInvolved []PartyType,
	contractSpecIDs []MetadataAddress,
) *ScopeSpecification {
	return &ScopeSpecification{
		SpecificationId: specificationID,
		Description:     description,
		OwnerAddresses:  ownerAddresses,
		PartiesInvolved: partiesInvolved,
		ContractSpecIds: contractSpecIDs,
	}
}

// ValidateBasic performs basic format checking of data in a ScopeSpecification
func (s *ScopeSpecification) ValidateBasic() error {
	prefix, err := VerifyMetadataAddressFormat(s.SpecificationId)
	if err != nil {
		return fmt.Errorf("invalid scope specification id: %w", err)
	}
	if prefix != PrefixScopeSpecification {
		return fmt.Errorf("invalid scope specification id prefix (expected: %s, got %s)", PrefixScopeSpecification, prefix)
	}
	if s.Description != nil {
		err = s.Description.ValidateBasic("ScopeSpecification.Description")
		if err != nil {
			return err
		}
	}
	if len(s.OwnerAddresses) < 1 {
		return errors.New("the ScopeSpecification must have at least one owner")
	}
	for i, owner := range s.OwnerAddresses {
		if _, err = sdk.AccAddressFromBech32(owner); err != nil {
			return fmt.Errorf("invalid owner address at index %d on ScopeSpecification: %w", i, err)
		}
	}
	if len(s.PartiesInvolved) == 0 {
		return errors.New("the ScopeSpecification must have at least one party involved")
	}
	for i, contractSpecID := range s.ContractSpecIds {
		prefix, err = VerifyMetadataAddressFormat(contractSpecID)
		if err != nil {
			return fmt.Errorf("invalid contract specification id at index %d: %w", i, err)
		}
		if prefix != PrefixContractSpecification {
			return fmt.Errorf("invalid contract specification id prefix at index %d (expected: %s, got %s)",
				i, PrefixContractSpecification, prefix)
		}
	}
	return nil
}

// String implements stringer interface
func (s ScopeSpecification) String() string {
	out, _ := yaml.Marshal(s)
	return string(out)
}

// NewScopeSpecification creates a new ScopeSpecification instance.
func NewContractSpecification(
	specificationID MetadataAddress,
	description *Description,
	ownerAddresses []string,
	partiesInvolved []PartyType,
	source isContractSpecification_Source,
	className string,
	conditionSpecs []*RecordSpecification,
) *ContractSpecification {
	return &ContractSpecification{
		SpecificationId: specificationID,
		Description:     description,
		OwnerAddresses:  ownerAddresses,
		PartiesInvolved: partiesInvolved,
		Source:          source,
		ClassName:       className,
		ConditionSpecs:  conditionSpecs,
	}
}

// ValidateBasic performs basic format checking of data in a ScopeSpecification
func (s *ContractSpecification) ValidateBasic() error {
	prefix, err := VerifyMetadataAddressFormat(s.SpecificationId)
	if err != nil {
		return fmt.Errorf("invalid contract specification id: %w", err)
	}
	if prefix != PrefixContractSpecification {
		return fmt.Errorf("invalid contract specification id prefix (expected: %s, got %s)", PrefixContractSpecification, prefix)
	}
	if s.Description != nil {
		err = s.Description.ValidateBasic("ContractSpecification.Description")
		if err != nil {
			return err
		}
	}
	if len(s.OwnerAddresses) == 0 {
		return fmt.Errorf("invalid owner addresses count (expected > 0 got: %d)", len(s.OwnerAddresses))
	}
	for i, owner := range s.OwnerAddresses {
		if _, err = sdk.AccAddressFromBech32(owner); err != nil {
			return fmt.Errorf("invalid owner address at index %d: %w", i, err)
		}
	}
	for i, owner := range s.OwnerAddresses {
		if _, err = sdk.AccAddressFromBech32(owner); err != nil {
			return fmt.Errorf("invalid owner address at index %d: %w", i, err)
		}
	}
	if len(s.PartiesInvolved) == 0 {
		return fmt.Errorf("invalid parties involved count (expected > 0 got: %d)", len(s.PartiesInvolved))
	}
	switch source := s.Source.(type) {
	case *ContractSpecification_ResourceId:
		_, err = VerifyMetadataAddressFormat(source.ResourceId)
		if err != nil {
			return fmt.Errorf("invalid source resource id: %w", err)
		}
	case *ContractSpecification_Hash:
		if len(source.Hash) == 0 {
			return errors.New("source hash cannot be empty")
		}
	default:
		return errors.New("unknown source type")
	}
	if len(s.ClassName) == 0 {
		return errors.New("class name cannot be empty")
	}
	if len(s.ClassName) > maxContractSpecificationClassNameLength {
		return fmt.Errorf("class name exceeds maximum length (expected <= %d got: %d)",
			maxContractSpecificationClassNameLength, len(s.ClassName))
	}
	if len(s.ConditionSpecs) == 0 {
		return fmt.Errorf("invalid condition specs count (expected > 0 got: %d)", len(s.ConditionSpecs))
	}
	for i, c := range s.ConditionSpecs {
		if err = c.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid condition spec at index %d: %w", i, err)
		}
	}
	return nil
}

// String implements stringer interface
func (s ContractSpecification) String() string {
	out, _ := yaml.Marshal(s)
	return string(out)
}

func (s *RecordSpecification) ValidateBasic() error {
	// TODO: implement
	return nil
}

// NewDescription creates a new Description instance.
func NewDescription(name, description, websiteURL, iconURL string) *Description {
	return &Description{
		Name:        name,
		Description: description,
		WebsiteUrl:  websiteURL,
		IconUrl:     iconURL,
	}
}

// ValidateBasic performs basic format checking of data in an Description.
// The path parameter is used to provide extra context to any error messages.
// e.g. If the name field is invalid in this description, and the path provided is "ScopeSpecification.Description",
// the error message will contain "ScopeSpecification.Description.Name" and the problem.
// Provide "" if there is no context you wish to provide.
func (d *Description) ValidateBasic(path string) error {
	if len(d.Name) == 0 {
		return fmt.Errorf("description %s cannot be empty", makeFieldString(path, "Name"))
	}
	if len(d.Name) > maxDescriptionNameLength {
		return fmt.Errorf("description %s exceeds maximum length (expected <= %d got: %d)",
			makeFieldString(path, "Name"), maxDescriptionNameLength, len(d.Name))
	}
	if len(d.Description) > maxDescriptionDescriptionLength {
		return fmt.Errorf("description %s exceeds maximum length (expected <= %d got: %d)",
			makeFieldString(path, "Description"), maxDescriptionDescriptionLength, len(d.Description))
	}
	err := validateURLBasic(d.WebsiteUrl, false, path, "WebsiteUrl")
	if err != nil {
		return err
	}
	err = validateURLBasic(d.IconUrl, false, path, "IconUrl")
	if err != nil {
		return err
	}
	return nil
}

// String implements stringer interface
func (d Description) String() string {
	out, _ := yaml.Marshal(d)
	return string(out)
}

// validateURLBasic - Helper function to check if a url string is superficially valid.
// The path and fieldName parameters are combined using makeFieldString for error messages.
func validateURLBasic(url string, required bool, path string, fieldName string) error {
	if len(url) == 0 {
		if required {
			return fmt.Errorf("url %s cannot be empty", makeFieldString(path, fieldName))
		}
		return nil
	}
	if len(url) > maxURLLength {
		return fmt.Errorf("url %s exceeds maximum length (expected <= %d got: %d)",
			makeFieldString(path, fieldName), maxURLLength, len(url))
	}
	isAllowedProtocol := false
	for _, r := range urlProtocolsAllowedRegexps {
		if r.MatchString(url) {
			isAllowedProtocol = true
			break
		}
	}
	if !isAllowedProtocol {
		return fmt.Errorf("url %s must use the http, https, or data protocol", makeFieldString(path, fieldName))
	}
	return nil
}

// makeFieldString: Helper function to create a string for a field meant for use in an error message for that field.
// If path is empty, then fieldName is returned unaltered.
// If path is not empty, "(path) fieldName" is returned.
func makeFieldString(path string, fieldName string) string {
	if len(path) == 0 {
		return fieldName
	}
	return fmt.Sprintf("(%s) %s", path, fieldName)
}
