package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Validate validates the RegistryKey
func (m *RegistryKey) Validate() error {
	if m == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("registry key cannot be nil")
	}

	// Validate NFT ID
	if strings.TrimSpace(m.NftId) == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("nft_id cannot be empty")
	}

	// Validate Asset Class ID
	if strings.TrimSpace(m.AssetClassId) == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("asset_class_id cannot be empty")
	}

	return nil
}

// Validate validates the RegistryEntry
func (m *RegistryEntry) Validate() error {
	if m == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("registry entry cannot be nil")
	}

	// Validate key
	if m.Key == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("key cannot be nil")
	}
	if err := m.Key.Validate(); err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid key: %s", err)
	}

	// Validate roles
	if len(m.Roles) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("roles cannot be empty")
	}

	// Check for duplicate roles
	seenRoles := make(map[RegistryRole]bool)
	for i, role := range m.Roles {
		if err := role.Validate(); err != nil {
			return sdkerrors.ErrInvalidRequest.Wrapf("invalid role at index %d: %s", i, err)
		}
		if seenRoles[role.Role] {
			return sdkerrors.ErrInvalidRequest.Wrapf("duplicate role at index %d: %s", i, role.Role)
		}
		seenRoles[role.Role] = true
	}

	return nil
}

// Validate validates the RolesEntry
func (m *RolesEntry) Validate() error {
	if m == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("roles entry cannot be nil")
	}

	// Validate role
	if m.Role == RegistryRole_REGISTRY_ROLE_UNSPECIFIED {
		return sdkerrors.ErrInvalidRequest.Wrap("role cannot be unspecified")
	}

	// Validate addresses
	if len(m.Addresses) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("addresses cannot be empty")
	}

	// Check for duplicate addresses
	seen := make(map[string]bool)
	for i, address := range m.Addresses {
		if strings.TrimSpace(address) == "" {
			return sdkerrors.ErrInvalidAddress.Wrapf("address at index %d cannot be empty", i)
		}
		if _, err := sdk.AccAddressFromBech32(address); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid address at index %d: %s", i, err)
		}
		if seen[address] {
			return sdkerrors.ErrInvalidRequest.Wrapf("duplicate address at index %d: %s", i, address)
		}
		seen[address] = true
	}

	return nil
}

// Validate validates the GenesisState
func (m *GenesisState) Validate() error {
	if m == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("genesis state cannot be nil")
	}

	// Validate entries
	for i, entry := range m.Entries {
		if err := entry.Validate(); err != nil {
			return sdkerrors.ErrInvalidRequest.Wrapf("invalid entry at index %d: %s", i, err)
		}
	}

	return nil
}

// Validate validates the RegistryBulkUpdate
func (m *RegistryBulkUpdate) Validate() error {
	if m == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("registry bulk update cannot be nil")
	}

	// Validate entries
	if len(m.Entries) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("entries cannot be empty")
	}

	for i, entry := range m.Entries {
		if err := entry.Validate(); err != nil {
			return sdkerrors.ErrInvalidRequest.Wrapf("invalid entry at index %d: %s", i, err)
		}
	}

	return nil
}

// Validate validates the RegistryBulkUpdateEntry
func (m *RegistryBulkUpdateEntry) Validate() error {
	if m == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("registry bulk update entry cannot be nil")
	}

	// Validate key
	if m.Key == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("key cannot be nil")
	}
	if err := m.Key.Validate(); err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid key: %s", err)
	}

	// Validate roles
	if len(m.Roles) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("roles cannot be empty")
	}

	// Check for duplicate roles
	seenRoles := make(map[RegistryRole]bool)
	for i, role := range m.Roles {
		if role == RegistryRole_REGISTRY_ROLE_UNSPECIFIED {
			return sdkerrors.ErrInvalidRequest.Wrapf("role at index %d cannot be unspecified", i)
		}
		if seenRoles[role] {
			return sdkerrors.ErrInvalidRequest.Wrapf("duplicate role at index %d: %s", i, role)
		}
		seenRoles[role] = true
	}

	return nil
}
