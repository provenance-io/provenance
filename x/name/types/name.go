package types

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewNameRecord creates a name record binding that is restricted for child updates to the owner.
func NewNameRecord(name string, address sdk.AccAddress, restricted bool) NameRecord {
	return NameRecord{
		Name:       name,
		Address:    address.String(),
		Restricted: restricted,
	}
}

// implement fmt.Stringer
func (nr NameRecord) String() string {
	if nr.Restricted {
		return strings.TrimSpace(fmt.Sprintf(`%s: %s [restricted]`, nr.Name, nr.Address))
	}
	return strings.TrimSpace(fmt.Sprintf(`%s: %s`, nr.Name, nr.Address))
}

// Validate performs basic stateless validity checks.
func (nr NameRecord) Validate() error {
	if strings.TrimSpace(nr.Address) == "" {
		return ErrInvalidAddress
	}
	if strings.TrimSpace(nr.Name) == "" {
		return ErrNameSegmentTooShort
	}
	return nil
}

// NormalizeName lower-cases and strips out spaces around each segment in the provided string.
func NormalizeName(name string) string {
	nameSegments := strings.Split(name, ".")
	segments := make([]string, len(nameSegments))
	for i, nameSegment := range nameSegments {
		segments[i] = strings.ToLower(strings.TrimSpace(nameSegment))
	}
	return strings.Join(segments, ".")
}

// IsValidName returns true if the provided name is valid (without consideration of length limits).
// It is assumed that the name provided has already been normalized using NormalizeName.
func IsValidName(name string) bool {
	return ValidateName(name) == nil
}

// ValidateName returns an error if the provided name is not valid (without consideration of length limits).
// It is assumed that the name provided has already been normalized using NormalizeName.
func ValidateName(name string) error {
	for _, segment := range strings.Split(name, ".") {
		if err := ValidateNameSegment(segment); err != nil {
			return fmt.Errorf("invalid name %q: %w", name, err)
		}
	}
	return nil
}

// IsValidNameSegment returns true if the provided string is a valid name segment.
// Name segments can only contain dashes, digits, and lower-case letters.
// If it's not a uuid, it can have at most one dash.
//
// The length of the segment is not considered here because the length limits are defined in state.
func IsValidNameSegment(segment string) bool {
	return ValidateNameSegment(segment) == nil
}

// ValidateNameSegment returns an error if the provided string is not a valid name segment.
// Name segments can only contain dashes, digits, and lower-case letters.
// If it's not a uuid, it can have at most one dash.
//
// The length of the segment is not considered here because the length limits are defined in state.
func ValidateNameSegment(segment string) error {
	// Allow valid UUID
	if IsValidUUID(segment) {
		return nil
	}
	// Only allow a single dash if not a UUID
	if strings.Count(segment, "-") > 1 {
		return fmt.Errorf("segment %q has too many dashes", segment)
	}
	// Only allow dashes, lowercase characters and digits.
	for _, c := range segment {
		if c != '-' && !unicode.IsLower(c) && !unicode.IsDigit(c) {
			return fmt.Errorf("illegal character %q in name segment %q", string(c), segment)
		}
	}
	return nil
}

// IsValidUUID returns true if the provided string is a valid UUID string.
func IsValidUUID(str string) bool {
	if _, err := uuid.Parse(str); err == nil {
		return true
	}
	return false
}
