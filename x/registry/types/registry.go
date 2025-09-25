package types

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"

	metadataTypes "github.com/provenance-io/provenance/x/metadata/types"
)

const (
	registryKeyHrp = "reg"

	MaxLenAddress      = 90
	MaxLenAssetClassID = 128
	MaxLenNFTID        = 128
)

var alNumDashRx = regexp.MustCompile(`^[a-zA-Z0-9-.]+$`)

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
		return fmt.Errorf("registry key cannot be nil")
	}

	return errors.Join(
		ValidateNftID(m.NftId),
		ValidateClassID(m.AssetClassId),
	)
}

// Validate validates the RegistryEntry
func (m *RegistryEntry) Validate() error {
	if m == nil {
		return fmt.Errorf("registry entry cannot be nil")
	}

	var errs []error
	if err := m.Key.Validate(); err != nil {
		errs = append(errs, err)
	}

	// Validate roles
	if len(m.Roles) == 0 {
		errs = append(errs, fmt.Errorf("roles cannot be empty"))
	}

	for _, role := range m.Roles {
		if err := role.Validate(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// Validate validates the RolesEntry
func (m *RolesEntry) Validate() error {
	if m == nil {
		return fmt.Errorf("roles entry cannot be nil")
	}

	var errs []error
	if err := m.Role.Validate(); err != nil {
		errs = append(errs, err)
	}

	// Validate addresses
	if len(m.Addresses) == 0 {
		errs = append(errs, fmt.Errorf("addresses cannot be empty"))
	}

	// Check for duplicate addresses
	seen := make(map[string]bool)
	for _, address := range m.Addresses {
		if err := ValidateStringLength(address, 1, MaxLenAddress); err != nil {
			errs = append(errs, fmt.Errorf("address: %w", err))
		}
		if _, err := sdk.AccAddressFromBech32(address); err != nil {
			errs = append(errs, fmt.Errorf("address: %w", err))
		}
		if seen[address] {
			errs = append(errs, fmt.Errorf("duplicate address: %q", address))
		}
		seen[address] = true
	}

	return errors.Join(errs...)
}

func (m RegistryRole) Validate() error {
	var errs []error
	if _, ok := RegistryRole_name[int32(m)]; !ok {
		errs = append(errs, fmt.Errorf("invalid role"))
	}

	if m == RegistryRole_REGISTRY_ROLE_UNSPECIFIED {
		errs = append(errs, fmt.Errorf("cannot be unspecified"))
	}

	return errors.Join(errs...)
}

// ParseRegistryRole converts the provided string into a RegistryRole. The "REGISTRY_ROLE_" prefix is optional.
func ParseRegistryRole(str string) (RegistryRole, error) {
	name := strings.ToUpper(str)
	if !strings.HasPrefix(name, "REGISTRY_ROLE_") {
		name = "REGISTRY_ROLE_" + name
	}
	role, ok := RegistryRole_value[name]
	if !ok {
		return RegistryRole_REGISTRY_ROLE_UNSPECIFIED, NewErrCodeInvalidField("role", "invalid role")
	}
	rv := RegistryRole(role)
	if rv == RegistryRole_REGISTRY_ROLE_UNSPECIFIED {
		return rv, NewErrCodeInvalidField("role", "cannot be unspecified")
	}
	return rv, nil
}

// ValidRolesString returns a string containing all of the registry roles string values.
func ValidRolesString() string {
	roles := make([]string, 0, len(RegistryRole_name)-1)
	for _, roleID := range slices.Sorted(maps.Keys(RegistryRole_name)) {
		if roleID == 0 {
			continue
		}
		roles = append(roles, strings.TrimPrefix(RegistryRole_name[roleID], "REGISTRY_ROLE_"))
	}
	for _, role := range RegistryRole_value {
		roles = append(roles, RegistryRole(role).String())
	}
	return strings.Join(roles, "  ")
}

// Validate validates the RegistryBulkUpdate
func (m *MsgRegistryBulkUpdate) Validate() error {
	if m == nil {
		return fmt.Errorf("registry bulk update cannot be nil")
	}

	// Validate entries
	if len(m.Entries) == 0 || len(m.Entries) > 200 {
		return fmt.Errorf("entries cannot be empty or greater than 200")
	}

	for _, entry := range m.Entries {
		if err := entry.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateClassID validates the asset class id format
func ValidateClassID(classID string) error {
	if err := ValidateStringLength(classID, 1, MaxLenAssetClassID); err != nil {
		return err
	}

	if _, isScopeSpec := MetadataScopeSpecID(classID); isScopeSpec {
		// the class id is an asset class id
		if !alNumDashRx.MatchString(classID) {
			return fmt.Errorf("must only contain alphanumeric, '-', '.' characters")
		}
	}

	return nil
}

// ValidateNFTID validates the nft id format
func ValidateNftID(nftID string) error {
	if err := ValidateStringLength(nftID, 1, MaxLenNFTID); err != nil {
		return err
	}

	if _, isScope := MetadataScopeID(nftID); !isScope {
		// the nft id is an asset id
		if !alNumDashRx.MatchString(nftID) {
			return fmt.Errorf("must only contain alphanumeric, '-', '.' characters")
		}
	}

	return nil
}

// metadataScopeID returns the metadata address for a given bech32 string.
// The bool is true if it's for a scope, false if other or invalid.
func MetadataScopeID(bech32String string) (metadataTypes.MetadataAddress, bool) {
	addr, hrp, err := metadataTypes.ParseMetadataAddressFromBech32(bech32String)
	if err != nil {
		return nil, false
	}
	return addr, hrp == metadataTypes.PrefixScope
}

// metadataScopeSpecID returns the metadata address for a given bech32 string.
// The bool is true if it's for a scope spec, false if other or invalid.
func MetadataScopeSpecID(bech32String string) (metadataTypes.MetadataAddress, bool) {
	addr, hrp, err := metadataTypes.ParseMetadataAddressFromBech32(bech32String)
	if err != nil {
		return nil, false
	}
	return addr, hrp == metadataTypes.PrefixScopeSpecification
}
