package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

// Removes an account state
func (k Keeper) RemoveAccountState(ctx sdk.Context, rewardProgramID, epochID uint64, addr string) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetAccountStateKey(rewardProgramID, epochID, []byte(addr))
	if key == nil {
		return false
	}

	keyExists := store.Has(key)
	if keyExists {
		bz := store.Get(key)
		store.Delete(bz)
	}
	return keyExists
}

func (k Keeper) GetAccountState(ctx sdk.Context, rewardProgramID, epochID uint64, addr string) (state types.AccountState, err error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetAccountStateKey(rewardProgramID, epochID, []byte(addr))
	bz := store.Get(key)
	if len(bz) == 0 {
		return state, nil
	}
	err = k.cdc.Unmarshal(bz, &state)
	return state, err
}

func (k Keeper) SetAccountState(ctx sdk.Context, state *types.AccountState) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(state)
	key := types.GetAccountStateKey(state.GetRewardProgramId(), state.GetEpochId(), []byte(state.GetAddress()))
	store.Set(key, bz)
}

// Iterates over ALL the account states for a reward program's epoch
func (k Keeper) IterateAccountStates(ctx sdk.Context, rewardProgramID, epochID uint64, handle func(state types.AccountState) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetAccountStateKeyPrefix(rewardProgramID, epochID))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		record := types.AccountState{}
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return err
		}
		if handle(record) {
			break
		}
	}
	return nil
}
