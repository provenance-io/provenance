package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Migrate1To2 is a migration method that will delete the gas limits from state.
func (m Migrator) Migrate1To2(ctx sdk.Context) error {
	logger := m.keeper.Logger(ctx)
	logger.Info("Starting migration of x/trigger from 1 to 2.")
	m.DeleteGasLimits(ctx, logger)
	logger.Info("Done migrating x/trigger from 1 to 2.")
	return nil
}

// GasLimitKeyPrefix is an initial byte to help group all gas limit keys.
var GasLimitKeyPrefix = []byte{0x04}

// DeleteGasLimits identifies all the gas limit stored in state, then deletes them.
func (m Migrator) DeleteGasLimits(ctx sdk.Context, logger log.Logger) {
	logger.Info("Identifying trigger gas limits to delete")
	store := ctx.KVStore(m.keeper.storeKey)

	// We need to make sure the iterator is closed before deleting stuff,
	// but we also need to make sure it closes if things go weird.
	iter := storetypes.KVStorePrefixIterator(store, GasLimitKeyPrefix)
	closeIter := func() {
		if iter != nil {
			iter.Close()
			iter = nil
		}
	}
	defer closeIter()

	var toDelete [][]byte
	for ; iter.Valid(); iter.Next() {
		toDelete = append(toDelete, iter.Key())
	}
	closeIter()

	if len(toDelete) == 0 {
		logger.Info("No trigger gas limits to delete")
		return
	}

	logger.Info(fmt.Sprintf("Deleting %d trigger gas limits.", len(toDelete)))
	for _, key := range toDelete {
		store.Delete(key)
	}
	logger.Info(fmt.Sprintf("Done deleting %d trigger gas limits.", len(toDelete)))
}
