package simulation

import (
	"encoding/json"

	"fmt"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/provenance-io/provenance/x/reward/types"
)

// RandomizedGenState generates a random GenesisState for distribution
func RandomizedGenState(simState *module.SimulationState) {

	rewardsGenesis := types.GenesisState{}

	bz, err := json.MarshalIndent(&rewardsGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated marker parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&rewardsGenesis)
}