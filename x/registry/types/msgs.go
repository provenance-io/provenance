package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ValidateBasic validates the MsgRegisterNFT message
func (m *MsgRegisterNFT) ValidateBasic() error {
	// Verify the authority is a valid address
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	if err := m.Key.Validate(); err != nil {
		return err
	}

	// Validate roles
	for i, role := range m.Roles {
		if err := role.Validate(); err != nil {
			return sdkerrors.ErrInvalidRequest.Wrapf("invalid role at index %d: %s", i, err)
		}
	}

	return nil
}

// ValidateBasic validates the MsgGrantRole message
func (m *MsgGrantRole) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
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
	// Verify the authority is a valid address
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
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
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	if err := m.Key.Validate(); err != nil {
		return err
	}

	return nil
}

func validateAddresses(addrs []string) error {
	// Validate addresses
	if len(addrs) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("addresses cannot be empty")
	}
	for i, address := range addrs {
		if _, err := sdk.AccAddressFromBech32(address); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid address at index %d: %s", i, err)
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

// ValidateBasic validates the QueryHasRoleRequest
func (m *QueryHasRoleRequest) ValidateBasic() error {
	if err := m.Key.Validate(); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(m.Address); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid address: %s", err)
	}

	if err := m.Role.Validate(); err != nil {
		return err
	}

	return nil
}
