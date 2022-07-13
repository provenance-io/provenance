package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/x/reward/types"
	"github.com/tendermint/tendermint/types/time"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
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
	return sdk.NewInt64Coin(pioconfig.DefaultBondDenom, int64(simtypes.RandIntBetween(r, 1000, 10000000000)))
}

// GenMaxRewardsByAddress randomized MaxRewardByAddress
func GenMaxRewardsByAddress(r *rand.Rand) sdk.Coin {
	return sdk.NewInt64Coin(pioconfig.DefaultBondDenom, int64(simtypes.RandIntBetween(r, 1, 999)))
}

// MaxActionsFn randomized MaxActions
func MaxActionsFn(r *rand.Rand) uint64 {
	return uint64(simtypes.RandIntBetween(r, 100, 100000000))
}

// MinActionsFn randomized MinActions
func MinActionsFn(r *rand.Rand) uint64 {
	return uint64(simtypes.RandIntBetween(r, 0, 100))
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

	minDelegation := sdk.NewInt64Coin(pioconfig.DefaultBondDenom, int64(minActions))

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		uint64(simState.Rand.Intn(100000)),
		simState.Accounts[0].Address.String(),
		totalRewardsPool,
		maxRewardsByAddress,
		time.Now(),
		uint64(simState.Rand.Intn(100000)),
		uint64(simState.Rand.Intn(100)),
		uint64(simState.Rand.Intn(100000)),
		[]types.QualifyingAction{
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
	)

	rewards := types.NewGenesisState(
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
