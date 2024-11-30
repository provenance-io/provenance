package nav

import (
	"errors"
	"fmt"

	"cosmossdk.io/collections"

	"github.com/provenance-io/provenance/internal/provutils"
)

// String returns a string representation of this nav.
func (n *NetAssetValue) String() string {
	if n == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s=%s", n.Assets, n.Price)
}

// Validate returns an error if something about this nav is wrong.
func (n *NetAssetValue) Validate() error {
	if n == nil {
		return errors.New("nav cannot be nil")
	}
	if err := n.Assets.Validate(); err != nil {
		return fmt.Errorf("invalid assets %q: %w", n.Assets, err)
	}
	if n.Assets.IsZero() {
		return fmt.Errorf("invalid assets %q: cannot be zero", n.Assets)
	}
	if err := n.Price.Validate(); err != nil {
		return fmt.Errorf("invalid price %q: %w", n.Price, err)
	}
	return nil
}

// AsRecord returns a NetAssetValueRecord for this NetAssetValue including the provided info.
func (n *NetAssetValue) AsRecord(height int64, recordedBy string) *NetAssetValueRecord {
	return &NetAssetValueRecord{
		Assets:       n.Assets,
		Price:        n.Price,
		RecordHeight: height,
		RecordedBy:   recordedBy,
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
	return provutils.ValidateSlice(n, (*NetAssetValue).Validate)
}

// ValidateNAVs is the same as navs.Validate().
func ValidateNAVs(navs NAVs) error {
	return navs.Validate()
}

// AsRecords converts each of the provided navs into a NetAssetValueRecord with the provided info.
func (n NAVs) AsRecords(height int64, recordedBy string) NAVRecords {
	if n == nil {
		return nil
	}
	rv := make(NAVRecords, len(n))
	for i, entry := range n {
		rv[i] = entry.AsRecord(height, recordedBy)
	}
	return rv
}

// NAVsAsRecords converts each of the provided navs into a NetAssetValueRecord with the provided info.
func NAVsAsRecords(navs NAVs, height int64, recordedBy string) NAVRecords {
	return navs.AsRecords(height, recordedBy)
}

// String returns a string representation of this nav record.
func (n *NetAssetValueRecord) String() string {
	if n == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s=%s@%d by %q", n.Assets, n.Price, n.RecordHeight, n.RecordedBy)
}

// Validate returns an error if something about this nav record is wrong.
func (n *NetAssetValueRecord) Validate() error {
	if n == nil {
		return errors.New("nav record cannot be nil")
	}
	if err := n.Assets.Validate(); err != nil {
		return fmt.Errorf("invalid assets %q: %w", n.Assets, err)
	}
	if n.Assets.IsZero() {
		return fmt.Errorf("invalid assets %q: cannot be zero", n.Assets)
	}
	if err := n.Price.Validate(); err != nil {
		return fmt.Errorf("invalid price %q: %w", n.Price, err)
	}
	if err := ValidateRecordedBy(n.RecordedBy); err != nil {
		return err
	}
	return nil
}

// ValidateRecordedBy returns an error if the provided string cannot be used as a RecordedBy string.
func ValidateRecordedBy(recordedBy string) error {
	if len(recordedBy) == 0 {
		return fmt.Errorf("invalid recorded_by %q: cannot be empty", recordedBy)
	}
	if len(recordedBy) > RecordedByMaxLen {
		return fmt.Errorf("invalid recorded_by %q: length %d exceeds max %d",
			recordedBy[:7]+"..."+recordedBy[:len(recordedBy)-6],
			len(recordedBy), RecordedByMaxLen)
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
	return provutils.ValidateSlice(n, (*NetAssetValueRecord).Validate)
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
	if err := NAVRecords(g.Navs).Validate(); err != nil {
		return fmt.Errorf("invalid navs: %w", err)
	}
	return nil
}
