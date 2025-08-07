package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	registryKeyHrp = "reg"
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

	if err := m.Key.Validate(); err != nil {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid key: %s", err)
	}

	// Validate roles
	if len(m.Roles) == 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("roles cannot be empty")
	}

	// Check for duplicate roles
	for i, role := range m.Roles {
		if err := role.Validate(); err != nil {
			return sdkerrors.ErrInvalidRequest.Wrapf("invalid role at index %d: %s", i, err)
		}
	}

	return nil
}

// Validate validates the RolesEntry
func (m *RolesEntry) Validate() error {
	if m == nil {
		return sdkerrors.ErrInvalidRequest.Wrap("roles entry cannot be nil")
	}

	if err := m.Role.Validate(); err != nil {
		return err
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

func (m *RegistryRole) Validate() error {
	if _, ok := RegistryRole_value[m.String()]; !ok {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid role: %s", m.String())
	}

	// Validate role
	if *m == RegistryRole_REGISTRY_ROLE_UNSPECIFIED {
		return sdkerrors.ErrInvalidRequest.Wrap("role cannot be unspecified")
	}

	return nil
}

// Validate validates the GenesisState
func (m *GenesisState) Validate() error {
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

	for _, entry := range m.Entries {
		if err := entry.Validate(); err != nil {
			return sdkerrors.ErrInvalidRequest.Wrapf("invalid entry: %s", err)
		}
	}

	return nil
}

// Combine the asset class id and nft id into a bech32 string.
// Using bech32 here just allows us a readable identifier for the registry.
func (key RegistryKey) String() string {
	joined := strings.Join([]string{key.AssetClassId, key.NftId}, ":")

	b32, err := bech32.ConvertAndEncode(registryKeyHrp, []byte(joined))
	if err != nil {
		panic(err)
	}

	return b32
}

func StringToRegistryKey(s string) (*RegistryKey, error) {
	hrp, b, err := bech32.DecodeAndConvert(s)
	if err != nil {
		return nil, err
	}

	if hrp != registryKeyHrp {
		return nil, NewErrCodeInvalidHrp(hrp)
	}

	parts := strings.Split(string(b), ":")
	if len(parts) != 2 {
		return nil, NewErrCodeInvalidKey(s)
	}

	return &RegistryKey{
		AssetClassId: parts[0],
		NftId:        parts[1],
	}, nil
}
