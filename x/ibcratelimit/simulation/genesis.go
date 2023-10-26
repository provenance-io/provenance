package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/provenance-io/provenance/x/ibcratelimit"
)

// RandomizedGenState generates a random GenesisState for ibchooks
func RandomizedGenState(simState *module.SimulationState) {
	genesis := ibcratelimit.DefaultGenesis()

	bz, err := json.MarshalIndent(genesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated ibchooks parameters:\n%s\n", bz)
	simState.GenState[ibcratelimit.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}
