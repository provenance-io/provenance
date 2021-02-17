package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Max length for info.name
	maxDescriptionNameLength = 200
	// Max length for info.description
	maxDescriptionDescriptionLength = 5000
	// Max url length
	maxUrlLength = 2048
)

// NewScopeSpecification creates a new ScopeSpecification instance.
func NewScopeSpecification(
	specificationId MetadataAddress,
	description *Description,
	ownerAddresses []string,
	partiesInvolved []PartyType,
	sessionSpecIds []MetadataAddress,
) *ScopeSpecification {
	return &ScopeSpecification{
		SpecificationId: specificationId,
		Description:     description,
		OwnerAddresses:  ownerAddresses,
		PartiesInvolved: partiesInvolved,
		SessionSpecIds:  sessionSpecIds,
	}
}

// ValidateBasic performs basic format checking of data in a ScopeSpecification
func (scopeSpec *ScopeSpecification) ValidateBasic() error {
	prefix, err := VerifyMetadataAddressFormat(scopeSpec.SpecificationId)
	if err != nil {
		return fmt.Errorf("invalid scope specification id: %w", err)
	}
	if prefix != PrefixScopeSpecification {
		return fmt.Errorf("invalid scope specification id prefix (expected: %s, got %s)", PrefixScopeSpecification, prefix)
	}
	if scopeSpec.Description != nil {
		err = scopeSpec.Description.ValidateBasic("ScopeSpecification.Description")
		if err != nil {
			return err
		}
	}
	if len(scopeSpec.OwnerAddresses) < 1 {
		return errors.New("ScopeSpecification must have at least one owner")
	}
	for i, owner := range scopeSpec.OwnerAddresses {
		if _, err = sdk.AccAddressFromBech32(owner); err != nil {
			return fmt.Errorf("invalid owner address at index %d on ScopeSpecification: %w", i, err)
		}
	}
	if len(scopeSpec.PartiesInvolved) == 0 {
		return errors.New("ScopeSpecification must have at least one party involved")
	}
	for i, groupSpecId := range scopeSpec.SessionSpecIds {
		prefix, err = VerifyMetadataAddressFormat(groupSpecId)
		if err != nil {
			return fmt.Errorf("invalid group specification id at index %d: %w", i, err)
		}
		if prefix != PrefixGroupSpecification {
			return fmt.Errorf("invalid group specification id prefix at index %d (expected: %s, got %s)",
				i, PrefixGroupSpecification, prefix)
		}
	}
	return nil
}

// NewDescription creates a new Description instance.
func NewDescription(name, description, websiteUrl, iconUrl string) *Description {
	return &Description{
		Name:        name,
		Description: description,
		WebsiteUrl:  websiteUrl,
		IconUrl:     iconUrl,
	}
}

// ValidateBasic performs basic format checking of data in an Description.
// The path parameter is used to provide extra context to any error messages.
// e.g. If the name field is invalid in this info, and the path provided is "ScopeSpecification.Description",
// the error message will contain "ScopeSpecification.Description.Name" and the problem.
// Provide "" if there is no context you wish to provide.
func (info *Description) ValidateBasic(path string) error {
	if len(info.Name) == 0 {
		return fmt.Errorf("info %s cannot be empty", makeFieldString(path, "Name"))
	}
	if len(info.Name) > maxDescriptionNameLength {
		return fmt.Errorf("info %s exceeds maximum length (expected <= %d got: %d)",
			makeFieldString(path, "Name"), maxDescriptionNameLength, len(info.Name))
	}
	if len(info.Description) > maxDescriptionDescriptionLength {
		return fmt.Errorf("info %s exceeds maximum length (expected <= %d got: %d)",
			makeFieldString(path, "Description"), maxDescriptionDescriptionLength, len(info.Description))
	}
	err := validateUrlBasic(info.WebsiteUrl, false, path, "WebsiteUrl")
	if err != nil {
		return err
	}
	err = validateUrlBasic(info.IconUrl, false, path, "IconUrl")
	if err != nil {
		return err
	}
	return nil
}

// validateUrlBasic - Helper function to check if a url string is superficially valid.
// The path and fieldName parameters are combined using makeFieldString for error messages.
func validateUrlBasic(url string, required bool, path string, fieldName string) error {
	if len(url) == 0 {
		if required {
			return fmt.Errorf("url %s cannot be empty", makeFieldString(path, fieldName))
		}
		return nil
	}
	if len(url) > maxUrlLength {
		return fmt.Errorf("url %s exceeds maximum length (expected <= %d got: %d)",
			makeFieldString(path, fieldName), maxUrlLength, len(url))
	}
	if strings.ToLower(url[0:8]) != "https://" && strings.ToLower(url[0:7]) != "http://" {
		return fmt.Errorf("url %s must begin with either http:// or https://", makeFieldString(path, fieldName))
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