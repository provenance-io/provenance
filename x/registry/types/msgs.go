package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ValidateBasic validates the MsgRegisterNFT message
func (m *MsgRegisterNFT) ValidateBasic() error {
	// Validate authority
	if strings.TrimSpace(m.Authority) == "" {
		return sdkerrors.ErrInvalidAddress.Wrap("authority cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	// Validate key
	if m.Key == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("key cannot be nil")
	}
	if err := m.Key.Validate(); err != nil {
		return err
	}

	// Validate roles
	if len(m.Roles) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("roles cannot be empty")
	}
	for i, role := range m.Roles {
		if err := role.Validate(); err != nil {
			return sdkerrors.ErrInvalidRequest.Wrapf("invalid role at index %d: %s", i, err)
		}
	}

	return nil
}

// ValidateBasic validates the MsgGrantRole message
func (m *MsgGrantRole) ValidateBasic() error {
	// Validate authority
	if strings.TrimSpace(m.Authority) == "" {
		return sdkerrors.ErrInvalidAddress.Wrap("authority cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	// Validate key
	if m.Key == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("key cannot be nil")
	}
	if err := m.Key.Validate(); err != nil {
		return err
	}

	// Validate role
	if m.Role == RegistryRole_REGISTRY_ROLE_UNSPECIFIED {
		return sdkerrors.ErrInvalidRequest.Wrap("role cannot be unspecified")
	}

	// Validate addresses
	if len(m.Addresses) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("addresses cannot be empty")
	}
	for i, address := range m.Addresses {
		if strings.TrimSpace(address) == "" {
			return sdkerrors.ErrInvalidAddress.Wrapf("address at index %d cannot be empty", i)
		}
		if _, err := sdk.AccAddressFromBech32(address); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid address at index %d: %s", i, err)
		}
	}

	return nil
}

// ValidateBasic validates the MsgRevokeRole message
func (m *MsgRevokeRole) ValidateBasic() error {
	// Validate authority
	if strings.TrimSpace(m.Authority) == "" {
		return sdkerrors.ErrInvalidAddress.Wrap("authority cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	// Validate key
	if m.Key == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("key cannot be nil")
	}
	if err := m.Key.Validate(); err != nil {
		return err
	}

	// Validate role
	if m.Role == RegistryRole_REGISTRY_ROLE_UNSPECIFIED {
		return sdkerrors.ErrInvalidRequest.Wrap("role cannot be unspecified")
	}

	// Validate addresses
	if len(m.Addresses) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("addresses cannot be empty")
	}
	for i, address := range m.Addresses {
		if strings.TrimSpace(address) == "" {
			return sdkerrors.ErrInvalidAddress.Wrapf("address at index %d cannot be empty", i)
		}
		if _, err := sdk.AccAddressFromBech32(address); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid address at index %d: %s", i, err)
		}
	}

	return nil
}

// ValidateBasic validates the MsgUnregisterNFT message
func (m *MsgUnregisterNFT) ValidateBasic() error {
	// Validate authority
	if strings.TrimSpace(m.Authority) == "" {
		return sdkerrors.ErrInvalidAddress.Wrap("authority cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	// Validate key
	if m.Key == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("key cannot be nil")
	}
	if err := m.Key.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSigners returns the signers for the MsgRegisterNFT message
func (m *MsgRegisterNFT) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", err))
	}
	return []sdk.AccAddress{authority}
}

// GetSigners returns the signers for the MsgGrantRole message
func (m *MsgGrantRole) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", err))
	}
	return []sdk.AccAddress{authority}
}

// GetSigners returns the signers for the MsgRevokeRole message
func (m *MsgRevokeRole) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", err))
	}
	return []sdk.AccAddress{authority}
}

// GetSigners returns the signers for the MsgUnregisterNFT message
func (m *MsgUnregisterNFT) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", err))
	}
	return []sdk.AccAddress{authority}
}
