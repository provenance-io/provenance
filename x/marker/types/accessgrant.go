package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
)

var (
	_ AccessGrantI = (*AccessGrant)(nil)
)

// AccessList is an array of access permissions
type AccessList = []Access

// AccessGrantI defines an interface for interacting with roles assigned to a given address.
type AccessGrantI interface {
	proto.Message
	Validate() error
	GetAddress() sdk.AccAddress

	HasAccess(Access) bool
	GetAccessList() []Access

	AddAccess(Access) error
	RemoveAccess(Access) error

	MergeAdd(AccessGrant) error
	MergeRemove(AccessGrant) error
}

// NewAccessGrant creates a new AccessGrant object
func NewAccessGrant(address sdk.AccAddress, access AccessList) *AccessGrant { // nolint:interfacer
	return &AccessGrant{
		Permissions: access,
		Address:     address.String(),
	}
}

// AccessByName returns the Access value given a name of the access type.  Normalizes input with
// proper ACCESS_ prefix and case of name.
func AccessByName(name string) Access {
	name = strings.ToUpper(strings.TrimSpace(name))
	name = strings.TrimPrefix(name, "ACCESS_")
	name = "ACCESS_" + name
	result := Access_value[name]
	return Access(result)
}

// AccessListByNames takes a comma separate list of names and returns an AccessList for the values
func AccessListByNames(names string) AccessList {
	// check to see if there is only "one" entry
	if !strings.Contains(names, ",") {
		return []Access{AccessByName(names)}
	}
	namelist := strings.Split(names, ",")
	result := make(AccessList, len(namelist))
	for i, name := range namelist {
		result[i] = AccessByName(name)
	}
	return result
}

// ValidateGrants checks a collection of grants and returns any errors encountered or nil
func ValidateGrants(grants ...AccessGrant) error {
	registered := make(map[string]bool)
	for _, grant := range grants {
		if err := grant.Validate(); err != nil {
			return err
		}
		if _, exists := registered[grant.Address]; exists {
			return ErrDuplicateAccessEntry
		}
	}
	return nil
}

// GrantsForAddress return
func GrantsForAddress(account sdk.AccAddress, grants ...AccessGrant) AccessGrant { // nolint:interfacer
	for _, grant := range grants {
		if grant.Address == account.String() {
			return grant
		}
	}
	return AccessGrant{account.String(), []Access{}}
}

// GetAddress returns the account address the access grant belongs to
func (ag AccessGrant) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(ag.Address)
	if err != nil {
		panic(err)
	}
	return addr
}

// GetAccessList returns the current list of access this grant holds
func (ag AccessGrant) GetAccessList() AccessList {
	return ag.Permissions
}

// Validate performs checks to ensure this acccess grant is properly formed.
func (ag AccessGrant) Validate() error {
	if _, err := sdk.AccAddressFromBech32(ag.Address); err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}
	return validateAccess(ag.Permissions)
}

// HasAccess returns true if the current grant contains the specified access type
func (ag AccessGrant) HasAccess(access Access) bool {
	if ag.Address == "" {
		return false
	}
	return hasAccess(ag.Permissions, access)
}

// AddAccess adds the specified access type to the current access grant
func (ag *AccessGrant) AddAccess(access Access) error {
	updated, err := addAccess(ag.Permissions, access)
	if err != nil {
		return err
	}
	ag.Permissions = updated
	return nil
}

// RemoveAccess removes the specified access type from the current access grant
func (ag *AccessGrant) RemoveAccess(access Access) error {
	updated, err := removeAccess(ag.Permissions, access)
	if err != nil {
		return err
	}
	ag.Permissions = updated
	return nil
}

// MergeAdd looks for any missing permissions in the given grant and adds them to this instance.
func (ag *AccessGrant) MergeAdd(other AccessGrant) error {
	if err := other.Validate(); err != nil {
		return err
	}
	if other.Address != ag.Address {
		return fmt.Errorf("cannot merge in AccessGrant for different address")
	}
	for _, p := range other.GetAccessList() {
		if !ag.HasAccess(p) {
			ag.Permissions = append(ag.Permissions, p)
		}
	}
	return nil
}

// MergeRemove looks for permissions in this instance that exist in the given grant and removes them.
func (ag *AccessGrant) MergeRemove(other AccessGrant) error {
	if err := other.Validate(); err != nil {
		return err
	}
	if other.Address != ag.Address {
		return fmt.Errorf("cannot merge in AccessGrant for different address")
	}
	var newPerms []Access
	for _, p := range ag.Permissions {
		if !other.HasAccess(p) {
			newPerms = append(newPerms, p)
		}
	}
	ag.Permissions = newPerms
	return nil
}

// String implements stringer
func (ag AccessGrant) String() string {
	result := ""
	for _, p := range ag.Permissions {
		perm := strings.ToLower(strings.TrimPrefix(p.String(), "ACCESS_"))
		if result == "" {
			result = perm
		} else {
			result = fmt.Sprintf("%s, %s", result, perm)
		}
	}
	return fmt.Sprintf("AccessGrant: %s [%s]", ag.Address, result)
}

// IsOneOf returns true if the specified Access right is any of the provided options.
func (right Access) IsOneOf(rights ...Access) bool {
	if len(rights) == 0 {
		return false
	}
	for _, r := range rights {
		if right == r {
			return true
		}
	}
	return false
}

// Validate checks to see that the access list only contains valid entries and no duplicates
func validateAccess(accessList AccessList) error {
	// Empty list can't have invalid entries
	if len(accessList) == 0 {
		return nil
	}
	registered := make(map[Access]bool)
	for _, access := range accessList {
		if _, exists := registered[access]; exists {
			return ErrDuplicateAccessEntry
		}
		if access == Access_Unknown {
			return ErrAccessTypeInvalid
		}
		registered[access] = true
	}
	return nil
}

// HasAccess returns true if the AccessGrant allows the given access
func hasAccess(accessList AccessList, access Access) bool {
	// Empty addresses can have no access.
	if len(accessList) == 0 {
		return false
	}
	for _, a := range accessList {
		if a == access {
			return true
		}
	}
	return false
}

// AddAccess adds the given access to the array of Access privledges if not included
func addAccess(accessList AccessList, access Access) (AccessList, error) {
	if access == Access_Unknown {
		return nil, ErrAccessTypeInvalid
	}
	if hasAccess(accessList, access) {
		return nil, ErrDuplicateAccessEntry
	}
	return append(accessList, access), nil
}

// RemoveAccess removes the given access from the array of Access privledges (if included)
func removeAccess(accessList AccessList, access Access) (AccessList, error) {
	if access == Access_Unknown {
		return nil, ErrAccessTypeInvalid
	}
	if !hasAccess(accessList, access) {
		return nil, ErrAccessTypeNotGranted
	}
	var updatedAccess AccessList
	for _, a := range accessList {
		if a != access {
			updatedAccess = append(updatedAccess, a)
		}
	}
	return updatedAccess, nil
}
