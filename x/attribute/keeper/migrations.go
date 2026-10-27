package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/attribute/types"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate2to3 migrates state from consensus version 2 to 3.
//
// This migration moves the attribute module to cosmossdk.io/collections AND changes the key
// layout: a name used to be stored as sha256(reversed name), it is now stored as the reversed
// name itself followed by a 0x00 terminator. That makes wildcard lookups (everything under
// "kyc.pb") a single prefix scan, but it means every key under the 0x02, 0x03 and 0x04 prefixes
// has to be rewritten.
//
// The attribute records themselves are unchanged - only the keys move. So this reads every
// legacy record, wipes the legacy keys, and writes each record back through the collections,
// rebuilding the name/address counters and the expiration index as it goes.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	logger := m.keeper.Logger(ctx)
	logger.Info("attribute: Migrate2to3 starting - rewriting keys from hashed to reversed names.")

	attrs, err := m.keeper.readV2Attributes(ctx)
	if err != nil {
		return fmt.Errorf("attribute: could not read v2 attributes: %w", err)
	}
	logger.Info(fmt.Sprintf("attribute: read %d attribute(s) from the legacy store.", len(attrs)))

	// Clear the legacy attribute (0x02), name/addr lookup (0x03) and expiration (0x04) entries.
	// The counters and expiration entries are rebuilt below rather than translated, so that a
	// counter that drifted in the old layout doesn't carry its drift across.
	if err = m.keeper.deleteV2Data(ctx); err != nil {
		return fmt.Errorf("attribute: could not delete v2 data: %w", err)
	}

	for i, attr := range attrs {
		addrBz := attr.GetAddressBytes()
		if err = m.keeper.attributes.Set(ctx, types.BuildAttrTriple(attr), attr); err != nil {
			return fmt.Errorf("attribute: could not write attribute %d (%q on %s): %w", i, attr.Name, attr.Address, err)
		}
		if err = m.keeper.incNameAddrCount(ctx, attr.Name, addrBz); err != nil {
			return fmt.Errorf("attribute: could not count attribute %d (%q on %s): %w", i, attr.Name, attr.Address, err)
		}
		if err = m.keeper.addExpireEntry(ctx, attr); err != nil {
			return fmt.Errorf("attribute: could not index expiration for attribute %d (%q on %s): %w", i, attr.Name, attr.Address, err)
		}
	}

	logger.Info(fmt.Sprintf("attribute: Migrate2to3 done - %d attribute(s) rewritten.", len(attrs)))
	return nil
}

// readV2Attributes reads every attribute record stored under the legacy 0x02 prefix.
//
// The legacy key is [0x02][len(addr)][addr][sha256(reversed name)][sha256(value)], which can't
// be reversed back into a name - but it doesn't need to be. The value is a marshaled Attribute
// that already carries Name and Address, so the new keys are built from the record itself and
// the legacy key is never parsed.
func (k Keeper) readV2Attributes(ctx sdk.Context) ([]types.Attribute, error) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.AttributeKeyPrefix, storetypes.PrefixEndBytes(types.AttributeKeyPrefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close() //nolint:errcheck // close error safe to ignore in this context.

	var attrs []types.Attribute
	for ; iter.Valid(); iter.Next() {
		var attr types.Attribute
		if err = k.cdc.Unmarshal(iter.Value(), &attr); err != nil {
			return nil, fmt.Errorf("could not unmarshal attribute at key %X: %w", iter.Key(), err)
		}
		attrs = append(attrs, attr)
	}
	return attrs, nil
}

// deleteV2Data removes every entry under the three legacy key prefixes that this migration
// rewrites: attributes (0x02), name/address counters (0x03) and expirations (0x04). The params
// entry (0x05) is not touched - its key doesn't contain a name.
func (k Keeper) deleteV2Data(ctx sdk.Context) error {
	for _, prefix := range [][]byte{
		types.AttributeKeyPrefix,
		types.AttributeAddrLookupKeyPrefix,
		types.AttributeExpirationKeyPrefix,
	} {
		if err := k.deletePrefix(ctx, prefix); err != nil {
			return fmt.Errorf("could not delete entries under prefix %X: %w", prefix, err)
		}
	}
	return nil
}

// deletePrefix deletes every entry under the provided prefix. The keys are collected first and
// deleted afterwards, since deleting through a live iterator is undefined behavior.
func (k Keeper) deletePrefix(ctx sdk.Context, prefix []byte) error {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return err
	}

	var keys [][]byte
	for ; iter.Valid(); iter.Next() {
		// The iterator reuses its key buffer, so this has to be a copy.
		keys = append(keys, append([]byte(nil), iter.Key()...))
	}
	if err = iter.Close(); err != nil {
		return err
	}

	for _, key := range keys {
		if err = store.Delete(key); err != nil {
			return err
		}
	}
	return nil
}

// Compile-time assurance the keeper exposes a Logger that returns log.Logger.
var _ interface {
	Logger(sdk.Context) log.Logger
} = Keeper{}
