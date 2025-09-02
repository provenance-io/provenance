package types

import (
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

const (
	registryKeyHrp = "reg"
)

// Combine the asset class id and nft id into a bech32 string.
// Using bech32 here just allows us a readable identifier for the ledger.
func (m RegistryKey) String() string {
	// Use null byte as delimiter
	joined := m.AssetClassId + "\x00" + m.NftId

	b32, err := bech32.ConvertAndEncode(registryKeyHrp, []byte(joined))
	if err != nil {
		panic(err)
	}

	return b32
}

// Validate validates the RegistryKey
func (m *RegistryKey) Validate() error {
	if m == nil {
		return NewErrCodeInvalidField("registry key", "registry key cannot be nil")
	}

	if err := lenCheck("nft_id", m.NftId, 1, 128); err != nil {
		return err
	}

	if err := lenCheck("asset_class_id", m.AssetClassId, 1, 128); err != nil {
		return err
	}

	// Verify that the nft_id and asset_class_id do not contain a null byte
	if strings.Contains(m.NftId, "\x00") {
		return NewErrCodeInvalidField("nft_id", "must not contain a null byte")
	}

	if strings.Contains(m.AssetClassId, "\x00") {
		return NewErrCodeInvalidField("asset_class_id", "must not contain a null byte")
	}

	return nil
}

// Validate validates the RegistryEntry
func (m *RegistryEntry) Validate() error {
	if m == nil {
		return NewErrCodeInvalidField("registry entry", "registry entry cannot be nil")
	}

	if err := m.Key.Validate(); err != nil {
		return NewErrCodeInvalidField("key", err.Error())
	}

	// Validate roles
	if len(m.Roles) == 0 {
		return NewErrCodeInvalidField("roles", "roles cannot be empty")
	}

	for _, role := range m.Roles {
		if err := role.Validate(); err != nil {
			return NewErrCodeInvalidField("role", err.Error())
		}
	}

	return nil
}

// Validate validates the RolesEntry
func (m *RolesEntry) Validate() error {
	if m == nil {
		return NewErrCodeInvalidField("roles entry", "roles entry cannot be nil")
	}

	if err := m.Role.Validate(); err != nil {
		return err
	}

	// Validate addresses
	if len(m.Addresses) == 0 {
		return NewErrCodeInvalidField("addresses", "addresses cannot be empty")
	}

	// Check for duplicate addresses
	seen := make(map[string]bool)
	for _, address := range m.Addresses {
		if strings.TrimSpace(address) == "" {
			return NewErrCodeInvalidField("address", "address cannot be empty")
		}
		if _, err := sdk.AccAddressFromBech32(address); err != nil {
			return NewErrCodeInvalidField("address", address)
		}
		if seen[address] {
			return NewErrCodeInvalidField("address", address)
		}
		seen[address] = true
	}

	return nil
}

func (m *RegistryRole) Validate() error {
	if _, ok := RegistryRole_value[m.String()]; !ok {
		return NewErrCodeInvalidField("role", m.String())
	}

	// Validate role
	if *m == RegistryRole_REGISTRY_ROLE_UNSPECIFIED {
		return NewErrCodeInvalidField("role", "role cannot be unspecified")
	}

	return nil
}

// Validate validates the GenesisState
func (m *GenesisState) Validate() error {
	// Validate entries
	for _, entry := range m.Entries {
		if err := entry.Validate(); err != nil {
			return NewErrCodeInvalidField("entry", err.Error())
		}
	}

	return nil
}

// Validate validates the RegistryBulkUpdate
func (m *RegistryBulkUpdate) Validate() error {
	// Validate entries
	if len(m.Entries) == 0 {
		return NewErrCodeInvalidField("entries", "entries cannot be empty")
	}

	for _, entry := range m.Entries {
		if err := entry.Validate(); err != nil {
			return NewErrCodeInvalidField("entry", err.Error())
		}
	}

	return nil
}

// Validate validates the RegistryBulkUpdateEntry
func (m *RegistryBulkUpdateEntry) Validate() error {
	if m == nil {
		return NewErrCodeInvalidField("registry bulk update entry", "registry bulk update entry cannot be nil")
	}

	// Validate key
	if m.Key == nil {
		return NewErrCodeInvalidField("key", "key cannot be nil")
	}
	if err := m.Key.Validate(); err != nil {
		return NewErrCodeInvalidField("key", err.Error())
	}

	for _, entry := range m.Entries {
		if err := entry.Validate(); err != nil {
			return NewErrCodeInvalidField("entry", err.Error())
		}
	}

	return nil
}

// lenCheck checks if the string is nil or empty and if it is, returns a missing field error.
// It also checks if the string is less than the minimum length or greater than the maximum length and returns an invalid field error.
func lenCheck(field string, str string, minLength int, maxLength int) error {
	// empty string
	if minLength > 0 && str == "" {
		return NewErrCodeInvalidField(field, "value cannot be empty")
	}

	if len(str) < minLength {
		return NewErrCodeInvalidField(field, "must be greater than or equal to "+strconv.Itoa(minLength)+" characters")
	}

	if len(str) > maxLength {
		return NewErrCodeInvalidField(field, "must be less than or equal to "+strconv.Itoa(maxLength)+" characters")
	}

	return nil
}
