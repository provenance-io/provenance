package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/provenance-io/provenance/x/ibchooks/types"
)

// RandomizedGenState generates a random GenesisState for ibchooks
func RandomizedGenState(simState *module.SimulationState) {
	ibcHooksGenesis := types.DefaultGenesis()

	bz, err := json.MarshalIndent(ibcHooksGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated ibchooks parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(ibcHooksGenesis)
}
