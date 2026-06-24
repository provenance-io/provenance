package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var AllRequestMsgs = []sdk.Msg{
	(*MsgRegisterNFT)(nil),
	(*MsgGrantRole)(nil),
	(*MsgRevokeRole)(nil),
	(*MsgUnregisterNFT)(nil),
	(*MsgRegistryBulkUpdate)(nil),
	(*MsgSetRoles)(nil),
	(*MsgProposeRoleChange)(nil),
	(*MsgApproveRoleChange)(nil),
	(*MsgCreateRegistryClass)(nil),
	(*MsgUpdateRegistryClassRoleAuthorization)(nil),
	(*MsgUpdateParams)(nil),
}

// ValidateBasic validates the MsgRegisterNFT message
func (m MsgRegisterNFT) ValidateBasic() error {
	var errs []error
	// Verify the signer is a valid address
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if err := m.Key.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("key", "%s", err))
	}

	// Validate roles
	for _, role := range m.Roles {
		if err := role.Validate(); err != nil {
			errs = append(errs, NewErrCodeInvalidField("role", "%s", err))
		}
	}

	// registry_class_id is optional; validate format only when set.
	if m.RegistryClassId != "" {
		if err := ValidateRegistryClassID(m.RegistryClassId); err != nil {
			errs = append(errs, NewErrCodeInvalidField("registry_class_id", "%s", err))
		}
	}

	return errors.Join(errs...)
}

// ValidateBasic validates the MsgGrantRole message
func (m MsgGrantRole) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if err := m.Key.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("key", "%s", err))
	}

	if err := m.Role.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("role", "%s", err))
	}

	if err := validateAddresses(m.Addresses); err != nil {
		errs = append(errs, NewErrCodeInvalidField("addresses", "%s", err))
	}

	return errors.Join(errs...)
}

// ValidateBasic validates the MsgRevokeRole message
func (m MsgRevokeRole) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if err := m.Key.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("key", "%s", err))
	}

	if err := m.Role.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("role", "%s", err))
	}

	if err := validateAddresses(m.Addresses); err != nil {
		errs = append(errs, NewErrCodeInvalidField("addresses", "%s", err))
	}

	return errors.Join(errs...)
}

// ValidateBasic validates the MsgSetRoles message
func (m MsgSetRoles) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if err := m.Key.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("key", "%s", err))
	}

	errs = append(errs, validateRoleUpdates(m.RoleUpdates)...)

	return errors.Join(errs...)
}

// ValidateBasic validates the MsgProposeRoleChange message
func (m MsgProposeRoleChange) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if err := m.Key.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("key", "%s", err))
	}

	errs = append(errs, validateRoleUpdates(m.RoleUpdates)...)

	return errors.Join(errs...)
}

// validateRoleUpdates validates a batch of desired-state role updates. Each role must be a known,
// specified enum value and may appear at most once across the batch (the batch is a desired-state
// map, so a repeated role is ambiguous). Addresses within an update must be valid and free of
// duplicates, matching the invariants enforced by RolesEntry.Validate once the batch is applied.
func validateRoleUpdates(updates []RoleUpdate) []error {
	if len(updates) == 0 {
		return []error{NewErrCodeInvalidField("role_updates", "at least one role update is required")}
	}

	var errs []error
	seenRoles := make(map[RegistryRole]bool, len(updates))
	for i, update := range updates {
		if err := update.Role.Validate(); err != nil {
			errs = append(errs, NewErrCodeInvalidField("role_updates", "%d: role %s", i, err))
		} else if seenRoles[update.Role] {
			errs = append(errs, NewErrCodeInvalidField("role_updates", "%d: duplicate role %s", i, update.Role.ShortString()))
		} else {
			seenRoles[update.Role] = true
		}

		seenAddrs := make(map[string]bool, len(update.Addresses))
		for _, addr := range update.Addresses {
			if _, err := sdk.AccAddressFromBech32(addr); err != nil {
				errs = append(errs, NewErrCodeInvalidField("role_updates", "%d: invalid address: %s", i, err))
				continue
			}
			if seenAddrs[addr] {
				errs = append(errs, NewErrCodeInvalidField("role_updates", "%d: duplicate address: %s", i, addr))
				continue
			}
			seenAddrs[addr] = true
		}
	}
	return errs
}

// ValidateBasic validates the MsgApproveRoleChange message
func (m MsgApproveRoleChange) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if len(m.ChangeId) == 0 {
		errs = append(errs, NewErrCodeInvalidField("change_id", "change_id is required"))
	}

	return errors.Join(errs...)
}

// ValidateBasic validates the MsgCreateRegistryClass message
func (m MsgCreateRegistryClass) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if _, err := sdk.AccAddressFromBech32(m.Maintainer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("maintainer", "%s", err))
	}

	// The signer must be the maintainer being set.
	if m.Signer != m.Maintainer {
		errs = append(errs, NewErrCodeInvalidField("signer", "signer %q must match maintainer %q", m.Signer, m.Maintainer))
	}

	if err := ValidateRegistryClassID(m.RegistryClassId); err != nil {
		errs = append(errs, NewErrCodeInvalidField("registry_class_id", "%s", err))
	}

	if err := ValidateClassID(m.AssetClassId); err != nil {
		errs = append(errs, NewErrCodeInvalidField("asset_class_id", "%s", err))
	}

	errs = append(errs, validateRoleAuthorizations(m.RoleAuthorizations)...)

	return errors.Join(errs...)
}

// ValidateBasic validates the MsgUpdateRegistryClassRoleAuthorization message
func (m MsgUpdateRegistryClassRoleAuthorization) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if err := ValidateRegistryClassID(m.RegistryClassId); err != nil {
		errs = append(errs, NewErrCodeInvalidField("registry_class_id", "%s", err))
	}

	errs = append(errs, validateRoleAuthorizations(m.RoleAuthorizations)...)

	return errors.Join(errs...)
}

// ValidateBasic validates the MsgUpdateParams message
func (m MsgUpdateParams) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		errs = append(errs, NewErrCodeInvalidField("authority", "%s", err))
	}

	if err := m.Params.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("params", "%s", err))
	}

	return errors.Join(errs...)
}

// ValidateBasic validates the MsgUnregisterNFT message
func (m MsgUnregisterNFT) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if err := m.Key.Validate(); err != nil {
		errs = append(errs, NewErrCodeInvalidField("key", "%s", err))
	}

	return errors.Join(errs...)
}

// ValidateBasic validates the MsgRegistryBulkUpdate message
func (m MsgRegistryBulkUpdate) ValidateBasic() error {
	var errs []error
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		errs = append(errs, NewErrCodeInvalidField("signer", "%s", err))
	}

	if len(m.Entries) == 0 || len(m.Entries) > MaxRegistryBulkEntries {
		errs = append(errs, fmt.Errorf("entries cannot be empty or greater than %d", MaxRegistryBulkEntries))
	}

	for i, entry := range m.Entries {
		if err := entry.Validate(); err != nil {
			errs = append(errs, NewErrCodeInvalidField("entry", "%d: %s", i, err))
		}
	}

	return errors.Join(errs...)
}

func validateAddresses(addrs []string) error {
	var errs []error
	// Validate addresses
	if len(addrs) == 0 {
		errs = append(errs, fmt.Errorf("addresses cannot be empty"))
	}
	for _, address := range addrs {
		if _, err := sdk.AccAddressFromBech32(address); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// Validate validates the QueryGetRegistryRequest
func (m QueryGetRegistryRequest) Validate() error {
	var errs []error
	if err := m.Key.Validate(); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// Validate validates the QueryGetRegistriesRequest
func (m QueryGetRegistriesRequest) Validate() error {
	// The AssetClassId is optional, and there's nothing to validate on it.
	// There's nothing to validate with the pagination either.
	return nil
}

// Validate validates the QueryHasRoleRequest
func (m QueryHasRoleRequest) Validate() error {
	var errs []error
	if err := m.Key.Validate(); err != nil {
		errs = append(errs, err)
	}

	if _, err := sdk.AccAddressFromBech32(m.Address); err != nil {
		errs = append(errs, err)
	}

	if err := m.Role.Validate(); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// Validate validates the QueryPendingRoleChangeRequest
func (m QueryPendingRoleChangeRequest) Validate() error {
	if len(m.Id) == 0 {
		return NewErrCodeInvalidField("id", "id cannot be empty")
	}
	return nil
}

// Validate validates the QueryPendingRoleChangesRequest
func (m QueryPendingRoleChangesRequest) Validate() error {
	// The key is optional; validate it only when provided.
	if m.Key != nil {
		return m.Key.Validate()
	}
	return nil
}

// Validate validates the QueryRegistryClassRequest
func (m QueryRegistryClassRequest) Validate() error {
	if len(m.RegistryClassId) == 0 {
		return NewErrCodeInvalidField("registry_class_id", "registry_class_id cannot be empty")
	}
	return nil
}

// Validate validates the QueryRegistryClassesRequest
func (m QueryRegistryClassesRequest) Validate() error {
	// There's nothing to validate; pagination is optional.
	return nil
}

// ValidateStringLength checks several conditions in order:
//  1. length is not between the minimum and maximum length returns an ErrCodeInvalidField error.
//  2. contains whitespace returns an ErrCodeInvalidField error.
//  3. contains a null byte returns an ErrCodeInvalidField error.
//
// returns nil if all conditions are met.
func ValidateStringLength(str string, minLength int, maxLength int) error {
	var errs []error

	if len(str) > maxLength || len(str) < minLength {
		errs = append(errs, fmt.Errorf("must be between %d and %d characters", minLength, maxLength))
	}

	// cannot contain whitespace
	if len(str) != len(strings.TrimSpace(str)) {
		errs = append(errs, fmt.Errorf("cannot contain whitespace"))
	}

	// cannot contain a null byte
	if strings.Contains(str, "\x00") {
		errs = append(errs, fmt.Errorf("cannot contain a null byte"))
	}

	return errors.Join(errs...)
}
