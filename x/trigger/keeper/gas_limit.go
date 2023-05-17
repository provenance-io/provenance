package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/trigger/types"
)

const SetGasLimitCost = 2510

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
func (k Keeper) GetGasLimit(ctx sdk.Context, id types.TriggerID) (gasLimit uint64, err error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetGasLimitKey(id))
	if bz == nil {
		return 0, types.ErrGasLimitNotFound
	}

	gasLimit = types.GetGasLimitFromBytes(bz)
	return gasLimit, nil
}
