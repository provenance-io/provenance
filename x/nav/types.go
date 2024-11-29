package nav

import (
	"errors"
	"fmt"

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
func (n *NetAssetValue) AsRecord(height uint64, recordedBy string) *NetAssetValueRecord {
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
	if len(n.RecordedBy) > RecordedByMaxLen {
		return fmt.Errorf("invalid recorded_by %q: length %d exceeds max %d",
			n.RecordedBy[:7]+"..."+n.RecordedBy[:len(n.RecordedBy)-6],
			len(n.RecordedBy), RecordedByMaxLen)
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
