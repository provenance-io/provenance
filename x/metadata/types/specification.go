package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Max length for info.name
	maxInfoNameLength = 200
	// Max length for info.description
	maxInfoDescriptionLength = 5000
	// Max url length
	maxUrlLength = 2048
)

// NewScopeSpecification creates a new ScopeSpecification instance.
func NewScopeSpecification(
	specificationId MetadataAddress,
	info *Info,
	ownerAddresses []string,
	partiesInvolved []PartyType,
	groupSpecIds []MetadataAddress,
) *ScopeSpecification {
	return &ScopeSpecification{
		SpecificationId: specificationId,
		Info:            info,
		OwnerAddresses:  ownerAddresses,
		PartiesInvolved: partiesInvolved,
		GroupSpecIds:    groupSpecIds,
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
	if scopeSpec.Info != nil {
		err = scopeSpec.Info.ValidateBasic("ScopeSpecification.Info")
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
	for i, groupSpecId := range scopeSpec.GroupSpecIds {
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

// NewInfo creates a new Info instance.
func NewInfo(name, description, websiteUrl, iconUrl string) *Info {
	return &Info{
		Name:        name,
		Description: description,
		WebsiteUrl:  websiteUrl,
		IconUrl:     iconUrl,
	}
}

// ValidateBasic performs basic format checking of data in an Info.
// The path parameter is used to provide extra context to any error messages.
// e.g. If the name field is invalid in this info, and the path provided is "ScopeSpecification.Info",
// the error message will contain "ScopeSpecification.Info.Name" and the problem.
// Provide "" if there is no context you wish to provide.
func (info *Info) ValidateBasic(path string) error {
	if len(info.Name) == 0 {
		return fmt.Errorf("info %s cannot be empty", makeFieldString(path, "Name"))
	}
	if len(info.Name) > maxInfoNameLength {
		return fmt.Errorf("info %s exceeds maximum length (expected <= %d got: %d)",
			makeFieldString(path, "Name"), maxInfoNameLength, len(info.Name))
	}
	if len(info.Description) > maxInfoDescriptionLength {
		return fmt.Errorf("info %s exceeds maximum length (expected <= %d got: %d)",
			makeFieldString(path, "Description"), maxInfoDescriptionLength, len(info.Description))
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
	if strings.ToLower(url[0:8]) != "https://" && strings.ToLower(url[0:7]) != "http://" {
		return fmt.Errorf("url %s must begin with either http:// or https://", makeFieldString(path, fieldName))
	}
	if len(url) > maxUrlLength {
		return fmt.Errorf("url %s exceeds maximum length (expected <= %d got: %d)",
			makeFieldString(path, fieldName), maxUrlLength, len(url))
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