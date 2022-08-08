package keeper

import (
	"encoding/binary"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

func (k Keeper) GetRewardAccountState(ctx sdk.Context, rewardProgramID, rewardClaimPeriodID uint64, addr string) (state types.RewardAccountState, err error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetRewardAccountStateKey(rewardProgramID, rewardClaimPeriodID, types.MustAccAddressFromBech32(addr))
	bz := store.Get(key)
	if len(bz) == 0 {
		return state, nil
	}
	err = k.cdc.Unmarshal(bz, &state)

	// By default protobuf can't tell the difference between
	// an empty map and nil. It's solution is to always make
	// it nil because of the spec.
	if state.ActionCounter == nil {
		state.ActionCounter = make(map[string]uint64)
	}

	return state, err
}

func (k Keeper) SetRewardAccountState(ctx sdk.Context, state types.RewardAccountState) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&state)
	key := types.GetRewardAccountStateKey(state.GetRewardProgramId(), state.GetClaimPeriodId(), types.MustAccAddressFromBech32(state.GetAddress()))
	store.Set(key, bz)
	// since there is a significant use case of looking up this via address create a secondary index
	// [0x8] :: [addr-bytes::reward program id bytes]::[claim period id bytes] {}
	addressLookupKey := types.GetRewardAccountStateAddressLookupKey(types.MustAccAddressFromBech32(state.GetAddress()), state.GetRewardProgramId(), state.GetClaimPeriodId())
	// no need for a value a key can derive all the info needed
	store.Set(addressLookupKey, []byte{})
}

// IterateRewardAccountStates Iterates over the account states for a reward program's claim period
func (k Keeper) IterateRewardAccountStates(ctx sdk.Context, rewardProgramID, rewardClaimPeriodID uint64, handle func(state types.RewardAccountState) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetRewardAccountStateClaimPeriodKey(rewardProgramID, rewardClaimPeriodID))

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

// IterateRewardAccountStatesByAddress Iterates over the account states by address iterator
func (k Keeper) IterateRewardAccountStatesByAddress(ctx sdk.Context, addr sdk.AccAddress, handle func(state types.RewardAccountState) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetAllRewardAccountByAddressPartialKey(types.MustAccAddressFromBech32(addr.String())))
	return k.IterateRewardAccountStatesByLookUpIndex(ctx, addr, iterator, handle)
}

// IterateRewardAccountStatesByAddressAndRewardsID Iterates over the account states by address iterator and reward id
func (k Keeper) IterateRewardAccountStatesByAddressAndRewardsID(ctx sdk.Context, addr sdk.AccAddress, rewardsID uint64, handle func(state types.RewardAccountState) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetAllRewardAccountByAddressAndRewardsIDPartialKey(addr, rewardsID))
	return k.IterateRewardAccountStatesByLookUpIndex(ctx, addr, iterator, handle)
}

// IterateRewardAccountStatesByLookUpIndex iterates reward account states by secondary index // [0x8] :: [addr-bytes::reward program id bytes]::[claim period id bytes] {}
func (k Keeper) IterateRewardAccountStatesByLookUpIndex(ctx sdk.Context, addr sdk.AccAddress, iterator sdk.Iterator, handle func(state types.RewardAccountState) (stop bool)) error {
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		keyParsed, err := ParseRewardAccountLookUpKey(iterator.Key(), addr)
		if err != nil {
			return err
		}
		record, err := k.GetRewardAccountState(ctx, keyParsed.rewardID, keyParsed.claimID, addr.String())
		if err != nil {
			return err
		}
		if handle(record) {
			break
		}
	}
	return nil
}

// IterateAllRewardAccountStates Iterates over the account states for every reward program
func (k Keeper) IterateAllRewardAccountStates(ctx sdk.Context, handle func(state types.RewardAccountState) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetAllRewardAccountStateKey())

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

// IterateRewardAccountStatesForRewardProgram Iterates over the account states for a reward program
func (k Keeper) IterateRewardAccountStatesForRewardProgram(ctx sdk.Context, rewardProgramID uint64, handle func(state types.RewardAccountState) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetRewardProgramRewardAccountStateKey(rewardProgramID))

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

// Returns a list of account states for the reward program's claim period
func (k Keeper) GetRewardAccountStatesForClaimPeriod(ctx sdk.Context, rewardProgramID, claimPeriodID uint64) ([]types.RewardAccountState, error) {
	states := []types.RewardAccountState{}
	err := k.IterateRewardAccountStates(ctx, rewardProgramID, claimPeriodID, func(state types.RewardAccountState) (stop bool) {
		states = append(states, state)
		return false
	})
	return states, err
}

// Returns a list of account states for the reward program
func (k Keeper) GetRewardAccountStatesForRewardProgram(ctx sdk.Context, rewardProgramID uint64) ([]types.RewardAccountState, error) {
	states := []types.RewardAccountState{}
	err := k.IterateRewardAccountStatesForRewardProgram(ctx, rewardProgramID, func(state types.RewardAccountState) (stop bool) {
		states = append(states, state)
		return false
	})
	return states, err
}

// Changes the state for all account states in a reward program's claim period to be claimable
func (k Keeper) MakeRewardClaimsClaimableForPeriod(ctx sdk.Context, rewardProgramID, claimPeriodID uint64) error {
	states, err := k.GetRewardAccountStatesForClaimPeriod(ctx, rewardProgramID, claimPeriodID)
	for _, state := range states {
		state.ClaimStatus = types.RewardAccountState_CLAIM_STATUS_CLAIMABLE
		k.SetRewardAccountState(ctx, state)
	}
	return err
}

// Changes the state for all account states in a reward program to be expired if they are not claimed
func (k Keeper) ExpireRewardClaimsForRewardProgram(ctx sdk.Context, rewardProgramID uint64) error {
	states, err := k.GetRewardAccountStatesForRewardProgram(ctx, rewardProgramID)
	for _, state := range states {
		if state.ClaimStatus == types.RewardAccountState_CLAIM_STATUS_CLAIMED {
			continue
		}
		state.ClaimStatus = types.RewardAccountState_CLAIM_STATUS_EXPIRED
		k.SetRewardAccountState(ctx, state)
	}
	return err
}

func ParseRewardAccountLookUpKey(accountStateAddressLookupKey []byte, addr sdk.AccAddress) (RewardAccountLookup, error) {
	lengthOfAddress := int64(accountStateAddressLookupKey[1:2][0])
	address := sdk.AccAddress(accountStateAddressLookupKey[2 : lengthOfAddress+2])
	if !addr.Equals(address) {
		return RewardAccountLookup{}, fmt.Errorf("addresses do not match up")
	}
	rewardID := binary.BigEndian.Uint64(accountStateAddressLookupKey[lengthOfAddress+2 : lengthOfAddress+2+8])
	claimID := binary.BigEndian.Uint64(accountStateAddressLookupKey[lengthOfAddress+2+8 : lengthOfAddress+2+16])
	return RewardAccountLookup{
		addr:     addr,
		rewardID: rewardID,
		claimID:  claimID,
	}, nil
}
