package types

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"

	"github.com/provenance-io/provenance/internal/provutils"
	metadataTypes "github.com/provenance-io/provenance/x/metadata/types"
)

const (
	registryKeyHrp = "reg"

	MaxLenAddress          = 90
	MaxLenAssetClassID     = 128
	MaxLenNFTID            = 128
	MaxRegistryBulkEntries = 250
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

	var errs []error
	if err := ValidateNftID(m.NftId); err != nil {
		errs = append(errs, fmt.Errorf("nft id: %w", err))
	}
	if err := ValidateClassID(m.AssetClassId); err != nil {
		errs = append(errs, fmt.Errorf("asset class id: %w", err))
	}

	return errors.Join(errs...)
}

// CollKey returns the collections key that this RegistryKey represents.
func (m *RegistryKey) CollKey() collections.Pair[string, string] {
	return collections.Join(m.AssetClassId, m.NftId)
}

// Equals returns true if this RegistryKey equals the provided one.
func (m *RegistryKey) Equals(other *RegistryKey) bool {
	if m == other {
		return true
	}
	if m == nil || other == nil {
		return false
	}
	return m.NftId == other.NftId && m.AssetClassId == other.AssetClassId
}

// Validate validates the RegistryEntry
func (m *RegistryEntry) Validate() error {
	if m == nil {
		return fmt.Errorf("registry entry cannot be nil")
	}

	var errs []error
	if err := m.Key.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("key: %w", err))
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

// GetRoleAddrs gets all the addresses assigned the given role in this RegistryEntry.
func (m *RegistryEntry) GetRoleAddrs(role RegistryRole) []string {
	if m == nil {
		return nil
	}
	for _, entry := range m.Roles {
		if entry.Role == role {
			return entry.Addresses
		}
	}
	return nil
}

// Validate validates the RolesEntry
func (m *RolesEntry) Validate() error {
	if m == nil {
		return fmt.Errorf("roles entry cannot be nil")
	}

	var errs []error
	if err := m.Role.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("role: %w", err))
	}

	// Validate addresses
	if len(m.Addresses) == 0 {
		errs = append(errs, fmt.Errorf("addresses cannot be empty"))
	}

	// Check for duplicate addresses
	seen := make(map[string]bool)
	for i, address := range m.Addresses {
		if seen[address] {
			errs = append(errs, fmt.Errorf("duplicate address[%d]: %q", i, address))
			continue
		}
		seen[address] = true
		if err := ValidateStringLength(address, 1, MaxLenAddress); err != nil {
			errs = append(errs, fmt.Errorf("address[%d]: %w", i, err))
			continue
		}
		if _, err := sdk.AccAddressFromBech32(address); err != nil {
			errs = append(errs, fmt.Errorf("address[%d]: %w", i, err))
			continue
		}
	}

	return errors.Join(errs...)
}

// UnmarshalJSON implements json.Unmarshaler for RegistryRole.
func (r *RegistryRole) UnmarshalJSON(data []byte) error {
	value, err := provutils.EnumUnmarshalJSON(data, RegistryRole_value, RegistryRole_name)
	if err != nil {
		return err
	}
	*r = RegistryRole(value)
	return nil
}

// Validate returns an error if this DayCountConvention isn't a defined enum entry, or is unspecified.
func (r RegistryRole) Validate() error {
	if err := provutils.EnumValidateExists(r, RegistryRole_name); err != nil {
		return err
	}

	if r == RegistryRole_REGISTRY_ROLE_UNSPECIFIED {
		return fmt.Errorf("cannot be unspecified")
	}

	return nil
}

// ShortString returns this.String() with the "REGISTRY_ROLE_" prefix removed.
func (r RegistryRole) ShortString() string {
	return strings.TrimPrefix(r.String(), "REGISTRY_ROLE_")
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

// ValidateClassID validates the asset class id format
func ValidateClassID(classID string) error {
	if err := ValidateStringLength(classID, 1, MaxLenAssetClassID); err != nil {
		return err
	}

	if _, isScopeSpec := MetadataScopeSpecID(classID); !isScopeSpec {
		// the class id is an asset class id
		if !alNumDashRx.MatchString(classID) {
			return fmt.Errorf("%q must only contain alphanumeric, '-', '.' characters", classID)
		}
	}

	return nil
}

// ValidateNftID validates the nft id format
func ValidateNftID(nftID string) error {
	if err := ValidateStringLength(nftID, 1, MaxLenNFTID); err != nil {
		return err
	}

	if _, isScope := MetadataScopeID(nftID); !isScope {
		// the nft id is an asset id
		if !alNumDashRx.MatchString(nftID) {
			return fmt.Errorf("%q must only contain alphanumeric, '-', '.' characters", nftID)
		}
	}

	return nil
}

// MetadataScopeID returns the metadata address for a given bech32 string.
// The bool is true if it's for a scope, false if other or invalid.
func MetadataScopeID(bech32String string) (metadataTypes.MetadataAddress, bool) {
	addr, hrp, err := metadataTypes.ParseMetadataAddressFromBech32(bech32String)
	if err != nil {
		return nil, false
	}
	return addr, hrp == metadataTypes.PrefixScope
}

// MetadataScopeSpecID returns the metadata address for a given bech32 string.
// The bool is true if it's for a scope spec, false if other or invalid.
func MetadataScopeSpecID(bech32String string) (metadataTypes.MetadataAddress, bool) {
	addr, hrp, err := metadataTypes.ParseMetadataAddressFromBech32(bech32String)
	if err != nil {
		return nil, false
	}
	return addr, hrp == metadataTypes.PrefixScopeSpecification
}
