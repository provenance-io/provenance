package simulation

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/provenance-io/provenance/x/expiration/types"
)

// RandomizedGenState generates a random GenesisState for marker
func RandomizedGenState(simState *module.SimulationState) {
	expirationGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		// TODO do we need to add expirations?
		//Expirations: ...,
	}

	bz, err := json.MarshalIndent(&expirationGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated expiration parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&expirationGenesis)
}
