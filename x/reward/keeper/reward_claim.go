package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

// SetRewardClaim sets the reward program in the keeper
func (k Keeper) SetRewardClaim(ctx sdk.Context, rewardClaim types.RewardClaim) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&rewardClaim)
	store.Set(types.GetRewardClaimsKey([]byte(rewardClaim.Address)), bz)
}

// GetRewardClaim returns a RewardClaim by id
func (k Keeper) GetRewardClaim(ctx sdk.Context, addr string) (rewardClaim types.RewardClaim, err error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetRewardClaimsKey([]byte(addr))
	bz := store.Get(key)
	if len(bz) == 0 {
		return rewardClaim, err
	}
	err = k.cdc.Unmarshal(bz, &rewardClaim)
	return rewardClaim, err
}

// IterateRewardClaims  iterates all reward claims with the given handler function.
func (k Keeper) IterateRewardClaims(ctx sdk.Context, handle func(rewardClaim types.RewardClaim) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.RewardClaimKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		record := types.RewardClaim{}
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return err
		}
		if handle(record) {
			break
		}
	}
	return nil
}

func (k Keeper) RewardClaimIsValid(rewardClaim *types.RewardClaim) bool {
	return rewardClaim.Address != ""
}
