package simulation

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/provenance-io/provenance/x/trigger/types"
)

// RandomizedGenState generates a random GenesisState for distribution
func RandomizedGenState(simState *module.SimulationState) {
	triggers := types.NewGenesisState()

	bz, err := json.MarshalIndent(&triggers, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated reward programs:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(triggers)
}
