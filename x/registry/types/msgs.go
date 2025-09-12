package types

import (
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
func (m *MsgRegisterNFT) ValidateBasic() error {
	// Verify the signer is a valid address
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		return NewErrCodeInvalidField("signer", m.Signer)
	}

	if err := m.Key.Validate(); err != nil {
		return err
	}

	// Validate roles
	for _, role := range m.Roles {
		if err := role.Validate(); err != nil {
			return NewErrCodeInvalidField("role", role.String())
		}
	}

	return nil
}

// ValidateBasic validates the MsgGrantRole message
func (m *MsgGrantRole) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		return NewErrCodeInvalidField("signer", m.Signer)
	}

	if err := m.Key.Validate(); err != nil {
		return err
	}

	if err := m.Role.Validate(); err != nil {
		return err
	}

	if err := validateAddresses(m.Addresses); err != nil {
		return err
	}

	return nil
}

// ValidateBasic validates the MsgRevokeRole message
func (m *MsgRevokeRole) ValidateBasic() error {
	// Verify the signer is a valid address
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		return NewErrCodeInvalidField("signer", m.Signer)
	}

	if err := m.Key.Validate(); err != nil {
		return err
	}

	if err := m.Role.Validate(); err != nil {
		return err
	}

	if err := validateAddresses(m.Addresses); err != nil {
		return err
	}

	return nil
}

// ValidateBasic validates the MsgUnregisterNFT message
func (m *MsgUnregisterNFT) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		return NewErrCodeInvalidField("signer", m.Signer)
	}

	if err := m.Key.Validate(); err != nil {
		return err
	}

	return nil
}

func validateAddresses(addrs []string) error {
	// Validate addresses
	if len(addrs) == 0 {
		return NewErrCodeInvalidField("addresses", "addresses cannot be empty")
	}
	for _, address := range addrs {
		if _, err := sdk.AccAddressFromBech32(address); err != nil {
			return NewErrCodeInvalidField("addresses", address)
		}
	}

	return nil
}

// ValidateBasic validates the QueryGetRegistryRequest
func (m *QueryGetRegistryRequest) ValidateBasic() error {
	if err := m.Key.Validate(); err != nil {
		return err
	}

	return nil
}

// ValidateBasic validates the QueryGetRegistriesRequest
func (m *QueryGetRegistriesRequest) ValidateBasic() error {
	if strings.TrimSpace(m.AssetClassId) == "" {
		return NewErrCodeInvalidField("asset_class_id", "asset_class_id cannot be empty if provided")
	}

	return nil
}

// ValidateBasic validates the QueryHasRoleRequest
func (m *QueryHasRoleRequest) ValidateBasic() error {
	if err := m.Key.Validate(); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(m.Address); err != nil {
		return NewErrCodeInvalidField("address", m.Address)
	}

	if err := m.Role.Validate(); err != nil {
		return err
	}

	return nil
}
