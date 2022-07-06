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

func (k Keeper) SetRewardAccountState(ctx sdk.Context, state types.RewardAccountState) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&state)
	key := types.GetRewardAccountStateKey(state.GetRewardProgramId(), state.GetClaimPeriodId(), []byte(state.GetAddress()))
	store.Set(key, bz)
}

// IterateRewardAccountStates Iterates over the account states for a reward program's claim period
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

// Iterates over the account states for every reward program
func (k Keeper) IterateAllRewardAccountStates(ctx sdk.Context, handle func(state types.RewardAccountState) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetAllRewardAccountStateKeyPrefix())

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

// TODO test this
func (k Keeper) IterateRewardAccountStatesForRewardProgram(ctx sdk.Context, rewardProgramID uint64, handle func(state types.RewardAccountState) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetRewardProgramRewardAccountStateKeyPrefix(rewardProgramID))

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

// TODO Test this
func (k Keeper) GetRewardAccountStatesForClaimPeriod(ctx sdk.Context, rewardProgramID, claimPeriodID uint64) ([]types.RewardAccountState, error) {
	states := []types.RewardAccountState{}
	err := k.IterateRewardAccountStates(ctx, rewardProgramID, claimPeriodID, func(state types.RewardAccountState) (stop bool) {
		states = append(states, state)
		return false
	})
	return states, err
}

// TODO Test this
func (k Keeper) GetRewardAccountStatesForRewardProgram(ctx sdk.Context, rewardProgramID uint64) ([]types.RewardAccountState, error) {
	states := []types.RewardAccountState{}
	err := k.IterateRewardAccountStatesForRewardProgram(ctx, rewardProgramID, func(state types.RewardAccountState) (stop bool) {
		states = append(states, state)
		return false
	})
	return states, err
}

// TODO Test this
func (k Keeper) MakeRewardClaimsClaimableForPeriod(ctx sdk.Context, rewardProgramID, claimPeriodID uint64) error {
	states, err := k.GetRewardAccountStatesForClaimPeriod(ctx, rewardProgramID, claimPeriodID)
	for _, state := range states {
		state.ClaimStatus = types.RewardAccountState_CLAIMABLE
		k.SetRewardAccountState(ctx, state)
	}
	return err
}

// TODO Test this
func (k Keeper) ExpireRewardClaimsForRewardProgram(ctx sdk.Context, rewardProgramID uint64) error {
	states, err := k.GetRewardAccountStatesForRewardProgram(ctx, rewardProgramID)
	for _, state := range states {
		if state.ClaimStatus == types.RewardAccountState_CLAIMED {
			continue
		}
		state.ClaimStatus = types.RewardAccountState_EXPIRED
		k.SetRewardAccountState(ctx, state)
	}
	return err
}
