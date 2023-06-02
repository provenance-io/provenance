package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/trigger/types"
)

const SetGasLimitCost uint64 = 2510
const MaximumTriggerGas uint64 = 2000000

// SetGasLimit Sets a gas limit for a trigger
func (k Keeper) SetGasLimit(ctx sdk.Context, id types.TriggerID, gasLimit uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := types.GetGasLimitBytes(gasLimit)
	store.Set(types.GetGasLimitKey(id), bz)
}

// RemoveGasLimit Removes a gas limit for a trigger
func (k Keeper) RemoveGasLimit(ctx sdk.Context, id types.TriggerID) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetGasLimitKey(id)
	keyExists := store.Has(key)
	if keyExists {
		store.Delete(key)
	}
	return keyExists
}

// GetGasLimit Gets a gas limit by id
func (k Keeper) GetGasLimit(ctx sdk.Context, id types.TriggerID) (gasLimit uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetGasLimitKey(id))
	if bz == nil {
		panic("gas limit not found for trigger")
	}
	gasLimit = types.GetGasLimitFromBytes(bz)
	return
}

// IterateGasLimits Iterates through all the gas limits.
func (k Keeper) IterateGasLimits(ctx sdk.Context, handle func(gasLimit types.GasLimit) (stop bool, err error)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GasLimitKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		record := types.GetGasLimitFromBytes(iterator.Value())
		key := types.GetTriggerIDFromBytes(iterator.Key()[1:])
		stop, err := handle(types.GasLimit{TriggerId: key, Amount: record})
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}
	return nil
}

// GetAllGasLimits Gets all the gas limits within the store.
func (k Keeper) GetAllGasLimits(ctx sdk.Context) (gasLimits []types.GasLimit, err error) {
	err = k.IterateGasLimits(ctx, func(gasLimit types.GasLimit) (stop bool, err error) {
		gasLimits = append(gasLimits, gasLimit)
		return false, nil
	})
	return
}
