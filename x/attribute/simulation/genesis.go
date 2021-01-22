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

// GenMaxValueLength randomized MaxValueLength
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

	attributeGenesis := types.GenesisState{
		Params: types.Params{
			MaxValueLength: maxValueLength,
		},
		Attributes: []types.Attribute{},
	}

	bz, err := json.MarshalIndent(&attributeGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated attribute parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&attributeGenesis)
}
