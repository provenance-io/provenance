package keeper

import (
	"cosmossdk.io/collections"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/quarantine"
)

// This file is available only to unit tests and houses functions for doing
// things with private keeper package stuff.

var FundsHolderBalanceInvariantHelper = fundsHolderBalanceInvariantHelper

// WithFundsHolder creates a copy of this, setting the funds holder to the provided addr.
func (k Keeper) WithFundsHolder(addr sdk.AccAddress) Keeper {
	k.fundsHolder = addr
	return k
}

// WithBankKeeper creates a copy of this, setting the bank keeper to the provided one.
func (k Keeper) WithBankKeeper(bankKeeper quarantine.BankKeeper) Keeper {
	k.bankKeeper = bankKeeper
	return k
}

// GetCodec exposes this keeper's codec (cdc) for unit tests.
func (k Keeper) GetCodec() codec.BinaryCodec {
	return k.cdc
}

// GetStoreKey returns the keeper's schema for use in unit tests.
func (k Keeper) Getstorekey() collections.Schema {
	return k.GetSchema()
}

// BzToQuarantineRecord exposes bzToQuarantineRecord for unit tests.
func (k Keeper) BzToQuarantineRecord(bz []byte) (*quarantine.QuarantineRecord, error) {
	return k.bzToQuarantineRecord(bz)
}

// MustBzToQuarantineRecord exposes mustBzToQuarantineRecord for unit tests.
func (k Keeper) MustBzToQuarantineRecord(bz []byte) *quarantine.QuarantineRecord {
	return k.mustBzToQuarantineRecord(bz)
}

// SetQuarantineRecordSuffixIndex exposes setQuarantineRecordSuffixIndex for unit tests.
func (k Keeper) SetQuarantineRecordSuffixIndex(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress, value *quarantine.QuarantineRecordSuffixIndex) {
	k.setQuarantineRecordSuffixIndex(ctx, toAddr, fromAddr, value)
}

// BzToQuarantineRecordSuffixIndex exposes bzToQuarantineRecordSuffixIndex for unit tests.
func (k Keeper) BzToQuarantineRecordSuffixIndex(bz []byte) (*quarantine.QuarantineRecordSuffixIndex, error) {
	return k.bzToQuarantineRecordSuffixIndex(bz)
}

// MustBzToQuarantineRecordSuffixIndex exposes mustBzToQuarantineRecordSuffixIndex for unit tests.
func (k Keeper) MustBzToQuarantineRecordSuffixIndex(bz []byte) *quarantine.QuarantineRecordSuffixIndex {
	return k.mustBzToQuarantineRecordSuffixIndex(bz)
}

// GetQuarantineRecordSuffixIndex exposes getQuarantineRecordSuffixIndex for unit tests.
func (k Keeper) GetQuarantineRecordSuffixIndex(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress) *quarantine.QuarantineRecordSuffixIndex {
	return k.getQuarantineRecordSuffixIndex(ctx, toAddr, fromAddr)
}

// GetQuarantineRecordSuffixes exposes getQuarantineRecordSuffixes for unit tests.
func (k Keeper) GetQuarantineRecordSuffixes(ctx sdk.Context, toAddr sdk.AccAddress, fromAddrs []sdk.AccAddress) [][]byte {
	return k.getQuarantineRecordSuffixes(ctx, toAddr, fromAddrs)
}

// AddQuarantineRecordSuffixIndexes exposes addQuarantineRecordSuffixIndexes for unit tests.
func (k Keeper) AddQuarantineRecordSuffixIndexes(ctx sdk.Context, toAddr sdk.AccAddress, fromAddrs []sdk.AccAddress, suffix []byte) {
	k.addQuarantineRecordSuffixIndexes(ctx, toAddr, fromAddrs, suffix)
}

// DeleteQuarantineRecordSuffixIndexes exposes deleteQuarantineRecordSuffixIndexes for unit tests.
func (k Keeper) DeleteQuarantineRecordSuffixIndexes(ctx sdk.Context, toAddr sdk.AccAddress, fromAddrs []sdk.AccAddress, suffix []byte) {
	k.deleteQuarantineRecordSuffixIndexes(ctx, toAddr, fromAddrs, suffix)
}
