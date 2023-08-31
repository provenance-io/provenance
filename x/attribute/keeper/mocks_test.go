package keeper_test

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/attribute/types"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

// mockNameKeeper is a not-really-mocked name keeper. It allows injection of errors
// for a few endpoints, but uses an actual name keeper for normal functionality.
type mockNameKeeper struct {
	Parent                types.NameKeeper
	GetRecordByNameError  string
	SetNameRecordError    string
	UpdateNameRecordError string
}

var _ types.NameKeeper = (*mockNameKeeper)(nil)

// newMockNameKeeper creates a "mocked" name keeper backed by the provided name keeper.
func newMockNameKeeper(parent types.NameKeeper) *mockNameKeeper {
	return &mockNameKeeper{Parent: parent}
}

// WithGetRecordByNameError sets error that should be returned from GetRecordByName.
func (k *mockNameKeeper) WithGetRecordByNameError(err string) *mockNameKeeper {
	k.GetRecordByNameError = err
	return k
}

// WithSetNameRecordError sets error that should be returned from SetNameRecord.
func (k *mockNameKeeper) WithSetNameRecordError(err string) *mockNameKeeper {
	k.SetNameRecordError = err
	return k
}

// WithUpdateNameRecordError sets error that should be returned from UpdateNameRecord.
func (k *mockNameKeeper) WithUpdateNameRecordError(err string) *mockNameKeeper {
	k.UpdateNameRecordError = err
	return k
}

// ResolvesTo calls the parent's ResolvesTo function.
func (k *mockNameKeeper) ResolvesTo(ctx sdk.Context, name string, addr sdk.AccAddress) bool {
	return k.Parent.ResolvesTo(ctx, name, addr)
}

// Normalize calls the parent's Normalize function.
func (k *mockNameKeeper) Normalize(ctx sdk.Context, name string) (string, error) {
	return k.Parent.Normalize(ctx, name)
}

// GetRecordByName returns an error if desired, otherwise calls parent's GetRecordByName function.
func (k *mockNameKeeper) GetRecordByName(ctx sdk.Context, name string) (record *nametypes.NameRecord, err error) {
	if len(k.GetRecordByNameError) > 0 {
		return nil, errors.New(k.GetRecordByNameError)
	}
	return k.Parent.GetRecordByName(ctx, name)
}

// NameExists calls the parent's NameExists function.
func (k *mockNameKeeper) NameExists(ctx sdk.Context, name string) bool {
	return k.Parent.NameExists(ctx, name)
}

// SetAttributeKeeper calls the parent's SetAttributeKeeper function.
func (k *mockNameKeeper) SetAttributeKeeper(attrKeeper nametypes.AttributeKeeper) {
	k.Parent.SetAttributeKeeper(attrKeeper)
}

// SetNameRecord returns an error if desired, otherwise calls parent's SetNameRecord function.
func (k *mockNameKeeper) SetNameRecord(ctx sdk.Context, name string, addr sdk.AccAddress, restrict bool) error {
	if len(k.SetNameRecordError) > 0 {
		return errors.New(k.SetNameRecordError)
	}
	return k.Parent.SetNameRecord(ctx, name, addr, restrict)
}

// UpdateNameRecord returns an error if desired, otherwise calls parent's UpdateNameRecord function.
func (k *mockNameKeeper) UpdateNameRecord(ctx sdk.Context, name string, addr sdk.AccAddress, restrict bool) error {
	if len(k.UpdateNameRecordError) > 0 {
		return errors.New(k.UpdateNameRecordError)
	}
	return k.Parent.UpdateNameRecord(ctx, name, addr, restrict)
}

// IterateRecords calls the parent's IterateRecords function.
func (k *mockNameKeeper) IterateRecords(ctx sdk.Context, prefix []byte, handle func(nametypes.NameRecord) error) error {
	return k.Parent.IterateRecords(ctx, prefix, handle)
}
