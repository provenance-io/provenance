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

	if len(m.Entries) == 0 || len(m.Entries) > 200 {
		errs = append(errs, fmt.Errorf("entries cannot be empty or greater than 200"))
	}

	for _, entry := range m.Entries {
		if err := entry.Validate(); err != nil {
			errs = append(errs, NewErrCodeInvalidField("entry", "%s", err))
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
	var errs []error
	if strings.TrimSpace(m.AssetClassId) == "" {
		errs = append(errs, NewErrCodeInvalidField("asset_class_id", "cannot be empty if provided"))
	}

	return errors.Join(errs...)
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
