package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

// SetEligibilityCriteria sets the reward epoch reward distribution in the keeper
func (k Keeper) SetEligibilityCriteria(ctx sdk.Context, eligibilityCriteria types.EligibilityCriteria) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&eligibilityCriteria)
	store.Set(types.GetEligibilityCriteriaKey(eligibilityCriteria.Name), bz)
}

// GetEligibilityCriteria returns a reward eligibility criteria by name if it exists nil if it does not
func (k Keeper) GetEligibilityCriteria(ctx sdk.Context, name string) (eligibilityCriteria types.EligibilityCriteria, err error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetEligibilityCriteriaKey(name)
	bz := store.Get(key)
	if len(bz) == 0 {
		return eligibilityCriteria, err
	}
	err = k.cdc.Unmarshal(bz, &eligibilityCriteria)
	return eligibilityCriteria, err
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

func (k Keeper) EligibilityCriteriaIsValid(criteria *types.EligibilityCriteria) bool {
	return criteria.Name != ""
}
