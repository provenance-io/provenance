package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/provenance-io/provenance/x/attribute/types"
)

// Simulation parameter constants
const (
	MaxValueLength = "max_value_length"
)

// GenMaxValueLength randomized CommunityTax
func GenMaxValueLength(r *rand.Rand) uint32 {
	return r.Uint32()
}

// RandomizedGenState generates a random GenesisState for distribution
func RandomizedGenState(simState *module.SimulationState) {
	var maxValueLength uint32
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxValueLength, &maxValueLength, simState.Rand,
		func(r *rand.Rand) { maxValueLength = GenMaxValueLength(r) },
	)

	accountGenesis := types.GenesisState{
		Params: types.Params{
			MaxValueLength: maxValueLength,
		},
		Attributes: []types.Attribute{},
	}

	bz, err := json.MarshalIndent(&accountGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated account parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&accountGenesis)
}
