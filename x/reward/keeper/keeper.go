package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	epochkeeper "github.com/provenance-io/provenance/x/epoch/keeper"
	"github.com/provenance-io/provenance/x/reward/types"
)

const StoreKey = types.ModuleName

type Keeper struct {
	storeKey      sdk.StoreKey
	cdc           codec.BinaryCodec
	EpochKeeper   epochkeeper.Keeper
	stakingKeeper types.StakingKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key sdk.StoreKey,
	epochKeeper epochkeeper.Keeper,
	stakingKeeper types.StakingKeeper,
) Keeper {
	return Keeper{
		storeKey:      key,
		cdc:           cdc,
		EpochKeeper:   epochKeeper,
		stakingKeeper: stakingKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// SetRewardProgram sets the reward program in the keeper
func (k Keeper) SetRewardProgram(ctx sdk.Context, rewardProgram types.RewardProgram) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&rewardProgram)
	store.Set(types.GetRewardProgramKey(int64(rewardProgram.Id)), bz)
}

// GetRewardProgram returns a RewardProgram by id
func (k Keeper) GetRewardProgram(ctx sdk.Context, id int64) (rewardProgram types.RewardProgram, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetRewardProgramKey(id)
	bz := store.Get(key)
	if len(bz) == 0 {
		return rewardProgram, false
	}
	k.cdc.MustUnmarshal(bz, &rewardProgram)

	return rewardProgram, true
}

// IterateRewardPrograms iterates all reward programs with the given handler function.
func (k Keeper) IterateRewardPrograms(ctx sdk.Context, handle func(rewardProgram types.RewardProgram) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.RewardProgramKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		record := types.RewardProgram{}
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return err
		}
		if handle(record) {
			break
		}
	}
	return nil
}

// SetRewardClaim sets the reward program in the keeper
func (k Keeper) SetRewardClaim(ctx sdk.Context, rewardProgram types.RewardClaim) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&rewardProgram)
	store.Set(types.GetRewardClaimsKey([]byte(rewardProgram.Address)), bz)
}

// GetRewardClaim returns a RewardClaim by id
func (k Keeper) GetRewardClaim(ctx sdk.Context, addr []byte) (rewardClaim types.RewardClaim, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetRewardClaimsKey(addr)
	bz := store.Get(key)
	if len(bz) == 0 {
		return types.RewardClaim{}, false
	}
	k.cdc.MustUnmarshal(bz, &rewardClaim)

	return rewardClaim, true
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

// SetEpochRewardDistribution sets the EpochRewardDistribution in the keeper
func (k Keeper) SetEpochRewardDistribution(ctx sdk.Context, epochRewardDistribution types.EpochRewardDistribution) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&epochRewardDistribution)
	store.Set(types.GetEpochRewardDistributionKey(epochRewardDistribution.EpochId, fmt.Sprintf("%d", epochRewardDistribution.RewardProgramId)), bz)
}

// GetEpochRewardDistribution returns a EpochRewardDistribution by epoch id and reward id
func (k Keeper) GetEpochRewardDistribution(ctx sdk.Context, epochId string, rewardId uint64) (epochRewardDistribution types.EpochRewardDistribution, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetEpochRewardDistributionKey(epochId, fmt.Sprintf("%d", rewardId))
	bz := store.Get(key)
	if len(bz) == 0 {
		return epochRewardDistribution, false
	}
	k.cdc.MustUnmarshal(bz, &epochRewardDistribution)

	return epochRewardDistribution, true
}

// IterateEpochRewardDistributions  iterates all epoch reward distributions with the given handler function.
func (k Keeper) IterateEpochRewardDistributions(ctx sdk.Context, handle func(epochRewardDistribution types.EpochRewardDistribution) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.EpochRewardDistributionKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		record := types.EpochRewardDistribution{}
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return err
		}
		if handle(record) {
			break
		}
	}
	return nil
}

// SetEligibilityCriteria sets the reward epoch reward distribution in the keeper
func (k Keeper) SetEligibilityCriteria(ctx sdk.Context, eligibilityCriteria types.EligibilityCriteria) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&eligibilityCriteria)
	store.Set(types.GetEligibilityCriteriaKey(eligibilityCriteria.Name), bz)
}

// GetEligibilityCriteria returns a reward eligibility criteria by name if it exists nil if it does not
func (k Keeper) GetEligibilityCriteria(ctx sdk.Context, name string) (eligibilityCriteria types.EligibilityCriteria, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetEligibilityCriteriaKey(name)
	bz := store.Get(key)
	if len(bz) == 0 {
		return types.EligibilityCriteria{}, false
	}
	k.cdc.MustUnmarshal(bz, &eligibilityCriteria)

	return eligibilityCriteria, true
}

// IterateEligibilityCriterias  iterates all reward eligibility criterions with the given handler function.
func (k Keeper) IterateEligibilityCriterias(ctx sdk.Context, handle func(eligibilityCriteria types.EligibilityCriteria) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.EligibilityCriteriaKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		record := types.EligibilityCriteria{}
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return err
		}
		if handle(record) {
			break
		}
	}
	return nil
}

// SetActionDelegate sets the reward epoch reward distribution in the keeper
func (k Keeper) SetActionDelegate(ctx sdk.Context, actionDelegate types.ActionDelegate) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&actionDelegate)
	store.Set(types.GetActionDelegateKey(), bz)
}

// GetActionDelegate returns a action delegate
func (k Keeper) GetActionDelegate(ctx sdk.Context) (actionDelegate types.ActionDelegate, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetActionDelegateKey())
	if len(bz) == 0 {
		return actionDelegate, false
	}
	k.cdc.MustUnmarshal(bz, &actionDelegate)

	return actionDelegate, true
}

// SetActionTransferDelegations sets the reward epoch reward distribution in the keeper
func (k Keeper) SetActionTransferDelegations(ctx sdk.Context, actionTransferDelegations types.ActionTransferDelegations) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&actionTransferDelegations)
	store.Set(types.GetActionTransferDelegationsKey(), bz)
}

// GetActionTransferDelegations returns a action transfer delegations
func (k Keeper) GetActionTransferDelegations(ctx sdk.Context) (actionTransferDelegations types.ActionTransferDelegations, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetActionTransferDelegationsKey())
	if len(bz) == 0 {
		return actionTransferDelegations, false
	}
	k.cdc.MustUnmarshal(bz, &actionTransferDelegations)

	return actionTransferDelegations, true
}
