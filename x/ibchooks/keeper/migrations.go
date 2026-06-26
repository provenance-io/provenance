package keeper

import (
	"strconv"
	"strings"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/ibchooks/types"
)

// Migrator handles in-place store migrations for the x/ibchooks module.
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns a Migrator for the x/ibchooks module.
func NewMigrator(k Keeper) Migrator {
	return Migrator{keeper: k}
}

// Migrate1to2 re-keys legacy raw-string packet entries into the new collections:
//   - "channel::seq"      -> PacketCallbacks[(channel, seq)]   (prefix 0x02)
//   - "channel::seq::ack" -> PacketAckActors[(channel, seq)]   (prefix 0x03)
//
// Params (0x01) is already byte-identical under the new collections.Item and needs no migration.
// Legacy entries are ephemeral, so there are usually few (only in-flight packets) at upgrade time.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	store := m.keeper.storeService.OpenKVStore(ctx)
	it, err := store.Iterator(nil, nil)
	if err != nil {
		return err
	}

	type entry struct{ key, val []byte }
	var legacy []entry
	for ; it.Valid(); it.Next() {
		key := it.Key()
		if len(key) == 1 && key[0] == types.ParamsKeyBz[0] {
			continue
		}
		if len(key) > 0 && (key[0] == types.PacketCallbackKeyBz[0] || key[0] == types.PacketAckKeyBz[0]) {
			continue
		}
		legacy = append(legacy,
			entry{key: append([]byte(nil), key...), val: append([]byte(nil), it.Value()...)})
	}
	if cerr := it.Close(); cerr != nil {
		return cerr
	}

	for _, e := range legacy {
		parts := strings.Split(string(e.key), "::")
		switch {
		case len(parts) == 2:
			seq, perr := strconv.ParseUint(parts[1], 10, 64)
			if perr != nil {
				continue
			}
			if err := m.keeper.packetCallbacks.Set(ctx, collections.Join(parts[0], seq), string(e.val)); err != nil {
				return err
			}
		case len(parts) == 3 && parts[2] == "ack":
			seq, perr := strconv.ParseUint(parts[1], 10, 64)
			if perr != nil {
				continue
			}
			if err := m.keeper.packetAckActors.Set(ctx, collections.Join(parts[0], seq), e.val); err != nil {
				return err
			}
		default:
			continue
		}
		if err := store.Delete(e.key); err != nil {
			return err
		}
	}

	ctx.Logger().Info("ibchooks 1->2 migration complete: re-keyed legacy packet entries", "count", len(legacy))
	return nil
}
