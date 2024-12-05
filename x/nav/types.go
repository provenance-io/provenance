package nav

import (
	"errors"
	"fmt"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/provutils"
)

// baseNAV is an interface that has accessors for all NAV types.
type baseNAV interface {
	GetAssets() sdk.Coin
	GetPrice() sdk.Coin
	String() string
}

var (
	_ baseNAV = (*NetAssetValue)(nil)
	_ baseNAV = (*NetAssetValueRecord)(nil)
)

// validateBaseNAV returns an error if there's something wrong with the provided NAV.
func validateBaseNAV(nav baseNAV) error {
	if nav == nil {
		return errors.New("nav cannot be nil")
	}

	assets := nav.GetAssets()
	if err := assets.Validate(); err != nil {
		return fmt.Errorf("invalid assets %q: %w", assets, err)
	}
	if assets.IsZero() {
		return fmt.Errorf("invalid assets %q: cannot be zero", assets)
	}

	price := nav.GetPrice()
	if err := price.Validate(); err != nil {
		return fmt.Errorf("invalid price %q: %w", price, err)
	}
	if assets.Denom == price.Denom {
		return fmt.Errorf("nav assets %q and price %q must have different denoms", assets.String(), price.String())
	}
	return nil
}

// validateNoDups returns an error if two (or more) of the provided navs have the same assets and price denoms.
func validateNoDups[S ~[]E, E baseNAV](navs S) error {
	all := make(map[string]S)
	keys := make([]string, 0, len(navs))
	for _, nav := range navs {
		key := nav.GetAssets().Denom + " " + nav.GetPrice().Denom
		if _, have := all[key]; !have {
			keys = append(keys, key)
		}
		all[key] = append(all[key], nav)
	}

	var errs []error
	for _, key := range keys {
		if len(all[key]) > 1 {
			errs = append(errs, fmt.Errorf("cannot have multiple (%d) navs with the same asset (%q) and price (%q)",
				len(all[key]), all[key][0].GetAssets().Denom, all[key][0].GetPrice().Denom))
		}
	}
	return errors.Join(errs...)
}

// String returns a string representation of this nav.
func (n *NetAssetValue) String() string {
	if n == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s=%s", n.Assets, n.Price)
}

// Validate returns an error if something about this nav is wrong.
func (n *NetAssetValue) Validate() error {
	return validateBaseNAV(n)
}

// AsRecord returns a NetAssetValueRecord for this NetAssetValue including the provided info.
func (n *NetAssetValue) AsRecord(height uint64, source string) *NetAssetValueRecord {
	return &NetAssetValueRecord{
		Assets: n.Assets,
		Price:  n.Price,
		Height: height,
		Source: source,
	}
}

// NAVs is a slice of NetAssetValue entries.
type NAVs []*NetAssetValue

// String returns a string representation of these navs.
func (n NAVs) String() string {
	return provutils.SliceString(n)
}

// Validate returns an error if there's something wrong with any of these navs.
func (n NAVs) Validate() error {
	if err := provutils.ValidateSlice(n, (*NetAssetValue).Validate); err != nil {
		return err
	}
	return validateNoDups(n)
}

// ValidateNAVs is the same as navs.Validate().
func ValidateNAVs(navs NAVs) error {
	return navs.Validate()
}

// AsRecords converts each of the provided navs into a NetAssetValueRecord with the provided info.
func (n NAVs) AsRecords(height uint64, source string) NAVRecords {
	if n == nil {
		return nil
	}
	rv := make(NAVRecords, len(n))
	for i, entry := range n {
		rv[i] = entry.AsRecord(height, source)
	}
	return rv
}

// NAVsAsRecords converts each of the provided navs into a NetAssetValueRecord with the provided info.
func NAVsAsRecords(navs NAVs, height uint64, source string) NAVRecords {
	return navs.AsRecords(height, source)
}

// String returns a string representation of this nav record.
func (n *NetAssetValueRecord) String() string {
	if n == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s=%s@%d by %q", n.Assets, n.Price, n.Height, n.Source)
}

// Validate returns an error if something about this nav record is wrong.
func (n *NetAssetValueRecord) Validate() error {
	if err := validateBaseNAV(n); err != nil {
		return err
	}
	if err := ValidateSource(n.Source); err != nil {
		return err
	}
	return nil
}

// ValidateSource returns an error if the provided string cannot be used as a Source string.
func ValidateSource(source string) error {
	if len(source) == 0 {
		return fmt.Errorf("invalid source %q: cannot be empty", source)
	}
	if len(source) > SourceMaxLen {
		return fmt.Errorf("invalid source %q: length %d exceeds max %d",
			source[:7]+"..."+source[:len(source)-6],
			len(source), SourceMaxLen)
	}
	return nil
}

// AsNAV returns a NetAssetValue representation of this NetAssetValueRecord.
func (n *NetAssetValueRecord) AsNAV() *NetAssetValue {
	return &NetAssetValue{
		Assets: n.Assets,
		Price:  n.Price,
	}
}

// Key returns the state store key for this NAV.
func (n *NetAssetValueRecord) Key() collections.Pair[string, string] {
	return collections.Join(n.Assets.Denom, n.Price.Denom)
}

// NAVRecords is a slice of NetAssetValueRecord entries.
type NAVRecords []*NetAssetValueRecord

// String returns a string representation of these nav records.
func (n NAVRecords) String() string {
	return provutils.SliceString(n)
}

// Validate returns an error if there's something wrong with any of these nav records.
func (n NAVRecords) Validate() error {
	if err := provutils.ValidateSlice(n, (*NetAssetValueRecord).Validate); err != nil {
		return err
	}
	return validateNoDups(n)
}

// ValidateNAVRecords is the same as navs.Validate().
func ValidateNAVRecords(navs NAVRecords) error {
	return navs.Validate()
}

// DefaultGenesisState returns the default NAV module genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}

// Validate returns an error if anything is wrong with this GenesisState.
func (g GenesisState) Validate() error {
	if err := ValidateNAVRecords(g.Navs); err != nil {
		return fmt.Errorf("invalid navs: %w", err)
	}
	return nil
}

// NewEventSetNetAssetValue creates a new EventSetNetAssetValue for the provided nav record.
func NewEventSetNetAssetValue(nav *NetAssetValueRecord) *EventSetNetAssetValue {
	return &EventSetNetAssetValue{
		Assets: nav.Assets.String(),
		Price:  nav.Price.String(),
		Source: nav.Source,
	}
}
