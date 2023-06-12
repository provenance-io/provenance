package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/x/reward/types"
)

// Simulation parameter constants
const (
	MaxActions         = "max_actions"
	MinActions         = "min_actions"
	TotalRewardsPool   = "total_rewards_pool"
	MaxRewardByAddress = "max_reward_by_address"
)

// GenTotalRewardsPool randomized TotalRewardsPool
func GenTotalRewardsPool(r *rand.Rand) sdk.Coin {
	return sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().BondDenom, int64(randIntBetween(r, 1000, 10000000000)))
}

// GenMaxRewardsByAddress randomized MaxRewardByAddress
func GenMaxRewardsByAddress(r *rand.Rand) sdk.Coin {
	return sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().BondDenom, int64(randIntBetween(r, 1, 999)))
}

// MaxActionsFn randomized MaxActions
func MaxActionsFn(r *rand.Rand) uint64 {
	return uint64(randIntBetween(r, 100, 100000000))
}

// MinActionsFn randomized MinActions
func MinActionsFn(r *rand.Rand) uint64 {
	return uint64(r.Intn(101))
}

// RandomizedGenState generates a random GenesisState for distribution
func RandomizedGenState(simState *module.SimulationState) {
	var totalRewardsPool sdk.Coin
	simState.AppParams.GetOrGenerate(
		simState.Cdc, TotalRewardsPool, &totalRewardsPool, simState.Rand,
		func(r *rand.Rand) { totalRewardsPool = GenTotalRewardsPool(r) },
	)
	var maxRewardsByAddress sdk.Coin
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxRewardByAddress, &maxRewardsByAddress, simState.Rand,
		func(r *rand.Rand) { maxRewardsByAddress = GenMaxRewardsByAddress(r) },
	)

	var maxActions uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxActions, &maxActions, simState.Rand,
		func(r *rand.Rand) { maxActions = MaxActionsFn(r) },
	)

	var minActions uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MinActions, &minActions, simState.Rand,
		func(r *rand.Rand) { minActions = MinActionsFn(r) },
	)

	minDelegation := sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().BondDenom, int64(minActions))

	now := simState.GenTimestamp
	claimPeriodSeconds := uint64(simState.Rand.Intn(100000))
	claimPeriods := uint64(simState.Rand.Intn(100)) + 1
	maxRolloverPeriods := uint64(simState.Rand.Intn(10))
	expireClaimPeriods := uint64(simState.Rand.Intn(100000))
	expectedProgramEndTime := types.CalculateExpectedEndTime(now, claimPeriodSeconds, claimPeriods)
	programEndTimeMax := types.CalculateEndTimeMax(now, claimPeriodSeconds, claimPeriods, maxRolloverPeriods)
	rewardProgram := types.RewardProgram{
		Title:                   "title",
		Description:             "description",
		Id:                      uint64(simState.Rand.Intn(100000)),
		DistributeFromAddress:   simState.Accounts[0].Address.String(),
		TotalRewardPool:         totalRewardsPool,
		RemainingPoolBalance:    totalRewardsPool,
		ClaimedAmount:           sdk.NewInt64Coin(totalRewardsPool.Denom, 0),
		MaxRewardByAddress:      maxRewardsByAddress,
		ProgramStartTime:        now.UTC(),
		ExpectedProgramEndTime:  expectedProgramEndTime.UTC(),
		ProgramEndTimeMax:       programEndTimeMax.UTC(),
		ClaimPeriodSeconds:      claimPeriodSeconds,
		ClaimPeriods:            claimPeriods,
		MaxRolloverClaimPeriods: maxRolloverPeriods,
		ExpirationOffset:        expireClaimPeriods,
		State:                   types.RewardProgram_STATE_PENDING,
		QualifyingActions: []types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Vote{
					Vote: &types.ActionVote{
						MinimumActions:          minActions,
						MaximumActions:          maxActions,
						MinimumDelegationAmount: minDelegation,
					},
				},
			},
		},
		MinimumRolloverAmount: sdk.NewInt64Coin(totalRewardsPool.Denom, 100_000_000_000),
	}

	rewards := types.NewGenesisState(
		uint64(100001),
		[]types.RewardProgram{
			rewardProgram,
		},
		[]types.ClaimPeriodRewardDistribution{},
		[]types.RewardAccountState{},
	)

	bz, err := json.MarshalIndent(&rewards, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated reward programs:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(rewards)
}

// randIntBetween generates a random number between min and max inclusive.
func randIntBetween(r *rand.Rand, min, max int) int {
	return r.Intn(max-min+1) + min
}
