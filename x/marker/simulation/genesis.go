package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/provenance-io/provenance/x/marker/types"
)

// RandomizedGenState generates a random GenesisState for marker
func RandomizedGenState(simState *module.SimulationState) {
	markerGenesis := types.GenesisState{
		Params: types.DefaultParams(),
	}

	bz, err := json.MarshalIndent(&markerGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated marker parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&markerGenesis)
}
