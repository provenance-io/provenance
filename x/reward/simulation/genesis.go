package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/provenance-io/provenance/x/reward/types"
	"github.com/tendermint/tendermint/types/time"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// Simulation parameter constants
const (
	MaxActions = "max_actions"
	MinActions = "min_actions"
)

// MaxActionsFn randomized MaxActions
func MaxActionsFn(r *rand.Rand) uint64 {
	return uint64(simtypes.RandIntBetween(r, 100, 100000000))
}

func MinActionsFn(r *rand.Rand) uint64 {
	return uint64(simtypes.RandIntBetween(r, 0, 100))
}

// RandomizedGenState generates a random GenesisState for distribution
func RandomizedGenState(simState *module.SimulationState) {

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

	minDelegation := sdk.NewInt64Coin("stake", int64(minActions))

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdk.NewInt64Coin("stake", 10000),
		sdk.NewInt64Coin("stake", 10000),
		time.Now(),
		5,
		5,
		0,
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

	msgFeesGenesis := types.GenesisState{
		RewardPrograms: []types.RewardProgram{
			rewardProgram,
		},
	}

	bz, err := json.MarshalIndent(&msgFeesGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated msgfees parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&msgFeesGenesis)
}
