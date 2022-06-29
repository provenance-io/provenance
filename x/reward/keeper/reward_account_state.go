package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

// RemoveRewardAccountState Removes an account state
func (k Keeper) RemoveRewardAccountState(ctx sdk.Context, rewardProgramID, rewardClaimPeriodID uint64, addr string) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetRewardAccountStateKey(rewardProgramID, rewardClaimPeriodID, []byte(addr))
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

func (k Keeper) GetRewardAccountState(ctx sdk.Context, rewardProgramID, rewardClaimPeriodID uint64, addr string) (state types.RewardAccountState, err error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetRewardAccountStateKey(rewardProgramID, rewardClaimPeriodID, []byte(addr))
	bz := store.Get(key)
	if len(bz) == 0 {
		return state, nil
	}
	err = k.cdc.Unmarshal(bz, &state)
	return state, err
}

func (k Keeper) SetRewardAccountState(ctx sdk.Context, state *types.RewardAccountState) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(state)
	key := types.GetRewardAccountStateKey(state.GetRewardProgramId(), state.GetClaimPeriodId(), []byte(state.GetAddress()))
	store.Set(key, bz)
}

// IterateRewardAccountStates Iterates over ALL the account states for a reward program's claim period
func (k Keeper) IterateRewardAccountStates(ctx sdk.Context, rewardProgramID, rewardClaimPeriodID uint64, handle func(state types.RewardAccountState) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetRewardAccountStateKeyPrefix(rewardProgramID, rewardClaimPeriodID))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		record := types.RewardAccountState{}
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return err
		}
		if handle(record) {
			break
		}
	}
	return nil
}
