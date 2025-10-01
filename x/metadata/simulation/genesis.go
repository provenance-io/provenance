package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// RandomizedGenState generates a random GenesisState for marker
func RandomizedGenState(simState *module.SimulationState) {
	metadataGenesis := types.GenesisState{
		Params:          types.DefaultParams(),
		OSLocatorParams: types.DefaultOSLocatorParams(),
	}

	bz, err := json.MarshalIndent(&metadataGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated metadata parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&metadataGenesis)
}
